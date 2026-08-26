package actions

// collect_external_orders_action_test.go — the money-critical branches:
// (1) no token = refuse loudly, never poll unauthenticated;
// (2) THE PAID GATE — an unpaid brief is neither queued nor acked, a paid one
//     is queued AND acked, in one batch so the filter itself is what's proved;
// (3) a domain already 'queued' is acknowledged without a second insert (the
//     lost-ack retry), per the owner's 2026-07-31 repeat-domain ruling.
// The needs_human_review branches ride the shared insertWorkItem helper,
// which carries its own tests; the council round reviews that wiring.

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gqls/agentchassis/pkg/models"
	"github.com/gqls/agentchassis/platform/orchestration/types"
	"go.uber.org/zap"
)

func collectorParams(t *testing.T, db interface{}, ordersURL string) ActionParams {
	t.Helper()
	p := ActionParams{
		Logger:           zap.NewNop(),
		ExecutionContext: &types.ExecutionContext{},
		StepConfig:       models.Step{Config: map[string]interface{}{"orders_url": ordersURL}},
		CollectedData:    map[string]interface{}{},
	}
	if sqlDB, ok := db.(interface{ Ping() error }); ok {
		_ = sqlDB
	}
	return p
}

// stubBox serves a fixed order list and records what gets acked.
func stubBox(t *testing.T, orders []map[string]string, ackedRefs *[]string) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer test-token" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		switch {
		case r.Method == http.MethodGet:
			_ = json.NewEncoder(w).Encode(map[string]any{"orders": orders})
		case r.Method == http.MethodPost && strings.HasSuffix(r.URL.Path, "/ack"):
			body, _ := io.ReadAll(r.Body)
			var req struct {
				References []string `json:"references"`
			}
			_ = json.Unmarshal(body, &req)
			*ackedRefs = append(*ackedRefs, req.References...)
			_ = json.NewEncoder(w).Encode(map[string]any{"collected": len(req.References)})
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
}

func swapCollectorHTTP(t *testing.T, srv *httptest.Server) {
	t.Helper()
	orig := collectOrdersHTTP
	collectOrdersHTTP = func() *http.Client { return srv.Client() }
	t.Cleanup(func() { collectOrdersHTTP = orig })
}

func TestCollectorRefusesWithoutToken(t *testing.T) {
	t.Setenv("WEBDESIGN_BOX_ORDERS_TOKEN", "")
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	p := collectorParams(t, db, "https://example.invalid/internal/orders")
	p.DB = db
	_, err = CollectExternalOrdersAction(context.Background(), p)
	if err == nil || !strings.Contains(err.Error(), "WEBDESIGN_BOX_ORDERS_TOKEN") {
		t.Fatalf("want a loud refusal naming the env var, got %v", err)
	}
}

// THE PAID GATE, proved as a FILTER: two briefs in one batch, one paid, one
// not. Only the paid one may reach build_queue, and only the paid one may be
// acknowledged — an unpaid brief acked would be a customer's brief silently
// discarded; an unpaid brief queued would be a free build.
func TestCollectorPaidGateFiltersTheBatch(t *testing.T) {
	t.Setenv("WEBDESIGN_BOX_ORDERS_TOKEN", "test-token")

	var acked []string
	srv := stubBox(t, []map[string]string{
		{"reference": "BR-PAID11", "contact_email": "a@example.com", "domain": "paidsite.uk",
			"brief": "A paid site brief long enough to be real."},
		{"reference": "BR-WAIT22", "contact_email": "b@example.com", "domain": "waitsite.uk",
			"brief": "An unpaid site brief long enough to be real."},
	}, &acked)
	defer srv.Close()
	swapCollectorHTTP(t, srv)

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	// BR-PAID11: paid → new domain → INSERT.
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id FROM billing_orders")).
		WithArgs("BR-PAID11").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("order-uuid-1"))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT status FROM build_queue")).
		WithArgs("paidsite.uk").
		WillReturnRows(sqlmock.NewRows([]string{"status"})) // no rows
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO build_queue")).
		WillReturnResult(sqlmock.NewResult(1, 1))
	// BR-WAIT22: no paid order → nothing else may touch the DB for it.
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id FROM billing_orders")).
		WithArgs("BR-WAIT22").
		WillReturnRows(sqlmock.NewRows([]string{"id"})) // no rows

	p := collectorParams(t, db, srv.URL+"/internal/orders")
	p.DB = db
	out, err := CollectExternalOrdersAction(context.Background(), p)
	if err != nil {
		t.Fatalf("collector: %v", err)
	}
	result, ok := out.(map[string]interface{})
	if !ok {
		t.Fatalf("unexpected result shape %#v", out)
	}
	if result["queued"] != 1 || result["awaiting_payment"] != 1 {
		t.Fatalf("queued=%v awaiting_payment=%v, want 1 and 1", result["queued"], result["awaiting_payment"])
	}
	if len(acked) != 1 || acked[0] != "BR-PAID11" {
		t.Fatalf("acked %v, want exactly [BR-PAID11] — acking an unpaid brief discards it; not acking a paid one re-collects it for ever", acked)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet DB expectations: %v", err)
	}
}

// The lost-ack retry: a paid brief whose domain is already 'queued' is
// acknowledged WITHOUT a second insert — the work is already waiting.
func TestCollectorAlreadyQueuedIsAckedWithoutInsert(t *testing.T) {
	t.Setenv("WEBDESIGN_BOX_ORDERS_TOKEN", "test-token")

	var acked []string
	srv := stubBox(t, []map[string]string{
		{"reference": "BR-RETRY33", "contact_email": "a@example.com", "domain": "paidsite.uk",
			"brief": "A paid site brief long enough to be real."},
	}, &acked)
	defer srv.Close()
	swapCollectorHTTP(t, srv)

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id FROM billing_orders")).
		WithArgs("BR-RETRY33").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("order-uuid-1"))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT status FROM build_queue")).
		WithArgs("paidsite.uk").
		WillReturnRows(sqlmock.NewRows([]string{"status"}).AddRow("queued"))
	// NO insert expectation: one appearing fails ExpectationsWereMet.

	p := collectorParams(t, db, srv.URL+"/internal/orders")
	p.DB = db
	out, err := CollectExternalOrdersAction(context.Background(), p)
	if err != nil {
		t.Fatalf("collector: %v", err)
	}
	if len(acked) != 1 || acked[0] != "BR-RETRY33" {
		t.Fatalf("acked %v, want [BR-RETRY33]", acked)
	}
	result := out.(map[string]interface{})
	if result["queued"] != 0 {
		t.Fatalf("queued=%v, want 0 — already-queued must not double-insert", result["queued"])
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet DB expectations: %v", err)
	}
}
