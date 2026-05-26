// FILE: platform/orchestration/actions/emit_design_items_action.go
//
// EmitDesignItemsAction queues the two work items that drive a site's visual
// design: needs_composition (→ site-design-planner, which resolves palette/
// layout/typography and installs the css_theme + style_collection, setting
// sites.style_collection_id) and needs_design (→ webdesign-agent, which renders
// and commits styles.css), with needs_design depends_on needs_composition.
//
// WHY THIS EXISTS
//   The legacy site-planner path emitted these inside WriteBuildItemsAction.
//   build-site-planner (Phase 1) moved its terminal step to write_site_plan +
//   reconcile_site_plan, neither of which emits the design trigger. This action
//   restores it as an explicit, named PLAN-TIME step in build-site-planner.
//
// WHY PLAN-TIME (not reconcile)
//   reconcile_site_plan also runs on the scheduled reconcile tick — emitting
//   from there would backfill every existing NULL-collection site on a cadence.
//   This action runs only when the planner runs (fresh plan / explicit replan),
//   so it never backfills. Stale already-built sites are the improvement loop's
//   job via the missing_css discovery check.
//
// GUARD
//   Emits only when sites.style_collection_id IS NULL. install_site_composition
//   sets style_collection_id and refuses to re-run, so a replan of an already-
//   composed site is a no-op.
//
// REUSE NOTE
//   The composition+design insert block mirrors WriteBuildItemsAction's
//   (the canonical source). If kept long-term, extract a shared helper
//   emitInitialCompositionAndDesign(ctx, tx, siteID, batchID, createdBy, logger)
//   and call it from both, so the two copies cannot drift.
//
// REGISTRATION (registry.go):
//   "emit_design_items": {
//       Handler:     EmitDesignItemsAction,
//       Category:    "site",
//       Description: "Queue needs_composition + needs_design for a fresh build",
//       IsLocal:     true,
//   }

package actions

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"go.uber.org/zap"
)

var EmitDesignItemsInputSpec = datahelpers.ActionInputSpec{
	Required:   []string{"site_id"},
	Optional:   []string{},
	Defaults:   map[string]interface{}{},
	Deprecated: map[string]string{},
}

func init() {
	datahelpers.RegisterActionInputSpec("emit_design_items", EmitDesignItemsInputSpec)
}

func EmitDesignItemsAction(ctx context.Context, params ActionParams) (interface{}, error) {
	logger := params.Logger.With(zap.String("action", "emit_design_items"))

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}
	if params.DB == nil {
		return nil, fmt.Errorf("database connection required")
	}

	inputs, err := datahelpers.ExtractActionInputs(
		params.CollectedData, params.StepConfig.Config,
		EmitDesignItemsInputSpec, logger,
	)
	if err != nil {
		return nil, fmt.Errorf("input extraction failed: %w", err)
	}

	siteID, err := uuid.Parse(inputs.Get("site_id"))
	if err != nil {
		return nil, fmt.Errorf("invalid site_id %q: %w", inputs.Get("site_id"), err)
	}

	// Guard: skip if a composition is already installed. No backfill, no
	// duplicate emission on replan.
	var styleCollectionID sql.NullString
	if err = params.DB.QueryRowContext(ctx,
		`SELECT style_collection_id::text FROM sites WHERE id = $1`, siteID,
	).Scan(&styleCollectionID); err != nil {
		return nil, fmt.Errorf("load style_collection_id: %w", err)
	}
	if styleCollectionID.Valid {
		logger.Info("EmitDesignItemsAction: composition already installed — nothing to emit",
			zap.String("site_id", siteID.String()),
			zap.String("style_collection_id", styleCollectionID.String))
		return map[string]interface{}{"design_emitted": false, "reason": "composition_present"}, nil
	}

	batchID := uuid.New()
	tx, err := params.DB.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	// needs_composition → site-design-planner. Priority 7: ahead of pages
	// (50+) and the trailing rerender (99) so the design kit exists first.
	compSpec, _ := json.Marshal(map[string]interface{}{
		"reason": "build_pipeline_initial_composition",
	})
	if _, err = insertWorkItem(ctx, tx, workItem{
		siteID:       siteID,
		source:       "planner",
		pipeline:     "build",
		itemType:     "needs_composition",
		severity:     "high",
		summary:      "Resolve palette/layout/typography composition for the site",
		spec:         string(compSpec),
		priority:     7,
		handlerAgent: "site-design-planner",
		status:       "triaged",
		createdBy:    "build-site-planner",
		itemKey:      "needs_composition",
		batchID:      batchID,
	}, logger); err != nil {
		return nil, fmt.Errorf("insert needs_composition: %w", err)
	}

	// Resolve the open composition item id for needs_design.depends_on. The
	// partial unique index idx_swi_dedup guarantees at most one open row per
	// (site_id, item_key); if a prior open composition item exists, the insert
	// above no-opped and we depend on that one. If none is open, design runs
	// without a dependency (closest-to-correct: composition already ran/won't).
	var designDepends []uuid.UUID
	var compID uuid.UUID
	lookup := fmt.Sprintf(`
		SELECT id FROM site_work_items
		WHERE site_id = $1 AND item_key = 'needs_composition'
		  AND status NOT IN (%s)
		ORDER BY created_at DESC
		LIMIT 1
	`, sqlInList(workItemTerminalStatuses))
	switch err = tx.QueryRowContext(ctx, lookup, siteID).Scan(&compID); err {
	case nil:
		designDepends = []uuid.UUID{compID}
	case sql.ErrNoRows:
		logger.Info("EmitDesignItemsAction: no open composition item — design runs without dependency",
			zap.String("site_id", siteID.String()))
	default:
		logger.Warn("EmitDesignItemsAction: composition lookup failed — design runs without dependency",
			zap.Error(err))
	}

	// needs_design → webdesign-agent (renders + commits styles.css), gated on
	// composition completing so it never renders against a missing collection.
	if _, err = insertWorkItem(ctx, tx, workItem{
		siteID:       siteID,
		source:       "planner",
		pipeline:     "build",
		itemType:     "needs_design",
		severity:     "high",
		summary:      "Generate site stylesheet",
		spec:         "{}",
		priority:     8,
		handlerAgent: "webdesign-agent",
		status:       "triaged",
		createdBy:    "build-site-planner",
		itemKey:      "needs_design",
		batchID:      batchID,
		dependsOn:    designDepends,
	}, logger); err != nil {
		return nil, fmt.Errorf("insert needs_design: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}

	logger.Info("EmitDesignItemsAction: design trigger emitted",
		zap.String("site_id", siteID.String()),
		zap.String("composition_item_id", compID.String()))

	return map[string]interface{}{
		"design_emitted":      true,
		"composition_item_id": compID.String(),
		"batch_id":            batchID.String(),
	}, nil
}
