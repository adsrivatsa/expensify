package services

import (
	"context"
	"fmt"
	"sort"
	"time"

	"expensify/internal/db"
	"expensify/internal/models"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// CreateCategoryRequest holds the fields for a new custom category.
type CreateCategoryRequest struct {
	Name  string `json:"name"`
	Icon  string `json:"icon"`
	Color string `json:"color"`
}

// CategoryService manages spending categories.
type CategoryService interface {
	// GetCategories returns all default categories plus any the user created.
	GetCategories(ctx context.Context, userID string) ([]*models.Category, error)
	CreateCategory(ctx context.Context, userID string, req CreateCategoryRequest) (*models.Category, error)
	DeleteCategory(ctx context.Context, userID string, categoryID string) error
}

type categoryService struct {
	repo   db.CategoryRepository
	txRepo db.TransactionRepository
}

// NewCategoryService creates a new CategoryService.
func NewCategoryService(repo db.CategoryRepository, txRepo db.TransactionRepository) CategoryService {
	return &categoryService{repo: repo, txRepo: txRepo}
}

func (s *categoryService) GetCategories(ctx context.Context, userID string) ([]*models.Category, error) {
	uid, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, ErrInvalidID
	}

	defaults, err := s.repo.FindDefaultCategories(ctx)
	if err != nil {
		return nil, fmt.Errorf("fetching default categories: %w", err)
	}

	custom, err := s.repo.FindByUserID(ctx, uid)
	if err != nil {
		return nil, fmt.Errorf("fetching user categories: %w", err)
	}

	all := append(defaults, custom...)
	sort.Slice(all, func(i, j int) bool {
		if all[i].Name == "Other" {
			return false
		}
		if all[j].Name == "Other" {
			return true
		}
		return all[i].Name < all[j].Name
	})
	return all, nil
}

func (s *categoryService) CreateCategory(ctx context.Context, userID string, req CreateCategoryRequest) (*models.Category, error) {
	uid, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, ErrInvalidID
	}

	cat := &models.Category{
		UserID:    &uid,
		Name:      req.Name,
		Icon:      req.Icon,
		Color:     req.Color,
		IsDefault: false,
		CreatedAt: time.Now(),
	}

	created, err := s.repo.Create(ctx, cat)
	if err != nil {
		return nil, fmt.Errorf("creating category: %w", err)
	}
	return created, nil
}

func (s *categoryService) DeleteCategory(ctx context.Context, userID string, categoryID string) error {
	uid, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return ErrInvalidID
	}
	catID, err := primitive.ObjectIDFromHex(categoryID)
	if err != nil {
		return ErrInvalidID
	}

	// Block deletion if the user has any transactions referencing this category.
	hasTransactions, err := s.txRepo.ExistsByCategoryID(ctx, uid, catID)
	if err != nil {
		return fmt.Errorf("checking category transactions: %w", err)
	}
	if hasTransactions {
		return ErrCategoryInUse
	}

	// The repo enforces ownership: it only deletes when user_id matches.
	if err := s.repo.Delete(ctx, catID, uid); err != nil {
		if err == db.ErrNotFound {
			return ErrNotFound
		}
		return fmt.Errorf("deleting category: %w", err)
	}
	return nil
}
