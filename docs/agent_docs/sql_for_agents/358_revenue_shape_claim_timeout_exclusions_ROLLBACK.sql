-- 358_revenue_shape_claim_timeout_exclusions_ROLLBACK.sql
--
-- Inverse of 358: removes 'revenue_shape_cta' and 'missing_conversion_path'
-- from the claimed-item-timeout sweep's exclusion list.
--
-- ⚠ ROLLING BACK 358 ALONE REOPENS THE BYPASS WINDOW: the verifiers for these
-- two item types are registered in the live image (v1.0.1276+, init() in
-- check_revenue_shape.go), and an excluded-list rollback lets the timeout sweep
-- auto-complete stuck items PAST those verifiers — the exact two-day hole 305's
-- header records for the 151 lane. Only roll this back together with 361's
-- rollback (disable the checks) or after an image roll that unregisters the
-- verifiers. The lockstep test will fail the build at the next test run either
-- way — that is the designed alarm, not a nuisance.
--
-- Written 2026-08-09 (round-2 council objection: needle-gate discipline wants a
-- separate rollback file, not just a fail-closed forward file).

\set ON_ERROR_STOP on

BEGIN;

DO $pre$
DECLARE v_rows int;
BEGIN
    SELECT count(*) INTO v_rows FROM scheduled_tasks
    WHERE name = 'claimed-item-timeout'
      AND pre_query LIKE '%''dead_fragment_link'', ''literal_markdown'', ''unbuilt_internal_link'', ''revenue_shape_cta'', ''missing_conversion_path''%';
    IF v_rows <> 1 THEN
        RAISE EXCEPTION '358_ROLLBACK: pre-state mismatch (% rows) — the live list is not the post-358 string; re-read and recompose', v_rows;
    END IF;
END $pre$;

UPDATE scheduled_tasks
SET pre_query = replace(
    pre_query,
    '''dead_fragment_link'', ''literal_markdown'', ''unbuilt_internal_link'', ''revenue_shape_cta'', ''missing_conversion_path''',
    '''dead_fragment_link'', ''literal_markdown'', ''unbuilt_internal_link'''
)
WHERE name = 'claimed-item-timeout';

DO $post$
DECLARE v_rows int;
BEGIN
    SELECT count(*) INTO v_rows FROM scheduled_tasks
    WHERE name = 'claimed-item-timeout'
      AND pre_query LIKE '%revenue_shape_cta%';
    IF v_rows <> 0 THEN
        RAISE EXCEPTION '358_ROLLBACK: post-state mismatch (% rows still carry the entries)', v_rows;
    END IF;
END $post$;

COMMIT;
