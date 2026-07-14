# Register — rag-knowledge-base

6 concepts, consolidated from 6 raw extractions across units U22, U25. No
duplicates found within this category's raw material (each raw block covered a
distinct facet: the store, the two actions, the best-practices doctrine, the
deferred owning-agent, and a deferred future consumer). Note for the
consolidator reading other clusters: closely related material was tagged
model-infrastructure and finetuning-flywheel by several units (e.g. an
almost-identical "RAG best practices" doc and a "RAG pipeline deployment"
bundle) — those raw blocks were left in their originally-tagged category
registers per the per-category assignment rule, but they describe the same
real-world knowledge_base/rag_lookup/rag_index mechanism documented here and
should be read alongside RAGK-001/RAGK-004 rather than as independent
infrastructure.

### RAGK-001 — RAG knowledge_base (shared pgvector store)
- **status:** deployed
- **status-evidence:** migration 082 marked idempotent/deployed in the vertical-architecture handoff; live repo confirms `rag_actions.go` and related migrations exist and are registered.
- **what:** A shared (not per-agent) `knowledge_base` table storing chunked content with a `vector(768)` embedding (nomic-embed-text), collection/industry/domain classification, SHA256 dedup, an ivfflat cosine index, and a trigram fallback index for when the embedding provider is down. Any agent reads via `rag_lookup` and writes via `rag_index`; later extended with source-provenance columns. The supporting `ollama-adapter` k8s service (own kustomize base, `ollama/ollama` image, PVC-persisted nomic-embed-text pulled by an init container) is the embedding provider this table depends on.
- **sources:** docs020_llm_training_rag/004_rag_knowledge_base.sql; docs020_llm_training_rag/010_simple_explanation.md; docs020_llm_training_rag/008_README.md; docs/_archive/agent_docs/docs020_llm_training_rag/007_rag_deployment_README.md (deployment bundle: migrations, actions, ollama-adapter kustomize, patches to ai_actions.go/registry.go/anthropic.go)
- **relations:** rag_lookup (RAGK-002), rag_index (RAGK-003), Ollama adapter (model-infrastructure), vertical knowledge architecture, finetuning-flywheel's "knowledge_base RAG store + Flywheel B verification" entry (same table, flywheel-lane framing)
- **verify-later:** knowledge_base table + idx_kb_embedding (ivfflat) + idx_kb_content_trgm; knowledge_base_stats view; platform/orchestration/actions/rag_actions.go; deployments/kustomize/services/ollama-adapter/

### RAGK-002 — rag_lookup action (vector search + trigram fallback)
- **status:** partial
- **status-evidence:** action code written and a registry patch documented ("NEEDS PATCH — add 2 rag entries") in the vertical-architecture handoff; not confirmed applied at time of extraction.
- **what:** An action that embeds the query via Ollama, runs pgvector cosine similarity within a collection, and returns both structured `rag_results` and a combined `rag_context` string for prompt injection; falls back to Postgres trigram text search when Ollama is down (reported in `search_method`). Best practice: filter by metadata (vertical/component/quality) before ranking, and prepend the `search_query:` task prefix.
- **sources:** docs020_llm_training_rag/010_simple_explanation.md#rag_lookup; docs020_llm_training_rag/012b_rag_best_practices_v2.md; docs020_llm_training_rag/005_PATCHES.md#patch-03
- **relations:** rag_index (RAGK-003), knowledge_base (RAGK-001), content-writer RAG injection, nomic task-prefix bug (finetuning-flywheel)
- **verify-later:** GlobalActionRegistry entry rag_lookup; RAGLookupAction min_authority/filter support

### RAGK-003 — rag_index action (chunk, embed, dedup, store)
- **status:** partial
- **status-evidence:** new file `rag_actions.go` documented as "ready to add"; registry patch pending at time of extraction.
- **what:** An action that splits text into chunks (default ~1000 chars, 200 overlap, sentence-boundary), SHA256-hashes each for dedup, embeds via Ollama, and inserts into `knowledge_base` tagged by collection/metadata. If embedding fails, the chunk is still stored (searchable via trigram fallback). Intended to accept source_authority/vertical_slug/knowledge_type once the schema is extended.
- **sources:** docs020_llm_training_rag/010_simple_explanation.md#rag_index; docs020_llm_training_rag/012b_rag_best_practices_v2.md#implementation-priority
- **relations:** rag_lookup (RAGK-002), knowledge-indexer agent (RAGK-005), vertical research handler
- **verify-later:** GlobalActionRegistry entry rag_index; RAGIndexAction dedup on collection+content_hash

### RAGK-004 — RAG best practices — filter-first, quality gating, token budget
- **status:** aspirational
- **status-evidence:** dated 2026-03-24 best-practices doc; its own "Implementation Priority" section is a to-do list (add metadata columns, update actions) — i.e. doctrine documented ahead of full implementation.
- **what:** A methodology for the site-build RAG: always filter by structured metadata (vertical, component_type, source_quality) before embedding-similarity ranking; keep RAG context to 20-30% of the window and 3-5 examples; gate entries by source_quality (high/verified) for prompt injection; track embedding_model and never mix embedding spaces; prepend nomic task prefixes (search_document/search_query); prefer nomic-embed-text-v2-moe. Names five common RAG failure modes and their fixes. An earlier v1 of this doc (dated the same day, in the archive tree) was superseded by this v2.
- **sources:** docs020_llm_training_rag/012b_rag_best_practices_v2.md
- **relations:** rag_index (RAGK-003), rag_lookup (RAGK-002), knowledge sources (scraped/claude-output/human-curated/audit-insight), model-infrastructure's superseded "RAG best practices (filter-first-then-rank)" v1 entry (same doctrine, older doc)
- **verify-later:** knowledge_base metadata columns (vertical/component_type/source_quality); task-prefix handling in rag actions

### RAGK-005 — knowledge-indexer agent (deferred)
- **status:** aspirational
- **status-evidence:** "Future agent (owns the knowledge-building domain): knowledge-indexer agent ... For now, we implement the actions. The agent comes when we have a use case."
- **what:** A proposed but deliberately-unbuilt agent that would own the knowledge-building process (load indexing targets → web_scrape → rag_index → refresh), called by the maintenance orchestrator or build pipeline. Held back per the "reuse before creating — don't build an agent until the workflow demands one" principle; the rag_index/rag_lookup actions suffice for now.
- **sources:** docs020_llm_training_rag/001_rag_agent_distribution_architecture.md#item-4
- **relations:** rag_index (RAGK-003), vertical research handler (later realises this role)
- **verify-later:** agent_definitions for any knowledge-indexer/vertical-research-handler

### RAGK-006 — Concept-document RAG for content writers (v2+)
- **status:** aspirational
- **status-evidence:** strategic-planning doc "Why not RAG for v1?" + "RAG integration (v2+)" section; a companion doc drops RAG to "Offline reference (not ingested for v1) … agents in v2+ via RAG".
- **what:** A deferred design: when content surface outgrows page-level content_context, ingest the full concept document into the knowledge_base via the existing `rag_index` action and add a `rag_lookup` step to the content-writer workflow (query built from page slug + section function, results fed into the prompt). Deliberately not built for a 5-page v1 — the structured mission/roadmap fields already carry enough context.
- **sources:** docs/social001_vonc_tiktok_social/003c_spark_strategic_planning_architecture(2).md#Why-not-RAG, #RAG-integration; docs/social001_vonc_tiktok_social/003d_spark_strategic_planning_architecture.md#What-goes-where
- **relations:** mission/roadmap aspects; documentation-system; rag_index/rag_lookup (RAGK-002/003)
- **verify-later:** rag_index/rag_lookup actions; knowledge_base table
