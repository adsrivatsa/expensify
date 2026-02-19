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
	{Name: "Interest", Icon: "🏦", Color: "#74B9FF", IsDefault: true},
	{Name: "Dividends", Icon: "📈", Color: "#00B894", IsDefault: true},
	{Name: "Investment Sales", Icon: "💹", Color: "#6C5CE7", IsDefault: true},
	{Name: "Other", Icon: "📦", Color: "#B2BEC3", IsDefault: true},
}

// SeedDefaultCategories inserts any built-in categories not yet in the database,
// so adding new entries to defaultCategories is safe on existing deployments.
func SeedDefaultCategories(ctx context.Context, repo CategoryRepository) error {
	existing, err := repo.FindDefaultCategories(ctx)
	if err != nil {
		return err
	}

	present := make(map[string]bool, len(existing))
	for _, c := range existing {
		present[c.Name] = true
	}

	for i := range defaultCategories {
		cat := defaultCategories[i]
		if present[cat.Name] {
			continue
		}
		if _, err := repo.Create(ctx, &cat); err != nil {
			return err
		}
	}
	return nil
}
