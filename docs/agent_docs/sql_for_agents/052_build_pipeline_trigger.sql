-- build-pipeline-trigger agent definition
-- Heartbeat agent that processes the build queue and triggers dispatch loops.
--
-- Each invocation:
--   1. Runs seed_build_queue (processes queued entries → creates sites + first work items)
--   2. Queries for any site with pending work items
--   3. If found: spawns + calls build-dispatch-loop for that site
--   4. Completes
--
-- Designed to be triggered by a K8s CronJob every 30 minutes, or manually.
-- Only handles one site per invocation — the next heartbeat picks up the next.
-- Safe to run concurrently: dispatch loop's claim_work_item prevents double-processing.
--
-- Trigger: CronJob or manual kafka message to system.agent.build-pipeline-trigger.requests

INSERT INTO agent_definitions (
    type,
    display_name,
    description,
    category,
    default_config,
    is_active,
    capabilities,
    image_repository,
    image_tag,
    resources,
    topics,
    health_config,
    env_vars,
    version,
    delegation_preferences,
    agent_category,
    status,
    domain_tags,
    briefing_questionnaire,
    usage_count,
    is_snapshot,
    input_contract,
    output_contract
) VALUES (
             'build-pipeline-trigger',
             'Build Pipeline Trigger',
             'Heartbeat agent: seeds build queue entries, finds sites with pending work items, triggers dispatch loop. One site per invocation.',
             'orchestrator',
             '{
                 "workflow": {
                     "start_step": "seed_queue",
                     "steps": {

                         "seed_queue": {
                             "action": "seed_build_queue",
                             "config": {
                                 "max_entries": 10
                             },
                             "next_step": "find_dispatchable_site",
                             "description": "Process queued entries from build_queue — creates site records and initial work items",
                             "output_field": "seed_result"
                         },

                         "find_dispatchable_site": {
                             "action": "query_database",
                             "config": {
                                 "query": "SELECT DISTINCT ON (wi.site_id) wi.site_id::text, s.domain FROM site_work_items wi JOIN sites s ON s.id = wi.site_id WHERE wi.status IN (''triaged'', ''approved'') AND wi.domain = ''build'' AND wi.attempt_count < wi.max_attempts AND NOT EXISTS (SELECT 1 FROM site_work_items active WHERE active.site_id = wi.site_id AND active.status = ''claimed'') ORDER BY wi.site_id, wi.priority ASC LIMIT 1",
                                 "output_format": "object"
                             },
                             "next_step": "check_has_site",
                             "description": "Find a site with pending build items that has no actively claimed items",
                             "output_field": "dispatchable"
                         },

                         "check_has_site": {
                             "action": "conditional",
                             "config": {
                                 "condition": "dispatchable.count > 0",
                                 "then_step": "spawn_dispatch",
                                 "else_step": "complete_idle"
                             },
                             "description": "Check if there is a site needing dispatch"
                         },

                         "spawn_dispatch": {
                             "action": "spawn_agent",
                             "config": {
                                 "role": "dispatcher",
                                 "agent_type": "build-dispatch-loop",
                                 "error_step": "complete_idle"
                             },
                             "next_step": "call_dispatch",
                             "description": "Spawn dispatch loop for the site",
                             "output_field": "dispatch_spawned"
                         },

                         "call_dispatch": {
                             "action": "call_agent",
                             "config": {
                                 "agent_type": "build-dispatch-loop",
                                 "target_role": "dispatcher",
                                 "input_mapping": {
                                     "site_id": "dispatchable.rows.0.site_id",
                                     "domain":  "dispatchable.rows.0.domain"
                                 },
                                 "timeout_seconds": 900,
                                 "error_step": "complete_idle"
                             },
                             "next_step": "complete",
                             "description": "Call dispatch loop — it processes one item and chains for the rest",
                             "output_field": "dispatch_result"
                         },

                         "complete_idle": {
                             "action": "complete_workflow",
                             "config": {
                                 "output_fields": ["seed_result"],
                                 "success_message": "No sites need dispatching"
                             },
                             "description": "Nothing to dispatch this cycle"
                         },

                         "complete": {
                             "action": "complete_workflow",
                             "config": {
                                 "output_fields": ["seed_result", "dispatch_result"]
                             },
                             "description": "Pipeline trigger complete"
                         }
                     }
                 },
                 "processing_mode": "orchestrator",
                 "timeout_seconds": 1200
             }'::jsonb,
             true,
             '["trigger", "dispatch", "build-queue"]'::jsonb,
             'docker.io/aqls/agent-chassis',
             'v1.0.803',
             '{"limits": {"cpu": "250m", "memory": "256Mi"}, "requests": {"cpu": "50m", "memory": "128Mi"}}'::jsonb,
             '{"error": "system.errors.{type}", "process": "system.agent.{type}.process", "response": "system.responses.{type}"}'::jsonb,
             '{"port": 8080, "liveness_path": "/health", "readiness_path": "/ready", "initial_delay_seconds": 15}'::jsonb,
             '[]'::jsonb,
             1,
             '{"fallback_to_self": true, "prefer_delegation": true}'::jsonb,
             'coordinator',
             'experimental',
             '["build", "trigger", "pipeline"]'::jsonb,
             '{}'::jsonb,
             0,
             false,
             '{"required": [], "optional": [], "description": "No input required. Reads from build_queue and site_work_items tables."}'::jsonb,
             '{"produces": {"seed_result": "How many entries were seeded", "dispatch_result": "Result from the dispatch loop (if triggered)"}}'::jsonb
         )
    ON CONFLICT (type, version) DO UPDATE SET
    display_name = EXCLUDED.display_name,
                                       description = EXCLUDED.description,
                                       category = EXCLUDED.category,
                                       default_config = EXCLUDED.default_config,
                                       is_active = EXCLUDED.is_active,
                                       capabilities = EXCLUDED.capabilities,
                                       image_tag = EXCLUDED.image_tag,
                                       resources = EXCLUDED.resources,
                                       agent_category = EXCLUDED.agent_category,
                                       domain_tags = EXCLUDED.domain_tags,
                                       input_contract = EXCLUDED.input_contract,
                                       output_contract = EXCLUDED.output_contract,
                                       updated_at = now();

---

-- path fix

-- build-pipeline-trigger agent definition
-- Heartbeat agent that processes the build queue and triggers dispatch loops.
--
-- Each invocation:
--   1. Runs seed_build_queue (processes queued entries → creates sites + first work items)
--   2. Queries for any site with pending work items
--   3. If found: spawns + calls build-dispatch-loop for that site
--   4. Completes
--
-- Designed to be triggered by a K8s CronJob every 30 minutes, or manually.
-- Only handles one site per invocation — the next heartbeat picks up the next.
-- Safe to run concurrently: dispatch loop's claim_work_item prevents double-processing.
--
-- Trigger: CronJob or manual kafka message to system.agent.build-pipeline-trigger.requests

INSERT INTO agent_definitions (
    type,
    display_name,
    description,
    category,
    default_config,
    is_active,
    capabilities,
    image_repository,
    image_tag,
    resources,
    topics,
    health_config,
    env_vars,
    version,
    delegation_preferences,
    agent_category,
    status,
    domain_tags,
    briefing_questionnaire,
    usage_count,
    is_snapshot,
    input_contract,
    output_contract
) VALUES (
             'build-pipeline-trigger',
             'Build Pipeline Trigger',
             'Heartbeat agent: seeds build queue entries, finds sites with pending work items, triggers dispatch loop. One site per invocation.',
             'orchestrator',
             '{
                 "workflow": {
                     "start_step": "seed_queue",
                     "steps": {

                         "seed_queue": {
                             "action": "seed_build_queue",
                             "config": {
                                 "max_entries": 10
                             },
                             "next_step": "find_dispatchable_site",
                             "description": "Process queued entries from build_queue — creates site records and initial work items",
                             "output_field": "seed_result"
                         },

                         "find_dispatchable_site": {
                             "action": "query_database",
                             "config": {
                                 "query": "SELECT DISTINCT ON (wi.site_id) wi.site_id::text, s.domain FROM site_work_items wi JOIN sites s ON s.id = wi.site_id WHERE wi.status IN (''triaged'', ''approved'') AND wi.domain = ''build'' AND wi.attempt_count < wi.max_attempts AND NOT EXISTS (SELECT 1 FROM site_work_items active WHERE active.site_id = wi.site_id AND active.status = ''claimed'') ORDER BY wi.site_id, wi.priority ASC LIMIT 1",
                                 "output_format": "object"
                             },
                             "next_step": "check_has_site",
                             "description": "Find a site with pending build items that has no actively claimed items",
                             "output_field": "dispatchable"
                         },

                         "check_has_site": {
                             "action": "conditional",
                             "config": {
                                 "condition": "dispatchable.count > 0",
                                 "then_step": "spawn_dispatch",
                                 "else_step": "complete_idle"
                             },
                             "description": "Check if there is a site needing dispatch"
                         },

                         "spawn_dispatch": {
                             "action": "spawn_agent",
                             "config": {
                                 "role": "dispatcher",
                                 "agent_type": "build-dispatch-loop",
                                 "error_step": "complete_idle"
                             },
                             "next_step": "call_dispatch",
                             "description": "Spawn dispatch loop for the site",
                             "output_field": "dispatch_spawned"
                         },

                         "call_dispatch": {
                             "action": "call_agent",
                             "config": {
                                 "agent_type": "build-dispatch-loop",
                                 "target_role": "dispatcher",
                                 "input_mapping": {
                                     "site_id": "dispatchable.rows.0.site_id",
                                     "domain":  "dispatchable.rows.0.domain"
                                 },
                                 "timeout_seconds": 900,
                                 "error_step": "complete_idle"
                             },
                             "next_step": "complete",
                             "description": "Call dispatch loop — it processes one item and chains for the rest",
                             "output_field": "dispatch_result"
                         },

                         "complete_idle": {
                             "action": "complete_workflow",
                             "config": {
                                 "output_fields": ["seed_result"],
                                 "success_message": "No sites need dispatching"
                             },
                             "description": "Nothing to dispatch this cycle"
                         },

                         "complete": {
                             "action": "complete_workflow",
                             "config": {
                                 "output_fields": ["seed_result", "dispatch_result"]
                             },
                             "description": "Pipeline trigger complete"
                         }
                     }
                 },
                 "processing_mode": "orchestrator",
                 "timeout_seconds": 1200
             }'::jsonb,
             true,
             '["trigger", "dispatch", "build-queue"]'::jsonb,
             'docker.io/aqls/agent-chassis',
             'v1.0.803',
             '{"limits": {"cpu": "250m", "memory": "256Mi"}, "requests": {"cpu": "50m", "memory": "128Mi"}}'::jsonb,
             '{"error": "system.errors.{type}", "process": "system.agent.{type}.process", "response": "system.responses.{type}"}'::jsonb,
             '{"port": 8080, "liveness_path": "/health", "readiness_path": "/ready", "initial_delay_seconds": 15}'::jsonb,
             '[]'::jsonb,
             1,
             '{"fallback_to_self": true, "prefer_delegation": true}'::jsonb,
             'coordinator',
             'experimental',
             '["build", "trigger", "pipeline"]'::jsonb,
             '{}'::jsonb,
             0,
             false,
             '{"required": [], "optional": [], "description": "No input required. Reads from build_queue and site_work_items tables."}'::jsonb,
             '{"produces": {"seed_result": "How many entries were seeded", "dispatch_result": "Result from the dispatch loop (if triggered)"}}'::jsonb
         )
    ON CONFLICT (type, version) DO UPDATE SET
    display_name = EXCLUDED.display_name,
                                       description = EXCLUDED.description,
                                       category = EXCLUDED.category,
                                       default_config = EXCLUDED.default_config,
                                       is_active = EXCLUDED.is_active,
                                       capabilities = EXCLUDED.capabilities,
                                       image_tag = EXCLUDED.image_tag,
                                       resources = EXCLUDED.resources,
                                       agent_category = EXCLUDED.agent_category,
                                       domain_tags = EXCLUDED.domain_tags,
                                       input_contract = EXCLUDED.input_contract,
                                       output_contract = EXCLUDED.output_contract,
                                       updated_at = now();

--

-- data path fix

-- Fix build-pipeline-trigger: update input_mapping paths
-- After the Go patch (flatten first row), paths change from
-- dispatchable.rows.0.x to dispatchable.x
--
-- The conditional "dispatchable.count > 0" still works unchanged.

UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,call_dispatch,config,input_mapping}',
        '{
            "site_id": "dispatchable.site_id",
            "domain":  "dispatchable.domain"
        }'::jsonb
                     ),
    updated_at = now()
WHERE type = 'build-pipeline-trigger';

-- Verify
SELECT
    default_config->'workflow'->'steps'->'call_dispatch'->'config'->'input_mapping' as input_mapping,
    default_config->'workflow'->'steps'->'check_has_site'->'config'->'condition' as condition
FROM agent_definitions
WHERE type = 'build-pipeline-trigger';

-- backup
clients_db=# SELECT * FROM agent_definitions WHERE type = 'build-pipeline-trigger';
id                  |          type          |      display_name      |                                                            description                                                            |   category   |                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                             default_config                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                             | is_active |          created_at           |          updated_at           | deleted_at |              capabilities              |       image_repository       | image_tag | command |                                           resources                                           |                                                       topics                                                       |                                            health_config                                            | env_vars | version | previous_version_id | task_workflow | orchestrator_workflow | orchestration_workflow |                delegation_preferences                 | agent_category |    status    |           domain_tags            | briefing_questionnaire | usage_count | is_snapshot |                                                      input_contract                                                      |                                                         output_contract
--------------------------------------+------------------------+------------------------+-----------------------------------------------------------------------------------------------------------------------------------+--------------+----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------+-----------+-------------------------------+-------------------------------+------------+----------------------------------------+------------------------------+-----------+---------+-----------------------------------------------------------------------------------------------+--------------------------------------------------------------------------------------------------------------------+-----------------------------------------------------------------------------------------------------+----------+---------+---------------------+---------------+-----------------------+------------------------+-------------------------------------------------------+----------------+--------------+----------------------------------+------------------------+-------------+-------------+--------------------------------------------------------------------------------------------------------------------------+----------------------------------------------------------------------------------------------------------------------------------
 bb291e23-ca22-4e85-85d1-df7e24b70b42 | build-pipeline-trigger | Build Pipeline Trigger | Heartbeat agent: seeds build queue entries, finds sites with pending work items, triggers dispatch loop. One site per invocation. | orchestrator | {"workflow": {"steps": {"complete": {"action": "complete_workflow", "config": {"output_fields": ["seed_result", "dispatch_result"]}, "description": "Pipeline trigger complete"}, "seed_queue": {"action": "seed_build_queue", "config": {"max_entries": 10}, "next_step": "find_dispatchable_site", "description": "Process queued entries from build_queue — creates site records and initial work items", "output_field": "seed_result"}, "call_dispatch": {"action": "call_agent", "config": {"agent_type": "build-dispatch-loop", "error_step": "complete_idle", "target_role": "dispatcher", "input_mapping": {"domain": "dispatchable.domain", "site_id": "dispatchable.site_id"}, "timeout_seconds": 900}, "next_step": "complete", "description": "Call dispatch loop — it processes one item and chains for the rest", "output_field": "dispatch_result"}, "complete_idle": {"action": "complete_workflow", "config": {"output_fields": ["seed_result"], "success_message": "No sites need dispatching"}, "description": "Nothing to dispatch this cycle"}, "check_has_site": {"action": "conditional", "config": {"condition": "dispatchable.count > 0", "else_step": "complete_idle", "then_step": "spawn_dispatch"}, "description": "Check if there is a site needing dispatch"}, "spawn_dispatch": {"action": "spawn_agent", "config": {"role": "dispatcher", "agent_type": "build-dispatch-loop", "error_step": "complete_idle"}, "next_step": "call_dispatch", "description": "Spawn dispatch loop for the site", "output_field": "dispatch_spawned"}, "find_dispatchable_site": {"action": "query_database", "config": {"query": "SELECT DISTINCT ON (wi.site_id) wi.site_id::text, s.domain FROM site_work_items wi JOIN sites s ON s.id = wi.site_id WHERE wi.status IN ('triaged', 'approved') AND wi.domain = 'build' AND wi.attempt_count < wi.max_attempts AND NOT EXISTS (SELECT 1 FROM site_work_items active WHERE active.site_id = wi.site_id AND active.status = 'claimed') ORDER BY wi.site_id, wi.priority ASC LIMIT 1", "output_format": "object"}, "next_step": "check_has_site", "description": "Find a site with pending build items that has no actively claimed items", "output_field": "dispatchable"}}, "start_step": "seed_queue"}, "processing_mode": "orchestrator", "timeout_seconds": 1200} | t         | 2026-02-24 19:33:19.226953+00 | 2026-02-28 19:04:08.894085+00 |            | ["trigger", "dispatch", "build-queue"] | docker.io/aqls/agent-chassis | v1.0.817  |         | {"limits": {"cpu": "250m", "memory": "256Mi"}, "requests": {"cpu": "50m", "memory": "128Mi"}} | {"error": "system.errors.{type}", "process": "system.agent.{type}.process", "response": "system.responses.{type}"} | {"port": 8080, "liveness_path": "/health", "readiness_path": "/ready", "initial_delay_seconds": 15} | []       |       1 |                     |               |                       |                        | {"fallback_to_self": true, "prefer_delegation": true} | coordinator    | experimental | ["build", "trigger", "pipeline"] | {}                     |           0 | f           | {"optional": [], "required": [], "description": "No input required. Reads from build_queue and site_work_items tables."} | {"produces": {"seed_result": "How many entries were seeded", "dispatch_result": "Result from the dispatch loop (if triggered)"}}
(1 row)

-- remove domain filter

                    -- Migration: 067_dispatch_remove_domain_filter.sql
--
-- Problem: build-dispatch-loop and build-pipeline-trigger filter
-- item_domain = 'build', so work items with domain 'design' or 'content'
-- (e.g. needs_design, forced_text_colors) are never dispatched.
--
-- Fix: Remove the domain filter. The dispatch loop processes items by
-- priority regardless of domain. Each item's handler_agent field ensures
-- the correct agent is spawned.
--
-- Affected agents:
--   build-dispatch-loop    — load_next_item, check_remaining steps
--   build-pipeline-trigger — find_dispatchable_site raw SQL query
--
-- NOT changed:
--   site-work-orchestrator — its load_work_items filters by domain AND
--     handler_agent for a specific purpose (initial page content writing).
--     Its load_fix_items step already has no domain filter.

-- ============================================================================
-- 1. build-dispatch-loop: load_next_item — remove item_domain from config
-- ============================================================================
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,load_next_item,config}',
        (default_config->'workflow'->'steps'->'load_next_item'->'config') - 'item_domain'
                     ),
    updated_at = NOW()
WHERE type = 'build-dispatch-loop'
  AND is_active = true
  AND default_config->'workflow'->'steps'->'load_next_item'->'config' ? 'item_domain';

-- ============================================================================
-- 2. build-dispatch-loop: check_remaining — remove item_domain from config
-- ============================================================================
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,check_remaining,config}',
        (default_config->'workflow'->'steps'->'check_remaining'->'config') - 'item_domain'
                     ),
    updated_at = NOW()
WHERE type = 'build-dispatch-loop'
  AND is_active = true
  AND default_config->'workflow'->'steps'->'check_remaining'->'config' ? 'item_domain';

-- ============================================================================
-- 3. build-pipeline-trigger: find_dispatchable_site — remove domain filter
--    from the raw SQL query string.
--
--    Old:  ... AND wi.domain = 'build' AND wi.attempt_count ...
--    New:  ... AND wi.attempt_count ...
-- ============================================================================
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,find_dispatchable_site,config,query}',
        to_jsonb(
                REPLACE(
                        default_config->'workflow'->'steps'->'find_dispatchable_site'->'config'->>'query',
                        'AND wi.domain = ''build'' ',
                        ''
                )
        )
                     ),
    updated_at = NOW()
WHERE type = 'build-pipeline-trigger'
  AND is_active = true
  AND default_config->'workflow'->'steps'->'find_dispatchable_site'->'config'->>'query'
    LIKE '%wi.domain%';

-- ============================================================================
-- Verify
-- ============================================================================
-- Run after applying:
--
-- SELECT type,
--        default_config->'workflow'->'steps'->'load_next_item'->'config' as load_config,
--        default_config->'workflow'->'steps'->'check_remaining'->'config' as remaining_config
-- FROM agent_definitions WHERE type = 'build-dispatch-loop' AND is_active = true;
--
-- SELECT type,
--        default_config->'workflow'->'steps'->'find_dispatchable_site'->'config'->>'query' as query
-- FROM agent_definitions WHERE type = 'build-pipeline-trigger' AND is_active = true;


-- backup
bb291e23-ca22-4e85-85d1-df7e24b70b42 | build-pipeline-trigger | Build Pipeline Trigger | Heartbeat agent: seeds build queue entries, finds sites with pending work items, triggers dispatch loop. One site per invocation. | orchestrator | {"workflow": {"steps": {"complete": {"action": "complete_workflow", "config": {"output_fields": ["seed_result", "dispatch_result"]}, "description": "Pipeline trigger complete"}, "seed_queue": {"action": "seed_build_queue", "config": {"max_entries": 10}, "next_step": "find_dispatchable_site", "description": "Process queued entries from build_queue — creates site records and initial work items", "output_field": "seed_result"}, "call_dispatch": {"action": "call_agent", "config": {"agent_type": "build-dispatch-loop", "error_step": "complete_idle", "target_role": "dispatcher", "input_mapping": {"domain": "dispatchable.domain", "site_id": "dispatchable.site_id"}, "timeout_seconds": 900}, "next_step": "complete", "description": "Call dispatch loop — it processes one item and chains for the rest", "output_field": "dispatch_result"}, "complete_idle": {"action": "complete_workflow", "config": {"output_fields": ["seed_result"], "success_message": "No sites need dispatching"}, "description": "Nothing to dispatch this cycle"}, "check_has_site": {"action": "conditional", "config": {"condition": "dispatchable.count > 0", "else_step": "complete_idle", "then_step": "spawn_dispatch"}, "description": "Check if there is a site needing dispatch"}, "spawn_dispatch": {"action": "spawn_agent", "config": {"role": "dispatcher", "agent_type": "build-dispatch-loop", "error_step": "complete_idle"}, "next_step": "call_dispatch", "description": "Spawn dispatch loop for the site", "output_field": "dispatch_spawned"}, "find_dispatchable_site": {"action": "query_database", "config": {"query": "SELECT DISTINCT ON (wi.site_id) wi.site_id::text, s.domain FROM site_work_items wi JOIN sites s ON s.id = wi.site_id WHERE wi.status IN ('triaged', 'approved') AND wi.attempt_count < wi.max_attempts AND NOT EXISTS (SELECT 1 FROM site_work_items active WHERE active.site_id = wi.site_id AND active.status = 'claimed') ORDER BY wi.site_id, wi.priority ASC LIMIT 1", "output_format": "object"}, "next_step": "check_has_site", "description": "Find a site with pending build items that has no actively claimed items", "output_field": "dispatchable"}}, "start_step": "seed_queue"}, "processing_mode": "orchestrator", "timeout_seconds": 1200} | t         | 2026-02-24 19:33:19.226953+00 | 2026-03-09 19:52:41.034875+00 |            | ["trigger", "dispatch", "build-queue"] | docker.io/aqls/agent-chassis | v1.0.851  |         | {"limits": {"cpu": "250m", "memory": "256Mi"}, "requests": {"cpu": "50m", "memory": "128Mi"}} | {"error": "system.errors.{type}", "process": "system.agent.{type}.process", "response": "system.responses.{type}"} | {"port": 8080, "liveness_path": "/health", "readiness_path": "/ready", "initial_delay_seconds": 15} | []       |       1 |                     |               |                       |                        | {"fallback_to_self": true, "prefer_delegation": true} | coordinator    | experimental | ["build", "trigger", "pipeline"] | {}                     |           0 | f           | {"optional": [], "required": [], "description": "No input required. Reads from build_queue and site_work_items tables."} | {"produces": {"seed_result": "How many entries were seeded", "dispatch_result": "Result from the dispatch loop (if triggered)"}}

-- Also reduce the trigger's call_dispatch timeout (one item doesn't need 15 min)
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,call_dispatch,config,timeout_seconds}',
        '360'::jsonb
                     ),
    updated_at = NOW()
WHERE type = 'build-pipeline-trigger' AND deleted_at IS NULL;

---

-- increase parallelisation

-- Allow more concurrent dispatches
UPDATE scheduled_tasks
SET max_concurrent = 4, interval_seconds = 60
WHERE name = 'build-pipeline-trigger';

-- The pre_query only returns 1 site — change to return more
UPDATE scheduled_tasks
SET pre_query = '
SELECT COUNT(*)::text as pending_sites
FROM sites s
WHERE EXISTS (
    SELECT 1 FROM site_work_items wi
    WHERE wi.site_id = s.id
      AND wi.status = ''triaged''
      AND wi.domain = ''build''
      AND wi.attempt_count < wi.max_attempts
)
HAVING COUNT(*) > 0
'
WHERE name = 'build-pipeline-trigger';


---

-- 1. Reset stale claimed items
UPDATE site_work_items
SET status = 'triaged', claimed_by = NULL, claimed_at = NULL
WHERE status = 'claimed' AND domain = 'build'
    RETURNING id, item_type, (SELECT domain FROM sites WHERE id = site_id);

-- 2. Wire notify_scheduler into complete_idle path too
-- complete_idle currently goes straight to complete_workflow
-- Change it to: complete_idle → notify_scheduler → then complete_idle does its thing
--
-- Actually, the simpler fix: make complete_idle go through notify_scheduler first
-- by inserting notify_scheduler_idle before complete_idle

UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,notify_scheduler_idle}',
        '{
            "action": "query_database",
            "config": {
                "query": "UPDATE scheduled_tasks SET last_completed_at = NOW() WHERE name = ''build-pipeline-trigger''",
                "output_format": "object"
            },
            "next_step": "complete_idle",
            "description": "Tell scheduler this execution finished (idle path)",
            "output_field": "scheduler_notified_idle"
        }'::jsonb
                     ),
    updated_at = NOW()
WHERE type = 'build-pipeline-trigger' AND deleted_at IS NULL;

-- Point check_has_site else_step to notify_scheduler_idle instead of complete_idle
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,check_has_site,config,else_step}',
        '"notify_scheduler_idle"'::jsonb
                     ),
    updated_at = NOW()
WHERE type = 'build-pipeline-trigger' AND deleted_at IS NULL;

-- Also check: does the success path (complete) go through notify_scheduler?
SELECT
    default_config->'workflow'->'steps'->'check_has_site'->'config'->>'else_step' as idle_path,
    default_config->'workflow'->'steps'->'spawn_dispatch'->>'next_step' as dispatch_path,
    default_config->'workflow'->'steps'->'call_dispatch'->>'next_step' as after_dispatch,
    default_config->'workflow'->'steps'->'notify_scheduler'->>'next_step' as notify_goes_to
FROM agent_definitions
WHERE type = 'build-pipeline-trigger' AND deleted_at IS NULL;

-- 3. Check claimed-item-timeout
SELECT name, enabled, fire_message, last_triggered_at,
       NOW() - last_triggered_at as since_last
FROM scheduled_tasks
WHERE name = 'claimed-item-timeout';


----

-- catch truly stale items

-- 1. Fix claimed-item-timeout: reset stale claims regardless of attempt count
-- Items at max attempts go to 'failed' instead of back to 'triaged'
UPDATE scheduled_tasks
SET pre_query = '
WITH reset AS (
    UPDATE site_work_items
    SET status = CASE
            WHEN attempt_count + 1 >= max_attempts THEN ''failed''
            ELSE ''triaged''
        END,
        claimed_by = NULL,
        claimed_at = NULL,
        attempt_count = attempt_count + 1,
        error = CASE
            WHEN attempt_count + 1 >= max_attempts THEN ''Claim timed out (attempts exhausted)''
            ELSE NULL
        END
    WHERE status = ''claimed''
      AND claimed_at < NOW() - INTERVAL ''10 minutes''
    RETURNING id, item_type, handler_agent, status
)
SELECT COUNT(*)::text as reset_count,
       string_agg(DISTINCT handler_agent, '', '') as agents
FROM reset
HAVING COUNT(*) > 0
'
WHERE name = 'claimed-item-timeout';

-- 2. Add notify_scheduler to the success path of build-pipeline-trigger
-- Currently: call_dispatch → complete
-- Need:      call_dispatch → notify_scheduler → complete

-- First add the notify_scheduler step (it doesn't exist yet)
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,notify_scheduler}',
        '{
            "action": "query_database",
            "config": {
                "query": "UPDATE scheduled_tasks SET last_completed_at = NOW() WHERE name = ''build-pipeline-trigger''",
                "output_format": "object"
            },
            "next_step": "complete",
            "description": "Tell scheduler this execution finished (success path)",
            "output_field": "scheduler_notified"
        }'::jsonb
                     ),
    updated_at = NOW()
WHERE type = 'build-pipeline-trigger' AND deleted_at IS NULL;

-- Point call_dispatch to notify_scheduler instead of complete
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,call_dispatch,next_step}',
        '"notify_scheduler"'::jsonb
                     ),
    updated_at = NOW()
WHERE type = 'build-pipeline-trigger' AND deleted_at IS NULL;

-- Also point call_dispatch error_step through notify too
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,call_dispatch,config,error_step}',
        '"notify_scheduler_idle"'::jsonb
                     ),
    updated_at = NOW()
WHERE type = 'build-pipeline-trigger' AND deleted_at IS NULL;

-- 3. Force last_completed_at to now so the no-refire guard unblocks immediately
UPDATE scheduled_tasks
SET last_completed_at = NOW()
WHERE name = 'build-pipeline-trigger';

-- Verify all paths
SELECT
    default_config->'workflow'->'steps'->'check_has_site'->'config'->>'else_step' as idle_path,
    default_config->'workflow'->'steps'->'call_dispatch'->>'next_step' as success_path,
    default_config->'workflow'->'steps'->'call_dispatch'->'config'->>'error_step' as error_path,
    default_config->'workflow'->'steps'->'notify_scheduler'->>'next_step' as notify_to,
    default_config->'workflow'->'steps'->'notify_scheduler_idle'->>'next_step' as notify_idle_to
FROM agent_definitions
WHERE type = 'build-pipeline-trigger' AND deleted_at IS NULL;

-- Expected:
-- idle_path:      notify_scheduler_idle
-- success_path:   notify_scheduler
-- error_path:     notify_scheduler_idle
-- notify_to:      complete
-- notify_idle_to: complete_idle