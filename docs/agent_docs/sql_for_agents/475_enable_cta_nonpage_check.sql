-- 475 (_HOLD) — bug 299: enable the cta_nonpage_destination discovery check on
-- completeness-discovery-agent (the one live agent carrying misdirected_cta).
--
-- ✅ HOLD DISCHARGED 2026-08-19 — this file is now RELEASED (the _HOLD suffix is
-- off). What the hold asked for, and how it was actually satisfied:
--
--   The hold said: pod-verify the build-provenance stamp, then
--   `git merge-base --is-ancestor 757a0890a <stamp>`.
--   THAT CHECK WAS NOT AVAILABLE. The fleet rolled to v1.0.1316 at 17:13Z and
--   the "build provenance" line is a STARTUP line: the pods' earliest retained
--   log is 20:08Z, so the stamp had already scrolled (the documented
--   CLAUDE.md/LANDMINES trap — an absent line means "not in range", never
--   "unstamped"). A discovery grep for "some 40-hex string" in the binary is
--   forbidden for the same corpus's reasons, and the binary carries ONE stamp
--   string, not its ancestry, so probing for 757a0890a returns absent on a
--   binary that certainly contains it.
--
--   SUBSTITUTED, and it is the STRONGER check: probe for the CAPABILITY the
--   hold exists to guarantee, not for a commit that proxies it. Measured
--   2026-08-19 on BOTH live pods (agent-chassis-5ddd9744-86nqf / -8jlqh,
--   image v1.0.1316), each with a negative control that came out absent:
--     cta_nonpage_destination        PRESENT (both pods)
--     cta_names_nonpage_destination  PRESENT
--     cta_tel_malformed              PRESENT
--     cta_nonpage_destination_NOTREAL  absent  <- control
--   i.e. the very name this migration is about to put in a checks array is
--   demonstrably registered in the binary that will read it. That is the
--   property the fail-fast arm needs; a commit sha was only ever a proxy for it.
--
--   Also confirmed by reading the code (the council's prior_art_librarian seat
--   flagged that this file and the submission's edit-6 rationale disagreed —
--   this file was the correct half): discovery_checks.go:198-216 —
--   `allowUnregistered := configBoolOrDefault(config,
--   "allow_unregistered_checks", false)` and, on a registry miss with the lever
--   false, `return nil, fmt.Errorf("discovery check %q is not registered …")`.
--   So an unregistered name DOES fail the whole step, and worse than stated:
--   the return happens before tx.Commit() inside defer tx.Rollback(), so it
--   also discards every EARLIER check's findings in the same run.
--   The asset_reference_404 precedent (bugs_closed/084) is the same order,
--   walked the same way: probe the binary, THEN edit the checks array.
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

-- README rule: every migration touching agent_definitions opens with a snapshot.
-- The bespoke backup table below is kept as well — it is narrower and survives
-- independently of the snapshot machinery.
SELECT snapshot_agent('completeness-discovery-agent',
  '475_enable_cta_nonpage_check: pre-update');

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
