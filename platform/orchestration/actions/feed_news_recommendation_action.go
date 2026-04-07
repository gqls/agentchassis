// FILE: platform/orchestration/actions/feed_news_recommendation_action.go
//
// EvaluateNewsFeedAction runs after classification to determine whether
// the site should have a news feed. Reads the classification spec (industry,
// site_type) and writes content_features.news_feed into the classification
// aspect via deep merge.
//
// This avoids modifying the classifier's LLM prompt. It's a deterministic
// enrichment step: industry X → yes/no news recommendation + vertical keywords.
//
// CHANGE: Added SeparatePage field to verticalNewsConfig. When true, the
// discovery check (missing_news_page) will detect the need for a dedicated
// /news.html page and route it through content-gap-planner.
//
// Registration:
//   "evaluate_news_feed": {
//       Handler:     EvaluateNewsFeedAction,
//       Category:    "feed",
//       Description: "Determine if site should have a news feed based on classification",
//       IsLocal:     true,
//   },
//
// Workflow placement: after write_classification_spec, before planner
//
//   "evaluate_news_feed": {
//       "action": "evaluate_news_feed",
//       "config": {
//           "site_id": "site_record.site_id"
//       },
//       "output_field": "news_feed_recommendation",
//       "next_step": "call_planner"
//   }

package actions

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"go.uber.org/zap"
)

var EvaluateNewsFeedInputSpec = datahelpers.ActionInputSpec{
	Required: []string{"site_id"},
	Optional: []string{},
}

func init() {
	datahelpers.RegisterActionInputSpec("evaluate_news_feed", EvaluateNewsFeedInputSpec)
}

// verticalNewsConfig holds the news recommendation for an industry vertical.
type verticalNewsConfig struct {
	Recommended      bool     `json:"recommended"`
	Reason           string   `json:"reason"`
	VerticalKeywords []string `json:"vertical_keywords"`
	SourceTypes      []string `json:"source_types"`
	SeparatePage     bool     `json:"separate_page"` // true = create /news.html listing page
}

// verticalNewsMap maps industry/site_type keywords to news recommendations.
// This is the deterministic routing table — no LLM needed.
// Add new verticals here when expanding.
//
// SeparatePage: true for verticals where news is a primary audience draw
// (energy, sports, finance, technology). False for verticals where news
// is supplementary (mortgage, insurance, legal).
var verticalNewsMap = map[string]verticalNewsConfig{
	// Energy — news is primary draw (daily price movements)
	"energy": {
		Recommended:      true,
		Reason:           "Energy markets have daily price movements and supply disruption news",
		VerticalKeywords: []string{"energy prices", "oil prices", "gas prices", "energy market", "supply disruption"},
		SourceTypes:      []string{"rss", "news_search", "api_news"},
		SeparatePage:     true,
	},
	"gas": {
		Recommended:      true,
		Reason:           "Gas wholesale markets change daily — price, supply, and regulatory news",
		VerticalKeywords: []string{"wholesale gas prices", "natural gas market", "gas supply", "LNG", "energy regulation"},
		SourceTypes:      []string{"rss", "news_search", "api_news"},
		SeparatePage:     true,
	},
	"oil": {
		Recommended:      true,
		Reason:           "Oil markets are highly news-driven with global supply and geopolitical factors",
		VerticalKeywords: []string{"oil prices", "crude oil", "OPEC", "petroleum", "fuel prices"},
		SourceTypes:      []string{"rss", "news_search", "api_news"},
		SeparatePage:     true,
	},

	// Finance — news is primary draw (market data)
	"finance": {
		Recommended:      true,
		Reason:           "Financial services benefit from market news for authority and SEO freshness",
		VerticalKeywords: []string{"financial markets", "interest rates", "economic outlook", "investment"},
		SourceTypes:      []string{"news_search", "api_news"},
		SeparatePage:     true,
	},
	"mortgage": {
		Recommended:      true,
		Reason:           "Mortgage rates change frequently — rate news drives return visits",
		VerticalKeywords: []string{"mortgage rates", "interest rates", "housing market", "property prices"},
		SourceTypes:      []string{"news_search", "api_news"},
		SeparatePage:     false, // supplementary — homepage snippet is enough
	},
	"insurance": {
		Recommended:      true,
		Reason:           "Insurance industry has regulatory and market news relevant to consumers",
		VerticalKeywords: []string{"insurance market", "insurance regulation", "claims", "premiums"},
		SourceTypes:      []string{"news_search"},
		SeparatePage:     false, // supplementary
	},

	// Sports — news is primary draw (event-driven)
	"boxing": {
		Recommended:      true,
		Reason:           "Boxing has frequent fight announcements, results, and rankings changes",
		VerticalKeywords: []string{"boxing news", "boxing results", "fight announcements", "boxing rankings"},
		SourceTypes:      []string{"rss", "news_search", "api_news"},
		SeparatePage:     true,
	},
	"sports": {
		Recommended:      true,
		Reason:           "Sports verticals are inherently news-driven",
		VerticalKeywords: []string{"sports news", "match results", "tournament", "league standings"},
		SourceTypes:      []string{"rss", "news_search"},
		SeparatePage:     true,
	},
	"mma": {
		Recommended:      true,
		Reason:           "MMA/UFC has frequent event news and fight announcements",
		VerticalKeywords: []string{"MMA news", "UFC", "fight card", "MMA results"},
		SourceTypes:      []string{"rss", "news_search", "api_news"},
		SeparatePage:     true,
	},

	// Technology — news is primary draw (fast-moving)
	"technology": {
		Recommended:      true,
		Reason:           "Tech sector moves fast — product launches, security news, industry shifts",
		VerticalKeywords: []string{"technology news", "tech industry", "software", "cybersecurity"},
		SourceTypes:      []string{"rss", "news_search"},
		SeparatePage:     true,
	},
	"saas": {
		Recommended: false,
		Reason:      "SaaS product sites benefit more from case studies and product updates than external news",
	},
	"ai": {
		Recommended:      true,
		Reason:           "AI field changes rapidly — research, product launches, regulation",
		VerticalKeywords: []string{"artificial intelligence", "machine learning", "AI regulation", "AI research"},
		SourceTypes:      []string{"rss", "news_search", "api_news"},
		SeparatePage:     true,
	},

	// Healthcare — not recommended
	"veterinary": {
		Recommended: false,
		Reason:      "Local vet practices benefit more from pet care advice content than industry news",
	},
	"dental": {
		Recommended: false,
		Reason:      "Local dental practices benefit from oral health content, not industry news",
	},
	"healthcare": {
		Recommended: false,
		Reason:      "Healthcare sites need careful content curation due to medical advice regulations",
	},

	// Real estate — supplementary
	"property": {
		Recommended:      true,
		Reason:           "Property market has regular price, rate, and regulatory news",
		VerticalKeywords: []string{"property market", "house prices", "property news", "planning permission"},
		SourceTypes:      []string{"news_search", "api_news"},
		SeparatePage:     false,
	},
	"estate-agent": {
		Recommended:      true,
		Reason:           "Local property market news adds value and SEO freshness",
		VerticalKeywords: []string{"property market", "house prices", "local property news", "stamp duty"},
		SourceTypes:      []string{"news_search"},
		SeparatePage:     false,
	},

	// Legal — supplementary
	"legal": {
		Recommended:      true,
		Reason:           "Legal sector has regulatory changes and case law news",
		VerticalKeywords: []string{"legal news", "regulation", "court ruling", "legislation"},
		SourceTypes:      []string{"news_search"},
		SeparatePage:     false,
	},

	// Food & hospitality — not recommended
	"restaurant": {
		Recommended: false,
		Reason:      "Restaurant sites benefit from menus and events, not external news",
	},
	"catering": {
		Recommended: false,
		Reason:      "Catering sites are better served by portfolio and testimonials",
	},

	// Construction — supplementary
	"construction": {
		Recommended:      true,
		Reason:           "Construction has material price, regulation, and project news",
		VerticalKeywords: []string{"construction news", "building materials", "construction regulation", "planning"},
		SourceTypes:      []string{"news_search"},
		SeparatePage:     false,
	},
}

// EvaluateNewsFeedAction checks whether the site's industry/vertical
// should have a news feed and writes the recommendation to the
// classification spec via deep merge.
func EvaluateNewsFeedAction(ctx context.Context, params ActionParams) (interface{}, error) {
	logger := params.Logger.With(zap.String("action", "evaluate_news_feed"))

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}
	if params.DB == nil {
		return nil, fmt.Errorf("database connection required")
	}

	inputs, err := datahelpers.ExtractActionInputs(
		params.CollectedData,
		params.StepConfig.Config,
		EvaluateNewsFeedInputSpec,
		logger,
	)
	if err != nil {
		return nil, fmt.Errorf("input extraction failed: %w", err)
	}

	siteID, err := uuid.Parse(inputs.Get("site_id"))
	if err != nil {
		return nil, fmt.Errorf("invalid site_id: %w", err)
	}

	// Read current classification spec
	var specDataJSON []byte
	err = params.DB.QueryRowContext(ctx, `
		SELECT data FROM site_specs
		WHERE site_id = $1 AND aspect = 'classification' AND is_current = true
	`, siteID).Scan(&specDataJSON)
	if err == sql.ErrNoRows {
		logger.Info("EvaluateNewsFeedAction: no classification spec yet, skipping")
		return map[string]interface{}{
			"recommended": false,
			"reason":      "no classification spec available",
		}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("query classification spec: %w", err)
	}

	var specData map[string]interface{}
	if err := json.Unmarshal(specDataJSON, &specData); err != nil {
		return nil, fmt.Errorf("unmarshal spec: %w", err)
	}

	// Extract industry signals from classification
	industry, _ := specData["industry"].(string)
	siteType, _ := specData["site_type"].(string)
	category, _ := specData["category"].(string)

	// Also check identity spec for domain hints
	var domain string
	_ = params.DB.QueryRowContext(ctx, `
		SELECT domain FROM sites WHERE id = $1
	`, siteID).Scan(&domain)

	// Match against vertical news map
	config := matchVerticalNews(industry, siteType, category, domain, logger)

	if config == nil {
		// No match — default to not recommended
		logger.Info("EvaluateNewsFeedAction: no vertical match, defaulting to no news",
			zap.String("industry", industry),
			zap.String("site_type", siteType))

		return map[string]interface{}{
			"recommended": false,
			"reason":      "no matching vertical profile for news",
			"industry":    industry,
			"site_type":   siteType,
		}, nil
	}

	// Write the recommendation into the classification spec via deep merge
	newsFeedSpec := map[string]interface{}{
		"content_features": map[string]interface{}{
			"news_feed": map[string]interface{}{
				"recommended":       config.Recommended,
				"reason":            config.Reason,
				"vertical_keywords": config.VerticalKeywords,
				"source_types":      config.SourceTypes,
				"separate_page":     config.SeparatePage,
			},
		},
	}

	// Deep merge into existing classification spec
	mergedSpec := deepMergeNewsFeed(specData, newsFeedSpec)
	mergedJSON, err := json.Marshal(mergedSpec)
	if err != nil {
		return nil, fmt.Errorf("marshal merged spec: %w", err)
	}

	// Write via the standard site_specs versioning pattern
	// (mark old as superseded, insert new as current)
	tx, err := params.DB.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(ctx, `
		UPDATE site_specs
		SET is_current = false, superseded_at = NOW()
		WHERE site_id = $1 AND aspect = 'classification' AND is_current = true
	`, siteID)
	if err != nil {
		return nil, fmt.Errorf("supersede old spec: %w", err)
	}

	_, err = tx.ExecContext(ctx, `
		INSERT INTO site_specs (site_id, aspect, data, source, source_agent, is_current)
		VALUES ($1, 'classification', $2::jsonb, 'enrichment', 'evaluate_news_feed', true)
	`, siteID, string(mergedJSON))
	if err != nil {
		return nil, fmt.Errorf("insert enriched spec: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}

	logger.Info("EvaluateNewsFeedAction: wrote news feed recommendation",
		zap.String("site_id", siteID.String()),
		zap.Bool("recommended", config.Recommended),
		zap.Bool("separate_page", config.SeparatePage),
		zap.String("reason", config.Reason))

	return map[string]interface{}{
		"recommended":       config.Recommended,
		"reason":            config.Reason,
		"vertical_keywords": config.VerticalKeywords,
		"source_types":      config.SourceTypes,
		"separate_page":     config.SeparatePage,
		"industry":          industry,
		"site_type":         siteType,
	}, nil
}

// matchVerticalNews tries multiple signals to find a news config.
// Checks: industry, site_type, category, domain parts.
func matchVerticalNews(industry, siteType, category, domain string, logger *zap.Logger) *verticalNewsConfig {
	// Normalise for matching
	signals := []string{
		strings.ToLower(industry),
		strings.ToLower(siteType),
		strings.ToLower(category),
	}

	// Add domain-derived signals (e.g. "gaswholesalers" → "gas")
	domainLower := strings.ToLower(domain)
	for keyword := range verticalNewsMap {
		if strings.Contains(domainLower, keyword) {
			signals = append(signals, keyword)
		}
	}

	// Try each signal against the map
	for _, signal := range signals {
		if signal == "" {
			continue
		}
		if config, ok := verticalNewsMap[signal]; ok {
			logger.Info("matchVerticalNews: matched",
				zap.String("signal", signal),
				zap.Bool("recommended", config.Recommended))
			return &config
		}
		// Try partial matches (e.g. "gas wholesale" contains "gas")
		for key, config := range verticalNewsMap {
			if strings.Contains(signal, key) {
				logger.Info("matchVerticalNews: partial match",
					zap.String("signal", signal),
					zap.String("matched_key", key),
					zap.Bool("recommended", config.Recommended))
				return &config
			}
		}
	}

	return nil
}

// deepMergeNewsFeed merges the news_feed recommendation into the spec.
// Only adds/updates content_features.news_feed — doesn't touch other fields.
func deepMergeNewsFeed(existing, update map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	for k, v := range existing {
		result[k] = v
	}
	for k, v := range update {
		if existingMap, ok := result[k].(map[string]interface{}); ok {
			if updateMap, ok := v.(map[string]interface{}); ok {
				result[k] = deepMergeNewsFeed(existingMap, updateMap)
				continue
			}
		}
		result[k] = v
	}
	return result
}
