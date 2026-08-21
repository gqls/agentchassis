-- ROLLBACK for 517 — remove the ai_service block from the rewrite_negations step.
--
-- ⚠ This returns the repair to INERT (it will report
-- status=repair_unavailable, error="no ai_service configuration resolvable" and
-- splice nothing) rather than to any earlier working state. Rolling this back is
-- only sensible as a way of disabling the repair while keeping its detection —
-- if that is what you want, this is a legitimate use; if you want the step gone,
-- roll back 509 instead.

BEGIN;

SELECT snapshot_agent('page-content-writer', '517_ROLLBACK: pre-rollback');

UPDATE agent_definitions
   SET default_config = default_config #- '{workflow,steps,process_sections_loop,config,sub_workflow,steps,rewrite_negations,config,ai_service}',
       updated_at = now()
 WHERE type = 'page-content-writer' AND is_active
   AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

DO $$
DECLARE n int;
BEGIN
  SELECT count(*) INTO n FROM agent_definitions
   WHERE type = 'page-content-writer' AND is_active
     AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL
     AND (default_config #> '{workflow,steps,process_sections_loop,config,sub_workflow,steps,rewrite_negations,config,ai_service}') IS NULL;
  IF n <> 1 THEN
    RAISE EXCEPTION '517 ROLLBACK FAILED: the ai_service block is still present (n=%)', n;
  END IF;
  RAISE NOTICE '517 ROLLBACK OK — the repair is INERT again by design';
END $$;

COMMIT;
