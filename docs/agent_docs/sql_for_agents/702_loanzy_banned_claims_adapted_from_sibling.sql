-- 702_loanzy_banned_claims_adapted_from_sibling.sql
--
-- Fills loanzy.uk's empty banned_claims (697 shipped `"banned_claims": []`)
-- with FIVE patterns adapted from adversecreditmortgage.co.uk's six — the
-- bugs_open/414 lane's ask (2026-09-02), verified before writing:
--   * all five ran against ALL 27 SERVING pages (3 of 30 active rows 404 —
--     separate finding, lane NOTES): ZERO hits; a planted-text positive
--     control fires 5/5. The set bans nothing the site says today.
--   * pattern 3 is NARROWED from the sibling's `\bno (credit )?checks?\b`:
--     loanzy's calculator page truthfully says "There's no credit check
--     involved" ABOUT ITS OWN TOOL — the broad form would flag the site's own
--     honest sentence for ever. The narrowed form bans the predatory PROMISE
--     (no-credit-check LOANS), not the phrase.
--   * the sibling's sixth pattern (any literal % APR/rate) is DELIBERATELY
--     OMITTED: loanzy's worked examples ("9.9% APR") are the site's core
--     teaching device, illustrative and disclaimed in copy — adopting the ban
--     would flag every example page. On the sibling, a literal rate is a
--     price fact/financial promotion; here it is arithmetic pedagogy.
-- Supersede-and-merge (never edit): facts carried forward UNCHANGED; the
-- unique partial index idx_site_specs_current is the race backstop (a
-- concurrent writer aborts this txn rather than yielding 0-or-2 current rows).
--
-- (Authored as 699; renumbered 702 — the loancalculator lane took 699 for
-- their own register in the same hour, the owner's instruction propagating.)
--
-- Lane: docs/agent_docs/docs024_key_docs_latest/loanzy_uk_example_site/
-- Rollback: 702_..._ROLLBACK.sql
\set ON_ERROR_STOP on
BEGIN;

DO $$
DECLARE n int; nb int;
BEGIN
  SELECT count(*) INTO n FROM site_specs
   WHERE site_id='55213ded-03ec-40f7-8fc1-169de05e05c8' AND aspect='evidence_base' AND is_current;
  IF n <> 1 THEN RAISE EXCEPTION '702 ABORT: expected exactly 1 current loanzy register, found %', n; END IF;
  SELECT jsonb_array_length(COALESCE(data->'banned_claims','[]'::jsonb)) INTO nb FROM site_specs
   WHERE site_id='55213ded-03ec-40f7-8fc1-169de05e05c8' AND aspect='evidence_base' AND is_current;
  IF nb <> 0 THEN RAISE EXCEPTION '702 ABORT: banned_claims already has % entries - read before writing', nb; END IF;
END $$;

WITH cur AS (
  UPDATE site_specs SET is_current=false, superseded_at=now()
   WHERE site_id='55213ded-03ec-40f7-8fc1-169de05e05c8' AND aspect='evidence_base' AND is_current
   RETURNING data
)
INSERT INTO site_specs (site_id, aspect, data, source, source_agent, created_by, is_current, pinned, notes)
SELECT
  '55213ded-03ec-40f7-8fc1-169de05e05c8',
  'evidence_base',
  jsonb_set(cur.data, '{banned_claims}', '[
    {"pattern": "\\bguaranteed (acceptance|approval|loans?|yes)\\b", "reason": "No guaranteed-acceptance language, ever: unprovable, a financial-promotion exposure, and to this audience predatory (adapted from the adversecreditmortgage register M5 rule; mortgage -> loan)."},
    {"pattern": "\\b(everyone|anyone|all applicants) (is|are|will be) (accepted|approved)\\b", "reason": "The same promise wearing a different grammar (sibling set, verbatim)."},
    {"pattern": "\\bno[- ]credit[- ]check (loans?|lending|borrowing)\\b|\\bloans? with no credit checks?\\b", "reason": "A no-credit-check loan is the marker of the unregulated end. NARROWED from the sibling: loanzy''s calculator truthfully says ''no credit check involved'' about its own tool - the ban targets the lending promise, not the phrase."},
    {"pattern": "\\bbad credit (is )?(no|not a) (problem|issue|barrier)\\b", "reason": "Dismisses the reader''s actual situation and implies an outcome the site cannot know (sibling set, verbatim)."},
    {"pattern": "\\b(we|our team) (can|will) (get|secure) you (a|the) (loan|deal|approval)\\b", "reason": "Loanzy is not a broker and arranges nothing; this would misrepresent what the site is (adapted: mortgage -> loan)."}
  ]'::jsonb),
  'manual', NULL, 'loanzy_uk_example_site lane (migration 702)', true, true,
  'banned_claims filled per the 414 lane''s ask: 5 of the sibling''s 6 patterns, one narrowed (no-credit-check: the broad form flags loanzy''s own honest calculator sentence), one omitted with reason (literal %APR: loanzy''s worked examples are pedagogy, not price promotion). Verified 0 hits across all 27 serving pages + 5/5 planted-text positive control, 2026-09-02. Facts carried forward unchanged.'
FROM cur;

DO $$
DECLARE nfacts int; nb int; nbad int; ncur int;
BEGIN
  SELECT count(*) INTO ncur FROM site_specs
   WHERE site_id='55213ded-03ec-40f7-8fc1-169de05e05c8' AND aspect='evidence_base' AND is_current;
  IF ncur <> 1 THEN RAISE EXCEPTION '702 VERIFY: expected exactly 1 current row, found %', ncur; END IF;
  SELECT jsonb_array_length(data->'facts'), jsonb_array_length(data->'banned_claims') INTO nfacts, nb
   FROM site_specs WHERE site_id='55213ded-03ec-40f7-8fc1-169de05e05c8' AND aspect='evidence_base' AND is_current;
  IF nfacts <> 3 THEN RAISE EXCEPTION '702 VERIFY: facts were LOST - expected 3, found %', nfacts; END IF;
  IF nb <> 5 THEN RAISE EXCEPTION '702 VERIFY: expected 5 banned_claims, found %', nb; END IF;
  SELECT count(*) INTO nbad FROM site_specs s, jsonb_array_elements(s.data->'banned_claims') b
   WHERE s.site_id='55213ded-03ec-40f7-8fc1-169de05e05c8' AND s.aspect='evidence_base' AND s.is_current
     AND (length(COALESCE(b->>'pattern','')) < 10 OR length(COALESCE(b->>'reason','')) < 20);
  IF nbad <> 0 THEN RAISE EXCEPTION '702 VERIFY: % banned_claims entries with missing/thin pattern or reason', nbad; END IF;
  RAISE NOTICE '702 OK: loanzy banned_claims = 5 (facts carried: 3)';
END $$;
COMMIT;
