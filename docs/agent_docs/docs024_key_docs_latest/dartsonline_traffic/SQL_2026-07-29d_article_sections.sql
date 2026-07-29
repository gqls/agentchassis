-- SQL — dartsonline.com: give the nine buying guides a section layout (2026-07-29)
--
-- THE BLOCKER: all nine blog-post pages have `pages.sections = '[]'` AND zero
-- `site_plan_sections` rows. Nobody ever decided what blocks a guide page should
-- contain, so the writer had nothing to write and five earlier build attempts died
-- in the review queue (gemini_content_provider/README_where_we_are.md:270-320).
--
-- WHY THE PLAN TABLE, NOT pages.sections: `load_page_sections_from_spec_action.go`
-- resolves in order — (1) site_plan tables, (2) site_specs, (3) pages.sections,
-- (4) borrow from a same-role sibling. pages.sections is documented in that file as
-- the "materialised cache"; the plan tables are the authority, and the loader SYNCS
-- down to pages.sections itself once it resolves. Writing only the cache would leave
-- the authority empty and a future re-plan would wipe it. So write the authority.
--
-- Fallback 4 could not save these pages either: it borrows from a same-role sibling
-- WITH sections, and all nine siblings are the empty set. A whole role can be
-- sectionless and the sibling rule cannot see it.
--
-- THE LAYOUT is not invented here. `create_blog_posts_action.go:183` already encodes
-- the platform's canonical article default:
--     sections = []string{"hero", "article-body", "call-to-action"}
-- All three components verified live: content_components rows exist, section-level,
-- is_active = true.
--
-- NO blog-index page is created, and this is a deliberate change from the plan.
-- `guides-index` (/guides/index.html) is already deployed and already carries
-- `content-listing`, whose input_schema sources `query.blog_posts` — the very
-- resolver a blog index would use. It is the guides index already; it lists nothing
-- only because no guide has ever been built. A second listing page would duplicate
-- it and split internal links.
--
-- ORDERING NOTE: idx_site_plan_sections_key is UNIQUE on (plan_id, page_name,
-- ordering), so re-running this file is safe only with the NOT EXISTS guard below.

BEGIN;

CREATE TABLE IF NOT EXISTS bak_darts_plan_sections_20260729 AS
SELECT sps.* FROM site_plan_sections sps
JOIN site_plans sp ON sp.id = sps.plan_id
WHERE sp.site_id = '5fe8785b-223d-41a3-88ee-c07187622381';

INSERT INTO site_plan_sections (plan_id, page_name, ordering, component_name)
SELECT sp.id, p.name, v.ordering, v.component_name
FROM site_plans sp
JOIN pages p
  ON p.site_id = sp.site_id
 AND p.page_type = 'blog-post'
 AND p.status = 'active'
CROSS JOIN (VALUES (0, 'hero'), (1, 'article-body'), (2, 'call-to-action'))
  AS v(ordering, component_name)
WHERE sp.site_id = '5fe8785b-223d-41a3-88ee-c07187622381'
  AND sp.is_current = true
  AND NOT EXISTS (
        SELECT 1 FROM site_plan_sections x
        WHERE x.plan_id = sp.id AND x.page_name = p.name
      );

-- Sync the cache so the build can read it without waiting for the loader to resolve
-- and write back (the loader does this itself; doing it here makes the state legible
-- to the next person reading the pages table).
UPDATE pages
SET sections = '["hero","article-body","call-to-action"]'::jsonb, updated_at = now()
WHERE site_id = '5fe8785b-223d-41a3-88ee-c07187622381'
  AND page_type = 'blog-post'
  AND status = 'active'
  AND (sections IS NULL OR sections = '[]'::jsonb);

COMMIT;

-- Verify: nine pages, three sections each, cache agrees
SELECT p.name,
       (SELECT count(*) FROM site_plan_sections sps
          JOIN site_plans sp ON sp.id = sps.plan_id
         WHERE sp.site_id = p.site_id AND sp.is_current AND sps.page_name = p.name) AS plan_sections,
       jsonb_array_length(p.sections) AS cached_sections,
       p.build_status
FROM pages p
WHERE p.site_id = '5fe8785b-223d-41a3-88ee-c07187622381'
  AND p.page_type = 'blog-post' AND p.status = 'active'
ORDER BY p.name;
