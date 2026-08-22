-- 549_..._ROLLBACK.sql — restore generate_template's pre-549 budget config.
--
-- Restores ai_service.max_tokens to 16000 and REMOVES max_tokens_ceiling.
-- Note the ordering hazard: if the bugs_open/337 escalation code has rolled,
-- removing the ceiling only disarms escalation (safe — the step reverts to
-- fail-loud at 16000, i.e. the original bugs_open/337 failure mode returns).
-- Rolling back the cap while the failing section type is still planned
-- re-parks those pages; do it only if 24000 itself is implicated.

SELECT snapshot_agent('component-creator', 'migration 549 ROLLBACK: pre-revert');

BEGIN;

DO $$
DECLARE
    step_cap text;
BEGIN
    SELECT default_config#>>'{workflow,steps,generate_template,config,ai_service,max_tokens}'
    INTO step_cap
    FROM agent_definitions
    WHERE type='component-creator' AND is_active
      AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
    IF step_cap IS DISTINCT FROM '24000' THEN
        RAISE EXCEPTION 'ROLLBACK 549: max_tokens is % not 24000 — refusing to revert a value 549 did not write.', COALESCE(step_cap,'ABSENT');
    END IF;
END $$;

UPDATE agent_definitions
SET default_config = jsonb_set(default_config,
        '{workflow,steps,generate_template,config,ai_service,max_tokens}',
        to_jsonb(16000))
        #- '{workflow,steps,generate_template,config,ai_service,max_tokens_ceiling}',
    version = version + 1,
    updated_at = now()
WHERE type='component-creator' AND is_active
  AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

DO $$
DECLARE
    step_cap text;
    ceiling  text;
BEGIN
    SELECT default_config#>>'{workflow,steps,generate_template,config,ai_service,max_tokens}',
           default_config#>>'{workflow,steps,generate_template,config,ai_service,max_tokens_ceiling}'
    INTO step_cap, ceiling
    FROM agent_definitions
    WHERE type='component-creator' AND is_active
      AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
    IF step_cap IS DISTINCT FROM '16000' THEN
        RAISE EXCEPTION 'ROLLBACK 549: max_tokens is %, expected 16000.', COALESCE(step_cap,'ABSENT');
    END IF;
    IF ceiling IS NOT NULL THEN
        RAISE EXCEPTION 'ROLLBACK 549: max_tokens_ceiling still present (%).', ceiling;
    END IF;
END $$;

COMMIT;
