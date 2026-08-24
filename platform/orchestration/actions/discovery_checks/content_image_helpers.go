// FILE: platform/orchestration/actions/discovery_checks/content_image_helpers.go
//
// Pure helpers for check_content_image_missing, split out so the item-key and
// spec shapes — the contract with asset-deployer's content_card mode and with
// the dedup index — are unit-testable without a DB.

package discovery_checks

import (
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
)

// PageRerenderItemKey is the idx_swi_dedup key for a per-page page_rerender
// work item: page_rerender_<page>_<site>_<reason>, with "assemble" standing
// in for an empty reason. The reason suffix is load-bearing (bugs_open/024
// defect 6): page-rerender's check_rerender_mode branches on spec.reason, so a
// reason-less (assemble-only) item sitting open in the backlog must never be
// able to dedup-suppress the reason-bearing item that would actually
// re-render the sections.
//
// EXPORTED 2026-08-24 (bugs_open/384) for the same reason ContentImageItemKey
// was: the event-driven page-list invalidation (actions.requestPageListReresolve,
// called when a card or page image lands) and the page_list_stale sweep both
// file this item for the same page, and idx_swi_dedup collapses them onto ONE
// row only if the strings are byte-identical. actions.pageRerenderItemKey
// delegates here, so there is one spelling.
//
// idx_swi_dedup is (site_id, item_key) with NO item_type column, so this key
// space is shared with every other type on the site — and that is not
// theoretical: 20 (site_id, item_key) pairs have carried two item_types across
// site_work_items + site_work_items_archive as of 2026-08-24, one of them
// needs_page + page_rerender on `page_rerender:llm-cost-calculator`. Every one
// of the 20 is the COLON shape (`page_rerender:<page>` — the needs_page
// producers' key: 491 of 491 needs_page keys across live+archive, 0 with an
// underscore). This helper's underscore shape has never been minted by any
// other type, so the two cannot meet by construction; the residual hazard is
// hand dispatches that mint page_rerender rows with the colon shape (46 across
// live+archive), which is the namespace needs_page occupies. Baseline 20 is a
// ratchet, not a zero-invariant: a 21st pair is the signal (352 lane).
func PageRerenderItemKey(pageName string, siteID uuid.UUID, keyReason string) string {
	if keyReason == "" {
		keyReason = "assemble"
	}
	return fmt.Sprintf("page_rerender_%s_%s_%s", pageName, siteID, keyReason)
}

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
