-- 208_experience_reviewer_max_tokens.sql
--
-- Live failure 2026-07-25 (corr fcdf8e72, vonc-spark-game re-plan, round 4):
-- review_feasibility died with stop_reason=max_tokens (cap=8000, only 1649
-- chars recovered) — its own compact-JSON verdict was cut before it could
-- finish, on a round where the plan and the accumulated review context had
-- grown (three prior REVISE rounds' objections + the 207 liveness-evidence
-- paragraph feed into what the reviewer reads and reasons over). Same
-- mechanism class as bugs_closed/067 (a step re-emitting/reasoning over a
-- large artifact needs headroom the cap didn't give it) — but there the fix
-- was scoped to compose/reframe/repropose only, with reviewer seats left at
-- 8000 on the explicit (untested) assumption that "verdict JSON only" stays
-- small. 067's own addendum already flagged the risk of closing one step too
-- narrow. This is that risk landing for real, on a different agent
-- (experience-planner, not feature-designer) but the identical shape.
--
-- All FIVE reviewer seats share the same structure (extended-thinking model,
-- capped-compact JSON output) and the same 8000 cap, so any of them can hit
-- this on a large-enough round — not just feasibility. Raise all five
-- together (8000 -> 16000): enough headroom for the reasoning tokens that ate
-- the failed call's budget while keeping reviewers well below the 32000 given
-- to the whole-plan re-emitters (compose/reframe/recompose), since a reviewer
-- only ever emits a small, explicitly length-capped JSON object.
--
-- Workstream: gauntlet_dead_cta (experience re-plan re-fire). Config-only,
-- live immediately. ROLLBACK: five snapshots below (one per seat).

BEGIN;

SELECT snapshot_agent('experience-planner', '208_experience_reviewer_max_tokens: pre-update');

UPDATE agent_definitions
   SET default_config = jsonb_set(default_config, '{workflow,steps,review_journeys,config,ai_service,max_tokens}', '16000', true)
 WHERE type='experience-planner' AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

UPDATE agent_definitions
   SET default_config = jsonb_set(default_config, '{workflow,steps,review_feasibility,config,ai_service,max_tokens}', '16000', true)
 WHERE type='experience-planner' AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

UPDATE agent_definitions
   SET default_config = jsonb_set(default_config, '{workflow,steps,review_honesty,config,ai_service,max_tokens}', '16000', true)
 WHERE type='experience-planner' AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

UPDATE agent_definitions
   SET default_config = jsonb_set(default_config, '{workflow,steps,review_mvp,config,ai_service,max_tokens}', '16000', true)
 WHERE type='experience-planner' AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

UPDATE agent_definitions
   SET default_config = jsonb_set(default_config, '{workflow,steps,review_contracts,config,ai_service,max_tokens}', '16000', true)
 WHERE type='experience-planner' AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

DO $$
DECLARE
  mt_j text; mt_f text; mt_h text; mt_m text; mt_c text;
BEGIN
  SELECT default_config->'workflow'->'steps'->'review_journeys'->'config'->'ai_service'->>'max_tokens',
         default_config->'workflow'->'steps'->'review_feasibility'->'config'->'ai_service'->>'max_tokens',
         default_config->'workflow'->'steps'->'review_honesty'->'config'->'ai_service'->>'max_tokens',
         default_config->'workflow'->'steps'->'review_mvp'->'config'->'ai_service'->>'max_tokens',
         default_config->'workflow'->'steps'->'review_contracts'->'config'->'ai_service'->>'max_tokens'
    INTO mt_j, mt_f, mt_h, mt_m, mt_c
    FROM agent_definitions
   WHERE type='experience-planner' AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

  IF mt_j <> '16000' THEN RAISE EXCEPTION '208: review_journeys max_tokens did not land (got %)', mt_j; END IF;
  IF mt_f <> '16000' THEN RAISE EXCEPTION '208: review_feasibility max_tokens did not land (got %)', mt_f; END IF;
  IF mt_h <> '16000' THEN RAISE EXCEPTION '208: review_honesty max_tokens did not land (got %)', mt_h; END IF;
  IF mt_m <> '16000' THEN RAISE EXCEPTION '208: review_mvp max_tokens did not land (got %)', mt_m; END IF;
  IF mt_c <> '16000' THEN RAISE EXCEPTION '208: review_contracts max_tokens did not land (got %)', mt_c; END IF;
END $$;

COMMIT;
