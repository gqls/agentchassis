-- 043 treatment — ai-agent-orchestration.com/case-study-kafka-consumer-group-
-- remediation system-stats (2026-07-24).
--
-- WHAT WAS LIVE (rendered with junk suffixes: 70%, 8ms, 30+, 1000x), presented
-- as "the numbers from deployed infrastructure — not projected capacity from a
-- whitepaper": "70 Deployed Agents / 8 Departments Served / 30 Agent Types /
-- 1000 Concurrent Instances". The platform behind this site has no departments
-- and does not run 1000 concurrent instances; none of the four values matched
-- anything countable. Ironically the fabrication UNDERSTATED the real registry:
-- the platform actually runs 170 active agent definitions.
--
-- FIX: ground every figure in the platform's own database (this site is the
-- platform's shop window, so its own DB IS the primary source):
--   Deployed Agents           = count(agent_definitions active, non-snapshot)   [170 at apply time]
--   Live Sites in Production  = count(sites not archived/deleted/pool)          [13]
--   Backend Services          = 17 — the service manifests under
--                               deployments/kustomize/services/ (admin-dashboard,
--                               agent-chassis, analyser-adapter, auth-service,
--                               browser-runner-adapter, business-intel,
--                               content-creator-agent, core-manager, git-adapter,
--                               image-generator-adapter, kafka-scheduler,
--                               reasoning-agent, remote-job-spawner,
--                               thunder-adapter, vet-intel, web-scrape-adapter,
--                               web-search-adapter)
--   Automated Work Items Completed = count(site_work_items complete/verified)   [1,264 at apply time]
-- First three computed by subquery where the DB can count; the service count is
-- a manifest fact with its basis stated here. Footnote carries the as-of date.
--
-- Apply:  kubectl -n ai-persona-system exec -i postgres-clients-0 -- \
--           psql -U clients_user -d clients_db < THIS_FILE

\set site '2a8ebf9c-20a2-4c39-b191-840b012371da'
\set comp 'fdd92ad4-521a-4602-89cf-7ee1a66c10f1'

BEGIN;

UPDATE page_components pc
   SET content_data = content_data || jsonb_build_object(
     'footnote_text', 'Figures are live counts from the platform''s own database — agent registry, site register and completed work-item ledger — taken 2026-07-24. Architecture runs on Kubernetes, Kafka, and Postgres.',
     'stat1_label', 'Deployed Agents',
     'stat1_value', (SELECT count(*)::text FROM agent_definitions
                     WHERE is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL),
     'stat1_suffix', '',
     'stat1_description', 'Active agent definitions in the production registry — each independently configured, versioned, and running with observable state.',
     'stat2_label', 'Live Sites in Production',
     'stat2_value', (SELECT count(*)::text FROM sites
                     WHERE status NOT IN ('archived','deleted','pool')),
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
   ),
   updated_at = now()
 WHERE pc.page_id = (SELECT id FROM pages WHERE site_id = :'site'
                     AND name='case-study-kafka-consumer-group-remediation')
   AND pc.component_id = :'comp';

\echo '--- after ---'
SELECT e.k, left(e.v,60) FROM page_components pc
JOIN pages p ON p.id=pc.page_id
CROSS JOIN LATERAL jsonb_each_text(pc.content_data) e(k,v)
WHERE p.site_id = :'site' AND p.name='case-study-kafka-consumer-group-remediation'
  AND pc.component_id = :'comp'
  AND e.k ~ 'stat[0-9]_(label|value|suffix)' ORDER BY e.k;

INSERT INTO site_work_items
  (site_id, source, item_type, severity, summary, status, created_by, pipeline,
   priority, triaged_at, handler_agent, item_key, spec)
SELECT p.site_id, 'fabricated-stats-043-aiagent', 'page_rerender', 'medium',
  'Rerender case study — fabricated production metrics replaced with live platform counts (043)',
  'triaged', 'session-2026-07-24-043-treatment', 'build',
  20, now(), 'page-rerender',
  'page_rerender_' || p.name || '_043stats_' || p.site_id::text,
  jsonb_build_object('domain','ai-agent-orchestration.com','reason','cta_links_stale',
                     'page_id',p.id,'page_name',p.name,'filename',ltrim(p.url,'/'))
FROM pages p WHERE p.site_id = :'site' AND p.name='case-study-kafka-consumer-group-remediation';

\echo '--- queued ---'
SELECT status, count(*) FROM site_work_items WHERE source='fabricated-stats-043-aiagent' GROUP BY 1;

COMMIT;
