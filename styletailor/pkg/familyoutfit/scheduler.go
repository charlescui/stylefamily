package familyoutfit

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/pocketbase/pocketbase/styletailor/pkg/bailian"
)

// Scheduler handles periodic outfit generation.
type Scheduler struct {
	repo      *Repository
	generator *Generator
	images    *ImageGenerator
}

// NewScheduler creates a scheduler.
func NewScheduler(repo *Repository, b *bailian.Client) *Scheduler {
	return &Scheduler{
		repo:      repo,
		generator: NewGenerator(repo, b),
		images:    NewImageGenerator(b),
	}
}

// RunOnce generates the weekly outfit immediately.
func (s *Scheduler) RunOnce(ctx context.Context, season, occasion string) (*WeeklyOutfit, error) {
	outfit, err := s.generator.GenerateWeeklyOutfit(ctx, season, occasion)
	if err != nil {
		return nil, err
	}

	// Evaluate and generate try-on images for each plan.
	for i := range outfit.Plans {
		Evaluate(&outfit.Plans[i], outfit.Season, outfit.Occasion)
		if outfit.Plans[i].Status == "approved" {
			url, err := s.images.GenerateTryOn(ctx, outfit.Plans[i])
			if err != nil {
				fmt.Fprintf(os.Stderr, "generate try-on failed: %v\n", err)
			} else {
				outfit.Plans[i].TryonImageURL = url
			}
		}
	}

	if err := s.repo.SaveWeeklyOutfit(ctx, *outfit); err != nil {
		return nil, fmt.Errorf("save weekly outfit: %w", err)
	}

	// Generate docs.
	if err := s.GenerateDocs(outfit); err != nil {
		fmt.Fprintf(os.Stderr, "generate docs failed: %v\n", err)
	}

	return outfit, nil
}

// Start runs the scheduler loop. It checks every minute and triggers on Saturday 07:00.
func (s *Scheduler) Start(ctx context.Context) {
	fmt.Println("家庭穿搭定时任务已启动，每周六 07:00 自动生成穿搭方案")
	for {
		now := time.Now()
		if now.Weekday() == time.Saturday && now.Hour() == 7 && now.Minute() == 0 {
			if _, err := s.RunOnce(ctx, "", ""); err != nil {
				fmt.Fprintf(os.Stderr, "weekly outfit failed: %v\n", err)
			}
			time.Sleep(60 * time.Second)
		} else {
			time.Sleep(30 * time.Second)
		}
	}
}

// GenerateDocs writes a Chinese markdown document for the weekly outfit.
func (s *Scheduler) GenerateDocs(w *WeeklyOutfit) error {
	docsDir := filepath.Join("docs", w.WeekID)
	if err := os.MkdirAll(docsDir, 0755); err != nil {
		return err
	}
	readme := filepath.Join(docsDir, "README.md")
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("# %s\n\n", w.Theme))
	sb.WriteString(fmt.Sprintf("**日期**：%s  \n", w.Date))
	sb.WriteString(fmt.Sprintf("**季节**：%s  \n", w.Season))
	sb.WriteString(fmt.Sprintf("**场景**：%s  \n", w.Occasion))
	sb.WriteString(fmt.Sprintf("**总结**：%s\n\n", w.Summary))
	sb.WriteString("## 家庭成员穿搭\n\n")
	for _, plan := range w.Plans {
		sb.WriteString(fmt.Sprintf("### %s\n\n", plan.MemberName))
		sb.WriteString(fmt.Sprintf("**主题**：%s\n\n", plan.Theme))
		sb.WriteString(fmt.Sprintf("%s\n\n", plan.Description))
		sb.WriteString("| 单品 | 说明 |\n")
		sb.WriteString("|------|------|\n")
		for _, oi := range plan.Items {
			sb.WriteString(fmt.Sprintf("| %s | %s |\n", oi.Item.Name, oi.Reason))
		}
		sb.WriteString("\n")
		if plan.TryonImageURL != "" {
			sb.WriteString(fmt.Sprintf("![试穿效果](%s)\n\n", plan.TryonImageURL))
		}
		sb.WriteString(fmt.Sprintf("**综合评分**：%.2f\n\n", plan.Score))
	}
	return os.WriteFile(readme, []byte(sb.String()), 0644)
}
