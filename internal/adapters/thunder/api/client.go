// FILE: internal/adapters/thunder/api/client.go
//
// Thunder Compute REST API client.
//
// Scope: pure HTTP transport. CreateInstance, ListInstances, GetInstance,
// DeleteInstance, plus a WaitForRunning polling helper for the common
// "create then wait until ready" flow. No business logic, no DB, no SSH.
//
// Caller pattern (action handler):
//   client := api.NewClient(baseURL, token, logger)
//   resp, err := client.CreateInstance(ctx, req)
//   if err != nil { return ... }
//   inst, err := client.WaitForRunning(ctx, resp.Identifier, 5*time.Second, 5*time.Minute)
//   ...
//
// Authentication:
//   Authorization: Bearer <token>
//
// Errors are returned as *APIError for HTTP non-2xx, with StatusCode and
// raw body preserved. Network errors are returned as standard errors
// (caller can errors.Is(err, context.DeadlineExceeded) etc.).

package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"go.uber.org/zap"
)

// Client is the Thunder Compute API client. Safe for concurrent use.
// Always construct via NewClient.
type Client struct {
	baseURL    string // e.g. "https://api.thundercompute.com:8443/v1"
	token      string // bearer token, kept private
	httpClient *http.Client
	logger     *zap.Logger
}

// NewClient builds a Client with sensible HTTP defaults (30s timeout,
// connection pool). The caller is expected to read the bearer token from
// env and pass it in — the client doesn't reach into os.Getenv directly,
// so it's testable.
func NewClient(baseURL, token string, logger *zap.Logger) *Client {
	return &Client{
		baseURL: baseURL,
		token:   token,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		logger: logger.Named("thunder_api"),
	}
}

// ─────────────────────────────────────────────────────────────────────────
// API methods
// ─────────────────────────────────────────────────────────────────────────

// CreateInstance creates a new Thunder Compute instance.
// Returns immediately after the API accepts the create request; the
// instance will be in PENDING state. Use WaitForRunning to block until
// it's reachable via SSH.
func (c *Client) CreateInstance(ctx context.Context, req CreateInstanceRequest) (*CreateInstanceResponse, error) {
	var resp CreateInstanceResponse
	if err := c.do(ctx, http.MethodPost, "/instances/create", req, &resp); err != nil {
		return nil, fmt.Errorf("thunder create instance: %w", err)
	}
	c.logger.Info("Thunder instance create accepted",
		zap.Int("identifier", resp.Identifier),
		zap.String("uuid", resp.UUID),
		zap.Bool("server_generated_key", resp.Key != ""),
	)
	return &resp, nil
}

// ListInstances returns all instances visible to the authenticated token.
// Useful for the reaper and for reconciliation queries.
//
// VERIFIED 2026-05-20/21: GET /instances/list returns a JSON OBJECT keyed by
// string id, e.g. {"0":{...},"1":{...}} — NOT an array, and NOT wrapped in
// {instances:[...]}. The id is the MAP KEY; it is not a field inside the
// object, so we inject it into Instance.ID after decode.
func (c *Client) ListInstances(ctx context.Context) ([]Instance, error) {
	var byID map[string]Instance
	if err := c.do(ctx, http.MethodGet, "/instances/list", nil, &byID); err != nil {
		return nil, fmt.Errorf("thunder list instances: %w", err)
	}
	out := make([]Instance, 0, len(byID))
	for id, inst := range byID {
		inst.ID = id // the map key is the identifier; the object has no id field
		out = append(out, inst)
	}
	return out, nil
}

// GetInstance fetches a single instance by numeric identifier.
//
// VERIFIED 2026-05-21: Thunder has NO GET /instances/{id} endpoint — it 404s.
// The only way to read an instance is via /instances/list and matching on the
// map key. So this lists and filters. (Kept as a method for caller convenience
// and so WaitForRunning has a single per-poll call site.)
func (c *Client) GetInstance(ctx context.Context, identifier int) (*Instance, error) {
	all, err := c.ListInstances(ctx)
	if err != nil {
		return nil, err
	}
	for i := range all {
		if n, ok := all[i].IdentifierInt(); ok && n == identifier {
			return &all[i], nil
		}
	}
	// Not found in the list. Return an APIError with 404 so callers (and
	// WaitForRunning) can distinguish "not there yet" from a transport error.
	return nil, &APIError{
		StatusCode: http.StatusNotFound,
		Method:     http.MethodGet,
		Path:       "/instances/list",
		Body:       fmt.Sprintf("instance %d not present in list", identifier),
	}
}

// DeleteInstance terminates an instance. Idempotent: deleting an already-
// terminated instance returns a non-error (we treat 404 as success).
func (c *Client) DeleteInstance(ctx context.Context, identifier int) error {
	// Thunder uses POST /instances/{id}/delete (not REST-style DELETE).
	path := "/instances/" + strconv.Itoa(identifier) + "/delete"
	err := c.do(ctx, http.MethodPost, path, nil, nil)

	// 404 = already gone, treat as success.
	var apiErr *APIError
	if errors.As(err, &apiErr) && apiErr.StatusCode == http.StatusNotFound {
		c.logger.Info("Thunder instance already deleted (404)",
			zap.Int("identifier", identifier))
		return nil
	}
	if err != nil {
		return fmt.Errorf("thunder delete instance %d: %w", identifier, err)
	}
	c.logger.Info("Thunder instance delete accepted",
		zap.Int("identifier", identifier))
	return nil
}

// ─────────────────────────────────────────────────────────────────────────
// Polling helper
// ─────────────────────────────────────────────────────────────────────────

// WaitForRunning polls GetInstance until status == RUNNING, ctx expires,
// or the instance reaches a terminal-non-running status (TERMINATED,
// ERROR — those return ErrInstanceTerminal).
//
// pollInterval controls poll frequency; 5*time.Second is a reasonable
// default. The overall deadline comes from ctx, so callers should wrap
// with a timeout-context: ctx, cancel := context.WithTimeout(parent, 5*time.Minute).
func (c *Client) WaitForRunning(ctx context.Context, identifier int, pollInterval time.Duration) (*Instance, error) {
	if pollInterval < 1*time.Second {
		pollInterval = 1 * time.Second // protect Thunder's rate limits
	}
	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	// Do an immediate first check (don't wait for the first tick).
	for {
		inst, err := c.GetInstance(ctx, identifier)
		if err != nil {
			// Transient — log and retry on next tick. Non-transient
			// errors (auth, validation) will keep failing; the ctx
			// deadline will eventually break us out.
			c.logger.Warn("WaitForRunning: GetInstance failed, will retry",
				zap.Int("identifier", identifier),
				zap.Error(err))
		} else {
			c.logger.Debug("WaitForRunning poll",
				zap.Int("identifier", identifier),
				zap.String("status", inst.Status))
			if IsReadyStatus(inst.Status) {
				return inst, nil
			}
			// Running already handled above; IsTerminalStatus (case-insensitive)
			// catches failed/deleted. Thunder returns these UPPERCASE, so the
			// exact-match comparison this replaced would have missed them and
			// looped until ctx timeout.
			if IsTerminalStatus(inst.Status) {
				return inst, fmt.Errorf("%w: instance %d reached %s",
					ErrInstanceTerminal, identifier, inst.Status)
			}
		}

		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("WaitForRunning: %w (instance %d)",
				ctx.Err(), identifier)
		case <-ticker.C:
			// next poll
		}
	}
}

// ErrInstanceTerminal is returned by WaitForRunning when the instance
// reaches TERMINATED or ERROR before becoming RUNNING.
var ErrInstanceTerminal = errors.New("instance reached terminal non-running state")

// ─────────────────────────────────────────────────────────────────────────
// Internal: HTTP layer
// ─────────────────────────────────────────────────────────────────────────

// do is the single HTTP entry point used by all public methods.
// Marshals body, sets bearer auth, decodes response, maps non-2xx to APIError.
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
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	// Log the outbound call. Body is capped to avoid blowing log size.
	// public_key is a public key — not sensitive. The bearer token lives
	// in req.Header, not the JSON body, so it doesn't appear here.
	c.logger.Info("Thunder API request",
		zap.String("method", method),
		zap.String("path", path),
		zap.Int("body_bytes", len(bodyBytes)),
		zap.String("body_preview", previewJSON(bodyBytes, 2048)),
	)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		c.logger.Warn("Thunder API HTTP error",
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

	c.logger.Info("Thunder API response",
		zap.String("method", method),
		zap.String("path", path),
		zap.Int("status", resp.StatusCode),
		zap.Int("body_bytes", len(respBody)),
		zap.String("body_preview", previewJSON(respBody, 2048)),
	)

	// Map non-2xx to APIError. Caller can errors.As() to inspect.
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
// Use errors.As(err, &apiErr) to inspect StatusCode (e.g. for 404 fallback).
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
	return fmt.Sprintf("thunder api: %s %s returned %d: %s",
		e.Method, e.Path, e.StatusCode, body)
}

// IsAuth returns true for 401/403 — likely token misconfigured.
func (e *APIError) IsAuth() bool {
	return e.StatusCode == http.StatusUnauthorized || e.StatusCode == http.StatusForbidden
}

// IsNotFound returns true for 404.
func (e *APIError) IsNotFound() bool {
	return e.StatusCode == http.StatusNotFound
}

// IsRateLimit returns true for 429.
func (e *APIError) IsRateLimit() bool {
	return e.StatusCode == http.StatusTooManyRequests
}

// IsServer returns true for 5xx — Thunder-side problem, retry may help.
func (e *APIError) IsServer() bool {
	return e.StatusCode >= 500
}
