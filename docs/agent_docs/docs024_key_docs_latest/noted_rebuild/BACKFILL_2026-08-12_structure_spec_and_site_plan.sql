-- =============================================================================
-- noted.co.uk — backfill the `structure` spec and the site plan
-- 2026-08-12. Applied out of band (psql -f), NOT via the migration runner:
-- per-site setup, not a platform schema change.
-- =============================================================================
--
-- WHY THIS EXISTS — three things measured on 2026-08-12, each of which is a
-- live gap rather than a tidiness concern:
--
--  1. NO `structure` SPEC. `siteUsesFlatURLs` (site_url_shape.go) reads
--     `site_specs` aspect='structure' -> data->>'url_shape' and treats absent
--     spec, absent key, and any non-"flat" value ALL as nested. So this site's
--     URL shape has never been a decision — it has been a default nobody made.
--     It bit us the same day: the legacy-rescue tool canonicalised to
--     /tools/legacy-rescue/index.html while every content page is flat
--     (/about.html), which reads as an inconsistency and is not one.
--
--  2. THE PLAN DOES NOT KNOW ABOUT TWO PAGES THAT EXIST. `build-site-planner`
--     wrote 5 rows at 03:22 (index, how-it-works, migrate, about, contact).
--     `create_tool_component` later created `tool-legacy-rescue` and its
--     companion guide DIRECTLY, outside the plan. A plan that does not list a
--     live page is a retraction risk — the estate has a `page-retraction` agent
--     and an "assemble and deploy pages after plan reconcile" path, and
--     [INFERRED, not measured] a reconcile that treats the plan as authoritative
--     is exactly how a page nobody planned gets retired. Recording them is cheap
--     insurance; verifying that inference is not, so it is marked.
--
--  3. NO PRIVACY PAGE ANYWHERE. The owner approved privacy copy on 2026-08-12
--     (registered in evidence_base.supplied_copy.privacy) and there was NO
--     framework path to add one content page on demand — the planner made all
--     five at once, create_tool_component is tool-only, create_report_page
--     forces rebuild_policy='owned' which this lane forbids, and
--     needs_content_page writes content for a page that ALREADY EXISTS.
--     Adding the page to the PLAN is that missing path.
--
-- =============================================================================
-- WHAT THIS DELIBERATELY DOES NOT DO
-- =============================================================================
-- IT DOES NOT CHANGE ANY URL. `url_shape` is written as **"nested"**, which is
-- what this site already does. A backfill records reality; it does not move
-- live files. Flipping to "flat" is a MIGRATION and is costed at the bottom of
-- this file, unapplied.
--
-- Why that matters here specifically: site_url_shape.go carries a measured
-- warning — "upsertPage overwrites pages.url unconditionally and the deployer
-- takes the file path from it. Measured on loancalculator.co.uk 2026-08-10:
-- 24 of 26 live URLs would have moved." On noted the blast radius is smaller
-- but not zero, and it is enumerated below rather than asserted.
--
-- IT DOES NOT BUILD THE PRIVACY PAGE. It adds the page to the plan. Whether the
-- build path picks a planned-but-unbuilt page up on its own is UNMEASURED —
-- check after applying (query at the bottom) rather than assuming.
-- =============================================================================

BEGIN;

-- ---------------------------------------------------------------- guard ----
DO $$
DECLARE n int;
BEGIN
  SELECT count(*) INTO n FROM sites WHERE domain='noted.co.uk';
  IF n <> 1 THEN RAISE EXCEPTION 'expected exactly 1 noted.co.uk site row, found %', n; END IF;

  SELECT count(*) INTO n FROM site_plans sp JOIN sites s ON s.id=sp.site_id
   WHERE s.domain='noted.co.uk' AND sp.is_current;
  IF n <> 1 THEN RAISE EXCEPTION 'expected exactly 1 current site plan, found %', n; END IF;
END $$;

-- ------------------------------------------------------- structure spec ----
-- Supersede-then-insert, the same shape every SEED_*.sql uses. `pages` mirrors
-- the convention in the adopted sites' structure specs (a flat list of page
-- names); url_shape is the load-bearing key.
UPDATE site_specs SET is_current = false, superseded_at = now()
WHERE site_id = (SELECT id FROM sites WHERE domain='noted.co.uk')
  AND aspect = 'structure' AND is_current;

INSERT INTO site_specs (site_id, aspect, data, source, notes, is_current, pinned, created_by)
SELECT
  id,
  'structure',
  $st${
    "url_shape": "nested",
    "url_shape_note": "Written 2026-08-12 to RECORD the shape this site already has, not to change it. siteUsesFlatURLs treats anything other than the exact string 'flat' as nested, so this value is what the site has always done by default — it is now a decision rather than an accident. Content, landing and blog-post pages are ALWAYS flat regardless of this key (CanonicalisePage only routes tool/guide/game roles through nestedOrFlatURL), so on this site the key governs exactly one page today: tool-legacy-rescue. Changing it to 'flat' MOVES that page and is a migration, not an edit — see the costed block at the foot of BACKFILL_2026-08-12_structure_spec_and_site_plan.sql.",
    "pages": [
      "index",
      "how-it-works",
      "migrate",
      "about",
      "contact",
      "tool-legacy-rescue",
      "tool-legacy-rescue-guide",
      "privacy"
    ],
    "source": "backfill",
    "backfilled_on": "2026-08-12",
    "backfill_reason": "This site was built without a structure spec, so its URL shape was an unmade decision and the plan did not list two pages that exist. See the header of the backfill file."
  }$st$::jsonb,
  'manual',
  'Backfill 2026-08-12: records the existing nested url_shape (changes nothing) and lists all pages incl. the two created outside the plan by create_tool_component.',
  true, true, 'backfill (claude session, owner-directed)'
FROM sites WHERE domain='noted.co.uk';

-- ------------------------------------------------------- site plan rows ----
-- Added to the CURRENT plan rather than creating a new plan version: these
-- pages already exist (or, for privacy, are approved and intended), so this is
-- the plan catching up with reality, not a re-plan. A re-plan would re-derive
-- all five existing pages and is the thing to avoid.
INSERT INTO site_plan_pages (plan_id, name, role, slug, url, in_header, in_footer, nav_order, title, nav_label, meta_description)
SELECT sp.id, v.name, v.role, v.slug, v.url, v.in_header, v.in_footer, v.nav_order, v.title, v.nav_label, v.meta_description
FROM site_plans sp
JOIN sites s ON s.id = sp.site_id
CROSS JOIN (VALUES
  -- REALITY: created by create_tool_component 2026-08-12, never in the plan.
  -- URLs are copied from the live `pages` rows, NOT recomputed — the point is to
  -- record what is deployed, and a recomputation here would silently disagree
  -- with the served file.
  ('tool-legacy-rescue', 'tool', 'legacy-rescue', '/tools/legacy-rescue/index.html',
   false, true, 10,
   'Your notes from the old Noted',
   'Rescue old notes',
   'Find notes, voice recordings and photographs saved by the previous version of Noted in this browser, and save them to a file.'),

  ('tool-legacy-rescue-guide', 'blog-post', 'tool-legacy-rescue-guide', '/guides/tool-legacy-rescue-guide.html',
   false, false, 11,
   'Understanding Your notes from the old Noted',
   NULL,
   'How to bring notes from the previous version of Noted across to your account.'),

  -- INTENDED: the owner-approved privacy copy has nowhere to live. role=content
  -- canonicalises to /privacy.html on any url_shape, so this URL is stable
  -- whatever happens to the key above.
  ('privacy', 'content', 'privacy', '/privacy.html',
   false, true, 20,
   'Your notes, and what happens to them',
   'Your notes and privacy',
   'What an account is for, what you can take with you, and what happens if any of this changes.')
) AS v(name, role, slug, url, in_header, in_footer, nav_order, title, nav_label, meta_description)
WHERE s.domain = 'noted.co.uk' AND sp.is_current
  -- idempotent: re-running adds nothing
  AND NOT EXISTS (
    SELECT 1 FROM site_plan_pages x WHERE x.plan_id = sp.id AND x.name = v.name
  );

-- ---------------------------------------------------------------- assert ----
DO $$
DECLARE n int; shape text;
BEGIN
  SELECT count(*) INTO n
  FROM site_plan_pages spp JOIN site_plans sp ON sp.id=spp.plan_id JOIN sites s ON s.id=sp.site_id
  WHERE s.domain='noted.co.uk' AND sp.is_current;
  IF n <> 8 THEN RAISE EXCEPTION 'expected 8 planned pages after backfill (5 built + 2 tool + privacy), found %', n; END IF;

  SELECT ss.data->>'url_shape' INTO shape
  FROM site_specs ss JOIN sites s ON s.id=ss.site_id
  WHERE s.domain='noted.co.uk' AND ss.aspect='structure' AND ss.is_current;
  IF shape IS DISTINCT FROM 'nested' THEN RAISE EXCEPTION 'url_shape is %, expected nested', shape; END IF;

  -- the whole point of writing "nested" is that NOTHING moves
  SELECT count(*) INTO n FROM pages p JOIN sites s ON s.id=p.site_id
  WHERE s.domain='noted.co.uk' AND p.url = '/tools/legacy-rescue/index.html';
  IF n <> 1 THEN RAISE EXCEPTION 'the tool page URL is no longer /tools/legacy-rescue/index.html — something moved it'; END IF;

  RAISE NOTICE 'backfill applied: 8 planned pages, url_shape=nested, no URL moved';
END $$;

COMMIT;

-- =============================================================================
-- VERIFY (run after applying)
-- =============================================================================
SELECT spp.name, spp.role, spp.url, spp.in_header, spp.in_footer, spp.nav_order,
       (p.id IS NOT NULL) AS page_exists, p.build_status
FROM site_plan_pages spp
JOIN site_plans sp ON sp.id = spp.plan_id
JOIN sites s ON s.id = sp.site_id
LEFT JOIN pages p ON p.site_id = s.id AND p.name = spp.name
WHERE s.domain='noted.co.uk' AND sp.is_current
ORDER BY spp.nav_order NULLS LAST, spp.name;

-- Did adding a planned page produce a build item on its own? UNMEASURED — this
-- is the query that answers it. If nothing appears, the plan->build step needs
-- driving; do NOT hand-create a `pages` row instead (bugs_open/080).
SELECT wi.item_type, wi.status, wi.spec->>'page_name' AS page, wi.created_at
FROM site_work_items wi JOIN sites s ON s.id = wi.site_id
WHERE s.domain='noted.co.uk' AND wi.created_at > now() - interval '30 minutes'
ORDER BY wi.created_at DESC;

-- =============================================================================
-- NOT APPLIED — the "flat" migration, costed so the next session need not
-- rediscover it
-- =============================================================================
-- Flipping url_shape to "flat" would make the tool page canonicalise to
-- /tools/legacy-rescue.html instead of /tools/legacy-rescue/index.html.
--
-- Blast radius on THIS site, enumerated (not asserted): exactly ONE page.
-- CanonicalisePage routes only tool/guide/game roles through nestedOrFlatURL;
-- content, landing and blog-post are flat on every shape. noted has one tool
-- page and its guide is role=blog-post, so the guide does not move either.
--
--   SELECT p.name, p.url, p.page_type FROM pages p JOIN sites s ON s.id=p.site_id
--   WHERE s.domain='noted.co.uk' AND p.page_type IN ('tool','guide','game');
--
-- What would have to happen together, in this order:
--   1. flip url_shape to 'flat';
--   2. re-canonicalise / rebuild the tool page so pages.url follows;
--   3. re-point migrate's two CTAs — they hard-code the current URL in
--      page_components.content_data and would otherwise link to a 404:
--        scripts/initial_messages/130_section_editor/074_section_editor_noted_cta_urls.sh migrate
--      (edit its EDITS entries to the new URL first);
--   4. delete or redirect the orphaned /tools/legacy-rescue/index.html on the
--      box and in gqls/vm-sites — sitesync rsyncs --delete from the repo, so
--      removing it from the repo is sufficient and removing it ONLY from the box
--      is undone within 5 minutes.
--
-- WHEN: cheapest now, while the framework build is NOT public (the apex still
-- serves the legacy app from B2). After cutover this is a live URL change with
-- a real cost, so if flat is wanted, want it before cutover.
