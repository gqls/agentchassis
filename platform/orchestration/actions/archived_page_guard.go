// FILE: platform/orchestration/actions/archived_page_guard.go
//
// Deploy-time guard for `pages.status='archived'` (bugs_open/266).
//
// An archived page was rebuilt and re-stamped `deployed` by at least FOUR
// independent producers, none of which read `pages.status` — measured
// 2026-08-12 against `site_work_items`, each producer's completion matched to
// the row's own `deployed_at`:
//
//	reconcile_site_plan  -> needs_page (not_built)
//	image-build-handler  -> needs_page (image_landed)
//	page-rerender        -> page_rerender (cta_links_stale)
//	section-editor       -> section_edit
//
// The load-bearing observation is that for one of the two pages
// `ReconcileSitePlanAction` behaved CORRECTLY — it withheld the build and filed
// `owned_page_review`/needs_human_review, where it still sits — and
// `image-build-handler` rebuilt and deployed the page sixteen minutes later by
// an unrelated path. A predicate added to the producer named in the symptom
// would have closed the one door that was already shut.
//
// WHY git_commit, AND WHY THAT IS THE OPPOSITE CHOICE TO owned_page_guard.
// The guard next door sits at `assemble_page` and says in terms why it avoids
// git_commit: "git_commit is also how owned pages LEGITIMATELY deploy" —
// page-rerender (rerender_single_page) and section-editor (apply_section_edit)
// commit pages without passing through assemble_page, and migration 164 wants
// those to keep working. That reasoning is correct FOR OWNED and inverts here:
//
//	owned    = "not the generic pipeline's to rebuild"  -> some deploys are legitimate
//	archived = "nothing may deploy this"                -> no deploy is legitimate
//
// A state whose prohibition has no exceptions must be guarded at the seam every
// producer passes through, which is exactly the seam the exception-bearing state
// had to avoid. Measured at the live workflow definitions 2026-08-12: every one
// of the four producers above reaches `git_commit` — page-rerender and
// section-editor each have their own `deploy_page`/`git_commit` step, and
// page-build-handler's `deploy_page` is a `call_agent` to target_role
// `page_renderer`, i.e. into page-rerender's own git_commit.
//
// Placing this at `assemble_page` instead would have been INERT, not merely
// partial: no live, active, non-snapshot agent definition has a step whose
// action is `assemble_page` (same measurement). That is the owned guard's
// problem to weigh, not this file's — recorded in bugs_open/266 so it is not
// lost — but it is why the tempting "copy the neighbour" fix was rejected.
//
// RETRACTION IS NOT AFFECTED, and this was checked rather than assumed:
// bugs_closed/098's retraction runs `retract_page_deployment`, which dispatches
// `delete_file` — a different path from git_commit. Guarding git_commit cannot
// stop a retraction removing an archived page's file, which is the one deploy-ish
// operation on an archived page that MUST keep working.
//
// FAIL-OPEN, deliberately, matching owned_page_guard's posture and for its
// reason: a guard that failed closed would stop page deployment fleet-wide the
// first time this query hiccupped, which is worse than the harm it prevents.
// The window is made countable rather than silent — see ARCHIVED_PAGE_GUARD_UNCHECKED.
package actions

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"go.uber.org/zap"
)

// archivedPageStatus is the `pages.status` value meaning "retired; nothing may
// deploy this". Note this is `status`, NOT `build_status` — the two are
// different columns with overlapping vocabularies, and `build_status='deployed'`
// on a `status='archived'` row is precisely the state 266 is about.
const archivedPageStatus = "archived"

// archivedPageSkipReasonPrefix marks a skip caused by this guard. Matched by
// nothing — the skip protocol is a boolean — but it is the string to pod-grep to
// prove the guard is live on a running binary, and it tells an operator reading a
// skipped iteration WHY the deploy was refused.
const archivedPageSkipReasonPrefix = "ARCHIVED_PAGE_GUARD"

// resolveDeployTargetPage finds the page a DEPLOY step is working on.
//
// It is a superset of resolveGuardedPage rather than a change to it, on purpose.
// resolveGuardedPage knows the three composition loops' shapes
// (`current_page.*`, `current_item.spec.*`); the deploy seam additionally sees
// the flat dispatch shape that page-rerender and section-editor are driven with
// (`page_id` / `site_id` / `page_name`, per page-build-handler's `input_mapping`
// and RerenderSinglePageAction's own `input_fields` default). Widening
// resolveGuardedPage itself would change what reaches the OWNED guard, which is
// another bug's shipped fix — widening what reaches a function is a behaviour
// change to it even when the function is untouched.
//
// Returns ok=false when the page cannot be identified, which callers MUST treat
// as "carry on as before". An unidentifiable page is the normal case for the
// many git_commit steps that deploy no page at all (CSS, JS snippets, RSS,
// reports, whole-site deploys) — this guard must be invisible to them.
func resolveDeployTargetPage(ctx context.Context, db *sql.DB, collectedData map[string]interface{}, logger *zap.Logger) (pageID uuid.UUID, pageName string, ok bool) {
	if db == nil {
		return uuid.Nil, "", false
	}

	// 1. The composition-loop shapes, via the shared resolver, so a page
	//    identified at assemble time and at deploy time is the SAME page by
	//    construction.
	if id, name, found := resolveGuardedPage(ctx, db, collectedData, logger); found {
		return id, name, true
	}

	// 2. The flat dispatch shape. `input_data.page_id` is tried as well as the
	//    bare key because a called agent's payload is nested under input_data by
	//    the coordinator, and both spellings are live.
	for _, idPath := range []string{"page_id", "input_data.page_id"} {
		if raw := datahelpers.ExtractNestedFieldString(collectedData, idPath); raw != "" {
			if parsed, err := uuid.Parse(raw); err == nil {
				name := datahelpers.ExtractNestedFieldString(collectedData, "page_name")
				if name == "" {
					name = datahelpers.ExtractNestedFieldString(collectedData, "input_data.page_name")
				}
				return parsed, name, true
			}
		}
	}

	// 3. (site_id, page_name) — reuses RerenderSinglePageAction's own resolver, so
	//    the guard and the action it guards resolve a page identically. Exact, not
	//    a guess: pages carries UNIQUE (site_id, name).
	for _, sitePath := range []string{"site_id", "input_data.site_id", "site_record.site_id"} {
		siteIDStr := datahelpers.ExtractNestedFieldString(collectedData, sitePath)
		if siteIDStr == "" {
			continue
		}
		for _, namePath := range []string{"page_name", "input_data.page_name"} {
			name := datahelpers.ExtractNestedFieldString(collectedData, namePath)
			if name == "" {
				continue
			}
			id, err := resolvePageIDByName(ctx, db, siteIDStr, name)
			if err != nil {
				logger.Debug("resolveDeployTargetPage: name lookup missed",
					zap.String("site_path", sitePath), zap.String("page_name", name), zap.Error(err))
				continue
			}
			return id, name, true
		}
	}

	return uuid.Nil, "", false
}

// pageIsArchivedForGuard reads the retirement marker. **This is the ONLY place
// any pipeline may read deploy-time archived status from.** If you are tempted
// to inline `SELECT status` in a new caller, call this instead: two predicates
// for "is this page archived" is the drift class the neighbouring guard was
// objected to for, by the council's `reuse_agent` seat, on exactly this ground.
//
// Returns (archived, checked). **`checked=false` means the status could not be
// READ, not that the page is live** — indistinguishable in the bool alone, which
// is why the second return exists and why callers must report it.
//
// Fails OPEN on any error (see the file header). The window is logged at ERROR
// and persisted as ARCHIVED_PAGE_GUARD_UNCHECKED by the call sites, so it is
// countable rather than hoped-about:
//
//	SELECT count(*), max(occurred_at) FROM agent_error_log
//	WHERE error_code = 'ARCHIVED_PAGE_GUARD_UNCHECKED';
func pageIsArchivedForGuard(ctx context.Context, db *sql.DB, pageID uuid.UUID, logger *zap.Logger) (archived bool, checked bool) {
	if db == nil || pageID == uuid.Nil {
		return false, false
	}
	var status string
	if err := db.QueryRowContext(ctx, `
		SELECT COALESCE(status, '') FROM pages WHERE id = $1
	`, pageID).Scan(&status); err != nil {
		// sql.ErrNoRows lands here too, and correctly: a page that does not exist
		// is not a page this guard may refuse a deploy for. It is still unchecked,
		// and still reported.
		logger.Warn("pageIsArchivedForGuard: could not read pages.status; guard stands down for this page",
			zap.String("page_id", pageID.String()), zap.Error(err))
		return false, false
	}
	return status == archivedPageStatus, true
}
