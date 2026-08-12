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

	if w := factsGet(r, "webdesign.uk", ""); w.Code != http.StatusUnauthorized {
		t.Errorf("missing token: want 401, got %d", w.Code)
	}
	if w := factsGet(r, "webdesign.uk", "wrong-token"); w.Code != http.StatusUnauthorized {
		t.Errorf("wrong token: want 401, got %d", w.Code)
	}
	// No SELECT may run before auth passes — the DB not being consulted is
	// part of the contract, not an optimisation (a 401 that already queried
	// leaks timing about which domains exist).
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("DB touched on unauthenticated request: %v", err)
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
