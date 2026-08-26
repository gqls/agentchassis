// FILE: platform/orchestration/datahelpers/tombstones.go
//
// THE tombstone-exclusion predicate, in one place.
//
// A page_components row with build_status='removed' is a TOMBSTONE: the
// webdesign rebuild lane retires ported slots by status flip and keeps the
// bytes for ever (never DELETE), and the page assembler excludes those rows
// from every served page. Everything that means "the live population of this
// page/component" must exclude tombstones with the SAME predicate the
// assembler uses, or the two populations drift apart silently.
//
// WHY IS DISTINCT FROM AND NOT <> (council 89dcc04a, 2026-08-26, executing
// debug_historian's 21540c8e advisory): build_status is NULLABLE (default
// 'pending' covers only omitted columns; information_schema is_nullable=YES).
// The assembler keeps a NULL-status row — `IS DISTINCT FROM` — while a bare
// inequality would drop it under SQL three-valued logic. Before this
// constant existed, five call sites hand-spelled the clause and FOUR of them
// were NULL-unsafe: a NULL-status row would have been SERVED by the assembler
// yet invisible to every tool audit, selector census and staleness check —
// the exact inversion of the tombstone defect. Zero NULL rows existed
// fleet-wide when measured (2026-08-26); nothing constrains the column, so
// the constant is what keeps that latent door shut.
//
// USE THE HELPER, NEVER HAND-SPELL THE CLAUSE. TestNoHandSpelledTombstone-
// Predicate walks platform/orchestration and fails on any non-comment,
// non-test hand-spelling — in either the NULL-safe or NULL-unsafe form,
// because a correct copy is still a copy that can drift.

package datahelpers

// NotRemovedSQL is the bare tombstone-exclusion predicate, for queries whose
// FROM clause makes build_status unambiguous.
const NotRemovedSQL = `build_status IS DISTINCT FROM 'removed'`

// NotRemoved returns the tombstone-exclusion predicate qualified for a table
// alias: NotRemoved("pc") -> "pc.build_status IS DISTINCT FROM 'removed'".
// An empty alias returns the bare form.
func NotRemoved(alias string) string {
	if alias == "" {
		return NotRemovedSQL
	}
	return alias + "." + NotRemovedSQL
}
