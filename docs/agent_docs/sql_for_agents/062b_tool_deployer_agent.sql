-- ============================================================================
-- tool-deployer agent definition
--
-- Deploys a tool from the library to a site using the fork-on-deploy model.
-- Handles: add_tool work items
-- Go actions: deploy_tool_to_site (already registered)
-- ============================================================================

INSERT INTO agent_definitions (
    type, display_name, description, category, status,
    image_repository, image_tag,
    resources, default_config, input_contract, output_contract,
    domain_tags, agent_category, idle_timeout_seconds
) VALUES (
             'tool-deployer',
             'Tool Deployer',
             'Deploys a tool from the library to a site using the fork-on-deploy model. Forks the library component, creates a tool page, and links them. The page then flows through the normal render/deploy pipeline.',
             'specialist', 'active',
             'docker.io/aqls/agent-chassis', 'v1.0.861',
             '{"limits": {"cpu": "500m", "memory": "512Mi"}, "requests": {"cpu": "100m", "memory": "256Mi"}}'::jsonb,
             '{
                 "workflow": {
                     "start_step": "ensure_site_record",
                     "processing_mode": "orchestrator",
                     "timeout_seconds": 180,
                     "steps": {
                         "ensure_site_record": {
                             "action": "ensure_site_record",
                             "config": {},
                             "next_step": "check_has_tool_id",
                             "output_field": "site_record"
                         },
                         "check_has_tool_id": {
                             "action": "conditional",
                             "config": {
                                 "condition": "input_data.spec.tool_component_id != null",
                                 "then_step": "deploy_tool",
                                 "else_step": "lookup_library_tool"
                             },
                             "description": "Check if spec includes a specific library tool ID"
                         },
                         "lookup_library_tool": {
                             "action": "query_database",
                             "config": {
                                 "query": "SELECT id::text as tool_component_id, function, display_name FROM content_components WHERE component_level = ''tool'' AND forked_from IS NULL AND is_active = true AND function = $1 LIMIT 1",
                                 "params": ["input_data.spec.function"],
                                 "output_format": "object"
                             },
                             "next_step": "check_lookup_result",
                             "error_step": "complete_no_tool",
                             "description": "Find library tool by function name from suggestion",
                             "output_field": "library_tool"
                         },
                         "check_lookup_result": {
                             "action": "conditional",
                             "config": {
                                 "condition": "library_tool.tool_component_id != null",
                                 "then_step": "deploy_tool_from_lookup",
                                 "else_step": "complete_no_tool"
                             },
                             "description": "Check if we found the tool in the library"
                         },
                         "deploy_tool_from_lookup": {
                             "action": "deploy_tool_to_site",
                             "config": {
                                 "site_id": "site_record.site_id",
                                 "tool_component_id": "library_tool.tool_component_id",
                                 "nav_section": "Tools",
                                 "in_header": true,
                                 "in_footer": false
                             },
                             "next_step": "complete",
                             "error_step": "complete_error",
                             "description": "Fork library tool and create tool page (from lookup)",
                             "output_field": "deploy_result"
                         },
                         "deploy_tool": {
                             "action": "deploy_tool_to_site",
                             "config": {
                                 "site_id": "site_record.site_id",
                                 "tool_component_id": "input_data.spec.tool_component_id",
                                 "nav_section": "Tools",
                                 "in_header": true,
                                 "in_footer": false
                             },
                             "next_step": "complete",
                             "error_step": "complete_error",
                             "description": "Fork library tool and create tool page (from spec)",
                             "output_field": "deploy_result"
                         },
                         "complete": {
                             "action": "complete_workflow",
                             "config": {
                                 "output_fields": ["deploy_result"]
                             },
                             "description": "Tool deployed"
                         },
                         "complete_no_tool": {
                             "action": "complete_workflow",
                             "config": {
                                 "output_fields": [],
                                 "success_message": "No matching library tool found — needs tool-generator (not yet implemented)"
                             },
                             "description": "Skip — tool not in library"
                         },
                         "complete_error": {
                             "action": "complete_workflow",
                             "config": {
                                 "output_fields": ["deploy_result"],
                                 "success_message": "Tool deployment failed"
                             },
                             "description": "Error path"
                         }
                     }
                 }
             }'::jsonb,
             '{"required": ["site_id", "domain"], "optional": ["spec"]}'::jsonb,
             '{"produces": {"deploy_result": "fork_id, page_id, page_url, needs_rerender"}}'::jsonb,
             '["tools", "deployment", "fork"]'::jsonb,
             'specialist',
             120
         ) ON CONFLICT (type, version) DO UPDATE SET
    default_config = EXCLUDED.default_config,
    description = EXCLUDED.description,
    display_name = EXCLUDED.display_name,
    image_tag = EXCLUDED.image_tag,
    idle_timeout_seconds = EXCLUDED.idle_timeout_seconds,
    status = EXCLUDED.status,
    updated_at = NOW();

-- ============================================================================
-- Verify
-- ============================================================================
SELECT type, display_name, status
FROM agent_definitions
WHERE type = 'tool-deployer' AND deleted_at IS NULL;

-- ============================================================================
-- Unblock the existing add_tool items
-- (feasibility-recheck will promote them on next run, but we can do it now)
-- ============================================================================
UPDATE site_work_items
SET status = 'triaged'
WHERE item_type = 'add_tool'
  AND status = 'blocked';