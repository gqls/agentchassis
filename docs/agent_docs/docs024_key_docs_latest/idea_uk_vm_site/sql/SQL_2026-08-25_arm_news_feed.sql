-- SQL — idea.uk: arm the news feed (HANDOFF_2026-08-25 §4 item 1; the oldest live gap)
--
-- WHAT THIS IS. idea.uk has served a 404 at /data/latest-news.json since the site went
-- live, and has 0 rows in content_sources. Both halves of the framework's news machinery
-- select a site on ONE flag in its classification spec:
--
--   site_specs.aspect='classification' AND is_current
--   AND (data->'content_features'->'news_feed'->>'recommended')::boolean = true
--
-- `content-feed-trigger.find_news_sites` (6-hourly, LIVE, last completed 2026-08-25 14:49Z)
-- and `MissingNewsSourcesCheck` (check_news_feed.go:83-94, in the hourly completeness
-- rotation) both read it. idea.uk's classification spec (written 2026-06-21 by
-- domain-research-classifier, the only row it has) carries NO `content_features` key at
-- all, so neither path has ever been able to select it.
--
-- WHY HAND-AUTHORED rather than letting the platform write the key. The writer is
-- `evaluate_news_feed` (feed_news_recommendation_action.go). Its matcher,
-- matchVerticalNews, builds signals from classification.industry / .site_type /
-- .category and domain substrings; it never reads `industry_tags`. [MEASURED 2026-08-25]
-- 0 of 31 current classification specs carry an `industry` string (27 carry
-- `industry_tags`), so that first signal is empty fleet-wide. idea.uk's remaining signals
-- — "interactive-platform", "interactive", "idea.uk" — contain none of the 27 vertical
-- keys, so the matcher returns nil and the action writes NOTHING on no-match (it only
-- writes when a vertical matches). And its only live carrier, improvement-loop, sits
-- behind the `improvement-sweep` scheduled task, which is DISABLED (enabled=false, last
-- completed 2026-08-22). Same wall, same remedy, same reasoning as
-- dartsonline_traffic/SQL_2026-07-29e_arm_news_feed.sql; the fleet correlation is exact:
-- 9 of 9 sites with recommended=true have sources, 22 of 22 without the flag have none.
--
-- source_types is `news_search` ONLY, deliberately. `seed_content_sources_action.go`
-- creates one `news_search` source per vertical_keyword (real articles with real URLs,
-- fetched by the search adapter — fundamentallyai.com's four run at error_count 0) and,
-- if listed, one `api_news` source whose items are LLM-authored (xAI/Grok, web_search +
-- x_search). idea.uk's entire positioning is the honest assessment (honesty arc CLOSED
-- 2026-08-17, migration 454), and vetcomparison.uk's precedent is that a site remediated
-- for fabricated content must not publish LLM-authored news. Leaving `api_news` out costs
-- nothing to reverse — add it to source_types and the seeder's ON CONFLICT DO NOTHING
-- means a re-run creates only the missing source. The owner opts in; this file does not.
--
-- vertical_keywords are the search QUERIES (one source each, num_results 10). They are
-- journalism-shaped on purpose — webdesign.co.uk's first ingestion showed that
-- market-category queries return vendor landing pages and listicles, while
-- institution/authority queries return reporting. Retune after the first ingestion, as
-- that lane did; the names are the (site_id, name) dedup key, so a retune is a DELETE of
-- the old source plus a re-run, not an UPDATE of this spec alone.
--
-- separate_page=true means a page_type='news-index' page is expected:
-- render_news_section_action.go:216 gates data/news-archive.json AND the homepage
-- snippet's "More insights →" link on one existing, and MissingNewsPageCheck files a
-- missing_news_page gap (→ content-gap-planner creates a SECOND /news page) when none
-- does. idea.uk already has the listing: pages 4f381fed, NAMED news-index, sections
-- [hero, news-listing], deployed, serving /news/index.html — but TYPED section-index
-- (the "-index" name rule flattened it; bugs_closed/015's class). page_type is a routing
-- key, not a label. This file re-types it on BOTH layers, because
-- site_db_actions.go:1240 re-upserts `page_type = EXCLUDED.page_type` from the plan on
-- every save: the current plan row 0417d6ed (role section-index) AND the pages row.
-- ValidateRoles rule 1b (page_role_validator.go:157-166) trusts an explicit news-index,
-- so it sticks.
--
-- What this file does NOT do: create sources, fetch, triage, render or deploy. Those are
-- the framework's — content-feed-orchestrator (seed_sources → dispatch_sources →
-- triage → render_news_json → commit_news into the vm-sites repo → sitesync), reached
-- either by the 6-hourly trigger or by one direct dispatch (RUNBOOK). The homepage
-- already carries the latest-news slot and /tools/assets/latest-news.js fetches
-- /data/latest-news.json; /tools/assets/news-listing.js fetches /data/news-archive.json.
--
-- Prior rows in bak_ideauk_newsarm_20260825_* (three tables). Rollback: restore the three
-- rows from them (site_specs: delete the authored row, set the backup row is_current=true).

\set ON_ERROR_STOP on
BEGIN;

CREATE TABLE IF NOT EXISTS bak_ideauk_newsarm_20260825_site_specs AS
  SELECT * FROM site_specs
  WHERE site_id = '1244516d-014d-421c-88c6-090bb1e9552a' AND aspect = 'classification';
CREATE TABLE IF NOT EXISTS bak_ideauk_newsarm_20260825_pages AS
  SELECT * FROM pages WHERE id = '4f381fed-8f9d-4bb9-9b76-4ee243bebe33';
CREATE TABLE IF NOT EXISTS bak_ideauk_newsarm_20260825_site_plan_pages AS
  SELECT * FROM site_plan_pages WHERE id = '0417d6ed-53a2-427e-9a5d-dba4488709a0';

-- Preconditions. DO/RAISE, not SELECT: a non-empty SELECT cannot stop the COMMIT.
DO $$
DECLARE
  n_cur   int;
  has_cf  bool;
  pt      text;
  prole   text;
  n_src   int;
BEGIN
  SELECT count(*), COALESCE(bool_or(data ? 'content_features'), false)
    INTO n_cur, has_cf
  FROM site_specs
  WHERE site_id = '1244516d-014d-421c-88c6-090bb1e9552a'
    AND aspect = 'classification' AND is_current = true;
  IF n_cur <> 1 THEN
    RAISE EXCEPTION 'expected exactly 1 current classification spec for idea.uk, found %', n_cur;
  END IF;
  IF has_cf THEN
    RAISE EXCEPTION 'idea.uk classification already carries content_features — re-read before re-applying';
  END IF;

  SELECT page_type INTO pt FROM pages
  WHERE id = '4f381fed-8f9d-4bb9-9b76-4ee243bebe33'
    AND site_id = '1244516d-014d-421c-88c6-090bb1e9552a';
  IF pt IS DISTINCT FROM 'section-index' THEN
    RAISE EXCEPTION 'news page 4f381fed page_type is % (expected section-index)', COALESCE(pt, '<missing>');
  END IF;

  SELECT spp.role INTO prole
  FROM site_plan_pages spp JOIN site_plans sp ON sp.id = spp.plan_id
  WHERE spp.id = '0417d6ed-53a2-427e-9a5d-dba4488709a0'
    AND sp.site_id = '1244516d-014d-421c-88c6-090bb1e9552a' AND sp.is_current = true;
  IF prole IS DISTINCT FROM 'section-index' THEN
    RAISE EXCEPTION 'plan row 0417d6ed role is % (expected section-index on the CURRENT plan)', COALESCE(prole, '<missing>');
  END IF;

  SELECT count(*) INTO n_src FROM content_sources
  WHERE site_id = '1244516d-014d-421c-88c6-090bb1e9552a' AND is_active = true;
  IF n_src <> 0 THEN
    RAISE EXCEPTION 'idea.uk already has % active content_sources — the seeder is all-or-nothing, stop and read them', n_src;
  END IF;
END $$;

-- 1. Supersede the classifier's row and insert the enriched copy, the way
--    evaluate_news_feed does it (supersede + insert, never UPDATE in place).
WITH old AS (
  UPDATE site_specs
  SET is_current = false, superseded_at = now()
  WHERE site_id = '1244516d-014d-421c-88c6-090bb1e9552a'
    AND aspect = 'classification' AND is_current = true
  RETURNING site_id, data
)
INSERT INTO site_specs (site_id, aspect, data, source, source_agent, is_current, created_by, notes)
SELECT
  old.site_id,
  'classification',
  old.data || jsonb_build_object(
    'content_features', jsonb_build_object(
      'news_feed', jsonb_build_object(
        'recommended', true,
        'separate_page', true,
        'source_types', jsonb_build_array('news_search'),
        'vertical_keywords', jsonb_build_array(
          'UK startup funding rounds',
          'Innovate UK grants and competitions',
          'UK Intellectual Property Office patents',
          'Start Up Loans British Business Bank',
          'UK small business and startup news'
        ),
        'reason',
          'idea.uk helps people test and develop an idea before they commit to it: the '
          || 'guides cover funding, patents, testing, building and feedback, and the tools '
          || 'sit on the same ground. The news its readers return for is the funding, '
          || 'grant, IP and launch landscape in the UK. The classifier''s own reasoning '
          || 'records the news section as a significant planned expansion. Hand-authored '
          || '2026-08-25: matchVerticalNews reads industry/site_type/category and domain '
          || 'substrings only, never industry_tags, and none of interactive-platform / '
          || 'interactive / idea.uk contains a vertical key, so the automatic path cannot '
          || 'reach a decision here. news_search ONLY — no LLM-authored (api_news) items '
          || 'on a site whose product is the honest assessment; the owner may opt in by '
          || 'adding api_news to source_types and re-running the orchestrator.'
      )
    )
  ),
  'authored',
  'idea-uk-vm-site-workstream',
  true,
  'idea-uk-vm-site-workstream',
  'Arms the news feed (HANDOFF_2026-08-25 §4 item 1). Adds content_features.news_feed '
  || 'only; every other key is the 2026-06-21 classifier row verbatim. Prior row in '
  || 'bak_ideauk_newsarm_20260825_site_specs.'
FROM old;

-- 2. Re-type the existing news listing on BOTH layers (plan first — pages is a cache of it).
UPDATE site_plan_pages
SET role = 'news-index'
WHERE id = '0417d6ed-53a2-427e-9a5d-dba4488709a0' AND role = 'section-index';

UPDATE pages
SET page_type = 'news-index', updated_at = now()
WHERE id = '4f381fed-8f9d-4bb9-9b76-4ee243bebe33'
  AND site_id = '1244516d-014d-421c-88c6-090bb1e9552a'
  AND page_type = 'section-index';

-- 3. Verify, including the trigger's own predicate — a DO block so a miss aborts the txn.
DO $$
DECLARE
  rec     bool;
  pt      text;
  prole   text;
  picked  int;
BEGIN
  SELECT (data->'content_features'->'news_feed'->>'recommended')::boolean INTO rec
  FROM site_specs
  WHERE site_id = '1244516d-014d-421c-88c6-090bb1e9552a'
    AND aspect = 'classification' AND is_current = true;
  IF rec IS DISTINCT FROM true THEN
    RAISE EXCEPTION 'current classification does not read recommended=true after insert';
  END IF;

  SELECT page_type INTO pt FROM pages WHERE id = '4f381fed-8f9d-4bb9-9b76-4ee243bebe33';
  SELECT role INTO prole FROM site_plan_pages WHERE id = '0417d6ed-53a2-427e-9a5d-dba4488709a0';
  IF pt <> 'news-index' OR prole <> 'news-index' THEN
    RAISE EXCEPTION 'news page not re-typed on both layers: pages=% plan=%', pt, prole;
  END IF;

  -- content-feed-trigger.find_news_sites, reduced to this site.
  SELECT count(*) INTO picked
  FROM sites s
  JOIN site_specs ss ON ss.site_id = s.id AND ss.aspect = 'classification' AND ss.is_current = true
    AND (ss.data->'content_features'->'news_feed'->>'recommended')::boolean = true
  WHERE s.id = '1244516d-014d-421c-88c6-090bb1e9552a'
    AND EXISTS (SELECT 1 FROM pages p WHERE p.site_id = s.id AND p.build_status = 'deployed')
    AND NOT EXISTS (SELECT 1 FROM content_sources cs WHERE cs.site_id = s.id AND cs.is_active = true);
  IF picked <> 1 THEN
    RAISE EXCEPTION 'the trigger predicate would still not select idea.uk (picked=%)', picked;
  END IF;
END $$;

COMMIT;

-- Read-back (informational; runs after COMMIT).
SELECT source, source_agent, is_current, created_at,
       data->'content_features'->'news_feed'->>'recommended'  AS recommended,
       data->'content_features'->'news_feed'->'source_types'   AS source_types,
       jsonb_array_length(data->'content_features'->'news_feed'->'vertical_keywords') AS n_keywords
FROM site_specs
WHERE site_id = '1244516d-014d-421c-88c6-090bb1e9552a' AND aspect = 'classification'
ORDER BY created_at;
SELECT 'pages' AS layer, page_type AS type FROM pages WHERE id = '4f381fed-8f9d-4bb9-9b76-4ee243bebe33'
UNION ALL
SELECT 'plan', role FROM site_plan_pages WHERE id = '0417d6ed-53a2-427e-9a5d-dba4488709a0';
