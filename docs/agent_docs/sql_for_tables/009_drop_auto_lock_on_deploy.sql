-- File: 009_drop_auto_lock_on_deploy.sql
-- Date: 2026-07-09
--
-- Drops the auto_lock_on_deploy trigger and function introduced by
-- 008_page_components_and_schema_mode.sql (originally 043_schema_mode_infrastructure.sql).
--
-- WHY
--   The strict-mode subsystem 008 designed was only partially applied and then
--   abandoned:
--     - page_components.schema_snapshot / content_snapshot columns were never
--       created in production (the lock/unlock functions that reference them
--       would error if called);
--     - no Go code reads page_components.schema_mode or sites.strict_mode_trigger;
--     - auto_lock_on_deploy fired exactly once in the system's history
--       (gaswholesalers.com, 2026-04-03), producing one schema_mode='strict' row
--       that nothing consumes.
--
--   The trigger became an active liability once apply_section_edit's build_status
--   handling was fixed: the fix UPDATEs page_components.build_status to 'deployed',
--   which fires this BEFORE UPDATE trigger and would stamp every edited section
--   schema_mode='strict' — making edited rows the only locked rows on any site,
--   for a feature no reader honours.
--
--   Reversibility: the exact function body is preserved at
--   docs/social001_vonc_tiktok_social/minilobby_task/auto_lock_on_deploy.FUNCTION_BACKUP.sql
--
-- SCOPE
--   Drops only the trigger and its function. Leaves the orphaned
--   lock_section_to_strict / unlock_section_for_redesign functions and the
--   schema_mode / strict_mode_trigger columns in place (harmless, and out of
--   scope for this change). Normalises the single legacy strict row.

BEGIN;

DROP TRIGGER IF EXISTS trigger_auto_lock_on_deploy ON page_components;
DROP FUNCTION IF EXISTS auto_lock_on_deploy();

-- Normalise the one legacy row the trigger produced (gaswholesalers.com).
UPDATE page_components
SET schema_mode = NULL, locked_at = NULL, locked_by = NULL
WHERE schema_mode = 'strict';

DO $verify$
DECLARE trg INT; fn INT; strict_rows INT;
BEGIN
  SELECT COUNT(*) INTO trg FROM pg_trigger
    WHERE tgname = 'trigger_auto_lock_on_deploy' AND NOT tgisinternal;
  SELECT COUNT(*) INTO fn FROM pg_proc WHERE proname = 'auto_lock_on_deploy';
  SELECT COUNT(*) INTO strict_rows FROM page_components WHERE schema_mode = 'strict';

  IF trg <> 0 THEN RAISE EXCEPTION 'trigger still present'; END IF;
  IF fn  <> 0 THEN RAISE EXCEPTION 'function still present'; END IF;
  IF strict_rows <> 0 THEN RAISE EXCEPTION '% strict rows remain', strict_rows; END IF;

  RAISE NOTICE 'OK  auto_lock_on_deploy trigger + function dropped, 0 strict rows';
END $verify$;

COMMIT;
