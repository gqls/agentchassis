-- p4_21 — force a SECTION rerender of /report.html so p4_20's CTA URLs reach the page.
--
-- A plain `rerender-pages` item sets no spec.reason and takes the assemble-only
-- path, rebuilding from the STORED rendered_html — it can never apply a
-- content_data change. reason='section_data_resolved' routes through
-- rerender_page_sections, which re-renders every section from stored
-- content_data plus freshly resolved dynamic fields, with NO LLM involved.
-- See p3_04 for the full mechanism; this is the same manoeuvre, one page.
--
-- Queue checked before dispatch (CLAUDE.md: checking the page does not check the
-- queue): no open page_rerender exists for this page.

\set ON_ERROR_STOP on

BEGIN;

-- Refuse if any section lacks content_data: rerender_page_sections would escalate
-- the WHOLE page to the LLM content writer and rewrite live sales copy.
DO $guard$
DECLARE n int;
BEGIN
  SELECT count(*) INTO n
  FROM page_components pc
  WHERE pc.page_id = '41333d74-0c5a-4e12-b942-50ba4df793e6'
    AND (pc.content_data IS NULL OR pc.content_data = '{}'::jsonb);
  IF n > 0 THEN
    RAISE EXCEPTION
      'ABORT: % section(s) on /report.html have NULL/empty content_data; the rerender would escalate to the LLM content writer and rewrite live copy.', n;
  END IF;
END
$guard$;

-- Refuse if the CTA values are not actually in place — otherwise this dispatches
-- a render that changes nothing and reports success, which is the failure mode
-- this workstream keeps hitting.
DO $guard2$
DECLARE n int;
BEGIN
  SELECT count(*) INTO n
  FROM page_components
  WHERE page_id = '41333d74-0c5a-4e12-b942-50ba4df793e6'
    AND (content_data->>'cta_url' = '#request-a-report'
      OR content_data->>'primary_cta_url' = '#request-a-report');
  IF n <> 2 THEN
    RAISE EXCEPTION 'ABORT: expected 2 sections carrying #request-a-report, found %. Apply p4_20 first.', n;
  END IF;
END
$guard2$;

INSERT INTO site_work_items
  (site_id, source, item_type, severity, summary, spec, priority, handler_agent, status, created_by, item_key)
SELECT
  p.site_id,
  'manual-p4_21',
  'page_rerender',
  'medium',
  'Re-render /report.html sections to pick up p4_20 (hero + CTA now point at the request form, not /contact.html)',
  jsonb_build_object(
    'domain',    'idea.uk',
    'page_id',   p.id::text,
    'filename',  'report.html',
    'page_name', p.name,
    'reason',    'section_data_resolved'
  ),
  85,
  'page-rerender',
  'triaged',
  'idea.uk vm 7 session, p4_21',
  'page_rerender_report_p4_21_' || p.site_id::text
FROM pages p
WHERE p.id = '41333d74-0c5a-4e12-b942-50ba4df793e6';

COMMIT;

SELECT id, spec->>'page_name' AS page, spec->>'reason' AS reason, status, created_at
FROM site_work_items
WHERE created_by LIKE '%p4_21%';
