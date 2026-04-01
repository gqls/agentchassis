-- 055_page_build_handler.sql
--
-- page-build-handler agent definition
-- Self-contained handler for needs_content_page work items.
--
-- Flow:
--   1. ensure_site_record — load site record from DB
--   2. read_site_spec (site_plan) — load planning data
--   3. read_site_spec (briefing) — load briefing data for content quality
--   4. spawn + call page-content-writer — generates HTML + sections_metadata
--   5. save_page_sections — persist sections to page_components
--   6. update_page_status — mark page as content_ready
--
-- Does NOT git_commit — rerender handles deployment.
-- Can be called standalone (e.g. from CLI) with just site_id + spec.
--
-- Called by: build-dispatch-loop (handler for needs_content_page items)
-- Input from dispatch: site_id, domain, work_item_id, item_type, spec

INSERT INTO agent_definitions (
    type, display_name, description, category,
    default_config, is_active, capabilities,
    image_repository, image_tag, resources, topics,
    health_config, env_vars, version,
    delegation_preferences, agent_category, status,
    domain_tags, briefing_questionnaire,
    usage_count, is_snapshot, input_contract, output_contract
) VALUES (
             'page-build-handler',
             'Page Build Handler',
             'Self-contained handler for content page work items. Loads site context, calls page-content-writer, saves sections to page_components, updates page status. Rerender agent handles final assembly and deployment.',
             'orchestrator',
             '{
                 "workflow": {
                     "start_step": "ensure_site_record",
                     "steps": {

                         "ensure_site_record": {
                             "action": "ensure_site_record",
                             "config": {},
                             "next_step": "read_site_plan",
                             "description": "Load site record from database",
                             "output_field": "site_record"
                         },

                         "read_site_plan": {
                             "action": "read_site_spec",
                             "config": {
                                 "site_id": "site_record.site_id",
                                 "aspect": "site_plan"
                             },
                             "next_step": "read_briefing",
                             "description": "Load site plan from site_specs",
                             "output_field": "site_plan_spec"
                         },

                         "read_briefing": {
                             "action": "read_site_spec",
                             "config": {
                                 "site_id": "site_record.site_id",
                                 "aspect": "briefing"
                             },
                             "next_step": "spawn_writer",
                             "description": "Load briefing from site_specs for content quality",
                             "output_field": "briefing_spec"
                         },

                         "spawn_writer": {
                             "action": "spawn_agent",
                             "config": {
                                 "role": "content_writer",
                                 "agent_type": "page-content-writer"
                             },
                             "next_step": "call_writer",
                             "description": "Spawn page-content-writer",
                             "output_field": "spawn_writer"
                         },

                         "call_writer": {
                             "action": "call_agent",
                             "config": {
                                 "agent_type": "page-content-writer",
                                 "target_role": "content_writer",
                                 "input_mapping": {
                                     "current_page":   "input_data.spec",
                                     "site_record":    "site_record",
                                     "site_plan":      "site_plan_spec.data",
                                     "reviewed_brief": "briefing_spec.data"
                                 },
                                 "timeout_seconds": 180
                             },
                             "next_step": "save_sections",
                             "error_step": "complete_error",
                             "description": "Generate page content",
                             "output_field": "page_content"
                         },

                         "save_sections": {
                             "action": "save_page_sections",
                             "config": {
                                 "sections_metadata_field": "page_content.response.sections_metadata",
                                 "html_field":              "page_content.response.page_html",
                                 "page_name_field":         "input_data.spec.name",
                                 "site_id_field":           "site_record.site_id"
                             },
                             "next_step": "update_page_status",
                             "error_step": "complete_error",
                             "description": "Save rendered sections to page_components for rerender",
                             "output_field": "save_result"
                         },

                         "update_page_status": {
                             "action": "update_page_status",
                             "config": {
                                 "status": "content_ready",
                                 "page_id_field": "input_data.spec.id"
                             },
                             "next_step": "complete",
                             "description": "Mark page as content_ready (rerender sets deployed)",
                             "output_field": "status_updated"
                         },

                         "complete": {
                             "action": "complete_workflow",
                             "config": {
                                 "output_fields": ["page_content", "save_result"]
                             },
                             "description": "Page build complete"
                         },

                         "complete_error": {
                             "action": "complete_workflow",
                             "config": {
                                 "output_fields": ["page_content"],
                                 "success_message": "Page build completed with errors"
                             },
                             "description": "Error path — dispatch loop marks work item failed"
                         }
                     }
                 },
                 "processing_mode": "orchestrator",
                 "timeout_seconds": 300
             }'::jsonb,
             true,
             '["content-generation", "page-build", "section-persistence"]'::jsonb,
             'docker.io/aqls/agent-chassis', 'v1.0.803',
             '{"limits": {"cpu": "500m", "memory": "512Mi"}, "requests": {"cpu": "100m", "memory": "256Mi"}}'::jsonb,
             '{"error": "system.errors.{type}", "process": "system.agent.{type}.process", "response": "system.responses.{type}"}'::jsonb,
             '{"port": 8080, "liveness_path": "/health", "readiness_path": "/ready", "initial_delay_seconds": 15}'::jsonb,
             '[]'::jsonb, 1,
             '{"fallback_to_self": true, "prefer_delegation": true}'::jsonb,
             'coordinator', 'experimental',
             '["build", "content", "page"]'::jsonb, '{}'::jsonb,
             0, false,
             '{"required": ["site_id", "spec"], "optional": ["domain", "work_item_id", "item_type"]}'::jsonb,
             '{"produces": {"page_content": "page_html + sections_metadata", "save_result": "sections saved count"}}'::jsonb
         )
    ON CONFLICT (type, version) DO UPDATE SET
    display_name = EXCLUDED.display_name,
                                       description = EXCLUDED.description,
                                       category = EXCLUDED.category,
                                       default_config = EXCLUDED.default_config,
                                       is_active = EXCLUDED.is_active,
                                       capabilities = EXCLUDED.capabilities,
                                       image_tag = EXCLUDED.image_tag,
                                       resources = EXCLUDED.resources,
                                       agent_category = EXCLUDED.agent_category,
                                       domain_tags = EXCLUDED.domain_tags,
                                       input_contract = EXCLUDED.input_contract,
                                       output_contract = EXCLUDED.output_contract,
                                       updated_at = now();


---

-- add latest news
UPDATE pages
SET sections = '["hero", "features", "services-grid", "differentiators-section", "social_proof", "latest-news", "call_to_action"]'::jsonb,
    updated_at = NOW()
WHERE site_id = '5fe15466-4e2e-4ff2-981e-98c1b7074002'
  AND name = 'index';

