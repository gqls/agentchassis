-- 202_designer_compose_caps.sql
--
-- bugs_closed/067 sibling, found one run later (corr ffb74056, 2026-07-24): the
-- designer's INITIAL `design` step (and its `reframe` twin) carry the same
-- undersized max_tokens=16000 that 201 fixed on `repropose`. A grown capability
-- spec pushed the round-5 plan past the cap and the run died at `design` with
-- stop_reason=max_tokens. Whole-artifact emitters all get 32000; reviewer seats
-- (verdict JSON) stay at 8000 deliberately.
-- ROLLBACK: snapshot below.

BEGIN;
SELECT snapshot_agent('feature-designer', '202_designer_compose_caps: pre-update');

UPDATE agent_definitions
   SET default_config = jsonb_set(
         jsonb_set(
           default_config,
           '{workflow,steps,design,config,ai_service,max_tokens}', '32000'::jsonb, false),
         '{workflow,steps,reframe,config,ai_service,max_tokens}', '32000'::jsonb, false)
 WHERE type='feature-designer'
   AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

DO $$
DECLARE d int; r int; rp int;
BEGIN
  SELECT (default_config->'workflow'->'steps'->'design'->'config'->'ai_service'->>'max_tokens')::int,
         (default_config->'workflow'->'steps'->'reframe'->'config'->'ai_service'->>'max_tokens')::int,
         (default_config->'workflow'->'steps'->'repropose'->'config'->'ai_service'->>'max_tokens')::int
    INTO d, r, rp FROM agent_definitions
   WHERE type='feature-designer' AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
  IF d IS DISTINCT FROM 32000 OR r IS DISTINCT FROM 32000 OR rp IS DISTINCT FROM 32000 THEN
    RAISE EXCEPTION '202: caps are design=%, reframe=%, repropose=% — expected 32000 each', d, r, rp;
  END IF;
END $$;
COMMIT;
