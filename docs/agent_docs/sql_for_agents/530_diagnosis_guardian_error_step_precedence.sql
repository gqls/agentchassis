-- 530 — the `diagnosis_guardian` seat's error_step discipline is INVERTED, and
-- it mis-fires on every submission that puts error routing in the PREFERRED
-- place. This is the fix-proposer roster; 531 mirrors it into council-gate.
--
-- WHAT THE SEAT SAYS TODAY (live prompt_template, review_diagnosis_guardian):
--   "CONFIG-LEVEL error_step: the workflow coordinator reads ONLY
--    step.config.error_step -- a step-level error_step is parsed but silently
--    inert (a real, recurring trap). Any plan adding error routing must place
--    it inside config."
-- and, in the judging clause, "(d) does it place error_step outside config
-- (silently inert)".
--
-- WHAT THE COORDINATOR ACTUALLY DOES — platform/orchestration/coordinator.go,
-- routeToErrorStepOrFail (:3666-3679), read 2026-08-21 at HEAD 91cd28919:
--
--     // Check step-level first (parallel to NextStep) — preferred location
--     if step.ErrorStep != "" { return s.routeToErrorStep(...) }
--     // Fallback to config-level for backward compatibility
--     if errorStep, ok := step.Config["error_step"].(string); ok && errorStep != "" { ... }
--
-- The precedence is exactly the reverse of what the seat asserts: step-level is
-- checked FIRST and the code's own comment calls it the preferred location;
-- config-level is the backward-compatibility fallback. BOTH are read. Nothing
-- is inert.
--
-- WHY THIS IS WORTH A MIGRATION RATHER THAN A NOTE. A council seat's prompt is
-- the seat's standing discipline: it objects on it, in every relevant round,
-- against every author who did the RIGHT thing. That is the pathology RFC_022
-- was narrowed to fix — a signal that fires on the recommended remedy stops
-- discriminating, and reviewers learn to wave the seat through. It was caught
-- on `bugs_open/301`'s round (2026-08-19), refuted at source there, and has sat
-- un-fixed since because "tell the seat" had no owner.
--
-- SCOPE. Two verbatim replacements in one prompt_template on one step. No
-- behaviour changes; no code changes; the seat keeps every other discipline and
-- its non-veto status. The bullet is REPLACED rather than deleted because the
-- underlying concern (a plan that drops or misdirects error routing) is real —
-- what was wrong is the test it applied, not that it looked.
--
-- NULL-DIRECTION: every check below counts POSITIVE presence via position() > 0
-- or = 0 on a text value that is asserted NOT NULL first. No jsonb <>-vs-NULL
-- comparison exists here.

\set ON_ERROR_STOP on

BEGIN;

SELECT snapshot_agent('fix-proposer',
  '530_diagnosis_guardian_error_step_precedence.sql: pre-update');

DO $mig$
DECLARE
    v_old_bullet text := '- CONFIG-LEVEL error_step: the workflow coordinator reads ONLY step.config.error_step -- a step-level error_step is parsed but silently inert (a real, recurring trap). Any plan adding error routing must place it inside config.';
    v_new_bullet text := '- error_step, BOTH LOCATIONS ARE LIVE: routeToErrorStepOrFail (coordinator.go) checks the STEP-LEVEL error_step FIRST -- the code comment there calls it the preferred location -- and falls back to step.config.error_step for backward compatibility. Neither is inert. Do NOT object to error routing on placement. Object where a plan REMOVES routing that a failure path needs, or points it at a step that swallows the error.';
    v_old_judge  text := '(d) does it place error_step outside config (silently inert), or move loop work/tokens onto shared pods.';
    v_new_judge  text := '(d) does it remove a step''s error routing, or point it at a step that swallows the failure (placement is NOT the test -- step-level and config-level are both honoured), or move loop work/tokens onto shared pods.';
    v_old text;
    v_new text;
    v_rows int;
BEGIN
    SELECT default_config->'workflow'->'steps'->'review_diagnosis_guardian'->'config'->>'prompt_template'
      INTO v_old
      FROM agent_definitions
     WHERE type = 'fix-proposer' AND is_active
       AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

    IF v_old IS NULL THEN
        RAISE EXCEPTION '530: no live fix-proposer row, or review_diagnosis_guardian has no prompt_template';
    END IF;

    -- Anchors must each occur EXACTLY once; a second occurrence means the
    -- prompt has been restructured and a blind replace would edit the wrong one.
    IF (length(v_old) - length(replace(v_old, v_old_bullet, ''))) / length(v_old_bullet) <> 1 THEN
        RAISE EXCEPTION '530: the CONFIG-LEVEL error_step bullet does not occur exactly once — refusing to replace blind';
    END IF;
    IF (length(v_old) - length(replace(v_old, v_old_judge, ''))) / length(v_old_judge) <> 1 THEN
        RAISE EXCEPTION '530: judging clause (d) does not occur exactly once — refusing to replace blind';
    END IF;

    v_new := replace(replace(v_old, v_old_bullet, v_new_bullet), v_old_judge, v_new_judge);

    -- The reverse replacement must return the body to its EXACT pre-image.
    -- This is the control: it fails if either replacement touched anything else.
    IF replace(replace(v_new, v_new_bullet, v_old_bullet), v_new_judge, v_old_judge) IS DISTINCT FROM v_old THEN
        RAISE EXCEPTION '530: reverse-replacement control failed — the edit is not confined to the two anchors';
    END IF;

    UPDATE agent_definitions
       SET default_config = jsonb_set(default_config,
             ARRAY['workflow','steps','review_diagnosis_guardian','config','prompt_template'],
             to_jsonb(v_new), false),
           updated_at = now()
     WHERE type = 'fix-proposer' AND is_active
       AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

    GET DIAGNOSTICS v_rows = ROW_COUNT;
    IF v_rows <> 1 THEN
        RAISE EXCEPTION '530: expected 1 row, updated %', v_rows;
    END IF;
END $mig$;

-- Verify at the LIVE column, not at the variables above. RAISE, not SELECT:
-- ON_ERROR_STOP ignores a non-empty result set, so only an exception can stop
-- the COMMIT.
DO $verify$
DECLARE v_live text;
BEGIN
    SELECT default_config->'workflow'->'steps'->'review_diagnosis_guardian'->'config'->>'prompt_template'
      INTO v_live
      FROM agent_definitions
     WHERE type = 'fix-proposer' AND is_active
       AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

    IF position('reads ONLY step.config.error_step' in v_live) > 0 THEN
        RAISE EXCEPTION '530 VERIFY: the inverted claim is still live';
    END IF;
    IF position('silently inert' in v_live) > 0 THEN
        RAISE EXCEPTION '530 VERIFY: "silently inert" still appears in the seat prompt';
    END IF;
    IF position('checks the STEP-LEVEL error_step FIRST' in v_live) = 0 THEN
        RAISE EXCEPTION '530 VERIFY: the corrected bullet is not present';
    END IF;
    IF position('placement is NOT the test' in v_live) = 0 THEN
        RAISE EXCEPTION '530 VERIFY: the corrected judging clause is not present';
    END IF;
    -- Untouched-neighbour control: a discipline either side of the edits must
    -- survive. If a replace over-reached, one of these goes missing.
    IF position('THREE-TIER CITATIONS' in v_live) = 0
       OR position('TOKEN/POD ISOLATION' in v_live) = 0 THEN
        RAISE EXCEPTION '530 VERIFY: a neighbouring discipline was lost — the replace over-reached';
    END IF;

    RAISE NOTICE '530 OK: fix-proposer.review_diagnosis_guardian now states the real error_step precedence (step-level first, config-level fallback, neither inert).';
END $verify$;

COMMIT;
