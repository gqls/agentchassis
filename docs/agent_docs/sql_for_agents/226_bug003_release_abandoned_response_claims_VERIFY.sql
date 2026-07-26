-- ============================================================================
-- VERIFY for 226_bug003_release_abandoned_response_claims.sql
--
-- Run AFTER applying. Every check prints PASS or FAIL — read them, do not
-- assume. Checks 3-5 are the ones that matter: they assert the defect is gone,
-- not merely that the code changed.
-- ============================================================================

\echo '--- 1. the release clause is actually in the live function body ---'
SELECT CASE WHEN prosrc LIKE '%status = ''processing''%'
             AND prosrc LIKE '%INTERVAL ''15 minutes''%'
            THEN 'PASS: release clause present'
            ELSE 'FAIL: live function does not carry the clause' END AS check_1
FROM pg_proc WHERE proname = 'cleanup_expired_awaited_requests';

\echo '--- 2. the pre-existing clauses survived the CREATE OR REPLACE ---'
SELECT CASE WHEN prosrc LIKE '%status = ''cancelled''%'
             AND prosrc LIKE '%INTERVAL ''60 minutes''%'
             AND prosrc LIKE '%INTERVAL ''7 days''%'
            THEN 'PASS: expire + cancel + delete clauses all intact'
            ELSE 'FAIL: a pre-existing clause was lost' END AS check_2
FROM pg_proc WHERE proname = 'cleanup_expired_awaited_requests';

\echo '--- 3. no row older than the window is still parked in processing ---'
\echo '    (the ticker runs every 60s; give it two minutes before believing a FAIL)'
SELECT CASE WHEN count(*) = 0
            THEN 'PASS: no abandoned claims left'
            ELSE 'FAIL: ' || count(*) || ' still parked' END AS check_3
FROM awaited_requests
WHERE status = 'processing'
  AND processed_at IS NULL
  AND processing_started_at < NOW() - INTERVAL '20 minutes';

\echo '--- 4. NEGATIVE CONTROL: in-flight claims were NOT disturbed ---'
\echo '    A claim younger than the window must still be processing. If this'
\echo '    reads 0 because there is no traffic, it proves nothing — say so.'
SELECT count(*) AS young_claims_still_processing,
       CASE WHEN count(*) FILTER (WHERE processing_started_at < NOW() - INTERVAL '15 minutes') = 0
            THEN 'PASS: nothing inside the window was reset'
            ELSE 'FAIL: a live claim was reset' END AS check_4
FROM awaited_requests
WHERE status = 'processing' AND processed_at IS NULL;

\echo '--- 5. THE POINT: no live parent is stranded on an unreleasable claim ---'
SELECT CASE WHEN count(*) = 0
            THEN 'PASS: zero AWAITING_RESPONSES parents blocked on a dead claim'
            ELSE 'FAIL: ' || count(*) || ' parents still stranded' END AS check_5
FROM awaited_requests a
JOIN orchestration_states o ON o.orchestration_id = a.orchestration_id
WHERE a.status = 'processing' AND a.processed_at IS NULL
  AND o.status = 'AWAITING_RESPONSES'
  AND a.processing_started_at < NOW() - INTERVAL '20 minutes';

\echo '--- 6. released rows went on to be driven or aged out, not lost ---'
SELECT status, count(*)
FROM awaited_requests
WHERE processing_pod IS NULL AND status <> 'waiting'
GROUP BY status ORDER BY 2 DESC;

\echo '--- 7. completed rows were never resurrected (processed_at guard held) ---'
SELECT CASE WHEN count(*) = 0
            THEN 'PASS: no processed row was reset to waiting'
            ELSE 'FAIL: ' || count(*) || ' processed rows went backwards' END AS check_7
FROM awaited_requests
WHERE processed_at IS NOT NULL AND status = 'waiting';
