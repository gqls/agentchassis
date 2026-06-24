// sqlguard.go — a cheap pre-flight lint for model-WRITTEN diagnosis queries.
//
// READ THIS FIRST — what this is NOT:
// This is NOT the safety boundary. SQL cannot be proven read-only by string
// inspection: data-modifying CTEs (WITH t AS (DELETE ... RETURNING *) SELECT ...),
// statement stacking (psql -c runs "SELECT 1; DELETE ..."), COPY ... TO PROGRAM,
// and side-effecting functions (SELECT pg_terminate_backend(...)) all read like
// reads. The REAL guarantee is the EXECUTION SUBSTRATE:
//   - chassis: db.BeginTx(ctx, &sql.TxOptions{ReadOnly: true}) + QueryContext —
//     the engine rejects any write (incl. data-modifying CTEs), and the extended
//     protocol rejects multi-statement strings;
//   - harness: -psql pointed at a READ-ONLY ROLE (default_transaction_read_only,
//     SELECT-only grants), so stacking is moot;
//   - both: statement_timeout to bound a runaway scan.
//
// IsReadOnlySQL is DEFENCE IN DEPTH on top of that: it catches an obviously-wrong
// query early with a clear operator message (and guards against a mis-provisioned
// role), but the read-only transaction/role is what makes model-written SQL safe.
// Never run model SQL on a read-write connection on the strength of this lint
// alone.
package diagnose

import (
	"fmt"
	"strings"
)

// writeTokens are word-boundary tokens that must never appear in a read-only
// query. Includes the keywords that make a CTE data-modifying (insert/update/
// delete/merge) and the DDL / privilege / side-effect statements.
var writeTokens = []string{
	"insert", "update", "delete", "merge", "truncate",
	"drop", "alter", "create", "grant", "revoke",
	"copy", "call", "do", "vacuum", "analyze", "reindex",
	"refresh", "cluster", "lock", "comment", "security",
}

// IsReadOnlySQL returns nil if sql passes the pre-flight lint, else a descriptive
// error. Conservative by design (a secondary gate): it would rather reject a
// legitimate-but-unusual read than admit a write. The real guarantee is the
// read-only transaction/role around execution (see the file header).
func IsReadOnlySQL(sql string) error {
	s := strings.TrimSpace(sql)
	if s == "" {
		return fmt.Errorf("empty query")
	}

	// Drop a single trailing ';', then reject any remaining ';' — psql -c stacks
	// statements, so an internal ';' is the stacking-injection signal. (A ';'
	// inside a string literal would trip this too; for a read-only LINT that
	// over-rejection is acceptable — rewrite the query without the literal ';'.)
	body := strings.TrimSuffix(strings.TrimSpace(s), ";")
	if strings.Contains(body, ";") {
		return fmt.Errorf("multiple statements not allowed (found ';' mid-query)")
	}

	lower := strings.ToLower(body)
	if !strings.HasPrefix(lower, "select") && !strings.HasPrefix(lower, "with") {
		return fmt.Errorf("query must start with SELECT or WITH, got %q", firstWord(body))
	}

	for _, tok := range writeTokens {
		if containsWord(lower, tok) {
			return fmt.Errorf("write/DDL token %q is not allowed in a read-only query", tok)
		}
	}
	return nil
}

// containsWord reports whether tok appears in s on word boundaries (so "created_at"
// does not match the token "create", and "deleted" does not match "delete").
func containsWord(s, tok string) bool {
	from := 0
	for {
		i := strings.Index(s[from:], tok)
		if i < 0 {
			return false
		}
		i += from
		leftOK := i == 0 || !isWordByte(s[i-1])
		rightIdx := i + len(tok)
		rightOK := rightIdx >= len(s) || !isWordByte(s[rightIdx])
		if leftOK && rightOK {
			return true
		}
		from = i + len(tok)
	}
}

func isWordByte(b byte) bool {
	return b == '_' ||
		(b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z') ||
		(b >= '0' && b <= '9')
}

func firstWord(s string) string {
	s = strings.TrimSpace(s)
	if i := strings.IndexAny(s, " \t\n("); i > 0 {
		return s[:i]
	}
	return s
}
