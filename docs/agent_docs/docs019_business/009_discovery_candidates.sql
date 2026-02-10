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


-- fixes
ALTER TABLE business_intel.discovery_candidates
    ADD CONSTRAINT uq_discovery_candidates_source_url UNIQUE (source_url);

ALTER TABLE business_intel.discovery_candidates
    ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ DEFAULT now();


-- fix all
-- ==========================================================================
-- Migration: Fix discovery_candidates for area sweep
-- Run against: clients_db
--
-- Fixes:
--   1. Add proper UNIQUE constraint on source_url (partial index doesn't
--      satisfy ON CONFLICT). Drop the old partial index first.
--   2. Add updated_at column (referenced in ON CONFLICT SET clause)
--   3. Add detected_group and is_independent for chain tagging
-- ==========================================================================

-- Fix 1: Replace partial unique index with proper UNIQUE constraint
-- The existing partial index (WHERE source_url IS NOT NULL) doesn't satisfy
-- ON CONFLICT (source_url). We need a real constraint.
DROP INDEX IF EXISTS business_intel.idx_dc_unique_source;

ALTER TABLE business_intel.discovery_candidates
    ADD CONSTRAINT uq_discovery_candidates_source_url UNIQUE (source_url);

-- Fix 2: Add updated_at column
ALTER TABLE business_intel.discovery_candidates
    ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ DEFAULT NOW();

-- Backfill updated_at from created_at for existing rows
UPDATE business_intel.discovery_candidates
SET updated_at = created_at
WHERE updated_at IS NULL;

-- Fix 3: Add group detection columns
ALTER TABLE business_intel.discovery_candidates
    ADD COLUMN IF NOT EXISTS detected_group TEXT,
    ADD COLUMN IF NOT EXISTS is_independent BOOLEAN;

-- Index for group queries
CREATE INDEX IF NOT EXISTS idx_dc_detected_group
    ON business_intel.discovery_candidates (detected_group)
    WHERE detected_group IS NOT NULL;

-- Verify
SELECT column_name, data_type
FROM information_schema.columns
WHERE table_schema = 'business_intel'
  AND table_name = 'discovery_candidates'
ORDER BY ordinal_position;

---


-- promote candidates
-- ==========================================================================
-- Promote discovery candidates to businesses table
-- Run against: clients_db
--
-- Inserts pending candidates into businesses, updates candidates with
-- promoted_business_id. Skips candidates that already share a website_url
-- with an existing business.
--
-- After running, use bulk_vet_verify.sh or verify_promoted.sh to verify them.
-- ==========================================================================

-- Show what we're working with
SELECT status, COUNT(*),
       COUNT(*) FILTER (WHERE detected_group IS NOT NULL) AS group_affiliated
FROM business_intel.discovery_candidates
GROUP BY status ORDER BY status;

-- Step 1: Get veterinary vertical_id
-- (needed for the businesses INSERT)
DO $$
DECLARE
v_vertical_id UUID;
    v_candidate RECORD;
    v_business_id UUID;
    v_promoted INT := 0;
    v_skipped_dup INT := 0;
    v_skipped_dir INT := 0;
    v_existing_id UUID;
BEGIN
    -- Get veterinary vertical
SELECT id INTO v_vertical_id
FROM business_intel.business_verticals
WHERE slug = 'veterinary';

IF v_vertical_id IS NULL THEN
        RAISE EXCEPTION 'Veterinary vertical not found';
END IF;

    RAISE NOTICE 'Veterinary vertical_id: %', v_vertical_id;

    -- Loop through pending candidates
FOR v_candidate IN
SELECT id, name, website_url, postcode, detected_group, is_independent,
       address_snippet
FROM business_intel.discovery_candidates
WHERE status = 'pending'
ORDER BY created_at ASC
    LOOP
        -- Skip candidates without a website_url (from directory listings)
        -- These need enrichment from the directory-scraper agent first
        IF v_candidate.website_url IS NULL THEN
UPDATE business_intel.discovery_candidates
SET status = 'needs_enrichment',
    notes = 'From directory listing, no website URL yet',
    updated_at = NOW()
WHERE id = v_candidate.id;

v_skipped_dir := v_skipped_dir + 1;
CONTINUE;
END IF;

        -- Skip if name looks like a directory page (not a practice)
        IF v_candidate.name ILIKE '%vets near%'
           OR v_candidate.name ILIKE '%veterinarians near%'
           OR v_candidate.name ILIKE '%VETS DIRECTORY%'
           OR v_candidate.name ILIKE 'THE BEST%' THEN

UPDATE business_intel.discovery_candidates
SET status = 'dismissed',
    notes = 'Auto-dismissed: directory listing title',
    reviewed_at = NOW(),
    updated_at = NOW()
WHERE id = v_candidate.id;

v_skipped_dir := v_skipped_dir + 1;
CONTINUE;
END IF;

        -- Skip if website_url already exists in businesses
SELECT id INTO v_existing_id
FROM business_intel.businesses
WHERE website_url ILIKE v_candidate.website_url || '%'
           OR website_url ILIKE 'https://www.' || REPLACE(v_candidate.website_url, 'https://', '') || '%'
        LIMIT 1;

IF v_existing_id IS NOT NULL THEN
UPDATE business_intel.discovery_candidates
SET status = 'matched',
    matched_business_id = v_existing_id,
    match_method = 'website_url',
    match_confidence = 0.95,
    reviewed_at = NOW(),
    updated_at = NOW()
WHERE id = v_candidate.id;

v_skipped_dup := v_skipped_dup + 1;
CONTINUE;
END IF;

        -- Also skip if another candidate with the same website_url was already promoted
SELECT promoted_business_id INTO v_existing_id
FROM business_intel.discovery_candidates
WHERE website_url = v_candidate.website_url
  AND status = 'promoted'
  AND promoted_business_id IS NOT NULL
    LIMIT 1;

IF v_existing_id IS NOT NULL THEN
UPDATE business_intel.discovery_candidates
SET status = 'matched',
    matched_business_id = v_existing_id,
    match_method = 'website_url_candidate_dup',
    match_confidence = 0.95,
    reviewed_at = NOW(),
    updated_at = NOW()
WHERE id = v_candidate.id;

v_skipped_dup := v_skipped_dup + 1;
CONTINUE;
END IF;

        -- Insert into businesses
INSERT INTO business_intel.businesses (
    name, website_url, postcode, country,
    group_name, is_independent,
    vertical_id, business_type,
    verification_status, is_active,
    created_at
) VALUES (
             v_candidate.name,
             v_candidate.website_url,
             v_candidate.postcode,
             'GB',
             v_candidate.detected_group,
             COALESCE(v_candidate.is_independent, true),
             v_vertical_id,
             'veterinary_practice',
             'pending',
             true,
             NOW()
         )
    RETURNING id INTO v_business_id;

-- Update candidate with promoted status
UPDATE business_intel.discovery_candidates
SET status = 'promoted',
    promoted_business_id = v_business_id,
    reviewed_at = NOW(),
    updated_at = NOW()
WHERE id = v_candidate.id;

v_promoted := v_promoted + 1;
END LOOP;

    RAISE NOTICE '';
    RAISE NOTICE 'Promotion complete:';
    RAISE NOTICE '  Promoted:              %', v_promoted;
    RAISE NOTICE '  Skipped (duplicate):   %', v_skipped_dup;
    RAISE NOTICE '  Skipped (directory/no URL): %', v_skipped_dir;
END $$;

-- Summary after promotion
SELECT status, COUNT(*)
FROM business_intel.discovery_candidates
GROUP BY status ORDER BY status;

-- Show newly promoted businesses
SELECT b.id, b.name, b.website_url, b.postcode, b.group_name, b.verification_status
FROM business_intel.businesses b
WHERE b.verification_status = 'pending'
  AND b.created_at >= NOW() - INTERVAL '5 minutes'
ORDER BY b.created_at DESC;