// FILE: platform/orchestration/actions/discovery_checks/check_undeployed_assets.go
//
// "Which of this site's images were generated but never deployed?"
//
// ── bugs_open/142: the check had a population that could not contain the
//    answer, and evidence that looked in the wrong table ────────────────────
//
// It ran ONE query, `FROM assets a WHERE a.site_id = $1 AND NOT EXISTS (…a
// deployed page_component whose rendered_html mentions the filename…)`. That is
// wrong twice for the two brand-head artefacts (favicon, og_card):
//
//  1. THE POPULATION WAS THE ARTEFACT TABLE. A site whose og card was never
//     generated has no `assets` row, so it was structurally unexaminable. A
//     detector whose denominator is the artefact table cannot see a MISSING
//     artefact — it can only ever report on artefacts that exist. Measured
//     2026-07-31: idea.uk (61 deployed components) and webdesign.co.uk (6) each
//     served 404 for both brand-head files, had no asset row for either, and
//     had never produced a single finding.
//
//  2. THE DEPLOY EVIDENCE WAS THE WRONG TABLE. favicon and og_card are
//     referenced from the site HEAD — `injectBrandHeadTags` writes them into
//     `site_components` (slot `head`) — and never from a page component. So a
//     present, serving artefact could never look deployed and fired for ever.
//     Measured the same day: the shipping predicate would have raised 96
//     findings across all 14 live sites, including favicon+og_card on 12 of
//     them, while all 12 served both files 200 on the wire.
//
// The two halves now have separate populations, because they are separate
// questions:
//
//	page-referenced assets  (hero, icon, illustration, card, …)
//	    population: the site's `assets` rows        — unchanged, still correct
//	    evidence:   a deployed page_component's rendered_html
//
//	brand-head assets       (favicon, og_card)
//	    population: the brand-head purposes THEMSELVES, for any site with
//	                deployed pages — so absence is representable
//	    evidence:   an active `assets` row (see "why the row is the evidence")
//
// ── WHY THE ROW IS THE DEPLOY EVIDENCE FOR BRAND-HEAD ARTEFACTS ────────────
//
// Not because the two correlate — because one causes the other.
// `derive_brand_head_assets` calls `sendGitCommitRequest` FIRST and returns on
// its error; `recordDerivedAsset` runs only after the commit succeeded
// (derive_brand_head_assets_action.go:175-185). The row is written iff the file
// was committed to the site repo.
//
// Checked against the wire rather than assumed, 2026-07-31, all 15 sites with
// pages: 12 with an active og_card row served og-card.png 200; the 3 without
// served 404. No exception in either direction. The head reference is NOT
// evidence and is deliberately not used as such — `injectBrandHeadTags` emits
// the <link>/<meta> unconditionally, so 13 of 14 heads advertise a card whether
// or not one exists (idea.uk's advertises a 404).
//
// Residual, stated rather than hidden: the row over-reports if a file is later
// deleted from the site repo without the row being retired. That direction is a
// missed finding, not a false one — strictly weaker than the failure it
// replaces, which fired on every working site for ever.
//
// ── LANDMINE: DO NOT "FIX" THE LIKE WILDCARD IN THE PAGE-ASSET QUERY ───────
//
// `purpose` values contain underscores (`content_hero`, `sprite_sheet`,
// `hero_home`) and SQL LIKE treats `_` as ANY CHARACTER, so the pattern built
// from `content_hero` matches the real published filename `content-hero…`,
// which is spelled with a HYPHEN. The wildcard is doing load-bearing work by
// accident. Measured 2026-07-31: 38 of 38 `content_hero` assets read as
// deployed with the wildcard live and 0 of 38 with `_` escaped — so "correctly"
// escaping it manufactures 38 false findings. Pinned by
// TestUnderscoreWildcardIsLoadBearing.
//
// ── STATUS='detected' IS DELIBERATE ────────────────────────────────────────
//
// Every finding here lands in the queue bugs_open/083 is about, which has no
// running promoter. That is not fixed here and must not be fixed by writing
// 'triaged' directly — 083's own landmine says so, because it would bypass
// triage for every finding the platform raises. This change makes the detector
// correct; it does not make the queue drain.

package discovery_checks

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"sort"

	"go.uber.org/zap"

	"github.com/gqls/agentchassis/platform/storage"
)

func init() { Register(&UndeployedAssetsCheck{}) }

type UndeployedAssetsCheck struct{}

func (c *UndeployedAssetsCheck) Name() string { return "undeployed_assets" }

func (c *UndeployedAssetsCheck) Run(dctx DiscoveryCheckContext) (*CheckResult, error) {
	result := &CheckResult{}

	// ── Half 1: page-referenced assets ──────────────────────────────────────
	assets, err := findUndeployedAssets(dctx)
	if err != nil {
		return nil, err
	}
	if len(assets) > 0 {
		result.Findings = append(result.Findings, map[string]interface{}{
			"check":  "undeployed_assets",
			"count":  len(assets),
			"assets": assets,
		})
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
	}

	// ── Half 2: brand-head artefacts, whose population is the purpose list ──
	gaps, provenanceNotes, err := findBrandHeadAssetGaps(dctx)
	if err != nil {
		return nil, err
	}
	for _, note := range provenanceNotes {
		result.Findings = append(result.Findings, map[string]interface{}{
			"check":         "undeployed_assets",
			"observation":   "brand_head_provenance_url_unexpected",
			"purpose":       note.Purpose,
			"expected_path": note.ExpectedPath,
			"actual_url":    note.ActualURL,
			"note": "an active row exists but does not record the published path, so it is evidence " +
				"of neither deployment nor absence; raising a gap here would be a false claim " +
				"(see bugs_open/152 for the URL-rewrite defect that produces these)",
		})
	}
	for _, gap := range gaps {
		specJSON, _ := json.Marshal(map[string]interface{}{
			"check":           "undeployed_assets",
			"purpose":         gap.Purpose,
			"expected_path":   gap.ExpectedPath,
			"head_references": gap.HeadReferences,
			"has_logo":        gap.HasLogo,
		})

		// The handler's own precondition, predicted by the item rather than
		// discovered by it: derive_brand_head_assets returns
		// {"derived": false, "reason": "no active logo asset"} when the site has
		// no active asset_key='logo' row. Say so in the summary rather than
		// suppressing the finding — a check that silently drops the case it
		// cannot route is the failing-branch-with-no-branch shape this file was
		// filed about. (Measured 2026-07-31: all 14 sites with deployed pages
		// have one, so this reads as unreachable today. It is written for the
		// day it is not.)
		summary := fmt.Sprintf("Brand-head asset '%s' has never been generated — the site serves nothing at %s",
			gap.Purpose, gap.ExpectedPath)
		if gap.HeadReferences {
			summary += ", and the site head already points at it"
		}
		if !gap.HasLogo {
			summary += " (blocked: no active logo asset to derive from)"
		}

		result.Findings = append(result.Findings, map[string]interface{}{
			"check":           "undeployed_assets",
			"gap":             "brand_head_asset_missing",
			"purpose":         gap.Purpose,
			"expected_path":   gap.ExpectedPath,
			"head_references": gap.HeadReferences,
			"has_logo":        gap.HasLogo,
		})
		result.WorkItems = append(result.WorkItems, WorkItemSpec{
			SiteID:       dctx.SiteID,
			Source:       "discovery",
			Pipeline:     "design",
			ItemType:     "needs_brand_head_assets",
			Severity:     "high",
			Summary:      summary,
			SpecJSON:     string(specJSON),
			Priority:     60,
			HandlerAgent: "asset-deployer",
			Status:       "detected",
			CreatedBy:    dctx.AgentType,
			ItemKey:      fmt.Sprintf("needs_brand_head_assets:%s", gap.Purpose),
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

// findUndeployedAssets covers PAGE-REFERENCED assets only. Brand-head purposes
// are excluded because their reference lives in site_components, so this
// predicate can only ever return them as false positives — see the file header.
//
// The `LIKE … || purpose || …` pattern is deliberately left unescaped. Read the
// LANDMINE section of the file header before changing it.
func findUndeployedAssets(dctx DiscoveryCheckContext) ([]undeployedAssetFinding, error) {
	rows, err := dctx.DB.QueryContext(dctx.Ctx, `
		SELECT a.id, COALESCE(a.purpose, 'unknown'), a.asset_type, a.url
		FROM assets a
		WHERE a.site_id = $1
		  AND NOT (COALESCE(a.purpose, '') = ANY($2::text[]))
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
	`, dctx.SiteID, pqTextArray(brandHeadPurposes()))
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
	return findings, rows.Err()
}

// brandHeadAssetGap is one brand-head artefact the site does not have.
type brandHeadAssetGap struct {
	Purpose        string `json:"purpose"`
	ExpectedPath   string `json:"expected_path"`
	HeadReferences bool   `json:"head_references"`
	HasLogo        bool   `json:"has_logo"`
}

// brandHeadProvenanceNote records a site that HAS an active row for a brand-head
// purpose whose `url` is not the path the deriver publishes.
//
// Why this is observed but NOT filed as a work item. `recordDerivedAsset` writes
// the published path verbatim, so a row carrying anything else was written by
// some other path — measured 2026-07-31: gamesdesign.co.uk and robot-hands.com
// each hold exactly one active favicon and one active og_card row, both with the
// unresolved template literal `/assets/images/input-data.asset-key.jpg`. Both
// sites nevertheless serve favicon.png and og-card.png 200 on the wire, so the
// artefact IS deployed and the row's provenance is merely wrong.
//
// Filing "has never been generated" against them would therefore be a FALSE
// claim, which is worse than a missed finding — and this check exists because
// the old one made false claims. Filing some other item type would be inventing
// a queue for a defect that belongs to the asset-URL rewrite lane
// (bugs_open/152). So it is surfaced in Findings, where a reader of the run sees
// it, and deliberately raises nothing.
type brandHeadProvenanceNote struct {
	Purpose      string `json:"purpose"`
	ExpectedPath string `json:"expected_path"`
	ActualURL    string `json:"actual_url"`
}

// findBrandHeadAssetGaps iterates the brand-head PURPOSES rather than the site's
// asset rows — the whole point of bugs_open/142. Absence is the finding, so the
// population cannot be the artefact table.
//
// Gated on the site having at least one deployed page component: a site with
// nothing deployed has no public surface to serve a card from, and flagging it
// would file work against every scaffold row in the fleet.
func findBrandHeadAssetGaps(dctx DiscoveryCheckContext) ([]brandHeadAssetGap, []brandHeadProvenanceNote, error) {
	var deployedComponents int
	var hasLogo bool
	err := dctx.DB.QueryRowContext(dctx.Ctx, `
		SELECT
		    (SELECT count(*) FROM pages p
		       JOIN page_components pc ON pc.page_id = p.id
		      WHERE p.site_id = $1 AND pc.build_status = 'deployed'),
		    EXISTS (SELECT 1 FROM assets a
		             WHERE a.site_id = $1 AND a.asset_key = 'logo' AND a.status = 'active')
	`, dctx.SiteID).Scan(&deployedComponents, &hasLogo)
	if err != nil {
		return nil, nil, fmt.Errorf("brand_head population query failed: %w", err)
	}
	if deployedComponents == 0 {
		return nil, nil, nil
	}

	var gaps []brandHeadAssetGap
	var notes []brandHeadProvenanceNote
	for _, purpose := range brandHeadPurposes() {
		expected := storage.BrandHeadAssetPaths[purpose]

		var anyRow, rowAtPublishedPath, headRefs bool
		var sampleURL sql.NullString
		// NOTE: site_components.build_status is 'rendered', NEVER 'deployed'
		// (all 42 rows, 3 slots x 14 sites, measured 2026-07-31). Mirroring the
		// page_components predicate's build_status='deployed' here would make
		// this permanently blind. Do not "make the two consistent".
		err := dctx.DB.QueryRowContext(dctx.Ctx, `
			SELECT
			    EXISTS (SELECT 1 FROM assets a
			             WHERE a.site_id = $1 AND a.purpose = $2 AND a.status = 'active'),
			    EXISTS (SELECT 1 FROM assets a
			             WHERE a.site_id = $1 AND a.purpose = $2 AND a.status = 'active'
			               AND a.url = $3),
			    EXISTS (SELECT 1 FROM site_components sc
			             WHERE sc.site_id = $1 AND sc.slot_name = 'head'
			               AND position($3 IN COALESCE(sc.rendered_html, '')) > 0),
			    (SELECT a.url FROM assets a
			      WHERE a.site_id = $1 AND a.purpose = $2 AND a.status = 'active'
			      ORDER BY a.updated_at DESC LIMIT 1)
		`, dctx.SiteID, purpose, expected).Scan(&anyRow, &rowAtPublishedPath, &headRefs, &sampleURL)
		if err != nil {
			return nil, nil, fmt.Errorf("brand_head check for %q failed: %w", purpose, err)
		}

		switch {
		case rowAtPublishedPath:
			// The deriver's own record of a successful commit. Deployed.
		case anyRow:
			// A row exists but records a different URL, so it is not evidence
			// that THIS artefact was published — and not evidence that it was
			// not. Observed, never filed; see brandHeadProvenanceNote.
			notes = append(notes, brandHeadProvenanceNote{
				Purpose:      purpose,
				ExpectedPath: expected,
				ActualURL:    sampleURL.String,
			})
		default:
			gaps = append(gaps, brandHeadAssetGap{
				Purpose:        purpose,
				ExpectedPath:   expected,
				HeadReferences: headRefs,
				HasLogo:        hasLogo,
			})
		}
	}
	return gaps, notes, nil
}

// brandHeadPurposes returns the brand-head purposes in a stable order. Sorted
// because Go map iteration is randomised and both the emitted work items and
// the tests need a deterministic sequence.
func brandHeadPurposes() []string {
	out := make([]string, 0, len(storage.BrandHeadAssetPaths))
	for purpose := range storage.BrandHeadAssetPaths {
		out = append(out, purpose)
	}
	sort.Strings(out)
	return out
}

// pqTextArray renders a Go string slice as a Postgres text[] literal for use as
// a query parameter. lib/pq's array support is not imported in this package and
// the input is a compile-time constant set, so a literal is sufficient and
// keeps the dependency surface unchanged.
func pqTextArray(values []string) string {
	if len(values) == 0 {
		return "{}"
	}
	out := "{"
	for i, v := range values {
		if i > 0 {
			out += ","
		}
		out += `"` + v + `"`
	}
	return out + "}"
}
