-- SQL_2026-07-08_robothands_mission_brief.sql
--
-- Adds a mission_brief site_spec for robot-hands.com ahead of the from-scratch
-- re-plan, so the planner sees the news scope explicitly (## Mission block in
-- the build-site-planner prompt) rather than relying only on the
-- classification news_feed flag (RULE 11, homepage section only).
-- No current mission_brief row exists (verified 2026-07-08), so this is a
-- plain insert; the verify block aborts if that assumption is wrong.

\set ON_ERROR_STOP on

BEGIN;

DO $guard$
BEGIN
    IF EXISTS (
        SELECT 1 FROM site_specs
        WHERE site_id = '00ff3af5-dad8-4770-9f70-3edc267a3c92'
          AND aspect = 'mission_brief' AND is_current = true
    ) THEN
        RAISE EXCEPTION 'mission_brief already exists — merge instead of insert';
    END IF;
END
$guard$;

INSERT INTO site_specs (site_id, aspect, data, source, notes, is_current, created_by)
VALUES (
    '00ff3af5-dad8-4770-9f70-3edc267a3c92',
    'mission_brief',
    jsonb_build_object('text',
        'robot-hands.com is a reference and decision-support platform for engineers '
     || 'selecting industrial robot grippers and end-of-arm tooling: a filterable '
     || 'gripper catalogue, working selection tools (payload calculator, cycle-time '
     || 'estimator, MatchMatrix), and a learning centre. This build is a from-scratch '
     || 'rebuild of the site. NEWS IS IN SCOPE: include a latest-news section on the '
     || 'home page AND a dedicated news page (page_type blog-index, name "news", in '
     || 'the header navigation) carrying robotics/automation industry news — new '
     || 'gripper products, automation trends, and component price/supply shifts. The '
     || 'news feed content itself is populated continuously by the news pipeline '
     || 'after the build; the plan only needs the section and the page. Tool pages '
     || 'must be planned as page_type tool so their interactive components are built.'),
    'manual-recovery',
    'News hint for the 2026-07-08 from-scratch rebuild (imagery best-in-class workstream). Complements classification.content_features.news_feed.',
    true,
    'imagery-best-in-class-rebuild'
);

-- Verify
DO $verify$
DECLARE
    v_len int;
BEGIN
    SELECT length(data ->> 'text') INTO v_len
    FROM site_specs
    WHERE site_id = '00ff3af5-dad8-4770-9f70-3edc267a3c92'
      AND aspect = 'mission_brief' AND is_current = true;
    IF v_len IS NULL OR v_len < 100 THEN
        RAISE EXCEPTION 'mission_brief text missing or too short (len %)', v_len;
    END IF;
    RAISE NOTICE 'mission_brief inserted (text length %)', v_len;
END
$verify$;

COMMIT;
