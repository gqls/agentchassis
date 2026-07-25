// FILE: platform/messaging/validation_drop_test.go
package messaging

import "testing"

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
