package actions

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/gqls/agentchassis/pkg/models"
	checks "github.com/gqls/agentchassis/platform/orchestration/actions/discovery_checks"
	orchtypes "github.com/gqls/agentchassis/platform/orchestration/types"
)

// ── bugs_open/375: the SECOND writer of `complete` now has the same gate ──────
//
// These assert at the ACTION, not at the helper, because the defect was never in
// the gate — it was in the wiring. A helper that verifies correctly and is called
// by nobody is exactly the state this bug describes, and a test of the helper
// alone would have passed throughout.
//
// ⚠ THE MUTATION RULE (LANDMINES.md, "a mock's own bookkeeping cannot assert a
// NEGATIVE"). Every guarantee below was proved by NEUTERING the guard and
// requiring the test to fail — see NOTES_completion_verifier_gap.md for what was
// mutated and which tests fell over. A test asserting "the verifier was not
// called" passes just as happily when the whole arm is unreachable.
//
// ⚠ AND THE SIBLING TRAP ("a mutation that PASSES may have hit a guard in
// SERIES"): this arm ALREADY carries the terminal-decision guard
// (workItemCompletionGuardStatuses). Under sqlmock the UPDATE is mocked, so that
// guard cannot mask anything here — but a fixture run against a real row must sit
// in a status the terminal guard lets through (detected/claimed/triaged) or the
// mutation is vacuous.

const (
	testVerifiedItemType   = "test_375_verified_type"
	testUnverifiedItemType = "test_375_unverified_type"
)

// verifierCall records whether the registered verifier's predicate actually ran.
// It is SUPPORTING evidence only — the load-bearing assertions below are on the
// action's observable output (the item's fate and the recorded _verification),
// because that is what a human and a census read.
var verifierCall struct {
	ran      bool
	resolved bool
	target   checks.VerifyTarget
}

// installTestVerifier points the gate's lookup at a fixture for the duration of
// ONE test.
//
// ⚠ It deliberately does NOT call checks.RegisterVerifier. The registry has no
// removal, and it is read process-wide by
// TestClaimTimeoutExclusionCoversBothCompletionGates, which requires every
// verified item_type to be declared in livespec.ClaimedItemTimeoutExclusions —
// because a THIRD writer, the claimed-item-timeout sweep, completes rows directly
// with neither gate running (bugs_closed/317). The first draft of this file
// registered a synthetic type in init() and broke that guard. The guard was right;
// the fixture was wrong. See NOTES_completion_verifier_gap.md, misstep 3.
func installTestVerifier(t *testing.T, itemType string) {
	t.Helper()
	verifierCall.ran, verifierCall.resolved = false, false
	verifierCall.target = checks.VerifyTarget{}
	prev := verifierLookup
	verifierLookup = func(lookedUp string) (checks.ItemVerifier, checks.VerifierPolicy) {
		if lookedUp != itemType {
			return prev(lookedUp)
		}
		return func(ctx context.Context, db *sql.DB, tgt checks.VerifyTarget, l *zap.Logger) (checks.VerifyResult, error) {
			verifierCall.ran = true
			verifierCall.target = tgt
			return checks.VerifyResult{Resolved: verifierCall.resolved, Detail: "test predicate"}, nil
		}, checks.VerifierPolicy{}
	}
	t.Cleanup(func() { verifierLookup = prev })
}

// The seam must stay a seam. If production code ever re-points verifierLookup,
// every assertion in this file is about a fixture rather than about the platform.
func TestVerifierLookupIsNotASwitchInProduction(t *testing.T) {
	files, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatal(err)
	}
	needle := "verifierLookup" + " ="
	found := 0
	for _, f := range files {
		if strings.HasSuffix(f, "_test.go") {
			continue
		}
		src, err := os.ReadFile(f)
		if err != nil {
			t.Fatal(err)
		}
		found += strings.Count(string(src), needle)
	}
	// Exactly one: `var verifierLookup = checks.GetVerifier`. The comment above it
	// is written to avoid spelling the assignment, so this cannot pass vacuously on
	// prose describing itself.
	if found != 1 {
		t.Errorf("found %d assignments to the lookup seam in non-test source, want exactly 1 (the declaration).\n"+
			"Production must always use checks.GetVerifier; re-pointing it turns every test in this file into a\n"+
			"test of its own fixture.", found)
	}
}

// verifyRowReadSQL is the column list loadWorkItemVerifyRow must actually SELECT,
// as a regex for sqlmock's query matcher.
//
// ⚠ IT IS THE EXPECTATION, NOT THE RETURNED ROWS, THAT GUARDS THIS. sqlmock hands
// back whatever rows the test queued no matter what the statement says, so a test
// asserting only on the VALUES proves the plumbing and nothing about the query —
// exactly the "a mock's own bookkeeping cannot assert a NEGATIVE" trap, one level
// along. Measured: with a values-only assertion, replacing the spec column with a
// literal '{}' failed no test in this package. Matching the text is what closes it.
const verifyRowReadSQL = `SELECT item_type, COALESCE\(spec, '\{\}'::jsonb\), site_id, page_id`

// expectVerifyRowRead queues the single row read both arms share.
func expectVerifyRowRead(mock sqlmock.Sqlmock, itemType string) {
	mock.ExpectQuery(verifyRowReadSQL).
		WillReturnRows(sqlmock.NewRows([]string{"item_type", "spec", "site_id", "page_id"}).
			AddRow(itemType, []byte(`{}`), uuid.New(), nil))
}

func updateStatusParams(t *testing.T, db *sql.DB, itemID uuid.UUID, cfg map[string]interface{}) ActionParams {
	t.Helper()
	cfg["status"] = "complete"
	return ActionParams{
		StepConfig:    models.Step{Config: cfg},
		CollectedData: map[string]interface{}{"input_data": map[string]interface{}{"work_item_id": itemID.String()}},
		DB:            db,
		Logger:        zap.NewNop(),
		ExecutionContext: &orchtypes.ExecutionContext{
			StepName: "close_complete",
			Sender:   orchtypes.AgentIdentity{AgentType: "test-handler"},
		},
	}
}

// THE HEADLINE. A type with NO registered verifier is untouched by this change:
// no row read, no gate, and the completion UPDATE runs exactly as before. This is
// every live item of all six configured `complete` arms as of 2026-08-24, so it is
// the assertion that says "inert today".
func TestUpdateWorkItemStatus_NoVerifierRegistered_IsInertOnBothArms(t *testing.T) {
	for _, armed := range []bool{false, true} {
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("sqlmock: %v", err)
		}
		itemID := uuid.New()
		expectVerifyRowRead(mock, testUnverifiedItemType)
		mock.ExpectExec("UPDATE site_work_items").WillReturnResult(sqlmock.NewResult(0, 1))

		cfg := map[string]interface{}{}
		if armed {
			cfg[updateStatusVerifyConfigKey] = true
		}
		out, err := UpdateWorkItemStatusAction(context.Background(), updateStatusParams(t, db, itemID, cfg))
		if err != nil {
			t.Fatalf("armed=%v: %v", armed, err)
		}
		m := out.(map[string]interface{})
		if m["updated"] != true || m["status"] != "complete" {
			t.Errorf("armed=%v: an unverified type must still complete, got %v", armed, m)
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("armed=%v: %v", armed, err)
		}
		db.Close()
	}
}

// THE TRAP, MADE AUDIBLE. A verifier IS registered for the type, and the step did
// not arm the consult. The completion still happens — no liveness change — but the
// row now SAYS a guard was skipped, instead of the bypass passing in silence.
//
// This is the assertion that protects the person verifier_coverage_test.go's
// backlog invites to write one of these verifiers.
func TestUpdateWorkItemStatus_UnarmedWithVerifier_CompletesButRecordsTheBypass(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()
	itemID := uuid.New()
	installTestVerifier(t, testVerifiedItemType)

	expectVerifyRowRead(mock, testVerifiedItemType)
	// The recorded bypass must reach the ROW, not only the pod log — a log does
	// not survive a roll. Asserted on the UPDATE's own result argument.
	mock.ExpectExec("UPDATE site_work_items").
		WithArgs(itemID, "complete", sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 1))

	out, err := UpdateWorkItemStatusAction(context.Background(), updateStatusParams(t, db, itemID, map[string]interface{}{}))
	if err != nil {
		t.Fatalf("UpdateWorkItemStatusAction: %v", err)
	}
	if m := out.(map[string]interface{}); m["updated"] != true {
		t.Fatalf("an unarmed step must still complete — this change carries no liveness risk: %v", m)
	}
	if verifierCall.ran {
		t.Error("the verifier must NOT run unarmed: arming is a per-step decision (CQ-023)")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

// The recorded payload is checked at the gate, where its exact shape is readable.
// Split from the action test above on purpose: the action marshals it into an
// opaque JSON argument, and asserting a substring of that would be asserting on
// the mock's rendering rather than on the claim.
func TestVerifyBeforeUpdateStatusComplete_UnarmedRecordsWhatWasSkipped(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()
	installTestVerifier(t, testVerifiedItemType)
	expectVerifyRowRead(mock, testVerifiedItemType)

	payload, mayComplete := verifyBeforeUpdateStatusComplete(context.Background(), db, uuid.New(), false, zap.NewNop())
	if !mayComplete {
		t.Fatal("unarmed must never block")
	}
	if payload == nil {
		t.Fatal("a skipped guard must be RECORDED — a silent bypass is the whole bug")
	}
	if payload["status"] != "verifier_not_consulted" {
		t.Errorf("status = %v, want verifier_not_consulted", payload["status"])
	}
	if payload["item_type"] != testVerifiedItemType {
		t.Errorf("the record must name the type whose guard was skipped, got %v", payload["item_type"])
	}
	// The remedy has to name the key, or the record tells a reader what happened
	// and not what to do about it.
	if remedy, _ := payload["remedy"].(string); remedy == "" ||
		!strings.Contains(remedy, updateStatusVerifyConfigKey) {
		t.Errorf("remedy must name %q, got %q", updateStatusVerifyConfigKey, remedy)
	}
}

// ARMED AND THE DEFECT PERSISTS → the completion is REFUSED, and refused through
// the guarded writer's own attempt machinery rather than a second path invented
// here: attempt_count+1, claim released, 'triaged' or 'failed'.
func TestUpdateWorkItemStatus_ArmedAndDefectPersists_RefusesTheCompletion(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()
	itemID := uuid.New()
	installTestVerifier(t, testVerifiedItemType)

	expectVerifyRowRead(mock, testVerifiedItemType)
	// failUnverifiedCompletion's UPDATE — note it RETURNS the new status, which is
	// what distinguishes it from the plain completion UPDATE the mock would
	// otherwise match.
	mock.ExpectQuery("UPDATE site_work_items").
		WillReturnRows(sqlmock.NewRows([]string{"status"}).AddRow("triaged"))

	out, err := UpdateWorkItemStatusAction(context.Background(),
		updateStatusParams(t, db, itemID, map[string]interface{}{updateStatusVerifyConfigKey: true}))
	if err != nil {
		t.Fatalf("UpdateWorkItemStatusAction: %v", err)
	}
	m := out.(map[string]interface{})
	if m["completed"] != false || m["verified"] != false {
		t.Fatalf("an unverified completion must be refused, got %v", m)
	}
	if m["new_status"] != "triaged" || m["will_retry"] != true {
		t.Errorf("a refusal must route into the retry ladder, got %v", m)
	}
	if m["reason"] != "verification_failed" {
		t.Errorf("reason = %v, want verification_failed", m["reason"])
	}
	if !verifierCall.ran {
		t.Error("the armed arm must actually RUN the predicate")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

// THE NEGATIVE CONTROL, and it is not optional: a guard that refuses everything
// passes the test above. Same path, same arming, predicate satisfied → the item
// must complete.
func TestUpdateWorkItemStatus_ArmedAndResolved_StillCompletes(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()
	itemID := uuid.New()
	installTestVerifier(t, testVerifiedItemType)
	verifierCall.resolved = true

	expectVerifyRowRead(mock, testVerifiedItemType)
	mock.ExpectExec("UPDATE site_work_items").WillReturnResult(sqlmock.NewResult(0, 1))

	out, err := UpdateWorkItemStatusAction(context.Background(),
		updateStatusParams(t, db, itemID, map[string]interface{}{updateStatusVerifyConfigKey: true}))
	if err != nil {
		t.Fatalf("UpdateWorkItemStatusAction: %v", err)
	}
	if m := out.(map[string]interface{}); m["updated"] != true || m["status"] != "complete" {
		t.Fatalf("a satisfied predicate must still complete, got %v", m)
	}
	if !verifierCall.ran {
		t.Error("the predicate must have run")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

// A non-complete status must not touch the gate at all — no row read, no verifier.
// This is what keeps `failed` on the failure ladder and the six needs_human_review
// steps exactly where they were.
func TestUpdateWorkItemStatus_NonCompleteStatusesSkipTheGateEntirely(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()
	itemID := uuid.New()
	installTestVerifier(t, testVerifiedItemType)

	// No ExpectQuery for the row read: if the gate ran, sqlmock fails the call.
	mock.ExpectExec("UPDATE site_work_items").WillReturnResult(sqlmock.NewResult(0, 1))

	params := updateStatusParams(t, db, itemID, map[string]interface{}{updateStatusVerifyConfigKey: true})
	params.StepConfig.Config["status"] = "needs_human_review"
	if _, err := UpdateWorkItemStatusAction(context.Background(), params); err != nil {
		t.Fatalf("UpdateWorkItemStatusAction: %v", err)
	}
	if verifierCall.ran {
		t.Error("a non-complete status has no defect claim to verify")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

// The item's SPEC must actually reach the verifier — the column the shared row read
// exists to fetch.
//
// ⚠ THIS TEST EXISTS BECAUSE A MUTATION FOUND NOTHING. The council's guardian seat
// objected (corr 7a6add95, medium) that factoring loadWorkItemVerifyRow out of
// verifyBeforeComplete touches a helper the widest set of live agents depends on, and
// asked for mutation proof that the existing callers are unaffected. Two mutations
// were run. Emptying ItemType failed six tests, three of them pre-existing
// TestVerifyBeforeComplete_* ones — the extraction IS guarded. But replacing the spec
// column with a literal '{}' failed NOTHING: no test in this package had ever asserted
// that a verifier receives the item's real spec, so that column could be dropped
// silently by any future edit. The gap predates this change (it was the same when the
// SELECT was inline) and is closed here because the objection is what surfaced it.
func TestVerifyBeforeUpdateStatusComplete_ArmedPassesTheItemsSpecToTheVerifier(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()
	installTestVerifier(t, testVerifiedItemType)
	verifierCall.resolved = true

	itemID, siteID, pageID := uuid.New(), uuid.New(), uuid.New()
	mock.ExpectQuery(verifyRowReadSQL).
		WillReturnRows(sqlmock.NewRows([]string{"item_type", "spec", "site_id", "page_id"}).
			AddRow(testVerifiedItemType, []byte(`{"component_id":"abc","reason":"empty"}`), siteID, pageID))

	if _, mayComplete := verifyBeforeUpdateStatusComplete(context.Background(), db, itemID, true, zap.NewNop()); !mayComplete {
		t.Fatal("resolved predicate must permit completion")
	}
	if !verifierCall.ran {
		t.Fatal("the predicate must have run")
	}
	if got := verifierCall.target.Spec["component_id"]; got != "abc" {
		t.Errorf("the verifier was handed spec %v — it must receive the ITEM's spec, or it grades a defect it cannot locate", verifierCall.target.Spec)
	}
	// The scoping ids travel with it: bugs_open/213's Grades predicate reads these to
	// decide whether it speaks for this item at all.
	if verifierCall.target.SiteID != siteID {
		t.Errorf("SiteID = %v, want %v", verifierCall.target.SiteID, siteID)
	}
	if verifierCall.target.PageID == nil || *verifierCall.target.PageID != pageID {
		t.Errorf("PageID = %v, want %v", verifierCall.target.PageID, pageID)
	}
	if verifierCall.target.ItemID != itemID {
		t.Errorf("ItemID = %v, want %v", verifierCall.target.ItemID, itemID)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
