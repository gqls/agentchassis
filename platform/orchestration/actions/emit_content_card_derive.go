// FILE: platform/orchestration/actions/emit_content_card_derive.go
//
// bugs_open/114 — closing the imagery loop at the EVENT rather than at a sweep.
//
// THE GAP THIS FILLS. check_content_image_missing's design is a three-pass
// convergence, stated in its own header: pass 1 GENERATEs a missing content
// hero; a LATER pass sees the page has no card and files the DERIVE that builds
// one; pass 3 is silent. derive_card_asset is the estate's ONLY writer of
// assets.entity_type/entity_id, and both readers of that link —
// queryresolve.pageImageJoins (every listing card on the fleet) and the check's
// own sweep — additionally require purpose='card'. So until the DERIVE runs, a
// generated hero is invisible to every listing.
//
// The convergence is correct and it does not happen, because pass 2 is a sweep.
// `[MEASURED 2026-08-22]` site_discovery_rotation's newest last_selected_at for
// design-discovery-agent — the lane that runs this check — is 2026-08-11, while
// availability, completeness, quality and render-audit are all current
// (08-21/22). Eleven days. The bugs_open/230 staleness CronJob reports the gap
// DAILY and nothing consumes the report. On mortgagecalculator, ten content
// heroes generated on 08-15 have zero cards and zero entity links today.
//
// So the second pass runs HERE, at the moment the asset lands, in the same
// transaction as the page re-render that image-build-handler already emits. A
// generation now converges on its own; the sweep becomes a backstop for the
// stale-card case rather than the only route.
//
// WHAT THIS DELIBERATELY DOES NOT DO
//   - It does not write the entity link itself. derive_card_asset owns that
//     write, and a second writer of a column whose absence is invisible is how
//     this class of bug is made. This files the request; the existing handler
//     does the work, unchanged.
//   - It does not refresh a STALE card (one whose origin is a superseded
//     source). That needs the origin-lineage comparison the sweep already does
//     (contentImageAction's CardOriginID != source arm) and it is not urgent:
//     a stale card still renders an image. The gap worth closing at the event
//     is the ABSENT card, which renders nothing.
//   - It does not fire for pages with no content hero. The by-key existence
//     check below is the same convention plan_sections' Lane B resolver uses
//     (imageryplan.ContentHeroKey), so a page whose hero the resolver cannot
//     find does not get a card request either — one convention, two consumers.
//
// STATUS: 'triaged', NOT 'detected'. The sweep files these at 'detected'
// because discovery findings are promoted by a triage step. An item filed at
// 'detected' by an event would inherit exactly the dependency this exists to
// remove: claim_work_item takes ('triaged','approved') only, so a 'detected'
// item waits for a promoter that, like the sweep, may not run. This is the same
// status the sibling needs_page emit in this action already uses.
//
// DEDUP: the item key is ContentImageItemKey(pageName) — the discovery check's
// own exported helper, not a second spelling — so an event-filed item and a
// sweep-filed item collapse onto one row via idx_swi_dedup rather than
// producing two asset-deployer runs against the same page.

package actions

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/gqls/agentchassis/platform/orchestration/actions/discovery_checks"
	"github.com/gqls/agentchassis/platform/orchestration/imageryplan"
	"go.uber.org/zap"
)

// emitContentCardDerive files the card-derivation request for a page whose
// content hero has just landed, when that page has no card yet.
//
// It never returns an error: this is an opportunistic second half bolted to a
// re-render emit that has already succeeded, and failing the whole action —
// which would leave the work item failed and the page un-re-rendered — is a
// worse outcome than not deriving a card. Every disposition is returned as a
// string for the caller's log line, so a skip is visible rather than silent.
func emitContentCardDerive(
	ctx context.Context,
	tx *sql.Tx,
	siteID uuid.UUID,
	pageName string,
	batchID uuid.UUID,
	logger *zap.Logger,
) string {
	var pageID string
	var hasHero, hasCard bool

	// One round trip: the page's id, whether its content hero exists under the
	// resolver's own key convention, and whether an entity-linked card already
	// exists. The card predicate mirrors the readers exactly (entity_type,
	// entity_id, purpose='card', status='active') — asking a different question
	// here than the consumers ask is how a check comes to disagree with the
	// thing it is checking.
	err := tx.QueryRowContext(ctx, `
		SELECT p.id::text,
		       EXISTS (SELECT 1 FROM assets a
		                WHERE a.site_id = p.site_id
		                  AND a.asset_key = $3
		                  AND a.status = 'active'),
		       EXISTS (SELECT 1 FROM assets c
		                WHERE c.site_id = p.site_id
		                  AND c.entity_type = 'page'
		                  AND c.entity_id = p.id
		                  AND c.purpose = 'card'
		                  AND c.status = 'active')
		  FROM pages p
		 WHERE p.site_id = $1 AND p.name = $2
		 LIMIT 1
	`, siteID, pageName, imageryplan.ContentHeroKey(pageName)).Scan(&pageID, &hasHero, &hasCard)

	switch {
	case err == sql.ErrNoRows:
		logger.Info("emit_content_card_derive: no page row for that name, nothing to derive",
			zap.String("page", pageName))
		return "skipped_no_page"
	case err != nil:
		logger.Warn("emit_content_card_derive: lookup failed, card derivation not requested",
			zap.String("page", pageName), zap.Error(err))
		return "skipped_lookup_failed"
	case !hasHero:
		// Expected for a logo, favicon, sprite sheet or any brand asset landing
		// on a page-scoped item — those have no card to derive from.
		return "skipped_no_content_hero"
	case hasCard:
		// A card exists and is entity-linked, so the listing joins resolve. If it
		// is STALE the sweep's origin-lineage arm re-derives it; see the header.
		return "skipped_card_exists"
	}

	item, err := contentCardDeriveItem(siteID, pageID, pageName, batchID)
	if err != nil {
		logger.Warn("emit_content_card_derive: spec build failed",
			zap.String("page", pageName), zap.Error(err))
		return "skipped_spec_failed"
	}

	if _, err := insertWorkItem(ctx, tx, item, logger); err != nil {
		logger.Warn("emit_content_card_derive: insert failed, card derivation not requested",
			zap.String("page", pageName), zap.Error(err))
		return "skipped_insert_failed"
	}

	logger.Info("emit_content_card_derive: queued card derivation at the landing event rather than waiting for a discovery sweep (bugs_open/114)",
		zap.String("site_id", siteID.String()),
		zap.String("page", pageName),
		zap.String("page_id", pageID))
	return "raised"
}

// contentCardDeriveItem builds the work item this emitter files.
//
// Extracted as a pure function so the ITEM ITSELF is testable, not merely the
// helpers it calls. That distinction cost a mutation: an earlier cut inlined
// this and the accompanying test asserted
// discovery_checks.ContentImageItemKey(page) == "content_image:"+page — which
// is a test of the helper. Replacing the emitter's call with a hand-spelled
// "content-image:"+page (a hyphen where the contract has an underscore, the
// exact drift the shared helper exists to prevent) PASSED that test. A test
// that never touches the code under test passes every mutation of it; this is
// the second time in one session, the first being store_asset's URL derivation.
func contentCardDeriveItem(siteID uuid.UUID, pageID, pageName string, batchID uuid.UUID) (workItem, error) {
	specJSON, err := discovery_checks.ContentImageSpecJSON("flag_page_image_rebuild", pageID, pageName)
	if err != nil {
		return workItem{}, err
	}
	return workItem{
		siteID:       siteID,
		source:       "image-build-handler",
		pipeline:     "build",
		itemType:     "needs_content_image",
		severity:     "low",
		summary:      "Article \"" + pageName + "\" has no card image — derive from its hero",
		spec:         specJSON,
		priority:     65,
		handlerAgent: "asset-deployer",
		status:       "triaged",
		createdBy:    "image-build-handler",
		itemKey:      discovery_checks.ContentImageItemKey(pageName),
		batchID:      batchID,
	}, nil
}
