-- 541 (_HOLD) — bugs_open/198 + /211: enable the stylesheet_gutted discovery
-- check on design-discovery-agent.
--
-- ⚠ HELD DELIBERATELY. DO NOT APPLY UNTIL BOTH CONDITIONS BELOW ARE MET.
--
-- WHY IT IS HELD, and it is not ceremony. discovery_checks.go:198-216 reads
-- `allow_unregistered_checks` (default false) and, on a registry miss, does
-- `return nil, fmt.Errorf("discovery check %q is not registered …")`. That return
-- happens BEFORE tx.Commit() inside defer tx.Rollback(), so a name in this array
-- that the running binary does not know does not merely skip its own check — it
-- fails the whole step and DISCARDS EVERY EARLIER CHECK'S FINDINGS in the same
-- run. Config is live the instant it is applied; Go is inert until an image is
-- built and rolled. Applying this file before the roll therefore breaks design
-- discovery fleet-wide for as long as the gap lasts.
--
-- THE TWO CONDITIONS:
--
--   1. A chassis image carrying platform/orchestration/actions/discovery_checks/
--      check_stylesheet_gutted.go has ROLLED.
--
--   2. The CAPABILITY is probed on every live chassis pod, WITH A NEGATIVE
--      CONTROL IN THE SAME BREATH. Probe for the capability, not for a commit
--      that proxies it — 475's hold-discharge note is the worked precedent and
--      explains why the provenance-stamp route is not available (the stamp is a
--      STARTUP line and scrolls; a discovery grep for "some 40-hex string" is
--      forbidden; the binary carries one stamp string, not its ancestry):
--
--        for p in $(kubectl -n ai-persona-system get pods -l app=agent-chassis \
--                     -o name | cut -d/ -f2); do
--          echo "== $p"
--          kubectl -n ai-persona-system exec "$p" -- \
--            grep -ac 'stylesheet_gutted' /proc/1/exe          # expect >= 1
--          kubectl -n ai-persona-system exec "$p" -- \
--            grep -ac 'stylesheet_gutted_NOTREAL' /proc/1/exe  # expect 0 (control)
--        done
--
--      A control that comes out PRESENT means the probe matches everything and
--      proves nothing (the 40-zeros trap). Never use `strings` — it is absent
--      from the debian-slim images and behind a customary 2>/dev/null its failure
--      is indistinguishable from "not stamped".
--
-- THE COMPANION EDIT THE ROSTER TEST DEMANDS: add "stylesheet_gutted" to
-- liveConfiguredChecks in discovery_checks_registration_test.go IN THE SAME
-- COMMIT that applies this file. That fixture asserts the live agents' roster and
-- this migration is what changes the roster; they must move together or the test
-- is asserting a roster that no longer exists.
--
-- WHY design-discovery-agent: it is the agent that already carries
-- asset_reference_404 (the check whose blind spot this one fills — it fetches
-- this very URL and scores HTTP status only, so a 136-byte 200 reads as healthy
-- and RETRACTS), plus missing_css, palette_contrast and undeployed_assets. Its
-- check_pipeline is 'design', which is this finding's pipeline. Roster read live
-- 2026-08-21: 23 checks, asset_reference_404 among them.
--
-- WHAT IT ENABLES: one flag-only finding per site (needs no handler; the repair
-- is restore-from-git or a webdesign-agent run, both judgements — and routing it
-- automatically at webdesign-agent would re-roll the palette, which
-- check_generic_theme.go records causing four CSS rewrites in one day, one of
-- them putting a light background on a dark site). It self-clears via
-- CheckResult.Resolved once the served stylesheet defines everything again.
--
-- EXPECTED VOLUME ON FIRST ROTATION. Calibrated by hand across the fleet
-- 2026-08-21, before this file was written: 25 deployed/active sites, of which
-- exactly one still serves a gutted stylesheet (cookly.uk, 504 bytes, 0 :root —
-- its theme row was restored 2026-08-20 but the file deploy was refused by a
-- permission classifier and is still outstanding). remortgagecalculator.uk and
-- loanzy.uk were in that state the same morning and were restored in the same
-- session. So expect ~1 finding on the first full rotation, not a wave — and if
-- it reports ZERO on a rotation that included cookly.uk, that is a bug in the
-- check, not good news.
--
-- Register entry: IMP-055 (register/improvement-loop.md).

BEGIN;

-- README rule: every migration touching agent_definitions opens with a snapshot.
-- The bespoke backup table below is kept as well — it is narrower and survives
-- independently of the snapshot machinery.
SELECT snapshot_agent('design-discovery-agent',
  '541_enable_stylesheet_gutted_check: pre-update');

CREATE TABLE IF NOT EXISTS _backup_541_stylesheet_gutted_check AS
  SELECT id, type, default_config, now() AS backed_up_at
    FROM agent_definitions
   WHERE type = 'design-discovery-agent' AND is_active
     AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

-- Idempotent append (the 092 pattern): no-op when the name is already present.
UPDATE agent_definitions
   SET default_config = jsonb_set(default_config,
         '{workflow,steps,run_checks,config,checks}',
         (default_config #> '{workflow,steps,run_checks,config,checks}') || '["stylesheet_gutted"]'::jsonb)
 WHERE type = 'design-discovery-agent' AND is_active
   AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL
   AND (default_config #> '{workflow,steps,run_checks,config,checks}') IS NOT NULL
   AND NOT (default_config #> '{workflow,steps,run_checks,config,checks}') ? 'stylesheet_gutted';

-- Induced verification (DO/RAISE — a SELECT cannot stop the COMMIT):
DO $$
DECLARE n int;
BEGIN
  SELECT count(*) INTO n
    FROM agent_definitions
   WHERE type = 'design-discovery-agent' AND is_active
     AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL
     AND (default_config #> '{workflow,steps,run_checks,config,checks}') ? 'stylesheet_gutted';
  IF n <> 1 THEN
    RAISE EXCEPTION '541: expected exactly 1 design-discovery-agent carrying stylesheet_gutted, found % — step path or roster drifted, investigate before applying', n;
  END IF;
END $$;

COMMIT;
