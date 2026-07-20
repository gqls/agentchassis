-- p3_04_force_section_rerender.sql — idea.uk: actually get p3_03's template into the pages.
--
-- WHY THIS IS NEEDED (and why the previous run looked like it worked).
-- The rerender fired after p3_03 completed all 9 page_rerender items with status='complete',
-- deployed both pages, and even published /tools/assets/audience-check-form.js (HTTP 200,
-- 1469B — collectJSAssets reads content_components.js_content directly, so the ASSET landed).
-- But the page markup did not change: no <script src=…> ref, no #ac-result div, and the tool
-- cards still pointed at /audience-check.
--
-- Cause, from rerender_page_sections_action.go:47-51:
--     check_rerender_mode (conditional: reason==image_landed OR reason==section_data_resolved)
--       -> rerender_sections -> check_escalated -> save_sections -> render_page
--     else_step (no/other reason) -> render_page   (unchanged ASSEMBLE-ONLY path)
-- The items rerender-pages created carry NO spec.reason, so every one took the assemble-only
-- path: pages were rebuilt from the STORED page_components.rendered_html (dated 18/19 July)
-- and deployed unchanged. A template edit cannot reach a page down that path.
--
-- This inserts the same items WITH reason='section_data_resolved', which routes them through
-- rerender_page_sections — "re-render ALL of a page's sections from their STORED content_data
-- plus FRESHLY re-resolved dynamic fields, WITHOUT invoking the content writer (no LLM)".
-- That is exactly what is wanted: the copy is unchanged, only the template and one
-- query-resolved URL (tool-list.items, source query.pages_where_type:tool) need refreshing.
--
-- SCOPE: tools + index only. p3_03 touched audience-check-form (tools) and tool-list
-- (tools + index). The other 7 pages are unaffected — re-rendering them would be waste.
--
-- ⚠️ ESCALATION RISK, stated up front: if ANY section on these pages has NULL content_data,
-- rerender_page_sections escalates the WHOLE page to the content generator (needs_page ->
-- page-build-handler), which REGENERATES COPY VIA LLM. Checked before writing this: see the
-- guard below, which refuses rather than risk an unwanted rewrite of live sales copy.

\set ON_ERROR_STOP on

BEGIN;

-- Refuse if any section on either page lacks content_data — that would escalate to the
-- LLM content writer and rewrite live copy, which is not what this change is for.
DO $guard$
DECLARE
  n int;
BEGIN
  SELECT count(*) INTO n
  FROM page_components pc JOIN pages p ON p.id = pc.page_id
  WHERE p.site_id = '1244516d-014d-421c-88c6-090bb1e9552a'
    AND p.name IN ('tools','index')
    AND (pc.content_data IS NULL OR pc.content_data = '{}'::jsonb);
  IF n > 0 THEN
    RAISE EXCEPTION
      'ABORT: % section(s) on tools/index have NULL/empty content_data; rerender_page_sections would escalate the whole page to the LLM content writer and rewrite live copy. Investigate before forcing.', n;
  END IF;
END
$guard$;

INSERT INTO site_work_items
  (site_id, source, item_type, severity, summary, spec, priority, handler_agent, status, created_by, item_key)
SELECT
  p.site_id,
  'manual-p3_04',
  'page_rerender',
  'medium',
  'Re-render sections on ' || p.name || ' to pick up p3_03 (taster form AJAX + tool card URL)',
  jsonb_build_object(
    'domain',    'idea.uk',
    'page_id',   p.id::text,
    'filename',  p.name || '.html',
    'page_name', p.name,
    'reason',    'section_data_resolved'
  ),
  80,
  'page-rerender',
  'triaged',
  'idea.uk vm site 3 session, p3_04',
  'page_rerender_' || p.name || '_' || p.site_id::text
FROM pages p
WHERE p.site_id = '1244516d-014d-421c-88c6-090bb1e9552a'
  AND p.name IN ('tools','index');

COMMIT;

SELECT spec->>'page_name' AS page, spec->>'reason' AS reason, status, created_at
FROM site_work_items
WHERE site_id='1244516d-014d-421c-88c6-090bb1e9552a' AND item_type='page_rerender'
  AND created_by LIKE '%p3_04%' ORDER BY page;
