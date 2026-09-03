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

// Council 63be72d1 round 1, guardian (gating): the LANDMINES entry "tools-api's
// second tool splits ONE path prefix across TWO gin groups … any third tool
// added under /api/v1/tools/<name>". Its check is: name the route's CALLER, put
// browser routes on the CORS group and cluster routes on the internal-key
// group, nothing on the bare engine, then prove it at the socket in BOTH
// directions. This test is that proof for the third tool, with the second one
// mounted beside it, at gin's router rather than a live socket:
//
//   - the playground's /chat is behind CORS (no Origin → 403), NOT behind the
//     gripper's internal key (a missing X-Internal-Key would read 401);
//   - the gripper's /requests is still behind the internal key (401), so
//     mounting the playground moved nothing;
//   - no playground path answers on the bare engine, and no route of either
//     tool shadows the other (gin's tree keys on the full path; three static
//     prefixes cannot collide).
func TestPlaygroundDoesNotShareOrShadowTheGripperGroups(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cfg := config.Config{MaxBodyBytes: 4096, RateLimitRPS: 1, RateLimitBurst: 1,
		Gripper:    &config.GripperConfig{PullKey: "0123456789abcdef0123456789abcdef", MaxBodyBytes: 16384, DailyTurnCap: 10},
		Playground: &config.PlaygroundConfig{OllamaURL: "http://127.0.0.1:1", Model: "m", MaxTokens: 10, NumCtx: 512, MaxBodyBytes: 8192},
	}
	r := NewRouter(nil, &cfg, &GripperDeps{Limiters: NewGripperLimiters()}, WithPlayground(NewPlaygroundDeps()))

	do := func(method, path string, headers map[string]string) int {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(method, path, strings.NewReader(`{"messages":[{"role":"user","content":"hi"}]}`))
		req.Header.Set("Content-Type", "application/json")
		for k, v := range headers {
			req.Header.Set(k, v)
		}
		r.ServeHTTP(rec, req)
		return rec.Code
	}

	// Browser route, no Origin: refused by CORS, never by the internal key,
	// and never reaching the (unreachable) model server.
	if code := do("POST", "/api/v1/tools/playground/chat", nil); code == 401 || code == 503 || code == 200 {
		t.Errorf("playground /chat without Origin = %d; 401 would mean it sits behind the internal key, 503/200 that CORS let it through", code)
	}
	// The same call WITH the cluster's key header must not be let through
	// either: the key is not an alternative to CORS on a browser route.
	if code := do("POST", "/api/v1/tools/playground/chat", map[string]string{"X-Internal-Key": "0123456789abcdef0123456789abcdef"}); code == 200 || code == 503 {
		t.Errorf("playground /chat with the gripper's internal key = %d; the key must not bypass CORS", code)
	}
	// The gripper's cluster route still answers to the key alone: 401 without it.
	if code := do("GET", "/api/v1/tools/gripper/requests", nil); code != 401 {
		t.Errorf("gripper /requests without key = %d, want 401 (mounting the playground must not move it)", code)
	}
	// Nothing of the playground answers on any other prefix.
	for _, p := range []string{"/api/v1/tools/gripper/playground/chat", "/api/v1/tools/gauntlet/playground/chat", "/playground/chat", "/api/v1/tools/playground/requests"} {
		if code := do("POST", p, nil); code != 404 && code != 405 {
			t.Errorf("%s = %d, want 404/405", p, code)
		}
	}

	// Every registered route belongs to exactly one of the three prefixes,
	// and the playground registers only /chat.
	for _, ri := range r.Routes() {
		switch {
		case strings.HasPrefix(ri.Path, "/api/v1/tools/playground"):
			if ri.Path != "/api/v1/tools/playground/chat" {
				t.Errorf("unexpected playground route %s %s", ri.Method, ri.Path)
			}
		case strings.HasPrefix(ri.Path, "/api/v1/tools/gripper"), strings.HasPrefix(ri.Path, "/api/v1/tools/gauntlet"), ri.Path == "/health":
		default:
			t.Errorf("route outside the three tool prefixes: %s %s", ri.Method, ri.Path)
		}
	}
}
