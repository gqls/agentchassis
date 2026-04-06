# Batch LLM Processing — Architecture Options

## Date: 2026-03-26

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
states expire, Kafka messages TTL. We need a pattern that bridges this gap
without rewriting the core orchestration.

---

## Categorising Our LLM Calls

Not everything benefits equally from batching:

### Good batch candidates (non-blocking, can wait hours)
- Improvement loop audits (visual-design-auditor, content-quality-auditor, site-review-agent)
- Blog content planning
- Content rewrites triggered by audits
- Companies House LLM review (already batched in spirit)
- Vet price parsing (already discussed as batch-tolerant)
- Any scheduled/maintenance task

### Poor batch candidates (user-facing or blocking)
- Initial site build (user expects result in minutes)
- Briefing-agent (blocks the build pipeline)
- Site classifier (blocks everything downstream)
- Page content writer during initial build (pages depend on each other)

### Could go either way
- Improvement loop fixes (nobody's watching, but faster = more iterations)
- Asset generation
- Research agent calls

---

## Architecture Options

### Option A: Batch Table + Polling Agent (recommended)

The simplest approach that fits the existing architecture. No changes to the
core orchestration engine.

```
┌─────────────────┐     ┌──────────────────┐     ┌─────────────────┐
│ Agent (auditor)  │     │ llm_batch_queue   │     │ batch-submitter │
│                  │     │ (new table)       │     │ (scheduled task)│
│ Instead of LLM   │────▶│ Collects requests │────▶│ Every 5 min:    │
│ call, writes to  │     │ with all context  │     │ Submit to API   │
│ batch queue      │     │                   │     │ Store batch_id  │
└─────────────────┘     └──────────────────┘     └─────────────────┘
                                                          │
                         ┌──────────────────┐             │
                         │ batch-retriever   │◀────────────┘
                         │ (scheduled task)  │
                         │ Every 10 min:     │
                         │ Poll API status   │
                         │ On complete:      │
                         │  Parse results    │
                         │  Write to target  │
                         │  (work items,     │
                         │   collected_data)  │
                         └──────────────────┘
```

**New table:**

```sql
CREATE TABLE llm_batch_queue (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    -- Request
    model TEXT NOT NULL,
    prompt TEXT NOT NULL,
    max_tokens INT DEFAULT 4000,
    system_prompt TEXT,
    -- Context for routing the result back
    site_id UUID,
    agent_type TEXT NOT NULL,
    step_name TEXT,
    orchestration_id TEXT,
    work_item_id UUID,
    callback_action TEXT NOT NULL,    -- what to do with the result
    callback_config JSONB DEFAULT '{}',
    -- Batching
    batch_group TEXT DEFAULT 'default', -- group related requests
    priority INT DEFAULT 50,
    -- State
    status TEXT NOT NULL DEFAULT 'pending',  -- pending, submitted, complete, failed
    anthropic_batch_id TEXT,
    anthropic_request_id TEXT,        -- custom_id within the batch
    -- Result
    response_text TEXT,
    input_tokens INT,
    output_tokens INT,
    error TEXT,
    -- Timestamps
    created_at TIMESTAMPTZ DEFAULT NOW(),
    submitted_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ
);

CREATE INDEX idx_batch_queue_status ON llm_batch_queue(status, created_at);
CREATE INDEX idx_batch_queue_batch ON llm_batch_queue(anthropic_batch_id)
    WHERE anthropic_batch_id IS NOT NULL;
```

**How agents use it:**

Instead of calling `execute_llm_prompt`, batch-eligible agents call a new action
`queue_llm_batch` which writes to the table and immediately completes.

```json
{
    "action": "queue_llm_batch",
    "config": {
        "prompt": "...",
        "model": "claude-sonnet-4-6",
        "callback_action": "write_audit_findings",
        "callback_config": {
            "site_id": "site_record.site_id",
            "audit_source": "visual-design-audit"
        }
    }
}
```

The agent's workflow completes immediately after queuing. No waiting.

**Two scheduled tasks handle the rest:**

1. `batch-submitter` (every 5 minutes): Collects pending requests, groups them,
   submits to Anthropic Batch API, updates status to 'submitted'.

2. `batch-retriever` (every 10 minutes): Polls Anthropic for completed batches,
   parses results, executes the callback_action for each result.

**Pros:**
- No changes to orchestration engine
- Agents don't wait — they fire-and-forget to the queue
- Clear audit trail (every request and result in the table)
- Natural batching (requests accumulate, submitted together)
- Easy to monitor: `SELECT status, COUNT(*) FROM llm_batch_queue GROUP BY status`

**Cons:**
- Results don't flow back into the originating orchestration workflow
- Callback actions need to be self-contained (can't depend on collected_data
  from the original workflow)
- Two new scheduled tasks + two new Go actions

---

### Option B: Suspend/Resume Orchestration

The orchestration engine learns to park workflows and resume them later.

```
Agent starts workflow
  → Step 1: load context
  → Step 2: queue_and_suspend (writes batch request, parks orchestration state)
  → [hours pass]
  → batch-retriever detects completion
  → Sends resume message to orchestration
  → Step 3: process_result (continues with result injected into collected_data)
  → Step 4: write_findings
  → Step 5: complete
```

**New concepts:**
- `status: suspended` on orchestration_states
- `awaiting_batch_id` field on orchestration_states
- Resume mechanism: batch-retriever publishes to the agent's process topic
  with the orchestration_id, injecting the result

**Pros:**
- Results flow back into the original workflow naturally
- Agents can have multi-step post-processing that depends on collected_data
- Most similar to how the system works today (just a very long await)

**Cons:**
- Orchestration states could sit for 24 hours — needs TTL/cleanup changes
- Risk of stale state (what if the site changed while we waited?)
- More complex — touches the core orchestration engine
- Pod that started the workflow is long gone — resume must work on any pod

---

### Option C: Split Workflow Pattern

Each batch-eligible agent becomes two agents: a submitter and a processor.

```
visual-design-audit-submitter:
  load_context → render_prompt → queue_to_batch_table → complete

visual-design-audit-processor:
  (triggered by batch-retriever when results arrive)
  load_result → write_audit_findings → complete
```

**Pros:**
- Clean separation of concerns
- Each half is a simple, fast workflow
- No orchestration engine changes
- Processor has full access to context (can reload from DB)

**Cons:**
- Doubles the number of agent definitions for batch-eligible agents
- The "processor" needs to reconstruct context (site_id, etc.) from the batch queue row
- More moving parts to maintain

---

### Option D: Hybrid — Batch by Default, Sync as Override

Use the batch path for everything, but allow a `sync: true` flag that falls back
to the current direct API call. This lets initial builds use sync (fast) while
improvement loops use batch (cheap).

```json
{
    "action": "execute_llm_prompt",
    "config": {
        "ai_service": { "provider": "anthropic", "model": "claude-sonnet-4-6" },
        "batch_eligible": true
    }
}
```

When `batch_eligible: true`:
- If the agent is in a user-triggered workflow → call API directly (sync)
- If the agent is in a scheduled/maintenance workflow → queue to batch

Detection: check if the root trigger was a scheduled task or a manual build.

**Pros:**
- Minimal config changes — add one flag to existing workflows
- Best of both worlds: fast when needed, cheap when not
- Single code path that branches

**Cons:**
- "Am I in a batch context?" logic can be fragile
- Debugging harder — same agent behaves differently depending on trigger

---

## Recommendation

**Start with Option A (Batch Table + Polling Agent).**

It's the least invasive, doesn't touch the orchestration engine, and gives
immediate 50% savings on the highest-volume non-blocking calls. The table
provides full observability.

Apply it to these agents first:
1. visual-design-auditor (run_visual_llm_audit)
2. content-quality-auditor (run_content_llm_audit)
3. site-review-agent (run_strategic_review)
4. Vet price parsing (already batch-tolerant)

These are all fire-and-forget patterns where the result is written to a table
(site_work_items or similar) rather than returned to a waiting workflow.

**Future evolution:** If we find we need results to flow back into workflows
(e.g. for multi-step post-processing), implement Option B (suspend/resume)
as an upgrade. Option A and B aren't mutually exclusive — batch table handles
the simple cases, suspend/resume handles the complex ones.

---

## Implementation Sketch (Option A)

### Go Actions (2 new)

**queue_llm_batch** (~50 lines):
- Renders the prompt template (reuse existing template logic)
- Writes a row to `llm_batch_queue`
- Returns immediately with `{queued: true, batch_queue_id: "..."}`

**process_batch_result** (~80 lines):
- Called by batch-retriever for each completed result
- Reads callback_action and callback_config from the queue row
- Dispatches to the appropriate action (e.g. WriteAuditFindingsAction)
- Updates queue row to 'complete'

### Go Services (2 new, in kafka-scheduler or standalone)

**batch-submitter** (~100 lines):
- Queries `llm_batch_queue WHERE status = 'pending'`
- Groups by model (Anthropic requires same model per batch)
- Builds JSONL payload per Anthropic Batch API spec
- POST /v1/messages/batches
- Updates rows to 'submitted' with anthropic_batch_id

**batch-retriever** (~80 lines):
- Queries distinct anthropic_batch_ids WHERE status = 'submitted'
- GET /v1/messages/batches/{id} for each
- When processing_status = 'ended', GET results
- Parses each result, calls process_batch_result
- Updates rows to 'complete' or 'failed'

### Prompt Caching Bonus

Audit agents send the same site context (CSS, HTML samples, brief) with every
call. With batch + caching:
- First request in the batch pays cache write (1.25x)
- Subsequent requests for the same site read from cache (0.1x)
- Combined with 50% batch discount: up to 95% savings on input tokens

For a site with 3 audit calls sharing context, input cost goes from 3x to ~0.55x.

### Timeline

The batch infrastructure (table + 2 scheduled tasks + 2 actions) is maybe
2-3 sessions of work. The per-agent changes are small — swapping
`execute_llm_prompt` for `queue_llm_batch` in the workflow config.

---

## What NOT to Batch

Keep these on direct API calls:
- Initial site builds (user is waiting)
- Briefing agent (blocks the pipeline)
- Site classifier (everything depends on it)
- Chief strategist (one call per domain, user is watching)
- Any HITL/interactive workflow

The 50% saving applies to ~60-70% of total token spend (audits, maintenance,
content rewrites, batch processing). The remaining 30-40% (builds, classification)
stays synchronous.
