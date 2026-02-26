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


-- ============================================================================
-- 2. NAV LINK FIXER AGENT
-- Fixes broken navigation links in site_components (header/footer).
--
-- Strategy:
--   1. Load the site record
--   2. Fix the content_component template if it uses #{{.slug}} pattern
--   3. Force re-render site_components (header + footer)
--   4. If pages are deployed, trigger a rerender of all pages
--      (because header/footer are assembled into every page)
--
-- The template fix is done via update_component_template action.
-- The re-render uses the existing render_site_components action with force=true.
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
             'nav-link-fixer',
             'Nav Link Fixer',
             'Fixes broken navigation links in header/footer. Updates content_component templates to use {{.url}} instead of #{{.slug}}, then force re-renders site_components.',
             'orchestrator',
             '{
                 "workflow": {
                     "start_step": "ensure_site_record",
                     "steps": {

                         "ensure_site_record": {
                             "action": "ensure_site_record",
                             "config": {},
                             "next_step": "fix_nav_templates",
                             "description": "Load site record from database",
                             "output_field": "site_record"
                         },

                         "fix_nav_templates": {
                             "action": "fix_nav_link_templates",
                             "config": {
                                 "site_id": "site_record.site_id",
                                 "patterns": [
                                     {"find": "href=\"#{{.slug}}\"", "replace": "href=\"{{.url}}\""},
                                     {"find": "href=\"#{{ .slug }}\"", "replace": "href=\"{{ .url }}\""},
                                     {"find": "href=\"#{{.name}}\"", "replace": "href=\"{{.url}}\""}
                                 ]
                             },
                             "next_step": "rerender_site_components",
                             "description": "Fix header/footer templates: replace #slug links with proper URLs",
                             "output_field": "template_fix_result"
                         },

                         "rerender_site_components": {
                             "action": "render_site_components",
                             "config": {
                                 "slots": ["header", "footer"],
                                 "domain_field": "site_record.domain",
                                 "site_id_field": "site_record.site_id",
                                 "force_rerender": true
                             },
                             "next_step": "complete",
                             "description": "Force re-render header and footer with corrected templates",
                             "output_field": "rerender_result"
                         },

                         "complete": {
                             "action": "complete_workflow",
                             "config": {
                                 "output_fields": ["template_fix_result", "rerender_result"]
                             },
                             "description": "Nav fix complete. Pages need rerender to pick up new header/footer."
                         }
                     }
                 },
                 "processing_mode": "orchestrator",
                 "timeout_seconds": 120
             }'::jsonb,
             true,
             '["nav-fix", "template-repair", "site-component-rerender"]'::jsonb,
             'docker.io/aqls/agent-chassis', 'v1.0.803',
             '{"limits": {"cpu": "500m", "memory": "512Mi"}, "requests": {"cpu": "100m", "memory": "256Mi"}}'::jsonb,
             '{"error": "system.errors.{type}", "process": "system.agent.{type}.process", "response": "system.responses.{type}"}'::jsonb,
             '{"port": 8080, "liveness_path": "/health", "readiness_path": "/ready", "initial_delay_seconds": 15}'::jsonb,
             '[]'::jsonb, 1,
             '{"fallback_to_self": true, "prefer_delegation": true}'::jsonb,
             'coordinator', 'experimental',
             '["maintenance", "nav", "template-fix"]'::jsonb, '{}'::jsonb,
             0, false,
             '{"required": ["site_id"], "optional": ["domain", "spec", "work_item_id"]}'::jsonb,
             '{"produces": {"template_fix_result": "templates updated count", "rerender_result": "slots re-rendered"}}'::jsonb
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


-- ============================================================================
-- 3. NEW GO ACTION NEEDED: fix_nav_link_templates
--
-- This action does the template repair. Pseudocode:
--
--   func FixNavLinkTemplatesAction(ctx, params) {
--     siteID := resolve("site_id")
--     patterns := config["patterns"]  // [{find, replace}]
--
--     // Find component_ids assigned to this site's header and footer slots
--     rows := query(`
--       SELECT sc.slot_name, sc.component_id, cc.html_template
--       FROM site_components sc
--       JOIN content_components cc ON sc.component_id = cc.id
--       WHERE sc.site_id = $1 AND sc.slot_name IN ('header', 'footer')
--     `, siteID)
--
--     updated := 0
--     for each row {
--       newTemplate := row.html_template
--       for each pattern in patterns {
--         newTemplate = strings.ReplaceAll(newTemplate, pattern.find, pattern.replace)
--       }
--       if newTemplate != row.html_template {
--         exec(`UPDATE content_components SET html_template = $1 WHERE id = $2`,
--              newTemplate, row.component_id)
--         updated++
--       }
--     }
--     return {updated, slot_names}
--   }
--
-- Register as: "fix_nav_link_templates" in action registry
-- ============================================================================


-- ============================================================================
-- 4. IMMEDIATE FIX FOR GASWHOLESALERS.COM
-- Fix the template directly in the database, then re-render will pick it up.
-- ============================================================================

-- First, find the header template and see what it uses
-- SELECT sc.slot_name, sc.component_id,
--        SUBSTRING(cc.html_template FROM 'href="[^"]*"' FOR 60) as link_pattern
-- FROM site_components sc
-- JOIN content_components cc ON sc.component_id = cc.id
-- WHERE sc.site_id = '5fe15466-4e2e-4ff2-981e-98c1b7074002'
-- AND sc.slot_name IN ('header', 'footer');

-- Fix header template: replace #{{.slug}} with {{.url}}
-- (Run the SELECT first to verify the component_id, then update)
UPDATE content_components
SET html_template = REPLACE(
        REPLACE(
                REPLACE(html_template, 'href="#{{.slug}}"', 'href="{{.url}}"'),
                'href="#{{ .slug }}"', 'href="{{ .url }}"'
        ),
        'href="#{{.name}}"', 'href="{{.url}}"'
                    )
WHERE id IN (
    SELECT component_id FROM site_components
    WHERE site_id = '5fe15466-4e2e-4ff2-981e-98c1b7074002'
      AND slot_name IN ('header', 'footer')
);

-- NOTE: After running this UPDATE, the site needs:
-- 1. render_site_components with force_rerender=true (rebuilds header/footer HTML from fixed templates)
-- 2. rerender-pages (re-assembles all pages with new header/footer and git commits)
-- This can be done by inserting a needs_rerender work item and triggering the dispatch loop,
-- or by manually calling the rerender-pages agent.


-- ============================================================================
-- 5. IMPROVEMENT LOOP INTEGRATION
--
-- The improvement loop runs after initial build completes:
--
--   Initial build completes (all needs_* items done)
--   → Spawn quality-discovery-agent  (runs broken_nav_links, placeholder_contact, generic_theme)
--   → Spawn design-discovery-agent   (runs undeployed_assets, missing_css, duplicate_palette)
--   → Spawn completeness-discovery-agent (runs empty_sections)
--   → Findings written as work items with status='detected'
--   → Triage step promotes to status='triaged' with correct domain='build'
--   → Dispatch loop picks them up and calls the handler agents:
--       broken_nav_links    → nav-link-fixer
--       placeholder_contact → page-content-writer (with spec indicating which sections to re-write)
--       generic_theme       → webdesign-agent (with identity/classification context this time)
--       empty_sections      → page-content-writer
--       undeployed_assets   → asset-deployer
--       missing_css         → webdesign-agent
--   → After all fixes complete, needs_rerender assembles + deploys
--
-- The improvement loop can be triggered by:
-- a) A side-effect in the dispatch loop after the last build item completes
-- b) A scheduled job (maintenance-batch-scheduler)
-- c) Manual trigger
-- ============================================================================