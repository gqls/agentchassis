-- ============================================================================
-- Phase 2A — add locking columns to assets
--
-- Mirrors page_components and site_components exactly:
--   * locked_at timestamp with time zone (nullable, no default)
--   * locked_by text                     (nullable, no default)
--   * Partial index on (locked_at) WHERE locked_at IS NOT NULL
--
-- Lock convention (canonical reference: 031_locks.md, Pattern A):
--
-- Detection — a row is locked if:
--     locked_at IS NOT NULL
-- No time comparison. There is no `lock_type` or `lock_expires_at` column
-- in production today; timed expiry is documented design intent (004 v4,
-- 007 v4) but not implemented. The lock-expiry project will add both
-- columns uniformly across all four Pattern A tables in a single future
-- migration; until then `locked_at IS NOT NULL` is the complete check.
--
-- Classification — hard vs soft, by `locked_by` value:
--   Hard (human only can clear):  'admin', 'admin-removed', 'checkpoint'
--   Soft (agents can clear):      'deploy', 'manual', auditor names, anything else
--
-- The canonical Go-side classifier is
-- platform/orchestration/actions/check_component_lock.go's
-- CheckComponentLock function — it returns a ComponentLockStatus struct
-- with an IsHard field. New consumers of asset locks should mirror that
-- helper's logic rather than reinventing classification.
--
-- Discovery agents skip BOTH hard and soft locks (filter:
-- `locked_at IS NULL`). Execution agents skip hard locks; they may clear
-- soft locks when an explicit work item references the row.
--
-- locked_by vocabulary for assets (extends 013_content_governance.md's
-- table for page_components and site_components):
--   'manual'                       human upload via dashboard — HARD
--   'admin'                        human edit via dashboard — HARD
--   'visual-design-auditor'        Phase 4 auditor approved — SOFT
--   'imagery-quality-auditor'      Phase 6 auditor (future) — SOFT
--   'audit-pending'                transient, set at audit start, agent
--                                  clears at audit completion — SOFT
--
-- These are convention not constraint — no CHECK is added so future
-- callers can introduce new identifiers without a schema change.
--
-- This migration is purely additive. Existing rows retain locked_at=NULL,
-- meaning "never locked, available for any operation."
-- ============================================================================

BEGIN;

ALTER TABLE assets
    ADD COLUMN locked_at timestamp with time zone,
  ADD COLUMN locked_by text;

CREATE INDEX idx_assets_locked
    ON assets (locked_at)
    WHERE locked_at IS NOT NULL;


-- ----------------------------------------------------------------------------
-- Verification
-- ----------------------------------------------------------------------------
SELECT
    column_name,
    data_type,
    is_nullable
FROM information_schema.columns
WHERE table_name = 'assets'
  AND column_name IN ('locked_at', 'locked_by');

SELECT indexname, indexdef
FROM pg_indexes
WHERE tablename = 'assets'
  AND indexname = 'idx_assets_locked';

-- Sanity: existing rows are unaffected (all locked_at NULL, all status='active')
SELECT
    COUNT(*) AS total,
    COUNT(*) FILTER (WHERE locked_at IS NULL) AS unlocked,
    COUNT(*) FILTER (WHERE locked_at IS NOT NULL) AS locked,
    COUNT(*) FILTER (WHERE locked_by IS NOT NULL) AS has_locked_by
FROM assets;

COMMIT;


-- ============================================================================
-- ROLLBACK (only if reverting)
-- ============================================================================
-- BEGIN;
-- DROP INDEX IF EXISTS idx_assets_locked;
-- ALTER TABLE assets
--   DROP COLUMN IF EXISTS locked_at,
--   DROP COLUMN IF EXISTS locked_by;
-- COMMIT;