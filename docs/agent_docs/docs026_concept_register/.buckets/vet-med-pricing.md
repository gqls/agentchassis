
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
