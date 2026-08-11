package migrations

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/pocketbase/pocketbase/core"
)

func init() {
	core.SystemMigrations.Register(func(txApp core.App) error {
		return importSeedProducts(txApp)
	}, nil, "20260811_import_seed_products.go")
}

func importSeedProducts(app core.App) error {
	data, err := os.ReadFile(filepath.Join(app.DataDir(), "..", "migrations", "seed_products.json"))
	if err != nil {
		// Fallback for dev mode / go run
		if _, statErr := os.Stat("migrations/seed_products.json"); statErr == nil {
			data, err = os.ReadFile("migrations/seed_products.json")
		}
		if err != nil {
			return fmt.Errorf("read seed products: %w", err)
		}
	}
	var products []map[string]any
	if err := json.Unmarshal(data, &products); err != nil {
		return fmt.Errorf("unmarshal seed products: %w", err)
	}

	collection, err := app.FindCollectionByNameOrId("st_products")
	if err != nil {
		return fmt.Errorf("find st_products: %w", err)
	}

	for _, p := range products {
		record := core.NewRecord(collection)
		record.Set("sku", p["sku"])
		record.Set("name", p["name"])
		category, _ := p["category"].(string)
		record.Set("category", category)
		record.Set("tags", p["tags"])
		record.Set("metadata", p["metadata"])

		record.Set("image_url", p["image_url"])

		if err := app.Save(record); err != nil {
			// Log and continue; duplicates or bad rows shouldn't block the whole migration.
			fmt.Fprintf(os.Stderr, "failed to import product %v: %v\n", p["sku"], err)
		}
	}

	return nil
}

func mapCategory(cat string) string {
	switch cat {
	case "上衣":
		return "upper_body"
	case "下装", "裤装":
		return "lower_body"
	case "连衣/裙装":
		return "dresses"
	case "鞋履":
		return "shoes"
	case "配饰", "珠宝首饰":
		return "belt"
	case "箱包":
		return "hat"
	default:
		return "upper_body"
	}
}
