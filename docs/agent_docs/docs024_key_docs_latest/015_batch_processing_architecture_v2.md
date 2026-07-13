# 015 — Batch Processing Architecture

## Date: 2026-03-26 (v4: 2026-04-12)

---

## The Numbers

Anthropic's Message Batches API: 50% discount on input AND output tokens, stacks
with prompt caching (up to 95% savings combined). Up to 10,000 requests per batch.
Results within 24 hours, often much faster.

Current spend: ~$120 per 4 domains (19.6M input, 4.2M output). At 50% discount,
that drops to ~$60 per 4 domains. At scale (2,000 domains), the difference is
$15,000 vs $7,500+ per cycle.

---

## The Problem

Our agents currently work synchronously:

```
agent starts → calls LLM → waits 3-10 seconds → processes result → done
```

Batch API is asynchronous:

```
submit batch → poll every N minutes → results arrive 1-24 hours later → process
```

The system isn't built for 24-hour waits. Agent pods time out, orchestration
states expire, Kafka messages TTL. The stale orchestration sweeper runs every
60 seconds and would kill anything parked that long. We need a pattern that
sidesteps this entirely rather than fighting the orchestration engine's
timeout assumptions.

---

## Categorising Our LLM Calls

Not everything benefits equally from batching.

### Good batch candidates (non-blocking, can wait hours)
- Improvement loop audits (visual-design-auditor, content-quality-auditor, site-review-agent)
- Blog content planning and generation
- Content rewrites triggered by audits
- Companies House LLM review (already batched in spirit)
- Vet price parsing (already discussed as batch-tolerant)
- Any scheduled/maintenance task

### Poor batch candidates (user-facing or blocking)
- Initial site build (user expects result in minutes)
- Briefing-agent (blocks the build pipeline)
- Site classifier (blocks everything downstream)
- Chief strategist (one call per domain, user is watching)
- Page content writer during initial build (pages depend on each other)
- Any HITL/interactive workflow

### Decision rule for ambiguous cases

If the originating trigger is a **scheduled task**, use batch. If it's a
**user-initiated build or HITL response**, use sync. This is determined at
workflow-config time (which action the step uses), not at runtime — avoids
fragile detection logic.

Applies to:
- Improvement loop fixes → batch (nobody watching, scheduler-triggered)
- Asset generation during maintenance → batch
- Research agent calls during initial build → sync
- Research agent calls during improvement loop → batch

---

## Core Concept: Universal LLM Work Queue

The queue is not just an Anthropic batch queue. It's a **universal LLM work
queue** that can route requests to different backends depending on what's
available, what the request needs, and what's cheapest.

```
                                    ┌─────────────────────┐
                                    │ Anthropic Batch API  │
                                    │ (50% discount, async)│
                              ┌────▶│                     │
                              │     └─────────────────────┘
┌──────────┐    ┌─────────┐   │     ┌─────────────────────┐
│ Agents    │───▶│  Queue   │──┼────▶│ GPU endpoint (Llama) │
│ (fire &   │    │  Table   │  │     │ (spin up, drain,     │
│  forget)  │    │          │  │     │  shut down)          │
└──────────┘    └─────────┘   │     └─────────────────────┘
                              │     ┌─────────────────────┐
                              ├────▶│ Anthropic direct API │
                              │     │ (sync, full price)   │
                              │     └─────────────────────┘
                              │     ┌─────────────────────┐
                              └────▶│ Image generation     │
                                    │ (Flux, DALL-E, etc.) │
                                    └─────────────────────┘
```

The **submitter** reads from the queue and decides how to route based on:
1. The request's `provider` and `model` fields
2. The `ai_endpoint_health` table (is this endpoint up?)
3. The `llm_batch_config` table (is this provider enabled? what mode?)
4. The `llm_batch_agent_config` table (is this agent type opted in?)
5. Whether the item has been escalated to urgent

When batch is **off** for an agent type (the default), the queue action
executes the entire path inline — render, call LLM, execute callback —
behaving identically to the old workflow. The queue row is still written
(status `sync_executed`) for observability and cost tracking.

---

## Tables

### Queue Table

```sql
CREATE TABLE llm_batch_queue (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    -- Traceability
    client_id TEXT NOT NULL,
    correlation_id TEXT,
    site_id UUID,
    agent_type TEXT NOT NULL,
    step_name TEXT,
    orchestration_id TEXT,
    work_item_id UUID,

    -- Request (provider-agnostic)
    provider TEXT NOT NULL DEFAULT 'anthropic',  -- anthropic, ollama, gpu, openai, replicate...
    model TEXT NOT NULL,                          -- claude-sonnet-4-6, llama-3.3-70b, flux-dev...
    request_type TEXT NOT NULL DEFAULT 'text',    -- text, image, embedding
    system_prompt TEXT,
    prompt TEXT NOT NULL,                         -- RENDERED prompt, not a template
    max_tokens INT DEFAULT 4000,
    temperature NUMERIC(3,2) DEFAULT 0.7,
    request_params JSONB DEFAULT '{}',           -- provider-specific: image dimensions, LoRA, etc.

    -- Callback (what to do with the result)
    -- All values in callback_config are RESOLVED at queue time.
    -- Stores "site_id": "a1b2c3d4-..." not "site_id": "site_record.site_id"
    callback_action TEXT NOT NULL,
    callback_config JSONB DEFAULT '{}',

    -- Routing and priority
    batch_group TEXT GENERATED ALWAYS AS (
        provider || ':' || model || ':' || COALESCE(site_id::text, 'no-site')
    ) STORED,
    priority INT DEFAULT 50,                     -- 0=lowest, 100=highest
    urgent BOOLEAN DEFAULT false,                -- escalation flag

    -- State
    status TEXT NOT NULL DEFAULT 'pending',
    -- pending:       waiting to be submitted/processed
    -- submitted:     sent to async API, waiting for results
    -- processing:    being processed right now (GPU drain mode)
    -- complete:      result received and callback executed
    -- failed:        callback failed or API error
    -- expired:       abandoned after max retries
    -- superseded:    result arrived but item already handled via urgent path
    -- sync_executed: batch was off, executed inline (dry-run/fallback record)
    provider_batch_id TEXT,
    provider_request_id TEXT,

    -- Result (always populated, even for sync_executed — for cost tracking)
    response_text TEXT,
    response_url TEXT,                           -- for image generation
    input_tokens INT,
    output_tokens INT,
    error_message TEXT,
    retry_count INT DEFAULT 0,

    -- Timestamps
    created_at TIMESTAMPTZ DEFAULT NOW(),
    submitted_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ
);

CREATE INDEX idx_batch_queue_pending ON llm_batch_queue(provider, status, priority DESC, created_at)
    WHERE status = 'pending';
CREATE INDEX idx_batch_queue_urgent ON llm_batch_queue(status, urgent)
    WHERE urgent = true AND status = 'pending';
CREATE INDEX idx_batch_queue_submitted ON llm_batch_queue(provider_batch_id)
    WHERE status = 'submitted';
CREATE INDEX idx_batch_queue_processing ON llm_batch_queue(status)
    WHERE status = 'processing';
CREATE INDEX idx_batch_queue_failed ON llm_batch_queue(status, retry_count)
    WHERE status = 'failed';
CREATE INDEX idx_batch_queue_site ON llm_batch_queue(site_id, status);
CREATE INDEX idx_batch_queue_client ON llm_batch_queue(client_id, status);
```

### Batch Provider Config Table

Global and per-provider settings for the submitter. Controls on/off,
mode (batch vs drain), limits, and GPU auto-shutdown.

```sql
CREATE TABLE llm_batch_config (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    provider TEXT NOT NULL UNIQUE,        -- 'global', 'anthropic', 'gpu', 'replicate', etc.

    -- On/off
    enabled BOOLEAN DEFAULT true,         -- false = submitter skips this provider entirely
    mode TEXT DEFAULT 'batch',            -- batch | drain | disabled
                                          -- batch:    normal async batch submission
                                          -- drain:    sync calls as fast as possible (GPU mode)
                                          -- disabled: same as enabled=false but more explicit

    -- Limits
    max_concurrent_batches INT DEFAULT 5,
    max_requests_per_batch INT DEFAULT 10000,
    submit_interval_seconds INT DEFAULT 300,    -- how often submitter checks (default 5 min)
    drain_concurrency INT DEFAULT 4,            -- parallel workers in drain mode

    -- Endpoint (for drain mode / non-Anthropic providers)
    endpoint_url TEXT,                    -- e.g. http://gpu-llama:8080/v1/chat/completions
    api_key_env_var TEXT,                 -- e.g. ANTHROPIC_API_KEY, GPU_API_KEY

    -- Auto-shutdown (for GPU spin-up scenarios)
    drain_until TIMESTAMPTZ,             -- stop draining after this time
    drain_stop_when_empty BOOLEAN DEFAULT true,  -- stop when queue is empty

    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Seed defaults
INSERT INTO llm_batch_config (provider, enabled, mode) VALUES
    ('global', true, 'batch'),           -- global kill switch
    ('anthropic', true, 'batch'),
    ('gpu', false, 'drain'),
    ('replicate', false, 'batch');
```

### Per-Agent-Type Control Table

Controls which agent types are opted in to batching. Default is **off for
everything** — agents must be explicitly opted in. The `batch_group` column
lets you enable/disable groups of agents with a single UPDATE.

```sql
CREATE TABLE llm_batch_agent_config (
    agent_type TEXT PRIMARY KEY REFERENCES agent_definitions(type),
    batch_enabled BOOLEAN NOT NULL DEFAULT false,
    batch_group TEXT,                     -- logical grouping: 'audits', 'enrichment', 'content'
    notes TEXT,                           -- human notes: "enabled 2026-04-10, watching for 48hrs"
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Seed all batch-eligible agent types, all OFF
INSERT INTO llm_batch_agent_config (agent_type, batch_enabled, batch_group, notes) VALUES
    ('visual-design-auditor',   false, 'audits',     'First candidate for batch'),
    ('content-quality-auditor', false, 'audits',     'Second candidate'),
    ('site-review-agent',       false, 'audits',     'Third candidate'),
    ('ch-llm-reviewer',         false, 'enrichment', 'Already batch-oriented'),
    ('vet-price-parser',        false, 'enrichment', 'Batch-tolerant');
```

### Three-Gate Resolution (in QueueLLMBatchAction)

All three gates must be open for an item to be queued async. Any gate
closed = sync fallback (identical behaviour to old workflow).

```
1. Is llm_batch_config WHERE provider='global' → enabled = false?
   → Sync fallback. Full stop.

2. Does llm_batch_agent_config have a row for this agent_type?
   → No row: sync fallback. Must be explicitly opted in.
   → Row exists, batch_enabled = false: sync fallback.
   → Row exists, batch_enabled = true: proceed to gate 3.

3. Is the chosen provider's llm_batch_config → enabled = true?
   → No: sync fallback.
   → Yes: write to queue with status 'pending'.
```

---

## What Happens at Queue Time

`QueueLLMBatchAction` does the following regardless of whether batch is on or off:

1. **Render the prompt template** against collected_data using existing
   `datahelpers.RenderPromptTemplate`. The queue stores the **rendered prompt**,
   not the template. If the template uses `{{.site_context}}` or
   `{{range .existing_posts}}`, all of that is resolved now, while
   collected_data is still available.

2. **Resolve callback_config references.** The workflow config might say
   `"site_id": "site_record.site_id"`. The action resolves this against
   collected_data to get the actual UUID. The queue stores
   `"site_id": "a1b2c3d4-..."`. Uses the same `ExtractFields` / dot-path
   resolution the orchestration engine already has.

3. **Write the queue row** with rendered prompt, resolved callback_config,
   all traceability fields from execution context.

4. **Check the three gates** (global → agent_type → provider).

5a. **If batch ON:** Return immediately. Workflow proceeds to `complete`.

5b. **If batch OFF (sync fallback):**
   - Call the LLM directly (reuse existing `createAIClient` / `GenerateText`)
   - Call `LogLLMCall` (writes to `llm_call_log` for training data)
   - Execute the callback action with the response
   - Update queue row: `status = 'sync_executed'`, populate response_text, tokens
   - Return. Workflow proceeds to `complete`.

This means sync fallback **proves the entire restructured pipeline** — prompt
rendering, callback_config resolution, callback execution — before you ever
turn batch on. The `sync_executed` rows are your dry-run evidence.

---

## Callback Contract

When the retriever (or sync fallback) executes a callback, it provides:

```go
type BatchCallbackParams struct {
    DB             *sql.DB
    Logger         *zap.Logger
    ResponseText   string                 // LLM response
    ResponseURL    string                 // for image generation
    CallbackConfig map[string]interface{} // resolved values from queue row
    QueueRowID     uuid.UUID              // for logging/tracing
    AgentType      string                 // from queue row
    CorrelationID  string                 // from queue row
}
```

A callback does **NOT** receive: collected_data, orchestration state, workflow
plan, or anything from the original agent's runtime context. Everything it needs
must be in `callback_config` (resolved at queue time) and the LLM response.

**Test for callback eligibility:** Can this action do its job given only a
database connection, the LLM response text, and a handful of resolved IDs
(site_id, work_item_id)? If yes, it can be a callback. If it needs data from
a previous workflow step that isn't in the database, it can't.

First batch-eligible callbacks all pass:
- `write_audit_findings`: needs site_id + response text → writes to site_work_items
- `write_ch_review_result`: needs company_number + response text → updates ch_vet_companies
- `store_generated_asset`: needs site_id + asset_type + response URL → writes to site assets

---

## The Workflow Restructure

When an agent becomes batch-eligible, its workflow changes shape:

**Before:**
```json
"run_audit": {
    "action": "execute_llm_prompt",
    "config": { "ai_service": {...}, "prompt_template": "..." },
    "next_step": "write_findings",
    "output_field": "audit_result"
},
"write_findings": {
    "action": "write_audit_findings",
    "config": { "site_id": "site_record.site_id" },
    "next_step": "complete"
}
```

**After:**
```json
"run_audit": {
    "action": "queue_llm_batch",
    "config": {
        "prompt_template": "...",
        "system_prompt": "You are a visual design auditor...",
        "model": "claude-sonnet-4-6",
        "max_tokens": 4000,
        "callback_action": "write_audit_findings",
        "callback_config": {
            "site_id": "site_record.site_id",
            "audit_source": "visual-design-audit",
            "vertical": "veterinary",
            "prompt_variant": "v3_visual_audit"
        }
    },
    "next_step": "complete"
}
```

The `write_findings` step disappears from the workflow. It becomes the
callback, executed either inline (batch off) or by the retriever (batch on).

For multi-provider with fallback:

```json
{
    "action": "queue_llm_batch",
    "config": {
        "provider_preference": ["gpu", "anthropic"],
        "model_map": {
            "gpu": "llama-3.3-70b",
            "anthropic": "claude-sonnet-4-6"
        },
        "prompt_template": "...",
        "callback_action": "write_audit_findings",
        "callback_config": { "site_id": "site_record.site_id" }
    }
}
```

`QueueLLMBatchAction` checks `llm_batch_config` for each provider in preference
order. First enabled provider wins. Items auto-route to GPU when it's up, fall
back to Anthropic batch when it's not — without changing workflow definitions.

For image generation:

```json
{
    "action": "queue_llm_batch",
    "config": {
        "provider": "replicate",
        "model": "flux-dev",
        "request_type": "image",
        "prompt_template": "{{.logo_prompt}}",
        "request_params": { "width": 1024, "height": 1024 },
        "callback_action": "store_generated_asset",
        "callback_config": {
            "site_id": "site_record.site_id",
            "asset_type": "logo"
        }
    }
}
```

---

## Operational Controls

### 1. Turning Batch On/Off Completely

```sql
-- Kill switch: all agents fall back to sync
UPDATE llm_batch_config SET enabled = false WHERE provider = 'global';

-- Re-enable
UPDATE llm_batch_config SET enabled = true WHERE provider = 'global';

-- Disable just Anthropic batching (GPU drain can still run)
UPDATE llm_batch_config SET enabled = false WHERE provider = 'anthropic';
```

When global is off, every `QueueLLMBatchAction` call executes inline. Tasks
still complete normally, just at full price. The queue table accumulates
`sync_executed` rows as a record of what would have been batched.

### 2. Turning Batch On/Off by Agent Group

```sql
-- Enable batch for all audit agents
UPDATE llm_batch_agent_config SET batch_enabled = true, updated_at = NOW()
WHERE batch_group = 'audits';

-- Disable one specific agent
UPDATE llm_batch_agent_config SET batch_enabled = false, updated_at = NOW()
WHERE agent_type = 'content-quality-auditor';

-- Enable enrichment tasks
UPDATE llm_batch_agent_config SET batch_enabled = true,
    notes = 'enabled for enrichment batch test', updated_at = NOW()
WHERE batch_group = 'enrichment';

-- Check what's enabled
SELECT agent_type, batch_enabled, batch_group, notes, updated_at
FROM llm_batch_agent_config ORDER BY batch_group, agent_type;
```

### 3. GPU Spin-Up: Drain Mode

Spin up H100s, blast through the queue, shut down.

```sql
-- Spin up GPUs, then:
UPDATE llm_batch_config SET
    enabled = true,
    mode = 'drain',
    endpoint_url = 'http://gpu-llama-h100:8080/v1/chat/completions',
    drain_concurrency = 8,              -- 8 parallel workers
    drain_until = NOW() + INTERVAL '2 hours',
    drain_stop_when_empty = true
WHERE provider = 'gpu';
```

The submitter sees `mode = 'drain'` and switches behaviour:
- Instead of building JSONL and submitting to a batch API, it **runs a worker
  pool** that pulls items from the queue and makes direct sync calls to the
  GPU endpoint
- `drain_concurrency` controls parallelism (match to GPU capacity)
- Each worker: picks a `pending` row (FOR UPDATE SKIP LOCKED), sets
  `status = 'processing'`, calls the endpoint, executes the callback,
  sets `status = 'complete'`
- Stops when: queue is empty (if `drain_stop_when_empty`), or current time
  exceeds `drain_until`, whichever comes first

```sql
-- When done (or GPUs auto-stop):
UPDATE llm_batch_config SET enabled = false, mode = 'disabled'
WHERE provider = 'gpu';

-- Any remaining 'pending' items for GPU-eligible models stay in the queue.
-- They can be picked up later by another drain session, or you can
-- reassign them to Anthropic batch:
UPDATE llm_batch_queue SET provider = 'anthropic'
WHERE provider = 'gpu' AND status = 'pending';
```

### 4. Priority Override (Pull from Batch)

A queued item becomes urgent — a dependency appeared, or you want to preview
a site that has pending audit results.

```sql
-- Escalate a specific item
SELECT escalate_batch_item('<item-id>');

-- Escalate all pending items for a site
SELECT escalate_site_batch('<site-id>');
```

**If still pending:** Submitter processes it immediately via direct sync call
on its next run (urgent items are checked first, every cycle).

**If already submitted to Anthropic:** Can't cancel individual items in a batch.
Submitter makes a parallel sync call. When the batch result arrives later, the
retriever sees `status = 'complete'` (already handled) and marks it `superseded`.

```sql
CREATE OR REPLACE FUNCTION escalate_batch_item(item_id UUID)
RETURNS TEXT AS $$
DECLARE
    current_status TEXT;
BEGIN
    SELECT status INTO current_status FROM llm_batch_queue WHERE id = item_id;
    IF current_status = 'pending' THEN
        UPDATE llm_batch_queue SET urgent = true, priority = 100 WHERE id = item_id;
        RETURN 'escalated_pending';
    ELSIF current_status = 'submitted' THEN
        UPDATE llm_batch_queue SET urgent = true WHERE id = item_id;
        RETURN 'escalated_submitted_will_parallel';
    ELSE
        RETURN 'cannot_escalate_status_' || current_status;
    END IF;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION escalate_site_batch(target_site_id UUID)
RETURNS INT AS $$
DECLARE
    affected INT;
BEGIN
    UPDATE llm_batch_queue SET urgent = true, priority = 100
    WHERE site_id = target_site_id AND status IN ('pending', 'submitted');
    GET DIAGNOSTICS affected = ROW_COUNT;
    RETURN affected;
END;
$$ LANGUAGE plpgsql;
```

### 5. Provider Capability Constraints

Different providers have different limits. The `llm_batch_config` table
captures these, and the submitter respects them:

```sql
-- Anthropic: 10,000 per batch, 200K context
UPDATE llm_batch_config SET
    max_requests_per_batch = 10000
WHERE provider = 'anthropic';

-- GPU Ollama: no batch API, context window depends on model
UPDATE llm_batch_config SET
    mode = 'drain',
    drain_concurrency = 4,
    max_requests_per_batch = 1  -- effectively: process one at a time per worker
WHERE provider = 'gpu';

-- Replicate: async but not batched, each request is its own prediction
UPDATE llm_batch_config SET
    mode = 'batch',            -- "batch" here means: submit async, poll later
    max_requests_per_batch = 1 -- each item = one prediction
WHERE provider = 'replicate';
```

### 6. Multi-Provider Routing

Different requests need different backends. The `provider` + `request_type`
fields control routing. The submitter has a **provider adapter** per backend:

```go
type ProviderAdapter interface {
    SubmitBatch(ctx context.Context, items []QueueItem, config BatchConfig) (batchID string, err error)
    PollBatch(ctx context.Context, batchID string, config BatchConfig) ([]BatchResult, error)
    SubmitSync(ctx context.Context, item QueueItem, config BatchConfig) (BatchResult, error)
}
```

Initial adapters:
- `AnthropicBatchAdapter` — /v1/messages/batches
- `AnthropicSyncAdapter` — /v1/messages (urgent items, sync fallback)
- `OllamaAdapter` — sync only (drain mode), calls local/GPU Ollama endpoint
- `ReplicateAdapter` — async prediction API (future)

New providers added without changing the queue table or agent workflows.

---

## Submitter Logic

Runs on the interval defined in `llm_batch_config` per provider.

```
for each provider in llm_batch_config where enabled = true:

    # 1. Urgent items first — always sync, regardless of mode
    urgent = SELECT FROM llm_batch_queue
             WHERE urgent = true AND status IN ('pending', 'submitted')
             AND provider = this_provider
             FOR UPDATE SKIP LOCKED

    for each urgent_item:
        if status = 'submitted':
            # Can't cancel from batch — make parallel sync call
        result = adapter.SubmitSync(urgent_item)
        LogLLMCall(...)
        execute_callback(urgent_item, result)
        mark complete

    # 2. Mode-specific processing
    if mode = 'drain':
        if drain_until is set AND NOW() > drain_until:
            UPDATE llm_batch_config SET enabled = false WHERE provider = this_provider
            continue
        run_drain_workers(provider, drain_concurrency)
        # Each worker: pick pending row (SKIP LOCKED), set 'processing',
        # call endpoint sync, LogLLMCall, execute callback, set 'complete'

    elif mode = 'batch':
        pending = SELECT FROM llm_batch_queue
                  WHERE status = 'pending' AND provider = this_provider
                  ORDER BY batch_group, priority DESC, created_at
                  LIMIT max_requests_per_batch
                  FOR UPDATE SKIP LOCKED

        if len(pending) > 0:
            # Group by model (Anthropic requires same model per batch)
            # Within each group, order by batch_group for cache adjacency
            for each model_group:
                batch_id = adapter.SubmitBatch(model_group)
                update rows: status = 'submitted', provider_batch_id = batch_id
```

---

## Retriever Logic

Runs every 10 minutes.

```
# 1. Poll async batches
for each provider in llm_batch_config where mode = 'batch' and enabled = true:
    outstanding = SELECT DISTINCT provider_batch_id FROM llm_batch_queue
                  WHERE status = 'submitted' AND provider = this_provider

    for each batch_id:
        results = adapter.PollBatch(batch_id)
        if results is nil: continue  # not ready yet

        for each result:
            row = find queue row by provider_request_id (= queue row id)

            if row.status = 'complete':
                # Already handled via urgent escalation
                mark 'superseded'
                continue

            write response_text/response_url, tokens to row
            LogLLMCall(...)  # → llm_call_log for training data
            execute_callback(row, result)
            if callback ok: mark 'complete'
            else: mark 'failed', increment retry_count

# 2. Retry failed callbacks (response already stored, no re-submission)
failed = SELECT FROM llm_batch_queue
         WHERE status = 'failed' AND retry_count < 3
         AND completed_at < NOW() - INTERVAL '10 minutes' * retry_count
         FOR UPDATE SKIP LOCKED

for each failed_item:
    re-execute callback (response_text is already in the row)
    if ok: mark 'complete'
    else: increment retry_count
    if retry_count >= 3: mark 'expired'
```

---

## Training Data Collection (RAG + LoRA)

`llm_call_log` is the single source for all prompt/response pairs — used for
RAG retrieval and future LoRA fine-tuning.

### All paths feed llm_call_log

| Path | Who calls LogLLMCall | When |
|---|---|---|
| Sync (execute_llm_prompt) | ExecuteLLMPromptAction | Immediately after LLM response |
| Sync fallback (batch off) | QueueLLMBatchAction | Inline during fallback execution |
| Batch (async) | Retriever | When processing each completed result |
| Drain (GPU) | Drain worker | After each sync call to GPU endpoint |
| Urgent escalation | Submitter | After urgent sync call |

### Fields that travel through callback_config

`llm_call_log` has columns that `llm_batch_queue` doesn't carry at top level:
`vertical`, `prompt_variant`, `rag_context_used`, `work_item_id`. These are
captured at queue time in `callback_config` (resolved to actual values) so the
retriever can pass them through to `LogLLMCall`:

```go
LogLLMCall(db, logger, LLMCallLogParams{
    AgentType:       row.AgentType,
    StepName:        row.StepName,
    OrchestrationID: row.OrchestrationID,
    CorrelationID:   row.CorrelationID,
    Model:           row.Model,
    Provider:        row.Provider,
    PromptRendered:  row.Prompt,
    ResponseText:    row.ResponseText,
    InputTokens:     row.InputTokens,
    OutputTokens:    row.OutputTokens,
    LatencyMs:       0,  // batch has no per-request latency
    Success:         true,
    WorkItemID:      callbackConfig["work_item_id"],
    Vertical:        callbackConfig["vertical"],
    PromptVariant:   callbackConfig["prompt_variant"],
    RAGContextUsed:  callbackConfig["rag_context_used"],
})
```

### Latency field

Anthropic's batch API doesn't report per-request latency. `latency_ms` will
be 0 for batch results. For training data this doesn't matter. For dashboards,
filter by `latency_ms > 0` for sync timings, or use `submitted_at → completed_at`
from the queue table for batch turnaround.

### Two tables, two purposes

| Table | Purpose | Query for |
|---|---|---|
| `llm_call_log` | Training data (RAG, LoRA), analytics | Prompt quality, model comparison |
| `llm_batch_queue` | Operational | Routing, status, retries, cost by provider |

Don't query `llm_batch_queue` for training data. Don't query `llm_call_log`
for batch operational status.

---

## Prompt Caching Strategy

Audit agents send the same site context (CSS, HTML samples, brief) with every
call. With batch + caching:
- First request in the batch pays cache write (1.25x input)
- Subsequent requests for the same site read from cache (0.1x input)
- Combined with 50% batch discount: up to 95% savings on input tokens

For a site with 3 audit calls sharing context, input cost goes from 3x to ~0.55x.

The submitter orders requests within each batch by `batch_group`
(provider:model:site_id), so all requests for the same site are adjacent in
the JSONL payload. System prompts containing site context should be identical
across audit types for the same site — use a shared `build_site_audit_context()`
helper.

---

## Monitoring

```sql
-- Overview dashboard
SELECT provider, status, COUNT(*),
       SUM(input_tokens) as total_input,
       SUM(output_tokens) as total_output,
       AVG(EXTRACT(EPOCH FROM completed_at - created_at)) as avg_latency_s
FROM llm_batch_queue
WHERE created_at > NOW() - INTERVAL '24 hours'
GROUP BY provider, status
ORDER BY provider, status;

-- Per-agent-type batch status (is my rollout working?)
SELECT agent_type, status, COUNT(*), MAX(completed_at) as latest
FROM llm_batch_queue
WHERE created_at > NOW() - INTERVAL '24 hours'
GROUP BY agent_type, status
ORDER BY agent_type, status;

-- Sync fallback volume (how much would batch save?)
SELECT agent_type, COUNT(*) as sync_calls,
       SUM(input_tokens) as input_tokens,
       SUM(output_tokens) as output_tokens
FROM llm_batch_queue
WHERE status = 'sync_executed'
  AND created_at > NOW() - INTERVAL '7 days'
GROUP BY agent_type;

-- Stuck batches
SELECT provider_batch_id, provider, COUNT(*), MIN(submitted_at)
FROM llm_batch_queue
WHERE status = 'submitted'
  AND submitted_at < NOW() - INTERVAL '6 hours'
GROUP BY provider_batch_id, provider;

-- Urgent items not yet processed
SELECT id, agent_type, provider, created_at
FROM llm_batch_queue
WHERE urgent = true AND status IN ('pending', 'submitted');

-- GPU drain progress
SELECT
    COUNT(*) FILTER (WHERE status = 'complete') as done,
    COUNT(*) FILTER (WHERE status = 'processing') as in_flight,
    COUNT(*) FILTER (WHERE status = 'pending') as remaining,
    MIN(created_at) as oldest_pending
FROM llm_batch_queue
WHERE provider = 'gpu';

-- Cost by provider and client
SELECT client_id, provider, model,
       COUNT(*) as requests,
       SUM(input_tokens) as input_tokens,
       SUM(output_tokens) as output_tokens
FROM llm_batch_queue
WHERE status IN ('complete', 'sync_executed')
  AND completed_at > NOW() - INTERVAL '7 days'
GROUP BY client_id, provider, model;

-- Failed items needing attention
SELECT id, agent_type, callback_action, error_message, retry_count
FROM llm_batch_queue
WHERE status IN ('failed', 'expired')
  AND created_at > NOW() - INTERVAL '7 days'
ORDER BY retry_count DESC;

-- Batch enablement status overview
SELECT ac.agent_type, ac.batch_enabled, ac.batch_group, ac.notes,
       COUNT(q.id) FILTER (WHERE q.status = 'sync_executed') as recent_sync,
       COUNT(q.id) FILTER (WHERE q.status = 'complete') as recent_batch,
       COUNT(q.id) FILTER (WHERE q.status = 'pending') as pending
FROM llm_batch_agent_config ac
LEFT JOIN llm_batch_queue q ON q.agent_type = ac.agent_type
    AND q.created_at > NOW() - INTERVAL '24 hours'
GROUP BY ac.agent_type, ac.batch_enabled, ac.batch_group, ac.notes
ORDER BY ac.batch_group, ac.agent_type;
```

---

## What NOT to Batch

Keep these on direct API calls (existing `execute_llm_prompt` path, never
switched to `queue_llm_batch`):
- Initial site builds (user is waiting)
- Briefing agent (blocks the pipeline)
- Site classifier (everything depends on it)
- Chief strategist (one call per domain, user is watching)
- Page content writer during initial build
- Any HITL/interactive workflow

These agents keep `execute_llm_prompt` in their workflows. The batch
infrastructure doesn't touch them.

The 50% saving applies to ~60-70% of total token spend (audits, maintenance,
content rewrites, batch processing). The remaining 30-40% (builds, classification)
stays synchronous.

---

## Implementation

### Go Code

**QueueLLMBatchAction** (~100 lines, `actions/batch_queue_action.go`):
- Renders prompt template against collected_data (reuse `RenderPromptTemplate`)
- Resolves callback_config references against collected_data (reuse `ExtractFields`)
- Writes queue row (always, even in sync fallback → `sync_executed`)
- Checks three gates: global config → agent_type config → provider config
- If batch on: return immediately
- If batch off: call LLM, LogLLMCall, execute callback, update row, return

**BatchCallbackDispatcher** (~60 lines, `actions/batch_callback.go`):
- Function registry mapping callback_action names to Go functions
- Constructs `BatchCallbackParams` from queue row + response
- Called by: QueueLLMBatchAction (sync fallback), retriever, drain workers,
  submitter (urgent)
- Each callback function: `func(BatchCallbackParams) error`

**ProviderAdapter interface** + adapters (~150 lines each):
- `AnthropicBatchAdapter`: /v1/messages/batches submit + poll
- `AnthropicSyncAdapter`: /v1/messages for urgent/fallback
- `OllamaAdapter`: sync calls to Ollama endpoint
- `ReplicateAdapter`: create prediction + poll (future)

**Submitter** (~150 lines, in kafka-scheduler):
- Registered in `scheduled_tasks`, default every 5 minutes
- Iterates enabled providers, urgent-first, then mode-specific
- Drain mode: worker goroutine pool, respects concurrency limit + drain_until

**Retriever** (~120 lines, in kafka-scheduler):
- Registered in `scheduled_tasks`, default every 10 minutes
- Polls each provider's outstanding batches
- LogLLMCall + callback for each result
- Retry sweep for failed callbacks

---

## Rollout Plan

### Phase 0: Deploy infrastructure, everything OFF

1. Run migration: create `llm_batch_queue`, `llm_batch_config`, `llm_batch_agent_config`
2. Deploy `QueueLLMBatchAction` + `BatchCallbackDispatcher` code
3. Global config: `enabled = true` (so the action runs its resolution logic)
4. All agent_type rows: `batch_enabled = false` (so everything sync-fallbacks)
5. Do NOT deploy submitter or retriever yet — not needed

At this point, no agent workflows have changed. Nothing is different.

### Phase 1: Switch first agent workflow, batch still OFF

1. Update `visual-design-auditor` workflow: `execute_llm_prompt` → `queue_llm_batch`
2. The action renders, calls LLM inline, executes callback, writes `sync_executed` row
3. Verify: audit findings still appear in `site_work_items`?
4. Verify: `llm_call_log` entries still appearing?
5. Inspect: `SELECT * FROM llm_batch_queue WHERE agent_type = 'visual-design-auditor'`
   — prompts look right? callback_config has resolved UUIDs not template refs?

This proves the restructured pipeline works before batch is turned on.

### Phase 2: Switch remaining batch-eligible agents, batch still OFF

1. Update `content-quality-auditor`, `site-review-agent`, `ch-llm-reviewer`
2. Same verification for each
3. Let it run for a day or two
4. Use "sync fallback volume" monitoring query to see what batch would save

### Phase 3: Deploy submitter + retriever, enable batch for one agent

1. Deploy submitter and retriever scheduled tasks
2. Enable batch for one agent:
   ```sql
   UPDATE llm_batch_agent_config
   SET batch_enabled = true, notes = 'Phase 3 test', updated_at = NOW()
   WHERE agent_type = 'visual-design-auditor';
   ```
3. Watch: items appearing as `pending`?
4. Watch: submitter picking them up, `submitted` status?
5. Watch: retriever processing results, `complete` status?
6. Verify: `site_work_items` receiving findings?
7. Verify: `llm_call_log` entries from retriever?

Rollback: `UPDATE llm_batch_agent_config SET batch_enabled = false WHERE agent_type = 'visual-design-auditor';`
Next invocations go back to sync fallback. Already-queued items still get
processed by the retriever, but new items go sync.

### Phase 4: Enable batch by group

```sql
UPDATE llm_batch_agent_config SET batch_enabled = true, updated_at = NOW()
WHERE batch_group = 'audits';
```

### Phase 5 (future): GPU drain mode

1. Configure GPU provider in `llm_batch_config`
2. Set `provider_preference` on relevant agent workflows
3. Spin up GPU, enable drain, watch it process
4. Compare quality to Anthropic results via `llm_call_log`

---

## Future Considerations

- **Suspend/resume orchestrations** — if batch results need to flow back into
  multi-step workflows. Additive to this design.
- **Cost-based routing** — submitter auto-switches to cheaper provider when
  budget threshold hit.
- **Queue priority aging** — old pending items get priority bumped so they
  don't starve behind newer items.
- **Batch result webhooks** — Anthropic may add webhook support, eliminating
  the polling retriever.
- **Admin dashboard widget** — queue depth, drain progress, cost by provider.
  The monitoring queries above feed this directly.
- **Prompt template storage** — the queue stores rendered prompts (correct for
  execution) but loses the template. If you need to re-render with updated
  context, that's a case for suspend/resume or for the callback to re-read
  from the database.

---

## Timeline

- **Session 1:** Migration (3 tables), `QueueLLMBatchAction` with three-gate
  check and sync fallback, `BatchCallbackDispatcher` with first callback
- **Session 2:** Switch `visual-design-auditor` workflow, test sync fallback
  end-to-end, verify queue rows and llm_call_log
- **Session 3:** Switch remaining audit agents, run for verification
- **Session 4:** `AnthropicBatchAdapter`, submitter, retriever
- **Session 5:** Enable batch for one agent, test async end-to-end
- **Session 6:** Enable by group, drain mode + `OllamaAdapter` (when GPU ready)
- **Session 7:** Priority override / escalation, superseded handling
