// FILE: platform/orchestration/actions/feed_event_extraction_fields_test.go
//
// bugs_open/427 fix candidate #1: the ONLY change to VerifyAndRegisterCitationsAction
// itself is four names added to the field pass-through allowlist
// (evidence_citations.go). This test exercises that allowlist end-to-end so a
// future edit to the list (or a revert of it) shows up as a failing assertion
// rather than a silent behaviour change discovered downstream in a rendered page.
//
// Also pins the kind: bugs_open/427's peer session caught, before this was
// committed, that "event" is not in EvidenceFact.Kind's closed vocabulary
// (claims.go) and would have silently demoted every fact to "metric". This
// test uses kind="entity" (datahelpers.FactKindEntity) — the corrected value —
// so a regression back to "event" is visible here, not just in a warning log
// on every subsequent build.
package actions

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/gqls/agentchassis/pkg/models"
	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"github.com/gqls/agentchassis/platform/orchestration/types"
	"go.uber.org/zap"
)

// registeredEventFactHas is an sqlmock.Argument that decodes the marshalled
// site_specs.data JSON and checks the single registered fact's kind and
// event-shaped fields — a real JSON decode, not a substring check, for the
// same reason write_build_items_routing_test.go's specHandlerIs is one: a
// substring would match any register merely containing these words.
type registeredEventFactHas struct {
	kind        string
	eventDate   string
	venue       string
	broadcaster string
}

func (want registeredEventFactHas) Match(v driver.Value) bool {
	var raw []byte
	switch t := v.(type) {
	case []byte:
		raw = t
	case string:
		raw = []byte(t)
	default:
		return false
	}
	var eb struct {
		Facts []map[string]interface{} `json:"facts"`
	}
	if err := json.Unmarshal(raw, &eb); err != nil || len(eb.Facts) != 1 {
		return false
	}
	f := eb.Facts[0]
	if k, _ := f["kind"].(string); k != want.kind {
		return false
	}
	if d, _ := f["event_date"].(string); d != want.eventDate {
		return false
	}
	if vn, _ := f["venue"].(string); vn != want.venue {
		return false
	}
	if b, _ := f["broadcaster"].(string); b != want.broadcaster {
		return false
	}
	participants, ok := f["participants"].([]interface{})
	return ok && len(participants) == 2
}

// MUTATION THAT MUST BREAK IT: remove any of "event_date"/"venue"/
// "participants"/"broadcaster" from evidence_citations.go's field
// pass-through loop, or change the candidate's "kind" back to "event" —
// either way the matcher above fails to find a fact matching what was
// actually written.
func TestVerifyAndRegisterCitations_EventFieldsPassThrough(t *testing.T) {
	quote := "the WBC title fight is confirmed for 14 March 2026 at The O2, London"
	mux := http.NewServeMux()
	mux.HandleFunc("/fight-announced", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(`<html><body><p>` + quote + `</p></body></html>`))
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	siteID := uuid.New()

	// No existing register — the find-or-create path, same as a site with no
	// evidence_base row yet (bug 427 §3: 34 of 54 sites are in this state).
	mock.ExpectQuery(`SELECT id, data, COALESCE\(pinned, false\) FROM site_specs`).
		WithArgs(siteID.String()).
		WillReturnError(sql.ErrNoRows)

	mock.ExpectQuery(`SELECT to_char\(now\(\), 'YYYY-MM-DD'\)`).
		WillReturnRows(sqlmock.NewRows([]string{"to_char"}).AddRow("2026-09-02"))

	mock.ExpectBegin()
	mock.ExpectExec(`INSERT INTO site_specs`).
		WithArgs(siteID.String(), registeredEventFactHas{
			kind:        datahelpers.FactKindEntity,
			eventDate:   "2026-03-14",
			venue:       "The O2, London",
			broadcaster: "DAZN",
		}, sqlmock.AnyArg(), true).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	candidate := map[string]interface{}{
		"claim":        "The WBC title fight is confirmed for 14 March 2026 at The O2, London.",
		"kind":         datahelpers.FactKindEntity,
		"quote":        quote,
		"url":          srv.URL + "/fight-announced",
		"publisher":    "Example Wire",
		"title":        "Title fight confirmed",
		"event_date":   "2026-03-14",
		"venue":        "The O2, London",
		"participants": []interface{}{"Fighter A", "Fighter B"},
		"broadcaster":  "DAZN",
	}

	out, err := VerifyAndRegisterCitationsAction(context.Background(), ActionParams{
		DB:               db,
		Logger:           zap.NewNop(),
		ExecutionContext: &types.ExecutionContext{},
		StepConfig: models.Step{Config: map[string]interface{}{
			"candidates": "extracted.candidates",
		}},
		CollectedData: map[string]interface{}{
			"site_id": siteID.String(),
			"extracted": map[string]interface{}{
				"candidates": []interface{}{candidate},
			},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	m, ok := out.(map[string]interface{})
	if !ok {
		t.Fatalf("unexpected result shape %#v", out)
	}
	registered, _ := m["registered"].([]string)
	if len(registered) != 1 {
		t.Fatalf("registered = %#v, want exactly one fact", m["registered"])
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet or unexpected queries (the INSERT argument matcher is what actually "+
			"proves the fields/kind were written correctly): %v", err)
	}
}
