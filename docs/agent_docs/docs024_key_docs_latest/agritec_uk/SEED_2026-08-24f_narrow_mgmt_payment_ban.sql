-- ============================================================================
-- agritec.uk — the SECOND over-broad ban: it blocked the honest past tense
-- Written 2026-08-24. Applied out of band (psql -f), per-site setup.
--
-- WHAT HAPPENED. After narrowing the ticker ban (24e), the SFI explainer was
-- still blocked, on a different pattern of mine. The blocker detail:
--
--   value: "£20 per hectare, for up to 50 hectares"
--   location: "...Under SFI 2023, an annual management payment of £20 per
--             hectare, for up to 50 hectares and capped at £1,000, WAS PAID on
--             top of action payments..."
--
-- That sentence is exactly what this site should say. It is past tense, it names
-- the scheme year, and it is the honest use of registered fact CIT-f88b5cd. My
-- own migration (SEED_2026-08-22_sfi26_bans) said so in as many words: "the
-- HISTORICAL fact is registered and remains sayable in the past tense with its
-- scope attached; this bans the bare present-tense form."
--
-- The pattern did not implement that sentence. It was
--     £ ?20 ?(per|/) ?(ha|hectare)[^.]{0,60}(first )?50 ?(ha|hectares)
-- which tests SHAPE (rate near cap) and nothing about tense at all. My keep-arm
-- test case happened not to have "50 hectares" within 60 characters, so it
-- passed; a real explanation naturally states the rate and the cap together, and
-- fails. **The test agreed with the intent and the pattern never did** — the
-- test case was the weak link, not the reasoning.
--
-- WHY NOT SIMPLY DROP IT. Measured before deciding: of the five
-- management-payment patterns, p4 is the ONLY one that catches the retired
-- site's actual sentence, "You receive £20 per hectare for the first 50
-- hectares of land entered into your agreement". Dropping it would unban the
-- thing this whole exercise exists to stop. So it is narrowed, not removed.
--
-- THE REPLACEMENT anchors on PRESENT-TENSE ENTITLEMENT rather than on shape: a
-- subject (you / farmers / holdings / applicants), a present-tense receipt verb
-- with word boundaries, then the rate. Tested on eight strings, both arms:
--   BANS  "You receive £20 per hectare for the first 50 hectares"
--         "Farmers receive £20 per hectare for the first 50 hectares"
--         "you get £20 per hectare on the first 50 hectares"
--   KEEPS "Under SFI 2023, an annual management payment of £20 per hectare, for
--          up to 50 hectares and capped at £1,000, was paid..."   <- the blocked one
--         "Under the SFI 2023 offer, ... was available for up to 50 hectares"
--         "Farmers RECEIVED £20 per hectare under the 2023 offer"  <- \b stops
--                                                                     the past tense
--         "The SFI management payment has been removed for SFI26 agreements"
--         "You will not be paid an SFI management payment..."
-- Compiles under Go RE2.
--
-- THE PATTERN OF MY OWN ERRORS, worth naming: both over-broad bans caught a
-- TRUE, WELL-SOURCED sentence, and in both cases the pages simply refused to
-- build. That is the good failure mode — a ban that suppresses a sentence
-- quietly would never have surfaced at all.
-- ============================================================================

\set ON_ERROR_STOP on

BEGIN;

CREATE TEMP TABLE _cur ON COMMIT DROP AS
SELECT ss.id, ss.site_id, ss.data
FROM site_specs ss JOIN sites s ON s.id = ss.site_id
WHERE s.domain = 'agritec.uk' AND ss.aspect = 'evidence_base' AND ss.is_current;

DO $guard$
DECLARE n int; hit int;
BEGIN
  SELECT count(*) INTO n FROM _cur;
  IF n <> 1 THEN RAISE EXCEPTION 'expected 1 current evidence_base row, found %', n; END IF;
  SELECT count(*) INTO hit FROM _cur, LATERAL jsonb_array_elements(data->'banned_claims') b
   WHERE b->>'pattern' LIKE '£ ?20 ?(per|/) ?(ha|hectare)%';
  IF hit <> 1 THEN RAISE EXCEPTION 'the shape-based £20 pattern is not present exactly once (found %)', hit; END IF;
END
$guard$;

UPDATE site_specs SET is_current = false, superseded_at = now()
WHERE id IN (SELECT id FROM _cur);

INSERT INTO site_specs (site_id, aspect, data, source, notes, is_current, pinned, created_by)
SELECT
  _cur.site_id, 'evidence_base',
  jsonb_set(_cur.data, '{banned_claims}',
    (SELECT jsonb_agg(
       CASE WHEN b->>'pattern' LIKE '£ ?20 ?(per|/) ?(ha|hectare)%'
            THEN jsonb_build_object(
                   'pattern','\b(you|farmers?|holdings?|applicants?)\b[^.]{0,25}\b(receive|receives|get|gets|are paid|is paid)\b[^.]{0,40}£ ?20 ?(per|/) ?(ha|hectare)',
                   'reason','Removed SFI management payment, stated as a PRESENT entitlement. NARROWED 2026-08-24: the previous pattern tested shape (rate near cap) and not tense, so it blocked the honest past-tense sentence "Under SFI 2023, an annual management payment of £20 per hectare, for up to 50 hectares and capped at £1,000, was paid...", which is the correct use of registered fact CIT-f88b5cd and stopped two pages building. Not dropped, because it is the ONLY one of the five management-payment patterns that catches the retired site''s own "You receive £20 per hectare for the first 50 hectares". Tested both arms on 8 strings including "Farmers RECEIVED" (past tense, kept via word boundaries).')
            ELSE b END ORDER BY ord)
     FROM jsonb_array_elements(_cur.data->'banned_claims') WITH ORDINALITY t(b,ord))),
  'manual',
  'Narrowed the £20/50ha management-payment ban from a shape test to a present-tense entitlement test, after it blocked the SFI explainer for stating the payment correctly in the past tense.',
  true, true, 'agritec-workstream-2026-08-24'
FROM _cur;

COMMIT;
