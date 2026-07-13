// FILE: internal/adapters/banana/api/client.go
//
// Google Gemini Image (Nano Banana family) REST API client.
//
// Scope: pure HTTP transport. GenerateImage is the only public method
// today. No business logic, no DB, no caching.
//
// Caller pattern (image-generator action via adapter):
//   client := api.NewClient(baseURL, apiKey, logger)
//   resp, err := client.GenerateImage(ctx, req)
//   if err != nil { return ... }
//   // resp.Images[0].Base64Data is a base64 string; adapter decodes
//   // before passing to S3 upload.
//
// Authentication:
//   x-goog-api-key: <key>
//   (NOT Bearer — Gemini uses its own header. Vertex AI variant would use
//    OAuth2 — out of scope here, see TODO at top of constants.)
//
// Errors are returned as *APIError for HTTP non-2xx, with StatusCode and
// raw body preserved. Safety blocks (200 OK + finish_reason=SAFETY) are
// surfaced as ErrSafetyBlocked rather than returned silently as empty.

package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"go.uber.org/zap"
)

// DefaultBaseURL is the public Gemini API (AI Studio) entry point.
//
// TODO(banana/vertex): for higher rate limits or enterprise requirements,
// add a Vertex AI client variant that uses Google OAuth2 + region-prefixed
// URLs ("https://us-central1-aiplatform.googleapis.com/v1/projects/...").
// Out of scope for the initial integration.
const DefaultBaseURL = "https://generativelanguage.googleapis.com/v1beta"

// Client is the Banana (Gemini Image) API client. Safe for concurrent use.
// Always construct via NewClient.
type Client struct {
	baseURL    string
	apiKey     string // kept private; sent in x-goog-api-key header
	httpClient *http.Client
	logger     *zap.Logger
}

// NewClient builds a Client with sensible HTTP defaults.
// Image generation can take 5-15s for Pro tier, so timeout is 90s.
// Caller may pass a tighter ctx deadline at call-time to fail faster.
func NewClient(baseURL, apiKey string, logger *zap.Logger) *Client {
	if baseURL == "" {
		baseURL = DefaultBaseURL
	}
	return &Client{
		baseURL: baseURL,
		apiKey:  apiKey,
		httpClient: &http.Client{
			Timeout: 90 * time.Second,
		},
		logger: logger.Named("banana_api"),
	}
}

// ─────────────────────────────────────────────────────────────────────────
// API methods
// ─────────────────────────────────────────────────────────────────────────

// GenerateImage produces one or more images from a text prompt, optionally
// conditioned on reference images. Synchronous: returns when the Gemini
// API returns (no polling — this API is non-async unlike Stability batch).
//
// Reference images in req.ReferenceImages are passed as already-base64-
// encoded byte strings with their MIME type. The adapter layer handles
// the encoding step from raw bytes — keeps this client content-agnostic
// and avoids double-allocation of large blobs.
//
// Max 20 references per Google's docs; no client-side enforcement —
// Gemini returns 400 if exceeded.
//
// Response images are returned as base64 strings; adapter decodes before
// passing to S3 upload.
//
// Safety blocks: if the model declines to generate (NSFW, copyrighted
// content, etc.), Gemini returns 200 with finish_reason=SAFETY per
// candidate. This method surfaces that as ErrSafetyBlocked rather than
// returning empty image data silently.
func (c *Client) GenerateImage(ctx context.Context, req GenerateRequest) (*GenerateImageResult, error) {
	if req.Model == "" {
		return nil, errors.New("banana api: GenerateRequest.Model is required")
	}
	if req.Prompt == "" {
		return nil, errors.New("banana api: GenerateRequest.Prompt is required")
	}

	body := buildGenerateContentRequest(req)
	var raw GenerateContentResponse
	path := "/models/" + req.Model + ":generateContent"
	if err := c.do(ctx, http.MethodPost, path, body, &raw); err != nil {
		return nil, fmt.Errorf("banana generate image: %w", err)
	}

	result, err := extractImagesFromResponse(&raw)
	if err != nil {
		return nil, err
	}

	c.logger.Info("Banana image generation succeeded",
		zap.String("model", req.Model),
		zap.Int("images_returned", len(result.Images)),
		zap.Int("reference_images_sent", len(req.ReferenceImages)),
		zap.String("aspect_ratio", req.AspectRatio),
	)
	return result, nil
}

// ─────────────────────────────────────────────────────────────────────────
// Internal: request body construction
// ─────────────────────────────────────────────────────────────────────────

// buildGenerateContentRequest translates our flat GenerateRequest into
// Gemini's nested contents/parts structure. Reference images are added
// as inlineData parts after the text part.
//
// TODO(banana/aspect-ratio-path): docs show two distinct paths for aspect
// ratio depending on model family:
//   - gemini-2.5-flash-image: response_format.image.aspect_ratio
//   - gemini-3.1-flash-image-preview, gemini-3-pro-image-preview:
//     generationConfig.imageConfig.aspectRatio
//
// First real call against Pro: confirm imageConfig path is honored. If
// 2.5-flash rejects imageConfig, branch on model prefix here.
func buildGenerateContentRequest(req GenerateRequest) *GenerateContentRequest {
	parts := []ContentPart{
		{Text: req.Prompt},
	}
	for _, ref := range req.ReferenceImages {
		parts = append(parts, ContentPart{
			InlineData: &InlineData{
				MimeType: ref.MimeType,
				Data:     ref.Base64Data,
			},
		})
	}

	body := &GenerateContentRequest{
		Contents: []Content{
			{Parts: parts},
		},
		GenerationConfig: &GenerationConfig{
			ResponseModalities: []string{"Image"},
		},
	}
	if req.AspectRatio != "" {
		body.GenerationConfig.ImageConfig = &ImageConfig{
			AspectRatio: req.AspectRatio,
		}
	}
	return body
}

// extractImagesFromResponse walks Gemini's candidates/content/parts to
// collect inlineData parts (the images). Returns ErrSafetyBlocked if
// the response contained no images and at least one candidate was
// blocked by safety policy.
func extractImagesFromResponse(raw *GenerateContentResponse) (*GenerateImageResult, error) {
	result := &GenerateImageResult{}
	var anyBlocked bool

	for _, cand := range raw.Candidates {
		if cand.FinishReason == FinishReasonSafety ||
			cand.FinishReason == FinishReasonProhibitedContent ||
			cand.FinishReason == FinishReasonBlocklist {
			anyBlocked = true
			continue
		}
		for _, part := range cand.Content.Parts {
			if part.InlineData == nil {
				continue
			}
			result.Images = append(result.Images, GeneratedImage{
				MimeType:   part.InlineData.MimeType,
				Base64Data: part.InlineData.Data,
			})
		}
	}

	if len(result.Images) == 0 {
		if anyBlocked {
			return nil, ErrSafetyBlocked
		}
		// No images and no explicit safety block: surface as error rather
		// than silently empty. Indicates either malformed response or a
		// finish_reason we don't recognise.
		return nil, fmt.Errorf("banana: response contained no image parts and no safety block")
	}
	return result, nil
}

// ErrSafetyBlocked is returned when Gemini blocks generation due to
// safety policies (prompt content, output content, or blocklisted
// patterns). Caller should NOT retry the same prompt — it will block
// again.
var ErrSafetyBlocked = errors.New("banana: generation blocked by safety policy")

// ─────────────────────────────────────────────────────────────────────────
// Internal: HTTP layer
// ─────────────────────────────────────────────────────────────────────────

// do is the single HTTP entry point used by all public methods.
// Marshals body, sets x-goog-api-key auth, decodes response,
// maps non-2xx to APIError. out may be nil.
func (c *Client) do(ctx context.Context, method, path string, body, out interface{}) error {
	var bodyReader io.Reader
	var bodyBytes []byte
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("marshal request: %w", err)
		}
		bodyBytes = b
		bodyReader = bytes.NewReader(b)
	}

	url := c.baseURL + path
	req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("x-goog-api-key", c.apiKey)
	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	// Outbound log. Body preview is capped — reference images base64'd in
	// inlineData can be MB-scale and shouldn't bloat logs. The previewable
	// prefix typically contains the prompt text plus the start of the
	// first inlineData chunk.
	c.logger.Info("Banana API request",
		zap.String("method", method),
		zap.String("path", path),
		zap.Int("body_bytes", len(bodyBytes)),
		zap.String("body_preview", previewJSON(bodyBytes, 1024)),
	)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		c.logger.Warn("Banana API HTTP error",
			zap.String("method", method),
			zap.String("path", path),
			zap.Error(err),
		)
		return fmt.Errorf("http: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response: %w", err)
	}

	// Response bodies are large (1MB+ base64 image). Preview is short.
	c.logger.Info("Banana API response",
		zap.String("method", method),
		zap.String("path", path),
		zap.Int("status", resp.StatusCode),
		zap.Int("body_bytes", len(respBody)),
		zap.String("body_preview", previewJSON(respBody, 512)),
	)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return &APIError{
			StatusCode: resp.StatusCode,
			Method:     method,
			Path:       path,
			Body:       string(respBody),
		}
	}

	if out == nil || len(respBody) == 0 {
		return nil
	}
	if err := json.Unmarshal(respBody, out); err != nil {
		return fmt.Errorf("decode response (status %d): %w", resp.StatusCode, err)
	}
	return nil
}

// previewJSON returns a printable, length-bounded version of a JSON byte
// buffer suitable for logs. Truncated bodies get a "...(truncated)" suffix.
// Empty input returns "<empty>".
func previewJSON(b []byte, max int) string {
	if len(b) == 0 {
		return "<empty>"
	}
	if len(b) <= max {
		return string(b)
	}
	return string(b[:max]) + "...(truncated)"
}

// ─────────────────────────────────────────────────────────────────────────
// APIError
// ─────────────────────────────────────────────────────────────────────────

// APIError is returned by Client methods for non-2xx HTTP responses.
// Use errors.As(err, &apiErr) to inspect StatusCode (e.g. for retry decisions).
type APIError struct {
	StatusCode int
	Method     string
	Path       string
	Body       string // raw response body, useful for debugging
}

func (e *APIError) Error() string {
	const maxBodyInError = 200
	body := e.Body
	if len(body) > maxBodyInError {
		body = body[:maxBodyInError] + "..."
	}
	return fmt.Sprintf("banana api: %s %s returned %d: %s",
		e.Method, e.Path, e.StatusCode, body)
}

// IsAuth returns true for 401/403 — likely API key misconfigured.
func (e *APIError) IsAuth() bool {
	return e.StatusCode == http.StatusUnauthorized || e.StatusCode == http.StatusForbidden
}

// IsNotFound returns true for 404 — likely a typo in model name.
func (e *APIError) IsNotFound() bool {
	return e.StatusCode == http.StatusNotFound
}

// IsRateLimit returns true for 429 — back off + retry.
func (e *APIError) IsRateLimit() bool {
	return e.StatusCode == http.StatusTooManyRequests
}

// IsServer returns true for 5xx — Google-side problem, retry may help.
func (e *APIError) IsServer() bool {
	return e.StatusCode >= 500
}
