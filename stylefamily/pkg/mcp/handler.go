package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tools/router"
	"github.com/pocketbase/pocketbase/stylefamily/pkg/bailian"
)

// Handler exposes StyleFamily MCP tools via an HTTP API.
type Handler struct {
	app         core.App
	bailian     *bailian.Client
	repo        *Repository
	multiViewGen *MultiViewGenerator
	chatModel   string
	imageModel  string
	maxIter     int
}

// NewHandler creates the MCP handler.
func NewHandler(app core.App, b *bailian.Client) *Handler {
	return &Handler{
		app:         app,
		bailian:     b,
		repo:        NewRepository(app),
		multiViewGen: NewMultiViewGenerator(b),
		chatModel:   "qwen3.8-max",
		imageModel:  "qwen-image-2.0",
		maxIter:     3,
	}
}

// RegisterRoutes registers the MCP HTTP routes.
func (h *Handler) RegisterRoutes(r *router.Router[*core.RequestEvent]) {
	r.POST("/mcp/tools/list", func(e *core.RequestEvent) error {
		return h.handleToolsList(e.Response, e.Request)
	})
	r.POST("/mcp/tools/call", func(e *core.RequestEvent) error {
		return h.handleToolsCall(e.Response, e.Request)
	})
}

// toolsListResponse is the JSON-RPC style tools list.
type toolsListResponse struct {
	Tools []ToolDefinition `json:"tools"`
}

// ToolDefinition describes an exposed MCP tool.
type ToolDefinition struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Parameters  any    `json:"parameters"`
}

func (h *Handler) handleToolsList(w http.ResponseWriter, r *http.Request) error {
	resp := toolsListResponse{
		Tools: []ToolDefinition{
			{
				Name:        "stylefamily_generate_look",
				Description: "Generate a virtual try-on look for a user based on their body data, portrait and style preference, iterating with negative feedback until the result is good enough.",
				Parameters: map[string]any{
					"type": "object",
					"required": []string{"user_id", "body_data", "portrait_url", "occasion", "preference"},
					"properties": map[string]any{
						"user_id":      map[string]string{"type": "string", "description": "Unique user identifier"},
						"body_data":    map[string]string{"type": "string", "description": "JSON with height, weight, measurements and any body notes"},
						"portrait_url": map[string]string{"type": "string", "description": "URL to the user's portrait/avatar photo"},
						"occasion":     map[string]string{"type": "string", "description": "Where the outfit will be worn, e.g. wedding, office, party, daily"},
						"preference":   map[string]string{"type": "string", "description": "Free text style preference from the user"},
						"catalog_filter": map[string]string{"type": "string", "description": "Optional filters for the catalog query"},
					},
				},
			},
			{
				Name:        "stylefamily_feedback",
				Description: "Submit user feedback on a generated look to trigger a negative-feedback iteration.",
				Parameters: map[string]any{
					"type": "object",
					"required": []string{"request_id", "feedback"},
					"properties": map[string]any{
						"request_id": map[string]string{"type": "string", "description": "The look request ID"},
						"feedback":   map[string]string{"type": "string", "description": "What the user dislikes or wants changed"},
					},
				},
			},
			{
				Name:        "stylefamily_get_result",
				Description: "Fetch the status and result URLs of a previously submitted look request.",
				Parameters: map[string]any{
					"type": "object",
					"required": []string{"request_id"},
					"properties": map[string]any{
						"request_id": map[string]string{"type": "string", "description": "The look request ID"},
					},
				},
			},
		},
	}
	return writeJSON(w, resp)
}

// toolsCallRequest is the incoming tool call request.
type toolsCallRequest struct {
	Tool string         `json:"tool"`
	Args map[string]any `json:"args"`
}

func (h *Handler) handleToolsCall(w http.ResponseWriter, r *http.Request) error {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return err
	}
	var req toolsCallRequest
	if err := json.Unmarshal(body, &req); err != nil {
		return err
	}

	switch req.Tool {
	case "stylefamily_generate_look":
		return h.generateLook(w, r, req.Args)
	case "stylefamily_feedback":
		return h.feedback(w, r, req.Args)
	case "stylefamily_get_result":
		return h.getResult(w, r, req.Args)
	default:
		return writeJSON(w, map[string]any{"error": "unknown tool", "tool": req.Tool})
	}
}

func writeJSON(w http.ResponseWriter, v any) error {
	w.Header().Set("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

// generateLook performs the core StyleFamily pipeline.
func (h *Handler) generateLook(w http.ResponseWriter, r *http.Request, args map[string]any) error {
	ctx := r.Context()
	userID, _ := args["user_id"].(string)
	bodyDataStr, _ := args["body_data"].(string)
	portraitURL, _ := args["portrait_url"].(string)
	occasion, _ := args["occasion"].(string)
	preference, _ := args["preference"].(string)

	if userID == "" || portraitURL == "" || occasion == "" {
		return writeJSON(w, map[string]any{"error": "missing required args"})
	}

	var bodyData map[string]any
	_ = json.Unmarshal([]byte(bodyDataStr), &bodyData)

	// Upsert user profile.
	if _, err := h.repo.UpsertUser(ctx, userID, portraitURL, bodyData, nil); err != nil {
		return writeJSON(w, map[string]any{"error": fmt.Sprintf("upsert user failed: %v", err)})
	}

	// Seed mock catalog if empty.
	_ = h.repo.SeedProducts(ctx, mockProductsJSON(portraitURL))

	// Step A: understand user need and pick candidate categories.
	categoryList, prompts, err := h.need2Text(ctx, bodyDataStr, portraitURL, occasion, preference)
	if err != nil {
		return writeJSON(w, map[string]any{"error": fmt.Sprintf("need2text failed: %v", err)})
	}

	// Step B: query catalog for candidate garments.
	candidates := h.queryCatalog(ctx, categoryList, args["catalog_filter"])

	// Persist initial request.
	record, err := h.repo.CreateRequest(ctx, userID, portraitURL, occasion, preference, bodyData)
	if err != nil {
		return writeJSON(w, map[string]any{"error": fmt.Sprintf("create request failed: %v", err)})
	}

	// Step C: generate virtual try-on with hierarchical negative feedback.
	resultURLs := []string{}
	finalPrompt := preference
	iterationLog := []map[string]any{}
	for iter := 0; iter < h.maxIter; iter++ {
		iterResult, err := h.runTryOn(ctx, portraitURL, candidates, prompts, finalPrompt, iter)
		if err != nil {
			_ = h.repo.UpdateRequestResult(ctx, record.Id, resultURLs, 0, iterationLog)
			return writeJSON(w, map[string]any{"error": fmt.Sprintf("try-on failed: %v", err)})
		}
		score, critique, err := h.scoreImage(ctx, portraitURL, preference, iterResult)
		iterationLog = append(iterationLog, map[string]any{
			"iter":     iter,
			"score":    score,
			"critique": critique,
			"urls":     iterResult,
		})
		if err == nil && score >= 0.8 {
			resultURLs = iterResult
			break
		}
		finalPrompt = h.revisePrompt(finalPrompt, critique)
	}

	if len(resultURLs) == 0 && len(iterationLog) > 0 {
		resultURLs = iterationLog[len(iterationLog)-1]["urls"].([]string)
	}

	var finalScore float64
	if len(iterationLog) > 0 {
		finalScore = iterationLog[len(iterationLog)-1]["score"].(float64)
	}

	if err := h.repo.UpdateRequestResult(ctx, record.Id, resultURLs, finalScore, iterationLog); err != nil {
		return writeJSON(w, map[string]any{"error": fmt.Sprintf("update request failed: %v", err)})
	}

	// Step D: generate multi-view / pose variations.
	multiView := map[string]string{}
	if len(resultURLs) > 0 {
		mv, err := h.multiViewGen.GenerateViews(ctx, resultURLs[0], preference)
		if err == nil {
			multiView = mv
		}
	}

	return writeJSON(w, map[string]any{
		"request_id":   record.Id,
		"status":       "completed",
		"result_urls":  resultURLs,
		"multi_view":   multiView,
		"score":        finalScore,
	})
}

// feedback receives user critique and re-iterates the look.
func (h *Handler) feedback(w http.ResponseWriter, r *http.Request, args map[string]any) error {
	ctx := r.Context()
	requestID, _ := args["request_id"].(string)
	feedbackText, _ := args["feedback"].(string)
	if requestID == "" || feedbackText == "" {
		return writeJSON(w, map[string]any{"error": "missing request_id or feedback"})
	}

	record, err := h.repo.GetRequest(ctx, requestID)
	if err != nil {
		return writeJSON(w, map[string]any{"error": "request not found"})
	}

	portraitURL := record.GetString("portrait_url")
	preference := record.GetString("preference")
	revised := h.revisePrompt(preference, feedbackText)

	candidates := h.queryCatalog(ctx, []string{"upper_body", "lower_body"}, nil)
	prompts := map[string]string{"upper_body": revised, "lower_body": revised}
	resultURLs, err := h.runTryOn(ctx, portraitURL, candidates, prompts, revised, 0)
	if err != nil {
		return writeJSON(w, map[string]any{"error": fmt.Sprintf("feedback iteration failed: %v", err)})
	}

	score, critique, _ := h.scoreImage(ctx, portraitURL, revised, resultURLs)
	_ = critique

	if err := h.repo.UpdateRequestResult(ctx, record.Id, resultURLs, score, []map[string]any{}); err != nil {
		return writeJSON(w, map[string]any{"error": fmt.Sprintf("update request failed: %v", err)})
	}

	return writeJSON(w, map[string]any{
		"request_id": record.Id,
		"status":     "completed",
		"result_urls": resultURLs,
		"score":      score,
	})
}

// getResult returns the current request status and URLs.
func (h *Handler) getResult(w http.ResponseWriter, r *http.Request, args map[string]any) error {
	ctx := r.Context()
	requestID, _ := args["request_id"].(string)
	if requestID == "" {
		return writeJSON(w, map[string]any{"error": "missing request_id"})
	}
	record, err := h.repo.GetRequest(ctx, requestID)
	if err != nil {
		return writeJSON(w, map[string]any{"error": "request not found"})
	}
	return writeJSON(w, map[string]any{
		"request_id":   record.Id,
		"status":       record.GetString("status"),
		"result_urls":  record.Get("result_urls"),
		"final_score":  record.GetFloat("final_score"),
		"occasion":     record.GetString("occasion"),
		"preference":     record.GetString("preference"),
		"candidate_products": record.Get("candidate_products"),
	})
}

// need2Text uses Bailian chat to turn user need into structured prompts per category.
func (h *Handler) need2Text(ctx context.Context, bodyData, portraitURL, occasion, preference string) ([]string, map[string]string, error) {
	systemPrompt := "You are a fashion stylist. Given the user's body data, portrait, occasion and preference, return ONLY a JSON object with keys 'category' (array of garment categories) and 'prompts' (object mapping each category to a concise prompt)."
	userPrompt := fmt.Sprintf("Body data: %s\nPortrait: %s\nOccasion: %s\nPreference: %s", bodyData, portraitURL, occasion, preference)

	messages := []bailian.ChatMessage{
		{Role: "system", Content: systemPrompt},
		{Role: "user", Content: userPrompt},
	}
	text, err := h.bailian.Chat(ctx, h.chatModel, messages)
	if err != nil {
		return nil, nil, err
	}
	// Attempt JSON parse.
	var parsed struct {
		Category []string          `json:"category"`
		Prompts  map[string]string `json:"prompts"`
	}
	if err := json.Unmarshal([]byte(text), &parsed); err != nil {
		return []string{"upper_body"}, map[string]string{"upper_body": preference}, nil
	}
	return parsed.Category, parsed.Prompts, nil
}

// runTryOn generates the final look images for each category.
func (h *Handler) runTryOn(ctx context.Context, portraitURL string, candidates map[string][]map[string]string, prompts map[string]string, preference string, iter int) ([]string, error) {
	categories := make([]string, 0, len(prompts))
	for cat := range prompts {
		categories = append(categories, cat)
	}

	allURLs := []string{}
	for _, cat := range categories {
		garmentImages := ""
		if items, ok := candidates[cat]; ok && len(items) > 0 {
			garmentImages = items[0]["image"]
		}
		catPrompt := fmt.Sprintf("Fashion virtual try-on photo of a person wearing a stylish %s. Context: %s. Keep the person's face, pose and body consistent across views. Reference garment image: %s", cat, preference, garmentImages)
		if iter > 0 {
			catPrompt += fmt.Sprintf(". Iteration %d - avoid previous issues", iter)
		}
		negative := "blurry, distorted hands, extra limbs, inconsistent face, text, watermark"
		urls, err := h.bailian.GenerateImage(ctx, h.imageModel, catPrompt, negative, "1024*1024", 2)
		if err != nil {
			return nil, err
		}
		allURLs = append(allURLs, urls...)
	}
	return allURLs, nil
}

// scoreImage scores generated images using vision describe.
func (h *Handler) scoreImage(ctx context.Context, portraitURL, preference string, resultURLs []string) (float64, string, error) {
	if len(resultURLs) == 0 {
		return 0, "no images", nil
	}
	var total float64
	var critiques []string
	for _, url := range resultURLs {
		score, critique, err := h.bailian.ScoreImage(ctx, url, preference)
		if err != nil {
			continue
		}
		total += score
		critiques = append(critiques, critique)
	}
	if len(resultURLs) == 0 {
		return 0, "vision score failed", nil
	}
	avg := total / float64(len(resultURLs))
	critique := strings.Join(critiques, "; ")
	return avg, critique, nil
}

// revisePrompt rewrites the prompt based on critique/feedback.
func (h *Handler) revisePrompt(original, critique string) string {
	return fmt.Sprintf("%s. Previous issues to avoid: %s", original, critique)
}

// queryCatalog queries real product catalog via repository.
func (h *Handler) queryCatalog(ctx context.Context, categories []string, filter any) map[string][]map[string]string {
	candidates, err := h.repo.QueryCatalog(ctx, categories, nil)
	if err != nil {
		return map[string][]map[string]string{}
	}
	_ = filter
	return candidates
}

func mockProductsJSON(portraitURL string) string {
	return fmt.Sprintf(`[
		{"sku":"COAT-001","name":"Beige Wool Coat","category":"upper_body","tags":["coat","minimalist","beige","winter"],"image_url":"https://dashscope-0484.oss-accelerate.aliyuncs.com/7d/70/20260811/566f3ebc/128577b7-cac4-4180-ab01-40062ed2347c.png"},
		{"sku":"TRO-001","name":"Black Slim Trousers","category":"lower_body","tags":["trousers","black","office"],"image_url":"https://dashscope-0484.oss-accelerate.aliyuncs.com/7d/9a/20260811/566f3ebc/36586f74-63f5-4f8c-8fc8-75ce08e3c7cd.png"},
		{"sku":"BLO-001","name":"White Silk Blouse","category":"upper_body","tags":["blouse","white","silk","elegant"],"image_url":"https://dashscope-0484.oss-accelerate.aliyuncs.com/7d/47/20260811/566f3ebc/43ebbb9a-f48b-484d-a352-da547fa756b8.png"}
	]`)
}
