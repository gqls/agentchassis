-- SQL_2026-07-09_register_image_source_unsatisfiable.sql
--
-- Registers the new image_source_unsatisfiable discovery check on
-- design-discovery-agent's run_checks list. The check flags component
-- input_schema image fields sourced from a site_assets.<path> that no asset
-- key, plan imagery row, or image-role alias can supply (the 2026-07-09
-- robot-hands finding: empty src="" / shared placeholder hero), so every
-- future domain gets caught automatically.
--
-- RUN AFTER the chassis image containing check_image_source_unsatisfiable.go
-- is deployed — registering an unknown check name earlier just logs a skip,
-- but sequence it properly anyway.
--
-- Idempotent: the NOT ... ? guard makes re-runs no-ops.

\set ON_ERROR_STOP on

BEGIN;

-- Snapshot per convention before mutating agent default_config.
SELECT snapshot_agent('design-discovery-agent',
                      'register image_source_unsatisfiable discovery check (imagery best-in-class I0)');

UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,run_checks,config,checks}',
        (default_config #> '{workflow,steps,run_checks,config,checks}')
            || '["image_source_unsatisfiable"]'::jsonb,
        false),
    updated_at = now()
WHERE type = 'design-discovery-agent'
  AND is_active = true
  AND (is_snapshot IS NULL OR is_snapshot = false)
  AND NOT (default_config #> '{workflow,steps,run_checks,config,checks}')
          ? 'image_source_unsatisfiable';

-- Verify
DO $verify$
DECLARE
    v_has boolean;
BEGIN
    SELECT (default_config #> '{workflow,steps,run_checks,config,checks}')
           ? 'image_source_unsatisfiable'
      INTO v_has
      FROM agent_definitions
     WHERE type = 'design-discovery-agent'
       AND is_active = true
       AND (is_snapshot IS NULL OR is_snapshot = false);
    IF NOT COALESCE(v_has, false) THEN
        RAISE EXCEPTION 'image_source_unsatisfiable not present in checks array after update';
    END IF;
    RAISE NOTICE 'image_source_unsatisfiable registered on design-discovery-agent';
END
$verify$;

COMMIT;
