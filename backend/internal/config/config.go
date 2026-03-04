// Package config provides configuration for the application
package config

import (
	"os"
)

// Config holds the application configuration
type Config struct {
	DatabaseURL      string
	RedisURL         string
	JWTSecret        string
	Port             string
	AWSSaaSAccountID string
	AWSRegion        string
	LLMAPIKey        string
	LLMBaseURL       string
	LLMModel         string
}

// Load loads the configuration from environment variables
func Load() *Config {
	return &Config{
		DatabaseURL:      getEnv("DATABASE_URL", "postgres://finops:finops_secret@localhost:5432/finops?sslmode=disable"),
		RedisURL:         getEnv("REDIS_URL", "redis://localhost:6379/0"),
		JWTSecret:        getEnv("JWT_SECRET", "dev-secret-change-in-production"),
		Port:             getEnv("SERVER_PORT", "8080"),
		AWSSaaSAccountID: getEnv("AWS_SAAS_ACCOUNT_ID", ""),
		AWSRegion:        getEnv("AWS_REGION", "us-east-1"),
		LLMAPIKey:        getEnv("LLM_API_KEY", ""),
		LLMBaseURL:       getEnv("LLM_BASE_URL", "https://api.openai.com/v1"),
		LLMModel:         getEnv("LLM_MODEL", "gpt-4o-mini"),
	}
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}
