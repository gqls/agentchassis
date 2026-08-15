-- 427 - model-directory-publisher: drop the error_step routing added by 411
-- so a failed git-adapter commit FAILS the workflow loudly instead of
-- silently continuing to the next kind.
--
-- Why: council review of 411 (corr a3c418ea-4452-420d-b6e8-62ce78d5339e).
-- Round 1 flagged the error_step routing (a failed commit on one kind
-- routes past the failure into the next kind's render step) as a genuinely
-- new design choice not restoring any known prior behaviour - the FINDING
-- this migration follows on from explicitly could not recover what the
-- original 07-26 config's failure semantics were. Round 2 (bug_historian,
-- gating HIGH, every other seat approved) escalated: shipping that
-- mechanism live matches this platform's most-repeated recurring defect
-- shape - something discarded/skipped on failure with no fail-loud signal,
-- caught only after real loss (016b §9) - and additionally noted the
-- routing's whole safety story rests on step-level error_step surviving
-- plan conversion (bugs_closed/086's mechanism), which the plan never
-- checked for THIS agent's steps.
--
-- Rather than build a new fail-loud alert/work-item mechanism (itself a
-- second, unreviewed behaviour change riding inside a migration whose own
-- risks section already declined scope creep twice), the simpler and more
-- conservative fix is to remove the invented behaviour entirely: no
-- error_step means a failed step falls through to the coordinator's default
-- failWorkflow path, which marks the orchestration FAILED - visible in
-- orchestration_states and monitored exactly like any other workflow
-- failure on this platform, not silently masked as COMPLETED. This is a
-- smaller diff than 411 itself and removes the self-admitted "unverified
-- vs original intent" design choice rather than defending it further.
--
-- next_step chaining (the happy-path model->company->protocol sequence) is
-- UNCHANGED - only the three error_step keys 411 added are removed.

SELECT snapshot_agent('model-directory-publisher', '427_model_directory_publisher_drop_silent_continue_on_commit_failure.sql: pre-update');

BEGIN;

DO $do$
DECLARE
    n_active int;
    steps jsonb;
    n_steps int;
BEGIN
    SELECT count(*) INTO n_active FROM agent_definitions
    WHERE type = 'model-directory-publisher' AND is_active
      AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;
    IF n_active <> 1 THEN
        RAISE EXCEPTION '427: model-directory-publisher does not have exactly one active row (found %) - resolve before seeding', n_active;
    END IF;

    SELECT default_config#>'{workflow,steps}' INTO steps FROM agent_definitions
    WHERE type = 'model-directory-publisher' AND is_active
      AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

    SELECT count(*) INTO n_steps FROM jsonb_object_keys(steps);
    IF n_steps <> 7 THEN
        RAISE EXCEPTION '427: expected 7 steps (migration 411 not applied, or row has drifted further), found %', n_steps;
    END IF;

    -- Idempotency: already applied if none of the three commit steps still
    -- carries an error_step key.
    IF NOT (steps #> '{commit_model_directory}' ? 'error_step'
            OR steps #> '{commit_adoption_directory}' ? 'error_step'
            OR steps #> '{commit_protocol_directory}' ? 'error_step') THEN
        RAISE EXCEPTION '427: already applied - no commit step carries an error_step key';
    END IF;

    -- Drift guard: pin to exactly what 411 left behind.
    IF steps #>> '{commit_model_directory,error_step}' <> 'render_adoption_json' THEN
        RAISE EXCEPTION '427: commit_model_directory.error_step has drifted from 411''s known value - re-check before reapplying';
    END IF;
    IF steps #>> '{commit_adoption_directory,error_step}' <> 'render_protocol_json' THEN
        RAISE EXCEPTION '427: commit_adoption_directory.error_step has drifted from 411''s known value - re-check before reapplying';
    END IF;
    IF steps #>> '{commit_protocol_directory,error_step}' <> 'complete' THEN
        RAISE EXCEPTION '427: commit_protocol_directory.error_step has drifted from 411''s known value - re-check before reapplying';
    END IF;
    -- next_step must be untouched by this migration.
    IF steps #>> '{commit_model_directory,next_step}' <> 'render_adoption_json'
       OR steps #>> '{commit_adoption_directory,next_step}' <> 'render_protocol_json'
       OR steps #>> '{commit_protocol_directory,next_step}' <> 'complete' THEN
        RAISE EXCEPTION '427: a commit step''s next_step has drifted from 411''s known value - re-check before reapplying';
    END IF;
END $do$;

UPDATE agent_definitions
SET default_config = default_config
        #- '{workflow,steps,commit_model_directory,error_step}'
        #- '{workflow,steps,commit_adoption_directory,error_step}'
        #- '{workflow,steps,commit_protocol_directory,error_step}',
    updated_at = NOW()
WHERE type = 'model-directory-publisher' AND is_active
  AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

DO $do$
DECLARE
    steps jsonb;
    n_steps int;
BEGIN
    SELECT default_config#>'{workflow,steps}' INTO steps FROM agent_definitions
    WHERE type = 'model-directory-publisher' AND is_active
      AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

    SELECT count(*) INTO n_steps FROM jsonb_object_keys(steps);
    IF n_steps <> 7 THEN
        RAISE EXCEPTION '427: verify failed - step count changed, expected 7, got %', n_steps;
    END IF;
    IF steps #> '{commit_model_directory}' ? 'error_step'
       OR steps #> '{commit_adoption_directory}' ? 'error_step'
       OR steps #> '{commit_protocol_directory}' ? 'error_step' THEN
        RAISE EXCEPTION '427: verify failed - a commit step still carries error_step';
    END IF;
    -- next_step chain must survive untouched.
    IF steps #>> '{commit_model_directory,next_step}' <> 'render_adoption_json'
       OR steps #>> '{commit_adoption_directory,next_step}' <> 'render_protocol_json'
       OR steps #>> '{commit_protocol_directory,next_step}' <> 'complete' THEN
        RAISE EXCEPTION '427: verify failed - the happy-path next_step chain was disturbed';
    END IF;
    IF steps #>> '{render_adoption_json,config,kind}' <> 'company'
       OR steps #>> '{render_protocol_json,config,kind}' <> 'protocol' THEN
        RAISE EXCEPTION '427: verify failed - kind literals were disturbed';
    END IF;
END $do$;

COMMIT;

-- ROLLBACK recipe (hand-run, restores 411's error_step routing from the
-- agent_definitions_backup row this migration took via snapshot_agent()):
--   UPDATE agent_definitions live
--   SET default_config = bak.default_config
--   FROM (SELECT default_config FROM agent_definitions_backup
--         WHERE type = 'model-directory-publisher'
--           AND snapshot_reason = '427_model_directory_publisher_drop_silent_continue_on_commit_failure.sql: pre-update'
--         ORDER BY snapshot_taken_at DESC LIMIT 1) bak
--   WHERE live.type = 'model-directory-publisher' AND live.is_active
--     AND COALESCE(live.is_snapshot, false) = false AND live.deleted_at IS NULL;
