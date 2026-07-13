-- =====================================================================
-- idea.uk composition RE-RESOLVE — STEP 2 of 4: DETACH + CLEAR (mutating)
-- site_id = 1244516d-014d-421c-88c6-090bb1e9552a
--
-- Run ONLY after step 1's uniqueness checks were ALL 0. This detaches the
-- installed composition and deletes idea.uk's own composition rows so the
-- re-run recreates them cleanly (frees the unique names theme-/palette-/
-- typography-idea-uk) and supersedes the stale resolved_composition spec so
-- install_site_composition will write a fresh one.
--
-- Uses the bak_* tables from step 1 as the id source, so it does not depend
-- on the (about-to-be-detached) live chain. Palette/typography deletes are
-- guarded by source_site_id = idea.uk, so a shared/seed row is never removed.
--
-- FK-safe order: NULL sites pointer -> style_collections -> css_themes ->
-- palettes -> typography_sets.
-- =====================================================================
\set ON_ERROR_STOP on
BEGIN;

-- 1) detach so install can re-run and so the style_collection is unreferenced
UPDATE sites SET style_collection_id = NULL WHERE id = '1244516d-014d-421c-88c6-090bb1e9552a';

-- 2) delete idea.uk's composition rows (children first)
DELETE FROM style_collections WHERE id IN (SELECT id FROM bak_style_collections_idea_20260621);
DELETE FROM css_themes        WHERE id IN (SELECT id FROM bak_css_themes_idea_20260621);
DELETE FROM palettes          WHERE id IN (SELECT id FROM bak_palettes_idea_20260621)
                                AND source_site_id = '1244516d-014d-421c-88c6-090bb1e9552a';
DELETE FROM typography_sets   WHERE id IN (SELECT id FROM bak_typography_sets_idea_20260621)
                                AND source_site_id = '1244516d-014d-421c-88c6-090bb1e9552a';

-- 3) supersede the stale resolved_composition spec (install refuses to
--    overwrite an installed composition; the spec is the decision record)
UPDATE site_specs
   SET is_current = false, superseded_at = now()
 WHERE site_id = '1244516d-014d-421c-88c6-090bb1e9552a'
   AND aspect = 'resolved_composition'
   AND is_current;

-- 4) show the result before committing
\echo '--- after detach/clear: sites pointer should be NULL, no idea.uk theme/sc rows ---'
SELECT id, style_collection_id FROM sites WHERE id = '1244516d-014d-421c-88c6-090bb1e9552a';
SELECT count(*) AS idea_style_collections FROM style_collections WHERE id IN (SELECT id FROM bak_style_collections_idea_20260621);
SELECT count(*) AS idea_css_themes        FROM css_themes        WHERE id IN (SELECT id FROM bak_css_themes_idea_20260621);
SELECT aspect, is_current FROM site_specs WHERE site_id = '1244516d-014d-421c-88c6-090bb1e9552a' AND aspect = 'resolved_composition' ORDER BY created_at DESC LIMIT 3;

COMMIT;
\echo '--- COMMITTED. Now run step 3 (re-trigger), then step 4 (verify). ---'
