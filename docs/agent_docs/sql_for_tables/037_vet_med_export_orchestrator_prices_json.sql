-- med_export_agent_setup.sql
-- Configurable JSON export agents. One action, multiple sites via different configs.
--
-- Usage examples:
--   All prices to vetcomparison.co.uk:
--     {"agent_type":"med-json-exporter","input_data":{"domain":"vetcomparison.co.uk"}}
--
--   Dog-only prices to a hypothetical dog meds site:
--     {"agent_type":"med-json-exporter","input_data":{"domain":"dogmeds.co.uk","filters":{"species":["dog"]}}}
--
--   Single retailer export:
--     {"agent_type":"med-json-exporter","input_data":{"domain":"petdrugsonline-compare.co.uk","filters":{"retailers":["pet_drugs_online"]}}}

-- ============================================================================
-- 1. Export worker agent (generic — config comes from input_data)
-- ============================================================================

INSERT INTO agent_definitions (
    type, display_name, description, category,
    image_repository, image_tag,
    is_active, status,
    default_config,
    topics, env_vars,
    idle_timeout_seconds
) VALUES (
             'med-json-exporter',
             'Med JSON Exporter',
             'Exports current medicine prices as JSON and commits to a site git repo. Configurable: domain, filters (species/category/retailer), output files, data path.',
             'business_intel',
             'docker.io/aqls/agent-chassis', 'v1.0.942',
             true, 'active',
             '{
                 "workflow": {
                     "start_step": "export_json",
                     "steps": {
                         "export_json": {
                             "action": "med_export_json",
                             "config": {
                                 "domain": "vetcomparison.co.uk",
                                 "repo_name": "sites",
                                 "data_path": "data",
                                 "commit_message_prefix": "Update medicine prices",
                                 "outputs": {
                                     "index": true,
                                     "full": true,
                                     "by_letter": true,
                                     "metadata": true
                                 }
                             },
                             "next_step": "complete",
                             "description": "Query prices, build JSON, commit to git"
                         },
                         "complete": {
                             "action": "complete_workflow",
                             "description": "Export complete"
                         }
                     }
                 },
                 "processing_mode": "task"
             }'::jsonb,
             '["system.agent.business-intel.requests", "system.agent.business-intel.responses"]'::jsonb,
             '[]'::jsonb,
             0
         ) ON CONFLICT (type, version)
DO UPDATE SET
    default_config = EXCLUDED.default_config,
                  description = EXCLUDED.description,
                  updated_at = NOW();


-- ============================================================================
-- 2. Export orchestrator (spawns temporary pod)
-- ============================================================================

INSERT INTO agent_definitions (
    type, display_name, description, category,
    image_repository, image_tag,
    is_active, status,
    default_config,
    topics, env_vars,
    idle_timeout_seconds
) VALUES (
             'med-export-orchestrator',
             'Med Export Orchestrator',
             'Spawns a temporary pod to export medicine prices as JSON to a configured site.',
             'business_intel',
             'docker.io/aqls/agent-chassis', 'v1.0.942',
             true, 'active',
             '{
                 "workflow": {
                     "start_step": "spawn_exporter",
                     "steps": {
                         "spawn_exporter": {
                             "action": "spawn_agent",
                             "config": {
                                 "agent_type": "med-json-exporter",
                                 "role": "json_exporter"
                             },
                             "next_step": "call_exporter",
                             "output_field": "exporter_spawn",
                             "description": "Spawn a temporary pod for JSON export"
                         },
                         "call_exporter": {
                             "action": "call_agent",
                             "config": {
                                 "agent_type": "med-json-exporter",
                                 "target_role": "json_exporter",
                                 "input_mapping": {
                                     "input_data": "input_data"
                                 },
                                 "timeout_seconds": 120
                             },
                             "next_step": "complete",
                             "output_field": "export_result",
                             "description": "Send export work to spawned pod and wait"
                         },
                         "complete": {
                             "action": "complete_workflow",
                             "description": "Export complete"
                         }
                     }
                 }
             }'::jsonb,
             '["system.agent.business-intel.requests", "system.agent.business-intel.responses"]'::jsonb,
             '[]'::jsonb,
             0
         ) ON CONFLICT (type, version)
DO UPDATE SET
    default_config = EXCLUDED.default_config,
                  description = EXCLUDED.description,
                  updated_at = NOW();


-- ============================================================================
-- 3. Scheduled task for vetcomparison.co.uk (disabled initially)
-- ============================================================================

INSERT INTO scheduled_tasks (
    name, description,
    target_agent_type, target_topic,
    interval_seconds, enabled,
    input_data,
    concurrency_group
) VALUES (
             'med-export-json',
             'Export all medicine prices to vetcomparison.co.uk',
             'med-json-exporter',
             'system.agent.business-intel.requests',
             172800,
             false,
             -- PAYLOAD ONLY. The scheduler's fireTrigger (cmd/scheduler/main.go) supplies
             -- action="orchestrate" and config.agent_type (from target_agent_type) and wraps this
             -- column as input_data itself. Nesting an {action,config,input_data} envelope here
             -- double-wraps it and the payload never reaches the action — see bugs_closed/054.
             -- DOMAIN DELIBERATELY BLANK (fail-closed). Per the vetcomparison RUNBOOK standing
             -- safety rail: "we do NOT own vetcomparison.co.uk — never reintroduce it", and DB
             -- configs are blanked so nothing can export to a wrong domain. This med exporter is
             -- ALSO superseded by the generic directory_export_json (Phase 2). Only the owner sets
             -- a real domain, deliberately, when/if enabling this task — one at a time.
             '{"domain":""}'::jsonb,
             'med-collection'
         ) ON CONFLICT (name) DO NOTHING;


-- ============================================================================
-- Verify
-- ============================================================================
SELECT type, display_name, status
FROM agent_definitions
WHERE type LIKE 'med-%export%'
ORDER BY type;

