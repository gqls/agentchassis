BEGIN;
-- OWNER RULING 2026-09-03 (second): "accept a lesser redirect checker" — the 8th planned tool, scoped to what a browser can honestly do.
CREATE TEMP TABLE fire(site_id uuid, domain text, name text, fn text, prio int, cx text, descr text, related jsonb, tp text);
INSERT INTO fire VALUES ((SELECT id FROM sites WHERE domain='seotools.co.uk'),'seotools.co.uk','Redirect Chain Checker','tool-redirect-chain-checker',7,'simple',
'A redirect-chain analyser that works from what the user can see, because a browser is not allowed to observe another site''s redirect hops. Primary mode: a textarea where the user pastes a chain as one hop per line in the form "<status> <url>" (for example "301 http://example.com/old" then "302 https://example.com/new" then "200 https://www.example.com/final"), or pastes the raw response headers from a curl -I run; the tool parses it and reports hop count, the final destination, whether the chain loops, http-to-https and www hops that could be collapsed, temporary (302/307) hops sitting where a permanent (301/308) one belongs, method-changing 302s, and a plain verdict with the single-hop redirect it recommends instead. Show the chain as a vertical timeline with a status badge per hop and flag chains longer than two hops as too long for crawlers. Secondary mode: a URL field that runs a best-effort browser fetch and, where the target permits cross-origin reads, reports only whether the final response was redirected and the final URL it landed on; when the target blocks it, say so and point the user to the paste mode and to a "curl -I -L" command it prints for them to copy. Include a short explainer of why redirect chains matter for crawl budget and link equity. Runs entirely in the browser: no server, no API key, no third-party proxy. State plainly in the UI what the tool cannot see.',
'["tools","technical-seo-crawlers-compared","glossary"]'::jsonb,'tools');
DO $$ DECLARE n int; BEGIN
  SELECT count(*) INTO n FROM fire f JOIN sites s ON s.id=f.site_id WHERE s.locked_at IS NULL AND s.status='deployed'; IF n<>1 THEN RAISE EXCEPTION 'site guard: %', n; END IF;
  SELECT count(*) INTO n FROM fire f JOIN pages p ON p.site_id=f.site_id AND p.name=f.fn AND p.page_type='tool' AND p.status='active'
    WHERE NOT EXISTS (SELECT 1 FROM page_components pc JOIN content_components cc ON cc.id=pc.component_id WHERE pc.page_id=p.id AND pc.build_status<>'removed' AND cc.component_level='tool');
  IF n<>1 THEN RAISE EXCEPTION 'page guard: %', n; END IF;
  SELECT count(*) INTO n FROM fire f JOIN site_work_items w ON w.site_id=f.site_id AND w.item_key='add_tool_novel_'||f.domain||'_'||f.fn AND w.status NOT IN ('complete','verified','rejected','wont_fix','failed','unresolved','cancelled');
  IF n<>0 THEN RAISE EXCEPTION 'dedup guard: %', n; END IF;
  SELECT count(*) INTO n FROM content_components WHERE component_level='tool' AND is_active AND function IN (SELECT fn FROM fire); IF n<>0 THEN RAISE EXCEPTION 'library guard: %', n; END IF;
END $$;
INSERT INTO site_work_items (site_id, source, item_type, severity, summary, spec, priority, handler_agent, status, created_by, item_key, pipeline, triaged_at, approval_mode)
SELECT f.site_id, 'portfolio_positioning', 'add_tool', 'low', f.name,
  jsonb_build_object('name', f.name, 'function', f.fn, 'priority', f.prio, 'complexity', f.cx, 'description', f.descr,
                     'target_page', f.tp, 'related_pages', f.related, 'library_source', NULL, 'experience_plan', NULL, 'tool_component_id', NULL,
                     'owner_ruling', '2026-09-03 accept a lesser redirect checker (bugs_open/450 instance; browser-only scope, no backend exists)'),
  120, 'tool-generator', 'triaged', 'portfolio_positioning', 'add_tool_novel_'||f.domain||'_'||f.fn, 'build', now(), 'auto'
FROM fire f;
COMMIT;
SELECT left(id::text,8) AS id, summary, item_key, status, created_at FROM site_work_items WHERE created_by='portfolio_positioning' AND item_type='add_tool' AND spec->>'function'='tool-redirect-chain-checker';
