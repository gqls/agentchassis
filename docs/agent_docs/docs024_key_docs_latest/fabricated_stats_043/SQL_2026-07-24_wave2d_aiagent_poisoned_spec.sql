-- 043 wave-2d (2026-07-24) — the POISONED GIVEN: ai-agent-orchestration's own
-- site specs command the fabricated numbers, and a spec instruction beats every
-- downstream rule.
--
-- OBSERVED LIVE (the same afternoon migration 201 + evidence_base shipped): a
-- full-writer re-render of aiagent/index at 15:10 rewrote the system-stats
-- block BACK to "70+ / 8 / 30+ / 1000s" — overwriting the grounded values set
-- 45 minutes earlier. Root cause: the site's content_direction (and site_plan,
-- briefing) contain, verbatim, "(2) the system-stats component showing real
-- numbers — 70+ agents, 8 departments, 30+ agent types, thousands of
-- concurrent instances". The prompt renders content_direction under "follow
-- this closely" ABOVE the Verified Facts block, so the writer was ORDERED to
-- write those figures and told they are real. Rule 14 v2 permits given figures;
-- these were given — by a spec. **Candidate 3 cannot beat a poisoned spec: the
-- givens themselves must trace.** (Fourth mechanism for the 043 file, after the
-- schema fallback, the required-field conflict, and the example-shape seeding.)
--
-- PROVENANCE + TRUTH AUDIT of the spec claims (specs are agent-generated —
-- build-briefing-agent / build-site-planner / content-gap-planner, Apr-May):
--   "over 70 specialised AI agents"      TRUE, conservative (registry: 170)
--   "8 departments — Strategy, Research, Content, Design, Development,
--    Quality, Operations, Data"          TRUE as the platform's own taxonomy
--   "30+ agent types"                    TRUE, conservative (165 distinct)
--   "thousands of concurrent agent instances"  UNTRUE — nothing measures it.
--    The honest neighbour: 1,699 orchestrations in the 24h to 2026-07-24, so
--    "over a thousand orchestrations a day" is simply true.
-- So the generator embellished ONE clause past the truth, and the stats
-- instruction froze all four into copy law. My earlier evidence_base for this
-- site over-corrected the other way (banned "departments" outright — wrong: the
-- departments are a real self-taxonomy, not a fabricated client count).
--
-- FIX (all site_specs changes are versioned supersede+insert, never in-place):
--  1. briefing/identity: "managing thousands of concurrent agent instances in
--     production" -> "processing over a thousand orchestrations a day in
--     production" (true; 1,699/24h at audit).
--  2. content_direction/site_plan/briefing: the hardcoded stats instruction ->
--     "showing real registry counts, taken only from the Verified Facts list
--     for this site — never approximated, never invented".
--  3. evidence_base writer_block: refreshed — adds the 8-department taxonomy
--     (with its meaning), 160+ agent types, the orchestration-volume figure;
--     narrows the departments ban to the misframing ("departments SERVED").
--  4. index system-stats restored: 170 / 8 departments / 13 live sites /
--     1,267 work items (owner positioning kept where true, grounded elsewhere).
--  5. index re-render.
--
-- Apply:  kubectl -n ai-persona-system exec -i postgres-clients-0 -- \
--           psql -U clients_user -d clients_db < THIS_FILE

\set aio '2a8ebf9c-20a2-4c39-b191-840b012371da'

BEGIN;

-- ── 1+2. Patch the four aspects (versioned: supersede + insert patched copy) ──
CREATE TEMP TABLE _patched ON COMMIT DROP AS
SELECT aspect,
       replace(replace(
         data::text,
         'managing thousands of concurrent agent instances in production',
         'processing over a thousand orchestrations a day in production'
       ),
         'the system-stats component showing real numbers — 70+ agents, 8 departments, 30+ agent types, thousands of concurrent instances',
         'the system-stats component showing real registry counts, taken only from the Verified Facts list for this site — never approximated, never invented'
       )::jsonb AS data
FROM site_specs
WHERE site_id = :'aio' AND is_current
  AND aspect IN ('briefing','identity','content_direction','site_plan');

UPDATE site_specs SET is_current=false, superseded_at=now()
 WHERE site_id = :'aio' AND is_current
   AND aspect IN ('briefing','identity','content_direction','site_plan');

INSERT INTO site_specs (site_id, aspect, data, source, created_by, notes)
SELECT :'aio', aspect, data, 'manual', 'session-2026-07-24-043-treatment',
       'bug 043 wave-2d: de-poison the spec — untrue "thousands of concurrent instances" clause corrected to the true orchestration volume; hardcoded system-stats figure list repointed at Verified Facts. Prior version preserved (superseded).'
FROM _patched;

\echo '--- specs after patch: poisoned strings must be 0 ---'
SELECT aspect,
       (data::text LIKE '%thousands of concurrent%') AS still_poisoned,
       (data::text LIKE '%Verified Facts list for this site%') AS repointed
FROM site_specs
WHERE site_id = :'aio' AND is_current
  AND aspect IN ('briefing','identity','content_direction','site_plan')
ORDER BY aspect;

-- ── 3. evidence_base refresh (supersede + insert) ────────────────────────────
UPDATE site_specs SET is_current=false, superseded_at=now()
 WHERE site_id = :'aio' AND is_current AND aspect='evidence_base';

INSERT INTO site_specs (site_id, aspect, data, source, created_by, notes)
VALUES (:'aio', 'evidence_base', jsonb_build_object('writer_block',
$ao2$NUMBERS (state only these, with their listed meaning; dated snapshots up to a listed live count are fine — these are live registry counts taken 2026-07-24, recount before restating):
- more than 150 active agent definitions in the production registry (170 as of 2026-07-24); "over 70 specialised AI agents" is also true and may stand where already written
- 8 departments — Strategy, Research, Content, Design, Development, Quality, Operations, Data. This is the platform's OWN organisational taxonomy (a self-description); never frame it as departments of external clients ("departments served")
- more than 150 distinct agent types (165 as of 2026-07-24); "30+ agent types" is also true and may stand
- 13 live sites in production, built and operated end-to-end by the platform
- 17 backend services (the service manifests under deployments/kustomize/services)
- 1,267 automated work items completed (the platform's work-item ledger)
- over a thousand orchestrations a day (1,699 in the 24 hours to 2026-07-24)
- Architecture: Kubernetes, Kafka, Postgres — true and stated freely
NOT TRACKED / DOES NOT EXIST, NEVER STATE: clients served, "departments served", satisfaction rates, awards won, concurrent-instance counts ("thousands of concurrent instances" is not measured), uptime percentages. None of these are measured; every such figure at any value is an invention.$ao2$),
 'manual', 'session-2026-07-24-043-treatment',
 'bug 043 wave-2d: refreshed — adds the true taxonomy + volume figures the spec legitimately positions on; narrows the departments ban to the misframing');

-- ── 4. Restore the index stat block (owner positioning where true, grounded) ──
UPDATE page_components pc
   SET content_data = content_data || jsonb_build_object(
     'footnote_text', 'Figures are live counts from the platform''s own database — agent registry, site register and completed work-item ledger — taken 2026-07-24. Architecture runs on Kubernetes, Kafka, and Postgres.',
     'stat1_label', 'Deployed Agents',
     'stat1_value', (SELECT count(*)::text FROM agent_definitions
                     WHERE is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL),
     'stat1_suffix', '',
     'stat1_description', 'Active agent definitions in the production registry — each independently configured, versioned, and running with observable state.',
     'stat2_label', 'Departments',
     'stat2_value', '8',
     'stat2_suffix', '',
     'stat2_description', 'Strategy, Research, Content, Design, Development, Quality, Operations and Data — the platform''s own organisational structure, each led by a managing agent coordinating specialists.',
     'stat3_label', 'Live Sites in Production',
     'stat3_value', (SELECT count(*)::text FROM sites WHERE status NOT IN ('archived','deleted','pool')),
     'stat3_suffix', '',
     'stat3_description', 'Distinct production websites the platform builds and operates end-to-end — content, imagery, interactive tooling and deployment.',
     'stat4_label', 'Automated Work Items Completed',
     'stat4_value', (SELECT to_char(count(*), 'FM9,999,999') FROM site_work_items
                     WHERE status IN ('complete','verified')),
     'stat4_suffix', '',
     'stat4_description', 'Discrete pieces of site work — builds, re-renders, repairs, audits — detected, dispatched and completed by the platform''s own work-item pipeline.'
   ), updated_at = now()
 WHERE pc.page_id = (SELECT id FROM pages WHERE site_id = :'aio' AND name='index')
   AND pc.component_id = 'fdd92ad4-521a-4602-89cf-7ee1a66c10f1';

\echo '--- index stats after restore ---'
SELECT e.k, left(e.v,40) FROM page_components pc JOIN pages p ON p.id=pc.page_id
CROSS JOIN LATERAL jsonb_each_text(pc.content_data) e(k,v)
WHERE p.site_id = :'aio' AND p.name='index' AND pc.component_id='fdd92ad4-521a-4602-89cf-7ee1a66c10f1'
  AND e.k ~ 'stat[0-9]_(label|value)' ORDER BY e.k;

-- ── 5. Re-render ─────────────────────────────────────────────────────────────
INSERT INTO site_work_items
  (site_id, source, item_type, severity, summary, status, created_by, pipeline,
   priority, triaged_at, handler_agent, item_key, spec)
SELECT p.site_id, 'fabricated-stats-043-wave2d', 'page_rerender', 'medium',
  'Rerender index — spec de-poisoned, grounded stats restored (043 wave-2d)',
  'triaged', 'session-2026-07-24-043-treatment', 'build',
  20, now(), 'page-rerender',
  'page_rerender_' || p.name || '_043w2d_' || p.site_id::text,
  jsonb_build_object('domain','ai-agent-orchestration.com','reason','cta_links_stale',
                     'page_id',p.id,'page_name',p.name,'filename',ltrim(p.url,'/'))
FROM pages p WHERE p.site_id = :'aio' AND p.name='index';

\echo '--- queued ---'
SELECT status, count(*) FROM site_work_items WHERE source='fabricated-stats-043-wave2d' GROUP BY 1;

COMMIT;
