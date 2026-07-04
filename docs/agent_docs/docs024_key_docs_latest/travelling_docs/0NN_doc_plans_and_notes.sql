-- 0NN_doc_plans_and_notes.sql — travelling docs: PLAN + NOTES tables.
-- DRAFT 2026-07-04. Renumber 0NN to the next migration number.
-- DO NOT APPLY until verify_before_migration.sql returns clean
-- (no doc_plans/doc_notes collision; pipeline value set confirmed).
--
-- Design: PLAN_travelling_docs.md rev 4. Truth in Postgres; knowledge_base is
-- the derived RAG index; git mirror optional. Versioning reuses the
-- site_specs supersede pattern re-keyed to (subject_type, subject_key).
-- subject_key: tool  -> content_components.function (byte-for-byte)
--              pipeline -> a site_work_items.pipeline value (unconstrained
--              text in that table, so no FK/CHECK here either — convention).

BEGIN;

-- ── doc_plans — the intent (versioned; one current row per subject) ─────────
CREATE TABLE doc_plans (
    id             uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    subject_type   text NOT NULL CHECK (subject_type IN ('tool', 'pipeline')),
    subject_key    text NOT NULL,
    body           text NOT NULL,                 -- PLAN markdown; may embed a ```criteria json block
    source         text,                          -- 'tool-generator' | 'human' | 'migration' | ...
    source_agent   text,
    source_item_id uuid,
    notes          text,                          -- why this version superseded the last
    is_current     boolean NOT NULL DEFAULT true,
    pinned         boolean NOT NULL DEFAULT false,
    created_by     text,
    created_at     timestamptz NOT NULL DEFAULT now(),
    superseded_at  timestamptz,
    updated_at     timestamptz NOT NULL DEFAULT now()
);

-- One current PLAN per subject (partial unique, mirrors site_specs is_current).
CREATE UNIQUE INDEX idx_doc_plans_current
    ON doc_plans (subject_type, subject_key)
    WHERE is_current;

CREATE INDEX idx_doc_plans_subject ON doc_plans (subject_type, subject_key, created_at DESC);

COMMENT ON TABLE doc_plans IS
    'Travelling PLAN docs (intent) for tools/complex components and pipelines. Supersede-versioned; one is_current row per (subject_type, subject_key). Truth store; knowledge_base holds a derived RAG copy.';
COMMENT ON COLUMN doc_plans.subject_key IS
    'tool: content_components.function; pipeline: site_work_items.pipeline value.';

-- ── doc_notes — the history (append-only; one row per entry) ────────────────
CREATE TABLE doc_notes (
    id             uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    subject_type   text NOT NULL CHECK (subject_type IN ('tool', 'pipeline')),
    subject_key    text NOT NULL,
    site_id        uuid,                          -- set for a per-site incident; NULL = library/global
    body           text NOT NULL,                 -- Observed / Root cause / Fix / Verified (or diagnosis report)
    categories     jsonb NOT NULL DEFAULT '[]'::jsonb,
    source         text,                          -- 'tool-improver' | 'diagnosis-loop' | 'migration' | ...
    source_agent   text,
    source_item_id uuid,
    created_by     text,
    created_at     timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX idx_doc_notes_subject ON doc_notes (subject_type, subject_key, created_at DESC);

-- jsonb_ops (default opclass) — NOT jsonb_path_ops — so `categories ? 'tag'`
-- (existence) is index-backed for the cross-subject roll-up into 016/016b.
CREATE INDEX idx_doc_notes_categories ON doc_notes USING gin (categories jsonb_ops);

CREATE INDEX idx_doc_notes_site ON doc_notes (site_id) WHERE site_id IS NOT NULL;

COMMENT ON TABLE doc_notes IS
    'Travelling NOTES (append-only history) for tools/complex components and pipelines. One row per entry; categories tagged for roll-up. Includes persisted diagnosis reports (incl. unconfirmed dead ends).';

COMMIT;

-- Rollback (manual):
--   DROP TABLE IF EXISTS doc_notes; DROP TABLE IF EXISTS doc_plans;
