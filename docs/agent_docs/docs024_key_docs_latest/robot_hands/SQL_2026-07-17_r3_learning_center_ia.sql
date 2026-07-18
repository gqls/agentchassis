-- SQL_2026-07-17_r3_learning_center_ia.sql
--
-- R3 of HANDOFF_2026-07-17_robot_hands_site_fixes.md: learning-center URL
-- sprawl. Three URLs served one concept:
--   /learning-center.html       — category info-card grid; WAS the primary
--                                 nav target; its card CTAs are phantoms
--                                 (5 open cta/phantom items)
--   /learning-center-hub.html   — the REAL article listing (query.blog_posts
--                                 + F2.1 eligibility + imagery cards); was
--                                 reachable only directly
--   /learning-center/index.html — plan-residue section index; WAS the footer
--                                 utility nav target; its rebuild has failed
--                                 twice (needs_page churn)
--
-- IA decision (delegated by the handoff; matches its likely-IA): the HUB is
-- the canonical Learning Center.
--   1. Primary + utility nav → /learning-center-hub.html.
--   2. /learning-center.html KEPT active as a body-linked landing (archiving
--      it would flip every baked in-body link to it into phantom_internal_link
--      churn) but dropped from header nav flags. Folding its category grid
--      into the hub is future content work, noted in the handoff doc.
--   3. learning-center-index ARCHIVED (pure residue) + its failed rebuild
--      items closed.
-- Nav-flag edits on plan rows are NOT a re-plan (no build-site-planner run)
-- — the replan-clobbers-built-pages landmine does not apply.
--
-- Deploy note: nav is baked into pages at assembly. Applied mid-drain of the
-- R1 37-page batch; a follow-up rerender (components + pages) bakes the new
-- nav everywhere.

\set ON_ERROR_STOP on

CREATE TABLE IF NOT EXISTS site_nav_items_backup_20260717_r3 AS
SELECT * FROM site_nav_items WHERE site_id = '00ff3af5-dad8-4770-9f70-3edc267a3c92';
SELECT count(*) AS bak_nav_items FROM site_nav_items_backup_20260717_r3;

BEGIN;

-- 1. Primary nav: Learning Center → hub.
UPDATE site_nav_items
SET url = '/learning-center-hub.html',
    page_id = (SELECT id FROM pages WHERE site_id='00ff3af5-dad8-4770-9f70-3edc267a3c92' AND name='learning-center-hub'),
    updated_at = now()
WHERE id = '3ebd6f04-e093-4e91-a5dd-8acbb0f6e521';

-- 2. Utility (footer) nav: Learning Center → hub (was the residue index).
UPDATE site_nav_items
SET url = '/learning-center-hub.html',
    page_id = (SELECT id FROM pages WHERE site_id='00ff3af5-dad8-4770-9f70-3edc267a3c92' AND name='learning-center-hub'),
    updated_at = now()
WHERE id = '11cb2310-64c6-4ee5-aaa2-a0495c74d8ad';

-- 3. Page-row nav flags: hub owns the header slot; grid page demoted;
--    residue archived. nav_label kept 'Learning Center' on all.
UPDATE pages SET in_header = true, nav_order = 6, updated_at = now()
WHERE site_id='00ff3af5-dad8-4770-9f70-3edc267a3c92' AND name='learning-center-hub';

UPDATE pages SET in_header = false, nav_order = 210, updated_at = now()
WHERE site_id='00ff3af5-dad8-4770-9f70-3edc267a3c92' AND name='learning-center';

UPDATE pages SET status = 'archived', in_header = false, in_footer = false, updated_at = now()
WHERE site_id='00ff3af5-dad8-4770-9f70-3edc267a3c92' AND name='learning-center-index';

-- 4. Plan rows: keep site_plan_pages nav flags in lockstep so a future
--    PopulateNavTables run derives the same nav (flag edits only — NOT a
--    re-plan).
UPDATE site_plan_pages spp
SET in_header = true, nav_order = 6, nav_label = 'Learning Center'
FROM site_plans sp
WHERE sp.id = spp.plan_id AND sp.is_current = true
  AND sp.site_id = '00ff3af5-dad8-4770-9f70-3edc267a3c92'
  AND spp.name = 'learning-center-hub';

UPDATE site_plan_pages spp
SET in_header = false, nav_order = 210
FROM site_plans sp
WHERE sp.id = spp.plan_id AND sp.is_current = true
  AND sp.site_id = '00ff3af5-dad8-4770-9f70-3edc267a3c92'
  AND spp.name = 'learning-center';

UPDATE site_plan_pages spp
SET in_header = false, in_footer = false
FROM site_plans sp
WHERE sp.id = spp.plan_id AND sp.is_current = true
  AND sp.site_id = '00ff3af5-dad8-4770-9f70-3edc267a3c92'
  AND spp.name = 'learning-center-index';

-- 5. Close the two failed learning-center-index rebuild items (page retired).
UPDATE site_work_items
SET status = 'wont_fix', updated_at = now(),
    error = COALESCE(error || E'\n', '')
            || 'Closed 2026-07-17 (R3): learning-center-index archived as plan residue; canonical Learning Center is /learning-center-hub.html. See robot_hands/SQL_2026-07-17_r3_learning_center_ia.sql.'
WHERE id IN ('ee0934d6-cc09-4d6e-973d-ddc0d526c28b',   -- Full rebuild of learning-center-index (failed 2026-07-17)
             '54845d41-da5e-4b3f-a901-a49234e04a2c')   -- Re-render learning-center-index — stale residue (failed 2026-07-10)
  AND status = 'failed';

DO $verify$
DECLARE v_url text; v_cnt int;
BEGIN
    SELECT url INTO v_url FROM site_nav_items WHERE id='3ebd6f04-e093-4e91-a5dd-8acbb0f6e521';
    IF v_url <> '/learning-center-hub.html' THEN RAISE EXCEPTION 'primary nav not repointed (%)', v_url; END IF;
    SELECT url INTO v_url FROM site_nav_items WHERE id='11cb2310-64c6-4ee5-aaa2-a0495c74d8ad';
    IF v_url <> '/learning-center-hub.html' THEN RAISE EXCEPTION 'utility nav not repointed (%)', v_url; END IF;

    SELECT count(*) INTO v_cnt FROM pages
    WHERE site_id='00ff3af5-dad8-4770-9f70-3edc267a3c92'
      AND name='learning-center-index' AND status='archived';
    IF v_cnt <> 1 THEN RAISE EXCEPTION 'learning-center-index not archived'; END IF;

    SELECT count(*) INTO v_cnt FROM pages
    WHERE site_id='00ff3af5-dad8-4770-9f70-3edc267a3c92'
      AND name='learning-center' AND status='active' AND in_header=false;
    IF v_cnt <> 1 THEN RAISE EXCEPTION 'learning-center grid page not demoted-but-active'; END IF;

    RAISE NOTICE 'R3 applied: hub canonical in both nav groups; grid demoted; residue archived';
END
$verify$;

COMMIT;
