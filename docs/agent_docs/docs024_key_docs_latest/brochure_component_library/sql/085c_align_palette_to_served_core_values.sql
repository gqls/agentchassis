-- 085c_align_palette_to_served_core_values.sql
--
-- The served styles.css was NOT rendered from the palette row. Proven by
-- re-rendering the layout template locally with the palette row's values and
-- diffing against the served file: every CORE slot differs by a shade
-- (background #080E1C vs #090F1A, surface #111E33 vs #132239, text #E4EAF2 vs
-- #E8EDF3, text_muted #7E91A8 vs #8A9BB0, border #1B2D47 vs #1E3050) and
-- line-height 1.65 vs 1.6, while every specialised slot and every structural
-- rule matched byte-for-byte.
--
-- Cause: `analyze_design` emits a fresh `color_scheme` on each run and the
-- composition rule is "spec wins for core slots" (render_css_composition_
-- helpers.go:corePaletteKeys). design_intent.palette.reference_values is meant
-- to hold that steady, but the prompt hands it over as "starting points, not
-- exact targets ... you may adjust them" — so the pin is advisory by
-- construction and the stylesheet drifted away from its own palette row.
--
-- This aligns the row to what the site actually serves, so a future
-- deterministic re-render reproduces the live file instead of shifting it
-- again. The values chosen are the SERVED ones: they are what the contrast
-- audit was run against, and they pass.

BEGIN;

UPDATE palettes SET colours = colours
      || '{"background":"#080E1C","surface":"#111E33","text":"#E4EAF2","text_muted":"#7E91A8","border":"#1B2D47"}'::jsonb,
    updated_at = NOW()
 WHERE id = 'c7c5435f-7cde-4cb7-9398-045bbb5be84a';

UPDATE css_themes SET color_palette = color_palette
      || '{"background":"#080E1C","surface":"#111E33","text":"#E4EAF2","text_muted":"#7E91A8","border":"#1B2D47"}'::jsonb
 WHERE id = 'b62650b3-d4df-4cb3-b586-b09b262dafa4';

UPDATE site_specs
   SET data = jsonb_set(data, '{palette,reference_values}',
        (data->'palette'->'reference_values')
        || '{"background":"#080E1C","surface":"#111E33","text":"#E4EAF2","text_muted":"#7E91A8","border":"#1B2D47"}'::jsonb,
        true),
       updated_at = NOW()
 WHERE site_id = (SELECT id FROM sites WHERE domain = 'fundamentallyai.com')
   AND aspect = 'design_intent' AND is_current;

COMMIT;
