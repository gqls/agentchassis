-- SQL — dartsonline.com: repoint the two pages that make commercial promises
-- (owner decision D4, 2026-07-29)
--
-- WHAT IS LIVE RIGHT NOW, measured on the wire today:
--
-- /about.html — <title> "About Darts Online | Specialist Darts RETAILER", body says
--   "We stock across the full range" (x2) and "We carry the manufacturers players are
--   genuinely loyal to" (x1). No stock exists.
--
-- /shipping-returns.html — the more serious of the two. It promises, in order:
--   "All orders are dispatched promptly"
--   "you'll get a tracking notification the moment your parcel's on its way"
--   "We use reliable couriers"
--   "Standard UK delivery typically lands within 2–4 working days"
--   "Express options are available at checkout"
--   "Orders placed before our daily cut-off are processed the same day"
--   "Items can be returned within 30 days of delivery"
-- There is no checkout, no order, no courier and no cut-off. This is a fabricated
-- commercial policy: a consumer-protection problem if anyone relied on it, and the
-- page an affiliate reviewer would most want to see match reality.
--
-- WHY NOT JUST ARCHIVE IT: bugs_open/098 — archiving a page does NOT undeploy it. The
-- file keeps serving 200 with the false policy on it. Replacing the CONTENT at the same
-- URL is the only move that actually removes the claims from the internet today.
--
-- WHY THE URL SURVIVES: for an affiliate site "shipping and returns" is a real reader
-- question — it is just answered by "the retailer you buy from handles both". That is
-- honest, useful, and the exact page an affiliate reviewer expects. Renaming would also
-- change the URL, orphaning the deployed file (bugs_open/125 class) for no gain.
--
-- HOW THE CONTENT IS STEERED: `pages.page_spec->>'purpose'` is read into the content
-- brief by save_page_sections_action.go:462-466. Both pages have page_spec = NULL today,
-- so the writer had only the title and the site-level content_direction to go on — and
-- until an hour ago that direction told it to write shop copy. Setting an explicit
-- purpose is the per-page half of the same fix.

BEGIN;

CREATE TABLE IF NOT EXISTS bak_darts_pages_titles_20260729 AS
SELECT id, name, title, nav_label, meta_description, page_spec
FROM pages WHERE site_id = '5fe8785b-223d-41a3-88ee-c07187622381';

-- 1. About: drop "Retailer" from the title tag and state the page's real job.
UPDATE pages
SET title = 'About Darts Online | Spec-First Darts Guides',
    meta_description = 'Darts Online is a UK online darts publication: spec-first buying '
      || 'guides on barrel weight, tungsten percentage, shaft length and flight shape. '
      || 'We hold no stock — we help you choose.',
    page_spec = COALESCE(page_spec, '{}'::jsonb) || jsonb_build_object(
      'purpose',
      'Explain what this site is and who writes it. Darts Online is a UK-based, '
      || 'online-only darts publication. It publishes spec-first buying guides and darts '
      || 'news. It does NOT sell darts, hold stock, carry brands, run a warehouse or ship '
      || 'anything, and it has no premises or trading history to describe. Say plainly '
      || 'that we do not sell — that independence is the reason the advice can be '
      || 'straight. Do not name any brand as one we stock, represent or partner with. Do '
      || 'not state an address. Contact is darts@contactforsales.com / 07934 524 911.'
    ),
    updated_at = now()
WHERE site_id = '5fe8785b-223d-41a3-88ee-c07187622381' AND name = 'about';

-- 2. Shipping & Returns: same URL, truthful subject.
UPDATE pages
SET title = 'Shipping & Returns — How Buying Works | Darts Online',
    nav_label = 'Shipping & Returns',
    meta_description = 'Darts Online does not sell darts directly. When you buy from a '
      || 'retailer, their delivery and returns terms apply — here is what to check before '
      || 'you order.',
    page_spec = COALESCE(page_spec, '{}'::jsonb) || jsonb_build_object(
      'purpose',
      'Explain honestly how buying works for a reader of this site. The single most '
      || 'important fact: Darts Online does not sell darts, take orders, process payments '
      || 'or ship anything. Any purchase happens on a retailer''s own site, and THAT '
      || 'retailer''s delivery charges, delivery times, cancellation rights and returns '
      || 'policy apply — ours do not exist. Be useful rather than apologetic: tell the '
      || 'reader what to check on a retailer''s page before ordering (delivery cost and '
      || 'timescale, the returns window, whether darts can be returned once thrown, '
      || 'whether the weight quoted is the barrel weight or the assembled dart). State '
      || 'that UK shoppers have statutory rights against the seller, not against us. '
      || 'FORBIDDEN, and previously present on this page: any promise of dispatch, '
      || 'tracking, couriers, delivery windows, express options, order cut-offs, or a '
      || 'returns period offered by us. We have no checkout to offer them from.'
    ),
    updated_at = now()
WHERE site_id = '5fe8785b-223d-41a3-88ee-c07187622381' AND name = 'shipping-returns';

-- 3. Queue both rebuilds. Copy regeneration needs page-build-handler: a page_rerender
--    either reassembles stored rendered_html or re-renders from the component template,
--    and neither writes new copy.
INSERT INTO site_work_items
  (site_id, item_type, item_key, status, pipeline, priority, handler_agent, source, spec, created_by, summary)
SELECT '5fe8785b-223d-41a3-88ee-c07187622381', 'needs_page',
       'truth_reset:' || v.page || ':5fe8785b-223d-41a3-88ee-c07187622381',
       'triaged', 'build', 45, 'page-build-handler', 'dartsonline-traffic-workstream',
       jsonb_build_object(
         'reason', 'identity_corrected',
         'plan_id', '0fb05b75-04f4-4f4c-8890-c34d6a71012c',
         'page_name', v.page,
         'page_role', 'content',
         'note', 'Rewrite against the corrected identity/briefing/content_direction and the new page_spec.purpose. This page currently makes claims that are false.'),
       'dartsonline-traffic-workstream',
       'Rebuild ' || v.page || ' — live page makes commercial claims that are not true'
FROM (VALUES ('about'), ('shipping-returns')) AS v(page);

COMMIT;

SELECT name, title, left(page_spec->>'purpose', 70) AS purpose_starts
FROM pages
WHERE site_id = '5fe8785b-223d-41a3-88ee-c07187622381'
  AND name IN ('about','shipping-returns');
