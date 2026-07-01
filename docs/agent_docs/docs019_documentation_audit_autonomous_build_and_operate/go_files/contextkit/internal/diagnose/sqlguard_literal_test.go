// sqlguard_literal_test.go — regression for the string-literal false-positive:
// a keyword (or ';') INSIDE a quoted literal must not be read as SQL syntax, but
// a real statement keyword / real ';' still must be rejected. The gamesdesign
// loop hit this when a data_request filtered on the page slug
// 'tool-drop-rate-simulator' and the lint flagged "drop".
package diagnose

import "testing"

func TestIsReadOnlySQL_KeywordsInsideLiteralsAccepted(t *testing.T) {
	ok := []string{
		"SELECT name, build_status FROM pages WHERE name = 'tool-drop-rate-simulator'",  // "drop" in slug
		"SELECT id FROM pages WHERE url LIKE '%create-account%'",                        // "create" in literal
		"SELECT name FROM pages WHERE name = 'a;b'",                                     // ';' in literal
		"SELECT rendered_html FROM page_components WHERE content_hash = 'delete-me-42'", // "delete" in literal
	}
	for _, q := range ok {
		if err := IsReadOnlySQL(q); err != nil {
			t.Errorf("keyword inside a literal must be accepted, got error for %q: %v", q, err)
		}
	}
}

func TestIsReadOnlySQL_RealWritesStillRejectedAlongsideLiterals(t *testing.T) {
	bad := []struct{ q, why string }{
		{"SELECT name FROM pages WHERE name='drop'; DROP TABLE pages", "real stacked DROP after a literal 'drop'"},
		{"WITH t AS (DELETE FROM pages WHERE name='x-create-y' RETURNING *) SELECT * FROM t", "data-modifying CTE with a literal keyword"},
	}
	for _, c := range bad {
		if err := IsReadOnlySQL(c.q); err == nil {
			t.Errorf("expected rejection (%s) for %q, got nil", c.why, c.q)
		}
	}
}
