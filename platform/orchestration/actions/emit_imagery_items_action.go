// FILE: platform/orchestration/actions/emit_imagery_items_action.go
//
// EmitImageryItemsAction queues needs_imagery work items at BUILD TIME from the
// current plan's site_plan_imagery rows, so a fresh site deploys with its logo,
// hero and section images rather than waiting for the improvement loop.
//
// WHY THIS EXISTS
//   write_site_plan records the planner's image requests in site_plan_imagery
//   (flattenImageryBlock), but nothing on the build-site-planner path acted on
//   them — needs_imagery was emitted only by the unfulfilled_imagery_plan
//   discovery check, which runs in the improvement loop, not during the build.
//   Same shape as the dropped design trigger. This is the build-time emitter,
//   invoked as a build-site-planner workflow step.
//
//   The selection, classification, brand_update rule, item_key and spec body
//   are shared with the discovery check via the imageryplan package, so the two
//   emitters cannot drift. The differences are intentional and local:
//     - status 'triaged' (build path auto-dispatch) vs 'detected' (loop triages)
//     - section-scope priority clamped below the terminal rerender (see below)
//     - source 'planner', created_by 'build-site-planner'
//   item_key dedup (idx_swi_dedup) prevents the build-time and loop emitters
//   from double-queuing the same imagery row.

package actions

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"github.com/gqls/agentchassis/platform/orchestration/imageryplan"
	"go.uber.org/zap"
)

var EmitImageryItemsInputSpec = datahelpers.ActionInputSpec{
	CheckConfig: true,
	Required:    []string{"site_id"},
	Optional:    []string{},
	Defaults:    map[string]interface{}{},
}

func init() {
	datahelpers.RegisterActionInputSpec("emit_imagery_items", EmitImageryItemsInputSpec)
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

	rows, err := imageryplan.LoadCurrentPlan(ctx, params.DB, siteID)
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
		if emitted >= imageryplan.MaxPerPass {
			logger.Info("EmitImageryItemsAction: per-build cap reached; overflow handled by the loop",
				zap.Int("cap", imageryplan.MaxPerPass), zap.Int("plan_rows", len(rows)))
			break
		}

		assetKey := imageryplan.AssetKey(row)

		has, err := imageryplan.HasActiveAsset(ctx, tx, siteID, assetKey)
		if err != nil {
			logger.Warn("EmitImageryItemsAction: asset existence check failed",
				zap.String("asset_key", assetKey), zap.Error(err))
			continue
		}
		if has {
			continue
		}

		priority, severity := imageryplan.Classify(row.Scope, row.Kind, row.ScopeRef)
		// Build-time imagery must precede the terminal needs_rerender (priority
		// 99) so the first deploy includes it. Section-scope classifies at 100
		// (after rerender) for the loop's catch-up purpose; clamp it below 99
		// here so it lands in the first rerender instead.
		if priority >= 99 {
			priority = 98
		}

		specJSON, err := imageryplan.BuildSpec(row, "emit_imagery_items")
		if err != nil {
			logger.Warn("EmitImageryItemsAction: spec build failed",
				zap.String("asset_key", assetKey), zap.Error(err))
			continue
		}

		itemKey := imageryplan.ItemKey(row)
		if _, err = insertWorkItem(ctx, tx, workItem{
			siteID:       siteID,
			source:       "planner",
			pipeline:     "build",
			itemType:     "needs_imagery",
			severity:     severity,
			summary:      imageryplan.Summary(row),
			spec:         specJSON,
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
