package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"

	"github.com/gqls/agentchassis/platform/mailer"
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

	// Gripper is nil unless GRIPPER_ANTHROPIC_API_KEY is set. When nil the
	// gripper route group is not mounted at all, so an island .env that
	// predates the feature boots exactly as it did before (the routes 404
	// through Caddy the way every unregistered path does). Setting the key is
	// the opt-in; once opted in, the rest of the group's config is REQUIRED
	// and a gap fails the process loudly at start, matching how this file
	// treats DATABASE_URL and ANTHROPIC_API_KEY.
	Gripper *GripperConfig
}

// GripperConfig is the gripper-dossier route group's own configuration. It is
// deliberately separate from the gauntlet fields above: its own spend-capped
// Anthropic key (owner ruling — never the debate engine's), its own model, its
// own body cap, and the two things the gauntlet has no use for (a pull key for
// the cluster and an SMTP identity for the poller).
type GripperConfig struct {
	// APIKeyEnvVar names the env var holding the key; aiservice reads it by
	// name (config["api_key_env_var"]) rather than being handed the value.
	APIKeyEnvVar string
	Model        string
	// PullKey is what the cluster sends as X-Internal-Key on GET /requests.
	PullKey string
	// SMTP is the mailer identity for the poller's emails.
	SMTP mailer.Config
	// DailyTurnCap bounds chat turns per UTC day across every visitor.
	DailyTurnCap int
	// MaxBodyBytes caps a gripper request body. Its own value because the
	// gauntlet cap (MAX_INPUT_CHARS, default 2000 BYTES) is smaller than one
	// maximal /chat message.
	MaxBodyBytes int
}

// Env var names for the gripper group, in one place so the compose file, the
// runbook and this loader cannot drift apart silently.
const (
	GripperAPIKeyEnv       = "GRIPPER_ANTHROPIC_API_KEY"
	GripperModelEnv        = "GRIPPER_MODEL"
	GripperPullKeyEnv      = "GRIPPER_PULL_KEY"
	GripperSMTPPrefix      = "GRIPPER_SMTP" // _HOST/_PORT/_USER/_PASS/_FROM/_FROM_NAME/_REPLY_TO via mailer.FromEnv
	GripperDailyTurnCapEnv = "GRIPPER_DAILY_TURN_CAP"
	GripperMaxBodyBytesEnv = "GRIPPER_MAX_BODY_BYTES"

	// GripperDefaultModel: DESIGN §2's chat-call spec. Haiku is the whole cost
	// envelope (~$0.15/session worst case) — do not quietly promote it.
	GripperDefaultModel        = "claude-haiku-4-5"
	GripperDefaultDailyTurnCap = 2000
	GripperDefaultMaxBodyBytes = 16384
	// gripperMinPullKeyLen mirrors seed 208's own guard on the cluster side
	// (it refuses a pull_key under 24 chars), so both ends agree on what a
	// real key looks like.
	gripperMinPullKeyLen = 24
)

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

	g, err := loadGripper()
	if err != nil {
		return nil, err
	}
	cfg.Gripper = g

	return cfg, nil
}

// loadGripper returns nil, nil when the feature is not opted in.
func loadGripper() (*GripperConfig, error) {
	if os.Getenv(GripperAPIKeyEnv) == "" {
		return nil, nil
	}
	g := &GripperConfig{
		APIKeyEnvVar: GripperAPIKeyEnv,
		Model:        getEnvDefault(GripperModelEnv, GripperDefaultModel),
		PullKey:      os.Getenv(GripperPullKeyEnv),
	}
	if len(g.PullKey) < gripperMinPullKeyLen {
		return nil, fmt.Errorf("%s is required (>= %d chars) when %s is set", GripperPullKeyEnv, gripperMinPullKeyLen, GripperAPIKeyEnv)
	}
	smtp, err := mailer.FromEnv(GripperSMTPPrefix)
	if err != nil {
		return nil, fmt.Errorf("gripper mailer: %w (required when %s is set)", err, GripperAPIKeyEnv)
	}
	if _, err := mailer.New(smtp); err != nil {
		return nil, fmt.Errorf("gripper mailer: %w", err)
	}
	g.SMTP = smtp

	g.DailyTurnCap, err = strconv.Atoi(getEnvDefault(GripperDailyTurnCapEnv, strconv.Itoa(GripperDefaultDailyTurnCap)))
	if err != nil || g.DailyTurnCap <= 0 {
		return nil, fmt.Errorf("%s must be a positive integer", GripperDailyTurnCapEnv)
	}
	g.MaxBodyBytes, err = strconv.Atoi(getEnvDefault(GripperMaxBodyBytesEnv, strconv.Itoa(GripperDefaultMaxBodyBytes)))
	if err != nil || g.MaxBodyBytes <= 0 {
		return nil, fmt.Errorf("%s must be a positive integer", GripperMaxBodyBytesEnv)
	}
	return g, nil
}

func getEnvDefault(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}
