-- 569 ROLLBACK — put the `rewrite_negations` repair ceiling back to 2000.
--
-- ⚠ THIS RE-ARMS A MEASURED, TOTAL FAILURE ON DENSE PAGES. At 2000, every page
-- observed with >=9 repair targets returned `repair_unavailable` and repaired
-- NOTHING — 24.4% of all repair targets in the 2026-08-23 window. It fails
-- loudly (that is 517's design and it is sound), but it fails completely.
--
-- Reasons this rollback might legitimately be right:
--   * spend. If the dense-page repairs turn out to cost more than they are
--     worth, the honest fix is a LOWER ceiling chosen from measurement, not a
--     return to a number that predates any.
--   * a model change. If the step's model moves to one whose thinking behaviour
--     makes 16000 dangerous, re-measure before picking the replacement.
--
-- Before running this, read bugs_open/305 §27 — the ceiling is a stopgap for an
-- unbounded target count, and chunking is the structural answer.
--
-- Needle-gated on 16000 so it cannot clobber a later deliberate retune.

BEGIN;

SELECT snapshot_agent('page-content-writer',
                      '569_ROLLBACK: pre-update');

UPDATE agent_definitions
   SET default_config = jsonb_set(default_config,
         '{workflow,steps,process_sections_loop,config,sub_workflow,steps,rewrite_negations,config,ai_service,max_tokens}',
         '2000'::jsonb),
       updated_at = now()
 WHERE type = 'page-content-writer' AND is_active
   AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL
   AND default_config #>> '{workflow,steps,process_sections_loop,config,sub_workflow,steps,rewrite_negations,action}' = 'rewrite_negations'
   AND default_config #>> '{workflow,steps,process_sections_loop,config,sub_workflow,steps,rewrite_negations,config,ai_service,max_tokens}' = '16000';

DO $$
DECLARE n int;
BEGIN
  SELECT count(*) INTO n FROM agent_definitions
   WHERE type = 'page-content-writer' AND is_active
     AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL
     AND default_config #>> '{workflow,steps,process_sections_loop,config,sub_workflow,steps,rewrite_negations,config,ai_service,max_tokens}' = '2000';
  IF n <> 1 THEN
    RAISE EXCEPTION '569 ROLLBACK FAILED: expected exactly 1 page-content-writer back at max_tokens=2000, got %', n;
  END IF;
  RAISE NOTICE '569 ROLLBACK OK: ceiling back at 2000 — dense pages will return repair_unavailable again';
END $$;

COMMIT;
