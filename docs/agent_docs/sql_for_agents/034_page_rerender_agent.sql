-- ============================================================
-- 38_page_rerender_agent.sql
-- Creates page-rerender agent and updates rerender-pages to call it
-- ============================================================

-- ============================================================
-- 1. Create page-rerender agent (handles single page)
-- ============================================================
INSERT INTO agent_definitions (
    type,
    display_name,
    description,
    category,
    default_config,
    status,
    domain_tags,
    input_contract,
    output_contract
)
VALUES (
           'page-rerender',
           'Page Rerender Agent',
           'Renders and deploys a single page from stored sections. Called by rerender-pages for each page in a loop.',
           'specialist',
           '{
               "workflow": {
                   "start_step": "render_page",
                   "steps": {
                       "render_page": {
                           "action": "rerender_single_page",
                           "config": {
                               "input_fields": ["page_id", "site_id", "domain"],
                               "max_nav_items": 6
                           },
                           "description": "Render page from stored sections",
                           "output_field": "rendered_page",
                           "next_step": "check_skipped"
                       },
                       "check_skipped": {
                           "action": "conditional",
                           "config": {
                               "condition": "rendered_page.skipped == true",
                               "then_step": "complete_skipped",
                               "else_step": "deploy_page"
                           },
                           "description": "Skip deploy if no sections stored"
                       },
                       "deploy_page": {
                           "action": "git_commit",
                           "config": {
                               "repo_name": "sites",
                               "domain_field": "rendered_page.domain",
                               "content_field": "rendered_page.html",
                               "filename_field": "rendered_page.filename",
                               "commit_message": "Rerender: {{.filename}}"
                           },
                           "description": "Deploy rendered page to git",
                           "output_field": "deploy_result",
                           "next_step": "update_status"
                       },
                       "update_status": {
                           "action": "update_page_status",
                           "config": {
                               "status": "deployed",
                               "page_id_field": "input_data.page_id",
                               "commit_from": "deploy_result.commit_sha"
                           },
                           "description": "Update page status in database",
                           "output_field": "status_updated",
                           "next_step": "complete"
                       },
                       "complete_skipped": {
                           "action": "complete_workflow",
                           "config": {
                               "output_fields": ["rendered_page"]
                           },
                           "description": "Page skipped - no sections stored"
                       },
                       "complete": {
                           "action": "complete_workflow",
                           "config": {
                               "output_fields": ["rendered_page", "deploy_result", "status_updated"]
                           },
                           "description": "Page rerender complete"
                       }
                   }
               },
               "processing_mode": "orchestrator",
               "timeout_seconds": 120
           }'::jsonb,
           'active',
           '["rerender", "page", "deploy"]'::jsonb,
           '{
               "required": ["page_id", "site_id", "domain"],
               "description": "Renders and deploys a single page from stored sections"
           }'::jsonb,
           '{
               "produces": {
                   "rendered_page": "object with html, filename, domain, page_id, page_name",
                   "deploy_result": "git commit result with commit_sha",
                   "status_updated": "page status update result"
               }
           }'::jsonb
       )
    ON CONFLICT (type, version) DO UPDATE SET
    display_name = EXCLUDED.display_name,
                                       description = EXCLUDED.description,
                                       default_config = EXCLUDED.default_config,
                                       status = EXCLUDED.status,
                                       domain_tags = EXCLUDED.domain_tags,
                                       input_contract = EXCLUDED.input_contract,
                                       output_contract = EXCLUDED.output_contract,
                                       updated_at = NOW();

--

-- =============================================================================
-- PART 1: Minor update to page-rerender specialist
-- =============================================================================
-- Keep existing workflow, just ensure processing_mode is set correctly
-- and input_contract matches what the orchestrator will send via input_mapping.
-- =============================================================================

UPDATE agent_definitions
SET
    default_config = jsonb_set(
            default_config,
            '{processing_mode}',
            '"task"'
                     ),
    updated_at = NOW()
WHERE type = 'page-rerender';

--

-- ============================================================
-- Update page-rerender workflow: deploy HTML + JS assets together
-- ============================================================
-- Changes deploy_page step from content_field (single file) to
-- files_field (multi-file). The files map comes from
-- RerenderSinglePageAction which now includes JS assets.
--
-- Backward compatible: content_field kept as fallback in case
-- files map is empty (extractFilesForGit falls through to Method 3).

UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,deploy_page,config}',
        jsonb_build_object(
                'repo_name', 'sites',
                'domain_field', 'rendered_page.domain',
                'files_field', 'rendered_page.files',
                'content_field', 'rendered_page.html',
                'filename_field', 'rendered_page.filename',
                'commit_message', 'Rerender: {{.filename}}'
        )
                     ),
    default_config = jsonb_set(
            default_config,
            '{workflow,steps,deploy_page,description}',
            '"Deploy assembled page and JS assets to git"'::jsonb
                     ),
    updated_at = now()
WHERE type = 'page-rerender';

-- Verify
SELECT type,
       default_config->'workflow'->'steps'->'deploy_page'->'config'->>'files_field' as files_field,
    default_config->'workflow'->'steps'->'deploy_page'->'config'->>'content_field' as content_field_fallback
FROM agent_definitions
WHERE type = 'page-rerender';
--
-- consolidation page rerender/rerender pages or stg

UPDATE agent_definitions
SET default_config = jsonb_set(
        jsonb_set(default_config, '{workflow,start_step}', '"check_rerender_mode"'),
        '{workflow,steps}',
        (default_config #> '{workflow,steps}') || '{
           "check_rerender_mode": {
             "action": "conditional",
             "config": {"condition": "input_data.spec.reason == ''image_landed'' OR input_data.spec.reason == ''section_data_resolved''", "then_step": "rerender_sections", "else_step": "render_page"},
             "description": "Re-render sections only for re-render reasons; else assemble stored HTML"
           },
           "rerender_sections": {
             "action": "rerender_page_sections",
             "config": {"target_site_id": "input_data.site_id", "page_name": "input_data.spec.page_name", "reason": "input_data.spec.reason"},
             "output_field": "rerender_sections", "next_step": "check_escalated",
             "description": "Re-render all sections from stored content_data + fresh resolved fields (no LLM)"
           },
           "check_escalated": {
             "action": "conditional",
             "config": {"condition": "rerender_sections.escalated == true", "then_step": "complete", "else_step": "save_sections"},
             "description": "If a section had no content_data, the page was escalated to the writer — stop"
           },
           "save_sections": {
             "action": "save_page_sections",
             "config": {"site_id_field": "rerender_sections.site_id", "page_name_field": "input_data.spec.page_name", "sections_metadata_field": "rerender_sections.sections_metadata"},
             "output_field": "sections_saved", "next_step": "render_page",
             "description": "Persist re-rendered sections, then assemble + deploy via the existing path"
           }
         }'::jsonb
                     ),
    updated_at = now()
WHERE type = 'page-rerender';