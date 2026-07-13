-- NNN_rename_complete_keys_preferred.sql
--
-- ⚠ ORDERING: APPLY ONLY AFTER the Option A image (shared result contract:
-- datahelpers/result_contract.go + unified extractFinalResult + response size
-- guard) is DEPLOYED. Before that image, `result_from` is a key the response
-- builder does not read — applying this early resurrects the 1.27MB dump
-- fallback and the Kafka rejection (run 17933a83).
--
-- WHAT: move the four diagnose/index agents' complete steps to the PREFERRED
-- key (guidelines 003: "Use the preferred names"), all as result_from
-- (flatten). SHAPE CHANGE, deliberate: responses become the named object
-- DIRECTLY (e.g. the diagnosis map) instead of nested {key: {...}} — flagged.
-- Also fixes index-orchestrator's own imagined key: its result_from pointed at
-- "code-indexer_result", but the engine stores the call step's response under
-- the STEP NAME "call_indexer" (same class as the diagnose parent bug).
-- After this + the deployed image, resolveResultSpec's deprecation warns for
-- these agents go quiet.
--
-- Census (optional, before/after): find every other agent still on deprecated
-- keys:
--   SELECT type, default_config #> '{workflow,steps,complete,config}' AS cc
--   FROM agent_definitions
--   WHERE COALESCE(is_snapshot,false)=false AND deleted_at IS NULL
--     AND (default_config #> '{workflow,steps,complete,config}') ?| array['output_field','output_fields','output'];

BEGIN;

SELECT snapshot_agent('diagnose-agent',      'complete -> preferred result_from "diagnosis" (flatten; requires Option A image)');
UPDATE agent_definitions SET default_config = jsonb_set(default_config,
  '{workflow,steps,complete,config}', '{"result_from": "diagnosis"}'::jsonb), updated_at = now()
WHERE type='diagnose-agent' AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

SELECT snapshot_agent('diagnose-orchestrator','complete -> preferred result_from "call_diagnoser" (flatten; requires Option A image)');
UPDATE agent_definitions SET default_config = jsonb_set(default_config,
  '{workflow,steps,complete,config}', '{"result_from": "call_diagnoser"}'::jsonb), updated_at = now()
WHERE type='diagnose-orchestrator' AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

SELECT snapshot_agent('code-indexer',        'complete -> preferred result_from "index_result" (flatten; requires Option A image)');
UPDATE agent_definitions SET default_config = jsonb_set(default_config,
  '{workflow,steps,complete,config}', '{"result_from": "index_result"}'::jsonb), updated_at = now()
WHERE type='code-indexer' AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

SELECT snapshot_agent('index-orchestrator',  'complete result_from: "code-indexer_result" (imagined; stored under STEP NAME) -> "call_indexer"');
UPDATE agent_definitions SET default_config = jsonb_set(default_config,
  '{workflow,steps,complete,config}', '{"result_from": "call_indexer"}'::jsonb), updated_at = now()
WHERE type='index-orchestrator' AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

SELECT type, default_config #> '{workflow,steps,complete,config}' AS complete_config
FROM agent_definitions
WHERE type IN ('diagnose-agent','diagnose-orchestrator','code-indexer','index-orchestrator')
  AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL
ORDER BY type;

COMMIT;

-- ── REVERT (restores today's applied configs) ───────────────────────────────
-- BEGIN;
-- SELECT snapshot_agent('diagnose-agent','revert to output_fields');
-- UPDATE agent_definitions SET default_config = jsonb_set(default_config,'{workflow,steps,complete,config}','{"output_fields": ["diagnosis"]}'::jsonb), updated_at=now() WHERE type='diagnose-agent' AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
-- SELECT snapshot_agent('diagnose-orchestrator','revert to output_fields');
-- UPDATE agent_definitions SET default_config = jsonb_set(default_config,'{workflow,steps,complete,config}','{"output_fields": ["call_diagnoser"]}'::jsonb), updated_at=now() WHERE type='diagnose-orchestrator' AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
-- SELECT snapshot_agent('code-indexer','revert to output_fields');
-- UPDATE agent_definitions SET default_config = jsonb_set(default_config,'{workflow,steps,complete,config}','{"output_fields": ["index_result"]}'::jsonb), updated_at=now() WHERE type='code-indexer' AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
-- SELECT snapshot_agent('index-orchestrator','revert to result_from code-indexer_result');
-- UPDATE agent_definitions SET default_config = jsonb_set(default_config,'{workflow,steps,complete,config}','{"result_from": "code-indexer_result"}'::jsonb), updated_at=now() WHERE type='index-orchestrator' AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
-- COMMIT;
