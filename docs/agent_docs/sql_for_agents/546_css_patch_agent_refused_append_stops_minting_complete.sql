-- 546_css_patch_agent_refused_append_stops_minting_complete.sql
--
-- bugs_open/198 — the arm migration 542 MISSED, found hours later by validating the
-- workflow GRAPH rather than by reading the steps I had edited.
--
-- 542 rewired every `error_step` (plan_css_fix, save_css_to_db, deploy_css) through a
-- status-stamping step so a failure could no longer ride a success-labelled
-- `complete_workflow` and read as `complete`. It did not touch `check_saved`, because
-- `check_saved` is not an error_step — it is a `conditional_branch`, and its refusal
-- travels on `config.else_step`. That edge still pointed straight at `complete_error`.
--
-- So one door was left open, and it is the door 318 built on purpose:
--
--   `save_css_to_db` guards its append with `length($2) BETWEEN 1 AND 8192`. When the
--   model returns an empty or oversized `css_added` — the whole-document echo that
--   caused the ORIGINAL 2026-08-04 incident — the UPDATE matches ZERO rows. 318 added
--   `check_saved` precisely so that refusal would be loud rather than riding
--   git_commit's "no files → skipped, Success:true" path. It routes to `complete_error`,
--   which is a SUCCESS-labelled `complete_workflow`, so the parent dispatch loop stamps
--   the work item `complete`.
--
-- The guard fires, correctly, and the ledger records a repair that did not happen. That
-- is the same defect 542 exists to close, one step over — and worse here than elsewhere,
-- because this is the arm that catches the bug's founding failure mode.
--
-- HOW IT WAS FOUND, because the method transfers: a query that walks EVERY edge in the
-- workflow (`next_step`, `error_step`, `config.then_step`, `config.else_step`) and
-- resolves each target against the step map. It was written to catch a DANGLING edge
-- after 542's rewire; every edge resolved, and reading the resulting 18-row table is
-- what showed `check_saved | else | complete_error` sitting there untouched. **Reading
-- the steps you edited cannot find the step you did not edit.** The query is in the lane
-- RUNBOOK.
--
-- THE FIX: route it through `mark_append_refused` — a new stamping step, not the
-- existing `mark_step_failed`. Two reasons they are separate:
--
--   1. `mark_step_failed` carries NO literal `error_message`, deliberately, so that the
--      routed `__step_error` is what gets recorded. Here there IS no step error: the
--      query succeeded and returned zero rows. It would record nothing useful.
--   2. The remedy differs and an operator should be told which one they are looking at.
--      A refused append means the MODEL produced an unusable patch (empty, or > 8192
--      chars — i.e. it echoed a document instead of emitting rules). That is worth
--      saying in the row.
--
-- Status is `failed`, not `needs_human_review`: an unusable completion is a transient
-- model failure, and `failed` goes through `applyWorkItemFailureLadder` — retried with
-- backoff, counted in the promoter floor's denominator. A refusal that a retry could fix
-- must not be parked for a human (that is the distinction 542's `mark_base_unsafe` sits
-- on the other side of: an unsafe BASE cannot be retried away).
--
-- CONFIG IS LIVE IMMEDIATELY ON APPLY.

DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM agent_definitions
        WHERE type = 'css-patch-agent'
          AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL
          AND default_config #> '{workflow,steps}' ? 'mark_append_refused'
    ) THEN
        RAISE EXCEPTION '198/546: already applied — mark_append_refused already exists';
    END IF;
END $$;

BEGIN;

SELECT snapshot_agent('css-patch-agent',
  '546_css_patch_agent_refused_append_stops_minting_complete: pre-update');

-- ── DRIFT GUARD ────────────────────────────────────────────────────────────────
DO $$
DECLARE
    v_steps jsonb;
BEGIN
    SELECT default_config #> '{workflow,steps}'
      INTO v_steps
      FROM agent_definitions
     WHERE type = 'css-patch-agent'
       AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

    IF v_steps IS NULL THEN
        RAISE EXCEPTION '198/546: no live css-patch-agent row found';
    END IF;

    -- 542 must be in place: this migration completes its work and its reasoning
    -- assumes the other three exits are already stamped.
    IF NOT (v_steps ? 'mark_step_failed') THEN
        RAISE EXCEPTION '198/546: migration 542 is not applied (mark_step_failed absent) — apply it first';
    END IF;

    -- the edge being rewired, at its current value
    IF v_steps #>> '{check_saved,config,else_step}' <> 'complete_error' THEN
        RAISE EXCEPTION '198/546 drift: check_saved.else_step is %, expected complete_error',
            v_steps #>> '{check_saved,config,else_step}';
    END IF;
    IF v_steps #>> '{check_saved,config,condition}' <> 'css_saved.count >= 1' THEN
        RAISE EXCEPTION '198/546 drift: check_saved.condition is %, expected css_saved.count >= 1',
            v_steps #>> '{check_saved,config,condition}';
    END IF;

    -- and the guard whose refusal this arm reports — if the size bounds moved, the
    -- error message below would name the wrong number.
    IF position('BETWEEN 1 AND 8192' in (v_steps #>> '{save_css_to_db,config,query}')) = 0 THEN
        RAISE EXCEPTION '198/546 drift: save_css_to_db no longer carries the 1..8192 size guard';
    END IF;
END $$;

UPDATE agent_definitions
   SET default_config = jsonb_set(
         jsonb_set(
           default_config,
           '{workflow,steps,mark_append_refused}',
           jsonb_build_object(
             'action', 'update_work_item_status',
             'description',
               'bugs_open/198: the guarded append matched zero rows — the model returned an empty or '
               || 'oversized css_added. 318 made that refusal loud; without this step it rode '
               || 'complete_error (a success-labelled complete_workflow) and the item still read complete.',
             'next_step', 'complete_error',
             'config', jsonb_build_object(
               'status',        'failed',
               'error_message',
                 'css-patch refused the append (bugs_open/198): the guarded UPDATE matched no row, which '
                 || 'means css_added was empty or longer than 8192 characters — i.e. the model echoed a '
                 || 'whole document instead of emitting the new rules. Nothing was written and nothing was '
                 || 'deployed. This is retryable: the next attempt gets a fresh completion.',
               'result_fields', jsonb_build_object('failed_by', 'css_append_guard_198')
             )
           )
         ),
         '{workflow,steps,check_saved,config,else_step}',
         to_jsonb('mark_append_refused'::text)
       ),
       updated_at = NOW()
 WHERE type = 'css-patch-agent'
   AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

-- ── VERIFY ─────────────────────────────────────────────────────────────────────
DO $$
DECLARE
    v_steps jsonb;
    v_dangling int;
BEGIN
    SELECT default_config #> '{workflow,steps}'
      INTO v_steps
      FROM agent_definitions
     WHERE type = 'css-patch-agent'
       AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

    IF v_steps #>> '{check_saved,config,else_step}' <> 'mark_append_refused' THEN
        RAISE EXCEPTION '198/546 verify: check_saved.else_step not rewired';
    END IF;
    IF v_steps #>> '{mark_append_refused,config,status}' <> 'failed' THEN
        RAISE EXCEPTION '198/546 verify: mark_append_refused must record failed';
    END IF;
    IF v_steps #>> '{mark_append_refused,next_step}' <> 'complete_error' THEN
        RAISE EXCEPTION '198/546 verify: mark_append_refused does not reach a terminal';
    END IF;

    -- The check that found this defect, now run as a POST-CONDITION: every edge in the
    -- whole workflow must resolve to a real step. This is what catches the arm you did
    -- not think to look at — including any this migration orphans.
    SELECT count(*) INTO v_dangling
      FROM (
        SELECT e.v->>'next_step' AS tgt FROM jsonb_each(v_steps) AS e(k,v) WHERE e.v ? 'next_step'
        UNION ALL
        SELECT e.v->>'error_step' FROM jsonb_each(v_steps) AS e(k,v) WHERE e.v ? 'error_step'
        UNION ALL
        SELECT e.v->'config'->>'then_step' FROM jsonb_each(v_steps) AS e(k,v) WHERE e.v->'config' ? 'then_step'
        UNION ALL
        SELECT e.v->'config'->>'else_step' FROM jsonb_each(v_steps) AS e(k,v) WHERE e.v->'config' ? 'else_step'
      ) AS edges
     WHERE tgt IS NOT NULL AND NOT (v_steps ? tgt);

    IF v_dangling > 0 THEN
        RAISE EXCEPTION '198/546 verify: % workflow edge(s) point at a step that does not exist', v_dangling;
    END IF;

    -- and the property this whole sub-task is about: NO non-success exit reaches a
    -- terminal without stamping the item first.
    IF v_steps #>> '{check_has_css,config,else_step}' <> 'mark_no_css'
       OR v_steps #>> '{check_saved,config,else_step}' <> 'mark_append_refused'
       OR v_steps #>> '{plan_css_fix,error_step}' <> 'mark_step_failed'
       OR v_steps #>> '{save_css_to_db,error_step}' <> 'mark_step_failed'
       OR v_steps #>> '{deploy_css,error_step}' <> 'mark_step_failed'
       OR v_steps #>> '{load_current_css,error_step}' <> 'mark_no_css'
       OR v_steps #>> '{check_base_integrity,config,else_step}' <> 'mark_base_unsafe' THEN
        RAISE EXCEPTION '198/546 verify: a non-success exit still reaches a terminal unstamped';
    END IF;

    RAISE NOTICE '198/546: verified — all 7 non-success exits stamp before their terminal, 0 dangling edges';
END $$;

COMMIT;
