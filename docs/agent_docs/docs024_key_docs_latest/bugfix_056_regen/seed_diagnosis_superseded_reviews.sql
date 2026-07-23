-- seed_diagnosis_superseded_reviews.sql — REVIEW-BYPASS reconciliation sweep.
-- bugs_open/056 (regeneration_silently_drops_content…, the regen 056 — resolve
-- the ambiguous number by slug). 2026-07-23. Applies to clients_db.
--
-- ██ DEPLOY SEQUENCING ██ — apply ONLY AFTER a chassis image carrying the
-- reconcile_superseded_reviews action is live. Verify in-pod:
--   kubectl -n ai-persona-system exec <chassis-pod> -- \
--     sh -c 'strings /app/agent-chassis | grep -c reconcile_superseded_reviews'  (>= 1)
-- A seed naming an unregistered action fails at runtime (image first, then seed).
-- Verified live on v1.0.1150, 2026-07-23 (count 3 + REVIEW_SUPERSEDED literal 2).
--
-- WHAT THIS IS. The council-approved standalone reconciler (corr 1d8ef2c0,
-- round 2 APPROVED 2026-07-22): a deterministic sweep (no LLM) over
-- (work item parked at needs_human_review) × (its page deployed AFTER the
-- parking) pairs — the review-bypass-by-sibling-item mechanism the diagnosis
-- loop identified (corr b361298a) after refuting the originally-filed one.
-- Per pair it checks every previously-flagged blocker value against the
-- CURRENTLY-DEPLOYED page_components content (dropped-not-resolved vs
-- still-present), writes one REVIEW_SUPERSEDED_BY_PASSING_SAVE row to
-- agent_error_log and annotates the parked item's
-- result.superseded_by_passing_save (also the idempotence key). It never
-- blocks anything, closes nothing, and touches no item status — whether a
-- parked review should HOLD deploys is an owner policy call this sweep only
-- produces the evidence for (bugs_open/033: the review queue has no drain).
--
-- ██ SHIPS IN DRY_RUN ██ — dry_run=true: report pairs only; NO log rows, NO
-- annotations. Review a dry run (expect the fundamentallyai
-- needs_page:model-fine-tuning pair), then flip dry_run false (jsonb_set at
-- the bottom). Owner: MANUAL trigger for now (mirrors dormant-agents).

BEGIN;

SELECT snapshot_agent('diagnosis-superseded-reviews', 'pre-update: superseded-reviews v1 re-apply')
WHERE EXISTS (SELECT 1 FROM agent_definitions
              WHERE type='diagnosis-superseded-reviews' AND is_active
                AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL);

INSERT INTO agent_definitions (
    type, display_name, description, category, agent_category, status,
    is_active, version, capabilities,
    image_repository, image_tag, command, resources, topics, health_config, env_vars,
    default_config
)
SELECT
    'diagnosis-superseded-reviews',
    'Diagnosis Superseded-Reviews (review-bypass reconciler)',
    'Deterministic review-bypass reconciler (no LLM, bugs_open/056 regeneration): pages deployed after a sibling work item parked them at needs_human_review; flags previously-blocked values as dropped vs still-present against deployed content; writes REVIEW_SUPERSEDED_BY_PASSING_SAVE + annotates the parked item. Never blocks, closes nothing. Ships dry_run. Manual for now.',
    'diagnose', 'coordinator', 'experimental',
    true, 1, '["diagnose", "reconcile"]'::jsonb,
    d.image_repository, d.image_tag, d.command, d.resources, d.topics, d.health_config, d.env_vars,
    jsonb_build_object('workflow', jsonb_build_object(
      'start_step', 'sweep',
      'processing_mode', 'task',
      'timeout_seconds', 120,
      'steps', jsonb_build_object(

        'sweep', jsonb_build_object(
          'action', 'reconcile_superseded_reviews',
          'description', 'Scan for pages deployed past a parked needs_human_review; write supersession evidence (unless dry_run).',
          'output_field', 'superseded_reviews_result',
          'next_step', 'complete',
          'config', jsonb_build_object(
            'max_items', 25,
            'dry_run', true   -- FLIP to false once the dry-run pair list looks right
          )
        ),

        'complete', jsonb_build_object(
          'action', 'complete_workflow',
          'description', 'Superseded-review sweep done; pair report in the payload.',
          'config', jsonb_build_object(
            'output_fields', jsonb_build_array('superseded_reviews_result'),
            'success_message', 'diagnosis-superseded-reviews swept; pair report in payload; evidence rows written unless dry_run'
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

-- Flip dry_run off (after reviewing a dry-run pair report):
--   UPDATE agent_definitions
--   SET default_config = jsonb_set(default_config,
--         '{workflow,steps,sweep,config,dry_run}', 'false'::jsonb), updated_at=now()
--   WHERE type='diagnosis-superseded-reviews' AND is_active AND COALESCE(is_snapshot,false)=false;
--
-- Read written evidence:
--   SELECT occurred_at, error_message FROM agent_error_log
--   WHERE error_code='REVIEW_SUPERSEDED_BY_PASSING_SAVE' ORDER BY occurred_at DESC LIMIT 10;
--   SELECT item_key, result->'superseded_by_passing_save' FROM site_work_items
--   WHERE result ? 'superseded_by_passing_save';
--
-- Rollback: restore the pre-update snapshot from agent_definitions_backup, or
-- DELETE FROM agent_definitions WHERE type='diagnosis-superseded-reviews' AND version=1;
