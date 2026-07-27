package handlers

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"strings"
	"testing"

	"github.com/gqls/agentchassis/platform/aiservice"
)

// captureLog redirects the standard logger for the duration of fn and returns
// what was written. The flags are cleared so assertions do not depend on the
// timestamp prefix.
func captureLog(fn func()) string {
	var buf bytes.Buffer
	oldOut := log.Writer()
	oldFlags := log.Flags()
	log.SetOutput(&buf)
	log.SetFlags(0)
	defer func() {
		log.SetOutput(oldOut)
		log.SetFlags(oldFlags)
	}()
	fn()
	return buf.String()
}

// The whole point of bugs_open/083 is that a 503 with no record is
// undiagnosable. These tests assert the record EXISTS and CARRIES THE CAUSE —
// an implementation that logged a bare "call failed" would satisfy "something
// was logged" and still leave the bug open, so each case asserts the specific
// discriminating detail.

func TestLogAIFailure_RecordsTheUnderlyingError(t *testing.T) {
	out := captureLog(func() {
		logAIFailure("position", "generate", "round-123",
			errors.New("API request failed with status 529"))
	})

	for _, want := range []string{"gauntlet/position", "generate", "FAILED", "round-123", "status 529"} {
		if !strings.Contains(out, want) {
			t.Errorf("log line missing %q\ngot: %s", want, out)
		}
	}
}

// The discriminating case. A truncated completion is NOT an upstream fault — it
// is our own max_tokens cap — so it needs a different fix and must be
// distinguishable in the log. An implementation that just printed err would
// still produce a line, so this asserts the TRUNCATED label and the provider's
// own numbers, which only the IsTruncated branch can supply.
func TestLogAIFailure_NamesTruncationDistinctly(t *testing.T) {
	te := &aiservice.TruncatedError{
		Partial:      "the judge began to answer and then",
		OutputTokens: 2048,
		Reason:       "stop_reason=max_tokens",
		Provider:     "anthropic",
	}

	out := captureLog(func() { logAIFailure("defend", "generate", "round-456", te) })

	if !strings.Contains(out, "TRUNCATED") {
		t.Fatalf("a truncation must be labelled TRUNCATED, not folded into a generic failure\ngot: %s", out)
	}
	for _, want := range []string{"stop_reason=max_tokens", "output_tokens=2048", "provider=anthropic", "round-456"} {
		if !strings.Contains(out, want) {
			t.Errorf("truncation log missing %q\ngot: %s", want, out)
		}
	}
	if strings.Contains(out, "FAILED") {
		t.Errorf("truncation should not also be reported as a generic FAILED\ngot: %s", out)
	}
}

// A truncation reached through a wrapped error must still be recognised —
// IsTruncated uses errors.As precisely so callers need not unwrap by hand. If
// this regressed, a wrapped truncation would silently log as a generic failure
// and the §2 truncation question could never be settled.
func TestLogAIFailure_FindsTruncationThroughWrapping(t *testing.T) {
	te := &aiservice.TruncatedError{OutputTokens: 512, Reason: "done_reason=length", Provider: "ollama"}
	wrapped := fmt.Errorf("generate text: %w", te)

	out := captureLog(func() { logAIFailure("defend", "generate", "r1", wrapped) })

	if !strings.Contains(out, "TRUNCATED") {
		t.Fatalf("wrapped truncation was not recognised\ngot: %s", out)
	}
}

// "Unparseable" alone cannot distinguish a prose wrapper from a second JSON
// object from an empty completion, and those have different fixes — so the body
// snippet is load-bearing, not decoration.
func TestLogAIBadResponse_IncludesTheOffendingBody(t *testing.T) {
	body := "Sure! Here is the JSON you asked for:\n```json\n{\"verdict\":\"user wins\"}\n```"

	out := captureLog(func() {
		logAIBadResponse("defend", "json_unmarshal: invalid character 'S'", "round-789", body)
	})

	for _, want := range []string{"UNUSABLE", "round-789", "json_unmarshal", "Here is the JSON"} {
		if !strings.Contains(out, want) {
			t.Errorf("bad-response log missing %q\ngot: %s", want, out)
		}
	}
}

func TestLogAIBadResponse_CapsAVeryLongBody(t *testing.T) {
	body := strings.Repeat("x", maxLoggedBody*3)

	out := captureLog(func() { logAIBadResponse("position", "empty fields", "r2", body) })

	if !strings.Contains(out, "truncated for log") {
		t.Errorf("an oversized body must be capped\ngot %d chars", len(out))
	}
	if len(out) > maxLoggedBody*2 {
		t.Errorf("capped line is still too long: %d chars", len(out))
	}
	// The true length must survive the cap, or the log understates what arrived.
	if !strings.Contains(out, "body_chars="+fmt.Sprint(len(body))) {
		t.Errorf("capped line must still report the ORIGINAL body length\ngot: %s", out)
	}
}

func TestSnippet_LeavesAShortBodyIntact(t *testing.T) {
	if got := snippet("short"); got != "short" {
		t.Errorf("snippet altered a short body: %q", got)
	}
}
