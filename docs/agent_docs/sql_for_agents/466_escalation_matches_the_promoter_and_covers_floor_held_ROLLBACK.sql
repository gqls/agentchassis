-- 466 ROLLBACK — restore 453's pre_query (canary-held only, live-table history,
-- non-suppressing WHERE) and 453's description.
--
-- ⚠ Reverting reinstates three MEASURED defects: the task claims a shared
-- max_concurrent=1 slot on every idle tick; floor-held pairs (10 rows on
-- 2026-08-18) are escalated by nothing; and its history test disagrees with the
-- promoter's, so it can ask a human to canary a pair the promoter is already
-- promoting. Prefer adjusting the 3-day limit or the owner map over reverting.
--
-- Does NOT un-escalate rows 466 already moved. To return those:
--   UPDATE site_work_items SET status='detected', resolution_path=NULL,
--          result = result - 'held_pair_escalation', updated_at=now()
--    WHERE resolution_path='auto:held_pair_escalated'
--      AND result->'held_pair_escalation'->>'escalated_by' LIKE '%466%';

BEGIN;

UPDATE scheduled_tasks
SET pre_query = $Q$
    WITH held AS (
        SELECT wi.id, wi.item_type, wi.handler_agent, wi.created_at
        FROM site_work_items wi
        WHERE wi.status = 'detected'
          AND COALESCE(wi.handler_agent, '') <> ''
          AND NOT EXISTS (
            SELECT 1 FROM site_work_items done
            WHERE done.item_type = wi.item_type
              AND done.handler_agent = wi.handler_agent
              AND done.status = 'complete'
          )
          AND EXISTS (
            SELECT 1 FROM agent_definitions ad
            WHERE ad.type = wi.handler_agent
              AND ad.is_active
              AND COALESCE(ad.is_snapshot, false) = false
              AND ad.deleted_at IS NULL
          )
    ),
    overdue AS (
        SELECT h.*, p.oldest, (now()::date - p.oldest::date) AS days_waiting
        FROM held h
        JOIN (
            SELECT item_type, handler_agent, min(created_at) AS oldest
            FROM held GROUP BY 1, 2
            HAVING min(created_at) < now() - interval '3 days'
        ) p ON p.item_type = h.item_type AND p.handler_agent = h.handler_agent
    ),
    owners (item_type, owner) AS (
        VALUES
          ('placeholder_contact',
           'bugs_open/201 lane — docs024_key_docs_latest/bugfix_201_page_content_writer_dispatch'),
          ('page_component_status_drift',
           '(UNASSIGNED - claim this) check_page_component_status_drift.go added 2026-07-10, never touched since, no lane doc claims it')
    ),
    escalated AS (
        UPDATE site_work_items wi
        SET status = 'needs_human_review',
            resolution_path = 'auto:held_pair_escalated',
            result = COALESCE(wi.result, '{}'::jsonb) || jsonb_build_object(
                'held_pair_escalation', jsonb_build_object(
                    'at',            to_char(now() AT TIME ZONE 'UTC', 'YYYY-MM-DD"T"HH24:MI:SS"Z"'),
                    'reason',        'held by the detected-item-promoter known-good rule (SCH-026): this (item_type, handler_agent) pair has never completed one, so the promoter will not dispatch it until a human runs one by hand',
                    'days_waiting',  o.days_waiting,
                    'limit_days',    3,
                    'pair',          o.item_type || ' -> ' || o.handler_agent,
                    'owner',         COALESCE(ow.owner, '(UNASSIGNED - claim this) no owner named for this item_type in migration 453'),
                    'what_to_do',    'Promote ONE row of this pair by hand and watch it: UPDATE site_work_items SET status=''triaged'', pipeline=''build'', triaged_at=now(), spec=jsonb_set(COALESCE(spec,''{}''::jsonb),''{original_pipeline}'',to_jsonb(pipeline)) WHERE id=''<one id>''. If it completes, the pair becomes known-good and the promoter takes the rest automatically. If it fails, that is the finding — file it.',
                    'escalated_by',  'held-pair-canary-escalation (migration 453, owner decision 2026-08-17)'
                )
            ),
            updated_at = now()
        FROM overdue o
        LEFT JOIN owners ow ON ow.item_type = o.item_type
        WHERE wi.id = o.id
          AND wi.status = 'detected'
        RETURNING wi.id, wi.item_type, wi.handler_agent
    )
    SELECT COUNT(*)::text AS escalated,
           string_agg(DISTINCT item_type || '->' || handler_agent, ', ') AS pairs
    FROM escalated
    WHERE (SELECT COUNT(*) FROM escalated) > 0
$Q$,
    description = 'Owner decision 2026-08-17 (bugs_open/083). A (item_type, handler_agent) pair held by SCH-026''s known-good rule for more than 3 days is escalated from detected to needs_human_review, carrying its named owner and what the human must do. The owner map enriches; it never gates.',
    updated_at = now()
WHERE name = 'held-pair-canary-escalation';

DO $$
DECLARE q text;
BEGIN
    SELECT pre_query INTO q FROM scheduled_tasks WHERE name = 'held-pair-canary-escalation';
    IF q LIKE '%hold_kind%' OR q LIKE '%site_work_items_archive%' THEN
        RAISE EXCEPTION '466 ROLLBACK: 466''s shape survives — the restore did not take';
    END IF;
    IF q NOT LIKE '%WHERE (SELECT COUNT(*) FROM escalated) > 0%' OR q NOT LIKE '%interval ''3 days''%' THEN
        RAISE EXCEPTION '466 ROLLBACK: 453''s own text is missing — the row is now neither 453 nor 466, do NOT commit';
    END IF;
    RAISE NOTICE '466 ROLLBACK: held-pair-canary-escalation restored to 453''s pre_query (and its three defects).';
END $$;

COMMIT;
