package diagnose

import (
	"strings"
	"testing"
)

func TestParseVerdict_StringFormat(t *testing.T) {
	raw := []byte(`{
	  "outcome": "REFUTED",
	  "citations": [
	    {"tier": "runtime", "where": "agent_error_log", "quote": "content regression blocked: ~3k vs ~13k", "fresh": "2026-06-14"}
	  ],
	  "revised_hypothesis": "the regeneration is short upstream",
	  "next_scope": ["plan_sections_action.go:PlanSectionsAction"]
	}`)
	v, err := ParseVerdict(raw)
	if err != nil {
		t.Fatal(err)
	}
	if v.Outcome != Refuted {
		t.Fatalf("outcome: want Refuted, got %s", v.Outcome)
	}
	if len(v.Citations) != 1 || v.Citations[0].Tier != TierRuntime {
		t.Fatalf("citation tier not parsed: %+v", v.Citations)
	}
	if v.Citations[0].Quote == "" || v.RevisedHypothesis == "" || len(v.NextScope) != 1 {
		t.Fatalf("fields not mapped: %+v", v)
	}
}

func TestParseVerdict_UnknownOutcomeFailsSafe(t *testing.T) {
	// A malformed outcome must NOT confirm — fail safe to UNVERIFIABLE.
	raw := []byte(`{"outcome":"DEFINITELY","citations":[{"tier":"static","where":"x.go:Y","quote":"q"}]}`)
	v, err := ParseVerdict(raw)
	if err != nil {
		t.Fatal(err) // JSON itself is valid; the bad value is handled, not errored
	}
	if v.Outcome != Unverifiable {
		t.Fatalf("unknown outcome must fail safe to Unverifiable, got %s", v.Outcome)
	}
	if v.NeededEvidence == "" {
		t.Fatal("the outcome error should be surfaced in NeededEvidence")
	}
}

func TestParseVerdict_TierDefaultsStatic(t *testing.T) {
	raw := []byte(`{"outcome":"CONFIRMED","citations":[{"tier":"weird","where":"a.go:B","quote":"q"}]}`)
	v, _ := ParseVerdict(raw)
	if v.Citations[0].Tier != TierStatic {
		t.Fatalf("unknown tier should default to static, got %s", v.Citations[0].Tier)
	}
}

func TestParseVerdicts_ArrayScript_GamesdesignPath(t *testing.T) {
	// The real gamesdesign reasoning path, in the model wire format — the same
	// bytes the model would emit, used as a script.
	raw := []byte(`[
	  {"outcome":"REFUTED",
	   "citations":[{"tier":"runtime","where":"agent_error_log","quote":"content regression blocked"}],
	   "revised_hypothesis":"short upstream","next_scope":["plan_sections_action.go:PlanSectionsAction"]},
	  {"outcome":"REFUTED",
	   "citations":[{"tier":"static","where":"content_writer","quote":"max_tokens 2000"}],
	   "revised_hypothesis":"discarded at result extraction","next_scope":["coordinator.go:extractWorkflowResult"]},
	  {"outcome":"CONFIRMED",
	   "citations":[{"tier":"static","where":"result_spec.go:resolveResultSpec","quote":"singular output_field ignored"}]}
	]`)
	vs, err := ParseVerdicts(raw)
	if err != nil {
		t.Fatal(err)
	}
	if len(vs) != 3 {
		t.Fatalf("want 3 verdicts, got %d", len(vs))
	}
	if vs[0].Outcome != Refuted || vs[1].Outcome != Refuted || vs[2].Outcome != Confirmed {
		t.Fatalf("outcome sequence wrong: %s,%s,%s", vs[0].Outcome, vs[1].Outcome, vs[2].Outcome)
	}
	// feed it through the loop to confirm the wire format drives the scaffold
	res, err := Run("sections never reach save",
		Scope{Symbols: []string{"save_page_sections_action.go:SavePageSectionsAction"}},
		&fakeGather{}, &scriptVerdicter{verdicts: vs}, nil, DefaultConfig())
	if err != nil {
		t.Fatal(err)
	}
	if res.Status != Confirmed {
		t.Fatalf("wire-format script should drive loop to Confirmed, got %s (%s)", res.Status, res.StoppedBy)
	}
}

func TestParseVerdict_DataRequestsLintedAtParse(t *testing.T) {
	raw := []byte(`{
	  "outcome": "UNVERIFIABLE",
	  "needed_evidence": "need the page rows",
	  "data_requests": [
	    {"sql": "SELECT name, status FROM pages WHERE site_id = $1", "why": "what pages exist"},
	    {"sql": "WITH t AS (DELETE FROM pages RETURNING *) SELECT * FROM t", "why": "sneaky"},
	    {"sql": "SELECT 1; DROP TABLE pages", "why": "stacked"}
	  ]
	}`)
	v, err := ParseVerdict(raw)
	if err != nil {
		t.Fatal(err)
	}
	// only the read-only SELECT survives; the data-modifying CTE and the stacked
	// statement are dropped at parse
	if len(v.DataRequests) != 1 {
		t.Fatalf("expected 1 surviving data request, got %d: %+v", len(v.DataRequests), v.DataRequests)
	}
	if v.DataRequests[0].SQL != "SELECT name, status FROM pages WHERE site_id = $1" {
		t.Fatalf("wrong request survived: %q", v.DataRequests[0].SQL)
	}
	// the drops are surfaced in the trail, not silent
	if !strings.Contains(v.NeededEvidence, "dropped non-read-only data_request") {
		t.Fatalf("dropped requests not surfaced in NeededEvidence: %q", v.NeededEvidence)
	}
}
