-- ============================================================================
-- Flywheel B — Step 0: Schema + infrastructure checks
-- ============================================================================
-- Run these and share the output. We don't change anything yet.
-- These are read-only; safe to run against the live clients database.
--
-- Run: kubectl -n ai-persona-system exec -it deploy/postgres-clients -- \
--        psql -U clients_user -d clients_db
-- ============================================================================


-- ── 1. knowledge_base table exists and has expected columns ──

\echo '=== 1. knowledge_base columns ==='

SELECT column_name, data_type, is_nullable, column_default
FROM information_schema.columns
WHERE table_name = 'knowledge_base'
ORDER BY ordinal_position;


-- ── 2. knowledge_base indexes ──

\echo ''
\echo '=== 2. knowledge_base indexes ==='

SELECT indexname, indexdef
FROM pg_indexes
WHERE tablename = 'knowledge_base'
ORDER BY indexname;


-- ── 3. knowledge_base current contents ──

\echo ''
\echo '=== 3. knowledge_base row count + stats ==='

SELECT count(*) as total_rows,
       count(*) FILTER (WHERE embedding IS NOT NULL) as embedded_rows,
       count(DISTINCT collection) as collections,
       count(DISTINCT embedding_model) as embedding_models
FROM knowledge_base;

\echo ''
\echo '=== 3a. knowledge_base by collection (if any rows) ==='

SELECT collection, count(*) as chunks,
       count(*) FILTER (WHERE embedding IS NOT NULL) as embedded,
       min(created_at) as oldest,
       max(created_at) as newest
FROM knowledge_base
GROUP BY collection
ORDER BY chunks DESC;


-- ── 4. pgvector extension ──

\echo ''
\echo '=== 4. pgvector + pg_trgm extensions ==='

SELECT extname, extversion
FROM pg_extension
WHERE extname IN ('vector', 'pg_trgm')
ORDER BY extname;


-- ── 5. llm_call_log recent activity (confirms logging alive) ──

\echo ''
\echo '=== 5. llm_call_log — last 24h activity ==='

SELECT count(*) as calls_24h,
       count(DISTINCT agent_type) FILTER (WHERE agent_type IS NOT NULL AND agent_type != '') as distinct_agents,
       count(*) FILTER (WHERE success) as successful,
       count(*) FILTER (WHERE NOT success) as failed
FROM llm_call_log
WHERE created_at > NOW() - INTERVAL '24 hours';


-- ── 6. Ollama endpoint health as recorded ──

\echo ''
\echo '=== 6. ai_endpoint_health ==='

SELECT name, healthy, last_checked, last_healthy, error, check_mode
FROM ai_endpoint_health
ORDER BY name;


-- ── 7. Existing agent_definitions that might already use rag_lookup / rag_index ──

\echo ''
\echo '=== 7. Agents referencing rag_lookup or rag_index in their workflow ==='

SELECT type, is_active, is_snapshot,
       default_config::text LIKE '%rag_lookup%' as uses_rag_lookup,
       default_config::text LIKE '%rag_index%' as uses_rag_index
FROM agent_definitions
WHERE (default_config::text LIKE '%rag_lookup%'
       OR default_config::text LIKE '%rag_index%')
  AND deleted_at IS NULL
ORDER BY type;


-- ── 8. Any test agents we could borrow (type name hints) ──

\echo ''
\echo '=== 8. Agent types that look like test / sandbox agents ==='

SELECT type, is_active, is_snapshot, display_name, created_at
FROM agent_definitions
WHERE deleted_at IS NULL
  AND (type ILIKE '%test%' OR type ILIKE '%sample%' OR type ILIKE '%sandbox%'
       OR type ILIKE '%rag%' OR type ILIKE '%knowledge%')
ORDER BY type;


-- ── 9. Confirm rag_index / rag_lookup are reachable: quick check via sample work item ──

\echo ''
\echo '=== 9. Recent work items (to see what flows through the system) ==='

SELECT work_type, count(*) as cnt
FROM work_items
WHERE created_at > NOW() - INTERVAL '7 days'
GROUP BY work_type
ORDER BY cnt DESC
LIMIT 20;
