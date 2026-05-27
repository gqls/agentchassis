# Companies House Matching — Revised Cascade Plan

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
