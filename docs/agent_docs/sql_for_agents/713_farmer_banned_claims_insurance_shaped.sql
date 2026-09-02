-- 713_farmer_banned_claims_insurance_shaped.sql
--
-- Fills farmerinsurance.uk's empty banned_claims (698 shipped the policy facts
-- with banned_claims untouched at []) — the last register gap besides loancash
-- [the 414 lane's ~17:50 census: loancalculator 12/8, lendzy 8/5, loanzy 3/5,
-- farmerinsurance 7/0]. A facts-only register reads as "done" in every count
-- census while enforcing nothing at the build gate.
--
-- FIVE INSURANCE-SHAPED patterns — not the credit set transplanted: farmer is
-- an FSMA-flagged information site (not FCA-authorised, arranges nothing,
-- explicitly does not compare or rank insurers), so its falsehood surface is
-- payout promises, universal-acceptance claims, broker misrepresentation,
-- price-superiority claims, and unregulated-end markers.
--
-- CALIBRATED BEFORE WRITING, per the day's discipline (lendzy RUNBOOK §8b/8e +
-- the blind-zero landmine's three controls): all five patterns run over the
-- FULL served corpus with count reconciliation (18 active rows -> 18 fetched ->
-- 17 scanned + 1 accounted: /claims.html serves 404 — a bugs_open/437
-- mechanism-flow victim, needs_rebuild since 2026-08-27, so it can carry no
-- phrase); ZERO hits; planted-text positive control 5/5.
--
-- DELIBERATELY ABSENT, with reasons:
--   * a literal-%/premium-rate pattern — the banned-claims layer has NO
--     regulatory-citation exemption (the 414 lane verified: fad209b92 exempts
--     cited figures from the NUMBER scan only), and farmer quotes the £5m ELCI
--     statutory minimum beside its rule (fact LAW-ELCI-1998-R3);
--   * a first-person "we are FCA-authorised" pattern — already refused
--     FLEET-WIDE by CGV-033 (live v1.0.1317 on build+save paths); duplicating
--     it here would shadow the shared mechanism with a weaker local copy.
--
-- Lane: docs/agent_docs/docs024_key_docs_latest/loanzy_uk_example_site/
-- Rollback: 713_..._ROLLBACK.sql
\set ON_ERROR_STOP on
BEGIN;

DO $$
DECLARE n int; nb int;
BEGIN
  SELECT count(*) INTO n FROM site_specs
   WHERE site_id='99cae989-2413-430d-b026-59dfeeb638c0' AND aspect='evidence_base' AND is_current;
  IF n <> 1 THEN RAISE EXCEPTION '713 ABORT: expected exactly 1 current farmer register, found %', n; END IF;
  SELECT jsonb_array_length(COALESCE(data->'banned_claims','[]'::jsonb)) INTO nb FROM site_specs
   WHERE site_id='99cae989-2413-430d-b026-59dfeeb638c0' AND aspect='evidence_base' AND is_current;
  IF nb <> 0 THEN RAISE EXCEPTION '713 ABORT: banned_claims already has % entries - read before writing', nb; END IF;
END $$;

WITH cur AS (
  UPDATE site_specs SET is_current=false, superseded_at=now()
   WHERE site_id='99cae989-2413-430d-b026-59dfeeb638c0' AND aspect='evidence_base' AND is_current
   RETURNING data
)
INSERT INTO site_specs (site_id, aspect, data, source, source_agent, created_by, is_current, pinned, notes)
SELECT
  '99cae989-2413-430d-b026-59dfeeb638c0',
  'evidence_base',
  jsonb_set(cur.data, '{banned_claims}', '[
    {"pattern": "\\bguaranteed (payouts?|acceptance|approval|claims?|settlements?)\\b", "reason": "No guaranteed-outcome language, ever: no insurer''s payout is guaranteed, the claim is unprovable, and on an FSMA-flagged information site it is a financial-promotion exposure (insurance-shaped analogue of the adversecreditmortgage M5 rule)."},
    {"pattern": "\\b(all|every) claims? (is|are|will be) (paid|accepted|approved|settled)\\b", "reason": "The same promise wearing claims-handling grammar; ICOBS 8.1.1 requires fair handling, it does not make every claim payable, and the site''s own copy explains exactly when insurers push back."},
    {"pattern": "\\b(we|our team) (can|will) (get|secure|find|arrange) you (a|the) (policy|cover|quote|payout|settlement)\\b", "reason": "farmerinsurance is an information site: not FCA-authorised, arranges nothing, and says so on every page. This phrasing would misrepresent what the site is (sibling pattern, broker-shaped)."},
    {"pattern": "\\b(cheapest|lowest) (premiums?|prices?|quotes?|cover)\\b", "reason": "A price-superiority claim the site has forsworn in its own copy (''We do not compare named insurers against each other, rank them, or suggest that one is better value'') — and could not evidence if it wanted to."},
    {"pattern": "\\bno[- ]questions[- ]asked (cover|insurance|payouts?|claims?)\\b", "reason": "The unregulated-end marker in insurance clothing — the analogue of no-credit-check loans, and the very thing the site warns readers about in its unregulated-introducers passage."}
  ]'::jsonb),
  'manual', NULL, 'loanzy_uk_example_site lane (migration 713)', true, true,
  'banned_claims filled: five insurance-shaped patterns (payout guarantees, universal acceptance, broker misrepresentation, price superiority, no-questions-asked), calibrated at 0 hits over all 17 serving pages with a 5/5 planted control and count reconciliation (18th page is a 437 victim serving 404). Literal-rate pattern deliberately absent (banned-claims layer has no citation exemption; the register quotes the ELCI 5m statutory minimum); first-person FCA-authorisation claims deliberately absent (CGV-033 refuses them fleet-wide). Facts carried forward unchanged.'
FROM cur;

DO $$
DECLARE nfacts int; nb int; nbad int; ncur int; nesc int;
BEGIN
  SELECT count(*) INTO ncur FROM site_specs
   WHERE site_id='99cae989-2413-430d-b026-59dfeeb638c0' AND aspect='evidence_base' AND is_current;
  IF ncur <> 1 THEN RAISE EXCEPTION '713 VERIFY: expected exactly 1 current row, found %', ncur; END IF;
  SELECT jsonb_array_length(data->'facts'), jsonb_array_length(data->'banned_claims') INTO nfacts, nb
   FROM site_specs WHERE site_id='99cae989-2413-430d-b026-59dfeeb638c0' AND aspect='evidence_base' AND is_current;
  IF nfacts <> 7 THEN RAISE EXCEPTION '713 VERIFY: facts were LOST - expected 7, found %', nfacts; END IF;
  IF nb <> 5 THEN RAISE EXCEPTION '713 VERIFY: expected 5 banned_claims, found %', nb; END IF;
  SELECT count(*) INTO nbad FROM site_specs s, jsonb_array_elements(s.data->'banned_claims') b
   WHERE s.site_id='99cae989-2413-430d-b026-59dfeeb638c0' AND s.aspect='evidence_base' AND s.is_current
     AND (length(COALESCE(b->>'pattern','')) < 10 OR length(COALESCE(b->>'reason','')) < 20);
  IF nbad <> 0 THEN RAISE EXCEPTION '713 VERIFY: % entries with missing/thin pattern or reason', nbad; END IF;
  -- Double-escape landmine (695''s arms, both directions): a double-escaped
  -- pattern compiles cleanly in the Go consumer and matches NOTHING.
  SELECT count(*) INTO nesc FROM site_specs s3, jsonb_array_elements(s3.data->'banned_claims') b
   WHERE s3.site_id='99cae989-2413-430d-b026-59dfeeb638c0' AND s3.aspect='evidence_base' AND s3.is_current
     AND strpos(b->>'pattern', E'\\\\') > 0;
  IF nesc <> 0 THEN RAISE EXCEPTION '713 VERIFY: % stored pattern(s) carry a DOUBLE backslash - would match nothing', nesc; END IF;
  SELECT count(*) INTO nesc FROM site_specs s3, jsonb_array_elements(s3.data->'banned_claims') b
   WHERE s3.site_id='99cae989-2413-430d-b026-59dfeeb638c0' AND s3.aspect='evidence_base' AND s3.is_current
     AND strpos(b->>'pattern', E'\\b') > 0;
  IF nesc <> 5 THEN RAISE EXCEPTION '713 VERIFY: expected 5 patterns carrying a single-escaped word boundary, found %', nesc; END IF;
  RAISE NOTICE '713 OK: farmer banned_claims = 5 insurance-shaped (facts carried: 7)';
END $$;
COMMIT;
