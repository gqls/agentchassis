-- 489_page_component_history_retention_ROLLBACK.sql
--
-- Removes the retention mechanism: the scheduled task and its partial index.
-- ALREADY-NULLED PAYLOADS ARE NOT RECOVERABLE — this rollback stops future
-- pruning; it cannot restore bytes a past run removed. (That asymmetry is why
-- the forward migration's predicate takes only machine_made trigger-arm
-- payloads — the class whose artefact the platform's save-path snapshot has
-- always declined to keep — and why enablement precedes the first eligible row
-- by ~20 days.)
--
-- doc_notes report rows are left in place: they are the record that runs
-- happened, which outlives the mechanism.

BEGIN;

DELETE FROM scheduled_tasks WHERE name = 'page-component-history-retention';

DROP INDEX IF EXISTS idx_pch_retention_candidates;

DO $guard$
DECLARE
    n integer;
BEGIN
    SELECT count(*) INTO n FROM scheduled_tasks WHERE name = 'page-component-history-retention';
    IF n <> 0 THEN
        RAISE EXCEPTION 'mig489 rollback: task row still present';
    END IF;
END $guard$;

COMMIT;
