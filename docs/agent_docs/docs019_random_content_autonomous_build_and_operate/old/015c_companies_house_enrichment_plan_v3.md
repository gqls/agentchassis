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
load_ch_enrichment_batch (10 businesses per batch)
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

**Memory:** The pod runs with 2Gi limit / 512Mi request. Batch size was reduced from 20 to 10 after OOM kills caused by orchestration state growth during large batches.

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
| Vet industry title — "veterinary" | +0.15 | Whole word "veterinary" in company title |
| Vet industry title — "vet"/"vets" | +0.10 | Whole word "vet" or "vets" in company title |
| SIC prefix match | +0.25 | Any SIC code starts with a filter prefix |
| No SIC in result | -0.05 | Search results don't include SIC codes (normal for search API) |
| Wrong SIC codes | -0.15 | Has SIC codes but none match the filter |

**Minimum score threshold: 0.4**

The SIC filter uses prefix matching: filter value "75" matches SIC codes "75000", "75100", etc. Current filter prefixes in the workflow: `["75", "749", "869"]`.

The vet industry title bonus acts as a proxy for the missing SIC codes in search results. "Veterinary" or "vets" in the company title is a strong signal that the result is in the right industry. The bonus uses whole-word matching (splitting on whitespace, stripping trailing punctuation) to avoid false positives from words like "veteran" or "vetting".

### Scoring Examples

**Good match: "Erne Veterinary Group" → "ERNE VETERINARY GROUP LIMITED" (BT74)**
- Postcode BT74 matches: +0.35
- Name contains: +0.15
- Active: +0.15
- "veterinary" in title: +0.15
- No SIC in search result: -0.05
- **Total: 0.75 ✓**

**Good match without postcode: "Goddard Veterinary Group Eastcote" → "GODDARD VETERINARY GROUP LIMITED" (YO30 — HQ, not branch)**
- No postcode match: +0.00
- Word overlap (3/4 words): +0.15
- Active: +0.15
- "veterinary" in title: +0.15
- No SIC: -0.05
- **Total: 0.40 ✓** (vet title bonus pushes it over threshold)

**Wrong company: "Erne Veterinary Group" → "ERNES GROUP LONDON LTD" (NW11)**
- No postcode match: +0.00
- Word overlap (2/3 words): +0.13
- Active: +0.15
- No "veterinary"/"vet" in title: +0.00
- No SIC: -0.05
- **Total: 0.23 ✗**

**Wrong vet company: "Elms Veterinary Surgery" → "ALMA VETERINARY SURGERY LTD" (YO12)**
- No postcode match: +0.00
- Word overlap (1/3 "surgery"): +0.07
- Active: +0.15
- "veterinary" in title: +0.15
- No SIC: -0.05
- **Total: 0.32 ✗** (correctly rejected — name too different despite vet title)

### Known Limitations

1. **Search API returns `address`, not `registered_office_address`** — the search results use the JSON key `"address"` while the company profile API uses `"registered_office_address"`. The `chSearchItem` struct has both fields to handle this. The scoring function checks `item.Address.PostalCode` first, falling back to `item.RegisteredOfficeAddress.PostalCode`. This was the root cause of early zero-match results — the original struct only had `registered_office_address`, so postcodes were never parsed from search results and the +0.35 postcode bonus never applied.

2. **Search API doesn't return SIC codes** — the SIC scoring factor only works for companies we've previously fetched. First-time matches rely entirely on name + postcode + active status.

3. **Sole traders and partnerships** — many independent vet practices are not incorporated as limited companies and won't appear on Companies House at all. These correctly return `no_match`.

4. **Registered address vs trading address** — some practices trade from a different address than their registered office. The postcode match uses the registered address, which may be an accountant's office or home address. This reduces postcode hit rates for practices registered at non-trading addresses.

5. **Name variations** — "J Smith Veterinary Services Ltd" won't match well against "Smithfield Vets". The word-overlap scoring handles some of this but not all.

6. **False positives from postcode + partial name** — e.g. "Eglish Builders" matching "Eglish Veterinary Clinic" because the postcode area matches and "Eglish" overlaps. A post-fetch SIC validation step (checking the fetched profile's SIC codes start with "75") would catch these but is not yet implemented.

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

### Kustomize

Structure: `deployments/kustomize/services/business-intel/`
- `base/deployment.yaml` — bare deployment (no env vars)
- `overlays/production/uk_001/kustomization.yaml` — references base, configmap, patch, and image tag
- `overlays/production/uk_001/patch-deployment.yaml` — all env vars (database passwords, Kafka topics, CH API key)
- `overlays/production/uk_001/configmap.yaml` — `business-intel-config` with Kafka brokers and database connection details

The kustomize patch target must be `name: business-intel` (matching the base deployment name). See debug history item #8.

### Makefile

The `business-intel` deployment is included in `deploy-agents`, `redeploy-agents`, `update-kustomization-images`, and has standalone targets: `deploy-business-intel`, `logs-business-intel`, `restart-business-intel`.

## Backfill Progress

With ~2,725 unenriched businesses remaining (as of March 18, 2026) and batches of 10 at 1200s intervals:
- ~273 batches needed
- ~91 hours of wall time to complete full backfill
- Many will be `no_match` (sole traders, partnerships, corporate branches) — these complete quickly (no fetch step)

**Observed match rate:** ~85% from early batches (35 matched out of 41 processed). This is likely inflated by the initial batches covering more independent practices. Expect the overall rate to settle around 15-30% as the backfill covers more corporate branches and sole traders.

**Corporate branches** are a significant source of `no_match`. Practices owned by CVS, IVC Evidensia, Medivet, VetPartners etc. are typically registered under the parent company name at a central HQ address. The practice name and local postcode won't match the CH record. These are correctly identified as `no_match` — the business is already flagged as corporate-owned in the `businesses` table via the `group_name` field from web scraping.

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

## Financial Data from Filed Accounts (Planned)

The financial columns in `companies_house_data` (total_assets_gbp, net_worth_gbp, turnover_gbp, etc.) are currently unpopulated. The CH profile and officers endpoints don't include financial data — it's only available from the filed accounts documents.

### How It Works

CH accounts are filed as iXBRL (inline XBRL) — structured XHTML with tagged financial values. The pipeline to extract them:

1. **Get latest accounts filing:** `GET /company/{number}/filing-history?category=accounts&items_per_page=1`
    - Returns a `document_metadata` URL

2. **Get document metadata:** `GET {document_metadata_url}`
    - Returns available formats: `application/pdf` and `application/xhtml+xml`
    - Returns a content URL

3. **Download iXBRL:** `GET {content_url}` with `Accept: application/xhtml+xml`
    - **Must follow redirects** (`-L` in curl, `CheckRedirect` in Go HTTP client) — the content endpoint redirects to S3
    - Returns structured XHTML with `ix:nonFraction` elements containing financial values

### iXBRL Tags to Extract

Financial values are in `<ix:nonFraction>` elements with a `name` attribute and `contextRef` indicating the period:

```xml
<ix:nonFraction contextRef="CY_END" decimals="0" format="ixt2:numdotdecimal" 
    name="core:NetAssetsLiabilities" unitRef="GBP">960,294</ix:nonFraction>
```

Key tags and their DB mappings:

| iXBRL Tag | Context | DB Field | Available In |
|-----------|---------|----------|-------------|
| `core:NetAssetsLiabilities` | CY_END | `net_worth_gbp` | All filings |
| `core:TotalAssetsLessCurrentLiabilities` | CY_END | `total_assets_gbp` | All filings |
| `core:FixedAssets` | CY_END | — (store in raw) | All filings |
| `core:CurrentAssets` | CY_END | — (store in raw) | All filings |
| `core:AverageNumberEmployeesDuringPeriod` | CY | `employee_count` | Most filings |
| `core:TurnoverRevenue` | CY | `turnover_gbp` | Full accounts only |
| `core:ProfitLossForPeriod` | CY | `profit_loss_gbp` | Full accounts only |

Context suffixes: `CY_END` = current year-end balance sheet instant, `CY` = current year period, `PY_END` = prior year-end.

### What Small Companies Disclose

Most vet practices file as "small" or "micro" under abridged accounts. These typically include:
- Balance sheet (fixed assets, current assets, creditors, net assets, equity) ✓
- Employee count ✓
- Turnover / revenue ✗ (exempted under Section 444)
- Profit & loss ✗ (exempted)

The filing often contains `EntityHasTakenExemptionUnderCompaniesActInNotPublishingItsOwnProfitLossAccountTruefalse: true` confirming the P&L exemption.

Larger vet groups (medium/large companies, those filing full accounts) will include turnover and profit. These are the more valuable data points but cover fewer businesses.

### Accounts Type Field

The filing history response includes `description` which indicates the accounts type:
- `accounts-with-accounts-type-unaudited-abridged` — small company, limited data
- `accounts-with-accounts-type-total-exemption-full` — micro entity, very limited
- `accounts-with-accounts-type-full` — full accounts, includes P&L
- `accounts-with-accounts-type-medium` — medium company

Store this in `accounts_type` and use it to set expectations about which fields will be available.

### Implementation Approach

A new Go action `companies_house_fetch_accounts` would slot into the workflow after `fetch_details`:

```
search_ch → check_match → fetch_details → fetch_accounts → store_enrichment
```

The action would:
1. Call filing history API to get latest accounts document link
2. Call document metadata API to get content URL
3. Download iXBRL with redirect-following HTTP client
4. Parse `ix:nonFraction` elements using regex (the tags are consistent enough that full XML parsing isn't needed)
5. Return extracted values as a map for `store_enrichment` to persist

Regex pattern for extraction:
```
<ix:nonFraction[^>]*name="([^"]*)"[^>]*contextRef="(CY_END|CY)"[^>]*>([^<]+)</ix:nonFraction>
```

Rate limiting: 2 additional API calls per matched company (filing history + document download). With the existing 15s delay, this stays within limits.

### Tested Example: Eglish Builders (NI022532)

Filing history: `GET /company/NI022532/filing-history?category=accounts&items_per_page=1`
- Latest filing: unaudited abridged accounts for year ended 31 October 2024
- Document metadata URL: `https://document-api.company-information.service.gov.uk/document/1ZQnz8Hhz5d3vP--x104tx3onRnzxBNolcPz0S5w2Nc`
- Available as both PDF (47KB) and iXBRL (171KB)

Extracted values from iXBRL:
- Fixed assets: £413,384
- Current assets: £2,329,076
- Net assets (net worth): £960,294
- Total assets less current liabilities: £972,060
- Shareholders funds (equity): £960,294
- Average employees: 12
- Turnover: not disclosed (small company exemption)
- Profit/loss: not disclosed

## Future Enrichment Sources

The `business-intel` pod is designed to host multiple enrichment workflows beyond CH:

- **Job board salary data** — scrape Indeed, VetClick for salary benchmarking per practice and region
- **Pet population correlation** — PDSA/PFMA annual data loaded into search_areas for market size estimates
- **RCVS Register** — qualification dates for named vets imply approximate birth year
- **Credit data** — commercial credit scores and CCJ history

Each would be a separate scheduled task targeting the same `business-intel` pod with its own workflow in the agent definition.

## Debug History

Bugs found and fixed during initial deployment (March 2026):

1. **Stale loop variable** — Search and fetch actions used `ExtractNestedField` which resolved to the wrong business during loop iterations. Fixed by switching to `ExtractFields`.

2. **Over-aggressive name cleaning** — `cleanCompanySearchName` minimum length was 3 chars, causing "Erne Veterinary Group" to search as "Erne". Raised to 6 chars.

3. **Workflow condition** — `ch_search.matched != 0` was comparing boolean to integer. Changed to `ch_search.matched == true`.

4. **SIC filter exact match** — Original filter was `["75000"]` with exact string equality. Changed to prefix-based matching with `strings.HasPrefix` and broadened to `["75", "749", "869"]`.

5. **SIC penalty too harsh** — `-0.30` penalty for missing SIC codes. Since the search API never returns SIC codes, every result was penalised. Softened to `-0.05` for empty SIC (normal for search API) and `-0.15` for wrong SIC codes.

6. **Address field mismatch (root cause of zero matches)** — The CH search API returns address data under `"address"` but the Go struct only had `json:"registered_office_address"` (which is the profile API's key). Postcodes were never parsed from search results, so the +0.35 postcode bonus never applied. Fixed by adding an `Address` struct field with `json:"address"` tag, with scoring checking both fields.

7. **OOM kills** — Batch size of 20 caused orchestration state to grow beyond pod memory limits. Reduced to 10 and increased memory limit to 2Gi (later 3Gi).

8. **Kustomize patch target wrong** — `business-intel` overlay had `target.name: vet-intel` instead of `target.name: business-intel`. The patch with all env vars (database passwords, Kafka topics, CH API key) was never applied. Pod crashed with `CLIENTS_DB_PASSWORD is not set`.

9. **Vet industry title bonus** — Many vet practices registered as limited companies include "veterinary" or "vets" in their company title, but the CH search API doesn't return SIC codes. Without postcode match (e.g. branches registered at HQ), scores were too low. Added +0.15 for "veterinary" and +0.10 for "vet"/"vets" as whole words in the company title. Uses word-boundary matching to avoid false positives from "veteran" or "vetting".

## GDPR & Legal Notes

- All Companies House data is public record (filing is a legal requirement)
- Processing for legitimate business purposes (market analysis) is lawful under Article 6(1)(f)
- The `succession_risk` field is derived from public data, not from private profiling
- If data is sold to third parties: note the data source and legal basis; present age-related fields as "company maturity indicators" and "ownership tenure" rather than personal age profiling

