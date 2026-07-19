-- 171_experience_council_abstention_tolerant.sql — Experience Loop, CP2 unblock.
--
-- PROBLEM (observed, run 054b358a, 2026-07-18): every critic step carried
-- `error_step: complete_refused`. `review_feasibility` died on
-- `no text content in response (had 1 blocks)` — bugs_open/008 item 5, an
-- undecoded `stop_reason` (likely `refusal`) on Sonnet 5 — and took the WHOLE
-- run down after round 1. That run had just confirmed both of the major fixes
-- (compose truncation + load_context level filter) were working, and lost the
-- round anyway to one flaky seat.
--
-- FIX: the three ADVISORY critics fall through to the NEXT critic on error
-- instead of to complete_refused. `diagnose_council_decide` already reads an
-- absent review field as an ABSTENTION
-- (diagnose_council_decide_action.go:98-112 — written for the relevance filter
-- that skips irrelevant seats), so a dead critic simply does not vote and the
-- round still decides. It fails closed if ALL seats abstain (:141), so this
-- cannot degrade into "silence == approval".
--
-- NOT a blanket change. `review_honesty` KEEPS `error_step: complete_refused`
-- because it is the sole `hard_veto_from` seat: an abstaining honesty auditor
-- would let a plan reach "approved" with the anti-fabrication gate never
-- applied — the exact failure the loop exists to catch. A dead honesty auditor
-- must refuse the run.
--
-- Note the fall-through is to the NEXT critic, not straight to council_decide:
-- routing to council_decide would drop every critic AFTER the failed one too,
-- turning one flaky seat into three abstentions.
--
-- Config-only: live the moment it commits, no image roll. Seed 167 carries the
-- same change in-place, so a re-apply of 167 does not clobber this (the
-- patch-style re-seed landmine).

BEGIN;

SELECT snapshot_agent('experience-planner', 'pre-update: 171 abstention-tolerant advisory critics')
WHERE EXISTS (SELECT 1 FROM agent_definitions
              WHERE type='experience-planner' AND is_active
                AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL);

UPDATE agent_definitions
   SET default_config = jsonb_set(
         jsonb_set(
           jsonb_set(
             default_config,
             '{workflow,steps,review_journeys,config,error_step}',
             '"review_feasibility"'::jsonb, true),
           '{workflow,steps,review_feasibility,config,error_step}',
           '"review_honesty"'::jsonb, true),
         '{workflow,steps,review_mvp,config,error_step}',
         '"council_decide"'::jsonb, true),
       updated_at = now()
 WHERE type = 'experience-planner'
   AND COALESCE(is_snapshot,false) = false
   AND deleted_at IS NULL;

-- Assert the intended end state, including the deliberate asymmetry.
DO $$
DECLARE
  j text; f text; h text; m text;
BEGIN
  SELECT default_config->'workflow'->'steps'->'review_journeys'->'config'->>'error_step',
         default_config->'workflow'->'steps'->'review_feasibility'->'config'->>'error_step',
         default_config->'workflow'->'steps'->'review_honesty'->'config'->>'error_step',
         default_config->'workflow'->'steps'->'review_mvp'->'config'->>'error_step'
    INTO j, f, h, m
    FROM agent_definitions
   WHERE type='experience-planner'
     AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

  IF j <> 'review_feasibility' OR f <> 'review_honesty' OR m <> 'council_decide' THEN
    RAISE EXCEPTION 'advisory critic error_step routing wrong: journeys=%, feasibility=%, mvp=%', j, f, m;
  END IF;
  IF h <> 'complete_refused' THEN
    RAISE EXCEPTION 'honesty (hard_veto seat) must keep error_step=complete_refused, got %', h;
  END IF;
END $$;

COMMIT;

-- Rollback: set all four back to 'complete_refused' (restores the run-ending
-- behaviour, and with it the single-flaky-critic failure mode).
