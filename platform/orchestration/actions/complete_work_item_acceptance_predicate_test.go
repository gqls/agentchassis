// FILE: platform/orchestration/actions/complete_work_item_acceptance_predicate_test.go
//
// Gate 1c's tests (bugs_open/395).
//
// THE ONE THAT MATTERS MOST is TestTheLiveStoredPredicateIsUnevaluableUntilStripped,
// because it pins a trap that would have made this entire gate silently inert while
// looking like it worked: the shape stored in site_work_items.spec carries two
// emission-provenance keys, and EvaluateAcceptancePredicate enforces a closed key
// set, so evaluating the stored form verbatim returns `inapplicable` for EVERY live
// predicate. The only test in the estate over real live predicates
// (TestTheFirstLiveEmittedPredicatesStillRefuteAfterTheFix) hand-writes them WITHOUT
// those keys, so it exercises a shape the database does not contain and could never
// have caught this.
//
// THE CONTROLS, and this file is not honest without them. A gate that refused
// everything, or one that graded nothing at all, would pass a refutation test on its
// own. Three tests discriminate:
//
//   - TestAPredicateThatHoldsPermitsTheCompletion — the negative control, at the
//     gate. ⚠ It is a UNIT and NOT a live row: bugs_open/395 §6 asks for an item
//     whose predicate is satisfied after its fix, and no such row exists
//     [MEASURED 2026-08-25]. The two must not be quoted for one another.
//   - TestAnUnoptedTypeIsNeverGraded — the opt-in guarantee: a type absent from the
//     roster must not even reach the spec decode.
//   - TestAnItemWithNoPredicateIsNotGraded — the common case; most items of an armed
//     type carry no predicate at all.
package actions

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// liveStoredPredicateJSON is the `acceptance_predicate` object EXACTLY as it sits in
// site_work_items.spec for the worked case of bugs_open/395 — webdesign.co.uk's
// index finding, item b4c82ec3, written 2026-08-24 22:08:38Z and read back from the
// live database on 2026-08-25. Provenance keys included, because their presence is
// the whole point of this fixture.
const liveStoredPredicateJSON = `{
  "page": "index",
  "type": "text_order",
  "after": ["$cardinal"],
  "field": "meta_description",
  "before": ["paired", "pairing", "article", "guide"],
  "verdict_at_emission": "refutes",
  "evidence_at_emission": "meta_description of \"index\" states none of \"paired\", \"pairing\", \"article\", \"guide\", so nothing can precede \"$cardinal\""
}`

// liveIndexMeta is the served <meta name="description"> of webdesign.co.uk's index
// page, verbatim, AFTER the handler rebuilt and deployed it twice.
const liveIndexMeta = "Sixty-three browser tools for web design and development. No account, no upload, everything runs in your browser."

// satisfyingIndexMeta is liveIndexMeta rewritten so the item's own criterion is MET:
// a "before" needle ("paired") precedes the cardinal. It is the string a correct
// repair would have produced, and it is the negative control's input.
const satisfyingIndexMeta = "Browser tools for web design, each paired with a plain-English article explaining the concept behind it. Sixty-three of them, and no account is needed."

func liveStoredPredicate(t *testing.T) map[string]interface{} {
	t.Helper()
	var pred map[string]interface{}
	if err := json.Unmarshal([]byte(liveStoredPredicateJSON), &pred); err != nil {
		t.Fatalf("the live fixture does not decode: %v", err)
	}
	return pred
}

// ---------------------------------------------------------------------------
// The trap
// ---------------------------------------------------------------------------

// TestTheLiveStoredPredicateIsUnevaluableUntilStripped is the disconfirming half and
// the confirming half in one test, which is what makes it evidence rather than a
// restatement: the SAME predicate, over the SAME string, returns `inapplicable`
// stored and `refutes` stripped. Either assertion alone could pass for the wrong
// reason.
func TestTheLiveStoredPredicateIsUnevaluableUntilStripped(t *testing.T) {
	stored := liveStoredPredicate(t)
	subject := predSubject("index", "", liveIndexMeta)

	verdict, reason := EvaluateAcceptancePredicate(stored, subject)
	if verdict != PredicateInapplicable {
		t.Fatalf("the STORED shape evaluated to %s (%s).\n"+
			"This test exists because it must NOT: if the closed key set has been widened to admit "+
			"verdict_at_emission, a model can now write its own emission verdict and have the gate accept it "+
			"(bugs_closed/335's self-attribution failure). Check acceptancePredicateFields.", verdict, reason)
	}
	if !strings.Contains(reason, "verdict_at_emission") && !strings.Contains(reason, "evidence_at_emission") {
		t.Errorf("the rejection reason %q does not name the offending key — a consumer hitting this in "+
			"production would read it as a fault in the model's output", reason)
	}

	verdict, reason = EvaluateAcceptancePredicate(predicateForEvaluation(stored), subject)
	if verdict != PredicateRefutes {
		t.Fatalf("the STRIPPED live predicate evaluated to %s (%s), want refutes.\n"+
			"This is bugs_open/395's worked case: the page was rebuilt and deployed twice and its own "+
			"criterion is still unmet.", verdict, reason)
	}
}

// TestStampAndStripAreInverses is what stops a THIRD provenance key silently
// re-breaking every consumer. It does not enumerate the keys — it round-trips
// through both functions, so a key added to storedPredicate and forgotten in
// emissionProvenanceKeys fails here rather than in production.
func TestStampAndStripAreInverses(t *testing.T) {
	author := map[string]interface{}{
		"type": "text_order", "page": "index", "field": "meta_description",
		"before": []interface{}{"paired"}, "after": []interface{}{cardinalNeedle},
	}
	round := predicateForEvaluation(storedPredicate(author, "because the page says otherwise"))
	if !reflect.DeepEqual(author, round) {
		t.Fatalf("storedPredicate → predicateForEvaluation is not the identity.\ngot  %#v\nwant %#v\n"+
			"A provenance key was added to the stamp and not to emissionProvenanceKeys; every live "+
			"predicate now evaluates to `inapplicable` and gate 1c reads as permanently blind.", round, author)
	}
	// And the stamp must actually stamp — otherwise the round trip above is the
	// identity for the boring reason and this test proves nothing.
	stamped := storedPredicate(author, "why")
	if len(stamped) != len(author)+len(emissionProvenanceKeys) {
		t.Fatalf("storedPredicate added %d keys, want %d — this test's round trip would pass vacuously",
			len(stamped)-len(author), len(emissionProvenanceKeys))
	}
	if _, ok := author[emissionVerdictKey]; ok {
		t.Fatal("storedPredicate mutated its caller's map; the row's own decoded spec must not be rewritten")
	}
}

// ---------------------------------------------------------------------------
// The gate, end to end through verifyBeforeComplete
// ---------------------------------------------------------------------------

// specWithPredicate builds the item spec shape the gate reads.
func specWithPredicate(t *testing.T, pred map[string]interface{}) []byte {
	t.Helper()
	spec := map[string]interface{}{
		"page_name":       "index",
		"audit_source":    "offer-analysis",
		"acceptance_test": "The index page meta description mentions both the tool-article pairing and the no-account promise, in that order, before any catalogue count.",
	}
	if pred != nil {
		spec[acceptancePredicateKey] = pred
	}
	b, err := json.Marshal(spec)
	if err != nil {
		t.Fatal(err)
	}
	return b
}

// gateRun drives verifyBeforeComplete with a real *sql.DB (sqlmock) for both queries
// it makes: the item row, then gate 1c's page surface.
//
// ⚠ THE PAGE QUERY IS MATCHED ON ITS TEXT, not merely queued. sqlmock returns
// whatever rows a test queued regardless of the statement, so a test asserting only
// on values proves the plumbing and nothing about which query ran — the trap
// verifyRowReadSQL's comment records one file along.
func gateRun(t *testing.T, itemType string, spec []byte, meta string, expectPages bool) (map[string]interface{}, bool, *acceptancePredicateNote) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	mock.ExpectQuery(verifyRowReadSQL).
		WillReturnRows(sqlmock.NewRows([]string{"item_type", "spec", "site_id", "page_id"}).
			AddRow(itemType, spec, uuid.New(), nil))
	if expectPages {
		mock.ExpectQuery("FROM pages").WillReturnRows(
			sqlmock.NewRows([]string{"name", "title", "meta_description"}).
				AddRow("index", "Web Design Tools", meta))
	}

	payload, mayComplete, _, note := verifyBeforeComplete(
		context.Background(), db, uuid.New(), map[string]interface{}{"response": map[string]interface{}{"status": "complete"}}, zap.NewNop())

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("queries did not run as expected: %v", err)
	}
	return payload, mayComplete, note
}

// predicatePayload digs gate 1c's verdict out of the _verification payload.
func predicatePayload(t *testing.T, payload map[string]interface{}) map[string]interface{} {
	t.Helper()
	if payload == nil {
		t.Fatal("_verification payload is nil — gate 1c recorded nothing, so a false green is as invisible as before")
	}
	pred, ok := payload[acceptancePredicateKey].(map[string]interface{})
	if !ok {
		t.Fatalf("_verification carries no %q key: %#v", acceptancePredicateKey, payload)
	}
	return pred
}

// TestTheLiveFalseGreenIsRecordedOnTheRow is bugs_open/395's worked case driven
// through the real completion path. The item STILL COMPLETES — content_rewrite
// records rather than refuses — and that is the point being asserted: what changed
// is that the false green is now countable.
func TestTheLiveFalseGreenIsRecordedOnTheRow(t *testing.T) {
	payload, mayComplete, note := gateRun(t, "content_rewrite",
		specWithPredicate(t, liveStoredPredicate(t)), liveIndexMeta, true)

	if !mayComplete {
		t.Fatal("mayComplete = false: content_rewrite is declared predicateRecords, so this completion must " +
			"proceed. A blocked one here means the roster was promoted without the negative control " +
			"bugs_open/395 §6 requires, or without the claim-timeout exclusion.")
	}
	if note != nil {
		t.Fatalf("a blind-spot note was returned for a predicate that evaluated fine: %+v", note)
	}
	pred := predicatePayload(t, payload)
	if pred["verdict"] != string(PredicateRefutes) {
		t.Fatalf("verdict = %v, want refutes — this is the case the whole bug is about", pred["verdict"])
	}
	if pred["outcome"] != "recorded_only" {
		t.Errorf("outcome = %v, want recorded_only", pred["outcome"])
	}
	if pred["verdict_at_emission"] != string(PredicateRefutes) {
		t.Errorf("verdict_at_emission = %v; the emission verdict is echoed so a reader can see the "+
			"refutation was already there, without a second query", pred["verdict_at_emission"])
	}
	if s, _ := pred["promotion_owes"].(string); !strings.Contains(s, "NEGATIVE CONTROL") {
		t.Error("promotion_owes does not reach the row; a record-only arm with no stated debt is how it becomes permanent")
	}
}

// TestAPredicateThatHoldsPermitsTheCompletion is the NEGATIVE CONTROL, and without
// it every other test here is satisfied by a gate that refuses everything.
//
// ⚠ IT IS A UNIT, NOT A LIVE ROW. bugs_open/395 §6 asks for an item whose predicate
// is SATISFIED after its fix and which closed green through this gate; [MEASURED
// 2026-08-25] all three live predicates refute and no such row exists anywhere. This
// proves the permit arm computes; it does not prove the gate has ever permitted in
// production, and the two must never be quoted for one another.
func TestAPredicateThatHoldsPermitsTheCompletion(t *testing.T) {
	payload, mayComplete, note := gateRun(t, "content_rewrite",
		specWithPredicate(t, liveStoredPredicate(t)), satisfyingIndexMeta, true)

	if !mayComplete {
		t.Fatal("mayComplete = false on a SATISFIED predicate — the gate refuses everything")
	}
	if note != nil {
		t.Fatalf("a satisfied predicate produced a blind-spot note: %+v", note)
	}
	pred := predicatePayload(t, payload)
	if pred["verdict"] != string(PredicateHolds) {
		t.Fatalf("verdict = %v (%v), want holds.\nThe control string was written to satisfy the LIVE "+
			"predicate — \"paired\" before the cardinal — so a refutation here means the evaluator's "+
			"ordering arm moved, not that the fixture is wrong.", pred["verdict"], pred["detail"])
	}
	if pred["outcome"] != "permitted" {
		t.Errorf("outcome = %v, want permitted", pred["outcome"])
	}
	// The permit MUST be recorded. A gate that permits silently is indistinguishable
	// from one that never ran, which is the residual this whole roster entry carries.
	if _, ok := payload[acceptancePredicateKey]; !ok {
		t.Error("a permitted completion recorded nothing — the one arm that proves this gate is not simply refusing")
	}
}

// TestARefusingTypeBlocksTheCompletion drives the arm the live roster does not use,
// through the fixture seam. It is the only proof that a refusal reaches the caller.
func TestARefusingTypeBlocksTheCompletion(t *testing.T) {
	restore := acceptancePredicateGateFor
	acceptancePredicateGateFor = func(itemType string) (acceptancePredicateRule, bool) {
		if itemType != "synthetic_refusing_type" {
			return restore(itemType)
		}
		return acceptancePredicateRule{
			Why:        "fixture",
			OnRefuted:  predicateRefuses,
			RefusalWhy: "fixture",
		}, true
	}
	defer func() { acceptancePredicateGateFor = restore }()

	payload, mayComplete, note := gateRun(t, "synthetic_refusing_type",
		specWithPredicate(t, liveStoredPredicate(t)), liveIndexMeta, true)

	if mayComplete {
		t.Fatal("mayComplete = true on a still-refuting predicate for a type declaring predicateRefuses — " +
			"the completion this arm exists to stop")
	}
	if note != nil {
		t.Fatalf("a REFUSAL must not also be recorded as a blind spot — the blocked row is the record: %+v", note)
	}
	if payload["status"] != "acceptance_predicate_refuted" {
		t.Fatalf("status = %v, want acceptance_predicate_refuted — blockedCompletionReason switches on it", payload["status"])
	}
	msg, reason := blockedCompletionReason(payload)
	if reason != "acceptance_predicate_refuted" {
		t.Errorf("reason code = %q; an operator filtering the error column cannot tell this cause from a "+
			"handler that failed", reason)
	}
	if !strings.Contains(msg, "own stated acceptance criterion") || !strings.Contains(msg, "index") {
		t.Errorf("blocked message %q does not say whose criterion or which page — the other five causes "+
			"would send a reader looking for a handler that failed, and this one worked", msg)
	}
}

// ---------------------------------------------------------------------------
// The controls: this gate must be inert everywhere it has not been armed
// ---------------------------------------------------------------------------

// TestAnUnoptedTypeIsNeverGraded is the opt-in guarantee, asserted where it can
// actually fail: sqlmock queues NO page query, so if the gate reaches the surface
// read for an unopted type the mock errors.
func TestAnUnoptedTypeIsNeverGraded(t *testing.T) {
	payload, mayComplete, note := gateRun(t, "spacing_fix",
		specWithPredicate(t, liveStoredPredicate(t)), liveIndexMeta, false)

	if !mayComplete {
		t.Fatal("an unopted type was blocked — this gate is not opt-in")
	}
	if note != nil {
		t.Fatalf("an unopted type produced a note: %+v", note)
	}
	if payload != nil {
		t.Fatalf("an unopted type had a verdict recorded: %#v.\nA type absent from the roster must be "+
			"byte-identical to today, which is what \"the unsafe default is OFF\" has to mean", payload)
	}
}

// TestAnItemWithNoPredicateIsNotGraded covers the common case by a wide margin: most
// content_rewrite items come from producers that write no predicate. Again the page
// query is deliberately not queued.
func TestAnItemWithNoPredicateIsNotGraded(t *testing.T) {
	payload, mayComplete, note := gateRun(t, "content_rewrite", specWithPredicate(t, nil), liveIndexMeta, false)

	if !mayComplete || note != nil {
		t.Fatalf("an item with no predicate was not passed through cleanly (mayComplete=%v note=%+v)", mayComplete, note)
	}
	if payload != nil {
		t.Fatalf("an item with no predicate recorded a verdict: %#v — one row per completion would bury "+
			"the real signal", payload)
	}
}

// TestAnEmptyPageSurfaceBlamesThisGateAndIsRecorded is the emit side's own lesson,
// applied here (corr ef482d1c, editquality + debug_historian): a page query that
// matches nothing makes EVERY predicate unevaluable, and "found nothing to grade"
// is byte-identical to the acceptable outcome "no item carried a predicate". So the
// blind case must be loud, and must not send the next reader to the model.
func TestAnEmptyPageSurfaceBlamesThisGateAndIsRecorded(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()
	mock.ExpectQuery(verifyRowReadSQL).
		WillReturnRows(sqlmock.NewRows([]string{"item_type", "spec", "site_id", "page_id"}).
			AddRow("content_rewrite", specWithPredicate(t, liveStoredPredicate(t)), uuid.New(), nil))
	mock.ExpectQuery("FROM pages").
		WillReturnRows(sqlmock.NewRows([]string{"name", "title", "meta_description"}))

	payload, mayComplete, _, note := verifyBeforeComplete(
		context.Background(), db, uuid.New(), map[string]interface{}{}, zap.NewNop())

	if !mayComplete {
		t.Fatal("an unevaluable predicate blocked a completion — none of its three causes is evidence the handler failed")
	}
	if note == nil {
		t.Fatal("no blind-spot note returned. This is the silent-inert failure: every predicate reads " +
			"unevaluable and nothing is recorded, so the gate looks like it is working and grades nothing")
	}
	if !strings.Contains(note.Detail, "fault in this gate") {
		t.Errorf("the recorded reason %q blames the predicate or the page. With an EMPTY surface no page is "+
			"on it, and naming one sends the next reader to the model when the fault is the query's", note.Detail)
	}
	pred := predicatePayload(t, payload)
	if pred["outcome"] != "not_evaluable" {
		t.Errorf("outcome = %v, want not_evaluable", pred["outcome"])
	}
}

// ---------------------------------------------------------------------------
// The roster's own contract
// ---------------------------------------------------------------------------

// TestAcceptancePredicateRosterCarriesItsEvidence refuses an entry that ships
// without the declaration it is supposed to be making. It is the sibling of
// TestNoChangeGatesRosterCarriesItsEvidence and exists for the same reason: the zero
// value of OnRefuted is not a policy, and an entry written by somebody who never read
// the file must not be able to arm or excuse anything by accident.
func TestAcceptancePredicateRosterCarriesItsEvidence(t *testing.T) {
	for itemType, rule := range acceptancePredicateGates {
		if strings.TrimSpace(rule.Why) == "" {
			t.Errorf("%q: no Why. An entry without one is a guess about somebody else's producer", itemType)
		}
		switch rule.OnRefuted {
		case predicateUndeclared:
			t.Errorf("%q: OnRefuted is the zero value, which is NOT a policy. Declare predicateRecords or "+
				"predicateRefuses explicitly — see predicateGateOutcome.", itemType)
		case predicateRefuses:
			if strings.TrimSpace(rule.RefusalWhy) == "" {
				t.Errorf("%q: predicateRefuses without RefusalWhy. Blocking a completion on a shared path "+
					"needs the measurement that a still-refuting criterion here cannot be a repair.", itemType)
			}
		case predicateRecords:
			if strings.TrimSpace(rule.PromotionOwes) == "" {
				t.Errorf("%q: predicateRecords without PromotionOwes. A record-only arm with no stated debt "+
					"is how the arm becomes permanent — which is CLM-023's residual, twice over.", itemType)
			}
		}
	}
}

// TestAcceptancePredicateLookupIsNotASwitchInProduction — the fixture seam above must
// be a fixture. Mirrors TestVerifierLookupIsNotASwitchInProduction exactly.
func TestAcceptancePredicateLookupIsNotASwitchInProduction(t *testing.T) {
	files, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatal(err)
	}
	needle := "acceptancePredicateGateFor" + " ="
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
	// Exactly one: the declaration. The comment above it is written to avoid
	// spelling the assignment, so this cannot pass vacuously on prose describing
	// itself.
	if found != 1 {
		t.Errorf("found %d assignments to the roster lookup seam in non-test source, want exactly 1 "+
			"(the declaration). Production must always read acceptancePredicateGates; re-pointing it turns "+
			"every test in this file into a test of its own fixture.", found)
	}
}
