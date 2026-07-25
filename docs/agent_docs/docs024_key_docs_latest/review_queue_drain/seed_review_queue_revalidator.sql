-- seed_review_queue_revalidator.sql — the needs_human_review DRAIN.
-- bugs_open/033 (human_review_queue_has_no_working_surface). 2026-07-25.
-- Applies to clients_db.
--
-- ██ DEPLOY SEQUENCING ██ — apply ONLY AFTER a chassis image carrying the
-- revalidate_review_queue action is live. Verify in-pod:
--   kubectl -n ai-persona-system exec <chassis-pod> -- \
--     sh -c 'strings /app/agent-chassis | grep -c "auto:revalidated"'   (>= 1)
-- Grep a string the CHANGE CREATED ('auto:revalidated'), not one it merely
-- uses — an action name that already appears in an older binary proves nothing.
-- A seed naming an unregistered action fails at runtime (image first, then seed).
--
-- WHAT THIS IS. The auto-drain half of the owner's 2026-07-20 ruling on 033
-- ("split it — auto-drain what can be, queue the rest"), re-affirmed
-- 2026-07-22 ("this is a queue, not a bin — a human works these").
--
-- The queue had 370 parked items on 2026-07-25 and has never had ONE actioned
-- through the admin surface, which has been visible and reachable since
-- 2026-07-20. The blocker is not the size: 321 of the 370 describe a page that
-- has been REDEPLOYED since the item was filed, and nothing re-checks them, so
-- a ghost and a live finding are indistinguishable. This sweep re-evaluates
-- each parked finding against currently-deployed state and either closes it
-- with evidence (status='complete', resolution_path='auto:revalidated' — the
-- first Go writer that column has ever had) or stamps it as re-confirmed so the
-- survivors can be trusted.
--
-- Closing is REVERSIBLE by construction: every terminal status is excluded from
-- idx_swi_dedup, so a close releases the item's dedup key and the originating
-- check re-raises it if the finding is in fact still true. A wrong close costs
-- one re-raise; it cannot lose a finding.
--
-- ██ SHIPS IN DRY_RUN ██ — dry_run=true: verdicts reported in the payload, NO
-- status changes, NO annotations. Review a dry run first (expect roughly 51
-- 'resolved' across the three covered types), then flip dry_run false with the
-- jsonb_set at the bottom. Owner: MANUAL trigger for now (mirrors
-- diagnosis-superseded-reviews and dormant-agents).
--
-- COVERAGE IS DELIBERATELY PARTIAL. v1 revalidates unresolved_cta,
-- required_fields_missing and needs_section_data. Everything else returns
-- 'unknown' and is counted in the payload's uncovered_types so the gap is
-- REPORTED, never silently read as "nothing to drain".
-- cta_names_unknown_destination (the largest class) is excluded on purpose:
-- bugs_open/023 / cta_link_integrity owns that check and is mid-flight on it.

BEGIN;

SELECT snapshot_agent('diagnosis-review-queue-revalidator', 'pre-update: review-queue revalidator v1 re-apply')
WHERE EXISTS (SELECT 1 FROM agent_definitions
              WHERE type='diagnosis-review-queue-revalidator' AND is_active
                AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL);

INSERT INTO agent_definitions (
    type, display_name, description, category, agent_category, status,
    is_active, version, capabilities,
    image_repository, image_tag, command, resources, topics, health_config, env_vars,
    default_config
)
SELECT
    'diagnosis-review-queue-revalidator',
    'Review-Queue Revalidator (the needs_human_review drain)',
    'Deterministic drain for the needs_human_review queue (no LLM, bugs_open/033): re-evaluates each parked finding against currently-deployed page_components state. Closes the ones that provably no longer hold (status=complete, resolution_path=auto:revalidated, evidence in result.revalidation); stamps the ones that still hold so a human can see they were re-confirmed and when; reports the item_types it cannot judge. Never guesses: every ambiguity is unknown and stays queued. Ships dry_run. Manual for now.',
    'diagnose', 'coordinator', 'experimental',
    true, 1, '["diagnose", "reconcile"]'::jsonb,
    d.image_repository, d.image_tag, d.command, d.resources, d.topics, d.health_config, d.env_vars,
    jsonb_build_object('workflow', jsonb_build_object(
      'start_step', 'sweep',
      'processing_mode', 'task',
      'timeout_seconds', 180,
      'steps', jsonb_build_object(

        'sweep', jsonb_build_object(
          'action', 'revalidate_review_queue',
          'description', 'Re-evaluate parked needs_human_review findings against deployed state; close the resolved, stamp the rest (unless dry_run).',
          'output_field', 'revalidation_result',
          'next_step', 'complete',
          'config', jsonb_build_object(
            'max_items', 50,
            'dry_run', true   -- FLIP to false once the dry-run verdict list looks right
            -- optional narrowing for a first live run:
            --   'item_type', 'unresolved_cta'
            --   'site_id', '<uuid>'
          )
        ),

        'complete', jsonb_build_object(
          'action', 'complete_workflow',
          'description', 'Revalidation sweep done; per-item verdicts in the payload.',
          'config', jsonb_build_object(
            'output_fields', jsonb_build_array('revalidation_result'),
            'success_message', 'review-queue revalidation swept; verdicts in payload; items closed/stamped unless dry_run'
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

-- ── After reviewing a dry-run verdict list ──────────────────────────────────
-- Narrow the first LIVE run to one class before opening it up:
--   UPDATE agent_definitions
--   SET default_config = jsonb_set(
--         jsonb_set(default_config, '{workflow,steps,sweep,config,dry_run}', 'false'::jsonb),
--         '{workflow,steps,sweep,config,item_type}', '"unresolved_cta"'::jsonb),
--       updated_at = now()
--   WHERE type='diagnosis-review-queue-revalidator' AND is_active AND COALESCE(is_snapshot,false)=false;
--
-- Then widen by removing the item_type pin:
--   UPDATE agent_definitions
--   SET default_config = default_config #- '{workflow,steps,sweep,config,item_type}', updated_at=now()
--   WHERE type='diagnosis-review-queue-revalidator' AND is_active AND COALESCE(is_snapshot,false)=false;

-- ── Reading what it did ─────────────────────────────────────────────────────
-- What got closed, and on what evidence:
--   SELECT id, item_type, resolution_path, result->'revalidation'->>'reason'
--   FROM site_work_items WHERE resolution_path = 'auto:revalidated' ORDER BY completed_at DESC;
--
-- Survivors, with their re-confirmation stamp — this is the trust signal the
-- queue has never had:
--   SELECT item_type, result->'revalidation'->>'verdict' AS verdict,
--          result->'revalidation'->>'at' AS checked_at, count(*)
--   FROM site_work_items WHERE status='needs_human_review' AND result ? 'revalidation'
--   GROUP BY 1,2,3 ORDER BY 4 DESC;
--
-- Queue depth before/after:
--   SELECT count(*) FROM site_work_items WHERE status='needs_human_review';

-- ── Rollback ────────────────────────────────────────────────────────────────
-- Agent: restore the pre-update snapshot from agent_definitions_backup, or
--   DELETE FROM agent_definitions WHERE type='diagnosis-review-queue-revalidator' AND version=1;
-- Data: every close is individually reversible and self-identifying —
--   UPDATE site_work_items SET status='needs_human_review', completed_at=NULL,
--          resolution_path=NULL, result = result - 'revalidation'
--   WHERE resolution_path='auto:revalidated';
