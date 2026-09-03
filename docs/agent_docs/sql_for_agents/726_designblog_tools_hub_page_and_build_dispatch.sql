-- 726_designblog_tools_hub_page_and_build_dispatch.sql
--
-- OWNER RULING 2026-09-03 (designblog.co.uk session, decision 7: "do it now"): designblog's
-- nav has no Tools link while four tool pages serve. Root shape is SITE_DEFECT_CATEGORIES
-- 2.5 + 2.6, both measured on this site 2026-09-03: all four tool pages ALREADY carry
-- in_header=true with nav_order 1-4, and the page_type='tool' bar keeps them out of the
-- primary nav (correctly-configured rows that never render); nothing planned a tools hub, so
-- the family has no nav path at all. The estate remedy (portfolio positioning, the builder):
-- a tools section-index page with in_header set, then a nav rebuild.
--
-- This migration does the two data halves ON ONE SITE (aa51d9b8, designblog.co.uk):
--   (1) adds the tools hub to the live site plan (plan a265bb7c, the 2026-09-02 remake plan),
--       shaped exactly like its sibling section-index rows (criticism-index / feed-index),
--       nav_order 7 (visible nav currently ends at 6 = glossary; header max is 5-8 items and
--       the four barred tool rows do not render, so the visible count goes 6 -> 7);
--   (2) files the needs_page build item (page-build-handler, the default handler --
--       section-index is deliberately NOT in builderForPageType per the 206 lane, so the
--       default IS the correct route), item_key under idx_swi_dedup so a duplicate dispatch
--       cannot stack.
-- The nav link itself appears at the next chrome re-render after the page deploys -- the
-- pending GTM stale_chrome wave on this site covers that, or a later chrome rerender does.
-- The 444 gate's child-count arm resolves this page as PRODUCIBLE (four tool children exist
-- under /tools/), so this does not create another empty-listing page: its listing has four
-- real children on day one.
-- NOT in this migration, deliberately: the header pin (theme kits corrected their own
-- available-now advice 2026-09-03 -- the alternative headers need ~12 content_data variables
-- this site does not carry; pinning risks a broken header on a live site; held for the owner).
--
-- Apply: psql -f THIS FILE ONLY. Companion ROLLBACK removes the plan row and cancels the
-- item if still non-terminal (it does NOT touch a built page).

BEGIN;

DO $$
DECLARE n int;
BEGIN
  SELECT count(*) INTO n FROM sites WHERE id='aa51d9b8-511a-4bda-8207-a7e65c3abc4a' AND domain='designblog.co.uk';
  IF n <> 1 THEN RAISE EXCEPTION '726 REFUSED: designblog site row not found under expected id'; END IF;

  SELECT count(*) INTO n FROM site_plans WHERE id='a265bb7c-f6c1-4a97-bbd2-4065d5375675' AND site_id='aa51d9b8-511a-4bda-8207-a7e65c3abc4a';
  IF n <> 1 THEN RAISE EXCEPTION '726 REFUSED: expected plan a265bb7c not found for designblog — the plan may have been superseded; re-derive the live plan id before applying'; END IF;

  SELECT count(*) INTO n FROM site_plan_pages WHERE plan_id='a265bb7c-f6c1-4a97-bbd2-4065d5375675' AND name='tools-index';
  IF n <> 0 THEN RAISE EXCEPTION '726 REFUSED: tools-index already planned (found %) — nothing to do', n; END IF;

  SELECT count(*) INTO n FROM pages WHERE site_id='aa51d9b8-511a-4bda-8207-a7e65c3abc4a' AND name='tools-index';
  IF n <> 0 THEN RAISE EXCEPTION '726 REFUSED: a tools-index pages row already exists — plan and pages have diverged; investigate before applying'; END IF;
END $$;

INSERT INTO site_plan_pages (plan_id, name, role, slug, url, parent_section, in_header, in_footer, nav_order, title, meta_description, nav_label)
VALUES (
  'a265bb7c-f6c1-4a97-bbd2-4065d5375675',
  'tools-index',
  'section-index',
  'tools',
  '/tools/index.html',
  NULL,
  true,
  true,
  7,
  'Design Tools | Design Blog',
  'Free design tools: a WCAG contrast checker, CSS variable architect, unit converter and aspect-ratio calculator for working interface designers.',
  'Tools'
);

INSERT INTO site_work_items (site_id, item_type, item_key, status, priority, summary, spec, handler_agent, source, created_by)
SELECT
  'aa51d9b8-511a-4bda-8207-a7e65c3abc4a',
  'needs_page',
  'needs_page:tools-index',
  'triaged',
  10,
  'Build /tools/index.html (tools hub, section-index) — owner ruling 2026-09-03 decision 7: the four tool pages are type-barred from primary nav (SITE_DEFECT_CATEGORIES 2.5/2.6); the hub is the nav path. Planned in migration 726; nav link follows at next chrome re-render.',
  '{"reason": "new_page_planned", "handler": "page-build-handler", "page_name": "tools-index", "page_role": "section-index"}'::jsonb,
  'page-build-handler',
  'operator',
  'designblog-couk-session-2026-09-03'
WHERE NOT EXISTS (
  SELECT 1 FROM site_work_items
  WHERE site_id='aa51d9b8-511a-4bda-8207-a7e65c3abc4a'
    AND item_key='needs_page:tools-index'
    AND status NOT IN ('complete','verified','rejected','wont_fix','failed','unresolved','cancelled')
);

-- Verify (DO/RAISE): plan row present exactly once; exactly one live work item.
DO $$
DECLARE n int;
BEGIN
  SELECT count(*) INTO n FROM site_plan_pages WHERE plan_id='a265bb7c-f6c1-4a97-bbd2-4065d5375675' AND name='tools-index' AND role='section-index' AND in_header AND nav_label='Tools';
  IF n <> 1 THEN RAISE EXCEPTION '726 VERIFY: plan row not present exactly once (found %)', n; END IF;

  SELECT count(*) INTO n FROM site_work_items
   WHERE site_id='aa51d9b8-511a-4bda-8207-a7e65c3abc4a' AND item_key='needs_page:tools-index'
     AND status NOT IN ('complete','verified','rejected','wont_fix','failed','unresolved','cancelled');
  IF n <> 1 THEN RAISE EXCEPTION '726 VERIFY: expected exactly 1 live needs_page item, found %', n; END IF;

  RAISE NOTICE '726 OK: tools hub planned (nav_order 7, Tools) and build item filed on page-build-handler.';
END $$;

COMMIT;
