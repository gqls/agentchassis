// FILE: platform/orchestration/actions/rag_actions.go
//
// RAG (Retrieval-Augmented Generation) actions for shared knowledge base.
//
// rag_lookup:  Embed query → vector search → return relevant chunks
//              Falls back to trigram text search if embedding fails
// rag_index:   Chunk content → embed → store in knowledge_base table
//              Embedding failures are non-fatal: chunks stored without embeddings
//
// Both use Ollama for embeddings by default. Configure via step config:
//   "embedding_service": {
//       "provider": "ollama",
//       "model": "nomic-embed-text",
//       "api_url": "http://ollama-adapter.ai-persona-system.svc.cluster.local:11434"
//   }
//
// Reuses existing helpers:
//   - datahelpers.ExtractNestedFieldString for field path resolution
//   - datahelpers.GetStringField, datahelpers.GetIntField for config extraction

package actions

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/gqls/agentchassis/platform/aiservice"
	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"go.uber.org/zap"
)

// ============================================================================
// rag_lookup
// ============================================================================

func RAGLookupAction(ctx context.Context, params ActionParams) (interface{}, error) {
	config := params.StepConfig.Config
	logger := params.Logger

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}

	// Get query text
	queryText := resolveRAGConfigField(config, "query_field", "query", params.CollectedData)
	if queryText == "" {
		logger.Warn("rag_lookup: empty query, returning no results")
		return map[string]interface{}{
			"rag_results":   []interface{}{},
			"rag_context":   "",
			"result_count":  0,
			"search_method": "none",
		}, nil
	}

	collection := datahelpers.GetStringField(config, "collection", "industry_sites")
	industry := ""
	if indField, ok := config["industry_field"].(string); ok {
		industry = datahelpers.ExtractNestedFieldString(params.CollectedData, indField)
	}
	topK := datahelpers.GetIntField(config, "top_k", 5)

	// Try vector search first, fall back to trigram
	searchMethod := "vector"
	var results []map[string]interface{}
	var err error

	embClient, embErr := createRAGEmbeddingClient(ctx, config)
	if embErr == nil {
		promptText, prefixApplied := applyNomicPrefix(config, queryText, "search_query")
		queryEmbedding, embGenErr := embClient.GenerateEmbedding(ctx, promptText)
		if embGenErr == nil {
			logger.Info("rag_lookup: generated query embedding",
				zap.String("query_preview", truncateForLog(queryText, 100)),
				zap.Int("embedding_dims", len(queryEmbedding)),
				zap.Bool("prefix_applied", prefixApplied),
				zap.String("collection", collection))

			results, err = vectorSearchKB(ctx, params.DB, collection, industry, queryEmbedding, topK, logger)
		} else {
			logger.Warn("rag_lookup: embedding generation failed, falling back to trigram",
				zap.Error(embGenErr))
			searchMethod = "trigram"
		}
	} else {
		logger.Warn("rag_lookup: embedding client creation failed, falling back to trigram",
			zap.Error(embErr))
		searchMethod = "trigram"
	}

	// Trigram fallback
	if searchMethod == "trigram" {
		results, err = trigramSearchKB(ctx, params.DB, collection, industry, queryText, topK, logger)
	}

	if err != nil {
		return nil, fmt.Errorf("rag_lookup: search failed: %w", err)
	}

	// Build combined context string for prompt injection
	var contextParts []string
	for i, r := range results {
		header := fmt.Sprintf("[Source %d", i+1)
		if title, ok := r["title"].(string); ok && title != "" {
			header += ": " + title
		}
		header += "]"
		contextParts = append(contextParts, header+"\n"+r["content"].(string))
	}

	// Update usage counts in background
	go updateKBUsageCounts(params.DB, results, logger)

	return map[string]interface{}{
		"rag_results":   results,
		"rag_context":   strings.Join(contextParts, "\n\n---\n\n"),
		"result_count":  len(results),
		"search_method": searchMethod,
	}, nil
}

// ============================================================================
// rag_index
// ============================================================================

func RAGIndexAction(ctx context.Context, params ActionParams) (interface{}, error) {
	config := params.StepConfig.Config
	logger := params.Logger

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}

	contentField := datahelpers.GetStringField(config, "content_field", "content")
	content := datahelpers.ExtractNestedFieldString(params.CollectedData, contentField)
	if content == "" {
		logger.Warn("rag_index: no content found, skipping",
			zap.String("content_field", contentField))
		return map[string]interface{}{
			"indexed": false, "stored": 0, "skipped": 0,
			"reason": "no content at " + contentField,
		}, nil
	}

	collection := datahelpers.GetStringField(config, "collection", "industry_sites")
	industry := extractRAGOptionalField(config, "industry_field", params.CollectedData)
	domain := extractRAGOptionalField(config, "domain_field", params.CollectedData)
	sourceURL := extractRAGOptionalField(config, "source_url_field", params.CollectedData)

	chunkSize := datahelpers.GetIntField(config, "chunk_size", 1000)
	chunkOverlap := datahelpers.GetIntField(config, "chunk_overlap", 200)

	// Chunk the content
	chunks := chunkContent(content, chunkSize, chunkOverlap)
	logger.Info("rag_index: chunked content",
		zap.Int("total_length", len(content)),
		zap.Int("chunks", len(chunks)),
		zap.Int("chunk_size", chunkSize),
		zap.Int("overlap", chunkOverlap))

	// Create embedding client (non-fatal if unavailable)
	embClient, embErr := createRAGEmbeddingClient(ctx, config)
	if embErr != nil {
		logger.Warn("rag_index: embedding client unavailable, storing without embeddings",
			zap.Error(embErr))
	}

	stored := 0
	skipped := 0
	embeddingsFailed := 0

	agentType := ""
	orchID := ""
	if params.ExecutionContext != nil {
		agentType = params.Headers["agent_type"]
		orchID = params.ExecutionContext.OrchestrationID
	}

	// Per-chunk embedding deadline: a stalled ollama-adapter must degrade into
	// the non-fatal "store without embeddings" path below, never freeze the
	// action. Default mirrors the OllamaClient http.Client 120s cap (ollama on
	// CPU is slow, cold model loads slower) — so defaults change nothing; the
	// config key exists to TIGHTEN it per-step where a faster bound is wanted.
	embTimeout := time.Duration(datahelpers.GetIntField(config, "embedding_timeout_seconds", 120)) * time.Second

	prefixAppliedOnce := false
	for i, chunk := range chunks {
		// SHA256 for dedup (computed on ORIGINAL chunk, not prefixed version)
		hash := fmt.Sprintf("%x", sha256.Sum256([]byte(chunk)))

		// Generate embedding (non-fatal). Apply nomic task prefix to the
		// text going to the embedder, but keep chunk unprefixed for storage
		// and hashing.
		var embeddingStr *string
		if embClient != nil {
			embedInput, prefixApplied := applyNomicPrefix(config, chunk, "search_document")
			if prefixApplied && !prefixAppliedOnce {
				logger.Info("rag_index: applying nomic search_document prefix to embeddings",
					zap.String("collection", collection))
				prefixAppliedOnce = true
			}
			embCtx, cancel := context.WithTimeout(ctx, embTimeout)
			embedding, err := embClient.GenerateEmbedding(embCtx, embedInput)
			cancel()
			if err != nil {
				logger.Warn("rag_index: embedding failed for chunk, storing without",
					zap.Int("chunk", i), zap.Error(err))
				embeddingsFailed++
			} else {
				s := pgvectorString(embedding)
				embeddingStr = &s
			}
		}

		// Insert with ON CONFLICT DO NOTHING (dedup on collection + content_hash)
		var result sql.Result
		var insertErr error
		if embeddingStr != nil {
			result, insertErr = params.DB.ExecContext(ctx, `
				INSERT INTO knowledge_base (
					collection, industry, domain, content, content_hash,
					embedding, embedding_model,
					source_type, source_url, source_agent_type, source_orchestration_id
				) VALUES ($1, $2, $3, $4, $5, $6::vector, $7, $8, $9, $10, $11)
				ON CONFLICT (collection, content_hash) WHERE content_hash IS NOT NULL
				DO NOTHING`,
				collection, nullIfEmpty(industry), nullIfEmpty(domain),
				chunk, hash,
				*embeddingStr, "nomic-embed-text",
				"scrape", nullIfEmpty(sourceURL), nullIfEmpty(agentType), nullIfEmpty(orchID))
		} else {
			result, insertErr = params.DB.ExecContext(ctx, `
				INSERT INTO knowledge_base (
					collection, industry, domain, content, content_hash,
					source_type, source_url, source_agent_type, source_orchestration_id
				) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
				ON CONFLICT (collection, content_hash) WHERE content_hash IS NOT NULL
				DO NOTHING`,
				collection, nullIfEmpty(industry), nullIfEmpty(domain),
				chunk, hash,
				"scrape", nullIfEmpty(sourceURL), nullIfEmpty(agentType), nullIfEmpty(orchID))
		}

		if insertErr != nil {
			logger.Warn("rag_index: insert failed for chunk",
				zap.Int("chunk", i), zap.Error(insertErr))
			continue
		}

		rows, _ := result.RowsAffected()
		if rows > 0 {
			stored++
		} else {
			skipped++ // dedup hit
		}
	}

	logger.Info("rag_index: complete",
		zap.Int("stored", stored),
		zap.Int("skipped_dedup", skipped),
		zap.Int("embeddings_failed", embeddingsFailed),
		zap.String("collection", collection))

	return map[string]interface{}{
		"indexed":           true,
		"stored":            stored,
		"skipped":           skipped,
		"embeddings_failed": embeddingsFailed,
		"total_chunks":      len(chunks),
		"collection":        collection,
	}, nil
}

// ============================================================================
// Helpers
// ============================================================================

func vectorSearchKB(ctx context.Context, db *sql.DB, collection, industry string, embedding []float32, topK int, logger *zap.Logger) ([]map[string]interface{}, error) {
	embStr := pgvectorString(embedding)

	query := `
		SELECT id, title, content, source_url, domain,
		       1 - (embedding <=> $1::vector) as similarity
		FROM knowledge_base
		WHERE collection = $2
		  AND embedding IS NOT NULL`
	args := []interface{}{embStr, collection}

	if industry != "" {
		query += ` AND (industry = $3 OR industry IS NULL)`
		args = append(args, industry)
	}
	query += fmt.Sprintf(` ORDER BY embedding <=> $1::vector LIMIT %d`, topK)

	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("vector search query failed: %w", err)
	}
	defer rows.Close()

	return scanKBResults(rows)
}

func trigramSearchKB(ctx context.Context, db *sql.DB, collection, industry, queryText string, topK int, logger *zap.Logger) ([]map[string]interface{}, error) {
	query := `
		SELECT id, title, content, source_url, domain,
		       similarity(content, $1) as similarity
		FROM knowledge_base
		WHERE collection = $2
		  AND content % $1`
	args := []interface{}{queryText, collection}

	if industry != "" {
		query += ` AND (industry = $3 OR industry IS NULL)`
		args = append(args, industry)
	}
	query += fmt.Sprintf(` ORDER BY similarity(content, $1) DESC LIMIT %d`, topK)

	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("trigram search query failed: %w", err)
	}
	defer rows.Close()

	return scanKBResults(rows)
}

func scanKBResults(rows *sql.Rows) ([]map[string]interface{}, error) {
	var results []map[string]interface{}
	for rows.Next() {
		var id, content string
		var title, sourceURL, domain sql.NullString
		var similarity float64

		if err := rows.Scan(&id, &title, &content, &sourceURL, &domain, &similarity); err != nil {
			continue
		}
		result := map[string]interface{}{
			"id":         id,
			"content":    content,
			"similarity": similarity,
		}
		if title.Valid {
			result["title"] = title.String
		}
		if sourceURL.Valid {
			result["source_url"] = sourceURL.String
		}
		if domain.Valid {
			result["domain"] = domain.String
		}
		results = append(results, result)
	}
	return results, nil
}

func createRAGEmbeddingClient(ctx context.Context, config map[string]interface{}) (*aiservice.OllamaClient, error) {
	embConfig := map[string]interface{}{
		"provider": "ollama",
		"model":    "nomic-embed-text",
		"api_url":  "http://ollama-adapter.ai-persona-system.svc.cluster.local:11434",
	}

	if svc, ok := config["embedding_service"].(map[string]interface{}); ok {
		for k, v := range svc {
			embConfig[k] = v
		}
	}

	return aiservice.NewOllamaClient(ctx, embConfig)
}

func resolveRAGConfigField(config map[string]interface{}, fieldKey, directKey string, data map[string]interface{}) string {
	// Check for direct value first
	if val, ok := config[directKey].(string); ok && val != "" {
		return val
	}
	// Then try field path resolution
	if path, ok := config[fieldKey].(string); ok && path != "" {
		return datahelpers.ExtractNestedFieldString(data, path)
	}
	return ""
}

func extractRAGOptionalField(config map[string]interface{}, fieldKey string, data map[string]interface{}) string {
	if path, ok := config[fieldKey].(string); ok && path != "" {
		return datahelpers.ExtractNestedFieldString(data, path)
	}
	return ""
}

func chunkContent(content string, chunkSize, overlap int) []string {
	if len(content) <= chunkSize {
		return []string{content}
	}

	var chunks []string
	start := 0
	for start < len(content) {
		end := start + chunkSize
		if end > len(content) {
			end = len(content)
		}

		// Try to break at a sentence boundary
		if end < len(content) {
			for i := end; i > start+chunkSize/2; i-- {
				if content[i] == '.' || content[i] == '\n' {
					end = i + 1
					break
				}
			}
		}

		chunks = append(chunks, strings.TrimSpace(content[start:end]))

		// The final chunk ends the loop. Without this, start = end - overlap
		// lands before len(content), the loop re-enters, and the same tail
		// chunk is appended forever — 2Gi of duplicates in seconds (both
		// chassis OOMKills of 2026-07-09/10 were this loop on a ~3KB PLAN).
		if end == len(content) {
			break
		}
		next := end - overlap
		if next <= start { // guarantee forward progress for any overlap config
			next = end
		}
		start = next
	}
	return chunks
}

func pgvectorString(embedding []float32) string {
	parts := make([]string, len(embedding))
	for i, v := range embedding {
		parts[i] = fmt.Sprintf("%f", v)
	}
	return "[" + strings.Join(parts, ",") + "]"
}

func updateKBUsageCounts(db *sql.DB, results []map[string]interface{}, logger *zap.Logger) {
	if db == nil {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	for _, r := range results {
		if id, ok := r["id"].(string); ok {
			_, err := db.ExecContext(ctx, "UPDATE knowledge_base SET usage_count = usage_count + 1 WHERE id = $1", id)
			if err != nil {
				logger.Warn("rag: failed to update usage count", zap.String("id", id), zap.Error(err))
			}
		}
	}
}

// applyNomicPrefix prepends a nomic task prefix to the text when the
// configured embedding model is a nomic variant. Nomic embedding models
// (nomic-embed-text, nomic-embed-text-v2-moe, etc.) perform significantly
// better with "search_document: " on indexed content and "search_query: "
// on queries. Any non-nomic model receives the text unchanged.
//
// Returns the (possibly prefixed) text and a bool indicating whether a
// prefix was applied — used for logging so regressions are visible in
// llm_call_log and action logs.
func applyNomicPrefix(config map[string]interface{}, text, task string) (string, bool) {
	model := getEmbeddingModel(config)
	if !strings.HasPrefix(model, "nomic-embed-") {
		return text, false
	}
	// Don't double-prefix if caller already prepended one.
	if strings.HasPrefix(text, "search_document: ") || strings.HasPrefix(text, "search_query: ") {
		return text, false
	}
	return task + ": " + text, true
}

// getEmbeddingModel extracts the model name from the embedding_service
// config map, defaulting to nomic-embed-text to match the default in
// createRAGEmbeddingClient.
func getEmbeddingModel(config map[string]interface{}) string {
	if svc, ok := config["embedding_service"].(map[string]interface{}); ok {
		if m, ok := svc["model"].(string); ok && m != "" {
			return m
		}
	}
	return "nomic-embed-text"
}
