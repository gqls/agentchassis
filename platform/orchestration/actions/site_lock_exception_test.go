package actions

import (
	"go/ast"
	"go/parser"
	"go/token"
	"strings"
	"testing"
)

// bugs_open/396, after council ed821065 returned REVISE. `sites.locked_at` already
// existed and is enforced at exactly ONE gate — build-pipeline-trigger's
// find_dispatchable_site. What it lacked was granularity: it is all-or-nothing, so
// a lane needing "hold this site EXCEPT these items" reached for status='deferred'
// instead, and that is what stranded 52 rows.
//
// ⚠ THE FAILURE THESE TESTS EXIST TO PREVENT IS AN UNLOCK, NOT A STRAND.
// find_dispatchable_site selects a SITE. Teach it to select a locked site because
// one excepted item is dispatchable, and LoadWorkItemsAction — which runs next and
// has NEVER checked the lock — takes every dispatchable item on it. The site-gate
// half without the loader half turns a full hold into no hold at all.

// TestSiteLockExceptionSQLKeepsBothHalves pins the two arms of the fragment.
//
// MUTATION 1: delete the `locked_at ... IS NULL` arm -> FAILS (an unlocked site
// would stop yielding work, fleet-wide).
// MUTATION 2: delete the `lock_except_item_ids` arm -> FAILS (the exception list
// would be inert and a locked site would hold everything, which is the old
// behaviour wearing the new column's name).
func TestSiteLockExceptionSQLKeepsBothHalves(t *testing.T) {
	sql := siteLockExceptionSQL()

	if !strings.Contains(sql, "locked_at") || !strings.Contains(sql, "IS NULL") {
		t.Error("the fragment has lost its `locked_at IS NULL` arm: an UNLOCKED site would " +
			"no longer match, so the loader would return nothing for every site in the fleet")
	}
	if !strings.Contains(sql, "lock_except_item_ids") {
		t.Error("the fragment has lost its exception arm: a locked site would hold ALL its " +
			"items, which is the pre-396 behaviour under a new column's name")
	}
	if !strings.Contains(sql, "COALESCE") {
		t.Error("the fragment has lost its COALESCE: a NULL lock_except_item_ids must mean " +
			"THE FULL HOLD (what a lock has always meant), not a NULL comparison")
	}
	if !strings.HasPrefix(strings.TrimSpace(sql), "AND ") {
		t.Errorf("the fragment must start with AND — it is appended to an existing WHERE. Got: %.40q", strings.TrimSpace(sql))
	}
}

// TestSiteLockExceptionSQLBindsOnlyDollarOne is the quiet one that would cost a
// production incident. The fragment is appended to a statement whose caller then
// keeps appending `$2`, `$3` … from its own argIdx counter. If this fragment ever
// introduces a placeholder of its own, every later filter binds the wrong value.
//
// MUTATION: change a `$1` to `$2` -> FAILS.
func TestSiteLockExceptionSQLBindsOnlyDollarOne(t *testing.T) {
	sql := siteLockExceptionSQL()
	for _, bad := range []string{"$2", "$3", "$4", "$5"} {
		if strings.Contains(sql, bad) {
			t.Errorf("the fragment binds %s. It is appended BEFORE the caller's own "+
				"pipeline/handler/limit placeholders, which are numbered from its argIdx "+
				"counter — introducing a placeholder here shifts every one of them and they "+
				"silently bind the wrong values.", bad)
		}
	}
	if !strings.Contains(sql, "$1") {
		t.Error("the fragment no longer references $1 (the site id) — it must scope to the " +
			"site the loader was asked about, or it reads another site's lock")
	}
}

// TestLoadWorkItemsActionIsOptInForTheSiteLock proves the default is OFF.
//
// The key is read with a plain bool assertion, so an absent key yields false and
// the fragment is not appended — leaving the statement byte-identical to the
// pre-396 one. That default is what keeps site-work-orchestrator's two
// load_work_items steps (human-initiated, gating on nothing today) unchanged.
//
// ⚠ AST, not grep: a source scan would match the comment that explains the flag
// and pass even if the read were deleted.
//
// MUTATION: replace `config["honour_site_lock"].(bool)` with a literal `true`
// -> FAILS, because the ValueSpec no longer indexes that key.
func TestLoadWorkItemsActionIsOptInForTheSiteLock(t *testing.T) {
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

	readsKey, callsFragment := false, false
	ast.Inspect(fn, func(n ast.Node) bool {
		if lit, ok := n.(*ast.BasicLit); ok && lit.Kind == token.STRING &&
			strings.Contains(lit.Value, "honour_site_lock") {
			readsKey = true
		}
		if call, ok := n.(*ast.CallExpr); ok {
			if id, ok := call.Fun.(*ast.Ident); ok && id.Name == "siteLockExceptionSQL" {
				callsFragment = true
			}
		}
		return true
	})

	if !readsKey {
		t.Error("LoadWorkItemsAction does not read the `honour_site_lock` step-config key. " +
			"Either the opt-in is gone (and the lock now applies to every caller including " +
			"site-work-arbitrator's human-initiated flow), or it is hard-wired.")
	}
	if !callsFragment {
		t.Error("LoadWorkItemsAction never calls siteLockExceptionSQL. The fragment is defined " +
			"but not wired, so `honour_site_lock: true` would be silently inert — and if the " +
			"held config half (migration 633) is applied against such a binary, a locked site " +
			"with an exception list dispatches its ENTIRE queue.")
	}
}

// TestSiteLockExceptionSQLIsNotTheSelectorSpelling pins the divergence that a
// council reviewer already tried to close in their head (corr 175df761,
// editquality): the same rule is spelled twice, and the two spellings are NOT
// interchangeable.
//
//   - HERE (LoadWorkItemsAction): per-site and PARAMETERISED. The caller has
//     bound the site id as $1, so the lock is a scalar subquery.
//   - THERE (build-pipeline-trigger > find_dispatchable_site, migration 633):
//     a CROSS-SITE SCAN that already joins `sites s` and has NO $1 at all, so it
//     is spelled against the joined alias — `s.locked_at`, `s.lock_except_item_ids`.
//
// Drop this fragment into that query and it references a $1 that does not exist:
// site selection fails fleet-wide, and the build pipeline stops picking any site.
//
// MUTATION: rewrite the fragment against a bare joined alias (drop the
// `(SELECT … WHERE s.id = $1)` subqueries in favour of `s.locked_at`) so that it
// "matches the selector" -> FAILS here, which is the point: making the two the
// same is the change this test exists to refuse.
func TestSiteLockExceptionSQLIsNotTheSelectorSpelling(t *testing.T) {
	sql := siteLockExceptionSQL()

	// The per-site form MUST look the lock up through $1, not through a joined
	// alias — LoadWorkItemsAction's FROM is `site_work_items wi` alone; there is
	// no `sites s` in scope to reference bare.
	if !strings.Contains(sql, "FROM sites s WHERE s.id = $1") {
		t.Error("the fragment no longer scopes the lock lookup by $1. LoadWorkItemsAction's " +
			"FROM is `site_work_items wi` with no `sites` join, so a bare `s.locked_at` here " +
			"does not compile. If this was changed to match find_dispatchable_site's spelling, " +
			"revert it: that query is a cross-site scan with no $1 and the two are not " +
			"interchangeable (council 175df761).")
	}
}
