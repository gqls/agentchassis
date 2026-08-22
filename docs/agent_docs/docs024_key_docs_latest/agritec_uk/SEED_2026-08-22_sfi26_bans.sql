-- ============================================================================
-- agritec.uk — add SFI26 banned_claims to the evidence register
-- Written 2026-08-22. Applied out of band (psql -f), per-site setup.
--
-- WHY. Phase 2's first evidence run registered 10 verified facts about the
-- Sustainable Farming Incentive, and four of them say the same thing:
--
--   "the SFI management payment has been removed for SFI26 agreements"
--     — gov.uk SFI26 scheme rules and guidance
--   "You will not be paid: an SFI management payment for your SFI26 agreement"
--     — same page
--   "We will no longer offer the SFI management payment. It was intended as a
--    time-limited payment to support farmers transitioning into the new scheme."
--     — defrafarming.blog.gov.uk, 2026-02-24
--
-- Verified first-hand 2026-08-22 by fetching the GOV.UK page (HTTP 200) and
-- confirming each quote appears verbatim, because a machine-verified citation
-- proves the quote is on the page and NOT that we read it correctly
-- (bugs_open/161).
--
-- THE SITE BEING REPLACED STILL PAYS IT. agritec.uk/tools/elms-calculator.html
-- leads with a callout — "You receive £20 per hectare for the first 50 hectares
-- ... the first £1,000 of your SFI income is effectively guaranteed" — and its
-- JS computes Math.min(farmSize, 50) * 20 into a headline "SFI Management Pmt"
-- line. That is money a farmer will not receive under SFI26, presented as
-- guaranteed, next to a GOV.UK link that lends it authority. It is exactly the
-- bugs_open/288 class: a legislated figure encoded in a calculator, checked by
-- nothing, still running after the legislation moved.
--
-- WHAT THESE BANS DO AND DO NOT DO. They forbid asserting the management
-- payment as AVAILABLE. They deliberately do NOT ban the words themselves: the
-- explainer must be able to say it was removed, and why, and what it used to be
-- — that history is now one of our registered facts (CIT-f88b5cd..., correctly
-- scoped to "Under the SFI 2023 offer"). A ban that made the removal
-- unsayable would suppress the most useful thing this site can currently tell
-- an SFI reader.
--
-- The register's own trap applies to the historical fact: it is TRUE and it is
-- STALE-BY-DESIGN, and a writer handed it may state it in the present tense.
-- The last ban below is aimed at that specific failure.
-- ============================================================================

\set ON_ERROR_STOP on

BEGIN;

-- Supersede-then-insert, carrying the whole existing document forward and
-- touching ONLY banned_claims. Never UPDATE a spec row in place.
--
-- ⚠ THIS CANNOT BE ONE STATEMENT WITH A DATA-MODIFYING CTE, and the version that
--   tried it failed here on 2026-08-22 with:
--     duplicate key value violates unique constraint "idx_site_specs_current"
--   All CTEs in a statement run against ONE snapshot, so the INSERT's uniqueness
--   check does not see the sibling UPDATE's supersede — the old row is still
--   is_current as far as the partial index is concerned. It reads perfectly and
--   is simply wrong. Sequential statements in one transaction is the fix, and it
--   is what the oufe seed does. The failure was clean (BEGIN + ON_ERROR_STOP
--   rolled it all back, verified: bans/facts/writer_block unchanged).

CREATE TEMP TABLE _cur ON COMMIT DROP AS
SELECT ss.id, ss.site_id, ss.data
FROM site_specs ss JOIN sites s ON s.id = ss.site_id
WHERE s.domain = 'agritec.uk' AND ss.aspect = 'evidence_base' AND ss.is_current;

-- Abort loudly if the world is not what we think. A verify block of plain
-- SELECTs CANNOT stop a COMMIT (ON_ERROR_STOP ignores a non-empty result), so
-- this has to be DO/RAISE.
DO $guard$
DECLARE n int; f int;
BEGIN
  SELECT count(*) INTO n FROM _cur;
  IF n <> 1 THEN
    RAISE EXCEPTION 'expected exactly 1 current evidence_base row, found %', n;
  END IF;
  SELECT jsonb_array_length(data->'facts') INTO f FROM _cur;
  IF f IS DISTINCT FROM 10 THEN
    RAISE EXCEPTION 'expected the 10 registered SFI facts, found % - refusing to rewrite a document I do not recognise', f;
  END IF;
END
$guard$;

UPDATE site_specs SET is_current = false, superseded_at = now()
WHERE id IN (SELECT id FROM _cur);

INSERT INTO site_specs (site_id, aspect, data, source, notes, is_current, pinned, created_by)
SELECT
  _cur.site_id,
  'evidence_base',
  jsonb_set(_cur.data, '{banned_claims}',
    (_cur.data->'banned_claims') || $add$[

      {"pattern": "management payment[^.]{0,80}(is|are|will be) (available|offered|paid|payable)",
       "reason": "REMOVED FOR SFI26. gov.uk SFI26 scheme rules: 'the SFI management payment has been removed for SFI26 agreements' and 'You will not be paid: an SFI management payment for your SFI26 agreement'. Verified first-hand 2026-08-22. The retired hand-built calculator paid it as a headline line item; it must never be asserted as current again."},
      {"pattern": "(you|farmers?) (will|would|can) (receive|get|claim)[^.]{0,60}management payment",
       "reason": "Same removal, stated as an entitlement to the reader — which is also the subsidy-entitlement class this site never asserts."},
      {"pattern": "((effectively )?guaranteed[^.]{0,80}(£ ?1,?000|first 50)|(£ ?1,?000|first 50)[^.]{0,80}(effectively )?guaranteed)",
       "reason": "The retired calculator's exact framing: 'the first £1,000 of your SFI income is effectively guaranteed'. Guaranteed was never true and the payment no longer exists."},
      {"pattern": "£ ?20 ?(per|/) ?(ha|hectare)[^.]{0,60}(first )?50 ?(ha|hectares)",
       "reason": "The removed rate and cap stated together as a live offer. The HISTORICAL fact is registered (CIT-f88b5cd..., 'Under the SFI 2023 offer') and remains sayable in the past tense with its scope attached; this bans the bare present-tense form."},
      {"pattern": "management payment (is|remains|stands at|comes to) £",
       "reason": "Any present-tense monetary statement of the removed payment. PRESENT TENSE ONLY, deliberately: 'a management payment of £20 per hectare WAS available under the SFI 2023 offer' must stay sayable, because that scoped past-tense form is the honest way to use registered fact CIT-f88b5cd. Aimed at the register's own stale-fact risk (bugs_open/161) — the fact is true, dated and scoped, and a writer handed it may still state it in the present tense."}
        ]$add$::jsonb),
  'manual',
  'Phase 2, run 1 (correlation 1e8c7735-b922-450c-b261-cbfac3e2d5d6): 10 facts registered, 4 of them establishing that the SFI management payment is removed for SFI26. Bans added BEFORE any page exists so the first page is covered. Facts, writer_block, allowed_entities and governing_rule carried forward unchanged by jsonb_set on banned_claims alone.',
  true, true, 'agritec-workstream-2026-08-22'
FROM _cur;

COMMIT;

-- Verify: bans should rise 19 -> 24, and facts must STILL be 10 (a lost facts
-- array would mean the carry-forward failed, and it would look like success).
--   SELECT jsonb_array_length(data->'banned_claims') AS bans,
--          jsonb_array_length(data->'facts')         AS facts,
--          length(data->>'writer_block')             AS writer_block_chars
--     FROM site_specs ss JOIN sites s ON s.id=ss.site_id
--    WHERE s.domain='agritec.uk' AND ss.aspect='evidence_base' AND ss.is_current;
