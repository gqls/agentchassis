-- 695_lendzy_evidence_base_fca_handbook_citations.sql
--
-- Gives lendzy.co.uk an evidence register: EIGHT facts, each citing the FCA
-- Handbook rule that states it, with a verbatim quote.
--
-- WHY THIS FIRST, AND WHY IT IS NOT CODE. RFC_060 section 4: "populating registers for the
-- five register-less finance sites ... belongs to the site lanes, needs no RFC,
-- and would deliver most of the benefit of (ii) without any code at all. If only
-- one thing happens as a result of this RFC, it should be that." lendzy is one of
-- the five: it has ZERO evidence_base rows [MEASURED 2026-09-02], so its numeric
-- scan never arms and its daily refresher run is VACUOUSLY clean. A clean run
-- over an empty register means nothing, which is the worst kind of green.
--
-- Owner instruction 2026-09-02: make "checked against the FCA handbook, rule by
-- rule" TRUE rather than delete it (the phrase itself was bugs_closed/414). The
-- daily refresher (refresh_evidence_base_action.go + evidence_citations.go,
-- CLM-007/008, owned by the claims-verification lane) re-fetches every citation
-- URL and re-checks its verbatim quote against visible text, classifying 403/5xx
-- as unknown and a 200-with-quote-gone as citation_lost. THAT IS THE OWNER'S
-- "check with their online version each time", and it starts working on lendzy
-- the moment these rows exist. No new code, no new dependency.
--
-- EVERY QUOTE WAS VERIFIED THROUGH THE PRODUCTION MATCHER BEFORE BEING WRITTEN
-- HERE, not through a re-implementation of it. cmd/fcaquotecheck calls
-- datahelpers.VisibleTextFromHTML and datahelpers.QuoteFoundInText -- the exact
-- functions the refresher calls -- and every one of the eight returned true with
-- a deliberately-absent control returning false in the same run. A mirror of the
-- extraction passes happily while production disagrees; that is why the probe
-- exists. A quote that does not match is classified as DRIFT every day for ever,
-- and that false alarm is indistinguishable from a real one.
--
-- TWO FACTS CARRY corrects_site_citation, AND THIS IS THE POINT OF THE EXERCISE.
-- Checking each cited rule against the handbook text found two wrong attributions
-- in lendzy's live copy [MEASURED 2026-09-02]:
--   * the two-rollover limit is CONC 6.7.23, not CONC 6.7.17 (the definitions rule);
--   * the two-attempt continuous-payment-authority limit is CONC 7.6.12, not
--     CONC 6.7.23 (the refinance rule, a different chapter).
-- Both SUBSTANTIVE claims are true; only the attributions shifted. This migration
-- records the CORRECT rule in the register and notes what the page currently says.
-- IT DOES NOT TOUCH THE SERVED COPY -- rewriting published prose on an automated
-- finding is authority the owner withheld (bugs_open/320 section 15), so the copy
-- repair is his call and is tracked separately in the lane.
--
-- LIMIT OF WHAT THE DAILY CHECK CAN SEE (PLAN section B5, and it is structural):
-- handbook.fca.org.uk has NO rule-level URL. CONC 6.7 is 54 rules on one page
-- carrying both 6.7.17 and 6.7.23, so a quote from one verifies against a citation
-- naming the other -- same page, same bytes. The refresher therefore keeps these
-- quotes honest but CANNOT keep the `rule` field honest. That gap is filed as
-- RFC_060 section 3d / Q6 by the claims-verification lane, with a fix sketch.
-- Until it is built, `rule` is a HUMAN-VERIFIED field: it was checked by hand on
-- 2026-09-02 against the rule's own heading, and re-checking it is a human job.
--
-- Also relevant: handbook.fca.org.uk returns HTTP 200 for EVERY path, invented
-- rules included -- see LANDMINES.md. The eight URLs here were each confirmed by
-- <title>, never by status.
--
-- ROUND 2 (2026-09-02), answering the council's round-1 objections with evidence:
--
-- (a) editquality HIGH, the lossy round-trip: the hazard is real and DOCUMENTED,
--     and its own landmine text is the answer — the typed structs
--     (EvidenceBase/EvidenceFact/EvidenceSource) are lossy, and precisely for
--     that reason "the two live writers avoid it... both work on
--     map[string]interface{} and marshal that (refresh_evidence_base_action.go:683,
--     evidence_citations.go:350). That is deliberate." So the daily refresher —
--     the consumer this migration relies on — does NOT delete `rule` or
--     `corrects_site_citation`. What the objection still buys, because a FUTURE
--     writer could round-trip through the struct: the two corrections are now
--     ALSO recorded outside the round-trip surface entirely — a doc_notes row
--     (below), plus site_specs.notes (a real column, not inside data), plus the
--     lane NOTES and migration 696 itself. Losing the jsonb fields would now
--     lose a convenience, not the record.
--
-- (b) pinned (editquality med, prior_art med): NOT load-bearing, stated plainly —
--     write_site_spec ignores and drops `pinned` (its landmine). It is set as
--     convention only; durability rests on (a)'s analysis and the doc_notes row.
--
-- (c) the truncated site_id (two seats, low): review-sketch shorthand only. THIS
--     file has carried the full UUID throughout; a guard below now also resolves
--     the id against `sites` and aborts unless it names lendzy.co.uk.
--
-- (d) prior_art med, the dispatch absence: NAMED AND MEASURED. Scheduled task
--     `evidence-freshness`: enabled, interval 86400s, last_triggered_at
--     2026-09-02 09:08:57Z [MEASURED]. Its site selection is
--     `SELECT site_id FROM site_specs WHERE aspect='evidence_base' AND
--     is_current=true` (refresh_evidence_base_action.go:278) — fleet-wide by
--     aspect, no per-site registration. Lendzy is reached on the first daily
--     tick after this row exists, with no further action.
--
-- (e) compliance med, urgency of the live wrong citations: OVERTAKEN BY EVENTS
--     the same day — the owner ruled "please fix both"; migration 696 (committed,
--     council corr bb352ee8) corrects every storage layer including the
--     content_direction spec and the loancash fork, and files 11 rerenders.
--     Apply is sequenced after the announced chassis roll settles. Not routine,
--     and no longer merely tracked: fixed, pending apply.
--
-- (f) compliance low, the standing gap: correct, and it stays NAMED — a Handbook
--     re-numbering could invalidate `rule` silently, because the daily check
--     verifies URL+quote, not the rule id. The closer is the rule-span checker
--     (RFC_060 §3d/Q6), owner-approved to build 2026-09-02, claims-verification
--     lane's. Until it ships, `rule` is a human-verified field, dated.
--
-- Lane: docs/agent_docs/docs024_key_docs_latest/lendzy_co_uk/
-- Rollback: 695_..._ROLLBACK.sql

BEGIN;

-- GUARD. A current register must not already exist (idx_site_specs_current is
-- UNIQUE on (site_id, aspect) WHERE is_current). If another session has written
-- one since this file was authored, ABORT rather than supersede work unseen.
DO $$
DECLARE n int;
BEGIN
  SELECT count(*) INTO n FROM site_specs
   WHERE site_id = '8ff093d5-1f19-453b-9439-a10379bbcd76' AND aspect = 'evidence_base';
  IF n <> 0 THEN
    RAISE EXCEPTION '695 ABORT: lendzy already has % evidence_base row(s) - read them before writing', n;
  END IF;

  -- (c): the UUID must resolve to the site this migration believes it names.
  SELECT count(*) INTO n FROM sites
   WHERE id = '8ff093d5-1f19-453b-9439-a10379bbcd76' AND domain = 'lendzy.co.uk';
  IF n <> 1 THEN
    RAISE EXCEPTION '695 ABORT: site 8ff093d5-1f19-453b-9439-a10379bbcd76 does not resolve to lendzy.co.uk';
  END IF;
END $$;

INSERT INTO site_specs (site_id, aspect, data, source, source_agent, created_by, is_current, pinned, notes)
VALUES (
  '8ff093d5-1f19-453b-9439-a10379bbcd76',
  'evidence_base',
  '{"facts": [{"id": "FCA-CONC-5A-2-3", "kind": "policy", "rule": "CONC 5A.2.3", "claim": "An FCA rule caps the charges under a high-cost short-term credit agreement at 0.8% of the amount of credit per day.", "writer_line": "the 0.8% per day cost cap (CONC 5A.2.3)", "source": {"citation": {"url": "https://handbook.fca.org.uk/handbook/conc5a", "quote": "exceed or are capable of exceeding 0.8% of the amount of credit provided under the agreement calculated per day", "title": "FCA Handbook - CONC 5A Cost cap for high-cost short-term credit", "publisher": "Financial Conduct Authority", "accessed": "2026-09-02"}}, "verified_at": "2026-09-02"}, {"id": "FCA-CONC-5A-2-2", "kind": "policy", "rule": "CONC 5A.2.2", "claim": "An FCA rule caps the total charges under a high-cost short-term credit agreement at the amount of credit provided, so a borrower can never be charged more in interest and fees than they borrowed.", "writer_line": "the 100% total cost cap (CONC 5A.2.2)", "source": {"citation": {"url": "https://handbook.fca.org.uk/handbook/conc5a", "quote": "exceed or are capable of exceeding the amount of credit provided under the agreement", "title": "FCA Handbook - CONC 5A Cost cap for high-cost short-term credit", "publisher": "Financial Conduct Authority", "accessed": "2026-09-02"}}, "verified_at": "2026-09-02"}, {"id": "FCA-CONC-5A-2-14", "kind": "policy", "rule": "CONC 5A.2.14", "claim": "An FCA rule caps default charges under a high-cost short-term credit agreement at 15 pounds in total, whether for one breach or cumulatively across several.", "writer_line": "the £15 default fee cap (CONC 5A.2.14)", "source": {"citation": {"url": "https://handbook.fca.org.uk/handbook/conc5a", "quote": "cumulatively in relation to multiple breaches of the agreement) exceed or are capable of exceeding £15", "title": "FCA Handbook - CONC 5A Cost cap for high-cost short-term credit", "publisher": "Financial Conduct Authority", "accessed": "2026-09-02"}}, "verified_at": "2026-09-02"}, {"id": "FCA-CONC-6-7-23", "kind": "policy", "rule": "CONC 6.7.23", "claim": "An FCA rule prohibits a firm from refinancing high-cost short-term credit on more than two occasions, other than by exercising forbearance.", "writer_line": "the two-rollover limit (CONC 6.7.23)", "source": {"citation": {"url": "https://handbook.fca.org.uk/handbook/CONC/6/7.html", "quote": "must not refinance high-cost short-term credit (other than by exercising forbearance) on more than two occasions", "title": "FCA Handbook - CONC 6.7 Post contract: business practices", "publisher": "Financial Conduct Authority", "accessed": "2026-09-02"}}, "verified_at": "2026-09-02", "corrects_site_citation": "CONC 6.7.17 - the site attributes the two-rollover limit to the DEFINITIONS rule. 6.7.17 defines ''refinance'' for the range CONC 6.7.18 R to CONC 6.7.23 R; it does not state the limit."}, {"id": "FCA-CONC-7-6-12", "kind": "policy", "rule": "CONC 7.6.12", "claim": "An FCA rule prohibits a firm from making a further continuous payment authority request for a sum due for high-cost short-term credit after two previous requests on the same agreement have been refused.", "writer_line": "the two-attempt limit on continuous payment authority (CONC 7.6.12)", "source": {"citation": {"url": "https://handbook.fca.org.uk/handbook/CONC/7/6.html", "quote": "on two previous occasions and those previous payment requests have been refused", "title": "FCA Handbook - CONC 7.6 Exercise of continuous payment authority", "publisher": "Financial Conduct Authority", "accessed": "2026-09-02"}}, "verified_at": "2026-09-02", "corrects_site_citation": "CONC 6.7.23 - the site attributes the card-payment limit to the REFINANCE rule, which is in a different chapter and says nothing about payment requests."}, {"id": "FCA-CONC-5-2A-5", "kind": "policy", "rule": "CONC 5.2A.5", "claim": "An FCA rule requires a firm to undertake a creditworthiness assessment, and to have proper regard to its outcome for affordability risk, before entering into a regulated credit agreement.", "writer_line": "the affordability assessment duty (CONC 5.2A.5)", "source": {"citation": {"url": "https://handbook.fca.org.uk/handbook/CONC/5/2A.html", "quote": "undertaken a creditworthiness assessment", "title": "FCA Handbook - CONC 5.2A Creditworthiness assessment", "publisher": "Financial Conduct Authority", "accessed": "2026-09-02"}}, "verified_at": "2026-09-02"}, {"id": "FCA-CONC-7-3-4", "kind": "policy", "rule": "CONC 7.3.4", "claim": "An FCA rule requires a firm to treat customers in or approaching arrears or in default with forbearance and due consideration.", "writer_line": "the forbearance duty (CONC 7.3.4)", "source": {"citation": {"url": "https://handbook.fca.org.uk/handbook/CONC/7/3.html", "quote": "A firm must treat customers in or approaching arrears or in default with forbearance and due consideration", "title": "FCA Handbook - CONC 7.3 Treatment of customers in or approaching arrears or in default (including repossessions): lenders, owners and debt collectors", "publisher": "Financial Conduct Authority", "accessed": "2026-09-02"}}, "verified_at": "2026-09-02"}, {"id": "FCA-DISP-2-8-2", "kind": "policy", "rule": "DISP 2.8.2", "claim": "An FCA rule prevents the Financial Ombudsman Service from considering a complaint referred more than six months after the respondent sent its final response.", "writer_line": "the six-month FOS referral deadline (DISP 2.8.2)", "source": {"citation": {"url": "https://handbook.fca.org.uk/handbook/DISP/2/8.html", "quote": "more than six months after the date on which the respondent sent the complainant its final response", "title": "FCA Handbook - DISP 2.8 Was the complaint referred to the Financial Ombudsman Service in time?", "publisher": "Financial Conduct Authority", "accessed": "2026-09-02"}}, "verified_at": "2026-09-02"}], "banned_claims": []}'::jsonb,
  'manual',
  NULL,
  'lendzy_co_uk lane (migration 695)',
  true,
  true,
  'Eight FCA Handbook rule citations, each quote verified through datahelpers.QuoteFoundInText via cmd/fcaquotecheck on 2026-09-02 with a negative control in the same run. Two facts carry corrects_site_citation: the served pages attribute the rollover limit to CONC 6.7.17 (it is 6.7.23) and the CPA limit to CONC 6.7.23 (it is 7.6.12). Copy not touched - owner decision.'
);

-- Durable record of the two corrections OUTSIDE the jsonb round-trip surface
-- (round-2 (a)): a future struct-based writer can lose fields inside data; it
-- cannot touch this row.
INSERT INTO doc_notes (subject_type, subject_key, site_id, body, categories, source, created_by)
VALUES (
  'site', 'lendzy.co.uk', '8ff093d5-1f19-453b-9439-a10379bbcd76',
  'CITATION CORRECTIONS OF RECORD (2026-09-02, migration 695; served-copy fix is migration 696). '
  || 'The two-rollover limit for high-cost short-term credit is CONC 6.7.23 ("A firm must not refinance '
  || 'high-cost short-term credit (other than by exercising forbearance) on more than two occasions") - '
  || 'the site had cited CONC 6.7.17, the definitions rule. The two-attempt continuous payment authority '
  || 'limit is CONC 7.6.12 - the site had cited CONC 6.7.23, the refinance rule. Both substantive claims '
  || 'were always true; the attributions were shifted. Verified against the rule text at '
  || 'handbook.fca.org.uk on 2026-09-02; quotes verified through datahelpers.QuoteFoundInText with '
  || 'negative controls. The evidence register (site_specs aspect evidence_base, migration 695) carries '
  || 'the same corrections as facts FCA-CONC-6-7-23 and FCA-CONC-7-6-12 with corrects_site_citation keys.',
  '["lendzy","citation-correction","evidence-base"]'::jsonb,
  'lendzy_co_uk lane', 'lendzy_co_uk lane (migration 695)'
);

-- VERIFY as DO/RAISE. A verify block of bare SELECTs cannot stop the COMMIT.
DO $$
DECLARE nfacts int; ncited int; nrule int; ncorr int;
BEGIN
  SELECT jsonb_array_length(data->'facts') INTO nfacts FROM site_specs
   WHERE site_id = '8ff093d5-1f19-453b-9439-a10379bbcd76' AND aspect = 'evidence_base' AND is_current;
  IF nfacts <> 8 THEN
    RAISE EXCEPTION '695 VERIFY: expected 8 facts, found %', nfacts;
  END IF;

  SELECT count(*) INTO ncited FROM site_specs s, jsonb_array_elements(s.data->'facts') f
   WHERE s.site_id = '8ff093d5-1f19-453b-9439-a10379bbcd76' AND s.aspect = 'evidence_base' AND s.is_current
     AND f->'source'->'citation'->>'url' LIKE 'https://handbook.fca.org.uk/%'
     AND length(f->'source'->'citation'->>'quote') > 20;
  IF ncited <> 8 THEN
    RAISE EXCEPTION '695 VERIFY: expected 8 facts with an FCA handbook URL and a substantive quote, found %', ncited;
  END IF;

  SELECT count(*) INTO nrule FROM site_specs s, jsonb_array_elements(s.data->'facts') f
   WHERE s.site_id = '8ff093d5-1f19-453b-9439-a10379bbcd76' AND s.aspect = 'evidence_base' AND s.is_current
     AND f->>'rule' ~ '^(CONC|DISP) ';
  IF nrule <> 8 THEN
    RAISE EXCEPTION '695 VERIFY: expected 8 facts naming a CONC/DISP rule, found %', nrule;
  END IF;

  SELECT count(*) INTO ncorr FROM site_specs s, jsonb_array_elements(s.data->'facts') f
   WHERE s.site_id = '8ff093d5-1f19-453b-9439-a10379bbcd76' AND s.aspect = 'evidence_base' AND s.is_current
     AND f ? 'corrects_site_citation';
  IF ncorr <> 2 THEN
    RAISE EXCEPTION '695 VERIFY: expected 2 facts recording a corrected site citation, found %', ncorr;
  END IF;

  SELECT count(*) INTO ncorr FROM doc_notes
   WHERE site_id = '8ff093d5-1f19-453b-9439-a10379bbcd76'
     AND created_by = 'lendzy_co_uk lane (migration 695)'
     AND categories ? 'citation-correction';
  IF ncorr <> 1 THEN
    RAISE EXCEPTION '695 VERIFY: durable doc_notes correction record not found (%)', ncorr;
  END IF;

  RAISE NOTICE '695 OK: lendzy evidence register created - 8 FCA citations, 2 correcting a live misattribution, doc_notes record written';
END $$;

COMMIT;
