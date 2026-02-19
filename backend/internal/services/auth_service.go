package services

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"expensify/internal/db"
	"expensify/internal/models"

	"github.com/google/uuid"
)

const sessionDuration = 30 * 24 * time.Hour

// GoogleUserInfo holds the profile data returned by Google's userinfo endpoint.
type GoogleUserInfo struct {
	ID      string `json:"id"`
	Email   string `json:"email"`
	Name    string `json:"name"`
	Picture string `json:"picture"`
}

// AuthService handles Google OAuth and session management.
type AuthService interface {
	GetGoogleUserInfo(ctx context.Context, accessToken string) (*GoogleUserInfo, error)
	HandleCallback(ctx context.Context, info *GoogleUserInfo) (*models.Session, error)
	GetCurrentUser(ctx context.Context, token string) (*models.User, error)
	Logout(ctx context.Context, token string) error
}

type authService struct {
	userRepo    db.UserRepository
	sessionRepo db.SessionRepository
}

// NewAuthService creates a new AuthService.
func NewAuthService(userRepo db.UserRepository, sessionRepo db.SessionRepository) AuthService {
	return &authService{
		userRepo:    userRepo,
		sessionRepo: sessionRepo,
	}
}

func (s *authService) GetGoogleUserInfo(ctx context.Context, accessToken string) (*GoogleUserInfo, error) {
	resp, err := http.Get("https://www.googleapis.com/oauth2/v2/userinfo?access_token=" + accessToken)
	if err != nil {
		return nil, fmt.Errorf("fetching google user info: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("google userinfo returned status %d", resp.StatusCode)
	}

	var info GoogleUserInfo
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return nil, fmt.Errorf("decoding google user info: %w", err)
	}
	return &info, nil
}

// HandleCallback upserts the user and creates a new session.
func (s *authService) HandleCallback(ctx context.Context, info *GoogleUserInfo) (*models.Session, error) {
	user := &models.User{
		GoogleID: info.ID,
		Email:    info.Email,
		Name:     info.Name,
		Picture:  info.Picture,
	}
	saved, err := s.userRepo.Upsert(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("upserting user: %w", err)
	}

	session := &models.Session{
		UserID:    saved.ID,
		Token:     uuid.New().String(),
		ExpiresAt: time.Now().Add(sessionDuration),
	}
	created, err := s.sessionRepo.Create(ctx, session)
	if err != nil {
		return nil, fmt.Errorf("creating session: %w", err)
	}
	return created, nil
}

// GetCurrentUser validates the session token and returns the associated user.
func (s *authService) GetCurrentUser(ctx context.Context, token string) (*models.User, error) {
	session, err := s.sessionRepo.FindByToken(ctx, token)
	if err != nil {
		return nil, fmt.Errorf("finding session: %w", err)
	}
	if session == nil {
		return nil, ErrNotFound
	}
	if time.Now().After(session.ExpiresAt) {
		return nil, ErrSessionExpired
	}

	user, err := s.userRepo.FindByID(ctx, session.UserID)
	if err != nil {
		return nil, fmt.Errorf("finding user: %w", err)
	}
	if user == nil {
		return nil, ErrNotFound
	}
	return user, nil
}

func (s *authService) Logout(ctx context.Context, token string) error {
	if err := s.sessionRepo.Delete(ctx, token); err != nil {
		return fmt.Errorf("deleting session: %w", err)
	}
	return nil
}
