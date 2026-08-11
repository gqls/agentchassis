-- ROLLBACK for 393: restore the three components' rendered_html from the
-- migration_backups rows 393 wrote. NOTE: restoring re-ghosts the primary
-- buttons (1.05:1 label contrast) — see 393's header for why that is the
-- pre-fix state. After running, redeploy the three pages (RUNBOOK s10b) or
-- the bucket keeps serving the fixed CSS.

BEGIN;

UPDATE page_components c
SET rendered_html = b.old_value->>'rendered_html',
    updated_at = now()
FROM migration_backups b
WHERE b.migration_name = '393_fix_self_referential_css_vars_three_tool_components'
  AND b.target_table = 'page_components'
  AND b.target_id = c.id::text;

DO $$
DECLARE
    restored int;
BEGIN
    SELECT count(*) INTO restored
    FROM page_components c JOIN migration_backups b
      ON b.target_id = c.id::text
     AND b.migration_name = '393_fix_self_referential_css_vars_three_tool_components'
    WHERE c.rendered_html = b.old_value->>'rendered_html';
    IF restored <> 3 THEN
        RAISE EXCEPTION '393 ROLLBACK: expected 3 restored components, found %', restored;
    END IF;
END $$;

COMMIT;
