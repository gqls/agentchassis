-- ROLLBACK for 627: restores the pre-627 default_config from migration_backups.
BEGIN;
UPDATE agent_definitions ad
   SET default_config = (SELECT mb.old_value->'default_config' FROM migration_backups mb
                          WHERE mb.migration_name='627_page_content_writer_delete_july_substitutes'
                          ORDER BY mb.applied_at DESC LIMIT 1)
 WHERE ad.type='page-content-writer' AND ad.is_active AND COALESCE(ad.is_snapshot,false)=false AND ad.deleted_at IS NULL;
COMMIT;
