-- SEED — evidence-freshness agent, and repointing the task at it (bugs_open/074)
--
-- WHY THIS EXISTS: `evidence-freshness` has NEVER RUN. Not once since it was
-- seeded (claims_verification V4, SEED_evidence_freshness_scheduled_task.sql).
--
-- It carries its workflow INLINE, in scheduled_tasks.input_data.config.workflow,
-- and targets target_agent_type='generic'. The chassis DOES honour an inline
-- workflow — but only at body.config.workflow (processor.go selectWorkflow,
-- Priority 1, ahead of group discovery; 58 live orchestrations arrive that way
-- from DispatchFeedSourcesAction). The SCHEDULER cannot express it: fireTrigger
-- (cmd/scheduler/main.go) builds `config` itself from the row's COLUMNS and puts
-- the whole input_data column underneath, so an inline workflow lands at
-- body.input_data.config.workflow — one level below anything that reads it.
--
-- `generic`'s own workflow is a single no-op step, so every fire completed
-- instantly, stamped last_triggered_at AND last_completed_at, and did nothing.
-- Measured 2026-07-26, before this seed:
--
--   SELECT count(*) FROM orchestration_states WHERE workflow_plan::text LIKE '%refresh_evidence_base%';
--   --  0
--
-- THE FIX: carry the workflow where the chassis reads it — an agent_definitions
-- row — exactly as model_directory did on 07-25 for the identical shape
-- (SEED_directory_freshness_agent.sql, verified live: its sweep now runs, and
-- reports 0 due rather than never running). Config-only, no image roll: the
-- action `refresh_evidence_base` has been in the image all along and simply had
-- no caller (`strings /app/agent-chassis | grep -c refresh_evidence_base` → 11).
--
-- ── dry_run: true IS TEMPORARY, AND MUST BE FLIPPED ────────────────────────
--
-- This sweep has never run, and it rewrites the `writer_block` on every site
-- with `writer_block_managed: true` — today leopardessconsulting.co.uk (9
-- sql-sourced facts) and fundamentallyai.com (7), the latter written hours
-- earlier by another live workstream. So the FIRST pass reports and writes
-- nothing. **Stage 2 below is not optional**: a dry run left on is a silent
-- no-op with a green status, which is the exact defect this file is fixing.
--
-- ROLLBACK
-- --------
--   UPDATE scheduled_tasks SET enabled = false WHERE name = 'evidence-freshness';
--   -- (and, if the agent row must go too:)
--   UPDATE agent_definitions SET deleted_at = now() WHERE type = 'evidence-freshness';
-- Lossless: the action writes a NEW site_specs revision rather than mutating in
-- place, so every pre-refresh evidence base survives at is_current = false.

BEGIN;

-- ── Stage 1: the agent that carries the workflow ───────────────────────────
INSERT INTO agent_definitions (
    type, display_name, description, category, agent_category, status, is_active,
    image_repository, image_tag, input_contract, output_contract, default_config
)
SELECT
    'evidence-freshness',
    'Evidence Freshness Sweep',
    'V4 of the claims-verification layer: re-runs every live-verifiable (source.sql) fact in a site evidence base, updates value/verified_at from the query that declares the fact, regenerates the writer whitelist where the site opted in, and raises stale_evidence for human ruling when a live value drifts outside the tolerance its published wording allows — including when the site is UNDER-claiming.',
    'analyst', 'analyst', 'active', true,
    'docker.io/aqls/agent-chassis',
    (SELECT image_tag FROM agent_definitions
      WHERE type = 'claims-auditor' AND is_active AND COALESCE(is_snapshot,false) = false AND deleted_at IS NULL),
    '{"optional": ["site_id"]}'::jsonb,
    '{"produces": {"refresh_result": "per-site counts of facts checked, updated and drifted, with the writer_block action taken"}}'::jsonb,
    $cfg${
  "workflow": {
    "start_step": "refresh_evidence",
    "processing_mode": "orchestrator",
    "timeout_seconds": 600,
    "steps": {
      "refresh_evidence": {
        "action": "refresh_evidence_base",
        "config": {"dry_run": true},
        "next_step": "complete",
        "output_field": "refresh_result",
        "description": "Sweep every site with an evidence base; no site_id = all of them. dry_run is TEMPORARY — see stage 2, and bugs_open/074."
      },
      "complete": {
        "action": "complete_workflow",
        "config": {"output_fields": ["refresh_result"]}
      }
    }
  }
}$cfg$::jsonb
WHERE NOT EXISTS (
    SELECT 1 FROM agent_definitions WHERE type = 'evidence-freshness' AND deleted_at IS NULL
);

-- ── Stage 1b: repoint the task, and strip the workflow that was never read ──
UPDATE scheduled_tasks
SET target_agent_type = 'evidence-freshness',
    input_data        = '{}'::jsonb,
    description       = 'V4 of the claims-verification layer: re-run live-verifiable evidence_base facts, re-sync value/verified_at, regenerate the V2 writer whitelist, raise stale_evidence for human ruling when a live value drifts outside the tolerance its published wording allows (including under-claiming). Repointed 2026-07-26 from target_agent_type=generic + inline workflow — a shape the scheduler cannot express and nothing reads, so the sweep had never run (bugs_open/074).',
    last_triggered_at = NULL,
    updated_at        = now()
WHERE name = 'evidence-freshness';

SELECT name, target_agent_type, enabled, interval_seconds, input_data::text
FROM scheduled_tasks WHERE name = 'evidence-freshness';

COMMIT;

-- ── Stage 2: AFTER reading the dry-run report, turn the writes on ──────────
--
-- Read it first (the run is dispatched within one 30s scheduler tick):
--
--   SELECT jsonb_pretty(collected_data->'refresh_result')
--   FROM orchestration_states
--   WHERE workflow_plan::text LIKE '%refresh_evidence_base%'
--   ORDER BY created_at DESC LIMIT 1;
--
-- Then, and only then:
--
--   UPDATE agent_definitions
--   SET default_config = jsonb_set(default_config,
--         '{workflow,steps,refresh_evidence,config,dry_run}', 'false'::jsonb),
--       updated_at = now()
--   WHERE type = 'evidence-freshness' AND deleted_at IS NULL;
--
--   UPDATE scheduled_tasks SET last_triggered_at = NULL WHERE name = 'evidence-freshness';
--
-- Assert the flip actually landed — a dry run left on looks identical to a
-- healthy one from the outside, which is this bug all over again:
--
--   SELECT default_config->'workflow'->'steps'->'refresh_evidence'->'config'->>'dry_run'
--   FROM agent_definitions WHERE type = 'evidence-freshness' AND deleted_at IS NULL;
--   -- must read: false
--
-- ── Post-apply verification — verify the FAILING branch ────────────────────
--
-- 1) The run actually carried the action (0 before this seed):
--      SELECT count(*) FROM orchestration_states WHERE workflow_plan::text LIKE '%refresh_evidence_base%';
-- 2) The artefact, not the status — a NEW spec revision written by the sweep:
--      SELECT created_by, pinned, notes, created_at FROM site_specs
--      WHERE aspect='evidence_base' AND is_current ORDER BY created_at DESC;
--      -- expect created_by='evidence-refresher', pinned intact, a 'V4 freshness pass: …' note
-- 3) INDUCE A FAULT — the only proof a detector detects. On leopardess (NOT
--    fundamentallyai, whose register another workstream is actively using),
--    take a copy, corrupt one sql-sourced fact's value, backdate its
--    verified_at, force the task, and expect drifted >= 1 plus a stale_evidence
--    work item. The pass writes the live value back, so the fact restores
--    itself; cancel the work item afterwards.
