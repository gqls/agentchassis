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
