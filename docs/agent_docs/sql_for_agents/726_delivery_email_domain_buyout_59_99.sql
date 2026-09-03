-- 726_delivery_email_domain_buyout_59_99.sql
--
-- The delivery email still quotes the SUPERSEDED domain buy-out price.
--
-- OWNER RULING 2026-08-26 (night), verbatim from SQL_2026-08-26e's header:
--   "the domain price is incongruent with the cheap website pricing... Let's make
--    it £59.99" — buy-out £200 -> £59.99, rental stays £10/mo.
--
-- WHY IT WAS MISSED, and it is worth stating because the same gap will recur.
-- 08-26e censused "every £200 in the live SPECS" — evidence_base x6, identity x1,
-- briefing x1, strategy x1 — and swept those exactly, with count guards. The
-- delivery EMAIL is not a site spec: it is a step config on the
-- `delivery-email-sender` agent definition, so it sat outside that census's
-- population entirely. Measured 2026-09-03: `body_template` still reads
-- "buying it outright is a one-off 200 pounds", and it is the ONLY live agent
-- config carrying the stale figure.
--
-- Nothing has shipped it: the delivery chain has never run (zero orchestrations
-- for all four delivery agents, zero customer_access_tokens, zero zips, all
-- time, measured 2026-09-03). So this corrects the price BEFORE the first send
-- rather than after.
--
-- Style note: the template deliberately spells prices in words ("10 pounds a
-- month"), so the replacement matches — "a one-off 59.99 pounds" — rather than
-- introducing a £ glyph into an email body that currently contains none.
--
-- Apply: psql -f THIS FILE ONLY (never an unscoped runner --apply).
BEGIN;
DO $$
DECLARE n int; body text;
BEGIN
  SELECT count(*) INTO n FROM agent_definitions
   WHERE type='delivery-email-sender' AND is_active
     AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
  IF n <> 1 THEN
    RAISE EXCEPTION '726 REFUSED: expected exactly 1 active delivery-email-sender, found %', n;
  END IF;
  SELECT default_config #>> '{workflow,steps,send_email,config,body_template}' INTO body
    FROM agent_definitions WHERE type='delivery-email-sender' AND is_active
     AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
  IF body IS NULL THEN
    RAISE EXCEPTION '726 REFUSED: send_email.config.body_template not found — step renamed?';
  END IF;
  IF position('a one-off 200 pounds' IN body) = 0 THEN
    RAISE EXCEPTION '726 REFUSED: the anchor "a one-off 200 pounds" is not in the live body — already fixed, or the wording moved. Read it before editing.';
  END IF;
  PERFORM snapshot_agent('delivery-email-sender',
                         '726_delivery_email_domain_buyout_59_99.sql: pre-update (body still says 200 pounds)');
END $$;
UPDATE agent_definitions
   SET default_config = jsonb_set(default_config,
         '{workflow,steps,send_email,config,body_template}',
         to_jsonb(replace(default_config #>> '{workflow,steps,send_email,config,body_template}',
                          'a one-off 200 pounds', 'a one-off 59.99 pounds')), false),
       updated_at = now()
 WHERE type='delivery-email-sender' AND is_active
   AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
DO $$
DECLARE body text;
BEGIN
  SELECT default_config #>> '{workflow,steps,send_email,config,body_template}' INTO body
    FROM agent_definitions WHERE type='delivery-email-sender' AND is_active
     AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
  IF position('200 pounds' IN body) <> 0 THEN
    RAISE EXCEPTION '726 VERIFY FAILED: "200 pounds" still present';
  END IF;
  IF position('a one-off 59.99 pounds' IN body) = 0 THEN
    RAISE EXCEPTION '726 VERIFY FAILED: the new price is not present';
  END IF;
  IF position('10 pounds a month' IN body) = 0 THEN
    RAISE EXCEPTION '726 VERIFY FAILED: the RENTAL price was disturbed — it must stay 10 pounds a month';
  END IF;
  RAISE NOTICE '726 OK: buy-out now 59.99, rental untouched at 10 pounds a month';
END $$;
COMMIT;
