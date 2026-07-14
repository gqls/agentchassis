-- 0NN_diagnosis_silent_check.sql — SILENT-CHECK verification checker (Phase 2). 2026-07-14.
-- Renumber 0NN when filing. Applies to clients_db.
--
-- ██ DEPLOY SEQUENCING ██ — apply ONLY AFTER a chassis image carrying the
-- diagnose_silent_check action (> v1.0.1117) is live. Verify in-pod:
--   grep -ac diagnose_silent_check /proc/1/exe   (must be >= 1)
--
-- WHAT THIS IS (DESIGN_triage_and_escalation.md §"the hard case"; scope
-- reconciled 2026-07-14 with the empty-sections thread). The deterministic
-- verification checker (no LLM) for the defect class NO work item ever
-- touches: structural invariants violated in observable state while
-- site_work_items stays silent (the darts guides-index class).
--   * Check nav_linked_never_built (EMITS): page in the site's navigation,
--     build_status='planned' beyond the grace period, no covering work item.
--   * Check deployed_zero_components (REPORT-ONLY in v1): built/deployed page
--     serving zero components — may be a deliberate content removal, so it
--     stays on the report until the owner promotes it via emit_checks.
-- Findings emit as INERT silent_failure items (status='failed', unclaimable,
-- one per check+site, self-deduped); the EXISTING triage sweep routes them to
-- the loop under its own cap. Resolved findings are closed honestly.
-- One doc_note per sweep (categories silent-check+fixloop).
--
-- ██ SHIPS IN DRY_RUN ██ — dry_run=true: report only; NO items written, none
-- closed. Review the report, then flip dry_run to false (jsonb_set below).
-- Owner: MANUAL trigger for now (more awareness before autonomy).
--
-- Read the latest silent-check report:
--   SELECT body FROM doc_notes WHERE categories ? 'silent-check'
--   ORDER BY created_at DESC LIMIT 1;

BEGIN;

SELECT snapshot_agent('diagnosis-silent-check', 'pre-update: silent-check v1 re-apply')
WHERE EXISTS (SELECT 1 FROM agent_definitions
              WHERE type='diagnosis-silent-check' AND is_active
                AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL);

INSERT INTO agent_definitions (
    type, display_name, description, category, agent_category, status,
    is_active, version, capabilities,
    image_repository, image_tag, command, resources, topics, health_config, env_vars,
    default_config
)
SELECT
    'diagnosis-silent-check',
    'Diagnosis Silent-Check (verification checker)',
    'Deterministic verification checker (no LLM): structural invariants violated with NO covering work item (nav-linked pages never built; deployed pages with zero components report-only). Emits inert silent_failure items for triage; closes resolved ones. Ships dry_run. Manual for now.',
    'diagnose', 'coordinator', 'experimental',
    true, 1, '["diagnose", "verify"]'::jsonb,
    d.image_repository, d.image_tag, d.command, d.resources, d.topics, d.health_config, d.env_vars,
    jsonb_build_object('workflow', jsonb_build_object(
      'start_step', 'sweep',
      'processing_mode', 'task',
      'timeout_seconds', 120,
      'steps', jsonb_build_object(

        'sweep', jsonb_build_object(
          'action', 'diagnose_silent_check',
          'description', 'Verify structural invariants; emit silent_failure findings for triage (unless dry_run); close resolved ones; write the report note.',
          'output_field', 'silent_check_result',
          'next_step', 'complete',
          'config', jsonb_build_object(
            'grace_hours', 48,
            'settle_hours', 2,
            'max_emit', 3,
            'dry_run', true   -- FLIP to false when the preview looks right
          )
        ),

        'complete', jsonb_build_object(
          'action', 'complete_workflow',
          'description', 'Silent-check swept; report in doc_notes (categories silent-check+fixloop) and in the payload.',
          'config', jsonb_build_object(
            'output_fields', jsonb_build_array('silent_check_result'),
            'success_message', 'diagnosis-silent-check swept; report persisted to doc_notes (categories silent-check+fixloop)'
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
--   WHERE type='diagnosis-silent-check' AND is_active AND COALESCE(is_snapshot,false)=false;
--
-- Promote check 2 to an emitter (owner decision, only after its report-only
-- findings have been reviewed for deliberate removals):
--   UPDATE agent_definitions
--   SET default_config = jsonb_set(default_config,
--         '{workflow,steps,sweep,config,emit_checks}',
--         '["nav_linked_never_built","deployed_zero_components"]'::jsonb), updated_at=now()
--   WHERE type='diagnosis-silent-check' AND is_active AND COALESCE(is_snapshot,false)=false;
--
-- Rollback: restore the pre-update snapshot from agent_definitions_backup, or
-- DELETE FROM agent_definitions WHERE type='diagnosis-silent-check' AND version=1;
