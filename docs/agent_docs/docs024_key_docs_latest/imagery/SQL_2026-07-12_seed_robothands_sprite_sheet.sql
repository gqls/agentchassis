-- SQL_2026-07-12_seed_robothands_sprite_sheet.sql
--
-- Phase I2.1: hand-seed robot-hands.com's site-scope sprite_sheet plan row
-- (the planner prompt learns to emit these later; the scope doc allows
-- hand-seeding the testbed). Grid geometry + cell vocabulary locked with the
-- user (Turn 28): 3×3 @ 768² (256px cells), technical glyphs.
--
-- The prompt follows every hard-won rule: ONE image (a single grid), one
-- glyph per cell in reading order, single flat colour, flat selectable
-- background (NO transparency — abandoned 2026-05-20), no photorealism.
-- style_hints carries the grid plan for the sprite-CSS emit step; the
-- REQUESTED cell_names are recorded here, but the TRUE cell→name map is
-- assigned at the human eyeball gate after generation and written back.
--
-- Guarded: aborts if a sprite_sheet row already exists on the current plan.

\set ON_ERROR_STOP on

BEGIN;

DO $guard$
BEGIN
    IF EXISTS (
        SELECT 1 FROM site_plan_imagery spi
        JOIN site_plans sp ON sp.id = spi.plan_id AND sp.is_current
        WHERE sp.site_id = '00ff3af5-dad8-4770-9f70-3edc267a3c92'
          AND spi.kind = 'sprite_sheet'
    ) THEN
        RAISE EXCEPTION 'sprite_sheet row already exists on the current plan';
    END IF;
END
$guard$;

INSERT INTO site_plan_imagery (plan_id, scope, scope_ref, key, kind, prompt, style_hints, ordering, source)
SELECT sp.id, 'site', NULL, 'sprite_sheet_main', 'sprite_sheet',
       'A single image: a precise 3x3 grid of nine small flat line icons, one icon '
    || 'centred in each equal cell, reading order left-to-right top-to-bottom: '
    || 'a checkmark, a pressure gauge, a robotic gripper jaw, a cog, a bar chart, '
    || 'a download arrow into a tray, a right arrow, an info circle, a warning triangle. '
    || 'Uniform minimalist engineering line style, consistent stroke weight, single '
    || 'light grey colour (#C8CDD4) on a flat dark charcoal background (#1a1a2e), '
    || 'thin visible cell divider lines, no shadows, no gradients, no photorealism, no text.',
       jsonb_build_object(
           'rows', 3, 'cols', 3,
           'aspect_ratio', '1:1',
           'cell_names', jsonb_build_array('check','gauge','gripper','cog','chart','download','arrow','info','warning'),
           'style', 'flat light-grey line glyphs on #1a1a2e',
           'cell_names_verified', false),
       50, 'manual'
FROM site_plans sp
WHERE sp.site_id = '00ff3af5-dad8-4770-9f70-3edc267a3c92' AND sp.is_current
RETURNING key, kind;

DO $verify$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM site_plan_imagery spi
        JOIN site_plans sp ON sp.id = spi.plan_id AND sp.is_current
        WHERE sp.site_id = '00ff3af5-dad8-4770-9f70-3edc267a3c92'
          AND spi.key = 'sprite_sheet_main' AND spi.kind = 'sprite_sheet'
    ) THEN
        RAISE EXCEPTION 'sprite_sheet_main row not present after insert';
    END IF;
    RAISE NOTICE 'sprite_sheet_main seeded on the current robot-hands plan';
END
$verify$;

COMMIT;
