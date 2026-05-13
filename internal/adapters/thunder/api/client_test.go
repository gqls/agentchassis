// FILE: internal/adapters/thunder/api/client_test.go
//
// Unit tests using httptest. No real API access required; safe to run
// in CI. For real-API integration testing, write a separate test
// (e.g. client_integration_test.go) guarded by a build tag or env var.

package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"go.uber.org/zap/zaptest"
)

// ─────────────────────────────────────────────────────────────────────────
// CreateInstance
// ─────────────────────────────────────────────────────────────────────────

func TestCreateInstance_WithPublicKey(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request shape
		if r.Method != http.MethodPost {
			t.Errorf("method: got %q want POST", r.Method)
		}
		if r.URL.Path != "/instances/create" {
			t.Errorf("path: got %q want /instances/create", r.URL.Path)
		}
		if auth := r.Header.Get("Authorization"); auth != "Bearer test-token" {
			t.Errorf("auth: got %q want %q", auth, "Bearer test-token")
		}

		var req CreateInstanceRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		if req.PublicKey == "" {
			t.Error("PublicKey should be set in request")
		}

		// Server-side key empty because client supplied PublicKey
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(CreateInstanceResponse{
			Identifier: 42,
			Key:        "", // empty — client provided public key
			UUID:       "uuid-abc",
		})
	}))
	defer srv.Close()

	c := NewClient(srv.URL, "test-token", zaptest.NewLogger(t))
	ctx := context.Background()

	resp, err := c.CreateInstance(ctx, CreateInstanceRequest{
		GPU:       "a100",
		Mode:      "prototyping",
		PublicKey: "ssh-ed25519 AAAA...",
	})
	if err != nil {
		t.Fatalf("CreateInstance: %v", err)
	}
	if resp.Identifier != 42 {
		t.Errorf("Identifier: got %d want 42", resp.Identifier)
	}
	if resp.UUID != "uuid-abc" {
		t.Errorf("UUID: got %q want uuid-abc", resp.UUID)
	}
	if resp.Key != "" {
		t.Errorf("Key: got %q, expected empty (client provided public_key)", resp.Key)
	}
}

func TestCreateInstance_AuthError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error": "invalid token"}`))
	}))
	defer srv.Close()

	c := NewClient(srv.URL, "bad-token", zaptest.NewLogger(t))
	_, err := c.CreateInstance(context.Background(), CreateInstanceRequest{})
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected *APIError, got %T: %v", err, err)
	}
	if !apiErr.IsAuth() {
		t.Errorf("IsAuth() = false, want true (status %d)", apiErr.StatusCode)
	}
}

// ─────────────────────────────────────────────────────────────────────────
// ListInstances
// ─────────────────────────────────────────────────────────────────────────

func TestListInstances(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/instances" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		json.NewEncoder(w).Encode([]Instance{
			{Identifier: 1, UUID: "u1", Status: InstanceStatusRunning, IP: "10.0.0.1", GPU: "a100"},
			{Identifier: 2, UUID: "u2", Status: InstanceStatusPending, IP: "", GPU: "a100"},
		})
	}))
	defer srv.Close()

	c := NewClient(srv.URL, "tok", zaptest.NewLogger(t))
	got, err := c.ListInstances(context.Background())
	if err != nil {
		t.Fatalf("ListInstances: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("got %d instances, want 2", len(got))
	}
	if got[0].IP != "10.0.0.1" || got[1].IP != "" {
		t.Errorf("IPs: got %q, %q", got[0].IP, got[1].IP)
	}
}

// ─────────────────────────────────────────────────────────────────────────
// GetInstance with 404 fallback to ListInstances
// ─────────────────────────────────────────────────────────────────────────

func TestGetInstance_FallsBackToListOn404(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Direct fetch returns 404
		if r.URL.Path == "/instances/7" {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(`{"error": "not in index yet"}`))
			return
		}
		// List returns the instance
		if r.URL.Path == "/instances" {
			json.NewEncoder(w).Encode([]Instance{
				{Identifier: 7, UUID: "u7", Status: InstanceStatusPending},
			})
			return
		}
		t.Errorf("unexpected path %q", r.URL.Path)
	}))
	defer srv.Close()

	c := NewClient(srv.URL, "tok", zaptest.NewLogger(t))
	inst, err := c.GetInstance(context.Background(), 7)
	if err != nil {
		t.Fatalf("GetInstance: %v", err)
	}
	if inst.Identifier != 7 || inst.UUID != "u7" {
		t.Errorf("got identifier=%d uuid=%q", inst.Identifier, inst.UUID)
	}
}

func TestGetInstance_404IfNotInList(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Both direct fetch and list return empty/404
		if r.URL.Path == "/instances/99" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		json.NewEncoder(w).Encode([]Instance{}) // empty list
	}))
	defer srv.Close()

	c := NewClient(srv.URL, "tok", zaptest.NewLogger(t))
	_, err := c.GetInstance(context.Background(), 99)
	if err == nil {
		t.Fatal("expected 404 error, got nil")
	}
	var apiErr *APIError
	if !errors.As(err, &apiErr) || !apiErr.IsNotFound() {
		t.Errorf("expected APIError with IsNotFound, got %T %v", err, err)
	}
}

// ─────────────────────────────────────────────────────────────────────────
// DeleteInstance idempotency
// ─────────────────────────────────────────────────────────────────────────

func TestDeleteInstance_404TreatedAsSuccess(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || !strings.HasSuffix(r.URL.Path, "/delete") {
			t.Errorf("unexpected: %s %s", r.Method, r.URL.Path)
		}
		w.WriteHeader(http.StatusNotFound) // already gone
	}))
	defer srv.Close()

	c := NewClient(srv.URL, "tok", zaptest.NewLogger(t))
	if err := c.DeleteInstance(context.Background(), 42); err != nil {
		t.Errorf("DeleteInstance returned %v on 404, expected nil (idempotent)", err)
	}
}

func TestDeleteInstance_SuccessfulPath(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/instances/42/delete" {
			t.Errorf("path: got %q want /instances/42/delete", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	c := NewClient(srv.URL, "tok", zaptest.NewLogger(t))
	if err := c.DeleteInstance(context.Background(), 42); err != nil {
		t.Errorf("DeleteInstance: %v", err)
	}
}

// ─────────────────────────────────────────────────────────────────────────
// APIError classification
// ─────────────────────────────────────────────────────────────────────────

func TestAPIErrorClassification(t *testing.T) {
	cases := []struct {
		status                                    int
		isAuth, isNotFound, isRateLimit, isServer bool
	}{
		{401, true, false, false, false},
		{403, true, false, false, false},
		{404, false, true, false, false},
		{429, false, false, true, false},
		{500, false, false, false, true},
		{503, false, false, false, true},
		{400, false, false, false, false}, // none of the above
	}
	for _, tc := range cases {
		t.Run(fmt.Sprintf("status_%d", tc.status), func(t *testing.T) {
			e := &APIError{StatusCode: tc.status}
			if e.IsAuth() != tc.isAuth {
				t.Errorf("IsAuth = %v, want %v", e.IsAuth(), tc.isAuth)
			}
			if e.IsNotFound() != tc.isNotFound {
				t.Errorf("IsNotFound = %v, want %v", e.IsNotFound(), tc.isNotFound)
			}
			if e.IsRateLimit() != tc.isRateLimit {
				t.Errorf("IsRateLimit = %v, want %v", e.IsRateLimit(), tc.isRateLimit)
			}
			if e.IsServer() != tc.isServer {
				t.Errorf("IsServer = %v, want %v", e.IsServer(), tc.isServer)
			}
		})
	}
}

// ─────────────────────────────────────────────────────────────────────────
// WaitForRunning
// ─────────────────────────────────────────────────────────────────────────

func TestWaitForRunning_HappyPath(t *testing.T) {
	// Server returns PENDING twice, then RUNNING with IP populated.
	var callCount int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := atomic.AddInt32(&callCount, 1)
		status := InstanceStatusRunning
		ip := "10.0.0.42"
		if n < 3 {
			status = InstanceStatusPending
			ip = ""
		}
		json.NewEncoder(w).Encode(Instance{
			Identifier: 5, UUID: "u5", Status: status, IP: ip,
		})
	}))
	defer srv.Close()

	c := NewClient(srv.URL, "tok", zaptest.NewLogger(t))
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	inst, err := c.WaitForRunning(ctx, 5, 100*time.Millisecond)
	if err != nil {
		t.Fatalf("WaitForRunning: %v", err)
	}
	if !IsReadyStatus(inst.Status) {
		t.Errorf("status: got %q, want RUNNING", inst.Status)
	}
	if inst.IP != "10.0.0.42" {
		t.Errorf("IP: got %q want 10.0.0.42", inst.IP)
	}
	if atomic.LoadInt32(&callCount) < 3 {
		t.Errorf("expected at least 3 calls, got %d", callCount)
	}
}

func TestWaitForRunning_ContextTimeout(t *testing.T) {
	// Server always returns PENDING — WaitForRunning should hit ctx timeout.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(Instance{
			Identifier: 6, Status: InstanceStatusPending,
		})
	}))
	defer srv.Close()

	c := NewClient(srv.URL, "tok", zaptest.NewLogger(t))
	ctx, cancel := context.WithTimeout(context.Background(), 250*time.Millisecond)
	defer cancel()

	_, err := c.WaitForRunning(ctx, 6, 50*time.Millisecond)
	if err == nil {
		t.Fatal("expected timeout error, got nil")
	}
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("expected DeadlineExceeded, got %v", err)
	}
}

func TestWaitForRunning_TerminalErrorState(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(Instance{
			Identifier: 9, Status: InstanceStatusError,
		})
	}))
	defer srv.Close()

	c := NewClient(srv.URL, "tok", zaptest.NewLogger(t))
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	_, err := c.WaitForRunning(ctx, 9, 50*time.Millisecond)
	if err == nil {
		t.Fatal("expected terminal error, got nil")
	}
	if !errors.Is(err, ErrInstanceTerminal) {
		t.Errorf("expected ErrInstanceTerminal, got %v", err)
	}
}
