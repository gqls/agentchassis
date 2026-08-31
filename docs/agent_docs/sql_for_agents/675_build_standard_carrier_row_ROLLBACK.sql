-- 675 ROLLBACK — remove the carrier row. Safe while nothing injects it (pre-roll) or after
-- 676_ROLLBACK has stripped the opt-ins; with opt-ins live, removing the carrier degrades to
-- "no block" (voicestyle serves last-known-good until pod restart, then templates render the
-- heading over nothing) — strip the opt-ins first.
BEGIN;
DELETE FROM agent_default_configs WHERE config_name='build_standard_block';
DELETE FROM schema_migrations WHERE filename='675_build_standard_carrier_row.sql';
COMMIT;
