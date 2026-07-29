-- SQL_2026-07-29s_listings_that_list_nothing.sql
--
-- The Guides hub links to no guides, and the homepage's listing has been empty
-- for a fortnight. Both were found by reading the served pages, not the queue.
--
-- WHAT THE SERVED /guides/index.html CONTAINS: a hero, a heading "Guides &
-- Advice for Every Player", a subtitle — and zero links to any guide. A visitor
-- clicking Guides lands on a page that describes guides and offers none of them.
-- It is also the internal-linking failure that matters most for search: eight
-- guide pages with no hub linking to them.
--
-- WHY, and the answer is NOT the resolver. Run by hand, the resolver's own query
-- returns all eight:
--
--   SELECT p.name, p.url FROM pages p
--    WHERE p.site_id = '5fe8785b-…' AND p.page_type = 'blog-post'
--      AND p.status IN ('active','deployed') AND p.deployed_at IS NOT NULL
--      AND jsonb_typeof(p.sections) = 'array' AND jsonb_array_length(p.sections) > 0
--   -- 8 rows: barrel-weight, beginners, board-setup, brand-comparison,
--   --         flight-shapes, shaft-length, steel-tip-vs-soft-tip, tungsten-guide
--
-- The component's stored data says `"articles": []`, and its row was last
-- written **2026-07-20 10:59** — when the site had no guides at all. Everything
-- since has been REASSEMBLY, not regeneration: the page was redeployed at
-- 16:43 today with a corrected header and its 20-July body intact.
--
-- That is reassembly working exactly as designed, and it is the trap. A
-- page_rerender re-injects chrome and leaves content alone — which is precisely
-- why it was the right tool for the two meta-description fixes an hour ago, and
-- precisely why it is the WRONG tool here. A listing's contents are not chrome;
-- they are resolved at WRITE time, so a listing only learns about new pages when
-- something sends it back to the writer.
--
--   **A stale listing is invisible in every status.** build_status says
--   'deployed', deployed_at is today, the item said complete, and the page
--   serves 200. Only page_components.updated_at (2026-07-20) disagrees, and
--   nothing surfaces it.
--
-- THE HOMEPAGE, same class, older. Its `category-listing` component renders 315
-- bytes — an empty grid — and there has been an `empty_section` work item open
-- against it since 2026-07-14. The reason is structural rather than stale:
-- category-listing resolves `query.category_posts`, and this site has no
-- categories, so it can never fill. `content-listing` resolves `query.blog_posts`,
-- which returns the eight guides. So the homepage swaps the section it cannot
-- fill for the one it can, and gains eight internal links to the pages we most
-- want found.
--
-- Site: dartsonline.com  5fe8785b-223d-41a3-88ee-c07187622381

BEGIN;

CREATE TABLE IF NOT EXISTS bak_20260729s_pages AS
SELECT id, name, sections, suppressed_sections FROM pages
WHERE site_id = '5fe8785b-223d-41a3-88ee-c07187622381';

-- Homepage: category-listing (unfillable — no categories) -> content-listing
-- (fillable — eight deployed guides). product-grid and testimonials stay in
-- sections and stay suppressed; they are the affiliate-era placeholders and
-- removing them would lose the record that they were planned.
UPDATE pages SET
  sections = '["hero", "product-grid", "info-card-grid", "content-listing", "call-to-action", "testimonials"]'::jsonb,
  updated_at = now()
WHERE site_id = '5fe8785b-223d-41a3-88ee-c07187622381' AND name = 'index';

COMMIT;

-- ── regeneration, not reassembly ───────────────────────────────────────────
INSERT INTO site_work_items (
  site_id, source, pipeline, item_type, severity, summary, spec,
  priority, handler_agent, status, created_by, item_key, approval_mode
)
SELECT
  '5fe8785b-223d-41a3-88ee-c07187622381', 'dartsonline-traffic-workstream', 'build',
  'needs_page', 'high', v.summary,
  jsonb_build_object(
    'page_name', v.page_name, 'page_role', v.page_role,
    'plan_id', '0fb05b75-04f4-4f4c-8890-c34d6a71012c',
    'reason', 'listing_stale', 'note', v.note
  ),
  35, 'page-build-handler', 'triaged', 'dartsonline-traffic-workstream',
  'listings_2026_07_29:' || v.page_name || ':5fe8785b-223d-41a3-88ee-c07187622381', 'auto'
FROM (VALUES
  ('guides-index', 'section-index',
   'The Guides hub links to no guides — its listing was written 2026-07-20, before any existed',
   'MUST go to the writer, not to reassembly: page_components for this page were last written 2026-07-20 10:59 and every rebuild since has been a reassembly that preserved them. The content-listing section resolves query.blog_posts, which now returns eight deployed guides. Keep the existing hero and the section headings — the only thing wrong with this page is that its list is empty.'),
  ('index', 'landing',
   'Homepage listing swapped from category-listing (unfillable) to content-listing, and needs regenerating to fill it',
   'The homepage''s category-listing has rendered an empty grid since 2026-07-14 because it resolves query.category_posts and this site has no categories; sections now name content-listing instead, which resolves query.blog_posts. CAUTION: the hero and info-card-grid copy on this page was written deliberately today against the corrected content_direction and page_spec.purpose — "Read the Specs Before You Hit the Oche", card CTAs "Read the tungsten guide →" and "See flight comparisons →". Keep that voice. The forbidden phrases in page_spec.purpose still apply: no stock, no range, no catalogue, no checkout.')
) AS v(page_name, page_role, summary, note)
ON CONFLICT DO NOTHING;

-- ── verification ───────────────────────────────────────────────────────────
-- The status will say complete either way. Check the ARTEFACT, and check the
-- timestamp that gave this away in the first place:
--
--   SELECT p.name, pc.slot_name, pc.updated_at,
--          jsonb_array_length(COALESCE(pc.content_data->'articles','[]')) AS n_articles
--   FROM page_components pc JOIN pages p ON p.id=pc.page_id
--   WHERE p.site_id='5fe8785b-223d-41a3-88ee-c07187622381'
--     AND pc.slot_name IN ('content-listing','category-listing');
--   -- expect updated_at TODAY and n_articles = 8 on both
--
-- Then on the served page, counting the links a reader can actually click:
--   curl -s https://dartsonline.com/guides/index.html | grep -oE 'href="/blog/[^"]+"' | sort -u | wc -l
--   curl -s https://dartsonline.com/ | grep -oE 'href="/blog/[^"]+"' | sort -u | wc -l
--   -- expect 8 and 8; grip-styles is the ninth guide and belongs to another
--   -- lane, so it is not deployed and correctly will not appear.
