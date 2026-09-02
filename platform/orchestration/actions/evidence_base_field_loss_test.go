// FILE: platform/orchestration/actions/evidence_base_field_loss_test.go
//
// RFC_060 §1d/Q7 context (2026-09-02): four finance-site registers were written
// today (migrations 695/697/698/699), each carrying fact-level keys the typed
// EvidenceFact struct does not model — `rule`, `writer_line`,
// `corrects_site_citation`. writer_block_guidance_387_test.go already pins the
// SAME hazard for one specific top-level key (writer_block_guidance); this file
// generalises it to arbitrary FACT-level keys, because the landmine is about
// the shape of the round trip, not any one field's name, and two independent
// council rounds today (695's editquality HIGH; 699's bug_historian medium)
// raised exactly this — the class wants a structural answer.
//
// Proposal: the loancalculator lane (round-trip test). Discriminating-control
// requirement: lendzy, from a probe experience the same afternoon where a
// vacuous control shipped and only a second control arm caught it — a check
// with no way to fail proves nothing about the thing it claims to guard.

package actions

import (
	"context"
	"database/sql"
	"encoding/json"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
)

// unmodelledFact returns a fact map carrying the three unmodelled keys today's
// finance registers actually rely on, plus every field EvidenceFact DOES
// model, so the test cannot pass by accident of an empty/short map.
func unmodelledFact() map[string]interface{} {
	return map[string]interface{}{
		"id":          "F-CONC-6.7.23",
		"claim":       "must not refinance high-cost short-term credit on more than two occasions",
		"kind":        "capability",
		"source":      map[string]interface{}{"citation": "https://handbook.fca.org.uk/handbook/conc6/7"},
		"verified_at": "2026-09-02",
		// Unmodelled — not fields on datahelpers.EvidenceFact:
		"rule":                   "CONC 6.7.23",
		"writer_line":            "We do not refinance a loan more than twice.",
		"corrects_site_citation": "CONC 6.7.17",
	}
}

// TestUnmodelledFactKeysSurviveRefresherRoundTrip pins the PRODUCTION write
// path (writeRefreshedEvidenceBase, sqlmock against the real INSERT, not a
// shape-alike marshal in the test — the pin writer_block_guidance_387_test.go
// established). If a future commit swaps the map-based read/write for a typed
// EvidenceFact round trip, this test fails instead of five sites' `rule` and
// `writer_line` fields silently vanishing on the next scheduled refresh.
func TestUnmodelledFactKeysSurviveRefresherRoundTrip(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	siteID := uuid.New()
	rowID := uuid.New()
	eb := map[string]interface{}{
		"facts": []interface{}{unmodelledFact()},
	}

	mock.ExpectBegin()
	mock.ExpectExec("UPDATE site_specs SET is_current = false").
		WithArgs(rowID).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("INSERT INTO site_specs").
		WithArgs(siteID, jsonCarrying{
			`"rule":"CONC 6.7.23"`,
			`"writer_line":"We do not refinance a loan more than twice."`,
			`"corrects_site_citation":"CONC 6.7.17"`,
		}, sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	res := &siteRefreshResult{}
	if err := writeRefreshedEvidenceBase(context.Background(), db, siteID, rowID,
		eb, sql.NullBool{}, res, zap.NewNop()); err != nil {
		t.Fatalf("writeRefreshedEvidenceBase: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("the production write path did not carry the unmodelled fact keys: %v", err)
	}
}

// TestTypedFactRoundTripWouldLoseUnmodelledKeys is the discriminating control.
// It proves the check above has real power to fail: the SAME fact, round-
// tripped through the typed datahelpers.EvidenceFact struct instead of the
// production map, silently drops every field EvidenceFact does not declare.
// If this control ever started PASSING with the keys still present, it would
// mean EvidenceFact had grown a catch-all field and the hazard this file
// exists to guard against had already been closed a different way — worth
// noticing, not worth deleting the test over.
func TestTypedFactRoundTripWouldLoseUnmodelledKeys(t *testing.T) {
	raw, err := json.Marshal(unmodelledFact())
	if err != nil {
		t.Fatalf("marshal fixture: %v", err)
	}

	var typed datahelpers.EvidenceFact
	if err := json.Unmarshal(raw, &typed); err != nil {
		t.Fatalf("unmarshal into EvidenceFact: %v", err)
	}
	back, err := json.Marshal(typed)
	if err != nil {
		t.Fatalf("marshal EvidenceFact: %v", err)
	}

	var roundTripped map[string]interface{}
	if err := json.Unmarshal(back, &roundTripped); err != nil {
		t.Fatalf("unmarshal round-tripped json: %v", err)
	}

	for _, key := range []string{"rule", "writer_line", "corrects_site_citation"} {
		if _, present := roundTripped[key]; present {
			t.Fatalf("control did not demonstrate loss: %q survived a typed EvidenceFact round trip — "+
				"either the struct gained a catch-all field (update this file's comment) or this control is broken",
				key)
		}
	}
}
