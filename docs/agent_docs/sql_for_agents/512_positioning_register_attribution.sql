-- 512 — record HOW a domain came to be attributed to its register entry.
--
-- WHY, and it is a trust distinction rather than a tidy-up. Migration 511's
-- loader takes domains from the labelled fields (`primary:`, `twins:`,
-- `domains:`). That gave 76 rows — and left **82 of the 152 portfolio domains
-- with no row at all**, because they are named in an entry's PROSE instead
-- (measured 2026-08-20: all 82 appear inside a `###` entry, none outside one, and
-- none absent from the document).
--
-- Sweeping the prose recovers all 82, which matters now that the database is the
-- source of truth — a domain with no row is invisible to anything that asks the
-- table. But a domain named in `primary:` and a domain mentioned in a sentence
-- are not equally certain attributions, and collapsing that difference would make
-- the table look more confident than it is. So it is recorded:
--
--   attribution = 'field'  — from primary:/twins:/domains:. High confidence.
--   attribution = 'prose'  — swept from the entry body. The entry is almost
--                            certainly right (it is the entry the domain is
--                            discussed in) but no human labelled it, so a
--                            consumer that needs certainty should read raw_md.
--   attribution = 'exclusion-only' — no register entry at all; the row exists
--                            solely to carry exclude_from_build.
--
-- Rollback: 512_positioning_register_attribution_ROLLBACK.sql

BEGIN;

ALTER TABLE positioning_register
  ADD COLUMN IF NOT EXISTS attribution text;

COMMENT ON COLUMN positioning_register.attribution IS
'How this domain was attributed to its entry: field (labelled primary:/twins:/domains: — high confidence), prose (swept from the entry body — the entry is right but nobody labelled it, so read raw_md if certainty matters), or exclusion-only (no entry; the row exists to carry exclude_from_build).';

CREATE INDEX IF NOT EXISTS idx_positioning_register_attribution
  ON positioning_register (attribution) WHERE attribution IS NOT NULL;

DO $$
DECLARE n int;
BEGIN
  SELECT count(*) INTO n FROM information_schema.columns
   WHERE table_name='positioning_register' AND column_name='attribution';
  IF n <> 1 THEN RAISE EXCEPTION 'attribution column not added'; END IF;
  RAISE NOTICE '512 OK — attribution column added';
END $$;

COMMIT;
