package discovery_checks

import (
	"encoding/json"
	"fmt"

	"go.uber.org/zap"
)

func init() { Register(&UndeployedAssetsCheck{}) }

type UndeployedAssetsCheck struct{}

func (c *UndeployedAssetsCheck) Name() string { return "undeployed_assets" }

func (c *UndeployedAssetsCheck) Run(dctx DiscoveryCheckContext) (*CheckResult, error) {
	assets, err := findUndeployedAssets(dctx)
	if err != nil {
		return nil, err
	}
	if len(assets) == 0 {
		return &CheckResult{}, nil
	}

	result := &CheckResult{
		Findings: []map[string]interface{}{{
			"check":  "undeployed_assets",
			"count":  len(assets),
			"assets": assets,
		}},
	}

	for _, asset := range assets {
		specJSON, _ := json.Marshal(map[string]interface{}{
			"check":      "undeployed_assets",
			"asset_id":   asset.AssetID,
			"purpose":    asset.Purpose,
			"asset_type": asset.AssetType,
			"url":        asset.URL,
		})

		result.WorkItems = append(result.WorkItems, WorkItemSpec{
			SiteID:       dctx.SiteID,
			Source:       "discovery",
			Pipeline:     "design",
			ItemType:     "undeployed_asset",
			Severity:     "high",
			Summary:      fmt.Sprintf("Asset '%s' generated but not deployed to site", asset.Purpose),
			SpecJSON:     string(specJSON),
			Priority:     60,
			HandlerAgent: "asset-deployer",
			Status:       "detected",
			CreatedBy:    dctx.AgentType,
			ItemKey:      fmt.Sprintf("undeployed_asset:%s", asset.AssetID),
			BatchID:      dctx.BatchID,
		})
	}

	return result, nil
}

type undeployedAssetFinding struct {
	AssetID   string `json:"asset_id"`
	Purpose   string `json:"purpose"`
	AssetType string `json:"asset_type"`
	URL       string `json:"url"`
}

func findUndeployedAssets(dctx DiscoveryCheckContext) ([]undeployedAssetFinding, error) {
	rows, err := dctx.DB.QueryContext(dctx.Ctx, `
		SELECT a.id, COALESCE(a.purpose, 'unknown'), a.asset_type, a.url
		FROM assets a
		WHERE a.site_id = $1
		  AND NOT EXISTS (
		      SELECT 1 FROM page_components pc
		      JOIN pages p ON pc.page_id = p.id
		      WHERE p.site_id = a.site_id
		        AND pc.build_status = 'deployed'
		        AND (
		            pc.rendered_html LIKE '%/assets/images/' || COALESCE(a.purpose, '') || '.%'
		            OR pc.rendered_html LIKE '%/assets/images/' || COALESCE(a.purpose, '') || '-%'
		        )
		  )
		ORDER BY a.purpose
	`, dctx.SiteID)
	if err != nil {
		return nil, fmt.Errorf("undeployed_assets query failed: %w", err)
	}
	defer rows.Close()

	var findings []undeployedAssetFinding
	for rows.Next() {
		var f undeployedAssetFinding
		if err := rows.Scan(&f.AssetID, &f.Purpose, &f.AssetType, &f.URL); err != nil {
			dctx.Logger.Warn("Failed to scan undeployed asset", zap.Error(err))
			continue
		}
		findings = append(findings, f)
	}
	return findings, nil
}
