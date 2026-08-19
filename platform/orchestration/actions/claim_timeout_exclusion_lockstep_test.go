// FILE: platform/orchestration/actions/claim_timeout_exclusion_lockstep_test.go
//
// The claimed-item-timeout lockstep guard, WIDENED to cover both completion gates
// (bugs_open/317). It replaces TestRegisteredVerifiersMatchClaimTimeoutExclusion,
// which lived in discovery_checks and could only ever see one of them.
//
// WHY IT MOVED PACKAGE, which is the whole reason 317 existed. The sweep
// `claimed-item-timeout` auto-completes a claimed item past its timeout by writing
// the row directly, so NEITHER completion gate runs. Its only protection is an
// item_type exclusion list, and migration 220's own comment states the contract it
// was written to: "the LOCKSTEP TWIN of the RegisterVerifier() calls". That was
// complete when gate 2 was the only gate. Gate 1b (noChangeGates) arrived on
// 2026-08-13 with its own opt-in roster, and a type on THAT roster with no
// registered verifier fell outside the list — so for it the sweep was a completion
// path no gate could see.
//
// The old guard could not be widened where it stood: noChangeGates lives in package
// `actions`, and `actions` imports `discovery_checks`, so the test could not read
// both rosters from there. It reads both from here.
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
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"

	checks "github.com/gqls/agentchassis/platform/orchestration/actions/discovery_checks"
)

// claimTimeoutMigrationGlob locates the migrations that own the sweep's predicate.
//
// It deliberately matches MORE THAN ONE file and takes the newest. 220 is applied
// history and stays untouched — editing it would make its recorded checksum a lie —
// so a later migration owns the live predicate, and the guard must read whichever
// that is rather than the first one ever written.
const claimTimeoutMigrationGlob = "../../../docs/agent_docs/sql_for_agents/*_claimed_item_timeout_generic_evidence.sql"

var claimTimeoutExclusionRe = regexp.MustCompile(`item_type NOT IN \(([^)]*)\)`)

// itemTypeShapeRe is what a real item_type looks like. Anything else in the parsed
// list means the regex matched prose rather than the declaration.
var itemTypeShapeRe = regexp.MustCompile(`^[a-z][a-z0-9_]*$`)

func TestClaimTimeoutExclusionCoversBothCompletionGates(t *testing.T) {
	matches, err := filepath.Glob(claimTimeoutMigrationGlob)
	if err != nil {
		t.Fatalf("glob %s: %v", claimTimeoutMigrationGlob, err)
	}
	if len(matches) == 0 {
		t.Fatalf("no claim-timeout migration matches %s.\n"+
			"If it was renamed, fix claimTimeoutMigrationGlob — this guard is inert until you do.",
			claimTimeoutMigrationGlob)
	}
	// Lexicographic order is numeric order here: every file in sql_for_agents carries a
	// zero-padded 3-digit prefix. If that ever stops being true this picks the wrong file,
	// so the chosen name is reported on failure rather than left implicit.
	sort.Strings(matches)
	newest := matches[len(matches)-1]

	src, err := os.ReadFile(newest)
	if err != nil {
		t.Fatalf("read %s: %v", newest, err)
	}

	found := claimTimeoutExclusionRe.FindSubmatch(src)
	if found == nil {
		t.Fatalf("no exclusion clause found in %s.\n"+
			"Either the sweep no longer excludes gated item types — in which case it can now\n"+
			"auto-complete an item a gate would have blocked — or the SQL was reshaped and this\n"+
			"guard went vacuous. Both need a human.", newest)
	}

	excluded := map[string]bool{}
	for _, raw := range strings.Split(string(found[1]), ",") {
		if itemType := strings.Trim(strings.TrimSpace(raw), "'"); itemType != "" {
			excluded[itemType] = true
		}
	}

	// WELL-FORMEDNESS, and it guards this guard against the trap that nearly bit the
	// migration it reads (council editquality, gating objection, corr ff58ee4a).
	//
	// The regex takes the FIRST match in the file, so ANY prose above the real
	// declaration that spells the exclusion clause out is parsed instead of it — and
	// 482's own explanatory comment did exactly that while being written. The failure
	// is not silent (every gated type then reports as "NOT excluded"), but it reports
	// 14 confusing errors about the roster rather than the one true cause, which is
	// that the parse is looking at prose. Assert the shape so the real cause is named.
	for itemType := range excluded {
		if !itemTypeShapeRe.MatchString(itemType) {
			t.Fatalf("the exclusion clause in %s parsed to %q, which is not an item_type.\n"+
				"The regex takes the FIRST `item_type NOT IN (...)` in the file, so a COMMENT above the real\n"+
				"declaration that spells the clause out is read instead of it. Describe the clause in prose;\n"+
				"never spell it. (This is how 482 was authored — the trap is real and this is its check.)",
				filepath.Base(newest), itemType)
		}
	}

	// gated = the union of the two rosters. Both are read live from the code that
	// enforces them, never from a third copy here: a guard whose own copy can drift is
	// not a guard.
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
			"back to the gate-2-only contract that bugs_open/317 was filed about")
	}

	for itemType, why := range gated {
		if !excluded[itemType] {
			t.Errorf("item_type %q has %s but is NOT excluded in %s.\n"+
				"The claimed-item-timeout sweep writes the row directly, so it will auto-complete this\n"+
				"item on handler-orchestration evidence alone, with that gate never running\n"+
				"(bugs_open/317, /017, /021). Add %q to the exclusion clause in a NEW migration and apply it.",
				itemType, why, filepath.Base(newest), itemType)
		}
	}

	for itemType := range excluded {
		if _, isGated := gated[itemType]; !isGated {
			t.Errorf("item_type %q is excluded from the claim-timeout sweep in %s but NO gate can grade it.\n"+
				"Nothing can ever prove its completion, so it falls through to the timeout reset forever —\n"+
				"the churn bugs_open/006 §C was filed about. Remove it from the exclusion clause, or give it\n"+
				"the verifier or noChangeGates entry its exclusion implies.",
				itemType, filepath.Base(newest))
		}
	}
}
