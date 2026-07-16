// FILE: platform/orchestration/actions/discovery_checks/content_image_helpers.go
//
// Pure helpers for check_content_image_missing, split out so the item-key and
// spec shapes — the contract with asset-deployer's content_card mode and with
// the dedup index — are unit-testable without a DB.

package discovery_checks

import "encoding/json"

// contentImageItemKey is the dedup key for one page's card derivation. Keyed
// on page name (stable across re-plans that keep the page) rather than page
// id, mirroring how imagery item keys ride asset_key conventions.
func contentImageItemKey(pageName string) string {
	return "content_image:" + pageName
}

// contentImageSpecJSON builds the needs_content_image spec consumed by
// asset-deployer's content_card mode (derive_card_asset reads entity_type,
// entity_id and page_name; check_mode routing reads mode).
func contentImageSpecJSON(checkName, pageID, pageName string) (string, error) {
	b, err := json.Marshal(map[string]interface{}{
		"mode":        "content_card",
		"check":       checkName,
		"entity_type": "page",
		"entity_id":   pageID,
		"page_name":   pageName,
		"purpose":     "card",
	})
	if err != nil {
		return "", err
	}
	return string(b), nil
}
