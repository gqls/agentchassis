// FILE: platform/orchestration/actions/discovery_checks/check_news_feed.go
//
// Discovery checks for the news feed pipeline. Detects:
//   - missing_news_sources: spec recommends news but no content_sources exist
//   - missing_news_section: sources exist with items but no latest-news on homepage
//   - stale_news_section:   homepage has latest-news but items are older than threshold
//
// All three are Layer 1 (algorithmic, no LLM). They run as part of the
// completeness-discovery-agent alongside empty_sections, empty_blog, etc.
//
// To enable, add these check names to the completeness-discovery-agent's
// "checks" config array:
//
//   UPDATE agent_definitions
//   SET default_config = jsonb_set(
//       default_config,
//       '{workflow,steps,run_checks,config,checks}',
//       '["empty_sections", "empty_blog", "orphan_pages",
//         "missing_news_sources", "missing_news_section", "stale_news_section"]'::jsonb
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
}

// ===========================================================================
// MissingNewsSourcesCheck
// ===========================================================================
// Detects: site spec has content_features.news_feed.recommended = true
// but no content_sources rows exist for the site.
//
// Handler: configure-feed-sources (future) or creates a needs_human_review
// item until that agent exists.

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
			SiteID:       dctx.SiteID,
			Source:       "discovery",
			Pipeline:     "content",
			ItemType:     "missing_news_sources",
			Severity:     "medium",
			Summary:      "Site spec recommends news feed but no content sources are configured",
			SpecJSON:     string(specJSON),
			Priority:     70,
			HandlerAgent: "page-build-handler", // future: configure-feed-sources
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
// Handler: page-build-handler (adds latest-news section to homepage)

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
	var sectionCount int
	err = dctx.DB.QueryRowContext(dctx.Ctx, `
		SELECT COUNT(*) FROM page_components pc
		JOIN content_components cc ON cc.id = pc.component_id
		JOIN pages p ON p.id = pc.page_id
		WHERE p.site_id = $1
		  AND cc.function = 'latest-news'
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
		"check":      "missing_news_section",
		"item_count": itemCount,
		"page_name":  "index",
	})

	return &CheckResult{
		Findings: []map[string]interface{}{{
			"check":      "missing_news_section",
			"message":    "Feed items exist but homepage has no latest-news section",
			"item_count": itemCount,
		}},
		WorkItems: []WorkItemSpec{{
			SiteID:       dctx.SiteID,
			Source:       "discovery",
			Pipeline:     "content",
			ItemType:     "missing_news_section",
			Severity:     "medium",
			Summary:      fmt.Sprintf("Site has %d feed items but no latest-news section on homepage", itemCount),
			SpecJSON:     string(specJSON),
			Priority:     65,
			HandlerAgent: "page-build-handler",
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
// Handler: rerender-pages (rerenders the page with fresh items)

type StaleNewsSectionCheck struct{}

func (c *StaleNewsSectionCheck) Name() string { return "stale_news_section" }

func (c *StaleNewsSectionCheck) Run(dctx DiscoveryCheckContext) (*CheckResult, error) {
	// Check if a latest-news component exists on any page
	var pageID, pageComponentID string
	var pageName string
	err := dctx.DB.QueryRowContext(dctx.Ctx, `
		SELECT p.id::text, p.name, pc.id::text
		FROM page_components pc
		JOIN content_components cc ON cc.id = pc.component_id
		JOIN pages p ON p.id = pc.page_id
		WHERE p.site_id = $1
		  AND cc.function = 'latest-news'
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
			SiteID:       dctx.SiteID,
			Source:       "discovery",
			Pipeline:     "content",
			ItemType:     "stale_news_section",
			Severity:     "low",
			Summary:      fmt.Sprintf("News section on %s is %.0f days old", pageName, staleDays),
			SpecJSON:     string(specJSON),
			Priority:     80,
			HandlerAgent: "rerender-pages",
			Status:       "detected",
			CreatedBy:    dctx.AgentType,
			ItemKey:      fmt.Sprintf("stale_news_section:%s:%s", dctx.SiteID, pageName),
			BatchID:      dctx.BatchID,
		}},
	}, nil
}

// ===========================================================================
// AllSourcesErroringCheck (bonus — catches dead sources)
// ===========================================================================
// Detects: all content_sources for a site have error_count > 0
//
// Handler: needs_human_review (source URLs may have changed or gone offline)

type AllSourcesErroringCheck struct{}

func init() {
	Register(&AllSourcesErroringCheck{})
}

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
			SiteID:       dctx.SiteID,
			Source:       "discovery",
			Pipeline:     "content",
			ItemType:     "all_sources_erroring",
			Severity:     "high",
			Summary:      fmt.Sprintf("All %d content sources are failing — feed URLs may be dead", totalSources),
			SpecJSON:     string(specJSON),
			Priority:     50,
			HandlerAgent: "", // needs_human_review — no automated handler
			Status:       "detected",
			CreatedBy:    dctx.AgentType,
			ItemKey:      fmt.Sprintf("all_sources_erroring:%s", dctx.SiteID),
			BatchID:      dctx.BatchID,
		}},
	}, nil
}
