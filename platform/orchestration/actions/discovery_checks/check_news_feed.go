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
	"strings"
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

	// bugs_open/015: "no news-index page" is NOT the same as "no news page".
	// relojistas.com's planner emitted its Spanish news listing as
	// page_type='section-index', which orphaned it from every gate keying on
	// 'news-index' — including this one. Telling the planner to create a page
	// would have shipped an English /news.html alongside the existing Spanish
	// /noticias rather than fixing the mistyped page.
	//
	// So look for a page already OCCUPYING the role before claiming none
	// exists. The signal is structural, taken from 015's own "how to find
	// others" query — nav-visible, but with no sections, so it can never build
	// and is a live 404 in the navigation. Deliberately NOT a name-vocabulary
	// match ("news"/"noticias"/"nachrichten"…): that is the name-heuristic
	// shape bugs_open/044 was filed against, and it fails on exactly the
	// non-English sites this bug hurts most.
	candidates, cerr := findStrandedNavPages(dctx)
	if cerr != nil {
		return nil, fmt.Errorf("missing_news_page: candidate query failed: %w", cerr)
	}

	gapSpec := map[string]interface{}{
		"check":            "missing_news_page",
		"page_name":        "news",
		"page_type":        "news-index",
		"category":         "content_completeness",
		"news_feed_config": newsFeed,
	}
	summary := fmt.Sprintf("Site %s needs a dedicated /news.html page (separate_page=true in spec)", domain)
	message := fmt.Sprintf("Site spec recommends a separate news page but no news-index page exists for %s", domain)

	if len(candidates) > 0 {
		// The handler is content-gap-planner, an LLM reading these
		// natural-language fields — so changing what they SAY is the whole
		// intervention. Adding a new structured key nobody reads would be the
		// dead-config shape of bugs_open/025 and /042.
		gapSpec["approach"] = "retype_existing"
		gapSpec["retype_candidates"] = candidates
		gapSpec["description"] = fmt.Sprintf(
			"Site %s is recommended for a dedicated news page (separate_page=true) and has NO page of type 'news-index' — but it does have %d navigation-linked page(s) with no sections, which can never build and are dead links today: %s. "+
				"One of these is very likely the news listing already, created under the wrong page_type. Creating a new page would leave a duplicate alongside it, in the wrong language for a non-English site.",
			domain, len(candidates), describeCandidates(candidates))
		gapSpec["suggestion"] = "FIRST decide whether one of the listed stranded pages is already the news listing. If it is, RE-TYPE that page to page_type 'news-index' and give it the sections [hero, news-listing, call-to-action] — do NOT create a second page. Only if none of them is a news listing should you create a new page named 'news'. page_type is a routing key here, not a label: several independent gates key on 'news-index', so the wrong value silences all of them at once."
		summary = fmt.Sprintf("Site %s: news page missing, but %d stranded nav page(s) may be it mistyped — re-type, do not duplicate", domain, len(candidates))
		message = fmt.Sprintf("Site %s has no news-index page but %d nav-linked sectionless page(s) that may be it under the wrong page_type", domain, len(candidates))
	} else {
		gapSpec["approach"] = "new_page"
		gapSpec["description"] = fmt.Sprintf("Site %s is recommended for a dedicated news page (separate_page=true) but no news-index page exists. Create a /news.html page with a news-listing component to display the full news archive.", domain)
		gapSpec["suggestion"] = "Create a new page named 'news' with page_type 'news-index', a hero section, the news-listing component, and a call-to-action. Add it to header and footer navigation."
	}

	specJSON, _ := json.Marshal(gapSpec)

	return &CheckResult{
		Findings: []map[string]interface{}{{
			"check":   "missing_news_page",
			"message": message,
		}},
		WorkItems: []WorkItemSpec{{
			SiteID:       dctx.SiteID,
			Source:       "discovery",
			Pipeline:     "content",
			ItemType:     "missing_news_page",
			Severity:     "medium",
			Summary:      summary,
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

// ===========================================================================
// bugs_open/015 support — pages the navigation links to that can never build
// ===========================================================================

// strandedNavPage is a page linked from the header or footer that has no
// sections. page-build-handler no-ops it ("no sections ready to build") into
// needs_human_review, so it never deploys and the nav link stays a live 404.
// On relojistas.com this was the news listing itself, emitted by the planner as
// page_type='section-index' instead of 'news-index'.
type strandedNavPage struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	PageType    string `json:"page_type"`
	BuildStatus string `json:"build_status"`
}

// strandedNavPageCap bounds what goes into a work-item spec. Reported rather
// than silently applied — a truncated list that reads as complete is how a
// caller concludes "that is all of them".
const strandedNavPageCap = 10

// findStrandedNavPages returns nav-linked, section-less pages that have never
// successfully built and are not already typed 'news-index'.
//
// The predicate is deliberately STRUCTURAL — nav-visible, sections-empty,
// not deployed — and not a name match. A vocabulary list ("news", "noticias",
// "nachrichten") is the name-heuristic shape bugs_open/044 was filed against,
// and it would fail worst on the non-English sites this bug hurts most:
// relojistas' page is called "noticias-index", and the next site's will be
// called something else again.
//
// The `build_status <> 'deployed'` clause is NOT incidental, and was added
// after checking the predicate against live data rather than reasoning about
// it. Without it the query returned six pages on ai-agent-orchestration.com —
// tool pages and a blog index, all deployed and working, whose `sections` are
// legitimately empty because their content comes from elsewhere. Feeding those
// to the planner as pages that "can never build and are dead links today"
// would have been false, and would have invited it to re-type a live tool page
// into a news index. An empty `sections` array only means "stranded" when the
// page never built; on a deployed page it means the sections are somebody
// else's job.
//
// Vocabulary check (live, 2026-07-20): build_status has exactly three values —
// deployed (215), needs_rebuild (50), planned (20). needs_rebuild is kept
// deliberately: sectionless and nav-linked, it cannot rebuild either.
func findStrandedNavPages(dctx DiscoveryCheckContext) ([]strandedNavPage, error) {
	rows, err := dctx.DB.QueryContext(dctx.Ctx, `
		SELECT id::text, name, COALESCE(page_type, ''), COALESCE(build_status, '')
		FROM pages
		WHERE site_id = $1
		  AND COALESCE(page_type, '') <> 'news-index'
		  AND (COALESCE(in_header, false) OR COALESCE(in_footer, false))
		  AND jsonb_array_length(COALESCE(sections, '[]'::jsonb)) = 0
		  AND COALESCE(build_status, '') <> 'deployed'
		ORDER BY name
	`, dctx.SiteID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []strandedNavPage
	for rows.Next() {
		var p strandedNavPage
		if err := rows.Scan(&p.ID, &p.Name, &p.PageType, &p.BuildStatus); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	if len(out) > strandedNavPageCap {
		dctx.Logger.Info("missing_news_page: stranded nav page list capped",
			zap.Int("found", len(out)),
			zap.Int("reported", strandedNavPageCap))
		out = out[:strandedNavPageCap]
	}
	return out, nil
}

// describeCandidates renders the candidates for the natural-language spec the
// content-gap-planner reads. Name, current type and build status are all
// load-bearing: the type is the thing that is wrong, and the build status is
// what shows the page has never successfully built.
func describeCandidates(candidates []strandedNavPage) string {
	parts := make([]string, 0, len(candidates))
	for _, c := range candidates {
		pt := c.PageType
		if pt == "" {
			pt = "(no page_type)"
		}
		bs := c.BuildStatus
		if bs == "" {
			bs = "(no build_status)"
		}
		parts = append(parts, fmt.Sprintf("'%s' (page_type=%s, build_status=%s)", c.Name, pt, bs))
	}
	return strings.Join(parts, ", ")
}
