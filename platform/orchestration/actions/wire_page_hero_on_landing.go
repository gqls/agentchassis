// FILE: platform/orchestration/actions/wire_page_hero_on_landing.go
//
// bugs_open/412 fix candidate 1 — the hero WIRING, hoisted out of the build
// path to the landing event. Built by the bugfix_114_imagery_wiring lane under
// 412's number by that lane's recorded handover (412 §10, 2026-09-02).
//
// THE GAP THIS FILLS. A page-scoped content hero lands (generated, stored,
// deployed, active) and the only thing that can put its URL in front of a
// visitor is a full page build re-resolving `site_assets.hero` through
// plan_sections — a path measured FAILING at exactly that on 8 of 9 pages in
// one run (bugs_open/412 §8-9), for reasons the diagnosis loop returned
// UNVERIFIABLE on (the 114 lane's GAP 4). Meanwhile the LLM-free rerender path
// merges STORED content_data — so a value that never reaches storage cannot be
// served by the light path either. The result is bugs_open/114's class: the
// asset serves 200 and the page points at the generic site fallback for ever.
//
// THE FIX SHAPE: write the deployed URL into the page's hero-family component
// rows' content_data AT THE LANDING EVENT, in the same transaction as the
// re-render emit — the same construction as emitContentCardDerive (IMG-073),
// which has fired 193/193 at this seam. The write is a deterministic FLOOR
// under the flaky resolver, not a competitor to it: plan_sections either
// re-resolves the same value (Lane B reads the same asset by the same key
// convention), or resolves nothing and carryStored carries the stored value.
// Either way the page serves its own image.
//
// WHAT IT WRITES, AND — LOAD-BEARING — WHAT IT REFUSES TO TOUCH:
//   - only rows whose component is HERO-FAMILY (function 'hero', 'hero-*',
//     '*-hero', or category 'hero') — measured 2026-09-02, that predicate
//     separates the 8 hero variants from the interactive calculators that also
//     mention background_image in their templates;
//   - never a row whose rendered_html carries a bugs_open/357 structural
//     marker (the hero-declares-tool shape): the write would be inert AND
//     would recreate the false "wired" reading that confused the mcalc census;
//   - never a PAGE-SPECIFIC value: it fills hero_url only where BOTH hero_url
//     and background_image are empty, the legacy literal, or the site-wide
//     fallback — i.e. exactly where the page is showing the generic default,
//     which is 114's symptom. A value the planner set (hero-about.jpg) is not
//     fought.
//   - the URL is storage.DeployedWebPath(ContentHeroKey(page), purpose) — the
//     deployer's own derivation. Re-deriving it is the two-derivations defect
//     that poisoned sites.content_data for a month (IMG-072).
//
// OPT-IN, DEFAULT OFF (owner ruling 2026-08-02 §2: new authority on a shared
// seam ships as a field whose unsafe default is OFF). The step config key
// `wire_hero_on_landing` arms it; the held migration
// 710_arm_wire_hero_on_landing_HOLD.sql turns it on for image-build-handler
// after a chassis image carrying this file has rolled. Unarmed, this function
// is never called and behaviour is byte-identical to today.
//
// ⚠ ACCEPTANCE CAVEAT, recorded before anyone measures: finetuning.uk is
// EXCLUDED as evidence either way — not merely because migration 664 (08-31)
// hand-wired its heroes, but because (the 412 lane's own refinement, 09-02)
// 664 changed the JOIN while 649 changed the SCHEMA, so two defects overlap
// there and a single before/after cannot attribute a delta to either. Measure
// on IMG-077's `unwired` rollup sites instead — that state's predicate already
// requires an image-capable, non-fragment component, which is exactly the
// pre-check that keeps a not-image-capable instance from reading as this bug.

package actions

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/gqls/agentchassis/platform/orchestration/actions/discovery_checks"
	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"github.com/gqls/agentchassis/platform/orchestration/imageryplan"
	"github.com/gqls/agentchassis/platform/storage"
	"go.uber.org/zap"
)

// legacyHeroFallbackLiteral is the pre-IMG-072 poisoned default. Pages built
// before the 562 repair still carry it verbatim in content_data.
const legacyHeroFallbackLiteral = "/assets/images/hero.jpg"

// heroWireArmed is the opt-in gate, extracted pure so the OFF default is
// pinned by a test rather than by a source scan (a source-scan test makes
// comments load-bearing).
func heroWireArmed(config map[string]interface{}) bool {
	return configBoolOrDefault(config, "wire_hero_on_landing", false)
}

// wirePageHeroOnLanding writes the just-landed content hero's deployed URL into
// the page's hero-family component rows, where they currently show the generic
// fallback. Returns a disposition string for the caller's log line — like its
// sibling emitContentCardDerive it never returns an error, because failing the
// whole action (leaving the item failed and the page un-re-rendered) is a worse
// outcome than not wiring; every skip is visible.
func wirePageHeroOnLanding(
	ctx context.Context,
	tx *sql.Tx,
	siteID uuid.UUID,
	pageName string,
	logger *zap.Logger,
) string {
	heroKey := imageryplan.ContentHeroKey(pageName)

	// One round trip: does the page's content hero exist and what purpose does
	// it carry (for the path derivation), plus the site-wide fallback the
	// value-gate compares against.
	var purpose, siteFallback string
	err := tx.QueryRowContext(ctx, `
		SELECT COALESCE((SELECT a.purpose FROM assets a
		                  WHERE a.site_id = $1 AND a.asset_key = $2 AND a.status = 'active'
		                  LIMIT 1), ''),
		       COALESCE((SELECT s.content_data->>'hero_url' FROM sites s WHERE s.id = $1), '')
	`, siteID, heroKey).Scan(&purpose, &siteFallback)
	switch {
	case err != nil:
		logger.Warn("wire_page_hero_on_landing: lookup failed, hero not wired",
			zap.String("page", pageName), zap.Error(err))
		return "skipped_lookup_failed"
	case purpose == "":
		// Expected for a logo, icon or any non-content-hero landing routed
		// through the same page flow.
		return "skipped_no_content_hero"
	}

	webPath := storage.DeployedWebPath(heroKey, purpose)
	if webPath == "" {
		// DeployedWebPath contract: "" is unresolvable — skip loudly, never guess.
		logger.Warn("wire_page_hero_on_landing: no deployable web path for asset, hero not wired",
			zap.String("asset_key", heroKey), zap.String("purpose", purpose))
		return "skipped_no_web_path"
	}

	// The write. Every arm of the WHERE is load-bearing and mutation-tested:
	// hero-family only; non-fragment only (bugs_open/357); and the value-gate —
	// BOTH image fields currently empty / legacy / the site fallback, so a
	// page-specific value is never fought. jsonb_set on a COALESCEd object so a
	// NULL content_data row is wired rather than skipped.
	res, err := tx.ExecContext(ctx, `
		UPDATE page_components pc
		   SET content_data = jsonb_set(COALESCE(pc.content_data, '{}'::jsonb),
		                                '{hero_url}', to_jsonb($4::text)),
		       updated_at   = now()
		  FROM pages p, content_components cc
		 WHERE p.id = pc.page_id
		   AND p.site_id = $1 AND p.name = $2
		   AND cc.id = pc.component_id
		   AND (cc.function = 'hero' OR cc.function LIKE 'hero-%'
		        OR cc.function LIKE '%-hero' OR cc.category = 'hero')
		   AND `+datahelpers.NotRemoved("pc")+`
		   AND NOT (`+fragmentMarkerPredicate("pc.rendered_html")+`)
		   AND COALESCE(pc.content_data->>'hero_url', '') IN ('', $3, $5)
		   AND COALESCE(pc.content_data->>'background_image', '') IN ('', $3, $5)
	`, siteID, pageName, legacyHeroFallbackLiteral, webPath, siteFallback)
	if err != nil {
		logger.Warn("wire_page_hero_on_landing: write failed, hero not wired",
			zap.String("page", pageName), zap.Error(err))
		return "skipped_write_failed"
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		// Three honest causes fold here: no hero-family row on the page, the
		// row is a 357 fragment, or the row already carries a page-specific
		// value. The rollup census (IMG-077) is what tells them apart.
		return "skipped_no_eligible_component"
	}

	logger.Info("wire_page_hero_on_landing: wired the landed content hero into the page's stored hero fields (bugs_open/412 candidate 1)",
		zap.String("site_id", siteID.String()),
		zap.String("page", pageName),
		zap.String("web_path", webPath),
		zap.Int64("components", n))
	return fmt.Sprintf("wired:%d", n)
}

// fragmentMarkerPredicate builds the 357 structural-marker OR-chain over the
// given column, from the SHARED list (discovery_checks.InteractiveStructuralMarkers
// — this package's private copy is lockstep-pinned to it), so a marker added
// there is honoured here without a second edit.
func fragmentMarkerPredicate(column string) string {
	out := ""
	for i, m := range discovery_checks.InteractiveStructuralMarkers {
		if i > 0 {
			out += " OR "
		}
		out += column + " LIKE '%" + m + "%'"
	}
	return out
}
