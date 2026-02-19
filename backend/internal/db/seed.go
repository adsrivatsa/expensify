package db

import (
	"context"

	"expensify/internal/models"
)

var defaultCategories = []models.Category{
	{Name: "Food & Dining", Icon: "🍕", Color: "#FF6B6B", IsDefault: true},
	{Name: "Transportation", Icon: "🚗", Color: "#4ECDC4", IsDefault: true},
	{Name: "Shopping", Icon: "🛍️", Color: "#45B7D1", IsDefault: true},
	{Name: "Entertainment", Icon: "🎬", Color: "#96CEB4", IsDefault: true},
	{Name: "Health & Medical", Icon: "🏥", Color: "#FFEAA7", IsDefault: true},
	{Name: "Utilities", Icon: "⚡", Color: "#DDA0DD", IsDefault: true},
	{Name: "Housing", Icon: "🏠", Color: "#98D8C8", IsDefault: true},
	{Name: "Personal Care", Icon: "💆", Color: "#F7D794", IsDefault: true},
	{Name: "Education", Icon: "📚", Color: "#A29BFE", IsDefault: true},
	{Name: "Travel", Icon: "✈️", Color: "#FD79A8", IsDefault: true},
	{Name: "Gifts & Donations", Icon: "🎁", Color: "#55EFC4", IsDefault: true},
	{Name: "Other", Icon: "📦", Color: "#B2BEC3", IsDefault: true},
}

// SeedDefaultCategories inserts the built-in categories if none exist yet.
func SeedDefaultCategories(ctx context.Context, repo CategoryRepository) error {
	existing, err := repo.FindDefaultCategories(ctx)
	if err != nil {
		return err
	}
	if len(existing) > 0 {
		return nil
	}

	for i := range defaultCategories {
		cat := defaultCategories[i] // copy to avoid mutating the package-level slice
		if _, err := repo.Create(ctx, &cat); err != nil {
			return err
		}
	}
	return nil
}
