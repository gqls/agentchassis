
<!-- SOURCE: U01_docs024_numbered_core.md -->
### Companies House enrichment pipeline (bulk collect → local match → LLM review → detail fetch)
- **category:** companies-house-enrichment
- **status-signal:** deployed
- **status-evidence:** 017 status header with March 19 2026 results (5,780 companies, 634 confirmed matches, 23.2%)
- **what:** Five stages on the business-intel pod: bulk SIC collection into a local mirror (trigram-indexed); two-pass local matching (~10s, no API): postcode+name scoring cascade then GiST trigram name-only with three tiers (≥0.90+distinctive auto-accept / 0.50-0.90 pending_llm_review / reject) and a distinctive-word check against generic-word inflation; haiku LLM review in batches of 15 with industry context from step config (~$0.05/run); detail fetch (profile/officers/PSC, succession risk from officer DOBs); discovery of unmatched companies planned. Revised matching cascade (tiers 0–6: website company number, exact+geo, exact-unique, postcode+moderate, LLM full-context, corporate-group parent, HITL queue) is the forward plan.
- **sources:** 017 full
- **relations:** vertical profile registry; business-intel pod pattern
- **verify-later:** ch_vet_companies match_method distribution; cascade implemented?

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Vertical profile registry (generic-words/keywords/suffixes per industry)
- **category:** companies-house-enrichment
- **status-signal:** deployed
- **status-evidence:** 017 "Vertical Generalisation (Implemented)"; ch_vertical_profiles.go
- **what:** Matching heuristics (industry keywords, generic word lists, scoring bonuses, name-cleaning suffixes) live in a Go profile registry keyed by vertical_slug from step config; LLM industry context is free-text in the agent definition config — new verticals are config, not code. Principle: search generality → scoring specificity.
- **sources:** 017#Vertical Generalisation, #Name Cleaning
- **relations:** ch_local_match; generic LLM review
- **verify-later:** ch_vertical_profiles.go registry contents

<!-- SOURCE: U18_sql_for_agents.md -->
### Companies House enrichment chain (business-intel / ch-* agents)
- **category:** companies-house-enrichment
- **status-signal:** deployed
- **status-evidence:** 077–083 sequential build-out with scheduled tasks; 100 portfolio: "Thousands of veterinary practices collected, verified against Companies House records, and enriched with financial data."
- **what:** Multi-stage enrichment on the business-intel pod: ch-collector bulk-mirrors all SIC 75000 companies into ch_vet_companies (paginated, rate-limited); ch-matcher matches verified businesses against the mirror by postcode + name similarity (pure SQL/Go scoring, threshold 0.40, no API); ch-llm-reviewer classifies ambiguous matches (Haiku, 15 pairs/batch) as confirmed/rejected/uncertain; ch-detail-fetcher pulls profile/officers/PSC for confirmed matches and derives succession-risk signals; ch-company-scraper regex-extracts registration numbers from business website footers (generic across verticals); ch-accounts-fetcher parses filed iXBRL accounts into financial columns (net assets, turnover, employees). ch-enricher (077, renamed business-intel) was the original combined agent.
- **sources:** 077_business_intel_companies_house.sql; 079_companies_house_ch_matcher.sql; 080_companies_house_ch_llm_reviewer.sql; 082_company_number_scraper_ch_company_scraper.sql; 083_companies_house_ch_accounts_fetcher.sql
- **relations:** vet vertical pipeline (verified businesses input); scheduled_tasks entries per agent
- **verify-later:** ch_* actions; match-rate stats views; scheduled task cadence

<!-- SOURCE: U19_sql_tables_components.md -->
### Companies House enrichment with succession-risk signals
- **category:** companies-house-enrichment
- **status-signal:** deployed
- **status-evidence:** Schema + scheduled task (ch-enrichment every 20 min, seeded disabled "until Go actions are built") and a later applied accounts-fetch migration (accounts_fetched tracking, financial columns), indicating progression to live collection.
- **what:** Post-verification enrichment of business_intel.businesses: company identity/status/SIC, financials from filed accounts (accounts_type micro/small/medium/full, assets/net worth/turnover/PL, employees), officers and PSC JSONB, and derived owner-age/succession signals (owner_dob from CH month/year, estimated age, tenure, is_sole_director, is_corporate_owned → succession_risk high/medium/low/acquired). Deliberately polite rate limiting (~7% of CH's 600 req/5min). Match metadata records confidence/method/search query; accounts fetch is tracked separately on ch_vet_companies with an LLM-review exclusion filter.
- **sources:** docs/agent_docs/sql_for_tables/023_companies_house_data.sql
- **relations:** business-intel collection pipeline; http_request_log rate monitoring; vet vertical.
- **verify-later:** ch-enricher agent; enrichment coverage counts.

<!-- SOURCE: U24f_docs_archive_remaining_small.md -->
### Companies House matching cascade (revised 7-tier signal architecture)
- **category:** companies-house-enrichment
- **status-signal:** partial
- **status-evidence:** "Current matching achieves 676/2,767 (24.4%)... Target: 70-80% automated match rate, with HITL bringing it to 90%+." Presented as a revision (v2) of an earlier plan (docs/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/_archive/old/022_companies_houise_matching_cascade_plan.md, a v1 outside this unit's scope), and the enrichment domain itself has a live anchor doc (docs/agent_docs/docs024_key_docs_latest/017_companies_house_enrichment.md).
- **what:** A priority-ordered cascade replacing a flat two-pass matcher, where each business flows down tiers (matched / pass-to-next / queue-for-HITL) until resolved: Tier 0 scrapes the practice's own website for a company registration number (definitive, no CH API cost); Tier 1 exact-name+geography; Tier 2 exact-name unique-in-CH regardless of geography; Tier 3 postcode+moderate-name (raised threshold from 0.40→0.50 with a mandatory name-overlap component to cut false positives); Tier 4 LLM review with the top-3 trigram candidates and full business context (not just the single best match); Tier 5 corporate-group-parent mapping for chains sharing one CH registration (Medivet, CVS, Vets4Pets, IVC Evidensia, VetPartners — addressing ~800 corporate-branch businesses that a per-business match can never resolve 1:1); Tier 6 a human-review HITL queue for the remainder. New tables proposed: `businesses.company_number_scraped`, `ch_match_candidates`, `ch_corporate_groups`.
- **sources:** docs/_archive/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/_archive/archive_april_26/022b_companies_house_matching_cascade_plan_v2.md
- **relations:** companies-house-enrichment (017 anchor doc); the v1 predecessor plan (022_companies_houise_matching_cascade_plan.md)
- **verify-later:** business_intel.ch_match_candidates, business_intel.ch_corporate_groups, business_intel.businesses.company_number_scraped — confirm which tiers actually shipped and current match rate

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Companies House enrichment pipeline (bulk collect → local match → LLM review → detail fetch)
- **category:** companies-house-enrichment
- **status-signal:** deployed
- **status-evidence:** 017 status header with March 19 2026 results (5,780 companies, 634 confirmed matches, 23.2%)
- **what:** Five stages on the business-intel pod: bulk SIC collection into a local mirror (trigram-indexed); two-pass local matching (~10s, no API): postcode+name scoring cascade then GiST trigram name-only with three tiers (≥0.90+distinctive auto-accept / 0.50-0.90 pending_llm_review / reject) and a distinctive-word check against generic-word inflation; haiku LLM review in batches of 15 with industry context from step config (~$0.05/run); detail fetch (profile/officers/PSC, succession risk from officer DOBs); discovery of unmatched companies planned. Revised matching cascade (tiers 0–6: website company number, exact+geo, exact-unique, postcode+moderate, LLM full-context, corporate-group parent, HITL queue) is the forward plan.
- **sources:** 017 full
- **relations:** vertical profile registry; business-intel pod pattern
- **verify-later:** ch_vet_companies match_method distribution; cascade implemented?

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Vertical profile registry (generic-words/keywords/suffixes per industry)
- **category:** companies-house-enrichment
- **status-signal:** deployed
- **status-evidence:** 017 "Vertical Generalisation (Implemented)"; ch_vertical_profiles.go
- **what:** Matching heuristics (industry keywords, generic word lists, scoring bonuses, name-cleaning suffixes) live in a Go profile registry keyed by vertical_slug from step config; LLM industry context is free-text in the agent definition config — new verticals are config, not code. Principle: search generality → scoring specificity.
- **sources:** 017#Vertical Generalisation, #Name Cleaning
- **relations:** ch_local_match; generic LLM review
- **verify-later:** ch_vertical_profiles.go registry contents

<!-- SOURCE: U18_sql_for_agents.md -->
### Companies House enrichment chain (business-intel / ch-* agents)
- **category:** companies-house-enrichment
- **status-signal:** deployed
- **status-evidence:** 077–083 sequential build-out with scheduled tasks; 100 portfolio: "Thousands of veterinary practices collected, verified against Companies House records, and enriched with financial data."
- **what:** Multi-stage enrichment on the business-intel pod: ch-collector bulk-mirrors all SIC 75000 companies into ch_vet_companies (paginated, rate-limited); ch-matcher matches verified businesses against the mirror by postcode + name similarity (pure SQL/Go scoring, threshold 0.40, no API); ch-llm-reviewer classifies ambiguous matches (Haiku, 15 pairs/batch) as confirmed/rejected/uncertain; ch-detail-fetcher pulls profile/officers/PSC for confirmed matches and derives succession-risk signals; ch-company-scraper regex-extracts registration numbers from business website footers (generic across verticals); ch-accounts-fetcher parses filed iXBRL accounts into financial columns (net assets, turnover, employees). ch-enricher (077, renamed business-intel) was the original combined agent.
- **sources:** 077_business_intel_companies_house.sql; 079_companies_house_ch_matcher.sql; 080_companies_house_ch_llm_reviewer.sql; 082_company_number_scraper_ch_company_scraper.sql; 083_companies_house_ch_accounts_fetcher.sql
- **relations:** vet vertical pipeline (verified businesses input); scheduled_tasks entries per agent
- **verify-later:** ch_* actions; match-rate stats views; scheduled task cadence

<!-- SOURCE: U19_sql_tables_components.md -->
### Companies House enrichment with succession-risk signals
- **category:** companies-house-enrichment
- **status-signal:** deployed
- **status-evidence:** Schema + scheduled task (ch-enrichment every 20 min, seeded disabled "until Go actions are built") and a later applied accounts-fetch migration (accounts_fetched tracking, financial columns), indicating progression to live collection.
- **what:** Post-verification enrichment of business_intel.businesses: company identity/status/SIC, financials from filed accounts (accounts_type micro/small/medium/full, assets/net worth/turnover/PL, employees), officers and PSC JSONB, and derived owner-age/succession signals (owner_dob from CH month/year, estimated age, tenure, is_sole_director, is_corporate_owned → succession_risk high/medium/low/acquired). Deliberately polite rate limiting (~7% of CH's 600 req/5min). Match metadata records confidence/method/search query; accounts fetch is tracked separately on ch_vet_companies with an LLM-review exclusion filter.
- **sources:** docs/agent_docs/sql_for_tables/023_companies_house_data.sql
- **relations:** business-intel collection pipeline; http_request_log rate monitoring; vet vertical.
- **verify-later:** ch-enricher agent; enrichment coverage counts.

<!-- SOURCE: U24f_docs_archive_remaining_small.md -->
### Companies House matching cascade (revised 7-tier signal architecture)
- **category:** companies-house-enrichment
- **status-signal:** partial
- **status-evidence:** "Current matching achieves 676/2,767 (24.4%)... Target: 70-80% automated match rate, with HITL bringing it to 90%+." Presented as a revision (v2) of an earlier plan (docs/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/_archive/old/022_companies_houise_matching_cascade_plan.md, a v1 outside this unit's scope), and the enrichment domain itself has a live anchor doc (docs/agent_docs/docs024_key_docs_latest/017_companies_house_enrichment.md).
- **what:** A priority-ordered cascade replacing a flat two-pass matcher, where each business flows down tiers (matched / pass-to-next / queue-for-HITL) until resolved: Tier 0 scrapes the practice's own website for a company registration number (definitive, no CH API cost); Tier 1 exact-name+geography; Tier 2 exact-name unique-in-CH regardless of geography; Tier 3 postcode+moderate-name (raised threshold from 0.40→0.50 with a mandatory name-overlap component to cut false positives); Tier 4 LLM review with the top-3 trigram candidates and full business context (not just the single best match); Tier 5 corporate-group-parent mapping for chains sharing one CH registration (Medivet, CVS, Vets4Pets, IVC Evidensia, VetPartners — addressing ~800 corporate-branch businesses that a per-business match can never resolve 1:1); Tier 6 a human-review HITL queue for the remainder. New tables proposed: `businesses.company_number_scraped`, `ch_match_candidates`, `ch_corporate_groups`.
- **sources:** docs/_archive/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/_archive/archive_april_26/022b_companies_house_matching_cascade_plan_v2.md
- **relations:** companies-house-enrichment (017 anchor doc); the v1 predecessor plan (022_companies_houise_matching_cascade_plan.md)
- **verify-later:** business_intel.ch_match_candidates, business_intel.ch_corporate_groups, business_intel.businesses.company_number_scraped — confirm which tiers actually shipped and current match rate
