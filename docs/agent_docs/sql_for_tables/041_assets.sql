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

---

-- ============================================================================
-- Phase 2B — add asset_key column to assets (preparation for multi-image)
--
-- Goal: introduce a per-row asset key that distinguishes multiple images
-- with the same logical `purpose` (e.g. a Phase 3 adoption-mirror import
-- that produces five rows all with `purpose='adopted_image'`, differentiated
-- by `asset_key='adopted:<filename>'`).
--
-- Phase 2B is purely additive. After this migration:
--   * Every existing row has asset_key = purpose (backfilled).
--   * Future inserts that don't specify asset_key default to purpose
--     (Phase 2C makes the Go side honour this; Phase 2B alone leaves the
--     column nullable and population to the caller).
--   * Two unique constraints coexist:
--       OLD: idx_assets_site_purpose_unique  (site_id, purpose) WHERE purpose IS NOT NULL
--       NEW: idx_assets_site_asset_key_unique (site_id, asset_key) WHERE asset_key IS NOT NULL AND status = 'active'
--     They overlap harmlessly while asset_key = purpose for everything.
--
-- Phase 2C will update store_asset to write asset_key and switch ON CONFLICT
-- target. Phase 2D drops the old purpose-uniqueness constraint after Phase 2C
-- has been live and stable. Until 2D, the old constraint still blocks any
-- multi-image case where two rows share a purpose; that's expected and is
-- exactly what the staged sequence is designed for.
--
-- Two statements:
--   1. BEGIN/COMMIT block: ALTER TABLE + UPDATE backfill (atomic).
--   2. CREATE INDEX CONCURRENTLY (outside transaction; non-blocking on the
--      live table).
-- ============================================================================

BEGIN;

ALTER TABLE assets
    ADD COLUMN asset_key text;

-- Backfill from purpose for all existing rows
UPDATE assets
SET asset_key = purpose
WHERE purpose IS NOT NULL
  AND asset_key IS NULL;


-- ----------------------------------------------------------------------------
-- Verification (in same transaction so we abort if state is wrong)
-- ----------------------------------------------------------------------------
-- Every row that has a purpose should now have an asset_key matching it.
SELECT
    COUNT(*)                                                          AS total,
    COUNT(*) FILTER (WHERE purpose IS NOT NULL)                        AS has_purpose,
    COUNT(*) FILTER (WHERE asset_key IS NOT NULL)                      AS has_asset_key,
    COUNT(*) FILTER (WHERE purpose IS NOT NULL AND asset_key = purpose) AS asset_key_matches_purpose
FROM assets;

-- Sanity: there should be no rows where purpose is set but asset_key isn't
-- (otherwise the backfill missed something).
SELECT COUNT(*) AS backfill_gaps
FROM assets
WHERE purpose IS NOT NULL
  AND asset_key IS NULL;

COMMIT;


-- ============================================================================
-- Index creation (must be OUTSIDE transaction — CONCURRENTLY rule)
-- ============================================================================
-- This may take time on large tables; on this one (small, early-stage)
-- it'll be near-instant. Using CONCURRENTLY anyway for hygiene.

CREATE UNIQUE INDEX CONCURRENTLY IF NOT EXISTS idx_assets_site_asset_key_unique
    ON assets(site_id, asset_key)
    WHERE asset_key IS NOT NULL AND status = 'active';


-- ----------------------------------------------------------------------------
-- Post-creation verification (informational; can be commented out)
-- ----------------------------------------------------------------------------
SELECT indexname, indexdef
FROM pg_indexes
WHERE tablename = 'assets'
  AND indexname IN ('idx_assets_site_purpose_unique', 'idx_assets_site_asset_key_unique')
ORDER BY indexname;


-- ============================================================================
-- ROLLBACK (only if reverting)
-- ============================================================================
-- BEGIN;
-- DROP INDEX IF EXISTS idx_assets_site_asset_key_unique;
-- ALTER TABLE assets DROP COLUMN IF EXISTS asset_key;
-- COMMIT;


-----

clients_db=# -- ============================================================================
-- Phase 2D — pre-migration backup before dropping idx_assets_site_purpose_unique
--
-- DROP INDEX is reversible by re-creating the index (same definition is
-- preserved in this file's docstring), so this backup is light insurance.
-- It captures the live row count and the index definition just in case.
-- ============================================================================

BEGIN;

CREATE TABLE assets_backup_20260508_pre_phase2d_dropconstraint AS
SELECT * FROM assets;

-- Verification — backup row count should equal live row count
SELECT 'assets' AS table_name,
       (SELECT COUNT(*) FROM assets)                                              AS live,
       (SELECT COUNT(*) FROM assets_backup_20260508_pre_phase2d_dropconstraint)   AS backup;

-- Record the old index definition so we can recreate it for rollback
SELECT indexname, indexdef
FROM pg_indexes
WHERE tablename = 'assets' AND indexname = 'idx_assets_site_purpose_unique';

COMMIT;


-- ============================================================================
-- RESTORE — only if reverting Phase 2D
-- ============================================================================
-- The original index definition (recorded above for reference):
--
--   CREATE UNIQUE INDEX idx_assets_site_purpose_unique
--     ON public.assets USING btree (site_id, purpose)
-- the recreate will fail. Resolve those before re-creating.et_key),eurpose_unique
BEGIN
SELECT 11
           table_name | live | backup
------------+------+--------
 assets     |   11 |     11
(1 row)

           indexname            |                                                            indexdef
--------------------------------+--------------------------------------------------------------------------------------------------------------------------------
 idx_assets_site_purpose_unique | CREATE UNIQUE INDEX idx_assets_site_purpose_unique ON public.assets USING btree (site_id, purpose) WHERE (purpose IS NOT NULL)
    (1 row)

    COMMIT
    clients_db=# -- ============================================================================
-- Phase 2D — drop the old (site_id, purpose) unique constraint
--
-- Removes idx_assets_site_purpose_unique. After this, the only uniqueness
-- protection on assets is idx_assets_site_asset_key_unique (created in
-- Phase 2B). The multi-image case becomes possible at the schema level.
--
-- PRECONDITIONS (verify before running):
--   1. Phase 2B has been applied (idx_assets_site_asset_key_unique exists,
--      asset_key column populated for all existing rows).
--   2. Phase 2C has been deployed (StoreAssetAction writes asset_key on
--      every insert; ON CONFLICT target switched to (site_id, asset_key)).
--   3. No active row has purpose set without asset_key — i.e. no straggler
--      rows from the gap between Phase 2B migration and Phase 2C deploy.
--      The pre-drop check below verifies this.
--
-- This migration has two statements:
--   1. BEGIN/COMMIT block: pre-drop sanity check that fails the
--      transaction if any active row has purpose without asset_key.
--   2. DROP INDEX CONCURRENTLY IF EXISTS (must be outside transaction).
--
-- IMPORTANT: \set ON_ERROR_STOP on at the top is essential. Without it,
-- psql would proceed to the DROP INDEX statement even when the sanity
-- check transaction aborted — which would defeat the safety guard.
-- ============================================================================

    \set ON_ERROR_STOP on

-- ----------------------------------------------------------------------------
-- Pre-drop sanity check (in transaction so we abort on bad state)
-- ----------------------------------------------------------------------------
BEGIN;

--   ON assets(site_id, purpose) WHERE purpose IS NOT NULL;s_site_purpose_uniquet.';
unrecognized value "on--------------------------------------------------------------------------------Pre-dropsanitycheck(intransactionsoweabortonbadstate)------------------------------------------------------------------------------BEGIN;--Lookforactiverowsthathavepurposesetbutasset_keynotset.--Thesewouldbeinvisibletothenewindex(WHEREasset_keyISNOTNULL),--sodroppingtheoldindexwouldleavethemwithnouniquenessprotection.DO$$DECLAREstraggler_countinteger;BEGINSELECTCOUNT(*)INTOstraggler_countFROMassetsWHEREstatus=activeANDpurposeISNOTNULLANDasset_keyISNULL;IFstraggler_count>0THENRAISEEXCEPTIONPhase 2D abort: % active row(s) have purpose set without asset_key. These would lose uniqueness protection if the old index is dropped. Likely cause: rows inserted between Phase 2B migration and Phase 2C deploy by the unpatched StoreAssetAction. To resolve, backfill: UPDATE assets SET asset_key = purpose WHERE status = 'active' AND purpose IS NOT NULL AND asset_key IS NULL;  Then re-run this migration.,straggler_count;ENDIF;RAISENOTICEPre-drop check passed: 0 straggler rows.;END$$;--Alsoconfirmthenewindexexists.Ifitdoesnt, Phase 2B wasntapplied.DO$$BEGINIFNOTEXISTS(SELECT1FROMpg_indexesWHEREtablename=assetsANDindexname=idx_assets_site_asset_key_unique)THENRAISEEXCEPTIONPhase 2D abort: idx_assets_site_asset_key_unique does not exist. Phase 2B must be applied before Phase 2D.;ENDIF;RAISENOTICEPre-drop check passed: idx_assets_site_asset_key_unique present.;END$$;COMMIT;--------------------------------------------------------------------------------Droptheoldconstraint(mustbeOUTSIDEtransactionforCONCURRENTLY)------------------------------------------------------------------------------DROPINDEXCONCURRENTLYIFEXISTSidx_assets_site_purpose_unique;--------------------------------------------------------------------------------Post-dropverification------------------------------------------------------------------------------SELECTindexname,indexdefFROMpg_indexesWHEREtablename=assetsANDindexnameIN(idx_assets_site_purpose_unique,idx_assets_site_asset_key_unique)ORDERBYindexname;--Expected:onlyidx_assets_site_asset_key_uniquepresent.--============================================================================--ROLLBACK(recreatetheoldconstraint)--============================================================================--ThiswillFAILifanyrowsnowviolate(site_id,purpose)uniqueness—--whichistheverystatePhase2Denables.Resolveviolationsfirst.----CREATEUNIQUEINDEXCONCURRENTLYIFNOTEXISTSidx_assets_site_purpose_unique--ONassets(site_id,purpose)WHEREpurposeISNOTNULL;" for "ON_ERROR_STOP": Boolean expected
clients_db=#

             