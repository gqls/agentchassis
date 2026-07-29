-- SQL_2026-07-29r_build_batch_b.sql
--
-- Batch B. Batch A's quality was checked by READING the built pages, not by
-- reading their status:
--
--   new-arrivals  "Build Your First Setup ... how 22g versus 26g barrels, grip
--                 styles and tungsten percentages actually change your throw"
--   sale          "Spot a Genuine Darts Deal — a flashy barrel coating often
--                 hides a cheap alloy"
--   board-setup   1.73m bullseye height, 2.37m oche, and it correctly says to
--                 measure from the board FACE not the wall, because a bristle
--                 board's depth pushes you too far back. The numbers are right.
--   brand-comparison  the one flagged as most able to reinstate the false
--                 premise. It compares grip profiles, taper angles and tungsten
--                 density across Winmau, Red Dragon, Harrows and Target, and
--                 makes no claim about availability, price or our carrying any
--                 of it. Clean.
--
-- A wide-vocabulary sweep over all 11 pages with stored components — 49 phrases,
-- not the 5 that produced session 1's false "clean" — returns three hits, all
-- three the word "checkout", all three in the DARTS sense (finishing a leg on a
-- double: "tune your trajectory for a cleaner checkout"). That is the correct
-- usage and a good sign the writer knows the sport.
--
-- IN THIS BATCH
--   shaft-length, steel-tip-vs-soft-tip   the last two of the seven unbuilt
--       guides that are ours to build. grip-styles stays with the
--       gemini_content_provider lane.
--   index, about   HEAD ONLY. Both had meta descriptions over the 160-character
--       limit and were trimmed; they need the tag re-injected and nothing else.
--       These go as `page_rerender` to `page-rerender`, which reassembles from
--       stored content_data, NOT as `needs_page`, which would send them back to
--       the writer and rewrite copy that took two careful passes to get right.
--       Choosing the heavier item here would have been a self-inflicted wound.
--
-- HELD BACK: guides-index. It is a listing page, so rebuilding it before the
-- guides it lists have deployed would produce a correct-looking index of four
-- pages instead of nine. It goes in batch C, last, once the others are live.
--
-- Site: dartsonline.com  5fe8785b-223d-41a3-88ee-c07187622381

-- ── the last two guides ───────────────────────────────────────────────────
INSERT INTO site_work_items (
  site_id, source, pipeline, item_type, severity, summary, spec,
  priority, handler_agent, status, created_by, item_key, approval_mode
)
SELECT
  '5fe8785b-223d-41a3-88ee-c07187622381', 'dartsonline-traffic-workstream', 'build',
  'needs_page', 'medium', v.summary,
  jsonb_build_object(
    'page_name', v.page_name, 'page_role', 'blog-post',
    'plan_id', '0fb05b75-04f4-4f4c-8890-c34d6a71012c',
    'reason', 'not_built', 'note', v.note
  ),
  40, 'page-build-handler', 'triaged', 'dartsonline-traffic-workstream',
  'batch_b_2026_07_29:' || v.page_name || ':5fe8785b-223d-41a3-88ee-c07187622381', 'auto'
FROM (VALUES
  ('shaft-length',
   'Build shaft-length — short, medium and long, and what each does to the balance point',
   'Pairs with flight-shapes, built in batch A: shafts and flights are tuned together and neither guide should pretend the choice is independent. Say which one to change first and why (the cheap one).'),
  ('steel-tip-vs-soft-tip',
   'Build steel-tip-vs-soft-tip — the first choice, and the one everything else follows from',
   'This is the entry point to the whole guide set: the format decides board, legal weights, venue and most of the rest. It should read as the page a complete beginner lands on, and should point onward to barrel-weight and board-setup.')
) AS v(page_name, summary, note)
ON CONFLICT DO NOTHING;

-- ── head-only reassembly for the two trimmed descriptions ─────────────────
INSERT INTO site_work_items (
  site_id, source, pipeline, item_type, severity, summary, spec,
  priority, handler_agent, status, created_by, item_key, approval_mode
)
SELECT
  '5fe8785b-223d-41a3-88ee-c07187622381', 'dartsonline-traffic-workstream', 'build',
  'page_rerender', 'low',
  'Re-inject ' || v.page_name || '''s meta description (trimmed from over 160 characters)',
  jsonb_build_object(
    'page_name', v.page_name,
    'reason', 'meta_description_corrected',
    'note', 'HEAD ONLY. The body copy on this page is correct and was written deliberately today; this must reassemble from stored content_data and must not go back to the writer.'
  ),
  55, 'page-rerender', 'triaged', 'dartsonline-traffic-workstream',
  'batch_b_meta_2026_07_29:' || v.page_name || ':5fe8785b-223d-41a3-88ee-c07187622381', 'auto'
FROM (VALUES ('index'), ('about')) AS v(page_name)
ON CONFLICT DO NOTHING;

-- ── verification ───────────────────────────────────────────────────────────
--   SELECT spec->>'page_name', item_type, status, updated_at FROM site_work_items
--   WHERE item_key LIKE 'batch_b_%' ORDER BY item_type, 1;
--
-- For index and about the check is that the body did NOT change while the head
-- did. The obvious comparator does not work: page_components.content_hash is
-- NULL on all eight rows for these two pages, so a before/after on it would
-- compare nothing to nothing and pass. Use length + updated_at instead, and
-- take the baseline BEFORE queueing:
--   SELECT p.name, pc.slot_name, length(pc.rendered_html) AS len, pc.updated_at
--   FROM page_components pc JOIN pages p ON p.id=pc.page_id
--   WHERE p.site_id='5fe8785b-223d-41a3-88ee-c07187622381'
--     AND p.name IN ('index','about') ORDER BY p.name, pc.position;
-- Baseline, measured 2026-07-29 16:38Z (I typed plausible numbers here first and
-- then ran the query; every one of them was wrong. The figures below are the
-- query's):
--   about  (all updated 13:49:33Z): hero-about 2320 · about-content 3358 ·
--          differentiators 2562 · call-to-action 2227
--   index  (all updated 14:04:45Z): hero 3007 · info-card-grid 7144 ·
--          category-listing 315 · call-to-action 2336
-- If a length moves, the item went to the writer and the copy has been
-- rewritten — read the page before accepting it.
--
-- And on the served page:
--   curl -s https://dartsonline.com/index.html |
--     grep -o '<meta name="description"[^>]*>'
