-- Phase 1 — emit a needs_site_plan for idea.uk so build-site-planner composes the three
-- catalogued-but-uncomposed pages (guides-index, news-index, tool-audience-check).
--
-- WHY THIS IS SAFE (verified in code, not assumed):
--   * validate_plan's Pass A (v3_site_actions.go:4630-4652) UNIONS realised pages that the
--     LLM did not re-propose, and normaliseRealisedToPlanPage (:4458) CARRIES THEIR sections
--     — its own comment says that without this the sync's "<col> = EXCLUDED.<col>" would
--     clobber the adopted page's sections/meta_description/nav_order. So the 6 built pages
--     keep their composition.
--   * max_pages = 80 in the deployed planner config, against 9 pages — truncation cannot
--     drop a realised page.
--   * idea.uk's build currently deploys to B2, which DNS does not point at (the VM serves
--     the live site). A re-plan therefore cannot touch the live site. This is the right
--     moment to do it — BEFORE the nginx cutover.
--
-- SHAPE: copied from the historical row that worked (created 2026-06-21 by
-- build-briefing-agent via the create_work_item action, ran to status='complete').
-- create_work_item_action.go:113-117 defaults status to 'triaged', which is what the
-- dispatch loop claims (claim_work_item_action.go:102 — status IN ('triaged','approved')).
--
-- DOUBLE-EMIT GUARD: idx_swi_dedup is UNIQUE (site_id, item_key) WHERE the status is
-- non-terminal. Reusing item_key='site_plan_idea.uk' therefore makes a second run of this
-- file fail loudly rather than queue two re-plans. The 2026-06-21 row is 'complete'
-- (terminal), so it does not block this insert.

\set ON_ERROR_STOP on

-- ── BEFORE ───────────────────────────────────────────────────────────────────
\echo '=== BEFORE: pages (3 planned with 0 sections are the targets) ==='
SELECT p.url, p.name, p.page_type, p.build_status,
       jsonb_array_length(COALESCE(p.sections, '[]'::jsonb)) AS n_sections
FROM pages p
WHERE p.site_id = (SELECT id FROM sites WHERE domain = 'idea.uk')
ORDER BY p.nav_order, p.url;

\echo '=== BEFORE: any OPEN work item that would block this emit ==='
SELECT wi.item_type, wi.status, wi.item_key, wi.created_at
FROM site_work_items wi
WHERE wi.site_id = (SELECT id FROM sites WHERE domain = 'idea.uk')
  AND wi.item_key = 'site_plan_idea.uk'
  AND wi.status <> ALL (ARRAY['complete','verified','rejected','wont_fix','failed','unresolved']);

-- ── EMIT ─────────────────────────────────────────────────────────────────────
INSERT INTO site_work_items (
    site_id, source, item_type, severity, summary, spec,
    priority, handler_agent, status, created_by, item_key, pipeline, approval_mode
)
SELECT
    s.id,
    'manual-replan',
    'needs_site_plan',
    'high',
    'Re-plan: compose the 3 catalogued pages (guides-index, news-index, tool-audience-check)',
    '{}'::jsonb,
    15,
    'build-site-planner',
    'triaged',
    'manual-replan',
    'site_plan_idea.uk',
    'build',
    'auto'
FROM sites s
WHERE s.domain = 'idea.uk';

-- ── AFTER ────────────────────────────────────────────────────────────────────
\echo '=== AFTER: the emitted item ==='
SELECT wi.id, wi.item_type, wi.status, wi.handler_agent, wi.pipeline,
       wi.priority, wi.item_key, wi.created_at
FROM site_work_items wi
WHERE wi.site_id = (SELECT id FROM sites WHERE domain = 'idea.uk')
  AND wi.item_key = 'site_plan_idea.uk'
ORDER BY wi.created_at DESC
LIMIT 2;
