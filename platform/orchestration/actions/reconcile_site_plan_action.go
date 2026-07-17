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
//     pages_skipped_built: <int>,   already deployed at current plan version
//     pages_skipped_queued: <int>,  open work item already exists
//     rerender_emitted: <bool>,
//     batch_id: <uuid>,
//   }
//
// Decision per plan page (in order):
//   1. no row in pages / not deployed / stale plan  → candidate (else skip)
//   2. tool/game role OR pages.rebuild_policy='owned'
//        → emit owned_page_review (needs_human_review, NO handler) — the
//          generic builder clobbers owned pages (TP-004 / TL-001; guard
//          rail 1 of the experience loop)
//   3. existing open item for the key               → skip (queued)
//   4. otherwise                                    → emit needs_page
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
			"pages_skipped_queued": 0,
			"rerender_emitted":     false,
		}, nil
	}

	realised, err := loadRealisedPages(ctx, params.DB, siteID)
	if err != nil {
		return nil, fmt.Errorf("load realised pages: %w", err)
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

		decision := decideEmit(plan, real, planID)
		if decision == "skip_built" {
			skippedBuilt++
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
		zap.Int("pages_skipped_queued", skippedQueued),
		zap.Bool("rerender_emitted", rerenderEmitted),
	)

	return map[string]interface{}{
		"plan_id":              planID.String(),
		"pages_total":          len(planPages),
		"pages_emitted":        emitted,
		"pages_review_emitted": emittedReview,
		"pages_skipped_built":  skippedBuilt,
		"pages_skipped_queued": skippedQueued,
		"rerender_emitted":     rerenderEmitted,
		"batch_id":             batchID.String(),
	}, nil
}

// decideEmit returns one of: "missing", "not_built", "stale", "skip_built".
// Inputs:
//   - plan: the plan page record (always present — caller filters)
//   - realised: the realised pages row, may be zero-value if absent
//   - planID: the current plan id, for stale check
func decideEmit(plan planPageRecord, realised realisedPageRecord, planID uuid.UUID) string {
	if realised.BuildStatus == "" {
		return "missing" // no row in pages table
	}
	if realised.BuildStatus != "deployed" {
		return "not_built"
	}
	if !realised.BuiltFromPlanVersion.Valid {
		return "stale" // built but no plan-version recorded; treat as stale until set
	}
	if realised.BuiltFromPlanVersion.UUID != planID {
		return "stale"
	}
	return "skip_built"
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
