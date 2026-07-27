// FILE: platform/orchestration/actions/llm_call_logger_thinking_test.go
//
// bugs_open/110 candidate 2. The whole point of these four columns is that they
// can tell "the provider reported zero" from "the provider reported nothing",
// so that is what is tested. Everything else about them is an INSERT.

package actions

import "testing"

func TestThinkingTelemetry_AbsentKeysYieldNilNotZero(t *testing.T) {
	// An anthropic or ollama call leaves none of these keys in the options map.
	// Each must come back nil so the column is NULL — a row of zeroes would
	// assert those providers spent no thinking, which is a claim we cannot make
	// and did not measure.
	wire, reserve, thinking, total := thinkingTelemetry(map[string]interface{}{
		"__sent_max_tokens":     8000,
		"__usage_input_tokens":  4227,
		"__usage_output_tokens": 87,
	})
	for name, got := range map[string]interface{}{
		"wire": wire, "reserve": reserve, "thinking": thinking, "total": total,
	} {
		if got != nil {
			t.Errorf("%s: absent key must yield nil (SQL NULL), got %#v", name, got)
		}
	}
}

func TestThinkingTelemetry_ReportedZeroSurvivesAsZero(t *testing.T) {
	// The regression this guards: routing these through the existing
	// nullIfZero() helper, as every other int field in LLMCallLogParams does.
	// That would map a genuine 0 to NULL and make "spent no thinking" and
	// "provider has no thinking" the same row — the empty-vs-absent confusion
	// bugs_open/110 is itself an instance of.
	_, _, thinking, _ := thinkingTelemetry(map[string]interface{}{
		"__usage_thinking_tokens": 0,
	})
	if thinking == nil {
		t.Fatal("a reported 0 must survive as 0, not become nil/NULL")
	}
	if v, ok := thinking.(int); !ok || v != 0 {
		t.Fatalf("want int 0, got %#v", thinking)
	}
}

func TestThinkingTelemetry_ReadsTheKeysGeminiActuallyWrites(t *testing.T) {
	// Key names are a contract with platform/aiservice/gemini.go and there is no
	// compiler check on them: a typo yields nil, which is indistinguishable from
	// a provider that reports nothing, so the column would just be quietly NULL
	// forever. These are the literals gemini.go sets (grep them there before
	// changing anything here). Values are the writer's real measured shape:
	// visible 8000, wire 16192, reserve 8192.
	wire, reserve, thinking, total := thinkingTelemetry(map[string]interface{}{
		"__sent_wire_max_output_tokens":  16192,
		"__sent_thinking_reserve_tokens": 8192,
		"__usage_thinking_tokens":        2878,
		"__usage_total_tokens":           7105,
	})
	for _, c := range []struct {
		name string
		got  interface{}
		want int
	}{
		{"wire_max_output_tokens", wire, 16192},
		{"thinking_reserve_tokens", reserve, 8192},
		{"thinking_tokens", thinking, 2878},
		{"total_output_tokens", total, 7105},
	} {
		if v, ok := c.got.(int); !ok || v != c.want {
			t.Errorf("%s: want %d, got %#v", c.name, c.want, c.got)
		}
	}
}

func TestThinkingTelemetry_NilOptionsDoesNotPanic(t *testing.T) {
	// LogLLMCall is fire-and-forget on a path that must never take down a
	// workflow; a nil options map is reachable when a client returns early.
	wire, reserve, thinking, total := thinkingTelemetry(nil)
	if wire != nil || reserve != nil || thinking != nil || total != nil {
		t.Fatal("nil options must yield four nils")
	}
}

func TestThinkingTelemetry_WrongTypeIsTreatedAsAbsent(t *testing.T) {
	// A provider writing a float64 or string here is a bug in that provider, but
	// it must not become a panic inside a logging goroutine. nil is the safe
	// reading: NULL says "not captured", which is true.
	_, _, thinking, _ := thinkingTelemetry(map[string]interface{}{
		"__usage_thinking_tokens": "2878",
	})
	if thinking != nil {
		t.Fatalf("a non-int value must be treated as absent, got %#v", thinking)
	}
}
