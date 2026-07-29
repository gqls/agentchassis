-- SQL_2026-07-29n_news_page.sql
--
-- Create the news page work item directly, rather than waiting for a discovery
-- pass to raise it.
--
-- WHY THIS IS SAFE TO HAND-RAISE. MissingNewsPageCheck's preconditions are all
-- met and were checked one by one against the live row, not assumed:
--   classification.content_features.news_feed.recommended   = true   ✓
--   classification.content_features.news_feed.separate_page = true   ✓
--   pages with page_type='news-index'                       = 0      ✓
--   active content_sources                                  = 9      ✓
-- so the check would raise exactly this item on its next run. Raising it here
-- skips the wait, and — the actual reason — it would otherwise land at
-- status='detected', which is where such items go to die: the only other
-- missing_news_page items in the fleet are ai-agent-orchestration.com's, one
-- 'detected' since 2026-07-24 and one 'unresolved' since 2026-05-01. That is
-- bugs_open/083's stranded-findings problem, whose owner has ruled "routing is
-- NOT the bottleneck" and "decision pending — do not act". So this raises the
-- item for THIS SITE at 'triaged' — data, not a change to the mechanism.
--
-- approach = 'new_page', asserted rather than left to the handler. The
-- retype_existing branch exists for bugs_open/015, where relojistas.com's news
-- listing had been emitted as page_type='section-index' and creating a second
-- page would have shipped an English /news.html alongside the Spanish one. That
-- cannot apply here: after the nav rebuild this site has no nav-linked
-- sectionless page at all (guides-index, new-arrivals, sale and the four
-- utility pages all have sections; shop-index and brands-index are now
-- in_header=false).
--
-- THE ONE THING THAT MUST BE RIGHT. page_type is a routing key, not a label —
-- several independent gates key on the literal 'news-index'. bugs_open/081 is
-- the case where a mistyped page was DEPLOYED, and a deployed mistyped page has
-- no repair path; it has been looping for about three months and needs an owner
-- call. So verify the row BEFORE it builds, not after:
--
--   SELECT name, url, page_type, status FROM pages
--   WHERE site_id='5fe8785b-223d-41a3-88ee-c07187622381' AND page_type='news-index';
--   -- expect exactly: news | /news/index.html | news-index | active
--
-- If page_type comes back as anything else, fix it before the build item runs.
--
-- Site: dartsonline.com  5fe8785b-223d-41a3-88ee-c07187622381

INSERT INTO site_work_items (
  site_id, source, pipeline, item_type, severity, summary, spec,
  priority, handler_agent, status, created_by, item_key, approval_mode
) VALUES (
  '5fe8785b-223d-41a3-88ee-c07187622381',
  'dartsonline-traffic-workstream',
  'content',
  'missing_news_page',
  'medium',
  'dartsonline.com needs a dedicated /news.html page (separate_page=true in spec, 9 active sources, 14 relevant items)',
  jsonb_build_object(
    'check',       'missing_news_page',
    'page_name',   'news',
    'page_type',   'news-index',
    'category',    'content_completeness',
    'approach',    'new_page',
    'description', 'dartsonline.com is recommended for a dedicated news page (separate_page=true) and has no news-index page. It has 9 active content sources and 14 feed items already marked relevant, so there is something to show on day one. Create a /news.html page with a news-listing component to display the news archive.',
    'suggestion',  'Create a new page named ''news'' with page_type ''news-index'' (that exact literal — several gates key on it and the wrong value silences all of them at once), a hero section, the news-listing component, and a call-to-action. Add it to header and footer navigation. IMPORTANT for this site: it is a darts PUBLICATION that sells nothing, holds no stock and has no checkout. The hero must not promise products, offers or arrivals. The news here is PDC circuit results, rankings and equipment releases reported from other outlets with attribution — our value is the analysis alongside it, not the scoop.',
    'news_feed_config', (
      SELECT data->'content_features'->'news_feed' FROM site_specs
      WHERE site_id='5fe8785b-223d-41a3-88ee-c07187622381'
        AND aspect='classification' AND is_current
    )
  ),
  60,
  'content-gap-planner',
  'triaged',
  'dartsonline-traffic-workstream',
  'missing_news_page:5fe8785b-223d-41a3-88ee-c07187622381',
  'auto'
)
ON CONFLICT DO NOTHING;

-- ── verification ───────────────────────────────────────────────────────────
--   SELECT status, attempt_count, updated_at FROM site_work_items
--   WHERE item_key='missing_news_page:5fe8785b-223d-41a3-88ee-c07187622381';
--
-- Then the page_type check above, BEFORE the build item it spawns runs.
-- Then, once built, on the served page:
--   curl -sI https://dartsonline.com/news/index.html   # or /news.html — the
--                                                      # planner chooses; take
--                                                      # the URL from pages.url
