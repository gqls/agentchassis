-- migration_guides_url_to_canonical.sql
-- ----------------------------------------------------------------------------
-- STRUCTURAL (optional) — move the 5 guide pages from their blog-post-origin
-- URLs (/blog/guide-<slug>.html) to the canonical guide shape that
-- datahelpers.CanonicalisePage produces for role=guide:
--
--     /guides/<slug>/index.html        (slug = name with the "guide-" prefix dropped)
--
-- This makes guides a peer of tools (/tools/<slug>/index.html) and games
-- (/games/<slug>/index.html), and lets the guides hub at /guides/index.html
-- contain its children. NAME is unchanged (guide-<slug> is already canonical);
-- only url + a rebuild flip change here.
--
-- PRE-REQUISITE: run migration_retype_guides_to_guide.sql first (these rows must
-- already be page_type='guide').
--
-- BLAST RADIUS (handle via the separate steps at the foot, sized by the three
-- diagnostics in the handoff):
--   - the 5 guides rebuild+redeploy at the new paths (flip below)
--   - the OLD /blog/guide-*.html files become orphaned in the repo (the deployer
--     writes the new path; it does NOT delete the old file) — delete them as a
--     git cleanup
--   - pages whose rendered_html hardcodes /blog/guide-* links go stale until they
--     re-render (diagnostic c) — rebuild those too
--   - guide-list links are dynamic (resolved from pages.url by queryresolve), so
--     they pick up the new URLs automatically on the homepage's next rebuild
--
-- SCHEMA NOTE: pages.url has no format CHECK; the deployer already writes nested
-- <section>/<slug>/index.html paths (tools and games use them), so /guides/<slug>/
-- index.html is a supported output path.
--
-- Data-only change. Effective on COMMIT; files move on the subsequent rebuild.
-- ----------------------------------------------------------------------------

BEGIN;

-- SNAPSHOT (rollback safety): full backup of the rows being changed (captures
-- url, build_status, built_from_plan_version before the edit).
CREATE TABLE IF NOT EXISTS pages_bak_guides_url AS
SELECT * FROM pages
WHERE site_id = (SELECT id FROM sites WHERE domain = 'gamesdesign.co.uk')
  AND page_type = 'guide'
  AND name LIKE 'guide-%'
  AND url LIKE '/blog/guide-%';

-- Before: current url + the computed canonical url (verify the mapping looks
-- right; expect 5 rows).
SELECT name,
       url AS current_url,
       '/guides/' || substring(name from 7) || '/index.html' AS canonical_url,
       build_status
FROM pages
WHERE site_id = (SELECT id FROM sites WHERE domain = 'gamesdesign.co.uk')
  AND page_type = 'guide'
  AND name LIKE 'guide-%'
  AND url LIKE '/blog/guide-%'
ORDER BY name;

-- Move the URL to canonical and flip to rebuild so the page redeploys at the
-- new path. substring(name from 7) drops the 6-char "guide-" prefix.
UPDATE pages
SET url                    = '/guides/' || substring(name from 7) || '/index.html',
    build_status           = 'needs_rebuild',
    built_from_plan_version = NULL
WHERE site_id = (SELECT id FROM sites WHERE domain = 'gamesdesign.co.uk')
  AND page_type = 'guide'
  AND name LIKE 'guide-%'
  AND url LIKE '/blog/guide-%';

-- After: confirm the 5 rows now carry /guides/<slug>/index.html.
SELECT name, url, page_type, build_status
FROM pages
WHERE site_id = (SELECT id FROM sites WHERE domain = 'gamesdesign.co.uk')
  AND page_type = 'guide'
ORDER BY name;

COMMIT;

-- ----------------------------------------------------------------------------
-- ROLLBACK (if needed): restore url + build state from the snapshot.
--   UPDATE pages p
--   SET url = b.url, build_status = b.build_status,
--       built_from_plan_version = b.built_from_plan_version
--   FROM pages_bak_guides_url b WHERE p.id = b.id;
-- Drop the snapshot once satisfied: DROP TABLE pages_bak_guides_url;
-- ----------------------------------------------------------------------------

-- ============================================================================
-- FOLLOW-UPS (run after the diagnostics in the handoff tell you the scope):
--
-- 1) Rebuild pages that SHOW guide-list (homepage; guides hub once it has the
--    section) so list links point at the new URLs. Snapshot then flip:
--      CREATE TABLE IF NOT EXISTS pages_bak_guidelist_rebuild AS
--      SELECT id, name, url, build_status, built_from_plan_version, NOW() AS snapshot_at
--      FROM pages
--      WHERE site_id = (SELECT id FROM sites WHERE domain = 'gamesdesign.co.uk')
--        AND sections @> '["guide-list"]'::jsonb;
--      UPDATE pages SET build_status='needs_rebuild', built_from_plan_version=NULL
--      WHERE site_id = (SELECT id FROM sites WHERE domain = 'gamesdesign.co.uk')
--        AND sections @> '["guide-list"]'::jsonb;
--
-- 2) Rebuild pages whose rendered_html hardcodes /blog/guide-* links
--    (diagnostic c) so their inline links + link_registry re-sync.
--
-- 3) GIT CLEANUP (outside the DB): delete the orphaned old files
--    /blog/guide-economy-basics.html, /blog/guide-fairness-in-rng.html,
--    /blog/guide-p2p-architecture.html, /blog/guide-rng-design.html,
--    /blog/guide-skinner-box.html  (optionally add redirects).
-- ============================================================================
