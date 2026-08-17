-- 458 — detected-item-promoter: report what the doors HELD, and stop claiming
--       a concurrency slot on a tick that did nothing
--       (bugs_open/083, the guardian seat's open residual from council 05a3d1c8)
--
-- 430, 444 and 454 are ledger-recorded and are NOT edited. This file rewrites the
-- live `scheduled_tasks.detected-item-promoter` pre_query, guarded on a verbatim
-- pre-image so that a concurrent edit by another session makes it ABORT rather
-- than silently revert that session's work (444 -> 454 happened four hours apart
-- on this same column, by a lane that is still open).
--
-- ============================================================================
-- WHAT THE RESIDUAL ACTUALLY IS, AND WHY ITS STATED REMEDY DOES NOT WORK
-- ============================================================================
-- bugs_open/083 records the guardian's residual with a suggested cheap control:
--   "have the pre_query return a third column counting `detected` rows a door
--    held, so the scheduler's own log makes it visible."
--
-- [MEASURED 2026-08-17 at the running service] That remedy is INOPERATIVE as
-- stated, and the column alone would have been write-only. Every tick of this
-- task emits exactly one line and it carries no numbers at all:
--
--   {"caller":"scheduler/main.go:274",
--    "msg":"Pre-query task completed (no message fired)",
--    "task":"detected-item-promoter"}
--
-- Six such lines in `kubectl logs -l app=kafka-scheduler --tail=3000`, no counts.
-- For a `fire_message=false` task the pre_query result is merged into inputData
-- and then discarded unlogged, so `promoted` and `pairs` have never been readable
-- either. That is SCH-006's observability gap and it affects EVERY CTE-only task,
-- not this one — so it is fixed in the scheduler (cmd/scheduler/main.go, same
-- commit as this file), and this migration supplies the numbers worth logging.
-- THIS FILE IS INERT UNTIL kafka-scheduler IS REBUILT AND ROLLED.
--
-- ============================================================================
-- CHANGE 1 — one predicate evaluation, not two (why this is a restructure)
-- ============================================================================
-- The obvious way to count held rows is to write the door predicates a second
-- time, negated. That would create two copies of a rule that must stay in
-- lockstep — and this rule has already changed twice in one day (444 added the
-- pipeline allow-list and the success floor; 454 corrected both to count
-- `verified`). A mirrored copy would have drifted within four hours of being
-- written. It is the `idx_swi_dedup` / `workItemTerminalStatuses` lockstep trap
-- in a new place.
--
-- So the predicates move into a `scored` CTE that evaluates each door ONCE per
-- row as a boolean. `candidates` selects the rows where all doors pass; `held`
-- selects the rows where any door fails, and can say WHICH. There is exactly one
-- copy of every predicate, and the two sets are complementary by construction.
-- Every predicate below is byte-identical in meaning to the live one it replaces
-- — the verify block asserts the promotable set is UNCHANGED by this migration.
--
-- ============================================================================
-- CHANGE 2 — a tick that did nothing must return NO ROW
-- ============================================================================
-- The live final SELECT ends:
--     FROM promoted WHERE (SELECT COUNT(*) FROM promoted) > 0
-- That does NOT suppress the row. The target list is aggregate-only with no
-- GROUP BY, so Postgres returns exactly one row regardless — verified read-only
-- against an empty CTE: the WHERE form returns ('0', NULL), the HAVING form
-- returns 0 rows. Two consequences, both live today:
--
--   * the scheduler cannot distinguish an acting tick from an idle one: the
--     `dynamicData == nil` branch (main.go:200-216) is unreachable for this task,
--     so it logs the same completion line either way;
--   * the task claims its concurrency slot on EVERY tick, including idle ones.
--     `maintenance` is max_concurrent=1 with four other tasks in it, and
--     bugs_open/048's fix — release the slot on a no-op (dc2e4b61a) — works by
--     returning zero rows, which this shape opts out of.
--
-- HONEST SIZING: this is a door, not a repair. [MEASURED 2026-08-17] no
-- `maintenance` group-mate is starved today — all five are within 1.04 intervals
-- of due (`feasibility-recheck` 1.04, `database-cleanup` 0.22, `work-item-archiver`
-- 0.19, `stale-work-item-reaper` 0.11, this task 0.03). The claim is that the
-- shape is wrong and cheap to correct, NOT that it is currently costing throughput.
--
-- The suppression is `promoted = 0 AND held = 0`, NOT `promoted = 0`. Suppressing
-- on promoted alone would hide the held count on exactly the ticks where it is the
-- only thing worth saying — an idle tick that is holding rows is the state this
-- residual is about.
--
-- ============================================================================
-- WHAT IT REPORTS TODAY, AND WHY THAT IS A POSITIVE CONTROL
-- ============================================================================
-- [MEASURED 2026-08-17 22:5xZ] the live held set is 2 rows, and one of them is
-- 444's own door firing on a live row for the first time:
--     literal_markdown    -> page-build-handler   1 row  HELD BY THE SUCCESS FLOOR
--     placeholder_contact -> page-build-handler   1 row  HELD BY THE KNOWN-GOOD RULE
-- 444 recorded both doors as holding ZERO rows at apply and said "they are doors,
-- not repairs". One of them is now load-bearing and nothing can see it. That is
-- the residual, instantiated.
--
-- The verify block below therefore carries a POSITIVE control: `held` must be
-- non-zero, or the counter has not been shown capable of counting anything and
-- the assert is vacuous. It also carries a NEGATIVE control: the promotable set
-- must be IDENTICAL before and after, or the restructure changed behaviour.
--
-- ============================================================================
-- ORDER / ROLLBACK
-- ============================================================================
-- DB config: live at the scheduler's next tick after COMMIT, but the new columns
-- are only VISIBLE once kafka-scheduler carries the logging change. Applying this
-- before the roll is safe and does nothing observable except the suppression.
-- Rollback is a separate file (458_..._ROLLBACK.sql) restoring the pre-image
-- verbatim.

BEGIN;

-- GUARD: refuse if the live pre_query is not the exact text this file was
-- written against. Another session owns this column too; clobbering its edit
-- would be silent and would look like a revert nobody made.
DO $$
DECLARE
    live_md5 text;
    live_len int;
BEGIN
    SELECT md5(pre_query), length(pre_query) INTO live_md5, live_len
      FROM scheduled_tasks WHERE name = 'detected-item-promoter';

    IF live_md5 IS NULL THEN
        RAISE EXCEPTION '458: no detected-item-promoter row — nothing to guard';
    END IF;

    IF live_md5 <> '1d1efee2913929db7b6b5395d8421ecc' THEN
        RAISE EXCEPTION
          '458: ABORTING — the live pre_query is not the text this migration was written against (expected md5 1d1efee2913929db7b6b5395d8421ecc len 2340, found % len %). Another session has edited it since 2026-08-17 22:5xZ. Re-read the live column, re-derive this file against it, and do NOT force: overwriting is how one lane silently reverts another.',
          live_md5, live_len;
    END IF;
END $$;

UPDATE scheduled_tasks
SET pre_query = $Q$
    WITH scored AS (
        -- Every door evaluated ONCE per row. candidates and held are
        -- complementary halves of this set, so the two can never drift apart.
        SELECT wi.id,
               wi.item_type,
               wi.handler_agent,
               wi.created_at,
               -- DOOR-CLOSER 1 (444): only pipelines this transition has ever produced.
               (wi.pipeline IN ('build', 'content', 'design')) AS pipe_ok,
               EXISTS (
                 SELECT 1 FROM agent_definitions ad
                 WHERE ad.type = wi.handler_agent
                   AND ad.is_active
                   AND COALESCE(ad.is_snapshot, false) = false
                   AND ad.deleted_at IS NULL
               ) AS handler_ok,
               -- KNOWN-GOOD (430, corrected by 454): `verified` is a completion that
               -- also passed verification. Counting only 'complete' would hold a pair
               -- whose every success has been verified.
               EXISTS (
                 SELECT 1 FROM site_work_items done
                 WHERE done.item_type = wi.item_type
                   AND done.handler_agent = wi.handler_agent
                   AND done.status IN ('complete', 'verified')
               ) AS known_good,
               -- DOOR-CLOSER 2 (444, corrected by 454): >=5 terminal outcomes => must
               -- still be >=25% good. NB `failed` rows have no completed_at, so this
               -- keys on status only.
               (
                 SELECT (c + f) < 5 OR c >= 0.25 * (c + f)
                 FROM (
                   SELECT count(*) FILTER (WHERE h.status IN ('complete', 'verified')) AS c,
                          count(*) FILTER (WHERE h.status = 'failed')                  AS f
                   FROM site_work_items h
                   WHERE h.item_type = wi.item_type
                     AND h.handler_agent = wi.handler_agent
                 ) hist
               ) AS floor_ok
        FROM site_work_items wi
        WHERE wi.status = 'detected'
          AND COALESCE(wi.handler_agent, '') <> ''
    ),
    candidates AS (
        SELECT id FROM scored
        WHERE pipe_ok AND handler_ok AND known_good AND floor_ok
        ORDER BY created_at ASC
        LIMIT 20
    ),
    held AS (
        -- What the doors refused. Flag-only rows (no handler_agent) are NOT here:
        -- they are excluded by `scored` itself, because `detected` is where they
        -- belong permanently and holding is not what is happening to them.
        SELECT item_type, handler_agent,
               CASE WHEN NOT handler_ok  THEN 'handler not a live agent'
                    WHEN NOT pipe_ok     THEN 'pipeline not in allow-list'
                    WHEN NOT known_good  THEN 'pair has never completed one (awaiting a hand canary)'
                    ELSE                      'pair below the 25% success floor'
               END AS reason
        FROM scored
        WHERE NOT (pipe_ok AND handler_ok AND known_good AND floor_ok)
    ),
    promoted AS (
        UPDATE site_work_items wi
        SET status = 'triaged',
            triaged_at = now(),
            spec = jsonb_set(COALESCE(wi.spec, '{}'::jsonb), '{original_pipeline}', to_jsonb(wi.pipeline)),
            pipeline = 'build',
            updated_at = now()
        FROM candidates c
        WHERE wi.id = c.id
          AND wi.status = 'detected'
        RETURNING wi.id, wi.item_type, wi.handler_agent
    )
    SELECT COUNT(*)::text AS promoted,
           string_agg(DISTINCT item_type || '->' || handler_agent, ', ') AS pairs,
           (SELECT COUNT(*)::text FROM held) AS held,
           (SELECT string_agg(DISTINCT item_type || '->' || handler_agent || ' (' || reason || ')', '; ')
              FROM held) AS held_detail
    FROM promoted
    -- Suppress ONLY a tick with nothing to say. A tick that promoted nothing but
    -- is holding rows still has something to say, and saying it is the point.
    HAVING COUNT(*) > 0 OR (SELECT COUNT(*) FROM held) > 0
$Q$,
    updated_at = now()
WHERE name = 'detected-item-promoter';

-- ============================================================================
-- Verification. RAISE, not SELECT — a plain SELECT cannot stop the COMMIT.
-- ============================================================================
DO $$
DECLARE
    q                text;
    n_promotable_old int;
    n_promotable_new int;
    n_held           int;
BEGIN
    SELECT pre_query INTO q FROM scheduled_tasks WHERE name = 'detected-item-promoter';

    -- the new shape is present
    IF q NOT LIKE '%HAVING COUNT(*) > 0 OR (SELECT COUNT(*) FROM held) > 0%' THEN
        RAISE EXCEPTION '458: the suppression clause is not in the live pre_query';
    END IF;
    IF q NOT LIKE '%AS held_detail%' THEN
        RAISE EXCEPTION '458: the held reporting columns are not in the live pre_query';
    END IF;
    -- the OLD non-suppressing shape must be GONE (a rewrite that left it would
    -- still "work" and would still be invisible)
    IF q LIKE '%WHERE (SELECT COUNT(*) FROM promoted) > 0%' THEN
        RAISE EXCEPTION '458: the old non-suppressing WHERE survived the rewrite';
    END IF;

    -- 430/444/454's load-bearing parts must have SURVIVED (negative control on a
    -- copy-paste error: dropping the cap, the stamp or a door would still "work")
    IF q NOT LIKE '%LIMIT 20%'
       OR q NOT LIKE '%original_pipeline%'
       OR q NOT LIKE '%pipeline = ''build''%'
       OR q NOT LIKE '%wi.pipeline IN (''build'', ''content'', ''design'')%'
       OR q NOT LIKE '%0.25 * (c + f)%'
       OR q NOT LIKE '%(c + f) < 5%'
       OR q NOT LIKE '%IN (''complete'', ''verified'')%'
       OR q NOT LIKE '%ad.is_active%' THEN
        RAISE EXCEPTION '458: the rewrite LOST one of 430/444/454''s predicates (cap / stamp / rewrite / pipeline door / floor / verified / active-handler)';
    END IF;

    -- BEHAVIOUR UNCHANGED: the promotable set must be identical under the old
    -- shape and the new one. This is the negative control on the restructure.
    SELECT count(*) INTO n_promotable_old
      FROM site_work_items wi
      WHERE wi.status = 'detected'
        AND COALESCE(wi.handler_agent,'') <> ''
        AND wi.pipeline IN ('build','content','design')
        AND EXISTS (SELECT 1 FROM agent_definitions ad WHERE ad.type = wi.handler_agent
                      AND ad.is_active AND COALESCE(ad.is_snapshot,false) = false AND ad.deleted_at IS NULL)
        AND EXISTS (SELECT 1 FROM site_work_items d WHERE d.item_type = wi.item_type
                      AND d.handler_agent = wi.handler_agent AND d.status IN ('complete','verified'))
        AND (SELECT (c + f) < 5 OR c >= 0.25 * (c + f)
               FROM (SELECT count(*) FILTER (WHERE h.status IN ('complete','verified')) AS c,
                            count(*) FILTER (WHERE h.status = 'failed')                 AS f
                       FROM site_work_items h
                      WHERE h.item_type = wi.item_type AND h.handler_agent = wi.handler_agent) hist);

    WITH scored AS (
        SELECT (wi.pipeline IN ('build','content','design')) AS pipe_ok,
               EXISTS (SELECT 1 FROM agent_definitions ad WHERE ad.type = wi.handler_agent
                         AND ad.is_active AND COALESCE(ad.is_snapshot,false) = false AND ad.deleted_at IS NULL) AS handler_ok,
               EXISTS (SELECT 1 FROM site_work_items d WHERE d.item_type = wi.item_type
                         AND d.handler_agent = wi.handler_agent AND d.status IN ('complete','verified')) AS known_good,
               (SELECT (c + f) < 5 OR c >= 0.25 * (c + f)
                  FROM (SELECT count(*) FILTER (WHERE h.status IN ('complete','verified')) AS c,
                               count(*) FILTER (WHERE h.status = 'failed')                 AS f
                          FROM site_work_items h
                         WHERE h.item_type = wi.item_type AND h.handler_agent = wi.handler_agent) hist) AS floor_ok
        FROM site_work_items wi
        WHERE wi.status = 'detected' AND COALESCE(wi.handler_agent,'') <> ''
    )
    SELECT count(*) FILTER (WHERE pipe_ok AND handler_ok AND known_good AND floor_ok),
           count(*) FILTER (WHERE NOT (pipe_ok AND handler_ok AND known_good AND floor_ok))
      INTO n_promotable_new, n_held
      FROM scored;

    IF n_promotable_old <> n_promotable_new THEN
        RAISE EXCEPTION '458: the restructure CHANGED the promotable set (% before, % after). It must be behaviour-identical.',
            n_promotable_old, n_promotable_new;
    END IF;

    -- POSITIVE CONTROL: the held counter must be capable of counting something,
    -- or "it reports held rows" is an untested claim. If this ever fails because
    -- the pile is genuinely empty, say so and re-run when it is not — do not
    -- delete the control.
    IF n_held = 0 THEN
        RAISE EXCEPTION '458: POSITIVE CONTROL FAILED — the held set is empty, so this migration cannot demonstrate that its counter counts anything. Expected >=1 (literal_markdown->page-build-handler was held by the floor at 2026-08-17 22:5xZ).';
    END IF;

    RAISE NOTICE '458: promoter now reports what it held. Promotable unchanged at % (negative control OK). Held = % and visible for the first time (positive control OK). Columns are inert until kafka-scheduler carries the logging change.',
        n_promotable_new, n_held;
END $$;

COMMIT;
