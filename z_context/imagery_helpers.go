// FILE: platform/orchestration/actions/discovery_checks/imagery_helpers.go
//
// Shared helpers for the imagery discovery checks
// (unfulfilled_image_prompt, placeholder_image_in_use, image_url_404).
//
// Kept in a single file so the imagery checks read top-to-bottom without
// having to track shared functions defined in unrelated check files.

package discovery_checks

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"go.uber.org/zap"
)

// loadImagePromptsForSite reads site_specs.data.image_prompts for the current
// site_plan aspect. Returns the map of prompt_key -> prompt_text. Empty map
// (not error) if the spec is missing, the column is NULL, or the JSON is
// malformed — discovery checks should degrade gracefully rather than fail
// the whole discovery pass.
func loadImagePromptsForSite(dctx DiscoveryCheckContext) (map[string]string, error) {
	var raw []byte
	err := dctx.DB.QueryRowContext(dctx.Ctx, `
		SELECT data->'image_prompts'
		FROM site_specs
		WHERE site_id = $1
		  AND aspect = 'site_plan'
		  AND is_current = true
		LIMIT 1
	`, dctx.SiteID).Scan(&raw)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("site_plan spec query failed: %w", err)
	}
	if len(raw) == 0 {
		return nil, nil
	}

	// Primary path: prompts are a flat string map.
	var prompts map[string]string
	if err := json.Unmarshal(raw, &prompts); err == nil && prompts != nil {
		return prompts, nil
	}

	// Permissive fallback: nested objects per future planners. Stringify
	// scalar values only; drop anything else.
	var generic map[string]interface{}
	if err := json.Unmarshal(raw, &generic); err != nil {
		dctx.Logger.Warn("loadImagePromptsForSite: image_prompts not parseable",
			zap.Error(err))
		return nil, nil
	}
	out := make(map[string]string, len(generic))
	for k, v := range generic {
		if s, ok := v.(string); ok {
			out[k] = s
		}
	}
	return out, nil
}

// hasActiveAssetForPurpose returns true if any active asset exists for the
// given site_id + purpose. Only counts status='active' rows; tombstoned
// rows (status != 'active') are treated as absent so the discovery check
// will recommend regeneration.
//
// Note: post Phase 2D, multiple active rows can share a purpose (one per
// asset_key). Use hasActiveAssetForAssetKey when checking whether a
// SPECIFIC image variant is present. Use this function only when asking
// "is any image of this purpose present at all?" — typically not what
// the discovery checks want.
func hasActiveAssetForPurpose(dctx DiscoveryCheckContext, purpose string) (bool, error) {
	var n int
	err := dctx.DB.QueryRowContext(dctx.Ctx, `
		SELECT COUNT(*)
		FROM assets
		WHERE site_id = $1
		  AND purpose = $2
		  AND status = 'active'
	`, dctx.SiteID, purpose).Scan(&n)
	if err != nil {
		return false, err
	}
	return n > 0, nil
}

// hasActiveAssetForAssetKey returns true if an active asset exists for the
// given site_id + asset_key. This is the right granularity for the
// imagery discovery checks post Phase 2D (asset_key is the unique
// identifier per active row; purpose can repeat across variants).
func hasActiveAssetForAssetKey(dctx DiscoveryCheckContext, assetKey string) (bool, error) {
	var n int
	err := dctx.DB.QueryRowContext(dctx.Ctx, `
		SELECT COUNT(*)
		FROM assets
		WHERE site_id = $1
		  AND asset_key = $2
		  AND status = 'active'
	`, dctx.SiteID, assetKey).Scan(&n)
	if err != nil {
		return false, err
	}
	return n > 0, nil
}
