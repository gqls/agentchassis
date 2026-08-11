-- ============================================================================
-- 389 — park the contrast_failure backlog, then re-enable improvement-sweep
--       so the PAGE RE-RENDERS can drain (owner decision, 2026-08-11)
-- ============================================================================
-- OWNER DECISION, verbatim: "lets reenable improvement-sweep for the rerenders
-- for a short while - it will be expensive so I am wary of costs."
--
-- WHY THE PARK IS PART OF "FOR THE RERENDERS", not a liberty taken with it.
-- `improvement-sweep`'s pre_query skips any site with **50 or more** items in
-- (triaged, detected) on pipeline='build'. Migration 369's weekly render audit
-- filed 220 `contrast_failure` items overnight, and that pushed FIVE sites over
-- that guard — so enabling the sweep as-is would SKIP exactly the sites holding
-- the most re-renders. Measured 2026-08-11 before writing:
--
--   site                              backlog  contrast  rerenders   after park
--   robot-hands.com                      68       34        17       34  (in)
--   loancalculator.co.uk                 58        0        12       58  (still out)
--   vonc.com                             55       38        13       17  (in)
--   leopardessconsulting.co.uk           53        8        22       45  (in)
--   loanandmortgagecalculator.co.uk      52        3        40       49  (in)
--
-- So the park un-blocks 4 of the 5 over-guard sites, including the two with the
-- most re-render work (40 and 22). loancalculator stays out on its own backlog,
-- which is not this lane's to clear.
--
-- WHY PARK RATHER THAN PROMOTE. `triage_detected_items` is site-scoped and
-- TYPE-BLIND (`triage_detect_items_action.go:162-173` — `WHERE site_id = $1 AND
-- status = 'detected'`), so a sweep promotes contrast_failure alongside the
-- re-renders whether or not anyone wants it to. Those items route to
-- `css-patch-agent`, where **`bugs_open/213`'s false-complete defect is
-- unfixed** — it can stamp an item `complete` having written nothing. Promoting
-- 220 of them converts an honest, queryable backlog into 220 false closures, and
-- a `complete` row is far harder to find later than a `detected` one. The agreed
-- ordering is 213 FIRST; this park is what lets the re-renders proceed without
-- pre-empting it.
--
-- WHY `deferred` IS THE RIGHT PARK — checked, not assumed:
--   * `triage_detected_items` only promotes `status='detected'`, so a deferred
--     item cannot be swept up. Verified at the action's own WHERE clause.
--   * `deferred` is **NOT** in `idx_swi_dedup`'s terminal list
--     (complete/verified/rejected/wont_fix/failed/unresolved/cancelled), so a
--     parked row STILL HOLDS ITS DEDUP SLOT and the next weekly audit will not
--     re-file a duplicate. Verified against `pg_indexes.indexdef`. This is the
--     property that makes the park safe to leave in place for weeks.
--   * It is reversible by one UPDATE (foot of this file), and every parked row
--     records why in `spec.parked_*`.
--
-- COST CONTROL (the owner's stated concern). The sweep's own cadence was
-- `interval_seconds = 180` — a full improvement-loop run every 3 minutes, ~20
-- site-runs an hour, indefinitely. This migration re-enables it at **900s**
-- (~4/hour), which is a 5x cut in burn rate while still draining. **It is left
-- ENABLED and someone must turn it off** — see the STOP block at the foot.
--
-- Rollback (both halves) is at the foot of this file.

BEGIN;

-- ---------------------------------------------------------------------------
-- 1. Park every DETECTED contrast_failure item.
-- ---------------------------------------------------------------------------
DO $$
DECLARE
    n_before int;
    n_parked int;
    n_left   int;
BEGIN
    SELECT count(*) INTO n_before
      FROM site_work_items
     WHERE item_type = 'contrast_failure' AND status = 'detected';

    IF n_before = 0 THEN
        RAISE EXCEPTION '389: no detected contrast_failure items — already parked, or the premise is gone. STOP and re-measure.';
    END IF;

    UPDATE site_work_items
       SET status = 'deferred',
           spec = jsonb_set(
                    jsonb_set(
                      jsonb_set(COALESCE(spec, '{}'::jsonb),
                                '{parked_from_status}', '"detected"'::jsonb),
                      '{parked_reason}',
                      '"bugs_open/213 false-complete on css-patch-agent; parked 2026-08-11 so improvement-sweep can drain page_rerender without promoting these. Restore to detected when 213 is fixed - see migration 389 rollback."'::jsonb),
                    '{parked_by}', '"migration_389"'::jsonb),
           updated_at = now()
     WHERE item_type = 'contrast_failure' AND status = 'detected';

    GET DIAGNOSTICS n_parked = ROW_COUNT;

    IF n_parked <> n_before THEN
        RAISE EXCEPTION '389: parked % of % — race, STOP', n_parked, n_before;
    END IF;

    -- Negative control: nothing else moved.
    SELECT count(*) INTO n_left
      FROM site_work_items
     WHERE item_type = 'contrast_failure' AND status = 'detected';
    IF n_left <> 0 THEN
        RAISE EXCEPTION '389: % detected contrast_failure remain after park', n_left;
    END IF;

    RAISE NOTICE '389/park: % contrast_failure items detected -> deferred', n_parked;
END $$;

-- ---------------------------------------------------------------------------
-- 2. Re-enable improvement-sweep, at a reduced cadence.
-- ---------------------------------------------------------------------------
DO $$
DECLARE
    was_enabled  boolean;
    old_interval int;
    n_over_guard int;
BEGIN
    SELECT enabled, interval_seconds INTO was_enabled, old_interval
      FROM scheduled_tasks WHERE name = 'improvement-sweep';

    IF NOT FOUND THEN
        RAISE EXCEPTION '389: scheduled_tasks row improvement-sweep not found — STOP';
    END IF;
    IF was_enabled THEN
        RAISE EXCEPTION '389: improvement-sweep is ALREADY enabled — someone else moved it; STOP and reconcile';
    END IF;

    UPDATE scheduled_tasks
       SET enabled = true,
           interval_seconds = 900,
           updated_at = now()
     WHERE name = 'improvement-sweep';

    -- Prove the park actually bought what it was for: count sites still over
    -- the pre_query's own < 50 guard. Recorded in the log, not asserted as a
    -- fixed number, because other lanes file items continuously.
    SELECT count(*) INTO n_over_guard
      FROM sites s
     WHERE s.status IN ('active','deployed')
       AND (SELECT count(*) FROM site_work_items wi
             WHERE wi.site_id = s.id
               AND wi.status IN ('triaged','detected')
               AND wi.pipeline = 'build') >= 50;

    RAISE NOTICE '389/sweep: improvement-sweep enabled, interval %s -> 900s. Sites still over the <50 guard: %',
                 old_interval, n_over_guard;
END $$;

COMMIT;

-- ============================================================================
-- ⚠ STOP CONDITION — THIS TASK IS LEFT RUNNING AND MUST BE TURNED OFF
-- ============================================================================
-- The owner asked for "a short while". Nothing in this migration expires.
-- Turn it off with:
--
--   UPDATE scheduled_tasks SET enabled = false, updated_at = now()
--    WHERE name = 'improvement-sweep';
--
-- Watch cost while it runs (baseline was taken immediately before the apply and
-- is recorded in the lane's NOTES):
--
--   SELECT date_trunc('hour', created_at) AS hr, count(*) AS calls,
--          sum(input_tokens) AS in_tok, sum(output_tokens) AS out_tok
--     FROM llm_call_log WHERE created_at > now() - interval '6 hours'
--    GROUP BY 1 ORDER BY 1;
--
-- Watch progress:
--
--   SELECT status, count(*) FROM site_work_items
--    WHERE item_type='page_rerender' AND created_at > '2026-07-24' GROUP BY 1;
--
-- ============================================================================
-- ROLLBACK
-- ============================================================================
-- Half 2 (the sweep) — restore the original disabled state and cadence:
--   UPDATE scheduled_tasks SET enabled = false, interval_seconds = 180,
--          updated_at = now()
--    WHERE name = 'improvement-sweep';
--
-- Half 1 (the park) — ONLY once bugs_open/213 is fixed and live, else they will
-- promote straight into the defect this park exists to avoid:
--   UPDATE site_work_items
--      SET status = 'detected',
--          spec = (spec - 'parked_from_status' - 'parked_reason' - 'parked_by'),
--          updated_at = now()
--    WHERE item_type = 'contrast_failure'
--      AND status = 'deferred'
--      AND spec->>'parked_by' = 'migration_389';
-- (The `parked_by` predicate is what keeps this from disturbing any deferred
--  contrast_failure row parked by somebody else, now or later.)
-- Row-level backup: scratchpad/backups/backup_park_contrast_failure_20260811.tsv
