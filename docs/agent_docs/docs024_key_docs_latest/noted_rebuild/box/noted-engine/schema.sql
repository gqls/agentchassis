-- noted.co.uk — engine schema.
-- Applied by the engine at startup (idempotent), so the schema and the code
-- that reads it ship together and cannot drift.
--
-- MEDIA LIVES IN POSTGRES, WHICH REVISES PLAN_2026-08-10_noted_rebuild.md §3a-i.
-- That said media would go to B2 so noted's unbounded storage growth could never
-- fill the box's 50 GB disk and take the webdesign.uk shopfront down with it.
-- The concern was right; the remedy here is different and, for launch, better:
--
--   * the coupling is closed by a HARD QUOTA (per account, enforced in the same
--     transaction as the insert), not by moving the bytes elsewhere. A quota
--     bounds disk growth; a different storage backend only relocates it.
--   * media in Postgres is media inside the nightly encrypted dump. Media in B2
--     needs its own backup story, its own credential on the box, and a restore
--     that has to reunite two stores at the same point in time. At launch
--     volumes that is more ways to lose someone's recordings, not fewer.
--
-- Revisit when total media passes a few GB: the migration path is a
-- `storage_key` column beside `bytes`, filled for new rows, with `bytes` drained
-- in the background. Written now so the door is left open deliberately.

CREATE TABLE IF NOT EXISTS accounts (
    id              BIGSERIAL PRIMARY KEY,
    email           TEXT NOT NULL,
    -- Case- and whitespace-insensitive identity. Stored separately rather than
    -- lower()ing on read: a functional unique index would still let two rows
    -- differing only in case be created by two concurrent inserts on some
    -- collations, and this way the uniqueness is on the exact bytes compared.
    email_canonical TEXT NOT NULL UNIQUE,
    password_hash   TEXT NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    last_login_at   TIMESTAMPTZ,
    -- Bytes of media currently stored. Maintained transactionally with the
    -- media rows so the quota check cannot race an insert.
    media_bytes     BIGINT NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS sessions (
    token_hash  TEXT PRIMARY KEY,          -- sha256 of the cookie value, never the value
    account_id  BIGINT NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    expires_at  TIMESTAMPTZ NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_sessions_account ON sessions(account_id);
CREATE INDEX IF NOT EXISTS idx_sessions_expiry  ON sessions(expires_at);

CREATE TABLE IF NOT EXISTS notes (
    id          BIGSERIAL PRIMARY KEY,
    account_id  BIGINT NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    -- The id the browser app used. Carried so an import is idempotent: importing
    -- the same backup twice updates rather than duplicating. Unique per account,
    -- NOT globally — two people's backups can legitimately share an id.
    client_id   TEXT,
    title       TEXT NOT NULL DEFAULT '',
    content     TEXT NOT NULL DEFAULT '',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at  TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS idx_notes_account ON notes(account_id, updated_at DESC);
CREATE UNIQUE INDEX IF NOT EXISTS idx_notes_client_id
    ON notes(account_id, client_id) WHERE client_id IS NOT NULL;

CREATE TABLE IF NOT EXISTS media (
    id          BIGSERIAL PRIMARY KEY,
    note_id     BIGINT NOT NULL REFERENCES notes(id) ON DELETE CASCADE,
    account_id  BIGINT NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    kind        TEXT NOT NULL CHECK (kind IN ('audio','image')),
    mime        TEXT NOT NULL DEFAULT 'application/octet-stream',
    bytes       BYTEA NOT NULL,
    byte_len    BIGINT NOT NULL,
    ordering    INT NOT NULL DEFAULT 0,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_media_note ON media(note_id, kind, ordering);

-- Keep accounts.media_bytes true no matter which code path writes media. A
-- quota enforced only in the handler is a quota that the next handler forgets.
CREATE OR REPLACE FUNCTION media_bytes_maintain() RETURNS TRIGGER AS $$
BEGIN
    IF TG_OP = 'INSERT' THEN
        UPDATE accounts SET media_bytes = media_bytes + NEW.byte_len WHERE id = NEW.account_id;
    ELSIF TG_OP = 'DELETE' THEN
        UPDATE accounts SET media_bytes = GREATEST(0, media_bytes - OLD.byte_len) WHERE id = OLD.account_id;
    END IF;
    RETURN NULL;
END $$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trg_media_bytes ON media;
CREATE TRIGGER trg_media_bytes
    AFTER INSERT OR DELETE ON media
    FOR EACH ROW EXECUTE FUNCTION media_bytes_maintain();
