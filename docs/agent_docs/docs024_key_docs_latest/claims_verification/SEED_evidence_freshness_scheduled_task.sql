-- SEED — V4 evidence freshness on a cadence (claims-verification layer)
--
-- ── REWRITTEN 2026-07-26 (bugs_open/074). READ THIS BEFORE RE-APPLYING. ────
--
-- As first authored, this seed carried the workflow INLINE, in
-- input_data.config.workflow, targeting target_agent_type='generic'. **That
-- shape is undeliverable and the sweep never ran once.** The chassis honours an
-- inline workflow only at body.config.workflow, and the scheduler builds that
-- envelope from the row's COLUMNS (cmd/scheduler/main.go fireTrigger), so
-- anything under input_data lands a level too deep and is read by nothing.
-- `generic`'s own workflow is a single no-op step, so every fire completed
-- instantly, stamped BOTH timestamps, and did nothing — for as long as this task
-- had existed.
--
-- The workflow now lives where the chassis reads it, in an agent_definitions row
-- (bugfix_074_inline_workflow/SEED_evidence_freshness_agent.sql), and this task
-- points at it. Migration 217 adds a CHECK constraint, so re-applying the old
-- shape now FAILS with 23514 rather than succeeding and doing nothing.
--
-- Verified live 2026-07-26: the first real pass wrote 3 site_specs revisions as
-- `evidence-refresher` (pinned preserved) and raised 3 stale_evidence items; an
-- induced fault (a fact corrupted to 9,370 against a live 937) came back
-- `drifted` and was re-synced. Follow-on defect found doing that, filed as
-- bugs_open/091 — while an earlier stale_evidence item is open, a second,
-- different drift is dropped by the work-item dedup and still reported as
-- raised.
--
-- APPLY ONLY AFTER a chassis image carrying `refresh_evidence_base` is deployed
-- and pod-verified. CLAUDE.md: "Image first, then seeds (a seed naming an
-- unregistered action fails at runtime)." Verify first, against the POD:
--
--   kubectl -n ai-persona-system exec <agent-chassis-pod> -- \
--     sh -c 'strings /app/agent-chassis | grep -c refresh_evidence_base'
--
-- Expect a non-zero count. If it is 0 the action is not in the image and this
-- seed must NOT be applied.
--
-- Cadence rationale: the facts are counts over slow-moving business data
-- (verified businesses, agent definitions, deployed sites, feed items). Daily
-- is well inside any tolerance and keeps `verified_at` honest without hammering
-- the query layer; the action is idempotent and safely re-runnable, and it
-- writes nothing when nothing moved. 86400s = daily.
--
-- Scope: the action sweeps EVERY site holding a current evidence_base spec when
-- no site_id is supplied — today that is leopardessconsulting.co.uk alone, and
-- each site opted in later is picked up with no seed change.
--
-- Drift terminates at human review (`stale_evidence`, needs_human_review, no
-- automated handler): a number moving is a fact about the world, but changing
-- published copy because of it is a human decision (SPEC §10 open question 3).

INSERT INTO scheduled_tasks (
    name, description, interval_seconds, target_agent_type, target_topic,
    input_data, concurrency_group, max_concurrent, enabled, timeout_seconds
)
SELECT
    'evidence-freshness',
    'V4 of the claims-verification layer: re-run live-verifiable evidence_base facts, re-sync value/verified_at, regenerate the V2 writer whitelist, raise stale_evidence for human ruling when a live value drifts outside the tolerance its published wording allows (including under-claiming).',
    86400,
    -- The agent carrying the workflow — seed it FIRST
    -- (bugfix_074_inline_workflow/SEED_evidence_freshness_agent.sql). A task
    -- naming an agent type that does not exist falls through to discovery and
    -- runs something else's workflow, which is how this seed failed silently.
    'evidence-freshness',
    'system.agent.scheduled.requests',
    -- PAYLOAD ONLY. fireTrigger supplies action / config.agent_type / the
    -- input_data wrapper itself; a `config` or `workflow` key here is
    -- undeliverable (bugs_open/074) and `input_data` here would double-wrap
    -- (bugs_closed/054). The sweep needs no payload: no site_id means "every
    -- site with an evidence base", which is the intended default.
    '{}'::jsonb,
    'evidence-freshness',
    1,
    true,
    600
WHERE NOT EXISTS (SELECT 1 FROM scheduled_tasks WHERE name = 'evidence-freshness');

-- ── Post-apply verification ────────────────────────────────────────────────
--
-- 0. FIRST, the discriminating check — did a run ever carry the action at all?
--    A fired task, a COMPLETED orchestration and two advancing timestamps are
--    all equally true of the shape that never worked; this is not:
--
--   SELECT count(*) FROM orchestration_states
--   WHERE workflow_plan::text LIKE '%refresh_evidence_base%';   -- was 0 for the whole life of the old seed
--
-- 1. The row exists and is enabled:
--
--   SELECT name, target_agent_type, interval_seconds, enabled
--   FROM scheduled_tasks WHERE name = 'evidence-freshness';
--
-- 2. After the first fire (within `interval_seconds`, or set
--    last_triggered_at = NULL to bring it forward), the register's dates move:
--
--   SELECT (f->>'id') AS fact, f->>'value' AS value, f->>'verified_at' AS verified
--   FROM site_specs ss, jsonb_array_elements(ss.data->'facts') f
--   WHERE ss.aspect = 'evidence_base' AND ss.is_current = true
--     AND f->'source' ? 'sql'
--   ORDER BY 1;
--
-- 3. The spec row records the pass in its notes, and `pinned` survives:
--
--   SELECT created_by, pinned, notes FROM site_specs
--   WHERE aspect = 'evidence_base' AND is_current = true;
--
--   Expect created_by = 'evidence-refresher', pinned = true, and a note of the
--   form 'V4 freshness pass: N live-verifiable fact(s) checked, ...'.
--
-- 4. Any drift raised for a human (empty is the healthy steady state):
--
--   SELECT summary, status, spec->'drifted' FROM site_work_items
--   WHERE item_type = 'stale_evidence' ORDER BY created_at DESC;
--
-- 5. The regenerated whitelist still reads as its human authors wrote it —
--    caveats intact, numbers current:
--
--   SELECT data->>'writer_block' FROM site_specs
--   WHERE aspect = 'evidence_base' AND is_current = true;
--
-- ── Rollback ───────────────────────────────────────────────────────────────
--
--   UPDATE scheduled_tasks SET enabled = false WHERE name = 'evidence-freshness';
--
-- Disabling is sufficient and lossless: the action writes a new spec revision
-- rather than mutating in place, so every pre-refresh evidence base remains in
-- site_specs history (is_current = false) and can be restored by hand.
