// FILE: platform/orchestration/actions/request_render_audit_truncation_test.go
//
// Regression tests for bugs_open/242: a render audit truncated by max_pages
// must be distinguishable from a complete one. The step AWAITS, so its own
// result never survives the park (RFC_012 addendum 2) — the two things that
// must therefore happen at dispatch time are (1) the request payload carries
// pages_total/truncated for the adapter to echo back in the reply summary,
// and (2) the truncation lands as a durable agent_error_log row BEFORE the
// dispatch, so a failed send cannot unrecord it.
package actions

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/gqls/agentchassis/pkg/models"
	"github.com/gqls/agentchassis/platform/orchestration/types"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
)

// orderCheckingProducer fails the test if the produce happens while any DB
// expectation is still unmet — i.e. it pins "the durable row lands BEFORE the
// dispatch", not merely "both happened".
type orderCheckingProducer struct {
	capturingProducer
	t    *testing.T
	mock sqlmock.Sqlmock
}

func (p *orderCheckingProducer) ProduceWithValidation(ctx context.Context, topic string, headers map[string]string, key, value []byte) error {
	if err := p.mock.ExpectationsWereMet(); err != nil {
		p.t.Errorf("produce ran before all DB writes: %v — the truncation row must land BEFORE the dispatch", err)
	}
	return p.capturingProducer.ProduceWithValidation(ctx, topic, headers, key, value)
}

func renderAuditParams(db ActionParams, siteID string) ActionParams {
	p := db
	p.ExecutionContext = &types.ExecutionContext{
		CorrelationID:   "corr-1",
		OrchestrationID: "orch-1",
		ClientID:        "client-1",
		StepName:        "audit",
		ResponsesTopic:  "system.agent.test.responses",
		Sender:          types.AgentIdentity{AgentType: "render-audit-agent", PodName: "pod-1"},
	}
	p.CollectedData = map[string]interface{}{
		"site_record": map[string]interface{}{
			"site_id": siteID,
			"domain":  "example.com",
		},
	}
	return p
}

func TestRequestRenderAuditTruncationTravelsInRequestAndLandsDurably(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()
	siteID := uuid.New().String()

	// Three live pages, max_pages 2 — the cap bites.
	mock.ExpectQuery("FROM pages").
		WillReturnRows(sqlmock.NewRows([]string{"url"}).
			AddRow("/index.html").AddRow("/about.html").AddRow("/pricing.html"))
	// The durable record, and it must precede the produce (orderCheckingProducer).
	mock.ExpectExec("INSERT INTO agent_error_log").
		WillReturnResult(sqlmock.NewResult(0, 1))

	producer := &orderCheckingProducer{t: t, mock: mock}
	params := renderAuditParams(ActionParams{DB: db, Logger: zap.NewNop(), Producer: producer}, siteID)
	params.StepConfig = models.Step{Config: map[string]interface{}{"max_pages": 2}}

	if _, err := RequestRenderAuditAction(context.Background(), params); err != nil {
		t.Fatalf("action failed: %v", err)
	}

	data := producedSearchData(t, &producer.capturingProducer)
	if got, want := data["pages_total"], float64(3); got != want {
		t.Errorf("request pages_total = %v, want %v — the adapter cannot echo a total it never received", got, want)
	}
	if data["truncated"] != true {
		t.Errorf("request truncated = %v, want true", data["truncated"])
	}
	if urls, ok := data["urls"].([]interface{}); !ok || len(urls) != 2 {
		t.Errorf("urls = %v, want the 2 capped entries", data["urls"])
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestRequestRenderAuditNoTruncationWritesNoRow(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()
	siteID := uuid.New().String()

	// Two pages under a cap of 25 — no truncation, so NO agent_error_log
	// expectation is registered. agenterrors.Write is best-effort and would
	// swallow sqlmock's unexpected-call error, so the absence is asserted via
	// the logger instead: an attempted write against this mock MUST produce
	// the "Failed to write to agent_error_log" warn, therefore no such record
	// means no write was attempted (the no-op case checked, not assumed).
	mock.ExpectQuery("FROM pages").
		WillReturnRows(sqlmock.NewRows([]string{"url"}).
			AddRow("/index.html").AddRow("/about.html"))

	core, logs := observer.New(zap.WarnLevel)
	producer := &capturingProducer{}
	params := renderAuditParams(ActionParams{DB: db, Logger: zap.New(core), Producer: producer}, siteID)
	params.StepConfig = models.Step{Config: map[string]interface{}{}}

	if _, err := RequestRenderAuditAction(context.Background(), params); err != nil {
		t.Fatalf("action failed: %v", err)
	}

	data := producedSearchData(t, producer)
	if data["truncated"] != false {
		t.Errorf("request truncated = %v, want false", data["truncated"])
	}
	if got, want := data["pages_total"], float64(2); got != want {
		t.Errorf("request pages_total = %v, want %v — the total is stated even when nothing was cut", got, want)
	}
	for _, entry := range logs.All() {
		if entry.Message == "Failed to write to agent_error_log — the disposition stands but its durable trace is lost" {
			t.Errorf("a truncation row was attempted on an untruncated run")
		}
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}
