// FILE: internal/adapters/imagegenerator/stability/api/client.go
//
// Stability AI REST API client (v1 text-to-image).
//
// Scope: pure HTTP transport. TextToImage is the only public method
// today. No business logic, no dimension validation (the provider snaps
// to SDXLV1Dimensions before calling here), no env reads (caller passes
// the bearer token in).
//
// Authentication:
//   Authorization: Bearer <token>
//
// Errors:
//   - *APIError for HTTP non-2xx, with StatusCode + raw body preserved
//   - standard errors for network failures (caller can errors.Is for
//     context.DeadlineExceeded etc.)
//   - artifact-level finish_reason (SUCCESS vs CONTENT_FILTERED vs ERROR)
//     is the PROVIDER'S concern, not this client's — the client returns
//     the parsed response and lets the provider interpret.

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

// DefaultBaseURL is the Stability AI public API entry point.
const DefaultBaseURL = "https://api.stability.ai"

// Client is the Stability API client. Safe for concurrent use.
// Always construct via NewClient.
type Client struct {
	baseURL    string
	apiKey     string // kept private; sent as Bearer
	httpClient *http.Client
	logger     *zap.Logger
}

// NewClient builds a Client with sensible HTTP defaults.
// Image generation typically takes 5-30s on SDXL v1.0, so timeout is
// 60s. Caller may pass a shorter ctx deadline at call-time.
func NewClient(baseURL, apiKey string, logger *zap.Logger) *Client {
	if baseURL == "" {
		baseURL = DefaultBaseURL
	}
	return &Client{
		baseURL: baseURL,
		apiKey:  apiKey,
		httpClient: &http.Client{
			Timeout: 120 * time.Second,
		},
		logger: logger.Named("stability_api"),
	}
}

// ─────────────────────────────────────────────────────────────────────────
// API methods
// ─────────────────────────────────────────────────────────────────────────

// TextToImage calls POST /v1/generation/{engine}/text-to-image.
//
// engine is the engine ID (use api.EngineSDXL10 for SDXL v1.0).
// Caller is responsible for ensuring req.Width × req.Height is on the
// SDXLV1Dimensions whitelist — the API returns 400 with name=
// invalid_sdxl_v1_dimensions for off-whitelist pairs.
//
// Returns the parsed response with one or more artifacts. Interpretation
// of artifact.FinishReason (SUCCESS vs CONTENT_FILTERED vs ERROR) is the
// provider's concern.
func (c *Client) TextToImage(ctx context.Context, engine string, req TextToImageRequest) (*TextToImageResponse, error) {
	if engine == "" {
		return nil, errors.New("stability api: engine is required")
	}
	if len(req.TextPrompts) == 0 {
		return nil, errors.New("stability api: at least one text_prompt is required")
	}

	var resp TextToImageResponse
	path := "/v1/generation/" + engine + "/text-to-image"
	if err := c.do(ctx, http.MethodPost, path, req, &resp); err != nil {
		return nil, fmt.Errorf("stability text-to-image: %w", err)
	}
	return &resp, nil
}

// ─────────────────────────────────────────────────────────────────────────
// Internal: HTTP layer
// ─────────────────────────────────────────────────────────────────────────

// do is the single HTTP entry point used by all public methods.
// Marshals body, sets Bearer auth, decodes response, maps non-2xx to APIError.
// out may be nil for endpoints that return no body of interest.
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
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	// Outbound log. Body preview is generous on the request side — it's
	// just the prompt + small config, not large like a response.
	c.logger.Info("Stability API request",
		zap.String("method", method),
		zap.String("path", path),
		zap.Int("body_bytes", len(bodyBytes)),
		zap.String("body_preview", previewJSON(bodyBytes, 2048)),
	)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		c.logger.Warn("Stability API HTTP error",
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

	// Response bodies contain base64 image data (~1-3 MB per image).
	// Preview is short — the base64 prefix isn't informative.
	c.logger.Info("Stability API response",
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
// Use errors.As(err, &apiErr) to inspect StatusCode for retry decisions.
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
	return fmt.Sprintf("stability api: %s %s returned %d: %s",
		e.Method, e.Path, e.StatusCode, body)
}

// IsAuth returns true for 401/403 — likely API key misconfigured.
func (e *APIError) IsAuth() bool {
	return e.StatusCode == http.StatusUnauthorized || e.StatusCode == http.StatusForbidden
}

// IsNotFound returns true for 404 — likely a typo in engine ID.
func (e *APIError) IsNotFound() bool {
	return e.StatusCode == http.StatusNotFound
}

// IsRateLimit returns true for 429 — back off + retry.
func (e *APIError) IsRateLimit() bool {
	return e.StatusCode == http.StatusTooManyRequests
}

// IsServer returns true for 5xx — Stability-side problem, retry may help.
func (e *APIError) IsServer() bool {
	return e.StatusCode >= 500
}
