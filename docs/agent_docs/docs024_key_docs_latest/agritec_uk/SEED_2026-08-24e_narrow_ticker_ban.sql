-- ============================================================================
-- agritec.uk — my fabricated-ticker ban was blocking a CITATION. Narrow it.
-- Written 2026-08-24. Applied out of band (psql -f), per-site setup.
--
-- WHAT HAPPENED. Two of the six explainers would not build:
--   the-physics-of-horticultural-lighting  needs_rebuild, 0 components
--   stacking-agricultural-scheme-actions   needs_rebuild, 0 components
-- both failing `validate_content` with "1 blockers, 0 errors". The detail, from
-- agent_error_log CONTENT_VALIDATION_BLOCKER_DETAIL:
--
--   type: banned_claim
--   value: "Carbon Brief, May 2025"
--   location: "...the time, according to academic research published in 2023
--             (Carbon Brief, May 2025, citing Zakeri and Staffell 2023), which
--             is why the price a..."
--
-- So the writer did exactly what it was asked to do — attributed a registered
-- fact to its source, by name and date — and MY OWN BAN blocked the page for it.
-- The pattern was
--     (brent|crude|carbon|uk ets|day-ahead|ammonium nitrate)[^.]{0,30}[0-9]
-- and "CARBON Brief, May 2025" is `carbon` followed by a digit within 30 chars.
-- It was written to catch "UK Base Carbon: £45.00" off the retired ticker strip.
--
-- The irony is worth stating plainly: the rule I added to stop fabricated prices
-- was blocking the citation the owner asked for on 2026-08-24. A ban aimed at
-- invented figures caught a real source instead, and the only reason anyone
-- noticed is that it stopped a build rather than quietly suppressing a sentence.
--
-- WHY I DID NOT CATCH IT. I tested the SFI bans on BOTH arms — five must-ban
-- strings and four must-stay-sayable strings — and it caught two bad patterns.
-- For the ticker bans I tested only the positive arm ("UK Feed Wheat GBP
-- 189.50"). A ban tested only on what it should catch cannot tell you what else
-- it catches. Same lesson, same lane, four days apart, and this time it cost two
-- pages.
--
-- THE NEW PATTERN IS STRICTLY BETTER ON BOTH ARMS, tested before applying:
--   old: 3 false positives on legitimate prose, and it MISSED "AN Fertilizer:
--        £345/t" — which is on the retired homepage. Over-broad AND
--        under-inclusive at once.
--   new: 0 false positives, 0 misses on all five real ticker strings.
-- It works by requiring the multi-word instrument NAME as it actually appears on
-- the ticker, and a CURRENCY SYMBOL before the number. A price has a currency; a
-- publisher's name does not.
-- ============================================================================

\set ON_ERROR_STOP on

BEGIN;

CREATE TEMP TABLE _cur ON COMMIT DROP AS
SELECT ss.id, ss.site_id, ss.data
FROM site_specs ss JOIN sites s ON s.id = ss.site_id
WHERE s.domain = 'agritec.uk' AND ss.aspect = 'evidence_base' AND ss.is_current;

DO $guard$
DECLARE n int; hit int;
BEGIN
  SELECT count(*) INTO n FROM _cur;
  IF n <> 1 THEN RAISE EXCEPTION 'expected 1 current evidence_base row, found %', n; END IF;
  SELECT count(*) INTO hit FROM _cur, LATERAL jsonb_array_elements(data->'banned_claims') b
   WHERE b->>'pattern' = '(brent|crude|carbon|uk ets|day-ahead|ammonium nitrate)[^.]{0,30}[0-9]';
  IF hit <> 1 THEN RAISE EXCEPTION 'the over-broad ticker pattern is not present exactly once (found %)', hit; END IF;
END
$guard$;

UPDATE site_specs SET is_current = false, superseded_at = now()
WHERE id IN (SELECT id FROM _cur);

INSERT INTO site_specs (site_id, aspect, data, source, notes, is_current, pinned, created_by)
SELECT
  _cur.site_id, 'evidence_base',
  jsonb_set(_cur.data, '{banned_claims}',
    (SELECT jsonb_agg(
       CASE WHEN b->>'pattern' = '(brent|crude|carbon|uk ets|day-ahead|ammonium nitrate)[^.]{0,30}[0-9]'
            THEN jsonb_build_object(
                   'pattern','(brent crude|uk base carbon|carbon \(uk ets\)|day-ahead power|ammonium nitrate|an fertilizer)[^.]{0,25}[£$][0-9]',
                   'reason','fabricated-ticker class, NARROWED 2026-08-24. The original — (brent|crude|carbon|uk ets|day-ahead|ammonium nitrate)[^.]{0,30}[0-9] — blocked two pages from building because it matched "Carbon Brief, May 2025", a CITATION. It also missed "AN Fertilizer: £345/t", which is on the retired homepage: over-broad and under-inclusive at once. This version requires the instrument name as it appears on the ticker AND a currency symbol before the number, because a price has a currency and a publisher does not. Tested on both arms: 5/5 real ticker strings caught, 0 false positives.')
            ELSE b END ORDER BY ord)
     FROM jsonb_array_elements(_cur.data->'banned_claims') WITH ORDINALITY t(b,ord))),
  'manual',
  'Narrowed the fabricated-ticker ban after it blocked the-physics-of-horticultural-lighting and stacking-agricultural-scheme-actions at validate_content for matching "Carbon Brief, May 2025". Tested on both arms this time.',
  true, true, 'agritec-workstream-2026-08-24'
FROM _cur;

COMMIT;
