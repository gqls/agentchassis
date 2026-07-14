
<!-- SOURCE: U22_recent_small_docs.md -->
### business_intel schema (multi-vertical business intelligence platform)
- **category:** NEW:business-intelligence-platform
- **status-signal:** partial
- **status-evidence:** Agent definitions seeded with `status = 'experimental'`; verification_progress/discovery_stats views described as "used by bulk script monitoring"; ~3,500 vet practices already loaded.
- **what:** A separate `business_intel` schema inside `clients_db` (distinct from the public website-builder schema) that models businesses for data-collection verticals. Layered design: universal `businesses` table + `business_verticals` registry, with 1:1 vertical detail tables (`vet_practice_details`), pricing, products, people, and provenance. Seeded verticals: veterinary, online-pharmacy, seaweed-farming.
- **sources:** docs019_business/001_business_intel_schema.sql#layers, docs019_business/002_business_intel_actions.md#load_business_record
- **relations:** vet-practice-verifier, area-sweep discovery, vet-med-pricing (business_prices/product_prices)
- **verify-later:** schema `business_intel` in clients_db; tables businesses, business_verticals, vet_practice_details, business_prices, products, product_prices

<!-- SOURCE: U22_recent_small_docs.md -->
### Business verticals registry (business_intel)
- **category:** NEW:business-intelligence-platform
- **status-signal:** deployed
- **status-evidence:** `INSERT INTO business_intel.business_verticals ... ON CONFLICT (slug) DO NOTHING` seeds veterinary/online-pharmacy/seaweed-farming with `default_agent_type`.
- **what:** A `business_verticals` table keying each collection vertical by slug, display name, `default_agent_type` (e.g. `vet-practice-verifier`, `pharmacy-price-monitor`), and per-vertical `collection_config` JSONB. Businesses and collection tasks reference the vertical; used to scope which agent handles a business type. Distinct from the docs021 knowledge `vertical_registry`.
- **sources:** docs019_business/001_business_intel_schema.sql#seed-verticals
- **relations:** vertical_registry (docs021 — different table, same "vertical" idea), collection_tasks
- **verify-later:** business_intel.business_verticals rows and default_agent_type usage

<!-- SOURCE: U22_recent_small_docs.md -->
### Data observations provenance model
- **category:** NEW:business-intelligence-platform
- **status-signal:** deployed
- **status-evidence:** `store_business_verification` inserts a `data_observations` row per agent run with `raw_data JSONB`, source, confidence, orchestration_id.
- **what:** Every scrape/search/submission is recorded as a `data_observations` row carrying raw + extracted data, source type/name/url, extraction confidence, and the producing `orchestration_id`. Provides an audit trail and change history for business facts, separate from the current values on the business record. Temporal staleness columns (first_seen/last_confirmed/missed_count/is_stale) track freshness on prices and contacts.
- **sources:** docs019_business/001_business_intel_schema.sql#layer3, docs019_business/009_discovery_candidates.sql#temporal-tracking
- **relations:** business_intel schema, vet-practice-verifier
- **verify-later:** business_intel.data_observations; is_stale/missed_count columns; stale_contacts view

<!-- SOURCE: U22_recent_small_docs.md -->
### collection_tasks queue + batch claiming
- **category:** NEW:business-intelligence-platform
- **status-signal:** deployed
- **status-evidence:** `load_business_batch` claims pending tasks via `FOR UPDATE SKIP LOCKED`; unique partial index prevents duplicate pending tasks per business+task_type.
- **what:** A `collection_tasks` queue (task_type initial_verification/price_refresh/status_check/discovery; status pending→in_progress→completed/failed/needs_review; priority). Agents claim batches atomically with SKIP LOCKED and reset orphaned in_progress rows after crashes. `ensure_collection_tasks` backfills tasks for pending businesses.
- **sources:** docs019_business/001_business_intel_schema.sql#collection_tasks, docs019_business/002_business_intel_actions.md#load_business_batch, docs019_business/015_collection_tasks.sql
- **relations:** vet-batch-processor, maintenance_queue (same claim pattern)
- **verify-later:** business_intel.collection_tasks; idx_collection_tasks_unique_pending; load_business_batch action

<!-- SOURCE: U22_recent_small_docs.md -->
### vet-practice-verifier agent
- **category:** NEW:business-intelligence-platform
- **status-signal:** partial
- **status-evidence:** Seeded `status='experimental'`; SQL file is a long trail of iterative production fixes (Go-template dots, ai_service path, scraped_data path, prepare_context step) implying live debugging, plus "stuck at scrape_website - timeout goroutine lost" cleanup.
- **what:** Single-practice orchestrator workflow: load_business → search_practice (web_search) → scrape_website → prepare_context → extract_and_reconcile (LLM JSON extraction of business/vet_details/prices/staff/contacts) → store_results → scan_discoveries. Runs on claude-haiku-4-5. Callable standalone or spawned by the batch processor.
- **sources:** docs019_business/004_vet_practice_verifier.sql, docs019_business/002_business_intel_actions.md#store_business_verification
- **relations:** vet-batch-processor, prepare_extraction_context, scan_discovery_candidates, discovery_candidates
- **verify-later:** agent_definitions type='vet-practice-verifier'; actions load_business_record/store_business_verification

<!-- SOURCE: U22_recent_small_docs.md -->
### vet-batch-processor agent
- **category:** NEW:business-intelligence-platform
- **status-signal:** partial
- **status-evidence:** Three documented fix rounds (spawn verifier first; continue_on_error; "remove loop — loop steps can't re-expand", max_iterations 50→250) show it broke and was reworked in production.
- **what:** Single-pod orchestrator that claims a batch of pending verification tasks and processes them sequentially, spawning one reusable `vet-practice-verifier` and calling it per business. Designed for polite, low-throughput collection; drains the queue by re-running.
- **sources:** docs019_business/003_vet_batch_processor.sql, docs019_business/005_initial_messaging.md, docs019_business/006_initial_messaging.sh
- **relations:** vet-practice-verifier, collection_tasks, vet-pipeline-orchestrator
- **verify-later:** agent_definitions type='vet-batch-processor'; loop step re-expansion behaviour in chassis

<!-- SOURCE: U22_recent_small_docs.md -->
### Geographic area-sweep discovery system
- **category:** NEW:business-intelligence-platform
- **status-signal:** partial
- **status-evidence:** "We currently have around 3,500 practices and estimate ~5,000 in the UK"; converted from fire-and-forget dispatch to spawn+loop; costs "3,402 credits out of our 100k/month budget".
- **what:** Two-agent system (`area-sweep-orchestrator` + `area-sweep-discoverer`) that sweeps every UK postcode district (3,402 seeded in `search_areas`) via Firecrawl search, skips directory/aggregator domains, checks results against existing businesses and candidates, and inserts new finds into `discovery_candidates`. Go actions: load_unswept_areas, dispatch_area_discoverers, process_area_sweep.
- **sources:** docs019_business/011_area_sweep_discovery_system.md, docs019_business/010_district_search_areas_uk.sql, docs019_business/014_vet_pipeline_orchestrator.sql
- **relations:** discovery_candidates, vet-pipeline-orchestrator, search_result_cache
- **verify-later:** business_intel.search_areas seed count; actions load_unswept_areas/dispatch_area_discoverers/process_area_sweep

<!-- SOURCE: U22_recent_small_docs.md -->
### Discovery candidates + promotion pipeline
- **category:** NEW:business-intelligence-platform
- **status-signal:** partial
- **status-evidence:** `promote_candidates` PL/pgSQL loops pending candidates → businesses with dedup/dismiss logic; status flow pending→matched/promoted/dismissed/needs_enrichment.
- **what:** `discovery_candidates` stores practices found in search results that don't match existing records, with match_method/confidence and group detection. A promotion routine inserts website-bearing candidates into `businesses` (status 'pending'), skips URL duplicates and directory-title junk, and queues them for verification. `search_result_cache` stores raw results for later mining.
- **sources:** docs019_business/009_discovery_candidates.sql, docs019_business/012_promote_candidates.sql
- **relations:** area-sweep discovery, collection_tasks, scan_discovery_candidates
- **verify-later:** business_intel.discovery_candidates; promote_candidates action; discovery_summary/discovery_stats views

<!-- SOURCE: U22_recent_small_docs.md -->
### vet-pipeline-orchestrator (rolling pipeline)
- **category:** NEW:business-intelligence-platform
- **status-signal:** partial
- **status-evidence:** Multiple reworks: fire-and-forget → spawn+loop coordinator → thin coordinator calling child orchestrators; timeouts bumped to 12h sweep / 6h verify.
- **what:** A rolling coordinator that advances work from previous runs each time it runs: load unswept areas → sweep → promote discovery_candidates → ensure_collection_tasks → run batch verification. Evolved from firing Kafka messages to spawning/awaiting child orchestrators (area-sweep + batch-processor) with promotion between.
- **sources:** docs019_business/014_vet_pipeline_orchestrator.sql
- **relations:** area-sweep discovery, vet-batch-processor, promote_candidates
- **verify-later:** agent_definitions type='vet-pipeline-orchestrator'; ensure_collection_tasks action

<!-- SOURCE: U22_recent_small_docs.md -->
### prepare_extraction_context / scan_discovery_candidates actions
- **category:** NEW:business-intelligence-platform
- **status-signal:** partial
- **status-evidence:** Wired into vet-practice-verifier via UPDATE migrations adding prepare_context and scan_discoveries steps; "reuses the skipDomains map from scan_discovery_candidates.go".
- **what:** Two supporting Go actions in the vet pipeline: `prepare_extraction_context` formats search results + scraped content (max_content_length/max_snippets) into a clean `extraction_context` for the LLM step; `scan_discovery_candidates` scans a verifier's search results for unknown practices (skipping aggregator domains) and inserts them into discovery_candidates. Both illustrate the "complexity in Go actions, thin workflow" convention.
- **sources:** docs019_business/004_vet_practice_verifier.sql#prepare_context, #scan_discoveries, docs019_business/011_area_sweep_discovery_system.md
- **relations:** vet-practice-verifier, discovery_candidates, area-sweep process_area_sweep (shares skipDomains)
- **verify-later:** actions prepare_extraction_context, scan_discovery_candidates; skipDomains map

<!-- SOURCE: U22_recent_small_docs.md -->
### business_intel schema (multi-vertical business intelligence platform)
- **category:** NEW:business-intelligence-platform
- **status-signal:** partial
- **status-evidence:** Agent definitions seeded with `status = 'experimental'`; verification_progress/discovery_stats views described as "used by bulk script monitoring"; ~3,500 vet practices already loaded.
- **what:** A separate `business_intel` schema inside `clients_db` (distinct from the public website-builder schema) that models businesses for data-collection verticals. Layered design: universal `businesses` table + `business_verticals` registry, with 1:1 vertical detail tables (`vet_practice_details`), pricing, products, people, and provenance. Seeded verticals: veterinary, online-pharmacy, seaweed-farming.
- **sources:** docs019_business/001_business_intel_schema.sql#layers, docs019_business/002_business_intel_actions.md#load_business_record
- **relations:** vet-practice-verifier, area-sweep discovery, vet-med-pricing (business_prices/product_prices)
- **verify-later:** schema `business_intel` in clients_db; tables businesses, business_verticals, vet_practice_details, business_prices, products, product_prices

<!-- SOURCE: U22_recent_small_docs.md -->
### Business verticals registry (business_intel)
- **category:** NEW:business-intelligence-platform
- **status-signal:** deployed
- **status-evidence:** `INSERT INTO business_intel.business_verticals ... ON CONFLICT (slug) DO NOTHING` seeds veterinary/online-pharmacy/seaweed-farming with `default_agent_type`.
- **what:** A `business_verticals` table keying each collection vertical by slug, display name, `default_agent_type` (e.g. `vet-practice-verifier`, `pharmacy-price-monitor`), and per-vertical `collection_config` JSONB. Businesses and collection tasks reference the vertical; used to scope which agent handles a business type. Distinct from the docs021 knowledge `vertical_registry`.
- **sources:** docs019_business/001_business_intel_schema.sql#seed-verticals
- **relations:** vertical_registry (docs021 — different table, same "vertical" idea), collection_tasks
- **verify-later:** business_intel.business_verticals rows and default_agent_type usage

<!-- SOURCE: U22_recent_small_docs.md -->
### Data observations provenance model
- **category:** NEW:business-intelligence-platform
- **status-signal:** deployed
- **status-evidence:** `store_business_verification` inserts a `data_observations` row per agent run with `raw_data JSONB`, source, confidence, orchestration_id.
- **what:** Every scrape/search/submission is recorded as a `data_observations` row carrying raw + extracted data, source type/name/url, extraction confidence, and the producing `orchestration_id`. Provides an audit trail and change history for business facts, separate from the current values on the business record. Temporal staleness columns (first_seen/last_confirmed/missed_count/is_stale) track freshness on prices and contacts.
- **sources:** docs019_business/001_business_intel_schema.sql#layer3, docs019_business/009_discovery_candidates.sql#temporal-tracking
- **relations:** business_intel schema, vet-practice-verifier
- **verify-later:** business_intel.data_observations; is_stale/missed_count columns; stale_contacts view

<!-- SOURCE: U22_recent_small_docs.md -->
### collection_tasks queue + batch claiming
- **category:** NEW:business-intelligence-platform
- **status-signal:** deployed
- **status-evidence:** `load_business_batch` claims pending tasks via `FOR UPDATE SKIP LOCKED`; unique partial index prevents duplicate pending tasks per business+task_type.
- **what:** A `collection_tasks` queue (task_type initial_verification/price_refresh/status_check/discovery; status pending→in_progress→completed/failed/needs_review; priority). Agents claim batches atomically with SKIP LOCKED and reset orphaned in_progress rows after crashes. `ensure_collection_tasks` backfills tasks for pending businesses.
- **sources:** docs019_business/001_business_intel_schema.sql#collection_tasks, docs019_business/002_business_intel_actions.md#load_business_batch, docs019_business/015_collection_tasks.sql
- **relations:** vet-batch-processor, maintenance_queue (same claim pattern)
- **verify-later:** business_intel.collection_tasks; idx_collection_tasks_unique_pending; load_business_batch action

<!-- SOURCE: U22_recent_small_docs.md -->
### vet-practice-verifier agent
- **category:** NEW:business-intelligence-platform
- **status-signal:** partial
- **status-evidence:** Seeded `status='experimental'`; SQL file is a long trail of iterative production fixes (Go-template dots, ai_service path, scraped_data path, prepare_context step) implying live debugging, plus "stuck at scrape_website - timeout goroutine lost" cleanup.
- **what:** Single-practice orchestrator workflow: load_business → search_practice (web_search) → scrape_website → prepare_context → extract_and_reconcile (LLM JSON extraction of business/vet_details/prices/staff/contacts) → store_results → scan_discoveries. Runs on claude-haiku-4-5. Callable standalone or spawned by the batch processor.
- **sources:** docs019_business/004_vet_practice_verifier.sql, docs019_business/002_business_intel_actions.md#store_business_verification
- **relations:** vet-batch-processor, prepare_extraction_context, scan_discovery_candidates, discovery_candidates
- **verify-later:** agent_definitions type='vet-practice-verifier'; actions load_business_record/store_business_verification

<!-- SOURCE: U22_recent_small_docs.md -->
### vet-batch-processor agent
- **category:** NEW:business-intelligence-platform
- **status-signal:** partial
- **status-evidence:** Three documented fix rounds (spawn verifier first; continue_on_error; "remove loop — loop steps can't re-expand", max_iterations 50→250) show it broke and was reworked in production.
- **what:** Single-pod orchestrator that claims a batch of pending verification tasks and processes them sequentially, spawning one reusable `vet-practice-verifier` and calling it per business. Designed for polite, low-throughput collection; drains the queue by re-running.
- **sources:** docs019_business/003_vet_batch_processor.sql, docs019_business/005_initial_messaging.md, docs019_business/006_initial_messaging.sh
- **relations:** vet-practice-verifier, collection_tasks, vet-pipeline-orchestrator
- **verify-later:** agent_definitions type='vet-batch-processor'; loop step re-expansion behaviour in chassis

<!-- SOURCE: U22_recent_small_docs.md -->
### Geographic area-sweep discovery system
- **category:** NEW:business-intelligence-platform
- **status-signal:** partial
- **status-evidence:** "We currently have around 3,500 practices and estimate ~5,000 in the UK"; converted from fire-and-forget dispatch to spawn+loop; costs "3,402 credits out of our 100k/month budget".
- **what:** Two-agent system (`area-sweep-orchestrator` + `area-sweep-discoverer`) that sweeps every UK postcode district (3,402 seeded in `search_areas`) via Firecrawl search, skips directory/aggregator domains, checks results against existing businesses and candidates, and inserts new finds into `discovery_candidates`. Go actions: load_unswept_areas, dispatch_area_discoverers, process_area_sweep.
- **sources:** docs019_business/011_area_sweep_discovery_system.md, docs019_business/010_district_search_areas_uk.sql, docs019_business/014_vet_pipeline_orchestrator.sql
- **relations:** discovery_candidates, vet-pipeline-orchestrator, search_result_cache
- **verify-later:** business_intel.search_areas seed count; actions load_unswept_areas/dispatch_area_discoverers/process_area_sweep

<!-- SOURCE: U22_recent_small_docs.md -->
### Discovery candidates + promotion pipeline
- **category:** NEW:business-intelligence-platform
- **status-signal:** partial
- **status-evidence:** `promote_candidates` PL/pgSQL loops pending candidates → businesses with dedup/dismiss logic; status flow pending→matched/promoted/dismissed/needs_enrichment.
- **what:** `discovery_candidates` stores practices found in search results that don't match existing records, with match_method/confidence and group detection. A promotion routine inserts website-bearing candidates into `businesses` (status 'pending'), skips URL duplicates and directory-title junk, and queues them for verification. `search_result_cache` stores raw results for later mining.
- **sources:** docs019_business/009_discovery_candidates.sql, docs019_business/012_promote_candidates.sql
- **relations:** area-sweep discovery, collection_tasks, scan_discovery_candidates
- **verify-later:** business_intel.discovery_candidates; promote_candidates action; discovery_summary/discovery_stats views

<!-- SOURCE: U22_recent_small_docs.md -->
### vet-pipeline-orchestrator (rolling pipeline)
- **category:** NEW:business-intelligence-platform
- **status-signal:** partial
- **status-evidence:** Multiple reworks: fire-and-forget → spawn+loop coordinator → thin coordinator calling child orchestrators; timeouts bumped to 12h sweep / 6h verify.
- **what:** A rolling coordinator that advances work from previous runs each time it runs: load unswept areas → sweep → promote discovery_candidates → ensure_collection_tasks → run batch verification. Evolved from firing Kafka messages to spawning/awaiting child orchestrators (area-sweep + batch-processor) with promotion between.
- **sources:** docs019_business/014_vet_pipeline_orchestrator.sql
- **relations:** area-sweep discovery, vet-batch-processor, promote_candidates
- **verify-later:** agent_definitions type='vet-pipeline-orchestrator'; ensure_collection_tasks action

<!-- SOURCE: U22_recent_small_docs.md -->
### prepare_extraction_context / scan_discovery_candidates actions
- **category:** NEW:business-intelligence-platform
- **status-signal:** partial
- **status-evidence:** Wired into vet-practice-verifier via UPDATE migrations adding prepare_context and scan_discoveries steps; "reuses the skipDomains map from scan_discovery_candidates.go".
- **what:** Two supporting Go actions in the vet pipeline: `prepare_extraction_context` formats search results + scraped content (max_content_length/max_snippets) into a clean `extraction_context` for the LLM step; `scan_discovery_candidates` scans a verifier's search results for unknown practices (skipping aggregator domains) and inserts them into discovery_candidates. Both illustrate the "complexity in Go actions, thin workflow" convention.
- **sources:** docs019_business/004_vet_practice_verifier.sql#prepare_context, #scan_discoveries, docs019_business/011_area_sweep_discovery_system.md
- **relations:** vet-practice-verifier, discovery_candidates, area-sweep process_area_sweep (shares skipDomains)
- **verify-later:** actions prepare_extraction_context, scan_discovery_candidates; skipDomains map
