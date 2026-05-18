// FILE: internal/adapters/banana/api/types.go
//
// Shared request/response types for the Banana (Gemini Image) REST API
// client.
// Schema source: https://ai.google.dev/gemini-api/docs/image-generation
// (Gemini 2.5/3.x image generation reference, verified 2026-05-18)
//
// IMPORTANT: Gemini's API uses a nested contents/parts structure where
// the same Part can hold EITHER text OR inlineData (image). The struct
// reflects that with both fields being pointers/optional and
// omitempty-tagged so the JSON encoder doesn't send empty alternatives.
//
// Aspect ratio handling differs between model families — see
// buildGenerateContentRequest in client.go for the routing.
package api

// ─── Model identifiers ──────────────────────────────────────────────────────

// Banana model identifiers (verified 2026-05-18 from Google AI for
// Developers docs). The two preview tiers may rotate as Gemini releases
// progress — pin the model string in adapter config and bump when needed.
const (
	// ModelNanoBanana is the original Nano Banana (Gemini 2.5 Flash Image).
	// Lowest cost (~$0.039/image direct), GA. 1024px output, no resolution selector.
	ModelNanoBanana = "gemini-2.5-flash-image"

	// ModelNanoBanana2 is the preview tier of Gemini 3.1 Flash Image.
	// Resolution-tiered pricing. Faster than Pro, smarter than 2.5.
	ModelNanoBanana2 = "gemini-3.1-flash-image-preview"

	// ModelNanoBananaPro is Gemini 3 Pro Image (preview). "Thinking" mode
	// for complex instruction-following; up to 4K resolution. Best for
	// flat illustration, text rendering, and prompt adherence.
	// Best fit for our icon use case.
	ModelNanoBananaPro = "gemini-3-pro-image-preview"
)

// ─── Aspect ratios ──────────────────────────────────────────────────────────

// AspectRatio constants supported by Gemini Image. The full set is
// ten ratios spanning 1:1 through 21:9 per Google's docs.
// Adapter should map our internal aspect strings to these.
const (
	AspectRatio1to1  = "1:1"
	AspectRatio16to9 = "16:9"
	AspectRatio9to16 = "9:16"
	AspectRatio4to3  = "4:3"
	AspectRatio3to4  = "3:4"
	AspectRatio3to2  = "3:2"
	AspectRatio2to3  = "2:3"
	AspectRatio21to9 = "21:9"
	AspectRatio9to21 = "9:21"
	AspectRatio5to4  = "5:4"
)

// ─── Finish reasons ─────────────────────────────────────────────────────────

// FinishReason values seen in candidate responses. Used by client.go
// to detect safety blocks vs successful generations.
const (
	FinishReasonStop              = "STOP"
	FinishReasonMaxTokens         = "MAX_TOKENS"
	FinishReasonSafety            = "SAFETY"
	FinishReasonRecitation        = "RECITATION"
	FinishReasonProhibitedContent = "PROHIBITED_CONTENT"
	FinishReasonBlocklist         = "BLOCKLIST"
	FinishReasonOther             = "OTHER"
)

// ─── Adapter-facing request type ────────────────────────────────────────────

// GenerateRequest is the flat input the adapter passes to GenerateImage.
// The client transforms this into the nested Gemini contents/parts shape.
type GenerateRequest struct {
	// Model is the Gemini model identifier (see Model* constants).
	// Required.
	Model string

	// Prompt is the text prompt for image generation. Required.
	Prompt string

	// AspectRatio is one of the AspectRatio* constants. Optional — when
	// empty, Gemini defaults to 1:1 (or matches an input reference image
	// when one is provided).
	AspectRatio string

	// ReferenceImages are optional style/content anchors. Max 20.
	// Caller (adapter) is responsible for base64-encoding raw image bytes.
	ReferenceImages []ReferenceImage
}

// ReferenceImage is an input reference image already base64-encoded.
// Adapter handles the encode step from raw bytes (typically fetched from
// S3 or read off disk).
type ReferenceImage struct {
	// MimeType is the image MIME type ("image/png", "image/jpeg", "image/webp").
	MimeType string

	// Base64Data is the base64-encoded image bytes (no data: URL prefix).
	Base64Data string
}

// ─── Adapter-facing response type ───────────────────────────────────────────

// GenerateImageResult is the simplified response the adapter consumes.
// One GenerateRequest may produce multiple Images (if Gemini returns
// several candidates), though for non-batch single-prompt calls it's
// usually exactly one.
type GenerateImageResult struct {
	Images []GeneratedImage
}

// GeneratedImage is a single image returned by Gemini, still
// base64-encoded. The adapter decodes before passing to S3 upload to
// avoid double-allocating large blobs in this layer.
type GeneratedImage struct {
	MimeType   string
	Base64Data string
}

// ─── Gemini wire format (internal) ──────────────────────────────────────────
//
// The types below mirror Gemini's API surface — kept private-ish (lowercase
// init not used; types are exported because client.go needs them in the same
// package). External code should only see GenerateRequest/GenerateImageResult
// above.

// GenerateContentRequest is the top-level body for :generateContent.
type GenerateContentRequest struct {
	Contents         []Content         `json:"contents"`
	GenerationConfig *GenerationConfig `json:"generationConfig,omitempty"`

	// SafetySettings is reserved for future use. Currently omitted — the
	// default Google safety thresholds are appropriate for our use case.
	// SafetySettings []SafetySetting `json:"safetySettings,omitempty"`
}

// Content holds a sequence of parts. For generateContent the request
// typically has a single Content containing prompt + reference image parts.
type Content struct {
	Parts []ContentPart `json:"parts"`

	// Role is optional for image generation; leaving omitempty means
	// Gemini treats it as a user message implicitly.
	Role string `json:"role,omitempty"`
}

// ContentPart is a discriminated union: exactly one of Text or InlineData
// should be set per part.
type ContentPart struct {
	Text       string      `json:"text,omitempty"`
	InlineData *InlineData `json:"inlineData,omitempty"`
}

// InlineData carries a base64-encoded blob (used for reference images on
// the request side and generated images on the response side).
type InlineData struct {
	MimeType string `json:"mimeType"`
	Data     string `json:"data"` // base64-encoded
}

// GenerationConfig holds generation-time tuning parameters.
type GenerationConfig struct {
	// ResponseModalities tells the API what the output type should be.
	// For image generation we send ["Image"]; for interleaved text+image
	// (e.g. tutorial steps) it would be ["Image","Text"].
	ResponseModalities []string `json:"responseModalities,omitempty"`

	// ImageConfig holds image-specific generation knobs (aspect ratio etc).
	// Used by gemini-3.1-flash-image-preview and gemini-3-pro-image-preview.
	ImageConfig *ImageConfig `json:"imageConfig,omitempty"`

	// TODO(banana/aspect-ratio-path): gemini-2.5-flash-image uses
	// response_format.image.aspect_ratio instead of imageConfig. If we
	// want to support that model, add a ResponseFormat field here and
	// branch in buildGenerateContentRequest based on model prefix.
}

// ImageConfig is the image-specific knob set for Gemini 3.x image preview
// models. aspectRatio is the only field we use today.
type ImageConfig struct {
	AspectRatio string `json:"aspectRatio,omitempty"`
}

// ─── Response types ─────────────────────────────────────────────────────────

// GenerateContentResponse is the top-level :generateContent response.
type GenerateContentResponse struct {
	Candidates    []Candidate    `json:"candidates"`
	UsageMetadata *UsageMetadata `json:"usageMetadata,omitempty"`

	// PromptFeedback appears when the prompt itself was filtered.
	// We don't surface this field today; if needed for diagnostics,
	// fetch from this struct in the adapter.
	// PromptFeedback *PromptFeedback `json:"promptFeedback,omitempty"`
}

// Candidate is one model output. Image generation typically returns one
// candidate per request, but the API supports multiple in principle.
type Candidate struct {
	Content      Content `json:"content"`
	FinishReason string  `json:"finishReason"`
	Index        int     `json:"index"`

	// SafetyRatings is reserved — we only need FinishReason for the
	// "block vs success" decision today.
	// SafetyRatings []SafetyRating `json:"safetyRatings,omitempty"`
}

// UsageMetadata reports token usage. Useful for cost accounting if we
// surface it to a billing table later.
type UsageMetadata struct {
	PromptTokenCount     int `json:"promptTokenCount"`
	CandidatesTokenCount int `json:"candidatesTokenCount"`
	TotalTokenCount      int `json:"totalTokenCount"`
}
