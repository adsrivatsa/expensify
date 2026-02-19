//go:build integration

package db_test

import (
	"context"
	"testing"
	"time"

	"expensify/internal/db"
	"expensify/internal/models"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestSessionRepo_CreateAndFind(t *testing.T) {
	repo := db.NewSessionRepository(testDB(t))
	ctx := context.Background()

	session := &models.Session{
		UserID:    primitive.NewObjectID(),
		Token:     "test-token-123",
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}

	created, err := repo.Create(ctx, session)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if created.ID.IsZero() {
		t.Error("expected non-zero session ID")
	}

	found, err := repo.FindByToken(ctx, "test-token-123")
	if err != nil {
		t.Fatalf("FindByToken: %v", err)
	}
	if found == nil {
		t.Fatal("expected session, got nil")
	}
	if found.Token != session.Token {
		t.Errorf("token mismatch: got %q, want %q", found.Token, session.Token)
	}
}

func TestSessionRepo_FindByToken_Missing(t *testing.T) {
	repo := db.NewSessionRepository(testDB(t))
	ctx := context.Background()

	found, err := repo.FindByToken(ctx, "nonexistent")
	if err != nil {
		t.Fatalf("FindByToken: %v", err)
	}
	if found != nil {
		t.Error("expected nil for missing token")
	}
}

func TestSessionRepo_Delete(t *testing.T) {
	repo := db.NewSessionRepository(testDB(t))
	ctx := context.Background()

	session := &models.Session{
		UserID:    primitive.NewObjectID(),
		Token:     "delete-me",
		ExpiresAt: time.Now().Add(time.Hour),
	}
	repo.Create(ctx, session)

	if err := repo.Delete(ctx, "delete-me"); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	found, _ := repo.FindByToken(ctx, "delete-me")
	if found != nil {
		t.Error("expected session to be deleted")
	}
}

func TestSessionRepo_DeleteExpired(t *testing.T) {
	repo := db.NewSessionRepository(testDB(t))
	ctx := context.Background()

	uid := primitive.NewObjectID()
	// Create one expired and one active session.
	repo.Create(ctx, &models.Session{UserID: uid, Token: "expired", ExpiresAt: time.Now().Add(-time.Hour)})
	repo.Create(ctx, &models.Session{UserID: uid, Token: "active", ExpiresAt: time.Now().Add(time.Hour)})

	if err := repo.DeleteExpired(ctx); err != nil {
		t.Fatalf("DeleteExpired: %v", err)
	}

	expired, _ := repo.FindByToken(ctx, "expired")
	if expired != nil {
		t.Error("expected expired session to be removed")
	}
	active, _ := repo.FindByToken(ctx, "active")
	if active == nil {
		t.Error("expected active session to remain")
	}
}
