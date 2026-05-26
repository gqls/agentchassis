// FILE: platform/orchestration/actions/emit_imagery_items_action.go
//
// EmitImageryItemsAction queues needs_imagery work items at BUILD TIME from the
// current plan's site_plan_imagery rows, so a fresh site deploys with its logo,
// hero and section images rather than waiting for the improvement loop.
//
// WHY THIS EXISTS
//   write_site_plan records the planner's image requests in site_plan_imagery,
//   but nothing on the build-site-planner path acts on them — needs_imagery was
//   emitted only by the unfulfilled_imagery_plan discovery check, which runs in
//   the improvement loop (design-discovery), not during the build. This is the
//   same shape as the dropped design trigger. This action is the build-time
//   emitter, invoked as a build-site-planner workflow step.
//
// RELATIONSHIP TO THE DISCOVERY CHECK
//   This mirrors discovery_checks/check_unfulfilled_imagery_plan.go: same
//   selection (current plan's imagery rows, skip rows whose asset_key already
//   has an active asset), same priority bands, same per-pass cap, same
//   needs_imagery -> image-build-handler routing and item_key shape. The
//   differences are: emitted at status 'triaged' (build path auto-dispatch, vs
//   'detected' which the loop triages) and createdBy 'build-site-planner'.
//   The two coexist: the discovery check still covers stale sites in the loop;
//   item_key dedup (idx_swi_dedup) prevents the build-time and loop emitters
//   from double-queuing the same imagery row.
//
//   NOTE ON DUPLICATION: the selection + classification logic is duplicated
//   from the discovery check because the two live in different packages
//   (actions vs discovery_checks) and the helpers are unexported. If this pair
//   is kept, extract loadCurrentPlanImagery + classifyImageryRow into a shared
//   package that both import.
//
// GUARD
//   Emits only for imagery rows that have no active asset for their asset_key.
//   image-build-handler stores the asset, so a replan after images exist is a
//   no-op. Capped at maxImageryWorkItemsPerBuild per run; the loop's discovery
//   check picks up any overflow on subsequent passes.
//
// REGISTRATION (registry.go):
//   "emit_imagery_items": {
//       Handler:     EmitImageryItemsAction,
//       Category:    "site",
//       Description: "Queue needs_imagery from the current plan's site_plan_imagery rows",
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

// maxImageryWorkItemsPerBuild mirrors the discovery check's per-pass cap so a
// rich plan doesn't flood image-build-handler in one go. Overflow is picked up
// by the loop's unfulfilled_imagery_plan check.
const maxImageryWorkItemsPerBuild = 20

var EmitImageryItemsInputSpec = datahelpers.ActionInputSpec{
	Required: []string{"site_id"},
	Optional: []string{},
	Defaults: map[string]interface{}{},
}

func init() {
	datahelpers.RegisterActionInputSpec("emit_imagery_items", EmitImageryItemsInputSpec)
}

type plannedImageryRow struct {
	Scope       string
	ScopeRef    *string
	Key         string
	Kind        string
	Prompt      string
	StyleHints  json.RawMessage
	Constraints json.RawMessage
}

// classifyImageryRowPriority mirrors classifyImageryRow in the discovery check.
func classifyImageryRowPriority(scope, kind string, scopeRef *string) (priority int, severity string) {
	if scope == "page" && scopeRef != nil && *scopeRef == "index" && kind == "hero" {
		return 65, "high"
	}
	if scope == "site" && kind == "logo" {
		return 70, "high"
	}
	if scope == "site" {
		return 75, "high"
	}
	if scope == "page" && kind == "hero" {
		return 80, "medium"
	}
	if scope == "page" {
		return 90, "medium"
	}
	return 100, "low"
}

func EmitImageryItemsAction(ctx context.Context, params ActionParams) (interface{}, error) {
	logger := params.Logger.With(zap.String("action", "emit_imagery_items"))

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}
	if params.DB == nil {
		return nil, fmt.Errorf("database connection required")
	}

	inputs, err := datahelpers.ExtractActionInputs(
		params.CollectedData, params.StepConfig.Config,
		EmitImageryItemsInputSpec, logger,
	)
	if err != nil {
		return nil, fmt.Errorf("input extraction failed: %w", err)
	}
	siteID, err := uuid.Parse(inputs.Get("site_id"))
	if err != nil {
		return nil, fmt.Errorf("invalid site_id %q: %w", inputs.Get("site_id"), err)
	}

	rows, err := loadCurrentPlanImageryRows(ctx, params.DB, siteID)
	if err != nil {
		return nil, fmt.Errorf("load site_plan_imagery: %w", err)
	}
	if len(rows) == 0 {
		return map[string]interface{}{"imagery_emitted": 0}, nil
	}

	batchID := uuid.New()
	tx, err := params.DB.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	emitted := 0
	for _, row := range rows {
		if emitted >= maxImageryWorkItemsPerBuild {
			logger.Info("EmitImageryItemsAction: per-build cap reached; overflow handled by the loop",
				zap.Int("cap", maxImageryWorkItemsPerBuild), zap.Int("plan_rows", len(rows)))
			break
		}

		assetKey := row.Key

		has, err := assetKeyHasActiveAsset(ctx, tx, siteID, assetKey)
		if err != nil {
			logger.Warn("EmitImageryItemsAction: asset existence check failed",
				zap.String("asset_key", assetKey), zap.Error(err))
			continue
		}
		if has {
			continue
		}

		priority, severity := classifyImageryRowPriority(row.Scope, row.Kind, row.ScopeRef)
		brandUpdate := row.Scope == "site" ||
			(row.Scope == "page" && row.ScopeRef != nil && *row.ScopeRef == "index" && row.Kind == "hero")

		spec := map[string]interface{}{
			"check":        "emit_imagery_items",
			"scope":        row.Scope,
			"key":          row.Key,
			"kind":         row.Kind,
			"asset_key":    assetKey,
			"purpose":      row.Kind,
			"prompt":       row.Prompt,
			"brand_update": brandUpdate,
		}
		if row.ScopeRef != nil {
			spec["scope_ref"] = *row.ScopeRef
		}
		if len(row.StyleHints) > 0 {
			spec["style_hints"] = row.StyleHints
		}
		if len(row.Constraints) > 0 {
			spec["constraints"] = row.Constraints
		}
		specJSON, err := json.Marshal(spec)
		if err != nil {
			logger.Warn("EmitImageryItemsAction: spec marshal failed",
				zap.String("asset_key", assetKey), zap.Error(err))
			continue
		}

		scopeRefDisplay := "-"
		if row.ScopeRef != nil {
			scopeRefDisplay = *row.ScopeRef
		}
		itemKey := fmt.Sprintf("needs_imagery:%s:%s:%s", row.Scope, scopeRefDisplay, row.Key)

		if _, err = insertWorkItem(ctx, tx, workItem{
			siteID:       siteID,
			source:       "planner",
			pipeline:     "build",
			itemType:     "needs_imagery",
			severity:     severity,
			summary:      fmt.Sprintf("Imagery %s/%s (kind=%s) requested but no asset for %s", row.Scope, row.Key, row.Kind, assetKey),
			spec:         string(specJSON),
			priority:     priority,
			handlerAgent: "image-build-handler",
			status:       "triaged",
			createdBy:    "build-site-planner",
			itemKey:      itemKey,
			batchID:      batchID,
		}, logger); err != nil {
			return nil, fmt.Errorf("insert needs_imagery (%s): %w", itemKey, err)
		}
		emitted++
	}

	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}

	logger.Info("EmitImageryItemsAction: emitted needs_imagery",
		zap.String("site_id", siteID.String()), zap.Int("emitted", emitted), zap.Int("plan_rows", len(rows)))

	return map[string]interface{}{"imagery_emitted": emitted, "batch_id": batchID.String()}, nil
}

// loadCurrentPlanImageryRows mirrors loadCurrentPlanImagery in the discovery check.
func loadCurrentPlanImageryRows(ctx context.Context, db *sql.DB, siteID uuid.UUID) ([]plannedImageryRow, error) {
	rows, err := db.QueryContext(ctx, `
		SELECT spi.scope, spi.scope_ref, spi.key, spi.kind, spi.prompt,
		       COALESCE(spi.style_hints::text, ''),
		       COALESCE(spi.constraints::text, '')
		  FROM site_plan_imagery spi
		  JOIN site_plans sp ON sp.id = spi.plan_id
		 WHERE sp.site_id = $1
		   AND sp.is_current = true
		 ORDER BY
		     CASE spi.scope
		         WHEN 'site'    THEN 0
		         WHEN 'page'    THEN 1
		         WHEN 'section' THEN 2
		         ELSE 3
		     END,
		     spi.scope_ref NULLS FIRST,
		     spi.ordering
	`, siteID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []plannedImageryRow
	for rows.Next() {
		var r plannedImageryRow
		var scopeRef sql.NullString
		var styleHints, constraints string
		if err := rows.Scan(&r.Scope, &scopeRef, &r.Key, &r.Kind, &r.Prompt, &styleHints, &constraints); err != nil {
			return nil, err
		}
		if scopeRef.Valid {
			r.ScopeRef = &scopeRef.String
		}
		if styleHints != "" {
			r.StyleHints = json.RawMessage(styleHints)
		}
		if constraints != "" {
			r.Constraints = json.RawMessage(constraints)
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

// assetKeyHasActiveAsset mirrors hasActiveAssetForAssetKey in the discovery check.
func assetKeyHasActiveAsset(ctx context.Context, tx *sql.Tx, siteID uuid.UUID, assetKey string) (bool, error) {
	var n int
	err := tx.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM assets
		WHERE site_id = $1 AND asset_key = $2 AND status = 'active'
	`, siteID, assetKey).Scan(&n)
	if err != nil {
		return false, err
	}
	return n > 0, nil
}
