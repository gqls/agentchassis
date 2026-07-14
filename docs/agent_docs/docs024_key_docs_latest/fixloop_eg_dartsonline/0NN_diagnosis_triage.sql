-- 0NN_diagnosis_triage.sql — TRIAGE router (Phase 1). 2026-07-14.
-- Renumber 0NN when filing. Applies to clients_db.
--
-- ██ DEPLOY SEQUENCING ██ — apply ONLY AFTER a chassis image carrying the
-- diagnose_triage action (> v1.0.1114) is live. Verify in-pod:
--   grep -ac diagnose_triage /proc/1/exe   (must be >= 1)
--
-- WHAT THIS IS (DESIGN_triage_and_escalation.md; owner decisions 2026-07-14).
-- The deterministic router (no LLM) from the operational immune system into the
-- fix loop. Phase 1 handles the two failure flavours already present in the data:
--   * LOUD FAILURE (status='failed') — escalate the PATTERN to needs_diagnosis
--     (deduped by (item_type, handler, error signature); capped per sweep;
--     parked at awaiting_diagnosis, INERT until a human/dispatch claims it).
--   * NO HANDLER YET (item_type='capability_gap') — surfaced to the roadmap in
--     the report; NEVER escalated to the loop.
-- One doc_note per sweep (categories triage+fixloop) is the readable artifact.
--
-- ██ SHIPS IN DRY_RUN ██ — dry_run=true: the sweep previews what it WOULD
-- escalate and writes the report, but creates NO work items. Review the report,
-- then flip dry_run to false (jsonb_set below) to let it escalate for real.
-- Owner: MANUAL trigger for now; cadence hourly-later when trusted.
--
-- Read the latest triage report:
--   SELECT body FROM doc_notes WHERE categories ? 'triage'
--   ORDER BY created_at DESC LIMIT 1;

BEGIN;

SELECT snapshot_agent('diagnosis-triage', 'pre-update: triage router v1 re-apply')
WHERE EXISTS (SELECT 1 FROM agent_definitions
              WHERE type='diagnosis-triage' AND is_active
                AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL);

INSERT INTO agent_definitions (
    type, display_name, description, category, agent_category, status,
    is_active, version, capabilities,
    image_repository, image_tag, command, resources, topics, health_config, env_vars,
    default_config
)
SELECT
    'diagnosis-triage',
    'Diagnosis Triage (router)',
    'Deterministic router (no LLM): scans site_work_items, escalates loud-failure PATTERNS (deduped, capped) to needs_diagnosis, surfaces capability_gap items to the roadmap. Ships dry_run. Manual for now.',
    'diagnose', 'coordinator', 'experimental',
    true, 1, '["diagnose", "triage"]'::jsonb,
    d.image_repository, d.image_tag, d.command, d.resources, d.topics, d.health_config, d.env_vars,
    jsonb_build_object('workflow', jsonb_build_object(
      'start_step', 'sweep',
      'processing_mode', 'task',
      'timeout_seconds', 120,
      'steps', jsonb_build_object(

        'sweep', jsonb_build_object(
          'action', 'diagnose_triage',
          'description', 'Scan failures + capability gaps; escalate deduped patterns (unless dry_run); write the report note.',
          'output_field', 'triage_result',
          'next_step', 'complete',
          'config', jsonb_build_object(
            'window_hours', 336,
            'max_escalations', 3,
            'diagnose_handler', 'diagnose-orchestrator',
            'repo_owner', 'gqls',
            'repo_name', 'agentchassis',
            'ref', 'main',
            'dry_run', true   -- FLIP to false when the preview looks right
          )
        ),

        'complete', jsonb_build_object(
          'action', 'complete_workflow',
          'description', 'Triage swept; report in doc_notes (categories triage+fixloop) and in the payload.',
          'config', jsonb_build_object(
            'output_fields', jsonb_build_array('triage_result'),
            'success_message', 'diagnosis-triage swept; report persisted to doc_notes (categories triage+fixloop)'
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

-- Flip out of dry-run once the preview looks right:
--   UPDATE agent_definitions
--   SET default_config = jsonb_set(default_config,
--         '{workflow,steps,sweep,config,dry_run}', 'false'::jsonb), updated_at=now()
--   WHERE type='diagnosis-triage' AND is_active AND COALESCE(is_snapshot,false)=false;
--
-- Rollback: restore the pre-update snapshot from agent_definitions_backup, or
-- DELETE FROM agent_definitions WHERE type='diagnosis-triage' AND version=1;
