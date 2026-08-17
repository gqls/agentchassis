-- 444 — detected-item-promoter: two door-closers on the candidates CTE
--       (bugs_open/083 candidate 2, council round 1 corr 05a3d1c8 = REVISE)
--
-- 430 is ledger-recorded and MUST NOT be edited. This file rewrites the live
-- `scheduled_tasks.detected-item-promoter` pre_query with two added predicates
-- and nothing else. Everything not named below is byte-identical to 430's.
--
-- ============================================================================
-- DOOR-CLOSER 1 — the promoter may only promote pipelines it has ever promoted
-- ============================================================================
-- Council round 1 objection (editquality edit 1, MEDIUM; guardian edit 1,
-- MEDIUM): the pre_query rewrites `pipeline='build'` unconditionally, so a
-- detected row whose natural pipeline is `diagnose` or `report` would be
-- silently rerouted out of its own dispatch loop and into the build gate.
--
-- [MEASURED 2026-08-17] The objection has never fired, and the mechanism it
-- fears is real:
--   * lifetime promotions (every row ever carrying spec.original_pipeline,
--     BOTH implementations, n=1562): build 1055 / design 310 / content 197.
--     ZERO diagnose, ZERO report, ZERO anything else.
--   * pipelines that EXIST fleet-wide (all statuses, n=8421): build 7749,
--     content 490, design 130, diagnose 44, experience 7, maintenance 1.
--   * `diagnose-pipeline-trigger` is an enabled scheduled task, so `diagnose`
--     is a live loop with its own gate — the objection's premise is sound.
--
-- WHY AN ALLOW-LIST, NOT THE DENY-LIST THE OBJECTION IMPLIES. The obvious
-- patch is `AND wi.pipeline NOT IN ('diagnose','report')`. The measurement
-- above refutes it as the better shape: `report` is not a pipeline value that
-- exists on this table at all (0 rows), while `experience` (7) and
-- `maintenance` (1) DO exist and a deny-list would still rewrite them to
-- 'build'. So the deny-list names one value that cannot fire and misses two
-- that can. The allow-list names the three values the transition has actually
-- ever produced, and a new pipeline value is HELD rather than rerouted — which
-- is the same philosophy as the existing known-good rule (hold the unknown,
-- let a human look). `pipeline` is NOT NULL DEFAULT 'build', so there is no
-- NULL hole in an IN-list.
--
-- Fires on 0 rows today: nothing at `detected` carries a non-allow-listed
-- pipeline. This is a door, not a repair.
--
-- ============================================================================
-- DOOR-CLOSER 2 — a pair that has stopped working is no longer "known-good"
-- ============================================================================
-- NOT a council objection. This is 430's OWN submitted risk 2 — "the
-- known-good rule trusts ONE lifetime complete per pair; a pair whose only
-- complete was a false success qualifies" — which has now FIRED, measured.
--
-- [MEASURED 2026-08-17] Of 85 rows this promoter has promoted since it went
-- live, 19 ended `failed`. The counterfactual, computing each pair's record AS
-- IT STOOD at the instant of its promotion:
--     item_type / handler                  complete/failed at promo   promoted  failed
--     literal_markdown  / page-build-handler        1 / 28  =  3%        6        5
--     audit_tool        / tool-auditor             18 / 21  = 46%       12        8
--     phantom_internal_link / page-build-handler    7 /  8  = 47%        2        2
--     (every other pair)                                   >= 60%
-- `literal_markdown -> page-build-handler` cleared the gate on ONE lifetime
-- success against 28 failures, and the promoter then fed it 6 more, 5 of which
-- failed. That is the rule admitting a pair it should hold.
--
-- THE THRESHOLD IS SET BY THE CENSUS, NOT CHOSEN. Success ratios across all
-- 28 pairs holding >=1 lifetime complete: 3%, then 41, 42, 46, 50, 67, 79, 80,
-- 86, 86, 88, 89, 94, 96, 98, 99, 100... A floor anywhere in 10%-35% isolates
-- exactly the one pathological pair and touches nothing else; 25% sits in the
-- middle of that gap. The 41-47% pairs are a SEPARATE question (are those
-- handlers defective?) and are deliberately NOT bundled here.
--
-- MINIMUM SAMPLE, so the canary path is untouched: the floor applies only once
-- a pair has >=5 terminal outcomes. Below that the existing ">=1 complete"
-- rule stands alone, so a brand-new pair is not held by one unlucky failure.
-- Checked against live data: literal_markdown (29 terminals) HELD;
-- audit_tool (52), phantom_internal_link (17), empty_section (24) pass;
-- empty_internal_href (2 terminals) passes on the small-sample exemption.
-- Fires on 0 rows today — no `literal_markdown` row is currently `detected`.
--
-- LANDMINE THIS FILE WAS WRITTEN AROUND: `failed` rows carry NO `completed_at`
-- (0 of 265, against 5882 of 5921 `complete` rows). A pair-health query that
-- keys on `completed_at` therefore counts ZERO failures for every pair and
-- reports 100% success across the board — it cannot come out any other way.
-- Both predicates below key on `status` alone; the counterfactual above was
-- re-run on `updated_at` after the first attempt returned exactly that
-- all-100% result. See LANDMINES.md.
--
-- ============================================================================
-- ORDER / ROLLBACK
-- ============================================================================
-- DB config: live at the scheduler's next tick after COMMIT. No binary
-- dependency. Rollback is a separate file (444_..._ROLLBACK.sql), which
-- restores 430's pre_query verbatim — asked for by the council's
-- debug_historian seat (edit 1, LOW: "a separate rollback file alongside a
-- separate verify file, as the needle-gate discipline asks").
--
-- Interaction with the concurrent bugs_open/284 lane, checked before writing:
-- migration 443 added CHECK swi_no_handlerless_promotable — NOT(handler_agent
-- = '' AND status IN ('triaged','approved','claimed')). This promoter already
-- requires a non-empty handler with a live agent_definitions row, so the
-- constraint is a hard backstop BEHIND that predicate, not a conflict: a
-- future promoter bug that promoted a handler-less row would now abort the
-- tick loudly instead of misrouting silently.

BEGIN;

UPDATE scheduled_tasks
SET pre_query = $Q$
    WITH candidates AS (
        SELECT wi.id
        FROM site_work_items wi
        WHERE wi.status = 'detected'
          AND COALESCE(wi.handler_agent, '') <> ''
          -- DOOR-CLOSER 1: only pipelines this transition has ever produced.
          AND wi.pipeline IN ('build', 'content', 'design')
          AND EXISTS (
            SELECT 1 FROM agent_definitions ad
            WHERE ad.type = wi.handler_agent
              AND ad.is_active
              AND COALESCE(ad.is_snapshot, false) = false
              AND ad.deleted_at IS NULL
          )
          AND EXISTS (
            SELECT 1 FROM site_work_items done
            WHERE done.item_type = wi.item_type
              AND done.handler_agent = wi.handler_agent
              AND done.status = 'complete'
          )
          -- DOOR-CLOSER 2: once a pair has a real sample (>=5 terminal
          -- outcomes), it must still be succeeding at >=25%. NB `failed` rows
          -- have no completed_at, so this keys on status only.
          AND (
            SELECT (c + f) < 5 OR c >= 0.25 * (c + f)
            FROM (
              SELECT count(*) FILTER (WHERE h.status = 'complete') AS c,
                     count(*) FILTER (WHERE h.status = 'failed')   AS f
              FROM site_work_items h
              WHERE h.item_type = wi.item_type
                AND h.handler_agent = wi.handler_agent
            ) hist
          )
        ORDER BY wi.created_at ASC
        LIMIT 20
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
           string_agg(DISTINCT item_type || '->' || handler_agent, ', ') AS pairs
    FROM promoted
    WHERE (SELECT COUNT(*) FROM promoted) > 0
$Q$,
    updated_at = now()
WHERE name = 'detected-item-promoter';

-- ============================================================================
-- Verification. RAISE, not SELECT — a plain SELECT cannot stop the COMMIT.
-- Every assert below is disconfirmable: each has a stated value that a wrong
-- edit would change.
-- ============================================================================
DO $$
DECLARE
    n_rows            int;
    q                 text;
    n_before_holdable int;
    n_after_holdable  int;
    n_pipeline_held   int;
    n_ratio_held      int;
BEGIN
    SELECT count(*) INTO n_rows FROM scheduled_tasks WHERE name = 'detected-item-promoter';
    IF n_rows <> 1 THEN
        RAISE EXCEPTION '444: expected exactly 1 detected-item-promoter row, found %', n_rows;
    END IF;

    SELECT pre_query INTO q FROM scheduled_tasks WHERE name = 'detected-item-promoter';

    -- both new predicates present
    IF q NOT LIKE '%wi.pipeline IN (''build'', ''content'', ''design'')%' THEN
        RAISE EXCEPTION '444: door-closer 1 (pipeline allow-list) is not in the live pre_query';
    END IF;
    IF q NOT LIKE '%0.25 * (c + f)%' OR q NOT LIKE '%(c + f) < 5%' THEN
        RAISE EXCEPTION '444: door-closer 2 (success floor + small-sample exemption) is not in the live pre_query';
    END IF;

    -- 430's load-bearing parts must have SURVIVED the rewrite (negative control
    -- on a copy-paste error: if this file dropped the cap, the stamp or the
    -- known-good test, the promoter would still "work" and silently misbehave)
    IF q NOT LIKE '%LIMIT 20%'
       OR q NOT LIKE '%original_pipeline%'
       OR q NOT LIKE '%pipeline = ''build''%'
       OR q NOT LIKE '%done.status = ''complete''%'
       OR q NOT LIKE '%ad.is_active%' THEN
        RAISE EXCEPTION '444: the rewrite LOST one of 430''s predicates (cap / stamp / rewrite / known-good / active-handler)';
    END IF;

    -- Effect, measured against the live pile. Both door-closers must hold ZERO
    -- rows today: this is a door, not a repair. If either holds something, the
    -- predicate is wrong or the pile has changed since 2026-08-17 11:00Z and a
    -- human should look before this commits.
    SELECT count(*) INTO n_before_holdable
      FROM site_work_items wi
      WHERE wi.status = 'detected'
        AND COALESCE(wi.handler_agent,'') <> ''
        AND EXISTS (SELECT 1 FROM agent_definitions ad WHERE ad.type = wi.handler_agent
                      AND ad.is_active AND COALESCE(ad.is_snapshot,false) = false AND ad.deleted_at IS NULL)
        AND EXISTS (SELECT 1 FROM site_work_items d WHERE d.item_type = wi.item_type
                      AND d.handler_agent = wi.handler_agent AND d.status = 'complete');

    SELECT count(*) INTO n_pipeline_held
      FROM site_work_items wi
      WHERE wi.status = 'detected'
        AND COALESCE(wi.handler_agent,'') <> ''
        AND wi.pipeline NOT IN ('build','content','design')
        AND EXISTS (SELECT 1 FROM agent_definitions ad WHERE ad.type = wi.handler_agent
                      AND ad.is_active AND COALESCE(ad.is_snapshot,false) = false AND ad.deleted_at IS NULL)
        AND EXISTS (SELECT 1 FROM site_work_items d WHERE d.item_type = wi.item_type
                      AND d.handler_agent = wi.handler_agent AND d.status = 'complete');

    SELECT count(*) INTO n_ratio_held
      FROM site_work_items wi
      WHERE wi.status = 'detected'
        AND COALESCE(wi.handler_agent,'') <> ''
        AND wi.pipeline IN ('build','content','design')
        AND EXISTS (SELECT 1 FROM agent_definitions ad WHERE ad.type = wi.handler_agent
                      AND ad.is_active AND COALESCE(ad.is_snapshot,false) = false AND ad.deleted_at IS NULL)
        AND EXISTS (SELECT 1 FROM site_work_items d WHERE d.item_type = wi.item_type
                      AND d.handler_agent = wi.handler_agent AND d.status = 'complete')
        AND NOT (SELECT (c + f) < 5 OR c >= 0.25 * (c + f)
                   FROM (SELECT count(*) FILTER (WHERE h.status='complete') AS c,
                                count(*) FILTER (WHERE h.status='failed')   AS f
                           FROM site_work_items h
                          WHERE h.item_type = wi.item_type AND h.handler_agent = wi.handler_agent) hist);

    n_after_holdable := n_before_holdable - n_pipeline_held - n_ratio_held;

    IF n_pipeline_held <> 0 OR n_ratio_held <> 0 THEN
        RAISE EXCEPTION '444: a door-closer would HOLD live rows (pipeline % / ratio %) — expected 0 and 0. Read the rows before applying.',
            n_pipeline_held, n_ratio_held;
    END IF;

    -- POSITIVE CONTROL: the ratio predicate must be capable of holding
    -- something, or the assert above is vacuous. literal_markdown ->
    -- page-build-handler is the pair the counterfactual named; it must fail the
    -- floor when evaluated directly. If this ever passes, the pair has
    -- recovered (fine — say so) or the predicate is inert (not fine).
    IF (SELECT (c + f) < 5 OR c >= 0.25 * (c + f)
          FROM (SELECT count(*) FILTER (WHERE h.status='complete') AS c,
                       count(*) FILTER (WHERE h.status='failed')   AS f
                  FROM site_work_items h
                 WHERE h.item_type = 'literal_markdown' AND h.handler_agent = 'page-build-handler') hist) THEN
        RAISE EXCEPTION '444: POSITIVE CONTROL FAILED — literal_markdown->page-build-handler passes the 25%% floor, so the predicate cannot be shown to hold anything. Re-measure before applying.';
    END IF;

    RAISE NOTICE '444: detected-item-promoter door-closers live. Promotable pile unchanged at % (pipeline door held %, ratio door held %). Positive control OK: literal_markdown->page-build-handler is held by the floor.',
        n_after_holdable, n_pipeline_held, n_ratio_held;
END $$;

COMMIT;
