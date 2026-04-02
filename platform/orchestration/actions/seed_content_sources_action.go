// FILE: platform/orchestration/actions/seed_content_sources_action.go
//
// SeedContentSourcesAction reads the classification spec's
// content_features.news_feed recommendation and creates content_sources
// rows for the site if none exist yet.
//
// This bridges the gap between EvaluateNewsFeedAction (which writes the
// recommendation to the classification spec) and DispatchFeedSourcesAction
// (which queries content_sources for sources to fetch). Without this action,
// sites get a news_feed recommendation but never get any actual sources
// created — so the ingestion pipeline has nothing to work with.
//
// Idempotent: uses ON CONFLICT DO NOTHING on the (site_id, name) unique
// index so repeated calls are safe.
//
// Source creation logic:
//   - news_search: one source per vertical_keyword
//   - api_news: one source with all keywords (xAI/Grok)
//   - rss: skipped (requires manual URL — logged as info)
//   - scrape: skipped (requires manual URL)
//
// Registration (add to registry.go):
//
//	"seed_content_sources": {
//	    Handler:     SeedContentSourcesAction,
//	    Category:    "feed",
//	    Description: "Create content_sources from classification news_feed recommendation",
//	    IsLocal:     true,
//	},
//
// Workflow placement: first step of content-feed-orchestrator, before dispatch_sources.
//
//	"seed_sources": {
//	    "action": "seed_content_sources",
//	    "config": {"site_id": "input_data.site_id"},
//	    "next_step": "dispatch_sources",
//	    "output_field": "seed_result",
//	    "description": "Ensure content_sources exist from classification spec"
//	}

package actions

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"go.uber.org/zap"
)

var SeedContentSourcesInputSpec = datahelpers.ActionInputSpec{
	Required: []string{"site_id"},
	Optional: []string{},
}

func init() {
	datahelpers.RegisterActionInputSpec("seed_content_sources", SeedContentSourcesInputSpec)
}

// SeedContentSourcesAction checks whether a site needs news sources and
// creates them from the classification spec if they don't already exist.
func SeedContentSourcesAction(ctx context.Context, params ActionParams) (interface{}, error) {
	logger := params.Logger.With(zap.String("action", "seed_content_sources"))

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}
	if params.DB == nil {
		return nil, fmt.Errorf("database connection required")
	}

	inputs, err := datahelpers.ExtractActionInputs(
		params.CollectedData,
		params.StepConfig.Config,
		SeedContentSourcesInputSpec,
		logger,
	)
	if err != nil {
		return nil, fmt.Errorf("input extraction failed: %w", err)
	}

	siteID, err := uuid.Parse(inputs.Get("site_id"))
	if err != nil {
		return nil, fmt.Errorf("invalid site_id: %w", err)
	}

	// Check if sources already exist for this site
	var existingCount int
	err = params.DB.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM content_sources
		WHERE site_id = $1 AND is_active = true
	`, siteID).Scan(&existingCount)
	if err != nil {
		return nil, fmt.Errorf("count existing sources: %w", err)
	}

	if existingCount > 0 {
		logger.Info("SeedContentSourcesAction: sources already exist, skipping seed",
			zap.String("site_id", siteID.String()),
			zap.Int("existing_count", existingCount))
		return map[string]interface{}{
			"site_id":        siteID.String(),
			"seeded":         0,
			"has_sources":    true,
			"existing_count": existingCount,
		}, nil
	}

	// Read classification spec for news_feed recommendation
	var specDataJSON []byte
	err = params.DB.QueryRowContext(ctx, `
		SELECT data FROM site_specs
		WHERE site_id = $1 AND aspect = 'classification' AND is_current = true
	`, siteID).Scan(&specDataJSON)
	if err != nil {
		logger.Info("SeedContentSourcesAction: no classification spec, nothing to seed",
			zap.String("site_id", siteID.String()))
		return map[string]interface{}{
			"site_id":     siteID.String(),
			"seeded":      0,
			"has_sources": false,
			"reason":      "no classification spec",
		}, nil
	}

	var specData map[string]interface{}
	if err := json.Unmarshal(specDataJSON, &specData); err != nil {
		return nil, fmt.Errorf("unmarshal classification spec: %w", err)
	}

	// Navigate to content_features.news_feed
	newsFeed := extractNewsFeedConfig(specData)
	if newsFeed == nil {
		logger.Info("SeedContentSourcesAction: no news_feed config in classification spec",
			zap.String("site_id", siteID.String()))
		return map[string]interface{}{
			"site_id":     siteID.String(),
			"seeded":      0,
			"has_sources": false,
			"reason":      "no news_feed in classification spec",
		}, nil
	}

	recommended, _ := newsFeed["recommended"].(bool)
	if !recommended {
		logger.Info("SeedContentSourcesAction: news_feed not recommended for this site",
			zap.String("site_id", siteID.String()))
		return map[string]interface{}{
			"site_id":     siteID.String(),
			"seeded":      0,
			"has_sources": false,
			"reason":      "news_feed not recommended",
		}, nil
	}

	// Extract source_types and vertical_keywords
	sourceTypes := extractStringSlice(newsFeed, "source_types")
	verticalKeywords := extractStringSlice(newsFeed, "vertical_keywords")

	if len(sourceTypes) == 0 {
		logger.Info("SeedContentSourcesAction: no source_types specified",
			zap.String("site_id", siteID.String()))
		return map[string]interface{}{
			"site_id":     siteID.String(),
			"seeded":      0,
			"has_sources": false,
			"reason":      "no source_types in news_feed config",
		}, nil
	}

	if len(verticalKeywords) == 0 {
		logger.Info("SeedContentSourcesAction: no vertical_keywords specified",
			zap.String("site_id", siteID.String()))
		return map[string]interface{}{
			"site_id":     siteID.String(),
			"seeded":      0,
			"has_sources": false,
			"reason":      "no vertical_keywords in news_feed config",
		}, nil
	}

	// Also load domain for prompt context
	var domain string
	_ = params.DB.QueryRowContext(ctx, `
		SELECT domain FROM sites WHERE id = $1
	`, siteID).Scan(&domain)

	// Create sources per type
	seeded := 0
	skippedTypes := []string{}

	for _, sourceType := range sourceTypes {
		switch sourceType {
		case "news_search":
			n, err := seedNewsSearchSources(ctx, params, siteID, verticalKeywords, logger)
			if err != nil {
				logger.Warn("SeedContentSourcesAction: failed to seed news_search sources",
					zap.Error(err))
				continue
			}
			seeded += n

		case "api_news":
			n, err := seedAPINewsSources(ctx, params, siteID, domain, verticalKeywords, logger)
			if err != nil {
				logger.Warn("SeedContentSourcesAction: failed to seed api_news source",
					zap.Error(err))
				continue
			}
			seeded += n

		case "rss":
			// RSS requires specific feed URLs — can't auto-discover.
			// Discovery agents or manual config should add these later.
			logger.Info("SeedContentSourcesAction: skipping rss (requires manual URL config)",
				zap.String("site_id", siteID.String()))
			skippedTypes = append(skippedTypes, "rss")

		case "scrape":
			// Scrape requires specific page URLs — can't auto-discover.
			logger.Info("SeedContentSourcesAction: skipping scrape (requires manual URL config)",
				zap.String("site_id", siteID.String()))
			skippedTypes = append(skippedTypes, "scrape")

		default:
			logger.Warn("SeedContentSourcesAction: unknown source_type",
				zap.String("source_type", sourceType))
			skippedTypes = append(skippedTypes, sourceType)
		}
	}

	logger.Info("SeedContentSourcesAction: seeding complete",
		zap.String("site_id", siteID.String()),
		zap.Int("seeded", seeded),
		zap.Strings("skipped_types", skippedTypes),
		zap.Strings("vertical_keywords", verticalKeywords))

	return map[string]interface{}{
		"site_id":       siteID.String(),
		"seeded":        seeded,
		"has_sources":   seeded > 0,
		"skipped_types": skippedTypes,
		"source_types":  sourceTypes,
		"keywords":      verticalKeywords,
	}, nil
}

// seedNewsSearchSources creates one content_source per vertical keyword
// with source_type = 'news_search'.
func seedNewsSearchSources(ctx context.Context, params ActionParams, siteID uuid.UUID, keywords []string, logger *zap.Logger) (int, error) {
	seeded := 0
	for _, keyword := range keywords {
		keyword = strings.TrimSpace(keyword)
		if keyword == "" {
			continue
		}

		name := fmt.Sprintf("News Search: %s", keyword)
		config := map[string]interface{}{
			"query":       keyword,
			"num_results": 10,
		}
		configJSON, err := json.Marshal(config)
		if err != nil {
			logger.Warn("seedNewsSearchSources: marshal config failed",
				zap.String("keyword", keyword), zap.Error(err))
			continue
		}

		// ON CONFLICT DO NOTHING — idx_cs_site_name is UNIQUE on (site_id, name)
		result, err := params.DB.ExecContext(ctx, `
			INSERT INTO content_sources (site_id, source_type, name, config)
			VALUES ($1, 'news_search', $2, $3::jsonb)
			ON CONFLICT (site_id, name) DO NOTHING
		`, siteID, name, string(configJSON))
		if err != nil {
			logger.Warn("seedNewsSearchSources: insert failed",
				zap.String("keyword", keyword), zap.Error(err))
			continue
		}

		rows, _ := result.RowsAffected()
		if rows > 0 {
			seeded++
			logger.Info("seedNewsSearchSources: created source",
				zap.String("name", name),
				zap.String("keyword", keyword))
		}
	}
	return seeded, nil
}

// seedAPINewsSources creates one content_source with source_type = 'api_news'
// containing all vertical keywords. Uses xAI/Grok as the provider.
func seedAPINewsSources(ctx context.Context, params ActionParams, siteID uuid.UUID, domain string, keywords []string, logger *zap.Logger) (int, error) {
	// Build a prompt template from the keywords
	keywordList := strings.Join(keywords, ", ")
	promptTemplate := fmt.Sprintf(
		"Find the %d most important news items from the last {{.hours}} hours about: %s. "+
			"Focus on developments that would matter to professionals and businesses in this sector. "+
			"Return results as a JSON array.",
		10, keywordList,
	)

	name := "LLM News: " + domain
	if domain == "" {
		// Fallback: use first keyword
		if len(keywords) > 0 {
			name = "LLM News: " + keywords[0]
		} else {
			name = "LLM News: general"
		}
	}

	config := map[string]interface{}{
		"provider":        "xai",
		"model":           "grok-3-mini",
		"prompt_template": promptTemplate,
		"hours_lookback":  12,
		"max_items":       10,
		"keywords":        keywords,
	}
	configJSON, err := json.Marshal(config)
	if err != nil {
		return 0, fmt.Errorf("marshal api_news config: %w", err)
	}

	result, err := params.DB.ExecContext(ctx, `
		INSERT INTO content_sources (site_id, source_type, name, config)
		VALUES ($1, 'api_news', $2, $3::jsonb)
		ON CONFLICT (site_id, name) DO NOTHING
	`, siteID, name, string(configJSON))
	if err != nil {
		return 0, fmt.Errorf("insert api_news source: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows > 0 {
		logger.Info("seedAPINewsSources: created source",
			zap.String("name", name),
			zap.String("domain", domain))
		return 1, nil
	}

	return 0, nil
}

// extractNewsFeedConfig navigates specData to find content_features.news_feed.
func extractNewsFeedConfig(specData map[string]interface{}) map[string]interface{} {
	contentFeatures, ok := specData["content_features"].(map[string]interface{})
	if !ok {
		return nil
	}
	newsFeed, ok := contentFeatures["news_feed"].(map[string]interface{})
	if !ok {
		return nil
	}
	return newsFeed
}
