// FILE: internal/adapters/imagegenerator/stability/api/types.go
//
// Stability AI v1 REST API types for text-to-image.
// Schema source: https://platform.stability.ai/docs/api-reference
//
// SDXLV1Dimensions is the strict whitelist of dimension pairs the v0.9
// and v1.0 endpoints accept. Any other pair produces a 400 with
// name=invalid_sdxl_v1_dimensions. The whitelist is owned by the API
// layer (it's a wire-format constraint) and consumed by the provider
// when snapping abstract aspect-ratio labels to valid dimensions.

package api

// Engine identifiers for Stability's v1 generation endpoints.
const (
	// EngineSDXL10 is SDXL v1.0 — the current production engine.
	// Strict dimension whitelist (see SDXLV1Dimensions).
	EngineSDXL10 = "stable-diffusion-xl-1024-v1-0"

	// EngineSDXL09 is SDXL v0.9 — same dimension whitelist as v1.0.
	EngineSDXL09 = "stable-diffusion-xl-1024-v0-9"
)

// SDXLV1Dimensions enumerates the dimension pairs accepted by SDXL v0.9
// and v1.0. Any other pair returns 400 with name=invalid_sdxl_v1_dimensions.
// Source: the API's own error response message.
var SDXLV1Dimensions = [][2]int{
	{1024, 1024}, // 1:1
	{1152, 896},  // ~1.29:1 landscape
	{1216, 832},  // ~1.46:1 landscape
	{1344, 768},  // ~1.75:1 landscape (near 16:9)
	{1536, 640},  // 2.4:1 cinematic landscape
	{896, 1152},  // ~0.78:1 portrait
	{832, 1216},  // ~0.68:1 portrait
	{768, 1344},  // ~0.57:1 portrait (near 9:16)
	{640, 1536},  // ~0.42:1 ultra-portrait
}

// ─── Request types ─────────────────────────────────────────────────────────

// TextToImageRequest is the body for
// POST /v1/generation/{engine}/text-to-image.
type TextToImageRequest struct {
	// TextPrompts is the prompt array. At least one prompt with weight > 0
	// is required. Negative prompts have weight < 0 (typically -1.0).
	TextPrompts []TextPrompt `json:"text_prompts"`

	// Width and Height must be a pair on SDXLV1Dimensions. The provider
	// resolves the pair before constructing the request.
	Width  int `json:"width,omitempty"`
	Height int `json:"height,omitempty"`

	// Samples is the number of artifacts to generate (1..10, default 1).
	Samples int `json:"samples,omitempty"`

	// Steps is the diffusion step count (10..150, default 30 in this provider).
	// Higher = slower but slightly higher quality.
	Steps int `json:"steps,omitempty"`

	// CFGScale is the classifier-free guidance scale (0..35, default 7).
	// Higher = stricter prompt adherence, often at the cost of variety.
	CFGScale float64 `json:"cfg_scale,omitempty"`

	// StylePreset optionally biases the output toward a style
	// ("digital-art", "low-poly", "anime", "photographic", ...). Not used
	// today — we steer style via the prompt itself for determinism.
	StylePreset string `json:"style_preset,omitempty"`

	// Seed sets the random seed. 0 = random. Useful for reproducibility
	// in tests but otherwise not surfaced.
	Seed uint32 `json:"seed,omitempty"`
}

// TextPrompt is a single weighted prompt. Positive prompts have weight
// > 0 (typically 1.0); negative prompts have weight < 0 (typically -1.0).
type TextPrompt struct {
	Text   string  `json:"text"`
	Weight float64 `json:"weight,omitempty"`
}

// ─── Response types ────────────────────────────────────────────────────────

// TextToImageResponse holds the artifacts returned by the generation call.
type TextToImageResponse struct {
	Artifacts []Artifact `json:"artifacts"`
}

// Artifact is a single generated image. "Artifact" is Stability's term —
// the API historically supported other outputs; today only images come back.
type Artifact struct {
	// Base64 is the standard base64-encoded image bytes (no data: URL prefix).
	Base64 string `json:"base64"`

	// Seed is the random seed used to generate this artifact.
	Seed uint32 `json:"seed"`

	// FinishReason is the per-artifact outcome — see FinishReason* constants.
	FinishReason string `json:"finishReason"`
}

// FinishReason values returned in Artifact.FinishReason.
const (
	FinishReasonSuccess         = "SUCCESS"
	FinishReasonError           = "ERROR"
	FinishReasonContentFiltered = "CONTENT_FILTERED"
)
