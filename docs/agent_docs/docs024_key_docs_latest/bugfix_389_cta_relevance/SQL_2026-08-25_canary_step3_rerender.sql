-- bugs_open/391 retirement step 2, part 3: re-resolve the CTA hrefs after the label rewrite.
-- REQUIRED and time-critical: the label rewrite changed the button TEXT only, so between the
-- rewrite and this rerender the page serves a button naming one tool and linking to another —
-- bugs_closed/299's exact defect. Do not leave a page in that state.
--
-- ⚠ spec.page_name IS LOAD-BEARING (LANDMINES.md): a page-rerender dispatched with the reason but
-- WITHOUT page_name re-renders and then DISCARDS the result (sections_saved: 0, success: true) and
-- deploys the stale assembly. Both keys, every time.
BEGIN;
INSERT INTO site_work_items
  (site_id, page_id, source, created_by, item_type, handler_agent, status, severity, priority,
   summary, item_key, spec)
VALUES (
  '1368e337-dd1d-4799-bbb3-8221a1b79bcc',
  'a32b8822-db49-4e45-88f8-bda06d73de62',
  'bugfix_391_cta_relevance', 'bugfix_391_cta_relevance',
  'page_rerender', 'page-rerender', 'detected', 'high', 35,
  'Re-resolve CTA hrefs on technical-details: labels were rewritten to name the Fine-Tuning vs RAG vs Prompting Decision Guide but both still point at the retiring password tool',
  'cta_relink:a32b8822-db49-4e45-88f8-bda06d73de62',
  jsonb_build_object(
    'reason',    'cta_links_stale',
    'check',     'misdirected_cta',
    'page_id',   'a32b8822-db49-4e45-88f8-bda06d73de62',
    'page_name', 'technical-details',
    'fix',       'The hero and call-to-action labels now name a different tool; recompute the CTA targets so each href follows its own copy.',
    'original_pipeline', 'build'
  )
);
SELECT id, status, spec->>'reason' AS reason, spec->>'page_name' AS page_name
FROM site_work_items WHERE item_key='cta_relink:a32b8822-db49-4e45-88f8-bda06d73de62';
COMMIT;
