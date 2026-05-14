// FILE: platform/orchestration/actions/render_js_snippets_for_site_action_test.go
//
// Integration tests for render_js_snippets_for_site_action.go.
//
// Most important test in this file: TestLoadJSSnippetsForSite_OverlapSemantics.
// That test verifies the EXISTS + jsonb_array_elements_text overlap query
// returns the right rows. If someone reintroduces `applies_to && $1::jsonb`
// (operator does not exist: jsonb && jsonb), this test fails with the SQL
// error message in the test output.
//
// Requires a Postgres test DB. Tests skip if TEST_DB_HOST is unset (or
// the connection fails). Same env var pattern as orchestration_test.go:
//   TEST_DB_HOST, TEST_DB_PORT, TEST_DB_USER, TEST_DB_PASSWORD, TEST_DB_NAME
// Defaults: localhost, 5432, postgres, postgres, agentchassis_test.

package actions

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"strings"
	"testing"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"
)

// ---------------------------------------------------------------------------
// Test DB setup helpers (shared with render_css_from_spec_action_test.go via
// package scope — both files are in package actions)
// ---------------------------------------------------------------------------

// setupSnippetTestDB opens a connection to the test database using the same
// TEST_DB_* env vars as orchestration_test.go. Calls t.Skip if connection
// fails, so CI without a Postgres test DB doesn't see red.
func setupSnippetTestDB(t *testing.T) *sql.DB {
	t.Helper()

	host := getEnvOr("TEST_DB_HOST", "")
	if host == "" {
		t.Skip("TEST_DB_HOST not set — skipping integration test")
	}

	port := getEnvOr("TEST_DB_PORT", "5432")
	user := getEnvOr("TEST_DB_USER", "postgres")
	password := getEnvOr("TEST_DB_PASSWORD", "postgres")
	dbname := getEnvOr("TEST_DB_NAME", "agentchassis_test")

	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	db, err := sql.Open("pgx", connStr)
	if err != nil {
		t.Skipf("opening test DB failed: %v", err)
	}

	if err := db.Ping(); err != nil {
		t.Skipf("test DB unreachable: %v", err)
	}

	return db
}

func getEnvOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

// createSnippetTestSchema creates css_snippets and js_snippets with the
// real-world column types (jsonb applies_to). Idempotent — safe to call
// repeatedly. Test rows have unique names prefixed `_test_` so cleanup
// is selective and doesn't disturb other tests' data.
func createSnippetTestSchema(t *testing.T, db *sql.DB) {
	t.Helper()

	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS css_snippets (
			id            uuid PRIMARY KEY DEFAULT gen_random_uuid(),
			name          varchar(100) UNIQUE NOT NULL,
			description   text,
			css_content   text NOT NULL,
			semantic_tags jsonb DEFAULT '[]'::jsonb,
			applies_to    jsonb DEFAULT '[]'::jsonb,
			created_at    timestamp DEFAULT now()
		)
	`)
	require.NoError(t, err)

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS js_snippets (
			id            uuid PRIMARY KEY DEFAULT gen_random_uuid(),
			name          varchar(100) UNIQUE NOT NULL,
			description   text,
			js_content    text NOT NULL,
			semantic_tags jsonb DEFAULT '[]'::jsonb,
			applies_to    jsonb DEFAULT '[]'::jsonb,
			dependencies  jsonb DEFAULT '[]'::jsonb,
			is_active     boolean NOT NULL DEFAULT false,
			created_at    timestamp DEFAULT now()
		)
	`)
	require.NoError(t, err)
}

// cleanupTestSnippets removes only rows this test suite inserted. Names
// are prefixed `_test_` so a row from production data won't get touched.
func cleanupTestSnippets(t *testing.T, db *sql.DB) {
	t.Helper()
	_, _ = db.Exec(`DELETE FROM css_snippets WHERE name LIKE '\_test\_%' ESCAPE '\'`)
	_, _ = db.Exec(`DELETE FROM js_snippets  WHERE name LIKE '\_test\_%' ESCAPE '\'`)
}

// ---------------------------------------------------------------------------
// loadJSSnippetsForSite — the main tests
// ---------------------------------------------------------------------------

// TestLoadJSSnippetsForSite_OverlapSemantics is the regression test for
// the jsonb && jsonb bug. If the SQL reverts to using `&&`, this test
// fails with the operator-does-not-exist error.
func TestLoadJSSnippetsForSite_OverlapSemantics(t *testing.T) {
	db := setupSnippetTestDB(t)
	defer db.Close()
	createSnippetTestSchema(t, db)
	cleanupTestSnippets(t, db)
	defer cleanupTestSnippets(t, db)

	logger := zaptest.NewLogger(t)
	ctx := context.Background()

	// Seed: three snippets with various applies_to and is_active values
	mustExec(t, db, `
		INSERT INTO js_snippets (name, js_content, applies_to, is_active) VALUES
			('_test_news_formatter', 'function fmtA(){}', '["latest-news","news-listing"]'::jsonb, true),
			('_test_scroll_reveal',  'function fmtB(){}', '["section","card"]'::jsonb,            true),
			('_test_dormant',        'function fmtC(){}', '["latest-news"]'::jsonb,               false)
	`)

	t.Run("returns snippets whose applies_to overlaps components, ignoring is_active=false", func(t *testing.T) {
		got, err := loadJSSnippetsForSite(ctx, db, []string{"latest-news"}, logger)
		require.NoError(t, err)

		names := snippetNames(got)
		require.Contains(t, names, "_test_news_formatter")
		require.NotContains(t, names, "_test_scroll_reveal", "scroll_reveal does not apply to latest-news")
		require.NotContains(t, names, "_test_dormant", "dormant is is_active=false and must be excluded")
	})

	t.Run("multiple matching component values produce union of matches", func(t *testing.T) {
		got, err := loadJSSnippetsForSite(ctx, db, []string{"latest-news", "section"}, logger)
		require.NoError(t, err)

		names := snippetNames(got)
		require.Contains(t, names, "_test_news_formatter")
		require.Contains(t, names, "_test_scroll_reveal")
	})

	t.Run("no overlapping components produces empty result", func(t *testing.T) {
		got, err := loadJSSnippetsForSite(ctx, db, []string{"completely-unused-component"}, logger)
		require.NoError(t, err)
		require.Empty(t, got)
	})

	t.Run("results are ordered by name for deterministic output", func(t *testing.T) {
		// Insert two more snippets whose names sort in known order
		mustExec(t, db, `
			INSERT INTO js_snippets (name, js_content, applies_to, is_active) VALUES
				('_test_zzz_last',  'js', '["card"]'::jsonb, true),
				('_test_aaa_first', 'js', '["card"]'::jsonb, true)
		`)

		got, err := loadJSSnippetsForSite(ctx, db, []string{"card"}, logger)
		require.NoError(t, err)
		require.GreaterOrEqual(t, len(got), 2)

		// Find positions of _test_aaa_first and _test_zzz_last
		idxFirst := -1
		idxLast := -1
		for i, s := range got {
			if s.Name == "_test_aaa_first" {
				idxFirst = i
			}
			if s.Name == "_test_zzz_last" {
				idxLast = i
			}
		}
		require.NotEqual(t, -1, idxFirst)
		require.NotEqual(t, -1, idxLast)
		require.Less(t, idxFirst, idxLast, "results must be ordered by name")
	})
}

// ---------------------------------------------------------------------------
// extractComponentFunctionsList — pure unit tests, no DB
// ---------------------------------------------------------------------------

func TestExtractComponentFunctionsList(t *testing.T) {
	logger := zaptest.NewLogger(t)

	cases := []struct {
		name string
		data map[string]interface{}
		path string
		want []string
	}{
		{
			name: "nested path with []interface{} string list",
			data: map[string]interface{}{
				"site_context": map[string]interface{}{
					"all_component_functions": []interface{}{"hero", "features", "footer"},
				},
			},
			path: "site_context.all_component_functions",
			want: []string{"hero", "features", "footer"},
		},
		{
			name: "[]string directly",
			data: map[string]interface{}{
				"things": []string{"a", "b"},
			},
			path: "things",
			want: []string{"a", "b"},
		},
		{
			name: "missing path returns nil",
			data: map[string]interface{}{},
			path: "anywhere",
			want: nil,
		},
		{
			name: "non-string elements in []interface{} are skipped",
			data: map[string]interface{}{
				"mixed": []interface{}{"valid", 42, nil, "also-valid", ""},
			},
			path: "mixed",
			want: []string{"valid", "also-valid"},
		},
		{
			name: "wrong type returns nil",
			data: map[string]interface{}{
				"not_a_list": "hero",
			},
			path: "not_a_list",
			want: nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := extractComponentFunctionsList(tc.data, tc.path, logger)
			require.Equal(t, tc.want, got)
		})
	}
}

// ---------------------------------------------------------------------------
// buildJSSnippetsBundle / emptyJSSnippetsBundle — pure unit tests
// ---------------------------------------------------------------------------

func TestBuildJSSnippetsBundle(t *testing.T) {
	siteID := "00000000-0000-0000-0000-000000000001"

	t.Run("empty list produces header only", func(t *testing.T) {
		out := buildJSSnippetsBundle(siteID, nil)
		require.Contains(t, out, siteID)
		require.Contains(t, out, "0 active snippet(s)")
		require.NotContains(t, out, "/* --- ")
	})

	t.Run("includes each snippet with header and body", func(t *testing.T) {
		snippets := []jsSnippetRow{
			{Name: "alpha", Description: "first", JSContent: "function a(){}"},
			{Name: "beta", Description: "", JSContent: "function b(){}\n"},
		}
		out := buildJSSnippetsBundle(siteID, snippets)

		require.Contains(t, out, "2 active snippet(s)")
		require.Contains(t, out, "/* --- alpha — first --- */")
		require.Contains(t, out, "function a(){}")
		require.Contains(t, out, "/* --- beta --- */")
		require.Contains(t, out, "function b(){}")
	})

	t.Run("trailing newline is normalised", func(t *testing.T) {
		// Snippet without trailing newline: bundle should add one before next
		out := buildJSSnippetsBundle(siteID, []jsSnippetRow{
			{Name: "x", JSContent: "no_newline"},
			{Name: "y", JSContent: "also_no_newline"},
		})
		// Both bodies are present
		require.Contains(t, out, "no_newline")
		require.Contains(t, out, "also_no_newline")
		// Bodies aren't smashed onto one line
		require.NotContains(t, out, "no_newlinealso_no_newline")
		require.NotContains(t, out, "no_newline/*", "should have blank line before next header")
	})
}

func TestEmptyJSSnippetsBundle(t *testing.T) {
	out := emptyJSSnippetsBundle("site-id-123")
	require.Contains(t, out, "site-id-123")
	require.Contains(t, out, "no active snippets")
	// Must be valid JS — a comment is enough
	require.True(t, strings.HasPrefix(strings.TrimSpace(out), "/*"))
	require.True(t, strings.HasSuffix(strings.TrimSpace(out), "*/"))
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func mustExec(t *testing.T, db *sql.DB, q string, args ...interface{}) {
	t.Helper()
	_, err := db.ExecContext(context.Background(), q, args...)
	require.NoError(t, err)
}

func snippetNames(s []jsSnippetRow) []string {
	out := make([]string, len(s))
	for i, x := range s {
		out[i] = x.Name
	}
	return out
}

// Silence unused import warning if zap is referenced only in tests above
var _ = zap.NewNop
