-- 352_pages_noindex_flag.sql
-- bugs_open/232: opt-in per-page search-engine exclusion.
-- Read by getPageInfo/assemblePage (rerender_single_page_action.go), the
-- live assembly path for page-rerender. NOT read by the separate
-- assemble_page/AssemblePageAction build path (multipage_actions.go) --
-- see the LANDMINES entry filed alongside this migration.
BEGIN;

ALTER TABLE pages ADD COLUMN noindex boolean NOT NULL DEFAULT false;

COMMENT ON COLUMN pages.noindex IS
  'Opt-in (owner ruling 2026-08-02: new authority on a shared seam ships as an opt-in field, default OFF). When true, the page-rerender assembly path injects <meta name="robots" content="noindex, nofollow"> into the page head. Default false = zero behaviour change for every other page. bugs_open/232.';

-- Flip on for the one page this bug names: vonc.com /tools/gauntlet/round.html
-- (gauntlet-round-record) -- published visitor UGC + an AI verdict, potentially
-- naming real people, that should not be search-discoverable while a shared
-- link keeps working exactly as before.
UPDATE pages SET noindex = true
WHERE id = '4629451e-e4f2-4fe2-b258-35107b5cb51e';

-- Guard: at this moment the column is fresh, so exactly one row must be true.
-- A plain SELECT cannot stop the COMMIT; DO/RAISE can (standing practice).
DO $$
DECLARE n int;
BEGIN
  SELECT count(*) INTO n FROM pages WHERE noindex;
  IF n <> 1 THEN
    RAISE EXCEPTION 'bugs_open/232: expected exactly 1 noindex row after flip, found %', n;
  END IF;
END $$;

COMMIT;
