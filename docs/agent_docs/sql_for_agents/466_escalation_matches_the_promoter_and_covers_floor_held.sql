-- 466 — held-pair-canary-escalation: three corrections to my own migration 453
--
-- (a) its final WHERE does not suppress, so it claims a shared concurrency slot
--     on every idle tick;
-- (b) it escalates only CANARY-held pairs, so pairs held by 444/454's success
--     FLOOR are escalated by nothing and sit for ever;
-- (c) its held predicate is now OUT OF STEP with the promoter it describes, and
--     would tell a human to canary a pair the promoter has already promoted.
--
-- (a) was found and measured by the 458 author, who fixed the promoter and left
-- this task to its author. (b) I named in my own handoff. (c) I did not spot
-- until reading 453's live text today, and it is the worst of the three.
--
-- ============================================================================
-- (c) FIRST, BECAUSE IT IS A LIVE CONTRADICTION
-- ============================================================================
-- 453's held test asks `NOT EXISTS (… status = 'complete')` over
-- `site_work_items` ALONE. The promoter's equivalent test has since been
-- corrected twice — `454` added `verified` (a second terminal success status)
-- and `465` added `site_work_items_archive` (terminal rows older than 7 days
-- are MOVED there; the archive holds ~20k rows against ~8.7k live). 453 got
-- neither, so the two mechanisms now disagree about the same fact.
--
-- [MEASURED 2026-08-18] `empty_internal_href -> page-build-handler`:
--     what 453 sees (live `complete` only) ....... 0  -> "never completed, canary it"
--     what the promoter now sees (true) .......... 9  -> known-good, promote it
-- So 453 would escalate a request for a human canary on a pair with nine
-- lifetime successes that the promoter is promoting unattended. That is not a
-- cosmetic drift: the escalation's whole payload is an instruction to a person.
--
-- Both tests must be the same test. 466 makes 453's predicate identical in
-- scope to the promoter's: statuses `('complete','verified')`, population
-- `site_work_items UNION ALL site_work_items_archive`.
--
-- ============================================================================
-- (b) THE ESCALATION GAP — the disease this task exists to cure, one category over
-- ============================================================================
-- 453 requires ZERO successes, so it only ever sees pairs awaiting a first
-- canary. A pair with successes that is failing the 25% floor is held by the
-- promoter and escalated by NOTHING — no clock, no owner, no path out.
--
-- [MEASURED 2026-08-18] `literal_markdown -> page-build-handler`: 3 successes /
-- 36 failures = 8% (true, with archive), **10 rows** at `detected`, oldest
-- 08-17, growing. `bugs_open/184`'s consumer notice predicted the rows would
-- sit; nothing told anyone WHEN that had gone on too long. That is my design
-- gap, not a surprise.
--
-- The two cases need DIFFERENT words, because the remedies are opposite:
--   canary-held -> "nobody has ever watched this run; run one by hand."
--   floor-held  -> "this handler is failing; a canary would just add a failure.
--                   FIX THE HANDLER." Escalating a floor-held pair with canary
--                   instructions would be actively harmful — `bugs_open/300` is
--                   the case where a canary on a stale row would have written a
--                   `failed` and pushed the pair FURTHER under the floor.
--
-- ============================================================================
-- (a) THE NON-SUPPRESSING WHERE
-- ============================================================================
-- `… FROM escalated WHERE (SELECT COUNT(*) FROM escalated) > 0` returns exactly
-- one row regardless: the target list is aggregate-only with no GROUP BY.
-- Verified live this hour in the scheduler's own log —
-- `"task":"held-pair-canary-escalation","pre_query_result":"{\"escalated\":\"0\",\"pairs\":null}"`
-- on an idle tick. Consequences (458's header states them for the promoter and
-- they apply identically here): the `dynamicData == nil` branch is unreachable,
-- and the task claims its `maintenance` concurrency slot on every tick although
-- `max_concurrent=1` and `bugs_open/048`'s fix releases the slot by returning
-- zero rows. `HAVING` is the fix.
--
-- Following 458's subtlety rather than the naive form: suppress only when there
-- is nothing to say at all — `escalated = 0 AND watching = 0`. Suppressing on
-- `escalated` alone would go quiet on exactly the ticks where the interesting
-- state is "holding N pairs, none yet past the limit".
--
-- ============================================================================
-- SHAPE: ONE `classified` CTE, evaluated ONCE
-- ============================================================================
-- Deliberately NOT a second negated copy of the promoter's predicates. 458 made
-- that call for the same reason and it is the right one here: 444 added the
-- doors and 454 corrected them within four hours, and 465 corrected them again
-- the next day. A duplicated predicate would have drifted three times in three
-- days — and (c) above IS that drift, already realised.
--
-- Rollback: `466_..._ROLLBACK.sql` restores 453's pre_query verbatim.

BEGIN;

-- Guard: refuse unless the live text is 453's, so a concurrent edit ABORTS this
-- rather than being silently reverted.
DO $$
DECLARE q text;
BEGIN
    SELECT pre_query INTO q FROM scheduled_tasks WHERE name = 'held-pair-canary-escalation';
    IF q IS NULL THEN
        RAISE EXCEPTION '466: held-pair-canary-escalation has no pre_query — wrong database or the task is gone';
    END IF;
    IF q NOT LIKE '%WHERE (SELECT COUNT(*) FROM escalated) > 0%' THEN
        RAISE EXCEPTION '466: 453''s non-suppressing WHERE is absent — someone has already edited this task; re-read the live pre_query before applying';
    END IF;
    IF q NOT LIKE '%interval ''3 days''%' THEN
        RAISE EXCEPTION '466: 453''s 3-day limit is absent — this file rewrites 453''s shape and must not be applied to another';
    END IF;
END $$;

UPDATE scheduled_tasks
SET pre_query = $Q$
    WITH hist AS (
        -- ONE definition of "has this pair ever succeeded", identical in scope to
        -- the promoter's after 454 (verified counts) and 465 (the archive counts).
        SELECT item_type, handler_agent, status FROM site_work_items
        UNION ALL
        SELECT item_type, handler_agent, status FROM site_work_items_archive
    ),
    classified AS (
        SELECT wi.id, wi.item_type, wi.handler_agent, wi.created_at,
               h.c AS successes, h.f AS failures,
               CASE WHEN h.c = 0 THEN 'canary' ELSE 'floor' END AS hold_kind
        FROM site_work_items wi
        CROSS JOIN LATERAL (
            SELECT count(*) FILTER (WHERE status IN ('complete','verified')) AS c,
                   count(*) FILTER (WHERE status = 'failed')                AS f
            FROM hist
            WHERE hist.item_type = wi.item_type
              AND hist.handler_agent = wi.handler_agent
        ) h
        WHERE wi.status = 'detected'
          AND COALESCE(wi.handler_agent, '') <> ''
          AND EXISTS (
            SELECT 1 FROM agent_definitions ad
            WHERE ad.type = wi.handler_agent
              AND ad.is_active
              AND COALESCE(ad.is_snapshot, false) = false
              AND ad.deleted_at IS NULL
          )
          -- held by the promoter for a reason THIS task speaks to: either the
          -- pair has never succeeded (canary), or it has and is under the floor.
          -- A pair that is promotable is not held and must not appear here.
          AND (
            h.c = 0
            OR NOT ((h.c + h.f) < 5 OR h.c >= 0.25 * (h.c + h.f))
          )
    ),
    overdue AS (
        -- the clock runs on the PAIR, so a forgotten pair escalates together
        SELECT c.*, (now()::date - p.oldest::date) AS days_waiting
        FROM classified c
        JOIN (
            SELECT item_type, handler_agent, min(created_at) AS oldest
            FROM classified GROUP BY 1, 2
            HAVING min(created_at) < now() - interval '3 days'
        ) p ON p.item_type = c.item_type AND p.handler_agent = c.handler_agent
    ),
    owners (item_type, owner) AS (
        VALUES
          ('placeholder_contact',
           'bugs_open/201 lane — docs024_key_docs_latest/bugfix_201_page_content_writer_dispatch'),
          ('literal_markdown',
           'bugs_open/184 + bugs_open/201 lane — docs024_key_docs_latest/bugfix_201_page_content_writer_dispatch'),
          ('page_component_status_drift',
           '(UNASSIGNED - claim this) check_page_component_status_drift.go added 2026-07-10, never touched since, no lane doc claims it')
    ),
    escalated AS (
        UPDATE site_work_items wi
        SET status = 'needs_human_review',
            resolution_path = 'auto:held_pair_escalated',
            result = COALESCE(wi.result, '{}'::jsonb) || jsonb_build_object(
                'held_pair_escalation', jsonb_build_object(
                    'at',           to_char(now() AT TIME ZONE 'UTC', 'YYYY-MM-DD"T"HH24:MI:SS"Z"'),
                    'hold_kind',    o.hold_kind,
                    'days_waiting', o.days_waiting,
                    'limit_days',   3,
                    'pair',         o.item_type || ' -> ' || o.handler_agent,
                    'record',       o.successes || ' success(es) / ' || o.failures || ' failure(s), lifetime incl. archive',
                    'owner',        COALESCE(ow.owner, '(UNASSIGNED - claim this) no owner named for this item_type in migration 466'),
                    'reason',       CASE WHEN o.hold_kind = 'canary'
                        THEN 'held by the promoter''s known-good rule (SCH-026): this pair has never completed one, so nothing will dispatch it until a human runs one by hand'
                        ELSE 'held by the promoter''s SUCCESS FLOOR (migration 444/454): the pair succeeds below 25%, so the promoter has stopped feeding it' END,
                    'what_to_do',   CASE WHEN o.hold_kind = 'canary'
                        THEN 'Promote ONE row by hand and watch it: UPDATE site_work_items SET status=''triaged'', pipeline=''build'', triaged_at=now(), spec=jsonb_set(COALESCE(spec,''{}''::jsonb),''{original_pipeline}'',to_jsonb(pipeline)) WHERE id=''<one id>''. If it completes the pair becomes known-good and the promoter takes the rest. If it fails, that IS the finding — file it. FIRST check the row is not stale (bugs_open/300): a canary on a row whose page_components.id no longer exists hard-errors and writes a failure against the pair.'
                        ELSE 'Do NOT canary this one — a canary would add another failure and push the pair further under the floor (bugs_open/300 is that case). FIX THE HANDLER, or decide the pair is wrong and retire the producer. The rows are safe at detected meanwhile; they are not lost and carry no error.' END,
                    'escalated_by', 'held-pair-canary-escalation (migration 466, supersedes 453)'
                )
            ),
            updated_at = now()
        FROM overdue o
        LEFT JOIN owners ow ON ow.item_type = o.item_type
        WHERE wi.id = o.id
          AND wi.status = 'detected'
        RETURNING wi.id, wi.item_type, wi.handler_agent, o.hold_kind
    )
    SELECT COUNT(*)::text AS escalated,
           string_agg(DISTINCT item_type || '->' || handler_agent || ' (' || hold_kind || ')', ', ') AS pairs,
           (SELECT COUNT(*)::text FROM classified) AS watching,
           (SELECT string_agg(DISTINCT item_type || '->' || handler_agent || ' (' || hold_kind
                              || ', day ' || (now()::date - created_at::date) || ' of 3)', ', ')
              FROM classified) AS watching_detail
    FROM escalated
    -- HAVING, not WHERE: an aggregate-only target list with no GROUP BY returns
    -- one row under a WHERE regardless, which claims the concurrency slot on
    -- every idle tick. Suppress only when there is nothing to say AT ALL —
    -- "watching N, none overdue" is worth saying.
    HAVING COUNT(*) > 0 OR (SELECT COUNT(*) FROM classified) > 0
$Q$,
    description = 'Owner decision 2026-08-17 (bugs_open/083), corrected by 466. A pair the detected-item-promoter is holding — either never-completed (canary) or under the 444/454 success floor — for more than 3 days is escalated from detected to needs_human_review, carrying its named owner, its true lifetime record incl. archive, and remedy text that DIFFERS by hold kind. The owner map enriches; it never gates.',
    updated_at = now()
WHERE name = 'held-pair-canary-escalation';

-- ============================================================================
-- Verification. RAISE, not SELECT.
-- ============================================================================
DO $$
DECLARE
    q            text;
    n_watch      int;
    n_floor      int;
    n_canary     int;
    n_promotable int;
BEGIN
    SELECT pre_query INTO q FROM scheduled_tasks WHERE name = 'held-pair-canary-escalation';

    IF q LIKE '%WHERE (SELECT COUNT(*) FROM escalated) > 0%' THEN
        RAISE EXCEPTION '466: defect (a) survives — the non-suppressing WHERE is still there';
    END IF;
    IF q NOT LIKE '%HAVING COUNT(*) > 0%' THEN
        RAISE EXCEPTION '466: defect (a) not fixed — no HAVING in the final SELECT';
    END IF;
    IF q NOT LIKE '%site_work_items_archive%' OR q NOT LIKE '%IN (''complete'',''verified'')%' THEN
        RAISE EXCEPTION '466: defect (c) not fixed — the held test is still narrower than the promoter''s';
    END IF;
    IF q NOT LIKE '%hold_kind%' OR q NOT LIKE '%0.25 * (h.c + h.f)%' THEN
        RAISE EXCEPTION '466: defect (b) not fixed — floor-held pairs are still invisible to this task';
    END IF;
    IF q NOT LIKE '%interval ''3 days''%' THEN
        RAISE EXCEPTION '466: the 3-day limit was lost in the rewrite';
    END IF;

    -- The watched set, by kind, from the same definitions the task now uses.
    CREATE TEMP TABLE _h466 ON COMMIT DROP AS
      SELECT item_type, handler_agent, status FROM site_work_items
      UNION ALL SELECT item_type, handler_agent, status FROM site_work_items_archive;

    SELECT count(*) FILTER (WHERE h.c = 0),
           count(*) FILTER (WHERE h.c > 0),
           count(*)
      INTO n_canary, n_floor, n_watch
      FROM site_work_items wi
      CROSS JOIN LATERAL (
        SELECT count(*) FILTER (WHERE status IN ('complete','verified')) AS c,
               count(*) FILTER (WHERE status='failed') AS f
        FROM _h466 WHERE _h466.item_type=wi.item_type AND _h466.handler_agent=wi.handler_agent) h
      WHERE wi.status='detected' AND COALESCE(wi.handler_agent,'')<>''
        AND EXISTS (SELECT 1 FROM agent_definitions ad WHERE ad.type=wi.handler_agent AND ad.is_active
                      AND COALESCE(ad.is_snapshot,false)=false AND ad.deleted_at IS NULL)
        AND (h.c = 0 OR NOT ((h.c + h.f) < 5 OR h.c >= 0.25 * (h.c + h.f)));

    -- POSITIVE CONTROL for (b): the floor-held set must be NON-EMPTY, or the
    -- whole point of this migration is untested. literal_markdown is the case.
    IF n_floor = 0 THEN
        RAISE EXCEPTION '466: POSITIVE CONTROL FAILED — no floor-held rows are visible, so defect (b)''s fix cannot be shown to do anything. Expected literal_markdown->page-build-handler (10 rows on 2026-08-18). Re-measure before applying.';
    END IF;

    -- NEGATIVE CONTROL for (c), written to DISCRIMINATE old behaviour from new
    -- rather than to be true by construction. (My first draft asserted
    -- `promotable AND NOT promotable = 0`, i.e. P AND NOT P — always zero, no
    -- matter what the code does. That is the third tautological control this
    -- lane has written; the check for one is "could this ever have been
    -- non-zero?".)
    --
    -- The discriminating case is a named pair: `empty_internal_href ->
    -- page-build-handler` has 0 live `complete` rows but NINE successes once
    -- `verified` and the archive are counted (measured 2026-08-18). So:
    --   under 453's predicate it IS watched  (live-only complete = 0 -> "canary")
    --   under 466's predicate it is NOT      (9 successes at 64% -> promotable)
    -- A non-zero count here therefore means the rewrite did not take effect.
    SELECT count(*) INTO n_promotable
      FROM site_work_items wi
      CROSS JOIN LATERAL (
        SELECT count(*) FILTER (WHERE status IN ('complete','verified')) AS c,
               count(*) FILTER (WHERE status='failed') AS f
        FROM _h466 WHERE _h466.item_type=wi.item_type AND _h466.handler_agent=wi.handler_agent) h
      WHERE wi.status='detected'
        AND wi.item_type='empty_internal_href' AND wi.handler_agent='page-build-handler'
        AND (h.c = 0 OR NOT ((h.c + h.f) < 5 OR h.c >= 0.25 * (h.c + h.f)));
    IF n_promotable <> 0 THEN
        RAISE EXCEPTION '466: NEGATIVE CONTROL FAILED — empty_internal_href->page-build-handler is still in the watched set (% row(s)). It has 9 lifetime successes at 64%%, so this task is still using 453''s narrower history and would ask a human to canary a pair the promoter promotes.', n_promotable;
    END IF;

    RAISE NOTICE '466: escalation now matches the promoter and covers BOTH hold kinds. Watching % row(s): % canary-held, % floor-held. Controls: floor-held set is non-empty (so (b) is exercised) and no promotable row is watched (so (c)''s two tests agree).',
        n_watch, n_canary, n_floor;
END $$;

COMMIT;
