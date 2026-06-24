-- NNN_create_diagnose_ro_role.sql
--
-- Renumber NNN to the next number in your migration sequence before applying
-- (the repo uses 3-digit prefixes, e.g. 002_, 003_, 018_).
--
-- A SELECT-only login role for the diagnosis HARNESS connection (cmd/diagnose ->
-- dbcontext -> psql -c). It is the harness analogue of the chassis guarantee: in
-- the chassis the loop runs model-written queries inside db.BeginTx(ReadOnly)
-- (Guard 3); the harness has no Go tx around `psql -c`, so the connection itself
-- must be read-only. This is GRANT-based (SELECT-only) — NOT
--   ALTER ROLE diagnose_ro SET default_transaction_read_only = on
-- which is unreliable under pgbouncer transaction pooling (the session GUC does
-- not travel with a pooled server connection).
--
-- PRIVILEGES: this migration needs a role with CREATEROLE (and ownership/superuser
-- for GRANT). The app role clients_user may NOT have CREATEROLE — run it as the
-- DB owner/admin. The password is set OUT OF BAND from a secret (see below), so it
-- is never committed here.
--
-- Idempotent: re-running is safe (role created only if absent; GRANTs are
-- additive).

DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'diagnose_ro') THEN
        CREATE ROLE diagnose_ro LOGIN;
    END IF;
END
$$;

-- Set the password from a secret, NOT in this file. Either run separately:
--   ALTER ROLE diagnose_ro PASSWORD 'set-from-secret';
-- or apply with a psql variable:
--   psql -v ro_pw="$DIAGNOSE_RO_PASSWORD" -f NNN_create_diagnose_ro_role.sql
-- and uncomment the next line:
-- ALTER ROLE diagnose_ro PASSWORD :'ro_pw';

GRANT CONNECT ON DATABASE clients_db TO diagnose_ro;
GRANT USAGE ON SCHEMA public TO diagnose_ro;
GRANT SELECT ON ALL TABLES IN SCHEMA public TO diagnose_ro;

-- Future tables also become readable by the role automatically. Run this as the
-- role that OWNS/creates the tables (otherwise default privileges apply only to
-- objects that role creates).
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT SELECT ON TABLES TO diagnose_ro;

-- Verify after applying (should list SELECT, and nothing else):
--   \dp site_flows
--   SELECT has_table_privilege('diagnose_ro','site_flows','SELECT') AS can_read,
--          has_table_privilege('diagnose_ro','site_flows','DELETE') AS can_delete;
-- can_read = t, can_delete = f.

-- Down (manual, if ever needed):
--   REVOKE ALL ON ALL TABLES IN SCHEMA public FROM diagnose_ro;
--   REVOKE ALL ON SCHEMA public FROM diagnose_ro;
--   REVOKE ALL ON DATABASE clients_db FROM diagnose_ro;
--   DROP ROLE diagnose_ro;
