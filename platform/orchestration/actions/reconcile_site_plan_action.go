// FILE: platform/orchestration/actions/reconcile_site_plan_action.go
//
// ReconcileSitePlanAction reads the current plan for a site and emits
// needs_page:<name> work items for any plan page that is missing,
// not deployed, or built from an older plan version. Also emits a
// terminal needs_rerender item if any page items were emitted.
//
// No LLM. Pure read-from-DB, decide, write-work-items. Idempotent: the
// existing idx_swi_dedup partial unique index ensures repeated runs
// don't create duplicate work items for the same (site_id, item_key).
//
// Inputs (CollectedData):
//   - target_site_id  (declared in spec; deliberately not 'site_id' to
//                      avoid the nested-lookup collision discussed in
//                      doc 001's Field-name-collisions section)
//   - plan_id         (optional; if absent, reads the current plan from
//                      site_plans WHERE is_current = true)
//
// Returns:
//   {
//     plan_id: <uuid>,
//     pages_total: <int>,           total plan pages considered
//     pages_emitted: <int>,         needs_page items written
//     pages_skipped_built: <int>,   deployed & unchanged (includes re-stamped)
//     pages_restamped: <int>,       deployed, older plan, composition unchanged
//                                   → built_from_plan_version carried forward
//     pages_skipped_queued: <int>,  open work item already exists
//     rerender_emitted: <bool>,
//     batch_id: <uuid>,
//   }
//
// Decision per plan page (in order):
//   1. deployed at current plan version, OR deployed at an OLDER version but the
//      new plan proposes the same composition  → skip (the latter re-stamps
//      built_from_plan_version forward; /bugs_open/038)
//   2. no row in pages / not deployed / genuinely re-composed → candidate
//   3. tool/game role OR pages.rebuild_policy='owned'
//        → emit owned_page_review (needs_human_review, NO handler) — the
//          generic builder clobbers owned pages (TP-004 / TL-001; guard
//          rail 1 of the experience loop)
//   4. existing open item for the key               → skip (queued)
//   5. otherwise                                    → emit needs_page
//
// "Open" = status NOT IN ('complete','verified','rejected','wont_fix','failed').
// Same set the dedup index uses; consistent semantics.
//
// Updates sites.last_reconciled_at on success so the future scheduled
// tick can skip recently-reconciled sites.

package actions

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"sort"

	"github.com/google/uuid"
	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"go.uber.org/zap"
)

var ReconcileSitePlanInputSpec = datahelpers.ActionInputSpec{
	Required:   []string{"target_site_id"},
	Optional:   []string{"plan_id"},
	Defaults:   map[string]interface{}{},
	Deprecated: map[string]string{},
}

func init() {
	datahelpers.RegisterActionInputSpec("reconcile_site_plan", ReconcileSitePlanInputSpec)
}

// planPageRecord is what we read from site_plan_pages for the diff.
type planPageRecord struct {
	Name     string
	Role     string
	Title    sql.NullString
	NavOrder sql.NullInt32
}

// realisedPageRecord is what we read from pages for the diff.
type realisedPageRecord struct {
	BuildStatus          string
	BuiltFromPlanVersion uuid.NullUUID
	RebuildPolicy        string
}

// ReconcileSitePlanAction handler.
func ReconcileSitePlanAction(ctx context.Context, params ActionParams) (interface{}, error) {
	logger := params.Logger.With(zap.String("action", "reconcile_site_plan"))
	logger.Info("ReconcileSitePlanAction: Starting",
		zap.Any("collected_data_keys", datahelpers.GetMapKeys(params.CollectedData)),
	)

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}
	if params.DB == nil {
		return nil, fmt.Errorf("database connection required")
	}

	inputs, err := datahelpers.ExtractActionInputs(
		params.CollectedData, params.StepConfig.Config,
		ReconcileSitePlanInputSpec, logger,
	)
	if err != nil {
		return nil, fmt.Errorf("input extraction failed: %w", err)
	}

	siteID, err := uuid.Parse(inputs.Get("target_site_id"))
	if err != nil {
		return nil, fmt.Errorf("invalid target_site_id: %w", err)
	}

	// ── 1. Resolve plan_id ─────────────────────────────────────────────
	var planID uuid.UUID
	if planIDStr := inputs.Get("plan_id"); planIDStr != "" {
		planID, err = uuid.Parse(planIDStr)
		if err != nil {
			return nil, fmt.Errorf("invalid plan_id %q: %w", planIDStr, err)
		}
	} else {
		err = params.DB.QueryRowContext(ctx, `
			SELECT id FROM site_plans
			WHERE site_id = $1 AND is_current = true
		`, siteID).Scan(&planID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return nil, fmt.Errorf("no current site_plan for site %s", siteID)
			}
			return nil, fmt.Errorf("load current plan: %w", err)
		}
	}

	logger.Info("ReconcileSitePlanAction: resolved plan",
		zap.String("site_id", siteID.String()),
		zap.String("plan_id", planID.String()),
	)

	// ── 2. Read plan pages, realised pages, open work items ───────────
	planPages, err := loadPlanPages(ctx, params.DB, planID)
	if err != nil {
		return nil, fmt.Errorf("load plan pages: %w", err)
	}
	if len(planPages) == 0 {
		logger.Warn("ReconcileSitePlanAction: plan has zero pages, nothing to reconcile",
			zap.String("plan_id", planID.String()))
		return map[string]interface{}{
			"plan_id":              planID.String(),
			"pages_total":          0,
			"pages_emitted":        0,
			"pages_skipped_built":  0,
			"pages_restamped":      0,
			"pages_skipped_queued": 0,
			"rerender_emitted":     false,
		}, nil
	}

	realised, err := loadRealisedPages(ctx, params.DB, siteID)
	if err != nil {
		return nil, fmt.Errorf("load realised pages: %w", err)
	}

	// Section lists for the current plan and each version a deployed page was
	// built from, so decideEmit can tell a genuine re-composition (rebuild) from
	// an unchanged page (re-stamp + skip). /bugs_open/038.
	planSections, err := loadPlanSectionsByVersion(ctx, params.DB, planID, realised)
	if err != nil {
		return nil, fmt.Errorf("load plan sections: %w", err)
	}

	openItems, err := loadOpenPageItems(ctx, params.DB, siteID)
	if err != nil {
		return nil, fmt.Errorf("load open work items: %w", err)
	}

	// ── 3. Diff and emit ──────────────────────────────────────────────
	batchID := uuid.New()
	tx, err := params.DB.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	emitted := 0
	emittedReview := 0
	skippedBuilt := 0
	restamped := 0
	skippedQueued := 0

	// Iterate pages in name order so priorities are deterministic across runs.
	names := make([]string, 0, len(planPages))
	for name := range planPages {
		names = append(names, name)
	}
	sort.Strings(names)

	for _, name := range names {
		plan := planPages[name]
		real := realised[name]

		var builtFromSections []string
		if real.BuiltFromPlanVersion.Valid {
			builtFromSections = planSections[real.BuiltFromPlanVersion.UUID][name]
		}
		decision := decideEmit(real, planID, planSections[planID][name], builtFromSections)
		if decision == "skip_built" {
			skippedBuilt++
			continue
		}
		if decision == "restamp" {
			// Deployed from an older plan version, but the new plan proposes the
			// same composition — carry the stamp forward instead of rebuilding,
			// which would only regenerate identical copy (/bugs_open/038). Counts
			// as skipped-from-build so pages_skipped_built stays the total of
			// deployed-and-unchanged pages.
			if _, err = tx.ExecContext(ctx, `
				UPDATE pages SET built_from_plan_version = $1, updated_at = NOW()
				WHERE site_id = $2 AND name = $3
			`, planID, siteID, name); err != nil {
				return nil, fmt.Errorf("re-stamp built_from_plan_version for %q: %w", name, err)
			}
			skippedBuilt++
			restamped++
			continue
		}

		// Page-ownership guard (guard rail 1, experience loop). Tool/game-role
		// and rebuild_policy='owned' pages must never route to the generic
		// page builder — it produces a widget-less prose page where an
		// interactive tool belongs (TP-004; the vonc arena clobber). Emit a
		// review item with NO handler instead, mirroring
		// check_incomplete_page_group's tool-role branch.
		if plan.Role == "tool" || plan.Role == "game" || real.RebuildPolicy == "owned" {
			reviewKey := "owned_page_review:" + name
			if _, queued := openItems[reviewKey]; queued {
				skippedQueued++
				continue
			}
			spec := map[string]interface{}{
				"page_name":      name,
				"page_role":      plan.Role,
				"rebuild_policy": real.RebuildPolicy,
				"plan_id":        planID.String(),
				"reason":         decision,
				"fix": "Owned/interactive page is " + decision + ". Do NOT route to the " +
					"generic page builder. Build via the tool pipeline " +
					"(tool-generator/create_tool_component) or the owning experience " +
					"spec, or remove the page from the plan.",
			}
			specJSON, _ := json.Marshal(spec)
			_, err = tx.ExecContext(ctx, `
				INSERT INTO site_work_items (
					site_id, source, pipeline, item_type, severity, summary,
					spec, priority, status, created_by, item_key, batch_id
				) VALUES ($1, 'reconcile_site_plan', 'build', 'owned_page_review',
				          'high', $2, $3::jsonb, 30,
				          'needs_human_review', 'reconcile_site_plan', $4, $5)
				ON CONFLICT DO NOTHING
			`, siteID, fmt.Sprintf("Owned page %s is %s — needs owner-aware build, not the generic builder", name, decision),
				string(specJSON), reviewKey, batchID)
			if err != nil {
				return nil, fmt.Errorf("emit owned_page_review for %q: %w", name, err)
			}
			emittedReview++
			continue
		}

		itemKey := "needs_page:" + name
		if _, queued := openItems[itemKey]; queued {
			skippedQueued++
			continue
		}

		// emit
		spec := map[string]interface{}{
			"page_name": name,
			"page_role": plan.Role,
			"plan_id":   planID.String(),
			"reason":    decision,
		}
		specJSON, _ := json.Marshal(spec)

		summary := fmt.Sprintf("Build %s page (%s)", name, decision)
		priority := 50 + emitted // simple monotonic for ordering; not load-balancing

		_, err = tx.ExecContext(ctx, `
			INSERT INTO site_work_items (
				site_id, source, pipeline, item_type, severity, summary,
				spec, priority, handler_agent, status, created_by,
				item_key, batch_id
			) VALUES ($1, 'reconcile_site_plan', 'build', 'needs_page',
			          'medium', $2, $3::jsonb, $4, 'page-build-handler',
			          'triaged', 'reconcile_site_plan', $5, $6)
			ON CONFLICT DO NOTHING
		`, siteID, summary, string(specJSON), priority, itemKey, batchID)
		if err != nil {
			return nil, fmt.Errorf("emit needs_page for %q: %w", name, err)
		}
		emitted++
	}

	// ── 4. Rerender at end if any pages emitted ────────────────────────
	rerenderEmitted := false
	if emitted > 0 {
		rerenderSpec, _ := json.Marshal(map[string]interface{}{
			"refresh_site_components": true,
			"reason":                  "post_reconcile_assembly",
			"plan_id":                 planID.String(),
		})
		rerenderKey := fmt.Sprintf("reconcile_rerender:%s", planID.String())

		res, err := tx.ExecContext(ctx, `
			INSERT INTO site_work_items (
				site_id, source, pipeline, item_type, severity, summary,
				spec, priority, handler_agent, status, created_by,
				item_key, batch_id
			) VALUES ($1, 'reconcile_site_plan', 'build', 'needs_rerender',
			          'medium', 'Assemble and deploy pages after plan reconcile',
			          $2::jsonb, 99, 'rerender-pages',
			          'triaged', 'reconcile_site_plan', $3, $4)
			ON CONFLICT DO NOTHING
		`, siteID, string(rerenderSpec), rerenderKey, batchID)
		if err != nil {
			return nil, fmt.Errorf("emit needs_rerender: %w", err)
		}
		if n, _ := res.RowsAffected(); n > 0 {
			rerenderEmitted = true
		}
	}

	// ── 5. Update sites.last_reconciled_at ─────────────────────────────
	_, err = tx.ExecContext(ctx, `
		UPDATE sites SET last_reconciled_at = NOW() WHERE id = $1
	`, siteID)
	if err != nil {
		return nil, fmt.Errorf("update last_reconciled_at: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}

	logger.Info("ReconcileSitePlanAction: Complete",
		zap.String("site_id", siteID.String()),
		zap.String("plan_id", planID.String()),
		zap.String("batch_id", batchID.String()),
		zap.Int("pages_total", len(planPages)),
		zap.Int("pages_emitted", emitted),
		zap.Int("pages_review_emitted", emittedReview),
		zap.Int("pages_skipped_built", skippedBuilt),
		zap.Int("pages_restamped", restamped),
		zap.Int("pages_skipped_queued", skippedQueued),
		zap.Bool("rerender_emitted", rerenderEmitted),
	)

	return map[string]interface{}{
		"plan_id":              planID.String(),
		"pages_total":          len(planPages),
		"pages_emitted":        emitted,
		"pages_review_emitted": emittedReview,
		"pages_skipped_built":  skippedBuilt,
		"pages_restamped":      restamped,
		"pages_skipped_queued": skippedQueued,
		"rerender_emitted":     rerenderEmitted,
		"batch_id":             batchID.String(),
	}, nil
}

// decideEmit returns one of: "missing", "not_built", "stale", "skip_built",
// "restamp".
//
// Inputs:
//   - realised: the realised pages row, may be zero-value if absent
//   - planID: the current plan id, for the stale check
//   - currentSections:   the current plan's ordered, normalised section list
//     for this page (from site_plan_sections)
//   - builtFromSections: the same for the plan version the page was built from
//     (nil if that version has no sections for the page, or is unknown)
//
// "restamp" means: the page is deployed from an older plan version, but the new
// plan proposes the same composition, so it needs no rebuild — the caller
// carries built_from_plan_version forward and skips. This is the fix for
// /bugs_open/038: a re-plan writes a new site_plans row, so EVERY deployed page
// fails the plain built_from == planID equality and was being rebuilt (and its
// copy regenerated) even when the plan left the page untouched.
//
// The comparison is plan-to-plan on site_plan_sections, deliberately NOT
// plan-to-pages.sections. sync_pages runs before this action and overwrites
// pages.sections with the new plan's proposal (site_db_actions.go, `sections =
// EXCLUDED.sections`), so pages.sections always equals the current plan and
// cannot tell an unchanged page from a re-composed one. site_plan_sections is
// per-plan and immutable, so it can.
func decideEmit(realised realisedPageRecord, planID uuid.UUID, currentSections, builtFromSections []string) string {
	if realised.BuildStatus == "" {
		return "missing" // no row in pages table
	}
	if realised.BuildStatus != "deployed" {
		return "not_built"
	}
	if !realised.BuiltFromPlanVersion.Valid {
		return "stale" // built but no plan-version recorded; treat as stale until set
	}
	if realised.BuiltFromPlanVersion.UUID == planID {
		return "skip_built"
	}
	// Built from an older plan version. Only a genuine composition change is
	// stale; an unchanged page is re-stamped and skipped. Require the built-from
	// version to actually carry a section list for the page — a nil list means
	// we cannot positively confirm the old composition, so fall through to a
	// (safe) rebuild rather than skipping on absence of evidence.
	if builtFromSections != nil && sectionsEqual(currentSections, builtFromSections) {
		return "restamp"
	}
	return "stale"
}

// sectionsEqual reports whether two ordered, already-normalised section lists
// are identical. Order matters: it determines page layout.
func sectionsEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func loadPlanPages(ctx context.Context, db *sql.DB, planID uuid.UUID) (map[string]planPageRecord, error) {
	rows, err := db.QueryContext(ctx, `
		SELECT name, role, title, nav_order
		FROM site_plan_pages
		WHERE plan_id = $1
	`, planID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make(map[string]planPageRecord)
	for rows.Next() {
		var r planPageRecord
		if err := rows.Scan(&r.Name, &r.Role, &r.Title, &r.NavOrder); err != nil {
			return nil, err
		}
		out[r.Name] = r
	}
	return out, rows.Err()
}

func loadRealisedPages(ctx context.Context, db *sql.DB, siteID uuid.UUID) (map[string]realisedPageRecord, error) {
	rows, err := db.QueryContext(ctx, `
		SELECT name, COALESCE(build_status, ''), built_from_plan_version,
		       COALESCE(rebuild_policy, 'generic')
		FROM pages
		WHERE site_id = $1
	`, siteID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make(map[string]realisedPageRecord)
	for rows.Next() {
		var name string
		var r realisedPageRecord
		if err := rows.Scan(&name, &r.BuildStatus, &r.BuiltFromPlanVersion, &r.RebuildPolicy); err != nil {
			return nil, err
		}
		out[name] = r
	}
	return out, rows.Err()
}

// loadPlanSectionsByVersion loads, for the current plan and for every distinct
// plan version any DEPLOYED realised page was built from, the ordered and
// normalised component list per page name. Shape: map[planID]map[pageName][]string.
//
// The set of plan versions is small (the current plan plus a handful of
// build-time stamps), so one query per version is cheaper and clearer than a
// single ANY(...) with a driver-specific array literal.
func loadPlanSectionsByVersion(
	ctx context.Context,
	db *sql.DB,
	currentPlanID uuid.UUID,
	realised map[string]realisedPageRecord,
) (map[uuid.UUID]map[string][]string, error) {
	want := map[uuid.UUID]struct{}{currentPlanID: {}}
	for _, r := range realised {
		if r.BuildStatus == "deployed" && r.BuiltFromPlanVersion.Valid {
			want[r.BuiltFromPlanVersion.UUID] = struct{}{}
		}
	}

	out := make(map[uuid.UUID]map[string][]string, len(want))
	for planID := range want {
		secs, err := loadPlanSections(ctx, db, planID)
		if err != nil {
			return nil, fmt.Errorf("load plan sections for %s: %w", planID, err)
		}
		out[planID] = secs
	}
	return out, nil
}

// loadPlanSections returns the ordered, normalised component list per page name
// for a single plan version. Names are normalised with the same rule the rest
// of the build path uses (NormalizeComponentFunction), so a plan that wrote
// "call_to_action" and one that wrote "call-to-action" compare equal.
func loadPlanSections(ctx context.Context, db *sql.DB, planID uuid.UUID) (map[string][]string, error) {
	rows, err := db.QueryContext(ctx, `
		SELECT page_name, component_name
		FROM site_plan_sections
		WHERE plan_id = $1
		ORDER BY page_name, ordering
	`, planID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make(map[string][]string)
	for rows.Next() {
		var pageName, componentName string
		if err := rows.Scan(&pageName, &componentName); err != nil {
			return nil, err
		}
		out[pageName] = append(out[pageName], NormalizeComponentFunction(componentName))
	}
	return out, rows.Err()
}

// loadOpenPageItems returns a set of item_keys for needs_page items in
// open status (i.e. matching the dedup index's WHERE clause).
func loadOpenPageItems(ctx context.Context, db *sql.DB, siteID uuid.UUID) (map[string]struct{}, error) {
	rows, err := db.QueryContext(ctx, `
		SELECT item_key FROM site_work_items
		WHERE site_id = $1
		  AND item_type IN ('needs_page','owned_page_review')
		  AND item_key IS NOT NULL
		  AND status NOT IN ('complete','verified','rejected','wont_fix','failed')
	`, siteID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make(map[string]struct{})
	for rows.Next() {
		var k string
		if err := rows.Scan(&k); err != nil {
			return nil, err
		}
		out[k] = struct{}{}
	}
	return out, rows.Err()
}
