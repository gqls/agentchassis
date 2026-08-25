// FILE: platform/orchestration/actions/write_audit_findings_field_capability_test.go
//
// Tests for routing rule 3b (bugs_open/395 §9, bugs_open/320 §5).
//
// Every test here is written so that DELETING the thing it guards makes it FAIL.
// A test that passes whether or not the mechanism is present is the failure mode
// this estate keeps paying for, so each one names its own mutation.
package actions

import (
	"strings"
	"testing"

	"github.com/google/uuid"
)

// theWorkedCase is bugs_open/395 §1, as the router actually receives it: an
// offer-analysis content finding about webdesign.co.uk's index meta description,
// carrying the predicate CLM-024 emitted alongside its prose.
func theWorkedCase() auditFinding {
	return auditFinding{
		Category:    "content",
		Page:        "index",
		Severity:    "high",
		Description: "the index meta description leads with a catalogue count",
		AcceptanceTest: "The index page meta description mentions both the tool-article pairing and " +
			"the no-account promise, in that order, before any catalogue count.",
		AcceptancePredicate: map[string]interface{}{
			"type":   "text_order",
			"page":   "index",
			"field":  "meta_description",
			"before": []interface{}{"paired", "pairing", "article", "guide"},
			"after":  []interface{}{"$cardinal"},
		},
	}
}

func routeIt(f auditFinding) classifiedFinding {
	siteID := uuid.New()
	pages := map[string]pageInfo{"index": {ID: uuid.New(), Name: "index"}}
	return classifyFinding(f, pages, siteID, "offer-analysis")
}

// TestTheWorkedCaseNoLongerRoutesAtAHandlerThatCannotFixIt is the bug itself.
//
// MUTATION: delete the withUnwritableFieldGuard call in classifyFinding (or
// empty the pageFieldWriters entry) and this fails — the finding routes at
// page-build-handler exactly as it did on 2026-08-24 and on 2026-08-15.
func TestTheWorkedCaseNoLongerRoutesAtAHandlerThatCannotFixIt(t *testing.T) {
	c := routeIt(theWorkedCase())

	if c.HandlerAgent != "" {
		t.Fatalf("routed at handler %q, which cannot write pages.meta_description — this is bugs_open/395 §1 "+
			"and bugs_open/320 §5: the handler rebuilds, deploys, reports success and the column is "+
			"unchanged. Expected an empty handler_agent (capability_gap).", c.HandlerAgent)
	}
	if c.ItemType != "capability_gap" {
		t.Errorf("item_type = %q, want capability_gap", c.ItemType)
	}
	if c.Status != "deferred" {
		t.Errorf("status = %q, want 'deferred' — a promotable row dispatches impossible work", c.Status)
	}
	if got := c.Spec["unwritable_field"]; got != "meta_description" {
		t.Errorf("spec.unwritable_field = %v, want meta_description", got)
	}
	if got, _ := c.Spec["would_have_routed_at"].(string); got != "page-build-handler" {
		t.Errorf("spec.would_have_routed_at = %q, want page-build-handler — the row must record WHICH "+
			"handler was incapable, or whoever builds the capability cannot tell what to widen", got)
	}
	// The evidence must travel with the row, not live only in this package.
	if cap, _ := c.Spec["capability"].(string); !strings.Contains(cap, "save_page_meta_description") {
		t.Errorf("spec.capability does not carry the writer census; a reader cannot check the claim: %q", cap)
	}
}

// TestTheGuardReadsTheRejectedPredicateWrapperShape pins the trap that would
// make this guard silently inert.
//
// A rejected predicate is an AcceptancePredicateRejection — {verdict, reason,
// predicate:{…}} — so `field` is ONE LEVEL DOWN. Reading ["field"] off the
// rejection finds nothing, returns "", and the guard declines to fire on exactly
// the population it exists for, with no error and no log line.
//
// MUTATION: drop the ["predicate"] unwrapping in predicateTargetField and this
// fails while every other test in this file still passes.
func TestTheGuardReadsTheRejectedPredicateWrapperShape(t *testing.T) {
	f := theWorkedCase()
	// The emit gate refused it and moved it, preserving the predicate.
	f.AcceptancePredicate = nil
	f.AcceptancePredicateRejected = map[string]interface{}{
		"verdict": "holds",
		"reason":  "the condition already holds, so it asserts nothing about the defect",
		"predicate": map[string]interface{}{
			"type": "text_order", "page": "index", "field": "meta_description",
		},
	}

	if got := predicateTargetField(f); got != "meta_description" {
		t.Fatalf("predicateTargetField = %q, want meta_description. The rejected record NESTS the "+
			"predicate under 'predicate' (verify_acceptance_predicates_action.go:277-281); reading "+
			"['field'] off the wrapper silently yields '' and this guard stops firing.", got)
	}
	if c := routeIt(f); c.ItemType != "capability_gap" {
		t.Errorf("a finding whose REJECTED predicate names an unwritable field still routed at %q — "+
			"the two guards would blind each other", c.HandlerAgent)
	}
}

// TestAProseOnlyFindingIsUntouched is the negative control, and it is not
// optional: a guard that converts everything is indistinguishable from a guard
// that works. A prose acceptance_test is undecidable here and must route exactly
// as it did before.
//
// MUTATION: make predicateTargetField return a constant field and this fails.
func TestAProseOnlyFindingIsUntouched(t *testing.T) {
	f := theWorkedCase()
	f.AcceptancePredicate = nil
	f.AcceptancePredicateRejected = nil

	c := routeIt(f)
	if c.ItemType == "capability_gap" {
		t.Fatalf("a prose-only finding was converted to capability_gap. Prose is undecidable at this " +
			"seam by design — converting it would park work the estate CAN do.")
	}
	if c.HandlerAgent == "" {
		t.Errorf("prose-only finding lost its handler (%+v)", c)
	}
}

// TestAWritableFieldStillRoutes proves the roster discriminates rather than
// refusing everything that carries a predicate at all.
//
// MUTATION: make HandlerCanWriteField always return canWrite=false and this fails.
func TestAWritableFieldStillRoutes(t *testing.T) {
	f := theWorkedCase()
	// A field nobody has measured: not this rule's business, must fall through.
	f.AcceptancePredicate = map[string]interface{}{
		"type": "text_present", "page": "index", "field": "some_unmeasured_field",
	}

	c := routeIt(f)
	if c.ItemType == "capability_gap" {
		t.Fatalf("an UNMEASURED field was treated as unwritable. Absent from the roster means 'not " +
			"measured', which no caller may read as either capability or incapacity.")
	}
	if _, known, _ := HandlerCanWriteField("page-build-handler", "some_unmeasured_field"); known {
		t.Errorf("HandlerCanWriteField reported known=true for a field not in the roster")
	}
}

// TestPageFieldWritersCoversThePredicateVocabulary is the lockstep that stops
// this guard going quietly partial.
//
// The roster only guards fields it lists. The predicate vocabulary lives one
// file away in acceptancePredicateFields, and it WILL widen — the moment it
// admits a third field, a finding naming that field routes unguarded, and the
// guard reports success on a population it no longer covers. That is the
// "your own action narrows what your detector can see" shape, so the two are
// pinned to each other rather than left to a reader's memory.
func TestPageFieldWritersCoversThePredicateVocabulary(t *testing.T) {
	// Read the evaluator's OWN set rather than restating it. A mirrored literal
	// would agree with itself for ever and pass on the day the vocabulary
	// widened, which is the one day this test exists for.
	vocab := acceptancePredicateTextFields
	if len(vocab) == 0 {
		t.Fatal("acceptancePredicateTextFields is empty — this test would pass vacuously")
	}

	for f, readable := range vocab {
		if !readable {
			continue
		}
		if _, ok := pageFieldWriters[f]; !ok {
			t.Errorf("field %q is in the predicate vocabulary but has no pageFieldWriters entry, so a "+
				"finding naming it routes UNGUARDED. Measure its writers and add it (or state that any "+
				"handler may write it).", f)
		}
	}
	for f := range pageFieldWriters {
		if !vocab[f] {
			t.Errorf("pageFieldWriters has an entry for %q which no predicate can name — dead roster "+
				"weight that reads as coverage", f)
		}
	}
}

// TestPageFieldWritersEntriesCarryTheirEvidence — a NEGATIVE capability claim
// with no measurement and no date is an assertion, and this estate has a ruling
// about that (2026-08-22: a count of things carries the date it was counted).
// An entry that cannot be re-checked cannot be trusted to have stayed true.
func TestPageFieldWritersEntriesCarryTheirEvidence(t *testing.T) {
	if len(pageFieldWriters) == 0 {
		t.Fatal("roster is empty — this test would pass vacuously")
	}
	for field, rule := range pageFieldWriters {
		if strings.TrimSpace(rule.Why) == "" {
			t.Errorf("%q: no Why. A reviewer cannot check an incapacity claim without the writer census", field)
		}
		if !strings.Contains(rule.Why, "[MEASURED") {
			t.Errorf("%q: Why does not carry a [MEASURED …] marker, so an inference reads as a finding", field)
		}
		if strings.TrimSpace(rule.Measured) == "" {
			t.Errorf("%q: no Measured date. A negative capability claim goes stale BY ADDITION and the "+
				"date is what makes `git log --since` able to detect it", field)
		}
	}
}
