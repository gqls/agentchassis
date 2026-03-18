# Companies House Enrichment — Plan

## Objective

Add a Companies House lookup step to the vet-practice-verifier workflow to enrich business records with financial data, employee counts, and company registration details.

## What Companies House Gives Us

The Companies House API (free, no auth required for basic data, API key for higher rate limits) provides:

**Company Profile:**
- Registered company name, number, status (active/dissolved)
- Date of incorporation
- Registered office address
- SIC codes (Standard Industrial Classification — 75000 is "Veterinary activities")
- Company type (ltd, llp, plc)

**Officers:**
- Directors and secretaries with appointment dates
- Useful for cross-referencing head vet / practice owner

**Filing History → Accounts:**
- Annual accounts filed as PDF or iXBRL
- For small companies: total assets, net worth, employee count band
- For medium/large: turnover, profit/loss, detailed balance sheet
- Most vet practices file as "small" or "micro" — limited detail but employee count and total assets are usually present

**Persons of Significant Control (PSC):**
- Beneficial owners (25%+ shareholding)
- Useful for identifying group ownership (CVS, IVC Evidensia, etc.)

## API Details

Base URL: `https://api.company-information.service.gov.uk`

Key endpoints:
- `GET /search/companies?q={name}` — search by name
- `GET /company/{number}` — company profile
- `GET /company/{number}/officers` — directors/secretaries
- `GET /company/{number}/filing-history` — accounts and filings
- `GET /company/{number}/persons-with-significant-control` — beneficial owners

Rate limits: 600 requests per 5 minutes with API key, lower without.

Authentication: HTTP basic auth with API key as username, empty password. Free to register at https://developer.company-information.service.gov.uk/

## Matching Strategy

Vet practice names don't always match company registrations exactly. Matching approach:

1. **Search by name + postcode:** Query `/search/companies?q={practice_name}` and filter results by registered address postcode matching the practice postcode.

2. **SIC code filter:** Only consider companies with SIC code 75000 (Veterinary activities) or 86900 (Other human health activities — some mixed practices).

3. **Fuzzy matching:** Practice name "Oakwood Veterinary Surgery" might be registered as "Oakwood Vets Ltd" or "J Smith Veterinary Services Ltd". Use the LLM to pick the best match from search results given the practice name, address, and director names.

4. **Group detection:** If the PSC shows a corporate owner (e.g., "CVS Group PLC"), record the group relationship. This confirms/enriches the `group_name` field we already extract from websites.

## Data to Store

New table or extend existing:

```sql
CREATE TABLE IF NOT EXISTS business_intel.company_house_data (
    business_id UUID NOT NULL REFERENCES business_intel.businesses(id),
    company_number VARCHAR(10),
    company_name TEXT,
    company_status TEXT,          -- active, dissolved, etc.
    incorporation_date DATE,
    company_type TEXT,            -- ltd, llp, etc.
    sic_codes TEXT[],
    registered_address JSONB,
    
    -- Financial data (from latest accounts)
    accounts_date DATE,
    total_assets_gbp NUMERIC(12,2),
    net_worth_gbp NUMERIC(12,2),
    turnover_gbp NUMERIC(12,2),  -- only if full accounts filed
    profit_loss_gbp NUMERIC(12,2),
    employee_count_band TEXT,     -- "1-50", "51-250", etc.
    employee_count INTEGER,      -- exact if available
    accounts_type TEXT,           -- micro, small, medium, full
    
    -- Officers
    directors JSONB,             -- [{name, appointed_on, role, dob_month, dob_year}]
    
    -- Ownership
    psc JSONB,                   -- [{name, kind, notified_on, ownership_band}]
    parent_company TEXT,         -- corporate PSC name if applicable
    
    -- Owner age signals (for acquisition targeting)
    owner_dob_year INTEGER,      -- from Companies House (month/year disclosed)
    owner_dob_month INTEGER,
    owner_estimated_age INTEGER,  -- calculated from dob_year
    owner_appointment_year INTEGER, -- year first appointed as director
    owner_tenure_years INTEGER,   -- years since first appointment
    incorporation_years INTEGER,  -- years since company incorporated
    succession_risk TEXT,         -- high/medium/low (derived signal)
    
    -- Matching metadata
    match_confidence NUMERIC(3,2),
    match_method TEXT,           -- exact_name, fuzzy, llm_matched
    
    updated_at TIMESTAMP DEFAULT NOW(),
    
    PRIMARY KEY (business_id)
);
```

## Workflow Integration

Add two new steps to the vet-practice-verifier, after `load_business` and before `search_practice`:

```
load_business → search_companies_house → extract_financials → search_practice → scrape_website → ...
```

### Step: search_companies_house

New action: `companies_house_search`

- Takes `business_record.business.name` and `business_record.business.postcode`
- Calls Companies House search API
- Returns top 5 matches filtered by SIC code 75000

Config:
```json
{
    "action": "companies_house_search",
    "config": {
        "name_field": "business_record.business.name",
        "postcode_field": "business_record.business.postcode",
        "sic_filter": ["75000"],
        "max_results": 5
    },
    "next_step": "extract_financials",
    "output_field": "ch_search_results"
}
```

### Step: extract_financials

Uses LLM to match the right company and extract key data. Could also be done deterministically if the match is obvious (exact name + postcode match).

Config:
```json
{
    "action": "companies_house_enrich",
    "config": {
        "search_results_field": "ch_search_results",
        "business_record_field": "business_record",
        "fetch_officers": true,
        "fetch_psc": true,
        "fetch_latest_accounts": true
    },
    "next_step": "search_practice",
    "output_field": "company_data"
}
```

The `store_business_verification` action would then also persist the company data.

## Implementation Steps

1. **Register for Companies House API key** — free, instant
2. **Create `companies_house_search` action** — Go code, HTTP client, calls search endpoint
3. **Create `companies_house_enrich` action** — fetches profile, officers, PSC, latest filing for matched company
4. **Create `company_house_data` table** — migration SQL
5. **Update `store_business_verification`** — persist company data alongside other verification results
6. **Update verifier workflow** — add the two new steps
7. **Backfill** — run against already-verified businesses (separate batch, doesn't need scraping)

## Rate Limiting Considerations

At 600 req/5 min = 120/min, and ~3 API calls per business (search + profile + accounts):
- 40 businesses per minute
- 2,400 per hour
- Full backfill of 3,400 businesses: ~1.5 hours

This fits comfortably within the existing sequential processing model. No special rate limiting needed beyond what the batch processor already does.

## Job Board Salary Data (Future)

A separate enrichment step could scrape job boards for salary data:

- Indeed API or scrape: `{practice_name} veterinary` jobs
- VetClick, Vet Record Jobs
- Extract salary ranges, job titles, experience requirements

This is lower priority than Companies House (less structured, more scraping fragility) but would add salary benchmarking per practice and region. Could be a separate scheduled task rather than part of the verification workflow.

## Pet Population Correlation (Future)

Area-level pet density from PDSA/PFMA data could be loaded into the `search_areas` table as additional columns (estimated_pet_population, dog_count, cat_count). This gives a per-practice "market size" estimate based on catchment area.

This is a static data import, not an API integration — download the annual reports, extract the regional figures, load into the DB.

## Owner Age & Succession Signals (PE Targeting)

Private equity firms acquiring vet practices (CVS, IVC Evidensia, VetPartners) look for owner-operators approaching retirement. Several data sources give age signals:

### Companies House — Date of Birth (Direct)

The Officers endpoint returns **month and year of birth** for every director. This is public data, disclosed on every directorship filing. The API returns it as:

```json
{
    "name": "John Smith",
    "officer_role": "director",
    "appointed_on": "1998-03-15",
    "date_of_birth": {
        "month": 6,
        "year": 1958
    }
}
```

From this we derive:
- **Estimated age** (year-level precision)
- **Tenure as director** (appointed_on → now)
- **Years since incorporation** — long-incorporated single-owner practices suggest founder-operators

### Derived Succession Risk Score

Combine signals into a simple risk score:

| Signal | High risk | Medium risk | Low risk |
|--------|-----------|-------------|----------|
| Owner age | 60+ | 50-59 | Under 50 |
| Tenure | 20+ years | 10-19 years | Under 10 |
| Company age | 20+ years | 10-19 years | Under 10 |
| Corporate PSC | No (independent) | — | Yes (already acquired) |
| Multiple directors | Single director | — | Multiple (succession planned) |
| Recent officer changes | Resignations | — | New appointments |

A practice with a 63-year-old sole director who incorporated 30 years ago and has no corporate PSC scores high on all dimensions — strong acquisition candidate.

### Website Signals (Already Collected)

The LLM extraction already picks up clues from scraped content:
- **Head vet name** — cross-reference with Companies House director DOB
- **"Established in 19XX"** — many practice websites mention founding year
- **Staff page photos** — we don't analyse images but the text around them sometimes mentions qualifications year (e.g. "BVSc 1985") which implies graduation age ~23, giving approximate birth year
- **"Retiring" / "new ownership" mentions** — occasionally found in news sections

### RCVS Register (Future)

The Royal College of Veterinary Surgeons register lists every practising vet with their qualification year. Qualification year minus ~23 gives approximate birth year. The register is searchable at https://findavet.rcvs.org.uk/ — scraping it would give qualification dates for named vets already in our data.

### LinkedIn (Future, Handle Carefully)

LinkedIn profiles often show education dates which imply age. However, scraping LinkedIn violates their ToS and raises GDPR concerns around profiling. If pursued at all, this would need to be manual or via their official API with consent.

### Implementation Notes

- Companies House DOB data requires no extra API calls — it comes with the Officers endpoint we're already calling
- The `succession_risk` field is derived at storage time, not from an external source
- All data is public record (Companies House filings are a legal requirement)
- GDPR: processing public Companies House data for legitimate business purposes (market analysis) is lawful under Article 6(1)(f), but any dataset sold to third parties should note the data source and legal basis
- When selling data: present age-related fields as "company maturity indicators" and "ownership tenure" rather than personal age profiling — same data, better framing
