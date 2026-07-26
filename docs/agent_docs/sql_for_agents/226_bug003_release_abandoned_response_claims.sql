-- ============================================================================
-- 226_bug003_release_abandoned_response_claims.sql
--
-- bugs_open/003 — the last hole in F2's durable retry driver.
--
-- THE DEFECT. ProcessResponse takes an exclusive claim on the awaited request
-- ('waiting' -> 'processing', platform/orchestration/coordinator.go:247) BEFORE
-- it can still bail out — the orchestration state failing to load, or being
-- gone entirely. Every early exit after that claim returned without releasing
-- it, and NOTHING anywhere moved a row back out of 'processing':
--
--   * cleanup_expired_awaited_requests() (this function, before this migration)
--     expired only 'waiting' rows and cancelled only stale 'retrying' ones;
--   * F2's two claim paths take only 'expired' and stale 'retrying'
--     (ClaimAwaitedRequestForRetry / ClaimExpiredAwaitedRequestsForRetry).
--
-- So the row became invisible to every recovery path the platform has, and the
-- parent sat in AWAITING_RESPONSES with NO retry driver at all — bug 003's own
-- symptom, surviving inside bug 003's fix.
--
-- MEASURED 2026-07-26 21:15Z, before applying this: 181 rows parked in
-- 'processing', the oldest claimed 2026-06-26. 176 had lost their orchestration
-- entirely (pure leak — nothing reaps that status, so they accumulate for
-- ever). TWO had a live parent still waiting on them:
--   ba0051d3-1248-4afb-bf24-f89e77e6cc54  step deploy_page   stranded 80 min
--   7e6835ff-bbf5-46b2-83c1-50983b0e5742  step call_scraper  stranded 10 min
-- Both claiming pods were dead. The deploy_page one is a page that would never
-- have deployed.
--
-- THE FIX. Hand an abandoned claim back to the machinery that already exists,
-- rather than teaching a second subsystem about it. Reset to 'waiting' and the
-- rest of this function does the work on the same tick: the expire clause below
-- marks it 'expired' (its timeout_at is long past), F2's ticker then claims and
-- retries any whose parent is still AWAITING_RESPONSES, and the 7-day delete
-- finally reaps the orphans whose parent is gone. No Go change, no image roll.
--
-- WHY 15 MINUTES. A live response drive takes seconds, so this cannot race one.
-- It matches PROCESSED_MESSAGES_LEASE_SECONDS (900s default), which is the
-- window the same codebase already treats as "long enough that the holder is
-- dead" for the dedupe lease. Reusing that number rather than inventing one
-- keeps the two claim layers telling the same story.
--
-- ORDER MATTERS: the reset runs FIRST so a released row is expired on the same
-- tick rather than waiting another minute. deleted_count keeps its meaning —
-- it counts rows expired, which now legitimately includes freshly released
-- ones.
--
-- SAFETY. Guarded on processed_at IS NULL, so a row that genuinely completed
-- (CompleteAwaitedRequest sets status='processed' AND processed_at) can never
-- be resurrected. Orphans are NOT re-driven: F2's claim joins
-- orchestration_states on status='AWAITING_RESPONSES', so a released row whose
-- parent is gone simply expires and ages out. Nothing is re-executed that
-- should not be.
--
-- Mirror any edit into docs/agent_docs/sql_for_tables/ the way 205 was, or the
-- next re-seed clobbers it.
-- ============================================================================

CREATE OR REPLACE FUNCTION cleanup_expired_awaited_requests()
RETURNS INTEGER AS $$
DECLARE
deleted_count INTEGER;
BEGIN
    -- bugs_open/003: release response claims that were taken and then
    -- abandoned. See this migration's header for the full mechanism. Runs
    -- BEFORE the expire clause so a released row recovers on the same tick.
UPDATE awaited_requests
SET status = 'waiting',
    processing_started_at = NULL,
    processing_pod = NULL
WHERE status = 'processing'
  AND processed_at IS NULL
  AND processing_started_at < NOW() - INTERVAL '15 minutes';

    -- Mark expired requests
UPDATE awaited_requests
SET status = 'expired',
    processed_at = NOW()
WHERE status = 'waiting'
  AND timeout_at < NOW();

GET DIAGNOSTICS deleted_count = ROW_COUNT;

-- bugs_open/003 F2 backstop: cancel 'retrying' rows no actor is driving.
-- (The ticker reclaims live-but-orphaned claims at 5 min; 60 min here means
-- the orchestration itself moved on. processing_started_at is stamped by
-- every claim, so it cannot be NULL on a 'retrying' row.)
UPDATE awaited_requests
SET status = 'cancelled',
    processed_at = NOW()
WHERE status = 'retrying'
  AND processing_started_at < NOW() - INTERVAL '60 minutes';

-- Delete very old terminal requests (>7 days)
DELETE FROM awaited_requests
WHERE status IN ('processed','expired','cancelled','error')
  AND processed_at < NOW() - INTERVAL '7 days';

RETURN deleted_count;
END;
$$ LANGUAGE plpgsql;
