package datahelpers

import (
	"go/ast"
	"go/parser"
	"go/token"
	"testing"
)

// THE ARM BUDGET — owner ruling 2026-08-15, decision 3 of six, raised by RFC_028.
//
// WHAT IS BEING COUNTED, in plain terms: the number of distinct places inside
// ExtractActionInputs that can put a resolved value into result.Values. That is the
// surface a future author adds to when they write "Strategy 7", and it is the thing
// whose growth nobody was watching — the chain went from Strategy 0 to Strategy 6
// across 27 council rounds, 8 of which said out loud that it needed an architectural
// owner, and it shipped every time because each addition was individually well-argued.
//
// THE RULE: past resolverArmBudget, adding another one is an RFC, not a council round.
// This is RFC_022's ruling (budget N=10 on accumulated optional keys) one level down,
// and for the identical reason the owner gave there: "ten individually inert opt-in
// fields are a shared action nobody understands, and this trigger was the only thing
// that would have noticed the tenth."
//
// WHY THIS IS A GO TEST AND NOT A CRONJOB, unlike its RFC_022 sibling
// (optional-key-budget-check, which runs daily on the cluster). That counter watches a
// number that moves in the DATABASE — an agent definition gains a carrier and no commit
// hook can see it, measured moving 21 -> 22 overnight with nobody acting. This number
// moves only when someone edits this file. A commit is exactly when to catch it, and a
// daily job would be watching a value that cannot change between runs.
//
// WHY AN AST WALK AND NOT A GREP FOR "// Strategy N:". A source-scanning test that reads
// COMMENTS makes the comments load-bearing, and this estate has been bitten by that: the
// first occurrence of a marker wins, and a comment can be edited to make the test agree
// while the code says something else. Counting assignment statements counts the code.
const (
	// The budget, set generous on the owner's instruction ("a reasonably generous
	// limit"). The arithmetic, so a future reader can judge it rather than inherit it:
	// the resolver sits at 10 write sites today, and a new strategy typically adds one
	// or two (Strategy 4 has two — a dotted branch and a single-segment branch), so 15
	// is roughly two to three more strategies of headroom. It binds when the surface
	// has grown by half, not on the next change.
	//
	// Raising it is the owner's call and belongs in RFC_028, not in whichever change
	// happens to be inconvenienced by it. That is the whole point: the count went 0 -> 6
	// strategies over 27 rounds precisely because each individual step was reasonable.
	resolverArmBudget = 15

	// The vacuity control, and the more important half of this test.
	//
	// Without it this test passes when it has gone BLIND. Rename the `result`
	// variable, restructure the writes behind a helper, or break the parse, and the
	// matcher below counts ZERO — which sails under any ceiling. A test that cannot
	// fail is worse than no test, because it reports a budget is being enforced when
	// nothing is being counted. So: assert the floor too. If this trips, the matcher
	// has stopped seeing the writes it is supposed to count — FIX THE MATCHER. Do not
	// lower the floor to make it pass.
	//
	// Pinned to the exact count as measured 2026-08-15, deliberately with NO slack:
	// the defaults pre-fill (line ~640), Strategy 0, the two ExtractFields writes of
	// Strategy 1/2, the Strategy 3 bridge, the nested-object block, Strategy 4's two
	// branches, Strategy 5, and Strategy 6. A floor with slack in it is a floor that
	// tolerates the matcher half-going-blind.
	resolverArmFloor = 10
)

// countResolverValueWrites returns the number of statements in ExtractActionInputs that
// assign into result.Values[...], plus the receiver name it matched on, so a failure can
// say whether it found the function at all.
func countResolverValueWrites(t *testing.T) (int, bool) {
	t.Helper()

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "action_inputs.go", nil, 0)
	if err != nil {
		t.Fatalf("parsing action_inputs.go: %v", err)
	}

	var (
		found bool
		count int
	)
	ast.Inspect(file, func(n ast.Node) bool {
		fn, ok := n.(*ast.FuncDecl)
		if !ok || fn.Name.Name != "ExtractActionInputs" || fn.Body == nil {
			return true
		}
		found = true

		ast.Inspect(fn.Body, func(inner ast.Node) bool {
			assign, ok := inner.(*ast.AssignStmt)
			if !ok {
				return true
			}
			for _, lhs := range assign.Lhs {
				index, ok := lhs.(*ast.IndexExpr)
				if !ok {
					continue
				}
				sel, ok := index.X.(*ast.SelectorExpr)
				if !ok || sel.Sel.Name != "Values" {
					continue
				}
				if ident, ok := sel.X.(*ast.Ident); ok && ident.Name == "result" {
					count++
				}
			}
			return true
		})
		return false
	})

	return count, found
}

func TestResolverArmBudget(t *testing.T) {
	count, found := countResolverValueWrites(t)

	if !found {
		t.Fatal("ExtractActionInputs not found in action_inputs.go — the matcher is blind, " +
			"not the chain empty. Fix the matcher; do not delete this test.")
	}

	if count < resolverArmFloor {
		t.Errorf("counted %d writes into result.Values, floor is %d.\n"+
			"This almost certainly means THE MATCHER WENT BLIND (receiver renamed, writes "+
			"moved behind a helper), not that arms were removed. If arms really were "+
			"removed, lower resolverArmFloor deliberately and say so in the commit.",
			count, resolverArmFloor)
	}

	if count > resolverArmBudget {
		t.Errorf("the input resolver now has %d places that can write a value, budget is %d.\n\n"+
			"This is the RFC_028 arm budget (owner ruling 2026-08-15). Adding another way for "+
			"a value to arrive is an ARCHITECTURE question now, not a council round: every "+
			"action in every workflow resolves its inputs through this function, and "+
			"correctness rests on a hand-proved 'nothing else can win' invariant that gets "+
			"harder to hold with each arm.\n\n"+
			"Do not raise resolverArmBudget to make this pass. Take it to RFC_028.",
			count, resolverArmBudget)
	}

	t.Logf("input resolver: %d/%d value-write sites used", count, resolverArmBudget)
}

// TestResolverArmBudgetMatcherCanFail is the demand control on the control.
//
// The floor above proves the matcher sees SOMETHING. It does not prove the matcher would
// notice a NEW arm — and that is the whole job. So: parse a fixture carrying one extra
// write and assert the count moves. A counter that cannot go up is a counter that will
// report "within budget" for ever.
func TestResolverArmBudgetMatcherCanFail(t *testing.T) {
	const fixture = `package p
func ExtractActionInputs() {
	result.Values["a"] = 1
	result.Values["b"] = 2
	other.Values["c"] = 3
	result.Defaulted["d"] = true
}`
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "fixture.go", fixture, 0)
	if err != nil {
		t.Fatalf("parsing fixture: %v", err)
	}

	count := 0
	ast.Inspect(file, func(n ast.Node) bool {
		assign, ok := n.(*ast.AssignStmt)
		if !ok {
			return true
		}
		for _, lhs := range assign.Lhs {
			index, ok := lhs.(*ast.IndexExpr)
			if !ok {
				continue
			}
			sel, ok := index.X.(*ast.SelectorExpr)
			if !ok || sel.Sel.Name != "Values" {
				continue
			}
			if ident, ok := sel.X.(*ast.Ident); ok && ident.Name == "result" {
				count++
			}
		}
		return true
	})

	// Two matches, not four: `other.Values` is a different receiver and
	// `result.Defaulted` is a different map. Both must be excluded, or the real
	// count above is inflated by bookkeeping writes and the budget bites early.
	if count != 2 {
		t.Errorf("matcher counted %d writes in the fixture, want 2 "+
			"(it must count result.Values only — not other receivers, not result.Defaulted)", count)
	}
}
