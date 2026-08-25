-- 632_site_lock_exception_list.sql
--
-- `sites.lock_except_item_ids` — the column half of "hold this site's queue
-- EXCEPT these items". bugs_open/396, after council `ed821065` returned REVISE.
--
-- ── WHY, AND WHY IT IS NOT A PARK VERB ──
--
-- This lane first proposed a park verb (migration 621, WII-034) on the premise
-- that "the platform offers NO way to hold a site's work queue". The council's
-- prior_art seat refuted that premise and was right: `sites.locked_at` /
-- `sites.locked_by` already exist, are live on 3 of 51 sites, are gated by
-- `build-pipeline-trigger`'s `find_dispatchable_site`, and have admin lock/unlock
-- endpoints. Crucially the lock MUTATES NO WORK-ITEM ROW — so it strands nothing,
-- holds no `idx_swi_dedup` slot, and cannot produce the 23505 that stalled a page
-- for 22 days in bugs_open/328.
--
-- What is actually missing is narrower: the lock is ALL-OR-NOTHING. `sites` has
-- only `locked_at` and `locked_by`. That is precisely why the
-- mortgagecalculator_couk_adoption lane wrote a 15-second auto-defer backstop —
-- its own handoff calls the site lock "(a)" and item status "(b) the finer
-- control" — and that backstop is what made 38 unattributable parked rows.
--
-- So: give the lock the exception list it lacks, and the reason to reach for
-- `deferred` goes away. No row changes status, so nothing can be stranded.
--
-- ── THIS FILE IS THE SAFE HALF, AND IT IS DELIBERATELY SEPARATE ──
--
-- Adding a nullable column changes NO behaviour: nothing reads it until the
-- chassis binary carrying `honour_site_lock` has rolled AND the held config half
-- (633_..._HOLD.sql) is applied. Applying this file today is inert by
-- construction, which is why it is not itself a _HOLD file.
--
-- ⚠ THE ORDERING TRAP THAT SPLIT THESE FILES — read before touching 633.
-- `find_dispatchable_site` selects a SITE, not an item. If its pre_query is
-- taught to select a locked site because one excepted item is dispatchable, the
-- next step (`build-dispatch-loop > load_items` → `LoadWorkItemsAction`) loads
-- EVERY dispatchable item on that site, because that loader has never checked
-- the lock. The exception list would unlock the whole queue — the exact failure
-- the lock exists to prevent. The binary must ship the loader's
-- `honour_site_lock` arm FIRST. That is why 633 is a `_HOLD` file and this one
-- is not.
--
-- ROLLBACK: 632_site_lock_exception_list_ROLLBACK.sql

BEGIN;

ALTER TABLE sites
  ADD COLUMN IF NOT EXISTS lock_except_item_ids uuid[];

COMMENT ON COLUMN sites.lock_except_item_ids IS
  'bugs_open/396. Work-item ids that may still dispatch while sites.locked_at is set — "hold this site EXCEPT these". NULL/empty = the full hold, which is what a lock has always meant. Honoured by build-pipeline-trigger>find_dispatchable_site and, behind the opt-in step-config key honour_site_lock, by LoadWorkItemsAction. Both halves are required: the site gate alone selects the SITE, and the loader would then take every dispatchable item on it.';

-- ---------------------------------------------------------------------------
-- GUARDS. DO/RAISE, not bare SELECTs: ON_ERROR_STOP ignores a non-empty result,
-- so a verify block of SELECTs cannot stop this COMMIT.
-- ---------------------------------------------------------------------------
DO $guard$
DECLARE
    v_type   text;
    v_locked bigint;
    v_set    bigint;
BEGIN
    SELECT data_type INTO v_type FROM information_schema.columns
     WHERE table_name='sites' AND column_name='lock_except_item_ids';
    IF v_type IS NULL THEN
        RAISE EXCEPTION '632: sites.lock_except_item_ids was not created';
    END IF;
    IF v_type <> 'ARRAY' THEN
        RAISE EXCEPTION '632: sites.lock_except_item_ids is %, expected an array', v_type;
    END IF;

    -- NEGATIVE CONTROL: this file must change no behaviour. Every existing row
    -- must have a NULL exception list, or something else wrote it and the
    -- "inert by construction" claim above is false.
    SELECT count(*) INTO v_set FROM sites WHERE lock_except_item_ids IS NOT NULL;
    IF v_set <> 0 THEN
        RAISE EXCEPTION '632: % site row(s) already carry a non-NULL lock_except_item_ids — this file claims to be inert and is not', v_set;
    END IF;

    -- POSITIVE CONTROL: the lock this column modifies must actually be in use,
    -- or the premise of the whole change is stale. Without this, the negative
    -- control above passes just as well on a fleet where nothing is ever locked.
    SELECT count(*) INTO v_locked FROM sites WHERE locked_at IS NOT NULL;
    IF v_locked = 0 THEN
        RAISE EXCEPTION '632: no site is currently locked — the mechanism this column extends is not in use, so RE-MEASURE before applying (bugs_open/396 measured 3 of 51 locked on 2026-08-25)';
    END IF;

    RAISE NOTICE '632 OK: column added (array), 0 rows carry a value, % site(s) currently locked.', v_locked;
END;
$guard$;

COMMIT;
