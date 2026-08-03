-- 301 — render-audit-agent: the write tail (A0.3b) — findings become work items
--
-- vigilant_designer_offer_analysis programme, Phase A0.3b (PLAN 2026-08-02).
--
-- The 256 seed's header named this extension point deliberately: the audit's
-- findings stopped in collected_data (bugs_open/115 / 083 shape — a correct
-- measurement nobody consumed). The chassis action write_render_audit_findings
-- (councils e49f5935 r2 APPROVED 2026-08-03) files firm contrast failures at
-- css-patch-agent and attributed broken images at asset-deployer, born
-- 'detected' for the 286 single promoter. This migration wires it in:
--
--   site → audit → write_findings → complete    (+ complete_error unchanged)
--
-- ORDERING: image BEFORE config. The action ships in agent-chassis v1.0.1241
-- (pod-verified before this was applied). A workflow naming an action the
-- running binary does not register is bug-017 shaped: 'completed' with no work.
--
-- The step reads the audit step's own output_field (render_audit); the action
-- unwraps the coordinator's .response envelope itself and FAILS LOUD when the
-- audit is absent or still awaited — absent ≠ clean. error_step is STEP-LEVEL
-- (coordinator.routeToErrorStepOrFail checks step-level first; config-level is
-- the backward-compat fallback).
--
-- Apply alone (never a blanket --apply — the dir carries other sessions'
-- pending files):
--   kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -f - < 299_...sql
--   ./docs/agent_docs/sql_for_agents/run-migrations.sh --record-only 299

SELECT snapshot_agent('render-audit-agent', '301_render_audit_agent_write_findings_tail.sql: pre-update');

BEGIN;

UPDATE agent_definitions
SET default_config =
    jsonb_set(
        jsonb_set(
            jsonb_set(
                default_config,
                '{workflow,steps,write_findings}',
                $step$
{
  "action": "write_render_audit_findings",
  "config": {
    "site_id": "site_record.site_id"
  },
  "error_step": "complete_error",
  "next_step": "complete",
  "description": "File the audit's firm findings as routed work items (contrast_failure → css-patch-agent, attributed broken images → undeployed_asset/asset-deployer), born detected for the 286 promoter. Fails loud when render_audit is absent or awaited — a failed audit and a clean audit must never be read the same way.",
  "output_field": "findings_written"
}
                $step$::jsonb
            ),
            '{workflow,steps,audit,next_step}',
            '"write_findings"'
        ),
        '{workflow,steps,complete,config,multiple_output_fields}',
        '["render_audit", "findings_written"]'
    ),
    updated_at = now()
WHERE type = 'render-audit-agent'
  AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

-- VERIFY — DO/RAISE, not bare SELECTs: a SELECT's non-empty result cannot stop
-- the COMMIT (WRONG_CALLS/memory lesson), an exception can. Reads start_step
-- (the 256 lesson: the engine's key is start_step, never initial_step).
DO $verify$
DECLARE
    cfg jsonb;
BEGIN
    SELECT default_config INTO cfg
    FROM agent_definitions
    WHERE type = 'render-audit-agent'
      AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

    IF cfg IS NULL THEN
        RAISE EXCEPTION '301 VERIFY: no active render-audit-agent row';
    END IF;
    IF cfg #>> '{workflow,start_step}' IS DISTINCT FROM 'site' THEN
        RAISE EXCEPTION '301 VERIFY: start_step = %, want site', cfg #>> '{workflow,start_step}';
    END IF;
    IF cfg #>> '{workflow,steps,audit,next_step}' IS DISTINCT FROM 'write_findings' THEN
        RAISE EXCEPTION '301 VERIFY: audit.next_step = %, want write_findings', cfg #>> '{workflow,steps,audit,next_step}';
    END IF;
    IF cfg #>> '{workflow,steps,write_findings,action}' IS DISTINCT FROM 'write_render_audit_findings' THEN
        RAISE EXCEPTION '301 VERIFY: write_findings.action = %', cfg #>> '{workflow,steps,write_findings,action}';
    END IF;
    IF cfg #>> '{workflow,steps,write_findings,error_step}' IS DISTINCT FROM 'complete_error' THEN
        RAISE EXCEPTION '301 VERIFY: write_findings.error_step = % (must be step-level)', cfg #>> '{workflow,steps,write_findings,error_step}';
    END IF;
    IF NOT (cfg #> '{workflow,steps,complete,config,multiple_output_fields}') ? 'findings_written' THEN
        RAISE EXCEPTION '301 VERIFY: complete does not surface findings_written';
    END IF;
END
$verify$;

COMMIT;
