-- 374_decision_regression_claim_timeout_exclusion_ROLLBACK.sql
--
-- Inverse of 374: removes 'decision_regression' from the claimed-item-timeout
-- sweep's item_type exclusion list. Same targeted-replace shape, with the
-- assertions inverted.
--
-- Only correct while the Go verifier is NOT live. Rolling this back with
-- VerifyDecisionRegressionResolved in the running image re-opens the bypass the
-- migration exists to close — the sweep would auto-complete decision_regression
-- items on handler-orchestration evidence alone, which for a needs_human_review
-- item means on a person's unverified word.

\set ON_ERROR_STOP on

BEGIN;

DO $pre$
DECLARE v_rows int;
BEGIN
    SELECT count(*) INTO v_rows FROM scheduled_tasks
     WHERE pre_query LIKE '%''decision_regression''%';
    IF v_rows <> 1 THEN
        RAISE EXCEPTION 'expected exactly 1 scheduled_task excluding decision_regression, found % — 374 may not be applied', v_rows;
    END IF;
END $pre$;

UPDATE scheduled_tasks
   SET pre_query = replace(
           pre_query,
           ', ''missing_conversion_path'', ''decision_regression'')',
           ', ''missing_conversion_path'')'
       )
 WHERE pre_query LIKE '%''decision_regression''%';

DO $post$
DECLARE v_rows int;
BEGIN
    SELECT count(*) INTO v_rows FROM scheduled_tasks
     WHERE pre_query LIKE '%''decision_regression''%';
    IF v_rows <> 0 THEN
        RAISE EXCEPTION '374 ROLLBACK FAILED: decision_regression still excluded in % row(s)', v_rows;
    END IF;
END $post$;

COMMIT;
