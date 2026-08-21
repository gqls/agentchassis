-- 531 ROLLBACK — restores the council-gate diagnosis_guardian seat to its
-- pre-531 text: the inverted error_step discipline AND the mangled heading that
-- 099_SYNC_gate_roster.py's unanchored replace produced.
--
-- ⚠ Both things it restores are defects. Reverse only to undo a botched apply.
-- Pair with 530_..._ROLLBACK.sql; running one without the other leaves the two
-- rosters contradicting each other.
--
-- 377: every edit is downstream of <!--CACHE_BREAKPOINT-->, so the marker's
-- offset must be unchanged in this direction too. Asserted, same as forward.

\set ON_ERROR_STOP on

BEGIN;

SELECT snapshot_agent('council-gate',
  '531_..._ROLLBACK.sql: pre-rollback');

DO $rb$
DECLARE
    v_marker     text := '<!--CACHE_BREAKPOINT-->';
    v_old_bullet text := '- CONFIG-LEVEL error_step: the workflow coordinator reads ONLY step.config.error_step -- a step-level error_step is parsed but silently inert (a real, recurring trap). Any plan adding error routing must place it inside config.';
    v_new_bullet text := '- error_step, BOTH LOCATIONS ARE LIVE: routeToErrorStepOrFail (coordinator.go) checks the STEP-LEVEL error_step FIRST -- the code comment there calls it the preferred location -- and falls back to step.config.error_step for backward compatibility. Neither is inert. Do NOT object to error routing on placement. Object where a plan REMOVES routing that a failure path needs, or points it at a step that swallows the error.';
    v_old_judge  text := '(d) does it place error_step outside config (silently inert), or move loop work/tokens onto shared pods.';
    v_new_judge  text := '(d) does it remove a step''s error routing, or point it at a step that swallows the failure (placement is NOT the test -- step-level and config-level are both honoured), or move loop work/tokens onto shared pods.';
    v_old_head   text := '## The author''s stated rationale loop''s load-bearing disciplines';
    v_new_head   text := '## The diagnosis loop''s load-bearing disciplines';
    v_live text;
    v_back text;
    v_mark_old int;
    v_rows int;
BEGIN
    SELECT default_config->'workflow'->'steps'->'review_diagnosis_guardian'->'config'->>'prompt_template'
      INTO v_live
      FROM agent_definitions
     WHERE type = 'council-gate' AND is_active
       AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

    IF v_live IS NULL THEN
        RAISE EXCEPTION '531 ROLLBACK: no live council-gate row';
    END IF;

    v_mark_old := position(v_marker in v_live);
    IF v_mark_old = 0 THEN
        RAISE EXCEPTION '531 ROLLBACK: 377 cache breakpoint absent — stop and investigate';
    END IF;

    IF (length(v_live) - length(replace(v_live, v_new_bullet, ''))) / length(v_new_bullet) <> 1
       OR (length(v_live) - length(replace(v_live, v_new_judge, ''))) / length(v_new_judge) <> 1
       OR (length(v_live) - length(replace(v_live, v_new_head, ''))) / length(v_new_head) <> 1 THEN
        RAISE EXCEPTION '531 ROLLBACK: 531 text not present exactly once — nothing to reverse, or the prompt moved on';
    END IF;

    v_back := replace(replace(replace(v_live, v_new_bullet, v_old_bullet), v_new_judge, v_old_judge), v_new_head, v_old_head);

    IF position(v_marker in v_back) <> v_mark_old THEN
        RAISE EXCEPTION '531 ROLLBACK: the 377 cache breakpoint MOVED — refusing';
    END IF;

    UPDATE agent_definitions
       SET default_config = jsonb_set(default_config,
             ARRAY['workflow','steps','review_diagnosis_guardian','config','prompt_template'],
             to_jsonb(v_back), false),
           updated_at = now()
     WHERE type = 'council-gate' AND is_active
       AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

    GET DIAGNOSTICS v_rows = ROW_COUNT;
    IF v_rows <> 1 THEN
        RAISE EXCEPTION '531 ROLLBACK: expected 1 row, updated %', v_rows;
    END IF;
END $rb$;

DO $verify$
DECLARE v_live text;
BEGIN
    SELECT default_config->'workflow'->'steps'->'review_diagnosis_guardian'->'config'->>'prompt_template'
      INTO v_live
      FROM agent_definitions
     WHERE type = 'council-gate' AND is_active
       AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;
    IF position('reads ONLY step.config.error_step' in v_live) = 0
       OR position('## The author''s stated rationale loop' in v_live) = 0 THEN
        RAISE EXCEPTION '531 ROLLBACK VERIFY: the pre-531 text is not fully back';
    END IF;
    IF position('checks the STEP-LEVEL error_step FIRST' in v_live) > 0 THEN
        RAISE EXCEPTION '531 ROLLBACK VERIFY: 531 text still present';
    END IF;
    RAISE NOTICE '531 ROLLBACK OK: council-gate.review_diagnosis_guardian restored to its pre-531 text.';
END $verify$;

COMMIT;
