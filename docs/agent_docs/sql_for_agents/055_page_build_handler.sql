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

---
--

-- ============================================================================
-- 026h — Page Build Handler: read sections from site_specs.site_plan
-- ============================================================================
--
-- Problem: page-build-handler reads sections from pages.sections, which
-- doesn't include latest-news (or any other sections added after initial
-- page creation). site_specs.site_plan IS authoritative but wasn't consulted.
--
-- Fix: insert a load_spec_sections step that reads from site_specs.site_plan,
-- falls back to pages.sections, and syncs the two. Then plan_sections reads
-- from the spec-sourced list instead of the pages table directly.
--
-- Changes to page-build-handler workflow:
--   1. load_existing_content.next_step: "plan_sections" → "load_spec_sections"
--   2. New step: load_spec_sections (action: load_page_sections_from_spec)
--   3. plan_sections.config.sections: "page_record.sections" → "spec_sections.sections"
--
-- New flow (changed steps marked with *):
--   ensure_site_record → load_page_record → check_page_found
--     → load_existing_content* → load_spec_sections* → plan_sections*
--     → check_has_ready_sections → spawn_content_writer → call_content_writer
--     → check_content_produced → validate_content → save_sections
--     → update_status → spawn_rerender_agent → deploy_page → complete
--
-- Prerequisites:
--   - load_page_sections_from_spec_action.go already deployed in chassis
--   - Action registered as "load_page_sections_from_spec" in registry
--
-- Revert:
--   UPDATE agent_definitions
--   SET default_config = (SELECT default_config
--       FROM agent_def_page_build_handler_backup_20260402 LIMIT 1),
--       updated_at = NOW()
--   WHERE type = 'page-build-handler' AND deleted_at IS NULL;
-- ============================================================================

-- Step 1: Backup
CREATE TABLE IF NOT EXISTS agent_def_page_build_handler_backup_20260402 AS
SELECT * FROM agent_definitions
WHERE type = 'page-build-handler' AND deleted_at IS NULL;

-- Step 2: Change load_existing_content.next_step from "plan_sections" to "load_spec_sections"
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,load_existing_content,next_step}',
        '"load_spec_sections"'::jsonb
                     ),
    updated_at = NOW()
WHERE type = 'page-build-handler' AND deleted_at IS NULL;

-- Step 3: Also change load_existing_content.error_step from "plan_sections" to "load_spec_sections"
-- (if load_existing_content fails, we still want to try loading sections from spec)
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,load_existing_content,error_step}',
        '"load_spec_sections"'::jsonb
                     ),
    updated_at = NOW()
WHERE type = 'page-build-handler' AND deleted_at IS NULL;

-- Step 4: Add the new load_spec_sections step
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,load_spec_sections}',
        jsonb_build_object(
                'action', 'load_page_sections_from_spec',
                'config', jsonb_build_object(
                        'site_id', 'site_record.site_id',
                        'page_name', 'page_record.name',
                        'page_sections_fallback', 'page_record.sections'
                          ),
                'next_step', 'plan_sections',
                'error_step', 'plan_sections',
                'description', 'Load sections from site_specs.site_plan (authoritative), fall back to pages.sections',
                'output_field', 'spec_sections'
        )
                     ),
    updated_at = NOW()
WHERE type = 'page-build-handler' AND deleted_at IS NULL;

-- Step 5: Change plan_sections to read from spec_sections.sections
-- instead of page_record.sections
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,plan_sections,config,sections}',
        '"spec_sections.sections"'::jsonb
                     ),
    updated_at = NOW()
WHERE type = 'page-build-handler' AND deleted_at IS NULL;

-- Verify the changes
SELECT
    default_config->'workflow'->'steps'->'load_existing_content'->'next_step' as load_existing_next,
    default_config->'workflow'->'steps'->'load_spec_sections'->'action' as spec_action,
    default_config->'workflow'->'steps'->'plan_sections'->'config'->'sections' as plan_reads_from
FROM agent_definitions
WHERE type = 'page-build-handler' AND deleted_at IS NULL;

-- Expected output:
--  load_existing_next |          spec_action           |      plan_reads_from
-- --------------------+--------------------------------+---------------------------
--  "load_spec_sections" | "load_page_sections_from_spec" | "spec_sections.sections"


-- put error detection at config not at step level
UPDATE agent_definitions
SET default_config = (
    SELECT jsonb_set(
                   jsonb_set(
                           jsonb_set(
                                   jsonb_set(
                                           jsonb_set(
                                                   jsonb_set(
                                                           jsonb_set(
                                                                   default_config,
                                                                   '{workflow,steps,deploy_page,config,error_step}', '"complete_error"'
                                                           ),
                                                           '{workflow,steps,plan_sections,config,error_step}', '"complete_error"'
                                                   ),
                                                   '{workflow,steps,save_sections,config,error_step}', '"complete_error"'
                                           ),
                                           '{workflow,steps,validate_content,config,error_step}', '"mark_needs_review"'
                                   ),
                                   '{workflow,steps,load_spec_sections,config,error_step}', '"plan_sections"'
                           ),
                           '{workflow,steps,call_content_writer,config,error_step}', '"complete_error"'
                   ),
                   '{workflow,steps,load_existing_content,config,error_step}', '"load_spec_sections"'
           )
),
    updated_at = NOW()
WHERE type = 'page-build-handler';

