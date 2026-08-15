-- 418_improvement_loop_gains_brief_fidelity_HOLD.sql
--
-- ⚠ HELD (uppercase suffix: run-migrations.sh --apply skips this file; it lists it
-- under "Sidecars (hand-run only)"). DO NOT APPLY until the ordering condition below
-- holds, then: apply by hand → rename to drop _HOLD → --record-only.
--
-- ORDERING CONDITION. This wiring makes the improvement sweep dispatch
-- brief-fidelity-auditor on every due audit. If the running chassis predates commit
-- d6d56e540 (the bugs_open/279 routing fix), any off-vocabulary category the
-- auditor emits is MINTED as an unrouteable audit_finding_* row instead of filed as
-- a capability_gap — the exact defect the owner had cleaned up. Apply only when:
--
--   STAMP=$(kubectl -n ai-persona-system logs -l app=agent-chassis --tail=300 \
--           | grep -m1 'build provenance')   # or read the stamp per CLAUDE.md
--   git merge-base --is-ancestor d6d56e540 <stamp-commit>   # must exit 0
--
-- (Also required first: migration 417 applied — the auditor must speak the router
-- vocabulary before anything dispatches it. 417 is a normal migration and will be
-- live long before this; the guard below re-checks it anyway.)
--
-- Then:
--   kubectl -n ai-persona-system exec -i postgres-clients-0 -- \
--     psql -U clients_user -d clients_db -v ON_ERROR_STOP=1 < <this file>
--   git mv docs/agent_docs/sql_for_agents/418_improvement_loop_gains_brief_fidelity_HOLD.sql \
--          docs/agent_docs/sql_for_agents/418_improvement_loop_gains_brief_fidelity.sql
--   ./scripts/migration/run-migrations.sh --record-only 418_improvement_loop_gains_brief_fidelity.sql \
--     --note "<what you verified>"
--
-- WHAT IT DOES (owner decision 2026-08-15, bugs_open/279 candidate 3, half two):
-- inserts a spawn/call pair for brief-fidelity-auditor into the improvement loop's
-- audit chain, mirroring the offer-analyser pair exactly — including error_step
-- continuing the sweep ("one auditor must not strand it; the child run is the
-- record of whether it worked"). Placement: between call_offer_analyser and
-- record_audit_pass, so the chain becomes
--   … call_site_review → spawn_offer_analyser → call_offer_analyser
--     → spawn_brief_fidelity → call_brief_fidelity → record_audit_pass → triage_findings …
-- The auditor's own degraded case is already safe: its prompt returns [] when the
-- site has no brief text, which write_audit_findings treats as a recognised empty
-- observation.
--
-- ROLLBACK: repoint call_offer_analyser's next_step/error_step back to
-- record_audit_pass and delete the two steps:
--   UPDATE agent_definitions SET default_config =
--     jsonb_set(jsonb_set(default_config,
--       '{workflow,steps,call_offer_analyser,next_step}', '"record_audit_pass"'),
--       '{workflow,steps,call_offer_analyser,error_step}', '"record_audit_pass"')
--     #- '{workflow,steps,spawn_brief_fidelity}' #- '{workflow,steps,call_brief_fidelity}',
--     version = version + 1, updated_at = now()
--   WHERE type='improvement-loop' AND is_active
--     AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

BEGIN;

DO $$
DECLARE
    n_defs integer;
    next_now text;
    err_now text;
    already integer;
    vocab_ok integer;
BEGIN
    SELECT count(*) INTO n_defs FROM agent_definitions
    WHERE type='improvement-loop' AND is_active
      AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
    IF n_defs <> 1 THEN
        RAISE EXCEPTION 'MIGRATION 418: expected exactly 1 live improvement-loop, found %', n_defs;
    END IF;

    SELECT default_config->'workflow'->'steps'->'call_offer_analyser'->>'next_step',
           default_config->'workflow'->'steps'->'call_offer_analyser'->>'error_step',
           CASE WHEN default_config->'workflow'->'steps' ? 'spawn_brief_fidelity' THEN 1 ELSE 0 END
      INTO next_now, err_now, already
    FROM agent_definitions
    WHERE type='improvement-loop' AND is_active
      AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

    IF already = 1 THEN
        RAISE EXCEPTION 'MIGRATION 418: spawn_brief_fidelity already present — already applied';
    END IF;
    IF next_now IS DISTINCT FROM 'record_audit_pass' OR err_now IS DISTINCT FROM 'record_audit_pass' THEN
        RAISE EXCEPTION 'MIGRATION 418: call_offer_analyser points at %/% (expected record_audit_pass/record_audit_pass) — the chain has changed since this file was written; re-derive the splice point', next_now, err_now;
    END IF;

    -- 417 must be live: the auditor must not be dispatched speaking a dead category.
    SELECT (length(default_config::text) - length(replace(default_config::text,'choose by REPAIR SHAPE','')))/length('choose by REPAIR SHAPE')
      INTO vocab_ok
    FROM agent_definitions
    WHERE type='brief-fidelity-auditor' AND is_active
      AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
    IF vocab_ok <> 1 THEN
        RAISE EXCEPTION 'MIGRATION 418: brief-fidelity-auditor does not carry the 417 vocabulary (found %) — apply 417 first', vocab_ok;
    END IF;
END $$;

UPDATE agent_definitions
SET default_config = jsonb_set(jsonb_set(jsonb_set(jsonb_set(default_config,
        '{workflow,steps,call_offer_analyser,next_step}', '"spawn_brief_fidelity"'),
        '{workflow,steps,call_offer_analyser,error_step}', '"spawn_brief_fidelity"'),
        '{workflow,steps,spawn_brief_fidelity}', '{
            "action": "spawn_agent",
            "config": {"role": "brief_fidelity_auditor", "agent_type": "brief-fidelity-auditor"},
            "next_step": "call_brief_fidelity",
            "description": "Spawn the brief-fidelity auditor (owner decision 2026-08-15, bugs_open/279)",
            "output_field": "brief_fidelity_spawned"
        }'::jsonb),
        '{workflow,steps,call_brief_fidelity}', '{
            "action": "call_agent",
            "config": {
                "target_role": "brief_fidelity_auditor",
                "input_mapping": {"domain": "site_record.domain", "site_id": "site_record.site_id"},
                "timeout_seconds": 600
            },
            "next_step": "record_audit_pass",
            "error_step": "record_audit_pass",
            "description": "Brief-fidelity audit: grades the built site against its own brief; findings carry the router vocabulary (mig 417) under audit_source brief-fidelity-audit. error_step continues the sweep — one auditor must not strand it, and the child run is the record of whether it worked",
            "output_field": "brief_fidelity_result"
        }'::jsonb),
    version    = version + 1,
    updated_at = now()
WHERE type='improvement-loop' AND is_active
  AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

DO $$
DECLARE
    spliced text;
    chain_out text;
    n_steps integer;
BEGIN
    SELECT default_config->'workflow'->'steps'->'call_offer_analyser'->>'next_step',
           default_config->'workflow'->'steps'->'call_brief_fidelity'->>'next_step'
      INTO spliced, chain_out
    FROM agent_definitions
    WHERE type='improvement-loop' AND is_active
      AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

    IF spliced IS DISTINCT FROM 'spawn_brief_fidelity' OR chain_out IS DISTINCT FROM 'record_audit_pass' THEN
        RAISE EXCEPTION 'MIGRATION 418: splice failed (offer→% , brief→%)', spliced, chain_out;
    END IF;

    SELECT count(*) INTO n_steps
    FROM jsonb_object_keys((SELECT default_config->'workflow'->'steps'
                            FROM agent_definitions WHERE type='improvement-loop'
                              AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL)) k;
    IF n_steps <> 30 THEN
        RAISE EXCEPTION 'MIGRATION 418: expected 30 steps after the splice (28 + 2), found %', n_steps;
    END IF;

    RAISE NOTICE 'migration 418 OK: brief-fidelity wired into the audit chain (30 steps)';
END $$;

COMMIT;
