-- business_contact_details: stores individual contact items with type labels
-- Handles: multiple phone numbers, multiple emails, social links, opening hours etc.
-- The businesses.phone and businesses.email columns remain as "primary" quick-reference values.

CREATE TABLE IF NOT EXISTS business_intel.business_contact_details (
                                                                       id              UUID DEFAULT gen_random_uuid() PRIMARY KEY,
    business_id     UUID NOT NULL REFERENCES business_intel.businesses(id) ON DELETE CASCADE,
    contact_type    TEXT NOT NULL,       -- 'phone', 'email', 'social', 'fax', 'website'
    label           TEXT,                -- 'main', 'emergency', 'referrals', 'fax', 'mobile', 'reception', etc.
    value           TEXT NOT NULL,       -- the actual phone number, email, URL
    is_primary      BOOLEAN DEFAULT FALSE,
    source          TEXT,                -- 'website_scrape', 'search_results', 'manual', 'seed_import'
    source_url      TEXT,                -- where it was found
    confidence      NUMERIC(3,2) DEFAULT 0.5,
    verified_at     TIMESTAMPTZ,
    orchestration_id UUID,              -- which orchestration discovered this
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    updated_at      TIMESTAMPTZ DEFAULT NOW()
    );

CREATE INDEX IF NOT EXISTS idx_bcd_business_id ON business_intel.business_contact_details(business_id);
CREATE INDEX IF NOT EXISTS idx_bcd_type ON business_intel.business_contact_details(contact_type);
CREATE UNIQUE INDEX IF NOT EXISTS idx_bcd_unique_value
    ON business_intel.business_contact_details(business_id, contact_type, value);

-- Verify
SELECT column_name, data_type FROM information_schema.columns
WHERE table_schema = 'business_intel' AND table_name = 'business_contact_details'
ORDER BY ordinal_position;