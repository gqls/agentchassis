package actions

import (
	"strings"
	"testing"

	"github.com/gqls/agentchassis/pkg/diagnose"
)

// The code-search tier's route half. Mirrors TestWithPriorRequests (the F0.5
// data-request case) because the persistence argument is identical: a code
// answer that rides only the bundle after the requesting verdict is LOST the
// moment a guard refuses the confirm that follows, and the loop then re-asks a
// question it already had the answer to.
func TestWithPriorCodeRequests(t *testing.T) {
	cur := func(reqs ...diagnose.CodeRequest) []diagnose.CodeRequest { return reqs }

	t.Run("empty current verdict still re-forwards prior questions", func(t *testing.T) {
		seen := map[string]bool{
			diagnose.CodeRequestKey("symbol", "GenerateText"):   true,
			diagnose.CodeRequestKey("content", "%stop_reason%"): true,
		}
		got, _, _ := withPriorCodeRequests(nil, seen, 10)
		if len(got) != 2 {
			t.Fatalf("want both prior code questions forwarded, got %d: %v", len(got), got)
		}
		first := got[0].(map[string]interface{})
		if !strings.Contains(first["why"].(string), "persists across iterations") {
			t.Fatalf("re-forwarded question should say it is a re-run: %v", first)
		}
		// The kind must survive the round-trip through the guard-map key, or the
		// action cannot answer it (an unknown kind is dropped).
		if !diagnose.ValidCodeRequestKind(first["kind"].(string)) {
			t.Fatalf("re-forwarded kind must be answerable, got %q", first["kind"])
		}
	})

	t.Run("current questions keep their why and dedupe against seen", func(t *testing.T) {
		seen := map[string]bool{diagnose.CodeRequestKey("symbol", "GenerateText"): true}
		got, _, _ := withPriorCodeRequests(
			cur(diagnose.CodeRequest{Kind: "symbol", Query: "GenerateText", Why: "fresh"}), seen, 10)
		if len(got) != 1 {
			t.Fatalf("identical current+seen must not duplicate, got %d: %v", len(got), got)
		}
		if got[0].(map[string]interface{})["why"] != "fresh" {
			t.Fatalf("current question's why must be preserved: %v", got[0])
		}
	})

	t.Run("dedup is case-insensitive on the query, matching the action's own dedup", func(t *testing.T) {
		seen := map[string]bool{diagnose.CodeRequestKey("symbol", "generatetext"): true}
		got, _, _ := withPriorCodeRequests(
			cur(diagnose.CodeRequest{Kind: "symbol", Query: "GenerateText", Why: "fresh"}), seen, 10)
		if len(got) != 1 {
			t.Fatalf("case-different spellings of one question must collapse, got %d: %v", len(got), got)
		}
	})

	t.Run("cap honoured, current first", func(t *testing.T) {
		seen := map[string]bool{}
		for _, q := range []string{"a", "b", "c"} {
			seen[diagnose.CodeRequestKey("symbol", q)] = true
		}
		got, _, _ := withPriorCodeRequests(
			cur(diagnose.CodeRequest{Kind: "content", Query: "zzz", Why: "fresh"}), seen, 2)
		if len(got) != 2 {
			t.Fatalf("cap 2 not honoured: %v", got)
		}
		if got[0].(map[string]interface{})["query"] != "zzz" {
			t.Fatalf("current verdict's question must come first: %v", got)
		}
	})

	// The guard map round-trips through collected_data as JSON, so its keys are
	// DATA, not trusted internals. A malformed key must be skipped, never
	// forwarded as a question the action will then fail to answer.
	t.Run("malformed guard-map keys are skipped, not forwarded", func(t *testing.T) {
		seen := map[string]bool{
			"no-separator-here":                                  true,
			diagnose.CodeRequestKey("wat", "thing"):              true, // unknown kind
			diagnose.CodeRequestKey("symbol", ""):                true, // empty query
			diagnose.CodeRequestKey("ls", "platform/aiservice/"): true, // the only good one
		}
		got, _, malformed := withPriorCodeRequests(nil, seen, 10)
		if len(got) != 1 {
			t.Fatalf("only the well-formed key should survive, got %d: %v", len(got), got)
		}
		// Counted, not silently skipped (council round 5): a malformed key means
		// the collected_data round-trip corrupted them or CodeRequestKey's
		// encoding changed — a defect signal that must not look like silence.
		if malformed != 3 {
			t.Fatalf("want the 3 unaskable keys counted as malformed, got %d", malformed)
		}
		if got[0].(map[string]interface{})["query"] != "platform/aiservice/" {
			t.Fatalf("wrong survivor: %v", got[0])
		}
	})
}

// The route cap must NOT drop silently (council-gate eba040a9, bug-historian,
// medium). The spin guard credits a code question as progress on the promise
// that its answer arrives next gather; a question dropped here is never
// forwarded to be answered, and a re-forwarded prior one dropped here loses an
// answer that was persisting. Either way the trail must be able to say so.
func TestWithPriorCodeRequestsReportsDrops(t *testing.T) {
	t.Run("a question THIS verdict asked, dropped by the cap, is counted", func(t *testing.T) {
		cur := []diagnose.CodeRequest{
			{Kind: "symbol", Query: "A"},
			{Kind: "symbol", Query: "B"},
			{Kind: "symbol", Query: "C"},
		}
		got, dropped, _ := withPriorCodeRequests(cur, nil, 2)
		if len(got) != 2 {
			t.Fatalf("cap 2 not honoured: %v", got)
		}
		if dropped != 1 {
			t.Fatalf("the dropped question must be counted, got dropped=%d", dropped)
		}
	})

	t.Run("prior questions dropped by the cap are counted (the F0.5 answer-loss case)", func(t *testing.T) {
		seen := map[string]bool{}
		for _, q := range []string{"a", "b", "c", "d"} {
			seen[diagnose.CodeRequestKey("symbol", q)] = true
		}
		got, dropped, _ := withPriorCodeRequests(nil, seen, 2)
		if len(got) != 2 || dropped != 2 {
			t.Fatalf("want 2 forwarded / 2 counted as dropped, got %d / %d", len(got), dropped)
		}
	})

	t.Run("malformed keys are NOT counted as drops — they were never askable", func(t *testing.T) {
		seen := map[string]bool{
			"no-separator":                         true,
			diagnose.CodeRequestKey("wat", "x"):    true,
			diagnose.CodeRequestKey("symbol", "y"): true,
		}
		got, dropped, _ := withPriorCodeRequests(nil, seen, 10)
		if len(got) != 1 {
			t.Fatalf("only the well-formed key should forward, got %v", got)
		}
		if dropped != 0 {
			t.Fatalf("skipping an unaskable key is not a DROP; got dropped=%d", dropped)
		}
	})

	t.Run("nothing dropped when under the cap", func(t *testing.T) {
		_, dropped, _ := withPriorCodeRequests(
			[]diagnose.CodeRequest{{Kind: "ls", Query: "platform/"}}, nil, 10)
		if dropped != 0 {
			t.Fatalf("want 0, got %d", dropped)
		}
	})
}
