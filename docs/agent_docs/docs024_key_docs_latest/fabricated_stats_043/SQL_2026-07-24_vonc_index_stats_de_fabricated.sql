-- 043 treatment — vonc.com/index system-stats (2026-07-24).
--
-- WHAT WAS LIVE: fabricated live-activity theatre, rendered with junk suffixes:
--   "Takes Filed Today 14,203%" ("since midnight"), "Avg. Split % 61ms" ("hasn't
--   produced a majority verdict in over two months"), "Active Gauntlets 9+"
--   ("Nine Provocations live simultaneously"), "Archetypes in Play 9x" ("The
--   Contrarian is currently winning"), footnote "updated every 15 minutes during
--   active Gauntlet windows". NO takes/gauntlet-activity tables exist anywhere in
--   the database — there is nothing these numbers could have counted. Even the
--   archetype count was wrong about the site's own content: vonc documents EIGHT
--   archetypes (catalyst, judge, maker, mentor, oracle, scout, surgeon,
--   wildcard), and there is no archetype called "The Contrarian".
--   Same class the gauntlet_dead_cta thread stripped from the tool's own
--   interface on 2026-07-22 (12,847/94,210/38%/7 + invented leaderboard names) —
--   this is the homepage sibling it did not cover.
--
-- FIX: count what the site actually holds, computed from the pages register so
-- the values trace (bug 043 fix-candidate-1 style). Voice kept, fake liveness
-- dropped. Suffixes cleared (the R7c schema-fallback fix means empty now holds).
--
-- Ground truth (2026-07-24): 8 archetype entity pages, 3 interactive tools
-- (archetype-taster-quiz, arena, gauntlet), 17 deployed active pages.
--
-- Apply:  kubectl -n ai-persona-system exec -i postgres-clients-0 -- \
--           psql -U clients_user -d clients_db < THIS_FILE

\set site '9ec3b9ee-5b08-461b-b4f8-9e1e03579c74'
\set comp 'fdd92ad4-521a-4602-89cf-7ee1a66c10f1'

BEGIN;

UPDATE page_components pc
   SET content_data = content_data || jsonb_build_object(
     'section_headline', 'The Arena, Counted',
     'section_intro',    'No projections and no scoreboard theatre — these are the register counts of what the Arena actually holds.',
     'footnote_text',    'Counts come from the site''s own page register, as of 2026-07-24. When the roster grows, these numbers follow it.',
     'stat1_label', 'Archetypes Documented',
     'stat1_value', (SELECT count(*)::text FROM pages
                     WHERE site_id = :'site' AND page_type='entity-page'
                       AND status='active' AND deployed_at IS NOT NULL),
     'stat1_suffix', '',
     'stat1_description', 'Catalyst, Judge, Maker, Mentor, Oracle, Scout, Surgeon and Wildcard — each with its own documented profile page.',
     'stat2_label', 'Interactive Tools',
     'stat2_value', (SELECT count(*)::text FROM pages
                     WHERE site_id = :'site' AND page_type='tool'
                       AND status='active' AND deployed_at IS NOT NULL),
     'stat2_suffix', '',
     'stat2_description', 'The Archetype Taster Quiz, the Arena and the Gauntlet — all live, all client-side, no sign-up.',
     'stat3_label', 'Guides & Provocations Published',
     'stat3_value', (SELECT count(*)::text FROM pages
                     WHERE site_id = :'site' AND page_type IN ('blog-post','section-index')
                       AND status='active' AND deployed_at IS NOT NULL),
     'stat3_suffix', '',
     'stat3_description', 'Written pieces currently live. Small on purpose — nothing is published to inflate a counter.',
     'stat4_label', 'Pages Live',
     'stat4_value', (SELECT count(*)::text FROM pages
                     WHERE site_id = :'site' AND status='active' AND deployed_at IS NOT NULL),
     'stat4_suffix', '',
     'stat4_description', 'Every deployed page on the site, archetypes to tools. What you can click is what gets counted.'
   ),
   updated_at = now()
 WHERE pc.page_id = (SELECT id FROM pages WHERE site_id = :'site' AND name='index')
   AND pc.component_id = :'comp';

\echo '--- after (labels, values, suffixes must be honest/empty) ---'
SELECT e.k, left(e.v,60) FROM page_components pc
JOIN pages p ON p.id=pc.page_id
CROSS JOIN LATERAL jsonb_each_text(pc.content_data) e(k,v)
WHERE p.site_id = :'site' AND p.name='index' AND pc.component_id = :'comp'
  AND e.k ~ 'stat[0-9]_(label|value|suffix)' ORDER BY e.k;

INSERT INTO site_work_items
  (site_id, source, item_type, severity, summary, status, created_by, pipeline,
   priority, triaged_at, handler_agent, item_key, spec)
SELECT p.site_id, 'fabricated-stats-043-vonc', 'page_rerender', 'medium',
  'Rerender index — fabricated live-activity stats replaced with register counts (043)',
  'triaged', 'session-2026-07-24-043-treatment', 'build',
  20, now(), 'page-rerender',
  'page_rerender_' || p.name || '_043stats_' || p.site_id::text,
  jsonb_build_object('domain','vonc.com','reason','cta_links_stale',
                     'page_id',p.id,'page_name',p.name,'filename',ltrim(p.url,'/'))
FROM pages p WHERE p.site_id = :'site' AND p.name='index';

\echo '--- queued ---'
SELECT status, count(*) FROM site_work_items WHERE source='fabricated-stats-043-vonc' GROUP BY 1;

COMMIT;
