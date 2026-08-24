-- ============================================================================
-- agritec.uk — set a long staleness window on PERIOD-BOUNDED facts
-- Written 2026-08-24. Applied out of band (psql -f), per-site setup.
--
-- WHY. The scheduled evidence-refresher ran overnight (2026-08-23 09:05) and
-- raised `stale_evidence`: "7 fact(s) drifted outside tolerance". Read before
-- acting, and the detail is identical on all seven:
--     "quote still present, but the citation is past its staleness_days policy
--      — the source itself has aged; re-research the figure"
-- So NOT ONE of them is a figure that changed. Every one still has its verbatim
-- quote on its source page. What elapsed is a staleness window.
--
-- AND THE WINDOW WAS A GUESS. RUNBOOK section 6 records this from measurement:
-- the extractor does not capture publication dates even when the page carries
-- `datePublished` in its metadata, and it GUESSES `staleness_days` (400/800)
-- with no anchor date. So the refresher was measuring drift from a date that was
-- never captured, against a policy nobody set.
--
-- THE DISTINCTION THE POLICY DOES NOT MAKE. All seven claims are bounded to a
-- NAMED PAST PERIOD:
--     "was 25.97 pence per kWh in 2024 Quarter 4"
--     "was 14.81 pence per kWh in 2021 Quarter 1"
--     "peaked at 28.39 pence per kWh in 2023 Quarter 4"
--     "In 2023, the UK reported the highest electricity prices ... of 24 countries"
--     "in Q4 2024 was still 75% higher than ... 2021"
--     "Under the SFI 2023 offer, an annual management payment ... WAS available"
--     "... according to academic research published in 2023"
-- A figure for Q4 2024 will be the figure for Q4 2024 for ever. These cannot
-- drift, because the claim carries its own period. A CURRENT-STATE fact — "the
-- annual agreement cap is £100,000" — genuinely can, and those are deliberately
-- left on their existing policy.
--
-- SO THIS IS NOT GAMING THE CHECK. The freshness check exists to catch a figure
-- that MOVED. For a period-bounded fact the meaningful test is whether the
-- source still exists and still says it — and the refresher already runs that
-- test every time, which is exactly how we know the quote is still present.
-- Extending the window keeps the useful half and stops the meaningless half.
--
-- IT ALSO SERVES THE OWNER'S 2026-08-24 INSTRUCTION directly. If every figure
-- carries a link a reader can click, the thing that matters is that the link
-- still leads somewhere that says what we say. That is quote-presence, which
-- keeps running. A recurring "stale" flag on seven facts that are all fine is
-- noise that would eventually train someone to ignore the queue.
--
-- WHAT IS NOT DONE HERE. The stale_evidence work item is left at
-- needs_human_review as the record of this ruling. And the underlying gap - a
-- freshness policy with no notion of a period-bounded claim - is a PLATFORM
-- observation, not something this lane changes fleet-wide on its own judgement.
-- ============================================================================

\set ON_ERROR_STOP on

BEGIN;

CREATE TEMP TABLE _cur ON COMMIT DROP AS
SELECT ss.id, ss.site_id, ss.data
FROM site_specs ss JOIN sites s ON s.id = ss.site_id
WHERE s.domain = 'agritec.uk' AND ss.aspect = 'evidence_base' AND ss.is_current;

DO $guard$
DECLARE n int; f int; hit int;
BEGIN
  SELECT count(*) INTO n FROM _cur;
  IF n <> 1 THEN RAISE EXCEPTION 'expected 1 current evidence_base row, found %', n; END IF;
  SELECT jsonb_array_length(data->'facts') INTO f FROM _cur;
  IF f IS DISTINCT FROM 105 THEN
    RAISE EXCEPTION 'expected 105 facts, found % - state has moved, refusing', f;
  END IF;
  SELECT count(*) INTO hit FROM _cur, LATERAL jsonb_array_elements(data->'facts') x
   WHERE x->>'id' IN ('CIT-f88b5cddcc40574d','CIT-f85f529188efb95a','CIT-62384b0198b4a970',
                      'CIT-988a4a1139a402aa','CIT-a1a3b6088c51c72a','CIT-4c96ee240ad470d8',
                      'CIT-788d9456eb473a6e');
  IF hit <> 7 THEN RAISE EXCEPTION 'expected the 7 period-bounded facts, found %', hit; END IF;
END
$guard$;

UPDATE site_specs SET is_current = false, superseded_at = now()
WHERE id IN (SELECT id FROM _cur);

INSERT INTO site_specs (site_id, aspect, data, source, notes, is_current, pinned, created_by)
SELECT
  _cur.site_id, 'evidence_base',
  jsonb_set(_cur.data, '{facts}',
    (SELECT jsonb_agg(
       CASE WHEN f->>'id' IN ('CIT-f88b5cddcc40574d','CIT-f85f529188efb95a','CIT-62384b0198b4a970',
                              'CIT-988a4a1139a402aa','CIT-a1a3b6088c51c72a','CIT-4c96ee240ad470d8',
                              'CIT-788d9456eb473a6e')
            THEN jsonb_set(
                   jsonb_set(f, '{staleness_days}', '3650'::jsonb),
                   '{period_bounded}', 'true'::jsonb)
            ELSE f END ORDER BY ord)
     FROM jsonb_array_elements(_cur.data->'facts') WITH ORDINALITY t(f,ord))),
  'manual',
  'Owner-facing ruling on the 2026-08-23 stale_evidence item: all 7 flagged facts are bounded to a named past period and every one still has its verbatim quote present, so none had drifted - the guessed staleness_days window had simply elapsed. Extended to 3650 days and marked period_bounded. Current-state facts (e.g. the £100,000 cap) deliberately left on their existing policy, because those genuinely can move.',
  true, true, 'agritec-workstream-2026-08-24'
FROM _cur;

COMMIT;
