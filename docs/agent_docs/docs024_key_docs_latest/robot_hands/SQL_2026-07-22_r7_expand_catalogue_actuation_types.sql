-- R7 (2026-07-22) — EXPAND the gripper catalogue so the "six actuation
-- technologies" claim is TRUE, and make the stat blocks trace to the catalogue.
--
-- OWNER DECISION (2026-07-22): of the four options for the fabricated
-- "six actuation technologies" positioning (rewrite prose / EXPAND catalogue /
-- minimal strip / leave), the owner chose **expand the catalogue** — keep the
-- ambitious positioning and back it with real, datasheet-sourced products.
--
-- WHY THIS IS A DATA FIX, NOT A COPY EDIT.  bug_open/043 established the
-- `products` table (category='gripper', site_id below) as the index-of-record:
-- "the index holds five grippers." Those five are all electric parallel-jaw
-- grippers, so the site's core claim — "indexes grippers across six actuation
-- technologies: pneumatic, electric, vacuum, magnetic, soft-robotic, adhesive" —
-- advertised FOUR technologies (vacuum/magnetic/soft-robotic/adhesive) with ZERO
-- grippers and asserted pneumatic where the data (24 V DC specs) said electric.
-- This file adds ONE real, sourced product for each missing technology, so the
-- index genuinely spans six. Every figure below was read off a manufacturer or
-- distributor page on 2026-07-22 (verified_date). NO figure is invented — that is
-- the whole point of the exercise (see 020/043: the fabrication family this
-- workstream exists to fight).
--
-- The catalogue is a DATA index, not a browsable listing: the gripper-catalog
-- page is static prose (verified live 2026-07-22 — 0 gripper names rendered), and
-- MatchMatrix embeds the 5 parallel-jaw grippers as a static dataset. So the new
-- non-jaw grippers back the CLAIM and the STAT COUNTS but do not auto-generate
-- detail pages or MatchMatrix rows. That is a pre-existing rendering gap, recorded
-- for follow-up, not part of this decision.
--
-- SOURCES (each product's specifications come from the URL stored in its
--          content_data.source_url; figures verified 2026-07-22):
--   Festo DHPS-10-A (pneumatic) — 34.5 N/jaw closing (39 N opening) @ 6 bar,
--     3 mm stroke/jaw, 67 g, 2–8 bar, 0.02 mm repeat. Festo datasheet 1254040,
--     verified via https://www.motionworld.com/products/148789/festo-1254040-dhps-10-a-parallel-gripper
--     (the Festo PDF itself is binary and not fetch-readable).
--   OnRobot VG10 (vacuum) — 15 kg payload, electric dual-zone (no external air).
--     https://onrobot.com/en/products/vg10-vacuum-gripper
--   OnRobot Soft Gripper SG (soft-robotic) — 2.2 kg payload, 11–118 mm grip,
--     food-grade silicone. https://onrobot.com/en/products/soft-gripper
--   OnRobot Gecko SP5 (adhesive) — 5 kg payload, gecko/van der Waals adhesion
--     (no air, no electricity). https://onrobot.com/en/products/gecko-gripper
--   Schmalz SGM-HP 50 (magnetic) — 560 N holding (385 N with friction ring),
--     50 mm dia, workpiece up to 350 °C.
--     https://www.schmalz.com/en-us/vacuum-technology-for-automation/vacuum-components/special-grippers/magnetic-grippers/magnetic-grippers-sgm-hp-ht-306089/
--
-- Apply:  kubectl -n ai-persona-system exec -i postgres-clients-0 -- \
--           psql -U clients_user -d clients_db < THIS_FILE

\set site '00ff3af5-dad8-4770-9f70-3edc267a3c92'

BEGIN;

-- ── 1. Retrofit the structured actuation key on the existing 5 (all electric) ──
--    so a single query counts technologies. Their subcategory already reads
--    "Electric parallel gripper …"; this just makes it queryable.
UPDATE products
   SET specifications = specifications || '{"actuation":"electric"}'::jsonb,
       updated_at = now()
 WHERE site_id = :'site' AND category = 'gripper'
   AND (specifications->>'actuation') IS NULL;

-- ── 2. Insert one real, sourced gripper per missing actuation technology ──────
INSERT INTO products
  (site_id, name, slug, category, subcategory, status, specifications,
   content_data, published_at)
VALUES
  (:'site', 'Festo DHPS-10-A', 'festo-dhps-10-a', 'gripper',
   'Pneumatic parallel gripper', 'active',
   '{"manufacturer":"Festo","actuation":"pneumatic","gripping_force":"34.5 N per jaw closing (39 N opening) at 6 bar","stroke":"3 mm per jaw","weight":"67 g","operating_pressure":"2 to 8 bar","repeat_accuracy":"0.02 mm"}'::jsonb,
   '{"source_url":"https://www.motionworld.com/products/148789/festo-1254040-dhps-10-a-parallel-gripper","verified_date":"2026-07-22"}'::jsonb,
   now()),

  (:'site', 'OnRobot VG10', 'onrobot-vg10', 'gripper',
   'Electric vacuum gripper (dual-zone)', 'active',
   '{"manufacturer":"OnRobot","actuation":"vacuum","payload":"15 kg (35 lb)","type":"electric vacuum, dual-zone, built-in pump (no external air supply)"}'::jsonb,
   '{"source_url":"https://onrobot.com/en/products/vg10-vacuum-gripper","verified_date":"2026-07-22"}'::jsonb,
   now()),

  (:'site', 'OnRobot Soft Gripper SG', 'onrobot-soft-gripper-sg', 'gripper',
   'Soft-robotic silicone gripper (food-grade)', 'active',
   '{"manufacturer":"OnRobot","actuation":"soft-robotic","payload":"2.2 kg (depends on shape, softness, friction)","grip_range":"11 to 118 mm","material":"food-grade silicone"}'::jsonb,
   '{"source_url":"https://onrobot.com/en/products/soft-gripper","verified_date":"2026-07-22"}'::jsonb,
   now()),

  (:'site', 'OnRobot Gecko SP5', 'onrobot-gecko-sp5', 'gripper',
   'Adhesive (gecko / van der Waals) gripper', 'active',
   '{"manufacturer":"OnRobot","actuation":"adhesive","payload":"5 kg","principle":"gecko-inspired van der Waals adhesion (no air, no electricity)"}'::jsonb,
   '{"source_url":"https://onrobot.com/en/products/gecko-gripper","verified_date":"2026-07-22"}'::jsonb,
   now()),

  (:'site', 'Schmalz SGM-HP 50', 'schmalz-sgm-hp-50', 'gripper',
   'Magnetic gripper (permanent magnet, size 50)', 'active',
   '{"manufacturer":"Schmalz","actuation":"magnetic","holding_force":"560 N (without friction ring), 385 N (with friction ring)","diameter":"50 mm","max_workpiece_temperature":"350 C"}'::jsonb,
   '{"source_url":"https://www.schmalz.com/en-us/vacuum-technology-for-automation/vacuum-components/special-grippers/magnetic-grippers/magnetic-grippers-sgm-hp-ht-306089/","verified_date":"2026-07-22"}'::jsonb,
   now());

\echo '--- catalogue after insert (expect 10 models / 6 manufacturers / 6 technologies) ---'
SELECT
  count(*)                                            AS models,
  count(DISTINCT specifications->>'manufacturer')     AS manufacturers,
  count(DISTINCT specifications->>'actuation')        AS technologies
FROM products WHERE site_id = :'site' AND category = 'gripper' AND status = 'active';

\echo '--- the six actuation technologies now in the index ---'
SELECT specifications->>'actuation' AS actuation, count(*)
FROM products WHERE site_id = :'site' AND category = 'gripper' AND status = 'active'
GROUP BY 1 ORDER BY 1;

-- ── 3. Re-point the stat blocks at the catalogue (bug 043 fix candidate 1: a
--       stat value should trace to a query, never be typed by hand). Each value
--       below is COMPUTED from the products table, so it cannot drift from it. ──

-- about / content-block-about (keys: stat_1_value models, stat_2_value mfrs)
UPDATE page_components pc SET content_data = jsonb_set(jsonb_set(
    content_data,
    '{stat_1_value}',
    to_jsonb((SELECT count(*)::text FROM products
              WHERE site_id = :'site' AND category='gripper' AND status='active'))),
    '{stat_2_value}',
    to_jsonb((SELECT count(DISTINCT specifications->>'manufacturer')::text FROM products
              WHERE site_id = :'site' AND category='gripper' AND status='active'))),
    updated_at = now()
 WHERE pc.page_id = (SELECT id FROM pages WHERE site_id = :'site' AND name='about')
   AND pc.content_data->>'stat_1_label' = 'Gripper Models Indexed';

-- gripper-detail / system-stats (keys: stat1_value models, stat2_value mfrs,
--   stat4_value published figures = every spec value except manufacturer/actuation)
UPDATE page_components pc SET content_data = jsonb_set(jsonb_set(jsonb_set(
    content_data,
    '{stat1_value}',
    to_jsonb((SELECT count(*)::text FROM products
              WHERE site_id = :'site' AND category='gripper' AND status='active'))),
    '{stat2_value}',
    to_jsonb((SELECT count(DISTINCT specifications->>'manufacturer')::text FROM products
              WHERE site_id = :'site' AND category='gripper' AND status='active'))),
    '{stat4_value}',
    to_jsonb((SELECT count(*)::text FROM products p2, jsonb_each_text(p2.specifications) e(k,v)
              WHERE p2.site_id = :'site' AND p2.category='gripper' AND p2.status='active'
                AND e.k NOT IN ('manufacturer','actuation')))),
    updated_at = now()
 WHERE pc.page_id = (SELECT id FROM pages WHERE site_id = :'site' AND name='gripper-detail')
   AND pc.content_data->>'stat1_label' = 'Gripper Models Indexed';

\echo '--- stat blocks after update (about + gripper-detail) ---'
SELECT p.name, pc.content_data->>'stat_1_value' AS about_models,
       pc.content_data->>'stat1_value' AS gd_models,
       pc.content_data->>'stat_2_value' AS about_mfrs,
       pc.content_data->>'stat2_value' AS gd_mfrs,
       pc.content_data->>'stat4_value' AS gd_figures
FROM page_components pc JOIN pages p ON p.id=pc.page_id
WHERE p.site_id = :'site' AND p.name IN ('about','gripper-detail')
  AND (pc.content_data->>'stat_1_label'='Gripper Models Indexed'
    OR pc.content_data->>'stat1_label'='Gripper Models Indexed')
ORDER BY p.name;

-- index / brief-explanation already renders "Actuation technologies benchmarked = 6";
-- that value was fabricated before this file and is now BACKED (count above = 6),
-- so it needs no edit and no re-render. Asserted, not assumed:
\echo '--- index actuation stat is now backed (expect value 6 == technologies 6) ---'
SELECT pc.content_data->>'stat_1_value' AS index_shows,
       (SELECT count(DISTINCT specifications->>'actuation') FROM products
        WHERE site_id = :'site' AND category='gripper' AND status='active') AS real_count
FROM page_components pc JOIN pages p ON p.id=pc.page_id
WHERE p.site_id = :'site' AND p.name='index'
  AND pc.content_data->>'stat_1_label' = 'Actuation technologies benchmarked';

-- ── 4. Queue re-renders so the corrected counts reach the live pages. ─────────
--    Only about + gripper-detail changed a rendered number (index is unchanged).
--    reason='cta_links_stale' is the PROVEN reason that reaches the real
--    rerender_sections branch (per bug 024; these two pages already re-rendered
--    through it in R4d). handler_agent + item_key are BOTH required or the item
--    sits unroutable at 'triaged' looking queued (R4d correction). Distinct
--    item_key ('_r7stats_') so dedup does not collide with the R4d items.
INSERT INTO site_work_items
  (site_id, source, item_type, severity, summary, status, created_by, pipeline,
   priority, triaged_at, handler_agent, item_key, spec)
SELECT
  p.site_id,
  'robot-hands-r7-catalogue-expand',
  'page_rerender',
  'medium',
  'Rerender ' || p.name || ' — gripper index expanded to 6 actuation technologies; model/manufacturer/figure counts recomputed (R7)',
  'triaged',
  'session-2026-07-22-robot-hands',
  'build',
  20,
  now(),
  'page-rerender',
  'page_rerender_' || p.name || '_r7stats_' || p.site_id::text,
  jsonb_build_object(
    'domain',    'robot-hands.com',
    'reason',    'cta_links_stale',
    'page_id',   p.id,
    'page_name', p.name,
    'filename',  ltrim(p.url, '/')
  )
FROM pages p
WHERE p.site_id = :'site' AND p.name IN ('about', 'gripper-detail');

\echo '--- queued re-renders ---'
SELECT status, handler_agent, priority, count(*)
FROM site_work_items
WHERE site_id = :'site' AND source = 'robot-hands-r7-catalogue-expand'
GROUP BY 1,2,3;

COMMIT;
