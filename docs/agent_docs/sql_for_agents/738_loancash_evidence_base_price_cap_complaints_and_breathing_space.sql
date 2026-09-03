-- 738_loancash_evidence_base_price_cap_complaints_and_breathing_space.sql
--
-- Gives loancash.co.uk an evidence register: NINETEEN facts, each citing the FCA
-- Handbook rule or statutory provision that states it, with a verbatim quote; plus
-- the six-pattern sibling banned-claims set the other three lending sites carry.
--
-- WHY. RFC_060 §1d listed five register-less finance sites. Four are now done
-- (lendzy 695, loanzy 697, farmerinsurance 698+713, loancalculator 699).
-- loancash.co.uk was the last, and had ZERO evidence_base rows -- not an empty
-- register, an ABSENT one -- beside 14 other current specs, on a deployed site
-- serving 30 pages [MEASURED 2026-09-03]. Owner directed this lane to populate it,
-- 2026-09-03. Method: RUNBOOK_lendzy_co_uk.md §8, followed step for step.
--
-- ⚠ WHY THIS SITE IS THE HIGH-RISK ONE OF THE FIVE. The other four calculate or
-- compare; loancash EXPLAINS THE RULES THEMSELVES. Its 30 pages carry 338
-- regulatory-shaped sentences [MEASURED 2026-09-03, all 30 pages crawled at the
-- artefact with an invented-URL 404 control], of which only THREE cite a rule
-- number at all. The rest state hard figures -- 0.8%, £15, 100%, 8 weeks, 6 months,
-- 60 days, 3% per month -- in plain English with no citation. Every one of those is
-- a claim about what the law is, made to someone in financial difficulty, and until
-- this row exists nothing re-checks any of them.
--
-- WHAT THE REGISTER DISCHARGES. The loancash_couk_fca_validation lane (2026-08-11)
-- verified the three price-cap constants by hand and wrote: "What is still true is
-- the second half of the worry: nothing is checking ... What actually earns its keep
-- is something that reads the rulebook and shouts if it disagrees with what is on
-- our page." That is precisely this mechanism. That lane also named the complaint
-- deadline calculator as its highest-value unchecked item, because limitation
-- periods -- unlike the price cap, static since 02/01/2015 -- do move. Facts
-- FCA-DISP-2-8-2-1 / -2A / -2B register exactly those three deadlines.
--
-- EVERY QUOTE WAS VERIFIED THROUGH THE PRODUCTION MATCHER (cmd/fcaquotecheck ->
-- datahelpers.VisibleTextFromHTML / QuoteFoundInText) on 2026-09-03: 19/19 true,
-- with a deliberately-absent control returning false in every run, and every URL
-- confirmed by <title> (handbook.fca.org.uk 200s on wrong paths -- the documented
-- LANDMINE -- and legislation.gov.uk gets the same discipline). The stored URL is
-- the one that ANSWERED, after the redirect.
--
-- A NEAR-MISS, recorded because it is step 4 working: the DISP 2.8.2(2)(b)
-- three-year quote was first written with commas where the source has parentheses
-- ("became aware, or ought reasonably to have become aware,"). It returned FALSE on
-- the production matcher. Shipped, it would have classified as citation_lost drift
-- every day for ever -- a false alarm indistinguishable from a real one. Never
-- hand-transcribe a quote; paste it and let the matcher decide.
--
-- THREE FACTS CARRY corrects_site_citation -- the point of the exercise:
--   * CONC 5A.2.14 (default cap). Two pages frame the £15 as a per-missed-payment
--     fee ("can only be charged once per missed payment"; "one-off, regardless of
--     how many days late"). The rule caps default charges at £15 "whether in
--     relation to one breach or CUMULATIVELY in relation to multiple breaches of
--     the agreement" -- one £15 for the AGREEMENT. The site UNDERSTATES the
--     protection: a reader with two missed payments would accept a second £15 as
--     lawful. It is not.
--   * CONC 7.6.12 (continuous payment authority). guides/loan-sharks-and-illegal-
--     lending.html says a lender "cannot take more than one payment attempt of over
--     £1" without fresh permission. The limit is TWO requests that have been
--     REFUSED, and there is NO £1 threshold anywhere in CONC 7.6. The site's own
--     CPA page states the rule correctly, so this is an internal contradiction.
--   * CONC 5.2A.4 (affordability). guides/check-your-lender-is-authorised.html
--     attributes "the affordability checks" to "CONC 5A" -- the PRICE CAP chapter,
--     which contains no affordability rule. It is CONC 5.2A.4R/5.2A.5R, which the
--     site cites correctly on three other pages.
-- THE SERVED COPY IS NOT TOUCHED -- rewriting published prose on an automated
-- finding is authority the owner withheld (bugs_open/320 §15, the 695 precedent).
-- The copy repairs are his call, tracked in the lane.
--
-- THE META-LESSON HOLDS FOR A FOURTH LANE: lendzy 2 wrong of 7 · loanzy 1 ·
-- loancalculator 2 · loancash 3. Every lane that has read the source has found
-- errors. Run the method expecting to find them.
--
-- BANNED CLAIMS: the sibling lending set, and ONE WIDTH CHOICE THAT MATTERS.
-- All six patterns were compiled with the production prefix (regexp.Compile("(?i)"
-- + p), claims.go:468) and each was fired against a positive control AND against
-- this site's own copy: 6/6 compile, 6/6 match their positive, 0/6 match anything
-- on the 30 served pages [MEASURED 2026-09-03].
-- ⚠ The no-credit-check pattern is deliberately the NARROW loanzy/loancalculator
-- variant, which requires the product noun. lendzy.co.uk carries a wider bare
-- \bno credit checks?\b, and that variant FIRES on this site's correct consumer
-- advice that an employer salary advance involves "no interest and no credit check"
-- -- measured, 1 hit. Adopting "the shared set" without checking which width you
-- are copying would convict the site of its own accurate guidance. The set is not
-- one set; it is two, and the difference is invisible in a coverage count.
--
-- LIMIT ACCEPTED (695 §B5, structural): neither host has a rule-level URL, so the
-- daily refresher keeps QUOTES honest but not the `rule` field -- a section page
-- carries dozens of rules. Until RFC_060 §3d/Q6's rule-span checker ships, `rule`
-- is a HUMAN-VERIFIED field: checked by hand against each provision's own heading
-- on 2026-09-03.
--
-- Lane: docs/agent_docs/docs024_key_docs_latest/loancash_couk_fca_validation/
-- Rollback: 738_..._ROLLBACK.sql

BEGIN;

-- GUARD. A current register must not already exist (idx_site_specs_current is
-- UNIQUE on (site_id, aspect) WHERE is_current). If another session has written
-- one since this file was authored, ABORT rather than supersede work unseen.
DO $$
DECLARE n int;
BEGIN
  SELECT count(*) INTO n FROM site_specs
   WHERE site_id = 'ee4a8199-4f5b-4e2e-88ce-01e600721b74' AND aspect = 'evidence_base';
  IF n <> 0 THEN
    RAISE EXCEPTION '738 ABORT: loancash already has % evidence_base row(s) - read them before writing', n;
  END IF;
END $$;

INSERT INTO site_specs (site_id, aspect, data, source, source_agent, created_by, is_current, pinned, notes)
VALUES (
  'ee4a8199-4f5b-4e2e-88ce-01e600721b74',
  'evidence_base',
  '{"facts":[{"id":"FCA-CONC-5A-2-3","kind":"policy","rule":"CONC 5A.2.3","claim":"A firm must not enter into an agreement for high-cost short-term credit whose charges exceed, or are capable of exceeding, 0.8% per day of the amount of credit provided - and for this cap the amount of credit is the amount OUTSTANDING on the day in question (CONC 5A.2.7 R), not the sum originally advanced.","writer_line":"the 0.8%-per-day initial cost cap on high-cost short-term credit (CONC 5A.2.3)","source":{"citation":{"url":"https://handbook.fca.org.uk/handbook/conc5a/conc5as2","quote":"exceed or are capable of exceeding 0.8% of the amount of credit provided under the agreement calculated per day","title":"FCA Handbook - CONC 5A.2 Prohibition from entering into agreements for high-cost short-term credit","publisher":"Financial Conduct Authority","accessed":"2026-09-03"}},"verified_at":"2026-09-03"},{"id":"FCA-CONC-5A-2-2","kind":"policy","rule":"CONC 5A.2.2","claim":"A firm must not enter into an agreement for high-cost short-term credit whose charges, alone or combined with any other charge under the agreement or a connected agreement, exceed or are capable of exceeding the amount of credit provided - the 100% total cost cap.","writer_line":"the 100% total cost cap on high-cost short-term credit (CONC 5A.2.2)","source":{"citation":{"url":"https://handbook.fca.org.uk/handbook/conc5a/conc5as2","quote":"exceed or are capable of exceeding the amount of credit provided under the agreement","title":"FCA Handbook - CONC 5A.2 Prohibition from entering into agreements for high-cost short-term credit","publisher":"Financial Conduct Authority","accessed":"2026-09-03"}},"verified_at":"2026-09-03"},{"id":"FCA-CONC-5A-2-14","kind":"policy","rule":"CONC 5A.2.14","claim":"The £15 default cap is an AGGREGATE limit across the whole agreement: charges connected with a borrower breach must not exceed £15 whether in relation to one breach or CUMULATIVELY in relation to multiple breaches. Interest on a default charge, and on credit unpaid in breach, is separately capped at 0.8% per day.","writer_line":"the £15 cap on default charges, cumulative across every breach of the agreement (CONC 5A.2.14)","source":{"citation":{"url":"https://handbook.fca.org.uk/handbook/conc5a/conc5as2","quote":"cumulatively in relation to multiple breaches of the agreement) exceed or are capable of exceeding £15","title":"FCA Handbook - CONC 5A.2 Prohibition from entering into agreements for high-cost short-term credit","publisher":"Financial Conduct Authority","accessed":"2026-09-03"}},"verified_at":"2026-09-03","corrects_site_citation":"guides/the-payday-loan-price-cap.html says the default fee \"is capped at £15 and can only be charged once per missed payment, not once per day it stays missed\", and guides/jargon-buster.html says it is \"capped at £15, one-off, regardless of how many days late you are\". Both are right about what they DENY (no per-day accrual) and wrong about the unit they AFFIRM. CONC 5A.2.14R(1) caps default charges at £15 \"whether in relation to one breach or cumulatively in relation to multiple breaches of the agreement\" - it is one £15 for the AGREEMENT, not one per missed payment. A reader with two missed payments would accept a second £15 as lawful; it is not. The site understates the protection it exists to explain. Copy not touched; owner''s call."},{"id":"FCA-CONC-5A-2-10","kind":"policy","rule":"CONC 5A.2.10","claim":"A replacement (refinanced) high-cost short-term credit agreement cannot carry charges that, taken together with those under the earlier agreement, exceed the amount of credit provided under the combined effect of both - so rolling a loan over does not open a fresh 100% allowance.","writer_line":"rollovers stay inside the original total cost cap (CONC 5A.2.10)","source":{"citation":{"url":"https://handbook.fca.org.uk/handbook/conc5a/conc5as2","quote":"under the combined effect of the replacement agreement and the earlier agreement","title":"FCA Handbook - CONC 5A.2 Prohibition from entering into agreements for high-cost short-term credit","publisher":"Financial Conduct Authority","accessed":"2026-09-03"}},"verified_at":"2026-09-03"},{"id":"FCA-CONC-6-7-23","kind":"policy","rule":"CONC 6.7.23","claim":"A firm must not refinance high-cost short-term credit, other than by exercising forbearance, on more than two occasions - the two-rollover limit.","writer_line":"the two-rollover limit on high-cost short-term credit (CONC 6.7.23)","source":{"citation":{"url":"https://handbook.fca.org.uk/handbook/conc6/conc6s7","quote":"must not refinance high-cost short-term credit (other than by exercising forbearance) on more than two occasions","title":"FCA Handbook - CONC 6.7 Post contract: business practices","publisher":"Financial Conduct Authority","accessed":"2026-09-03"}},"verified_at":"2026-09-03"},{"id":"FCA-CONC-7-6-12","kind":"policy","rule":"CONC 7.6.12","claim":"A firm must not request a further payment under a continuous payment authority to collect a sum due for high-cost short-term credit once it has done so on two previous occasions in connection with the same agreement AND those two requests were refused.","writer_line":"the two-refused-attempts limit on a continuous payment authority (CONC 7.6.12)","source":{"citation":{"url":"https://handbook.fca.org.uk/handbook/conc7/conc7s6","quote":"on two previous occasions and those previous payment requests have been refused","title":"FCA Handbook - CONC 7.6 Exercise of continuous payment authority","publisher":"Financial Conduct Authority","accessed":"2026-09-03"}},"verified_at":"2026-09-03","corrects_site_citation":"guides/loan-sharks-and-illegal-lending.html says a regulated lender \"cannot take more than one payment attempt of over £1 through a continuous payment authority (CPA) without your fresh permission\". Two errors: the limit is TWO requests that have been REFUSED (CONC 7.6.12R), not more than one; and there is NO £1 threshold anywhere in CONC 7.6 - the phrase appears nowhere in the section. The site''s own dedicated page, guides/stopping-payments-the-cpa-rules.html, states the rule correctly (\"A lender can attempt to collect a payment using your CPA twice\"), so this is an internal contradiction, not a settled house view. Copy not touched; owner''s call."},{"id":"FCA-CONC-7-6-14","kind":"policy","rule":"CONC 7.6.14","claim":"A firm must not use a continuous payment authority to collect a sum due for high-cost short-term credit that is less than the full sum due at the time of the request, except under an agreed repayment plan the customer has expressly consented to.","writer_line":"the prohibition on part-payment collection by continuous payment authority (CONC 7.6.14)","source":{"citation":{"url":"https://handbook.fca.org.uk/handbook/conc7/conc7s6","quote":"if that sum is less than the full sum due at the time the request is made","title":"FCA Handbook - CONC 7.6 Exercise of continuous payment authority","publisher":"Financial Conduct Authority","accessed":"2026-09-03"}},"verified_at":"2026-09-03"},{"id":"FCA-CONC-5-2A-4","kind":"policy","rule":"CONC 5.2A.4","claim":"A firm must undertake a reasonable assessment of the creditworthiness of a customer before entering into a regulated credit agreement, and (CONC 5.2A.5 R) must be able to demonstrate it did so and had proper regard to the outcome for affordability risk.","writer_line":"the affordability assessment duty before lending (CONC 5.2A.4)","source":{"citation":{"url":"https://handbook.fca.org.uk/handbook/conc5/conc5s6","quote":"A firm must undertake a reasonable assessment of the creditworthiness of a customer before","title":"FCA Handbook - CONC 5.2A Creditworthiness assessment","publisher":"Financial Conduct Authority","accessed":"2026-09-03"}},"verified_at":"2026-09-03","corrects_site_citation":"guides/check-your-lender-is-authorised.html attributes \"the affordability checks\" to \"CONC 5A\". CONC 5A is the PRICE CAP chapter (\"Prohibition from entering into agreements for high-cost short-term credit\") and contains no affordability rule. The creditworthiness duty is CONC 5.2A.4R/5.2A.5R. The site cites CONC 5.2A correctly on three other pages, so this is a slip on one page rather than a house error. Copy not touched; owner''s call."},{"id":"FCA-DISP-1-6-2","kind":"policy","rule":"DISP 1.6.2","claim":"A respondent must, by the end of eight weeks after receiving a complaint, send the complainant a final response or a written response explaining why it cannot yet give one.","writer_line":"the eight-week deadline for a lender''s final response (DISP 1.6.2)","source":{"citation":{"url":"https://handbook.fca.org.uk/handbook/disp1/disp1s6","quote":"by the end of eight weeks after its receipt of the complaint","title":"FCA Handbook - DISP 1.6 Complaints time limit rules","publisher":"Financial Conduct Authority","accessed":"2026-09-03"}},"verified_at":"2026-09-03"},{"id":"FCA-DISP-2-8-2-1","kind":"policy","rule":"DISP 2.8.2(1)","claim":"The Financial Ombudsman Service cannot consider a complaint referred more than six months after the date the firm sent its final response, redress determination or summary resolution communication.","writer_line":"the six-month window to refer a complaint to the Ombudsman (DISP 2.8.2(1))","source":{"citation":{"url":"https://handbook.fca.org.uk/handbook/disp2/disp2s8","quote":"more than six months after the date on which the respondent sent the complainant its final response","title":"FCA Handbook - DISP 2.8 Was the complaint referred to the Financial Ombudsman Service in time?","publisher":"Financial Conduct Authority","accessed":"2026-09-03"}},"verified_at":"2026-09-03"},{"id":"FCA-DISP-2-8-2-2A","kind":"policy","rule":"DISP 2.8.2(2)(a)","claim":"The Ombudsman also cannot consider a complaint referred more than six years after the event complained of, unless the later three-year limb applies.","writer_line":"the six-year limit from the event complained of (DISP 2.8.2(2)(a))","source":{"citation":{"url":"https://handbook.fca.org.uk/handbook/disp2/disp2s8","quote":"six years after the event complained of","title":"FCA Handbook - DISP 2.8 Was the complaint referred to the Financial Ombudsman Service in time?","publisher":"Financial Conduct Authority","accessed":"2026-09-03"}},"verified_at":"2026-09-03"},{"id":"FCA-DISP-2-8-2-2B","kind":"policy","rule":"DISP 2.8.2(2)(b)","claim":"Where it is later than the six-year limit, the Ombudsman may still consider a complaint referred within three years of the date the complainant became aware, or ought reasonably to have become aware, that they had cause for complaint.","writer_line":"the three-year-from-awareness limb of the Ombudsman time limit (DISP 2.8.2(2)(b))","source":{"citation":{"url":"https://handbook.fca.org.uk/handbook/disp2/disp2s8","quote":"three years from the date on which the complainant became aware (or ought reasonably to have become aware) that he had cause for complaint","title":"FCA Handbook - DISP 2.8 Was the complaint referred to the Financial Ombudsman Service in time?","publisher":"Financial Conduct Authority","accessed":"2026-09-03"}},"verified_at":"2026-09-03"},{"id":"SI-2020-1311-REG26","kind":"policy","rule":"Debt Respite Scheme Regulations 2020, reg 26(2)","claim":"A breathing space moratorium continues for 60 days from the day it starts, unless it ends earlier on the debtor''s death or is cancelled.","writer_line":"the 60-day duration of a standard Breathing Space (SI 2020/1311 reg 26)","source":{"citation":{"url":"https://www.legislation.gov.uk/uksi/2020/1311/regulation/26/made","quote":"A moratorium continues for 60 days beginning with the date on which it started","title":"The Debt Respite Scheme (Breathing Space Moratorium and Mental Health Crisis Moratorium) (England and Wales) Regulations 2020","publisher":"legislation.gov.uk (The National Archives)","accessed":"2026-09-03"}},"verified_at":"2026-09-03"},{"id":"SI-2020-1311-REG32","kind":"policy","rule":"Debt Respite Scheme Regulations 2020, reg 32(2)(a)","claim":"A mental health crisis moratorium ends on the earliest of several events, one of which is the end of the period of 30 days beginning with the day the debtor stops receiving mental health crisis treatment - so it runs for the duration of treatment plus 30 days.","writer_line":"mental health crisis Breathing Space runs for the treatment plus 30 days (SI 2020/1311 reg 32)","source":{"citation":{"url":"https://www.legislation.gov.uk/uksi/2020/1311/regulation/32/made","quote":"the end of the period of 30 days beginning with the day on which the debtor stops receiving mental health crisis treatment","title":"The Debt Respite Scheme (Breathing Space Moratorium and Mental Health Crisis Moratorium) (England and Wales) Regulations 2020","publisher":"legislation.gov.uk (The National Archives)","accessed":"2026-09-03"}},"verified_at":"2026-09-03"},{"id":"SI-2020-1311-REG24","kind":"policy","rule":"Debt Respite Scheme Regulations 2020, reg 24(3)(g)","claim":"A debtor is eligible for a breathing space moratorium only if any previous breathing space moratorium ended more than 12 months before the date of the application.","writer_line":"the 12-month gap required between standard Breathing Spaces (SI 2020/1311 reg 24(3)(g))","source":{"citation":{"url":"https://www.legislation.gov.uk/uksi/2020/1311/regulation/24/made","quote":"that moratorium ended more than 12 months before the date of the application","title":"The Debt Respite Scheme (Breathing Space Moratorium and Mental Health Crisis Moratorium) (England and Wales) Regulations 2020","publisher":"legislation.gov.uk (The National Archives)","accessed":"2026-09-03"}},"verified_at":"2026-09-03"},{"id":"SI-2020-1311-REG16","kind":"policy","rule":"Debt Respite Scheme Regulations 2020, reg 16(2)(b)","claim":"During a breathing space moratorium the debtor must make any payment due in relation to an ongoing liability as it falls due - rent, mortgage, tax and similar bills arising during the 60 days are still payable.","writer_line":"ongoing liabilities remain payable during a Breathing Space (SI 2020/1311 reg 16(2)(b))","source":{"citation":{"url":"https://www.legislation.gov.uk/uksi/2020/1311/regulation/16/made","quote":"make any payment due in relation to an ongoing liability as it falls due to be paid during the moratorium period","title":"The Debt Respite Scheme (Breathing Space Moratorium and Mental Health Crisis Moratorium) (England and Wales) Regulations 2020","publisher":"legislation.gov.uk (The National Archives)","accessed":"2026-09-03"}},"verified_at":"2026-09-03"},{"id":"SI-2013-2589-ART2","kind":"policy","rule":"Credit Unions (Maximum Interest Rate on Loans) Order 2013, art 2","claim":"The maximum interest rate a credit union may charge on a loan, specified for the purposes of section 11(5) of the Credit Unions Act 1979, is three per cent per month (in force 1 April 2014, raised from 2%).","writer_line":"the 3%-per-month statutory ceiling on credit union loan interest (SI 2013/2589 art 2)","source":{"citation":{"url":"https://www.legislation.gov.uk/uksi/2013/2589/made","quote":"The rate specified for the purposes of section 11(5) of the Credit Unions Act 1979 is three per cent per month","title":"The Credit Unions (Maximum Interest Rate on Loans) Order 2013","publisher":"legislation.gov.uk (The National Archives)","accessed":"2026-09-03"}},"verified_at":"2026-09-03"},{"id":"FSMA-2000-S19","kind":"policy","rule":"Financial Services and Markets Act 2000 s.19","claim":"No person may carry on a regulated activity in the United Kingdom, or purport to do so, unless they are an authorised person or an exempt person - the general prohibition.","writer_line":"the general prohibition on unauthorised regulated activity (FSMA 2000 s.19)","source":{"citation":{"url":"https://www.legislation.gov.uk/ukpga/2000/8/section/19","quote":"No person may carry on a regulated activity in the United Kingdom, or purport to do so, unless he is","title":"Financial Services and Markets Act 2000","publisher":"legislation.gov.uk (The National Archives)","accessed":"2026-09-03"}},"verified_at":"2026-09-03"},{"id":"FSMA-2000-S23","kind":"policy","rule":"Financial Services and Markets Act 2000 s.23","claim":"A person who contravenes the general prohibition commits a criminal offence, punishable on summary conviction by up to six months imprisonment and on indictment by up to two years - the statutory basis for describing unauthorised lending as illegal.","writer_line":"unauthorised lending is a criminal offence (FSMA 2000 s.23)","source":{"citation":{"url":"https://www.legislation.gov.uk/ukpga/2000/8/section/23","quote":"A person who contravenes the general prohibition is guilty of an offence and liable","title":"Financial Services and Markets Act 2000","publisher":"legislation.gov.uk (The National Archives)","accessed":"2026-09-03"}},"verified_at":"2026-09-03"}],"banned_claims":[{"pattern":"\\bguaranteed (acceptance|approval|loans?|yes)\\b","reason":"No guaranteed-acceptance language, ever: unprovable, a financial-promotion breach, and the marker of the unauthorised end of this market. Sibling set, adopted verbatim from loanzy.uk / loancalculator.co.uk."},{"pattern":"\\b(everyone|anyone|all applicants) (is|are|will be) (accepted|approved)\\b","reason":"The same promise wearing a different grammar. Sibling set, verbatim."},{"pattern":"\\bno[- ]credit[- ]check (loans?|lending|borrowing)\\b|\\bloans? with no credit checks?\\b","reason":"A no-credit-check loan is the marker of the unregulated end. NARROW variant deliberately: it requires the product noun, so it does not fire on this site''s correct consumer advice that an employer salary advance involves \"no interest and no credit check\". lendzy.co.uk''s wider \\bno credit checks?\\b DOES fire on that sentence - measured 2026-09-03."},{"pattern":"\\bbad credit (is )?(no|not a) (problem|issue|barrier)\\b","reason":"Dismisses the reader''s actual situation and implies an outcome. Sibling set, verbatim."},{"pattern":"\\b(we|our team) (can|will) (get|secure) you (a|the) (loan|deal|approval)\\b","reason":"loancash is an information site: it lends nothing and arranges nothing. Sibling set, verbatim."},{"pattern":"\\b(we lend|borrow from us|apply (with|through) us|we are a (lender|broker)|we (can )?arrange (a |your )?loan)\\b","reason":"Archetype constraint: never reposition a rules-explainer as a lender or broker. Adopted from loancalculator.co.uk, whose archetype is the same."}]}'::jsonb,
  'manual',
  NULL,
  'loancash_couk_fca_validation lane (migration 738)',
  true,
  true,
  'Nineteen citations - eleven FCA Handbook rules (CONC 5A.2.2/2.3/2.10/2.14, CONC 6.7.23, CONC 7.6.12/7.6.14, CONC 5.2A.4, DISP 1.6.2, DISP 2.8.2(1) and 2.8.2(2)(a)+(b)) and eight statutory (Debt Respite Scheme Regs 2020 regs 16/24/26/32, Credit Unions Maximum Interest Rate Order 2013 art 2, FSMA 2000 ss.19 and 23) - every quote verified through datahelpers.QuoteFoundInText via cmd/fcaquotecheck on 2026-09-03 with a negative control in the same run, every URL confirmed by title. Three facts carry corrects_site_citation: the £15 default cap is cumulative across the agreement not per missed payment, the CPA limit is two REFUSED attempts with no £1 threshold, and affordability is CONC 5.2A not CONC 5A. Copy not touched - owner decision, per the 695 precedent. Six banned_claims adopted from the sibling lending set, all compiled with the production (?i) prefix and measured against the served pages: 0 hits.'
);

-- VERIFY as DO/RAISE. A verify block of bare SELECTs cannot stop the COMMIT.
DO $$
DECLARE nfacts int; ncited int; nrule int; ncorr int; nban int;
BEGIN
  SELECT jsonb_array_length(data->'facts') INTO nfacts FROM site_specs
   WHERE site_id = 'ee4a8199-4f5b-4e2e-88ce-01e600721b74' AND aspect = 'evidence_base' AND is_current;
  IF nfacts <> 19 THEN
    RAISE EXCEPTION '738 VERIFY: expected 19 facts, found %', nfacts;
  END IF;

  SELECT count(*) INTO ncited FROM site_specs s, jsonb_array_elements(s.data->'facts') f
   WHERE s.site_id = 'ee4a8199-4f5b-4e2e-88ce-01e600721b74' AND s.aspect = 'evidence_base' AND s.is_current
     AND (f->'source'->'citation'->>'url' LIKE 'https://handbook.fca.org.uk/%'
       OR f->'source'->'citation'->>'url' LIKE 'https://www.legislation.gov.uk/%')
     AND length(f->'source'->'citation'->>'quote') > 20;
  IF ncited <> 19 THEN
    RAISE EXCEPTION '738 VERIFY: expected 19 facts with a handbook/legislation URL and a substantive quote, found %', ncited;
  END IF;

  SELECT count(*) INTO nrule FROM site_specs s, jsonb_array_elements(s.data->'facts') f
   WHERE s.site_id = 'ee4a8199-4f5b-4e2e-88ce-01e600721b74' AND s.aspect = 'evidence_base' AND s.is_current
     AND f->>'rule' ~ '^(CONC |DISP |Debt Respite|Credit Unions|Financial Services)';
  IF nrule <> 19 THEN
    RAISE EXCEPTION '738 VERIFY: expected 19 facts naming a CONC/DISP/statutory rule, found %', nrule;
  END IF;

  SELECT count(*) INTO ncorr FROM site_specs s, jsonb_array_elements(s.data->'facts') f
   WHERE s.site_id = 'ee4a8199-4f5b-4e2e-88ce-01e600721b74' AND s.aspect = 'evidence_base' AND s.is_current
     AND f ? 'corrects_site_citation';
  IF ncorr <> 3 THEN
    RAISE EXCEPTION '738 VERIFY: expected 3 facts recording a corrected site citation, found %', ncorr;
  END IF;

  SELECT jsonb_array_length(data->'banned_claims') INTO nban FROM site_specs
   WHERE site_id = 'ee4a8199-4f5b-4e2e-88ce-01e600721b74' AND aspect = 'evidence_base' AND is_current;
  IF nban <> 6 THEN
    RAISE EXCEPTION '738 VERIFY: expected 6 banned_claims patterns, found %', nban;
  END IF;

  RAISE NOTICE '738 OK: loancash evidence register created - 19 citations (11 handbook, 8 statutory), 3 correcting a live misstatement, 6 banned-claim patterns';
END $$;

COMMIT;
