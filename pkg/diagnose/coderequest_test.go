package diagnose

import (
	"strings"
	"testing"
)

// The wire boundary for the code-search tier. The closed kind-set is the ONLY
// validation here (unlike data_requests there is no SQL to lint — the model
// supplies a pattern, never a query), so this is where it has to hold.
func TestParseVerdictCodeRequests(t *testing.T) {
	t.Run("valid kinds survive, normalised", func(t *testing.T) {
		v, err := ParseVerdict([]byte(`{
		  "outcome": "UNVERIFIABLE",
		  "code_requests": [
		    {"kind": "SYMBOL", "query": "  GenerateText  ", "why": "other adapters?"},
		    {"kind": "content", "query": "%stop_reason%"},
		    {"kind": "ls", "query": "platform/aiservice/"}
		  ]
		}`))
		if err != nil {
			t.Fatalf("parse: %v", err)
		}
		if len(v.CodeRequests) != 3 {
			t.Fatalf("want 3 code requests, got %d: %+v", len(v.CodeRequests), v.CodeRequests)
		}
		if v.CodeRequests[0].Kind != "symbol" {
			t.Errorf("kind should be lower-cased, got %q", v.CodeRequests[0].Kind)
		}
		if v.CodeRequests[0].Query != "GenerateText" {
			t.Errorf("query should be trimmed, got %q", v.CodeRequests[0].Query)
		}
	})

	// An unanswerable question that vanishes silently reads to the verdicter
	// exactly like a question that came back empty — i.e. "the mechanism is
	// absent", which is the most dangerous wrong answer this tier can produce.
	// So the drop must be visible in the trail.
	t.Run("unknown kind is dropped LOUDLY, not silently", func(t *testing.T) {
		v, err := ParseVerdict([]byte(`{
		  "outcome": "UNVERIFIABLE",
		  "code_requests": [{"kind": "grep", "query": "stop_reason", "why": "x"}]
		}`))
		if err != nil {
			t.Fatalf("parse: %v", err)
		}
		if len(v.CodeRequests) != 0 {
			t.Fatalf("unknown kind must not be forwarded: %+v", v.CodeRequests)
		}
		if !strings.Contains(v.NeededEvidence, "grep") {
			t.Errorf("the dropped kind should be named in needed_evidence, got %q", v.NeededEvidence)
		}
		if !strings.Contains(v.NeededEvidence, "not absent") {
			t.Errorf("needed_evidence must warn against reading the drop as absence, got %q", v.NeededEvidence)
		}
	})

	t.Run("empty query is skipped without comment", func(t *testing.T) {
		v, err := ParseVerdict([]byte(`{
		  "outcome": "UNVERIFIABLE",
		  "code_requests": [{"kind": "symbol", "query": "   "}]
		}`))
		if err != nil {
			t.Fatalf("parse: %v", err)
		}
		if len(v.CodeRequests) != 0 {
			t.Fatalf("empty query must not be forwarded: %+v", v.CodeRequests)
		}
	})

	// A verdict from before the code tier existed must parse unchanged.
	t.Run("absent code_requests is not an error", func(t *testing.T) {
		v, err := ParseVerdict([]byte(`{"outcome":"REFUTED","citations":[{"tier":"static","quote":"q","where":"w"}]}`))
		if err != nil {
			t.Fatalf("parse: %v", err)
		}
		if len(v.CodeRequests) != 0 {
			t.Fatalf("want none, got %+v", v.CodeRequests)
		}
	})
}

// The spin guard. A verdict that asks a NEW code question is making progress on
// exactly the same reasoning as one issuing a new data_request: the answer
// arrives in the next gather, so stopping now halts the loop one iteration
// before the evidence it just asked for. Without this the code tier would be
// actively harmful — asking a question would look like spinning.
func TestGuardCreditsNewCodeRequestAsProgress(t *testing.T) {
	staleCitation := []Citation{{Tier: TierStatic, Quote: "q", Where: "w"}}

	t.Run("a new code_request keeps the loop alive on stale citations", func(t *testing.T) {
		seen := map[string]bool{"w|q": true} // citation already seen -> no new evidence
		seenReq := map[string]bool{}
		seenCode := map[string]bool{}
		var hyp []string

		v := Verdict{
			Outcome:      Unverifiable,
			Citations:    staleCitation,
			CodeRequests: []CodeRequest{{Kind: "symbol", Query: "GenerateText"}},
		}
		if stop := guardAfter(v, Scope{}, 5, seen, seenReq, seenCode, &hyp, "h"); stop != "" {
			t.Fatalf("a new code question is progress; guard stopped with %q", stop)
		}
		if !seenCode[CodeRequestKey("symbol", "GenerateText")] {
			t.Errorf("the question should be recorded so a RE-issue does not count again")
		}
	})

	t.Run("re-issuing the SAME code_request does not count as progress", func(t *testing.T) {
		seen := map[string]bool{"w|q": true}
		seenReq := map[string]bool{}
		seenCode := map[string]bool{CodeRequestKey("symbol", "GenerateText"): true}
		var hyp []string

		v := Verdict{
			Outcome:      Unverifiable,
			Citations:    staleCitation,
			CodeRequests: []CodeRequest{{Kind: "SYMBOL", Query: " generatetext "}}, // same question, different spelling
		}
		if stop := guardAfter(v, Scope{}, 5, seen, seenReq, seenCode, &hyp, "h"); stop != "evidence-not-growing" {
			t.Fatalf("a re-issued question is a spin; want evidence-not-growing, got %q", stop)
		}
	})

	// Nil map = a loop state written before this tier existed. It must degrade to
	// the old behaviour, not panic.
	t.Run("nil code-request memory is safe", func(t *testing.T) {
		seen := map[string]bool{}
		var hyp []string
		v := Verdict{Outcome: Unverifiable, Citations: staleCitation,
			CodeRequests: []CodeRequest{{Kind: "symbol", Query: "X"}}}
		_ = guardAfter(v, Scope{}, 5, seen, map[string]bool{}, nil, &hyp, "h")
	})
}

// The guard memory rides collected_data between iterations; if it does not
// round-trip, every iteration re-credits the same question as new and the spin
// guard silently stops working.
func TestLoopStateCarriesCodeRequestMemory(t *testing.T) {
	st := InitLoopState("h", Scope{}, 5, false)
	st.SeenCodeRequests[CodeRequestKey("content", "%stop_reason%")] = true

	var back LoopState
	if err := DecodeLoopState(EncodeLoopState(&st), &back); err != nil {
		t.Fatalf("round-trip: %v", err)
	}
	if !back.SeenCodeRequests[CodeRequestKey("content", "%stop_reason%")] {
		t.Fatalf("code-request memory lost in round-trip: %+v", back.SeenCodeRequests)
	}
}
