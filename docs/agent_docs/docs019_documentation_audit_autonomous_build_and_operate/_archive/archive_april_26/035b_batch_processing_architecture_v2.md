# Batch LLM Processing — Architecture

## Date: 2026-03-26 (v2: 2026-04-06)

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

## Architecture: Batch Table + Polling Agent

The simplest approach that fits the existing architecture. No changes to the
core orchestration engine. Agents fire-and-forget to the queue. The orchestration
completes immediately — no suspended states, no sweeper conflicts, no stale
collected_data after 24 hours.

```
┌─────────────────┐     ┌──────────────────┐     ┌─────────────────┐
│ Agent (auditor)  │     │ llm_batch_queue   │     │ batch-submitter │
│                  │     │ (new table)       │     │ (scheduled task)│
│ Instead of LLM   │────▶│ Collects requests │────▶│ Every 5 min:    │
│ call, writes to  │     │ with all context  │     │ Group by model  │
│ batch queue      │     │                   │     │ + site_id       │
└─────────────────┘     └──────────────────┘     │ Submit to API   │
                                                  └─────────────────┘
                                                          │
                         ┌──────────────────┐             │
                         │ batch-retriever   │◀────────────┘
                         │ (scheduled task)  │
                         │ Every 10 min:     │
                         │ Poll API status   │
                         │ On complete:      │
                         │  Parse results    │
                         │  Execute callback │
                         │  Mark complete    │
                         │  (or failed+retry)│
                         └──────────────────┘
```

### The Table

```sql
CREATE TABLE llm_batch_queue (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    -- Traceability (so results can be followed through the log chain)
    client_id TEXT NOT NULL,
    correlation_id TEXT,
    site_id UUID,
    agent_type TEXT NOT NULL,
    step_name TEXT,
    orchestration_id TEXT,
    work_item_id UUID,

    -- Request
    model TEXT NOT NULL,
    system_prompt TEXT,
    prompt TEXT NOT NULL,
    max_tokens INT DEFAULT 4000,
    temperature NUMERIC(3,2) DEFAULT 0.7,

    -- Callback (what to do with the result)
    callback_action TEXT NOT NULL,
    callback_config JSONB DEFAULT '{}',

    -- Batching strategy
    -- Group by (model, site_id) to maximise prompt cache hits:
    -- audit requests for the same site share context → cache reads at 0.1x
    batch_group TEXT GENERATED ALWAYS AS (
        model || ':' || COALESCE(site_id::text, 'no-site')
    ) STORED,
    priority INT DEFAULT 50,

    -- State
    status TEXT NOT NULL DEFAULT 'pending',
    -- pending:   waiting to be submitted
    -- submitted: sent to Anthropic, waiting for results
    -- complete:  result received and callback executed
    -- failed:    callback failed or API error
    -- expired:   abandoned after max retries
    anthropic_batch_id TEXT,
    anthropic_request_id TEXT,

    -- Result
    response_text TEXT,
    input_tokens INT,
    output_tokens INT,
    error_message TEXT,
    retry_count INT DEFAULT 0,

    -- Timestamps
    created_at TIMESTAMPTZ DEFAULT NOW(),
    submitted_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ
);

CREATE INDEX idx_batch_queue_pending ON llm_batch_queue(status, priority DESC, created_at)
    WHERE status = 'pending';
CREATE INDEX idx_batch_queue_submitted ON llm_batch_queue(anthropic_batch_id)
    WHERE status = 'submitted';
CREATE INDEX idx_batch_queue_failed ON llm_batch_queue(status, retry_count)
    WHERE status = 'failed';
CREATE INDEX idx_batch_queue_site ON llm_batch_queue(site_id, status);
CREATE INDEX idx_batch_queue_client ON llm_batch_queue(client_id, status);
```

### How Agents Use It

Instead of calling `execute_llm_prompt`, batch-eligible agents call a new action
`queue_llm_batch` which writes to the table and immediately completes.

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

The agent's workflow completes immediately after queuing. No waiting, no
suspended orchestration states, no sweeper conflicts.

### Submitter (scheduled task, every 5 minutes)

```
1. SELECT * FROM llm_batch_queue WHERE status = 'pending'
   ORDER BY priority DESC, created_at
   LIMIT 10000
   FOR UPDATE SKIP LOCKED

2. Group rows by batch_group (model + site_id)
   — This ensures same-site requests are adjacent in the batch,
     maximising prompt cache hits

3. For each group, cap at 10,000 requests per batch (Anthropic limit)

4. Build JSONL payload per Anthropic Batch API spec:
   Each request uses the queue row id as custom_id

5. POST /v1/messages/batches
   — One batch per (model) since Anthropic requires same model per batch
   — Respect concurrent batch limits (check API docs for current limit)

6. UPDATE rows SET status = 'submitted',
   anthropic_batch_id = <returned_id>,
   submitted_at = NOW()
```

Within each batch, order requests so same-site requests are adjacent. The first
request for a site pays the cache write cost (1.25x input), subsequent requests
for the same site read from cache (0.1x input). Combined with 50% batch discount
this yields up to 95% savings on input tokens for audit batches.

### Retriever (scheduled task, every 10 minutes)

```
1. SELECT DISTINCT anthropic_batch_id FROM llm_batch_queue
   WHERE status = 'submitted'

2. For each batch_id:
   GET /v1/messages/batches/{batch_id}

3. If processing_status = 'ended':
   GET /v1/messages/batches/{batch_id}/results
   Stream JSONL response

4. For each result line:
   a. Match to queue row via custom_id (= queue row id)
   b. Extract response text, token counts
   c. Execute callback_action with callback_config + response_text
   d. If callback succeeds:
        UPDATE SET status = 'complete', response_text = ...,
        input_tokens = ..., output_tokens = ..., completed_at = NOW()
   e. If callback fails:
        UPDATE SET status = 'failed', error_message = ...,
        retry_count = retry_count + 1
      Continue processing remaining results in the batch.
      Don't abort the whole batch for one failure.

5. If processing_status = 'failed' or 'expired':
   UPDATE all rows for that batch SET status = 'failed',
   error_message = 'batch-level failure: ' || processing_status
```

### Failure and Retry

Failed individual items (callback execution failed) are retried by a sweep
within the retriever:

```sql
-- Retry failed items up to 3 times, with backoff
SELECT * FROM llm_batch_queue
WHERE status = 'failed'
  AND retry_count < 3
  AND completed_at < NOW() - INTERVAL '10 minutes' * retry_count
FOR UPDATE SKIP LOCKED
```

For these, the retriever re-executes the callback (the response_text is already
stored, no need to re-submit to Anthropic). After 3 failures, status changes to
'expired' and surfaces in monitoring.

Items where the Anthropic API itself returned an error (rate limit, content
filter, etc.) are requeued as 'pending' for re-submission in the next batch,
up to 3 times.

### Monitoring

```sql
-- Dashboard query
SELECT status, COUNT(*), AVG(output_tokens) as avg_tokens,
       AVG(EXTRACT(EPOCH FROM completed_at - created_at)) as avg_latency_seconds
FROM llm_batch_queue
WHERE created_at > NOW() - INTERVAL '24 hours'
GROUP BY status;

-- Stuck batches (submitted >6 hours ago, not complete)
SELECT anthropic_batch_id, COUNT(*), MIN(submitted_at)
FROM llm_batch_queue
WHERE status = 'submitted'
  AND submitted_at < NOW() - INTERVAL '6 hours'
GROUP BY anthropic_batch_id;

-- Failed items needing attention
SELECT id, agent_type, callback_action, error_message, retry_count
FROM llm_batch_queue
WHERE status = 'failed' AND retry_count >= 3;

-- Cost tracking per client
SELECT client_id,
       SUM(input_tokens) as total_input,
       SUM(output_tokens) as total_output,
       COUNT(*) as total_requests
FROM llm_batch_queue
WHERE status = 'complete'
  AND completed_at > NOW() - INTERVAL '7 days'
GROUP BY client_id;
```

---

## Implementation

### Go Actions (2 new, in existing actions package)

**QueueLLMBatchAction** (~60 lines):
- Renders the prompt template (reuse existing `renderPromptTemplate` logic)
- Extracts client_id, correlation_id, site_id from execution context / collected_data
- Writes a row to `llm_batch_queue`
- Logs via existing `LogLLMCall` with a `batch_queued` flag
- Returns immediately with `{queued: true, batch_queue_id: "..."}`

**ProcessBatchResultAction** (~80 lines):
- Called by batch-retriever for each completed result
- Reads callback_action and callback_config from the queue row
- Resolves site_id, work_item_id from the row (self-contained, doesn't need
  collected_data from the original workflow)
- Dispatches to the appropriate action function (e.g. WriteAuditFindingsAction)
- Updates queue row status

### Scheduled Tasks (2 new, in kafka-scheduler)

**batch-submitter** (~120 lines):
- Registered in `scheduled_tasks` table, runs every 5 minutes
- Uses `FOR UPDATE SKIP LOCKED` so multiple scheduler pods don't double-submit
- Groups by model, orders by batch_group within each batch for cache adjacency
- Calls Anthropic Batch API via existing HTTP client patterns
- Caps at 10,000 requests per API call

**batch-retriever** (~100 lines):
- Registered in `scheduled_tasks` table, runs every 10 minutes
- Polls Anthropic for each outstanding batch_id
- Streams JSONL results, processes each individually
- Failed callbacks don't block other results
- Retries failed callbacks on subsequent runs

### Changes to Existing Agents

For each batch-eligible agent, the workflow change is minimal — swap one step:

Before:
```json
"run_audit": {
    "action": "execute_llm_prompt",
    "config": { "ai_service": {...}, "prompt_template": "..." },
    "next_step": "write_findings"
}
```

After:
```json
"run_audit": {
    "action": "queue_llm_batch",
    "config": {
        "prompt_template": "...",
        "model": "claude-sonnet-4-6",
        "callback_action": "write_audit_findings",
        "callback_config": { "site_id": "site_record.site_id" }
    },
    "next_step": "complete"
}
```

The `write_findings` step moves into the callback — it's executed by the
retriever when results arrive, not by the original agent workflow.

### First Batch-Eligible Agents

1. **visual-design-auditor** — `run_visual_llm_audit` → callback writes to `site_work_items`
2. **content-quality-auditor** — `run_content_llm_audit` → callback writes to `site_work_items`
3. **site-review-agent** — `run_strategic_review` → callback writes to `site_work_items`
4. **ch-llm-reviewer** — already batch-oriented, swap API call for queue write

These are all fire-and-forget patterns where the result is written to a table
rather than returned to a waiting workflow.

---

## Prompt Caching Strategy

Audit agents send the same site context (CSS, HTML samples, brief) with every
call. With batch + caching:
- First request in the batch pays cache write (1.25x input)
- Subsequent requests for the same site read from cache (0.1x input)
- Combined with 50% batch discount: up to 95% savings on input tokens

For a site with 3 audit calls sharing context, input cost goes from 3x to ~0.55x.

To maximise this: the submitter groups by `batch_group` (model + site_id), so
all requests for the same site are adjacent in the JSONL payload. System prompts
containing site context should be identical across audit types for the same site
— use a shared `build_site_audit_context()` helper that both visual and content
auditors call.

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

## Future Evolution

If we find we need batch results to flow back into multi-step workflows
(e.g. audit → fix → re-audit as a single orchestration), we can add a
**suspend/resume** capability to the orchestration engine:

- New status: `SUSPENDED` on orchestration_states (excluded from sweeper)
- `awaiting_batch_id` field
- Retriever publishes a response message to resume the orchestration
- Requires TTL/cleanup changes and stale-state handling

This is additive — the batch table stays, suspend/resume is an alternative
consumption pattern for cases where fire-and-forget isn't sufficient. Build it
when there's a concrete need, not preemptively.

---

## Timeline

- **Session 1:** Table creation, `QueueLLMBatchAction`, basic submitter
- **Session 2:** Retriever, `ProcessBatchResultAction`, retry logic
- **Session 3:** Wire up first 2 agents, monitoring queries, test end-to-end
- 