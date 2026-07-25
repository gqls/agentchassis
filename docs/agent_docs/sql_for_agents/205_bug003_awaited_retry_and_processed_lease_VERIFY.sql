-- 205_bug003_awaited_retry_and_processed_lease_VERIFY.sql
--
-- Read-only. Safe to run before OR after 205. Check every row by eye — a
-- migration that reports success is not the same as a platform that retries
-- (bugs_open/003's whole history is status rows that lied).

\echo '=== 1. Is 205 recorded in the ledger? (bugs_open/007 trap) ==='
SELECT filename, applied_at FROM schema_migrations WHERE filename LIKE '205_%';

\echo ''
\echo '=== 2. awaited_requests CHECK admits retrying (must contain the word) ==='
SELECT pg_get_constraintdef(oid)
FROM pg_constraint
WHERE conname = 'awaited_requests_status_check';

\echo ''
\echo '=== 3. processed_messages columns (status default complete, lease nullable) ==='
SELECT column_name, data_type, column_default, is_nullable
FROM information_schema.columns
WHERE table_name = 'processed_messages'
  AND column_name IN ('status','lease_expires_at');

\echo ''
\echo '=== 4. Pre-roll every row must be complete; post-roll processing is transient ==='
SELECT status, count(*) FROM processed_messages GROUP BY status;

\echo ''
\echo '=== 5. Both partial indexes present ==='
SELECT indexname FROM pg_indexes
WHERE indexname IN ('idx_awaited_requests_retrying','idx_processed_messages_processing');

\echo ''
\echo '=== 6. Cleanup function carries the retrying sweep (grep the source) ==='
SELECT position('retrying' in prosrc) > 0 AS function_has_retrying_sweep
FROM pg_proc WHERE proname = 'cleanup_expired_awaited_requests';

\echo ''
\echo '=== 7. awaited_requests status distribution (retrying must stay transient) ==='
SELECT status, count(*) FROM awaited_requests GROUP BY status ORDER BY 2 DESC;

\echo ''
\echo '=== 8. F4 config landed (600 / 900) ==='
SELECT type, idle_timeout_seconds FROM agent_definitions
WHERE type IN ('diagnose-agent','image-generator')
  AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;
