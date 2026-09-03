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

	"encoding/json"
	checks "github.com/gqls/agentchassis/platform/orchestration/actions/discovery_checks"
	"strings"
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

// workItemRevalidatableStatuses is the set revalidate_review_queue may
// re-examine and close on POSITIVE re-observation. It sits with its two
// siblings above so the differences are visible rather than discovered:
//
//	terminal (dedup / ON CONFLICT):  complete verified rejected wont_fix cancelled failed unresolved
//	closed   (retraction):           complete verified rejected wont_fix cancelled
//	revalidatable (this list):       needs_human_review unresolved
//
// ⚠ `failed` IS DELIBERATELY ABSENT, AND THAT IS THE INTERESTING PART.
// RFC_010 Decision 2 puts `unresolved` and `failed` together on the RETRACTION
// path, and the obvious move was to copy the pair across. Measuring the other
// consumers first is what stopped it: the widening's actual blast radius today
// is not on `required_fields_missing` (0 rows) but on `needs_page` — **17
// `failed` rows**, whose revalidator is a different one (revalidateNeedsPage,
// "was the PAGE built?") and whose population this action's own header defers by
// name: *"failures parked by FailWorkItemAction's status_override branch, which
// does not increment attempt_count so they neither retry nor age out. Real
// defect, open owner decision (033 D2), not this sweep's."* Adding `failed`
// would quietly overrule that deferral from inside an unrelated change. It also
// was never needed by the argument below, which is entirely about `unresolved`:
// the two-strike counter brands items `unresolved`, never `failed`.
// Blast radius as narrowed: 1 `needs_page` row, 0 of every other type.
//
// WHY `unresolved` IS HERE (2026-08-04). Until this, the sweep
// selected `needs_human_review` alone — and it was BUILDING A QUEUE IT COULD NOT
// THEN DRAIN. Every close it makes writes `complete` onto a row, which feeds
// insertWorkItem's two-strike counter (`status IN ('complete','failed') AND
// created_at > NOW() - INTERVAL '7 days'`, load_work_item_actions.go:1237).
// Discovery re-raises the finding on the next pass; after the SECOND close
// inside seven days the third re-raise is born `unresolved` — and `unresolved`
// was not in this list, so the sweep could never see it again. Discovery-check
// items do not set recurrenceExpected, so the counter is not skipped for them.
// Measured 2026-08-04: 5 `required_fields_missing` item_keys already sit at 1
// strike, i.e. one close away from that state.
//
// The semantics match RFC_010 Decision 2, which put `unresolved` on the
// retraction path for the same reason: it does not mean "this stopped being a
// problem", it means "we gave up". That is exactly the row that should close
// when something POSITIVELY OBSERVES the condition healthy, and this sweep only ever closes on positive evidence (every named
// field populated; every ambiguity resolves to 'unknown' and stays open).
//
// ⚠ IT IS INTERPOLATED IN THREE PLACES AND THEY MUST NOT DRIFT: the selection in
// loadParkedReviewItems, and the two write-time CAS guards in
// recordRevalidation. The guards re-check the status so a row that changed
// underneath the sweep is not clobbered — so widening the selection alone would
// select the new rows and then silently update nothing.
//
// ⚠ NOT for any `ON CONFLICT` clause — only workItemTerminalStatuses may be
// interpolated there (see that list's comment; using another raises 42P10).
var workItemRevalidatableStatuses = []string{
	"needs_human_review",
	"unresolved",
}

// workItemStatusOverrideAllowed is the set `FailWorkItemAction`'s
// `status_override` step-config key may write. It sits with its three siblings
// above so the differences are visible rather than discovered:
//
//	terminal (dedup / ON CONFLICT):  complete verified rejected wont_fix cancelled failed unresolved
//	closed   (retraction):           complete verified rejected wont_fix cancelled
//	revalidatable:                   needs_human_review unresolved
//	status_override (this list):     needs_human_review blocked wont_fix unresolved cancelled
//
// WHY IT EXISTS (bugs_open/396). `status_override` was a bare string written
// straight into `site_work_items.status` with no allow-list, no enum and no CHECK
// constraint — validated against NONE of the three predicates that actually decide
// a row's fate. Pick a status outside all of them and the row becomes
// undispatchable (`claim_work_item_action.go` claims `triaged`/`approved`),
// un-promotable (the promoter takes `detected`) and still holding its
// `(site_id, item_key)` slot (`deferred` is absent from `idx_swi_dedup`'s
// exclusion list) — so the detector cannot re-file it and any other session
// dispatching that page hits 23505, a failure that reads as "already queued" and
// means "queued and abandoned". Every field on such a row looks healthy.
//
// THE RULE EACH ENTRY HAS TO MEET: a status may be written here only if the row
// can still LEAVE it — either because something named moves it on, or because it
// is terminal and therefore releases the dedup slot. Each line names its exit:
//
//	needs_human_review — HandleRetryWorkItem and the resolve endpoint both take it
//	blocked            — the `feasibility-recheck` scheduled task selects it
//	wont_fix           — terminal: releases the slot, so the detector can re-file
//	unresolved         — terminal (see the closed-vs-terminal note above)
//	cancelled          — terminal, joined the closed set in migration 157
//
// ⚠ `deferred` IS DELIBERATELY ABSENT, and it is the reason this list exists.
// It is the one status that is neither claimable, nor promotable, nor terminal.
// It is also the natural word to reach for when you mean "park this", which is
// exactly the trap. To park work deliberately, use the PARK VERB — migration 621's
// `park_work_items(...)` (register WII-034) — which records who, why and what
// would release it, because those are required arguments rather than a habit.
//
// [MEASURED 2026-08-25] a recursive walk over EVERY `agent_definitions` row —
// snapshots and soft-deleted included — found `status_override` on 4 steps, in 3
// agents, and **every one of them is `needs_human_review`**. No other value has
// ever been configured. So this allow-list narrows a capability nobody uses and
// its blast radius today is zero; it is a door being shut before it is walked
// through, not a behaviour change.
var workItemStatusOverrideAllowed = []string{
	"needs_human_review",
	"blocked",
	"wont_fix",
	"unresolved",
	"cancelled",
}

// statusOverrideRefusedCode names the refusal in agent_error_log. It is its own
// code rather than a generic one because it answers a question no existing code
// answers: "a step config asked for a status this item could never leave, and we
// declined". A reader querying it wants the misconfigured STEP, not a failure.
const statusOverrideRefusedCode = "WORK_ITEM_STATUS_OVERRIDE_REFUSED"

// failWorkItemTemplateFallbackCode names the case where an opt-in
// `error_message_template` did not render and the item was parked with the
// static `error_message` instead (bugs_open/440, RFC_062 phase 3).
//
// Its own code, for the same reason as the one above: it answers "the refusal
// message a human is reading is the FALLBACK, not the intended one". Nothing
// failed — the item parked correctly — so a reader querying it wants the
// misconfigured TEMPLATE, and folding this into a generic failure code would
// bury the one fact that distinguishes it.
const failWorkItemTemplateFallbackCode = "FAIL_WORK_ITEM_MESSAGE_TEMPLATE_FALLBACK"

// statusOverrideAllowed reports whether `status_override` may write s.
//
// The fail direction is deliberate and is the opposite of the caller's
// convenience: an unrecognised value is REFUSED and the item falls through to
// the normal failure ladder, which ages and retries it. Honouring an unknown
// status is the branch that can strand a row for ever; falling back can only
// cost an extra attempt.
func statusOverrideAllowed(s string) bool {
	for _, allowed := range workItemStatusOverrideAllowed {
		if s == allowed {
			return true
		}
	}
	return false
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

// workItemStatusRequiresRegisteredHandler reports whether a row at this status
// is in, or headed into, the dispatch loop's hands — the Go mirror of CHECK
// swi_no_handlerless_promotable's status list (migration 443). At these statuses
// a NAMED handler must be a registered agent, or the row can only ever become
// claim's handler-not-registered `blocked` (bugs_open/291: tool-auditor filed
// 14 real findings at 'hitl-review', a handler that has never existed).
//
// Everywhere else a handler name is decorative-or-future BY DESIGN, and policing
// it would break deliberate idioms: 'detected' is judged at promotion
// (workItemRoutableSQL); 'deferred' capability_gap rows name an UNBUILT builder
// on purpose (load_work_item_actions.go, unavailableBuilders); and parked
// 'needs_human_review' rows are never claimed, several producers naming the
// unregistered pseudo-handler 'human-review' there. Widening this set to those
// statuses would demote every one of them to blocked — recreating
// bugs_closed/284 inside 291's fix.
func workItemStatusRequiresRegisteredHandler(status string) bool {
	for _, s := range workItemDispatchableStatuses {
		if status == s {
			return true
		}
	}
	return status == "claimed"
}

// workItemStatusHeadsForDispatch reports whether a row born at this status is
// ON ITS WAY to a handler — either already in the dispatch loop's hands, or
// waiting for the promoter that will put it there.
//
// It is `workItemStatusRequiresRegisteredHandler` PLUS `detected`, and it is a
// SEPARATE function rather than a widening of that one, deliberately. Widening
// 291's set is the bugs_closed/284 trap: `detected` with no handler is the
// platform's flag-only idiom, and demoting those to `blocked` is precisely the
// regression 291's own test (TestStatusRequiresRegisteredHandler_ExactlyCheck443sList)
// exists to catch. The two questions genuinely differ:
//
//   - REGISTRATION (291) is re-judged downstream. A `detected` row naming an
//     unregistered handler is held back by the promoter's own routability test
//     (workItemRoutableSQL, WDS-017) and released the moment the agent is
//     seeded — so the door can safely ignore `detected` and let promotion decide.
//   - OWNERSHIP (bugs_open/333) is re-judged by NOBODY. No promoter, no
//     scheduled task and no claim-time branch reads pages.rebuild_policy. If the
//     door does not look at a `detected` row, nothing ever will: it is promoted
//     on its next tick and refused by the handler exactly as before.
//
// Hence `detected` is IN this set and OUT of that one. The parked idioms —
// `needs_human_review`, `deferred`, `blocked` — stay out of BOTH: they are not
// heading anywhere, and a row already parked cannot be parked harder.
func workItemStatusHeadsForDispatch(status string) bool {
	return status == "detected" || workItemStatusRequiresRegisteredHandler(status)
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

// countUnroutableDetected answers what the promoter's routability guard held
// back on this site: `detected` rows that name no handler, or name one that is
// not a registered agent. It returns the total and a per-item_type breakdown.
//
// It exists so the guard is not a SILENT cap. A filter that quietly promotes
// fewer rows than it used to is indistinguishable, in every log and every step
// output, from a site that simply had less work — and "the machinery stopped
// doing something" is exactly the shape this estate keeps failing to notice.
// Held-back rows are the DESIGNED outcome for a flag-only finding, so this is
// not an error path; it is the count that makes the design legible.
func countUnroutableDetected(ctx context.Context, db *sql.DB, siteID uuid.UUID) (int64, map[string]int64, error) {
	query := fmt.Sprintf(`
		SELECT wi.item_type, count(*)
		FROM site_work_items wi
		WHERE wi.site_id = $1
		  AND wi.status = 'detected'
		  AND NOT %s
		GROUP BY 1
	`, workItemRoutableSQL("wi"))

	rows, err := db.QueryContext(ctx, query, siteID)
	if err != nil {
		return 0, nil, fmt.Errorf("count unroutable detected items: %w", err)
	}
	defer rows.Close()

	var total int64
	byType := map[string]int64{}
	for rows.Next() {
		var itemType string
		var n int64
		if err := rows.Scan(&itemType, &n); err != nil {
			return 0, nil, fmt.Errorf("scan unroutable detected items: %w", err)
		}
		byType[itemType] = n
		total += n
	}
	if err := rows.Err(); err != nil {
		return 0, nil, fmt.Errorf("iterate unroutable detected items: %w", err)
	}
	return total, byType, nil
}

// workItemHandlerRegisteredSQL renders "this handler is a registered agent" as a
// SQL boolean expression over the given handler expression (a placeholder like
// `$1`, or a column reference like `wi.handler_agent`).
//
// It exists so the test has ONE definition. It is the second half of the claim
// path's routability test (claim_work_item_action.go), and the promoter now
// applies the same test one step earlier — see workItemRoutableSQL. Two hand-kept
// copies of one predicate is the drift class this file already documents twice
// (idx_swi_dedup vs workItemTerminalStatuses; the dispatchable-status trio).
//
// ⚠ DELIBERATELY NOT filtered on is_active / is_snapshot, because the claim path
// is not: a snapshot or deactivated definition still satisfies claim's check and
// still dispatches. Narrowing it HERE without narrowing claim would hold back rows
// claim would have accepted, and nothing downstream would ever promote them —
// the scheduled `detected-item-promoter` is stricter still. If that posture is
// wrong it is wrong in both places, and both are this one function.
// The definition now lives in discovery_checks.HandlerRegisteredSQL — one
// renderer for all three readers (owner ruling 2026-08-17). This stays as a thin
// delegation so every call site in this package is unchanged, which is the other
// half of what the council asked for: the claim path keeps its exact text.
func workItemHandlerRegisteredSQL(handlerExpr string) string {
	return checks.HandlerRegisteredSQL(handlerExpr)
}

// workItemRoutableSQL renders the FULL test the claim path applies before it will
// dispatch a row: the row names a handler, and that handler is registered.
// `alias` is the site_work_items alias in the caller's query.
//
// WHY IT EXISTS (bugs_open/284). ClaimWorkItemAction applies this test AFTER
// claiming, and a row that fails it is stamped `blocked` with an error naming a
// routing failure. That is right for a row that was MEANT to be dispatched. It is
// wrong for the platform's flag-only findings — a palette nothing can repaint, a
// VM nothing can restart, an image reference nothing can repoint — which name no
// handler ON PURPOSE and are meant to sit visible for a human. Measured
// 2026-08-16: 60 such rows across 4 item_types on 15+ sites, every one of them
// promoted into the dispatch queue by TriageDetectedItemsAction, which filtered on
// neither field. `blocked` is not terminal, so each one also holds its dedup slot
// and its check can never file that finding again; and feasibility-recheck can
// never release it, because no agent type is the empty string.
//
// Applying the test at PROMOTION makes that outcome unreachable from the promoter
// rather than merely unlikely, and it costs nothing a caller must remember.
// workItemRetryNotPendingSQL renders the predicate "this item is not sitting out
// a retry cooldown", for an optional table alias ("" for an unaliased UPDATE).
//
// WHY A COMPLETION PATH CARES ABOUT A RETRY STAMP (bugs_open/344). The failure
// ladder (WII-024) re-triages a failed item with `retry_after = now() + backoff`.
// `triaged` is NOT in the completion guard list — deliberately and correctly for
// the world that list was written in, where completing from an in-progress status
// WAS the normal claim flow. The moment the ladder shipped, `triaged` also became
// a POST-FAILURE state, and the dispatch loop's mark_complete — which runs on
// every returned saga, including a handler that failed a step and then ended via
// a success-labelled complete_workflow — began overwriting those rows to
// `complete` about two seconds after the failure. Measured live 2026-08-21: a
// mortgagecalculator.co.uk item stamped `retry_after` 11:02:50 and `completed_at`
// 10:32:52. The retry is cancelled and a failed build is recorded as done.
//
// WHY THIS PREDICATE RATHER THAN ADDING `triaged` TO THE GUARD LIST. A future
// `retry_after` at completion time is DISCRIMINATING, not a heuristic: an item
// that failed, waited out its cooldown, was re-claimed and genuinely succeeded has
// a retry_after in the PAST by then, because the claim path itself refuses to
// re-claim before the stamp expires (proven on the two natural transient rows of
// 2026-08-20 — stamps 18:34:00/18:54:33, re-claims 18:34:25/18:56:58, completions
// after). So a future stamp can only mean THIS saga's own failure write scheduled
// a retry seconds ago, and completing would contradict a decision the same
// contract made. Guarding on the status word `triaged` instead would refuse every
// legitimate completion of an item something re-triaged mid-run for unrelated
// reasons, and would protect the word rather than the decision it happens to carry.
//
// ⚠ IT IS NOT ENOUGH ON ITS OWN. `claimed-item-timeout`'s pre_query auto-completes
// items in raw SQL and never reaches this code, exactly as it bypasses both
// completion gates (bugs_open/317's mechanism). The same predicate is applied to
// that sweep's two auto-complete CTEs by its own migration; they are one contract
// in two media and must not drift.
func workItemRetryNotPendingSQL(alias string) string {
	col := "retry_after"
	if alias != "" {
		col = alias + ".retry_after"
	}
	return "(" + col + " IS NULL OR " + col + " <= NOW())"
}

func workItemRoutableSQL(alias string) string {
	col := alias + ".handler_agent"
	return "(COALESCE(" + col + ", '') <> '' AND " + workItemHandlerRegisteredSQL(col) + ")"
}

// workItemHandlerRefusesOwnedPagesSQL renders "this handler DECLARES that it
// refuses owned pages" (bugs_open/333's door).
//
// WHY A DECLARATION AND NOT A LIST OF NAMES. The door needs to know which
// handlers a finding must not be routed to when its page is owned. A Go slice of
// handler names would be one more hand-maintained copy of a fact the database
// already holds, and it would be wrong in the dangerous direction the day a
// handler opts in: the door would keep filing at a handler that had started
// refusing. Reading the opt-in itself means a handler that adopts
// `refuse_owned_page` is covered by the door in the same migration that makes it
// refuse — no code change, no roll.
//
// Like workItemHandlerRegisteredSQL above, this is a THIN DELEGATION to the
// renderer in discovery_checks, for the same import-direction reason and to keep
// one definition of the predicate. See HandlerDeclaresOwnedPageRefusalSQL for
// why it is not an EXISTS and why its is_active/is_snapshot posture is the
// opposite of the registration check's.
func workItemHandlerRefusesOwnedPagesSQL(handlerExpr string) string {
	return checks.HandlerDeclaresOwnedPageRefusalSQL(handlerExpr)
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

// ── bugs_open/394: THE INVERSE OF THIS COMPOSITION IS NOT SAFE TO WRITE ────
//
// A first cut of the render-audit coverage cursor added a pagePathFromContrastKey
// here, splitting a contrast_failure key on its first '#' to recover the page
// path. It was REMOVED after the council's editquality seat objected, and the
// measurement it prompted says why, so the next author does not re-add it:
//
//   - a SELECTOR may contain '#'. Live 2026-08-26:
//     `contrast_failure:/tools/sfi26-revenue-stacker/index.html#BUTTON#c-tool-…`,
//     and the `describe` scheme emits `tag#id.classes` by construction — two
//     selector schemes coexist here deliberately and permanently (LANDMINES,
//     "The render-audit package now holds TWO selector-composition schemes";
//     register VIZ-016 / WII-016).
//   - a PAGE URL may contain '#'. `idea.uk` has BOTH `/tools.html#audience-check`
//     and `/tools.html` ACTIVE, with 35 open contrast_failure rows. Splitting on
//     the first '#' turns the first into the second — a path that IS a real page
//     on that site — so the wrong page is selected, silently and successfully.
//
// The safe direction is FORWARD, and the grader already goes that way: build the
// prefix from the page with workItemKey(...) and prefix-match
// (write_render_audit_findings_action.go:748). Do that instead of parsing.

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
// workItemFromSpec maps a check's WorkItemSpec onto the runner's workItem.
//
// Extracted from the discovery runner's inline literal because there are now TWO
// places a check-authored spec becomes a row: the runner's insert loop, and the
// RECEIPT resolveWorkItems writes below. Two hand-kept copies of one field
// mapping is the drift class this estate keeps paying for — a field added to the
// spec and mapped in only one of them is silently dropped on the other path, and
// nothing would fail.
func workItemFromSpec(wi checks.WorkItemSpec) workItem {
	return workItem{
		siteID:             wi.SiteID,
		pageID:             wi.PageID,
		source:             wi.Source,
		pipeline:           wi.Pipeline,
		itemType:           wi.ItemType,
		severity:           wi.Severity,
		summary:            wi.Summary,
		spec:               wi.SpecJSON,
		priority:           wi.Priority,
		handlerAgent:       wi.HandlerAgent,
		status:             wi.Status,
		createdBy:          wi.CreatedBy,
		itemKey:            wi.ItemKey,
		batchID:            wi.BatchID,
		recurrenceExpected: wi.RecurrenceExpected,
	}
}

// It never infers. It closes what it is told to close, and only that.
//
// ── THE RECEIPT COUPLING (bugs_open/469) ────────────────────────────────────
// When ResolvedFinding.Receipt is set, this function writes that item FIRST and
// REFUSES the retraction if it cannot make it durable. See the field's own
// comment for why: for a divergence check, "the finding no longer reproduces"
// and "the damage completed" can be the SAME observation, and closing on the
// first while it means the second launders a destroyed page into "resolved".
//
// The refusal is a returned error, which the runner already logs and counts —
// deliberately NOT a silent skip, because a retraction that quietly did not
// happen is indistinguishable from one that did.
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

	// ── The receipt, BEFORE the close, in the SAME transaction. ─────────────
	// Order is the whole point: a run that dies between the two must leave the
	// finding OPEN, never the loss unrecorded. Postgres gives that for free
	// inside one tx; writing the receipt second would not.
	if r.Receipt != nil {
		rc := workItemFromSpec(*r.Receipt)
		if rc.itemKey == "" {
			return 0, fmt.Errorf("resolve %s/%s: receipt has no item_key — it could never be found again; retraction WITHHELD",
				r.ItemType, r.ItemKey)
		}
		inserted, rErr := insertWorkItem(ctx, tx, rc, logger)
		if rErr != nil {
			return 0, fmt.Errorf("resolve %s/%s: receipt %q could not be written — retraction WITHHELD: %w",
				r.ItemType, r.ItemKey, rc.itemKey, rErr)
		}
		if !inserted {
			// insertWorkItem reports FALSE for several different reasons — an
			// open row already holds the key (ON CONFLICT DO NOTHING), the
			// anti-churn brake held it back, the two-strike rule dropped it. Only
			// the first means "a durable record already describes this loss".
			// CONFIRM it; never assume, because assuming turns a dropped receipt
			// into a silent close, which is the exact failure this guard exists
			// to prevent.
			var present int
			if qErr := tx.QueryRowContext(ctx, fmt.Sprintf(`
				SELECT 1 FROM site_work_items
				 WHERE site_id = $1 AND item_key = $2 AND status NOT IN (%s)
				 LIMIT 1`, sqlInList(workItemClosedStatuses)),
				siteID, rc.itemKey).Scan(&present); qErr != nil {
				return 0, fmt.Errorf("resolve %s/%s: receipt %q was neither inserted nor found open (%v) — retraction WITHHELD",
					r.ItemType, r.ItemKey, rc.itemKey, qErr)
			}
			logger.Info("Retraction receipt already open — reusing it rather than filing a duplicate",
				zap.String("check", checkName),
				zap.String("receipt_item_key", rc.itemKey))
		}
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
	//
	// `evidence` is appended CONDITIONALLY rather than always-with-NULL: the
	// six-argument statement below is byte-identical to the one every existing
	// caller's test asserts, so a check that sets no Evidence cannot be broken by
	// this field existing.
	evidenceSQL, args := "", []interface{}{checkName, r.Reason, siteID, r.ItemType, r.ItemKey, batchID}
	if len(r.Evidence) > 0 {
		blob, mErr := json.Marshal(r.Evidence)
		if mErr != nil {
			return 0, fmt.Errorf("resolve %s/%s: evidence is not marshalable — retraction WITHHELD rather than closed without it: %w",
				r.ItemType, r.ItemKey, mErr)
		}
		evidenceSQL = " || $7::jsonb"
		args = append(args, string(blob))
	}

	res, err := tx.ExecContext(ctx, fmt.Sprintf(`
		UPDATE site_work_items
		   SET status       = 'complete',
		       completed_at = now(),
		       updated_at   = now(),
		       result       = COALESCE(result, '{}'::jsonb) || jsonb_build_object(
		                          'resolved_by', $1::text,
		                          'reason',      $2::text,
		                          'resolved_at', now()::text)%s
		 WHERE site_id   = $3
		   AND item_type = $4
		   AND ($5::text = '' OR item_key = $5::text)
		   AND status NOT IN (%s)
		   AND batch_id IS DISTINCT FROM $6::uuid`, evidenceSQL, sqlInList(workItemClosedStatuses)),
		args...)
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

// emitRequiredFieldsMissing is the ONE render-time producer of
// required_fields_missing items (bugs_open/342, council bb7f5d0e round 6,
// reuse_agent seat).
//
// It exists because I wrote it twice. The chrome path and the two section-editor
// routes each had a near-identical emitter — same item_type, same handler, same
// key shape, differing only in the surface they name and whether they are armed
// — and a second seat had just praised collapsing two copies of a declared-type
// check into one primitive. Three copies of one write is how three subtly
// different behaviours start: one gets a severity bump, one forgets the handler,
// and the router quietly stops seeing a third of the population.
//
// THE KEY SHAPE IS SHARED WITH check_required_fields_missing DELIBERATELY
// (required_fields_missing:<scope>:<slot-or-function>), so the post-deploy
// producer and this render-time one co-dedup on one key instead of filing two
// items for one defect. That check scans build_status='deployed' — rows that
// made it — which is exactly why a render-time producer was needed: a section
// that renders empty is dropped and never becomes such a row.
//
// Failures are logged, never returned: no render or edit may fail because a note
// could not be written.
// pageContext is what the ROUTER needs and what this emitter did not supply for
// its first day alive (see the header note above). Zero value = "no page", which
// is the chrome surface's honest state, not a caller forgetting.
type pageContext struct {
	id   *uuid.UUID
	name string
	slot string
}

// requiredFieldsMissingRouting decides an item's key, handler and intake status
// from one question: is there a page? Factored out of the emitter so a test can
// exercise the decision that was WRONG — the emitter itself needs a database,
// and both defects it carried (an unroutable spec and a key that could not
// co-dedup) are computable without one.
//
// WITH a page this is the same finding the post-deploy check files, so it takes
// the SAME key — `required_fields_missing:<page_id>:<slot_name>`,
// check_required_fields_missing.go:180 — and the two producers genuinely
// co-dedup. ⚠ They did NOT before 2026-08-23: this producer keyed on
// `<site_id>:<component function>`, so the "matching item_key" claim in the bug
// file, the register and the APPROVED council submission was false, and the two
// would have filed separate items for one defect.
//
// WITHOUT a page — the chrome surface, whose slots hang off the SITE — the
// page-resolving router structurally cannot classify the item, so handing it
// over buys three failed attempts and a parked item, which is exactly what the
// editor route demonstrated (item a31da7f3).
//
// THE RESIDUE IS A `capability_gap`, AND THAT IS NOT AN INVENTION — it is
// `bugs_closed/077`'s convention, for exactly this shape: *a detector whose
// predicate is wider than its handler's remit files the residue as a capability
// gap*, with `gap_kind` distinguishing `handler_remit` (a handler exists, this
// finding is outside it — us) from `handler_missing` (none exists at all).
// 45 such rows are live, all with an empty handler, and — the part that decides
// it — **`diagnose_triage_action.go`'s roadmap sweep reads
// `item_type='capability_gap' OR status='deferred'`, grouped by
// `spec->>'builder_needed'`**, so a gap filed this way has a CONSUMER.
//
// ⚠ The first version of this parked at `needs_human_review` with the
// `required_fields_missing` type — invented, and nothing swept it, so the items
// would have aged forever. Two council seats caught it on trail a0ef0b07
// (`bug_historian`, medium, naming 077; `debug_historian`, low, naming the
// no-consumer risk). The irony is worth keeping: this same change ships a
// landmine saying *reusing a type is not reusing its contract*, and the seats
// caught me NOT reusing a type that already fitted.
//
// Deliberately still an EMPTY handler_agent, matching all 45 live rows and
// `bugs_closed/291`: an unregistered handler name is born `blocked` and never
// claimed, which reads as a queue bug rather than as a roadmap entry.
//
// BOTH halves of the page context are required before taking the routed path:
// the classify step resolves the page by name AND the component by slot, so
// half the context must not be allowed to look like all of it. The fallback key
// carries the SURFACE so a page-shaped residue and a chrome finding on the same
// site cannot collide on a shared scope name (council a0ef0b07, editquality).
func requiredFieldsMissingRouting(siteID uuid.UUID, page pageContext, surface, scopeKey string) (itemType, itemKey, handler, status string) {
	if page.id != nil && page.slot != "" {
		return "required_fields_missing",
			fmt.Sprintf("required_fields_missing:%s:%s", *page.id, page.slot),
			"required-fields-missing-handler", "detected"
	}
	return "capability_gap",
		fmt.Sprintf("capability_gap:required_fields_missing:%s:%s:%s", siteID, surface, scopeKey),
		"", "deferred"
}

func emitRequiredFieldsMissing(ctx context.Context, db *sql.DB, siteID uuid.UUID,
	page pageContext, componentID *uuid.UUID, scopeKey, label, surface, source string,
	absent []string, extraSpec map[string]interface{}, logger *zap.Logger) {

	if len(absent) == 0 {
		return
	}
	if db == nil || siteID == uuid.Nil {
		logger.Warn("required_fields_missing: no site identity available, item not filed",
			zap.String("surface", surface), zap.Strings("absent_required_fields", absent))
		return
	}

	spec := map[string]interface{}{
		"surface":        surface,
		"missing_fields": absent,
		"source":         source,
		"detected_at":    "render",
		"fix": "Schema-required source:llm field(s) rendered EMPTY. Go's missingkey=zero renders " +
			"an absent field as empty with no error, so the section ships blank rather than " +
			"failing, and page assembly drops a visually-empty section (bugs_open/342). Supply " +
			"the values, or change the schema if they are not really required. NOTE this was " +
			"detected AT RENDER: the post-deploy check cannot see it, because it reads rows that " +
			"have already reached a deployed build status and a dropped section never becomes one.",
	}
	// THE FIELDS THE ROUTER ACTUALLY READS. required-fields-missing-handler's
	// `classify` step resolves the page by `spec->>'page_name'` and the component
	// by `spec->>'slot_name'`; without them it classifies the item `malformed`,
	// routes to mark_failed, and the attempt ladder parks it. That is not a
	// theory — it happened to the first item this emitter ever filed
	// (`a31da7f3`, 2026-08-22): three attempts, three `route: "malformed"`
	// classifications, terminal status `failed`, no repair attempted.
	//
	// ⚠ THE LESSON, because it is the reusable half: REUSING A TYPE IS NOT
	// REUSING ITS CONTRACT. This producer was written to reuse
	// `required_fields_missing` precisely so the existing router would handle
	// it, that reasoning was reviewed and approved, and nobody — me least of
	// all — checked that the items carried the keys the router reads.
	if page.name != "" {
		spec["page_name"] = page.name
	}
	if page.slot != "" {
		spec["slot_name"] = page.slot
	}
	for k, v := range extraSpec {
		spec[k] = v
	}
	specJSON, _ := json.Marshal(spec)

	summary := fmt.Sprintf("%s left %d schema-required field(s) empty: %s",
		label, len(absent), strings.Join(absent, ", "))
	if len(summary) > 250 {
		summary = summary[:247] + "..."
	}

	// ── Key, handler and status all follow from ONE question: is there a page? ──
	//
	// WITH a page, this is the same finding the post-deploy check files, so it
	// takes the SAME key — `required_fields_missing:<page_id>:<slot_name>`,
	// check_required_fields_missing.go:180 — and the two producers genuinely
	// co-dedup. ⚠ They did NOT before 2026-08-23: this emitter keyed on
	// `<site_id>:<component function>`, so the "matching item_key" claim in the
	// bug file, the register and the council submission was FALSE and the two
	// producers would have filed separate items for one defect. Corrected at
	// source rather than only in the docs.
	//
	// WITHOUT a page — the chrome surface, whose slots hang off the SITE — the
	// page-resolving router structurally cannot classify the item, so handing it
	// over would buy three failed attempts and a parked item, which is what the
	// editor route just demonstrated. It is filed for a human instead: honest
	// about who can act on it, and visible rather than fake-routed. Not a
	// phantom handler (bugs_closed/291: an unregistered handler is born blocked
	// and never claimed, which reads as a queue bug).
	itemType, itemKey, handler, status := requiredFieldsMissingRouting(siteID, page, surface, scopeKey)

	// A capability_gap says WHAT IS MISSING, not just that something is: the
	// roadmap sweep groups by `spec->>'builder_needed'`, so a gap with no
	// builder named lands in a `?` bucket and tells the reader nothing.
	if itemType == "capability_gap" {
		spec["gap_kind"] = "handler_remit"
		spec["builder_needed"] = "a required_fields_missing router that can service the " +
			surface + " surface (required-fields-missing-handler resolves pages by " +
			"spec.page_name; this surface has no page)"
		spec["finding_type"] = "required_fields_missing"
		specJSON, _ = json.Marshal(spec)
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		logger.Warn("required_fields_missing: begin tx failed", zap.String("surface", surface), zap.Error(err))
		return
	}
	inserted, err := insertWorkItem(ctx, tx, workItem{
		siteID:       siteID,
		pageID:       page.id,
		componentID:  componentID,
		source:       source,
		pipeline:     "build",
		itemType:     itemType,
		severity:     "medium",
		summary:      summary,
		spec:         string(specJSON),
		priority:     50,
		handlerAgent: handler,
		status:       status,
		createdBy:    source,
		itemKey:      itemKey,
	}, logger)
	if err != nil {
		_ = tx.Rollback()
		logger.Warn("required_fields_missing: insert failed", zap.String("surface", surface), zap.Error(err))
		return
	}
	if err := tx.Commit(); err != nil {
		logger.Warn("required_fields_missing: commit failed", zap.String("surface", surface), zap.Error(err))
		return
	}
	if inserted {
		logger.Info("required_fields_missing item filed at RENDER time (bugs_open/342)",
			zap.String("surface", surface), zap.String("scope", scopeKey),
			zap.Strings("absent_required_fields", absent))
	}
}

// siteLockExceptionSQL renders the site-lock arm that `LoadWorkItemsAction`
// appends when its step config sets `honour_site_lock: true` (bugs_open/396).
//
// It is a function rather than an inline string for the same reason
// `refreshOpenWorkItemSQL` is: the test that pins its two halves reads THE
// FRAGMENT THAT RUNS, instead of re-deriving its own copy and agreeing with
// itself.
//
// BOTH HALVES ARE LOAD-BEARING AND THE TEST PINS BOTH:
//
//   - `(SELECT s.locked_at …) IS NULL` — an UNLOCKED site is unaffected, which is
//     every site but three as of 2026-08-25. Lose this and the loader would
//     select nothing anywhere.
//   - `wi.id = ANY(COALESCE(s.lock_except_item_ids, ARRAY[]::uuid[]))` — on a
//     LOCKED site, only the excepted items. The COALESCE is what makes a NULL
//     exception list mean the full hold rather than an error, so "locked with no
//     exceptions" keeps meaning exactly what a lock has always meant.
//
// It takes the site id from `$1`, which the caller already binds as the only
// positional argument at that point in the statement — deliberately, so this
// fragment adds no parameter and cannot shift the caller's argIdx bookkeeping.
//
// ⚠⚠ DO NOT REUSE THIS FRAGMENT IN `find_dispatchable_site`. THE TWO SPELLINGS OF
// THIS RULE ARE DELIBERATELY DIFFERENT AND MUST STAY DIFFERENT.
//
// This one is PER-SITE and PARAMETERISED: the caller has already bound the site
// id as `$1`, so the lock is looked up by scalar subquery.
//
// The dispatch selector (`build-pipeline-trigger > find_dispatchable_site`,
// migration 633) is a CROSS-SITE SCAN that already `JOIN sites s ON s.id =
// wi.site_id` and has NO `$1` site parameter at all. There it is spelled against
// the joined alias:
//
//	WHERE (s.locked_at IS NULL
//	       OR wi.id = ANY(COALESCE(s.lock_except_item_ids, ARRAY[]::uuid[])))
//
// Dropping THIS fragment into THAT query would reference a `$1` that does not
// exist and break site selection fleet-wide. A council reviewer read the plan for
// 633 and concluded exactly that had been done (corr `175df761`, editquality) —
// which is the tell that the next person to "DRY these up" will try it. The rule
// is one rule; the two SQL contexts are not the same context, and there is no
// shared form that is correct in both.
func siteLockExceptionSQL() string {
	return `
		  AND (
		    (SELECT s.locked_at FROM sites s WHERE s.id = $1) IS NULL
		    OR wi.id = ANY(COALESCE(
		         (SELECT s.lock_except_item_ids FROM sites s WHERE s.id = $1),
		         ARRAY[]::uuid[]))
		  )`
}

// workItemNotGovernorShedSQL renders "this row is NOT withheld by the LLM spend
// governor" (D4 stage B; register AGOV-013; design of record in
// doc_plans('pipeline','spend-governor')).
//
// The governor sheds work in the owner-ruled order of 2026-08-31: shed_level 1
// withholds LLM-bearing MAINTENANCE items, level 2 adds BUILD items, level 3
// adds RESEARCH. Classes live in `governor_work_class_map` as data; the current
// level lives in `governor_state`, recomputed every 120s by the
// `spend-governor-state` task. LLM-free rows are never withheld — withholding
// them saves nothing and stops serving.
//
// THE LOGIC LIVES IN ONE PLACE AND IT IS NOT HERE. `governor_admits(item_type)`
// (migration 675) is the single canonical predicate — this renderer emits a
// one-line CALL, and so does the dispatch selector's config text (migration
// 674). That was the council's architecture revision on corr 8f4bb57d r1: four
// hand-copied spellings of one rule across two Go files and a jsonb-stored
// query is a drift machine no VERIFY string-compare can keep honest; a function
// makes the lockstep structural. Do NOT re-spell the shed logic in Go — the
// posture rules (fail-open on an unreadable governor; unmapped item_type sheds
// earliest as maintenance+llm_bearing; enabled=false is identity) are execution-
// proven by 675's verify probes and documented on the function itself.
//
// ⚠ The SELECTOR must carry the same call (674 does this), or shed-only sites
// become selection hogs: the selector would rank a site by rows the loader then
// refuses to load, and the site never goes busy — the bugs_closed/413
// starvation shape, governor edition.
//
// ⚠ ORDERING: the function must exist before any step config flips
// honour_spend_governor on. 675 applies with stage A (inert — nothing calls
// it); the flags arrive only via held 674, whose preflight asserts the
// function exists.
func workItemNotGovernorShedSQL(alias string) string {
	return "governor_admits(" + alias + ".item_type)"
}
