-- 570 ROLLBACK — drop the (error_code, occurred_at DESC) index on agent_error_log.
--
-- Safe and cheap: dropping an index cannot lose data. What it DOES cost is measured and worth
-- knowing before you run it — the strike ladder in page_build_failure_guard.go:131 goes back to
-- reading ~8,000 buffers per refused deploy stamp instead of 3, and cmd/content-loss-check's
-- family re-grade goes back to a sequential scan of the whole table.
--
-- Plain DROP INDEX (not CONCURRENTLY) to match the forward migration's transaction.

BEGIN;

DO $do$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_indexes
                  WHERE tablename='agent_error_log' AND indexname='idx_error_log_code_time') THEN
    RAISE EXCEPTION '570 ROLLBACK REFUSED: idx_error_log_code_time does not exist — 570 was never applied, or something else already dropped it';
  END IF;
END
$do$;

DROP INDEX idx_error_log_code_time;

DO $do$
BEGIN
  IF EXISTS (SELECT 1 FROM pg_indexes
              WHERE tablename='agent_error_log' AND indexname='idx_error_log_code_time') THEN
    RAISE EXCEPTION '570 ROLLBACK: the index is still present';
  END IF;
  -- negative control: dropping ours must not have taken any of the other four
  IF NOT (EXISTS (SELECT 1 FROM pg_indexes WHERE tablename='agent_error_log' AND indexname='agent_error_log_pkey')
      AND EXISTS (SELECT 1 FROM pg_indexes WHERE tablename='agent_error_log' AND indexname='idx_error_log_time')
      AND EXISTS (SELECT 1 FROM pg_indexes WHERE tablename='agent_error_log' AND indexname='idx_error_log_unresolved')) THEN
    RAISE EXCEPTION '570 ROLLBACK: a pre-existing index was lost';
  END IF;
  RAISE NOTICE '570 ROLLBACK: applied — error_code is unindexed again.';
END
$do$;

COMMIT;
