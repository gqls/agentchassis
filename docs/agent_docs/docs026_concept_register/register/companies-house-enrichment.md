# Register — companies-house-enrichment

4 concepts, consolidated from 10 raw extractions (5 unique blocks, each appearing
twice due to exact whole-block duplication in the cluster input file) across units
U01, U18, U19, U24f.

### CH-001 — Companies House enrichment pipeline (bulk collect → local match → LLM review → detail fetch)
- **status:** deployed
- **status-evidence:** Doc 017 status header with March 19 2026 results (5,780 companies, 634 confirmed matches, 23.2%); concrete SQL build-out (077–083) sequential with scheduled tasks; 100 portfolio doc: "Thousands of veterinary practices collected, verified against Companies House records, and enriched with financial data."
- **what:** Multi-stage enrichment on the business-intel pod, described consistently by a narrative doc (017) and by the concrete agent SQL migrations that implement it (077–083): bulk SIC 75000 collection into a trigram-indexed local mirror (ch-collector, paginated, rate-limited); two-pass local matching (~10s, no API) — postcode+name scoring cascade then GiST trigram name-only with three tiers (≥0.90+distinctive auto-accept / 0.50–0.90 pending_llm_review / reject) plus a distinctive-word check against generic-word inflation (ch-matcher, threshold 0.40); Haiku LLM review in batches of 15 with industry context from step config, ~$0.05/run (ch-llm-reviewer, classifies confirmed/rejected/uncertain); detail fetch of profile/officers/PSC with succession risk derived from officer DOBs (ch-detail-fetcher); ch-company-scraper regex-extracts registration numbers from business website footers (generic across verticals); ch-accounts-fetcher parses filed iXBRL accounts into financial columns. ch-enricher (077) was the original combined agent, later renamed business-intel. Discovery of unmatched companies is planned but not yet built.
- **sources:** 017 full; 077_business_intel_companies_house.sql; 079_companies_house_ch_matcher.sql; 080_companies_house_ch_llm_reviewer.sql; 082_company_number_scraper_ch_company_scraper.sql; 083_companies_house_ch_accounts_fetcher.sql
- **relations:** Vertical profile registry (CH-002); Companies House enrichment with succession-risk signals (CH-003, the underlying schema); Companies House matching cascade revision (CH-004, the forward plan for this pipeline); vet vertical pipeline (verified businesses input, business-intel-collection.md BIC-001)
- **verify-later:** ch_vet_companies match_method distribution; cascade implemented?; ch_* actions; match-rate stats views; scheduled task cadence

### CH-002 — Vertical profile registry (generic-words/keywords/suffixes per industry)
- **status:** deployed
- **status-evidence:** 017 "Vertical Generalisation (Implemented)"; ch_vertical_profiles.go.
- **what:** Matching heuristics (industry keywords, generic word lists, scoring bonuses, name-cleaning suffixes) live in a Go profile registry keyed by vertical_slug from step config; LLM industry context is free-text in the agent definition config — new verticals are config, not code. Principle: search generality → scoring specificity.
- **sources:** 017#Vertical Generalisation, #Name Cleaning
- **relations:** Companies House enrichment pipeline (CH-001, ch-matcher / ch-llm-reviewer)
- **verify-later:** ch_vertical_profiles.go registry contents

### CH-003 — Companies House enrichment with succession-risk signals
- **status:** deployed
- **status-evidence:** Schema + scheduled task (ch-enrichment every 20 min, seeded disabled "until Go actions are built") and a later applied accounts-fetch migration (accounts_fetched tracking, financial columns), indicating progression to live collection.
- **what:** Post-verification enrichment of business_intel.businesses: company identity/status/SIC, financials from filed accounts (accounts_type micro/small/medium/full, assets/net worth/turnover/PL, employees), officers and PSC JSONB, and derived owner-age/succession signals (owner_dob from CH month/year, estimated age, tenure, is_sole_director, is_corporate_owned → succession_risk high/medium/low/acquired). Deliberately polite rate limiting (~7% of CH's 600 req/5min). Match metadata records confidence/method/search query; accounts fetch is tracked separately on ch_vet_companies with an LLM-review exclusion filter.
- **sources:** docs/agent_docs/sql_for_tables/023_companies_house_data.sql
- **relations:** Companies House enrichment pipeline (CH-001, this is its underlying schema); http_request_log rate monitoring; vet vertical
- **verify-later:** ch-enricher agent; enrichment coverage counts

### CH-004 — Companies House matching cascade (revised 7-tier signal architecture)
- **status:** partial
- **status-evidence:** "Current matching achieves 676/2,767 (24.4%)... Target: 70-80% automated match rate, with HITL bringing it to 90%+." Presented as a v2 revision of an earlier plan (v1, docs019 archive, out of this unit's scope); the enrichment domain has a live anchor doc (017, CH-001).
- **what:** A priority-ordered cascade replacing the flat two-pass matcher in CH-001, where each business flows down tiers (matched / pass-to-next / queue-for-HITL) until resolved: Tier 0 scrapes the practice's own website for a company registration number (definitive, no CH API cost); Tier 1 exact-name+geography; Tier 2 exact-name unique-in-CH regardless of geography; Tier 3 postcode+moderate-name (raised threshold from 0.40→0.50 with a mandatory name-overlap component to cut false positives); Tier 4 LLM review with the top-3 trigram candidates and full business context (not just the single best match); Tier 5 corporate-group-parent mapping for chains sharing one CH registration (Medivet, CVS, Vets4Pets, IVC Evidensia, VetPartners — addressing ~800 corporate-branch businesses a per-business match can never resolve 1:1); Tier 6 a human-review HITL queue for the remainder. New tables proposed: `businesses.company_number_scraped`, `ch_match_candidates`, `ch_corporate_groups`.
- **sources:** docs/_archive/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/_archive/archive_april_26/022b_companies_house_matching_cascade_plan_v2.md
- **relations:** Companies House enrichment pipeline (CH-001, the anchor it revises); the v1 predecessor plan (022_companies_houise_matching_cascade_plan.md, out of scope)
- **verify-later:** business_intel.ch_match_candidates, business_intel.ch_corporate_groups, business_intel.businesses.company_number_scraped — confirm which tiers actually shipped and current match rate
