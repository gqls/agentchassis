-- ROLLBACK for 628: restores the owner-approved v2 voice text from migration_backups.
BEGIN;
UPDATE agent_default_configs c
   SET config = (SELECT mb.old_value->'config' FROM migration_backups mb
                  WHERE mb.migration_name='628_house_voice_v3_form_rewrite'
                  ORDER BY mb.applied_at DESC LIMIT 1)
 WHERE c.config_name='voice_style_block';
COMMIT;
