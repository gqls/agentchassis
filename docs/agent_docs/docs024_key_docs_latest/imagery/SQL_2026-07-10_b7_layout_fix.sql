-- SQL_2026-07-10_b7_layout_fix.sql
--
-- RUNBOOK B7 resolution (user decision 2026-07-10: NO brochure fallback).
-- Root cause per the needs_new_layout_candidate spec: the layout matcher had
-- "fallback — no classification tags" and site_tags [] — robot-hands'
-- old-format classification carries no industry_tags, so the scheme-aware
-- matcher (resolve_composition_layout / resolveLayoutByTags) had nothing to
-- score. The library ALREADY has the right layout: tool-portal-dark
-- (scheme=dark, category=interactive, tags incl. interactive-platform, tools,
-- calculators, technical-reference, professional-dark) — grown from a prior
-- instance of exactly this gap.
--
-- Fix (three steps in one transaction):
--   1. Supersede classification with industry_tags matching the site's real
--      nature (interactive gripper-selection platform, calculators, dark
--      technical aesthetic).
--   2. Close the needs_new_layout_candidate item (wont_fix + note).
--   3. Queue a fresh needs_composition item (site-design-planner re-resolves
--      the composition; with tags + scheme=dark the weighted matcher should
--      now pick tool-portal-dark).

\set ON_ERROR_STOP on

BEGIN;

-- 1. Supersede classification, adding industry_tags (news_feed block and the
--    rest of the data carried over unchanged).
WITH old AS (
    UPDATE site_specs
    SET is_current = false, superseded_at = now()
    WHERE site_id = '00ff3af5-dad8-4770-9f70-3edc267a3c92'
      AND aspect = 'classification'
      AND is_current = true
    RETURNING site_id, data
)
INSERT INTO site_specs (site_id, aspect, data, source, notes, is_current, created_by)
SELECT
    site_id,
    'classification',
    data || jsonb_build_object(
        'industry_tags', jsonb_build_array(
            'interactive-platform', 'tools', 'tool-portal', 'calculators',
            'technical-reference', 'professional-dark', 'engineering',
            'developer-tools', 'utility-platform')
    ),
    'manual-recovery',
    'B7 layout fix 2026-07-10: added industry_tags so the scheme-aware layout matcher can score (was "fallback — no classification tags"). Tags aligned with tool-portal-dark. See RUNNING_NOTES Turn 18.',
    true,
    'imagery-best-in-class-i1'
FROM old;

-- 2. Close the layout-gap review item per the user decision.
UPDATE site_work_items
SET status = 'wont_fix',
    error = COALESCE(error || E'\n', '')
            || 'Resolved 2026-07-10: user rejected brochure fallback; classification industry_tags added and composition re-queued (expected match: tool-portal-dark).',
    updated_at = now()
WHERE site_id = '00ff3af5-dad8-4770-9f70-3edc267a3c92'
  AND item_type = 'needs_new_layout_candidate'
  AND status = 'needs_human_review';

-- 3. Queue the re-composition.
INSERT INTO site_work_items
    (site_id, source, item_type, severity, summary, spec, priority,
     handler_agent, status, triaged_at, created_by, item_key, pipeline)
VALUES
    ('00ff3af5-dad8-4770-9f70-3edc267a3c92', 'manual-recovery',
     'needs_composition', 'high',
     'Re-resolve composition after industry_tags fix — expect tool-portal-dark instead of brochure-formal',
     '{"reason": "b7_layout_fix_reresolve"}'::jsonb, 7,
     'site-design-planner', 'triaged', now(),
     'imagery-best-in-class-i1', 'needs_composition_b7_fix', 'build');

-- Verify
DO $verify$
DECLARE
    v_tags int;
    v_open int;
    v_item int;
BEGIN
    SELECT jsonb_array_length(data->'industry_tags') INTO v_tags
    FROM site_specs
    WHERE site_id = '00ff3af5-dad8-4770-9f70-3edc267a3c92'
      AND aspect = 'classification' AND is_current = true;
    IF COALESCE(v_tags, 0) < 5 THEN
        RAISE EXCEPTION 'industry_tags missing/short after update (got %)', v_tags;
    END IF;

    SELECT count(*) INTO v_open FROM site_work_items
    WHERE site_id = '00ff3af5-dad8-4770-9f70-3edc267a3c92'
      AND item_type = 'needs_new_layout_candidate' AND status = 'needs_human_review';
    IF v_open <> 0 THEN
        RAISE EXCEPTION 'layout-gap item still open';
    END IF;

    SELECT count(*) INTO v_item FROM site_work_items
    WHERE site_id = '00ff3af5-dad8-4770-9f70-3edc267a3c92'
      AND item_key = 'needs_composition_b7_fix' AND status = 'triaged';
    IF v_item <> 1 THEN
        RAISE EXCEPTION 'needs_composition item not queued';
    END IF;

    RAISE NOTICE 'B7 fix applied: tags added, review item closed, re-composition queued';
END
$verify$;

COMMIT;
