-- B16.3 (owner decision 2026-07-19) — give gamesdesign.co.uk, finetuning.uk and
-- leopardessconsulting.co.uk an `imagery_style_guide`, each with a
-- `kinds.content_hero` override, BEFORE the tool-surface sweep spends 19
-- generations on them.
--
-- WHY (see /bugs_open/027): D14 gave `content_hero` its own kind (flat duotone
-- illustration, routed to Banana) but did NOT add it to directionAppliesToKind.
-- So on a site with no `kinds.content_hero` override the generator falls back to
-- the free-text design_intent.imagery_direction — written to describe PHOTOGRAPHY
-- — and hands it to a flat-illustration kind. robot-hands.com was the only site
-- in the fleet with a style guide at all, so it was the only site where the kind
-- behaved. These three rows close that for the three sites that are about to
-- generate; the fleet-wide code half stays open in 027 §5(a).
--
-- Config only: site_specs is read at generation time, so this is LIVE
-- IMMEDIATELY — no image build, no chassis roll.
--
-- SHAPE follows the proven robot-hands row (SQL_2026-07-17_d14_style_guide_
-- content_hero_override.sql + the 07-18 `avoid` tightening): an override
-- REPLACES the guide-level fields wholesale for its kind, so each override
-- restates palette/medium/mood/avoid in full rather than relying on a merge.
--
-- PALETTES are taken from each site's design_intent.palette.reference_values
-- where it has them (gamesdesign, leopardess) — that is the pinned source of
-- truth for colour (see the generic_theme colour-churn landmine). finetuning.uk
-- has no reference_values, so its colours come from design_intent.colour_mood
-- prose and are marked as such in `notes`.
--
-- REFERENCE ANCHORS — each decided by LOOKING at the candidate, not by assuming:
--   leopardess  → ["hero_home"]. Inspected 2026-07-19: flat antique-gold linework
--                 on near-black, no text, generous negative space — exactly the
--                 "same hand as the logo" its design_intent asks for, and Banana-
--                 generated (banana/gemini-3-pro-image-preview), so it is the
--                 house style rather than an SDXL accident.
--   gamesdesign → []. Its heroes are all SDXL photographic, and the homepage's
--                 own /assets/images/hero.jpg does not even serve (301 → 404).
--                 Anchoring a flat kind to a photographic hero is the
--                 2026-05-20 contamination failure.
--   finetuning  → []. Inspected 2026-07-19: its `hero` is a teal/charcoal mark on
--                 a PALE GREY ground — anchoring to it would import exactly the
--                 white background that D14's `avoid` had to be tightened to
--                 exclude on robot-hands.
--
-- Applied: 2026-07-19. Backup + verify per house practice.

\set ON_ERROR_STOP on
BEGIN;

-- Backup: every current imagery_style_guide row before we touch anything.
CREATE TABLE IF NOT EXISTS site_specs_imagery_guide_backup_20260719 AS
SELECT * FROM site_specs WHERE aspect = 'imagery_style_guide';

-- Supersede any existing current guide on the three targets (there are none
-- today — verified — but this keeps the migration re-runnable and respects the
-- partial unique index on (site_id, aspect) WHERE is_current).
UPDATE site_specs sd
   SET is_current = false, superseded_at = now(), updated_at = now()
  FROM sites s
 WHERE s.id = sd.site_id
   AND sd.aspect = 'imagery_style_guide'
   AND sd.is_current = true
   AND s.domain IN ('gamesdesign.co.uk', 'finetuning.uk', 'leopardessconsulting.co.uk');

-- ── gamesdesign.co.uk ────────────────────────────────────────────────────────
-- Voice: "professional-grade developer utility… visually austere". Its
-- design_intent explicitly rejects photography ("Heavy imagery, photography, or
-- stock images — the aesthetic is intentionally data and text driven"), so even
-- the BASE voice here is flat/schematic rather than photographic. Cyan #00bcd4
-- on near-black #121212 is non-negotiable per its palette guidance.
INSERT INTO site_specs (site_id, aspect, data, source, created_by, notes, is_current)
SELECT s.id, 'imagery_style_guide', $json$
{
  "medium": "flat technical diagram illustration, schematic linework, HUD and debug-overlay visual language, no photography",
  "mood": "precise, systems-oriented, utilitarian, confident, restrained",
  "palette": "near-black ground (#121212) with slightly elevated dark surface (#1e1e1e), a single cyan accent (#00bcd4) for shapes and linework, light grey (#e0e0e0) secondary",
  "avoid": "photography, photorealism, photographic texture, stock imagery, people, hands, offices, glowing brains, circuit-board cliches, robot faces, rounded bubbly playful shapes, marketing gloss, text, lettering, numerals, watermarks, white background, pale background, light background, bright full-bleed colour field",
  "reference_asset_keys": [],
  "kinds": {
    "content_hero": {
      "medium": "flat duotone editorial illustration",
      "mood": "bold simple silhouette of the subject, minimal detail, legible at small size, technical",
      "palette": "near-black ground (#121212), cyan (#00bcd4) flat shapes and linework, light grey (#e0e0e0) secondary accents only",
      "avoid": "photorealism, photographic texture, gradients, 3D rendering, drop shadows, text, lettering, numerals, logos, watermarks, busy detail, colour outside the palette, white background, pale background, light background, bright full-bleed colour field",
      "reference_asset_keys": []
    }
  }
}
$json$::jsonb,
       'manual-recovery', 'imagery-i3-b16.3',
       'B16.3 / bugs_open 027. Palette from design_intent.palette.reference_values (pinned). No reference anchors: existing heroes are SDXL photographic and hero.jpg does not serve.',
       true
  FROM sites s WHERE s.domain = 'gamesdesign.co.uk';

-- ── finetuning.uk ────────────────────────────────────────────────────────────
-- Voice: "looking at well-designed infrastructure — calm, competent, purposeful".
-- Its imagery_direction already carries a strong avoid-list (AI cliches, stock
-- people); that is folded into `avoid` here so it reaches the NEGATIVE prompt
-- instead of being prepended to the positive one, which is where it belongs.
INSERT INTO site_specs (site_id, aspect, data, source, created_by, notes, is_current)
SELECT s.id, 'imagery_style_guide', $json$
{
  "medium": "abstract geometric illustration of systems and infrastructure — network patterns, data-flow structures, calm architectural forms",
  "mood": "calm, competent, purposeful, precise, trustworthy",
  "palette": "deep navy and charcoal grounds, electric teal and cool blue accents, warm off-white for light elements only",
  "avoid": "stock photography of people, handshakes, offices, laptops, anyone pointing at a screen, robots, glowing brains, circuit boards, neural-network filigree, AI cliches, bright gradients, neon, flashiness, text, lettering, numerals, watermarks, white background, pale background, light background",
  "reference_asset_keys": [],
  "kinds": {
    "content_hero": {
      "medium": "flat duotone editorial illustration",
      "mood": "bold simple silhouette of the subject, minimal detail, legible at small size, calm and precise",
      "palette": "deep navy to charcoal ground, electric teal flat shapes and linework, cool blue secondary accents only",
      "avoid": "photorealism, photographic texture, gradients, 3D rendering, drop shadows, text, lettering, numerals, logos, watermarks, busy detail, robots, glowing brains, circuit boards, colour outside the palette, white background, pale background, light background, bright full-bleed colour field",
      "reference_asset_keys": []
    }
  }
}
$json$::jsonb,
       'manual-recovery', 'imagery-i3-b16.3',
       'B16.3 / bugs_open 027. No palette.reference_values on this site — colours derived from design_intent.colour_mood prose. No reference anchors: its `hero` asset sits on a pale grey ground and would import a white background.',
       true
  FROM sites s WHERE s.domain = 'finetuning.uk';

-- ── leopardessconsulting.co.uk ───────────────────────────────────────────────
-- Voice: its imagery_direction is the most specific of the three and needed no
-- invention — "diagrams: flat, gold on charcoal, drawn in the same hand as the
-- logo". NOTE the hard constraint carried into `avoid`: this site forbids image
-- models drawing CHARTS, because a model draws the appearance of a chart and
-- invents the numbers (real charts are code-rendered from a real series with a
-- source and a date). That is the same lesson as bugs_open/011.
INSERT INTO site_specs (site_id, aspect, data, source, created_by, notes, is_current)
SELECT s.id, 'imagery_style_guide', $json$
{
  "medium": "flat diagrammatic illustration, antique-gold linework on charcoal, drawn in the same hand as the logo",
  "mood": "explanatory, calm, precise, editorial, quietly authoritative",
  "palette": "charcoal to near-black ground (#0D0D0D), antique gold (#836E32) flat shapes and linework, warm off-white (#FAF8F4) sparingly",
  "avoid": "photographs of people, offices, handshakes, glowing brains, robot faces, neural-network filigree, photorealism, photographic texture, 3D rendering, gradients, drop shadows, charts, graphs, plotted axes or any figure that implies real data, text, lettering, numerals, watermarks, white background, pale background, bright full-bleed colour field",
  "reference_asset_keys": ["hero_home"],
  "kinds": {
    "content_hero": {
      "medium": "flat duotone editorial illustration, antique-gold linework on charcoal",
      "mood": "bold simple silhouette of the subject, minimal detail, generous negative space, legible at small size",
      "palette": "charcoal to near-black ground (#0D0D0D), antique gold (#836E32) flat shapes and linework only",
      "avoid": "photorealism, photographic texture, gradients, 3D rendering, drop shadows, charts, graphs, plotted axes or any figure that implies real data, text, lettering, numerals, logos, watermarks, busy detail, colour outside the palette, white background, pale background, light background, bright full-bleed colour field",
      "reference_asset_keys": ["hero_home"]
    }
  }
}
$json$::jsonb,
       'manual-recovery', 'imagery-i3-b16.3',
       'B16.3 / bugs_open 027. Palette from design_intent.palette.reference_values (pinned). Anchored to hero_home — inspected 2026-07-19: flat gold linework on near-black, no text, Banana-generated, matches the "same hand as the logo" direction.',
       true
  FROM sites s WHERE s.domain = 'leopardessconsulting.co.uk';

-- Verify: exactly one current guide per target, each carrying a content_hero
-- override, and the anchor list is what we intended.
SELECT s.domain,
       count(*)                                             AS current_rows,
       bool_and(sd.data->'kinds' ? 'content_hero')          AS has_content_hero_override,
       (array_agg(sd.data->'kinds'->'content_hero'->'reference_asset_keys'))[1] AS anchors
  FROM site_specs sd JOIN sites s ON s.id = sd.site_id
 WHERE sd.aspect = 'imagery_style_guide' AND sd.is_current = true
 GROUP BY s.domain ORDER BY s.domain;

COMMIT;
