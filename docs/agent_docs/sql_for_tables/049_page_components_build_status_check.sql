-- 049_page_components_build_status_check.sql
-- APPLIED 2026-07-11
--
-- page_components.build_status was free text. That is what let apply_section_edit
-- invent 'approved' unnoticed and remove a live section from every discovery
-- check's audit surface (all filter build_status = 'deployed'). This CHECK turns
-- any future invented value from a silent invisibility into a loud write failure.
-- (PLAN_generalise_fixes_to_fleet.md §4; the drift check
-- page_component_status_drift covers the residual case — a legitimate 'approved'
-- row whose deploy step failed and never advanced.)
--
-- Pre-flight surveys (2026-07-11):
--   data:    SELECT DISTINCT build_status FROM page_components
--            → deployed (597), pending (20). Nothing outside the set.
--   writers: Go literals write only 'deployed' / 'pending' / 'approved';
--            UpdatePageStatusAction mirrors 'deployed';
--            UpdatePageComponentsStatusAction takes config-fed free text, but the
--            only workflow using it (content-reviewer) passes 'approved'.
--   'removed' and 'needs_rebuild' have no writer today but are read-filtered
--   (v3_site_actions.go: build_status != 'removed') and are documented known
--   states in check_page_component_status_drift's knownComponentStatuses —
--   retained so those writers can be added without a migration.
--
-- NULL passes a CHECK by SQL semantics; no NULLs exist and writers always set it.
--
-- Reversal:
--   ALTER TABLE page_components DROP CONSTRAINT page_components_build_status_check;

BEGIN;

-- verify nothing outside the set landed since the survey
DO $$
DECLARE bad INT;
BEGIN
  SELECT COUNT(*) INTO bad FROM page_components
  WHERE build_status IS NOT NULL
    AND build_status NOT IN ('deployed','pending','approved','removed','needs_rebuild');
  IF bad > 0 THEN
    RAISE EXCEPTION 'aborting: % page_components rows carry a build_status outside the proposed set', bad;
  END IF;
END $$;

ALTER TABLE page_components
  ADD CONSTRAINT page_components_build_status_check
  CHECK (build_status IN ('deployed','pending','approved','removed','needs_rebuild'));

-- prove it holds
SELECT conname, pg_get_constraintdef(oid)
FROM pg_constraint
WHERE conrelid = 'page_components'::regclass AND conname = 'page_components_build_status_check';

COMMIT;
