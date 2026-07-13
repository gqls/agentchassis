package diagnose

import "testing"

func TestIsReadOnlySQL_Accepts(t *testing.T) {
	ok := []string{
		"SELECT name, status FROM pages WHERE site_id = $1",
		"  select 1  ",
		"SELECT * FROM pages;", // single trailing ; allowed
		"WITH recent AS (SELECT * FROM site_work_items WHERE site_id = $1 ORDER BY created_at DESC LIMIT 30) SELECT * FROM recent",
		// word-boundary: these columns/identifiers must NOT trip the write tokens
		"SELECT created_at, updated_at, build_status FROM page_components WHERE page_id = $1",
		"SELECT item_type, claimed_at FROM site_work_items WHERE error IS NOT NULL",
	}
	for _, q := range ok {
		if err := IsReadOnlySQL(q); err != nil {
			t.Errorf("expected OK, got error for %q: %v", q, err)
		}
	}
}

func TestIsReadOnlySQL_Rejects(t *testing.T) {
	bad := []struct {
		q, why string
	}{
		{"", "empty"},
		{"DELETE FROM pages", "starts with DELETE"},
		{"UPDATE pages SET status='x'", "starts with UPDATE"},
		{"DROP TABLE pages", "DDL"},
		{"TRUNCATE pages", "DDL"},
		{"GRANT SELECT ON pages TO x", "privilege"},
		{"COPY (SELECT 1) TO PROGRAM 'rm -rf /'", "COPY TO PROGRAM"},
		{"SELECT 1; DELETE FROM pages", "statement stacking"},
		{"SELECT 1; DROP TABLE pages;", "stacking with trailing ;"},
		{"WITH t AS (DELETE FROM pages RETURNING *) SELECT * FROM t", "data-modifying CTE"},
		{"WITH t AS (UPDATE pages SET status='x' RETURNING *) SELECT * FROM t", "data-modifying CTE"},
		{"VACUUM pages", "side-effect"},
		{"CALL some_proc()", "CALL"},
		{"SHOW search_path", "not SELECT/WITH"}, // benign but not our shape; lint rejects → caught by exec layer anyway
	}
	for _, c := range bad {
		if err := IsReadOnlySQL(c.q); err == nil {
			t.Errorf("expected rejection (%s) for %q, got nil", c.why, c.q)
		}
	}
}
