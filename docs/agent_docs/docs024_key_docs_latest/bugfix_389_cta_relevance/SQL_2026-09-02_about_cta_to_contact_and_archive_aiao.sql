-- OWNER DECISION 2026-09-02: "that about button/text should point to the contact page I think."
-- Resolves the ONE field left undecided after the 08-31 ruling, and unblocks the last archive.
--
-- Part A — the about button.
--   ai-agent-orchestration.com/about.html, slot content-block-about, cta_url -> /contact.html.
--   Note its `cta_text` is NULL in content_data; the served label "Learn More About Us" comes from
--   the component template, not the stored data. That does not affect durability: KEEP #1's
--   predicate (storedCTADestinationIsAuthored) is URL-based, not label-based, so /contact.html is
--   held whatever supplies the wording.
--
-- Part B — archive the last tool page.
--   With every labelled field now resolved, archiving is safe: the three remaining UNLABELLED
--   fields have no copy for the positional pick to contradict, so it cannot produce a kind
--   mismatch, and the nav_order 1 -> 900 demotion makes its choice sensible. Archiving also
--   drops the page out of loadResolverPageSet, which is what lets those fields re-resolve at all.
--   Archiving does NOT unpublish -- the page keeps serving until the retraction (step 5).
--
-- Part C — publish: one no-LLM rerender per affected page. spec.page_name is LOAD-BEARING.

BEGIN;

-- A
UPDATE page_components pc
SET content_data = jsonb_set(pc.content_data, '{cta_url}', '"/contact.html"'::jsonb, false),
    updated_at = now()
FROM pages p
WHERE p.id = pc.page_id
  AND p.site_id = '2a8ebf9c-20a2-4c39-b191-840b012371da'
  AND p.url = '/about.html'
  AND pc.slot_name = 'content-block-about'
  AND pc.content_data->>'cta_url' = '/tools/password-entropy.html';

-- B
UPDATE pages
SET status = 'archived', updated_at = now()
WHERE site_id = '2a8ebf9c-20a2-4c39-b191-840b012371da'
  AND url = '/tools/password-entropy.html'
  AND status = 'active';

-- C
INSERT INTO site_work_items
  (site_id, page_id, source, created_by, item_type, handler_agent, status, severity, priority,
   summary, item_key, spec)
SELECT DISTINCT p.site_id, p.id, 'bugfix_391_cta_relevance', 'bugfix_391_cta_relevance',
       'page_rerender', 'page-rerender', 'detected', 'high', 30,
       'Final CTA pass on ' || p.name || ': publish the contact repoint / re-resolve the last unlabelled field now the tool is archived',
       'final_cta_pass:' || p.id::text,
       jsonb_build_object(
         'reason','cta_links_stale', 'check','misdirected_cta',
         'page_id', p.id::text, 'page_name', p.name,
         'fix','The retiring tool is archived. Publish the authored /contact.html destination where present, and recompute any remaining CTA that still points at the archived page.',
         'original_pipeline','build')
FROM pages p
JOIN page_components pc ON pc.page_id = p.id
WHERE p.status = 'active'
  AND (pc.content_data::text LIKE '%password-entropy%' OR pc.rendered_html LIKE '%password-entropy%')
  AND NOT EXISTS (SELECT 1 FROM site_work_items w WHERE w.item_key = 'final_cta_pass:' || p.id::text);

DO $$
DECLARE n_active int; n_about int; n_items int; n_lib int;
BEGIN
  SELECT count(*) INTO n_active FROM pages
   WHERE url='/tools/password-entropy.html' AND status='active';
  IF n_active <> 0 THEN
    RAISE EXCEPTION 'expected 0 tool pages still active, found %', n_active;
  END IF;

  SELECT count(*) INTO n_about FROM page_components pc JOIN pages p ON p.id=pc.page_id
   WHERE p.site_id='2a8ebf9c-20a2-4c39-b191-840b012371da' AND p.url='/about.html'
     AND pc.slot_name='content-block-about' AND pc.content_data->>'cta_url'='/contact.html';
  IF n_about <> 1 THEN
    RAISE EXCEPTION 'about cta not repointed (% rows)', n_about;
  END IF;

  -- Owner decision 1's carve-out: the shared library component STAYS active.
  SELECT count(*) INTO n_lib FROM content_components
   WHERE name ILIKE '%password-entropy%' AND is_active;
  IF n_lib <> 1 THEN
    RAISE EXCEPTION 'expected exactly 1 active library component, found %', n_lib;
  END IF;

  SELECT count(*) INTO n_items FROM site_work_items WHERE item_key LIKE 'final_cta_pass:%';
  RAISE NOTICE 'all three tool pages archived; about cta -> /contact.html; % final rerenders queued; library component intact', n_items;
END $$;

COMMIT;
