-- NNN_fix_orchestrator_complete_key.sql
--
-- Run 73ed55c6 follow-up #1: the parent's complete had output_fields
-- ["diagnose-agent_result"], but jsonb_object_keys on the parent row shows the
-- child's response is stored under the STEP NAME — `call_diagnoser` (the
-- engine stores each step's result under its step name, plus output_field when
-- configured; the call step configures none). `diagnose-agent_result` never
-- existed — the original result_from config was written against an imagined
-- key, and my 2026-07-03 migration repeated the imagined name. With the key
-- absent, extractFinalResult's output_fields branch found nothing and fell to
-- the dump fallback (benign at 14KB, but the same dead-key class).
--
-- FIX: point at the real key. Plural output_fields deliberately — it is the
-- ONLY spelling both readers act on (extractFinalResult current;
-- resolveResultSpec deprecated-alias), per the two-readers analysis; preferred
-- names stay vetoed until the Go unification. Response becomes
-- {"call_diagnoser": <child result>} (~7–10KB).
-- Optional later refinement: unwrap to the diagnosis with a dotted
-- output_field once the stored shape is confirmed
-- (SELECT jsonb_object_keys(collected_data->'call_diagnoser') …).
-- Verification on the NEXT diagnose run:
--   SELECT status, pg_column_size(collected_data),
--          (collected_data ? 'call_diagnoser') AS has_child_result
--   FROM orchestration_states WHERE correlation_id='<new>'::uuid;

BEGIN;

SELECT snapshot_agent('diagnose-orchestrator',
  'complete output_fields: ["diagnose-agent_result"] (imagined key; response stored under STEP NAME) -> ["call_diagnoser"]');

UPDATE agent_definitions
SET default_config = jsonb_set(
      default_config,
      '{workflow,steps,complete,config}',
      '{"output_fields": ["call_diagnoser"]}'::jsonb),
    updated_at = now()
WHERE type = 'diagnose-orchestrator'
  AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

SELECT type, default_config #> '{workflow,steps,complete,config}' AS complete_config
FROM agent_definitions
WHERE type = 'diagnose-orchestrator'
  AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

COMMIT;

-- ── REVERT ──────────────────────────────────────────────────────────────────
-- BEGIN;
-- SELECT snapshot_agent('diagnose-orchestrator','revert complete to diagnose-agent_result');
-- UPDATE agent_definitions SET default_config = jsonb_set(default_config,
--   '{workflow,steps,complete,config}', '{"output_fields": ["diagnose-agent_result"]}'::jsonb),
--   updated_at = now()
-- WHERE type='diagnose-orchestrator' AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
-- COMMIT;
