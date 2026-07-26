-- 203_agent_run_stats.sql — durable per-agent run record (bugs_open/060).
--
-- WHY. There is no durable record of which agents have ever run:
-- agent_definitions.usage_count is unmaintained (0 for all agents), and
-- orchestration_states is pruned hourly at 24h by database-cleanup. So the
-- lifetime "has this capability ever run" question — the one answered wrong for
-- section-editor (bugs_closed/002 D) — cannot be answered. This table is the
-- durable substrate: one row per agent type, upserted at orchestration START
-- (start = work was ROUTED to the agent, which is the dormancy signal 044 wants),
-- keyed on the RESOLVED real agent type (see the owner_agent_type fix shipping in
-- the same image). It is NOT pruned — it accumulates for the life of the fleet.
--
-- Keyed on the resolved dispatched type, so unlike the step-fingerprint method it
-- has no mirrored-agent blind spot and no council/subtree false-negatives.
--
-- SAFE TO APPLY ANYTIME: additive, starts empty. Nothing reads it until the
-- chassis image carrying the RecordAgentRun writer + the rewired dormant detector
-- is live; nothing writes it until then either. Additive-only, no snapshot_agent
-- (touches no agent_definitions row).

BEGIN;

CREATE TABLE IF NOT EXISTS agent_run_stats (
    agent_type            text        PRIMARY KEY,
    run_count             bigint      NOT NULL DEFAULT 0,
    first_ran_at          timestamptz NOT NULL DEFAULT now(),
    last_ran_at           timestamptz NOT NULL DEFAULT now(),
    last_orchestration_id uuid,
    updated_at            timestamptz NOT NULL DEFAULT now()
);

COMMENT ON TABLE agent_run_stats IS
  'Durable per-agent-type run record (bugs_open/060). Upserted at orchestration start on the resolved real agent type. NOT pruned (unlike orchestration_states). Source of truth for the dormant-agents detector (bugs_open/044).';

-- last_ran_at drives the dormancy read; index it for the detector''s scans.
CREATE INDEX IF NOT EXISTS idx_agent_run_stats_last_ran ON agent_run_stats (last_ran_at);

COMMIT;

-- Verify:
--   SELECT agent_type, run_count, first_ran_at, last_ran_at
--   FROM agent_run_stats ORDER BY last_ran_at DESC;
-- Rollback:
--   DROP TABLE IF EXISTS agent_run_stats;
