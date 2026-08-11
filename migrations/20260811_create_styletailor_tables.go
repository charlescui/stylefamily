package migrations

import (
	"fmt"

	"github.com/pocketbase/pocketbase/core"
)

func init() {
	core.SystemMigrations.Register(func(txApp core.App) error {
		return createStyleTailorTables(txApp)
	}, nil, "20260811_create_styletailor_tables.go")
}

func createStyleTailorTables(app core.App) error {
	// st_users
	users := core.NewBaseCollection("st_users")
	users.ListRule = nil
	users.ViewRule = nil
	users.CreateRule = nil
	users.UpdateRule = nil
	users.DeleteRule = nil
	users.Fields.Add(&core.TextField{Name: "user_id", Required: true})
	users.Fields.Add(&core.TextField{Name: "portrait_url", Required: false})
	users.Fields.Add(&core.JSONField{Name: "body_data", Required: false})
	users.Fields.Add(&core.JSONField{Name: "preferences", Required: false})
	users.Fields.Add(&core.AutodateField{Name: "created", OnCreate: true})
	users.Fields.Add(&core.AutodateField{Name: "updated", OnCreate: true, OnUpdate: true})
	if err := app.Save(users); err != nil {
		return fmt.Errorf("save st_users: %w", err)
	}

	// st_products
	products := core.NewBaseCollection("st_products")
	products.ListRule = nil
	products.ViewRule = nil
	products.CreateRule = nil
	products.UpdateRule = nil
	products.DeleteRule = nil
	products.Fields.Add(&core.TextField{Name: "sku", Required: true})
	products.Fields.Add(&core.TextField{Name: "name", Required: true})
	products.Fields.Add(&core.SelectField{
		Name:      "category",
		Required:  true,
		Values:    []string{"upper_body", "lower_body", "dresses", "shoes", "hat", "glasses", "belt", "scarf"},
		MaxSelect: 1,
	})
	products.Fields.Add(&core.JSONField{Name: "tags", Required: false})
	products.Fields.Add(&core.JSONField{Name: "metadata", Required: false})
	products.Fields.Add(&core.URLField{Name: "image_url", Required: false})
	products.Fields.Add(&core.AutodateField{Name: "created", OnCreate: true})
	products.Fields.Add(&core.AutodateField{Name: "updated", OnCreate: true, OnUpdate: true})
	if err := app.Save(products); err != nil {
		return fmt.Errorf("save st_products: %w", err)
	}

	// st_requests
	requests := core.NewBaseCollection("st_requests")
	requests.ListRule = nil
	requests.ViewRule = nil
	requests.CreateRule = nil
	requests.UpdateRule = nil
	requests.DeleteRule = nil
	requests.Fields.Add(&core.TextField{Name: "user_id", Required: true})
	requests.Fields.Add(&core.TextField{Name: "portrait_url", Required: true})
	requests.Fields.Add(&core.JSONField{Name: "body_data", Required: false})
	requests.Fields.Add(&core.TextField{Name: "occasion", Required: true})
	requests.Fields.Add(&core.TextField{Name: "preference", Required: true})
	requests.Fields.Add(&core.JSONField{Name: "candidate_products", Required: false})
	requests.Fields.Add(&core.JSONField{Name: "result_urls", Required: false})
	requests.Fields.Add(&core.TextField{Name: "status", Required: true})
	requests.Fields.Add(&core.NumberField{Name: "final_score", Required: false})
	requests.Fields.Add(&core.JSONField{Name: "iteration_log", Required: false})
	requests.Fields.Add(&core.AutodateField{Name: "created", OnCreate: true})
	requests.Fields.Add(&core.AutodateField{Name: "updated", OnCreate: true, OnUpdate: true})
	if err := app.Save(requests); err != nil {
		return fmt.Errorf("save st_requests: %w", err)
	}

	// st_feedback
	feedback := core.NewBaseCollection("st_feedback")
	feedback.ListRule = nil
	feedback.ViewRule = nil
	feedback.CreateRule = nil
	feedback.UpdateRule = nil
	feedback.DeleteRule = nil
	feedback.Fields.Add(&core.TextField{Name: "request_id", Required: true})
	feedback.Fields.Add(&core.TextField{Name: "user_id", Required: true})
	feedback.Fields.Add(&core.NumberField{Name: "rating", Required: false})
	feedback.Fields.Add(&core.TextField{Name: "comment", Required: false})
	feedback.Fields.Add(&core.JSONField{Name: "metadata", Required: false})
	feedback.Fields.Add(&core.AutodateField{Name: "created", OnCreate: true})
	feedback.Fields.Add(&core.AutodateField{Name: "updated", OnCreate: true, OnUpdate: true})
	if err := app.Save(feedback); err != nil {
		return fmt.Errorf("save st_feedback: %w", err)
	}

	return nil
}
