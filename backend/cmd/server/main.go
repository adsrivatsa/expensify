package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"expensify/internal/api"
	"expensify/internal/config"
	"expensify/internal/db"
	"expensify/internal/services"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

func main() {
	cfg := config.Load()

	// Database
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	mongoClient, err := db.Connect(ctx, cfg.MongoURI, cfg.DatabaseName)
	if err != nil {
		log.Fatalf("connecting to mongo: %v", err)
	}
	defer func() {
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer shutdownCancel()
		if err := mongoClient.Disconnect(shutdownCtx); err != nil {
			log.Printf("disconnecting from mongo: %v", err)
		}
	}()

	// Indexes
	if err := db.EnsureSessionIndexes(context.Background(), mongoClient.DB); err != nil {
		log.Printf("warning: could not ensure session indexes: %v", err)
	}
	if err := db.EnsureTransactionIndexes(context.Background(), mongoClient.DB); err != nil {
		log.Printf("warning: could not ensure transaction indexes: %v", err)
	}

	// Repositories
	userRepo := db.NewUserRepository(mongoClient.DB)
	sessionRepo := db.NewSessionRepository(mongoClient.DB)
	catRepo := db.NewCategoryRepository(mongoClient.DB)
	txRepo := db.NewTransactionRepository(mongoClient.DB)

	// Seed default categories
	if err := db.SeedDefaultCategories(context.Background(), catRepo); err != nil {
		log.Printf("warning: could not seed default categories: %v", err)
	}

	// Services
	authSvc := services.NewAuthService(userRepo, sessionRepo)
	catSvc := services.NewCategoryService(catRepo, txRepo)
	txSvc := services.NewTransactionService(txRepo, catRepo)

	// OAuth config
	oauthCfg := &oauth2.Config{
		ClientID:     cfg.GoogleClientID,
		ClientSecret: cfg.GoogleClientSecret,
		RedirectURL:  cfg.GoogleRedirectURL,
		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.email",
			"https://www.googleapis.com/auth/userinfo.profile",
		},
		Endpoint: google.Endpoint,
	}

	// Router
	router := api.NewRouter(authSvc, catSvc, txSvc, oauthCfg, cfg.FrontendURL, cfg.SecureCookies)

	// Server
	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		log.Printf("server listening on :%s", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	<-quit
	log.Println("shutting down server...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("server forced shutdown: %v", err)
	}
	log.Println("server stopped")
}
