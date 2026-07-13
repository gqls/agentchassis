-- Fix database-cleanup to always return a row
UPDATE scheduled_tasks
SET pre_query = REPLACE(
        pre_query,
        'HAVING
            (SELECT COUNT(*) FROM deleted_errors) > 0
            OR (SELECT COUNT(*) FROM deleted_audit) > 0
            OR (SELECT COUNT(*) FROM deleted_orchestrations) > 0
            OR (SELECT COUNT(*) FROM deleted_stale) > 0',
        '-- Always returns a row so scheduler marks task as executed'
                )
WHERE name = 'database-cleanup';