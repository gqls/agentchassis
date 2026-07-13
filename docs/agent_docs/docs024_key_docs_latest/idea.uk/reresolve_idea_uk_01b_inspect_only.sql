-- =====================================================================
-- idea.uk RE-RESOLVE — STEP 1b: INSPECT + UNIQUENESS (backups already exist)
-- site_id = 1244516d-014d-421c-88c6-090bb1e9552a
--
-- Run this INSTEAD of re-running step 1 (the bak_* tables from step 1 already
-- exist; re-running step 1 would fail with "relation already exists"). Run it
-- AFTER applying migration_layouts_scheme_and_light_tool_portal.sql — it reads
-- layouts.scheme and checks for tool-portal-light, both added by that migration.
--
-- EYEBALL the four uniqueness counts: they MUST all be 0 before step 2.
-- A count > 0 means a row is shared with another site — STOP, do not delete it.
-- =====================================================================
\set ON_ERROR_STOP on

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

\echo '--- palette + typography provenance (step 2 deletes only when source_site_id = idea.uk) ---'
SELECT 'palette' AS kind, id, name, origin, source_site_id FROM palettes WHERE id IN (SELECT id FROM bak_palettes_idea_20260621)
UNION ALL
SELECT 'typography' AS kind, id, name, origin, source_site_id FROM typography_sets WHERE id IN (SELECT id FROM bak_typography_sets_idea_20260621);

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

\echo '--- confirm the migration landed: tool-portal-light exists + schemes set ---'
SELECT name, scheme, is_active FROM layouts
WHERE name IN ('tool-portal-light','tool-portal-dark','soft-editorial')
ORDER BY name;
