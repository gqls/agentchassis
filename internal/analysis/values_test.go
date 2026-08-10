package analysis

import (
	"go/ast"
	"go/parser"
	"go/token"
	"strings"
	"testing"
)

// bugs_open/223 phase 2 — the analyser dropped every package-level var and const
// at the AST walk, so code_symbols could show a reader every USE of a declaration
// and never the declaration itself. Two consumers were stopped by that: the
// landmine-verifier answered "possibly inlined or renamed" about a live `var`, and
// a 090 run stopped at UNVERIFIABLE naming the spec whose Defaults map it could
// not see (bugs_open/231).
//
// The LINE SPAN is what these tests care most about. For a value the body IS the
// evidence, so a span that captured only the name would index the identifier and
// lose the thing anyone actually needs.

func parseValues(t *testing.T, src string) []ValueDef {
	t.Helper()
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "x.go", src, parser.ParseComments)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	var out []ValueDef
	for _, decl := range f.Decls {
		d, ok := decl.(*ast.GenDecl)
		if !ok || (d.Tok != token.VAR && d.Tok != token.CONST) {
			continue
		}
		for _, spec := range d.Specs {
			if vs, ok := spec.(*ast.ValueSpec); ok {
				out = append(out, valueDefs(fset, d, vs)...)
			}
		}
	}
	return out
}

func find(t *testing.T, vals []ValueDef, name string) ValueDef {
	t.Helper()
	for _, v := range vals {
		if v.Name == name {
			return v
		}
	}
	t.Fatalf("no ValueDef named %q in %d values", name, len(vals))
	return ValueDef{}
}

func TestValueDefsCoversVarAndConst(t *testing.T) {
	vals := parseValues(t, `package p

// Patterns is the doc line.
var Patterns = []string{"a", "b"}

const MaxTokens = 2048

var unexported int
`)
	if len(vals) != 3 {
		t.Fatalf("want 3 values, got %d: %+v", len(vals), vals)
	}
	p := find(t, vals, "Patterns")
	if p.Kind != "var" || !p.Exported || p.Doc != "Patterns is the doc line." {
		t.Errorf("Patterns: %+v", p)
	}
	if m := find(t, vals, "MaxTokens"); m.Kind != "const" || !m.Exported {
		t.Errorf("MaxTokens: %+v", m)
	}
	if u := find(t, vals, "unexported"); u.Exported || u.Type != "int" {
		t.Errorf("unexported: %+v — a DECLARED type must be captured verbatim", u)
	}
}

// A grouped block is the common shape for a pattern table or a policy list, and
// each member must get its OWN span — slicing the decl span per member would
// repeat the entire block once per name.
func TestValueDefsSpansAreSpecLevelInsideABlock(t *testing.T) {
	vals := parseValues(t, `package p

const (
	// First is one.
	First = "one"

	Second = "two"
)
`)
	first, second := find(t, vals, "First"), find(t, vals, "Second")
	if first.StartLine == second.StartLine {
		t.Fatalf("members of a block must have distinct spans, both start at %d", first.StartLine)
	}
	if first.EndLine >= second.StartLine {
		t.Errorf("First's span (%d-%d) must not swallow Second (starts %d)", first.StartLine, first.EndLine, second.StartLine)
	}
	if first.Doc != "First is one." {
		t.Errorf("a doc comment INSIDE a block sits on the spec, got %q", first.Doc)
	}
}

// A lone declaration takes the DECL span, so the keyword and its doc comment are
// inside the slice — the body reads as source rather than as a fragment.
func TestValueDefsLoneDeclarationIncludesTheKeyword(t *testing.T) {
	src := `package p

// Table explains itself.
var Table = map[string]int{
	"a": 1,
	"b": 2,
}
`
	v := find(t, parseValues(t, src), "Table")
	lines := strings.Split(src, "\n")
	body := strings.Join(lines[v.StartLine-1:v.EndLine], "\n")
	if !strings.HasPrefix(strings.TrimSpace(body), "var Table") {
		t.Errorf("the slice must start at the keyword, got:\n%s", body)
	}
	if !strings.Contains(body, `"b": 2`) {
		t.Errorf("THE BODY IS THE EVIDENCE — the whole literal must be inside the span, got:\n%s", body)
	}
}

// The blank identifier must never be indexed: `var _ Iface = (*T)(nil)` appears
// several times in one file and code_symbols' identity is (repo, path, symbol),
// so every `_` in a file collides on UPSERT and all but one silently vanish.
func TestValueDefsSkipsBlankIdentifier(t *testing.T) {
	vals := parseValues(t, `package p

var _ Stringer = (*T)(nil)
var _ Marshaler = (*T)(nil)
var Real = 1
`)
	if len(vals) != 1 || vals[0].Name != "Real" {
		t.Fatalf("only Real should be indexed, got %+v", vals)
	}
}

// `var a, b = f()` is TWO addressable symbols; a lookup for b must find it.
func TestValueDefsOneEntryPerName(t *testing.T) {
	vals := parseValues(t, `package p

var alpha, beta = compute()
`)
	if len(vals) != 2 {
		t.Fatalf("want 2 entries for a two-name spec, got %d: %+v", len(vals), vals)
	}
	if find(t, vals, "alpha").StartLine != find(t, vals, "beta").StartLine {
		t.Error("names sharing one declaration must share its span — the declaration IS one region")
	}
}

// An inferred type must stay EMPTY rather than be guessed. This package parses; it
// does not type-check, and a guessed type rendered into a signature would be a
// confident wrong answer of exactly the kind bugs_open/223 is about.
func TestValueDefsDoesNotInferAType(t *testing.T) {
	if v := find(t, parseValues(t, "package p\n\nvar x = compute()\n"), "x"); v.Type != "" {
		t.Errorf("an inferred type must be empty, got %q", v.Type)
	}
}
