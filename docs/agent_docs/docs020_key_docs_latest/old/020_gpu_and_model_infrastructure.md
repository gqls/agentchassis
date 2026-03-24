# GPU and Model Infrastructure — Architecture Plan

## Date: 2026-03-24

---

## Context

LLM costs on Claude API are ~$120 per 4 domains over 2-3 weeks (19.6M input tokens, 4.2M output), projecting to $15,000-30,000 at 2,000 domains. A significant portion of this spend is triage loops — audit agents finding problems, fix agents addressing them, re-audits creating more work.

Testing on ThunderCompute shows Llama 3.3 70B on GPU (H100, $1.38/hr) produces comparable quality to Claude for classification and content generation. Mistral Small 3 (24B) on CPU is adequate for low-stakes tasks but weaker on classification and design.

The system needs to support multiple AI endpoints (Claude API, GPU Ollama, CPU Ollama) with agents pointed at whichever is appropriate, and handle unavailability of any endpoint gracefully.

---

## Model Quality Assessment (Tested 2026-03-24)

Tests run against vetcomparison.uk classification, content writing, and web design prompts.

### Classification

| Model | Correct? | Reasoning Quality | Score |
|---|---|---|---|
| Claude (reference) | content YES | Deep — affiliate vertical, SEO, listing fees | 9/10 |
| Llama 3.3 70B (H100) | content YES | Adequate — reviews, guides, comparisons | 8/10 |
| Mistral Small 3 (CPU) | tools NO | Surface — latched on "comparison" = "tool" | 5/10 |

### Content Generation (with 16-rule prompt)

| Model | JSON valid? | Rules followed? | CTA quality | Score |
|---|---|---|---|---|
| Claude | Yes | All 16 | "Compare Vets Near You" — specific | 9/10 |
| Llama 70B | Yes | All 16 | "Search for Vets" — specific | 9/10 |
| Mistral 24B | Yes | Broke 7, 12, 14 | "Get Started" — generic | 6/10 |

### Web Design

| Model | Industry-distinctive? | Fonts | All fields? | Score |
|---|---|---|---|---|
| Claude | Yes — forest green, teal, amber | Inter + DM Sans | 8/8 | 9/10 |
| Llama 70B | Yes — sage/olive, cream, grey | Lato + Merriweather | 8/8 | 7/10 |
| Mistral 24B | No — Material Design defaults | Arial + Georgia | 5/8 | 3/10 |

### Recommended Model Assignment

| Agent | Model | Endpoint | Why |
|---|---|---|---|
| chief-strategist | Claude Opus 4.6 | Anthropic API | One call per domain, highest leverage structural decisions |
| webdesign-agent | Claude Sonnet 4.6 | Anthropic API | Design quality gap is significant |
| build-site-planner | Claude Sonnet 4.6 | Anthropic API | Page structure quality matters |
| site-classifier | Llama 3.3 70B | GPU Ollama | Got classification right, close to Claude |
| page-content-writer | Llama 3.3 70B | GPU Ollama | Matched Claude quality with good prompts |
| briefing-agent | Mistral Small 3 | CPU Ollama | Low stakes, structured output |
| Triage/rewrite agents | Llama 3.3 70B | GPU Ollama | Quality matters here — weak fixes create more triage loops |

### Cost Projection at 2,000 Domains

| Component | Cost |
|---|---|
| Claude Opus (planner, 1 call x 2000) | ~$600 |
| Claude Sonnet (design + planner, 2 calls x 2000) | ~$240 |
| GPU rental (content, classification, triage) | ~$70-150 |
| CPU Mistral (embeddings, briefing) | $0 |
| **Total** | **~$910-990** |

vs ~$15,000-30,000 all-Claude. ~95% reduction.

---

## Core Change: Back-to-Triage on AI Unavailability

### The Problem

Currently, when an LLM endpoint is unreachable:

1. GenerateText fails with connection error
2. Retry logic tries 4 times (5s, 15s, 30s, 60s delays) — burns ~2 minutes
3. Workflow fails
4. Work item marked failed, attempt_count incremented
5. Item eventually hits max_attempts and goes to needs_human_review

This treats "endpoint is down" the same as "model produced garbage." An item that was never actually attempted gets penalised as if it failed. When the endpoint comes back, the item may already be exhausted.

### The Fix

Distinguish connection errors from real failures. Connection errors release the item back to triaged without counting an attempt.

**Error categories:**

| Error Type | Examples | What Happens |
|---|---|---|
| Connection unavailable | Connection refused, DNS lookup failed, timeout on connect | Release to triaged, no attempt counted |
| Credits/auth exhausted | 401, 402 | Release to triaged, no attempt counted, log warning |
| Temporary overload | 529, 503, 502 | Retry (existing logic), then release to triaged if all retries fail |
| Model error | 404 model not found | Real failure, count attempt |
| Bad output | 200 but unparseable response | Real failure, count attempt |
| Success | 200 with good response | Normal completion |

### Code Change 1: AIUnavailableError type

New file or add to existing errors in the actions package:

```go
// AIUnavailableError indicates the AI endpoint is unreachable.
// The work item should be released back to triaged, not marked as failed.
type AIUnavailableError struct {
    Provider string
    Model    string
    Endpoint string
    Cause    error
}

func (e *AIUnavailableError) Error() string {
    return fmt.Sprintf("AI endpoint unavailable: provider=%s model=%s endpoint=%s: %v",
        e.Provider, e.Model, e.Endpoint, e.Cause)
}

func isAIUnavailable(err error) bool {
    errStr := err.Error()
    return strings.Contains(errStr, "connection refused") ||
        strings.Contains(errStr, "no such host") ||
        strings.Contains(errStr, "i/o timeout") ||
        strings.Contains(errStr, "connection reset") ||
        strings.Contains(errStr, "dial tcp") ||
        strings.Contains(errStr, "status 401") ||
        strings.Contains(errStr, "status 402") ||
        strings.Contains(errStr, "credit") ||
        strings.Contains(errStr, "requires more system memory")
}
```

### Code Change 2: Fast-fail in ExecuteLLMPromptAction

In ai_actions.go, before the existing retry logic:

```go
// Call the AI service
result, err := aiClient.GenerateText(ctx, renderedPrompt, options)
if err != nil {
    // Fast-fail on connection errors — don't retry, the endpoint is down
    if isAIUnavailable(err) {
        params.Logger.Warn("AI endpoint unavailable — releasing item back to queue",
            zap.String("provider", provider),
            zap.String("model", resolvedModel),
            zap.Error(err))

        // Log the failed call (for monitoring which endpoints are down)
        LogLLMCall(params.DB, params.Logger, LLMCallLogParams{
            AgentType: params.AgentType,
            Model:     modelAlias,
            Provider:  provider,
            LatencyMs: int(time.Since(llmCallStart).Milliseconds()),
            Success:   false,
            ErrorMessage: err.Error(),
        })

        return nil, &AIUnavailableError{
            Provider: provider,
            Model:    resolvedModel,
            Cause:    err,
        }
    }

    // Existing retry logic for transient errors (529, 503, etc.)
    errStr := err.Error()
    // ... existing code ...
}
```

### Code Change 3: Coordinator releases items on AIUnavailableError

In the coordinator's error handling path (where workflow failures result in work item status changes):

```go
// When a workflow fails, check why before marking the work item
func handleWorkflowFailure(ctx context.Context, db *sql.DB, itemID string, err error) {
    if isAIUnavailableFromError(err) {
        // AI endpoint was down — release item, don't count as attempt
        db.ExecContext(ctx, `
            UPDATE site_work_items
            SET status = 'triaged',
                claimed_by = NULL,
                claimed_at = NULL,
                error = $2,
                updated_at = NOW()
            WHERE id = $1
        `, itemID, "AI endpoint unavailable: " + err.Error())

        // Do NOT increment attempt_count
        return
    }

    // Real failure — existing logic
    db.ExecContext(ctx, `
        UPDATE site_work_items
        SET status = 'failed',
            attempt_count = attempt_count + 1,
            error = $2,
            updated_at = NOW()
        WHERE id = $1
    `, itemID, err.Error())
}
```

### What This Gives You

- **Claude goes down**: Items bounce back to triaged, process when Claude recovers. No operator intervention.
- **GPU is off**: Items that point at GPU endpoint fail fast, go back to triaged. No wasted attempts, no false failures.
- **Credits exhausted**: Items go back to triaged with a warning in the error field. Operator sees the warning and adds credits.
- **Any new endpoint added in future**: Same behaviour automatically.

### What This Does NOT Do

- Does not prevent the dispatch loop from repeatedly claiming and releasing GPU items when GPU is off. Each cycle: claim, spawn handler, handler tries LLM, connection refused, back to triaged. Fast (under a second) but noisy.
- Does not prioritise or deprioritise items based on endpoint availability.

These are the problems the GPU scheduling mechanism (discussed below) would solve as an optimisation layer on top.

---

## GPU Scheduling: Options Under Discussion

The back-to-triage change is the foundation — it makes the system safe regardless of which scheduling approach is chosen. The following options are being evaluated for the next session:

### Option A: Priority-Based Deprioritisation

GPU items have priority set to 900 when GPU is off, restored to original priority when GPU is on. Dispatch loop's existing sort-by-priority naturally skips them.

**Pros:** Zero dispatch loop code changes. Uses existing priority mechanism. Gradual (priority 50 = "process if nothing else to do" vs 900 = "never"). Works for any resource constraint.

**Cons:** Need to track original_priority in spec. Batch SQL update on GPU state change.

### Option B: Simple Boolean Flag

A gpu_available flag checked by load_work_items query.

**Pros:** Explicit. Easy to query queue depth.

**Cons:** New filter logic in dispatch. Doesn't generalise to other constraints.

### Option C: Health-Check Auto-Discovery

Periodic job pings ollama-gpu:11434, adjusts priorities or flag automatically.

**Pros:** Fully automatic. Detects crashes.

**Cons:** 60-second detection lag. Needs A or B underneath.

### Option D: Back-to-Triage Only (no scheduling)

Let items fail fast and return to triaged. Accept the noise.

**Pros:** Simplest. No new mechanisms.

**Cons:** Noisy when GPU is off. Unnecessary pod startups.

**Not yet decided.** To be discussed in next session.

---

## Three Endpoints

| Endpoint | URL | What | Availability |
|---|---|---|---|
| Claude API | https://api.anthropic.com/v1/messages | Opus/Sonnet for planning, design | Always (unless credits exhausted) |
| CPU Ollama | http://ollama-adapter.ai-persona-system.svc.cluster.local:11434 | Mistral Small 3 for briefing, embeddings | Always (runs in K8s cluster) |
| GPU Ollama | http://ollama-gpu.ai-persona-system.svc.cluster.local:11434 | Llama 70B for content, classification, triage | Intermittent (ThunderCompute on-demand) |

The ollama-gpu K8s Service is created manually (Endpoints + Service) when a GPU instance is running. When no GPU is running, the Service doesn't exist and DNS lookups fail — triggering back-to-triage.

---

## Open Questions for Next Session

1. **GPU scheduling mechanism** — priority vs flag vs health-check vs combination
2. **Triage drain loop** — audit phase structure to prevent infinite audit/fix/re-audit cycles. Key ideas: frozen audit batches, acceptance criteria on findings, phases that close.
3. **Prompt optimisation per model** — prompts written for Claude may need adjustment for Llama. How to manage without doubling maintenance.
4. **vLLM vs Ollama on GPU** — Ollama serves one request at a time. At scale, vLLM continuous batching gives higher throughput.
5. **Fallback ai_service** — optional fallback_ai_service in step config for graceful degradation. Discussed but not yet decided whether to implement.
6. **ThunderCompute GPU 1 pre-allocation** — 2-GPU instances show GPU 1 with 77GB consumed. Single-GPU instances work fine.

---

## Current Infrastructure State

### Deployed and Working

- Ollama adapter (CPU): 2 replicas, nomic-embed-text + mistral-small3.1, emptyDir storage
- Knowledge base table: pgvector, 1 test row
- LLM call log table: created but not yet populated (needs chassis v1.0.900 deploy)
- Model aliases: claude-sonnet-4-6, claude-opus-4-6 mapped
- Agent definitions backup: agent_definitions_backup_20260322 (107 definitions)

### Not Yet Deployed

- LLM call logging in chassis (needs v1.0.900 rebuild with incremented tag)
- RAG actions (registered but not workflow-tested)
- Migration 083 (model swap/revert functions)
- Back-to-triage error handling (this document)
- GPU batch adapter (designed but scope reduced — may just be K8s Service + manual SQL)

---

## Key Files

| File | Status | What |
|---|---|---|
| ai_actions.go | Patched, needs deploy v1.0.900 | LLM logging + ollama + back-to-triage (to add) |
| llm_call_logger.go | Patched with visibility logging | Fire-and-forget LLM call logger |
| anthropic.go | Patched, needs deploy | Usage token capture |
| ollama.go | Ready | Ollama AI provider |
| rag_actions.go | Ready, not workflow-tested | RAG lookup and index |
| 083_model_swap_and_rollback.sql | Written, not applied | snapshot/swap/revert functions |
| agent_definitions_backup_20260322 | In database | Nuclear revert backup |
| agent_backup_and_swap_reference.md | Written | Operator reference |
| canine_biology_implementation_plan.md | Written | Knowledge base content plan |
| gpu_and_model_infrastructure.md | This document | GPU and model infrastructure plan |

---

## Standing Decisions

1. Back-to-triage on AI unavailability is the foundation. All scheduling layers on top.
2. Agent definitions are the control plane for model routing. Swap ai_service to change where calls go.
3. Snapshots before every swap. Revert is one SQL call per agent.
4. The dispatch loop doesn't change. It loads items, claims them, spawns handlers.
5. Triage quality matters. Weak fix models create more triage loops. Content and triage agents should use the best available model.
6. Three endpoints are permanent infrastructure. GPU endpoint is intermittent but back-to-triage handles this.
