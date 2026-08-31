-- ROLLBACK for 668: restores every touched site's offer_ordering data from migration_backups.
-- Note: if the producer regenerated a row between apply and rollback, this overwrites the
-- regeneration - acceptable, regenerations are re-derivable by the analyser.
BEGIN;
UPDATE site_specs sp
   SET data = mb.old_value->'data'
  FROM migration_backups mb
 WHERE mb.migration_name='668_offer_ordering_expansion_10_points'
   AND mb.target_id = sp.id::text
   AND sp.aspect='offer_ordering' AND sp.is_current;
COMMIT;
