// FILE: platform/orchestration/actions/checkpoint_for_review_action.go
//
// CheckpointForReviewAction allows any agent to pause its pipeline at a
// human review point WITHOUT suspending the orchestration.
//
// What it does:
//   1. Extracts review data from collected_data (configurable)
//   2. Optionally saves it to site_specs (versioned, auditable)
//   3. Creates a site_work_item with status needs_human_review
//   4. Embeds "on_approve" instructions in the work item spec
//   5. The workflow then completes normally
//
// The dashboard shows the item. When the admin approves (with optional edits),
// the dashboard endpoint:
//   a. Updates site_specs with the corrected data
//   b. Creates the follow-on work item specified in on_approve
//   c. Marks the review item complete
//
// This replaces the request_human_input → suspended orchestration pattern
// with a database-mediated handoff between short-lived orchestrations.
//
// Workflow config example:
//   "request_review": {
//       "action": "checkpoint_for_review",
//       "config": {
//           "item_type": "needs_brief_review",
//           "summary_from": "Review brief for {{domain}}",
//           "severity": "high",
//           "review_fields_from": "generated_brief",
//           "save_spec_aspect": "generated_brief",
//           "site_id_from": "site_record.site_id",
//           "page_id_from": "current_page.id",
//           "on_approve": {
//               "item_type": "ready_to_build",
//               "handler_agent": "pageflow-builder",
//               "include_fields": ["reviewed_brief", "site_record"]
//           }
//       },
//       "next_step": "complete"
//   }
//
// Registration:
//   "checkpoint_for_review": {
//       Handler:     CheckpointForReviewAction,
//       Category:    "orchestration",
//       Description: "Save work and create human review item",
//       IsLocal:     true,
//   },

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

var CheckpointForReviewInputSpec = datahelpers.ActionInputSpec{
	Required:   []string{"site_id"},
	Optional:   []string{"item_type", "summary_from", "severity", "review_fields_from", "save_spec_aspect", "page_id_from", "on_approve"},
	Defaults:   map[string]interface{}{"severity": "medium", "item_type": "needs_human_review"},
	Deprecated: map[string]string{},
}

func init() {
	datahelpers.RegisterActionInputSpec("checkpoint_for_review", CheckpointForReviewInputSpec)
}

func CheckpointForReviewAction(ctx context.Context, params ActionParams) (interface{}, error) {
	logger := params.Logger.With(zap.String("action", "checkpoint_for_review"))

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}
	if params.DB == nil {
		return nil, fmt.Errorf("database connection required")
	}

	config := params.StepConfig.Config

	// ── Extract site_id ──────────────────────────────────────────────────
	siteIDPath := "site_record.site_id"
	if s, ok := config["site_id_from"].(string); ok && s != "" {
		siteIDPath = s
	}
	siteIDStr := datahelpers.ExtractNestedFieldString(params.CollectedData, siteIDPath)
	if siteIDStr == "" {
		siteIDStr = datahelpers.ExtractNestedFieldString(params.CollectedData, "input_data.site_id")
	}
	if siteIDStr == "" {
		return nil, fmt.Errorf("site_id not found at path %q or input_data.site_id", siteIDPath)
	}
	siteID, err := uuid.Parse(siteIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid site_id %q: %w", siteIDStr, err)
	}

	// ── Extract config values ────────────────────────────────────────────
	itemType, _ := config["item_type"].(string)
	if itemType == "" {
		itemType = "needs_human_review"
	}

	severity, _ := config["severity"].(string)
	if severity == "" {
		severity = "medium"
	}

	// ── Build summary ────────────────────────────────────────────────────
	summary := "Review required"
	if s, ok := config["summary_from"].(string); ok && s != "" {
		summary = expandSimpleTemplate(s, params.CollectedData)
	}

	// ── Extract review data ──────────────────────────────────────────────
	reviewFieldsFrom, _ := config["review_fields_from"].(string)
	var reviewData interface{}
	if reviewFieldsFrom != "" {
		reviewData = datahelpers.ExtractNestedField(params.CollectedData, reviewFieldsFrom)
	}
	if reviewData == nil {
		// Fallback: use the full collected_data minus internal fields
		reviewData = stripInternalFields(params.CollectedData)
	}

	// ── Optional: page_id ────────────────────────────────────────────────
	var pageID *uuid.UUID
	if pidPath, ok := config["page_id_from"].(string); ok && pidPath != "" {
		pidStr := datahelpers.ExtractNestedFieldString(params.CollectedData, pidPath)
		if pidStr != "" {
			if parsed, err := uuid.Parse(pidStr); err == nil {
				pageID = &parsed
			}
		}
	}

	// ── Optional: save to site_specs ─────────────────────────────────────
	specAspect, _ := config["save_spec_aspect"].(string)
	if specAspect != "" {
		reviewJSON, err := json.Marshal(reviewData)
		if err != nil {
			logger.Warn("Failed to marshal review data for spec save", zap.Error(err))
		} else {
			if err := saveCheckpointSpec(ctx, params, siteID, specAspect, reviewJSON, logger); err != nil {
				logger.Warn("Failed to save checkpoint spec", zap.Error(err))
				// Non-fatal — work item still gets created
			}
		}
	}

	// ── Build work item spec ─────────────────────────────────────────────
	spec := map[string]interface{}{
		"review_data":    reviewData,
		"checkpoint":     true,
		"source_agent":   params.AgentType,
		"correlation_id": params.ExecutionContext.CorrelationID,
	}

	// Get domain for context
	domain := datahelpers.ExtractNestedFieldString(params.CollectedData, "site_record.domain")
	if domain == "" {
		domain = datahelpers.ExtractNestedFieldString(params.CollectedData, "input_data.domain")
	}
	if domain != "" {
		spec["domain"] = domain
	}

	if specAspect != "" {
		spec["spec_aspect"] = specAspect
	}

	// Embed on_approve instructions
	if onApprove, ok := config["on_approve"].(map[string]interface{}); ok {
		spec["on_approve"] = onApprove
	}

	specJSON, err := json.Marshal(spec)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal work item spec: %w", err)
	}

	// ── Determine priority ───────────────────────────────────────────────
	priority := 20 // Human review items are high priority
	switch severity {
	case "high":
		priority = 10
	case "low":
		priority = 40
	}

	// ── Create work item ─────────────────────────────────────────────────
	var newItemID uuid.UUID
	err = params.DB.QueryRowContext(ctx, `
		INSERT INTO site_work_items (
			site_id, source, domain, item_type, severity, summary,
			spec, page_id, priority, handler_agent, status, created_by
		) VALUES ($1, 'checkpoint', 'build', $2, $3, $4, $5::jsonb, $6, $7, 'human-review', 'needs_human_review', $8)
		RETURNING id
	`, siteID, itemType, severity, summary,
		string(specJSON), pageID, priority,
		params.AgentType,
	).Scan(&newItemID)

	if err != nil {
		return nil, fmt.Errorf("failed to create review work item: %w", err)
	}

	logger.Info("Checkpoint created for human review",
		zap.String("work_item_id", newItemID.String()),
		zap.String("item_type", itemType),
		zap.String("site_id", siteIDStr),
		zap.String("spec_aspect", specAspect),
		zap.String("source_agent", params.AgentType))

	return map[string]interface{}{
		"work_item_id":  newItemID.String(),
		"item_type":     itemType,
		"status":        "needs_human_review",
		"spec_aspect":   specAspect,
		"checkpoint":    true,
		"review_queued": true,
	}, nil
}

// saveCheckpointSpec saves the review data to site_specs with versioning.
func saveCheckpointSpec(ctx context.Context, params ActionParams, siteID uuid.UUID, aspect string, data []byte, logger *zap.Logger) error {
	tx, err := params.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Supersede current version
	_, err = tx.ExecContext(ctx, `
		UPDATE site_specs
		SET is_current = false, superseded_at = NOW()
		WHERE site_id = $1 AND aspect = $2 AND is_current = true
	`, siteID, aspect)
	if err != nil {
		return err
	}

	// Insert new version
	_, err = tx.ExecContext(ctx, `
		INSERT INTO site_specs (site_id, aspect, data, source, created_by, is_current)
		VALUES ($1, $2, $3, 'checkpoint', $4, true)
	`, siteID, aspect, data, params.AgentType)
	if err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	logger.Info("Checkpoint data saved to site_specs",
		zap.String("site_id", siteID.String()),
		zap.String("aspect", aspect))

	return nil
}

// expandSimpleTemplate replaces {{.path}} tokens with values from collected data.
// Handles simple cases like "Review brief for {{.site_record.domain}}".
func expandSimpleTemplate(template string, data map[string]interface{}) string {
	result := template
	// Find all {{.path}} patterns
	for {
		start := strings.Index(result, "{{.")
		if start == -1 {
			break
		}
		end := strings.Index(result[start:], "}}")
		if end == -1 {
			break
		}
		end += start + 2

		path := result[start+3 : end-2]
		value := datahelpers.ExtractNestedFieldString(data, path)
		if value == "" {
			// Try without the leading dot path
			value = datahelpers.ExtractNestedFieldString(data, strings.TrimPrefix(path, "."))
		}
		if value == "" {
			value = path // Leave the path name as placeholder
		}

		result = result[:start] + value + result[end:]
	}
	return result
}

// stripInternalFields removes chassis internal keys from collected data
// so the review spec only contains meaningful business data.
func stripInternalFields(data map[string]interface{}) map[string]interface{} {
	clean := make(map[string]interface{})
	for k, v := range data {
		if strings.HasPrefix(k, "__") && strings.HasSuffix(k, "__") {
			continue
		}
		clean[k] = v
	}
	return clean
}
