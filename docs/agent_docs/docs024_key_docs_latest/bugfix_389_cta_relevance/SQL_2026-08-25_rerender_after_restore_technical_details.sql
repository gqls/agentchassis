-- Re-render finetuning.uk/technical-details.html so the SERVED bytes pick up the two text
-- blocks restored by SQL_2026-08-25_restore_technical_details_blocks.sql. Until this runs
-- the canary still serves the same "The model and its licence" section three times.
--
-- reason=cta_links_stale is the no-LLM path. The page's CTA fields already name and link the
-- Fine-Tuning vs RAG vs Prompting Decision Guide and that destination is a valid page, so
-- applyCTARecompute's KEEP #2 returns early and no CTA moves -- this is for the re-render and
-- deploy, not for a CTA change.
--
-- spec.page_name IS LOAD-BEARING (LANDMINES.md): without it the rerender discards its own
-- result (sections_saved: 0, success: true) and deploys the stale assembly.
BEGIN;
INSERT INTO site_work_items
  (site_id, page_id, source, created_by, item_type, handler_agent, status, severity, priority,
   summary, item_key, spec)
VALUES (
  '1368e337-dd1d-4799-bbb3-8221a1b79bcc',
  'a32b8822-db49-4e45-88f8-bda06d73de62',
  'bugfix_391_cta_relevance', 'bugfix_391_cta_relevance',
  'page_rerender', 'page-rerender', 'detected', 'high', 30,
  'Re-render technical-details after restoring two text blocks this lane''s own canary content_rewrite destroyed: the page currently serves the same licence section three times',
  'content_restore_rerender:a32b8822-db49-4e45-88f8-bda06d73de62',
  jsonb_build_object(
    'reason',    'cta_links_stale',
    'check',     'misdirected_cta',
    'page_id',   'a32b8822-db49-4e45-88f8-bda06d73de62',
    'page_name', 'technical-details',
    'fix',       'Two generic-text-block sections were restored in content_data; re-render the page from content_data and deploy so the served page stops repeating one section three times.',
    'original_pipeline', 'build'
  )
);
SELECT id, status, spec->>'page_name' AS page_name
FROM site_work_items WHERE item_key='content_restore_rerender:a32b8822-db49-4e45-88f8-bda06d73de62';
COMMIT;
