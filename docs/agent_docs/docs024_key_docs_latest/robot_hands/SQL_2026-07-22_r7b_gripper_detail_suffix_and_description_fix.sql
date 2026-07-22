-- R7b (2026-07-22) — correct gripper-detail's system-stats block: the live page
-- still carried the placeholder-suffix bug, and R7's count bump left two
-- descriptions stating figures that no longer match the catalogue.
--
-- FOUND BY VERIFYING THE REAL LIVE URL (not the status, not the wrong URL).
-- gripper-detail deploys to /entities/gripper-detail.html (pages.url), NOT
-- /gripper-detail.html at root — I checked root first and saw a stale May-02 file
-- with empty stats, which briefly read as a delivery gap. On the correct URL R7's
-- values ARE live (10 / 6 / 4 / 39), but:
--   1. stat{1..4}_suffix are unedited generic-template placeholders — %, ms, +, x
--      — so the page rendered "10%", "6ms", "4+", "39x". This is bug_open/043
--      point (b), the "nobody read the rendered output" tell, STILL LIVE here
--      (the 07-20 containment cleared it on about — which has no suffix keys — but
--      not on gripper-detail). Clear all four.
--   2. stat2_description listed five manufacturers ("Schunk, OnRobot, Robotiq,
--      Zimmer Group and Festo") while R7 set the value to 6 — add Schmalz.
--   3. stat4_description asserted "two models publish no payload rating and two
--      publish no IP rating". True for the original five; with ten it is 4 with no
--      payload and 7 with no IP rating (verified:
--        count(*) FILTER (WHERE NOT specifications ? 'payload') = 4,
--        count(*) FILTER (WHERE NOT specifications ? 'ip_rating') = 7).
--      Genericised so there is no hardcoded count to drift again — the same
--      fabrication-family lesson R7 was about.
--
-- Apply:  kubectl -n ai-persona-system exec -i postgres-clients-0 -- \
--           psql -U clients_user -d clients_db < THIS_FILE

\set site '00ff3af5-dad8-4770-9f70-3edc267a3c92'

BEGIN;

UPDATE page_components pc
   SET content_data = content_data
       || '{"stat1_suffix":"","stat2_suffix":"","stat3_suffix":"","stat4_suffix":""}'::jsonb
       || jsonb_build_object(
            'stat2_description',
            'Schunk, OnRobot, Robotiq, Zimmer Group, Festo and Schmalz. Figures are reproduced as each manufacturer publishes them; where one does not publish a parameter it is shown as unpublished, never estimated.',
            'stat4_description',
            'Individual specification figures across the indexed grippers. Coverage is uneven by design — not every manufacturer publishes every parameter (payload and IP rating are the most often omitted), and MatchMatrix flags those gaps rather than filling them.'
          ),
       updated_at = now()
 WHERE pc.page_id = (SELECT id FROM pages WHERE site_id = :'site' AND name='gripper-detail')
   AND pc.slot_name = 'system-stats';

\echo '--- stat values + suffixes after fix (suffixes must be empty) ---'
SELECT e.k, e.v
FROM page_components pc JOIN pages p ON p.id=pc.page_id
CROSS JOIN LATERAL jsonb_each_text(pc.content_data) e(k,v)
WHERE p.site_id = :'site' AND p.name='gripper-detail' AND pc.slot_name='system-stats'
  AND e.k ~ 'stat[0-9]_(value|suffix)' ORDER BY e.k;

-- Re-render just gripper-detail so the corrected block reaches /entities/…
INSERT INTO site_work_items
  (site_id, source, item_type, severity, summary, status, created_by, pipeline,
   priority, triaged_at, handler_agent, item_key, spec)
SELECT
  p.site_id, 'robot-hands-r7b-suffix-fix', 'page_rerender', 'medium',
  'Rerender gripper-detail — clear placeholder stat suffixes (%/ms/+/x) and refresh count-driven descriptions (R7b)',
  'triaged', 'session-2026-07-22-robot-hands', 'build',
  20, now(), 'page-rerender',
  'page_rerender_' || p.name || '_r7bsuffix_' || p.site_id::text,
  jsonb_build_object('domain','robot-hands.com','reason','cta_links_stale',
                     'page_id',p.id,'page_name',p.name,'filename',ltrim(p.url,'/'))
FROM pages p
WHERE p.site_id = :'site' AND p.name = 'gripper-detail';

\echo '--- queued ---'
SELECT status, handler_agent, count(*) FROM site_work_items
WHERE source='robot-hands-r7b-suffix-fix' GROUP BY 1,2;

COMMIT;
