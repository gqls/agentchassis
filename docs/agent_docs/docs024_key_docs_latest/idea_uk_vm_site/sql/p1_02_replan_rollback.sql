-- Phase 1 rollback — undo the re-plan's damage, keep its two wins.
--
-- WHAT THE RE-PLAN DID (plan ff03bdef, 2026-07-14 16:00:54):
--   ✅ composed guides-index and news-index (2 sections each)     <- KEEP
--   ❌ left tool-audience-check uncomposed (0 sections)            <- separate route
--   ⚠️  invented 10 pages (5 tools + 5 blog posts), all uncomposed <- DELETE
--   ⚠️  regressed 4 built pages by re-proposing them with different sections:
--         index   dropped info-card-grid
--         about   dropped hero-about + info-card-grid, gained generic hero
--         contact dropped hero-contact, gained generic hero
--         report  dropped generic-text-block + info-card-grid
--       and shuffled nav metadata (index.in_header f->t, contact.in_header t->f,
--       privacy.nav_order 90->190).                                <- RESTORE
--
-- WHY IT HAPPENED (the handoff's premise is wrong, fleet-wide):
--   normaliseRealisedToPlanPage's union (v3_site_actions.go:4630) only protects pages the
--   LLM does NOT re-propose. For pages it DOES re-propose, the LLM's section list wins and
--   the realised composition is overwritten. reconcilePlanWithRealised force-preserves only
--   ADOPTION-LOCKED pages, and idea.uk's are not adopted. Separately, the union carries
--   sections=[] forward faithfully, so a catalogued-but-uncomposed page is preserved AS
--   EMPTY — a re-plan can never compose it. That is why tool-audience-check is still empty.
--
-- Backups taken first: _ideauk_bak_20260714_{pages,plan_pages,plan_sections,work_items}.
-- Note those are POST-run snapshots; the authoritative PRE-run state is the old plan
-- 32be2797 (its 21 sections match the pre-run counts exactly: index 6, about 4, report 4,
-- contact 3, tools 3, privacy 1).

\set ON_ERROR_STOP on
BEGIN;

\set SID '1244516d-014d-421c-88c6-090bb1e9552a'
\set OLD_PLAN '32be2797-e60b-44e1-812b-16eadc076c61'
\set NEW_PLAN 'ff03bdef-3bb2-40eb-93ff-efa70f46b6b8'

-- The 9 pages that should exist. Anything else is an invention of the re-plan.
CREATE TEMP TABLE keep_pages(name text) ON COMMIT DROP;
INSERT INTO keep_pages VALUES
  ('index'),('tools'),('report'),('guides-index'),('news-index'),
  ('about'),('contact'),('tool-audience-check'),('privacy');

-- The 6 pages that were already built, whose composition must be restored verbatim.
CREATE TEMP TABLE built_pages(name text) ON COMMIT DROP;
INSERT INTO built_pages VALUES
  ('index'),('tools'),('report'),('about'),('contact'),('privacy');

-- ── 1. Restore the 6 built pages' sections in the CURRENT plan, from the old plan ──
DELETE FROM site_plan_sections
WHERE plan_id = :'NEW_PLAN' AND page_name IN (SELECT name FROM built_pages);

INSERT INTO site_plan_sections
  (plan_id, page_name, ordering, component_name, component_version_id, palette_id, layout_id, typography_set_id)
SELECT :'NEW_PLAN', page_name, ordering, component_name, component_version_id, palette_id, layout_id, typography_set_id
FROM site_plan_sections
WHERE plan_id = :'OLD_PLAN' AND page_name IN (SELECT name FROM built_pages);

-- ── 2. Drop the 10 invented pages from the current plan ──
DELETE FROM site_plan_pages
WHERE plan_id = :'NEW_PLAN' AND name NOT IN (SELECT name FROM keep_pages);

-- ── 3. Restore nav metadata on the current plan's pages, from the old plan ──
UPDATE site_plan_pages n
SET in_header = o.in_header, in_footer = o.in_footer, nav_order = o.nav_order
FROM site_plan_pages o
WHERE n.plan_id = :'NEW_PLAN' AND o.plan_id = :'OLD_PLAN' AND n.name = o.name;

-- ── 4. Restore pages.sections for the 6 built pages (ordered array of component names) ──
UPDATE pages p
SET sections = s.arr, updated_at = NOW()
FROM (
  SELECT page_name, jsonb_agg(component_name ORDER BY ordering) AS arr
  FROM site_plan_sections
  WHERE plan_id = :'OLD_PLAN' AND page_name IN (SELECT name FROM built_pages)
  GROUP BY page_name
) s
WHERE p.site_id = :'SID' AND p.name = s.page_name;

-- ── 5. Restore pages nav metadata to the PRE-RUN values (captured before the re-plan) ──
UPDATE pages p SET in_header = v.in_header, nav_order = v.nav_order, updated_at = NOW()
FROM (VALUES
  ('index',               false,  1),
  ('tools',               true,   2),
  ('report',              true,   3),
  ('guides-index',        true,   4),
  ('news-index',          true,   5),
  ('about',               true,   6),
  ('contact',             true,   7),
  ('tool-audience-check', false, 30),
  ('privacy',             false, 90)
) AS v(name, in_header, nav_order)
WHERE p.site_id = :'SID' AND p.name = v.name;

-- ── 6. Delete the 10 invented pages and their work items ──
DELETE FROM site_work_items
WHERE site_id = :'SID'
  AND item_type = 'needs_page'
  AND spec->>'page_name' NOT IN (SELECT name FROM keep_pages);

-- the plan-wide reconcile rerender would try to assemble all 11 uncomposed pages
DELETE FROM site_work_items
WHERE site_id = :'SID' AND item_type = 'needs_rerender' AND status = 'detected';

DELETE FROM pages
WHERE site_id = :'SID' AND name NOT IN (SELECT name FROM keep_pages);

-- ── 7. Re-triage only what genuinely needs building ──
-- about + contact were REBUILT AND DEPLOYED with the regressed sections (16:07, 16:12),
-- so their B2 copies are degraded and must be rebuilt from the restored composition.
-- guides-index + news-index have never been built (these are the two wins).
UPDATE site_work_items SET status = 'triaged', updated_at = NOW()
WHERE site_id = :'SID' AND item_type = 'needs_page' AND status = 'detected'
  AND spec->>'page_name' IN ('about','contact','guides-index','news-index');

-- index/tools/report/privacy were NOT rebuilt, so their deployed B2 copies still match the
-- now-restored composition. Nothing to do — retire their items rather than churn the site.
UPDATE site_work_items SET status = 'wont_fix', updated_at = NOW()
WHERE site_id = :'SID' AND item_type = 'needs_page' AND status IN ('detected','triaged')
  AND spec->>'page_name' IN ('index','tools','report','privacy');

-- tool-audience-check stays PAUSED at 'detected' until it is actually composed.

-- imagery: heroes for the two new pages + genuine gaps on index. Let these run.
UPDATE site_work_items SET status = 'triaged', updated_at = NOW()
WHERE site_id = :'SID' AND item_type = 'needs_imagery' AND status = 'detected';

COMMIT;

-- ── VERIFY ───────────────────────────────────────────────────────────────────
\echo '=== pages after rollback (expect 9; 6 built restored; 2 composed; 1 empty) ==='
SELECT p.name, p.page_type, p.build_status, p.in_header, p.nav_order,
       jsonb_array_length(COALESCE(p.sections,'[]'::jsonb)) AS n_sections
FROM pages p WHERE p.site_id = :'SID' ORDER BY p.nav_order;

\echo '=== restored sections match the old plan exactly? (expect 0 diffs) ==='
WITH old AS (SELECT page_name, ordering, component_name FROM site_plan_sections WHERE plan_id = :'OLD_PLAN'
             AND page_name IN ('index','tools','report','about','contact','privacy')),
     cur AS (SELECT page_name, ordering, component_name FROM site_plan_sections WHERE plan_id = :'NEW_PLAN'
             AND page_name IN ('index','tools','report','about','contact','privacy'))
SELECT COALESCE(o.page_name,c.page_name) AS page, COALESCE(o.component_name,c.component_name) AS component,
       CASE WHEN c.component_name IS NULL THEN 'MISSING' ELSE 'EXTRA' END AS diff
FROM old o FULL OUTER JOIN cur c
  ON o.page_name=c.page_name AND o.ordering=c.ordering AND o.component_name=c.component_name
WHERE o.component_name IS NULL OR c.component_name IS NULL;

\echo '=== work queue after rollback ==='
SELECT item_type, status, count(*) FROM site_work_items
WHERE site_id = :'SID' AND status NOT IN ('complete','verified','rejected','wont_fix')
GROUP BY 1,2 ORDER BY 1,2;
