-- cleanup_robot_hands_icon_artifact.sql
--
-- Removes the botched icon_cross_technology artifact from robot-hands.com.
-- Per TODO item 25, option 2: delete asset row, expire work item, let discovery
-- re-emit per-actuation-type icons after the planner runs with the new prompt.
--
-- Site: robot-hands.com (00ff3af5-dad8-4770-9f70-3edc267a3c92)
-- Work item: 3cacc0dd-4bb0-44a1-a3ca-87a3911423a8 (status=triaged with stale claim metadata)
-- Asset: created 2026-05-15 15:55:36, the 6-panel SDXL output that isn't an icon
--
-- Run AFTER phase_2g_planner_imagery_prompt_decomposition.sql succeeds.
-- Run BEFORE re-triggering build-site-planner — otherwise the new plan won't
-- see this slot as unfulfilled.
--
-- Includes git cleanup hint at the bottom — must be run separately by the operator.

BEGIN;

-- ----------------------------------------------------------------------------
-- Sanity check: confirm we're about to delete what we expect
-- ----------------------------------------------------------------------------
DO $$
DECLARE
  asset_count int;
  work_item_count int;
  plan_imagery_count int;
BEGIN
  SELECT count(*) INTO asset_count
  FROM assets
  WHERE site_id = '00ff3af5-dad8-4770-9f70-3edc267a3c92'
    AND asset_key = 'icon_cross_technology';

  SELECT count(*) INTO work_item_count
  FROM site_work_items
  WHERE id = '3cacc0dd-4bb0-44a1-a3ca-87a3911423a8';

  SELECT count(*) INTO plan_imagery_count
  FROM site_plan_imagery spi
  JOIN site_plans sp ON sp.id = spi.plan_id
  WHERE sp.site_id = '00ff3af5-dad8-4770-9f70-3edc267a3c92'
    AND spi.key = 'icon_cross_technology';

  RAISE NOTICE 'About to delete: % asset(s), update % work item(s), delete % plan imagery entry/entries',
    asset_count, work_item_count, plan_imagery_count;

  IF asset_count = 0 AND work_item_count = 0 AND plan_imagery_count = 0 THEN
    RAISE NOTICE 'Nothing to clean up — likely already run.';
  END IF;
END $$;

-- ----------------------------------------------------------------------------
-- Backup
-- ----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS migration_backups (
  id serial PRIMARY KEY,
  migration_name text NOT NULL,
  applied_at timestamptz NOT NULL DEFAULT now(),
  target_table text,
  target_id text,
  old_value jsonb,
  notes text
);

INSERT INTO migration_backups (migration_name, target_table, target_id, old_value, notes)
SELECT
  'cleanup_robot_hands_icon_artifact',
  'assets',
  id::text,
  to_jsonb(a.*),
  'asset row pre-deletion'
FROM assets a
WHERE site_id = '00ff3af5-dad8-4770-9f70-3edc267a3c92'
  AND asset_key = 'icon_cross_technology';

INSERT INTO migration_backups (migration_name, target_table, target_id, old_value, notes)
SELECT
  'cleanup_robot_hands_icon_artifact',
  'site_work_items',
  id::text,
  to_jsonb(wi.*),
  'work item pre-update'
FROM site_work_items wi
WHERE id = '3cacc0dd-4bb0-44a1-a3ca-87a3911423a8';

INSERT INTO migration_backups (migration_name, target_table, target_id, old_value, notes)
SELECT
  'cleanup_robot_hands_icon_artifact',
  'site_plan_imagery',
  spi.id::text,
  to_jsonb(spi.*),
  'plan imagery entry pre-deletion'
FROM site_plan_imagery spi
JOIN site_plans sp ON sp.id = spi.plan_id
WHERE sp.site_id = '00ff3af5-dad8-4770-9f70-3edc267a3c92'
  AND spi.key = 'icon_cross_technology';

-- ----------------------------------------------------------------------------
-- 1. Delete the asset row (the 6-panel image that shouldn't exist)
-- ----------------------------------------------------------------------------
DELETE FROM assets
WHERE site_id = '00ff3af5-dad8-4770-9f70-3edc267a3c92'
  AND asset_key = 'icon_cross_technology';

-- ----------------------------------------------------------------------------
-- 2. Expire the work item so it won't claim again
-- ----------------------------------------------------------------------------
UPDATE site_work_items
SET status = 'wont_fix',
    claimed_at = NULL,
    claimed_by = NULL,
    error = 'Superseded by item 25 — replaced with per-actuation-type icons after planner prompt update'
WHERE id = '3cacc0dd-4bb0-44a1-a3ca-87a3911423a8';

-- ----------------------------------------------------------------------------
-- 3. Delete the site_plan_imagery entry under any plan for this site
-- (so even non-current plans don't reference the old key)
-- ----------------------------------------------------------------------------
DELETE FROM site_plan_imagery spi
USING site_plans sp
WHERE spi.plan_id = sp.id
  AND sp.site_id = '00ff3af5-dad8-4770-9f70-3edc267a3c92'
  AND spi.key = 'icon_cross_technology';

-- ----------------------------------------------------------------------------
-- Verify
-- ----------------------------------------------------------------------------
SELECT
  (SELECT count(*) FROM assets
   WHERE site_id = '00ff3af5-dad8-4770-9f70-3edc267a3c92'
     AND asset_key = 'icon_cross_technology') AS remaining_assets,
  (SELECT count(*) FROM site_work_items
   WHERE id = '3cacc0dd-4bb0-44a1-a3ca-87a3911423a8'
     AND status != 'wont_fix') AS work_items_not_expired,
  (SELECT count(*) FROM site_plan_imagery spi
   JOIN site_plans sp ON sp.id = spi.plan_id
   WHERE sp.site_id = '00ff3af5-dad8-4770-9f70-3edc267a3c92'
     AND spi.key = 'icon_cross_technology') AS remaining_plan_entries;
-- All three should be 0.

COMMIT;

-- ----------------------------------------------------------------------------
-- Git cleanup (run separately, outside this SQL)
-- ----------------------------------------------------------------------------
-- The deployed image file remains in git. Remove it manually:
--
--   cd <your-sites-repo>/robot-hands.com
--   git pull
--   # Find the file (likely assets/images/icon-cross-technology.jpg or .png):
--   ls assets/images/ | grep -i icon
--   git rm assets/images/icon-cross-technology.jpg  # adjust if extension differs
--   git commit -m "Remove botched icon — superseded by per-actuation-type icons"
--   git push
--
-- After deploy_image_asset re-runs for the new per-type icon work items, the
-- repository will have icon-parallel-jaw.jpg, icon-servo-motor.jpg, etc.