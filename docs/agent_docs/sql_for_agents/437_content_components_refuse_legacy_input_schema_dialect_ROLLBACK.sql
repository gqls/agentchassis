-- 437 ROLLBACK — drop chk_input_schema_no_legacy_dialect and restore the three rows
-- from the backup table migration 437 created. Hand-run sidecar (uppercase suffix:
-- the runner lists it and never applies it).
--
-- Order matters: the constraint must go BEFORE the restore, or the restore itself is
-- refused — which is exactly the guarantee this file is undoing.

BEGIN;

DO $$
BEGIN
  IF to_regclass('public.content_components_bak_20260816_265_legacy_dialect') IS NULL THEN
    RAISE EXCEPTION 'backup table content_components_bak_20260816_265_legacy_dialect is absent — 437 was not applied here, or the backup was dropped; nothing to restore';
  END IF;
END $$;

ALTER TABLE public.content_components
  DROP CONSTRAINT IF EXISTS chk_input_schema_no_legacy_dialect;

UPDATE public.content_components c
   SET input_schema = b.input_schema,
       updated_at   = b.updated_at
  FROM public.content_components_bak_20260816_265_legacy_dialect b
 WHERE c.id = b.id;

DO $$
DECLARE n int;
BEGIN
  SELECT count(*) INTO n FROM public.content_components WHERE input_schema ? 'properties';
  IF n <> 3 THEN
    RAISE EXCEPTION 'restore left % legacy row(s), expected 3 — aborting', n;
  END IF;
  RAISE NOTICE '437 ROLLBACK: constraint dropped, 3 rows restored';
END $$;

-- The backup table is left in place deliberately; drop it by hand once satisfied:
--   DROP TABLE public.content_components_bak_20260816_265_legacy_dialect;

COMMIT;
