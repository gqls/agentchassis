-- 0NN_fix_implementer_orchestrator.sql — F1.1b(c) fix: run the implementer as a
-- DEDICATED POD, not in-chassis. 2026-07-13. Applies to clients_db.
--
-- WHY. The first end-to-end run failed at read_current_files with
-- "GITHUB_READ_TOKEN not in env": fix-implementer was fired via the generic
-- orchestrate path and ran IN the shared chassis pod, so the isRepoCloningAgent
-- spawn gate (which injects the read token into a DEDICATED pod) never fired.
-- An explicit prior decision forbids the chassis pod from holding the token.
--
-- FIX (owner decision 2026-07-13: "dedicated implementer pod that uses the
-- git-adapter"). This wrapper spawns fix-implementer as its OWN k8s Job —
-- exactly as diagnose-orchestrator spawns diagnose-agent — so the gate injects
-- GITHUB_READ_TOKEN into the ephemeral implementer pod (reaped after). Reads use
-- that gated token in-pod; WRITES still go through the git-adapter
-- (git_adapter_request). The chassis never holds a GitHub token; the implementer
-- pod holds only the READ token, only for its lifetime.
--
-- No image rebuild: isRepoCloningAgent already lists fix-implementer (deployed
-- in v1.0.1110); this only adds the wrapper agent + retargets the trigger
-- (092 now targets fix-implementer-orchestrator).
--
-- Input: {"fix_correlation_id": "<correlation of an APPROVED plan>"} — forwarded
-- verbatim to the spawned fix-implementer.

BEGIN;

SELECT snapshot_agent('fix-implementer-orchestrator', 'pre-update: F1.1b(c) wrapper re-apply')
WHERE EXISTS (SELECT 1 FROM agent_definitions
              WHERE type='fix-implementer-orchestrator' AND is_active
                AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL);

INSERT INTO agent_definitions (
    type, display_name, description, category, agent_category, status,
    is_active, version, capabilities,
    image_repository, image_tag, command, resources, topics, health_config, env_vars,
    default_config
)
SELECT
    'fix-implementer-orchestrator',
    'Fix Implementer Orchestrator (F1.1b(c))',
    'Spawns fix-implementer as a DEDICATED pod (so the isRepoCloningAgent gate injects the read token) and forwards the fix_correlation_id. Mirrors diagnose-orchestrator -> diagnose-agent. Keeps repo reads off the shared chassis pod; writes still go via the git-adapter.',
    'diagnose', 'coordinator', 'experimental',
    true, 1, '["diagnose", "fix-implementation"]'::jsonb,
    d.image_repository, d.image_tag, d.command, d.resources, d.topics, d.health_config, d.env_vars,
    jsonb_build_object('workflow', jsonb_build_object(
      'start_step', 'spawn_implementer',
      'processing_mode', 'orchestrator',
      'timeout_seconds', 2000,
      'steps', jsonb_build_object(

        'spawn_implementer', jsonb_build_object(
          'action', 'spawn_agent',
          'description', 'Spawn a dedicated fix-implementer pod — the spawn gate injects GITHUB_READ_TOKEN into it (never the chassis).',
          'next_step', 'call_implementer',
          'config', jsonb_build_object(
            'role', 'implementer',
            'agent_type', 'fix-implementer'
          )
        ),

        'call_implementer', jsonb_build_object(
          'action', 'call_agent',
          'description', 'Hand the approved-plan correlation to the spawned implementer and await its PR result.',
          'next_step', 'complete',
          'config', jsonb_build_object(
            'agent_type', 'fix-implementer',
            'target_role', 'implementer',
            'timeout_seconds', 1800,
            'input_mapping', jsonb_build_object(
              'fix_correlation_id', 'input_data.fix_correlation_id'
            )
          )
        ),

        'complete', jsonb_build_object(
          'action', 'complete_workflow',
          'description', 'Forward the implementer result (pr_result / gate / commit_prep).',
          'config', jsonb_build_object('result_from', 'call_implementer')
        )
      )
    ))
FROM agent_definitions d
WHERE d.type = 'diagnose-orchestrator'
  AND COALESCE(d.is_snapshot, false) = false AND d.deleted_at IS NULL
ON CONFLICT (type, version) DO UPDATE
   SET default_config = EXCLUDED.default_config,
       description    = EXCLUDED.description,
       updated_at     = now();

COMMIT;

-- Rollback: restore the pre-update snapshot from agent_definitions_backup, or
-- DELETE FROM agent_definitions WHERE type='fix-implementer-orchestrator' AND version=1;
