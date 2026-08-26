package actions

import (
	"go/ast"
	"go/parser"
	"go/token"
	"strings"
	"testing"
)

// bugs_open/413 — the dispatch selector and the item loader are ONE ordering contract.
//
// build-pipeline-trigger > find_dispatchable_site (migration 657) ranks sites by
// min(created_at) over each site's top-K eligible rows UNDER THIS FILE'S ORDERING
// (`ORDER BY wi.priority ASC, wi.created_at ASC`), K read live from
// build-dispatch-loop > load_items > max_items. That is what makes a "pin"
// unrepresentable: a site can only win selection on work this loader will actually
// take. Before 657 the selector ranked sites by their single globally-oldest row,
// and one old worst-priority row could win selection for its site for ever while
// never being loaded — 16 of 25 eligible sites were pinned when measured
// (2026-08-26 ~20:1xZ), with waits beyond 11 h invisible to every aggregate meter.
//
// ⚠ CHANGING THE ORDER BY BELOW THEREFORE CHANGES SITE SELECTION FLEET-WIDE, even
// though this function never selects a site. If you change it, you MUST re-derive
// the selector's window in the same task: the live query's window clause must
// mirror the new ordering, and 657's _VERIFY sidecar
// (docs/agent_docs/sql_for_agents/657_selector_ranks_sites_by_loadable_work_HOLD_VERIFY.sql,
// assertion 2) pins the DB-side half of this lockstep. This test is the Go-side half.
//
// ⚠ AST, not grep: a source scan would match this very comment and pass vacuously
// (a-source-scanning-test-makes-comments-load-bearing). ParseFile drops comments.

const loaderOrderingFragment = "ORDER BY wi.priority ASC, wi.created_at ASC"

// TestLoadWorkItemsOrderingMirrorsTheSelectorWindow pins the loader's ordering
// literal inside LoadWorkItemsAction.
//
// MUTATION 1: swap to created_at-major (or add DESC) -> FAILS (a literal with
// "ORDER BY" no longer carries the pinned fragment).
// MUTATION 2: delete the ORDER BY entirely -> FAILS (no literal carries it; the
// loader would return an arbitrary K-subset while the selector still ranks sites
// by a specific one).
// MUTATION 3: split the LIMIT off into a separate appended literal -> FAILS (the
// window and the cap must stay one statement fragment: K = max_items bounds THIS
// ordering, not some later re-sort).
func TestLoadWorkItemsOrderingMirrorsTheSelectorWindow(t *testing.T) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "load_work_item_actions.go", nil, 0) // 0 = drop comments
	if err != nil {
		t.Fatalf("parse: %v", err)
	}

	var fn *ast.FuncDecl
	for _, d := range file.Decls {
		if f, ok := d.(*ast.FuncDecl); ok && f.Name.Name == "LoadWorkItemsAction" {
			fn = f
			break
		}
	}
	if fn == nil {
		t.Fatal("LoadWorkItemsAction not found")
	}

	var orderByLits []string
	ast.Inspect(fn, func(n ast.Node) bool {
		if lit, ok := n.(*ast.BasicLit); ok && lit.Kind == token.STRING &&
			strings.Contains(lit.Value, "ORDER BY") {
			orderByLits = append(orderByLits, lit.Value)
		}
		return true
	})

	if len(orderByLits) == 0 {
		t.Fatal("LoadWorkItemsAction no longer carries an ORDER BY literal. Without it the " +
			"loader returns an arbitrary K-subset, while find_dispatchable_site (migration " +
			"657) still ranks sites by a window that claims to mirror this ordering — the " +
			"selector would promise work the loader does not deliver, which is bugs_open/413's " +
			"mechanism wearing a new face.")
	}
	if len(orderByLits) > 1 {
		t.Errorf("LoadWorkItemsAction carries %d ORDER BY literals; expected exactly one. "+
			"A second ordering means the selector's window (which mirrors ONE ordering) no "+
			"longer describes what a pick will load.", len(orderByLits))
	}

	lit := orderByLits[0]
	if !strings.Contains(lit, loaderOrderingFragment) {
		t.Errorf("the loader's ordering literal changed: got %s, want it to contain %q. "+
			"This ordering is HALF A CONTRACT: find_dispatchable_site ranks sites by "+
			"min(created_at) over each site's top-K rows under exactly this ordering "+
			"(bugs_open/413, migration 657). Re-derive the selector window in the same task "+
			"and update 657's _VERIFY (assertion 2) — then, and only then, update this test.",
			lit, loaderOrderingFragment)
	}
	if !strings.Contains(lit, "LIMIT $") {
		t.Errorf("the ordering literal no longer carries its LIMIT placeholder (got %s). "+
			"max_items must cap THIS ordering in the same fragment — a detached LIMIT can "+
			"drift to capping a different sort, and K = max_items is what the selector's "+
			"window reads live (657's K agreement with 658).", lit)
	}
}
