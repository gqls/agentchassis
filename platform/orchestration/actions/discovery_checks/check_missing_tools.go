package discovery_checks

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"go.uber.org/zap"
)

func init() { Register(&MissingToolsCheck{}) }

type MissingToolsCheck struct{}

func (c *MissingToolsCheck) Name() string { return "missing_tools" }

func (c *MissingToolsCheck) Run(dctx DiscoveryCheckContext) (*CheckResult, error) {
	suggestions, err := findMissingTools(dctx)
	if err != nil {
		return nil, err
	}
	if len(suggestions) == 0 {
		return &CheckResult{}, nil
	}

	result := &CheckResult{
		Findings: []map[string]interface{}{{
			"check":       "missing_tools",
			"tools_found": len(suggestions),
			"suggestions": suggestions,
		}},
	}

	for _, tool := range suggestions {
		specJSON, _ := json.Marshal(map[string]interface{}{
			"check":             "missing_tools",
			"tool_component_id": tool.ComponentID,
			"tool_function":     tool.Function,
			"tool_display_name": tool.DisplayName,
			"match_reason":      tool.MatchReason,
		})

		result.WorkItems = append(result.WorkItems, WorkItemSpec{
			SiteID:       dctx.SiteID,
			Source:       "discovery",
			Pipeline:     "build",
			ItemType:     "add_tool",
			Severity:     "low",
			Summary:      fmt.Sprintf("Suggested tool: %s", tool.DisplayName),
			SpecJSON:     string(specJSON),
			Priority:     120,
			HandlerAgent: "tool-deployer",
			Status:       "detected",
			CreatedBy:    dctx.AgentType,
			ItemKey:      fmt.Sprintf("add_tool:%s", tool.Function),
			BatchID:      dctx.BatchID,
		})
	}

	return result, nil
}

type toolSuggestion struct {
	ComponentID string `json:"component_id"`
	Function    string `json:"function"`
	DisplayName string `json:"display_name"`
	Category    string `json:"category"`
	MatchReason string `json:"match_reason"`
}

func findMissingTools(dctx DiscoveryCheckContext) ([]toolSuggestion, error) {
	// Step 1: Get site's industry and category for matching
	var siteType sql.NullString
	var industryTags sql.NullString

	// Try site_specs first (most detailed)
	_ = dctx.DB.QueryRowContext(dctx.Ctx, `
		SELECT
			data->>'industry' as industry,
			data->>'site_type' as site_type
		FROM site_specs
		WHERE site_id = $1 AND aspect = 'classification' AND is_current = true
	`, dctx.SiteID).Scan(&industryTags, &siteType)

	// Fallback to sites table
	if !siteType.Valid || siteType.String == "" {
		_ = dctx.DB.QueryRowContext(dctx.Ctx, `
			SELECT site_type FROM sites WHERE id = $1
		`, dctx.SiteID).Scan(&siteType)
	}

	dctx.Logger.Info("findMissingTools: Site context",
		zap.String("site_id", dctx.SiteID.String()),
		zap.String("site_type", siteType.String),
		zap.String("industry", industryTags.String),
	)

	// Step 2: Find library tools not already deployed to this site
	rows, err := dctx.DB.QueryContext(dctx.Ctx, `
		SELECT
			lib.id::text,
			lib.function,
			lib.display_name,
			lib.category,
			lib.semantic_tags::text
		FROM content_components lib
		WHERE lib.component_level = 'tool'
		  AND lib.forked_from IS NULL
		  AND lib.is_active = true
		  AND lib.html_template != ''
		  AND NOT EXISTS (
		    SELECT 1
		    FROM content_components fork
		    JOIN page_components pc ON pc.component_id = fork.id
		    JOIN pages p ON pc.page_id = p.id
		    WHERE fork.forked_from = lib.id
		      AND p.site_id = $1
		      AND fork.is_active = true
		  )
		ORDER BY lib.display_name
	`, dctx.SiteID)
	if err != nil {
		return nil, fmt.Errorf("missing_tools query failed: %w", err)
	}
	defer rows.Close()

	var suggestions []toolSuggestion
	for rows.Next() {
		var compID, function, displayName, tagsStr string
		var category sql.NullString
		if err := rows.Scan(&compID, &function, &displayName, &category, &tagsStr); err != nil {
			dctx.Logger.Warn("Failed to scan tool", zap.Error(err))
			continue
		}

		// Parse semantic_tags (stored as JSON array or comma-separated)
		var tags []string
		if err := json.Unmarshal([]byte(tagsStr), &tags); err != nil {
			tags = strings.Split(tagsStr, ",")
		}

		reason := matchToolToSite(tags, siteType.String, industryTags.String)
		if reason == "" {
			continue
		}

		suggestions = append(suggestions, toolSuggestion{
			ComponentID: compID,
			Function:    function,
			DisplayName: displayName,
			Category:    category.String,
			MatchReason: reason,
		})
	}

	if len(suggestions) > 0 {
		dctx.Logger.Info("findMissingTools: Found suggestions",
			zap.String("site_id", dctx.SiteID.String()),
			zap.Int("count", len(suggestions)),
		)
	}

	return suggestions, nil
}

// matchToolToSite checks if a tool is relevant to a site based on
// the tool's semantic_tags and the site's type/industry.
// Returns a match reason string, or "" if not relevant.
func matchToolToSite(toolTags []string, siteType, industry string) string {
	// Universal tools — relevant to all sites
	universalTags := map[string]bool{
		"security": true, "password": true, "privacy": true,
	}

	// Site-type affinity
	siteTypeAffinity := map[string][]string{
		"tools":     {"calculator", "generator", "converter", "analyzer", "webdev"},
		"content":   {"calculator", "generator", "social-media"},
		"ecommerce": {"calculator", "statistics", "reviews", "ecommerce"},
		"corporate": {"calculator", "statistics", "marketing"},
		"landing":   {"calculator", "marketing"},
		"portfolio": {"design", "generator", "creative"},
	}

	// Industry affinity
	industryAffinity := map[string][]string{
		"marketing":   {"ab-testing", "statistics", "marketing", "conversion"},
		"technology":  {"webdev", "css", "design", "generator", "ai"},
		"design":      {"design", "css", "generator", "creative", "visual"},
		"photography": {"image", "photo-editing", "background-removal", "creative"},
		"ecommerce":   {"ecommerce", "reviews", "ranking", "statistics"},
		"finance":     {"calculator", "statistics"},
		"education":   {"calculator", "statistics", "generator"},
	}

	// Check universal match
	for _, tag := range toolTags {
		if universalTags[tag] {
			return "universal: " + tag
		}
	}

	// Check site type affinity
	if affinityTags, ok := siteTypeAffinity[siteType]; ok {
		for _, tag := range toolTags {
			for _, aff := range affinityTags {
				if tag == aff {
					return "site_type:" + siteType + " matched:" + tag
				}
			}
		}
	}

	// Check industry affinity
	industryLower := strings.ToLower(industry)
	for ind, affinityTags := range industryAffinity {
		if strings.Contains(industryLower, ind) {
			for _, tag := range toolTags {
				for _, aff := range affinityTags {
					if tag == aff {
						return "industry:" + ind + " matched:" + tag
					}
				}
			}
		}
	}

	return ""
}
