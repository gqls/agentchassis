-- ============================================================================
-- agritec.uk — vet the three LED-efficacy facts: drop one, date-stamp two
-- Written 2026-08-22. Applied out of band (psql -f), per-site setup.
--
-- Phase 2 run 5 (correlation d7908777-7ab9-47bc-9636-b14c756733ae) returned
-- three photon-efficacy facts, all from one Greenhouse Product News column by
-- Runkle and Bugbee — who are, genuinely, the two most-cited academics in
-- horticultural lighting, so this is expert authorship in a trade venue rather
-- than trade press repeating a vendor.
--
-- THE PAGE IS DATED 2017-07-03. Its own metadata carries
-- datePublished":"2017-07-03T14:40:34+00:00" and the extractor recorded
-- published: (none) for all three facts, guessing staleness_days at 400/800
-- with no anchor date. So the register's staleness machinery was set up to
-- measure drift from a date it never captured, on a page nine years old.
--
-- DROPPED — CIT-4730bf4274192f38, "Many new LED fixtures now exceed 2.0 µmol/J".
--   True in 2017 and misleading in 2026: the word "now" is doing the damage. The
--   DLC has raised its photon-efficacy threshold twice since (V3.0 by 21%, V4.0
--   by a further 8.7%), and the retired agritec site's own table — unsourced,
--   but not invented from nothing — put premium fixtures at 3.2. A reader told
--   that good LEDs "now exceed 2.0" would under-specify their rig and
--   over-estimate their running cost. A time-sensitive claim with no date
--   attached is the exact shape this register exists to keep off the page.
--
-- KEPT, with the date moved into the writer_line where the writer will see it:
--   CIT-2311a1c6ba7236ff — HPS at ~0.9 (400-W SE, magnetic) and ~1.7 µmol/J
--     (1,000-W DE, electronic). HPS is mature technology; these do not move.
--   CIT-bd51f5d7f3cb7541 — theoretical maximum 4.6 to 5.1 µmol/J. A physics
--     limit set by photon energy and diode efficiency; it does not age at all.
--
-- WHY THE VALUE FIELDS ARE LEFT ALONE. Both kept facts have a `value` that is
-- lossy on its own — 1.7 is one of two figures in its claim, 5.1 is the top of a
-- range. That would matter if `value` were the only thing a writer saw. It is
-- not: each writer_line embeds {value} in a sentence that restores the context
-- ("...compared to ~0.9 µmol/J for a 400-W single-ended HPS", "is 4.6 to
-- {value} µmol/J"). The writer_line is the control, and it already holds.
--
-- CONSEQUENCE FOR THE CALCULATOR (T1). LED efficacy stays a USER INPUT with no
-- asserted default, which is what the retired tool already did — it is one of
-- its five numeric fields, read off the operator's own fixture datasheet. That
-- is decision D4 working as intended: no citable current figure exists in a
-- form this pipeline can reach, so the number is not published, it is asked for.
-- These three facts serve the EXPLAINER (G1, the physics of light), where a
-- dated, attributed comparison of HPS against the theoretical LED ceiling is
-- exactly the right teaching material.
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
  IF f IS DISTINCT FROM 19 THEN
    RAISE EXCEPTION 'expected 19 facts, found % - another session has written here, refusing', f;
  END IF;
  SELECT count(*) INTO hit FROM _cur, LATERAL jsonb_array_elements(data->'facts') x
   WHERE x->>'id' IN ('CIT-4730bf4274192f38','CIT-2311a1c6ba7236ff','CIT-bd51f5d7f3cb7541');
  IF hit <> 3 THEN RAISE EXCEPTION 'expected the 3 LED facts, found %', hit; END IF;
END
$guard$;

UPDATE site_specs SET is_current = false, superseded_at = now()
WHERE id IN (SELECT id FROM _cur);

INSERT INTO site_specs (site_id, aspect, data, source, notes, is_current, pinned, created_by)
SELECT
  _cur.site_id, 'evidence_base',
  jsonb_set(_cur.data, '{facts}',
    (SELECT jsonb_agg(
              CASE
                WHEN f->>'id' IN ('CIT-2311a1c6ba7236ff','CIT-bd51f5d7f3cb7541')
                THEN jsonb_set(
                       jsonb_set(f, '{published}', '"2017-07"'::jsonb),
                       '{writer_line}',
                       to_jsonb(replace(f->>'writer_line',
                                        'Greenhouse Product News)',
                                        'Greenhouse Product News, July 2017)')))
                ELSE f
              END ORDER BY ord)
       FROM jsonb_array_elements(_cur.data->'facts') WITH ORDINALITY t(f,ord)
      WHERE f->>'id' <> 'CIT-4730bf4274192f38')),
  'manual',
  'Phase 2 run 5 vetting: source page dated 2017-07-03 in its own metadata while the extractor recorded published:(none) and guessed staleness_days. Dropped the one time-sensitive claim ("many new fixtures NOW exceed 2.0 µmol/J" - true in 2017, misleading in 2026 after two DLC threshold rises). Kept the two durable ones (mature HPS technology; a physics ceiling) and moved the 2017 date into their writer_line so the writer states it. LED efficacy stays a USER INPUT on the calculator - no citable current figure exists in a form this pipeline can reach.',
  true, true, 'agritec-workstream-2026-08-22'
FROM _cur;

COMMIT;
