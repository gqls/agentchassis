-- ROLLBACK for 622 — restore the un-guarded sitemap rotation pre_query.
--
-- What this restores is the DEFECT: a site selected before its pages deploy burns
-- its 3-day slot and serves no sitemap until the next rotation. Measured live on
-- homegarden.uk and cv1.co.uk, 2026-08-25. Only roll this back if the guard is
-- itself causing harm, and say what.

BEGIN;

UPDATE scheduled_tasks
   SET pre_query = regexp_replace(
         pre_query,
         E'\n    -- 622: do not spend.*?AND pg\\.deployed_at IS NOT NULL\\)',
         '', 'sn'),
       updated_at = now()
 WHERE name = 'sitemap-refresh-rotation';

DELETE FROM schema_migrations WHERE filename = '622_sitemap_rotation_skips_a_site_with_nothing_to_list.sql';

DO $$
DECLARE v_pq text;
BEGIN
    SELECT pre_query INTO v_pq FROM scheduled_tasks WHERE name='sitemap-refresh-rotation';
    IF v_pq LIKE '%pg.deployed_at IS NOT NULL%' THEN
        RAISE EXCEPTION 'rollback: the 622 guard is still present';
    END IF;
    IF v_pq NOT LIKE '%locked_at IS NULL%' THEN
        RAISE EXCEPTION 'rollback: removed too much — the locked_at guard is gone';
    END IF;
    RAISE NOTICE '622 rollback OK — guard removed, locked_at intact.';
END $$;

COMMIT;
