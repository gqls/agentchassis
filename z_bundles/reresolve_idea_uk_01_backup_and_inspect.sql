-- =====================================================================
-- idea.uk composition RE-RESOLVE — STEP 1 of 4: BACKUP + INSPECT (safe)
-- site_id = 1244516d-014d-421c-88c6-090bb1e9552a
--
-- Run this FIRST (read-only except it creates bak_* tables). It backs up
-- every row the re-resolve will touch, then prints the current composition
-- chain and the uniqueness checks. EYEBALL the uniqueness counts: they MUST
-- all be 0 before you run step 2. If any is > 0, a row is shared with another
-- site — STOP and do not delete it.
--
-- Re-running: CREATE TABLE ... AS will error if a bak_* table already exists
-- (deliberate — prevents clobbering a prior backup). Drop or rename the old
-- bak_* tables if you genuinely need to re-run.
-- =====================================================================
\set ON_ERROR_STOP on

-- ---- backups (the whole composition chain for idea.uk) ----
CREATE TABLE bak_sites_idea_20260621 AS
  SELECT * FROM sites WHERE id = '1244516d-014d-421c-88c6-090bb1e9552a';

CREATE TABLE bak_site_specs_idea_20260621 AS
  SELECT * FROM site_specs WHERE site_id = '1244516d-014d-421c-88c6-090bb1e9552a';

CREATE TABLE bak_style_collections_idea_20260621 AS
  SELECT * FROM style_collections
  WHERE id IN (SELECT style_collection_id FROM sites WHERE id = '1244516d-014d-421c-88c6-090bb1e9552a');

CREATE TABLE bak_css_themes_idea_20260621 AS
  SELECT * FROM css_themes
  WHERE id IN (
    SELECT css_theme_id FROM style_collections
    WHERE id IN (SELECT style_collection_id FROM sites WHERE id = '1244516d-014d-421c-88c6-090bb1e9552a')
  );

CREATE TABLE bak_palettes_idea_20260621 AS
  SELECT * FROM palettes
  WHERE source_site_id = '1244516d-014d-421c-88c6-090bb1e9552a'
     OR id IN (
       SELECT palette_id FROM css_themes
       WHERE id IN (SELECT css_theme_id FROM style_collections
                    WHERE id IN (SELECT style_collection_id FROM sites WHERE id = '1244516d-014d-421c-88c6-090bb1e9552a'))
     );

CREATE TABLE bak_typography_sets_idea_20260621 AS
  SELECT * FROM typography_sets
  WHERE id IN (
    SELECT typography_set_id FROM css_themes
    WHERE id IN (SELECT css_theme_id FROM style_collections
                 WHERE id IN (SELECT style_collection_id FROM sites WHERE id = '1244516d-014d-421c-88c6-090bb1e9552a'))
  );

\echo '--- backups created: bak_sites/site_specs/style_collections/css_themes/palettes/typography_sets _idea_20260621 ---'

-- ---- current composition chain ----
\echo '--- current chain (site -> style_collection -> css_theme -> palette/typography/layout) ---'
SELECT s.id AS site_id, s.style_collection_id,
       sc.css_theme_id,
       t.palette_id, t.typography_set_id, t.layout_id,
       l.name AS layout_name, l.scheme AS layout_scheme
FROM sites s
LEFT JOIN style_collections sc ON sc.id = s.style_collection_id
LEFT JOIN css_themes t         ON t.id = sc.css_theme_id
LEFT JOIN layouts l            ON l.id = t.layout_id
WHERE s.id = '1244516d-014d-421c-88c6-090bb1e9552a';

-- ---- the palette/typography rows' provenance (source_site_id guards the deletes) ----
\echo '--- palette + typography provenance (delete in step 2 only fires when source_site_id = idea.uk) ---'
SELECT 'palette' AS kind, id, name, origin, source_site_id FROM palettes WHERE id IN (SELECT id FROM bak_palettes_idea_20260621)
UNION ALL
SELECT 'typography' AS kind, id, name, origin, source_site_id FROM typography_sets WHERE id IN (SELECT id FROM bak_typography_sets_idea_20260621);

-- ---- uniqueness checks — ALL must be 0 before step 2 ----
\echo '--- uniqueness checks (every count MUST be 0; >0 = shared row, do NOT delete) ---'
SELECT 'other_sites_using_this_style_collection' AS check, count(*) AS n
FROM sites
WHERE style_collection_id IN (SELECT id FROM bak_style_collections_idea_20260621)
  AND id <> '1244516d-014d-421c-88c6-090bb1e9552a'
UNION ALL
SELECT 'other_style_collections_using_this_css_theme', count(*)
FROM style_collections
WHERE css_theme_id IN (SELECT id FROM bak_css_themes_idea_20260621)
  AND id NOT IN (SELECT id FROM bak_style_collections_idea_20260621)
UNION ALL
SELECT 'other_css_themes_using_this_palette', count(*)
FROM css_themes
WHERE palette_id IN (SELECT id FROM bak_palettes_idea_20260621)
  AND id NOT IN (SELECT id FROM bak_css_themes_idea_20260621)
UNION ALL
SELECT 'other_css_themes_using_this_typography_set', count(*)
FROM css_themes
WHERE typography_set_id IN (SELECT id FROM bak_typography_sets_idea_20260621)
  AND id NOT IN (SELECT id FROM bak_css_themes_idea_20260621);

-- ---- confirm the migration landed (the matcher needs these) ----
\echo '--- confirm tool-portal-light exists + schemes set (run AFTER the layouts migration) ---'
SELECT name, scheme, is_active FROM layouts
WHERE name IN ('tool-portal-light','tool-portal-dark','soft-editorial')
ORDER BY name;
