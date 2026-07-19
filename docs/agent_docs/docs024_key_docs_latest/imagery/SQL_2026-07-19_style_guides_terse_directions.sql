-- B16.3 follow-up — make the ACCENT COLOUR survive the 200-char direction cap.
--
-- WHY (pilot evidence, 2026-07-19, gamesdesign.co.uk):
-- The two pilot content heroes came back with the right GROUND but an invented
-- ACCENT (orange/navy on one, a teal field on the other). The stored prompt says
-- why — `assets.origin_prompt` for content_hero_tool_ehp_calculator ends:
--
--   "... colour palette: near-black ground (#121212). Header image for a web-based
--    tool representing: Effective Health (EHP) Calculator ..."
--
-- The cyan is simply GONE. Mechanism, both parts verified in code:
--   * composeDirection (imagery_style_guide.go:136) joins medium -> mood ->
--     "colour palette: "+palette, so the palette is ALWAYS LAST; and
--   * composeImagePromptWithDirection truncates the whole direction at
--     maxImageryDirectionInPrompt = 200 (generate_image_actions.go:1037).
-- So a verbose medium+mood silently eats the colour instruction — the single most
-- brand-identifying part of the guide — and the model picks its own accent.
--
-- This is why robot-hands.com looked fine and gamesdesign did not: robot-hands'
-- shorter mood leaves its "electric blue (#0080FF)" just inside the 200 chars.
-- The same config shape works or fails on whether the accent lands before char 200.
-- That is a platform defect, not a content one — recorded in /bugs_open/027 §5(c);
-- the durable fix is to compose the palette FIRST (or raise/omit the cap for
-- structured guides), which is code and wants the council gate.
--
-- THIS migration is the config-side mitigation only: keep medium+mood terse and
-- put the ACCENT FIRST inside the palette string, so the accent survives even if
-- the tail is cut. Live immediately; no image roll.
--
-- Applied: 2026-07-19.

\set ON_ERROR_STOP on
BEGIN;

CREATE TABLE IF NOT EXISTS site_specs_imagery_guide_backup_20260719b AS
SELECT * FROM site_specs WHERE aspect = 'imagery_style_guide';

-- gamesdesign.co.uk — content_hero direction, terse, accent first.
-- Composed length: "flat duotone editorial illustration. bold simple silhouette,
-- minimal detail. colour palette: cyan #00bcd4 on near-black #121212" ≈ 150 chars.
UPDATE site_specs SET
  data = jsonb_set(data, '{kinds,content_hero}', $j$
  {
    "medium": "flat duotone editorial illustration",
    "mood": "bold simple silhouette, minimal detail",
    "palette": "cyan #00bcd4 on near-black #121212, light grey accents",
    "avoid": "photorealism, photographic texture, gradients, 3D rendering, drop shadows, text, lettering, words, numerals, labels, captions, logos, watermarks, busy detail, colour outside the palette, white background, pale background, bright full-bleed colour field",
    "reference_asset_keys": []
  }
  $j$::jsonb),
  updated_at = now()
 WHERE aspect='imagery_style_guide' AND is_current=true
   AND site_id = (SELECT id FROM sites WHERE domain='gamesdesign.co.uk');

-- finetuning.uk
UPDATE site_specs SET
  data = jsonb_set(data, '{kinds,content_hero}', $j$
  {
    "medium": "flat duotone editorial illustration",
    "mood": "bold simple silhouette, minimal detail",
    "palette": "electric teal on deep navy #1B2A41, cool blue accents",
    "avoid": "photorealism, photographic texture, gradients, 3D rendering, drop shadows, text, lettering, words, numerals, labels, captions, logos, watermarks, busy detail, robots, glowing brains, circuit boards, colour outside the palette, white background, pale background, bright full-bleed colour field",
    "reference_asset_keys": []
  }
  $j$::jsonb),
  updated_at = now()
 WHERE aspect='imagery_style_guide' AND is_current=true
   AND site_id = (SELECT id FROM sites WHERE domain='finetuning.uk');

-- leopardessconsulting.co.uk — keeps its anchor; the anchor carries the style
-- that the truncated text cannot.
UPDATE site_specs SET
  data = jsonb_set(data, '{kinds,content_hero}', $j$
  {
    "medium": "flat diagrammatic illustration",
    "mood": "bold simple silhouette, generous negative space",
    "palette": "antique gold #836E32 on near-black #0D0D0D",
    "avoid": "photorealism, photographic texture, gradients, 3D rendering, drop shadows, charts, graphs, plotted axes or any figure implying real data, text, lettering, words, numerals, labels, captions, logos, watermarks, busy detail, colour outside the palette, white background, pale background, bright full-bleed colour field",
    "reference_asset_keys": ["hero_home"]
  }
  $j$::jsonb),
  updated_at = now()
 WHERE aspect='imagery_style_guide' AND is_current=true
   AND site_id = (SELECT id FROM sites WHERE domain='leopardessconsulting.co.uk');

-- Verify: the composed direction each site will now send, and its length against
-- the 200-char cap. Anything at or over 200 will still lose its palette tail.
SELECT s.domain,
       length(
         (sd.data->'kinds'->'content_hero'->>'medium') || '. ' ||
         (sd.data->'kinds'->'content_hero'->>'mood')   || '. colour palette: ' ||
         (sd.data->'kinds'->'content_hero'->>'palette')
       ) AS composed_len,
       left((sd.data->'kinds'->'content_hero'->>'medium') || '. ' ||
            (sd.data->'kinds'->'content_hero'->>'mood')   || '. colour palette: ' ||
            (sd.data->'kinds'->'content_hero'->>'palette'), 200) AS what_the_model_sees
  FROM site_specs sd JOIN sites s ON s.id=sd.site_id
 WHERE sd.aspect='imagery_style_guide' AND sd.is_current=true
 ORDER BY s.domain;

COMMIT;
