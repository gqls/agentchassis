-- NNN_seed_index_orchestrator.sql  (v2 — corrected against the REAL schema)
--
-- v1 FAILED: `column "name" of relation "agent_definitions" does not exist` —
-- the column list was written from memory despite the preamble mandating \d
-- first. Corrected from the user's \d paste (2026-07-02):
--   * there is NO `name` column — `display_name` (varchar 255, NOT NULL) only;
--   * `category` (varchar 50) is NOT NULL with NO default — must be supplied;
--     it is DISTINCT from `agent_category` (text, nullable, CHECK
--     strategist|executor|analyst|integrator|coordinator|specialist);
--   * `status` (text, default 'experimental', CHECK active|experimental|…);
--   * `version` integer (default 1), UNIQUE (type, version);
--   * snapshot_agent writes into THIS table (is_snapshot), so the source-row
--     SELECT filters is_snapshot AND deleted_at.
-- Everything else has defaults (image_*, topics, resources, health_config…).
--
-- PURPOSE (unchanged, §7B.1): run 6dfa37cd proved orchestrate+agent_type is
-- ADOPTED IN-PLACE on the chassis pod, which never holds GITHUB_READ_TOKEN.
-- Wrap the indexer in the proven diagnose-orchestrator spawn pattern so the
-- spawned agent-code-indexer pod receives the secretKeyRef (gate line live on
-- image 7c65fbdf64). category/agent_category/status are COPIED from the live
-- diagnose-orchestrator row rather than guessed.

BEGIN;

INSERT INTO agent_definitions
  (type, display_name, description, category, agent_category, status, version, default_config)
SELECT
  'index-orchestrator',
  'Index Orchestrator',
  'Spawns the code-indexer as its own pod (isRepoCloningAgent injects GITHUB_READ_TOKEN) and forwards owner/repo/ref/language; returns the index result. §7B.1 of RUNBOOK_code_retrieval_route.md.',
  d.category,
  d.agent_category,
  d.status,
  1,
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
FROM agent_definitions d
WHERE d.type = 'diagnose-orchestrator'
  AND COALESCE(d.is_snapshot, false) = false
  AND d.deleted_at IS NULL
ORDER BY d.version DESC
LIMIT 1;

-- verify — expect ONE row: category/agent_category/status copied, version 1,
-- workflow starting at spawn_indexer with the four-field input_mapping
SELECT type, display_name, category, agent_category, status, version,
       default_config #> '{workflow,start_step}'                                  AS start_step,
       default_config #> '{workflow,steps,call_indexer,config,input_mapping}'     AS input_mapping
FROM agent_definitions
WHERE type = 'index-orchestrator' AND deleted_at IS NULL;

COMMIT;

-- ── REVERT ────────────────────────────────────────────────────────────────
-- BEGIN;
-- DELETE FROM agent_definitions WHERE type = 'index-orchestrator';
-- COMMIT;
