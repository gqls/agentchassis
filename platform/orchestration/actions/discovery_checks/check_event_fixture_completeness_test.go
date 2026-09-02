// FILE: platform/orchestration/actions/discovery_checks/check_event_fixture_completeness_test.go
//
// Pure classification tests, each with a mutation that kills it (run, not
// asserted — each mutation was applied, compiled, tested to fail, then
// reverted by restoring the exact original line, never `git restore`):
//
//	M1 drop the event_date gate           -> TestNoEventDateIsNotAnEventFact
//	M2 drop the citation url/quote check  -> TestMissingCitationIsUnevidenced
//	M3 drop the venue/broadcaster/participant checks -> TestMissingOptionalFieldIsIncomplete
//	M4 the all-present case               -> TestFullyPopulatedFactIsComplete
//
// Plus one Run()-level test per outcome (file, retract, quiet) using sqlmock,
// since the site-scoped work-item wiring (gap key, capability_gap shape,
// RFC_010 retraction) is not exercised by the pure predicate alone.

package discovery_checks

import (
	"context"
	"database/sql"
	"encoding/json"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

func mkEventFact(id, eventDate string, withCitation bool, venue, broadcaster string, nParticipants int) map[string]interface{} {
	f := map[string]interface{}{"id": id, "kind": "entity"}
	if eventDate != "" {
		f["event_date"] = eventDate
	}
	if withCitation {
		f["source"] = map[string]interface{}{"citation": map[string]interface{}{
			"url": "https://example.com/" + id, "quote": "the fixture",
		}}
	}
	if venue != "" {
		f["venue"] = venue
	}
	if broadcaster != "" {
		f["broadcaster"] = broadcaster
	}
	if nParticipants > 0 {
		ps := make([]interface{}, nParticipants)
		for i := range ps {
			ps[i] = "P"
		}
		f["participants"] = ps
	}
	return f
}

// M1.
func TestNoEventDateIsNotAnEventFact(t *testing.T) {
	f := mkEventFact("F1", "", true, "Arena", "DAZN", 2)
	if got := eventFixtureVerdict(f); got != "" {
		t.Errorf("a fact with no event_date must not be classified as an event fact, got %q", got)
	}
}

// M2. The gate this whole check exists for (council REVISE, compliance HIGH's
// twin — a fact must not be treated as trustworthy without a citation).
func TestMissingCitationIsUnevidenced(t *testing.T) {
	cases := []map[string]interface{}{
		mkEventFact("F1", "2027-01-01", false, "Arena", "DAZN", 2),
		{"id": "F2", "event_date": "2027-01-01", "venue": "Arena", "broadcaster": "DAZN",
			"participants": []interface{}{"A", "B"},
			"source":       map[string]interface{}{"citation": map[string]interface{}{"url": "https://x"}}}, // quote missing
		{"id": "F3", "event_date": "2027-01-01", "venue": "Arena", "broadcaster": "DAZN",
			"participants": []interface{}{"A", "B"},
			"source":       map[string]interface{}{"citation": map[string]interface{}{"quote": "x"}}}, // url missing
	}
	for _, f := range cases {
		if got := eventFixtureVerdict(f); got != "unevidenced" {
			t.Errorf("fact %v: want unevidenced, got %q", f["id"], got)
		}
	}
}

// M3.
func TestMissingOptionalFieldIsIncomplete(t *testing.T) {
	cases := []struct {
		name string
		f    map[string]interface{}
	}{
		{"no venue", mkEventFact("F1", "2027-01-01", true, "", "DAZN", 2)},
		{"no broadcaster", mkEventFact("F2", "2027-01-01", true, "Arena", "", 2)},
		{"one participant", mkEventFact("F3", "2027-01-01", true, "Arena", "DAZN", 1)},
		{"no participants", mkEventFact("F4", "2027-01-01", true, "Arena", "DAZN", 0)},
	}
	for _, c := range cases {
		if got := eventFixtureVerdict(c.f); got != "incomplete" {
			t.Errorf("%s: want incomplete, got %q", c.name, got)
		}
	}
}

// M4. The positive control: without it, a passing check could be convicting
// everything.
func TestFullyPopulatedFactIsComplete(t *testing.T) {
	f := mkEventFact("F1", "2027-01-01", true, "Arena", "DAZN", 2)
	if got := eventFixtureVerdict(f); got != "complete" {
		t.Errorf("want complete, got %q", got)
	}
}

func newEventFixtureMock(t *testing.T) (*sql.DB, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return db, mock
}

func registerJSON(t *testing.T, facts []map[string]interface{}) []byte {
	t.Helper()
	b, err := json.Marshal(map[string]interface{}{"facts": facts})
	if err != nil {
		t.Fatalf("marshal fixture: %v", err)
	}
	return b
}

func TestRun_FilesCapabilityGapWhenIncompleteOrUnevidenced(t *testing.T) {
	db, mock := newEventFixtureMock(t)
	siteID := uuid.New()

	facts := []map[string]interface{}{
		mkEventFact("EVT-good", "2027-01-01", true, "Arena", "DAZN", 2),
		mkEventFact("EVT-thin", "2027-02-01", true, "", "", 2),
		mkEventFact("EVT-bad", "2027-03-01", false, "", "", 0),
	}
	mock.ExpectQuery("FROM site_specs").WithArgs(siteID).
		WillReturnRows(sqlmock.NewRows([]string{"data"}).AddRow(registerJSON(t, facts)))

	dctx := DiscoveryCheckContext{Ctx: context.Background(), DB: db, SiteID: siteID,
		Pipeline: "content", AgentType: "test", BatchID: uuid.New(), Logger: zap.NewNop()}
	res, err := (&EventFixtureCompletenessCheck{}).Run(dctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(res.WorkItems) != 1 {
		t.Fatalf("want exactly 1 capability_gap item, got %d: %+v", len(res.WorkItems), res.WorkItems)
	}
	wi := res.WorkItems[0]
	if wi.ItemType != "capability_gap" || wi.Status != "deferred" || wi.HandlerAgent != "" {
		t.Errorf("wrong shape for the established capability_gap convention: %+v", wi)
	}
	if wi.ItemKey != "event_fixture_completeness:"+siteID.String() {
		t.Errorf("item key = %q, want the site-scoped gap key", wi.ItemKey)
	}
	if len(res.Resolved) != 0 {
		t.Errorf("must not retract while a gap still exists, got %+v", res.Resolved)
	}
}

func TestRun_RetractsWhenEverythingIsClean(t *testing.T) {
	db, mock := newEventFixtureMock(t)
	siteID := uuid.New()

	facts := []map[string]interface{}{
		mkEventFact("EVT-good1", "2027-01-01", true, "Arena", "DAZN", 2),
		mkEventFact("EVT-good2", "2027-02-01", true, "Stadium", "ESPN", 2),
	}
	mock.ExpectQuery("FROM site_specs").WithArgs(siteID).
		WillReturnRows(sqlmock.NewRows([]string{"data"}).AddRow(registerJSON(t, facts)))

	dctx := DiscoveryCheckContext{Ctx: context.Background(), DB: db, SiteID: siteID,
		Pipeline: "content", AgentType: "test", BatchID: uuid.New(), Logger: zap.NewNop()}
	res, err := (&EventFixtureCompletenessCheck{}).Run(dctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(res.WorkItems) != 0 {
		t.Errorf("must not file anything when every fact is clean, got %+v", res.WorkItems)
	}
	if len(res.Resolved) != 1 || res.Resolved[0].ItemKey != "event_fixture_completeness:"+siteID.String() {
		t.Fatalf("want exactly one retraction for the site's gap key, got %+v", res.Resolved)
	}
}

func TestRun_NoEventFactsAtAllFilesNothingAndRetractsNothing(t *testing.T) {
	db, mock := newEventFixtureMock(t)
	siteID := uuid.New()

	// A register that exists but has no event-shaped facts at all — the
	// ordinary state for most sites. Must not retract (RFC_010: retraction is
	// a POSITIVE observation of health, never derived from "found nothing").
	facts := []map[string]interface{}{
		{"id": "F1", "kind": "metric", "value": 42.0},
	}
	mock.ExpectQuery("FROM site_specs").WithArgs(siteID).
		WillReturnRows(sqlmock.NewRows([]string{"data"}).AddRow(registerJSON(t, facts)))

	dctx := DiscoveryCheckContext{Ctx: context.Background(), DB: db, SiteID: siteID,
		Pipeline: "content", AgentType: "test", BatchID: uuid.New(), Logger: zap.NewNop()}
	res, err := (&EventFixtureCompletenessCheck{}).Run(dctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(res.WorkItems) != 0 || len(res.Resolved) != 0 {
		t.Fatalf("a site with no event facts must neither file nor retract, got work_items=%+v resolved=%+v",
			res.WorkItems, res.Resolved)
	}
}
