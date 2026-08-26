// FILE: platform/orchestration/datahelpers/scan_completeness.go
//
// ONE definition of "did this row scan lose anything the query gave me"
// (bugs_open/410, instance 3).
//
// THE DEFECT THIS CLOSES. A `for rows.Next()` loop whose `rows.Scan` error
// branch logs and `continue`s returns FEWER rows than the cursor yielded, with
// NO error. The caller cannot tell a short result from a genuinely short table,
// so the work completes green and the artefact is left looking freshly built —
// not blank, not erroring, not obviously stale: REBUILT. That is the shape
// bugs_open/410 was filed for, across three seams in three lanes in one week,
// all failing toward the quiet default.
//
// WHICH COUNT — this is the load-bearing decision and the intuitive version is
// the wrong one. bugs_open/410 pins it, and the reason is worth restating here
// because a guard that fires on healthy input gets loosened within a week, and
// a loosened guard is a dead one:
//
//   - WRONG: rows kept, compared against "rows that exist in the table for this
//     key". Queries legitimately filter in SQL — the motivating caller's own
//     WHERE drops `build_status = 'removed'` tombstones — so that comparison
//     fires on every page carrying a removed component, which is a large and
//     entirely healthy population.
//   - RIGHT: rows the CURSOR YIELDED, compared against rows KEPT. It needs no
//     second query, so it also cannot race a concurrent writer the way a
//     re-count would — which matters on a shared tree where another lane may be
//     writing the same page.
//
// In one line: the guard's job is "did I lose anything the query gave me", not
// "does the result have as many rows as I expected". Only the first is knowable
// inside the function, and only the first is invariant to legitimate filtering.
//
// THIS IS PROPAGATION, NOT INVENTION. Two callers in
// platform/orchestration/actions already implement this guard by hand, and both
// were required by a review seat rather than volunteered:
//
//   - collectPageSections (validate_page_content_surface_sections.go) refuses
//     component grain for a whole page when its canonical reader resolves fewer
//     sections than the metadata held — council round 3ed2b792, bug_historian.
//   - scanBlogArticles (rebuild_blog_listing_action.go) counts offered against
//     kept across a rows.Scan loop and errors when every offered row failed —
//     council round 170147b4, bug_historian, which rejected a first cut that
//     logged-and-skipped unconditionally and documented the exposure in prose
//     without closing it.
//
// scanBlogArticles is the GRADED sibling: it tolerates partial loss (one
// malformed post must not blank a live listing) and errors only on total loss.
// That policy is right for a projection and wrong for a wholesale replace — see
// the caller's own comment in loadStoredSections. It is deliberately NOT
// converted to this helper: it is already guarded, and a behaviour-neutral
// refactor of the listing rebuild would widen a bug fix's blast radius for
// nothing. Whoever next touches it can adopt these counters.
//
// POLICY IS THE CALLER'S. This file ships exactly one function, the strict one,
// because that is the policy the first caller exercises. A graded twin shipped
// "for completeness" would be a mechanism nobody runs, which this estate has
// been bitten by before and which its reviewers object to by name.
package datahelpers

import "fmt"

// ScanShortfall reports whether a row-scanning loop silently dropped rows.
//
// offered is incremented once per `rows.Next()` returning true — i.e. rows the
// cursor actually handed over. kept is what survived the scan, typically
// len(out). subject names the reader, and appears in the error.
//
// Returns nil when nothing was lost, INCLUDING when offered is zero: a
// genuinely empty result set is not a failure and must never be reported as
// one. It is the caller's own no-rows path that decides what an empty result
// means.
//
// Usage — three lines grafted onto an existing loop, no restructuring:
//
//	var out []T
//	offered := 0
//	for rows.Next() {
//	        offered++
//	        if err := rows.Scan(...); err != nil {
//	                // scan-loss:accepted: counted — ScanShortfall below refuses
//	                logger.Warn("...", zap.Error(err))
//	                continue
//	        }
//	        out = append(out, v)
//	}
//	if err := rows.Err(); err != nil {
//	        return nil, err
//	}
//	return out, ScanShortfall(offered, len(out), "my_reader: what it read")
//
// Keeping the per-row Warn and continue is deliberate rather than lazy: on a
// mixed failure it records EVERY failing row's cause, where an immediate return
// would record only the first — and the first is rarely the informative one when
// a projection has drifted.
func ScanShortfall(offered, kept int, subject string) error {
	if kept >= offered {
		return nil
	}
	return fmt.Errorf("%s: kept %d of %d rows the cursor yielded (%d lost to scan failures) — "+
		"refusing the partial result: a silently thinned scan ships an artefact that looks "+
		"freshly built (bugs_open/410)",
		subject, kept, offered, offered-kept)
}
