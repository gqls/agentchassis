// FILE: platform/orchestration/actions/discovery_checks/content_image_helpers.go
//
// Pure helpers for check_content_image_missing, split out so the item-key and
// spec shapes — the contract with asset-deployer's content_card mode and with
// the dedup index — are unit-testable without a DB.

package discovery_checks

import "encoding/json"

// ContentImageItemKey is the dedup key for one page's card derivation. Keyed
// on page name (stable across re-plans that keep the page) rather than page
// id, mirroring how imagery item keys ride asset_key conventions.
//
// EXPORTED 2026-08-22 (bugs_open/114) so the event-driven emitter in the
// actions package files a byte-identical key. The sweep and the event must
// collapse onto ONE item via idx_swi_dedup; two spellings of this key would
// mean two items, two asset-deployer runs and two upserts of the same card.
func ContentImageItemKey(pageName string) string {
	return "content_image:" + pageName
}

// ContentImageSpecJSON builds the needs_content_image spec consumed by
// asset-deployer's content_card mode (derive_card_asset reads entity_type,
// entity_id and page_name; check_mode routing reads mode).
//
// EXPORTED 2026-08-22 (bugs_open/114) for the same reason as the key above:
// the discovery sweep and the image-build-handler's terminal step both file
// this item, and derive_card_asset reads these exact field names. Two
// hand-maintained spellings of one consumer's contract is the drift class this
// estate keeps paying for — so there is one, here.
func ContentImageSpecJSON(checkName, pageID, pageName string) (string, error) {
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
