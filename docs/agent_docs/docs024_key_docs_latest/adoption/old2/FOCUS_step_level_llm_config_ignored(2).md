# FOCUS — LLM Config Resolution Order Has Shadowing Bugs

Date: 2026-05-18 (revised)
Status: workaround applied for site-adoption-agent (top-level ai_service.max_tokens=16000); structural fix pending

This doc went through one wrong iteration before landing on the right mental model. The earlier version claimed "step-level config is never read." That's wrong. The actual situation is more subtle.

## The real bug — shadowing, not exclusion

`ExecuteLLMPromptAction` resolves `aiServiceConfig` in this order (lines 110-145 of the file):

1. `agentConfig["ai_service"]` — agent's top-level `default_config.ai_service`
2. *Only if (1) returned nil*: `workflow.steps.<currentStep>.config.ai_service`
3. *Only if (1) AND (2) returned nil*: `params.StepConfig.Config["ai_service"]`

The problem: **once (1) hits, (2) and (3) are skipped entirely**. Even if (1)'s ai_service is missing the field you care about (max_tokens, model, etc.), the chassis doesn't fall through to look at the step.

`max_tokens` is then read from `agentConfig["max_tokens"]` (top of default_config) first, falling back to `aiServiceConfig["max_tokens"]`. If `aiServiceConfig` was set from (1), step-level max_tokens inside the step's ai_service is invisible.

| Where you can put max_tokens | When the chassis reads it |
|---|---|
| `default_config.max_tokens` (very top) | Always — highest priority |
| `default_config.ai_service.max_tokens` | When (1) above is absent |
| `default_config.workflow.steps.<step>.config.ai_service.max_tokens` | Only when **no** top-level ai_service exists at all |
| `default_config.workflow.steps.<step>.config.max_tokens` (sibling of ai_service) | **NEVER** — chassis only reads max_tokens from within ai_service objects |

Most agents work fine because they have no top-level `ai_service`. The chassis falls through to (2) and reads each step's own `ai_service`, max_tokens included. The data confirms this — 38 of 60 active agents have step-level `ai_service` working today.

The agents that hit the bug are those with a top-level `ai_service` that lacks the field. Today: `site-adoption-agent` (lacked max_tokens, now patched), and `feed-triage` (has one shadowed step).

## How site-adoption-agent failed

Pre-patch state:
- Top-level `ai_service`: `{model, provider, api_key_env_var}` — present, but no `max_tokens`
- Step `analyze_site.config.ai_service`: `{model, provider, api_key_env_var}` — present, but no `max_tokens`
- Step `analyze_site.config.max_tokens: 32000` — present, but in a location the chassis never reads

Resolution:
1. Top-level ai_service found → `aiServiceConfig` populated → fallback paths skipped
2. `agentConfig["max_tokens"]` not present
3. `aiServiceConfig["max_tokens"]` (top-level) not present
4. No `options["max_tokens"]` set
5. AnthropicClient.GenerateText hardcoded fallback: `"max_tokens": 2048`

Output capped at 2048. analyze_site needed ~3500 tokens for 20 pages → JSON truncated mid-array → 8 pages dropped.

The `max_tokens: 32000` in the step config was always going to be ignored — it's a sibling of `ai_service`, not inside it. Even with the structural fix below, that location wouldn't be read; the fix is to put max_tokens at one of the live locations.

## The workaround (already applied)

Added `max_tokens: 16000` to `site-adoption-agent.default_config.ai_service`. Now the top-level ai_service has the field, the chassis reads 16000, all steps in the agent cap at 16000. Covers ~80 pages of analyze_site output and over-provides for the smaller steps (classify_archetype naturally hits ~1300, generate_design_intent ~1500).

For other agents that hit the same bug, the same fix works: add `max_tokens` to the top-level `ai_service`. The 38-agent list with no top-level ai_service doesn't need this patch — they're already reading step-level.

## Temperature — separate, more troubling

The temperature path is even simpler:

```go
if temp, ok := agentConfig["temperature"].(float64); ok {
    options["temperature"] = temp
}
```

Only one read: `agentConfig["temperature"]` — the very top of default_config. Not inside `ai_service`, not from step config. So:

| Where you can put temperature | Read by chassis? |
|---|---|
| `default_config.temperature` (very top) | Yes |
| `default_config.ai_service.temperature` | **No** |
| `default_config.workflow.steps.<step>.config.temperature` | **No** |
| `default_config.workflow.steps.<step>.config.ai_service.temperature` | **No** |

What we have today (per query 3):
- 9 agents have top-level `temperature` set: `component-creator=0.4`, `content-creator-about=0.7`, `content-creator-contact=0.7`, `content-creator-cta=0.8`, `content-creator-features=0.6`, `content-creator-hero=0.7`, `content-creator-hero-without-research=0.7`, `content-creator-testimonials=0.7`, `content_researcher=0.7`, `copywriter=0.7`, `reasoning=0.2`, `researcher=0.3`, `simple-content-writer-with-approval=0.7`
- 6 step-level temperature settings are dead config:
  - `site-adoption-agent`: 4 steps (analyze_site=0.2, classify_archetype=0.2, derive_content_direction=0.2, generate_design_intent=0.3)
  - `tool-recreation-handler`: 2 steps

## Temperature observability gap

A second, related issue surfaced from the audit: `llm_call_log.temperature` is **NULL for every call** over the past 14 days, even for the 9 agents that have top-level temperature set. Two possibilities:

A. The chassis IS setting temperature on the API call but isn't writing it to `llm_call_log.temperature`. Logging bug; behaviour is correct.

B. The chassis isn't actually setting temperature at all (the `if temp, ok` check fails for some reason — likely because the JSONB float is being unmarshalled as something other than float64, e.g. as json.Number). Real bug; everything runs at API default (~1.0).

To distinguish without code changes: check if `component-creator` output style differs visibly from `content-creator-cta`. Component-creator has temp=0.4 (deterministic), content-creator-cta has 0.8 (creative). If they read similarly, B. If component output is more structured/predictable, A.

More directly: instrument the chassis to log `options["temperature"]` at the callsite, deploy on one agent, observe.

Either way, the missing observability is a problem — we have no way to verify what temperature is actually being used today without code instrumentation.

## The structural fix (revised)

Three changes in `ExecuteLLMPromptAction`:

### 1. Per-field resolution, not per-object

Instead of resolving the whole `aiServiceConfig` object once and then reading fields from it, resolve each LLM parameter independently with its own fallback chain:

```go
// max_tokens: step config first, then agent config,
// then step ai_service, then top-level ai_service
var maxTokens int
if v, ok := readInt(params.StepConfig.Config, "max_tokens"); ok {
    maxTokens = v
} else if v, ok := readInt(agentConfig, "max_tokens"); ok {
    maxTokens = v
} else if v, ok := readNestedInt(params.StepConfig.Config, "ai_service", "max_tokens"); ok {
    maxTokens = v
} else if v, ok := readNestedInt(agentConfig, "ai_service", "max_tokens"); ok {
    maxTokens = v
}
if maxTokens > 0 { options["max_tokens"] = maxTokens }
```

Same shape for `temperature`, `model`, and any future LLM parameter.

This makes step-level fields override agent-level fields per-field, rather than the current all-or-nothing object resolution.

### 2. Raise the hardcoded floor in `AnthropicClient.GenerateText`

`platform/aiservice/anthropic.go` `"max_tokens": 2048` is too low a fallback. Raise to 8000 so missing-config doesn't silently truncate everything.

### 3. Fix the logging gap

In the `llm_call_log` insert path, capture `options["temperature"]` and `options["max_tokens"]` from what was actually sent to the API. Right now `temperature` is universally NULL in the log, which makes step-level diagnosis impossible.

## Things to check after the structural fix lands

Audit query — find every step whose output was being silently truncated:

```sql
SELECT agent_type, step_name, model,
       COUNT(*) AS calls,
       AVG(output_tokens)::int AS avg_out,
       MAX(output_tokens) AS max_out,
       SUM(CASE WHEN output_tokens >= 2048 THEN 1 ELSE 0 END) AS at_cap
FROM llm_call_log
WHERE success = true AND created_at > now() - interval '14 days'
GROUP BY agent_type, step_name, model
ORDER BY at_cap DESC;
```

Today (pre-fix) we know:
- `component-creator.generate_template` — avg 5927, max 13665, at_cap 138/153. Step-level ai_service.max_tokens is being read (no top-level), so this is genuinely natural responses, not truncation. Some at-cap rows may be legitimate completion at the configured cap.
- `feed-triage.score_relevance` — avg 2965, max 3243, at_cap 55/55. Top-level ai_service has max_tokens (per query 1) — likely the configured cap is the natural ceiling here.
- `site-adoption-agent.analyze_site` — avg 2048, max 2048, at_cap 10/10. Truncated — fixed by the 16000 patch, should drop dramatically on next call.
- `site-adoption-agent.derive_content_direction` — avg 2048, max 2048, at_cap 9/10. Same fix applies, same agent.

After the patch, re-run the audit; `analyze_site` and `derive_content_direction` should sit well below 2048 on average, with maxes well above.

## What changed from the previous version of this doc

- **Removed**: "step-level config is never read" claim
- **Added**: shadowing mechanism — top-level ai_service blocks step fallback even when fields are missing
- **Corrected**: 4-row table showing exactly where each location is read from
- **Confirmed**: query results from 2026-05-18 establish actual blast radius (2 affected agents for max_tokens, 6 dead step-temperature settings)
- **New**: temperature observability gap — `llm_call_log.temperature` is NULL across the board

## References

- `platform/orchestration/actions/execute_llm_prompt_action.go` lines 110-145 (aiServiceConfig resolution), 209 (temperature read), 213-218 (max_tokens reads)
- `platform/aiservice/anthropic.go` `GenerateText` function, hardcoded `"max_tokens": 2048`
- Original diagnosis: 2026-05-17 → 2026-05-18
- Symptom debug guide entry: section 6.6 of `016_debugging_guide_v2_12_.md`
- Related: `023_llm_quality_testing.md` — its per-step model swap works on agents without top-level ai_service, fails on agents with it
- LLM quality follow-up: the audit query above feeds into knowing whether existing prompt baselines are real or truncation-masked
