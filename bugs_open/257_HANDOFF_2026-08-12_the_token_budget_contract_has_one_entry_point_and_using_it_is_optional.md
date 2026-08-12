# 257 — the token-budget contract has exactly ONE enforced entry point and using it is OPTIONAL, so every direct-to-provider call silently runs at 2048

**Filed:** 2026-08-12 by the `provocation_pipeline` lane, at the owner's request
("write up a bugs_open record for the two copies of one rule budget bug so it's not
lost"). **Status: OPEN, unowned.** **Severity: not a live burn — a latent defect class
with three known instances, one of which is live and unrecorded (§4).**

**Class:** structural. The mechanism is a contract that is not enforced at its own
boundary, so conformance depends on each caller remembering. **Sibling of
`bugs_open/205`, not a duplicate** — 205 is one *instance* (a step whose budget nobody
configured); this is the *class* (a budget that is configured and never arrives, plus
every future caller that will reproduce it). They want different fixes and 205's fix
does not close this.

## Declared substitute for the 090 diagnosis loop (owner ruling 2026-07-31)

This file asserts a cross-cutting root cause without a `090` run. The substitute, stated
plainly rather than omitted:

- **The failure was reproduced live, three times, and the fix verified at the pod.**
  Orchestrations `c49adc3f`, `f6a2ab74`, `9c9d52db` (2026-08-10) all died
  `stop_reason=max_tokens`; the last reported `output_tokens=2048` **after**
  migration 372 had set `ai_service.max_tokens = 8000` on that very step.
- **Every link was read first-hand** and is quoted below with file:line.
- **Four independent council seats reached the same structural conclusion** in round
  `65d153f0` (`reuse_agent`, `constitution`, `prior_art_librarian`, `architecture`) —
  which is stronger corroboration for a *design* claim than one loop iteration, because
  they disagreed with the filing session rather than confirming it.
- **The census the `architecture` seat asked for has now been run** (§4). It found a
  third instance nobody had recorded.

## 1. The mechanism, quoted

The provider clients read the token budget from **one place only** — the `options` map
passed to `GenerateText`:

```go
// platform/aiservice/anthropic.go:109
requestBody := map[string]interface{}{ "model": c.model, "max_tokens": 2048, ... }
// platform/aiservice/anthropic.go:147
if maxTokens, ok := options["max_tokens"]; ok { requestBody["max_tokens"] = maxTokens }
```

`ai_service.max_tokens` in `agent_definitions` is **not** read by the client. It is read
by `ExecuteAIStepAction`, which translates it into that map:

```go
// platform/orchestration/actions/ai_actions.go:358-364
var maxTokens float64
if maxTokens, ok = agentConfig["max_tokens"].(float64); ok { options["max_tokens"] = int(maxTokens) }
else if maxTokens, ok = aiServiceConfig["max_tokens"].(float64); ok { options["max_tokens"] = int(maxTokens) }
else { logger.Warn("max_tokens not configured at any level; provider hardcoded default will apply") }
```

**So the contract is: "call through `ExecuteAIStepAction`, or your configuration is
inert."** Nothing states that, nothing checks it, and the compiler is happy either way —
`GenerateText` accepts `nil`.

## 2. Why it is invisible, which is the part that makes it a bug worth filing

Three independent things hide it, and they compound:

1. **The config looks right.** Reading `agent_definitions` shows `max_tokens: 8000`. The
   value is present, well-formed, at the documented path. **A key nothing reads is
   indistinguishable, from the config side, from one that works.**
2. **The error echoes the OLD number.** After setting 8000 the failure still said
   `output_tokens=2048`. The filing session read that as "still too small" and **changed
   live config a second time** before opening the call site. *A value the error repeats
   back unchanged after you have changed it is the signature of a key nothing reads, and
   no further config change can help.*
3. **These calls write no `llm_call_log` row**, because that table is written by
   `ExecuteAIStepAction` — the very function being bypassed. **Measured 2026-08-12:**
   `SELECT ... FROM llm_call_log WHERE agent_type ILIKE '%reason%'` returns **0 rows**,
   for a service that is deployed and running. So the fleet's own instrument for "which
   steps are truncating" is blind to exactly the calls that are most likely to.

## 3. The duplication the council objected to

The fix shipped on 2026-08-10 (`36b2dc54e`, `llm_options.go`) gave the two provocation
actions a helper that builds the map. It works, and it left **two hand-maintained
implementations of one precedence rule** — `ai_actions.go:358-364` and
`llmOptionsFromConfig`. Four seats objected. `architecture` put it best:

> *"the options-map contract has exactly one enforced entry point and it is optional to
> use ... any action author who reaches for `client.GenerateText` instead of
> `ExecuteAIStepAction` reproduces this bug by construction."*

**There are now also two different warning strings for the same condition** — 205's
`"max_tokens not configured at any level"` (`ai_actions.go`) and
`"no max_tokens configured at step or ai_service level"` (`llm_options.go`) — so even
the observability of the defect has forked.

⚠ **The filing session's precedence claim was WRONG and is corrected here:**
`llm_options.go` originally said its precedence "mirrors `ExecuteAIStepAction`". It does
not. That function's outer key is `agentConfig` =
`CollectedData["agent_config"]` → `agentDef.DefaultConfig` (`ai_actions.go:180,219`) —
the **agent's** whole config. The helper's is the **step's**. *The two copies had already
drifted before either was reviewed*, which is the objection proving itself inside a week.

## 4. Census — every direct provider-client call site [MEASURED 2026-08-12]

The `architecture` seat flagged this as unanswerable by the code index (it does not index
function bodies, so a content search returns zero regardless of how many call sites
exist). Run as a grep instead, excluding `platform/aiservice/` itself and tests:

| call site | options passed | budget reaching the API |
|---|---|---|
| `ai_actions.go:415,624,764` | `options` (built at :358) | **correct** — the canonical path |
| `provocation_gate_action.go:842` | `opts` | fixed 2026-08-10 |
| `provocation_generator_action.go:548` | `opts` | fixed 2026-08-10 |
| `companies_house_llm_review_action.go:133` | `llmOptions` | sets `max_tokens` — OK |
| `execute_vision_prompt_action.go:197` | `options` | sets `max_tokens` — OK |
| `contentcreator/agent.go:305` | `options` | sets `max_tokens` — OK |
| `tools-api/handlers/defend.go:101` | `{"max_tokens": 8192}` | hardcoded, config-blind, but not 2048 |
| `tools-api/handlers/position.go:90` | `{"max_tokens": 8192}` | same |
| **`internal/agents/reasoning/agent.go:127`** | **`nil`** | **2048, always** |

**THE NEW FINDING: `reasoning-agent` is hard-pinned to 2048 output tokens and cannot be
configured.** The file contains **zero** occurrences of `max_tokens`. It is one of the
14 deployed backend services, it has its own kustomize overlay, and nothing in
`agent_definitions` can raise its ceiling.

`[UNMEASURED]` — **whether it is actually truncating in production is NOT established,
and the usual instrument cannot answer it** (§2.3: no `llm_call_log` rows). Do not record
this as a live burn until someone checks. **The check:** the client returns a real error
on a cut (`anthropic.go:246` → `TruncatedError`), so look for that error text in the
reasoning-agent pod's logs or in whatever it writes on failure — *not* in `llm_call_log`,
which will be empty either way and would read as "no truncations".

## 5. Fix candidates, ordered by what makes the bad state unrepresentable

1. **Resolve the budget inside the provider client** (`aiservice.NewClient` already
   receives the whole `ai_service` config — `ai_actions.go:1397-1399`). If the client
   read `max_tokens` from the config it was constructed with, **every** caller would
   inherit it, including `nil`-passing ones, and the defect class would stop being
   reachable. This is the only candidate that removes the choice rather than documenting
   it. It is also the one with real blast radius (every LLM call in the fleet), so it
   wants its own council round and a staged rollout.
2. **Extract `ai_actions.go:358-364` into `llmOptionsFromConfig`** and have
   `ExecuteAIStepAction` call it, deleting its inline copy. This is what the four seats
   actually asked for. It ends the drift but leaves the contract optional — a future
   direct caller still reproduces the bug.
3. **Fix `reasoning-agent` specifically** (§4). Narrow, safe, independently worthwhile
   whatever happens to 1 and 2.
4. **A lint or census in CI** over direct `GenerateText`/`GenerateWithImages` call sites,
   so the next one is caught at review rather than by a production truncation. The grep
   in §4 is the whole implementation.
5. **Do NOT** simply widen the warning. 205 already added one and it did not prevent
   this; a second warning for the same condition is what §3 is complaining about.

## 6. How to verify any fix

- **At the pod, both replicas**, never at git or the tag — grep a string the change ADDS
  *and* one it REMOVES (a removed string is the stronger control).
- **Behaviourally:** a step whose configured `max_tokens` exceeds 2048 must produce a
  completion longer than 2048 output tokens. `output_tokens` in the response is the
  measurement; the config value is not.
- **The disconfirming result must be possible.** If the check would pass on a binary
  that still ignores the config, it is not a check. The provocation lane's own
  `TestNoProvocationActionCallsAModelWithAnEmptyOptionsMap` is mutation-proven for this
  reason — delete the fix and it fails naming file and line.

## 7. Related

`bugs_open/205` (the unconfigured instance; its candidate 1 added the first warning) ·
`bugs_open/244` (council prompt spend — adjacent, same table, different mechanism) ·
council round **`65d153f0`** REVISE, where four seats raised §3 ·
`docs024_key_docs_latest/provocation_pipeline/NOTES_provocation_pipeline.md` under
2026-08-10 for the full reproduction · `WRONG_CALLS.md` 2026-08-10 for the two live
config changes made against a key nothing reads.
