-- ==========================================================================
-- Migration: observation tracking, search caching, and discovery candidates
-- ==========================================================================

-- 1. Search result cache
-- Stores raw search results for later mining by discovery process
CREATE TABLE IF NOT EXISTS business_intel.search_result_cache (
                                                                  id                UUID DEFAULT gen_random_uuid() PRIMARY KEY,
    business_id       UUID REFERENCES business_intel.businesses(id) ON DELETE SET NULL,
    query             TEXT NOT NULL,
    results_json      JSONB NOT NULL,          -- the raw results array from the search adapter
    provider          TEXT,                     -- 'firecrawl', 'duckduckgo', etc.
    result_count      INT DEFAULT 0,
    searched_at       TIMESTAMPTZ DEFAULT NOW(),
    orchestration_id  UUID,
    discovery_scanned BOOLEAN DEFAULT FALSE,    -- has the discovery process looked at this yet?
    discovery_scanned_at TIMESTAMPTZ
    );

CREATE INDEX IF NOT EXISTS idx_src_business_id ON business_intel.search_result_cache(business_id);
CREATE INDEX IF NOT EXISTS idx_src_not_scanned ON business_intel.search_result_cache(discovery_scanned) WHERE discovery_scanned = FALSE;
CREATE INDEX IF NOT EXISTS idx_src_searched_at ON business_intel.search_result_cache(searched_at DESC);

-- 2. Discovery candidates
-- Practices found in search results that don't match existing records
CREATE TABLE IF NOT EXISTS business_intel.discovery_candidates (
                                                                   id                  UUID DEFAULT gen_random_uuid() PRIMARY KEY,
    name                TEXT,
    website_url         TEXT,
    address_snippet     TEXT,                   -- whatever address info was in the search snippet
    phone               TEXT,
    postcode            TEXT,                   -- extracted if possible
    source_business_id  UUID REFERENCES business_intel.businesses(id) ON DELETE SET NULL,
    source_query        TEXT,                   -- the search query that found this
    source_url          TEXT,                   -- the search result URL
    matched_business_id UUID REFERENCES business_intel.businesses(id) ON DELETE SET NULL,
    match_confidence    NUMERIC(3,2) DEFAULT 0,
    match_method        TEXT,                   -- 'website_url', 'postcode_name', 'phone', etc.
    status              TEXT DEFAULT 'pending', -- pending, matched, promoted, dismissed
    promoted_business_id UUID REFERENCES business_intel.businesses(id) ON DELETE SET NULL,
    created_at          TIMESTAMPTZ DEFAULT NOW(),
    reviewed_at         TIMESTAMPTZ,
    notes               TEXT
    );

CREATE INDEX IF NOT EXISTS idx_dc_status ON business_intel.discovery_candidates(status);
CREATE INDEX IF NOT EXISTS idx_dc_website ON business_intel.discovery_candidates(website_url);
CREATE INDEX IF NOT EXISTS idx_dc_postcode ON business_intel.discovery_candidates(postcode);
-- Prevent duplicate candidates from the same source URL
CREATE UNIQUE INDEX IF NOT EXISTS idx_dc_unique_source
    ON business_intel.discovery_candidates(source_url) WHERE source_url IS NOT NULL;

-- 3. Temporal tracking on business_contact_details
ALTER TABLE business_intel.business_contact_details
    ADD COLUMN IF NOT EXISTS first_seen_at TIMESTAMPTZ DEFAULT NOW(),
    ADD COLUMN IF NOT EXISTS last_confirmed_at TIMESTAMPTZ DEFAULT NOW(),
    ADD COLUMN IF NOT EXISTS check_count INT DEFAULT 1,
    ADD COLUMN IF NOT EXISTS missed_count INT DEFAULT 0,
    ADD COLUMN IF NOT EXISTS is_stale BOOLEAN DEFAULT FALSE;

-- 4. Temporal tracking on business_prices
ALTER TABLE business_intel.business_prices
    ADD COLUMN IF NOT EXISTS first_seen_at TIMESTAMPTZ DEFAULT NOW(),
    ADD COLUMN IF NOT EXISTS last_confirmed_at TIMESTAMPTZ DEFAULT NOW(),
    ADD COLUMN IF NOT EXISTS check_count INT DEFAULT 1,
    ADD COLUMN IF NOT EXISTS missed_count INT DEFAULT 0,
    ADD COLUMN IF NOT EXISTS is_stale BOOLEAN DEFAULT FALSE;

-- 5. Temporal tracking on vet_practice_details
ALTER TABLE business_intel.vet_practice_details
    ADD COLUMN IF NOT EXISTS last_checked_at TIMESTAMPTZ DEFAULT NOW(),
    ADD COLUMN IF NOT EXISTS check_count INT DEFAULT 1;

-- 6. Verification cycle tracking on businesses
-- Lets us know how many times this business has been through the pipeline
ALTER TABLE business_intel.businesses
    ADD COLUMN IF NOT EXISTS verification_count INT DEFAULT 0,
    ADD COLUMN IF NOT EXISTS first_verified_at TIMESTAMPTZ;

-- Backfill first_verified_at for already-verified businesses
UPDATE business_intel.businesses
SET first_verified_at = last_verified_at,
    verification_count = 1
WHERE verification_status = 'verified'
  AND first_verified_at IS NULL
  AND last_verified_at IS NOT NULL;

-- 8. View: stale data report
-- Shows contact details not confirmed in the last 3 checks
CREATE OR REPLACE VIEW business_intel.stale_contacts AS
SELECT
    b.name AS business_name,
    b.postcode,
    bcd.contact_type,
    bcd.label,
    bcd.value,
    bcd.last_confirmed_at,
    bcd.check_count,
    bcd.missed_count,
    bcd.is_stale,
    b.verification_count AS business_check_count,
    AGE(NOW(), bcd.last_confirmed_at) AS time_since_confirmed
FROM business_intel.business_contact_details bcd
         JOIN business_intel.businesses b ON b.id = bcd.business_id
WHERE bcd.missed_count >= 3 OR bcd.is_stale = TRUE
ORDER BY bcd.missed_count DESC, bcd.last_confirmed_at ASC;

-- 9. View: discovery pipeline status
CREATE OR REPLACE VIEW business_intel.discovery_summary AS
SELECT
    status,
    COUNT(*) AS count,
    COUNT(DISTINCT website_url) AS unique_websites,
    MIN(created_at) AS earliest,
    MAX(created_at) AS latest
FROM business_intel.discovery_candidates
GROUP BY status;

-- 10. View: verification progress (used by bulk script monitoring)
CREATE OR REPLACE VIEW business_intel.verification_progress AS
SELECT
    COALESCE(b.verification_count, 0) AS times_checked,
    COUNT(*) AS businesses,
    COUNT(*) FILTER (WHERE b.verification_status = 'verified') AS verified,
    COUNT(*) FILTER (WHERE b.phone IS NOT NULL) AS have_phone,
    COUNT(*) FILTER (WHERE b.email IS NOT NULL) AS have_email,
    ROUND(AVG(b.confidence_score)::numeric, 2) AS avg_confidence,
    MIN(b.last_verified_at) AS oldest_check,
    MAX(b.last_verified_at) AS newest_check
FROM business_intel.businesses b
         JOIN business_intel.business_verticals bv ON bv.id = b.vertical_id
WHERE bv.slug = 'veterinary' AND b.is_active = true
GROUP BY COALESCE(b.verification_count, 0)
ORDER BY times_checked;

-- 11. View: discovery stats (used by bulk script monitoring)
CREATE OR REPLACE VIEW business_intel.discovery_stats AS
SELECT
    (SELECT COUNT(*) FROM business_intel.search_result_cache) AS searches_cached,
    (SELECT SUM(result_count) FROM business_intel.search_result_cache) AS total_results_cached,
    (SELECT COUNT(*) FROM business_intel.search_result_cache WHERE NOT discovery_scanned) AS unscanned_searches,
    (SELECT COUNT(*) FROM business_intel.discovery_candidates) AS total_candidates,
    (SELECT COUNT(*) FROM business_intel.discovery_candidates WHERE status = 'pending') AS pending_review,
    (SELECT COUNT(*) FROM business_intel.discovery_candidates WHERE status = 'promoted') AS promoted,
    (SELECT COUNT(*) FROM business_intel.discovery_candidates WHERE status = 'dismissed') AS dismissed;

-- Verify tables created
SELECT table_name FROM information_schema.tables
WHERE table_schema = 'business_intel'
ORDER BY table_name;