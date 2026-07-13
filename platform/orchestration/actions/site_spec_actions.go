// FILE: platform/orchestration/actions/site_spec_actions.go
//
// Actions for the site_specs table (versioned site specification storage).
//
// Actions:
//   WriteSiteSpecAction  — deep-merge partial update → insert new current row
//   ReadSiteSpecAction   — read one aspect or all current aspects for a site
//
// Registration (add to registry.go):
//   "write_site_spec": { Handler: WriteSiteSpecAction, IsLocal: true },
//   "read_site_spec":  { Handler: ReadSiteSpecAction,  IsLocal: true },

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

// ============================================================================
// ActionInputSpecs
// ============================================================================

// WriteSiteSpecInputSpec
//
// Workflow config example:
//
//	"write_identity_spec": {
//	    "action": "write_site_spec",
//	    "config": {
//	        "site_id": "site_record.site_id",
//	        "spec_data": "classifier_result.identity",
//	        "aspect": "identity",
//	        "source": "classifier",
//	        "source_agent": "domain-research-classifier"
//	    },
//	    "output_field": "identity_spec_written",
//	    "next_step": "..."
//	}
//
// site_id and spec_data are paths resolved from collected_data.
// aspect, source, source_agent, source_item_id, notes are config literals.
var WriteSiteSpecInputSpec = datahelpers.ActionInputSpec{
	Required: []string{"site_id", "spec_data"},
	Optional: []string{},
	Defaults: map[string]interface{}{},
	Deprecated: map[string]string{
		"site_id_field":   "site_id",
		"spec_data_field": "spec_data",
	},
}

// ReadSiteSpecInputSpec
//
// Workflow config example (single aspect):
//
//	"read_identity": {
//	    "action": "read_site_spec",
//	    "config": {
//	        "site_id": "site_record.site_id",
//	        "aspect": "identity"
//	    },
//	    "output_field": "identity_spec",
//	    "next_step": "..."
//	}
//
// Workflow config example (all current aspects):
//
//	"read_all_specs": {
//	    "action": "read_site_spec",
//	    "config": {
//	        "site_id": "site_record.site_id"
//	    },
//	    "output_field": "all_specs",
//	    "next_step": "..."
//	}
//
// site_id is a path resolved from collected_data.
// aspect is a config literal (omit to read all).
var ReadSiteSpecInputSpec = datahelpers.ActionInputSpec{
	Required:   []string{"site_id"},
	Optional:   []string{},
	Defaults:   map[string]interface{}{},
	Deprecated: map[string]string{},
}

func init() {
	datahelpers.RegisterActionInputSpec("write_site_spec", WriteSiteSpecInputSpec)
	datahelpers.RegisterActionInputSpec("read_site_spec", ReadSiteSpecInputSpec)
}

// ============================================================================
// ACTION: write_site_spec
//
// Deep-merges a partial update over the current spec for (site_id, aspect),
// then inserts a new row with the merged (complete) value. The old row is
// marked is_current = false, superseded_at = now().
//
// If no current row exists, the update becomes the initial complete value.
//
// Data inputs (via ActionInputSpec):
//   - site_id (required)   — path to site UUID in collected_data
//   - spec_data (required) — path to the JSONB data to write/merge
//
// Config literals (read directly):
//   - aspect (required)        — e.g. "identity", "tone", "strategy"
//   - source (required)        — e.g. "classifier", "hitl", "planner"
//   - source_agent (optional)  — agent type that wrote this
//   - source_item_id (optional)— work item UUID that caused this write
//   - notes (optional)         — human-readable reason
//   - created_by (optional)    — defaults to source_agent or source
// ============================================================================

func WriteSiteSpecAction(ctx context.Context, params ActionParams) (interface{}, error) {
	logger := params.Logger.With(zap.String("action", "write_site_spec"))

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}
	if params.DB == nil {
		return nil, fmt.Errorf("database connection required")
	}

	// Extract path-resolved inputs
	inputs, err := datahelpers.ExtractActionInputs(
		params.CollectedData,
		params.StepConfig.Config,
		WriteSiteSpecInputSpec,
		logger,
	)
	if err != nil {
		return nil, fmt.Errorf("input extraction failed: %w", err)
	}

	siteIDStr := inputs.Get("site_id")
	siteID, err := uuid.Parse(siteIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid site_id %q: %w", siteIDStr, err)
	}

	specData := inputs.GetRaw("spec_data")
	if specData == nil {
		return nil, fmt.Errorf("spec_data is nil")
	}

	// Coerce a string spec_data to an object.
	// JSON string  → parse it (LLM may emit spec as a JSON-encoded string)
	// Plain string → wrap as {"text": value} (mission_brief, roadmap_brief, etc.)
	// Already-parsed objects pass through unchanged.
	if strVal, isString := specData.(string); isString {
		strVal = datahelpers.StripCodeFences(strVal)
		var parsed map[string]interface{}
		if err := json.Unmarshal([]byte(strVal), &parsed); err == nil {
			logger.Info("write_site_spec: coerced JSON string spec_data to object")
			specData = parsed
		} else if strVal != "" {
			logger.Info("write_site_spec: wrapped plain string spec_data as {text:...}",
				zap.Int("len", len(strVal)))
			specData = map[string]interface{}{"text": strVal}
		} else {
			return nil, fmt.Errorf("spec_data is an empty string")
		}
	}

	specMap, ok := specData.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("spec_data must be a JSON object, got %T", specData)
	}

	// Read config literals
	config := params.StepConfig.Config
	aspect, _ := config["aspect"].(string)
	if aspect == "" {
		return nil, fmt.Errorf("aspect is required in config")
	}
	source, _ := config["source"].(string)
	if source == "" {
		return nil, fmt.Errorf("source is required in config")
	}

	sourceAgent, _ := config["source_agent"].(string)
	notes, _ := config["notes"].(string)

	var sourceItemID *uuid.UUID
	if itemIDStr, ok := config["source_item_id"].(string); ok && itemIDStr != "" {
		parsed, err := uuid.Parse(itemIDStr)
		if err == nil {
			sourceItemID = &parsed
		}
	}

	createdBy, _ := config["created_by"].(string)
	if createdBy == "" {
		createdBy = sourceAgent
	}
	if createdBy == "" {
		createdBy = source
	}
	// Auto-format content_direction specs: add a "formatted" field that
	// contains the entire spec as readable text. The content writer reads
	// {{.site_specs.specs.content_direction.formatted}} — one field regardless
	// of how the structured data is organised.
	// This runs for every content_direction write — classifier, adoption, HITL.
	if aspect == "content_direction" {
		formatted := datahelpers.FormatContentDirection(specMap)
		if formatted != "" {
			specMap["formatted"] = formatted
			logger.Info("Auto-formatted content_direction spec",
				zap.Int("formatted_len", len(formatted)))
		}
	}

	// --- Begin transaction ---
	tx, err := params.DB.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	// 1. Read current spec for this site+aspect
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

	// 2. Deep merge update over current
	merged := siteSpecDeepMerge(currentData, specMap)

	// Normalise services to object format if this is an identity spec
	if aspect == "identity" {
		normaliseServicesField(merged)
	}

	mergedJSON, err := json.Marshal(merged)
	if err != nil {
		return nil, fmt.Errorf("marshal merged spec: %w", err)
	}

	// 3. Mark old row as superseded (if exists)
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

	// 4. Insert new current row
	var newID uuid.UUID
	err = tx.QueryRowContext(ctx, `
		INSERT INTO site_specs (
			site_id, aspect, data,
			source, source_agent, source_item_id, notes,
			is_current, created_by
		) VALUES (
			$1, $2, $3::jsonb,
			$4, $5, $6, $7,
			true, $8
		) RETURNING id
	`, siteID, aspect, string(mergedJSON),
		source, siteSpecNilIfEmpty(sourceAgent), sourceItemID, siteSpecNilIfEmpty(notes),
		createdBy,
	).Scan(&newID)
	if err != nil {
		return nil, fmt.Errorf("insert new spec: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}

	logger.Info("WriteSiteSpecAction: spec written",
		zap.String("site_id", siteID.String()),
		zap.String("aspect", aspect),
		zap.String("source", source),
		zap.String("spec_id", newID.String()),
		zap.Bool("had_previous", oldID != nil),
	)

	return map[string]interface{}{
		"spec_id":      newID.String(),
		"site_id":      siteID.String(),
		"aspect":       aspect,
		"source":       source,
		"had_previous": oldID != nil,
		"written":      true,
	}, nil
}

// normaliseServicesField ensures services is always []{"name":"...","description":"..."}
// Handles: []string, []interface{} containing strings, and already-correct objects.
func normaliseServicesField(data map[string]interface{}) {
	if services, ok := data["services"]; ok {
		data["services"] = normaliseToNameDescArray(services)
	}
}

// normaliseToNameDescArray converts various service formats to a consistent
// []map[string]interface{}{"name": "...", "description": "..."} structure.
func normaliseToNameDescArray(raw interface{}) interface{} {
	arr, ok := raw.([]interface{})
	if !ok {
		return raw // not an array, leave as-is
	}
	if len(arr) == 0 {
		return raw
	}

	// Check first element to determine format
	switch arr[0].(type) {
	case string:
		// Plain strings -> convert to objects
		result := make([]interface{}, len(arr))
		for i, item := range arr {
			if name, ok := item.(string); ok {
				result[i] = map[string]interface{}{
					"name":        name,
					"description": "",
				}
			} else {
				result[i] = item
			}
		}
		return result

	case map[string]interface{}:
		// Already objects - ensure they have name and description keys
		for _, item := range arr {
			if m, ok := item.(map[string]interface{}); ok {
				if _, hasName := m["name"]; !hasName {
					for _, key := range []string{"title", "service_name", "label"} {
						if val, exists := m[key]; exists {
							m["name"] = val
							break
						}
					}
				}
				if _, hasDesc := m["description"]; !hasDesc {
					m["description"] = ""
				}
			}
		}
		return arr

	default:
		return raw
	}
}

// ============================================================================
// ACTION: read_site_spec
//
// Reads current spec(s) for a site. If aspect is specified in config, returns
// a single aspect's data. If aspect is omitted, returns all current aspects
// assembled into one map keyed by aspect name.
//
// Data inputs (via ActionInputSpec):
//   - site_id (required) — path to site UUID in collected_data
//
// Config literals (read directly):
//   - aspect (optional) — specific aspect to read. Omit for all.
// ============================================================================

func ReadSiteSpecAction(ctx context.Context, params ActionParams) (interface{}, error) {
	logger := params.Logger.With(zap.String("action", "read_site_spec"))

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}
	if params.DB == nil {
		return nil, fmt.Errorf("database connection required")
	}

	inputs, err := datahelpers.ExtractActionInputs(
		params.CollectedData,
		params.StepConfig.Config,
		ReadSiteSpecInputSpec,
		logger,
	)
	if err != nil {
		return nil, fmt.Errorf("input extraction failed: %w", err)
	}

	siteIDStr := inputs.Get("site_id")
	siteID, err := uuid.Parse(siteIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid site_id %q: %w", siteIDStr, err)
	}

	config := params.StepConfig.Config
	aspect, _ := config["aspect"].(string)

	// --- Single aspect mode ---
	if aspect != "" {
		var dataJSON []byte
		var specID uuid.UUID
		var source string

		err := params.DB.QueryRowContext(ctx, `
			SELECT id, data, source FROM site_specs
			WHERE site_id = $1 AND aspect = $2 AND is_current = true
		`, siteID, aspect).Scan(&specID, &dataJSON, &source)

		if err == sql.ErrNoRows {
			logger.Info("ReadSiteSpecAction: no current spec found",
				zap.String("site_id", siteID.String()),
				zap.String("aspect", aspect),
			)
			return map[string]interface{}{
				"found":   false,
				"aspect":  aspect,
				"site_id": siteID.String(),
				"data":    map[string]interface{}{},
			}, nil
		}
		if err != nil {
			return nil, fmt.Errorf("query spec: %w", err)
		}

		var data map[string]interface{}
		if err := json.Unmarshal(dataJSON, &data); err != nil {
			return nil, fmt.Errorf("unmarshal spec data: %w", err)
		}

		return map[string]interface{}{
			"found":   true,
			"spec_id": specID.String(),
			"aspect":  aspect,
			"site_id": siteID.String(),
			"source":  source,
			"data":    data,
		}, nil
	}

	// --- All aspects mode ---
	rows, err := params.DB.QueryContext(ctx, `
		SELECT aspect, data, source FROM site_specs
		WHERE site_id = $1 AND is_current = true
		ORDER BY aspect
	`, siteID)
	if err != nil {
		return nil, fmt.Errorf("query all specs: %w", err)
	}
	defer rows.Close()

	specs := make(map[string]interface{})
	aspects := []string{}

	for rows.Next() {
		var rowAspect, rowSource string
		var dataJSON []byte

		if err := rows.Scan(&rowAspect, &dataJSON, &rowSource); err != nil {
			logger.Warn("ReadSiteSpecAction: scan error", zap.Error(err))
			continue
		}

		var data map[string]interface{}
		if err := json.Unmarshal(dataJSON, &data); err != nil {
			logger.Warn("ReadSiteSpecAction: unmarshal error",
				zap.String("aspect", rowAspect), zap.Error(err))
			continue
		}

		specs[rowAspect] = data
		aspects = append(aspects, rowAspect)
	}

	logger.Info("ReadSiteSpecAction: specs loaded",
		zap.String("site_id", siteID.String()),
		zap.Int("aspect_count", len(aspects)),
		zap.Strings("aspects", aspects),
	)

	return map[string]interface{}{
		"site_id":      siteID.String(),
		"specs":        specs,
		"aspects":      aspects,
		"aspect_count": len(aspects),
	}, nil
}

// ============================================================================
// Private helpers
// ============================================================================

// siteSpecDeepMerge merges src into dst recursively. For nested maps, values
// in src override values in dst at the leaf level. Non-map values in src
// replace dst. Neither input is modified — returns a new map.
func siteSpecDeepMerge(dst, src map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})

	for k, v := range dst {
		result[k] = v
	}

	for k, srcVal := range src {
		dstVal, exists := result[k]
		if !exists {
			result[k] = srcVal
			continue
		}

		srcMap, srcIsMap := srcVal.(map[string]interface{})
		dstMap, dstIsMap := dstVal.(map[string]interface{})
		if srcIsMap && dstIsMap {
			result[k] = siteSpecDeepMerge(dstMap, srcMap)
		} else {
			result[k] = srcVal
		}
	}

	return result
}

// siteSpecNilIfEmpty returns nil for empty strings so DB columns get NULL.
func siteSpecNilIfEmpty(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}
