-- SQL_2026-07-29p_imagery_style_guide_and_prompts.sql
--
-- THE FIFTH HOME OF THE FALSE PREMISE — the imagery plan — plus the source of a
-- defect I verified by looking rather than by inference.
--
-- MEASURED, NOT ASSUMED. Two things were checked against the artefacts before
-- anything here was written.
--
-- (1) HOW MUCH IMAGERY THE SITE ACTUALLY USES. 33 assets exist (14 hero, 17
--     icon, favicon, og_card). The served homepage contains exactly ONE <img>
--     (the logo) and one CSS background-image (hero-home.jpg), and four emoji
--     where a card grid's icons should be: 🏆 📐 🖐 ⚖. So "33 assets barely
--     referenced" is 2 of 33 on the front page, counted rather than estimated.
--
-- (2) WHY THE 17 ICONS CANNOT SHIP. I fetched one and looked at it: mid-grey
--     linework on a near-white ground. The site background is #111520. But the
--     interesting part is that this is not a generation accident — every icon
--     prompt SAYS SO, verbatim:
--
--       "a darker grey (#4A4A4A) line on a flat solid light grey (#EEEEEE)
--        background"
--
--     So the 17 stranded `undeployed_asset` items are not a queue problem to
--     drain; they are 17 correct renderings of a wrong instruction. Deploying
--     them would put white tiles on a stadium-dark site. They stay undeployed
--     and the PROMPTS are fixed instead — otherwise the next generation round
--     produces 17 more of the same.
--
-- AND THE FALSE PREMISE, AGAIN, IN A FIFTH PLACE. The imagery plan describes a
-- shop:
--     icon_free_shipping     "a delivery truck with a checkmark"
--     icon_specialist_range  "a curated product range"
--     hero_shipping          "a cardboard shipping box partially open,
--                             revealing protective foam padding"
-- This site ships nothing, holds no range and has no checkout. Left alone, the
-- next imagery run would spend real generation credits drawing a delivery truck
-- for a publication that has never posted anything. Identity, briefing,
-- classification, content_direction, page_spec.purpose and site_plan_pages were
-- the first five readers to be corrected; this is the sixth surface and the
-- second one that WRITES rather than reads.
--
-- WHAT THIS FILE DOES
--   1. Creates the `imagery_style_guide` spec, which is the gate every later
--      generation reads. Anchored to hero_home, which I fetched from the live
--      site and looked at: a tight close-up of darts grouped in a board, dark
--      ground, shallow depth of field — genuinely good, and exactly the
--      reference the rest should match.
--   2. Rewrites all seven icon prompts onto dark ground in the palette's own
--      values, keeping the "one uniform background / no gradients / no shadows /
--      no checkerboard / no transparency / no photorealism" guards, which are
--      clearly there because somebody was bitten.
--   3. Replaces icon_free_shipping (untrue) with icon_independent (true, and the
--      site's actual selling point).
--   4. Rewrites the three retail hero prompts to match what those pages now are.
--   5. Repoints hero_guides from the archived `guides` page to `guides-index`.
--
-- NOT DONE HERE: generating anything. No image is generated, no credit spent,
-- nothing is wired into a page. This is the instruction set the next round reads.
-- Read every generated PNG before wiring it — green signals lie.
--
-- Site: dartsonline.com  5fe8785b-223d-41a3-88ee-c07187622381
-- Plan: 0fb05b75-04f4-4f4c-8890-c34d6a71012c (current)

BEGIN;

CREATE TABLE IF NOT EXISTS bak_20260729p_site_plan_imagery AS
SELECT * FROM site_plan_imagery
WHERE plan_id = '0fb05b75-04f4-4f4c-8890-c34d6a71012c';

-- ── 1. the style guide spec (supersede-then-insert; none exists yet) ───────
UPDATE site_specs SET is_current = false, superseded_at = now()
WHERE site_id = '5fe8785b-223d-41a3-88ee-c07187622381'
  AND aspect = 'imagery_style_guide' AND is_current;

INSERT INTO site_specs (site_id, aspect, data, source, source_agent, created_by, notes)
VALUES (
  '5fe8785b-223d-41a3-88ee-c07187622381',
  'imagery_style_guide',
  jsonb_build_object(
    'medium', 'photographic close-up of darts equipment, shot dark: shallow depth of field, a single subject sharp against an unlit background, hard directional light picking out metal and grip texture',
    'mood', 'stadium-night, precise, competitive; the look of a lit board in a dim venue rather than a lit product on a white table',
    'palette', 'near-black and deep navy grounds (#111520 background, #1E2436 surface); board red and green where a board is in shot; a single vivid red-orange accent (#E8311A); light text-grey highlights (#F0F2F7); no bright white fields',
    'avoid', 'white, pale or light-grey backgrounds of any kind — the site ground is #111520 and a pale tile reads as a hole in the page (all 17 existing icons were generated pale because the prompts asked for #EEEEEE); manufacturer logos, brand marks or recognisable product branding — we have no relationship with any manufacturer and a rendered logo is a claim we cannot make; text, lettering, numerals or watermarks; delivery vans, parcels, boxes, packaging, shopping baskets, price tags, checkouts or anything implying we sell or ship; multi-panel collages or grids of small scenes; pastel palettes; stock-photo people playing darts casually; identifiable faces',
    'kinds', jsonb_build_object(
      'icon', jsonb_build_object(
        'medium', 'flat single-weight linework, one subject, centred',
        'palette', 'light grey (#F0F2F7) line on a flat solid dark ground (#1E2436), with #E8311A permitted for one emphasis stroke only',
        'avoid', 'light or white grounds, gradients, drop shadows, transparency, checkerboard, photorealism, more than one subject, text'
      ),
      'hero', jsonb_build_object(
        'medium', 'close-up photography, one subject, shallow depth of field',
        'palette', 'dark ground, metal and board colour carrying the image',
        'avoid', 'flat-lay arrangements of many items, white or seamless studio backgrounds, anything resembling a catalogue shot'
      ),
      'content_hero', jsonb_build_object(
        'medium', 'close-up photography of the single component the article is about — the barrel, the flight, the shaft, the board face',
        'palette', 'dark ground, one sharp subject',
        'avoid', 'people, hands holding product towards camera, white grounds, composite scenes'
      )
    ),
    'reference_asset_keys', jsonb_build_array('hero_home')
  ),
  'hand_authored',
  'dartsonline-traffic-workstream',
  'dartsonline-traffic-workstream',
  'Written 2026-07-29. Anchored on hero_home after fetching and LOOKING at the live /assets/images/hero-home.jpg. The avoid list is not generic: every clause names something already present in this site''s own imagery plan or output — the #EEEEEE ground that produced 17 unusable icons, and the delivery-truck/shipping-box prompts for a site that ships nothing.'
);

-- ── 2. icon prompts: dark ground, in the palette's own values ──────────────
-- The three no-* guards are preserved verbatim from the originals. They read
-- like scar tissue and there is no reason to find out why the hard way.
UPDATE site_plan_imagery SET
  prompt = replace(
    prompt,
    'a darker grey (#4A4A4A) line on a flat solid light grey (#EEEEEE) background',
    'a light grey (#F0F2F7) line on a flat solid dark navy (#1E2436) background'
  )
WHERE plan_id = '0fb05b75-04f4-4f4c-8890-c34d6a71012c'
  AND kind = 'icon'
  AND prompt LIKE '%#EEEEEE%';

-- ── 3. the untrue icon, replaced rather than reworded ──────────────────────
-- Keeping the key `icon_free_shipping` on an icon that no longer means free
-- shipping is how a config key comes to mean something nobody can derive from
-- its name. Delete and insert.
DELETE FROM site_plan_imagery
WHERE plan_id = '0fb05b75-04f4-4f4c-8890-c34d6a71012c' AND key = 'icon_free_shipping';

INSERT INTO site_plan_imagery (plan_id, scope, scope_ref, key, kind, prompt, style_hints, ordering, source)
VALUES (
  '0fb05b75-04f4-4f4c-8890-c34d6a71012c', 'section', 'index:2', 'icon_independent', 'icon',
  'A single minimalist flat icon representing independence — a set of balance scales with a dart resting in one pan, line illustration style, a light grey (#F0F2F7) line on a flat solid dark navy (#1E2436) background, one single uniform background colour, no gradients, no shadows, no checkerboard, no transparency, no photorealism',
  -- source must be one of llm|classifier|manual|adoption (chk_source). A first
  -- attempt used 'hand_authored', which is the vocabulary site_specs.source
  -- takes, and the whole transaction rolled back — the two tables do not share
  -- a source vocabulary and nothing says so at either site.
  '{"aspect_ratio": "1:1"}'::jsonb, 20, 'manual'
);

-- Same problem, milder: "a curated product range" on a site with no range.
UPDATE site_plan_imagery SET
  prompt = 'A single minimalist flat icon representing spec-first advice — a pair of calipers measuring a dart barrel, line illustration style, a light grey (#F0F2F7) line on a flat solid dark navy (#1E2436) background, one single uniform background colour, no gradients, no shadows, no checkerboard, no transparency, no photorealism'
WHERE plan_id = '0fb05b75-04f4-4f4c-8890-c34d6a71012c' AND key = 'icon_specialist_range';

-- icon_all_brands keeps its key and its four barrel silhouettes — coverage
-- across manufacturers is true and is not a claim to stock any of them. The
-- prompt is made to say so, so a future generator cannot read "brand variety"
-- as "brands we carry".
UPDATE site_plan_imagery SET
  prompt = 'A single minimalist flat icon representing coverage across manufacturers with no allegiance to any — four different barrel profile silhouettes side by side, no logos or brand marks of any kind, line illustration style, a light grey (#F0F2F7) line on a flat solid dark navy (#1E2436) background, one single uniform background colour, no gradients, no shadows, no checkerboard, no transparency, no photorealism'
WHERE plan_id = '0fb05b75-04f4-4f4c-8890-c34d6a71012c' AND key = 'icon_all_brands';

-- ── 4. the three retail hero prompts ──────────────────────────────────────
UPDATE site_plan_imagery SET
  prompt = 'A close-up of a single tungsten barrel held at an angle so the machined grip rings and the laser-etched weight marking catch a hard rim light, everything behind it falling into unlit darkness. Shallow depth of field, cold metal against near-black. No packaging, no price tags, no branding, no text.'
WHERE plan_id = '0fb05b75-04f4-4f4c-8890-c34d6a71012c' AND key = 'hero_sale';

UPDATE site_plan_imagery SET
  prompt = 'A first set of three darts laid out unassembled on a dark surface — barrels, shafts and flights separated, arranged as if about to be put together for the first time. Hard directional light, deep shadow, near-black ground. No packaging, no branding, no text, no hands.'
WHERE plan_id = '0fb05b75-04f4-4f4c-8890-c34d6a71012c' AND key = 'hero_new_arrivals';

UPDATE site_plan_imagery SET
  prompt = 'A dart tip resting against the wire of a board segment, extreme close-up, so the point and the wire edge are the only things sharp. Dark ground, board red and green blurred behind. Nothing about parcels, packaging or delivery — this page explains what to check before buying from someone else. No text, no branding.'
WHERE plan_id = '0fb05b75-04f4-4f4c-8890-c34d6a71012c' AND key = 'hero_shipping';

-- ── 5. repoint the hub hero at the page that exists ───────────────────────
-- `guides` was archived on 2026-07-29 as an orphan of a superseded plan; the
-- live hub is `guides-index` at /guides/index.html. `shop` and `brands` are
-- deliberately left pointing at their archived pages: those hubs are not being
-- built until an affiliate feed exists, and repointing them would be a change
-- that reads as a decision to build them.
UPDATE site_plan_imagery SET scope_ref = 'guides-index'
WHERE plan_id = '0fb05b75-04f4-4f4c-8890-c34d6a71012c' AND key = 'hero_guides';

COMMIT;

-- ── verification ───────────────────────────────────────────────────────────
-- The point of this file is that NOTHING pale and NOTHING retail survives in the
-- instructions. Both are one query, and both must return 0:
--
--   SELECT key, prompt FROM site_plan_imagery
--   WHERE plan_id='0fb05b75-04f4-4f4c-8890-c34d6a71012c'
--     AND (prompt ILIKE '%#EEEEEE%' OR prompt ILIKE '%light grey (#EEE%'
--          OR prompt ILIKE '%white background%' OR prompt ILIKE '%crisp white%');
--
--   SELECT key, prompt FROM site_plan_imagery
--   WHERE plan_id='0fb05b75-04f4-4f4c-8890-c34d6a71012c'
--     AND (prompt ILIKE '%shipping%' OR prompt ILIKE '%delivery%'
--          OR prompt ILIKE '%parcel%' OR prompt ILIKE '%product range%');
--
-- CAREFUL — the second one is exactly the trap this site has already sprung
-- twice: my own new prompts CONTAIN the words "packaging", "price tags" and
-- "delivery" inside their negative clauses ("No packaging, no price tags"), and
-- a naive ILIKE will match its own prohibition. Read the rows, do not just count
-- them. That is the third time on this site: honesty_rails matched "stock",
-- cta_style.never_use matched "Add to Bag".
--
-- And the guide itself:
--   SELECT jsonb_pretty(data) FROM site_specs
--   WHERE site_id='5fe8785b-223d-41a3-88ee-c07187622381'
--     AND aspect='imagery_style_guide' AND is_current;
--   SELECT count(*) FROM site_specs
--   WHERE site_id='5fe8785b-223d-41a3-88ee-c07187622381'
--     AND aspect='imagery_style_guide' AND is_current;   -- must be exactly 1
