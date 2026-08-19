-- 486 ROLLBACK — remove the judged branch and restore the pre-486 wiring.
-- The Go fix_type stays in the binary (inert without these steps).
BEGIN;
-- Same double-apply guard, inverted: rolling back an unapplied 486 would snapshot
-- and "restore" a row that never changed.
DO $$
DECLARE cfg jsonb;
BEGIN
  SELECT default_config INTO cfg FROM agent_definitions
   WHERE type='component-template-fixer' AND is_active
     AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
  IF cfg IS NULL OR NOT (cfg #> '{workflow,steps}') ? 'check_scope_route' THEN
    RAISE EXCEPTION '486 rollback: judged steps not present — 486 is not applied; nothing to roll back.';
  END IF;
END $$;
SELECT snapshot_agent('component-template-fixer', '486 rollback pre-image');
UPDATE agent_definitions
SET default_config = jsonb_set(
      jsonb_set(
        jsonb_set(default_config, '{workflow,steps}',
          (default_config #> '{workflow,steps}')
            - 'check_scope_route' - 'scope_script_llm' - 'apply_judged_write'
            - 'check_judged_result' - 'judged_refusal' - 'create_section_edit_delivery'),
        '{workflow,steps,apply_fix,next_step}', '"check_needs_rerender"'),
      '{workflow,steps,create_rerender,next_step}', '"compose_note"'),
    updated_at = now()
WHERE type = 'component-template-fixer' AND is_active
  AND COALESCE(is_snapshot,false) = false AND deleted_at IS NULL
  AND (default_config #> '{workflow,steps}') ? 'check_scope_route';
DO $$
DECLARE cfg jsonb;
BEGIN
  SELECT default_config INTO cfg FROM agent_definitions
   WHERE type='component-template-fixer' AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
  IF (cfg #> '{workflow,steps}') ? 'check_scope_route' THEN
    RAISE EXCEPTION '486 rollback: judged steps still present';
  END IF;
  IF cfg #>> '{workflow,steps,apply_fix,next_step}' <> 'check_needs_rerender' THEN
    RAISE EXCEPTION '486 rollback: apply_fix wiring not restored';
  END IF;
END $$;
COMMIT;
