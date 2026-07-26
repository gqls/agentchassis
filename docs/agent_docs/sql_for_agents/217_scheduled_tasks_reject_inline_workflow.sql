-- 217_scheduled_tasks_reject_inline_workflow.sql
--
-- bugs_open/074 — a scheduled task that carries its workflow inline is accepted
-- without complaint, fires, reports success, and does nothing. Three tasks were
-- authored this way; one of them (`evidence-freshness`) had never run once.
--
-- THE MECHANISM (read the code, not the symptom)
-- ----------------------------------------------
-- The chassis DOES honour an inline workflow — at `body.config.workflow`
-- (platform/messaging/processor.go, selectWorkflow Priority 1, taken ahead of
-- group discovery). It is a live production path, not a testing hook: 58
-- orchestration_states rows carry the inline workflow that
-- DispatchFeedSourcesAction dispatches that way (measured 2026-07-26).
--
-- The SCHEDULER cannot express it. fireTrigger (cmd/scheduler/main.go) builds
-- the message `config` itself, out of the row's COLUMNS —
--     {"action":"orchestrate","config":{"agent_type":<target_agent_type>},
--      "input_data":<the whole input_data column>}
-- — so a workflow authored inside input_data lands at
-- body.input_data.config.workflow, one level below anything that reads it. It
-- is then discarded in silence, and the target agent's own workflow runs. For
-- `generic` that is a single complete_workflow no-op, so the task fires, stamps
-- last_triggered_at AND last_completed_at, and every health check reads green.
--
-- WHY A CONSTRAINT AND NOT "MAKE THE SHAPE WORK"
-- ---------------------------------------------
-- bugs_closed/054 settled the contract for this column: scheduled_tasks.input_data
-- is the PAYLOAD ONLY — fireTrigger supplies action/config/input_data — and
-- teaching fireTrigger to reach into the payload is the regression class 042
-- warned about ("a legitimate payload field named input_data/action/config would
-- be silently eaten"). A workflow's home is an agent_definitions row, where it is
-- versioned, snapshotted and discoverable. So: refuse the shape at the moment it
-- is authored, rather than accept it and drop it at 03:00 every night.
--
-- This is live the moment it is applied — no image roll. The scheduler also
-- gained a WARN for the same shape (cmd/scheduler/main.go), which is
-- belt-and-braces: with this constraint in place the state it detects cannot be
-- authored, but a constraint can be dropped and a row can be restored.
--
-- SAFETY (checked 2026-07-26)
-- ---------------------------
-- * Zero rows violate it: `evidence-freshness` was repointed at a new
--   `evidence-freshness` agent_definitions row (its sweep then ran for the FIRST
--   TIME — 3 site_specs revisions written by `evidence-refresher`, 3
--   stale_evidence items raised for human review), and the disabled
--   `adoption-tracker-freshness` had its never-read workflow stripped, preserved
--   verbatim in bugs_closed/074.
-- * The admin API cannot trip it: pipeline_admin_handlers.go only ever UPDATEs
--   `enabled` and `interval_seconds` (a PATCH body with exactly those fields).
-- * It is deliberately NARROW — `input_data->'config' ? 'workflow'` only. The
--   other envelope-shaped rows (`diagnose-pipeline-trigger`,
--   `ai-endpoint-health-check`, `ch-fetch-accounts`) carry no workflow, are the
--   054 family, and stay legal. Widening this to ban `config`/`action`/
--   `input_data` outright would break two live tasks for no gain.
-- * Not enabled-only: a disabled row carrying the trap is a landmine waiting for
--   whoever enables it.
--
-- ROLLBACK
-- --------
--   ALTER TABLE scheduled_tasks DROP CONSTRAINT scheduled_tasks_no_inline_workflow;

BEGIN;

ALTER TABLE scheduled_tasks DROP CONSTRAINT IF EXISTS scheduled_tasks_no_inline_workflow;

ALTER TABLE scheduled_tasks
    ADD CONSTRAINT scheduled_tasks_no_inline_workflow
    CHECK (NOT (input_data -> 'config' ? 'workflow'));

COMMENT ON CONSTRAINT scheduled_tasks_no_inline_workflow ON scheduled_tasks IS
    'bugs_open/074: input_data is the PAYLOAD ONLY (bugs_closed/054) — the scheduler builds the message envelope from the row''s columns, so a workflow authored at input_data.config.workflow lands one level too deep and is discarded in silence while the task reports success. Put the workflow in an agent_definitions row and point target_agent_type at it (see SEED_evidence_freshness_agent.sql / SEED_directory_freshness_agent.sql).';

COMMIT;

-- ── Post-apply verification — assert the FAILING branch, not the happy one ──
--
-- Three checks, verbatim SQL in
-- docs/agent_docs/docs024_key_docs_latest/bugfix_074_inline_workflow/RUNBOOK_inline_workflow.md
-- (kept there rather than inline: the runner's idempotency lint reads comment
-- text, so a pasteable probe in this header raises a false "unguarded insert"):
--
--   1) a probe row carrying input_data.config.workflow is REFUSED — SQLSTATE
--      23514, naming scheduled_tasks_no_inline_workflow;
--   2) POSITIVE CONTROL — the same probe row with an ordinary payload is
--      accepted, so check 1 is not passing for the wrong reason;
--   3) no surviving row violates it:
--        SELECT name FROM scheduled_tasks WHERE input_data->'config' ? 'workflow';
--      -- expect 0 rows
