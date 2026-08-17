-- SQL_2026-08-17b — the new commercial terms, owner-ruled 2026-08-17.
--
-- WHAT CHANGED (owner's words): payment comes BEFORE the build ("I don't think we do
-- build before payment"); the customer does NOT see the site before paying; after paying
-- they get the ZIP plus hosting instructions, AND a limited-time preview link (~1 month)
-- named as an added benefit; no refunds, justified "against the deal - a really cheap
-- starter site - but you get what you're given".
--
-- THREE facts change and NOTHING ELSE MAY. The other twelve are carried through by jsonb
-- surgery rather than retyped: this row is 15 facts + an 8KB writer_block, and a retype is
-- how a supersede silently drops an attestation. writer_block is edited by ANCHORED
-- replace on its own sentences, same discipline.
--
-- `payment_after_approval` is RENAMED to `payment_upfront`: an id asserting the opposite of
-- its claim is a trap for every later reader. Safe because no plan row pins it — checked:
-- webdesign.uk's only assigned_fact_ids are price_total, price_is_total_no_vat,
-- build_duration.
--
-- THE SWITCH MOVES WITH THE COPY. billing_settings.payment_timing is a REAL setting read by
-- auth-service (repository.go:247), and the old fact itself said to re-check it before
-- restating. Copy saying "pay first" while the system is set to 'after_approval' would be
-- worse than the state we are leaving. Flipped to 'upfront' in the same transaction.
--
-- CLAIMS-GATE RULE OBEYED (from writer_block itself): never "no refund" bare — the gate
-- reads a bare "no" as an intensifier, so "there is no refund" scans as a refund PROMISE
-- and blocks the page. Every refund sentence here carries "do not".

BEGIN;

WITH cur AS (
  SELECT ss.id, ss.site_id, ss.data, ss.pinned
    FROM site_specs ss JOIN sites s ON s.id = ss.site_id
   WHERE s.domain = 'webdesign.uk' AND ss.aspect = 'evidence_base' AND ss.is_current
),
rebuilt AS (
  SELECT c.site_id, c.pinned,
    jsonb_set(
      jsonb_set(c.data, '{facts}', (
        SELECT jsonb_agg(
          CASE
            WHEN f->>'id' = 'payment_after_approval' THEN jsonb_build_object(
              'id','payment_upfront','kind','capability',
              'claim','Payment is taken before the site is built. The customer does not see the site before paying; the preview link and the files come after payment.',
              'source', jsonb_build_object('sql','SELECT payment_timing FROM billing_settings','attested_by','owner, 2026-08-17 (supersedes payment_after_approval, attested 2026-08-11; the billing_settings switch was moved to upfront in the same transaction)'),
              'verified_at','2026-08-17',
              'writer_line','You pay first, and the site is built after that.')
            WHEN f->>'id' = 'no_refund' THEN jsonb_build_object(
              'id','no_refund','kind','attestation',
              'claim','We do not offer refunds. The price buys one build of a starter site, handed over as it is, and that is what the low fixed price pays for.',
              'source', jsonb_build_object('attested_by','owner, 2026-08-17 (re-justified against the deal; the previous reason, that the customer approved before paying, is retired with payment_after_approval)'),
              'verified_at','2026-08-17',
              'writer_line','We do not offer refunds. The price buys one build, handed over as it is.')
            WHEN f->>'id' = 'delivery_preview_and_zip' THEN jsonb_build_object(
              'id','delivery_preview_and_zip','kind','capability',
              'claim','After payment the customer receives a ZIP of the finished site to keep and host wherever they like, with instructions for putting it online. A preview link is also provided so they can see the site working before hosting it themselves; it stays live for about a month, and the ZIP is theirs permanently.',
              'source', jsonb_build_object('attested_by','owner, 2026-08-17 (preview named as an added benefit, roughly one month; the ZIP is the permanent copy)'),
              'verified_at','2026-08-17',
              'writer_line','You get the finished site as a ZIP to keep, with instructions for putting it online. There is a preview link too, so you can see it working straight away; it stays up for about a month.')
            ELSE f
          END ORDER BY ord)
        FROM jsonb_array_elements(c.data->'facts') WITH ORDINALITY AS t(f, ord)
      )),
      '{writer_block}',
      to_jsonb(
        replace(
          replace(
            replace(
              c.data->>'writer_block',
              'The customer pays after they have seen the finished site on a private preview link and approved it. Nothing is taken before that.',
              'The customer pays before the site is built, and does not see it before paying. After payment they get the finished site as a ZIP to keep, with instructions for putting it online themselves, and a preview link so they can see it working straight away; the preview stays up for about a month and the ZIP is permanent.'),
            'They get a private preview link and then a ZIP of the finished site, which they host wherever they like.',
            'They get a ZIP of the finished site to keep and host wherever they like, plus a preview link, live for about a month, so they can see it working before they host it.'),
          'that payment comes after approval',
          'that payment is taken before the build'
        )
      )
    ) AS newdata
  FROM cur c
),
retire AS (
  UPDATE site_specs ss SET is_current = false, superseded_at = now()
   WHERE ss.id = (SELECT id FROM cur) RETURNING 1
)
INSERT INTO site_specs (site_id, aspect, data, source, notes, is_current, created_by, pinned)
SELECT r.site_id, 'evidence_base', r.newdata, 'owner-ruling',
       'SQL_2026-08-17b: payment before build; preview as an added benefit (~1 month) alongside the permanent ZIP; no refunds re-justified against the deal. Supersedes the 2026-08-14b row.',
       true, 'webdesign_uk_build_service lane, owner ruling 2026-08-17', r.pinned
  FROM rebuilt r, retire;

-- the switch moves with the copy
UPDATE billing_settings SET payment_timing = 'upfront', updated_at = now() WHERE id = 1;

DO $$
DECLARE d jsonb; wb text; n int;
BEGIN
  SELECT ss.data INTO d FROM site_specs ss JOIN sites s ON s.id=ss.site_id
   WHERE s.domain='webdesign.uk' AND ss.aspect='evidence_base' AND ss.is_current;

  SELECT count(*) INTO n FROM jsonb_array_elements(d->'facts');
  IF n <> 15 THEN RAISE EXCEPTION 'fact count changed: % (expected 15) - a fact was dropped', n; END IF;

  IF NOT EXISTS (SELECT 1 FROM jsonb_array_elements(d->'facts') f WHERE f->>'id'='payment_upfront') THEN
    RAISE EXCEPTION 'payment_upfront missing'; END IF;
  IF EXISTS (SELECT 1 FROM jsonb_array_elements(d->'facts') f WHERE f->>'id'='payment_after_approval') THEN
    RAISE EXCEPTION 'the retired payment_after_approval id is still present'; END IF;
  -- the untouched twelve must still be there, exactly once each
  SELECT count(*) INTO n FROM unnest(ARRAY['price_total','price_is_total_no_vat','ai_built','build_duration','no_changes_included','no_lock_in','hosting_and_domain_not_included','taking_it_further','yours_to_change','queue_limited','contact','third_party_options']) k
   WHERE (SELECT count(*) FROM jsonb_array_elements(d->'facts') f WHERE f->>'id' = k) <> 1;
  IF n <> 0 THEN RAISE EXCEPTION '% untouched fact(s) lost or duplicated', n; END IF;

  wb := d->>'writer_block';
  IF position('pays after they have seen' in wb) > 0 OR position('payment comes after approval' in wb) > 0 THEN
    RAISE EXCEPTION 'writer_block still instructs the OLD terms - the wire and the facts would disagree'; END IF;
  IF position('pays before the site is built' in wb) = 0 THEN
    RAISE EXCEPTION 'writer_block did not take the new payment sentence'; END IF;
  IF length(wb) < 7000 THEN RAISE EXCEPTION 'writer_block truncated: % chars', length(wb); END IF;

  IF (SELECT payment_timing FROM billing_settings WHERE id=1) <> 'upfront' THEN
    RAISE EXCEPTION 'billing_settings.payment_timing did not move - copy and system would disagree'; END IF;
END $$;

COMMIT;
