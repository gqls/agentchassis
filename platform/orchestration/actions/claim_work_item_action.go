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

	// Fallback: resolve dot-notation path if ExtractActionInputs returned literal
	if itemIDStr == "" || strings.Contains(itemIDStr, ".") {
		if resolved := resolveConfigPath(params.StepConfig.Config, "work_item_id", params.CollectedData, logger); resolved != "" {
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

	return map[string]interface{}{
		"claimed":      true,
		"work_item_id": claimedItemID,
		"claimed_by":   claimedBy,
	}, nil
}
