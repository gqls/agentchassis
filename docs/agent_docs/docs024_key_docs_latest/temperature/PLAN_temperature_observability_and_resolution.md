# PLAN — temperature logging, A/B resolution, and per-field config resolution

Takes forward the temperature thread from `FOCUS_chassis_config_location_bugs.md` and
`FOCUS_step_level_llm_config_ignored(3).md`. Those docs left the temperature path at
"confirmed-broken-observability, root-cause-unverified" as of 2026-05-18. This plan
records the verification done against the current chassis snapshot and sequences the
remaining work.

## Verification against current chassis snapshot (2026-05-26)

Source checked: `production_agent-chassis-actions-current_context.txt` (repo snapshot,
file mtime 2026-05-26 13:38). This is repo context, not a guarantee of the deployed
image — see Step 2 for the live confirmation.

| Element | Location in snapshot | State |
|---|---|---|
| `llm_call_log` schema | `schemas_all` | `temperature real`, `max_tokens integer`, `model_resolved` all present. No migration needed. |
| Writer INSERT | `llm_call_logger.go` line 36714 | 20 columns. `temperature` and `max_tokens` **not** inserted. |
| `LLMCallLogParams` struct | `llm_call_logger.go` line 36755 | No `Temperature`, no `MaxTokens` field. |
| Failed-call log site | `execute_llm_prompt_action.go` line 10991 | Does not pass temperature/max_tokens. |
| Success log site | `execute_llm_prompt_action.go` line 11115 | Does not pass temperature/max_tokens. |
| Temperature read | line 10937 | Still single read: `agentConfig["temperature"].(float64)`. Per-field fix not applied. |
| max_tokens read | lines 10941–10946 | Two-level fallback: `agentConfig` then `aiServiceConfig`. |
| `options["temperature"]` / `options["max_tokens"]` | lines 10937–10946 | Computed and passed to `aiClient.GenerateText` (line 10983). Available to log cheaply. |
| `UseNumber()` in actions file | — | None found. Leans toward the `float64` assertion succeeding for DB-loaded config, but the upstream message-decode path (coordinator) is not in this snapshot, so possibility B is not fully ruled out statically. |

**Conclusion:** the planned `[next chassis deploy]` logging change has not landed in this
snapshot. The schema column exists but is never written, so `llm_call_log.temperature`
remains NULL by construction.

## Step 1 — Observability: capture temperature and max_tokens in the writer

Smallest change that makes possibility A (logging gap) vs B (temperature never set →
API default ~1.0) distinguishable from the log. Reuses the existing `LogLLMCall` /
`LLMCallLogParams` path; no new function or table.

Changes, all in `platform/orchestration/actions/llm_call_logger.go` plus the two call
sites in `execute_llm_prompt_action.go`:

1. Add two fields to `LLMCallLogParams`. Use pointer/`interface{}` for temperature so an
   explicitly configured `0.0` (fully deterministic) is distinguishable from "unset".
   `max_tokens` can reuse the existing `nullIfZero` pattern since 0 is never a real cap.

   ```go
   // additive — no existing field renamed
   Temperature interface{} // nil when not set in options; float64 otherwise
   MaxTokens   int          // 0 → logged as NULL via nullIfZero
   ```

2. Add both columns to the INSERT. Column list grows from 20 to 22; placeholders
   renumber to `$21`, `$22`. **Flag:** this renumbering is the one place to check
   carefully against the live column order — verify against `schemas_all` before deploy.

   ```go
   INSERT INTO llm_call_log (
       ..., prompt_variant, rag_context_used,
       temperature, max_tokens
   ) VALUES (..., $19, $20, $21, $22)
   ```
   Append `params.Temperature` (interface{}, nil-safe) and `nullIfZero(params.MaxTokens)`.

3. At both call sites, capture from the `options` map already built at lines 10937–10946,
   i.e. the values actually sent to the API:

   ```go
   var loggedTemp interface{}
   if t, ok := options["temperature"].(float64); ok {
       loggedTemp = t
   }
   loggedMaxTokens, _ := options["max_tokens"].(int)
   ```
   Pass `Temperature: loggedTemp, MaxTokens: loggedMaxTokens` into both the failed-call
   (line 10991) and success (line 11115) `LLMCallLogParams`.

No workflow variable names touched. No subworkflow added. Change is confined to one
action file and its logger helper.

## Step 2 — Deploy, then re-run the audit query to resolve A vs B

First confirm the deployed image actually matches the snapshot before assuming the
change is needed (in case a later deploy already differs):

```sql
-- Live state of the observability gap. If 0 non-null over 14 days,
-- deployed code matches the snapshot and Step 1 is genuinely outstanding.
SELECT COUNT(*) AS total,
       COUNT(temperature) AS non_null_temp,
       COUNT(max_tokens)  AS non_null_maxtok
FROM llm_call_log
WHERE created_at > now() - interval '14 days';
```

After Step 1 deploys, re-run the same query (or scope to calls after the deploy time):

- **Non-null temperature now appears** → it was a pure logging gap (possibility A). The
  read works; temperature was reaching the API all along. Proceed to Step 3 only to lift
  the dead step-level settings.
- **Still NULL even though the writer now inserts `options["temperature"]`** → means
  `options["temperature"]` itself is unset, i.e. the `float64` assertion at line 10937 is
  failing (possibility B). Every call has been running at the API default ~1.0. Then
  trace the upstream decode path: check the coordinator's message unmarshalling and the
  `agent_definitions` loader for a `json.Decoder` with `UseNumber()`; if present, JSONB
  numbers arrive as `json.Number` and the assertion drops silently. Fix is a `json.Number`
  fallback in the read (folded into Step 3's helpers).

## Step 3 — Structural fix: per-field resolution for temperature

From `FOCUS_chassis_config_location_bugs.md`. Extend the temperature read to the same
fallback chain `max_tokens` effectively has, so step-level settings stop being dead
config. Define the resolution helpers once and reuse for temperature, max_tokens, and
model rather than duplicating the assertion inline.

```go
// resolve once, reuse for every LLM param
func readFloat(m map[string]interface{}, key string) (float64, bool) {
    switch v := m[key].(type) {
    case float64:
        return v, true
    case json.Number: // possibility-B guard
        f, err := v.Float64()
        return f, err == nil
    }
    return 0, false
}
```

Temperature fallback chain (step config → agent top → step ai_service → top-level
ai_service), mirroring the proposed `max_tokens` chain:

```go
var temperature float64
if v, ok := readFloat(params.StepConfig.Config, "temperature"); ok {
    temperature = v
} else if v, ok := readFloat(agentConfig, "temperature"); ok {
    temperature = v
} else if v, ok := readNestedFloat(params.StepConfig.Config, "ai_service", "temperature"); ok {
    temperature = v
} else if v, ok := readNestedFloat(agentConfig, "ai_service", "temperature"); ok {
    temperature = v
}
// only set if found, so Step 5's default can apply when truly unset
```

This lifts the 6 currently-dead step-level temperatures:
`site-adoption-agent` (analyze_site 0.2, classify_archetype 0.2, derive_content_direction
0.2, generate_design_intent 0.3) and `tool-recreation-handler` (2 steps).

## Step 4 — Validate the dead step-level temperatures take effect

After Step 3 deploys, the 6 step-level settings should show their configured values in
`llm_call_log.temperature` (now observable thanks to Step 1):

```sql
SELECT agent_type, step_name,
       COUNT(*) AS calls,
       array_agg(DISTINCT temperature) AS temps_seen
FROM llm_call_log
WHERE success = true
  AND created_at > now() - interval '3 days'
  AND agent_type IN ('site-adoption-agent','tool-recreation-handler')
GROUP BY agent_type, step_name
ORDER BY agent_type, step_name;
```

Expected: `analyze_site`, `classify_archetype`, `derive_content_direction` show 0.2;
`generate_design_intent` shows 0.3; the two tool-recreation-handler steps show their
configured values. Outputs for these extraction/classification steps should also become
more deterministic.

## Step 5 — Hardening: default temperature when none configured (optional, last)

Anthropic's default ~1.0 is creative — likely too high for extraction/classification
prompts. Once Steps 1–3 confirm the path end to end, consider a chassis-level default
(~0.4) applied only when no temperature is set at any level, overridable by an explicit
value. Hold this until the read path is proven, so we don't stack a default on top of an
unverified read.

## Variable-name / change-surface notes

- `LLMCallLogParams` gains `Temperature` and `MaxTokens` — additive, nothing renamed.
- INSERT column count 20 → 22; placeholder renumber `$21`/`$22` is the one spot to verify
  against live column order before deploy.
- Step 3 adds `readFloat` / `readNestedFloat` helpers (and a `json.Number` guard); these
  replace the inline single assertion at line 10937. The replacement is intentional and
  noted here.
- No workflow YAML/JSON variable names changed. No subworkflows introduced. Sub-agent
  spawning model unaffected.

## Sequencing summary

1. Writer change (observability) — cheap, unblocks everything.
2. Deploy + audit query — resolves A vs B from the log.
3. Per-field resolution fix — lifts dead step-level temperatures.
4. Validate the 6 settings land.
5. Default-temperature hardening — only after the path is proven.

## References

- `platform/orchestration/actions/execute_llm_prompt_action.go` — read at 10937,
  max_tokens 10941–10946, options→GenerateText 10983, log sites 10991 / 11115.
- `platform/orchestration/actions/llm_call_logger.go` — INSERT 36714, struct 36755.
- `platform/aiservice/anthropic.go` — `GenerateText`, hardcoded `max_tokens: 2048` fallback.
- `FOCUS_chassis_config_location_bugs.md`, `FOCUS_step_level_llm_config_ignored(3).md`
- `016_debugging_guide_v2(15).md` section 6.6
- Original diagnosis: 2026-05-17 → 2026-05-18. Snapshot verification: 2026-05-26.
