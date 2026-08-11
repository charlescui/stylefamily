package mcp

import (
	"context"
	"fmt"
	"strings"

	"github.com/pocketbase/pocketbase/stylefamily/pkg/bailian"
)

// MultiViewGenerator generates additional pose/angle images from a base look image.
type MultiViewGenerator struct {
	bailian *bailian.Client
}

// NewMultiViewGenerator creates a generator.
func NewMultiViewGenerator(b *bailian.Client) *MultiViewGenerator {
	return &MultiViewGenerator{bailian: b}
}

// ViewSpec describes a target view.
type ViewSpec struct {
	Name   string
	Prompt string
}

// Generate poses returns several angle prompts for one look.
func (g *MultiViewGenerator) GenerateViews(ctx context.Context, baseImageURL, preference string) (map[string]string, error) {
	views := []ViewSpec{
		{Name: "front", Prompt: fmt.Sprintf("Full-body front view fashion photo, same person and outfit as reference. Context: %s", preference)},
		{Name: "side", Prompt: fmt.Sprintf("Full-body side view fashion photo, same person and outfit, consistent face and body. Context: %s", preference)},
		{Name: "back", Prompt: fmt.Sprintf("Full-body back view fashion photo, same person and outfit from behind. Context: %s", preference)},
		{Name: "walking", Prompt: fmt.Sprintf("Dynamic full-body photo of the same person walking naturally, same outfit. Context: %s", preference)},
	}

	results := make(map[string]string)
	for _, v := range views {
		negative := "blurry, distorted face, different person, different clothes, text, watermark"
		urls, err := g.bailian.GenerateImage(ctx, "qwen-image-2.0", v.Prompt, negative, "1024*1024", 1)
		if err != nil {
			continue
		}
		if len(urls) > 0 {
			results[v.Name] = urls[0]
		}
	}
	return results, nil
}

// ComposePrompt injects a reference image URL into the prompt for consistency.
func ComposePrompt(baseImageURL, preference, pose string) string {
	return fmt.Sprintf("Reference image: %s. Generate a %s fashion photo with the same model, same outfit and consistent face. Context: %s", baseImageURL, pose, preference)
}

// FilterURLs removes common placeholder URLs.
func FilterURLs(urls []string) []string {
	filtered := []string{}
	for _, u := range urls {
		if u != "" && !strings.Contains(u, "example.com") {
			filtered = append(filtered, u)
		}
	}
	return filtered
}
