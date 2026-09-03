-- 756_form_submissions_and_site_form_routes_ROLLBACK.sql — undo 756.
--
-- Drops site_form_routes and form_submissions. Safe only while they are still inert: nothing in
-- the Go tree reads or writes either table until the phase 2 receiver ships.
--
-- ⚠ IT REFUSES WHILE form_submissions HOLDS ANY ROW, and that refusal is the point. A submission
-- is a lead — possibly one someone paid for, and the only copy, since the whole design stores
-- before it notifies. Dropping real leads is not a rollback, it is data loss with a tidy name.
-- If you genuinely mean to discard them, export them first and then delete them explicitly; the
-- extra step is deliberate.
--
-- A route row, by contrast, is regenerable config, so its presence does not block. Any token in
-- it dies with the table: re-running 756 mints new ones, and any form markup already stamped with
-- an old token will be refused by the receiver until the site is re-rendered. That is the correct
-- failure — a stale token should stop working, not quietly match a new row.

BEGIN;

DO $$
DECLARE
  v_subs   bigint := 0;
  v_routes bigint := 0;
BEGIN
  IF to_regclass('public.form_submissions') IS NULL
     AND to_regclass('public.site_form_routes') IS NULL THEN
    RAISE NOTICE '756 rollback: neither table exists — nothing to do.';
    RETURN;
  END IF;

  IF to_regclass('public.form_submissions') IS NOT NULL THEN
    EXECUTE 'SELECT count(*) FROM form_submissions' INTO v_subs;
  END IF;
  IF to_regclass('public.site_form_routes') IS NOT NULL THEN
    EXECUTE 'SELECT count(*) FROM site_form_routes' INTO v_routes;
  END IF;

  IF v_subs > 0 THEN
    RAISE EXCEPTION
      'REFUSING: form_submissions holds % row(s). These are real enquiries and this migration is their only store. Export them, delete them deliberately, then re-run this rollback.', v_subs;
  END IF;

  RAISE NOTICE '756 rollback: proceeding — 0 submissions, % route row(s) will be dropped.', v_routes;
END $$;

-- Order matters: form_submissions references site_form_routes.
DROP TABLE IF EXISTS form_submissions;
DROP TABLE IF EXISTS site_form_routes;

DO $$
BEGIN
  IF to_regclass('public.form_submissions') IS NOT NULL
     OR to_regclass('public.site_form_routes') IS NOT NULL THEN
    RAISE EXCEPTION 'verify: a table survived the drop';
  END IF;
  RAISE NOTICE '756 rollback: OK — both tables dropped.';
END $$;

COMMIT;
