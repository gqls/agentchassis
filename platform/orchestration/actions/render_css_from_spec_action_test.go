// FILE: platform/orchestration/actions/render_css_from_spec_action_test.go
//
// Integration tests for loadComponentCSSSnippets in
// render_css_from_spec_action.go.
//
// Shares test helpers (setupSnippetTestDB, createSnippetTestSchema,
// cleanupTestSnippets, mustExec) with render_js_snippets_for_site_action_test.go
// via package scope.
//
// Most important test: TestLoadComponentCSSSnippets_OverlapSemantics —
// regression guard for the jsonb && jsonb operator bug. Reverting to
// `applies_to && $1::jsonb` would fail this test with the operator-
// does-not-exist error in the test logs.

package actions

import (
	"context"
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

// TestLoadComponentCSSSnippets_OverlapSemantics is the regression test
// for the jsonb && jsonb bug. Same shape as the JS analog. Verifies
// the EXISTS overlap check works on real Postgres.
func TestLoadComponentCSSSnippets_OverlapSemantics(t *testing.T) {
	db := setupSnippetTestDB(t)
	defer db.Close()
	createSnippetTestSchema(t, db)
	cleanupTestSnippets(t, db)
	defer cleanupTestSnippets(t, db)

	logger := zaptest.NewLogger(t)
	ctx := context.Background()

	// Seed: distinctive CSS bodies so we can identify which made it through
	mustExec(t, db, `
		INSERT INTO css_snippets (name, css_content, applies_to) VALUES
			('_test_news_grid',    '.news-grid{display:grid}',  '["latest-news"]'::jsonb),
			('_test_listing_page', '.listing{padding:2rem}',    '["news-listing"]'::jsonb),
			('_test_card_styles',  '.card{border-radius:8px}',  '["card","feature"]'::jsonb),
			('_test_unused',       '.unused{display:none}',     '["nothing-uses-this"]'::jsonb)
	`)

	t.Run("matching component returns matching snippet CSS", func(t *testing.T) {
		got := loadComponentCSSSnippets(ctx, db, []string{"latest-news"}, logger)
		require.NotEmpty(t, got)
		require.Contains(t, got, ".news-grid{display:grid}")
		require.NotContains(t, got, ".unused")
	})

	t.Run("multiple matching components combine", func(t *testing.T) {
		got := loadComponentCSSSnippets(ctx, db, []string{"latest-news", "card"}, logger)
		require.Contains(t, got, ".news-grid")
		require.Contains(t, got, ".card{border-radius:8px}")
		require.NotContains(t, got, ".unused")
	})

	t.Run("no overlapping components returns empty string", func(t *testing.T) {
		got := loadComponentCSSSnippets(ctx, db, []string{"completely-unused-component"}, logger)
		require.Empty(t, got)
	})

	t.Run("includes the component-specific styles header when matches exist", func(t *testing.T) {
		got := loadComponentCSSSnippets(ctx, db, []string{"latest-news"}, logger)
		require.Contains(t, got, "/* === Component-specific styles === */")
	})

	t.Run("results ordered by name", func(t *testing.T) {
		got := loadComponentCSSSnippets(ctx, db, []string{"latest-news", "news-listing"}, logger)
		// _test_listing_page should appear after _test_news_grid alphabetically
		idxGrid := strings.Index(got, "_test_news_grid")
		idxListing := strings.Index(got, "_test_listing_page")
		// They appear in the log lines via Info("included snippet"), but the
		// css_content bodies are concatenated without the names. The order
		// guarantee is in the SQL ORDER BY name — verify via positions of
		// the distinctive body strings:
		idxGridBody := strings.Index(got, ".news-grid{display:grid}")
		idxListingBody := strings.Index(got, ".listing{padding:2rem}")
		require.NotEqual(t, -1, idxGridBody)
		require.NotEqual(t, -1, idxListingBody)
		require.Less(t, idxListingBody, idxGridBody,
			"_test_listing_page (alpha-before _test_news_grid) should appear first")
		_ = idxGrid
		_ = idxListing
	})
}

// TestLoadComponentCSSSnippets_EdgeCases — graceful handling of empty
// inputs. These paths return "" (the original behavior) and are tested
// to make sure that contract is preserved.
func TestLoadComponentCSSSnippets_EdgeCases(t *testing.T) {
	logger := zaptest.NewLogger(t)
	ctx := context.Background()

	t.Run("nil db returns empty", func(t *testing.T) {
		got := loadComponentCSSSnippets(ctx, nil, []string{"hero"}, logger)
		require.Empty(t, got)
	})

	t.Run("empty components returns empty (no DB query needed)", func(t *testing.T) {
		db := setupSnippetTestDB(t)
		defer db.Close()
		got := loadComponentCSSSnippets(ctx, db, []string{}, logger)
		require.Empty(t, got)
	})

	t.Run("nil components returns empty", func(t *testing.T) {
		db := setupSnippetTestDB(t)
		defer db.Close()
		got := loadComponentCSSSnippets(ctx, db, nil, logger)
		require.Empty(t, got)
	})
}

// ─────────────────────────────────────────────────────────────────────────────
// The append order of the three renderer-owned :root blocks (bugs_open/122).
//
// buildLegibleInkDefaults skips any custom-property name the ASSEMBLED CSS
// already defines. That makes its position load-bearing: run it before
// buildTokenAliases and a later block can define the same name and win, so the
// companion is emitted-and-inert — the dead-config shape the whole seam exists
// to avoid.
//
// Round 1 of this change stated the dependency in a comment. The council's
// bug_historian and guardian seats both objected that a positional invariant
// with no positive assertion is not an invariant, and they were right: nothing
// would have failed if a later edit moved the call.
//
// This reads the AST rather than the source text. A regex or strings.Index over
// the file would be satisfied by the ORDER OF THE COMMENTS — first occurrence
// wins — so a reorder that left the comments in place would pass. Parsing means
// only real call expressions count.
// ─────────────────────────────────────────────────────────────────────────────

func TestRenderCSS_InkCompanionsComeAfterTokenAliases(t *testing.T) {
	const file = "render_css_from_spec_action.go"

	fset := token.NewFileSet()
	parsed, err := parser.ParseFile(fset, file, nil, 0) // 0 => comments discarded
	if err != nil {
		t.Fatalf("parse %s: %v", file, err)
	}

	// Offset of each call expression, in source order, ignoring comments.
	pos := map[string]int{}
	ast.Inspect(parsed, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		ident, ok := call.Fun.(*ast.Ident)
		if !ok {
			return true
		}
		switch ident.Name {
		case "buildSectionDefaults", "buildTokenAliases", "buildLegibleInkDefaults":
			if _, seen := pos[ident.Name]; !seen {
				pos[ident.Name] = int(call.Pos())
			}
		}
		return true
	})

	for _, name := range []string{"buildSectionDefaults", "buildTokenAliases", "buildLegibleInkDefaults"} {
		if _, ok := pos[name]; !ok {
			t.Fatalf("no call to %s found in %s — the renderer no longer appends that block, "+
				"which is a bigger change than a reorder", name, file)
		}
	}

	if pos["buildTokenAliases"] > pos["buildLegibleInkDefaults"] {
		t.Errorf("buildLegibleInkDefaults is called BEFORE buildTokenAliases.\n" +
			"It skips names already present in the assembled CSS, so from this order a " +
			"later block can define --color-primary-ink and win: the companion is emitted " +
			"but inert, which is exactly the dead-config shape this seam exists to avoid " +
			"(bugs_open/122). Restore step 12 after step 11.")
	}
	if pos["buildSectionDefaults"] > pos["buildLegibleInkDefaults"] {
		t.Errorf("buildLegibleInkDefaults is called BEFORE buildSectionDefaults; " +
			"it must see the fully assembled CSS. Restore step 12 last.")
	}
}

// The skip must read the ASSEMBLED css, not the layout template. If the first
// argument is ever changed to something narrower, the "already defined" check
// stops seeing blocks appended by steps 9-11 and starts redefining names that
// are already present.
func TestRenderCSS_InkCompanionsAreCheckedAgainstTheAssembledCSS(t *testing.T) {
	fset := token.NewFileSet()
	parsed, err := parser.ParseFile(fset, "render_css_from_spec_action.go", nil, 0)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}

	found := false
	ast.Inspect(parsed, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		if ident, ok := call.Fun.(*ast.Ident); !ok || ident.Name != "buildLegibleInkDefaults" {
			return true
		}
		found = true
		if len(call.Args) == 0 {
			t.Error("buildLegibleInkDefaults called with no arguments")
			return false
		}
		arg, ok := call.Args[0].(*ast.Ident)
		if !ok || arg.Name != "renderedCSS" {
			t.Errorf("first argument is %s, want renderedCSS — the skip check must see "+
				"every block appended by steps 9-11, or it will redefine names that "+
				"are already present", types.ExprString(call.Args[0]))
		}
		return false
	})
	if !found {
		t.Fatal("no call to buildLegibleInkDefaults found")
	}
}
