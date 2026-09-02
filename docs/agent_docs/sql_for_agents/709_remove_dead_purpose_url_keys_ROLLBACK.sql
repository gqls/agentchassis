-- 709_remove_dead_purpose_url_keys_ROLLBACK.sql — restores ONLY the four keys
-- 709 deleted, from the per-row backup, onto the CURRENT content_data (never a
-- wholesale overwrite: the live row may legitimately have moved since the
-- apply, and clobbering unrelated keys with a stale snapshot would be a second
-- incident). Sidecar, hand-run only (SIDECAR_RE excludes it from --apply).
--
-- Idempotent: a key already present on the live row is left alone (the
-- jsonb_set only fires where the live row lacks the key), so a partial
-- rollback can be re-run safely.

BEGIN;

DO $$
DECLARE n int;
BEGIN
    SELECT count(*) INTO n FROM bak_site_dead_purpose_urls_20260902;
    IF n = 0 THEN
        RAISE EXCEPTION '709 rollback: backup table bak_site_dead_purpose_urls_20260902 is empty — nothing to restore from; if 709 never applied there is nothing to roll back';
    END IF;
END $$;

UPDATE sites s
   SET content_data = s.content_data ||
       (SELECT COALESCE(jsonb_object_agg(k, b.content_data->k), '{}'::jsonb)
          FROM bak_site_dead_purpose_urls_20260902 b,
               unnest(ARRAY['icon_url','content_hero_url','illustration_url','sprite_sheet_url']) AS k
         WHERE b.id = s.id
           AND b.content_data ? k
           AND NOT (s.content_data ? k))
  FROM bak_site_dead_purpose_urls_20260902 bb
 WHERE bb.id = s.id;

-- Verify: every backed-up key is present again on its live row.
DO $$
DECLARE missing int;
BEGIN
    SELECT count(*) INTO missing
      FROM bak_site_dead_purpose_urls_20260902 b
      JOIN sites s ON s.id = b.id
      CROSS JOIN unnest(ARRAY['icon_url','content_hero_url','illustration_url','sprite_sheet_url']) AS k
     WHERE b.content_data ? k AND NOT (s.content_data ? k);
    IF missing <> 0 THEN
        RAISE EXCEPTION '709 rollback verify: % backed-up key(s) still absent from live rows', missing;
    END IF;
    RAISE NOTICE '709 rollback OK: every backed-up key is present on its live row again';
END $$;

COMMIT;
