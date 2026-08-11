package mcp

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/pocketbase/pocketbase/core"
)

// Repository provides typed access to StyleTailor collections.
type Repository struct {
	app core.App
}

// NewRepository creates a repository.
func NewRepository(app core.App) *Repository {
	return &Repository{app: app}
}

// UpsertUser creates or updates a user profile.
func (r *Repository) UpsertUser(ctx context.Context, userID, portraitURL string, bodyData, preferences map[string]any) (*core.Record, error) {
	collection, err := r.app.FindCollectionByNameOrId("st_users")
	if err != nil {
		return nil, fmt.Errorf("find st_users: %w", err)
	}

	record, err := r.app.FindFirstRecordByFilter("st_users", "user_id={:uid}", map[string]any{"uid": userID})
	if err != nil {
		record = core.NewRecord(collection)
	}

	record.Set("user_id", userID)
	if portraitURL != "" {
		record.Set("portrait_url", portraitURL)
	}
	if bodyData != nil {
		record.Set("body_data", bodyData)
	}
	if preferences != nil {
		record.Set("preferences", preferences)
	}

	if err := r.app.Save(record); err != nil {
		return nil, err
	}
	return record, nil
}

// CreateRequest persists a new request record.
func (r *Repository) CreateRequest(ctx context.Context, userID, portraitURL, occasion, preference string, bodyData map[string]any) (*core.Record, error) {
	collection, err := r.app.FindCollectionByNameOrId("st_requests")
	if err != nil {
		return nil, fmt.Errorf("find st_requests: %w", err)
	}
	record := core.NewRecord(collection)
	record.Set("user_id", userID)
	record.Set("portrait_url", portraitURL)
	record.Set("occasion", occasion)
	record.Set("preference", preference)
	record.Set("status", "processing")
	if bodyData != nil {
		record.Set("body_data", bodyData)
	}
	if err := r.app.Save(record); err != nil {
		return nil, err
	}
	return record, nil
}

// GetRequest fetches a request by id.
func (r *Repository) GetRequest(ctx context.Context, id string) (*core.Record, error) {
	return r.app.FindRecordById("st_requests", id)
}

// UpdateRequestResult updates the result and status.
func (r *Repository) UpdateRequestResult(ctx context.Context, id string, resultURLs []string, score float64, iterationLog []map[string]any) error {
	record, err := r.app.FindRecordById("st_requests", id)
	if err != nil {
		return err
	}
	record.Set("result_urls", resultURLs)
	record.Set("final_score", score)
	record.Set("status", "completed")
	if len(iterationLog) > 0 {
		record.Set("iteration_log", iterationLog)
	}
	return r.app.Save(record)
}

// AddFeedback saves user feedback for a request.
func (r *Repository) AddFeedback(ctx context.Context, requestID, userID string, rating float64, comment string) (*core.Record, error) {
	collection, err := r.app.FindCollectionByNameOrId("st_feedback")
	if err != nil {
		return nil, err
	}
	record := core.NewRecord(collection)
	record.Set("request_id", requestID)
	record.Set("user_id", userID)
	record.Set("rating", rating)
	record.Set("comment", comment)
	if err := r.app.Save(record); err != nil {
		return nil, err
	}
	return record, nil
}

// QueryCatalog returns candidate products by category and optional tags.
func (r *Repository) QueryCatalog(ctx context.Context, categories []string, tags []string) (map[string][]map[string]string, error) {
	result := make(map[string][]map[string]string)
	for _, cat := range categories {
		records, err := r.app.FindRecordsByFilter("st_products", "category={:cat}", "-created", 5, 0, map[string]any{"cat": cat})
		if err != nil {
			return nil, err
		}
		items := []map[string]string{}
		for _, rec := range records {
			items = append(items, map[string]string{
				"id":       rec.Id,
				"sku":      rec.GetString("sku"),
				"name":     rec.GetString("name"),
				"image":    rec.GetString("image_url"),
				"category": rec.GetString("category"),
			})
		}
		result[cat] = items
	}
	_ = tags
	return result, nil
}

// SeedProducts inserts mock products from a JSON string if the table is empty.
func (r *Repository) SeedProducts(ctx context.Context, data string) error {
	count, err := r.app.CountRecords("st_products")
	if err != nil {
		return err
	}
	if count > 0 {
		return nil
	}
	var products []map[string]any
	if err := json.Unmarshal([]byte(data), &products); err != nil {
		return err
	}
	collection, err := r.app.FindCollectionByNameOrId("st_products")
	if err != nil {
		return err
	}
	for _, p := range products {
		record := core.NewRecord(collection)
		record.Set("sku", p["sku"])
		record.Set("name", p["name"])
		record.Set("category", p["category"])
		record.Set("tags", p["tags"])
		record.Set("image_url", p["image_url"])
		if err := r.app.Save(record); err != nil {
			return err
		}
	}
	return nil
}
