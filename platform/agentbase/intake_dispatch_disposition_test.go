// FILE: platform/agentbase/intake_dispatch_disposition_test.go
package agentbase

import (
	"fmt"
	"testing"

	perrors "github.com/gqls/agentchassis/platform/errors"
)

// intakeEventDisposition is the whole of bugs_open/239's change at this layer,
// and both wrong answers are expensive in opposite directions: calling a
// transient fault `done` loses the message for ever, and calling a terminal
// refusal `retry` replays a message that can never succeed until its attempts
// run out.
//
// Before the fix there was no decision at all — MarkEventDone was
// unconditional, so a dispatch that resolved nothing and ran nothing was
// recorded as an event that had been handled.
func TestIntakeEventDisposition(t *testing.T) {
	terminal := perrors.New(perrors.ErrDispatchUnresolvable, "DISPATCH_FAIL_CLOSED parse_failure").Build()
	transient := perrors.New(perrors.ErrDispatchLookupUnavailable, "DISPATCH_LOOKUP_RETRYABLE").
		AsRetryable(nil).Build()

	tests := []struct {
		name string
		err  error
		want string
	}{
		{"success", nil, intakeDispositionDone},
		{"terminal dispatch refusal", terminal, intakeDispositionFailed},
		{"transient lookup fault", transient, intakeDispositionRetry},
		// Everything else keeps the pre-fix contract exactly: the agent RAN and
		// the failure is the parent's to retry, so the intake row is done.
		{"ordinary handler error", fmt.Errorf("step execute_llm_prompt failed"), intakeDispositionDone},
		{"untyped error mentioning dispatch", fmt.Errorf("dispatch went wrong"), intakeDispositionDone},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := intakeEventDisposition(tc.err); got != tc.want {
				t.Fatalf("disposition = %q, want %q", got, tc.want)
			}
		})
	}
}

// TestIntakeEventDisposition_SurvivesWrapping: the classification keys on the
// typed code via errors.As, so a %w wrap on the way up cannot silently change
// an event's fate. A bare type assertion would answer `done` here, which is the
// defect bugs_open/195 fixed one layer down and this must not reintroduce.
func TestIntakeEventDisposition_SurvivesWrapping(t *testing.T) {
	terminal := perrors.New(perrors.ErrDispatchUnresolvable, "DISPATCH_FAIL_CLOSED parse_failure").Build()
	wrapped := fmt.Errorf("processing message: %w", error(terminal))

	if got := intakeEventDisposition(wrapped); got != intakeDispositionFailed {
		t.Fatalf("wrapped terminal refusal → %q, want %q", got, intakeDispositionFailed)
	}
}
