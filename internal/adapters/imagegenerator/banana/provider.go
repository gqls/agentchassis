// FILE: internal/adapters/imagegenerator/banana/provider.go
//
// Banana (Gemini Image) provider — implements provider.Provider for
// Google's Nano Banana family (gemini-2.5-flash-image,
// gemini-3.1-flash-image-preview, gemini-3-pro-image-preview).
//
// What this provider does:
//   - Translate provider.Request to Gemini's contents/parts wire format
//   - Map our semantic aspect ratio labels to Gemini's supported set
//   - Fetch reference images via the injected provider.ReferenceFetcher
//     and base64-encode them for inlineData parts
//   - Surface safety blocks as provider.ErrSafetyBlocked
//
// Construction takes config from the parent imagegenerator adapter
// (which reads env vars). This file does NOT call os.Getenv — that's
// the parent adapter's job.
//
// NegativePrompt on provider.Request is ignored here (Gemini has no
// negative-prompt concept). Provider logs at debug level if one is
// provided; callers shouldn't rely on it being honoured.

package banana

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"go.uber.org/zap"

	"github.com/gqls/agentchassis/internal/adapters/imagegenerator/banana/api"
	"github.com/gqls/agentchassis/internal/adapters/imagegenerator/provider"
)

// ProviderName is the stable identifier used by the action layer when
// routing kinds to providers.
const ProviderName = "banana"

// ─── Config ────────────────────────────────────────────────────────────────

// Config holds construction-time parameters. Build it in the parent
// imagegenerator adapter from env vars, then pass to New().
type Config struct {
	// APIKey is the Gemini API key (sent in the x-goog-api-key header).
	// Required; empty → New() errors.
	// Parent adapter reads from BANANA_API_KEY env (new env var).
	APIKey string

	// BaseURL overrides api.DefaultBaseURL. Empty = use the default.
	BaseURL string

	// DefaultModel is the model used for every call (no per-request
	// override today). Recommend api.ModelNanoBananaPro for best
	// instruction-following on flat-illustration prompts.
	DefaultModel string
}

// ─── Provider ──────────────────────────────────────────────────────────────

// Provider is the provider.Provider implementation for Google Gemini Image.
type Provider struct {
	cfg     Config
	client  *api.Client
	logger  *zap.Logger
	fetcher provider.ReferenceFetcher // may be nil if reference URIs never used
}

// Compile-time check: Provider satisfies provider.Provider.
var _ provider.Provider = (*Provider)(nil)

// New constructs a Provider.
//
// fetcher may be nil if the caller is sure no request will include
// reference URIs. If a request DOES include URIs and fetcher is nil,
// GenerateImage returns an error rather than silently dropping them.
// Recommend always passing a fetcher — wiring is cheap.
func New(cfg Config, fetcher provider.ReferenceFetcher, logger *zap.Logger) (*Provider, error) {
	if cfg.APIKey == "" {
		return nil, errors.New("banana provider: APIKey required")
	}
	if cfg.DefaultModel == "" {
		cfg.DefaultModel = api.ModelNanoBananaPro
	}
	return &Provider{
		cfg:     cfg,
		client:  api.NewClient(cfg.BaseURL, cfg.APIKey, logger),
		logger:  logger.Named("banana_provider"),
		fetcher: fetcher,
	}, nil
}

// Name implements provider.Provider.
func (p *Provider) Name() string { return ProviderName }

// GenerateImage implements provider.Provider.
func (p *Provider) GenerateImage(ctx context.Context, req provider.Request) (*provider.Result, error) {
	if req.Prompt == "" {
		return nil, fmt.Errorf("%w: prompt required", provider.ErrInvalidRequest)
	}

	// Gemini has no negative-prompt concept. Log + ignore if provided so
	// callers don't think it's being honoured.
	if req.NegativePrompt != "" {
		p.logger.Debug("NegativePrompt provided but Banana ignores it",
			zap.String("negative_prompt_preview", truncate(req.NegativePrompt, 80)))
	}

	aspectRatio := mapAspectRatio(req.AspectRatio, p.logger)

	refs, err := p.fetchAndEncodeReferences(ctx, req.ReferenceImageURIs)
	if err != nil {
		return nil, fmt.Errorf("banana provider: fetch references: %w", err)
	}

	apiReq := api.GenerateRequest{
		Model:           p.cfg.DefaultModel,
		Prompt:          req.Prompt,
		AspectRatio:     aspectRatio,
		ReferenceImages: refs,
	}

	apiResp, err := p.client.GenerateImage(ctx, apiReq)
	if err != nil {
		// Translate API-layer safety block to provider sentinel so the
		// action can errors.Is against a single canonical value.
		if errors.Is(err, api.ErrSafetyBlocked) {
			return nil, provider.ErrSafetyBlocked
		}
		return nil, fmt.Errorf("banana provider: generate: %w", err)
	}

	if len(apiResp.Images) == 0 {
		// Defensive — client.go already checks this case, but explicit
		// branch protects against future client changes.
		return nil, errors.New("banana provider: no images in successful response")
	}

	img := apiResp.Images[0]
	raw, err := base64.StdEncoding.DecodeString(img.Base64Data)
	if err != nil {
		return nil, fmt.Errorf("banana provider: decode base64: %w", err)
	}

	p.logger.Info("Banana image generation succeeded",
		zap.String("model", p.cfg.DefaultModel),
		zap.String("kind", req.Kind),
		zap.String("aspect_ratio", aspectRatio),
		zap.Int("reference_count", len(refs)),
		zap.Int("bytes", len(raw)),
	)

	return &provider.Result{
		ImageBytes:   raw,
		MimeType:     img.MimeType,
		ProviderName: p.Name(),
		ModelID:      p.cfg.DefaultModel,
	}, nil
}

// ─── Internal helpers ──────────────────────────────────────────────────────

// mapAspectRatio translates the planner's semantic labels (from
// style_hints.aspect_ratio) to Gemini's supported set.
//
// All ten labels in api.AspectRatio* are passed through verbatim.
// Unknown labels return "" (Gemini falls back to 1:1 default or matches
// the first reference image). Logs a debug note on unknown input so
// silent fallthrough is visible.
func mapAspectRatio(label string, logger *zap.Logger) string {
	if label == "" {
		return ""
	}
	label = strings.TrimSpace(label)
	switch label {
	case api.AspectRatio1to1,
		api.AspectRatio16to9, api.AspectRatio9to16,
		api.AspectRatio4to3, api.AspectRatio3to4,
		api.AspectRatio3to2, api.AspectRatio2to3,
		api.AspectRatio21to9, api.AspectRatio9to21,
		api.AspectRatio5to4:
		return label
	default:
		logger.Debug("Aspect ratio label not in Gemini's set; sending no aspect_ratio",
			zap.String("label", label))
		return ""
	}
}

// fetchAndEncodeReferences fetches each URI via the injected
// ReferenceFetcher and base64-encodes the bytes for inlineData parts.
//
// Returns nil slice (not error) when no URIs are provided. Returns an
// error if URIs are provided but fetcher is nil — surfaces wiring bugs
// loudly rather than dropping references silently.
func (p *Provider) fetchAndEncodeReferences(ctx context.Context, uris []string) ([]api.ReferenceImage, error) {
	if len(uris) == 0 {
		return nil, nil
	}
	if p.fetcher == nil {
		return nil, errors.New("banana provider: reference URIs provided but no ReferenceFetcher injected at New()")
	}
	out := make([]api.ReferenceImage, 0, len(uris))
	for _, uri := range uris {
		raw, mime, err := p.fetcher.Fetch(ctx, uri)
		if err != nil {
			return nil, fmt.Errorf("fetch %s: %w", uri, err)
		}
		out = append(out, api.ReferenceImage{
			MimeType:   mime,
			Base64Data: base64.StdEncoding.EncodeToString(raw),
		})
	}
	return out, nil
}

// truncate caps a string for log preview. Used for the negative-prompt
// debug log line to avoid printing huge text blobs.
func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "..."
}
