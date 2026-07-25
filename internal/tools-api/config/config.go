package config

import (
	"errors"
	"os"
	"strconv"
)

// Config holds all runtime configuration for the tools-api service.
type Config struct {
	Port            string
	DatabaseURL     string
	AnthropicAPIKey string
	Model           string
	MaxBodyBytes    int
	RateLimitRPS    float64
	RateLimitBurst  int
}

// Load reads configuration from environment variables and returns errors
// instead of panicking on required-but-missing variables.
func Load() (*Config, error) {
	cfg := &Config{}

	cfg.Port = getEnvDefault("PORT", "8080")

	cfg.DatabaseURL = os.Getenv("DATABASE_URL")
	if cfg.DatabaseURL == "" {
		return nil, errors.New("DATABASE_URL is required but not set")
	}

	cfg.AnthropicAPIKey = os.Getenv("ANTHROPIC_API_KEY")
	if cfg.AnthropicAPIKey == "" {
		return nil, errors.New("ANTHROPIC_API_KEY is required but not set")
	}

	cfg.Model = getEnvDefault("GAUNTLET_MODEL", "claude-sonnet-5")

	maxInputChars, err := strconv.Atoi(getEnvDefault("MAX_INPUT_CHARS", "2000"))
	if err != nil {
		return nil, errors.New("MAX_INPUT_CHARS must be an integer")
	}
	cfg.MaxBodyBytes = maxInputChars

	rateLimitRPS, err := strconv.ParseFloat(getEnvDefault("RATE_LIMIT_RPS", "1.0"), 64)
	if err != nil {
		return nil, errors.New("RATE_LIMIT_RPS must be a float")
	}
	cfg.RateLimitRPS = rateLimitRPS

	rateLimitBurst, err := strconv.Atoi(getEnvDefault("RATE_LIMIT_BURST", "5"))
	if err != nil {
		return nil, errors.New("RATE_LIMIT_BURST must be an integer")
	}
	cfg.RateLimitBurst = rateLimitBurst

	return cfg, nil
}

func getEnvDefault(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}
