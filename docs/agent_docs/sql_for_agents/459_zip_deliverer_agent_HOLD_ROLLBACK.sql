-- 459 ROLLBACK: retire the zip-deliverer + zip-deliverable-dispatch
-- definitions seeded by 459_zip_deliverer_agent_HOLD.sql.
--
-- Both were NEW rows (nothing repurposed, no snapshot to restore) — the
-- rollback is a soft delete, matching how the estate retires definitions.
-- Idempotent: a second run updates zero rows.
--
-- Apply:
--   kubectl -n ai-persona-system exec -i postgres-clients-0 -- \
--     psql -U clients_user -d clients_db -v ON_ERROR_STOP=1 \
--     -f - < docs/agent_docs/sql_for_agents/459_zip_deliverer_agent_HOLD_ROLLBACK.sql

BEGIN;

UPDATE agent_definitions
   SET is_active = false, deleted_at = now(), updated_at = now()
 WHERE type IN ('zip-deliverer', 'zip-deliverable-dispatch')
   AND deleted_at IS NULL;

DO $verify$
DECLARE
  n int;
BEGIN
  SELECT count(*) INTO n FROM agent_definitions
   WHERE type IN ('zip-deliverer', 'zip-deliverable-dispatch') AND deleted_at IS NULL;
  IF n <> 0 THEN
    RAISE EXCEPTION '459 rollback verify: % live zip agent definition(s) remain', n;
  END IF;
END $verify$;

COMMIT;
