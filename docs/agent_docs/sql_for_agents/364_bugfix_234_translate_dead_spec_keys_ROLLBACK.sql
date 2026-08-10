-- ============================================================================
-- 364_bugfix_234_translate_dead_spec_keys_ROLLBACK.sql
--
-- Reverses 364: renames spec_literal/spec_paths back to `spec` on the three
-- steps, restoring the (dead) pre-364 spelling. Values copied verbatim from
-- the live keys.
--
-- ⚠ Do NOT apply this after the bugs_open/234 CODE half has rolled: that
-- binary declares `spec` a REMOVED key on create_work_item and hard-fails the
-- whole workflow at validation, on every message, for any definition carrying
-- it — rolling back the data alone would kill improvement-loop and
-- deduplicate-sections outright. Roll back the code first, or not at all.
-- ============================================================================

BEGIN;

DO $$
DECLARE
  il_cfg jsonb; nc_cfg jsonb; qr_cfg jsonb;
BEGIN
  SELECT default_config->'workflow'->'steps'->'insert_rerender_item'->'config',
         default_config->'workflow'->'steps'->'record_not_converging'->'config'
    INTO il_cfg, nc_cfg
    FROM agent_definitions
   WHERE type='improvement-loop' AND is_active
     AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
  SELECT default_config->'workflow'->'steps'->'queue_rerender'->'config'
    INTO qr_cfg
    FROM agent_definitions
   WHERE type='deduplicate-sections' AND is_active
     AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

  IF NOT (il_cfg ? 'spec_literal' AND nc_cfg ? 'spec_literal' AND qr_cfg ? 'spec_paths') THEN
    RAISE EXCEPTION '364 ROLLBACK: 364 does not appear applied (spec_literal on il=% nc=%, spec_paths on qr=%)',
      il_cfg ? 'spec_literal', nc_cfg ? 'spec_literal', qr_cfg ? 'spec_paths';
  END IF;
  IF il_cfg ? 'spec' OR nc_cfg ? 'spec' OR qr_cfg ? 'spec' THEN
    RAISE EXCEPTION '364 ROLLBACK: DRIFT — a `spec` key already present. Re-measure.';
  END IF;
END $$;

UPDATE agent_definitions
SET default_config =
      jsonb_set(
        default_config #- '{workflow,steps,insert_rerender_item,config,spec_literal}',
        '{workflow,steps,insert_rerender_item,config,spec}',
        default_config#>'{workflow,steps,insert_rerender_item,config,spec_literal}'),
    updated_at = now()
WHERE type='improvement-loop' AND is_active
  AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

UPDATE agent_definitions
SET default_config =
      jsonb_set(
        default_config #- '{workflow,steps,record_not_converging,config,spec_literal}',
        '{workflow,steps,record_not_converging,config,spec}',
        default_config#>'{workflow,steps,record_not_converging,config,spec_literal}'),
    updated_at = now()
WHERE type='improvement-loop' AND is_active
  AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

UPDATE agent_definitions
SET default_config =
      jsonb_set(
        default_config #- '{workflow,steps,queue_rerender,config,spec_paths}',
        '{workflow,steps,queue_rerender,config,spec}',
        default_config#>'{workflow,steps,queue_rerender,config,spec_paths}'),
    updated_at = now()
WHERE type='deduplicate-sections' AND is_active
  AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

DO $$
DECLARE
  il_cfg jsonb; nc_cfg jsonb; qr_cfg jsonb;
BEGIN
  SELECT default_config->'workflow'->'steps'->'insert_rerender_item'->'config',
         default_config->'workflow'->'steps'->'record_not_converging'->'config'
    INTO il_cfg, nc_cfg
    FROM agent_definitions
   WHERE type='improvement-loop' AND is_active
     AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
  SELECT default_config->'workflow'->'steps'->'queue_rerender'->'config'
    INTO qr_cfg
    FROM agent_definitions
   WHERE type='deduplicate-sections' AND is_active
     AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

  IF il_cfg ? 'spec_literal' OR nc_cfg ? 'spec_literal' OR qr_cfg ? 'spec_paths'
     OR NOT (il_cfg ? 'spec' AND nc_cfg ? 'spec' AND qr_cfg ? 'spec') THEN
    RAISE EXCEPTION '364 ROLLBACK VERIFY: rename-back incomplete (il=% nc=% qr=%)',
      il_cfg, nc_cfg, qr_cfg;
  END IF;
END $$;

COMMIT;
