package diagnose

import (
	"encoding/json"
	"strings"
	"testing"
)

// bugs_open/056: the original key delimiter was a literal NUL ("\x00").
// json.Marshal encodes a NUL as \u0000 — the ONE Unicode escape Postgres
// jsonb rejects (22P05) — and SeenCodeRequests round-trips through the jsonb
// collected_data column, so the first persist after any code request killed
// the whole diagnosis run. The key must therefore never contain a NUL.
func TestCodeRequestKeySurvivesJSONBMarshal(t *testing.T) {
	key := CodeRequestKey("symbol", "GenerateText")
	if strings.ContainsRune(key, 0) {
		t.Fatalf("CodeRequestKey contains a NUL byte: %q", key)
	}
	marshalled, err := json.Marshal(map[string]bool{key: true})
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if strings.Contains(string(marshalled), `\u0000`) {
		t.Fatalf("marshalled SeenCodeRequests carries the jsonb-fatal escape: %s", marshalled)
	}
}

// Builder and reader must stay in lockstep or cross-iteration dedup breaks
// silently — it reads as "malformed keys" in the route's counter, not a crash.
func TestSplitCodeRequestKeyIsTheInverse(t *testing.T) {
	kind, query, ok := SplitCodeRequestKey(CodeRequestKey("  Symbol ", "  GenerateText  "))
	if !ok {
		t.Fatal("SplitCodeRequestKey did not find the separator")
	}
	if kind != "symbol" || query != "generatetext" {
		t.Fatalf("round-trip mangled the key: kind=%q query=%q", kind, query)
	}
	if !ValidCodeRequestKind(kind) {
		t.Fatalf("split kind %q no longer passes ValidCodeRequestKind", kind)
	}
}
