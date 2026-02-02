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

