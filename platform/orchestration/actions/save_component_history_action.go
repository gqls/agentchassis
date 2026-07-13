// FILE: platform/orchestration/actions/save_component_history_action.go
//
// Saves the current content_data of a page_component to page_component_history
// before it gets overwritten. Called as a workflow step before content edits.
//
// Designed to be used in content-writing and section-editing workflows:
//
//   "save_history": {
//       "action": "save_component_history",
//       "config": {
//           "component_instance_id": "edit_context.page_component.id",
//           "source": "section-editor",
//           "source_item_id": "current_item.id"
//       },
//       "output_field": "history_saved",
//       "next_step": "apply_edit"
//   }
//
// If the component has no content_data (NULL or empty), no history row is
// created and the action succeeds with saved=false.
//
// Registration (add to registry.go):
//   "save_component_history": { Handler: SaveComponentHistoryAction, IsLocal: true },

package actions

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"go.uber.org/zap"
)

var SaveComponentHistoryInputSpec = datahelpers.ActionInputSpec{
	Required:   []string{"component_instance_id"},
	Optional:   []string{},
	Defaults:   map[string]interface{}{},
	Deprecated: map[string]string{},
}

func init() {
	datahelpers.RegisterActionInputSpec("save_component_history", SaveComponentHistoryInputSpec)
}

func SaveComponentHistoryAction(ctx context.Context, params ActionParams) (interface{}, error) {
	logger := params.Logger.With(zap.String("action", "save_component_history"))

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}
	if params.DB == nil {
		return nil, fmt.Errorf("database connection required")
	}

	inputs, err := datahelpers.ExtractActionInputs(
		params.CollectedData,
		params.StepConfig.Config,
		SaveComponentHistoryInputSpec,
		logger,
	)
	if err != nil {
		return nil, fmt.Errorf("input extraction failed: %w", err)
	}

	pcIDStr := inputs.Get("component_instance_id")
	pcID, err := uuid.Parse(pcIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid component_instance_id %q: %w", pcIDStr, err)
	}

	// Config literals
	config := params.StepConfig.Config
	source, _ := config["source"].(string)
	if source == "" {
		source = "unknown"
	}

	var sourceItemID *uuid.UUID
	if itemIDStr, ok := config["source_item_id"].(string); ok && itemIDStr != "" {
		// source_item_id may be a path — try resolving from collected_data first
		resolved := datahelpers.ExtractNestedFieldString(params.CollectedData, itemIDStr)
		if resolved != "" {
			itemIDStr = resolved
		}
		parsed, err := uuid.Parse(itemIDStr)
		if err == nil {
			sourceItemID = &parsed
		}
	}

	// Read current state of the component
	var contentDataJSON []byte
	var pageID, siteID uuid.UUID

	err = params.DB.QueryRowContext(ctx, `
		SELECT pc.content_data, pc.page_id, p.site_id
		FROM page_components pc
		JOIN pages p ON pc.page_id = p.id
		WHERE pc.id = $1
	`, pcID).Scan(&contentDataJSON, &pageID, &siteID)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("page_component %s not found", pcID)
	}
	if err != nil {
		return nil, fmt.Errorf("query page_component: %w", err)
	}

	// If no content_data, nothing to save
	if len(contentDataJSON) == 0 || string(contentDataJSON) == "null" {
		logger.Info("SaveComponentHistoryAction: no content_data to save",
			zap.String("component_instance_id", pcID.String()),
		)
		return map[string]interface{}{
			"saved":                 false,
			"reason":                "no_content_data",
			"component_instance_id": pcID.String(),
		}, nil
	}

	// Verify it's valid JSON before saving
	var check interface{}
	if err := json.Unmarshal(contentDataJSON, &check); err != nil {
		logger.Warn("SaveComponentHistoryAction: content_data is not valid JSON, skipping",
			zap.String("component_instance_id", pcID.String()),
			zap.Error(err),
		)
		return map[string]interface{}{
			"saved":                 false,
			"reason":                "invalid_json",
			"component_instance_id": pcID.String(),
		}, nil
	}

	// Insert history row
	var historyID uuid.UUID
	err = params.DB.QueryRowContext(ctx, `
		INSERT INTO page_component_history (
			component_id, page_id, site_id, content_data,
			source, source_item_id
		) VALUES (
			$1, $2, $3, $4::jsonb,
			$5, $6
		) RETURNING id
	`, pcID, pageID, siteID, string(contentDataJSON),
		source, sourceItemID,
	).Scan(&historyID)
	if err != nil {
		return nil, fmt.Errorf("insert history: %w", err)
	}

	logger.Info("SaveComponentHistoryAction: history saved",
		zap.String("history_id", historyID.String()),
		zap.String("component_instance_id", pcID.String()),
		zap.String("page_id", pageID.String()),
		zap.String("source", source),
	)

	return map[string]interface{}{
		"saved":                 true,
		"history_id":            historyID.String(),
		"component_instance_id": pcID.String(),
		"page_id":               pageID.String(),
		"site_id":               siteID.String(),
	}, nil
}
