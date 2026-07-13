# GPU and Model Infrastructure — Architecture Plan (v4)

## Date: 2026-03-25

---

## Context

LLM costs on Claude API are ~$120 per 4 domains over 2-3 weeks (19.6M input tokens, 4.2M output), projecting to $15,000-30,000 at 2,000 domains. A significant portion of this spend was triage loops — audit agents finding problems, fix agents addressing them, re-audits creating more work. The triage drain fix (session 2026-03-25) reduces audit token usage by ~65-70%.

Testing on ThunderCompute shows Llama 3.3 70B on GPU (H100, $1.38/hr) produces comparable quality to Claude for classification and content generation. Mistral Small 3 (24B) on CPU is adequate for low-stakes structured tasks (briefing Q&A, simple extraction).

The system now supports multiple AI endpoints (Claude API, GPU Ollama, CPU Ollama) with agents pointed at whichever is appropriate, and handles unavailability of any endpoint gracefully via the endpoint health table and back-to-triage error handling.

---

## What's Deployed (as of 2026-03-25)

### Tables and Functions

| Item | Migration | Status |
|---|---|---|
| `ai_endpoint_health` table (3 seed rows) | 085 | Applied, verified |
| `ai_endpoint_status` view | 085 | Applied |
| `update_endpoint_health()` function | 085 | Applied |
| `snapshot_agent()` function | 083 | Applied |
| `swap_agent_model()` function | 083 | Applied |
| `revert_agent()` function | 083 | Applied |
| `agent_snapshots` view | 083 | Applied |
| `llm_call_log` flywheel columns | 085 | Applied |
| `endpoint-health-checker` agent definition | 085 | Applied |
| `ai-endpoint-health-check` scheduled task | 085 | Applied |

### Go Code (deployed in chassis)

| File | What |
|---|---|
| `write_audit_findings_action.go` | Captures `current_value`, `acceptance_test`, `max_fix_attempts` into work item spec |
| `ai_errors.go` | `AIUnavailableError` type + `isAIUnavailable()` helper |
| `check_endpoint_health_action.go` | Pings Ollama (GET /api/tags) and Claude (1-token haiku) |
| `llm_call_logger.go` | Fire-and-forget logging to `llm_call_log` |
| `ollama.go` | Ollama AI provider |

### Go Patches Written (for next chassis deploy)

| File | What |
|---|---|
| `ai_actions_patch.go` | Fast-fail on `isAIUnavailable` before retry loop, reactive health table update |
| `fail_work_item_patch.go` | Release to triaged without counting attempt on AI unavailability |
| `claim_work_item_patch.go` | Check endpoint health before claiming + `extractAIEndpointFromHandler()` |

### Endpoint Health State

| Name | URL | Status | Check Interval |
|---|---|---|---|
| claude | `https://api.anthropic.com/v1/messages` | UP | 3600s |
| cpu-ollama | `http://ollama-adapter.ai-persona-system.svc.cluster.local:11434` | UP | 60s |
| gpu-ollama | `http://ollama-gpu.ai-persona-system.svc.cluster.local:11434` | DOWN (never started) | 30s |

Active pinging starts after the three Go patches are deployed. Currently seeded values only.

### LLM Call Logging

Working. 9 calls logged successfully on first verification. Known issue: `agent_type` column is empty — `params.AgentType` not passed through to `LogLLMCall`. Low priority fix.

---

## Decisions Made

1. **Endpoint health table is the GPU scheduler.** No separate batch mechanism needed. GPU is either healthy (items flow) or unhealthy (items wait).
2. **Back-to-triage on AI unavailability** is the reactive safety net beneath the health table.
3. **No fallback model chains for now.** Items wait for their configured endpoint. Quality over speed.
4. **Priority means importance only.** Not infrastructure availability or processing tier.
5. **Items don't know about models.** Agent definitions know about endpoints. The health table knows about availability.
6. **Claude health is dual-mode.** Reactive (402/401 marks unhealthy) + active (hourly haiku ping ~$0.002/month auto-recovers).
7. **Agent definitions are the control plane for model routing.** `swap_agent_model()` to change, `revert_agent()` to undo.
8. **Audit findings carry acceptance criteria.** Specific, testable criteria enable cheap verification and bounded fix attempts.
9. **Section locking is the termination condition.** Good-enough sections lock and stop consuming tokens.
10. **Three improvement channels operate independently.** RAG, LoRA, prompt evolution — each valuable alone, compound together.

---

## Model Quality Assessment (Tested 2026-03-24)

Tests run against vetcomparison.uk — classification, content writing, and web design prompts.

| Task | Claude | Llama 3.3 70B (H100) | Mistral Small 3 (CPU) |
|---|---|---|---|
| Classification | 9/10 | 8/10 | 5/10 |
| Content (16 rules) | 9/10 | 9/10 | 6/10 |
| Web Design | 9/10 | 7/10 | 3/10 |

### Recommended Model Assignment

| Agent | Model | Endpoint | Why |
|---|---|---|---|
| chief-strategist | Claude Opus | Anthropic API | One call per domain, worth the cost |
| site-classifier | Claude Sonnet | Anthropic API | Needs strong reasoning (consider Llama 70B on GPU later) |
| page-content-writer | Claude Sonnet | Anthropic API | Highest token volume — swap to Llama 70B on GPU when quality confirmed |
| webdesign-agent | Claude Sonnet | Anthropic API | Design quality matters — test Llama 70B |
| briefing-agent | Mistral Small 3 | CPU Ollama | Lowest risk — structured Q&A extraction. Swap ready. |
| visual-design-auditor | Claude Sonnet | Anthropic API | Needs design judgement |
| content-quality-auditor | Claude Sonnet | Anthropic API | Needs content understanding |

---

## Model Swap Procedure

### Using swap_agent_model() (preferred)

```sql
-- Check current config
SELECT type,
    default_config->'workflow'->'steps'->'STEP_NAME'->'config'->'ai_service' as ai_service
FROM agent_definitions
WHERE type = 'AGENT_TYPE' AND is_active = true;

-- Swap (snapshots automatically)
SELECT swap_agent_model(
    'AGENT_TYPE',
    'STEP_NAME',
    '{"provider": "ollama", "model": "mistral-small3.1",
      "api_url": "http://ollama-adapter.ai-persona-system.svc.cluster.local:11434",
      "max_tokens": 1500}'::jsonb
);

-- Verify
SELECT * FROM agent_snapshots WHERE type = 'AGENT_TYPE';

-- Revert if needed
SELECT revert_agent('AGENT_TYPE');
```

### Common Agent Types and Their LLM Steps

| Agent Type | LLM Step Name | Current Model |
|---|---|---|
| briefing-agent | infer_via_llm | claude-haiku-4-5 |
| site-classifier | classify_site | claude-sonnet-4-6 |
| chief-strategist | generate_build_plan | claude-opus-4-6 |
| page-content-writer | generate_content | claude-sonnet-4-6 |
| webdesign-agent | analyze_design | claude-sonnet-4-5 |
| visual-design-auditor | run_visual_llm_audit | claude-sonnet-4-6 |
| content-quality-auditor | run_content_llm_audit | claude-sonnet-4-6 |
| site-review-agent | run_strategic_review | claude-sonnet-4-6 |

### Evaluating a Swap

```sql
-- After swapping, trigger a build then check:
SELECT agent_type, model, provider, success, latency_ms
FROM llm_call_log
WHERE created_at > NOW() - INTERVAL '30 minutes'
ORDER BY created_at DESC LIMIT 20;
```

---

## Health Check Architecture

### Active Checks (scheduler-driven)

The `endpoint-health-checker` agent runs every 30 seconds via `ai-endpoint-health-check` scheduled task:
- Ollama endpoints: `GET {url}/api/tags` — 200 = healthy
- Claude: 1-token haiku request — 200 = healthy, 402 = credits exhausted, 401 = auth failed

### Reactive Checks (failure-driven)

When an LLM call fails with a connection or credit error (detected by `isAIUnavailable()`):
1. The call returns `AIUnavailableError`
2. `ai_actions.go` reactively calls `update_endpoint_health(url, false, error)`
3. `FailWorkItemAction` detects AI unavailability in the error message and releases the item to triaged without incrementing `attempt_count`

### Dispatch Integration

Before claiming a work item, `ClaimWorkItemAction`:
1. Looks up the handler agent's definition
2. Finds the first `ai_service` config block
3. Extracts the endpoint URL
4. Checks `ai_endpoint_health` — if unhealthy, releases the claim

---

## ThunderCompute Notes

- Single H100 ($1.38/hr) works for Llama 3.3 70B with 8K context (via Modelfile `PARAMETER num_ctx 8192`)
- 2-GPU instances consistently show GPU 1 with 77GB pre-allocated — ThunderCompute platform issue. Use single-GPU.
- Must use Ollama template for pre-configured CUDA drivers
- `OLLAMA_KEEP_ALIVE=-1` prevents model unload. `OLLAMA_FLASH_ATTENTION=1` for memory efficiency.
- Default `num_ctx=262144` is a bug in Ollama model metadata — always override via Modelfile.
- All ThunderCompute instances are currently stopped.

---

## Implementation Status

### Done

- [x] Ollama adapter on CPU cluster (2 replicas, mistral-small3.1 + nomic-embed-text)
- [x] LLM call logging in chassis (deployed, populating)
- [x] ai_endpoint_health table + seed data
- [x] Model swap functions (snapshot/swap/revert)
- [x] Structured audit findings with acceptance criteria
- [x] Audit pass cap (3 passes per site)
- [x] Section locking exclusion in auditor queries
- [x] Health check action code (check_endpoint_health_action.go)
- [x] AIUnavailableError type (ai_errors.go)
- [x] Flywheel columns on llm_call_log

### Next Deploy

- [ ] Back-to-triage fast-fail in ai_actions.go
- [ ] Health check before claim in claim_work_item_action.go
- [ ] Release on AI unavailability in FailWorkItemAction
- [ ] Active health pinging starts working end-to-end

### Ready to Execute (SQL)

- [ ] Briefing-agent swap to Mistral Small 3 (swap_briefing_agent_to_mistral.sql)

### Future

- [ ] Fix empty agent_type in llm_call_log
- [ ] Wire work_item_id through orchestration context to LogLLMCall
- [ ] Fix verification step (cheap LLM call checking acceptance_test)
- [ ] Swap site-classifier to Llama 70B on GPU
- [ ] RAG actions (registered, not workflow-tested)
- [ ] LoRA training pipeline
- [ ] Training data export from llm_call_log
