package api

import (
	"expensify/internal/middleware"
	"expensify/internal/services"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"golang.org/x/oauth2"
)

// NewRouter builds and returns the fully configured chi router.
func NewRouter(
	authSvc services.AuthService,
	catSvc services.CategoryService,
	txSvc services.TransactionService,
	oauthCfg *oauth2.Config,
	frontendURL string,
	secureCookies bool,
) *chi.Mux {
	r := chi.NewRouter()

	// Global middleware
	r.Use(chimiddleware.Logger)
	r.Use(chimiddleware.Recoverer)
	r.Use(chimiddleware.RequestID)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{frontendURL},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Content-Type"},
		AllowCredentials: true,
		MaxAge:           86400, // cache preflight for 24 h
	}))

	authHandler := NewAuthHandler(authSvc, oauthCfg, frontendURL, secureCookies)
	catHandler := NewCategoryHandler(catSvc)
	txHandler := NewTransactionHandler(txSvc)

	// Public auth routes
	r.Route("/auth", func(r chi.Router) {
		r.Get("/google", authHandler.GoogleLogin)
		r.Get("/google/callback", authHandler.GoogleCallback)
		r.Post("/logout", authHandler.Logout)
	})

	// Protected routes
	r.Group(func(r chi.Router) {
		r.Use(middleware.Authenticate(authSvc))

		r.Get("/auth/me", authHandler.Me)

		r.Route("/api/categories", func(r chi.Router) {
			r.Get("/", catHandler.List)
			r.Post("/", catHandler.Create)
			r.Delete("/{id}", catHandler.Delete)
		})

		r.Route("/api/transactions", func(r chi.Router) {
			r.Get("/", txHandler.List)
			r.Post("/", txHandler.Create)
			r.Put("/{id}", txHandler.Update)
			r.Delete("/{id}", txHandler.Delete)
		})

		r.Get("/api/cashflow/summary", txHandler.Summary)
	})

	return r
}
