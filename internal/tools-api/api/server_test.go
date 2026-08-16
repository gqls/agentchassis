package api

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/gqls/agentchassis/internal/tools-api/config"
)

// TestNewRouterRegistersRoutes exists because `go build` cannot catch the failure
// mode that actually threatens this file: gin panics at REGISTRATION time when
// two routes conflict in its tree. Adding GET /round/:slug beside the existing
// POST /round is exactly the shape that can do it (a static segment and a
// wildcard child under one path), and the panic would land on the island at
// process start — i.e. a service that will not boot, discovered after a swap.
func TestNewRouterRegistersRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := &config.Config{MaxBodyBytes: 4096, RateLimitRPS: 1, RateLimitBurst: 1}

	var r *gin.Engine
	func() {
		defer func() {
			if p := recover(); p != nil {
				t.Fatalf("NewRouter panicked registering routes: %v", p)
			}
		}()
		// A nil pool is fine: NewRouter only registers handlers, and nothing here
		// dials the database until a request is served.
		r = NewRouter(nil, cfg, nil)
	}()

	want := map[string]string{
		"POST /api/v1/tools/gauntlet/round":          "create a round",
		"POST /api/v1/tools/gauntlet/position":       "counter + challenge",
		"POST /api/v1/tools/gauntlet/defend":         "verdict + reasons",
		"POST /api/v1/tools/gauntlet/publish":        "publish a completed round",
		"GET /api/v1/tools/gauntlet/round/:slug":     "public record read",
		"OPTIONS /api/v1/tools/gauntlet/publish":     "preflight for publish",
		"OPTIONS /api/v1/tools/gauntlet/round/:slug": "preflight for the record read",
		"GET /health": "health",
	}

	got := map[string]bool{}
	for _, ri := range r.Routes() {
		got[ri.Method+" "+ri.Path] = true
	}
	for route, why := range want {
		if !got[route] {
			t.Errorf("route %q (%s) is not registered", route, why)
		}
	}
}

// TestPostRoundIsNotShadowed guards the specific worry above: that introducing
// the /round/:slug child silently turns POST /round into something else. Both
// must be present and distinct.
func TestPostRoundIsNotShadowed(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := NewRouter(nil, &config.Config{MaxBodyBytes: 4096, RateLimitRPS: 1, RateLimitBurst: 1}, nil)

	var postRound, getSlug int
	for _, ri := range r.Routes() {
		if ri.Method == "POST" && ri.Path == "/api/v1/tools/gauntlet/round" {
			postRound++
		}
		if ri.Method == "GET" && ri.Path == "/api/v1/tools/gauntlet/round/:slug" {
			getSlug++
		}
	}
	if postRound != 1 {
		t.Errorf("expected exactly 1 POST /round, found %d", postRound)
	}
	if getSlug != 1 {
		t.Errorf("expected exactly 1 GET /round/:slug, found %d", getSlug)
	}
}

// TestGripperGroupIsOptIn: with cfg.Gripper nil the gripper routes must NOT
// exist (an island .env that predates the feature boots exactly as before),
// and with it set all four must — three under the browser prefix and the
// cluster's /requests beside them.
func TestGripperGroupIsOptIn(t *testing.T) {
	gin.SetMode(gin.TestMode)
	base := config.Config{MaxBodyBytes: 4096, RateLimitRPS: 1, RateLimitBurst: 1}

	off := NewRouter(nil, &base, nil)
	for _, ri := range off.Routes() {
		if strings.HasPrefix(ri.Path, "/api/v1/tools/gripper") {
			t.Errorf("gripper route %s %s registered with Gripper config nil", ri.Method, ri.Path)
		}
	}

	on := base
	on.Gripper = &config.GripperConfig{PullKey: "0123456789abcdef0123456789abcdef", MaxBodyBytes: 16384, DailyTurnCap: 10}
	deps := &GripperDeps{Store: nil, Generator: nil, Limiters: NewGripperLimiters()}
	r := NewRouter(nil, &on, deps)

	want := []string{
		"POST /api/v1/tools/gripper/session",
		"OPTIONS /api/v1/tools/gripper/session",
		"POST /api/v1/tools/gripper/chat",
		"OPTIONS /api/v1/tools/gripper/chat",
		"POST /api/v1/tools/gripper/submit",
		"OPTIONS /api/v1/tools/gripper/submit",
		"GET /api/v1/tools/gripper/requests",
	}
	got := map[string]bool{}
	for _, ri := range r.Routes() {
		got[ri.Method+" "+ri.Path] = true
	}
	for _, w := range want {
		if !got[w] {
			t.Errorf("gripper route %q not registered", w)
		}
	}
	// The gauntlet group must be untouched by mounting a second tool.
	if !got["POST /api/v1/tools/gauntlet/round"] {
		t.Errorf("gauntlet POST /round missing after mounting gripper")
	}
}

// TestGripperRequestsIsNotBehindCORS: the cluster sends no Origin header, so
// /requests must answer on X-Internal-Key alone (401 without it, never 403
// from CORS), while a browser route without an Origin must still be CORS-refused.
func TestGripperRequestsIsNotBehindCORS(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cfg := config.Config{MaxBodyBytes: 4096, RateLimitRPS: 1, RateLimitBurst: 1,
		Gripper: &config.GripperConfig{PullKey: "0123456789abcdef0123456789abcdef", MaxBodyBytes: 16384, DailyTurnCap: 10}}
	r := NewRouter(nil, &cfg, &GripperDeps{Limiters: NewGripperLimiters()})

	// No key → 401 from InternalKey, proving CORS did not run first (CORS with a
	// nil pool would panic on the DB call, and with a real pool would 403).
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/v1/tools/gripper/requests", nil)
	r.ServeHTTP(rec, req)
	if rec.Code != 401 {
		t.Fatalf("GET /requests without key: want 401, got %d body=%s", rec.Code, rec.Body.String())
	}
	rec = httptest.NewRecorder()
	req = httptest.NewRequest("GET", "/api/v1/tools/gripper/requests", nil)
	req.Header.Set("X-Internal-Key", "wrong-key-of-the-right-length!!!")
	r.ServeHTTP(rec, req)
	if rec.Code != 401 {
		t.Fatalf("GET /requests with wrong key: want 401, got %d", rec.Code)
	}
}
