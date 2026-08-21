-- 548_seed_webdesign_uk_theme_row_from_deploy_repo_ROLLBACK.sql
--
-- Reverses 548: empties webdesign.uk's theme row again.
--
-- ⚠ WHAT ROLLING BACK COSTS, and it is not symmetrical with the other rollbacks in
-- this series. Emptying the row does NOT re-expose the site to a clobber — migration
-- 542's gate refuses to patch a base under 4096 bytes, so an empty row is now SAFE.
-- What it costs is PATCHABILITY: every contrast finding for webdesign.uk parks for a
-- human instead of being applied, forever, because nothing else fills that row until
-- a webdesign-agent design run does it via migration 543.
--
-- So the only honest reason to run this is that the seeded content turned out to be
-- WRONG (e.g. the blob was stale, or carried a pre-2026-08-14 ink derivation that a
-- fresh render would not produce). In that case empty it and re-seed from a correct
-- source rather than leaving it empty — an empty row is a safe holding state, not a
-- destination.
--
-- Guarded on the exact md5 548 wrote, so this cannot discard content that a
-- webdesign-agent render (543) has since persisted, or that another lane has fixed.

BEGIN;

UPDATE css_themes ct
   SET css_content = '',
       version = version + 1,
       updated_at = NOW(),
       description = NULL
  FROM style_collections sc, sites s
 WHERE sc.css_theme_id = ct.id
   AND s.style_collection_id = sc.id
   AND s.domain = 'webdesign.uk'
   AND md5(ct.css_content) = 'a582e515df3a31eeff30359c073205a9';

DO $$
DECLARE
    v_bytes int;
BEGIN
    SELECT octet_length(ct.css_content) INTO v_bytes
      FROM sites s
      JOIN style_collections sc ON sc.id = s.style_collection_id
      JOIN css_themes ct ON ct.id = sc.css_theme_id
     WHERE s.domain = 'webdesign.uk';

    IF v_bytes <> 0 THEN
        RAISE EXCEPTION '198/548 ROLLBACK: row is % bytes, not 0 — the md5 guard did not match, so nothing was reverted (this is the guard working: somebody else owns that content now)', v_bytes;
    END IF;

    RAISE NOTICE '198/548 ROLLBACK: verified — webdesign.uk row emptied; the site is SAFE but no longer patchable until it is re-seeded or a design run fills it';
END $$;

COMMIT;
