-- 0NN_fixloop_digest.sql — the AWARENESS SURFACE agent (v1). 2026-07-13.
-- Renumber 0NN when filing. Applies to clients_db.
--
-- ██ DEPLOY SEQUENCING ██ — apply ONLY AFTER a chassis image carrying the
-- fixloop_digest action (> v1.0.1113) is live. Verify in-pod:
--   grep -ac fixloop_digest /proc/1/exe   (must be >= 1)
--
-- WHAT THIS IS. Owner standing rule (2026-07-12): "more awareness BEFORE wider
-- autonomy". This agent composes a DETERMINISTIC digest (no LLM anywhere in
-- the path — an awareness surface that could hallucinate what the system did
-- would defeat itself) of the fix loop's activity in a window:
--   runs by loop agent type (status, terminal, gate verdicts, PR urls);
--   decisions per correlation (artifact kinds, latest council decision + why);
--   agent-config snapshots (the "what changed about the machine itself" ledger
--   — every seeded/updated agent leaves one with a reason).
-- Persisted to doc_notes (pipeline/diagnose, categories ["digest","fixloop"]).
--
-- Read the latest digest:
--   SELECT body FROM doc_notes WHERE categories ? 'digest'
--   ORDER BY created_at DESC LIMIT 1;
--
-- v1 is MANUAL-TRIGGER ONLY (093_TRIGGER_fixloop_digest_v1.sh). A scheduled
-- cadence (daily via kafka-scheduler) is a deliberate later enablement, once
-- the owner is happy with the content — the shipped-disabled tradition.

BEGIN;

SELECT snapshot_agent('fixloop-digest', 'pre-update: awareness surface v1 re-apply')
WHERE EXISTS (SELECT 1 FROM agent_definitions
              WHERE type='fixloop-digest' AND is_active
                AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL);

INSERT INTO agent_definitions (
    type, display_name, description, category, agent_category, status,
    is_active, version, capabilities,
    image_repository, image_tag, command, resources, topics, health_config, env_vars,
    default_config
)
SELECT
    'fixloop-digest',
    'Fix-loop Digest (awareness surface)',
    'Composes a deterministic digest (no LLM) of fix-loop activity: runs, council decisions, gate/PR outcomes, and agent-config snapshots in a window. Persists to doc_notes (pipeline/diagnose, categories digest+fixloop). Owner rule: more awareness before wider autonomy.',
    'diagnose', 'coordinator', 'experimental',
    true, 1, '["diagnose", "observability"]'::jsonb,
    d.image_repository, d.image_tag, d.command, d.resources, d.topics, d.health_config, d.env_vars,
    jsonb_build_object('workflow', jsonb_build_object(
      'start_step', 'compose_digest',
      'processing_mode', 'task',
      'timeout_seconds', 120,
      'steps', jsonb_build_object(

        'compose_digest', jsonb_build_object(
          'action', 'fixloop_digest',
          'description', 'Gather facts by SQL, render by Go, persist to doc_notes.',
          'output_field', 'digest_result',
          'next_step', 'complete',
          'config', jsonb_build_object(
            'window_hours', 24,
            'subject_type', 'pipeline',
            'subject_key', 'diagnose',
            'persist', true
          )
        ),

        'complete', jsonb_build_object(
          'action', 'complete_workflow',
          'description', 'Digest composed and persisted; body also in the completion payload.',
          'config', jsonb_build_object(
            'output_fields', jsonb_build_array('digest_result'),
            'success_message', 'fix-loop digest persisted to doc_notes (categories digest+fixloop)'
          )
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
-- DELETE FROM agent_definitions WHERE type='fixloop-digest' AND version=1;
