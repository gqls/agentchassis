-- ============================================================================
-- 025_thunder_adapter_schema.sql
-- ============================================================================
-- Schema for the thunder-adapter (see 013_thunder_adapter_design.md).
--
-- Two tables:
--   thunder_instances — one row per Thunder VM ever provisioned. Source of
--                       truth for the adapter and the reaper.
--   thunder_config    — singleton-row global configuration (cost cap,
--                       concurrency limit, default uptime, hourly rate).
--
-- Two views:
--   thunder_spend_24h        — computes rolling 24h cost from instances,
--                              including running estimates.
--   thunder_provision_check  — combines config + spend + active count to
--                              answer "can we provision a new VM right now?"
--
-- Step Zero check (per dev guide §0):
--   Searched for: thunder, gpu instance, vm tracking, instance lifecycle
--     - no existing schema. New tables justified.
--   Searched for: budget cap, cost tracking, rolling window
--     - found: page_growth_budget.go uses 7-day rolling window with COUNT(*)
--       FILTER (WHERE created_at > NOW() - INTERVAL '7 days') — same pattern,
--       different scope. Reused for our 24h cost window.
--   Searched for: existing thunder_instance_id references
--     - model_lifecycle.training_runs.thunder_instance_id (TEXT) already
--       exists from migration 019. We FK from thunder_instances.training_run_id
--       to training_runs.id (UUID); the reverse linkage stays as a TEXT
--       breadcrumb in training_runs.
--
-- Decisions (from user, recorded in 013):
--   - daily cap: $100
--   - max concurrent: 2
--   - default uptime: 18 hours
--   - estimated new run cost: $25 (covers iter_0-sized run with margin)
--   - assumed hourly rate: $1.80 (A100 80GB single)
-- ============================================================================


-- ---------------------------------------------------------------------------
-- thunder_instances
-- ---------------------------------------------------------------------------
-- One row per Thunder Compute VM ever provisioned.
--
-- Inserted at provision time BEFORE the API call returns success to the
-- caller — even if the call later fails or the adapter crashes, we have a
-- record of "we asked Thunder to make a VM". The reaper relies on this.
--
-- status transitions:
--   provisioning → running         (Thunder API confirmed instance is up)
--   running → decommissioning      (decommission_instance action started)
--   decommissioning → decommissioned (Thunder API confirmed deletion)
--   * → reaped                     (reaper killed it for orphan/over-time/cap)
--   * → lost                       (Thunder shows it gone but we didn't
--                                   initiate; sets cost_usd=NULL)
--   provisioning → failed          (Thunder API rejected provisioning)
--
-- decommissioned + reaped are terminal (no further updates expected).
-- ---------------------------------------------------------------------------

CREATE TABLE IF NOT EXISTS thunder_instances (
                                                 id                          UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    -- Thunder identity
    thunder_instance_id         TEXT NOT NULL UNIQUE,           -- from Thunder API e.g. "ti_abc123"
    instance_type               TEXT NOT NULL,                  -- e.g. "a100_80gb_single"
    instance_ip                 TEXT,                           -- populated when status='running'
    ssh_user                    TEXT NOT NULL DEFAULT 'ubuntu',
    ssh_key_secret_name         TEXT NOT NULL,                  -- name of k8s secret holding private key

-- Lifecycle
    status                      TEXT NOT NULL CHECK (status IN
('provisioning','running','decommissioning',
                                                     'decommissioned','reaped','lost','failed')),
    max_uptime_hours            INTEGER NOT NULL DEFAULT 18,

    -- Linkage to training_runs (the typical caller)
    training_run_id             UUID REFERENCES model_lifecycle.training_runs(id),
    requested_by                TEXT,                           -- agent_type that called provision_instance

-- Cost (snapshot at provision; updated on decommission)
    hourly_rate_usd             NUMERIC NOT NULL DEFAULT 1.80,
    cost_usd                    NUMERIC,                        -- final cost when decommissioned

-- Reaper bookkeeping
    reaped_at                   TIMESTAMPTZ,
    reaped_reason               TEXT,                           -- 'orphan_no_match' / 'over_uptime' / 'cost_cap_breach'

-- Failure
    error_message               TEXT,

    -- Timestamps
    provisioned_at              TIMESTAMPTZ,                    -- adapter sent provision request
    running_since               TIMESTAMPTZ,                    -- Thunder API confirmed running
    decommission_requested_at   TIMESTAMPTZ,                    -- adapter sent decommission request
    decommissioned_at           TIMESTAMPTZ,                    -- Thunder API confirmed gone
    created_at                  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at                  TIMESTAMPTZ NOT NULL DEFAULT NOW()
    );

-- Reaper queries: anything still active
CREATE INDEX IF NOT EXISTS idx_thunder_instances_active
    ON thunder_instances(status, running_since)
    WHERE status IN ('provisioning', 'running', 'decommissioning');

-- Linkage from training_runs (find the VM that ran this training)
CREATE INDEX IF NOT EXISTS idx_thunder_instances_training_run
    ON thunder_instances(training_run_id) WHERE training_run_id IS NOT NULL;

-- For 24h rolling window queries (decommissioned recently)
CREATE INDEX IF NOT EXISTS idx_thunder_instances_decommissioned_recent
    ON thunder_instances(decommissioned_at) WHERE decommissioned_at IS NOT NULL;


-- ---------------------------------------------------------------------------
-- updated_at trigger
-- ---------------------------------------------------------------------------

CREATE OR REPLACE FUNCTION thunder_set_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trg_thunder_instances_updated_at ON thunder_instances;
CREATE TRIGGER trg_thunder_instances_updated_at
    BEFORE UPDATE ON thunder_instances
    FOR EACH ROW
    EXECUTE FUNCTION thunder_set_updated_at();


-- ---------------------------------------------------------------------------
-- thunder_config — singleton-row global configuration
-- ---------------------------------------------------------------------------
-- Single-row pattern enforced by CHECK on a fixed PK value. Avoids the trap
-- of multi-row config where rows accumulate and queries become unclear.
-- ---------------------------------------------------------------------------

CREATE TABLE IF NOT EXISTS thunder_config (
                                              singleton                   CHAR(1) PRIMARY KEY DEFAULT 'X' CHECK (singleton = 'X'),

    daily_cap_usd               NUMERIC NOT NULL DEFAULT 100,
    max_concurrent_instances    INTEGER NOT NULL DEFAULT 2,
    default_hard_uptime_hours   INTEGER NOT NULL DEFAULT 18,
    default_hourly_rate_usd     NUMERIC NOT NULL DEFAULT 1.80,
    estimated_new_run_cost_usd  NUMERIC NOT NULL DEFAULT 25,    -- pessimistic per-run, used in cap pre-check

    is_paused                   BOOLEAN NOT NULL DEFAULT false,
    pause_reason                TEXT,

    updated_at                  TIMESTAMPTZ NOT NULL DEFAULT NOW()
    );

DROP TRIGGER IF EXISTS trg_thunder_config_updated_at ON thunder_config;
CREATE TRIGGER trg_thunder_config_updated_at
    BEFORE UPDATE ON thunder_config
    FOR EACH ROW
    EXECUTE FUNCTION thunder_set_updated_at();

-- Seed the singleton row with the user's confirmed values
INSERT INTO thunder_config (singleton)
VALUES ('X')
    ON CONFLICT (singleton) DO NOTHING;


-- ---------------------------------------------------------------------------
-- thunder_spend_24h — rolling 24-hour cost computation
-- ---------------------------------------------------------------------------
-- Sums cost_usd for instances decommissioned in the last 24 hours, plus an
-- estimate for currently-running instances based on their elapsed time and
-- recorded hourly rate. Computed on every read; no static counter to drift.
-- ---------------------------------------------------------------------------

CREATE OR REPLACE VIEW thunder_spend_24h AS
SELECT
    COALESCE(SUM(cost_usd) FILTER (
        WHERE decommissioned_at > NOW() - INTERVAL '24 hours'
    ), 0)::NUMERIC AS decommissioned_spend_24h,

    COALESCE(SUM(
                     CASE WHEN status IN ('provisioning', 'running') AND running_since IS NOT NULL
                              THEN hourly_rate_usd * EXTRACT(EPOCH FROM (NOW() - running_since)) / 3600.0
                          ELSE 0
                         END
             ), 0)::NUMERIC AS running_estimated_spend,

    (
        COALESCE(SUM(cost_usd) FILTER (
            WHERE decommissioned_at > NOW() - INTERVAL '24 hours'
        ), 0)
            +
        COALESCE(SUM(
                         CASE WHEN status IN ('provisioning', 'running') AND running_since IS NOT NULL
                                  THEN hourly_rate_usd * EXTRACT(EPOCH FROM (NOW() - running_since)) / 3600.0
                              ELSE 0
                             END
                 ), 0)
        )::NUMERIC AS total_24h_spend,

    COUNT(*) FILTER (WHERE status IN ('provisioning', 'running'))::INTEGER AS active_count
FROM thunder_instances;


-- ---------------------------------------------------------------------------
-- thunder_provision_check — does a provision request pass all gates?
-- ---------------------------------------------------------------------------
-- The adapter SELECTs from this view at the start of every provision_instance
-- request. If can_provision = false, returns to caller with denial_reason.
-- ---------------------------------------------------------------------------

CREATE OR REPLACE VIEW thunder_provision_check AS
SELECT
    c.daily_cap_usd,
    c.max_concurrent_instances,
    c.estimated_new_run_cost_usd,
    c.default_hard_uptime_hours,
    c.default_hourly_rate_usd,
    c.is_paused,
    c.pause_reason,

    s.total_24h_spend,
    s.active_count,

    (
        NOT c.is_paused
            AND s.active_count < c.max_concurrent_instances
            AND (s.total_24h_spend + c.estimated_new_run_cost_usd) <= c.daily_cap_usd
        ) AS can_provision,

    CASE
        WHEN c.is_paused THEN 'paused: ' || COALESCE(c.pause_reason, 'unknown')
        WHEN s.active_count >= c.max_concurrent_instances THEN 'concurrent_cap_reached'
        WHEN (s.total_24h_spend + c.estimated_new_run_cost_usd) > c.daily_cap_usd THEN 'cost_cap_would_exceed'
        ELSE NULL
        END AS denial_reason

FROM thunder_config c
         CROSS JOIN thunder_spend_24h s;


-- ============================================================================
-- Sanity checks (run after applying)
-- ============================================================================

-- Confirm tables and views exist
\dt thunder_instances
\dt thunder_config
\dv thunder_spend_24h
\dv thunder_provision_check

-- Confirm singleton row was seeded with confirmed defaults
SELECT daily_cap_usd, max_concurrent_instances, default_hard_uptime_hours,
       default_hourly_rate_usd, estimated_new_run_cost_usd,
       is_paused, pause_reason
FROM thunder_config;

-- Should show: 100 | 2 | 18 | 1.80 | 25 | f | NULL

-- Confirm the views compute zero spend on empty tables
SELECT * FROM thunder_spend_24h;
-- Should show: 0 | 0 | 0 | 0

SELECT can_provision, denial_reason, daily_cap_usd, max_concurrent_instances,
       total_24h_spend, active_count
FROM thunder_provision_check;
-- Should show: true | NULL | 100 | 2 | 0 | 0