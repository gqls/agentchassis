// FILE: platform/messaging/retryable_transient_test.go
//
// bugs_open/197. Nothing pinned the retryable-side classifier at all before
// this file — zero test hits on the old isRecoverableError, and
// platform/errors had no test file — which is exactly how its case-sensitive
// needle list ran for months while ~30% of its real input population was
// misclassified terminal.
//
// The census cases below use REAL error messages from agent_error_log
// (population measured 2026-08-06: 2,996 rows; the figures live in
// retryable_transient.go's header). Each needle is pinned in BOTH
// directions: a case that fails if the needle is removed, and a case that
// fails if the list is widened to the rejected bare-digit patterns.
package messaging

import (
	"fmt"
	"testing"

	"github.com/gqls/agentchassis/platform/errors"
)

func TestMatchedTransientFailure(t *testing.T) {
	t.Run("CensusHeadline_nestedDeadlineExceededIsTransient", func(t *testing.T) {
		// The 885-row class: deeply nested workflow wrappers around
		// context.DeadlineExceeded prose. The old classifier made every one
		// of these TERMINAL (no needle matched). Removing "deadline exceeded"
		// from the list fails this case.
		msg := fmt.Errorf("workflow failed: Request 7f3a2c1e-8b4d-4e6f-9a0b-1c2d3e4f5a6b failed: " +
			"workflow failed: step generate_content failed: failed to execute action " +
			"execute_llm_prompt: context deadline exceeded")
		if got := MatchedTransientFailure(msg); got != "deadline exceeded" {
			t.Errorf("MatchedTransientFailure = %q, want %q", got, "deadline exceeded")
		}
	})

	t.Run("CensusCaseFold_capitalisedNeedlesNowMatch", func(t *testing.T) {
		// The 882-row class: capitalised variants missed on case alone —
		// bugs_closed/195's capital-letter defect, live on this sibling until
		// now. Removing the ToLower fold fails all three.
		for _, tc := range []struct{ msg, want string }{
			{"API error: Timeout waiting for model response", "timeout"},
			{"Connection reset by peer", "connection"},
			{"Temporary failure in name resolution", "temporary"},
		} {
			if got := MatchedTransientFailure(fmt.Errorf("%s", tc.msg)); got != tc.want {
				t.Errorf("MatchedTransientFailure(%q) = %q, want %q", tc.msg, got, tc.want)
			}
		}
	})

	t.Run("PermanentErrorIsNotTransient", func(t *testing.T) {
		// The bugs_closed/195 error. It never reaches this classifier in
		// production (MatchedPermanentFailure drops it first), but if it did,
		// it must not read as retryable — a static config error cannot be
		// cured by retrying.
		msg := fmt.Errorf("WORKFLOW_INVALID: Invalid workflow configuration (caused by: step 'done' with action 'complete' requires a topic)")
		if got := MatchedTransientFailure(msg); got != "" {
			t.Errorf("a statically-broken workflow must not classify transient, got %q", got)
		}
	})

	t.Run("RejectedDigits_hexUUIDCannotClassify", func(t *testing.T) {
		// Pins the REJECTION of bare "502"/"503"/"504": the dominant census
		// message shape nests correlation UUIDs, which are hex, so a bare
		// digit needle would classify an arbitrary permanent failure as
		// retryable off its own correlation id. This message's only "503"
		// sits inside the UUID. Re-adding a bare-digit needle fails this.
		msg := fmt.Errorf("workflow failed: Request a1b2503c-9d8e-4f6a-b5c4-d3e2f1a0b9c8 failed: " +
			"persist_plan discarded the design")
		if got := MatchedTransientFailure(msg); got != "" {
			t.Errorf("a hex UUID must never satisfy a needle, got %q", got)
		}
	})

	t.Run("SSLOverMatchIsKnownAndBounded", func(t *testing.T) {
		// The documented over-match, asserted so it stays VISIBLE rather than
		// discovered again: a TLS-certificate failure matches "connection"
		// inside "establish a secure connection" and buys three futile
		// retries. If this case ever returns "", the over-match was fixed —
		// update the file header's known-over-matches list in the same change.
		msg := fmt.Errorf("API error: An SSL/TLS certificate error occurred while trying to " +
			"establish a secure connection to this website")
		if got := MatchedTransientFailure(msg); got != "connection" {
			t.Errorf("the documented SSL over-match changed: got %q, want %q", got, "connection")
		}
	})

	t.Run("AuthorRetryableOutranksEverything", func(t *testing.T) {
		// An explicit AsRetryable survives %w wrapping and reports through
		// the retryable: prefix — the author's intent, not the list.
		de := errors.New(errors.ErrExternalService, "provider hiccup").
			AsRetryable(nil).Build()
		wrapped := fmt.Errorf("calling provider: %w", de)
		if got := MatchedTransientFailure(wrapped); got != "retryable:EXTERNAL_SERVICE_ERROR" {
			t.Errorf("MatchedTransientFailure = %q, want retryable:EXTERNAL_SERVICE_ERROR", got)
		}
	})

	t.Run("TypedCodeAnswersWithoutNeedleHelp", func(t *testing.T) {
		// ErrAgentTimeout with prose that matches NO needle — the typed
		// branch must answer alone. Removing the code list fails this.
		de := errors.New(errors.ErrAgentTimeout, "attempt budget exhausted").Build()
		if got := MatchedTransientFailure(de); got != "code:AGENT_TIMEOUT" {
			t.Errorf("MatchedTransientFailure = %q, want code:AGENT_TIMEOUT", got)
		}
	})

	t.Run("RetryableFalseFallsThroughToNeedles", func(t *testing.T) {
		// InternalError defaults Retryable=false while WRAPPING a transient
		// cause. Classifying on the flag alone would make every wrapped
		// timeout terminal — this bug re-created one level up. The fall-through
		// is load-bearing; making Retryable=false return "" fails this.
		wrapped := errors.InternalError("processing message", fmt.Errorf("context deadline exceeded"))
		if got := MatchedTransientFailure(wrapped); got != "deadline exceeded" {
			t.Errorf("a Retryable=false DomainError wrapping transient prose must fall through, got %q", got)
		}
	})

	t.Run("UnclassifiableIsTerminal", func(t *testing.T) {
		if got := MatchedTransientFailure(fmt.Errorf("probe: unclassifiable failure xyzzy")); got != "" {
			t.Errorf("MatchedTransientFailure = %q, want empty", got)
		}
	})

	t.Run("NilIsNotTransient", func(t *testing.T) {
		// The old isRecoverableError had no nil guard and would panic here.
		if got := MatchedTransientFailure(nil); got != "" {
			t.Errorf("MatchedTransientFailure(nil) = %q, want empty", got)
		}
	})
}

// TestTransientNeedlesAreTheJudgedList pins the pattern set byte-for-byte,
// in the style of TestValidationNeedlesAreTheOnesBothLayersUsed: any change
// to the list must come through this test and re-argue the header's
// admission criteria — in particular the REJECTED entries (bare digits,
// "unreachable", "network"), whose reasons are in the file header.
func TestTransientNeedlesAreTheJudgedList(t *testing.T) {
	want := []string{
		"deadline exceeded",
		"timeout",
		"connection",
		"temporary",
		"dial tcp",
		"too many requests",
		"rate limit",
		"service unavailable",
		"bad gateway",
	}
	if len(transientErrorNeedles) != len(want) {
		t.Fatalf("transientErrorNeedles = %v, want %v", transientErrorNeedles, want)
	}
	for i, w := range want {
		if transientErrorNeedles[i] != w {
			t.Errorf("transientErrorNeedles[%d] = %q, want %q", i, transientErrorNeedles[i], w)
		}
	}
}
