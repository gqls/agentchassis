package actions

// bugs_open/345 (Fable review F4): the pre-store rejection for a
// schema/template mismatch used to say only "template variables and schema
// fields do not match" — no field names — so the retry prompt's "change
// exactly what it says was wrong" was unsatisfiable, and fed retries failed
// byte-identically (item 2396218a, 3 attempts, 2026-08-24). These tests pin
// the two properties the fix rests on:
//
//  1. the message NAMES every offender, both directions, and
//  2. it is DETERMINISTIC — the enriched message now feeds candidate 2's
//     byte-identical repeat detector, so a message whose content or order
//     varied between attempts would silently defeat the mechanism it exists
//     to inform. Determinism is asserted by scoring the same inputs
//     repeatedly, which is exactly the operation a retry performs.
//
// Mutation map (each guard fails its own test and only that one):
//   - drop the sort in describeSchemaTemplateMismatch     → TestMismatchMessageIsDeterministic (order flips run to run via map input)
//   - restore the `break` in either scoring loop          → TestMismatchMessageNamesEveryOffender (a second offender goes unnamed)
//   - restore the `if synced` gate on Direction 2         → TestMismatchMessageNamesEveryOffender (orphan hidden behind a var miss)
//   - drop the join / return bare constant unconditionally→ TestMismatchMessageNamesEveryOffender

import (
	"strings"
	"testing"
)

// twoOrphansOneUnknown: schema declares alpha/beta/gamma, template renders
// only gamma and additionally references an undeclared {{.delta}}. So the
// full mismatch is: alpha and beta orphaned (Direction 2), delta unknown
// (Direction 1).
const mismatchSchema = `{"fields":{"alpha":{"type":"text"},"beta":{"type":"text"},"gamma":{"type":"text"}}}`
const mismatchTemplate = `<section data-component="x"><div>{{.gamma}} {{.delta}}</div></section>`

func TestMismatchMessageNamesEveryOffender(t *testing.T) {
	score := scoreComponent("", "x", mismatchTemplate, mismatchSchema, "section")
	if score.SchemaTemplateSynced {
		t.Fatal("fixture defect: this template/schema pair must be out of sync")
	}
	msg := describeSchemaTemplateMismatch(score)

	// The stable prefix survives (anything grepping the old constant still
	// matches), and every offender is named — including the SECOND orphan,
	// which the pre-fix `break` could never report, and the Direction-2
	// orphans, which the pre-fix `if synced` gate hid whenever Direction 1
	// also missed.
	if !strings.Contains(msg, "template variables and schema fields do not match") {
		t.Fatalf("stable prefix lost: %q", msg)
	}
	for _, want := range []string{
		`schema field "alpha" has no template variable`,
		`schema field "beta" has no template variable`,
		`template var {{.delta}} has no schema entry`,
	} {
		if !strings.Contains(msg, want) {
			t.Fatalf("offender not named: want %q in %q", want, msg)
		}
	}
	// gamma is fine on both sides and must NOT be blamed.
	if strings.Contains(msg, `"gamma"`) || strings.Contains(msg, "{{.gamma}}") {
		t.Fatalf("non-offender blamed: %q", msg)
	}
}

func TestMismatchMessageIsDeterministic(t *testing.T) {
	// Two layers, attacked separately — a first cut of this test only scored
	// real inputs, and the sort-removal mutation PASSED it: the producer's
	// sorted iteration was a guard in series, masking the helper's own
	// (the a-mutation-that-passes-may-have-hit-a-guard-in-series lesson,
	// caught in this change's own mutation round on 2026-08-24).
	//
	// Layer 1 — the HELPER alone must normalise order: hand it the same
	// issues in two different orders and require identical output. This is
	// the case a future producer that appends unsorted would create.
	fwd := ComponentQualityResult{QualityIssues: []string{
		`schema field "alpha" has no template variable`,
		`schema field "beta" has no template variable`,
		`template var {{.delta}} has no schema entry`,
	}}
	rev := ComponentQualityResult{QualityIssues: []string{
		`template var {{.delta}} has no schema entry`,
		`schema field "beta" has no template variable`,
		`schema field "alpha" has no template variable`,
	}}
	if a, b := describeSchemaTemplateMismatch(fwd), describeSchemaTemplateMismatch(rev); a != b {
		t.Fatalf("the helper's output depends on issue ORDER:\n  %q\nvs\n  %q\n"+
			"a varying message defeats byte-identical repeat termination (345 candidate 2)", a, b)
	}

	// Layer 2 — the PRODUCER must emit deterministically too (Direction 2
	// iterates a Go map; range order varies within one process).
	first := describeSchemaTemplateMismatch(scoreComponent("", "x", mismatchTemplate, mismatchSchema, "section"))
	for i := 0; i < 20; i++ {
		got := describeSchemaTemplateMismatch(scoreComponent("", "x", mismatchTemplate, mismatchSchema, "section"))
		if got != first {
			t.Fatalf("message varies between identical scoring runs (iteration %d):\n  %q\nvs\n  %q", i, first, got)
		}
	}
}

func TestMismatchMessageFallsBackToBareConstant(t *testing.T) {
	// A score with no matching per-field issues keeps the old wording —
	// no second wording is invented for an unknown shape.
	score := ComponentQualityResult{QualityIssues: []string{"something else entirely"}}
	if got := describeSchemaTemplateMismatch(score); got != "template variables and schema fields do not match" {
		t.Fatalf("unexpected message for no-detail case: %q", got)
	}
}
