-- 192_model_directory_schema.sql — 2026-07-22, model_directory_pipeline Phase A
--
-- Context: fleet-wide, continually-updated AI model directory (open+closed
-- models: what they do, cited cost, owner, where to find/use them), to be
-- followed by a company AI-agent adoption tracker on the same schema (new
-- `kind`, new `field` values — no further migration expected for that).
-- Full design: docs/agent_docs/docs024_key_docs_latest/model_directory_pipeline/
-- PLAN_2026-07-22_model_directory_pipeline.md.
--
-- Two tables, deliberately kept separate:
--   directory_entities — the thing (a model; later a company or protocol).
--   directory_claims   — individually CITED facts about a thing (a price, a
--     licence, a claimed ROI %). Cost/spec facts are claims, never a jsonb
--     convenience field on the entity — an uncited "attributes" blob is
--     exactly the kind of fact that goes stale silently, which this split
--     exists to catch. `directory_entities.attributes` is reserved for
--     genuinely structural/filtering metadata that is not itself a
--     verifiable claim (modality tags, category, logo URL).
--
-- Versioning mirrors the site_specs / doc_plans convention already live in
-- this schema: no integer version column, "current" is is_current=true,
-- enforced by a partial unique index, history kept via superseded_at. A
-- claim that fails re-verification gets a NEW current row (status
-- citation_lost), the old one superseded — never a silent in-place edit.
--
-- citation jsonb shape matches datahelpers.Citation exactly (publisher,
-- title, url, quote, accessed, published) so the existing deterministic
-- verifier (evidence_citations.go / datahelpers.QuoteFoundInText) can be
-- reused unchanged in Phase B — this migration does not touch that code.
--
-- No seed data (owner decision: pipeline before publish — no hand-curated
-- launch content). This migration is schema-only.
--
-- Verify after applying:
--   \d directory_entities
--   \d directory_claims
--   -- partial-unique-index behaviour: insert two is_current rows for the
--   -- same (entity_id, field) and confirm the second is rejected.
--
-- Rollback: DROP TABLE IF EXISTS directory_claims; DROP TABLE IF EXISTS directory_entities;

\set ON_ERROR_STOP on
BEGIN;

CREATE TABLE directory_entities (
    id            uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    kind          text NOT NULL,                    -- 'model' | later 'company' | 'protocol'
    slug          text NOT NULL,
    name          text NOT NULL,
    owner         text,                              -- who runs/owns it
    summary       text,
    links         jsonb NOT NULL DEFAULT '{}',       -- {docs, weights, wrapper_url, video_urls: [...]}
    attributes    jsonb NOT NULL DEFAULT '{}',        -- structural/filtering metadata ONLY — not a verifiable claim
    status        text NOT NULL DEFAULT 'active',    -- active | archived | superseded
    discovered_by text,
    created_at    timestamptz NOT NULL DEFAULT now(),
    updated_at    timestamptz NOT NULL DEFAULT now(),
    UNIQUE (kind, slug)
);

COMMENT ON TABLE directory_entities IS
    'A directory subject: an AI model now, later a company or protocol (kind). Structural/filtering metadata only in attributes — verifiable facts belong in directory_claims.';

CREATE INDEX idx_directory_entities_kind ON directory_entities (kind, status);

CREATE TABLE directory_claims (
    id             uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    entity_id      uuid NOT NULL REFERENCES directory_entities(id),
    field          text NOT NULL,                    -- 'price_input_per_mtok' | 'context_window' | 'license' | 'roi_pct' (later) | ...
    value          text,
    unit           text,
    citation       jsonb NOT NULL,                   -- {publisher,title,url,quote,accessed,published} — datahelpers.Citation shape
    status         text NOT NULL DEFAULT 'found',    -- found | citation_lost | fetch_error | pending
    staleness_days integer NOT NULL DEFAULT 200,
    is_current     boolean NOT NULL DEFAULT true,
    superseded_at  timestamptz,
    verified_at    timestamptz,
    created_by     text,
    created_at     timestamptz NOT NULL DEFAULT now()
);

COMMENT ON TABLE directory_claims IS
    'One cited, individually re-verifiable fact about a directory_entities row. One is_current row per (entity_id, field); a failed re-verification supersedes and inserts a new row with status citation_lost, never edits in place.';

CREATE UNIQUE INDEX idx_directory_claims_current
    ON directory_claims (entity_id, field)
    WHERE is_current;

CREATE INDEX idx_directory_claims_stale ON directory_claims (verified_at) WHERE is_current;

DO $$
BEGIN
    IF to_regclass('public.directory_entities') IS NULL THEN
        RAISE EXCEPTION '192: directory_entities was not created';
    END IF;
    IF to_regclass('public.directory_claims') IS NULL THEN
        RAISE EXCEPTION '192: directory_claims was not created';
    END IF;
    IF to_regclass('public.idx_directory_claims_current') IS NULL THEN
        RAISE EXCEPTION '192: idx_directory_claims_current was not created';
    END IF;
END $$;

COMMIT;
