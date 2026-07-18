-- SQL_2026-07-17_r6_retire_dead_article_rows.sql
--
-- R6 of HANDOFF_2026-07-17_robot_hands_site_fixes.md: 6 of the 9 blog-post
-- rows are status='active' but never really shipped. Decision per row
-- (delegated to this session by the handoff):
--   RETIRE all 6 — none carries content worth building:
--   * 3 plan-era scaffolds (placeholder titles): news-post,
--     learning-center-article, learning-center-post
--   * 3 never-built duplicate slugs of the REAL deployed guides:
--     grip-force-friction-calculator-guide, gripper-cycle-time-estimator-guide,
--     gripper-payload-calculator-guide (the real ones live at
--     /blog/tool-gripper-payload-calculator-guide.html and
--     /guides/tool-{grip-force-friction-calculator,gripper-cycle-time-estimator}-guide.html)
--
-- The LISTING side is already fixed in code by the imagery session (F2.1
-- listedOnly, commit 4e35c8064, verified live in the running pod
-- 2026-07-17: blog_posts demands deployed_at + non-empty sections).
-- Archiving the rows kills the remaining churn: imagery sweeps, "re-render
-- after asset landed" items, incomplete_page_group noise.
--
-- Per the handoff's caution, the card/hero assets generated for these rows
-- this week (2026-07-16/17 imagery passes) are superseded in the same
-- transaction so they don't sit as active orphans.
--
-- Plan rows (site_plan_pages learning-center-article/learning-center-post,
-- role blog-post) are left untouched — plan surgery belongs to the R3
-- learning-center IA decision; F2.1 makes them harmless to listings.

\set ON_ERROR_STOP on

CREATE TABLE IF NOT EXISTS pages_backup_20260717_r6 AS
SELECT * FROM pages
WHERE site_id = '00ff3af5-dad8-4770-9f70-3edc267a3c92'
  AND name IN ('news-post','learning-center-article','learning-center-post',
               'grip-force-friction-calculator-guide',
               'gripper-cycle-time-estimator-guide',
               'gripper-payload-calculator-guide');
SELECT count(*) AS bak_pages FROM pages_backup_20260717_r6;

BEGIN;

-- 1. Archive the six page rows (house convention: pages.status='archived').
UPDATE pages
SET status = 'archived', updated_at = now()
WHERE site_id = '00ff3af5-dad8-4770-9f70-3edc267a3c92'
  AND name IN ('news-post','learning-center-article','learning-center-post',
               'grip-force-friction-calculator-guide',
               'gripper-cycle-time-estimator-guide',
               'gripper-payload-calculator-guide')
  AND status = 'active';

-- 2. Supersede their card + hero assets (active, generated 2026-07-16/17).
--    The real tool-* guides' assets (card_tool_*, content_hero_tool_*) are
--    NOT touched.
UPDATE assets
SET status = 'superseded'
WHERE site_id = '00ff3af5-dad8-4770-9f70-3edc267a3c92'
  AND status = 'active'
  AND asset_key IN (
    'card_news_post','card_learning_center_article','card_learning_center_post',
    'card_grip_force_friction_calculator_guide',
    'card_gripper_cycle_time_estimator_guide',
    'card_gripper_payload_calculator_guide',
    'content_hero_news_post','content_hero_learning_center_article',
    'content_hero_learning_center_post',
    'content_hero_grip_force_friction_calculator_guide',
    'content_hero_gripper_cycle_time_estimator_guide',
    'content_hero_gripper_payload_calculator_guide');

-- 3. Close the five open "re-render after its image asset landed" items that
--    reference retired rows (they are needs_human_review; this is the review
--    outcome).
UPDATE site_work_items
SET status = 'wont_fix', updated_at = now(),
    error = COALESCE(error || E'\n', '')
            || 'Closed 2026-07-17 (R6): page row retired (scaffold/duplicate, never deployed); assets superseded. See robot_hands/SQL_2026-07-17_r6_retire_dead_article_rows.sql.'
WHERE id IN ('7695c973-350c-4345-9f69-672aa76182ca',   -- grip-force-friction-calculator-guide
             '4d6896c2-e108-4866-a118-fe8e2de98bb4',   -- gripper-cycle-time-estimator-guide
             '71d0a15d-da34-4cf6-8840-2006463965a5',   -- gripper-payload-calculator-guide
             '4681eb44-14cf-4ddf-8676-71949daf1ea4',   -- learning-center-article
             '5a7130a7-c530-447e-8b55-7a41e5a3152d')   -- news-post
  AND status = 'needs_human_review';

DO $verify$
DECLARE v_pages int; v_assets int; v_items int; v_real int;
BEGIN
    SELECT count(*) INTO v_pages FROM pages
    WHERE site_id = '00ff3af5-dad8-4770-9f70-3edc267a3c92'
      AND name IN ('news-post','learning-center-article','learning-center-post',
                   'grip-force-friction-calculator-guide',
                   'gripper-cycle-time-estimator-guide',
                   'gripper-payload-calculator-guide')
      AND status = 'archived';
    IF v_pages <> 6 THEN RAISE EXCEPTION 'expected 6 archived pages, got %', v_pages; END IF;

    -- The three real guides must remain active.
    SELECT count(*) INTO v_real FROM pages
    WHERE site_id = '00ff3af5-dad8-4770-9f70-3edc267a3c92'
      AND name IN ('tool-gripper-payload-calculator-guide',
                   'tool-grip-force-friction-calculator-guide',
                   'tool-gripper-cycle-time-estimator-guide')
      AND status = 'active';
    IF v_real <> 3 THEN RAISE EXCEPTION 'real guides disturbed (% active)', v_real; END IF;

    SELECT count(*) INTO v_assets FROM assets
    WHERE site_id = '00ff3af5-dad8-4770-9f70-3edc267a3c92'
      AND status = 'active'
      AND asset_key IN (
        'card_news_post','card_learning_center_article','card_learning_center_post',
        'card_grip_force_friction_calculator_guide',
        'card_gripper_cycle_time_estimator_guide',
        'card_gripper_payload_calculator_guide',
        'content_hero_news_post','content_hero_learning_center_article',
        'content_hero_learning_center_post',
        'content_hero_grip_force_friction_calculator_guide',
        'content_hero_gripper_cycle_time_estimator_guide',
        'content_hero_gripper_payload_calculator_guide');
    IF v_assets <> 0 THEN RAISE EXCEPTION '% dead-page assets still active', v_assets; END IF;

    SELECT count(*) INTO v_items FROM site_work_items
    WHERE id IN ('7695c973-350c-4345-9f69-672aa76182ca','4d6896c2-e108-4866-a118-fe8e2de98bb4',
                 '71d0a15d-da34-4cf6-8840-2006463965a5','4681eb44-14cf-4ddf-8676-71949daf1ea4',
                 '5a7130a7-c530-447e-8b55-7a41e5a3152d')
      AND status = 'wont_fix';
    IF v_items <> 5 THEN RAISE EXCEPTION 'expected 5 closed items, got %', v_items; END IF;

    RAISE NOTICE 'R6 applied: 6 rows archived, 12 assets superseded, 5 items closed; real guides intact';
END
$verify$;

COMMIT;
