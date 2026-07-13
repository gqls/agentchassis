-- W9 step 3 (read-only): assets schema (before any url-flip SQL) + the presigned inventory.
\d assets

SELECT asset_key, status,
       (url LIKE '%X-Amz-Expires%') AS presigned,
       created_at
FROM assets
WHERE site_id = (SELECT id FROM sites WHERE domain = 'idea.uk')
ORDER BY created_at;
