// FILE: platform/orchestration/actions/owned_page_guard.go
//
// The guard on the GENERIC composition loops (bugs_open/208, bugs_open/450).
//
// The question this file answers is "MAY THE GENERIC BUILDER WRITE THIS PAGE?",
// which has two independent NO answers. The file is still called
// owned_page_guard because ownership was the first of them and the marker
// literals (OWNED_PAGE_GUARD, owned_page_review) are load-bearing in matchers,
// dedup keys and operator queries across the estate; renaming them would cost
// more than the tidiness is worth. Both answers travel as a refusal CLASS —
// refusalOwned, refusalToolPending — so a receipt says which one applied.
//
// The second answer, added 2026-09-03 (bugs_open/450): A TOOL PAGE WITH NO TOOL
// ON IT. A plan names tool pages before the tools exist; the tool arrives later
// from the design rotation under a name the planner never saw; meanwhile five
// generic producers (unbuilt_internal_link first, then empty_section,
// needs_page, needs_content_page, page_rerender) route the page to
// page-build-handler, which builds exactly what the plan said — prose about the
// tool — and deploys it. It serves 200 at full weight with the tool's own
// headline and no form, so every status-shaped check passes. 61 such pages
// across 10 sites as of 2026-09-03. The predicate is DERIVED from the page's
// contents rather than stored as a flag; see toolShellPredicateFor for why that
// distinction is the whole design.
//
// pages.rebuild_policy='owned' marks a page that belongs to a tool/widget or is a
// runtime-fill shell — Experience Loop guard rail 1, migration 164, mechanising
// TL-001 after the vonc arena clobber. Migration 164 named two refusals:
// ReconcileSitePlanAction emits owned_page_review instead of needs_page, and
// SavePageSectionsAction hard-refuses a generic section save.
//
// Those two were sufficient for the route 164 modelled (needs_page ->
// page-build-handler, whose live order is save_sections -> update_status ->
// deploy_page, so the refusal precedes any commit). They are NOT sufficient for
// the three loops that compose and commit in the opposite order — measured
// 2026-08-06, page-rebuild, pageflow-builder and site-work-orchestrator all run:
//
//	assemble_page -> deploy_page (git_commit) -> save_sections -> update_page_status
//
// There the regenerated HTML is committed to the site repo, which the site
// deploys from, one step BEFORE the guard refuses the database write. The live
// tool is destroyed and page_components still describes it.
//
// This file holds the shared parts of the fix so the three touch points cannot
// drift apart: page identity resolution, the policy read, and the review-item
// emission. Keeping the precondition in a helper rather than in each caller is
// deliberate — a precondition parked in a caller is one port away from gone
// (the lesson recorded on bugs_closed/145).
//
// WHY assemble_page and not git_commit: git_commit is also how owned pages
// LEGITIMATELY deploy. page-rerender (rerender_single_page) and section-editor
// (apply_section_edit) both commit pages, and migration 164 says in terms that
// re-assembly of existing page_components "is deliberately NOT gated — it is how
// owned pages deploy". assemble_page is used by exactly the three loops above and
// by nothing else, and is fed freshly LLM-written HTML
// (content_field: page_content.response.page_html) in all three. It is therefore
// the one seam that means "generic composition about to be committed".
package actions

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/google/uuid"
	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"go.uber.org/zap"
)

// ownedRebuildPolicy is the pages.rebuild_policy value meaning "not the generic
// pipeline's to rebuild" (migration 164's CHECK allows only 'generic'|'owned').
const ownedRebuildPolicy = "owned"

// The refusal CLASSES this guard can return. A page may be off-limits to the
// generic builder for two unrelated reasons, and the receipts, the log lines and
// the operator's next action differ, so the reason travels with the verdict
// rather than being inferred from a bool.
//
//	refusalOwned       — pages.rebuild_policy='owned' (migration 164; adopted or
//	                     ported pages, runtime-fill shells, verbatim tools)
//	refusalToolPending — a TOOL page with no tool on it yet (bugs_open/450)
const (
	refusalOwned       = "owned"
	refusalToolPending = "tool_pending"
)

// disableToolShellRefusalEnv disarms the tool_pending arm ONLY; the owned arm
// (migration 164, live since August) is unaffected, and the whole door still has
// its own switch at the writeWorkItem seam. Ships ARMED — the owner has ruled
// against default-OFF switches that rot unexercised, so this exists to be
// disarmed in anger, and the split exists so a misfire in the new arm cannot
// cost the old one (the 444 CONTRIB's independent-rollback argument, as an env
// lever rather than a config key).
const disableToolShellRefusalEnv = "DISABLE_TOOL_SHELL_REFUSAL"

// toolShellPredicateFor renders the "this is a tool page with no tool on it"
// test against one table alias (both the page_type and the id come off it).
// ONE rendering, because this predicate now decides whether five producers may
// write to a page, and two copies of it would be the drift this file exists to
// prevent.
//
// bugs_open/450: a plan names tool pages BEFORE their tools exist — tools arrive
// from the design rotation hours-to-days later, under names the planner never
// saw (seotools: 0 of 7 planned names matched what tool-deployer built). The
// generic builder then fills the page with prose about the tool, deploys it, and
// every status-shaped check passes. The only honest question is whether a live
// tool actually occupies the page, so that is what this asks.
//
// DERIVED, NEVER STORED — and that is the whole design. The obvious fix is to
// mark the page (rebuild_policy, or a new 'tool-pending' value), but NOTHING in
// this estate has ever UPDATEd rebuild_policy: there are two INSERT-time writers
// and no transition anywhere, so a mark set at plan time is a mark nobody clears
// and the page would be protected for ever. A derived predicate self-clears the
// instant tool-deployer inserts the tool component — verified ordering, and it
// matters for a live lane: deploy_tool_action.go INSERTs the component (step 5)
// BEFORE it raises the companion needs_content_page (step 6), so the tool
// pipeline's own follow-up work is never refused by this.
//
// DELIBERATELY NOT discovery_checks.toolEligibilityWhere, which looks similar.
// That fragment also requires the page to carry exactly ONE component, because
// it is answering "is this a ported tool whose identity we must infer". This is
// answering "does a live tool occupy this page RIGHT NOW", which has no such
// condition — the 450 shells carry two components and would be invisible to it.
// NotRemoved is the shared spelling of the liveness filter (matching the
// assembler); a tombstoned tool slot still refuses generic builds, which is
// protective rather than accidental.
func toolShellPredicateFor(alias string) string {
	return `(` + alias + `.page_type = 'tool' AND NOT EXISTS (
		SELECT 1 FROM page_components pc_g
		JOIN content_components cc_g ON cc_g.id = pc_g.component_id
		WHERE pc_g.page_id = ` + alias + `.id
		  AND ` + datahelpers.NotRemoved("pc_g") + `
		  AND cc_g.component_level = 'tool'
		  AND cc_g.is_active = true))`
}

// genericBuildRefusal derives the verdict from the two facts read off the page
// row. Pure, so the decision table is testable without a database and the
// kill-switch has exactly one place it can be honoured.
//
// Returns ("", false) when the generic builder may proceed.
func genericBuildRefusal(policy string, toolShell bool) (refused bool, class string) {
	if policy == ownedRebuildPolicy {
		return true, refusalOwned
	}
	if toolShell && toolShellRefusalArmed() {
		return true, refusalToolPending
	}
	return false, ""
}

// toolShellRefusalArmed is the ONE reading of the kill switch, so the SQL-side
// exclusion and the Go-side verdict can never disagree about whether the arm is
// live. Two readings of one env var is the same drift class as two predicates.
func toolShellRefusalArmed() bool {
	return os.Getenv(disableToolShellRefusalEnv) == ""
}

// ownedPageSkipReasonPrefix marks a refusal caused by this guard. It is the
// string to pod-grep to prove the guard is live on a running binary, and it is
// what tells an operator reading a skipped iteration WHY the page was left alone.
//
// > **CORRECTED 2026-08-18:** this comment used to say "It is matched by
// > nothing — the skip protocol is a boolean". That was already false when
// > written (SavePageSectionsAction and UpdatePageStatusAction both match it in
// > an upstream skip_reason) and it is now load-bearing in a third place:
// > SavePageSectionsAction leads its refusal ERROR with it, so that
// > UpdateWorkItemStatusAction's owned_page_refusal_status can tell an ownership
// > refusal apart from a real save failure. The full matcher set, so a future
// > edit knows what it would break:
// >
// >	SavePageSectionsAction     — an upstream skip_reason, and its own error
// >	UpdatePageStatusAction     — an upstream skip_reason (refuses the 'deployed' stamp)
// >	UpdateWorkItemStatusAction — __step_error.message (chooses the terminal status)
// >
// > Changing this literal changes an item's terminal STATUS, not only a log line.
// >
// > EMITTERS (errors that LEAD with it, added 2026-08-19, bugs_open/301):
// > SavePageSectionsAction's refusal, and LoadPageRecordAction's early refusal
// > (refuseOwnedPageIfConfigured, behind the refuse_owned_page opt-in) — both
// > feed UpdateWorkItemStatusAction's matcher above via error_step routing.
const ownedPageSkipReasonPrefix = "OWNED_PAGE_GUARD"

// resolveGuardedPage finds the page a composition step is working on.
//
// The three consumers name it differently, so all four shapes are tried in
// order of authority (an id beats a name lookup):
//
//	page-rebuild / pageflow-builder : current_page.id, then current_page.name
//	site-work-orchestrator          : current_item.spec.id, then .spec.name
//
// current_item.spec is the page record for build items — WriteBuildItemsAction
// marshals what queryPagesForBuild returned, PLUS `handler` (routing provenance,
// added 2026-08-25, bugs_open/206: the handler the emit chose, so a later reader
// can tell a mint from a hand repair of handler_agent). Nothing here reads it —
// the paths below are explicit — but "exactly what queryPagesForBuild returned"
// stopped being true, and this guard's correctness rests on reading by path
// rather than on the shape being closed. A fix item minted by a
// discovery check carries only its own spec shape, which is why the name path
// exists and why an unresolved page must fail OPEN rather than block the loop.
//
// Returns ok=false when the page cannot be identified, which callers must treat
// as "carry on as before".
func resolveGuardedPage(ctx context.Context, db *sql.DB, collectedData map[string]interface{}, logger *zap.Logger) (pageID uuid.UUID, pageName string, ok bool) {
	if db == nil {
		return uuid.Nil, "", false
	}

	// Direct id paths first — no query needed, and no ambiguity.
	for _, idPath := range []string{"current_page.id", "current_item.spec.id"} {
		if raw := datahelpers.ExtractNestedFieldString(collectedData, idPath); raw != "" {
			if parsed, err := uuid.Parse(raw); err == nil {
				name := datahelpers.ExtractNestedFieldString(collectedData, "current_page.name")
				if name == "" {
					name = datahelpers.ExtractNestedFieldString(collectedData, "current_item.spec.name")
				}
				return parsed, name, true
			}
		}
	}

	// Name paths need the site to disambiguate: page names are unique per site,
	// not globally.
	siteIDStr := datahelpers.ExtractNestedFieldString(collectedData, "site_record.site_id")
	if siteIDStr == "" {
		siteIDStr = datahelpers.ExtractNestedFieldString(collectedData, "site_id")
	}
	siteID, err := uuid.Parse(siteIDStr)
	if err != nil {
		return uuid.Nil, "", false
	}

	for _, namePath := range []string{"current_page.name", "current_item.spec.name", "current_item.spec.page_name"} {
		name := datahelpers.ExtractNestedFieldString(collectedData, namePath)
		if name == "" {
			continue
		}
		// Reuses save_page_sections' own lookup so the two guards resolve a page
		// the same way by construction.
		id, _, lookupErr := saveSectionsLookupPageID(ctx, db, siteID, name)
		if lookupErr != nil {
			logger.Debug("resolveGuardedPage: name lookup missed",
				zap.String("path", namePath), zap.String("page_name", name), zap.Error(lookupErr))
			continue
		}
		return id, name, true
	}

	return uuid.Nil, "", false
}

// pageRefusesGenericBuild answers the one question every generic composition
// path needs: **may the generic builder write this page?** **This is the ONLY
// place any pipeline may read that from** — SavePageSectionsAction's
// migration-164 refusal goes through it too. If you are tempted to inline
// `SELECT rebuild_policy` in a new caller, call this instead: two predicates for
// "may we build this" is precisely the drift class this guard exists to prevent,
// and the council's `reuse_agent` seat objected on exactly that ground when this
// file first duplicated the save path's inline query.
//
// It was named pageIsOwnedForGuard until 2026-09-03 (bugs_open/450), when the
// answer stopped being reducible to ownership: a tool page with no tool on it is
// equally not the generic builder's to compose, for an unrelated reason. The
// rename is deliberate — it compile-forces every call site to be re-read rather
// than letting one keep asking the narrower question by accident.
//
// Returns (refused, class, checked). **`checked=false` means the page row could
// not be READ, not that the page is fair game** — indistinguishable in the bool
// alone, which is why the third return exists and why callers must report it.
//
// Fails OPEN on any error, matching SavePageSectionsAction's long-standing posture
// ("a scan error means an older schema, in which case the guard stands down"). A
// guard that failed CLOSED would stop generic page building fleet-wide the first
// time this query hiccupped — worse than the harm it prevents — and the
// selection-level exclusion already carries the load for the two loops with most
// of the traffic.
//
// The council's `bug_historian` seat objected (medium) that fail-open leaves a
// SILENT window: an owned page whose lookup times out is composed and committed,
// logged only. Answered by making the window LOUD rather than by failing closed —
// `checked=false` is logged at ERROR and persisted to `agent_error_log` as
// OWNED_PAGE_GUARD_UNCHECKED by the assemble call site, so the window is countable
// rather than hoped-about:
//
//	SELECT count(*), max(created_at) FROM agent_error_log
//	WHERE error_code = 'OWNED_PAGE_GUARD_UNCHECKED';
func pageRefusesGenericBuild(ctx context.Context, db *sql.DB, pageID uuid.UUID, logger *zap.Logger) (refused bool, class string, checked bool) {
	// The nil test happens HERE, before db is passed as an interface value.
	// A nil *sql.DB boxed into dbQueryer is NOT == nil, so readGenericBuildPolicy
	// cannot make this check for us — it would panic instead.
	if db == nil || pageID == uuid.Nil {
		return false, "", false
	}
	policy, toolShell, err := readGenericBuildPolicy(ctx, db, pageID)
	if err != nil {
		logger.Error("ownedPageGuard: page build policy UNREADABLE — guard standing down, page treated as generic",
			zap.String("page_id", pageID.String()), zap.Error(err))
		return false, "", false
	}
	refused, class = genericBuildRefusal(policy, toolShell)
	return refused, class, true
}

// readGenericBuildPolicy is the policy SQL itself, taken out of
// pageRefusesGenericBuild so a caller holding a *sql.Tx can run the same
// statement without opening a second connection — writeWorkItem's door
// (bugs_open/333) reads the policy inside the caller's transaction, and a
// *sql.DB read there would be outside it.
//
// Extracting the query rather than adding a second one is the whole point: the
// comment above says this file is the ONLY place a pipeline may read build
// policy, and that stays literally true — there is one statement, and both
// postures call it.
//
// ONE ROUND TRIP, not two. The shell test rides the existing primary-key read as
// a second output column rather than becoming a follow-up query, so the door in
// writeWorkItem — which every work-item write crosses — pays no extra
// round trip. The EXISTS subplan only executes for tool-typed rows; for
// everything else the planner short-circuits on the page_type test.
//
// IT DELIBERATELY RETURNS THE ERROR INSTEAD OF A `checked` BOOL, because the two
// callers want opposite things from `sql.ErrNoRows`:
//
//   - pageRefusesGenericBuild (the composition guards) treats ANY error as
//     "unreadable" — checked=false, logged at ERROR, counted as
//     OWNED_PAGE_GUARD_UNCHECKED. A page that a live composition step is holding
//     and that does not resolve is a real anomaly there.
//   - the writeWorkItem door treats ErrNoRows as "not refused", at Debug. A work
//     item can legitimately carry a page_id that no longer resolves — the
//     config-driven create_work_item takes whatever `input_data.spec.page_id`
//     says (create_work_item_action.go) — and a page that does not exist cannot
//     be owned by anything. Failing loud there would demote real findings on the
//     strength of a stale id.
//
// Collapsing those two into one bool is what would make this a shared helper
// with a wrong answer for one of its callers.
func readGenericBuildPolicy(ctx context.Context, q dbQueryer, pageID uuid.UUID) (policy string, toolShell bool, err error) {
	if err := q.QueryRowContext(ctx, `
		SELECT COALESCE(pages.rebuild_policy, 'generic'),
		       `+toolShellPredicateFor("pages")+`
		FROM pages WHERE pages.id = $1
	`, pageID).Scan(&policy, &toolShell); err != nil {
		return "", false, err
	}
	return policy, toolShell, nil
}

// emitOwnedPageReviewItem records that a generic build was refused for an owned
// page, so the request does not vanish silently.
//
// Deliberately the SAME item_type and item_key namespace ReconcileSitePlanAction
// already uses (reconcile_site_plan_action.go:238-266): one deterministic key per
// page, arbitrated by the partial unique index on (site_id, item_key) over open
// statuses, with ON CONFLICT DO NOTHING. So repeated dispatches converge on one
// row rather than accumulating, and our emission dedups against reconcile's for
// free instead of competing with it.
//
// Errors are logged and swallowed: a guard must never fail because its reporting
// failed, and the refusal itself is what protects the page.
func emitOwnedPageReviewItem(ctx context.Context, db *sql.DB, siteID uuid.UUID, pageName, source, reason, class string, logger *zap.Logger) {
	if db == nil || siteID == uuid.Nil || pageName == "" {
		return
	}
	if class == "" {
		class = refusalOwned
	}

	spec := map[string]interface{}{
		"page_name":     pageName,
		"refusal_class": class,
		"reason":        reason,
		"refused_by":    source,
		"fix":           refusalFixAdvice(class),
	}
	// Only the owned class may assert a policy value. Stamping
	// "rebuild_policy": "owned" on a tool_pending row would be a falsehood — the
	// page IS 'generic', and that is precisely why the old guard let it through.
	if class == refusalOwned {
		spec["rebuild_policy"] = ownedRebuildPolicy
	}
	specJSON, err := json.Marshal(spec)
	if err != nil {
		logger.Warn("ownedPageGuard: could not marshal review spec", zap.Error(err))
		return
	}

	if _, err := db.ExecContext(ctx, `
		INSERT INTO site_work_items (
			site_id, source, pipeline, item_type, severity, summary,
			spec, priority, status, created_by, item_key
		) VALUES ($1, $2, 'build', 'owned_page_review',
		          'high', $3, $4::jsonb, 30,
		          'needs_human_review', $2, $5)
		ON CONFLICT DO NOTHING
	`, siteID, source,
		refusalSummary(class, pageName),
		string(specJSON),
		"owned_page_review:"+pageName,
	); err != nil {
		logger.Warn("ownedPageGuard: could not emit owned_page_review item",
			zap.String("page_name", pageName), zap.Error(err))
		return
	}

	logger.Info("ownedPageGuard: owned_page_review recorded",
		zap.String("page_name", pageName), zap.String("source", source),
		zap.String("refusal_class", class))
}

// refusalSummary and refusalFixAdvice keep the human-facing halves of the two
// refusal classes side by side, because they are the only thing an operator
// reading the queue actually gets. The item_type and item_key stay shared
// (owned_page_review:<name>) so a page collects ONE row however it was refused
// and however many producers tried — the dedup arbitration reconcile's own
// emitter already relies on.
func refusalSummary(class, pageName string) string {
	if class == refusalToolPending {
		return fmt.Sprintf("Tool page %s has no tool component yet — excluded from generic rebuilds so it is not filled with prose about the tool", pageName)
	}
	return fmt.Sprintf("Owned page %s was excluded from a generic rebuild — needs owner-aware handling, not the generic builder", pageName)
}

func refusalFixAdvice(class string) string {
	if class == refusalToolPending {
		return "This page is typed 'tool' but carries no tool component, so the generic builder " +
			"would publish prose about a tool that is not there (bugs_open/450). Do NOT route it " +
			"to the generic page builder. The tool pipeline creates the component itself " +
			"(design-discovery evaluate_tools -> tool-suggester add_tool -> tool-deployer), after " +
			"which this refusal STOPS ON ITS OWN — it is derived from the page's contents, not a " +
			"flag anyone has to clear. If the page is not really a tool page, correct " +
			"pages.page_type; if the tool is wanted now, mint an add_tool item for it."
	}
	return "Owned/interactive page was excluded from a generic rebuild. Do NOT route it to " +
		"the generic page builder — it produces a prose page where an interactive tool " +
		"belongs. Rebuild via the tool pipeline (tool-generator/create_tool_component), " +
		"edit via apply_section_edit/section-editor, or change pages.rebuild_policy if the " +
		"page genuinely is the pipeline's to compose."
}

// censusExcludedOwnedPages names the owned pages that the selection WOULD have
// returned had the ownership exclusion not been applied, and records a review item
// for each.
//
// Without this the exclusion is silent: the page stays at needs_rebuild, is
// re-excluded on every subsequent run, and an operator who explicitly asked for it
// gets neither the rebuild nor a reason. The predicate is deliberately the exact
// inverse of ownedPageExclusionSQL — same status filter, same liveness filter, the
// ownership test flipped — so it cannot drift into reporting a different population
// from the one that was excluded.
//
// Returns the page names for the action result. Errors are logged and swallowed:
// reporting must never fail a selection.
func censusExcludedRefusedPages(ctx context.Context, db *sql.DB, siteID uuid.UUID, statuses []string, includeAll bool, source string, logger *zap.Logger) []string {
	if db == nil || siteID == uuid.Nil {
		return nil
	}

	// The exact inverse of genericBuildExclusionSQL: that clause is
	// `AND NOT (owned OR shell)`, so the census selects `(owned OR shell)`, with
	// the shell disjunct present exactly when the exclusion carries it. Reading
	// the class per row lets the receipt say WHICH reason applied.
	refused := `(COALESCE(pages.rebuild_policy, 'generic') = 'owned' OR ` + toolShellPredicateFor("pages") + `)`
	classExpr := `CASE WHEN COALESCE(pages.rebuild_policy, 'generic') = 'owned' THEN '` + refusalOwned +
		`' ELSE '` + refusalToolPending + `' END`
	if !toolShellRefusalArmed() {
		refused = `(COALESCE(pages.rebuild_policy, 'generic') = 'owned')`
		classExpr = `'` + refusalOwned + `'`
	}

	query := `SELECT pages.name, ` + classExpr + ` FROM pages
	          WHERE pages.site_id = $1 AND pages.status = 'active'
	            AND ` + refused + `
	          ORDER BY pages.name`
	args := []interface{}{siteID}

	if !includeAll {
		placeholders := make([]string, len(statuses))
		for i, s := range statuses {
			placeholders[i] = fmt.Sprintf("$%d", i+2)
			args = append(args, s)
		}
		query = fmt.Sprintf(`SELECT pages.name, %s FROM pages
		          WHERE pages.site_id = $1 AND pages.status = 'active'
		            AND COALESCE(pages.build_status, 'planned') IN (%s)
		            AND %s
		          ORDER BY pages.name`, classExpr, strings.Join(placeholders, ", "), refused)
	}

	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		logger.Warn("ownedPageGuard: excluded-page census failed", zap.Error(err))
		return nil
	}
	defer rows.Close()

	var names []string
	classOf := map[string]string{}
	for rows.Next() {
		var name, class string
		if scanErr := rows.Scan(&name, &class); scanErr != nil {
			logger.Warn("ownedPageGuard: census scan failed", zap.Error(scanErr))
			continue
		}
		names = append(names, name)
		classOf[name] = class
	}
	if rowsErr := rows.Err(); rowsErr != nil {
		logger.Warn("ownedPageGuard: census iteration failed", zap.Error(rowsErr))
	}
	// Close before the INSERTs: holding the result set open while writing on the
	// same *sql.DB can pin a connection unnecessarily.
	rows.Close()

	if len(names) == 0 {
		return nil
	}

	logger.Warn("ownedPageGuard: pages excluded from a generic build selection",
		zap.String("site_id", siteID.String()),
		zap.Strings("pages", names),
		zap.String("source", source),
	)

	for _, name := range names {
		class := classOf[name]
		reason := ownedPageSkipReasonPrefix + ": excluded at selection — an owned page is not the generic builder's to compose"
		if class == refusalToolPending {
			reason = ownedPageSkipReasonPrefix + ": excluded at selection — a tool page with no tool component is not the generic builder's to compose"
		}
		emitOwnedPageReviewItem(ctx, db, siteID, name, source, reason, class, logger)
	}

	return names
}

// upstreamAssemblySkipped reports whether the assemble step for THIS iteration
// declared a skip, and why.
//
// Reads the same collected_data shape checkUpstreamSkipped's first branch reads
// (assembled_page.skipped), which every live consumer populates: all three
// assemble_page steps declare output_field "assembled_page", measured 2026-08-06.
func upstreamAssemblySkipped(collectedData map[string]interface{}) (bool, string) {
	assembled, ok := collectedData["assembled_page"].(map[string]interface{})
	if !ok {
		return false, ""
	}
	skipped, _ := assembled["skipped"].(bool)
	if !skipped {
		return false, ""
	}
	reason, _ := assembled["skip_reason"].(string)
	return true, reason
}
