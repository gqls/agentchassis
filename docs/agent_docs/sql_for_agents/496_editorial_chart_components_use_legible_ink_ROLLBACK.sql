-- ROLLBACK for 496. Restores both component templates from the backup table
-- 496 created. Hand-run only; not picked up by the migration runner.
\set ON_ERROR_STOP on
BEGIN;
UPDATE content_components cc
   SET html_template = b.html_template, updated_at = now()
  FROM content_components_backup_20260820_legible_ink b
 WHERE b.id = cc.id;
DO $$
BEGIN
  IF EXISTS (SELECT 1 FROM content_components
              WHERE function IN ('evidence-chart','evidence-timeseries') AND is_active
                AND html_template LIKE '%-ink, var(--color-%') THEN
    RAISE EXCEPTION 'rollback did not restore both templates';
  END IF;
END $$;
COMMIT;
