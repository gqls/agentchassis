// FILE: platform/orchestration/datahelpers/nullable_helpers.go
//
// Additional nullable helpers for SQL parameter building.

package datahelpers

import (
	"database/sql"
	"strings"
)

// NullableInt returns sql.NullInt64 for optional integer fields
func NullableInt(i int) sql.NullInt64 {
	if i == 0 {
		return sql.NullInt64{}
	}
	return sql.NullInt64{Int64: int64(i), Valid: true}
}

// NullableInt64 returns sql.NullInt64 for optional int64 fields
func NullableInt64(i int64) sql.NullInt64 {
	if i == 0 {
		return sql.NullInt64{}
	}
	return sql.NullInt64{Int64: i, Valid: true}
}

// PGTextArrayLiteral converts a Go []string into a PostgreSQL array literal
// suitable for passing as a parameter to INSERT/UPDATE statements targeting
// a text[] column, when used with an explicit ::text[] cast in the SQL.
//
// Format produced: {"elem1","elem2","elem3"}
// Empty or nil input produces: {}
// Per-element backslash and double-quote are escaped per PG array-literal rules.
//
// Example usage:
//
//	tags := []string{"foo", "bar"}
//	_, err := db.ExecContext(ctx,
//	    `INSERT INTO style_collections (industry_tags) VALUES ($1::text[])`,
//	    datahelpers.PGTextArrayLiteral(tags),
//	)
//
// NOTE: Do NOT json.Marshal a []string for a text[] column — jsonb and text[]
// are different Postgres types and the INSERT will fail with
// "expression is of type jsonb".
//
// This helper exists because the codebase uses database/sql (wrapping
// jackc/pgx/v5/stdlib) with neither lib/pq's pq.Array nor pgx's pgtype
// array types imported. Passing an array literal string is the cheapest
// path that works under both drivers.
func PGTextArrayLiteral(tags []string) string {
	if len(tags) == 0 {
		return "{}"
	}
	var b strings.Builder
	b.WriteByte('{')
	for i, t := range tags {
		if i > 0 {
			b.WriteByte(',')
		}
		b.WriteByte('"')
		// Escape backslash first, then double-quote
		for j := 0; j < len(t); j++ {
			c := t[j]
			switch c {
			case '\\', '"':
				b.WriteByte('\\')
				b.WriteByte(c)
			default:
				b.WriteByte(c)
			}
		}
		b.WriteByte('"')
	}
	b.WriteByte('}')
	return b.String()
}
