-- 541 (_HOLD) — bugs_open/198 + /211: enable the stylesheet_gutted discovery
-- check on design-discovery-agent.
--
-- ✅ HOLD DISCHARGED 2026-08-26 — this file is now RELEASED (the _HOLD suffix
-- is off). How the two conditions were satisfied, each with its control:
--
--   1. ROLLED: every live chassis pod (217/217 seen within 30 min, ONE distinct
--      build commit, 2fb40a960) runs an image whose commit has e34b33a36 — this
--      check's birth — as an ancestor. Control: today's HEAD correctly NOT an
--      ancestor of the rolled commit.
--   2. CAPABILITY, probed on every pod: the binary now REPORTS its checks
--      itself — service_binary_capabilities kind='discovery_check',
--      name='stylesheet_gutted' present for 217/217 live pods; negative control
--      'stylesheet_gutted_NOTREAL' = 0 rows. Stronger than (and superseding)
--      the exec-grep recipe below, which predates the capability registry; the
--      recipe is kept as written for images that lack the registry.
--
--   RE-CALIBRATED before applying, with the check's OWN code, not a proxy
--   (the 2026-08-21 correction below is why that distinction is load-bearing):
--   the real Run() driven over the exported live corpus of all **31** deployed
--   sites as of 2026-08-26 (corpus exported with this check's own predicate
--   SQL, stylesheets fetched live). Result: **0 filed / 29 resolved /
--   2 declined-to-judge** (lampenkap.com — its linked stylesheet 404s, which is
--   asset_reference_404's finding, not ours; loanandmortgagecalculator.co.uk —
--   skip branch not identified [UNVERIFIED], files nothing either way). The
--   08-21 "0 of 25" figure held at 0 of 31; six sites were born in between.
--
--   Applied by hand 2026-08-26 09:07Z (runner --apply would have taken other
--   lanes' pending files); ledger row '541_enable_stylesheet_gutted_check.sql',
--   applied_by=record-only, checksum = this file's post-discharge md5. Live row
--   verified independently after COMMIT: 24 checks, stylesheet_gutted present.
--   Companion liveConfiguredChecks edit rides the same commit, as the header
--   below demands.
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
-- EXPECTED VOLUME ON FIRST ROTATION: **ZERO**, and that is a corrected figure.
--
-- > CORRECTED 2026-08-21, same session, before this file was ever applied. The
-- > first version of this comment said "expect ~1 finding (cookly.uk)". That was
-- > measured with a `:root`-presence PROXY, not with the check's own predicate,
-- > and it was wrong in both directions.
-- >
-- > Running the REAL predicate across all 25 deployed/active sites showed the
-- > check as first written would have filed on **NINETEEN** — seventeen of them
-- > for the same four component-invented names (--color-hero-title,
-- > --color-hero-subtitle, --color-secondary-text, --color-secondary-hover)
-- > that NO site's stylesheet has ever defined, including in the pre-clobber
-- > originals. That is a real defect and a DIFFERENT one; filing it here would
-- > have buried this check's signal under seventeen copies of another.
-- > The check now gates on the renderer's GUARANTEED vocabulary
-- > (rendererGuaranteedTokens, kept in step with canonicalCSSTokens by a parity
-- > test). Re-measured with that gate: **0 of 25 sites** would file.
-- > cookly.uk, the one genuine positive when this work started, was restored
-- > mid-session and now serves 18,047 bytes.
--
-- So this ships as a REGRESSION GUARD with no live positive — the same posture
-- as asset_reference_404 (IMP-051), and the same caveat: a guard with no live
-- positive can rot unexercised, so every branch is proven by an induced fault in
-- the test file rather than by hope. A clean first rotation is therefore the
-- EXPECTED result and is not evidence the check works. What would be evidence:
-- re-point the test fixtures at a real gutted stylesheet, or wait for the next
-- incident and confirm it files.
--
-- COUNCIL ROUND d3187418 — APPROVED r1, and the two checkable objections were
-- CHECKED rather than accepted (2026-08-21):
--
--   edit-quality [medium]: "four agent types carry TWO active rows and only the
--   higher version loads; this UPDATE has no version filter." MEASURED — all
--   four discovery agents carry exactly ONE active non-snapshot row today
--   (design/completeness/quality/availability, count=1 each). The hazard is real
--   in general and does not apply here; the DO/RAISE expecting n=1 is what turns
--   a future second row into a loud abort rather than a silent half-write, which
--   is the protection the objection asks for.
--
--   prior-art [medium]: "the whole HOLD rests on an asserted file:line claim."
--   VERIFIED in the source: discovery_checks.go has `defer tx.Rollback()` at
--   :141, the unregistered-name `return nil, fmt.Errorf(...)` at ~:208, and
--   `tx.Commit()` at :286 — so the return does precede the commit inside the
--   deferred rollback, and earlier checks' findings in that run ARE discarded.
--   The hold is justified on read code, not on memory.
--
--   Also found while checking: there IS an escape hatch —
--   `allow_unregistered_checks` (config bool, default false, discovery_checks.go
--   :198). Setting it true tolerates an unrolled name with a Warn instead of
--   failing the step. That is a legitimate alternative to holding this file, but
--   it weakens the guard for EVERY check on that agent, so the hold stays the
--   default and this is recorded as an option, not a recommendation.
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
