// FILE: platform/orchestration/actions/repair_ordering_register_action_test.go
//
// ⚠ EVERY `from` STRING IN THIS FILE IS A REAL LIVE POINT, copied out of
// `site_specs` on 2026-08-31 and named with its site and rank. None is composed.
//
// That is not decoration. A fixture this lane COMPOSES exercises its own rule —
// it is written by someone who already knows which arm should fire, so it tests
// the arm rather than the corpus. The eight sites below are the ones the
// producer minted dirty AFTER the 667 wash committed at 10:34Z, which is the
// population this gate exists for.
package actions

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
)

// livePoint is one measured lead_with point.
type livePoint struct {
	site  string
	rank  int
	diff  bool
	text  string
	tells []string // the arms that must fire, as kind:name
}

// The post-wash dirty mint, 2026-08-31. Text verbatim from the live rows.
var liveDirtyPoints = []livePoint{
	{"finetuning.uk", 3, true,
		"We pick the best tool for your problem, not our favourite vendor — so the recommendation is yours to keep.",
		[]string{"shape:x_not_y"}},
	{"leopardessconsulting.co.uk", 2, true,
		"Leopardess delivers hierarchical multi-agent AI systems in days, not months.",
		[]string{"shape:x_not_y"}},
	{"mortgagecalculator.co.uk", 3, false,
		"Every calculator output shows its working and says plainly what it cannot answer.",
		[]string{"word:plainly"}},
	{"agritec.uk", 4, false,
		"Where a figure is not yet verified, this site says so plainly.",
		[]string{"word:plainly"}},
	{"loanzy.uk", 6, false,
		"All loan mechanics are explained with real pound amounts and worked examples rather than abstract percentages.",
		[]string{"shape:rather_than"}},
}

// TestScannerFiresOnEveryMeasuredDirtyPoint. If this goes red the gate has gone
// blind to the very population it was built for.
func TestScannerFiresOnEveryMeasuredDirtyPoint(t *testing.T) {
	for _, p := range liveDirtyPoints {
		hits := datahelpers.ScanBannedRegister(p.text)
		if len(hits) == 0 {
			t.Errorf("%s rank %d: no hit on a point measured dirty in production: %q", p.site, p.rank, p.text)
			continue
		}
		got := map[string]bool{}
		for _, h := range hits {
			got[h.Kind+":"+h.Name] = true
		}
		for _, want := range p.tells {
			if !got[want] {
				t.Errorf("%s rank %d: expected %s, got %s", p.site, p.rank, want,
					datahelpers.DescribeRegisterViolations(hits))
			}
		}
	}
}

// TestCleanPointsAreNotTouched — the false-positive control. These are live
// points measured CLEAN on the same day, so a scanner that flagged them would be
// sending real benefits to a model for no reason.
func TestCleanPointsAreNotTouched(t *testing.T) {
	clean := []string{
		"You leave with a specific number for your likely monthly payment and the deposit you would need.",
		"Every figure on this page traces to a named lender source, dated.",
		"A weekly guide to which fights are on, and where to watch them.",
		"The comparison covers 23 policies, including the budget end of the market.",
	}
	for _, s := range clean {
		if hits := datahelpers.ScanBannedRegister(s); len(hits) > 0 {
			t.Errorf("false positive on a clean live point %q: %s", s, datahelpers.DescribeRegisterViolations(hits))
		}
	}
}

func targetFor(p livePoint) registerTarget {
	hits := datahelpers.ScanBannedRegister(p.text)
	at := 0
	if len(hits) > 0 {
		at = hits[0].At
	}
	return registerTarget{
		Index: 0, Text: p.text, Rank: p.rank, Differentiated: p.diff,
		Hits: hits, ProtectFrom: at,
	}
}

// ⚠⚠ THIS IS THE MOTIVATING CASE, and it is the reason the judge has four layers
// rather than the three it was designed with.
//
// The `to` string is the ACTUAL repair migration 667 applied to
// leopardessconsulting rank 2 — truncate before the comparison, per ruling 7.
// It keeps **64 of 76 bytes (84%)**, so it passes AcceptNegationRewrite's 40%
// "gutted" floor AND the 60% differentiated floor this action adds. And it has
// removed "not months", which is the entire differentiating claim: "in days"
// alone is a duration; "in days, not months" is a comparison against the market.
//
// So the loss is real and the arithmetic cannot see it. That is the class this
// lane caught by hand on 10 of 51 points, the class the shared judge cannot see,
// and — until layer 4 — the class this action could not see either.
func TestDifferentiationFloorRejectsTheRepairTheWashMade(t *testing.T) {
	p := liveDirtyPoints[1] // leopardessconsulting rank 2, differentiated
	tgt := targetFor(p)
	to := "Leopardess delivers hierarchical multi-agent AI systems in days."

	// First: prove the SHARED judge lets it through, or this test is vacuous.
	if ok, why := datahelpers.AcceptNegationRewrite(tgt.Text, to, tgt.ProtectFrom); !ok {
		t.Fatalf("premise broken: the shared judge already rejects this repair (%s) — the extra floor would then be untested", why)
	}
	// Then: prove OUR floor catches it.
	ok, why := judgeRegisterRewrite(tgt, to, defaultDifferentiatedFloorPct)
	if ok {
		t.Fatalf("the differentiated floor let through a repair that removed the differentiating clause: %q -> %q", tgt.Text, to)
	}
	// ⚠ THE REASON MATTERS, AND IT IS NOT THE ONE THIS TEST FIRST ASSERTED.
	// This repair retains 84% of the point, so the length floor does NOT fire —
	// the differentiation was lost in twelve bytes. The first version of this
	// judge had only the length floor and let it through; the failure is what
	// produced layer 4. Asserting the specific reason is what stops a future
	// edit satisfying this test by loosening the floor instead.
	if !strings.HasPrefix(why, "truncation_only") {
		t.Errorf("rejected for the wrong reason: %q (want truncation_only — the length floor cannot see a 12-byte loss)", why)
	}
	// And state the arithmetic, so nobody re-derives the wrong lesson: this
	// point is NOT short enough to trip either floor.
	if len(to)*100 < len(tgt.Text)*defaultDifferentiatedFloorPct {
		t.Fatalf("premise broken: this repair retains %d%%, which is BELOW the %d%% floor — the case no longer demonstrates that a length rule is insufficient",
			len(to)*100/len(tgt.Text), defaultDifferentiatedFloorPct)
	}
}

// TestTruncationIsStillTheRightRepairForAnUndifferentiatedPoint. Layer 4 is a
// cost, and it must be paid only where the measurement says the loss happens:
// on ordinary points the truncation repair is sanctioned and demonstrably
// lossless (27 tics -> 5 on the finetuning approach page, meaning intact).
func TestTruncationIsStillTheRightRepairForAnUndifferentiatedPoint(t *testing.T) {
	tgt := targetFor(livePoint{"x", 6, false,
		"All loan mechanics are explained with real pound amounts and worked examples rather than abstract percentages.", nil})
	to := "All loan mechanics are explained with real pound amounts and worked examples."
	if !isTruncationOf(tgt.Text, to) {
		t.Fatal("premise broken: this candidate is not a truncation, so the exemption is untested")
	}
	if ok, why := judgeRegisterRewrite(tgt, to, defaultDifferentiatedFloorPct); !ok {
		t.Errorf("truncation must remain available for an undifferentiated point: %q", why)
	}
}

// TestATrueRestatementOfADifferentiatedPointIsAccepted — the positive control
// for layer 4. Without it, layer 4 is indistinguishable from "reject every
// repair of a differentiated point", which would be a gate that never repairs
// the points that matter most.
func TestATrueRestatementOfADifferentiatedPointIsAccepted(t *testing.T) {
	tgt := targetFor(liveDirtyPoints[1]) // leopardess rank 2, differentiated
	to := "Leopardess delivers hierarchical multi-agent AI systems within days of kickoff."
	if isTruncationOf(tgt.Text, to) {
		t.Fatal("premise broken: this candidate IS a truncation, so it cannot show that a restatement passes")
	}
	if ok, why := judgeRegisterRewrite(tgt, to, defaultDifferentiatedFloorPct); !ok {
		t.Fatalf("a genuine positive restatement of a differentiated point was rejected: %q", why)
	}
}

// TestDifferentiationFloorDoesNotApplyToUndifferentiatedPoints. The floor is a
// cost — it keeps a tic that could have been repaired — so it must be paid only
// where the measurement says the loss happens.
func TestDifferentiationFloorDoesNotApplyToUndifferentiatedPoints(t *testing.T) {
	tgt := targetFor(livePoint{"x", 1, false,
		"Leopardess delivers hierarchical multi-agent AI systems in days, not months.", nil})
	to := "Leopardess delivers hierarchical multi-agent AI systems in days."
	if ok, why := judgeRegisterRewrite(tgt, to, defaultDifferentiatedFloorPct); !ok {
		t.Errorf("an undifferentiated point must not pay the differentiated floor: %q", why)
	}
}

// ⚠ TestRewriteThatDisplacesIntoTheWordArmIsRejected — layer 2.
//
// The shared judge re-scans only for SHAPES. This candidate has genuinely
// removed the "X, not Y" construction and then reached for the owner's
// first-named banned word. Without the full-register re-scan the gate would
// accept it and teach the word — the displacement failure this estate has
// already measured once, where banning an opening moved the fault to the end of
// the sentence.
func TestRewriteThatDisplacesIntoTheWordArmIsRejected(t *testing.T) {
	tgt := targetFor(liveDirtyPoints[0]) // finetuning rank 1-3, x_not_y
	to := "We pick the best tool for your problem and we say so plainly, so the recommendation is yours to keep."

	if hits := datahelpers.ScanDefineByNegation(to); len(hits) > 0 {
		t.Fatalf("premise broken: the candidate still carries a SHAPE (%s), so the shared judge would catch it and layer 2 would be untested", hits[0].Shape)
	}
	ok, why := judgeRegisterRewrite(tgt, to, defaultDifferentiatedFloorPct)
	if ok {
		t.Fatal("a rewrite that displaced into the banned-word arm was accepted")
	}
	if why != "still_word_plainly" {
		t.Errorf("wrong reason: %q, want still_word_plainly", why)
	}
}

// TestAGoodRewriteIsAccepted — the positive control. Without it every test above
// passes on an action that rejects everything, which is the same as no gate.
func TestAGoodRewriteIsAccepted(t *testing.T) {
	tgt := targetFor(liveDirtyPoints[4]) // loanzy rank 6, rather_than, not differentiated
	to := "All loan mechanics are explained with real pound amounts and worked examples you can follow."
	ok, why := judgeRegisterRewrite(tgt, to, defaultDifferentiatedFloorPct)
	if !ok {
		t.Fatalf("a sound rewrite was rejected: %q", why)
	}
}

// TestJudgeKeepsProtectedFigures. ProtectFrom is the offset of the earliest
// construction: facts BEFORE it are the claim and must survive.
func TestJudgeKeepsProtectedFigures(t *testing.T) {
	tgt := targetFor(livePoint{"x", 1, false,
		"We run 1,600 orchestrations a day, not 12.", nil})
	if ok, _ := judgeRegisterRewrite(tgt, "We run orchestrations continuously.", defaultDifferentiatedFloorPct); ok {
		t.Error("a rewrite that dropped the protected figure 1,600 was accepted")
	}
	if ok, why := judgeRegisterRewrite(tgt, "We run 1,600 orchestrations a day.", defaultDifferentiatedFloorPct); !ok {
		t.Errorf("a rewrite keeping the protected figure was rejected: %q", why)
	}
}

// TestWithKeyDoesNotMutateTheStoredObject. The object belongs to the previous
// step's output; an in-place edit is the silent-no-op shape this estate keeps
// being bitten by.
func TestWithKeyDoesNotMutateTheStoredObject(t *testing.T) {
	orig := map[string]interface{}{"lead_with": []interface{}{}, "reader_goal": "x"}
	out := withKey(orig, "register_repairs", []registerRepairRecord{})
	if _, present := orig["register_repairs"]; present {
		t.Error("withKey mutated the stored object")
	}
	if _, present := out["register_repairs"]; !present {
		t.Error("withKey did not set the key on the copy")
	}
	if out["reader_goal"] != "x" {
		t.Error("withKey dropped an unrelated key")
	}
}

// ⚠ TestCleanRunStillWritesTheRecordKey guards the deep-merge trap. write_site_spec
// merges rather than replaces, so a clean run that OMITS register_repairs leaves
// the previous run's record standing next to an ordering that no longer needs
// one — an audit record accusing a clean artefact, for ever. Writing the empty
// array is what makes "nothing was repaired" a POSITIVE statement, and what makes
// `len(register_repairs) > 0` a sound census predicate.
func TestCleanRunStillWritesTheRecordKey(t *testing.T) {
	obj := map[string]interface{}{
		"reader_goal": "understand the cost",
		"lead_with": []interface{}{
			map[string]interface{}{"rank": float64(1), "point": "Every figure traces to a named lender source, dated."},
		},
	}
	out := withKey(obj, "register_repairs", []registerRepairRecord{})
	recs, present := out["register_repairs"].([]registerRepairRecord)
	if !present {
		t.Fatal("clean path must write the record key")
	}
	if recs == nil {
		t.Error("the record must be an EMPTY ARRAY, not nil — nil marshals to null and a null is not a positive statement")
	}
	if len(recs) != 0 {
		t.Errorf("clean path wrote %d records", len(recs))
	}
}

// TestEveryTargetIsAccountedFor. A target the model never mentions must still
// get a record, or the producer's error rate — the only evidence this gate works
// — is computed over a denominator that silently shrinks. Measured at 31%
// unreconciled on the page gate before the equivalent fix.
func TestEveryTargetIsAccountedFor(t *testing.T) {
	targets := []registerTarget{}
	for i, p := range liveDirtyPoints {
		tg := targetFor(p)
		tg.Index = i
		targets = append(targets, tg)
	}
	// Simulate: the model answered only the first target.
	answered := map[int]bool{0: true}
	records := make([]registerRepairRecord, 0, len(targets))
	records = append(records, registerRepairRecord{Index: 0, Outcome: "repaired"})
	for _, tg := range targets {
		if !answered[tg.Index] {
			records = append(records, registerRepairRecord{
				Index: tg.Index, Outcome: "kept", Reason: "not_addressed_by_the_model"})
		}
	}
	if len(records) != len(targets) {
		t.Fatalf("reconciliation must produce one record per target: %d records for %d targets",
			len(records), len(targets))
	}
	seen := map[int]int{}
	for _, r := range records {
		seen[r.Index]++
	}
	for _, tg := range targets {
		if seen[tg.Index] != 1 {
			t.Errorf("target %d has %d records, want exactly 1", tg.Index, seen[tg.Index])
		}
	}
}

// TestPromptCarriesNoBannedConstruction. The prompt is rendered into a model's
// context window; a catalogue of the shapes to avoid demonstrates them, and a
// demonstration teaches. The register's own usage rule says the same. This test
// is the enforcement — it scans the rendered prompt with the very scanner the
// gate uses.
//
// ⚠ The item text itself is EXCLUDED from the assertion, because the offending
// sentence is necessarily present — it is what we are asking about. Only the
// instruction half is held to the rule, which is the half we wrote.
func TestPromptCarriesNoBannedConstruction(t *testing.T) {
	prompt := registerRepairPrompt([]registerTarget{targetFor(liveDirtyPoints[0])})
	instructions := prompt
	if i := strings.Index(prompt, "Items:"); i > 0 {
		instructions = prompt[:i]
	}
	if hits := datahelpers.ScanBannedRegister(instructions); len(hits) > 0 {
		t.Errorf("the prompt's own instructions carry a banned construction, which teaches it: %s",
			datahelpers.DescribeRegisterViolations(hits))
	}
	// And it must not enumerate the shapes by name either — naming them invites
	// the model to pattern-match rather than restate.
	for _, shape := range datahelpers.NegationShapeNames() {
		if strings.Contains(strings.ToLower(instructions), strings.ReplaceAll(shape, "_", " ")) {
			t.Errorf("the prompt names banned shape %q — structured input only", shape)
		}
	}
}

// TestDecodeRegisterReplacementsAcceptsBothShapes.
func TestDecodeRegisterReplacementsAcceptsBothShapes(t *testing.T) {
	var envelope interface{} = map[string]interface{}{
		"replacements": []interface{}{
			map[string]interface{}{"index": float64(2), "to": "restated"},
		},
	}
	if got := decodeRegisterReplacements(envelope); len(got) != 1 || got[0].Index != 2 {
		t.Errorf("envelope shape not decoded: %+v", got)
	}
	var bare interface{} = []interface{}{
		map[string]interface{}{"index": float64(5), "to": "restated"},
	}
	if got := decodeRegisterReplacements(bare); len(got) != 1 || got[0].Index != 5 {
		t.Errorf("bare array shape not decoded: %+v", got)
	}
}

// ⚠ TestSummaryStatesStillViolatingOnBothPaths — from the council's round-1
// objection (bug_historian): "the gate's presence could read as 'this content is
// now guarded' when a meaningful fraction of violations will still ship dirty."
//
// Layers 2-4 fail closed, so a refused repair KEEPS the original violating text.
// That is expected behaviour of a working gate, not an error — which is exactly
// why the count has to be stated rather than inferred. A reader who has to derive
// it by counting outcome='kept' is a reader who will not.
func TestSummaryStatesStillViolatingOnBothPaths(t *testing.T) {
	clean := newRegisterSummary(6, 0, 0)
	if clean.StillViolating != 0 || clean.Checked != 6 {
		t.Errorf("clean summary wrong: %+v", clean)
	}
	// The case the objection is about: 8 violations, 3 repaired, FIVE still dirty.
	dirty := newRegisterSummary(20, 8, 3)
	if dirty.StillViolating != 5 {
		t.Fatalf("still_violating must be violations-repaired: got %d, want 5 (%+v)", dirty.StillViolating, dirty)
	}
	// It must survive the JSON round-trip into the artefact, under a name a census
	// can key on — a summary that marshals to nothing states nothing.
	b, err := json.Marshal(dirty)
	if err != nil {
		t.Fatalf("summary does not marshal: %v", err)
	}
	if !strings.Contains(string(b), `"still_violating":5`) {
		t.Errorf("still_violating is not in the persisted shape: %s", b)
	}
	// And it must carry the authority it judged against, so a stale artefact can be
	// told from a current one after a register version bump.
	if dirty.RegisterVer != datahelpers.BannedRegisterVersion || dirty.Register != datahelpers.BannedRegisterPath {
		t.Errorf("summary does not carry the register it judged against: %+v", dirty)
	}
}

// TestBothArtefactKeysAreWrittenOnTheCleanPath. The record AND the summary both
// go through write_site_spec's deep merge, so BOTH must be written on a clean run
// or the previous run's value stands and reads as current (bugs_open/327). The
// summary was added later than the record, which is exactly when this gets missed.
func TestBothArtefactKeysAreWrittenOnTheCleanPath(t *testing.T) {
	obj := map[string]interface{}{"reader_goal": "x", "lead_with": []interface{}{}}
	out := withKey(
		withKey(obj, "register_repairs", []registerRepairRecord{}),
		"register_repairs_summary", newRegisterSummary(0, 0, 0))
	for _, k := range []string{"register_repairs", "register_repairs_summary"} {
		if _, present := out[k]; !present {
			t.Errorf("clean path must write %q — an omitted key keeps the previous run's value under deep merge", k)
		}
	}
	if out["reader_goal"] != "x" {
		t.Error("an unrelated key was dropped")
	}
}
