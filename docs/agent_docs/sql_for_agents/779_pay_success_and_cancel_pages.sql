-- 779_pay_success_and_cancel_pages.sql — 2026-09-04, stripe_and_payments lane
--
-- WHY: a customer who pays for a website on webdesign.uk lands on a bare 404.
-- `internal/auth-service/billing/stripe.go:55-56` mints every checkout's
-- landing as `{billing_public_base_url}/pay/success?o=<order uuid>` (and
-- /pay/cancel), `publicBaseURL` defaults to https://webdesign.uk
-- (cmd/auth-service/main.go:148), and neither page has ever existed. Found
-- 2026-08-27 by trial run 1 when the owner hit it live; still true on
-- 2026-09-04. Nothing loses money — the webhook is the truth, not the redirect
-- — but it is the worst possible post-purchase moment, and the owner is about
-- to run a second end-to-end trial (paper-cups.com, voucher WD-KN3WU-9PZN4).
--
-- WHY THESE URLs NEED NO CODE CHANGE — [MEASURED 2026-09-04 16:1xZ]. The box
-- vhost is `try_files $uri $uri/ =404` with `index index.html`
-- (webdesign_uk_build_service/box/webdesign.uk.nginx), and this site already
-- serves twelve pages from subdirectories. So an extensionless directory URL
-- already resolves, proven against a live analogue rather than reasoned:
--     /tools/css-variables            -> 200   (extensionless, no trailing slash)
--     /tools/css-variables/           -> 200
--     /tools/css-variables/index.html -> 200
--     /pay/success                    -> 404
--     /invented-dir-probe             -> 404   (control: the host 404s invented paths)
-- Therefore a page at /pay/success/index.html is served at /pay/success, and
-- NO nginx location, NO `billing_public_base_url` change and NO Go change is
-- needed. That last point is load-bearing: Go is inert until the next roll, and
-- the owner's trial is before it.
--
-- THROUGH THE FRAMEWORK, per the owner ruling of 2026-08-04. These are two
-- `needs_content_page` work items on the site's own build pipeline, handled by
-- `page-build-handler` — the same path that has completed 365 of these. This
-- migration writes NO html, NO rendered_html and NO page rows; it asks the
-- framework to build them. A pay-success page is precisely the "however small,
-- however temporary" case that ruling was written about, and it is being
-- resisted on the one lane whose product is framework-built sites.
--
-- WHAT THE COPY MAY AND MAY NOT SAY — the constraints are in each `suggestion`
-- and every one of them is a measured fact, not a preference:
--   * NO TIMESCALE. Delivery is gated on a human approval (`delivery-review-filer`
--     files a checkpoint; `HandleApproveWorkItem` is the sole writer of the
--     predicate the delivery gate reads). There is no SLA, so any number would
--     be invented.
--   * DO NOT SAY THE BUILD HAS STARTED. `collect_external_orders_action.go`
--     routes two paid cases to `needs_human_review` instead of `build_queue` —
--     a brief naming no domain, and a domain whose row is past 'queued'. Those
--     customers see this same page.
--   * DO NOT DISPLAY A REFERENCE. `?o=` carries `billing_orders.id`, a UUID —
--     verified in the one real stored webhook payload:
--     success_url ...?o=36744bf0-ca85-465a-9d73-d8176bd2525d while that order's
--     external_reference is BR-9AUZ59. A static page cannot map one to the other.
--   * NO LINK TO VISIT, and nothing about what the delivered files contain
--     (that sentence is already false on the delivery email and is the owner's
--     "leave it" until the instructions page exists).
--
-- ⚠ POSITIVE FORM IS A GATE REQUIREMENT, NOT A STYLE NOTE: the claims gate
-- reads a bare "no" as an INTENSIFIER — "there is no refund" scans as a refund
-- PROMISE and blocks the page (LANDMINES, webdesign.uk lane). The cancel page
-- is the one at risk, because "you have not been charged" is its central fact.
-- Its suggestion states this explicitly so the writer reaches for the positive
-- construction first.
--
-- ROLLBACK: 779_..._ROLLBACK.sql deletes the two items by item_key. If the
-- framework has already built the pages, the ROLLBACK says how to retire them
-- and deliberately does NOT delete them blind.

\set ON_ERROR_STOP on
BEGIN;

WITH site AS (
  SELECT id FROM sites WHERE domain = 'webdesign.uk'
)
INSERT INTO site_work_items
  (site_id, source, item_type, item_key, severity, pipeline, priority,
   created_by, handler_agent, status, summary, spec)
SELECT site.id, 'owner-request', 'needs_content_page', v.item_key, 'medium', 'build', 40,
       'stripe_and_payments_lane', 'page-build-handler', 'triaged', v.summary, v.spec
FROM site, (VALUES
  (
    'needs_content_page:pay-success:2026-09-04',
    'Build /pay/success/ — the page a customer sees straight after paying, which today is a 404',
    jsonb_build_object(
      'check', 'gap_plan_new_page',
      'page_url', '/pay/success/index.html',
      'page_name', 'pay-success',
      'reason', 'Every Stripe checkout this platform creates sends the buyer to '
             || '{billing_public_base_url}/pay/success?o=<order id> when they finish paying '
             || '(internal/auth-service/billing/stripe.go:55). No such page has ever existed, so a '
             || 'customer who has just paid real money lands on a bare 404. Confirmed live 2026-09-04 '
             || 'and in the stored webhook payload of the one real payment taken (2026-08-27).',
      'sections', jsonb_build_array('hero', 'Generic Text Block'),
      'suggestion',
         'Write the short page a customer sees immediately after paying for their website. '
      || 'Exactly three things are true for every single reader, and the page should say those three warmly and stop: '
      || 'we have received their payment; we have their brief; and we will email them when their site is ready. '
      || 'Two short sections: a hero that confirms the payment has gone through, and one text block that says what happens next in plain prose. '
      || 'HARD CONSTRAINTS, each for a measured reason. '
      || 'Give NO timescale of any kind — not days, not weeks, not "shortly", not "soon" — because delivery waits on a human approval step and any number would be invented. '
      || 'Do NOT say the build has started or is underway: some paid orders wait for a person instead, and those customers see this page too. '
      || 'Offer NO link for them to click or visit — none exists yet. '
      || 'Do NOT show an order number or reference: the web address carries an internal identifier rather than the BR- reference from their brief, so the page cannot display anything meaningful. If a reference is worth mentioning at all, invite them to quote the one from their brief when they get in touch. '
      || 'Say nothing about what the delivered files contain. Say nothing about price, refunds, hosting or domains. '
      || 'Write every sentence in the positive form. Follow the site content_direction and the house voice: British English, plain human prose, short paragraphs, no marketing language and no exclamation marks. This is a receipt, not a sales page — calm and brief is the whole brief.'
    )
  ),
  (
    'needs_content_page:pay-cancel:2026-09-04',
    'Build /pay/cancel/ — where Stripe returns someone who leaves checkout without paying; today a 404',
    jsonb_build_object(
      'check', 'gap_plan_new_page',
      'page_url', '/pay/cancel/index.html',
      'page_name', 'pay-cancel',
      'reason', 'The cancel half of the same defect: stripe.go:56 sends anyone who backs out of '
             || 'checkout to {billing_public_base_url}/pay/cancel?o=<order id>, and that page has '
             || 'never existed either. Someone who hesitated over a purchase is shown a 404, which '
             || 'reads as a broken site at the exact moment they were already unsure.',
      'sections', jsonb_build_array('hero', 'Generic Text Block'),
      'suggestion',
         'Write the short page someone sees when they leave the payment page without completing payment. '
      || 'The reader is someone who hesitated, so the tone is unbothered and helpful rather than disappointed or persuasive. '
      || 'What is true: their card was not charged, their brief is safe and still with us, and they can pay whenever they are ready by replying to us. '
      || 'Two short sections: a hero, and one text block telling them how to pick it up again. '
      || 'HARD CONSTRAINTS. Apply no pressure and no urgency — no "do not miss out", no deadline, no discount, no reason-to-hurry of any kind. '
      || 'Give NO timescale. Offer NO link to retry payment: the checkout link is created per order and this page cannot generate one, so invite them to reply to their email or use the contact page instead. '
      || 'Do NOT show an order number: the web address carries an internal identifier, not their brief reference. '
      || 'Say nothing about price, refunds, hosting or domains. '
      || 'IMPORTANT — WRITE THE REASSURANCE IN THE POSITIVE. The central fact is that no money has left their account, and a bare negative ("you have not been charged", "there is no charge") is misread by this platform''s claims gate as the very promise it denies, which blocks the page. Prefer positive constructions such as "your card is untouched", "your money stays where it is", or "payment is still to be made". '
      || 'Follow the site content_direction and the house voice: British English, plain human prose, short paragraphs, no marketing language.'
    )
  )
) AS v(item_key, summary, spec)
WHERE NOT EXISTS (
  SELECT 1 FROM site_work_items x WHERE x.item_key = v.item_key
);

-- Verify with DO/RAISE, never a bare SELECT: ON_ERROR_STOP does NOT abort on a
-- non-empty result set, so a SELECT-based check cannot stop the COMMIT
-- (LANDMINES / RFC_006).
DO $$
DECLARE
  n_items    int;
  n_wrongsite int;
  v_site     uuid;
BEGIN
  SELECT id INTO v_site FROM sites WHERE domain = 'webdesign.uk';
  IF v_site IS NULL THEN
    RAISE EXCEPTION 'webdesign.uk has no sites row — nothing was seeded';
  END IF;

  SELECT count(*) INTO n_items FROM site_work_items
   WHERE item_key IN ('needs_content_page:pay-success:2026-09-04',
                      'needs_content_page:pay-cancel:2026-09-04');
  IF n_items <> 2 THEN
    RAISE EXCEPTION 'expected exactly 2 pay-page work items, found %', n_items;
  END IF;

  SELECT count(*) INTO n_wrongsite FROM site_work_items
   WHERE item_key IN ('needs_content_page:pay-success:2026-09-04',
                      'needs_content_page:pay-cancel:2026-09-04')
     AND (site_id <> v_site OR status <> 'triaged'
          OR handler_agent <> 'page-build-handler' OR pipeline <> 'build');
  IF n_wrongsite > 0 THEN
    RAISE EXCEPTION '% pay-page item(s) are on the wrong site or not dispatchable', n_wrongsite;
  END IF;

  RAISE NOTICE 'OK: 2 needs_content_page items triaged for webdesign.uk (%), handler page-build-handler', v_site;
END $$;

COMMIT;
