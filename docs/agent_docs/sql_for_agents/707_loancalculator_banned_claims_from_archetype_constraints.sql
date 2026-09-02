-- 707_loancalculator_banned_claims_from_archetype_constraints.sql
--
-- Populates loancalculator.co.uk's banned_claims — migration 699 created the
-- register with 12 facts but shipped `banned_claims: []`. This is the second
-- half of the register work, prompted by the bugfix_414 lane's relay
-- (2026-09-02): the site's `site_archetype.constraints` reads like a compliance
-- layer in PROSE ("Never make calculators appear to give regulated financial
-- advice", "Never add real lender recommendations...", "Never reposition the
-- site as a lender or broker...") and NOTHING ENFORCES IT — banned_claims
-- patterns ARE enforced, at the build gate, the persistence floor
-- (claimsGuardBeforePersist refuses the save) and the post-deploy sweep.
--
-- EIGHT patterns: five adapted from adversecreditmortgage.co.uk's audited set
-- (mortgage→loan), three translated from this site's own archetype
-- constraints. Go-regex, compiled case-insensitive by datahelpers/claims.go
-- (a pattern that fails Compile silently degrades to a QuoteMeta LITERAL —
-- all eight were compiled and probe-fired through Go's own engine before this
-- file was written, 8/8). Matches negated in the same clause are dropped
-- (negatedClaimMatch), but the census, not the negation guard, is the
-- load-bearing safety.
--
-- CENSUS BEFORE ARMING [MEASURED 2026-09-02]: every shipped pattern was run
-- against the FULL visible text of all 28 served pages: ZERO matches for all
-- eight, with planted-text positive controls firing 8/8 in the Go engine.
-- Arming them cannot refuse a save of any current page — including the
-- pending bugs_open/397 GTM rerender wave.
--
-- TWO LESSONS INHERITED FROM SIBLING MIGRATION 702 (loanzy), deliberately:
--   * the no-credit-check pattern is the NARROWED form. loanzy's calculator
--     truthfully says "no credit check involved" about its own tool, and this
--     site's credit-health-check page makes the same honest promise in other
--     words today — a future rewrite could use the literal phrase and be
--     refused for ever by the broad form. The ban targets the lending PROMISE
--     (no-credit-check loans), not the phrase. Re-censused in its narrowed
--     form (the hyphenated alternative is NOT a subset of the broad form):
--     0 matches, control fires.
--   * supersede-and-merge, NOT an in-place jsonb_set UPDATE. The daily
--     refresher's write-back is a CAS keyed on the row id it read
--     (writeRefreshedEvidenceBase): an in-place edit keeps the id, so a
--     refresher that read before this migration and wrote after would CAS-
--     SUCCEED and silently restore banned_claims: [] — a lost update. The
--     supersede changes the current row, the refresher affects 0 rows and
--     skips, and idx_site_specs_current is the race backstop for any
--     concurrent writer.
--
-- ONE PATTERN FROM THE PRECEDENT SET IS DELIBERATELY EXCLUDED, with the
-- measurement: the literal-rate pattern (\b[0-9]+(\.[0-9]+)?% (apr|apcr|rate)\b)
-- matches TWICE on this site's live copy — tools/compare-loans.html's
-- deliberately illustrative "a 7.9% APR loan and an 8.4% APR loan can be
-- compared like for like". On a broker-shaped site a literal rate is a stale
-- price fact and a promotion exposure; on an educational calculator site
-- illustrative APRs ARE the content. Adopting it would refuse saves of a page
-- doing its job — RFC_060 §1c's false-positive class, measured here rather
-- than repeated. (Same call, same reason, as loanzy's 702.)
--
-- NOTE ON created_by: this supersede writes a new current row with THIS
-- migration's created_by, so 699's ROLLBACK guard (keyed on 699's created_by)
-- refuses from now on — correct per its own header ("the register has moved
-- on"). 707's rollback keys on the shape (12 facts + 8 patterns), not the
-- author.
--
-- Lane: docs/agent_docs/docs024_key_docs_latest/loancalculator_couk/
-- Rollback: 707_..._ROLLBACK.sql

\set ON_ERROR_STOP on
BEGIN;

DO $$
DECLARE n int; nb int; nfacts int;
BEGIN
  SELECT count(*) INTO n FROM site_specs
   WHERE site_id='0162cde4-633e-45e9-8ca6-87a6b2fe1d26' AND aspect='evidence_base' AND is_current;
  IF n <> 1 THEN RAISE EXCEPTION '707 ABORT: expected exactly 1 current loancalculator register, found %', n; END IF;
  SELECT jsonb_array_length(COALESCE(data->'banned_claims','[]'::jsonb)),
         jsonb_array_length(data->'facts') INTO nb, nfacts FROM site_specs
   WHERE site_id='0162cde4-633e-45e9-8ca6-87a6b2fe1d26' AND aspect='evidence_base' AND is_current;
  IF nb <> 0 THEN RAISE EXCEPTION '707 ABORT: banned_claims already has % entries - read before writing', nb; END IF;
  IF nfacts <> 12 THEN RAISE EXCEPTION '707 ABORT: expected the 12 facts 699 wrote, found % - the register has changed shape, read before writing', nfacts; END IF;
END $$;

WITH cur AS (
  UPDATE site_specs SET is_current=false, superseded_at=now()
   WHERE site_id='0162cde4-633e-45e9-8ca6-87a6b2fe1d26' AND aspect='evidence_base' AND is_current
   RETURNING data, pinned
)
INSERT INTO site_specs (site_id, aspect, data, source, source_agent, created_by, is_current, pinned, notes)
SELECT
  '0162cde4-633e-45e9-8ca6-87a6b2fe1d26',
  'evidence_base',
  jsonb_set(cur.data, '{banned_claims}', '[
    {"pattern": "\\bguaranteed (acceptance|approval|loans?|yes)\\b",
     "reason": "Guaranteed-acceptance language is unprovable, a financial-promotion exposure, and predatory to this audience (adapted from the adversecreditmortgage audited set; mortgage -> loan). Zero matches on live copy, census 2026-09-02."},
    {"pattern": "\\b(everyone|anyone|all applicants) (is|are|will be) (accepted|approved)\\b",
     "reason": "The same promise wearing a different grammar (sibling set, verbatim). Zero matches on live copy, census 2026-09-02."},
    {"pattern": "\\bno[- ]credit[- ]check (loans?|lending|borrowing)\\b|\\bloans? with no credit checks?\\b",
     "reason": "A no-credit-check loan is the marker of the unregulated end. NARROWED per sibling migration 702: a calculator site truthfully describes its own tools as involving no credit check, so the ban targets the lending PROMISE, not the phrase. Zero matches on live copy in this narrowed form, census 2026-09-02."},
    {"pattern": "\\bbad credit (is )?(no|not a) (problem|issue|barrier)\\b",
     "reason": "Dismisses the reader''s actual situation and implies an outcome the site cannot know (sibling set, verbatim). Zero matches on live copy, census 2026-09-02."},
    {"pattern": "\\b(we|our team) (can|will) (get|secure) you (a|the) (loan|deal|approval)\\b",
     "reason": "The site is not a broker and arranges nothing; this would misrepresent what it is. Archetype constraint: never reposition as a lender or broker. Zero matches on live copy, census 2026-09-02."},
    {"pattern": "\\bwe (recommend|advise) (that )?you (take out|borrow|apply)\\b",
     "reason": "Archetype constraint, translated: never make the calculators appear to give regulated financial advice. The site is not FCA-authorised to recommend a loan or course of action (its own legal.html says so). Zero matches on live copy, census 2026-09-02."},
    {"pattern": "\\b(we lend|borrow from us|apply (with|through) us|we are a (lender|broker)|we (can )?arrange (a |your )?loan)\\b",
     "reason": "Archetype constraint, translated: never reposition the site as a lender or broker without appropriate regulatory framing. Zero matches on live copy, census 2026-09-02."},
    {"pattern": "\\b(best|top) (uk )?(loan|lender|deal)s?\\b",
     "reason": "Archetype constraint, translated: never add real lender recommendations or ranked product tables without FCA authorisation disclaimers - a superlative recommendation is the one-line form of a ranked table. Zero matches on live copy (the guides'' ''best single figure'' phrasing does not match), census 2026-09-02."}
  ]'::jsonb),
  'manual', NULL, 'loancalculator_couk lane (migration 707)', true, cur.pinned,
  'banned_claims filled per the 414 lane''s ask: 5 sibling patterns (one narrowed per 702''s no-credit-check lesson), 3 translated from this site''s own archetype constraints, literal-%APR omitted on a 2-match census (illustrative APRs are the site''s pedagogy). All 8 compiled+probe-fired in Go 8/8; 0 matches across all 28 served pages, census 2026-09-02. Facts carried forward unchanged from 699. Supersede-and-merge per 702 (the refresher CAS makes in-place edits lose).'
FROM cur;

DO $$
DECLARE nfacts int; nb int; nbad int; ncur int;
BEGIN
  SELECT count(*) INTO ncur FROM site_specs
   WHERE site_id='0162cde4-633e-45e9-8ca6-87a6b2fe1d26' AND aspect='evidence_base' AND is_current;
  IF ncur <> 1 THEN RAISE EXCEPTION '707 VERIFY: expected exactly 1 current row, found %', ncur; END IF;
  SELECT jsonb_array_length(data->'facts'), jsonb_array_length(data->'banned_claims') INTO nfacts, nb
   FROM site_specs WHERE site_id='0162cde4-633e-45e9-8ca6-87a6b2fe1d26' AND aspect='evidence_base' AND is_current;
  IF nfacts <> 12 THEN RAISE EXCEPTION '707 VERIFY: facts were LOST - expected 12, found %', nfacts; END IF;
  IF nb <> 8 THEN RAISE EXCEPTION '707 VERIFY: expected 8 banned_claims, found %', nb; END IF;
  SELECT count(*) INTO nbad FROM site_specs s, jsonb_array_elements(s.data->'banned_claims') b
   WHERE s.site_id='0162cde4-633e-45e9-8ca6-87a6b2fe1d26' AND s.aspect='evidence_base' AND s.is_current
     AND (length(COALESCE(b->>'pattern','')) < 10 OR length(COALESCE(b->>'reason','')) < 20);
  IF nbad <> 0 THEN RAISE EXCEPTION '707 VERIFY: % banned_claims entries with missing/thin pattern or reason', nbad; END IF;

  SELECT count(*) INTO nbad FROM site_specs s, jsonb_array_elements(s.data->'facts') f
   WHERE s.site_id='0162cde4-633e-45e9-8ca6-87a6b2fe1d26' AND s.aspect='evidence_base' AND s.is_current
     AND f ? 'corrects_site_citation';
  IF nbad <> 2 THEN RAISE EXCEPTION '707 VERIFY: the 2 corrects_site_citation fields must survive the merge, found %', nbad; END IF;

  RAISE NOTICE '707 OK: loancalculator banned_claims = 8 (facts carried: 12, corrections carried: 2); literal-APR excluded on a 2-match census';
END $$;
COMMIT;
