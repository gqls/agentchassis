// FILE: platform/orchestration/actions/update_site_spec_from_item_action.go
//
// Handler action for needs_spec_update work items. Reads the work item's
// spec field which contains {aspect, field, suggested_value}, loads the
// current spec for that aspect, merges the field, writes a new version.
//
// This is the mechanical counterpart to WriteSiteSpecAction — that action
// is used in workflows where the caller provides the data. This action
// reads the data from the dispatched work item.
//
// Uses the same deep-merge and versioning pattern as WriteSiteSpecAction.
//
// Registration:
//   "update_site_spec_from_item": {
//       Handler:     UpdateSiteSpecFromItemAction,
//       Category:    "site",
//       Description: "Apply a spec update from a work item (needs_spec_update handler)",
//       IsLocal:     true,
//   },

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

var UpdateSiteSpecFromItemInputSpec = datahelpers.ActionInputSpec{
	CheckConfig: true,
	Required:    []string{"site_id"},
	Optional:    []string{"spec", "work_item_id"},
	Defaults:    map[string]interface{}{},
	Deprecated:  map[string]string{},
}

func init() {
	datahelpers.RegisterActionInputSpec("update_site_spec_from_item", UpdateSiteSpecFromItemInputSpec)
}

func UpdateSiteSpecFromItemAction(ctx context.Context, params ActionParams) (interface{}, error) {
	logger := params.Logger.With(zap.String("action", "update_site_spec_from_item"))

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}
	if params.DB == nil {
		return nil, fmt.Errorf("database connection required")
	}

	inputs, err := datahelpers.ExtractActionInputs(
		params.CollectedData, params.StepConfig.Config,
		UpdateSiteSpecFromItemInputSpec, logger,
	)
	if err != nil {
		return nil, fmt.Errorf("input extraction failed: %w", err)
	}

	siteIDStr := inputs.Get("site_id")
	siteID, err := uuid.Parse(siteIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid site_id: %w", err)
	}

	// Extract the update details from the work item spec
	// Expected shape: {aspect, field, suggested_value, description, ...}
	specRaw := inputs.GetRaw("spec")
	if specRaw == nil {
		// Try input_data.spec
		specRaw = datahelpers.ExtractNestedField(params.CollectedData, "input_data.spec")
	}

	specMap, ok := specRaw.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("spec must be a JSON object, got %T", specRaw)
	}

	// Extract the update parameters
	aspect, _ := specMap["aspect"].(string)
	if aspect == "" {
		// Try to infer from category
		category, _ := specMap["category"].(string)
		switch category {
		case "identity", "metadata":
			aspect = "identity"
		case "config":
			aspect = "site_config"
		default:
			aspect = "identity" // safe default — most spec updates are identity fields
		}
	}

	field, _ := specMap["field"].(string)
	suggestedValue := specMap["suggested_value"]

	// If field + suggested_value are set, do a targeted field update
	// If not, look for a broader data merge (e.g. the whole spec is a patch)
	if field != "" && suggestedValue != nil {
		return applyFieldUpdate(ctx, params.DB, siteID, aspect, field, suggestedValue, logger)
	}

	// Check for description-only items (audit found missing metadata but
	// didn't specify what value to set). These need human review.
	description, _ := specMap["description"].(string)
	if description != "" && field == "" {
		logger.Info("UpdateSiteSpecFromItemAction: no field/value specified, cannot auto-update",
			zap.String("description", description),
			zap.String("aspect", aspect))
		return map[string]interface{}{
			"updated":     false,
			"reason":      "no field or value specified — needs human review",
			"aspect":      aspect,
			"description": description,
		}, nil
	}

	return map[string]interface{}{
		"updated": false,
		"reason":  "insufficient data for spec update",
	}, nil
}

func applyFieldUpdate(ctx context.Context, db *sql.DB, siteID uuid.UUID, aspect, field string, value interface{}, logger *zap.Logger) (interface{}, error) {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	// Read current spec
	var currentDataJSON []byte
	var oldID *uuid.UUID
	err = tx.QueryRowContext(ctx, `
		SELECT id, data FROM site_specs
		WHERE site_id = $1 AND aspect = $2 AND is_current = true
	`, siteID, aspect).Scan(&oldID, &currentDataJSON)

	var currentData map[string]interface{}
	if err == sql.ErrNoRows {
		currentData = make(map[string]interface{})
	} else if err != nil {
		return nil, fmt.Errorf("read current spec: %w", err)
	} else {
		if err := json.Unmarshal(currentDataJSON, &currentData); err != nil {
			return nil, fmt.Errorf("unmarshal current spec: %w", err)
		}
	}

	// Check if value is actually different
	existingValue, exists := currentData[field]
	if exists {
		existingJSON, _ := json.Marshal(existingValue)
		newJSON, _ := json.Marshal(value)
		if string(existingJSON) == string(newJSON) {
			logger.Info("UpdateSiteSpecFromItemAction: value unchanged, skipping",
				zap.String("aspect", aspect),
				zap.String("field", field))
			return map[string]interface{}{
				"updated": false,
				"reason":  "value already matches",
				"aspect":  aspect,
				"field":   field,
			}, nil
		}
	}

	// Apply the update
	currentData[field] = value
	mergedJSON, err := json.Marshal(currentData)
	if err != nil {
		return nil, fmt.Errorf("marshal merged spec: %w", err)
	}

	// Supersede old row
	if oldID != nil {
		_, err = tx.ExecContext(ctx, `
			UPDATE site_specs
			SET is_current = false, superseded_at = now()
			WHERE id = $1
		`, *oldID)
		if err != nil {
			return nil, fmt.Errorf("supersede old spec: %w", err)
		}
	}

	// Insert new current row
	var newID uuid.UUID
	err = tx.QueryRowContext(ctx, `
		INSERT INTO site_specs (
			site_id, aspect, data, source, source_agent,
			is_current, created_by, notes
		) VALUES (
			$1, $2, $3::jsonb, 'spec-updater', 'spec-updater',
			true, 'spec-updater', $4
		) RETURNING id
	`, siteID, aspect, string(mergedJSON),
		fmt.Sprintf("Updated field '%s' via spec-updater", field),
	).Scan(&newID)

	if err != nil {
		return nil, fmt.Errorf("insert new spec: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}

	logger.Info("UpdateSiteSpecFromItemAction: spec updated",
		zap.String("site_id", siteID.String()),
		zap.String("aspect", aspect),
		zap.String("field", field),
		zap.String("spec_id", newID.String()),
		zap.Bool("had_previous", oldID != nil))

	return map[string]interface{}{
		"updated":      true,
		"spec_id":      newID.String(),
		"aspect":       aspect,
		"field":        field,
		"had_previous": oldID != nil,
	}, nil
}
