-- ============================================================================
-- 019_model_lifecycle_schema.sql
-- ============================================================================
-- Tables tracking what happens to training_exports.runs:
--   training_runs   — one row per QLoRA run, links back to its source export
--   artefacts       — LoRA adapter binaries, registered after upload to S3
--   evaluations    — flywheel D evaluation results per (artefact, eval_set, judge)
--
-- Why a new schema rather than extending training_exports:
--   training_exports.* captures "data exports we produced" and is read mostly
--   at training-job-creation time. model_lifecycle.* is read every flywheel
--   cycle and after every adapter we ship, so it earns its own namespace.
--
-- Style matches training_exports (UPPERCASE types, IF NOT EXISTS, comments).
-- ============================================================================

CREATE SCHEMA IF NOT EXISTS model_lifecycle;


-- ---------------------------------------------------------------------------
-- training_runs — one row per QLoRA training run
-- ---------------------------------------------------------------------------
-- Created in 'pending' state when a training-flywheel-orchestrator dispatches
-- work to the model-trainer agent. Updated to 'running' when the Thunder VM
-- starts training, 'complete' when the adapter is uploaded and registered,
-- 'failed' if any step errored.
--
-- hyperparameters captures everything needed to reproduce the run — base_model,
-- epochs, batch, grad_accum, lr, lora_r, lora_alpha, max_seq, seed, plus any
-- new knobs we add in iter_N. Stored as JSONB so we can query/index it
-- without altering the schema as the surface evolves.
--
-- thunder_instance_id is captured for post-mortem only — Thunder ephemeral
-- instances are deleted after each run, so this is just a breadcrumb.
-- ---------------------------------------------------------------------------

CREATE TABLE IF NOT EXISTS model_lifecycle.training_runs (
    id                      UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    -- Source data
    export_id               UUID NOT NULL REFERENCES training_exports.runs(id),

    -- Lifecycle state
    status                  TEXT NOT NULL DEFAULT 'pending'
        CHECK (status IN ('pending', 'running', 'complete', 'failed')),

    -- Reproducibility — full hyperparameter set (see action input spec)
    hyperparameters         JSONB NOT NULL,

    -- Timing
    started_at              TIMESTAMPTZ,
    completed_at            TIMESTAMPTZ,
    train_runtime_s         NUMERIC,                -- from manifest, NULL until complete

    -- Outcome metrics
    final_loss              NUMERIC,                -- trailing-average training loss
    peak_vram_gb            NUMERIC,                -- to validate VRAM headroom for next run

    -- Cost (USD) — from instance_uptime × hourly_rate; helps the fleet planner
    cost_usd                NUMERIC,

    -- Failure handling
    error_message           TEXT,                   -- populated only when status='failed'

    -- Provenance
    thunder_instance_id     TEXT,                   -- breadcrumb for post-mortem (instance is gone)
    triggered_by            UUID,                   -- agent that dispatched the run
    orchestration_id        UUID,                   -- the orchestration that owns this run

    -- Timestamps
    created_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Look up runs by export to answer "what did we train from this dataset?"
CREATE INDEX IF NOT EXISTS idx_model_lifecycle_training_runs_export
    ON model_lifecycle.training_runs(export_id, created_at DESC);

-- Find still-running or failed runs (operational queries)
CREATE INDEX IF NOT EXISTS idx_model_lifecycle_training_runs_status
    ON model_lifecycle.training_runs(status, created_at DESC)
    WHERE status IN ('pending', 'running', 'failed');


-- ---------------------------------------------------------------------------
-- artefacts — LoRA adapter binaries
-- ---------------------------------------------------------------------------
-- One row per uploaded adapter. Decoupled from training_runs so we can
-- support requantization (e.g. fp16 → int8) or repacking (e.g. PEFT → GGUF)
-- as separate artefact rows referring to the same training run.
--
-- storage_uri uses the URI scheme (s3://bucket/key, file:///path, etc.)
-- rather than encoding bucket+key separately — keeps storage backend
-- pluggable.
--
-- format names the on-disk shape, not just the file extension. e.g.
-- 'lora_safetensors_fp16' is meaningfully different from 'lora_safetensors_fp32'
-- because tools that load the adapter need to know the dtype.
-- ---------------------------------------------------------------------------

CREATE TABLE IF NOT EXISTS model_lifecycle.artefacts (
    id                      UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    training_run_id         UUID NOT NULL REFERENCES model_lifecycle.training_runs(id),

    -- Storage
    storage_uri             TEXT NOT NULL,          -- s3://bucket/path/adapter_model.safetensors
    sha256                  TEXT NOT NULL,          -- integrity check at load time
    size_bytes              BIGINT NOT NULL,

    -- Shape — controls how downstream code loads it
    format                  TEXT NOT NULL,          -- 'lora_safetensors_fp16', 'gguf_q4', etc.

    -- Metadata
    base_model              TEXT NOT NULL,          -- denormalized from training_runs.hyperparameters for query convenience
    notes                   TEXT,                   -- free-form, e.g. "merged adapter for inference"

    created_at              TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Find artefacts for a given run (typically 1, sometimes more if requantized)
CREATE INDEX IF NOT EXISTS idx_model_lifecycle_artefacts_run
    ON model_lifecycle.artefacts(training_run_id);

-- Find artefacts by base model (to compare adapters across iter_N for the same base)
CREATE INDEX IF NOT EXISTS idx_model_lifecycle_artefacts_base_model
    ON model_lifecycle.artefacts(base_model, created_at DESC);


-- ---------------------------------------------------------------------------
-- evaluations — flywheel D results per (artefact, eval_set, judge_model)
-- ---------------------------------------------------------------------------
-- Multiple eval rows can exist for the same artefact: same eval set with
-- different judges (Opus, Sonnet, human), or different eval sets (held_out_v1,
-- held_out_v2, novel_industry_v1).
--
-- l1_metrics and l2_metrics keep the evaluation pipeline outputs as JSONB
-- so we don't need a schema migration every time we add a metric. The
-- contract is enforced by build_report.py and the evaluator agent, not the
-- database.
--
-- deployment_decision is intentionally a free-text TEXT field rather than
-- a CHECK constraint — different sites have different shipping bars and we
-- don't want to lock the vocabulary prematurely. Examples: 'shipped_internal',
-- 'shipped_low_stakes', 'rejected_voice_gap', 'pending_review'.
-- ---------------------------------------------------------------------------

CREATE TABLE IF NOT EXISTS model_lifecycle.evaluations (
    id                      UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    artefact_id             UUID NOT NULL REFERENCES model_lifecycle.artefacts(id),

    -- What we evaluated against
    eval_set_uri            TEXT NOT NULL,          -- e.g. 'flywheel_D/held_out_cases_v1.jsonl'
    eval_set_n_cases        INTEGER NOT NULL,
    judge_model             TEXT NOT NULL,          -- e.g. 'claude-opus-4-7'

    -- Results
    l1_metrics              JSONB NOT NULL,         -- structural — schema match, length, forbidden, etc.
    l2_metrics              JSONB NOT NULL,         -- judge — head-to-head, mean dim scores, position bias
    report_uri              TEXT,                   -- where the human-readable report lives

    -- Decision (set by human after review; nullable until then)
    deployment_decision     TEXT,                   -- free-text vocabulary, see docstring
    deployment_decision_at  TIMESTAMPTZ,
    deployment_decision_by  TEXT,                   -- email / handle of decision-maker

    -- Timestamps
    created_at              TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Find evals for an artefact (to compare across judges or eval sets)
CREATE INDEX IF NOT EXISTS idx_model_lifecycle_evaluations_artefact
    ON model_lifecycle.evaluations(artefact_id, created_at DESC);

-- Find evaluations by judge across artefacts (to track judge drift over iter_N)
CREATE INDEX IF NOT EXISTS idx_model_lifecycle_evaluations_judge
    ON model_lifecycle.evaluations(judge_model, created_at DESC);

-- Find pending decisions (operational query)
CREATE INDEX IF NOT EXISTS idx_model_lifecycle_evaluations_pending_decision
    ON model_lifecycle.evaluations(created_at DESC)
    WHERE deployment_decision IS NULL;


-- ---------------------------------------------------------------------------
-- View: deployable_adapters — adapters whose latest evaluation greenlit them
-- ---------------------------------------------------------------------------
-- The chassis can read this view directly to find the current canonical
-- adapter to load. We treat 'shipped_*' decisions as deployable (anything
-- starting with 'shipped' counts) and pick the most recently created.
-- ---------------------------------------------------------------------------

CREATE OR REPLACE VIEW model_lifecycle.deployable_adapters AS
SELECT DISTINCT ON (a.base_model)
    a.id                AS artefact_id,
    a.base_model,
    a.storage_uri,
    a.format,
    a.sha256,
    a.size_bytes,
    a.created_at        AS adapter_created_at,
    e.deployment_decision,
    e.deployment_decision_at,
    e.report_uri
FROM model_lifecycle.artefacts a
JOIN model_lifecycle.evaluations e ON e.artefact_id = a.id
WHERE e.deployment_decision LIKE 'shipped_%'
ORDER BY a.base_model, a.created_at DESC;


-- ---------------------------------------------------------------------------
-- View: latest_training_run_per_export — for "what's the current iter_N?"
-- ---------------------------------------------------------------------------

CREATE OR REPLACE VIEW model_lifecycle.latest_training_run_per_export AS
SELECT DISTINCT ON (export_id)
    id                  AS training_run_id,
    export_id,
    status,
    final_loss,
    train_runtime_s,
    cost_usd,
    completed_at,
    created_at
FROM model_lifecycle.training_runs
ORDER BY export_id, created_at DESC;


-- ---------------------------------------------------------------------------
-- updated_at trigger for training_runs (other tables are append-mostly)
-- ---------------------------------------------------------------------------

CREATE OR REPLACE FUNCTION model_lifecycle.set_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trg_model_lifecycle_training_runs_updated_at
    ON model_lifecycle.training_runs;

CREATE TRIGGER trg_model_lifecycle_training_runs_updated_at
    BEFORE UPDATE ON model_lifecycle.training_runs
    FOR EACH ROW
    EXECUTE FUNCTION model_lifecycle.set_updated_at();


-- ============================================================================
-- Sanity checks (run after applying)
-- ============================================================================
-- \dn model_lifecycle
-- \dt model_lifecycle.*
-- \dv model_lifecycle.*
-- \di model_lifecycle.*
