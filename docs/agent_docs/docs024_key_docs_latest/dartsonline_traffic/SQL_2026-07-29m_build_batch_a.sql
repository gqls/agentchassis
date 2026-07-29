-- SQL_2026-07-29m_build_batch_a.sql
--
-- Batch A of the content release. The nav table was rebuilt first (SQL ...l,
-- item complete 16:20Z) and now reads, verified:
--
--   primary:  Guides /guides/index.html · Start Here /new-arrivals.html · Deals /sale.html
--   utility:  Home · About · Contact · Shipping & Returns
--
-- No /shop.html, no /brands.html, no /guides.html. So every page built from here
-- bakes in a header with no dead links, which is why this file runs SECOND.
--
-- WHAT IS IN THIS BATCH AND WHY THOSE
--   sale, new-arrivals   the two pages still serving retail copy on a site that
--                        sells nothing — the most damaging pages an affiliate
--                        reviewer could open. Repurposed in SQL ...k with an
--                        explicit page_spec.purpose that NAMES the sentences to
--                        remove; the writer reads purpose directly
--                        (save_page_sections_action.go:462-466).
--   contact              carries the three dead nav links; nothing else wrong.
--   tungsten-guide,      four of the seven unbuilt buying guides. These are the
--   board-setup,         traffic engine: they are what a darts buyer actually
--   brand-comparison,    searches for, and they have existed as titles since
--   flight-shapes        6 July with no sections to build from until 29 July.
--
-- DELIBERATELY NOT IN IT
--   grip-styles          owned by the gemini_content_provider lane, which has an
--                        open item on it for its own P7 verification
--                        (9fdb87b4, created 2026-07-27). Do not compete.
--   guides-index         held for batch B, so it is rebuilt AFTER the guides it
--                        is supposed to list exist.
--   shaft-length,        held for batch B purely to keep the batch watchable —
--   steel-tip-vs-soft-tip  the first four are the quality check on the corrected
--                        content_direction across four different guide shapes.
--   shop-index,          retail hubs with nothing to put in them until an
--   brands-index,        affiliate feed exists (PLAN phase 6). Now
--   brand/product-detail in_header=false, so nothing links to them.
--
-- plan_id is the CURRENT plan 0fb05b75 (2026-07-22 18:08Z). The two pre-existing
-- needs_page items for tungsten-guide and board-setup carry bb0e5c5d, which was
-- superseded 107 minutes after it was written — they are left alone rather than
-- promoted, and these fresh items supersede them in practice.
--
-- Site: dartsonline.com  5fe8785b-223d-41a3-88ee-c07187622381

INSERT INTO site_work_items (
  site_id, source, pipeline, item_type, severity, summary, spec,
  priority, handler_agent, status, created_by, item_key, approval_mode
)
SELECT
  '5fe8785b-223d-41a3-88ee-c07187622381',
  'dartsonline-traffic-workstream',
  'build',
  'needs_page',
  v.severity,
  v.summary,
  jsonb_build_object(
    'page_name', v.page_name,
    'page_role', v.page_role,
    'plan_id',   '0fb05b75-04f4-4f4c-8890-c34d6a71012c',
    'reason',    v.reason,
    'note',      v.note
  ),
  v.priority,
  'page-build-handler',
  'triaged',
  'dartsonline-traffic-workstream',
  'batch_a_2026_07_29:' || v.page_name || ':5fe8785b-223d-41a3-88ee-c07187622381',
  'auto'
FROM (VALUES
  ('sale', 'landing', 'purpose_corrected', 30, 'high',
   'sale still serves shop copy on a site that sells nothing — rewrite to the corrected purpose',
   'The served page says "We cut prices across our sale range" and "We move high-density tungsten barrels, shafts, and flights into clearance regularly". There is no sale, no stock and no checkout. page_spec.purpose now names those sentences as forbidden and asks for an evergreen article on judging a darts discount from the outside. Title and meta_description are already corrected.'),

  ('new-arrivals', 'landing', 'purpose_corrected', 30, 'high',
   'new-arrivals promises arrivals on a site with no stock — rewrite as the Start Here hub',
   'Repurposed to "New to Darts? Start Here": a signposting hub that puts the beginner decisions in the order they actually arise and links each to the guide that answers it. page_spec.purpose forbids any reference to new stock, arrivals, ranges or drops.'),

  ('contact', 'content', 'nav_corrected', 50, 'medium',
   'contact still serves the three dead nav links',
   'Content is fine. It carries /shop.html, /brands.html and /guides.html in its header because it has not been rebuilt since the nav data was corrected. A rebuild against the new site_nav_items is the whole fix.'),

  ('tungsten-guide', 'blog-post', 'not_built', 40, 'medium',
   'Build tungsten-guide — planned since 6 July, buildable since the sections landed',
   'Nine buying guides existed as titles with no site_plan_sections rows, so five build attempts died with nothing to write. Sections were backfilled 2026-07-29 and content_direction no longer instructs shop copy. This is 80% vs 90% vs 95% tungsten: what the percentage changes about barrel diameter and therefore grouping.'),

  ('board-setup', 'blog-post', 'not_built', 40, 'medium',
   'Build board-setup — height, oche distance and mounting',
   'The one guide on this site that is pure measurement and has a single correct answer (1.73m to the bullseye, 2.37m oche, 2.93m diagonal). Get the numbers right and cite the governing body; this is the page that will be checked hardest by readers.'),

  ('brand-comparison', 'blog-post', 'not_built', 40, 'medium',
   'Build brand-comparison — Red Dragon vs Winmau vs Harrows and the rest',
   'CAUTION: this is the guide most able to reintroduce the false premise. We have no relationship with any of these manufacturers and hold none of their stock. Compare what the ranges are known for, in the same spec-first voice as the other guides, and make no claim about availability, pricing or our carrying any of it.'),

  ('flight-shapes', 'blog-post', 'not_built', 40, 'medium',
   'Build flight-shapes — standard, slim, kite and what each does to the flight path',
   'Pairs with shaft-length, which is batch B: flights and shafts are tuned together, so this page should say so rather than pretend the choice is independent.')
) AS v(page_name, page_role, reason, priority, severity, summary, note)
ON CONFLICT DO NOTHING;

-- ── verification ───────────────────────────────────────────────────────────
--   SELECT spec->>'page_name', status, attempt_count, updated_at
--   FROM site_work_items
--   WHERE item_key LIKE 'batch_a_2026_07_29:%' ORDER BY priority, spec->>'page_name';
--
-- `error` holds the LAST failure and is NOT cleared on re-claim — a stale
-- message there while status is 'complete' means nothing. Read status, and for a
-- run that says COMPLETED but failed a step, read __step_error
-- (orchestration_states.error is empty in that case — bugs_open/099).
--
-- Then, on the SERVED page and not the DB:
--   curl -s https://dartsonline.com/sale.html | grep -ciE 'clearance|cut prices|sale range|shop the sale'
--   curl -s https://dartsonline.com/new-arrivals.html | grep -ciE 'new arrivals|just landed|in stock|our range'
--   for p in sale.html new-arrivals.html contact.html; do
--     curl -s "https://dartsonline.com/$p" | grep -oE '/(shop|brands|guides)\.html' | sort -u; done   # expect empty
