# Register — rag-retrieval

1 concept, consolidated from 2 raw extractions across units U18_sql_for_agents, U19_sql_tables_components.

Note on source material: the assigned cluster input file contained this entire category duplicated byte-for-byte exactly twice (a mechanical bucketing artifact, verified via diff). Counts and merges below are computed on the de-duplicated set of 2 real raw blocks, which turned out to describe the same table from two different SQL-doc directories (each with their own independent 001-numbering, hence the apparently conflicting migration numbers below — not a real conflict).

---

### RAGR-001 — knowledge_base: shared pgvector RAG store (rag_index/rag_lookup)
*(merged from 2 raw blocks describing the same table from two different SQL-migration directories)*
- **status:** deployed
- **status-evidence:** 041_rag_knowledge_base.sql (sql_for_agents numbering) creates the table; 105_rag_test_agent.sql verifies chassis registration; 141_reenable_index_plan_after_chunk_fix.sql lands the first real tool_docs rows after a chunk-loop saga. Independently, migration 082 (sql_for_tables numbering, idempotent) adds pgvector + pg_trgm, and a later migration (048) confirms live extension versions on clients_db (vector 0.8.0) and describes knowledge_base as the "proven SHAPE".
- **what:** A shared (not per-agent) embedded knowledge store for scraped exemplar sites, research, curated industry info, component usage patterns, and marketing/industry content chunks — queryable by any content-creating agent. Rows carry collection + industry + domain classification, content with content_hash dedup per collection, vector(768) embeddings (nomic-embed-text via the ollama-adapter) with an IVFFlat cosine index, and a trigram GIN fallback for keyword retrieval when embeddings are unavailable. Collections partition use-cases (industry_sites, research, components, tool_docs, flywheel_b_chassis_test). Additional lifecycle columns track source, quality_score, and usage_count, with a stats view over them. An embedding-model column tracks provenance; changing embedding dimensions requires a column ALTER + reindex. code_symbols is a sibling store with the same shape for a different domain (code, not content).
- **sources:** 041_rag_knowledge_base.sql; 105_rag_test_agent.sql; 141_reenable_index_plan_after_chunk_fix.sql; docs/agent_docs/sql_for_tables/025_llm_call_log_rag_knowledge_base.sql#082; docs/agent_docs/sql_for_tables/048_NNN_create_code_symbols_index.sql#WHY-A-SIBLING
- **relations:** travelling docs workstream (doc_plans indexed into tool_docs — an active workstream per project memory); rag_actions.go; code-indexer (separate code_symbols store); ollama-adapter embedder; Component creation contract (CTS-057, notes knowledge_base + rag_lookup as the proposed future storage location for the component-creation contract itself if it starts to churn)
- **verify-later:** knowledge_base row counts per collection; rag_lookup consumers; collections actually in use; retrieval actions
