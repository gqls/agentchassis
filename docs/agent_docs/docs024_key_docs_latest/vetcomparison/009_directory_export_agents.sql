-- ============================================================================
-- 009_directory_export_agents.sql
-- ============================================================================
-- Agent pair + disabled scheduled task for the generic directory exporter
-- (directory_export_json). Modelled on the med-json-exporter pair (037).
--
-- ⚠️ image_tag: set to the FIRST chassis build containing
-- directory_export_json (registered 2026-07-16; not yet deployed at seed
-- time — current fleet tag v1.0.1126 does NOT contain it). Update image_tag
-- before enabling the task. Task is seeded DISABLED.
-- ============================================================================

BEGIN;

INSERT INTO agent_definitions (
    type, display_name, description, category,
    image_repository, image_tag, is_active, status,
    default_config, topics, env_vars, idle_timeout_seconds
) VALUES (
    'directory-json-exporter',
    'Directory JSON Exporter',
    'Exports a vertical''s verified business directory, k-anonymous price aggregates, and consented/attributed price lists as JSON to a site repo. Domain+vertical are required config; attributed prices are fail-closed off unless the site config enables them.',
    'business_intel',
    'docker.io/aqls/agent-chassis', 'v1.0.1126',
    true, 'active',
    '{
        "workflow": {
            "start_step": "export_json",
            "steps": {
                "export_json": {
                    "action": "directory_export_json",
                    "config": {},
                    "next_step": "complete",
                    "description": "Query directory + prices, build JSON, commit to git"
                },
                "complete": { "action": "complete_workflow", "description": "Export complete" }
            }
        },
        "processing_mode": "task"
    }'::jsonb,
    '["system.agent.business-intel.requests", "system.agent.business-intel.responses"]'::jsonb,
    '[]'::jsonb, 0
) ON CONFLICT (type, version) DO UPDATE SET
    default_config = EXCLUDED.default_config,
    description = EXCLUDED.description, updated_at = NOW();

INSERT INTO agent_definitions (
    type, display_name, description, category,
    image_repository, image_tag, is_active, status,
    default_config, topics, env_vars, idle_timeout_seconds
) VALUES (
    'directory-export-orchestrator',
    'Directory Export Orchestrator',
    'Spawns a temporary pod to run the generic directory/price JSON export for a configured site.',
    'business_intel',
    'docker.io/aqls/agent-chassis', 'v1.0.1126',
    true, 'active',
    '{
        "workflow": {
            "start_step": "spawn_exporter",
            "steps": {
                "spawn_exporter": {
                    "action": "spawn_agent",
                    "config": { "agent_type": "directory-json-exporter", "role": "json_exporter" },
                    "next_step": "call_exporter",
                    "output_field": "exporter_spawn",
                    "description": "Spawn a temporary pod for JSON export"
                },
                "call_exporter": {
                    "action": "call_agent",
                    "config": {
                        "agent_type": "directory-json-exporter",
                        "target_role": "json_exporter",
                        "input_mapping": { "input_data": "input_data" },
                        "timeout_seconds": 120
                    },
                    "next_step": "complete",
                    "output_field": "export_result",
                    "description": "Send export work to spawned pod and wait"
                },
                "complete": { "action": "complete_workflow", "description": "Export complete" }
            }
        }
    }'::jsonb,
    '["system.agent.business-intel.requests", "system.agent.business-intel.responses"]'::jsonb,
    '[]'::jsonb, 0
) ON CONFLICT (type, version) DO UPDATE SET
    default_config = EXCLUDED.default_config,
    description = EXCLUDED.description, updated_at = NOW();

-- vetcomparison.uk task — DISABLED until the image containing the action is
-- deployed and image_tag above is updated. Owner decisions 2026-07-16 baked
-- into config: attributed on, min_n 3, live filename preserved.
INSERT INTO scheduled_tasks (
    name, description, target_agent_type, target_topic,
    interval_seconds, enabled, input_data, concurrency_group
) VALUES (
    'directory-export-json',
    'Export vetcomparison.uk directory + price data',
    'directory-json-exporter',
    'system.agent.business-intel.requests',
    172800, false,
    '{"action":"orchestrate","config":{"agent_type":"directory-json-exporter"},
      "input_data":{
        "vertical":"veterinary","domain":"vetcomparison.uk",
        "repo_name":"sites","data_path":"data",
        "business_type_ilike":"%vet%",
        "commit_message_prefix":"Update vetcomparison.uk directory data",
        "outputs":{"directory":true,"directory_filename":"vet-full-index.json",
                   "aggregates":{"enabled":true,"min_n":3},
                   "claimed_prices":true,"attributed_prices":true}}}'::jsonb,
    'vetcomparison-uk-exports'
) ON CONFLICT (name) DO NOTHING;

SELECT type, status, image_tag FROM agent_definitions WHERE type LIKE 'directory-%' ORDER BY type;
SELECT name, enabled FROM scheduled_tasks WHERE name = 'directory-export-json';

COMMIT;
