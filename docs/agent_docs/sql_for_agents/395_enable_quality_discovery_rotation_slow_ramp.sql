-- ============================================================================
-- 395 — re-enable ONE discovery rotation (quality) on a slow ramp
--       (owner decision, 2026-08-12)
-- ============================================================================
-- OWNER DECISION: after reading the cost and lock-out evidence from the
-- 2026-08-11 improvement-sweep run, the owner chose "enable the discovery
-- rotations, slowly" over raising the sweep's guard or making its discovery
-- step conditional. This migration does the first, minimal step of that: ONE
-- rotation, on a longer poll interval than it shipped with.
--
-- WHY THE ROTATION AND NOT THE SWEEP. These are two different mechanisms and
-- only one of them is safe to leave running:
--   * `improvement-sweep` (improvement-loop) both FINDS and DISPATCHES work. Its
--     pre_query skips a site with >= 50 items in (triaged, detected) on
--     pipeline='build' — and because PROMOTION moves a row from one member of
--     that list to the other WITHOUT reducing the count, the sweep re-locks
--     sites with its own output. Measured over its 5h29m run on 2026-08-11:
--     sites over the guard 5 -> 1 (after migration 389's 226-row park) -> 8, and
--     ~105 items filed per hour against ~48/h completed. Register IMP-016 and
--     the corrected `LANDMINES.md` entry carry the full figures.
--   * SCH-025's `site-discovery-rotation-*` tasks (this file) only FIND. They
--     have NO backlog cap, and they are fair by construction: the pre_query
--     STAMPS `site_discovery_rotation.last_selected_at = now()` in the same
--     statement that selects the site, so a selected site is immediately
--     not-due for 7 days. That is the fix for the sweep's `ORDER BY
--     sites.updated_at ASC` starvation (IMP-010), where nothing the sweep does
--     advances its own sort key.
--
-- WHY 10800s IS "SLOWLY", AND WHY THE INTERVAL IS NOT THE COST DIAL.
-- Read the pre_query before reasoning about the rate: it is `LIMIT 1` (one site
-- per fire) against a 7-day due window, with stamp-on-selection. So the WORK
-- RATE IS BOUNDED BY THE ROTATION PERIOD, NOT BY THE CADENCE: with 22
-- active/deployed sites and a 7-day period, steady state is 22/7 ~= 3.1 site
-- passes per DAY however often the task polls. Once every site is stamped, `due`
-- returns zero rows and a fire dispatches nothing and costs nothing.
--   The only thing the interval actually sizes is the INITIAL RAMP, when all 22
--   sites are due at once:
--     3600s  (as shipped) -> 22 passes in ~22 hours   (whole fleet in a day)
--     10800s (this file)   -> 22 passes in ~66 hours   (~8/day, ~3-day ramp)
--   10800s is chosen so the spend can be judged after a handful of sites rather
--   than after the entire fleet. It is not a throttle on steady state; nothing
--   is, except the 7-day window.
--
-- COST, MARKED AS AN ESTIMATE BECAUSE IT IS ONE.
-- [ESTIMATED — UPPER BOUND, not measured for this agent] From the 2026-08-11
-- sweep: ~22 fires consumed ~4.03M input tokens, i.e. ~180k input tokens per
-- fire. That figure is an UPPER BOUND for a rotation fire, because a sweep fire
-- is discovery PLUS triage PLUS dispatch, while a rotation fire is discovery
-- only. So the initial 22-site ramp is <= ~4M input tokens spread over ~3 days,
-- and steady state <= ~550k/day. Measure it against the real thing rather than
-- trusting this arithmetic — the watch query is at the foot of this file.
--
-- ⚠ WHAT THIS BUYS AND WHAT IT DOES NOT: OBSERVE-ONLY.
-- SCH-025 findings are born `detected` and there is NO triage carrier running
-- (`improvement-sweep` is disabled and stays disabled — this migration does not
-- touch it). So this restores DETECTION, not repair: problems become queryable,
-- and nothing fixes them automatically. That is the state register IMP-016
-- describes as "544 detected findings accumulating with nothing able to consume
-- them", and it is being entered deliberately this time, with two consequences
-- written down rather than discovered later:
--   1. `detected` rows COUNT toward `improvement-sweep`'s >= 50 guard. Anyone
--      re-enabling the sweep later must re-read the guard census first, because
--      this rotation will have been quietly raising the numbers it reads.
--   2. This lane's own 226 `contrast_failure` items stay PARKED at `deferred`
--      (migration 389) and are unaffected: `deferred` is not `detected`, and the
--      park's real trigger is now "contrast_failure gets a registered verifier",
--      NOT "bugs_open/213 closes" (corrected 2026-08-12 — contrast_failure
--      appears at no RegisterVerifier* call site, so promoting those items would
--      mint completions that are ungraded by construction).
--
-- SCOPE: `quality` ONLY. `design` and `completeness` stay disabled. The owner
-- asked for slowly; three rotations at once is not slowly, and one is enough to
-- price the mechanism. Adding the others later is one UPDATE each (at the foot).
-- `site-render-audit-rotation` is deliberately untouched — it is this lane's
-- contrast audit, already enabled, and measured at zero LLM spend.
-- ============================================================================

BEGIN;

-- Guard: refuse if the row is not in the state this migration assumes. Written
-- as a DO/RAISE rather than a SELECT because ON_ERROR_STOP ignores a non-empty
-- result set — a verify block made of SELECTs cannot stop the COMMIT.
DO $$
DECLARE
  v_enabled  boolean;
  v_interval integer;
  v_sweep    boolean;
BEGIN
  SELECT enabled, interval_seconds INTO v_enabled, v_interval
    FROM scheduled_tasks WHERE name = 'site-discovery-rotation-quality';

  IF NOT FOUND THEN
    RAISE EXCEPTION '395: scheduled_tasks row site-discovery-rotation-quality is MISSING — SCH-025 may not be seeded on this database';
  END IF;

  IF v_enabled THEN
    RAISE EXCEPTION '395: site-discovery-rotation-quality is ALREADY enabled — another session has enabled it; re-read the cost baseline before changing the cadence underneath them';
  END IF;

  -- The whole rationale above depends on the sweep being OFF. If someone has
  -- re-enabled it, this migration's "observe-only, no triage carrier" claim is
  -- false and its cost estimate is wrong, so refuse rather than proceed.
  SELECT enabled INTO v_sweep FROM scheduled_tasks WHERE name = 'improvement-sweep';
  IF v_sweep THEN
    RAISE EXCEPTION '395: improvement-sweep is ENABLED — this migration assumes it is off (observe-only, and its >= 50 guard reads the rows this rotation files). Resolve the conflict with that session first';
  END IF;

  RAISE NOTICE '395: pre-state OK — rotation disabled at interval %s, sweep off', v_interval;
END $$;

UPDATE scheduled_tasks
   SET enabled          = true,
       interval_seconds = 10800,   -- ~3-day initial ramp; see arithmetic above
       updated_at       = now()
 WHERE name = 'site-discovery-rotation-quality';

-- Post-check, same DO/RAISE discipline: assert row IDENTITY, not just a count.
DO $$
DECLARE
  v_enabled  boolean;
  v_interval integer;
  v_due      integer;
BEGIN
  SELECT enabled, interval_seconds INTO v_enabled, v_interval
    FROM scheduled_tasks WHERE name = 'site-discovery-rotation-quality';

  IF NOT (v_enabled AND v_interval = 10800) THEN
    RAISE EXCEPTION '395: post-check FAILED — enabled=%, interval=%', v_enabled, v_interval;
  END IF;

  -- How many sites are due right now. This is the ramp size, and it is the
  -- number to expect passes for over the next ~3 days.
  SELECT count(*) INTO v_due
    FROM sites s
    LEFT JOIN site_discovery_rotation r
      ON r.site_id = s.id AND r.agent_type = 'quality-discovery-agent'
   WHERE s.status IN ('active', 'deployed')
     AND COALESCE(r.last_selected_at, '-infinity'::timestamptz) < now() - interval '7 days';

  RAISE NOTICE '395: OK — quality rotation enabled at 10800s; % site(s) due for the initial ramp', v_due;
END $$;

COMMIT;

-- ============================================================================
-- WATCH IT (run these, do not trust the estimate above)
-- ============================================================================
-- Spend per hour. Compare against a no-rotation baseline taken the same way;
-- for reference, 2026-08-11 pre-sweep ran ~248k input tokens/h and the driven
-- sweep ~806k/h. Count CALLS and TOKENS together — calls/hour alone did not
-- discriminate the sweep's 3.2x, because each fire is a full site pass.
--   SELECT date_trunc('hour', created_at) AS hr, count(*) AS calls,
--          sum(input_tokens) AS in_tok, sum(output_tokens) AS out_tok
--     FROM llm_call_log WHERE created_at > now() - interval '12 hours'
--    GROUP BY 1 ORDER BY 1;
--
-- Is it actually rotating, and fairly? Every site should acquire a stamp over
-- the ramp, and none should be re-selected inside 7 days.
--   SELECT s.domain, r.last_selected_at
--     FROM sites s LEFT JOIN site_discovery_rotation r
--       ON r.site_id = s.id AND r.agent_type = 'quality-discovery-agent'
--    WHERE s.status IN ('active','deployed')
--    ORDER BY r.last_selected_at ASC NULLS FIRST;
--
-- What is it finding, and is it accumulating unconsumed (expected: yes)?
--   SELECT item_type, status, count(*) FROM site_work_items
--    WHERE created_at > '2026-08-12' GROUP BY 1,2 ORDER BY 3 DESC;
--
-- The sweep's guard census, because this rotation raises the numbers it reads:
--   SELECT count(*) FILTER (WHERE b >= 50) AS locked_out FROM (
--     SELECT (SELECT count(*) FROM site_work_items wi
--              WHERE wi.site_id = s.id AND wi.status IN ('triaged','detected')
--                AND wi.pipeline = 'build') b
--       FROM sites s WHERE s.status IN ('active','deployed')) t;
--
-- ============================================================================
-- STOP IT (one UPDATE, no migration needed — nothing expires on its own)
-- ============================================================================
--   UPDATE scheduled_tasks SET enabled = false, updated_at = now()
--    WHERE name = 'site-discovery-rotation-quality';
--
-- ADD THE OTHER TWO, if the quality one proves affordable:
--   UPDATE scheduled_tasks SET enabled = true, interval_seconds = 10800,
--          updated_at = now()
--    WHERE name IN ('site-discovery-rotation-design',
--                   'site-discovery-rotation-completeness');
-- ============================================================================
