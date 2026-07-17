-- 0NN_feature_implementer_orchestrator.sql — feature builder delta 2: run the
-- stage-loop implementer as a DEDICATED POD (v1 DRAFT). 2026-07-17. Applies to
-- clients_db. Renumber 0NN when filing.
--
-- ██ STATUS: DRAFT, NOT APPLIED ██ — ships in the repo per the seed discipline.
-- Apply together with 0NN_feature_implementer.sql, AFTER the image carrying
-- feature-implementer on the isRepoCloningAgent gate (commit c19b5d097+).
--
-- WHY (the fix loop's proven lesson, 2026-07-13): fired via the generic
-- orchestrate path, the implementer runs IN the shared chassis pod, where the
-- spawn gate never fires and GITHUB_READ_TOKEN is (by explicit prior decision)
-- absent — the first fix-implementer run died exactly there. This wrapper
-- spawns feature-implementer as its OWN pod so the gate injects the READ token
-- for the pod's lifetime; WRITES still go only through the git-adapter.
--
-- Input: {"fix_correlation_id": "<correlation of an APPROVED staged plan>",
--         "base_ref": "<optional; default main>"} — forwarded verbatim.

BEGIN;

SELECT snapshot_agent('feature-implementer-orchestrator', 'pre-update: feature builder delta 2 wrapper re-apply')
WHERE EXISTS (SELECT 1 FROM agent_definitions
              WHERE type='feature-implementer-orchestrator' AND is_active
                AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL);

INSERT INTO agent_definitions (
    type, display_name, description, category, agent_category, status,
    is_active, version, capabilities,
    image_repository, image_tag, command, resources, topics, health_config, env_vars,
    default_config
)
SELECT
    'feature-implementer-orchestrator',
    'Feature Implementer Orchestrator (FB delta 2)',
    'Spawns feature-implementer as a DEDICATED pod (so the isRepoCloningAgent gate injects the read token) and forwards fix_correlation_id + base_ref. Mirrors fix-implementer-orchestrator. Keeps repo reads off the shared chassis pod; writes still go via the git-adapter.',
    'diagnose', 'coordinator', 'experimental',
    true, 1, '["diagnose", "feature-implementation"]'::jsonb,
    d.image_repository, d.image_tag, d.command, d.resources, d.topics, d.health_config, d.env_vars,
    jsonb_build_object('workflow', jsonb_build_object(
      'start_step', 'spawn_implementer',
      'processing_mode', 'orchestrator',
      'timeout_seconds', 3800,
      'steps', jsonb_build_object(

        'spawn_implementer', jsonb_build_object(
          'action', 'spawn_agent',
          'description', 'Spawn a dedicated feature-implementer pod — the spawn gate injects GITHUB_READ_TOKEN into it (never the chassis).',
          'next_step', 'call_implementer',
          'config', jsonb_build_object(
            'role', 'implementer',
            'agent_type', 'feature-implementer'
          )
        ),

        'call_implementer', jsonb_build_object(
          'action', 'call_agent',
          'description', 'Hand the approved staged-plan correlation (+ optional base_ref) to the spawned implementer and await its PR result.',
          'next_step', 'complete',
          'config', jsonb_build_object(
            'agent_type', 'feature-implementer',
            'target_role', 'implementer',
            'timeout_seconds', 3600,
            'input_mapping', jsonb_build_object(
              'fix_correlation_id', 'input_data.fix_correlation_id',
              'base_ref', 'input_data.base_ref'
            )
          )
        ),

        'complete', jsonb_build_object(
          'action', 'complete_workflow',
          'description', 'Forward the implementer result (pr_result / gate / stage).',
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
-- DELETE FROM agent_definitions WHERE type='feature-implementer-orchestrator' AND version=1;
