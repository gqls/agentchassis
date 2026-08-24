-- ============================================================================
-- agritec.uk — the guides hub lists nothing; swap guide-list for blog-listing
-- Written 2026-08-24. Applied out of band (psql -f), per-site setup.
--
-- THE DEFECT. /guides/index.html has the sections the roadmap brief required —
-- ["hero","guide-list"] — and the guide-list instance is `deployed` carrying
-- 4,937 chars of rendered HTML. It contains ZERO anchors and ZERO cards: just a
-- heading ("Explainers behind the calculators") and a CTA. Three of the six
-- explainers have no inbound link from anywhere on the site.
--
-- This is the exact defect the rebuild exists to fix, reproduced on the rebuild.
--
-- THE CAUSE IS MY OWN DECISION, and both halves of it were right on their own.
-- `guide-list` resolves page_type='guide'. The explainers were deliberately
-- built as page_type='blog-post' because the measurement said blog-post yields
-- ~1,600 words and guide yields ~511 — and it delivered: the four rendered so
-- far run 1,400-1,726 words against the retired site's 315-453. So the depth
-- decision worked and orphaned the pages at the same time.
--
-- The lesson, which is now in the ledger: **a list component and the page type
-- it resolves are a PAIR.** The roadmap brief said "build the hub with an
-- explicit list section from day one", which was necessary and not sufficient,
-- because it named one half of the pair.
--
-- IT WILL NOT SELF-HEAL. Checked rather than assumed: rerender_single_page_action
-- READS pages.sections and assembles from existing components; nothing in the
-- rerender path rewrites the section plan. The queued page_rerender items would
-- re-render an empty list indefinitely.
--
-- THE FIX WAS VERIFIED AT THE ARTEFACT BEFORE BEING ADOPTED. blog-listing is the
-- component that resolves blog-posts, and fundamentallyai.com's
-- /platform-log/index.html is the same shape as this hub: a section-index with
-- ["hero","blog-listing"]. **That page is the subject of bugs_open/309** ("six
-- unclickable cards so every article is orphaned"), so copying the pattern blind
-- would have been copying a bug. I read its rendered HTML first: 16 anchors with
-- real /blog/... hrefs. 309's defect is not present there now, and the component
-- does what it says.
--
-- WHAT THIS DOES NOT DO. It does not hand-write any HTML or any component
-- instance — that would be the 2026-08-04 ruling. It changes the section PLAN
-- and raises a needs_page item so page-build-handler builds the section the way
-- it built every other one. The framework still writes the content.
-- ============================================================================

\set ON_ERROR_STOP on

BEGIN;

DO $guard$
DECLARE sec jsonb; n int;
BEGIN
  SELECT p.sections INTO sec FROM pages p JOIN sites s ON s.id=p.site_id
   WHERE s.domain='agritec.uk' AND p.url='/guides/index.html';
  IF sec IS NULL THEN RAISE EXCEPTION 'guides hub not found'; END IF;
  IF sec::text NOT LIKE '%guide-list%' THEN
    RAISE EXCEPTION 'guides hub sections are % - expected to contain guide-list; another session may have fixed this already', sec::text;
  END IF;
  -- the dedup index refuses a second live needs_page:guides-index; check first so
  -- the failure is a sentence rather than a 42P10.
  SELECT count(*) INTO n FROM site_work_items wi JOIN sites s ON s.id=wi.site_id
   WHERE s.domain='agritec.uk' AND wi.item_key='needs_page:guides-index'
     AND wi.status <> ALL (ARRAY['complete','verified','rejected','wont_fix','failed','unresolved','cancelled']);
  IF n > 0 THEN RAISE EXCEPTION 'a live needs_page:guides-index already exists (%) - not queueing a duplicate', n; END IF;
END
$guard$;

UPDATE pages p
   SET sections = '["hero","blog-listing"]'::jsonb,
       build_status = 'needs_rebuild',
       updated_at = now()
  FROM sites s
 WHERE s.id = p.site_id AND s.domain='agritec.uk' AND p.url='/guides/index.html';

-- Retire the empty guide-list instance so the rebuild does not assemble it
-- alongside the new one. build_status='removed' is the documented
-- assembly-excluded tombstone.
UPDATE page_components pc
   SET build_status = 'removed', updated_at = now()
  FROM pages p, sites s, content_components cc
 WHERE pc.page_id = p.id AND p.site_id = s.id AND cc.id = pc.component_id
   AND s.domain='agritec.uk' AND p.url='/guides/index.html' AND cc.function='guide-list';

INSERT INTO site_work_items
  (site_id, item_type, item_key, pipeline, priority, handler_agent, status, source, created_by, summary, spec)
SELECT s.id, 'needs_page', 'needs_page:guides-index', 'build', 52, 'page-build-handler', 'triaged', 'operator', 'agritec-workstream-2026-08-24',
       'Rebuild guides-index: guide-list resolved page_type=guide and listed nothing; the explainers are blog-post',
       jsonb_build_object(
         'reason','section_plan_changed',
         'page_name','guides-index',
         'page_role','section-index',
         'detail','guide-list rendered 0 anchors because it resolves page_type=guide while the six explainers are page_type=blog-post (chosen for depth: 1400-1726 words vs guide-shape 511). Swapped to blog-listing, verified rendering 16 real anchors on fundamentallyai /platform-log/index.html.')
FROM sites s WHERE s.domain='agritec.uk';

COMMIT;

-- Verify: sections swapped, old instance tombstoned, one live needs_page queued.
--   SELECT sections, build_status FROM pages p JOIN sites s ON s.id=p.site_id
--    WHERE s.domain='agritec.uk' AND p.url='/guides/index.html';
-- Then, after it rebuilds, the check that actually matters:
--   count anchors to /blog/ in the hub's rendered_html — must be 6.
