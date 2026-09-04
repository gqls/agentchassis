-- ROLLBACK for 778 (bugs_open/477).
--
-- ⚠ THIS IS DESTRUCTIVE IN A WAY THE OTHER ROLLBACKS IN THIS LANE ARE NOT.
-- Dropping this table destroys the only durable record of who each site was
-- delivered to. For the backfilled row it destroys the LAST machine-readable
-- copy: `orchestration_states` is a ~26-hour rolling window and the idea.uk
-- delivery run aged out on 2026-09-04. After that, dropping this table leaves the
-- address in prose only (bugs_open/477 and 778's header).
--
-- So the guard refuses whenever the table holds anything, rather than warning.
-- If you genuinely intend to discard it, empty the table deliberately first —
-- which forces the decision to be made about the ROWS, where it belongs, rather
-- than as a side effect of removing a schema object.

BEGIN;

DO $$
DECLARE n int;
BEGIN
  IF to_regclass('public.site_deliveries') IS NULL THEN
    RAISE NOTICE '778 ROLLBACK: site_deliveries does not exist; nothing to do.';
    RETURN;
  END IF;
  SELECT count(*) INTO n FROM site_deliveries;
  IF n > 0 THEN
    RAISE EXCEPTION '778 ROLLBACK REFUSED: site_deliveries holds % row(s). Dropping it destroys the only durable record of who those sites were delivered to, and for any backfilled row the source (orchestration_states) has already aged out. Empty the table deliberately first if that is really the intent.', n;
  END IF;
END $$;

DROP TABLE IF EXISTS site_deliveries;

COMMIT;
