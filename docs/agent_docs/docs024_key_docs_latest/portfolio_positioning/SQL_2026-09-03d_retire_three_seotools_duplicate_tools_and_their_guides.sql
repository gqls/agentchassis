-- 2026-09-03 — retire three copyonline.co.uk tool pages that duplicate seotools.co.uk, plus their
-- companion guides. Owner ruling 2026-09-03 (evening): "retire the duplicate tool pages, but check
-- it's correct first using broader search because we have retired good pages by mistake before."
--
-- THE BROADER SEARCH, all [MEASURED 2026-09-03 18:5xZ] before this file was written:
--   1. DUPLICATION at the component level, fleet-wide: tool-serp-snippet-previewer, tool-title-tag-scorer
--      and tool-keyword-intent-classifier each carry a live, deployed page on seotools.co.uk (and
--      keyword-intent-classifier on websitepromotion.co.uk too). Shared library components; the
--      content_components rows are NOT touched by this file.
--   2. THE LOANZY PRECEDENT (bugfix_311 NOTES: four tool pages archived that the site should have
--      held; owner: "it should have held pages"; all four unarchived). The check that precedent
--      demands: does the site's approved brief or its sighted strategy name these tools ANYWHERE?
--      brief:    serp=false  title_tag=false  keyword_intent=false   control 'headline scorer'=true
--      strategy: serp=false  title_tag=false  keyword_intent=false   control 'headline scorer'=true
--      The site should not hold them. The control proves the predicate can hit.
--   3. INBOUND LINKS, three sources (runbook_unpublish_primitive §"Audit a retraction's graph"), with
--      /tools/website-brief-starter/index.html as the positive control (body 9 / chrome footer / nav 1):
--      the three GUIDES have zero inbound from body, chrome or nav — already orphans (the checker's
--      own orphan_blog_posts finding at 18:05:36Z agrees). The three TOOLS are linked from the two
--      surviving pages' tool-cta/guide prose, from each other, from the footer chrome, and from one
--      nav row each. The retraction action refuses on editorial inbound and names referrers; that
--      refusal is the guard, and it is expected on the tools until their referrers are regenerated.
--   4. NOTHING IS PUBLISHED: sites.publish_target/published_at/last_deployed_at NULL, the domain still
--      serves the previous Drupal 7 install, every recorded pages.url 404s against a working
--      invented-URL control. Retiring these is an internal tidy.
--   5. Open work items on the six pages: 0.
--
-- WHY THE GUIDES TOO: each is "Understanding <tool> | Guide", created by the tool deployer as that
-- tool's companion, orphaned, and not in the brief. A guide to a tool that no longer exists is not a
-- page to keep. The two surviving tools' guides are NOT touched.
--
-- WHAT THIS FILE DOES: sets pages.status='archived' on exactly six named rows. Retraction (the
-- deployed artefact + nav rows) is the SECOND half and is fired separately via
-- docs/agent_docs/sql_for_agents/216_TRIGGER_page_retraction.sh — by decision 2026-08-04 archiving
-- never triggers it automatically.
BEGIN;
DO $g$
DECLARE n int; pub timestamptz;
BEGIN
  SELECT published_at INTO pub FROM sites WHERE id='3d965325-519a-4515-b79f-50c886954a80' AND domain='copyonline.co.uk';
  IF NOT FOUND THEN RAISE EXCEPTION 'REFUSED: site id/domain pair not found'; END IF;
  IF pub IS NOT NULL THEN RAISE EXCEPTION 'REFUSED: site has been published (%); re-audit at the served bytes first', pub; END IF;
  SELECT count(*) INTO n FROM pages WHERE site_id='3d965325-519a-4515-b79f-50c886954a80' AND status='active'
     AND id IN ('3ae2096f-98c6-4589-92cd-b2f343140fbb','260bf59f-49e4-489e-b460-6fac226e7ff9',
                '9fae1345-84a3-42dd-aa7b-b22ca314d335','7e158fe2-6c39-4eb3-8f45-5d2b2bf097d7',
                '09fdbca9-4d88-4011-907b-b5adf1206a82','3e50e18c-be80-4452-b232-70d12ecd8a44')
     AND name IN ('tool-serp-snippet-previewer','tool-serp-snippet-previewer-guide','tool-title-tag-scorer',
                  'tool-title-tag-scorer-guide','tool-keyword-intent-classifier','tool-keyword-intent-classifier-guide');
  IF n <> 6 THEN RAISE EXCEPTION 'REFUSED: expected 6 active rows matching id AND name, found %', n; END IF;
  SELECT count(*) INTO n FROM site_work_items WHERE site_id='3d965325-519a-4515-b79f-50c886954a80'
     AND page_id IN ('3ae2096f-98c6-4589-92cd-b2f343140fbb','260bf59f-49e4-489e-b460-6fac226e7ff9','9fae1345-84a3-42dd-aa7b-b22ca314d335','7e158fe2-6c39-4eb3-8f45-5d2b2bf097d7','09fdbca9-4d88-4011-907b-b5adf1206a82','3e50e18c-be80-4452-b232-70d12ecd8a44')
     AND status NOT IN ('complete','cancelled','rejected','failed');
  IF n <> 0 THEN RAISE EXCEPTION 'REFUSED: % open work item(s) on these pages', n; END IF;
END $g$;
UPDATE pages SET status='archived'
 WHERE site_id='3d965325-519a-4515-b79f-50c886954a80' AND status='active'
   AND id IN ('3ae2096f-98c6-4589-92cd-b2f343140fbb','260bf59f-49e4-489e-b460-6fac226e7ff9',
              '9fae1345-84a3-42dd-aa7b-b22ca314d335','7e158fe2-6c39-4eb3-8f45-5d2b2bf097d7',
              '09fdbca9-4d88-4011-907b-b5adf1206a82','3e50e18c-be80-4452-b232-70d12ecd8a44');
DO $v$
DECLARE a int; t int;
BEGIN
  SELECT count(*) FILTER (WHERE status='archived'), count(*) INTO a, t FROM pages WHERE site_id='3d965325-519a-4515-b79f-50c886954a80';
  IF a <> 6 OR t <> 10 THEN RAISE EXCEPTION 'VERIFY: expected 6 archived of 10, got % of %', a, t; END IF;
  IF (SELECT count(*) FROM pages WHERE site_id='3d965325-519a-4515-b79f-50c886954a80' AND status='active'
        AND name IN ('tool-website-brief-starter','tool-website-brief-starter-guide','tool-insight-injector','tool-insight-injector-guide')) <> 4
  THEN RAISE EXCEPTION 'VERIFY: a surviving page was touched'; END IF;
END $v$;
COMMIT;
