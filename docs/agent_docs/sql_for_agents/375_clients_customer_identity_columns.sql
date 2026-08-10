-- 375: clients gains customer identity columns
--
-- Owner ruling 2026-08-10 (ai_site_selling_automation PLAN §1): customer
-- identity lives on the existing clients -> networks -> sites chain, BIZ-014
-- shape — customer-shaped columns are ADDED to clients; sites gains no
-- ownership columns. Stripe linkage stays on the existing external_id.
--
-- Additive, all nullable. customer_status is deliberately NOT named `status`:
-- sites.status is load-bearing for tools-api CORS (status='deployed') and
-- auth-service subscription.status already means only "a row exists" — a third
-- overloaded `status` invites the same trap.
--
-- Blast radius measured before writing (2026-08-10): zero Go readers of
-- FROM clients across internal/ platform/ cmd/; zero positional
-- INSERT INTO clients in go/sql/sh. The admin API's /clients endpoints read a
-- DIFFERENT store (clients_info, lazily created, absent live) and are
-- untouched by this file.
--
-- Rollback recipe (do not run as part of this file):
--   ALTER TABLE clients
--     DROP COLUMN IF EXISTS email, DROP COLUMN IF EXISTS phone,
--     DROP COLUMN IF EXISTS tier, DROP COLUMN IF EXISTS customer_status,
--     DROP COLUMN IF EXISTS notes;

BEGIN;

ALTER TABLE clients
  ADD COLUMN IF NOT EXISTS email varchar(255),
  ADD COLUMN IF NOT EXISTS phone varchar(50),
  ADD COLUMN IF NOT EXISTS tier varchar(50),
  ADD COLUMN IF NOT EXISTS customer_status varchar(50),
  ADD COLUMN IF NOT EXISTS notes text;

COMMENT ON COLUMN clients.email IS
  'Customer contact email (ai_site_selling lane). sites.email predates this and is per-site delivery contact; this is the account-level one.';
COMMENT ON COLUMN clients.phone IS
  'Customer contact phone (account-level; see email comment).';
COMMENT ON COLUMN clients.tier IS
  'Product tier, e.g. done_for_you (GBP 1200 ruling 2026-08-10). Free-form until the tier vocabulary is ruled.';
COMMENT ON COLUMN clients.customer_status IS
  'Customer lifecycle. Deliberately not named status — see migration 375 header.';
COMMENT ON COLUMN clients.notes IS
  'Free-form admin notes shown in the admin dashboard Customers tab.';

DO $$
DECLARE
  n int;
BEGIN
  SELECT count(*) INTO n FROM information_schema.columns
   WHERE table_schema = 'public' AND table_name = 'clients'
     AND column_name IN ('email', 'phone', 'tier', 'customer_status', 'notes');
  IF n <> 5 THEN
    RAISE EXCEPTION '375: expected 5 customer columns on clients, found %', n;
  END IF;
END $$;

COMMIT;
