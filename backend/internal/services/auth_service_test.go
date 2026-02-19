package services_test

import (
	"context"
	"testing"
	"time"

	"expensify/internal/models"
	"expensify/internal/services"
	"expensify/internal/testutil"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

func newAuthSvc(userRepo *testutil.MockUserRepo, sessionRepo *testutil.MockSessionRepo) services.AuthService {
	return services.NewAuthService(userRepo, sessionRepo)
}

func TestAuthService_HandleCallback_NewUser(t *testing.T) {
	userID := primitive.NewObjectID()
	userRepo := &testutil.MockUserRepo{
		UpsertFn: func(_ context.Context, u *models.User) (*models.User, error) {
			u.ID = userID
			return u, nil
		},
	}
	sessionRepo := &testutil.MockSessionRepo{
		CreateFn: func(_ context.Context, s *models.Session) (*models.Session, error) {
			s.ID = primitive.NewObjectID()
			return s, nil
		},
	}

	svc := newAuthSvc(userRepo, sessionRepo)
	info := &services.GoogleUserInfo{
		ID:      "google-1",
		Email:   "alice@example.com",
		Name:    "Alice",
		Picture: "https://example.com/pic.jpg",
	}

	session, err := svc.HandleCallback(context.Background(), info)
	if err != nil {
		t.Fatalf("HandleCallback: %v", err)
	}
	if session == nil {
		t.Fatal("expected a session, got nil")
	}
	if session.Token == "" {
		t.Error("expected a non-empty session token")
	}
	if session.UserID != userID {
		t.Errorf("session.UserID: got %v, want %v", session.UserID, userID)
	}
}

func TestAuthService_HandleCallback_ExistingUser(t *testing.T) {
	existingID := primitive.NewObjectID()
	callCount := 0

	userRepo := &testutil.MockUserRepo{
		UpsertFn: func(_ context.Context, u *models.User) (*models.User, error) {
			callCount++
			u.ID = existingID
			return u, nil
		},
	}
	sessionRepo := &testutil.MockSessionRepo{
		CreateFn: func(_ context.Context, s *models.Session) (*models.Session, error) {
			s.ID = primitive.NewObjectID()
			return s, nil
		},
	}

	svc := newAuthSvc(userRepo, sessionRepo)
	info := &services.GoogleUserInfo{ID: "google-2", Email: "bob@example.com", Name: "Bob"}

	// Call twice to simulate returning user.
	svc.HandleCallback(context.Background(), info)
	session, err := svc.HandleCallback(context.Background(), info)
	if err != nil {
		t.Fatalf("second HandleCallback: %v", err)
	}
	if session.UserID != existingID {
		t.Errorf("session.UserID mismatch: got %v, want %v", session.UserID, existingID)
	}
	if callCount != 2 {
		t.Errorf("Upsert should have been called twice, called %d times", callCount)
	}
}

func TestAuthService_GetCurrentUser_Valid(t *testing.T) {
	userID := primitive.NewObjectID()
	expectedUser := &models.User{ID: userID, Email: "carol@example.com"}

	sessionRepo := &testutil.MockSessionRepo{
		FindByTokenFn: func(_ context.Context, token string) (*models.Session, error) {
			return &models.Session{
				UserID:    userID,
				Token:     token,
				ExpiresAt: time.Now().Add(time.Hour),
			}, nil
		},
	}
	userRepo := &testutil.MockUserRepo{
		FindByIDFn: func(_ context.Context, id primitive.ObjectID) (*models.User, error) {
			if id == userID {
				return expectedUser, nil
			}
			return nil, nil
		},
	}

	svc := newAuthSvc(userRepo, sessionRepo)
	user, err := svc.GetCurrentUser(context.Background(), "valid-token")
	if err != nil {
		t.Fatalf("GetCurrentUser: %v", err)
	}
	if user.Email != expectedUser.Email {
		t.Errorf("email: got %q, want %q", user.Email, expectedUser.Email)
	}
}

func TestAuthService_GetCurrentUser_ExpiredSession(t *testing.T) {
	sessionRepo := &testutil.MockSessionRepo{
		FindByTokenFn: func(_ context.Context, token string) (*models.Session, error) {
			return &models.Session{
				Token:     token,
				ExpiresAt: time.Now().Add(-time.Hour), // already expired
			}, nil
		},
	}

	svc := newAuthSvc(&testutil.MockUserRepo{}, sessionRepo)
	_, err := svc.GetCurrentUser(context.Background(), "expired-token")
	if err != services.ErrSessionExpired {
		t.Errorf("expected ErrSessionExpired, got %v", err)
	}
}

func TestAuthService_GetCurrentUser_MissingSession(t *testing.T) {
	sessionRepo := &testutil.MockSessionRepo{
		FindByTokenFn: func(_ context.Context, _ string) (*models.Session, error) {
			return nil, nil
		},
	}

	svc := newAuthSvc(&testutil.MockUserRepo{}, sessionRepo)
	_, err := svc.GetCurrentUser(context.Background(), "ghost-token")
	if err != services.ErrNotFound {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestAuthService_Logout(t *testing.T) {
	deleted := false
	sessionRepo := &testutil.MockSessionRepo{
		DeleteFn: func(_ context.Context, token string) error {
			if token == "my-token" {
				deleted = true
			}
			return nil
		},
	}

	svc := newAuthSvc(&testutil.MockUserRepo{}, sessionRepo)
	if err := svc.Logout(context.Background(), "my-token"); err != nil {
		t.Fatalf("Logout: %v", err)
	}
	if !deleted {
		t.Error("expected session to be deleted")
	}
}
