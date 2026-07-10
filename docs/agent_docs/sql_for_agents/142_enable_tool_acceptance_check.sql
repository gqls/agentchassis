-- 142_enable_tool_acceptance_check.sql — Stage 5: enable the Tier-2 static
-- acceptance check in the design discovery sweep. DB-only; idempotent.
--
-- WHAT: adds "tool_acceptance" to design-discovery-agent's run_checks list
-- (where tool_health already runs). The check itself is
-- platform/orchestration/actions/discovery_checks/check_tool_acceptance.go:
-- for each active tool with a CURRENT PLAN criteria fence, it fetches the
-- DEPLOYED page and asserts the statically-visible subset under the ANCHOR
-- RULE (validate a selector's leftmost id/class token, never the whole path;
-- confirm, never refute; -EDIT ids skipped). No criteria → needs_criteria
-- note, never a fake pass. Failures → improve_tool item (criteria embedded as
-- acceptance_test) + acceptance-fail note. Shares tool_health's 7-day
-- improve_tool cooldown per component.
--
-- SAFE TO APPLY BEFORE THE CHASSIS DEPLOYS the new check:
-- run_discovery_checks warns "Unknown discovery check — not registered" and
-- skips unknown names (discovery_checks.go), so the check simply activates
-- with the next image.

BEGIN;

SELECT snapshot_agent('design-discovery-agent', '142_enable_tool_acceptance_check.sql: pre-update');

UPDATE agent_definitions
SET default_config = CASE
      WHEN default_config #> '{workflow,steps,run_checks,config,checks}' ? 'tool_acceptance'
        THEN default_config
      ELSE jsonb_set(default_config,
             '{workflow,steps,run_checks,config,checks}',
             (default_config #> '{workflow,steps,run_checks,config,checks}') || '["tool_acceptance"]'::jsonb)
    END
WHERE type = 'design-discovery-agent' AND deleted_at IS NULL;

-- Pipeline note (runbook §3: workflow-altering migrations leave number/what/why).
INSERT INTO doc_notes (subject_type, subject_key, body, categories, source, created_by)
VALUES ('pipeline', 'build',
$note$## 142: Tier-2 static acceptance check enabled (Stage 5)
Observed: tools pass completeness+validation while shipping behavioural bugs (economy-simulator, twice); composers can invent selectors (#xpTableBody).
Root cause: no layer asserted the PLAN's acceptance criteria against the deployed artifact.
Fix: discovery check tool_acceptance added to design-discovery-agent.run_checks — anchor-rule static verification of criteria selectors/assets/status against the deployed page; needs_criteria note when no PLAN criteria; failures create improve_tool items carrying the criteria as acceptance_test plus an acceptance-fail note.
Verified: unit tests (anchor rule incl. the #tableWrap/#xpTableBody founding cases, -EDIT skip, shell checks); check activates when the carrying image deploys (unknown names warn+skip until then).
Categories: migration$note$,
'["migration"]'::jsonb, 'human', 'run-migrations.sh');

DO $$
DECLARE n int;
BEGIN
    SELECT count(*) INTO n FROM agent_definitions
    WHERE type = 'design-discovery-agent' AND deleted_at IS NULL
      AND default_config #> '{workflow,steps,run_checks,config,checks}' ? 'tool_acceptance'
      AND default_config #> '{workflow,steps,run_checks,config,checks}' ? 'tool_health';
    IF n <> 1 THEN RAISE EXCEPTION 'tool_acceptance enablement incomplete (found %)', n; END IF;
END $$;

COMMIT;

-- Verify:
--   SELECT default_config #> '{workflow,steps,run_checks,config,checks}'
--   FROM agent_definitions WHERE type='design-discovery-agent' AND deleted_at IS NULL;
-- Rollback: restore the snapshot, or remove the array element:
--   UPDATE agent_definitions SET default_config = jsonb_set(default_config,
--     '{workflow,steps,run_checks,config,checks}',
--     (SELECT jsonb_agg(e) FROM jsonb_array_elements(
--        default_config #> '{workflow,steps,run_checks,config,checks}') e
--      WHERE e <> '"tool_acceptance"'::jsonb))
--   WHERE type='design-discovery-agent' AND deleted_at IS NULL;
