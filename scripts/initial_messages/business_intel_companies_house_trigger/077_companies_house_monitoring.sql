-- ==========================================================================
-- Companies House Enrichment Monitoring Queries
-- ==========================================================================


-- =========================================================================
-- DASHBOARD
-- =========================================================================

SELECT
    (SELECT COUNT(*) FROM business_intel.companies_house_data) as total_enriched,
    (SELECT COUNT(*) FROM business_intel.companies_house_data WHERE match_method != 'no_match') as matched,
    (SELECT COUNT(*) FROM business_intel.companies_house_data WHERE match_method = 'no_match') as no_match,
                                                                  (SELECT COUNT(*) FROM business_intel.businesses b
    JOIN business_intel.business_verticals bv ON bv.id = b.vertical_id
    LEFT JOIN business_intel.companies_house_data ch ON ch.business_id = b.id
WHERE bv.slug = 'veterinary' AND b.verification_status = 'verified'
  AND ch.business_id IS NULL) as remaining;


-- =========================================================================
-- SUCCESSION RISK — the PE targeting view
-- =========================================================================

-- Risk breakdown
SELECT succession_risk, COUNT(*)
FROM business_intel.companies_house_data
GROUP BY succession_risk
ORDER BY COUNT(*) DESC;

-- High-risk practices — prime acquisition targets
SELECT b.name, b.town, b.postcode, b.phone, b.website_url,
       ch.owner_name, ch.owner_estimated_age, ch.owner_tenure_years,
       ch.is_sole_director, ch.company_name, ch.company_number,
       ch.succession_risk
FROM business_intel.companies_house_data ch
         JOIN business_intel.businesses b ON b.id = ch.business_id
WHERE ch.succession_risk = 'high'
ORDER BY ch.owner_estimated_age DESC NULLS LAST
    LIMIT 20;

-- Already acquired practices (corporate ownership)
SELECT b.name, b.town, ch.parent_company_name, ch.parent_company_number,
       ch.company_name, ch.company_number
FROM business_intel.companies_house_data ch
         JOIN business_intel.businesses b ON b.id = ch.business_id
WHERE ch.is_corporate_owned = true
ORDER BY ch.parent_company_name, b.name;

-- Group ownership summary — which PE firms own the most
SELECT ch.parent_company_name, COUNT(*) as practices_owned
FROM business_intel.companies_house_data ch
WHERE ch.is_corporate_owned = true
  AND ch.parent_company_name IS NOT NULL
GROUP BY ch.parent_company_name
ORDER BY COUNT(*) DESC
    LIMIT 20;


-- =========================================================================
-- OWNER AGE ANALYSIS
-- =========================================================================

-- Age distribution of practice owners
SELECT
    CASE
        WHEN owner_estimated_age >= 65 THEN '65+'
        WHEN owner_estimated_age >= 60 THEN '60-64'
        WHEN owner_estimated_age >= 55 THEN '55-59'
        WHEN owner_estimated_age >= 50 THEN '50-54'
        WHEN owner_estimated_age >= 40 THEN '40-49'
        WHEN owner_estimated_age > 0 THEN 'under 40'
        ELSE 'unknown'
        END as age_band,
    COUNT(*)
FROM business_intel.companies_house_data
WHERE match_method != 'no_match'
GROUP BY age_band
ORDER BY age_band;

-- Sole directors over 55 — highest succession risk
SELECT b.name, b.town, b.postcode, b.phone,
       ch.owner_name, ch.owner_estimated_age, ch.owner_tenure_years,
       ch.company_name
FROM business_intel.companies_house_data ch
         JOIN business_intel.businesses b ON b.id = ch.business_id
WHERE ch.is_sole_director = true
  AND ch.owner_estimated_age >= 55
  AND ch.is_corporate_owned = false
ORDER BY ch.owner_estimated_age DESC;


-- =========================================================================
-- COMPANY DETAILS
-- =========================================================================

-- Recent enrichments
SELECT b.name, ch.company_name, ch.company_number,
       ch.company_status, ch.company_type,
       ch.match_confidence, ch.match_method,
       ch.enriched_at
FROM business_intel.companies_house_data ch
         JOIN business_intel.businesses b ON b.id = ch.business_id
WHERE ch.match_method != 'no_match'
ORDER BY ch.enriched_at DESC
    LIMIT 20;

-- Company status breakdown
SELECT ch.company_status, COUNT(*)
FROM business_intel.companies_house_data ch
WHERE ch.company_status IS NOT NULL
GROUP BY ch.company_status
ORDER BY COUNT(*) DESC;

-- Incorporation age (how long established)
SELECT
    CASE
        WHEN ch.incorporation_date < NOW() - INTERVAL '30 years' THEN '30+ years'
    WHEN ch.incorporation_date < NOW() - INTERVAL '20 years' THEN '20-30 years'
    WHEN ch.incorporation_date < NOW() - INTERVAL '10 years' THEN '10-20 years'
    WHEN ch.incorporation_date < NOW() - INTERVAL '5 years' THEN '5-10 years'
    WHEN ch.incorporation_date IS NOT NULL THEN 'under 5 years'
    ELSE 'unknown'
END as company_age,
    COUNT(*)
FROM business_intel.companies_house_data ch
WHERE ch.match_method != 'no_match'
GROUP BY company_age
ORDER BY company_age;


-- =========================================================================
-- MATCH QUALITY
-- =========================================================================

-- Confidence distribution
SELECT
    CASE
        WHEN match_confidence >= 0.8 THEN 'high (0.8+)'
        WHEN match_confidence >= 0.6 THEN 'medium (0.6-0.8)'
        WHEN match_confidence >= 0.4 THEN 'low (0.4-0.6)'
        ELSE 'no match'
        END as confidence_band,
    COUNT(*)
FROM business_intel.companies_house_data
GROUP BY confidence_band
ORDER BY confidence_band;

-- Match method breakdown
SELECT match_method, COUNT(*)
FROM business_intel.companies_house_data
GROUP BY match_method
ORDER BY COUNT(*) DESC;

-- Low confidence matches worth reviewing
SELECT b.name, b.postcode, ch.company_name,
       ch.company_number, ch.match_confidence, ch.match_method,
       ch.search_query
FROM business_intel.companies_house_data ch
         JOIN business_intel.businesses b ON b.id = ch.business_id
WHERE ch.match_confidence BETWEEN 0.4 AND 0.6
ORDER BY ch.match_confidence ASC
    LIMIT 20;


-- =========================================================================
-- OFFICERS / DIRECTORS
-- =========================================================================

-- View directors for a practice (example — replace business name)
-- SELECT b.name,
--        officer->>'name' as director,
--        officer->>'officer_role' as role,
--        officer->>'appointed_on' as appointed,
--        officer->>'dob_year' as birth_year,
--        officer->>'nationality' as nationality
-- FROM business_intel.companies_house_data ch
-- JOIN business_intel.businesses b ON b.id = ch.business_id,
--      jsonb_array_elements(ch.officers) as officer
-- WHERE b.name ILIKE '%oakwood%';


-- =========================================================================
-- COMBINED VIEW — practice + financials + risk
-- =========================================================================

-- Full practice intelligence report
SELECT b.name, b.town, b.postcode, b.phone, b.website_url,
       b.group_name,
       ch.company_name, ch.company_number, ch.company_status,
       ch.incorporation_date,
       ch.owner_name, ch.owner_estimated_age, ch.owner_tenure_years,
       ch.is_sole_director, ch.is_corporate_owned,
       ch.parent_company_name,
       ch.succession_risk,
       ch.match_confidence,
       bp_count.price_count,
       vpd.species_treated, vpd.num_vets, vpd.emergency_service
FROM business_intel.businesses b
         JOIN business_intel.business_verticals bv ON bv.id = b.vertical_id
         LEFT JOIN business_intel.companies_house_data ch ON ch.business_id = b.id
         LEFT JOIN business_intel.vet_practice_details vpd ON vpd.business_id = b.id
         LEFT JOIN (
    SELECT business_id, COUNT(*) as price_count
    FROM business_intel.business_prices
    WHERE is_current = true
    GROUP BY business_id
) bp_count ON bp_count.business_id = b.id
WHERE bv.slug = 'veterinary'
  AND b.verification_status = 'verified'
ORDER BY ch.succession_risk DESC NULLS LAST, ch.owner_estimated_age DESC NULLS LAST
    LIMIT 30;


============================================================================================

-- Compare business names vs search queries
SELECT b.name as business_name, ch.search_query,
       LENGTH(b.name) as name_len, LENGTH(ch.search_query) as query_len
FROM business_intel.companies_house_data ch
         JOIN business_intel.businesses b ON b.id = ch.business_id
WHERE ch.search_query != b.name
ORDER BY LENGTH(ch.search_query)
    LIMIT 15;

SELECT b.name, ch.search_query, ch.match_method
FROM business_intel.companies_house_data ch
         JOIN business_intel.businesses b ON b.id = ch.business_id
ORDER BY ch.enriched_at DESC
    LIMIT 10;
-------------------------------------------------------------------------------------------
