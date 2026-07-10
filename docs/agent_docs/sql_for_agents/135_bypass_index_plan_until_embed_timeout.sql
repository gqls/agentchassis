-- 0NN_bypass_index_plan_until_embed_timeout.sql — stop tool creation hanging.
-- DRAFT 2026-07-09. Renumber 0NN. DB-only; effective immediately. REVERSIBLE.
--
-- INCIDENT (orchestration 46cd5299, 2026-07-09): the run wrote its PLAN
-- (Task-3 proof) then sat in EXECUTING_STEP at `index_plan` for 2641s+. The
-- workflow's timeout_seconds=480 did NOT terminate it — observation: the
-- workflow timeout does not govern in-process action execution (hypothesis:
-- it governs awaited responses / child orchestrations; unconfirmed).
--
-- MECHANISM (rag_actions.go, RAGIndexAction): content -> chunkContent(1000/200)
-- -> for each chunk: createRAGEmbeddingClient() then GenerateEmbedding(ctx,...)
-- against http://ollama-adapter.ai-persona-system.svc.cluster.local:11434,
-- then INSERT. Embedding FAILURES are non-fatal by design ("storing without
-- embeddings"); a STALL is not a failure, and no deadline is applied. So a
-- reachable-but-unresponsive embedder hangs the step, and with it the run.
--
-- WHY THIS IS SAFE: index_plan is the LAST doc step. write_plan has already
-- persisted the PLAN to doc_plans (Postgres = truth). The knowledge_base copy
-- is a derived retrieval convenience (the standing "KB tool_docs write" item).
-- Bypassing it loses nothing and unhangs every future tool creation.
--
-- STRUCTURAL FIX (next chassis build — this migration is the stopgap):
--   * give the embedding path an explicit deadline:
--       embCtx, cancel := context.WithTimeout(ctx, 15*time.Second); defer cancel()
--     around GenerateEmbedding (and/or an http.Client{Timeout: ...} inside
--     aiservice.OllamaClient), so a stall degrades to the existing non-fatal
--     "store without embeddings" path;
--   * consider bounding total rag_index time (chunks x per-chunk deadline);
--   * consider an action-level deadline in the chassis so no action can hang a
--     workflow past timeout_seconds.
-- Re-enable index_plan (below) once that ships.

BEGIN;

SELECT snapshot_agent('tool-generator', '0NN_bypass_index_plan_until_embed_timeout.sql: pre-update');

WITH cur AS (
    SELECT id FROM agent_definitions
    WHERE type = 'tool-generator' AND deleted_at IS NULL
    ORDER BY version DESC LIMIT 1
)
UPDATE agent_definitions ad
SET default_config = jsonb_set(
        ad.default_config,
        '{workflow,steps,write_plan,next_step}',
        '"complete"'::jsonb,
        true),
    updated_at = now()
FROM cur WHERE ad.id = cur.id;

-- index_plan is intentionally LEFT IN PLACE: defined, unreachable, ready to
-- re-enable. Record why, in the definition itself.
WITH cur AS (
    SELECT id FROM agent_definitions
    WHERE type = 'tool-generator' AND deleted_at IS NULL
    ORDER BY version DESC LIMIT 1
)
UPDATE agent_definitions ad
SET default_config = jsonb_set(
        ad.default_config,
        '{workflow,steps,index_plan,description}',
        to_jsonb('BYPASSED 2026-07-09 (write_plan -> complete): rag_index has no deadline on its per-chunk Ollama embedding call and hung a run for 44+ min. Re-enable by setting write_plan.next_step back to index_plan once the embedding path has a context timeout.'::text),
        true),
    updated_at = now()
FROM cur WHERE ad.id = cur.id;

DO $$
DECLARE n int;
BEGIN
    SELECT count(*) INTO n FROM agent_definitions
    WHERE type = 'tool-generator' AND deleted_at IS NULL
      AND default_config #>> '{workflow,steps,write_plan,next_step}' = 'complete'
      AND default_config #>> '{workflow,steps,index_plan,action}' = 'rag_index'
      AND default_config #>> '{workflow,steps,save_tool,next_step}' = 'compose_plan';
    IF n <> 1 THEN RAISE EXCEPTION 'index_plan bypass incomplete (found %)', n; END IF;
END $$;

COMMIT;

-- Verify:
--   SELECT default_config #>> '{workflow,steps,write_plan,next_step}' AS write_next,
--          default_config #>> '{workflow,steps,index_plan,action}'    AS index_step_still_defined
--   FROM agent_definitions WHERE type='tool-generator' AND deleted_at IS NULL
--   ORDER BY version DESC LIMIT 1;   -- expect: complete | rag_index
--
-- Re-enable after the timeout ships:
--   SELECT snapshot_agent('tool-generator','reenable index_plan');
--   UPDATE agent_definitions SET default_config =
--     jsonb_set(default_config,'{workflow,steps,write_plan,next_step}','"index_plan"'::jsonb,true),
--     updated_at = now()
--   WHERE type='tool-generator' AND deleted_at IS NULL;
--
-- The stuck orchestration (46cd5299) is not touched here. Let the idle reaper
-- (3600s) close it, or inspect first:
--   SELECT status, current_step, EXTRACT(EPOCH FROM (NOW()-last_activity))::int AS since_s
--   FROM orchestration_states WHERE correlation_id='1923badd-870c-48ad-98e0-bc18297e8579'::uuid;
