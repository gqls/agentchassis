-- 0NN_reenable_index_plan.sql — undo the index_plan bypass, once the embedding
-- deadline has SHIPPED. DRAFT 2026-07-09. Renumber 0NN. DB-only; REVERSIBLE.
--
-- PRECONDITION (do NOT apply before this is true): the chassis binary running
-- the shared `generic` pod must include the rag_index per-chunk embedding
-- deadline — platform/orchestration/actions/rag_actions.go now wraps
-- GenerateEmbedding in context.WithTimeout(ctx, embTimeout) (default 120s,
-- matching the OllamaClient http.Client cap — ollama is slow, so defaults do
-- not tighten anything; config key embedding_timeout_seconds can lower it
-- per-step), so a stalled ollama-adapter degrades into the existing non-fatal
-- "store without embeddings" path instead of freezing the step. This is the
-- structural fix named in 0NN_bypass_index_plan_until_embed_timeout.sql
-- (§STRUCTURAL FIX). NOTE the 016b v8 correction: the 44-min "hang" was an
-- OOMKilled pod, not a missing deadline — this deadline is hygiene, and the
-- worst case per chunk was already 120s via the http client.
--
-- Verify the deadline is live before applying, e.g. confirm the running image
-- tag matches the build that carries the rag_actions.go change.
--
-- Effect: tool-generator runs save_tool -> compose_plan -> write_plan ->
-- index_plan -> complete again (PLAN also indexed into knowledge_base tool_docs).

BEGIN;

SELECT snapshot_agent('tool-generator', '0NN_reenable_index_plan.sql: pre-update');

WITH cur AS (
    SELECT id FROM agent_definitions
    WHERE type = 'tool-generator' AND deleted_at IS NULL
    ORDER BY version DESC LIMIT 1
)
UPDATE agent_definitions ad
SET default_config = jsonb_set(
        jsonb_set(
            ad.default_config,
            '{workflow,steps,write_plan,next_step}',
            '"index_plan"'::jsonb,
            true),
        -- refresh the description so it no longer reads as BYPASSED
        '{workflow,steps,index_plan,description}',
        to_jsonb('Index the PLAN into knowledge_base tool_docs. Re-enabled once rag_index gained a per-chunk embedding deadline (context.WithTimeout, default 15s) so an unresponsive embedder degrades to store-without-embeddings rather than hanging the step.'::text),
        true),
    updated_at = now()
FROM cur WHERE ad.id = cur.id;

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
--   SELECT default_config #>> '{workflow,steps,write_plan,next_step}' AS write_next,
--          default_config #>> '{workflow,steps,index_plan,action}'    AS index_action
--   FROM agent_definitions WHERE type='tool-generator' AND deleted_at IS NULL
--   ORDER BY version DESC LIMIT 1;   -- expect: index_plan | rag_index
--
-- Proof (next organic tool creation): a knowledge_base row lands for the PLAN
-- without the run hanging —
--   SELECT collection, count(*) FROM knowledge_base
--   WHERE collection = 'tool_docs' GROUP BY 1;
--
-- Rollback: restore from the snapshot, or re-run the bypass
--   (write_plan.next_step -> 'complete').
