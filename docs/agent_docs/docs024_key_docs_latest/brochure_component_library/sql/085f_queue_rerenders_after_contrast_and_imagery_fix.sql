-- 085f_queue_rerenders_after_contrast_and_imagery_fix.sql
--
-- Re-render every page whose sections are complete, so the component template
-- changes (085b/085d) and the content_data imagery (085e) reach the served
-- pages. reason='section_data_resolved' routes page-rerender to its
-- rerender_sections pre-pass: every section re-rendered from stored
-- content_data through the CURRENT template, no LLM.
--
-- The three pages omitted (llm-cost-calculator-guide, platform-log-index,
-- tool-decision-record) each hold a section with NULL content_data, and a page
-- carrying one escalates to the content writer, which REWRITES the copy. That
-- is not what this change is for.
--
-- Fresh rows, not a re-queue: bugs_open/070's reaper keys on created_at, so a
-- re-queued row is born stale and parks itself.

INSERT INTO site_work_items (
  site_id, item_type, item_key, status, pipeline, summary, spec,
  handler_agent, source, created_by, attempt_count, max_attempts, created_at, updated_at)
SELECT p.site_id,
       'page_rerender',
       'page_rerender_' || p.name || '_199733a8-ac9c-4c30-b2ce-65ecdac6f3bd_contrast_imagery_20260727',
       'triaged', 'build',
       'Republish ' || p.name || ': palette contrast fix + generated imagery wired in',
       jsonb_build_object(
         'domain',    'fundamentallyai.com',
         'page_id',   p.id::text,
         'page_name', p.name,
         'filename',  CASE WHEN p.name = 'self-correction-leopardessconsulting'
                           THEN 'blog/' || p.name || '.html'
                           ELSE p.name || '.html' END,
         'reason',    'section_data_resolved'),
       'page-rerender', 'operator:brochure_component_library', 'operator:brochure_component_library',
       0, 3, NOW(), NOW()
  FROM pages p
 WHERE p.site_id = '199733a8-ac9c-4c30-b2ce-65ecdac6f3bd'
   AND p.name IN ('index','capabilities','about','contact','model-fine-tuning',
                  'multi-agent-review-council','self-correction-leopardessconsulting',
                  'tool-model-approach-selector-guide','tool-model-approach-selector',
                  'llm-cost-calculator');
