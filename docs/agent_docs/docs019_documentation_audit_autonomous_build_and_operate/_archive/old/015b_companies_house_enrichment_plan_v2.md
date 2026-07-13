# Companies House Enrichment — Implementation & Plan

## Status: Implemented — Tuning In Progress

The CH enrichment pipeline is live, running as a standalone agent on the `business-intel` pod. The core loop works: load batch → search CH → score matches → fetch details → store enrichment data. Current work is tuning match scoring to improve hit rates.

## Architecture

### Deployment

CH enrichment runs on the `business-intel` static pod (single replica), not as part of the vet-practice-verifier workflow. This separation keeps verification and enrichment independent — enrichment can run against already-verified businesses without re-verifying them.

- **Pod:** `business-intel` (image: `agent-chassis`, `AGENT_TYPE=business-intel`)
- **Agent definition:** `business-intel` (renamed from `ch-enricher`)
- **Scheduled task:** `ch-enrichment` (fires every 1200s, concurrency group: `ch-enrichment`, max_concurrent: 1)
- **Kafka topics:** `system.agent.business-intel.requests`, `system.agent.business-intel.responses`
- **Kustomize:** `deployments/kustomize/services/business-intel/`

### Workflow

```
load_ch_enrichment_batch (20 businesses per batch)
  → check_batch (any businesses to process?)
  → process_batch (loop, sequential, continue_on_error)
      → companies_house_search (search by name, score by postcode + name similarity)
      → check_match (ch_search.matched == true)
          → companies_house_fetch (profile, officers, PSC)
          → store_ch_enrichment (write to DB, derive succession signals)
        OR
          → store_ch_enrichment (record no_match)
  → notify_scheduler (update last_completed_at)
  → complete
```

Each business in the batch has a 15-second delay before API calls to stay well within CH rate limits (~2 businesses/minute).

## What Companies House Gives Us

The Companies House API (free, API key for rate limits) provides:

**Company Profile** (from `/company/{number}`):
- Registered company name, number, status (active/dissolved)
- Date of incorporation
- Registered office address
- SIC codes (Standard Industrial Classification — 75000 is "Veterinary activities")
- Company type (ltd, llp, plc)

**Officers** (from `/company/{number}/officers`):
- Directors and secretaries with appointment dates
- Date of birth (month + year) — used for succession risk scoring
- Useful for cross-referencing head vet / practice owner

**Filing History → Accounts:**
- Annual accounts filed as PDF or iXBRL
- For small companies: total assets, net worth, employee count band
- For medium/large: turnover, profit/loss, detailed balance sheet
- Most vet practices file as "small" or "micro" — limited detail but employee count and total assets are usually present

**Persons of Significant Control (PSC)** (from `/company/{number}/persons-with-significant-control`):
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

**Important:** The search endpoint does NOT return SIC codes. SIC codes are only available from the company profile endpoint (the fetch step). This affects scoring — see below.

## Matching Strategy (Implemented)

Matching is fully deterministic — no LLM involved. The `scoreCHMatches` function scores each search result against the business name and postcode.

### Name Cleaning

Before searching, `cleanCompanySearchName` strips common suffixes that hurt search relevance:
- " - Veterinary Practice", " Veterinary Surgery", " Veterinary Centre", " Veterinary Clinic"
- " Veterinary Hospital", " Veterinary Group", " Vets", " Ltd", " Limited"

If stripping reduces the name below 6 characters, the original name is used instead. This prevents single-word searches like "Erne" (from "Erne Veterinary Group") that return too many irrelevant results.

### Scoring (current implementation)

Each search result is scored on four factors:

| Factor | Score | Condition |
|--------|-------|-----------|
| Postcode prefix match | +0.35 | Outward code matches (e.g. "BT74" == "BT74") |
| Name — exact match | +0.25 | Case-insensitive exact match |
| Name — contains | +0.15 | One name contains the other |
| Name — word overlap | +0.0 to +0.20 | Proportion of significant words (>3 chars) that appear in both |
| Company active | +0.15 | company_status == "active" |
| SIC prefix match | +0.25 | Any SIC code starts with a filter prefix |
| No SIC in result | -0.05 | Search results don't include SIC codes (normal for search API) |
| Wrong SIC codes | -0.15 | Has SIC codes but none match the filter |

**Minimum score threshold: 0.4**

The SIC filter uses prefix matching: filter value "75" matches SIC codes "75000", "75100", etc. Current filter prefixes in the workflow: `["75", "749", "869"]`.

### Scoring Examples

**Good match: "Erne Veterinary Group" → "ERNE VETERINARY GROUP LIMITED" (BT74)**
- Postcode BT74 matches: +0.35
- Name contains: +0.15
- Active: +0.15
- No SIC in search result: -0.05
- **Total: 0.60 ✓**

**Wrong company: "Erne Veterinary Group" → "ERNES GROUP LONDON LTD" (NW11)**
- No postcode match: +0.00
- Word overlap (2/3 words): +0.13
- Active: +0.15
- No SIC: -0.05
- **Total: 0.23 ✗**

### Known Limitations

1. **Search API doesn't return SIC codes** — the SIC scoring factor only works for companies we've previously fetched. First-time matches rely entirely on name + postcode + active status.

2. **Sole traders and partnerships** — many independent vet practices are not incorporated as limited companies and won't appear on Companies House at all. These correctly return `no_match`.

3. **Registered address vs trading address** — some practices trade from a different address than their registered office. The postcode match uses the registered address, which may be an accountant's office or home address. This reduces postcode hit rates for practices registered at non-trading addresses.

4. **Name variations** — "J Smith Veterinary Services Ltd" won't match well against "Smithfield Vets". The word-overlap scoring handles some of this but not all.

## Data Storage

Table: `business_intel.companies_house_data`

```sql
CREATE TABLE IF NOT EXISTS business_intel.companies_house_data (
    business_id            UUID NOT NULL REFERENCES business_intel.businesses(id),
    company_number         VARCHAR(10),
    company_name           TEXT,
    company_status         TEXT,
    company_type           TEXT,
    incorporation_date     DATE,
    cessation_date         DATE,
    sic_codes              TEXT[],
    registered_address     JSONB,

    -- Financial data (from latest accounts)
    accounts_date          DATE,
    accounts_type          TEXT,
    total_assets_gbp       NUMERIC(12,2),
    net_worth_gbp          NUMERIC(12,2),
    turnover_gbp           NUMERIC(12,2),
    profit_loss_gbp        NUMERIC(12,2),
    employee_count         INTEGER,
    employee_count_band    TEXT,

    -- Officers and ownership
    officers               JSONB,
    psc                    JSONB,
    parent_company_name    TEXT,
    parent_company_number  VARCHAR(10),

    -- Owner age & succession signals
    owner_name             TEXT,
    owner_dob_year         INTEGER,
    owner_dob_month        INTEGER,
    owner_estimated_age    INTEGER,
    owner_appointment_date DATE,
    owner_tenure_years     INTEGER,
    is_sole_director       BOOLEAN,
    is_corporate_owned     BOOLEAN,
    succession_risk        TEXT,          -- high/medium/low/unknown

    -- Matching metadata
    match_confidence       NUMERIC(3,2),
    match_method           TEXT,          -- name_postcode, name_search, no_match
    search_query           TEXT,
    search_results_count   INTEGER,

    -- Timestamps
    enriched_at            TIMESTAMPTZ DEFAULT NOW(),
    enrichment_source      TEXT DEFAULT 'companies-house-api',
    raw_response           JSONB,
    created_at             TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at             TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    PRIMARY KEY (business_id)
);

CREATE INDEX idx_ch_succession_risk ON business_intel.companies_house_data (succession_risk)
    WHERE succession_risk IS NOT NULL;
```

## Go Actions

File: `platform/orchestration/actions/companies_house_actions.go`

Four actions registered in the action registry:

1. **`load_ch_enrichment_batch`** — Queries verified businesses without CH data, returns batch of N businesses
2. **`companies_house_search`** — Searches CH by business name, scores results by postcode + name similarity + SIC
3. **`companies_house_fetch`** — Fetches company profile, officers, PSC for a matched company number
4. **`store_ch_enrichment`** — Writes enrichment data to DB, derives succession risk signals

### Loop Variable Extraction

The search and fetch actions use `datahelpers.ExtractFields` (not `ExtractNestedField`) to read loop variables. `ExtractNestedField` resolves to stale values during loops because it searches deeply nested collected_data paths and can find a previous iteration's data. The store action already used `ExtractFields` — the search and fetch actions were updated to match.

### Rate Limiting

Each search and fetch call has a configurable delay (default 15s via `delay_ms` in workflow config). At ~2 businesses per minute with 3 API calls each, this stays well within CH's 600 requests per 5 minutes limit.

## Operational Notes

### Kafka Topics

- `system.agent.business-intel.requests` — retention 7 days
- `system.agent.business-intel.responses` — **retention 1 hour** (short retention prevents stale response replay on pod restart; see "Stale Response Problem" below)

### Stale Response Problem (Resolved)

When the business-intel pod restarts, it gets a new pod UID which becomes a new Kafka consumer group for the responses topic. With `auto.offset.reset=earliest` (the default), the new consumer group reads from offset 0 and replays every historical response. Completed orchestration responses cause a `RESPONSE_ORPHANED` claim loop that blocks the pod from processing new work.

**Resolution:** Set short retention (1 hour) on the responses topic. Historical responses older than 1 hour are automatically purged. Any legitimate in-flight response completes well within 1 hour. The reaper catches anything stuck beyond 90 minutes.

**If the problem recurs:** Delete and recreate the responses topic to clear stale messages, then restart the pod.

### Scheduled Task

```sql
-- ch-enrichment scheduled task
name: ch-enrichment
interval_seconds: 1200
target_agent_type: business-intel
target_topic: system.agent.business-intel.requests
concurrency_group: ch-enrichment
max_concurrent: 1
pre_query: SELECT COUNT(*)::text as unenriched
           FROM business_intel.businesses b
           JOIN business_intel.business_verticals bv ON bv.id = b.vertical_id
           LEFT JOIN business_intel.companies_house_data ch ON ch.business_id = b.id
           WHERE bv.slug = 'veterinary'
             AND b.verification_status = 'verified'
             AND ch.business_id IS NULL
           HAVING COUNT(*) > 0
```

The pre_query only fires the task when there are unenriched businesses. Once the backfill completes, the task will skip silently until new businesses are verified.

### Terraform

Kafka topics are defined in `deployments/terraform/modules/kafka_topics/main.tf`. The business-intel topics were added alongside vet-intel topics.

### Makefile

The `business-intel` deployment is included in `deploy-agents`, `redeploy-agents`, `update-kustomization-images`, and has standalone targets: `deploy-business-intel`, `logs-business-intel`, `restart-business-intel`.

## Backfill Progress

With ~2,339 unenriched businesses (as of March 2026) and batches of 20 at 1200s intervals:
- ~117 batches needed
- ~39 hours of wall time to complete full backfill
- Many will be `no_match` (sole traders, partnerships) — these complete quickly (no fetch step)

## Owner Age & Succession Signals (PE Targeting)

Private equity firms acquiring vet practices (CVS, IVC Evidensia, VetPartners) look for owner-operators approaching retirement. The CH Officers endpoint provides month and year of birth for every director:

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

### Derived Succession Risk Score

Stored in `succession_risk` field, derived at storage time by `store_ch_enrichment`:

| Signal | High risk | Medium risk | Low risk |
|--------|-----------|-------------|----------|
| Owner age | 60+ | 50-59 | Under 50 |
| Tenure | 20+ years | 10-19 years | Under 10 |
| Company age | 20+ years | 10-19 years | Under 10 |
| Corporate PSC | No (independent) | — | Yes (already acquired) |
| Multiple directors | Single director | — | Multiple (succession planned) |
| Recent officer changes | Resignations | — | New appointments |

## Future Enrichment Sources

The `business-intel` pod is designed to host multiple enrichment workflows beyond CH:

- **Job board salary data** — scrape Indeed, VetClick for salary benchmarking per practice and region
- **Pet population correlation** — PDSA/PFMA annual data loaded into search_areas for market size estimates
- **RCVS Register** — qualification dates for named vets imply approximate birth year
- **Credit data** — commercial credit scores and CCJ history

Each would be a separate scheduled task targeting the same `business-intel` pod with its own workflow in the agent definition.

## GDPR & Legal Notes

- All Companies House data is public record (filing is a legal requirement)
- Processing for legitimate business purposes (market analysis) is lawful under Article 6(1)(f)
- The `succession_risk` field is derived from public data, not from private profiling
- If data is sold to third parties: note the data source and legal basis; present age-related fields as "company maturity indicators" and "ownership tenure" rather than personal age profiling
