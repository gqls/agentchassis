// Tests for D1 half two (bugs_open/213): site-scoped retraction of
// dark_section_audit findings after N consecutive silences.
//
// EVERY TEST HERE IS A CONTROL FOR A SPECIFIC WRONG VERSION, because a
// retraction that only ever confirms "the finding is gone, close it" passes
// against an implementation that closes far too much. The wrong versions each
// of these kills, in order: retract on the FIRST silence; treat an
// unrecognised reply as silence; count a run as silent because it FILED
// nothing; close another producer's rows; write to rows the stale reaper is
// watching.

package actions

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/gqls/agentchassis/pkg/models"
	"github.com/gqls/agentchassis/platform/orchestration/types"
	"go.uber.org/zap"
)

// argJSONContains matches a JSON-ish argument containing a substring. Used to
// pin what a streak write actually stores.
type argJSONContains struct{ substr string }

func (a argJSONContains) Match(v driver.Value) bool {
	switch s := v.(type) {
	case string:
		return strings.Contains(s, a.substr)
	case []byte:
		return strings.Contains(string(s), a.substr)
	}
	return false
}

// auditParams builds the live shape: config values are DOT-PATHS into
// collected_data, never literals (bugs_open/264 — audit_source is Required with
// no default, and migration 399 gives every auditor the
// audit_source_literal.audit_source path used here).
func auditParams(db *sql.DB, siteID uuid.UUID, findings interface{}) ActionParams {
	return ActionParams{
		DB:               db,
		Logger:           zap.NewNop(),
		ExecutionContext: &types.ExecutionContext{},
		StepConfig: models.Step{Config: map[string]interface{}{
			"site_id":      "site_record.site_id",
			"audit_source": "audit_source_literal.audit_source",
		}},
		CollectedData: map[string]interface{}{
			"site_record":          map[string]interface{}{"site_id": siteID.String()},
			"audit_source_literal": map[string]interface{}{"audit_source": "visual-design-audit"},
			"audit_result":         map[string]interface{}{"result": findings},
		},
	}
}

// darkSectionFinding is a finding that classifies as dark_section_audit.
var darkSectionFinding = map[string]interface{}{
	"category":    "dark_section",
	"severity":    "high",
	"description": "the hero section renders dark text on a dark band",
	"page":        "index",
}

// expectPagesLoad pins loadSitePages, which runs before anything else.
func expectPagesLoad(mock sqlmock.Sqlmock, siteID uuid.UUID) {
	mock.ExpectQuery("FROM pages").
		WithArgs(siteID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "page_type", "sections"}))
}

// candidateRows builds the shape loadAuditRetractionCandidates scans.
func candidateRows() *sqlmock.Rows {
	return sqlmock.NewRows([]string{"id", "item_key", "status", "spec", "result"})
}

const ourSpec = `{"audit_source":"visual-design-audit","page_name":"index"}`

// ---------------------------------------------------------------------------
// THE CENTRAL CONTROL: one silence is not enough, and three is.
// ---------------------------------------------------------------------------

// A silent run against a fresh row must BUMP, never retract. This is the test
// that fails if N is quietly dropped to 1 — which is WII-016's shape and is
// wrong here, because that producer measures and this one asks an LLM.
func TestAuditRetraction_FirstSilenceBumpsAndDoesNotRetract(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()
	siteID := uuid.New()
	itemID := uuid.New()

	expectPagesLoad(mock, siteID)
	mock.ExpectBegin()
	mock.ExpectQuery("item_type = 'dark_section_audit'").
		WithArgs(siteID).
		WillReturnRows(candidateRows().
			AddRow(itemID, "visual-design-audit_dark_section_audit_index_x", "failed", ourSpec, `{}`))
	// The streak write, and it must carry the count AND the source: a streak
	// with no source attached would survive a producer rename and let two
	// producers' silences add up.
	mock.ExpectExec("UPDATE site_work_items").
		WithArgs(itemID, 1, "visual-design-audit").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	out, err := WriteAuditFindingsAction(context.Background(), auditParams(db, siteID, []interface{}{}))
	if err != nil {
		t.Fatalf("action failed: %v", err)
	}
	r := retractionFor(t, out, "dark_section_audit")
	if !r.Silent || r.StreaksBumped != 1 || r.Retracted != 0 {
		t.Fatalf("want silent, 1 bump, 0 retracted; got %+v", r)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

// The third silence closes it. Note the row arrives already carrying
// silent_runs=2, which is how the streak actually accumulates across runs.
func TestAuditRetraction_ThirdSilenceRetracts(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()
	siteID := uuid.New()
	itemID := uuid.New()
	key := "visual-design-audit_dark_section_audit_index_x"

	expectPagesLoad(mock, siteID)
	mock.ExpectBegin()
	mock.ExpectQuery("item_type = 'dark_section_audit'").
		WithArgs(siteID).
		WillReturnRows(candidateRows().AddRow(itemID, key, "failed", ourSpec,
			`{"retraction":{"silent_runs":2,"audit_source":"visual-design-audit"}}`))
	// resolveWorkItems' args are (check, reason, site, item_type, item_key, batch).
	// The reason must NAME the streak — a closed row whose stated cause does not
	// say what was observed is indistinguishable later from a hand closure.
	mock.ExpectExec("UPDATE site_work_items").
		WithArgs("visual-design-audit", argJSONContains{"3 consecutive runs"}, siteID,
			"dark_section_audit", key, sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	out, err := WriteAuditFindingsAction(context.Background(), auditParams(db, siteID, []interface{}{}))
	if err != nil {
		t.Fatalf("action failed: %v", err)
	}
	r := retractionFor(t, out, "dark_section_audit")
	if r.Retracted != 1 || r.StreaksBumped != 0 {
		t.Fatalf("want 1 retracted, 0 bumps; got %+v", r)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

// A run that DOES report a dark-section finding resets the streak, even at 2.
// Without this the counter is a lifetime tally rather than a consecutive one,
// and every ticket eventually closes regardless of what the audit says.
func TestAuditRetraction_AFindingResetsTheStreak(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()
	siteID := uuid.New()
	itemID := uuid.New()

	expectPagesLoad(mock, siteID)
	// The filing half runs as before: blocked keys, blocked check, dedup, insert.
	mock.ExpectQuery("status = 'blocked'").
		WillReturnRows(sqlmock.NewRows([]string{"item_key"}))
	mock.ExpectQuery("SELECT EXISTS").WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))
	mock.ExpectQuery("SELECT EXISTS").WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))
	mock.ExpectExec("INSERT INTO site_work_items").WillReturnResult(sqlmock.NewResult(0, 1))

	mock.ExpectBegin()
	mock.ExpectQuery("item_type = 'dark_section_audit'").
		WithArgs(siteID).
		WillReturnRows(candidateRows().AddRow(itemID, "k", "failed", ourSpec,
			`{"retraction":{"silent_runs":2,"audit_source":"visual-design-audit"}}`))
	// The reset REMOVES the key rather than writing a zero, so a row that never
	// went silent and one whose streak was broken look the same afterwards.
	mock.ExpectExec("UPDATE site_work_items").
		WithArgs(itemID).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	out, err := WriteAuditFindingsAction(context.Background(),
		auditParams(db, siteID, []interface{}{darkSectionFinding}))
	if err != nil {
		t.Fatalf("action failed: %v", err)
	}
	r := retractionFor(t, out, "dark_section_audit")
	if r.Silent || r.StreaksReset != 1 || r.Retracted != 0 {
		t.Fatalf("want not-silent, 1 reset, 0 retracted; got %+v", r)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

// ---------------------------------------------------------------------------
// THE AVAILABILITY CONTROL: a reply we did not understand is not silence.
// ---------------------------------------------------------------------------

// An unrecognised reply produces zero findings and no parse error — exactly
// what a clean site produces — and must retract NOTHING and touch NOTHING.
// Without this control, every audit whose response shape drifts would close
// that site's tickets three runs later.
//
// ⚠ THIS TEST IS WRITTEN AGAINST THE FUNCTION, NOT THROUGH THE ACTION, AND
// THAT IS THE WHOLE POINT. The obvious version — drive the action with an
// unrecognised reply, set no ExpectBegin, and assert no "retraction" key comes
// back — PASSES AGAINST A REMOVED GUARD, which is how it was first written and
// what the mutation run caught. With the guard gone the code DOES open a
// transaction, sqlmock refuses it as unexpected, the action logs the error and
// returns no retraction key, and the assertion is satisfied by the mock's
// refusal rather than by the code's. A mock's own bookkeeping cannot assert a
// negative. Calling the function directly makes the same refusal an ERROR
// RETURN, which the assertion below can actually see.
func TestAuditRetraction_UnrecognisedReplyIsNotSilence(t *testing.T) {
	for _, tc := range []struct {
		name string
		raw  interface{}
	}{
		{"object with no findings key", map[string]interface{}{"overall_score": float64(7)}},
		{"array of non-objects", []interface{}{"a", "b"}},
		{"scalar", float64(7)},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, _, recognised := parseAuditFindings(tc.raw, zap.NewNop())
			if recognised {
				t.Fatalf("fixture wrong: %#v must parse as UNRECOGNISED", tc.raw)
			}

			// A mock with NO expectations at all: any database call whatsoever
			// fails it. The guard's contract is that there is not one.
			db, _, err := sqlmock.New()
			if err != nil {
				t.Fatalf("sqlmock: %v", err)
			}
			defer db.Close()

			out, rErr := retractSilentAuditFindings(context.Background(), db, uuid.New(),
				"visual-design-audit", uuid.New(), map[string]bool{}, recognised, zap.NewNop())
			if rErr != nil {
				t.Fatalf("an unrecognised reply must touch no database at all, got: %v", rErr)
			}
			if out != nil {
				t.Fatalf("an unrecognised reply must report no retraction, got %#v", out)
			}
		})
	}
}

// The companion in the other direction, and the reason the guard cannot be
// "shapeRecognised is always false in this branch": a RECOGNISED empty reply
// reaching the very same early return DOES open the transaction and DOES
// advance the streak. Without this, deleting the retraction call from the
// zero-findings branch would go unnoticed — and that branch is where the
// silence case actually arrives.
func TestAuditRetraction_RecognisedEmptyReplyStillRetracts(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()
	siteID := uuid.New()
	itemID := uuid.New()

	expectPagesLoad(mock, siteID)
	mock.ExpectBegin()
	mock.ExpectQuery("item_type = 'dark_section_audit'").
		WithArgs(siteID).
		WillReturnRows(candidateRows().AddRow(itemID, "k", "failed", ourSpec, `{}`))
	mock.ExpectExec("UPDATE site_work_items").
		WithArgs(itemID, 1, "visual-design-audit").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	// The wrapped-object shape with an empty list — a real auditor reporting a
	// clean site, which reaches the len(findings)==0 early return.
	out, err := WriteAuditFindingsAction(context.Background(),
		auditParams(db, siteID, map[string]interface{}{"overall_score": float64(9), "findings": []interface{}{}}))
	if err != nil {
		t.Fatalf("action failed: %v", err)
	}
	m := out.(map[string]interface{})
	if m["shape_recognised"] != true {
		t.Fatalf("a wrapped empty findings list must be reported as recognised: %#v", m)
	}
	r := retractionFor(t, out, "dark_section_audit")
	if !r.Silent || r.StreaksBumped != 1 {
		t.Fatalf("want silent with 1 bump, got %+v", r)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

// ---------------------------------------------------------------------------
// THE INHERITED TRAP: silence is measured on what was OBSERVED, not FILED.
// ---------------------------------------------------------------------------

// A run that observes a dark-section defect and files NOTHING — because the
// dedup check found the row already open — is NOT silent. This is WII-016's
// landmine in this producer's shape: `items_created == 0` is the number that
// looks like silence and is not.
func TestAuditRetraction_ObservedButDedupedIsNotSilence(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()
	siteID := uuid.New()
	itemID := uuid.New()

	expectPagesLoad(mock, siteID)
	mock.ExpectQuery("status = 'blocked'").
		WillReturnRows(sqlmock.NewRows([]string{"item_key"}))
	mock.ExpectQuery("SELECT EXISTS").WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))
	// The dedup check says the row is already open → nothing is inserted.
	mock.ExpectQuery("SELECT EXISTS").WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

	mock.ExpectBegin()
	mock.ExpectQuery("item_type = 'dark_section_audit'").
		WithArgs(siteID).
		WillReturnRows(candidateRows().AddRow(itemID, "k", "failed", ourSpec,
			`{"retraction":{"silent_runs":2,"audit_source":"visual-design-audit"}}`))
	mock.ExpectExec("UPDATE site_work_items").WithArgs(itemID).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	out, err := WriteAuditFindingsAction(context.Background(),
		auditParams(db, siteID, []interface{}{darkSectionFinding}))
	if err != nil {
		t.Fatalf("action failed: %v", err)
	}
	m := out.(map[string]interface{})
	if m["items_created"] != 0 {
		t.Fatalf("fixture wrong: this run must file nothing, got %#v", m["items_created"])
	}
	r := retractionFor(t, out, "dark_section_audit")
	if r.Silent {
		t.Fatalf("a run that OBSERVED the defect and filed nothing is not silent: %+v", r)
	}
	if r.Retracted != 0 || r.StreaksReset != 1 {
		t.Fatalf("want 0 retracted, 1 reset; got %+v", r)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

// ---------------------------------------------------------------------------
// THE CO-FILING CONTROL: one producer's silence is not evidence about another.
// ---------------------------------------------------------------------------

// Rows filed by a DIFFERENT audit_source must be left entirely alone — no
// streak, no retraction — however silent this producer is. WII-016's guardian
// objection, answered structurally rather than by census.
func TestAuditRetraction_AnotherProducersRowsAreUntouched(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()
	siteID := uuid.New()
	ours, theirs := uuid.New(), uuid.New()

	expectPagesLoad(mock, siteID)
	mock.ExpectBegin()
	mock.ExpectQuery("item_type = 'dark_section_audit'").
		WithArgs(siteID).
		WillReturnRows(candidateRows().
			AddRow(theirs, "content-quality-audit_dark_section_audit_index_x", "failed",
				`{"audit_source":"content-quality-audit"}`,
				`{"retraction":{"silent_runs":2,"audit_source":"visual-design-audit"}}`).
			AddRow(ours, "visual-design-audit_dark_section_audit_index_x", "failed", ourSpec, `{}`))
	// EXACTLY ONE write, and it is ours. The other row carries a streak of 2
	// deliberately: if producer scope were dropped, it would retract HERE, and
	// this test would see two writes.
	mock.ExpectExec("UPDATE site_work_items").
		WithArgs(ours, 1, "visual-design-audit").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	out, err := WriteAuditFindingsAction(context.Background(), auditParams(db, siteID, []interface{}{}))
	if err != nil {
		t.Fatalf("action failed: %v", err)
	}
	r := retractionFor(t, out, "dark_section_audit")
	if r.Candidates != 1 || r.Retracted != 0 || r.StreaksBumped != 1 {
		t.Fatalf("want 1 candidate, 0 retracted, 1 bump; got %+v", r)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

// ---------------------------------------------------------------------------
// THE REAPER CONTROL: never write to a row the stale reaper is watching.
// ---------------------------------------------------------------------------

// `triaged` and `claimed` rows must be skipped entirely. The reason is not
// taste: `trg_site_work_items_updated_at` bumps updated_at on EVERY write to
// this table, and `stale-work-item-reaper` parks items whose updated_at is
// older than 48h while `triaged`. A streak write every 15 minutes would make
// that row unreapable for ever — migration 237's header names this hazard
// explicitly. The `deferred` row in the same fixture is the positive control:
// it proves the skip is keyed on the two in-flight statuses and is not simply
// skipping everything.
func TestAuditRetraction_InFlightRowsAreNeverWritten(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()
	siteID := uuid.New()
	triaged, claimed, deferred := uuid.New(), uuid.New(), uuid.New()

	expectPagesLoad(mock, siteID)
	mock.ExpectBegin()
	mock.ExpectQuery("item_type = 'dark_section_audit'").
		WithArgs(siteID).
		WillReturnRows(candidateRows().
			AddRow(triaged, "k1", "triaged", ourSpec,
				`{"retraction":{"silent_runs":2,"audit_source":"visual-design-audit"}}`).
			AddRow(claimed, "k2", "claimed", ourSpec,
				`{"retraction":{"silent_runs":2,"audit_source":"visual-design-audit"}}`).
			AddRow(deferred, "k3", "deferred", ourSpec,
				`{"retraction":{"silent_runs":2,"audit_source":"visual-design-audit"}}`))
	// ONLY the deferred row is touched, and it retracts (streak reaches 3).
	mock.ExpectExec("UPDATE site_work_items").
		WithArgs("visual-design-audit", sqlmock.AnyArg(), siteID, "dark_section_audit", "k3", sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	out, err := WriteAuditFindingsAction(context.Background(), auditParams(db, siteID, []interface{}{}))
	if err != nil {
		t.Fatalf("action failed: %v", err)
	}
	r := retractionFor(t, out, "dark_section_audit")
	if r.SkippedInFlight != 2 {
		t.Fatalf("want 2 in-flight rows skipped, got %+v", r)
	}
	if r.Retracted != 1 {
		t.Fatalf("want the deferred row retracted (the positive control), got %+v", r)
	}
	// A retraction that closes a PARKED row must say so — a park draining
	// silently is what that counter exists to prevent.
	if r.RetractedPark != 1 {
		t.Fatalf("want retracted_parked=1, got %+v", r)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

// ---------------------------------------------------------------------------
// THE OPT-IN CONTROL: a type not on the roster is untouched.
// ---------------------------------------------------------------------------

// The roster is the whole of the authority. A silent run whose site holds only
// a NON-gated type must load no candidates for it and write nothing — the
// unsafe default is OFF, per the 2026-08-02 shared-seam ruling. Asserted by
// the absence of any query naming that type, which sqlmock enforces because an
// unexpected query fails the call.
func TestAuditRetraction_UngatedItemTypeIsInert(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()
	siteID := uuid.New()

	expectPagesLoad(mock, siteID)
	mock.ExpectBegin()
	// dark_section_audit is the ONLY gated type, so this is the only load.
	// cta_improvement rows exist on this site and are never asked about.
	mock.ExpectQuery("item_type = 'dark_section_audit'").
		WithArgs(siteID).
		WillReturnRows(candidateRows())
	mock.ExpectCommit()

	out, err := WriteAuditFindingsAction(context.Background(), auditParams(db, siteID, []interface{}{}))
	if err != nil {
		t.Fatalf("action failed: %v", err)
	}
	m := out.(map[string]interface{})
	rr := m["retraction"].(map[string]interface{})
	if len(rr) != 1 {
		t.Fatalf("exactly one gated type expected in the report, got %#v", rr)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

// ---------------------------------------------------------------------------
// helpers
// ---------------------------------------------------------------------------

func retractionFor(t *testing.T, out interface{}, itemType string) silenceRetractionResult {
	t.Helper()
	m, ok := out.(map[string]interface{})
	if !ok {
		t.Fatalf("action returned %T, want map", out)
	}
	raw, present := m["retraction"]
	if !present {
		t.Fatalf("no retraction reported: %#v", m)
	}
	per, ok := raw.(map[string]interface{})
	if !ok {
		t.Fatalf("retraction is %T, want map", raw)
	}
	r, ok := per[itemType].(silenceRetractionResult)
	if !ok {
		t.Fatalf("no %s entry in retraction report: %#v", itemType, per)
	}
	return r
}

// ---------------------------------------------------------------------------
// THE SHARED-PATH CONTROL (council 54e3b698, `guardian`, medium — actioned)
// ---------------------------------------------------------------------------

// Hoisting classification above the blocked/dedup/insert filters changed the
// loop EVERY producer runs through, not just this one's. Six live agents call
// write_audit_findings; the retraction work above only pins dark_section_audit,
// so on its own it would leave the other five unevidenced — which is exactly
// what the guardian seat objected to.
//
// This drives a NON-gated item_type (cta -> cta_improvement, absent from
// silenceRetractionGates) through the full filter chain and pins that the order
// and the effects are unchanged: blocked-key load, per-finding blocked EXISTS,
// dedup EXISTS, then the INSERT carrying that finding's own item_type, handler
// and dedup key. sqlmock is ordered, so a reordering of these four fails here.
func TestWriteAuditFindings_UngatedProducerPathIsUnchanged(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()
	siteID := uuid.New()
	pageID := uuid.New()

	mock.ExpectQuery("FROM pages").
		WithArgs(siteID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "page_type", "sections"}).
			AddRow(pageID, "index", "content", "[]"))
	mock.ExpectQuery("status = 'blocked'").
		WillReturnRows(sqlmock.NewRows([]string{"item_key"}))
	// The broader blocked check must carry the PRODUCER scope (bugs_open/279,
	// the bugfix_213 lane's contribution): without spec->>'audit_source' = $4,
	// a blocked capability_gap filed by the discovery path mutes every unrouted
	// audit category site-wide (18 such rows on 14 sites when found). Pinning
	// the regex AND the args proves the clause and the value are both in the
	// query this action actually runs.
	mock.ExpectQuery(`SELECT EXISTS[\s\S]*spec->>'audit_source'`).
		WithArgs(siteID, "cta_improvement", pageID, "visual-design-audit").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))
	mock.ExpectQuery("SELECT EXISTS").
		WithArgs(siteID, "visual-design-audit_cta_improvement_index_"+siteID.String()).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))
	// `status` became a parameter (position 10) when the bugs_open/279 fix let
	// the capability_gap fallback file as 'deferred'; a routed finding still
	// inserts 'detected', and this control pins that.
	mock.ExpectExec("INSERT INTO site_work_items").
		WithArgs(siteID, "discovery", "cta_improvement", "medium", sqlmock.AnyArg(),
			argJSONContains{`"audit_source":"visual-design-audit"`}, pageID, sqlmock.AnyArg(),
			"component-template-fixer", "detected", "visual-design-audit",
			"visual-design-audit_cta_improvement_index_"+siteID.String(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 1))

	// The retraction pass still runs — and asks about dark_section_audit ONLY.
	// An ungated type must never reach the candidate loader; sqlmock fails an
	// unexpected query, so a second load here would fail this test.
	mock.ExpectBegin()
	mock.ExpectQuery("item_type = 'dark_section_audit'").
		WithArgs(siteID).
		WillReturnRows(candidateRows())
	mock.ExpectCommit()

	out, err := WriteAuditFindingsAction(context.Background(),
		auditParams(db, siteID, []interface{}{map[string]interface{}{
			"category":    "cta",
			"severity":    "medium",
			"description": "the hero call to action is below the fold",
			"page":        "index",
		}}))
	if err != nil {
		t.Fatalf("action failed: %v", err)
	}
	m := out.(map[string]interface{})
	if m["items_created"] != 1 {
		t.Fatalf("the ungated producer must still file its item: %#v", m)
	}
	stats, _ := m["classification_stats"].(map[string]int)
	if stats["cta_improvement"] != 1 {
		t.Fatalf("classification_stats lost the ungated type: %#v", m["classification_stats"])
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

// The companion: the dedup filter must still SUPPRESS an ungated finding, and
// suppression must still be counted. Without this, the test above would pass
// against a reordering that ran the insert before the filters.
func TestWriteAuditFindings_UngatedProducerDedupStillSuppresses(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()
	siteID := uuid.New()

	expectPagesLoad(mock, siteID)
	mock.ExpectQuery("status = 'blocked'").
		WillReturnRows(sqlmock.NewRows([]string{"item_key"}))
	mock.ExpectQuery("SELECT EXISTS").WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))
	mock.ExpectQuery("SELECT EXISTS").WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))
	// No INSERT: an unexpected Exec would fail this test.
	mock.ExpectBegin()
	mock.ExpectQuery("item_type = 'dark_section_audit'").
		WithArgs(siteID).
		WillReturnRows(candidateRows())
	mock.ExpectCommit()

	out, err := WriteAuditFindingsAction(context.Background(),
		auditParams(db, siteID, []interface{}{map[string]interface{}{
			"category": "cta", "severity": "medium", "description": "d", "page": "index",
		}}))
	if err != nil {
		t.Fatalf("action failed: %v", err)
	}
	m := out.(map[string]interface{})
	if m["items_created"] != 0 || m["items_skipped"] != 1 {
		t.Fatalf("want created=0 skipped=1, got %#v", m)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}
