// internal/analysis/symbolbody_test.go  (package analysis)
//
// Behaviour test: analyse a small fixture, then assert ReadSymbolBody slices the
// analyser's spans the way cmd/assembler does — func body from the `func` line to
// its closing brace (NO doc comment), types whole, receiver-qualified methods
// disambiguated, whole-file when no symbol is given, and clean errors (not
// panics) for unknown path/symbol.
//
// (The reference tool is contextkit's `cmd/assembler`; this header said
// `cmd/bundle`, which has never existed in this repo — corrected 2026-07-31 with
// the same citation error in symbolbody.go's own header, per bugs_open/145.)
//
// TestReadSymbolBodyRefusesUnanalysedPaths below is the 145 regression: the
// analyser Output, not the disk, decides what is readable.
package analysis

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const symbolBodyFixture = `package fixture

import (
	"errors"
	"fmt"
)

// ErrFixture is a package-level var (bugs_open/223 phase 2 records these).
var ErrFixture = errors.New("fixture")

// MaxFixture is a package-level const.
const MaxFixture = 42

// Greeter does things.
type Greeter struct {
	Name string
}

// Hello greets by name.
func Hello(name string) string {
	return fmt.Sprintf("hi %s", name)
}

func (g Greeter) Greet() string {
	return "hello " + g.Name
}

// Greet on a different receiver, SAME method name (collision on purpose).
func (h *Helper) Greet() string {
	return "x"
}

type Helper struct{}
`

func writeSymbolBodyFixture(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "f.go"), []byte(symbolBodyFixture), 0o644); err != nil {
		t.Fatal(err)
	}
	return dir
}

func TestReadSymbolBody(t *testing.T) {
	dir := writeSymbolBodyFixture(t)
	out, err := Analyse(dir)
	if err != nil {
		t.Fatalf("Analyse: %v", err)
	}

	get := func(symbol string) string {
		t.Helper()
		body, err := ReadSymbolBody(dir, out, symbol)
		if err != nil {
			t.Fatalf("ReadSymbolBody(%q): %v", symbol, err)
		}
		return body
	}

	// Plain func: starts at `func Hello`, ends at its `}`, NO doc comment.
	hello := get("f.go:Hello")
	if !strings.HasPrefix(hello, "func Hello(name string) string {") {
		t.Errorf("Hello should start at the func line, got:\n%s", hello)
	}
	if strings.Contains(hello, "Hello greets") {
		t.Errorf("Hello must NOT include its doc comment:\n%s", hello)
	}
	if !strings.HasSuffix(strings.TrimRight(hello, "\n"), "}") {
		t.Errorf("Hello should end at the closing brace:\n%s", hello)
	}

	// Type.
	greeter := get("f.go:Greeter")
	if !strings.HasPrefix(greeter, "type Greeter struct {") || !strings.Contains(greeter, "Name string") {
		t.Errorf("Greeter body wrong:\n%s", greeter)
	}
	if strings.Contains(greeter, "Greeter does things") {
		t.Errorf("Greeter must NOT include its doc comment:\n%s", greeter)
	}

	// Method-name collision resolved by receiver-qualified form.
	if g := get("f.go:Greeter.Greet"); !strings.Contains(g, `"hello " + g.Name`) {
		t.Errorf("Greeter.Greet should resolve to the Greeter receiver:\n%s", g)
	}
	if h := get("f.go:Helper.Greet"); !strings.Contains(h, `return "x"`) {
		t.Errorf("Helper.Greet should resolve to the Helper receiver (pointer):\n%s", h)
	}

	// Whole file (no ":Symbol").
	whole := get("f.go")
	if !strings.Contains(whole, "package fixture") || !strings.Contains(whole, "type Helper struct{}") {
		t.Errorf("whole-file read should return the entire file:\n%s", whole)
	}

	// Errors, not panics.
	if _, err := ReadSymbolBody(dir, out, "f.go:Nope"); err == nil {
		t.Error("expected error for unknown symbol")
	}
	if _, err := ReadSymbolBody(dir, out, "nofile.go:Hello"); err == nil {
		t.Error("expected error for unknown path")
	}
}

// TestReadSymbolBodyRefusesUnanalysedPaths is the bugs_open/145 regression: the
// analyser Output, not the disk, decides what ReadSymbolBody will read.
//
// Every case here writes a file that DOES exist on disk under `root`, so a pass
// cannot come from the read merely failing — the refusal has to come from the
// membership check. That is the "assert the mechanism FIRED" shape: the positive
// control at the end proves the same fixture is readable through the analysed
// path, so a green run cannot mean the function is simply refusing everything.
func TestReadSymbolBodyRefusesUnanalysedPaths(t *testing.T) {
	// Two levels on purpose: `root` is the analysed checkout, and `base` is its
	// parent, so the traversal cases below can name a file that REALLY EXISTS
	// outside root. Measured, not assumed — against the pre-fix code
	// filepath.Join(root, "../outside.go") resolved and the file came back.
	base := t.TempDir()
	dir := filepath.Join(base, "repo")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}

	// The analysed file — the positive control's target.
	if err := os.WriteFile(filepath.Join(dir, "f.go"), []byte(symbolBodyFixture), 0o644); err != nil {
		t.Fatal(err)
	}

	// Files the analyser will NOT put in Output, each for a different reason, and
	// each with a distinctive marker so a leak is unmistakable in the assertion.
	const secretMarker = "SUPER_SECRET_VALUE_145"

	// Outside the analysed root entirely — the traversal target.
	if err := os.WriteFile(filepath.Join(base, "outside.go"), []byte("package outside\n\n// "+secretMarker+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	unanalysed := map[string]string{
		// 1. Not Go at all — the class the filing named (docs, .env-shaped files).
		"notes.md": "# notes\n" + secretMarker + "\n",
		".env":     "API_TOKEN=" + secretMarker + "\n",
		// 2. Go, but excluded by the analyser's own rules — the cases a plain
		//    ".go extension" check would have WRONGLY admitted. This is the
		//    measured reason Output membership beats an extension allow-list.
		"f_test.go":          "package fixture\n\n// " + secretMarker + "\n",
		"vendor/dep/dep.go":  "package dep\n\n// " + secretMarker + "\n",
		"testdata/sample.go": "package sample\n\n// " + secretMarker + "\n",
	}
	for rel, body := range unanalysed {
		abs := filepath.Join(dir, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(abs), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(abs, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	out, err := Analyse(dir)
	if err != nil {
		t.Fatalf("Analyse: %v", err)
	}

	// Guard the premise: if the analyser ever starts including these, this test is
	// asserting nothing and must be rewritten rather than left quietly passing.
	for rel := range unanalysed {
		if findFile(out, rel) != nil {
			t.Fatalf("premise broken: analyser now includes %q in Output; "+
				"this test no longer exercises the boundary", rel)
		}
	}
	if findFile(out, "f.go") == nil {
		t.Fatal("premise broken: analyser did not include f.go, so the positive control is meaningless")
	}

	// The bare-path (whole-file) branch must refuse all of them.
	for rel := range unanalysed {
		body, err := ReadSymbolBody(dir, out, rel)
		if err == nil {
			t.Errorf("ReadSymbolBody(%q) returned %d bytes; a file the analyser did not parse must be refused", rel, len(body))
		}
		if strings.Contains(body, secretMarker) {
			t.Errorf("ReadSymbolBody(%q) LEAKED unanalysed file contents", rel)
		}
	}

	// Directory traversal is refused by the same single check — no path outside
	// the analysed set can be named, so filepath.Join cannot be walked out of root.
	// "../outside.go" is the LIVE case (that file exists and leaked before the fix);
	// the absolute path is belt-and-braces for a differently-shaped escape.
	for _, esc := range []string{"../outside.go", "/etc/passwd"} {
		body, err := ReadSymbolBody(dir, out, esc)
		if err == nil {
			t.Errorf("ReadSymbolBody(%q) returned %d bytes; must refuse a path outside the analysis", esc, len(body))
		}
		if strings.Contains(body, secretMarker) {
			t.Errorf("ReadSymbolBody(%q) LEAKED a file from outside the analysed root", esc)
		}
	}

	// POSITIVE CONTROL — the mechanism is selective, not blanket. Both branches
	// still work for the analysed file, so the refusals above are the boundary
	// doing its job rather than the function being broken.
	whole, err := ReadSymbolBody(dir, out, "f.go")
	if err != nil {
		t.Fatalf("positive control: whole-file read of an ANALYSED file must still work: %v", err)
	}
	if !strings.Contains(whole, "package fixture") || !strings.Contains(whole, "type Helper struct{}") {
		t.Errorf("positive control: whole-file read did not return the whole file:\n%s", whole)
	}
	sym, err := ReadSymbolBody(dir, out, "f.go:Hello")
	if err != nil {
		t.Fatalf("positive control: symbol read of an ANALYSED file must still work: %v", err)
	}
	if !strings.HasPrefix(sym, "func Hello(name string) string {") {
		t.Errorf("positive control: symbol slice changed:\n%s", sym)
	}
}

// TestReadSymbolBodyResolvesIndexSpellings is the bugs_open/261 regression: the
// symbol spellings the CODE INDEX produces must resolve here, because the index's
// spelling is what reaches this function.
//
// The two producers of a "path:Symbol" handle had drifted apart. code_symbols
// writes a method as `(` + Receiver.Type + `).` + Name
// (code_symbols_actions.go:598 — `(*SagaCoordinator).applyResponseToState`), and
// diagnose_assemble_bundle's scopeFromCodeResults concatenates that column
// verbatim into the scope; spanOf split on the last dot and compared the raw
// prefix against receiverType(), so `(*SagaCoordinator)` was tested against
// `SagaCoordinator` and never matched. Package-level values were a second, plainer
// gap: the analyser has recorded them since bugs_open/223 phase 2 and spanOf
// searched only Functions and Types.
//
// Measured before the fix, fleet-wide over all 460 bundles ever assembled: 321
// symbol-read failures, of which 301 were the receiver form and 20 were
// package-level values. NOT ONE was a genuinely absent symbol — so this is the
// whole of what that error message has ever meant.
//
// Each case below is a spelling some real producer emits, and the collision pair
// is what stops the fix being "ignore everything before the dot": Greeter.Greet
// and Helper.Greet must still resolve to DIFFERENT bodies.
func TestReadSymbolBodyResolvesIndexSpellings(t *testing.T) {
	dir := writeSymbolBodyFixture(t)
	out, err := Analyse(dir)
	if err != nil {
		t.Fatalf("Analyse: %v", err)
	}

	// Guard the premise for the values half: if the analyser ever stops recording
	// package-level values, the value cases below would pass or fail for reasons
	// that have nothing to do with spanOf, so fail loudly instead.
	fi := findFile(out, "f.go")
	if fi == nil {
		t.Fatal("premise broken: f.go not in Output")
	}
	if len(fi.Values) == 0 {
		t.Fatal("premise broken: analyser recorded no package-level values, so the var/const cases assert nothing")
	}

	cases := []struct {
		symbol string // as some producer really spells it
		want   string // a fragment unique to the RIGHT body
		why    string
	}{
		{"f.go:(*Helper).Greet", `return "x"`, "index form, pointer receiver — code_symbols_actions.go:598"},
		{"f.go:(Greeter).Greet", `"hello " + g.Name`, "index form, value receiver"},
		{"f.go:Greeter.Greet", `"hello " + g.Name`, "dotted form — unchanged, the pre-existing convention"},
		{"f.go:Helper.Greet", `return "x"`, "dotted form, pointer receiver — unchanged"},
		{"f.go:*Helper.Greet", `return "x"`, "starred-dotted, as a human writes it from a call site"},
		{"f.go:ErrFixture", "errors.New", "package-level var — bugs_open/223 phase 2 kinds"},
		{"f.go:MaxFixture", "= 42", "package-level const"},
	}
	for _, c := range cases {
		body, err := ReadSymbolBody(dir, out, c.symbol)
		if err != nil {
			t.Errorf("ReadSymbolBody(%q) [%s]: %v", c.symbol, c.why, err)
			continue
		}
		if !strings.Contains(body, c.want) {
			t.Errorf("ReadSymbolBody(%q) [%s] resolved to the WRONG body — want a fragment %q, got:\n%s",
				c.symbol, c.why, c.want, body)
		}
	}

	// NEGATIVE CONTROLS — the widening must not turn spanOf into "match anything".
	// Without these, a fix that ignored the receiver entirely would pass every case
	// above while silently returning the first Greet it found.
	for _, bad := range []string{
		"f.go:(*Nope).Greet",  // right method name, receiver that does not exist
		"f.go:(*Helper).Nope", // right receiver, method that does not exist
		"f.go:Nope",           // nothing of that name at all
		"f.go:(*Helper).Name", // a struct FIELD is not a symbol
	} {
		if body, err := ReadSymbolBody(dir, out, bad); err == nil {
			t.Errorf("ReadSymbolBody(%q) should not resolve, got %d bytes:\n%s", bad, len(body), body)
		}
	}
}
