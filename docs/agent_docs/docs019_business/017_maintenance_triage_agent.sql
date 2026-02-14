-- ============================================================================
-- MIGRATION: maintenance_queue table + maintenance-triage agent
--
-- Apply with:
--   kubectl -n ai-persona-system exec -it deploy/api-server -- \
--     psql -U clients_user -d clients_db -f /tmp/migration.sql
--
-- Or copy-paste sections into psql.
-- ============================================================================

-- ============================================================================
-- PART 1: maintenance_queue table and helper functions
-- ============================================================================

CREATE TABLE IF NOT EXISTS maintenance_queue (
                                                 id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    site_id         UUID NOT NULL REFERENCES sites(id),

    -- Task classification
    task_type       TEXT NOT NULL,          -- 'page_rebuild', 'css_update', 'nav_fix', 'link_repair', 'content_refresh'
    priority        INT NOT NULL DEFAULT 5, -- 1=urgent, 5=normal, 10=low
    reason          TEXT,                   -- 'stale_content', 'missing_page', 'broken_links', 'manual_request', 'automated_scan'

-- Task payload
-- For page_rebuild: { "pages": ["use-cases", "privacy"], "detected_at": "..." }
-- For css_update:   { "force_regenerate": true }
-- For nav_fix:      { "remove_stale": true, "add_missing": true }
    payload         JSONB NOT NULL DEFAULT '{}'::jsonb,

    -- Execution tracking
    status          TEXT NOT NULL DEFAULT 'pending',  -- pending, claimed, in_progress, complete, failed, cancelled
    claimed_by      TEXT,                              -- agent_id that picked this up
    claimed_at      TIMESTAMPTZ,
    started_at      TIMESTAMPTZ,
    completed_at    TIMESTAMPTZ,

    -- Result
    result          JSONB,                             -- specialist agent output
    error_message   TEXT,
    retry_count     INT NOT NULL DEFAULT 0,
    max_retries     INT NOT NULL DEFAULT 3,

    -- Metadata
    requested_by    TEXT,                               -- 'maintenance-triage', 'manual', 'cron'
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
    );

-- Indexes for queue operations
CREATE INDEX IF NOT EXISTS idx_maintenance_queue_pending
    ON maintenance_queue (priority ASC, created_at ASC)
    WHERE status = 'pending';

CREATE INDEX IF NOT EXISTS idx_maintenance_queue_site
    ON maintenance_queue (site_id, status);

CREATE INDEX IF NOT EXISTS idx_maintenance_queue_type
    ON maintenance_queue (task_type, status);


-- Claim function: atomically claim the next pending task (single row)
-- Used by: claim_maintenance_task PL/pgSQL (not used by Go code directly,
--          but useful for manual debugging and future use)
CREATE OR REPLACE FUNCTION claim_maintenance_task(
    p_task_type TEXT,
    p_agent_id TEXT
) RETURNS maintenance_queue AS $$
DECLARE
claimed_task maintenance_queue;
BEGIN
UPDATE maintenance_queue
SET status = 'claimed',
    claimed_by = p_agent_id,
    claimed_at = NOW(),
    updated_at = NOW()
WHERE id = (
    SELECT id FROM maintenance_queue
    WHERE status = 'pending'
      AND task_type = p_task_type
      AND retry_count < max_retries
    ORDER BY priority ASC, created_at ASC
    LIMIT 1
    FOR UPDATE SKIP LOCKED
            )
            RETURNING * INTO claimed_task;

RETURN claimed_task;
END;
$$ LANGUAGE plpgsql;


-- Complete function: mark task as done
CREATE OR REPLACE FUNCTION complete_maintenance_task(
    p_task_id UUID,
    p_result JSONB DEFAULT '{}'::jsonb
) RETURNS VOID AS $$
BEGIN
UPDATE maintenance_queue
SET status = 'complete',
    result = p_result,
    completed_at = NOW(),
    updated_at = NOW()
WHERE id = p_task_id;
END;
$$ LANGUAGE plpgsql;


-- Fail function: increment retry, revert to pending or mark failed
CREATE OR REPLACE FUNCTION fail_maintenance_task(
    p_task_id UUID,
    p_error TEXT
) RETURNS VOID AS $$
BEGIN
UPDATE maintenance_queue
SET retry_count = retry_count + 1,
    error_message = p_error,
    status = CASE
                 WHEN retry_count + 1 >= max_retries THEN 'failed'
                 ELSE 'pending'  -- back in queue for retry
        END,
    claimed_by = NULL,
    claimed_at = NULL,
    updated_at = NOW()
WHERE id = p_task_id;
END;
$$ LANGUAGE plpgsql;


-- ============================================================================
-- PART 2: maintenance-triage agent definition
-- ============================================================================

INSERT INTO agent_definitions (
    type, display_name, description, category, default_config,
    is_active, capabilities, image_repository, image_tag,
    topics, health_config, env_vars, version,
    delegation_preferences, agent_category, status, domain_tags,
    input_contract, output_contract
) VALUES (
             'maintenance-triage',
             'Maintenance Triage Agent',
             'Scans deployed sites for maintenance issues (stale pages, missing content, orphan nav items). Populates the maintenance_queue and dispatches specialist agents to resolve issues. Can scan one site or all sites.',
             'orchestrator',
             '{
                 "workflow": {
                     "start_step": "scan_and_queue",
                     "processing_mode": "orchestrator",
                     "timeout_seconds": 1800,
                     "steps": {

                         "scan_and_queue": {
                             "action": "scan_sites_for_maintenance",
                             "config": {
                                 "stale_threshold_days": 30,
                                 "checks": ["stale_pages", "missing_content", "orphan_nav"],
                                 "domain_field": "input_data.domain",
                                 "deduplicate": true
                             },
                             "output_field": "scan_results",
                             "next_step": "check_dry_run",
                             "description": "Scan sites for maintenance issues and insert tasks into queue"
                         },

                         "check_dry_run": {
                             "action": "conditional",
                             "config": {
                                 "condition": "input_data.dry_run == true",
                                 "then_step": "complete_dry_run",
                                 "else_step": "prepare_rebuild_dispatches"
                             },
                             "description": "If dry_run, skip dispatch and report findings only"
                         },

                         "complete_dry_run": {
                             "action": "complete_workflow",
                             "config": {
                                 "output_fields": ["scan_results"],
                                 "success_message": "Dry run complete - tasks queued but not dispatched"
                             },
                             "description": "Complete without dispatching (dry run mode)"
                         },

                         "prepare_rebuild_dispatches": {
                             "action": "prepare_rebuild_dispatches",
                             "config": {
                                 "task_type": "page_rebuild",
                                 "max_tasks": 50,
                                 "flag_pages": true
                             },
                             "output_field": "rebuild_dispatches",
                             "next_step": "check_has_rebuilds",
                             "description": "Claim page_rebuild tasks from queue, flag pages as needs_rebuild, group by site"
                         },

                         "check_has_rebuilds": {
                             "action": "conditional",
                             "config": {
                                 "condition": "rebuild_dispatches.dispatch_count > 0",
                                 "then_step": "spawn_rebuilder",
                                 "else_step": "complete_no_work"
                             },
                             "description": "Check if any sites need page rebuilds"
                         },

                         "complete_no_work": {
                             "action": "complete_workflow",
                             "config": {
                                 "output_fields": ["scan_results", "rebuild_dispatches"],
                                 "success_message": "Scan complete - no dispatches needed"
                             },
                             "description": "Complete when scan found no actionable issues"
                         },

                         "spawn_rebuilder": {
                             "action": "spawn_agent",
                             "config": {
                                 "role": "rebuilder",
                                 "agent_type": "page-rebuild"
                             },
                             "output_field": "rebuilder_agent",
                             "next_step": "rebuild_loop",
                             "description": "Spawn page-rebuild agent"
                         },

                         "rebuild_loop": {
                             "action": "loop",
                             "config": {
                                 "mode": "sequential",
                                 "items_field": "rebuild_dispatches.dispatches",
                                 "item_variable": "current_dispatch",
                                 "max_iterations": 50,
                                 "sub_workflow": {
                                     "start_step": "call_rebuilder",
                                     "steps": {
                                         "call_rebuilder": {
                                             "action": "call_agent",
                                             "config": {
                                                 "agent_type": "page-rebuild",
                                                 "target_role": "rebuilder",
                                                 "input_mapping": {
                                                     "domain": "current_dispatch.domain"
                                                 },
                                                 "timeout_seconds": 900
                                             },
                                             "output_field": "rebuild_result",
                                             "next_step": "mark_dispatch_complete",
                                             "description": "Call page-rebuild for this site"
                                         },

                                         "mark_dispatch_complete": {
                                             "action": "mark_maintenance_complete",
                                             "config": {
                                                 "task_ids_field": "current_dispatch.task_ids",
                                                 "result_field": "rebuild_result"
                                             },
                                             "output_field": "tasks_marked",
                                             "next_step": "complete_dispatch",
                                             "description": "Mark queued tasks as complete"
                                         },

                                         "complete_dispatch": {
                                             "action": "loop_complete",
                                             "description": "Dispatch complete for this site"
                                         }
                                     }
                                 }
                             },
                             "output_field": "dispatches_completed",
                             "next_step": "complete",
                             "description": "Dispatch page-rebuild for each site with pending tasks"
                         },

                         "complete": {
                             "action": "complete_workflow",
                             "config": {
                                 "output_fields": ["scan_results", "rebuild_dispatches", "dispatches_completed"]
                             },
                             "description": "Triage complete"
                         }
                     }
                 }
             }'::jsonb,
             true,
             '[]'::jsonb,
             'docker.io/aqls/agent-chassis',
             'v1.0.770',
             '{"error": "system.errors.{type}", "process": "system.agent.{type}.process", "response": "system.responses.{type}"}'::jsonb,
             '{"port": 8080, "liveness_path": "/health", "readiness_path": "/ready", "initial_delay_seconds": 30}'::jsonb,
             '[]'::jsonb,
             1,
             '{"fallback_to_self": true, "prefer_delegation": true}'::jsonb,
             'coordinator',
             'active',
             '["maintenance", "triage", "scanning"]'::jsonb,
             '{"required": [], "optional": ["domain", "stale_threshold_days", "dry_run"], "description": "Omit domain to scan all deployed sites. Set dry_run=true to scan and queue without dispatching. stale_threshold_days defaults to 30."}'::jsonb,
             '{"produces": {"scan_results": "Issues found per site", "dispatches_completed": "Specialists dispatched and results"}}'::jsonb
         )
    ON CONFLICT (type, version) DO UPDATE SET
    display_name = EXCLUDED.display_name,
                                       description = EXCLUDED.description,
                                       default_config = EXCLUDED.default_config,
                                       image_tag = EXCLUDED.image_tag,
                                       input_contract = EXCLUDED.input_contract,
                                       output_contract = EXCLUDED.output_contract,
                                       status = EXCLUDED.status,
                                       updated_at = NOW();


-- ============================================================================
-- VERIFY
-- ============================================================================

SELECT 'maintenance_queue table' as item,
       (SELECT COUNT(*) FROM information_schema.tables WHERE table_name = 'maintenance_queue') as exists;

SELECT type, display_name, status, agent_category
FROM agent_definitions
WHERE type = 'maintenance-triage';