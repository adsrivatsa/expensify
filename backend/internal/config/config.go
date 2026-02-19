package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	MongoURI           string
	DatabaseName       string
	GoogleClientID     string
	GoogleClientSecret string
	GoogleRedirectURL  string
	SessionSecret      string
	FrontendURL        string
	Port               string
	SecureCookies      bool // set true in production (HTTPS)
}

func Load() *Config {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, reading from environment")
	}

	return &Config{
		MongoURI:           getEnv("MONGO_URI", "mongodb://localhost:27017"),
		DatabaseName:       getEnv("DB_NAME", "expensify"),
		GoogleClientID:     getEnv("GOOGLE_CLIENT_ID", ""),
		GoogleClientSecret: getEnv("GOOGLE_CLIENT_SECRET", ""),
		GoogleRedirectURL:  getEnv("GOOGLE_REDIRECT_URL", "http://localhost:8080/auth/google/callback"),
		SessionSecret:      getEnv("SESSION_SECRET", "change-me-in-production"),
		FrontendURL:        getEnv("FRONTEND_URL", "http://localhost:5173"),
		Port:               getEnv("PORT", "8080"),
		SecureCookies:      getEnv("SECURE_COOKIES", "") == "true",
	}
}

func getEnv(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}
