-- RETIREMENT STEP 2 of 3, CANARY PAGE: prove archiving actually unblocked the re-resolution.
--
-- THE HYPOTHESIS UNDER TEST, stated so it can fail: before the archive, applyCTARecompute's
-- KEEP #2 returned early for these fields because the stored destination was in `validPages`
-- and so counted as "a real, sensible destination". `loadResolverPageSet` selects
-- `status NOT IN ('deleted','archived')`, so archiving finetuning.uk's page should have
-- dropped it out of that set, KEEP #2 should now fail, KEEP #3 cannot catch a relative
-- /tools/... path, and control should reach the positional pick — which the nav_order 1 -> 900
-- demotion already made correct.
--
-- If the hypothesis is WRONG, this rerender completes and /about.html still serves
-- href="/tools/password-entropy.html" twice. That is the disconfirming result, and it is
-- readable at the served bytes in one grep. One page first, not eleven.
--
-- reason=cta_links_stale is the no-LLM path.
-- spec.page_name IS LOAD-BEARING (LANDMINES): without it the rerender discards its own result
-- (sections_saved: 0, success: true) and deploys the stale assembly.
BEGIN;
INSERT INTO site_work_items
  (site_id, page_id, source, created_by, item_type, handler_agent, status, severity, priority,
   summary, item_key, spec)
VALUES (
  '1368e337-dd1d-4799-bbb3-8221a1b79bcc',
  'c0c68034-469f-420c-90bd-d3c0fc0e13d2',
  'bugfix_391_cta_relevance', 'bugfix_391_cta_relevance',
  'page_rerender', 'page-rerender', 'detected', 'high', 30,
  'Retirement step 2 canary: re-resolve the two CTA hrefs on finetuning.uk/about that still point at the now-archived password-entropy tool',
  'retire_relink:c0c68034-469f-420c-90bd-d3c0fc0e13d2',
  jsonb_build_object(
    'reason',    'cta_links_stale',
    'check',     'misdirected_cta',
    'page_id',   'c0c68034-469f-420c-90bd-d3c0fc0e13d2',
    'page_name', 'about',
    'fix',       'The CTA destination /tools/password-entropy.html is now archived and must not be linked. Recompute the CTA targets so each href points at a live, relevant tool.',
    'original_pipeline', 'build'
  )
);
SELECT id, status, spec->>'page_name' AS page_name FROM site_work_items
WHERE item_key='retire_relink:c0c68034-469f-420c-90bd-d3c0fc0e13d2';
COMMIT;
