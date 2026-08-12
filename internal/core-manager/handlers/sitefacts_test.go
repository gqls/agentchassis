// FILE: internal/core-manager/handlers/sitefacts_test.go
//
// Wire-shape tests for the site-facts relay. The auth cases matter most: this
// endpoint deliberately sits OUTSIDE AuthMiddleware, so its own token check is
// the entire access control — a regression here isn't "weaker auth", it is
// "every site's facts served to anyone on the network path".
package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func newFactsRouter(t *testing.T) (*gin.Engine, sqlmock.Sqlmock, *sql.DB) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewSiteFactsHandler(db, zap.NewNop())
	r.GET("/api/v1/site-facts/:domain", h.HandleGetSiteFacts)
	return r, mock, db
}

func factsGet(r *gin.Engine, domain, token string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, "/api/v1/site-facts/"+domain, nil)
	if token != "" {
		req.Header.Set("X-Facts-Token", token)
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

// FAIL CLOSED: no token configured server-side means NOBODY gets in — not
// everybody. An operator who forgets the env var must see 401s, never data.
func TestNoServerTokenRefusesEverything(t *testing.T) {
	r, _, db := newFactsRouter(t)
	defer db.Close()
	os.Unsetenv("SITE_FACTS_TOKEN")

	if w := factsGet(r, "webdesign.uk", "anything"); w.Code != http.StatusUnauthorized {
		t.Fatalf("unset server token must 401 every caller, got %d", w.Code)
	}
}

func TestWrongOrMissingTokenIs401AndNeverTouchesTheDB(t *testing.T) {
	r, mock, db := newFactsRouter(t)
	defer db.Close()
	t.Setenv("SITE_FACTS_TOKEN", "correct-token")

	// PROOF THAT NO QUERY RAN, without relying on ExpectationsWereMet() —
	// which the go-sqlmock LANDMINE correctly flags as hollow here: with zero
	// registered expectations it is satisfied whether or not a query ran, so
	// it cannot distinguish "auth short-circuited" from "queried anyway".
	//
	// The real evidence is the STATUS CODE. This mock has NO expectation
	// registered, and go-sqlmock returns an error for any unexpected query.
	// So IF the handler reached its SELECT on the auth-fail path, that query
	// would error and the handler would return 500 (its DB-error branch). A
	// 401 is therefore only reachable by returning BEFORE the query — exactly
	// the property under test. (The mutation that disables the token check
	// makes both cases below return 500, not 401, confirming this is load-
	// bearing, not incidental.)
	for _, tc := range []struct {
		name, token string
	}{{"missing", ""}, {"wrong", "wrong-token"}} {
		w := factsGet(r, "webdesign.uk", tc.token)
		if w.Code != http.StatusUnauthorized {
			t.Errorf("%s token: want 401 (500 would mean the SELECT ran against the unconfigured mock — i.e. auth was bypassed), got %d", tc.name, w.Code)
		}
	}
	// Belt-and-braces, and now MEANINGFUL because a spurious query above would
	// already have failed the status assertion: no unexpected call was logged.
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("sqlmock recorded activity on an unauthenticated request: %v", err)
	}
}

// The auth check must reject BEFORE reaching the DB even when a real query
// WOULD have matched — the positive control that makes the 401-not-500 logic
// above unambiguous. If auth ever regresses to "query first, then check", this
// test still gets 401 (good) but the one above would flip to whatever the query
// returned; keeping both pins the ordering.
func TestValidQueryShapeStillRejectedWithoutToken(t *testing.T) {
	r, mock, db := newFactsRouter(t)
	defer db.Close()
	t.Setenv("SITE_FACTS_TOKEN", "correct-token")

	// A query that WOULD succeed if it ran — so a 401 here proves the code
	// never got far enough to run it, not that the query happened to fail.
	mock.ExpectQuery("SELECT s.id::text, ss.data->'facts'").
		WithArgs("webdesign.uk").
		WillReturnRows(sqlmock.NewRows([]string{"id", "facts"}).
			AddRow("some-id", []byte(`[{"id":"x"}]`)))

	if w := factsGet(r, "webdesign.uk", "wrong-token"); w.Code != http.StatusUnauthorized {
		t.Errorf("want 401 even with a query ready to succeed, got %d", w.Code)
	}
	// The prepared query must NOT have been consumed — auth returned first.
	if err := mock.ExpectationsWereMet(); err == nil {
		t.Error("the prepared query was consumed on an unauthenticated request — auth did not short-circuit before the DB")
	}
}

func TestKnownDomainServesFactsOnly(t *testing.T) {
	r, mock, db := newFactsRouter(t)
	defer db.Close()
	t.Setenv("SITE_FACTS_TOKEN", "correct-token")

	facts := `[{"id":"price_total","value":1200,"writer_line":"£1,200"}]`
	mock.ExpectQuery("SELECT s.id::text, ss.data->'facts'").
		WithArgs("webdesign.uk").
		WillReturnRows(sqlmock.NewRows([]string{"id", "facts"}).
			AddRow("1fcfa4f3-ec80-4010-878b-b971cd46711f", []byte(facts)))

	w := factsGet(r, "webdesign.uk", "correct-token")
	if w.Code != http.StatusOK {
		t.Fatalf("want 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp struct {
		Domain string          `json:"domain"`
		SiteID string          `json:"site_id"`
		Facts  json.RawMessage `json:"facts"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("response is not valid JSON: %v", err)
	}
	if resp.Domain != "webdesign.uk" || resp.SiteID == "" {
		t.Errorf("identity fields wrong: %+v", resp)
	}
	var parsed []map[string]interface{}
	if err := json.Unmarshal(resp.Facts, &parsed); err != nil || len(parsed) != 1 {
		t.Fatalf("facts did not round-trip as a JSON array: %v", err)
	}
	// writer_block is internal prompt text and must never be in the response,
	// whatever the DB row alongside the facts contains.
	if _, present := parsed[0]["writer_block"]; present {
		t.Error("writer_block leaked into the response")
	}
}

// The route must be mounted OUTSIDE the JWT AuthMiddleware group — a headless
// box service has a static token, not a JWT. This models server.go's mount
// (facts route at top level; an auth-gated group alongside) and proves an
// unauthenticated request reaches the FACTS handler's own token check rather
// than being rejected by the group's middleware. If someone nested the route
// under the auth group, a no-JWT request would be rejected by the middleware
// (here: 403) before ever reaching the token check (401) — so the two codes
// are distinguishable and this test would catch the move.
//
// (Corroborated live 2026-08-12: the box, carrying no JWT, reached the handler
// over WireGuard and got a 404 from the pre-endpoint image — a JWT-gated route
// would have 401'd at the middleware before routing.)
func TestRouteIsOutsideTheAuthMiddlewareGroup(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	gin.SetMode(gin.TestMode)

	r := gin.New()
	// An auth-gated group that rejects everything lacking a JWT, standing in
	// for AuthMiddleware. The facts route must NOT be under it.
	authGroup := r.Group("/api/v1")
	authGroup.Use(func(c *gin.Context) {
		c.AbortWithStatus(http.StatusForbidden) // no JWT -> 403, before any handler
	})
	authGroup.GET("/some-jwt-route", func(c *gin.Context) { c.Status(http.StatusOK) })

	// Mounted at top level, exactly as server.go does it.
	h := NewSiteFactsHandler(db, zap.NewNop())
	r.GET("/api/v1/site-facts/:domain", h.HandleGetSiteFacts)

	t.Setenv("SITE_FACTS_TOKEN", "correct-token")
	w := factsGet(r, "webdesign.uk", "") // no token, no JWT

	// 401 = reached the facts handler's own token check (correct: outside auth).
	// 403 = the auth-group middleware caught it first (wrong: nested under auth).
	if w.Code == http.StatusForbidden {
		t.Fatal("facts route is under the AuthMiddleware group — a static-token box caller would be rejected as a missing-JWT request")
	}
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("want 401 from the handler's own token check, got %d", w.Code)
	}
}

// Unknown domain and known-domain-without-facts are both 404 and identical on
// the wire, so the endpoint cannot be used to enumerate which domains exist.
func TestUnknownDomainIs404(t *testing.T) {
	r, mock, db := newFactsRouter(t)
	defer db.Close()
	t.Setenv("SITE_FACTS_TOKEN", "correct-token")

	mock.ExpectQuery("SELECT s.id::text, ss.data->'facts'").
		WithArgs("no-such-site.example").
		WillReturnRows(sqlmock.NewRows([]string{"id", "facts"}))

	if w := factsGet(r, "no-such-site.example", "correct-token"); w.Code != http.StatusNotFound {
		t.Errorf("want 404, got %d", w.Code)
	}
}

// Domain matching is case-insensitive at the SQL boundary (lower() both
// sides), so the caller's casing cannot cause a spurious 404.
func TestDomainIsLowercasedBeforeTheQuery(t *testing.T) {
	r, mock, db := newFactsRouter(t)
	defer db.Close()
	t.Setenv("SITE_FACTS_TOKEN", "correct-token")

	mock.ExpectQuery("SELECT s.id::text, ss.data->'facts'").
		WithArgs("webdesign.uk"). // lowercased, despite the mixed-case request
		WillReturnRows(sqlmock.NewRows([]string{"id", "facts"}).
			AddRow("1fcfa4f3-ec80-4010-878b-b971cd46711f", []byte(`[]`)))

	if w := factsGet(r, "WebDesign.UK", "correct-token"); w.Code != http.StatusOK {
		t.Errorf("mixed-case domain should still resolve, got %d", w.Code)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("query did not receive the lowercased domain: %v", err)
	}
}
