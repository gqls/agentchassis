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
	// statements, so an internal ';' is the stacking-injection signal. Literals are
	// blanked first (below), so a ';' inside a string literal no longer trips this.
	body := strings.TrimSuffix(strings.TrimSpace(s), ";")

	// Blank the CONTENTS of string literals / quoted identifiers before the stacking
	// and write-token checks, so a keyword or ';' that appears INSIDE a literal — e.g.
	// a page slug like 'tool-drop-rate-simulator', whose "drop" otherwise reads as DDL
	// — is not mistaken for SQL syntax. The read-only transaction/role (see header)
	// stays the real guarantee, so blanking here cannot admit a write; it only removes
	// false rejections.
	scan := stripQuoted(body)
	if strings.Contains(scan, ";") {
		return fmt.Errorf("multiple statements not allowed (found ';' mid-query)")
	}

	lower := strings.ToLower(scan)
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

// stripQuoted blanks the CONTENTS of single-quoted string literals and
// double-quoted identifiers, keeping the delimiters, so a keyword or ';' that
// appears INSIDE a literal/identifier (e.g. the page slug
// 'tool-drop-rate-simulator', which contains "drop") cannot trip the
// statement-stacking or write-token checks. Handles the SQL escapes ” and "".
// Lint-accuracy only: the read-only transaction/role remains the real guarantee,
// so blanking literals here cannot admit an actual write.
func stripQuoted(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	i := 0
	for i < len(s) {
		c := s[i]
		if c == '\'' || c == '"' {
			q := c
			b.WriteByte(q)
			i++
			for i < len(s) {
				if s[i] == q {
					if i+1 < len(s) && s[i+1] == q { // doubled = escaped delimiter, stay in
						i += 2
						continue
					}
					break
				}
				i++ // drop literal/identifier content
			}
			if i < len(s) { // closing delimiter
				b.WriteByte(q)
				i++
			}
			continue
		}
		b.WriteByte(c)
		i++
	}
	return b.String()
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
