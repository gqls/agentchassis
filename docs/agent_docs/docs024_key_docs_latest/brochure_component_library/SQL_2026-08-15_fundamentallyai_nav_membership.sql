-- FILE: SQL_2026-08-15_fundamentallyai_nav_membership.sql
--
-- 2b: give fundamentallyai.com's orphaned /tools.html its nav membership, then let
-- the framework materialise it. NOT YET APPLIED — step 1 (the UPDATE on pages) was
-- refused by the harness permission classifier in the authoring session, so this is
-- committed ready to run rather than as a record of something done.
--
-- WHY: owner, 2026-08-12: "we need links to the tools from the platform log and
-- elsewhere, and a tools entry in the top nav would probably be nice."
--
-- CORRECTED DIAGNOSIS — my first one named the wrong table:
--   The nav is built from pages.in_header / pages.in_footer (dedicated index
--   idx_pages_nav btree (site_id, in_header, nav_order) WHERE status='active') and
--   materialised into site_nav_items. It is NOT built from site_plan_pages.
--   I originally inferred the plan because its six in_header rows matched the six
--   served nav items exactly — but both tables carry the same flags for those pages,
--   so that match was a COINCIDENCE, not a mechanism. Caught by reading a completed
--   nav_drift item, whose own fix string says "rebuild site_nav_items FROM PAGES".
--   Consequence: this touches NO plan row, so it cannot collide with the 215 front's
--   concurrent plan surgery on this same site.
--
-- THE DEFECT: /tools.html is active and serves 200 (27,163 bytes) with
--   in_header = false AND in_footer = false — so it is reachable from nowhere.
--
-- LABEL NOTE: the nav builder derives its label by title-casing the page NAME and
--   ignores nav_label — which is why the footer reads "Llm Cost Calculator" while
--   that page's nav_label says "Tools / LLM Provider Cost Comparison Calculator".
--   For 'tools' the derived label is "Tools", which is what we want, so no label
--   plumbing is required. nav_label is corrected here anyway so the row stops lying,
--   but do NOT expect it to be what renders.
--
-- WHY BOTH STEPS ARE IN ONE TRANSACTION: the nav_drift handler rebuilds the nav FROM
--   pages. Filing it without the declaration would rebuild the identical nav,
--   complete successfully, and report a fix that never happened.
--
-- REVERT: UPDATE pages SET in_header=false, nav_order=203, nav_label='Tools / Index'
--          WHERE site_id='199733a8-ac9c-4c30-b2ce-65ecdac6f3bd' AND name='tools';
--         plus cancel the nav_drift item printed below.

BEGIN;

-- preconditions --------------------------------------------------------------
DO $chk$
DECLARE h bool; f bool; st text;
BEGIN
  SELECT in_header, in_footer, status INTO h, f, st FROM pages
   WHERE site_id='199733a8-ac9c-4c30-b2ce-65ecdac6f3bd' AND name='tools';
  IF NOT FOUND THEN RAISE EXCEPTION 'no tools page on this site'; END IF;
  IF st <> 'active' THEN RAISE EXCEPTION 'tools page is %, not active', st; END IF;
  IF h THEN
    RAISE EXCEPTION 'tools.in_header is ALREADY true — another session has been here; re-read before changing, the gap is elsewhere';
  END IF;
END $chk$;

-- step 1: DECLARE the membership ---------------------------------------------
UPDATE pages
   SET in_header = true,
       nav_order = 4,           -- header currently runs 1,2,3,10,11,12; this sits after Capabilities
       nav_label = 'Tools'
 WHERE site_id='199733a8-ac9c-4c30-b2ce-65ecdac6f3bd' AND name='tools';

-- step 2: MATERIALISE it -------------------------------------------------------
-- Shape cloned from a completed nav_drift item (e2c3a18e, 2026-08-15 00:51,
-- 3-minute turnaround; 23 of 23 such items complete) — never guess a work-item spec.
INSERT INTO site_work_items
  (site_id, page_id, item_type, item_key, summary, spec, handler_agent, pipeline,
   priority, severity, approval_mode, max_attempts, source, created_by, status)
SELECT
  '199733a8-ac9c-4c30-b2ce-65ecdac6f3bd', p.id, 'nav_drift',
  'nav_rebuild:199733a8-ac9c-4c30-b2ce-65ecdac6f3bd:tools-20260815',
  'Nav membership declared for tools — rebuild nav tables and re-render chrome',
  jsonb_build_object(
    'reason','nav_membership_declared',
    'fix','A page has nav membership declared (in_header/in_footer). Rebuild site_nav_items from pages and re-render chrome so the link ships.',
    'page_id', p.id,
    'page_url', p.url,
    'page_name','tools',
    'in_header', true,
    'in_footer', false,
    'requested_by','brochure_contrast_front_thread',
    'context','/tools.html has been live and serving 200 (27,163 bytes) while reachable from nowhere: both nav flags were false. Owner asked on 2026-08-12 for a Tools entry in the top nav. Expected also to correct llm-cost-calculator, which carries in_header=true yet materialised into the FOOTER group only — the stored nav disagrees with the declared flags.'
  ),
  'page-build-handler','build',30,'medium','auto',3,
  'operator:brochure_component_library','brochure_contrast_front_thread','triaged'
FROM pages p
WHERE p.site_id='199733a8-ac9c-4c30-b2ce-65ecdac6f3bd' AND p.name='tools'
ON CONFLICT DO NOTHING;

-- postconditions ---------------------------------------------------------------
DO $post$
DECLARE n int;
BEGIN
  SELECT count(*) INTO n FROM pages
   WHERE site_id='199733a8-ac9c-4c30-b2ce-65ecdac6f3bd' AND name='tools' AND in_header;
  IF n <> 1 THEN RAISE EXCEPTION 'post: tools.in_header not set'; END IF;

  SELECT count(*) INTO n FROM site_work_items
   WHERE created_by='brochure_contrast_front_thread' AND item_type='nav_drift';
  IF n <> 1 THEN RAISE EXCEPTION 'post: expected 1 nav_drift item, found %', n; END IF;

  -- if a /tools.html nav item already exists then the gap was NOT what this fix assumes
  SELECT count(*) INTO n FROM site_nav_items
   WHERE site_id='199733a8-ac9c-4c30-b2ce-65ecdac6f3bd' AND url='/tools.html';
  IF n <> 0 THEN
    RAISE EXCEPTION 'post: a /tools.html nav item already exists (n=%) — stop, the diagnosis is wrong', n;
  END IF;
END $post$;

SELECT name, in_header, in_footer, nav_order, nav_label FROM pages
 WHERE site_id='199733a8-ac9c-4c30-b2ce-65ecdac6f3bd' AND name='tools';

COMMIT;

-- VERIFY AFTER THE HANDLER RUNS — at the artefact, not at the item status:
--   curl -s https://fundamentallyai.com/index.html | grep -o 'href="/tools.html"[^<]*'
--   and confirm the header nav now shows seven items, with the label reading "Tools".
-- A 'complete' nav_drift item over an unchanged served nav is the failure mode to
-- watch for here; the served page is the only thing that settles it.
