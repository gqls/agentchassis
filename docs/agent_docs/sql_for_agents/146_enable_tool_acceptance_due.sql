-- 146_enable_tool_acceptance_due.sql — Tier 4 goes CONTINUOUS: enable the
-- periodic acceptance sweep in the design discovery agent. DB-only; idempotent.
--
-- WHAT: adds "tool_acceptance_due" to design-discovery-agent's run_checks
-- (beside tool_health and tool_acceptance). The check
-- (discovery_checks/check_tool_acceptance_due.go) emits ONE acceptance_run
-- work item (handler tool-acceptance-agent, status 'triaged', pipeline
-- 'build', priority 90 — after builds/rerenders so acceptance tests the NEW
-- page) for every active tool with a DEPLOYED page and a current PLAN
-- criteria fence, unless a verdict note landed within 7 days or a run is
-- already open. Emitting straight to 'triaged' follows the
-- create_rerender_items precedent: acceptance needs no human judgment, and
-- 'detected' items were observed sitting unswept on this site.
--
-- SAFE TO APPLY BEFORE THE CHASSIS DEPLOYS the check (run_discovery_checks
-- warn-skips unknown names — the 142 precedent); it activates with the next
-- image, which also carries the correct-while-touching fix declared here:
-- check_tool_acceptance's improve_tool cooldown now EXCLUDES cancelled items
-- (a cancelled item = resolved another way; the tool is re-checkable — the
-- recorded 2026-07-10 follow-up).

BEGIN;

SELECT snapshot_agent('design-discovery-agent', '146_enable_tool_acceptance_due.sql: pre-update');

UPDATE agent_definitions
SET default_config = CASE
      WHEN default_config #> '{workflow,steps,run_checks,config,checks}' ? 'tool_acceptance_due'
        THEN default_config
      ELSE jsonb_set(default_config,
             '{workflow,steps,run_checks,config,checks}',
             (default_config #> '{workflow,steps,run_checks,config,checks}') || '["tool_acceptance_due"]'::jsonb)
    END
WHERE type = 'design-discovery-agent' AND deleted_at IS NULL;

-- Pipeline note (runbook §3).
INSERT INTO doc_notes (subject_type, subject_key, body, categories, source, created_by)
VALUES ('pipeline', 'build',
$note$## 146: periodic Tier-4 acceptance sweep enabled (tool_acceptance_due)
Observed: acceptance runs existed only as manual 087 triggers; nothing re-verifies tools over time or after changes land.
Root cause: not-applicable (new capability — the Stage 6 trigger-points item).
Fix: discovery check tool_acceptance_due added to design-discovery-agent.run_checks — emits acceptance_run items (handler tool-acceptance-agent, status triaged, priority 90) for deployed tools with current PLAN criteria; 7-day verdict cooldown; skips open runs. Also in the same image: check_tool_acceptance's improve_tool cooldown now excludes cancelled items.
Verified: build + package tests green; activates when the carrying image deploys (unknown names warn+skip until then).
Categories: migration$note$,
'["migration"]'::jsonb, 'human', 'run-migrations.sh');

DO $$
DECLARE n int;
BEGIN
    SELECT count(*) INTO n FROM agent_definitions
    WHERE type = 'design-discovery-agent' AND deleted_at IS NULL
      AND default_config #> '{workflow,steps,run_checks,config,checks}' ? 'tool_acceptance_due'
      AND default_config #> '{workflow,steps,run_checks,config,checks}' ? 'tool_acceptance'
      AND default_config #> '{workflow,steps,run_checks,config,checks}' ? 'tool_health';
    IF n <> 1 THEN RAISE EXCEPTION 'tool_acceptance_due enablement incomplete (found %)', n; END IF;
END $$;

COMMIT;

-- Verify:
--   SELECT default_config #> '{workflow,steps,run_checks,config,checks}'
--   FROM agent_definitions WHERE type='design-discovery-agent' AND deleted_at IS NULL;
-- Proof (after the carrying image deploys + a sweep runs): an acceptance_run
-- item appears, the dispatch loop drives tool-acceptance-agent, and a fresh
-- acceptance-run note lands WITHOUT any manual trigger:
--   SELECT item_key, status FROM site_work_items WHERE item_type='acceptance_run';
--   SELECT subject_key, categories, created_at FROM doc_notes
--   WHERE source='tool-acceptance' ORDER BY created_at DESC LIMIT 5;
-- Rollback: restore the snapshot, or remove the array element (142's recipe).
