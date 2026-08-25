// FILE: platform/orchestration/actions/unarmed_completer_lockstep_test.go
//
// The build-time half of bugs_open/375 candidate 4: an arm that completes a
// VERIFIED item type without consulting its verifier must be a decision on the
// record, not an accident.
//
// WHY IT LIVES IN THIS PACKAGE, and not in discovery_checks. It has to read BOTH
// the verifier registry (discovery_checks) and the declaration (livespec), and
// `actions` imports discovery_checks while the reverse would be a cycle. That is
// the same reason claim_timeout_exclusion_lockstep_test.go sits here, and this
// test is deliberately modelled on it: same shape, same two-directions contract,
// same honest statement of what a Go test cannot see.
//
// WHAT IT CATCHES, stated as the sequence it is written for. Somebody works
// verifier_coverage_test.go's catMechanical backlog — the list that says of itself
// "These SHOULD get verifiers — this is the actionable backlog, not an excuse
// list". They write a verifier for required_fields_missing, register it, and the
// coverage guard goes green. Nothing is protected: all three of that type's
// `complete` arms are update_work_item_status steps, and WII-030's consult is
// opt-in per step. Before this test, their next signal would have been a
// result._verification marker on a live item that had ALREADY completed
// unverified. Now the build stops them, at the moment they type RegisterVerifier.
//
// ⚠ WHAT IT DELIBERATELY DOES NOT DO: force the arm to be armed. Arming is a live
// behaviour change with a documented hazard per type — CQ-023 records that a
// required_fields_missing verifier fail-closes the `converted` arm — so the guard
// demands a REASON, not a switch. That mirrors itemTypesWithoutVerifiers, which
// makes an unverified type legitimate as long as somebody wrote down why.
//
// ⚠ AND WHAT IT CANNOT SEE, which is half the contract. Go cannot read
// agent_definitions, so this test cannot tell whether the declaration still matches
// production. A new agent completing through this writer is INVISIBLE here. That
// arm is cmd/config-key-audit --unarmed-verified-completers, and treating this test
// as the whole guard reproduces, one level down, the exact criticism the council
// levelled at the first cut of verifier_coverage_test.go.
package actions

import (
	"testing"

	"github.com/gqls/agentchassis/platform/livespec"
	checks "github.com/gqls/agentchassis/platform/orchestration/actions/discovery_checks"
)

func TestUnarmedCompletersOfVerifiedTypesAreAcknowledged(t *testing.T) {
	verified := map[string]bool{}
	for _, itemType := range checks.RegisteredVerifierItemTypes() {
		verified[itemType] = true
	}
	// Both halves must be non-empty or the comparison proves nothing — the same
	// guard claim_timeout_exclusion_lockstep_test.go makes, for the same reason: a
	// registry that failed to initialise would make every assertion below pass.
	if len(verified) == 0 {
		t.Fatal("zero verifiers registered — init() ordering broke or the registry moved; " +
			"every assertion in this test would pass vacuously")
	}
	if len(livespec.UnarmedVerifiedCompleters) == 0 {
		t.Fatal("livespec.UnarmedVerifiedCompleters is empty — either every completer now arms the gate " +
			"(check that before deleting this guard) or the declaration was lost")
	}

	seen := map[string]bool{}
	for _, c := range livespec.UnarmedVerifiedCompleters {
		key := c.Agent + "." + c.Step
		if seen[key] {
			t.Errorf("%s is declared twice — a duplicate makes the live-set comparison in "+
				"cmd/config-key-audit --unarmed-verified-completers unable to balance", key)
		}
		seen[key] = true

		if c.Agent == "" || c.Step == "" || c.ItemType == "" {
			t.Errorf("declaration entry %+v is incomplete — Agent, Step and ItemType are what the "+
				"live-set comparison joins on", c)
			continue
		}
		if c.Why == "" {
			t.Errorf("%s (%s) has no Why — an entry nobody can read is a list that stops being maintained",
				key, c.ItemType)
		}

		// THE LOAD-BEARING ASSERTION. The moment this type gains a verifier, this
		// arm is actively skipping a guard, and that must be a decision somebody
		// wrote down rather than a silence.
		if verified[c.ItemType] && c.Acknowledged == "" {
			t.Errorf("item_type %q HAS a registered verifier, and %s completes it through "+
				"update_work_item_status WITHOUT setting verify_before_complete — so the verifier is "+
				"consulted by nothing on this path and verifier_coverage_test.go is green anyway.\n"+
				"Either arm the step (config: verify_before_complete: true — read that type's close paths "+
				"FIRST, register CQ-023 records a fail-close hazard on required_fields_missing's converted "+
				"arm), or set Acknowledged on this entry with the reason it stays unarmed.\n"+
				"See bugs_open/375 and WII-030.", c.ItemType, key)
		}

		// The reverse direction, which is what keeps the list honest rather than
		// merely long: an Acknowledged reason on a type with no verifier is
		// asserting a bypass that cannot happen, and it will read as settled to the
		// next person deciding whether to write that verifier.
		if !verified[c.ItemType] && c.Acknowledged != "" {
			t.Errorf("%s declares Acknowledged for item_type %q, which has NO registered verifier — "+
				"there is no guard being skipped, so the acknowledgement claims a decision nobody had to "+
				"make. Move the content into Why, or register the verifier the acknowledgement implies.",
				key, c.ItemType)
		}
	}
}
