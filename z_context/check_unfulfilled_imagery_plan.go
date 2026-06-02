// FILE: platform/orchestration/actions/discovery_checks/check_unfulfilled_imagery_plan.go
//
// Discovery check: site_plan_imagery rows whose assets are missing.
//
// Phase 2G step 4. Reads the current plan's imagery rows (populated by
// build-site-planner + write_site_plan_action's flattenImageryBlock) and emits
// a needs_imagery work item for each row whose asset_key is not yet present in
// active assets for the site.
//
// The selection, priority bands, brand_update rule, item_key shape and spec
// body now live in the shared imageryplan package, so this loop-side emitter
// and the build-time actions.EmitImageryItemsAction cannot drift. This check
// keeps its own concerns: status 'detected' (the loop triages), the per-row
// asset-key collision logging, and the canonical section-scope priority (100)
// for catch-up ordering (the build-time emitter clamps that below the terminal
// rerender; the loop does not).
//
// Sibling of check_unfulfilled_image_prompt.go (Phase 1.1). Both coexist during
// the transition window; the older check de-registers at Phase 2G step 6.

package discovery_checks

import (
	"github.com/gqls/agentchassis/platform/orchestration/imageryplan"
	"go.uber.org/zap"
)

func init() { Register(&UnfulfilledImageryPlanCheck{}) }

type UnfulfilledImageryPlanCheck struct{}

func (c *UnfulfilledImageryPlanCheck) Name() string { return "unfulfilled_imagery_plan" }

func (c *UnfulfilledImageryPlanCheck) Run(dctx DiscoveryCheckContext) (*CheckResult, error) {
	imagery, err := imageryplan.LoadCurrentPlan(dctx.Ctx, dctx.DB, dctx.SiteID)
	if err != nil {
		return nil, err
	}
	if len(imagery) == 0 {
		return &CheckResult{}, nil
	}

	result := &CheckResult{}
	emitted := 0
	seenAssetKeys := make(map[string]bool, len(imagery))

	for _, row := range imagery {
		if emitted >= imageryplan.MaxPerPass {
			dctx.Logger.Info("unfulfilled_imagery_plan: per-pass cap reached; remaining rows picked up next pass",
				zap.Int("cap", imageryplan.MaxPerPass),
				zap.Int("plan_rows_total", len(imagery)),
				zap.Int("emitted_this_pass", emitted))
			break
		}

		assetKey := imageryplan.AssetKey(row)

		// Cross-scope collision detection — visibility only, non-blocking.
		if seenAssetKeys[assetKey] {
			dctx.Logger.Warn("unfulfilled_imagery_plan: asset_key collision across imagery rows",
				zap.String("asset_key", assetKey),
				zap.String("scope", row.Scope),
				zap.String("key", row.Key))
		}
		seenAssetKeys[assetKey] = true

		hasAsset, err := hasActiveAssetForAssetKey(dctx, assetKey)
		if err != nil {
			dctx.Logger.Warn("unfulfilled_imagery_plan: asset existence check failed",
				zap.String("asset_key", assetKey),
				zap.Error(err))
			continue
		}
		if hasAsset {
			continue
		}

		priority, severity := imageryplan.Classify(row.Scope, row.Kind, row.ScopeRef)

		specJSON, err := imageryplan.BuildSpec(row, "unfulfilled_imagery_plan")
		if err != nil {
			dctx.Logger.Warn("unfulfilled_imagery_plan: spec build failed",
				zap.String("asset_key", assetKey),
				zap.Error(err))
			continue
		}

		result.Findings = append(result.Findings, map[string]interface{}{
			"check":     "unfulfilled_imagery_plan",
			"scope":     row.Scope,
			"key":       row.Key,
			"kind":      row.Kind,
			"asset_key": assetKey,
		})

		result.WorkItems = append(result.WorkItems, WorkItemSpec{
			SiteID: dctx.SiteID,
			Source: "discovery",
			// Pipeline is the destination, not the origin. needs_imagery routes
			// to image-build-handler (build pipeline) regardless of which
			// discovery agent invoked us.
			Pipeline:     "build",
			ItemType:     "needs_imagery",
			Severity:     severity,
			Summary:      imageryplan.Summary(row),
			SpecJSON:     specJSON,
			Priority:     priority,
			HandlerAgent: "image-build-handler",
			Status:       "detected",
			CreatedBy:    dctx.AgentType,
			ItemKey:      imageryplan.ItemKey(row),
			BatchID:      dctx.BatchID,
		})
		emitted++
	}

	if emitted > 0 {
		dctx.Logger.Info("unfulfilled_imagery_plan: emitted work items",
			zap.Int("count", emitted),
			zap.Int("plan_rows", len(imagery)))
	}
	return result, nil
}
