// FILE: platform/orchestration/actions/discovery_checks/check_news_feed.go
//
// Discovery checks for the news feed pipeline. Detects:
//   - missing_news_sources: spec recommends news but no content_sources exist
//   - missing_news_section: sources exist with items but no latest-news on homepage
//   - stale_news_section:   homepage has latest-news but items are older than threshold
//   - all_sources_erroring: all content_sources for a site have errors
//   - missing_news_page:    spec says separate_page=true but no news-index page exists
//
// All five are Layer 1 (algorithmic, no LLM). They run as part of the
// completeness-discovery-agent alongside empty_sections, empty_blog, etc.
//
// Handler agent routing:
//   - missing_news_sources  → content-feed-orchestrator (seeds sources via SeedContentSourcesAction)
//   - missing_news_section  → content-gap-planner (LLM decides how to add latest-news to homepage)
//   - stale_news_section    → content-feed-orchestrator (re-runs ingest + triage + render + commit)
//   - all_sources_erroring  → content-feed-orchestrator (re-seeds may replace broken sources)
//   - missing_news_page     → content-gap-planner (creates /news.html with news-listing component)
//
// To enable, add these check names to the completeness-discovery-agent's
// "checks" config array (already done in current deployment):
//
//   UPDATE agent_definitions
//   SET default_config = jsonb_set(
//       default_config,
//       '{workflow,steps,run_checks,config,checks}',
//       '["empty_sections", "empty_blog", "orphan_pages",
//         "missing_news_sources", "missing_news_section", "stale_news_section",
//         "all_sources_erroring", "missing_news_page"]'::jsonb
//   )
//   WHERE type = 'completeness-discovery-agent' AND deleted_at IS NULL;

package discovery_checks

import (
	"encoding/json"
	"fmt"
	"time"

	"go.uber.org/zap"
)

func init() {
	Register(&MissingNewsSourcesCheck{})
	Register(&MissingNewsSectionCheck{})
	Register(&StaleNewsSectionCheck{})
	Register(&AllSourcesErroringCheck{})
	Register(&MissingNewsPageCheck{})
}

// ===========================================================================
// MissingNewsSourcesCheck
// ===========================================================================
// Detects: site spec has content_features.news_feed.recommended = true
// but no content_sources rows exist for the site.
//
// Handler: content-feed-orchestrator — its first step is seed_content_sources
// which reads the classification spec and creates content_sources rows.
// After seeding it dispatches ingesters, runs triage, and renders.

type MissingNewsSourcesCheck struct{}

func (c *MissingNewsSourcesCheck) Name() string { return "missing_news_sources" }

func (c *MissingNewsSourcesCheck) Run(dctx DiscoveryCheckContext) (*CheckResult, error) {
	// Check if site spec recommends news
	var specData []byte
	err := dctx.DB.QueryRowContext(dctx.Ctx, `
		SELECT data FROM site_specs
		WHERE site_id = $1 AND aspect = 'classification' AND is_current = true
	`, dctx.SiteID).Scan(&specData)
	if err != nil {
		// No classification spec — can't determine if news is recommended
		return &CheckResult{}, nil
	}

	var spec map[string]interface{}
	if err := json.Unmarshal(specData, &spec); err != nil {
		return &CheckResult{}, nil
	}

	// Navigate to content_features.news_feed.recommended
	contentFeatures, _ := spec["content_features"].(map[string]interface{})
	if contentFeatures == nil {
		return &CheckResult{}, nil
	}
	newsFeed, _ := contentFeatures["news_feed"].(map[string]interface{})
	if newsFeed == nil {
		return &CheckResult{}, nil
	}
	recommended, _ := newsFeed["recommended"].(bool)
	if !recommended {
		return &CheckResult{}, nil
	}

	// Check if content_sources exist
	var sourceCount int
	err = dctx.DB.QueryRowContext(dctx.Ctx, `
		SELECT COUNT(*) FROM content_sources
		WHERE site_id = $1 AND is_active = true
	`, dctx.SiteID).Scan(&sourceCount)
	if err != nil {
		return nil, fmt.Errorf("missing_news_sources: query failed: %w", err)
	}

	if sourceCount > 0 {
		return &CheckResult{}, nil
	}

	// Spec says news, no sources exist
	dctx.Logger.Info("MissingNewsSourcesCheck: spec recommends news but no sources configured",
		zap.String("site_id", dctx.SiteID.String()))

	specJSON, _ := json.Marshal(map[string]interface{}{
		"check":             "missing_news_sources",
		"news_feed_config":  newsFeed,
		"vertical_keywords": newsFeed["vertical_keywords"],
	})

	return &CheckResult{
		Findings: []map[string]interface{}{{
			"check":   "missing_news_sources",
			"message": "Site spec recommends news feed but no content_sources are configured",
		}},
		WorkItems: []WorkItemSpec{{
			SiteID:   dctx.SiteID,
			Source:   "discovery",
			Pipeline: "content",
			ItemType: "missing_news_sources",
			Severity: "medium",
			Summary:  "Site spec recommends news feed but no content sources are configured",
			SpecJSON: string(specJSON),
			Priority: 70,
			// CHANGED: was "page-build-handler" — that agent can't seed content sources.
			// content-feed-orchestrator's first step (seed_content_sources) reads the
			// classification spec and creates content_sources rows automatically.
			HandlerAgent: "content-feed-orchestrator",
			Status:       "detected",
			CreatedBy:    dctx.AgentType,
			ItemKey:      fmt.Sprintf("missing_news_sources:%s", dctx.SiteID),
			BatchID:      dctx.BatchID,
		}},
	}, nil
}

// ===========================================================================
// MissingNewsSectionCheck
// ===========================================================================
// Detects: content_sources exist and have relevant/ingested items,
// but no page has a latest-news page_component.
//
// Handler: content-gap-planner — it reads the gap description and decides
// whether to add latest-news to the homepage or create a dedicated page.
// It creates work items for page-build-handler to execute the plan.

type MissingNewsSectionCheck struct{}

func (c *MissingNewsSectionCheck) Name() string { return "missing_news_section" }

func (c *MissingNewsSectionCheck) Run(dctx DiscoveryCheckContext) (*CheckResult, error) {
	// Check if any feed items exist
	var itemCount int
	err := dctx.DB.QueryRowContext(dctx.Ctx, `
		SELECT COUNT(*) FROM content_feed_items
		WHERE site_id = $1
		  AND status IN ('ingested', 'relevant')
	`, dctx.SiteID).Scan(&itemCount)
	if err != nil {
		return nil, fmt.Errorf("missing_news_section: item count query failed: %w", err)
	}

	if itemCount == 0 {
		// No items — nothing to display yet
		return &CheckResult{}, nil
	}

	// Check if a latest-news component exists on any page
	// Two paths: (1) component_id linked to content_components with function='latest-news'
	// (2) rendered_html contains data-component="latest-news" (for pages built before
	//     component_id linking was added — common for older/adopted sites)
	var sectionCount int
	err = dctx.DB.QueryRowContext(dctx.Ctx, `
		SELECT COUNT(*) FROM page_components pc
		JOIN pages p ON p.id = pc.page_id
		WHERE p.site_id = $1
		  AND (
		      -- Path 1: linked via component_id
		      EXISTS (
		          SELECT 1 FROM content_components cc
		          WHERE cc.id = pc.component_id AND cc.function = 'latest-news'
		      )
		      OR
		      -- Path 2: HTML contains the data-component attribute (unlinked rows)
		      (pc.component_id IS NULL AND pc.rendered_html LIKE '%data-component="latest-news"%')
		  )
	`, dctx.SiteID).Scan(&sectionCount)
	if err != nil {
		return nil, fmt.Errorf("missing_news_section: section query failed: %w", err)
	}

	if sectionCount > 0 {
		return &CheckResult{}, nil
	}

	// Items exist but no section to display them
	dctx.Logger.Info("MissingNewsSectionCheck: feed items exist but no latest-news section",
		zap.String("site_id", dctx.SiteID.String()),
		zap.Int("item_count", itemCount))

	specJSON, _ := json.Marshal(map[string]interface{}{
		"check":       "missing_news_section",
		"item_count":  itemCount,
		"page_name":   "index",
		"description": "Site has ingested news feed items but the homepage has no latest-news section to display them. Consider adding a latest-news component to the homepage.",
		"suggestion":  "Add a latest-news section to the homepage to show recent industry news",
		"category":    "content_completeness",
	})

	return &CheckResult{
		Findings: []map[string]interface{}{{
			"check":      "missing_news_section",
			"message":    "Feed items exist but homepage has no latest-news section",
			"item_count": itemCount,
		}},
		WorkItems: []WorkItemSpec{{
			SiteID:   dctx.SiteID,
			Source:   "discovery",
			Pipeline: "content",
			ItemType: "missing_news_section",
			Severity: "medium",
			Summary:  fmt.Sprintf("Site has %d feed items but no latest-news section on homepage", itemCount),
			SpecJSON: string(specJSON),
			Priority: 65,
			// CHANGED: was "page-build-handler" — it can't decide where to add a section.
			// content-gap-planner reads the gap spec, decides whether to add to an
			// existing page or create a new one, then creates targeted work items
			// for page-build-handler.
			HandlerAgent: "content-gap-planner",
			Status:       "detected",
			CreatedBy:    dctx.AgentType,
			ItemKey:      fmt.Sprintf("missing_news_section:%s", dctx.SiteID),
			BatchID:      dctx.BatchID,
		}},
	}, nil
}

// ===========================================================================
// StaleNewsSectionCheck
// ===========================================================================
// Detects: homepage has a latest-news component but the newest relevant
// item is older than the freshness threshold (default 72 hours).
//
// Handler: content-feed-orchestrator — re-runs the full pipeline:
// seed (noop if sources exist) → dispatch ingesters → triage → render → commit.
// This is the right fix because stale news means we need fresh content,
// not a page rerender.

type StaleNewsSectionCheck struct{}

func (c *StaleNewsSectionCheck) Name() string { return "stale_news_section" }

func (c *StaleNewsSectionCheck) Run(dctx DiscoveryCheckContext) (*CheckResult, error) {
	// Check if a latest-news component exists on any page
	// Handles both linked (component_id set) and unlinked (data-component in HTML) rows
	var pageID, pageComponentID string
	var pageName string
	err := dctx.DB.QueryRowContext(dctx.Ctx, `
		SELECT p.id::text, p.name, pc.id::text
		FROM page_components pc
		JOIN pages p ON p.id = pc.page_id
		WHERE p.site_id = $1
		  AND (
		      EXISTS (
		          SELECT 1 FROM content_components cc
		          WHERE cc.id = pc.component_id AND cc.function = 'latest-news'
		      )
		      OR (pc.component_id IS NULL AND pc.rendered_html LIKE '%data-component="latest-news"%')
		  )
		LIMIT 1
	`, dctx.SiteID).Scan(&pageID, &pageName, &pageComponentID)
	if err != nil {
		// No latest-news section — not our problem (MissingNewsSectionCheck handles that)
		return &CheckResult{}, nil
	}

	// Find the newest relevant item
	var newestItemAge float64 // hours
	err = dctx.DB.QueryRowContext(dctx.Ctx, `
		SELECT EXTRACT(EPOCH FROM (NOW() - MAX(
			COALESCE(source_published_at, created_at)
		))) / 3600.0
		FROM content_feed_items
		WHERE site_id = $1
		  AND status IN ('relevant', 'ingested')
	`, dctx.SiteID).Scan(&newestItemAge)
	if err != nil {
		// No items at all — will be caught by other checks
		return &CheckResult{}, nil
	}

	// Default freshness threshold: 72 hours (3 days)
	freshnessThresholdHours := 72.0

	// Check site settings for custom threshold
	var settingsJSON []byte
	if err := dctx.DB.QueryRowContext(dctx.Ctx, `
		SELECT COALESCE(settings->'maintenance_profile'->'content_feed'->>'max_age_hours', '72')
		FROM sites WHERE id = $1
	`, dctx.SiteID).Scan(&settingsJSON); err == nil {
		// Parse custom threshold
		var customHours float64
		if json.Unmarshal(settingsJSON, &customHours) == nil && customHours > 0 {
			freshnessThresholdHours = customHours
		}
	}

	if newestItemAge < freshnessThresholdHours {
		// News is fresh enough
		return &CheckResult{}, nil
	}

	// News section is stale
	staleDays := newestItemAge / 24
	dctx.Logger.Info("StaleNewsSectionCheck: news section is stale",
		zap.String("site_id", dctx.SiteID.String()),
		zap.Float64("newest_item_hours", newestItemAge),
		zap.Float64("threshold_hours", freshnessThresholdHours))

	specJSON, _ := json.Marshal(map[string]interface{}{
		"check":               "stale_news_section",
		"page_name":           pageName,
		"page_id":             pageID,
		"page_component_id":   pageComponentID,
		"newest_item_age_hrs": int(newestItemAge),
		"threshold_hours":     int(freshnessThresholdHours),
	})

	return &CheckResult{
		Findings: []map[string]interface{}{{
			"check":            "stale_news_section",
			"message":          fmt.Sprintf("Latest news on %s is %.0f days old (threshold: %.0f hours)", pageName, staleDays, freshnessThresholdHours),
			"newest_age_hours": int(newestItemAge),
			"page":             pageName,
		}},
		WorkItems: []WorkItemSpec{{
			SiteID:   dctx.SiteID,
			Source:   "discovery",
			Pipeline: "content",
			ItemType: "stale_news_section",
			Severity: "low",
			Summary:  fmt.Sprintf("News section on %s is %.0f days old", pageName, staleDays),
			SpecJSON: string(specJSON),
			Priority: 80,
			// CHANGED: was "rerender-pages" — rerendering doesn't bring new content.
			// content-feed-orchestrator re-runs ingest + triage + render + commit,
			// which produces fresh news items.
			HandlerAgent: "content-feed-orchestrator",
			Status:       "detected",
			CreatedBy:    dctx.AgentType,
			ItemKey:      fmt.Sprintf("stale_news_section:%s:%s", dctx.SiteID, pageName),
			BatchID:      dctx.BatchID,
		}},
	}, nil
}

// ===========================================================================
// AllSourcesErroringCheck
// ===========================================================================
// Detects: all content_sources for a site have error_count > 0
//
// Handler: content-feed-orchestrator — its seed_content_sources step will
// attempt to create new sources from the classification spec if none exist
// or if all existing ones are erroring. The dispatch step will re-try
// fetching from sources (error_count doesn't block dispatch, only
// exponential backoff on next_fetch_at does).

type AllSourcesErroringCheck struct{}

func (c *AllSourcesErroringCheck) Name() string { return "all_sources_erroring" }

func (c *AllSourcesErroringCheck) Run(dctx DiscoveryCheckContext) (*CheckResult, error) {
	var totalSources, erroringSources int
	err := dctx.DB.QueryRowContext(dctx.Ctx, `
		SELECT
			COUNT(*) as total,
			COUNT(*) FILTER (WHERE error_count > 0) as erroring
		FROM content_sources
		WHERE site_id = $1 AND is_active = true
	`, dctx.SiteID).Scan(&totalSources, &erroringSources)
	if err != nil {
		return nil, fmt.Errorf("all_sources_erroring: query failed: %w", err)
	}

	if totalSources == 0 || erroringSources < totalSources {
		return &CheckResult{}, nil
	}

	dctx.Logger.Info("AllSourcesErroringCheck: all sources have errors",
		zap.String("site_id", dctx.SiteID.String()),
		zap.Int("total", totalSources),
		zap.Int("erroring", erroringSources))

	// Load source details for the spec
	rows, err := dctx.DB.QueryContext(dctx.Ctx, `
		SELECT name, source_type, error_count, COALESCE(last_error, '')
		FROM content_sources
		WHERE site_id = $1 AND is_active = true AND error_count > 0
		ORDER BY error_count DESC
		LIMIT 10
	`, dctx.SiteID)
	if err != nil {
		return nil, fmt.Errorf("all_sources_erroring: detail query failed: %w", err)
	}
	defer rows.Close()

	var sources []map[string]interface{}
	for rows.Next() {
		var name, sourceType, lastError string
		var errorCount int
		if err := rows.Scan(&name, &sourceType, &errorCount, &lastError); err != nil {
			continue
		}
		sources = append(sources, map[string]interface{}{
			"name":        name,
			"source_type": sourceType,
			"error_count": errorCount,
			"last_error":  lastError,
		})
	}

	specJSON, _ := json.Marshal(map[string]interface{}{
		"check":            "all_sources_erroring",
		"total_sources":    totalSources,
		"erroring_sources": erroringSources,
		"sources":          sources,
		"checked_at":       time.Now().UTC().Format(time.RFC3339),
	})

	return &CheckResult{
		Findings: []map[string]interface{}{{
			"check":          "all_sources_erroring",
			"message":        fmt.Sprintf("All %d content sources have errors", totalSources),
			"source_details": sources,
		}},
		WorkItems: []WorkItemSpec{{
			SiteID:   dctx.SiteID,
			Source:   "discovery",
			Pipeline: "content",
			ItemType: "all_sources_erroring",
			Severity: "high",
			Summary:  fmt.Sprintf("All %d content sources are failing — feed URLs may be dead", totalSources),
			SpecJSON: string(specJSON),
			Priority: 50,
			// CHANGED: was "" (empty) — dispatch would fail trying to spawn empty agent type.
			// content-feed-orchestrator will re-seed sources (adding new ones from the
			// classification spec) and re-dispatch. If the original sources are truly dead,
			// the new seeded sources may work. If not, this item will recur and eventually
			// need manual attention (error_count keeps incrementing).
			HandlerAgent: "content-feed-orchestrator",
			Status:       "detected",
			CreatedBy:    dctx.AgentType,
			ItemKey:      fmt.Sprintf("all_sources_erroring:%s", dctx.SiteID),
			BatchID:      dctx.BatchID,
		}},
	}, nil
}

// ===========================================================================
// MissingNewsPageCheck
// ===========================================================================
// Detects: classification spec has news_feed.separate_page = true
// but no page with page_type = 'news-index' exists.
//
// This fires after the evaluate_news_feed enrichment step has run and
// set separate_page = true for verticals where news is a primary draw
// (energy, sports, finance, technology, etc).
//
// Handler: content-gap-planner — reads the gap spec and creates a news
// page with news-listing component, then creates work items for
// page-build-handler to generate the content and deploy.

type MissingNewsPageCheck struct{}

func (c *MissingNewsPageCheck) Name() string { return "missing_news_page" }

func (c *MissingNewsPageCheck) Run(dctx DiscoveryCheckContext) (*CheckResult, error) {
	// Check if site spec recommends a separate news page
	var specData []byte
	err := dctx.DB.QueryRowContext(dctx.Ctx, `
		SELECT data FROM site_specs
		WHERE site_id = $1 AND aspect = 'classification' AND is_current = true
	`, dctx.SiteID).Scan(&specData)
	if err != nil {
		return &CheckResult{}, nil
	}

	var spec map[string]interface{}
	if err := json.Unmarshal(specData, &spec); err != nil {
		return &CheckResult{}, nil
	}

	contentFeatures, _ := spec["content_features"].(map[string]interface{})
	if contentFeatures == nil {
		return &CheckResult{}, nil
	}
	newsFeed, _ := contentFeatures["news_feed"].(map[string]interface{})
	if newsFeed == nil {
		return &CheckResult{}, nil
	}

	// Check both recommended and separate_page flags
	recommended, _ := newsFeed["recommended"].(bool)
	separatePage, _ := newsFeed["separate_page"].(bool)

	if !recommended || !separatePage {
		return &CheckResult{}, nil
	}

	// Check if a news-index page already exists
	var pageCount int
	err = dctx.DB.QueryRowContext(dctx.Ctx, `
		SELECT COUNT(*) FROM pages
		WHERE site_id = $1 AND page_type = 'news-index'
	`, dctx.SiteID).Scan(&pageCount)
	if err != nil {
		return nil, fmt.Errorf("missing_news_page: page query failed: %w", err)
	}

	if pageCount > 0 {
		return &CheckResult{}, nil
	}

	// Also check if there are any content sources or items — no point creating
	// a news page if there's nothing to show yet. The missing_news_sources
	// check handles that case separately.
	var sourceCount int
	err = dctx.DB.QueryRowContext(dctx.Ctx, `
		SELECT COUNT(*) FROM content_sources
		WHERE site_id = $1 AND is_active = true
	`, dctx.SiteID).Scan(&sourceCount)
	if err != nil {
		return nil, fmt.Errorf("missing_news_page: source query failed: %w", err)
	}

	if sourceCount == 0 {
		// No sources yet — missing_news_sources will fire first,
		// and this check will pick up on the next discovery pass
		// once sources exist.
		return &CheckResult{}, nil
	}

	dctx.Logger.Info("MissingNewsPageCheck: spec recommends separate news page but none exists",
		zap.String("site_id", dctx.SiteID.String()),
		zap.Int("source_count", sourceCount))

	// Get domain for the description
	var domain string
	_ = dctx.DB.QueryRowContext(dctx.Ctx, `
		SELECT domain FROM sites WHERE id = $1
	`, dctx.SiteID).Scan(&domain)

	specJSON, _ := json.Marshal(map[string]interface{}{
		"check":            "missing_news_page",
		"page_name":        "news",
		"page_type":        "news-index",
		"description":      fmt.Sprintf("Site %s is recommended for a dedicated news page (separate_page=true) but no news-index page exists. Create a /news.html page with a news-listing component to display the full news archive.", domain),
		"suggestion":       "Create a new page named 'news' with page_type 'news-index', a hero section, the news-listing component, and a call-to-action. Add it to header and footer navigation.",
		"category":         "content_completeness",
		"news_feed_config": newsFeed,
	})

	return &CheckResult{
		Findings: []map[string]interface{}{{
			"check":   "missing_news_page",
			"message": fmt.Sprintf("Site spec recommends a separate news page but no news-index page exists for %s", domain),
		}},
		WorkItems: []WorkItemSpec{{
			SiteID:       dctx.SiteID,
			Source:       "discovery",
			Pipeline:     "content",
			ItemType:     "missing_news_page",
			Severity:     "medium",
			Summary:      fmt.Sprintf("Site %s needs a dedicated /news.html page (separate_page=true in spec)", domain),
			SpecJSON:     string(specJSON),
			Priority:     60,
			HandlerAgent: "content-gap-planner",
			Status:       "detected",
			CreatedBy:    dctx.AgentType,
			ItemKey:      fmt.Sprintf("missing_news_page:%s", dctx.SiteID),
			BatchID:      dctx.BatchID,
		}},
	}, nil
}
