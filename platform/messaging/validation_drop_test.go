// FILE: platform/messaging/validation_drop_test.go
package messaging

import (
	"fmt"
	"testing"

	"github.com/gqls/agentchassis/platform/errors"
)

// TestMatchedValidationNeedle pins the substring classifier behind the
// dropped-validation-error branch (bugs_open/034). It documents BOTH halves:
// genuine validation errors are still classified (so retry behaviour is
// unchanged), and the unanchored match still catches non-validation runtime
// failures — which is precisely why recordDroppedValidationError now persists a
// durable row and records which needle fired.
//
// Moved here from platform/agentbase when the two layers' drifted needle lists
// were merged into one. The agentbase side keeps a lockstep test.
func TestMatchedValidationNeedle(t *testing.T) {
	tests := []struct {
		name string
		err  string
		want string
	}{
		// Genuine validation errors — behaviour preserved.
		{"client_id required", "client_id is required to execute a workflow", "is required"},
		{"generic validation", "payload failed validation", "validation"},
		{"invalid field", "field foo is invalid", "invalid"},
		// "missing required header" does NOT contain the needle "is required" —
		// the needle carries the word "is". It matches on "missing" alone, which
		// is exactly the needle agentbase used to lack.
		{"missing required header", "missing required header", "missing"},
		{"missing only", "missing client_id header", "missing"},
		// Earlier needles still win when several are present.
		{"ordering: is required before missing", "client_id is required, header missing", "is required"},

		// Unanchored-match hazard (bugs_open/034): these are NOT validation
		// errors, yet they match. The classifier still drops them; the point of
		// the fix is that the drop is now durably recorded, not that these stop
		// matching.
		{"truncated-LLM parse failure", "unrecoverable after control-char repair: invalid character 'w' after object key:value pair", "invalid"},
		{"db driver error", "pq: invalid connection", "invalid"},
		{"nil-pointer recovery", "runtime error: invalid memory address or nil pointer dereference", "invalid"},

		// Ordinary failures that must NOT be classified as validation errors.
		{"plain timeout", "context deadline exceeded", ""},
		{"kafka unavailable", "broker not available", ""},
		{"empty", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := MatchedValidationNeedle(tt.err); got != tt.want {
				t.Errorf("MatchedValidationNeedle(%q) = %q, want %q", tt.err, got, tt.want)
			}
		})
	}
}

// TestValidationNeedlesAreTheOnesBothLayersUsed guards the drift that
// bugs_open/034 turned up: agentbase classified on three needles and messaging
// on four, so an error containing only "missing" was dropped-without-retry by
// one layer and retried by the other, depending purely on which path the message
// took. There is now one list; this pins its contents so a future edit to either
// layer has to come through here.
func TestValidationNeedlesAreTheOnesBothLayersUsed(t *testing.T) {
	want := []string{"is required", "validation", "invalid", "missing"}

	if len(ValidationErrorNeedles) != len(want) {
		t.Fatalf("ValidationErrorNeedles = %v, want %v", ValidationErrorNeedles, want)
	}
	for i, w := range want {
		if ValidationErrorNeedles[i] != w {
			t.Errorf("ValidationErrorNeedles[%d] = %q, want %q", i, ValidationErrorNeedles[i], w)
		}
	}
}

// TestMatchedPermanentFailure pins the typed classifier (bugs_open/195).
//
// The load-bearing case is ReproducesTheBug: the exact error a rejected
// workflow produces returns "" from MatchedValidationNeedle — that IS the bug,
// and the test asserts it explicitly so nobody "fixes" the needle list and
// quietly removes the reason this seam exists.
func TestMatchedPermanentFailure(t *testing.T) {
	workflowInvalid := errors.New(errors.ErrWorkflowInvalid, "Invalid workflow configuration").
		WithCause(fmt.Errorf("step 'done' with action 'complete' requires a topic")).
		Build()

	t.Run("ReproducesTheBug_needleMissesButCodeMatches", func(t *testing.T) {
		// The whole of bugs_open/195 in two assertions.
		if got := MatchedValidationNeedle(workflowInvalid.Error()); got != "" {
			t.Errorf("the substring list is expected to MISS this error (that is the bug); got %q", got)
		}
		if got := MatchedPermanentFailure(workflowInvalid); got != "code:WORKFLOW_INVALID" {
			t.Errorf("MatchedPermanentFailure = %q, want code:WORKFLOW_INVALID", got)
		}
	})

	t.Run("SurvivesPercentWWrapping", func(t *testing.T) {
		// The reason for errors.As over a bare type assertion.
		wrapped := fmt.Errorf("processing message: %w", workflowInvalid)
		if got := MatchedPermanentFailure(wrapped); got != "code:WORKFLOW_INVALID" {
			t.Errorf("a %%w-wrapped DomainError must still classify; got %q", got)
		}
	})

	t.Run("DegradesToNeedleWhenChainIsBroken", func(t *testing.T) {
		// %v discards the chain. Documented, not desired: the fallback is why
		// this degrades to "unclassified" rather than to a wrong answer.
		flattened := fmt.Errorf("processing message: %v", workflowInvalid)
		if got := MatchedPermanentFailure(flattened); got != "" {
			t.Errorf("a %%v-flattened error carries no code and matches no needle; got %q", got)
		}
	})

	t.Run("ExplicitlyRetryableIsNeverPermanent", func(t *testing.T) {
		retryable := errors.New(errors.ErrValidation, "validation failed").
			AsRetryable(nil).
			Build()
		// Note it would match the "validation" needle — the typed guard must win.
		if got := MatchedPermanentFailure(retryable); got != "" {
			t.Errorf("an author who said AsRetryable outranks the code list; got %q", got)
		}
	})

	t.Run("UntypedTransientIsNotPermanent", func(t *testing.T) {
		if got := MatchedPermanentFailure(fmt.Errorf("context deadline exceeded")); got != "" {
			t.Errorf("MatchedPermanentFailure = %q, want empty", got)
		}
	})

	t.Run("LegacyFallbackHazardPreserved", func(t *testing.T) {
		// pq: invalid connection is a driver fault, not a validation error. It
		// still matches, exactly as before — this fix does not narrow the
		// fallback, and the hazard stays visible (bugs_closed/034 candidate 2).
		if got := MatchedPermanentFailure(fmt.Errorf("pq: invalid connection")); got != "invalid" {
			t.Errorf("fallback behaviour changed: got %q, want invalid", got)
		}
	})

	t.Run("NilIsNotPermanent", func(t *testing.T) {
		if got := MatchedPermanentFailure(nil); got != "" {
			t.Errorf("MatchedPermanentFailure(nil) = %q, want empty", got)
		}
	})
}
