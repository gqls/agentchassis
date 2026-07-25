-- SQL_p3_design_intent_pin.sql — webdesign.co.uk, phase 3
--
-- WHEN THIS RUNS: after the domain-research-classifier leg has completed and
-- before webdesign-agent (needs_design) is ever released. That ordering is the
-- whole point. The classifier has its own write_design_intent_spec step, and
-- WriteSiteSpecAction (site_spec_actions.go:120-310) deep-merges NEW over OLD
-- with NO `pinned` guard — site_specs.pinned is honoured only by evidence_base
-- code. A pin written earlier would be partly overwritten by the classifier.
--
-- WHY A PIN AT ALL: webdesign-agent's analyze_design step renders only the
-- STRUCTURED design_intent.palette / .typography blocks into its prompt. Free
-- text (colour_mood, style_direction, avoid) is never rendered. A site whose
-- design_intent carries only prose therefore falls into the LLM's "no design
-- intent exists -> invent" branch and re-rolls its core colours on EVERY run —
-- which is how robot-hands.com got four CSS rewrites in one day, one of them
-- rolling a light background onto a dark site (see robot_hands/
-- SQL_2026-07-17_r1b_design_intent_palette_pin.sql, the pattern this copies).
--
-- WHAT THE CLASSIFIER ALREADY GOT RIGHT (read before writing, 2026-07-25): it
-- honoured the mission brief and wrote all eight canonical palette slots at
-- exactly the owner's values, plus a 12-entry avoid list. So this is NOT a
-- correction — it is a hardening. What it adds:
--   * palette.character and palette.guidance   (the prompt renders these; the
--     classifier wrote reference_values with no surrounding prose at all)
--   * typography: mono_font, base_size, line_height, character, guidance
--     (it wrote only font_family + heading_font)
--   * spacing block — radius 12px and container max, otherwise unstated
--   * dark_light: "light" as an explicit field, not merely implied by
--     style_direction: "modern-light"
--   * primary_hover / primary_text, which the source design uses
--
-- Note on slots: dropNonColourKeys in resolve_composition_reference_helpers.go
-- filters anything outside the canonical eight on the design_reference path.
-- primary_hover/primary_text survive here because extractPaletteSignal's
-- design_intent branch does not filter — but they only reach the CSS if the
-- layout template references them. They are recorded as intent either way.

\set ON_ERROR_STOP on

BEGIN;

WITH old AS (
    UPDATE site_specs
    SET is_current = false, superseded_at = now()
    WHERE site_id = (SELECT id FROM sites WHERE domain = 'webdesign.co.uk')
      AND aspect = 'design_intent'
      AND is_current = true
    RETURNING site_id, data
)
INSERT INTO site_specs (site_id, aspect, data, source, notes, is_current, created_by)
SELECT
    site_id,
    'design_intent',
    data || '{
      "dark_light": "light",
      "style_direction": "warm-minimalist-light",
      "palette": {
        "character": "Warm minimalist. A soft warm pearl field with pure white panels lifting off it, a muted sage primary and a soft terracotta accent. Calm, natural, considered — the feel of a well-made reference shelf rather than a software product. The site is LIGHT ONLY and must never acquire a dark mode.",
        "reference_values": {
          "primary": "#5c6b5d",
          "primary_hover": "#4a574b",
          "primary_text": "#ffffff",
          "secondary": "#8a9a86",
          "accent": "#d4a373",
          "background": "#f9f8f6",
          "surface": "#ffffff",
          "text": "#2b2b2b",
          "text_muted": "#717171",
          "border": "#edece9"
        },
        "guidance": "HARD REQUIREMENT: light warm scheme, no dark variant, ever. Background stays #f9f8f6 and must never go darker than #f0efec. Cards and panels are pure white. Text is #2b2b2b — never pure black. Every neutral carries a warm undertone; cool greys are wrong here. Shadows are warm and soft: rgba(43,43,43,0.04) / 0.06 / 0.08 at 8px / 16px / 32px blur. Border radius is 12px on cards, panels and interactive elements. Card hover is the only animation: a 4px lift, one shadow step up, and the border tinted toward the primary. Keep these reference values exactly unless the owner explicitly approves a change."
      },
      "typography": {
        "character": "Inter for everything readable; Fira Code strictly as a small monospace accent — the wordmark suffix, badges, value labels, code. Body text must never be monospace.",
        "reference_values": {
          "font_family": "''Inter'', system-ui, -apple-system, sans-serif",
          "heading_font": "''Inter'', system-ui, -apple-system, sans-serif",
          "mono_font": "''Fira Code'', ui-monospace, SFMono-Regular, monospace",
          "base_size": "16px",
          "line_height": "1.6"
        },
        "guidance": "Headings are Inter 700/800 with tight negative letter-spacing (-0.5px to -1.5px at display sizes). Mono accents are Fira Code at 0.75-0.9rem only. Base 16px, line-height 1.6. Both faces load from Google Fonts via the site head; do not substitute a different family."
      },
      "spacing": {
        "character": "Generous and quiet. Sections breathe; cards do the organising.",
        "reference_values": {
          "radius": "12px",
          "container_max_width": "1200px",
          "section_padding": "4rem",
          "card_padding": "1.75rem",
          "grid_gap": "1.5rem"
        },
        "guidance": "12px radius everywhere. 1200px container. 4rem between dashboard sections, 1.5rem between cards in a grid."
      }
    }'::jsonb,
    'manual-pin',
    'Phase 3 design pin, 2026-07-25. Hardens the classifier''s design_intent: adds palette.character/guidance, the full typography block (mono_font/base_size/line_height + prose), a spacing block and an explicit dark_light=light. The classifier had already written all 8 palette reference_values correctly from the mission brief — this is hardening, not correction. Lands after the classifier leg and before needs_design is released, because WriteSiteSpecAction deep-merges new over old with no pinned guard. See webdesign_couk/SQL_p3_design_intent_pin.sql.',
    true,
    'webdesign-couk-standup'
FROM old;

DO $verify$
DECLARE
    v_site   uuid;
    v_bg     text;
    v_mono   text;
    v_avoid  int;
    v_slots  int;
BEGIN
    SELECT id INTO v_site FROM sites WHERE domain = 'webdesign.co.uk';

    SELECT data->'palette'->'reference_values'->>'background',
           data->'typography'->'reference_values'->>'mono_font'
      INTO v_bg, v_mono
      FROM site_specs
     WHERE site_id = v_site AND aspect = 'design_intent' AND is_current = true;

    IF v_bg IS DISTINCT FROM '#f9f8f6' THEN
        RAISE EXCEPTION 'palette pin missing or wrong (background=%)', v_bg;
    END IF;
    IF v_mono IS NULL THEN
        RAISE EXCEPTION 'typography pin missing (mono_font is null)';
    END IF;

    -- The classifier''s own work must survive the merge. If these are gone the
    -- merge overwrote rather than extended, which is the failure this pattern
    -- exists to prevent.
    SELECT jsonb_array_length(data->'avoid') INTO v_avoid
      FROM site_specs
     WHERE site_id = v_site AND aspect = 'design_intent' AND is_current = true;
    IF COALESCE(v_avoid, 0) < 10 THEN
        RAISE EXCEPTION 'classifier avoid list lost or truncated (len=%)', v_avoid;
    END IF;

    SELECT count(*) INTO v_slots
      FROM site_specs,
           jsonb_object_keys(data->'palette'->'reference_values') k
     WHERE site_id = v_site AND aspect = 'design_intent' AND is_current = true;
    IF v_slots < 10 THEN
        RAISE EXCEPTION 'palette slot count too low (%)', v_slots;
    END IF;

    -- Exactly one current row for the aspect (idx_site_specs_current enforces
    -- this, but assert it so a future edit to this file cannot quietly break it).
    IF (SELECT count(*) FROM site_specs
         WHERE site_id = v_site AND aspect = 'design_intent' AND is_current) <> 1 THEN
        RAISE EXCEPTION 'design_intent is not single-current';
    END IF;

    RAISE NOTICE 'Phase 3 pin applied: warm-minimalist light palette + typography + spacing pinned for webdesign.co.uk';
END
$verify$;

COMMIT;
