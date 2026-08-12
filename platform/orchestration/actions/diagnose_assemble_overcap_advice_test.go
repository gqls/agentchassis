// FILE: platform/orchestration/actions/diagnose_assemble_overcap_advice_test.go
//
// bugs_open/267 — the bundle told the model to re-request a whole file that its
// own arithmetic had already refuted, and the loop spent an iteration obeying it.
// Worked case: platform/orchestration/coordinator.go, 169,139 chars against a
// 60,000-char budget, advised "Put THIS SYMBOL ALONE in next_scope to read it
// whole" by the over-cap marker itself.
//
// Every test here is written to FAIL against the pre-fix code, and each asserts
// BOTH branches of the narrowing it pins. The failure mode to guard against is
// not "the advice is still there" — it is a BLANKET removal, which the bug's §4
// names explicitly: a fix that deletes the sentence outright passes any test that
// only looks for its absence. So the achievable case must keep it, verbatim.
//
// Fixture arithmetic is asserted before every assertion that depends on it: a cap
// test that never trips the cap asserts nothing (016b §9).

package actions

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gqls/agentchassis/internal/analysis"
)

// fnSpec is one function to synthesise: a name, an optional receiver type
// (rendered as a method, which is where the canonical "(*Recv).Name" spelling
// matters), and a body length in padding lines.
type fnSpec struct {
	name  string
	recv  string // "" for a plain func; "*Big" for a pointer method
	lines int
}

// overCapFixture writes ONE Go file holding the given functions and returns the
// checkout root plus the analyser Output describing it.
//
// Spans are recorded WHILE the file is written rather than computed afterwards.
// A hand-counted StartLine/EndLine is the classic way to write a fixture that
// agrees with itself and with nothing else — and every size assertion below is
// derived through analysis.ReadSymbolBody against these spans, so a wrong span
// would silently make the expected and actual numbers agree on a fiction.
func overCapFixture(t *testing.T, path string, fns []fnSpec) (root string, out analysis.Output) {
	t.Helper()
	root = t.TempDir()

	var src strings.Builder
	src.WriteString("package fixture\n\n")
	line := 3 // next line to be written, 1-indexed

	fi := analysis.FileInfo{Path: path, Package: "fixture"}
	for _, f := range fns {
		start := line
		sig := fmt.Sprintf("func %s()", f.name)
		if f.recv != "" {
			sig = fmt.Sprintf("func (r %s) %s()", f.recv, f.name)
		}
		fmt.Fprintf(&src, "%s {\n", sig)
		line++
		for i := 0; i < f.lines; i++ {
			fmt.Fprintf(&src, "\t_ = %d // padding line %d of %s, long enough to give this body a real length\n", i, i, f.name)
			line++
		}
		src.WriteString("}\n")
		end := line
		line++
		src.WriteString("\n")
		line++

		def := analysis.FuncDef{Name: f.name, Signature: sig, StartLine: start, EndLine: end}
		if f.recv != "" {
			def.Receiver = &analysis.Param{Name: "r", Type: f.recv}
		}
		fi.Functions = append(fi.Functions, def)
	}

	if err := os.WriteFile(filepath.Join(root, path), []byte(src.String()), 0o600); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
	return root, analysis.Output{Root: root, Files: []analysis.FileInfo{fi}}
}

// collectedFor round-trips the Output through JSON, because collected_data
// carries the analyser's decoded MAP in production — and "repo_analysis.root",
// which the fix's file-size check reads, is resolved out of that same map.
func collectedFor(t *testing.T, out analysis.Output) map[string]interface{} {
	t.Helper()
	jb, err := json.Marshal(out)
	if err != nil {
		t.Fatalf("marshal Output: %v", err)
	}
	var asMap map[string]interface{}
	if err := json.Unmarshal(jb, &asMap); err != nil {
		t.Fatalf("unmarshal Output: %v", err)
	}
	return map[string]interface{}{"repo_analysis": asMap}
}

func wholeFileSize(t *testing.T, root, path string) int {
	t.Helper()
	src, err := os.ReadFile(filepath.Join(root, path))
	if err != nil {
		t.Fatalf("fixture: read %s: %v", path, err)
	}
	return len(src)
}

// THE BUG. A bare file path whose whole-file body exceeds the WHOLE budget can
// never render under any next_scope, and the marker printed by the cap holds both
// numbers at the moment it decides.
func TestOverCapAdvice_OversizedWholeFileIsNotOfferedWhole(t *testing.T) {
	root, out := overCapFixture(t, "big_file.go", []fnSpec{
		{name: "Alpha", lines: 300},
		{name: "Beta", lines: 60},
		{name: "Gamma", lines: 5},
		{name: "Delta", recv: "*Big", lines: 30},
	})
	fileChars := wholeFileSize(t, root, "big_file.go")

	// The budget is chosen to straddle the file's own symbols: Beta fits, Alpha does
	// not. That is what makes the suggestion list a real assertion — a budget above
	// every symbol would let a fix that suggests indiscriminately pass, which is the
	// same defect one level down.
	body := func(sym string) string {
		t.Helper()
		s, err := analysis.ReadSymbolBody(root, out, sym)
		if err != nil {
			t.Fatalf("fixture: %v", err)
		}
		return s
	}
	alphaChars, betaChars := len(body("big_file.go:Alpha")), len(body("big_file.go:Beta"))
	budget := betaChars + 500
	if fileChars <= budget {
		t.Fatalf("fixture is wrong: the whole file (%d) must exceed the budget (%d) or this test asserts nothing", fileChars, budget)
	}
	if alphaChars <= budget {
		t.Fatalf("fixture is wrong: Alpha (%d) must exceed the budget (%d) so that suggesting it would be a defect", alphaChars, budget)
	}

	m := bodyCapRun(t, collectedFor(t, out), []string{"big_file.go"}, budget)
	bundle := m["bundle"].(string)
	section := inScopeSection(bundle)

	// Pre-fix, this section ended "Put THIS SYMBOL ALONE in next_scope to read it
	// whole." — the request the numbers on the same line refute.
	if strings.Contains(section, "read it whole") {
		t.Fatalf("the bundle still advises reading whole a file that is %d chars against a %d-char budget.\n--- section ---\n%s", fileChars, budget, section)
	}
	if !strings.Contains(section, "body omitted") {
		t.Fatalf("the omission itself stopped being reported — 164's accounting must survive this fix.\n--- section ---\n%s", section)
	}
	// It must say what to ask for INSTEAD, with sizes, in the spelling next_scope
	// resolves.
	for _, want := range []string{"big_file.go:Beta", "chars)"} {
		if !strings.Contains(section, want) {
			t.Errorf("the marker does not name %q — refusing a request without naming a reachable one just moves the dead end.\n--- section ---\n%s", want, section)
		}
	}
	// The suggestion must not repeat the very bug it is closing: Alpha is over the
	// budget, so naming it would be the same impossible request one indirection out.
	if strings.Contains(section, "big_file.go:Alpha") {
		t.Errorf("the marker suggests Alpha (%d chars), which cannot fit the %d-char budget either.\n--- section ---\n%s", alphaChars, budget, section)
	}

	// The METHOD, if suggested, must carry the canonical receiver spelling that
	// splitReceiver accepts — a bare "Delta" would be ambiguous coming back.
	if strings.Contains(section, "big_file.go:Delta") && !strings.Contains(section, "big_file.go:(*Big).Delta") {
		t.Errorf("a method was suggested in a spelling next_scope cannot disambiguate.\n--- section ---\n%s", section)
	}

	// bugs_open/267, second half: the file whose whole-file body did NOT render is
	// exactly the file whose symbol list the model needs. Pre-fix the sibling
	// section skipped it as "whole file already included", which was false.
	if !strings.Contains(bundle, "## Same-file signatures") {
		t.Fatalf("no same-file signature section for the file that could not be rendered whole.\n--- bundle ---\n%s", bundle)
	}
	for _, want := range []string{"big_file.go:Alpha", "big_file.go:Beta"} {
		if !strings.Contains(bundle, want) {
			t.Errorf("the omitted file's symbol %q is not listed anywhere in the bundle.\n--- bundle ---\n%s", want, bundle)
		}
	}
}

// A symbol that alone exceeds the whole budget has nothing smaller to be asked
// for — a function does not subdivide — so the marker must offer no remedy rather
// than an impossible one.
func TestOverCapAdvice_OversizedSingleSymbolOffersNoImpossibleRemedy(t *testing.T) {
	root, out := overCapFixture(t, "one.go", []fnSpec{{name: "Huge", lines: 400}})
	huge, err := analysis.ReadSymbolBody(root, out, "one.go:Huge")
	if err != nil {
		t.Fatalf("fixture: %v", err)
	}
	budget := len(huge) - 1000
	if budget <= 0 || len(huge) <= budget {
		t.Fatalf("fixture is wrong: the symbol (%d) must exceed the budget (%d)", len(huge), budget)
	}

	section := inScopeSection(bodyCapRun(t, collectedFor(t, out), []string{"one.go:Huge"}, budget)["bundle"].(string))
	if strings.Contains(section, "read it whole") {
		t.Fatalf("a symbol %d chars against a %d-char budget is still advised to be re-requested whole.\n--- section ---\n%s", len(huge), budget, section)
	}
	if !strings.Contains(section, "one.go:Huge") || !strings.Contains(section, "body omitted") {
		t.Fatalf("the omission must still be named where the body would have gone.\n--- section ---\n%s", section)
	}
}

// THE NEGATIVE CONTROL, and the reason this fix is a narrowing rather than a
// deletion (267 §4). A body that fits the budget ALONE, and was dropped only
// because earlier bodies had already spent it, is exactly the case the original
// sentence was written for. It must survive verbatim.
func TestOverCapAdvice_AchievableRerequestKeepsTheOriginalAdviceVerbatim(t *testing.T) {
	root, out := overCapFixture(t, "two.go", []fnSpec{
		{name: "Aaa", lines: 60},
		{name: "Zzz", lines: 60},
	})
	first, err := analysis.ReadSymbolBody(root, out, "two.go:Aaa")
	if err != nil {
		t.Fatalf("fixture: %v", err)
	}
	second, err := analysis.ReadSymbolBody(root, out, "two.go:Zzz")
	if err != nil {
		t.Fatalf("fixture: %v", err)
	}
	// Each fits alone; together they do not. Both halves asserted, because a budget
	// that failed either would test the other branch without saying so.
	budget := len(first) + len(second) - 100
	if len(second) > budget {
		t.Fatalf("fixture is wrong: the omitted body (%d) must fit the budget (%d) ALONE", len(second), budget)
	}
	if len(first)+len(second) <= budget {
		t.Fatalf("fixture is wrong: the two bodies (%d) must not fit together under %d", len(first)+len(second), budget)
	}

	section := inScopeSection(bodyCapRun(t, collectedFor(t, out), []string{"two.go:Aaa", "two.go:Zzz"}, budget)["bundle"].(string))
	if !strings.Contains(section, "Put THIS SYMBOL ALONE in next_scope to read it whole.") {
		t.Fatalf("the achievable advice was removed — this fix must be conditional, not a blanket deletion.\n--- section ---\n%s", section)
	}
	// And the summary line must still say the omitted one can be re-asked singly,
	// because here it genuinely can.
	if !strings.Contains(section, "re-request them singly in next_scope.") {
		t.Errorf("the coverage summary dropped advice that is true for every omitted symbol here.\n--- section ---\n%s", section)
	}
}

// The summary line is read by the same model that chooses next_scope, so it must
// not promise for the whole set what is only true of a subset. Pre-fix it said
// "re-request them singly" unconditionally.
func TestOverCapAdvice_SummaryDoesNotPromiseSinglyForTheUnaskable(t *testing.T) {
	root, out := overCapFixture(t, "mix.go", []fnSpec{
		{name: "Aaa", lines: 400},
		{name: "Zzz", lines: 400},
	})
	body, err := analysis.ReadSymbolBody(root, out, "mix.go:Aaa")
	if err != nil {
		t.Fatalf("fixture: %v", err)
	}
	budget := len(body) / 2
	if len(body) <= budget {
		t.Fatalf("fixture is wrong: every body (%d) must exceed the budget (%d)", len(body), budget)
	}

	section := inScopeSection(bodyCapRun(t, collectedFor(t, out), []string{"mix.go:Aaa", "mix.go:Zzz"}, budget)["bundle"].(string))
	if !strings.Contains(section, "This section is INCOMPLETE") {
		t.Fatalf("164's coverage line must still be written.\n--- section ---\n%s", section)
	}
	if strings.Contains(section, "re-request them singly in next_scope.") {
		t.Fatalf("the summary still advises re-requesting singly although no omitted body fits the budget alone.\n--- section ---\n%s", section)
	}
}

// bugs_open/267 candidate 2 — the same unconditional invitation in the sibling
// section's "+N more" marker. BOTH branches, on ONE fixture: the only thing that
// differs between the two runs is the budget, so a blanket removal fails the
// second half and a missing fix fails the first.
func TestSiblingSignatures_BareFilePathOfferedOnlyWhenItCouldFit(t *testing.T) {
	var fns []fnSpec
	for i := 0; i < 40; i++ {
		fns = append(fns, fnSpec{name: fmt.Sprintf("helperNumber%02d", i), lines: 2})
	}
	root, out := overCapFixture(t, "many.go", fns)
	fileChars := wholeFileSize(t, root, "many.go")

	// capChars small enough that the per-file share truncates and a "+N more"
	// marker is emitted at all — assert that, or neither branch is under test.
	const sigCap = 600

	over := siblingSignatures(out, []string{"many.go:helperNumber00"}, sigCap,
		bodyCapView{repoRoot: root, budget: fileChars - 1})
	if !strings.Contains(over, "more in this file") {
		t.Fatalf("fixture is wrong: no +N-more marker was emitted, so neither branch is exercised.\n--- got ---\n%s", over)
	}
	if strings.Contains(over, "put the bare file path in next_scope to see it whole") {
		t.Fatalf("a file of %d chars is still offered whole against a %d-char budget.\n--- got ---\n%s", fileChars, fileChars-1, over)
	}

	// The other branch: the same file, a budget it fits. The advice is correct here
	// and must be unchanged.
	fits := siblingSignatures(out, []string{"many.go:helperNumber00"}, sigCap,
		bodyCapView{repoRoot: root, budget: fileChars + 1})
	if !strings.Contains(fits, "put the bare file path in next_scope to see it whole") {
		t.Fatalf("a file that DOES fit the budget lost its bare-path advice — that is a blanket removal, not a narrowing.\n--- got ---\n%s", fits)
	}

	// And the unknown case — no root, no budget — keeps the old wording, which is
	// what leaves every pre-existing caller byte-identical.
	unknown := siblingSignatures(out, []string{"many.go:helperNumber00"}, sigCap, bodyCapView{})
	if unknown != fits {
		t.Errorf("an unknown file size must degrade to the existing advice, byte for byte.\n--- unknown ---\n%s\n--- fits ---\n%s", unknown, fits)
	}
}
