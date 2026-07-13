-- 0NN_diagnose_load_runtime_error_step.sql — let a code-only diagnosis survive.
-- DRAFT 2026-07-06. Renumber 0NN.
--
-- Incident (2026-07-06, correlation 48f937e5-…): a subjectless/anchorless smoke
-- diagnosis failed the WHOLE child workflow at step load_runtime:
--   "diagnose_load_runtime: need at least one of site_id / correlation_id /
--    domain in collected_data"
-- so the run never reached verdict/emit/persist_note. Structural gap: runtime
-- evidence is an OPTIONAL bundle tier (diagnose_assemble_bundle omits the
-- section when empty), but the step has no effective error routing, making the
-- tier mandatory in practice.
--
-- Fix (001 §16 mechanism): config-level error_step -> "assemble" on
-- load_runtime. The SUCCESS path already goes to assemble, so success and
-- failure converge on the same next step — an anchorless run proceeds with a
-- code+schema bundle. error_step MUST be inside config (step-level is silently
-- ignored — coordinator reads step.Config["error_step"]).
--
-- COALESCE guards the case where the step has no config object at all
-- (jsonb_set does not create missing parents).
--
-- Targets the CURRENT version row (type + max version, deleted_at IS NULL).
-- Follow-up (separate, next chassis build): soften diagnose_load_runtime to
-- return {runtime_evidence:"", skipped:true, reason} when no anchor is given,
-- keeping the hard error for genuine DB failures.

BEGIN;

WITH current_diag AS (
    SELECT id
    FROM agent_definitions
    WHERE type = 'diagnose-agent'
      AND deleted_at IS NULL
    ORDER BY version DESC
    LIMIT 1
)
UPDATE agent_definitions ad
SET default_config = jsonb_set(
        ad.default_config,
        '{workflow,steps,load_runtime,config}',
        COALESCE(ad.default_config #> '{workflow,steps,load_runtime,config}', '{}'::jsonb)
            || '{"error_step": "assemble"}'::jsonb,
        true
    ),
    updated_at = now()
FROM current_diag
WHERE ad.id = current_diag.id;

-- Guard: exactly one row carries the routing.
DO $$
DECLARE n int;
BEGIN
    SELECT count(*) INTO n
    FROM agent_definitions
    WHERE type = 'diagnose-agent' AND deleted_at IS NULL
      AND default_config #>> '{workflow,steps,load_runtime,config,error_step}' = 'assemble';
    IF n <> 1 THEN
        RAISE EXCEPTION 'expected exactly 1 diagnose-agent with load_runtime error routing, found %', n;
    END IF;
END $$;

COMMIT;

-- Verify after apply:
--   SELECT jsonb_pretty(default_config #> '{workflow,steps,load_runtime}')
--   FROM agent_definitions
--   WHERE type='diagnose-agent' AND deleted_at IS NULL
--   ORDER BY version DESC LIMIT 1;
--
-- Rollback (manual):
--   UPDATE agent_definitions
--   SET default_config = default_config #- '{workflow,steps,load_runtime,config,error_step}',
--       updated_at = now()
--   WHERE type='diagnose-agent' AND deleted_at IS NULL;
