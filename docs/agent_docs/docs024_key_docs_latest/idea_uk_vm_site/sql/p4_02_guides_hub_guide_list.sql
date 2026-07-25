-- p4_02_guides_hub_guide_list.sql — idea.uk: make /guides/index.html actually list the guides.
--
-- RUN ONLY AFTER p4_01 (+p4_01b) IS VERIFIED LIVE at https://idea.uk/guides/patents/index.html.
-- The guard below enforces that: the hub's listing is a REQUIRED, min_items:1 query field, so
-- swapping it in while no guide is fetchable resolves an empty required list.
--
-- THE DEFECT BEING FIXED. idea.uk's guides hub carries `content-listing`
-- (aa3e4b68-bcea-49ca-890a-c111acefa551), whose `articles` is a STATIC array in content_data with
-- no query source. It has always been empty, so the hub renders a heading and nothing else
-- (601 bytes rendered, verified 2026-07-25; live page shows "Guides" with no cards). Adding guide
-- pages would never populate it — this is the derived-vs-static shape of bugs_open/023 in reverse:
-- a listing that cannot see the pages it is supposed to list.
--
-- THE FIX. Swap the section to `guide-list_pre_037` (9d5e461a-8981-4ecc-b236-05895edfc15d), whose
-- schema sources `items` from `query.pages_where_type:guide` (queryresolve.go:81 ->
-- resolvePagesWhereType) with FetchablePageEligibilitySQL — so it lists every guide page that has
-- actually shipped, and keeps listing them as more are added, with no further edits here. This is
-- the same component the fleet's only populated guides hub uses (gamesdesign.co.uk, 7,758 bytes).
--
-- slot_name MUST move with the component (see p4_01b): rerender_page_sections keys its component
-- lookup on slot_name, and the two components have different `function` values
-- ('content-listing' -> 'guide-list'). Leaving the old slot_name would silently carry the stale
-- empty HTML and report success.
--
-- The LLM-sourced copy fields are AUTHORED here rather than left to the content writer, for the
-- same reason p4_01's body is authored: this hub introduces legal/financial guidance and the copy
-- should not drift. All non-query fields are filled, so a section_data_resolved rerender has no
-- reason to escalate to the LLM.

\set ON_ERROR_STOP on

BEGIN;

-- ---------------------------------------------------------------------------
-- 0. Guards.
-- ---------------------------------------------------------------------------
DO $guard$
DECLARE n int;
BEGIN
  -- The patents guide must be FETCHABLE, or the hub's required items list resolves empty.
  SELECT count(*) INTO n
  FROM pages p
  WHERE p.site_id = '1244516d-014d-421c-88c6-090bb1e9552a'
    AND p.page_type = 'guide'
    AND p.status IN ('active','deployed')
    AND (p.deployed_at IS NOT NULL OR p.build_status = 'deployed');
  IF n < 1 THEN
    RAISE EXCEPTION 'ABORT: no fetchable page_type=guide on idea.uk yet (deployed_at IS NOT NULL OR build_status=deployed). Verify p4_01 live first.';
  END IF;

  SELECT count(*) INTO n FROM content_components
   WHERE id = '9d5e461a-8981-4ecc-b236-05895edfc15d' AND is_active;
  IF n <> 1 THEN
    RAISE EXCEPTION 'ABORT: guide-list_pre_037 not found/active — re-ground the component id.';
  END IF;
END
$guard$;

-- ---------------------------------------------------------------------------
-- 1. Snapshot the section we are replacing (recoverable, per the read-before-write rule).
-- ---------------------------------------------------------------------------
DROP TABLE IF EXISTS bak_ideauk_guideshub_20260725;
CREATE TABLE bak_ideauk_guideshub_20260725 AS
SELECT pc.* FROM page_components pc JOIN pages p ON p.id = pc.page_id
WHERE p.site_id = '1244516d-014d-421c-88c6-090bb1e9552a'
  AND p.url = '/guides/index.html';

-- ---------------------------------------------------------------------------
-- 2. Swap the listing section: component + slot_name + authored copy.
--    `items` is deliberately NOT set — it is query-resolved on every rerender.
-- ---------------------------------------------------------------------------
UPDATE page_components pc
SET component_id = '9d5e461a-8981-4ecc-b236-05895edfc15d',
    slot_name    = 'guide-list',
    content_data = jsonb_build_object(
      'eyebrow_label',    'Guides',
      'section_heading',  'Working an idea out properly, one stage at a time',
      'section_intro',    'Plain-English guides to the decisions that actually cost people money — what to protect and when, what to test before you build, and where the funding really comes from. UK-focused, honest about what we do not know, and free.',
      'cta_heading',      'Not sure your idea is ready for any of this?',
      'cta_subtext',      'Start with the evidence. A Verified Idea Report researches one idea properly — the market, the competition, and a specific next step — for £29.',
      'cta_button_label', 'Get a verified idea report',
      'cta_url',          '/report.html',
      'empty_state_text', 'More guides are being written. The first covers patents and protecting an idea.'
    ),
    updated_at = now()
FROM pages p
WHERE p.id = pc.page_id
  AND p.site_id = '1244516d-014d-421c-88c6-090bb1e9552a'
  AND p.url = '/guides/index.html'
  AND pc.slot_name = 'content-listing';

-- ---------------------------------------------------------------------------
-- 3. pages.sections is the page-level slot-name list and must move with the swap
--    (see p4_01c for why an out-of-date sections array is not cosmetic). It was
--    ["hero","content-listing"]; it becomes ["hero","guide-list"].
-- ---------------------------------------------------------------------------
UPDATE pages p
SET sections = (
      SELECT COALESCE(jsonb_agg(pc.slot_name ORDER BY pc.position), '[]'::jsonb)
      FROM page_components pc
      WHERE pc.page_id = p.id AND COALESCE(pc.slot_name,'') <> ''
    ),
    updated_at = now()
WHERE p.site_id = '1244516d-014d-421c-88c6-090bb1e9552a'
  AND p.url = '/guides/index.html';

DO $guard2$
DECLARE n int; secs jsonb;
BEGIN
  SELECT count(*) INTO n
  FROM page_components pc JOIN pages p ON p.id = pc.page_id
  WHERE p.site_id = '1244516d-014d-421c-88c6-090bb1e9552a'
    AND p.url = '/guides/index.html'
    AND pc.slot_name = 'guide-list';
  IF n <> 1 THEN
    RAISE EXCEPTION 'ABORT: expected exactly 1 guide-list section on the hub, found %.', n;
  END IF;

  SELECT sections INTO secs FROM pages
   WHERE site_id = '1244516d-014d-421c-88c6-090bb1e9552a' AND url = '/guides/index.html';
  IF NOT (secs ? 'guide-list') OR (secs ? 'content-listing') THEN
    RAISE EXCEPTION 'ABORT: pages.sections did not track the swap: %', secs;
  END IF;
END
$guard2$;

COMMIT;

SELECT pc.position, pc.slot_name, cc.name AS component,
       pc.content_data->>'section_heading' AS heading,
       pc.content_data->>'cta_url' AS cta_url
FROM page_components pc
JOIN pages p ON p.id = pc.page_id
JOIN content_components cc ON cc.id = pc.component_id
WHERE p.site_id = '1244516d-014d-421c-88c6-090bb1e9552a'
  AND p.url = '/guides/index.html'
ORDER BY pc.position;

-- What the hub will resolve into `items`:
SELECT p.url, COALESCE(p.title,p.name) AS title, p.nav_label, p.nav_order
FROM pages p
WHERE p.site_id = '1244516d-014d-421c-88c6-090bb1e9552a'
  AND p.page_type = 'guide'
  AND p.status IN ('active','deployed')
  AND (p.deployed_at IS NOT NULL OR p.build_status = 'deployed')
ORDER BY COALESCE(p.nav_order,100), p.name;
