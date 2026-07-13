-- ============================================================================
-- 006_unify_prices_schema.sql
-- ============================================================================
--
-- Unifies the two price tables (business_prices for services and
-- product_prices for products) under a single polymorphic schema rooted on
-- the existing business_intel.products catalog.
--
-- SUPERSEDES the previously planned 001_business_med_prices_migration.sql,
-- which would have created a third sibling table. That file should not be
-- applied — the existing products + product_prices schema is the right
-- shape, it just needed extending to cover services too.
--
-- WHAT THIS DOES
--   (A) Adds a `kind` discriminator column to business_intel.products.
--       Defaults to 'medicine' for backwards compatibility with all existing
--       rows. Accepts: medicine, service, supplement, other.
--   (B) Migrates distinct (service_category, service_name) tuples from
--       business_prices into products as rows with kind='service'.
--   (C) Migrates business_prices price rows into product_prices, joined to
--       the new service products via category+name.
--   (D) Marks business_prices as deprecated via table comment (does NOT
--       drop it — that happens in a later migration after Go cutover).
--
-- WHAT THIS DOES NOT DO
--   - Drop business_prices. Go's insertPrice still writes there until the
--     paired Go change ships. Dropping prematurely breaks the verifier.
--   - Touch the med_products table or the online retailer price pipeline.
--     Those continue to use their own catalog (med_products) for
--     retailer-side pricing. A future pass could merge med_products into
--     products as kind='medicine' rows with a retailer_id column on
--     product_prices, but that's a separate piece of work.
--   - Reconcile naming. product_prices is technically misnamed now that
--     it covers services too — "offering_prices" would be more honest. But
--     renaming touches every reader in Go, so we accept the slight
--     misnomer and document it via the kind column.
--
-- ORDER OF OPERATIONS (operator runbook)
--   1. Ship the paired Go change (formerly Go-B): update insertPrice and
--      add a medicine-write helper, both targeting products + product_prices
--      with kind='service' or kind='medicine'. Deploy chassis.
--   2. Apply this migration. Backfills historical business_prices.
--   3. Verify row counts via the SELECT at the end.
--   4. (Future migration) DROP business_prices once no reader references it.
--
-- This migration is IDEMPOTENT. Safe to re-run if Go-B writes new rows to
-- business_prices in the interim — the second run will pick them up.
-- ============================================================================

BEGIN;

-- ----------------------------------------------------------------------------
-- (A) Add kind discriminator to products
-- ----------------------------------------------------------------------------
ALTER TABLE business_intel.products
    ADD COLUMN IF NOT EXISTS kind text NOT NULL DEFAULT 'medicine';

-- Add the CHECK constraint idempotently (ADD CONSTRAINT IF NOT EXISTS is
-- PG14+; use a DO block for broader compatibility).
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint
        WHERE conname = 'products_kind_check'
          AND conrelid = 'business_intel.products'::regclass
    ) THEN
        ALTER TABLE business_intel.products
            ADD CONSTRAINT products_kind_check
            CHECK (kind IN ('medicine', 'service', 'supplement', 'other'));
    END IF;
END $$;

CREATE INDEX IF NOT EXISTS idx_bi_products_kind
    ON business_intel.products (kind);

COMMENT ON COLUMN business_intel.products.kind IS
    'Discriminator for the type of offering. Existing rows default to '
    '"medicine" for backwards compat. Services migrated from business_prices '
    'use "service". Allowed values enforced by products_kind_check.';

-- ----------------------------------------------------------------------------
-- (B) Insert distinct service offerings as products of kind=service
-- ----------------------------------------------------------------------------
-- Slug pattern: `service-{lowercased,hyphenated category-name}`
-- Same (category, name) tuple from any business produces the same slug,
-- so all observations of the same service share one product row.
INSERT INTO business_intel.products (
    slug, name, category, kind, requires_prescription, is_active
)
SELECT DISTINCT
    'service-' || regexp_replace(
        lower(service_category || '-' || service_name),
        '[^a-z0-9]+', '-', 'g'
    ) AS slug,
    service_name AS name,
    service_category AS category,
    'service' AS kind,
    -- prescription category implies prescription requirement; others don't
    (service_category = 'prescription') AS requires_prescription,
    true AS is_active
FROM business_intel.business_prices
WHERE service_category IS NOT NULL
  AND service_name IS NOT NULL
  AND length(trim(service_category)) > 0
  AND length(trim(service_name)) > 0
ON CONFLICT (slug) DO NOTHING;

-- ----------------------------------------------------------------------------
-- (C) Migrate business_prices rows into product_prices
-- ----------------------------------------------------------------------------
-- Join business_prices to the newly-inserted service products via
-- (kind, category, name). NOT EXISTS check keeps the migration idempotent
-- across re-runs.
--
-- Field mapping:
--   bp.business_id      -> pp.business_id
--   p.id                -> pp.product_id (resolved via category+name join)
--   bp.price_gbp        -> pp.price_gbp
--   bp.price_qualifier  -> pp.price_qualifier
--   bp.includes_vat     -> pp.includes_vat (default true if null)
--   bp.source_url       -> pp.product_url (closest semantic match — where
--                         the price was observed)
--   bp.source           -> pp.source (or "migrated_from_business_prices"
--                         if null on the source row)
--   bp.observed_at      -> pp.observed_at
--   bp.is_current       -> pp.is_current
INSERT INTO business_intel.product_prices (
    product_id, business_id, price_gbp, price_qualifier, includes_vat,
    product_url, source, observed_at, is_current
)
SELECT
    p.id                                              AS product_id,
    bp.business_id                                    AS business_id,
    bp.price_gbp                                      AS price_gbp,
    bp.price_qualifier                                AS price_qualifier,
    COALESCE(bp.includes_vat, true)                   AS includes_vat,
    bp.source_url                                     AS product_url,
    COALESCE(NULLIF(bp.source, ''),
             'migrated_from_business_prices')         AS source,
    COALESCE(bp.observed_at, now())                   AS observed_at,
    COALESCE(bp.is_current, true)                     AS is_current
FROM business_intel.business_prices bp
JOIN business_intel.products p
    ON p.kind     = 'service'
   AND p.category = bp.service_category
   AND p.name     = bp.service_name
WHERE NOT EXISTS (
    -- Already-migrated rows skip. Match on (business_id, product_id,
    -- observed_at) — same business observing the same product at the same
    -- moment is a duplicate.
    SELECT 1 FROM business_intel.product_prices pp
    WHERE pp.business_id = bp.business_id
      AND pp.product_id  = p.id
      AND pp.observed_at = COALESCE(bp.observed_at, now())
);

-- ----------------------------------------------------------------------------
-- (D) Mark business_prices as deprecated
-- ----------------------------------------------------------------------------
COMMENT ON TABLE business_intel.business_prices IS
    'DEPRECATED — service prices migrated to product_prices via products '
    'with kind=service. New writes should go to products + product_prices. '
    'This table is retained for backwards compatibility during the Go '
    'cutover and will be dropped in a follow-up migration once '
    'store_business_verification has been updated to write directly to '
    'product_prices. See 006_unify_prices_schema.sql.';

-- ----------------------------------------------------------------------------
-- Verification — eyeball these before COMMIT
-- ----------------------------------------------------------------------------
-- Expected results after migration:
--   business_prices_all                = N (unchanged, not dropped)
--   business_prices_current_true       = M (subset of N)
--   products_kind_service              = D (distinct service rows seen)
--   product_prices_for_services        ≈ N (one product_prices row per
--                                         business_prices row, allowing for
--                                         deduplication if any business had
--                                         dupes on observed_at)
--   products_kind_medicine             = existing medicine catalog
SELECT 'business_prices_all'           AS metric, COUNT(*) AS rows
FROM business_intel.business_prices
UNION ALL
SELECT 'business_prices_current_true', COUNT(*)
FROM business_intel.business_prices WHERE is_current = true
UNION ALL
SELECT 'products_kind_service', COUNT(*)
FROM business_intel.products WHERE kind = 'service'
UNION ALL
SELECT 'products_kind_medicine', COUNT(*)
FROM business_intel.products WHERE kind = 'medicine'
UNION ALL
SELECT 'product_prices_for_services', COUNT(*)
FROM business_intel.product_prices pp
JOIN business_intel.products p ON p.id = pp.product_id
WHERE p.kind = 'service'
UNION ALL
SELECT 'product_prices_for_medicines', COUNT(*)
FROM business_intel.product_prices pp
JOIN business_intel.products p ON p.id = pp.product_id
WHERE p.kind = 'medicine';

COMMIT;

-- ============================================================================
-- ROLLBACK NOTES
-- ============================================================================
-- To revert this migration:
--
--   BEGIN;
--   -- Remove migrated price rows (they're identifiable by the source tag)
--   DELETE FROM business_intel.product_prices
--   WHERE source = 'migrated_from_business_prices'
--      OR product_id IN (
--          SELECT id FROM business_intel.products WHERE kind = 'service'
--      );
--
--   -- Remove service products
--   DELETE FROM business_intel.products WHERE kind = 'service';
--
--   -- Drop the kind column and constraint
--   ALTER TABLE business_intel.products DROP CONSTRAINT IF EXISTS products_kind_check;
--   DROP INDEX IF EXISTS business_intel.idx_bi_products_kind;
--   ALTER TABLE business_intel.products DROP COLUMN IF EXISTS kind;
--
--   -- Restore the comment
--   COMMENT ON TABLE business_intel.business_prices IS NULL;
--   COMMIT;
--
-- Note: rollback DELETE on product_prices uses the source tag for safety —
-- if Go was updated to write directly to product_prices in between, those
-- rows have source values like 'scrape' or 'verifier' and will be preserved.
-- Only rows tagged 'migrated_from_business_prices' are removed.
