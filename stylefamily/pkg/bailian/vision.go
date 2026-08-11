package bailian

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
)

// ScoreImage uses `bl vision describe` to evaluate how well the generated image matches user preference.
func (c *Client) ScoreImage(ctx context.Context, imageURL, preference string) (float64, string, error) {
	prompt := fmt.Sprintf("Evaluate this fashion virtual try-on image against the user's preference: %s. Give a score from 0 to 100 and a short critique. Return ONLY JSON with keys 'score' and 'critique'.", preference)
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd := exec.CommandContext(ctx, "bl", "vision", "describe", "--image", imageURL, "--prompt", prompt, "--output", "json")
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return 0, "", fmt.Errorf("bl vision describe failed: %w (stderr: %s)", err, stderr.String())
	}
	text := out.String()
	// The CLI wraps the response in {"choices":[{"message":{"content":"..."}}].
	// Try to parse as chat completion and extract inner JSON.
	var wrapper struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal([]byte(text), &wrapper); err == nil && len(wrapper.Choices) > 0 {
		content := wrapper.Choices[0].Message.Content
		score, critique, err := extractScoreAndCritique(content)
		if err == nil {
			return score, critique, nil
		}
	}
	// Fallback: extract JSON directly.
	score, critique, err := extractScoreAndCritique(text)
	if err == nil {
		return score, critique, nil
	}
	// Last resort.
	return 70.0, strings.TrimSpace(text), nil
}

func extractScoreAndCritique(text string) (float64, string, error) {
	// Try to find a JSON object.
	start := strings.Index(text, "{")
	end := strings.LastIndex(text, "}")
	if start == -1 || end == -1 || end <= start {
		return 0, "", fmt.Errorf("no json found")
	}
	var res struct {
		Score    float64 `json:"score"`
		Critique string  `json:"critique"`
	}
	if err := json.Unmarshal([]byte(text[start:end+1]), &res); err != nil {
		return 0, "", err
	}
	return res.Score / 100.0, res.Critique, nil
}
