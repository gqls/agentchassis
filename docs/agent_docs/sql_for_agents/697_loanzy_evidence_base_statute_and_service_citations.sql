-- 697_loanzy_evidence_base_statute_and_service_citations.sql
--
-- Gives loanzy.uk an evidence register: THREE facts — the statutory claim its
-- pages lean on hardest, and the provenance of the two services it signposts.
-- Owner instruction 2026-09-02 via the lendzy lane ("tell the other sites to do
-- theirs too"); method = lendzy RUNBOOK §8 / migration 695, RFC_060 §4.
-- loanzy had ZERO evidence_base rows [MEASURED 2026-09-02], so its numeric scan
-- never arms and its daily refresher run is VACUOUSLY clean.
--
-- EVERY QUOTE VERIFIED THROUGH THE PRODUCTION MATCHER (cmd/fcaquotecheck ->
-- datahelpers.VisibleTextFromHTML/QuoteFoundInText) on 2026-09-02, each with a
-- deliberately-absent control returning false in the same run.
--
-- ONE FACT CARRIES corrects_site_citation — the lendzy-class finding here:
-- loanzy's served pages describe the Money and Pensions Service as
-- "FCA-authorised" / "FCA-regulated" alongside StepChange. StepChange IS
-- FCA-regulated (its own page states it, cited below). MaPS is NOT an
-- FCA-authorised firm — it is the government-sponsored guidance body
-- (gov.uk, cited below). The substantive signposting is sound; the
-- ATTRIBUTION of regulatory status to MaPS is wrong. Copy not touched —
-- rewriting published prose on an automated finding is authority the owner
-- withheld (bugs_open/320 §15); the finding is routed to the copy lane.
--
-- SOURCE-STABILITY NOTE (the failure case the lendzy lane asked to hear
-- about): maps.org.uk and moneyhelper.org.uk both sit behind a Cloudflare
-- challenge ("Just a moment...") — a citation there would classify as drift
-- for ever. The MaPS fact therefore cites gov.uk (curl-clean, stable), found
-- at the organisation's gov.uk slug 'single-financial-guidance-body' (the
-- body's founding name; the page titles and describes MaPS).
--
-- Lane: docs/agent_docs/docs024_key_docs_latest/loanzy_uk_example_site/
-- Rollback: 697_..._ROLLBACK.sql
\set ON_ERROR_STOP on
BEGIN;

DO $$
DECLARE n int;
BEGIN
  SELECT count(*) INTO n FROM site_specs
   WHERE site_id = '55213ded-03ec-40f7-8fc1-169de05e05c8' AND aspect = 'evidence_base';
  IF n <> 0 THEN
    RAISE EXCEPTION '697 ABORT: loanzy already has % evidence_base row(s) - read them before writing', n;
  END IF;
END $$;

INSERT INTO site_specs (site_id, aspect, data, source, source_agent, created_by, is_current, pinned, notes)
VALUES (
  '55213ded-03ec-40f7-8fc1-169de05e05c8',
  'evidence_base',
  '{"facts": [{"id": "LAW-CCA-66A", "kind": "policy", "rule": "Consumer Credit Act 1974 s.66A", "claim": "A debtor may withdraw from a regulated consumer credit agreement, without giving any reason, within 14 days beginning with the day after the relevant day.", "writer_line": "your legal right under the Consumer Credit Act 1974 to withdraw from a regulated credit agreement within 14 days of signing, without needing to give a reason", "source": {"citation": {"url": "https://www.legislation.gov.uk/ukpga/1974/39/section/66A", "quote": "before the end of the period of 14 days beginning with the day after the relevant day", "title": "Consumer Credit Act 1974 - Section 66A", "publisher": "legislation.gov.uk", "accessed": "2026-09-02"}}, "verified_at": "2026-09-02"}, {"id": "SVC-STEPCHANGE-FCA", "kind": "policy", "rule": "FCA authorisation (firm status)", "claim": "StepChange Debt Charity is authorised and regulated by the Financial Conduct Authority.", "writer_line": "free help is available from FCA-regulated services such as StepChange", "source": {"citation": {"url": "https://www.stepchange.org/about-us.aspx", "quote": "authorised and regulated by the Financial Conduct Authority", "title": "About StepChange Debt Charity", "publisher": "StepChange Debt Charity", "accessed": "2026-09-02"}}, "verified_at": "2026-09-02"}, {"id": "SVC-MAPS-GOV", "kind": "policy", "rule": "Financial Guidance and Claims Act 2018 (body status)", "claim": "The Money and Pensions Service (MaPS) is the government-sponsored financial guidance body, successor to the Money Advice Service, the Pensions Advisory Service and Pension Wise; it is not an FCA-authorised firm.", "writer_line": "free, independent guidance is available from ... the Money and Pensions Service (MoneyHelper)", "source": {"citation": {"url": "https://www.gov.uk/government/organisations/single-financial-guidance-body", "quote": "The Money and Pensions Service (MaPS) replaces the 3 existing providers of government-sponsored financial guidance", "title": "Money and Pensions Service - GOV.UK", "publisher": "GOV.UK", "accessed": "2026-09-02"}}, "verified_at": "2026-09-02", "corrects_site_citation": "Served pages group MaPS under \"FCA-authorised services\" / \"FCA-regulated services\" alongside StepChange. MaPS is a government-sponsored statutory body, not an FCA-authorised firm - the signposting is sound, the regulatory attribution is not. Copy repair is the owner/copy lane call."}], "banned_claims": []}'::jsonb,
  'manual',
  NULL,
  'loanzy_uk_example_site lane (migration 697)',
  true,
  true,
  'Three citations: CCA 1974 s.66A (the 14-day withdrawal right the pages state), StepChange FCA status (their own page), MaPS provenance (gov.uk - NOT an FCA firm; corrects the served copy grouping it under FCA-authorised services). Every quote verified through cmd/fcaquotecheck with a negative control, 2026-09-02. maps.org.uk/moneyhelper.org.uk rejected as sources: Cloudflare challenge = perpetual false drift.'
);

DO $$
DECLARE nfacts int; ncited int; ncorr int;
BEGIN
  SELECT jsonb_array_length(data->'facts') INTO nfacts FROM site_specs
   WHERE site_id = '55213ded-03ec-40f7-8fc1-169de05e05c8' AND aspect = 'evidence_base' AND is_current;
  IF nfacts <> 3 THEN RAISE EXCEPTION '697 VERIFY: expected 3 facts, found %', nfacts; END IF;
  SELECT count(*) INTO ncited FROM site_specs s, jsonb_array_elements(s.data->'facts') f
   WHERE s.site_id = '55213ded-03ec-40f7-8fc1-169de05e05c8' AND s.aspect='evidence_base' AND s.is_current
     AND f->'source'->'citation'->>'url' ~ '^https://(www\.legislation\.gov\.uk|www\.stepchange\.org|www\.gov\.uk)/'
     AND length(f->'source'->'citation'->>'quote') > 20;
  IF ncited <> 3 THEN RAISE EXCEPTION '697 VERIFY: expected 3 substantively-quoted facts at stable hosts, found %', ncited; END IF;
  SELECT count(*) INTO ncorr FROM site_specs s, jsonb_array_elements(s.data->'facts') f
   WHERE s.site_id = '55213ded-03ec-40f7-8fc1-169de05e05c8' AND s.aspect='evidence_base' AND s.is_current
     AND f ? 'corrects_site_citation';
  IF ncorr <> 1 THEN RAISE EXCEPTION '697 VERIFY: expected 1 corrects_site_citation fact, found %', ncorr; END IF;
  RAISE NOTICE '697 OK: loanzy evidence register created - 3 citations, 1 correcting a live misattribution';
END $$;
COMMIT;
