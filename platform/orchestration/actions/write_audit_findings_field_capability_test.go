// FILE: platform/orchestration/actions/write_audit_findings_field_capability_test.go
//
// Tests for routing rule 3b (bugs_open/395 §9, bugs_open/320 §5).
//
// Every test here is written so that DELETING the thing it guards makes it FAIL.
// A test that passes whether or not the mechanism is present is the failure mode
// this estate keeps paying for, so each one names its own mutation.
package actions

import (
	"os"
	"path/filepath"
	"slices"
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

// TestNoExistingRouteIsIncorrectlyOverridden answers the council's `guidelines`
// seat (021cb965): the guard wraps the AGGREGATE result of a router that decides
// in several places, and the suite only covered one worked case plus a prose
// control — not the full route matrix. In particular, a route that had ALREADY
// parked the finding (capability_gap / empty handler) must not be re-stamped.
//
// The assertion is total over the category universe and is stated as a
// BICONDITIONAL, so it cannot be satisfied by a guard that fires too little OR
// too much: a route that had a handler must become capability_gap; a route that
// had none must come back byte-identical.
func TestNoExistingRouteIsIncorrectlyOverridden(t *testing.T) {
	siteID := uuid.New()
	pages := map[string]pageInfo{"index": {ID: uuid.New(), Name: "index"}}

	withPred := func(f auditFinding) auditFinding {
		f.AcceptancePredicate = map[string]interface{}{
			"type": "text_present", "page": "index", "field": "meta_description",
		}
		return f
	}

	checkedParked, checkedRouted := 0, 0
	for _, cat := range classifyCategoryUniverse() {
		for _, page := range []string{"index", "pricing", "site-wide", ""} {
			base := auditFinding{Category: cat, Page: page, Severity: "medium", Description: "x"}

			// The route as it stands today, with no predicate attached.
			before := classifyFindingRoute(base, pages, siteID, "offer-analysis")
			// The same finding, now naming an unwritable field.
			after := classifyFinding(withPred(base), pages, siteID, "offer-analysis")

			if strings.TrimSpace(before.HandlerAgent) == "" {
				// Already parked / unrouted. The guard must be a no-op here —
				// re-stamping someone else's capability_gap would destroy its
				// gap_kind, its dedup key and its builder_needed.
				checkedParked++
				if after.ItemType != before.ItemType || after.DedupKey != before.DedupKey {
					t.Errorf("category %q page %q was ALREADY unrouted (item_type %q, dedup %q) and the "+
						"guard re-stamped it as %q / %q — it must return an unrouted finding untouched",
						cat, page, before.ItemType, before.DedupKey, after.ItemType, after.DedupKey)
				}
				continue
			}

			checkedRouted++
			if after.ItemType != "capability_gap" || after.HandlerAgent != "" {
				t.Errorf("category %q page %q routes at %q, which cannot write meta_description, and the "+
					"guard let it through as item_type %q handler %q",
					cat, page, before.HandlerAgent, after.ItemType, after.HandlerAgent)
			}
		}
	}

	// Both arms must have been exercised, or this passes vacuously — the shape
	// that made a sibling test in this package assert nothing for weeks.
	if checkedParked == 0 {
		t.Error("no already-unrouted route was exercised: the no-op arm is untested")
	}
	if checkedRouted == 0 {
		t.Error("no routed category was exercised: the conversion arm is untested")
	}
	t.Logf("route matrix covered: %d already-unrouted, %d routed", checkedParked, checkedRouted)
}

// TestThePageFieldWriterRosterIsDefinedExactlyOnce answers the council's
// `reuse_agent` and `architecture` seats (021cb965): the helper is a shared
// contract with an agreed second consumer in another lane, and the coordination
// is a chat message. A chat message cannot stop a concurrent lane defining its
// own copy of the map — which is the DRIFT THIS HELPER EXISTS TO PREVENT, so
// resting on it would reproduce the founding incident one level down.
//
// So the single-source rule is a build failure rather than a request in a doc
// comment.
//
// ⚠ The needle is split so that it cannot match the text of THIS test, and the
// assertion is on the COUNT — a source-scan whose needle matches its own prose
// passes vacuously (LANDMINES: "a source-scanning test makes your COMMENTS
// load-bearing").
func TestThePageFieldWriterRosterIsDefinedExactlyOnce(t *testing.T) {
	files, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatal(err)
	}
	needle := "pageFieldWriters" + " = map["
	found, where := 0, []string{}
	for _, f := range files {
		if strings.HasSuffix(f, "_test.go") {
			continue
		}
		src, err := os.ReadFile(f)
		if err != nil {
			t.Fatal(err)
		}
		if n := strings.Count(string(src), needle); n > 0 {
			found += n
			where = append(where, f)
		}
	}
	if found != 1 {
		t.Fatalf("pageFieldWriters is declared %d time(s) in %v, want exactly 1. A second roster is two "+
			"hand-maintained answers to 'can this handler write this field', which drift silently because "+
			"each side looks internally correct. Call HandlerCanWriteField instead of copying the map.",
			found, where)
	}
}

// TestPageFieldWritersIsTotalOverTheRoutableHandlers — bugs_open/395 §9, and the
// test that would have caught it on day one.
//
// The roster makes a NEGATIVE capability claim. Before 2026-08-31 an absent
// handler read as "cannot write", so a handler nobody had CONSIDERED was
// indistinguishable from one measured and found incapable — and that is exactly
// how the `title` entry shipped claiming no handler could write it while
// content-gap-planner could, and does, 989 times.
//
// The three tests that already existed did not catch it and could not have: the
// vocabulary lockstep, the [MEASURED marker assertion and the Measured date
// assertion are all SHAPE tests. They passed identically over the true entry and
// the false one. A marker proves a measurement was CLAIMED, never that it was
// COMPLETE — so the only thing that helps is asserting the map is TOTAL over the
// set it claims about.
func TestPageFieldWritersIsTotalOverTheRoutableHandlers(t *testing.T) {
	if len(routableHandlers) == 0 {
		t.Fatal("routableHandlers is empty — this test would pass vacuously")
	}
	if len(pageFieldWriters) == 0 {
		t.Fatal("pageFieldWriters is empty — this test would pass vacuously")
	}
	for field, rule := range pageFieldWriters {
		for _, h := range routableHandlers {
			if _, ok := rule.WritableBy[h]; !ok {
				t.Errorf("pageFieldWriters[%q].WritableBy has no verdict for routable handler %q. "+
					"Absence is NOT a measurement: measure whether that handler can write the column "+
					"(through its own spawn closure, resolved via the action registry — never a "+
					"config-text search) and record an explicit true/false with its date in Why.", field, h)
			}
		}
		// The other direction: a verdict for a handler the router can never name
		// is dead weight that reads as coverage.
		for h := range rule.WritableBy {
			if !slices.Contains(routableHandlers, h) {
				t.Errorf("pageFieldWriters[%q].WritableBy names %q, which classifyFindingRoute cannot "+
					"produce — either the router lost a route or this entry is stale", field, h)
			}
		}
	}
}

// TestHandlerCanWriteFieldReportsAnUnmeasuredHandlerAsUNKNOWN pins the safety arm
// directly, because the totality test above can only fail when someone forgets —
// it cannot prove what happens WHEN they forget. Mutation check: delete the
// `if !measured` block in HandlerCanWriteField and this fails while the totality
// test still passes, which is precisely the window bugs_open/395 §9 lived in.
func TestHandlerCanWriteFieldReportsAnUnmeasuredHandlerAsUNKNOWN(t *testing.T) {
	const unmeasured = "zzz-handler-that-was-never-considered"
	for field := range pageFieldWriters {
		canWrite, known, _ := HandlerCanWriteField(unmeasured, field)
		if known {
			t.Errorf("field %q: an unmeasured handler reported known=true — the guard would treat "+
				"'nobody thought about this' as 'proven incapable' and park a finding that may be fixable", field)
		}
		if canWrite {
			t.Errorf("field %q: an unmeasured handler reported canWrite=true", field)
		}
	}
	// And the measured negative must still be reported as MEASURED, or the arm
	// above has simply disabled the roster.
	if canWrite, known, _ := HandlerCanWriteField("page-build-handler", "meta_description"); !known || canWrite {
		t.Errorf("page-build-handler/meta_description: want known=true canWrite=false, got known=%v canWrite=%v "+
			"— the unmeasured arm must not swallow a real measurement", known, canWrite)
	}
	// The corrected entry, pinned so a revert is loud (bugs_open/395 §9).
	if canWrite, known, _ := HandlerCanWriteField("content-gap-planner", "title"); !known || !canWrite {
		t.Errorf("content-gap-planner/title: want known=true canWrite=TRUE, got known=%v canWrite=%v — "+
			"it reaches apply_gap_plan, whose applyExistingPage runs a bare UPDATE pages SET title", known, canWrite)
	}
}

// TestRoutableHandlersMatchesTheRouter is what makes routableHandlers a
// MEASUREMENT rather than a list somebody once wrote down. It scans
// classifyFindingRoute's own file for the handler names the router can emit and
// fails when they drift from the roster's universe — so adding a route without
// measuring its capability is a build failure, not a silent gap.
//
// Without this, the totality test above is circular: routableHandlers would be
// whatever the roster already covers, and the roster would be total over itself.
//
// ⚠⚠ THIS TEST IS TECH DEBT, AND BOTH THE `editquality` AND `guardian` SEATS WERE
// RIGHT TO SAY SO (council 76231f57). It derives ground truth by parsing another Go
// file's LITERAL FORMATTING — `HandlerAgent: "x"` on one line — so it couples the
// router's coding style to this test passing, and its own first draft over-read and
// harvested a DedupKey format string as a handler name. It is the sole guarantor that
// routableHandlers is a measurement rather than a list somebody typed, which makes a
// fragile mechanism load-bearing.
//
// It is kept because the alternative available today is worse (no check at all, which
// is what let bugs_open/395 §9 ship), and because its failure direction is loud: a
// router edit this scan cannot parse makes it UNDER-read, which fails the set-equality
// assertion rather than passing it. But the durable fix is for the router to expose its
// handler universe as a Go value that both it and this roster read — one declaration,
// no parsing — at which point this test should be DELETED, not maintained. Whoever
// touches classifyFindingRoute's routing shape next is best placed to do that.
//
// ⚠ COMMENTS AND STRUCT KEYS ARE STRIPPED FIRST. A source scan that reads its own
// prose passes vacuously (LANDMINES: "a source-scanning test makes your COMMENTS
// load-bearing"), and this file's neighbours discuss these very agent names — the
// router file's own comments name css-patch-agent, webdesign-agent and
// color-variable-fixer in prose about why a route was changed. Scanning raw text
// would have silently "passed" on a name no code path can produce.
func TestRoutableHandlersMatchesTheRouter(t *testing.T) {
	src, err := os.ReadFile("write_audit_findings_action.go")
	if err != nil {
		t.Fatal(err)
	}

	// Strip // comments line-wise. Crude but sufficient: the two patterns below
	// are code-only shapes, and a stray // inside a string literal would only
	// ever cause this test to under-read, i.e. fail loudly, never pass wrongly.
	var code strings.Builder
	for _, line := range strings.Split(string(src), "\n") {
		if i := strings.Index(line, "//"); i >= 0 {
			line = line[:i]
		}
		code.WriteString(line)
		code.WriteString("\n")
	}
	body := code.String()

	found := map[string]bool{}
	// (a) direct routes: `HandlerAgent: "x"` — skipping "" (a deliberate park).
	//
	// ⚠ BOUNDED TO THE SAME LINE, and the first draft was not: Rule 1 spells it
	// `HandlerAgent: handler` (a variable, no quote), so an unbounded search ran
	// on to the next quoted literal in the struct and harvested the DedupKey
	// format string "%s_%s_%s_%s" as a handler name. The test failed loudly, which
	// is the only reason it was caught — an over-reading scan of this shape is
	// otherwise indistinguishable from a real drift report.
	for _, line := range strings.Split(body, "\n") {
		i := strings.Index(line, "HandlerAgent:")
		if i < 0 {
			continue
		}
		seg := line[i+len("HandlerAgent:"):]
		q := strings.Index(seg, `"`)
		if q < 0 {
			continue // `HandlerAgent: handler` — the designRouting branch covers it
		}
		if e := strings.Index(seg[q+1:], `"`); e > 0 {
			found[seg[q+1:q+1+e]] = true
		}
	}
	// (b) designRouting's values, reached by Rule 1's `HandlerAgent: handler`.
	if i := strings.Index(body, "designRouting = map[string]string{"); i >= 0 {
		blk := body[i:]
		if e := strings.Index(blk, "\n}"); e > 0 {
			blk = blk[:e]
		}
		for _, line := range strings.Split(blk, "\n") {
			if c := strings.LastIndex(line, ":"); c >= 0 {
				v := strings.Trim(strings.TrimSpace(line[c+1:]), `",`)
				if v != "" && strings.Contains(v, "-") {
					found[v] = true
				}
			}
		}
	}

	if len(found) == 0 {
		t.Fatal("scanned the router and found no handler names — the scan broke, and a broken scan " +
			"here reads exactly like a clean fleet")
	}
	for h := range found {
		if !slices.Contains(routableHandlers, h) {
			t.Errorf("the router can emit handler %q but routableHandlers does not list it, so the "+
				"roster makes no claim about it and HandlerCanWriteField answers 'not measured' for "+
				"every field. Measure it and add it — bugs_open/395 §9 is what an unmeasured routable "+
				"handler costs.", h)
		}
	}
	for _, h := range routableHandlers {
		if !found[h] {
			t.Errorf("routableHandlers lists %q but the router cannot emit it — either a route was "+
				"removed (drop it here too) or this scan has stopped seeing the router's shape", h)
		}
	}
}

// TestTheDriftAuditDeferralTriggerHasNotFired encodes an OWNER RULING of
// 2026-09-02 on RFC_057 §6 Q1, and it exists because this lane's founding finding
// was that a rule living only as prose gets broken by the next producer. A
// deferral recorded in a document is exactly that shape; a deferral with a
// mechanical trigger is not.
//
// THE RULING, in plain terms. RFC_057 §4 asks for a live-drift audit: something
// that notices when a handler this roster says CANNOT write a field has since
// GAINED that capability. It is not built. The owner ruled it may be DEFERRED —
// on the reasoning that staleness fails in the SAFE direction (the roster
// over-refuses, which files a visible capability_gap row that two readers already
// consume) whereas the COMPLETENESS failure, which was silent, is the one now
// closed by TestPageFieldWritersIsTotalOverTheRoutableHandlers.
//
// THE TRIGGER. The deferral was granted for a roster of TWO fields, where the
// whole population can be re-read by hand in a minute. It does not extend to a
// roster that has grown. RFC_057 §4's own words: "one entry is a measurement, ten
// entries with no audit is a stale map with an enforcement mechanism attached to
// it."
//
// So this test fails the moment the roster gains a third field — not to block
// that change, but to put the owed work in front of the person making it, at the
// moment they are already measuring writers and have the context to build it.
// Delete this test in the same commit that ships the audit.
func TestTheDriftAuditDeferralTriggerHasNotFired(t *testing.T) {
	const deferredWhileRosterHasAtMost = 2

	if len(pageFieldWriters) > deferredWhileRosterHasAtMost {
		fields := make([]string, 0, len(pageFieldWriters))
		for f := range pageFieldWriters {
			fields = append(fields, f)
		}
		slices.Sort(fields)
		t.Fatalf(""+
			"THE DRIFT-AUDIT DEFERRAL HAS EXPIRED. The roster now holds %d fields (%v); the owner's "+
			"2026-09-02 deferral of RFC_057 §6 Q1 was granted for at most %d.\n\n"+
			"WHAT IS OWED: a live-drift audit that fails when a handler this roster records as UNABLE "+
			"to write a field has since GAINED that capability. Staleness is the one failure mode "+
			"nothing here can see — the totality test catches a handler never CONSIDERED, not a "+
			"capability ACQUIRED after the census.\n\n"+
			"HOW TO BUILD IT (RFC_057 §3, and this is the part that is easy to get wrong): resolve "+
			"each handler's reachable actions through the ACTION REGISTRY, never a config-text search "+
			"for the column name — an agent can write a column without ever naming it (upsertPage "+
			"writes meta_description and is reached via the action sync_pages_to_db; a column-name "+
			"census saw 1 of 3 writers). And never a workflow.steps walk: it misses steps nested in a "+
			"loop's sub_workflow and returns a confident zero (register WII-031). Use "+
			"jsonb_path_query_array(default_config, 'strict $.**.action'), and follow each handler's "+
			"SPAWN CLOSURE, because call_agent/spawn_agent mean a step list bounds what a handler does "+
			"itself, not what it can cause.\n\n"+
			"THEN DELETE THIS TEST in the same commit. bugs_open/395 §9, RFC_057 §4/§8.",
			len(pageFieldWriters), fields, deferredWhileRosterHasAtMost)
	}
}
