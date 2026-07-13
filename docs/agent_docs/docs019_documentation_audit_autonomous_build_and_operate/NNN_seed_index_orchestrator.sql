-- NNN_seed_index_orchestrator.sql
--
-- §7B.1 (RUNBOOK_code_retrieval_route.md): run 6dfa37cd proved the code-indexer
-- workflow is NOT spawned — `orchestrate` + agent_type=code-indexer is ADOPTED
-- IN-PLACE by the generic chassis pod (log: agent_type="generic", pod
-- agent-chassis-7c65fbdf64-69wr9 executing request_analysis). The chassis pod,
-- by design, never holds GITHUB_READ_TOKEN, so analyse_repo_local fails there.
-- isRepoCloningAgent (now including "code-indexer", image 7c65fbdf64) is
-- consulted at SPAWN time only — dormant until something spawns the indexer.
--
-- FIX: wrap the indexer in a spawning orchestrator, mirroring the PROVEN
-- diagnose-orchestrator pattern verbatim (spawn_agent -> call_agent ->
-- complete; workflow JSON mirrored from this session's live dump of
-- diagnose-orchestrator). The spawned agent-code-indexer-<id> pod then
-- receives the secretKeyRef via the gate, loads code-indexer's (§7B-migrated)
-- default_config, and runs analyse_repo_local with the token present.
--
-- Alternatives considered and REJECTED:
--   (B) put the token on the chassis deployment — forbidden by the token
--       design ("the spawning chassis pod never holds it").
--   (C) revert §7B to the adapter path (its code was exonerated; an explicit
--       ref WOULD fetch the right tree) — but the adapter returns the FULL
--       analysis Output over Kafka, and at 572 files that reply is multi-MB
--       against Kafka's ~1MB default max message; fine at 69 files in §6D,
--       a probable new failure now. In-process stays structurally right.
--   (D) a bare k8s Job — a parallel mechanism outside "every agent is an
--       orchestrator"; loses the message/trace logging the agents give.
--
-- SCHEMA RULE: before applying, confirm the column list with `\d
-- agent_definitions`. The INSERT below names only columns used throughout this
-- project (type, name, display_name, version, agent_category, status,
-- default_config) and COPIES version/agent_category/status from the proven
-- diagnose-orchestrator row rather than guessing values/types. If \d shows
-- additional NOT NULL columns without defaults, extend the SELECT accordingly.
-- New type ⇒ nothing to snapshot (snapshot_agent is for UPDATEs).

BEGIN;

INSERT INTO agent_definitions
  (type, name, display_name, version, agent_category, status, default_config)
SELECT
  'index-orchestrator',
  'Index Orchestrator',
  'Index Orchestrator',
  version,
  agent_category,
  status,
  '{
    "workflow": {
      "start_step": "spawn_indexer",
      "steps": {
        "spawn_indexer": {
          "action": "spawn_agent",
          "config": { "agent_type": "code-indexer", "role": "indexer" },
          "next_step": "call_indexer",
          "description": "Spawn the code-indexer as its own pod (isRepoCloningAgent injects GITHUB_READ_TOKEN via secretKeyRef)"
        },
        "call_indexer": {
          "action": "call_agent",
          "config": {
            "agent_type": "code-indexer",
            "target_role": "indexer",
            "timeout_seconds": 1800,
            "input_mapping": {
              "owner?": "input_data.owner",
              "repo?": "input_data.repo",
              "ref?": "input_data.ref",
              "language?": "input_data.language"
            }
          },
          "next_step": "complete",
          "description": "Forward owner/repo/ref/language to the spawned indexer; await the index result"
        },
        "complete": {
          "action": "complete_workflow",
          "config": { "result_from": "code-indexer_result" },
          "description": "Return the index result to the caller"
        }
      }
    },
    "processing_mode": "orchestrator",
    "timeout_seconds": 1800
  }'::jsonb
FROM agent_definitions
WHERE type = 'diagnose-orchestrator';

-- verify — expect one row, workflow starting at spawn_indexer
SELECT type, name, version, agent_category, status,
       default_config #> '{workflow,start_step}' AS start_step,
       default_config #> '{workflow,steps,call_indexer,config,input_mapping}' AS input_mapping
FROM agent_definitions
WHERE type = 'index-orchestrator';

COMMIT;

-- ── REVERT ────────────────────────────────────────────────────────────────
-- BEGIN;
-- DELETE FROM agent_definitions WHERE type = 'index-orchestrator';
-- COMMIT;
