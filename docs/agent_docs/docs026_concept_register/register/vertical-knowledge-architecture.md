# Register — vertical-knowledge-architecture

4 concepts, consolidated from 8 raw extractions across unit U22. (The cluster input file contained this category's raw blocks twice, back-to-back and byte-identical; each pair is merged into one entry below. No cross-unit duplication found — all raw blocks for this category came from U22_recent_small_docs.md.)

### VKA-001 — Vertical knowledge architecture (overview)
- **status:** aspirational
- **status-evidence:** "The architecture ... is now designed but not yet fully implemented"; implementation todo is phased 0-8 with only Phase 0 partially begun.
- **what:** The strategic pivot from a flat build pipeline to routing each domain to a specialised vertical (veterinary, energy_wholesale, finance_mortgage, seasonal_gifts, generic) that maintains its own knowledge-base collection, research strategy, content patterns, and monetisation config. Verticals are logical first (shared infra, knowledge_base collections + rag_lookup/index), physical later (dispatch_agent). Compounding moat: the tenth domain benefits from the first nine's research.
- **sources:** docs021.../020_vertical_cluster_architecture.md; docs021.../025_session_handoff_vertical_architecture.md; docs020.../011_where_we_are.md
- **relations:** VKA-002 vertical_registry, VKA-003 research/build cluster separation, RAG knowledge_base, site classifier vertical output, BIZ-022 verticals designed
- **verify-later:** vertical_registry table; agent_definitions tagged 'vertical-orch'; knowledge_base collection usage

### VKA-002 — vertical_registry table + knowledge-base provenance extensions
- **status:** aspirational
- **status-evidence:** "These additions to the database, not yet applied"; seed INSERT for 5 verticals given as a to-do.
- **what:** A `vertical_registry` table mapping vertical_slug → research/build orchestrator types, knowledge_collection, research_sources, content_patterns/page_type_library, monetisation_config, refresh_schedule, and maturity_stage; plus knowledge_base column extensions (source_authority 1-5, source_url, source_date, vertical_slug, knowledge_type) for provenance-weighted retrieval. Seeds veterinary/energy/mortgage/seasonal/generic.
- **sources:** docs021.../020_vertical_cluster_architecture.md#8, docs021.../026_implementation_todo_vertical_architecture(2).md#2.1
- **relations:** VKA-001 vertical knowledge architecture, rag_lookup min_authority, BIP-002 business_verticals (different table, same idea)
- **verify-later:** vertical_registry table; knowledge_base.source_authority/vertical_slug/knowledge_type columns

### VKA-003 — Research/build cluster separation
- **status:** aspirational
- **status-evidence:** "Phase 2: Physical separation ... implementation deferred"; described as designed, Phase 8 in the todo.
- **what:** A two-cluster model separating messy/slow/shared research (web scraping, PDF parsing, LLM knowledge extraction, validation, indexing) from structured/fast/per-site build. They communicate only via the shared knowledge_base (Postgres reads/writes) and Kafka orchestration; build dispatches a research request when it hits a knowledge gap. Justified by independent scaling, failure isolation, clean logs, tighter build-cluster network policy, and clean cost attribution.
- **sources:** docs021.../020_vertical_cluster_architecture.md#5, docs021.../025_session_handoff_vertical_architecture.md#layer-5
- **relations:** VKA-001 vertical knowledge architecture, DispatchAgentAction, VKA-004 vertical research handler
- **verify-later:** any research-cluster agent definitions; dispatch_agent used for research requests

### VKA-004 — Vertical research handler + knowledge accumulation loop
- **status:** aspirational
- **status-evidence:** "Phase 4 (Research Handler)" is an unchecked todo; `needs_vertical_research` work-item type not yet in the schema.
- **what:** A `vertical-research-handler` agent that processes `needs_vertical_research` work items (identify sources → scrape → parse → LLM extract structured knowledge chunks → validate quality/confidence → rag_index with source_authority/vertical_slug). Realises the knowledge-accumulation loop: first domain bears foundational research cost, gaps become research items at priority 1-4 that content items (priority 10-17) depend on, and indexed knowledge benefits all future domains in the vertical.
- **sources:** docs021.../026_implementation_todo_vertical_architecture(2).md#phase-4, docs021.../020_vertical_cluster_architecture.md#6
- **relations:** VKA-003 research/build separation, rag_index, work-item lifecycle, knowledge-indexer agent
- **verify-later:** needs_vertical_research item type; vertical-research-handler agent; check_knowledge_coverage action
