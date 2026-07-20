-- R4d (2026-07-20) — queue re-renders so R4/R4b/R4c reach the live pages.
--
-- The three preceding files changed `page_components.content_data`, which is the
-- durable source but is INERT until each page re-renders. Twelve pages carry a
-- changed CTA, a corrected statistic, or both.
--
-- Reason is `cta_links_stale` deliberately. Per /bugs_open/024 the rerender
-- action picks its branch from the reason: an item arriving with no usable reason
-- falls to stale-HTML assembly and the page never actually changes while every
-- status still reports green. `cta_links_stale` is one of the reasons that
-- reaches the real `rerender_sections` branch.
--
-- Priority 20 + triaged_at set, per the RUNBOOK: `build-pipeline-trigger` runs
-- every 30s but processes ONE item per site at a time and skips a site that has
-- anything `claimed`, so a batch lands behind whatever churn is already queued.
-- Priority is ASC — lower runs sooner. robot-hands carries a large inherited
-- backlog (115 unresolved at the last count), so an unpromoted batch would take
-- hours.
--
-- Apply:  kubectl -n ai-persona-system exec -i postgres-clients-0 -- \
--           psql -U clients_user -d clients_db < THIS_FILE

\set site '00ff3af5-dad8-4770-9f70-3edc267a3c92'

BEGIN;

INSERT INTO site_work_items
  (site_id, source, item_type, severity, summary, status, created_by, pipeline,
   priority, triaged_at, spec)
SELECT
  p.site_id,
  'robot-hands-r4-cta-pairing',
  'page_rerender',
  'medium',
  'Rerender ' || p.name || ' — CTA label/URL pairing corrected and unsourced statistics replaced (R4/R4b/R4c)',
  'triaged',
  'session-2026-07-20-robot-hands',
  'build',
  20,
  now(),
  jsonb_build_object(
    'domain',    'robot-hands.com',
    'reason',    'cta_links_stale',
    'page_id',   p.id,
    'page_name', p.name,
    'filename',  ltrim(p.url, '/')
  )
FROM pages p
WHERE p.site_id = :'site'
  AND p.name IN (
    'about', 'gripper-detail', 'index',
    'gripper-payload-calculator', 'gripper-selection-guide',
    'how-it-works', 'how-to-specify-a-gripper', 'learning-center-hub',
    'matchmatrix', 'matchmatrix-methodology', 'product-detail', 'services'
  );

\echo '--- queued ---'
SELECT status, priority, count(*)
FROM site_work_items
WHERE site_id = :'site' AND source = 'robot-hands-r4-cta-pairing'
GROUP BY 1,2;

COMMIT;

-- ---------------------------------------------------------------------------
-- CORRECTION, same session, ~15 min later. The 12 items above sat at `triaged`
-- and were NEVER claimed, while the pipeline was demonstrably alive (6 unrelated
-- items completed in that window) and they were the ONLY `triaged` items
-- fleet-wide. The INSERT was modelled on an existing row's *visible* fields and
-- I missed two that the dispatch loop needs:
--
--   handler_agent  must be 'page-rerender'  — I left it NULL, so the dispatch
--                  loop had no handler to route to and silently skipped them
--   item_key       the dedup key; NULL is accepted but leaves the row outside
--                  idx_swi_dedup
--
-- Nothing errored and nothing warned: an unroutable item just sits at `triaged`
-- looking queued. That is indistinguishable from "waiting behind the backlog",
-- which is exactly what I assumed for the first ten minutes.
-- ---------------------------------------------------------------------------
UPDATE site_work_items
   SET handler_agent = 'page-rerender',
       item_key = 'page_rerender_' || (spec->>'page_name') || '_r4cta_' || site_id::text,
       updated_at = now()
 WHERE source = 'robot-hands-r4-cta-pairing'
   AND handler_agent IS NULL;

\echo '--- after the correction ---'
SELECT status, handler_agent, priority, count(*)
FROM site_work_items WHERE source = 'robot-hands-r4-cta-pairing'
GROUP BY 1,2,3;
