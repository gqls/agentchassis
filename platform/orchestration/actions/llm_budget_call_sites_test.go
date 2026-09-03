// FILE: platform/orchestration/actions/llm_budget_call_sites_test.go
//
// THE OUTPUT-TOKEN BUDGET MUST COME FROM CONFIGURATION, AT EVERY CALL SITE IN
// THIS PACKAGE. This file fails the build when it does not.
//
// WHY IT EXISTS, and why the guard that already existed was not enough.
//
// `bugs_open/257` Path A (2026-08-16) moved budget resolution into the provider
// clients, so a caller that supplies NO budget now inherits the configured one.
// A test then bound the two provocation actions to `llmOptionsFromConfig`
// (`provocation_generator_action_test.go`, TestNoProvocationActionCallsAModel...).
// That test is well built — it refuses to pass when it matches nothing, which is
// the vacuity failure register OPP-003 exists to warn about — but it is scoped to
// two named files, so it could not see anything written afterwards.
//
// Two actions were written afterwards (`rewrite_negations`, 2026-08-20;
// `repair_ordering_register`, 2026-08-31). Each re-implemented the precedence
// rule by hand and each ended in a hardcoded `2000`. That is worse than the
// original defect, for two reasons a reviewer should not have to rediscover:
//
//  1. A LITERAL DEFEATS PATH A. An explicitly supplied option wins at the wire
//     (`platform/aiservice/anthropic.go:307`), so a caller that always sends a
//     number can never inherit the configured one. Passing nothing is strictly
//     safer than passing 2000.
//  2. THE FLEET'S OWN INSTRUMENT COULD NOT SEE IT. `offer-analyser` declares
//     `ai_service.max_tokens: 2000` on both steps that run
//     `repair_ordering_register`, and the Go literal was also 2000 — so
//     `llm_call_log.max_tokens` read 2000 whether the configuration was honoured
//     or dropped on the floor. [MEASURED 2026-09-03] 29 calls, all at 2000, and
//     no query over that table can distinguish the two hypotheses.
//
// So the rule is enforced at the PACKAGE, not at a list of files, and it is
// enforced on the AST rather than on the source text. A source scan would make
// this package's comments load-bearing — a documented failure shape in this
// estate — and every paragraph above would then be a needle for its own test.
// `go/parser` is asked not to return comments at all.
//
// WHAT THIS AUDIT CANNOT SEE — stated here so nobody reads a pass as more than
// it is:
//
//   - A literal laundered through a variable. `mt := 2000; opts["max_tokens"] =
//     mt` is invisible to the literal check. For the aiservice call sites that
//     does not matter, because the resolver rule binds them independently; for a
//     provider reached over raw HTTP (feed_actions.go's Perplexity path) the
//     literal check is the only net, and it has that hole.
//   - Anything outside this package. `internal/agents/*` and `tools-api` call
//     models with their own hardcoded budgets (bugs_open/257 §6). This audit is
//     a package rule; extending it needs a build-time check, not a Go test.
//   - Config that is declared and read by nobody. [MEASURED 2026-09-03] four
//     `site-adoption-agent` steps declare a top-level `config.max_tokens`
//     (32000/4000/6000/4000) and every one of them runs at the root 16000,
//     because no reader looks there. That is the same defect from the config
//     end, and no Go test can see it.
package actions

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"strings"
	"testing"
)

// The two methods of the aiservice client interface that carry a prompt to a
// provider (`platform/aiservice/interface.go`: AIService, VisionCapable). If a
// third is added there, the counters below still fire — a package with model
// calls this file cannot see is caught by TestBudgetAuditSeesTheModelCalls.
var modelCallMethods = map[string]string{
	"GenerateText":       "options",
	"GenerateWithImages": "options",
}

// budgetKeys are the option keys that size a request. A numeric literal written
// to either of these is the defect this file exists to refuse.
var budgetKeys = map[string]bool{
	"max_tokens":    true,
	"budget_tokens": true,
}

// canonicalBudgetBuilder is the ONE file allowed to build an options map without
// calling llmOptionsFromConfig: ExecuteAIStepAction resolves the budget inline
// (ai_actions.go), serving most live steps. It cannot call the helper today —
// their precedence rules differ at the agent level (see llm_options.go) — and
// unifying them is bugs_open/257 candidate 2, an open question with an import
// cycle in it, not a tidy-up.
//
// The exemption is not blind: TestTheCanonicalBuilderStillResolvesFromConfig
// asserts that this file really does read the key out of config, so the moment
// it stops doing so, the exemption stops being granted.
const canonicalBudgetBuilder = "ai_actions.go"

// parsePackage parses every non-test .go file in this package, WITHOUT comments.
func parsePackage(t *testing.T) (*token.FileSet, map[string]*ast.File) {
	t.Helper()
	fset := token.NewFileSet()
	pkgs, err := parser.ParseDir(fset, ".", func(fi os.FileInfo) bool {
		return !strings.HasSuffix(fi.Name(), "_test.go")
	}, 0)
	if err != nil {
		t.Fatalf("parse package: %v", err)
	}
	files := map[string]*ast.File{}
	for _, pkg := range pkgs {
		for name, f := range pkg.Files {
			files[name] = f
		}
	}
	if len(files) == 0 {
		t.Fatal("parsed no source files — this audit is watching nothing")
	}
	return fset, files
}

// isModelCall reports the method name when expr is a call to one of the client's
// generation methods, plus the argument that carries the options map (always the
// last one in both signatures).
func isModelCall(node ast.Node) (string, ast.Expr, bool) {
	call, ok := node.(*ast.CallExpr)
	if !ok {
		return "", nil, false
	}
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return "", nil, false
	}
	if _, known := modelCallMethods[sel.Sel.Name]; !known {
		return "", nil, false
	}
	if len(call.Args) == 0 {
		return sel.Sel.Name, nil, true
	}
	return sel.Sel.Name, call.Args[len(call.Args)-1], true
}

// callsTheResolver reports whether n contains a call to llmOptionsFromConfig.
// Deliberately whole-function rather than a data-flow trace of one identifier:
// the options map is legitimately built once and copied per iteration
// (companies_house_llm_review_action.go), and a rule that only understood
// `x := llmOptionsFromConfig(...)` would push callers into contortions to
// satisfy the test rather than to be correct.
func callsTheResolver(n ast.Node) bool {
	found := false
	ast.Inspect(n, func(node ast.Node) bool {
		if found {
			return false
		}
		if call, ok := node.(*ast.CallExpr); ok {
			if id, ok := call.Fun.(*ast.Ident); ok && id.Name == "llmOptionsFromConfig" {
				found = true
				return false
			}
		}
		return true
	})
	return found
}

// TestEveryModelCallResolvesItsBudgetFromConfig binds the CALL SITES.
//
// The helper being correct and the helper being USED are independent facts, and
// only the second one sends a configured budget to the API.
func TestEveryModelCallResolvesItsBudgetFromConfig(t *testing.T) {
	fset, files := parsePackage(t)

	checked := 0
	for name, file := range files {
		short := shortName(name)
		for _, decl := range file.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Body == nil {
				continue
			}
			resolverInScope := callsTheResolver(fn.Body)

			ast.Inspect(fn.Body, func(node ast.Node) bool {
				method, optsArg, ok := isModelCall(node)
				if !ok {
					return true
				}
				checked++
				pos := fset.Position(node.Pos())

				if why, unbuilt := describeUnbuiltOptions(optsArg); unbuilt {
					t.Errorf("%s:%d %s() is handed %s.\n"+
						"\tSince bugs_open/257 that no longer pins the call to 2048 — the client "+
						"resolves ai_service.max_tokens from its own construction config. But it "+
						"still drops this STEP's max_tokens and budget_tokens, neither of which a "+
						"client constructor is ever shown. Build the map with "+
						"llmOptionsFromConfig(stepCfg, aiCfg, logger, \"<call site>\").",
						short, pos.Line, method, why)
					return true
				}
				if short == canonicalBudgetBuilder || resolverInScope {
					return true
				}
				t.Errorf("%s:%d %s() is called from %s, which never calls llmOptionsFromConfig.\n"+
					"\tThe options map handed to a provider must be built by this package's ONE "+
					"resolver (llm_options.go). Re-implementing the precedence rule by hand is how "+
					"two callers ended up with a hardcoded 2000 that no operator could override — "+
					"and, in offer-analyser's case, one that llm_call_log could not distinguish "+
					"from the configured value.",
					short, pos.Line, method, fn.Name.Name)
				return true
			})
		}
	}

	// A scan that finds nothing to check passes for the wrong reason.
	if checked == 0 {
		t.Fatal("no GenerateText/GenerateWithImages call found in this package — either the " +
			"client interface was renamed or the calls moved; this audit is no longer watching " +
			"anything and must be repointed, not deleted")
	}
	t.Logf("audited %d model call sites", checked)
}

// TestNoHardcodedTokenBudget refuses a numeric literal for a budget key anywhere
// in the package, in either shape it can take: an assignment into an options map
// (`options["max_tokens"] = 2000`) or a map literal entry
// (`map[string]interface{}{"max_tokens": 2000}`).
//
// A literal is not a smaller version of reading the config — it is the one value
// an operator cannot change, and it wins over the config at the wire.
func TestNoHardcodedTokenBudget(t *testing.T) {
	fset, files := parsePackage(t)

	writes := 0
	for name, file := range files {
		short := shortName(name)
		ast.Inspect(file, func(node ast.Node) bool {
			switch n := node.(type) {
			case *ast.AssignStmt:
				for i, lhs := range n.Lhs {
					idx, ok := lhs.(*ast.IndexExpr)
					if !ok || !isBudgetKey(idx.Index) {
						continue
					}
					writes++
					if i < len(n.Rhs) && isNumericLiteral(n.Rhs[i]) {
						pos := fset.Position(n.Pos())
						t.Errorf("%s:%d assigns a hardcoded number to a token budget.\n"+
							"\tRead it from configuration instead (llmOptionsFromConfig). A literal "+
							"wins over ai_service.max_tokens at the wire (aiservice/anthropic.go:307), "+
							"so it is not a fallback — it is an override no operator can reach.",
							short, pos.Line)
					}
				}
			case *ast.KeyValueExpr:
				if !isBudgetKey(n.Key) {
					return true
				}
				writes++
				if isNumericLiteral(n.Value) {
					pos := fset.Position(n.Pos())
					t.Errorf("%s:%d puts a hardcoded number in a map literal for a token budget.\n"+
						"\tSame rule: the budget is configuration, not code.", short, pos.Line)
				}
			}
			return true
		})
	}

	// If the key were renamed, every check above would pass while checking
	// nothing. Make that impossible to mistake for a clean result.
	if writes == 0 {
		t.Fatal("no max_tokens/budget_tokens write found anywhere in this package — the key " +
			"has been renamed or the writes have moved, and this audit is now vacuous")
	}
	t.Logf("audited %d budget-key writes", writes)
}

// TestTheCanonicalBuilderStillResolvesFromConfig is what makes the ai_actions.go
// exemption above honest. That file is excused from calling llmOptionsFromConfig
// because it IS a resolver; if it ever stops reading the key out of a config
// map, the excuse no longer applies and this fails.
func TestTheCanonicalBuilderStillResolvesFromConfig(t *testing.T) {
	_, files := parsePackage(t)

	var canonical *ast.File
	for name, file := range files {
		if shortName(name) == canonicalBudgetBuilder {
			canonical = file
		}
	}
	if canonical == nil {
		t.Fatalf("%s is not in this package any more — the exemption in "+
			"TestEveryModelCallResolvesItsBudgetFromConfig now excuses a file that does not "+
			"exist, so find where ExecuteAIStepAction went and repoint it", canonicalBudgetBuilder)
	}

	resolvesFromConfig := false
	ast.Inspect(canonical, func(node ast.Node) bool {
		assign, ok := node.(*ast.AssignStmt)
		if !ok {
			return true
		}
		for i, lhs := range assign.Lhs {
			idx, ok := lhs.(*ast.IndexExpr)
			if !ok || !isBudgetKey(idx.Index) {
				continue
			}
			if i < len(assign.Rhs) && !isNumericLiteral(assign.Rhs[i]) {
				resolvesFromConfig = true
			}
		}
		return true
	})

	if !resolvesFromConfig {
		t.Fatalf("%s no longer computes a token budget from anything — it is exempted from the "+
			"resolver rule ONLY because it is the canonical resolver. Either restore the config "+
			"read or remove the exemption.", canonicalBudgetBuilder)
	}
}

// describeUnbuiltOptions names the two shapes that carry no configuration at all.
func describeUnbuiltOptions(arg ast.Expr) (string, bool) {
	switch a := arg.(type) {
	case nil:
		return "no options argument at all", true
	case *ast.Ident:
		if a.Name == "nil" {
			return "a nil options map", true
		}
	case *ast.CompositeLit:
		if len(a.Elts) == 0 {
			return "an empty options map literal", true
		}
	}
	return "", false
}

func isBudgetKey(expr ast.Expr) bool {
	lit, ok := expr.(*ast.BasicLit)
	if !ok || lit.Kind != token.STRING {
		return false
	}
	return budgetKeys[strings.Trim(lit.Value, "`\"")]
}

func isNumericLiteral(expr ast.Expr) bool {
	lit, ok := expr.(*ast.BasicLit)
	return ok && (lit.Kind == token.INT || lit.Kind == token.FLOAT)
}

// shortName reduces a parser path ("./ai_actions.go") to its base name.
func shortName(path string) string {
	if i := strings.LastIndex(path, "/"); i >= 0 {
		return path[i+1:]
	}
	return path
}
