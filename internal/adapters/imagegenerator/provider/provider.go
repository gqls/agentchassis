// FILE: internal/adapters/imagegenerator/provider/provider.go
//
// Provider interface for image-generation backends used by the
// imagegenerator adapter.
//
// Implementations live alongside this package:
//   - internal/adapters/imagegenerator/stability/  (SDXL via Stability AI)
//   - internal/adapters/imagegenerator/banana/     (Gemini Image / Nano Banana family)
//
// Routing between providers lives in the imagegenerator adapter
// (generate_image_actions.go switches kind → provider). The interface
// lets the action layer hold heterogeneous providers and swap them by
// configuration without coupling to any one API's wire format.
//
// This package exists as a SUB-package of imagegenerator (rather than as
// types in the imagegenerator package itself) to avoid an import cycle:
// imagegenerator imports stability + banana to construct them, and
// stability + banana import this package to satisfy the Provider
// interface. Both directions terminate cleanly.

package provider

import (
	"context"
	"errors"
)

// Provider is implemented by each image generation backend.
type Provider interface {
	// Name identifies the provider in logs, asset records, and routing.
	// Must be stable across versions; used as the lookup key when the
	// action layer holds a kind→provider map.
	Name() string

	// GenerateImage produces an image from the abstract domain request.
	//
	// Returns:
	//   - (*Result, nil) on success. ImageBytes is raw decoded image bytes
	//     ready for upload; MimeType, ProviderName, ModelID populated.
	//   - (nil, ErrSafetyBlocked) when the provider declined for safety.
	//     Caller MUST NOT retry the same prompt.
	//   - (nil, ErrInvalidRequest) for caller-side request errors
	//     (empty prompt, malformed params). Retry pointless.
	//   - (nil, other error) for transient/provider-side failures.
	//     Retry may help, subject to caller's retry policy.
	GenerateImage(ctx context.Context, req Request) (*Result, error)
}

// Request is the provider-agnostic domain input.
type Request struct {
	// Kind is the image purpose: "icon", "hero", "illustration", "logo",
	// "infographic". Providers MAY use Kind to apply kind-specific
	// defaults (e.g. dimensions when AspectRatio is empty, kind-tuned
	// negative prompts when NegativePrompt is empty).
	Kind string

	// Prompt is the positive text prompt. Required; empty → ErrInvalidRequest.
	Prompt string

	// NegativePrompt is the negative text prompt. Empty means "use provider
	// default for this kind".
	//
	// Every provider must HONOUR this, by its own best available means — SDXL
	// passes it as true negative conditioning; Banana has no such parameter
	// and folds it into the positive prompt as a prohibition clause. What a
	// provider must NOT do is accept it and discard it: that was bugs_open/028,
	// which made every site's `avoid` list inert fleet-wide without a single
	// error, because the field kept being accepted long after anything read it.
	// A new provider that genuinely cannot express a constraint should fail or
	// warn loudly, not swallow it.
	NegativePrompt string

	// AspectRatio is a semantic ratio label ("1:1", "16:9", "4:3", "3:4",
	// "9:16", "3:2", "2:3", "21:9", "9:21", "5:4", "4:5"). Providers
	// translate to provider-specific dimensions. Empty means "use
	// provider default for this kind".
	AspectRatio string

	// ReferenceImageURIs are optional style/content anchors. S3 URIs
	// (s3://...) or HTTPS URLs. Providers that support reference images
	// fetch them via the injected ReferenceFetcher; providers that don't
	// support log and ignore. Currently: Banana supports up to 20;
	// Stability v1 REST ignores.
	ReferenceImageURIs []string
}

// Result is the provider-agnostic output.
type Result struct {
	// ImageBytes is the raw decoded image data, ready for S3 upload.
	ImageBytes []byte

	// MimeType is the actual MIME type (e.g. "image/png", "image/jpeg").
	MimeType string

	// ProviderName matches Name() of the producing provider. Recorded
	// in assets.origin_model alongside ModelID for provenance.
	ProviderName string

	// ModelID is the provider-specific model identifier:
	//   - Stability: "stable-diffusion-xl-1024-v1-0"
	//   - Banana:    "gemini-3-pro-image-preview"
	// Together with ProviderName, identifies the exact engine that
	// produced the image for reproducibility and cost accounting.
	ModelID string
}

// ReferenceFetcher is the seam for fetching reference image bytes from
// a URI. The imagegenerator adapter provides one implementation that
// wraps the project's S3 client (and falls through to HTTP for URLs).
// Providers that consume reference images accept a ReferenceFetcher at
// construction.
//
// Implementations should return a sensible MIME type even when the
// source doesn't carry one (default to "image/png" for unknown — Gemini
// accepts PNG/JPEG/WEBP).
type ReferenceFetcher interface {
	Fetch(ctx context.Context, uri string) (bytes []byte, mimeType string, err error)
}

// ErrSafetyBlocked is returned when a provider declined to generate
// for safety reasons (NSFW, copyright, blocklist, etc). Retry of the
// same prompt will produce the same block.
var ErrSafetyBlocked = errors.New("imagegenerator: generation blocked by safety policy")

// ErrInvalidRequest is returned for caller-side request errors
// (empty prompt, malformed parameters, etc). Retry pointless.
var ErrInvalidRequest = errors.New("imagegenerator: invalid request")
