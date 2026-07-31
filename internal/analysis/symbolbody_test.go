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

import "fmt"

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
