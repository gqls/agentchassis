-- SQL_2026-07-29k_plan_and_retail_pages_truth.sql
--
-- THE FOURTH HOME OF THE FALSE PREMISE, and a correction to my own claim.
--
-- Session 1 closed by reporting "all nine built pages clean" of shop language.
-- That claim was WRONG, and the way it was wrong is the lesson. The sweep read
-- page_components.rendered_html for a fixed list of banned phrases ("stock",
-- "Add to Bag", "filter our ranges", "Portland", ...). It answered THAT
-- question. It did not answer the question I reported, which was whether the
-- pages still read like a shop. The served /sale.html says, today:
--
--   "We cut prices across our sale range."
--   "We move high-density tungsten barrels, shafts, and flights into
--    clearance regularly."
--   "Testing a new grip profile ... costs less when you shop the sale section."
--
-- None of those phrases were on my list, so the sweep printed "sale :: clean".
-- The cheap check that would have caught it in ten seconds was to curl the
-- served page and read it, which is what found it now. Logged in WRONG_CALLS.md.
--
-- The second finding is structural and matters more. site_plan_pages — the row
-- set a replan/reconcile rebuilds pages FROM — still carries the retail
-- identity that was corrected everywhere else:
--
--   index    title: "Darts Online | Specialist Darts Equipment & Accessories"
--   about    title: "About Darts Online | Specialist Darts Retailer"
--   sale     title: "Sale | Darts Deals & Clearance | Darts Online"
--
-- So identity, briefing, content_direction and per-page page_spec.purpose were
-- all corrected, and the PLAN would have quietly restored the lie on the next
-- reconcile. Four homes, not three.
--
-- WHAT THIS DOES
--   1. Fixes the plan's titles and nav so a reconcile cannot revert the truth.
--   2. Repurposes the two retail landing pages into honest ones that keep their
--      URLs. They are NOT archived: bugs_open/098 establishes that archiving
--      does not undeploy, so an archived /sale.html would go on serving
--      "we cut prices" to every visitor and crawler. A live page that tells the
--      truth beats an archived one that lies.
--        sale.html         -> "How to Spot a Genuine Darts Deal" (nav: Deals)
--        new-arrivals.html -> "New to Darts? Start Here"          (nav: Start Here)
--      Both keep the hero + call-to-action section shape they already have, so
--      no new section machinery is needed.
--   3. Puts Guides in the header, which it never was.
--
-- NOT done here, deliberately: shop-index and brands-index stay out of the
-- header. They are retail hubs with nothing to put in them until an affiliate
-- feed exists (PLAN phase 6). GetNavItems prunes never-deployed targets anyway,
-- so listing them would be a no-op that reads as a decision.
--
-- Site: dartsonline.com  5fe8785b-223d-41a3-88ee-c07187622381

BEGIN;

-- ── backups ────────────────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS bak_20260729k_site_plan_pages AS
SELECT spp.* FROM site_plan_pages spp
JOIN site_plans sp ON sp.id = spp.plan_id
WHERE sp.site_id = '5fe8785b-223d-41a3-88ee-c07187622381' AND sp.is_current;

CREATE TABLE IF NOT EXISTS bak_20260729k_pages AS
SELECT * FROM pages
WHERE site_id = '5fe8785b-223d-41a3-88ee-c07187622381';

-- ── 1. the plan: titles and header membership ──────────────────────────────
UPDATE site_plan_pages spp SET
  title = 'Darts Online | Spec-First Darts Buying Guides & News'
FROM site_plans sp
WHERE sp.id = spp.plan_id AND sp.is_current
  AND sp.site_id = '5fe8785b-223d-41a3-88ee-c07187622381'
  AND spp.name = 'index';

UPDATE site_plan_pages spp SET
  title = 'About Darts Online | Spec-First Darts Guides'
FROM site_plans sp
WHERE sp.id = spp.plan_id AND sp.is_current
  AND sp.site_id = '5fe8785b-223d-41a3-88ee-c07187622381'
  AND spp.name = 'about';

UPDATE site_plan_pages spp SET
  title      = 'How to Spot a Genuine Darts Deal | Darts Online',
  nav_label  = 'Deals',
  meta_description = 'What a real darts discount looks like: how RRP works on tungsten barrels, why a cheap set can still be the wrong set, and what to check before you buy.'
FROM site_plans sp
WHERE sp.id = spp.plan_id AND sp.is_current
  AND sp.site_id = '5fe8785b-223d-41a3-88ee-c07187622381'
  AND spp.name = 'sale';

UPDATE site_plan_pages spp SET
  title      = 'New to Darts? Start Here | Darts Online',
  nav_label  = 'Start Here',
  meta_description = 'Never bought a set of darts before? Start here: what actually matters in a first setup, in what order, and which guide answers each question.'
FROM site_plans sp
WHERE sp.id = spp.plan_id AND sp.is_current
  AND sp.site_id = '5fe8785b-223d-41a3-88ee-c07187622381'
  AND spp.name = 'new-arrivals';

-- Guides belongs in the header of a guides site. It never was.
UPDATE site_plan_pages spp SET
  in_header = true, nav_order = 3, nav_label = 'Guides'
FROM site_plans sp
WHERE sp.id = spp.plan_id AND sp.is_current
  AND sp.site_id = '5fe8785b-223d-41a3-88ee-c07187622381'
  AND spp.name = 'guides-index';

-- setup-builder is planned, not built. Keep it out of the header until it is,
-- or the nav gains a link that GetNavItems will prune and nobody will know why.
UPDATE site_plan_pages spp SET in_header = false
FROM site_plans sp
WHERE sp.id = spp.plan_id AND sp.is_current
  AND sp.site_id = '5fe8785b-223d-41a3-88ee-c07187622381'
  AND spp.name = 'tool-setup-builder';

-- ── 2. the pages: titles, nav, and the purpose the writer reads ────────────
UPDATE pages SET
  title            = 'How to Spot a Genuine Darts Deal | Darts Online',
  nav_label        = 'Deals',
  nav_order        = 5,
  in_header        = true,
  meta_description = 'What a real darts discount looks like: how RRP works on tungsten barrels, why a cheap set can still be the wrong set, and what to check before you buy.',
  build_status     = 'needs_rebuild',
  page_spec        = COALESCE(page_spec, '{}'::jsonb) || jsonb_build_object('purpose',
    'An evergreen buying-advice article about judging darts discounts. THIS SITE SELLS NOTHING AND RUNS NO SALE. '
    'FORBIDDEN, and currently present on this page: "We cut prices across our sale range", "We move high-density '
    'tungsten barrels, shafts, and flights into clearance regularly", "shop the sale section", "our sale range", '
    '"clearance prices", and any first-person claim to hold, price, discount, move or clear stock. '
    'Write instead about how a player judges a darts deal from the outside: what RRP actually means on tungsten '
    'barrels and how little it constrains the seller; why a 90% tungsten set at a headline discount can still be '
    'worse value than an 80% set at full price, because the tungsten percentage changes the barrel diameter and '
    'therefore how tightly you can group; what "end of line" and "discontinued" mean and why a discontinued barrel '
    'is a real risk if you ever want to replace one dart of three; why new-model launches are when last season''s '
    'sets move, and how to use that; and the checklist that stops a cheap set being the wrong set — weight, '
    'tungsten percentage, grip pattern, and whether the shafts and flights you already own will fit. '
    'Close by pointing at the barrel-weight and tungsten guides, which answer the two questions this page raises.'),
  updated_at = now()
WHERE site_id = '5fe8785b-223d-41a3-88ee-c07187622381' AND name = 'sale';

UPDATE pages SET
  title            = 'New to Darts? Start Here | Darts Online',
  nav_label        = 'Start Here',
  nav_order        = 4,
  in_header        = true,
  meta_description = 'Never bought a set of darts before? Start here: what actually matters in a first setup, in what order, and which guide answers each question.',
  build_status     = 'needs_rebuild',
  page_spec        = COALESCE(page_spec, '{}'::jsonb) || jsonb_build_object('purpose',
    'A signposting hub for somebody who has just decided to take darts seriously and does not yet know what to ask. '
    'THIS SITE SELLS NOTHING AND HAS NO STOCK, NO RANGE AND NO ARRIVALS. '
    'FORBIDDEN, and currently present on this page: any reference to new stock, new arrivals, latest ranges, product '
    'drops, "what''s landed", or a checkout. '
    'Write instead the order of decisions a new player actually faces, and say which of our guides answers each one: '
    'steel tip or soft tip first, because it decides everything after it; then barrel weight, because it is the '
    'single choice that most changes the throw; then tungsten percentage, because it sets how slim the barrel can '
    'be; then grip; then shaft length and flight shape as a pair, because they are tuned together and are the '
    'cheapest thing to change later. Say plainly what a beginner does NOT need to spend money on yet. '
    'This page is navigation with an opinion, not a lesson — keep each step short and link out to the guide that '
    'covers it properly.'),
  updated_at = now()
WHERE site_id = '5fe8785b-223d-41a3-88ee-c07187622381' AND name = 'new-arrivals';

-- Guides into the header; setup-builder out until it exists.
UPDATE pages SET in_header = true,  nav_order = 3, nav_label = 'Guides', updated_at = now()
WHERE site_id = '5fe8785b-223d-41a3-88ee-c07187622381' AND name = 'guides-index';

UPDATE pages SET in_header = false, updated_at = now()
WHERE site_id = '5fe8785b-223d-41a3-88ee-c07187622381' AND name = 'tool-setup-builder';

-- contact carries the three dead nav links and is already needs_rebuild;
-- guides-index carries them too and is not.
UPDATE pages SET build_status = 'needs_rebuild', updated_at = now()
WHERE site_id = '5fe8785b-223d-41a3-88ee-c07187622381' AND name = 'guides-index';

COMMIT;

-- ── verification (run after) ───────────────────────────────────────────────
-- Per-key, not a blob ILIKE: a purpose string that NAMES a forbidden phrase
-- would otherwise match its own prohibition. That trap has already been hit
-- twice on this site (honesty_rails matched "stock", cta_style.never_use
-- matched "Add to Bag").
--
--   SELECT name, title, nav_label, in_header, nav_order, build_status
--   FROM pages WHERE site_id='5fe8785b-223d-41a3-88ee-c07187622381'
--     AND (in_header OR name IN ('sale','new-arrivals')) ORDER BY nav_order;
--
--   SELECT spp.name, spp.title, spp.in_header, spp.nav_label
--   FROM site_plan_pages spp JOIN site_plans sp ON sp.id=spp.plan_id
--   WHERE sp.site_id='5fe8785b-223d-41a3-88ee-c07187622381' AND sp.is_current
--     AND spp.title ILIKE '%retail%' OR spp.title ILIKE '%clearance%';   -- expect 0
--
-- And after the rebuilds land, on the SERVED page, not the DB:
--   curl -s https://dartsonline.com/sale.html | grep -ciE 'clearance|cut prices|sale range|shop the sale'
