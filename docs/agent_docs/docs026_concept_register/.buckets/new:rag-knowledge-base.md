
<!-- SOURCE: U22_recent_small_docs.md -->
### RAG knowledge_base (shared pgvector store)
- **category:** NEW:rag-knowledge-base
- **status-signal:** deployed
- **status-evidence:** "knowledge_base table is sitting empty in Postgres, waiting for content ... ivfflat index"; migration 082 marked idempotent/deployed in the vertical-architecture handoff.
- **what:** A shared (not per-agent) `knowledge_base` table storing chunked content with a `vector(768)` embedding (nomic-embed-text), collection/industry/domain classification, SHA256 dedup, ivfflat cosine index, and a trigram fallback index. Any agent reads via `rag_lookup`, any writes via `rag_index`. Later extended with source provenance columns (docs021).
- **sources:** docs020.../004_rag_knowledge_base.sql, docs020.../010_simple_explanation.md, docs020.../008_README.md
- **relations:** rag_lookup, rag_index, Ollama provider, vertical knowledge architecture
- **verify-later:** knowledge_base table + idx_kb_embedding (ivfflat) + idx_kb_content_trgm; knowledge_base_stats view

<!-- SOURCE: U22_recent_small_docs.md -->
### rag_lookup action (vector search + trigram fallback)
- **category:** NEW:rag-knowledge-base
- **status-signal:** partial
- **status-evidence:** Registry patch written ("NEEDS PATCH — add 2 rag entries"); action code written but registry.go patch listed as not-yet-applied in the handoff.
- **what:** An action that embeds the query via Ollama, runs pgvector cosine similarity within a collection, and returns both structured `rag_results` and a combined `rag_context` string for prompt injection; falls back to Postgres trigram text search when Ollama is down (reported in `search_method`). Best practice: filter by metadata (vertical/component/quality) before ranking, and prepend `search_query:` task prefix.
- **sources:** docs020.../010_simple_explanation.md#rag_lookup, docs020.../012b_rag_best_practices_v2.md, docs020.../005_PATCHES.md#patch-03
- **relations:** rag_index, knowledge_base, content-writer RAG injection
- **verify-later:** GlobalActionRegistry entry rag_lookup; RAGLookupAction min_authority/filter support

<!-- SOURCE: U22_recent_small_docs.md -->
### rag_index action (chunk, embed, dedup, store)
- **category:** NEW:rag-knowledge-base
- **status-signal:** partial
- **status-evidence:** New file `rag_actions.go` "ready to add"; registry patch pending; non-fatal embedding failure behaviour specified in the revised plan.
- **what:** An action that splits text into chunks (default ~1000 chars, 200 overlap, sentence-boundary), SHA256-hashes each for dedup, embeds via Ollama, and inserts into `knowledge_base` tagged by collection/metadata. If embedding fails the chunk is still stored (searchable via trigram). Intended to accept source_authority/vertical_slug/knowledge_type once schema extended.
- **sources:** docs020.../010_simple_explanation.md#rag_index, docs020.../012b_rag_best_practices_v2.md#implementation-priority
- **relations:** rag_lookup, knowledge-indexer agent, vertical research handler
- **verify-later:** GlobalActionRegistry entry rag_index; RAGIndexAction dedup on collection+content_hash

<!-- SOURCE: U22_recent_small_docs.md -->
### RAG best practices — filter-first, quality gating, token budget
- **category:** NEW:rag-knowledge-base
- **status-signal:** aspirational
- **status-evidence:** Dated 2026-03-24 best-practices doc; "Implementation Priority" is a to-do list (add metadata columns, update actions), i.e. not yet applied.
- **what:** A methodology for the site-build RAG: always filter by structured metadata (vertical, component_type, source_quality) before embedding-similarity ranking; keep RAG context to 20-30% of the window and 3-5 examples; gate entries by source_quality (high/verified) for prompt injection; track embedding_model and never mix embedding spaces; prepend nomic task prefixes (search_document/search_query); prefer nomic-embed-text-v2-moe. Names five common RAG failures and their fixes.
- **sources:** docs020.../012b_rag_best_practices_v2.md
- **relations:** rag_index, rag_lookup, knowledge sources (scraped/claude-output/human-curated/audit-insight)
- **verify-later:** knowledge_base metadata columns (vertical/component_type/source_quality); task-prefix handling in rag actions

<!-- SOURCE: U22_recent_small_docs.md -->
### knowledge-indexer agent (deferred)
- **category:** NEW:rag-knowledge-base
- **status-signal:** aspirational
- **status-evidence:** "Future agent (owns the knowledge-building domain): knowledge-indexer agent ... For now, we implement the actions. The agent comes when we have a use case."
- **what:** A proposed but deliberately-unbuilt agent that would own the knowledge-building process (load indexing targets → web_scrape → rag_index → refresh), called by the maintenance orchestrator or build pipeline. Held back per the "reuse before creating — don't build an agent until the workflow demands one" principle; the rag_index/rag_lookup actions suffice for now.
- **sources:** docs020.../001_rag_agent_distribution_architecture.md#item-4
- **relations:** rag_index, vertical research handler (later realises this role)
- **verify-later:** agent_definitions for any knowledge-indexer/vertical-research-handler

<!-- SOURCE: U25_leopardess_social.md -->
### Concept-document RAG for content writers (v2+)
- **category:** NEW:rag-knowledge-base
- **status-signal:** aspirational
- **status-evidence:** 003c "Why not RAG for v1?" + "RAG integration (v2+)"; 003d drops RAG to "Offline reference (not ingested for v1) … agents in v2+ via RAG".
- **what:** Deferred design: when content surface outgrows page-level content_context, ingest the full concept document into the knowledge_base via the existing rag_index action and add a rag_lookup step to the content-writer workflow (query built from page slug + section function, results into the prompt). Deliberately not built for a 5-page v1 — the structured mission/roadmap fields carry enough context.
- **sources:** docs/social001_vonc_tiktok_social/003c_spark_strategic_planning_architecture(2).md#Why-not-RAG, #RAG-integration (family-delta); docs/social001_vonc_tiktok_social/003d_spark_strategic_planning_architecture.md#What-goes-where
- **relations:** mission/roadmap aspects; documentation-system
- **verify-later:** rag_index/rag_lookup actions; knowledge_base table

<!-- SOURCE: U22_recent_small_docs.md -->
### RAG knowledge_base (shared pgvector store)
- **category:** NEW:rag-knowledge-base
- **status-signal:** deployed
- **status-evidence:** "knowledge_base table is sitting empty in Postgres, waiting for content ... ivfflat index"; migration 082 marked idempotent/deployed in the vertical-architecture handoff.
- **what:** A shared (not per-agent) `knowledge_base` table storing chunked content with a `vector(768)` embedding (nomic-embed-text), collection/industry/domain classification, SHA256 dedup, ivfflat cosine index, and a trigram fallback index. Any agent reads via `rag_lookup`, any writes via `rag_index`. Later extended with source provenance columns (docs021).
- **sources:** docs020.../004_rag_knowledge_base.sql, docs020.../010_simple_explanation.md, docs020.../008_README.md
- **relations:** rag_lookup, rag_index, Ollama provider, vertical knowledge architecture
- **verify-later:** knowledge_base table + idx_kb_embedding (ivfflat) + idx_kb_content_trgm; knowledge_base_stats view

<!-- SOURCE: U22_recent_small_docs.md -->
### rag_lookup action (vector search + trigram fallback)
- **category:** NEW:rag-knowledge-base
- **status-signal:** partial
- **status-evidence:** Registry patch written ("NEEDS PATCH — add 2 rag entries"); action code written but registry.go patch listed as not-yet-applied in the handoff.
- **what:** An action that embeds the query via Ollama, runs pgvector cosine similarity within a collection, and returns both structured `rag_results` and a combined `rag_context` string for prompt injection; falls back to Postgres trigram text search when Ollama is down (reported in `search_method`). Best practice: filter by metadata (vertical/component/quality) before ranking, and prepend `search_query:` task prefix.
- **sources:** docs020.../010_simple_explanation.md#rag_lookup, docs020.../012b_rag_best_practices_v2.md, docs020.../005_PATCHES.md#patch-03
- **relations:** rag_index, knowledge_base, content-writer RAG injection
- **verify-later:** GlobalActionRegistry entry rag_lookup; RAGLookupAction min_authority/filter support

<!-- SOURCE: U22_recent_small_docs.md -->
### rag_index action (chunk, embed, dedup, store)
- **category:** NEW:rag-knowledge-base
- **status-signal:** partial
- **status-evidence:** New file `rag_actions.go` "ready to add"; registry patch pending; non-fatal embedding failure behaviour specified in the revised plan.
- **what:** An action that splits text into chunks (default ~1000 chars, 200 overlap, sentence-boundary), SHA256-hashes each for dedup, embeds via Ollama, and inserts into `knowledge_base` tagged by collection/metadata. If embedding fails the chunk is still stored (searchable via trigram). Intended to accept source_authority/vertical_slug/knowledge_type once schema extended.
- **sources:** docs020.../010_simple_explanation.md#rag_index, docs020.../012b_rag_best_practices_v2.md#implementation-priority
- **relations:** rag_lookup, knowledge-indexer agent, vertical research handler
- **verify-later:** GlobalActionRegistry entry rag_index; RAGIndexAction dedup on collection+content_hash

<!-- SOURCE: U22_recent_small_docs.md -->
### RAG best practices — filter-first, quality gating, token budget
- **category:** NEW:rag-knowledge-base
- **status-signal:** aspirational
- **status-evidence:** Dated 2026-03-24 best-practices doc; "Implementation Priority" is a to-do list (add metadata columns, update actions), i.e. not yet applied.
- **what:** A methodology for the site-build RAG: always filter by structured metadata (vertical, component_type, source_quality) before embedding-similarity ranking; keep RAG context to 20-30% of the window and 3-5 examples; gate entries by source_quality (high/verified) for prompt injection; track embedding_model and never mix embedding spaces; prepend nomic task prefixes (search_document/search_query); prefer nomic-embed-text-v2-moe. Names five common RAG failures and their fixes.
- **sources:** docs020.../012b_rag_best_practices_v2.md
- **relations:** rag_index, rag_lookup, knowledge sources (scraped/claude-output/human-curated/audit-insight)
- **verify-later:** knowledge_base metadata columns (vertical/component_type/source_quality); task-prefix handling in rag actions

<!-- SOURCE: U22_recent_small_docs.md -->
### knowledge-indexer agent (deferred)
- **category:** NEW:rag-knowledge-base
- **status-signal:** aspirational
- **status-evidence:** "Future agent (owns the knowledge-building domain): knowledge-indexer agent ... For now, we implement the actions. The agent comes when we have a use case."
- **what:** A proposed but deliberately-unbuilt agent that would own the knowledge-building process (load indexing targets → web_scrape → rag_index → refresh), called by the maintenance orchestrator or build pipeline. Held back per the "reuse before creating — don't build an agent until the workflow demands one" principle; the rag_index/rag_lookup actions suffice for now.
- **sources:** docs020.../001_rag_agent_distribution_architecture.md#item-4
- **relations:** rag_index, vertical research handler (later realises this role)
- **verify-later:** agent_definitions for any knowledge-indexer/vertical-research-handler

<!-- SOURCE: U25_leopardess_social.md -->
### Concept-document RAG for content writers (v2+)
- **category:** NEW:rag-knowledge-base
- **status-signal:** aspirational
- **status-evidence:** 003c "Why not RAG for v1?" + "RAG integration (v2+)"; 003d drops RAG to "Offline reference (not ingested for v1) … agents in v2+ via RAG".
- **what:** Deferred design: when content surface outgrows page-level content_context, ingest the full concept document into the knowledge_base via the existing rag_index action and add a rag_lookup step to the content-writer workflow (query built from page slug + section function, results into the prompt). Deliberately not built for a 5-page v1 — the structured mission/roadmap fields carry enough context.
- **sources:** docs/social001_vonc_tiktok_social/003c_spark_strategic_planning_architecture(2).md#Why-not-RAG, #RAG-integration (family-delta); docs/social001_vonc_tiktok_social/003d_spark_strategic_planning_architecture.md#What-goes-where
- **relations:** mission/roadmap aspects; documentation-system
- **verify-later:** rag_index/rag_lookup actions; knowledge_base table
