-- Teardown for gamesdesign.co.uk, PINNED TO site_id to avoid the domain race
-- that bit us repeatedly: the old domain-keyed script (a) deleted the wrong row
-- when two same-domain rows existed, and (b) raced a live cascade so the verify
-- lied (read 0 while a concurrent adoption re-inserted the row).
--
-- PRECONDITIONS (check BEFORE running):
--   1. Nothing in flight:
--        SELECT status,current_step,owner_agent_type,updated_at
--        FROM orchestration_states
--        WHERE site_id = (SELECT id FROM sites WHERE domain='gamesdesign.co.uk')
--          AND status NOT IN ('COMPLETED','FAILED');     -- must be 0 rows
--   2. Know the exact id(s):
--        SELECT id, created_at FROM sites WHERE domain='gamesdesign.co.uk';
--      If MORE THAN ONE row, run this block once per id (set :sid each time).
--   3. No adoption trigger is firing while this runs.
--
-- Usage: replace the \set below with the actual UUID, then run the file.
\set sid '00000000-0000-0000-0000-000000000000'   -- <<< SET THIS

BEGIN;

-- Step 0: confirm the target
SELECT id, domain, status, created_at FROM sites WHERE id = :'sid';

-- Step 1: break non-cascading FKs
UPDATE sites SET style_collection_id = NULL WHERE id = :'sid';
UPDATE style_collections SET css_theme_id = NULL
  WHERE source_site_id = :'sid' OR source_domain = 'gamesdesign.co.uk';

-- Step 2: non-cascading child tables (by site_id)
DELETE FROM site_work_items        WHERE site_id = :'sid';
DELETE FROM content_feed_items     WHERE site_id = :'sid';
DELETE FROM maintenance_queue      WHERE site_id = :'sid';
DELETE FROM page_component_history WHERE site_id = :'sid';
DELETE FROM link_registry          WHERE target_site_id = :'sid';

-- Step 3: the site row (cascades pages, site_specs, site_components, site_areas,
-- site_nav_*, site_snapshots, site_flows, navigation_structures,
-- research_results, redirects, products, content_items, content_sources,
-- assets, affiliate_products, link_registry via source_site_id, page_components
-- via pages). page_components/research_results/redirects cascade through pages
-- (pages_site_id_fkey ON DELETE CASCADE).
DELETE FROM sites WHERE id = :'sid';

-- Step 4: library rows attributed by domain text (no race once site is gone)
DELETE FROM style_collections WHERE source_domain = 'gamesdesign.co.uk';
DELETE FROM css_themes        WHERE source_domain = 'gamesdesign.co.uk';
DELETE FROM palettes          WHERE source_domain = 'gamesdesign.co.uk';
DELETE FROM typography_sets   WHERE source_domain = 'gamesdesign.co.uk';
DELETE FROM layouts
  WHERE source_domain = 'gamesdesign.co.uk'
    AND name NOT IN ('tool-portal-dark','social-lobby','brochure-formal',
                     'brochure-bold','technical-precise','affiliate-hub');

-- Step 5: verify (domain-based reads are fine here — they're checks, not deletes)
SELECT
  (SELECT count(*) FROM sites WHERE domain='gamesdesign.co.uk') AS sites,
  (SELECT count(*) FROM site_plans sp JOIN sites s ON s.id=sp.site_id
     WHERE s.domain='gamesdesign.co.uk') AS plans,
  (SELECT count(*) FROM style_collections WHERE source_domain='gamesdesign.co.uk') AS style_collections;

-- If sites=0 and plans=0, COMMIT. Otherwise ROLLBACK and investigate
-- (likely a second site row — re-run with its id).
COMMIT;


-- is it clean check
-- (1) Must return 0 rows — nothing running on this site
SELECT status, current_step, owner_agent_type, updated_at
FROM orchestration_states
WHERE site_id = (SELECT id FROM sites WHERE domain='gamesdesign.co.uk')
  AND status NOT IN ('COMPLETED','FAILED')
ORDER BY updated_at DESC LIMIT 5;

-- (2) How many site rows exist for this domain, and their ids
SELECT id, created_at FROM sites WHERE domain='gamesdesign.co.uk';

-- (3) Residue check — both should be 0 if a teardown already happened
SELECT
    (SELECT count(*) FROM sites WHERE domain='gamesdesign.co.uk') AS sites,
    (SELECT count(*) FROM site_plans sp JOIN sites s ON s.id=sp.site_id
WHERE s.domain='gamesdesign.co.uk') AS plans,
    (SELECT count(*) FROM orchestration_states
WHERE site_id IN (SELECT id FROM sites WHERE domain='gamesdesign.co.uk')) AS orch_rows;

