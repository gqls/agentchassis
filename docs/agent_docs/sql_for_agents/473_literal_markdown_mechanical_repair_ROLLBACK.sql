-- 473 ROLLBACK — surgical inverse: remove exactly what 473 added and nothing
-- else, so any migration that touched page-rerender after 473 survives intact
-- (a whole-config restore from a backup would clobber it — that is why this is
-- NOT a restore-from-snapshot).

BEGIN;

SELECT snapshot_agent('page-rerender', '473_ROLLBACK: pre-revert');

UPDATE agent_definitions
   SET default_config = jsonb_set(
         default_config #- '{workflow,steps,rerender_sections,config,strip_literal_markdown}',
         '{workflow,steps,check_rerender_mode,config,condition}',
         to_jsonb(replace(
           default_config #>> '{workflow,steps,check_rerender_mode,config,condition}',
           ' OR input_data.spec.reason == ''literal_markdown''',
           ''
         ))
       ),
       updated_at = now()
 WHERE type = 'page-rerender' AND is_active
   AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL
   AND default_config #>> '{workflow,steps,check_rerender_mode,config,condition}'
       LIKE '%literal_markdown%';

DO $$
DECLARE cond text;
BEGIN
  SELECT default_config #>> '{workflow,steps,check_rerender_mode,config,condition}'
    INTO cond FROM agent_definitions
   WHERE type = 'page-rerender' AND is_active
     AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;
  IF cond LIKE '%literal_markdown%' THEN
    RAISE EXCEPTION '473 ROLLBACK FAILED: condition still carries literal_markdown: %', cond;
  END IF;
  IF EXISTS (
    SELECT 1 FROM agent_definitions
     WHERE type = 'page-rerender' AND is_active
       AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL
       AND default_config #> '{workflow,steps,rerender_sections,config,strip_literal_markdown}' IS NOT NULL
  ) THEN
    RAISE EXCEPTION '473 ROLLBACK FAILED: strip flag still present';
  END IF;
  RAISE NOTICE '473 ROLLBACK OK: condition = %', cond;
END $$;

COMMIT;
