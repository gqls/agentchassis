-- ROLLBACK for 667: restores every touched site's offer_ordering data from migration_backups.
BEGIN;
UPDATE site_specs sp
   SET data = mb.old_value->'data'
  FROM migration_backups mb
 WHERE mb.migration_name='667_offer_ordering_register_wash_41_points'
   AND mb.target_id = sp.id::text
   AND sp.aspect='offer_ordering' AND sp.is_current;
COMMIT;
