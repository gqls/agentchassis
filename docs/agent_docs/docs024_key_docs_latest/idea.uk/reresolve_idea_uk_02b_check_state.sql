-- =====================================================================
-- idea.uk RE-RESOLVE — STEP 2b: STATE CHECK (read-only)
-- site_id = 1244516d-014d-421c-88c6-090bb1e9552a
--
-- Run this AFTER 'ROLLBACK;' (to release the open transaction's lock), as a
-- FILE (kubectl exec -i ... < this), never pasted. It reports the COMMITTED
-- state so you can tell whether step 2's effect is already in place.
--
-- INTERPRETATION:
--   * detached = NULL, all four composition counts = 0, current
--     resolved_composition = 0  -> STEP 2 IS ALREADY DONE -> go to step 3.
--   * style_collection_id = 54c5a076..., counts = 1, resolved_composition = 1
--     -> STEP 2 HAS NOT TAKEN -> re-run reresolve_idea_uk_02_detach_and_clear.sql
--        as a FILE, then re-run this check.
-- =====================================================================

\echo '--- 1. is idea.uk detached? (style_collection_id NULL if step 2 took effect) ---'
SELECT id, style_collection_id FROM sites WHERE id = '1244516d-014d-421c-88c6-090bb1e9552a';

\echo '--- 2. do idea.uk composition rows still exist? (0 each = deleted) ---'
SELECT
 (SELECT count(*) FROM style_collections WHERE id = '54c5a076-e389-407b-a6ff-3e48c3c8d2d3') AS style_collection,
 (SELECT count(*) FROM css_themes        WHERE id = 'cbecf8a4-0206-461c-a2da-805301b5662d') AS css_theme,
 (SELECT count(*) FROM palettes          WHERE name = 'palette-idea-uk')                    AS palette_idea_uk,
 (SELECT count(*) FROM typography_sets   WHERE name = 'typography-idea-uk')                  AS typography_idea_uk;

\echo '--- 3. is there still a CURRENT resolved_composition spec? (0 = superseded) ---'
SELECT count(*) AS current_resolved_composition
FROM site_specs
WHERE site_id = '1244516d-014d-421c-88c6-090bb1e9552a' AND aspect = 'resolved_composition' AND is_current;

\echo '--- 4. backups sanity (should be 1 each) ---'
SELECT
 (SELECT count(*) FROM bak_style_collections_idea_20260625) AS bak_sc,
 (SELECT count(*) FROM bak_css_themes_idea_20260625)        AS bak_theme,
 (SELECT count(*) FROM bak_palettes_idea_20260625)          AS bak_palette,
 (SELECT count(*) FROM bak_typography_sets_idea_20260625)   AS bak_typo;
