// FILE: platform/orchestration/actions/diagnose_assemble_bodycap_test.go
//
// bugs_open/164 — the diagnosis bundle's body cap used to `break`, so ONE
// oversized symbol silently dropped the whole rest of scope. Scope is
// sort.Strings-SORTED upstream (pkg/diagnose/loop.go:390,:416), so the casualties
// were an ALPHABETICAL tail, not a least-relevant one.
//
// These tests are written to FAIL against the pre-fix loop. Each one forces the
// cap and then asserts on BOTH branches — the marker appearing AND the following
// symbol's body surviving. A cap test that never trips the cap asserts nothing
// (016b §9, and 137's "verify a narrowing with BOTH branches": the finding branch
// alone is satisfied by deleting the guard).

package actions

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gqls/agentchassis/internal/analysis"
	"github.com/gqls/agentchassis/pkg/models"
	orchtypes "github.com/gqls/agentchassis/platform/orchestration/types"
	"go.uber.org/zap"
)

// bodyCapFixture writes a throwaway checkout holding one big and one small Go
// file and returns the collected_data an assemble step would see.
//
// The names matter: "aaa_big.go" sorts BEFORE "zzz_small.go", which is the real
// production ordering (scope arrives sorted) and the exact arrangement in which
// the old `break` destroyed the small file. A fixture that put the big file last
// would pass against the bug.
func bodyCapFixture(t *testing.T, bigLines, smallLines int) (map[string]interface{}, string, analysis.Output) {
	t.Helper()
	root := t.TempDir()

	mk := func(name string, n int) {
		var b strings.Builder
		b.WriteString("package fixture\n\n")
		b.WriteString("func Target() {\n")
		for i := 0; i < n; i++ {
			fmt.Fprintf(&b, "\t_ = %d // padding line %d to give this body a real length\n", i, i)
		}
		b.WriteString("}\n")
		if err := os.WriteFile(filepath.Join(root, name), []byte(b.String()), 0o600); err != nil {
			t.Fatalf("write %s: %v", name, err)
		}
	}

	mk("aaa_big.go", bigLines)
	mk("zzz_small.go", smallLines)

	// The analyser Output is both the span source and (since bugs_closed/145) the
	// READ BOUNDARY, so every readable fixture file must appear here.
	out := analysis.Output{Root: root, Files: []analysis.FileInfo{
		{Path: "aaa_big.go", Package: "fixture", Functions: []analysis.FuncDef{
			{Name: "Target", Signature: "func Target()", StartLine: 3, EndLine: 3 + bigLines + 1},
		}},
		{Path: "zzz_small.go", Package: "fixture", Functions: []analysis.FuncDef{
			{Name: "Target", Signature: "func Target()", StartLine: 3, EndLine: 3 + smallLines + 1},
		}},
	}}

	// collected_data carries the analyser Output as a decoded MAP in production
	// (the adapter's Output), which is also what repo_root_field "repo_analysis.root"
	// resolves against — so round-trip it rather than handing the action a struct
	// it would never see.
	jb, err := json.Marshal(out)
	if err != nil {
		t.Fatalf("marshal Output: %v", err)
	}
	var asMap map[string]interface{}
	if err := json.Unmarshal(jb, &asMap); err != nil {
		t.Fatalf("unmarshal Output: %v", err)
	}

	return map[string]interface{}{"repo_analysis": asMap}, root, out
}

// bodyOf returns EXACTLY what the action will render for a symbol — the span the
// analyser Output describes, not the whole file. Deriving the expected text
// through the same reader the action uses is what makes the byte-identity control
// below a real control rather than a restatement of the fixture.
func bodyOf(t *testing.T, root string, out analysis.Output, sym string) string {
	t.Helper()
	body, err := analysis.ReadSymbolBody(root, out, sym)
	if err != nil {
		t.Fatalf("fixture: ReadSymbolBody(%s): %v", sym, err)
	}
	return body
}

func bodyCapRun(t *testing.T, collected map[string]interface{}, scope []string, capChars int) map[string]interface{} {
	t.Helper()
	collected["scope"] = scope
	out, err := DiagnoseAssembleBundleAction(context.Background(), ActionParams{
		Context:          context.Background(),
		Logger:           zap.NewNop(),
		ExecutionContext: &orchtypes.ExecutionContext{Action: "process"},
		CollectedData:    collected,
		StepConfig: models.Step{Config: map[string]interface{}{
			"scope_field":    "scope",
			"max_body_chars": capChars,
			"persist_bundle": false, // no DB in a unit test; egress is not under test
		}},
	})
	if err != nil {
		t.Fatalf("action error: %v", err)
	}
	m, ok := out.(map[string]interface{})
	if !ok {
		t.Fatalf("expected map result, got %T", out)
	}
	return m
}

// inScopeSection returns just the "## In-scope code" section, so an assertion
// about what the verdicter sees as CODE cannot be satisfied by a mention of the
// same symbol in the sibling-signatures section further down.
func inScopeSection(bundle string) string {
	start := strings.Index(bundle, "## In-scope code")
	if start < 0 {
		return ""
	}
	rest := bundle[start+len("## In-scope code"):]
	if end := strings.Index(rest, "\n## "); end >= 0 {
		return rest[:end]
	}
	return rest
}

// THE BUG. A body that does not fit must be SKIPPED, not end the loop, and the
// symbol after it must still render.
func TestBundleBodyCap_OversizedSymbolDoesNotEvictTheRest(t *testing.T) {
	collected, root, out := bodyCapFixture(t, 400, 3)
	bigBody := bodyOf(t, root, out, "aaa_big.go:Target")
	smallBody := bodyOf(t, root, out, "zzz_small.go:Target")
	// A cap that admits the small body but not the big one — the whole point is
	// that the cap genuinely trips, so assert the fixture arithmetic first.
	capChars := len(smallBody) + 200
	if len(bigBody) <= capChars {
		t.Fatalf("fixture is wrong: big body (%d) must exceed the cap (%d) or this test asserts nothing", len(bigBody), capChars)
	}

	m := bodyCapRun(t, collected, []string{"aaa_big.go:Target", "zzz_small.go:Target"}, capChars)
	section := inScopeSection(m["bundle"].(string))

	// Branch 1 — the omission is NAMED where the body would have gone. Before the
	// fix `truncated` was set and the bundle text said nothing at all.
	if !strings.Contains(section, "aaa_big.go:Target") || !strings.Contains(section, "body omitted") {
		t.Fatalf("oversized symbol not named as omitted in the bundle TEXT.\n--- section ---\n%s", section)
	}
	// The marker must carry the numbers a reader needs to act on it.
	if !strings.Contains(section, fmt.Sprintf("%d chars", len(bigBody))) {
		t.Errorf("omission marker does not state the body's real size (%d chars)", len(bigBody))
	}

	// Branch 2 — THE REGRESSION. The alphabetically-later small symbol still
	// renders. This is the assertion the pre-fix `break` fails.
	if !strings.Contains(section, "zzz_small.go:Target") {
		t.Fatalf("the symbol AFTER the oversized one was dropped — the `break` regression is back.\n--- section ---\n%s", section)
	}
	if !strings.Contains(section, "padding line 2") {
		t.Fatalf("zzz_small.go was named but its BODY is absent — a heading is not a body.\n--- section ---\n%s", section)
	}
	// And the big body itself must NOT be inlined (that would defeat the cap).
	if strings.Contains(section, "padding line 399") {
		t.Fatal("the oversized body was rendered anyway — the cap no longer caps")
	}

	if m["truncated"] != true {
		t.Errorf("truncated should be true when a body was dropped, got %v", m["truncated"])
	}
	if m["symbol_count"] != 1 {
		t.Errorf("symbol_count should count only RENDERED bodies (1), got %v", m["symbol_count"])
	}
	// The coverage line the verdicter can read (:304 logs it; a log is not a bundle).
	if !strings.Contains(section, "This section is INCOMPLETE") || !strings.Contains(section, "1 of 2") {
		t.Errorf("no readable coverage line naming 1 of 2 rendered.\n--- section ---\n%s", section)
	}
}

// The whole scope oversized — the manifestation actually seen in production
// (3 of 254 bundles rendered the "## In-scope code" heading with NOTHING under
// it, because the FIRST body alone exceeded the cap).
func TestBundleBodyCap_EveryBodyOversizedStillExplainsItself(t *testing.T) {
	collected, _, _ := bodyCapFixture(t, 400, 400)
	m := bodyCapRun(t, collected, []string{"aaa_big.go:Target", "zzz_small.go:Target"}, 100)
	section := inScopeSection(m["bundle"].(string))

	if m["symbol_count"] != 0 {
		t.Fatalf("fixture wrong: expected 0 rendered bodies, got %v", m["symbol_count"])
	}
	// Pre-fix this section was EMPTY: heading, blank line, next heading. The
	// verdicter could not tell "no code in scope" from "all of it was dropped".
	if strings.TrimSpace(section) == "" {
		t.Fatal("the in-scope section is empty — this is the exact production artefact 164 was filed for")
	}
	for _, want := range []string{"aaa_big.go:Target", "zzz_small.go:Target", "body omitted", "0 of 2"} {
		if !strings.Contains(section, want) {
			t.Errorf("section does not mention %q.\n--- section ---\n%s", want, section)
		}
	}
}

// A scope entry the analyser never parsed is refused by bugs_closed/145's read
// boundary. That refusal was ALSO silent to the verdicter, in the same loop and
// six lines up — fixing only the cap half would leave the sibling heuristic
// (016b §9 / bug 021), so both discard paths report. They must report
// DIFFERENTLY: a read failure is a defect signal, not a coverage signal
// (bd003f67a's ruling at the sibling cap in this same file).
func TestBundleBodyCap_UnreadableSymbolIsNamedAndDistinguished(t *testing.T) {
	collected, _, _ := bodyCapFixture(t, 3, 3)
	m := bodyCapRun(t, collected, []string{"aaa_big.go:Target", "not/in/analysis.go:Ghost"}, 60000)
	section := inScopeSection(m["bundle"].(string))

	if !strings.Contains(section, "not/in/analysis.go:Ghost") || !strings.Contains(section, "body unavailable") {
		t.Fatalf("an unreadable scope entry vanished from the bundle.\n--- section ---\n%s", section)
	}
	if !strings.Contains(section, "TOOLING failure") {
		t.Errorf("the unreadable marker must not read as a coverage cap — it is a defect signal")
	}
	// A read failure is NOT a truncation: conflating them would make the cap's own
	// rate unmeasurable again, which is why 164 had to be filed [UNMEASURED].
	if m["truncated"] != false {
		t.Errorf("truncated must stay false when nothing was dropped for SIZE, got %v", m["truncated"])
	}
	// The readable sibling still renders.
	if !strings.Contains(section, "padding line 2") {
		t.Errorf("the readable symbol's body is missing.\n--- section ---\n%s", section)
	}
}

// NEGATIVE CONTROL (164's stated requirement). A scope that fits entirely must
// produce the OLD bytes — otherwise the fix moves every existing diagnosis's
// baseline. Asserted as an exact match against the pre-fix format string rather
// than by absence of the new wording, because "no marker appeared" would also
// pass if the section's shape had changed some other way.
func TestBundleBodyCap_FittingScopeIsByteIdenticalToThePreFixFormat(t *testing.T) {
	collected, root, out := bodyCapFixture(t, 40, 3)
	m := bodyCapRun(t, collected, []string{"aaa_big.go:Target", "zzz_small.go:Target"}, 60000)
	bundle := m["bundle"].(string)

	// Exactly what the pre-fix loop emitted for two fitting symbols.
	want := "## In-scope code\n\n" +
		fmt.Sprintf("### %s\n```go\n%s\n```\n\n", "aaa_big.go:Target", bodyOf(t, root, out, "aaa_big.go:Target")) +
		fmt.Sprintf("### %s\n```go\n%s\n```\n\n", "zzz_small.go:Target", bodyOf(t, root, out, "zzz_small.go:Target"))
	if !strings.Contains(bundle, want) {
		t.Fatalf("a fitting scope no longer renders byte-identically to the pre-fix format.\n--- got section ---\n%s", inScopeSection(bundle))
	}
	for _, unwanted := range []string{"body omitted", "body unavailable", "INCOMPLETE"} {
		if strings.Contains(bundle, unwanted) {
			t.Errorf("bundle carries %q although nothing was dropped — every existing diagnosis's baseline just moved", unwanted)
		}
	}
	if m["truncated"] != false || m["symbol_count"] != 2 {
		t.Errorf("want truncated=false symbol_count=2, got %v / %v", m["truncated"], m["symbol_count"])
	}
}
