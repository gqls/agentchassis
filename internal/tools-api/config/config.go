package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

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

	// Playground is the third tool (finetuning.uk's public demo chat against a
	// self-hosted model). Same opt-in shape: nil unless PLAYGROUND_OLLAMA_URL
	// is set, and then every field is validated at start.
	Playground *PlaygroundConfig
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

	p, err := loadPlayground()
	if err != nil {
		return nil, err
	}
	cfg.Playground = p

	return cfg, nil
}

// PlaygroundConfig is the third tool: finetuning.uk's public demo chat against
// a self-hosted open-weight model (finetuning_uk_service PLAN Phase P; owner
// decision 2026-09-03: a public demo on the in-cluster Ollama plus booked GPU
// hours). Present only when PLAYGROUND_OLLAMA_URL is set — the gripper's
// opt-in shape: an unset env mounts nothing, so a deployment that has not been
// told about the demo serves exactly the routes it served before.
type PlaygroundConfig struct {
	// OllamaURL is the base URL of the model server, e.g.
	// http://ollama-adapter.ai-persona-system.svc.cluster.local:11434 (no
	// trailing slash; one is stripped if given).
	OllamaURL string
	// Model is the Ollama model name the demo answers with.
	Model string
	// MaxTokens caps num_predict per reply. The demo runs on CPU at ~14 tok/s
	// (measured 2026-09-03), so this bounds the visitor's wait as much as cost.
	MaxTokens int
	// NumCtx is the context window handed to Ollama; it bounds memory.
	NumCtx int
	// MaxBodyBytes is the request body cap for the playground routes.
	MaxBodyBytes int
}

const (
	PlaygroundOllamaURLEnv    = "PLAYGROUND_OLLAMA_URL"
	PlaygroundModelEnv        = "PLAYGROUND_MODEL"
	PlaygroundMaxTokensEnv    = "PLAYGROUND_MAX_TOKENS"
	PlaygroundNumCtxEnv       = "PLAYGROUND_NUM_CTX"
	PlaygroundMaxBodyBytesEnv = "PLAYGROUND_MAX_BODY_BYTES"

	PlaygroundDefaultModel        = "finetuning-demo"
	PlaygroundDefaultMaxTokens    = 150
	PlaygroundDefaultNumCtx       = 2048
	PlaygroundDefaultMaxBodyBytes = 8192
)

func loadPlayground() (*PlaygroundConfig, error) {
	url := strings.TrimRight(os.Getenv(PlaygroundOllamaURLEnv), "/")
	if url == "" {
		return nil, nil
	}
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		return nil, fmt.Errorf("%s must be an http(s) URL", PlaygroundOllamaURLEnv)
	}
	p := &PlaygroundConfig{
		OllamaURL: url,
		Model:     getEnvDefault(PlaygroundModelEnv, PlaygroundDefaultModel),
	}
	var err error
	if p.MaxTokens, err = positiveIntEnv(PlaygroundMaxTokensEnv, PlaygroundDefaultMaxTokens); err != nil {
		return nil, err
	}
	if p.NumCtx, err = positiveIntEnv(PlaygroundNumCtxEnv, PlaygroundDefaultNumCtx); err != nil {
		return nil, err
	}
	if p.MaxBodyBytes, err = positiveIntEnv(PlaygroundMaxBodyBytesEnv, PlaygroundDefaultMaxBodyBytes); err != nil {
		return nil, err
	}
	return p, nil
}

func positiveIntEnv(key string, def int) (int, error) {
	v, err := strconv.Atoi(getEnvDefault(key, strconv.Itoa(def)))
	if err != nil || v <= 0 {
		return 0, fmt.Errorf("%s must be a positive integer", key)
	}
	return v, nil
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
