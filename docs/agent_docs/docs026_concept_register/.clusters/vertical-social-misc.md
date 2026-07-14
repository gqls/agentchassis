# Cluster: vertical-social-misc
Categories included: vet-med-pricing, companies-house-enrichment, canine-biology, new:business-intel-collection, social-media, vonc, new:site-chatbot, new:saas-isolation-architecture, new:email-infrastructure, admin-dashboard-and-api, research-agents, new:entity-data, new:deploy-mechanics-reference, new:public-api, adopting-and-scraping


<!-- SOURCE: U01_docs024_numbered_core.md -->
### Vet med pricing pipeline (discovery → scrape+evidence → export)
- **category:** vet-med-pricing
- **status-signal:** deployed
- **status-evidence:** 008 dated 2026-04-08 with per-retailer coverage stats; scheduled tasks configured (disabled pending verification)
- **what:** business-intel pod spawns Job pods per stage: URL discovery (category scraper or Firecrawl /map, deny-list filtered, upserted to med_retailer_listings) → price collection (Firecrawl scrape+screenshot, section truncation, retailer regex cascade → CPU Mistral fallback when £ present but 0 variants; snapshots + evidence rows; materialized view refresh) → JSON export (index/full/by-letter/metadata files, git commit → live). Multi-site via input_data filters (species/categories/retailers). Evidence chain: markdown + content hash + screenshot re-uploaded to B2, indefinitely.
- **sources:** 008 full
- **relations:** med-* wrapper orchestrators; LLM fallback as training data
- **verify-later:** business_intel.med_* tables; scheduled tasks enabled?

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Unified polymorphic products/product_prices schema (kind discriminator)
- **category:** vet-med-pricing
- **status-signal:** aspirational
- **status-evidence:** Migration file header: "SUPERSEDES the previously planned 001_business_med_prices_migration.sql, which would have created a third sibling table." Handoff doc: "Nothing applied to production yet."
- **what:** Unifies two previously-separate price tables (`business_prices` for services, `product_prices` for products) under the existing `business_intel.products` catalog by adding a `kind` discriminator column, migrating distinct `(service_category, service_name)` tuples into `products` rows with `kind='service'`, and backfilling matching price rows into `product_prices`. Does not drop `business_prices` (only marks it deprecated) until the paired Go write-path change ships; idempotent by design.
- **sources:** vetcomparison/006_unify_prices_schema.sql#header,#(A)(B)(C)(D), vetcomparison/HANDOFF_2026-05-18_vetcomparison_uk_planning.md#§2
- **relations:** business_prices deprecation migration pattern; vet-json-exporter / vet-export-orchestrator agent pair
- **verify-later:** business_intel.products.kind column + products_kind_check constraint

<!-- SOURCE: U13_docs024_small_dirs.md -->
### business_prices deprecation migration pattern
- **category:** vet-med-pricing
- **status-signal:** aspirational
- **status-evidence:** "Migration 006 idempotency depends on Go-B being deployed... The cleanest order is Go-B first, then 006 once."
- **what:** A phased table-retirement discipline: mark the old table deprecated via `COMMENT ON TABLE` rather than dropping it immediately; keep it functional until a paired Go code change stops writing to it; only drop it in a later, separate migration. Rollback notes tag migrated rows by `source = 'migrated_from_business_prices'` so a later rollback doesn't accidentally delete real rows.
- **sources:** vetcomparison/006_unify_prices_schema.sql#(D),#Rollback notes, vetcomparison/HANDOFF_2026-05-18_vetcomparison_uk_planning.md#Go-B section
- **relations:** Unified polymorphic products/product_prices schema
- **verify-later:** COMMENT ON TABLE business_intel.business_prices; Go insertPrice rewrite status

<!-- SOURCE: U13_docs024_small_dirs.md -->
### vetcomparison.uk V1 rebuild scope
- **category:** vet-med-pricing
- **status-signal:** aspirational
- **status-evidence:** "Status: Five artefacts written, two operations blocked on Go work, schema unification decided. Nothing applied to production yet."
- **what:** A deliberately narrow relaunch scope for `vetcomparison.uk` (replacing a broken prototype pointing at a domain the user doesn't own): online-only medicine search, a vet directory with filter/sort, a news feed, and two adopted guide pages — explicitly excluding a local-pharmacy comparison panel until per-vet medicine price data accumulates, planned as "V1.5."
- **sources:** vetcomparison/HANDOFF_2026-05-18_vetcomparison_uk_planning.md#TL;DR,#Phase 9
- **relations:** LLM-driven content_features recommendation; vet-json-exporter / vet-export-orchestrator agent pair
- **verify-later:** sites table row for domain vetcomparison.uk

<!-- SOURCE: U13_docs024_small_dirs.md -->
### LLM-driven content_features recommendation (news/tools/guides moved from Go to classifier prompt)
- **category:** vet-med-pricing
- **status-signal:** aspirational
- **status-evidence:** "Migration 005 does this... Apply ANY TIME. Self-contained, no paired Go change required" but still listed as unapplied
- **what:** Moves the decision of whether a site should get a news feed / tools / guides out of a hardcoded Go `verticalNewsMap` lookup and into the `domain-research-classifier` LLM prompt, which now emits a `content_features` block. A companion operation removes the now-redundant `enrich_news_feed` step, leaving `EvaluateNewsFeedAction`/`verticalNewsMap` as orphaned dead code.
- **sources:** vetcomparison/HANDOFF_2026-05-18_vetcomparison_uk_planning.md#§3,#Go-C
- **relations:** vetcomparison.uk V1 rebuild scope; Design/composition work-item emission gap (same class: reorg leaving orphaned old code paths)
- **verify-later:** domain-research-classifier prompt; EvaluateNewsFeedAction / verticalNewsMap deletion status

<!-- SOURCE: U13_docs024_small_dirs.md -->
### vet-json-exporter / vet-export-orchestrator agent pair (wrapper pattern)
- **category:** vet-med-pricing
- **status-signal:** aspirational
- **status-evidence:** "Go-A — vet_export_json action... None are written in this session"; agent_definitions rows are "safe to land" but "cannot do anything until Go-A ships"
- **what:** A new specialist/wrapper agent pair modelled directly on the existing `med-json-exporter`/`med-export-orchestrator` shape: `vet-json-exporter` (single `vet_export_json` action, reading confirmed vet-practice matches plus service prices via the new unified products/product_prices join) wrapped by `vet-export-orchestrator` (spawn → call → complete). Blocked on a still-owed DB query needed to build the consult/Rx/vaccination mapping switch.
- **sources:** vetcomparison/HANDOFF_2026-05-18_vetcomparison_uk_planning.md#Go-A,#Owed DB queries
- **relations:** Unified polymorphic products/product_prices schema; vetcomparison.uk V1 rebuild scope
- **verify-later:** platform/orchestration/actions/med_export_json_action.go (model to copy); registry.go "vet_export_json" registration (not yet present)

<!-- SOURCE: U18_sql_for_agents.md -->
### Vet vertical data pipeline (area sweep, batch processor, practice verifier, med pricing)
- **category:** vet-med-pricing
- **status-signal:** deployed
- **status-evidence:** Tuning migrations show live operation (063 "up the max iterations to clear the backlog" → 1700; 063b prompt refined to extract registration_number).
- **what:** The veterinary vertical's collection stack: area-sweep-orchestrator loads un-swept UK postcode districts and dispatches area-sweep-discoverer per district (web search → discovery candidates for unknown businesses); vet-batch-processor works candidate batches; vet-practice-verifier web-searches each business (postcode/town fallback query template) and LLM-extracts/reconciles structured practice data including Companies House number. Med pricing: med-url-discoverer scrapes retailer category pages for product URLs, med-url-mapper uses Firecrawl /map site-wide, med-price-collector scrapes prices; each has a thin spawn→call orchestrator wrapper and scheduled task.
- **sources:** 037_area_sweep_discoverer.sql; 038_area_sweep_orchestrator.sql; 063_vet_batch_processor.sql; 063b_vet_practice_verifier.sql; 092_vet_med_pricing_agent.sql; 095_vet_med_firecrawl_url_agent.sql; 096_vet_med_url_discover_orchestrator.sql
- **relations:** companies-house chain consumes verified businesses; business-intel pod/topics
- **verify-later:** search_areas / businesses tables; scheduled task states

<!-- SOURCE: U19_sql_tables_components.md -->
### Vet med pricing schema (products / retailers / listings / snapshots)
- **category:** vet-med-pricing
- **status-signal:** deployed
- **status-evidence:** Schema migration applied (028, duplicated in 030), retailer seeds with live corrections ("URL structure changed", "Domain is animed.co.uk not animeddirect.co.uk — updated from plan"), test seeds, manual listing matches.
- **what:** Four tables + matview in business_intel: med_products (canonical catalog: generic/brand/manufacturer/species[]/category/form/strength, prescription flag), med_retailers (4 UK pharmacies with group ownership — IVC Evidensia, CVS, Covetrus, Independent — category_urls for discovery, delivery costs, scrape_config hints), med_retailer_listings (retailer URL per product with match_confidence/match_method manual|llm|exact_name, NULL product until matched, denormalised last_price), med_price_snapshots (per size_variant price history incl. typical_vet_price), and med_price_current materialized view (latest price per listing/variant within 14 days) for export.
- **sources:** docs/agent_docs/sql_for_tables/028_vet_med_prices.sql; docs/agent_docs/sql_for_tables/029_vet_med_retailers.sql; docs/agent_docs/sql_for_tables/029b_vet_med_test_seed.sql; docs/agent_docs/sql_for_tables/033-business_intel.med_retailer_listings.sql
- **relations:** scrape evidence; spawn orchestrators; JSON export.
- **verify-later:** matview refresh cadence; listing match coverage.

<!-- SOURCE: U19_sql_tables_components.md -->
### Med scrape evidence store
- **category:** vet-med-pricing
- **status-signal:** deployed
- **status-evidence:** Table comment: "Raw scraped page content as evidence of prices. One row per page fetch. Retention: keep at least 90 days."
- **what:** Every price is traceable to the page it came from: one row per fetch with the Firecrawl markdown content, SHA256 content_hash for unchanged-page detection, variants_found vs prices_stored accounting, and response metadata — the audit trail for price provenance.
- **sources:** docs/agent_docs/sql_for_tables/032_business_intel_med_scrape_evidence
- **relations:** med pricing schema; vet-med-pricing evidence requirement (doc 008).
- **verify-later:** evidence rows per scrape run; retention enforcement.

<!-- SOURCE: U19_sql_tables_components.md -->
### Med URL discovery via Firecrawl /map
- **category:** vet-med-pricing
- **status-signal:** partial
- **status-evidence:** med-url-mapper seeded with status 'experimental' ("Particularly useful for VioVet where category-page scraping misses products"); registry.go entry supplied as a comment, i.e. Go side pending at write time.
- **what:** A second, broader product-URL discovery path using Firecrawl's /map endpoint site-wide, alongside category-page crawling; wrapped in the standard spawn orchestrator (med-url-map-orchestrator).
- **sources:** docs/agent_docs/sql_for_tables/035_vet_med_url_mapper_and_orchestrator.sql
- **relations:** spawn-orchestrator pattern; med pricing discovery.
- **verify-later:** med_map_urls in registry.go; experimental → active status.

<!-- SOURCE: U19_sql_tables_components.md -->
### Configurable med price JSON export to site repos
- **category:** vet-med-pricing
- **status-signal:** deployed
- **status-evidence:** med-json-exporter agent seeded 'active' with full config (domain, repo, data_path, outputs index/full/by_letter/metadata); scheduled task med-export-json seeded every 48h but enabled=false initially.
- **what:** One generic export action serves many consumer sites via config: query med_price_current, apply filters (species/category/retailers), build JSON artefacts, and commit them into the target site's git repo (e.g. vetcomparison.co.uk /data). The price data pipeline's publishing edge — sites consume static JSON, not the DB.
- **sources:** docs/agent_docs/sql_for_tables/037_vet_med_export_orchestrator_prices_json.sql
- **relations:** deployment-github (commit path); client-side JSON rendering pattern.
- **verify-later:** exports landing in site repos; task enablement.

<!-- SOURCE: U24b_docs_archive_finetuning.md -->
### thunder-reaper + cost gate (spend backstop)
- **category:** vet-med-pricing
- **status-signal:** deployed
- **status-evidence:** STATUS_thunder_adapter §1 "3.5 thunder-reaper … Deployed and verified end-to-end (2026-05-14)"; NOTES(39) "Reaper SAVE (done, verified)" bumped max_uptime_hours 18→48
- **what:** Two DB-driven cost controls. The `thunder-reaper` scheduled task (every 15 min) decommissions `running` instances older than their per-instance `max_uptime_hours` (default 18; the cap is ours, not Thunder's — computed from `running_since`, extendable per-row without a Thunder call). The provisioning cost gate is the `thunder_provision_check` view (checks 24h spend + estimated_new_run_cost vs `thunder_config.daily_cap_usd`, defaults cap $100 / per-run $25), called before every create; a ~9h/~$18 run needs the estimate/cap made realistic.
- **sources:** docubundle/.../STATUS_thunder_adapter_2026-06_04.md#1; flywheel_docs/FOCUS_finetuning_flywheel_and_service(21).md#14 (Spend gating); phase5/RUNBOOK_iter0_pretrigger(3).md#5
- **relations:** shared decommission_instance action with the monitor; distinct from the completion-monitor
- **verify-later:** thunder_provision_check view; thunder_config (daily_cap_usd, estimated_new_run_cost_usd, max_uptime_hours); migration 028

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Vet med pricing pipeline (discovery → scrape+evidence → export)
- **category:** vet-med-pricing
- **status-signal:** deployed
- **status-evidence:** 008 dated 2026-04-08 with per-retailer coverage stats; scheduled tasks configured (disabled pending verification)
- **what:** business-intel pod spawns Job pods per stage: URL discovery (category scraper or Firecrawl /map, deny-list filtered, upserted to med_retailer_listings) → price collection (Firecrawl scrape+screenshot, section truncation, retailer regex cascade → CPU Mistral fallback when £ present but 0 variants; snapshots + evidence rows; materialized view refresh) → JSON export (index/full/by-letter/metadata files, git commit → live). Multi-site via input_data filters (species/categories/retailers). Evidence chain: markdown + content hash + screenshot re-uploaded to B2, indefinitely.
- **sources:** 008 full
- **relations:** med-* wrapper orchestrators; LLM fallback as training data
- **verify-later:** business_intel.med_* tables; scheduled tasks enabled?

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Unified polymorphic products/product_prices schema (kind discriminator)
- **category:** vet-med-pricing
- **status-signal:** aspirational
- **status-evidence:** Migration file header: "SUPERSEDES the previously planned 001_business_med_prices_migration.sql, which would have created a third sibling table." Handoff doc: "Nothing applied to production yet."
- **what:** Unifies two previously-separate price tables (`business_prices` for services, `product_prices` for products) under the existing `business_intel.products` catalog by adding a `kind` discriminator column, migrating distinct `(service_category, service_name)` tuples into `products` rows with `kind='service'`, and backfilling matching price rows into `product_prices`. Does not drop `business_prices` (only marks it deprecated) until the paired Go write-path change ships; idempotent by design.
- **sources:** vetcomparison/006_unify_prices_schema.sql#header,#(A)(B)(C)(D), vetcomparison/HANDOFF_2026-05-18_vetcomparison_uk_planning.md#§2
- **relations:** business_prices deprecation migration pattern; vet-json-exporter / vet-export-orchestrator agent pair
- **verify-later:** business_intel.products.kind column + products_kind_check constraint

<!-- SOURCE: U13_docs024_small_dirs.md -->
### business_prices deprecation migration pattern
- **category:** vet-med-pricing
- **status-signal:** aspirational
- **status-evidence:** "Migration 006 idempotency depends on Go-B being deployed... The cleanest order is Go-B first, then 006 once."
- **what:** A phased table-retirement discipline: mark the old table deprecated via `COMMENT ON TABLE` rather than dropping it immediately; keep it functional until a paired Go code change stops writing to it; only drop it in a later, separate migration. Rollback notes tag migrated rows by `source = 'migrated_from_business_prices'` so a later rollback doesn't accidentally delete real rows.
- **sources:** vetcomparison/006_unify_prices_schema.sql#(D),#Rollback notes, vetcomparison/HANDOFF_2026-05-18_vetcomparison_uk_planning.md#Go-B section
- **relations:** Unified polymorphic products/product_prices schema
- **verify-later:** COMMENT ON TABLE business_intel.business_prices; Go insertPrice rewrite status

<!-- SOURCE: U13_docs024_small_dirs.md -->
### vetcomparison.uk V1 rebuild scope
- **category:** vet-med-pricing
- **status-signal:** aspirational
- **status-evidence:** "Status: Five artefacts written, two operations blocked on Go work, schema unification decided. Nothing applied to production yet."
- **what:** A deliberately narrow relaunch scope for `vetcomparison.uk` (replacing a broken prototype pointing at a domain the user doesn't own): online-only medicine search, a vet directory with filter/sort, a news feed, and two adopted guide pages — explicitly excluding a local-pharmacy comparison panel until per-vet medicine price data accumulates, planned as "V1.5."
- **sources:** vetcomparison/HANDOFF_2026-05-18_vetcomparison_uk_planning.md#TL;DR,#Phase 9
- **relations:** LLM-driven content_features recommendation; vet-json-exporter / vet-export-orchestrator agent pair
- **verify-later:** sites table row for domain vetcomparison.uk

<!-- SOURCE: U13_docs024_small_dirs.md -->
### LLM-driven content_features recommendation (news/tools/guides moved from Go to classifier prompt)
- **category:** vet-med-pricing
- **status-signal:** aspirational
- **status-evidence:** "Migration 005 does this... Apply ANY TIME. Self-contained, no paired Go change required" but still listed as unapplied
- **what:** Moves the decision of whether a site should get a news feed / tools / guides out of a hardcoded Go `verticalNewsMap` lookup and into the `domain-research-classifier` LLM prompt, which now emits a `content_features` block. A companion operation removes the now-redundant `enrich_news_feed` step, leaving `EvaluateNewsFeedAction`/`verticalNewsMap` as orphaned dead code.
- **sources:** vetcomparison/HANDOFF_2026-05-18_vetcomparison_uk_planning.md#§3,#Go-C
- **relations:** vetcomparison.uk V1 rebuild scope; Design/composition work-item emission gap (same class: reorg leaving orphaned old code paths)
- **verify-later:** domain-research-classifier prompt; EvaluateNewsFeedAction / verticalNewsMap deletion status

<!-- SOURCE: U13_docs024_small_dirs.md -->
### vet-json-exporter / vet-export-orchestrator agent pair (wrapper pattern)
- **category:** vet-med-pricing
- **status-signal:** aspirational
- **status-evidence:** "Go-A — vet_export_json action... None are written in this session"; agent_definitions rows are "safe to land" but "cannot do anything until Go-A ships"
- **what:** A new specialist/wrapper agent pair modelled directly on the existing `med-json-exporter`/`med-export-orchestrator` shape: `vet-json-exporter` (single `vet_export_json` action, reading confirmed vet-practice matches plus service prices via the new unified products/product_prices join) wrapped by `vet-export-orchestrator` (spawn → call → complete). Blocked on a still-owed DB query needed to build the consult/Rx/vaccination mapping switch.
- **sources:** vetcomparison/HANDOFF_2026-05-18_vetcomparison_uk_planning.md#Go-A,#Owed DB queries
- **relations:** Unified polymorphic products/product_prices schema; vetcomparison.uk V1 rebuild scope
- **verify-later:** platform/orchestration/actions/med_export_json_action.go (model to copy); registry.go "vet_export_json" registration (not yet present)

<!-- SOURCE: U18_sql_for_agents.md -->
### Vet vertical data pipeline (area sweep, batch processor, practice verifier, med pricing)
- **category:** vet-med-pricing
- **status-signal:** deployed
- **status-evidence:** Tuning migrations show live operation (063 "up the max iterations to clear the backlog" → 1700; 063b prompt refined to extract registration_number).
- **what:** The veterinary vertical's collection stack: area-sweep-orchestrator loads un-swept UK postcode districts and dispatches area-sweep-discoverer per district (web search → discovery candidates for unknown businesses); vet-batch-processor works candidate batches; vet-practice-verifier web-searches each business (postcode/town fallback query template) and LLM-extracts/reconciles structured practice data including Companies House number. Med pricing: med-url-discoverer scrapes retailer category pages for product URLs, med-url-mapper uses Firecrawl /map site-wide, med-price-collector scrapes prices; each has a thin spawn→call orchestrator wrapper and scheduled task.
- **sources:** 037_area_sweep_discoverer.sql; 038_area_sweep_orchestrator.sql; 063_vet_batch_processor.sql; 063b_vet_practice_verifier.sql; 092_vet_med_pricing_agent.sql; 095_vet_med_firecrawl_url_agent.sql; 096_vet_med_url_discover_orchestrator.sql
- **relations:** companies-house chain consumes verified businesses; business-intel pod/topics
- **verify-later:** search_areas / businesses tables; scheduled task states

<!-- SOURCE: U19_sql_tables_components.md -->
### Vet med pricing schema (products / retailers / listings / snapshots)
- **category:** vet-med-pricing
- **status-signal:** deployed
- **status-evidence:** Schema migration applied (028, duplicated in 030), retailer seeds with live corrections ("URL structure changed", "Domain is animed.co.uk not animeddirect.co.uk — updated from plan"), test seeds, manual listing matches.
- **what:** Four tables + matview in business_intel: med_products (canonical catalog: generic/brand/manufacturer/species[]/category/form/strength, prescription flag), med_retailers (4 UK pharmacies with group ownership — IVC Evidensia, CVS, Covetrus, Independent — category_urls for discovery, delivery costs, scrape_config hints), med_retailer_listings (retailer URL per product with match_confidence/match_method manual|llm|exact_name, NULL product until matched, denormalised last_price), med_price_snapshots (per size_variant price history incl. typical_vet_price), and med_price_current materialized view (latest price per listing/variant within 14 days) for export.
- **sources:** docs/agent_docs/sql_for_tables/028_vet_med_prices.sql; docs/agent_docs/sql_for_tables/029_vet_med_retailers.sql; docs/agent_docs/sql_for_tables/029b_vet_med_test_seed.sql; docs/agent_docs/sql_for_tables/033-business_intel.med_retailer_listings.sql
- **relations:** scrape evidence; spawn orchestrators; JSON export.
- **verify-later:** matview refresh cadence; listing match coverage.

<!-- SOURCE: U19_sql_tables_components.md -->
### Med scrape evidence store
- **category:** vet-med-pricing
- **status-signal:** deployed
- **status-evidence:** Table comment: "Raw scraped page content as evidence of prices. One row per page fetch. Retention: keep at least 90 days."
- **what:** Every price is traceable to the page it came from: one row per fetch with the Firecrawl markdown content, SHA256 content_hash for unchanged-page detection, variants_found vs prices_stored accounting, and response metadata — the audit trail for price provenance.
- **sources:** docs/agent_docs/sql_for_tables/032_business_intel_med_scrape_evidence
- **relations:** med pricing schema; vet-med-pricing evidence requirement (doc 008).
- **verify-later:** evidence rows per scrape run; retention enforcement.

<!-- SOURCE: U19_sql_tables_components.md -->
### Med URL discovery via Firecrawl /map
- **category:** vet-med-pricing
- **status-signal:** partial
- **status-evidence:** med-url-mapper seeded with status 'experimental' ("Particularly useful for VioVet where category-page scraping misses products"); registry.go entry supplied as a comment, i.e. Go side pending at write time.
- **what:** A second, broader product-URL discovery path using Firecrawl's /map endpoint site-wide, alongside category-page crawling; wrapped in the standard spawn orchestrator (med-url-map-orchestrator).
- **sources:** docs/agent_docs/sql_for_tables/035_vet_med_url_mapper_and_orchestrator.sql
- **relations:** spawn-orchestrator pattern; med pricing discovery.
- **verify-later:** med_map_urls in registry.go; experimental → active status.

<!-- SOURCE: U19_sql_tables_components.md -->
### Configurable med price JSON export to site repos
- **category:** vet-med-pricing
- **status-signal:** deployed
- **status-evidence:** med-json-exporter agent seeded 'active' with full config (domain, repo, data_path, outputs index/full/by_letter/metadata); scheduled task med-export-json seeded every 48h but enabled=false initially.
- **what:** One generic export action serves many consumer sites via config: query med_price_current, apply filters (species/category/retailers), build JSON artefacts, and commit them into the target site's git repo (e.g. vetcomparison.co.uk /data). The price data pipeline's publishing edge — sites consume static JSON, not the DB.
- **sources:** docs/agent_docs/sql_for_tables/037_vet_med_export_orchestrator_prices_json.sql
- **relations:** deployment-github (commit path); client-side JSON rendering pattern.
- **verify-later:** exports landing in site repos; task enablement.

<!-- SOURCE: U24b_docs_archive_finetuning.md -->
### thunder-reaper + cost gate (spend backstop)
- **category:** vet-med-pricing
- **status-signal:** deployed
- **status-evidence:** STATUS_thunder_adapter §1 "3.5 thunder-reaper … Deployed and verified end-to-end (2026-05-14)"; NOTES(39) "Reaper SAVE (done, verified)" bumped max_uptime_hours 18→48
- **what:** Two DB-driven cost controls. The `thunder-reaper` scheduled task (every 15 min) decommissions `running` instances older than their per-instance `max_uptime_hours` (default 18; the cap is ours, not Thunder's — computed from `running_since`, extendable per-row without a Thunder call). The provisioning cost gate is the `thunder_provision_check` view (checks 24h spend + estimated_new_run_cost vs `thunder_config.daily_cap_usd`, defaults cap $100 / per-run $25), called before every create; a ~9h/~$18 run needs the estimate/cap made realistic.
- **sources:** docubundle/.../STATUS_thunder_adapter_2026-06_04.md#1; flywheel_docs/FOCUS_finetuning_flywheel_and_service(21).md#14 (Spend gating); phase5/RUNBOOK_iter0_pretrigger(3).md#5
- **relations:** shared decommission_instance action with the monitor; distinct from the completion-monitor
- **verify-later:** thunder_provision_check view; thunder_config (daily_cap_usd, estimated_new_run_cost_usd, max_uptime_hours); migration 028

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

<!-- SOURCE: U21_legacy_docs_b.md -->
### Canine biology knowledge tree (1M-agent demo)
- **category:** canine-biology
- **status-signal:** aspirational
- **status-evidence:** docs016/003c dated 2026-03-02, "Status: Working draft for further iteration"; docs016/004 (2026-03-03) demotes it: "best treated as marketing spend, not as a product... Build one branch (cardiovascular) as a polished showcase."
- **what:** A hierarchical agent swarm building a citable Labrador-reference knowledge tree: 7 levels of decomposition (root → body systems → aspects → subtopics → specific topics → mechanisms → molecular detail, branching 8–12), ~800K–1M agents across nine roles (Opus decomposers/synthesisers at top levels; BioMistral 7B research and finding-synthesis; non-LLM paper fetchers hitting PubMed; SciSpacy NER entity extractors; embedded-3B relevance filters; mermaid/FLUX diagram agents; 7B validators flagging cross-branch contradictions). Design priorities: accuracy over completeness; no reader-visible text from 3B models; phased rollout (125K live agents on five priority branches, background fill, then continuous PubMed-monitoring updates ~500-1000 agents/week); every node auditable (agent, prompt, sources, model); correction/discussion layer with versioning; pathway/mechanism cross-layer. Honest-risk section: credibility vs Plumb's/Merck, theatrical agent count, hallucination persistence, front-end decisive, costs 2-3x estimates ($2.2K-8.5K full run).
- **sources:** docs016_dogs_medicine_pathways/003c_canine_biology_project_baseline_v3.md; docs016_dogs_medicine_pathways/002_project_outline.md; docs016_dogs_medicine_pathways/004_medical_business_reality_assessment.md
- **relations:** canine-biology category (docs018 feature plans); multicluster worker pools; model-tiering; business strategy demotion.
- **verify-later:** any decomposer/leaf agent definitions; knowledge tree tables (expected absent).

<!-- SOURCE: U22_recent_small_docs.md -->
### Canine biology knowledge base (veterinary seeding)
- **category:** canine-biology
- **status-signal:** aspirational
- **status-evidence:** "The canine biology project stops being aspirational and becomes the working proof..." — future tense; "knowledge base is empty" in the RAG explainer.
- **what:** The first real RAG content and proof-of-concept for the veterinary vertical: structured LLM extraction (breed health profiles for top 20 UK breeds, 30-40 procedures, top 30 conditions, nutrition/vaccination/behaviour) into ~300-500 self-contained 200-500-word chunks, validated (self-consistency, cross-reference, structural), embedded via Ollama, and indexed into `collection: "veterinary"`. Structured JSON with confidence markers, not prose.
- **sources:** docs023.../018_canine_biology.md, docs023.../001_canine_biology_grok_plan.md
- **relations:** RAG knowledge_base, vertical knowledge architecture, text LoRA (vet extractor), deep research domain authority
- **verify-later:** knowledge_base rows collection='veterinary'; knowledge-extractor agent

<!-- SOURCE: U22_recent_small_docs.md -->
### Interactive Biological Explorer + experiment engine (aspirational vision)
- **category:** canine-biology
- **status-signal:** abandoned
- **status-evidence:** The grandiose Grok "Final Consolidated Plan" (multi-scale explorer, knowledge graph, experiment engine, 14-week timeline) is explicitly downgraded in the later doc: "The original 1M-agent design was aspirational. This plan is practical."
- **what:** An early, much larger vision: a public Next.js/Three.js/Cytoscape web app allowing drill-down from a pseudo-photographic Labrador image → organ systems → cells → biochemical pathways → genes, backed by a PostgreSQL/Neo4j knowledge graph (Gene/Protein/Metabolite/Reaction/Organ nodes), plus an agent-driven "theoretical experiment engine" running SciPy ODE simulations. Superseded by the practical RAG-seeding plan; the explorer/graph/experiment layers were dropped.
- **sources:** docs023.../001_canine_biology_grok_plan.md, docs023.../018_canine_biology.md#1
- **relations:** canine biology knowledge base (the practical replacement), image LoRA
- **verify-later:** n/a (not built; abandoned scope)

<!-- SOURCE: U24a_docs_archive_classic_and_docs024_misc.md -->
### Canine-biology per-vertical knowledge + LoRA project
- **category:** canine-biology
- **status-signal:** partial
- **status-evidence:** 018_canine_biology (identical to live docs023): §5 RAG in content generation "Authoritative Domain Knowledge"; §6 Text LoRA (4-bit QLoRA→GGUF Q4_K_M→Ollama), §7 Image LoRA (SDXL, ~£35-95 first pass); FOCUS_imagery_assessment §8 "Status: planned, not started."
- **what:** The reference per-vertical knowledge/fine-tuning project: research → chunk+embed into the RAG knowledge base → text LoRA fine-tune (Unsloth QLoRA, deployed via Ollama Modelfile) → image LoRA (60-90 curated images, SDXL/PixArt) for consistent per-vertical visual style. The image LoRA presupposes an adapter that accepts a `model:"vet-diagram-v1"` field — which the current Stability-only adapter cannot, so it is blocked on the provider-router work.
- **sources:** 018_canine_biology.md#6-text-lora-fine-tuning, #7-image-lora-fine-tuning; FOCUS_imagery_assessment(1).md#8-per-vertical-training-infrastructure
- **relations:** quality flywheel; RAG best practices; image provider router; vision auditor
- **verify-later:** knowledge_base pgvector rows; training_runs; Ollama LoRA Modelfiles

<!-- SOURCE: U21_legacy_docs_b.md -->
### Canine biology knowledge tree (1M-agent demo)
- **category:** canine-biology
- **status-signal:** aspirational
- **status-evidence:** docs016/003c dated 2026-03-02, "Status: Working draft for further iteration"; docs016/004 (2026-03-03) demotes it: "best treated as marketing spend, not as a product... Build one branch (cardiovascular) as a polished showcase."
- **what:** A hierarchical agent swarm building a citable Labrador-reference knowledge tree: 7 levels of decomposition (root → body systems → aspects → subtopics → specific topics → mechanisms → molecular detail, branching 8–12), ~800K–1M agents across nine roles (Opus decomposers/synthesisers at top levels; BioMistral 7B research and finding-synthesis; non-LLM paper fetchers hitting PubMed; SciSpacy NER entity extractors; embedded-3B relevance filters; mermaid/FLUX diagram agents; 7B validators flagging cross-branch contradictions). Design priorities: accuracy over completeness; no reader-visible text from 3B models; phased rollout (125K live agents on five priority branches, background fill, then continuous PubMed-monitoring updates ~500-1000 agents/week); every node auditable (agent, prompt, sources, model); correction/discussion layer with versioning; pathway/mechanism cross-layer. Honest-risk section: credibility vs Plumb's/Merck, theatrical agent count, hallucination persistence, front-end decisive, costs 2-3x estimates ($2.2K-8.5K full run).
- **sources:** docs016_dogs_medicine_pathways/003c_canine_biology_project_baseline_v3.md; docs016_dogs_medicine_pathways/002_project_outline.md; docs016_dogs_medicine_pathways/004_medical_business_reality_assessment.md
- **relations:** canine-biology category (docs018 feature plans); multicluster worker pools; model-tiering; business strategy demotion.
- **verify-later:** any decomposer/leaf agent definitions; knowledge tree tables (expected absent).

<!-- SOURCE: U22_recent_small_docs.md -->
### Canine biology knowledge base (veterinary seeding)
- **category:** canine-biology
- **status-signal:** aspirational
- **status-evidence:** "The canine biology project stops being aspirational and becomes the working proof..." — future tense; "knowledge base is empty" in the RAG explainer.
- **what:** The first real RAG content and proof-of-concept for the veterinary vertical: structured LLM extraction (breed health profiles for top 20 UK breeds, 30-40 procedures, top 30 conditions, nutrition/vaccination/behaviour) into ~300-500 self-contained 200-500-word chunks, validated (self-consistency, cross-reference, structural), embedded via Ollama, and indexed into `collection: "veterinary"`. Structured JSON with confidence markers, not prose.
- **sources:** docs023.../018_canine_biology.md, docs023.../001_canine_biology_grok_plan.md
- **relations:** RAG knowledge_base, vertical knowledge architecture, text LoRA (vet extractor), deep research domain authority
- **verify-later:** knowledge_base rows collection='veterinary'; knowledge-extractor agent

<!-- SOURCE: U22_recent_small_docs.md -->
### Interactive Biological Explorer + experiment engine (aspirational vision)
- **category:** canine-biology
- **status-signal:** abandoned
- **status-evidence:** The grandiose Grok "Final Consolidated Plan" (multi-scale explorer, knowledge graph, experiment engine, 14-week timeline) is explicitly downgraded in the later doc: "The original 1M-agent design was aspirational. This plan is practical."
- **what:** An early, much larger vision: a public Next.js/Three.js/Cytoscape web app allowing drill-down from a pseudo-photographic Labrador image → organ systems → cells → biochemical pathways → genes, backed by a PostgreSQL/Neo4j knowledge graph (Gene/Protein/Metabolite/Reaction/Organ nodes), plus an agent-driven "theoretical experiment engine" running SciPy ODE simulations. Superseded by the practical RAG-seeding plan; the explorer/graph/experiment layers were dropped.
- **sources:** docs023.../001_canine_biology_grok_plan.md, docs023.../018_canine_biology.md#1
- **relations:** canine biology knowledge base (the practical replacement), image LoRA
- **verify-later:** n/a (not built; abandoned scope)

<!-- SOURCE: U24a_docs_archive_classic_and_docs024_misc.md -->
### Canine-biology per-vertical knowledge + LoRA project
- **category:** canine-biology
- **status-signal:** partial
- **status-evidence:** 018_canine_biology (identical to live docs023): §5 RAG in content generation "Authoritative Domain Knowledge"; §6 Text LoRA (4-bit QLoRA→GGUF Q4_K_M→Ollama), §7 Image LoRA (SDXL, ~£35-95 first pass); FOCUS_imagery_assessment §8 "Status: planned, not started."
- **what:** The reference per-vertical knowledge/fine-tuning project: research → chunk+embed into the RAG knowledge base → text LoRA fine-tune (Unsloth QLoRA, deployed via Ollama Modelfile) → image LoRA (60-90 curated images, SDXL/PixArt) for consistent per-vertical visual style. The image LoRA presupposes an adapter that accepts a `model:"vet-diagram-v1"` field — which the current Stability-only adapter cannot, so it is blocked on the provider-router work.
- **sources:** 018_canine_biology.md#6-text-lora-fine-tuning, #7-image-lora-fine-tuning; FOCUS_imagery_assessment(1).md#8-per-vertical-training-infrastructure
- **relations:** quality flywheel; RAG best practices; image provider router; vision auditor
- **verify-later:** knowledge_base pgvector rows; training_runs; Ollama LoRA Modelfiles

<!-- SOURCE: U19_sql_tables_components.md -->
### Business-intel sweep/verify collection pipeline (vet-intel)
- **category:** NEW:business-intel-collection
- **status-signal:** deployed
- **status-evidence:** Operational scheduled tasks: vet-batch-verify (claims pending collection_tasks), vet-task-reset → broadened vet-cleanup self-healer (fails orchestrations stuck AWAITING_RESPONSES >20 min, resets stuck collection_tasks, "breaks the stall chain"), vet-sweep-continue (batches of 200 unswept areas); later re-pointed at a dedicated vet-intel pod on system.agent.vet-intel.requests.
- **what:** The area-sweep → collection_tasks → batch-verify pipeline that builds the verified business directory (vertical: veterinary) which CH enrichment then deepens. Includes the operational self-healing pattern and the dedicated-pod routing decision (vet-intel instead of the generic agent).
- **sources:** docs/agent_docs/sql_for_tables/020_scheduled_tasks.sql#vet-tasks and #vet-cleanup and #vet-intel-setup; docs/agent_docs/sql_for_tables/023_companies_house_data.sql#pre-query
- **relations:** companies-house enrichment; batch-processing; scheduler self-healing.
- **verify-later:** business_intel.businesses / collection_tasks schemas (defined elsewhere); vet-intel agent definition.

<!-- SOURCE: U19_sql_tables_components.md -->
### Business-intel sweep/verify collection pipeline (vet-intel)
- **category:** NEW:business-intel-collection
- **status-signal:** deployed
- **status-evidence:** Operational scheduled tasks: vet-batch-verify (claims pending collection_tasks), vet-task-reset → broadened vet-cleanup self-healer (fails orchestrations stuck AWAITING_RESPONSES >20 min, resets stuck collection_tasks, "breaks the stall chain"), vet-sweep-continue (batches of 200 unswept areas); later re-pointed at a dedicated vet-intel pod on system.agent.vet-intel.requests.
- **what:** The area-sweep → collection_tasks → batch-verify pipeline that builds the verified business directory (vertical: veterinary) which CH enrichment then deepens. Includes the operational self-healing pattern and the dedicated-pod routing decision (vet-intel instead of the generic agent).
- **sources:** docs/agent_docs/sql_for_tables/020_scheduled_tasks.sql#vet-tasks and #vet-cleanup and #vet-intel-setup; docs/agent_docs/sql_for_tables/023_companies_house_data.sql#pre-query
- **relations:** companies-house enrichment; batch-processing; scheduler self-healing.
- **verify-later:** business_intel.businesses / collection_tasks schemas (defined elsewhere); vet-intel agent definition.

<!-- SOURCE: U25_leopardess_social.md -->
### The Forge — AI-seeded community knowledge platform
- **category:** social-media
- **status-signal:** abandoned
- **status-evidence:** 001 doc "Status: Concept stage. Parked for future development." (file dated Mar 2026; no later reference builds on it).
- **what:** Product concept predating Spark: AI answers are published as explicit first drafts into a categorised community feed; humans validate/challenge/add experience/fork; the AI synthesises human input into improved versions, visibly showing evolution. Key dynamics: correcting the AI is rewarding not adversarial; "Beat the AI" competing answers earn domain reputation; AI as fair debate synthesiser; vertical Forge embeds (e.g. vets refining pet-owner answers). Open questions recorded: moderation at scale, cold start, AI identity, expert monetisation.
- **sources:** docs/social001_vonc_tiktok_social/001_concept_the_forge_humans_edit_ai_responses.md; docs/social001_vonc_tiktok_social/002pre_whole_chat (origin transcript)
- **relations:** Spark (successor concept, keeps AI-as-participant-not-oracle DNA); hitl
- **verify-later:** none (never built)

<!-- SOURCE: U25_leopardess_social.md -->
### Spark — AI game-master social platform (core concept)
- **category:** social-media
- **status-signal:** partial
- **status-evidence:** 002e "Status: Active exploration. All mechanics candidates for live testing"; the v1 content-first site is live on vonc.com (minilobby docs, 2026-07).
- **what:** "AI-driven provocation engine where the world is the content and your take is the game." AI occupies the game-master/producer role — framing, structure, scoring, recaps, synthesis, hosting, matchmaking — and is deliberately absent from responses, humour, personality, takes ("producer, not performer"; when something is funny, a human did it). Differentiators: opinion-first entry ("what's your take?" beats create-from-scratch), ephemeral game-structured challenges, rooms not feeds, showcase+competitive dual mode. Includes the AI-slop strategy (restrict AI to strengths; deliberately-bad AI takes as straight-man seeds) and moderation-by-design (AI referee/yellow cards, positive selection, provocation design as tone control, ephemerality limits damage, "Cursed" as community signal). Second-screen positioning: "TikTok: watch. Twitter: shout. Spark: play."
- **sources:** docs/social001_vonc_tiktok_social/002e_concept_spark(6).md#Positioning, #AI-Role, #The-AI-Slop-Problem, #Moderation; docs/social001_vonc_tiktok_social/trigger_script/004_submit_vonc_trigger.sh (mission brief)
- **relations:** The Forge (predecessor); vonc.com v1 site; provocation engine; archetype system
- **verify-later:** site_specs aspects mission/roadmap for vonc site 9ec3b9ee

<!-- SOURCE: U25_leopardess_social.md -->
### Arena + Stage dual modes and their mechanic families
- **category:** social-media
- **status-signal:** aspirational
- **status-evidence:** 003d roadmap: v3 "Live challenge rooms. Arena mode. Chains. Duels." status "speculative"; v4 "Stage mode … speculative".
- **what:** Two complementary energies that feed each other: Arena (competitive — provocations, reaction vocabulary Genius/Delusional/Suspicious/Based/Cursed, remix chains with shared credit, 60-second duels with AI referee, Misfit Mashup) and Stage (showcase — springboard challenges, Fire/Inspired/Stealing-This/Teach-Me/Vibe reactions, quality-curated Tune In follow, niche discovery rooms, Glow-Up progress reels, Teach-Me-triggered mentorship, AI-matchmade collabs, Rising newcomer spotlight, Taste Graph curator prestige). Flywheel: Stage showcases become Arena provocations; Arena winners get Stage suggestions; over time the platform generates its own culture.
- **sources:** docs/social001_vonc_tiktok_social/002e_concept_spark(6).md#Two-Modes, #Arena-Mechanics, #Stage-Mechanics; 002b (first appearance of the Arena/Stage split)
- **relations:** Spark core; rooms-not-feeds; emergent profiles
- **verify-later:** none built (v3/v4)

<!-- SOURCE: U25_leopardess_social.md -->
### Rooms-not-feeds architecture and the engagement-depth spectrum
- **category:** social-media
- **status-signal:** aspirational
- **status-evidence:** 002e: room mechanics and The Drift are design prose; v1 ships only static pages (003d "v1_content_first").
- **what:** Structural anti-feed design: a Lobby of 3–5 live rooms with energy indicators; room zones (Floor active / Gallery spectating with zero barrier); crowd-energy feedback (cards heat, reactions ripple, tug-of-war splits); ephemeral challenges with lasting recaps; Director's Cut AI replays; serialised prediction challenges; multi-format rotation; "The Crowd Speaks" synthesis; Moments (remarkable outcomes elevated to permanent shareable record). The Drift is the passive snackable stream; each depth level (Drift → Lobby → Gallery → Floor → solo modes) is complete in itself. Sound/haptics design included.
- **sources:** docs/social001_vonc_tiktok_social/002e_concept_spark(6).md#Shared-Mechanics, #The-Drift, #Sound-and-Haptics
- **relations:** Arena/Stage; lobby-grid component (v1's static echo of the lobby)
- **verify-later:** none built

<!-- SOURCE: U25_leopardess_social.md -->
### Behavioural archetype system + Daily Gauntlet
- **category:** social-media
- **status-signal:** partial
- **status-evidence:** 003d gauntlet page spec with 8 archetypes; RUNBOOK_minilobby §0: gauntlet tool + archetype-taster-quiz deployed, archetype hub (8 entity pages) live 2026-07-12.
- **what:** Identity engine: emergent profiles built from behaviour (can't be faked), producing archetypes (Surgeon, Wildcard, Oracle, Catalyst, Judge, Maker, Scout, Mentor) with secondary tendencies, earned via the Daily Gauntlet (5 provocations, 5 minutes, scored on speed/originality/consistency/topic preference) and shareable as visual cards — the "viral Trojan horse" (BuzzFeed/MBTI/Hogwarts dynamics on demonstrated behaviour, works with zero community). Radical anonymity in Arena (no usernames during play, reveal after, reputation never boosts visibility) and rotating achievable status games instead of permanent karma. On vonc v1 this exists as client-side tools + a canon archetype content set (088/089 note live archetype-combinations copy had drifted off-canon).
- **sources:** docs/social001_vonc_tiktok_social/003d_spark_strategic_planning_architecture.md#gauntlet content_context; docs/social001_vonc_tiktok_social/002e_concept_spark(6).md#The-archetype-as-viral-Trojan-horse, #Identity-and-Profiles, #Radical-Anonymity; docs/social001_vonc_tiktok_social/minilobby_task/088_archetype_entity_pages.sql (header)
- **relations:** archetype hub build; content-first launch; Daily Gauntlet page
- **verify-later:** /tools/gauntlet/index.html and /tools/archetype-taster-quiz/ on vonc.com

<!-- SOURCE: U25_leopardess_social.md -->
### Cold-start design: AI sparring partner and solo-first completeness
- **category:** social-media
- **status-signal:** aspirational
- **status-evidence:** 003d: "v2_sparring_and_interaction … status: directional, depends_on backend API infrastructure".
- **what:** The empty-room solution: first 10 seconds are a provocation card + text input + timer with no signup ("Engage first. Understand second. Commit third."); the AI is a transparent sparring partner (counter-arguments, not chatbot small-talk) so the experience is complete for one person; the scraping+AI flywheel self-fills content; invites are challenges not invitations; explicit scale thresholds (1 / 5–20 / 50–200 / 500+) each designed to feel complete. Sparring is also the top per-user cost risk with named mitigations (rate limits, smallest viable model, cached counters).
- **sources:** docs/social001_vonc_tiktok_social/002e_concept_spark(6).md#Cold-Start-Design, #The-First-Visit; 002 (original fuller treatment: Spark Solo, landing page day 1)
- **relations:** AI cost architecture; content-first launch
- **verify-later:** none built (v2)

<!-- SOURCE: U25_leopardess_social.md -->
### Provocation engine — layered content production architecture
- **category:** social-media
- **status-signal:** aspirational
- **status-evidence:** 002e technical-layers section is design; NOTES_provocations-archive-list (2026-07-09): "the archive feed is hand-committed until the Phase-3 pipeline emits provocations.json".
- **what:** Six-layer production line: (1) Raw Feed — scrapers pull social-interest trends (not newspaper news); (2) Framing Engine — cheap local models (Mistral/Llama 8B CPU) generate 5–10 provocation candidates per item, ~2,000/day; (3) Curation Gate — stronger model or human picks the best 15–20, learns from engagement; (4) Mashup Engine — foundation models find non-obvious connections, 2–5 calls/day; (5) Serialisation Tracker — narrative threads, mostly database; (6) Niche Detector — embeddings/ML clustering for Stage rooms. Content tone contract: interesting not dark, slightly adult never gruesome, competitive without money betting.
- **sources:** docs/social001_vonc_tiktok_social/002e_concept_spark(6).md#Provocation-Engine, #Content-Tone; docs/social001_vonc_tiktok_social/002_concept_spark.md#The-Provocation-Engine (original layer detail)
- **relations:** Phase-3 provocations.json pipeline (its v1 delivery vehicle); AI cost architecture; news-feed-pipeline (sibling scraping infrastructure)
- **verify-later:** any provocation-orchestrator agent definition; scheduled_tasks for provocation refresh

<!-- SOURCE: U25_leopardess_social.md -->
### AI cost architecture: fixed background vs per-user scaling
- **category:** social-media
- **status-signal:** aspirational
- **status-evidence:** 002d/002e cost tables are design projections ("Scaling Scenarios … 100,000 DAU"); nothing metered in production.
- **what:** Cost-shaping principle: most AI cost is fixed (content production scales with ~15–20 challenges/day, not users) and runs on cheap local models (~£5/day background at scale); the only linearly-scaling cost is per-user interactive AI (sparring, Gauntlet scoring) and is throttled by design. Projected £0.003–0.008/user/day, so a £3–5/month subscription covers compute several times over; break-even on subscriptions before ads/brand revenue.
- **sources:** docs/social001_vonc_tiktok_social/002e_concept_spark(6).md#AI-Cost-Architecture; 002d (first appearance of the full cost tables)
- **relations:** provocation engine; cold-start sparring; revenue model
- **verify-later:** none (projection)

<!-- SOURCE: U25_leopardess_social.md -->
### Content-first launch strategy for Spark (vonc.com as destination)
- **category:** social-media
- **status-signal:** partial
- **status-evidence:** 003d "current_phase: v1_content_first … Static S3 site. Provocations as SEO content. Daily Gauntlet as viral archetype quiz"; the site is live per minilobby docs.
- **what:** Don't launch a social platform; launch a content destination that happens to have interactive features. Daily provocations with the AI's take are SEO pages with shareable URLs; every provocation/response generates a self-contained share card ("47 people responded. Think you'd do better?" — the TikTok growth pattern for text); the archetype quiz works with zero community; vertical clustering happens organically by arrival path (the Reddit model — nobody joins "Reddit"). First-week content calendar defined (daily provocations, Gauntlet, weird-stat micro-content; weekly mashup + prediction post). "The first visit experience IS the pitch."
- **sources:** docs/social001_vonc_tiktok_social/002e_concept_spark(6).md#Launch-Strategy; docs/social001_vonc_tiktok_social/003d_spark_strategic_planning_architecture.md#The-initial-request
- **relations:** Spark core; vonc.com v1 site; provocation engine
- **verify-later:** vonc.com live pages; provocations archive content

<!-- SOURCE: U25_leopardess_social.md -->
### Motivation hierarchy and designed user journey
- **category:** social-media
- **status-signal:** aspirational
- **status-evidence:** 002c introduced the motivation tiers; 002d the journey; both remain design prose in 002e.
- **what:** Retention design: four motivation tiers (identity/status; belonging/connection; growth/learning; financial reward — "financial reward follows demonstrated value, nobody buys prominence") mapped to user types (casual → professional creator; acquire at Tier 2). A staged journey — first 5 seconds to month 6+ — engineered as intrigue → creation → validation → habit → identity → community → mastery → purpose, with engineered social moments (first reaction, "12 people have a similar profile", first remix) and the principle that the platform reveals itself through use, never onboarding.
- **sources:** docs/social001_vonc_tiktok_social/002e_concept_spark(6).md#What's-In-It-For-Users, #The-User-Journey
- **relations:** archetype system; games/puzzle retention; revenue model
- **verify-later:** none built

<!-- SOURCE: U25_leopardess_social.md -->
### Games and daily-puzzle retention ecosystem
- **category:** social-media
- **status-signal:** aspirational
- **status-evidence:** 002e: "Explore this in detail — could be the primary retention driver" (marked to-explore, not planned into any roadmap phase).
- **what:** A flagged expansion space beyond the Gauntlet: Wordle-style daily challenges with shareable result grids; daily games generated from scraping output (higher/lower with real stats, trend trivia); competitive/timed/bracket formats; streak mechanics (participation, prediction accuracy, creativity); seasonal tournaments; micro-games as retention bridges — a whole daily-play ecosystem tied to real-world content, potentially rivalling NYT Games.
- **sources:** docs/social001_vonc_tiktok_social/002e_concept_spark(6).md#Games-Daily-Puzzles-and-Gamification
- **relations:** Daily Gauntlet; predictions/time capsules
- **verify-later:** none built

<!-- SOURCE: U25_leopardess_social.md -->
### Spark revenue model
- **category:** social-media
- **status-signal:** aspirational
- **status-evidence:** 002e "Revenue Model (future)" — all items future-tense.
- **what:** Low subscription (£3–5/month) covering AI costs; brand-sponsored challenges (creators selected on niche reputation, not follower count — meritocratic sponsorship); revenue share on high-engagement showcases with challenge-driven supply control against content mills; creator subscription channels; collab marketplace; vertical expert consultations. Prediction staking uses reputation tokens, never money.
- **sources:** docs/social001_vonc_tiktok_social/002e_concept_spark(6).md#Revenue-Model, #Tier-4
- **relations:** AI cost architecture; motivation hierarchy
- **verify-later:** none built

<!-- SOURCE: U25_leopardess_social.md -->
### Vertical integration of Spark mechanics into domain sites
- **category:** social-media
- **status-signal:** aspirational
- **status-evidence:** 003d: "v4_and_beyond … Vertical integration … speculative"; 002e "Verticals after mechanics proven".
- **what:** The same mechanics re-flavoured per vertical: vet/pet (wholesome), finance (prediction-heavy), fashion (image-dominant), food (constraint challenges), with vonc.com as the unconstrained proving ground. Echoes The Forge's vertical embed idea; also the component-library payoff — a second site with Spark features reuses the generated interactive-platform components.
- **sources:** docs/social001_vonc_tiktok_social/002e_concept_spark(6).md#Vertical-Integration; docs/social001_vonc_tiktok_social/003d_spark_strategic_planning_architecture.md#Second-build-reuses-everything
- **relations:** component selector/creator; The Forge
- **verify-later:** none built

<!-- SOURCE: U25_leopardess_social.md -->
### The Forge — AI-seeded community knowledge platform
- **category:** social-media
- **status-signal:** abandoned
- **status-evidence:** 001 doc "Status: Concept stage. Parked for future development." (file dated Mar 2026; no later reference builds on it).
- **what:** Product concept predating Spark: AI answers are published as explicit first drafts into a categorised community feed; humans validate/challenge/add experience/fork; the AI synthesises human input into improved versions, visibly showing evolution. Key dynamics: correcting the AI is rewarding not adversarial; "Beat the AI" competing answers earn domain reputation; AI as fair debate synthesiser; vertical Forge embeds (e.g. vets refining pet-owner answers). Open questions recorded: moderation at scale, cold start, AI identity, expert monetisation.
- **sources:** docs/social001_vonc_tiktok_social/001_concept_the_forge_humans_edit_ai_responses.md; docs/social001_vonc_tiktok_social/002pre_whole_chat (origin transcript)
- **relations:** Spark (successor concept, keeps AI-as-participant-not-oracle DNA); hitl
- **verify-later:** none (never built)

<!-- SOURCE: U25_leopardess_social.md -->
### Spark — AI game-master social platform (core concept)
- **category:** social-media
- **status-signal:** partial
- **status-evidence:** 002e "Status: Active exploration. All mechanics candidates for live testing"; the v1 content-first site is live on vonc.com (minilobby docs, 2026-07).
- **what:** "AI-driven provocation engine where the world is the content and your take is the game." AI occupies the game-master/producer role — framing, structure, scoring, recaps, synthesis, hosting, matchmaking — and is deliberately absent from responses, humour, personality, takes ("producer, not performer"; when something is funny, a human did it). Differentiators: opinion-first entry ("what's your take?" beats create-from-scratch), ephemeral game-structured challenges, rooms not feeds, showcase+competitive dual mode. Includes the AI-slop strategy (restrict AI to strengths; deliberately-bad AI takes as straight-man seeds) and moderation-by-design (AI referee/yellow cards, positive selection, provocation design as tone control, ephemerality limits damage, "Cursed" as community signal). Second-screen positioning: "TikTok: watch. Twitter: shout. Spark: play."
- **sources:** docs/social001_vonc_tiktok_social/002e_concept_spark(6).md#Positioning, #AI-Role, #The-AI-Slop-Problem, #Moderation; docs/social001_vonc_tiktok_social/trigger_script/004_submit_vonc_trigger.sh (mission brief)
- **relations:** The Forge (predecessor); vonc.com v1 site; provocation engine; archetype system
- **verify-later:** site_specs aspects mission/roadmap for vonc site 9ec3b9ee

<!-- SOURCE: U25_leopardess_social.md -->
### Arena + Stage dual modes and their mechanic families
- **category:** social-media
- **status-signal:** aspirational
- **status-evidence:** 003d roadmap: v3 "Live challenge rooms. Arena mode. Chains. Duels." status "speculative"; v4 "Stage mode … speculative".
- **what:** Two complementary energies that feed each other: Arena (competitive — provocations, reaction vocabulary Genius/Delusional/Suspicious/Based/Cursed, remix chains with shared credit, 60-second duels with AI referee, Misfit Mashup) and Stage (showcase — springboard challenges, Fire/Inspired/Stealing-This/Teach-Me/Vibe reactions, quality-curated Tune In follow, niche discovery rooms, Glow-Up progress reels, Teach-Me-triggered mentorship, AI-matchmade collabs, Rising newcomer spotlight, Taste Graph curator prestige). Flywheel: Stage showcases become Arena provocations; Arena winners get Stage suggestions; over time the platform generates its own culture.
- **sources:** docs/social001_vonc_tiktok_social/002e_concept_spark(6).md#Two-Modes, #Arena-Mechanics, #Stage-Mechanics; 002b (first appearance of the Arena/Stage split)
- **relations:** Spark core; rooms-not-feeds; emergent profiles
- **verify-later:** none built (v3/v4)

<!-- SOURCE: U25_leopardess_social.md -->
### Rooms-not-feeds architecture and the engagement-depth spectrum
- **category:** social-media
- **status-signal:** aspirational
- **status-evidence:** 002e: room mechanics and The Drift are design prose; v1 ships only static pages (003d "v1_content_first").
- **what:** Structural anti-feed design: a Lobby of 3–5 live rooms with energy indicators; room zones (Floor active / Gallery spectating with zero barrier); crowd-energy feedback (cards heat, reactions ripple, tug-of-war splits); ephemeral challenges with lasting recaps; Director's Cut AI replays; serialised prediction challenges; multi-format rotation; "The Crowd Speaks" synthesis; Moments (remarkable outcomes elevated to permanent shareable record). The Drift is the passive snackable stream; each depth level (Drift → Lobby → Gallery → Floor → solo modes) is complete in itself. Sound/haptics design included.
- **sources:** docs/social001_vonc_tiktok_social/002e_concept_spark(6).md#Shared-Mechanics, #The-Drift, #Sound-and-Haptics
- **relations:** Arena/Stage; lobby-grid component (v1's static echo of the lobby)
- **verify-later:** none built

<!-- SOURCE: U25_leopardess_social.md -->
### Behavioural archetype system + Daily Gauntlet
- **category:** social-media
- **status-signal:** partial
- **status-evidence:** 003d gauntlet page spec with 8 archetypes; RUNBOOK_minilobby §0: gauntlet tool + archetype-taster-quiz deployed, archetype hub (8 entity pages) live 2026-07-12.
- **what:** Identity engine: emergent profiles built from behaviour (can't be faked), producing archetypes (Surgeon, Wildcard, Oracle, Catalyst, Judge, Maker, Scout, Mentor) with secondary tendencies, earned via the Daily Gauntlet (5 provocations, 5 minutes, scored on speed/originality/consistency/topic preference) and shareable as visual cards — the "viral Trojan horse" (BuzzFeed/MBTI/Hogwarts dynamics on demonstrated behaviour, works with zero community). Radical anonymity in Arena (no usernames during play, reveal after, reputation never boosts visibility) and rotating achievable status games instead of permanent karma. On vonc v1 this exists as client-side tools + a canon archetype content set (088/089 note live archetype-combinations copy had drifted off-canon).
- **sources:** docs/social001_vonc_tiktok_social/003d_spark_strategic_planning_architecture.md#gauntlet content_context; docs/social001_vonc_tiktok_social/002e_concept_spark(6).md#The-archetype-as-viral-Trojan-horse, #Identity-and-Profiles, #Radical-Anonymity; docs/social001_vonc_tiktok_social/minilobby_task/088_archetype_entity_pages.sql (header)
- **relations:** archetype hub build; content-first launch; Daily Gauntlet page
- **verify-later:** /tools/gauntlet/index.html and /tools/archetype-taster-quiz/ on vonc.com

<!-- SOURCE: U25_leopardess_social.md -->
### Cold-start design: AI sparring partner and solo-first completeness
- **category:** social-media
- **status-signal:** aspirational
- **status-evidence:** 003d: "v2_sparring_and_interaction … status: directional, depends_on backend API infrastructure".
- **what:** The empty-room solution: first 10 seconds are a provocation card + text input + timer with no signup ("Engage first. Understand second. Commit third."); the AI is a transparent sparring partner (counter-arguments, not chatbot small-talk) so the experience is complete for one person; the scraping+AI flywheel self-fills content; invites are challenges not invitations; explicit scale thresholds (1 / 5–20 / 50–200 / 500+) each designed to feel complete. Sparring is also the top per-user cost risk with named mitigations (rate limits, smallest viable model, cached counters).
- **sources:** docs/social001_vonc_tiktok_social/002e_concept_spark(6).md#Cold-Start-Design, #The-First-Visit; 002 (original fuller treatment: Spark Solo, landing page day 1)
- **relations:** AI cost architecture; content-first launch
- **verify-later:** none built (v2)

<!-- SOURCE: U25_leopardess_social.md -->
### Provocation engine — layered content production architecture
- **category:** social-media
- **status-signal:** aspirational
- **status-evidence:** 002e technical-layers section is design; NOTES_provocations-archive-list (2026-07-09): "the archive feed is hand-committed until the Phase-3 pipeline emits provocations.json".
- **what:** Six-layer production line: (1) Raw Feed — scrapers pull social-interest trends (not newspaper news); (2) Framing Engine — cheap local models (Mistral/Llama 8B CPU) generate 5–10 provocation candidates per item, ~2,000/day; (3) Curation Gate — stronger model or human picks the best 15–20, learns from engagement; (4) Mashup Engine — foundation models find non-obvious connections, 2–5 calls/day; (5) Serialisation Tracker — narrative threads, mostly database; (6) Niche Detector — embeddings/ML clustering for Stage rooms. Content tone contract: interesting not dark, slightly adult never gruesome, competitive without money betting.
- **sources:** docs/social001_vonc_tiktok_social/002e_concept_spark(6).md#Provocation-Engine, #Content-Tone; docs/social001_vonc_tiktok_social/002_concept_spark.md#The-Provocation-Engine (original layer detail)
- **relations:** Phase-3 provocations.json pipeline (its v1 delivery vehicle); AI cost architecture; news-feed-pipeline (sibling scraping infrastructure)
- **verify-later:** any provocation-orchestrator agent definition; scheduled_tasks for provocation refresh

<!-- SOURCE: U25_leopardess_social.md -->
### AI cost architecture: fixed background vs per-user scaling
- **category:** social-media
- **status-signal:** aspirational
- **status-evidence:** 002d/002e cost tables are design projections ("Scaling Scenarios … 100,000 DAU"); nothing metered in production.
- **what:** Cost-shaping principle: most AI cost is fixed (content production scales with ~15–20 challenges/day, not users) and runs on cheap local models (~£5/day background at scale); the only linearly-scaling cost is per-user interactive AI (sparring, Gauntlet scoring) and is throttled by design. Projected £0.003–0.008/user/day, so a £3–5/month subscription covers compute several times over; break-even on subscriptions before ads/brand revenue.
- **sources:** docs/social001_vonc_tiktok_social/002e_concept_spark(6).md#AI-Cost-Architecture; 002d (first appearance of the full cost tables)
- **relations:** provocation engine; cold-start sparring; revenue model
- **verify-later:** none (projection)

<!-- SOURCE: U25_leopardess_social.md -->
### Content-first launch strategy for Spark (vonc.com as destination)
- **category:** social-media
- **status-signal:** partial
- **status-evidence:** 003d "current_phase: v1_content_first … Static S3 site. Provocations as SEO content. Daily Gauntlet as viral archetype quiz"; the site is live per minilobby docs.
- **what:** Don't launch a social platform; launch a content destination that happens to have interactive features. Daily provocations with the AI's take are SEO pages with shareable URLs; every provocation/response generates a self-contained share card ("47 people responded. Think you'd do better?" — the TikTok growth pattern for text); the archetype quiz works with zero community; vertical clustering happens organically by arrival path (the Reddit model — nobody joins "Reddit"). First-week content calendar defined (daily provocations, Gauntlet, weird-stat micro-content; weekly mashup + prediction post). "The first visit experience IS the pitch."
- **sources:** docs/social001_vonc_tiktok_social/002e_concept_spark(6).md#Launch-Strategy; docs/social001_vonc_tiktok_social/003d_spark_strategic_planning_architecture.md#The-initial-request
- **relations:** Spark core; vonc.com v1 site; provocation engine
- **verify-later:** vonc.com live pages; provocations archive content

<!-- SOURCE: U25_leopardess_social.md -->
### Motivation hierarchy and designed user journey
- **category:** social-media
- **status-signal:** aspirational
- **status-evidence:** 002c introduced the motivation tiers; 002d the journey; both remain design prose in 002e.
- **what:** Retention design: four motivation tiers (identity/status; belonging/connection; growth/learning; financial reward — "financial reward follows demonstrated value, nobody buys prominence") mapped to user types (casual → professional creator; acquire at Tier 2). A staged journey — first 5 seconds to month 6+ — engineered as intrigue → creation → validation → habit → identity → community → mastery → purpose, with engineered social moments (first reaction, "12 people have a similar profile", first remix) and the principle that the platform reveals itself through use, never onboarding.
- **sources:** docs/social001_vonc_tiktok_social/002e_concept_spark(6).md#What's-In-It-For-Users, #The-User-Journey
- **relations:** archetype system; games/puzzle retention; revenue model
- **verify-later:** none built

<!-- SOURCE: U25_leopardess_social.md -->
### Games and daily-puzzle retention ecosystem
- **category:** social-media
- **status-signal:** aspirational
- **status-evidence:** 002e: "Explore this in detail — could be the primary retention driver" (marked to-explore, not planned into any roadmap phase).
- **what:** A flagged expansion space beyond the Gauntlet: Wordle-style daily challenges with shareable result grids; daily games generated from scraping output (higher/lower with real stats, trend trivia); competitive/timed/bracket formats; streak mechanics (participation, prediction accuracy, creativity); seasonal tournaments; micro-games as retention bridges — a whole daily-play ecosystem tied to real-world content, potentially rivalling NYT Games.
- **sources:** docs/social001_vonc_tiktok_social/002e_concept_spark(6).md#Games-Daily-Puzzles-and-Gamification
- **relations:** Daily Gauntlet; predictions/time capsules
- **verify-later:** none built

<!-- SOURCE: U25_leopardess_social.md -->
### Spark revenue model
- **category:** social-media
- **status-signal:** aspirational
- **status-evidence:** 002e "Revenue Model (future)" — all items future-tense.
- **what:** Low subscription (£3–5/month) covering AI costs; brand-sponsored challenges (creators selected on niche reputation, not follower count — meritocratic sponsorship); revenue share on high-engagement showcases with challenge-driven supply control against content mills; creator subscription channels; collab marketplace; vertical expert consultations. Prediction staking uses reputation tokens, never money.
- **sources:** docs/social001_vonc_tiktok_social/002e_concept_spark(6).md#Revenue-Model, #Tier-4
- **relations:** AI cost architecture; motivation hierarchy
- **verify-later:** none built

<!-- SOURCE: U25_leopardess_social.md -->
### Vertical integration of Spark mechanics into domain sites
- **category:** social-media
- **status-signal:** aspirational
- **status-evidence:** 003d: "v4_and_beyond … Vertical integration … speculative"; 002e "Verticals after mechanics proven".
- **what:** The same mechanics re-flavoured per vertical: vet/pet (wholesome), finance (prediction-heavy), fashion (image-dominant), food (constraint challenges), with vonc.com as the unconstrained proving ground. Echoes The Forge's vertical embed idea; also the component-library payoff — a second site with Spark features reuses the generated interactive-platform components.
- **sources:** docs/social001_vonc_tiktok_social/002e_concept_spark(6).md#Vertical-Integration; docs/social001_vonc_tiktok_social/003d_spark_strategic_planning_architecture.md#Second-build-reuses-everything
- **relations:** component selector/creator; The Forge
- **verify-later:** none built

<!-- SOURCE: U23_docs_root_vonc.md -->
### Spark daily-provocation product (vonc.com)
- **category:** vonc
- **status-signal:** partial
- **status-evidence:** HANDOFF 2026-07-09 §0: index + arena + archive live and browser-verified; but "the data file is currently hand-committed... a Phase-3 pipeline will emit it"; v1 roadmap features (daily_provocation_generation_from_scraping) not built.
- **what:** vonc.com / "Spark" — an AI daily-provocation platform: one charged provocation per day, users file a position, "the Gauntlet" scores the room, users get an Archetype. "The product IS the landing page": a single provocation card fills the screen; daily static regeneration; AI as producer (frames/scores/curates), not performer. v1 = daily provocations + Gauntlet; v3 concept = live challenge rooms. Serves as the platform's live test bed for the runtime-fill mechanism.
- **sources:** docs/PLAN_provocation-card(3).md#source-spec; docs/HANDOFF_2026-07-09_vonc_spark_minilobby_trim(2).md#§0/§2; docs/RUNNING_NOTES_vonc(36).md#2026-06-25-~17:15; docs/RUNNING_NOTES_vonc_v2(28).md#carried-forward-state
- **relations:** runtime-fill mechanism; Phase-3 provocation data pipeline; provocation-card/lobby-grid/provocations-archive-list components
- **verify-later:** live vonc.com; sites row 9ec3b9ee-5b08-461b-b4f8-9e1e03579c74; site_specs aspects (mission, roadmap, cta)

<!-- SOURCE: U23_docs_root_vonc.md -->
### Phase-3 provocation data pipeline (provocation-generator + orchestrator + render action + daily schedule)
- **category:** vonc
- **status-signal:** aspirational
- **status-evidence:** Phase-1 diagnostics confirmed "a clean slate — nothing exists yet" (2026-06-25); FX-4 checkbox never ticked; provocations.json still hand-committed as of 2026-07-09.
- **what:** The pipeline that would generate `/data/provocations.json` daily: clone the news pipeline — seed content_sources (trending-topic scraping targets) → reuse feed-ingester → NEW provocation-generator agent (LLM: raw topics → provocations + AI takes; generative analogue of feed-triage) → NEW render_provocations_section Go action (mirror of render_news_section; Go struct defines the JSON shape; returns a files map for git_commit) → provocation-orchestrator (clone of content-feed-orchestrator) → scheduled_tasks row `provocation-refresh` (daily; the column is `name`, not task_name). Open questions recorded: sources, volume per day, archive-page reads.
- **sources:** docs/PLAN_spark_provocation_pipeline.md; docs/RUNBOOK_phase2_provocation_js(29).md#data-deploy + #gap-1; docs/RUNBOOK_vonc_migrations(14).md#step-8
- **relations:** news feed pipeline (the template); provocations.json contract (the target shape); Spark product
- **verify-later:** absence of provocation-* agent_definitions/scheduled_tasks/content_sources for vonc; render_news_section_action.go as the model

<!-- SOURCE: U23_docs_root_vonc.md -->
### provocations.json data contract (today / lobby / arena / archive)
- **category:** vonc
- **status-signal:** deployed
- **status-evidence:** v3 served live (curl grep '"archive"' = 1, 2026-07-07); all three loaders verified filling from it in a browser.
- **what:** The versioned feed contract for Spark's runtime-fill sections: `generated_at`; `today` {eyebrow, headline (may carry `<em>`), body, primary_cta/secondary_cta {label,url}, stats ×3}; `lobby` [4 × {icon,title,desc,url}] (becomes dead after the mini-lobby trim); `arena` — an OBJECT {eyebrow,title,subtitle,cta_label,cta,cards[≤6]} because the grid's header + CTA need data too; `archive` {entries[≤24] {date,title,teaser,stat,url}, newest-first}. Evolved v1→v3 in provocations.sample.json; hand-committed interim, the fixed generation target for Phase 3.
- **sources:** docs/RUNNING_NOTES_vonc(36).md#2026-06-29-~17:35 (confirmed-good shape); docs/PLAN_lobby-grid(2).md#build-progress; docs/SPEC_provocations-archive-list.md#data-contract; docs/provocations.sample(3).json
- **relations:** the three loaders; Phase-3 pipeline; runtime-fill mechanism
- **verify-later:** live vonc.com/data/provocations.json keys

<!-- SOURCE: U23_docs_root_vonc.md -->
### provocation-card component (daily hero card) + mini-lobby trim
- **category:** vonc
- **status-signal:** partial
- **status-evidence:** "Live and working via Path-2 loader" (PLAN status); trim CONFIRMED 2026-07-04, drafted 2026-07-09, blocked on the bundle verdict — not executed within this corpus.
- **what:** The Spark centrepiece: single daily contested claim + AI take + 3 stats + 2 CTAs + (currently) a 4-card mini-lobby, filled at runtime from `today`/`lobby` by provocation-card-loader against the `.pc-*` DOM contract. JS-required by design — do NOT "fix" by baking content. Known limitation: the underlying template is Mode-B broken (loader masks it; JS-off shows `<no value>`). NEXT TASK: trim the mini-lobby (template pc-card-grid block, loader lobby fill, the orphaned 1fr-1fr media query, the dead hover script) because lobby-grid owns the arena role — with the method itself under a bundle verdict since HTML patching is the rejected mechanism.
- **sources:** docs/PLAN_provocation-card(3).md; docs/HANDOFF_2026-07-09_vonc_spark_minilobby_trim(2).md#§4; docs/RUNNING_NOTES_vonc(36).md#2026-06-29-~17:20; docs/provocation_card_loader.js (header)
- **relations:** lobby-grid overlap decision; sanctioned edit paths; runtime-fill mechanism
- **verify-later:** content_components 6163ff14 html_template (pc-card-grid still present?); js_snippets provocation-card-loader lobby block

<!-- SOURCE: U23_docs_root_vonc.md -->
### lobby-grid arena component (six-room grid)
- **category:** vonc
- **status-signal:** deployed
- **status-evidence:** "lobby-grid DONE (browser-verified)" 2026-07-04 — six arena cards + pulsing stat dots + "Enter the Arena" live; PLAN_lobby-grid marked DELIVERED 2026-07-09.
- **what:** The Arena lobby: 6-card grid (1 featured spanning 2 cols, 4 standard, 1 wide), each card icon (SVG inner markup with emoji fallback)/tag/title/desc/stat + pulsing dot, plus header and CTA — filled at runtime from `arena` by lobby-grid-loader. Honest v1 semantics decision: "live rooms" is a v3 concept, so in v1 the grid shows TODAY'S PROVOCATIONS as enterable cards. Confirmed decisions: lobby-grid is the primary "today's provocations grid" (D-A) with the `arena` object as feed (D-B). Its build was deliberately the reference implementation for the loader-builder design.
- **sources:** docs/PLAN_lobby-grid(2).md; docs/RUNNING_NOTES_vonc_v2(28).md#2026-07-04-lobby-grid-verified; docs/lobby_grid_loader.js (header); docs/lobby_grid_install.sql (header)
- **relations:** provocation-card mini-lobby trim; loader-builder reference; marker REPLACE anchoring incident
- **verify-later:** js_snippets lobby-grid-loader; live index data-component="lobby-grid"

<!-- SOURCE: U23_docs_root_vonc.md -->
### brief-explanation static explainer (regeneration, not a loader)
- **category:** vonc
- **status-signal:** deployed
- **status-evidence:** 083 succeeded 2026-07-01 (in-place update, quality 50→100, 0→20 fields); rendered with real copy on the live index 2026-07-03.
- **what:** The "what is Spark / how it works" index explainer — STABLE brand content (eyebrow, heading with `<mark>`, description, exactly 3 numbered steps, 3 stats, 2 CTAs, illustration+badge) that belongs in build-time HTML for SEO and no-JS robustness. Establishes the key distinction: Option-2 runtime loaders are ONLY for daily-changing data shells; static shells that happen to be empty are fixed by REGENERATION with a real schema — two different resolutions for the same empty-shell symptom. Its stat fields were later re-sourced static→llm to stop generic SaaS fallbacks leaking.
- **sources:** docs/PLAN_brief-explanation(1).md; docs/RUNNING_NOTES_vonc(36).md#2026-06-29-~18:00 + #2026-07-01-~12:46; docs/RUNNING_NOTES_vonc_v2(28).md#2026-07-04-static-field-discrepancy
- **relations:** static-vs-dynamic distinction; shared component library (58363894 shared ×3 sites); component regeneration in place
- **verify-later:** content_components 58363894 field sources; idea.uk/robot-hands pending instances

<!-- SOURCE: U23_docs_root_vonc.md -->
### provocations-archive-list component + provocations archive page
- **category:** vonc
- **status-signal:** deployed
- **status-evidence:** "PROVOCATIONS-INDEX THREAD DONE" 2026-07-08: page live, 8 rows fill, ghost row eliminated; live confirm grep = 2 on 2026-07-09.
- **what:** The Provocations Archive at /provocations/index.html — destination of every primary CTA — as a single self-contained runtime-fill section: llm header fields (nothing can defer), a hidden clone-template row the loader clones per `archive.entries[]` (variable-length list vs lobby-grid's fixed six), a visible empty state so the page ships before data lands, CTA back to today. Built via the full arc: component (70d6662a, 084 trigger) → plan row → pages.sections unblock → first real build (~5 min after ten 33–65s no-ops) → loader + data → ghost-row CSS fix.
- **sources:** docs/SPEC_provocations-archive-list.md; docs/RUNNING_NOTES_vonc_v2(28).md#2026-07-08; docs/RUNBOOK_phase2_provocation_js(29).md#you-are-here; docs/provocations_archive_loader.js (header)
- **relations:** complete_error family (its 404 was the trigger); generation-time guards (first live validation); CTA graph
- **verify-later:** pages e4b3b195 build_status; live /provocations/index.html

<!-- SOURCE: U23_docs_root_vonc.md -->
### Option 1 — build-time static content for the daily shells (rejected alternative)
- **category:** vonc
- **status-signal:** abandoned
- **status-evidence:** Early migrations-runbook versions carried "Recommendation: Option 1 for the first deployable version — get real content" (dropped line); final: "DECISION MADE: Option 2... Option 1 would freeze a single set of provocations permanently, defeating the daily-content product."
- **what:** The rejected fix for the empty index shells: regenerate them WITH proper input_schemas so the content writer fills them at build time. Briefly the recommended first-version route in early runbook versions, then dropped when the original Spark roadmap (daily provocations via client-side JS) was recovered — build-time content would bake one day's provocations permanently. Survives only in its correct form: genuinely static shells (brief-explanation) ARE fixed by regeneration.
- **sources:** docs/RUNBOOK_vonc_migrations(14).md#step-7 (decision); early-version dropped lines (family diff); docs/PLAN_spark_provocation_pipeline.md#why-option-2
- **relations:** static-vs-dynamic distinction; brief-explanation (where Option-1 logic is right)
- **verify-later:** none (historical)

<!-- SOURCE: U25_leopardess_social.md -->
### Phase-3 provocation pipeline (automated provocations.json emission)
- **category:** vonc
- **status-signal:** aspirational
- **status-evidence:** RUNNING_NOTES_minilobby 2026-07-11: "There is no Phase-3 emitter yet; all prior commits to the file were hand-made."
- **what:** The missing producer for the runtime-fill economy: a provocation-orchestrator + scheduled refresh generating /data/provocations.json ({generated_at, today, arena, archive}) daily from the scraping/framing engine, replacing hand-committed sample data. The dead `lobby` key was dropped 2026-07-11 (commit c244ddc) after the mini-lobby trim. Until it exists, vonc's "daily" provocation is static.
- **sources:** docs/social001_vonc_tiktok_social/tool_docs/PLAN_provocation-card(4).md#Data-contract; docs/social001_vonc_tiktok_social/tool_docs/SPEC_provocations-archive-list(1).md#Data-contract; docs/social001_vonc_tiktok_social/minilobby_task/RUNNING_NOTES_minilobby_task.md#2026-07-10/11
- **relations:** provocation engine (the design it implements); runtime-fill mechanism; scheduler-and-tasks
- **verify-later:** agent_definitions for any provocation-orchestrator; scheduled_tasks; sites repo /data/provocations.json history

<!-- SOURCE: U25_leopardess_social.md -->
### vonc.com Spark v1 site (the live testbed)
- **category:** vonc
- **status-signal:** deployed
- **status-evidence:** 083_update_work_items.sh lists the 8 live pages; HANDOFF §3 "Index page — live, six sections … Provocations archive — CLOSED 2026-07-08"; archetype hub live 2026-07-12.
- **what:** The built v1: index (hero, provocation-card, gauntlet-cta, brief-explanation, lobby-grid, system-stats), /provocations/index.html archive, about, contact, archetypes hub + 8 entity pages, blog/provocation, and two tools (gauntlet, archetype-taster-quiz). Serves as the platform's live test bed for runtime-fill, component generation, discovery checks and the section-editor; "the landing page IS the product — a provocation card, not a marketing page".
- **sources:** docs/social001_vonc_tiktok_social/trigger_script/083_update_work_items.sh; docs/social001_vonc_tiktok_social/minilobby_task/HANDOFF_2026-07-09_vonc_spark_minilobby_trim(2).md#3; docs/social001_vonc_tiktok_social/minilobby_task/RUNBOOK_minilobby_task.md#0
- **relations:** content-first launch; runtime-fill; archetype hub; site id 9ec3b9ee-5b08-461b-b4f8-9e1e03579c74
- **verify-later:** live vonc.com pages; pages table for the site

<!-- SOURCE: U25_leopardess_social.md -->
### Archetype hub built with existing machinery (entity pages + query-resolved grid)
- **category:** vonc
- **status-signal:** deployed
- **status-evidence:** RUNNING_NOTES_minilobby 2026-07-12: "End state, live-verified: archetypes.html shows 8 cards … all 8 detail pages HTTP 200, each with its icon."
- **what:** Fix for a page that had "rendered zero archetypes": archetype-grid is build-time query-resolved (items source query.pages_where_type) — a third content mode beside static and runtime-fill — and its page_type value was kebab-forbidden (chk_page_type_kebab_case) with zero matching pages. Approach A created 8 site_plan_pages (role entity-page), 24 plan sections, 8 page-scope site_plan_imagery hero rows consuming the 8 orphaned icon assets via kind-alias resolution, plus 8 pages rows (page-build-handler loads pages, never creates them), then reconcile_site_plan emitted the builds. 089 re-authored generic writer copy from the spec's archetype canon via content_data (light no-LLM rerender).
- **sources:** docs/social001_vonc_tiktok_social/minilobby_task/088_archetype_entity_pages.sql (header); 089_archetype_page_copy.sql (header); docs/social001_vonc_tiktok_social/minilobby_task/RUNNING_NOTES_minilobby_task.md#2026-07-12
- **relations:** site-plan-and-reconciler; behavioural archetype system; illustration/section-imagery resolution
- **verify-later:** pages page_type='entity-page' rows; archetype-grid input_schema source; chk_page_type_kebab_case

<!-- SOURCE: U23_docs_root_vonc.md -->
### Spark daily-provocation product (vonc.com)
- **category:** vonc
- **status-signal:** partial
- **status-evidence:** HANDOFF 2026-07-09 §0: index + arena + archive live and browser-verified; but "the data file is currently hand-committed... a Phase-3 pipeline will emit it"; v1 roadmap features (daily_provocation_generation_from_scraping) not built.
- **what:** vonc.com / "Spark" — an AI daily-provocation platform: one charged provocation per day, users file a position, "the Gauntlet" scores the room, users get an Archetype. "The product IS the landing page": a single provocation card fills the screen; daily static regeneration; AI as producer (frames/scores/curates), not performer. v1 = daily provocations + Gauntlet; v3 concept = live challenge rooms. Serves as the platform's live test bed for the runtime-fill mechanism.
- **sources:** docs/PLAN_provocation-card(3).md#source-spec; docs/HANDOFF_2026-07-09_vonc_spark_minilobby_trim(2).md#§0/§2; docs/RUNNING_NOTES_vonc(36).md#2026-06-25-~17:15; docs/RUNNING_NOTES_vonc_v2(28).md#carried-forward-state
- **relations:** runtime-fill mechanism; Phase-3 provocation data pipeline; provocation-card/lobby-grid/provocations-archive-list components
- **verify-later:** live vonc.com; sites row 9ec3b9ee-5b08-461b-b4f8-9e1e03579c74; site_specs aspects (mission, roadmap, cta)

<!-- SOURCE: U23_docs_root_vonc.md -->
### Phase-3 provocation data pipeline (provocation-generator + orchestrator + render action + daily schedule)
- **category:** vonc
- **status-signal:** aspirational
- **status-evidence:** Phase-1 diagnostics confirmed "a clean slate — nothing exists yet" (2026-06-25); FX-4 checkbox never ticked; provocations.json still hand-committed as of 2026-07-09.
- **what:** The pipeline that would generate `/data/provocations.json` daily: clone the news pipeline — seed content_sources (trending-topic scraping targets) → reuse feed-ingester → NEW provocation-generator agent (LLM: raw topics → provocations + AI takes; generative analogue of feed-triage) → NEW render_provocations_section Go action (mirror of render_news_section; Go struct defines the JSON shape; returns a files map for git_commit) → provocation-orchestrator (clone of content-feed-orchestrator) → scheduled_tasks row `provocation-refresh` (daily; the column is `name`, not task_name). Open questions recorded: sources, volume per day, archive-page reads.
- **sources:** docs/PLAN_spark_provocation_pipeline.md; docs/RUNBOOK_phase2_provocation_js(29).md#data-deploy + #gap-1; docs/RUNBOOK_vonc_migrations(14).md#step-8
- **relations:** news feed pipeline (the template); provocations.json contract (the target shape); Spark product
- **verify-later:** absence of provocation-* agent_definitions/scheduled_tasks/content_sources for vonc; render_news_section_action.go as the model

<!-- SOURCE: U23_docs_root_vonc.md -->
### provocations.json data contract (today / lobby / arena / archive)
- **category:** vonc
- **status-signal:** deployed
- **status-evidence:** v3 served live (curl grep '"archive"' = 1, 2026-07-07); all three loaders verified filling from it in a browser.
- **what:** The versioned feed contract for Spark's runtime-fill sections: `generated_at`; `today` {eyebrow, headline (may carry `<em>`), body, primary_cta/secondary_cta {label,url}, stats ×3}; `lobby` [4 × {icon,title,desc,url}] (becomes dead after the mini-lobby trim); `arena` — an OBJECT {eyebrow,title,subtitle,cta_label,cta,cards[≤6]} because the grid's header + CTA need data too; `archive` {entries[≤24] {date,title,teaser,stat,url}, newest-first}. Evolved v1→v3 in provocations.sample.json; hand-committed interim, the fixed generation target for Phase 3.
- **sources:** docs/RUNNING_NOTES_vonc(36).md#2026-06-29-~17:35 (confirmed-good shape); docs/PLAN_lobby-grid(2).md#build-progress; docs/SPEC_provocations-archive-list.md#data-contract; docs/provocations.sample(3).json
- **relations:** the three loaders; Phase-3 pipeline; runtime-fill mechanism
- **verify-later:** live vonc.com/data/provocations.json keys

<!-- SOURCE: U23_docs_root_vonc.md -->
### provocation-card component (daily hero card) + mini-lobby trim
- **category:** vonc
- **status-signal:** partial
- **status-evidence:** "Live and working via Path-2 loader" (PLAN status); trim CONFIRMED 2026-07-04, drafted 2026-07-09, blocked on the bundle verdict — not executed within this corpus.
- **what:** The Spark centrepiece: single daily contested claim + AI take + 3 stats + 2 CTAs + (currently) a 4-card mini-lobby, filled at runtime from `today`/`lobby` by provocation-card-loader against the `.pc-*` DOM contract. JS-required by design — do NOT "fix" by baking content. Known limitation: the underlying template is Mode-B broken (loader masks it; JS-off shows `<no value>`). NEXT TASK: trim the mini-lobby (template pc-card-grid block, loader lobby fill, the orphaned 1fr-1fr media query, the dead hover script) because lobby-grid owns the arena role — with the method itself under a bundle verdict since HTML patching is the rejected mechanism.
- **sources:** docs/PLAN_provocation-card(3).md; docs/HANDOFF_2026-07-09_vonc_spark_minilobby_trim(2).md#§4; docs/RUNNING_NOTES_vonc(36).md#2026-06-29-~17:20; docs/provocation_card_loader.js (header)
- **relations:** lobby-grid overlap decision; sanctioned edit paths; runtime-fill mechanism
- **verify-later:** content_components 6163ff14 html_template (pc-card-grid still present?); js_snippets provocation-card-loader lobby block

<!-- SOURCE: U23_docs_root_vonc.md -->
### lobby-grid arena component (six-room grid)
- **category:** vonc
- **status-signal:** deployed
- **status-evidence:** "lobby-grid DONE (browser-verified)" 2026-07-04 — six arena cards + pulsing stat dots + "Enter the Arena" live; PLAN_lobby-grid marked DELIVERED 2026-07-09.
- **what:** The Arena lobby: 6-card grid (1 featured spanning 2 cols, 4 standard, 1 wide), each card icon (SVG inner markup with emoji fallback)/tag/title/desc/stat + pulsing dot, plus header and CTA — filled at runtime from `arena` by lobby-grid-loader. Honest v1 semantics decision: "live rooms" is a v3 concept, so in v1 the grid shows TODAY'S PROVOCATIONS as enterable cards. Confirmed decisions: lobby-grid is the primary "today's provocations grid" (D-A) with the `arena` object as feed (D-B). Its build was deliberately the reference implementation for the loader-builder design.
- **sources:** docs/PLAN_lobby-grid(2).md; docs/RUNNING_NOTES_vonc_v2(28).md#2026-07-04-lobby-grid-verified; docs/lobby_grid_loader.js (header); docs/lobby_grid_install.sql (header)
- **relations:** provocation-card mini-lobby trim; loader-builder reference; marker REPLACE anchoring incident
- **verify-later:** js_snippets lobby-grid-loader; live index data-component="lobby-grid"

<!-- SOURCE: U23_docs_root_vonc.md -->
### brief-explanation static explainer (regeneration, not a loader)
- **category:** vonc
- **status-signal:** deployed
- **status-evidence:** 083 succeeded 2026-07-01 (in-place update, quality 50→100, 0→20 fields); rendered with real copy on the live index 2026-07-03.
- **what:** The "what is Spark / how it works" index explainer — STABLE brand content (eyebrow, heading with `<mark>`, description, exactly 3 numbered steps, 3 stats, 2 CTAs, illustration+badge) that belongs in build-time HTML for SEO and no-JS robustness. Establishes the key distinction: Option-2 runtime loaders are ONLY for daily-changing data shells; static shells that happen to be empty are fixed by REGENERATION with a real schema — two different resolutions for the same empty-shell symptom. Its stat fields were later re-sourced static→llm to stop generic SaaS fallbacks leaking.
- **sources:** docs/PLAN_brief-explanation(1).md; docs/RUNNING_NOTES_vonc(36).md#2026-06-29-~18:00 + #2026-07-01-~12:46; docs/RUNNING_NOTES_vonc_v2(28).md#2026-07-04-static-field-discrepancy
- **relations:** static-vs-dynamic distinction; shared component library (58363894 shared ×3 sites); component regeneration in place
- **verify-later:** content_components 58363894 field sources; idea.uk/robot-hands pending instances

<!-- SOURCE: U23_docs_root_vonc.md -->
### provocations-archive-list component + provocations archive page
- **category:** vonc
- **status-signal:** deployed
- **status-evidence:** "PROVOCATIONS-INDEX THREAD DONE" 2026-07-08: page live, 8 rows fill, ghost row eliminated; live confirm grep = 2 on 2026-07-09.
- **what:** The Provocations Archive at /provocations/index.html — destination of every primary CTA — as a single self-contained runtime-fill section: llm header fields (nothing can defer), a hidden clone-template row the loader clones per `archive.entries[]` (variable-length list vs lobby-grid's fixed six), a visible empty state so the page ships before data lands, CTA back to today. Built via the full arc: component (70d6662a, 084 trigger) → plan row → pages.sections unblock → first real build (~5 min after ten 33–65s no-ops) → loader + data → ghost-row CSS fix.
- **sources:** docs/SPEC_provocations-archive-list.md; docs/RUNNING_NOTES_vonc_v2(28).md#2026-07-08; docs/RUNBOOK_phase2_provocation_js(29).md#you-are-here; docs/provocations_archive_loader.js (header)
- **relations:** complete_error family (its 404 was the trigger); generation-time guards (first live validation); CTA graph
- **verify-later:** pages e4b3b195 build_status; live /provocations/index.html

<!-- SOURCE: U23_docs_root_vonc.md -->
### Option 1 — build-time static content for the daily shells (rejected alternative)
- **category:** vonc
- **status-signal:** abandoned
- **status-evidence:** Early migrations-runbook versions carried "Recommendation: Option 1 for the first deployable version — get real content" (dropped line); final: "DECISION MADE: Option 2... Option 1 would freeze a single set of provocations permanently, defeating the daily-content product."
- **what:** The rejected fix for the empty index shells: regenerate them WITH proper input_schemas so the content writer fills them at build time. Briefly the recommended first-version route in early runbook versions, then dropped when the original Spark roadmap (daily provocations via client-side JS) was recovered — build-time content would bake one day's provocations permanently. Survives only in its correct form: genuinely static shells (brief-explanation) ARE fixed by regeneration.
- **sources:** docs/RUNBOOK_vonc_migrations(14).md#step-7 (decision); early-version dropped lines (family diff); docs/PLAN_spark_provocation_pipeline.md#why-option-2
- **relations:** static-vs-dynamic distinction; brief-explanation (where Option-1 logic is right)
- **verify-later:** none (historical)

<!-- SOURCE: U25_leopardess_social.md -->
### Phase-3 provocation pipeline (automated provocations.json emission)
- **category:** vonc
- **status-signal:** aspirational
- **status-evidence:** RUNNING_NOTES_minilobby 2026-07-11: "There is no Phase-3 emitter yet; all prior commits to the file were hand-made."
- **what:** The missing producer for the runtime-fill economy: a provocation-orchestrator + scheduled refresh generating /data/provocations.json ({generated_at, today, arena, archive}) daily from the scraping/framing engine, replacing hand-committed sample data. The dead `lobby` key was dropped 2026-07-11 (commit c244ddc) after the mini-lobby trim. Until it exists, vonc's "daily" provocation is static.
- **sources:** docs/social001_vonc_tiktok_social/tool_docs/PLAN_provocation-card(4).md#Data-contract; docs/social001_vonc_tiktok_social/tool_docs/SPEC_provocations-archive-list(1).md#Data-contract; docs/social001_vonc_tiktok_social/minilobby_task/RUNNING_NOTES_minilobby_task.md#2026-07-10/11
- **relations:** provocation engine (the design it implements); runtime-fill mechanism; scheduler-and-tasks
- **verify-later:** agent_definitions for any provocation-orchestrator; scheduled_tasks; sites repo /data/provocations.json history

<!-- SOURCE: U25_leopardess_social.md -->
### vonc.com Spark v1 site (the live testbed)
- **category:** vonc
- **status-signal:** deployed
- **status-evidence:** 083_update_work_items.sh lists the 8 live pages; HANDOFF §3 "Index page — live, six sections … Provocations archive — CLOSED 2026-07-08"; archetype hub live 2026-07-12.
- **what:** The built v1: index (hero, provocation-card, gauntlet-cta, brief-explanation, lobby-grid, system-stats), /provocations/index.html archive, about, contact, archetypes hub + 8 entity pages, blog/provocation, and two tools (gauntlet, archetype-taster-quiz). Serves as the platform's live test bed for runtime-fill, component generation, discovery checks and the section-editor; "the landing page IS the product — a provocation card, not a marketing page".
- **sources:** docs/social001_vonc_tiktok_social/trigger_script/083_update_work_items.sh; docs/social001_vonc_tiktok_social/minilobby_task/HANDOFF_2026-07-09_vonc_spark_minilobby_trim(2).md#3; docs/social001_vonc_tiktok_social/minilobby_task/RUNBOOK_minilobby_task.md#0
- **relations:** content-first launch; runtime-fill; archetype hub; site id 9ec3b9ee-5b08-461b-b4f8-9e1e03579c74
- **verify-later:** live vonc.com pages; pages table for the site

<!-- SOURCE: U25_leopardess_social.md -->
### Archetype hub built with existing machinery (entity pages + query-resolved grid)
- **category:** vonc
- **status-signal:** deployed
- **status-evidence:** RUNNING_NOTES_minilobby 2026-07-12: "End state, live-verified: archetypes.html shows 8 cards … all 8 detail pages HTTP 200, each with its icon."
- **what:** Fix for a page that had "rendered zero archetypes": archetype-grid is build-time query-resolved (items source query.pages_where_type) — a third content mode beside static and runtime-fill — and its page_type value was kebab-forbidden (chk_page_type_kebab_case) with zero matching pages. Approach A created 8 site_plan_pages (role entity-page), 24 plan sections, 8 page-scope site_plan_imagery hero rows consuming the 8 orphaned icon assets via kind-alias resolution, plus 8 pages rows (page-build-handler loads pages, never creates them), then reconcile_site_plan emitted the builds. 089 re-authored generic writer copy from the spec's archetype canon via content_data (light no-LLM rerender).
- **sources:** docs/social001_vonc_tiktok_social/minilobby_task/088_archetype_entity_pages.sql (header); 089_archetype_page_copy.sql (header); docs/social001_vonc_tiktok_social/minilobby_task/RUNNING_NOTES_minilobby_task.md#2026-07-12
- **relations:** site-plan-and-reconciler; behavioural archetype system; illustration/section-imagery resolution
- **verify-later:** pages page_type='entity-page' rows; archetype-grid input_schema source; chk_page_type_kebab_case

<!-- SOURCE: U19_sql_tables_components.md -->
### site_chat_turns per-domain chatbot logging
- **category:** NEW:site-chatbot
- **status-signal:** aspirational
- **status-evidence:** Migration drafted with "NOTE ON NUMBERING: this snapshot only shows migrations up to 085. Confirm the next free migration number... before applying" — written against a snapshot, application unconfirmed in this unit.
- **what:** End-user chatbot turns from the site chatbot edge worker: one row per prompt/answer (PII), populated by a Layer-1 puller draining the edge sink with idempotent ingest via edge-supplied uuid PK; bounding outcomes (refused off-topic, capped), provenance for "why did it say that" (model, context pack_version, grounding_ids chunk list), token/latency columns name-aligned to llm_call_log, GDPR-conscious salted client_ip_hash instead of raw IPs, edge vs ingest timestamps, per-site cascade delete. Explicitly distinct from llm_call_log (build-time flywheel vs end-user data with its own retention/access profile).
- **sources:** docs/agent_docs/sql_for_tables/046_site_chat_turns.sql
- **relations:** llm_call_log; rag-retrieval (context packs / grounding chunks); edge workers.
- **verify-later:** table existence in production; edge worker + Layer-1 puller implementations.

<!-- SOURCE: U22_recent_small_docs.md -->
### Site chatbot edge worker (synchronous, not an orchestrated agent)
- **category:** NEW:site-chatbot
- **status-signal:** aspirational
- **status-evidence:** Design doc with "Suggested build order (structural first)" and "Open decisions" — canonical design, nothing deployed.
- **what:** The canonical design for a per-domain chatbot on static-S3 sites: a synchronous request/response handler on a provider-agnostic serverless edge worker (Cloudflare first), NOT run through Kafka/the chassis. Deliberate documented exception to "every agent is an orchestrator" — Kafka's async failure modes (offset replay, phantom-complete, no streaming) are wrong for live chat, and a central nginx VM would drag static traffic behind a hackable box and lose S3's hack-resistance. Worker: resolve domain → load context pack → guard limits → compose bounded prompt → stream LLM tokens (SSE) → fire-and-forget record turn.
- **sources:** docs025.../FOCUS_site_chatbot_edge_worker_and_context_pack(1).md
- **relations:** context pack, provider-agnostic deps adapters, site_chat_turns, isolated chat environment
- **verify-later:** any edge worker deploy; /api/chat route registration

<!-- SOURCE: U22_recent_small_docs.md -->
### Build-time context pack (per-domain bounded context)
- **category:** NEW:site-chatbot
- **status-signal:** aspirational
- **status-evidence:** Section 7 defines the JSON shape and versioning; produced by an unbuilt `chat-context-builder` agent.
- **what:** One per-domain JSON document published to static storage at install time, carrying identity, scope (instructions/refusal message/banned topics), build-time-selected grounding chunks (bounded by token budget), suggested model, and operational limits. The worker holds no per-site logic — the pack is the entire bounded context. Grounding is selected on Layer 1 via Ollama embeddings + pgvector; v2 optionally ships chunk vectors for in-worker per-question retrieval plus a narrow stateless embedding endpoint.
- **sources:** docs025.../FOCUS_site_chatbot_edge_worker_and_context_pack(1).md#7
- **relations:** chat-context-builder, RAG knowledge_base (install-time reuse), three-layer bounding
- **verify-later:** context-pack schema; chat-context-builder agent; pack publish-to-S3 step

<!-- SOURCE: U22_recent_small_docs.md -->
### site-chat-installer orchestration (install_chat maintenance task)
- **category:** NEW:site-chatbot
- **status-signal:** aspirational
- **status-evidence:** "Resolved: Install is a separate orchestration, triggered via a maintenance_queue install_chat task — not a build-pipeline stage." Not built.
- **what:** Chat install is its own orchestration (triggered by a `maintenance_queue` `install_chat` task, build pipeline untouched), spawning three sub-agents: `chat-context-builder` (build+publish the pack via Ollama+pgvector), `chat-widget-installer` (fork the chat widget through the existing component/tool pipeline; only difference is it POSTs to /api/chat), and `chat-route-registrar` (record the route + mark chat installed on the site, reversible via uninstall_chat). Supersedes the older `chat-suggester` gating agent from the FOCUS base version.
- **sources:** docs025.../FOCUS_site_chatbot_edge_worker_and_context_pack(1).md#install-path, docs025.../FOCUS_site_chatbot_edge_worker_and_context_pack.md (delta: chat-suggester)
- **relations:** maintenance_queue, context pack, component/tool pipeline, chat-suggester (superseded)
- **verify-later:** site-chat-installer + sub-agent definitions; install_chat maintenance task_type

<!-- SOURCE: U22_recent_small_docs.md -->
### Provider-agnostic worker (deps adapters)
- **category:** NEW:site-chatbot
- **status-signal:** aspirational
- **status-evidence:** Section 5 defines interfaces and a Cloudflare shim; "Best practice" reference impls listed, no code shipped.
- **what:** The portability strategy: a Web-platform-only core `handleChat(request, deps)` plus a ~20-line per-platform shim. Three (v2: four) small adapters — ContextStore (HTTP GET of static pack), LLMClient (Anthropic Messages over fetch, swappable to self-hosted), TurnSink (queue/D1, fire-and-forget), and v2 Embedder — each with a Cloudflare and a portable HTTP impl. Nothing vendor-specific in the core; Cloudflare/Deno/Fastly/Vercel/self-host are drop-in. Rate limiting is the least-portable concern (WAF + in-pack per-session cap floor).
- **sources:** docs025.../FOCUS_site_chatbot_edge_worker_and_context_pack(1).md#5, #6
- **relations:** edge worker, context pack, pluggable billing/LLM/storage adapter discipline
- **verify-later:** handleChat core + adapter interfaces if implemented

<!-- SOURCE: U22_recent_small_docs.md -->
### site_chat_turns table (turn recording)
- **category:** NEW:site-chatbot
- **status-signal:** aspirational
- **status-evidence:** Migration 086 written and "schema-checked" against live `sites`, but header notes "this snapshot only shows migrations up to 085. Confirm the next free migration number ... before applying."
- **what:** A `site_chat_turns` table logging each end-user prompt/answer turn per domain (question/answer as PII, refused/capped flags, model, pack_version, grounding_ids, tokens/latency named to match llm_call_log, salted client_ip_hash never raw IP). Deliberately separate from the build-time `llm_call_log` (different owner, privacy profile, and access pattern). Edge-supplied turn uuid is the PK for idempotent ingest (ON CONFLICT DO NOTHING); populated by a Layer-1 puller draining the edge sink.
- **sources:** docs025.../086_site_chat_turns.sql, docs025.../FOCUS_site_chatbot_edge_worker_and_context_pack(1).md#8
- **relations:** llm_call_log (kept separate), TurnSink, isolated chat environment (isolated-DB variant drops the FK)
- **verify-later:** site_chat_turns migration number; Layer-1 turn puller

<!-- SOURCE: U22_recent_small_docs.md -->
### Three-layer bounding (retrieval / prompt / operational)
- **category:** NEW:site-chatbot
- **status-signal:** aspirational
- **status-evidence:** Section 9 "Bounded context — the three layers"; part of the design.
- **what:** A precise decomposition of "bounded" to stop chatbot drift: retrieval bounding (only this site's grounding is in the pack, frozen at build time), prompt bounding (system prompt pins identity and emits an exact refusal message for out-of-scope questions), and operational bounding (input length, output tokens, turns/session, history window, rate limiting from pack.limits). Conflating the three is where bots go off-topic.
- **sources:** docs025.../FOCUS_site_chatbot_edge_worker_and_context_pack(1).md#9
- **relations:** context pack, edge worker composeSystemPrompt
- **verify-later:** composeSystemPrompt refusal enforcement; pack.limits guards

<!-- SOURCE: U22_recent_small_docs.md -->
### Isolated chat environment (satellite; load/hack/bug vectors)
- **category:** NEW:site-chatbot
- **status-signal:** aspirational
- **status-evidence:** "Current lean (not committed — kept open): Option Y-copy ... experiment in a sandbox." Explicitly undecided.
- **what:** A plan to run the chatbot's server-side pieces (turn store, drain, analytics, optionally chat workflow code) on infrastructure separate from the core build cluster, severing three blast vectors — load (turn write-load), hack (compromised edge worker's reachable radius), bug (chat code faulting the shared chassis). Deliberately does NOT reuse the coupled multi-cluster dispatch (which shares core Kafka/Postgres). Option X = minimal satellite (maybe no chassis at all); Option Y = full cut-down chassis (Y-copy config-only vs Y-slim purpose-built image). Boundary is one-directional, async, egress-from-core only.
- **sources:** docs025.../PLAN_isolated_chat_environment(4).md, docs025.../PLAN_isolated_chat_environment(1).md
- **relations:** remote-job-spawner (NOT reused), site_chat_turns, boundary contract, building-as-a-service
- **verify-later:** any separate chat cluster/DB; isolated-DB variant of migration 086

<!-- SOURCE: U22_recent_small_docs.md -->
### Simple paid multi-domain chat (freemium + day-pass)
- **category:** NEW:site-chatbot
- **status-signal:** aspirational
- **status-evidence:** "Status: discussion draft — direction firming."
- **what:** A deliberately simple "fast lane" route: the FOCUS edge worker + a light paywall, multi-tenant-by-Host, add a domain by publishing config + DNS (no chassis/Kafka/satellite). Monetisation is freemium + a flat day-pass (£2-5) rather than counted credits, because card processing's fixed ~20-30p fee makes sub-£5 one-off charges poor. Entitlement is a stateless signed `{domain, expiry}` token issued via a synchronous Stripe guest-checkout `redeem` (no accounts, no webhook on the critical path, no edge KV). The real cost driver is the free taster + abuse, not paying users.
- **sources:** docs025.../PLAN_simple_paid_multidomain_chat(1).md, docs025.../PLAN_simple_paid_multidomain_chat.md
- **relations:** edge worker, context pack, chat lanes (fast lane), commercial model/billing adapter
- **verify-later:** paywall gate + redeem endpoint; day-pass token signing/validation

<!-- SOURCE: U22_recent_small_docs.md -->
### Chat lanes (fast/slow/job) + warm-adapter maturation
- **category:** NEW:site-chatbot
- **status-signal:** aspirational
- **status-evidence:** Section 11 "What the agent framework buys chat (lanes and maturation)"; "This is still open by design; needs further conversation."
- **what:** A model splitting chat by what it does: fast lane (bounded Q&A, synchronous/streamed, no framework — ships independently); slow lane (turns needing work — live research, structured-data queries, running a site's tool, in-answer charts, multi-step tasks — routed by a cheap intent classifier, user warned it's slower); job lane (long-running submissions like "build me a site", ack + status + deliver). Maturation path: prove a slow-lane capability as a spawned agent (~12s cold), promote popular ones to warm adapters, end-state a resident chat-orchestrator adapter that fans out without spawning per turn.
- **sources:** docs025.../PLAN_isolated_chat_environment(4).md#11
- **relations:** simple paid multi-domain chat (fast lane), building-as-a-service (job lane), warm adapters
- **verify-later:** intent classifier; any resident chat-orchestrator adapter

<!-- SOURCE: U24f_docs_archive_remaining_small.md -->
### isolated chat satellite architecture (three blast vectors: load/hack/bug)
- **category:** NEW:site-chatbot
- **status-signal:** aspirational
- **status-evidence:** "Current lean (not committed — kept open)." Explicitly a plan document with an open central decision (Option X vs Y).
- **what:** A plan to run the site-chatbot's server-side pieces (turn storage, drain/analytics, and any chat workflow code) on infrastructure **separate** from the core build cluster, so that live chat traffic, a compromise of the internet-facing edge worker, or a chat-code bug cannot degrade or reach the webdesign/build system. Deliberately **not** built on the existing multi-cluster dispatch (Phase 4a, `remote-job-spawner`), which shares cluster A's Kafka/Postgres by design — the chat satellite instead reuses only the chassis *binaries* and action code, deployed against its own Kafka/DB/storage, with a one-directional async boundary (core publishes install triggers and content; nothing on the chat side has synchronous or write access back into core). Two options are weighed: **Option X (minimal, recommended MVP)** — pack-building stays on core, the satellite is just a turn store + puller + analytics, possibly needing no Kafka/chassis at all; **Option Y (full satellite chassis)** — the whole chat pipeline including install/pack-building moves to a cut-down copy of the chassis on the satellite. A worked "building-and-hosting-as-a-service via chat" example (a customer types a domain into another site's chatbot and gets a fully built, hosted site with its own chatbot) reframes the satellite as a second, customer-facing instance of the whole platform and pushes the design toward Option Y for that specific use case.
- **sources:** docs/_archive/agent_docs/docs025_ai_chatbot_idea_uk/excellent_discussions/PLAN_isolated_chat_environment(3).md
- **relations:** multicluster (Phase 4a, the pattern explicitly rejected as a template); SaaS commercial model (below, same document, §13)
- **verify-later:** whether any satellite cluster / separate chat Postgres exists; site_chat_turns table; remote-job-spawner (the Phase 4a mechanism used as contrast)

<!-- SOURCE: U19_sql_tables_components.md -->
### site_chat_turns per-domain chatbot logging
- **category:** NEW:site-chatbot
- **status-signal:** aspirational
- **status-evidence:** Migration drafted with "NOTE ON NUMBERING: this snapshot only shows migrations up to 085. Confirm the next free migration number... before applying" — written against a snapshot, application unconfirmed in this unit.
- **what:** End-user chatbot turns from the site chatbot edge worker: one row per prompt/answer (PII), populated by a Layer-1 puller draining the edge sink with idempotent ingest via edge-supplied uuid PK; bounding outcomes (refused off-topic, capped), provenance for "why did it say that" (model, context pack_version, grounding_ids chunk list), token/latency columns name-aligned to llm_call_log, GDPR-conscious salted client_ip_hash instead of raw IPs, edge vs ingest timestamps, per-site cascade delete. Explicitly distinct from llm_call_log (build-time flywheel vs end-user data with its own retention/access profile).
- **sources:** docs/agent_docs/sql_for_tables/046_site_chat_turns.sql
- **relations:** llm_call_log; rag-retrieval (context packs / grounding chunks); edge workers.
- **verify-later:** table existence in production; edge worker + Layer-1 puller implementations.

<!-- SOURCE: U22_recent_small_docs.md -->
### Site chatbot edge worker (synchronous, not an orchestrated agent)
- **category:** NEW:site-chatbot
- **status-signal:** aspirational
- **status-evidence:** Design doc with "Suggested build order (structural first)" and "Open decisions" — canonical design, nothing deployed.
- **what:** The canonical design for a per-domain chatbot on static-S3 sites: a synchronous request/response handler on a provider-agnostic serverless edge worker (Cloudflare first), NOT run through Kafka/the chassis. Deliberate documented exception to "every agent is an orchestrator" — Kafka's async failure modes (offset replay, phantom-complete, no streaming) are wrong for live chat, and a central nginx VM would drag static traffic behind a hackable box and lose S3's hack-resistance. Worker: resolve domain → load context pack → guard limits → compose bounded prompt → stream LLM tokens (SSE) → fire-and-forget record turn.
- **sources:** docs025.../FOCUS_site_chatbot_edge_worker_and_context_pack(1).md
- **relations:** context pack, provider-agnostic deps adapters, site_chat_turns, isolated chat environment
- **verify-later:** any edge worker deploy; /api/chat route registration

<!-- SOURCE: U22_recent_small_docs.md -->
### Build-time context pack (per-domain bounded context)
- **category:** NEW:site-chatbot
- **status-signal:** aspirational
- **status-evidence:** Section 7 defines the JSON shape and versioning; produced by an unbuilt `chat-context-builder` agent.
- **what:** One per-domain JSON document published to static storage at install time, carrying identity, scope (instructions/refusal message/banned topics), build-time-selected grounding chunks (bounded by token budget), suggested model, and operational limits. The worker holds no per-site logic — the pack is the entire bounded context. Grounding is selected on Layer 1 via Ollama embeddings + pgvector; v2 optionally ships chunk vectors for in-worker per-question retrieval plus a narrow stateless embedding endpoint.
- **sources:** docs025.../FOCUS_site_chatbot_edge_worker_and_context_pack(1).md#7
- **relations:** chat-context-builder, RAG knowledge_base (install-time reuse), three-layer bounding
- **verify-later:** context-pack schema; chat-context-builder agent; pack publish-to-S3 step

<!-- SOURCE: U22_recent_small_docs.md -->
### site-chat-installer orchestration (install_chat maintenance task)
- **category:** NEW:site-chatbot
- **status-signal:** aspirational
- **status-evidence:** "Resolved: Install is a separate orchestration, triggered via a maintenance_queue install_chat task — not a build-pipeline stage." Not built.
- **what:** Chat install is its own orchestration (triggered by a `maintenance_queue` `install_chat` task, build pipeline untouched), spawning three sub-agents: `chat-context-builder` (build+publish the pack via Ollama+pgvector), `chat-widget-installer` (fork the chat widget through the existing component/tool pipeline; only difference is it POSTs to /api/chat), and `chat-route-registrar` (record the route + mark chat installed on the site, reversible via uninstall_chat). Supersedes the older `chat-suggester` gating agent from the FOCUS base version.
- **sources:** docs025.../FOCUS_site_chatbot_edge_worker_and_context_pack(1).md#install-path, docs025.../FOCUS_site_chatbot_edge_worker_and_context_pack.md (delta: chat-suggester)
- **relations:** maintenance_queue, context pack, component/tool pipeline, chat-suggester (superseded)
- **verify-later:** site-chat-installer + sub-agent definitions; install_chat maintenance task_type

<!-- SOURCE: U22_recent_small_docs.md -->
### Provider-agnostic worker (deps adapters)
- **category:** NEW:site-chatbot
- **status-signal:** aspirational
- **status-evidence:** Section 5 defines interfaces and a Cloudflare shim; "Best practice" reference impls listed, no code shipped.
- **what:** The portability strategy: a Web-platform-only core `handleChat(request, deps)` plus a ~20-line per-platform shim. Three (v2: four) small adapters — ContextStore (HTTP GET of static pack), LLMClient (Anthropic Messages over fetch, swappable to self-hosted), TurnSink (queue/D1, fire-and-forget), and v2 Embedder — each with a Cloudflare and a portable HTTP impl. Nothing vendor-specific in the core; Cloudflare/Deno/Fastly/Vercel/self-host are drop-in. Rate limiting is the least-portable concern (WAF + in-pack per-session cap floor).
- **sources:** docs025.../FOCUS_site_chatbot_edge_worker_and_context_pack(1).md#5, #6
- **relations:** edge worker, context pack, pluggable billing/LLM/storage adapter discipline
- **verify-later:** handleChat core + adapter interfaces if implemented

<!-- SOURCE: U22_recent_small_docs.md -->
### site_chat_turns table (turn recording)
- **category:** NEW:site-chatbot
- **status-signal:** aspirational
- **status-evidence:** Migration 086 written and "schema-checked" against live `sites`, but header notes "this snapshot only shows migrations up to 085. Confirm the next free migration number ... before applying."
- **what:** A `site_chat_turns` table logging each end-user prompt/answer turn per domain (question/answer as PII, refused/capped flags, model, pack_version, grounding_ids, tokens/latency named to match llm_call_log, salted client_ip_hash never raw IP). Deliberately separate from the build-time `llm_call_log` (different owner, privacy profile, and access pattern). Edge-supplied turn uuid is the PK for idempotent ingest (ON CONFLICT DO NOTHING); populated by a Layer-1 puller draining the edge sink.
- **sources:** docs025.../086_site_chat_turns.sql, docs025.../FOCUS_site_chatbot_edge_worker_and_context_pack(1).md#8
- **relations:** llm_call_log (kept separate), TurnSink, isolated chat environment (isolated-DB variant drops the FK)
- **verify-later:** site_chat_turns migration number; Layer-1 turn puller

<!-- SOURCE: U22_recent_small_docs.md -->
### Three-layer bounding (retrieval / prompt / operational)
- **category:** NEW:site-chatbot
- **status-signal:** aspirational
- **status-evidence:** Section 9 "Bounded context — the three layers"; part of the design.
- **what:** A precise decomposition of "bounded" to stop chatbot drift: retrieval bounding (only this site's grounding is in the pack, frozen at build time), prompt bounding (system prompt pins identity and emits an exact refusal message for out-of-scope questions), and operational bounding (input length, output tokens, turns/session, history window, rate limiting from pack.limits). Conflating the three is where bots go off-topic.
- **sources:** docs025.../FOCUS_site_chatbot_edge_worker_and_context_pack(1).md#9
- **relations:** context pack, edge worker composeSystemPrompt
- **verify-later:** composeSystemPrompt refusal enforcement; pack.limits guards

<!-- SOURCE: U22_recent_small_docs.md -->
### Isolated chat environment (satellite; load/hack/bug vectors)
- **category:** NEW:site-chatbot
- **status-signal:** aspirational
- **status-evidence:** "Current lean (not committed — kept open): Option Y-copy ... experiment in a sandbox." Explicitly undecided.
- **what:** A plan to run the chatbot's server-side pieces (turn store, drain, analytics, optionally chat workflow code) on infrastructure separate from the core build cluster, severing three blast vectors — load (turn write-load), hack (compromised edge worker's reachable radius), bug (chat code faulting the shared chassis). Deliberately does NOT reuse the coupled multi-cluster dispatch (which shares core Kafka/Postgres). Option X = minimal satellite (maybe no chassis at all); Option Y = full cut-down chassis (Y-copy config-only vs Y-slim purpose-built image). Boundary is one-directional, async, egress-from-core only.
- **sources:** docs025.../PLAN_isolated_chat_environment(4).md, docs025.../PLAN_isolated_chat_environment(1).md
- **relations:** remote-job-spawner (NOT reused), site_chat_turns, boundary contract, building-as-a-service
- **verify-later:** any separate chat cluster/DB; isolated-DB variant of migration 086

<!-- SOURCE: U22_recent_small_docs.md -->
### Simple paid multi-domain chat (freemium + day-pass)
- **category:** NEW:site-chatbot
- **status-signal:** aspirational
- **status-evidence:** "Status: discussion draft — direction firming."
- **what:** A deliberately simple "fast lane" route: the FOCUS edge worker + a light paywall, multi-tenant-by-Host, add a domain by publishing config + DNS (no chassis/Kafka/satellite). Monetisation is freemium + a flat day-pass (£2-5) rather than counted credits, because card processing's fixed ~20-30p fee makes sub-£5 one-off charges poor. Entitlement is a stateless signed `{domain, expiry}` token issued via a synchronous Stripe guest-checkout `redeem` (no accounts, no webhook on the critical path, no edge KV). The real cost driver is the free taster + abuse, not paying users.
- **sources:** docs025.../PLAN_simple_paid_multidomain_chat(1).md, docs025.../PLAN_simple_paid_multidomain_chat.md
- **relations:** edge worker, context pack, chat lanes (fast lane), commercial model/billing adapter
- **verify-later:** paywall gate + redeem endpoint; day-pass token signing/validation

<!-- SOURCE: U22_recent_small_docs.md -->
### Chat lanes (fast/slow/job) + warm-adapter maturation
- **category:** NEW:site-chatbot
- **status-signal:** aspirational
- **status-evidence:** Section 11 "What the agent framework buys chat (lanes and maturation)"; "This is still open by design; needs further conversation."
- **what:** A model splitting chat by what it does: fast lane (bounded Q&A, synchronous/streamed, no framework — ships independently); slow lane (turns needing work — live research, structured-data queries, running a site's tool, in-answer charts, multi-step tasks — routed by a cheap intent classifier, user warned it's slower); job lane (long-running submissions like "build me a site", ack + status + deliver). Maturation path: prove a slow-lane capability as a spawned agent (~12s cold), promote popular ones to warm adapters, end-state a resident chat-orchestrator adapter that fans out without spawning per turn.
- **sources:** docs025.../PLAN_isolated_chat_environment(4).md#11
- **relations:** simple paid multi-domain chat (fast lane), building-as-a-service (job lane), warm adapters
- **verify-later:** intent classifier; any resident chat-orchestrator adapter

<!-- SOURCE: U24f_docs_archive_remaining_small.md -->
### isolated chat satellite architecture (three blast vectors: load/hack/bug)
- **category:** NEW:site-chatbot
- **status-signal:** aspirational
- **status-evidence:** "Current lean (not committed — kept open)." Explicitly a plan document with an open central decision (Option X vs Y).
- **what:** A plan to run the site-chatbot's server-side pieces (turn storage, drain/analytics, and any chat workflow code) on infrastructure **separate** from the core build cluster, so that live chat traffic, a compromise of the internet-facing edge worker, or a chat-code bug cannot degrade or reach the webdesign/build system. Deliberately **not** built on the existing multi-cluster dispatch (Phase 4a, `remote-job-spawner`), which shares cluster A's Kafka/Postgres by design — the chat satellite instead reuses only the chassis *binaries* and action code, deployed against its own Kafka/DB/storage, with a one-directional async boundary (core publishes install triggers and content; nothing on the chat side has synchronous or write access back into core). Two options are weighed: **Option X (minimal, recommended MVP)** — pack-building stays on core, the satellite is just a turn store + puller + analytics, possibly needing no Kafka/chassis at all; **Option Y (full satellite chassis)** — the whole chat pipeline including install/pack-building moves to a cut-down copy of the chassis on the satellite. A worked "building-and-hosting-as-a-service via chat" example (a customer types a domain into another site's chatbot and gets a fully built, hosted site with its own chatbot) reframes the satellite as a second, customer-facing instance of the whole platform and pushes the design toward Option Y for that specific use case.
- **sources:** docs/_archive/agent_docs/docs025_ai_chatbot_idea_uk/excellent_discussions/PLAN_isolated_chat_environment(3).md
- **relations:** multicluster (Phase 4a, the pattern explicitly rejected as a template); SaaS commercial model (below, same document, §13)
- **verify-later:** whether any satellite cluster / separate chat Postgres exists; site_chat_turns table; remote-job-spawner (the Phase 4a mechanism used as contrast)

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Isolated chat/satellite architecture ("Y-copy") for SaaS build isolation
- **category:** NEW:saas-isolation-architecture
- **status-signal:** aspirational
- **status-evidence:** "Current lean (open): Option Y-copy... not committed — kept open" (PLAN_isolated_chat_environment(5).md §5); "Nothing in this plan has been deployed or applied yet" (stripe/001commentary.md worked-example thread)
- **what:** A plan to run the site chatbot's server-side pieces (turn storage, drain/analytics, chat workflow code) on infrastructure separate from the core build cluster, decomposing "don't let chat interfere with build" into three distinct threats — load (turn-ingestion write-load competing with builds), hack (a compromised internet-facing edge worker reaching core data), bug (chat code faults degrading shared chassis/Kafka/DB) — via a strictly one-directional, async, egress-from-core-only boundary. Two sizing options were weighed: minimal Option X (turn store + puller + analytics only) vs full Option Y (a cut-down copy of the whole chassis on a separate cluster); the "current lean" is Y-copy (deploy the existing monolithic chassis image against new Kafka/Postgres/storage, curate the agent_definitions seed) as an experimentation sandbox. The plan escalated once chat was reframed as the intake front-end to a full build-as-a-service product: an anonymous, internet-triggered, token-spending build pipeline must not run on core, which rules out minimal Option X and pushes toward full-chassis Option Y as a second, customer-facing instance of the whole platform.
- **sources:** tools/tool_widget_clobber/PLAN_isolated_chat_environment(5).md#1-5,9,11-13, tools/tool_widget_clobber/PLAN_isolated_chat_environment(2).md#13, stripe/001commentary.md#worked-example section, stripe/001commentary.md#§13
- **relations:** Multi-cluster dispatch (Phase 4a) — the coupled model explicitly NOT reused; Agent-to-adapter capability maturation path; Conversational build-intake via briefing-agent chat; Operator-vs-vendor business model fork; Entitlement gate architecture
- **verify-later:** whether any satellite infrastructure has actually been stood up; `086_site_chat_turns.sql` isolated-DB variant; `TurnSink` implementation

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Conversational build-intake via briefing-agent chat
- **category:** NEW:saas-isolation-architecture
- **status-signal:** aspirational
- **status-evidence:** Described purely as a designed flow: "instead of a static briefing form, a briefing-agent conducts the intake as a conversation"
- **what:** Reframes an existing chatbot feature as the entry point to the whole build pipeline: a customer types a domain + rough spec into a chat, a `briefing-agent` conducts the intake as dialogue rather than a static form, then hands off to `intake-orchestrator` on the satellite to create the site row and kick the build. Reuses the existing build pipeline unchanged once the brief is solid; the chat drops into a "job lane" for the build duration.
- **sources:** stripe/001commentary.md#worked example section
- **relations:** Isolated chat/satellite architecture (Y-copy); New-domain build pipeline stage chain
- **verify-later:** 018_briefing_questionnaire fields; 002_intake_orchestrator entry contract

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Isolated chat/satellite architecture ("Y-copy") for SaaS build isolation
- **category:** NEW:saas-isolation-architecture
- **status-signal:** aspirational
- **status-evidence:** "Current lean (open): Option Y-copy... not committed — kept open" (PLAN_isolated_chat_environment(5).md §5); "Nothing in this plan has been deployed or applied yet" (stripe/001commentary.md worked-example thread)
- **what:** A plan to run the site chatbot's server-side pieces (turn storage, drain/analytics, chat workflow code) on infrastructure separate from the core build cluster, decomposing "don't let chat interfere with build" into three distinct threats — load (turn-ingestion write-load competing with builds), hack (a compromised internet-facing edge worker reaching core data), bug (chat code faults degrading shared chassis/Kafka/DB) — via a strictly one-directional, async, egress-from-core-only boundary. Two sizing options were weighed: minimal Option X (turn store + puller + analytics only) vs full Option Y (a cut-down copy of the whole chassis on a separate cluster); the "current lean" is Y-copy (deploy the existing monolithic chassis image against new Kafka/Postgres/storage, curate the agent_definitions seed) as an experimentation sandbox. The plan escalated once chat was reframed as the intake front-end to a full build-as-a-service product: an anonymous, internet-triggered, token-spending build pipeline must not run on core, which rules out minimal Option X and pushes toward full-chassis Option Y as a second, customer-facing instance of the whole platform.
- **sources:** tools/tool_widget_clobber/PLAN_isolated_chat_environment(5).md#1-5,9,11-13, tools/tool_widget_clobber/PLAN_isolated_chat_environment(2).md#13, stripe/001commentary.md#worked-example section, stripe/001commentary.md#§13
- **relations:** Multi-cluster dispatch (Phase 4a) — the coupled model explicitly NOT reused; Agent-to-adapter capability maturation path; Conversational build-intake via briefing-agent chat; Operator-vs-vendor business model fork; Entitlement gate architecture
- **verify-later:** whether any satellite infrastructure has actually been stood up; `086_site_chat_turns.sql` isolated-DB variant; `TurnSink` implementation

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Conversational build-intake via briefing-agent chat
- **category:** NEW:saas-isolation-architecture
- **status-signal:** aspirational
- **status-evidence:** Described purely as a designed flow: "instead of a static briefing form, a briefing-agent conducts the intake as a conversation"
- **what:** Reframes an existing chatbot feature as the entry point to the whole build pipeline: a customer types a domain + rough spec into a chat, a `briefing-agent` conducts the intake as dialogue rather than a static form, then hands off to `intake-orchestrator` on the satellite to create the site row and kick the build. Reuses the existing build pipeline unchanged once the brief is solid; the chat drops into a "job lane" for the build duration.
- **sources:** stripe/001commentary.md#worked example section
- **relations:** Isolated chat/satellite architecture (Y-copy); New-domain build pipeline stage chain
- **verify-later:** 018_briefing_questionnaire fields; 002_intake_orchestrator entry contract

<!-- SOURCE: U04_idea_uk.md -->
### Operator email identity: leopardess.uk + deterministic per-site addresses + email spec aspect
- **category:** NEW:email-infrastructure
- **status-signal:** partial
- **status-evidence:** "Status: design, not yet implemented in the chassis. idea.uk… carries these values in its env"; the identity scheme is live for idea.uk (idea-uk@leopardess.uk), the aspect/provisioner are design-only.
- **what:** One neutral operator domain (leopardess.uk — also given a one-page identity site) fronts all sites' transactional/support mail. Per-site address = deterministic encoding (lowercase, dots→dashes, @operator_domain), resolved by matching against the known-domain set, never by reversing; collisions detected at assignment and stored overrides win. A new `site_specs` aspect `email` (no DDL) carries per-site identity/status/provider; a future `email-provisioner` agent flips provisioned=false→true (same provision-and-write-back shape as model-trainer/Thunder). Refined 2026-06-06: prefer a **specific forwarder per published site** over a server catch-all (no backscatter; only forward addresses that exist).
- **sources:** idea.uk/EMAIL_identity_in_site_spec(5).md; idea.uk/idea_uk_architecture_and_deployment(6).md (2026-06-05 correction); idea.uk/leopardess_uk_index.html
- **relations:** transactional sending realities; site-spec aspect model (021); feasibility-recheck promotion mechanism.
- **verify-later:** site_specs DISTINCT aspect for 'email' (expect absent); 021 doc aspect list.

<!-- SOURCE: U04_idea_uk.md -->
### Transactional email sending realities (587-only, relay filtering, SES + per-domain DKIM)
- **category:** NEW:email-infrastructure
- **status-signal:** deployed
- **status-evidence:** "DECISIVE: MailChannels blocks leopardess.uk DIRECT outbound too → must leave MailChannels" (2026-06-11); SES live in production same day; EMAIL doc header codifies the lesson for the future provisioner.
- **what:** Hard-won operational truths now standing framework guidance: cloud boxes can't use outbound SMTP 25/465 (Hetzner leaves only 587 submission open — the cPanel UI advertising 465 misleads); Go's smtp.SendMail does STARTTLS not implicit-TLS, so a 465 path needs a tls.Dial branch; shared-host relays (Clook→MailChannels) content-filter legitimate transactional mail (a `From:`-like line + raw JSON in a body triggered "Spam Content"); therefore transactional sending needs a **dedicated sender (AWS SES eu-west-2) with the operator domain's own DKIM**, bodies kept clean, and the mailer async/bounded so a hung send can't freeze the request path. Gotcha: SES SMTP_USER is the AKIA access-key-id, not the IAM user name (535s otherwise). Chronology: Clook both-ways → catch-all/Default-Address fixes → MailChannels blocks → SES.
- **sources:** idea.uk/idea_uk_architecture_and_deployment(6).md (2026-06-05/06/10/11 updates); idea.uk/EMAIL_identity_in_site_spec(5).md (2026-06-11 header + operational note); idea.uk/running_notes(63).md (email checkpoints)
- **relations:** operator email identity; deliverable quality standards (clean bodies).
- **verify-later:** /etc/idea/idea.env SMTP block (on the box); smtpSend in service.go.

<!-- SOURCE: U04_idea_uk.md -->
### Operator email identity: leopardess.uk + deterministic per-site addresses + email spec aspect
- **category:** NEW:email-infrastructure
- **status-signal:** partial
- **status-evidence:** "Status: design, not yet implemented in the chassis. idea.uk… carries these values in its env"; the identity scheme is live for idea.uk (idea-uk@leopardess.uk), the aspect/provisioner are design-only.
- **what:** One neutral operator domain (leopardess.uk — also given a one-page identity site) fronts all sites' transactional/support mail. Per-site address = deterministic encoding (lowercase, dots→dashes, @operator_domain), resolved by matching against the known-domain set, never by reversing; collisions detected at assignment and stored overrides win. A new `site_specs` aspect `email` (no DDL) carries per-site identity/status/provider; a future `email-provisioner` agent flips provisioned=false→true (same provision-and-write-back shape as model-trainer/Thunder). Refined 2026-06-06: prefer a **specific forwarder per published site** over a server catch-all (no backscatter; only forward addresses that exist).
- **sources:** idea.uk/EMAIL_identity_in_site_spec(5).md; idea.uk/idea_uk_architecture_and_deployment(6).md (2026-06-05 correction); idea.uk/leopardess_uk_index.html
- **relations:** transactional sending realities; site-spec aspect model (021); feasibility-recheck promotion mechanism.
- **verify-later:** site_specs DISTINCT aspect for 'email' (expect absent); 021 doc aspect list.

<!-- SOURCE: U04_idea_uk.md -->
### Transactional email sending realities (587-only, relay filtering, SES + per-domain DKIM)
- **category:** NEW:email-infrastructure
- **status-signal:** deployed
- **status-evidence:** "DECISIVE: MailChannels blocks leopardess.uk DIRECT outbound too → must leave MailChannels" (2026-06-11); SES live in production same day; EMAIL doc header codifies the lesson for the future provisioner.
- **what:** Hard-won operational truths now standing framework guidance: cloud boxes can't use outbound SMTP 25/465 (Hetzner leaves only 587 submission open — the cPanel UI advertising 465 misleads); Go's smtp.SendMail does STARTTLS not implicit-TLS, so a 465 path needs a tls.Dial branch; shared-host relays (Clook→MailChannels) content-filter legitimate transactional mail (a `From:`-like line + raw JSON in a body triggered "Spam Content"); therefore transactional sending needs a **dedicated sender (AWS SES eu-west-2) with the operator domain's own DKIM**, bodies kept clean, and the mailer async/bounded so a hung send can't freeze the request path. Gotcha: SES SMTP_USER is the AKIA access-key-id, not the IAM user name (535s otherwise). Chronology: Clook both-ways → catch-all/Default-Address fixes → MailChannels blocks → SES.
- **sources:** idea.uk/idea_uk_architecture_and_deployment(6).md (2026-06-05/06/10/11 updates); idea.uk/EMAIL_identity_in_site_spec(5).md (2026-06-11 header + operational note); idea.uk/running_notes(63).md (email checkpoints)
- **relations:** operator email identity; deliverable quality standards (clean bodies).
- **verify-later:** /etc/idea/idea.env SMTP block (on the box); smtpSend in service.go.

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Admin dashboard + nginx gateway architecture
- **category:** admin-dashboard-and-api
- **status-signal:** deployed
- **status-evidence:** 012 full; 013 phases 1-11 ✅ except user portal
- **what:** React SPA served by nginx that also gateways /api/v1/auth→auth-service and /api/v1→core-manager (rate limits, timeouts, immutable asset caching, security headers). Views: Sites (lock badges), Work Items (three review flows: placeholder/checkpoint/standard; bulk retry; cross-site tab), Pages (three-level browser, Fields/HTML/Brief tabs, page-purpose bar, suppressed-section restore), Direction (spec cards, pin/propagate), Media (assets + references). Access via port-forward today; WireGuard/bastion planned.
- **sources:** 012 full; 013 status table
- **relations:** content governance edit paths; admin API endpoints
- **verify-later:** frontends/admin-dashboard; nginx conf

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Admin API current state: dual-auth gateway, inventory, and fix blocks
- **category:** admin-dashboard-and-api
- **status-signal:** partial
- **status-evidence:** P3 headings: known issues (bugs "code won't work as written", hardcoded/mock data, missing wiring) with blocks A–F and a target route map
- **what:** Two services one gateway: auth-service handles auth/user/subscription/projects directly and proxies admin site routes to core-manager; dual auth validation on both sides. The doc inventories every current route, catalogues bugs/mock data/unregistered handlers/design concerns, and sequences fixes: A fix bugs, B wire handlers, C replace hardcoded data, D performance, E new site-domain endpoints, F agent-definition admin improvements.
- **sources:** P3_admin_api_plan.md (header-scan)
- **relations:** admin dashboard; public API plan
- **verify-later:** which blocks completed (012 suggests site-domain endpoints largely live)

<!-- SOURCE: U09_adoption.md -->
### Core-manager API server surface (spec pin/unpin among admin routes)
- **category:** admin-dashboard-and-api
- **status-signal:** deployed
- **status-evidence:** old2/server.go (Core Manager API server, gin router, persona repo, kafka producer); PLAN_lock_coherence correction: "server.go routes POST /sites/:site_id/specs/:aspect/pin and .../unpin to specAdminHandlers.HandlePinSpec / HandleUnpinSpec (Phase 4 'Spec Direction Control')".
- **what:** The core-manager HTTP API (internal/core-manager) is a separate reader/writer surface from the chassis — notably exposing spec pin/unpin endpoints that keep Pattern B semantics alive even though chassis code has zero `pinned` references. Any lock-model retirement must account for admin-API consumers, not just chassis greps.
- **sources:** old2/server.go, PLAN_lock_coherence.md#pinning-verification
- **relations:** lock-model coherence step 5; site_specs supersede-then-insert writes
- **verify-later:** specAdminHandlers read/write of site_specs.pinned

<!-- SOURCE: U12_docs024_archives.md -->
### Work-item HITL model: approve/reject endpoints on pending_review status
- **category:** admin-dashboard-and-api
- **status-signal:** superseded
- **status-evidence:** `007a_public_api_plan_v1.md`: `POST /work-items/:item_id/approve|reject`; live `P2_public_api_plan.md`/`P3_admin_api_plan.md` have no approve/reject endpoints or `pending_review`/`rejected` statuses anywhere.
- **what:** The original API plan modelled human review as a binary approval gate on work items, with specs read-only initially. Replaced end-to-end by `needs_human_review` items with three resolution paths (provide missing spec data + retry, retry unchanged, or dismiss with a resolution note), and `PATCH /specs/:aspect` as a first-class, versioned write path feeding that retry flow.
- **sources:** old/older1/007a_public_api_plan_v1.md#"Work Items (build progress + HITL)"; docs024_key_docs_latest/P2_public_api_plan.md#"HITL Review Flow"
- **relations:** content-governance (locks, HITL)
- **verify-later:** grep core-manager handlers for any surviving `pending_review`/`HandleApproveWorkItem`/`HandleRejectWorkItem`.

<!-- SOURCE: U12_docs024_archives.md -->
### Admin work-item reassign + force-complete override endpoints
- **category:** admin-dashboard-and-api
- **status-signal:** superseded
- **status-evidence:** `008a_admin_api_plan_v1.md` E3 table has `reassign`/`force-complete`; live `P3_admin_api_plan.md`'s equivalent table has neither, only generic `PATCH`, `retry`, `resolve` (all Implemented).
- **what:** The original admin plan gave two narrow, single-purpose override endpoints for stuck work items: reassign the handler agent, or force-mark-complete with an arbitrary result. Generalised instead into one `PATCH` endpoint plus the shared `retry`/`resolve` pair — reassign and force-complete as distinct named actions never shipped.
- **sources:** old/older1/008a_admin_api_plan_v1.md#"E3: Work item administration"; docs024_key_docs_latest/P3_admin_api_plan.md#"E3: Work item administration + HITL review — IMPLEMENTED"
- **relations:** work-item HITL model (above)
- **verify-later:** confirm `site_admin_handlers.go` has no `HandleReassignWorkItem`/`HandleForceComplete`.

<!-- SOURCE: U12_docs024_archives.md -->
### WireGuard VPN admin-access implementation detail
- **category:** admin-dashboard-and-api
- **status-signal:** superseded
- **status-evidence:** Archive contains full runnable K8s manifests and nginx configs; live `012_admin_dashboard.md`'s condensed section keeps only one-line summaries, drops every YAML/config block.
- **what:** Three documented approaches to securely expose the admin dashboard without public ingress: (A) WireGuard-in-cluster with full K8s manifests, (B) external VM bastion with WireGuard + nginx + TLS + rate limiting, (C) plain `kubectl port-forward`. The live doc retains only the decision framework, not the deployable configuration.
- **sources:** archive_april_26/019_admin_access_infrastructure.md (whole file); docs024_key_docs_latest/012_admin_dashboard.md#"Network Access Options"
- **relations:** admin-dashboard-and-api; auth-service JWT/RequireRole security layer
- **verify-later:** check whether WireGuard was ever actually deployed or whether the system is still on Option C.

<!-- SOURCE: U24a_docs_archive_classic_and_docs024_misc.md -->
### Public REST API for the site-building pipeline
- **category:** admin-dashboard-and-api
- **status-signal:** aspirational
- **status-evidence:** 007b build-order table (2026-04) lists Blocks 0-6 (site_ownership, HandleCreateSite, pages, work-items, specs, briefing bridge, websockets) all "Not started" except the admin subset; only admin `site_admin_handlers.go` is "Written — ready to deploy".
- **what:** Plan to expose `sites`, `pages`, `site_work_items`, `site_specs`, `assets`, and briefing over `/api/v1/sites/*`, tenant-scoped via a new `site_ownership` junction table, so users can submit domains, watch build progress, and resolve HITL reviews over HTTP. Reads/writes the same DB the agents use (build_queue, dispatch loop pick changes up); Kafka only touched for the briefing HTTP→Kafka bridge. The user-scoped public half was never built.
- **sources:** archive_april_26/007b_public_api_plan_v2.md#public-api-endpoints, #ownership-model, #build-order
- **relations:** depends on site_ownership; complements Admin API (008b); superseded reference by live docs/api/reference.html
- **verify-later:** internal/core-manager/sites/*.go (planned, may not exist); site_ownership migration

<!-- SOURCE: U24a_docs_archive_classic_and_docs024_misc.md -->
### site_ownership table / ownership model
- **category:** admin-dashboard-and-api
- **status-signal:** abandoned
- **status-evidence:** 007b: "The `sites` table has no ownership columns"; site_ownership migration listed Block 0 "Not started"; 008b Block E notes admin endpoints "work without the site_ownership migration because they're admin-only".
- **what:** Proposed `site_ownership(site_id, client_id, user_id, role)` junction table to attach user identity to agent-created sites (which carry none), enabling per-user scoping of the public API. Chosen as a junction (not columns on `sites`) because sites can be shared and `sites` has 15+ FKs. Never created; admin API sidesteps it.
- **sources:** archive_april_26/007b_public_api_plan_v2.md#ownership-model; 008b_admin_api_plan_v2.md#block-e
- **relations:** blocks the public API; `assign`/`trigger-build` admin endpoints
- **verify-later:** grep for site_ownership in migrations and core-manager

<!-- SOURCE: U24a_docs_archive_classic_and_docs024_misc.md -->
### Admin API current-state audit (bugs, hardcoded data, blocks)
- **category:** admin-dashboard-and-api
- **status-signal:** partial
- **status-evidence:** 008b §3 lists live bugs (B1 MySQL syntax `CURDATE()` in dashboard, B2 `orchestrator_state` vs `orchestration_states`, B3 cartesian join, B4 missing agent-instances proxy) and H1-H7 hardcoded/mock values (Kafka topics, agent health, usage metrics); build order marks A-D "Not started", E3 "Implemented".
- **what:** Full inventory of the two-service gateway admin API (auth-service handles auth/user/subscription/projects directly; core-manager handles templates/personas/clients/system/workflows/agents via `.Any()` proxy). Documents JWT dual-validation, real vs hardcoded endpoints, and a repair plan. The site/work-item HITL admin subset (Block E3) is deployed; Kafka-topic listing, dashboard aggregation, and agent-health remain mock.
- **sources:** archive_april_26/008b_admin_api_plan_v2.md#current-admin-endpoints, #known-issues, #implementation-plan
- **relations:** gateway proxy pattern; HITL review flow; site admin handlers
- **verify-later:** internal/core-manager/admin/{dashboard_handlers,system_handlers,agent_handlers,site_admin_handlers}.go

<!-- SOURCE: U24a_docs_archive_classic_and_docs024_misc.md -->
### React admin dashboard for build review
- **category:** admin-dashboard-and-api
- **status-signal:** partial
- **status-evidence:** 007b: `site-admin-dashboard.jsx` "Written — uses mock data until API connected (toggle `useMock=false`)", Tailwind utility classes; three views (Dashboard, Review Queue, Review Detail).
- **what:** A React frontend rendering site cards with progress bars, a `needs_human_review` queue with Review/Retry/Dismiss, and a review-detail view with an editable identity-spec JSON + "Save Spec & Retry". Runs on mock data pending API wiring.
- **sources:** archive_april_26/007b_public_api_plan_v2.md#react-admin-dashboard; 008b#files-summary
- **relations:** Admin API Block E; HITL review flow
- **verify-later:** site-admin-dashboard.jsx location

<!-- SOURCE: U26_misc_dirs.md -->
### AI Persona Platform public API
- **category:** admin-dashboard-and-api
- **status-signal:** superseded
- **status-evidence:** docs/api/reference.html is a generated Redoc bundle titled "AI Persona Platform API" covering the persona-era surface; the current API surface is the admin dashboard/nginx gateway (spine 012) and the persona-instance concepts do not appear in current docs.
- **what:** The v1 REST surface of the persona era: JWT auth (register/login/refresh/validate/logout), user profile/password/delete, projects CRUD, subscription with usage stats and quota checks, persona template listing, persona instance list/create, health check, and a WebSocket connection endpoint. Documents the original productisation of the platform as "AI personas" for end users.
- **sources:** docs/api/reference.html (tags: Authentication, Users, Projects, Subscriptions, Templates, Instances, System, WebSocket; paths /api/v1/auth/*, /api/v1/projects, /api/v1/subscription/*, /api/v1/personas/instances)
- **relations:** three-database architecture (auth DB backs these endpoints); superseded by current admin-dashboard-and-api (012)
- **verify-later:** which endpoints survive in core-manager/api-gateway code

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Admin dashboard + nginx gateway architecture
- **category:** admin-dashboard-and-api
- **status-signal:** deployed
- **status-evidence:** 012 full; 013 phases 1-11 ✅ except user portal
- **what:** React SPA served by nginx that also gateways /api/v1/auth→auth-service and /api/v1→core-manager (rate limits, timeouts, immutable asset caching, security headers). Views: Sites (lock badges), Work Items (three review flows: placeholder/checkpoint/standard; bulk retry; cross-site tab), Pages (three-level browser, Fields/HTML/Brief tabs, page-purpose bar, suppressed-section restore), Direction (spec cards, pin/propagate), Media (assets + references). Access via port-forward today; WireGuard/bastion planned.
- **sources:** 012 full; 013 status table
- **relations:** content governance edit paths; admin API endpoints
- **verify-later:** frontends/admin-dashboard; nginx conf

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Admin API current state: dual-auth gateway, inventory, and fix blocks
- **category:** admin-dashboard-and-api
- **status-signal:** partial
- **status-evidence:** P3 headings: known issues (bugs "code won't work as written", hardcoded/mock data, missing wiring) with blocks A–F and a target route map
- **what:** Two services one gateway: auth-service handles auth/user/subscription/projects directly and proxies admin site routes to core-manager; dual auth validation on both sides. The doc inventories every current route, catalogues bugs/mock data/unregistered handlers/design concerns, and sequences fixes: A fix bugs, B wire handlers, C replace hardcoded data, D performance, E new site-domain endpoints, F agent-definition admin improvements.
- **sources:** P3_admin_api_plan.md (header-scan)
- **relations:** admin dashboard; public API plan
- **verify-later:** which blocks completed (012 suggests site-domain endpoints largely live)

<!-- SOURCE: U09_adoption.md -->
### Core-manager API server surface (spec pin/unpin among admin routes)
- **category:** admin-dashboard-and-api
- **status-signal:** deployed
- **status-evidence:** old2/server.go (Core Manager API server, gin router, persona repo, kafka producer); PLAN_lock_coherence correction: "server.go routes POST /sites/:site_id/specs/:aspect/pin and .../unpin to specAdminHandlers.HandlePinSpec / HandleUnpinSpec (Phase 4 'Spec Direction Control')".
- **what:** The core-manager HTTP API (internal/core-manager) is a separate reader/writer surface from the chassis — notably exposing spec pin/unpin endpoints that keep Pattern B semantics alive even though chassis code has zero `pinned` references. Any lock-model retirement must account for admin-API consumers, not just chassis greps.
- **sources:** old2/server.go, PLAN_lock_coherence.md#pinning-verification
- **relations:** lock-model coherence step 5; site_specs supersede-then-insert writes
- **verify-later:** specAdminHandlers read/write of site_specs.pinned

<!-- SOURCE: U12_docs024_archives.md -->
### Work-item HITL model: approve/reject endpoints on pending_review status
- **category:** admin-dashboard-and-api
- **status-signal:** superseded
- **status-evidence:** `007a_public_api_plan_v1.md`: `POST /work-items/:item_id/approve|reject`; live `P2_public_api_plan.md`/`P3_admin_api_plan.md` have no approve/reject endpoints or `pending_review`/`rejected` statuses anywhere.
- **what:** The original API plan modelled human review as a binary approval gate on work items, with specs read-only initially. Replaced end-to-end by `needs_human_review` items with three resolution paths (provide missing spec data + retry, retry unchanged, or dismiss with a resolution note), and `PATCH /specs/:aspect` as a first-class, versioned write path feeding that retry flow.
- **sources:** old/older1/007a_public_api_plan_v1.md#"Work Items (build progress + HITL)"; docs024_key_docs_latest/P2_public_api_plan.md#"HITL Review Flow"
- **relations:** content-governance (locks, HITL)
- **verify-later:** grep core-manager handlers for any surviving `pending_review`/`HandleApproveWorkItem`/`HandleRejectWorkItem`.

<!-- SOURCE: U12_docs024_archives.md -->
### Admin work-item reassign + force-complete override endpoints
- **category:** admin-dashboard-and-api
- **status-signal:** superseded
- **status-evidence:** `008a_admin_api_plan_v1.md` E3 table has `reassign`/`force-complete`; live `P3_admin_api_plan.md`'s equivalent table has neither, only generic `PATCH`, `retry`, `resolve` (all Implemented).
- **what:** The original admin plan gave two narrow, single-purpose override endpoints for stuck work items: reassign the handler agent, or force-mark-complete with an arbitrary result. Generalised instead into one `PATCH` endpoint plus the shared `retry`/`resolve` pair — reassign and force-complete as distinct named actions never shipped.
- **sources:** old/older1/008a_admin_api_plan_v1.md#"E3: Work item administration"; docs024_key_docs_latest/P3_admin_api_plan.md#"E3: Work item administration + HITL review — IMPLEMENTED"
- **relations:** work-item HITL model (above)
- **verify-later:** confirm `site_admin_handlers.go` has no `HandleReassignWorkItem`/`HandleForceComplete`.

<!-- SOURCE: U12_docs024_archives.md -->
### WireGuard VPN admin-access implementation detail
- **category:** admin-dashboard-and-api
- **status-signal:** superseded
- **status-evidence:** Archive contains full runnable K8s manifests and nginx configs; live `012_admin_dashboard.md`'s condensed section keeps only one-line summaries, drops every YAML/config block.
- **what:** Three documented approaches to securely expose the admin dashboard without public ingress: (A) WireGuard-in-cluster with full K8s manifests, (B) external VM bastion with WireGuard + nginx + TLS + rate limiting, (C) plain `kubectl port-forward`. The live doc retains only the decision framework, not the deployable configuration.
- **sources:** archive_april_26/019_admin_access_infrastructure.md (whole file); docs024_key_docs_latest/012_admin_dashboard.md#"Network Access Options"
- **relations:** admin-dashboard-and-api; auth-service JWT/RequireRole security layer
- **verify-later:** check whether WireGuard was ever actually deployed or whether the system is still on Option C.

<!-- SOURCE: U24a_docs_archive_classic_and_docs024_misc.md -->
### Public REST API for the site-building pipeline
- **category:** admin-dashboard-and-api
- **status-signal:** aspirational
- **status-evidence:** 007b build-order table (2026-04) lists Blocks 0-6 (site_ownership, HandleCreateSite, pages, work-items, specs, briefing bridge, websockets) all "Not started" except the admin subset; only admin `site_admin_handlers.go` is "Written — ready to deploy".
- **what:** Plan to expose `sites`, `pages`, `site_work_items`, `site_specs`, `assets`, and briefing over `/api/v1/sites/*`, tenant-scoped via a new `site_ownership` junction table, so users can submit domains, watch build progress, and resolve HITL reviews over HTTP. Reads/writes the same DB the agents use (build_queue, dispatch loop pick changes up); Kafka only touched for the briefing HTTP→Kafka bridge. The user-scoped public half was never built.
- **sources:** archive_april_26/007b_public_api_plan_v2.md#public-api-endpoints, #ownership-model, #build-order
- **relations:** depends on site_ownership; complements Admin API (008b); superseded reference by live docs/api/reference.html
- **verify-later:** internal/core-manager/sites/*.go (planned, may not exist); site_ownership migration

<!-- SOURCE: U24a_docs_archive_classic_and_docs024_misc.md -->
### site_ownership table / ownership model
- **category:** admin-dashboard-and-api
- **status-signal:** abandoned
- **status-evidence:** 007b: "The `sites` table has no ownership columns"; site_ownership migration listed Block 0 "Not started"; 008b Block E notes admin endpoints "work without the site_ownership migration because they're admin-only".
- **what:** Proposed `site_ownership(site_id, client_id, user_id, role)` junction table to attach user identity to agent-created sites (which carry none), enabling per-user scoping of the public API. Chosen as a junction (not columns on `sites`) because sites can be shared and `sites` has 15+ FKs. Never created; admin API sidesteps it.
- **sources:** archive_april_26/007b_public_api_plan_v2.md#ownership-model; 008b_admin_api_plan_v2.md#block-e
- **relations:** blocks the public API; `assign`/`trigger-build` admin endpoints
- **verify-later:** grep for site_ownership in migrations and core-manager

<!-- SOURCE: U24a_docs_archive_classic_and_docs024_misc.md -->
### Admin API current-state audit (bugs, hardcoded data, blocks)
- **category:** admin-dashboard-and-api
- **status-signal:** partial
- **status-evidence:** 008b §3 lists live bugs (B1 MySQL syntax `CURDATE()` in dashboard, B2 `orchestrator_state` vs `orchestration_states`, B3 cartesian join, B4 missing agent-instances proxy) and H1-H7 hardcoded/mock values (Kafka topics, agent health, usage metrics); build order marks A-D "Not started", E3 "Implemented".
- **what:** Full inventory of the two-service gateway admin API (auth-service handles auth/user/subscription/projects directly; core-manager handles templates/personas/clients/system/workflows/agents via `.Any()` proxy). Documents JWT dual-validation, real vs hardcoded endpoints, and a repair plan. The site/work-item HITL admin subset (Block E3) is deployed; Kafka-topic listing, dashboard aggregation, and agent-health remain mock.
- **sources:** archive_april_26/008b_admin_api_plan_v2.md#current-admin-endpoints, #known-issues, #implementation-plan
- **relations:** gateway proxy pattern; HITL review flow; site admin handlers
- **verify-later:** internal/core-manager/admin/{dashboard_handlers,system_handlers,agent_handlers,site_admin_handlers}.go

<!-- SOURCE: U24a_docs_archive_classic_and_docs024_misc.md -->
### React admin dashboard for build review
- **category:** admin-dashboard-and-api
- **status-signal:** partial
- **status-evidence:** 007b: `site-admin-dashboard.jsx` "Written — uses mock data until API connected (toggle `useMock=false`)", Tailwind utility classes; three views (Dashboard, Review Queue, Review Detail).
- **what:** A React frontend rendering site cards with progress bars, a `needs_human_review` queue with Review/Retry/Dismiss, and a review-detail view with an editable identity-spec JSON + "Save Spec & Retry". Runs on mock data pending API wiring.
- **sources:** archive_april_26/007b_public_api_plan_v2.md#react-admin-dashboard; 008b#files-summary
- **relations:** Admin API Block E; HITL review flow
- **verify-later:** site-admin-dashboard.jsx location

<!-- SOURCE: U26_misc_dirs.md -->
### AI Persona Platform public API
- **category:** admin-dashboard-and-api
- **status-signal:** superseded
- **status-evidence:** docs/api/reference.html is a generated Redoc bundle titled "AI Persona Platform API" covering the persona-era surface; the current API surface is the admin dashboard/nginx gateway (spine 012) and the persona-instance concepts do not appear in current docs.
- **what:** The v1 REST surface of the persona era: JWT auth (register/login/refresh/validate/logout), user profile/password/delete, projects CRUD, subscription with usage stats and quota checks, persona template listing, persona instance list/create, health check, and a WebSocket connection endpoint. Documents the original productisation of the platform as "AI personas" for end users.
- **sources:** docs/api/reference.html (tags: Authentication, Users, Projects, Subscriptions, Templates, Instances, System, WebSocket; paths /api/v1/auth/*, /api/v1/projects, /api/v1/subscription/*, /api/v1/personas/instances)
- **relations:** three-database architecture (auth DB backs these endpoints); superseded by current admin-dashboard-and-api (012)
- **verify-later:** which endpoints survive in core-manager/api-gateway code

<!-- SOURCE: U14_docs019_runbooks.md -->
### vertical-exemplar-researcher — the exemplar-research relay hop
- **category:** research-agents
- **status-signal:** deployed
- **status-evidence:** builder_route(21) "§B4 CLOSED — QUALITY VERIFIED 2026-07-06 … ✔ CONSUMPTION PROVEN: the strategy's gap_opportunity QUOTES the hop … ✔ TRANSMISSION THREE HOPS DEEP".
- **what:** The first new build of the builder route: a reuse-only agent (one DB row, zero new Go) inserted as needs_vertical_research between classifier and strategist. Twelve-step workflow: read specs → LLM exemplar selection (3 of the vertical's best sites, flat keys, own domain forbidden) → 3× shallow firecrawl + format → synthesis LLM (per-exemplar success factors, cross-exemplar patterns, adopt/adapt/avoid lessons, differentiation opportunity — REASONS NOT COPIES) → write_site_spec aspect=vertical_landscape → chain needs_strategy. Verified end-to-end on dartsonline: real vertical leaders selected, causal synthesis, quoted by the strategy, differentiator surfaced in the plan. Design calls: shallow-many vs adoption's deep-one; specs-not-messages; strategist prompt nudge so the research is read.
- **sources:** docs019/RUNBOOK_builder_route(21).md#B4 (design calls, change-set, re-run verification)
- **relations:** work-item relay spine; adoption fidelity; coverage baseline (curated-list reuse candidate); image_tag trap (its spawn incident)
- **verify-later:** NNN_seed_vertical_exemplar_researcher.sql; vertical_landscape spec rows; reroute migration chain_config

<!-- SOURCE: U15_docs019_running_notes.md -->
### Vertical-exemplar-researcher / competitor synthesis hop
- **category:** research-agents
- **status-signal:** deployed
- **status-evidence:** "§B4 CLOSED on quality... Landscape verified: three real vertical leaders; causal synthesis (reasons not copies); confidence 0.82. Strategy QUOTES the hop and builds the moat on it" (NOTES_running_synthesis_v4(39).md, 2026-07-06).
- **what:** A new relay hop (`needs_vertical_research` → `vertical-exemplar-researcher`) inserted between the domain classifier and the strategist to close a gap where the classifier captured `competitors_found` names but nothing ever researched them: it runs shallow crawls of 3 vertical exemplars (vs. adoption's one deep crawl of the site itself), synthesises causal reasons (not copied content) into a `site_specs` row (`aspect=vertical_landscape`), which the strategist prompt reads wholesale and demonstrably used to shape a real site's differentiator. Its first live deployment stalled because the seed migration copied `agent_definitions` columns from a donor missing the spawn-consumed `command`/`image_tag` columns (defaulted to the stale `latest` image tag) — fixed by copying from a fresher donor and flagged as a recurring `image_tag` default-value trap.
- **sources:** NOTES_running_synthesis_v4(39).md 2026-07-04 through 07-06 (§B4 sequence, full).
- **relations:** Work-item relay / builder-generations architecture; roadmap-phase enforcement gap.

<!-- SOURCE: U16_docs019_design_plans.md -->
### Vertical-exemplar research hop (best-of-niche synthesis into vertical_landscape)
- **category:** research-agents
- **status-signal:** deployed
- **status-evidence:** HANDOFF_builder_thread: "§B4 vertical-exemplar-researcher LIVE and quality-verified end to end on dartsonline.com … causal synthesis (confidence 0.82) → strategy QUOTING the landscape".
- **what:** A new relay hop between classification and strategy: find the vertical's best existing sites, read three of them shallowly (deliberate budget: limit 6, markdown, main-content only, depth 1 — vs adoption's one-site-deep 30/rawHtml/4), and distil WHY they succeed — reasons, not copies — into spec aspect vertical_landscape for the strategist and planner. Reuse-only (every step an existing action; the whole agent is one DB row, no Go, no image build); written as a spec because specs are the per-site shared memory across hops; inserted via reroute (classifier chains needs_vertical_research; researcher chains needs_strategy onward; priority 7 below strategy's 8 in the ascending ladder); an optional strategist prompt nudge makes the strategy step weigh the new aspect (research nobody reads is wasted). First bare-domain→deployed-site milestone followed.
- **sources:** README_flows.md (the plain-language explainer); NNN_seed_vertical_exemplar_researcher(2).sql; NNN_reroute_classifier_to_vertical_research.sql; NNN_strategist_vertical_landscape_nudge.sql; HANDOFF_builder_thread.md#2
- **relations:** relay spine; adoption pipeline (contrasting crawl budget); site-spec-and-classifier
- **verify-later:** vertical-exemplar-researcher row; vertical_landscape aspects in site_specs

<!-- SOURCE: U18_sql_for_agents.md -->
### research-agent (cited web research into research_results)
- **category:** research-agents
- **status-signal:** deployed
- **status-evidence:** Defined active v1.0.575 (v2/024); idle timeout set in 075; classifier v2 (003) depends on it.
- **what:** Web-search research specialist that extracts relevant quotes, synthesises findings with full source attribution and stores in a research_results table for citation ([0], [1] markers consumed by page-content-writer prompts).
- **sources:** sql_for_agents_v2/024_research_agent.sql; 024_research_agent.sql; 003_site_classifier.sql
- **relations:** spawned by page-content-writer and site-classifier v2
- **verify-later:** research_results table usage

<!-- SOURCE: U19_sql_tables_components.md -->
### research_results with source attribution
- **category:** research-agents
- **status-signal:** deployed
- **status-evidence:** Table created in 004 PART 5 with sources JSONB format (url, title, domain, accessed_at, quotes, relevance_score); 009 patches add result_type and data/findings columns the code expects; training exports read result_type='tool_recreation_training'.
- **what:** Research findings persisted per site/page/component with full source attribution and expiry (expires_at refresh signal); page_components.research_id links content to the research that informed it, with sources_displayed controlling on-page attribution. Also doubles as generic typed result storage (result_type) e.g. tool recreation training triples.
- **sources:** docs/agent_docs/sql_for_components/004_component_architecture_schema.sql#PART5; docs/agent_docs/sql_for_tables/009_research_results.sql; docs/agent_docs/sql_for_tables/025_llm_call_log_rag_knowledge_base.sql#exports
- **relations:** content grounding; finetuning flywheel (training triples); content_items origin_research_id.
- **verify-later:** research-agent writers; result_type vocabulary.

<!-- SOURCE: U21_legacy_docs_b.md -->
### Research agent with cited sources
- **category:** research-agents
- **status-signal:** deployed
- **status-evidence:** docs015/001 step-by-step verified pipeline (extract_topic → build_search_query → web_search → prepare_urls → batch_webscrape → format_research_content → synthesize → insert_research_result); docs012/010 principle "Research is cited — all LLM-generated content must cite sources, which are stored."
- **what:** A self-contained research agent: composes a search query from raw inputs, searches, selects top URLs, batch-scrapes them via the webscrape adapter, formats findings with snippet context, synthesizes a JSON summary (key points, recommendations, confidence), and persists to research_results with full source list — returning a research_id that content sections reference. Backing store for research-driven components (FAQ, long-form).
- **sources:** docs015_data_flow_verification/001_data_flow_verification.md; docs012_site_maps_and_components/012_summary_of_all_before_this_in_this_folder.md#research-agent; docs017_legacy_agent_rules_images_design_keydocs/019b_agent_architecture_v5_with_tickets_news.md#Active-Agents
- **relations:** render_mode needs_research; batch_webscrape action; adapters (webscrape); current research-agents category.
- **verify-later:** research_results table; prepare_urls/batch_webscrape/format_research_content in registry.

<!-- SOURCE: U22_recent_small_docs.md -->
### Chat differentiator ideation agent
- **category:** research-agents
- **status-signal:** aspirational
- **status-evidence:** "A low-risk, internal use of the agent framework ... Also still needs work — treat the output as candidates, not commitments."
- **what:** A proposed internal agent (runs on our own data, no isolation concerns) that, given a domain + audience, runs the asset × AI-capability combination and proposes ranked candidate payable differentiators split into "test now (cheap)" vs "score/consider (expensive)", each naming the asset and capability it depends on. Can spawn sub-agents to research willingness-to-pay or check whether a data feed exists/what it costs. Re-runnable across all domains whenever a new AI capability is added — the mechanism for catching early-adopter opportunities. Idea generator feeding human judgement, not an automated builder.
- **sources:** docs025.../PLAN_simple_paid_multidomain_chat(1).md#11
- **relations:** payable-differentiator framework, research-agents
- **verify-later:** any ideation/differentiator agent definition

<!-- SOURCE: U24c_docs_archive_traffic_probe.md -->
### Wayback/archive.org grounding method + limitation
- **category:** research-agents
- **status-signal:** partial
- **status-evidence:** running_notes 2026-06-13(b) "archive.org: Claude CAN web_fetch archive pages but ONLY when a search surfaces the exact URL; canNOT enumerate CDX on demand and the sandbox can't reach archive.org directly".
- **what:** Each probe page is grounded in the domain's old vertical via a Wayback snapshot. Constraint: the sandbox can't reach archive.org directly and can't enumerate CDX on demand, so grounding a NEW domain requires the operator to supply the Wayback URL/snapshot, or Claude uses web search + the domain name.
- **sources:** traffic_probe_running_notes(27).md#2026-06-13-b, traffic_probe_runbook.md#3
- **relations:** feeds intent-probe page content
- **verify-later:** archive.org.results/{relojistas,wayfaringlondoner}

<!-- SOURCE: U24e_docs_archive_idea_uk.md -->
### Capability watchlist + real-world event watchlist (dual standing research workflows)
- **category:** research-agents
- **status-signal:** aspirational
- **status-evidence:** `running_notes(44).md`: "The capability list in the method is a starter; the watchlist workflow itself isn't designed" (open thread, never closed within this file); later: "Real-world event watchlist promoted to a second standing workflow... Both are recurring research workflows that fire re-runs of ideation."
- **what:** Two proposed recurring background research workflows: (1) a capability watchlist tracking new AI capabilities that beat the model's self-knowledge (agentic browsing, million-token contexts, real-time voice, etc.) — the "early-adopter mechanism"; (2) an event/window watchlist tracking scheme deadlines, regulation changes and application windows per domain (proven by the agritec SFI26 Window 1 case, which turned a "consider later" candidate into "test now"). Both are meant to trigger automatic re-runs of the ideation method across domains, but the trigger mechanism itself was never designed/built within this archive's timeframe.
- **sources:** `running_notes(44).md` ("Capability watchlist warrants its own workflow", "Watchlist should track scheme/event windows, not just AI capabilities", "Real-world event watchlist promoted to a second standing workflow")
- **relations:** idea generation method
- **verify-later:** whether any scheduled_task / agent implements this in the live chassis

<!-- SOURCE: U26_misc_dirs.md -->
### Deep-research domain insight agent
- **category:** research-agents
- **status-signal:** abandoned
- **status-evidence:** 016 designs a "domain-insight-agent" deciding when deep social research pays ("Value Multiple: 50-100x"); tied to the abandoned domain-flipping context, though its research-orchestration DNA resembles the later research agents.
- **what:** A strategic classifier that assesses whether a domain/topic merits multi-platform deep research (Reddit/LinkedIn/Twitter/Facebook/YouTube community mining, influencer mapping, sentiment threading) versus standard development, then deploys the appropriate research agent squad to synthesise unique content, tools and FAQs from real community pain points — the claimed competitive moat over single-LLM or SEO-tool approaches.
- **sources:** docs/architecture/016-competitive-advantge.md#enhanced-domain-analysis-agent; docs/architecture/016-competitive-advantge.md#deep-research-workflows-by-domain-type
- **relations:** domain value maximisation; topic amplifier engine; current research-agents lineage
- **verify-later:** n/a

<!-- SOURCE: U14_docs019_runbooks.md -->
### vertical-exemplar-researcher — the exemplar-research relay hop
- **category:** research-agents
- **status-signal:** deployed
- **status-evidence:** builder_route(21) "§B4 CLOSED — QUALITY VERIFIED 2026-07-06 … ✔ CONSUMPTION PROVEN: the strategy's gap_opportunity QUOTES the hop … ✔ TRANSMISSION THREE HOPS DEEP".
- **what:** The first new build of the builder route: a reuse-only agent (one DB row, zero new Go) inserted as needs_vertical_research between classifier and strategist. Twelve-step workflow: read specs → LLM exemplar selection (3 of the vertical's best sites, flat keys, own domain forbidden) → 3× shallow firecrawl + format → synthesis LLM (per-exemplar success factors, cross-exemplar patterns, adopt/adapt/avoid lessons, differentiation opportunity — REASONS NOT COPIES) → write_site_spec aspect=vertical_landscape → chain needs_strategy. Verified end-to-end on dartsonline: real vertical leaders selected, causal synthesis, quoted by the strategy, differentiator surfaced in the plan. Design calls: shallow-many vs adoption's deep-one; specs-not-messages; strategist prompt nudge so the research is read.
- **sources:** docs019/RUNBOOK_builder_route(21).md#B4 (design calls, change-set, re-run verification)
- **relations:** work-item relay spine; adoption fidelity; coverage baseline (curated-list reuse candidate); image_tag trap (its spawn incident)
- **verify-later:** NNN_seed_vertical_exemplar_researcher.sql; vertical_landscape spec rows; reroute migration chain_config

<!-- SOURCE: U15_docs019_running_notes.md -->
### Vertical-exemplar-researcher / competitor synthesis hop
- **category:** research-agents
- **status-signal:** deployed
- **status-evidence:** "§B4 CLOSED on quality... Landscape verified: three real vertical leaders; causal synthesis (reasons not copies); confidence 0.82. Strategy QUOTES the hop and builds the moat on it" (NOTES_running_synthesis_v4(39).md, 2026-07-06).
- **what:** A new relay hop (`needs_vertical_research` → `vertical-exemplar-researcher`) inserted between the domain classifier and the strategist to close a gap where the classifier captured `competitors_found` names but nothing ever researched them: it runs shallow crawls of 3 vertical exemplars (vs. adoption's one deep crawl of the site itself), synthesises causal reasons (not copied content) into a `site_specs` row (`aspect=vertical_landscape`), which the strategist prompt reads wholesale and demonstrably used to shape a real site's differentiator. Its first live deployment stalled because the seed migration copied `agent_definitions` columns from a donor missing the spawn-consumed `command`/`image_tag` columns (defaulted to the stale `latest` image tag) — fixed by copying from a fresher donor and flagged as a recurring `image_tag` default-value trap.
- **sources:** NOTES_running_synthesis_v4(39).md 2026-07-04 through 07-06 (§B4 sequence, full).
- **relations:** Work-item relay / builder-generations architecture; roadmap-phase enforcement gap.

<!-- SOURCE: U16_docs019_design_plans.md -->
### Vertical-exemplar research hop (best-of-niche synthesis into vertical_landscape)
- **category:** research-agents
- **status-signal:** deployed
- **status-evidence:** HANDOFF_builder_thread: "§B4 vertical-exemplar-researcher LIVE and quality-verified end to end on dartsonline.com … causal synthesis (confidence 0.82) → strategy QUOTING the landscape".
- **what:** A new relay hop between classification and strategy: find the vertical's best existing sites, read three of them shallowly (deliberate budget: limit 6, markdown, main-content only, depth 1 — vs adoption's one-site-deep 30/rawHtml/4), and distil WHY they succeed — reasons, not copies — into spec aspect vertical_landscape for the strategist and planner. Reuse-only (every step an existing action; the whole agent is one DB row, no Go, no image build); written as a spec because specs are the per-site shared memory across hops; inserted via reroute (classifier chains needs_vertical_research; researcher chains needs_strategy onward; priority 7 below strategy's 8 in the ascending ladder); an optional strategist prompt nudge makes the strategy step weigh the new aspect (research nobody reads is wasted). First bare-domain→deployed-site milestone followed.
- **sources:** README_flows.md (the plain-language explainer); NNN_seed_vertical_exemplar_researcher(2).sql; NNN_reroute_classifier_to_vertical_research.sql; NNN_strategist_vertical_landscape_nudge.sql; HANDOFF_builder_thread.md#2
- **relations:** relay spine; adoption pipeline (contrasting crawl budget); site-spec-and-classifier
- **verify-later:** vertical-exemplar-researcher row; vertical_landscape aspects in site_specs

<!-- SOURCE: U18_sql_for_agents.md -->
### research-agent (cited web research into research_results)
- **category:** research-agents
- **status-signal:** deployed
- **status-evidence:** Defined active v1.0.575 (v2/024); idle timeout set in 075; classifier v2 (003) depends on it.
- **what:** Web-search research specialist that extracts relevant quotes, synthesises findings with full source attribution and stores in a research_results table for citation ([0], [1] markers consumed by page-content-writer prompts).
- **sources:** sql_for_agents_v2/024_research_agent.sql; 024_research_agent.sql; 003_site_classifier.sql
- **relations:** spawned by page-content-writer and site-classifier v2
- **verify-later:** research_results table usage

<!-- SOURCE: U19_sql_tables_components.md -->
### research_results with source attribution
- **category:** research-agents
- **status-signal:** deployed
- **status-evidence:** Table created in 004 PART 5 with sources JSONB format (url, title, domain, accessed_at, quotes, relevance_score); 009 patches add result_type and data/findings columns the code expects; training exports read result_type='tool_recreation_training'.
- **what:** Research findings persisted per site/page/component with full source attribution and expiry (expires_at refresh signal); page_components.research_id links content to the research that informed it, with sources_displayed controlling on-page attribution. Also doubles as generic typed result storage (result_type) e.g. tool recreation training triples.
- **sources:** docs/agent_docs/sql_for_components/004_component_architecture_schema.sql#PART5; docs/agent_docs/sql_for_tables/009_research_results.sql; docs/agent_docs/sql_for_tables/025_llm_call_log_rag_knowledge_base.sql#exports
- **relations:** content grounding; finetuning flywheel (training triples); content_items origin_research_id.
- **verify-later:** research-agent writers; result_type vocabulary.

<!-- SOURCE: U21_legacy_docs_b.md -->
### Research agent with cited sources
- **category:** research-agents
- **status-signal:** deployed
- **status-evidence:** docs015/001 step-by-step verified pipeline (extract_topic → build_search_query → web_search → prepare_urls → batch_webscrape → format_research_content → synthesize → insert_research_result); docs012/010 principle "Research is cited — all LLM-generated content must cite sources, which are stored."
- **what:** A self-contained research agent: composes a search query from raw inputs, searches, selects top URLs, batch-scrapes them via the webscrape adapter, formats findings with snippet context, synthesizes a JSON summary (key points, recommendations, confidence), and persists to research_results with full source list — returning a research_id that content sections reference. Backing store for research-driven components (FAQ, long-form).
- **sources:** docs015_data_flow_verification/001_data_flow_verification.md; docs012_site_maps_and_components/012_summary_of_all_before_this_in_this_folder.md#research-agent; docs017_legacy_agent_rules_images_design_keydocs/019b_agent_architecture_v5_with_tickets_news.md#Active-Agents
- **relations:** render_mode needs_research; batch_webscrape action; adapters (webscrape); current research-agents category.
- **verify-later:** research_results table; prepare_urls/batch_webscrape/format_research_content in registry.

<!-- SOURCE: U22_recent_small_docs.md -->
### Chat differentiator ideation agent
- **category:** research-agents
- **status-signal:** aspirational
- **status-evidence:** "A low-risk, internal use of the agent framework ... Also still needs work — treat the output as candidates, not commitments."
- **what:** A proposed internal agent (runs on our own data, no isolation concerns) that, given a domain + audience, runs the asset × AI-capability combination and proposes ranked candidate payable differentiators split into "test now (cheap)" vs "score/consider (expensive)", each naming the asset and capability it depends on. Can spawn sub-agents to research willingness-to-pay or check whether a data feed exists/what it costs. Re-runnable across all domains whenever a new AI capability is added — the mechanism for catching early-adopter opportunities. Idea generator feeding human judgement, not an automated builder.
- **sources:** docs025.../PLAN_simple_paid_multidomain_chat(1).md#11
- **relations:** payable-differentiator framework, research-agents
- **verify-later:** any ideation/differentiator agent definition

<!-- SOURCE: U24c_docs_archive_traffic_probe.md -->
### Wayback/archive.org grounding method + limitation
- **category:** research-agents
- **status-signal:** partial
- **status-evidence:** running_notes 2026-06-13(b) "archive.org: Claude CAN web_fetch archive pages but ONLY when a search surfaces the exact URL; canNOT enumerate CDX on demand and the sandbox can't reach archive.org directly".
- **what:** Each probe page is grounded in the domain's old vertical via a Wayback snapshot. Constraint: the sandbox can't reach archive.org directly and can't enumerate CDX on demand, so grounding a NEW domain requires the operator to supply the Wayback URL/snapshot, or Claude uses web search + the domain name.
- **sources:** traffic_probe_running_notes(27).md#2026-06-13-b, traffic_probe_runbook.md#3
- **relations:** feeds intent-probe page content
- **verify-later:** archive.org.results/{relojistas,wayfaringlondoner}

<!-- SOURCE: U24e_docs_archive_idea_uk.md -->
### Capability watchlist + real-world event watchlist (dual standing research workflows)
- **category:** research-agents
- **status-signal:** aspirational
- **status-evidence:** `running_notes(44).md`: "The capability list in the method is a starter; the watchlist workflow itself isn't designed" (open thread, never closed within this file); later: "Real-world event watchlist promoted to a second standing workflow... Both are recurring research workflows that fire re-runs of ideation."
- **what:** Two proposed recurring background research workflows: (1) a capability watchlist tracking new AI capabilities that beat the model's self-knowledge (agentic browsing, million-token contexts, real-time voice, etc.) — the "early-adopter mechanism"; (2) an event/window watchlist tracking scheme deadlines, regulation changes and application windows per domain (proven by the agritec SFI26 Window 1 case, which turned a "consider later" candidate into "test now"). Both are meant to trigger automatic re-runs of the ideation method across domains, but the trigger mechanism itself was never designed/built within this archive's timeframe.
- **sources:** `running_notes(44).md` ("Capability watchlist warrants its own workflow", "Watchlist should track scheme/event windows, not just AI capabilities", "Real-world event watchlist promoted to a second standing workflow")
- **relations:** idea generation method
- **verify-later:** whether any scheduled_task / agent implements this in the live chassis

<!-- SOURCE: U26_misc_dirs.md -->
### Deep-research domain insight agent
- **category:** research-agents
- **status-signal:** abandoned
- **status-evidence:** 016 designs a "domain-insight-agent" deciding when deep social research pays ("Value Multiple: 50-100x"); tied to the abandoned domain-flipping context, though its research-orchestration DNA resembles the later research agents.
- **what:** A strategic classifier that assesses whether a domain/topic merits multi-platform deep research (Reddit/LinkedIn/Twitter/Facebook/YouTube community mining, influencer mapping, sentiment threading) versus standard development, then deploys the appropriate research agent squad to synthesise unique content, tools and FAQs from real community pain points — the claimed competitive moat over single-LLM or SEO-tool approaches.
- **sources:** docs/architecture/016-competitive-advantge.md#enhanced-domain-analysis-agent; docs/architecture/016-competitive-advantge.md#deep-research-workflows-by-domain-type
- **relations:** domain value maximisation; topic amplifier engine; current research-agents lineage
- **verify-later:** n/a

<!-- SOURCE: U21_legacy_docs_b.md -->
### Entity data agent family (structured data drives pages)
- **category:** NEW:entity-data
- **status-signal:** partial
- **status-evidence:** docs017/019b: site_entities/site_entity_relationships "(exist)", entity_sources/entity_sync_log "(planned)"; "First implementation target: boxing ticket/events site, then football tickets, then finance"; no later doc confirms the sync pipeline ran.
- **what:** Real-world entities (events, performers, venues, ticket tiers, products, articles) stored as typed JSONB rows with relationships, synced from configured sources (API/scrape/feed with field_mapping, poll intervals, rate limits), change-logged, and driving template-rendered pages with minimal LLM. Entity lifecycle is state-based, not time-based (announced → on_sale → selling_fast → sold_out → event_day → past → historical/cancelled) with per-state page and nav behaviour; status transitions auto-detected from source data. entity_sources.news_triggers defines which changes are newsworthy, bridging to the feed pipeline. Three of four stress-tested site types need it.
- **sources:** docs017_legacy_agent_rules_images_design_keydocs/019b_agent_architecture_v5_with_tickets_news.md#5-Entity-Data-Agent-Family; docs017_legacy_agent_rules_images_design_keydocs/019b_agent_architecture_v5_with_tickets_news.md#Entity-Lifecycle; docs017_legacy_agent_rules_images_design_keydocs/023_maintenance_architecture_unified_v6.md#Entity-Data
- **relations:** news feed pipeline (entity_event triggers); tickets vertical; products tables (superseded by entities); dogs-medicine entities unrelated.
- **verify-later:** site_entities/site_entity_relationships rows; entity_sources/entity_sync_log existence; entity-data-agent definition.

<!-- SOURCE: U21_legacy_docs_b.md -->
### Events/tickets vertical (boxing first target)
- **category:** NEW:entity-data
- **status-signal:** abandoned
- **status-evidence:** docs017/019b "Events / Tickets Site (first target — boxing, then football)... API sources: Ticketmaster, SeatGeek, BoxRec"; entity examples "Fury vs Joshua"; no boxing/tickets site appears in later portfolio lists.
- **what:** The planned first entity-driven site type: dense entity relationships (event↔performer↔venue↔ticket_tier), frequently-updating ticket tier data (price/availability) flowing to pages quickly, state-transition-driven news (fight announced, on sale, sold out, results), contextual per-event/per-performer navigation, and past events retained as permanent SEO assets.
- **sources:** docs017_legacy_agent_rules_images_design_keydocs/019b_agent_architecture_v5_with_tickets_news.md#Entity-Types-for-Events-Tickets; docs017_legacy_agent_rules_images_design_keydocs/019b_agent_architecture_v5_with_tickets_news.md#Site-Type-Stress-Tests
- **relations:** entity data family; news feed pipeline.
- **verify-later:** any boxing/tickets site records.

<!-- SOURCE: U21_legacy_docs_b.md -->
### Entity data agent family (structured data drives pages)
- **category:** NEW:entity-data
- **status-signal:** partial
- **status-evidence:** docs017/019b: site_entities/site_entity_relationships "(exist)", entity_sources/entity_sync_log "(planned)"; "First implementation target: boxing ticket/events site, then football tickets, then finance"; no later doc confirms the sync pipeline ran.
- **what:** Real-world entities (events, performers, venues, ticket tiers, products, articles) stored as typed JSONB rows with relationships, synced from configured sources (API/scrape/feed with field_mapping, poll intervals, rate limits), change-logged, and driving template-rendered pages with minimal LLM. Entity lifecycle is state-based, not time-based (announced → on_sale → selling_fast → sold_out → event_day → past → historical/cancelled) with per-state page and nav behaviour; status transitions auto-detected from source data. entity_sources.news_triggers defines which changes are newsworthy, bridging to the feed pipeline. Three of four stress-tested site types need it.
- **sources:** docs017_legacy_agent_rules_images_design_keydocs/019b_agent_architecture_v5_with_tickets_news.md#5-Entity-Data-Agent-Family; docs017_legacy_agent_rules_images_design_keydocs/019b_agent_architecture_v5_with_tickets_news.md#Entity-Lifecycle; docs017_legacy_agent_rules_images_design_keydocs/023_maintenance_architecture_unified_v6.md#Entity-Data
- **relations:** news feed pipeline (entity_event triggers); tickets vertical; products tables (superseded by entities); dogs-medicine entities unrelated.
- **verify-later:** site_entities/site_entity_relationships rows; entity_sources/entity_sync_log existence; entity-data-agent definition.

<!-- SOURCE: U21_legacy_docs_b.md -->
### Events/tickets vertical (boxing first target)
- **category:** NEW:entity-data
- **status-signal:** abandoned
- **status-evidence:** docs017/019b "Events / Tickets Site (first target — boxing, then football)... API sources: Ticketmaster, SeatGeek, BoxRec"; entity examples "Fury vs Joshua"; no boxing/tickets site appears in later portfolio lists.
- **what:** The planned first entity-driven site type: dense entity relationships (event↔performer↔venue↔ticket_tier), frequently-updating ticket tier data (price/availability) flowing to pages quickly, state-transition-driven news (fight announced, on sale, sold out, results), contextual per-event/per-performer navigation, and past events retained as permanent SEO assets.
- **sources:** docs017_legacy_agent_rules_images_design_keydocs/019b_agent_architecture_v5_with_tickets_news.md#Entity-Types-for-Events-Tickets; docs017_legacy_agent_rules_images_design_keydocs/019b_agent_architecture_v5_with_tickets_news.md#Site-Type-Stress-Tests
- **relations:** entity data family; news feed pipeline.
- **verify-later:** any boxing/tickets site records.

<!-- SOURCE: U24d_docs_archive_adoption_content_quality.md -->
### Chassis deploy-mechanism reference (targets A–F)
- **category:** NEW:deploy-mechanics-reference
- **status-signal:** deployed
- **status-evidence:** live `docubundle/GUIDE_deploy_from_context_packs.md` names six distinct deploy mechanisms (A: chassis image rebuild+rollout, B: DB/SQL migration, C: work-item insert, D: orchestration `orchestrate` trigger via kcat, E: generated static site via git→GitHub Actions→Backblaze, F: idea.uk standalone binary) and a per-project quick reference mapping each named task to its mechanism(s).
- **what:** A structured taxonomy of "what shipping a change actually means" per target: the agent-chassis Kubernetes image is a different deploy surface from the sites it builds (Backblaze-hosted static output) which is different again from the idea.uk box (file-based, cPanel, no k8s/DB). The archived draft only had this half-formed (a looser walkthrough focused on one task, skinner-box); the live version generalized it into the reusable A–F reference.
- **sources:** adoption/docubundle/GUIDE_deploy_from_context_packs(1).md; live docubundle/GUIDE_deploy_from_context_packs.md
- **relations:** adapters (033/035), deployment-github (034), storage-architecture (032)
- **verify-later:** confirm the A–F reference still matches current `makefile.txt`/`kustomization.yaml` targets.

<!-- SOURCE: U24d_docs_archive_adoption_content_quality.md -->
### Chassis deploy-mechanism reference (targets A–F)
- **category:** NEW:deploy-mechanics-reference
- **status-signal:** deployed
- **status-evidence:** live `docubundle/GUIDE_deploy_from_context_packs.md` names six distinct deploy mechanisms (A: chassis image rebuild+rollout, B: DB/SQL migration, C: work-item insert, D: orchestration `orchestrate` trigger via kcat, E: generated static site via git→GitHub Actions→Backblaze, F: idea.uk standalone binary) and a per-project quick reference mapping each named task to its mechanism(s).
- **what:** A structured taxonomy of "what shipping a change actually means" per target: the agent-chassis Kubernetes image is a different deploy surface from the sites it builds (Backblaze-hosted static output) which is different again from the idea.uk box (file-based, cPanel, no k8s/DB). The archived draft only had this half-formed (a looser walkthrough focused on one task, skinner-box); the live version generalized it into the reusable A–F reference.
- **sources:** adoption/docubundle/GUIDE_deploy_from_context_packs(1).md; live docubundle/GUIDE_deploy_from_context_packs.md
- **relations:** adapters (033/035), deployment-github (034), storage-architecture (032)
- **verify-later:** confirm the A–F reference still matches current `makefile.txt`/`kustomization.yaml` targets.

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Public API plan: site_ownership junction + user-facing build/HITL endpoints
- **category:** NEW:public-api
- **status-signal:** aspirational
- **status-evidence:** P2 is an implementation plan (blocks 0–6, build order); Block 3 admin subset "implemented" per its own notes
- **what:** site_ownership junction table (site/client/user/role) rather than columns on sites (shared sites; 15+ FKs untouched); all public queries scope through it. POST /sites writes build_queue + ownership (seed picks it up; 409 on existing). Endpoints for sites/status (work-item progress rollup), pages, work items with the HITL review flow (needs_human_review → provide-data-and-retry / retry / dismiss; retry converts to content_rewrite), specs read+write, assets, briefing HTTP-to-Kafka bridge, WebSocket build events.
- **sources:** P2 full
- **relations:** admin API; needs_human_review status; build_queue
- **verify-later:** site_ownership table; which blocks landed

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Public API plan: site_ownership junction + user-facing build/HITL endpoints
- **category:** NEW:public-api
- **status-signal:** aspirational
- **status-evidence:** P2 is an implementation plan (blocks 0–6, build order); Block 3 admin subset "implemented" per its own notes
- **what:** site_ownership junction table (site/client/user/role) rather than columns on sites (shared sites; 15+ FKs untouched); all public queries scope through it. POST /sites writes build_queue + ownership (seed picks it up; 409 on existing). Endpoints for sites/status (work-item progress rollup), pages, work items with the HITL review flow (needs_human_review → provide-data-and-retry / retry / dismiss; retry converts to content_rewrite), specs read+write, assets, briefing HTTP-to-Kafka bridge, WebSocket build events.
- **sources:** P2 full
- **relations:** admin API; needs_human_review status; build_queue
- **verify-later:** site_ownership table; which blocks landed

<!-- SOURCE: U22_recent_small_docs.md -->
### Polite-scraping throttle (REQUEST_THROTTLE_MS)
- **category:** adopting-and-scraping
- **status-signal:** aspirational
- **status-evidence:** "(Optional) Throttle adapters — if you want the 5s delays between requests, add the throttle code and set REQUEST_THROTTLE_MS=5000 on the webscrape and web-search adapter deployments."
- **what:** An optional per-adapter throttle env var adding fixed delays between outbound web-scrape/web-search requests, to keep bulk vet data collection polite and avoid rate-limit/blocking. Presented as opt-in infra config, not verified as deployed.
- **sources:** docs019_business/005_initial_messaging.md#before-running
- **relations:** area-sweep discovery, vet-practice-verifier
- **verify-later:** REQUEST_THROTTLE_MS handling in webscrape/web-search adapters

<!-- SOURCE: U22_recent_small_docs.md -->
### Polite-scraping throttle (REQUEST_THROTTLE_MS)
- **category:** adopting-and-scraping
- **status-signal:** aspirational
- **status-evidence:** "(Optional) Throttle adapters — if you want the 5s delays between requests, add the throttle code and set REQUEST_THROTTLE_MS=5000 on the webscrape and web-search adapter deployments."
- **what:** An optional per-adapter throttle env var adding fixed delays between outbound web-scrape/web-search requests, to keep bulk vet data collection polite and avoid rate-limit/blocking. Presented as opt-in infra config, not verified as deployed.
- **sources:** docs019_business/005_initial_messaging.md#before-running
- **relations:** area-sweep discovery, vet-practice-verifier
- **verify-later:** REQUEST_THROTTLE_MS handling in webscrape/web-search adapters
