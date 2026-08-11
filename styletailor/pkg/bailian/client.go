package bailian

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

// Client wraps the Bailian CLI (`bl`) for model/image generation tasks.
type Client struct {
	// MaxRetries controls how many times to retry failed CLI calls.
	MaxRetries int
	// Timeout for each CLI invocation.
	Timeout time.Duration
}

// NewClient creates a new Bailian CLI client using the default `bl` binary on PATH.
func NewClient() *Client {
	return &Client{
		MaxRetries: 3,
		Timeout:    120 * time.Second,
	}
}

// ChatRequest describes a chat completion request to Bailian models.
type ChatRequest struct {
	Model    string        `json:"model"`
	Messages []ChatMessage `json:"messages"`
	Stream   bool          `json:"stream,omitempty"`
}

// ChatMessage is a single message in a chat request.
type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ChatResponse is the expected JSON response from `bl chat`.
type ChatResponse struct {
	Choices []struct {
		Message ChatMessage `json:"message"`
	} `json:"choices"`
	ErrorMessage string `json:"error_message,omitempty"`
	RequestID    string `json:"request_id,omitempty"`
}

// Chat invokes a Bailian chat model and returns the assistant text.
func (c *Client) Chat(ctx context.Context, model string, messages []ChatMessage) (string, error) {
	if model == "" {
		model = os.Getenv("BAILIAN_DEFAULT_CHAT_MODEL")
		if model == "" {
			model = "qwen3.8-max"
		}
	}

	// Build `bl text chat` arguments.
	var args []string
	args = append(args, "text", "chat", "--model", model, "--output", "json")

	// First message with system role becomes --system; user/assistant become --message.
	for _, m := range messages {
		switch m.Role {
		case "system":
			args = append(args, "--system", m.Content)
		case "user":
			args = append(args, "--message", m.Content)
		case "assistant":
			args = append(args, "--message", "assistant:"+m.Content)
		default:
			args = append(args, "--message", m.Content)
		}
	}

	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd := exec.CommandContext(ctx, "bl", args...)
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("bl text chat failed: %w (stderr: %s)", err, stderr.String())
	}

	var resp ChatResponse
	if err := json.Unmarshal(out.Bytes(), &resp); err != nil {
		// Fallback: return raw stdout if not JSON.
		return strings.TrimSpace(out.String()), nil
	}
	if resp.ErrorMessage != "" {
		return "", fmt.Errorf("bailian chat error: %s", resp.ErrorMessage)
	}
	if len(resp.Choices) == 0 {
		return strings.TrimSpace(out.String()), nil
	}
	return resp.Choices[0].Message.Content, nil
}

// ImageGenerationRequest describes an image generation request to Bailian.
type ImageGenerationRequest struct {
	Model       string `json:"model"`
	Prompt      string `json:"prompt"`
	NegativePrompt string `json:"negative_prompt,omitempty"`
	Size        string `json:"size,omitempty"`
	N           int    `json:"n,omitempty"`
}

// ImageGenerationResponse holds the URLs returned by Bailian image generation.
type ImageGenerationResponse struct {
	URLs    []string `json:"urls"`
	Images  []struct {
		URL string `json:"url"`
	} `json:"images"`
	Saved        []string `json:"saved"`
	Total        int      `json:"total"`
	ErrorMessage string   `json:"error_message,omitempty"`
	RequestID    string   `json:"request_id,omitempty"`
}

// GenerateImage asks Bailian for an image.
func (c *Client) GenerateImage(ctx context.Context, model, prompt, negative, size string, n int) ([]string, error) {
	if model == "" {
		model = os.Getenv("BAILIAN_DEFAULT_IMAGE_MODEL")
		if model == "" {
			model = "qwen-image-2.0"
		}
	}
	if n <= 0 {
		n = 1
	}
	if size == "" {
		size = "1024*1024"
	}

	var args []string
	args = append(args, "image", "generate", "--model", model, "--prompt", prompt, "--size", size, "--n", fmt.Sprintf("%d", n), "--output", "json")
	if negative != "" {
		args = append(args, "--negative-prompt", negative)
	}

	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd := exec.CommandContext(ctx, "bl", args...)
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("bl image generate failed: %w (stderr: %s)", err, stderr.String())
	}

	var resp ImageGenerationResponse
	if err := json.Unmarshal(out.Bytes(), &resp); err != nil {
		return nil, fmt.Errorf("failed to parse image response: %w (stdout: %s)", err, out.String())
	}
	if resp.ErrorMessage != "" {
		return nil, fmt.Errorf("bailian image error: %s", resp.ErrorMessage)
	}

	urls := make([]string, 0, len(resp.Images)+len(resp.URLs))
	for _, u := range resp.URLs {
		if u != "" {
			urls = append(urls, u)
		}
	}
	for _, img := range resp.Images {
		if img.URL != "" {
			urls = append(urls, img.URL)
		}
	}
	return urls, nil
}
