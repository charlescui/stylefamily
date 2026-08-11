package migrations

import (
	"fmt"

	"github.com/pocketbase/pocketbase/core"
)

func init() {
	core.SystemMigrations.Register(func(txApp core.App) error {
		return createFamilyOutfitTables(txApp)
	}, nil, "20260812_add_family_outfit_tables")
}

func createFamilyOutfitTables(app core.App) error {
	// st_family_members
	family := core.NewBaseCollection("st_family_members")
	family.ListRule = nil
	family.ViewRule = nil
	family.CreateRule = nil
	family.UpdateRule = nil
	family.DeleteRule = nil
	family.Fields.Add(&core.TextField{Name: "member_id", Required: true})
	family.Fields.Add(&core.TextField{Name: "display_name", Required: true})
	family.Fields.Add(&core.TextField{Name: "gender", Required: false})
	family.Fields.Add(&core.NumberField{Name: "age", Required: false})
	family.Fields.Add(&core.NumberField{Name: "height_cm", Required: false})
	family.Fields.Add(&core.NumberField{Name: "weight_kg", Required: false})
	family.Fields.Add(&core.JSONField{Name: "style_preferences", Required: false})
	family.Fields.Add(&core.JSONField{Name: "favorite_colors", Required: false})
	family.Fields.Add(&core.JSONField{Name: "avoid_colors", Required: false})
	family.Fields.Add(&core.TextField{Name: "size", Required: false})
	family.Fields.Add(&core.JSONField{Name: "common_occasions", Required: false})
	family.Fields.Add(&core.URLField{Name: "portrait_url", Required: false})
	family.Fields.Add(&core.TextField{Name: "model_prompt", Required: false})
	family.Fields.Add(&core.AutodateField{Name: "created", OnCreate: true})
	family.Fields.Add(&core.AutodateField{Name: "updated", OnCreate: true, OnUpdate: true})
	if err := app.Save(family); err != nil {
		return fmt.Errorf("save st_family_members: %w", err)
	}

	// st_wardrobe_items
	wardrobe := core.NewBaseCollection("st_wardrobe_items")
	wardrobe.ListRule = nil
	wardrobe.ViewRule = nil
	wardrobe.CreateRule = nil
	wardrobe.UpdateRule = nil
	wardrobe.DeleteRule = nil
	wardrobe.Fields.Add(&core.TextField{Name: "sku", Required: true})
	wardrobe.Fields.Add(&core.TextField{Name: "name", Required: true})
	wardrobe.Fields.Add(&core.SelectField{
		Name:      "category",
		Required:  true,
		Values:    []string{"upper", "lower", "dress", "outer", "shoes", "accessory"},
		MaxSelect: 1,
	})
	wardrobe.Fields.Add(&core.JSONField{Name: "seasons", Required: false})
	wardrobe.Fields.Add(&core.TextField{Name: "color", Required: false})
	wardrobe.Fields.Add(&core.TextField{Name: "color_name", Required: false})
	wardrobe.Fields.Add(&core.JSONField{Name: "style", Required: false})
	wardrobe.Fields.Add(&core.JSONField{Name: "size", Required: false})
	wardrobe.Fields.Add(&core.URLField{Name: "image_url", Required: false})
	wardrobe.Fields.Add(&core.JSONField{Name: "tags", Required: false})
	wardrobe.Fields.Add(&core.NumberField{Name: "price", Required: false})
	wardrobe.Fields.Add(&core.BoolField{Name: "available", Required: false})
	wardrobe.Fields.Add(&core.AutodateField{Name: "created", OnCreate: true})
	wardrobe.Fields.Add(&core.AutodateField{Name: "updated", OnCreate: true, OnUpdate: true})
	if err := app.Save(wardrobe); err != nil {
		return fmt.Errorf("save st_wardrobe_items: %w", err)
	}

	// st_weekly_outfits
	weekly := core.NewBaseCollection("st_weekly_outfits")
	weekly.ListRule = nil
	weekly.ViewRule = nil
	weekly.CreateRule = nil
	weekly.UpdateRule = nil
	weekly.DeleteRule = nil
	weekly.Fields.Add(&core.TextField{Name: "week_id", Required: true})
	weekly.Fields.Add(&core.TextField{Name: "date", Required: true})
	weekly.Fields.Add(&core.TextField{Name: "season", Required: false})
	weekly.Fields.Add(&core.TextField{Name: "occasion", Required: false})
	weekly.Fields.Add(&core.TextField{Name: "theme", Required: false})
	weekly.Fields.Add(&core.TextField{Name: "summary", Required: false})
	weekly.Fields.Add(&core.JSONField{Name: "plans", Required: false})
	weekly.Fields.Add(&core.BoolField{Name: "published", Required: false})
	weekly.Fields.Add(&core.AutodateField{Name: "created", OnCreate: true})
	weekly.Fields.Add(&core.AutodateField{Name: "updated", OnCreate: true, OnUpdate: true})
	if err := app.Save(weekly); err != nil {
		return fmt.Errorf("save st_weekly_outfits: %w", err)
	}

	return nil
}
