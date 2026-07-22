-- 055_seed_allowlist_ROLLBACK.sql — remove the allowlist key, restoring the
-- default contamination behaviour for fundamentallyai.com. Idempotent (the `-`
-- operator on an absent key is a no-op). Also the way to re-seed a changed list:
-- run this, then re-run 055_seed_allowlist.sql.
UPDATE sites
SET content_data = content_data - 'allowed_reference_domains'
WHERE domain = 'fundamentallyai.com'
RETURNING domain,
          content_data ? 'allowed_reference_domains' AS key_present_should_be_false;
