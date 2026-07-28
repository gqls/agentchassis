-- p4_25 — build /report/example/index.html (the specimen page created by p4_24).
-- reason=section_data_resolved: a NEW page has no rendered_html, so the
-- assemble-only path would produce nothing and report success.
\set ON_ERROR_STOP on
BEGIN;
DO $g$
DECLARE bad int;
BEGIN
  SELECT count(*) INTO bad FROM page_components pc
  JOIN pages p ON p.id=pc.page_id
  WHERE p.site_id='1244516d-014d-421c-88c6-090bb1e9552a' AND p.name='report-example'
    AND (pc.slot_name IS NULL OR pc.content_data IS NULL OR pc.content_data='{}'::jsonb);
  IF bad > 0 THEN
    RAISE EXCEPTION 'ABORT: % section(s) have NULL slot_name or empty content_data — the render would produce an empty page and report COMPLETED', bad;
  END IF;
END $g$;
INSERT INTO site_work_items
  (site_id, source, item_type, severity, summary, spec, priority, handler_agent, status, created_by, item_key)
SELECT p.site_id, 'manual-p4_25', 'page_rerender', 'medium',
  'Build /report/example/index.html — the specimen report page',
  jsonb_build_object('domain','idea.uk','page_id',p.id::text,
                     'filename','report/example/index.html',
                     'page_name',p.name,'reason','section_data_resolved'),
  85, 'page-rerender', 'triaged', 'idea.uk vm 7 session, p4_25',
  'page_rerender_report_example_p4_25_' || p.site_id::text
FROM pages p WHERE p.site_id='1244516d-014d-421c-88c6-090bb1e9552a' AND p.name='report-example';
COMMIT;
SELECT id, spec->>'filename' AS file, spec->>'reason' AS reason, status FROM site_work_items WHERE created_by LIKE '%p4_25%';
