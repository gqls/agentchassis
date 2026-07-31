package api

import (
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
		r = NewRouter(nil, cfg)
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
	r := NewRouter(nil, &config.Config{MaxBodyBytes: 4096, RateLimitRPS: 1, RateLimitBurst: 1})

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
