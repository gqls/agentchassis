-- 214_build_dispatch_watchdog.sql
--
-- bugs_open/029 — turn a silent fleet-wide build halt into a durable row.
--
-- WHY
-- ---
-- 029's core complaint is not that builds stall; it is that they stall
-- *silently*: "Nothing errors, nothing alerts, no site reports a failure —
-- builds simply stop happening everywhere." Three reproductions were each found
-- by a human debugging something else, 22-25 minutes in.
--
-- Migration 213 makes the trigger honest (gate predicate == dispatcher
-- predicate), so "not firing" now means "nothing dispatchable". That fixes the
-- false heartbeat but it still does not SAY anything when the pipeline stops.
-- This does.
--
-- WHAT IT WATCHES
-- ---------------
-- The case file's own third-occurrence section worked out the right signal the
-- hard way, after being misled by the AWAITING_RESPONSES pool census:
--
--   "The signals that actually decide it: (a) does a *new* trigger reach
--    COMPLETED, and (b) completed_at throughput against the triaged backlog.
--    Check throughput-vs-backlog, not the pool census."
--
-- So: raise a row when NOTHING has completed fleet-wide for 45 minutes AND
-- there is either dispatchable work waiting or a claim stuck past its reset.
--
-- Note it deliberately uses completed_at, NOT updated_at — site_work_items.
-- updated_at is not maintained (bugs_open/035), and the case file records
-- getting the right answer from it for the wrong reason.
--
-- WHY 45 MINUTES
-- --------------
-- claimed-item-timeout resets orphaned claims at 40 min and that is the
-- load-bearing self-heal. A threshold below it fires during normal recovery and
-- trains everyone to ignore the row. At 45 min the self-heal has had its chance
-- and demonstrably failed. The handler timeout is 1200s (20 min), so a single
-- slow build cannot trip it either.
--
-- KNOWN BENIGN CASE (documented rather than engineered around)
-- ------------------------------------------------------------
-- An item blocked on an unmet `depends_on` keeps its site "dispatchable" from
-- find_dispatchable_site's point of view (that query does not check
-- dependencies; load_work_items does). A fleet whose only remaining work is
-- dependency-deadlocked will therefore raise one row per hour. That is a real
-- condition worth knowing about, not a false alarm — but it is not an outage,
-- so read the context payload before treating it as one.
--
-- ROLLBACK
-- --------
--   DELETE FROM scheduled_tasks WHERE name = 'build-dispatch-watchdog';
--   -- or, to silence without forgetting:
--   UPDATE scheduled_tasks SET enabled = false WHERE name = 'build-dispatch-watchdog';

BEGIN;

INSERT INTO scheduled_tasks (
    name, description, interval_seconds,
    target_agent_type, target_topic, fire_message,
    concurrency_group, max_concurrent, timeout_seconds, enabled,
    pre_query
) VALUES (
    'build-dispatch-watchdog',
    'bugs_open/029: raises one agent_error_log row when build dispatch has stalled — no completions fleet-wide for 45 min while work is dispatchable or a claim is stuck past its reset. CTE-only, fires no message.',
    300,
    'generic',
    'system.internal.noop',
    false,
    -- Its own concurrency group on purpose. Adding a watchdog to `maintenance`
    -- or `claim-management` would have it compete for those groups' single slot,
    -- which is the head-of-queue starvation bugs_open/048 documented.
    'watchdog',
    1,
    60,
    true,
    $pq$
WITH dispatchable AS (
    SELECT count(*) AS n FROM (
        SELECT DISTINCT wi.site_id
          FROM site_work_items wi
          JOIN sites s ON s.id = wi.site_id
         WHERE wi.status IN ('triaged', 'approved')
           AND wi.attempt_count < wi.max_attempts
           AND s.locked_at IS NULL
           AND NOT EXISTS (
               SELECT 1 FROM site_work_items active
                WHERE active.site_id = wi.site_id
                  AND active.status = 'claimed'
           )
    ) d
),
stuck_claims AS (
    SELECT count(*) AS n FROM site_work_items
     WHERE status = 'claimed'
       AND claimed_at < NOW() - INTERVAL '45 minutes'
),
throughput AS (
    SELECT count(*) AS n FROM site_work_items
     WHERE completed_at > NOW() - INTERVAL '45 minutes'
),
recent_alert AS (
    SELECT count(*) AS n FROM agent_error_log
     WHERE error_code = 'BUILD_DISPATCH_STALLED'
       AND occurred_at > NOW() - INTERVAL '1 hour'
),
alerted AS (
    INSERT INTO agent_error_log (
        agent_type, step_name, action,
        error_message, error_code, severity, context
    )
    SELECT
        'build-pipeline-trigger',
        'find_dispatchable_site',
        'scheduled_gate',
        format(
            'BUILD DISPATCH STALLED: 0 work items completed fleet-wide in 45 min while %s site(s) are dispatchable and %s claim(s) are stuck past the 40-min reset. See bugs_open/029.',
            (SELECT n FROM dispatchable), (SELECT n FROM stuck_claims)
        ),
        'BUILD_DISPATCH_STALLED',
        'error',
        jsonb_build_object(
            'dispatchable_sites', (SELECT n FROM dispatchable),
            'stuck_claims',       (SELECT n FROM stuck_claims),
            'completions_45m',    (SELECT n FROM throughput),
            'detected_by',        'build-dispatch-watchdog',
            'bug',                'bugs_open/029'
        )
    WHERE (SELECT n FROM throughput) = 0
      AND ((SELECT n FROM dispatchable) > 0 OR (SELECT n FROM stuck_claims) > 0)
      AND (SELECT n FROM recent_alert) = 0
    RETURNING id
)
SELECT (SELECT n FROM dispatchable)::text  AS dispatchable_sites,
       (SELECT n FROM stuck_claims)::text  AS stuck_claims,
       (SELECT count(*) FROM alerted)::text AS alerts_raised
HAVING (SELECT count(*) FROM alerted) > 0
$pq$
)
ON CONFLICT (name) DO UPDATE
SET description       = EXCLUDED.description,
    interval_seconds  = EXCLUDED.interval_seconds,
    target_agent_type = EXCLUDED.target_agent_type,
    target_topic      = EXCLUDED.target_topic,
    fire_message      = EXCLUDED.fire_message,
    concurrency_group = EXCLUDED.concurrency_group,
    max_concurrent    = EXCLUDED.max_concurrent,
    timeout_seconds   = EXCLUDED.timeout_seconds,
    enabled           = EXCLUDED.enabled,
    pre_query         = EXCLUDED.pre_query,
    updated_at        = NOW();

DO $guard$
DECLARE
    v_pre     text;
    v_fire    boolean;
    v_enabled boolean;
    v_group   text;
BEGIN
    SELECT pre_query, fire_message, enabled, concurrency_group
      INTO v_pre, v_fire, v_enabled, v_group
      FROM scheduled_tasks WHERE name = 'build-dispatch-watchdog';

    IF v_pre IS NULL THEN
        RAISE EXCEPTION '214: build-dispatch-watchdog was not created';
    END IF;
    IF v_fire IS DISTINCT FROM false THEN
        RAISE EXCEPTION '214: watchdog must be CTE-only (fire_message=false), got %', v_fire;
    END IF;
    IF v_enabled IS DISTINCT FROM true THEN
        RAISE EXCEPTION '214: watchdog is not enabled';
    END IF;
    IF v_group IS DISTINCT FROM 'watchdog' THEN
        RAISE EXCEPTION '214: watchdog must sit in its own concurrency group, got %', v_group;
    END IF;
    IF v_pre NOT LIKE '%completed_at%' THEN
        RAISE EXCEPTION '214: watchdog must measure throughput via completed_at (updated_at is not maintained — bugs_open/035)';
    END IF;
    IF v_pre LIKE '%AWAITING_RESPONSES%' THEN
        RAISE EXCEPTION '214: watchdog must not key on the pool census — that is the signal that misled 029 three times';
    END IF;

    RAISE NOTICE '214: build-dispatch-watchdog installed (300s, CTE-only, group=watchdog)';
END
$guard$;

COMMIT;
