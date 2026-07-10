-- 140_rebypass_index_plan_chunk_loop.sql — bypass index_plan again until the
-- chunkContent fix deploys. DB-only; effective immediately. REVERSIBLE.
--
-- ROOT CAUSE — NOW CONFIRMED (supersedes 135's "no deadline" and the later
-- "slow leak" hypotheses): rag_actions.go chunkContent() never terminated on
-- content longer than chunk_size. The final chunk ends at len(content), then
-- start = end - overlap re-enters the loop and appends the same ~200-char tail
-- forever — ~2Gi of duplicate chunks in seconds. Both chassis OOMKills were
-- this loop:
--   * 2026-07-09 ~13:08 UTC — 23s into index_plan on the 2,982-char
--     tool-xp-curve-designer PLAN (misread as a stall, then as a leak);
--   * 2026-07-10 12:08 UTC — minutes into run 75c512bf (proof run for 139) on
--     the 3,010-char tool-drop-rate-tuner PLAN. Pod restart count 1,
--     Last State: OOMKilled, exit 137.
-- Content <= chunk_size returns early, which is why small rag_index calls
-- never crashed. The embedding deadline shipped in v1.0.1102 was hygiene; it
-- could not help — the loop runs BEFORE any embedding call.
--
-- FIX (same working tree as this file): chunkContent breaks after the final
-- chunk and enforces forward progress for any overlap config; regression
-- tests in rag_actions_chunk_test.go. Until that binary deploys, ANY tool
-- creation OOMs the shared generic pod — so bypass again now.
--
-- What run 75c512bf left behind (all healthy; only indexing was lost):
-- tool-drop-rate-tuner component + page (build_status planned) + current PLAN
-- (3,010 chars, fence intact). knowledge_base tool_docs remains 0 rows.
--
-- RE-ENABLE: 141 (drafted in travelling_docs/, moved here only after the
-- fixed chassis image is verified running — do not renumber it in early; a
-- premature run of the runner must not re-enable this).

BEGIN;

SELECT snapshot_agent('tool-generator', '140_rebypass_index_plan_chunk_loop.sql: pre-update');

UPDATE agent_definitions
SET default_config = jsonb_set(
        jsonb_set(
            default_config,
            '{workflow,steps,write_plan,next_step}',
            '"complete"'::jsonb,
            true),
        '{workflow,steps,index_plan,description}',
        to_jsonb('BYPASSED again 2026-07-10 (write_plan -> complete): chunkContent() in rag_actions.go loops forever on content > chunk_size (confirmed cause of both OOMKills). Re-enable via 141 once the chassis image carrying the chunkContent fix + rag_actions_chunk_test.go is deployed.'::text),
        true)
WHERE type = 'tool-generator' AND deleted_at IS NULL;

-- Pipeline note (runbook §3: workflow-altering migrations leave number/what/why).
INSERT INTO doc_notes (subject_type, subject_key, body, categories, source, created_by)
VALUES ('pipeline', 'build',
$note$## 140: index_plan re-bypassed — OOM root cause confirmed as chunkContent loop
Observed: chassis pod OOMKilled (exit 137) twice inside index_plan/rag_index: 2026-07-09 on the 2,982-char xp-curve PLAN (23s in), 2026-07-10 on the 3,010-char drop-rate-tuner PLAN (run 75c512bf).
Root cause: rag_actions.go chunkContent() never terminates on content > chunk_size — the tail chunk re-enters via start = end - overlap and duplicates until the 2Gi limit. Small content returns early, which hid it. Supersedes the earlier missing-deadline (135) and slow-leak readings.
Fix: loop now breaks on the final chunk and enforces forward progress; regression tests added. Workflow: write_plan.next_step -> complete until the fixed image deploys; 141 re-enables.
Verified: tests pass; go build clean. tool-drop-rate-tuner's component/page/PLAN from the crashed run are intact; only the KB index write was lost (tool_docs still 0).
Categories: migration, diagnosis$note$,
'["migration","diagnosis"]'::jsonb, 'human', 'run-migrations.sh');

DO $$
DECLARE n int;
BEGIN
    SELECT count(*) INTO n FROM agent_definitions
    WHERE type = 'tool-generator' AND deleted_at IS NULL
      AND default_config #>> '{workflow,steps,write_plan,next_step}' = 'complete'
      AND default_config #>> '{workflow,steps,index_plan,action}' = 'rag_index'
      AND default_config #>> '{workflow,steps,save_tool,next_step}' = 'compose_plan';
    IF n <> 1 THEN RAISE EXCEPTION 'index_plan re-bypass incomplete (found %)', n; END IF;

    SELECT count(*) INTO n FROM doc_notes
    WHERE subject_type='pipeline' AND subject_key='build'
      AND categories ? 'migration' AND body LIKE '%140: index_plan re-bypassed%';
    IF n < 1 THEN RAISE EXCEPTION 'migration note missing'; END IF;
END $$;

COMMIT;

-- Verify:
--   SELECT default_config #>> '{workflow,steps,write_plan,next_step}' AS write_next
--   FROM agent_definitions WHERE type='tool-generator' AND deleted_at IS NULL;
--   -- expect: complete
-- Rollback: restore the snapshot, or set write_plan.next_step back to 'index_plan'
-- (that is exactly what 141 does — only do it on a fixed binary).
