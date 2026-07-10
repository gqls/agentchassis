-- SQL_2026-07-10_register_component_template_corrupted.sql
--
-- Registers the component_template_corrupted discovery check on
-- design-discovery-agent. The check is the quality→regeneration BRIDGE: it
-- detects active components used on the site whose html_template is baked
-- rendered output (literal '<no value>' holes, or zero {{...}} variables
-- while the schema declares fields) and emits needs_component_regeneration
-- items to component-creator — the same shape as the proven system-stats
-- manual precedent, with a cross-site open-item guard and a 5-per-pass cap.
--
-- RUN AFTER the chassis image containing
-- check_component_template_corrupted.go is deployed.
--
-- Idempotent: the NOT ... ? guard makes re-runs no-ops.

\set ON_ERROR_STOP on

BEGIN;

SELECT snapshot_agent('design-discovery-agent',
                      'register component_template_corrupted discovery check (imagery best-in-class, baked-template bridge)');

UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,run_checks,config,checks}',
        (default_config #> '{workflow,steps,run_checks,config,checks}')
            || '["component_template_corrupted"]'::jsonb,
        false),
    updated_at = now()
WHERE type = 'design-discovery-agent'
  AND is_active = true
  AND (is_snapshot IS NULL OR is_snapshot = false)
  AND NOT (default_config #> '{workflow,steps,run_checks,config,checks}')
          ? 'component_template_corrupted';

DO $verify$
DECLARE
    v_has boolean;
BEGIN
    SELECT (default_config #> '{workflow,steps,run_checks,config,checks}')
           ? 'component_template_corrupted'
      INTO v_has
      FROM agent_definitions
     WHERE type = 'design-discovery-agent'
       AND is_active = true
       AND (is_snapshot IS NULL OR is_snapshot = false);
    IF NOT COALESCE(v_has, false) THEN
        RAISE EXCEPTION 'component_template_corrupted not present in checks array after update';
    END IF;
    RAISE NOTICE 'component_template_corrupted registered on design-discovery-agent';
END
$verify$;

COMMIT;
