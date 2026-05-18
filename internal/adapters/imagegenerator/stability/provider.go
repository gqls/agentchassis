// FILE: internal/adapters/imagegenerator/stability/provider.go
//
// Stability AI provider — implements provider.Provider for SDXL v1.0.
//
// Owns all SDXL-specific knowledge:
//   - The strict dimension whitelist (snap aspect ratios to valid pairs)
//   - Kind-specific dimension defaults (icon → square, hero → near-16:9)
//   - Kind-specific negative prompt defaults
//
// Construction takes config from the parent imagegenerator adapter
// (which reads env vars). This file does NOT call os.Getenv — that's
// the parent adapter's job.
//
// Reference images: the v1 REST API doesn't support reference image
// conditioning. Requests with reference URIs are logged and the
// references ignored. For style-anchored generation, route the kind to
// a provider that supports it (banana).

package stability

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"math"

	"go.uber.org/zap"

	"github.com/gqls/agentchassis/internal/adapters/imagegenerator/provider"
	"github.com/gqls/agentchassis/internal/adapters/imagegenerator/stability/api"
)

// ProviderName is the stable identifier used by the action layer when
// routing kinds to providers.
const ProviderName = "stability"

// ─── Kind defaults ─────────────────────────────────────────────────────────
//
// When the request doesn't specify AspectRatio (or supplies an unknown
// label), these dimensions apply per kind. All dimension pairs are on
// the SDXL v1.0 whitelist (api.SDXLV1Dimensions). When NegativePrompt is
// empty in the request, these per-kind defaults encode common SDXL
// failure modes for each image kind.

type kindDefault struct {
	Dimensions     [2]int
	NegativePrompt string
}

var kindDefaults = map[string]kindDefault{
	"icon": {
		Dimensions:     [2]int{1024, 1024},
		NegativePrompt: "background detail, shadows, photorealistic, 3D render, text, complex, busy, multiple subjects",
	},
	"hero": {
		Dimensions:     [2]int{1344, 768}, // closest landscape on whitelist to 16:9
		NegativePrompt: "text, watermark, logo, low quality, blurry, distorted",
	},
	"logo": {
		Dimensions:     [2]int{1024, 1024},
		NegativePrompt: "photorealistic, complex background, text artifacts, signature, watermark",
	},
	"illustration": {
		Dimensions:     [2]int{1024, 1024},
		NegativePrompt: "watermark, signature, text, low quality",
	},
	"infographic": {
		Dimensions:     [2]int{1024, 1024},
		NegativePrompt: "watermark, signature, blurry",
	},
}

// fallbackDefault applies when req.Kind is unknown. Conservative: square,
// generic negative prompt.
var fallbackDefault = kindDefault{
	Dimensions:     [2]int{1024, 1024},
	NegativePrompt: "watermark, signature, low quality",
}

// ─── Aspect ratio label → fraction ─────────────────────────────────────────
//
// Lookup from the planner's semantic labels (style_hints.aspect_ratio) to
// float ratios used by snapToSDXLWhitelist to find the closest valid pair.

var aspectRatioFractions = map[string]float64{
	"1:1":  1.0,
	"16:9": 16.0 / 9.0,
	"9:16": 9.0 / 16.0,
	"4:3":  4.0 / 3.0,
	"3:4":  3.0 / 4.0,
	"3:2":  3.0 / 2.0,
	"2:3":  2.0 / 3.0,
	"21:9": 21.0 / 9.0,
	"9:21": 9.0 / 21.0,
	"5:4":  5.0 / 4.0,
	"4:5":  4.0 / 5.0,
}

// ─── Config ────────────────────────────────────────────────────────────────

// Config holds construction-time parameters. Build it in the parent
// imagegenerator adapter from env vars, then pass to New().
type Config struct {
	// APIKey is the Stability AI bearer token. Required; empty → New() errors.
	// Parent adapter reads from STABILITY_API_KEY env (existing pattern).
	APIKey string

	// BaseURL overrides api.DefaultBaseURL (testing / staging). Empty = use
	// api.DefaultBaseURL. Parent adapter can pass STABILITY_API_ENDPOINT.
	BaseURL string

	// Engine selects the SDXL variant. Empty = api.EngineSDXL10.
	Engine string

	// Steps is the diffusion step count (10..150). 0 = 30.
	Steps int

	// CFGScale is the classifier-free guidance scale (0..35). 0 = 7.
	CFGScale float64

	// Samples is the number of artifacts per request (1..10). 0 = 1.
	// The provider currently surfaces only the first artifact; higher
	// values cost more without benefit unless the action layer is
	// extended to evaluate multiple candidates.
	Samples int
}

// ─── Provider ──────────────────────────────────────────────────────────────

// Provider is the provider.Provider implementation for Stability AI.
type Provider struct {
	cfg    Config
	client *api.Client
	logger *zap.Logger
}

// Compile-time check: Provider satisfies provider.Provider.
var _ provider.Provider = (*Provider)(nil)

// New constructs a Provider and validates required config. Applies
// defaults for Engine/Steps/CFGScale/Samples.
func New(cfg Config, logger *zap.Logger) (*Provider, error) {
	if cfg.APIKey == "" {
		return nil, errors.New("stability provider: APIKey required")
	}
	if cfg.Engine == "" {
		cfg.Engine = api.EngineSDXL10
	}
	if cfg.Steps == 0 {
		cfg.Steps = 30
	}
	if cfg.CFGScale == 0 {
		cfg.CFGScale = 7
	}
	if cfg.Samples == 0 {
		cfg.Samples = 1
	}
	return &Provider{
		cfg:    cfg,
		client: api.NewClient(cfg.BaseURL, cfg.APIKey, logger),
		logger: logger.Named("stability_provider"),
	}, nil
}

// Name implements provider.Provider.
func (p *Provider) Name() string { return ProviderName }

// GenerateImage implements provider.Provider.
func (p *Provider) GenerateImage(ctx context.Context, req provider.Request) (*provider.Result, error) {
	if req.Prompt == "" {
		return nil, fmt.Errorf("%w: prompt required", provider.ErrInvalidRequest)
	}

	// Resolve dimensions + negative prompt from kind defaults, allowing
	// the caller's request to override.
	defaults, known := kindDefaults[req.Kind]
	if !known {
		defaults = fallbackDefault
		p.logger.Debug("Unknown kind, using fallback default",
			zap.String("kind", req.Kind))
	}

	width, height := p.resolveDimensions(req.AspectRatio, defaults.Dimensions)
	negPrompt := req.NegativePrompt
	if negPrompt == "" {
		negPrompt = defaults.NegativePrompt
	}

	if len(req.ReferenceImageURIs) > 0 {
		p.logger.Warn("Stability v1 REST API does not support reference images; ignoring",
			zap.Int("reference_count", len(req.ReferenceImageURIs)),
			zap.String("kind", req.Kind))
	}

	prompts := []api.TextPrompt{
		{Text: req.Prompt, Weight: 1.0},
	}
	if negPrompt != "" {
		prompts = append(prompts, api.TextPrompt{
			Text:   negPrompt,
			Weight: -1.0,
		})
	}

	apiReq := api.TextToImageRequest{
		TextPrompts: prompts,
		Width:       width,
		Height:      height,
		Samples:     p.cfg.Samples,
		Steps:       p.cfg.Steps,
		CFGScale:    p.cfg.CFGScale,
	}

	resp, err := p.client.TextToImage(ctx, p.cfg.Engine, apiReq)
	if err != nil {
		return nil, fmt.Errorf("stability provider: %w", err)
	}
	if len(resp.Artifacts) == 0 {
		return nil, errors.New("stability provider: no artifacts in response")
	}

	// Samples=1 by default, so one artifact comes back. If Samples > 1,
	// the provider still returns only the first — provider.Result is
	// single-image. Extending to multi-image would require widening
	// Result and the action layer's handling.
	art := resp.Artifacts[0]
	switch art.FinishReason {
	case api.FinishReasonSuccess:
		// fall through to decode
	case api.FinishReasonContentFiltered:
		return nil, provider.ErrSafetyBlocked
	case api.FinishReasonError:
		return nil, errors.New("stability provider: artifact finish_reason=ERROR")
	default:
		return nil, fmt.Errorf("stability provider: unrecognised finish_reason %q", art.FinishReason)
	}

	raw, err := base64.StdEncoding.DecodeString(art.Base64)
	if err != nil {
		return nil, fmt.Errorf("stability provider: decode base64: %w", err)
	}

	p.logger.Info("Stability image generation succeeded",
		zap.String("engine", p.cfg.Engine),
		zap.String("kind", req.Kind),
		zap.Int("width", width),
		zap.Int("height", height),
		zap.Uint32("seed", art.Seed),
		zap.Int("bytes", len(raw)),
	)

	return &provider.Result{
		ImageBytes:   raw,
		MimeType:     "image/png", // SDXL v1 always returns PNG
		ProviderName: p.Name(),
		ModelID:      p.cfg.Engine,
	}, nil
}

// ─── Dimension resolution ──────────────────────────────────────────────────

// resolveDimensions returns (width, height) for the given aspect ratio
// label and kind default.
//
//   - If aspectRatio is "", returns kindDefault directly.
//   - If aspectRatio matches a known label, snaps to the closest pair on
//     api.SDXLV1Dimensions.
//   - If aspectRatio is unrecognised, logs a warning and uses kindDefault.
func (p *Provider) resolveDimensions(aspectRatio string, kindDefault [2]int) (int, int) {
	if aspectRatio == "" {
		return kindDefault[0], kindDefault[1]
	}
	target, ok := aspectRatioFractions[aspectRatio]
	if !ok {
		p.logger.Warn("Unknown aspect ratio label, using kind default",
			zap.String("aspect_ratio", aspectRatio),
			zap.Int("default_width", kindDefault[0]),
			zap.Int("default_height", kindDefault[1]))
		return kindDefault[0], kindDefault[1]
	}
	return snapToSDXLWhitelist(target)
}

// snapToSDXLWhitelist returns the dimension pair from api.SDXLV1Dimensions
// whose aspect ratio is closest to the target ratio.
//
// The whitelist includes both landscape and portrait entries, so
// orientation is preserved naturally (target < 1.0 picks a portrait
// entry; target > 1.0 picks landscape).
func snapToSDXLWhitelist(target float64) (int, int) {
	// Initial guess: square (centre of the ratio space).
	bestW, bestH := 1024, 1024
	bestDiff := math.Abs(target - 1.0)

	for _, d := range api.SDXLV1Dimensions {
		ar := float64(d[0]) / float64(d[1])
		diff := math.Abs(target - ar)
		if diff < bestDiff {
			bestDiff = diff
			bestW, bestH = d[0], d[1]
		}
	}
	return bestW, bestH
}
