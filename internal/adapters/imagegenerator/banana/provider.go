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
// NegativePrompt on provider.Request has no wire-format equivalent here
// (Gemini takes no negative-prompt parameter), so this provider TRANSLATES
// it into an explicit prohibition clause on the positive prompt rather than
// dropping it — see foldNegativeIntoPrompt.
//
// bugs_open/028: it used to drop it, logging at Debug. That made every
// `avoid` list in every site's imagery_style_guide inert the moment a kind
// moved to Banana, and since bugs_open/011 every declared kind is on Banana.
// The list was still assembled, logged, and shipped over Kafka — it just
// died here, at Debug level, where nobody reads it. Four of nine sampled
// gamesdesign.co.uk content heroes then violated their own avoid list in
// its own terms (numerals rendered, near-white grounds), and a style-guide
// edit that coincided with a lucky re-roll was written up in three places
// as a "hard-won fact" about how avoid lists work. A constraint a caller
// can set, that the platform silently discards, is worse than one it cannot
// set at all: it manufactures false beliefs about the system.

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

	// bugs_open/028 — Gemini takes no negative-prompt parameter, so fold the
	// caller's constraint into the positive prompt as an explicit prohibition
	// rather than discarding it. Logged at Info, not Debug: this rewrites what
	// the model is asked for, and the previous Debug line is precisely why the
	// discard went unnoticed for a release cycle.
	prompt := req.Prompt
	if req.NegativePrompt != "" {
		prompt = foldNegativeIntoPrompt(prompt, req.NegativePrompt)
		p.logger.Info("Banana: folded NegativePrompt into positive prompt as a prohibition clause",
			zap.String("kind", req.Kind),
			// Wide preview on purpose: this line is the evidence trail. The
			// stored assets.origin_prompt is written by the ACTION layer and so
			// records the prompt before this fold — it will never show these
			// terms, and a thread checking there would wrongly conclude the
			// avoid list is still inert. Real avoid lists run to ~350 chars.
			zap.String("negative_prompt", truncate(req.NegativePrompt, 400)),
			zap.Int("prompt_len_before", len(req.Prompt)),
			zap.Int("prompt_len_after", len(prompt)))
	}

	aspectRatio := mapAspectRatio(req.AspectRatio, p.logger)

	refs, err := p.fetchAndEncodeReferences(ctx, req.ReferenceImageURIs)
	if err != nil {
		return nil, fmt.Errorf("banana provider: fetch references: %w", err)
	}

	apiReq := api.GenerateRequest{
		Model:           p.cfg.DefaultModel,
		Prompt:          prompt,
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

// foldNegativeIntoPrompt expresses a negative prompt as an explicit
// prohibition clause appended to the positive prompt (bugs_open/028).
//
// WHY A TRAILING CLAUSE, AND WHY HERE. Gemini is instruction-tuned and does
// follow a plainly-worded prohibition; it simply has no parameter to put one
// in. The translation belongs in this provider and not in the action layer
// for three reasons:
//
//   - It is a fact about Gemini, so it lives with the code that knows Gemini.
//     The action layer would have to ask "which provider will serve this?",
//     and the only honest answer lives in routing.go — a second hand-maintained
//     provider list in a second package is the exact drift shape routing.go's
//     own header was written to stop.
//   - Doing it here covers every caller at once: the style guide's `avoid`,
//     every kindDefaults.NegativePrompt, and the input_data.constraints
//     folding, for every kind, with no further edits.
//   - It cannot disturb the palette. The action layer's direction is capped at
//     maxImageryDirectionInPrompt (200) and the palette is composed LAST, so
//     appending avoid terms up there would silently push the brand colours off
//     the end — that is bugs_open/027 §4b, a live defect. By the time a prompt
//     reaches this function the cap has already been applied, so nothing this
//     function adds can evict anything.
//
// The clause says "contain or use" deliberately: avoid lists mix objects
// ("text", "people", "watermarks") with styles ("photorealism", "gradients",
// "3D rendering"), and "do not include" reads as a no-op against a style.
//
// NOTE this is a translation, not a guarantee. A prohibition in the positive
// prompt is a softer instrument than SDXL's true negative conditioning, and
// on this evidence Gemini honours it imperfectly — the same nine samples that
// exposed this bug had "near-black" IN the positive prompt and still came back
// pale. Verify by counting violations across a set, never on one image.
func foldNegativeIntoPrompt(prompt, negative string) string {
	prompt = strings.TrimSpace(prompt)
	// Trailing separators are common in assembled avoid lists (the action
	// layer joins guide `avoid` onto kind defaults with ", ").
	negative = strings.Trim(strings.TrimSpace(negative), ",;. ")
	if negative == "" {
		return prompt
	}
	if prompt == "" {
		// Defensive: GenerateImage rejects an empty prompt before calling
		// this, so a bare constraint clause can never reach the model.
		return prompt
	}
	sep := " "
	if !endsWithSentenceBoundary(prompt) {
		sep = ". "
	}
	return prompt + sep + "Do not depict any of the following — the image must not contain or use: " + negative + "."
}

// endsWithSentenceBoundary reports whether s ends with ., !, or ?. Keeps the
// folded clause from producing "...in the image.. Do not depict".
func endsWithSentenceBoundary(s string) bool {
	if s == "" {
		return false
	}
	switch s[len(s)-1] {
	case '.', '!', '?':
		return true
	}
	return false
}

// truncate caps a string for log preview. Used for the negative-prompt
// log line to avoid printing huge text blobs.
func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "..."
}
