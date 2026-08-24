-- ============================================================================
-- agritec.uk — fix the guides hub in the SITE PLAN, which is the actual authority
-- Written 2026-08-24. Applied out of band (psql -f), per-site setup.
--
-- SUPERSEDES the approach in SEED_2026-08-24c, which was REVERTED BY THE
-- PIPELINE within five minutes. Keeping both files because the failure is the
-- lesson.
--
-- WHAT 24c DID, AND WHY IT DID NOT HOLD. It set pages.sections to
-- ["hero","blog-listing"], tombstoned the empty guide-list page_component
-- (build_status='removed'), and raised a needs_page. page-build-handler ran at
-- 14:08:35 and completed with a commit. Afterwards:
--     pages.sections            = ["hero", "guide-list"]      <- reverted
--     guide-list page_component = deployed, 4,986 chars       <- resurrected
--
-- Two things happened, and I had written down the second one myself in that very
-- commit message as a follow-up risk:
--   1. `pages.sections` is a DERIVED artefact. page-build-handler rebuilds a page
--      FROM THE PLAN and rewrites pages.sections from it. Editing the derived
--      copy changes what the next build overwrites, not what it produces.
--   2. LANDMINES, "build_status='removed' is a tombstone only until an automated
--      section edit touches the row": the rebuild wrote the component back to a
--      live status unconditionally. I cited that landmine as a follow-up to check
--      LATER; it fired inside five minutes, on the row I had just tombstoned.
--
-- THE AUTHORITY IS site_plan_sections, keyed (plan_id, page_name, ordering):
--     guides-index  0  hero
--     guides-index  1  guide-list     <- this is the row that decides
--
-- So the correct change is one UPDATE to the plan. Then a rebuild derives
-- everything else, including pages.sections and the component instances, the way
-- it derives them for every other page. No tombstone is needed and none is used:
-- letting the builder replace the section is the whole point.
--
-- THE GENERAL SHAPE, which is worth more than this instance: when a fix is
-- reverted by the system rather than failing, you edited a projection instead of
-- a source. The tell is that the write SUCCEEDED and then un-happened.
-- ============================================================================

\set ON_ERROR_STOP on

BEGIN;

DO $guard$
DECLARE n int; live int;
BEGIN
  SELECT count(*) INTO n
    FROM site_plan_sections sps JOIN site_plans sp ON sp.id=sps.plan_id JOIN sites s ON s.id=sp.site_id
   WHERE s.domain='agritec.uk' AND sps.page_name='guides-index' AND sps.component_name='guide-list';
  IF n <> 1 THEN
    RAISE EXCEPTION 'expected exactly 1 guide-list plan row for guides-index, found % - another session may have changed this', n;
  END IF;
  SELECT count(*) INTO live FROM site_work_items wi JOIN sites s ON s.id=wi.site_id
   WHERE s.domain='agritec.uk' AND wi.item_key='needs_page:guides-index'
     AND wi.status <> ALL (ARRAY['complete','verified','rejected','wont_fix','failed','unresolved','cancelled']);
  IF live > 0 THEN RAISE EXCEPTION 'a live needs_page:guides-index already exists (%)', live; END IF;
END
$guard$;

UPDATE site_plan_sections sps
   SET component_name = 'blog-listing'
  FROM site_plans sp, sites s
 WHERE sps.plan_id = sp.id AND sp.site_id = s.id
   AND s.domain='agritec.uk' AND sps.page_name='guides-index' AND sps.component_name='guide-list';

UPDATE pages p
   SET build_status = 'needs_rebuild', updated_at = now()
  FROM sites s
 WHERE s.id = p.site_id AND s.domain='agritec.uk' AND p.url='/guides/index.html';

INSERT INTO site_work_items
  (site_id, item_type, item_key, pipeline, priority, handler_agent, status, source, created_by, summary, spec)
SELECT s.id, 'needs_page', 'needs_page:guides-index', 'build', 51, 'page-build-handler', 'triaged',
       'operator', 'agritec-workstream-2026-08-24',
       'Rebuild guides-index from the corrected plan: blog-listing, not guide-list',
       jsonb_build_object(
         'reason','section_plan_changed',
         'page_name','guides-index',
         'page_role','section-index',
         'detail','site_plan_sections row changed guide-list -> blog-listing. guide-list resolves page_type=guide and rendered 0 anchors; the six explainers are page_type=blog-post, chosen for depth (1400-1726 words vs the guide shape at ~511). Previous attempt edited pages.sections, which page-build-handler overwrites from the plan.')
FROM sites s WHERE s.domain='agritec.uk';

COMMIT;
