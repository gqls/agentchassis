# Companies House Enrichment — Implementation & Plan

## Status: Full Pipeline — Collect → Match → LLM Review → Detail Fetch

The CH enrichment pipeline runs as five stages on the `business-intel` pod:

1. **Bulk collect** — 5,780 active SIC 75000 companies collected into `ch_vet_companies` (completed March 19, 2026)
2. **Local matching** — Two-pass matching (postcode+name, then trigram name-only) in ~10 seconds, no API calls
3. **LLM review** — claude-haiku-4-5 reviews ambiguous matches with group/location context, ~$0.05 per run
4. **Detail fetch** — Fetches officers/PSC/profile from CH API for confirmed matches, derives succession risk (deploying)
5. **Discovery** — Unmatched CH companies as new leads (planned)

### Match Results (March 19, 2026)

| Method | Count | Avg Confidence |
|---|---|---|
| `local_postcode_exact` | 362 | 0.62 |
| `llm_confirmed` | 141 | 0.67 |
| `local_postcode` | 67 | 0.49 |
| `name_trigram` (auto-accepted) | 64 | 1.00 |
| `pending_llm_review` | 68 | 0.51 |
| `llm_uncertain` | 15 | 0.62 |

**Total confirmed: 634** (362 + 141 + 67 + 64) from 2,730 businesses = **23.2%**. Plus 83 pending/uncertain for manual review. ~5,100 unmatched CH companies are discovery candidates.

The original per-business search approach (`ch-enrichment` scheduled task) is disabled.

## Architecture

### Deployment

All CH workflows run on the `business-intel` static pod (single replica). Different agent definitions share the same pod via Kafka message routing — the scheduler sends `config.agent_type` in the message, and `selectWorkflow` → `FindBestGroup` loads the correct workflow from `agent_definitions`.

- **Pod:** `business-intel` (image: `agent-chassis`, `AGENT_TYPE=business-intel`)
- **Kafka topics:** `system.agent.business-intel.requests` (2 partitions), `system.agent.business-intel.responses`
- **Kustomize:** `deployments/kustomize/services/business-intel/`

### Agent Definitions

| Agent Type | Purpose | Scheduled Task | Interval | Concurrency Group |
|---|---|---|---|---|
| `business-intel` | Per-business CH search (legacy) | `ch-enrichment` (disabled) | 1200s | ch-enrichment |
| `ch-collector` | Bulk collect SIC companies | `ch-vet-collect` (disabled) | monthly | ch-matching |
| `ch-matcher` | Local matching | `ch-local-match` | daily | ch-matching |
| `ch-llm-reviewer` | LLM review of ambiguous matches | `ch-llm-review` | daily | ch-matching |
| `ch-detail-fetcher` | Fetch officers/PSC for matches | `ch-detail-fetch` | 1200s | ch-enrichment |

All agent definitions use `ON CONFLICT (type, version)` for upserts. All workflows are minimal — `action → complete` — with stats, scheduler notification, and business logic handled inside Go actions.

### Go Actions

File locations in `platform/orchestration/actions/`:

| Action | File | Description |
|---|---|---|
| `ch_bulk_collect` | `companies_house_bulk_collect_action.go` | Paginate CH advanced search, store all results locally |
| `ch_local_match` | `companies_house_local_match_action.go` | Two-pass matching: postcode+name then trigram name-only |
| `ch_llm_review` | `companies_house_llm_review_action.go` | LLM review of pending candidates (generic, industry context from config) |
| `ch_detail_fetch` | `companies_house_detail_fetch_action.go` | Fetch profile/officers/PSC, derive succession risk, store enrichment |
| — | `ch_vertical_profiles.go` | Vertical-specific heuristics registry (generic words, keywords, suffixes) |
| `load_ch_enrichment_batch` | `companies_house_actions.go` | Legacy: load businesses for per-business search |
| `companies_house_search` | `companies_house_actions.go` | Legacy: search CH API by name |
| `companies_house_fetch` | `companies_house_actions.go` | Fetch company profile, officers, PSC (reused by detail fetch) |
| `store_ch_enrichment` | `companies_house_actions.go` | Legacy: write enrichment data to DB |

### Shared Helpers

The detail fetch action reuses helpers from `companies_house_actions.go` — no duplicated logic:
`chAPIGet`, `extractOfficersList`, `extractPSCList`, `deriveOwnerSignals`, `deriveSuccessionRisk`, `nullStr`, `nullInt`, `pgArrayFromInterface`.

## What Companies House Gives Us

The Companies House API (free, API key for rate limits) provides:

**Company Profile** (from `/company/{number}`): registered name, number, status, incorporation date, SIC codes, registered office address, company type.

**Officers** (from `/company/{number}/officers`): directors/secretaries with appointment dates and date of birth (month + year) for succession risk scoring.

**PSC** (from `/company/{number}/persons-with-significant-control`): beneficial owners (25%+ shareholding) for group ownership identification.

**Filing History → Accounts** (planned): annual accounts as iXBRL with tagged financial values — net assets, employee count, turnover (if disclosed).

**API:** Base URL `https://api.company-information.service.gov.uk`. Rate limits: 600 requests per 5 minutes. HTTP basic auth with API key as username.

## Stage 1: Bulk Collection (Implemented)

The `ch_bulk_collect` action paginates through CH advanced search for a given SIC code. All pagination is handled inside the single action call. The action ensures the table exists, fetches 100 companies per page with 2s delay, upserts each company, and populates `company_name_cleaned` for trigram matching.

**Result:** 5,780 active SIC 75000 companies in 58 API calls (~2.5 minutes).

### Table: `business_intel.ch_vet_companies`

Key columns: `company_number` (PK), `company_name`, `company_name_cleaned`, `postcode`, `postcode_prefix`, `locality`, `matched_business_id`, `match_confidence`, `match_method`, `details_fetched`, `details_fetched_at`.

Match methods: `local_postcode`, `local_postcode_exact`, `name_trigram`, `pending_llm_review`, `llm_confirmed`, `llm_uncertain`.

### Indexes

- `idx_ch_vet_postcode_prefix` — B-tree on `postcode_prefix` WHERE `company_status = 'active'`
- `idx_ch_vet_name_trgm_gist` — GiST trigram on `lower(company_name)` for `<->` distance queries
- `idx_ch_vet_name_cleaned_gist` — GiST trigram on `company_name_cleaned` for pass 2 matching (~4ms per lookup)

### Notable Data Points

- **SK9 3RN**: 554 Vets4Pets companies (Pets at Home HQ, Cheadle)
- **YO30 4UZ**: 68 IVC Evidensia/Linnaeus companies (York HQ)
- **TA1 3DU**: Company formation agent address

## Stage 2: Local Matching (Implemented)

### Two-Pass Architecture

The `ch_local_match` action processes all unmatched businesses in a single call (~10 seconds total). It loads a `CHVerticalProfile` from the profile registry based on `vertical_slug` in step config, which provides industry keywords, generic words, and scoring bonuses.

**Pass 1 — Postcode + name scoring**

For each business, query `ch_vet_companies` by postcode prefix, then score candidates:

| Factor | Score | Notes |
|---|---|---|
| Postcode prefix match | +0.20 | Always present (filtered by query) |
| Exact full postcode | +0.15 | Bonus when inward code also matches |
| Name exact (cleaned) | +0.30 | After stripping suffixes from profile |
| Name contains | +0.20 | One cleaned name contains the other |
| Name word overlap | 0 to +0.25 | Proportion of words >3 chars in both |
| Industry keyword in CH name | +bonus from profile | Default +0.10 for veterinary |
| **Threshold** | **0.40** | |

**Pass 2 — Trigram name-only (no postcode requirement)**

For businesses not matched in pass 1, search the entire table using GiST trigram index on `company_name_cleaned`:

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

Three tiers:
- Similarity ≥ 0.90 + distinctive word → **auto-accept** (`name_trigram`)
- Similarity 0.50–0.90 + distinctive word → **store as `pending_llm_review`**
- Similarity < 0.50 or no distinctive word → **reject**

### Distinctive Word Overlap Check

Prevents false positives where common industry words inflate trigram scores. Requires at least one non-generic word from the business name to appear in the CH name. The generic words list comes from `profile.GenericWords`.

Examples:
- "WW Mobile Veterinary Services" → "HAYLOFT MOBILE VETERINARY SERVICES" — "ww" too short, rest generic → **rejected** ✓
- "Shropshire Farm Vets" → "SHROPSHIRE FARM VETS LIMITED" — "shropshire" is distinctive → **accepted** ✓

### Name Cleaning

`cleanCompanySearchName` strips suffixes for search generality, keeping industry terms. Principle: **search generality → scoring specificity**. "Castle Veterinary Group" → "Castle Veterinary" finds both "CASTLE VETERINARY SURGERY LTD" and "CASTLE VETERINARY CENTRE LTD".

## Stage 3: LLM Review (Implemented)

The `ch_llm_review` action is generic — industry-specific context comes from step config, not hardcoded in Go.

### How It Works

1. Load all rows where `match_method = 'pending_llm_review'`, including business group_name and town
2. Batch 15 pairs per LLM call
3. Build prompt with industry context from `config.industry_context` and `config.industry_name`
4. LLM responds YES/NO/UNCERTAIN for each
5. Update: `llm_confirmed`, clear match (`llm_rejected`), or `llm_uncertain`

### Configuration

`ai_service` is in the step config (not agent_config top-level) because on multi-type pods, agent_config comes from the pod's own type, not the message's type.

Industry context (corporate groups, formation addresses, name patterns) is in `config.industry_context` — a free-text block injected into the prompt. Changing the vet-specific knowledge requires only a SQL update to the agent definition, not a code change.

### Prompt Design

The prompt includes for each candidate: business name, town, postcode, group ownership, CH company name, registered postcode, similarity score.

Key improvements over initial version:
- **Group context** — LLM sees "Group: Independent" vs "VETS4PETS LIMITED" and knows they're different brands
- **Formation agent addresses** — prompt explicitly names SK9 3RN, YO30 4UZ, B90 4BN with their owners
- **Brand mismatch rules** — "If listed as Independent but matched to a corporate chain, that's NO"

Results: 141 confirmed, 181 rejected (49% rejection rate), 15 uncertain. Vets4Pets false positives correctly rejected — only actual Vets4Pets/Companion Care practices confirmed at SK9 3RN.

### Cost

~367 candidates × ~100 tokens per judgment = ~37,000 tokens. On Haiku: ~$0.05.

## Stage 4: Detail Fetch (Deploying)

The `ch_detail_fetch` action fetches company profile, officers, and PSC from the CH API for confirmed matches, derives succession risk signals, and stores enrichment data.

### How It Works

1. Load from `ch_vet_companies` where `matched_business_id IS NOT NULL AND details_fetched = false` and match_method not pending/uncertain
2. For each company: fetch profile, officers (15s delay), PSC (15s delay) via `chAPIGet`
3. Derive signals using existing `extractOfficersList`, `deriveOwnerSignals`, `deriveSuccessionRisk`
4. Upsert into `companies_house_data` (same schema as legacy `StoreCHEnrichmentAction`)
5. Mark `details_fetched = true` in `ch_vet_companies`

### Timing

50 companies per batch × ~45s each (3 API calls at 15s delay) = ~37 minutes per batch. Scheduled every 20 minutes — pre_query only fires when unfetched matches exist. ~640 confirmed matches = ~13 batches = ~8 hours to complete backfill.

### What Gets Stored

`companies_house_data` receives: company_number, company_name, status, type, incorporation_date, SIC codes, registered_address, officers (JSONB), PSC (JSONB), parent_company_name/number, owner signals (name, DOB, age, tenure), succession_risk (high/medium/low/unknown), match_confidence, match_method.

## Stage 5: Discovery (Planned)

~5,100 unmatched CH companies with SIC 75000 are potential new leads — vet companies we don't have in our businesses table. Some are equine, agricultural, holding companies, or pet shops. A discovery workflow could look up registered addresses, search for websites, and feed new practices into the verification pipeline.

## Vertical Generalisation (Implemented)

The pipeline is now structured for multi-vertical support. Industry-specific heuristics are separated from the generic matching logic.

### Vertical Profile Registry

File: `ch_vertical_profiles.go`

```go
type CHVerticalProfile struct {
    Slug                 string
    SICCodes             []string          // for bulk collection
    IndustryKeywords     []string          // scoring bonus in pass 1
    IndustryKeywordBonus float64           // bonus amount
    GenericWords         map[string]bool   // excluded from distinctive word check
    NameStripSuffixes    []string          // for name cleaning
    BusinessVerticalSlug string            // filter on businesses table
}
```

Currently has `veterinary` profile. Adding a new vertical (e.g. seaweed farming) requires:

### What lives where

| What | Where | To add a vertical |
|---|---|---|
| SIC codes, keywords, generic words, suffixes | `ch_vertical_profiles.go` | Add new map entry |
| Vertical slug | Workflow SQL step config (`vertical_slug`) | New agent definition |
| Corporate groups, formation addresses | Workflow SQL step config (`industry_context`) | New agent definition |
| LLM model choice | Workflow SQL step config (`ai_service`) | New agent definition |
| Table name `ch_vet_companies` | Hardcoded in SQL queries | Rename to `ch_sic_companies` + `collection_slug` column |

The Go actions (`ch_local_match`, `ch_llm_review`, `ch_bulk_collect`, `ch_detail_fetch`) are generic — no vet-specific code in the action logic. The vertical profile and workflow config provide all industry context.

## Data Storage

### `business_intel.companies_house_data` (enrichment output)

Stores detailed enrichment data for matched companies. Populated by `ch_detail_fetch` (new pipeline) or `store_ch_enrichment` (legacy). Key fields: company_number, officers (JSONB), psc (JSONB), owner signals, succession_risk.

### `business_intel.ch_vet_companies` (local mirror)

Local copy of all CH companies with SIC 75000. Populated by `ch_bulk_collect`, matched by `ch_local_match`, reviewed by `ch_llm_review`, fetched by `ch_detail_fetch`.

## Owner Age & Succession Signals

The CH Officers endpoint provides month and year of birth for every director. The `deriveOwnerSignals` and `deriveSuccessionRisk` functions (in `companies_house_actions.go`) compute:

| Signal | High risk | Medium risk | Low risk |
|--------|-----------|-------------|----------|
| Owner age | 60+ | 50-59 | Under 50 |
| Tenure | 20+ years | 10-19 years | Under 10 |
| Corporate PSC | No (independent) | — | Yes (already acquired) |
| Single director | Yes | — | Multiple (succession planned) |

## Operational Notes

### Scheduled Tasks

```
ch-vet-collect:   disabled (run manually, monthly refresh)
ch-local-match:   daily, concurrency: ch-matching
ch-llm-review:    daily, concurrency: ch-matching
ch-detail-fetch:  every 20 min, concurrency: ch-enrichment
ch-enrichment:    disabled (legacy)
```

### Single-Pod Contention

The business-intel pod processes one orchestration at a time. Matching and enrichment use different concurrency groups (`ch-matching` vs `ch-enrichment`) but still queue on the same pod. The pre_query on each task prevents triggering when there's no work.

### Message Routing for Different Agent Types

The scheduler sends `{"config":{"agent_type":"ch-detail-fetcher"}}` to the business-intel requests topic. The pod's `selectWorkflow`:
1. Loads business-intel agent definition (Priority 3 fallback)
2. Extracts `config.agent_type` from the message
3. Calls `FindBestGroup("ch-detail-fetcher")` → looks up agent_definitions
4. Uses the ch-detail-fetcher workflow (Priority 2)

### ai_service on Shared Pods

On multi-type pods, `agent_config` in collected_data comes from the pod's own type (business-intel), not the message's type. So `ai_service` for LLM actions must be in the step config, not agent_config top-level. This is documented in `ch_llm_review_action.go`.

## Financial Data from Filed Accounts (Planned)

A `companies_house_fetch_accounts` action would fetch iXBRL documents and parse `ix:nonFraction` elements for net assets, total assets, employee count, turnover (if disclosed). Most vet practices file as small/micro.

## Debug History

1. **Stale loop variable** — `ExtractNestedField` → `ExtractFields`.
2. **Name cleaning min length** — Raised from 3 to 6 chars.
3. **Workflow condition** — `!= 0` → `== true`.
4. **SIC filter** — Exact → prefix-based with `strings.HasPrefix`.
5. **SIC penalty** — `-0.30` → `-0.05` (empty) / `-0.15` (wrong).
6. **Address field mismatch** — Added `Address` struct field for search API's `"address"` key.
7. **OOM kills** — Batch 20→10, memory to 3Gi.
8. **Kustomize patch target** — `name: vet-intel` → `name: business-intel`.
9. **Vet industry title bonus** — +0.15 "veterinary", +0.10 "vet"/"vets".
10. **Name cleaner** — Keep "Veterinary"/"Vets", strip generic suffixes. Search generality → scoring specificity.
11. **SK9 formation agent** — 554 Vets4Pets at SK9 3RN. Expected — corporate branches don't match by local postcode.
12. **Trigram index type** — GIN → GiST for `<->` distance operator (~4ms per lookup).
13. **Cleaned name column** — "SHROPSHIRE FARM VETS LIMITED" similarity: 0.72 raw → 1.0 cleaned.
14. **LLM false positives** — Initial prompt confirmed "YourVets Nuneaton" → "NUNEATON VETS4PETS LIMITED". Fixed by adding group_name, town, and corporate group knowledge to prompt. 49% rejection rate on ambiguous candidates.
15. **ai_service location** — On shared pods, agent_config comes from pod type, not message type. Moved ai_service to step config for LLM actions.

## GDPR & Legal Notes

- All Companies House data is public record (filing is a legal requirement)
- Processing for legitimate business purposes (market analysis) is lawful under Article 6(1)(f)
- The `succession_risk` field is derived from public data, not from private profiling
- If data is sold to third parties: note the data source and legal basis; present age-related fields as "company maturity indicators" rather than personal age profiling
