-- SQL_2026-07-17_r1b_design_intent_palette_pin.sql
--
-- R1 follow-up (the webdesign churn leg). Mechanism found 2026-07-17:
-- webdesign-agent's analyze_design LLM step invents a fresh color_scheme
-- EVERY run, and render_css_from_spec's merge lets the spec WIN on core
-- slots (primary/secondary/accent/background/surface/text/text_muted/
-- border). The prompt template only surfaces the STRUCTURED
-- design_intent.palette block (character / reference_values / guidance);
-- robot-hands' design_intent had only free-text colour_mood ("deep
-- charcoal...near-black"), which the template never renders — so the LLM
-- fell into its "no design intent exists → invent" branch on every run.
-- Result: core colours hue-drifted across today's four CSS renders and the
-- 20:31 run rolled a LIGHT background (#F4F5F7) onto the dark site — live.
-- This is also the same class as the recurring generic_theme misfire.
--
-- Fix: supersede design_intent keeping every existing field, ADDING the
-- structured palette + typography blocks with the user-approved
-- tool-portal-dark values (same set written to palettes.colours by
-- SQL_2026-07-17_r1_dark_theme_restore.sql) as reference_values, plus
-- hard guidance that the dark scheme is non-negotiable. After this,
-- re-trigger webdesign-agent and verify the committed styles.css is dark.

\set ON_ERROR_STOP on

BEGIN;

WITH old AS (
    UPDATE site_specs
    SET is_current = false, superseded_at = now()
    WHERE site_id = '00ff3af5-dad8-4770-9f70-3edc267a3c92'
      AND aspect = 'design_intent'
      AND is_current = true
    RETURNING site_id, data
)
INSERT INTO site_specs (site_id, aspect, data, source, notes, is_current, created_by)
SELECT
    site_id,
    'design_intent',
    data || '{
      "palette": {
        "character": "Dark-first engineering dashboard. Near-black blue-tinted backgrounds, high-contrast light text, restrained accents. The site ships the tool-portal-dark layout (scheme=dark) — the palette must never drift light.",
        "reference_values": {
          "primary": "#1A1F2E",
          "primary_hover": "#2563eb",
          "primary_text": "#ffffff",
          "secondary": "#C8D8E8",
          "accent": "#E8500A",
          "background": "#0F1218",
          "surface": "#1E2535",
          "surface_alt": "#1a1a1a",
          "text": "#E2E8F0",
          "text_muted": "#7A8FA6",
          "border": "#2D3A4A"
        },
        "guidance": "HARD REQUIREMENT: dark scheme. background must remain near-black (luminance at or below #14161C); text must remain light (contrast >= 7:1 on background). Never output a light or pastel background — the user rejected the light brochure look twice (B7 2026-07-10, D13 gate 2026-07-17). Adjust hues only within the dark scheme; keep the reference values unless a change is clearly justified by the brand character."
      },
      "typography": {
        "character": "Technical, systematic. Body in a legible humanist sans; headings/mono accents from the IBM Plex family. Body text must NEVER be monospace.",
        "reference_values": {
          "font_family": "IBM Plex Sans, Inter, system-ui, sans-serif",
          "heading_font": "IBM Plex Mono, JetBrains Mono, monospace",
          "base_size": "16px",
          "line_height": "1.65"
        },
        "guidance": "Keep base_size at 16px and body font sans-serif. Monospace is for headings, spec values and code only."
      }
    }'::jsonb,
    'manual-recovery',
    'R1b palette pin 2026-07-17: added structured design_intent.palette/typography (the only blocks the analyze_design prompt renders) so webdesign-agent stops re-inventing core colours each run; 20:31 run had rolled a light background onto the dark site. Values = user-approved tool-portal-dark set, same as palettes.colours. See robot_hands/SQL_2026-07-17_r1b_design_intent_palette_pin.sql.',
    true,
    'robot-hands-r1'
FROM old;

DO $verify$
DECLARE v_bg text; v_cnt int;
BEGIN
    SELECT data->'palette'->'reference_values'->>'background' INTO v_bg
    FROM site_specs
    WHERE site_id = '00ff3af5-dad8-4770-9f70-3edc267a3c92'
      AND aspect = 'design_intent' AND is_current = true;
    IF v_bg <> '#0F1218' THEN RAISE EXCEPTION 'palette pin missing (bg=%)', v_bg; END IF;

    SELECT count(*) INTO v_cnt FROM site_specs
    WHERE site_id = '00ff3af5-dad8-4770-9f70-3edc267a3c92'
      AND aspect = 'design_intent' AND is_current = true
      AND data ? 'colour_mood' AND data ? 'avoid';
    IF v_cnt <> 1 THEN RAISE EXCEPTION 'existing design_intent fields lost'; END IF;

    RAISE NOTICE 'R1b applied: design_intent palette/typography pinned dark';
END
$verify$;

COMMIT;
