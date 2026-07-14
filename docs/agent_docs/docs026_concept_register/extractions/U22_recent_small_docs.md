# EXTRACTION U22 — recent small docs (business, LLM/RAG, multiclustering, domain authority, canine biology, chatbot idea.uk)
Extracted 2026-07-13. Files in scope: 52. Concepts found: 64.

## Coverage
| file | treatment |
|---|---|
| docs019_business/001_business_intel_schema.sql | full |
| docs019_business/002_business_intel_actions.md | full |
| docs019_business/003_vet_batch_processor.sql | full |
| docs019_business/004_vet_practice_verifier.sql | full |
| docs019_business/005_initial_messaging.md | full |
| docs019_business/006_initial_messaging.sh | header-scan |
| docs019_business/007_seed_vets.sql | skipped-generated (1.27MB vet seed INSERTs) |
| docs019_business/008_business_contact_details_table.sql | full |
| docs019_business/009_discovery_candidates.sql | full |
| docs019_business/010_district_search_areas_uk.sql | header-scan (schema head; 3,402 district seed rows) |
| docs019_business/011_area_sweep_discovery_system.md | full |
| docs019_business/012_promote_candidates.sql | full |
| docs019_business/014_vet_pipeline_orchestrator.sql | full |
| docs019_business/015_collection_tasks.sql | full |
| docs019_business/016_maintenance_queue_table.sql | full |
| docs019_business/017_maintenance_triage_agent.sql | full |
| docs020_llm_training_rag/001_rag_agent_distribution_architecture.md | full |
| docs020_llm_training_rag/002_automated_go_action_create_and_build_pipeline.md | full |
| docs020_llm_training_rag/003_llm_model_upgrades_and_logging.sql | full |
| docs020_llm_training_rag/004_rag_knowledge_base.sql | full |
| docs020_llm_training_rag/005_PATCHES.md | full |
| docs020_llm_training_rag/006_rag_deployment_bundle.tar.gz | skipped-binary |
| docs020_llm_training_rag/008_README.md | full |
| docs020_llm_training_rag/009_023_session_handoff_vertical_architecture(1).md | full |
| docs020_llm_training_rag/010_simple_explanation.md | full |
| docs020_llm_training_rag/011_where_we_are.md | full |
| docs020_llm_training_rag/012b_rag_best_practices_v2.md | full |
| docs020_llm_training_rag/ollamaetc.zip | skipped-binary |
| docs021_multiclustering/014_multi_cluster_dispatch.md | full |
| docs021_multiclustering/015_scaling_analysis.md | full |
| docs021_multiclustering/020_vertical_cluster_architecture.md | full |
| docs021_multiclustering/021_2026-02-27-...-million-agent-scaling-plan.txt | header-scan (196KB chat transcript) |
| docs021_multiclustering/021_2026-02-28-...-multi-cluster-dispatch-design.txt | header-scan (273KB chat transcript) |
| docs021_multiclustering/021_2026-03-02-...-multi-cluster-dispatch-implementation.txt | header-scan (338KB chat transcript) |
| docs021_multiclustering/024_handoff_summary_2026_03_02.md | full |
| docs021_multiclustering/025_session_handoff_vertical_architecture.md | full |
| docs021_multiclustering/026_implementation_todo_vertical_architecture(2).md | full |
| docs021_multiclustering/multiclusteretc.zip | skipped-binary |
| docs022_domain_authority/001_domain_content_strategy_framework.md | full |
| docs022_domain_authority/002_domain_content_strategy_framework_v2.md | family-latest (identical to 001 bar trailing whitespace) |
| docs022_domain_authority/003_deep_domain_research_authority.md | full |
| docs022_domain_authority/old/004_classifier_notes.md | full |
| docs023_canine_biology/001_canine_biology_grok_plan.md | full |
| docs023_canine_biology/018_canine_biology.md | full |
| docs025_ai_chatbot_idea_uk/excellent_discussions/086_site_chat_turns.sql | full |
| docs025_ai_chatbot_idea_uk/excellent_discussions/FOCUS_site_chatbot_edge_worker_and_context_pack(1).md | family-latest |
| docs025_ai_chatbot_idea_uk/excellent_discussions/FOCUS_site_chatbot_edge_worker_and_context_pack.md | family-delta (older: had `chat-suggester` agent, since replaced) |
| docs025_ai_chatbot_idea_uk/excellent_discussions/PLAN_isolated_chat_environment(4).md | family-latest |
| docs025_ai_chatbot_idea_uk/excellent_discussions/PLAN_isolated_chat_environment(1).md | family-delta |
| docs025_ai_chatbot_idea_uk/excellent_discussions/PLAN_isolated_chat_environment.md | family-delta |
| docs025_ai_chatbot_idea_uk/excellent_discussions/PLAN_simple_paid_multidomain_chat(1).md | family-latest |
| docs025_ai_chatbot_idea_uk/excellent_discussions/PLAN_simple_paid_multidomain_chat.md | family-delta |

## Concepts

### business_intel schema (multi-vertical business intelligence platform)
- **category:** NEW:business-intelligence-platform
- **status-signal:** partial
- **status-evidence:** Agent definitions seeded with `status = 'experimental'`; verification_progress/discovery_stats views described as "used by bulk script monitoring"; ~3,500 vet practices already loaded.
- **what:** A separate `business_intel` schema inside `clients_db` (distinct from the public website-builder schema) that models businesses for data-collection verticals. Layered design: universal `businesses` table + `business_verticals` registry, with 1:1 vertical detail tables (`vet_practice_details`), pricing, products, people, and provenance. Seeded verticals: veterinary, online-pharmacy, seaweed-farming.
- **sources:** docs019_business/001_business_intel_schema.sql#layers, docs019_business/002_business_intel_actions.md#load_business_record
- **relations:** vet-practice-verifier, area-sweep discovery, vet-med-pricing (business_prices/product_prices)
- **verify-later:** schema `business_intel` in clients_db; tables businesses, business_verticals, vet_practice_details, business_prices, products, product_prices

### Business verticals registry (business_intel)
- **category:** NEW:business-intelligence-platform
- **status-signal:** deployed
- **status-evidence:** `INSERT INTO business_intel.business_verticals ... ON CONFLICT (slug) DO NOTHING` seeds veterinary/online-pharmacy/seaweed-farming with `default_agent_type`.
- **what:** A `business_verticals` table keying each collection vertical by slug, display name, `default_agent_type` (e.g. `vet-practice-verifier`, `pharmacy-price-monitor`), and per-vertical `collection_config` JSONB. Businesses and collection tasks reference the vertical; used to scope which agent handles a business type. Distinct from the docs021 knowledge `vertical_registry`.
- **sources:** docs019_business/001_business_intel_schema.sql#seed-verticals
- **relations:** vertical_registry (docs021 — different table, same "vertical" idea), collection_tasks
- **verify-later:** business_intel.business_verticals rows and default_agent_type usage

### Data observations provenance model
- **category:** NEW:business-intelligence-platform
- **status-signal:** deployed
- **status-evidence:** `store_business_verification` inserts a `data_observations` row per agent run with `raw_data JSONB`, source, confidence, orchestration_id.
- **what:** Every scrape/search/submission is recorded as a `data_observations` row carrying raw + extracted data, source type/name/url, extraction confidence, and the producing `orchestration_id`. Provides an audit trail and change history for business facts, separate from the current values on the business record. Temporal staleness columns (first_seen/last_confirmed/missed_count/is_stale) track freshness on prices and contacts.
- **sources:** docs019_business/001_business_intel_schema.sql#layer3, docs019_business/009_discovery_candidates.sql#temporal-tracking
- **relations:** business_intel schema, vet-practice-verifier
- **verify-later:** business_intel.data_observations; is_stale/missed_count columns; stale_contacts view

### collection_tasks queue + batch claiming
- **category:** NEW:business-intelligence-platform
- **status-signal:** deployed
- **status-evidence:** `load_business_batch` claims pending tasks via `FOR UPDATE SKIP LOCKED`; unique partial index prevents duplicate pending tasks per business+task_type.
- **what:** A `collection_tasks` queue (task_type initial_verification/price_refresh/status_check/discovery; status pending→in_progress→completed/failed/needs_review; priority). Agents claim batches atomically with SKIP LOCKED and reset orphaned in_progress rows after crashes. `ensure_collection_tasks` backfills tasks for pending businesses.
- **sources:** docs019_business/001_business_intel_schema.sql#collection_tasks, docs019_business/002_business_intel_actions.md#load_business_batch, docs019_business/015_collection_tasks.sql
- **relations:** vet-batch-processor, maintenance_queue (same claim pattern)
- **verify-later:** business_intel.collection_tasks; idx_collection_tasks_unique_pending; load_business_batch action

### vet-practice-verifier agent
- **category:** NEW:business-intelligence-platform
- **status-signal:** partial
- **status-evidence:** Seeded `status='experimental'`; SQL file is a long trail of iterative production fixes (Go-template dots, ai_service path, scraped_data path, prepare_context step) implying live debugging, plus "stuck at scrape_website - timeout goroutine lost" cleanup.
- **what:** Single-practice orchestrator workflow: load_business → search_practice (web_search) → scrape_website → prepare_context → extract_and_reconcile (LLM JSON extraction of business/vet_details/prices/staff/contacts) → store_results → scan_discoveries. Runs on claude-haiku-4-5. Callable standalone or spawned by the batch processor.
- **sources:** docs019_business/004_vet_practice_verifier.sql, docs019_business/002_business_intel_actions.md#store_business_verification
- **relations:** vet-batch-processor, prepare_extraction_context, scan_discovery_candidates, discovery_candidates
- **verify-later:** agent_definitions type='vet-practice-verifier'; actions load_business_record/store_business_verification

### vet-batch-processor agent
- **category:** NEW:business-intelligence-platform
- **status-signal:** partial
- **status-evidence:** Three documented fix rounds (spawn verifier first; continue_on_error; "remove loop — loop steps can't re-expand", max_iterations 50→250) show it broke and was reworked in production.
- **what:** Single-pod orchestrator that claims a batch of pending verification tasks and processes them sequentially, spawning one reusable `vet-practice-verifier` and calling it per business. Designed for polite, low-throughput collection; drains the queue by re-running.
- **sources:** docs019_business/003_vet_batch_processor.sql, docs019_business/005_initial_messaging.md, docs019_business/006_initial_messaging.sh
- **relations:** vet-practice-verifier, collection_tasks, vet-pipeline-orchestrator
- **verify-later:** agent_definitions type='vet-batch-processor'; loop step re-expansion behaviour in chassis

### Geographic area-sweep discovery system
- **category:** NEW:business-intelligence-platform
- **status-signal:** partial
- **status-evidence:** "We currently have around 3,500 practices and estimate ~5,000 in the UK"; converted from fire-and-forget dispatch to spawn+loop; costs "3,402 credits out of our 100k/month budget".
- **what:** Two-agent system (`area-sweep-orchestrator` + `area-sweep-discoverer`) that sweeps every UK postcode district (3,402 seeded in `search_areas`) via Firecrawl search, skips directory/aggregator domains, checks results against existing businesses and candidates, and inserts new finds into `discovery_candidates`. Go actions: load_unswept_areas, dispatch_area_discoverers, process_area_sweep.
- **sources:** docs019_business/011_area_sweep_discovery_system.md, docs019_business/010_district_search_areas_uk.sql, docs019_business/014_vet_pipeline_orchestrator.sql
- **relations:** discovery_candidates, vet-pipeline-orchestrator, search_result_cache
- **verify-later:** business_intel.search_areas seed count; actions load_unswept_areas/dispatch_area_discoverers/process_area_sweep

### Discovery candidates + promotion pipeline
- **category:** NEW:business-intelligence-platform
- **status-signal:** partial
- **status-evidence:** `promote_candidates` PL/pgSQL loops pending candidates → businesses with dedup/dismiss logic; status flow pending→matched/promoted/dismissed/needs_enrichment.
- **what:** `discovery_candidates` stores practices found in search results that don't match existing records, with match_method/confidence and group detection. A promotion routine inserts website-bearing candidates into `businesses` (status 'pending'), skips URL duplicates and directory-title junk, and queues them for verification. `search_result_cache` stores raw results for later mining.
- **sources:** docs019_business/009_discovery_candidates.sql, docs019_business/012_promote_candidates.sql
- **relations:** area-sweep discovery, collection_tasks, scan_discovery_candidates
- **verify-later:** business_intel.discovery_candidates; promote_candidates action; discovery_summary/discovery_stats views

### vet-pipeline-orchestrator (rolling pipeline)
- **category:** NEW:business-intelligence-platform
- **status-signal:** partial
- **status-evidence:** Multiple reworks: fire-and-forget → spawn+loop coordinator → thin coordinator calling child orchestrators; timeouts bumped to 12h sweep / 6h verify.
- **what:** A rolling coordinator that advances work from previous runs each time it runs: load unswept areas → sweep → promote discovery_candidates → ensure_collection_tasks → run batch verification. Evolved from firing Kafka messages to spawning/awaiting child orchestrators (area-sweep + batch-processor) with promotion between.
- **sources:** docs019_business/014_vet_pipeline_orchestrator.sql
- **relations:** area-sweep discovery, vet-batch-processor, promote_candidates
- **verify-later:** agent_definitions type='vet-pipeline-orchestrator'; ensure_collection_tasks action

### maintenance_queue + claim/complete/fail functions
- **category:** content-governance
- **status-signal:** partial
- **status-evidence:** "MAINTENANCE QUEUE TABLE (for future use) ... For now, pages are flagged manually and the agent is triggered via generic-agent." Table + PL/pgSQL claim/complete/fail functions defined.
- **what:** A generic site-maintenance work queue (`maintenance_queue`) keyed by site_id with task_type (page_rebuild/css_update/nav_fix/link_repair/content_refresh), priority, JSONB payload, retry logic, and atomic `claim_maintenance_task`/`complete_maintenance_task`/`fail_maintenance_task` SQL functions using SKIP LOCKED. Later reused as the trigger surface for the chatbot install (`install_chat` task).
- **sources:** docs019_business/016_maintenance_queue_table.sql, docs019_business/017_maintenance_triage_agent.sql, docs025.../FOCUS_site_chatbot_edge_worker_and_context_pack(1).md#install-path
- **relations:** maintenance-triage agent, site-chat-installer, improvement-loop
- **verify-later:** maintenance_queue table + claim/complete/fail functions

### maintenance-triage agent
- **category:** improvement-loop
- **status-signal:** aspirational
- **status-evidence:** Defined with dry_run mode; workflow scans sites, queues page_rebuild tasks, spawns page-rebuild agent; described alongside "for future use" queue.
- **what:** An orchestrator that scans deployed sites for maintenance issues (stale pages, missing pages, broken links, CSS drift), inserts tasks into `maintenance_queue`, then dispatches specialist agents (page-rebuild) per affected site. Supports dry_run (scan+queue without dispatch) and a configurable stale_threshold_days.
- **sources:** docs019_business/017_maintenance_triage_agent.sql
- **relations:** maintenance_queue, page-rebuild agent, improvement-loop
- **verify-later:** agent_definitions type='maintenance-triage'; actions scan_sites_for_maintenance/prepare_rebuild_dispatches

### Polite-scraping throttle (REQUEST_THROTTLE_MS)
- **category:** adopting-and-scraping
- **status-signal:** aspirational
- **status-evidence:** "(Optional) Throttle adapters — if you want the 5s delays between requests, add the throttle code and set REQUEST_THROTTLE_MS=5000 on the webscrape and web-search adapter deployments."
- **what:** An optional per-adapter throttle env var adding fixed delays between outbound web-scrape/web-search requests, to keep bulk vet data collection polite and avoid rate-limit/blocking. Presented as opt-in infra config, not verified as deployed.
- **sources:** docs019_business/005_initial_messaging.md#before-running
- **relations:** area-sweep discovery, vet-practice-verifier
- **verify-later:** REQUEST_THROTTLE_MS handling in webscrape/web-search adapters

### RAG knowledge_base (shared pgvector store)
- **category:** NEW:rag-knowledge-base
- **status-signal:** deployed
- **status-evidence:** "knowledge_base table is sitting empty in Postgres, waiting for content ... ivfflat index"; migration 082 marked idempotent/deployed in the vertical-architecture handoff.
- **what:** A shared (not per-agent) `knowledge_base` table storing chunked content with a `vector(768)` embedding (nomic-embed-text), collection/industry/domain classification, SHA256 dedup, ivfflat cosine index, and a trigram fallback index. Any agent reads via `rag_lookup`, any writes via `rag_index`. Later extended with source provenance columns (docs021).
- **sources:** docs020.../004_rag_knowledge_base.sql, docs020.../010_simple_explanation.md, docs020.../008_README.md
- **relations:** rag_lookup, rag_index, Ollama provider, vertical knowledge architecture
- **verify-later:** knowledge_base table + idx_kb_embedding (ivfflat) + idx_kb_content_trgm; knowledge_base_stats view

### rag_lookup action (vector search + trigram fallback)
- **category:** NEW:rag-knowledge-base
- **status-signal:** partial
- **status-evidence:** Registry patch written ("NEEDS PATCH — add 2 rag entries"); action code written but registry.go patch listed as not-yet-applied in the handoff.
- **what:** An action that embeds the query via Ollama, runs pgvector cosine similarity within a collection, and returns both structured `rag_results` and a combined `rag_context` string for prompt injection; falls back to Postgres trigram text search when Ollama is down (reported in `search_method`). Best practice: filter by metadata (vertical/component/quality) before ranking, and prepend `search_query:` task prefix.
- **sources:** docs020.../010_simple_explanation.md#rag_lookup, docs020.../012b_rag_best_practices_v2.md, docs020.../005_PATCHES.md#patch-03
- **relations:** rag_index, knowledge_base, content-writer RAG injection
- **verify-later:** GlobalActionRegistry entry rag_lookup; RAGLookupAction min_authority/filter support

### rag_index action (chunk, embed, dedup, store)
- **category:** NEW:rag-knowledge-base
- **status-signal:** partial
- **status-evidence:** New file `rag_actions.go` "ready to add"; registry patch pending; non-fatal embedding failure behaviour specified in the revised plan.
- **what:** An action that splits text into chunks (default ~1000 chars, 200 overlap, sentence-boundary), SHA256-hashes each for dedup, embeds via Ollama, and inserts into `knowledge_base` tagged by collection/metadata. If embedding fails the chunk is still stored (searchable via trigram). Intended to accept source_authority/vertical_slug/knowledge_type once schema extended.
- **sources:** docs020.../010_simple_explanation.md#rag_index, docs020.../012b_rag_best_practices_v2.md#implementation-priority
- **relations:** rag_lookup, knowledge-indexer agent, vertical research handler
- **verify-later:** GlobalActionRegistry entry rag_index; RAGIndexAction dedup on collection+content_hash

### Ollama provider + ollama-adapter
- **category:** model-infrastructure
- **status-signal:** partial
- **status-evidence:** "The Ollama adapter is a pod running the Ollama inference server with nomic-embed-text loaded" in one doc; but the RAG deploy handoff still lists "Deploy Ollama adapter" as a not-yet-done next step — conflicting claims across sessions.
- **what:** An `ollama.go` provider implementing the AIService interface (GenerateText via /api/chat, GenerateEmbedding via /api/embeddings) plus an `ollama-adapter` kustomize deployment (third-party `ollama/ollama` image, PVC for model persistence, init container pulling nomic-embed-text, single replica, ClusterIP 11434). Provides local embeddings and a path to self-hosted local LLMs.
- **sources:** docs020.../008_README.md, docs020.../009_023_session_handoff_vertical_architecture(1).md, docs021.../026_implementation_todo_vertical_architecture(2).md#0.3
- **relations:** rag_index, rag_lookup, self-hosted LLM inference, LoRA fine-tuning (GGUF via ollama create)
- **verify-later:** aiservice/ollama.go; deployments/kustomize/services/ollama-adapter/*; createAIClient "ollama" case

### llm_call_log (build-time training flywheel)
- **category:** finetuning-flywheel
- **status-signal:** partial
- **status-evidence:** "The llm_call_log table is capturing every LLM call from every agent ... This started logging the moment you deployed the new chassis image" vs handoff listing the Go patches as ready-but-not-committed.
- **what:** A table capturing every `execute_llm_prompt` call (agent_type, step, model, rendered prompt, response, input/output tokens, latency, success) via a fire-and-forget goroutine logger. Feeds cost/latency analytics (`llm_call_stats` view) and accumulates toward the 200+-examples-per-agent fine-tuning threshold. Cleanup function exists but nothing calls it (table-bloat risk flagged, ~1GB/month).
- **sources:** docs020.../003_llm_model_upgrades_and_logging.sql, docs020.../005_PATCHES.md#patch-01-02, docs020.../001_rag_agent_distribution_architecture.md#item-2
- **relations:** anthropic.go usage capture patch, site_chat_turns (deliberately separate log), LoRA training data export
- **verify-later:** llm_call_log table + cleanup_old_llm_logs; LogLLMCall in ai_actions.go; anthropic.go __usage_input_tokens write-back

### Model alias upgrades (Sonnet/Opus 4.5–4.6)
- **category:** model-infrastructure
- **status-signal:** deployed
- **status-evidence:** Migration 081 does idempotent `UPDATE agent_definitions SET default_config = replace(... claude-haiku-4-5 ... claude-sonnet-4-5 ...)`; handoff records chief-strategist→opus-4-6, planners/classifiers→sonnet-4-6, stale claude-3.x refs replaced.
- **what:** SQL migrations that upgrade per-agent model references in `agent_definitions.default_config` — planning/strategy agents to the strongest tier (chief-strategist→opus, site/domain planners+classifiers→sonnet), content generation kept on haiku for cost, and all stale `claude-3.x` aliases modernised. Model aliases resolve to API strings; both original and resolved names logged.
- **sources:** docs020.../003_llm_model_upgrades_and_logging.sql, docs020.../009_023_session_handoff_vertical_architecture(1).md#done
- **relations:** llm_call_log (logs resolved model), model_aliases.go
- **verify-later:** agent_definitions model values for chief-strategist/site-planner/site-classifier; model_aliases.go 4.6 entries

### Automated Go action build pipeline (compiler pod)
- **category:** NEW:action-build-pipeline
- **status-signal:** aspirational
- **status-evidence:** "This is a medium-term investment"; whole doc is a design with a numbered ordered rollout, no deployment claim.
- **what:** A design for an in-cluster compiler pod that watches git for LLM-written Go action files, compiles the full chassis, runs tests, has a second-LLM review stage, builds an image via kaniko, and deploys per an HITL dial (manual→staging→auto) with rollback via recorded previous_tag. Uses an `action_build_jobs` job/audit table; git stays the source of truth, replacing GitHub Actions. Closes the loop: LLM identifies missing capability → writes action → compiled/tested/deployed → wires into workflow JSON.
- **sources:** docs020.../002_automated_go_action_create_and_build_pipeline.md
- **relations:** modular discovery-check registry (init() pattern), HITL, tool-lifecycle
- **verify-later:** action_build_jobs table; any compiler-service/ deployment

### Field-path-resolution duplication tech debt
- **category:** development-guide
- **status-signal:** partial
- **status-evidence:** "the codebase has at least 18 functions that resolve dot-separated field paths ... This is the single biggest code hygiene issue"; datahelpers canonical vs 9+ scattered duplicates enumerated.
- **what:** A recognised code-hygiene problem: ~18 near-duplicate dot-path resolution helpers (resolveFieldPath, ExtractNestedField, GetFieldFromPath, etc.) spread across datahelpers and the actions package, differing subtly in arg order, logging, and `.response` unwrapping. Canonical is `datahelpers.ExtractNestedField`; the standing rule is reuse datahelpers before adding new resolvers. Related recurring bug: Go-template paths need leading dots (`{{.x.y}}`) and input_mappers are compulsory.
- **sources:** docs020.../001_rag_agent_distribution_architecture.md#field-path, docs019_business/004_vet_practice_verifier.sql#go-template-fixes
- **relations:** datahelpers, rag_actions.go helper cleanup
- **verify-later:** datahelpers vs actions-package path resolvers; NullableString/TruncateString/NullableInt in datahelpers

### RAG best practices — filter-first, quality gating, token budget
- **category:** NEW:rag-knowledge-base
- **status-signal:** aspirational
- **status-evidence:** Dated 2026-03-24 best-practices doc; "Implementation Priority" is a to-do list (add metadata columns, update actions), i.e. not yet applied.
- **what:** A methodology for the site-build RAG: always filter by structured metadata (vertical, component_type, source_quality) before embedding-similarity ranking; keep RAG context to 20-30% of the window and 3-5 examples; gate entries by source_quality (high/verified) for prompt injection; track embedding_model and never mix embedding spaces; prepend nomic task prefixes (search_document/search_query); prefer nomic-embed-text-v2-moe. Names five common RAG failures and their fixes.
- **sources:** docs020.../012b_rag_best_practices_v2.md
- **relations:** rag_index, rag_lookup, knowledge sources (scraped/claude-output/human-curated/audit-insight)
- **verify-later:** knowledge_base metadata columns (vertical/component_type/source_quality); task-prefix handling in rag actions

### knowledge-indexer agent (deferred)
- **category:** NEW:rag-knowledge-base
- **status-signal:** aspirational
- **status-evidence:** "Future agent (owns the knowledge-building domain): knowledge-indexer agent ... For now, we implement the actions. The agent comes when we have a use case."
- **what:** A proposed but deliberately-unbuilt agent that would own the knowledge-building process (load indexing targets → web_scrape → rag_index → refresh), called by the maintenance orchestrator or build pipeline. Held back per the "reuse before creating — don't build an agent until the workflow demands one" principle; the rag_index/rag_lookup actions suffice for now.
- **sources:** docs020.../001_rag_agent_distribution_architecture.md#item-4
- **relations:** rag_index, vertical research handler (later realises this role)
- **verify-later:** agent_definitions for any knowledge-indexer/vertical-research-handler

### DispatchAgentAction (remote dispatch via Kafka)
- **category:** multicluster
- **status-signal:** partial
- **status-evidence:** "successfully tested ... Created K8s Job agent-copywriter-f8f34764 in ~640ms"; but "Registry patch: Add dispatch_agent to GlobalActionRegistry" and full-workflow test listed as not done.
- **what:** An action mirroring `SpawnAgentAction` exactly (identical helpers/variable names) except step 7 publishes a `DispatchRequest` to `system.dispatch.requests` instead of creating a local K8s Job. Gives a dual-path spawn model — `spawn_agent` (local) and `dispatch_agent` (remote) — chosen per workflow step, with the parent unaware the child is remote. Longer startup/consumer waits for cross-cluster latency.
- **sources:** docs021.../014_multi_cluster_dispatch.md#1, docs021.../024_handoff_summary_2026_03_02.md
- **relations:** remote-job-spawner, SpawnAgentAction, vertical research/build cluster separation
- **verify-later:** actions/dispatch_actions.go; GlobalActionRegistry dispatch_agent entry

### remote-job-spawner service
- **category:** multicluster
- **status-signal:** deployed
- **status-evidence:** "The remote-job-spawner is deployed to the primary cluster and successfully tested" with logged Job creation (2026-03-02).
- **what:** A standalone stateless Go service (renamed from agent-dispatcher) consuming `system.dispatch.requests`, filtering by `target_cluster` header, and creating local K8s Jobs with the same spec as local spawn — no Postgres dependency (parent already wrote DB records). Confirms to `system.dispatch.responses`; scales horizontally via consumer groups; deployed per remote cluster with `CLUSTER_ID`.
- **sources:** docs021.../014_multi_cluster_dispatch.md#2, docs021.../015_scaling_analysis.md
- **relations:** DispatchAgentAction, system.dispatch.* topics, isolated chat environment (explicitly NOT reused for chat)
- **verify-later:** cmd/remote-job-spawner/main.go; system.dispatch.requests/responses topics; va001 cluster deployment

### Multi-cluster scaling tiers (10K/100K/1M)
- **category:** multicluster
- **status-signal:** aspirational
- **status-evidence:** Explicit phased plan Phase 1-5 with "Current" only at Phase 1 stubbed stress test; per-tier "architectural change" tables.
- **what:** A scaling analysis mapping each agent-count tier to its primary bottleneck and the single architectural change that unlocks it: 10K = topic-creation churn (no change); 50-100K = K8s Job scheduling + Kafka partition count (multi-cluster dispatch + shared topic pools + 8-10 brokers); 1M = per-agent K8s overhead + LLM cost + cross-DC latency (worker pools, regional Kafka/MirrorMaker2, distributed DB, self-hosted GPU inference). Key principle: each jump is one change, agent code never changes.
- **sources:** docs021.../015_scaling_analysis.md, docs021.../021_2026-02-27-...-million-agent-scaling-plan.txt
- **relations:** shared topic pools, worker pools, self-hosted LLM inference
- **verify-later:** current Kafka broker/topic counts; orchestration_states partitioning

### Shared topic pools (replace per-agent topics)
- **category:** multicluster
- **status-signal:** aspirational
- **status-evidence:** "Shared topic pools (needed at this tier) ... Route messages by agent ID in headers instead of by topic isolation."
- **what:** A planned change at the 50-100K tier that replaces two dedicated Kafka topics per agent with a fixed set of partitioned pool topics (e.g. `system.agent-work.requests`, 50-100 partitions), routing by agent ID in headers with the chassis filtering. Eliminates topic-creation churn (the 10K-tier ceiling) entirely.
- **sources:** docs021.../015_scaling_analysis.md#50-100k
- **relations:** worker pools, multi-cluster scaling tiers
- **verify-later:** any shared-topic-pool routing in chassis message handling

### Worker pool architecture (replace per-agent Jobs)
- **category:** multicluster
- **status-signal:** aspirational
- **status-evidence:** "Worker pools replace per-agent Jobs ... This is the biggest code change in the scaling roadmap."
- **what:** The 1M-tier change: long-running chassis pods pull agent work from shared Kafka pools and run multiple workflows concurrently as goroutines, so scaling is a Deployment replica count instead of Job creation. Reuses the existing orchestration engine; the agent doesn't know whether it's a Job, a dispatched Job, or a goroutine.
- **sources:** docs021.../015_scaling_analysis.md#1m, docs021.../014_multi_cluster_dispatch.md#whats-not-done
- **relations:** shared topic pools, self-hosted LLM inference
- **verify-later:** any long-running worker/goroutine-pool mode in chassis

### Self-hosted LLM inference (vLLM/GPU at scale)
- **category:** model-infrastructure
- **status-signal:** aspirational
- **status-evidence:** "Phase 2: Self-Hosted LLM Validation — Deploy vLLM or llama.cpp serving a 7B model"; cost tables for a 48-hour million-agent run.
- **what:** A plan to serve 7B models (Mistral/Llama 3/Qwen 2.5) on GPU via vLLM with continuous batching to escape per-token API costs at scale (1,000-2,000 req/min per A100). Bridges to the Ollama/local-model path and the LoRA fine-tuning targets. Estimated hybrid GPU+CPU cost $1,000-3,000 for a 48-hour million-agent burst.
- **sources:** docs021.../015_scaling_analysis.md#phase-2, docs021.../015_scaling_analysis.md#cost-estimates
- **relations:** Ollama provider, LoRA fine-tuning, worker pools
- **verify-later:** any vLLM/GPU inference deployment or stub_llm action

### Vertical knowledge architecture
- **category:** NEW:vertical-knowledge-architecture
- **status-signal:** aspirational
- **status-evidence:** "The architecture ... is now designed but not yet fully implemented"; implementation todo is phased 0-8 with only Phase 0 partially begun.
- **what:** The strategic pivot from a flat build pipeline to routing each domain to a specialised vertical (veterinary, energy_wholesale, finance_mortgage, seasonal_gifts, generic) that maintains its own knowledge-base collection, research strategy, content patterns, and monetisation config. Verticals are logical first (shared infra, knowledge_base collections + rag_lookup/index), physical later (dispatch_agent). Compounding moat: the tenth domain benefits from the first nine's research.
- **sources:** docs021.../020_vertical_cluster_architecture.md, docs021.../025_session_handoff_vertical_architecture.md, docs020.../011_where_we_are.md
- **relations:** vertical_registry, research/build cluster separation, RAG knowledge_base, site classifier vertical output
- **verify-later:** vertical_registry table; agent_definitions tagged 'vertical-orch'; knowledge_base collection usage

### vertical_registry table + knowledge-base provenance extensions
- **category:** NEW:vertical-knowledge-architecture
- **status-signal:** aspirational
- **status-evidence:** "These additions to the database, not yet applied"; seed INSERT for 5 verticals given as a to-do.
- **what:** A `vertical_registry` table mapping vertical_slug → research/build orchestrator types, knowledge_collection, research_sources, content_patterns/page_type_library, monetisation_config, refresh_schedule, and maturity_stage; plus knowledge_base column extensions (source_authority 1-5, source_url, source_date, vertical_slug, knowledge_type) for provenance-weighted retrieval. Seeds veterinary/energy/mortgage/seasonal/generic.
- **sources:** docs021.../020_vertical_cluster_architecture.md#8, docs021.../026_implementation_todo_vertical_architecture(2).md#2.1
- **relations:** vertical knowledge architecture, rag_lookup min_authority, business_verticals (different table)
- **verify-later:** vertical_registry table; knowledge_base.source_authority/vertical_slug/knowledge_type columns

### Research/build cluster separation
- **category:** NEW:vertical-knowledge-architecture
- **status-signal:** aspirational
- **status-evidence:** "Phase 2: Physical separation ... implementation deferred"; described as designed, Phase 8 in the todo.
- **what:** A two-cluster model separating messy/slow/shared research (web scraping, PDF parsing, LLM knowledge extraction, validation, indexing) from structured/fast/per-site build. They communicate only via the shared knowledge_base (Postgres reads/writes) and Kafka orchestration; build dispatches a research request when it hits a knowledge gap. Justified by independent scaling, failure isolation, clean logs, tighter build-cluster network policy, and clean cost attribution.
- **sources:** docs021.../020_vertical_cluster_architecture.md#5, docs021.../025_session_handoff_vertical_architecture.md#layer-5
- **relations:** vertical knowledge architecture, DispatchAgentAction, vertical research handler
- **verify-later:** any research-cluster agent definitions; dispatch_agent used for research requests

### Vertical research handler + knowledge accumulation loop
- **category:** NEW:vertical-knowledge-architecture
- **status-signal:** aspirational
- **status-evidence:** "Phase 4 (Research Handler)" is an unchecked todo; `needs_vertical_research` work-item type not yet in the schema.
- **what:** A `vertical-research-handler` agent that processes `needs_vertical_research` work items (identify sources → scrape → parse → LLM extract structured knowledge chunks → validate quality/confidence → rag_index with source_authority/vertical_slug). Realises the knowledge-accumulation loop: first domain bears foundational research cost, gaps become research items at priority 1-4 that content items (priority 10-17) depend on, and indexed knowledge benefits all future domains in the vertical.
- **sources:** docs021.../026_implementation_todo_vertical_architecture(2).md#phase-4, docs021.../020_vertical_cluster_architecture.md#6
- **relations:** research/build separation, rag_index, work-item lifecycle, knowledge-indexer agent
- **verify-later:** needs_vertical_research item type; vertical-research-handler agent; check_knowledge_coverage action

### Verticals designed (revenue models + knowledge clusters)
- **category:** business-strategy
- **status-signal:** aspirational
- **status-evidence:** Revenue projection tables labelled "months 12-18" with market data "verified through research"; no live-site revenue claimed.
- **what:** Five verticals worked out with specific knowledge clusters, source lists, page-type libraries, monetisation, and 24-month revenue projections: veterinary/vetcomparison.uk (insurance affiliate £15-35 + listings, £1,960-7,875/mo), energy/gaswholesalers.com (qualified leads £30-60, £1,250-5,350), finance_mortgage/mortgagecalculator.co.uk (broker leads £50-150, £16,500-44,000 — highest value), seasonal_gifts/xmaspresents.com (affiliate 3-17%), plus a "sell the domain not develop" premium pathway (design.co.uk £20-100k).
- **sources:** docs021.../020_vertical_cluster_architecture.md#3, docs021.../025_session_handoff_vertical_architecture.md#verticals-designed, docs022.../003_deep_domain_research_authority.md
- **relations:** vertical knowledge architecture, domain content strategy framework, premium-domain pathway
- **verify-later:** vertical_registry monetisation_config; any live vetcomparison/gaswholesalers/mortgagecalculator sites

### Vertical-specific planner variants
- **category:** design-composition
- **status-signal:** aspirational
- **status-evidence:** Phase 3.5 todo: "Create veterinary/energy/mortgage/seasonal site planner prompt variant" — all unchecked.
- **what:** Separate agent definitions using the same planner Go code but vertical-tuned prompt templates, so a well-established vertical produces better plans than a generic planner with config injected. Each knows its vertical's page types, conversion funnel, and per-page guidance (e.g. every breed-health page links to "find a vet for this breed"; every mortgage calculator has lead capture below results).
- **sources:** docs021.../026_implementation_todo_vertical_architecture(2).md#3.5
- **relations:** site-planner, vertical knowledge architecture, unified site spec
- **verify-later:** agent_definitions for veterinary/energy/mortgage/seasonal site-planner variants

### Fine-tuning pipeline (LoRA flywheel, deferred)
- **category:** finetuning-flywheel
- **status-signal:** aspirational
- **status-evidence:** "Phase 7: Fine-Tuning Pipeline (Deferred) ... becomes relevant after 200+ successful examples per agent type accumulate."
- **what:** The deferred end of the flywheel: once llm_call_log has 200+ examples/agent, export JSONL training data, QLoRA fine-tune a 7B via Unsloth on rented GPU, export GGUF, load into Ollama (`ollama create`), and switch the agent definition to `provider: ollama`. First candidates: site-classifier (high volume, short output), then domain-research-classifier, then the vertical knowledge extractor. Purpose: drive per-call inference cost to ~zero.
- **sources:** docs021.../026_implementation_todo_vertical_architecture(2).md#phase-7, docs023.../018_canine_biology.md#6, docs020.../010_simple_explanation.md
- **relations:** llm_call_log, Ollama provider, canine biology text LoRA, self-hosted LLM inference
- **verify-later:** training-data export queries; any GGUF/Ollama custom models; agent_definitions using provider:ollama

### CSS generation bug (webdesign-agent design_spec not applied)
- **category:** styling-render-pipeline
- **status-signal:** partial
- **status-evidence:** "the webdesign-agent reported css_deployed: success:true ... But the deployed styles.css in git still contains the default blue template — the design_spec colors were never applied" (unsolved, 2026-03-02).
- **what:** A documented production defect: the webdesign-agent generates a correct `design_spec` (industry colours/fonts) but the generated/deployed CSS reverts to the default blue template. Three suspected causes: design_spec not reaching the template in structured form, an over-long prompt reproducing literal template CSS, or `content_field` resolution (`generated_css.result`) losing the CSS in the response envelope. Flagged for stage-2 debugging.
- **sources:** docs021.../024_handoff_summary_2026_03_02.md#the-css-bug
- **relations:** webdesign-agent, git_commit content_field resolution, unified site spec design_intent
- **verify-later:** webdesign-agent generate_css/deploy_css steps; extractFilesForGit content_field handling

### build-dispatch-loop self-chaining removal
- **category:** improvement-loop
- **status-signal:** deployed
- **status-evidence:** "Fix: Loop back to load_next_item internally ... Status: Applied to production DB. Verified live definition matches migration."
- **what:** A fix (migration 063) removing the build-dispatch-loop's self-respawn pattern (spawn_next_dispatch → call_next_dispatch), which repeatedly left the parent stuck in AWAITING_RESPONSES when the child's Kafka response was lost to topic retention/pod restarts. Now loops back to load_next_item internally (9 steps vs 13), timeout bumped 900→1800s.
- **sources:** docs021.../024_handoff_summary_2026_03_02.md#fixes-applied
- **relations:** work-item lifecycle, dispatch loop, orchestration timeouts
- **verify-later:** build-dispatch-loop agent definition step count; migration 063

### Domain content strategy framework (15-question)
- **category:** business-strategy
- **status-signal:** aspirational
- **status-evidence:** "For the content generation pipeline, the 15-question framework should feed into the briefing/research phase" — prescriptive/should, not implemented.
- **what:** A systematic three-layer, 15-question methodology for deciding what content a domain needs to compete: Layer 1 (who visits, intent, satisfaction, money flow), Layer 2 (competitor pages, buying journey, real questions, bookmarkable hook), Layer 3 (best page on the topic, original element, format, next action). Worked examples for gaswholesalers.com and vetcomparison.uk with verified lead/affiliate rates. Questions 5-7 require real competitive research.
- **sources:** docs022.../001_domain_content_strategy_framework.md, docs022.../002_domain_content_strategy_framework_v2.md
- **relations:** domain-strategist prompt, deep research domain authority, site classifier
- **verify-later:** domain-strategist agent prompt; briefing/research phase incorporating the framework

### Deep research domain authority strategy
- **category:** business-strategy
- **status-signal:** aspirational
- **status-evidence:** "The multi-cluster knowledge base approach lets you build that deep knowledge layer for any domain" — strategy doc, canine project cited as proof-of-concept not production.
- **what:** The thesis that content wins on E-E-A-T by synthesising primary/authoritative sources (BSAVA, Ofgem, PRA/FCA, swap-rate data) into knowledge consumers can't easily find, rather than rephrasing published synthesis. A repeatable 6-step pipeline (niche mapping → primary-source identification → multi-cluster KB construction → gap identification → content architecture → generation from KB) creates a defensible moat: depth consistency, cross-cluster synthesis, and update efficiency competitors can't copy by rewriting one article.
- **sources:** docs022.../003_deep_domain_research_authority.md
- **relations:** vertical knowledge architecture, domain content strategy framework, canine biology KB
- **verify-later:** research-agent primary-source handling; knowledge_base source_authority weighting

### Unified site spec (status-tagged single document)
- **category:** site-spec-and-classifier
- **status-signal:** aspirational
- **status-evidence:** Doc lives under `docs022_domain_authority/old/`; "Extend the current classifier output ... Steps 1-3 can happen incrementally" — proposed, and archived.
- **what:** A proposal for the site-classifier to emit one unified spec covering classification, identity, design_intent, content_direction, pages, features, SEO, and maintenance_profile — every item tagged `status` (deployed/planned/blocked). The "dream" is the whole doc; the "build" is the non-blocked subset. Downstream agents (planner enriches rather than decides pages; design/content agents implement explicit intent; audit agents treat the spec as ground truth; HITL edits it).
- **sources:** docs022.../old/004_classifier_notes.md
- **relations:** site-classifier vertical/disposition output, feasibility/blocked-handler pattern, design_intent, HITL
- **verify-later:** site_specs.spec_type='unified_spec'; classifier identity/design_intent/content_direction fields

### Feasibility / blocked-handler pattern
- **category:** site-spec-and-classifier
- **status-signal:** partial
- **status-evidence:** "the claim action catches it ... Item marked 'blocked', error='Handler agent not registered' ... weeks later ... Feasibility-recheck task finds it ... promoted to 'triaged'." Describes an existing dispatch/claim mechanism.
- **what:** A pattern where planners/discovery emit work items naming an intended handler even if that agent doesn't exist yet. The claim action checks agent_definitions; unknown handlers mark the item `blocked` with an error; a periodic feasibility-recheck task promotes blocked items to `triaged` once the handler is deployed. A Go `check_feasibility` action can also pre-block at spec time based on the agent registry.
- **sources:** docs022.../old/004_classifier_notes.md#feasibility-assessment
- **relations:** unified site spec, work-item lifecycle, tool-lifecycle
- **verify-later:** claim action handler-existence check; feasibility-recheck scheduled task

### Content-site valuation model (24-32x)
- **category:** business-strategy
- **status-signal:** aspirational
- **status-evidence:** "using current market multiples of 24-32x monthly profit (Empire Flippers averaging 24x, premium 30-35x)" — used throughout as a projection basis.
- **what:** The valuation basis underpinning the domain portfolio strategy: content/affiliate sites sell at ~24-32x monthly profit, so a £1,500-3,000/mo site is worth ~£36k-96k. Combined with verified per-niche lead/affiliate rates to project each domain's asset value and justify the knowledge-base investment. The portfolio is framed as the testing ground toward a £25k+ annual revenue target and a two-tier service→pipeline-sale model.
- **sources:** docs022.../002_domain_content_strategy_framework_v2.md#monetisation, docs021.../025_session_handoff_vertical_architecture.md#market-data-verified
- **relations:** verticals designed, commercial model (chatbot docs), premium-domain pathway
- **verify-later:** n/a (business assumption)

### Canine biology knowledge base (veterinary seeding)
- **category:** canine-biology
- **status-signal:** aspirational
- **status-evidence:** "The canine biology project stops being aspirational and becomes the working proof..." — future tense; "knowledge base is empty" in the RAG explainer.
- **what:** The first real RAG content and proof-of-concept for the veterinary vertical: structured LLM extraction (breed health profiles for top 20 UK breeds, 30-40 procedures, top 30 conditions, nutrition/vaccination/behaviour) into ~300-500 self-contained 200-500-word chunks, validated (self-consistency, cross-reference, structural), embedded via Ollama, and indexed into `collection: "veterinary"`. Structured JSON with confidence markers, not prose.
- **sources:** docs023.../018_canine_biology.md, docs023.../001_canine_biology_grok_plan.md
- **relations:** RAG knowledge_base, vertical knowledge architecture, text LoRA (vet extractor), deep research domain authority
- **verify-later:** knowledge_base rows collection='veterinary'; knowledge-extractor agent

### Interactive Biological Explorer + experiment engine (aspirational vision)
- **category:** canine-biology
- **status-signal:** abandoned
- **status-evidence:** The grandiose Grok "Final Consolidated Plan" (multi-scale explorer, knowledge graph, experiment engine, 14-week timeline) is explicitly downgraded in the later doc: "The original 1M-agent design was aspirational. This plan is practical."
- **what:** An early, much larger vision: a public Next.js/Three.js/Cytoscape web app allowing drill-down from a pseudo-photographic Labrador image → organ systems → cells → biochemical pathways → genes, backed by a PostgreSQL/Neo4j knowledge graph (Gene/Protein/Metabolite/Reaction/Organ nodes), plus an agent-driven "theoretical experiment engine" running SciPy ODE simulations. Superseded by the practical RAG-seeding plan; the explorer/graph/experiment layers were dropped.
- **sources:** docs023.../001_canine_biology_grok_plan.md, docs023.../018_canine_biology.md#1
- **relations:** canine biology knowledge base (the practical replacement), image LoRA
- **verify-later:** n/a (not built; abandoned scope)

### Text LoRA — veterinary knowledge extractor
- **category:** finetuning-flywheel
- **status-signal:** aspirational
- **status-evidence:** Phase E todo "Text LoRA fine-tuning (week 6-7)" unchecked; full Unsloth/QLoRA recipe given as instructions.
- **what:** A concrete recipe to fine-tune a local 7-8B model (Unsloth QLoRA, r=16, 3 epochs) on accumulated knowledge-extraction examples, export Q4_K_M GGUF, load into Ollama, and swap `knowledge-extractor` to the local model to eliminate Claude API cost per extraction. Training data accrues naturally during the canine research phase (50 breeds + 30 conditions + 40 procedures ≈ 120, need 200+).
- **sources:** docs023.../018_canine_biology.md#6
- **relations:** fine-tuning pipeline, llm_call_log, Ollama provider
- **verify-later:** vet-extractor GGUF/Ollama model; knowledge-extractor agent provider

### Image LoRA — scientific illustration style
- **category:** imagery
- **status-signal:** aspirational
- **status-evidence:** Phase F todo "Image LoRA fine-tuning (week 7-8)" unchecked; "SDXL recommended for diagrams" over FLUX.
- **what:** A plan to train an image LoRA (SDXL/PixArt preferred over FLUX for clean diagrams) on 60-90 curated, captioned veterinary/biological illustrations so the image-generator produces consistent anatomical cross-sections, pathway diagrams, procedure illustrations, and infographics across a site. Served via serverless (Replicate/RunPod) rather than an in-cluster GPU initially.
- **sources:** docs023.../018_canine_biology.md#7
- **relations:** image-generator adapter, canine biology KB, self-hosted inference
- **verify-later:** image-generator adapter LoRA support; any vet-diagram LoRA weights

### Site chatbot edge worker (synchronous, not an orchestrated agent)
- **category:** NEW:site-chatbot
- **status-signal:** aspirational
- **status-evidence:** Design doc with "Suggested build order (structural first)" and "Open decisions" — canonical design, nothing deployed.
- **what:** The canonical design for a per-domain chatbot on static-S3 sites: a synchronous request/response handler on a provider-agnostic serverless edge worker (Cloudflare first), NOT run through Kafka/the chassis. Deliberate documented exception to "every agent is an orchestrator" — Kafka's async failure modes (offset replay, phantom-complete, no streaming) are wrong for live chat, and a central nginx VM would drag static traffic behind a hackable box and lose S3's hack-resistance. Worker: resolve domain → load context pack → guard limits → compose bounded prompt → stream LLM tokens (SSE) → fire-and-forget record turn.
- **sources:** docs025.../FOCUS_site_chatbot_edge_worker_and_context_pack(1).md
- **relations:** context pack, provider-agnostic deps adapters, site_chat_turns, isolated chat environment
- **verify-later:** any edge worker deploy; /api/chat route registration

### Build-time context pack (per-domain bounded context)
- **category:** NEW:site-chatbot
- **status-signal:** aspirational
- **status-evidence:** Section 7 defines the JSON shape and versioning; produced by an unbuilt `chat-context-builder` agent.
- **what:** One per-domain JSON document published to static storage at install time, carrying identity, scope (instructions/refusal message/banned topics), build-time-selected grounding chunks (bounded by token budget), suggested model, and operational limits. The worker holds no per-site logic — the pack is the entire bounded context. Grounding is selected on Layer 1 via Ollama embeddings + pgvector; v2 optionally ships chunk vectors for in-worker per-question retrieval plus a narrow stateless embedding endpoint.
- **sources:** docs025.../FOCUS_site_chatbot_edge_worker_and_context_pack(1).md#7
- **relations:** chat-context-builder, RAG knowledge_base (install-time reuse), three-layer bounding
- **verify-later:** context-pack schema; chat-context-builder agent; pack publish-to-S3 step

### site-chat-installer orchestration (install_chat maintenance task)
- **category:** NEW:site-chatbot
- **status-signal:** aspirational
- **status-evidence:** "Resolved: Install is a separate orchestration, triggered via a maintenance_queue install_chat task — not a build-pipeline stage." Not built.
- **what:** Chat install is its own orchestration (triggered by a `maintenance_queue` `install_chat` task, build pipeline untouched), spawning three sub-agents: `chat-context-builder` (build+publish the pack via Ollama+pgvector), `chat-widget-installer` (fork the chat widget through the existing component/tool pipeline; only difference is it POSTs to /api/chat), and `chat-route-registrar` (record the route + mark chat installed on the site, reversible via uninstall_chat). Supersedes the older `chat-suggester` gating agent from the FOCUS base version.
- **sources:** docs025.../FOCUS_site_chatbot_edge_worker_and_context_pack(1).md#install-path, docs025.../FOCUS_site_chatbot_edge_worker_and_context_pack.md (delta: chat-suggester)
- **relations:** maintenance_queue, context pack, component/tool pipeline, chat-suggester (superseded)
- **verify-later:** site-chat-installer + sub-agent definitions; install_chat maintenance task_type

### Provider-agnostic worker (deps adapters)
- **category:** NEW:site-chatbot
- **status-signal:** aspirational
- **status-evidence:** Section 5 defines interfaces and a Cloudflare shim; "Best practice" reference impls listed, no code shipped.
- **what:** The portability strategy: a Web-platform-only core `handleChat(request, deps)` plus a ~20-line per-platform shim. Three (v2: four) small adapters — ContextStore (HTTP GET of static pack), LLMClient (Anthropic Messages over fetch, swappable to self-hosted), TurnSink (queue/D1, fire-and-forget), and v2 Embedder — each with a Cloudflare and a portable HTTP impl. Nothing vendor-specific in the core; Cloudflare/Deno/Fastly/Vercel/self-host are drop-in. Rate limiting is the least-portable concern (WAF + in-pack per-session cap floor).
- **sources:** docs025.../FOCUS_site_chatbot_edge_worker_and_context_pack(1).md#5, #6
- **relations:** edge worker, context pack, pluggable billing/LLM/storage adapter discipline
- **verify-later:** handleChat core + adapter interfaces if implemented

### site_chat_turns table (turn recording)
- **category:** NEW:site-chatbot
- **status-signal:** aspirational
- **status-evidence:** Migration 086 written and "schema-checked" against live `sites`, but header notes "this snapshot only shows migrations up to 085. Confirm the next free migration number ... before applying."
- **what:** A `site_chat_turns` table logging each end-user prompt/answer turn per domain (question/answer as PII, refused/capped flags, model, pack_version, grounding_ids, tokens/latency named to match llm_call_log, salted client_ip_hash never raw IP). Deliberately separate from the build-time `llm_call_log` (different owner, privacy profile, and access pattern). Edge-supplied turn uuid is the PK for idempotent ingest (ON CONFLICT DO NOTHING); populated by a Layer-1 puller draining the edge sink.
- **sources:** docs025.../086_site_chat_turns.sql, docs025.../FOCUS_site_chatbot_edge_worker_and_context_pack(1).md#8
- **relations:** llm_call_log (kept separate), TurnSink, isolated chat environment (isolated-DB variant drops the FK)
- **verify-later:** site_chat_turns migration number; Layer-1 turn puller

### Three-layer bounding (retrieval / prompt / operational)
- **category:** NEW:site-chatbot
- **status-signal:** aspirational
- **status-evidence:** Section 9 "Bounded context — the three layers"; part of the design.
- **what:** A precise decomposition of "bounded" to stop chatbot drift: retrieval bounding (only this site's grounding is in the pack, frozen at build time), prompt bounding (system prompt pins identity and emits an exact refusal message for out-of-scope questions), and operational bounding (input length, output tokens, turns/session, history window, rate limiting from pack.limits). Conflating the three is where bots go off-topic.
- **sources:** docs025.../FOCUS_site_chatbot_edge_worker_and_context_pack(1).md#9
- **relations:** context pack, edge worker composeSystemPrompt
- **verify-later:** composeSystemPrompt refusal enforcement; pack.limits guards

### Isolated chat environment (satellite; load/hack/bug vectors)
- **category:** NEW:site-chatbot
- **status-signal:** aspirational
- **status-evidence:** "Current lean (not committed — kept open): Option Y-copy ... experiment in a sandbox." Explicitly undecided.
- **what:** A plan to run the chatbot's server-side pieces (turn store, drain, analytics, optionally chat workflow code) on infrastructure separate from the core build cluster, severing three blast vectors — load (turn write-load), hack (compromised edge worker's reachable radius), bug (chat code faulting the shared chassis). Deliberately does NOT reuse the coupled multi-cluster dispatch (which shares core Kafka/Postgres). Option X = minimal satellite (maybe no chassis at all); Option Y = full cut-down chassis (Y-copy config-only vs Y-slim purpose-built image). Boundary is one-directional, async, egress-from-core only.
- **sources:** docs025.../PLAN_isolated_chat_environment(4).md, docs025.../PLAN_isolated_chat_environment(1).md
- **relations:** remote-job-spawner (NOT reused), site_chat_turns, boundary contract, building-as-a-service
- **verify-later:** any separate chat cluster/DB; isolated-DB variant of migration 086

### Building-and-hosting-as-a-service via chat (recursive platform)
- **category:** business-strategy
- **status-signal:** aspirational
- **status-evidence:** "Recorded because it sharply reframes the satellite ... (Discussion artefact; revisit as it firms up.)"
- **what:** A worked example where a chat box on design.co.uk becomes the intake+orchestration front-end to the whole build platform offered as a service: conversational briefing (a briefing-agent interview replacing the static form) → satellite intake orchestrator → full build workflows on the satellite → hybrid S3+lambda hosting → the new site itself gets a chatbot (recursion). Requires the full chassis on the satellite (rules out Option X for this use case) and surfaces new SaaS concerns: cost/abuse gating, accounts/billing/quotas, feeding reusable building blocks one-directionally from core.
- **sources:** docs025.../PLAN_isolated_chat_environment(4).md#12
- **relations:** isolated chat environment (Y-copy), briefing-agent/intake orchestrator, commercial model, simple paid multi-domain chat
- **verify-later:** briefing-agent conversational mode; satellite intake orchestrator

### Commercial model + entitlement seams (billing adapter)
- **category:** payments
- **status-signal:** partial
- **status-evidence:** "billing/identity is mostly reuse, not new" — the auth service already has a `subscriptions` table with `stripe_customer_id`, tier definitions, JWT carrying client_id+tier; "live checkout-session creation and webhooks were not evident ... verify before relying."
- **what:** The saleability design: operator-primary (operate thousands of domains), vendor-optional (sell a domain + its backend, rarely the whole framework). Isolation unit = the satellite; separability unit = the domain (partition by site_id/domain, extractable + swappable credentials). Seams to honour now: ownership via existing clients→networks→sites hierarchy (re-parent network_id to sell), a pluggable billing adapter (Stripe first, generalise stripe_* columns to provider_*), two entitlement gates (build-submission reusing site_work_items.approval_mode → a pending_entitlement hold; maintenance-run filtering the heartbeat site-selection queries as a cost valve), a saas_cheap-vs-portfolio build-tier riding the existing batch/sync rail, and snapshot-able building blocks for whole-instance sales.
- **sources:** docs025.../PLAN_isolated_chat_environment(4).md#13, docs025.../PLAN_simple_paid_multidomain_chat(1).md#2
- **relations:** auth-service subscriptions, site_work_items.approval_mode, batch processing (scheduled→batch), building-as-a-service
- **verify-later:** auth subscriptions table + Stripe webhook wiring; site_work_items.approval_mode; heartbeat site-selection queries

### Simple paid multi-domain chat (freemium + day-pass)
- **category:** NEW:site-chatbot
- **status-signal:** aspirational
- **status-evidence:** "Status: discussion draft — direction firming."
- **what:** A deliberately simple "fast lane" route: the FOCUS edge worker + a light paywall, multi-tenant-by-Host, add a domain by publishing config + DNS (no chassis/Kafka/satellite). Monetisation is freemium + a flat day-pass (£2-5) rather than counted credits, because card processing's fixed ~20-30p fee makes sub-£5 one-off charges poor. Entitlement is a stateless signed `{domain, expiry}` token issued via a synchronous Stripe guest-checkout `redeem` (no accounts, no webhook on the critical path, no edge KV). The real cost driver is the free taster + abuse, not paying users.
- **sources:** docs025.../PLAN_simple_paid_multidomain_chat(1).md, docs025.../PLAN_simple_paid_multidomain_chat.md
- **relations:** edge worker, context pack, chat lanes (fast lane), commercial model/billing adapter
- **verify-later:** paywall gate + redeem endpoint; day-pass token signing/validation

### Chat lanes (fast/slow/job) + warm-adapter maturation
- **category:** NEW:site-chatbot
- **status-signal:** aspirational
- **status-evidence:** Section 11 "What the agent framework buys chat (lanes and maturation)"; "This is still open by design; needs further conversation."
- **what:** A model splitting chat by what it does: fast lane (bounded Q&A, synchronous/streamed, no framework — ships independently); slow lane (turns needing work — live research, structured-data queries, running a site's tool, in-answer charts, multi-step tasks — routed by a cheap intent classifier, user warned it's slower); job lane (long-running submissions like "build me a site", ack + status + deliver). Maturation path: prove a slow-lane capability as a spawned agent (~12s cold), promote popular ones to warm adapters, end-state a resident chat-orchestrator adapter that fans out without spawning per turn.
- **sources:** docs025.../PLAN_isolated_chat_environment(4).md#11
- **relations:** simple paid multi-domain chat (fast lane), building-as-a-service (job lane), warm adapters
- **verify-later:** intent classifier; any resident chat-orchestrator adapter

### Payable-differentiator framework (asset × AI capability)
- **category:** business-strategy
- **status-signal:** aspirational
- **status-evidence:** "This is the method ... It is a starting point and needs more thought and testing — the menus below are incomplete, the scoring is rough."
- **what:** A method for justifying a paid chat when there's no proprietary data: value comes not from the model (everyone has it) but from a hard-to-reproduce asset (proprietary/paid data feed, an owned process/output, a well-built tool, a commercial partnership, or early access to a new AI capability) combined with AI for a paying audience. Maintain two menus (assets; AI-capabilities-worth-using-now) and pair one of each per domain. Worked examples: websitedesign.com (package our site-spec/plan as a starter prompt for Bolt/Lovable — strongest), gaswholesalers.com (buy oil/gas data feeds), agritec.uk (partnership vouchers). Prioritise by reproducibility, willingness-to-pay, build cost, cross-domain reuse; test willingness-to-pay before building.
- **sources:** docs025.../PLAN_simple_paid_multidomain_chat(1).md#10
- **relations:** simple paid multi-domain chat, ideation agent, verticals designed
- **verify-later:** n/a (strategy)

### Chat differentiator ideation agent
- **category:** research-agents
- **status-signal:** aspirational
- **status-evidence:** "A low-risk, internal use of the agent framework ... Also still needs work — treat the output as candidates, not commitments."
- **what:** A proposed internal agent (runs on our own data, no isolation concerns) that, given a domain + audience, runs the asset × AI-capability combination and proposes ranked candidate payable differentiators split into "test now (cheap)" vs "score/consider (expensive)", each naming the asset and capability it depends on. Can spawn sub-agents to research willingness-to-pay or check whether a data feed exists/what it costs. Re-runnable across all domains whenever a new AI capability is added — the mechanism for catching early-adopter opportunities. Idea generator feeding human judgement, not an automated builder.
- **sources:** docs025.../PLAN_simple_paid_multidomain_chat(1).md#11
- **relations:** payable-differentiator framework, research-agents
- **verify-later:** any ideation/differentiator agent definition

### Layer-1 / Layer-2 hack-resistance model
- **category:** system-architecture
- **status-signal:** deployed
- **status-evidence:** Stated as existing platform fact (Appendix B): "Layer 1 ... publishes outward and pulls inward but never serves inbound public traffic. Layer 2 is client delivery; today that is static assets on Backblaze S3 with nothing in the request path."
- **what:** The security posture the whole chatbot design defends: Layer 1 (core K8s cluster — agents, Kafka, Postgres, all credentials) never accepts inbound public traffic; it only publishes outward (site assets, data exports, context packs to S3) and pulls inward (recorded turns). Layer 2 is static-on-S3 with nothing in the request path — "nothing to compromise." The edge worker is the only new Layer-2 compute, and the whole chatbot design is arranged to preserve this (no API keys in the page, no central VM in front of static content). Sister-project appendix documents the "nginx box keeps getting hacked" experience that motivates it.
- **sources:** docs025.../FOCUS_site_chatbot_edge_worker_and_context_pack(1).md#1, #Appendix-B, docs025.../PLAN_isolated_chat_environment(4).md#Appendix
- **relations:** edge worker, isolated chat environment boundary contract, deploy-to-S3 path
- **verify-later:** ingress rules on ai-persona-system; B2 publish path

### maintenance_queue as generic install/uninstall trigger surface
- **category:** content-governance
- **status-signal:** partial
- **status-evidence:** Chatbot design reuses it: "Installation is requested by enqueuing a maintenance task — task_type='install_chat' on the existing maintenance_queue (which already has site_id, payload, status, retries)."
- **what:** The recognition that the existing `maintenance_queue` (built for page rebuilds) is a reusable, generic trigger surface for opt-in site add-ons — chat install/uninstall being the first — without touching the build pipeline. Establishes a pattern: new per-site capabilities enqueue a maintenance task rather than becoming a build-pipeline stage. (Cross-cut between docs019 infra and docs025 chatbot.)
- **sources:** docs025.../FOCUS_site_chatbot_edge_worker_and_context_pack(1).md#install-path, docs019_business/016_maintenance_queue_table.sql
- **relations:** maintenance_queue, site-chat-installer, maintenance-triage
- **verify-later:** maintenance_queue task_type values in use

### prepare_extraction_context / scan_discovery_candidates actions
- **category:** NEW:business-intelligence-platform
- **status-signal:** partial
- **status-evidence:** Wired into vet-practice-verifier via UPDATE migrations adding prepare_context and scan_discoveries steps; "reuses the skipDomains map from scan_discovery_candidates.go".
- **what:** Two supporting Go actions in the vet pipeline: `prepare_extraction_context` formats search results + scraped content (max_content_length/max_snippets) into a clean `extraction_context` for the LLM step; `scan_discovery_candidates` scans a verifier's search results for unknown practices (skipping aggregator domains) and inserts them into discovery_candidates. Both illustrate the "complexity in Go actions, thin workflow" convention.
- **sources:** docs019_business/004_vet_practice_verifier.sql#prepare_context, #scan_discoveries, docs019_business/011_area_sweep_discovery_system.md
- **relations:** vet-practice-verifier, discovery_candidates, area-sweep process_area_sweep (shares skipDomains)
- **verify-later:** actions prepare_extraction_context, scan_discovery_candidates; skipDomains map

### Agent hierarchy tree navigation (ltree paths + subtree summaries + live viewer)
- **category:** NEW:agent-tree-navigation
- **status-signal:** aspirational
- **status-evidence:** Raw design-session transcript only ("The data model changes are small... The bigger piece of work is the API endpoints and the tree viewer UI"); no implementation claimed, buried inside a 273KB chat-transcript file the rest of the extraction treats as header-scan.
- **what:** A proposal for navigating the `orchestration_states` parent/child tree at massive scale (millions of rows, 8-10 levels deep) without recursive-CTE cost: add an `ltree`-typed `tree_path` column (materialised ancestry path, set cheaply at spawn time by prepending the parent's own path), enrich the existing `subtree_agents` jsonb with rolling status/type/failure counts so a UI can show summaries and only fetch detail on expand, add a `tags` jsonb column (GIN-indexed) for semantic queries ("find all bankrupt fast-food agents" rather than tree position), and a lightweight `agent_tree_index` table (~200 bytes/row, no heavy jsonb blobs) so a million-row tree fits comfortably in cache. Proposed REST API (`/trees/{correlation_id}`, `/agents/{id}/children`, `/agents/{id}/subtree`, `/trees/{id}/search?agent_type=...&status=...`) plus a WebSocket live tree viewer fed from existing Kafka response topics, giving filesystem-like drill-down ("root > uk-economy > fast-food-sector > dominos-agent-47").
- **sources:** docs021_multiclustering/021_2026-02-28-20-03-32-multi-cluster-dispatch-design.txt (sections "The fundamental query patterns" through "The user experience")
- **relations:** Multi-cluster scaling tiers, orchestration_states schema, Agent swarm simulation ideas (this viewer was requested specifically to make the swarm-simulation ideas practically navigable)
- **verify-later:** orchestration_states.tree_path / tags columns; any agent_tree_index table; core-manager tree API endpoints

### Agent swarm simulation ideas (never built — hierarchical/fractal use-case brainstorm)
- **category:** NEW:agent-swarm-simulations
- **status-signal:** aspirational
- **status-evidence:** Pure ideation inside a raw chat transcript; closing exchange asks to "report all your ideas into a document... I'd like to use the document as a web page of use-cases... to try and get people interested in triggering their own project ideas" — recorded as marketing/pitch material, not a build plan.
- **what:** A large brainstormed catalogue of speculative applications for the platform's hierarchical spawn/call architecture at extreme scale (up to 1M agents), produced across two rounds. Flat-swarm ideas: an LLM-agent economy simulation (the author's top pick — emergent price equilibrium/monopolies, visually renderable), collaborative Wikipedia cross-fact-checking, a million-agent code-review swarm auditing a large codebase, emergent-language formation between agents with no shared vocabulary, distributed collaborative micro-fiction, adversarial red-team-vs-blue-team war-gaming. Hierarchy-specific ideas exploiting the platform's distinguishing trait (every parent/child stores its result independently, decomposition is semantically meaningful): recursive market-research report trees, organisational/corporate-directive-cascade simulation, fractal ecological modelling (e.g. the Amazon basin), legislation-impact analysis trees, a hierarchical self-debugging swarm for the platform's own Kafka/K8s/Postgres stack, scientific-literature mapping into a queryable tree, supply-chain stress testing, evolutionary/genetic idea-generation trees, plus a further batch: historical counterfactual simulation, language-family evolution trees, musical composition by fractal decomposition, personal health-trajectory modelling, adversarial peer-review simulation, M&A due-diligence trees, climate-migration modelling, argument/debate mapping, ecosystem-succession modelling, judicial case-reasoning trees, disaster-response command-structure coordination, and personal knowledge management. None were built or scoped into a plan.
- **sources:** docs021_multiclustering/021_2026-02-27-18-19-36-million-agent-scaling-plan.txt (the exchanges following "what other impressive things could we do with 1M vaguely intelligent agents?" and "...for the hierarchical/fractal model that I currently have")
- **relations:** Agent hierarchy tree navigation, Multi-cluster scaling tiers (10K/100K/1M), Worker pool architecture
- **verify-later:** n/a (pure ideation, no code or schema artefact)
