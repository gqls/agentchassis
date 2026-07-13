# 017 — Companies House Enrichment

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
-e 

---

## Matching Cascade — Revised Plan (from 022b)

## Problem

Current matching achieves 676/2,767 (24.4%). Analysis shows the gap breaks down as:

| Category | Est. Count | Why unmatched |
|---|---|---|
| Corporate branches | ~800 | 50 Medivet branches share one CH registration |
| Weak pass 1 matches blocking better ones | ~100 | 0.40 threshold too permissive, grabs wrong company |
| Name mismatch (different trading vs registered name) | ~400-600 | "Acacia Vets" registered as "DR J SMITH VETERINARY SERVICES LTD" |
| Sole traders/partnerships | ~300-500 | Not on Companies House at all |
| Low confidence / not real practices | ~180 | Directories, aggregator listings |
| Missing/bad postcodes | ~37 | Can't match geographically |

## Revised Architecture: Matching Cascade

A priority-ordered cascade where each tier uses the most specific signal available. Businesses flow down through tiers until matched or exhausted. Each tier produces one of: `matched`, `pass_to_next_tier`, or `queue_for_hitl`.

### Pre-filter

Exclude from matching:
- `confidence_score < 0.40` (~180 businesses)
- `business_type` containing "directory" or "listing" (~46)
- Businesses already matched in `ch_vet_companies`

Reduces matching universe from 2,767 to ~2,550.

### Tier 0 — Company Number from Website Scrape

**Signal:** Company registration number found on the practice's website.
**Confidence:** Definitive — company numbers are unique identifiers.
**Method:** For each business with `website_url`, fetch the homepage and key pages (about, footer, terms, privacy policy). Extract company number using regex patterns:
- `company\s*(number|no\.?|reg\.?)\s*:?\s*(\d{7,8})`
- `registered\s*(in\s*(england|wales|scotland))?\s*(number|no\.?)?\s*:?\s*(\d{7,8})`
- `SC\d{6}` (Scottish companies)
- `NI\d{6}` (Northern Irish companies)

Look up extracted number against `ch_vet_companies.company_number`. Direct match, no ambiguity.

**Expected yield:** ~30-40% of businesses with websites show registration numbers. That's ~750-1,000 matches.
**Cost:** Web fetches only, no CH API calls. Rate limit web fetches to be polite (~1/sec).
**Implementation:** New action `ch_scrape_company_number`. Stores extracted number in a new column `businesses.company_number_scraped` for reuse. Matches directly against `ch_vet_companies`.

### Tier 1 — Exact Name + Geographic Confirmation

**Signal:** Cleaned name similarity ≥ 0.90 AND (same postcode prefix OR same town).
**Confidence:** High — distinctive name + geographic proximity.
**Method:** For each unmatched business, query `ch_vet_companies` by `company_name_cleaned` similarity ≥ 0.90. If any candidate shares postcode prefix or has the same town/locality, auto-accept.

**Expected yield:** Most of the current pass 1 exact matches + some that currently fall to pass 2.
**Implementation:** Part of the revised `ch_local_match` action.

### Tier 2 — Exact Name, Unique in CH

**Signal:** Cleaned name similarity ≥ 0.90, different geography, but only one CH company has that cleaned name.
**Confidence:** High — if there's only one "MARLOW VETS LIMITED" in all of CH, it must be our Marlow Vets.
**Method:** Same trigram query, but check if the result is unique. If only one CH company matches at ≥ 0.90, auto-accept regardless of geography.

**Expected yield:** Catches the ~39 "blocked_by_existing_match" cases and similar.
**Implementation:** Part of the revised `ch_local_match` action.

### Tier 3 — Postcode + Moderate Name (Revised Pass 1)

**Signal:** Same postcode prefix + name scoring above threshold.
**Confidence:** Moderate — requires genuine name overlap, not just vet words.
**Method:** Current postcode+name scoring but with raised threshold (0.50 instead of 0.40) and a minimum name similarity requirement. The postcode bonus (+0.20) alone plus a vet word (+0.10) shouldn't clear the threshold — there must be actual name overlap.

**Expected yield:** Similar to current pass 1 but fewer false positives.
**Implementation:** Part of the revised `ch_local_match` action.

### Tier 4 — LLM Review with Full Context

**Signal:** Trigram candidates 0.50-0.90, reviewed by LLM with full business context.
**Confidence:** Moderate to high depending on LLM judgment.
**Method:** For each unmatched business, find top 3 CH candidates by trigram similarity. Present to LLM with: business name, town, county, postcode, group_name, business_type, head_vet_name (if available), website domain. LLM picks the best match or says "none".

Key improvement over current approach: present multiple candidates per business, not just the single best trigram match. The LLM can compare and choose.

**Expected yield:** ~100-200 additional matches from the ambiguous zone.
**Cost:** ~$0.05-0.10 on Haiku per run.
**Implementation:** Revised `ch_llm_review` action.

### Tier 5 — Corporate Group Parent Matching

**Signal:** Business group_name maps to a known CH parent company.
**Confidence:** High for the mapping, but it's a many-to-one relationship.
**Method:** Maintain a lookup table of group_name → CH company_number(s):
- "Medivet" → "MEDIVET GROUP LIMITED" (various company numbers)
- "CVS Vets" / "CVS Group" → "CVS GROUP PLC"
- "Vets4Pets" / "Vets for Pets" → parent + "[LOCATION] VETS4PETS LIMITED"
- "IVC Evidensia" / "Linnaeus" → IVC parent companies
- "VetPartners" → "VETPARTNERS LIMITED"

This doesn't give a 1:1 match but links the branch to its corporate parent. Store as `match_method = 'group_parent'` with the parent company_number.

For Vets4Pets specifically: match "[Location] Vets4Pets" → "[LOCATION] VETS4PETS LIMITED" at SK9 using the location word.

**Expected yield:** ~300-500 of the 800 corporate branches.
**Implementation:** New action `ch_group_match` with a configurable group→company mapping in step config.

### Tier 6 — HITL Queue

**Signal:** Human judgment.
**Confidence:** Highest — human-verified.
**Method:** All remaining unmatched businesses are queued for human review. The admin dashboard shows:
- Business details (name, town, postcode, website, group, business_type)
- Top 5 CH candidates ranked by combined signals
- One-click approve, reject, or skip

The HITL queue is populated by the matching pipeline and drained by admin users. Approved matches are stored with `match_method = 'manual'`.

**Implementation:** Populate a `ch_match_candidates` table (or use existing admin API patterns). Dashboard integration is separate work — noted but not built here.

## Data Changes

### New column on businesses table

```sql
ALTER TABLE business_intel.businesses 
    ADD COLUMN IF NOT EXISTS company_number_scraped VARCHAR(10);
```

### New table for HITL candidates (optional, could use existing admin patterns)

```sql
CREATE TABLE IF NOT EXISTS business_intel.ch_match_candidates (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    business_id UUID NOT NULL REFERENCES business_intel.businesses(id),
    company_number VARCHAR(10) NOT NULL,
    company_name TEXT,
    match_score NUMERIC(3,2),
    match_signals JSONB,      -- {name_sim: 0.85, same_town: true, same_postcode: false}
    rank INTEGER,              -- 1 = best candidate
    status TEXT DEFAULT 'pending',  -- pending, approved, rejected
    reviewed_by UUID,
    reviewed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW()
);
```

### Corporate group lookup

```sql
CREATE TABLE IF NOT EXISTS business_intel.ch_corporate_groups (
    group_pattern TEXT NOT NULL,          -- regex or exact match on businesses.group_name
    company_number VARCHAR(10),           -- CH company number of parent
    company_name TEXT,
    match_type TEXT DEFAULT 'parent',     -- parent, branch_pattern
    branch_pattern TEXT,                  -- e.g. "{LOCATION} VETS4PETS LIMITED"
    notes TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW()
);
```

## Actions

| Action | Tier | Description |
|---|---|---|
| `ch_scrape_company_number` | 0 | Scrape websites for company registration numbers |
| `ch_local_match` (revised) | 1, 2, 3 | Cascading name+geography matching |
| `ch_llm_review` (revised) | 4 | LLM review with multiple candidates per business |
| `ch_group_match` | 5 | Corporate group parent matching |
| `ch_populate_hitl` | 6 | Populate HITL queue with remaining candidates |

## Implementation Order

1. **Pre-filter** — exclude low-confidence and directories from matching
2. **Revise `ch_local_match`** — implement tiers 1-3 as the new cascade, replacing current two-pass
3. **Tier 0: `ch_scrape_company_number`** — highest value, definitive matches
4. **Tier 5: `ch_group_match`** — addresses the largest gap (800 corporate branches)
5. **Tier 4: Revise `ch_llm_review`** — present multiple candidates
6. **Tier 6: HITL population** — depends on admin dashboard progress

## Expected Final Match Rates

| Tier | New Matches | Running Total | % of 2,550 |
|---|---|---|---|
| Pre-filter | -180 excluded | 2,550 universe | — |
| Tier 0 (company number scrape) | ~750-1,000 | ~750-1,000 | 29-39% |
| Tier 1 (exact name + geography) | ~400-500 | ~1,150-1,500 | 45-59% |
| Tier 2 (exact name, unique) | ~50-80 | ~1,200-1,580 | 47-62% |
| Tier 3 (postcode + name) | ~100-150 | ~1,300-1,730 | 51-68% |
| Tier 4 (LLM review) | ~100-200 | ~1,400-1,930 | 55-76% |
| Tier 5 (corporate groups) | ~300-500 | ~1,700-2,430 | 67-95% |
| Tier 6 (HITL) | remaining | ~2,100-2,550 | 82-100% |
| Unmatched (sole traders etc.) | ~100-300 | — | — |

Target: **70-80% automated match rate**, with HITL bringing it to 90%+.
