// FILE: internal/analysis/handle_producer_singleton_test.go
//
// RFC_027, owner ruling 2026-08-15: the `path:Symbol` handle grammar gets ONE
// authoritative producer. This test is what keeps it at one.
//
// The grammar's two halves — the writer that spells a method handle and the
// reader that parses it — must agree byte for byte, or a handle resolves to
// nothing or, worse, to the WRONG body. bugs_closed/261 measured the cost when
// they disagreed: 301 unreadable bodies across 44 diagnosis runs, and the loop
// reported "symbol not found", which every reader took at face value.
//
// Four bugs (189, 261, 267, 269) were four defects in that one mechanism. Each
// was closed as a point fix; the ruling collapsed the last independent producer
// (code_symbols_actions.go, the live code_symbols row writer) onto
// CanonicalSymbolName. Reintroducing a hand-rolled copy is the drift this test
// exists to fail on.
//
// It walks the AST rather than the file text on purpose: a text scan for
// "Receiver.Type" would make every COMMENT mentioning the field load-bearing,
// and this package's comments discuss the spelling at length.
package analysis

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// repoRoot walks up from this package to the module root.
func repoRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	for i := 0; i < 8; i++ {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		dir = filepath.Dir(dir)
	}
	t.Fatal("could not locate the module root (no go.mod found walking up)")
	return ""
}

// isReceiverTypeAccess reports whether n is the expression `<x>.Receiver.Type`.
func isReceiverTypeAccess(n ast.Node) bool {
	outer, ok := n.(*ast.SelectorExpr)
	if !ok || outer.Sel == nil || outer.Sel.Name != "Type" {
		return false
	}
	inner, ok := outer.X.(*ast.SelectorExpr)
	return ok && inner.Sel != nil && inner.Sel.Name == "Receiver"
}

// liveSourceRoots are the trees that actually build into the fleet's binaries —
// the same scope RFC_027 §3 used for its own census. docs/ and scripts/ hold
// archived copies of this very code (an older analyser, an older
// code_symbols_actions) which are not compiled, not callable, and would make
// this test permanently red without saying anything true about the estate.
var liveSourceRoots = []string{"internal", "platform", "pkg", "cmd"}

func TestOnlyThisPackageBuildsAMethodHandle(t *testing.T) {
	root := repoRoot(t)
	owner := filepath.Join("internal", "analysis")

	var offenders []string
	walk := func(tree string) error {
		return filepath.Walk(filepath.Join(root, tree), func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return nil // an unreadable path is not this test's business
			}
			if info.IsDir() {
				switch info.Name() {
				case ".git", "vendor", "node_modules":
					return filepath.SkipDir
				}
				return nil
			}
			if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
				return nil
			}

			rel, relErr := filepath.Rel(root, path)
			if relErr != nil || strings.HasPrefix(rel, owner) {
				return nil // the owning package IS the one permitted producer
			}

			fset := token.NewFileSet()
			file, parseErr := parser.ParseFile(fset, path, nil, 0) // no comments parsed
			if parseErr != nil {
				return nil // not compilable Go; the build catches that, not this test
			}
			ast.Inspect(file, func(n ast.Node) bool {
				if isReceiverTypeAccess(n) {
					pos := fset.Position(n.Pos())
					offenders = append(offenders, fmt.Sprintf("%s:%d", rel, pos.Line))
				}
				return true
			})
			return nil
		})
	}

	for _, tree := range liveSourceRoots {
		if err := walk(tree); err != nil {
			t.Fatalf("walk %s: %v", tree, err)
		}
	}

	if len(offenders) > 0 {
		t.Errorf("a method handle is being built outside internal/analysis, at %v.\n"+
			"Call analysis.CanonicalSymbolName(fn) instead. Two producers of this grammar is "+
			"what bugs_closed/261 cost 301 unreadable bodies (RFC_027, owner ruling 2026-08-15).",
			offenders)
	}
}
