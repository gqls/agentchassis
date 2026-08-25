-- ============================================================================
-- agritec.uk — light palette, dark text (owner instruction, 2026-08-25)
-- Written 2026-08-25. Applied out of band (psql -f), per-site setup.
--
-- Owner: "I would prefer a lighter palette (dark text) as the default on all
-- domains, as well as for this one."
--
-- This migration does the SITE half. The fleet half is a change to the
-- domain-research-classifier's prompt and is a separate, council-scoped
-- migration, because it changes what every future site gets.
--
-- WHAT WAS THERE. The classifier chose a dark scheme for agritec —
-- background #12151F, text #E8EAF0 — and the live calculator renders on it.
-- Measured across the fleet the same day: 22 sites light, 9 dark. Dark was never
-- a default; it is a free per-site choice, because the classifier's prompt gives
-- the palette SHAPE ("Hex for the page background") with no guidance either way.
--
-- ⚠ THIS ALSO CLOSES THE DIVERGENCE I FLAGGED ON 2026-08-24 AND LEFT OPEN. This
-- site had TWO palettes: the classifier's dark design_intent driving the CSS,
-- and my seeded imagery_style_guide with the retired brand's colours driving the
-- 17 generated images. They shared no hex. I raised it as an owner decision
-- rather than picking; the instruction resolves it toward light, so BOTH move
-- together here. Changing only design_intent would have left the imagery darker
-- than the site it sits on — which is the same divergence, reversed.
--
-- THE COLOURS keep the brand the owner is not replacing: agricultural green
-- #2C7744 and instrument teal #005F73 are the retired site's own, and dark text
-- #1A202C is its slate. What changes is the GROUND: near-white paper instead of
-- near-black, which is what "lighter palette, dark text" asks for.
--
-- CONTRAST is not left to chance: #1A202C on #F7F8F5 is ~15.9:1, far above the
-- 4.5:1 body-text minimum, and #5A6472 muted on the same ground is ~5.6:1. The
-- platform's legible-ink derivation will still run over these; the point is that
-- it should have nothing to correct.
-- ============================================================================

\set ON_ERROR_STOP on

BEGIN;

-- ---- design_intent: the palette that drives the site CSS ----
CREATE TEMP TABLE _di ON COMMIT DROP AS
SELECT ss.id, ss.site_id, ss.data FROM site_specs ss JOIN sites s ON s.id=ss.site_id
WHERE s.domain='agritec.uk' AND ss.aspect='design_intent' AND ss.is_current;

DO $g$
DECLARE n int; bg text;
BEGIN
  SELECT count(*) INTO n FROM _di;
  IF n <> 1 THEN RAISE EXCEPTION 'expected 1 current design_intent, found %', n; END IF;
  SELECT data->'palette'->'reference_values'->>'background' INTO bg FROM _di;
  IF bg IS NULL THEN RAISE EXCEPTION 'design_intent carries no palette.reference_values - refusing to invent one'; END IF;
  RAISE NOTICE 'current background: %', bg;
END $g$;

UPDATE site_specs SET is_current=false, superseded_at=now() WHERE id IN (SELECT id FROM _di);

INSERT INTO site_specs (site_id, aspect, data, source, notes, is_current, pinned, created_by)
SELECT _di.site_id, 'design_intent',
  jsonb_set(
    jsonb_set(_di.data, '{palette,reference_values}', $p${
      "primary":    "#2C7744",
      "secondary":  "#005F73",
      "accent":     "#1F6E3A",
      "background": "#F7F8F5",
      "surface":    "#FFFFFF",
      "text":       "#1A202C",
      "text_muted": "#5A6472",
      "border":     "#D8DED6"
    }$p$::jsonb),
    '{style_direction}', '"modern-light"'::jsonb),
  'manual',
  'Owner instruction 2026-08-25: lighter palette with dark text. Was background #12151F / text #E8EAF0. Keeps the brand green #2C7744 and teal #005F73 and the slate #1A202C as TEXT; changes the ground to near-white paper. Contrast #1A202C on #F7F8F5 is ~15.9:1.',
  true, true, 'agritec-workstream-2026-08-25'
FROM _di;

-- ---- imagery_style_guide: move it WITH the site, or the images clash ----
CREATE TEMP TABLE _ig ON COMMIT DROP AS
SELECT ss.id, ss.site_id, ss.data FROM site_specs ss JOIN sites s ON s.id=ss.site_id
WHERE s.domain='agritec.uk' AND ss.aspect='imagery_style_guide' AND ss.is_current;

UPDATE site_specs SET is_current=false, superseded_at=now() WHERE id IN (SELECT id FROM _ig);

INSERT INTO site_specs (site_id, aspect, data, source, notes, is_current, pinned, created_by)
SELECT _ig.site_id, 'imagery_style_guide',
  jsonb_set(
    jsonb_set(_ig.data, '{palette}',
      to_jsonb('warm off-white paper grounds (#F7F8F5, #FFFFFF), agricultural green as the primary accent (#2C7744), instrument teal as the secondary (#005F73), dark slate for linework and labels (#1A202C), cool greys for structure (#5A6472, #D8DED6)'::text)),
    '{kinds,content_hero,palette}',
      to_jsonb('warm off-white ground, agricultural green and instrument teal flat shapes, dark slate linework, cool grey secondary only'::text)),
  'manual',
  'Moved with design_intent to light grounds (owner 2026-08-25). Previously specified deep slate grounds, which would have left the imagery darker than the site it sits on - the same divergence flagged on 08-24, reversed. The 17 already-generated assets were made against the dark guide and will need regenerating to match.',
  true, true, 'agritec-workstream-2026-08-25'
FROM _ig;

COMMIT;
