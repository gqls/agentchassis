-- 757_form_submissions_inbox_ISLAND_ROLLBACK.sql — undo 757, on the ISLAND.
--
-- TARGET: the ISLAND Postgres, NOT clients_db. Same invocation as 757:
--   cd /opt/island && docker compose exec -T postgres psql -U tools_api -d tools_api -v ON_ERROR_STOP=1 < this_file.sql
--
-- ⚠ IT REFUSES WHILE ANY ROW IS STILL 'pending'. A pending row is a submission that a visitor
-- believes they sent and that the cluster has not yet collected — the one state in this design
-- where the island holds the only copy. Dropping the table then is losing someone's enquiry.
-- Pulled rows do not block: the cluster already has them in form_submissions (756), so the island
-- copy is a receipt, not the record.
--
-- If the pending rows are genuinely junk (a spam burst you want gone), delete them explicitly and
-- re-run. The extra step is deliberate — it is the difference between discarding data on purpose
-- and discarding it as a side effect of tidying up.

BEGIN;

DO $$
DECLARE
  v_pending bigint := 0;
  v_pulled  bigint := 0;
BEGIN
  IF to_regclass('public.form_submissions_inbox') IS NULL THEN
    RAISE NOTICE '757 rollback: form_submissions_inbox does not exist — nothing to do.';
    RETURN;
  END IF;

  EXECUTE 'SELECT count(*) FROM form_submissions_inbox WHERE status = ''pending''' INTO v_pending;
  EXECUTE 'SELECT count(*) FROM form_submissions_inbox WHERE status = ''pulled'''  INTO v_pulled;

  IF v_pending > 0 THEN
    RAISE EXCEPTION
      'REFUSING: % row(s) still pending — the cluster has not collected them and this is their only copy. Collect them, or delete them deliberately, then re-run.', v_pending;
  END IF;

  RAISE NOTICE '757 rollback: proceeding — 0 pending, % already-pulled receipt row(s) will go.', v_pulled;
END $$;

DROP TABLE IF EXISTS form_submissions_inbox;

DO $$
BEGIN
  IF to_regclass('public.form_submissions_inbox') IS NOT NULL THEN
    RAISE EXCEPTION 'verify: the table survived the drop';
  END IF;
  RAISE NOTICE '757 rollback: OK — inbox dropped.';
END $$;

COMMIT;
