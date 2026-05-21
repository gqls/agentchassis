-- 053_lock_expiry.sql
-- Timed lock expiry — the project doc 004 v4 designed and docs 031_locks /
-- 031_LOCKS_should_locks_expire approved (2026-05-08). Implemented now
-- (Option A) because adoption-faithfulness is a concrete driver, not just
-- anticipatory.
--
-- Adds lock_type + lock_expires_at to all four Pattern A lock-bearing tables
-- in one transaction, so no consumer ever has to know which tables have the
-- columns and which don't (the coherence requirement docs 031 insisted on).
--
-- Approved policy (with the new `adoption` source added):
--   locked_by                          lock_type   expiry
--   admin / admin-removed / checkpoint  permanent   —
--   manual                              permanent   —
--   deploy                              timed       +30 days
--   visual-design-auditor / imagery-quality-auditor  timed  +90 days
--   adoption  (NEW — faithful first pass)            timed  +90 days
--   audit-pending                       (not a lock; cleared on completion)
--
-- The expanded "is this row unlocked / improvable?" predicate becomes:
--   (locked_at IS NULL
--    OR (lock_type = 'timed' AND lock_expires_at IS NOT NULL AND lock_expires_at < NOW()))
-- Human/permanent locks (lock_expires_at NULL) are never improvable by agents.
--
-- This migration is SCHEMA + BACKFILL only. The Go follow-on (filter sweep on
-- the 11 `locked_at IS NULL` callsites, CheckComponentLock extension, the
-- adoption directive writer, the convergence layer's adoption_locked check,
-- and the expired_review_locks discovery check) is tracked in
-- FOCUS_adoption_faithfulness_via_locks.md and lands as separate code changes.

BEGIN;

-- ---------------------------------------------------------------------------
-- 1. Columns on all four Pattern A tables
-- ---------------------------------------------------------------------------
ALTER TABLE page_components
    ADD COLUMN IF NOT EXISTS lock_type       text,
    ADD COLUMN IF NOT EXISTS lock_expires_at timestamptz;

ALTER TABLE site_components
    ADD COLUMN IF NOT EXISTS lock_type       text,
    ADD COLUMN IF NOT EXISTS lock_expires_at timestamptz;

ALTER TABLE site_plan_directives
    ADD COLUMN IF NOT EXISTS lock_type       text,
    ADD COLUMN IF NOT EXISTS lock_expires_at timestamptz;

ALTER TABLE assets
    ADD COLUMN IF NOT EXISTS lock_type       text,
    ADD COLUMN IF NOT EXISTS lock_expires_at timestamptz;

-- ---------------------------------------------------------------------------
-- 2. CHECK constraint on the vocabulary (named, idempotent)
-- ---------------------------------------------------------------------------
DO $$
DECLARE t text;
BEGIN
    FOREACH t IN ARRAY ARRAY['page_components','site_components','site_plan_directives','assets']
    LOOP
        IF NOT EXISTS (
            SELECT 1 FROM pg_constraint
            WHERE conname = 'chk_' || t || '_lock_type'
        ) THEN
            EXECUTE format(
                'ALTER TABLE %I ADD CONSTRAINT %I CHECK (lock_type IS NULL OR lock_type IN (''permanent'',''timed'',''review''))',
                t, 'chk_' || t || '_lock_type'
            );
END IF;
END LOOP;
END $$;

-- ---------------------------------------------------------------------------
-- 3. Backfill existing locked rows per policy.
--
-- CONSERVATIVE CHOICE: existing timed locks get lock_expires_at measured from
-- NOW(), not from locked_at. Existing rows predate this policy and were
-- effectively permanent; measuring from NOW() gives a grace window after
-- migration so nothing releases on day one. New locks created by the writers
-- (Go follow-on) will measure from locked_at going forward.
--
-- Unlocked rows (locked_at IS NULL) are left with lock_type NULL — they are
-- simply unlocked, no expiry concept applies.
-- ---------------------------------------------------------------------------
DO $$
DECLARE t text;
BEGIN
    FOREACH t IN ARRAY ARRAY['page_components','site_components','site_plan_directives','assets']
    LOOP
        -- Human / permanent locks
        EXECUTE format($q$
            UPDATE %I SET lock_type = 'permanent'
            WHERE locked_at IS NOT NULL AND lock_type IS NULL
              AND locked_by IN ('admin','admin-removed','checkpoint','manual')
        $q$, t);

        -- deploy auto-locks -> timed 30d (grace from NOW)
EXECUTE format($q$
                   UPDATE %I SET lock_type = 'timed', lock_expires_at = NOW() + interval '30 days'
                   WHERE locked_at IS NOT NULL AND lock_type IS NULL
                   AND locked_by = 'deploy'
                   $q$, t);

-- auditor approvals -> timed 90d (grace from NOW)
EXECUTE format($q$
                   UPDATE %I SET lock_type = 'timed', lock_expires_at = NOW() + interval '90 days'
                   WHERE locked_at IS NOT NULL AND lock_type IS NULL
                   AND locked_by IN ('visual-design-auditor','imagery-quality-auditor')
                   $q$, t);

-- Any other existing locked row of unknown source -> treat as permanent
-- (conservative: don't silently expire something we can't classify).
EXECUTE format($q$
                   UPDATE %I SET lock_type = 'permanent'
            WHERE locked_at IS NOT NULL AND lock_type IS NULL
        $q$, t);
END LOOP;
END $$;

-- ---------------------------------------------------------------------------
-- 4. Partial index to make expired-timed-lock sweeps cheap
-- ---------------------------------------------------------------------------
CREATE INDEX IF NOT EXISTS idx_page_components_timed_lock
    ON page_components (lock_expires_at)
    WHERE lock_type = 'timed';
CREATE INDEX IF NOT EXISTS idx_site_components_timed_lock
    ON site_components (lock_expires_at)
    WHERE lock_type = 'timed';
CREATE INDEX IF NOT EXISTS idx_site_plan_directives_timed_lock
    ON site_plan_directives (lock_expires_at)
    WHERE lock_type = 'timed';
CREATE INDEX IF NOT EXISTS idx_assets_timed_lock
    ON assets (lock_expires_at)
    WHERE lock_type = 'timed';

COMMIT;

-- ---------------------------------------------------------------------------
-- Verification (run after COMMIT)
-- ---------------------------------------------------------------------------
-- Columns present on all four
--   SELECT table_name, column_name FROM information_schema.columns
--   WHERE column_name IN ('lock_type','lock_expires_at')
--     AND table_name IN ('page_components','site_components','site_plan_directives','assets')
--   ORDER BY table_name, column_name;
--   -- expect 8 rows (2 per table)
--
-- Backfill landed (no locked row left without a lock_type)
--   SELECT 'page_components' AS t, COUNT(*) AS locked_without_type
--   FROM page_components WHERE locked_at IS NOT NULL AND lock_type IS NULL
--   UNION ALL SELECT 'site_components', COUNT(*) FROM site_components WHERE locked_at IS NOT NULL AND lock_type IS NULL
--   UNION ALL SELECT 'site_plan_directives', COUNT(*) FROM site_plan_directives WHERE locked_at IS NOT NULL AND lock_type IS NULL
--   UNION ALL SELECT 'assets', COUNT(*) FROM assets WHERE locked_at IS NOT NULL AND lock_type IS NULL;
--   -- expect all zero
--
-- Distribution check
--   SELECT lock_type, COUNT(*) FROM page_components GROUP BY lock_type ORDER BY 1;