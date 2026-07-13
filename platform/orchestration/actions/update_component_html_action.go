// FILE: platform/orchestration/actions/update_component_html_action.go
//
// UpdateComponentHTMLAction updates the html_template of a content_component.
// Used by tool-improver and other agents that modify component HTML.
//
// Optionally snapshots the current HTML before overwriting (create_version = true).
// Also marks associated page_components as build_status='pending' so the
// rerender pipeline picks them up.
//
// Workflow config example:
//
//	"update_component": {
//	    "action": "update_component_html",
//	    "config": {
//	        "component_id": "tool_data.component_id",
//	        "html_field":   "improved_html.result",
//	        "create_version": true,
//	        "version_note":   "Pre-improvement snapshot"
//	    },
//	    "output_field": "update_result",
//	    "next_step": "create_rerender_item"
//	}
//
// Registration (add to registry.go):
//
//	"update_component_html": { Handler: UpdateComponentHTMLAction, Category: "site", IsLocal: true },

package actions

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"go.uber.org/zap"
)

// ============================================================================
// ActionInputSpec
// ============================================================================

var UpdateComponentHTMLInputSpec = datahelpers.ActionInputSpec{
	Required:   []string{"component_id", "html_field"},
	Optional:   []string{},
	Defaults:   map[string]interface{}{},
	Deprecated: map[string]string{},
}

func init() {
	datahelpers.RegisterActionInputSpec("update_component_html", UpdateComponentHTMLInputSpec)
}

// ============================================================================
// ACTION: update_component_html
// ============================================================================

func UpdateComponentHTMLAction(ctx context.Context, params ActionParams) (interface{}, error) {
	logger := params.Logger.With(
		zap.String("action", "update_component_html"),
		zap.String("step_name", params.ExecutionContext.StepName),
	)

	logger.Info("UpdateComponentHTMLAction: Starting")

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
		UpdateComponentHTMLInputSpec,
		logger,
	)
	if err != nil {
		return nil, fmt.Errorf("input extraction failed: %w", err)
	}

	componentIDStr := inputs.Get("component_id")
	componentID, err := uuid.Parse(componentIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid component_id %q: %w", componentIDStr, err)
	}

	newHTML := inputs.Get("html_field")
	if newHTML == "" {
		return nil, fmt.Errorf("html_field resolved to empty string")
	}

	// Config literals
	config := params.StepConfig.Config
	createVersion := false
	if v, ok := config["create_version"].(bool); ok {
		createVersion = v
	}
	versionNote := "Pre-edit snapshot"
	if v, ok := config["version_note"].(string); ok && v != "" {
		versionNote = v
	}

	// --- 1. Load current component to verify it exists ---
	var currentHTML string
	var componentFunction string
	var componentName string
	err = params.DB.QueryRowContext(ctx, `
		SELECT html_template, function, name
		FROM content_components
		WHERE id = $1 AND is_active = true
	`, componentID).Scan(&currentHTML, &componentFunction, &componentName)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("component %s not found or inactive", componentIDStr)
		}
		return nil, fmt.Errorf("load component: %w", err)
	}

	logger.Info("UpdateComponentHTMLAction: Component loaded",
		zap.String("component_id", componentIDStr),
		zap.String("function", componentFunction),
		zap.Int("current_html_len", len(currentHTML)),
		zap.Int("new_html_len", len(newHTML)),
	)

	// --- 2. Optionally snapshot current HTML to component_versions ---
	// Best-effort — a snapshot failure is logged but doesn't block the UPDATE.
	// Uses live schema columns: version_number (computed as MAX+1),
	// change_description, changed_by. The older code targeted a column
	// named version_note which no longer exists — that INSERT was silently
	// failing and the snapshot never actually landed.
	versioned := false
	if createVersion && currentHTML != "" {
		var nextVersionNumber int
		if vErr := params.DB.QueryRowContext(ctx, `
			SELECT COALESCE(MAX(version_number), 0) + 1
			FROM component_versions
			WHERE component_id = $1::uuid
		`, componentID).Scan(&nextVersionNumber); vErr != nil {
			logger.Warn("UpdateComponentHTMLAction: could not compute next version_number, defaulting to 1",
				zap.Error(vErr))
			nextVersionNumber = 1
		}

		_, err = params.DB.ExecContext(ctx, `
			INSERT INTO component_versions (
				component_id, version_number,
				html_template, change_description, changed_by,
				created_at
			) VALUES ($1::uuid, $2, $3, $4, $5, NOW())
		`, componentID, nextVersionNumber, currentHTML, versionNote, "update_component_html")
		if err != nil {
			// Log but don't fail — versioning is best-effort
			logger.Warn("UpdateComponentHTMLAction: Failed to create version snapshot",
				zap.Error(err),
				zap.String("component_id", componentIDStr),
			)
		} else {
			versioned = true
			logger.Info("UpdateComponentHTMLAction: Version snapshot created",
				zap.String("component_id", componentIDStr),
				zap.Int("version_number", nextVersionNumber),
			)
		}
	}

	// --- 3. Update the component HTML ---
	result, err := params.DB.ExecContext(ctx, `
		UPDATE content_components
		SET html_template = $1, updated_at = NOW()
		WHERE id = $2 AND is_active = true
	`, newHTML, componentID)
	if err != nil {
		return nil, fmt.Errorf("update html_template: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return nil, fmt.Errorf("no rows updated for component %s", componentIDStr)
	}

	// --- 4. Mark associated page_components as pending rebuild ---
	// Do NOT set rendered_html here. The template lives on the component;
	// rendered_html on page_components is the per-page render after
	// variable substitution and cannot be recreated by copying the new
	// template verbatim. The rerender pipeline regenerates rendered_html
	// correctly; we just flag that it needs to run.
	rebuildResult, err := params.DB.ExecContext(ctx, `
		UPDATE page_components
		SET build_status = 'pending', updated_at = NOW()
		WHERE component_id = $1
	`, componentID)
	if err != nil {
		logger.Warn("UpdateComponentHTMLAction: Failed to update page_components",
			zap.Error(err),
		)
	}
	var pagesMarked int64
	if rebuildResult != nil {
		pagesMarked, _ = rebuildResult.RowsAffected()
	}

	logger.Info("UpdateComponentHTMLAction: Complete",
		zap.String("component_id", componentIDStr),
		zap.String("function", componentFunction),
		zap.Bool("versioned", versioned),
		zap.Int64("pages_marked_pending", pagesMarked),
	)

	return map[string]interface{}{
		"updated":              true,
		"component_id":         componentIDStr,
		"function":             componentFunction,
		"name":                 componentName,
		"versioned":            versioned,
		"pages_marked_pending": pagesMarked,
		"html_length":          len(newHTML),
	}, nil
}
