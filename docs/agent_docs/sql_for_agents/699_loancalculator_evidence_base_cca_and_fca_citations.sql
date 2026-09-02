-- 699_loancalculator_evidence_base_cca_and_fca_citations.sql
--
-- Gives loancalculator.co.uk an evidence register: TWELVE facts, each citing the
-- statutory provision or FCA Handbook rule that states it, with a verbatim quote.
--
-- WHY. RFC_060 §1b: loancalculator.co.uk is one of the five register-less finance
-- sites — ZERO evidence_base rows [MEASURED 2026-09-02], so ScanUnregisteredNumbers
-- never arms and the daily citation check passes over an empty set. §4: populating
-- these registers "belongs to the site lanes, needs no RFC, and would deliver most
-- of the benefit ... If only one thing happens as a result of this RFC, it should
-- be that." Owner directed 2026-09-02 (relayed by the lendzy lane with the method);
-- pattern = migration 695 (lendzy), followed here structurally.
--
-- WHAT THE SITE ASSERTS, measured at the ARTEFACT (RUNBOOK_lendzy §8 step 1): all
-- 28 served pages were crawled 2026-09-02; 232 regulatory-shaped sentences
-- extracted. Unlike lendzy (payday-lending: CONC 5A caps), this site's external
-- assertions are overwhelmingly CONSUMER CREDIT ACT 1974 rights — settlement
-- figures, voluntary termination, the 14-day withdrawal, default notices, the
-- dealer-as-agent rule — plus FCA affordability/forbearance/complaints rules. So
-- this register cites legislation.gov.uk for the statutes and handbook.fca.org.uk
-- for the rules. The daily refresher re-fetches ANY citation URL, so both hosts
-- get the same daily quote re-check from the moment this row exists.
--
-- EVERY QUOTE WAS VERIFIED THROUGH THE PRODUCTION MATCHER (cmd/fcaquotecheck →
-- datahelpers.VisibleTextFromHTML / QuoteFoundInText) on 2026-09-02: 12/12 true,
-- with the deliberately-absent control false in the same run, and every URL
-- confirmed by <title> (both hosts 200 on wrong paths; handbook.fca.org.uk is the
-- documented LANDMINE, and legislation.gov.uk gets the same discipline).
--
-- TWO FACTS CARRY corrects_site_citation — the point of the exercise, as on lendzy:
--   * tools/settlement-calculator.html says the written settlement figure arrives
--     "usually within ten working days". The prescribed period is TWELVE working
--     days — SI 1983/1564 reg 4, quoted verbatim in fact SI-1983-1564-REG4.
--   * tools/overpayment-calculator.html attributes an ERC-free allowance of "up to
--     10% of your outstanding balance in any 12-month period" to the CCA 1974. No
--     such 10% rule exists in the Act: the statutory threshold is £8,000 of early
--     repayment in any 12-month period (s.95A(2)(a)), with compensation capped at
--     1% / 0.5% of the payment (s.95A(3)-(4)). A 10% fee-free allowance is a
--     product feature of some (mostly mortgage) agreements, not a CCA right.
-- THE SERVED COPY IS NOT TOUCHED — rewriting published prose on an automated
-- finding is authority the owner withheld (bugs_open/320 §15, the 695 precedent);
-- the copy repairs are his call, tracked in the lane.
--
-- A NEAR-MISS WORTH RECORDING (it is the method working on its own author): the
-- "12 working days" period was first traced to SI 1983/1569 (Prescribed Periods
-- for Giving Information) — whose Schedule covers ss.77-79/103/107-110 and NOT
-- s.97. The right instrument is SI 1983/1564 (Settlement Information), whose reg 4
-- names s.97(1) explicitly. Registering the first guess would have committed the
-- exact attribution error this register exists to catch.
--
-- ATTRIBUTION NUANCE, noted not corrected: several pages attribute affordability
-- checks to the FCA's "Consumer Duty". The substantive assessment rule is CONC
-- 5.2A.5 (registered here, as on lendzy); PRIN 2A exists and the framing is not
-- flatly wrong, so it gets no corrects_site_citation.
--
-- ARMING THE NUMERIC SCAN IS MEASURED-SAFE FOR THIS SITE: RFC_060 §1c armed the
-- scan locally against all five register-less sites' components (474, nothing
-- written) and its 5 findings included ZERO on loancalculator.co.uk; the
-- regulatory-citation exclusions (fad209b92) are live fleet-wide.
--
-- LIMIT ACCEPTED (695's §B5, structural): neither host has a rule-level URL, so
-- the daily check keeps QUOTES honest but not the `rule` field. Until RFC_060
-- §3d/Q6's rule-span checker ships, `rule` is a HUMAN-VERIFIED field: checked by
-- hand against each provision's own heading on 2026-09-02.
--
-- Lane: docs/agent_docs/docs024_key_docs_latest/loancalculator_couk/
-- Rollback: 699_..._ROLLBACK.sql

BEGIN;

-- GUARD. A current register must not already exist (idx_site_specs_current is
-- UNIQUE on (site_id, aspect) WHERE is_current). If another session has written
-- one since this file was authored, ABORT rather than supersede work unseen.
DO $$
DECLARE n int;
BEGIN
  SELECT count(*) INTO n FROM site_specs
   WHERE site_id = '0162cde4-633e-45e9-8ca6-87a6b2fe1d26' AND aspect = 'evidence_base';
  IF n <> 0 THEN
    RAISE EXCEPTION '699 ABORT: loancalculator already has % evidence_base row(s) - read them before writing', n;
  END IF;
END $$;

INSERT INTO site_specs (site_id, aspect, data, source, source_agent, created_by, is_current, pinned, notes)
VALUES (
  '0162cde4-633e-45e9-8ca6-87a6b2fe1d26',
  'evidence_base',
  '{"facts": [
    {"id": "CCA-1974-S97", "kind": "policy", "rule": "Consumer Credit Act 1974 s.97", "claim": "A lender under a regulated consumer credit agreement must, on the borrower''s request, give a statement of the amount needed to discharge the debt (the settlement figure), with prescribed particulars of how it is arrived at.", "writer_line": "the statutory right to a settlement figure on request (Consumer Credit Act 1974 s.97)", "source": {"citation": {"url": "https://www.legislation.gov.uk/ukpga/1974/39/section/97", "quote": "give the debtor a statement in the prescribed form indicating", "title": "Consumer Credit Act 1974 - Section 97: Duty to give information", "publisher": "legislation.gov.uk (The National Archives)", "accessed": "2026-09-02"}}, "verified_at": "2026-09-02", "corrects_site_citation": "tools/settlement-calculator.html says the written figure arrives ''usually within ten working days''. The prescribed period is TWELVE working days - SI 1983/1564 reg 4 (fact SI-1983-1564-REG4). Copy not touched; owner''s call."},
    {"id": "SI-1983-1564-REG4", "kind": "policy", "rule": "Consumer Credit (Settlement Information) Regulations 1983, reg 4", "claim": "The prescribed period within which a lender must give the s.97 settlement statement is 12 working days after receiving the request.", "writer_line": "the 12-working-day deadline for a settlement statement (SI 1983/1564 reg 4)", "source": {"citation": {"url": "https://www.legislation.gov.uk/uksi/1983/1564/made", "quote": "The period of 12 working days is hereby prescribed for the purposes of section 97(1) of the Act", "title": "The Consumer Credit (Settlement Information) Regulations 1983", "publisher": "legislation.gov.uk (The National Archives)", "accessed": "2026-09-02"}}, "verified_at": "2026-09-02"},
    {"id": "CCA-1974-S99", "kind": "policy", "rule": "Consumer Credit Act 1974 s.99", "claim": "At any time before the final payment falls due, the borrower under a regulated hire-purchase or conditional sale agreement may terminate the agreement by giving notice (voluntary termination).", "writer_line": "the voluntary termination right (Consumer Credit Act 1974 s.99)", "source": {"citation": {"url": "https://www.legislation.gov.uk/ukpga/1974/39/section/99", "quote": "the debtor shall be entitled to terminate the agreement by giving notice to any person entitled or authorised to receive the sums payable under the agreement", "title": "Consumer Credit Act 1974 - Section 99: Right to terminate hire-purchase etc. agreements", "publisher": "legislation.gov.uk (The National Archives)", "accessed": "2026-09-02"}}, "verified_at": "2026-09-02"},
    {"id": "CCA-1974-S100", "kind": "policy", "rule": "Consumer Credit Act 1974 s.100", "claim": "On voluntary termination the borrower''s liability is topped up to one-half of the total price, unless the agreement provides for less or a court orders a smaller sum.", "writer_line": "the half-of-total-price rule on voluntary termination (Consumer Credit Act 1974 s.100)", "source": {"citation": {"url": "https://www.legislation.gov.uk/ukpga/1974/39/section/100", "quote": "one-half of the total price exceeds the aggregate of the sums paid and the sums due in respect of the total price immediately before the termination", "title": "Consumer Credit Act 1974 - Section 100: Liability of debtor on termination of hire-purchase etc. agreement", "publisher": "legislation.gov.uk (The National Archives)", "accessed": "2026-09-02"}}, "verified_at": "2026-09-02"},
    {"id": "CCA-1974-S66A", "kind": "policy", "rule": "Consumer Credit Act 1974 s.66A", "claim": "A borrower under a regulated consumer credit agreement (other than an excluded agreement) may withdraw from it without giving any reason, by oral or written notice, within 14 days beginning the day after the relevant day.", "writer_line": "the 14-day right of withdrawal (Consumer Credit Act 1974 s.66A)", "source": {"citation": {"url": "https://www.legislation.gov.uk/ukpga/1974/39/section/66A", "quote": "give oral or written notice of the withdrawal to the creditor before the end of the period of 14 days", "title": "Consumer Credit Act 1974 - Section 66A: Withdrawal from consumer credit agreement", "publisher": "legislation.gov.uk (The National Archives)", "accessed": "2026-09-02"}}, "verified_at": "2026-09-02"},
    {"id": "CCA-1974-S87", "kind": "policy", "rule": "Consumer Credit Act 1974 s.87", "claim": "A default notice must be served before a lender can, by reason of the borrower''s breach, terminate the agreement, demand earlier payment, recover possession of goods or land, or enforce security.", "writer_line": "the default notice requirement (Consumer Credit Act 1974 s.87)", "source": {"citation": {"url": "https://www.legislation.gov.uk/ukpga/1974/39/section/87", "quote": "is necessary before the creditor or owner can become entitled, by reason of any breach by the debtor or hirer of a regulated agreement", "title": "Consumer Credit Act 1974 - Section 87: Need for default notice", "publisher": "legislation.gov.uk (The National Archives)", "accessed": "2026-09-02"}}, "verified_at": "2026-09-02"},
    {"id": "CCA-1974-S56", "kind": "policy", "rule": "Consumer Credit Act 1974 s.56", "claim": "Antecedent negotiations conducted by a credit-broker or supplier in the relevant transactions are deemed to be conducted as agent of the lender as well as in their own capacity, so the lender shares responsibility for what is said in the sale.", "writer_line": "the dealer-as-agent rule (Consumer Credit Act 1974 s.56)", "source": {"citation": {"url": "https://www.legislation.gov.uk/ukpga/1974/39/section/56", "quote": "shall be deemed to be conducted by the negotiator in the capacity of agent of the creditor as well as in his actual capacity", "title": "Consumer Credit Act 1974 - Section 56: Antecedent negotiations", "publisher": "legislation.gov.uk (The National Archives)", "accessed": "2026-09-02"}}, "verified_at": "2026-09-02"},
    {"id": "CCA-1974-S95A", "kind": "policy", "rule": "Consumer Credit Act 1974 s.95A", "claim": "On early repayment of part of a fixed-rate regulated agreement, the lender may claim compensation only where the payment exceeds 8,000 pounds - or payments total more than 8,000 pounds in any 12-month period - and the amount is capped at 1% of the payment (0.5% where a year or less remains), or the remaining interest if lower.", "writer_line": "the 8,000-pound/12-month threshold and 1%/0.5% cap on early-repayment compensation (Consumer Credit Act 1974 s.95A)", "source": {"citation": {"url": "https://www.legislation.gov.uk/ukpga/1974/39/section/95A", "quote": "where more than one such payment is made in any 12 month period, the total of those payments exceeds £8,000", "title": "Consumer Credit Act 1974 - Section 95A: Compensatory amount", "publisher": "legislation.gov.uk (The National Archives)", "accessed": "2026-09-02"}}, "verified_at": "2026-09-02", "corrects_site_citation": "tools/overpayment-calculator.html attributes an ERC-free allowance of ''up to 10% of your outstanding balance in any 12-month period'' to the Consumer Credit Act 1974. No 10% rule exists in the Act: the statutory threshold is 8,000 pounds of early repayment in any 12-month period (s.95A(2)(a)). A 10% fee-free allowance is a product feature of some (mostly mortgage) agreements, not a CCA right. Copy not touched; owner''s call."},
    {"id": "SI-2004-1483-REG5-6", "kind": "policy", "rule": "Consumer Credit (Early Settlement) Regulations 2004, regs 5-6", "claim": "For full early settlement, the interest rebate is calculated to a settlement date 28 days after the borrower''s notice, and where the agreement runs more than a year the lender may defer that date by one further month - the mechanism behind the ''one to two months'' interest'' an early settlement typically costs.", "writer_line": "the 28-day settlement date plus one-month deferment behind typical early-settlement charges (Consumer Credit (Early Settlement) Regulations 2004 regs 5-6)", "source": {"citation": {"url": "https://www.legislation.gov.uk/uksi/2004/1483/made", "quote": "the settlement date for calculation of the rebate may be deferred by", "title": "The Consumer Credit (Early Settlement) Regulations 2004", "publisher": "legislation.gov.uk (The National Archives)", "accessed": "2026-09-02"}}, "verified_at": "2026-09-02"},
    {"id": "FCA-CONC-5-2A-5", "kind": "policy", "rule": "CONC 5.2A.5", "claim": "An FCA rule requires a firm to undertake a creditworthiness assessment, and to have proper regard to its outcome for affordability risk, before entering into a regulated credit agreement.", "writer_line": "the affordability assessment duty (CONC 5.2A.5)", "source": {"citation": {"url": "https://handbook.fca.org.uk/handbook/CONC/5/2A.html", "quote": "undertaken a creditworthiness assessment", "title": "FCA Handbook - CONC 5.2A Creditworthiness assessment", "publisher": "Financial Conduct Authority", "accessed": "2026-09-02"}}, "verified_at": "2026-09-02"},
    {"id": "FCA-CONC-7-3-4", "kind": "policy", "rule": "CONC 7.3.4", "claim": "An FCA rule requires a firm to treat customers in or approaching arrears or in default with forbearance and due consideration.", "writer_line": "the forbearance duty (CONC 7.3.4)", "source": {"citation": {"url": "https://handbook.fca.org.uk/handbook/CONC/7/3.html", "quote": "A firm must treat customers in or approaching arrears or in default with forbearance and due consideration", "title": "FCA Handbook - CONC 7.3 Treatment of customers in or approaching arrears or in default (including repossessions): lenders, owners and debt collectors", "publisher": "Financial Conduct Authority", "accessed": "2026-09-02"}}, "verified_at": "2026-09-02"},
    {"id": "FCA-DISP-1-3-1", "kind": "policy", "rule": "DISP 1.3.1", "claim": "An FCA rule requires firms to establish, implement and maintain effective and transparent procedures for the reasonable and prompt handling of complaints - the process a borrower exhausts before escalating to the Financial Ombudsman Service.", "writer_line": "the complaints-handling requirement behind the route to the Financial Ombudsman (DISP 1.3.1)", "source": {"citation": {"url": "https://handbook.fca.org.uk/handbook/DISP/1/3.html", "quote": "Effective and transparent procedures for the reasonable and prompt handling of complaints must be established, implemented and maintained", "title": "FCA Handbook - DISP 1.3 Complaints handling rules", "publisher": "Financial Conduct Authority", "accessed": "2026-09-02"}}, "verified_at": "2026-09-02"}
  ], "banned_claims": []}'::jsonb,
  'manual',
  NULL,
  'loancalculator_couk lane (migration 699)',
  true,
  true,
  'Twelve citations - eight statutory (Consumer Credit Act 1974 ss.56/66A/87/95A/97/99/100 + SIs 1983/1564 and 2004/1483) and three FCA Handbook rules (CONC 5.2A.5, CONC 7.3.4, DISP 1.3.1) - every quote verified through datahelpers.QuoteFoundInText via cmd/fcaquotecheck on 2026-09-02 with a negative control in the same run, every URL confirmed by title. Two facts carry corrects_site_citation: the served settlement page says ten working days (it is 12 - SI 1983/1564 reg 4), and the overpayment page attributes a 10%-of-balance ERC-free allowance to the CCA (the statutory threshold is 8000 pounds per 12 months - s.95A). Copy not touched - owner decision, per the 695 precedent.'
);

-- VERIFY as DO/RAISE. A verify block of bare SELECTs cannot stop the COMMIT.
DO $$
DECLARE nfacts int; ncited int; nrule int; ncorr int;
BEGIN
  SELECT jsonb_array_length(data->'facts') INTO nfacts FROM site_specs
   WHERE site_id = '0162cde4-633e-45e9-8ca6-87a6b2fe1d26' AND aspect = 'evidence_base' AND is_current;
  IF nfacts <> 12 THEN
    RAISE EXCEPTION '699 VERIFY: expected 12 facts, found %', nfacts;
  END IF;

  SELECT count(*) INTO ncited FROM site_specs s, jsonb_array_elements(s.data->'facts') f
   WHERE s.site_id = '0162cde4-633e-45e9-8ca6-87a6b2fe1d26' AND s.aspect = 'evidence_base' AND s.is_current
     AND (f->'source'->'citation'->>'url' LIKE 'https://handbook.fca.org.uk/%'
       OR f->'source'->'citation'->>'url' LIKE 'https://www.legislation.gov.uk/%')
     AND length(f->'source'->'citation'->>'quote') > 20;
  IF ncited <> 12 THEN
    RAISE EXCEPTION '699 VERIFY: expected 12 facts with a handbook/legislation URL and a substantive quote, found %', ncited;
  END IF;

  SELECT count(*) INTO nrule FROM site_specs s, jsonb_array_elements(s.data->'facts') f
   WHERE s.site_id = '0162cde4-633e-45e9-8ca6-87a6b2fe1d26' AND s.aspect = 'evidence_base' AND s.is_current
     AND f->>'rule' ~ '^(Consumer Credit|CONC |DISP )';
  IF nrule <> 12 THEN
    RAISE EXCEPTION '699 VERIFY: expected 12 facts naming a CCA/SI/CONC/DISP rule, found %', nrule;
  END IF;

  SELECT count(*) INTO ncorr FROM site_specs s, jsonb_array_elements(s.data->'facts') f
   WHERE s.site_id = '0162cde4-633e-45e9-8ca6-87a6b2fe1d26' AND s.aspect = 'evidence_base' AND s.is_current
     AND f ? 'corrects_site_citation';
  IF ncorr <> 2 THEN
    RAISE EXCEPTION '699 VERIFY: expected 2 facts recording a corrected site citation, found %', ncorr;
  END IF;

  RAISE NOTICE '699 OK: loancalculator evidence register created - 12 citations (9 statute/SI, 3 handbook), 2 correcting a live misattribution';
END $$;

COMMIT;
