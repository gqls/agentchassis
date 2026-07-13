-- 058_quality_checks_and_fixers.sql
--
-- 1. quality-discovery-agent — runs the new checks (broken_nav_links, placeholder_contact, generic_theme)
-- 2. nav-link-fixer agent — fixes #slug links in header/footer templates, re-renders
-- 3. Template fix for immediate gaswholesalers.com repair
--
-- Discovery agents detect problems. Fixer agents fix them.
-- The dispatch loop connects them via site_work_items.

-- ============================================================================
-- 1. QUALITY DISCOVERY AGENT
-- Runs content quality checks that require no LLM budget.
-- Complements existing design-discovery-agent and completeness-discovery-agent.
-- ============================================================================

INSERT INTO agent_definitions (
    type, display_name, description, category,
    default_config, is_active, capabilities,
    image_repository, image_tag, resources, topics,
    health_config, env_vars, version,
    delegation_preferences, agent_category, status,
    domain_tags, briefing_questionnaire,
    usage_count, is_snapshot, input_contract, output_contract
) VALUES (
             'quality-discovery-agent',
             'Quality Discovery Agent',
             'Scans sites for quality issues: broken nav links (#slug instead of /page.html), placeholder/fabricated contact info, generic unthemed CSS. All checks are algorithmic — no LLM budget needed. Writes findings to site_work_items.',
             'specialist',
             '{
                 "workflow": {
                     "start_step": "ensure_site_record",
                     "steps": {
                         "ensure_site_record": {
                             "action": "ensure_site_record",
                             "config": {"input_fields": ["site_id", "domain"]},
                             "next_step": "run_checks",
                             "description": "Load site record from domain or site_id",
                             "output_field": "site_record"
                         },
                         "run_checks": {
                             "action": "run_discovery_checks",
                             "config": {
                                 "site_id": "site_record.site_id",
                                 "check_domain": "build",
                                 "checks": ["broken_nav_links", "placeholder_contact", "generic_theme"]
                             },
                             "next_step": "complete",
                             "description": "Run quality checks and write findings to site_work_items",
                             "output_field": "discovery_result"
                         },
                         "complete": {
                             "action": "complete_workflow",
                             "config": {"output_fields": ["discovery_result"]},
                             "description": "Quality discovery complete"
                         }
                     }
                 },
                 "processing_mode": "task",
                 "timeout_seconds": 120
             }'::jsonb,
             true,
             '["discovery", "quality_audit", "nav_check", "contact_check", "theme_check"]'::jsonb,
             'docker.io/aqls/agent-chassis', 'v1.0.803',
             '{"limits": {"cpu": "500m", "memory": "512Mi"}, "requests": {"cpu": "100m", "memory": "256Mi"}}'::jsonb,
             '{"error": "system.errors.{type}", "process": "system.agent.{type}.process", "response": "system.responses.{type}"}'::jsonb,
             '{"port": 8080, "liveness_path": "/health", "readiness_path": "/ready", "initial_delay_seconds": 15}'::jsonb,
             '[]'::jsonb, 1,
             '{"fallback_to_self": true, "prefer_delegation": true}'::jsonb,
             'analyst', 'experimental',
             '["quality", "discovery", "nav", "contact", "theme"]'::jsonb, '{}'::jsonb,
             0, false,
             '{"optional": ["site_id", "domain"], "required": [], "description": "Pass site_id or domain (at least one)."}'::jsonb,
             '{"produces": {"discovery_result": "items_inserted count, findings details, batch_id"}}'::jsonb
         )
    ON CONFLICT (type, version) DO UPDATE SET
    display_name = EXCLUDED.display_name,
                                       description = EXCLUDED.description,
                                       default_config = EXCLUDED.default_config,
                                       is_active = EXCLUDED.is_active,
                                       capabilities = EXCLUDED.capabilities,
                                       agent_category = EXCLUDED.agent_category,
                                       domain_tags = EXCLUDED.domain_tags,
                                       input_contract = EXCLUDED.input_contract,
                                       output_contract = EXCLUDED.output_contract,
                                       updated_at = now();

