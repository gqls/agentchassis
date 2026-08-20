-- bugs_open/299 (slug home_page_cta_names_the_brief_starter_tool_and_dials_the_phone_instead)
-- The last two CTAs on webdesign.uk that the framework could NOT fix by itself,
-- both now answered by the owner (2026-08-20).
--
-- ── WHY THESE TWO AND NOT THE OTHER THREE ────────────────────────────────────
-- The tel: normaliser has already self-repaired every case it could read
-- unambiguously, as each page rerendered through the fixed keep branch:
--   faq/hero              tel:+44 (0) 7934 524 911 -> tel:+447934524911  (auto, 08-19 20:39)
--   faq/call-to-action    tel:+44 (0) 7934 524 911 -> tel:+447934524911  (auto, 08-19 20:39)
--   index/call-to-action  tel:+44 (0) 7934 524 911 -> tel:+447934524911  (auto, 08-19 20:39)
--   how-it-works/call-to-action                    -- still raw; will self-heal on its
--                                                     next rerender. NOT touched here:
--                                                     it is a GENUINE phone CTA and the
--                                                     machine can fix it without a human.
-- What remains are exactly the two the code deliberately refused to guess.
--
-- ── (1) contact/hero — the undialable number ─────────────────────────────────
-- `tel:+4407934524911` has the "(0)" trunk digit collapsed INTO the number, so no
-- phone can dial it. NormalizeTelHref REFUSES this shape by design rather than
-- inventing digits (a "+440…" result is the tell), and check_cta_nonpage filed it
-- as `cta_tel_malformed` with the reason "a human must state the intended number".
-- **The owner stated it, 2026-08-20: +44 7934 524 911 — CONFIRMED.**
-- It also matches the display text the page has always carried.
--
-- ── (2) index/call-to-action — the unintended phone button ───────────────────
-- This is bug 299 itself. **The owner has confirmed it was never an intentional
-- phone button** (2026-08-20). It appears to be a leftover: on 2026-08-13 that
-- section read "Prefer to talk it through first? Call +44 (0) 7934 524 911 or
-- email…", which IS a phone CTA; the copy was then rewritten four times and the
-- href never moved.
--
-- ⚠ AND IT WOULD NOW BE PRESERVED FOR EVER IF LEFT ALONE. The bugs_open/299 fix
-- teaches the framework that a tel: destination is AUTHORED and must be kept —
-- which is what stops the faq and how-it-works "call us" buttons being destroyed.
-- The framework cannot distinguish authored-on-purpose from inherited-by-accident
-- (there is no provenance field on content_data; that is bugs_open/248's finding
-- and bugs_open/308's candidate 1). So the keep will faithfully defend a mistake
-- until a human removes it. This file is that human.
--
-- DESTINATION CHOSEN = /faq.html, i.e. what the button's own copy already says:
--   "Read the full terms in our FAQ before you pay."
-- Not the Brief Starter, because primary_cta_url on the same component is ALREADY
-- /tools/website-brief-starter/index.html — pointing both buttons at one page is
-- not a call to action, it is a duplicate.
--
-- ── WHY THIS STICKS (both write paths agree, so nothing has to guess) ────────
--   BUILD path  (setCTAField):     label-match runs FIRST and the copy names the
--                                  FAQ, so it resolves to /faq.html — the same
--                                  answer this file writes.
--   REPAIR path (applyCTARecompute): KEEP #2 preserves a stored url that is a
--                                  valid, non-utility page and not the page
--                                  itself. /faq.html is all three.
-- Contrast with today's value: a tel: takes KEEP #3 on the repair path for ever.
--
-- ── VERIFY (and the false-pass trap) ─────────────────────────────────────────
--   curl -s https://preview.webdesign.uk/index.html | grep -A3 'cta-btn-secondary'
-- Assert on the anchor whose TEXT names a destination. The nav and footer already
-- link the Brief Starter correctly, so a page-wide grep for a URL passes today
-- while the button is broken. The page must be redeployed for this to show.
--
-- ROLLBACK: the previous values are in the WHERE clauses below; swap them.

BEGIN;

-- (1) the undialable number, now confirmed by the owner
UPDATE page_components pc
   SET content_data = jsonb_set(pc.content_data, '{secondary_cta_url}', '"tel:+447934524911"'::jsonb),
       updated_at = now()
  FROM pages p, sites s
 WHERE pc.page_id = p.id AND p.site_id = s.id
   AND s.domain = 'webdesign.uk' AND p.name = 'contact' AND p.status = 'active'
   AND pc.slot_name = 'hero'
   -- anchored on the exact broken value, so this cannot double-apply or fire on
   -- a row somebody has already repaired differently:
   AND pc.content_data->>'secondary_cta_url' = 'tel:+4407934524911';

-- (2) bug 299's button: point it where its own copy says
UPDATE page_components pc
   SET content_data = jsonb_set(
         jsonb_set(pc.content_data, '{secondary_cta_url}', '"/faq.html"'::jsonb),
         '{secondary_cta_target_title}', '"Frequently Asked Questions"'::jsonb),
       updated_at = now()
  FROM pages p, sites s
 WHERE pc.page_id = p.id AND p.site_id = s.id
   AND s.domain = 'webdesign.uk' AND p.name = 'index' AND p.status = 'active'
   AND pc.slot_name = 'call-to-action'
   AND pc.content_data->>'secondary_cta_url' LIKE 'tel:%';

-- Induced verification (DO/RAISE — a SELECT cannot stop the COMMIT):
DO $$
DECLARE n int;
BEGIN
  SELECT count(*) INTO n
    FROM page_components pc JOIN pages p ON p.id = pc.page_id JOIN sites s ON s.id = p.site_id
   WHERE s.domain = 'webdesign.uk' AND p.status = 'active'
     AND p.name = 'contact' AND pc.slot_name = 'hero'
     AND pc.content_data->>'secondary_cta_url' = 'tel:+447934524911';
  IF n <> 1 THEN
    RAISE EXCEPTION '299-close (1): expected exactly 1 repaired contact/hero tel:, found % — the row drifted since the 2026-08-20 read; investigate', n;
  END IF;

  SELECT count(*) INTO n
    FROM page_components pc JOIN pages p ON p.id = pc.page_id JOIN sites s ON s.id = p.site_id
   WHERE s.domain = 'webdesign.uk' AND p.status = 'active'
     AND p.name = 'index' AND pc.slot_name = 'call-to-action'
     AND pc.content_data->>'secondary_cta_url' = '/faq.html'
     AND pc.content_data->>'primary_cta_url' = '/tools/website-brief-starter/index.html';
  IF n <> 1 THEN
    RAISE EXCEPTION '299-close (2): index call-to-action is not (secondary=/faq.html, primary=brief-starter) — found % matching rows; investigate before assuming this applied', n;
  END IF;

  -- The control that matters: the GENUINE phone buttons must be untouched.
  SELECT count(*) INTO n
    FROM page_components pc JOIN pages p ON p.id = pc.page_id JOIN sites s ON s.id = p.site_id
   WHERE s.domain = 'webdesign.uk' AND p.status = 'active'
     AND p.name IN ('faq','how-it-works')
     AND pc.content_data->>'secondary_cta_url' LIKE 'tel:%';
  IF n <> 3 THEN
    RAISE EXCEPTION '299-close (control): expected the 3 GENUINE phone CTAs on faq/how-it-works to survive untouched, found % — this file must never remove a real call-us button', n;
  END IF;
END $$;

COMMIT;

-- Post-apply: the page must be REDEPLOYED for the served HTML to change. Until
-- then the live page still shows the old href and that is expected, not failure.
