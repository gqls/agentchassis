-- Phase 1 — resolve tool-audience-check per owner decision (2026-07-15):
-- retarget the tool-list card straight at the live tool (/audience-check) and retire the
-- empty static page. No interstitial.
--
-- MECHANISM (traced, not guessed):
--   * The "Free Audience Check" card is rendered by the tool-list component via
--     source "query.pages_where_type:tool" → resolvePagesWhereType (queryresolve.go:155),
--     which SELECTs pages WHERE page_type='tool' AND status IN ('active','deployed') and uses
--     each page's `url` as the card href. So the card href IS tool-audience-check.url.
--   * Repointing that url to /audience-check makes the card link to the live tool. Post-cutover
--     nginx proxies /audience-check to the VM binary (exact-match location wins over any static
--     file), so the card lands on the real tool.
--   * "Retire the static page" = it must never emit a static artefact, and reconcile must stop
--     queueing it. decideEmit (reconcile_site_plan_action.go:293) returns skip_built ONLY when
--     build_status='deployed' AND built_from_plan_version = the current plan id. So we mark this
--     row as a POINTER: build_status='deployed' (truthful — the tool it points to IS deployed,
--     on the box) pinned to the current plan. It has no sections, so nothing assembles a file.
--
-- This is the general shape for a VM-hosted tool surfaced in a chassis tool-list: a page_type
-- 'tool' row acting as a catalogue pointer (url = the tool's live path, no static artefact).

\set ON_ERROR_STOP on
\set SID '1244516d-014d-421c-88c6-090bb1e9552a'
\set PLAN 'ff03bdef-3bb2-40eb-93ff-efa70f46b6b8'
BEGIN;

-- 1. Repoint + mark as pointer (skip_built forever).
UPDATE pages
SET url = '/audience-check',
    build_status = 'deployed',
    built_from_plan_version = :'PLAN',
    updated_at = NOW()
WHERE site_id = :'SID' AND name = 'tool-audience-check';

-- 2. Keep the plan page url consistent (audits compare plan vs realised url).
UPDATE site_plan_pages
SET url = '/audience-check'
WHERE plan_id = :'PLAN' AND name = 'tool-audience-check';

-- 3. Retire the paused build item — the page is intentionally never built.
UPDATE site_work_items
SET status = 'wont_fix', updated_at = NOW()
WHERE site_id = :'SID' AND item_type = 'needs_page'
  AND spec->>'page_name' = 'tool-audience-check'
  AND status NOT IN ('complete','verified','rejected','wont_fix');

-- 4. Re-render the two pages that carry the tool-list card, so the baked-in href updates
--    from /tools/audience-check/index.html to /audience-check. (Their current deployed HTML
--    still has the old link.) Emit fresh needs_page for index + tools.
INSERT INTO site_work_items
  (site_id, source, item_type, severity, summary, spec, priority, handler_agent, status, created_by, item_key, pipeline, approval_mode)
SELECT :'SID', 'manual-repair', 'needs_page', 'high',
       'Re-render ' || v.page || ' to update tool-list card link → /audience-check',
       jsonb_build_object('reason','rebuild','plan_id', :'PLAN','page_name', v.page,'page_role', v.role),
       55, 'page-build-handler', 'triaged', 'manual-repair',
       'needs_page:' || v.page, 'build', 'auto'
FROM (VALUES ('index','landing'), ('tools','content')) AS v(page, role);

COMMIT;

-- ── VERIFY ───────────────────────────────────────────────────────────────────
\echo '=== tool-audience-check is now a pointer (url=/audience-check, deployed, pinned) ==='
SELECT name, page_type, status, build_status, url,
       (built_from_plan_version = :'PLAN') AS pinned_to_current_plan,
       jsonb_array_length(COALESCE(sections,'[]'::jsonb)) AS n_sections
FROM pages WHERE site_id = :'SID' AND name = 'tool-audience-check';

\echo '=== index + tools re-render queued ==='
SELECT spec->>'page_name' AS page, status FROM site_work_items
WHERE site_id = :'SID' AND source='manual-repair' AND spec->>'page_name' IN ('index','tools') ORDER BY 1;
