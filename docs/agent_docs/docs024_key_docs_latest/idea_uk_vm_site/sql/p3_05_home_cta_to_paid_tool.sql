-- p3_05_home_cta_to_paid_tool.sql — idea.uk: point the HOME PAGE CTAs at the paid £29 tool.
--
-- OWNER REPORT (2026-07-22/23): "the links on the home page still don't go to the paid for tool."
--
-- DIAGNOSIS (see RUNNING_NOTES §X.9). /report.html IS the paid tool (it carries the
-- action="/request" form + "Request a report" + £29). But the home page's four CTA sections were
-- NEVER link-resolved: none has a cta_url or resolved_data. So each button fell back to:
--   - hero {{.cta_url}}/{{.secondary_cta_url}}            -> RenderContext default /contact.html
--   - brief-explanation "Get Started"/"Learn More"        -> href="#" HARDCODED in the template
--   - tool-list "Browse All Tools" {{.cta_url}}           -> /contact.html (query.section_index_for:tool
--                                                            finds no tool section-index -> nil)
--   - call-to-action {{.primary/secondary_cta_url}}       -> empty/skip
-- The fleet-wide resolver (resolve_internal_links / applyCTARecompute) only ever points CTAs at
-- content HUBS (page_type='section-index'), explicitly excluding landing pages — so it would NEVER
-- choose /report.html (page_type='landing'). Funnelling the home page to the paid tool is idea.uk
-- business intent that must be set EXPLICITLY. Owner confirmed the mapping (AskUserQuestion 2026-07-23):
--   hero primary "See how it works"                       -> /report.html
--   hero secondary "Request a verified idea report"       -> /report.html
--   brief "Get Started" / "Learn More"                    -> /report.html (both)
--   tool-list "Browse All Tools"                          -> /tools.html
--   call-to-action "Run the free idea check"              -> /tools.html#audience-check
--   call-to-action "See what a verified idea report..."   -> /report.html
--
-- WHY THIS RENDERS AND HOLDS:
--   * hero/call-to-action cta url fields are source=renderer. applyCTARecompute runs ONLY for
--     reason='cta_links_stale' (rerender_page_sections_action.go:287); a section_data_resolved
--     rerender does NOT recompute them, so the value in content_data is used. contextToInterfaceMap
--     defaults cta_url to /contact.html then the ContentData merge OVERRIDES it (component_library.go),
--     and mergeIntoRenderContext captures ALL content_data keys for template access (v3_site_actions.go
--     :1465-1471), so secondary_cta_url/primary_cta_url surface too.
--   * tool-list.cta_url is query.section_index_for:tool; that query returns nil for idea.uk (no tool
--     section-index), and the resolver "assigns any NON-nil value" (section_index_for.go), so the
--     content_data value survives re-resolution.
--   * brief-explanation hardcodes href="#" with NO url field, so it needs a TEMPLATE edit. The edit is
--     GATED ({{if .cta_primary_url}}...{{else}}#{{end}}) and backward-compatible: the 3 other instances
--     (verified: none sets cta_primary_url/cta_secondary_url) still render "#" exactly as before.
--
-- Locking of these sections is deferred to p3_06 (AFTER the live page is verified), so a lock cannot
-- interfere with this rerender (the rerender path does not gate on lock, but verify-before-lock is safer).

\set ON_ERROR_STOP on

BEGIN;

-- ---------------------------------------------------------------------------
-- 1. Shared brief-explanation template: gate the two hardcoded href="#" CTAs.
--    Backward-compatible no-op for any instance that does not set the url fields.
-- ---------------------------------------------------------------------------
UPDATE content_components
SET html_template = replace(replace(html_template,
      '<a href="#" class="brief-explanation__cta-primary">{{.cta_primary_label}}</a>',
      '<a href="{{if .cta_primary_url}}{{.cta_primary_url}}{{else}}#{{end}}" class="brief-explanation__cta-primary">{{.cta_primary_label}}</a>'),
      '<a href="#" class="brief-explanation__cta-secondary">{{.cta_secondary_label}}</a>',
      '<a href="{{if .cta_secondary_url}}{{.cta_secondary_url}}{{else}}#{{end}}" class="brief-explanation__cta-secondary">{{.cta_secondary_label}}</a>'),
    updated_at = now()
WHERE name = 'brief-explanation';

-- ---------------------------------------------------------------------------
-- 2. Set the CTA target URLs on the INDEX page's four CTA sections (merge; copy preserved).
-- ---------------------------------------------------------------------------
UPDATE page_components pc
SET content_data = pc.content_data || '{"cta_url":"/report.html","secondary_cta_url":"/report.html"}'::jsonb
FROM content_components cc
WHERE cc.id = pc.component_id AND cc.name = 'hero'
  AND pc.page_id = 'b147b925-a389-491a-8cce-ef6ba926d86a';

UPDATE page_components pc
SET content_data = pc.content_data || '{"primary_cta_url":"/tools.html#audience-check","secondary_cta_url":"/report.html"}'::jsonb
FROM content_components cc
WHERE cc.id = pc.component_id AND cc.name = 'call-to-action'
  AND pc.page_id = 'b147b925-a389-491a-8cce-ef6ba926d86a';

UPDATE page_components pc
SET content_data = pc.content_data || '{"cta_url":"/tools.html"}'::jsonb
FROM content_components cc
WHERE cc.id = pc.component_id AND cc.name = 'tool-list'
  AND pc.page_id = 'b147b925-a389-491a-8cce-ef6ba926d86a';

UPDATE page_components pc
SET content_data = pc.content_data || '{"cta_primary_url":"/report.html","cta_secondary_url":"/report.html"}'::jsonb
FROM content_components cc
WHERE cc.id = pc.component_id AND cc.name = 'brief-explanation'
  AND pc.page_id = 'b147b925-a389-491a-8cce-ef6ba926d86a';

-- ---------------------------------------------------------------------------
-- 3. Guard: refuse if any index section has NULL/empty content_data (would escalate the
--    whole page to the LLM content writer and rewrite live sales copy).
-- ---------------------------------------------------------------------------
DO $guard$
DECLARE n int;
BEGIN
  SELECT count(*) INTO n
  FROM page_components pc
  WHERE pc.page_id = 'b147b925-a389-491a-8cce-ef6ba926d86a'
    AND (pc.content_data IS NULL OR pc.content_data = '{}'::jsonb);
  IF n > 0 THEN
    RAISE EXCEPTION 'ABORT: % index section(s) have NULL/empty content_data; section_data_resolved would escalate to the LLM content writer.', n;
  END IF;
END
$guard$;

-- ---------------------------------------------------------------------------
-- 4. Rerender the INDEX page's sections from stored content_data (reason=section_data_resolved,
--    so no CTA recompute, no LLM). This regenerates rendered_html with the new CTA hrefs.
-- ---------------------------------------------------------------------------
INSERT INTO site_work_items
  (site_id, source, item_type, severity, summary, spec, priority, handler_agent, status, created_by, item_key)
SELECT
  p.site_id,
  'manual-p3_05',
  'page_rerender',
  'medium',
  'Re-render index sections to point home-page CTAs at the paid tool (p3_05)',
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
  'idea.uk vm site 5 session, p3_05',
  'page_rerender_p3_05_' || p.name || '_' || p.site_id::text
FROM pages p
WHERE p.id = 'b147b925-a389-491a-8cce-ef6ba926d86a';

COMMIT;

-- Read-back
SELECT cc.name AS section, pc.content_data->>'cta_url' AS cta_url,
       pc.content_data->>'secondary_cta_url' AS sec_url,
       pc.content_data->>'primary_cta_url' AS pri_url,
       pc.content_data->>'cta_primary_url' AS brief_pri, pc.content_data->>'cta_secondary_url' AS brief_sec
FROM page_components pc JOIN content_components cc ON cc.id=pc.component_id
WHERE pc.page_id='b147b925-a389-491a-8cce-ef6ba926d86a'
  AND cc.name IN ('hero','call-to-action','tool-list','brief-explanation')
ORDER BY pc.position;

SELECT spec->>'page_name' AS page, spec->>'reason' AS reason, status, created_at
FROM site_work_items WHERE created_by LIKE '%p3_05%' ORDER BY created_at DESC LIMIT 3;
