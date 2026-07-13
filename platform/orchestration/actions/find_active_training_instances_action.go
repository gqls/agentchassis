// FILE: platform/orchestration/actions/find_active_training_instances_action.go
//
// find_active_training_instances: the thunder-training-monitor's first step. It
// queries clients_db for Thunder instances that are currently running a training
// job and returns them as a list the monitor's loop step fans out over (one
// spawned monitor sub-agent per instance — "spawn sub-agents, not subworkflows").
//
// Output shape (consumed by LoopAction):
//   {
//     "instances": [
//       {"provisioning_id": "...", "training_run_id": "...",
//        "thunder_instance_id": "...", "instance_ip": "..."},
//       ...
//     ],
//     "count": <n>
//   }
// The monitor's loop step sets iterate_over: "<this step name>.instances" and a
// loop_var (e.g. "instance"); each substep's input_mapping then reads
// instance.provisioning_id / instance.training_run_id. The key is ALWAYS present
// (even as []), so a tick with no active instances is a graceful loop skip rather
// than a missing-collection error.
//
// Population (verified against schemas_all — public.thunder_instances, clients_db):
//   status = 'running'                 -> only live boxes ('decommissioning'/
//                                         'decommissioned'/'reaped'/'lost'/'failed'
//                                         excluded by the status filter)
//   training_run_id IS NOT NULL        -> it's actually running a training job
//   decommission_requested_at IS NULL  -> don't re-probe a box a prior tick has
//                                         already asked to decommission
// id and training_run_id are cast ::text so they scan into Go strings regardless
// of the driver's uuid handling; provisioning_id is thunder_instances.id, which is
// what ssh_get_status and decommission_instance resolve against.
//
// DB: params.DB must be the clients_db pool (same binding as the other Phase 5
// training/thunder actions — mark_training_run_*, record_probe_streak).

package actions

import (
	"context"
	"database/sql"
	"fmt"

	"go.uber.org/zap"
)

// FindActiveTrainingInstancesAction returns the running training instances for
// the monitor to fan out over.
func FindActiveTrainingInstancesAction(ctx context.Context, params ActionParams) (interface{}, error) {
	logger := params.Logger.With(zap.String("action", "find_active_training_instances"))

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}
	if params.DB == nil {
		return nil, fmt.Errorf("find_active_training_instances: database connection required")
	}

	rows, err := params.DB.QueryContext(ctx, `
		SELECT id::text,
		       training_run_id::text,
		       thunder_instance_id,
		       instance_ip
		  FROM public.thunder_instances
		 WHERE status = 'running'
		   AND training_run_id IS NOT NULL
		   AND decommission_requested_at IS NULL
		 ORDER BY running_since ASC NULLS LAST
	`)
	if err != nil {
		return nil, fmt.Errorf("find_active_training_instances: query failed: %w", err)
	}
	defer rows.Close()

	instances := make([]map[string]interface{}, 0)
	for rows.Next() {
		var (
			provisioningID    string
			trainingRunID     string
			thunderInstanceID string
			instanceIP        sql.NullString
		)
		if err := rows.Scan(&provisioningID, &trainingRunID, &thunderInstanceID, &instanceIP); err != nil {
			return nil, fmt.Errorf("find_active_training_instances: scan failed: %w", err)
		}
		item := map[string]interface{}{
			"provisioning_id":     provisioningID,
			"training_run_id":     trainingRunID,
			"thunder_instance_id": thunderInstanceID,
		}
		if instanceIP.Valid {
			item["instance_ip"] = instanceIP.String
		}
		instances = append(instances, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("find_active_training_instances: row iteration failed: %w", err)
	}

	logger.Info("Found active training instances",
		zap.Int("count", len(instances)),
	)

	return map[string]interface{}{
		"instances": instances,
		"count":     len(instances),
	}, nil
}
