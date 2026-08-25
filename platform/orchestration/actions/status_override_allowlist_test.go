package actions

import (
	"go/ast"
	"go/parser"
	"go/token"
	"testing"
)

// bugs_open/396. `status_override` was a bare step-config string written straight
// into site_work_items.status with no allow-list. A value outside the three
// predicates that decide a row's fate — dispatch (`triaged`/`approved`), the
// promoter (`detected`), and idx_swi_dedup's exclusion set — leaves the row
// undispatchable, un-promotable AND still holding its (site_id,item_key) slot,
// with every field on it looking healthy.
//
// THESE TESTS ARE WRITTEN TO FAIL ON THE MUTATION THAT MATTERS, which is not
// "someone deletes the guard" but "someone adds `deferred` to the list because
// it is the natural word for park this".

// TestStatusOverrideRefusesDeferred is the motivating case, pinned by name.
//
// MUTATION: add "deferred" to workItemStatusOverrideAllowed and this fails.
func TestStatusOverrideRefusesDeferred(t *testing.T) {
	if statusOverrideAllowed("deferred") {
		t.Fatal("`deferred` is allowed as a status_override. It is the ONE status that is " +
			"neither claimable (claim_work_item_action.go takes triaged/approved), nor " +
			"promotable (the promoter takes detected), nor terminal (it is absent from " +
			"idx_swi_dedup's exclusion list, so the row keeps its dedup slot and the " +
			"detector cannot re-file). That combination is bugs_open/396. " +
			"To park work deliberately use park_work_items() — migration 621, WII-034.")
	}
}

// TestStatusOverrideEveryEntryHasAnExit is the general form: a status may be
// written here only if the row can still LEAVE it. Terminal statuses qualify
// because they release the dedup slot; non-terminal ones qualify only if
// something named moves them on.
//
// MUTATION: add any status without an exit — 'diagnosing', 'claimed', a typo —
// and this fails, naming it.
func TestStatusOverrideEveryEntryHasAnExit(t *testing.T) {
	terminal := map[string]bool{}
	for _, s := range workItemTerminalStatuses {
		terminal[s] = true
	}

	// Non-terminal statuses that are nevertheless legitimate here, each with the
	// live consumer that gets the row out again. If you add to this map you are
	// asserting that consumer exists — go and read it first.
	namedConsumer := map[string]string{
		"needs_human_review": "HandleRetryWorkItem and the admin resolve endpoint both select it",
		"blocked":            "the `feasibility-recheck` scheduled task selects it",
	}

	if len(workItemStatusOverrideAllowed) == 0 {
		t.Fatal("the allow-list is empty — that is not a narrowing, it is a disabled feature; " +
			"needs_human_review is the only value ever configured fleet-wide and must stay allowed")
	}

	for _, s := range workItemStatusOverrideAllowed {
		if terminal[s] {
			continue
		}
		if _, ok := namedConsumer[s]; ok {
			continue
		}
		t.Errorf("status_override allows %q, which is neither terminal nor has a named consumer. "+
			"A row written to it can never leave: nothing dispatches it, nothing promotes it, "+
			"and it still holds its idx_swi_dedup slot so the detector cannot re-file. "+
			"Either name the consumer that moves it on (and read that code first), or remove it. "+
			"bugs_open/396.", s)
	}
}

// TestStatusOverrideAllowsTheOnlyValueEverConfigured guards the opposite
// direction — an allow-list narrowed until it breaks production.
//
// [MEASURED 2026-08-25] a recursive walk over EVERY agent_definitions row,
// snapshots and soft-deleted included, found status_override on 4 steps in 3
// agents (component-template-fixer ×2, page-build-handler, tool-improver) and
// every one of them is `needs_human_review`. No other value has ever been
// configured, so this is the whole of the live surface.
//
// MUTATION: remove "needs_human_review" from the list and this fails.
func TestStatusOverrideAllowsTheOnlyValueEverConfigured(t *testing.T) {
	if !statusOverrideAllowed("needs_human_review") {
		t.Fatal("needs_human_review is refused, and it is the ONLY value configured " +
			"anywhere in agent_definitions (4 steps, 3 agents, measured 2026-08-25 " +
			"including snapshots and deleted rows). This allow-list would break every " +
			"live HITL refusal step.")
	}
}

// TestStatusOverrideRejectsUnknown pins the default direction: unknown is
// refused, not waved through.
func TestStatusOverrideRejectsUnknown(t *testing.T) {
	for _, s := range []string{"", "parked", "on_hold", "TRIAGED", "needs-human-review", "held"} {
		if statusOverrideAllowed(s) {
			t.Errorf("statusOverrideAllowed(%q) = true; unknown values must be refused so the "+
				"item falls through to the failure ladder rather than being stranded", s)
		}
	}
}

// TestFailWorkItemActionCallsTheGuard proves the allow-list is actually WIRED,
// not merely defined.
//
// ⚠ IT PARSES THE AST RATHER THAN GREPPING THE SOURCE, deliberately. A
// source-scan test makes your own comments load-bearing: the needle matches the
// comment that explains the guard and the test passes vacuously even if the call
// is deleted. The AST contains no comments, so this can only pass on real code.
//
// MUTATION: delete the `!statusOverrideAllowed(statusOverride)` call from
// FailWorkItemAction and this fails, even with every comment left in place.
func TestFailWorkItemActionCallsTheGuard(t *testing.T) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "load_work_item_actions.go", nil, 0) // 0 = drop comments
	if err != nil {
		t.Fatalf("parse load_work_item_actions.go: %v", err)
	}

	var fn *ast.FuncDecl
	for _, d := range file.Decls {
		if f, ok := d.(*ast.FuncDecl); ok && f.Name.Name == "FailWorkItemAction" {
			fn = f
			break
		}
	}
	if fn == nil {
		t.Fatal("FailWorkItemAction not found in load_work_item_actions.go")
	}

	found := false
	ast.Inspect(fn, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		if id, ok := call.Fun.(*ast.Ident); ok && id.Name == "statusOverrideAllowed" {
			found = true
			return false
		}
		return true
	})

	if !found {
		t.Fatal("FailWorkItemAction does not call statusOverrideAllowed. The allow-list is " +
			"defined but not wired, so status_override is honoured verbatim again and " +
			"bugs_open/396's black hole is reachable from any step config.")
	}
}
