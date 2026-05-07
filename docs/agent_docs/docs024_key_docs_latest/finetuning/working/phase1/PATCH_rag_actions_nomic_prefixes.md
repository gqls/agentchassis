# PATCH — nomic task prefixes for rag_actions.go

Applies `search_document:` and `search_query:` task prefixes to text sent
to nomic embedding models. Per nomic docs, these prefixes materially
improve retrieval quality. Our empirical test on 2026-04-21 (FOCUS §2.4b)
confirmed: without prefixes, wrong ranking on a BOAS-specific query.
With prefixes, correct ranking with a wider margin.

Detection is based on the configured embedding model string — if it
starts with `nomic-embed-`, prefixes are applied. Any other embedding
provider/model is untouched (safe for later provider swaps).

**File:** `platform/orchestration/actions/rag_actions.go`

**Scope:** Three changes — one helper, one line in `RAGLookupAction`,
one line in `RAGIndexAction`. The chunk stored in `knowledge_base` is
still the original un-prefixed content; only the text going to
`GenerateEmbedding` gets the prefix.

---

## Change 1 — add helper at the bottom of the file

In the helpers section (near `resolveRAGConfigField`, `chunkContent`, etc.),
add this function:

```go
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
```

`strings` is already imported at the top of the file (used by
`strings.Join` in `RAGLookupAction`). No new imports needed.

---

## Change 2 — RAGLookupAction

**Line 14029 of the context file** (approximate line in your repo — search
for the call site `embClient.GenerateEmbedding(ctx, queryText)` inside
`RAGLookupAction`):

Before:
```go
	embClient, embErr := createRAGEmbeddingClient(ctx, config)
	if embErr == nil {
		queryEmbedding, embGenErr := embClient.GenerateEmbedding(ctx, queryText)
		if embGenErr == nil {
			logger.Info("rag_lookup: generated query embedding",
				zap.String("query_preview", truncateForLog(queryText, 100)),
				zap.Int("embedding_dims", len(queryEmbedding)),
				zap.String("collection", collection))
```

After:
```go
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
```

The changed lines:
- New line before the `GenerateEmbedding` call: build the prefixed version
- `GenerateEmbedding` now receives `promptText` (possibly prefixed) instead of `queryText`
- New `zap.Bool("prefix_applied", prefixApplied)` in the log line

The trigram fallback on line 14050 is untouched — it still uses the
original `queryText`, which is correct because trigram search is a text
match, not an embedding similarity.

---

## Change 3 — RAGIndexAction

**Line 14143 of the context file** (approximate — search for the call
site `embClient.GenerateEmbedding(ctx, chunk)` inside the `for i, chunk
:= range chunks` loop in `RAGIndexAction`):

Before:
```go
	for i, chunk := range chunks {
		// SHA256 for dedup
		hash := fmt.Sprintf("%x", sha256.Sum256([]byte(chunk)))

		// Generate embedding (non-fatal)
		var embeddingStr *string
		if embClient != nil {
			embedding, err := embClient.GenerateEmbedding(ctx, chunk)
			if err != nil {
				logger.Warn("rag_index: embedding failed for chunk, storing without",
					zap.Int("chunk", i), zap.Error(err))
				embeddingsFailed++
			} else {
				s := pgvectorString(embedding)
				embeddingStr = &s
			}
		}
```

After:
```go
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
			embedding, err := embClient.GenerateEmbedding(ctx, embedInput)
			if err != nil {
				logger.Warn("rag_index: embedding failed for chunk, storing without",
					zap.Int("chunk", i), zap.Error(err))
				embeddingsFailed++
			} else {
				s := pgvectorString(embedding)
				embeddingStr = &s
			}
		}
```

The changed bits:
- `prefixAppliedOnce` flag outside the loop to log the prefix activation exactly once per action run (not per chunk — would spam logs)
- `embedInput` built with `applyNomicPrefix` and passed to `GenerateEmbedding`
- `hash` is still computed on the raw `chunk`, not the prefixed version
- The chunk stored in the DB (in the `INSERT INTO knowledge_base ... VALUES ..., chunk, ...` lines below) is still `chunk`, not `embedInput` — no change needed there

---

## What this fix does not change

- `knowledge_base` schema — no change
- Row storage format — chunks stored as original text
- Trigram fallback — still searches original `queryText`
- `embedding_model` column — still stored as the configured model name
- Any existing workflows, agent definitions, or Kafka envelopes

---

## Building and deploying

Standard chassis flow:

```bash
# from the agent chassis repo root
git add platform/orchestration/actions/rag_actions.go
git commit -m "rag_actions: apply nomic task prefixes to embeddings

Nomic embedding models (nomic-embed-text variants) produce materially
better retrieval when text is prefixed with 'search_document: ' (for
indexed content) or 'search_query: ' (for queries). Empirical test
(FOCUS §2.4b, 2026-04-21) confirmed correct ranking on BOAS query
only with prefixes applied.

- Added applyNomicPrefix helper with nomic-* model detection
- RAGLookupAction prefixes queries before embedding
- RAGIndexAction prefixes chunks before embedding (chunk storage
  and hash computation remain on original text)
- Trigram fallback untouched
- Logs prefix_applied flag for observability

Ref: FOCUS §2.4b action items"

git push
```

GitHub Actions picks this up and produces a new chassis image. Then:

```bash
# (after the image is pushed to the registry)
make update-kustomization-images IMAGE_TAG=v1.0.XXX
make deploy-agents IMAGE_TAG=v1.0.XXX
```

---

## Test plan after deploy

The goal is to confirm the prefix is being applied in production, not
just that nothing crashed. Three checks:

### A. Chassis logs show the prefix was applied

Re-run the rag-test-agent from flywheel B step 3 (same Kafka trigger,
different content / query). Then:

```bash
kubectl -n ai-persona-system logs -l app=agent-chassis --tail=500 \
  | grep -E "rag_(lookup|index)" | grep "prefix_applied"
```

Expected: at least one log line showing `"prefix_applied": true`.

### B. Retrieval quality improves on the known test case

Run the step-2 manual test again, but this time use the rag-test-agent
instead of manual curl — the agent will apply prefixes automatically.
Index the same four content pieces, query with the same query, expect
French Bulldog ranked first with a wider margin than Labrador.

Or — faster smoke test — just verify that indexing + lookup still
produce sane similarity scores on one piece of content. Regression
check, not quality test.

### C. No regression on non-nomic configurations

If any workflow in the system uses a non-nomic embedding model (search
agent_definitions for `embedding_service`), re-run it and confirm no
prefix is applied:

```bash
# Search definitions for non-nomic embedding models
kubectl -n ai-persona-system exec -i pod/postgres-clients-0 -- \
  psql -U clients_user -d clients_db -c "
SELECT type, default_config #> '{workflow,steps}' ?| ARRAY['rag_lookup','rag_index'] as uses_rag
FROM agent_definitions
WHERE default_config::text ~ 'embedding_service'
  AND default_config::text !~ 'nomic-embed-'
  AND deleted_at IS NULL;"
```

Expected: zero rows, or any rows that come back should still work
correctly with `prefix_applied: false` in their logs.

---

## Risk assessment

- **Risk: prefixes break an existing RAG workflow.** Low. The only
  existing RAG-using agent we found was our test agent (from
  `uses_rag_lookup`/`uses_rag_index` query in step 0). If more appear,
  they all use nomic-embed-text and benefit from prefixes.
- **Risk: old un-prefixed `knowledge_base` content retrieved poorly
  after fix.** Current `knowledge_base` is empty-ish (test rows only,
  which we cleaned up). Near-zero data at stake.
- **Risk: trigger edge case where content already starts with
  "search_document: "** — helper guards against double-prefixing.
- **Rollback:** revert the three code changes and rebuild. Chassis image
  tag can also be pinned to the pre-fix version via `kubectl set image`
  if we need to revert fast without rebuilding.
