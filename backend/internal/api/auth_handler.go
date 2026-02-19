package api

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"time"

	"expensify/internal/middleware"
	"expensify/internal/services"

	"golang.org/x/oauth2"
)

// AuthHandler handles Google OAuth and session endpoints.
type AuthHandler struct {
	authSvc       services.AuthService
	oauthCfg      *oauth2.Config
	frontendURL   string
	secureCookies bool
}

// NewAuthHandler constructs an AuthHandler.
func NewAuthHandler(authSvc services.AuthService, oauthCfg *oauth2.Config, frontendURL string, secureCookies bool) *AuthHandler {
	return &AuthHandler{
		authSvc:       authSvc,
		oauthCfg:      oauthCfg,
		frontendURL:   frontendURL,
		secureCookies: secureCookies,
	}
}

// GoogleLogin redirects the user to the Google OAuth consent screen.
func (h *AuthHandler) GoogleLogin(w http.ResponseWriter, r *http.Request) {
	state, err := generateState()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to generate state")
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "oauth_state",
		Value:    state,
		MaxAge:   300,
		HttpOnly: true,
		Secure:   h.secureCookies,
		SameSite: http.SameSiteLaxMode,
		Path:     "/",
	})

	url := h.oauthCfg.AuthCodeURL(state, oauth2.AccessTypeOnline)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

// GoogleCallback handles the redirect from Google after the user authenticates.
func (h *AuthHandler) GoogleCallback(w http.ResponseWriter, r *http.Request) {
	// Validate state to prevent CSRF.
	stateCookie, err := r.Cookie("oauth_state")
	if err != nil || stateCookie.Value != r.URL.Query().Get("state") {
		writeError(w, http.StatusBadRequest, "invalid oauth state")
		return
	}
	http.SetCookie(w, &http.Cookie{Name: "oauth_state", MaxAge: -1, Secure: h.secureCookies, Path: "/"})

	code := r.URL.Query().Get("code")
	token, err := h.oauthCfg.Exchange(r.Context(), code)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to exchange oauth code")
		return
	}

	userInfo, err := h.authSvc.GetGoogleUserInfo(r.Context(), token.AccessToken)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to fetch google profile")
		return
	}

	session, err := h.authSvc.HandleCallback(r.Context(), userInfo)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create session")
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    session.Token,
		MaxAge:   int(30 * 24 * time.Hour / time.Second),
		HttpOnly: true,
		Secure:   h.secureCookies,
		SameSite: http.SameSiteLaxMode,
		Path:     "/",
	})

	http.Redirect(w, r, h.frontendURL+"/dashboard", http.StatusTemporaryRedirect)
}

// Me returns the currently authenticated user's profile.
func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	writeJSON(w, http.StatusOK, user)
}

// Logout deletes the session and clears the cookie.
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session")
	if err == nil {
		_ = h.authSvc.Logout(r.Context(), cookie.Value)
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    "",
		MaxAge:   -1,
		Secure:   h.secureCookies,
		SameSite: http.SameSiteLaxMode,
		Path:     "/",
	})
	writeJSON(w, http.StatusOK, map[string]string{"message": "logged out"})
}

func generateState() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
