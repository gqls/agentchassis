-- 530 ROLLBACK — restores the fix-proposer diagnosis_guardian seat's ORIGINAL
-- (and factually inverted) error_step discipline.
--
-- ⚠ Reversing this puts a claim back that is REFUTED at
-- platform/orchestration/coordinator.go:3666-3679. Reverse it only to undo a
-- botched apply, never because the old text "looked more familiar".
--
-- Pair with 531_..._ROLLBACK.sql: running only one leaves the two rosters
-- saying opposite things about the same coordinator.

\set ON_ERROR_STOP on

BEGIN;

SELECT snapshot_agent('fix-proposer',
  '530_..._ROLLBACK.sql: pre-rollback');

DO $rb$
DECLARE
    v_old_bullet text := '- CONFIG-LEVEL error_step: the workflow coordinator reads ONLY step.config.error_step -- a step-level error_step is parsed but silently inert (a real, recurring trap). Any plan adding error routing must place it inside config.';
    v_new_bullet text := '- error_step, BOTH LOCATIONS ARE LIVE: routeToErrorStepOrFail (coordinator.go) checks the STEP-LEVEL error_step FIRST -- the code comment there calls it the preferred location -- and falls back to step.config.error_step for backward compatibility. Neither is inert. Do NOT object to error routing on placement. Object where a plan REMOVES routing that a failure path needs, or points it at a step that swallows the error.';
    v_old_judge  text := '(d) does it place error_step outside config (silently inert), or move loop work/tokens onto shared pods.';
    v_new_judge  text := '(d) does it remove a step''s error routing, or point it at a step that swallows the failure (placement is NOT the test -- step-level and config-level are both honoured), or move loop work/tokens onto shared pods.';
    v_live text;
    v_back text;
    v_rows int;
BEGIN
    SELECT default_config->'workflow'->'steps'->'review_diagnosis_guardian'->'config'->>'prompt_template'
      INTO v_live
      FROM agent_definitions
     WHERE type = 'fix-proposer' AND is_active
       AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

    IF v_live IS NULL THEN
        RAISE EXCEPTION '530 ROLLBACK: no live fix-proposer row';
    END IF;
    IF (length(v_live) - length(replace(v_live, v_new_bullet, ''))) / length(v_new_bullet) <> 1
       OR (length(v_live) - length(replace(v_live, v_new_judge, ''))) / length(v_new_judge) <> 1 THEN
        RAISE EXCEPTION '530 ROLLBACK: 530 text not present exactly once — nothing to reverse, or the prompt moved on';
    END IF;

    v_back := replace(replace(v_live, v_new_bullet, v_old_bullet), v_new_judge, v_old_judge);

    UPDATE agent_definitions
       SET default_config = jsonb_set(default_config,
             ARRAY['workflow','steps','review_diagnosis_guardian','config','prompt_template'],
             to_jsonb(v_back), false),
           updated_at = now()
     WHERE type = 'fix-proposer' AND is_active
       AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

    GET DIAGNOSTICS v_rows = ROW_COUNT;
    IF v_rows <> 1 THEN
        RAISE EXCEPTION '530 ROLLBACK: expected 1 row, updated %', v_rows;
    END IF;
END $rb$;

DO $verify$
DECLARE v_live text;
BEGIN
    SELECT default_config->'workflow'->'steps'->'review_diagnosis_guardian'->'config'->>'prompt_template'
      INTO v_live
      FROM agent_definitions
     WHERE type = 'fix-proposer' AND is_active
       AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;
    IF position('reads ONLY step.config.error_step' in v_live) = 0 THEN
        RAISE EXCEPTION '530 ROLLBACK VERIFY: the original bullet is not back';
    END IF;
    IF position('checks the STEP-LEVEL error_step FIRST' in v_live) > 0 THEN
        RAISE EXCEPTION '530 ROLLBACK VERIFY: 530 text still present';
    END IF;
    RAISE NOTICE '530 ROLLBACK OK: fix-proposer.review_diagnosis_guardian restored to its pre-530 text.';
END $verify$;

COMMIT;
