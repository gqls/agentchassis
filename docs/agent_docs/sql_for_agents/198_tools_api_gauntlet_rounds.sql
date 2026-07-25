-- ============================================================
-- Migration : 198_tools_api_gauntlet_rounds.sql
-- Purpose   : Create gauntlet_rounds table for tools-api
--             Stores each public visitor debate round.
-- RE-VERIFIED: site_work_items (69-value item_type enum) has no
--             entry for a public visitor debate interaction and
--             lacks a client_ip_hash column — this is a genuinely
--             new table, not a duplicate of any existing structure.
-- Applied by: human migration runner (image-first-then-seed)
-- ============================================================

ASSERT (SELECT count(*) FROM information_schema.tables
        WHERE table_schema = 'public'
          AND table_name   = 'gauntlet_rounds') = 0,
       'Migration 198: gauntlet_rounds already exists — skipping';

CREATE TABLE gauntlet_rounds (
    id             uuid        PRIMARY KEY DEFAULT gen_random_uuid(),
    site_id        uuid        NOT NULL REFERENCES sites(id),
    tool           text        NOT NULL DEFAULT 'gauntlet',
    provocation    jsonb,
    position_text  text,
    counter        jsonb,
    defence_text   text,
    verdict        jsonb,
    client_ip_hash text,
    created_at     timestamptz NOT NULL DEFAULT now(),
    updated_at     timestamptz NOT NULL DEFAULT now()
);

-- Speed up per-site round listing, newest first.
CREATE INDEX gauntlet_rounds_site_id_created_at_idx
    ON gauntlet_rounds (site_id, created_at DESC);

-- Ledger entry
-- table : gauntlet_rounds
-- cols  : id, site_id, tool, provocation, position_text,
--         counter, defence_text, verdict, client_ip_hash,
--         created_at, updated_at
-- fk    : site_id -> sites(id)
-- index : (site_id, created_at DESC)
