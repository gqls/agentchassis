-- W9 step 4: flip idea.uk's presigned asset URLs to the repo-local paths the deployer
-- already committed. Naming convention verified against the repo listing:
-- /assets/images/<asset_key with _ -> ->.jpg (optimiser normalises extensions to .jpg).
-- The unsigned S3 object path is PRESERVED into storage_path (COALESCE = only if empty),
-- provider recorded; trg_assets_updated_at bumps updated_at. Guarded to idea.uk +
-- presigned rows only => idempotent (post-flip urls no longer match the guard).
UPDATE assets
SET storage_path     = COALESCE(storage_path, split_part(url, '?', 1)),
    storage_provider = COALESCE(storage_provider, 'backblaze'),
    filename         = replace(asset_key, '_', '-') || '.jpg',
    url              = '/assets/images/' || replace(asset_key, '_', '-') || '.jpg'
WHERE site_id = (SELECT id FROM sites WHERE domain = 'idea.uk')
  AND url LIKE '%X-Amz-Expires%'
  AND asset_key IS NOT NULL
RETURNING asset_key, url, filename, left(storage_path, 78) AS s3_object_preserved;
-- Expect: UPDATE 18, every url now /assets/images/<key-hyphenated>.jpg.

-- Verify:
SELECT count(*) FILTER (WHERE url LIKE '%X-Amz-Expires%') AS still_presigned,   -- expect 0
       count(*) FILTER (WHERE url LIKE '/assets/images/%') AS local_urls        -- expect 18
FROM assets
WHERE site_id = (SELECT id FROM sites WHERE domain = 'idea.uk') AND asset_key IS NOT NULL;
