-- ============================================================================
-- agent_definitions — intent-collection-orchestrator + intent-collector (P4)
-- Infra fields (image/tag/resources/health/topics/delegation) mirror the LIVE
-- med-export-orchestrator / med-json-exporter pair verbatim — the canonical
-- scheduler-triggered wrapper + task-worker. Only type/display_name/description/
-- default_config differ.
--
-- CONFIRM one thing: topics are copied from med-export (business-intel). The
-- scheduler reaches the orchestrator via target_topic=system.agent.generic.requests
-- (set in intent_events_migration.sql) by agent_type, independent of the topics
-- field — same as med-export-json drives med-export-orchestrator. If you keep a
-- dedicated data/maintenance topic taxonomy, swap both topics here and the
-- scheduled task's target_topic together.
--
-- VERSIONING: agent_definitions is UNIQUE on (type, version), NOT type alone —
-- the table keeps versioned rows. These INSERTs omit version (defaults to 1) and
-- use ON CONFLICT (type, version) DO NOTHING for idempotency on v1. To revise a
-- workflow later, INSERT a new row with the next version (and flip is_active),
-- or UPDATE the v1 row in place — don't expect ON CONFLICT (type) to work.
-- ============================================================================

-- Parent: thin wrapper. spawn → call → complete. No substantive work in-chassis.
INSERT INTO agent_definitions
    (type, display_name, description, category, default_config, is_active,
     capabilities, image_repository, image_tag, resources, topics, health_config,
     delegation_preferences, status, idle_timeout_seconds)
VALUES (
    'intent-collection-orchestrator',
    'Intent Collection Orchestrator',
    'Spawns a temporary pod to pull intent events + visit stats from VM-hosted backend sites into intent_events / intent_site_stats.',
    'business_intel',
    '{"workflow": {"start_step": "spawn_collector", "steps": {
        "spawn_collector": {"action": "spawn_agent", "config": {"role": "collector", "agent_type": "intent-collector"}, "next_step": "call_collector", "description": "Spawn a temporary pod for intent collection", "output_field": "collector_spawn"},
        "call_collector": {"action": "call_agent", "config": {"agent_type": "intent-collector", "target_role": "collector", "timeout_seconds": 280}, "next_step": "complete", "description": "Run the collection in the spawned pod and wait", "output_field": "collect_result"},
        "complete": {"action": "complete_workflow", "description": "Collection complete"}
    }}}'::jsonb,
    true,
    '[]'::jsonb,
    'docker.io/aqls/agent-chassis',
    'v1.0.1063',
    '{"limits": {"cpu": "500m", "memory": "1Gi"}, "requests": {"cpu": "100m", "memory": "256Mi"}}'::jsonb,
    '["system.agent.business-intel.requests", "system.agent.business-intel.responses"]'::jsonb,
    '{"port": 8080, "liveness_path": "/health", "readiness_path": "/ready", "initial_delay_seconds": 30}'::jsonb,
    '{"fallback_to_self": true, "prefer_delegation": true}'::jsonb,
    'active',
    0
)
ON CONFLICT (type, version) DO NOTHING;

-- Child: task worker. One step → collect_intent_events (self-queries all VM
-- backend sites, pulls /events + /stats). processing_mode:"task" like the
-- med-json-exporter worker.
INSERT INTO agent_definitions
    (type, display_name, description, category, default_config, is_active,
     capabilities, image_repository, image_tag, resources, topics, health_config,
     delegation_preferences, status, idle_timeout_seconds)
VALUES (
    'intent-collector',
    'Intent Collector',
    'Pulls new intent events and visit/event counters from every VM-hosted backend site''s /events and /stats endpoints into intent_events and intent_site_stats.',
    'business_intel',
    '{"workflow": {"start_step": "collect", "steps": {
        "collect": {"action": "collect_intent_events", "next_step": "complete", "description": "Pull /events + /stats from all VM backend sites and upsert"},
        "complete": {"action": "complete_workflow", "description": "Collection complete"}
    }}, "processing_mode": "task"}'::jsonb,
    true,
    '[]'::jsonb,
    'docker.io/aqls/agent-chassis',
    'v1.0.1063',
    '{"limits": {"cpu": "500m", "memory": "1Gi"}, "requests": {"cpu": "100m", "memory": "256Mi"}}'::jsonb,
    '["system.agent.business-intel.requests", "system.agent.business-intel.responses"]'::jsonb,
    '{"port": 8080, "liveness_path": "/health", "readiness_path": "/ready", "initial_delay_seconds": 30}'::jsonb,
    '{"fallback_to_self": true, "prefer_delegation": true}'::jsonb,
    'active',
    0
)
ON CONFLICT (type, version) DO NOTHING;
