-- 553: asset-deployer's input_contract catches up with what the action does —
-- `s3_uri` stops being required, `asset_id` becomes a named optional.
--
-- WHY. Since migration 324 (input_fields, applied 2026-08-06) and commit
-- 91dda3243 (209 Phase 2, live v1.0.1276), `deploy_image_asset` resolves its
-- source EITHER from an explicit `s3_uri` OR from `asset_id` via the asset
-- row's own storage_path/url (deploy_image_asset_action.go:323-330), and skips
-- LOUDLY when it has neither. But the agent's input_contract still says
--   required: ["domain","s3_uri"]
-- and `ValidateInputContract` (input_mapping.go:301-329) hard-fails a call_agent
-- dispatch missing any required field — so an asset_id-only deploy, the exact
-- shape bugs_open/155's closure recipe demands and the shape migrations
-- 348/401/402 built for their callers, is REFUSED at the contract before the
-- action can run. The contract is the last place in the platform still
-- asserting the pre-155 world where s3_uri was the only way to name a source.
--
-- WHAT CHANGES. required: ["domain"]. s3_uri moves to optional; asset_id is
-- added to optional. The description states the either/or rule so a reader of
-- the contract learns the real requirement (one of s3_uri | asset_id), which
-- a flat required-list cannot express — the action's own loud skip is the
-- runtime enforcement of that disjunction, and `ValidateInputContract` treats
-- optional as documentation only, so nothing new can be silently dropped.
--
-- BLAST RADIUS, measured before writing (2026-08-22): the contract is enforced
-- only on the call_agent input_mapping path with a LITERAL agent_type
-- (call_agent.go:1004-1015), and exactly ONE live step is that shape —
-- image-build-handler's call_asset_deployer, which maps BOTH s3_uri and
-- asset_id (401/417). build-dispatch-loop's call_handler resolves its
-- agent_type dynamically, so it never reaches this validation. So no existing
-- caller's behaviour changes — this only stops the contract refusing a caller
-- shape the action itself supports. Config change, live on apply, no image
-- needed (the action code has been live since v1.0.1276).
--
-- IDEMPOTENT: keyed on the current contract shape; a re-run touches 0 rows.

BEGIN;

-- Backup the exact row being changed.
CREATE TABLE IF NOT EXISTS bak_asset_deployer_contract_20260822 AS
SELECT * FROM agent_definitions WHERE false;
INSERT INTO bak_asset_deployer_contract_20260822
SELECT * FROM agent_definitions
WHERE type='asset-deployer' AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

-- Pre-state guard: refuse if the contract is not the shape this file was
-- written against (a concurrent edit must abort, not be overwritten).
DO $$
DECLARE c jsonb;
BEGIN
  SELECT input_contract INTO c FROM agent_definitions
   WHERE type='asset-deployer' AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
  IF c IS NULL THEN
    RAISE EXCEPTION 'REFUSED: no input_contract on the live asset-deployer row — restructured since this file was written';
  END IF;
  IF c->'required' IS DISTINCT FROM '["domain","s3_uri"]'::jsonb THEN
    RAISE EXCEPTION 'REFUSED: input_contract.required is % — not the ["domain","s3_uri"] this file expects; re-read before applying', c->'required';
  END IF;
END $$;

UPDATE agent_definitions
SET input_contract = jsonb_build_object(
      'required', '["domain"]'::jsonb,
      'optional', '["s3_uri","asset_id","purpose","asset_key"]'::jsonb,
      'description',
        'Provide domain plus ONE of s3_uri (explicit source object) or asset_id '
        || '(the asset row''s own storage_path/url is used — the post-155 identity path). '
        || 'Neither present => the action skips loudly ("no storage URI found"); the '
        || 'disjunction is enforced at the action, not expressible in a flat required list. '
        || 'Optional purpose (default: hero) controls resize dimensions; optional asset_key '
        || 'fixes the filename when it differs from purpose. The deploy PATH is derived from '
        || '(asset_key, purpose) and cannot be chosen — an explicit deploy_path is refused '
        || '(bugs_open/179 finding A).'
    ),
    updated_at = NOW()
WHERE type='asset-deployer' AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL
  AND input_contract->'required' = '["domain","s3_uri"]'::jsonb;

-- Verify loudly (SELECTs cannot stop a COMMIT — DO/RAISE).
DO $$
DECLARE c jsonb;
BEGIN
  SELECT input_contract INTO c FROM agent_definitions
   WHERE type='asset-deployer' AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
  IF c->'required' IS DISTINCT FROM '["domain"]'::jsonb THEN
    RAISE EXCEPTION 'required not relaxed: %', c->'required';
  END IF;
  IF NOT (c->'optional' ? 's3_uri' AND c->'optional' ? 'asset_id'
          AND c->'optional' ? 'purpose' AND c->'optional' ? 'asset_key') THEN
    RAISE EXCEPTION 'optional list wrong: %', c->'optional';
  END IF;
  RAISE NOTICE 'OK: input_contract = %', c;
END $$;

COMMIT;

-- ROLLBACK, if ever wanted (restores the exact pre-change row):
--   UPDATE agent_definitions a SET input_contract = b.input_contract
--     FROM bak_asset_deployer_contract_20260822 b WHERE a.id = b.id;
