-- SQL_2026-07-10_robothands_imagery_style_guide.sql
--
-- Phase I1: seed the imagery_style_guide site_specs aspect for
-- robot-hands.com — the structured, prompt-ready distillation of its
-- design_intent (imagery_direction + colour_mood). Read by generate_image
-- (imagery_style_guide.go) once the I1 chassis code deploys: photographic
-- kinds get medium+mood+palette, icons palette only, logos nothing; `avoid`
-- goes to the negative prompt; reference_asset_keys anchor Banana
-- generations to the approved brand heroes (resolved to stable s3:// URIs).
--
-- Guarded insert (aborts if a current row exists — merge instead) per the
-- site_specs supersede-row convention.

\set ON_ERROR_STOP on

BEGIN;

DO $guard$
BEGIN
    IF EXISTS (
        SELECT 1 FROM site_specs
        WHERE site_id = '00ff3af5-dad8-4770-9f70-3edc267a3c92'
          AND aspect = 'imagery_style_guide' AND is_current = true
    ) THEN
        RAISE EXCEPTION 'imagery_style_guide already exists — supersede instead of insert';
    END IF;
END
$guard$;

INSERT INTO site_specs (site_id, aspect, data, source, notes, is_current, created_by)
VALUES (
    '00ff3af5-dad8-4770-9f70-3edc267a3c92',
    'imagery_style_guide',
    jsonb_build_object(
        'palette', 'deep charcoal and near-black backgrounds, electric blue accents (#0080FF to #00B4FF), light grey highlights, occasional amber (#FFB800) callouts',
        'medium', 'industrial photography of real robotic grippers and end-effectors in manufacturing environments, close-up machined-surface detail, dark atmospheric lighting with focused highlights',
        'mood', 'precise, technical, engineered, high-contrast, serious',
        'avoid', 'stock photos of people pointing at screens, generic technology abstractions, decorative colour, cartoonish rendering, watermarks, text overlays',
        'reference_asset_keys', jsonb_build_array('hero_canonical', 'hero_home')
    ),
    'manual-recovery',
    'Phase I1 brand consistency layer: distilled from design_intent (imagery_direction + colour_mood). See PLAN_imagery_best_in_class.md.',
    true,
    'imagery-best-in-class-i1'
);

DO $verify$
DECLARE
    v_palette text;
BEGIN
    SELECT data->>'palette' INTO v_palette
    FROM site_specs
    WHERE site_id = '00ff3af5-dad8-4770-9f70-3edc267a3c92'
      AND aspect = 'imagery_style_guide' AND is_current = true;
    IF v_palette IS NULL OR length(v_palette) < 20 THEN
        RAISE EXCEPTION 'imagery_style_guide palette missing after insert';
    END IF;
    RAISE NOTICE 'imagery_style_guide seeded for robot-hands.com';
END
$verify$;

COMMIT;
