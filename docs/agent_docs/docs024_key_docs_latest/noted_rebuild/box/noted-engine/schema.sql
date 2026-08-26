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
    kind        TEXT NOT NULL CHECK (kind IN ('audio','image','video')),
    mime        TEXT NOT NULL DEFAULT 'application/octet-stream',
    bytes       BYTEA NOT NULL,
    byte_len    BIGINT NOT NULL,
    ordering    INT NOT NULL DEFAULT 0,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_media_note ON media(note_id, kind, ordering);

-- 2026-08-24: 'video' joins the kinds (owner ask — the media pasteboard, stage 1).
-- The CHECK above only applies to a freshly created table; on an existing
-- database CREATE TABLE IF NOT EXISTS is a no-op, so this pair is the real
-- migration. Idempotent like everything else here, re-run at every startup.
ALTER TABLE media DROP CONSTRAINT IF EXISTS media_kind_check;
ALTER TABLE media ADD CONSTRAINT media_kind_check CHECK (kind IN ('audio','image','video'));

-- 2026-08-25 (OWNER RULING — reverses this file's 08-10 header): media bytes
-- move to Backblaze B2 so a paid storage increase is a bucket setting, not a
-- disk. `storage_key`/`b2_file_id` set = the bytes live in B2 and `bytes` is
-- NULL; rows with storage_key NULL keep serving from `bytes` (no drain needed).
-- The quota stays exactly as it is — byte_len still counts, the trigger still
-- maintains accounts.media_bytes — it is now the abuse valve and the future
-- paid tier's lever rather than a disk-protection measure.
ALTER TABLE media ADD COLUMN IF NOT EXISTS storage_key TEXT;
ALTER TABLE media ADD COLUMN IF NOT EXISTS b2_file_id TEXT;
ALTER TABLE media ALTER COLUMN bytes DROP NOT NULL;

-- 2026-08-25, pasteboard stage 3: a caption belongs to the media ROW (not the
-- board layout), so it survives re-arrangement and feeds any future editor
-- view of the same item.
ALTER TABLE media ADD COLUMN IF NOT EXISTS caption TEXT NOT NULL DEFAULT '';

-- 2026-08-25, pasteboard stage 2: per-note board arrangement, client-owned,
-- versioned shape {v:1, items:[...]} with coordinates as fractions of board
-- width. Absent = the note has never been arranged (linear view).
ALTER TABLE notes ADD COLUMN IF NOT EXISTS layout JSONB;

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

-- 2026-08-26, large uploads (PLAN_2026-08-26_large_uploads.md — the 25 MB
-- blocker gating the paid tier). Two parts.
--
-- Per-account limit overrides: NULL = the process-wide env default applies, so
-- every existing account behaves exactly as before this ALTER. Setting them is
-- the paid tier's whole lever — 1 TB is an UPDATE, not a deploy.
ALTER TABLE accounts ADD COLUMN IF NOT EXISTS media_quota_override_bytes BIGINT;
ALTER TABLE accounts ADD COLUMN IF NOT EXISTS max_upload_override_bytes  BIGINT;

-- A pending upload RESERVES quota from `begin`: the quota checks count
-- declared_bytes of open reservations beside accounts.media_bytes, so two
-- concurrent begins cannot promise the same headroom twice. `finish` converts
-- the reservation into a media row in one transaction; abort or the reaper
-- releases it. parts records what B2 has confirmed: {"1": {"size":..,
-- "sha1":".."}, ...} — JSONB for the same ship-with-the-binary reason as
-- notes.layout.
CREATE TABLE IF NOT EXISTS pending_uploads (
    id             BIGSERIAL PRIMARY KEY,
    account_id     BIGINT NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    note_id        BIGINT NOT NULL REFERENCES notes(id) ON DELETE CASCADE,
    kind           TEXT NOT NULL CHECK (kind IN ('audio','image','video')),
    mime           TEXT NOT NULL DEFAULT 'application/octet-stream',
    declared_bytes BIGINT NOT NULL CHECK (declared_bytes > 0),
    -- Chosen at begin (B2 needs >=2 parts of >=5MB except the last, Cloudflare
    -- caps a request under ~100MB, so the size is per-upload arithmetic).
    -- PERSISTED rather than recomputed so an upload started under one binary
    -- validates identically under the next.
    part_size_bytes BIGINT NOT NULL CHECK (part_size_bytes > 0),
    storage_key    TEXT NOT NULL,
    b2_file_id     TEXT NOT NULL,
    parts          JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_pending_uploads_account ON pending_uploads(account_id);
CREATE INDEX IF NOT EXISTS idx_pending_uploads_age     ON pending_uploads(created_at);
