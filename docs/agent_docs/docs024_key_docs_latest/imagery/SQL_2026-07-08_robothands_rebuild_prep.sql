-- SQL_2026-07-08_robothands_rebuild_prep.sql
--
-- Imagery best-in-class workstream, Phase I0: rebuild robot-hands.com from
-- scratch with news scope. See PLAN_imagery_best_in_class.md and
-- RUNNING_NOTES_imagery_best_in_class.md (Turn 3).
--
-- Steps:
--   0. Backup all site_specs rows for the site (doc 009 ritual).
--   1. Supersede adoption-residue aspects (design, structure).
--   2. Supersede classification with a news-enabled copy: adds
--      content_features.news_feed (read by planner RULE 11 at plan time and
--      by the check_news_feed discovery family afterwards); drops the
--      deprecated recommended_builder=pageflow-builder key.
--   3. Retire five stale pre-2G unfulfilled_hero_variant items (triaged
--      2026-05-08) so they cannot dispatch mid-rebuild.
--   4. Insert the needs_site_plan trigger item (same shape as the
--      build-briefing-agent emissions for dartsonline/vonc).
--
-- Idempotence: NOT idempotent (INSERTs); the verify block aborts the
-- transaction on any unexpected state. Backup table is create-if-not-exists.

\set ON_ERROR_STOP on

-- ── 0. Backup (outside transaction) ──

CREATE TABLE IF NOT EXISTS site_specs_backup_20260708_robothands_rebuild AS
SELECT * FROM site_specs WHERE site_id = '00ff3af5-dad8-4770-9f70-3edc267a3c92';

SELECT
    (SELECT count(*) FROM site_specs
     WHERE site_id = '00ff3af5-dad8-4770-9f70-3edc267a3c92') AS live_rows,
    (SELECT count(*) FROM site_specs_backup_20260708_robothands_rebuild) AS backup_rows;

BEGIN;

-- ── 1. Supersede adoption-residue aspects ──

UPDATE site_specs
SET is_current = false, superseded_at = now()
WHERE site_id = '00ff3af5-dad8-4770-9f70-3edc267a3c92'
  AND aspect IN ('design', 'structure')
  AND is_current = true;

-- ── 2. Supersede classification with news-enabled copy ──

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
    (data - 'recommended_builder') || jsonb_build_object(
        'content_features', jsonb_build_object(
            'news_feed', jsonb_build_object(
                'recommended', true,
                'source_types', jsonb_build_array('rss', 'news_search', 'api_news'),
                'separate_page', true,
                'vertical_keywords', jsonb_build_array(
                    'industrial robot grippers', 'robotic end effectors',
                    'end-of-arm tooling', 'factory automation',
                    'collaborative robots', 'pneumatic grippers',
                    'electric grippers', 'machine tending'),
                'reason', 'News scope added at 2026-07-08 from-scratch rebuild: robotics/automation news for engineers selecting grippers (imagery best-in-class workstream).'
            )
        )
    ),
    'manual-recovery',
    'Rebuild 2026-07-08: added content_features.news_feed (planner RULE 11 + news discovery checks); dropped deprecated recommended_builder (pageflow-builder retired).',
    true,
    'imagery-best-in-class-rebuild'
FROM old;

-- ── 3. Retire stale pre-2G imagery items ──

UPDATE site_work_items
SET status = 'wont_fix',
    error = COALESCE(error || E'\n', '')
            || 'Superseded by 2026-07-08 from-scratch rebuild (imagery best-in-class workstream)',
    updated_at = now()
WHERE site_id = '00ff3af5-dad8-4770-9f70-3edc267a3c92'
  AND item_type = 'unfulfilled_hero_variant'
  AND status = 'triaged';

-- ── 4. Insert the rebuild trigger ──

INSERT INTO site_work_items
    (site_id, source, item_type, severity, summary, spec, priority,
     handler_agent, status, created_by, item_key, pipeline)
VALUES
    ('00ff3af5-dad8-4770-9f70-3edc267a3c92', 'manual-recovery',
     'needs_site_plan', 'high',
     'Rebuild robot-hands.com from scratch with news scope (imagery best-in-class workstream, Phase I0)',
     '{}'::jsonb, 15,
     'build-site-planner', 'detected', 'imagery-best-in-class-rebuild',
     'site_plan_robot-hands.com', 'build');

-- ── Verify ──

DO $verify$
DECLARE
    v_news text;
    v_residue int;
    v_item int;
BEGIN
    SELECT data #>> '{content_features,news_feed,recommended}' INTO v_news
    FROM site_specs
    WHERE site_id = '00ff3af5-dad8-4770-9f70-3edc267a3c92'
      AND aspect = 'classification' AND is_current = true;
    IF v_news IS DISTINCT FROM 'true' THEN
        RAISE EXCEPTION 'news_feed.recommended not true after update (got %)', v_news;
    END IF;

    SELECT count(*) INTO v_residue FROM site_specs
    WHERE site_id = '00ff3af5-dad8-4770-9f70-3edc267a3c92'
      AND aspect IN ('design', 'structure') AND is_current = true;
    IF v_residue <> 0 THEN
        RAISE EXCEPTION 'adoption-residue aspects still current: %', v_residue;
    END IF;

    SELECT count(*) INTO v_item FROM site_work_items
    WHERE site_id = '00ff3af5-dad8-4770-9f70-3edc267a3c92'
      AND item_type = 'needs_site_plan' AND status = 'detected';
    IF v_item <> 1 THEN
        RAISE EXCEPTION 'expected 1 detected needs_site_plan item, got %', v_item;
    END IF;

    RAISE NOTICE 'rebuild prep OK: news flag set, residue superseded, stale items retired, trigger inserted';
END
$verify$;

COMMIT;
