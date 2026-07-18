// FILE: platform/orchestration/actions/discovery_checks/check_content_image_missing.go
//
// Discovery check: a site LISTS its articles somewhere (a component consumes
// `query.blog_posts`), but one or more of those articles is missing its
// content imagery. Two-mode emitter (Phase I3; D13 2026-07-16):
//
//   - GENERATE — the article has NO image of its own (no planner-emitted page
//     hero, no Lane B content hero): emit a `needs_imagery` GENERATION item
//     under the ContentHeroKey convention, with the prompt composed from the
//     article's own title + description. It rides image-build-handler's
//     proven generic path (call_imagery_gen → store → deploy → flag_rebuild),
//     and the per-site imagery style guide layers medium/mood/palette onto
//     the prompt at generation time exactly as for plan imagery. The article
//     page re-renders with its new hero via the normal image-landed flow.
//
//   - DERIVE — the article HAS a source image (plan hero, or a content hero
//     from a previous pass) but its CARD is missing or STALE-BY-ORIGIN
//     (card.origin_asset_id no longer matches the current preferred source —
//     e.g. the card was cut from the site fallback before a per-article hero
//     existed, or the hero was regenerated): emit the `needs_content_image`
//     derive item for asset-deployer's content_card mode.
//
// Convergence: pass 1 generates missing content heroes; pass 2 sees each
// card's origin no longer matches and re-derives; pass 3 is silent. The
// entity link + origin lineage ARE the fulfilment stamp — no separate stamp.
//
// The site-scope brand hero is deliberately NOT a generation-suppressing
// source (D13: every listed article deserves its own image); it remains the
// last-resort DERIVE source inside derive_card_asset only.
//
// v1 covers entity_type='page' (articles); news items (I5) and products (I6)
// get their own sweeps against the same entity link. DB-ONLY by house
// convention. Cost containment: contentImageMaxPerPass caps TOTAL emissions
// (generation + derive) per pass, and dedup item keys stop double-queueing.

package discovery_checks

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/gqls/agentchassis/platform/orchestration/actions/queryresolve"
	"github.com/gqls/agentchassis/platform/orchestration/imageryplan"
	"go.uber.org/zap"
)

func init() { Register(&ContentImageMissingCheck{}) }

type ContentImageMissingCheck struct{}

func (c *ContentImageMissingCheck) Name() string { return "content_image_missing" }

// contentImageMaxPerPass caps emissions per discovery pass — a site adopting
// this check with a large article back-catalogue drains over a few passes
// instead of flooding the queue (and, for generation items, the image API).
const contentImageMaxPerPass = 10

// contentImageRow is one listed article's imagery state, as swept from the DB.
type contentImageRow struct {
	PageID          string
	PageName        string
	Title           string
	MetaDescription string
	PlanHeroID      string // "" = no planner-emitted page hero asset
	ContentHeroID   string // "" = no Lane B content hero asset
	CardID          string // "" = no entity-linked card asset
	CardOriginID    string // card's origin_asset_id lineage ("" if null)
}

// contentImageAction is the pure per-article decision (testable without a DB):
// "generate" when the article has no image of its own; "derive" when a source
// exists but the card is missing or stale-by-origin; "" when fulfilled.
func contentImageAction(r contentImageRow) string {
	source := r.PlanHeroID
	if source == "" {
		source = r.ContentHeroID
	}
	if source == "" {
		return "generate"
	}
	if r.CardID == "" || r.CardOriginID != source {
		return "derive"
	}
	return ""
}

// contentHeroPrompt composes the generation subject from the article's own
// content. Deliberately subject-only: the per-site imagery style guide
// prepends medium/mood/palette at generation time (imagery_style_guide.go),
// so this stays one brand voice without double-direction.
func contentHeroPrompt(title, metaDescription string) string {
	subject := strings.TrimSpace(title)
	if d := strings.TrimSpace(metaDescription); d != "" {
		subject += " — " + d
	}
	return fmt.Sprintf(
		"Article header image representing: %s. Concrete subject matter from the article topic, no text or lettering in the image.",
		subject)
}

func (c *ContentImageMissingCheck) Run(dctx DiscoveryCheckContext) (*CheckResult, error) {
	// Gate: something on this site actually lists articles.
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

	// Sweep every listed-type page's imagery state in one query. The inline
	// 'content_hero_' || replace(...) MUST match imageryplan.ContentHeroKey.
	//
	// Eligibility (F2.1, 2026-07-17): only articles that actually shipped.
	// Plan-era scaffold rows and never-built /blog/ duplicates sit
	// status='active' with empty sections; generating imagery for them is
	// wasted spend on pages that 404. The predicate is queryresolve's shared
	// constant, not a copy: the listing and this sweep must agree on which
	// articles exist, and two hand-maintained strings would drift silently.
	rows, err := dctx.DB.QueryContext(dctx.Ctx, `
		SELECT p.id::text, p.name,
		       COALESCE(p.title, p.name)          AS title,
		       COALESCE(p.meta_description, '')   AS meta_description,
		       COALESCE(ph.id::text, '')          AS plan_hero_id,
		       COALESCE(ch.id::text, '')          AS content_hero_id,
		       COALESCE(ca.id::text, '')          AS card_id,
		       COALESCE(ca.origin_asset_id::text, '') AS card_origin_id
		  FROM pages p
		  LEFT JOIN LATERAL (
		      SELECT a.id
		        FROM site_plan_imagery spi
		        JOIN site_plans sp ON sp.id = spi.plan_id AND sp.is_current = true
		        JOIN assets a ON a.site_id = p.site_id AND a.asset_key = spi.key AND a.status = 'active'
		       WHERE sp.site_id = p.site_id AND spi.kind = 'hero'
		         AND spi.scope = 'page' AND spi.scope_ref = p.name
		       LIMIT 1
		  ) ph ON true
		  LEFT JOIN assets ch
		    ON ch.site_id = p.site_id AND ch.status = 'active'
		   AND ch.asset_key = 'content_hero_' || replace(p.name, '-', '_')
		  LEFT JOIN assets ca
		    ON ca.site_id = p.site_id AND ca.entity_type = 'page'
		   AND ca.entity_id = p.id AND ca.purpose = 'card' AND ca.status = 'active'
		 WHERE p.site_id = $1
		   AND p.page_type = 'blog-post'
		   AND p.status IN ('active', 'deployed')`+
		queryresolve.ListedPageEligibilitySQL+`
		 ORDER BY p.name
	`, dctx.SiteID)
	if err != nil {
		return nil, fmt.Errorf("content_image_missing: sweep failed: %w", err)
	}
	defer rows.Close()

	result := &CheckResult{}
	emitted := 0
	for rows.Next() {
		var r contentImageRow
		if err := rows.Scan(&r.PageID, &r.PageName, &r.Title, &r.MetaDescription,
			&r.PlanHeroID, &r.ContentHeroID, &r.CardID, &r.CardOriginID); err != nil {
			dctx.Logger.Warn("content_image_missing: scan failed", zap.Error(err))
			continue
		}
		if emitted >= contentImageMaxPerPass {
			dctx.Logger.Info("content_image_missing: per-pass cap reached; remainder next pass",
				zap.Int("cap", contentImageMaxPerPass))
			break
		}

		action := contentImageAction(r)
		if action == "" {
			continue
		}

		var item WorkItemSpec
		switch action {
		case "generate":
			item, err = c.generationItem(dctx, r)
		case "derive":
			item, err = c.deriveItem(dctx, r)
		}
		if err != nil {
			return nil, err
		}

		result.Findings = append(result.Findings, map[string]interface{}{
			"check":     c.Name(),
			"action":    action,
			"page_name": r.PageName,
			"entity_id": r.PageID,
		})
		result.WorkItems = append(result.WorkItems, item)
		emitted++
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("content_image_missing: rows iter failed: %w", err)
	}

	if emitted > 0 {
		dctx.Logger.Info("content_image_missing: emitted items",
			zap.Int("count", emitted), zap.Int("cap", contentImageMaxPerPass))
	}
	return result, nil
}

// generationItem shapes the Lane B content-hero GENERATION request. It is a
// standard needs_imagery item (imageryplan.BuildSpec) so image-build-handler's
// generic path handles it unchanged; scope/scope_ref make flag_rebuild
// re-render the article page when the image lands.
func (c *ContentImageMissingCheck) generationItem(dctx DiscoveryCheckContext, r contentImageRow) (WorkItemSpec, error) {
	row := imageryplan.Row{
		Scope:    "page",
		ScopeRef: &r.PageName,
		Key:      imageryplan.ContentHeroKey(r.PageName),
		// content_hero (D14, was "hero"): its own kind so routing (Banana,
		// which honours style anchors — the Stability path does not) and the
		// style guide's per-kind override can differ from plan heroes. The
		// D13 gate failed on exactly this: SDXL drifted off the free-text
		// style direction card-to-card (colour, medium, text artefacts).
		Kind:       "content_hero",
		Prompt:     contentHeroPrompt(r.Title, r.MetaDescription),
		StyleHints: json.RawMessage(`{"aspect_ratio":"16:9"}`),
	}
	specJSON, err := imageryplan.BuildSpec(row, c.Name())
	if err != nil {
		return WorkItemSpec{}, fmt.Errorf("content_image_missing: build generation spec: %w", err)
	}
	priority, severity := imageryplan.Classify(row.Scope, row.Kind, row.ScopeRef)
	return WorkItemSpec{
		SiteID:       dctx.SiteID,
		Source:       "discovery",
		Pipeline:     "build",
		ItemType:     "needs_imagery",
		Severity:     severity,
		Summary:      fmt.Sprintf("Listed article %q has no image of its own — generate its content hero", r.PageName),
		SpecJSON:     specJSON,
		Priority:     priority,
		HandlerAgent: "image-build-handler",
		Status:       "detected",
		CreatedBy:    dctx.AgentType,
		ItemKey:      imageryplan.ItemKey(row),
		BatchID:      dctx.BatchID,
	}, nil
}

// deriveItem shapes the card DERIVE request for asset-deployer's content_card
// mode (derive_card_asset re-crops the article's current preferred source and
// refreshes the entity link + origin lineage).
func (c *ContentImageMissingCheck) deriveItem(dctx DiscoveryCheckContext, r contentImageRow) (WorkItemSpec, error) {
	specJSON, err := contentImageSpecJSON(c.Name(), r.PageID, r.PageName)
	if err != nil {
		return WorkItemSpec{}, err
	}
	reason := "has no card image"
	if r.CardID != "" {
		reason = "card is stale (derived from a superseded source)"
	}
	return WorkItemSpec{
		SiteID:       dctx.SiteID,
		Source:       "discovery",
		Pipeline:     "build",
		ItemType:     "needs_content_image",
		Severity:     "low",
		Summary:      fmt.Sprintf("Article %q %s — derive from its hero", r.PageName, reason),
		SpecJSON:     specJSON,
		Priority:     65,
		HandlerAgent: "asset-deployer",
		Status:       "detected",
		CreatedBy:    dctx.AgentType,
		ItemKey:      contentImageItemKey(r.PageName),
		BatchID:      dctx.BatchID,
	}, nil
}
