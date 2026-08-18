-- ROLLBACK for 479 — removes the reclaim arm from held-pair-canary-escalation by
-- the same anchored replacement, restoring the pre-479 text byte-exactly.
--
-- Pre-image: md5 db96f482ae8792dd19f9eb343fd38308, length 7803, captured live
-- 2026-08-18 immediately before 479 was applied.
--
-- WHAT REVERTING GIVES UP, so the choice is deliberate: the escalation door goes
-- back to opening ONE WAY. A finding escalated because its pair was held will stay
-- in the human queue for ever, even after that pair qualifies — which migration
-- 465 showed is a real event (it released a pair that had been held as "never
-- completed" while holding nine lifetime successes).
--
-- Does NOT re-escalate rows the arm already reclaimed. Those sit at `detected`
-- with a `result.held_pair_reclaimed` block, which is the intended end state; the
-- promoter owns them. To put one back by hand:
--   UPDATE site_work_items SET status='needs_human_review',
--          resolution_path='auto:held_pair_escalated', updated_at=now()
--    WHERE id='<id>' AND result ? 'held_pair_reclaimed';

BEGIN;

DO $$
DECLARE
    q text;
BEGIN
    SELECT pre_query INTO q FROM scheduled_tasks WHERE name = 'held-pair-canary-escalation';
    IF q IS NULL THEN
        RAISE EXCEPTION '479 ROLLBACK: no held-pair-canary-escalation row';
    END IF;
    IF md5(q) = 'db96f482ae8792dd19f9eb343fd38308' THEN
        RAISE NOTICE '479 ROLLBACK: already at the pre-479 text — nothing to do.';
    ELSIF q NOT LIKE '%reclaimable AS (%' THEN
        -- neither 479's output nor the pre-image: a third edit landed and blind
        -- restoration would revert it.
        RAISE EXCEPTION '479 ROLLBACK: ABORTING — the live pre_query is neither 479''s output nor the pre-image (md5 %). Another session has edited it since; read the live column before restoring anything.', md5(q);
    END IF;
END $$;

-- Undo replacement 3, then 2, then 1 — reverse order, same anchors.
UPDATE scheduled_tasks
SET pre_query = replace(pre_query,
E'    HAVING COUNT(*) > 0 OR (SELECT COUNT(*) FROM classified) > 0\n'
'        OR (SELECT COUNT(*) FROM reclaimed) > 0',
E'    HAVING COUNT(*) > 0 OR (SELECT COUNT(*) FROM classified) > 0')
WHERE name = 'held-pair-canary-escalation';

UPDATE scheduled_tasks
SET pre_query = replace(pre_query,
E'    SELECT COUNT(*)::text AS escalated,\n'
'           (SELECT COUNT(*)::text FROM reclaimed) AS reclaimed,\n'
'           (SELECT string_agg(DISTINCT item_type || ''->'' || handler_agent, '', '')\n'
'              FROM reclaimed) AS reclaimed_pairs,',
E'    SELECT COUNT(*)::text AS escalated,')
WHERE name = 'held-pair-canary-escalation';

UPDATE scheduled_tasks
SET pre_query = replace(pre_query,
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
'    escalated AS (',
E'    escalated AS (')
WHERE name = 'held-pair-canary-escalation';

DO $$
DECLARE
    live_md5 text;
BEGIN
    SELECT md5(pre_query) INTO live_md5 FROM scheduled_tasks WHERE name = 'held-pair-canary-escalation';
    IF live_md5 <> 'db96f482ae8792dd19f9eb343fd38308' THEN
        RAISE EXCEPTION '479 ROLLBACK: restored text does not match the captured pre-image (got %, want db96f482ae8792dd19f9eb343fd38308). The restoration is NOT byte-exact — do not commit it.', live_md5;
    END IF;
    RAISE NOTICE '479 ROLLBACK: pre_query restored byte-exactly to the pre-479 text.';
END $$;

COMMIT;
