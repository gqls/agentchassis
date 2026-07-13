-- 0NN_diagnosis_artifacts.sql — F0.1a of the diagnosis→fix loop.
-- 2026-07-09. Renumber 0NN to the next migration number when filing under
-- platform/database/migrations/ (that directory is hand-applied; the repo's
-- scripts/migration/run-migrations.sh is empty, so there is no version table
-- to update).
--
-- Pre-flight: verify_before_migration_diagnosis_artifacts.sql returned clean
-- on clients_db 2026-07-09 (no table, no index-name collisions, no CHECK on
-- site_work_items.pipeline; live pipelines: build/content/design/maintenance,
-- so the 'diagnose' namespace of F0.1c is free).
--
-- Design: Q-A, decided 2026-07-07. Each diagnosis iteration's bundle is today
-- ephemeral and in-memory; the loop can neither be audited nor replayed. This
-- table is the durable egress. The write happens INSIDE
-- DiagnoseAssembleBundleAction (Go) rather than as a new workflow step, so the
-- diagnose-agent workflow's shape — emit → persist_note → complete, an active
-- surface owned by the tools chat — is not touched at all.
--
-- Unified table with kind ∈ {bundle, iteration_note} per the (9) refinement:
-- start unified, split only if retention diverges. It has not yet.
--
-- Relationship to doc_notes: doc_notes holds the TERMINAL, human-facing
-- diagnosis note, written by the tools chat's persist_note wiring. This table
-- holds the machine-replayable per-iteration evidence. Different readers,
-- different retention. We write nothing into doc_notes.

BEGIN;

CREATE TABLE IF NOT EXISTS diagnosis_artifacts (
    id              uuid PRIMARY KEY DEFAULT gen_random_uuid(),

    -- Correlation, not orchestration, is the loop's identity: one diagnosis run
    -- spans many orchestration steps. ExecutionContext.CorrelationID is a string,
    -- not necessarily a uuid, so text — do not tighten this to uuid.
    correlation_id  text NOT NULL,
    orchestration_id text,                        -- convenience join to orchestration_states

    iteration       integer NOT NULL CHECK (iteration >= 1),
    kind            text NOT NULL CHECK (kind IN ('bundle', 'iteration_note')),
    body            text NOT NULL,                -- bundle markdown, or the note prose

    -- Null for a site-less (code-only) diagnosis. The 'diagnose' pipeline
    -- namespace allows anchorless runs; load_runtime already survives them.
    site_id         uuid,

    -- symbol_count, truncated, scope, hypothesis, step name for an
    -- iteration_note. Kept as jsonb so the shape can move without a migration.
    metadata        jsonb NOT NULL DEFAULT '{}'::jsonb,

    source_agent    text,
    created_by      text,
    created_at      timestamptz NOT NULL DEFAULT now(),

    -- Retention knob, per kind (Q-A). A sweep deletes expired, unpinned rows;
    -- bundles are large (~60KB × ≤5 iterations) and expire sooner than notes.
    -- NULL expires_at = keep indefinitely. pinned always wins.
    expires_at      timestamptz,
    pinned          boolean NOT NULL DEFAULT false
);

-- The read path: "give me every artifact for this run, in order".
CREATE INDEX IF NOT EXISTS idx_diagnosis_artifacts_correlation
    ON diagnosis_artifacts (correlation_id, iteration, kind);

-- Exactly one bundle per (run, iteration). A workflow step retry re-runs
-- assemble for the same iteration; the Go write-through upserts on this index
-- so a retry replaces rather than duplicates. Notes are deliberately NOT
-- covered: per-STEP notes mean several rows per iteration (F0.3).
CREATE UNIQUE INDEX IF NOT EXISTS idx_diagnosis_artifacts_bundle_current
    ON diagnosis_artifacts (correlation_id, iteration)
    WHERE kind = 'bundle';

-- The retention sweep's index. Partial: only rows that can ever expire.
CREATE INDEX IF NOT EXISTS idx_diagnosis_artifacts_expiry
    ON diagnosis_artifacts (expires_at)
    WHERE expires_at IS NOT NULL AND NOT pinned;

COMMENT ON TABLE diagnosis_artifacts IS
    'Durable per-iteration egress for the diagnosis loop: kind=bundle is the machine-replayable evidence bundle the verdict step read; kind=iteration_note is the loop''s own reasoning trail (hypothesis, scope chosen and why, requests issued, verdict grounds). Written through from inside diagnose_assemble_bundle (Go), not by a workflow step. The terminal human-facing note lives in doc_notes instead.';
COMMENT ON COLUMN diagnosis_artifacts.correlation_id IS
    'ExecutionContext.CorrelationID — the identity of one diagnosis RUN across all its orchestration steps. Text, not uuid: the chassis does not guarantee uuid form.';
COMMENT ON COLUMN diagnosis_artifacts.iteration IS
    '1-based. Derived in assemble as route.diagnose_state.iteration + 1, defaulting to 1 on the first pass when diagnose_route has not yet run.';
COMMENT ON COLUMN diagnosis_artifacts.expires_at IS
    'Retention knob. NULL = keep indefinitely. Sweep deletes WHERE expires_at < now() AND NOT pinned. Bundles expire sooner than notes.';

COMMIT;

-- Rollback (manual):
--   DROP TABLE IF EXISTS diagnosis_artifacts;
--   (indexes drop with the table)
