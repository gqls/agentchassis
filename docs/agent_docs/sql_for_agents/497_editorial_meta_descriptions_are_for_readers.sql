-- 497_editorial_meta_descriptions_are_for_readers.sql
--
-- Replaces the meta_description on both editorial feature pages. The old ones
-- were HAND-AUTHORED BY THIS LANE in migrations 492 and 495 and read as
-- commissioning notes to a writer rather than descriptions for a reader.
--
-- The dartsonline one, 291 chars, verbatim:
--   "...Set against the calendar itself — 30 Players Championship events a
--    season through 2024, 34 since 2025 — these are one story about schedule
--    density, not four about discipline."
-- "these are one story about X, not four about Y" is the PREMISE TEST from this
-- lane's own design doc — an instruction about how to frame the piece — leaking
-- into the sentence a search engine prints under the title.
--
-- FOUND BY the meta_description_never_backfilled lane (bugs_open/320, 339),
-- which eliminated the site planner, the tool path, apply_gap_plan_action and
-- the rerender before handing it over. Their chain was exhaustive over CODE
-- PATHS and the answer was not on one: a session wrote the string into a seed.
--
-- ⚠ THE FINDING THAT MATTERS BEYOND THESE TWO PAGES, recorded here because the
-- fix itself is trivial: a seed migration writing pages.meta_description
-- directly bypasses EVERY producer-side control. 339's proposed remedy for the
-- tool path — "don't pass the brief as a candidate at all, compose the public
-- sentence separately" — cannot apply here, because there is no composer in the
-- path to fix. For a hand-authored row the guard (PublicMetaDescription) is the
-- ONLY control, and it is measured not to fire in the 200-320 band. So this
-- class needs either a guard that catches it or a check at seed time; nothing
-- upstream of the row exists to correct.
--
-- Replacements are reader-facing, both under 160 characters (search snippets
-- truncate around there; the originals were 242 and 291), and carry none of the
-- brief-shaped constructions.
--
-- After apply: assemble-only page-rerender on both pages so the served <head>
-- carries the new text. ROLLBACK: restore from the backup table below.

\set ON_ERROR_STOP on

CREATE TABLE IF NOT EXISTS pages_backup_20260821_meta_desc AS
SELECT id, name, site_id, meta_description FROM pages
 WHERE name IN ('robot-demand-step-change','darts-calendar-density');

BEGIN;

DO $$
DECLARE n int;
BEGIN
  SELECT count(*) INTO n FROM pages
   WHERE name='darts-calendar-density' AND meta_description LIKE '%not four about discipline%';
  IF n <> 1 THEN RAISE EXCEPTION 'dartsonline brief-shaped description not found (already fixed?)'; END IF;
END $$;

UPDATE pages SET meta_description =
  'Top players are skipping tournaments and the PDC has noticed. But the calendar grew: 30 Players Championship events a season through 2024, 34 since 2025.',
  updated_at = now()
 WHERE name='darts-calendar-density'
   AND site_id='5fe8785b-223d-41a3-88ee-c07187622381';

UPDATE pages SET meta_description =
  'Record robot installations, rising orders, one headline calling US demand weak. Five years of IFR figures show a step up in 2021, then a plateau at altitude.',
  updated_at = now()
 WHERE name='robot-demand-step-change'
   AND site_id='00ff3af5-dad8-4770-9f70-3edc267a3c92';

-- Verify: both under 160, neither carrying a brief-shaped construction.
DO $$
DECLARE bad int;
BEGIN
  SELECT count(*) INTO bad FROM pages
   WHERE name IN ('robot-demand-step-change','darts-calendar-density')
     AND (length(meta_description) > 160
       OR meta_description ILIKE '%not four about%'
       OR meta_description ILIKE '%these are one story%'
       OR meta_description ILIKE '%an editorial feature%');
  IF bad > 0 THEN RAISE EXCEPTION '% description(s) still too long or brief-shaped', bad; END IF;
  IF (SELECT count(*) FROM pages
       WHERE name IN ('robot-demand-step-change','darts-calendar-density')
         AND length(meta_description) BETWEEN 100 AND 160) <> 2 THEN
    RAISE EXCEPTION 'expected 2 descriptions in the 100-160 band';
  END IF;
END $$;

COMMIT;
