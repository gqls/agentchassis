The monitor itself is a periodic fan-out dispatcher, not the thing that probes the box. Its workflow is:

find_instances (find_active_training_instances) — queries clients_db for thunder_instances rows that are running, have a training_run_id, and have no decommission pending, returning the list. So step one is "which GPUs are we currently running training on?"
monitor_loop — for each such instance (sequentially, up to 25, continue_on_error:true so one bad box doesn't abort the tick): spawn_worker spawns a thunder-training-monitor-worker Job, then call_worker calls it with that instance's provisioning_id + training_run_id and awaits its terminal result.
done.

it's neither purely "check VM access" nor "check the GPU" — it's lifecycle/cost management. 

The monitor asks the DB "what's still running," and each worker then probes its box (via ssh_get_status: reachable? training process still alive / finished?) and decommissions when the run is done or the box is dead. The probe answers both your sub-questions at once — ssh_get_status returns reachable (VM access) and runs a status command (is the GPU job still going).

SELECT jsonb_pretty(default_config) FROM agent_definitions
WHERE type='thunder-training-monitor-worker' AND (is_snapshot IS NULL OR is_snapshot=false) AND deleted_at IS NULL;

\d thunder_instances