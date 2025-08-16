package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv" // You'll need to run: go get github.com/joho/godotenv
)

// Config holds all application-wide configuration loaded from environment variables.
type Config struct {
	DatabaseURL   string
	Auth0Domain   string
	Auth0Audience string
	AppEnv        string
	GCSBucketName string
	SentryDSN     string
	OpenAIAPIKey  string
}

// LoadConfig reads configuration from environment variables or a .env file.
// It is the single source of truth for application configuration.
func LoadConfig() (*Config, error) {
	// Load .env file if it exists. This is great for local development.
	// In production, these will be set directly in the environment.
	_ = godotenv.Load()

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		return nil, fmt.Errorf("FATAL: DATABASE_URL environment variable not set")
	}

	auth0Domain := os.Getenv("AUTH0_DOMAIN")
	if auth0Domain == "" {
		return nil, fmt.Errorf("FATAL: AUTH0_DOMAIN environment variable not set")
	}

	auth0Audience := os.Getenv("AUTH0_AUDIENCE")
	if auth0Audience == "" {
		return nil, fmt.Errorf("FATAL: AUTH0_AUDIENCE environment variable not set")
	}

	gcsBucketName := os.Getenv("GCS_BUCKET_NAME")
	if gcsBucketName == "" {
		return nil, fmt.Errorf("FATAL: GCS_BUCKET_NAME environment variable not set")
	}

	sentryDSN := os.Getenv("SENTRY_DSN")
	if sentryDSN == "" {
		return nil, fmt.Errorf("FATAL: SENTRY_DSN environment variable not set")
	}

	openAIKey := os.Getenv("OPENAI_API_KEY")
	if openAIKey == "" {
		return nil, fmt.Errorf("FATAL: OPENAI_API_KEY environment variable not set")
	}
	// AppEnv can have a default value
	appEnv := os.Getenv("APP_ENV")
	if appEnv == "" {
		appEnv = "development"
	}

	return &Config{
		DatabaseURL:   dbURL,
		Auth0Domain:   auth0Domain,
		Auth0Audience: auth0Audience,
		AppEnv:        appEnv,
		GCSBucketName: gcsBucketName,
		SentryDSN:     sentryDSN,
		OpenAIAPIKey:  openAIKey,
	}, nil
}
