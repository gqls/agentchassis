// FILE: platform/orchestration/actions/create_tool_cross_link_items.go
//
// Creates content_rewrite work items so that existing content pages
// reference newly created tools contextually. Called by tool-suggester
// after tools have been created, using the related_pages field from
// each suggestion.
//
// Example: tool-suggester suggests "Fuel Cost Estimator" with
// related_pages: ["services", "fleet-fuel-services"]. This action
// creates content_rewrite items for both pages, each telling
// page-build-handler to weave in a contextual reference to the tool.
//
// Registration:
//
//	"create_tool_cross_link_items": {
//	    Handler:     CreateToolCrossLinkItemsAction,
//	    Category:    "site",
//	    Description: "Create content_rewrite items to cross-link tools from related pages",
//	    IsLocal:     true,
//	},

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

// ============================================================================
// ActionInputSpec
// ============================================================================

var CreateToolCrossLinkItemsInputSpec = datahelpers.ActionInputSpec{
	Required:   []string{"site_id"},
	Optional:   []string{"suggestions"},
	Defaults:   map[string]interface{}{},
	Deprecated: map[string]string{},
}

func init() {
	datahelpers.RegisterActionInputSpec("create_tool_cross_link_items", CreateToolCrossLinkItemsInputSpec)
}

// ============================================================================
// ACTION: create_tool_cross_link_items
// ============================================================================

func CreateToolCrossLinkItemsAction(ctx context.Context, params ActionParams) (interface{}, error) {
	logger := params.Logger.With(
		zap.String("action", "create_tool_cross_link_items"),
	)

	logger.Info("CreateToolCrossLinkItemsAction: Starting")

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}
	if params.DB == nil {
		return nil, fmt.Errorf("database connection required")
	}

	// --- Resolve inputs ---
	inputs, err := datahelpers.ExtractActionInputs(
		params.CollectedData,
		params.StepConfig.Config,
		CreateToolCrossLinkItemsInputSpec,
		logger,
	)
	if err != nil {
		return nil, fmt.Errorf("input extraction failed: %w", err)
	}

	siteIDStr := inputs.Get("site_id")
	siteID, err := uuid.Parse(siteIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid site_id: %w", err)
	}

	// --- Extract suggestions from collected data ---
	// The suggestions come from evaluation.result.suggestions (the same array
	// that create_items_loop iterates over)
	suggestionsRaw := datahelpers.ExtractNestedField(params.CollectedData, "evaluation.result.suggestions")
	if suggestionsRaw == nil {
		logger.Info("CreateToolCrossLinkItemsAction: No suggestions found, skipping")
		return map[string]interface{}{"items_created": 0, "reason": "no suggestions"}, nil
	}

	suggestions, ok := suggestionsRaw.([]interface{})
	if !ok {
		logger.Info("CreateToolCrossLinkItemsAction: Suggestions not an array, skipping",
			zap.String("type", fmt.Sprintf("%T", suggestionsRaw)))
		return map[string]interface{}{"items_created": 0, "reason": "suggestions not array"}, nil
	}

	// --- Load page map for page_id resolution ---
	pageMap := make(map[string]uuid.UUID) // page name → page ID
	rows, err := params.DB.QueryContext(ctx, `
		SELECT name, id FROM pages
		WHERE site_id = $1 AND status = 'active'
	`, siteID)
	if err != nil {
		logger.Warn("CreateToolCrossLinkItemsAction: Failed to load pages", zap.Error(err))
	} else {
		defer rows.Close()
		for rows.Next() {
			var name string
			var id uuid.UUID
			if rows.Scan(&name, &id) == nil {
				pageMap[name] = id
			}
		}
	}

	// --- Process each suggestion's related_pages ---
	itemsCreated := 0
	for _, sugRaw := range suggestions {
		sug, ok := sugRaw.(map[string]interface{})
		if !ok {
			continue
		}

		toolName, _ := sug["name"].(string)
		toolFunction, _ := sug["function"].(string)
		toolDescription, _ := sug["description"].(string)

		if toolName == "" || toolFunction == "" {
			continue
		}

		// Extract related_pages array
		relatedPagesRaw, ok := sug["related_pages"].([]interface{})
		if !ok || len(relatedPagesRaw) == 0 {
			continue
		}

		toolPageURL := fmt.Sprintf("/tools/%s.html", toolFunction)

		for _, pageRaw := range relatedPagesRaw {
			pageName, ok := pageRaw.(string)
			if !ok || pageName == "" {
				continue
			}

			// Skip tool pages — cross-linking tool-to-tool isn't useful
			if strings.HasPrefix(pageName, "tool-") {
				logger.Info("CreateToolCrossLinkItemsAction: Skipping tool page",
					zap.String("page_name", pageName),
					zap.String("tool", toolFunction))
				continue
			}

			// Resolve page_id
			pageID, exists := pageMap[pageName]
			if !exists {
				logger.Info("CreateToolCrossLinkItemsAction: Page not found, skipping",
					zap.String("page_name", pageName),
					zap.String("tool", toolFunction))
				continue
			}

			// Build the content_rewrite spec
			spec, _ := json.Marshal(map[string]interface{}{
				"page":        pageName,
				"page_name":   pageName, // dispatch loop maps spec.page_name → input_data.page_name
				"page_id":     pageID.String(),
				"description": fmt.Sprintf("Add a contextual reference to the %s tool on this page where it naturally fits the topic.", toolName),
				"suggestion": fmt.Sprintf(
					"Weave a natural reference to '%s' (%s) into the existing content. "+
						"Link to %s. The reference should feel like a helpful suggestion, not an advertisement — "+
						"e.g. 'Use our %s to see how this applies to your situation' or "+
						"'Try our %s to estimate your costs'. "+
						"Place it where it's most contextually relevant to what the page is discussing. "+
						"Do NOT add a new section — integrate it into existing paragraphs or after a relevant point.",
					toolName, toolDescription, toolPageURL,
					strings.ToLower(toolName), strings.ToLower(toolName),
				),
				"acceptance_test": fmt.Sprintf(
					"Page contains at least one inline link to %s with anchor text referencing the tool",
					toolPageURL,
				),
				"source":            "tool-suggester",
				"tool_function":     toolFunction,
				"tool_display_name": toolName,
				"work_item_type":    "content_rewrite",
				"max_fix_attempts":  1,
			})

			itemKey := fmt.Sprintf("tool_crosslink:%s:%s:%s", toolFunction, pageName, siteID)

			// The ON CONFLICT WHERE clause must imply idx_swi_dedup's predicate
			// (see workItemTerminalStatuses) — a hardcoded stale list here
			// previously made this insert fail 42P10 and only Warn-log.
			_, err := params.DB.ExecContext(ctx, fmt.Sprintf(`
				INSERT INTO site_work_items (
					site_id, source, pipeline, item_type, severity, summary,
					spec, page_id, priority, handler_agent, status, created_by, item_key
				) VALUES (
					$1, 'tool-suggester', 'build', 'content_rewrite', 'low',
					$2, $3::jsonb, $4, 110, 'page-build-handler', 'triaged', 'tool-suggester', $5
				) ON CONFLICT (site_id, item_key)
				  WHERE item_key IS NOT NULL
				  AND status NOT IN (%s)
				  DO NOTHING
			`, sqlInList(workItemTerminalStatuses)), siteID,
				fmt.Sprintf("Add %s tool reference to %s page", toolName, pageName),
				string(spec), pageID, itemKey,
			)
			if err != nil {
				logger.Warn("CreateToolCrossLinkItemsAction: Failed to create cross-link item",
					zap.String("tool", toolFunction),
					zap.String("page", pageName),
					zap.Error(err))
				continue
			}

			itemsCreated++
			logger.Info("CreateToolCrossLinkItemsAction: Created cross-link item",
				zap.String("tool", toolFunction),
				zap.String("page", pageName),
				zap.String("page_id", pageID.String()))
		}
	}

	logger.Info("CreateToolCrossLinkItemsAction: Complete",
		zap.Int("items_created", itemsCreated),
		zap.Int("suggestions_processed", len(suggestions)))

	return map[string]interface{}{
		"items_created": itemsCreated,
	}, nil
}
