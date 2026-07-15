-- Phase 1 — rebuild about + contact from the RESTORED composition.
--
-- page_components (the actual rendered artifact) proved both pages deployed REGRESSED:
--   about   → generic `hero` (not hero-about), missing info-card-grid   [deployed 16:07]
--   contact → generic `hero` (not hero-contact), missing contact-info    [deployed 16:54]
-- The rollback restored pages.sections + the plan sections, but neither page rebuilt because
-- its needs_page item was already terminal ('complete'), so the re-triage had nothing to catch.
-- Emit fresh build items so they re-render from the restored composition.
--
-- Shape copied from the live reconcile_site_plan needs_page items (handler page-build-handler,
-- status triaged so the dispatch loop claims it). item_key needs_page:<page> is safe: the
-- dedup index only blocks a second NON-terminal row, and the existing rows are all complete.
--
-- NOTE (separate, pre-existing): contact-info escalates needs_section_data (needs_human_review)
-- because it wants a business contact email. That gap predates this incident; the rebuild will
-- restore hero-contact regardless, and the contact-info email is an owner decision tracked apart.

\set ON_ERROR_STOP on
\set SID '1244516d-014d-421c-88c6-090bb1e9552a'
\set PLAN 'ff03bdef-3bb2-40eb-93ff-efa70f46b6b8'

INSERT INTO site_work_items
  (site_id, source, item_type, severity, summary, spec, priority, handler_agent, status, created_by, item_key, pipeline, approval_mode)
SELECT s.id, 'manual-repair', 'needs_page', 'high',
       'Rebuild ' || v.page || ' from restored composition (regression repair)',
       jsonb_build_object('reason','rebuild','plan_id', :'PLAN','page_name', v.page,'page_role', v.role),
       52, 'page-build-handler', 'triaged', 'manual-repair',
       'needs_page:' || v.page, 'build', 'auto'
FROM sites s, (VALUES ('about','content'), ('contact','content')) AS v(page, role)
WHERE s.domain = 'idea.uk';

\echo '=== emitted ==='
SELECT spec->>'page_name' AS page, status, handler_agent, priority, item_key
FROM site_work_items
WHERE site_id = :'SID' AND item_type='needs_page' AND source='manual-repair'
ORDER BY page;
