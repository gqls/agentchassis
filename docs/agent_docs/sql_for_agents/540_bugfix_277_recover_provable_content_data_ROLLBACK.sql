-- 540 ROLLBACK — restores the three components to content_data = NULL from the
-- backup table, which holds their exact pre-image.
--
-- This is a RESTORE, not a reconstruction: it copies back from
-- page_components_bak_20260822_277_recover rather than setting NULL blindly, so
-- it cannot clobber a value written by someone else after 540. If a row's
-- content_data is no longer the one 540 wrote, this refuses it.
--
-- ⚠ Reversing makes these three components UNFILLABLE again
-- (ContentDataCanFillTemplate → false), which is the state that made them
-- unrepairable in the first place. Reverse only to undo a botched apply.

\set ON_ERROR_STOP on

BEGIN;

DO $rb$
DECLARE n int;
BEGIN
  IF to_regclass('public.page_components_bak_20260822_277_recover') IS NULL THEN
    RAISE EXCEPTION '540 ROLLBACK: backup table is gone — refusing to guess the pre-image';
  END IF;

  UPDATE page_components pc
     SET content_data = b.content_data, updated_at = now()
    FROM page_components_bak_20260822_277_recover b
   WHERE b.id = pc.id
     AND md5(pc.rendered_html) = md5(b.rendered_html);

  GET DIAGNOSTICS n = ROW_COUNT;
  IF n <> 3 THEN
    RAISE EXCEPTION '540 ROLLBACK: restored % rows, expected 3 (rendered_html moved on one of them)', n;
  END IF;

  SELECT count(*) INTO n FROM page_components
   WHERE id IN ('e50a9dbc-569c-41c5-ac01-bc564dc9a53a',
                'bd1f5219-c230-4143-93d7-7ece0f4d8e9f',
                '2b9d24d7-9e04-401b-a0b5-0e16e7731895')
     AND content_data IS NULL;
  IF n <> 3 THEN
    RAISE EXCEPTION '540 ROLLBACK: % of 3 are back to NULL content_data, expected 3', n;
  END IF;

  RAISE NOTICE '540 ROLLBACK OK: three components restored to their pre-540 pre-image (content_data NULL).';
END $rb$;

COMMIT;
