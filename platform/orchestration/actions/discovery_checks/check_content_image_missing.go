// FILE: platform/orchestration/actions/discovery_checks/check_content_image_missing.go
//
// Discovery check: a site LISTS its articles somewhere (a component consumes
// `query.blog_posts`), but one or more of those articles has no entity-linked
// CARD asset — so the listing renders the heavier page hero (or nothing) where
// the purpose-built 800×450 card crop should be. Emits a needs_content_image
// work item per missing card; asset-deployer's content_card mode runs
// derive_card_asset, which crops the page's hero and writes the entity-linked
// assets row.
//
// This is Phase I3 of the imagery workstream (Lane B: content-linked card
// imagery). v1 covers entity_type='page' (articles); news items (I5) and
// products (I6) get their own sweeps against the same entity link.
//
// DB-ONLY, by house convention (see check_image_url_404's header). Quiet
// unless BOTH gates pass:
//   - a component on an active page of this site actually consumes
//     query.blog_posts (no consumer → cards would be invisible work), AND
//   - the page is DERIVABLE: it has a plan hero with an active asset, or the
//     site has a site-scope brand hero to fall back to. Without this gate the
//     handler would complete derived:false and the check would re-emit every
//     pass — the churn the sprite_css_missing stamp exists to prevent.
//
// Idempotence comes from the entity link itself: once derive_card_asset
// upserts the active card row, the sweep's anti-join goes quiet.

package discovery_checks

import (
	"fmt"

	"go.uber.org/zap"
)

func init() { Register(&ContentImageMissingCheck{}) }

type ContentImageMissingCheck struct{}

func (c *ContentImageMissingCheck) Name() string { return "content_image_missing" }

// contentImageMaxPerPass caps emissions per discovery pass — a site adopting
// this check with a large article back-catalogue drains over a few passes
// instead of flooding the queue (same reasoning as imagery's MaxPerPass).
const contentImageMaxPerPass = 10

func (c *ContentImageMissingCheck) Run(dctx DiscoveryCheckContext) (*CheckResult, error) {
	// Gate 1: something on this site actually lists articles.
	var consumers int
	err := dctx.DB.QueryRowContext(dctx.Ctx, `
		SELECT COUNT(*)
		  FROM page_components pc
		  JOIN content_components cc ON cc.id = pc.component_id
		  JOIN pages p ON p.id = pc.page_id
		 WHERE p.site_id = $1
		   AND p.status IN ('active', 'deployed')
		   AND cc.input_schema::text LIKE '%query.blog_posts%'
	`, dctx.SiteID).Scan(&consumers)
	if err != nil {
		return nil, fmt.Errorf("content_image_missing: consumer scan failed: %w", err)
	}
	if consumers == 0 {
		return &CheckResult{}, nil
	}

	// Gate 2 input: a site-scope brand hero makes EVERY page derivable.
	var siteHasBrandHero bool
	err = dctx.DB.QueryRowContext(dctx.Ctx, `
		SELECT EXISTS (
		    SELECT 1
		      FROM site_plan_imagery spi
		      JOIN site_plans sp ON sp.id = spi.plan_id AND sp.is_current = true
		      JOIN assets a ON a.site_id = sp.site_id AND a.asset_key = spi.key AND a.status = 'active'
		     WHERE sp.site_id = $1 AND spi.kind = 'hero' AND spi.scope = 'site')
	`, dctx.SiteID).Scan(&siteHasBrandHero)
	if err != nil {
		return nil, fmt.Errorf("content_image_missing: brand hero scan failed: %w", err)
	}

	// The sweep: listed-type pages with no active entity-linked card, that are
	// derivable (own plan hero, or the site fallback).
	rows, err := dctx.DB.QueryContext(dctx.Ctx, `
		SELECT p.id::text, p.name
		  FROM pages p
		  LEFT JOIN assets ca
		    ON ca.site_id = p.site_id AND ca.entity_type = 'page'
		   AND ca.entity_id = p.id AND ca.purpose = 'card' AND ca.status = 'active'
		 WHERE p.site_id = $1
		   AND p.page_type = 'blog-post'
		   AND p.status IN ('active', 'deployed')
		   AND ca.id IS NULL
		   AND ($2 OR EXISTS (
		        SELECT 1
		          FROM site_plan_imagery spi
		          JOIN site_plans sp ON sp.id = spi.plan_id AND sp.is_current = true
		          JOIN assets a ON a.site_id = p.site_id AND a.asset_key = spi.key AND a.status = 'active'
		         WHERE sp.site_id = p.site_id AND spi.kind = 'hero'
		           AND spi.scope = 'page' AND spi.scope_ref = p.name))
		 ORDER BY p.name
		 LIMIT $3
	`, dctx.SiteID, siteHasBrandHero, contentImageMaxPerPass)
	if err != nil {
		return nil, fmt.Errorf("content_image_missing: sweep failed: %w", err)
	}
	defer rows.Close()

	result := &CheckResult{}
	for rows.Next() {
		var pageID, pageName string
		if err := rows.Scan(&pageID, &pageName); err != nil {
			dctx.Logger.Warn("content_image_missing: scan failed", zap.Error(err))
			continue
		}

		specJSON, err := contentImageSpecJSON(c.Name(), pageID, pageName)
		if err != nil {
			return nil, err
		}

		result.Findings = append(result.Findings, map[string]interface{}{
			"check":       c.Name(),
			"entity_type": "page",
			"entity_id":   pageID,
			"page_name":   pageName,
		})
		result.WorkItems = append(result.WorkItems, WorkItemSpec{
			SiteID: dctx.SiteID,
			Source: "discovery",
			// Pipeline is the destination, not the origin: needs_content_image
			// is handled by asset-deployer on the build pipeline (cf. the same
			// note in check_unfulfilled_imagery_plan).
			Pipeline:     "build",
			ItemType:     "needs_content_image",
			Severity:     "low",
			Summary:      fmt.Sprintf("Article %q is listed but has no card image (derive from its hero)", pageName),
			SpecJSON:     specJSON,
			Priority:     65,
			HandlerAgent: "asset-deployer",
			Status:       "detected",
			CreatedBy:    dctx.AgentType,
			ItemKey:      contentImageItemKey(pageName),
			BatchID:      dctx.BatchID,
		})
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("content_image_missing: rows iter failed: %w", err)
	}

	if n := len(result.WorkItems); n > 0 {
		dctx.Logger.Info("content_image_missing: emitting needs_content_image items",
			zap.Int("count", n), zap.Int("cap", contentImageMaxPerPass))
	}
	return result, nil
}
