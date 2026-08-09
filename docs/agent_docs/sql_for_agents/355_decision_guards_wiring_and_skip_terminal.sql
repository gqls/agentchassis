-- 355 — RFC_015 stage 2/3 wiring: drive the decision_guards check, and let a
--       gated refusal terminate cleanly
--
-- WHY (both halves measured live 2026-08-09, after the roll that shipped the
-- RFC_015 Go code):
--
-- A. THE GUARD CHECK IS UNDRIVEN. check_decision_guards.go registers itself in
--    Go, but every discovery agent runs an explicit allow-list of check names
--    (quality-discovery-agent: 6 names; design-discovery-agent: 23). A check
--    absent from every list never runs, and its silence is indistinguishable
--    from "nothing to report" — the exact shape of the fleet lesson "a silent
--    mechanism is usually UNDRIVEN, not missing". Proven: a throwaway decision
--    with a deliberately impossible guard produced ZERO findings across a full
--    quality-discovery run.
--
-- B. A GATED REFUSAL REPORTS AS A FAILED ITEM. apply_section_edit's decision
--    gate (and the lock gate it copies) returns {success:true, skipped:true}
--    precisely so a refusal is NOT a failed orchestration. The workflow then
--    walks on to deploy_page → update_page_status, which cannot find a page_id
--    for an edit that never happened, so the item lands `failed` with
--    "could not determine page_id". Measured on work item 27d6b485 (the
--    deliberate no-citation proof): the gate itself behaved perfectly —
--    edit_result carried decision_gated:true and named D-001 — and the item
--    still read as a failure. The lock gate has ALWAYS had this latent flaw;
--    it was never exercised because no section_edit had hit a locked component
--    (all 13 historical section_edit items completed). The decision gate is
--    simply the first thing to walk that path.
--
-- SCOPE: config only, live immediately, no roll. Additive for A (one name
-- appended to one list); B adds a skip branch that only fires on a result that
-- previously crashed downstream, so no working path changes.
--
-- SHAPE NOTE (checked, not assumed): the branch uses action `conditional_branch`
-- with a STRING condition and then_step/else_step — the shape live agents use
-- (webdesign-agent check_update_db). An invented {condition_field, operator,
-- true_step} shape was written first and would have failed at runtime, which
-- `go build` cannot see. A missing `skipped` key resolves to nil and takes
-- else_step, so every normal edit is unaffected. `conditional` is registered
-- but Deprecated; `conditional_branch` is the live name.
--
-- RE-RUN SAFE: both statements are idempotent (guarded by a WHERE that stops
-- matching once applied); the DO block reports rather than raising, because a
-- partially-applied state here is benign.

BEGIN;

SELECT snapshot_agent('quality-discovery-agent', 'pre-355: add decision_guards to the check list (RFC_015)');
SELECT snapshot_agent('section-editor',          'pre-355: skip branch so a gated refusal terminates cleanly (RFC_015)');

-- A. drive the check
UPDATE agent_definitions
   SET default_config = jsonb_set(default_config,
         '{workflow,steps,run_checks,config,checks}',
         (default_config->'workflow'->'steps'->'run_checks'->'config'->'checks') || '["decision_guards"]'::jsonb)
 WHERE type = 'quality-discovery-agent'
   AND is_active AND COALESCE(is_snapshot,false) = false AND deleted_at IS NULL
   AND NOT (default_config->'workflow'->'steps'->'run_checks'->'config'->'checks' @> '["decision_guards"]'::jsonb);

-- B. a skipped edit goes straight to complete, never to deploy/update_page_status
UPDATE agent_definitions
   SET default_config = jsonb_set(
         jsonb_set(default_config,
           '{workflow,steps,check_edit_skipped}',
           '{"action":"conditional_branch",
             "config":{"condition":"edit_result.skipped == true",
                       "then_step":"complete","else_step":"deploy_page"},
             "description":"RFC_015: a decision-gated or lock-gated refusal is a clean terminal skip, not a failed deploy (see 355 header B). A normal edit has no edit_result.skipped, which resolves to nil and takes else_step."}'::jsonb),
         '{workflow,steps,apply_edit,next_step}', '"check_edit_skipped"')
 WHERE type = 'section-editor'
   AND is_active AND COALESCE(is_snapshot,false) = false AND deleted_at IS NULL
   AND default_config->'workflow'->'steps'->'apply_edit'->>'next_step' = 'deploy_page';

DO $$
DECLARE has_check bool; branch text;
BEGIN
  SELECT default_config->'workflow'->'steps'->'run_checks'->'config'->'checks' @> '["decision_guards"]'::jsonb
    INTO has_check FROM agent_definitions
   WHERE type='quality-discovery-agent' AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
  SELECT default_config->'workflow'->'steps'->'apply_edit'->>'next_step'
    INTO branch FROM agent_definitions
   WHERE type='section-editor' AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
  IF NOT COALESCE(has_check,false) THEN
    RAISE EXCEPTION '355A failed: decision_guards absent from quality-discovery-agent checks';
  END IF;
  IF branch IS DISTINCT FROM 'check_edit_skipped' THEN
    RAISE EXCEPTION '355B failed: apply_edit.next_step is %, expected check_edit_skipped', branch;
  END IF;
  RAISE NOTICE '355 applied: decision_guards driven; section-editor skip branch live';
END $$;

COMMIT;
