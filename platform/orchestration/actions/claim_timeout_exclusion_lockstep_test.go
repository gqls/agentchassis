// FILE: platform/orchestration/actions/claim_timeout_exclusion_lockstep_test.go
//
// The claimed-item-timeout lockstep guard, covering BOTH completion gates
// (bugs_closed/317), now asserting against platform/livespec rather than against a
// migration file (bugs_open/363, council b3676918 APPROVED round 2).
//
// WHY IT MOVED PACKAGE (bugs_closed/317, unchanged). The sweep
// `claimed-item-timeout` auto-completes a claimed item past its timeout by writing
// the row directly, so NEITHER completion gate runs. Its only protection is an
// item_type exclusion list. Gate 1b (`noChangeGates`) lives in package `actions`,
// and `actions` imports `discovery_checks`, so a test in `discovery_checks` could
// never read both rosters. It reads both from here.
//
// WHY IT NO LONGER READS A MIGRATION (bugs_open/363, new). It used to glob
// `*_claimed_item_timeout_generic_evidence.sql`, take the newest match and parse an
// `item_type NOT IN (...)` clause out of it. Three things were wrong with that, and
// the first two are structural:
//
//  1. A migration is APPEND-ONLY HISTORY. `schema_migrations` records a checksum of
//     the file, so editing an applied migration makes that record a lie. The file is
//     frozen; the live object is not. Asserting the file's text is an assertion that
//     cannot fail in the direction that matters.
//  2. The glob could not see the edits that actually happened. Migrations 322, 331
//     and 374 each amended this live clause with `SET pre_query = replace(...)`, and
//     none of their filenames match that pattern; 524 edited the same column again.
//     Widening the glob is guessing at a naming convention.
//  3. Because the migration applies a `replace()` of a tail fragment, the applied
//     SQL never contains the whole clause — so migration 482 had to spell the list
//     out in a PROSE COMMENT for this test to parse. The declaration was a comment.
//
// The declaration now lives in `platform/livespec`, in a file that is allowed to
// change. ⚠ What that does NOT yet give us: nothing compares livespec to the LIVE
// `scheduled_tasks.pre_query`. That is the phase-2 auditor, and until it ships this
// guard proves Go and the declaration agree, not that either matches production.
//
// THE CONTRACT, both directions:
//
//	excluded  ⇔  (has a registered verifier)  OR  (has a noChangeGates entry)
//
// Forward: a type either gate can block must be excluded, or the sweep completes
// past the gate. Reverse: an excluded type no gate can grade would fall through to
// the timeout reset forever — the churn bugs_open/006 §C was filed about — so an
// exclusion must be earned by a gate that exists.
package actions

import (
	"strings"
	"testing"

	"github.com/gqls/agentchassis/platform/livespec"
	checks "github.com/gqls/agentchassis/platform/orchestration/actions/discovery_checks"
)

func TestClaimTimeoutExclusionCoversBothCompletionGates(t *testing.T) {
	excluded := map[string]bool{}
	for _, itemType := range livespec.ClaimedItemTimeoutExclusions {
		excluded[itemType] = true
	}
	if len(excluded) != len(livespec.ClaimedItemTimeoutExclusions) {
		t.Fatalf("the declared exclusion list contains a duplicate (%d entries, %d distinct)",
			len(livespec.ClaimedItemTimeoutExclusions), len(excluded))
	}

	// gated = the union of the two rosters. Both are read live from the code that
	// enforces them, never from a third copy here: a guard whose own copy can drift
	// is not a guard.
	gated := map[string]string{}
	for _, itemType := range checks.RegisteredVerifierItemTypes() {
		gated[itemType] = "a registered verifier (gate 2)"
	}
	for itemType := range noChangeGates {
		if existing, dup := gated[itemType]; dup {
			gated[itemType] = existing + " and a noChangeGates entry (gate 1b)"
		} else {
			gated[itemType] = "a noChangeGates entry (gate 1b)"
		}
	}

	// Both halves must be non-empty or the comparison proves nothing. Asserted
	// separately so a failure says WHICH roster went missing.
	if len(checks.RegisteredVerifierItemTypes()) == 0 {
		t.Fatal("zero verifiers registered — init() ordering broke or the registry moved; " +
			"this guard would be comparing against a half-empty set and proving less than it appears to")
	}
	if len(noChangeGates) == 0 {
		t.Fatal("noChangeGates is empty — gate 1b is inert; this guard would silently narrow " +
			"back to the gate-2-only contract that bugs_closed/317 was filed about")
	}

	for itemType, why := range gated {
		if !excluded[itemType] {
			t.Errorf("item_type %q has %s but is NOT declared in livespec.ClaimedItemTimeoutExclusions.\n"+
				"The claimed-item-timeout sweep writes the row directly, so it will auto-complete this\n"+
				"item on handler-orchestration evidence alone, with that gate never running\n"+
				"(bugs_closed/317, bugs_open/017, /021). Add %q to the declaration AND ship a migration\n"+
				"amending the live pre_query — the declaration alone changes nothing in production.",
				itemType, why, itemType)
		}
	}

	for itemType := range excluded {
		if _, isGated := gated[itemType]; !isGated {
			t.Errorf("item_type %q is declared excluded from the claim-timeout sweep but NO gate can grade it.\n"+
				"Nothing can ever prove its completion, so it falls through to the timeout reset forever —\n"+
				"the churn bugs_open/006 §C was filed about. Remove it from the declaration, or give it\n"+
				"the verifier or noChangeGates entry its exclusion implies.", itemType)
		}
	}
}

// TestCooldownRendererMatchesTheDeclaration pins the Go renderer to the declared
// predicate. It lives here because workItemRetryNotPendingSQL is an `actions`
// symbol that livespec's own tests cannot reach.
//
// Fragment containment rather than exact equality (council: debug_historian): an
// exact compare on rendered SQL turns a harmless whitespace change into a red test
// that costs a reviewer an hour, while the drift that actually matters is the
// BOUNDARY.
func TestCooldownRendererMatchesTheDeclaration(t *testing.T) {
	got := workItemRetryNotPendingSQL("wi")
	if !strings.Contains(got, "retry_after <= NOW()") {
		t.Errorf("renderer produced %q, which does not carry the non-strict boundary the declaration requires (%q).\n"+
			"With a strict '<' an item is claimable a moment before it is completable, and the disagreement\n"+
			"only ever surfaces as a race.", got, livespec.WorkItemRetryNotPendingAliased)
	}
	if got != livespec.WorkItemRetryNotPendingAliased {
		t.Logf("note: renderer %q differs cosmetically from the declaration %q — not a failure, but if the\n"+
			"difference is semantic the phase-2 auditor will report it against the live row.",
			got, livespec.WorkItemRetryNotPendingAliased)
	}
}
