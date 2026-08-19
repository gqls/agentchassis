-- 495 — component-template-fixer HONOURS ITS OWN REFUSAL FLAG: a run whose
--       fix_result carries action='needs_review' parks the work item at
--       needs_human_review instead of completing it (bugs_open/323).
--
-- WHY. FixComponentTemplateAction has, since 2026-03-14, returned
--   {fixed:false, action:"needs_review", reason:"fix_type requires LLM-driven
--    changes, not programmatic HTML edits"}
-- for every cta_improvement / nav_restructure item, and the same action key on
-- every other REFUSAL arm in the file (locked chrome slot, unrecognised fix_type,
-- chrome_overflow_fix with no/unsafe slot or selector, repair_page_component_status
-- refusals) — 13 emission sites. Its idempotent no-op arms ("already has flex CSS",
-- "already deployed", "already patched") carry NO action key. So the handler's own
-- vocabulary already separates "I cannot / will not" from "nothing to do".
--
-- NOTHING READS IT. This workflow branches check_needs_rerender on fix_result.fixed
-- only; its else path is compose_note → append_note → complete_workflow, and the
-- build-dispatch-loop then calls complete_work_item, whose gates read
-- response.status (never set here) and the opt-in numeric roster (cta_improvement is
-- not on it). The comment at fix_component_template_action.go:58 saying the flag
-- "stops the dispatch loop recording the work item as done" is false.
--
-- [MEASURED 2026-08-19, live+archive, every component-template-fixer row ever]:
--   action=needs_review          470  (468 cta_improvement + 2 responsive_fix "missing slot_name")
--   fixed=false, no action key   299  (226 spacing_fix + 72 responsive_fix + 1 instance_scope_conversion)
--   overlap either way             0
-- So this change would have parked 470 refusals and blocked 0 legitimate completions.
-- cta_improvement has NEVER reported fixed=true (0 of 993 lifetime, 22 sites).
--
-- PATTERN, NOT INVENTION. page-build-handler's mark_needs_review / mark_writer_skipped
-- ("park the work item visibly instead of letting the dispatch loop stamp it complete")
-- use fail_work_item + status_override=needs_human_review; the 283 lane's pending
-- 486_HOLD adds judged_refusal to THIS agent with the same shape, keyed on the same
-- fix_result.action field. complete_work_item's guard (load_work_item_actions.go,
-- "status NOT IN ('needs_human_review', …)") then preserves the parked status — no
-- change to the shared completion path.
--
-- COMPOSES WITH 486_HOLD. 486 rewires apply_fix.next_step (→ check_scope_route) and
-- guards on its pre-image value. This file deliberately hangs the new edge off
-- check_needs_rerender.else_step (pre-image 'compose_note') and touches neither
-- apply_fix nor any step 486 adds, so the two apply in either order.
--
-- SIDE EFFECT, STATED. Once the row is parked, complete_work_item skips its result
-- write, so the handler's reason is NOT in site_work_items.result. It IS in
-- site_work_items.error (the literal below names the mechanism), in the run's
-- orchestration state, and — because compose_note's prompt is amended here to title
-- refusals and quote fix_result.reason — in the doc_notes 'fix' entry for the site.
--
-- ROLLBACK: 495_fixer_parks_refusals_as_needs_human_review_ROLLBACK.sql.
BEGIN;

-- 0. Double-application guard BEFORE the snapshot (a second apply would snapshot the
--    patched row as a "pre-image" and poison a newest-first rollback).
DO $$
DECLARE cfg jsonb; n int;
BEGIN
  SELECT count(*) INTO n FROM agent_definitions
   WHERE type='component-template-fixer' AND is_active
     AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
  IF n <> 1 THEN
    RAISE EXCEPTION '495: expected exactly 1 live component-template-fixer row, found %', n;
  END IF;
  SELECT default_config INTO cfg FROM agent_definitions
   WHERE type='component-template-fixer' AND is_active
     AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
  IF (cfg #> '{workflow,steps}') ? 'check_refused' OR (cfg #> '{workflow,steps}') ? 'park_refused' THEN
    RAISE EXCEPTION '495: already applied (check_refused/park_refused present). Nothing to do.';
  END IF;
  IF cfg #>> '{workflow,steps,check_needs_rerender,config,else_step}' <> 'compose_note' THEN
    RAISE EXCEPTION '495: pre-image drift — check_needs_rerender.else_step is % (expected compose_note). Re-derive before applying.',
      cfg #>> '{workflow,steps,check_needs_rerender,config,else_step}';
  END IF;
  IF cfg #>> '{workflow,steps,compose_note,config,prompt_template}'
       NOT LIKE '%If the fix result shows fixed=false, title it no-op fix pass and say nothing changed.%' THEN
    RAISE EXCEPTION '495: pre-image drift — compose_note prompt no longer ends with the expected sentence. Re-derive before applying.';
  END IF;
END $$;

SELECT snapshot_agent('component-template-fixer', '495 pre-image: park action=needs_review refusals as needs_human_review (bugs_open/323)');

-- 1. Two new steps.
UPDATE agent_definitions
SET default_config = jsonb_set(
      default_config,
      '{workflow,steps}',
      (default_config #> '{workflow,steps}') || $STEPS${
  "check_refused": {
    "action": "conditional",
    "config": {
      "condition": "fix_result.action == 'needs_review'",
      "then_step": "park_refused",
      "else_step": "compose_note"
    },
    "description": "bugs_open/323: the handler marks every REFUSAL (cannot do this fix_type; slot locked; subject unidentifiable) with action=needs_review and every idempotent no-op without it. A refusal is parked for a human; a no-op continues to the note exactly as before."
  },
  "park_refused": {
    "action": "fail_work_item",
    "config": {
      "work_item_id": "input_data.work_item_id",
      "status_override": "needs_human_review",
      "error_message": "component-template-fixer REFUSED this item (fix_result.action=needs_review): the fix_type is outside what this handler can do, or the target slot is locked or unidentifiable. Nothing was changed. The handler's reason is quoted in the doc_notes 'fix' entry for this site and in fix_result of the run (bugs_open/323).",
      "error_step": "compose_note"
    },
    "next_step": "compose_note",
    "description": "Park the work item visibly (needs_human_review) instead of letting the dispatch loop stamp it complete — the page-build-handler pattern. Docs must never fail the fix: a missing work_item_id falls through to the note."
  }
}$STEPS$::jsonb
    ),
    updated_at = now()
WHERE type = 'component-template-fixer' AND is_active
  AND COALESCE(is_snapshot,false) = false AND deleted_at IS NULL;

-- 2. Rewire the no-op edge through the refusal check (guarded on the pre-image value).
UPDATE agent_definitions
SET default_config = jsonb_set(
      default_config,
      '{workflow,steps,check_needs_rerender,config,else_step}',
      '"check_refused"'
    ),
    updated_at = now()
WHERE type = 'component-template-fixer' AND is_active
  AND COALESCE(is_snapshot,false) = false AND deleted_at IS NULL
  AND default_config #>> '{workflow,steps,check_needs_rerender,config,else_step}' = 'compose_note';

-- 3. The NOTES entry names a refusal as a refusal and quotes the reason.
UPDATE agent_definitions
SET default_config = jsonb_set(
      default_config,
      '{workflow,steps,compose_note,config,prompt_template}',
      to_jsonb(replace(
        default_config #>> '{workflow,steps,compose_note,config,prompt_template}',
        'If the fix result shows fixed=false, title it no-op fix pass and say nothing changed.',
        'If the fix result shows action=needs_review, title it "refused: <fix_type>", put fix_result.reason VERBATIM on the Fix line, and say the work item was parked for human review. Otherwise, if the fix result shows fixed=false, title it no-op fix pass and say nothing changed.'
      ))
    ),
    updated_at = now()
WHERE type = 'component-template-fixer' AND is_active
  AND COALESCE(is_snapshot,false) = false AND deleted_at IS NULL;

-- 4. Verify (DO/RAISE — a SELECT cannot stop the COMMIT).
DO $$
DECLARE cfg jsonb; steps jsonb; s text; nxt text;
BEGIN
  SELECT default_config INTO cfg FROM agent_definitions
   WHERE type='component-template-fixer' AND is_active
     AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
  steps := cfg #> '{workflow,steps}';

  IF cfg #>> '{workflow,steps,check_needs_rerender,config,else_step}' <> 'check_refused' THEN
    RAISE EXCEPTION '495: check_needs_rerender.else_step is % (want check_refused)', cfg #>> '{workflow,steps,check_needs_rerender,config,else_step}';
  END IF;
  IF cfg #>> '{workflow,steps,check_needs_rerender,config,then_step}' <> 'create_rerender' THEN
    RAISE EXCEPTION '495: check_needs_rerender.then_step changed unexpectedly: %', cfg #>> '{workflow,steps,check_needs_rerender,config,then_step}';
  END IF;
  IF cfg #>> '{workflow,steps,check_refused,action}' <> 'conditional'
     OR cfg #>> '{workflow,steps,check_refused,config,condition}' <> 'fix_result.action == ''needs_review'''
     OR cfg #>> '{workflow,steps,check_refused,config,then_step}' <> 'park_refused'
     OR cfg #>> '{workflow,steps,check_refused,config,else_step}' <> 'compose_note' THEN
    RAISE EXCEPTION '495: check_refused step malformed: %', cfg #> '{workflow,steps,check_refused}';
  END IF;
  IF cfg #>> '{workflow,steps,park_refused,action}' <> 'fail_work_item'
     OR cfg #>> '{workflow,steps,park_refused,config,status_override}' <> 'needs_human_review'
     OR cfg #>> '{workflow,steps,park_refused,config,work_item_id}' <> 'input_data.work_item_id'
     OR cfg #>> '{workflow,steps,park_refused,next_step}' <> 'compose_note' THEN
    RAISE EXCEPTION '495: park_refused step malformed: %', cfg #> '{workflow,steps,park_refused}';
  END IF;
  IF cfg #>> '{workflow,steps,compose_note,config,prompt_template}' NOT LIKE '%action=needs_review%' THEN
    RAISE EXCEPTION '495: compose_note prompt amendment not present';
  END IF;
  -- Every step the two new steps point at must exist (a dangling next_step is a
  -- silent workflow failure at runtime, not a migration error).
  FOREACH s IN ARRAY ARRAY['park_refused','compose_note'] LOOP
    IF NOT (steps ? s) THEN RAISE EXCEPTION '495: check_refused/park_refused points at missing step %', s; END IF;
  END LOOP;
  RAISE NOTICE '495 verified: refusals (fix_result.action=needs_review) now park at needs_human_review';
END $$;

INSERT INTO doc_notes (subject_type, subject_key, body, categories, source, created_by)
VALUES ('pipeline', 'build',
  '495 (bugs_open/323): component-template-fixer now honours its own refusal flag. Any fix_result carrying action=needs_review (13 refusal arms in fix_component_template_action.go — the cta_improvement/nav_restructure punt, unrecognised fix_type, locked chrome slot, chrome_overflow_fix without a usable slot/selector, repair_page_component_status refusals) parks the work item at needs_human_review via fail_work_item instead of completing it. Idempotent no-ops (fixed=false, no action key) are unchanged. Measured 2026-08-19: 470 historical refusals, 0 legitimate completions affected. The reason is quoted in the fixer''s doc_notes fix entry and in the item''s error column.',
  '["fix","migration","bugs_open/323","component-template-fixer"]'::jsonb,
  'migration', '495_fixer_parks_refusals_as_needs_human_review.sql');

COMMIT;
