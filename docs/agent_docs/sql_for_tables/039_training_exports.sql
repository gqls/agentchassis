-- ============================================================================
-- Training exports schema — flywheel A v3
-- ============================================================================
-- Replaces ephemeral JSONL files with Postgres-backed storage.
--
-- Rationale:
--   - JSONL files landed on ephemeral chassis pods and vanished on restart
--   - Sizes (21MB-2GB full sweep) fit Postgres TOAST comfortably
--   - Queryable metadata; no second storage system to operate
--   - Datasets are named versioned artefacts via runs.id (UUID)
--
-- Two tables:
--   runs — one row per export (filter, counts, timestamps)
--   rows — one row per training record, FK'd to runs, CASCADE delete
--
-- Naming: schema `training_exports`. Not `training` because that name
-- could be confused with the actual model training pipeline (flywheel C).
--
-- Usage:
--   - INSERT into runs, get id back, then bulk INSERT into rows
--   - Update runs.rows_exported, size_bytes when done
--   - At training time: \copy (SELECT jsonb_build_object(...) FROM rows
--                              WHERE export_id = ? ORDER BY row_index) TO file
-- ============================================================================

CREATE SCHEMA IF NOT EXISTS training_exports;

-- ---------------------------------------------------------------------------
-- runs — one row per completed export
-- ---------------------------------------------------------------------------
-- Captures what was exported, when, with what filter, and the counts.
-- Source of truth for "which exports exist and what's in them".
-- ---------------------------------------------------------------------------

CREATE TABLE IF NOT EXISTS training_exports.runs (
                                                     id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    -- Filter criteria (matches llm_call_log columns)
    agent_type        TEXT NOT NULL,
    step_name         TEXT NOT NULL,
    model_filter      TEXT,                    -- NULL = no filter

-- Counts
    rows_seen         INTEGER NOT NULL DEFAULT 0,
    rows_exported     INTEGER NOT NULL DEFAULT 0,
    rows_skipped      JSONB NOT NULL DEFAULT '{}'::jsonb,  -- {invalid_json: N, scan_error: N, ...}

-- Dataset shape
    format            TEXT NOT NULL DEFAULT 'chatml',
    export_version    TEXT NOT NULL DEFAULT '1',
    size_bytes        BIGINT NOT NULL DEFAULT 0,           -- sum of messages+metadata JSONB sizes

-- Provenance
    triggered_by      UUID,                    -- optional caller agent id
    orchestration_id  UUID,                    -- the orchestration that produced this export
    source_notes      TEXT,                    -- free-form, e.g. "backfilled from JSONL 2026-04-23"

-- Timestamps
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    completed_at      TIMESTAMPTZ              -- NULL if the export didn't complete cleanly
    );

-- Finding recent exports for a given agent/step combo
CREATE INDEX IF NOT EXISTS idx_training_exports_runs_agent_step
    ON training_exports.runs(agent_type, step_name, created_at DESC);

-- Finding exports by model filter (for comparing models side-by-side)
CREATE INDEX IF NOT EXISTS idx_training_exports_runs_model
    ON training_exports.runs(model_filter, created_at DESC)
    WHERE model_filter IS NOT NULL;


-- ---------------------------------------------------------------------------
-- rows — one row per training record
-- ---------------------------------------------------------------------------
-- Each row is a complete training example (prompt + assistant response +
-- metadata). messages and metadata are stored as JSONB so we can query
-- into them if needed (e.g. filter by metadata.model at training time,
-- even within a single export).
--
-- ON DELETE CASCADE: dropping a run cleans up its rows automatically.
-- ---------------------------------------------------------------------------

CREATE TABLE IF NOT EXISTS training_exports.rows (
                                                     id         BIGSERIAL PRIMARY KEY,
                                                     export_id  UUID NOT NULL REFERENCES training_exports.runs(id) ON DELETE CASCADE,
    row_index  INTEGER NOT NULL,               -- 0-based position within the export

    messages   JSONB NOT NULL,                 -- ChatML: [{role, content}, {role, content}]
    metadata   JSONB NOT NULL                  -- source_log_id, agent_type, created_at, model, etc.
    );

-- Reading rows in export order
CREATE INDEX IF NOT EXISTS idx_training_exports_rows_export_index
    ON training_exports.rows(export_id, row_index);

-- Uniqueness: within an export, each source_log_id should appear once
-- (metadata.source_log_id is the unique key of the original llm_call_log row).
-- This prevents accidental duplicates if an export is re-run or resumed.
CREATE UNIQUE INDEX IF NOT EXISTS ux_training_exports_rows_export_source
    ON training_exports.rows(export_id, (metadata->>'source_log_id'));


-- ---------------------------------------------------------------------------
-- Convenience view — recent exports summary
-- ---------------------------------------------------------------------------

CREATE OR REPLACE VIEW training_exports.recent_runs AS
SELECT
    r.id,
    r.agent_type,
    r.step_name,
    r.model_filter,
    r.rows_exported,
    r.rows_seen,
    r.size_bytes,
    r.format,
    r.export_version,
    r.created_at,
    r.completed_at,
    (r.completed_at - r.created_at) AS duration
FROM training_exports.runs r
ORDER BY r.created_at DESC
    LIMIT 100;


-- ---------------------------------------------------------------------------
-- Verification
-- ---------------------------------------------------------------------------

SELECT table_schema, table_name
FROM information_schema.tables
WHERE table_schema = 'training_exports'
ORDER BY table_name;

SELECT indexname, indexdef
FROM pg_indexes
WHERE schemaname = 'training_exports'
ORDER BY indexname;
