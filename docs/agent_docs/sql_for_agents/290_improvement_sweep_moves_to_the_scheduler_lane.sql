-- 290 — improvement-sweep: move the producer onto the scheduler lane
--
-- vigilant_designer_offer_analysis programme, Phase A0.1 (PLAN 2026-08-02).
--
-- This migration deliberately does NOT touch `enabled`. The owner's 2026-07-29
-- ruling (improvement loop stopped deliberately during heavy development)
-- stands until the separate G1 go recorded in the PLAN; this file moves the
-- topic only, so that when G1 comes, one flag flip is the whole change.
--
-- WHY THE TOPIC MOVES. improvement-sweep still produces to
-- system.agent.generic.requests. That topic is NOT dead — it is the chassis's
-- MAIN request lane (REQUESTS_TOPIC on the live agent-chassis deployment,
-- verified 2026-08-02) — but it is the wrong lane for a scheduler producer,
-- for the reason recorded on setupExtraRequestLanes (platform/agentbase/
-- agent.go, the bugs_open/030 work): scheduler chores measured at 93% of the
-- shared lane's traffic and held interactive dispatches behind 8–15 minute
-- chains, so scheduled dispatch got its own lane. Every working enabled task
-- (18/18, measured 2026-07-30 by the robot_hands_checker_gaps lane) produces
-- to system.agent.scheduled.requests; improvement-sweep (enabled=f since
-- 2026-05-02) simply predates the split.
--
-- PRECISION NOTE, because a wrong mechanism propagates: the checker-gaps NOTES
-- call generic.requests "the default topic nothing consumes". Imprecise — the
-- live deployment consumes it as its main lane. What IS true: the one recorded
-- scheduler fire at it (oneshot-discovery-aao, 2026-07-26) produced no
-- orchestration, and the mechanism there is unresolved (a stamped
-- last_triggered_at is not a delivered message). Aligning this producer with
-- the 18/18 working population removes the variable either way.
--
-- ORDERING SAFETY. agent.go's own rule is "the image that consumes a lane
-- ships BEFORE any producer points at it". Already satisfied: the live
-- deployment carries EXTRA_REQUEST_TOPICS=system.agent.scheduled.requests,…
-- so the consumer exists today.

BEGIN;

UPDATE scheduled_tasks
SET target_topic = 'system.agent.scheduled.requests',
    updated_at   = now()
WHERE name = 'improvement-sweep'
  AND target_topic = 'system.agent.generic.requests';

-- ENFORCING CHECK — the exact post-conditions, or the file rolls back:
-- (i) the row exists; (ii) it produces to the scheduler lane; (iii) this
-- migration changed nothing else about it that G1 depends on (interval,
-- target agent). `enabled` is reported, not asserted: this file must leave it
-- as found, and asserting false would wrongly abort a re-run after G1.
DO $$
DECLARE
    r record;
BEGIN
    SELECT target_topic, target_agent_type, interval_seconds, enabled
    INTO r
    FROM scheduled_tasks WHERE name = 'improvement-sweep';

    IF r IS NULL THEN
        RAISE EXCEPTION 'improvement-sweep row not found';
    END IF;
    IF r.target_topic <> 'system.agent.scheduled.requests' THEN
        RAISE EXCEPTION 'improvement-sweep still produces to %', r.target_topic;
    END IF;
    IF r.target_agent_type <> 'improvement-loop' THEN
        RAISE EXCEPTION 'improvement-sweep target_agent_type unexpectedly %', r.target_agent_type;
    END IF;

    RAISE NOTICE 'improvement-sweep -> system.agent.scheduled.requests (interval %ss); enabled=% — G1 is a separate owner decision',
        r.interval_seconds, r.enabled;
END $$;

COMMIT;

-- ── ROLLBACK ──
-- UPDATE scheduled_tasks SET target_topic='system.agent.generic.requests', updated_at=now()
-- WHERE name='improvement-sweep';
