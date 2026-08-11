-- 391: billing — vouchers, one-off orders, webhook event dedup, settings
--
-- Owner rulings 2026-08-11 (ai_site_selling_automation PLAN §1b/§1c):
-- £149 all-in per site; voucher codes single-use, nameable to a recipient,
-- with an expiry, dropping the price to £10 or £55; payment collected after
-- approval while the system is tested, moving up-front later (a SWITCH, not a
-- constant); NO refund code anywhere (refunds are manual, behind the scenes).
-- The serving surface is auth-service (ruling §1b.1) but billing truth lives
-- HERE in clients_db, beside the customer identity it prices (PLAN §2.7) —
-- the auth service's own MySQL is external shared hosting (rs17.uk-noc.com)
-- and is not where payment state belongs.
--
-- Design notes:
-- * vouchers.redeemed_at IS NULL is the single-use invariant; redemption is
--   one atomic UPDATE ... WHERE redeemed_at IS NULL AND expires_at > now()
--   RETURNING — zero rows = invalid/used/expired, race-safe by construction.
-- * billing_orders.amount_pence is computed server-side (voucher price or
--   list price), never client-supplied. status vocabulary is deliberately
--   tiny: created | paid. There is NO refunded status — per ruling.
-- * billing_events PK (provider, provider_event_id) is webhook idempotency:
--   INSERT ... ON CONFLICT DO NOTHING, and a conflict means already handled.
-- * billing_settings is a one-row table (id=1 CHECK) holding the payment
--   timing switch; the CHECK constraint is the vocabulary.
-- * clients.external_id (unique) remains the Stripe customer linkage
--   (migration 375's stated intent); no column added for it here.
--
-- Rollback recipe (do not run as part of this file):
--   DROP TABLE IF EXISTS billing_events;
--   DROP TABLE IF EXISTS billing_orders;
--   DROP TABLE IF EXISTS vouchers;
--   DROP TABLE IF EXISTS billing_settings;

BEGIN;

CREATE TABLE IF NOT EXISTS vouchers (
    id                    uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    code                  varchar(64) NOT NULL UNIQUE,
    drops_price_to_pence  integer NOT NULL CHECK (drops_price_to_pence >= 0),
    recipient_name        varchar(255),
    expires_at            timestamptz NOT NULL,
    redeemed_at           timestamptz,
    redeemed_by_client_id uuid REFERENCES clients(id),
    created_at            timestamptz NOT NULL DEFAULT now()
);

COMMENT ON TABLE vouchers IS
  'Single-use discount codes the owner hands out (ruling 2026-08-11: £10 and £55 variants against the £149 list price). redeemed_at IS NULL = still valid; redemption is one atomic UPDATE, see billing service.';

CREATE TABLE IF NOT EXISTS billing_orders (
    id                      uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    client_id               uuid NOT NULL REFERENCES clients(id),
    kind                    text NOT NULL DEFAULT 'site_build',
    amount_pence            integer NOT NULL CHECK (amount_pence >= 0),
    voucher_id              uuid REFERENCES vouchers(id),
    status                  text NOT NULL DEFAULT 'created'
                            CHECK (status IN ('created','paid')),
    provider                text NOT NULL DEFAULT 'stripe',
    provider_session_id     text,
    provider_customer_id    text,
    paid_at                 timestamptz,
    created_at              timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_billing_orders_client
    ON billing_orders (client_id, created_at DESC);

COMMENT ON TABLE billing_orders IS
  'One row per checkout the platform issues (£149 site build, or voucher price). Marked paid ONLY by a signature-verified webhook — never by a browser redirect. No refunded status by design (owner ruling 2026-08-11: refunds are manual and unadvertised).';

CREATE TABLE IF NOT EXISTS billing_events (
    provider          text NOT NULL,
    provider_event_id text NOT NULL,
    type              text NOT NULL,
    order_id          uuid,
    payload           jsonb,
    processed_at      timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (provider, provider_event_id)
);

COMMENT ON TABLE billing_events IS
  'Webhook idempotency + audit. The PK is the dedup: INSERT ON CONFLICT DO NOTHING, and a conflict means the event was already processed.';

CREATE TABLE IF NOT EXISTS billing_settings (
    id              smallint PRIMARY KEY CHECK (id = 1),
    payment_timing  text NOT NULL DEFAULT 'after_approval'
                    CHECK (payment_timing IN ('after_approval','upfront')),
    updated_at      timestamptz NOT NULL DEFAULT now()
);

INSERT INTO billing_settings (id) VALUES (1)
ON CONFLICT (id) DO NOTHING;

COMMENT ON TABLE billing_settings IS
  'One-row config. payment_timing is the ruled switch (2026-08-11): after_approval while the system is tested, upfront later. Flip via the admin API, no build needed.';

-- Verify (RFC_006 lesson: SELECTs cannot stop a COMMIT — use DO/RAISE)
DO $$
BEGIN
  IF (SELECT count(*) FROM billing_settings) <> 1 THEN
    RAISE EXCEPTION 'billing_settings must hold exactly one row';
  END IF;
  IF NOT EXISTS (
    SELECT 1 FROM information_schema.table_constraints
    WHERE table_name = 'billing_events' AND constraint_type = 'PRIMARY KEY'
  ) THEN
    RAISE EXCEPTION 'billing_events PK (the webhook dedup) is missing';
  END IF;
END $$;

COMMIT;
