// FILE: platform/orchestration/actions/work_items_common.go   (NEW FILE)
//
// Centralised terminal-status vocabulary for site_work_items and shared
// SQL helpers. Every callsite that previously inlined this list now
// references the single source of truth here.
//
// The two-strike rule in insertWorkItem is INTENTIONALLY left alone.
// It counts 'complete' AND 'failed' because its job is to break cycles
// between discover agents and fix agents, not to budget retries. A
// discover agent that keeps re-finding an issue after the fix agent
// reports `complete` would loop forever if we only counted 'failed'.
// Two-strike's "terminal count >= 2" semantics catches exactly that.
//
// The real failure mode we hit on gamesdesign — re-cascading a site
// whose prior cascade completed successfully — is an item_key scoping
// issue, not a two-strike issue. Fix is to give cascade re-runs their
// own item_key namespace (e.g. suffixed with a cascade_run_id), not
// to weaken two-strike. That's a separate piece of work.

package actions

// workItemTerminalStatuses is the canonical set of statuses that mark
// a work item as done. idx_swi_dedup and every ON CONFLICT WHERE clause
// on site_work_items must agree with this list — otherwise partial-
// index inference fails (SQLSTATE 42P10).
//
// When adding or removing a status here, also update the DB migration
// that defines idx_swi_dedup.
var workItemTerminalStatuses = []string{
	"complete",
	"failed",
	"verified",
	"rejected",
	"wont_fix",
	"unresolved",
}

// sqlInList formats a Go string slice as a SQL IN literal list for
// interpolation into a query string. No escaping — callers must supply
// already-safe const values (these are package-level constants).
//
//	sqlInList(workItemTerminalStatuses)
//	// "'complete','failed','verified','rejected','wont_fix','unresolved'"
func sqlInList(statuses []string) string {
	out := ""
	for i, s := range statuses {
		if i > 0 {
			out += ","
		}
		out += "'" + s + "'"
	}
	return out
}
