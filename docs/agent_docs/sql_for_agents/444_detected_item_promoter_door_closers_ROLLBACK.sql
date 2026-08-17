-- 444 ROLLBACK — restore 430's detected-item-promoter pre_query verbatim.
--
-- Use when: a door-closer is holding rows it should not (symptom: the
-- promotable pile stops draining and `SELECT count(*) FROM site_work_items
-- WHERE status='detected' AND COALESCE(handler_agent,'')<>''` climbs while the
-- task keeps ticking). DB config — live at the scheduler's next tick.
--
-- The body below is 430's pre_query byte-for-byte. It is NOT an edit of the
-- ledger-recorded 430 file; it is a copy, so applying this leaves the live row
-- exactly as 430 left it.

BEGIN;

UPDATE scheduled_tasks
SET pre_query = $Q$
    WITH candidates AS (
        SELECT wi.id
        FROM site_work_items wi
        WHERE wi.status = 'detected'
          AND COALESCE(wi.handler_agent, '') <> ''
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

DO $$
DECLARE q text;
BEGIN
    SELECT pre_query INTO q FROM scheduled_tasks WHERE name = 'detected-item-promoter';
    IF q LIKE '%0.25 * (c + f)%' OR q LIKE '%wi.pipeline IN (''build'', ''content'', ''design'')%' THEN
        RAISE EXCEPTION '444 ROLLBACK: a door-closer is still present — the restore did not take';
    END IF;
    IF q NOT LIKE '%LIMIT 20%' OR q NOT LIKE '%original_pipeline%' OR q NOT LIKE '%done.status = ''complete''%' THEN
        RAISE EXCEPTION '444 ROLLBACK: 430''s own predicates are missing — do NOT commit, the row is now neither 430 nor 444';
    END IF;
    RAISE NOTICE '444 ROLLBACK: detected-item-promoter restored to 430''s pre_query.';
END $$;

COMMIT;
