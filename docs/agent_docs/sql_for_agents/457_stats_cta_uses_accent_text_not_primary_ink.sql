-- 457_stats_cta_uses_accent_text_not_primary_ink.sql
--
-- CORRECTS A REGRESSION MIGRATION 456 INTRODUCED, measured at the artefact.
--
-- 456 repointed foreground `--color-primary` declarations to `--color-primary-ink`
-- across 12 templates. It did so **regardless of the ground each declaration sits
-- on**, and that is the defect: `--color-primary-ink` is derived to clear the
-- contrast floor against the PAGE grounds — background, surface, and the composited
-- section overlay (`buildLegibleInkDefaults`, palette_specialised_slots.go:~720).
-- It carries NO guarantee against a fill.
--
-- `.stats-cta` is an ACCENT-FILLED button:
--
--   .stats-cta { background: var(--color-accent, #7dd3fc); color: <foreground>; }
--
-- so on ai-agent-orchestration.com 456 changed its label from `#0D1117` on `#F0A500`
-- (near-black on amber — legible) to `#768eb2` on `#F0A500`, **measured 1.61:1 on
-- 2026-08-18 by scripts/render_audit.py** on `/index.html`, element `a.stats-cta`
-- ("View full report"). One firm failure, introduced by the repair.
--
-- THE RENDERER ALREADY EMITS THE RIGHT TOKEN, and the component library was already
-- expected to use it — palette_specialised_slots.go:105-107 says in as many words:
-- "`--color-primary-text` on a primary-filled button, `--color-accent-text` on an
-- accent-filled one". `--color-accent-text` is derived from the palette TEXT colour
-- against `palette["accent"]` AS ITS GROUND (line 729), which is exactly this case.
-- Verified live before writing this migration: `--color-accent-text` computes to
-- `#294155` on this site (getComputedStyle, not read from a stylesheet).
--
-- SCOPE: one declaration in one template. `system-stats` carried exactly ONE ink
-- reference after 456 and it was this one, so the anchor cannot over-reach; the
-- dry run confirmed ink_before=1 -> ink_left=0, fixed_decls=1.
--
-- ⚠ SIBLINGS DELIBERATELY LEFT ALONE. `system-stats-leo` and `system-stats-leopardess`
-- carry the same `.stats-cta` rule with a BARE `var(--color-primary, #1a1a2e)`. They
-- were not in 456's set (this site does not render them) and they are not broken in
-- the same way — a dark primary on an amber fill is legible. Repointing them to
-- accent-text would be correct-by-symmetry and is NOT done here, because neither
-- site has been measured. Recorded so the asymmetry reads as a decision.
--
-- ⚠ THE GENERAL LESSON, for whoever does the remaining 144 templates: a foreground
-- repoint is only safe when the declaration's own rule block does not set a FILL.
-- Census the block, not the declaration. Of 456's 36 repointed declarations exactly
-- one sat on a fill, which is why the site got better overall (44 -> 33 firm
-- failures) while one element got worse.
--
-- ROLLBACK: 457_stats_cta_uses_accent_text_not_primary_ink_ROLLBACK.sql

BEGIN;

UPDATE content_components
SET html_template = regexp_replace(
      html_template,
      '(background:\s*var\(--color-accent[^;]*\);\s*)color:\s*var\(--color-primary-ink,var\(--color-primary,(\s*)(#[0-9a-fA-F]{3,8})\)\)',
      '\1color: var(--color-accent-text, var(--color-primary,\2\3))',
      'g'),
    updated_at = now()
WHERE name = 'system-stats'
  AND html_template ~ 'background:\s*var\(--color-accent[^;]*\);\s*color:\s*var\(--color-primary-ink,';

DO $$
DECLARE
  fixed      int;
  ink_left   int;
  bare_ink   int;
BEGIN
  SELECT count(*) INTO fixed FROM content_components
   WHERE name = 'system-stats'
     AND html_template LIKE '%color: var(--color-accent-text, var(--color-primary%';
  IF fixed <> 1 THEN
    RAISE EXCEPTION '457: expected exactly 1 system-stats row carrying the accent-text rule, found %', fixed;
  END IF;

  -- The whole point: no ink may remain on an accent-filled rule in this template.
  SELECT count(*) INTO ink_left FROM content_components
   WHERE name = 'system-stats'
     AND html_template ~ 'background:\s*var\(--color-accent[^;]*\);\s*color:\s*var\(--color-primary-ink,';
  IF ink_left <> 0 THEN
    RAISE EXCEPTION '457: % accent-filled rule(s) in system-stats still paint with --color-primary-ink', ink_left;
  END IF;

  -- 456's invariant still holds: never a bare ink reference anywhere.
  SELECT count(*) INTO bare_ink FROM content_components
   WHERE html_template ~ 'var\(\s*--color-(primary|accent)-ink\s*\)';
  IF bare_ink <> 0 THEN
    RAISE EXCEPTION '457: % content_component(s) carry a BARE ink reference', bare_ink;
  END IF;

  RAISE NOTICE '457 OK: .stats-cta now paints with --color-accent-text (two-level fallback). Re-render system-stats placements to propagate.';
END $$;

COMMIT;
