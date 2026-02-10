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