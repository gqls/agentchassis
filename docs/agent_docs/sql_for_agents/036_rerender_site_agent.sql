


-- =============================================================================
-- PART 2: Create rerender-site orchestrator
-- =============================================================================
-- Workflow:
--   ensure_site_record → spawn_deployer → spawn_rerenderer
--   → render_site_components → get_pages → rerender_loop → trigger_deploy → complete
--
-- The loop sub_workflow is minimal: just call_agent + loop_complete.
-- All per-page logic lives in the page-rerender agent's own workflow.
--
-- Input: { "domain": "example.com" } or { "site_id": "..." }
-- =============================================================================

INSERT INTO agent_definitions (
    id,
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
    status,
    domain_tags,
    input_contract,
    output_contract
) VALUES (
             gen_random_uuid(),
             'rerender-site',
             'Rerender Site',
             'Re-renders and deploys all pages for a site from stored components. Re-renders site-level components (header, footer, head), then spawns page-rerender agent for each page. Used after design changes, component updates, or nav changes.',
             'orchestrator',
             '{
                 "workflow": {
                     "start_step": "ensure_site_record",
                     "steps": {
                         "ensure_site_record": {
                             "action": "ensure_site_record",
                             "config": {
                                 "store_brief_in_content_data": false
                             },
                             "next_step": "spawn_deployer",
                             "description": "Load existing site record",
                             "output_field": "site_record"
                         },
                         "spawn_deployer": {
                             "action": "spawn_agent",
                             "config": {
                                 "role": "deployer",
                                 "agent_type": "deployer-agent"
                             },
                             "next_step": "spawn_rerenderer",
                             "description": "Spawn deployer for final Cloudflare push",
                             "output_field": "deployer_agent"
                         },
                         "spawn_rerenderer": {
                             "action": "spawn_agent",
                             "config": {
                                 "role": "rerenderer",
                                 "agent_type": "page-rerender"
                             },
                             "next_step": "render_site_components",
                             "description": "Spawn page-rerender agent for per-page work",
                             "output_field": "rerenderer_agent"
                         },
                         "render_site_components": {
                             "action": "render_site_components",
                             "config": {
                                 "slots": ["header", "footer", "head"],
                                 "force_rerender": true
                             },
                             "next_step": "get_pages",
                             "description": "Re-render site-level components with current data",
                             "output_field": "site_components_rendered"
                         },
                         "get_pages": {
                             "action": "get_pages_for_rerender",
                             "config": {
                                 "include_statuses": ["deployed", "active"]
                             },
                             "next_step": "rerender_loop",
                             "description": "Get all deployed pages for reassembly",
                             "output_field": "pages_for_rerender"
                         },
                         "rerender_loop": {
                             "action": "loop",
                             "config": {
                                 "mode": "sequential",
                                 "items_field": "pages_for_rerender.pages",
                                 "item_variable": "current_page",
                                 "max_iterations": 50,
                                 "sub_workflow": {
                                     "start_step": "call_rerender",
                                     "steps": {
                                         "call_rerender": {
                                             "action": "call_agent",
                                             "config": {
                                                 "agent_type": "page-rerender",
                                                 "target_role": "rerenderer",
                                                 "input_mapping": {
                                                     "page_id": "current_page.page_id",
                                                     "site_id": "site_record.site_id",
                                                     "domain": "site_record.domain"
                                                 },
                                                 "timeout_seconds": 120
                                             },
                                             "next_step": "complete_page",
                                             "description": "Spawn page-rerender for this page",
                                             "output_field": "page_result"
                                         },
                                         "complete_page": {
                                             "action": "loop_complete",
                                             "description": "Page rerender iteration complete"
                                         }
                                     }
                                 }
                             },
                             "next_step": "trigger_deploy",
                             "description": "Rerender each page via page-rerender agent",
                             "output_field": "pages_rerendered"
                         },
                         "trigger_deploy": {
                             "action": "call_agent",
                             "config": {
                                 "agent_type": "deployer-agent",
                                 "target_role": "deployer",
                                 "input_mapping": {
                                     "site_record": "site_record",
                                     "pages_built": "pages_rerendered"
                                 },
                                 "timeout_seconds": 180
                             },
                             "next_step": "complete",
                             "description": "Trigger Cloudflare deployment",
                             "output_field": "deployment_result"
                         },
                         "complete": {
                             "action": "complete_workflow",
                             "config": {
                                 "output_fields": ["site_record", "site_components_rendered", "pages_rerendered", "deployment_result"]
                             },
                             "description": "Rerender complete"
                         }
                     }
                 },
                 "processing_mode": "orchestrator",
                 "timeout_seconds": 600
             }'::jsonb,
             true,
             '["orchestration", "rerender", "component-based"]',
             'docker.io/aqls/agent-chassis',
             'v1.0.746',
             '{"limits": {"cpu": "500m", "memory": "512Mi"}, "requests": {"cpu": "100m", "memory": "128Mi"}}',
             '{"error": "system.errors.{type}", "process": "system.agent.{type}.process", "response": "system.responses.{type}"}',
             '{"port": 8080, "liveness_path": "/health", "readiness_path": "/ready", "initial_delay_seconds": 30}',
             '[]',
             1,
             'active',
             '["website", "rerender", "component-based"]',
             '{"expects": {"input_data.domain": "string - site domain to rerender", "input_data.site_id": "uuid - alternative to domain"}, "required": ["input_data"]}'::jsonb,
             '{"produces": {"pages_rerendered": "number - count of pages rerendered", "deploy_url": "string - the live site URL"}}'::jsonb
         )
    ON CONFLICT (type, version) WHERE deleted_at IS NULL
    DO UPDATE SET
    display_name = EXCLUDED.display_name,
           description = EXCLUDED.description,
           category = EXCLUDED.category,
           default_config = EXCLUDED.default_config,
           is_active = EXCLUDED.is_active,
           capabilities = EXCLUDED.capabilities,
           image_tag = EXCLUDED.image_tag,
           status = EXCLUDED.status,
           domain_tags = EXCLUDED.domain_tags,
           input_contract = EXCLUDED.input_contract,
           output_contract = EXCLUDED.output_contract,
           updated_at = NOW();