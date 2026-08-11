package familyoutfit

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tools/router"
)

// WebHandler provides HTTP handlers for family outfit UI and API.
type WebHandler struct {
	app core.App
	repo *Repository
}

// NewWebHandler creates a web handler.
func NewWebHandler(app core.App) *WebHandler {
	return &WebHandler{app: app, repo: NewRepository(app)}
}

// RegisterRoutes registers routes.
func (h *WebHandler) RegisterRoutes(r *router.Router[*core.RequestEvent]) {
	r.GET("/api/family_members", func(e *core.RequestEvent) error {
		members, err := h.repo.ListFamilyMembers(e.Request.Context())
		if err != nil {
			return e.JSON(http.StatusInternalServerError, map[string]any{"error": err.Error()})
		}
		return e.JSON(http.StatusOK, members)
	})

	r.GET("/api/latest_outfit", func(e *core.RequestEvent) error {
		outfit, err := h.repo.GetLatestWeeklyOutfit(e.Request.Context())
		if err != nil {
			return e.JSON(http.StatusOK, map[string]any{"plans": []any{}})
		}
		return e.JSON(http.StatusOK, outfit)
	})

	r.POST("/api/generate_outfit", func(e *core.RequestEvent) error {
		return e.JSON(http.StatusOK, map[string]any{"message": "use CLI or scheduler to generate outfits"})
	})
}

// SeedDemoData inserts demo family members and wardrobe items if collections are empty.
func SeedDemoData(app core.App) error {
	repo := NewRepository(app)
	ctx := context.Background()
	members, err := repo.ListFamilyMembers(ctx)
	if err != nil {
		return err
	}
	if len(members) == 0 {
		for _, m := range defaultFamilyMembers() {
			if err := repo.UpsertFamilyMember(ctx, m); err != nil {
				return err
			}
		}
	}
	items, err := repo.ListWardrobeItems(ctx)
	if err != nil {
		return err
	}
	if len(items) == 0 {
		if err := seedWardrobeItems(app); err != nil {
			return err
		}
	}
	return nil
}

func defaultFamilyMembers() []FamilyMember {
	return []FamilyMember{
		{
			MemberID:         "mom",
			DisplayName:      "妈妈",
			Gender:           "female",
			Age:              38,
			HeightCm:         165,
			WeightKg:         55,
			StylePreferences: []string{"优雅", "简约"},
			FavoriteColors:   []string{"米白", "浅蓝", "驼色"},
			AvoidColors:      []string{"亮黄", "荧光粉"},
			Size:             "M",
			CommonOccasions:  []string{"通勤", "周末聚会", "接送孩子"},
			PortraitURL:      "",
			ModelPrompt:      "一位38岁亚洲女性，中等身材，气质优雅，穿着时尚，站立姿势，浅色背景，全身照",
		},
		{
			MemberID:         "dad",
			DisplayName:      "爸爸",
			Gender:           "male",
			Age:              40,
			HeightCm:         178,
			WeightKg:         72,
			StylePreferences: []string{"商务休闲", "简约"},
			FavoriteColors:   []string{"藏青", "灰色", "白色"},
			AvoidColors:      []string{"粉红", "橙色"},
			Size:             "L",
			CommonOccasions:  []string{"通勤", "商务会议", "周末亲子"},
			PortraitURL:      "",
			ModelPrompt:      "一位40岁亚洲男性，身材匀称，商务休闲风格，穿着整洁，站立姿势，浅色背景，全身照",
		},
		{
			MemberID:         "daughter",
			DisplayName:      "女儿",
			Gender:           "female",
			Age:              10,
			HeightCm:         140,
			WeightKg:         32,
			StylePreferences: []string{"甜美", "休闲"},
			FavoriteColors:   []string{"粉色", "浅紫", "白色"},
			AvoidColors:      []string{"黑色", "深灰"},
			Size:             "S",
			CommonOccasions:  []string{"上学", "周末活动", "生日聚会"},
			PortraitURL:      "",
			ModelPrompt:      "一位10岁亚洲女孩，活泼可爱，穿着甜美休闲服装，站立姿势，浅色背景，全身照",
		},
	}
}

func seedWardrobeItems(app core.App) error {
	collection, err := app.FindCollectionByNameOrId("st_wardrobe_items")
	if err != nil {
		return err
	}
	items := []map[string]any{
		{"sku": "W001", "name": "米白色真丝衬衫", "category": "upper", "seasons": []string{"spring", "summer", "autumn"}, "color": "#f5f5dc", "color_name": "米白", "style": []string{"优雅", "简约"}, "size": []string{"S", "M", "L"}, "tags": []string{"百搭", "通勤"}, "price": 299.0, "available": true},
		{"sku": "W002", "name": "高腰阔腿西装裤", "category": "lower", "seasons": []string{"spring", "autumn"}, "color": "#2f3542", "color_name": "藏青", "style": []string{"优雅", "商务"}, "size": []string{"S", "M", "L"}, "tags": []string{"通勤", "显瘦"}, "price": 259.0, "available": true},
		{"sku": "W003", "name": "浅蓝色针织开衫", "category": "outer", "seasons": []string{"spring", "autumn"}, "color": "#87ceeb", "color_name": "浅蓝", "style": []string{"优雅", "休闲"}, "size": []string{"S", "M", "L"}, "tags": []string{"百搭", "温柔"}, "price": 199.0, "available": true},
		{"sku": "M001", "name": "白色商务衬衫", "category": "upper", "seasons": []string{"spring", "summer", "autumn"}, "color": "#ffffff", "color_name": "白色", "style": []string{"商务", "简约"}, "size": []string{"M", "L", "XL"}, "tags": []string{"通勤", "商务"}, "price": 249.0, "available": true},
		{"sku": "M002", "name": "深灰色休闲西裤", "category": "lower", "seasons": []string{"spring", "autumn", "winter"}, "color": "#696969", "color_name": "深灰", "style": []string{"商务休闲", "简约"}, "size": []string{"M", "L", "XL"}, "tags": []string{"通勤", "百搭"}, "price": 279.0, "available": true},
		{"sku": "M003", "name": "藏青色西装外套", "category": "outer", "seasons": []string{"spring", "autumn", "winter"}, "color": "#1e3a5f", "color_name": "藏青", "style": []string{"商务"}, "size": []string{"M", "L", "XL"}, "tags": []string{"商务", "通勤"}, "price": 599.0, "available": true},
		{"sku": "K001", "name": "粉色纯棉T恤", "category": "upper", "seasons": []string{"spring", "summer"}, "color": "#ffb7c5", "color_name": "粉色", "style": []string{"甜美", "休闲"}, "size": []string{"S"}, "tags": []string{"甜美", "舒适"}, "price": 79.0, "available": true},
		{"sku": "K002", "name": "白色蕾丝半身裙", "category": "lower", "seasons": []string{"spring", "summer"}, "color": "#ffffff", "color_name": "白色", "style": []string{"甜美"}, "size": []string{"S"}, "tags": []string{"甜美", "可爱"}, "price": 129.0, "available": true},
		{"sku": "K003", "name": "浅紫色针织外套", "category": "outer", "seasons": []string{"spring", "autumn"}, "color": "#e6e6fa", "color_name": "浅紫", "style": []string{"甜美", "休闲"}, "size": []string{"S"}, "tags": []string{"甜美", "保暖"}, "price": 159.0, "available": true},
	}
	for _, item := range items {
		record := core.NewRecord(collection)
		for k, v := range item {
			record.Set(k, v)
		}
		if err := app.Save(record); err != nil {
			return err
		}
	}
	return nil
}

// MarshalJSON helper to serialize weekly outfit.
func MarshalJSON(v any) ([]byte, error) {
	return json.Marshal(v)
}
