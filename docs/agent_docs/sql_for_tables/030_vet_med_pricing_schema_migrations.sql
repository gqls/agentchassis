-- med_pricing_schema_migration.sql
-- Creates the 4 med pricing tables + materialized view in business_intel schema.
-- Run against clients_db.

BEGIN;

-- ============================================================================
-- 1. med_products — Canonical medicine catalog
-- ============================================================================
CREATE TABLE IF NOT EXISTS business_intel.med_products (
                                                           id              TEXT PRIMARY KEY,          -- e.g. 'm_metacam_dog_100ml'
                                                           name            TEXT NOT NULL,             -- 'Metacam Oral Suspension (Dog)'
                                                           generic_name    TEXT,                      -- 'Meloxicam'
                                                           brand           TEXT,                      -- 'Metacam'
                                                           manufacturer    TEXT,                      -- 'Boehringer Ingelheim'
                                                           species         TEXT[],                    -- {'dog','cat'}
                                                           category        TEXT,                      -- 'nsaid', 'antiparasitic', 'cardiac', etc.
                                                           form            TEXT,                      -- 'oral_suspension', 'tablet', 'spot_on', 'injection'
                                                           strength        TEXT,                      -- '1.5mg/ml', '5.4mg'
                                                           pack_size       TEXT,                      -- '100ml', '100 tabs', '3 pipettes'
                                                           prescription_required BOOLEAN DEFAULT true,
                                                           is_active       BOOLEAN DEFAULT true,
                                                           metadata        JSONB DEFAULT '{}',
                                                           created_at      TIMESTAMPTZ DEFAULT NOW(),
    updated_at      TIMESTAMPTZ DEFAULT NOW()
    );

CREATE INDEX IF NOT EXISTS idx_med_products_name
    ON business_intel.med_products USING gin (to_tsvector('english', name));
CREATE INDEX IF NOT EXISTS idx_med_products_category
    ON business_intel.med_products (category);
CREATE INDEX IF NOT EXISTS idx_med_products_species
    ON business_intel.med_products USING gin (species);

-- ============================================================================
-- 2. med_retailers — Tracked pharmacies
-- ============================================================================
CREATE TABLE IF NOT EXISTS business_intel.med_retailers (
                                                            id              TEXT PRIMARY KEY,          -- 'pet_drugs_online'
                                                            name            TEXT NOT NULL,             -- 'Pet Drugs Online'
                                                            domain          TEXT NOT NULL,             -- 'petdrugsonline.co.uk'
                                                            group_name      TEXT,                      -- 'IVC Evidensia'
                                                            base_url        TEXT NOT NULL,             -- 'https://www.petdrugsonline.co.uk'
                                                            category_urls   TEXT[],                    -- URLs to crawl for product discovery
                                                            delivery_cost   NUMERIC,                  -- standard delivery cost
                                                            free_delivery_threshold NUMERIC,          -- free delivery above this amount
                                                            is_active       BOOLEAN DEFAULT true,
                                                            scrape_config   JSONB DEFAULT '{}',       -- retailer-specific extraction hints
                                                            created_at      TIMESTAMPTZ DEFAULT NOW(),
    updated_at      TIMESTAMPTZ DEFAULT NOW()
    );

-- ============================================================================
-- 3. med_retailer_listings — Product URLs per retailer
-- ============================================================================
CREATE TABLE IF NOT EXISTS business_intel.med_retailer_listings (
                                                                    id              UUID DEFAULT gen_random_uuid() PRIMARY KEY,
    retailer_id     TEXT NOT NULL REFERENCES business_intel.med_retailers(id),
    product_id      TEXT REFERENCES business_intel.med_products(id),  -- NULL until matched
    retailer_url    TEXT NOT NULL,
    retailer_product_name TEXT,               -- their name for this product
    retailer_sku    TEXT,                     -- their internal ID if available
    match_confidence NUMERIC,                 -- how confident is the product mapping
    match_method    TEXT,                     -- 'manual', 'llm', 'exact_name'
    is_active       BOOLEAN DEFAULT true,
    last_scraped_at TIMESTAMPTZ,
    last_price      NUMERIC,                  -- most recent price (denormalized)
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    updated_at      TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE (retailer_id, retailer_url)
    );

CREATE INDEX IF NOT EXISTS idx_med_listings_product
    ON business_intel.med_retailer_listings (product_id);
CREATE INDEX IF NOT EXISTS idx_med_listings_retailer
    ON business_intel.med_retailer_listings (retailer_id);
CREATE INDEX IF NOT EXISTS idx_med_listings_unmatched
    ON business_intel.med_retailer_listings (retailer_id)
    WHERE product_id IS NULL AND is_active = true;

-- ============================================================================
-- 4. med_price_snapshots — Price history
-- ============================================================================
CREATE TABLE IF NOT EXISTS business_intel.med_price_snapshots (
                                                                  id              BIGSERIAL PRIMARY KEY,
                                                                  listing_id      UUID NOT NULL REFERENCES business_intel.med_retailer_listings(id),
    product_id      TEXT REFERENCES business_intel.med_products(id),
    retailer_id     TEXT NOT NULL REFERENCES business_intel.med_retailers(id),
    size_variant    TEXT,                     -- '100ml', '180ml', '100 tabs'
    price           NUMERIC NOT NULL,
    currency        TEXT DEFAULT 'GBP',
    in_stock        BOOLEAN,
    typical_vet_price NUMERIC,               -- TVP if shown on page
    collected_at    TIMESTAMPTZ DEFAULT NOW(),
    collection_method TEXT DEFAULT 'scrape',  -- 'scrape', 'feed', 'manual'
    raw_data        JSONB DEFAULT '{}',       -- full extracted data for debugging
    UNIQUE (listing_id, size_variant, collected_at)
    );

CREATE INDEX IF NOT EXISTS idx_med_prices_product
    ON business_intel.med_price_snapshots (product_id, collected_at DESC);
CREATE INDEX IF NOT EXISTS idx_med_prices_retailer
    ON business_intel.med_price_snapshots (retailer_id, collected_at DESC);
CREATE INDEX IF NOT EXISTS idx_med_prices_latest
    ON business_intel.med_price_snapshots (listing_id, collected_at DESC);

-- ============================================================================
-- 5. med_price_current — Materialized view for quick export
-- ============================================================================
CREATE MATERIALIZED VIEW IF NOT EXISTS business_intel.med_price_current AS
SELECT DISTINCT ON (ps.listing_id, ps.size_variant)
    ps.listing_id,
    ps.product_id,
    ps.retailer_id,
    r.name AS retailer_name,
    r.group_name,
    r.domain,
    l.retailer_url,
    p.name AS product_name,
    p.brand,
    p.category,
    p.species,
    p.pack_size,
    ps.size_variant,
    ps.price,
    ps.in_stock,
    ps.typical_vet_price,
    ps.collected_at
FROM business_intel.med_price_snapshots ps
    JOIN business_intel.med_retailer_listings l ON l.id = ps.listing_id
    JOIN business_intel.med_retailers r ON r.id = ps.retailer_id
    LEFT JOIN business_intel.med_products p ON p.id = ps.product_id
WHERE ps.collected_at > NOW() - INTERVAL '14 days'
ORDER BY ps.listing_id, ps.size_variant, ps.collected_at DESC;

CREATE UNIQUE INDEX IF NOT EXISTS idx_med_price_current
    ON business_intel.med_price_current (listing_id, size_variant);

COMMIT;

