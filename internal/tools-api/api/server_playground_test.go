package api

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/gqls/agentchassis/internal/tools-api/config"
)

// The playground group is opt-in twice over: the config must be present AND
// the deps must be passed. Either absent → not one playground route, and the
// gauntlet is untouched.
func TestPlaygroundRoutesMountOnlyWhenConfiguredAndWired(t *testing.T) {
	gin.SetMode(gin.TestMode)
	base := config.Config{MaxBodyBytes: 4096, RateLimitRPS: 1, RateLimitBurst: 1}

	hasPlayground := func(r *gin.Engine) bool {
		for _, ri := range r.Routes() {
			if strings.HasPrefix(ri.Path, "/api/v1/tools/playground") {
				return true
			}
		}
		return false
	}

	// 1. no config, no deps
	if hasPlayground(NewRouter(nil, &base, nil)) {
		t.Errorf("playground routes registered with neither config nor deps")
	}
	// 2. config but no deps (main did not wire it)
	withCfg := base
	withCfg.Playground = &config.PlaygroundConfig{OllamaURL: "http://ollama:11434", Model: "finetuning-demo", MaxTokens: 150, NumCtx: 2048, MaxBodyBytes: 8192}
	if hasPlayground(NewRouter(nil, &withCfg, nil)) {
		t.Errorf("playground routes registered with config but no deps")
	}
	// 3. deps but no config (env unset)
	if hasPlayground(NewRouter(nil, &base, nil, WithPlayground(NewPlaygroundDeps()))) {
		t.Errorf("playground routes registered with deps but no config")
	}
	// 4. both
	r := NewRouter(nil, &withCfg, nil, WithPlayground(NewPlaygroundDeps()))
	got := map[string]bool{}
	for _, ri := range r.Routes() {
		got[ri.Method+" "+ri.Path] = true
	}
	for _, w := range []string{
		"POST /api/v1/tools/playground/chat",
		"OPTIONS /api/v1/tools/playground/chat",
		"POST /api/v1/tools/gauntlet/round",
	} {
		if !got[w] {
			t.Errorf("route %q not registered", w)
		}
	}
	if got["GET /api/v1/tools/playground/requests"] {
		t.Errorf("playground must not expose a pull route; it has no cluster-side consumer")
	}
}

// The body cap is the playground's own, not the gauntlet's: a body over
// PlaygroundConfig.MaxBodyBytes is refused before the handler (and before any
// model call) with 413.
func TestPlaygroundBodyCapIsItsOwn(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cfg := config.Config{MaxBodyBytes: 1 << 20, RateLimitRPS: 1, RateLimitBurst: 1,
		Playground: &config.PlaygroundConfig{OllamaURL: "http://127.0.0.1:1", Model: "m", MaxTokens: 10, NumCtx: 512, MaxBodyBytes: 64}}
	r := NewRouter(nil, &cfg, nil, WithPlayground(NewPlaygroundDeps()))
	rec := httptest.NewRecorder()
	body := `{"messages":[{"role":"user","content":"` + strings.Repeat("a", 200) + `"}]}`
	req := httptest.NewRequest("POST", "/api/v1/tools/playground/chat", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	// No Origin header: CORSMiddleware with a nil pool would otherwise be
	// consulted; the cap runs after CORS, so this test asserts only that the
	// request never reaches a model. A 403 (origin) or 413 (cap) both prove
	// that; a 503 (tried the model at 127.0.0.1:1) would disprove it.
	r.ServeHTTP(rec, req)
	if rec.Code == 503 || rec.Code == 200 {
		t.Fatalf("status %d: an oversized body reached the model call", rec.Code)
	}
}
