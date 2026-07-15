-- migration_cleanup_bare_guide_duplicates.sql
-- ----------------------------------------------------------------------------
-- Remove the 5 SPURIOUS bare blog-post pages on gamesdesign.co.uk:
--   economy-basics, fairness-in-rng, p2p-architecture, rng-design, skinner-box
--
-- Provenance (confirmed): created 2026-06-03 20:25:30 by a post-adoption planner
-- pass (~6h after the 14:36:50 adoption batch), page_type=blog-post,
-- sections=[], build_status=planned, n_rendered=0 — created, never built, no
-- content. They duplicate the faithful adopted guide-<topic> pages (now typed
-- `guide`, at /guides/<slug>/index.html). This is the documented
-- "planner ignores adopted state" pattern (FOCUS_planner_ignores_adopted_state.md
-- / doc 029): a second surface invents parallel pages after adoption.
--
-- Confirmed inert: not guide-typed (so absent from guide-list), and 0 rows in
-- link_registry for their /blog/<topic>.html urls.
--
-- This migration is DURABLE: it also removes them from the CURRENT site_plan
-- (site_plan_pages + site_plan_sections) so reconcile_site_plan will NOT
-- re-create them, and terminalises any open needs_page work items
-- (site_work_items.page_id has NO FK to pages, so those would otherwise linger
-- and hold the idx_swi_dedup slot).
--
-- It does NOT stop build-site-planner / blog-content-planner re-inventing them
-- on a FUTURE plan_site run — that upstream gap is tracked separately (identify
-- the creating surface via query A1, then tighten its prompt/logic).
--
-- FK NOTES (checked): a pages delete CASCADES page_components,
-- link_registry.source_page_id, flow_pages, research_results; SET-NULLs
-- site_nav_items. It BLOCKS on the non-cascading refs link_registry.target_page_id,
-- page_component_history.page_id, redirects.source_page_id — the DO-block guard
-- below aborts the txn if any of those exist for these pages.
--
-- Data-only. Effective on COMMIT.
-- ----------------------------------------------------------------------------

BEGIN;

-- ---- GUARD: abort if any non-cascading reference to the 5 bare pages exists ----
DO $$
DECLARE
  blocking int;
BEGIN
  WITH bare AS (
    SELECT id FROM pages
    WHERE site_id = (SELECT id FROM sites WHERE domain = 'gamesdesign.co.uk')
      AND page_type = 'blog-post'
      AND name IN ('economy-basics','fairness-in-rng','p2p-architecture','rng-design','skinner-box')
  )
  SELECT
      (SELECT count(*) FROM link_registry        WHERE target_page_id IN (SELECT id FROM bare))
    + (SELECT count(*) FROM page_component_history WHERE page_id        IN (SELECT id FROM bare))
    + (SELECT count(*) FROM redirects             WHERE source_page_id  IN (SELECT id FROM bare))
  INTO blocking;

  IF blocking > 0 THEN
    RAISE EXCEPTION 'Aborting: % non-cascading reference(s) to the bare pages exist (link_registry.target_page_id / page_component_history / redirects). Resolve those first.', blocking;
  END IF;
END $$;

-- ---- SNAPSHOTS (rollback safety) ----
-- 1) the pages rows themselves
CREATE TABLE IF NOT EXISTS pages_bak_del_bare AS
SELECT * FROM pages
WHERE site_id = (SELECT id FROM sites WHERE domain = 'gamesdesign.co.uk')
  AND page_type = 'blog-post'
  AND name IN ('economy-basics','fairness-in-rng','p2p-architecture','rng-design','skinner-box');

-- 2) current-plan plan-page rows for these names
CREATE TABLE IF NOT EXISTS splanpages_bak_del_bare AS
SELECT spp.* FROM site_plan_pages spp
JOIN site_plans sp ON sp.id = spp.plan_id AND sp.is_current = true
WHERE sp.site_id = (SELECT id FROM sites WHERE domain = 'gamesdesign.co.uk')
  AND spp.name IN ('economy-basics','fairness-in-rng','p2p-architecture','rng-design','skinner-box');

-- 3) current-plan plan-section rows for these page_names (orphan cleanup)
CREATE TABLE IF NOT EXISTS splansecs_bak_del_bare AS
SELECT sps.* FROM site_plan_sections sps
JOIN site_plans sp ON sp.id = sps.plan_id AND sp.is_current = true
WHERE sp.site_id = (SELECT id FROM sites WHERE domain = 'gamesdesign.co.uk')
  AND sps.page_name IN ('economy-basics','fairness-in-rng','p2p-architecture','rng-design','skinner-box');

-- 4) work items that reference these pages (by page_id or by needs_page key)
CREATE TABLE IF NOT EXISTS swi_bak_del_bare AS
SELECT * FROM site_work_items
WHERE site_id = (SELECT id FROM sites WHERE domain = 'gamesdesign.co.uk')
  AND (
        page_id IN (SELECT id FROM pages
                    WHERE site_id = (SELECT id FROM sites WHERE domain = 'gamesdesign.co.uk')
                      AND name IN ('economy-basics','fairness-in-rng','p2p-architecture','rng-design','skinner-box'))
     OR item_key IN ('needs_page:economy-basics','needs_page:fairness-in-rng',
                     'needs_page:p2p-architecture','needs_page:rng-design','needs_page:skinner-box')
      );

-- Before counts (for the record).
SELECT
  (SELECT count(*) FROM pages_bak_del_bare)        AS pages_to_delete,
  (SELECT count(*) FROM splanpages_bak_del_bare)   AS plan_pages_to_delete,
  (SELECT count(*) FROM splansecs_bak_del_bare)    AS plan_sections_to_delete,
  (SELECT count(*) FROM swi_bak_del_bare)          AS work_items_to_terminalise;

-- ---- 1) Terminalise the work items (no FK to pages; do this while page_id resolves) ----
UPDATE site_work_items
SET status     = 'wont_fix',
    error      = 'Spurious duplicate page removed (bare blog-post sibling of an adopted guide); see migration_cleanup_bare_guide_duplicates.sql',
    updated_at = NOW()
WHERE id IN (SELECT id FROM swi_bak_del_bare)
  AND status NOT IN ('complete','verified','rejected','wont_fix','failed','unresolved');

-- ---- 2) Remove from the current plan (so the reconciler won't re-create) ----
DELETE FROM site_plan_sections
WHERE id IN (SELECT id FROM splansecs_bak_del_bare);

DELETE FROM site_plan_pages
WHERE id IN (SELECT id FROM splanpages_bak_del_bare);

-- ---- 3) Delete the page rows (cascades page_components/link_registry.source/
--          flow_pages/research_results; SET-NULLs site_nav_items) ----
DELETE FROM pages
WHERE id IN (SELECT id FROM pages_bak_del_bare);

-- After: the bare names should be gone; the guide-<topic> rows remain.
SELECT name, page_type, url, build_status
FROM pages
WHERE site_id = (SELECT id FROM sites WHERE domain = 'gamesdesign.co.uk')
  AND (name LIKE 'guide-%'
       OR name IN ('economy-basics','fairness-in-rng','p2p-architecture','rng-design','skinner-box'))
ORDER BY name;

COMMIT;

-- ----------------------------------------------------------------------------
-- ROLLBACK (if needed) — re-insert from the snapshots (run inside one txn):
--   INSERT INTO pages             SELECT * FROM pages_bak_del_bare;
--   INSERT INTO site_plan_pages   SELECT * FROM splanpages_bak_del_bare;
--   INSERT INTO site_plan_sections SELECT * FROM splansecs_bak_del_bare;
--   UPDATE site_work_items w SET status = b.status, error = b.error, updated_at = NOW()
--     FROM swi_bak_del_bare b WHERE w.id = b.id;
-- (Note: cascaded child rows — page_components etc. — are NOT restored by this;
--  these pages had none worth keeping. If that ever matters, snapshot children too.)
-- Drop snapshots when satisfied:
--   DROP TABLE pages_bak_del_bare, splanpages_bak_del_bare, splansecs_bak_del_bare, swi_bak_del_bare;
-- ----------------------------------------------------------------------------
