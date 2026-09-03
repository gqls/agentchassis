-- 732_designblog_tools_hub_pages_row_and_refile.sql
--
-- COMPLETES 726, whose dispatch half was built on a wrong premise (mine — diagnosed at the
-- config, 2026-09-03): page-build-handler CANNOT CREATE a page. Its `check_page_found` step
-- routes `page_record.found == false` to `complete_error` (its own description: "audit
-- findings for new pages will skip here"), so 726's needs_page item for the not-yet-existing
-- tools-index completed at 10:30:20Z having built NOTHING — a complete work item is not a
-- changed artefact, at my own dispatch. Page CREATION belongs to the plan pipeline's
-- sync_pages — and a full replan is deliberately NOT used here: gamedesign's same-day canary
-- exposed validate_site_plan Pass C dropping any NEW page whose slug matches a realised
-- section stem (slugOf returns the first path segment), and tools-index (slug "tools", four
-- realised tool pages under /tools/) is squarely that shape. So this migration materialises
-- what sync_pages would have written, surgically:
--   (1) the THREE site_plan_sections rows 726 omitted (the placement landmine: a section
--       placement lives in three places; missing one gets it dropped by a complete
--       re-render) — composition hero / tool-list / call-to-action, the measured working
--       pattern for tools hubs (7 of 10 deployed hubs fleet-wide use tool-list; the
--       section-index twin gamesdesign.co.uk/tools-index uses exactly this shape). NOT the
--       sibling section-indexes' fallback trio (hero/generic-text-block/call-to-action) —
--       that is the manifesto-page composition this critique started from; tool-list is the
--       component that actually lists the site's tools.
--   (2) the pages row (build_status 'planned' — the vocabulary's pre-build value),
--   (3) a FRESH needs_page item under the same key (the 10:30Z item is terminal, so
--       idx_swi_dedup admits a new one; NOT EXISTS guard kept for the non-terminal case).
-- With the page found, page-build-handler's normal path (plan_sections -> writer ->
-- save_sections -> deploy) applies. The Tools nav link then renders at the next chrome
-- re-render (the pending GTM stale_chrome wave).
--
-- Apply: psql -f THIS FILE ONLY. Companion ROLLBACK refuses once the page has components.

BEGIN;

DO $$
DECLARE n int;
BEGIN
  SELECT count(*) INTO n FROM site_plan_pages
   WHERE plan_id='a265bb7c-f6c1-4a97-bbd2-4065d5375675' AND name='tools-index';
  IF n <> 1 THEN RAISE EXCEPTION '732 REFUSED: 726 plan row missing (found %) — apply 726 first', n; END IF;

  SELECT count(*) INTO n FROM pages
   WHERE site_id='aa51d9b8-511a-4bda-8207-a7e65c3abc4a' AND name='tools-index';
  IF n <> 0 THEN RAISE EXCEPTION '732 REFUSED: pages row already exists (found %) — nothing to do', n; END IF;

  SELECT count(*) INTO n FROM site_plan_sections
   WHERE plan_id='a265bb7c-f6c1-4a97-bbd2-4065d5375675' AND page_name='tools-index';
  IF n <> 0 THEN RAISE EXCEPTION '732 REFUSED: plan sections already present (found %) — investigate', n; END IF;
END $$;

INSERT INTO site_plan_sections (plan_id, page_name, ordering, component_name)
VALUES
  ('a265bb7c-f6c1-4a97-bbd2-4065d5375675', 'tools-index', 0, 'hero'),
  ('a265bb7c-f6c1-4a97-bbd2-4065d5375675', 'tools-index', 1, 'tool-list'),
  ('a265bb7c-f6c1-4a97-bbd2-4065d5375675', 'tools-index', 2, 'call-to-action');

INSERT INTO pages (site_id, name, title, meta_description, page_type, url, nav_label,
                   nav_order, in_header, in_footer, sections, status, build_status, rebuild_policy)
VALUES (
  'aa51d9b8-511a-4bda-8207-a7e65c3abc4a',
  'tools-index',
  'Design Tools | Design Blog',
  'Free design tools: a WCAG contrast checker, CSS variable architect, unit converter and aspect-ratio calculator for working interface designers.',
  'section-index',
  '/tools/index.html',
  'Tools',
  7,
  true,
  true,
  '["hero", "tool-list", "call-to-action"]'::jsonb,
  'active',
  'planned',
  'generic'
);

INSERT INTO site_work_items (site_id, item_type, item_key, status, priority, summary, spec, handler_agent, source, created_by)
SELECT
  'aa51d9b8-511a-4bda-8207-a7e65c3abc4a',
  'needs_page',
  'needs_page:tools-index',
  'triaged',
  10,
  'Build /tools/index.html (tools hub) — RE-FILE of 726''s item, which completed without building: page-build-handler cannot create a page (check_page_found -> complete_error). The pages row now exists (migration 732, composition hero/tool-list/call-to-action); normal build path applies.',
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

DO $$
DECLARE n int;
BEGIN
  SELECT count(*) INTO n FROM pages
   WHERE site_id='aa51d9b8-511a-4bda-8207-a7e65c3abc4a' AND name='tools-index'
     AND page_type='section-index' AND in_header AND sections = '["hero", "tool-list", "call-to-action"]'::jsonb;
  IF n <> 1 THEN RAISE EXCEPTION '732 VERIFY: pages row not present exactly once with expected shape (found %)', n; END IF;

  SELECT count(*) INTO n FROM site_plan_sections
   WHERE plan_id='a265bb7c-f6c1-4a97-bbd2-4065d5375675' AND page_name='tools-index';
  IF n <> 3 THEN RAISE EXCEPTION '732 VERIFY: expected 3 plan section rows, found %', n; END IF;

  SELECT count(*) INTO n FROM site_work_items
   WHERE site_id='aa51d9b8-511a-4bda-8207-a7e65c3abc4a' AND item_key='needs_page:tools-index'
     AND status NOT IN ('complete','verified','rejected','wont_fix','failed','unresolved','cancelled');
  IF n <> 1 THEN RAISE EXCEPTION '732 VERIFY: expected exactly 1 live needs_page item, found %', n; END IF;

  RAISE NOTICE '732 OK: tools-index exists in all three places; build item re-filed.';
END $$;

COMMIT;
