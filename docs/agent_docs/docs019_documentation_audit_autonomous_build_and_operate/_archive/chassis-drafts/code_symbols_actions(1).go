// FILE: platform/orchestration/actions/code_symbols_actions.go
//
// DRAFT for the agent-chassis repo. Does not compile in the contextkit
// container — built in your env.
//
// Code-symbol retrieval + indexing actions, the code analogue of rag_actions.go.
// They REUSE the rag embedding helpers (same package): createRAGEmbeddingClient,
// applyNomicPrefix, pgvectorString, getEmbeddingModel, resolveRAGConfigField,
// nullIfEmpty. They DIFFER from rag in three deliberate ways:
//   - source: symbols from the analyser (analysis.Output), NOT char-window chunks
//     (chunkContent fragments code, so it is not used here);
//   - target: the code_symbols table (own columns/identity), NOT knowledge_base;
//   - conflict: identity upsert on (repo,path,symbol) DO UPDATE, NOT content
//     DO NOTHING — code is versioned, symbols change in place and are pruned.
//
// index_code_symbols:  analysis.Output → upsert rows, embed changed symbols
//                      (non-fatal), prune symbols absent from this commit.
// lookup_code_symbols: embed query → vector search code_symbols → trigram
//                      fallback → top-k + combined code_context.

package analyser

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/gqls/agentchassis/internal/analysis"
	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"go.uber.org/zap"
)

// ============================================================================
// lookup_code_symbols
// ============================================================================

func LookupCodeSymbolsAction(ctx context.Context, params ActionParams) (interface{}, error) {
	config := params.StepConfig.Config
	logger := params.Logger

	if params.ExecutionContext != nil && params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}

	queryText := resolveRAGConfigField(config, "query_field", "query", params.CollectedData)
	if queryText == "" {
		logger.Warn("lookup_code_symbols: empty query, returning no results")
		return map[string]interface{}{
			"code_results":  []interface{}{},
			"code_context":  "",
			"result_count":  0,
			"search_method": "none",
		}, nil
	}

	repo := resolveRAGConfigField(config, "repo_field", "repo", params.CollectedData)
	topK := datahelpers.GetIntField(config, "top_k", 12)

	searchMethod := "vector"
	var results []map[string]interface{}
	var err error

	embClient, embErr := createRAGEmbeddingClient(ctx, config)
	if embErr == nil {
		promptText, prefixApplied := applyNomicPrefix(config, queryText, "search_query")
		queryEmbedding, embGenErr := embClient.GenerateEmbedding(ctx, promptText)
		if embGenErr == nil {
			logger.Info("lookup_code_symbols: generated query embedding",
				zap.String("query_preview", truncateForLog(queryText, 100)),
				zap.Int("embedding_dims", len(queryEmbedding)),
				zap.Bool("prefix_applied", prefixApplied),
				zap.String("repo", repo))
			results, err = vectorSearchCodeSymbols(ctx, params.DB, repo, queryEmbedding, topK)
		} else {
			logger.Warn("lookup_code_symbols: embedding generation failed, falling back to trigram",
				zap.Error(embGenErr))
			searchMethod = "trigram"
		}
	} else {
		logger.Warn("lookup_code_symbols: embedding client creation failed, falling back to trigram",
			zap.Error(embErr))
		searchMethod = "trigram"
	}

	if searchMethod == "trigram" {
		results, err = trigramSearchCodeSymbols(ctx, params.DB, repo, queryText, topK)
	}
	if err != nil {
		return nil, fmt.Errorf("lookup_code_symbols: search failed: %w", err)
	}

	// Combined context: one line per hit (path:symbol + signature). Bodies are
	// NOT stored here — the assembler reads them from the repo at commit_sha.
	var parts []string
	for i, r := range results {
		path, _ := r["path"].(string)
		symbol, _ := r["symbol"].(string)
		sig, _ := r["signature"].(string)
		header := fmt.Sprintf("[%d] %s :: %s", i+1, path, symbol)
		if sig != "" {
			parts = append(parts, header+"\n"+sig)
		} else {
			parts = append(parts, header)
		}
	}

	return map[string]interface{}{
		"code_results":  results,
		"code_context":  strings.Join(parts, "\n\n"),
		"result_count":  len(results),
		"search_method": searchMethod,
	}, nil
}

// ============================================================================
// index_code_symbols
// ============================================================================

func IndexCodeSymbolsAction(ctx context.Context, params ActionParams) (interface{}, error) {
	config := params.StepConfig.Config
	logger := params.Logger

	if params.ExecutionContext != nil && params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}

	// Resolve repo + commit + the analyser Output from collected_data. Field
	// names are prefixed (repo, not site_id/domain) to dodge the
	// ExtractActionInputs nested-source collisions (doc 001).
	repo := resolveRAGConfigField(config, "repo_field", "repo", params.CollectedData)
	if repo == "" {
		repo = datahelpers.ExtractNestedFieldString(params.CollectedData, "input_data.repo")
	}
	if repo == "" {
		return nil, fmt.Errorf("index_code_symbols: repo not found (set config.repo, config.repo_field, or input_data.repo)")
	}
	commitField := datahelpers.GetStringField(config, "commit_field", "repo_analysis.commit_sha")
	commitSHA := datahelpers.ExtractNestedFieldString(params.CollectedData, commitField)

	analysisField := datahelpers.GetStringField(config, "analysis_field", "repo_analysis.output")
	rawOutput := datahelpers.ExtractNestedField(params.CollectedData, analysisField)
	if rawOutput == nil {
		return nil, fmt.Errorf("index_code_symbols: no analysis output at %q", analysisField)
	}
	// Re-marshal the (map[string]interface{}) field back to JSON and decode into
	// the typed Output — the analyser ran behind the adapter, so collected_data
	// holds the JSON-decoded form, not the Go struct.
	outBytes, err := json.Marshal(rawOutput)
	if err != nil {
		return nil, fmt.Errorf("index_code_symbols: re-marshal analysis output: %w", err)
	}
	var out analysis.Output
	if err := json.Unmarshal(outBytes, &out); err != nil {
		return nil, fmt.Errorf("index_code_symbols: decode analysis output: %w", err)
	}

	rows := flattenSymbols(out)
	if len(rows) == 0 {
		logger.Warn("index_code_symbols: analyser produced no symbols", zap.String("repo", repo))
		return map[string]interface{}{"indexed": true, "repo": repo, "symbols": 0, "upserted": 0, "pruned": 0}, nil
	}

	// Existing content hashes for this repo, so unchanged symbols are not
	// re-embedded (embedding is the expensive call).
	existing := loadExistingHashes(ctx, params.DB, repo, logger)

	embClient, embErr := createRAGEmbeddingClient(ctx, config)
	if embErr != nil {
		logger.Warn("index_code_symbols: embedding client unavailable, storing without embeddings", zap.Error(embErr))
	}
	embModel := getEmbeddingModel(config)

	upserted, embedded, embeddingsFailed := 0, 0, 0
	for _, r := range rows {
		// Embed only when new or changed; COALESCE in the upsert keeps the
		// existing embedding otherwise.
		var embArg interface{} // nil → NULL::vector
		if embClient != nil && existing[r.path+"\x00"+r.symbol] != r.hash {
			embedInput, _ := applyNomicPrefix(config, r.content, "search_document")
			emb, gErr := embClient.GenerateEmbedding(ctx, embedInput)
			if gErr != nil {
				logger.Warn("index_code_symbols: embedding failed for symbol, storing without",
					zap.String("symbol", r.symbol), zap.Error(gErr))
				embeddingsFailed++
			} else {
				embArg = pgvectorString(emb)
				embedded++
			}
		}

		_, execErr := params.DB.ExecContext(ctx, `
			INSERT INTO code_symbols (
				repo, commit_sha, path, symbol, kind, signature, doc,
				line_start, line_end, content, content_hash, embedding, embedding_model
			) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12::vector,$13)
			ON CONFLICT (repo, path, symbol) DO UPDATE SET
				commit_sha      = EXCLUDED.commit_sha,
				kind            = EXCLUDED.kind,
				signature       = EXCLUDED.signature,
				doc             = EXCLUDED.doc,
				line_start      = EXCLUDED.line_start,
				line_end        = EXCLUDED.line_end,
				content         = EXCLUDED.content,
				content_hash    = EXCLUDED.content_hash,
				embedding       = COALESCE(EXCLUDED.embedding, code_symbols.embedding),
				embedding_model = EXCLUDED.embedding_model,
				updated_at      = now()`,
			repo, nullIfEmpty(commitSHA), r.path, r.symbol, r.kind,
			nullIfEmpty(r.signature), nullIfEmpty(r.doc),
			r.lineStart, r.lineEnd, r.content, r.hash, embArg, embModel)
		if execErr != nil {
			logger.Warn("index_code_symbols: upsert failed", zap.String("symbol", r.symbol), zap.Error(execErr))
			continue
		}
		upserted++
	}

	// Prune symbols absent from this commit. Only safe when a commit_sha is
	// known; a working-tree index (no commit) retains stale rows and logs.
	pruned := 0
	if commitSHA != "" {
		res, pErr := params.DB.ExecContext(ctx,
			`DELETE FROM code_symbols WHERE repo = $1 AND commit_sha IS DISTINCT FROM $2`, repo, commitSHA)
		if pErr != nil {
			logger.Warn("index_code_symbols: prune failed", zap.Error(pErr))
		} else {
			n, _ := res.RowsAffected()
			pruned = int(n)
		}
	} else {
		logger.Warn("index_code_symbols: no commit_sha — skipping prune (stale symbols retained)",
			zap.String("repo", repo))
	}

	logger.Info("index_code_symbols: complete",
		zap.String("repo", repo), zap.String("commit_sha", commitSHA),
		zap.Int("symbols", len(rows)), zap.Int("upserted", upserted),
		zap.Int("embedded", embedded), zap.Int("embeddings_failed", embeddingsFailed),
		zap.Int("pruned", pruned))

	return map[string]interface{}{
		"indexed":           true,
		"repo":              repo,
		"commit_sha":        commitSHA,
		"symbols":           len(rows),
		"upserted":          upserted,
		"embedded":          embedded,
		"embeddings_failed": embeddingsFailed,
		"pruned":            pruned,
	}, nil
}

// ============================================================================
// Helpers (code_symbols-specific; the KB equivalents hardcode knowledge_base)
// ============================================================================

type symbolRow struct {
	path, symbol, kind, signature, doc, content, hash string
	lineStart, lineEnd                                int
}

// flattenSymbols turns the analyser Output into one row per function/method and
// per type. Method symbol names carry the receiver, e.g. "(*OllamaClient).GenerateEmbedding".
func flattenSymbols(out analysis.Output) []symbolRow {
	var rows []symbolRow
	for _, f := range out.Files {
		for _, fn := range f.Functions {
			kind := "func"
			name := fn.Name
			if fn.Receiver != nil {
				kind = "method"
				name = "(" + fn.Receiver.Type + ")." + fn.Name
			}
			content := composeSymbolContent(kind, name, fn.Signature, fn.Doc, f.Path)
			rows = append(rows, symbolRow{
				path: f.Path, symbol: name, kind: kind,
				signature: fn.Signature, doc: fn.Doc, content: content,
				hash: sha256hex(content), lineStart: fn.StartLine, lineEnd: fn.EndLine,
			})
		}
		for _, td := range f.Types {
			kind := td.Kind // struct | interface | alias — all in the CHECK set
			if kind == "" {
				kind = "type"
			}
			content := composeSymbolContent(kind, td.Name, "", td.Doc, f.Path)
			rows = append(rows, symbolRow{
				path: f.Path, symbol: td.Name, kind: kind,
				signature: "", doc: td.Doc, content: content,
				hash: sha256hex(content), lineStart: td.StartLine, lineEnd: td.EndLine,
			})
		}
	}
	return rows
}

// composeSymbolContent builds the searchable text (embedded AND trigram-matched):
// kind + symbol + signature + doc + path.
func composeSymbolContent(kind, symbol, signature, doc, path string) string {
	var b strings.Builder
	b.WriteString(kind)
	b.WriteString(" ")
	b.WriteString(symbol)
	b.WriteString("\n")
	if signature != "" {
		b.WriteString(signature)
		b.WriteString("\n")
	}
	if doc != "" {
		b.WriteString(doc)
		b.WriteString("\n")
	}
	b.WriteString(path)
	return b.String()
}

func sha256hex(s string) string {
	return fmt.Sprintf("%x", sha256.Sum256([]byte(s)))
}

func loadExistingHashes(ctx context.Context, db *sql.DB, repo string, logger *zap.Logger) map[string]string {
	existing := map[string]string{}
	rows, err := db.QueryContext(ctx, `SELECT path, symbol, content_hash FROM code_symbols WHERE repo = $1`, repo)
	if err != nil {
		logger.Warn("index_code_symbols: could not load existing hashes (will re-embed all)", zap.Error(err))
		return existing
	}
	defer rows.Close()
	for rows.Next() {
		var p, s, h string
		if rows.Scan(&p, &s, &h) == nil {
			existing[p+"\x00"+s] = h
		}
	}
	return existing
}

func vectorSearchCodeSymbols(ctx context.Context, db *sql.DB, repo string, embedding []float32, topK int) ([]map[string]interface{}, error) {
	embStr := pgvectorString(embedding)
	query := `
		SELECT id, repo, path, symbol, kind, signature, doc, line_start, line_end, commit_sha,
		       1 - (embedding <=> $1::vector) AS similarity
		FROM code_symbols
		WHERE embedding IS NOT NULL`
	args := []interface{}{embStr}
	if repo != "" {
		query += ` AND repo = $2`
		args = append(args, repo)
	}
	query += fmt.Sprintf(` ORDER BY embedding <=> $1::vector LIMIT %d`, topK)

	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("code vector search failed: %w", err)
	}
	defer rows.Close()
	return scanCodeSymbolRows(rows)
}

func trigramSearchCodeSymbols(ctx context.Context, db *sql.DB, repo, queryText string, topK int) ([]map[string]interface{}, error) {
	query := `
		SELECT id, repo, path, symbol, kind, signature, doc, line_start, line_end, commit_sha,
		       similarity(content, $1) AS similarity
		FROM code_symbols
		WHERE content % $1`
	args := []interface{}{queryText}
	if repo != "" {
		query += ` AND repo = $2`
		args = append(args, repo)
	}
	query += fmt.Sprintf(` ORDER BY similarity(content, $1) DESC LIMIT %d`, topK)

	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("code trigram search failed: %w", err)
	}
	defer rows.Close()
	return scanCodeSymbolRows(rows)
}

func scanCodeSymbolRows(rows *sql.Rows) ([]map[string]interface{}, error) {
	var results []map[string]interface{}
	for rows.Next() {
		var id, repo, path, symbol, kind string
		var signature, doc, commitSHA sql.NullString
		var lineStart, lineEnd sql.NullInt64
		var similarity float64

		if err := rows.Scan(&id, &repo, &path, &symbol, &kind, &signature, &doc,
			&lineStart, &lineEnd, &commitSHA, &similarity); err != nil {
			continue
		}
		r := map[string]interface{}{
			"id":         id,
			"repo":       repo,
			"path":       path,
			"symbol":     symbol,
			"kind":       kind,
			"similarity": similarity,
		}
		if signature.Valid {
			r["signature"] = signature.String
		}
		if doc.Valid {
			r["doc"] = doc.String
		}
		if commitSHA.Valid {
			r["commit_sha"] = commitSHA.String
		}
		if lineStart.Valid {
			r["line_start"] = lineStart.Int64
		}
		if lineEnd.Valid {
			r["line_end"] = lineEnd.Int64
		}
		results = append(results, r)
	}
	return results, nil
}
