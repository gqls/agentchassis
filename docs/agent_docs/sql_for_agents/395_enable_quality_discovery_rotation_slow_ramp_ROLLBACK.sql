-- ============================================================================
-- 395 ROLLBACK — disable the quality discovery rotation, restore 3600s
-- ============================================================================
-- Run BY HAND, deliberately. The migration runner never picks up an
-- UPPERCASE-suffixed sidecar (SIDECAR_RE, bugs_open/007).
--
-- Restores the pre-395 state exactly: enabled=false, interval_seconds=3600.
--
-- NOTE ON WHAT ROLLBACK CANNOT UNDO: any `detected` work items the rotation
-- already filed stay filed, and any `site_discovery_rotation.last_selected_at`
-- stamps stay stamped (so those sites will not be re-selected for 7 days from
-- their stamp, even after a later re-enable). That is by design — the stamps are
-- the fairness ledger, and un-stamping would re-introduce the starvation SCH-025
-- exists to fix. If you genuinely need the findings gone, cancel them by
-- item_type + created_at window; do not delete rotation rows.
-- ============================================================================

BEGIN;

DO $$
DECLARE
  v_enabled boolean;
BEGIN
  SELECT enabled INTO v_enabled
    FROM scheduled_tasks WHERE name = 'site-discovery-rotation-quality';

  IF NOT FOUND THEN
    RAISE EXCEPTION '395 ROLLBACK: scheduled_tasks row site-discovery-rotation-quality is MISSING';
  END IF;

  IF NOT v_enabled THEN
    RAISE NOTICE '395 ROLLBACK: already disabled — restoring interval only';
  END IF;
END $$;

UPDATE scheduled_tasks
   SET enabled          = false,
       interval_seconds = 3600,
       updated_at       = now()
 WHERE name = 'site-discovery-rotation-quality';

DO $$
DECLARE
  v_enabled  boolean;
  v_interval integer;
BEGIN
  SELECT enabled, interval_seconds INTO v_enabled, v_interval
    FROM scheduled_tasks WHERE name = 'site-discovery-rotation-quality';

  IF v_enabled OR v_interval <> 3600 THEN
    RAISE EXCEPTION '395 ROLLBACK: post-check FAILED — enabled=%, interval=%', v_enabled, v_interval;
  END IF;

  RAISE NOTICE '395 ROLLBACK: OK — disabled, interval restored to 3600s';
END $$;

COMMIT;
