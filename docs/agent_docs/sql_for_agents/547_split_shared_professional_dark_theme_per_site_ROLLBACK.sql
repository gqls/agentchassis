-- 547_split_shared_professional_dark_theme_per_site_ROLLBACK.sql
--
-- Reverses 547: repoints finetuning.uk and gaswholesalers.com back at the shared
-- seed collection `3196d966` and deletes the two per-site collections and themes.
--
-- ⚠ WHAT ROLLING BACK COSTS. Both sites go back to sharing ONE 1,649-byte theme row
-- that matches neither of their stylesheets (13,988 and 20,271 bytes). Migration
-- 542's `site_count <= 1` gate will refuse to patch either of them again, which is
-- SAFE but means no contrast fix can ever be applied to either site until they are
-- split again. Nothing is destroyed: the seed row is untouched by 547, and the
-- per-site rows deleted here are reconstructible from each site's served file.
--
-- ⚠ Deletion order matters: `style_collections.css_theme_id` is a foreign key onto
-- `css_themes`, and `sites.style_collection_id` onto `style_collections`. Repoint
-- the sites FIRST, then drop the collections, then the themes.

BEGIN;

UPDATE sites
   SET style_collection_id = '3196d966-24ef-4415-9dc8-1afbc02166ca',
       updated_at = NOW()
 WHERE domain IN ('finetuning.uk','gaswholesalers.com');

DELETE FROM style_collections
 WHERE name IN ('collection-finetuning-uk','collection-gaswholesalers-com');

DELETE FROM css_themes
 WHERE name IN ('theme-finetuning-uk','theme-gaswholesalers-com');

DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM css_themes WHERE name IN ('theme-finetuning-uk','theme-gaswholesalers-com')) THEN
        RAISE EXCEPTION '198/547 ROLLBACK: a per-site theme survives';
    END IF;
    IF EXISTS (SELECT 1 FROM style_collections WHERE name IN ('collection-finetuning-uk','collection-gaswholesalers-com')) THEN
        RAISE EXCEPTION '198/547 ROLLBACK: a per-site collection survives';
    END IF;
    IF (SELECT count(*) FROM sites WHERE style_collection_id='3196d966-24ef-4415-9dc8-1afbc02166ca') <> 2 THEN
        RAISE EXCEPTION '198/547 ROLLBACK: the two sites are not both back on the shared collection';
    END IF;
    IF (SELECT octet_length(css_content) FROM css_themes WHERE id='fecb962d-3ace-4c19-b08f-088eba46ea53') <> 1649 THEN
        RAISE EXCEPTION '198/547 ROLLBACK: the seed theme is not its original 1649 bytes';
    END IF;

    RAISE NOTICE '198/547 ROLLBACK: verified — both sites share the seed again (and are unpatchable again)';
END $$;

COMMIT;
