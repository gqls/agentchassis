-- p4_23 — rerender /report.html so p4_22's maxlength attributes reach the page.
-- A template edit CANNOT arrive via the assemble-only path; reason must be
-- section_data_resolved. Same guards as p4_21: no NULL content_data (else the
-- rerender escalates to the LLM writer and rewrites live sales copy).
\set ON_ERROR_STOP on
BEGIN;
DO $g$
DECLARE n int;
BEGIN
  SELECT count(*) INTO n FROM page_components
  WHERE page_id='41333d74-0c5a-4e12-b942-50ba4df793e6'
    AND (content_data IS NULL OR content_data='{}'::jsonb);
  IF n > 0 THEN RAISE EXCEPTION 'ABORT: % section(s) lack content_data', n; END IF;
END $g$;
INSERT INTO site_work_items
  (site_id, source, item_type, severity, summary, spec, priority, handler_agent, status, created_by, item_key)
SELECT p.site_id, 'manual-p4_23', 'page_rerender', 'medium',
  'Re-render /report.html to pick up p4_22 (maxlength on the request form fields)',
  jsonb_build_object('domain','idea.uk','page_id',p.id::text,'filename','report.html',
                     'page_name',p.name,'reason','section_data_resolved'),
  85, 'page-rerender', 'triaged', 'idea.uk vm 7 session, p4_23',
  'page_rerender_report_p4_23_' || p.site_id::text
FROM pages p WHERE p.id='41333d74-0c5a-4e12-b942-50ba4df793e6';
COMMIT;
SELECT id, spec->>'reason' AS reason, status FROM site_work_items WHERE created_by LIKE '%p4_23%';
