package familyoutfit

import (
	"context"
	"fmt"

	"github.com/pocketbase/pocketbase/core"
)

// FamilyMember represents a family member record.
type FamilyMember struct {
	ID               string   `json:"id"`
	MemberID         string   `json:"member_id"`
	DisplayName      string   `json:"display_name"`
	Gender           string   `json:"gender"`
	Age              int      `json:"age"`
	HeightCm         int      `json:"height_cm"`
	WeightKg         int      `json:"weight_kg"`
	StylePreferences []string `json:"style_preferences"`
	FavoriteColors   []string `json:"favorite_colors"`
	AvoidColors      []string `json:"avoid_colors"`
	Size             string   `json:"size"`
	CommonOccasions  []string `json:"common_occasions"`
	PortraitURL      string   `json:"portrait_url"`
	ModelPrompt      string   `json:"model_prompt"`
}

// WardrobeItem represents a clothing item in the wardrobe.
type WardrobeItem struct {
	ID       string   `json:"id"`
	SKU      string   `json:"sku"`
	Name     string   `json:"name"`
	Category string   `json:"category"`
	Seasons  []string `json:"seasons"`
	Color    string   `json:"color"`
	ColorName string  `json:"color_name"`
	Style    []string `json:"style"`
	Size     []string `json:"size"`
	ImageURL string   `json:"image_url"`
	Tags     []string `json:"tags"`
	Price    float64  `json:"price"`
	Available bool    `json:"available"`
}

// OutfitItem is a selected wardrobe item for an outfit.
type OutfitItem struct {
	Item   WardrobeItem `json:"item"`
	Reason string       `json:"reason"`
}

// OutfitPlan represents one member's outfit plan.
type OutfitPlan struct {
	MemberID         string       `json:"member_id"`
	MemberName       string       `json:"member_name"`
	Theme            string       `json:"theme"`
	Description      string       `json:"description"`
	Items            []OutfitItem `json:"items"`
	Score            float64      `json:"score"`
	Scores           ScoreDetails `json:"scores"`
	TryonImageURL    string       `json:"tryon_image_url"`
	Status           string       `json:"status"`
}

// ScoreDetails holds evaluation scores.
type ScoreDetails struct {
	Season       float64 `json:"season"`
	Occasion     float64 `json:"occasion"`
	Color        float64 `json:"color"`
	Style        float64 `json:"style"`
	Completeness float64 `json:"completeness"`
	Overall      float64 `json:"overall"`
}

// WeeklyOutfit is the generated weekly outfit recommendation.
type WeeklyOutfit struct {
	WeekID    string       `json:"week_id"`
	Date      string       `json:"date"`
	Season    string       `json:"season"`
	Occasion  string       `json:"occasion"`
	Theme     string       `json:"theme"`
	Summary   string       `json:"summary"`
	Plans     []OutfitPlan `json:"plans"`
	Published bool         `json:"published"`
}

// Repository provides typed access to family outfit collections.
type Repository struct {
	app core.App
}

// NewRepository creates a repository.
func NewRepository(app core.App) *Repository {
	return &Repository{app: app}
}

// ListFamilyMembers returns all family members.
func (r *Repository) ListFamilyMembers(ctx context.Context) ([]FamilyMember, error) {
	records, err := r.app.FindRecordsByFilter("st_family_members", "", "created", 100, 0, nil)
	if err != nil {
		return nil, err
	}
	members := make([]FamilyMember, 0, len(records))
	for _, rec := range records {
		members = append(members, recordToFamilyMember(rec))
	}
	return members, nil
}

// UpsertFamilyMember creates or updates a family member.
func (r *Repository) UpsertFamilyMember(ctx context.Context, m FamilyMember) error {
	collection, err := r.app.FindCollectionByNameOrId("st_family_members")
	if err != nil {
		return err
	}
	var record *core.Record
	if existing, err := r.app.FindFirstRecordByFilter("st_family_members", "member_id={:mid}", map[string]any{"mid": m.MemberID}); err == nil {
		record = existing
	} else {
		record = core.NewRecord(collection)
	}
	record.Set("member_id", m.MemberID)
	record.Set("display_name", m.DisplayName)
	record.Set("gender", m.Gender)
	record.Set("age", m.Age)
	record.Set("height_cm", m.HeightCm)
	record.Set("weight_kg", m.WeightKg)
	record.Set("style_preferences", m.StylePreferences)
	record.Set("favorite_colors", m.FavoriteColors)
	record.Set("avoid_colors", m.AvoidColors)
	record.Set("size", m.Size)
	record.Set("common_occasions", m.CommonOccasions)
	record.Set("portrait_url", m.PortraitURL)
	record.Set("model_prompt", m.ModelPrompt)
	return r.app.Save(record)
}

// ListWardrobeItems returns all wardrobe items.
func (r *Repository) ListWardrobeItems(ctx context.Context) ([]WardrobeItem, error) {
	records, err := r.app.FindRecordsByFilter("st_wardrobe_items", "", "created", 500, 0, nil)
	if err != nil {
		return nil, err
	}
	items := make([]WardrobeItem, 0, len(records))
	for _, rec := range records {
		items = append(items, recordToWardrobeItem(rec))
	}
	return items, nil
}

// SaveWeeklyOutfit persists a weekly outfit.
func (r *Repository) SaveWeeklyOutfit(ctx context.Context, w WeeklyOutfit) error {
	collection, err := r.app.FindCollectionByNameOrId("st_weekly_outfits")
	if err != nil {
		return err
	}
	record, err := r.app.FindFirstRecordByFilter("st_weekly_outfits", "week_id={:wid}", map[string]any{"wid": w.WeekID})
	if err != nil {
		record = core.NewRecord(collection)
	}
	record.Set("week_id", w.WeekID)
	record.Set("date", w.Date)
	record.Set("season", w.Season)
	record.Set("occasion", w.Occasion)
	record.Set("theme", w.Theme)
	record.Set("summary", w.Summary)
	record.Set("plans", w.Plans)
	record.Set("published", w.Published)
	return r.app.Save(record)
}

// ListWeeklyOutfits returns weekly outfits ordered by date desc.
func (r *Repository) ListWeeklyOutfits(ctx context.Context, limit int) ([]WeeklyOutfit, error) {
	if limit <= 0 {
		limit = 10
	}
	records, err := r.app.FindRecordsByFilter("st_weekly_outfits", "", "-created", limit, 0, nil)
	if err != nil {
		return nil, err
	}
	outfits := make([]WeeklyOutfit, 0, len(records))
	for _, rec := range records {
		outfits = append(outfits, recordToWeeklyOutfit(rec))
	}
	return outfits, nil
}

// GetLatestWeeklyOutfit returns the most recent weekly outfit.
func (r *Repository) GetLatestWeeklyOutfit(ctx context.Context) (*WeeklyOutfit, error) {
	records, err := r.app.FindRecordsByFilter("st_weekly_outfits", "", "-created", 1, 0, nil)
	if err != nil {
		return nil, err
	}
	if len(records) == 0 {
		return nil, fmt.Errorf("no weekly outfit found")
	}
	w := recordToWeeklyOutfit(records[0])
	return &w, nil
}

func recordToFamilyMember(rec *core.Record) FamilyMember {
	return FamilyMember{
		ID:               rec.Id,
		MemberID:         rec.GetString("member_id"),
		DisplayName:      rec.GetString("display_name"),
		Gender:           rec.GetString("gender"),
		Age:              rec.GetInt("age"),
		HeightCm:         rec.GetInt("height_cm"),
		WeightKg:         rec.GetInt("weight_kg"),
		StylePreferences: getStringSlice(rec, "style_preferences"),
		FavoriteColors:   getStringSlice(rec, "favorite_colors"),
		AvoidColors:      getStringSlice(rec, "avoid_colors"),
		Size:             rec.GetString("size"),
		CommonOccasions:  getStringSlice(rec, "common_occasions"),
		PortraitURL:      rec.GetString("portrait_url"),
		ModelPrompt:      rec.GetString("model_prompt"),
	}
}

func recordToWardrobeItem(rec *core.Record) WardrobeItem {
	return WardrobeItem{
		ID:        rec.Id,
		SKU:       rec.GetString("sku"),
		Name:      rec.GetString("name"),
		Category:  rec.GetString("category"),
		Seasons:   getStringSlice(rec, "seasons"),
		Color:     rec.GetString("color"),
		ColorName: rec.GetString("color_name"),
		Style:     getStringSlice(rec, "style"),
		Size:      getStringSlice(rec, "size"),
		ImageURL:  rec.GetString("image_url"),
		Tags:      getStringSlice(rec, "tags"),
		Price:     rec.GetFloat("price"),
		Available: rec.GetBool("available"),
	}
}

func recordToWeeklyOutfit(rec *core.Record) WeeklyOutfit {
	return WeeklyOutfit{
		WeekID:    rec.GetString("week_id"),
		Date:      rec.GetString("date"),
		Season:    rec.GetString("season"),
		Occasion:  rec.GetString("occasion"),
		Theme:     rec.GetString("theme"),
		Summary:   rec.GetString("summary"),
		Plans:     rec.Get("plans").([]OutfitPlan),
		Published: rec.GetBool("published"),
	}
}

func getStringSlice(rec *core.Record, field string) []string {
	val := rec.Get(field)
	if val == nil {
		return nil
	}
	if arr, ok := val.([]string); ok {
		return arr
	}
	if arr, ok := val.([]any); ok {
		out := make([]string, 0, len(arr))
		for _, v := range arr {
			if s, ok := v.(string); ok {
				out = append(out, s)
			}
		}
		return out
	}
	return nil
}
