-- 216_site_work_items_touch_updated_at.sql
--
-- bugs_open/035 — site_work_items.updated_at is not maintained, so a completed
-- item reads as "never picked up" (updated_at == created_at) and the documented
-- remedy for a dropped dispatch is a resubmit, i.e. a credit spend.
--
-- WHY A TRIGGER
-- -------------
-- Measured 2026-07-26: 134 of 941 complete rows had updated_at > created_at
-- (86% wrong), zero triggers on the table. Since the bug was filed (07-20)
-- several Go writers gained explicit `updated_at = NOW()` (claim_work_item,
-- complete_work_item_verification, load_work_item_actions, the admin confirm
-- handler) — the exact "several writers must all remember" drift the bug file
-- argues against. One trigger covers every writer, present and future; the
-- explicit writers coexist harmlessly (the trigger overwrites with the same
-- NOW()).
--
-- SAFETY (checked 2026-07-26, evidence in the bug file's resolution section)
-- --------------------------------------------------------------------------
-- * No reader of site_work_items.updated_at exists: not in Go, not in the
--   admin dashboard, not in scripts, not in scheduled_tasks / agent_definitions
--   config (the sweepers that match a grep read updated_at on OTHER tables,
--   and stale-work-item-reaper filters on created_at).
-- * No index involves updated_at; idx_swi_dedup keys (site_id, item_key).
-- * public.set_updated_at() already exists and backs trg_<table>_updated_at
--   triggers on site_specs, layouts, site_plans, content_feed_items and
--   model_lifecycle.training_runs — reused here, deliberately NOT re-created.
-- * No backfill: completed_at carries the historical truth; inventing an
--   updated_at for old rows would fabricate timestamps (bug file, "How to
--   verify a fix" #3).
--
-- ROLLBACK
-- --------
--   DROP TRIGGER trg_site_work_items_updated_at ON site_work_items;

BEGIN;

DROP TRIGGER IF EXISTS trg_site_work_items_updated_at ON site_work_items;
CREATE TRIGGER trg_site_work_items_updated_at
BEFORE UPDATE ON site_work_items
FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- Guard: live-fire the trigger inside this transaction, on a probe row that is
-- inserted and deleted here so it is never visible to any other session and
-- never survives the file. Deliberately NOT a no-op write on a real historical
-- row: that would stamp today's timestamp onto it, which is the same
-- fabrication the "no backfill" rule forbids.
--
-- The probe is backdated an hour because now() is FIXED for the whole
-- transaction: a row inserted and updated in one txn gets the identical
-- timestamp from both defaults and from the trigger, so an in-txn probe at
-- now() would read as "updated_at did not move" even with the trigger working
-- correctly. Backdating created_at/updated_at makes the trigger's now() the
-- only thing that can move the column.
DO $$
DECLARE
  v_site    uuid;
  v_id      uuid;
  v_created timestamptz;
  v_updated timestamptz;
BEGIN
  SELECT id INTO v_site FROM sites ORDER BY created_at ASC LIMIT 1;
  IF v_site IS NULL THEN
    RAISE EXCEPTION '216 guard: no site row available to hang a probe work item on';
  END IF;

  INSERT INTO site_work_items
    (site_id, source, item_type, summary, created_by, status, created_at, updated_at)
  VALUES
    (v_site, 'migration_216_probe', 'migration_probe',
     'probe: assert trg_site_work_items_updated_at fires (rolled back)',
     'migration_216', 'cancelled', now() - interval '1 hour', now() - interval '1 hour')
  RETURNING id, created_at INTO v_id, v_created;

  UPDATE site_work_items SET status = status WHERE id = v_id;
  SELECT updated_at INTO v_updated FROM site_work_items WHERE id = v_id;

  DELETE FROM site_work_items WHERE id = v_id;

  IF v_updated IS NULL OR v_updated <= v_created THEN
    RAISE EXCEPTION '216 guard: updated_at did not move on UPDATE (got %, created %) — trigger not firing',
      v_updated, v_created;
  END IF;
END $$;

COMMIT;
