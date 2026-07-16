# Pre-registered rubrics — two real-case dispatches, 2026-07-16

Registered BEFORE the runs, per this workstream's benchmark discipline (fixloop
docs are blinded out of the loop's corpus). Grade the verdicts against these.

## BUG A — `GenerateText` discards `stop_reason` (silent truncation)

- **Symptom dispatched:** mechanism-only. `GenerateText`
  (`platform/aiservice/anthropic.go`) unmarshals the API response into a struct
  carrying only `Content` and `Usage`; `stop_reason` is never decoded, so a
  `max_tokens`-truncated HTTP 200 is returned as a complete success. **No
  downstream page-rendering claim this time** (that clause was stale in the prior
  run — the render_component guard now catches it loudly).
- **Live evidence:** 17 `llm_call_log` rows where `output_tokens == max_tokens`,
  all `success=true`, `error_message` NULL, across 5 agent types.
- **Expected CONFIRMED** citing the response struct (`anthropic.go`, the
  `var response struct { Content … Usage … }` block — no `stop_reason` field).
- **Expected fix:** add `StopReason` to the struct; return an error/typed signal
  when it is `"max_tokens"`. Fail loud at the client boundary.
- **PASS** = CONFIRMED grounded on that struct, symptom_check closes.
  **PARTIAL** = UNVERIFIABLE naming the struct.
  **FAIL** = REFUTED, or a fix that only bumps a caller's max_tokens.

## BUG B — root `ai_service` shadows step-level (dead per-step config)

- **Symptom dispatched:** `ExecuteLLMPromptAction` (`ai_actions.go`) resolves
  `ai_service` by checking the agent's ROOT `ai_service` FIRST and only falling
  back to the step's block `if aiServiceConfig == nil`. So when a root block
  exists, the step's entire `ai_service` — including `max_tokens` — is dead. This
  INVERTS the documented runbook gotcha ("max_tokens lives INSIDE a step's
  ai_service; root is dead config").
- **Proof by experiment (this session):** `diagnose-agent` verdict logged
  `max_tokens=2048` while its step block said 8000; only after moving max_tokens
  to the ROOT block did the call log 32000.
- **Live evidence:** 17 agents fleet-wide have a root `ai_service` with no
  `max_tokens` (→ hardcoded 2048); 10 of them (whole `content-creator-*` family)
  ALSO declare `max_tokens` elsewhere — dead config their authors believe live.
- **Expected CONFIRMED** citing `ai_actions.go` (root read first, `if
  aiServiceConfig == nil` fallback).
- **Expected fix:** step-level should override or merge with root (step wins),
  OR read `max_tokens` from the step block even when a root block exists.
- **PASS** = CONFIRMED grounded on the precedence logic. **FAIL** = REFUTED, or a
  verdict that just says "set max_tokens on the root" without naming the
  shadowing as the defect.

## Cross-cutting note
Both are config-plumbing bugs with the SAME consequence (2048 truncation), and
BUG A is why that truncation is invisible. A strong loop may connect them; that
is a bonus, not required for PASS.
