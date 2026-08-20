-- 496_editorial_chart_components_use_legible_ink.sql
--
-- Repoints three text colours in the two evidence-chart components onto the
-- renderer's LEGIBLE-INK companions, fixing measured contrast failures on live
-- editorial pages without introducing a single colour literal.
--
-- WHAT WAS MEASURED (scripts/render_audit.py, 2026-08-20, at the served pages,
-- both sites' stylesheets confirmed healthy at 25-27 KB first — a clobbered
-- stylesheet reports FEWER contrast failures, so that control is not optional):
--   robot-hands  /insights/robot-demand-step-change.html
--     .evidence-chart__eyebrow   1.14:1  rgb(26,31,46)  on rgb(15,18,24)
--     .ev-ts__sources a          4.38:1  x5 (one per observation)
--     .ev-ts__eyebrow            PASSED on this palette
--   dartsonline  /insights/darts-calendar-density.html
--     .evidence-chart__eyebrow   1.11:1  rgb(26,31,46)  on rgb(17,21,32)
--     .ev-ts__sources a          3.71:1  x5
--     .ev-ts__eyebrow            4.24:1  <- fails HERE and not on robot-hands
--
-- WHY A TOKEN AND NOT A COLOUR. The obvious fix — pick an ink that clears AA on
-- every palette carrying these components — is IMPOSSIBLE. Those palettes include
-- light-background sites (leopardessconsulting #0D0D0D-ish ink, noted.co.uk) and
-- dark ones (dartsonline #111520, robot-hands #0F1218). No single literal clears
-- AA on both, so that approach converges on a value that fails two sites while
-- appearing to have been proven. The renderer already computes a per-palette
-- answer: buildLegibleInkDefaults / legibleInkFor
-- (platform/orchestration/actions/palette_specialised_slots.go), live since
-- v1.0.1298, register entry VIZ-014. Measured in the served CSS 2026-08-20:
--   robot-hands  --color-primary-ink #94a0c2  --color-accent-ink #f77f47
--   dartsonline  --color-primary-ink #94a0c2  --color-accent-ink #f18072
-- (the accent companion differs per site, which is the tell that it is derived
-- rather than a shared literal.) Its target is inkMinContrast = 5.0, ABOVE the
-- 4.5 AA floor; inkFloorContrast (4.5) is a separate constant for the `-text`
-- slots and a test fails if the two are merged.
--
-- WHY THE TWO-LEVEL FALLBACK IS LOAD-BEARING, NOT TIDINESS. Written as
-- `var(--color-primary-ink, var(--color-primary, #1e40af))`, an ABSENT companion
-- falls through to today's exact behaviour. That matters three ways:
--   1. oufe.com's stylesheet is currently clobbered (bugs_open/198) and therefore
--      has no companion — this change is a NO-OP there, not a breakage, so the
--      outstanding restores are NOT a dependency of this migration;
--   2. it is how the mechanism's own kill-switch works (`legible_ink_enabled`
--      false emits nothing and every consumer falls back), quoted from
--      buildLegibleInkDefaults' own comment;
--   3. the hard-coded literal is preserved as the last resort, so the component
--      still renders valid CSS with no palette at all.
--
-- ⚠ THE TRAP THIS AVOIDS, from VIZ-014's own corrected history: between
-- 2026-08-06 and 2026-08-14 `-ink` resolved in practice to `--color-text`, so a
-- repoint silently STRIPPED the brand colour — and because render_audit.py
-- measures contrast, de-branding scores a CLEAN PASS. Fixed 2026-08-14:
-- colour.LegibleVariant now gets first refusal and moves the source in HSL
-- LIGHTNESS ONLY, hue and saturation preserved. Verified live before writing this
-- (the two sites' accent companions differ, and both differ from --color-text).
-- ANY FUTURE REPOINT MUST RE-CHECK THAT, because a clean audit cannot see it.
--
-- BLAST RADIUS. evidence-chart: 8 instances / 5 sites; evidence-timeseries:
-- 3 instances / 3 sites. Live effect is bounded by locking: every editorial
-- instance is `lock_type='permanent'` and keeps its STORED rendered_html until
-- deliberately re-rendered, so nothing changes on a live page until this lane
-- re-renders it. Unlocked instances (fundamentallyai, leopardessconsulting) pick
-- the fix up on their next ordinary render, which is the intent.
--
-- ROLLBACK: 496_..._ROLLBACK.sql restores both templates from the backup table
-- this migration creates.

\set ON_ERROR_STOP on

CREATE TABLE IF NOT EXISTS content_components_backup_20260820_legible_ink AS
SELECT * FROM content_components
 WHERE function IN ('evidence-chart','evidence-timeseries') AND is_active;

BEGIN;

-- Guard: the exact strings must be present exactly once each, or the anchors
-- have moved and a blind replace would hit the wrong rule (or nothing).
DO $$
DECLARE n int;
BEGIN
  SELECT count(*) INTO n FROM content_components
   WHERE function='evidence-chart' AND is_active
     AND html_template LIKE '%color: var(--color-primary, #1e40af);%';
  IF n <> 1 THEN RAISE EXCEPTION 'evidence-chart eyebrow anchor: expected 1 row, got %', n; END IF;

  SELECT count(*) INTO n FROM content_components
   WHERE function='evidence-timeseries' AND is_active
     AND html_template LIKE '%color: var(--color-accent, #c49a3c); margin: 0 0 0.5rem; }%'
     AND html_template LIKE '%.ev-ts__sources a { color: var(--color-accent, #c49a3c); }%';
  IF n <> 1 THEN RAISE EXCEPTION 'evidence-timeseries anchors: expected 1 row, got %', n; END IF;

  IF EXISTS (SELECT 1 FROM content_components
              WHERE function IN ('evidence-chart','evidence-timeseries') AND is_active
                AND html_template LIKE '%-ink, var(--color-%') THEN
    RAISE EXCEPTION 'already repointed - refusing double apply';
  END IF;
END $$;

UPDATE content_components
   SET html_template = replace(html_template,
         'color: var(--color-primary, #1e40af);',
         'color: var(--color-primary-ink, var(--color-primary, #1e40af));'),
       updated_at = now()
 WHERE function='evidence-chart' AND is_active;

UPDATE content_components
   SET html_template = replace(
         replace(html_template,
           'color: var(--color-accent, #c49a3c); margin: 0 0 0.5rem; }',
           'color: var(--color-accent-ink, var(--color-accent, #c49a3c)); margin: 0 0 0.5rem; }'),
         '.ev-ts__sources a { color: var(--color-accent, #c49a3c); }',
         '.ev-ts__sources a { color: var(--color-accent-ink, var(--color-accent, #c49a3c)); }'),
       updated_at = now()
 WHERE function='evidence-timeseries' AND is_active;

-- Verify: three repoints present, and NO bare pre-repoint form left behind.
DO $$
DECLARE n int;
BEGIN
  SELECT count(*) INTO n FROM content_components
   WHERE function='evidence-chart' AND is_active
     AND html_template LIKE '%var(--color-primary-ink, var(--color-primary, #1e40af));%';
  IF n <> 1 THEN RAISE EXCEPTION 'evidence-chart repoint did not land'; END IF;

  SELECT count(*) INTO n FROM content_components
   WHERE function='evidence-timeseries' AND is_active
     AND html_template LIKE '%var(--color-accent-ink, var(--color-accent, #c49a3c)); margin%'
     AND html_template LIKE '%.ev-ts__sources a { color: var(--color-accent-ink, var(--color-accent, #c49a3c)); }%';
  IF n <> 1 THEN RAISE EXCEPTION 'evidence-timeseries repoints did not land'; END IF;

  -- the eyebrow/link rules must no longer carry the UNWRAPPED token
  IF EXISTS (SELECT 1 FROM content_components
              WHERE function='evidence-chart' AND is_active
                AND html_template LIKE '%  color: var(--color-primary, #1e40af);%') THEN
    RAISE EXCEPTION 'evidence-chart still carries the unwrapped eyebrow colour';
  END IF;
END $$;

COMMIT;
