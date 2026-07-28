-- ============================================================================
-- 259_image_build_handler_error_steps_repointed.sql
--   bugs_closed/086 — the owner call the per-handler audit left outstanding
--
-- WHY. Seed 220 disabled ten step-level `error_step` handlers that routed to
-- `complete` (a `complete_workflow` terminal), because the converter fix
-- dca5649b3 armed all 55 declared handlers at once and a failure at those ten
-- would have ended the orchestration GREEN. The audit (7e3a6d89d) then reviewed
-- them one by one and found that eight should stay disabled — six of their
-- agents own no error terminal at all, so `complete` was the author reaching for
-- the only terminal that existed, and one is oufe's via bugs_open/126.
--
-- The two exceptions are image-build-handler's. That agent DOES have a real
-- error terminal, its five `call_*` steps all use it, and these two alone
-- bypassed it. Left disabled they fail generically: the orchestration dies and
-- the triggering site_work_items row is left neither complete nor failed.
--
-- OWNER RULING 2026-07-28: re-point both at `mark_work_item_failed`
-- (option 1 of the three put to him; option 2 was leave disabled, option 3 was
-- re-enable as authored at `-> complete`, argued against because a silent
-- flag_rebuild failure manufactures bugs_open/114 on demand).
--
-- These are the only two of the ten with live traffic — 4 orchestrations each on
-- 2026-07-28, all COMPLETED with __step_error NULL, so the failing branch has
-- still never been exercised. The exposure is real, not theoretical.
--
-- WHAT WAS CHECKED BEFORE WRITING THIS (2026-07-28, live):
--   * the converter carries step-level error_step on the running chassis —
--     v1.0.1192, pod agent-chassis-f757fcf65-bg9t7:
--       grep -c 'Step declares an error_step that is not in the plan' -> 1  (fix marker)
--       grep -c 'routeToErrorStepOrFail'                              -> 3  (positive control)
--       grep -c 'zzz_not_a_real_marker_zzz'                           -> 0  (negative control)
--   * persisted plans really carry it: 66 step-level error_step entries across
--     image-build-handler plans created in the last 2 days.
--   * the target lane works and terminates: 2 orchestrations reached
--     mark_work_item_failed on 07-28, both COMPLETED at current_step
--     'complete_error'. No cycle — mark_work_item_failed's own error_step is
--     complete_error, a terminal.
--   * `mark_work_item_failed` reads input_data.work_item_id with
--     skip_if_missing:true, so a manual trigger without a work item no-ops
--     rather than erroring.
--
-- TWO CONSEQUENCES, STATED RATHER THAN GLOSSED:
--   1. flag_rebuild runs AFTER mark_work_item_complete. If it fails, the row is
--      flipped complete -> failed. UpdateWorkItemStatusAction's UPDATE is
--      unconditional on the current status (v3_site_actions.go:4644-4661) and
--      only the 'complete' branch sets completed_at, so such a row ends up
--      status='failed' with completed_at still populated and attempt_count
--      incremented twice. Odd-looking, and deliberately so: it is a truer record
--      than a green orchestration over unreferenced imagery.
--   2. That flip does NOT cause the image to be regenerated. The claim query
--      takes only 'triaged'/'approved' (claim_work_item_action.go:102), so a
--      failed item parks for triage rather than retrying.
--
-- The real error text is not lost: with no error_message literal and a non-
-- 'complete' status the action records __step_error.message into the row's
-- error column (v3_site_actions.go:4599-4617).
--
-- The `error_step_disabled_086` marker is DROPPED on these two, not kept beside
-- the live key. Keeping it would leave the fleet-wide "how many are disabled"
-- count reading 10 when only 8 are — the marker exists to make an inert handler
-- visible, and these two are no longer inert. Their history lives here, in
-- bugs_closed/086 and in git.
--
-- LIVE IMMEDIATELY — DB config, no image roll needed.
--
-- REVERT (back to disabled, i.e. seed 220's state):
--   UPDATE agent_definitions SET default_config = jsonb_set(
--       default_config #- ARRAY['workflow','steps','flag_rebuild','error_step'],
--       ARRAY['workflow','steps','flag_rebuild','error_step_disabled_086'], '"complete"'::jsonb)
--    WHERE type='image-build-handler' AND is_active
--      AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
--   -- and the same for 'mark_work_item_complete'. Descriptions revert with the
--   -- pre-change snapshot this seed takes (snapshot_reason LIKE '259_%').
-- ============================================================================

DO $$
DECLARE
  def_id     uuid;
  step_json  jsonb;
  missing    int;
BEGIN
  SELECT id INTO def_id FROM agent_definitions
  WHERE type='image-build-handler' AND is_active
    AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

  IF def_id IS NULL THEN
    RAISE EXCEPTION 'no single active image-build-handler row — read the table before rerunning';
  END IF;

  -- Guard: the target step must exist in this agent. An error_step naming a
  -- step that is not in the plan is only a WARN in the converter
  -- (processing.go: "Step declares an error_step that is not in the plan"),
  -- so a typo here would be silently inert — exactly the class 086 is about.
  SELECT count(*) INTO missing FROM agent_definitions
  WHERE id = def_id AND default_config->'workflow'->'steps' ? 'mark_work_item_failed';
  IF missing <> 1 THEN
    RAISE EXCEPTION 'image-build-handler has no mark_work_item_failed step — refusing to point at it';
  END IF;

  -- Snapshot before touching. NOTE: the two-arg form writes to
  -- agent_definitions_backup; the one-arg form writes an is_snapshot row into
  -- agent_definitions itself. Looking in the wrong table has already produced a
  -- confident and wrong "the safety net does not exist".
  PERFORM snapshot_agent('image-build-handler',
    '259_image_build_handler_error_steps_repointed.sql: pre-update');

  -- ---- mark_work_item_complete --------------------------------------------
  SELECT default_config->'workflow'->'steps'->'mark_work_item_complete'
    INTO step_json FROM agent_definitions WHERE id = def_id;

  IF NOT (step_json ? 'error_step_disabled_086') THEN
    RAISE WARNING 'mark_work_item_complete carries no error_step_disabled_086 — already applied, or the row moved';
  ELSE
    step_json := (step_json - 'error_step_disabled_086')
                 || jsonb_build_object(
                      'error_step', 'mark_work_item_failed',
                      'description',
                      'Mark the triggering site_work_items row as complete. Gracefully no-ops if input_data.work_item_id is absent. error_step is mark_work_item_failed (owner ruling 2026-07-28, bugs_closed/086 audit): if the bookkeeping update itself fails, the row must not be left neither complete nor failed. It routes to the agent''s real error terminal, so the run ends at complete_error with the reason recorded on the item, rather than green.');
    UPDATE agent_definitions
       SET default_config = jsonb_set(default_config,
             ARRAY['workflow','steps','mark_work_item_complete'], step_json),
           updated_at = NOW()
     WHERE id = def_id;
    RAISE NOTICE 're-pointed mark_work_item_complete -> mark_work_item_failed';
  END IF;

  -- ---- flag_rebuild --------------------------------------------------------
  SELECT default_config->'workflow'->'steps'->'flag_rebuild'
    INTO step_json FROM agent_definitions WHERE id = def_id;

  IF NOT (step_json ? 'error_step_disabled_086') THEN
    RAISE WARNING 'flag_rebuild carries no error_step_disabled_086 — already applied, or the row moved';
  ELSE
    step_json := (step_json - 'error_step_disabled_086')
                 || jsonb_build_object(
                      'error_step', 'mark_work_item_failed',
                      'description',
                      'Imagery landed: flag the page needs_rebuild and emit needs_page so plan_sections re-resolves the now-present asset. Page scope uses scope_ref directly; section scope (scope_ref "<page>:<ordinal>") maps to its page prefix. Still a no-op for site scope (logo). error_step is mark_work_item_failed (owner ruling 2026-07-28, bugs_closed/086 audit): a silent failure here leaves imagery generated, deployed and never referenced, which is bugs_open/114. This step runs after mark_work_item_complete, so the failure path flips the row complete -> failed — deliberately, because a failed row with the reason attached is a truer record than a green run over unreferenced imagery. It does not trigger a retry: the claim query takes only triaged/approved.');
    UPDATE agent_definitions
       SET default_config = jsonb_set(default_config,
             ARRAY['workflow','steps','flag_rebuild'], step_json),
           updated_at = NOW()
     WHERE id = def_id;
    RAISE NOTICE 're-pointed flag_rebuild -> mark_work_item_failed';
  END IF;
END $$;

-- ---------------------------------------------------------------------------
-- post-checks
-- ---------------------------------------------------------------------------

-- 1: the two steps are live and point at the failure lane (want 2 rows,
--    error_step = mark_work_item_failed, disabled_086 empty)
SELECT s.key AS step,
       s.value->>'error_step'              AS error_step,
       s.value->>'error_step_disabled_086' AS disabled_086
FROM agent_definitions d, jsonb_each(d.default_config->'workflow'->'steps') s
WHERE d.type='image-build-handler' AND d.is_active
  AND COALESCE(d.is_snapshot,false)=false AND d.deleted_at IS NULL
  AND s.key IN ('mark_work_item_complete','flag_rebuild')
ORDER BY 1;

-- 2: nothing routes to `complete` at step level anywhere on the fleet (want 0)
--    — seed 220's invariant, unchanged by this seed
SELECT count(*) AS still_routing_to_complete
FROM agent_definitions, jsonb_each(default_config->'workflow'->'steps')
WHERE is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL
  AND value->>'error_step'='complete' AND value->'config'->>'error_step' IS NULL;

-- 3: the other eight stay disabled and visible (want 8, was 10)
SELECT count(*) AS still_disabled_086
FROM agent_definitions, jsonb_each(default_config->'workflow'->'steps')
WHERE is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL
  AND value ? 'error_step_disabled_086';

-- 4: live step-level handlers fleet-wide (want 46, was 44 — the two re-armed)
SELECT count(*) AS live_step_level_handlers
FROM agent_definitions, jsonb_each(default_config->'workflow'->'steps')
WHERE is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL
  AND value->>'error_step' IS NOT NULL AND value->'config'->>'error_step' IS NULL;

-- 5: the snapshot exists (want >= 1 row, taken just now)
SELECT type, snapshot_taken_at, snapshot_reason
FROM agent_definitions_backup
WHERE snapshot_reason LIKE '259_%'
ORDER BY snapshot_taken_at DESC LIMIT 3;

-- 6: no step in this agent points at a step that does not exist (want 0 rows)
SELECT s.key AS step, s.value->>'error_step' AS dangling_error_step
FROM agent_definitions d, jsonb_each(d.default_config->'workflow'->'steps') s
WHERE d.type='image-build-handler' AND d.is_active
  AND COALESCE(d.is_snapshot,false)=false AND d.deleted_at IS NULL
  AND s.value->>'error_step' IS NOT NULL
  AND NOT (d.default_config->'workflow'->'steps' ? (s.value->>'error_step'));
