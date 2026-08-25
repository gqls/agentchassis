// FILE: platform/orchestration/actions/verify_acceptance_predicates_action_test.go
//
// EVERY STRING IN THE CORPUS BELOW IS REAL, and that is the point rather than
// realism for its own sake. The design question this action answers — which
// acceptance-test clauses can be checked mechanically, and which look checkable
// and are not — cannot be settled from invented examples: it turns entirely on how
// the model actually writes, and on what the pages actually say. Each was read
// from the live estate on 2026-08-24 (site_specs, site_work_items, pages) and the
// served artefact was checked for the load-bearing one: the meta tag served by
// https://webdesign.co.uk/ is byte-identical to pages.meta_description.
//
// The two most valuable tests here are the ones that would have FAILED an obvious
// implementation:
//   - TestWordBoundariesPreventTheFalseRefutation — a plain substring search finds
//     "we" inside "web" on a site about web design, refuting a test that holds.
//   - TestAPredicateThatAlreadyHoldsIsRejected — the vacuous case, which is the
//     one that would grade green for ever.

package actions

import (
	"context"
	"reflect"
	"strings"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/gqls/agentchassis/pkg/models"
	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	orchtypes "github.com/gqls/agentchassis/platform/orchestration/types"
	"go.uber.org/zap"
)

// ---- the live corpus, verbatim (read 2026-08-24) --------------------------

const (
	// webdesign.co.uk / index. THE WORKED CASE. site_work_items row
	// 08-24 10:13, status `complete` at 10:34 with a commit sha and a deploy —
	// and its own acceptance test is refuted by this string:
	// "The index page meta description must state the zero-data or zero-account
	//  promise before any catalogue count or category list."
	wdIndexMeta = "Sixty-three browser tools for web design and development. No account, no upload, everything runs in your browser."

	// webdesign.co.uk / tools-index. The same failure in DIGIT form.
	wdToolsIndexMeta = "63 browser-based tools for design, colour, layout and creative workflow. Nothing uploads, nothing installs, nothing asks you to sign in."

	// webdesign.co.uk / about. The tone_shift finding's test: "must not contain
	// the word 'curated' … must contain at least one of 'no account', 'no
	// upload', 'nothing stored', 'client-side'". Both halves fail.
	wdAboutMeta = "Curated guides and tools for modern web development, from HTML semantics to responsive design."

	// webdesign.co.uk / learn-index. The NEGATIVE CONTROL for word boundaries:
	// an earlier census entry asked for "no instance of 'we' or 'our'", and this
	// string contains neither as a word — but contains "web" twice.
	wdLearnIndexMeta = "Learn the reasoning behind web design: why breakpoints fail, how to use math for layouts, and when to break the rules."
)

// wdIndexFixed is the ONLY invented string in this file, and it is a MUTATION of
// wdIndexMeta rather than a fresh sentence: the same clauses, reordered, which is
// exactly what the finding asks for. It is what a correct fix looks like, and the
// predicate must stop refuting it. A gate that cannot be turned off by the repair
// it demands is not a gate, it is a permanent alarm.
const wdIndexFixed = "No account, no upload, everything runs in your browser. Sixty-three browser tools for web design and development."

func predSubject(page, title, meta string) AcceptancePredicateSubject {
	return AcceptancePredicateSubject{Page: page, Title: title, MetaDescription: meta}
}

func orderPred(page string, before []interface{}, after []interface{}) map[string]interface{} {
	return map[string]interface{}{
		"type": "text_order", "page": page, "field": "meta_description",
		"before": before, "after": after,
	}
}

// ---- the evaluator, over live strings ------------------------------------

func TestTheWorkedFailureIsRefutedByItsOwnField(t *testing.T) {
	pred := orderPred("index",
		[]interface{}{"no account", "no upload", "nothing stored"},
		[]interface{}{cardinalNeedle})

	verdict, reason := EvaluateAcceptancePredicate(pred, predSubject("index", "", wdIndexMeta))
	if verdict != PredicateRefutes {
		t.Fatalf("verdict = %s (%s); the live meta opens with a word cardinal and must refute", verdict, reason)
	}
	// The reason has to name both sides AND their offsets, because a reader who
	// cannot see them cannot tell a real refutation from a confident one. Note
	// the case: the evidence is quoted AS THE PAGE WROTE IT ("No account", not
	// the needle "no account"), which is what makes it checkable against the
	// served string by eye.
	if !strings.Contains(reason, "Sixty-three") || !strings.Contains(reason, "No account") {
		t.Errorf("reason does not quote both sides in the page's own case: %s", reason)
	}
	if !strings.Contains(reason, " at 0,") {
		t.Errorf("reason does not carry the offsets: %s", reason)
	}
}

func TestTheFixTurnsTheSamePredicateOff(t *testing.T) {
	pred := orderPred("index",
		[]interface{}{"no account", "no upload", "nothing stored"},
		[]interface{}{cardinalNeedle})

	verdict, reason := EvaluateAcceptancePredicate(pred, predSubject("index", "", wdIndexFixed))
	if verdict != PredicateHolds {
		t.Fatalf("verdict = %s (%s); the reordered clauses satisfy the very test that demanded them", verdict, reason)
	}
}

func TestDigitFormCountsToo(t *testing.T) {
	pred := orderPred("tools-index",
		[]interface{}{"nothing uploads", "nothing installs"},
		[]interface{}{cardinalNeedle})

	verdict, _ := EvaluateAcceptancePredicate(pred, predSubject("tools-index", "", wdToolsIndexMeta))
	if verdict != PredicateRefutes {
		t.Fatalf("verdict = %s; \"63\" at offset 0 is a count and must refute", verdict)
	}
}

// TestOnlyTheWordScannerCatchesTheWordForm is the discrimination proof: strip the
// word half of the cardinal scanner and the worked case survives. It is here
// because the same hole is what verify_report_prose has, and why this lane could
// not reuse it (bugs_closed/335).
func TestOnlyTheWordScannerCatchesTheWordForm(t *testing.T) {
	if got := len(digitCardinalSpans(wdIndexMeta)); got != 0 {
		t.Fatalf("digit spans in the worked case = %d, want 0 — the defect is spelled out", got)
	}
	if pos, matched := firstCardinalMatch(wdIndexMeta); pos != 0 || !strings.EqualFold(matched, "Sixty-three") {
		t.Fatalf("firstCardinalMatch = (%d, %q), want (0, \"Sixty-three\")", pos, matched)
	}
}

func TestTextAbsentCatchesTheClaimCaseInsensitively(t *testing.T) {
	pred := map[string]interface{}{
		"type": "text_absent", "page": "about", "field": "meta_description",
		"values": []interface{}{"curated", "hand-picked"},
	}
	verdict, reason := EvaluateAcceptancePredicate(pred, predSubject("about", "", wdAboutMeta))
	if verdict != PredicateRefutes {
		t.Fatalf("verdict = %s (%s); the live meta opens \"Curated\"", verdict, reason)
	}
}

// TestWordBoundariesPreventTheFalseRefutation — the single most important test in
// this file. A substring implementation refutes here, and refuting a page that
// PASSES is worse than not checking it at all.
func TestWordBoundariesPreventTheFalseRefutation(t *testing.T) {
	pred := map[string]interface{}{
		"type": "text_absent", "page": "learn-index", "field": "meta_description",
		"values": []interface{}{"we", "our"},
	}
	verdict, reason := EvaluateAcceptancePredicate(pred, predSubject("learn-index", "", wdLearnIndexMeta))
	if verdict != PredicateHolds {
		t.Fatalf("verdict = %s (%s); \"web\" is not \"we\"", verdict, reason)
	}
	// POSITIVE CONTROL, in the same breath: the needle must still fire on the
	// word it is for, or the test above is passing because nothing works.
	verdict, _ = EvaluateAcceptancePredicate(pred,
		predSubject("learn-index", "", "Learn why we built our layout tools the way we did."))
	if verdict != PredicateRefutes {
		t.Fatalf("control failed: standalone \"we\" must refute, got %s", verdict)
	}
	// And a plain substring search WOULD have refuted the real string — stated
	// as an assertion so the claim in the comment cannot rot.
	if !strings.Contains(strings.ToLower(wdLearnIndexMeta), "we") {
		t.Fatal("the corpus string no longer contains the substring \"we\"; this test's premise has moved")
	}
}

func TestTextPresentCountsDistinctValues(t *testing.T) {
	// leopardessconsulting.co.uk's index finding: "names at least two of the
	// three recorded infrastructure components".
	pred := map[string]interface{}{
		"type": "text_present", "page": "index", "field": "meta_description",
		"values": []interface{}{"Kubernetes", "Kafka", "Postgres"}, "min": float64(2),
	}
	verdict, reason := EvaluateAcceptancePredicate(pred,
		predSubject("index", "", "We build production AI agent systems on Kubernetes for UK engineering teams."))
	if verdict != PredicateRefutes {
		t.Fatalf("verdict = %s (%s); one of three is fewer than two", verdict, reason)
	}
	verdict, reason = EvaluateAcceptancePredicate(pred,
		predSubject("index", "", "Agent systems on Kubernetes and Kafka, with Postgres workflow state."))
	if verdict != PredicateHolds {
		t.Fatalf("verdict = %s (%s); three of three satisfies min 2", verdict, reason)
	}
}

func TestAnEmptyMetaRefutesAnOrderClaim(t *testing.T) {
	// A real state on this estate, and a genuine failure of "state X before Y":
	// nothing is stated at all.
	verdict, reason := EvaluateAcceptancePredicate(
		orderPred("contact", []interface{}{"no account"}, []interface{}{cardinalNeedle}),
		predSubject("contact", "Contact", ""))
	if verdict != PredicateRefutes {
		t.Fatalf("verdict = %s (%s)", verdict, reason)
	}
}

func TestNothingToBeBeforeIsNotAFailure(t *testing.T) {
	// The promise is stated and the page contains no count at all: the ordering
	// clause is satisfied vacuously, and vacuously satisfied is HOLDS, which is
	// why such a predicate is refused at emission rather than stored.
	verdict, _ := EvaluateAcceptancePredicate(
		orderPred("index", []interface{}{"no account"}, []interface{}{cardinalNeedle}),
		predSubject("index", "", "No account needed, and nothing you type leaves the browser."))
	if verdict != PredicateHolds {
		t.Fatalf("verdict = %s, want holds", verdict)
	}
}

// ---- the refusals ---------------------------------------------------------

func TestUnknownTypeIsInapplicableNotAPass(t *testing.T) {
	verdict, reason := EvaluateAcceptancePredicate(map[string]interface{}{
		"type": "selector_exists", "page": "index", "field": "meta_description",
		"values": []interface{}{"x"},
	}, predSubject("index", "", wdIndexMeta))
	if verdict != PredicateInapplicable {
		t.Fatalf("verdict = %s, want inapplicable", verdict)
	}
	if !strings.Contains(reason, "text_order") {
		t.Errorf("the refusal must name what IS evaluable: %s", reason)
	}
}

func TestAKeyNoCheckerReadsRejectsThePredicate(t *testing.T) {
	// P7's rule, borrowed from experience_criteria.go: a field the runner never
	// reads makes the artefact appear to assert something it does not.
	verdict, reason := EvaluateAcceptancePredicate(map[string]interface{}{
		"type": "text_absent", "page": "about", "field": "meta_description",
		"values": []interface{}{"curated"}, "selector": ".hero h1",
	}, predSubject("about", "", wdAboutMeta))
	if verdict != PredicateInapplicable {
		t.Fatalf("verdict = %s, want inapplicable (the selector is inert here)", verdict)
	}
	if !strings.Contains(reason, "selector") {
		t.Errorf("the refusal must name the offending key: %s", reason)
	}
}

func TestBodyTextIsNotReadable(t *testing.T) {
	// The offer surface passes no page content, so a predicate over body copy is
	// a claim about something its author never read (features_open/030 v2(a)).
	verdict, reason := EvaluateAcceptancePredicate(map[string]interface{}{
		"type": "text_absent", "page": "index", "field": "rendered_html",
		"values": []interface{}{"curated"},
	}, predSubject("index", "", wdIndexMeta))
	if verdict != PredicateInapplicable || !strings.Contains(reason, "meta_description and title") {
		t.Fatalf("verdict = %s (%s)", verdict, reason)
	}
}

func TestTheCardinalNeedleIsRefusedOutsideAfter(t *testing.T) {
	for _, pred := range []map[string]interface{}{
		{"type": "text_absent", "page": "index", "field": "meta_description", "values": []interface{}{cardinalNeedle}},
		{"type": "text_present", "page": "index", "field": "meta_description", "values": []interface{}{cardinalNeedle}},
		orderPred("index", []interface{}{cardinalNeedle}, []interface{}{"no account"}),
	} {
		verdict, reason := EvaluateAcceptancePredicate(pred, predSubject("index", "", wdIndexMeta))
		if verdict != PredicateInapplicable {
			t.Errorf("%s: verdict = %s (%s), want inapplicable", pred["type"], verdict, reason)
		}
	}
}

func TestAnUnsatisfiableMinIsRefused(t *testing.T) {
	verdict, reason := EvaluateAcceptancePredicate(map[string]interface{}{
		"type": "text_present", "page": "index", "field": "meta_description",
		"values": []interface{}{"Kafka"}, "min": float64(3),
	}, predSubject("index", "", wdIndexMeta))
	if verdict != PredicateInapplicable || !strings.Contains(reason, "never hold") {
		t.Fatalf("verdict = %s (%s)", verdict, reason)
	}
}

func TestMalformedNeedleListsAreRefused(t *testing.T) {
	cases := []map[string]interface{}{
		{"type": "text_absent", "page": "a", "field": "title"},
		{"type": "text_absent", "page": "a", "field": "title", "values": []interface{}{}},
		{"type": "text_absent", "page": "a", "field": "title", "values": []interface{}{""}},
		{"type": "text_absent", "page": "a", "field": "title", "values": []interface{}{float64(3)}},
		{"type": "text_present", "page": "a", "field": "title", "values": []interface{}{"x"}, "min": "two"},
	}
	for i, pred := range cases {
		if verdict, _ := EvaluateAcceptancePredicate(pred, predSubject("a", "T", "M")); verdict != PredicateInapplicable {
			t.Errorf("case %d: verdict = %s, want inapplicable", i, verdict)
		}
	}
}

func TestASingleStringWhereAnArrayBelongsIsAccepted(t *testing.T) {
	verdict, _ := EvaluateAcceptancePredicate(map[string]interface{}{
		"type": "text_absent", "page": "about", "field": "meta_description",
		"values": "curated",
	}, predSubject("about", "", wdAboutMeta))
	if verdict != PredicateRefutes {
		t.Fatalf("verdict = %s, want refutes", verdict)
	}
}

// ---- the action ----------------------------------------------------------

func predicateParams(t *testing.T, db interface{ Close() error }, findings interface{}, cfg map[string]interface{}) ActionParams {
	t.Helper()
	base := map[string]interface{}{
		"site_id":        "site_record.site_id",
		"findings_field": "analysis.result.findings",
	}
	for k, v := range cfg {
		base[k] = v
	}
	return ActionParams{
		Context:          context.Background(),
		Logger:           zap.NewNop(),
		ExecutionContext: &orchtypes.ExecutionContext{Action: "process", StepName: "gate"},
		StepConfig:       models.Step{Action: "verify_acceptance_predicates", Config: base},
		CollectedData: map[string]interface{}{
			"site_record": map[string]interface{}{"site_id": uuid.New().String()},
			"analysis":    map[string]interface{}{"result": map[string]interface{}{"findings": findings}},
		},
	}
}

// livePagesMock returns a mock DB whose pages query answers with the real
// webdesign.co.uk rows.
func livePagesMock(t *testing.T) (*sqlmock.Sqlmock, ActionParams, func()) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	mock.ExpectQuery("FROM pages").WillReturnRows(
		sqlmock.NewRows([]string{"name", "title", "meta_description"}).
			AddRow("index", "Web Design Tools", wdIndexMeta).
			AddRow("about", "About", wdAboutMeta).
			AddRow("learn-index", "Learn", wdLearnIndexMeta))
	params := ActionParams{}
	params.DB = db
	return &mock, params, func() { db.Close() }
}

func runPredicateGate(t *testing.T, findings interface{}) map[string]interface{} {
	t.Helper()
	_, dbParams, closeDB := livePagesMock(t)
	defer closeDB()
	params := predicateParams(t, nil, findings, nil)
	params.DB = dbParams.DB
	out, err := VerifyAcceptancePredicatesAction(context.Background(), params)
	if err != nil {
		t.Fatalf("action error: %v", err)
	}
	res, ok := out.(map[string]interface{})
	if !ok {
		t.Fatalf("result is %T", out)
	}
	return res
}

func TestARefutingPredicateIsKeptWithItsEvidence(t *testing.T) {
	res := runPredicateGate(t, []interface{}{
		map[string]interface{}{
			"category": "content", "page": "index", "acceptance_test": "prose",
			"acceptance_predicate": orderPred("", []interface{}{"no account"}, []interface{}{cardinalNeedle}),
		},
	})
	if res["kept"] != 1 || res["rejected"] != 0 {
		t.Fatalf("kept=%v rejected=%v", res["kept"], res["rejected"])
	}
	f := res["findings"].([]interface{})[0].(map[string]interface{})
	stored, ok := f[acceptancePredicateKey].(map[string]interface{})
	if !ok {
		t.Fatalf("the predicate was not stored: %v", f)
	}
	if stored["verdict_at_emission"] != string(PredicateRefutes) || stored["evidence_at_emission"] == "" {
		t.Errorf("a stored predicate must carry the verdict that earned it its place: %v", stored)
	}
	if _, present := f[acceptancePredicateRejectedKey]; present {
		t.Error("a kept predicate must not also be recorded as rejected")
	}
}

// TestAPredicateThatAlreadyHoldsIsRejected — the vacuous case. Without this rule
// the artefact stores a check that can never fail and reads as verification.
func TestAPredicateThatAlreadyHoldsIsRejected(t *testing.T) {
	res := runPredicateGate(t, []interface{}{
		map[string]interface{}{
			"category": "content", "page": "about", "acceptance_test": "prose",
			// "shibboleth" appears in no live meta description, so this can
			// never refute anything: exactly the shape that must not be stored.
			"acceptance_predicate": map[string]interface{}{
				"type": "text_absent", "field": "meta_description",
				"values": []interface{}{"shibboleth"},
			},
		},
	})
	if res["kept"] != 0 || res["rejected"] != 1 {
		t.Fatalf("kept=%v rejected=%v", res["kept"], res["rejected"])
	}
	f := res["findings"].([]interface{})[0].(map[string]interface{})
	if _, present := f[acceptancePredicateKey]; present {
		t.Error("a refused predicate must not remain under the key a consumer reads")
	}
	rej, ok := f[acceptancePredicateRejectedKey].(AcceptancePredicateRejection)
	if !ok {
		t.Fatalf("no rejection record: %v", f)
	}
	if rej.Verdict != string(PredicateHolds) || rej.Predicate == nil {
		t.Errorf("the rejection must say what it was and why: %+v", rej)
	}
	// The finding itself survives intact — the prose is the valuable part.
	if f["acceptance_test"] != "prose" || f["category"] != "content" {
		t.Errorf("the finding was altered beyond its predicate keys: %v", f)
	}
}

func TestSilenceIsFreeAndLeavesNothingBehind(t *testing.T) {
	res := runPredicateGate(t, []interface{}{
		map[string]interface{}{"category": "tone", "page": "index", "acceptance_test": "judgement only"},
	})
	if res["checked"] != 1 || res["kept"] != 0 || res["rejected"] != 0 {
		t.Fatalf("checked=%v kept=%v rejected=%v", res["checked"], res["kept"], res["rejected"])
	}
	f := res["findings"].([]interface{})[0].(map[string]interface{})
	if len(f) != 3 {
		t.Errorf("a finding with no predicate must pass through untouched, got %v", f)
	}
}

func TestAPageNotOnTheSurfaceIsRefusedNotAssumed(t *testing.T) {
	res := runPredicateGate(t, []interface{}{
		map[string]interface{}{
			"category": "gap", "page": "pricing", "acceptance_test": "prose",
			"acceptance_predicate": map[string]interface{}{
				"type": "text_absent", "field": "meta_description", "values": []interface{}{"curated"},
			},
		},
	})
	if res["rejected"] != 1 {
		t.Fatalf("rejected=%v; a predicate about a page that does not exist cannot be evaluated", res["rejected"])
	}
	f := res["findings"].([]interface{})[0].(map[string]interface{})
	rej := f[acceptancePredicateRejectedKey].(AcceptancePredicateRejection)
	if rej.Verdict != string(PredicateInapplicable) || !strings.Contains(rej.Reason, "pricing") {
		t.Errorf("the refusal must name the page: %+v", rej)
	}
}

func TestHomepageAliasResolvesTheSameWayTheWriteDoes(t *testing.T) {
	res := runPredicateGate(t, []interface{}{
		map[string]interface{}{
			"category": "content", "page": "homepage", "acceptance_test": "prose",
			"acceptance_predicate": orderPred("", []interface{}{"no account"}, []interface{}{cardinalNeedle}),
		},
	})
	if res["kept"] != 1 {
		t.Fatalf("kept=%v; \"homepage\" must resolve to \"index\" here exactly as it does in classifyFinding", res["kept"])
	}
}

// TestAnUnresolvableFindingsPathOmitsTheKey guards the retraction trap: an empty
// array is write_audit_findings' "the auditor found nothing", which ARMS silence
// retraction. A path that did not resolve is not that statement.
func TestAnUnresolvableFindingsPathOmitsTheKey(t *testing.T) {
	_, dbParams, closeDB := livePagesMock(t)
	defer closeDB()
	params := predicateParams(t, nil, nil, map[string]interface{}{"findings_field": "analysis.result.nowhere"})
	params.DB = dbParams.DB
	out, err := VerifyAcceptancePredicatesAction(context.Background(), params)
	if err != nil {
		t.Fatalf("an unresolvable path must not fail the run: %v", err)
	}
	res := out.(map[string]interface{})
	if _, present := res["findings"]; present {
		t.Error("the findings key must be ABSENT, so write_audit_findings takes its own nil branch and retracts nothing")
	}
	if res["findings_resolved"] != false {
		t.Errorf("findings_resolved = %v, want false", res["findings_resolved"])
	}
}

func TestAGenuinelyEmptyArrayIsPassedThrough(t *testing.T) {
	res := runPredicateGate(t, []interface{}{})
	list, present := res["findings"]
	if !present {
		t.Fatal("a resolved-but-empty findings array must survive as an empty array: it is the auditor's 'nothing wrong here'")
	}
	if l, ok := list.([]interface{}); !ok || len(l) != 0 {
		t.Errorf("findings = %v", list)
	}
}

func TestEveryFindingSurvivesEvenWhenEveryPredicateIsRefused(t *testing.T) {
	bad := map[string]interface{}{"type": "nonsense"}
	res := runPredicateGate(t, []interface{}{
		map[string]interface{}{"category": "content", "page": "index", "acceptance_predicate": bad},
		map[string]interface{}{"category": "tone", "page": "about", "acceptance_predicate": "not an object"},
		map[string]interface{}{"category": "gap", "page": "learn-index"},
	})
	if res["checked"] != 3 || res["rejected"] != 2 || res["kept"] != 0 {
		t.Fatalf("checked=%v kept=%v rejected=%v", res["checked"], res["kept"], res["rejected"])
	}
	if got := len(res["findings"].([]interface{})); got != 3 {
		t.Fatalf("findings out = %d, want 3 — this gate must never remove a finding", got)
	}
}

func TestMissingSiteIDFailsLoudly(t *testing.T) {
	// The opposite policy from the predicate arm, deliberately: a misconfigured
	// step must not run 130 findings past a gate that could not read a page.
	_, dbParams, closeDB := livePagesMock(t)
	defer closeDB()
	params := predicateParams(t, nil, []interface{}{}, nil)
	params.DB = dbParams.DB
	params.CollectedData["site_record"] = map[string]interface{}{}
	if _, err := VerifyAcceptancePredicatesAction(context.Background(), params); err == nil {
		t.Fatal("a step that cannot resolve site_id must fail rather than silently reject every predicate")
	}
}

// TestFindingsFromListPopulatesEveryTaggedField is the lockstep this change owed
// itself. write_audit_findings decodes findings TWICE — through struct tags for a
// JSON string, and by hand for a native list — and the native list is the shape
// an upstream action hands over. A field added to the struct and forgotten in
// findingsFromList is dropped in silence, on the only path this feature uses.
func TestFindingsFromListPopulatesEveryTaggedField(t *testing.T) {
	item := map[string]interface{}{
		"category": "cta", "severity": "high", "description": "d", "suggestion": "s",
		"page": "index", "fix_type": "cta_improvement", "affected_component": "hero",
		"current_value": "cv", "acceptance_test": "at", "max_fix_attempts": float64(2),
		"affected_pages":                []interface{}{"index"},
		"acceptance_predicate":          map[string]interface{}{"type": "text_absent"},
		"acceptance_predicate_rejected": map[string]interface{}{"verdict": "holds"},
	}
	got, recognised := findingsFromList([]interface{}{item})
	if !recognised || len(got) != 1 {
		t.Fatalf("recognised=%v n=%d", recognised, len(got))
	}
	v := reflect.ValueOf(got[0])
	typ := v.Type()
	for i := 0; i < typ.NumField(); i++ {
		tag := strings.Split(typ.Field(i).Tag.Get("json"), ",")[0]
		if tag == "" || tag == "-" {
			continue
		}
		if _, offered := item[tag]; !offered {
			t.Errorf("field %s (json %q) is not covered by this fixture — extend the fixture", typ.Field(i).Name, tag)
			continue
		}
		if v.Field(i).IsZero() {
			t.Errorf("findingsFromList dropped %q: the struct has the field, the hand-written map path does not", tag)
		}
	}
}

// TestAnEmptySubjectSetIsReportedAsAFaultHere is the council's medium objection
// made mechanical (editquality + debug_historian, corr ef482d1c): if the page
// query ever stops matching, every predicate is refused and the gate goes inert
// while reading exactly like "the model wrote nothing storable today". The
// measured premise of that risk does not hold on `pages` right now — status is
// `active` 805 / `archived` 66 as of 2026-08-24, and the query returns 35-137
// rows per enrolled site — so this test MUTATES the state the objection is about
// rather than waiting for it: an empty subject set must produce a rejection that
// blames the gate, and a `subjects_loaded` of 0 in the record.
func TestAnEmptySubjectSetIsReportedAsAFaultHere(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()
	// The mutation: the query matches nothing.
	mock.ExpectQuery("FROM pages").WillReturnRows(
		sqlmock.NewRows([]string{"name", "title", "meta_description"}))

	params := predicateParams(t, nil, []interface{}{
		map[string]interface{}{
			"category": "content", "page": "index", "acceptance_test": "prose",
			"acceptance_predicate": orderPred("", []interface{}{"no account"}, []interface{}{cardinalNeedle}),
		},
	}, nil)
	params.DB = db

	out, err := VerifyAcceptancePredicatesAction(context.Background(), params)
	if err != nil {
		t.Fatalf("an empty page set must not fail the run — the findings still deserve filing: %v", err)
	}
	res := out.(map[string]interface{})
	if res["subjects_loaded"] != 0 {
		t.Errorf("subjects_loaded = %v, want 0 stated positively in the record", res["subjects_loaded"])
	}
	if res["kept"] != 0 || res["rejected"] != 1 {
		t.Fatalf("kept=%v rejected=%v", res["kept"], res["rejected"])
	}
	f := res["findings"].([]interface{})[0].(map[string]interface{})
	rej := f[acceptancePredicateRejectedKey].(AcceptancePredicateRejection)
	if strings.Contains(rej.Reason, "\"index\"") || !strings.Contains(rej.Reason, "fault in the gate") {
		t.Errorf("the refusal must blame this step, not the named page: %q", rej.Reason)
	}
	// And the finding itself still goes through.
	if f["acceptance_test"] != "prose" {
		t.Errorf("the finding was damaged: %v", f)
	}
}

// TestTheLifecycleArmIsTheSharedHelper pins the query's lifecycle predicate to
// datahelpers.PageWantedLivePredicateFor rather than a hand-written literal.
// `pages.status` has two live values and two of the spellings in circulation are
// INERT (`<> 'deleted'` excludes nothing; `IN ('active','deployed')` works by
// accident) — the helper is where that vocabulary is written down once.
func TestTheLifecycleArmIsTheSharedHelper(t *testing.T) {
	if got := datahelpers.PageWantedLivePredicateFor(""); got != "status = 'active'" {
		t.Fatalf("the helper now returns %q — this action's query embeds it, so re-read the "+
			"offer surface's own filter and confirm the two still select the same population", got)
	}
}

// ---- the first LIVE-EMITTED predicates, and what they say about a closed item -

// TestTheFirstLiveEmittedPredicatesStillRefuteAfterTheFix pins the three
// predicates the model actually wrote on the first live run of this gate
// (webdesign.co.uk, 2026-08-24 22:08Z, corr 4caba084) against the pages as
// SERVED the following morning — verbatim from `pages.meta_description`, and
// confirmed byte-identical to the served `<meta name="description">`.
//
// WHY THIS IS THE MOST VALUABLE TEST IN THE FILE. The `index` finding was
// dispatched to page-build-handler, which REBUILT AND DEPLOYED the page
// (commit ee88ba3c, 22:25Z) and closed the item `complete` — and the page was
// rebuilt again at 2026-08-25 11:23Z. Its own stored predicate still refutes,
// because the meta description is unchanged. That is the false-green this
// feature exists to make visible, demonstrated end to end on a real item with a
// machine-checkable verdict attached to it, rather than read by eye.
//
// It also pins the needle sets THE MODEL CHOSE, which no hand-written fixture
// would have guessed: for `index` it went for the differentiator words
// ("paired", "pairing", "article", "guide") rather than the promise words, so
// the refutation arrives through the "states none of `before` at all" arm — the
// arm that exists because "state X before Y" is unmet when X is absent.
func TestTheFirstLiveEmittedPredicatesStillRefuteAfterTheFix(t *testing.T) {
	cases := []struct {
		page, meta string
		pred       map[string]interface{}
		wantIn     string
	}{
		{
			// status `complete` — rebuilt, deployed, closed. Still refuted.
			page: "index",
			meta: "Sixty-three browser tools for web design and development. No account, no upload, everything runs in your browser.",
			pred: map[string]interface{}{
				"type": "text_order", "page": "index", "field": "meta_description",
				"before": []interface{}{"paired", "pairing", "article", "guide"},
				"after":  []interface{}{cardinalNeedle},
			},
			wantIn: "states none of",
		},
		{
			page: "about",
			meta: "Curated guides and tools for modern web development, from HTML semantics to responsive design.",
			pred: map[string]interface{}{
				"type": "text_absent", "page": "about", "field": "meta_description",
				"values": []interface{}{"curated", "hand-picked"},
			},
			wantIn: `contains "Curated" at 0`,
		},
		{
			page: "tools-index",
			meta: "63 browser-based tools for design, colour, layout and creative workflow. Nothing uploads, nothing installs, nothing asks you to sign in.",
			pred: map[string]interface{}{
				"type": "text_order", "page": "tools-index", "field": "meta_description",
				"before": []interface{}{"nothing uploads", "nothing installs", "no account", "no sign"},
				"after":  []interface{}{cardinalNeedle},
			},
			wantIn: `"63" appears at 0`,
		},
	}
	for _, c := range cases {
		verdict, reason := EvaluateAcceptancePredicate(c.pred, predSubject(c.page, "", c.meta))
		if verdict != PredicateRefutes {
			t.Errorf("%s: verdict = %s (%s), want refutes", c.page, verdict, reason)
			continue
		}
		if !strings.Contains(reason, c.wantIn) {
			t.Errorf("%s: reason %q does not carry %q — the evidence shape has changed", c.page, reason, c.wantIn)
		}
	}
}

// TestTheLiveRunsSilentFindingHadNoPredicateToEvaluate records the other half of
// that run: of four findings the model wrote three predicates and left one
// (learn-index) bare. The gate reported checked=4 kept=3 rejected=0 with
// subjects_loaded=137. There is nothing to evaluate for a bare finding, and this
// test exists to state that the silence arm FIRED on its first live outing —
// unprompted, since the key is deliberately absent from the prompt's OUTPUT
// skeleton. ⚠ rejected=0 also means the REFUSAL arm has never fired in
// production, exactly like CLM-023's enforcement arm: it is proven by the
// mutation tests above, not in the wild, and a clean run must never be quoted as
// evidence that it works.
func TestTheLiveRunsSilentFindingHadNoPredicateToEvaluate(t *testing.T) {
	res := runPredicateGate(t, []interface{}{
		map[string]interface{}{"category": "content", "page": "learn-index", "acceptance_test": "prose"},
	})
	if res["checked"] != 1 || res["kept"] != 0 || res["rejected"] != 0 {
		t.Fatalf("checked=%v kept=%v rejected=%v", res["checked"], res["kept"], res["rejected"])
	}
	if res["subjects_loaded"] == 0 {
		t.Error("subjects_loaded must be a positive count here — 0 is the gate's own fault signal")
	}
}
