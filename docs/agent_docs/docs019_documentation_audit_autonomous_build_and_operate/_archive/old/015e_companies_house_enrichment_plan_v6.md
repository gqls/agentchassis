# Companies House Enrichment — Implementation & Plan

## Status: Bulk Collection + Local Matching — Implemented

The CH enrichment pipeline has transitioned from per-business API search to a three-stage local matching approach:

1. **Bulk collect** — All 5,780 active SIC 75000 companies collected into `ch_vet_companies` table (completed March 19, 2026)
2. **Local matching** — Two-pass matching (postcode+name, then trigram name-only) runs in ~10 seconds with no API calls. 429 postcode matches + auto-accepted name matches, with ambiguous candidates queued for LLM review.
3. **LLM review** — claude-haiku-4-5 reviews ambiguous matches (similarity 0.50–0.90) in batches of 15. Classifies as confirmed/rejected/uncertain.

The original per-business search approach (`ch-enrichment` scheduled task) is disabled. The existing `companies_house_data` table continues to store enrichment data — only the matching method has changed.

## Architecture

### Deployment

All CH workflows run on the `business-intel` static pod (single replica). Different agent definitions share the same pod via Kafka message routing — the scheduler sends `config.agent_type` in the message, and `selectWorkflow` → `FindBestGroup` loads the correct workflow from `agent_definitions`.

- **Pod:** `business-intel` (image: `agent-chassis`, `AGENT_TYPE=business-intel`)
- **Kafka topics:** `system.agent.business-intel.requests` (2 partitions), `system.agent.business-intel.responses`
- **Kustomize:** `deployments/kustomize/services/business-intel/`

### Agent Definitions

| Agent Type | Purpose | Scheduled Task | Interval |
|---|---|---|---|
| `business-intel` | Per-business CH search enrichment (legacy) | `ch-enrichment` (disabled) | 1200s |
| `ch-collector` | Bulk collect all SIC 75000 companies | `ch-vet-collect` (disabled, run manually) | 30 days |
| `ch-matcher` | Local matching against `ch_vet_companies` | `ch-local-match` | daily |
| `ch-llm-reviewer` | LLM review of ambiguous matches | `ch-llm-review` | daily |

All agent definitions use `ON CONFLICT (type, version)` for upserts. All workflows are minimal — `action → complete` — with stats collection, scheduler notification, and business logic handled inside the Go actions.

### Go Actions

File locations in `platform/orchestration/actions/`:

| Action | File | Description |
|---|---|---|
| `ch_bulk_collect` | `companies_house_bulk_collect_action.go` | Paginate CH advanced search, store all results locally |
| `ch_local_match` | `companies_house_local_match_action.go` | Two-pass matching: postcode+name then trigram name-only |
| `ch_llm_review` | `companies_house_llm_review_action.go` | LLM review of pending_llm_review candidates |
| `load_ch_enrichment_batch` | `companies_house_actions.go` | Load verified businesses for per-business search (legacy) |
| `companies_house_search` | `companies_house_actions.go` | Search CH API by name, score by postcode+name (legacy) |
| `companies_house_fetch` | `companies_house_actions.go` | Fetch company profile, officers, PSC |
| `store_ch_enrichment` | `companies_house_actions.go` | Write enrichment data to DB, derive succession signals |

## What Companies House Gives Us

The Companies House API (free, API key for rate limits) provides:

**Company Profile** (from `/company/{number}`):
- Registered company name, number, status (active/dissolved)
- Date of incorporation, SIC codes (75000 = "Veterinary activities")
- Registered office address, company type (ltd, llp, plc)

**Officers** (from `/company/{number}/officers`):
- Directors and secretaries with appointment dates
- Date of birth (month + year) — used for succession risk scoring

**Persons of Significant Control** (from `/company/{number}/persons-with-significant-control`):
- Beneficial owners (25%+ shareholding)
- Useful for identifying group ownership (CVS, IVC Evidensia, etc.)

**Filing History → Accounts** (planned, not yet implemented):
- Annual accounts filed as iXBRL with tagged financial values
- For small companies: total assets, net worth, employee count band
- For medium/large: turnover, profit/loss, detailed balance sheet

**API Details:**
- Base URL: `https://api.company-information.service.gov.uk`
- Advanced search: `GET /advanced-search/companies?sic_codes=75000&company_status=active&size=100`
- Rate limits: 600 requests per 5 minutes with API key
- Authentication: HTTP basic auth with API key as username, empty password

## Stage 1: Bulk Collection (Implemented)

### How It Works

The `ch_bulk_collect` action paginates through the CH advanced search API for a given SIC code:

```
GET /advanced-search/companies?sic_codes=75000&company_status=active&size=100&start_index=0
```

All pagination is handled inside the single action call — no workflow loop. The action:
1. Ensures the `ch_vet_companies` table exists (idempotent)
2. Fetches 100 companies per page with 2s delay between pages
3. Upserts each company (safe for re-runs)
4. Populates `company_name_cleaned` for trigram matching
5. Queries stats and notifies the scheduler on completion

**Result:** 5,780 active SIC 75000 companies collected in 58 API calls (~2.5 minutes).

### Table: `business_intel.ch_vet_companies`

```sql
company_number      VARCHAR(10) PRIMARY KEY
company_name        TEXT NOT NULL
company_name_cleaned TEXT          -- lowercase, stripped of Ltd/Group/Surgery etc. for trigram matching
company_status      TEXT
company_type        TEXT
date_of_creation    DATE
date_of_cessation   DATE
sic_codes           TEXT[]
registered_address  JSONB
postcode            TEXT
postcode_prefix     TEXT           -- outward code (e.g. "BT74") for matching
locality            TEXT

-- Matching state
matched_business_id UUID           -- FK to businesses table, NULL if unmatched
matched_at          TIMESTAMPTZ
match_confidence    NUMERIC(3,2)
match_method        TEXT           -- local_postcode, local_postcode_exact, name_trigram, pending_llm_review, llm_confirmed, llm_rejected, llm_uncertain

-- Detail fetch state
details_fetched     BOOLEAN DEFAULT FALSE
details_fetched_at  TIMESTAMPTZ

-- Discovery state
is_discovered       BOOLEAN DEFAULT FALSE
discovery_status    TEXT DEFAULT 'pending'

-- Timestamps
collected_at        TIMESTAMPTZ DEFAULT NOW()
updated_at          TIMESTAMPTZ DEFAULT NOW()
```

### Indexes

```sql
-- Postcode prefix for pass 1 matching
idx_ch_vet_postcode_prefix ON ch_vet_companies (postcode_prefix) WHERE company_status = 'active'

-- GiST trigram for pass 2 name-only matching (~4ms per lookup)
idx_ch_vet_name_trgm_gist ON ch_vet_companies USING gist (lower(company_name) gist_trgm_ops)
idx_ch_vet_name_cleaned_gist ON ch_vet_companies USING gist (company_name_cleaned gist_trgm_ops)

-- Unmatched companies for discovery
idx_ch_vet_unmatched ON ch_vet_companies (matched_business_id) WHERE matched_business_id IS NULL AND company_status = 'active'
```

### Notable Data Points

- **SK9 3RN** has 554 companies — this is Pets at Home HQ in Cheadle. All Vets4Pets branches are registered there.
- **YO30 4UZ** has 68 companies — IVC Evidensia/Linnaeus HQ in York. Many IVC/Linnaeus branches registered there.
- **TA1 3DU** — another company formation address with many vet companies.

## Stage 2: Local Matching (Implemented)

### Two-Pass Architecture

The `ch_local_match` action processes all unmatched businesses in a single action call (~10 seconds total):

**Pass 1 — Postcode + name scoring**

For each business, find all CH companies with the same postcode prefix, then score by name similarity:

| Factor | Score | Notes |
|---|---|---|
| Postcode prefix match | +0.20 | Always present (filtered by query) |
| Exact full postcode | +0.15 | Bonus when inward code also matches |
| Name exact (cleaned) | +0.30 | After stripping Ltd/Group/Surgery etc. |
| Name contains | +0.20 | One cleaned name contains the other |
| Name word overlap | 0 to +0.25 | Proportion of words >3 chars in both |
| Vet word in CH name | +0.10 | Confirmatory (all are SIC 75000) |
| **Threshold** | **0.40** | |

**Pass 2 — Trigram name-only (no postcode requirement)**

For businesses not matched in pass 1, search the entire `ch_vet_companies` table using PostgreSQL trigram similarity on `company_name_cleaned`:

```sql
SELECT company_number, company_name, postcode,
       similarity(company_name_cleaned, $1) as sim
FROM business_intel.ch_vet_companies
WHERE company_status = 'active'
  AND matched_business_id IS NULL
  AND company_name_cleaned % $1
ORDER BY company_name_cleaned <-> $1
LIMIT 3
```

This uses the GiST trigram index for ~4ms per lookup. Catches companies registered at accountant/HQ addresses.

Three tiers:
- Similarity ≥ 0.90 + distinctive word overlap → **auto-accept** (method: `name_trigram`)
- Similarity 0.50–0.90 + distinctive word overlap → **store as `pending_llm_review`**
- Similarity < 0.50 or no distinctive word → **reject**

### Distinctive Word Overlap Check

Prevents false positives where common vet words inflate trigram scores. Requires at least one non-generic word from the business name to appear in the CH name.

Generic words (excluded): veterinary, vet, vets, animal, pet, pets, paws, mobile, services, clinic, centre, surgery, practice, hospital, group, limited, ltd, equine, farm, emergency, referrals.

Examples:
- "WW Mobile Veterinary Services" → "HAYLOFT MOBILE VETERINARY SERVICES" — "ww" (too short), rest generic → **rejected** ✓
- "Shropshire Farm Vets" → "SHROPSHIRE FARM VETS LIMITED" — "shropshire" is distinctive → **accepted** ✓

### Match Results (March 19, 2026)

| Method | Count | Avg Confidence |
|---|---|---|
| `local_postcode_exact` | 362 | 0.62 |
| `name_trigram` (auto-accepted) | ~50-80 | 0.88 |
| `local_postcode` | 67 | 0.49 |
| `pending_llm_review` | ~100-200 | varies |

Total: ~530 confirmed matches + pending LLM review candidates from 2,730 businesses processed.

### Name Cleaning

`cleanCompanySearchName` strips suffixes for search generality, keeping industry terms:

**Stripped:** Ltd, Limited, LLP, PLC, Group, Surgery, Centre, Center, Clinic, Hospital, Practice
**Kept:** Veterinary, Vets (industry signal)
**Stripped:** Parenthetical locations — "(Maldon)", "(Bickley)"
**Stripped:** Dash-separated generic descriptors — "- Veterinary Practice"
**Min length:** 6 chars (falls back to original if too short)

Principle: **search generality → scoring specificity**. "Castle Veterinary Group" → search "Castle Veterinary" to find both "CASTLE VETERINARY SURGERY LTD" and "CASTLE VETERINARY CENTRE LTD".

## Stage 3: LLM Review (Implemented, Not Yet Run)

The `ch_llm_review` action reviews `pending_llm_review` entries using claude-haiku-4-5.

### How It Works

1. Load all rows where `match_method = 'pending_llm_review'`
2. Batch 15 pairs per LLM call
3. For each pair, provide: practice name, postcode, CH company name, registered postcode, similarity score
4. LLM responds YES/NO/UNCERTAIN for each
5. Update match_method: `llm_confirmed`, clear match (`llm_rejected`), or `llm_uncertain`

### Prompt Structure

```
You are matching UK veterinary practice trading names to their Companies House
registered company names.

For each pair below, determine if they are the SAME business entity. Consider:
- The practice may trade under a different name than its registered company
- The registered office may be at an accountant's address, not the practice
- Corporate groups register branches under the parent or location-specific company
- Geographic proximity matters

1.
  Practice: "Four Paws Veterinary Clinic" (postcode: G5 8RS)
  CH Company: "FOUR WET PAWS LTD" (registered: GL52 8PG)
  Similarity score: 0.71

Respond: 1: YES / NO / UNCERTAIN
```

### Cost

~200 candidates × ~100 tokens per judgment = ~20,000 tokens. On Haiku: ~$0.05.

## Stage 4: Detail Fetch (Planned)

For confirmed matches (all methods except `pending_llm_review` and `llm_rejected`), fetch officers/PSC/accounts from the CH API and store in `companies_house_data`.

The existing `companies_house_fetch` and `store_ch_enrichment` actions handle this. A new action would:
1. Load matched rows from `ch_vet_companies` where `details_fetched = false`
2. Call CH API for each with rate limiting (15s delay)
3. Store enrichment data in `companies_house_data`
4. Mark `details_fetched = true` in `ch_vet_companies`

This replaces the current search-per-business enrichment pipeline entirely.

## Stage 5: Discovery (Planned)

Unmatched `ch_vet_companies` rows (5,351 as of March 19, 2026) are vet companies we don't have in our businesses table. Some are equine, agricultural, holding companies, or pet shops with SIC 75000. But many are small independent practices not found via web scraping.

A discovery workflow could:
- Look up the company's registered address
- Search for a website associated with that name or address
- Feed new practices into the vet-intel verification pipeline

## Future: Vertical Generalisation

The pipeline is currently vet-specific. Before adding the next vertical (e.g. seaweed farming, insect farming), the hardcoded values should be extracted into a vertical profile registry:

**Currently vet-specific:**
- SIC code "75000" in collector
- Table name `ch_vet_companies`
- Name cleaning suffixes (Surgery, Clinic, Practice — vet/health terms)
- Vet title bonus scoring
- Generic words list in distinctive word check
- Vertical slug "veterinary" in business loader query

**Planned approach:**
- `VerticalProfile` struct in Go mapping slug → SIC codes, industry keywords, generic words, strip suffixes, score bonuses
- Rename table to `ch_sic_companies` with a `collection_slug` column
- Workflow config carries `vertical_slug` — actions look up the profile
- One set of actions serves all verticals; new vertical = new profile entry + new agent definition

## Data Storage

### `business_intel.companies_house_data` (enrichment output — unchanged)

Stores the detailed enrichment data for matched companies. Populated by `store_ch_enrichment` after `companies_house_fetch` retrieves officers/PSC.

Key fields: company_number, company_name, officers (JSONB), psc (JSONB), owner_name, owner_estimated_age, succession_risk, match_confidence, match_method.

### `business_intel.ch_vet_companies` (local mirror — new)

Local copy of all CH companies with SIC 75000. Populated by `ch_bulk_collect`, matched by `ch_local_match`, reviewed by `ch_llm_review`. See schema above.

## Owner Age & Succession Signals

Private equity firms acquiring vet practices look for owner-operators approaching retirement. The CH Officers endpoint provides month and year of birth for every director.

### Derived Succession Risk Score

Stored in `succession_risk` field, derived at storage time by `store_ch_enrichment`:

| Signal | High risk | Medium risk | Low risk |
|--------|-----------|-------------|----------|
| Owner age | 60+ | 50-59 | Under 50 |
| Tenure | 20+ years | 10-19 years | Under 10 |
| Corporate PSC | No (independent) | — | Yes (already acquired) |
| Single director | Yes | — | Multiple (succession planned) |

## Operational Notes

### Kafka Topics

- `system.agent.business-intel.requests` — 2 partitions, retention 7 days
- `system.agent.business-intel.responses` — retention 1 hour (prevents stale response replay)

### Kustomize

Structure: `deployments/kustomize/services/business-intel/`
- `base/deployment.yaml` — bare deployment (no env vars)
- `overlays/production/uk_001/kustomization.yaml` — references base, configmap, patch, and image tag
- `overlays/production/uk_001/patch-deployment.yaml` — all env vars (database passwords, Kafka topics, CH API key)
- `overlays/production/uk_001/configmap.yaml` — `business-intel-config` with Kafka brokers and database connection details

The kustomize patch target must be `name: business-intel` (matching the base deployment name). See debug history item #8.

### Scheduled Tasks

```
ch-vet-collect:  disabled (run manually, monthly refresh)
ch-local-match:  daily, concurrency group: ch-matching
ch-llm-review:   daily, concurrency group: ch-matching (runs after local-match)
ch-enrichment:   disabled (legacy per-business search)
```

### Single-Pod Contention

The business-intel pod processes one orchestration at a time. Collection/matching/review messages queue behind enrichment batches if the legacy pipeline is running. When triggering collection or matching:
1. Disable `ch-enrichment`
2. Cancel active orchestrations
3. Purge the requests topic (set retention to 1s, wait, restore)
4. Restart pod, then trigger

### Message Routing for Different Agent Types

The scheduler sends `{"config":{"agent_type":"ch-collector"}}` to the business-intel requests topic. The pod's `selectWorkflow` function:
1. Loads business-intel agent definition (Priority 3 fallback)
2. Extracts `config.agent_type` = "ch-collector" from the message
3. Calls `FindBestGroup("ch-collector")` which looks up ch-collector in `agent_definitions`
4. Uses the ch-collector workflow (Priority 2)

This is how multiple workflows share a single pod without code changes.

## Financial Data from Filed Accounts (Planned)

The financial columns in `companies_house_data` are currently unpopulated. A new `companies_house_fetch_accounts` action would:
1. Get latest accounts filing from `/company/{number}/filing-history?category=accounts`
2. Download iXBRL document (must follow redirects)
3. Parse `ix:nonFraction` elements for: net assets, total assets, employee count, turnover (if disclosed)
4. Store in `companies_house_data` financial fields

Most vet practices file as small/micro — net assets and employee count available, turnover/profit usually exempt.

## Debug History

Bugs found and fixed during deployment (March 2026):

1. **Stale loop variable** — Search/fetch actions used `ExtractNestedField` → switched to `ExtractFields`.
2. **Name cleaning min length** — Raised from 3 to 6 chars.
3. **Workflow condition** — Changed `!= 0` to `== true` for boolean comparison.
4. **SIC filter** — Changed from exact `["75000"]` to prefix-based `["75", "749", "869"]` with `strings.HasPrefix`.
5. **SIC penalty** — Softened from `-0.30` to `-0.05` (empty) / `-0.15` (wrong).
6. **Address field mismatch** — CH search API returns `"address"`, not `"registered_office_address"`. Added dual struct field.
7. **OOM kills** — Batch size reduced 20→10, memory increased to 3Gi.
8. **Kustomize patch target** — `name: vet-intel` instead of `name: business-intel`. Env vars never applied.
9. **Vet industry title bonus** — Added +0.15 for "veterinary", +0.10 for "vet"/"vets" in company title.
10. **Name cleaner stripping too much** — Kept "Veterinary"/"Vets", strip only generic suffixes. Search generality → scoring specificity.
11. **SK9 formation agent** — 554 Vets4Pets companies at SK9 3RN (Pets at Home HQ). Corporate branches don't match by local postcode — expected behaviour.
12. **Trigram index type** — GIN index supports `LIKE`/`ILIKE` but not `<->` distance operator. Switched to GiST for `ORDER BY ... <->` queries (~4ms per lookup).
13. **Cleaned name column** — Raw company names include "LIMITED"/"LTD" which reduces trigram similarity. Added `company_name_cleaned` column with stripped suffixes. "SHROPSHIRE FARM VETS LIMITED" → similarity 0.72 raw, 1.0 cleaned.

## GDPR & Legal Notes

- All Companies House data is public record (filing is a legal requirement)
- Processing for legitimate business purposes (market analysis) is lawful under Article 6(1)(f)
- The `succession_risk` field is derived from public data, not from private profiling
- If data is sold to third parties: note the data source and legal basis; present age-related fields as "company maturity indicators" rather than personal age profiling
