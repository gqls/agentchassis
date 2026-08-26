-- 659 ROLLBACK — drop the intake join column and its index.
--
-- ⚠ Only safe while NO auth-service binary that names the column is running:
-- the post-659 binary's INSERT ... RETURNING lists external_reference, and
-- dropping the column under it breaks every order creation. Roll the binary
-- back first, column second — the reverse of the forward order.

BEGIN;
DROP INDEX IF EXISTS idx_billing_orders_external_ref;
ALTER TABLE billing_orders DROP COLUMN IF EXISTS external_reference;
COMMIT;
