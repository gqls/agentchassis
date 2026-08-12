-- ROLLBACK for 396_thunder_provision_claims.sql
--
-- Run BY HAND, deliberately. The migration runner never executes an
-- UPPERCASE-suffixed sidecar (SIDECAR_RE in scripts/migration/run-migrations.sh).
--
-- ⚠ Dropping this table removes the ONLY thing standing between one provision
-- request and four billable GPUs (bugs_open/259 — the chassis retry driver
-- re-dispatches an expired await with a fresh request_id). The adapter binary
-- treats an absent table as a hard error and refuses to provision rather than
-- provisioning unguarded, so after this drop provisioning FAILS CLOSED until
-- the table is restored. That is deliberate: the alternative is spending money.
--
-- It also discards the audit trail of every provision attempt, which is the
-- only durable record a FAILED provision leaves (bugs_open/258 defect 3).

BEGIN;

DROP TRIGGER IF EXISTS trg_thunder_provision_claims_updated_at ON thunder_provision_claims;
DROP TABLE IF EXISTS thunder_provision_claims;

DO $$
BEGIN
  IF EXISTS (
    SELECT 1 FROM information_schema.tables
    WHERE table_name = 'thunder_provision_claims'
  ) THEN
    RAISE EXCEPTION '396 ROLLBACK: table still present';
  END IF;
END $$;

COMMIT;
