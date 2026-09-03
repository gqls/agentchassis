-- Rollback for 733: remove the network-level default. Existing per-site rows are untouched;
-- with the value absent, seedAnalyticsDefault seeds nothing (its first guard).
UPDATE networks
   SET settings = jsonb_set(settings, '{analytics}', (settings->'analytics') - 'gtm_container_id'),
       updated_at = now()
 WHERE id = '00000000-0000-0000-0000-000000000002'
   AND settings->'analytics' ? 'gtm_container_id';
