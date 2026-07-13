-- migration_rebuild_list_pages.sql
-- ----------------------------------------------------------------------------
-- THREAD A — re-render the pages that carry a guide-list (and game-list)
-- section so they pick up the now-correctly-typed guide pages
-- (page_type='guide', status active) and game pages.
--
-- WHY THIS WORKS: guide-list resolves its items at plan_sections time via
-- query.pages_where_type:guide (resolvePagesWhereType selects pages WHERE
-- page_type=$2 AND status IN ('active','deployed')). The guides are now typed
-- 'guide' and were adopted as active, so a rebuild of the page holding the
-- guide-list re-runs the resolver and the cards appear.
--
-- guide-list is already Tier-D (items.source = query.pages_where_type:guide),
-- so it populates on rebuild. game-list populates on rebuild ONLY IF its
-- component is also Tier-D — the pre-check below reports each component's
-- source. If game-list is still flat (post1_* style), apply
-- migration_game_list_tier_d.sql FIRST, then re-run this flip.
--
-- Self-targeting and robust to versioned section names: it matches any page
-- whose sections array contains an element starting "guide-list" or
-- "game-list" (so it works whether sections stores "guide-list" or
-- "guide-list_pre_037"). Setting build_status='needs_rebuild' only queues a
-- rebuild; it is harmless to pages that don't match.
-- ----------------------------------------------------------------------------

BEGIN;

-- (0) Pre-check: which list components are Tier-D (query-sourced) yet?
--     query_sourced=t means a rebuild will populate that list from the pages
--     table; query_sourced=f means it is still flat and a rebuild will NOT
--     fix it (apply the component's tier-D migration first).
SELECT name, component_level,
       (input_schema::text LIKE '%query.pages_where_type%') AS query_sourced
FROM content_components
WHERE name LIKE 'guide-list%'
   OR name LIKE 'game-list%'
   OR name LIKE 'tool-list%'
ORDER BY name;

-- (1) Snapshot the pages about to be flipped (rollback safety).
CREATE TABLE IF NOT EXISTS pages_bak_relist AS
SELECT * FROM pages
WHERE site_id = (SELECT id FROM sites WHERE domain = 'gamesdesign.co.uk')
  AND jsonb_typeof(sections) = 'array'
  AND EXISTS (
        SELECT 1 FROM jsonb_array_elements_text(sections) AS s
        WHERE s LIKE 'guide-list%' OR s LIKE 'game-list%'
      );

-- (2) Show exactly what will be flipped (and their current sections).
SELECT name, page_type, build_status, sections
FROM pages
WHERE site_id = (SELECT id FROM sites WHERE domain = 'gamesdesign.co.uk')
  AND jsonb_typeof(sections) = 'array'
  AND EXISTS (
        SELECT 1 FROM jsonb_array_elements_text(sections) AS s
        WHERE s LIKE 'guide-list%' OR s LIKE 'game-list%'
      )
ORDER BY name;

-- (3) Flip to needs_rebuild.
UPDATE pages
SET build_status = 'needs_rebuild'
WHERE site_id = (SELECT id FROM sites WHERE domain = 'gamesdesign.co.uk')
  AND jsonb_typeof(sections) = 'array'
  AND EXISTS (
        SELECT 1 FROM jsonb_array_elements_text(sections) AS s
        WHERE s LIKE 'guide-list%' OR s LIKE 'game-list%'
      );

COMMIT;

-- ----------------------------------------------------------------------------
-- ROLLBACK (restore prior build_status from the snapshot):
--   UPDATE pages p
--   SET build_status = b.build_status
--   FROM pages_bak_relist b WHERE p.id = b.id;
-- Drop snapshot once satisfied: DROP TABLE pages_bak_relist;
--
-- NOTE: if step (2) returns no rows, the list sections are stored under a
-- string the LIKE didn't catch — send me the sections values and I'll adjust
-- the match (no harm done; nothing was flipped).
-- ----------------------------------------------------------------------------
