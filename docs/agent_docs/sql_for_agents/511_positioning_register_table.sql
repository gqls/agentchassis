-- 511 — the positioning register becomes a DATABASE table.
--
-- OWNER RULING 2026-08-19 (RFC_037): *"The data should be in a database. The
-- database should be the source of truth."* — the larger of the two options
-- offered, taken deliberately. `REGISTER_positioning.md` stops being
-- authoritative and becomes a rendering of, an input to, or retired in favour of
-- this table.
--
-- AND THE RULING THAT SIZES IT: *"For the 40 non finance sites add a registry,
-- also for the rest of the 2000 .uk domains."* (~1,500 by the owner's later
-- correction). So this is not a table for 44 finance entries; it is a table for
-- the whole estate, and 44 is where it starts.
--
-- ── THE ONE DESIGN DECISION THAT MATTERS ─────────────────────────────────────
--
-- `raw_md` HOLDS THE WHOLE ORIGINAL ENTRY, AND IT IS AUTHORITATIVE.
--
-- The register is a hand-written document whose value is its REASONING — why two
-- domains are two businesses, what separates them, what each must never drift
-- into. A parser can extract `audience` and `mode` reliably (46/49 and 31/49 of
-- the current entries carry them as labelled fields), but the argument lives in
-- prose that no schema anticipates: 49 entries use 18 different field names, and
-- `owns:` appears in exactly ONE of them while the same idea appears as prose in
-- dozens.
--
-- **A lossy parse would therefore not be a migration; it would be a deletion.**
-- Making the database the source of truth is only safe if nothing is lost in
-- getting there. So the columns below are a convenience INDEX over the entry —
-- for querying, for the collision invariant, for feeding a prompt — and `raw_md`
-- is the entry. When the two disagree, `raw_md` wins, and a reader that needs to
-- be right consults it.
--
-- This is the same discipline `EvidenceBase` failed at (LANDMINES: parsing
-- `evidence_base` through its typed struct and writing it back DELETES every
-- citation and writer_line the struct does not model). The lesson there arrived
-- after the damage; here it is designed in.
--
-- ── WHAT IS NOT DECIDED HERE ─────────────────────────────────────────────────
--   * the markdown file's fate — generated view, one-time input, or retired. The
--     ruling says it stops being the source of truth; which of the three it
--     becomes is a separate call, and until it is made BOTH must not be edited
--     (two hand-maintained copies of one roster is the drift class
--     `099_SYNC_gate_roster.py` exists to prevent).
--   * neighbour SELECTION at scale. At 44 entries neighbours are hand-named; at
--     1,500 they cannot be, so `neighbours` here records what the entry SAYS and
--     a rule for deriving them is still owed (RFC_037 open).
--   * the collision invariant as a CONSTRAINT. `(family, audience, mode)` is the
--     register's stated rule, but the current entries do not populate those three
--     consistently enough for a UNIQUE index to be anything but a blocker. A
--     daily check is the honest first step; recorded as a residual, not built.
--
-- Rollback: 511_positioning_register_table_ROLLBACK.sql

BEGIN;

CREATE TABLE IF NOT EXISTS positioning_register (
    id              uuid PRIMARY KEY DEFAULT gen_random_uuid(),

    -- One row per DOMAIN, not per entry: a twin needs to be findable by its own
    -- name (that is the whole point of recording twins), and the brief writer is
    -- given a domain, never an entry code.
    domain          text NOT NULL,
    entry_code      text,               -- M1, L9, G3 … the human handle
    family          text,               -- MORTGAGE, LOAN/DEBT, GIFTS/SEASONAL …
    is_primary      boolean NOT NULL DEFAULT true,
    primary_domain  text,               -- set on a twin row, pointing at its primary

    -- The convenience index. Nullable throughout, deliberately: an entry that
    -- does not state its `mode` must not be forced to invent one, and a NULL
    -- here is a truthful "the entry does not say".
    proposition     text,
    audience        text,
    stage           text,
    mode            text,
    stance          text,

    -- What the entry says about its neighbours: [{"domain"|"code": …, "rule": …}]
    neighbours      jsonb NOT NULL DEFAULT '[]'::jsonb,
    must_nots       jsonb NOT NULL DEFAULT '[]'::jsonb,

    -- AUTHORITATIVE. See the header. Never NULL for a parsed entry.
    raw_md          text,

    -- Status vocabulary kept open text rather than an enum: the register's own
    -- legend already carries LIVE / BUILT / ADOPTED / HOLD / REMAKE / — and it
    -- has gained values repeatedly. An enum would make the next one a migration.
    status          text,

    -- Owner ruling 2026-08-20 SIMPLIFIED this: the 50 test domains need "no
    -- special status, just keep them out of the list to build for now". So this
    -- is a build-queue exclusion, NOT the elaborate reserved-state the earlier
    -- design proposed. A row may exist for a test domain; it simply must not be
    -- picked up for a build.
    exclude_from_build boolean NOT NULL DEFAULT false,
    exclude_reason  text,

    source_file     text,
    parsed_at       timestamptz,
    created_at      timestamptz NOT NULL DEFAULT now(),
    updated_at      timestamptz NOT NULL DEFAULT now(),
    created_by      text NOT NULL DEFAULT 'unknown'
);

-- A domain appears once. This is the constraint that makes the table a register
-- rather than a log, and it is what lets the loader be idempotent (upsert by
-- domain) instead of accumulating duplicates on every re-run.
CREATE UNIQUE INDEX IF NOT EXISTS idx_positioning_register_domain
    ON positioning_register (lower(domain));

CREATE INDEX IF NOT EXISTS idx_positioning_register_family
    ON positioning_register (family) WHERE family IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_positioning_register_entry
    ON positioning_register (entry_code) WHERE entry_code IS NOT NULL;

-- Partial, because the exclusion list is the thing a build dispatcher asks about
-- and it will always be a small minority of a ~1,500-row table.
CREATE INDEX IF NOT EXISTS idx_positioning_register_excluded
    ON positioning_register (domain) WHERE exclude_from_build;

COMMENT ON TABLE positioning_register IS
'The portfolio positioning register (RFC_037, owner ruling 2026-08-19: the database is the source of truth). One row per domain. raw_md holds the entire original entry and is AUTHORITATIVE — the typed columns are a convenience index over it, because the register''s value is its reasoning and a lossy parse would be a deletion rather than a migration.';

COMMENT ON COLUMN positioning_register.raw_md IS
'The whole original entry. Authoritative: when a typed column disagrees with this, this wins.';

COMMENT ON COLUMN positioning_register.exclude_from_build IS
'Build-queue exclusion (owner 2026-08-20). NOT a reserved state — the owner explicitly declined that ceremony for the test domains.';

DO $$
DECLARE n_cols int; n_idx int;
BEGIN
    SELECT count(*) INTO n_cols FROM information_schema.columns
     WHERE table_name = 'positioning_register';
    IF n_cols < 20 THEN RAISE EXCEPTION 'positioning_register has only % columns', n_cols; END IF;

    SELECT count(*) INTO n_idx FROM pg_indexes
     WHERE tablename = 'positioning_register' AND indexname = 'idx_positioning_register_domain';
    IF n_idx <> 1 THEN
      RAISE EXCEPTION 'the unique domain index is missing — the loader would accumulate duplicates';
    END IF;

    -- Prove the uniqueness actually bites, rather than trusting the DDL.
    BEGIN
        INSERT INTO positioning_register (domain, created_by) VALUES ('zz-probe.uk', '511-probe');
        INSERT INTO positioning_register (domain, created_by) VALUES ('ZZ-Probe.uk', '511-probe');
        RAISE EXCEPTION 'the unique index did NOT reject a case-different duplicate';
    EXCEPTION WHEN unique_violation THEN
        NULL;  -- correct
    END;
    DELETE FROM positioning_register WHERE created_by = '511-probe';

    RAISE NOTICE '511 OK — positioning_register created, % columns, case-insensitive uniqueness PROVEN', n_cols;
END $$;

COMMIT;
