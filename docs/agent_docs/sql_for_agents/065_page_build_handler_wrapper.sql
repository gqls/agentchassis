-- ============================================================================
-- page-build-handler: Wrapper handler for page-content-writer
--
-- Solves the specialist vs handler problem: page-content-writer generates
-- content but doesn't persist it. This handler spawns page-content-writer,
-- captures its output, then runs save_page_sections + update_page_status
-- to persist content to page_components.
--
-- No new Go code needed — uses existing actions in a workflow.
--
-- Pre-flight check (Step 0):
--   Existing actions used: ensure_site_record, spawn_agent, call_agent,
--   save_page_sections, update_page_status, page-rerender (agent),
--   complete_workflow. All exist and are registered.
-- ============================================================================

INSERT INTO agent_definitions (
    type, display_name, description, agent_category, status,
    image_repository, image_tag, category,
    default_config, input_contract, output_contract, domain_tags
) VALUES (
             'page-build-handler',
             'Page Build Handler',
             'Wrapper handler for page-content-writer. Spawns the specialist, captures output, persists sections to page_components via save_page_sections, updates page status, assembles and deploys via page-rerender. Dispatch-loop compatible.',
             'specialist', 'active',
             'docker.io/aqls/agent-chassis', 'v1.0.831', 'specialist',
             '{"workflow":{"start_step":"ensure_site_record","processing_mode":"orchestrator","timeout_seconds":600,"steps":{"ensure_site_record":{"action":"ensure_site_record","config":{"store_brief_in_content_data":false},"next_step":"spawn_content_writer","description":"Load site record for context","output_field":"site_record"},"spawn_content_writer":{"action":"spawn_agent","config":{"role":"content_writer","agent_type":"page-content-writer"},"next_step":"call_content_writer","description":"Spawn page-content-writer specialist","output_field":"writer_agent"},"call_content_writer":{"action":"call_agent","config":{"target_role":"content_writer","input_mapping":{"site_id":"site_record.site_id","domain":"site_record.domain","current_page":"input_data.spec","site_record":"site_record","site_plan":"site_record.content_data","reviewed_brief?":"site_record.content_data.reviewed_brief"},"timeout_seconds":300},"next_step":"save_sections","error_step":"complete_error","description":"Call page-content-writer to generate content","output_field":"page_content"},"save_sections":{"action":"save_page_sections","config":{"html_field":"page_content.response.page_html","sections_metadata_field":"page_content.response.sections_metadata","page_name_field":"input_data.spec.page_name","site_id_field":"site_record.site_id"},"next_step":"update_status","error_step":"complete_error","description":"Persist generated sections to page_components","output_field":"sections_saved"},"update_status":{"action":"update_page_status","config":{"status":"deployed","page_name_field":"input_data.spec.page_name","site_id_field":"site_record.site_id"},"next_step":"spawn_rerender_agent","description":"Mark page as deployed","output_field":"status_updated"},"spawn_rerender_agent":{"action":"spawn_agent","config":{"role":"page_renderer","agent_type":"page-rerender"},"next_step":"deploy_page","description":"Spawn page-rerender for assembly and deploy","output_field":"rerender_agent"},"deploy_page":{"action":"call_agent","config":{"target_role":"page_renderer","input_mapping":{"site_id":"site_record.site_id","domain":"site_record.domain","page_id":"sections_saved.page_id"},"timeout_seconds":120},"next_step":"complete","error_step":"complete_error","description":"Assemble page from stored components and deploy to git","output_field":"deploy_result"},"complete":{"action":"complete_workflow","config":{"output_fields":["sections_saved","deploy_result"]},"description":"Page build complete"},"complete_error":{"action":"complete_workflow","config":{"output_fields":["page_content","sections_saved"],"success_message":"Page build completed with errors"},"description":"Page build completed with errors"}}}}'::jsonb,
             '{"required": ["site_id", "domain"], "optional": ["page_name", "page_id", "sections"]}'::jsonb,
             '{"produces": {"sections_saved": "save result with page_id and section count", "deploy_result": "git commit result"}}'::jsonb,
             '["build", "content", "handler"]'::jsonb
         ) ON CONFLICT (type, version) DO UPDATE SET
    default_config = EXCLUDED.default_config,
    description = EXCLUDED.description,
    image_tag = EXCLUDED.image_tag,
    updated_at = NOW();

-- Update existing needs_content_page items to use page-build-handler
UPDATE site_work_items
SET handler_agent = 'page-build-handler'
WHERE item_type = 'needs_content_page'
  AND handler_agent = 'page-content-writer'
  AND status IN ('detected', 'triaged', 'wont_fix');

-- Create a new blog content work item for finetuning.uk
-- (previous one was wont_fix because page-content-writer didn't persist)
INSERT INTO site_work_items (
    site_id, source, domain, item_type, severity, summary,
    page_id, priority, handler_agent, status, created_by, spec
) VALUES (
             '1368e337-dd1d-4799-bbb3-8221a1b79bcc',
             'human', 'build', 'needs_content_page', 'high',
             'Generate content for blog page - currently empty',
             (SELECT id FROM pages WHERE site_id = '1368e337-dd1d-4799-bbb3-8221a1b79bcc' AND name = 'blog'),
             50, 'page-build-handler', 'triaged', 'human',
             '{"page_name": "blog", "sections": ["hero", "article-grid", "call-to-action"]}'::jsonb
         );

-- Verify
SELECT type, display_name, status
FROM agent_definitions WHERE type = 'page-build-handler' AND deleted_at IS NULL;

SELECT item_type, handler_agent, status
FROM site_work_items
WHERE site_id = '1368e337-dd1d-4799-bbb3-8221a1b79bcc'
  AND item_type = 'needs_content_page'
ORDER BY created_at DESC LIMIT 2;