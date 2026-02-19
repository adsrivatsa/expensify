//go:build integration

package db_test

import (
	"context"
	"testing"
	"time"

	"expensify/internal/db"
	"expensify/internal/models"
)

func TestUserRepo_Upsert_NewUser(t *testing.T) {
	repo := db.NewUserRepository(testDB(t))
	ctx := context.Background()

	user := &models.User{
		GoogleID: "google-123",
		Email:    "alice@example.com",
		Name:     "Alice",
		Picture:  "https://example.com/alice.jpg",
	}

	saved, err := repo.Upsert(ctx, user)
	if err != nil {
		t.Fatalf("Upsert: %v", err)
	}
	if saved.ID.IsZero() {
		t.Error("expected non-zero ID after upsert")
	}
	if saved.Email != user.Email {
		t.Errorf("email: got %q, want %q", saved.Email, user.Email)
	}
}

func TestUserRepo_Upsert_ExistingUser(t *testing.T) {
	repo := db.NewUserRepository(testDB(t))
	ctx := context.Background()

	user := &models.User{GoogleID: "google-456", Email: "bob@example.com", Name: "Bob"}
	first, _ := repo.Upsert(ctx, user)

	// Simulate a profile update.
	user.Name = "Bobby"
	user.Email = "bobby@example.com"
	second, err := repo.Upsert(ctx, user)
	if err != nil {
		t.Fatalf("second Upsert: %v", err)
	}
	if first.ID != second.ID {
		t.Error("expected same ID on upsert of existing user")
	}
	if second.Name != "Bobby" {
		t.Errorf("name: got %q, want Bobby", second.Name)
	}
}

func TestUserRepo_FindByGoogleID(t *testing.T) {
	repo := db.NewUserRepository(testDB(t))
	ctx := context.Background()

	repo.Upsert(ctx, &models.User{GoogleID: "gid-abc", Email: "c@example.com", Name: "Carol"})

	found, err := repo.FindByGoogleID(ctx, "gid-abc")
	if err != nil {
		t.Fatalf("FindByGoogleID: %v", err)
	}
	if found == nil {
		t.Fatal("expected user, got nil")
	}
	if found.Email != "c@example.com" {
		t.Errorf("email: got %q, want c@example.com", found.Email)
	}

	missing, err := repo.FindByGoogleID(ctx, "does-not-exist")
	if err != nil {
		t.Fatalf("FindByGoogleID missing: %v", err)
	}
	if missing != nil {
		t.Error("expected nil for missing user")
	}
}

func TestUserRepo_FindByID(t *testing.T) {
	repo := db.NewUserRepository(testDB(t))
	ctx := context.Background()

	saved, _ := repo.Upsert(ctx, &models.User{GoogleID: "gid-xyz", Email: "d@example.com", Name: "Dave"})

	found, err := repo.FindByID(ctx, saved.ID)
	if err != nil {
		t.Fatalf("FindByID: %v", err)
	}
	if found == nil || found.ID != saved.ID {
		t.Error("FindByID returned wrong or nil user")
	}
}

func TestUserRepo_FindByID_Timestamps(t *testing.T) {
	repo := db.NewUserRepository(testDB(t))
	ctx := context.Background()

	before := time.Now().Add(-time.Second)
	saved, _ := repo.Upsert(ctx, &models.User{GoogleID: "gid-ts", Email: "ts@example.com", Name: "Timmy"})

	if saved.CreatedAt.Before(before) {
		t.Error("created_at should be set to roughly now")
	}
}
