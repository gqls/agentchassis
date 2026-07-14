
<!-- SOURCE: U18_sql_for_agents.md -->
### RAG knowledge base (shared pgvector store) and rag_index/rag_lookup
- **category:** NEW:rag-retrieval
- **status-signal:** deployed
- **status-evidence:** 041 creates knowledge_base (vector(768), nomic-embed-text, content_hash dedup); 105 rag-test-agent verifies chassis registration; 141 finally lands first tool_docs rows after the chunk-loop saga.
- **what:** Shared (not per-agent) embedded knowledge store for scraped exemplar sites, research, curated industry info and component usage patterns, queryable by any content-creating agent. Collections partition use-cases (industry_sites, research, components, tool_docs, flywheel_b_chassis_test). Embedding-model column tracks provenance; changing dimensions requires column ALTER + reindex.
- **sources:** 041_rag_knowledge_base.sql; 105_rag_test_agent.sql; 141_reenable_index_plan_after_chunk_fix.sql
- **relations:** travelling docs (doc_plans indexed into tool_docs); rag_actions.go; code-indexer (separate code_symbols store)
- **verify-later:** knowledge_base row counts per collection; rag_lookup consumers

<!-- SOURCE: U19_sql_tables_components.md -->
### knowledge_base RAG store
- **category:** NEW:rag-retrieval
- **status-signal:** deployed
- **status-evidence:** Migration 082 (idempotent) with pgvector + pg_trgm; 048 later confirms live extension versions on clients_db (vector 0.8.0) and describes knowledge_base as the "proven SHAPE".
- **what:** Industry/marketing content chunks for retrieval: collection + industry + domain classification, content with content_hash dedup per collection, vector(768) embeddings (nomic-embed-text via ollama-adapter) with IVFFlat cosine index, trigram GIN fallback for keyword retrieval when embeddings are unavailable, source tracking, quality_score and usage_count lifecycle, stats view.
- **sources:** docs/agent_docs/sql_for_tables/025_llm_call_log_rag_knowledge_base.sql#082; docs/agent_docs/sql_for_tables/048_NNN_create_code_symbols_index.sql#WHY-A-SIBLING
- **relations:** code_symbols (sibling shape); ollama-adapter embedder; content grounding.
- **verify-later:** collections in use; retrieval actions.

<!-- SOURCE: U18_sql_for_agents.md -->
### RAG knowledge base (shared pgvector store) and rag_index/rag_lookup
- **category:** NEW:rag-retrieval
- **status-signal:** deployed
- **status-evidence:** 041 creates knowledge_base (vector(768), nomic-embed-text, content_hash dedup); 105 rag-test-agent verifies chassis registration; 141 finally lands first tool_docs rows after the chunk-loop saga.
- **what:** Shared (not per-agent) embedded knowledge store for scraped exemplar sites, research, curated industry info and component usage patterns, queryable by any content-creating agent. Collections partition use-cases (industry_sites, research, components, tool_docs, flywheel_b_chassis_test). Embedding-model column tracks provenance; changing dimensions requires column ALTER + reindex.
- **sources:** 041_rag_knowledge_base.sql; 105_rag_test_agent.sql; 141_reenable_index_plan_after_chunk_fix.sql
- **relations:** travelling docs (doc_plans indexed into tool_docs); rag_actions.go; code-indexer (separate code_symbols store)
- **verify-later:** knowledge_base row counts per collection; rag_lookup consumers

<!-- SOURCE: U19_sql_tables_components.md -->
### knowledge_base RAG store
- **category:** NEW:rag-retrieval
- **status-signal:** deployed
- **status-evidence:** Migration 082 (idempotent) with pgvector + pg_trgm; 048 later confirms live extension versions on clients_db (vector 0.8.0) and describes knowledge_base as the "proven SHAPE".
- **what:** Industry/marketing content chunks for retrieval: collection + industry + domain classification, content with content_hash dedup per collection, vector(768) embeddings (nomic-embed-text via ollama-adapter) with IVFFlat cosine index, trigram GIN fallback for keyword retrieval when embeddings are unavailable, source tracking, quality_score and usage_count lifecycle, stats view.
- **sources:** docs/agent_docs/sql_for_tables/025_llm_call_log_rag_knowledge_base.sql#082; docs/agent_docs/sql_for_tables/048_NNN_create_code_symbols_index.sql#WHY-A-SIBLING
- **relations:** code_symbols (sibling shape); ollama-adapter embedder; content grounding.
- **verify-later:** collections in use; retrieval actions.
