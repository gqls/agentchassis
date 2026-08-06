-- 321: backfill assets.storage_path from the presigned url — fleet-wide mirror
-- of idea.uk's w9_04 (IMG-053), for the bugs_open/152 + /155 fix.
--
-- WHY: the fix (chassis image carrying storage.AssetSourceRef) resolves an
-- asset's SOURCE object from the row itself: storage_path first, url as the
-- legacy fallback. A presigned url still names the object (only the path is
-- read, never fetched), so this backfill is not required for correctness —
-- it exists so the remaining presigned-only rows (205 measured 2026-08-06)
-- become durable BEFORE anything flips or hand-repairs their url, which is
-- exactly how 49 rows got stranded (url = local web path, storage_path NULL,
-- source unrecoverable).
--
-- ORDERING: independent of the image roll — write-only, no reader depends on
-- it until AssetSourceRef ships, and the old readers ignore storage_path
-- entirely. Safe to apply before or after.
--
-- IDEMPOTENT: the WHERE clause only fills NULLs; a re-run touches 0 rows.

BEGIN;

UPDATE assets
SET storage_path     = split_part(url, '?', 1),
    storage_provider = COALESCE(storage_provider, 'backblaze')
WHERE storage_path IS NULL
  AND url LIKE 'https://%'
  AND url LIKE '%X-Amz-%';

-- Verify loudly: a SELECT cannot stop the COMMIT (ON_ERROR_STOP ignores a
-- non-empty result) — use DO/RAISE per the standing migration practice.
DO $$
DECLARE n integer;
BEGIN
  SELECT count(*) INTO n FROM assets
  WHERE storage_path IS NULL AND url LIKE '%X-Amz-%';
  IF n > 0 THEN
    RAISE EXCEPTION 'backfill incomplete: % presigned rows still lack storage_path', n;
  END IF;
END $$;

COMMIT;
