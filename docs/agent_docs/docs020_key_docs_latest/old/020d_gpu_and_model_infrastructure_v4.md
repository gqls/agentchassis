# GPU and Model Infrastructure — Architecture Plan

## Date: 2026-03-24

---

## Context

LLM costs on Claude API are ~$120 per 4 domains over 2-3 weeks (19.6M input tokens, 4.2M output), projecting to $15,000-30,000 at 2,000 domains. A significant portion of this spend is triage loops — audit agents finding problems, fix agents addressing them, re-audits creating more work.

Testing on ThunderCompute shows Llama 3.3 70B on GPU (H100, $1.38/hr) produces comparable quality to Claude for classification and content generation. Mistral Small 3 (24B) on CPU is adequate for low-stakes tasks but weaker on classification and design.

The system needs to support multiple AI endpoints (Claude API, GPU Ollama, CPU Ollama) with agents pointed at whichever is appropriate, and handle unavailability of any endpoint gracefully — without special-casing any provider.

---

## Decisions Made

1. **Endpoint health table** is the mechanism for GPU scheduling and all AI availability management.
2. **Back-to-triage on AI unavailability** is the reactive safety net beneath the health table.
3. **No fallback model chains for now.** Items wait for their configured endpoint. Keep things simple, keep quality high.
4. **No priority overloading.** Priority means importance, not infrastructure availability.
5. **No processing tiers on work items.** Items don't know about GPU. Agent definitions know about endpoints. The health table knows about availability.
6. **Claude gets a cheap hourly health check.** A single haiku token (~$0.000003 per check, $0.002/month) confirms credits are valid. Immediate detection when credits run out (reactive, from real call failures). Automatic recovery within an hour when credits are topped up (active, from hourly ping). No operator intervention needed for credit recovery.
7. **GPU scheduling is just the health table.** Start ThunderCompute → health check notices within 30 seconds → items flow. Stop ThunderCompute → health check notices → items wait. No separate scheduling mechanism needed.

---

## Model Quality Assessment (Tested 2026-03-24)

Tests run against vetcomparison.uk — classification, content writing, and web design prompts.

### Classification

| Model | Correct? | Reasoning Quality | Score |
|---|---|---|---|
| Claude (reference) | content ✓ | Deep — affiliate vertical, SEO, listing fees | 9/10 |
| Llama 3.3 70B (H100) | content ✓ | Adequate — reviews, guides, comparisons | 8/10 |
| Mistral Small 3 (CPU) | tools ✗ | Surface — latched on "comparison" = "tool" | 5/10 |

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
| chief-strategist | Claude Opus 4.6 | Claude API | One call per domain, highest leverage structural decisions |
| webdesign-agent | Claude Sonnet 4.6 | Claude API | Design quality gap is significant |
| build-site-planner | Claude Sonnet 4.6 | Claude API | Page structure quality matters |
| site-classifier | Llama 3.3 70B | GPU Ollama | Got classification right, close to Claude |
| page-content-writer | Llama 3.3 70B | GPU Ollama | Matched Claude quality with good prompts |
| briefing-agent | Mistral Small 3 | CPU Ollama | Low stakes, structured output |
| Triage/rewrite agents | Llama 3.3 70B | GPU Ollama | Quality matters — weak fixes create more triage loops |

### Cost Projection at 2,000 Domains

| Component | Cost |
|---|---|
| Claude Opus (planner, 1 call × 2000) | ~$600 |
| Claude Sonnet (design + planner, 2 calls × 2000) | ~$240 |
| GPU rental (content, classification, triage) | ~$70-150 |
| CPU Mistral (embeddings, briefing) | $0 |
| **Total** | **~$910-990** |

vs ~$15,000-30,000 all-Claude. ~95% reduction.

---

## Architecture: Three Layers

### Layer 1: Endpoint Health Table (Proactive)

A table tracking which AI endpoints are available right now. The dispatch loop checks this before claiming items. If an item's handler depends on an unhealthy endpoint, the item is skipped — stays `triaged` at its real priority, untouched.

```sql
CREATE TABLE ai_endpoint_health (
    endpoint_url TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    healthy BOOLEAN NOT NULL DEFAULT false,
    last_checked TIMESTAMPTZ,
    last_healthy TIMESTAMPTZ,
    error TEXT,
    check_interval_seconds INT DEFAULT 60,
    check_mode TEXT NOT NULL DEFAULT 'active',
    ping_path TEXT DEFAULT '/api/tags'
);
```

**Check modes:**

| Mode | How It Works | Used For |
|---|---|---|
| `active` | Scheduler pings the endpoint periodically. Free or near-free. | Ollama endpoints (free GET /api/tags), Claude (cheap haiku ping ~$0.000003/check), future cluster endpoints |
| `reactive` | Health updated by real call failures only. No pinging. | Future paid APIs where no cheap ping exists |

Note: Claude uses active mode with a long interval (3600s = hourly). The "ping" sends a 1-token haiku request costing ~$0.002/month. This means credits being topped up are detected automatically within an hour — no operator intervention needed. The reactive path still fires immediately when a real call hits 402, so credit exhaustion is detected instantly.

**Initial data:**

```sql
INSERT INTO ai_endpoint_health VALUES
(
    'https://api.anthropic.com/v1/messages',
    'claude',
    true,          -- assume healthy until a real call fails
    NOW(), NOW(), NULL,
    3600,          -- check once per hour (~$0.002/month)
    'active',      -- cheap haiku ping + reactive on real call failures
    'claude_ping'  -- special handler: sends 1-token haiku request instead of GET
),
(
    'http://ollama-adapter.ai-persona-system.svc.cluster.local:11434',
    'cpu-ollama',
    true,
    NOW(), NOW(), NULL,
    60,            -- check every 60 seconds
    'active',
    '/api/tags'
),
(
    'http://ollama-gpu.ai-persona-system.svc.cluster.local:11434',
    'gpu-ollama',
    false,         -- unhealthy until GPU is started
    NOW(), NULL, NULL,
    30,            -- check every 30 seconds (GPU state changes often)
    'active',
    '/api/tags'
);
```

**Health checker in kafka-scheduler (periodic task, ~40 lines of Go):**

```go
func checkEndpointHealth(db *sql.DB, anthropicAPIKey string) {
    rows, _ := db.Query(`
        SELECT endpoint_url, name, ping_path 
        FROM ai_endpoint_health 
        WHERE check_mode = 'active'
          AND check_interval_seconds > 0
          AND (last_checked IS NULL 
               OR last_checked < NOW() - (check_interval_seconds || ' seconds')::interval)
    `)
    defer rows.Close()
    
    for rows.Next() {
        var url, name, pingPath string
        rows.Scan(&url, &name, &pingPath)
        
        var healthy bool
        var errMsg string
        
        if pingPath == "claude_ping" {
            // Special handler: cheap 1-token haiku request (~$0.000003)
            healthy, errMsg = pingClaude(url, anthropicAPIKey)
        } else {
            // Standard: GET the ping path (free for Ollama, internal services)
            healthy, errMsg = pingEndpoint(url + pingPath)
        }
        
        if healthy {
            db.Exec(`
                UPDATE ai_endpoint_health 
                SET healthy = true, last_checked = NOW(), last_healthy = NOW(), error = NULL
                WHERE endpoint_url = $1
            `, url)
        } else {
            db.Exec(`
                UPDATE ai_endpoint_health 
                SET healthy = false, last_checked = NOW(), error = $1
                WHERE endpoint_url = $2
            `, errMsg, url)
        }
    }
}

func pingEndpoint(url string) (bool, string) {
    client := &http.Client{Timeout: 3 * time.Second}
    resp, err := client.Get(url)
    if err != nil {
        return false, err.Error()
    }
    defer resp.Body.Close()
    return resp.StatusCode == 200, ""
}

// pingClaude sends a 1-token haiku request to verify credits are valid.
// Cost: ~$0.000003 per check. At once/hour = ~$0.002/month.
func pingClaude(baseURL string, apiKey string) (bool, string) {
    body := `{"model":"claude-haiku-4-5-20251001","max_tokens":1,"messages":[{"role":"user","content":"1"}]}`
    req, err := http.NewRequest("POST", baseURL, strings.NewReader(body))
    if err != nil {
        return false, err.Error()
    }
    req.Header.Set("x-api-key", apiKey)
    req.Header.Set("anthropic-version", "2023-06-01")
    req.Header.Set("content-type", "application/json")
    
    client := &http.Client{Timeout: 10 * time.Second}
    resp, err := client.Do(req)
    if err != nil {
        return false, err.Error()
    }
    defer resp.Body.Close()
    
    switch resp.StatusCode {
    case 200:
        return true, ""
    case 402:
        return false, "credits exhausted"
    case 401:
        return false, "authentication failed"
    case 529:
        return true, "" // overloaded but reachable — credits are valid
    default:
        return true, "" // any non-auth error means the API is reachable
    }
}
```

**Dispatch loop integration (in claim_work_item action):**

```go
// After confirming handler agent exists, check its AI endpoint health
aiEndpoint := extractAIEndpointFromConfig(handlerAgentConfig)
if aiEndpoint != "" {
    var healthy bool
    err := db.QueryRowContext(ctx, `
        SELECT healthy FROM ai_endpoint_health WHERE endpoint_url = $1
    `, aiEndpoint).Scan(&healthy)
    
    if err == nil && !healthy {
        logger.Info("Skipping item — AI endpoint unhealthy",
            zap.String("item_id", itemID),
            zap.String("handler", handlerAgent),
            zap.String("endpoint", aiEndpoint))
        
        // Release the claim — item stays triaged
        db.ExecContext(ctx, `
            UPDATE site_work_items 
            SET claimed_by = NULL, claimed_at = NULL, updated_at = NOW()
            WHERE id = $1
        `, itemID)
        
        return map[string]interface{}{
            "claimed": false,
            "reason":  "ai_endpoint_unavailable",
            "endpoint": aiEndpoint,
        }, nil
    }
}
```

**How `extractAIEndpointFromConfig` works:**

Walks the handler's agent definition config, finds the first `ai_service` block, returns the `api_url` (for Ollama) or the base URL (for Anthropic). This is a read from the already-loaded agent definition — no extra DB query.

### Layer 2: Back-to-Triage on AI Unavailability (Reactive Safety Net)

Catches cases the health table missed: endpoint dies between health check and actual use, race conditions, unexpected error types.

**Error categories:**

| Error Type | Examples | What Happens |
|---|---|---|
| Connection unavailable | Connection refused, DNS failed, timeout | Release to triaged, no attempt counted |
| Credits/auth exhausted | 401, 402 | Release to triaged, no attempt counted, **mark endpoint unhealthy** |
| Temporary overload | 529, 503, 502 | Retry (existing logic), then release to triaged |
| Model error | 404 model not found | Real failure, count attempt |
| Bad output | 200 but unparseable | Real failure, count attempt |
| Success | 200 with good response | Normal completion |

**Code change 1: AIUnavailableError type**

```go
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

**Code change 2: Fast-fail in ExecuteLLMPromptAction**

```go
result, err := aiClient.GenerateText(ctx, renderedPrompt, options)
if err != nil {
    if isAIUnavailable(err) {
        params.Logger.Warn("AI endpoint unavailable — releasing item back to queue",
            zap.String("provider", provider),
            zap.String("model", resolvedModel),
            zap.Error(err))

        // Log the failed call for monitoring
        LogLLMCall(params.DB, params.Logger, LLMCallLogParams{
            AgentType:    params.AgentType,
            Model:        modelAlias,
            Provider:     provider,
            LatencyMs:    int(time.Since(llmCallStart).Milliseconds()),
            Success:      false,
            ErrorMessage: err.Error(),
        })

        // Update health table reactively (especially for Claude credit failures)
        updateEndpointHealth(params.DB, apiURL, false, err.Error())

        return nil, &AIUnavailableError{
            Provider: provider,
            Model:    resolvedModel,
            Endpoint: apiURL,
            Cause:    err,
        }
    }

    // Existing retry logic for transient errors (529, 503, etc.)
    // ...
}
```

**Code change 3: Coordinator releases items on AIUnavailableError**

```go
if isAIUnavailableFromError(workflowError) {
    db.ExecContext(ctx, `
        UPDATE site_work_items
        SET status = 'triaged',
            claimed_by = NULL,
            claimed_at = NULL,
            error = $2,
            updated_at = NOW()
        WHERE id = $1
    `, itemID, "AI endpoint unavailable: " + workflowError.Error())
    // Do NOT increment attempt_count
    return
}
```

**The reactive health update for Claude:**

When a Claude call fails with 401/402, `updateEndpointHealth` marks the Claude endpoint as unhealthy immediately. The health table now says `claude: healthy = false`. All subsequent dispatch cycles skip Claude-dependent items.

**Auto-recovery:** The hourly haiku ping detects when credits are topped up. Within an hour of adding credits, the ping succeeds, the health table flips back to `healthy = true`, and items start flowing. No operator SQL needed.

**Manual override (if you can't wait an hour):**
```sql
UPDATE ai_endpoint_health SET healthy = true, error = NULL WHERE name = 'claude';
```

### Layer 3: GPU Lifecycle (Operational, Not Code)

Starting and stopping the GPU is a manual operation. The health table makes it automatic from the system's perspective:

```
Operator starts ThunderCompute H100 instance
  → Creates K8s Service "ollama-gpu" pointing at instance IP
  → Within 30 seconds: scheduler health check pings /api/tags → healthy = true
  → Next dispatch cycle: GPU-dependent items start being claimed
  → Items process in priority order (highest priority first)

Operator stops ThunderCompute instance  
  → Deletes K8s Service (or leaves it — ping will fail either way)
  → Within 30 seconds: scheduler health check → healthy = false
  → In-flight items: back-to-triage catches connection errors
  → Next dispatch cycle: GPU items skipped, stay triaged
  → Other items (Claude, CPU Ollama) continue normally
```

No scheduling mechanism. No batch adapter. No priority manipulation. The health table is the scheduler.

**Operator commands (manual for now):**

```bash
# Start GPU work session
# 1. Start ThunderCompute instance (via dashboard or tnr CLI)
# 2. Get the instance IP
# 3. Create K8s service:
kubectl -n ai-persona-system apply -f - <<EOF
apiVersion: v1
kind: Endpoints
metadata:
  name: ollama-gpu
subsets:
  - addresses:
      - ip: <THUNDER_IP>
    ports:
      - port: 11434
        name: http
---
apiVersion: v1
kind: Service
metadata:
  name: ollama-gpu
spec:
  ports:
    - port: 11434
      targetPort: 11434
      name: http
EOF
# 4. Health check auto-discovers within 30 seconds
# 5. Items start flowing

# Stop GPU work session
kubectl -n ai-persona-system delete svc ollama-gpu
kubectl -n ai-persona-system delete endpoints ollama-gpu
# Health check marks unhealthy within 30 seconds
# Stop ThunderCompute instance via dashboard
```

**Monitoring during GPU session:**

```sql
-- How many items are waiting for GPU?
SELECT COUNT(*) as waiting
FROM site_work_items wi
JOIN agent_definitions ad ON ad.type = wi.handler_agent AND ad.is_active = true
WHERE wi.status = 'triaged'
  AND ad.default_config::text LIKE '%ollama-gpu%';

-- Is the GPU healthy?
SELECT name, healthy, last_checked, last_healthy, error
FROM ai_endpoint_health
WHERE name = 'gpu-ollama';

-- What's being processed right now?
SELECT wi.item_type, wi.status, wi.handler_agent, wi.claimed_at
FROM site_work_items wi
WHERE wi.status = 'claimed'
ORDER BY wi.claimed_at DESC
LIMIT 10;
```

---

## How Everything Flows Together

```
ITEM CREATED BY PLANNER:
  needs_content_page, priority 10, handler: page-content-writer
  (item knows nothing about GPU, models, or endpoints)

DISPATCH LOOP CLAIMS ITEM:
  1. Load triaged items sorted by priority
  2. For each item, try to claim:
     a. Check handler agent exists in agent_definitions → yes
     b. Extract handler's ai_service endpoint URL from config
     c. Check ai_endpoint_health table: is that endpoint healthy?
        → YES: claim the item, spawn handler
        → NO: skip this item, try next one
  3. Item stays triaged at its real priority until endpoint is healthy

HANDLER RUNS:
  1. Workflow executes, reaches execute_llm_prompt step
  2. Calls GenerateText on configured endpoint
     → SUCCESS: normal flow, log to llm_call_log
     → CONNECTION ERROR: fast-fail, return AIUnavailableError
        → Back-to-triage: item released, attempt not counted
        → Health table updated reactively (for paid APIs)
     → REAL FAILURE (bad output): normal failure, attempt counted

GPU OFF → GPU ON:
  Nothing changes in items, agents, or dispatch loop code.
  Health check flips gpu-ollama to healthy.
  Next dispatch cycle picks up GPU items in priority order.

GPU ON → GPU OFF:
  Health check flips gpu-ollama to unhealthy within 30 seconds.
  Items already claimed: back-to-triage catches connection errors.
  Unclaimed items: dispatch loop skips them (endpoint unhealthy).
  Claude and CPU items: completely unaffected.
```

---

## Three Endpoints

| Endpoint | Name | URL | Health Mode | Check Interval | Availability |
|---|---|---|---|---|---|
| Claude API | claude | https://api.anthropic.com/v1/messages | active (cheap haiku ping) | 3600s (hourly) | Always (auto-recovers after credits topped up) |
| CPU Ollama | cpu-ollama | http://ollama-adapter...svc.cluster.local:11434 | active | 60s | Always (K8s cluster) |
| GPU Ollama | gpu-ollama | http://ollama-gpu...svc.cluster.local:11434 | active | 30s | Intermittent (ThunderCompute) |
| Future: other clusters | per-cluster | http://ollama-{cluster}...svc.cluster.local:11434 | active | 60s | Varies |

Adding a new endpoint (new GPU provider, new cluster, fine-tuned model server) is one INSERT into the health table. The scheduler starts pinging it. Agents that point at it start having their items claimed when it's healthy.

---

## Database Changes

```sql
-- Single new table
CREATE TABLE ai_endpoint_health (
    endpoint_url TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    healthy BOOLEAN NOT NULL DEFAULT false,
    last_checked TIMESTAMPTZ,
    last_healthy TIMESTAMPTZ,
    error TEXT,
    check_interval_seconds INT DEFAULT 60,
    check_mode TEXT NOT NULL DEFAULT 'active',
    ping_path TEXT DEFAULT '/api/tags'
);

-- Seed data
INSERT INTO ai_endpoint_health VALUES
('https://api.anthropic.com/v1/messages', 'claude', true, NOW(), NOW(), NULL, 3600, 'active', 'claude_ping'),
('http://ollama-adapter.ai-persona-system.svc.cluster.local:11434', 'cpu-ollama', true, NOW(), NOW(), NULL, 60, 'active', '/api/tags'),
('http://ollama-gpu.ai-persona-system.svc.cluster.local:11434', 'gpu-ollama', false, NOW(), NULL, NULL, 30, 'active', '/api/tags');

-- View for operator
CREATE OR REPLACE VIEW ai_endpoint_status AS
SELECT 
    name,
    healthy,
    CASE WHEN healthy THEN 'UP' ELSE 'DOWN' END as status,
    error,
    last_checked,
    last_healthy,
    CASE 
        WHEN last_healthy IS NULL THEN 'never'
        ELSE age(now(), last_healthy)::text 
    END as down_since,
    check_mode
FROM ai_endpoint_health
ORDER BY name;
```

No changes to site_work_items. No changes to agent_definitions schema. No new columns anywhere except the one new table.

---

## Code Changes

### New Files

| File | What | Size |
|---|---|---|
| `platform/orchestration/actions/ai_errors.go` | AIUnavailableError type + isAIUnavailable helper | ~40 lines |

### Modified Files

| File | Change | Size |
|---|---|---|
| `ai_actions.go` | Add fast-fail on connection error before retry logic. Add reactive health update for paid APIs. | ~20 lines added |
| `claim_work_item_action.go` | After handler existence check, check endpoint health. Skip if unhealthy. | ~15 lines added |
| `kafka-scheduler` (periodic task) | Add endpoint health check task. Ping Ollama endpoints (GET), ping Claude (cheap haiku request). Update table. | ~60 lines |

### Not Changed

- Dispatch loop logic (load_work_items, process_item)
- Work item schema
- Agent definition schema
- Workflow definitions
- Priority handling

---

## Implementation Order

### Step 1: Get LLM Logging Working

- [ ] Rebuild chassis with incremented tag (v1.0.900) to avoid node image cache
- [ ] Deploy and verify llm_call_log populates
- [ ] Prerequisite for measuring everything that follows

### Step 2: Create ai_endpoint_health Table

- [ ] Run migration (the CREATE TABLE + INSERT above)
- [ ] Verify with `SELECT * FROM ai_endpoint_status`

### Step 3: Back-to-Triage Error Handling

- [ ] Add AIUnavailableError type and isAIUnavailable helper
- [ ] Modify ExecuteLLMPromptAction to fast-fail on connection errors
- [ ] Add reactive health table update on credit/auth failures
- [ ] Modify coordinator error handling to release items to triaged on AIUnavailableError
- [ ] Test: point an agent at a non-existent endpoint, verify item goes back to triaged

### Step 4: Health Check in Dispatch Loop

- [ ] Modify claim_work_item to check endpoint health before claiming
- [ ] Add extractAIEndpointFromConfig helper
- [ ] Test: set gpu-ollama to healthy=false, verify GPU items are skipped

### Step 5: Scheduler Health Check Task

- [ ] Add periodic endpoint ping task to kafka-scheduler
- [ ] Test: start/stop an Ollama instance, verify health table updates within 30-60 seconds

### Step 6: Apply Model Swap Infrastructure

- [ ] Run migration 083 (snapshot/swap/revert functions)
- [ ] Take fresh backup before any swaps

### Step 7: Swap First Agent to Local Model

- [ ] Swap briefing-agent to Mistral Small 3 on CPU Ollama
- [ ] Trigger a build, verify it works
- [ ] Check llm_call_log

### Step 8: Set Up GPU Endpoint

- [ ] Start ThunderCompute H100
- [ ] Pull llama3.3:70b, create llama70b Modelfile (num_ctx 8192)
- [ ] Create K8s Endpoints + Service "ollama-gpu"
- [ ] Verify health check detects it within 30 seconds

### Step 9: Swap Content Agents to GPU

- [ ] Swap page-content-writer, site-classifier to Llama 70B on GPU
- [ ] Trigger builds, compare output quality
- [ ] Monitor llm_call_log

---

## Open Questions for Next Session

1. **vLLM vs Ollama on GPU** — at scale with concurrent handlers, vLLM continuous batching gives higher throughput. When does this become the bottleneck?
2. **ThunderCompute GPU 1 issue** — 2-GPU instances consistently show GPU 1 with 77GB consumed. Single-GPU H100 instances work. Raised as platform issue.
3. **Llama 4 Scout/Maverick testing** — MoE architecture with 17B active params might outperform Llama 3.3 70B for our tasks while being faster. Needs testing.
4. **LoRA training automation** — adapter or workflow with actions for data export, training, evaluation, and deployment.

---

## Triage Drain Loop Fix

### The Problem

The current triage pipeline spent the majority of tokens on audit/fix cycles across 4 domains:

| Domain | design-audit items | completeness-discovery items | Total items |
|---|---|---|---|
| ai-agent-orchestration.com | 210 | 147 | 622 |
| finetuning.uk | 253 | 88 | 508 |
| gaswholesalers.com | 191 | 82 | 438 |
| leopardessconsulting.co.uk | 191 | 52 | 314 |

Discovery agents created 845+ items from design-audit alone across 4 domains over ~10 days. Each creates LLM calls for both discovery and fixing. The loop has no termination condition.

### Fix 1: Structured Audit Findings with Acceptance Criteria

The audit agent produces specific, testable findings instead of vague complaints:

```json
{
  "findings": [
    {
      "component": "hero",
      "finding": "Subheadline doesn't differentiate from competitors",
      "current_value": "We help you find the best care for your pet",
      "problem": "Could appear on any pet-related site. Doesn't mention what makes this site different.",
      "suggestions": [
        "Mention the comparison/review functionality",
        "Reference the UK geographic focus",
        "Reference specific data types (costs, services, location)"
      ],
      "acceptance_test": "Subheadline must contain at least TWO of: comparison/reviewing, UK/local, specific data like costs or services",
      "acceptance_levels": {
        "acceptable": "Contains at least ONE specific differentiator",
        "good": "Contains TWO+ differentiators and reads naturally",
        "excellent": "Differentiators, trust signal, and action motivation"
      },
      "minimum_required": "acceptable",
      "severity": "medium"
    }
  ]
}
```

The fixer has a concrete target. Verification checks the acceptance test, not a full re-audit.

### Fix 2: Capped Audit Passes

Maximum 3 audit passes per site. Each produces a numbered batch:

```
Batch 1 (initial audit): top 5 findings → fix → verify against criteria → batch closed
Batch 2 (re-audit): top 5 findings → fix → verify → batch closed
Batch 3 (final): top 3 findings → fix or accept → batch closed
Site → audit_complete. No more automatic audits.
```

The prompt tells the auditor "find the top 3-5 most impactful problems" — not everything.

### Fix 3: Section Locking

Components/sections that pass verification get locked. Locked sections are skipped by subsequent audits:

- **Section lock**: component is good, auditor skips it
- **Page lock**: all sections done, no audit or fix activity
- **Site lock**: entire site done, maintenance only
- **Unlock**: always manual — human decides to reopen

Simplest implementation: `locked_at` timestamp on `page_components`. Audit prompt says "skip locked sections."

### Fix 4: Fix Verification Against Criteria (not Re-Audit)

After fixing, a cheap verification call checks the acceptance test:

```
"The acceptance test is: 'Contains at least one specific differentiator.'
The new text is: 'Compare veterinary practices across the UK by services and fees.'
Does it pass? YES or NO with brief explanation."
```

This is a tiny LLM call — could run on Mistral on CPU. Pass → item complete, lock section. Fail → one more attempt. After 2 attempts → escalate to needs_human_review.

### Fix 5: Per-Page Sequential Processing

Work items for the same page process in order via `depends_on`:

```
fix_hero_content (priority 10, page: index)
  → fix_hero_colors (priority 11, depends_on: fix_hero_content)
    → fix_hero_responsive (priority 12, depends_on: fix_hero_colors)
```

Same-page items are sequential. Different-page items can be parallel. Uses existing `depends_on` mechanism.

### Estimated Token Reduction

```
Current (per domain):   ~8+ audit passes, ~50+ fix items = ~88K+ tokens
Proposed (per domain):  3 audit passes, ~9-15 fix items + 15 verification = ~30K tokens
Reduction: ~65-70%
```

Combined with model tier savings (Llama 70B for fixing, Claude for auditing): ~90% total cost reduction on triage.

---

## Quality Improvement Flywheel

### Architecture

```
KNOWLEDGE COLLECTION
  Scrape top sites per vertical → extract structured examples → quality gate → knowledge_base
  Every successful Claude output → llm_call_log → filtered by audit outcome → knowledge_base
  Audit insights about what works → knowledge_base

        │                │                │
        ▼                ▼                ▼
      RAG            LoRA Training    Prompt Evolution
  Inject relevant    Train on best    A/B test variants
  examples into      examples per     using audit success
  prompts at call    task type        rate as fitness
  time                                function
        │                │                │
        └────────────────┼────────────────┘
                         ▼
               SITE PRODUCTION
         Content writer uses: base model + LoRA + RAG + best prompt variant
                         │
                         ▼
                  QUALITY GATE
         Audit → acceptance criteria → pass → lock → log as good example
                                     → fail → fix (2 max) → escalate or accept
                         │
                         ▼
                  FEEDBACK LOOP
         Good outputs → knowledge collection (training data)
         Prompt variant success rates → prompt evolution
         Audit insights → RAG knowledge base
```

### Three Improvement Channels

**RAG** — improves by accumulating more examples. Scrape more sites, process more domains, store more successful outputs. No retraining needed. See `rag_best_practices.md` for detailed guidance.

**LoRA** — improves by retraining on accumulated successful outputs. Start early — some training is better than none. Process matters more than data volume initially.

**Prompts** — improve through deliberate A/B testing. Create challenger variants with specific hypotheses. Test 10-20 sites each. Compare audit success rates. Promote winners.

### Prompt Evolution (Deliberate, Not Random)

The operator or an LLM creates a small set of candidate prompts with hypotheses:

```json
"prompt_config": {
    "active_variant": "v3_with_examples",
    "testing_variant": "v4_role_based",
    "testing_allocation": 0.2,
    "variants": {
        "v3_with_examples": {"template": "...", "promoted_at": "2026-03-20"},
        "v4_role_based": {"template": "...", "hypothesis": "role framing improves specificity"}
    }
}
```

80% of work goes to the proven variant. 20% to the challenger. After enough data, promote or discard.

### Quality Metrics

**Measurable now (once llm_call_log is deployed):**

| Metric | What It Measures | How |
|---|---|---|
| First-pass success rate | Content that passes audit without rewrite | Join llm_call_log to work item outcomes |
| Rewrite count | Fix cycles before acceptance | Count content_rewrite items per original |
| Lock rate | Sections accepted after pass 1 vs 2 vs 3 | locked_at timing relative to audit batches |
| Escalation rate | Model's failure ceiling | needs_human_review count |
| Token efficiency | Output per successful section | output_tokens from llm_call_log |
| JSON validity rate | Fundamental compliance | Parse success in response handling |

**Measurable later (when sites have traffic):**

| Metric | What It Measures | Feedback Loop |
|---|---|---|
| Bounce rate | Content engagement | High-engagement content becomes premium training data |
| Time on page | Content quality | Long-read pages indicate good content |
| CTA click-through | Conversion effectiveness | Train on CTAs that actually convert |
| Affiliate conversion | Revenue impact | Ultimate quality signal |
| SEO ranking | Search relevance | Content that ranks well trains better content |

### Scraped Data Quality Gate (AI Slop Prevention)

Not all scraped sites are good training data. Quality signals:

- **Domain age >3 years**: likely human-written
- **Specific details**: real names, addresses, dated content = human
- **Custom design**: professional photography, branded elements = investment in content
- **Content personality**: consistent voice across pages = human author
- **AI probability assessment**: LLM scores content authenticity

**HITL option:** With flag on, human reviews quality assessment. With flag off, system uses quality threshold automatically. Everything gets scraped for competitive intel, only quality-approved sites contribute to training set.

### LLM Call Log Additions for Flywheel

```sql
ALTER TABLE llm_call_log
ADD COLUMN IF NOT EXISTS work_item_id UUID,
ADD COLUMN IF NOT EXISTS prompt_variant TEXT DEFAULT 'default',
ADD COLUMN IF NOT EXISTS vertical TEXT,
ADD COLUMN IF NOT EXISTS rag_context_used BOOLEAN DEFAULT false;

CREATE INDEX IF NOT EXISTS idx_llm_log_flywheel
ON llm_call_log(agent_type, prompt_variant, success, vertical)
WHERE success = true;
```

- `work_item_id` — direct link to outcome (did this output survive audit?)
- `prompt_variant` — which version produced this (for A/B testing)
- `vertical` — industry (for filtering training data by vertical)
- `rag_context_used` — whether RAG examples were injected (to measure RAG impact)

### LoRA Training as Adapter/Workflow

Training runs as a workflow with specialised actions:

```
training-orchestrator workflow:
  1. export_training_data (action)
     → Query llm_call_log + work item outcomes
     → Filter for quality (first-pass success, no rewrites)
     → Filter by vertical if specified
     → Format as instruction/response JSONL
     → Write to S3/Backblaze

  2. start_gpu_instance (action)
     → Call ThunderCompute API
     → Wait for instance ready

  3. run_training (action)
     → SSH/API call to GPU instance
     → Execute Unsloth QLoRA training script
     → Monitor progress

  4. evaluate_model (action)
     → Run held-out test set through new LoRA
     → Compare metrics to previous LoRA
     → Produce quality report

  5. deploy_or_reject (action)
     → If quality >= previous: deploy new LoRA to Ollama
     → If quality < previous: discard, log result

  6. stop_gpu_instance (action)
     → Stop ThunderCompute instance

  7. log_training_run (action)
     → Record results to training_runs table
```

The adapter handles ThunderCompute lifecycle. The actions handle data preparation, execution, and evaluation. This keeps the workflow readable and the complexity in Go.

---

## Models to Evaluate

### Currently Tested

| Model | Size | Where | Quality |
|---|---|---|---|
| Claude Opus 4.6 | API | Anthropic | Reference (9/10 all tasks) |
| Claude Sonnet 4.6 | API | Anthropic | Close to Opus for most tasks |
| Llama 3.3 70B (Q4) | 40GB | GPU (H100) | Classification 8/10, content 9/10, design 7/10 |
| Mistral Small 3 (24B) | 15GB | CPU cluster | Classification 5/10, content 6/10, design 3/10 |

### Worth Testing

| Model | Why | Deployment |
|---|---|---|
| Llama 4 Scout (109B total, 17B active) | MoE — only 17B active per token, fits single H100 with int4. 10M context. May match 70B quality with faster inference. | GPU (ThunderCompute) |
| Llama 4 Maverick (400B total, 17B active) | 128 experts, potentially higher quality than Scout. Needs more VRAM. | GPU (may need 2x H100) |
| nomic-embed-text-v2-moe | Upgraded embeddings — same 768 dims, better multilingual, MoE. | CPU cluster (Ollama) |
| BGE-M3 (568M) | Better long-document embeddings if RAG entries get longer. | CPU cluster |

Llama 4 is released under the Llama Community License (not OSI-approved open source, but free for <700M monthly active users — fine for our scale). MoE architecture means faster inference than Llama 3.3 70B despite larger total parameter count.

---

## Current Infrastructure State

### Deployed and Working

- Ollama adapter (CPU): 2 replicas, nomic-embed-text + mistral-small3.1, emptyDir, ephemeral-storage requests
- Knowledge base table: pgvector, 1 test row
- LLM call log table: created but not yet populated (needs v1.0.900 deploy)
- Model aliases: claude-sonnet-4-6, claude-opus-4-6 mapped
- Agent definitions backup: agent_definitions_backup_20260322 (107 definitions)

### Tested but Not Persistent

- Llama 3.3 70B on ThunderCompute H100: classification 8/10, content 9/10, design 7/10
- Mistral Small 3 on CPU cluster: classification 5/10, content 6/10, design 3/10

### Not Yet Deployed

- LLM call logging in chassis (needs v1.0.900)
- RAG actions (registered, not workflow-tested)
- Migration 083 (model swap/revert functions)
- ai_endpoint_health table
- Back-to-triage error handling
- Health check in claim_work_item
- Scheduler health check task
- Structured audit findings with acceptance criteria
- Section locking
- Flywheel llm_call_log columns (work_item_id, prompt_variant, vertical, rag_context_used)
- Training data export pipeline
- Quality-gated scraping extraction

---

## Key Files

| File | Status | What |
|---|---|---|
| ai_actions.go | Patched, needs v1.0.900 deploy + back-to-triage addition | LLM logging + ollama + fast-fail |
| llm_call_logger.go | Patched with visibility logging | Fire-and-forget LLM call logger |
| anthropic.go | Patched, needs deploy | Usage token capture |
| ollama.go | Ready | Ollama AI provider |
| rag_actions.go | Ready, not workflow-tested | RAG lookup and index |
| ai_errors.go | To create | AIUnavailableError type |
| claim_work_item_action.go | To modify | Endpoint health check before claim |
| 083_model_swap_and_rollback.sql | Written, not applied | snapshot/swap/revert functions |
| 084_ai_endpoint_health.sql | To create | Endpoint health table + seed data |
| agent_definitions_backup_20260322 | In database | Nuclear revert backup |
| agent_backup_and_swap_reference.md | Written | Operator reference |
| canine_biology_implementation_plan.md | Written | Knowledge base content plan |
| rag_best_practices.md | Written | RAG quality and retrieval guidance |
| gpu_and_model_infrastructure.md | This document | Architecture plan |

---

## Standing Decisions

1. **Endpoint health table is the GPU scheduler.** No separate batch mechanism needed. GPU is either healthy (items flow) or unhealthy (items wait).
2. **Back-to-triage is the safety net.** Catches anything the health table misses. No item is penalised for infrastructure being down.
3. **No fallback chains for now.** Items wait for their configured endpoint. Quality over speed.
4. **Priority means importance.** Not infrastructure availability. Not processing tier.
5. **Items don't know about models.** Agent definitions know about endpoints. The health table knows about availability. Items just say what they need done.
6. **Claude health is dual-mode.** Reactive: first failed call (402/401) marks it unhealthy immediately. Active: hourly cheap haiku ping (~$0.002/month) auto-recovers when credits are topped up. No operator intervention needed for credit recovery.
7. **Active health checks for free endpoints.** Scheduler pings Ollama endpoints every 30-60 seconds. GPU discovery is automatic.
8. **Agent definitions are the control plane for model routing.** Swap ai_service to change where calls go. Snapshot before swapping. Revert if needed.
9. **Triage quality matters.** Weak fix models create more triage loops. Content and triage agents should use the best available model, not the cheapest.
10. **The dispatch loop doesn't change structurally.** It gains one health check in claim_work_item. Everything else stays the same.
11. **Audit findings must carry acceptance criteria.** Vague findings create expensive fix loops. Specific, testable criteria enable cheap verification and bounded fix attempts.
12. **Section locking is the termination condition.** Good-enough sections lock and stop consuming tokens. Unlock is always manual.
13. **Three improvement channels operate independently.** RAG (more examples), LoRA (better weights), prompts (better instructions). Each is valuable alone. Together they compound.
14. **Start LoRA training early.** Process matters more than data volume. Build the automation, validate the pipeline, improve incrementally.
15. **Quality-gate all training data.** AI slop from scraped sites and mediocre outputs from the pipeline must not contaminate training. Filter by audit outcome and content authenticity signals.
