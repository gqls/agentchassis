// FILE: platform/orchestration/adapter_retry_cap_test.go
//
// The adapter-action retry cap (bugs_open/075). These tests model the LIVE
// shape of the loop that ran on 2026-07-25: each cycle re-executes the step,
// which mints a brand new awaited request, and createAwaitedRequest hardcodes
// RetryVersion: 0 — so retry_version is 0 on every single cycle and any cap
// that reads it alone can never fire.
//
// TestOldAssignmentRuleNeverCaps is a deliberate negative control: it encodes
// the rule this fix replaced and asserts it loops for ever. Without it, the
// tests below would pass against code that never worked, which is the failure
// mode a lockstep test walked into on bugs_open/076.

package orchestration

import "testing"

// The live loop: a fresh rv=0 request every cycle, the durable counter carried
// in execution_metadata across cycles (and across the pod deaths F2 recovers
// from). The cap must stop the third re-execution's successor.
func TestAdapterRetryCapFiresOnFreshRequestsPinnedAtZero(t *testing.T) {
	counts := map[string]int{}
	const step = "deploy_page"

	for cycle := 1; cycle <= maxStepRetries; cycle++ {
		attempt, capped := nextAdapterRetryAttempt(counts, step, 0)
		if capped {
			t.Fatalf("cycle %d: capped too early (attempt=%d, counter=%d)", cycle, attempt, counts[step])
		}
		if attempt != cycle {
			t.Fatalf("cycle %d: attempt = %d, want %d", cycle, attempt, cycle)
		}
		counts[step] = attempt
	}

	attempt, capped := nextAdapterRetryAttempt(counts, step, 0)
	if !capped {
		t.Fatalf("after %d re-executions the cap did not fire (attempt=%d) — this is the 2026-07-25 infinite loop",
			maxStepRetries, attempt)
	}
	if counts[step] != maxStepRetries {
		t.Fatalf("durable counter = %d, want %d", counts[step], maxStepRetries)
	}
}

// NEGATIVE CONTROL. The rule that shipped before this fix, applied to the same
// sequence: RetryCount[step] = retryVersion + 1. It is an assignment, so the
// counter is pinned at 1 for ever and no threshold on it is reachable. If this
// test ever fails, the control has stopped modelling the old defect and the
// test above is no longer evidence of anything.
func TestOldAssignmentRuleNeverCaps(t *testing.T) {
	counts := map[string]int{}
	const step = "deploy_page"
	const retryVersion = 0 // what a freshly minted adapter request always carries

	for cycle := 1; cycle <= 50; cycle++ {
		counts[step] = retryVersion + 1 // the old line, verbatim in behaviour
		if counts[step] >= maxStepRetries {
			t.Fatalf("cycle %d: the old rule reached the cap (counter=%d) — control invalid", cycle, counts[step])
		}
	}
}

// The message path keeps working: there retry_version is authoritative because
// the SAME request row is resent and incremented, and it may be ahead of the
// per-step map (which only the adapter path writes).
func TestRetryVersionLeadsWhenItIsAhead(t *testing.T) {
	counts := map[string]int{"call_agent": 1}

	attempt, capped := nextAdapterRetryAttempt(counts, "call_agent", 2)
	if capped {
		t.Fatalf("capped at retry_version=2 with cap %d", maxStepRetries)
	}
	if attempt != 3 {
		t.Fatalf("attempt = %d, want 3 (retry_version 2 leads the counter 1)", attempt)
	}

	if _, capped := nextAdapterRetryAttempt(counts, "call_agent", maxStepRetries); !capped {
		t.Fatalf("retry_version = %d did not cap", maxStepRetries)
	}
}

// A step nobody has retried starts at attempt 1, and one step's budget must not
// consume another's.
func TestRetryBudgetIsPerStep(t *testing.T) {
	counts := map[string]int{"deploy_page": maxStepRetries}

	attempt, capped := nextAdapterRetryAttempt(counts, "commit_files", 0)
	if capped || attempt != 1 {
		t.Fatalf("untouched step: attempt=%d capped=%v, want 1/false", attempt, capped)
	}

	if _, capped := nextAdapterRetryAttempt(counts, "deploy_page", 0); !capped {
		t.Fatal("exhausted step was not capped")
	}
}
