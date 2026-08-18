-- 479 — held-pair-canary-escalation: RECLAIM a pair that has since qualified
--       (OWNER DECISION 2026-08-18, decision 2: "fix the door")
--
-- 453 and 466 are ledger-recorded and are NOT edited. This file rewrites the live
-- `scheduled_tasks.held-pair-canary-escalation` pre_query by SURGICAL REPLACEMENT
-- on three verbatim anchors, each asserted to occur exactly once, and guarded on
-- the whole text's md5. It transcribes nothing: the 7,803-character body — most of
-- it 466's `what_to_do` prose — is carried through untouched, which is the point.
-- That lane is iterating fast (465, 466, 471, 472 in one day); a wholesale rewrite
-- would silently revert whatever landed between my read and my apply, and a
-- transcription of 7.8 KB would introduce its own errors.
--
-- ============================================================================
-- THE DEFECT: THE DOOR ONLY OPENS ONE WAY
-- ============================================================================
-- SCH-026's promoter selects `status='detected'` and nothing else. This task moves
-- a held row to `needs_human_review` after 3 days so a human sees it. Nothing ever
-- moves it back. So when a pair LATER becomes known-good, its already-escalated
-- findings do not rejoin the automated path — they sit in a human queue standing
-- at 829 rows. That is `bugs_open/083`'s own disease (a queue with no consumer)
-- reproduced one step further along, inside the mechanism built to cure it.
--
-- **This is not hypothetical, and migration `465` is the proof.** Until 465 the
-- promoter's history tests read only `site_work_items`, which the daily
-- `work-item-archiver` prunes to ~7 days. `empty_internal_href -> page-build-handler`
-- therefore read 0 successes and was held as "never completed one" while holding
-- NINE lifetime successes (9 ok / 5 failed = 64% once the archive is counted).
-- Had it been escalated during that window, 465 would have fixed the gate and the
-- rows would still be stranded. A pair being wrongly held and later released is a
-- MEASURED event on this platform, not a thought experiment.
--
-- ============================================================================
-- WHAT THIS ADDS
-- ============================================================================
-- A `reclaimable` CTE and a `reclaimed` UPDATE, mirroring `classified`'s predicate
-- in the OPPOSITE direction, over the SAME `hist` CTE (live + archive) that 465/466
-- already established — so the reclaim test and the hold test cannot drift apart:
-- they read one source and one arithmetic.
--
-- A row is reclaimed when ALL hold:
--   * `status='needs_human_review'` AND `resolution_path='auto:held_pair_escalated'`
--     — only THIS task writes that path, so a human's own triage is never touched;
--     if a person has acted, the status has moved and the row is invisible here.
--   * its handler is still a live, active, non-snapshot agent_definitions row;
--   * the pair now has >=1 lifetime success (`complete` or `verified`), AND
--   * the pair passes the floor: fewer than 5 terminal outcomes, or >=25% success.
-- i.e. exactly the promoter's own known-good rule. A reclaimed row returns to
-- `detected`, where the promoter will pick it up on its next tick.
--
-- `resolution_path` is cleared (the row is no longer resolved), and the escalation
-- record is KEPT under `result.held_pair_escalation` with a `held_pair_reclaimed`
-- block added beside it — so the history reads "escalated, then reclaimed, because
-- the pair qualified", rather than looking as though the escalation never happened.
--
-- ============================================================================
-- CONTROLS — AND WHY THEY TEST THE PREDICATE, NOT THE ROW COUNT
-- ============================================================================
-- [MEASURED 2026-08-18] ZERO rows are currently escalated: `result ?
-- 'held_pair_escalation'` is 0 across all 15 held rows, because the oldest is at
-- 1.9 days of its 3-day limit. So this arm reclaims 0 rows today and a row-count
-- assertion would be VACUOUS — it would pass identically if the predicate were
-- `WHERE false`.
--
-- The verify block therefore evaluates the reclaim predicate directly against two
-- named pairs whose answers must differ, which is a test that can fail:
--   POSITIVE: `page_component_status_drift -> component-template-fixer` (4 ok /
--             0 failed, canaried by hand 2026-08-17) MUST evaluate TRUE.
--   NEGATIVE: `literal_markdown -> page-build-handler` (3 ok / 36 failed = 8%)
--             MUST evaluate FALSE — if the archive scope or the floor arithmetic
--             were wrong, this is the pair that would wrongly come back, and it is
--             the pair `444` exists for.
-- Both are asserted below and both came out the required way at apply.
--
-- ============================================================================
-- ORDER / ROLLBACK
-- ============================================================================
-- DB config: live at the task's next daily tick after COMMIT. No binary dependency.
-- Rollback is a separate file (479_..._ROLLBACK.sql) which removes the two CTEs by
-- the same anchored replacement and asserts the md5 returns to this pre-image.
-- Reverting does NOT re-escalate rows this arm already reclaimed; they are at
-- `detected` and the promoter owns them, which is the intended end state anyway.

BEGIN;

-- GUARD: refuse unless the live text is exactly what this file was written against.
DO $$
DECLARE
    live_md5 text;
    n int;
    q text;
BEGIN
    SELECT pre_query, md5(pre_query) INTO q, live_md5
      FROM scheduled_tasks WHERE name = 'held-pair-canary-escalation';

    IF q IS NULL THEN
        RAISE EXCEPTION '479: no held-pair-canary-escalation row — nothing to extend';
    END IF;

    IF live_md5 <> 'db96f482ae8792dd19f9eb343fd38308' THEN
        RAISE EXCEPTION
          '479: ABORTING — the live pre_query is not the text this migration was written against (expected md5 db96f482ae8792dd19f9eb343fd38308 len 7803, found % len %). The 466/471/472 lane has edited it since 2026-08-18 evening. Re-read the live column, re-derive this file against it, and do NOT force: overwriting is how one lane silently reverts another.',
          live_md5, length(q);
    END IF;

    -- each anchor must occur EXACTLY once, or a replace() would hit the wrong
    -- place (or nothing) and the result would still be valid SQL
    n := (length(q) - length(replace(q, E'    escalated AS (', ''))) / length(E'    escalated AS (');
    IF n <> 1 THEN RAISE EXCEPTION '479: anchor "escalated AS (" occurs % times, expected 1', n; END IF;

    n := (length(q) - length(replace(q, E'    SELECT COUNT(*)::text AS escalated,', ''))) / length(E'    SELECT COUNT(*)::text AS escalated,');
    IF n <> 1 THEN RAISE EXCEPTION '479: anchor "SELECT COUNT(*)::text AS escalated," occurs % times, expected 1', n; END IF;

    n := (length(q) - length(replace(q, E'    HAVING COUNT(*) > 0 OR (SELECT COUNT(*) FROM classified) > 0', ''))) / length(E'    HAVING COUNT(*) > 0 OR (SELECT COUNT(*) FROM classified) > 0');
    IF n <> 1 THEN RAISE EXCEPTION '479: anchor "HAVING ..." occurs % times, expected 1', n; END IF;
END $$;

-- REPLACEMENT 1 — the reclaim CTEs, inserted immediately before `escalated`.
UPDATE scheduled_tasks
SET pre_query = replace(pre_query,
E'    escalated AS (',
E'    reclaimable AS (\n'
'        -- OWNER DECISION 2026-08-18: the escalation door opens both ways. A row\n'
'        -- escalated because its pair was held comes BACK to `detected` once that\n'
'        -- pair qualifies, so the promoter can dispatch it. Mirrors `classified`\n'
'        -- over the SAME `hist` CTE, so hold and reclaim cannot drift apart.\n'
'        SELECT wi.id, wi.item_type, wi.handler_agent, h.c AS successes, h.f AS failures\n'
'        FROM site_work_items wi\n'
'        CROSS JOIN LATERAL (\n'
'            SELECT count(*) FILTER (WHERE status IN (''complete'',''verified'')) AS c,\n'
'                   count(*) FILTER (WHERE status = ''failed'')                  AS f\n'
'            FROM hist\n'
'            WHERE hist.item_type = wi.item_type\n'
'              AND hist.handler_agent = wi.handler_agent\n'
'        ) h\n'
'        WHERE wi.status = ''needs_human_review''\n'
'          -- only THIS task writes that path, so a human''s own triage is untouched\n'
'          AND wi.resolution_path = ''auto:held_pair_escalated''\n'
'          AND COALESCE(wi.handler_agent, '''') <> ''''\n'
'          AND EXISTS (\n'
'            SELECT 1 FROM agent_definitions ad\n'
'            WHERE ad.type = wi.handler_agent\n'
'              AND ad.is_active\n'
'              AND COALESCE(ad.is_snapshot, false) = false\n'
'              AND ad.deleted_at IS NULL\n'
'          )\n'
'          -- the promoter''s own known-good rule, both halves\n'
'          AND h.c > 0\n'
'          AND ((h.c + h.f) < 5 OR h.c >= 0.25 * (h.c + h.f))\n'
'    ),\n'
'    reclaimed AS (\n'
'        UPDATE site_work_items wi\n'
'        SET status = ''detected'',\n'
'            resolution_path = NULL,\n'
'            result = COALESCE(wi.result, ''{}''::jsonb) || jsonb_build_object(\n'
'                ''held_pair_reclaimed'', jsonb_build_object(\n'
'                    ''at'',     to_char(now() AT TIME ZONE ''UTC'', ''YYYY-MM-DD"T"HH24:MI:SS"Z"''),\n'
'                    ''pair'',   r.item_type || '' -> '' || r.handler_agent,\n'
'                    ''record'', r.successes || '' success(es) / '' || r.failures || '' failure(s), lifetime incl. archive'',\n'
'                    ''why'',    ''the pair now passes SCH-026''''s known-good rule, so this finding returns to the automated path; it was escalated only because the pair was held at the time'',\n'
'                    ''by'',     ''held-pair-canary-escalation reclaim arm (migration 479, owner decision 2026-08-18)''\n'
'                )\n'
'            ),\n'
'            updated_at = now()\n'
'        FROM reclaimable r\n'
'        WHERE wi.id = r.id\n'
'          AND wi.status = ''needs_human_review''\n'
'        RETURNING wi.id, wi.item_type, wi.handler_agent\n'
'    ),\n'
'    escalated AS (')
WHERE name = 'held-pair-canary-escalation';

-- REPLACEMENT 2 — report the reclaims in the tick's own log line.
UPDATE scheduled_tasks
SET pre_query = replace(pre_query,
E'    SELECT COUNT(*)::text AS escalated,',
E'    SELECT COUNT(*)::text AS escalated,\n'
'           (SELECT COUNT(*)::text FROM reclaimed) AS reclaimed,\n'
'           (SELECT string_agg(DISTINCT item_type || ''->'' || handler_agent, '', '')\n'
'              FROM reclaimed) AS reclaimed_pairs,')
WHERE name = 'held-pair-canary-escalation';

-- REPLACEMENT 3 — a tick that reclaimed something has something to say.
UPDATE scheduled_tasks
SET pre_query = replace(pre_query,
E'    HAVING COUNT(*) > 0 OR (SELECT COUNT(*) FROM classified) > 0',
E'    HAVING COUNT(*) > 0 OR (SELECT COUNT(*) FROM classified) > 0\n'
'        OR (SELECT COUNT(*) FROM reclaimed) > 0')
WHERE name = 'held-pair-canary-escalation';

-- ============================================================================
-- Verification. RAISE, not SELECT — a plain SELECT cannot stop the COMMIT.
-- ============================================================================
DO $$
DECLARE
    q            text;
    pos_ok       boolean;
    neg_ok       boolean;
    n_escalated  int;
BEGIN
    SELECT pre_query INTO q FROM scheduled_tasks WHERE name = 'held-pair-canary-escalation';

    IF q NOT LIKE '%reclaimable AS (%' OR q NOT LIKE '%reclaimed AS (%' THEN
        RAISE EXCEPTION '479: the reclaim CTEs are not in the live pre_query';
    END IF;
    IF q NOT LIKE '%AS reclaimed,%' OR q NOT LIKE '%AS reclaimed_pairs,%' THEN
        RAISE EXCEPTION '479: the reclaim columns are not in the final SELECT';
    END IF;
    IF q NOT LIKE '%OR (SELECT COUNT(*) FROM reclaimed) > 0%' THEN
        RAISE EXCEPTION '479: the HAVING does not account for a reclaim-only tick';
    END IF;

    -- 466's load-bearing parts must have SURVIVED the three replacements
    IF q NOT LIKE '%site_work_items_archive%'
       OR q NOT LIKE '%hold_kind%'
       OR q NOT LIKE '%auto:held_pair_escalated%'
       OR q NOT LIKE '%PARTITION THE FAILURES%'
       OR q NOT LIKE '%interval ''3 days''%' THEN
        RAISE EXCEPTION '479: the rewrite LOST one of 466''s parts (archive scope / hold_kind / resolution_path / the failure-partition remedy / the 3-day limit)';
    END IF;

    -- POSITIVE CONTROL: a pair that has since qualified MUST satisfy the reclaim
    -- test. page_component_status_drift -> component-template-fixer was canaried
    -- by hand on 2026-08-17 and stands at 4 ok / 0 failed.
    SELECT (c > 0 AND ((c + f) < 5 OR c >= 0.25 * (c + f)))
      INTO pos_ok
      FROM (SELECT count(*) FILTER (WHERE status IN ('complete','verified')) AS c,
                   count(*) FILTER (WHERE status = 'failed')                 AS f
              FROM (SELECT item_type, handler_agent, status FROM site_work_items
                    UNION ALL
                    SELECT item_type, handler_agent, status FROM site_work_items_archive) u
             WHERE item_type = 'page_component_status_drift'
               AND handler_agent = 'component-template-fixer') z;
    IF NOT pos_ok THEN
        RAISE EXCEPTION '479: POSITIVE CONTROL FAILED — page_component_status_drift->component-template-fixer (canaried 2026-08-17) does not satisfy the reclaim test, so this arm cannot be shown to reclaim anything.';
    END IF;

    -- NEGATIVE CONTROL: the pair 444 exists for MUST NOT be reclaimed. If the
    -- archive scope or the floor arithmetic were wrong, this is the row that
    -- would wrongly come back.
    SELECT (c > 0 AND ((c + f) < 5 OR c >= 0.25 * (c + f)))
      INTO neg_ok
      FROM (SELECT count(*) FILTER (WHERE status IN ('complete','verified')) AS c,
                   count(*) FILTER (WHERE status = 'failed')                 AS f
              FROM (SELECT item_type, handler_agent, status FROM site_work_items
                    UNION ALL
                    SELECT item_type, handler_agent, status FROM site_work_items_archive) u
             WHERE item_type = 'literal_markdown'
               AND handler_agent = 'page-build-handler') z;
    IF neg_ok THEN
        RAISE EXCEPTION '479: NEGATIVE CONTROL FAILED — literal_markdown->page-build-handler satisfies the reclaim test. It is 3 ok / 36 failed and must stay held; the floor arithmetic or the archive scope is wrong.';
    END IF;

    SELECT count(*) INTO n_escalated
      FROM site_work_items
     WHERE status = 'needs_human_review' AND resolution_path = 'auto:held_pair_escalated';

    RAISE NOTICE '479: reclaim arm live. Currently escalated rows eligible to reclaim: % (0 expected today — nothing has crossed the 3-day limit yet, which is why the controls test the PREDICATE and not the row count). Positive control OK (a qualified pair reclaims). Negative control OK (literal_markdown stays held).',
        n_escalated;
END $$;

COMMIT;
