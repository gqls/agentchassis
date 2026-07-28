// FILE: platform/orchestration/actions/load_existing_content_action.go
//
// LoadExistingContentAction checks if the work item has mode: "recreate"
// and if so, loads existing page content from research_results.
// This is used by the page-build-handler before spawning the content writer.
//
// For non-adoption pages (no mode: "recreate"), returns empty — no-op.
// For adopted pages, returns the raw markdown and structured content
// from the adoption crawl, which the content writer can use as source material.
//
// Registration:
//   "load_existing_content": {
//       Handler:     LoadExistingContentAction,
//       Category:    "site",
//       Description: "Load existing page content from research_results for recreate mode",
//       IsLocal:     true,
//   },

package actions

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"go.uber.org/zap"
)

var LoadExistingContentInputSpec = datahelpers.ActionInputSpec{
	CheckConfig: true,
	Required:    []string{"site_id", "page_name"},
	Optional:    []string{"page_id", "mode"},
	Defaults:    map[string]interface{}{},
	Deprecated:  map[string]string{},
}

func init() {
	datahelpers.RegisterActionInputSpec("load_existing_content", LoadExistingContentInputSpec)
}

func LoadExistingContentAction(ctx context.Context, params ActionParams) (interface{}, error) {
	logger := params.Logger.With(zap.String("action", "load_existing_content"))

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}
	if params.DB == nil {
		return map[string]interface{}{"has_existing": false, "reason": "no_db"}, nil
	}

	inputs, err := datahelpers.ExtractActionInputs(
		params.CollectedData, params.StepConfig.Config,
		LoadExistingContentInputSpec, logger,
	)
	if err != nil {
		// Non-fatal — just means no existing content
		logger.Info("LoadExistingContent: input extraction failed, skipping",
			zap.Error(err))
		return map[string]interface{}{"has_existing": false, "reason": "no_inputs"}, nil
	}

	// Check if this is a recreate mode page
	mode := inputs.Get("mode")

	if mode != "recreate" {
		logger.Info("LoadExistingContent: not recreate mode, skipping",
			zap.String("mode", mode))
		return map[string]interface{}{"has_existing": false, "reason": "not_recreate"}, nil
	}

	siteIDStr := inputs.Get("site_id")
	siteID, err := uuid.Parse(siteIDStr)
	if err != nil {
		return map[string]interface{}{"has_existing": false, "reason": "invalid_site_id"}, nil
	}

	pageName := inputs.Get("page_name")

	// Try to find adoption page content in research_results
	// First try by page_id if available
	pageIDStr := inputs.Get("page_id")

	var dataJSON []byte
	var found bool

	if pageIDStr != "" {
		pageID, err := uuid.Parse(pageIDStr)
		if err == nil {
			err = params.DB.QueryRowContext(ctx, `
				SELECT data FROM research_results
				WHERE site_id = $1 AND page_id = $2
				  AND result_type = 'adoption_page'
				ORDER BY created_at DESC LIMIT 1
			`, siteID, pageID).Scan(&dataJSON)
			if err == nil {
				found = true
			}
		}
	}

	// Fallback: search by page_name in the data field
	if !found {
		err = params.DB.QueryRowContext(ctx, `
			SELECT data FROM research_results
			WHERE site_id = $1
			  AND result_type = 'adoption_page'
			  AND data->>'page_name' = $2
			ORDER BY created_at DESC LIMIT 1
		`, siteID, pageName).Scan(&dataJSON)
		if err == nil {
			found = true
		}
	}

	if !found {
		logger.Info("LoadExistingContent: no adoption content found",
			zap.String("site_id", siteIDStr),
			zap.String("page_name", pageName))
		return map[string]interface{}{
			"has_existing": false,
			"reason":       "no_adoption_content",
			"mode":         "recreate",
		}, nil
	}

	var contentData map[string]interface{}
	if err := json.Unmarshal(dataJSON, &contentData); err != nil {
		logger.Warn("LoadExistingContent: failed to parse stored content",
			zap.Error(err))
		return map[string]interface{}{"has_existing": false, "reason": "parse_error"}, nil
	}

	// Extract the existing content and raw markdown
	existingContent, _ := contentData["existing_content"]
	rawMarkdown := ""
	if ec, ok := existingContent.(map[string]interface{}); ok {
		rawMarkdown, _ = ec["raw_markdown"].(string)
	}

	logger.Info("LoadExistingContent: found adoption content",
		zap.String("page_name", pageName),
		zap.Int("markdown_length", len(rawMarkdown)),
		zap.Bool("has_structured_content", existingContent != nil))

	return map[string]interface{}{
		"has_existing":     true,
		"mode":             "recreate",
		"existing_content": existingContent,
		"raw_markdown":     rawMarkdown,
		"adopted_from":     contentData["adopted_from"],
	}, nil
}
