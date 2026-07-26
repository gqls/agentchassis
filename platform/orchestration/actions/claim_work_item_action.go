// FILE: platform/orchestration/actions/claim_work_item_action.go
//
// Atomically claims a work item for processing. Sets status='claimed',
// claimed_by, claimed_at. Uses optimistic locking — returns claimed: false
// if the item was already claimed or no longer eligible.
//
// Workflow config:
//   "claim_item": {
//       "action": "claim_work_item",
//       "config": {
//           "work_item_id": "pending_items.items.0.id"
//       },
//       "output_field": "claimed_item"
//   }
//
// Registration:
//   "claim_work_item": {
//       Handler:     ClaimWorkItemAction,
//       Category:    "site",
//       Description: "Atomically claim a work item for processing, preventing double-dispatch",
//       IsLocal:     true,
//   },

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

var ClaimWorkItemInputSpec = datahelpers.ActionInputSpec{
	Required:   []string{"work_item_id"},
	Optional:   []string{},
	Defaults:   map[string]interface{}{},
	Deprecated: map[string]string{},
}

func init() {
	datahelpers.RegisterActionInputSpec("claim_work_item", ClaimWorkItemInputSpec)
}

func ClaimWorkItemAction(ctx context.Context, params ActionParams) (interface{}, error) {
	logger := params.Logger.With(zap.String("action", "claim_work_item"))

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}
	if params.DB == nil {
		return nil, fmt.Errorf("database connection required")
	}

	inputs, err := datahelpers.ExtractActionInputs(
		params.CollectedData,
		params.StepConfig.Config,
		ClaimWorkItemInputSpec,
		logger,
	)
	if err != nil {
		return nil, fmt.Errorf("input extraction failed: %w", err)
	}

	itemIDStr := inputs.Get("work_item_id")

	// Fallback: resolve dot-notation path if ExtractActionInputs returned literal.
	// With Strategy 0 in ExtractActionInputs, this should no longer be needed.
	// If it triggers, something upstream has regressed.
	if itemIDStr == "" || strings.Contains(itemIDStr, ".") {
		if resolved := resolveConfigPath(params.StepConfig.Config, "work_item_id", params.CollectedData, logger); resolved != "" {
			logger.Warn("ClaimWorkItemAction: Strategy 0 missed config path, using fallback",
				zap.String("original", itemIDStr),
				zap.String("resolved", resolved),
			)
			itemIDStr = resolved
		}
	}

	itemID, err := uuid.Parse(itemIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid work_item_id %q: %w", itemIDStr, err)
	}

	claimedBy := params.AgentType
	if claimedBy == "" {
		claimedBy = "dispatch-loop"
	}

	// Atomic claim: only succeeds if item is still triaged/approved
	var claimedItemID string
	err = params.DB.QueryRowContext(ctx, `
		UPDATE site_work_items
		SET status = 'claimed',
		    claimed_by = $2,
		    claimed_at = NOW()
		WHERE id = $1
		  AND status IN ('triaged', 'approved')
		  AND attempt_count < max_attempts
		RETURNING id::text
	`, itemID, claimedBy).Scan(&claimedItemID)

	if err == sql.ErrNoRows {
		logger.Info("ClaimWorkItemAction: item not claimable (already claimed or ineligible)",
			zap.String("item_id", itemIDStr),
		)
		return map[string]interface{}{
			"claimed":      false,
			"work_item_id": itemIDStr,
			"reason":       "already_claimed_or_ineligible",
		}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("claim failed: %w", err)
	}

	logger.Info("ClaimWorkItemAction: claimed",
		zap.String("item_id", claimedItemID),
		zap.String("claimed_by", claimedBy),
	)

	// Handler existence check: verify the handler_agent is registered
	// in agent_definitions before returning. If not, release the claim
	// and mark as 'blocked' — the feasibility-recheck scheduled task
	// will promote it back to 'triaged' when the handler is deployed.
	var handlerAgent sql.NullString
	handlerLookupErr := params.DB.QueryRowContext(ctx, `
		SELECT handler_agent FROM site_work_items WHERE id = $1
	`, itemID).Scan(&handlerAgent)

	// The lookup error is now captured rather than discarded, because the
	// no-handler branch below cannot otherwise tell "this item has no handler"
	// from "we failed to read it". A transient DB error would leave handlerAgent
	// zero-valued and blocking on that would park a perfectly routable item.
	// On a read failure, log and fall through to the pre-existing behaviour
	// (skip the handler checks, let the item dispatch) — the same outcome this
	// code has always had when the read failed silently.
	if handlerLookupErr != nil {
		logger.Warn("ClaimWorkItemAction: could not read handler_agent, skipping handler checks",
			zap.String("item_id", claimedItemID),
			zap.Error(handlerLookupErr))
	}

	// No handler at all is the degenerate case of "handler not registered", and
	// it takes the same exit: an item nothing can route is blocked here, on its
	// first claim, rather than dispatched to a spawn_agent that must fail.
	// Without this it reaches spawn_handler with an empty agent_type_field,
	// fails into the loop's error_step, and is recorded as the generic
	// "Handler failed" — three attempts spent, and an error message that names
	// the wrong problem. 'blocked' is also the honest state: feasibility-recheck
	// promotes only where EXISTS(agent_definitions WHERE type = handler_agent),
	// and no agent type is the empty string, so this cannot become a retry loop.
	// Reachable via '' since migration 217 made the column NOT NULL DEFAULT ''
	// (bugs_closed/078); the !Valid arm covers a NULL from any pre-217 replica.
	if handlerLookupErr == nil && (!handlerAgent.Valid || handlerAgent.String == "") {
		params.DB.ExecContext(ctx, `
			UPDATE site_work_items
			SET status = 'blocked',
			    claimed_at = NULL,
			    claimed_by = NULL,
			    error = 'No handler_agent set — item cannot be routed to any agent',
			    updated_at = NOW()
			WHERE id = $1
		`, itemID)

		logger.Warn("ClaimWorkItemAction: work item has no handler_agent, item blocked",
			zap.String("item_id", claimedItemID))

		return map[string]interface{}{
			"claimed":       false,
			"work_item_id":  claimedItemID,
			"blocked":       true,
			"handler_agent": "",
			"reason":        "handler_not_set",
		}, nil
	}

	if handlerAgent.Valid && handlerAgent.String != "" {
		var handlerExists bool
		params.DB.QueryRowContext(ctx, `
			SELECT EXISTS(
				SELECT 1 FROM agent_definitions
				WHERE type = $1 AND deleted_at IS NULL
			)
		`, handlerAgent.String).Scan(&handlerExists)

		if !handlerExists {
			// Release claim, mark as blocked
			params.DB.ExecContext(ctx, `
				UPDATE site_work_items
				SET status = 'blocked',
				    claimed_at = NULL,
				    claimed_by = NULL,
				    error = 'Handler agent not registered: ' || $2,
				    updated_at = NOW()
				WHERE id = $1
			`, itemID, handlerAgent.String)

			logger.Info("ClaimWorkItemAction: handler not registered, item blocked",
				zap.String("item_id", claimedItemID),
				zap.String("handler_agent", handlerAgent.String))

			return map[string]interface{}{
				"claimed":       false,
				"work_item_id":  claimedItemID,
				"blocked":       true,
				"handler_agent": handlerAgent.String,
				"reason":        "handler_not_registered",
			}, nil
		}
	}

	// ── AI endpoint health check ──
	// After confirming the handler agent exists, check whether its AI
	// endpoint is healthy. If not, release the claim — the item stays
	// triaged and will be picked up when the endpoint recovers.
	if handlerAgent.Valid && handlerAgent.String != "" {
		aiEndpoint := extractAIEndpointFromHandler(ctx, params.DB, handlerAgent.String, logger)
		if aiEndpoint != "" {
			var healthy bool
			err := params.DB.QueryRowContext(ctx, `
				SELECT healthy FROM ai_endpoint_health WHERE endpoint_url = $1
			`, aiEndpoint).Scan(&healthy)

			if err == nil && !healthy {
				// Release the claim — item goes back to triaged
				params.DB.ExecContext(ctx, `
					UPDATE site_work_items
					SET status = 'triaged',
					    claimed_at = NULL,
					    claimed_by = NULL,
					    updated_at = NOW()
					WHERE id = $1
				`, itemID)

				logger.Info("ClaimWorkItemAction: AI endpoint unhealthy, releasing item",
					zap.String("item_id", claimedItemID),
					zap.String("handler", handlerAgent.String),
					zap.String("endpoint", aiEndpoint))

				return map[string]interface{}{
					"claimed":      false,
					"work_item_id": claimedItemID,
					"reason":       "ai_endpoint_unavailable",
					"endpoint":     aiEndpoint,
				}, nil
			}
			// If err != nil (no row in health table), assume healthy — don't block
			// items just because the health table hasn't been populated for this endpoint.
		}
	}

	return map[string]interface{}{
		"claimed":      true,
		"work_item_id": claimedItemID,
		"claimed_by":   claimedBy,
	}, nil
}

// extractAIEndpointFromHandler looks up the handler agent's definition,
// finds the first step with an ai_service config, and returns the endpoint URL.
// Returns "" if no AI endpoint is configured (e.g. algorithmic-only agents).
func extractAIEndpointFromHandler(ctx context.Context, db *sql.DB, handlerType string, logger *zap.Logger) string {
	var configJSON []byte
	err := db.QueryRowContext(ctx, `
		SELECT default_config FROM agent_definitions
		WHERE type = $1 AND is_active = true AND deleted_at IS NULL
		  AND (is_snapshot IS NULL OR is_snapshot = false)
		ORDER BY version DESC LIMIT 1
	`, handlerType).Scan(&configJSON)
	if err != nil {
		return ""
	}

	// Parse to find ai_service in any step
	var config map[string]interface{}
	if err := json.Unmarshal(configJSON, &config); err != nil {
		return ""
	}

	// Navigate: config.workflow.steps.*.config.ai_service
	workflow, _ := config["workflow"].(map[string]interface{})
	if workflow == nil {
		return ""
	}
	steps, _ := workflow["steps"].(map[string]interface{})
	if steps == nil {
		return ""
	}

	for _, stepVal := range steps {
		step, _ := stepVal.(map[string]interface{})
		if step == nil {
			continue
		}
		stepConfig, _ := step["config"].(map[string]interface{})
		if stepConfig == nil {
			continue
		}
		aiService, _ := stepConfig["ai_service"].(map[string]interface{})
		if aiService == nil {
			continue
		}

		// Found an ai_service block — extract the endpoint URL
		provider, _ := aiService["provider"].(string)
		apiURL, _ := aiService["api_url"].(string)

		if apiURL != "" {
			return apiURL
		}
		// Anthropic uses a known base URL
		if provider == "anthropic" {
			return "https://api.anthropic.com/v1/messages"
		}
	}
	return ""
}
