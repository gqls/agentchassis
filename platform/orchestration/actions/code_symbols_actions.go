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
//
// BODIES (2026-07-27, D11 layer 1, council 18fe4035): each row also carries the
// symbol's SOURCE TEXT in code_symbols.body, sliced from the local checkout by
// the line span the analyser already recorded. Until this, `content` was the only
// searchable text and it holds DECLARATIONS ONLY (composeSymbolContent: kind +
// symbol + signature + doc + path), so diagnose_code_lookup's `content` kind —
// which documents itself as matching "symbol source bodies" — returned zero rows
// for every string literal, route, registry key or table name inside a function.
// A zero read as absence is how a reviewer concludes "no such code exists" about
// code that does (bugs_open/108 defect B).
//
// body is a SEPARATE COLUMN and is deliberately absent from content_hash: that
// hash is the re-embed trigger (loadExistingHashes), so folding bodies into
// content would silently re-embed and re-skew all 4,535 vectors as a side effect
// of a search fix. 243_code_symbols_body_column_VERIFY.sql asserts that.

package actions

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gqls/agentchassis/internal/analysis"
	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"go.uber.org/zap"
)

// ============================================================================
// shared repo-label resolution
// ============================================================================

// resolveCodeRepoLabel resolves the owner/repo label that BOTH index_code_symbols
// and lookup_code_symbols key on, so the writer and the reader can never diverge
// (a divergence is what caused lookup to query a bare "agentchassis" against rows
// stored under "gqls/agentchassis" -> 0 hits). Resolution order:
//  1. config.repo / config.repo_field — explicit override (non-git corpora,
//     e.g. "domain:kruste.com");
//  2. COMPOSE owner/repo from the analyser reply — the default, which matches
//     what was actually fetched and stored
//     (repo_analysis.owner + "/" + repo_analysis.repo);
//  3. input_data.repo — last-resort fallback.
//
// config keys owner_field / repo_name_field default to the request step's
// output_field "repo_analysis".
func resolveCodeRepoLabel(config map[string]interface{}, collected map[string]interface{}) string {
	repo := resolveRAGConfigField(config, "repo_field", "repo", collected)
	if repo == "" {
		ownerPath := datahelpers.GetStringField(config, "owner_field", "repo_analysis.owner")
		namePath := datahelpers.GetStringField(config, "repo_name_field", "repo_analysis.repo")
		owner := datahelpers.ExtractNestedFieldString(collected, ownerPath)
		name := datahelpers.ExtractNestedFieldString(collected, namePath)
		if owner != "" && name != "" {
			repo = owner + "/" + name
		}
	}
	if repo == "" {
		repo = datahelpers.ExtractNestedFieldString(collected, "input_data.repo")
	}
	return repo
}

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

	repo := resolveCodeRepoLabel(config, params.CollectedData)
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

	// Combined context: one line per hit (path:symbol + signature). Bodies ARE
	// stored now (code_symbols.body) but are deliberately NOT rendered here: this
	// is retrieval SEEDING for the bundle assembler, which reads the body from the
	// checkout at commit_sha, and a top-k of full function bodies would blow the
	// caller's context. The body column serves diagnose_code_lookup's `content`
	// kind, which caps its excerpt.
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

	// Resolve the owner/repo label via the shared resolver (same logic the lookup
	// now uses, so writer and reader can't diverge). LABEL CONVENTION (2026-06-11):
	// code_symbols.repo is the "owner/repo" form (e.g. "gqls/agentchassis"),
	// composed from the analyser reply's owner + repo by default. See
	// resolveCodeRepoLabel for the full resolution order. Field names are prefixed
	// (repo, not site_id/domain) to dodge ExtractActionInputs collisions (doc 001).
	repo := resolveCodeRepoLabel(config, params.CollectedData)
	if repo == "" {
		return nil, fmt.Errorf("index_code_symbols: repo label not found (set config.repo, config.repo_field, or let it compose from %q owner+repo)", "repo_analysis")
	}
	commitField := datahelpers.GetStringField(config, "commit_field", "repo_analysis.commit_sha")
	commitSHA := datahelpers.ExtractNestedFieldString(params.CollectedData, commitField)

	// The ref the analyse step fetched, and the fetched commit's own committer
	// date (bugs_open/108 defect A). Both were always known at this point —
	// analyse_repo_local returns them beside commit_sha — and both were
	// discarded here, which is why the freshness banner had nothing to key on
	// but the row clock. Absent-or-unparseable persists as NULL, the honest
	// "unrecorded" state the banner renders as UNKNOWN, never as fresh.
	refField := datahelpers.GetStringField(config, "ref_field", "repo_analysis.ref")
	indexedRef := datahelpers.ExtractNestedFieldString(params.CollectedData, refField)
	commitTimeField := datahelpers.GetStringField(config, "commit_time_field", "repo_analysis.commit_time")
	var commitTime interface{} // nil → NULL::timestamptz
	if s := datahelpers.ExtractNestedFieldString(params.CollectedData, commitTimeField); s != "" {
		if t, perr := time.Parse(time.RFC3339, s); perr == nil {
			commitTime = t.UTC()
		} else {
			logger.Warn("index_code_symbols: unparseable commit_time — storing NULL (freshness banner will read UNKNOWN)",
				zap.String("commit_time", s), zap.Error(perr))
		}
	}

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

	rows := flattenSymbols(out, logger)
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

	upserted, embedded, embeddingsFailed, withBody := 0, 0, 0, 0
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

		// body is assigned PLAINLY, never COALESCEd onto the existing value the
		// way embedding is. Preserving an old body on a failed slice looks safer
		// and is not: content_hash covers the DECLARATION text only (kind +
		// symbol + signature + doc + path), so a function whose body changed
		// while its signature did not has an UNCHANGED hash — there is no cheap
		// test for "this body is still current". line_start/line_end above are
		// overwritten from EXCLUDED regardless, so a preserved body would end up
		// contradicting the very span it claims to be. NULL is the honest state.
		// ref and commit_time are assigned plainly, like body and for the same
		// reason: a re-index at a new ref must overwrite, and an unrecorded
		// commit_time must persist as NULL rather than be COALESCEd onto a stale
		// value it would then contradict.
		_, execErr := params.DB.ExecContext(ctx, `
			INSERT INTO code_symbols (
				repo, commit_sha, ref, commit_time, path, symbol, kind, signature, doc,
				line_start, line_end, content, body, content_hash, embedding, embedding_model
			) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15::vector,$16)
			ON CONFLICT (repo, path, symbol) DO UPDATE SET
				commit_sha      = EXCLUDED.commit_sha,
				ref             = EXCLUDED.ref,
				commit_time     = EXCLUDED.commit_time,
				kind            = EXCLUDED.kind,
				signature       = EXCLUDED.signature,
				doc             = EXCLUDED.doc,
				line_start      = EXCLUDED.line_start,
				line_end        = EXCLUDED.line_end,
				content         = EXCLUDED.content,
				body            = EXCLUDED.body,
				content_hash    = EXCLUDED.content_hash,
				embedding       = COALESCE(EXCLUDED.embedding, code_symbols.embedding),
				embedding_model = EXCLUDED.embedding_model,
				updated_at      = now()`,
			repo, nullIfEmpty(commitSHA), nullIfEmpty(indexedRef), commitTime, r.path, r.symbol, r.kind,
			nullIfEmpty(r.signature), nullIfEmpty(r.doc),
			r.lineStart, r.lineEnd, r.content, r.body, r.hash, embArg, embModel)
		if execErr != nil {
			logger.Warn("index_code_symbols: upsert failed", zap.String("symbol", r.symbol), zap.Error(execErr))
			continue
		}
		upserted++
		if r.body != nil {
			withBody++
		}
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
		zap.Int("with_body", withBody), zap.Int("pruned", pruned))

	return map[string]interface{}{
		"indexed":           true,
		"repo":              repo,
		"commit_sha":        commitSHA,
		"symbols":           len(rows),
		"upserted":          upserted,
		"embedded":          embedded,
		"embeddings_failed": embeddingsFailed,
		// with_body is reported next to upserted so a run that indexed everything
		// and sliced nothing is legible from the orchestration record alone,
		// without anyone thinking to query the column.
		"with_body": withBody,
		"pruned":    pruned,
	}, nil
}

// ============================================================================
// Helpers (code_symbols-specific; the KB equivalents hardcode knowledge_base)
// ============================================================================

type symbolRow struct {
	path, symbol, kind, signature, doc, content, hash string
	lineStart, lineEnd                                int
	// body is the symbol's source text, or nil when it could not be sliced.
	// A POINTER, not a string, so "could not read it" persists as NULL and can
	// never be confused with a genuinely empty body — the empty-vs-absent
	// confusion this whole change exists to remove, reintroduced one layer down.
	body *string
}

// flattenSymbols turns the analyser Output into one row per function/method and
// per type. Method symbol names carry the receiver, e.g. "(*OllamaClient).GenerateEmbedding".
//
// Bodies are sliced from out.Root — the LOCAL checkout. That works because the
// live code-indexer workflow's first step is analyse_repo_local, which fetches
// the tarball into this pod's own temp dir and deliberately does NOT clean it up,
// so out.Root is a real path THIS process can read. It would NOT work under the
// original request_repo_analysis wiring (seed 118), where the analyser adapter
// parses in its own pod and returns line spans whose root does not exist here:
// every read would fail and every body would be NULL. That is a degrade, not a
// break — and the log line below says which happened rather than leaving it to
// be inferred from a column of NULLs.
func flattenSymbols(out analysis.Output, logger *zap.Logger) []symbolRow {
	var rows []symbolRow
	filesRead, fileReadErrs, bodiesSliced, sliceErrs := 0, 0, 0, 0
	var firstReadErr error

	for _, f := range out.Files {
		// ONE read per FILE, not per symbol: a file with 40 functions is read
		// once and sliced 40 times. No re-parse — the spans are already recorded.
		src, srcErr := readIndexedFile(out.Root, f.Path)
		if srcErr != nil {
			fileReadErrs++
			if firstReadErr == nil {
				firstReadErr = srcErr
			}
		} else {
			filesRead++
		}

		slice := func(start, end int) *string {
			if srcErr != nil {
				return nil
			}
			b, err := analysis.SliceLines(src, start, end)
			if err != nil {
				// A bad span is the analyser and the file disagreeing. Leave the
				// body NULL: a wrong body is worse than a missing one, because a
				// missing one is visible in the coverage count and a wrong one is
				// only visible to whoever acts on it.
				sliceErrs++
				return nil
			}
			bodiesSliced++
			return &b
		}

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
				body: slice(fn.StartLine, fn.EndLine),
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
				body: slice(td.StartLine, td.EndLine),
			})
		}
	}

	if logger != nil {
		fields := []zap.Field{
			zap.String("root", out.Root),
			zap.Int("files_read", filesRead),
			zap.Int("file_read_errors", fileReadErrs),
			zap.Int("bodies_sliced", bodiesSliced),
			zap.Int("slice_errors", sliceErrs),
			zap.Int("symbols", len(rows)),
		}
		if firstReadErr != nil {
			fields = append(fields, zap.NamedError("first_read_error", firstReadErr))
		}
		if bodiesSliced == 0 && len(rows) > 0 {
			// Loud: every content check will keep behaving as it did before this
			// change, and nothing else in the system would say so.
			logger.Warn("index_code_symbols: NO symbol bodies could be sliced — "+
				"code_symbols.body will be NULL for this repo and content checks stay declaration-only. "+
				"Expected cause: the workflow's analyse step is not analyse_repo_local, so out.Root is not a local checkout", fields...)
		} else {
			logger.Info("index_code_symbols: sliced symbol bodies from the local checkout", fields...)
		}
	}
	return rows
}

// readIndexedFile reads one analysed file from the checkout root. Paths in the
// analyser Output are slash-relative to Root (analysis.FileInfo.Path), so they
// are converted with filepath.FromSlash — the same conversion ReadSymbolBody
// does, kept identical on purpose. An empty root is an error rather than a read
// of the process's working directory, which would silently index whatever
// happened to sit at that relative path in the container image.
func readIndexedFile(root, slashRelPath string) ([]byte, error) {
	if root == "" {
		return nil, fmt.Errorf("no checkout root in the analysis output (bodies unavailable)")
	}
	return os.ReadFile(filepath.Join(root, filepath.FromSlash(slashRelPath)))
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
	// lookup_code_symbols is a SCOPE SEEDER: its rows become "path:Symbol" scope
	// entries (diagnose_assemble_bundle_action.go scopeFromCodeResults) which the
	// assembler slices into Go bodies. A row with no sliceable Go body can never be
	// a valid result, so non-code kinds are excluded at the source rather than
	// filtered downstream. Reuses codeKindsCSV — the SAME allow-list the D12 guard
	// tags with — so there is exactly one answer to "what is code".
	args = append(args, codeKindsCSV)
	query += fmt.Sprintf(` AND kind = ANY(string_to_array($%d, ','))`, len(args))
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
	// Same scope-seeder invariant as the vector path above, and deliberately the
	// same list: two searches feeding one consumer must not disagree about which
	// rows are eligible.
	args = append(args, codeKindsCSV)
	query += fmt.Sprintf(` AND kind = ANY(string_to_array($%d, ','))`, len(args))
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
