-- ============================================================================
-- 249_chassis_intake_events_VERIFY.sql
--
-- Hand-run companion to 249. UPPERCASE-suffixed so run-migrations.sh's
-- SIDECAR_RE never auto-applies it. SELECT-only; re-runnable.
--
-- Run once after applying 249 (all four checks return the stated value on an
-- idle table), and again after the flag flips (check 4 is the one whose answer
-- changes — rows appearing is the point).
-- ============================================================================

-- 1. Both tables exist with the expected columns. Expect 2 rows.
SELECT table_name, count(*) AS columns
FROM information_schema.columns
WHERE table_name IN ('chassis_intake_events','chassis_orchestration_claims')
GROUP BY table_name
ORDER BY table_name;

-- 2. The transport idempotency key is UNIQUE, not merely an index. Expect 1.
SELECT count(*) FROM pg_indexes
WHERE tablename = 'chassis_intake_events'
  AND indexname = 'idx_cie_exactly_once'
  AND indexdef LIKE 'CREATE UNIQUE INDEX%';

-- 3. The status vocabularies match what the Go writes (intake_repo.go).
--    Expect both check constraints present. Expect 2.
SELECT count(*) FROM information_schema.check_constraints
WHERE constraint_name LIKE '%chassis_intake_events%'
  AND (check_clause LIKE '%pending%' OR check_clause LIKE '%request%');

-- 4. Liveness after the flag flips: intake rows arriving, claims cycling.
--    Before the flip both counts are 0 — that is the dark check.
SELECT
  (SELECT count(*) FROM chassis_intake_events
    WHERE received_at > NOW() - INTERVAL '15 minutes')                AS intake_last_15m,
  (SELECT count(*) FROM chassis_intake_events
    WHERE status IN ('pending','running'))                            AS in_flight_now,
  (SELECT count(*) FROM chassis_orchestration_claims)                 AS live_claims,
  (SELECT count(*) FROM chassis_intake_events WHERE status='failed')  AS failed_ever;
