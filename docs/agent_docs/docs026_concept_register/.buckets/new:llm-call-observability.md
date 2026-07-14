
<!-- SOURCE: U13_docs024_small_dirs.md -->
### Temperature/max_tokens logging gap in llm_call_log
- **category:** NEW:llm-call-observability
- **status-signal:** partial
- **status-evidence:** "the schema column exists but is never written, so llm_call_log.temperature remains NULL by construction" — verified against a 2026-05-26 chassis snapshot
- **what:** Although `llm_call_log` already has `temperature real` and `max_tokens integer` columns, the Go writer (`llm_call_logger.go`) never populates them, and the two call sites in `execute_llm_prompt_action.go` don't pass the values through — even though the actual values sent to the LLM API are already computed a few lines earlier. This makes it impossible to observe from the log alone whether a configured temperature ever reached the API call.
- **sources:** temperature/PLAN_temperature_observability_and_resolution.md#Verification table,#Step 1
- **relations:** Per-field LLM config resolution fallback chain; Possibility-A-vs-B diagnostic method
- **verify-later:** platform/orchestration/actions/llm_call_logger.go LLMCallLogParams struct; execute_llm_prompt_action.go call sites

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Per-field LLM config resolution fallback chain (temperature parity with max_tokens)
- **category:** NEW:llm-call-observability
- **status-signal:** aspirational
- **status-evidence:** "Per-field fix not applied" — temperature read is "Still single read: agentConfig['temperature'].(float64)" versus max_tokens' existing two-level fallback
- **what:** Proposes lifting temperature to the same multi-level fallback chain max_tokens already has (step config → agent top → step ai_service → top-level ai_service) via shared `readFloat`/`readNestedFloat` helpers, replacing the single inline float64 type assertion. Would activate 6 currently-dead step-level temperature settings configured in the DB but never actually taking effect.
- **sources:** temperature/PLAN_temperature_observability_and_resolution.md#Step 3,#Step 4
- **relations:** Temperature/max_tokens logging gap; Possibility-A-vs-B diagnostic method
- **verify-later:** readFloat/readNestedFloat helpers (proposed, not yet added)

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Possibility-A-vs-B diagnostic method for silent LLM config failures
- **category:** NEW:llm-call-observability
- **status-signal:** partial
- **status-evidence:** "Smallest change that makes possibility A (logging gap) vs B (temperature never set → API default ~1.0) distinguishable from the log" with an exact before/after SQL audit query
- **what:** A diagnostic technique for a silently-broken config field: ship the cheapest possible observability fix first before attempting any structural resolution fix, then re-run a COUNT(*)/COUNT(temperature)/COUNT(max_tokens) audit query pre- and post-deploy. If temperature becomes non-null post-deploy, the bug was pure logging gap; if still null, the upstream read itself is silently failing — distinguishing the two determines whether every historical LLM call ran at the intended temperature.
- **sources:** temperature/PLAN_temperature_observability_and_resolution.md#Step 2,#Sequencing summary
- **relations:** Temperature/max_tokens logging gap; Per-field LLM config resolution fallback chain
- **verify-later:** SQL audit query in Step 2

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Default temperature hardening (chassis-level fallback ~0.4)
- **category:** NEW:llm-call-observability
- **status-signal:** aspirational
- **status-evidence:** Explicitly sequenced last and conditional: "Hold this until the read path is proven, so we don't stack a default on top of an unverified read."
- **what:** Once the observability and per-field resolution fixes are proven, proposes a chassis-level default temperature (~0.4) applied only when none is configured at any level, overridable by an explicit value — reasoning that Anthropic's API default (~1.0) is likely too high for the extraction/classification-style prompts most affected.
- **sources:** temperature/PLAN_temperature_observability_and_resolution.md#Step 5
- **relations:** Per-field LLM config resolution fallback chain
- **verify-later:** n/a (not yet implemented; gated on Steps 1-3)

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Temperature/max_tokens logging gap in llm_call_log
- **category:** NEW:llm-call-observability
- **status-signal:** partial
- **status-evidence:** "the schema column exists but is never written, so llm_call_log.temperature remains NULL by construction" — verified against a 2026-05-26 chassis snapshot
- **what:** Although `llm_call_log` already has `temperature real` and `max_tokens integer` columns, the Go writer (`llm_call_logger.go`) never populates them, and the two call sites in `execute_llm_prompt_action.go` don't pass the values through — even though the actual values sent to the LLM API are already computed a few lines earlier. This makes it impossible to observe from the log alone whether a configured temperature ever reached the API call.
- **sources:** temperature/PLAN_temperature_observability_and_resolution.md#Verification table,#Step 1
- **relations:** Per-field LLM config resolution fallback chain; Possibility-A-vs-B diagnostic method
- **verify-later:** platform/orchestration/actions/llm_call_logger.go LLMCallLogParams struct; execute_llm_prompt_action.go call sites

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Per-field LLM config resolution fallback chain (temperature parity with max_tokens)
- **category:** NEW:llm-call-observability
- **status-signal:** aspirational
- **status-evidence:** "Per-field fix not applied" — temperature read is "Still single read: agentConfig['temperature'].(float64)" versus max_tokens' existing two-level fallback
- **what:** Proposes lifting temperature to the same multi-level fallback chain max_tokens already has (step config → agent top → step ai_service → top-level ai_service) via shared `readFloat`/`readNestedFloat` helpers, replacing the single inline float64 type assertion. Would activate 6 currently-dead step-level temperature settings configured in the DB but never actually taking effect.
- **sources:** temperature/PLAN_temperature_observability_and_resolution.md#Step 3,#Step 4
- **relations:** Temperature/max_tokens logging gap; Possibility-A-vs-B diagnostic method
- **verify-later:** readFloat/readNestedFloat helpers (proposed, not yet added)

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Possibility-A-vs-B diagnostic method for silent LLM config failures
- **category:** NEW:llm-call-observability
- **status-signal:** partial
- **status-evidence:** "Smallest change that makes possibility A (logging gap) vs B (temperature never set → API default ~1.0) distinguishable from the log" with an exact before/after SQL audit query
- **what:** A diagnostic technique for a silently-broken config field: ship the cheapest possible observability fix first before attempting any structural resolution fix, then re-run a COUNT(*)/COUNT(temperature)/COUNT(max_tokens) audit query pre- and post-deploy. If temperature becomes non-null post-deploy, the bug was pure logging gap; if still null, the upstream read itself is silently failing — distinguishing the two determines whether every historical LLM call ran at the intended temperature.
- **sources:** temperature/PLAN_temperature_observability_and_resolution.md#Step 2,#Sequencing summary
- **relations:** Temperature/max_tokens logging gap; Per-field LLM config resolution fallback chain
- **verify-later:** SQL audit query in Step 2

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Default temperature hardening (chassis-level fallback ~0.4)
- **category:** NEW:llm-call-observability
- **status-signal:** aspirational
- **status-evidence:** Explicitly sequenced last and conditional: "Hold this until the read path is proven, so we don't stack a default on top of an unverified read."
- **what:** Once the observability and per-field resolution fixes are proven, proposes a chassis-level default temperature (~0.4) applied only when none is configured at any level, overridable by an explicit value — reasoning that Anthropic's API default (~1.0) is likely too high for the extraction/classification-style prompts most affected.
- **sources:** temperature/PLAN_temperature_observability_and_resolution.md#Step 5
- **relations:** Per-field LLM config resolution fallback chain
- **verify-later:** n/a (not yet implemented; gated on Steps 1-3)
