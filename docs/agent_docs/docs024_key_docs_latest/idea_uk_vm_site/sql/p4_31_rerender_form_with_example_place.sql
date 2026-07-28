-- p4_31 — publish p4_30 (the £8 example place on the request form).
-- Template edit, so reason=section_data_resolved; the assemble-only path cannot
-- apply it. Same required-field guard as p4_27 (RUNBOOK TRAP 1b).
\set ON_ERROR_STOP on
BEGIN;
DO $g$
DECLARE missing text;
BEGIN
  SELECT string_agg(cc.function || '.' || f.key, ', ') INTO missing
  FROM page_components pc
  JOIN content_components cc ON cc.id = pc.component_id
  CROSS JOIN LATERAL jsonb_each(cc.input_schema->'fields') AS f(key,val)
  WHERE pc.page_id = '41333d74-0c5a-4e12-b942-50ba4df793e6'
    AND (f.val->>'required')::boolean IS TRUE AND NOT (pc.content_data ? f.key);
  IF missing IS NOT NULL THEN
    RAISE EXCEPTION 'ABORT: required field(s) missing: %', missing;
  END IF;
END $g$;
INSERT INTO site_work_items
  (site_id, source, item_type, severity, summary, spec, priority, handler_agent, status, created_by, item_key)
SELECT p.site_id, 'manual-p4_31', 'page_rerender', 'medium',
  'Re-render /report.html to publish the £8 example place on the form',
  jsonb_build_object('domain','idea.uk','page_id',p.id::text,'filename','report.html',
                     'page_name',p.name,'reason','section_data_resolved'),
  85, 'page-rerender', 'triaged', 'idea.uk vm 7 session, p4_31',
  'page_rerender_report_p4_31_' || p.site_id::text
FROM pages p WHERE p.id = '41333d74-0c5a-4e12-b942-50ba4df793e6';
COMMIT;
SELECT id, status FROM site_work_items WHERE created_by LIKE '%p4_31%';
