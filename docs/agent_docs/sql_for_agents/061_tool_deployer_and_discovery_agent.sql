-- 065b_tool_deployer_agent.sql
--
-- Agent definition for tool-deployer and discovery check wiring.
--
-- Flow:
--   improvement-loop → design-discovery-agent (runs "missing_tools" check)
--     → writes site_work_item with handler_agent='tool-deployer', spec includes tool_component_id
--     → triage promotes to 'triaged'
--     → build-dispatch-loop spawns tool-deployer
--     → tool-deployer: loads work item → forks tool → creates page → completes item
--     → improvement-loop inserts needs_rerender → pages redeployed
--
-- Manual triage entry for adding a tool to a specific site:
--   INSERT INTO site_work_items (
--     site_id, source, domain, item_type, severity, summary,
--     spec, priority, handler_agent, status, created_by, item_key
--   ) VALUES (
--     '<site_uuid>', 'manual', 'build', 'add_tool', 'low',
--     'Add A/B Test Calculator tool',
--     '{"tool_component_id": "<tool_uuid>"}'::jsonb,
--     80, 'tool-deployer', 'triaged', 'admin',
--     'add_tool:<tool_function>'
--   );
--
-- Prerequisites:
--   - deploy_tool_action.go compiled and deployed
--   - Action registered in registry.go:
--       "deploy_tool_to_site": {
--           Handler:     DeployToolToSiteAction,
--           Category:    "build",
--           Description: "Fork library tool and create tool page for a site",
--       },
--   - Discovery check addition in run_discovery_checks_action.go:
--       findMissingTools function + containsCheck block

-- ============================================================
-- 1. Tool-deployer agent definition
-- ============================================================
-- Handles work items of type: add_tool
-- The work item spec MUST contain: tool_component_id
-- The work item provides: site_id (from the item itself)

--
-- Fixes from v1:
--   role → category
--   tags → capabilities
--   image → image_repository
--   image_version → image_tag
--   resource_limits → resources
--   health_check → health_config
--   replicas → removed (not a column)
--   delegation_config → delegation_preferences
--   tier → agent_category
--   lifecycle_stage → status
--   pipeline_tags → domain_tags
--   ON CONFLICT (type) → ON CONFLICT (type, version)
--
-- Flow:
--   improvement-loop → design-discovery-agent (runs "missing_tools" check)
--     → writes site_work_item with handler_agent='tool-deployer'
--     → triage promotes to 'triaged'
--     → build-dispatch-loop spawns tool-deployer
--     → tool-deployer: loads work item → forks tool → creates page → completes item
--
-- Prerequisites:
--   - deploy_tool_action.go compiled and deployed
--   - Action registered in registry.go
--   - findMissingTools added to run_discovery_checks_action.go

-- ============================================================
-- 1. Tool-deployer agent definition
-- ============================================================

INSERT INTO agent_definitions (
    type, display_name, description, category,
    default_config, is_active, capabilities, image_repository, image_tag,
    resources, topics, health_config, env_vars,
    delegation_preferences, agent_category, status,
    domain_tags, input_contract, output_contract
) VALUES (
             'tool-deployer',
             'Tool Deployer',
             'Deploys tools from the library to sites. Forks the canonical tool component (creating a site-owned copy), creates a tool page, and links the fork as a page_component. The site then owns the tool independently.',
             'specialist',
             '{
               "workflow": {
                 "steps": {
                   "load_item": {
                     "action": "load_work_items",
                     "config": {
                       "site_id": "input_data.site_id",
                       "item_domain": "build",
                       "handler_agent": "tool-deployer",
                       "max_items": 1
                     },
                     "next_step": "check_has_item",
                     "description": "Load next add_tool work item for this site",
                     "output_field": "work_items"
                   },
                   "check_has_item": {
                     "action": "conditional",
                     "config": {
                       "condition": "work_items.count > 0",
                       "then_step": "deploy_tool",
                       "else_step": "complete_empty"
                     },
                     "description": "Check if there are tool items to process"
                   },
                   "deploy_tool": {
                     "action": "deploy_tool_to_site",
                     "config": {
                       "site_id": "input_data.site_id",
                       "tool_component_id": "work_items.items.0.spec.tool_component_id",
                       "nav_section": "Tools",
                       "in_header": true,
                       "in_footer": false
                     },
                     "next_step": "complete_item",
                     "error_step": "fail_item",
                     "description": "Fork library tool, create page, link component",
                     "output_field": "deploy_result"
                   },
                   "complete_item": {
                     "action": "complete_work_item",
                     "config": {
                       "work_item_id": "work_items.items.0.id",
                       "result": "deploy_result"
                     },
                     "next_step": "complete",
                     "description": "Mark work item as complete",
                     "output_field": "item_completed"
                   },
                   "fail_item": {
                     "action": "fail_work_item",
                     "config": {
                       "work_item_id": "work_items.items.0.id",
                       "error_message": "Tool deployment failed"
                     },
                     "next_step": "complete",
                     "description": "Mark work item as failed"
                   },
                   "complete_empty": {
                     "action": "complete_workflow",
                     "config": {
                       "success_message": "No tool items to process",
                       "output_fields": ["work_items"]
                     },
                     "description": "No items found"
                   },
                   "complete": {
                     "action": "complete_workflow",
                     "config": {
                       "output_fields": ["deploy_result", "item_completed"]
                     },
                     "description": "Tool deployment complete"
                   }
                 },
                 "start_step": "load_item"
               },
               "processing_mode": "task",
               "timeout_seconds": 120
             }'::jsonb,
             true,
             '["build", "tools", "deployment"]'::jsonb,
             'docker.io/aqls/agent-chassis',
             'latest',
             '{"limits": {"cpu": "500m", "memory": "512Mi"}, "requests": {"cpu": "100m", "memory": "256Mi"}}'::jsonb,
             '{"error": "system.errors.{type}", "process": "system.agent.{type}.process", "response": "system.responses.{type}"}'::jsonb,
             '{"port": 8080, "liveness_path": "/health", "readiness_path": "/ready", "initial_delay_seconds": 15}'::jsonb,
             '[]'::jsonb,
             '{"fallback_to_self": true, "prefer_delegation": true}'::jsonb,
             'executor',
             'experimental',
             '["build", "tools"]'::jsonb,
             '{"optional": ["domain"], "required": ["site_id"], "description": "Receives site_id from dispatch loop. Loads add_tool work items and deploys tools."}'::jsonb,
             '{"produces": {"deploy_result": "fork_id, page_id, page_url, tool_function"}}'::jsonb
         )
    ON CONFLICT (type, version) DO UPDATE SET
    default_config = EXCLUDED.default_config,
                                       display_name = EXCLUDED.display_name,
                                       description = EXCLUDED.description,
                                       updated_at = NOW();


-- ============================================================
-- 2. Add "missing_tools" to design-discovery-agent checks
-- ============================================================

UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,run_checks,config,checks}',
        '["undeployed_assets", "missing_css", "duplicate_palette", "hardcoded_section_colors", "forced_text_colors", "missing_tools"]'::jsonb
                     ),
    updated_at = NOW()
WHERE type = 'design-discovery-agent';


-- ============================================================
-- 3. Verify
-- ============================================================
SELECT type, display_name,
       default_config->'workflow'->'start_step' as start_step
FROM agent_definitions
WHERE type = 'tool-deployer';

SELECT type,
       default_config->'workflow'->'steps'->'run_checks'->'config'->'checks' as checks
FROM agent_definitions
WHERE type = 'design-discovery-agent';

