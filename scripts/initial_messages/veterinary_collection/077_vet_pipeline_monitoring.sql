-- ==========================================================================
-- Vet Pipeline Monitoring Queries
-- ==========================================================================
-- Quick reference for checking pipeline status, debugging stalls,
-- and inspecting collected data.
-- ==========================================================================


-- =========================================================================
-- DASHBOARD — run this to get overall status at a glance
-- =========================================================================

SELECT
    (SELECT COUNT(*) FROM business_intel.collection_tasks WHERE status = 'completed') as tasks_done,
    (SELECT COUNT(*) FROM business_intel.collection_tasks WHERE status = 'in_progress') as tasks_active,
    (SELECT COUNT(*) FROM business_intel.collection_tasks WHERE status = 'pending') as tasks_pending,
    (SELECT COUNT(*) FROM business_intel.businesses b
                              JOIN business_intel.business_verticals bv ON bv.id = b.vertical_id
     WHERE bv.slug = 'veterinary' AND b.verification_status = 'verified') as verified,
    (SELECT COUNT(*) FROM business_intel.business_prices WHERE is_current = TRUE) as current_prices;


-- =========================================================================
-- SWEEP PROGRESS
-- =========================================================================

-- Areas swept vs remaining
SELECT
    COUNT(*) FILTER (WHERE last_swept_at IS NOT NULL) as swept,
    COUNT(*) FILTER (WHERE last_swept_at IS NULL) as unswept
FROM business_intel.search_areas;

-- Most recent sweep activity
SELECT district_code, area_name, last_swept_at, candidates_found
FROM business_intel.search_areas
WHERE last_swept_at IS NOT NULL
ORDER BY last_swept_at DESC
    LIMIT 10;

-- Discovery candidates by status
SELECT status, COUNT(*)
FROM business_intel.discovery_candidates
GROUP BY status;


-- =========================================================================
-- BUSINESS STATUS
-- =========================================================================

-- Verification status breakdown
SELECT verification_status, COUNT(*)
FROM business_intel.businesses b
         JOIN business_intel.business_verticals bv ON bv.id = b.vertical_id
WHERE bv.slug = 'veterinary'
GROUP BY verification_status;

-- Collection tasks breakdown
SELECT status, COUNT(*)
FROM business_intel.collection_tasks
GROUP BY status;

-- Recently verified businesses
SELECT b.name, b.website_url, b.town, b.postcode,
       b.phone, b.group_name, b.verification_status,
       b.confidence_score, b.updated_at
FROM business_intel.businesses b
         JOIN business_intel.business_verticals bv ON bv.id = b.vertical_id
WHERE bv.slug = 'veterinary'
  AND b.verification_status = 'verified'
ORDER BY b.updated_at DESC
    LIMIT 10;

-- Businesses with missing data (postcode, no URL, etc.)
SELECT
    COUNT(*) FILTER (WHERE postcode IS NULL OR postcode = '') as missing_postcode,
    COUNT(*) FILTER (WHERE website_url IS NULL OR website_url = '') as missing_website,
    COUNT(*) FILTER (WHERE phone IS NULL OR phone = '') as missing_phone
FROM business_intel.businesses b
         JOIN business_intel.business_verticals bv ON bv.id = b.vertical_id
WHERE bv.slug = 'veterinary'
  AND b.verification_status = 'verified';


-- =========================================================================
-- PRICES
-- =========================================================================

-- Price count
SELECT COUNT(*) FROM business_intel.business_prices WHERE is_current = TRUE;

-- Recent prices with business names
SELECT b.name, bp.service_category, bp.service_name,
       bp.price_gbp, bp.price_qualifier
FROM business_intel.business_prices bp
         JOIN business_intel.businesses b ON b.id = bp.business_id
WHERE bp.is_current = TRUE
ORDER BY b.updated_at DESC
    LIMIT 20;

-- Price summary by category
SELECT bp.service_category, COUNT(*),
       ROUND(AVG(bp.price_gbp), 2) as avg_price,
       MIN(bp.price_gbp) as min_price,
       MAX(bp.price_gbp) as max_price
FROM business_intel.business_prices bp
WHERE bp.is_current = TRUE AND bp.price_gbp IS NOT NULL
GROUP BY bp.service_category
ORDER BY COUNT(*) DESC;


-- =========================================================================
-- VET-SPECIFIC DATA
-- =========================================================================

-- Species treated
SELECT b.name, vpd.species_treated
FROM business_intel.vet_practice_details vpd
         JOIN business_intel.businesses b ON b.id = vpd.business_id
WHERE vpd.species_treated IS NOT NULL
  AND array_length(vpd.species_treated, 1) > 0
ORDER BY b.updated_at DESC
    LIMIT 10;

-- Practice details
SELECT b.name, vpd.emergency_service, vpd.num_vets,
       vpd.head_vet_name, vpd.accepting_new_clients,
       vpd.has_own_lab, vpd.has_imaging
FROM business_intel.vet_practice_details vpd
         JOIN business_intel.businesses b ON b.id = vpd.business_id
ORDER BY b.updated_at DESC
    LIMIT 10;

-- Contact details
SELECT b.name, bcd.contact_type, bcd.label, bcd.value, bcd.is_primary
FROM business_intel.business_contact_details bcd
         JOIN business_intel.businesses b ON b.id = bcd.business_id
ORDER BY b.name, bcd.contact_type, bcd.is_primary DESC
    LIMIT 20;


-- =========================================================================
-- SCHEDULER STATUS
-- =========================================================================

-- Are scheduled tasks firing?
SELECT name, last_triggered_at,
       EXTRACT(EPOCH FROM (NOW() - last_triggered_at))::int as seconds_ago,
    enabled, concurrency_group, max_concurrent, timeout_seconds
FROM scheduled_tasks
WHERE name LIKE 'vet-%' OR name = 'stale-orchestration-reaper'
ORDER BY name;


-- =========================================================================
-- ACTIVE ORCHESTRATIONS
-- =========================================================================

-- Currently running
SELECT owner_agent_type, status, current_step,
       EXTRACT(EPOCH FROM (NOW() - last_activity))::int as idle_seconds
FROM orchestration_states
WHERE status NOT IN ('COMPLETED', 'FAILED')
  AND created_at > NOW() - INTERVAL '2 hours'
ORDER BY created_at DESC
    LIMIT 10;

-- Stuck orchestrations (AWAITING_RESPONSES for >20 min)
SELECT orchestration_id, owner_agent_type, current_step,
       updated_at,
       EXTRACT(EPOCH FROM (NOW() - last_activity))::int as idle_seconds
FROM orchestration_states
WHERE status = 'AWAITING_RESPONSES'
  AND last_activity < NOW() - INTERVAL '20 minutes'
ORDER BY last_activity ASC;

-- Check a specific orchestration
-- SELECT orchestration_id, status, current_step, error, updated_at,
--        responses_topic, requests_topic
-- FROM orchestration_states
-- WHERE orchestration_id = '<ORCHESTRATION_ID>';

-- Check awaited requests for an orchestration
-- SELECT key as request_id,
--        value->>'step_name' as step_name,
--        value->>'timeout_at' as timeout_at,
--        value->>'responses_topic' as responses_topic
-- FROM orchestration_states,
--      jsonb_each(awaited_requests)
-- WHERE orchestration_id = '<ORCHESTRATION_ID>';

-- Recent verifier outcomes
SELECT status, COUNT(*)
FROM orchestration_states
WHERE owner_agent_type = 'vet-practice-verifier'
  AND created_at > NOW() - INTERVAL '6 hours'
GROUP BY status;

-- Verifier errors
SELECT error, COUNT(*)
FROM orchestration_states
WHERE owner_agent_type = 'vet-practice-verifier'
  AND status = 'FAILED'
  AND created_at > NOW() - INTERVAL '6 hours'
GROUP BY error
ORDER BY COUNT(*) DESC;


-- =========================================================================
-- AWAITED REQUESTS
-- =========================================================================

-- Expired awaited requests (should have been caught by reaper)
SELECT request_id, orchestration_id, step_name,
       timeout_at, retry_version, status
FROM awaited_requests
WHERE status = 'waiting'
  AND timeout_at < NOW() - INTERVAL '5 minutes'
ORDER BY timeout_at ASC
    LIMIT 10;


-- =========================================================================
-- MANUAL FIXES
-- =========================================================================

-- Reset orphaned in_progress tasks
-- UPDATE business_intel.collection_tasks
-- SET status = 'pending', started_at = NULL, orchestration_id = NULL
-- WHERE status = 'in_progress';

-- Fail stuck orchestrations
-- UPDATE orchestration_states
-- SET status = 'FAILED', error = 'manual reset', updated_at = NOW()
-- WHERE status = 'AWAITING_RESPONSES'
--   AND last_activity < NOW() - INTERVAL '30 minutes';

-- Dismiss non-UK businesses
-- UPDATE business_intel.businesses
-- SET verification_status = 'dismissed'
-- WHERE id IN (
--     SELECT b.id FROM business_intel.businesses b
--     JOIN business_intel.business_verticals bv ON bv.id = b.vertical_id
--     WHERE bv.slug = 'veterinary'
--       AND b.verification_status = 'pending'
--       AND (b.postcode IS NULL OR b.postcode = '')
--       AND (b.website_url NOT LIKE '%.co.uk%' AND b.website_url NOT LIKE '%.uk%')
-- );

-- Cancel tasks for dismissed businesses
-- UPDATE business_intel.collection_tasks ct
-- SET status = 'cancelled'
-- FROM business_intel.businesses b
-- WHERE ct.business_id = b.id
--   AND b.verification_status = 'dismissed'
--   AND ct.status = 'pending';

-- Disable vet scheduled tasks
-- UPDATE scheduled_tasks SET enabled = false WHERE name LIKE 'vet-%';