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

// workItemKey builds the canonical deduplication key for a site_work_items
// row. The contract is "{itemType}:{target}", with the prefix EQUAL to the
// row's item_type, so that:
//   - idx_swi_dedup (UNIQUE on (site_id, item_key) over non-terminal rows)
//     collapses exactly the rows that represent the same unit of work, and
//   - the key is safe to filter / group by type (its prefix encodes the type).
//
// Every creator should build item_key through this helper — whether it
// inserts via insertWorkItem(workItem{...}) or an inline INSERT ... VALUES —
// rather than fmt.Sprintf-ing its own prefix. Hand-rolled prefixes are how the
// keys drifted from their item_type in the first place (e.g. flag_page_image_
// rebuild minting page_rerender:<page> for a needs_page row, and adoption
// minting needs_page:<name> for BOTH a content page and a tool recreation).
//
// Deliberate shared-namespace use is allowed but must be commented at the
// callsite: when two item_types are genuinely the SAME dedup unit on the SAME
// handler and SHOULD collapse together, a creator may pass the namespace-owning
// type rather than the row's own item_type — e.g. an adoption needs_content_page
// that must co-dedup with a planner needs_page build of the same page would call
// workItemKey("needs_page", page.Name). State which namespace is shared and why,
// so the exception is visible to the "prefix == item_type" invariant.
func workItemKey(itemType, target string) string {
	return itemType + ":" + target
}
