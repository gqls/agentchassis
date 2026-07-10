-- 141_reenable_index_plan_after_chunk_fix.sql — undo the 140 bypass once the
-- chunkContent fix is DEPLOYED. DRAFT 2026-07-10, parked in travelling_docs/
-- ON PURPOSE: move it into docs/agent_docs/sql_for_agents/ only when the
-- precondition below holds, then `./scripts/migration/run-migrations.sh --apply`.
--
-- PRECONDITION: the running chassis image must carry the rag_actions.go
-- chunkContent termination fix (break on final chunk + forward-progress guard;
-- regression tests in rag_actions_chunk_test.go). Check the running image tag
-- is built from a commit containing that change:
--   kubectl -n ai-persona-system get pods -l app=agent-chassis \
--     -o jsonpath='{.items[0].spec.containers[0].image}'
--   git log --oneline -3 -- platform/orchestration/actions/rag_actions.go
-- On an UNFIXED binary this migration re-arms the OOM: any tool creation kills
-- the shared generic pod (140's incident record).
--
-- PROOF after applying: re-run the 085 trigger with
--   SPEC_FUNCTION=tool-drop-rate-tuner (same function — create_tool_component
--   updates in place; its PLAN is machine-written v1, superseding it loses
--   nothing hand-made), and expect the full path save_tool -> compose_plan ->
--   write_plan -> index_plan -> complete plus, for the first time ever:
--   SELECT count(*) FROM knowledge_base WHERE collection='tool_docs';  -- > 0

BEGIN;

SELECT snapshot_agent('tool-generator', '141_reenable_index_plan_after_chunk_fix.sql: pre-update');

UPDATE agent_definitions
SET default_config = jsonb_set(
        jsonb_set(
            default_config,
            '{workflow,steps,write_plan,next_step}',
            '"index_plan"'::jsonb,
            true),
        '{workflow,steps,index_plan,description}',
        to_jsonb('Index the PLAN into knowledge_base tool_docs. Re-enabled after the chunkContent termination fix deployed (the confirmed cause of both index_plan OOMKills — see 140); per-chunk embedding deadline (120s default) retained as hygiene.'::text),
        true)
WHERE type = 'tool-generator' AND deleted_at IS NULL;

-- Pipeline note (runbook §3).
INSERT INTO doc_notes (subject_type, subject_key, body, categories, source, created_by)
VALUES ('pipeline', 'build',
$note$## 141: index_plan re-enabled after chunkContent fix deployed
Observed: 140 had bypassed index_plan because chunkContent() looped forever on content > chunk_size, OOMKilling the shared pod on every PLAN-sized index.
Root cause: closed by the chassis fix (final-chunk break + forward-progress guard, regression-tested).
Fix: write_plan.next_step -> index_plan restored; tool PLANs index into knowledge_base tool_docs from here on.
Verified: fixed image confirmed running before apply; proof run expected to land the first tool_docs rows.
Categories: migration$note$,
'["migration"]'::jsonb, 'human', 'run-migrations.sh');

DO $$
DECLARE n int;
BEGIN
    SELECT count(*) INTO n FROM agent_definitions
    WHERE type = 'tool-generator' AND deleted_at IS NULL
      AND default_config #>> '{workflow,steps,write_plan,next_step}' = 'index_plan'
      AND default_config #>> '{workflow,steps,index_plan,action}'    = 'rag_index'
      AND default_config #>> '{workflow,steps,save_tool,next_step}'  = 'compose_plan';
    IF n <> 1 THEN RAISE EXCEPTION 'index_plan re-enable incomplete (found %)', n; END IF;
END $$;

COMMIT;

-- Verify:
--   SELECT default_config #>> '{workflow,steps,write_plan,next_step}' AS write_next
--   FROM agent_definitions WHERE type='tool-generator' AND deleted_at IS NULL;
--   -- expect: index_plan
-- Rollback: restore the snapshot, or re-run the 140 UPDATE.
