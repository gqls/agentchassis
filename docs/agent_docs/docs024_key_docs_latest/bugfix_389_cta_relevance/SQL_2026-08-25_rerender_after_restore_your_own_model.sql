-- Re-render finetuning.uk/your-own-model.html so the SERVED bytes pick up the two
-- text blocks restored by SQL_2026-08-25_restore_your_own_model_blocks.sql.
-- Until this runs the page still serves the same "How it works" copy three times.
--
-- reason=cta_links_stale is the no-LLM path. The page's CTA fields already name and
-- link the Fine-Tuning vs RAG vs Prompting Decision Guide, and that destination is a
-- valid page, so applyCTARecompute's KEEP #2 returns early and no CTA moves -- this
-- dispatch is for the re-render and deploy, not for a CTA change.
--
-- spec.page_name IS LOAD-BEARING (LANDMINES.md): without it the rerender discards its
-- own result (sections_saved: 0, success: true) and deploys the stale assembly.
BEGIN;
INSERT INTO site_work_items
  (site_id, page_id, source, created_by, item_type, handler_agent, status, severity, priority,
   summary, item_key, spec)
VALUES (
  '1368e337-dd1d-4799-bbb3-8221a1b79bcc',
  'a8909fc1-f1ff-43fe-842c-5ce364b8b182',
  'bugfix_391_cta_relevance', 'bugfix_391_cta_relevance',
  'page_rerender', 'page-rerender', 'detected', 'high', 30,
  'Re-render your-own-model after restoring two text blocks this lane''s own content_rewrite destroyed: the page currently serves the same How it works copy three times',
  'content_restore_rerender:a8909fc1-f1ff-43fe-842c-5ce364b8b182',
  jsonb_build_object(
    'reason',    'cta_links_stale',
    'check',     'misdirected_cta',
    'page_id',   'a8909fc1-f1ff-43fe-842c-5ce364b8b182',
    'page_name', 'your-own-model',
    'fix',       'Two generic-text-block sections were restored in content_data; re-render the page from content_data and deploy so the served page stops repeating one section three times.',
    'original_pipeline', 'build'
  )
);
SELECT id, status, spec->>'reason' AS reason, spec->>'page_name' AS page_name
FROM site_work_items WHERE item_key='content_restore_rerender:a8909fc1-f1ff-43fe-842c-5ce364b8b182';
COMMIT;
