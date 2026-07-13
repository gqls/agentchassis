// internal/analysis/symbolbody_test.go  (package analysis)
//
// Behaviour test: analyse a small fixture, then assert ReadSymbolBody slices the
// analyser's spans the way cmd/bundle does — func body from the `func` line to
// its closing brace (NO doc comment), types whole, receiver-qualified methods
// disambiguated, whole-file when no symbol is given, and clean errors (not
// panics) for unknown path/symbol.
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
