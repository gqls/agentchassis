-- 201_designer_repropose_cap.sql
--
-- B4 designer finding (2026-07-24, corr 7773219b; also explains the 07-23
-- corr c2a9fd27 death): feature-designer's `repropose` step caps max_tokens at
-- 16000 while a full staged plan runs ~26k chars — so EVERY revise cycle dies
-- at repropose (stop_reason=max_tokens; the 07-23 instance hung then FAILED
-- with no recorded error on the older image). The designer has therefore never
-- completed a revise round; only round-1 approvals ever land. Raise to 32000,
-- matching compose. Config-only. ROLLBACK: snapshot below.

BEGIN;
SELECT snapshot_agent('feature-designer', '201_designer_repropose_cap: pre-update');

UPDATE agent_definitions
   SET default_config = jsonb_set(
         default_config,
         '{workflow,steps,repropose,config,ai_service,max_tokens}',
         '32000'::jsonb, false)
 WHERE type='feature-designer'
   AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

DO $$
DECLARE v int;
BEGIN
  SELECT (default_config->'workflow'->'steps'->'repropose'->'config'->'ai_service'->>'max_tokens')::int
    INTO v FROM agent_definitions
   WHERE type='feature-designer' AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
  IF v IS DISTINCT FROM 32000 THEN
    RAISE EXCEPTION '201: repropose max_tokens is %, expected 32000', v;
  END IF;
END $$;
COMMIT;
