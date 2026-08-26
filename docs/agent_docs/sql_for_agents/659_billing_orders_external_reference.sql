-- 659: billing_orders.external_reference — the order-intake join key.
--
-- OWNER RULING 2026-08-26: payment joins a committed brief by an ORDER
-- REFERENCE, never by carrying the brief ("the brief will change"). The chat
-- bot mints BR-XXXXXX when it stores a brief on the site box; the customer
-- quotes it at payment; the dashboard's create-order call records it here;
-- and collect_external_orders releases the brief into build_queue only when a
-- PAID row carries its reference. NULL is the normal state for every order
-- with no brief behind it (domain rental, buy-out, historic rows).
--
-- ORDERING: safe to apply ahead of the auth-service roll that writes it — the
-- running binary neither selects nor inserts the column until the roll, and a
-- column addition breaks nothing that ignores it. The REVERSE order breaks:
-- the new binary's INSERT ... RETURNING names the column. So: migration
-- first, roll second — ordinary config-before-image.
--
-- Not UNIQUE, deliberately: a checkout that fails after order creation is
-- retried as a NEW order quoting the SAME reference (billing's own recovery
-- bias, service.go), so two rows sharing a reference is a legitimate history.
-- The collector takes the newest PAID row.

BEGIN;

ALTER TABLE billing_orders ADD COLUMN IF NOT EXISTS external_reference TEXT;

CREATE INDEX IF NOT EXISTS idx_billing_orders_external_ref
    ON billing_orders (external_reference)
    WHERE external_reference IS NOT NULL;

COMMENT ON COLUMN billing_orders.external_reference IS
    'Chat-minted brief reference (BR-XXXXXX) joining this payment to a committed brief on the site box. NULL for orders with no brief behind them. Collector: collect_external_orders (releases briefs on the newest PAID row per reference). Owner ruling 2026-08-26: reference, never the brief.';

DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM information_schema.columns
     WHERE table_name = 'billing_orders' AND column_name = 'external_reference'
  ) THEN
    RAISE EXCEPTION '659 verify failed: external_reference column missing';
  END IF;
  IF NOT EXISTS (
    SELECT 1 FROM pg_indexes
     WHERE tablename = 'billing_orders' AND indexname = 'idx_billing_orders_external_ref'
  ) THEN
    RAISE EXCEPTION '659 verify failed: idx_billing_orders_external_ref missing';
  END IF;
END $$;

COMMIT;
