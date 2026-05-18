# FOCUS — Step-level LLM Config Settings Are Silently Ignored

Date: 2026-05-18
Status: workaround applied (top-level ai_service.max_tokens for site-adoption-agent), structural fix pending

## What's wrong

In `platform/orchestration/actions/execute_llm_prompt_action.go`, `ExecuteLLMPromptAction` reads LLM parameters (max_tokens, temperature) from two places only:

1. `agentConfig["max_tokens"]` — the agent's top-level `default_config`
2. `aiServiceConfig["max_tokens"]` — the top-level `ai_service` object inside `default_config`

It never reads `params.StepConfig.Config["max_tokens"]` — the step's own config.

This means every `workflow.steps.<step>.config.max_tokens` setting in every agent definition is **dead config**. Same for `temperature`. The values appear correct when you query `agent_definitions`, but the chassis ignores them.

When no value is found in either supported location, `options["max_tokens"]` is never set in the call to the AI client. `AnthropicClient.GenerateText` in `platform/aiservice/anthropic.go` (around line 110460 of the consolidated chassis context) then falls back to a hardcoded `"max_tokens": 2048` in its request body.

## How we found this

2026-05-17 gamesdesign.co.uk readoption. analyze_site step has `max_tokens: 32000` in its step config. The actual call truncated at output_tokens=2048, dropping 8 of the 20 pages mid-JSON. The LLM hadn't filtered them — `prompt_rendered LIKE '%pathfinding%'` etc. all returned `true`. The cap was 2048, not 32000.

Traced through:
- `llm_call_log` showed `output_tokens: 2048` (suspicious round number)
- Chassis read order in `ExecuteLLMPromptAction` (lines 213-218) only checks `agentConfig` and `aiServiceConfig`
- Neither had `max_tokens` on `site-adoption-agent`
- `AnthropicClient.GenerateText` hardcodes 2048 as its default request body value

## Workaround applied 2026-05-18

Patched `site-adoption-agent.default_config.ai_service` to add `max_tokens: 16000`. The chassis's existing fallback chain picks that up. All steps in the agent now use the same ceiling (16000), which is wasteful for `classify_archetype` (~1500 tokens naturally) but covers `analyze_site` for ~80-page sites.

Same workaround can be applied to any other agent that hits truncation: add `max_tokens` to the top-level `ai_service` object.

## What's still wrong after the workaround

- The workaround applies one max_tokens to **all** LLM steps in the agent. Steps that need different ceilings (analyze_site: 32000 if we want to handle 150-page sites; classify_archetype: 4000 to bound the response) can't have them.
- `temperature` is similarly ignored. Step-level `temperature: 0.2` in analyze_site is dead config. The Anthropic API receives no temperature parameter and uses its default (around 1.0). Same fallback fix (add `temperature` to top-level `ai_service`) would work for a global override, but again — one value for all steps.
- This affects **every agent** in the platform, not just site-adoption-agent. Any agent with multi-step LLM workflows where steps want different settings has the same dead-config problem.

## The structural fix

In `ExecuteLLMPromptAction`, change the max_tokens read from:

```go
var maxTokens float64
if maxTokens, ok = agentConfig["max_tokens"].(float64); ok {
    options["max_tokens"] = int(maxTokens)
} else if maxTokens, ok = aiServiceConfig["max_tokens"].(float64); ok {
    options["max_tokens"] = int(maxTokens)
}
```

to:

```go
var maxTokens float64
// Step-level config has highest priority (per-step granularity)
if maxTokens, ok = params.StepConfig.Config["max_tokens"].(float64); ok {
    options["max_tokens"] = int(maxTokens)
} else if maxTokens, ok = agentConfig["max_tokens"].(float64); ok {
    options["max_tokens"] = int(maxTokens)
} else if maxTokens, ok = aiServiceConfig["max_tokens"].(float64); ok {
    options["max_tokens"] = int(maxTokens)
}
```

Same pattern for the temperature read around line ~370 of the same file:

```go
if temp, ok := params.StepConfig.Config["temperature"].(float64); ok {
    options["temperature"] = temp
} else if temp, ok := agentConfig["temperature"].(float64); ok {
    options["temperature"] = temp
}
```

Also worth raising the hardcoded `2048` in `AnthropicClient.GenerateText` to something less footgun-y — say 8000. If config is missing entirely, 2048 truncates almost everything; 8000 truncates only large outputs and is a safer floor.

After the fix:
- `site-adoption-agent`'s analyze_site uses its existing 32000 step setting (overrides the 16000 top-level)
- classify_archetype uses 4000, derive_content_direction uses 6000, generate_design_intent uses 4000
- The top-level 16000 becomes the fallback for steps that don't specify
- Existing config across every agent starts working

## Things to check after the structural fix lands

When the code change deploys, audit every agent's step-level settings to make sure the existing values are reasonable. Some may have been set high speculatively when the author thought "this is the cap" but in practice the agent has been running on 2048 and producing acceptable output — once the cap lifts, costs and latency could increase. Verify each step's natural output size:

```sql
-- For each agent and step, see average output_tokens in recent successful calls
SELECT agent_type, step_name, model,
       COUNT(*) AS calls,
       AVG(output_tokens)::int AS avg_out,
       MAX(output_tokens) AS max_out,
       PERCENTILE_CONT(0.95) WITHIN GROUP (ORDER BY output_tokens)::int AS p95_out
FROM llm_call_log
WHERE success = true
  AND created_at > now() - interval '14 days'
GROUP BY agent_type, step_name, model
ORDER BY agent_type, step_name;
```

If `max_out` is close to 2048 for many step/agent combinations, those steps have been silently truncated for a long time — the structural fix will improve their output quality, possibly significantly.

## Related but out of scope

- The chassis may have similar dead-config issues for other AI service parameters: `model` (currently read only from `aiServiceConfig["model"]`), `top_p`, `top_k`, `stop_sequences`. Audit during the structural fix.
- The 2048 hardcoded default in `AnthropicClient.GenerateText` is one of several. Other providers' clients (`OpenAIClient`, etc.) likely have their own defaults — should be aligned.

## Wider scope — `ai_service` itself has the same bug

While auditing this, found the same pattern applies to the whole `ai_service` object read, not just its `max_tokens` field. In `ExecuteLLMPromptAction` around lines 110-145:

```go
// Order of resolution for aiServiceConfig (the object passed to createAIClient)
1. agentConfig["ai_service"]          ← top-level, checked FIRST
2. Any ai_service found inside workflow.steps (loop search)
3. params.StepConfig.Config["ai_service"]  ← step config, checked LAST
```

For step-level overrides to do anything, the order needs to be reversed (step-most-specific first). Today, **if the agent has a top-level `ai_service`, every step's own `ai_service` is ignored.**

This is exactly the override that `023_llm_quality_testing.md` lines 171-219 are trying to apply when swapping one step to Ollama. The SQL writes the new `ai_service` into `workflow.steps.<step>.config.ai_service` — and the chassis silently keeps using the agent's top-level `ai_service` instead. The model swap appears applied at the database level but doesn't take effect at runtime.

**This affects model swaps, provider swaps, per-step Anthropic key separation, anything that relies on step-level ai_service.** Worth verifying with a test (swap one step to a deliberately-broken model URL; if the step still succeeds, the override isn't being read).

Verification query for any agent suspected of having ignored overrides:

```sql
SELECT
    type,
    default_config->'ai_service'->>'model' AS top_level_model,
    step_key.key AS step_name,
    step_key.value->'config'->'ai_service'->>'model' AS step_model
FROM agent_definitions,
     LATERAL jsonb_each(default_config->'workflow'->'steps') AS step_key
WHERE deleted_at IS NULL
  AND is_active = true
  AND step_key.value->'config'->'ai_service' IS NOT NULL
  AND (step_key.value->'config'->'ai_service'->>'model') IS DISTINCT FROM
      (default_config->'ai_service'->>'model')
ORDER BY type, step_name;
```

Any row this returns is a step whose `ai_service` override is currently being ignored.

## The full structural fix (revised)

Three changes in `ExecuteLLMPromptAction`:

1. Reverse the `aiServiceConfig` resolution order: step config first, then workflow search, then agent top-level.
2. Add step-first lookup for `max_tokens` (per the original section above).
3. Add step-first lookup for `temperature` (per the original section above).

After this lands:
- Doc 023's model-swap SQL actually takes effect
- analyze_site's `max_tokens: 32000` actually takes effect
- Per-step temperatures actually take effect
- The current workaround (top-level ai_service with max_tokens=16000) becomes the fallback default for steps that don't override

Need to do the structural fix as one change set because changing the resolution order without also covering max_tokens/temperature would leave the bug half-fixed and confusing.

## References

- `platform/orchestration/actions/execute_llm_prompt_action.go` lines 110-145 (aiServiceConfig resolution), 213-218 (max_tokens), ~370 (temperature)
- `platform/aiservice/anthropic.go` `GenerateText` function, hardcoded `"max_tokens": 2048`
- Original diagnosis: this conversation, 2026-05-17 → 2026-05-18
- Symptom debug guide entry: section 6.6 of `016_debugging_guide_v2_11_.md`
- Related: `023_llm_quality_testing.md` — its model-swap SQL is currently a no-op against agents with top-level ai_service
- LLM quality follow-up: when implementing quality measurement, the audit above feeds into knowing whether existing baselines are real or truncation-masked
