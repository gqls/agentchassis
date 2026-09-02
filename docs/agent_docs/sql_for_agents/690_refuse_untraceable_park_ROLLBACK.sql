-- 690_refuse_untraceable_park_ROLLBACK.sql
--
-- Withdraws migration 690's guard. bugs_open/396.
--
-- ⚠ IF YOU ARE IN A HURRY, THE TRIGGER ALONE IS THE SWITCH AND IT IS ONE STATEMENT:
--
--     DROP TRIGGER trg_site_work_items_park_provenance ON site_work_items;
--
-- That takes effect immediately, needs no image roll, and leaves the function in place so
-- re-arming is a single CREATE TRIGGER. Run this whole file only when you want the function
-- gone as well.
--
-- ⚠ WHAT YOU ARE TURNING BACK ON, stated so the decision is informed: without the trigger,
-- any raw `UPDATE site_work_items SET status='deferred'` on a row with a named handler_agent
-- succeeds silently. Such a row is selected by NOTHING (claim takes triaged/approved, the
-- promoter takes detected) and still holds its idx_swi_dedup slot, so the detector cannot
-- re-file it and another session dispatching that page hits 23505 — a failure that reads as
-- "already queued" and means "queued and abandoned". 170 such unattributable rows existed on
-- 2026-09-02; one of that shape blocked bugs_open/328 for 22 days.
--
-- This file is idempotent and safe to run when the objects are already absent.

BEGIN;

DO $rb$
DECLARE
    v_had_trigger boolean;
    v_had_func    boolean;
BEGIN
    v_had_trigger := EXISTS (SELECT 1 FROM pg_trigger
                              WHERE tgrelid = 'site_work_items'::regclass
                                AND tgname  = 'trg_site_work_items_park_provenance'
                                AND NOT tgisinternal);
    v_had_func := EXISTS (SELECT 1 FROM pg_proc WHERE proname = 'refuse_untraceable_park');

    RAISE NOTICE '690 ROLLBACK: trigger present before = %, function present before = %',
                 v_had_trigger, v_had_func;
END
$rb$;

DROP TRIGGER IF EXISTS trg_site_work_items_park_provenance ON site_work_items;
DROP FUNCTION IF EXISTS refuse_untraceable_park();

-- POST-CHECK. A SELECT here could not stop the COMMIT, so this is DO/RAISE (LANDMINES).
DO $rbpost$
BEGIN
    IF EXISTS (SELECT 1 FROM pg_trigger
                WHERE tgrelid = 'site_work_items'::regclass
                  AND tgname  = 'trg_site_work_items_park_provenance'
                  AND NOT tgisinternal) THEN
        RAISE EXCEPTION '690 ROLLBACK FAILED: the trigger is still attached';
    END IF;
    IF EXISTS (SELECT 1 FROM pg_proc WHERE proname = 'refuse_untraceable_park') THEN
        RAISE EXCEPTION '690 ROLLBACK FAILED: refuse_untraceable_park() still exists';
    END IF;
    RAISE NOTICE '690 ROLLBACK COMPLETE: guard removed. Untraceable parks are possible again.';
END
$rbpost$;

COMMIT;
