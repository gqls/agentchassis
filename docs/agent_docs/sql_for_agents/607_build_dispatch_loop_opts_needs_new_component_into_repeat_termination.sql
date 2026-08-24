-- 607_build_dispatch_loop_opts_needs_new_component_into_repeat_termination.sql
--
-- bugs_open/345 candidate 2, the OPT-IN. The mechanism (byte-identical repeat
-- failures end the budget early, applyWorkItemFailureLadder) is live on
-- v1.0.1334+ and council-APPROVED (Council-Reviewed:
-- f1f1fc37-35e9-45fd-88d7-fcc3ddcf9eb0), but it is a per-item-type opt-in in
-- the CALLER's step config with the unsafe default OFF (RFC_010 §2), and as of
-- 2026-08-24 NOTHING opts in — measured with strpos, not LIKE (underscore is a
-- LIKE wildcard), with a positive control (last_error_code found in 2 agents)
-- and a negative one (invented key found in 0).
--
-- WHY needs_new_component, AND WHY NOW. The Fable adversarial review (bug file
-- 2026-08-24 night section) found the demand is live TODAY, not historical:
-- item 2396218a took 3 byte-identical rejections on 2026-08-24 with attempts 2
-- and 3 CARRYING the typed retry feedback (candidate 1) and failing anyway;
-- three more items sat triaged with identical-rejection histories. The honest
-- benefit is ~1 generation per repeat-failing item (the 52-burn was the
-- pre-307 uncounted loop and cannot recur) — but each firing is also the
-- SIGNAL that feedback did not change the outcome, which nothing else records.
--
-- ⚠ WHAT A FIRING MEANS — CLASS-CONDITIONAL, do not over-read it (F2a): the
-- retry prompt renders only for the 3 producer validation codes (migration
-- 563). A firing on an item whose terminal error is one of those codes means
-- "feedback was shown and did not help". A firing on any OTHER error class
-- (e.g. the chk_section_type_kebab_case insert failure on item eb38ae2b)
-- means "this class HAS no feedback channel" — deterministic, un-fed, and
-- stopping it early is pure saving.
--
-- ⚠ HOW A FIRING IS READ. The Go marker (result.terminated_on_repeat, same
-- commit as this file) is INERT until the next chassis roll. Until that rolls,
-- a firing is visible only as: status='failed' AND attempt_count <
-- max_attempts on this item type after this migration's applied_at — and that
-- proxy has a ~7-day shelf life (the archiver drains terminal rows, WII-024).
--
-- ⚠ THE LATENT CONSTANT CHANNEL THIS ARMS (F5). This very step's config
-- carries error_message: "Handler failed" — the fallback literal used only
-- when the routed __step_error is missing. Two DIFFERENT failures that both
-- fall back to it become byte-identical and would repeat-terminate falsely.
-- 0 rows fleet-wide carry that text today (the guard below RAISEs NOTICE with
-- the live count rather than trusting this comment); if that count ever goes
-- non-zero, read bugs_open/345 F5 before trusting any firing census.
--
-- SCOPE (F6): ONLY build-dispatch-loop's nested mark_failed step —
-- workflow.steps.process_item.config.sub_workflow.steps.mark_failed.config.
-- That is the ONLY failure step needs_new_component items can reach: 21/21
-- ladder-written failures of this type carry handled_by='build-dispatch-loop',
-- and site-work-orchestrator's loops either filter to a different handler or
-- have no failure step at all. **1 covering step as of 2026-08-24** — the
-- opt-in is per call-site by design, so a future failure step does NOT
-- inherit it; re-run the census in the bug file before assuming coverage.
--
-- ORDERING: none. The ladder reads the key defensively (absent = today's
-- behaviour, byte-identical), and the binary carrying the mechanism is
-- already live. Applying this before the marker's roll only affects HOW a
-- firing is read (see above), not whether it is safe.
--
-- ROLLBACK: 607_..._ROLLBACK.sql removes the single key.

BEGIN;

DO $guard$
DECLARE
  cfg jsonb;
  n int;
  handler_failed_rows int;
BEGIN
  SELECT count(*) INTO n FROM agent_definitions
   WHERE type='build-dispatch-loop' AND is_active
     AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
  IF n <> 1 THEN
    RAISE EXCEPTION '607: expected exactly 1 live build-dispatch-loop row, found %', n;
  END IF;

  SELECT default_config->'workflow'->'steps'->'process_item'->'config'->'sub_workflow'
         ->'steps'->'mark_failed'
    INTO cfg
  FROM agent_definitions
   WHERE type='build-dispatch-loop' AND is_active
     AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

  IF cfg IS NULL THEN
    RAISE EXCEPTION '607: mark_failed not found at the expected nested path — the workflow has been restructured; re-read it before editing';
  END IF;
  IF cfg->>'action' <> 'fail_work_item' THEN
    RAISE EXCEPTION '607: mark_failed action is %, not fail_work_item — this is not the step this migration was written against', cfg->>'action';
  END IF;

  -- The anchor keys read from the live row on 2026-08-24; a restructured
  -- config aborts here rather than being silently reshaped.
  IF NOT (cfg->'config' ? 'work_item_id' AND cfg->'config' ? 'error_message') THEN
    RAISE EXCEPTION '607: mark_failed config is missing its anchor keys (work_item_id/error_message)';
  END IF;

  IF cfg->'config' ? 'stop_on_repeat_failure_item_types' THEN
    RAISE EXCEPTION '607: stop_on_repeat_failure_item_types already present — another session has applied this';
  END IF;

  -- The latent constant channel (F5), counted live rather than asserted:
  -- "Handler failed" is this step's fallback error_message, and any row
  -- carrying it verbatim is one where two DIFFERENT failures could look
  -- byte-identical to the repeat rule this migration arms.
  SELECT count(*) INTO handler_failed_rows FROM site_work_items WHERE error = 'Handler failed';
  RAISE NOTICE '607: rows carrying the bare "Handler failed" fallback: % (expected 0 — if non-zero, read bugs_open/345 F5 before trusting any firing census)', handler_failed_rows;

  -- Shop convention (run-migrations.sh header): every migration touching
  -- agent_definitions opens with a snapshot, so the pre-write row is
  -- recoverable without git archaeology. Taken only after every guard has
  -- passed — a refused apply must not leave a snapshot implying it ran.
  PERFORM snapshot_agent('build-dispatch-loop',
    '607_build_dispatch_loop_opts_needs_new_component_into_repeat_termination.sql: pre-update');
END
$guard$;

UPDATE agent_definitions
SET default_config = jsonb_set(
      default_config,
      '{workflow,steps,process_item,config,sub_workflow,steps,mark_failed,config,stop_on_repeat_failure_item_types}',
      '["needs_new_component"]'::jsonb,
      true)
WHERE type='build-dispatch-loop' AND is_active
  AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

DO $verify$
DECLARE
  c jsonb;
BEGIN
  SELECT default_config->'workflow'->'steps'->'process_item'->'config'->'sub_workflow'
         ->'steps'->'mark_failed'->'config'
    INTO c
  FROM agent_definitions
   WHERE type='build-dispatch-loop' AND is_active
     AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

  IF c->'stop_on_repeat_failure_item_types' IS DISTINCT FROM '["needs_new_component"]'::jsonb THEN
    RAISE EXCEPTION '607 VERIFY: stop_on_repeat_failure_item_types is %, not ["needs_new_component"]',
      c->'stop_on_repeat_failure_item_types';
  END IF;
  IF NOT (c ? 'work_item_id' AND c ? 'error_message') THEN
    RAISE EXCEPTION '607 VERIFY: a pre-existing config key was lost';
  END IF;
  IF (SELECT count(*) FROM jsonb_object_keys(c)) <> 3 THEN
    RAISE EXCEPTION '607 VERIFY: expected exactly 3 config keys after the write, found %',
      (SELECT count(*) FROM jsonb_object_keys(c));
  END IF;

  RAISE NOTICE '607 OK: needs_new_component is opted in to repeat termination on build-dispatch-loop.mark_failed';
END
$verify$;

COMMIT;
