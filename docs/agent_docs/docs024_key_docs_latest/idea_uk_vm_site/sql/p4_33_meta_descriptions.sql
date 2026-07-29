-- p4_33 — meta descriptions for the 6 pages serving an EMPTY one (demand lane,
-- 2026-07-29). The homepage — the only page real Googlebot crawls — served
-- <meta name="description" content=""> until now.
--
-- privacy is deliberately SKIPPED: nginx 301s /privacy.html to the tool's
-- canonical /privacy, so the static file (and its meta) is never served.
--
-- Copy was grounded against each live page's actual content before writing
-- (about's "thinking partner" line is the page's own). All under 155 chars,
-- comfortably below datahelpers.PublicMetaDescription's 320 brief-detection
-- bound.
--
-- Dispatch is PLAIN page_rerender (no spec.reason = assemble mode): the head
-- is rebuilt from pages.meta_description on every render
-- (rerender_single_page_action.go:357), and assemble mode never inspects
-- content_data — so pages with derived/NULL sections (both hubs) cannot
-- escalate to the LLM writer.
\set ON_ERROR_STOP on
\set site '1244516d-014d-421c-88c6-090bb1e9552a'

BEGIN;

UPDATE pages SET meta_description = v.d, updated_at = now()
FROM (VALUES
  ('index',        'Most people with a serious idea have nowhere good to work it out. Plain-English guides, free tools, and a researched £29 report that pushes back honestly.'),
  ('about',        'Why idea.uk exists: a thinking partner for the part of the process that usually happens alone — not a cheerleader, a rigorous place to think.'),
  ('contact',      'How to reach idea.uk with a question about an idea, a tool or a report — email, phone, or the form on this page.'),
  ('tools',        'Four ways to test an idea before you build it: a free audience check, a patent steer, a funding-route finder, and the £29 Verified Idea Report.'),
  ('guides-index', 'Nine plain-English guides in journey order: creating ideas, building it, testing it, user acceptance, feedback loops, patents, copyright and funding.'),
  ('news-index',   'Notes and updates from idea.uk — new guides, new tools, and what has changed.')
) AS v(name, d)
WHERE pages.site_id = :'site' AND pages.name = v.name;

DO $g$
DECLARE n int;
BEGIN
  SELECT count(*) INTO n FROM pages
  WHERE site_id='1244516d-014d-421c-88c6-090bb1e9552a'
    AND name IN ('index','about','contact','tools','guides-index','news-index')
    AND COALESCE(meta_description,'') <> ''
    AND length(meta_description) BETWEEN 40 AND 320;
  IF n <> 6 THEN
    RAISE EXCEPTION 'ABORT: expected 6 pages with a sane meta_description, found %', n;
  END IF;
END $g$;

INSERT INTO site_work_items
  (site_id, source, item_type, severity, summary, spec, priority, handler_agent, status, created_by, item_key)
SELECT p.site_id, 'manual-p4_33', 'page_rerender', 'low',
  'Rebuild ' || p.url || ' head with the new meta description (assemble mode)',
  jsonb_build_object('domain','idea.uk','page_id',p.id::text,
                     'filename', ltrim(p.url,'/'),
                     'page_name', p.name),
  60, 'page-rerender', 'triaged', 'idea.uk vm 9 session, p4_33',
  'page_rerender_' || p.name || '_p4_33_' || p.site_id::text
FROM pages p
WHERE p.site_id = :'site'
  AND p.name IN ('index','about','contact','tools','guides-index','news-index');

COMMIT;

SELECT name, length(meta_description) AS len FROM pages
WHERE site_id = :'site'
  AND name IN ('index','about','contact','tools','guides-index','news-index')
ORDER BY name;
SELECT count(*) AS queued FROM site_work_items WHERE created_by LIKE '%p4_33%';
