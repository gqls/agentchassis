-- p4_27 — supply the generic-text-block's REQUIRED `heading`, then rebuild.
--
-- p4_25's build reported COMPLETED and produced nothing. Cause, from
-- rerender_page_sections_action.go:273 — missingRequiredLLMFields() found the
-- text block's required `heading` absent, so the action ESCALATED the page to
-- the content writer rather than rendering a blank one. That is the right
-- behaviour; the mistake was mine, for creating sections without checking which
-- schema fields are required.
--
-- The escalation raised a needs_page item for page-build-handler, which
-- regenerates copy via LLM. CANCELLED before it ran: this page reproduces a real
-- report verbatim and publishes the claim that nothing was reworded, so letting a
-- writer near it would have made that claim false silently.
--
-- RUNBOOK Phase 5 documents the slot_name trap but not this one. Added there.
\set ON_ERROR_STOP on
BEGIN;
UPDATE page_components pc
SET content_data = jsonb_set(pc.content_data, '{heading}',
                             '"The report, exactly as it was sent"'::jsonb, true),
    updated_at = now()
FROM pages p
WHERE p.id = pc.page_id AND p.site_id='1244516d-014d-421c-88c6-090bb1e9552a'
  AND p.name='report-example' AND pc.position=2;

-- Refuse to dispatch if ANY required LLM field is still missing on ANY section —
-- otherwise this repeats the same silent escalation.
DO $g$
DECLARE missing text;
BEGIN
  SELECT string_agg(cc.function || '.' || f.key, ', ') INTO missing
  FROM page_components pc
  JOIN pages p ON p.id = pc.page_id
  JOIN content_components cc ON cc.id = pc.component_id
  CROSS JOIN LATERAL jsonb_each(cc.input_schema->'fields') AS f(key,val)
  WHERE p.site_id='1244516d-014d-421c-88c6-090bb1e9552a' AND p.name='report-example'
    AND (f.val->>'required')::boolean IS TRUE
    AND NOT (pc.content_data ? f.key);
  IF missing IS NOT NULL THEN
    RAISE EXCEPTION 'ABORT: required field(s) still missing: % — the render would escalate to the LLM writer again', missing;
  END IF;
END $g$;

INSERT INTO site_work_items
  (site_id, source, item_type, severity, summary, spec, priority, handler_agent, status, created_by, item_key)
SELECT p.site_id, 'manual-p4_27', 'page_rerender', 'medium',
  'Rebuild /report/example/index.html after supplying the required heading',
  jsonb_build_object('domain','idea.uk','page_id',p.id::text,
                     'filename','report/example/index.html',
                     'page_name',p.name,'reason','section_data_resolved'),
  85, 'page-rerender', 'triaged', 'idea.uk vm 7 session, p4_27',
  'page_rerender_report_example_p4_27_' || p.site_id::text
FROM pages p WHERE p.site_id='1244516d-014d-421c-88c6-090bb1e9552a' AND p.name='report-example';
COMMIT;
SELECT id, status FROM site_work_items WHERE created_by LIKE '%p4_27%';
