-- Migration NNN — code_symbols: the context tool's per-repo code index.
-- Renumber NNN to the next sequential migration number before applying.
--
-- WHAT
--   A sibling to knowledge_base. It reuses knowledge_base's proven SHAPE
--   (pgvector(768) + a trigram index on the searchable text + idempotent dedup)
--   and the same embedder (aiservice.AIService.GenerateEmbedding, nomic-embed-text
--   via the ollama-adapter) — but it is keyed and shaped for CODE, not for the
--   marketing knowledge base.
--
-- WHY A SIBLING TABLE, NOT A knowledge_base COLLECTION
--   knowledge_base is marketing-knowledge-shaped (industry, domain, title,
--   source_url, quality_score, usage_count) on a web-parse / quality lifecycle.
--   Code symbols are SHA-keyed, commit-indexed, and identified by (repo, path,
--   symbol). Riding in knowledge_base under collection='chassis_code' would
--   overload `collection` and leave half its columns permanently null. We reuse
--   the shape and the embedder, not the table.
--
-- DELIBERATE DEPARTURES FROM THE USUAL CONVENTIONS (flagged on purpose)
--   * No `version` / `previous_version_id`. Those are for system-of-record
--     entities. This is a rebuildable cache; it is versioned by `commit_sha`,
--     not per row.
--   * No `deleted_at` soft-delete. A row for a deleted function has no value, so
--     stale symbols are PRUNED (hard delete) on re-index. Restoring a deleted
--     symbol means re-indexing the commit that has it.
--   These are intentional; they are not an oversight of the soft-delete/versioning
--   conventions used elsewhere.
--
-- BEFORE APPLYING (reuse-before-recreate, and the §6.1 backup-table footgun)
--   1. Confirm nothing already fills this role:   \dt *code*    \dt *symbol*
--      This file uses plain CREATE TABLE (not CREATE TABLE IF NOT EXISTS) on
--      purpose: if a table of this name already exists with a different shape,
--      it should ERROR here, not silently no-op (the migration-110 trap in the
--      debugging guide §6.1).
--   2. HNSW needs pgvector >= 0.5.0. The knowledge_base index is IVFFlat, which
--      does NOT prove HNSW is available — confirm with  \dx  (look at the vector
--      extension version). If HNSW is unavailable, use the IVFFlat fallback at the
--      foot of this file (the same access method knowledge_base already uses).

-- Already present (knowledge_base uses both); declared here for self-documentation.
-- IF NOT EXISTS is safe for extensions (unlike for the table above).
CREATE EXTENSION IF NOT EXISTS vector;
CREATE EXTENSION IF NOT EXISTS pg_trgm;

CREATE TABLE code_symbols (
    id              uuid         PRIMARY KEY DEFAULT gen_random_uuid(),
    repo            varchar(255) NOT NULL,   -- analogue of knowledge_base.collection; the partition key.
                                             -- Multi-tenant scoping (plan B5) is deferred: prefix the repo
                                             -- value (e.g. 'tenant/agent-chassis') or add a tenant_id later.
    commit_sha      varchar(64),             -- provenance SHA, stamped by the indexing workflow. NULLable so a
                                             -- local/harness run can index a working tree without a commit.
    path            text         NOT NULL,   -- file path relative to repo root, e.g. 'cmd/dbcontext/main.go'.
    symbol          varchar(255) NOT NULL,   -- symbol name from the analyser, e.g. 'SpawnAgentAction' or
                                             -- '(*OllamaClient).GenerateEmbedding'.
    kind            text         NOT NULL
                    CHECK (kind IN ('func','method','struct','interface','alias','type','var','const')),
                                             -- the analyser's Go kinds. Keep this list covering every `Kind`
                                             -- the analyser emits; widen it when a new-language analyser is added.
    signature       text,                    -- one-line signature from the analyser.
    doc             text,                    -- doc comment, if any.
    line_start      integer,                 -- the symbol's line range. The assembler reads the BODY from the
    line_end        integer,                 -- repo at commit_sha; full bodies are NOT stored here (keeps rows small).
    content         text         NOT NULL,   -- the searchable text that is embedded AND trigram-matched:
                                             -- name + kind + signature + doc + path.
    content_hash    varchar(64)  NOT NULL,   -- hash of `content`; lets re-index SKIP re-embedding unchanged symbols.
    embedding       vector(768),             -- nomic-embed-text dimension. NULLable so lexical (trigram) retrieval
                                             -- still works when an embedding is deferred or fails.
    embedding_model varchar(100) DEFAULT 'nomic-embed-text',
    metadata        jsonb        NOT NULL DEFAULT '{}',  -- extensibility: package, exported flag, calls/called-by, etc.
    created_at      timestamptz  NOT NULL DEFAULT now(),
    updated_at      timestamptz  NOT NULL DEFAULT now(),

    -- Identity: one current row per symbol. The indexing upsert targets this key.
    CONSTRAINT uq_code_symbols_identity UNIQUE (repo, path, symbol)
);

-- Semantic kNN (HNSW, cosine) — PRIMARY. Matches knowledge_base's cosine ops.
-- HNSW is preferred over IVFFlat here because this index is updated incrementally
-- per commit; IVFFlat trains centroids on the data at build time and degrades (needs
-- REINDEX) as rows churn, whereas HNSW absorbs inserts without retraining.
CREATE INDEX idx_code_symbols_embedding_hnsw
    ON code_symbols USING hnsw (embedding vector_cosine_ops)
    WITH (m = 16, ef_construction = 64);

-- Lexical (trigram) — mirrors knowledge_base.idx_kb_content_trgm.
CREATE INDEX idx_code_symbols_content_trgm
    ON code_symbols USING gin (content gin_trgm_ops);

-- Filtering by repo + kind (e.g. "interfaces in repo X").
CREATE INDEX idx_code_symbols_repo_kind
    ON code_symbols (repo, kind);

-- Change-detection lookups during incremental re-index.
CREATE INDEX idx_code_symbols_repo_hash
    ON code_symbols (repo, content_hash);

COMMENT ON TABLE code_symbols IS
    'Per-repo code-symbol index for the context tool. Sibling to knowledge_base; reuses pgvector(768)+trigram, keyed for code (repo,path,symbol). SHA-versioned, pruned (no soft-delete).';

-- ============================================================================
-- Usage patterns (parameterised; these run in the indexing/build actions, not in
-- this migration). Included as the contract this table is built to serve.
-- ============================================================================
--
-- INDEXING upsert — one row per symbol. The workflow SELECTs the existing
-- content_hash first and calls GenerateEmbedding only for new/changed symbols; the
-- DO UPDATE ... WHERE then avoids a needless write when nothing changed:
--
--   INSERT INTO code_symbols
--     (repo, commit_sha, path, symbol, kind, signature, doc, line_start, line_end,
--      content, content_hash, embedding, embedding_model, metadata)
--   VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14)
--   ON CONFLICT (repo, path, symbol) DO UPDATE SET
--     commit_sha      = EXCLUDED.commit_sha,
--     kind            = EXCLUDED.kind,
--     signature       = EXCLUDED.signature,
--     doc             = EXCLUDED.doc,
--     line_start      = EXCLUDED.line_start,
--     line_end        = EXCLUDED.line_end,
--     content         = EXCLUDED.content,
--     content_hash    = EXCLUDED.content_hash,
--     embedding       = EXCLUDED.embedding,
--     embedding_model = EXCLUDED.embedding_model,
--     metadata        = EXCLUDED.metadata,
--     updated_at      = now()
--   WHERE code_symbols.content_hash IS DISTINCT FROM EXCLUDED.content_hash;
--
-- PRUNE symbols that no longer exist in the code (hard delete — the rebuildable-cache
-- reason there is no deleted_at). $2 is the array of current 'path|symbol' keys:
--
--   DELETE FROM code_symbols
--   WHERE repo = $1
--     AND (path || '|' || symbol) <> ALL ($2::text[]);
--
-- SEMANTIC retrieval (HNSW cosine; $1 = query embedding, $2 = repo, $3 = limit):
--
--   SELECT path, symbol, kind, signature, 1 - (embedding <=> $1) AS score
--   FROM code_symbols
--   WHERE repo = $2 AND embedding IS NOT NULL
--   ORDER BY embedding <=> $1
--   LIMIT $3;
--   -- tune recall per session with:  SET hnsw.ef_search = 40;  (raise for better recall)
--
-- LEXICAL retrieval (trigram; $1 = query text, $2 = repo, $3 = limit):
--
--   SELECT path, symbol, kind, signature, similarity(content, $1) AS sim
--   FROM code_symbols
--   WHERE repo = $2 AND content % $1
--   ORDER BY sim DESC
--   LIMIT $3;
--
-- HYBRID in SQL — reciprocal-rank fusion of the two lists; this is what replaces the
-- in-Go `fuse`. $1 = query embedding, $2 = query text, $3 = repo, $4 = per-list depth,
-- $5 = final limit. 60 is the usual RRF constant:
--
--   WITH sem AS (
--     SELECT id, row_number() OVER (ORDER BY embedding <=> $1) AS rnk
--     FROM code_symbols WHERE repo = $3 AND embedding IS NOT NULL
--     ORDER BY embedding <=> $1 LIMIT $4
--   ),
--   lex AS (
--     SELECT id, row_number() OVER (ORDER BY similarity(content, $2) DESC) AS rnk
--     FROM code_symbols WHERE repo = $3 AND content % $2
--     ORDER BY similarity(content, $2) DESC LIMIT $4
--   ),
--   fused AS (
--     SELECT id, SUM(1.0 / (60 + rnk)) AS score
--     FROM (SELECT id, rnk FROM sem UNION ALL SELECT id, rnk FROM lex) u
--     GROUP BY id
--   )
--   SELECT cs.path, cs.symbol, cs.kind, cs.signature, f.score
--   FROM fused f JOIN code_symbols cs ON cs.id = f.id
--   ORDER BY f.score DESC
--   LIMIT $5;
--
-- ============================================================================
-- FALLBACK if HNSW is unavailable (pgvector < 0.5.0): IVFFlat, like knowledge_base.
-- IVFFlat must be created AFTER rows are loaded (it trains on existing data), and
-- `lists` should scale with row count (~ rows / 1000, a small minimum). Re-run
-- REINDEX as the corpus grows — exactly the cost HNSW avoids.
--
--   -- DROP INDEX idx_code_symbols_embedding_hnsw;
--   -- CREATE INDEX idx_code_symbols_embedding_ivf
--   --   ON code_symbols USING ivfflat (embedding vector_cosine_ops) WITH (lists = 100);
--
-- ============================================================================
-- ROLLBACK:
--   DROP TABLE IF EXISTS code_symbols;
