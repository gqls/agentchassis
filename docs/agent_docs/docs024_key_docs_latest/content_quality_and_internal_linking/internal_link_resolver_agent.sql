-- internal-link-resolver — agent_definitions row.
--
-- A called sub-agent (like research-agent): workflow lives in
-- default_config.workflow, processing_mode "task" inside default_config
-- (agent_definitions has NO processing_mode column). Spawned by
-- page-content-writer (role link_resolver) and called once per page with the
-- section plan; returns the sections with CTA destinations resolved from real
-- pages, plus an unresolved list. Thin workflow — the logic is in the
-- resolve_internal_links Go action.
--
-- Apply AFTER the chassis image containing resolve_internal_links_action.go
-- (+ its registry entry) is rolled, and set image_tag below to that batch's tag.
-- Idempotent: guarded by NOT EXISTS.

INSERT INTO agent_definitions (
    id, type, display_name, description, category, agent_category,
    default_config,
    is_active, status, version,
    capabilities, image_repository, image_tag,
    resources, topics, health_config, env_vars,
    delegation_preferences, domain_tags, briefing_questionnaire,
    input_contract, output_contract, idle_timeout_seconds
)
SELECT
    gen_random_uuid(),
    'internal-link-resolver',
    'Internal Link Resolver',
    'Resolves intent-appropriate internal link destinations (hero/CTA buttons) for a page from the real pages set — never a fabricated target. Augments section resolved_data at build time; reports unresolved CTAs for review. v1 rules are deterministic (top content hubs); the agent boundary allows an LLM intent-matching upgrade without changing callers.',
    'specialist',
    'specialist',
    $cfg$
    {
        "workflow": {
            "steps": {
                "resolve_links": {
                    "action": "resolve_internal_links",
                    "config": {
                        "site_id": "input_data.site_id",
                        "page_type": "input_data.page_type",
                        "page_name": "input_data.page_name",
                        "sections": "input_data.sections"
                    },
                    "next_step": "complete",
                    "description": "Augment CTA-bearing sections with destinations resolved from real pages",
                    "output_field": "link_resolution"
                },
                "complete": {
                    "action": "complete_workflow",
                    "config": {
                        "output": {
                            "sections_ready": "link_resolution.sections_ready",
                            "unresolved": "link_resolution.unresolved"
                        }
                    },
                    "description": "Return augmented sections and unresolved CTA report"
                }
            },
            "start_step": "resolve_links"
        },
        "processing_mode": "task",
        "timeout_seconds": 120
    }
    $cfg$::jsonb,
    true,
    'active',
    1,
    '["internal-linking", "link-resolution"]'::jsonb,
    'docker.io/aqls/agent-chassis',
    'v1.0.1060',
    '{"limits": {"cpu": "500m", "memory": "512Mi"}, "requests": {"cpu": "100m", "memory": "256Mi"}}'::jsonb,
    '{"error": "system.errors.{type}", "process": "system.agent.{type}.process", "response": "system.responses.{type}"}'::jsonb,
    '{"port": 8080, "liveness_path": "/health", "readiness_path": "/ready", "initial_delay_seconds": 15}'::jsonb,
    '[]'::jsonb,
    '{"fallback_to_self": true, "prefer_delegation": true}'::jsonb,
    '[]'::jsonb,
    '{}'::jsonb,
    '{"required": ["site_id", "sections"], "optional": ["page_type", "page_name"], "description": "site_id and the section_plan sections_ready list; page_type/page_name refine destination choice (own-hub exclusion)."}'::jsonb,
    '{"produces": {"sections_ready": "Section list with CTA destinations written into resolved_data", "unresolved": "CTA fields with no resolvable real-page destination"}}'::jsonb,
    180
WHERE NOT EXISTS (
    SELECT 1 FROM agent_definitions WHERE type = 'internal-link-resolver'
);

-- Verify
SELECT type, agent_category, is_active, status,
       default_config->'workflow'->>'start_step'            AS start_step,
       default_config->>'processing_mode'                   AS processing_mode,
       jsonb_object_keys(default_config->'workflow'->'steps') AS steps
FROM agent_definitions
WHERE type = 'internal-link-resolver';
