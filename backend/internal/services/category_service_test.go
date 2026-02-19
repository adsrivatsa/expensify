package services_test

import (
	"context"
	"testing"

	"expensify/internal/db"
	"expensify/internal/models"
	"expensify/internal/services"
	"expensify/internal/testutil"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestCategoryService_GetCategories_CombinesDefaultAndCustom(t *testing.T) {
	userID := primitive.NewObjectID()

	defaults := []*models.Category{
		{ID: primitive.NewObjectID(), Name: "Food", IsDefault: true},
		{ID: primitive.NewObjectID(), Name: "Travel", IsDefault: true},
	}
	custom := []*models.Category{
		{ID: primitive.NewObjectID(), Name: "Coffee", UserID: &userID},
	}

	repo := &testutil.MockCategoryRepo{
		FindDefaultCategoriesFn: func(_ context.Context) ([]*models.Category, error) { return defaults, nil },
		FindByUserIDFn:          func(_ context.Context, _ primitive.ObjectID) ([]*models.Category, error) { return custom, nil },
	}

	svc := services.NewCategoryService(repo)
	cats, err := svc.GetCategories(context.Background(), userID.Hex())
	if err != nil {
		t.Fatalf("GetCategories: %v", err)
	}
	if len(cats) != 3 {
		t.Errorf("expected 3 categories (2 default + 1 custom), got %d", len(cats))
	}
}

func TestCategoryService_GetCategories_InvalidUserID(t *testing.T) {
	svc := services.NewCategoryService(&testutil.MockCategoryRepo{})
	_, err := svc.GetCategories(context.Background(), "not-an-object-id")
	if err != services.ErrInvalidID {
		t.Errorf("expected ErrInvalidID, got %v", err)
	}
}

func TestCategoryService_CreateCategory(t *testing.T) {
	userID := primitive.NewObjectID()

	repo := &testutil.MockCategoryRepo{
		CreateFn: func(_ context.Context, cat *models.Category) (*models.Category, error) {
			cat.ID = primitive.NewObjectID()
			return cat, nil
		},
	}

	svc := services.NewCategoryService(repo)
	req := services.CreateCategoryRequest{Name: "Gym", Icon: "🏋", Color: "#ff0000"}

	created, err := svc.CreateCategory(context.Background(), userID.Hex(), req)
	if err != nil {
		t.Fatalf("CreateCategory: %v", err)
	}
	if created.Name != "Gym" {
		t.Errorf("name: got %q, want Gym", created.Name)
	}
	if created.IsDefault {
		t.Error("custom category should not be marked as default")
	}
	if created.UserID == nil || *created.UserID != userID {
		t.Error("user_id should be set to the requesting user")
	}
}

func TestCategoryService_DeleteCategory_Success(t *testing.T) {
	userID := primitive.NewObjectID()
	catID := primitive.NewObjectID()
	deleted := false

	repo := &testutil.MockCategoryRepo{
		DeleteFn: func(_ context.Context, id, uid primitive.ObjectID) error {
			if id == catID && uid == userID {
				deleted = true
				return nil
			}
			return db.ErrNotFound
		},
	}

	svc := services.NewCategoryService(repo)
	if err := svc.DeleteCategory(context.Background(), userID.Hex(), catID.Hex()); err != nil {
		t.Fatalf("DeleteCategory: %v", err)
	}
	if !deleted {
		t.Error("expected repo.Delete to be called")
	}
}

func TestCategoryService_DeleteCategory_NotOwned(t *testing.T) {
	userID := primitive.NewObjectID()
	catID := primitive.NewObjectID()

	repo := &testutil.MockCategoryRepo{
		DeleteFn: func(_ context.Context, _, _ primitive.ObjectID) error { return db.ErrNotFound },
	}

	svc := services.NewCategoryService(repo)
	err := svc.DeleteCategory(context.Background(), userID.Hex(), catID.Hex())
	if err != services.ErrNotFound {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestCategoryService_DeleteCategory_InvalidIDs(t *testing.T) {
	svc := services.NewCategoryService(&testutil.MockCategoryRepo{})

	if err := svc.DeleteCategory(context.Background(), "bad", primitive.NewObjectID().Hex()); err != services.ErrInvalidID {
		t.Errorf("expected ErrInvalidID for bad userID, got %v", err)
	}
	if err := svc.DeleteCategory(context.Background(), primitive.NewObjectID().Hex(), "bad"); err != services.ErrInvalidID {
		t.Errorf("expected ErrInvalidID for bad catID, got %v", err)
	}
}
