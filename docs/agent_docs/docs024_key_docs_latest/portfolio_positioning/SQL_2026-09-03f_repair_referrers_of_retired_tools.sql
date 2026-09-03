-- 2026-09-03 — repair the surviving pages that still link to the three retired tools, so the
-- retraction (refused at 18:58Z, run 45a7eba9, "still linked from live content") can pass.
-- The retraction audit named 13 inbound links: footer chrome ×3; tool-cta ×3 (tool-insight-injector);
-- hero ×2 + call-to-action ×2 (both surviving guides); article-body ×3 (prose, both surviving guides).
--
-- Mechanisms, each the estate's own and none hand-editing content:
--   * CTA slots (tool-cta / hero / call-to-action): a page_rerender with reason=cta_links_stale —
--     the check_misdirected_cta repair path; page-rerender recomputes CTA targets against VALID pages
--     (archived excluded), stored content_data merged LAST. Mirrors that check's own item exactly.
--   * Footer chrome: needs_rerender with refresh_site_components=true → rerender-pages regenerates
--     header/footer/head from loadFetchablePageSet (status NOT IN deleted/archived) and re-assembles
--     the four active pages. Mirrors gamedesign's SEED_2026-09-03g.
--   * article-body prose: handled by the phantom_internal_links route (separate insert below the
--     verify block, filed only if the rebuild path is confirmed to re-resolve prose links).
-- Why by hand and not by waiting: completeness-discovery-agent selected ONE site in the last four
-- hours across 40 in rotation [MEASURED 2026-09-03 19:1xZ]; the checks would file these eventually.
BEGIN;
INSERT INTO site_work_items (site_id, page_id, source, pipeline, item_type, severity, summary, spec, priority, handler_agent, status, created_by, item_key)
SELECT '3d965325-519a-4515-b79f-50c886954a80', p.id, 'portfolio_positioning lane', 'content', 'page_rerender', 'high',
       'Recompute CTA targets on '||p.name||' after three tool pages were retired (retraction 45a7eba9 refused on these links)',
       jsonb_build_object('check','misdirected_cta','page_name',p.name,'page_id',p.id::text,
         'reason','cta_links_stale','routing_reason','cta_links_stale',
         'fix','Three tool pages are archived; CTA slots on this page still point at them. A cta_links_stale rerender recomputes CTA targets from real (active) pages.',
         'lane','portfolio_positioning','retired_targets',jsonb_build_array('/tools/serp-snippet-previewer/index.html','/tools/title-tag-scorer/index.html','/tools/keyword-intent-classifier/index.html')),
       35, 'page-rerender', 'triaged', 'portfolio_positioning lane 2026-09-03',
       'retired_tool_cta_repair:'||p.name||':2026-09-03'
  FROM pages p WHERE p.site_id='3d965325-519a-4515-b79f-50c886954a80' AND p.status='active'
   AND p.name IN ('tool-insight-injector','tool-insight-injector-guide','tool-website-brief-starter-guide')
ON CONFLICT DO NOTHING;

INSERT INTO site_work_items (site_id, source, pipeline, item_type, severity, summary, spec, priority, handler_agent, status, created_by, item_key)
VALUES ('3d965325-519a-4515-b79f-50c886954a80', 'portfolio_positioning lane', 'build', 'needs_rerender', 'medium',
  'Refresh site chrome + re-assemble the 4 active pages after 3 tool pages (and their guides) were retired; footer still links to the archived tools',
  '{"reason":"post_reconcile_assembly","refresh_site_components":true,"why":"six pages archived 2026-09-03 (three seotools-duplicate tools + companion guides); footer chrome rendered_html still carries hrefs to the three archived tool pages; render_site_components rebuilds from loadFetchablePageSet which excludes archived","lane":"portfolio_positioning"}'::jsonb,
  60, 'rerender-pages', 'triaged', 'portfolio_positioning lane 2026-09-03', 'chrome_refresh_retired_tools_3d965325_2026-09-03')
ON CONFLICT DO NOTHING;

DO $v$
DECLARE n int;
BEGIN
  SELECT count(*) INTO n FROM site_work_items WHERE site_id='3d965325-519a-4515-b79f-50c886954a80'
     AND item_key LIKE 'retired_tool_cta_repair:%:2026-09-03' AND status='triaged' AND spec->>'reason'='cta_links_stale' AND spec ? 'page_name';
  IF n <> 3 THEN RAISE EXCEPTION 'VERIFY: expected 3 cta repair items, found %', n; END IF;
  SELECT count(*) INTO n FROM site_work_items WHERE item_key='chrome_refresh_retired_tools_3d965325_2026-09-03' AND status='triaged' AND (spec->>'refresh_site_components')::boolean;
  IF n <> 1 THEN RAISE EXCEPTION 'VERIFY: chrome refresh item missing'; END IF;
END $v$;
COMMIT;

-- ---------------------------------------------------------------------------------------------
-- PART 2 (same evening, after reading resolve_internal_links_action.go): the CTA recompute covers
-- ONLY ctaFieldNames (hero, call-to-action, …) — NOT tool-cta and NOT article-body prose. Those two
-- surfaces are what check_phantom_internal_links routes to page-build-handler ("rebuild the page;
-- build-time link resolution re-runs") — the framework rewrites the content, nobody hand-edits it.
-- File exactly what that check would file, one item per (page, slot, href), DERIVED from the stored
-- html rather than typed, with the check's own item_key shape so a later sweep dedups against it.
BEGIN;
INSERT INTO site_work_items (site_id, page_id, source, pipeline, item_type, severity, summary, spec, priority, handler_agent, status, created_by, item_key)
SELECT DISTINCT p.site_id, p.id, 'discovery', 'content', 'phantom_internal_link', 'high',
       format('phantom_internal_link in page_component (%s:%s): href %s has no matching page', p.name, cc.function, t.href),
       jsonb_build_object('check','phantom_internal_links','issue_type','phantom_internal_link','surface','page_component',
         'page_name',p.name,'page_id',p.id::text,'slot_name',cc.function,'href',t.href,'occurrences',1,
         'fix','Internal href has no matching pages.url row (or is empty). Page surfaces: rebuild the page (build-time link resolution re-runs). Site surfaces: re-render site components from real-page data.',
         'lane','portfolio_positioning','filed_by_hand_because','completeness-discovery-agent selected 1 site in 4h across 40; the three targets were archived 2026-09-03 and retraction 45a7eba9 refused on these links'),
       35, 'page-build-handler', 'triaged', 'portfolio_positioning lane 2026-09-03',
       'phantom_internal_link:page_component:'||p.name||':'||cc.function||':'||t.href
  FROM pages p JOIN page_components pc ON pc.page_id=p.id JOIN content_components cc ON cc.id=pc.component_id,
       (VALUES ('/tools/serp-snippet-previewer/index.html'),('/tools/title-tag-scorer/index.html'),('/tools/keyword-intent-classifier/index.html')) AS t(href)
 WHERE p.site_id='3d965325-519a-4515-b79f-50c886954a80' AND p.status='active'
   AND cc.function IN ('tool-cta','article-body')
   AND pc.rendered_html LIKE '%href="'||t.href||'"%'
ON CONFLICT DO NOTHING;
DO $v$
DECLARE n int;
BEGIN
  SELECT count(*) INTO n FROM site_work_items WHERE site_id='3d965325-519a-4515-b79f-50c886954a80' AND item_type='phantom_internal_link' AND status='triaged' AND created_by LIKE 'portfolio_positioning%';
  IF n < 4 OR n > 6 THEN RAISE EXCEPTION 'VERIFY: expected 4-6 phantom items (tool-cta ×3 on one page + article-body ×1-3), found %', n; END IF;
END $v$;
COMMIT;
