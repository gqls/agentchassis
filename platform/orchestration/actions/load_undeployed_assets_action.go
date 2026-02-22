// FILE: platform/orchestration/actions/load_undeployed_assets_action.go
//
// Thin action wrapping findUndeployedAssets (from discovery_checks.go).
// Used by asset-deploy-agent to query which assets need deploying,
// then loop over them calling deploy_image_asset for each.
//
// Returns assets in loop-friendly format with s3_uri and purpose fields
// matching what deploy_image_asset expects via its InputSpec.

package actions

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"go.uber.org/zap"
)

var LoadUndeployedAssetsInputSpec = datahelpers.ActionInputSpec{
	Required:   []string{"site_id"},
	Optional:   []string{},
	Defaults:   map[string]interface{}{},
	Deprecated: map[string]string{},
}

func init() {
	datahelpers.RegisterActionInputSpec("load_undeployed_assets", LoadUndeployedAssetsInputSpec)
}

// LoadUndeployedAssetsAction queries assets in the assets table that aren't
// referenced in any deployed page HTML for the site.
//
// Data inputs (via ActionInputSpec):
//   - site_id (required) — resolved from collectedData
//
// Output:
//
//	{
//	  "assets": [
//	    {"asset_id": "...", "purpose": "logo", "asset_type": "logo", "s3_uri": "s3://..."},
//	    {"asset_id": "...", "purpose": "hero", "asset_type": "image", "s3_uri": "s3://..."}
//	  ],
//	  "count": 2,
//	  "has_assets": true
//	}
//
// Note: each asset item uses "s3_uri" (not "url") to match deploy_image_asset's InputSpec.
func LoadUndeployedAssetsAction(ctx context.Context, params ActionParams) (interface{}, error) {
	logger := params.Logger.With(
		zap.String("action", "load_undeployed_assets"),
		zap.String("step_name", params.ExecutionContext.StepName),
	)

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}
	if params.DB == nil {
		return nil, fmt.Errorf("database connection required")
	}

	inputs, err := datahelpers.ExtractActionInputs(
		params.CollectedData,
		params.StepConfig.Config,
		LoadUndeployedAssetsInputSpec,
		logger,
	)
	if err != nil {
		return nil, fmt.Errorf("input extraction failed: %w", err)
	}

	siteIDStr := inputs.Get("site_id")
	siteID, err := uuid.Parse(siteIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid site_id: %w", err)
	}

	// findUndeployedAssets is defined in discovery_checks.go (same package)
	findings, err := findUndeployedAssets(ctx, params.DB, siteID, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to query undeployed assets: %w", err)
	}

	// Map to loop-friendly format with field names matching deploy_image_asset InputSpec
	assets := make([]interface{}, 0, len(findings))
	for _, f := range findings {
		assets = append(assets, map[string]interface{}{
			"asset_id":   f.AssetID,
			"purpose":    f.Purpose,
			"asset_type": f.AssetType,
			"s3_uri":     f.URL, // "url" from assets table is the S3 URI; rename to match deploy_image_asset
		})
	}

	logger.Info("Loaded undeployed assets",
		zap.String("site_id", siteIDStr),
		zap.Int("count", len(assets)),
	)

	return map[string]interface{}{
		"assets":     assets,
		"count":      len(assets),
		"has_assets": len(assets) > 0,
	}, nil
}
