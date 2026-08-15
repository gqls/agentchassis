-- ROLLBACK for 417 — restore asset_id? (optional) on image-build-handler's
-- call_asset_deployer input_mapping, undoing the strict `!` marker.
--
-- Safe against ANY binary: the `?` spelling is understood by every chassis
-- build since migration 401. Use this if the strict flip turns out to fire on a
-- live refusal branch the 2026-08-15 measurement did not see (13/13 spawns
-- carried asset_id; zero refusal spawns in the retained window) and the owner
-- prefers the old silent-skip while that branch is redesigned.

BEGIN;

SELECT snapshot_agent('image-build-handler',
    '417_..._ROLLBACK.sql: pre-rollback');

UPDATE agent_definitions
   SET default_config = jsonb_set(
         default_config #- '{workflow,steps,call_asset_deployer,config,input_mapping,asset_id!}',
         '{workflow,steps,call_asset_deployer,config,input_mapping,asset_id?}',
         COALESCE(
           default_config #> '{workflow,steps,call_asset_deployer,config,input_mapping,asset_id!}',
           '"asset_stored.asset_id"'::jsonb
         ),
         true
       ),
       updated_at = now()
 WHERE type = 'image-build-handler'
   AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL
   AND default_config #> '{workflow,steps,call_asset_deployer,config,input_mapping}' ? 'asset_id!';

DO $$
DECLARE mapping jsonb;
BEGIN
    SELECT default_config #> '{workflow,steps,call_asset_deployer,config,input_mapping}' INTO mapping
    FROM agent_definitions
    WHERE type = 'image-build-handler'
      AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;
    IF mapping IS NULL OR NOT (mapping ? 'asset_id?') OR (mapping ? 'asset_id!') THEN
        RAISE EXCEPTION '417 rollback: expected asset_id? present and asset_id! absent; mapping is %', mapping;
    END IF;
END $$;

COMMIT;
