-- p3_07_header_cta_to_paid_tool.sql — idea.uk: point the site-header "Get Started" button at the paid tool.
--
-- OWNER DECISION (AskUserQuestion 2026-07-23): after the body-CTA fix (p3_05/06), point the header
-- "Get Started" button at /report.html too — site-wide (it is chrome, on all 9 pages).
--
-- WHY A TEMPLATE EDIT. The header CTA is <a href="{{.cta_url}}" class="btn-primary"> and cta_url is
-- source=renderer: render_site_components_action.go:155-156 hard-sets it to the CONTACT nav page URL
-- fleet-wide, and the renderer does NOT read site_components.content_data, so there is no per-site data
-- override. idea.uk's site-header component (f420f3fa-…) is idea.uk-specific (1 site uses it), so
-- editing its template is safe and scoped. /report.html is a REAL page (the paid tool), not a phantom.
--
-- The {{if .cta_url}} gate is preserved (cta_url stays non-empty -> the button still renders); only the
-- href changes. Both occurrences (desktop header-actions + mobile-actions) are updated.
--
-- After this, a chrome refresh must run to re-render site_components and reassemble all pages:
--   agent_type=rerender-pages, input_data{site_id, domain, refresh_site_components:true}
-- (fired separately via Kafka — see RUNBOOK / RUNNING_NOTES §X.9).
--
-- ROLLBACK: bak_ideauk_header_20260723 holds the pre-change row.

\set ON_ERROR_STOP on

BEGIN;

CREATE TABLE IF NOT EXISTS bak_ideauk_header_20260723 AS
  SELECT now() AS snapshot_at, * FROM content_components WHERE id='f420f3fa-43a2-4a2f-b2e1-39770d45b494';

-- Guard: refuse unless exactly the 2 expected CTA anchors are present (so a template that has since
-- changed shape is not silently mis-edited).
DO $guard$
DECLARE n int;
BEGIN
  SELECT (length(html_template) - length(replace(html_template, 'href="{{.cta_url}}" class="btn-primary"', '')))
         / length('href="{{.cta_url}}" class="btn-primary"')
    INTO n
  FROM content_components WHERE id='f420f3fa-43a2-4a2f-b2e1-39770d45b494';
  IF n <> 2 THEN
    RAISE EXCEPTION 'ABORT: expected 2 header CTA anchors, found %; template shape changed, investigate.', n;
  END IF;
END
$guard$;

UPDATE content_components
SET html_template = replace(html_template,
      'href="{{.cta_url}}" class="btn-primary"',
      'href="/report.html" class="btn-primary"'),
    updated_at = now()
WHERE id='f420f3fa-43a2-4a2f-b2e1-39770d45b494';

COMMIT;

-- Read-back: the template should now carry /report.html and no bare {{.cta_url}} href.
SELECT
  (length(html_template) - length(replace(html_template, 'href="/report.html" class="btn-primary"',''))) / length('href="/report.html" class="btn-primary"') AS report_ctas,
  html_template LIKE '%href="{{.cta_url}}" class="btn-primary"%' AS still_has_old
FROM content_components WHERE id='f420f3fa-43a2-4a2f-b2e1-39770d45b494';
