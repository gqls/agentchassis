-- 465 ROLLBACK — restore the live-table-only success tests (458's pre_query).
--
-- Use when: the union's cost is unacceptable, or including archived history
-- promotes something it should not. Reverses exactly 465's two substitutions
-- and nothing else, so any later edit by another lane survives.
--
-- ⚠ Reverting reinstates a KNOWN, MEASURED defect: both success tests go back
-- to reading a ~7-day window and calling it lifetime, which was stranding
-- empty_internal_href -> page-build-handler (9 lifetime successes, read as
-- zero). Prefer narrowing the union to a date range over reverting it.

BEGIN;

UPDATE scheduled_tasks
SET pre_query = replace(
      replace(
        pre_query,
        'SELECT 1 FROM (SELECT item_type, handler_agent, status FROM site_work_items UNION ALL SELECT item_type, handler_agent, status FROM site_work_items_archive) done',
        'SELECT 1 FROM site_work_items done'
      ),
      'FROM (SELECT item_type, handler_agent, status FROM site_work_items UNION ALL SELECT item_type, handler_agent, status FROM site_work_items_archive) h',
      'FROM site_work_items h'
    ),
    updated_at = now()
WHERE name = 'detected-item-promoter';

DO $$
DECLARE q text; n int;
BEGIN
    SELECT pre_query INTO q FROM scheduled_tasks WHERE name = 'detected-item-promoter';
    n := (length(q) - length(replace(q, 'site_work_items_archive', '')));
    IF n <> 0 THEN
        RAISE EXCEPTION '465 ROLLBACK: archive references remain — the restore did not take';
    END IF;
    IF q NOT LIKE '%SELECT 1 FROM site_work_items done%' OR q NOT LIKE '%FROM site_work_items h%' THEN
        RAISE EXCEPTION '465 ROLLBACK: 458''s anchors are missing — the row is now neither 458 nor 465, do NOT commit';
    END IF;
    RAISE NOTICE '465 ROLLBACK: promoter restored to live-table-only history (458''s shape).';
END $$;

COMMIT;
