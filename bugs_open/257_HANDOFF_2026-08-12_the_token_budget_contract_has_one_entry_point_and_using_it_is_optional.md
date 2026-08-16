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

---

## §Path A COMMITTED 2026-08-16 — budget resolution moved INTO the provider clients (candidate 1)

*This section by the fixing session (lane:
`docs024_key_docs_latest/bugfix_257_token_budget_at_the_client/`). Council
correlation `85971036-5b9a-42ca-aeac-d77a65397c86` — `Council-Submitted:`
trailer; **read the verdict before treating this as reviewed.***

**Taken as OPEN and UNOWNED**, re-checked two ways because one is not enough:
`who-owns.py 257` finds only the filing commit and `(none identified)` for the
owning workstream; a grep of ~20 live session transcripts (the only instrument
that sees *uncommitted* work) shows no session on it. ⚠ `who-owns.py` still
printed `VERDICT: OWNED or recently active` — it fires on any recent commit, and
the filing IS a recent commit. Read its body, not its verdict line.

### The bug is still valid — but every line number in §1 has MOVED

The prompt-caching work (`bugs_open/244`) shifted them. Mechanism unchanged:

| §1 claim | filed as | verified at HEAD 2026-08-16 |
|---|---|---|
| provider hardcodes 2048 | `anthropic.go:109` | **`anthropic.go:185`** |
| only `options` can override | `anthropic.go:147` | **`anthropic.go:280-281`** |
| inline precedence copy | `ai_actions.go:358-364` | unmoved |
| `reasoning-agent` passes `nil` | `reasoning/agent.go:127` | unmoved; still 0 hits for `max_tokens` in the package |

### TWO CORRECTIONS TO THIS FILE'S OWN §4 CENSUS

1. **Gemini carries the identical hardcoded 2048** (`gemini.go`, `generate()`:
   `visibleBudget := 2048`). §4 is Anthropic-only because the filing session read
   one client. Fixing one and leaving its twin is 016b §9's own recurring shape —
   both are fixed here, plus ollama.
2. **There is a TENTH direct call site**, added after filing:
   `internal/tools-api/handlers/gripper.go:161` (commit `f967d9307`). It passes a
   Go constant so it is unharmed — but `llm_options.go`'s SCOPE note predicted
   *"a THIRD caller should be the extraction, not another paste"* and a third
   caller arrived within a week. **Re-run the census; do not quote §4.**

### What shipped

`platform/aiservice/max_tokens.go` (new) + the three provider clients. Precedence
at the wire:

1. `options["max_tokens"]` — per-call, **wins. UNCHANGED.**
2. **`ai_service.max_tokens` from the CONSTRUCTION config — NEW.**
3. provider fallback — **UNCHANGED**: `DefaultMaxOutputTokens` (2048) for
   anthropic/gemini; ollama still omits `num_predict` entirely.

Not new plumbing: every constructor already receives the whole `ai_service` map
and already reads `model` from it with a default fallback. It was discarding
`max_tokens`.

**Ollama is deliberately asymmetric** — it uses `configMaxTokens` (0 = "nothing
chosen"), *not* `resolvedMaxTokens`. This bug is about a *configured* budget
failing to arrive; imposing a 2048 ceiling where there never was one would be a
regression dressed as a fix.

**A zero-budget floor was added at the wire** (`if budget <= 0`) after measuring
that the existing suite passes with the field unset — so a struct-literal client
sending `max_tokens: 0` (a 400) was invisible to every test. It also means **no
existing test was edited** to accommodate this change.

### Blast radius — MEASURED. This change alters ZERO requests in the fleet today.

All **seven** client construction sites enumerated (`grep -rn
'aiservice\.(NewClient|NewAnthropicClient|NewGeminiClient|NewOllamaClient)\('`):

| site | config has `max_tokens`? | options carry it? | effect |
|---|---|---|---|
| `ai_actions.go:1398` (canonical) | sometimes | **whenever the config does** | **none — proof below** |
| `contentcreator/agent.go:123` | no | always (`:720`, default 3000) | none |
| **`reasoning/agent.go:71`** | **yes — 2048** | **no (`nil`)** | **key becomes LIVE; value unchanged** |
| `tools-api` gripper/position/defend | no (inline literals) | always | none |
| `rag_actions.go:374` | n/a — embeddings only | n/a | none |

**The canonical path is provably unchanged**: `createAIClient(ctx,
aiServiceConfig)` (`:287`) and the options build (`:361`) read the **same merged
map** from `resolveAIServiceConfig` (`:243`). Key present → options carry it →
rule 1 → identical. Absent → client default absent → rule 3 → identical.
Agent-level only (outside `ai_service`, never passed to a constructor) → options
carry it → identical. They cannot disagree; they are one map.

Live census: 193 active agents; **3** set `ai_service.max_tokens` (diagnose-agent
32000, feed-triage 4000, site-adoption-agent 16000), **10** set agent-level. All
13 go through the canonical path, so all 13 are unaffected.

⚠ **Not claiming existing config will "self-heal"** — `WRONG_CALLS.md:1262`
records that exact claim being made from a count and refuted by the query
(bugs_open/009 said 17 agents / 10 self-healing; live returned 16 and **zero**).
Queried first. Answer: one site off the canonical path names the key, and its
value makes the request identical.

**One honest divergence:** `ai_actions.go:361` reads `.(float64)` only; the new
reader also accepts `int`/`int64`/`json.Number`, because the key arrives as
float64 from jsonb and **int from YAML** — which is exactly what
`configs/reasoning-agent.yaml` carries. Impossible from jsonb, so it cannot bite
the canonical path; stated rather than left to be discovered.

### §4's `[UNMEASURED]` QUESTION: ANSWERED — `reasoning-agent` is NOT a live burn

Measured by the route §4 itself prescribed (**the pod, not `llm_call_log`**):

- 3 replicas, all `1/1 Running`.
- **Entire log is startup.** Zero request-handling lines, zero `TruncatedError`.
- `grep -rn "system.agent.reasoning.requests"` → the consumer, plus **one** hit:
  `core-manager/admin/system_handlers.go:132`, a hardcoded *"known topics"* list
  in an admin endpoint. **No producer exists in the codebase.**
- `llm_call_log` for `%reason%` → 0 rows (uninformative alone, per §2.3;
  consistent).

**A deployed service consuming a topic nothing writes to.** The hard pin was
real; the burn is zero. **Do not record this as an outage.** Path A makes the key
live; its VALUE is deliberately left at 2048, because raising a budget with no
truncation evidence is what §2.2 warns against.

*Side finding, recorded not acted on:* that config also pins `model:
"claude-3-opus-20240229"`, passed through unchanged by `ResolveModelAlias`. Costs
nothing while there is no traffic; whoever wires up a producer should revisit it.

### Tests — mutation-proven, and one mutation found a real defect in my own test

`platform/aiservice/max_tokens_test.go`. Wire-shape assertions against the
request BODY, clients built through the REAL constructors with only the transport
swapped. Negative controls per provider (unconfigured anthropic/gemini must still
send 2048; unconfigured ollama must send **no** `num_predict`), so "always send
N" fails.

**11 mutations.** M1-M8, M11 killed. **M9 SURVIVED** — deleting the `> 0` guard
left `TestUnusableConfiguredValuesFallBackRatherThanBeingSent` passing, because
the wire-level floor sits in **SERIES** with the resolver and rescued the value.
The test was passing on the floor while naming the guard. Fixed by asserting
`configMaxTokens` **directly** on its `(int, bool)` pair. Logged in
`WRONG_CALLS.md`. **M10 is an equivalent mutant** (Go returns the zero value for
a nil-map read) and is documented as one in the source.

Candidate 4 shipped as `TestEveryImplementedProviderResolvesItsBudgetFromConfig`
— behavioural, not the source-scan lint §5 suggested, because register **OPP-003**
records that shape examining zero files and printing a clean result. It reads
`factory.go`'s `case "x":` lines and fails if a provider both constructs and has
no budget-contract test, and fails if the scan matches nothing.

### Candidate 2 deliberately NOT done — with the reasons, so nobody rediscovers them

Four seats asked for `ai_actions.go:358-364` to become a call to
`llmOptionsFromConfig`. Not bundled here:

- Path A removes the **harm** the duplication caused. Forgetting both copies used
  to cost a silent 2048; it now costs an inconsistency.
- **The copies are not equivalent.** `ai_actions.go`'s outer key is `agentConfig`
  (the *agent's* whole config); the helper's is the *step's*. The helper requires
  `> 0` (so `max_tokens: 0`, today a 400, would become a warning) and folds in
  `budget_tokens`, whose precedence at `:377` is `aiServiceConfig`-only. Three
  behaviour changes on a path serving **127 live steps across 55 agents**.
- That is the trade `llm_options.go`'s own SCOPE note refused, for the same reason.

**`llm_options.go`'s header comment was corrected in place and dated** — this
change makes its "silently gets the provider's hardcoded fallback whatever the
config says" sentence FALSE, and a false explanation beside working code is how
folklore starts. The correction also states why the helper is **still** the right
thing to call (step config; `budget_tokens`), so it is not read as obsolete.

### To close (fixed AND live bar)

1. Rides the next whole-fleet release (owner runs `make release`). **Go code is
   inert until an image rolls.**
2. **`reasoning-agent` is its own service and image** — the chassis roll does not
   carry it. Its `ai_service.max_tokens` stays inert until *it* is rebuilt.
3. Post-roll, per service (one image serves many deployments — read the stamp of
   the service you mean): `kubectl -n ai-persona-system logs -l app=<svc>
   --tail=300 | grep -m1 'build provenance'`, then
   `git merge-base --is-ancestor <commit> <stamped sha>`.
4. **Behavioural proof, not a marker grep**: build a client from a config with
   `max_tokens` well above 2048, call it with nil options, and assert
   `output_tokens > 2048` is *reachable*. The config value is not the
   measurement; the response is.
5. Read the council verdict at correlation `85971036-...` before recording this as
   reviewed.

### ⚠ CORRECTION 2026-08-16 — the first council submission was REFUSED, not reviewed

The section above names council correlation `85971036-5b9a-42ca-aeac-d77a65397c86`. **That run never
reviewed anything.** It reached `current_step='complete_invalid'`, `status='COMPLETED'`, with no
council report: two of its eight edits named TWO files in one `.file` field
(`"platform/aiservice/anthropic.go + gemini.go"`), and `editProblems()` in
`diagnose_persist_fix_plan_action.go` refuses any path containing whitespace, a leading `/`, or `..`.

**LIVE CORRELATION IS NOW `366efae9-f2a8-46ae-a6a6-b9d68a4cc6ae`** — same plan, restructured to one
file per edit (two edits merged because they touched the same test file; no content dropped).

⚠ **Commits `5cafe18ef`, `171121579`, `665c6773a` carry the DEAD correlation in their
`Council-Submitted:` trailer**, so `098` will never credit them. Forward-only forbids an amend; a
follow-up commit names the live correlation. **Do not read those three commits as un-reviewed** —
read this line.

⚠ **A refusal records NOTHING readable**: `diagnosis_artifacts` for that correlation → 0 rows of any
kind, `collected_data` → only `input_data`, `__step_error` → empty (the gate's `persist_submission`
sets no `repair_step`). The cause was established by re-running the validator's own predicate over the
submission file, not read from any message. Landmine filed; `097` gained a fifth client-side trap so
the next thread cannot lose 30 minutes to it. Full account: the lane's NOTES, misstep 3, and
`WRONG_CALLS.md`.

### Still OPEN after Path A — two items, both raised by the council and both needing a human call

Recorded here rather than only in the lane docs, because this file is where the next thread looks.

**1. Candidate 2 — the duplicated precedence rule (unchanged from §5, still open).** Path A did NOT
merge `ai_actions.go:358-364` with `llmOptionsFromConfig`. What it removed is the *consequence* of them
disagreeing — a silent 2048 — not the duplication. The `reuse_agent` seat objected (medium) that Path A
adds a THIRD reader. Answered structurally, and the answer constrains any future attempt:

> Both existing resolvers live in `package actions`, and `platform/orchestration/actions` **imports**
> `platform/aiservice` (`ai_actions.go:13`). So `aiservice` importing either is a genuine **import
> cycle**. Unification cannot go in that direction. It has to be either (a) a new leaf package both
> import, or (b) `datahelpers` — which does not cycle, but `go list -deps ./platform/aiservice` returns
> **only itself** today, so that gives the estate's lowest-level provider package its first dependency
> on the orchestration layer.

The three behavioural differences that make a naive merge unsafe are in the §Path A section above.

**2. `llm_call_log` blindness — §2.3's mechanism is still completely unaddressed, and Path A does not
touch it.** Raised by the `editquality` seat as the plan's one MISSING item, and it is right:

> *"No edit adds any llm_call_log write (or equivalent observability) for calls that bypass
> ExecuteAIStepAction; direct callers remain invisible to the fleet's truncation instrument even after
> their budget is now correctly resolved."*

**Why it was not fixed here, structurally rather than as an excuse:** the provider clients hold
`apiKey`, `model`, `maxTokens` and an `http.Client` and **no database handle of any kind** — a grep for
`sql.`/`pgx`/`DB` across `platform/aiservice` returns one comment and nothing else. A bypassing caller
has no DB either. The write has to happen at the **caller's** layer, so it is a different change with a
different blast radius.

⚠ **This matters more now, not less.** Before Path A, a direct caller was invisible *and* pinned to
2048 — so at least the number was predictable. Now it is invisible *and* running at whatever its
construction config says. **Every instrument this estate has built for truncation reads
`llm_call_log`** — `fleet-step-token-pressure` (LCO-007), the 205 census, the 183 monitoring — so all
of them are structurally blind to exactly this population. What Path A does keep honest: the clients
still write `__sent_max_tokens` back into the options map, the sole feed for `llm_call_log.max_tokens`,
so the column is correct for the budgets Path A newly enables (asserted by
`TestSentMaxTokensRecordsAConfigSourcedBudget`) — *for any caller that persists at all*.

**Owner call needed on both**: in scope for this bug, or separate lanes?
