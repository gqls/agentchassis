package actions

import (
	"encoding/json"
	"errors"
	"fmt"
	"testing"

	"github.com/gqls/agentchassis/platform/aiservice"
)

// The decision table for bugs_open/337's escalated retry. The wiring in
// ExecuteLLMPromptAction is a few lines around this function; the function holds
// every branch, so the table here is the behaviour.

func truncated(tokens int) error {
	return &aiservice.TruncatedError{
		Partial:      "<section>cut mid-",
		OutputTokens: tokens,
		Reason:       "stop_reason=max_tokens",
		Provider:     "anthropic",
	}
}

func TestTruncationEscalationApplies(t *testing.T) {
	cfgWith := func(v interface{}) map[string]interface{} {
		return map[string]interface{}{"max_tokens": float64(16000), "max_tokens_ceiling": v}
	}

	cases := []struct {
		name        string
		err         error
		cfg         map[string]interface{}
		sent        int
		sentKnown   bool
		wantCeiling int
		wantOK      bool
	}{
		// The motivating case: generate_template cut at 16,000, ceiling 32,000.
		{"truncated with taller ceiling escalates", truncated(16000), cfgWith(float64(32000)), 16000, true, 32000, true},

		// No key -> byte-for-byte the old behaviour (the opt-in contract).
		{"truncated with no ceiling stays off", truncated(16000),
			map[string]interface{}{"max_tokens": float64(16000)}, 16000, true, 0, false},

		// AN UNKNOWN SENT CAP REFUSES — council round 1, editquality (gating,
		// and right). options["__sent_max_tokens"] is absent on an unconfigured
		// ollama step (ollama.go:121 writes it only alongside num_predict), and
		// reading that absence as a 0 baseline would make the ceiling>cap
		// refusal below vacuous: every positive ceiling would clear 0. These two
		// rows are the difference between the refusal being ENFORCED and merely
		// being described in a comment — the first fails if the guard is dropped,
		// the second proves the sent VALUE is not what carries the decision.
		{"unknown sent cap refuses even with a taller ceiling", truncated(16000), cfgWith(float64(32000)), 0, false, 0, false},
		{"unknown sent cap refuses regardless of the stale int beside it", truncated(16000), cfgWith(float64(32000)), 16000, false, 0, false},

		// A ceiling at or below the sent cap must refuse: retrying at the same
		// height is the deterministic waste this mechanism exists to end.
		{"ceiling equal to sent cap refuses", truncated(16000), cfgWith(float64(16000)), 16000, true, 0, false},
		{"ceiling below sent cap refuses", truncated(16000), cfgWith(float64(8000)), 16000, true, 0, false},

		// Only a typed truncation escalates — never a 5xx, never nil. A 529
		// retried at a taller cap would just be a more expensive 529.
		{"non-truncation error never escalates", errors.New("529 overloaded"), cfgWith(float64(32000)), 16000, true, 0, false},
		{"wrapped non-truncation error never escalates", fmt.Errorf("call: %w", errors.New("overloaded")), cfgWith(float64(32000)), 16000, true, 0, false},
		{"nil error never escalates", nil, cfgWith(float64(32000)), 16000, true, 0, false},

		// A wrapped TruncatedError still counts — errors.As, not equality.
		{"wrapped truncation escalates", fmt.Errorf("AI call failed: %w", truncated(16000)), cfgWith(float64(32000)), 16000, true, 32000, true},

		// An unconfigured step runs at the provider fallback (2048); a declared
		// ceiling still rescues it.
		{"escalates over the provider fallback", truncated(2048), cfgWith(float64(32000)), 2048, true, 32000, true},

		// Zero / negative / junk ceilings are unset, not tiny.
		{"zero ceiling is unset", truncated(16000), cfgWith(float64(0)), 16000, true, 0, false},
		{"negative ceiling is unset", truncated(16000), cfgWith(float64(-1)), 16000, true, 0, false},
		{"string ceiling is unset", truncated(16000), cfgWith("32000"), 16000, true, 0, false},
		{"nil config never escalates", truncated(16000), nil, 16000, true, 0, false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ceiling, ok := truncationEscalationApplies(tc.err, tc.cfg, tc.sent, tc.sentKnown)
			if ok != tc.wantOK || ceiling != tc.wantCeiling {
				t.Fatalf("truncationEscalationApplies(...) = (%d, %v), want (%d, %v)",
					ceiling, ok, tc.wantCeiling, tc.wantOK)
			}
		})
	}
}

// The same key arrives as different Go types depending on the path (jsonb ->
// float64, viper YAML -> int, UseNumber decoders -> json.Number). A
// float64-only read silently drops the YAML case — the exact defect
// aiservice/max_tokens.go documents on `ai_actions.go:361` — so every arrival
// type is pinned here.
func TestTruncationEscalationCeilingNumericTypes(t *testing.T) {
	for name, v := range map[string]interface{}{
		"float64 (jsonb)":       float64(32000),
		"int (viper yaml)":      int(32000),
		"int64":                 int64(32000),
		"json.Number (decoder)": json.Number("32000"),
	} {
		t.Run(name, func(t *testing.T) {
			got, ok := truncationEscalationCeiling(map[string]interface{}{"max_tokens_ceiling": v})
			if !ok || got != 32000 {
				t.Fatalf("truncationEscalationCeiling(%s) = (%d, %v), want (32000, true)", name, got, ok)
			}
		})
	}
	if _, ok := truncationEscalationCeiling(map[string]interface{}{"max_tokens_ceiling": json.Number("not-a-number")}); ok {
		t.Fatal("unparseable json.Number must read as unset")
	}
}
