# Architecture Review: LLM Optimization Proposals

## 1. The Field Path Resolution Problem

Before anything else — the codebase has **at least 18 functions** that resolve dot-separated
field paths from nested maps. This is the single biggest code hygiene issue I can see,
and my initial proposal made it worse by adding a 19th.

### What exists in datahelpers (the right place):

| Function | Signature | Notes |
|----------|-----------|-------|
| `ExtractNestedField` | `(data, path) → interface{}` | Auto-unwraps `.response` wrappers |
| `ExtractNestedFieldString` | `(data, path) → string` | String convenience wrapper |
| `ExtractNestedFieldMap` | `(data, path) → map` | Map convenience wrapper |
| `GetFieldFromPath` | `(data, path, logger) → interface{}, error` | Logged, returns error |
| `GetFieldFromPathWithDefault` | `(data, path, default, logger) → interface{}` | With default |
| `GetValueByPath` | `(data, path, logger) → interface{}, bool` | Returns found bool |
| `FindByPath` | `(data, path, logger) → interface{}` | Recursive unwrap + fallbacks |
| `getFieldByPath` | `(data, path) → interface{}` | Private, simple traversal |
| `resolveFieldPath` | `(data, path) → interface{}` | Private, with array index support |

### What exists scattered across the actions package (shouldn't be there):

| Function | File | Difference from datahelpers version |
|----------|------|-------------------------------------|
| `resolveFieldPath` | entity_state_actions.go | Same as datahelpers private version |
| `resolveFieldPath` | workflow_actions.go | Duplicate of above |
| `resolveFieldPathForSpawn` | spawn_actions.go | Args swapped (path, data) vs (data, path) |
| `resolveFieldPathCallAgent` | call_agent.go | Returns string instead of interface{} |
| `resolveFieldPathQuestionnaire` | fetch_agent_questionnaire.go | Adds logger |
| `resolveFieldPathWithStepData` | fetch_agent_questionnaire.go | Different nesting logic |
| `resolveFieldValue` | conditional_branch_action.go | Adds logger, returns interface{} |
| `extractFieldValue` | multipage_actions.go | Returns string |
| `getNestedFieldValue` | (somewhere in actions) | Returns interface{}, error |

### My initial proposal added:

| Function | File | What it duplicates |
|----------|------|--------------------|
| `resolveRAGFieldPath` | rag_actions.go | `ExtractNestedFieldString` exactly |
| `nullStr` | rag_actions.go | `NullableString` in datahelpers |
| `truncateStr` | rag_actions.go | `TruncateString` in datahelpers |
| `nullIfEmpty` | llm_call_logger.go | `NullableString` in datahelpers (also exists in helpers.go) |
| `truncateForLog` | llm_call_logger.go | `TruncateString` + suffix |

### What to do:

**Immediate (for my new code):** Use the existing datahelpers functions. No new path
resolution functions. Specifically:

- `resolveRAGFieldPath` → use `datahelpers.ExtractNestedFieldString`
- `nullStr` / `nullIfEmpty` → use `datahelpers.NullableString`
- `truncateStr` / `truncateForLog` → use `datahelpers.TruncateString`

**Gradual (tech debt):** The 9+ duplicates in the actions package should migrate to
calling datahelpers. Each one is slightly different because they were written at different
times. `ExtractNestedField` is the canonical one — it handles the `.response` unwrapping
that most of the others try to do. The actions-package versions could become thin wrappers
or be replaced entirely. This is a separate cleanup task, not part of this work.

---

## 2. Reviewing Each Proposal Against Agent Checklist

### Item 1: Model Upgrades (SQL only)

**Assessment: Fine as-is.** This is just SQL updates to agent_definitions.default_config.
No new code, no new agents, no architectural concerns.

### Item 2: LLM Call Logging

**Assessment: Correctly scoped as infrastructure, not an agent.**

This is a cross-cutting concern that belongs inside the `execute_llm_prompt` action.
It's not an agent, it's not a workflow — it's instrumentation. The fire-and-forget
goroutine pattern is right for this.

**Concerns to address:**

1. **Table growth**: The `llm_call_log` table will grow fast. The cleanup function exists
   but nothing calls it. Need either `pg_cron` or a step in the maintenance-catch-all
   agent's workflow.

2. **Import path**: `llm_call_logger.go` should import from `datahelpers` for
   `NullableString` and `TruncateString` rather than defining its own. This means
   the actions package imports datahelpers (which it already does).

3. **The logger shouldn't have its own DB helpers**: Remove `nullIfEmpty`, `nullIfZero`,
   `truncateForLog` from the logger file. Use datahelpers where possible, and for
   `nullIfZero` (which genuinely doesn't exist yet), add it to datahelpers since it's
   a generally useful utility.

### Item 3: Ollama Provider

**Assessment: Correctly scoped as infrastructure.**

Same interface, new backend. The `ollama.go` file is a clean implementation of
`AIService`. The `createAIClient` switch addition is minimal.

**Concerns:**

1. **No health check**: If Ollama is down, every RAG/embedding call fails. Should add a
   simple health check (GET `/api/tags`) that runs on startup or first use. Not blocking
   for initial deployment but should be a fast follow.

2. **Model pull**: Ollama needs models pulled before they can be used. The deployment guide
   covers this but there's no code-level validation. If someone deploys without pulling
   the model, errors won't be obvious. Consider a startup check that verifies the
   configured model exists.

3. **Embedding dimension**: The `knowledge_base.embedding` column is `vector(768)` which
   is correct for nomic-embed-text. If someone switches to a model with different
   dimensions (e.g., text-embedding-ada-002 at 1536), they'll get silent failures.
   Document this clearly. Could also store the dimension per-row but that prevents the
   ivfflat index from working efficiently.

### Item 4: RAG Actions

**Assessment: Needs rethinking.** The actions themselves (`rag_lookup`, `rag_index`) are
correct as actions. But I was sloppy about two things:

**Problem A: Duplicate helpers.** Already covered above — use datahelpers.

**Problem B: Missing the agent boundary.**

Looking at the agent checklist:

> "Each specialist agent is self-contained and independently callable.
>  The agent handles its own data gathering rather than relying on callers
>  to provide everything."

The `rag_index` action does chunking + embedding + storage. That's fine as a building
block. But the *process* of building the knowledge base (decide what to scrape → scrape
→ chunk → embed → store → periodically refresh) is an agent's domain.

The right decomposition:

```
Actions (building blocks):
  rag_index   — chunk, embed, store content in knowledge_base
  rag_lookup  — embed query, vector search, return context

Future agent (owns the knowledge-building domain):
  knowledge-indexer agent
    workflow:
      1. load_indexing_targets  — query sites/URLs to index
      2. web_scrape             — existing action
      3. rag_index              — new action
      4. complete

  Called by:
    - site-maintenance-orchestrator (after discovery, index competitor sites)
    - standalone (manually index a URL or content)
    - build pipeline (after scraping exemplar sites)
```

For now, we implement the actions. The agent comes when we have a use case that
exercises the full pipeline. This follows the checklist principle of "Reuse Before
Creating" — don't build an agent until the workflow demands one.

**Problem C: The rag_lookup action embeds the query.**

This means every RAG lookup needs Ollama running with the embedding model. If we're
just doing keyword search or if Ollama is down, it fails completely. Consider:

- The `knowledge_base` table already has a trigram index (`idx_kb_content_trgm`)
- Add a fallback: if embedding fails, fall back to trigram text search
- This makes RAG degradation graceful rather than cliff-edge

---

## 3. Maintenance Concerns

### What could go wrong over time:

**llm_call_log table bloat:**
- With 30 agents making LLM calls, and multiple sites building per day, this grows fast.
- Prompts and responses can be 10-50KB each.
- At 1000 LLM calls/day × 30KB average = 30MB/day = ~1GB/month.
- The cleanup function exists but nothing calls it.
- **Fix:** Either pg_cron schedule, or add cleanup as a step in maintenance-catch-all.

**knowledge_base index rebuild:**
- ivfflat index on pgvector needs periodic REINDEX after significant data changes.
- With 100+ clusters (`lists = 100`), index build time grows with data volume.
- **Fix:** Schedule REINDEX CONCURRENTLY in the catch-all or a cron job.

**Embedding model lock-in:**
- Column is `vector(768)`. Changing models means re-embedding everything.
- **Fix:** Store `embedding_model` per row (already in schema). When model changes,
  run a migration job to re-embed. The `embedding_model` column lets you query
  "which rows need re-embedding."

**Ollama availability:**
- Single replica, no persistence concern (models re-pull), but availability matters
  for both RAG lookup and indexing.
- **Fix:** PVC for model storage (in deployment guide). Consider 2 replicas if demand
  grows. Add readiness probe.

### Dependency chain:

```
Content agents → rag_lookup → Ollama (embedding) → pgvector (search)
                                  ↓
                            Needs: model pulled, service running, GPU (optional)

Build pipeline → rag_index → Ollama (embedding) → pgvector (insert)
                                  ↓
                            Same dependencies
```

Both paths depend on Ollama. If Ollama is down:
- rag_lookup should fall back to trigram search (degraded but functional)
- rag_index should queue content for later indexing (or skip gracefully)

Neither should hard-fail the parent workflow.

---

## 4. Revised Plan

### What to change from my initial proposal:

1. **Remove all duplicate helpers from rag_actions.go and llm_call_logger.go.**
   Import from datahelpers instead.

2. **Add `NullableInt` to datahelpers** (the one genuinely new utility — nullIfZero).

3. **Add trigram fallback to rag_lookup** so it doesn't hard-depend on Ollama.

4. **Make rag_index failures non-fatal** — if embedding fails, log and skip, don't
   error the workflow.

5. **Add a note in the knowledge_base migration** about embedding dimension being
   model-dependent.

6. **Don't build a knowledge-indexer agent yet.** The actions are sufficient. The agent
   comes when we have a pipeline that needs orchestration (scrape → index → refresh).

### What stays the same:

- Model upgrades (SQL) — unchanged
- LLM call logging table schema — unchanged
- Anthropic client token capture — unchanged
- Ollama provider — unchanged
- Registry additions for rag_lookup/rag_index — unchanged
- Training data export queries — unchanged

### Deployment order is still:
1. SQL migrations (model upgrades + tables)
2. Go code (logging + Ollama provider + RAG actions)
3. Deploy Ollama (when ready for RAG/local models)
4. Wire RAG into workflows (when knowledge base has content)