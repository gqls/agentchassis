-- NNN_move_diagnose_workflow_to_default_config.sql
-- DON'T USE THIS
-- Renumber NNN to the next number in your migration sequence (3-digit prefixes).
--
-- FIX: the seeded diagnose agents (diagnose-agent, diagnose-orchestrator, created
-- 2026-06-20) carry their workflow in `orchestration_workflow` (the `json` column),
-- but every WORKING agent (build-dispatch-loop, site-work-orchestrator,
-- domain-research-classifier, ...) carries it in `default_config` (jsonb). The
-- workflow loader reads default_config; a workflow left in orchestration_workflow
-- will not load. This moves it.
--
-- It copies the column across verbatim (no re-typing — preserves the exact seeded
-- workflow), nulls orchestration_workflow, and merges in the two top-level keys the
-- working agents have that the seeded workflow lacked: processing_mode +
-- timeout_seconds. (The seeded orchestration_workflow held ONLY {"workflow":{...}};
-- the reference agents put processing_mode/timeout_seconds at default_config top
-- level — e.g. site-work-orchestrator. Without processing_mode the agent may not
-- run as an orchestrator.)
--
-- Schema (verified from \d agent_definitions):
--   default_config jsonb NOT NULL DEFAULT '{}'::jsonb
--   orchestration_workflow json (nullable)
-- The cast orchestration_workflow::jsonb (json -> jsonb) is valid; `||` merges
-- top-level keys.
--
-- Idempotent + non-clobbering: runs only where orchestration_workflow IS NOT NULL
-- (so a second run is a no-op) and default_config = '{}' (so it never overwrites a
-- default_config that already holds config). default_config is NOT NULL, so the
-- guards also stop it ever writing NULL.

-- diagnose-agent (the worker that runs the loop)
UPDATE agent_definitions
SET default_config = orchestration_workflow::jsonb
      || jsonb_build_object('processing_mode', 'orchestrator', 'timeout_seconds', 1800),
    orchestration_workflow = NULL,
    updated_at = now()
WHERE type = 'diagnose-agent'
  AND orchestration_workflow IS NOT NULL
  AND default_config = '{}'::jsonb;

-- diagnose-orchestrator (the thin spawn-and-forward wrapper)
UPDATE agent_definitions
SET default_config = orchestration_workflow::jsonb
      || jsonb_build_object('processing_mode', 'orchestrator', 'timeout_seconds', 600),
    orchestration_workflow = NULL,
    updated_at = now()
WHERE type = 'diagnose-orchestrator'
  AND orchestration_workflow IS NOT NULL
  AND default_config = '{}'::jsonb;

-- Verify (expect has_workflow = t, orch_wf_null = t for both):
--   SELECT type,
--          (default_config ? 'workflow')          AS has_workflow,
--          (default_config ->> 'processing_mode')  AS processing_mode,
--          (orchestration_workflow IS NULL)        AS orch_wf_null
--   FROM agent_definitions
--   WHERE type IN ('diagnose-agent','diagnose-orchestrator');

-- FLAG (not fixed here — confirm, then a follow-up if needed):
--  1. ai_service: the moved default_config has NO ai_service block. If diagnose_run's
--     verdict needs agent-level model config (vs the verdict_prompt_ref 'diagnose-verdict-v1'
--     carrying it), add it — site-adoption-agent shows the shape:
--       default_config || '{"ai_service":{"model":"claude-sonnet-4-6","provider":"anthropic","api_key_env_var":"ANTHROPIC_API_KEY"}}'::jsonb
--  2. timeout alignment: diagnose-orchestrator's call_diagnoser step has timeout_seconds
--     600; the diagnose-agent loop (<=5 iterations, each a model verdict + gather) may
--     exceed that. If it times out, raise BOTH the orchestrator's call_diagnoser
--     timeout and the agent's timeout_seconds together.
