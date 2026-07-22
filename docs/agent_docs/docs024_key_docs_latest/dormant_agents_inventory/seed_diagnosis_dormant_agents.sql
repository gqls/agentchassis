-- seed_diagnosis_dormant_agents.sql — DORMANT-AGENTS capability inventory sweep.
-- bugs_open/044. 2026-07-21. Applies to clients_db.
--
-- ██ DEPLOY SEQUENCING ██ — apply ONLY AFTER a chassis image carrying the
-- diagnose_dormant_agents action is live. Verify in-pod (chassis pod name):
--   kubectl -n ai-persona-system exec <chassis-pod> -- \
--     sh -c 'grep -ac diagnose_dormant_agents /proc/1/exe'   (must be >= 1)
-- A seed naming an unregistered action fails at runtime (image first, then seed).
--
-- WHAT THIS IS (bugs_open/044). The deterministic capability inventory (no LLM):
-- the producer-side mirror of the silent-check checker. silent-check finds work
-- no worker touches; this finds a WORKER no work routes to — an active agent
-- whose workflow has never been observed running. Method: the STEP FINGERPRINT
-- (a workflow step key unique to exactly one agent), looked for among the
-- top-level step keys ever seen in a real orchestration_states.workflow_plan.
-- owner_agent_type is deliberately NOT used (95k+ rows carry 'generic').
--   * Age floor (default 14d): a freshly-seeded agent that simply has not run
--     yet is reported but not emitted.
--   * Mirrored-agent blind spot (agents with no unique step, e.g. council-gate):
--     listed, NEVER flagged — unmeasurable by this method.
-- Findings emit as INERT dormant_agent items (status='dormant',
-- pipeline='maintenance', unclaimable, one per agent, self-deduped) anchored to
-- the system.internal pseudo-site; a human triages wire / retire / paused.
-- Items are closed honestly once the agent is observed to run.
-- One doc_note per sweep (categories dormant-agents+fixloop).
--
-- ██ SHIPS IN DRY_RUN ██ — dry_run=true: report only; NO items written, none
-- closed. Review the report, then flip dry_run to false (jsonb_set below).
-- Owner: MANUAL trigger for now (more awareness before autonomy).
--
-- Read the latest dormant-agents report:
--   SELECT body FROM doc_notes WHERE categories ? 'dormant-agents'
--   ORDER BY created_at DESC LIMIT 1;

BEGIN;

SELECT snapshot_agent('diagnosis-dormant-agents', 'pre-update: dormant-agents v1 re-apply')
WHERE EXISTS (SELECT 1 FROM agent_definitions
              WHERE type='diagnosis-dormant-agents' AND is_active
                AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL);

INSERT INTO agent_definitions (
    type, display_name, description, category, agent_category, status,
    is_active, version, capabilities,
    image_repository, image_tag, command, resources, topics, health_config, env_vars,
    default_config
)
SELECT
    'diagnosis-dormant-agents',
    'Diagnosis Dormant-Agents (capability inventory)',
    'Deterministic capability inventory (no LLM, bugs_open/044): active agents whose workflow has never been observed running (step-fingerprint method; owner_agent_type deliberately unused), age-floored, mirrored-agent blind spot never flagged. Emits inert dormant_agent items for human triage; closes them once the agent runs. Ships dry_run. Manual for now.',
    'diagnose', 'coordinator', 'experimental',
    true, 1, '["diagnose", "inventory"]'::jsonb,
    d.image_repository, d.image_tag, d.command, d.resources, d.topics, d.health_config, d.env_vars,
    jsonb_build_object('workflow', jsonb_build_object(
      'start_step', 'sweep',
      'processing_mode', 'task',
      'timeout_seconds', 120,
      'steps', jsonb_build_object(

        'sweep', jsonb_build_object(
          'action', 'diagnose_dormant_agents',
          'description', 'Inventory active agents; emit dormant_agent findings for human triage (unless dry_run); close resolved ones; write the report note.',
          'output_field', 'dormant_agents_result',
          'next_step', 'complete',
          'config', jsonb_build_object(
            'age_floor_days', 14,
            'max_emit', 10,
            'dry_run', true   -- FLIP to false when the preview looks right
          )
        ),

        'complete', jsonb_build_object(
          'action', 'complete_workflow',
          'description', 'Dormant-agents swept; report in doc_notes (categories dormant-agents+fixloop) and in the payload.',
          'config', jsonb_build_object(
            'output_fields', jsonb_build_array('dormant_agents_result'),
            'success_message', 'diagnosis-dormant-agents swept; report persisted to doc_notes (categories dormant-agents+fixloop)'
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

-- ██ DO NOT FLIP dry_run OFF YET ██ (added 2026-07-22 — bugs_open/060).
-- orchestration_states is pruned hourly at 24h by the database-cleanup task, so
-- the observation window is ~1-2 days — far shorter than the 14d age floor.
-- Under a short window a still-ACTIVE agent that merely has not run today reads
-- as never-observed (fix-proposer is the canonical false positive). The
-- guarded build (> v1.0.1149) REFUSES to emit while the window < the age floor;
-- the CURRENTLY-LIVE binary (v1.0.1149) does NOT have that guard, so flipping
-- dry_run off now would write false-positive items. Keep dry_run=true until:
--   (a) a chassis image carrying the WINDOW GUARD is live (verify in-pod:
--       strings show the "WINDOW TOO SHORT" banner literal), AND
--   (b) bugs_open/060 gives a durable run signal (usage_count is dead), OR the
--       window genuinely exceeds the age floor.
-- Flip (only when the above hold):
--   UPDATE agent_definitions
--   SET default_config = jsonb_set(default_config,
--         '{workflow,steps,sweep,config,dry_run}', 'false'::jsonb), updated_at=now()
--   WHERE type='diagnosis-dormant-agents' AND is_active AND COALESCE(is_snapshot,false)=false;
--
-- Raise the emit cap once the report has been reviewed (default 10; ~70 findings
-- are past the age floor, so coverage is capped, not complete, at the default):
--   UPDATE agent_definitions
--   SET default_config = jsonb_set(default_config,
--         '{workflow,steps,sweep,config,max_emit}', '80'::jsonb), updated_at=now()
--   WHERE type='diagnosis-dormant-agents' AND is_active AND COALESCE(is_snapshot,false)=false;
--
-- Rollback: restore the pre-update snapshot from agent_definitions_backup, or
-- DELETE FROM agent_definitions WHERE type='diagnosis-dormant-agents' AND version=1;
