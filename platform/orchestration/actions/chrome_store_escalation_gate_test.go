// FILE: platform/orchestration/actions/chrome_store_escalation_gate_test.go
//
// bugs_open/423, council dc62975f round 1, guardian HIGH.
//
// The first cut of the fix routed a STORE refusal into bugs_open/260's
// escalation by analogy: a template that cannot execute may fail the step, so
// (the argument went) a store that refuses may too. The seat asked for the blast
// radius instead of the analogy, and the enumeration killed it — all SEVEN live
// workflows that dispatch render_site_components declare no error_step, no
// on_error and no continue_on_error, so a hard step failure has no handler
// anywhere and takes the whole orchestration down.
//
// So escalation became opt-in with the unsafe default OFF, and these tests pin
// the three properties that make that gate worth having. They drive
// escalateUnservedChromeSlot — the arm the action itself calls — rather than
// restating its logic.
package actions

import (
	"errors"
	"fmt"
	"testing"
)

// MUTATION KILLED: the gate deleted, i.e. the caller escalating every renderErr
// as the first cut did. Unset config must mean today's behaviour byte for byte,
// which is the whole claim that makes this change safe to roll to seven
// pipelines.
func TestStoreRefusalDoesNotEscalateByDefault(t *testing.T) {
	refused := &chromeStoreRefusedError{errors.New("invalid byte sequence for encoding \"UTF8\": 0x80")}

	for _, cfg := range []map[string]interface{}{
		nil,
		{},
		{"escalate_chrome_store_failure": false},
		{"escalate_chrome_store_failure": "true"}, // mistyped: must fail OPEN
		{"escalate_chrome_store_failure": 1},      // mistyped: must fail OPEN
	} {
		if escalateUnservedChromeSlot(cfg, refused) {
			t.Errorf("a store refusal escalated to a step failure with config %v — "+
				"unarmed must mean today's behaviour, and 7 workflows have no error_step", cfg)
		}
	}
}

// MUTATION KILLED: an arm that ignores its config key (returns false always),
// which would make the capability unreachable — the dead-mechanism half of the
// same trap.
func TestStoreRefusalEscalatesWhenArmed(t *testing.T) {
	refused := &chromeStoreRefusedError{errors.New("store refused")}
	if !escalateUnservedChromeSlot(map[string]interface{}{"escalate_chrome_store_failure": true}, refused) {
		t.Error("an armed step did not escalate a store refusal — the gate is unreachable")
	}
}

// MUTATION KILLED: gating the WRONG half — i.e. making bugs_open/260's existing
// execution-failure escalation opt-in too. That authority was reviewed and
// shipped; this change must not quietly withdraw it. A wrapped store refusal
// must still be recognised through fmt.Errorf's %w chain, or the gate leaks.
func TestExecutionFailureStillEscalatesUnconditionally(t *testing.T) {
	execErr := errors.New("template: executing \"footer\": range can't iterate over 3")
	if !escalateUnservedChromeSlot(map[string]interface{}{}, execErr) {
		t.Error("an execution failure stopped escalating — bugs_open/260's ruling was withdrawn by accident")
	}
	if !escalateUnservedChromeSlot(nil, execErr) {
		t.Error("an execution failure stopped escalating with nil config")
	}

	// The store refusal must stay recognisable after being wrapped again.
	wrapped := fmt.Errorf("chrome slot %q: %w", "footer",
		&chromeStoreRefusedError{errors.New("0x80")})
	if escalateUnservedChromeSlot(map[string]interface{}{}, wrapped) {
		t.Error("a WRAPPED store refusal escalated — errors.As must see through the chain, " +
			"or every future wrapping re-arms the escalation silently")
	}
}
