package familyoutfit

import (
	"context"
	"fmt"
	"strings"

	"github.com/pocketbase/pocketbase/stylefamily/pkg/bailian"
)

// ImageGenerator generates try-on images.
type ImageGenerator struct {
	bailian *bailian.Client
}

// NewImageGenerator creates an image generator.
func NewImageGenerator(b *bailian.Client) *ImageGenerator {
	return &ImageGenerator{bailian: b}
}

// GenerateTryOn generates a placeholder or AI try-on image for an outfit plan.
func (ig *ImageGenerator) GenerateTryOn(ctx context.Context, plan OutfitPlan) (string, error) {
	// Build a prompt from selected items and member info.
	itemNames := make([]string, 0, len(plan.Items))
	for _, it := range plan.Items {
		itemNames = append(itemNames, it.Item.Name)
	}
	prompt := fmt.Sprintf("Virtual try-on photo of a person wearing %s. Clean background, full body, high quality fashion photo.", strings.Join(itemNames, ", "))
	negative := "blurry, distorted hands, extra limbs, text, watermark, inconsistent face"
	urls, err := ig.bailian.GenerateImage(ctx, "", prompt, negative, "1024*1024", 1)
	if err != nil {
		return "", err
	}
	if len(urls) == 0 {
		return "", fmt.Errorf("no image returned")
	}
	return urls[0], nil
}
