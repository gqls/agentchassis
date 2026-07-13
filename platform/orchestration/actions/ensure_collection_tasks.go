// FILE: platform/orchestration/actions/ensure_collection_tasks.go
//
// EnsureCollectionTasksAction creates collection_tasks for any businesses
// with verification_status='pending' that don't already have a pending task.
//
// This is a safety net that runs between promote and verify in the pipeline.
// It catches cases where:
//   - A previous promote run failed to create tasks (DB errors, schema issues)
//   - Businesses were imported manually without tasks
//   - Tasks were accidentally deleted
//
// Idempotent — the partial unique index on (business_id, task_type)
// WHERE status='pending' prevents duplicates even without ON CONFLICT.
//
// Workflow config:
//
//	"ensure_tasks": {
//	    "action": "ensure_collection_tasks",
//	    "config": {
//	        "vertical_slug": "veterinary",
//	        "task_type": "initial_verification",
//	        "task_priority": 5
//	    },
//	    "output_field": "ensure_result"
//	}
//
// Registration:
//   "ensure_collection_tasks": EnsureCollectionTasksAction,

package actions

import (
	"context"
	"fmt"

	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"go.uber.org/zap"
)

var EnsureCollectionTasksInputSpec = datahelpers.ActionInputSpec{
	Required: []string{"vertical_slug"},
	Optional: []string{"task_type", "task_priority"},
	Defaults: map[string]interface{}{
		"task_type":     "initial_verification",
		"task_priority": 5,
	},
}

func init() {
	datahelpers.RegisterActionInputSpec("ensure_collection_tasks", EnsureCollectionTasksInputSpec)
}

func EnsureCollectionTasksAction(ctx context.Context, params ActionParams) (interface{}, error) {
	params.Logger.Info("EnsureCollectionTasksAction: Starting")

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}

	if params.DB == nil {
		return nil, fmt.Errorf("no database connection available")
	}

	inputs, err := datahelpers.ExtractActionInputs(
		params.CollectedData,
		params.StepConfig.Config,
		EnsureCollectionTasksInputSpec,
		params.Logger,
	)
	if err != nil {
		return nil, fmt.Errorf("input extraction failed: %w", err)
	}

	verticalSlug := inputs.Get("vertical_slug")
	taskType := inputs.Get("task_type")
	taskPriority := inputs.GetInt("task_priority", 5)

	// Get vertical ID
	var verticalID string
	err = params.DB.QueryRowContext(ctx,
		`SELECT id FROM business_intel.business_verticals WHERE slug = $1`,
		verticalSlug,
	).Scan(&verticalID)
	if err != nil {
		return nil, fmt.Errorf("vertical '%s' not found: %w", verticalSlug, err)
	}

	// Insert tasks for any pending businesses missing them
	result, err := params.DB.ExecContext(ctx, `
		INSERT INTO business_intel.collection_tasks
			(business_id, task_type, vertical_id, priority, status, created_at, updated_at)
		SELECT 
			b.id, $2, b.vertical_id, $3, 'pending', NOW(), NOW()
		FROM business_intel.businesses b
		WHERE b.verification_status IN ('pending', 'seed_import')
		  AND b.vertical_id = $1
		  AND NOT EXISTS (
			  SELECT 1 FROM business_intel.collection_tasks ct
			  WHERE ct.business_id = b.id
				AND ct.task_type = $2
				AND ct.status IN ('pending', 'in_progress')
		  )`,
		verticalID, taskType, taskPriority)
	if err != nil {
		return nil, fmt.Errorf("failed to backfill collection_tasks: %w", err)
	}

	created, _ := result.RowsAffected()

	params.Logger.Info("EnsureCollectionTasksAction: done",
		zap.Int64("tasks_created", created),
		zap.String("vertical", verticalSlug),
		zap.String("task_type", taskType))

	return map[string]interface{}{
		"tasks_created": created,
		"vertical_slug": verticalSlug,
		"task_type":     taskType,
	}, nil
}
