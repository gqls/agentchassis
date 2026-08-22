-- 554: asset-retraction — remove the baked "dry_run": false an operator edit
-- left in the LIVE step config, restoring seed 446's reviewed, safe shape.
--
-- WHAT HAPPENED. Seed 446 (recorded 2026-08-20) ships the retract step with NO
-- dry_run key — the action's own default is dry-run TRUE, and its header rules
-- the unsafe side defaults OFF (owner ruling 2026-08-02 §2, cited in
-- retract_asset_files_action.go). The staged gaswholesalers dispatch script
-- noted that step_overrides propagation was UNVERIFIED and named its fallback:
-- "a one-off agent_definitions config UPDATE (snapshot first)". A snapshot was
-- taken 2026-08-20 17:31 and the live row now carries "dry_run": false — armed
-- — while the step's own description STILL SAYS "dry run unless the step
-- config carries dry_run:false". The arming was never reverted; the live row's
-- last write is 2026-08-22 08:36.
--
-- WHY IT MATTERS. Every dispatcher that believes the description (and the
-- prepared script, which prints "Dry run (default)") DELETES on what it
-- believes is an audit run. Measured consequence 2026-08-22: ten dispatches
-- intended as dry runs each deleted their target file. The outcome happened to
-- be the owner-authorised end state, reached through a pre-dispatch
-- estate-wide reference audit and the action's five-guard chain — but the
-- process safety the dry-run default exists to provide was ABSENT, and the
-- next operator gets no audit pass at all. LANDMINES.md carries the entry.
--
-- WHAT THIS DOES. Deletes the dry_run key from the live retract step config.
-- The action's compiled default (dry-run TRUE) then governs again; an armed
-- run once more requires the explicit literal, per the reviewed design.
-- All 13 stale-logo retractions this key enabled are COMPLETE, so nothing
-- in flight depends on the armed state.
--
-- IDEMPOTENT: keyed on the key's presence; a re-run touches 0 rows.

BEGIN;

CREATE TABLE IF NOT EXISTS bak_asset_retraction_20260822 AS
SELECT * FROM agent_definitions WHERE false;
INSERT INTO bak_asset_retraction_20260822
SELECT * FROM agent_definitions
WHERE type='asset-retraction' AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

-- Pre-state guard: refuse unless the live config is exactly the armed shape
-- this file was written against (a concurrent edit aborts, never overwritten).
DO $$
DECLARE cfg jsonb;
BEGIN
  SELECT default_config #> '{workflow,steps,retract,config}' INTO cfg
  FROM agent_definitions
  WHERE type='asset-retraction' AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
  IF cfg IS NULL THEN
    RAISE EXCEPTION 'REFUSED: retract step config missing — agent restructured since this file was written';
  END IF;
  IF cfg->'dry_run' IS DISTINCT FROM 'false'::jsonb THEN
    RAISE EXCEPTION 'REFUSED: retract config dry_run is % — not the armed false this file exists to remove; nothing to do or the shape moved', cfg->'dry_run';
  END IF;
END $$;

UPDATE agent_definitions
SET default_config = default_config #- '{workflow,steps,retract,config,dry_run}',
    updated_at = NOW()
WHERE type='asset-retraction' AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL
  AND default_config #> '{workflow,steps,retract,config}' ? 'dry_run';

-- Verify loudly (DO/RAISE — SELECTs cannot stop a COMMIT).
DO $$
DECLARE cfg jsonb;
BEGIN
  SELECT default_config #> '{workflow,steps,retract,config}' INTO cfg
  FROM agent_definitions
  WHERE type='asset-retraction' AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
  IF cfg ? 'dry_run' THEN
    RAISE EXCEPTION 'dry_run still present after update: %', cfg;
  END IF;
  IF NOT (cfg ? 'site_id_field' AND cfg ? 'paths_field') THEN
    RAISE EXCEPTION 'a field reference was lost: %', cfg;
  END IF;
  RAISE NOTICE 'OK: retract config = % (action default dry-run TRUE governs again)', cfg;
END $$;

COMMIT;

-- ROLLBACK (re-arms — only with a live operation's explicit need):
--   UPDATE agent_definitions a SET default_config = b.default_config
--     FROM bak_asset_retraction_20260822 b WHERE a.id = b.id;
