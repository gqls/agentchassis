-- 475 (_HOLD) — bug 299: enable the cta_nonpage_destination discovery check on
-- completeness-discovery-agent (the one live agent carrying misdirected_cta).
--
-- ⚠ HELD until the image carrying commit 757a0890a is POD-VERIFIED live on
-- agent-chassis (build-provenance stamp + `git merge-base --is-ancestor
-- 757a0890a <stamp>`). An unregistered check name fails the WHOLE
-- run_discovery_checks step on an old binary — the asset_reference_404
-- precedent. Rename away the _HOLD suffix only after that check passes.
--
-- What it enables: review-only findings (needs_human_review, no handler) —
-- cta_names_nonpage_destination (copy names a real page, href is tel:/mailto:)
-- and cta_tel_malformed. Calibrated 2026-08-18 over the whole fleet before the
-- code was committed: 17 findings, 17/17 hand-reviewed true, 0 false
-- (bugfix_299_cta_dials_phone/CALIBRATION_2026-08-18_cta_nonpage_report.md).
-- Expect ~that volume on first rotation, then a trickle.
--
-- The COMPANION EDIT the roster test demands: add "cta_nonpage_destination" to
-- liveConfiguredChecks in discovery_checks_registration_test.go IN THE SAME
-- COMMIT that applies this file — that list asserts the live agents' roster,
-- and this migration is what changes the roster.

BEGIN;

CREATE TABLE IF NOT EXISTS _backup_475_cta_nonpage_check AS
  SELECT id, type, default_config, now() AS backed_up_at
    FROM agent_definitions
   WHERE type = 'completeness-discovery-agent' AND is_active
     AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

-- Idempotent append (the 092 pattern): no-op when the name is already present.
UPDATE agent_definitions
   SET default_config = jsonb_set(default_config,
         '{workflow,steps,run_checks,config,checks}',
         (default_config #> '{workflow,steps,run_checks,config,checks}') || '["cta_nonpage_destination"]'::jsonb)
 WHERE type = 'completeness-discovery-agent' AND is_active
   AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL
   AND (default_config #> '{workflow,steps,run_checks,config,checks}') IS NOT NULL
   AND NOT (default_config #> '{workflow,steps,run_checks,config,checks}') ? 'cta_nonpage_destination';

-- Induced verification (DO/RAISE — a SELECT cannot stop the COMMIT):
DO $$
DECLARE n int;
BEGIN
  SELECT count(*) INTO n
    FROM agent_definitions
   WHERE type = 'completeness-discovery-agent' AND is_active
     AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL
     AND (default_config #> '{workflow,steps,run_checks,config,checks}') ? 'cta_nonpage_destination';
  IF n <> 1 THEN
    RAISE EXCEPTION '475: expected exactly 1 completeness-discovery-agent carrying cta_nonpage_destination, found % — step path or roster drifted, investigate before applying', n;
  END IF;
END $$;

COMMIT;
