-- ============================================================================
-- Migration 086: site_chat_turns — per-domain chatbot prompt/answer log
-- ============================================================================
-- Records each end-user chat turn (one prompt + one answer) from the site
-- chatbot edge worker. Distinct from llm_call_log, which is the build-time
-- training flywheel: this table is end-user owned, carries user-submitted text
-- (PII), and has its own retention/access profile (per-site analytics for the
-- site owner). Token/latency column names are aligned to llm_call_log on
-- purpose so the two LLM-logging tables read the same way.
--
-- Population path: the edge worker writes each turn to its sink (queue or edge
-- store); a Layer-1 puller drains that sink and inserts here. The worker
-- supplies `id` (its own turn uuid) so re-ingest is idempotent via
-- ON CONFLICT (id) DO NOTHING. `created_at` is the edge timestamp (when the turn
-- happened); `ingested_at` is when Layer 1 stored it.
--
-- NOTE ON NUMBERING: this snapshot only shows migrations up to 085. Confirm the
-- next free migration number AND the file-prefix against the live migrations
-- directory before applying; renumber if 086 is taken.
-- ============================================================================

CREATE TABLE IF NOT EXISTS site_chat_turns (
    -- Edge-supplied turn uuid (PK enables idempotent ingest). Falls back to a
    -- generated uuid if the puller does not supply one.
    id                uuid                     NOT NULL DEFAULT gen_random_uuid(),

    site_id           uuid                     NOT NULL,
    domain            character varying(255)   NOT NULL,   -- denormalised, matches sites.domain
    session_id        character varying(255),              -- opaque client-generated session id

    -- The conversation turn. Treat question/answer as PII.
    question          text                     NOT NULL,
    answer            text                     NOT NULL,

    -- Bounding outcomes
    refused           boolean                  NOT NULL DEFAULT false,  -- hit the off-topic refusal path
    capped            boolean                  NOT NULL DEFAULT false,  -- hit a turn/length limit

    -- Provenance / debugging
    model             character varying(100),
    pack_version      integer,                              -- context-pack version that produced the answer
    grounding_ids     text[],                               -- chunk ids fed into the prompt, for "why did it say that"

    -- Cost / performance (names aligned to llm_call_log)
    input_tokens      integer,
    output_tokens     integer,
    latency_ms        integer,

    -- Abuse handling without storing raw IPs (store a salted hash, not the IP).
    client_ip_hash    text,

    error_message     text,                                 -- set if the turn failed at the edge

    created_at        timestamp with time zone NOT NULL DEFAULT now(),  -- edge time (turn happened)
    ingested_at       timestamp with time zone NOT NULL DEFAULT now(),  -- Layer 1 store time

    CONSTRAINT site_chat_turns_pkey PRIMARY KEY (id),

    -- Deleting a site removes its chat history. If audit-retention beyond site
    -- deletion is wanted instead, make site_id nullable and switch to
    -- ON DELETE SET NULL.
    CONSTRAINT site_chat_turns_site_id_fkey
        FOREIGN KEY (site_id) REFERENCES sites(id) ON DELETE CASCADE
);

-- Per-site analytics / recent-first listing
CREATE INDEX IF NOT EXISTS idx_site_chat_turns_site
    ON site_chat_turns (site_id, created_at DESC);

-- Lookup by domain (the worker keys on Host)
CREATE INDEX IF NOT EXISTS idx_site_chat_turns_domain
    ON site_chat_turns (domain, created_at DESC);

-- Reconstruct a single conversation in order
CREATE INDEX IF NOT EXISTS idx_site_chat_turns_session
    ON site_chat_turns (session_id, created_at)
    WHERE session_id IS NOT NULL;

-- Review off-topic / refused attempts (scope tuning)
CREATE INDEX IF NOT EXISTS idx_site_chat_turns_refused
    ON site_chat_turns (site_id, created_at DESC)
    WHERE refused;

-- Surface failed turns
CREATE INDEX IF NOT EXISTS idx_site_chat_turns_errors
    ON site_chat_turns (created_at DESC)
    WHERE error_message IS NOT NULL;

COMMENT ON TABLE site_chat_turns IS
    'End-user chatbot prompt/answer turns, per domain. Populated by the Layer-1 puller draining the edge worker sink. Separate from llm_call_log (build-time flywheel).';
COMMENT ON COLUMN site_chat_turns.id IS
    'Edge-supplied turn uuid; PK enables idempotent ingest via ON CONFLICT (id) DO NOTHING.';
COMMENT ON COLUMN site_chat_turns.client_ip_hash IS
    'Salted hash of client IP for abuse detection. Do NOT store raw IPs (GDPR).';
