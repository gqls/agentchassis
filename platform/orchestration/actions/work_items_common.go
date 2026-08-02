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

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"go.uber.org/zap"

	checks "github.com/gqls/agentchassis/platform/orchestration/actions/discovery_checks"
)

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
	// "cancelled" joined the closed set in migration 157 (2026-07-16):
	// a cancelled row must not hold the dedup slot. This list is interpolated
	// into insertWorkItem's ON CONFLICT ... WHERE, whose clause MUST imply
	// idx_swi_dedup's predicate or every keyed insert fails with SQLSTATE
	// 42P10 ("no unique or exclusion constraint matching the ON CONFLICT
	// specification") — which is exactly what happened between 157 being
	// applied and this line landing. Keep list and index in lockstep.
	"cancelled",
}

// workItemClosedStatuses is the set a RETRACTION must not disturb: an item that
// has already reached a settled conclusion.
//
// IT IS DELIBERATELY NOT workItemTerminalStatuses, AND THE DIFFERENCE IS THE
// POINT (RFC_010, owner ruling 2026-08-02 "Decision 2: `unresolved` is OPEN").
// The two lists sit together so the difference is visible rather than
// discovered:
//
//	terminal (dedup / ON CONFLICT):  complete verified rejected wont_fix cancelled failed unresolved
//	closed   (retraction):           complete verified rejected wont_fix cancelled
//
// `unresolved` and `failed` are absent here on purpose. Neither means "this
// stopped being a problem" — they mean "we gave up" and "the handler errored".
// If a check now POSITIVELY OBSERVES the underlying condition to be healthy,
// those are exactly the rows that should close; leaving them open is how
// `unresolved` became the landfill RFC_010 describes (not terminal, not
// deduplicated, not retractable, so copies accumulate silently — nine of eleven
// items in one incident were duplicates of two findings).
//
// ⚠ THIS LIST IS NOT INTERPOLATED INTO ANY `ON CONFLICT` CLAUSE and must never
// be. Only workItemTerminalStatuses may be, because only it matches
// idx_swi_dedup's predicate; using this one there would fail partial-index
// inference with SQLSTATE 42P10 on every keyed insert, which is the breakage
// migration 157 already caused once. Retraction is a plain UPDATE and touches
// no index predicate, which is why it can honour the ruling today while the
// dedup half of the same ruling waits on a coupled schema+binary change.
var workItemClosedStatuses = []string{
	"complete",
	"verified",
	"rejected",
	"wont_fix",
	"cancelled",
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

// workItemDispatchableStatuses is the canonical set of statuses at which an
// item is waiting for the dispatch loop, i.e. promoted and not yet claimed.
// Its two siblings must agree with it:
//
//   - claim_work_item_action.go — the atomic claim's `status IN (...)` guard;
//   - load_work_item_actions.go — the dispatcher's selection query.
//
// Same lockstep obligation as workItemTerminalStatuses / idx_swi_dedup above,
// with a softer failure: a drift here does not error, it silently changes what
// "this site has work waiting" means.
var workItemDispatchableStatuses = []string{
	"triaged",
	"approved",
}

// countDispatchableWorkItems answers a question about the SITE, not about the
// caller's own result set: how many work items are sitting in a dispatchable
// status for this pipeline right now, whoever put them there.
//
// WHY IT EXISTS (bugs_open/150). Three live agents run the triage_detected_items
// step, the promotion is unconditional over the site, and the parent's copy runs
// last — so the parent legitimately reports `promoted: 0` for 67 findings a child
// promoted seconds earlier, and a branch reading that lands the improvement loop
// on "No issues found — site is clean". A step's own result is not the site's
// state; this is the site's state.
//
// The predicate is deliberately NARROWER than the dispatcher's selection query,
// which additionally filters attempt_count, approval_mode and depends_on. Those
// clauses answer "will the loader return this row on its next tick"; this
// function answers "is there unfinished promoted work here". Anything they would
// exclude — an attempt-exhausted item, one awaiting approval, one whose
// dependency has not landed — is still work, and counting it can only err toward
// NOT clean. That is the safe direction for every caller: a needless closing
// rerender costs a render, a false "clean" costs the finding.
func countDispatchableWorkItems(ctx context.Context, db *sql.DB, siteID uuid.UUID, pipeline string) (int64, error) {
	query := fmt.Sprintf(`
		SELECT count(*)
		FROM site_work_items
		WHERE site_id = $1
		  AND status IN (%s)
		  AND pipeline = $2
	`, sqlInList(workItemDispatchableStatuses))

	var count int64
	if err := db.QueryRowContext(ctx, query, siteID, pipeline).Scan(&count); err != nil {
		return 0, fmt.Errorf("count dispatchable work items: %w", err)
	}
	return count, nil
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

// resolveWorkItems closes work items a discovery check has POSITIVELY OBSERVED
// to be fixed. It is the runner-side half of checks.CheckResult.Resolved
// (RFC_010, owner ruling 2026-08-02).
//
// One implementation, called from one place, deliberately: the alternative is
// each check spelling its own UPDATE, and two hand-maintained copies of one
// status vocabulary is the drift class this estate keeps paying for — it is
// literally what idx_swi_dedup vs workItemTerminalStatuses cost once already.
// Before this, exactly one check (backend_unreachable) could retract, via its
// own inline query.
//
// It never infers. It closes what it is told to close, and only that.
func resolveWorkItems(
	ctx context.Context,
	tx *sql.Tx,
	siteID uuid.UUID,
	checkName string,
	batchID uuid.UUID,
	r checks.ResolvedFinding,
	logger *zap.Logger,
) (int, error) {
	// Validation is a REFUSAL, not a best guess. Each of these means the check
	// is making a claim it has not specified, and the wide branch is far too
	// destructive to reach by accident — closing every open item of a type for
	// a site because a field was left blank is precisely the failure the
	// owner's 2026-08-02 ruling ("ships as an opt-in field, unsafe default OFF")
	// exists to prevent.
	switch {
	case r.ItemType == "":
		return 0, fmt.Errorf("resolve: ItemType is empty — refusing to guess which item type %q meant", checkName)
	case r.Reason == "":
		return 0, fmt.Errorf("resolve: Reason is empty for item_type %q — a row that closes itself with no "+
			"stated cause is indistinguishable later from one a human closed by hand", r.ItemType)
	case r.ItemKey == "" && !r.AllOfType:
		return 0, fmt.Errorf("resolve: neither ItemKey nor AllOfType set for item_type %q — set AllOfType "+
			"explicitly if you really mean every open item of this type for this site", r.ItemType)
	case r.ItemKey != "" && r.AllOfType:
		return 0, fmt.Errorf("resolve: both ItemKey (%q) and AllOfType set for item_type %q — these mean "+
			"different things; pick one", r.ItemKey, r.ItemType)
	}

	// `batch_id IS DISTINCT FROM $batch` — never close what THIS run just
	// raised. A check that both files and resolves the same key in one run is
	// contradicting itself; without this it would thrash an item open/closed on
	// every sweep and look stable in aggregate.
	//
	// The status predicate is workItemClosedStatuses, NOT
	// workItemTerminalStatuses: `unresolved` and `failed` are retractable on
	// purpose (owner ruling, Decision 2 — see that list's comment). This is a
	// plain UPDATE and infers no partial index, so it cannot raise 42P10.
	res, err := tx.ExecContext(ctx, fmt.Sprintf(`
		UPDATE site_work_items
		   SET status       = 'complete',
		       completed_at = now(),
		       updated_at   = now(),
		       result       = COALESCE(result, '{}'::jsonb) || jsonb_build_object(
		                          'resolved_by', $1::text,
		                          'reason',      $2::text,
		                          'resolved_at', now()::text)
		 WHERE site_id   = $3
		   AND item_type = $4
		   AND ($5::text = '' OR item_key = $5::text)
		   AND status NOT IN (%s)
		   AND batch_id IS DISTINCT FROM $6::uuid`, sqlInList(workItemClosedStatuses)),
		checkName, r.Reason, siteID, r.ItemType, r.ItemKey, batchID)
	if err != nil {
		return 0, fmt.Errorf("resolve %s/%s: %w", r.ItemType, r.ItemKey, err)
	}

	n, _ := res.RowsAffected()
	if n > 0 {
		logger.Info("Retracted work items no longer reproducing",
			zap.String("check", checkName),
			zap.String("item_type", r.ItemType),
			zap.String("item_key", r.ItemKey),
			zap.Bool("all_of_type", r.AllOfType),
			zap.String("reason", r.Reason),
			zap.Int64("items", n))
	}
	return int(n), nil
}
