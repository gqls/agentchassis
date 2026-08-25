package api

// Tests for the delivery-only listener (RFC_054 Q2, owner ruling 2026-08-25).
//
// The change this file guards is a CONTAINMENT: the admin port serves no
// customer-facing route, so widening the box's nginx location cannot reach the
// API that holds every site's data. Three properties carry that, and each is
// asserted here rather than described:
//
//  1. the delivery engine serves the delivery routes and NOTHING else;
//  2. the opt-in defaults OFF — an unset port serves the routes nowhere;
//  3. a delivery route mounted on the main router refuses to build the server.
//
// Property 3 is the one that has to survive future sessions, because it is the
// only one a well-meaning edit can undo silently.

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/gqls/agentchassis/internal/core-manager/handlers"
	"go.uber.org/zap"
)

// stubDeliveryDeps drives the HTTP layer without a database. The handler
// package's own tests cover what ConfirmTransfer does; these cover which port
// can reach it at all.
type stubDeliveryDeps struct {
	logger *zap.Logger
	calls  int
}

func (s *stubDeliveryDeps) ConfirmTransfer(c *gin.Context, token string) error {
	s.calls++
	return nil
}

func (s *stubDeliveryDeps) Logger() *zap.Logger { return s.logger }

func newTestDeliveryHandler() (*handlers.DeliveryHandler, *stubDeliveryDeps) {
	deps := &stubDeliveryDeps{logger: zap.NewNop()}
	return handlers.NewDeliveryHandler(deps), deps
}

func init() { gin.SetMode(gin.TestMode) }

func mustDeliveryEngine(t *testing.T, h *handlers.DeliveryHandler) *gin.Engine {
	t.Helper()
	e, err := newDeliveryEngine(h)
	if err != nil {
		t.Fatalf("newDeliveryEngine: %v", err)
	}
	return e
}

// TestDeliveryEngineServesDeliveryRoutesOnly is the "delivery-only" claim,
// asserted as an exact route table rather than as a sample of paths that happen
// to 404. A sample cannot fail when someone adds a route; a count can.
func TestDeliveryEngineServesDeliveryRoutesOnly(t *testing.T) {
	h, _ := newTestDeliveryHandler()
	engine := mustDeliveryEngine(t, h)

	got := map[string]bool{}
	for _, r := range engine.Routes() {
		got[r.Method+" "+r.Path] = true
	}

	want := []string{"GET /c/:token", "POST /c/:token"}
	for _, w := range want {
		if !got[w] {
			t.Errorf("delivery engine is missing route %q; has %v", w, got)
		}
	}
	if len(got) != len(want) {
		t.Errorf("delivery engine serves %d routes, want exactly %d (%v).\n"+
			"A route added here gives back part of what this listener bought: "+
			"the box proxies to this port, so anything mounted here is reachable "+
			"from the internet.", len(got), len(want), got)
	}
}

// TestDeliveryEngineDoesNotServeAdminPaths is the containment stated the way an
// attacker would test it: the paths that matter on the admin port must not
// answer here.
func TestDeliveryEngineDoesNotServeAdminPaths(t *testing.T) {
	h, _ := newTestDeliveryHandler()
	engine := mustDeliveryEngine(t, h)

	for _, path := range []string{
		"/health",
		"/api/v1/admin/work-items",
		"/api/v1/admin/clients",
		"/api/v1/agents/bootstrap",
		"/api/v1/site-facts/webdesign.uk",
		"/",
	} {
		rec := httptest.NewRecorder()
		engine.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))
		if rec.Code != http.StatusNotFound {
			t.Errorf("delivery listener answered %s with %d, want 404: this port is "+
				"proxied from the public internet", path, rec.Code)
		}
	}
}

// TestDeliveryEngineReachesTheHandler is the control for the test above. Without
// it, a delivery engine that served nothing at all — the way a typo'd route
// would — would pass every 404 assertion and look like perfect containment.
func TestDeliveryEngineReachesTheHandler(t *testing.T) {
	h, deps := newTestDeliveryHandler()
	engine := mustDeliveryEngine(t, h)

	const token = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa" // 43 chars, token-shaped

	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/c/"+token, nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("GET /c/<token> on the delivery listener returned %d, want 200", rec.Code)
	}
	if deps.calls != 0 {
		t.Errorf("GET reached the database %d times, want 0 (the mail-scanner "+
			"mitigation: GET renders, POST confirms)", deps.calls)
	}

	rec = httptest.NewRecorder()
	engine.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/c/"+token, nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("POST /c/<token> on the delivery listener returned %d, want 200", rec.Code)
	}
	if deps.calls != 1 {
		t.Errorf("POST confirmed %d times, want 1", deps.calls)
	}
}

// TestDeliveryListenerIsOptInAndDefaultsOff asserts the direction the empty case
// falls. If this ever inverts, an unconfigured deployment starts a public
// listener, which is the failure the opt-in exists to make impossible.
func TestDeliveryListenerIsOptInAndDefaultsOff(t *testing.T) {
	h, _ := newTestDeliveryHandler()

	srvOff, err := newDeliveryServer("", h)
	if err != nil {
		t.Fatalf("newDeliveryServer(\"\"): %v", err)
	}
	if srv := srvOff; srv != nil {
		t.Errorf("an unset delivery port produced a listener on %q; the default "+
			"must be OFF (CLAUDE.md 2026-08-02 §2: new authority on a shared seam "+
			"ships opt-in with the unsafe default off)", srv.Addr)
	}

	srv, err := newDeliveryServer("8090", h)
	if err != nil {
		t.Fatalf("newDeliveryServer: %v", err)
	}
	if srv == nil {
		t.Fatal("a configured delivery port produced no listener")
	}
	if srv.Addr != ":8090" {
		t.Errorf("delivery listener bound %q, want %q", srv.Addr, ":8090")
	}
	if srv.Handler == nil {
		t.Error("delivery listener has no handler: it would answer nothing on a port the box proxies to")
	}
}

// TestAssertNoDeliveryRoutesRejectsCustomerRoutesOnTheAdminRouter is property 3.
//
// This is the assertion that has to outlive everyone currently working on this
// lane: it is what turns "the admin port carries no customer route" from a
// comment into something the process refuses to start without.
func TestAssertNoDeliveryRoutesRejectsCustomerRoutesOnTheAdminRouter(t *testing.T) {
	clean := gin.RoutesInfo{
		{Method: http.MethodGet, Path: "/health"},
		{Method: http.MethodGet, Path: "/api/v1/admin/work-items"},
		{Method: http.MethodPost, Path: "/api/v1/agents/bootstrap"},
		{Method: http.MethodGet, Path: "/api/v1/site-facts/:domain"},
	}
	if err := assertNoDeliveryRoutes(clean); err != nil {
		t.Fatalf("a router with no delivery routes was rejected: %v", err)
	}

	for _, tc := range []struct {
		name  string
		route gin.RouteInfo
	}{
		{"confirm page re-mounted", gin.RouteInfo{Method: http.MethodGet, Path: "/c/:token"}},
		{"confirm POST re-mounted", gin.RouteInfo{Method: http.MethodPost, Path: "/c/:token"}},
		{"download route added", gin.RouteInfo{Method: http.MethodGet, Path: "/d/:token"}},
		{"root catch-all captures /c/ by prefix", gin.RouteInfo{Method: http.MethodGet, Path: "/*any"}},
		{"root catch-all, other param name", gin.RouteInfo{Method: http.MethodPost, Path: "/*filepath"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if err := assertNoDeliveryRoutes(append(clean, tc.route)); err == nil {
				t.Errorf("%s %s was accepted on the main admin router; it must refuse "+
					"to build the server", tc.route.Method, tc.route.Path)
			}
		})
	}

	// The wildcard arm must not be a blunt "no wildcards": a wildcard that cannot
	// reach a delivery path is legitimate, and refusing it would make this guard
	// something people route around rather than keep.
	for _, ok := range []gin.RouteInfo{
		{Method: http.MethodGet, Path: "/api/*path"},
		{Method: http.MethodGet, Path: "/static/*filepath"},
	} {
		if err := assertNoDeliveryRoutes(append(clean, ok)); err != nil {
			t.Errorf("wildcard %s %s cannot capture /c/ or /d/ and must be allowed: %v",
				ok.Method, ok.Path, err)
		}
	}
}

// TestMainRouterCarriesNoDeliveryRoute is the same property checked against the
// route table the service actually builds, rather than a synthetic one.
//
// It cannot call NewServer — that needs a database and a Kafka broker — so it
// asserts the next best thing that is still real: registering the delivery
// routes is what trips the guard, so a router built the way setupRoutes builds
// one, plus the delivery handler, must be rejected. If someone re-adds
// RegisterRoutes to setupRoutes, NewServer takes this path and returns an error.
func TestMainRouterCarriesNoDeliveryRoute(t *testing.T) {
	h, _ := newTestDeliveryHandler()

	main := gin.New()
	main.GET("/health", func(c *gin.Context) {})
	if err := assertNoDeliveryRoutes(main.Routes()); err != nil {
		t.Fatalf("baseline admin router rejected: %v", err)
	}

	h.RegisterRoutes(main)
	if err := assertNoDeliveryRoutes(main.Routes()); err == nil {
		t.Error("the admin router accepted the delivery routes; NewServer would " +
			"start with customer routes on the port that serves every site's data")
	}
}

// TestDeliveryRoutesMustBeServableThroughTheBox is the landmine, asserted.
//
// The box proxies ONE anchored regex per prefix. A suffix route compiles,
// registers, and serves perfectly from inside the cluster — and is 404'd at the
// box, where there is no log line, no metric and no error to find. So the shape
// is checked at construction, and GET/POST parity falls out of it: both verbs can
// only ever be the one canonical path, which is what the second-click design
// requires, since there the METHOD is the security distinction.
func TestDeliveryRoutesMustBeServableThroughTheBox(t *testing.T) {
	canonical := gin.RoutesInfo{
		{Method: http.MethodGet, Path: "/c/:token"},
		{Method: http.MethodPost, Path: "/c/:token"},
	}
	if err := assertRoutesAreBoxServable(canonical); err != nil {
		t.Fatalf("the canonical route pair was rejected: %v", err)
	}
	// /d/ is not registered yet, but when it is, this shape must pass.
	if err := assertRoutesAreBoxServable(gin.RoutesInfo{{Method: http.MethodGet, Path: "/d/:token"}}); err != nil {
		t.Errorf("the download route's canonical shape was rejected: %v", err)
	}

	for _, tc := range []struct {
		name  string
		route gin.RouteInfo
	}{
		// The exact shape LANDMINES warns about: "POST /c/<token> must live on
		// the SAME path as the GET — a suffix route compiles, passes every test,
		// and 404s."
		{"the suffix route from the landmine", gin.RouteInfo{Method: http.MethodPost, Path: "/c/:token/confirm"}},
		{"a nested segment", gin.RouteInfo{Method: http.MethodGet, Path: "/c/confirm/:token"}},
		{"a differently named param", gin.RouteInfo{Method: http.MethodGet, Path: "/c/:id"}},
		{"a bare prefix with no token", gin.RouteInfo{Method: http.MethodGet, Path: "/c/"}},
		{"an unrelated route smuggled onto the listener", gin.RouteInfo{Method: http.MethodGet, Path: "/health"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if err := assertRoutesAreBoxServable(gin.RoutesInfo{tc.route}); err == nil {
				t.Errorf("%s %s was accepted, but the box would answer it with a local "+
					"404 that never reaches this service", tc.route.Method, tc.route.Path)
			}
		})
	}
}

// The parity property stated directly, because it is the one a future edit is
// most likely to break: registering the POST at a suffix while leaving the GET
// alone passes every inside-the-cluster test and breaks only real customers.
func TestGetAndPostShareOneDeliveryPath(t *testing.T) {
	h, _ := newTestDeliveryHandler()
	engine := mustDeliveryEngine(t, h)

	byMethod := map[string]string{}
	for _, r := range engine.Routes() {
		byMethod[r.Method] = r.Path
	}
	get, post := byMethod[http.MethodGet], byMethod[http.MethodPost]
	if get == "" || post == "" {
		t.Fatalf("expected both verbs registered, got %v", byMethod)
	}
	if get != post {
		t.Errorf("GET is %q and POST is %q: they must be the SAME path. The box "+
			"admits one anchored regex, and the second-click design makes the METHOD "+
			"the security distinction, so differing paths break the customer flow "+
			"while every test in this package still passes", get, post)
	}
}

// badRegistrar registers a route the box could never serve. It exists to prove
// the assertion is WIRED into newDeliveryEngine, not merely present in the file.
type badRegistrar struct{ path string }

func (b badRegistrar) RegisterRoutes(r gin.IRouter) {
	r.POST(b.path, func(c *gin.Context) {})
}

// TestNewDeliveryEngineRefusesAnUnservableRouteTable is the WIRING test.
//
// Without it, deleting the assertRoutesAreBoxServable call from
// newDeliveryEngine left every other test in this file passing — the direct-call
// tests proved the function worked and nothing proved it ran. That is the same
// shape as the config-validation mutation that passed in platform/delivery, and
// it is why "the guard exists" and "the guard is wired" are two claims.
func TestNewDeliveryEngineRefusesAnUnservableRouteTable(t *testing.T) {
	if _, err := newDeliveryEngine(badRegistrar{path: "/c/:token/confirm"}); err == nil {
		t.Error("newDeliveryEngine built an engine whose POST sits at a suffix path: " +
			"the box would 404 it locally and the handler would never be reached")
	}

	// And the listener must refuse to exist at all, so a bad table cannot reach
	// a port the box proxies to.
	if _, err := newDeliveryServer("8090", badRegistrar{path: "/c/:token/confirm"}); err == nil {
		t.Error("newDeliveryServer returned a listener for an unservable route table")
	}

	// Control: a canonical table must still build, or this test would pass for
	// a build that refuses everything.
	if _, err := newDeliveryEngine(badRegistrar{path: "/c/:token"}); err != nil {
		t.Errorf("the canonical path was refused: %v", err)
	}
}
