package middleware

import (
	"context"
	"errors"
	"net/http"

	"expensify/internal/models"
	"expensify/internal/services"
)

type contextKey string

const UserContextKey contextKey = "user"

// Authenticate validates the session cookie and injects the user into the request context.
// Requests without a valid, non-expired session are rejected with 401.
func Authenticate(authSvc services.AuthService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie("session")
			if err != nil {
				http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
				return
			}

			user, err := authSvc.GetCurrentUser(r.Context(), cookie.Value)
			if err != nil {
				if errors.Is(err, services.ErrSessionExpired) || errors.Is(err, services.ErrNotFound) {
					http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
					return
				}
				http.Error(w, `{"error":"internal server error"}`, http.StatusInternalServerError)
				return
			}

			ctx := context.WithValue(r.Context(), UserContextKey, user)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// UserFromContext retrieves the authenticated user from the request context.
func UserFromContext(ctx context.Context) *models.User {
	user, _ := ctx.Value(UserContextKey).(*models.User)
	return user
}
