-- ============================================================================
-- 469 ROLLBACK — restore the render-audit rotation's 7-day re-visit window
-- ============================================================================
-- Hand-run only; run-migrations.sh never auto-applies an UPPERCASE sidecar.
-- Use if the ~2.3x rise in audit volume is unwanted, or if sites are observed
-- slipping their turn (symptom: last_selected_at gaps materially wider than
-- the window, i.e. the rotation cannot keep up with what it makes due).
-- ============================================================================

BEGIN;

DO $$
DECLARE n int;
BEGIN
    SELECT count(*) INTO n FROM scheduled_tasks
     WHERE name = 'site-render-audit-rotation' AND pre_query LIKE '%interval ''3 days''%';
    IF n <> 1 THEN
        RAISE EXCEPTION '469 ROLLBACK: no 3-day literal on site-render-audit-rotation '
                        '(matched % rows) — already rolled back, or rewritten since', n;
    END IF;
END $$;

UPDATE scheduled_tasks
   SET pre_query  = replace(pre_query, 'interval ''3 days''', 'interval ''7 days'''),
       updated_at = now()
 WHERE name = 'site-render-audit-rotation';

DO $$
DECLARE bad int;
BEGIN
    SELECT count(*) INTO bad FROM scheduled_tasks
     WHERE name = 'site-render-audit-rotation'
       AND (pre_query LIKE '%interval ''3 days''%' OR pre_query NOT LIKE '%interval ''7 days''%');
    IF bad <> 0 THEN
        RAISE EXCEPTION '469 ROLLBACK VERIFY FAILED (% row(s)) — refusing to commit', bad;
    END IF;
END $$;

COMMIT;
