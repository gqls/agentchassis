-- p4_28 — two finishing steps for the specimen work.
--
-- (a) Backfill pages.sections for the new page. The rerender path does NOT write
--     it (RUNBOOK Phase 5 step 5), and a page with sections='[]' is invisible to
--     ListedPageEligibilitySQL and to the imagery sweep — it renders and serves
--     fine, so nothing complains, and it is quietly missing from every derivation.
--     Derived from page_components in position order so it cannot drift from what
--     is actually on the page.
--
-- (b) Publish p4_26's card changes to /report.html. content_data edit, so
--     reason=section_data_resolved; the assemble-only path could not apply it.
\set ON_ERROR_STOP on
\set site '1244516d-014d-421c-88c6-090bb1e9552a'

BEGIN;

UPDATE pages p
SET sections = (
      SELECT coalesce(jsonb_agg(pc.slot_name ORDER BY pc.position), '[]'::jsonb)
      FROM page_components pc WHERE pc.page_id = p.id
    ),
    updated_at = now()
WHERE p.site_id = :'site' AND p.name = 'report-example';

DO $g$
DECLARE n int;
BEGIN
  SELECT count(*) INTO n FROM page_components pc
  JOIN pages p ON p.id = pc.page_id
  WHERE p.site_id = '1244516d-014d-421c-88c6-090bb1e9552a' AND p.name = 'report'
    AND (pc.content_data IS NULL OR pc.content_data = '{}'::jsonb);
  IF n > 0 THEN RAISE EXCEPTION 'ABORT: % section(s) on /report.html lack content_data', n; END IF;
END $g$;

INSERT INTO site_work_items
  (site_id, source, item_type, severity, summary, spec, priority, handler_agent, status, created_by, item_key)
SELECT p.site_id, 'manual-p4_28', 'page_rerender', 'medium',
  'Re-render /report.html to publish p4_26 (cards now point at the specimen, not at themselves)',
  jsonb_build_object('domain','idea.uk','page_id',p.id::text,'filename','report.html',
                     'page_name',p.name,'reason','section_data_resolved'),
  85, 'page-rerender', 'triaged', 'idea.uk vm 7 session, p4_28',
  'page_rerender_report_p4_28_' || p.site_id::text
FROM pages p WHERE p.site_id = :'site' AND p.name = 'report';

COMMIT;

SELECT name, jsonb_pretty(sections) AS sections FROM pages
WHERE site_id = :'site' AND name = 'report-example';
