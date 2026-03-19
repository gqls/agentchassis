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


Accounts
-- Check what we'd get: how many matched companies have filings?
SELECT ch.company_number, ch.company_name, ch.company_status
FROM business_intel.companies_house_data ch
WHERE ch.company_number IS NOT NULL AND ch.company_number != ''
LIMIT 10;

Test with curl on a known company:

# Try Eglish Builders which has been around since 1989
curl -s -u "bd727e00-7972-4195-a576-d97faad6043f:" \
  "https://api.company-information.service.gov.uk/company/NI022532/filing-history?category=accounts&items_per_page=1" | python3 -m json.tool

 # Try Eglish Builders which has been around since 1989
curl -s -u "bd727e00-7972-4195-a576-d97faad6043f:" \
  "https://api.company-information.service.gov.uk/company/NI022532/filing-history?category=accounts&items_per_page=1" | python3 -m json.tool
{
    "filing_history_status": "filing-history-available",
    "items": [
        {
            "transaction_id": "MzQ3NTc5OTIzMWFkaXF6a2N4",
            "barcode": "XE7YRA7K",
            "type": "AA",
            "date": "2025-07-31",
            "category": "accounts",
            "description": "accounts-with-accounts-type-unaudited-abridged",
            "description_values": {
                "made_up_date": "2024-10-31"
            },
            "pages": 8,
            "links": {
                "self": "/company/NI022532/filing-history/MzQ3NTc5OTIzMWFkaXF6a2N4",
                "document_metadata": "https://document-api.company-information.service.gov.uk/document/1ZQnz8Hhz5d3vP--x104tx3onRnzxBNolcPz0S5w2Nc"
            }
        }
    ],
    "items_per_page": 1,
    "start_index": 0,
    "total_count": 38
}

The metadata shows both formats available. Fetch the iXBRL (structured, parseable):
curl -s -u "bd727e00-7972-4195-a576-d97faad6043f:" \
  -H "Accept: application/xhtml+xml" \
  "https://document-api.company-information.service.gov.uk/document/1ZQnz8Hhz5d3vP--x104tx3onRnzxBNolcPz0S5w2Nc/content" \
  -o /tmp/eglish_accounts.xhtml

# Check what financial tags are in there
grep -i "FixedAssets\|CurrentAssets\|NetAssets\|TurnoverRevenue\|ProfitLoss\|NumberEmployees\|ShareholderFunds\|Equity" /tmp/eglish_accounts.xhtml | head -20

/tmp/eglish_accounts.xhtml is an empty document:
-rw-rw-r--  1 ant  ant      0 Mar 18 13:09 eglish_accounts.xhtml

The document API redirects. Add -L to follow:
curl -s -L -u "bd727e00-7972-4195-a576-d97faad6043f:" \
  -H "Accept: application/xhtml+xml" \
  "https://document-api.company-information.service.gov.uk/document/1ZQnz8Hhz5d3vP--x104tx3onRnzxBNolcPz0S5w2Nc/content" \
  -o /tmp/eglish_accounts.xhtml

ls -la /tmp/eglish_accounts.xhtml


----
why no match
              -- What kinds of businesses are no_match?
SELECT b.group_name, COUNT(*) as no_match_count
FROM business_intel.companies_house_data ch
         JOIN business_intel.businesses b ON b.id = ch.business_id
WHERE ch.match_method = 'no_match'
GROUP BY b.group_name
ORDER BY COUNT(*) DESC
    LIMIT 15;

-- How many no_matches are corporate vs independent?
SELECT
    CASE WHEN b.group_name IN ('Independent', '') OR b.group_name IS NULL THEN 'Independent'
         ELSE 'Corporate/Group' END as ownership,
    COUNT(*) as no_match_count
FROM business_intel.companies_house_data ch
         JOIN business_intel.businesses b ON b.id = ch.business_id
WHERE ch.match_method = 'no_match'
GROUP BY 1;

-- Pick 5 independent no_matches with their search queries
SELECT b.name, b.postcode, ch.search_query
FROM business_intel.companies_house_data ch
         JOIN business_intel.businesses b ON b.id = ch.business_id
WHERE ch.match_method = 'no_match'
  AND b.group_name = 'Independent'
ORDER BY RANDOM()
    LIMIT 5;

# Replace with actual names from the query above
curl -s -u "bd727e00-7972-4195-a576-d97faad6043f:" \
  "https://api.company-information.service.gov.uk/search/companies?q=PRACTICE_NAME_HERE&items_per_page=5" | python3 -m json.tool | head -40

                                      name                 | postcode |            search_query
-------------------------------------+----------+-------------------------------------
 Catley Cross Veterinary Clinic      | CO9 2PE  | Catley Cross
 Coastal Veterinary Group            | PE31 7NY | Coastal
 Churchill Vets                      | E4 6AG   | Churchill
 Eagle Vets (Wingham)                | CT3 1AR  | Eagle Vets (Wingham)
 Cameron & Greig Veterinary Practice | KY13 9XU | Cameron & Greig Veterinary Practice
 Eagle Vets (Wingham)                       | CT3 1AR  | Eagle Vets (Wingham)
 Endell Veterinary Group                    | SP1 3UH  | Endell
 Coastal Veterinary Group                   | BT40 1BA | Coastal
 Clarendon Veterinary Centre (West Wickham) | BR4 0PY  | Clarendon Veterinary Centre (West Wickham)
 Ease Vets                                  | DE24 8GZ | Ease Vets

# Replace with actual names from the query above
curl -s -u "bd727e00-7972-4195-a576-d97faad6043f:" \
  "https://api.company-information.service.gov.uk/search/companies?q="PRACTICE_NAME_HERE"&items_per_page=5" | python3 -m json.tool | head -40

# Replace with actual names from the query above
curl -s -u "bd727e00-7972-4195-a576-d97faad6043f:" \
  "https://api.company-information.service.gov.uk/search/companies?q="Catley Cross Veterinary Clinic"&items_per_page=5" | python3 -m json.tool | head -40

# Replace with actual names from the query above
curl -s -u "bd727e00-7972-4195-a576-d97faad6043f:" \
  "https://api.company-information.service.gov.uk/search/companies?q="Coastal Veterinary Group"&items_per_page=5" | python3 -m json.tool | head -40

# Replace with actual names from the query above
curl -s -u "bd727e00-7972-4195-a576-d97faad6043f:" \
  "https://api.company-information.service.gov.uk/search/companies?q="Churchill Vets "&items_per_page=5" | python3 -m json.tool | head -40

# Replace with actual names from the query above
curl -s -u "bd727e00-7972-4195-a576-d97faad6043f:" \
  "https://api.company-information.service.gov.uk/search/companies?q="Eagle Vets (Wingham)"&items_per_page=5" | python3 -m json.tool | head -40

# Replace with actual names from the query above
curl -s -u "bd727e00-7972-4195-a576-d97faad6043f:" \
  "https://api.company-information.service.gov.uk/search/companies?q="Endell Veterinary Group "&items_per_page=5" | python3 -m json.tool | head -40

# Replace with actual names from the query above
curl -s -u "bd727e00-7972-4195-a576-d97faad6043f:" \
  "https://api.company-information.service.gov.uk/search/companies?q="Coastal Veterinary Group"&items_per_page=5" | python3 -m json.tool | head -40

# Replace with actual names from the query above
curl -s -u "bd727e00-7972-4195-a576-d97faad6043f:" \
  "https://api.company-information.service.gov.uk/search/companies?q="Clarendon Veterinary Centre (West Wickham)"&items_per_page=5" | python3 -m json.tool | head -40

# Replace with actual names from the query above
curl -s -u "bd727e00-7972-4195-a576-d97faad6043f:" \
  "https://api.company-information.service.gov.uk/search/companies?q="Ease Vets"&items_per_page=5" | python3 -m json.tool | head -40

--- results
curl -s -u "bd727e00-7972-4195-a576-d97faad6043f:" \
  "https://api.company-information.service.gov.uk/search/companies?q="Catley Cross Veterinary Clinic"&items_per_page=5" | python3 -m json.tool | head -40
{
    "items": [
        {
            "kind": "searchresults#company",
            "description_identifier": [
                "dissolved-on"
            ],
            "company_status": "dissolved",
            "date_of_creation": "2011-05-23",
            "date_of_cessation": "2013-05-21",
            "company_type": "ltd",
            "company_number": "07642599",
            "address": {
                "address_line_1": "Alexandra Park Road",
                "country": "United Kingdom",
                "locality": "London",
                "postal_code": "N10 2AB",
                "premises": "22a"
            },
            "title": "CATLEY LIMITED",
            "address_snippet": "22a Alexandra Park Road, London, United Kingdom, N10 2AB",
            "description": "07642599 - Dissolved on 21 May 2013",
            "links": {
                "self": "/company/07642599"
            },
            "snippet": "",
            "matches": {
                "snippet": []
            }
        },
        {
            "kind": "searchresults#company",
            "description_identifier": [
                "incorporated-on"
            ],
            "company_status": "active",
            "date_of_creation": "2013-02-22",
            "company_type": "ltd",
            "company_number": "08416241",
            "address": {
ant@ant-XPS-15-9500:~/projects/agentchassis$ curl -s -u "bd727e00-7972-4195-a576-d97faad6043f:" \
  "https://api.company-information.service.gov.uk/search/companies?q="Coastal Veterinary Group"&items_per_page=5" | python3 -m json.tool | head -40
{
    "items": [
        {
            "kind": "searchresults#company",
            "description_identifier": [
                "incorporated-on"
            ],
            "company_status": "active",
            "date_of_creation": "2007-07-12",
            "company_type": "ltd",
            "company_number": "06310902",
            "address": {
                "address_line_1": "Addison Road",
                "locality": "Sudbury",
                "postal_code": "CO10 2YW",
                "premises": "C/O A&B Glass Co Ltd",
                "region": "Suffolk"
            },
            "title": "COASTAL LIMITED",
            "address_snippet": "C/O A&B Glass Co Ltd, Addison Road, Sudbury, Suffolk, CO10 2YW",
            "description": "06310902 - Incorporated on 12 July 2007",
            "links": {
                "self": "/company/06310902"
            },
            "snippet": "",
            "matches": {
                "snippet": []
            }
        },
        {
            "kind": "searchresults#company",
            "description_identifier": [
                "incorporated-on"
            ],
            "company_status": "active",
            "date_of_creation": "2025-05-06",
            "company_type": "ltd",
            "company_number": "16430629",
            "address": {
                "address_line_1": "Wenlock Road",
ant@ant-XPS-15-9500:~/projects/agentchassis$ curl -s -u "bd727e00-7972-4195-a576-d97faad6043f:" \
  "https://api.company-information.service.gov.uk/search/companies?q="Coastal Veterinary Group"&items_per_page=5" | python3 -m json.tool | head -40
{
    "items": [
        {
            "kind": "searchresults#company",
            "description_identifier": [
                "incorporated-on"
            ],
            "company_status": "active",
            "date_of_creation": "2007-07-12",
            "company_type": "ltd",
            "company_number": "06310902",
            "address": {
                "address_line_1": "Addison Road",
                "locality": "Sudbury",
                "postal_code": "CO10 2YW",
                "premises": "C/O A&B Glass Co Ltd",
                "region": "Suffolk"
            },
            "title": "COASTAL LIMITED",
            "address_snippet": "C/O A&B Glass Co Ltd, Addison Road, Sudbury, Suffolk, CO10 2YW",
            "description": "06310902 - Incorporated on 12 July 2007",
            "links": {
                "self": "/company/06310902"
            },
            "snippet": "",
            "matches": {
                "snippet": []
            }
        },
        {
            "kind": "searchresults#company",
            "description_identifier": [
                "incorporated-on"
            ],
            "company_status": "active",
            "date_of_creation": "2025-05-06",
            "company_type": "ltd",
            "company_number": "16430629",
            "address": {
                "address_line_1": "Wenlock Road",
--- results

=======================================================================================
API
=======================================================================================
# Try the advanced search API
curl -s -u "bd727e00-7972-4195-a576-d97faad6043f:" \
  "https://api.company-information.service.gov.uk/advanced-search/companies?sic_codes=75000&size=5" | python3 -m json.tool | head -40

# How many SIC 75000 companies exist? And check pagination params
curl -s -u "bd727e00-7972-4195-a576-d97faad6043f:" \
  "https://api.company-information.service.gov.uk/advanced-search/companies?sic_codes=75000&size=1" | python3 -m json.tool | grep -E "total_results|hits|start_index|items_per_page"

# Try filtering for active only
curl -s -u "bd727e00-7972-4195-a576-d97faad6043f:" \
  "https://api.company-information.service.gov.uk/advanced-search/companies?sic_codes=75000&company_status=active&size=1" | python3 -m json.tool | grep -E "total_results|hits"

# Check pagination works
curl -s -u "bd727e00-7972-4195-a576-d97faad6043f:" \
  "https://api.company-information.service.gov.uk/advanced-search/companies?sic_codes=75000&company_status=active&size=5&start_index=0" | python3 -m json.tool | head -10

max page size
curl -s -u "bd727e00-7972-4195-a576-d97faad6043f:" \
  "https://api.company-information.service.gov.uk/advanced-search/companies?sic_codes=75000&company_status=active&size=100&start_index=0" | python3 -m json.tool | python3 -c "import json,sys; d=json.load(sys.stdin); print(f'items: {len(d.get(\"items\",[]))}, hits: {d.get(\"hits\")}')"












