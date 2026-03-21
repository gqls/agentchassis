// FILE: platform/orchestration/datahelpers/nullable_helpers.go
//
// Additional nullable helpers for SQL parameter building.

package datahelpers

import "database/sql"

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
