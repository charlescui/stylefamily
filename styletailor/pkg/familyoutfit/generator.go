package familyoutfit

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/pocketbase/pocketbase/styletailor/pkg/bailian"
)

// Generator generates outfit plans based on family members and wardrobe.
type Generator struct {
	repo    *Repository
	bailian *bailian.Client
	chatModel string
}

// NewGenerator creates a generator.
func NewGenerator(repo *Repository, b *bailian.Client) *Generator {
	return &Generator{
		repo:    repo,
		bailian: b,
		chatModel: "qwen3.8-max",
	}
}

// GenerateWeeklyOutfit generates a weekly outfit for all family members.
func (g *Generator) GenerateWeeklyOutfit(ctx context.Context, season, occasion string) (*WeeklyOutfit, error) {
	members, err := g.repo.ListFamilyMembers(ctx)
	if err != nil {
		return nil, fmt.Errorf("list family members: %w", err)
	}
	if len(members) == 0 {
		return nil, fmt.Errorf("no family members found")
	}
	items, err := g.repo.ListWardrobeItems(ctx)
	if err != nil {
		return nil, fmt.Errorf("list wardrobe: %w", err)
	}

	if season == "" {
		season = currentSeason()
	}
	if occasion == "" {
		occasion = "周末家庭聚会"
	}

	now := time.Now()
	weekly := &WeeklyOutfit{
		WeekID:    fmt.Sprintf("weekly_%s", now.Format("2006-01-02")),
		Date:      now.Format("2006-01-02"),
		Season:    season,
		Occasion:  occasion,
		Theme:     "本周家庭穿搭",
		Published: true,
	}

	plans := make([]OutfitPlan, 0, len(members))
	for _, m := range members {
		plan := g.generatePlan(ctx, m, items, season, occasion)
		plans = append(plans, plan)
	}
	weekly.Plans = plans
	weekly.Summary = fmt.Sprintf("本周共 %d 套穿搭，通过 %d 套", len(plans), countApproved(plans))
	return weekly, nil
}

func (g *Generator) generatePlan(ctx context.Context, member FamilyMember, items []WardrobeItem, season, occasion string) OutfitPlan {
	plan := OutfitPlan{
		MemberID:   member.MemberID,
		MemberName: member.DisplayName,
		Theme:      fmt.Sprintf("%s的%s穿搭", member.DisplayName, season),
		Status:     "pending",
	}

	// Filter by season and availability.
	candidates := make([]WardrobeItem, 0)
	for _, it := range items {
		if it.Available && contains(it.Seasons, season) {
			candidates = append(candidates, it)
		}
	}

	// Prefer matching style and avoid colors.
	preferred := make([]WardrobeItem, 0)
	fallback := make([]WardrobeItem, 0)
	for _, it := range candidates {
		if matchesStyle(it, member.StylePreferences) && !containsString(member.AvoidColors, it.ColorName) {
			preferred = append(preferred, it)
		} else {
			fallback = append(fallback, it)
		}
	}
	if len(preferred) == 0 {
		preferred = fallback
	}

	// Use LLM if possible, otherwise select by category.
	selected := g.selectItems(ctx, member, preferred, season, occasion)
	plan.Items = selected
	plan.Description = "基于商品库自动匹配的搭配方案"
	return plan
}

func (g *Generator) selectItems(ctx context.Context, member FamilyMember, items []WardrobeItem, season, occasion string) []OutfitItem {
	// Group by category and pick one per category.
	byCat := make(map[string]WardrobeItem)
	for _, it := range items {
		if _, ok := byCat[it.Category]; !ok {
			byCat[it.Category] = it
		}
	}
	result := make([]OutfitItem, 0, len(byCat))
	for _, it := range byCat {
		result = append(result, OutfitItem{
			Item:   it,
			Reason: fmt.Sprintf("匹配%s的%s穿搭需求", member.DisplayName, season),
		})
	}
	return result
}

// Evaluate scores the outfit plan across multiple dimensions.
func Evaluate(plan *OutfitPlan, season, occasion string) {
	items := plan.Items
	if len(items) == 0 {
		plan.Score = 0
		plan.Status = "rejected"
		return
	}

	scores := ScoreDetails{
		Season:       seasonScore(items, season),
		Occasion:     occasionScore(items, occasion),
		Color:        colorScore(items),
		Style:        styleScore(items),
		Completeness: completenessScore(items),
	}
	overall := (scores.Season + scores.Occasion + scores.Color + scores.Style + scores.Completeness) / 5.0
	scores.Overall = overall
	plan.Scores = scores
	plan.Score = overall
	if overall >= 0.75 {
		plan.Status = "approved"
	} else {
		plan.Status = "rejected"
	}
}

func seasonScore(items []OutfitItem, season string) float64 {
	if len(items) == 0 {
		return 0
	}
	count := 0
	for _, it := range items {
		if contains(it.Item.Seasons, season) {
			count++
		}
	}
	return float64(count) / float64(len(items))
}

func occasionScore(items []OutfitItem, occasion string) float64 {
	if len(items) == 0 {
		return 0
	}
	matched := 0
	for _, it := range items {
		tags := it.Item.Tags
		if contains(tags, occasion) || contains(tags, "百搭") || contains(tags, "通勤") {
			matched++
		}
	}
	return float64(matched) / float64(len(items))
}

func colorScore(items []OutfitItem) float64 {
	colors := make(map[string]struct{})
	for _, it := range items {
		colors[it.Item.ColorName] = struct{}{}
	}
	n := len(colors)
	if n <= 2 {
		return 1.0
	}
	if n <= 4 {
		return 0.8
	}
	if n <= 5 {
		return 0.6
	}
	return 0.4
}

func styleScore(items []OutfitItem) float64 {
	styleCount := make(map[string]int)
	total := 0
	for _, it := range items {
		for _, s := range it.Item.Style {
			styleCount[s]++
			total++
		}
	}
	if total == 0 {
		return 0
	}
	max := 0
	for _, c := range styleCount {
		if c > max {
			max = c
		}
	}
	return float64(max) / (float64(total) / float64(len(items)))
}

func completenessScore(items []OutfitItem) float64 {
	cats := make(map[string]struct{})
	for _, it := range items {
		cats[it.Item.Category] = struct{}{}
	}
	required := map[string]struct{}{"upper": {}, "lower": {}, "shoes": {}}
	if _, hasDress := cats["dress"]; hasDress {
		required = map[string]struct{}{"dress": {}, "shoes": {}}
	}
	matched := 0
	for k := range required {
		if _, ok := cats[k]; ok {
			matched++
		}
	}
	return float64(matched) / float64(len(required))
}

func currentSeason() string {
	m := time.Now().Month()
	switch m {
	case 3, 4, 5:
		return "spring"
	case 6, 7, 8:
		return "summer"
	case 9, 10, 11:
		return "autumn"
	default:
		return "winter"
	}
}

func contains[T comparable](arr []T, v T) bool {
	for _, a := range arr {
		if a == v {
			return true
		}
	}
	return false
}

func containsString(arr []string, v string) bool {
	for _, a := range arr {
		if strings.TrimSpace(a) == strings.TrimSpace(v) {
			return true
		}
	}
	return false
}

func matchesStyle(item WardrobeItem, prefs []string) bool {
	for _, p := range prefs {
		for _, s := range item.Style {
			if strings.Contains(s, p) || strings.Contains(p, s) {
				return true
			}
		}
	}
	return false
}

func countApproved(plans []OutfitPlan) int {
	count := 0
	for _, p := range plans {
		if p.Status == "approved" {
			count++
		}
	}
	return count
}

// TryGeneratePlanWithLLM attempts to use LLM for better selection.
func (g *Generator) TryGeneratePlanWithLLM(ctx context.Context, member FamilyMember, items []WardrobeItem, season, occasion string) ([]OutfitItem, error) {
	// Build prompt and call Bailian chat.
	itemText := ""
	for _, it := range items {
		itemText += fmt.Sprintf("- %s: %s (category: %s, color: %s, style: %s, tags: %s)\n",
			it.SKU, it.Name, it.Category, it.ColorName, strings.Join(it.Style, ","), strings.Join(it.Tags, ","))
	}
	prompt := fmt.Sprintf("为家庭成员 %s 设计一套%s适合%s的穿搭。可选商品：\n%s\n请返回 JSON {\"items\":[{\"sku\":\"\",\"reason\":\"\"}]}",
		member.DisplayName, season, occasion, itemText)
	messages := []bailian.ChatMessage{
		{Role: "system", Content: "你是专业家庭穿搭顾问。请仅返回 JSON。"},
		{Role: "user", Content: prompt},
	}
	text, err := g.bailian.Chat(ctx, g.chatModel, messages)
	if err != nil {
		return nil, err
	}
	var parsed struct {
		Items []struct {
			SKU    string `json:"sku"`
			Reason string `json:"reason"`
		} `json:"items"`
	}
	if err := json.Unmarshal([]byte(text), &parsed); err != nil {
		return nil, err
	}
	itemMap := make(map[string]WardrobeItem)
	for _, it := range items {
		itemMap[it.SKU] = it
	}
	result := make([]OutfitItem, 0, len(parsed.Items))
	for _, sel := range parsed.Items {
		if it, ok := itemMap[sel.SKU]; ok {
			result = append(result, OutfitItem{Item: it, Reason: sel.Reason})
		}
	}
	return result, nil
}
