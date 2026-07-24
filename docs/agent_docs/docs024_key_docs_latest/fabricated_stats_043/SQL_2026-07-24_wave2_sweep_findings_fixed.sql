-- 043 fleet sweep — wave 2 (2026-07-24): fix what the sweep found beyond the
-- three system-stats pages. Full sweep record lives in bugs_open/043 (dated
-- section, same day).
--
-- FINDINGS FIXED HERE:
-- (A) robot-hands/index brief-explanation — REGRESSION, the sweep's biggest
--     find: re-rendered 2026-07-24 10:54 and the LLM RE-INVENTED
--     "Gripper Models Indexed 2,400+" (removed by R4c on 07-20) plus
--     "MatchMatrix Queries Run: Growing daily" (nothing tracks queries). Live
--     proof that value-containment on an agent-rendered component gets
--     re-fabricated by a later render — THE motivating case for 043 candidate 3
--     (the scalar prompt rule, shipped separately today). Values recomputed
--     from the products table; the untrackable "queries run" stat replaced by
--     the countable "Published Figures Held".
-- (B) vonc about (content-block-about + gauntlet-cta) and index (gauntlet-cta +
--     brief-explanation) — the same fabricated arena theatre as the system-stats
--     block fixed in wave 1, including pure template mad-libs ("Happy Customers
--     14,203", "Avg. Rating: 6 Archetypes", "Setup Time: 4h 12m") and figures
--     wrong about the site's own content ("6 Archetypes" — it documents 8).
--     Replaced with register counts (computed) and honest format facts.
-- (C) ai-agent-orchestration index (system-stats) + about (content-block-about)
--     — same fabricated production metrics as the case-study page fixed in
--     wave 1 ("70+ agents / 8 departments / 1,000s concurrent"; about had
--     template mad-libs: "Satisfaction Rate: 30+", "Awards Won: 30 yrs").
--     Grounded in the platform's own registry (agents/sites/services/work
--     items), computed where the DB can count.
-- (D) system-stats-leopardess — a clone of system-stats with the SAME junk unit
--     fallbacks (%/ms/+/x) in its input_schema. Zero consumers today, so zero
--     live impact — this clears a dormant landmine for its first consumer
--     (same class as the R7c root fix).
--
-- RECORDED BUT NOT TOUCHED (see 043):
--   finetuning.uk/about "Clients Served 11+ / Satisfaction Rate 100%" — needs
--     its owner's real story; no honest replacement derivable from the DB.
--   vonc gauntlet-interface persisted 12,847/94,210/7 + schema fallbacks —
--     INERT (template no longer references the fields; live page verified
--     clean); gauntlet_dead_cta thread's territory.
--   leopardessconsulting/about "Agent Definitions 150+" — VERIFIED TRUE
--     (registry holds 170).
--
-- Apply:  kubectl -n ai-persona-system exec -i postgres-clients-0 -- \
--           psql -U clients_user -d clients_db < THIS_FILE

\set rh   '00ff3af5-dad8-4770-9f70-3edc267a3c92'
\set vonc '9ec3b9ee-5b08-461b-b4f8-9e1e03579c74'
\set aio  '2a8ebf9c-20a2-4c39-b191-840b012371da'

BEGIN;

-- ── (A) robot-hands/index brief-explanation ─────────────────────────────────
UPDATE page_components pc
   SET content_data = content_data || jsonb_build_object(
     'stat_1_label', 'Gripper Models Indexed',
     'stat_1_value', (SELECT count(*)::text FROM products
                      WHERE site_id = :'rh' AND category='gripper' AND status='active'),
     'stat_2_label', 'Actuation Technologies Benchmarked',
     'stat_2_value', (SELECT count(DISTINCT specifications->>'actuation')::text FROM products
                      WHERE site_id = :'rh' AND category='gripper' AND status='active'),
     'stat_3_label', 'Published Figures Held',
     'stat_3_value', (SELECT count(*)::text FROM products p2, jsonb_each_text(p2.specifications) e(k,v)
                      WHERE p2.site_id = :'rh' AND p2.category='gripper' AND p2.status='active'
                        AND e.k NOT IN ('manufacturer','actuation'))
   ),
   updated_at = now()
 WHERE pc.id = 'e4e6bb54-a761-41c9-8fc1-2e7e255c95a5';

-- ── (B) vonc — about + index non-system-stats components ────────────────────
UPDATE page_components pc
   SET content_data = content_data || jsonb_build_object(
     'stat_1_label', 'Archetypes Documented', 'stat_1_value',
        (SELECT count(*)::text FROM pages WHERE site_id=:'vonc' AND page_type='entity-page' AND status='active' AND deployed_at IS NOT NULL),
     'stat_2_label', 'Interactive Tools', 'stat_2_value',
        (SELECT count(*)::text FROM pages WHERE site_id=:'vonc' AND page_type='tool' AND status='active' AND deployed_at IS NOT NULL),
     'stat_3_label', 'Pages Live', 'stat_3_value',
        (SELECT count(*)::text FROM pages WHERE site_id=:'vonc' AND status='active' AND deployed_at IS NOT NULL)
   ), updated_at = now()
 WHERE pc.id = '02432c5c-670a-43d7-b49b-c4e15010796c';  -- about / content-block-about

UPDATE page_components pc
   SET content_data = content_data || jsonb_build_object(
     'stat_1_label', 'Archetypes Documented', 'stat_1_value',
        (SELECT count(*)::text FROM pages WHERE site_id=:'vonc' AND page_type='entity-page' AND status='active' AND deployed_at IS NOT NULL),
     'stat_2_label', 'Guides & Provocations', 'stat_2_value',
        (SELECT count(*)::text FROM pages WHERE site_id=:'vonc' AND page_type IN ('blog-post','section-index') AND status='active' AND deployed_at IS NOT NULL),
     'stat_3_label', 'Sign-up Required', 'stat_3_value', 'None'
   ), updated_at = now()
 WHERE pc.id = '33a2a442-94db-40d6-b1ad-1bee0e8bf6a2';  -- about / gauntlet-cta

UPDATE page_components pc
   SET content_data = content_data || jsonb_build_object(
     'stat_1_label', 'Archetypes Documented', 'stat_1_value',
        (SELECT count(*)::text FROM pages WHERE site_id=:'vonc' AND page_type='entity-page' AND status='active' AND deployed_at IS NOT NULL),
     'stat_2_label', 'Interactive Tools', 'stat_2_value',
        (SELECT count(*)::text FROM pages WHERE site_id=:'vonc' AND page_type='tool' AND status='active' AND deployed_at IS NOT NULL),
     'stat_3_label', 'Sign-up Required', 'stat_3_value', 'None'
   ), updated_at = now()
 WHERE pc.id = '038625f6-061d-4679-81b5-e4eff36549be';  -- index / gauntlet-cta

UPDATE page_components pc
   SET content_data = content_data || jsonb_build_object(
     'stat_1_label', 'Interactive Tools', 'stat_1_value',
        (SELECT count(*)::text FROM pages WHERE site_id=:'vonc' AND page_type='tool' AND status='active' AND deployed_at IS NOT NULL),
     'stat_2_label', 'Archetypes Documented', 'stat_2_value',
        (SELECT count(*)::text FROM pages WHERE site_id=:'vonc' AND page_type='entity-page' AND status='active' AND deployed_at IS NOT NULL),
     'stat_3_label', 'Free to Play', 'stat_3_value', '100%'
   ), updated_at = now()
 WHERE pc.id = '74945a7d-2988-4c83-a78b-159193f6c4b9';  -- index / brief-explanation

-- ── (C) ai-agent-orchestration — index system-stats + about ─────────────────
UPDATE page_components pc
   SET content_data = content_data || jsonb_build_object(
     'footnote_text', 'Figures are live counts from the platform''s own database — agent registry, site register and completed work-item ledger — taken 2026-07-24. Architecture runs on Kubernetes, Kafka, and Postgres.',
     'stat1_label', 'Deployed Agents',
     'stat1_value', (SELECT count(*)::text FROM agent_definitions
                     WHERE is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL),
     'stat1_suffix', '',
     'stat1_description', 'Active agent definitions in the production registry — each independently configured, versioned, and running with observable state.',
     'stat2_label', 'Live Sites in Production',
     'stat2_value', (SELECT count(*)::text FROM sites WHERE status NOT IN ('archived','deleted','pool')),
     'stat2_suffix', '',
     'stat2_description', 'Distinct production websites the platform builds and operates end-to-end — content, imagery, interactive tooling and deployment.',
     'stat3_label', 'Backend Services',
     'stat3_value', '17',
     'stat3_suffix', '',
     'stat3_description', 'Purpose-built services behind the fleet — orchestration, scheduling, and adapters for web scraping, search, imagery, git and more — each deployed and scaled independently on Kubernetes.',
     'stat4_label', 'Automated Work Items Completed',
     'stat4_value', (SELECT to_char(count(*), 'FM9,999,999') FROM site_work_items
                     WHERE status IN ('complete','verified')),
     'stat4_suffix', '',
     'stat4_description', 'Discrete pieces of site work — builds, re-renders, repairs, audits — detected, dispatched and completed by the platform''s own work-item pipeline.'
   ), updated_at = now()
 WHERE pc.page_id = (SELECT id FROM pages WHERE site_id = :'aio' AND name='index')
   AND pc.component_id = 'fdd92ad4-521a-4602-89cf-7ee1a66c10f1';

UPDATE page_components pc
   SET content_data = content_data || jsonb_build_object(
     'stat_1_label', 'Active Agents',
     'stat_1_value', (SELECT count(*)::text FROM agent_definitions
                      WHERE is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL),
     'stat_2_label', 'Live Sites in Production',
     'stat_2_value', (SELECT count(*)::text FROM sites WHERE status NOT IN ('archived','deleted','pool')),
     'stat_3_label', 'Backend Services',
     'stat_3_value', '17'
   ), updated_at = now()
 WHERE pc.page_id = (SELECT id FROM pages WHERE site_id = :'aio' AND name='about')
   AND pc.content_data->>'stat_1_label' = 'Clients Served';

-- ── (D) system-stats-leopardess clone — clear the dormant junk fallbacks ────
UPDATE content_components
   SET input_schema = jsonb_set(jsonb_set(jsonb_set(jsonb_set(
         input_schema,
         '{fields,stat1_suffix,fallback}', '""'::jsonb),
         '{fields,stat2_suffix,fallback}', '""'::jsonb),
         '{fields,stat3_suffix,fallback}', '""'::jsonb),
         '{fields,stat4_suffix,fallback}', '""'::jsonb),
       updated_at = now()
 WHERE name = 'system-stats-leopardess';

\echo '--- verify: no fabricated values left on the touched rows ---'
SELECT s.domain, p.name AS page, e.k, left(e.v,40) AS v
FROM page_components pc JOIN pages p ON p.id=pc.page_id JOIN sites s ON s.id=p.site_id
CROSS JOIN LATERAL jsonb_each_text(pc.content_data) e(k,v)
WHERE pc.id IN ('e4e6bb54-a761-41c9-8fc1-2e7e255c95a5','02432c5c-670a-43d7-b49b-c4e15010796c',
                '33a2a442-94db-40d6-b1ad-1bee0e8bf6a2','038625f6-061d-4679-81b5-e4eff36549be',
                '74945a7d-2988-4c83-a78b-159193f6c4b9')
  AND e.k ~ '^stat[_]?[0-9]+_?(label|value)$'
ORDER BY s.domain, p.name, pc.id, e.k;

\echo '--- ai-agent index/about after ---'
SELECT p.name, e.k, left(e.v,40) FROM page_components pc
JOIN pages p ON p.id=pc.page_id
CROSS JOIN LATERAL jsonb_each_text(pc.content_data) e(k,v)
WHERE p.site_id = :'aio' AND p.name IN ('index','about')
  AND e.k ~ '^stat[_]?[0-9]+_?(label|value)$' ORDER BY p.name, e.k;

-- ── Re-renders: rh/index, vonc/about, aio/index + about (vonc/index and the
--    other wave-1 pages are already queued or complete) ──────────────────────
INSERT INTO site_work_items
  (site_id, source, item_type, severity, summary, status, created_by, pipeline,
   priority, triaged_at, handler_agent, item_key, spec)
SELECT p.site_id, 'fabricated-stats-043-wave2', 'page_rerender', 'medium',
  'Rerender ' || p.name || ' — sweep wave-2 fabricated stats replaced with traced figures (043)',
  'triaged', 'session-2026-07-24-043-treatment', 'build',
  20, now(), 'page-rerender',
  'page_rerender_' || p.name || '_043w2_' || p.site_id::text,
  jsonb_build_object('domain', s.domain, 'reason','cta_links_stale',
                     'page_id',p.id,'page_name',p.name,'filename',ltrim(p.url,'/'))
FROM pages p JOIN sites s ON s.id=p.site_id
WHERE (p.site_id = :'rh' AND p.name='index')
   OR (p.site_id = :'vonc' AND p.name='about')
   OR (p.site_id = :'aio' AND p.name IN ('index','about'));

\echo '--- queued ---'
SELECT status, count(*) FROM site_work_items WHERE source='fabricated-stats-043-wave2' GROUP BY 1;

COMMIT;
