//go:build integration

package db_test

import (
	"context"
	"testing"

	"expensify/internal/db"
	"expensify/internal/models"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

func seedDefaults(t *testing.T, repo db.CategoryRepository) {
	t.Helper()
	if err := db.SeedDefaultCategories(context.Background(), repo); err != nil {
		t.Fatalf("seeding defaults: %v", err)
	}
}

func TestCategoryRepo_FindDefaultCategories(t *testing.T) {
	repo := db.NewCategoryRepository(testDB(t))
	seedDefaults(t, repo)

	cats, err := repo.FindDefaultCategories(context.Background())
	if err != nil {
		t.Fatalf("FindDefaultCategories: %v", err)
	}
	if len(cats) == 0 {
		t.Error("expected default categories, got none")
	}
	for _, c := range cats {
		if !c.IsDefault {
			t.Errorf("category %q should have is_default=true", c.Name)
		}
	}
}

func TestCategoryRepo_CreateAndFindByUserID(t *testing.T) {
	repo := db.NewCategoryRepository(testDB(t))
	ctx := context.Background()
	userID := primitive.NewObjectID()

	cat := &models.Category{
		UserID: &userID,
		Name:   "Coffee",
		Icon:   "☕",
		Color:  "#6F4E37",
	}
	created, err := repo.Create(ctx, cat)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if created.ID.IsZero() {
		t.Error("expected non-zero ID")
	}

	byUser, err := repo.FindByUserID(ctx, userID)
	if err != nil {
		t.Fatalf("FindByUserID: %v", err)
	}
	if len(byUser) != 1 || byUser[0].Name != "Coffee" {
		t.Errorf("unexpected categories: %v", byUser)
	}
}

func TestCategoryRepo_FindByID(t *testing.T) {
	repo := db.NewCategoryRepository(testDB(t))
	ctx := context.Background()
	uid := primitive.NewObjectID()

	created, _ := repo.Create(ctx, &models.Category{UserID: &uid, Name: "Gym", Icon: "🏋", Color: "#fff"})

	found, err := repo.FindByID(ctx, created.ID)
	if err != nil {
		t.Fatalf("FindByID: %v", err)
	}
	if found == nil || found.ID != created.ID {
		t.Error("FindByID returned wrong result")
	}
}

func TestCategoryRepo_FindByIDs(t *testing.T) {
	repo := db.NewCategoryRepository(testDB(t))
	ctx := context.Background()
	uid := primitive.NewObjectID()

	c1, _ := repo.Create(ctx, &models.Category{UserID: &uid, Name: "A", Icon: "a", Color: "#aaa"})
	c2, _ := repo.Create(ctx, &models.Category{UserID: &uid, Name: "B", Icon: "b", Color: "#bbb"})

	found, err := repo.FindByIDs(ctx, []primitive.ObjectID{c1.ID, c2.ID})
	if err != nil {
		t.Fatalf("FindByIDs: %v", err)
	}
	if len(found) != 2 {
		t.Errorf("expected 2 categories, got %d", len(found))
	}
}

func TestCategoryRepo_Delete_OwnedCategory(t *testing.T) {
	repo := db.NewCategoryRepository(testDB(t))
	ctx := context.Background()
	uid := primitive.NewObjectID()

	cat, _ := repo.Create(ctx, &models.Category{UserID: &uid, Name: "ToDelete", Icon: "x", Color: "#000"})

	if err := repo.Delete(ctx, cat.ID, uid); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	found, _ := repo.FindByID(ctx, cat.ID)
	if found != nil {
		t.Error("expected category to be deleted")
	}
}

func TestCategoryRepo_Delete_WrongUser(t *testing.T) {
	repo := db.NewCategoryRepository(testDB(t))
	ctx := context.Background()
	ownerID := primitive.NewObjectID()
	otherID := primitive.NewObjectID()

	cat, _ := repo.Create(ctx, &models.Category{UserID: &ownerID, Name: "Protected", Icon: "x", Color: "#000"})

	err := repo.Delete(ctx, cat.ID, otherID)
	if err != db.ErrNotFound {
		t.Errorf("expected ErrNotFound when deleting another user's category, got %v", err)
	}
}

func TestCategoryRepo_Delete_DefaultCategory(t *testing.T) {
	repo := db.NewCategoryRepository(testDB(t))
	ctx := context.Background()
	seedDefaults(t, repo)

	defaults, _ := repo.FindDefaultCategories(ctx)
	anyUID := primitive.NewObjectID()

	err := repo.Delete(ctx, defaults[0].ID, anyUID)
	if err != db.ErrNotFound {
		t.Errorf("expected ErrNotFound when deleting default category, got %v", err)
	}
}
