# Batch LLM Processing — Architecture

## Date: 2026-03-26 (v2: 2026-04-06, v3: 2026-04-06)

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
3. Global and per-provider batch settings (is batching enabled?)
4. Whether the item has been escalated to urgent

---

## The Table

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
    prompt TEXT NOT NULL,
    max_tokens INT DEFAULT 4000,
    temperature NUMERIC(3,2) DEFAULT 0.7,
    request_params JSONB DEFAULT '{}',           -- provider-specific: image dimensions, LoRA,
                                                  -- aspect_ratio, negative_prompt, etc.

    -- Callback (what to do with the result)
    callback_action TEXT NOT NULL,
    callback_config JSONB DEFAULT '{}',

    -- Routing and priority
    batch_group TEXT GENERATED ALWAYS AS (
        provider || ':' || model || ':' || COALESCE(site_id::text, 'no-site')
    ) STORED,
    priority INT DEFAULT 50,                     -- 0=lowest, 100=highest
    urgent BOOLEAN DEFAULT false,                -- escalation flag, see "Priority Override"

    -- State
    status TEXT NOT NULL DEFAULT 'pending',
    -- pending:     waiting to be submitted/processed
    -- submitted:   sent to async API (Anthropic batch), waiting for results
    -- processing:  being processed right now (GPU drain mode)
    -- complete:    result received and callback executed
    -- failed:      callback failed or API error
    -- expired:     abandoned after max retries
    -- superseded:  result arrived but item was already handled via urgent path
    provider_batch_id TEXT,                      -- Anthropic batch_id, or GPU job_id, etc.
    provider_request_id TEXT,                    -- custom_id within the batch

    -- Result
    response_text TEXT,
    response_url TEXT,                           -- for image generation: URL of generated image
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

### Batch Control Table

Global and per-provider settings for the submitter. This is how you turn
batching on/off, switch to GPU drain mode, etc.

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
    ('global', true, 'batch'),
    ('anthropic', true, 'batch'),
    ('gpu', false, 'drain'),
    ('replicate', false, 'batch');
```

---

## Operational Controls

### 1. Turning Batching On/Off Completely

```sql
-- Kill switch: disable all batching globally
-- QueueLLMBatchAction falls back to direct ExecuteLLMPromptAction
UPDATE llm_batch_config SET enabled = false WHERE provider = 'global';

-- Re-enable
UPDATE llm_batch_config SET enabled = true WHERE provider = 'global';

-- Disable just Anthropic batching (GPU drain can still run)
UPDATE llm_batch_config SET enabled = false WHERE provider = 'anthropic';
```

When `global.enabled = false`, `QueueLLMBatchAction` checks this flag before
writing to the queue. If disabled, it calls `ExecuteLLMPromptAction` directly
(sync, full price) and still writes a row to the queue with `status = 'sync_executed'`
for cost tracking. The agent workflow doesn't need to change at all.

### 2. GPU Spin-Up: Drain Mode

You spin up one or more H100 GPUs for a couple of hours, want to blast through
the queue, then shut down.

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

Which items go to GPU vs Anthropic? The agent workflow config determines this:

```json
{
    "action": "queue_llm_batch",
    "config": {
        "provider": "gpu",
        "model": "llama-3.3-70b",
        "prompt_template": "...",
        "callback_action": "write_audit_findings"
    }
}
```

Or, for agents that should use GPU when available and fall back to Anthropic:

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
        "callback_action": "write_audit_findings"
    }
}
```

`QueueLLMBatchAction` checks `llm_batch_config` for each provider in preference
order. First enabled provider wins. This means you can have items auto-route to
GPU when it's up, and fall back to Anthropic batch when it's not — without
changing any workflow definitions.

### 3. Priority Override (Pull from Batch)

A queued item becomes urgent — maybe a dependency appeared, or you want to
preview a site that has pending audit results.

```sql
-- Escalate a specific item
UPDATE llm_batch_queue SET urgent = true, priority = 100
WHERE id = '<item-id>' AND status = 'pending';

-- Escalate all pending items for a site
UPDATE llm_batch_queue SET urgent = true, priority = 100
WHERE site_id = '<site-id>' AND status = 'pending';
```

The submitter has a fast path: before doing normal batch submission, it checks
for urgent pending items and processes them immediately via direct sync API call
(full price, but instant). After the sync call completes and the callback
executes, the row is marked `complete`.

If the item has already been `submitted` to an Anthropic batch (can't cancel
individual items in a batch):

```sql
-- Mark it urgent — submitter will make a parallel sync call
UPDATE llm_batch_queue SET urgent = true WHERE id = '<item-id>';
```

The submitter makes a direct sync call for the same prompt. When the batch
result eventually arrives, the retriever sees `urgent = true` and
`status = 'complete'` (already handled) and marks it `superseded` instead
of executing the callback twice.

For admin convenience:

```sql
-- Helper function (or admin API endpoint)
CREATE OR REPLACE FUNCTION escalate_batch_item(item_id UUID)
RETURNS TEXT AS $$
DECLARE
    current_status TEXT;
BEGIN
    SELECT status INTO current_status FROM llm_batch_queue WHERE id = item_id;

    IF current_status = 'pending' THEN
        UPDATE llm_batch_queue SET urgent = true, priority = 100
        WHERE id = item_id;
        RETURN 'escalated_pending';
    ELSIF current_status = 'submitted' THEN
        UPDATE llm_batch_queue SET urgent = true
        WHERE id = item_id;
        RETURN 'escalated_submitted_will_parallel';
    ELSE
        RETURN 'cannot_escalate_status_' || current_status;
    END IF;
END;
$$ LANGUAGE plpgsql;

-- Escalate all pending work for a site
CREATE OR REPLACE FUNCTION escalate_site_batch(target_site_id UUID)
RETURNS INT AS $$
DECLARE
    affected INT;
BEGIN
    UPDATE llm_batch_queue SET urgent = true, priority = 100
    WHERE site_id = target_site_id
      AND status IN ('pending', 'submitted');
    GET DIAGNOSTICS affected = ROW_COUNT;
    RETURN affected;
END;
$$ LANGUAGE plpgsql;
```

### 4. Multi-Provider Routing

Different requests need different backends. The `provider` + `request_type`
fields control this.

**Text LLM (Anthropic batch):**
```json
{ "provider": "anthropic", "model": "claude-sonnet-4-6", "request_type": "text" }
```

**Text LLM (GPU, Llama):**
```json
{ "provider": "gpu", "model": "llama-3.3-70b", "request_type": "text" }
```

**Image generation (Replicate/Flux):**
```json
{
    "provider": "replicate",
    "model": "flux-dev",
    "request_type": "image",
    "prompt": "Modern veterinary clinic logo, clean lines...",
    "request_params": {
        "width": 1024,
        "height": 1024,
        "num_inference_steps": 50
    }
}
```

**Image generation (local GPU):**
```json
{
    "provider": "gpu",
    "model": "flux-schnell",
    "request_type": "image",
    "prompt": "...",
    "request_params": { "width": 512, "height": 512 }
}
```

The submitter has a **provider adapter** for each provider type. Each adapter
knows how to:
- Format the request for that provider's API
- Submit (batch or sync depending on mode)
- Parse the response back into `response_text` or `response_url`

```go
type ProviderAdapter interface {
    SubmitBatch(ctx context.Context, items []QueueItem, config BatchConfig) (batchID string, err error)
    PollBatch(ctx context.Context, batchID string, config BatchConfig) ([]BatchResult, error)
    SubmitSync(ctx context.Context, item QueueItem, config BatchConfig) (BatchResult, error)
}
```

Initial adapters:
- `AnthropicBatchAdapter` — uses /v1/messages/batches
- `AnthropicSyncAdapter` — uses /v1/messages (for urgent items, fallback)
- `OllamaAdapter` — sync only (drain mode), calls local/GPU Ollama endpoint
- `ReplicateAdapter` — async (create prediction, poll for result)

New providers can be added without changing the queue table or agent workflows.

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

---

## How Agents Use It

Unchanged from v2 for basic text batching:

```json
{
    "action": "queue_llm_batch",
    "config": {
        "prompt_template": "Audit the visual design of this site...\n{{.site_context}}",
        "system_prompt": "You are a visual design auditor...",
        "model": "claude-sonnet-4-6",
        "max_tokens": 4000,
        "callback_action": "write_audit_findings",
        "callback_config": {
            "site_id": "site_record.site_id",
            "audit_source": "visual-design-audit"
        }
    }
}
```

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

## Submitter Logic (revised)

The submitter runs on the interval defined in `llm_batch_config` for each
provider. Pseudocode:

```
for each provider in llm_batch_config where enabled = true:

    # 1. Always handle urgent items first (any provider, sync)
    urgent_items = SELECT FROM llm_batch_queue
                   WHERE urgent = true AND status = 'pending'
                   AND provider = this_provider
                   FOR UPDATE SKIP LOCKED

    for each urgent_item:
        result = adapter.SubmitSync(urgent_item)
        execute_callback(urgent_item, result)
        mark complete

    # 2. Check mode
    if mode = 'drain':
        if drain_until is set AND NOW() > drain_until:
            set enabled = false, continue
        run_drain_workers(provider, drain_concurrency)

    elif mode = 'batch':
        pending = SELECT FROM llm_batch_queue
                  WHERE status = 'pending' AND provider = this_provider
                  ORDER BY batch_group, priority DESC, created_at
                  LIMIT max_requests_per_batch
                  FOR UPDATE SKIP LOCKED

        if len(pending) > 0:
            group by model (API requires same model per batch)
            for each model_group:
                batch_id = adapter.SubmitBatch(model_group)
                update rows: status = 'submitted', provider_batch_id = batch_id
```

---

## Retriever Logic (revised)

```
# 1. Poll async batches (Anthropic, Replicate)
for each provider in llm_batch_config where mode = 'batch':
    outstanding = SELECT DISTINCT provider_batch_id FROM llm_batch_queue
                  WHERE status = 'submitted' AND provider = this_provider

    for each batch_id:
        results = adapter.PollBatch(batch_id)
        if results is nil: continue  # not ready yet

        for each result:
            row = find queue row by provider_request_id
            if row.urgent AND row.status = 'complete':
                mark 'superseded', skip callback
                continue
            write response_text/response_url, tokens
            execute_callback(row, result)
            mark complete (or failed)

# 2. Retry failed callbacks (response_text already stored)
failed = SELECT FROM llm_batch_queue
         WHERE status = 'failed' AND retry_count < 3
         AND completed_at < NOW() - INTERVAL '10 minutes' * retry_count
         FOR UPDATE SKIP LOCKED

for each failed_item:
    re-execute callback
    if success: mark complete
    else: increment retry_count
    if retry_count >= 3: mark expired
```

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
```

---

## What NOT to Batch

Keep these on direct API calls (existing `execute_llm_prompt` path):
- Initial site builds (user is waiting)
- Briefing agent (blocks the pipeline)
- Site classifier (everything depends on it)
- Chief strategist (one call per domain, user is watching)
- Page content writer during initial build
- Any HITL/interactive workflow

The 50% saving applies to ~60-70% of total token spend (audits, maintenance,
content rewrites, batch processing). The remaining 30-40% (builds, classification)
stays synchronous.

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

## Training Data Collection (RAG + LoRA)

`llm_call_log` is the single source for all prompt/response pairs — used for
RAG retrieval and future LoRA fine-tuning. The batch path must feed it the same
way the sync path already does via `LogLLMCall`.

### How it works

**Sync path (existing, unchanged):** `ExecuteLLMPromptAction` calls `LogLLMCall`
immediately after the LLM response. All fields populated including `latency_ms`.

**Batch path (new):** The retriever calls `LogLLMCall` after processing each
completed result. It has access to the queue row which carries all the
traceability fields (`agent_type`, `agent_id`, `step_name`, `orchestration_id`,
`correlation_id`, `model`, `provider`).

**Sync fallback path:** When `global.enabled = false` and `QueueLLMBatchAction`
falls back to `ExecuteLLMPromptAction`, the existing `LogLLMCall` inside
`ExecuteLLMPromptAction` fires as normal. The queue row is also written with
`status = 'sync_executed'` and `response_text` populated, but the authoritative
training record is in `llm_call_log`.

### Fields that need forwarding

`llm_call_log` has columns that `llm_batch_queue` doesn't carry at top level:
`vertical`, `prompt_variant`, `rag_context_used`, `work_item_id`. These must
be captured at queue time and stored in `callback_config` so the retriever can
pass them through to `LogLLMCall`:

```json
{
    "action": "queue_llm_batch",
    "config": {
        "callback_action": "write_audit_findings",
        "callback_config": {
            "site_id": "site_record.site_id",
            "vertical": "veterinary",
            "prompt_variant": "v3_audit",
            "rag_context_used": true,
            "work_item_id": "current_item.id"
        }
    }
}
```

The retriever's `LogLLMCall` call:

```go
LogLLMCall(db, logger, LLMCallLogParams{
    AgentType:       row.AgentType,
    AgentID:         "", // no specific agent instance in batch path
    StepName:        row.StepName,
    OrchestrationID: row.OrchestrationID,
    CorrelationID:   row.CorrelationID,
    Model:           row.Model,
    Provider:        row.Provider,
    PromptTemplate:  "", // not stored separately in queue
    PromptRendered:  row.Prompt,
    ResponseText:    row.ResponseText,
    InputTokens:     row.InputTokens,
    OutputTokens:    row.OutputTokens,
    LatencyMs:       0, // batch doesn't have per-request latency
    Success:         true,
    WorkItemID:      callbackConfig["work_item_id"],
    Vertical:        callbackConfig["vertical"],
    PromptVariant:   callbackConfig["prompt_variant"],
    RAGContextUsed:  callbackConfig["rag_context_used"],
})
```

### Latency field

Anthropic's batch API doesn't report per-request latency. The `latency_ms`
field will be 0 for batch results. For training data extraction this doesn't
matter. For cost/performance dashboards, filter by `latency_ms > 0` to get
sync-only timings, or use `submitted_at → completed_at` from the queue table
for batch turnaround times.

### Two tables, two purposes

| Table | Purpose | Writes to it |
|---|---|---|
| `llm_call_log` | Training data (RAG, LoRA), analytics | `LogLLMCall` from sync path + retriever |
| `llm_batch_queue` | Operational (routing, status, retries, cost) | `QueueLLMBatchAction` + retriever updates |

Don't query `llm_batch_queue` for training data. Don't query `llm_call_log`
for batch operational status. Each table has one job.

---

## Implementation

### Go Code

**QueueLLMBatchAction** (~80 lines, `actions/batch_queue_action.go`):
- Checks `llm_batch_config` global enabled flag
- If disabled: falls back to `ExecuteLLMPromptAction` directly (which calls
  `LogLLMCall` as normal), also writes queue row with `status = 'sync_executed'`
  and `response_text` populated for cost tracking
- If enabled: resolves provider (direct or from preference list), renders
  prompt, writes queue row, returns immediately
- Reuses existing `renderPromptTemplate` and `getAIServiceConfig`

**ProcessBatchResultAction** (~80 lines, `actions/batch_result_action.go`):
- Dispatches to callback action by name (function registry lookup)
- Self-contained: reads all context from the queue row + callback_config
- Handles the `superseded` check for urgent items

**ProviderAdapter interface** + adapters (~150 lines each):
- `AnthropicBatchAdapter`: /v1/messages/batches submit + poll
- `AnthropicSyncAdapter`: /v1/messages for urgent/fallback
- `OllamaAdapter`: sync calls to Ollama endpoint
- `ReplicateAdapter`: create prediction + poll (future)

**Submitter service** (~150 lines, in kafka-scheduler):
- Registered in `scheduled_tasks`, runs every 5 minutes (configurable)
- Iterates providers, handles urgent-first, then mode-specific logic
- Drain mode: spawns worker goroutines, respects concurrency limit

**Retriever service** (~120 lines, in kafka-scheduler):
- Registered in `scheduled_tasks`, runs every 10 minutes
- Polls each provider's outstanding batches
- Calls `LogLLMCall` for each completed result (training data → `llm_call_log`)
- Then executes the callback action
- Handles superseded detection, failure retry

### Changes to Existing Agents

Same as v2 — swap `execute_llm_prompt` for `queue_llm_batch` in workflow config.
No other changes needed. The provider preference / fallback logic is in the
action, not the workflow.

### First Batch-Eligible Agents

1. **visual-design-auditor** — callback writes to `site_work_items`
2. **content-quality-auditor** — callback writes to `site_work_items`
3. **site-review-agent** — callback writes to `site_work_items`
4. **ch-llm-reviewer** — swap API call for queue write

---

## Future Considerations

- **Suspend/resume orchestrations** — if we need batch results to flow back
  into multi-step workflows. Additive to this design.
- **Cost-based routing** — submitter could check current spend and auto-switch
  to cheaper provider when budget threshold is hit.
- **Queue priority aging** — items that have been pending too long get priority
  bumped automatically so they don't starve.
- **Batch result webhooks** — Anthropic may add webhook support, eliminating
  the polling retriever for that provider.
- **Admin dashboard widget** — queue depth, drain progress, cost by provider.
  The monitoring queries above feed this directly.

---

## Timeline

- **Session 1:** Tables (`llm_batch_queue`, `llm_batch_config`), `QueueLLMBatchAction`
  with global on/off and sync fallback
- **Session 2:** `AnthropicBatchAdapter`, submitter, retriever for Anthropic path
- **Session 3:** Wire up first 2 audit agents, test end-to-end
- **Session 4:** Drain mode + `OllamaAdapter` for GPU path
- **Session 5:** Priority override / escalation, superseded handling
- **Session 6:** Image generation adapter (when needed)
