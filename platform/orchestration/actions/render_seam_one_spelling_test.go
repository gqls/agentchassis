// FILE: platform/orchestration/actions/render_seam_one_spelling_test.go
//
// THE CHECK THAT KEEPS ONE SPELLING ONE SPELLING (owner ruling 2026-08-21,
// closing RFC_041 §5's open question).
//
// Until 2026-08-21 this package offered two ways to render a component:
// `RenderTemplateReportingMissing`, which returns the empty-placeholder and
// dead-URL-attribute reports, and a one-line `RenderTemplate` wrapper that threw
// both away. Nine of the twelve call sites used the short one — which is how
// bugs_open/238 shipped five `<img src="">` to a live homepage while the call
// had the field names in hand. The discard lived INSIDE the wrapper, where no
// reviewer of the call site could see it.
//
// The wrapper is deleted. A caller that does not want the reports now writes
// `out, _, _, err :=`, so the discard is in the diff. These tests exist because
// nothing else stops someone adding the convenience back in six months — and
// the version that comes back is always reasonable-looking.
//
// WHY AST AND NOT GREP. A source-scanning check that matches text makes your
// COMMENTS load-bearing: the sentence above literally contains
// "RenderTemplateReportingMissing", and a grep-based rule would fail on this
// file's own prose. Parsing declares that away — a comment is not a call
// expression and not a function declaration, so it cannot trip these rules and
// they cannot be worked around by rewording. Test files are skipped: a test may
// legitimately call anything, and including them would make the rules describe
// the tests rather than the package.
package actions

import (
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"sort"
	"strings"
	"testing"
)

// declaredTemplateExecutors enumerates every function in this package that
// EXECUTES a Go template. Not "parses" — executes: `missingBareFields` builds a
// template to WALK it and never runs it, and sweeping parse-only helpers in here
// would pad the list with entries nobody needs to justify, which is how an
// allow-list stops discriminating.
//
// The list is grouped, because the two groups ask different questions.
//
// COMPONENT-HTML RENDERERS — these execute a component's html_template, so each
// one is a DIALECT a template author can land in without knowing:
//
//	executeGoTemplate     — the one true executor (call_agent.go). FuncMap +
//	                        missingkey=zero. Everything RenderTemplate renders
//	                        goes through it.
//	RenderTemplateWithMap — the contact-info block on the legacy whole-page
//	                        rerender path. NO FuncMap, NO missingkey=zero, so
//	                        {{safe}}/{{default}}/{{isset}} are PARSE errors here
//	                        (bugs_closed/260 §13g). Its whole call chain is
//	                        currently unreachable — RerenderSitePagesAction is in
//	                        no GlobalActionRegistry entry — which is the only
//	                        reason the divergence has not bitten.
//	renderBlogTemplate    — the blog listing (rebuild_blog_listing_action.go).
//	                        Same missing FuncMap, and unlike the one above this
//	                        path IS registered and live. Found by this very test
//	                        on 2026-08-21; its silent substitution of a generic
//	                        default listing on a parse failure was removed the
//	                        same day.
//
// NOT COMPONENT RENDERERS — they execute their own small templates from config
// or code, never a component's html_template, so the dialect question does not
// arise. Listed anyway so a NEW one cannot arrive unannounced and so nobody has
// to re-derive this set the way this test just did:
//
//	buildCommitMessage             — git commit message from step config
//	generateStoragePath            — object-storage path from step config
//	extractSearchQuery             — search query from step config
//	evaluateConditionTemplateFormat — workflow condition expression
//	RenderCSSFromSpecAction        — CSS from a design spec
//
// ⚠ IF THIS TEST FAILED AND YOU ARE ABOUT TO ADD AN ENTRY: say which group, and
// if it is the first group, say what its language is and how a template author
// is supposed to know which executor they are writing for. A third dialect is
// how bugs_closed/260 happened twice.
var declaredTemplateExecutors = map[string]string{
	// component-HTML renderers (dialect matters)
	"executeGoTemplate":     "the one true executor: FuncMap + missingkey=zero",
	"RenderTemplateWithMap": "legacy contact-info; NO FuncMap (bugs_closed/260 §13g); call chain unregistered",
	"renderBlogTemplate":    "blog listing; NO FuncMap; LIVE path; silent default-template substitution removed 2026-08-21",
	// not component renderers (their own small templates)
	"buildCommitMessage":              "git commit message from step config",
	"generateStoragePath":             "object-storage path from step config",
	"extractSearchQuery":              "search query from step config",
	"evaluateConditionTemplateFormat": "workflow condition expression",
	"RenderCSSFromSpecAction":         "CSS from a design spec",
	// bugs_open/440 / WII-038. OPERATOR-AUTHORED, which is the one in this group
	// that most needs its language written down: a person writes this template into
	// a step config (`error_message_template` on fail_work_item), so they are the
	// "template author" the group above is about, one seam along.
	// ITS LANGUAGE: plain text/template over the run's collected_data, NO FuncMap,
	// and `missingkey=error` — the OPPOSITE of executeGoTemplate's missingkey=zero.
	// So {{safe}}/{{default}}/{{isset}} are parse errors here (as in
	// RenderTemplateWithMap), AND a path that is merely absent is a RENDER error
	// rather than an empty string. Both are deliberate: this renders a refusal
	// message a human acts on, so failing loudly and falling back to the static
	// error_message beats emitting a confidently wrong one. How an author knows
	// which executor they are writing for: this dialect is reachable from exactly
	// one config key, and the key's own doc header states it.
	"renderFailWorkItemMessage": "fail_work_item refusal message from step config; NO FuncMap; missingkey=ERROR (not zero); <no value> guarded; falls back to the static error_message and files FAIL_WORK_ITEM_MESSAGE_TEMPLATE_FALLBACK",
}

// parsePackageFuncs returns funcName -> body, for non-test files in this package.
func parsePackageFuncs(t *testing.T) (map[string]*ast.FuncDecl, *token.FileSet) {
	t.Helper()
	fset := token.NewFileSet()
	pkgs, err := parser.ParseDir(fset, ".", func(fi fs.FileInfo) bool {
		return !strings.HasSuffix(fi.Name(), "_test.go")
	}, 0)
	if err != nil {
		t.Fatalf("parsing the package: %v", err)
	}
	funcs := map[string]*ast.FuncDecl{}
	for _, pkg := range pkgs {
		for _, f := range pkg.Files {
			for _, decl := range f.Decls {
				if fd, ok := decl.(*ast.FuncDecl); ok && fd.Name != nil && fd.Body != nil {
					funcs[fd.Name.Name] = fd
				}
			}
		}
	}
	if len(funcs) < 100 {
		// A traversal that finds almost nothing would let every rule below pass
		// for the wrong reason. This package has hundreds of functions.
		t.Fatalf("CONTROL FAILED: parsed only %d functions — the traversal is broken, so a green result here would mean nothing", len(funcs))
	}
	return funcs, fset
}

// callsNamed reports whether fn's body calls the named plain function.
func callsNamed(fn *ast.FuncDecl, name string) bool {
	found := false
	ast.Inspect(fn.Body, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		if id, ok := call.Fun.(*ast.Ident); ok && id.Name == name {
			found = true
		}
		return true
	})
	return found
}

// TestOnlyOneExportedRenderTemplateSpelling is the owner's ask, literally: one
// spelling. Any new exported RenderTemplate* symbol has to argue with this test.
func TestOnlyOneExportedRenderTemplateSpelling(t *testing.T) {
	funcs, _ := parsePackageFuncs(t)

	var found []string
	for name, fd := range funcs {
		if fd.Recv != nil || !ast.IsExported(name) {
			continue
		}
		if strings.HasPrefix(name, "RenderTemplate") {
			found = append(found, name)
		}
	}
	sort.Strings(found)

	want := []string{"RenderTemplate", "RenderTemplateWithMap"}
	if strings.Join(found, ",") != strings.Join(want, ",") {
		t.Fatalf("exported RenderTemplate* symbols = %v, want %v.\n"+
			"A NEW one means two spellings again — the shape that let bugs_open/238 discard the\n"+
			"dead-control report at nine call sites invisibly. If the new symbol is a convenience\n"+
			"wrapper, delete it and make the caller write `out, _, _, err :=`; if it is genuinely a\n"+
			"different renderer, it belongs in declaredTemplateExecutors with its dialect stated.",
			found, want)
	}
}

// TestOneEntryPointToTheExecutor is the deeper half: it is not enough that there
// is one NAME, if a second function reaches past it into executeGoTemplate and
// skips the missing-field report, the <no value> strip and the form_action
// seeding that live in the seam.
func TestOneEntryPointToTheExecutor(t *testing.T) {
	funcs, _ := parsePackageFuncs(t)

	var callers []string
	for name, fd := range funcs {
		if callsNamed(fd, "executeGoTemplate") {
			callers = append(callers, name)
		}
	}
	sort.Strings(callers)

	if len(callers) != 1 || callers[0] != "RenderTemplate" {
		t.Fatalf("functions calling executeGoTemplate = %v, want exactly [RenderTemplate].\n"+
			"Everything the seam does AROUND the execution — seeding form_action, the InstanceID\n"+
			"report, missingBareFields, the <no value> strip, and the error that replaced the\n"+
			"regex fallback (bugs_closed/260) — is skipped by a caller that goes straight to the\n"+
			"executor.", callers)
	}
}

// TestTemplateExecutorsAreDeclared catches the drift that actually happened:
// bugs_closed/260 §13g found a SECOND, independent Go-template executor with a
// different language, and nothing anywhere said it existed. Any third one now
// fails the build until it is declared with its dialect.
//
// The comparison is EXACT in both directions on purpose. A stale entry — a
// declared executor that no longer constructs a template — fails too, because an
// allow-list that only ever grows is how the check stops meaning anything.
func TestTemplateExecutorsAreDeclared(t *testing.T) {
	funcs, _ := parsePackageFuncs(t)

	actual := map[string]bool{}
	for name, fd := range funcs {
		var constructs, executes bool
		ast.Inspect(fd.Body, func(n ast.Node) bool {
			call, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}
			sel, ok := call.Fun.(*ast.SelectorExpr)
			if !ok || sel.Sel == nil {
				return true
			}
			if pkg, ok := sel.X.(*ast.Ident); ok && pkg.Name == "template" && sel.Sel.Name == "New" {
				constructs = true
			}
			// Execute / ExecuteTemplate on anything. Narrow enough in this
			// package that a false positive would itself be worth reading.
			if strings.HasPrefix(sel.Sel.Name, "Execute") {
				executes = true
			}
			return true
		})
		// EXECUTES, not merely constructs: missingBareFields builds a template to
		// WALK it and never runs it, and a parse-only helper carries no dialect
		// risk. Requiring both keeps the list to functions that actually render.
		if constructs && executes {
			actual[name] = true
		}
	}

	var undeclared, stale []string
	for name := range actual {
		if _, ok := declaredTemplateExecutors[name]; !ok {
			undeclared = append(undeclared, name)
		}
	}
	for name := range declaredTemplateExecutors {
		if !actual[name] {
			stale = append(stale, name)
		}
	}
	sort.Strings(undeclared)
	sort.Strings(stale)

	if len(undeclared) > 0 {
		t.Errorf("UNDECLARED template executor(s) %v — a new executor is a new DIALECT, and the "+
			"last one to appear unannounced (RenderTemplateWithMap) accepts a language the component "+
			"seam does not, so an ordinary {{safe}} is a parse error in one and fine in the other. "+
			"Declare it in declaredTemplateExecutors and say what its language is.", undeclared)
	}
	if len(stale) > 0 {
		t.Errorf("declaredTemplateExecutors lists %v, which no longer construct a template — "+
			"remove the entry. An allow-list that only grows stops discriminating.", stale)
	}
}
