-- 495 ROLLBACK — restore the pre-495 component-template-fixer workflow:
-- drop check_refused/park_refused, re-point check_needs_rerender.else_step at
-- compose_note, restore the compose_note prompt sentence. Hand-run only.
-- Prefer restoring the snapshot_agent pre-image if it is the newest snapshot for
-- this agent; this file is the surgical alternative when later migrations
-- (e.g. 486) have landed on top and the snapshot would undo them too.
BEGIN;

DO $$
DECLARE cfg jsonb;
BEGIN
  SELECT default_config INTO cfg FROM agent_definitions
   WHERE type='component-template-fixer' AND is_active
     AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
  IF NOT ((cfg #> '{workflow,steps}') ? 'check_refused') THEN
    RAISE EXCEPTION '495 ROLLBACK: check_refused not present — 495 is not applied';
  END IF;
END $$;

SELECT snapshot_agent('component-template-fixer', '495 rollback pre-image');

UPDATE agent_definitions
SET default_config = jsonb_set(
      jsonb_set(
        default_config,
        '{workflow,steps}',
        (default_config #> '{workflow,steps}') - 'check_refused' - 'park_refused'
      ),
      '{workflow,steps,check_needs_rerender,config,else_step}',
      '"compose_note"'
    ),
    updated_at = now()
WHERE type='component-template-fixer' AND is_active
  AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL
  AND default_config #>> '{workflow,steps,check_needs_rerender,config,else_step}' = 'check_refused';

UPDATE agent_definitions
SET default_config = jsonb_set(
      default_config,
      '{workflow,steps,compose_note,config,prompt_template}',
      to_jsonb(replace(
        default_config #>> '{workflow,steps,compose_note,config,prompt_template}',
        'If the fix result shows action=needs_review, title it "refused: <fix_type>", put fix_result.reason VERBATIM on the Fix line, and say the work item was parked for human review. Otherwise, if the fix result shows fixed=false, title it no-op fix pass and say nothing changed.',
        'If the fix result shows fixed=false, title it no-op fix pass and say nothing changed.'
      ))
    ),
    updated_at = now()
WHERE type='component-template-fixer' AND is_active
  AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

DO $$
DECLARE cfg jsonb;
BEGIN
  SELECT default_config INTO cfg FROM agent_definitions
   WHERE type='component-template-fixer' AND is_active
     AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
  IF (cfg #> '{workflow,steps}') ? 'check_refused' OR (cfg #> '{workflow,steps}') ? 'park_refused'
     OR cfg #>> '{workflow,steps,check_needs_rerender,config,else_step}' <> 'compose_note'
     OR cfg #>> '{workflow,steps,compose_note,config,prompt_template}' LIKE '%action=needs_review%' THEN
    RAISE EXCEPTION '495 ROLLBACK: post-image check failed';
  END IF;
END $$;

COMMIT;
