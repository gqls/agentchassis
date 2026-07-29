-- SQL_2026-07-29o_meta_descriptions.sql
--
-- Meta descriptions for every page that will exist. Eighteen of twenty-one
-- pages had none: only index, about and shipping-returns were authored, and
-- those three were authored by hand earlier today. `assemblePage` emits
-- content="" when the column is empty, so every other page has been shipping a
-- blank description tag — which is worse than no tag, because a search engine
-- treats it as an authored answer to "what is this page".
--
-- Written BEFORE batch A builds rather than after, so the builds inject them
-- instead of needing a second pass.
--
-- Both surfaces are written, because they are read by different things and
-- disagreeing is how one silently reverts the other:
--   pages.meta_description            what the renderer injects now
--   site_plan_pages.meta_description  what a reconcile rebuilds pages FROM
-- The plan half is the same lesson as SQL ...k: the plan regenerates the pages,
-- so fixing only the pages fixes only until the next reconcile.
--
-- HOUSE RULES APPLIED. British English. No claim to stock, sell, carry, ship or
-- discount anything. Where a description touches manufacturers it says plainly
-- that we have no relationship with them — that independence is the reason to
-- trust the comparison, so it is stated rather than hidden. Each is written to
-- be the sentence a searcher reads before deciding to click, not a keyword list.
--
-- NOT WRITTEN, deliberately: shop-index, brands-index, brand-detail,
-- product-detail and tool-setup-builder. The first four are retail furniture
-- with nothing to put in them until an affiliate feed exists, and the fifth is
-- unbuilt. A description for a page that does not exist is a claim about a page
-- that does not exist.
--
-- Site: dartsonline.com  5fe8785b-223d-41a3-88ee-c07187622381

BEGIN;

CREATE TABLE IF NOT EXISTS bak_20260729o_meta AS
SELECT id, name, meta_description FROM pages
WHERE site_id = '5fe8785b-223d-41a3-88ee-c07187622381';

UPDATE pages p SET meta_description = v.md, updated_at = now()
FROM (VALUES
  ('barrel-weight',
   'How much your dart should weigh and what changes when it does: 21-26g for steel tip, why 22-24g suits most players, and how weight shifts your release.'),
  ('beginners',
   'Starting darts? What you actually need first, what can wait, and the three specs - weight, tungsten, grip - that decide how a set feels in the hand.'),
  ('board-setup',
   'Dartboard height and oche distance, exactly: 1.73m to the bullseye, 2.37m throw, 2.93m diagonal. Plus mounting, lighting and protecting the wall.'),
  ('brand-comparison',
   'Red Dragon, Winmau, Harrows, Target and Unicorn compared on what each range is known for. We have no relationship with any of them and sell none of it.'),
  ('flight-shapes',
   'Standard, slim, kite and pear flights explained: how surface area changes drag and lift, where the dart sits in the board, and why shafts matter too.'),
  ('grip-styles',
   'Ring, razor, shark and knurled grips compared: how much grip you actually need, what it does to your release, and why the wrong one ruins a good barrel.'),
  ('guides-index',
   'Every Darts Online buying guide in one place: barrel weight, tungsten percentage, shaft length, flight shape, grip, board setup and steel versus soft tip.'),
  ('shaft-length',
   'Short, medium and long dart shafts: how length moves the balance point, changes the flight angle, and pairs with flight shape. What to change first.'),
  ('steel-tip-vs-soft-tip',
   'Steel tip or soft tip: the first choice you make and the one everything else follows from. Boards, weights, venues, and what each format rewards.'),
  ('tungsten-guide',
   '80%, 90% or 95% tungsten: what the percentage actually buys. Higher density means a slimmer barrel at the same weight, which means tighter grouping.'),
  ('contact',
   'Get in touch with Darts Online, a UK online darts publication. Questions about kit, corrections to a guide, or anything you think we have got wrong.')
) AS v(name, md)
WHERE p.site_id = '5fe8785b-223d-41a3-88ee-c07187622381' AND p.name = v.name;

UPDATE site_plan_pages spp SET meta_description = p.meta_description
FROM pages p, site_plans sp
WHERE sp.id = spp.plan_id AND sp.is_current
  AND sp.site_id = '5fe8785b-223d-41a3-88ee-c07187622381'
  AND p.site_id  = '5fe8785b-223d-41a3-88ee-c07187622381'
  AND p.name = spp.name
  AND p.meta_description IS NOT NULL AND p.meta_description <> '';

COMMIT;

-- ── verification ───────────────────────────────────────────────────────────
-- Length is the check that matters: Google truncates around 155-160 characters,
-- and a description cut mid-clause reads as carelessness on the one line a
-- searcher actually sees. Assert the number, do not eyeball it.
--
--   SELECT name, length(meta_description) AS len, meta_description
--   FROM pages WHERE site_id='5fe8785b-223d-41a3-88ee-c07187622381'
--     AND meta_description IS NOT NULL AND meta_description <> ''
--   ORDER BY len DESC;                    -- expect every len <= 160
--
--   SELECT name FROM pages
--   WHERE site_id='5fe8785b-223d-41a3-88ee-c07187622381' AND status='active'
--     AND COALESCE(meta_description,'')=''
--   ORDER BY name;                        -- expect only the 5 deliberate skips
--
-- And after the builds, on the SERVED page — the DB column is not the tag:
--   curl -s https://dartsonline.com/blog/tungsten-guide.html |
--     grep -o '<meta name="description"[^>]*>'
