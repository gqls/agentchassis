-- FILE: SQL_2026-08-14_kraft_palette.sql
--
-- webdesign.uk: the kraft-and-ink palette (owner approved the proposed values,
-- 2026-08-14). Four colour copies exist and the landmine says a theme change
-- must rewrite them all or component regens keep baking the old colours:
--   1. design_intent.palette (site_specs)          — the record + the ADVISORY
--      anchor the analyze_design prompt renders (reference_values are
--      "starting points", NOT a pin — the corrected landmine).
--   2. sites.content_data.color_scheme             — the last design pass's spec.
--   3. palettes.colours (43feb06e…)                 — the theme's palette row.
--   4. style_collections.color_palette (9c78c279…) — read by component renders.
-- All four are single-site (measured: 1 site / 1 collection / 1 theme), so no
-- shared seam. Green (#2d5016 / #C84B1F rust) leaves everywhere.
-- `secondary` differs by copy on purpose: design_intent uses it as a second
-- SURFACE (#efe6d3); the scheme/palette/collection copies use it as a dark
-- companion tone (#4a4437) — matching each copy's existing semantics.

BEGIN;

UPDATE site_specs SET is_current=false, superseded_at=now()
 WHERE site_id='1fcfa4f3-ec80-4010-878b-b971cd46711f' AND aspect='design_intent' AND is_current=true;

INSERT INTO site_specs (site_id, aspect, data, source, notes, is_current, created_by, pinned)
SELECT '1fcfa4f3-ec80-4010-878b-b971cd46711f', 'design_intent', $json$
{
  "avoid": [
    "Gradient backgrounds or gradient overlays on text",
    "Animated hero sections, particle effects, scroll-jacking",
    "Browser-frame mockups, device frames, angle-shot laptops",
    "Wireframe sketch illustrations or colour-swatch imagery",
    "Bold colour blocks used purely decoratively",
    "Cards with drop shadows and rounded corners as a layout default",
    "Icon libraries used to add visual noise to plain text sections",
    "Any element that looks like it came from a SaaS product marketing template"
  ],
  "palette": {
    "character": "Kraft paper and ink with one gold. The pages are the paper the site's pen-and-ink illustrations are drawn on; the only colour anywhere is the golden-egg gold, which the imagery reserves for the finished product. Warm, matt, printed — never glossy, never corporate blue-grey.",
    "reference_values": {
      "background": "#f7f1e6",
      "surface": "#efe6d3",
      "secondary": "#efe6d3",
      "border": "#d8c3a0",
      "text": "#1a1a1a",
      "primary": "#1a1a1a",
      "text_muted": "#6f6553",
      "accent": "#8a6410"
    },
    "guidance": "Hold these values closely: they are matched to the site's generated hero artwork (kraft ground #d8c3a0–#c8ab82, near-black ink, gold #c8961e on the output element). Accent #8a6410 is a bronze dark enough to pass WCAG AA as link text on the paper background; do not lighten it toward the egg-gold for text. Buttons and primary CTAs: fill #c8961e (the egg gold) with near-black #1a1a1a text. No green anywhere — the forest accent belonged to the retired brand. No pure white and no cool greys."
  },
  "typography": {
    "reference_values": {
      "font_family": "'Inter', system-ui, -apple-system, sans-serif",
      "heading_font": "'Playfair Display', Georgia, serif"
    }
  },
  "colour_mood": "Kraft paper, black ink, one gold. The whole site reads as a well-made drawing on cheap paper: warm paper ground, near-black text, and the golden-egg gold reserved for buttons and the things the customer gets. Deliberately matt and printed. The old near-white/forest-green document palette is retired with the £1,200 offer.",
  "style_direction": "modern-light",
  "typography_mood": "A neutral grotesque for body copy — the kind that disappears and lets the words do the work. A slightly characterful serif or semi-display face for headings only — enough to signal that someone made deliberate choices, not enough to draw attention to itself. The pairing should feel like a well-designed British print publication from the early 2000s: confident, unhurried, professional without being corporate.",
  "imagery_direction": "The site's imagery is its pen-and-ink illustration set (imagery_style_guide): Heath Robinson contraptions, the marble run, the cardboard box, the trade counter, the scruffy goose with the golden egg — an unorthodox process producing an immaculate result, on kraft ground with gold only on the output. Page colours exist to sit these drawings on their own paper. No stock photography, no browser mockups, no device frames.",
  "layout_preference": "Single-column reading flow for all prose pages. Generous line length (65–75 characters). Ample vertical whitespace — padding does the work that decoration usually does. Navigation is minimal: wordmark left, five page links right, nothing else. Each page opens with its kraft-and-ink hero illustration behind a clear headline; below that, single-column reading flow."
}
$json$::jsonb,
 'owner_ruling',
 'Kraft-and-ink palette for the scrappy £149 brand (owner approved 2026-08-14). '
 'Replaces the "well-printed document" near-white/forest palette. Also retires '
 'the stale "No hero image" layout line — every page now opens with its '
 'generated kraft hero.',
 true, 'ai-site-selling-automation-2026-08-14', COALESCE(pinned,false)
FROM site_specs
WHERE site_id='1fcfa4f3-ec80-4010-878b-b971cd46711f' AND aspect='design_intent' AND is_current=false
ORDER BY superseded_at DESC NULLS LAST LIMIT 1;

UPDATE sites SET content_data = jsonb_set(content_data, '{color_scheme}', '{
  "text":"#1a1a1a","accent":"#8a6410","border":"#d8c3a0","primary":"#1a1a1a",
  "surface":"#efe6d3","secondary":"#4a4437","background":"#f7f1e6","text_muted":"#6f6553"}'::jsonb)
 WHERE id='1fcfa4f3-ec80-4010-878b-b971cd46711f';

UPDATE palettes SET colours='{
  "text":"#1a1a1a","accent":"#8a6410","border":"#d8c3a0","primary":"#1a1a1a",
  "surface":"#efe6d3","secondary":"#4a4437","background":"#f7f1e6","text_muted":"#6f6553"}'::jsonb,
 updated_at=now() WHERE id='43feb06e-009e-4404-97ef-c06dcdc58e72';

UPDATE style_collections SET color_palette='{
  "text":"#1a1a1a","accent":"#8a6410","border":"#d8c3a0","primary":"#1a1a1a",
  "surface":"#efe6d3","secondary":"#4a4437","background":"#f7f1e6","text_muted":"#6f6553"}'::jsonb,
 updated_at=now() WHERE id='9c78c279-665d-450d-9ba5-fca42c5a8653';

DO $$
DECLARE n int; di jsonb;
BEGIN
  SELECT count(*) INTO n FROM site_specs
   WHERE site_id='1fcfa4f3-ec80-4010-878b-b971cd46711f' AND aspect='design_intent' AND is_current;
  IF n<>1 THEN RAISE EXCEPTION 'expected 1 current design_intent, got %', n; END IF;

  SELECT data INTO di FROM site_specs
   WHERE site_id='1fcfa4f3-ec80-4010-878b-b971cd46711f' AND aspect='design_intent' AND is_current;
  IF di->'palette'->'reference_values'->>'background' <> '#f7f1e6' THEN
    RAISE EXCEPTION 'design_intent background not kraft paper'; END IF;
  IF di->>'layout_preference' LIKE '%No hero image%' THEN
    RAISE EXCEPTION 'the stale no-hero-image line survived'; END IF;

  -- all four copies must agree on the accent, or a component regen bakes green back
  SELECT count(*) INTO n FROM (
    SELECT content_data->'color_scheme'->>'accent' AS a FROM sites WHERE id='1fcfa4f3-ec80-4010-878b-b971cd46711f'
    UNION SELECT colours->>'accent' FROM palettes WHERE id='43feb06e-009e-4404-97ef-c06dcdc58e72'
    UNION SELECT color_palette->>'accent' FROM style_collections WHERE id='9c78c279-665d-450d-9ba5-fca42c5a8653'
    UNION SELECT data->'palette'->'reference_values'->>'accent' FROM site_specs
      WHERE site_id='1fcfa4f3-ec80-4010-878b-b971cd46711f' AND aspect='design_intent' AND is_current) x
  WHERE x.a='#8a6410';
  IF n<>1 THEN RAISE EXCEPTION 'the four accent copies do not agree'; END IF;
END $$;

COMMIT;
