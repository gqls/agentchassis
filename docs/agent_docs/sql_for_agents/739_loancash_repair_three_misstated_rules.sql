-- 739_loancash_repair_three_misstated_rules.sql
--
-- Files FOUR `content_rewrite` work items repairing THREE wrong regulatory claims on
-- loancash.co.uk (the £15 default cap error appears on two pages).
--
-- WHY, AND THE AUTHORITY. Migration 738 gave loancash an evidence register and recorded
-- three wrong live claims as `corrects_site_citation`, deliberately WITHOUT touching the
-- served copy: rewriting published prose on an automated finding is authority the owner
-- withheld (the 695 precedent, `bugs_open/320` §15). **The owner lifted that hold on
-- 2026-09-03 — "fix the loancash wrong sentences" — recorded as RFC_060 §3g D2.** The hold
-- is lifted FOR THESE THREE FINDINGS ON THIS SITE ONLY; it is not a general licence.
--
-- WHY WORK ITEMS AND NOT AN EDIT. Two owner rulings: every site goes through the framework
-- (2026-08-04), and the framework writes the content, not the session (2026-08-06). So the
-- repair is dispatched to `page-build-handler` with the exact wrong wording, the exact
-- required correction, and an acceptance test readable at the served bytes.
--
-- WHAT IS WRONG, all three read at the SERVED page and checked against the primary source
-- with every quote verified through the production matcher (cmd/fcaquotecheck) on 2026-09-03:
--
--   1. THE £15 DEFAULT CAP IS CUMULATIVE, NOT PER MISSED PAYMENT — and the site therefore
--      UNDERSTATES the protection it exists to explain. CONC 5A.2.14R(1) caps charges
--      connected with a breach at £15 "whether in relation to one breach or cumulatively in
--      relation to multiple breaches of the agreement". `guides/the-payday-loan-price-cap.html`
--      says the fee "can only be charged once per missed payment"; `guides/jargon-buster.html`
--      says "capped at £15, one-off, regardless of how many days late you are". Both are right
--      about what they DENY (no per-day accrual) and wrong about the unit they AFFIRM. A reader
--      who missed two payments would accept a second £15 as lawful. It is not.
--      → items 1 and 2 (register fact FCA-CONC-5A-2-14).
--
--   2. THE CPA LIMIT IS TWO REFUSED ATTEMPTS, AND THERE IS NO £1 THRESHOLD.
--      `guides/loan-sharks-and-illegal-lending.html` says a regulated lender "cannot take more
--      than one payment attempt of over £1" without fresh permission. CONC 7.6.12R prohibits a
--      further request once two previous requests for the same sum have been REFUSED; the string
--      "£1" appears NOWHERE in CONC 7.6. The site's OWN
--      `guides/stopping-payments-the-cpa-rules.html` states the rule correctly, so this is an
--      internal contradiction on one page — that page is the reference to be made consistent
--      WITH, and must not be changed.
--      → item 3 (register facts FCA-CONC-7-6-12, FCA-CONC-7-6-14).
--
--   3. AFFORDABILITY IS CONC 5.2A, NOT "CONC 5A".
--      `guides/check-your-lender-is-authorised.html` attributes the affordability checks to
--      CONC 5A — the price-cap chapter, which contains no affordability rule. The duty is
--      CONC 5.2A.4R/5.2A.5R, which this site cites correctly on three other pages.
--      → item 4 (register fact FCA-CONC-5-2A-4).
--
-- ⚠ A `page_rerender` WAS QUEUED ON ALL FOUR PAGES AT 13:18 UTC TODAY, and it is NOT a repair.
-- A rerender regenerates from `content_data`, so it re-ships this wrong wording byte for byte
-- and completes successfully. Do not read a completed rerender as a fix, and do not cancel
-- these items because a rerender ran.
--
-- ⚠ AND A COMPLETED `content_rewrite` IS NOT A REPAIRED ARTEFACT (016b §9, and
-- complete_work_item_acceptance_predicate.go's own note that NO verifier is registered for
-- `content_rewrite`, so its acceptance gate is inert). Every item carries an
-- `acceptance_test` naming the served URL and the exact string that must or must not appear.
-- VERIFY AT THE SERVED BYTES.
--
-- Dedup: `idx_swi_dedup` is UNIQUE on (site_id, item_key) for non-terminal statuses; all four
-- keys are distinct and none exists on this site today (checked 2026-09-03). No `content_rewrite`
-- in a dispatchable state existed on any of the four pages.
--
-- Lane: docs/agent_docs/docs024_key_docs_latest/loancash_couk_fca_validation/
-- Register: docs/agent_docs/sql_for_agents/738_loancash_evidence_base_price_cap_complaints_and_breathing_space.sql
-- Rollback: 739_..._ROLLBACK.sql

BEGIN;

-- GUARD 1: the site_id must resolve to the domain (RUNBOOK_lendzy §8e — a mistyped uuid
-- populates another site and every count below still passes, being scoped to the same wrong id).
DO $$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM sites WHERE id = 'ee4a8199-4f5b-4e2e-88ce-01e600721b74' AND domain = 'loancash.co.uk') THEN
    RAISE EXCEPTION '739 ABORT: site_id does not resolve to loancash.co.uk';
  END IF;
END $$;

-- GUARD 2: the register these items cite must exist, or the suggestions point at nothing.
DO $$
DECLARE n int;
BEGIN
  SELECT count(*) INTO n FROM site_specs
   WHERE site_id = 'ee4a8199-4f5b-4e2e-88ce-01e600721b74' AND aspect = 'evidence_base' AND is_current;
  IF n <> 1 THEN
    RAISE EXCEPTION '739 ABORT: expected exactly 1 current evidence_base for loancash (migration 738), found %', n;
  END IF;
END $$;

-- GUARD 3: refuse to double-file. Any of these four keys already open means another pass ran.
DO $$
DECLARE n int;
BEGIN
  SELECT count(*) INTO n FROM site_work_items
   WHERE site_id = 'ee4a8199-4f5b-4e2e-88ce-01e600721b74'
     AND item_key IN ('loancash_claim_repair_price_cap_default_fee_cumulative',
                      'loancash_claim_repair_jargon_buster_default_fee_unit',
                      'loancash_claim_repair_loan_sharks_cpa_two_attempts',
                      'loancash_claim_repair_affordability_rule_reference')
     AND status NOT IN ('complete','verified','rejected','wont_fix','failed','unresolved','cancelled');
  IF n <> 0 THEN
    RAISE EXCEPTION '739 ABORT: % of the four repair items already open - read them before filing again', n;
  END IF;
END $$;

INSERT INTO site_work_items
  (site_id, source, pipeline, item_type, severity, summary, spec, page_id, affected_url,
   priority, handler_agent, status, created_by, item_key, approval_mode)
VALUES
  ('ee4a8199-4f5b-4e2e-88ce-01e600721b74', 'loancash-claims-primary-source-verification', 'build', 'content_rewrite',
   'high', 'Correct the default-fee passage: the £15 cap is CUMULATIVE across the whole agreement, not per missed payment (CONC 5A.2.14R(1)) — the page currently understates the protection',
   '{"origin":"primary_source_verification","category":"content","page_name":"guide-the-payday-loan-price-cap","current_value":"Default fees are capped separately Miss a repayment and the lender can add a default fee, but that fee is capped at £15 and can only be charged once per missed payment, not once per day it stays missed. […] A lender adding repeated £15 charges for the same missed payment is charging you illegally.","suggestion":"Rewrite the \"Default fees are capped separately\" passage so it states the cap as the rule states it. CONC 5A.2.14R(1) caps charges connected with a borrower breach at £15 \"whether in relation to one breach or cumulatively in relation to multiple breaches of the agreement\" — so £15 is the maximum in DEFAULT FEES FOR THE WHOLE AGREEMENT, however many payments are missed, not £15 per missed payment. KEEP the two things the page already gets right: that the fee does not accrue per day, and that interest on the unpaid balance stays inside the same 0.8%/day and 100% total limits (CONC 5A.2.14R(2)-(3)). CHANGE the unit: remove \"can only be charged once per missed payment\", and widen the closing warning — repeated £15 charges are unlawful whether they are for the SAME missed payment or for DIFFERENT ones. The registered fact FCA-CONC-5A-2-14 (migration 738) carries the verbatim rule text and the writer_line to use. Do not add a rule number to the body if the page''s existing style does not carry them.","acceptance_test":"Load https://loancash.co.uk/guides/the-payday-loan-price-cap.html and confirm the served bytes state that £15 is the maximum total in default fees across the agreement however many payments are missed, AND contain no sentence stating or implying that a fresh £15 may be charged for each missed payment. The words \"once per missed payment\" must not appear.","affected_component":"Guide body — the ''Default fees are capped separately'' section","description":"Correct the default-fee passage: the £15 cap is CUMULATIVE across the whole agreement, not per missed payment (CONC 5A.2.14R(1)) — the page currently understates the protection","max_fix_attempts":2,"evidence_register_migration":"738","authority":"OWNER RULING 2026-09-03 (RFC_060 §3g D2): ''fix the loancash wrong sentences''. This LIFTS the standing hold from the 695 precedent / bugs_open/320 §15 — under which corrects_site_citation recorded a finding and the served copy was deliberately left alone pending the owner — FOR THESE THREE FINDINGS ON THIS SITE ONLY. It is not a general licence to rewrite published prose on an automated finding.","verified_how":"The rule text was read at the primary source and the quote verified through the PRODUCTION matcher (cmd/fcaquotecheck -> datahelpers.VisibleTextFromHTML / QuoteFoundInText) on 2026-09-03, with a deliberately-absent control returning false in the same run. The wrong wording was read from the SERVED page, not from content_data.","do_not_touch":"/guides/stopping-payments-the-cpa-rules.html is CORRECT on the CPA two-attempt rule and must not be changed; it is the reference the loan-sharks page should be made consistent WITH.","beware":"A page_rerender was queued on this page at 2026-09-03 13:18 UTC. A rerender regenerates from content_data and therefore re-ships this wrong wording unchanged — it is not a repair and must not be read as one. Verify at the SERVED bytes, never at the work item status: a ''complete'' content_rewrite is not a repaired artefact."}'::jsonb,
   '8b768e35-f125-4318-8567-5956db43bee9', '/guides/the-payday-loan-price-cap.html', 20, 'page-build-handler', 'triaged',
   'loancash_couk_fca_validation lane (migration 739)', 'loancash_claim_repair_price_cap_default_fee_cumulative', 'auto'),
  ('ee4a8199-4f5b-4e2e-88ce-01e600721b74', 'loancash-claims-primary-source-verification', 'build', 'content_rewrite',
   'high', 'Correct the default-fee glossary entry: "regardless of how many days late" states the wrong unit — the £15 cap is cumulative across the agreement (CONC 5A.2.14R(1))',
   '{"origin":"primary_source_verification","category":"content","page_name":"guide-jargon-buster","current_value":"It is capped at £15, one-off, regardless of how many days late you are.","suggestion":"The glossary entry for the default fee is right that the charge is one-off rather than daily, but it frames the unit as per-instance. CONC 5A.2.14R(1) caps default charges at £15 \"whether in relation to one breach or cumulatively in relation to multiple breaches of the agreement\". Change \"regardless of how many days late you are\" to make the cumulative unit explicit — the £15 is the most that can be charged in default fees across the whole agreement, however many payments are missed and however late they are. Registered fact FCA-CONC-5A-2-14 (migration 738) carries the verbatim rule text.","acceptance_test":"Load https://loancash.co.uk/guides/jargon-buster.html and confirm the default-fee entry states the £15 is the total across the agreement however many payments are missed. The phrase \"regardless of how many days late\" must not be the only qualifier.","affected_component":"Jargon-buster glossary — the default fee entry","description":"Correct the default-fee glossary entry: \"regardless of how many days late\" states the wrong unit — the £15 cap is cumulative across the agreement (CONC 5A.2.14R(1))","max_fix_attempts":2,"evidence_register_migration":"738","authority":"OWNER RULING 2026-09-03 (RFC_060 §3g D2): ''fix the loancash wrong sentences''. This LIFTS the standing hold from the 695 precedent / bugs_open/320 §15 — under which corrects_site_citation recorded a finding and the served copy was deliberately left alone pending the owner — FOR THESE THREE FINDINGS ON THIS SITE ONLY. It is not a general licence to rewrite published prose on an automated finding.","verified_how":"The rule text was read at the primary source and the quote verified through the PRODUCTION matcher (cmd/fcaquotecheck -> datahelpers.VisibleTextFromHTML / QuoteFoundInText) on 2026-09-03, with a deliberately-absent control returning false in the same run. The wrong wording was read from the SERVED page, not from content_data.","do_not_touch":"/guides/stopping-payments-the-cpa-rules.html is CORRECT on the CPA two-attempt rule and must not be changed; it is the reference the loan-sharks page should be made consistent WITH.","beware":"A page_rerender was queued on this page at 2026-09-03 13:18 UTC. A rerender regenerates from content_data and therefore re-ships this wrong wording unchanged — it is not a repair and must not be read as one. Verify at the SERVED bytes, never at the work item status: a ''complete'' content_rewrite is not a repaired artefact."}'::jsonb,
   '2f0107ea-0e5d-48f1-8606-6390a346d013', '/guides/jargon-buster.html', 20, 'page-build-handler', 'triaged',
   'loancash_couk_fca_validation lane (migration 739)', 'loancash_claim_repair_jargon_buster_default_fee_unit', 'auto'),
  ('ee4a8199-4f5b-4e2e-88ce-01e600721b74', 'loancash-claims-primary-source-verification', 'build', 'content_rewrite',
   'high', 'Correct the CPA sentence: the limit is TWO REFUSED attempts (CONC 7.6.12R) and there is no £1 threshold anywhere in CONC 7.6 — this page contradicts the site''s own CPA guide',
   '{"origin":"primary_source_verification","category":"content","page_name":"guide-loan-sharks-and-illegal-lending","current_value":"Compare that with a regulated lender, which must run an affordability check before lending, cannot take more than one payment attempt of over £1 through a continuous payment authority (CPA) without your fresh permission, and is bound by the price cap on what it can charge.","suggestion":"Two errors in one clause. (1) The limit is not \"more than one payment attempt\": CONC 7.6.12R prohibits a FURTHER request once a firm has made two previous requests for the same sum AND both were REFUSED — i.e. two failed attempts, then it must stop and contact the customer. (2) There is NO £1 threshold in CONC 7.6; the string does not appear anywhere in the section. Delete the \"of over £1\" qualifier entirely rather than replacing it with another figure. Rewrite the clause to match the site''s OWN correct treatment at /guides/stopping-payments-the-cpa-rules.html (\"A lender can attempt to collect a payment using your CPA twice. If both attempts fail … the lender must not make a third attempt for that payment without contacting you first\"), so the two pages agree. Registered facts FCA-CONC-7-6-12 and FCA-CONC-7-6-14 (migration 738) carry the verbatim rule text.","acceptance_test":"Load https://loancash.co.uk/guides/loan-sharks-and-illegal-lending.html and confirm the served bytes: (a) state the CPA limit as two attempts, (b) contain no \"£1\" threshold claim about continuous payment authorities, and (c) do not contradict /guides/stopping-payments-the-cpa-rules.html, which is correct and must not be changed.","affected_component":"Guide body — the regulated-lender comparison paragraph","description":"Correct the CPA sentence: the limit is TWO REFUSED attempts (CONC 7.6.12R) and there is no £1 threshold anywhere in CONC 7.6 — this page contradicts the site''s own CPA guide","max_fix_attempts":2,"evidence_register_migration":"738","authority":"OWNER RULING 2026-09-03 (RFC_060 §3g D2): ''fix the loancash wrong sentences''. This LIFTS the standing hold from the 695 precedent / bugs_open/320 §15 — under which corrects_site_citation recorded a finding and the served copy was deliberately left alone pending the owner — FOR THESE THREE FINDINGS ON THIS SITE ONLY. It is not a general licence to rewrite published prose on an automated finding.","verified_how":"The rule text was read at the primary source and the quote verified through the PRODUCTION matcher (cmd/fcaquotecheck -> datahelpers.VisibleTextFromHTML / QuoteFoundInText) on 2026-09-03, with a deliberately-absent control returning false in the same run. The wrong wording was read from the SERVED page, not from content_data.","do_not_touch":"/guides/stopping-payments-the-cpa-rules.html is CORRECT on the CPA two-attempt rule and must not be changed; it is the reference the loan-sharks page should be made consistent WITH.","beware":"A page_rerender was queued on this page at 2026-09-03 13:18 UTC. A rerender regenerates from content_data and therefore re-ships this wrong wording unchanged — it is not a repair and must not be read as one. Verify at the SERVED bytes, never at the work item status: a ''complete'' content_rewrite is not a repaired artefact."}'::jsonb,
   '0988c0b6-cccb-4254-b9b4-187d7b9f5cfe', '/guides/loan-sharks-and-illegal-lending.html', 20, 'page-build-handler', 'triaged',
   'loancash_couk_fca_validation lane (migration 739)', 'loancash_claim_repair_loan_sharks_cpa_two_attempts', 'auto'),
  ('ee4a8199-4f5b-4e2e-88ce-01e600721b74', 'loancash-claims-primary-source-verification', 'build', 'content_rewrite',
   'medium', 'Correct the rule reference: affordability is CONC 5.2A, not "CONC 5A" (which is the price-cap chapter and contains no affordability rule)',
   '{"origin":"primary_source_verification","category":"content","page_name":"guide-check-your-lender-is-authorised","current_value":"It means the lender must follow the FCA''s consumer credit rules, including the affordability checks set out under CONC 5A and the price cap on high-cost short-term credit: 0.8% per day in interest and fees, a £15 cap on default fees, and a 100% cap on total cost.","suggestion":"The sentence attributes the affordability checks to \"CONC 5A\". CONC 5A is titled \"Prohibition from entering into agreements for high-cost short-term credit\" and contains only the cost caps — it has no affordability rule. The creditworthiness duty is CONC 5.2A (CONC 5.2A.4R / 5.2A.5R), which THIS SITE cites correctly on three other pages. Either correct the reference to CONC 5.2A for affordability while leaving CONC 5A attached to the price cap, or drop the rule numbers from this sentence altogether — both are acceptable; what must not survive is affordability attributed to CONC 5A. Registered facts FCA-CONC-5-2A-4 and FCA-CONC-5A-2-2/2-3/2-14 (migration 738).","acceptance_test":"Load https://loancash.co.uk/guides/check-your-lender-is-authorised.html and confirm the served bytes do not attribute affordability checks to CONC 5A. Either affordability is attributed to CONC 5.2A, or no rule number is attached to it.","affected_component":"Guide body — the ''what authorisation means'' paragraph","description":"Correct the rule reference: affordability is CONC 5.2A, not \"CONC 5A\" (which is the price-cap chapter and contains no affordability rule)","max_fix_attempts":2,"evidence_register_migration":"738","authority":"OWNER RULING 2026-09-03 (RFC_060 §3g D2): ''fix the loancash wrong sentences''. This LIFTS the standing hold from the 695 precedent / bugs_open/320 §15 — under which corrects_site_citation recorded a finding and the served copy was deliberately left alone pending the owner — FOR THESE THREE FINDINGS ON THIS SITE ONLY. It is not a general licence to rewrite published prose on an automated finding.","verified_how":"The rule text was read at the primary source and the quote verified through the PRODUCTION matcher (cmd/fcaquotecheck -> datahelpers.VisibleTextFromHTML / QuoteFoundInText) on 2026-09-03, with a deliberately-absent control returning false in the same run. The wrong wording was read from the SERVED page, not from content_data.","do_not_touch":"/guides/stopping-payments-the-cpa-rules.html is CORRECT on the CPA two-attempt rule and must not be changed; it is the reference the loan-sharks page should be made consistent WITH.","beware":"A page_rerender was queued on this page at 2026-09-03 13:18 UTC. A rerender regenerates from content_data and therefore re-ships this wrong wording unchanged — it is not a repair and must not be read as one. Verify at the SERVED bytes, never at the work item status: a ''complete'' content_rewrite is not a repaired artefact."}'::jsonb,
   'd4f791b6-4cb6-4095-aacf-3731e00a4039', '/guides/check-your-lender-is-authorised.html', 40, 'page-build-handler', 'triaged',
   'loancash_couk_fca_validation lane (migration 739)', 'loancash_claim_repair_affordability_rule_reference', 'auto');

-- VERIFY as DO/RAISE. A verify block of bare SELECTs cannot stop the COMMIT.
DO $$
DECLARE n int; npage int; nacc int;
BEGIN
  SELECT count(*) INTO n FROM site_work_items
   WHERE site_id = 'ee4a8199-4f5b-4e2e-88ce-01e600721b74' AND created_by = 'loancash_couk_fca_validation lane (migration 739)';
  IF n <> 4 THEN
    RAISE EXCEPTION '739 VERIFY: expected 4 repair items, found %', n;
  END IF;

  SELECT count(*) INTO npage FROM site_work_items w JOIN pages p ON p.id = w.page_id
   WHERE w.site_id = 'ee4a8199-4f5b-4e2e-88ce-01e600721b74' AND w.created_by = 'loancash_couk_fca_validation lane (migration 739)'
     AND p.site_id = 'ee4a8199-4f5b-4e2e-88ce-01e600721b74';
  IF npage <> 4 THEN
    RAISE EXCEPTION '739 VERIFY: expected all 4 items to name a page belonging to this site, found %', npage;
  END IF;

  SELECT count(*) INTO nacc FROM site_work_items
   WHERE site_id = 'ee4a8199-4f5b-4e2e-88ce-01e600721b74' AND created_by = 'loancash_couk_fca_validation lane (migration 739)'
     AND length(spec->>'acceptance_test') > 60
     AND length(spec->>'current_value')  > 40
     AND spec->>'suggestion' IS NOT NULL
     AND status = 'triaged' AND handler_agent = 'page-build-handler';
  IF nacc <> 4 THEN
    RAISE EXCEPTION '739 VERIFY: expected 4 dispatchable items each carrying current_value, suggestion and a substantive acceptance_test, found %', nacc;
  END IF;

  RAISE NOTICE '739 OK: 4 content_rewrite items filed for loancash - 3 misstated rules across 4 pages, dispatchable to page-build-handler';
END $$;

COMMIT;
