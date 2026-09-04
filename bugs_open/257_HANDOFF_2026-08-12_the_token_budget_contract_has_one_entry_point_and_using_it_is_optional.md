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

### ✅ COUNCIL APPROVED — round 2, corr `366efae9-f2a8-46ae-a6a6-b9d68a4cc6ae` (2026-08-16)

*Verdict READ before this line was written.* `decision: approved`, `decided_by: "approved with 3
advisory objection(s) — none high-severity"`, **`gated_by_truncation: false`** (checked deliberately —
the architecture seat has a recorded history of truncated reviews), 4 seats abstained.

Round 1 on the same trail was **REVISE**, gated by `reuse_agent`. Two of its three substantive
objections were right and produced real work: a reuse check I had genuinely skipped, and vision-path
coverage I had not thought about. Full account in the lane NOTES.

**The three advisory residuals, and what was done about each:**

1. **`editquality` (low, edit 7)** — the `reasoning-agent.yaml` edit is doc-only like edits 6 and 8,
   but was not grouped under my own "accepted as comment-only, kept deliberately" concession, so it
   *reads* as a functional fix. **Accepted; it is a fair honesty point.** The value is unchanged at
   2048 and the rationale said so, but the grouping should have made that obvious without reading.
2. **`bug_historian` (low, edit 5)** — candidate 4's provider scan guards only providers reachable
   from `factory.go`'s `case` list; a client constructed outside the factory would not be caught.
   **MEASURED and mostly closed:** five production sites bypass the factory
   (`rag_actions.go:374`, `reasoning/agent.go:71`, `tools-api` defend/gripper/position) — and all five
   are covered anyway, because the per-provider tests call `NewAnthropicClient`/`NewGeminiClient`/
   `NewOllamaClient` **directly** rather than through the factory. The genuine residual is narrower
   than it sounds: a FUTURE provider whose constructor exists but is never added to the factory switch.
   Recorded in the test's own comment.
3. **`bug_historian` (low, edit 2)** — struct-literal constructions outside test files were never
   enumerated the way the `GenerateText` call sites were. **MEASURED and closed:**
   `grep -rn '&(AnthropicClient|GeminiClient|OllamaClient){' --include=*.go` minus tests returns
   **exactly three hits — `anthropic.go:69`, `gemini.go:186`, `ollama.go:62` — and all three ARE the
   constructors themselves.** No production code builds these clients as a bare struct. The floor's
   remaining population is the 7 test fakes across 5 files, which is what it was written for.

**Also noted by the seat and worth carrying:** `bugs_closed/076`
(*truncated_llm_responses_tolerated_at_113_unguarded_call_sites*) was not cited in `grounded_in`
despite being the closest title in the family — along with `bugs_closed/012` and `046`, it is the
silent-truncation lineage this bug belongs to.

⚠ **The commits carry `Council-Submitted:`, not `Council-Reviewed:`** — deliberately, and correctly:
they were made before the verdict, forward-only forbids an amend, and `098` resolves the correlation at
report time, so they are credited automatically now that it is approved. **Do not read them as
un-reviewed.** ⚠ Commits `5cafe18ef`, `171121579`, `665c6773a` carry the earlier DEAD correlation
`85971036` (refused unreviewed); `3f918bd34` records the supersession.

## ✅ LIVE 2026-08-16 at v1.0.1305 — verified at the artefact, with controls; and a CORRECTION to my own census

Fleet rolled 22:07:55Z. **Both target images shipped** — the chassis roll carried
`reasoning-agent:v1.0.1305` too, so the "needs its OWN image" caveat above is satisfied.

### Deploy proven at the binary, not at git or the tag

| check | result |
|---|---|
| `reasoning-agent` build stamp (from its own log) | `6a782274b626c9f4977c9246d905deebb097cb1f` |
| `git merge-base --is-ancestor 5cafe18ef 6a782274b` | **YES** — the fix is in the deployed build |
| `agent-chassis` binary probe, positive control | expected sha **present** in `/proc/1/exe` |
| `agent-chassis` binary probe, **negative control** | fabricated sha **absent** — the probe discriminates |
| deployed `configs/reasoning-agent.yaml` carries the 257 annotation | **PRESENT**, with a fabricated-string control **absent** |

⚠ The chassis' `build provenance` line had already **scrolled out of `--tail=400`** — the documented
landmine. The binary probe has no shelf life and is what settled it.

### `reasoning-agent` survived reading its own config for the first time

3 pods `Running`, **0 restarts**, 0 error/panic/fatal lines. Worth checking rather than assuming: this
is the first build in the service's life where that YAML `max_tokens` is actually parsed (as an `int`,
which is exactly the type the pre-existing `float64`-only reader would have dropped).

### The central claim tested live: "this alters zero requests in the fleet"

`llm_call_log`, 12h before the roll vs since:

| | value |
|---|---|
| calls in the 12h before | 584 |
| **calls since the roll (DEMAND CONTROL — the zero is not blind)** | **36** across 5 distinct steps |
| **steps whose `max_tokens` changed across the boundary** | **0** |
| truncations before / after (`error_message ILIKE '%stop_reason=max_tokens%'`, the only correct predicate) | 0 / 0 |

⚠ **This is an EARLY, PARTIAL sample and must not be read as fleet-wide proof** — 5 steps have run
since the roll, out of a much larger live population. It is a real check with a real demand control,
not a complete one. Re-run it after a day of normal load.

### ⚠ CORRECTION — my census in the §Path A section is UNDERSTATED 5×, and the live data caught it

The §Path A section says *"3 set `ai_service.max_tokens` … 10 set agent-level … all 13 go through the
canonical path"*. **The true figure is 68 distinct agents.**

`llm_call_log` showed `feed-triage / score_relevance` sending **8192** where my census recorded
feed-triage at **4000**. The reason: root `ai_service.max_tokens = 4000` AND a **step-level**
`workflow.steps.score_relevance.config.ai_service.max_tokens = 8192`, which `resolveAIServiceConfig`
overlays over the root (`ai_actions.go:77-87`). **My census only ever read the root block.**

Depth-aware, run 2026-08-16 post-roll: **193** active agents · **3** root `ai_service.max_tokens` ·
**10** agent-level `max_tokens` · **67** step-level `config.ai_service.max_tokens` · **68** distinct
agents once deduplicated.

**The conclusion is unchanged, and precisely why matters:** the blast-radius argument never rested on
the count. It rests on `createAIClient` (`:287`) and the options build (`:361`) reading the **same
merged map** from `resolveAIServiceConfig` (`:243`) — and that merge already *includes* the step-level
overlay. So client-default and options cannot disagree at any nesting level, for 13 agents or 68. The
number was decorating an argument that did not need it, which is exactly why nothing failed when it was
wrong. Logged in `WRONG_CALLS.md`.

### What is verified, and what is NOT — stated so the next reader does not over-read this

- **VERIFIED:** the fix is in the deployed binaries (both images, with controls); the config half
  shipped; `reasoning-agent` is healthy; the canonical path is unchanged on live traffic so far.
- **NOT VERIFIED LIVE, and it has no live exerciser:** the distinguishing behaviour — *a nil-options
  caller whose config exceeds 2048 actually sending >2048*. The only nil-options caller in the estate is
  `reasoning-agent`, its budget is deliberately 2048, and **nothing publishes to its topic**. Its config
  is baked into the image (no ConfigMap volume), so manufacturing this case needs a yaml edit plus a
  rebuild, and would be a synthetic exerciser rather than a real one. The behaviour is proven
  deterministically instead by the mutation-tested wire-shape tests, which assert the request BODY.
  **Do not write "behaviourally verified end to end" in this file without doing that work.**

---

## §2026-09-03 RE-VALIDATED — still OPEN, and the class has RETURNED in a shape that defeats Path A

*This section by the session that picked the bug up on 2026-09-03 after 18 days quiet. It wrote **no
code**; the fleet build of that day carries nothing from it. Cold-start for the next thread:
`docs/agent_docs/docs024_key_docs_latest/bugfix_257_token_budget_at_the_client/HANDOFF_2026-09-03_continue_here.md`*

**The bug is still valid.** Path A remains live and correct (MDL-041, v1.0.1305). Both items listed as
"still open after Path A" are untouched. What is new is that the defect class reproduced itself twice in
code written *after* the fix, exactly as the `architecture` seat said it would — *"any action author who
reaches for `client.GenerateText` instead of `ExecuteAIStepAction` reproduces this bug by
construction"*.

### The new sub-class: a hand-rolled literal fallback

| file:line | added | what it does |
|---|---|---|
| `rewrite_negations_action.go:537-543` | 2026-08-20, `a5d5d0728` | reads `aiServiceConfig["max_tokens"]` as `float64` only, `else options["max_tokens"] = 2000` |
| `repair_ordering_register_action.go:436-440` | 2026-08-31, `f7156fb54` | identical block, identical literal |

1. **It defeats Path A.** Path A resolves the budget at the client *when the caller supplies none*.
   These callers always supply one, so rule 1 wins at the wire (`anthropic.go:307`) and the client's
   resolution never runs. Passing a literal is now strictly worse than passing `nil`.
2. **Copies of the precedence rule have gone from three to FIVE** — `ai_actions.go:357-360`,
   `llmOptionsFromConfig`, `aiservice/max_tokens.go`, plus these two. Neither new copy reads the
   **step's** config and neither forwards `budget_tokens`, so both are less capable than the helper in
   their own Go package. `llm_options.go`'s SCOPE note predicted precisely this ("a THIRD caller should
   be the extraction, not another paste") and was ignored twice.
3. **⚠ THE LITERAL EQUALS THE CONFIGURED VALUE, SO `llm_call_log` CANNOT DETECT THE DEFECT.**
   `offer-analyser` declares `ai_service.max_tokens = 2000` on both steps running
   `repair_ordering_register`; the Go fallback is also 2000. [MEASURED 2026-09-03] 29 calls, all at
   `max_tokens=2000`, 0 truncations, max output 600. **That reading is equally consistent with the code
   working and with 257 being live at that call site.** The original defect showed a *wrong* number;
   this one shows the *right* number for a possibly wrong reason, which is worse.

### The worked example, and a near-miss it caused

[MEASURED 2026-09-03] every `rewrite_negations` call ever logged: **99 at 2000** (2026-08-21 → 08-23,
**4 truncated**), **3,245 at 16000** (08-23 → today, 0 truncated).

Those four truncations are **NOT** the Go literal biting — I nearly recorded that they were. Migration
`517` set the step's config to 2000 on 08-21 with a stated rationale; migration `569` raised it to 16000
on 08-23 after the 305 lane measured a ten-target page repairing zero. The configured value and the Go
literal were **the same number** for those three days, so no query could have separated the hypotheses.
Settled by reading the two migrations. Logged in `WRONG_CALLS.md`.

**The residual is real:** 569's repair lives only in configuration. Line 542 still says 2000, so any
future step or loop body running this action without declaring the key silently inherits 2000 and 569's
measured fix does not transfer. Migration 517 documents how easily that happens — a loop substep is
named `process_sections_loop_iter_N_rewrite_negations`, which `resolveAIServiceConfig` cannot resolve at
all.

### §4's census re-run — it has now been wrong THREE times, do not quote it

[MEASURED 2026-09-03] **twelve** production call sites, not nine. §4 missed the Gemini client, missed a
tenth site that arrived within days of filing, and predates these two. Grouping and the grep are in the
handoff §4. Supporting live config census, 208 active agents: **149** steps declare
`config.ai_service.max_tokens`, 10 agent-level, 3 root, **7** declare a top-level `config.max_tokens`
outside `ai_service` (possibly a sixth spelling nobody reads), **0** declare a value `<= 0`.

### The fix plan, unsubmitted

In the handoff, §5. Summary: replace both hand-rolled blocks with `llmOptionsFromConfig` (three named
behaviour changes, all measured inert on today's fleet); then generalise
`TestNoProvocationActionCallsAModelWithAnEmptyOptionsMap` from two named files to the whole package, so
the sixth paste cannot merge quietly. ⚠ The test must not land before the fix — on this shared tree it
would break HEAD for every other session — and must not allow-list the two current violations.

**Candidate 2 (merging `ai_actions.go` with the helper) stays open and stays a human call**; the import
cycle constraint recorded on 2026-08-16 still holds. Step 1 takes the copies from five to three without
answering it.

⚠ **When verifying any of this, pick a step whose configured budget DIFFERS from the literal** — or
§(3) above applies to your own check and it cannot fail.

---

## §2026-09-03b — ROUND 2 FIX COMMITTED (`51357cf51`), guard bound at the package, three MORE sites found

Written by the session that implemented the plan in the section above. **Still OPEN**: the code is
committed but not yet rolled, and this file's own bar is *fixed AND live*.

**Council:** submission `c8660cfb-690d-4dd2-8b1f-25828305133e`, committed with `Council-Submitted:`.
Submission JSON:
`docs/agent_docs/docs024_key_docs_latest/bugfix_257_token_budget_at_the_client/COUNCIL_SUBMISSION_2026-09-03_257_round2_no_hardcoded_budgets.json`

### What shipped

All five direct model callers in `platform/orchestration/actions` now resolve their budget through
`llmOptionsFromConfig`, and no numeric literal reaches a budget key anywhere in the package.

| file | before | after |
|---|---|---|
| `rewrite_negations_action.go` | hand-rolled block + literal `2000` | resolver |
| `repair_ordering_register_action.go` | hand-rolled block + literal `2000` | resolver |
| `companies_house_llm_review_action.go` | `map[string]interface{}{}` — an EMPTY map | resolver |
| `execute_vision_prompt_action.go` | float64-only read, **no `> 0` guard** | resolver |
| `feed_actions.go` (`fetchViaPerplexity`) | `"max_tokens": 4096` in a raw HTTP body | `source_config["max_tokens"]`, 4096 as default |

New: `platform/orchestration/actions/llm_budget_call_sites_test.go` (package-wide AST audit, three
tests). Corrected in place: `llm_options.go`'s SCOPE note, whose prediction *"a THIRD caller should be
the extraction, not another paste"* was ignored twice.

### ⚠ FOURTH CORRECTION TO §4's CENSUS — and this time the census was not merely stale, it was WRONG

The section above says the census "has now been wrong THREE times". Four. And the two new errors are a
different kind from the first three, which were staleness by addition:

1. **`companies_house_llm_review_action.go` was classified "reads config, no literal — acceptable".** It
   does not read config for the budget at all; it passed an **empty options map**. Since Path A that is
   not a burn — the client resolves from its construction config — but the step's own `max_tokens` and
   `budget_tokens` never reached the wire, and no client constructor is ever shown a step config.
2. **`execute_vision_prompt_action.go` was classified the same way.** It read the key float64-only with
   **no `> 0` guard**, so a configured `max_tokens: 0` would be sent verbatim — a hard 400, per
   `platform/aiservice/max_tokens.go`. Latent ([MEASURED 2026-09-03] no active agent declares a
   non-positive budget), but recorded as acceptable when it was not.
3. **`feed_actions.go:607` was never in the census at all, in any of its four runs.** `fetchViaPerplexity`
   builds a raw HTTP chat-completions body and never touches `platform/aiservice`. **Every census this
   bug has run greps `\.GenerateText(`, so all four were structurally blind to it.** It carried
   `"max_tokens": 4096`.

**The transferable point** (now a `LANDMINES.md` entry): a census of "who hardcodes a budget" keyed on
the client interface can only see callers who use the client. A provider reached over raw HTTP is
invisible to it and the census reads complete. The AST audit found it on its first run because it asks
"is a numeric literal written to a budget key anywhere in this package", which is transport-agnostic.

### The guard, and why it is not an allow-list

- **AST, not source text** (`go/parser` asked for no comments), so this package's prose cannot become a
  needle for its own test.
- **`ai_actions.go` is the ONE exemption** — it *is* the canonical resolver, and cannot call the helper
  (import cycle + agent-level vs step-level precedence: candidate 2). `TestTheCanonicalBuilderStillResolvesFromConfig`
  asserts it still assigns a budget from a non-literal, so the exemption cannot outlive its reason.
- **The two live violations were fixed in the same commit, never allow-listed.**
- **Non-vacuity counters** `t.Fatal` on zero model calls or zero budget writes. Today: `audited 10 model
  call sites`, `audited 8 budget-key writes`.
- **[MEASURED 2026-09-03] Mutation-proven, four mutations in an isolated HEAD checkout** (never in the
  shared tree): reinstating the literal → FAIL at `rewrite_negations_action.go:553`; hand-rolling the
  map → FAIL *"called from runRegisterRepair, which never calls llmOptionsFromConfig"*; passing `nil` →
  FAIL *"is handed a nil options map"*; renaming the exempt file → FAIL on all four `ai_actions.go`
  sites plus the exemption's own repoint message.

### Blast radius — measured, not asserted: ZERO requests change today

[MEASURED 2026-09-03, live `agent_definitions`, 208 active] Every step reached by these five call sites
declares its budget under `ai_service` and none at step top level; no active agent declares
`budget_tokens` anywhere. `offer-analyser.repair_ordering_register` 2000 ·
`offer-analyser.repair_hierarchy_register` 2000 · `page-content-writer.rewrite_negations` 16000 ·
`design-critique-agent.critique` 16000 · `tool-acceptance-agent.look` 4000 · `ch-llm-reviewer.review`
2000. So the helper's two extra capabilities (step-level precedence, `budget_tokens` forwarding) are
inert on today's fleet, and each step sends exactly the number it sends now.

**Pinned pre-change baseline from `llm_call_log`, for the post-roll comparison:**
`repair_ordering_register` 29 @2000 (max out 600) · `repair_hierarchy_register` 6 @2000 ·
`rewrite_negations` (loop substeps) 99 @2000 with 4 truncations 2026-08-21..23, then 3,245 @16000 with
none · `critique` 5 @6000 + 1 @16000 · `look` 91 @4000 · `ch_llm_review` no rows.

**The one residual behaviour change, stated:** an UNCONFIGURED call at these sites now reaches
`aiservice.DefaultMaxOutputTokens` (2048) via Path A instead of a private `2000`. No live step is
unconfigured, so this changes nothing today; it is the right direction regardless, because 2048 is the
estate's single documented floor and 2000 was one action's opinion.

### §4's OTHER open question — ANSWERED, and it splits

[MEASURED 2026-09-03] The **7** steps declaring a top-level `config.max_tokens`:

- **`html-developer-chunked` x3 ARE read** — `getMaxTokens(config, 16000)` at `html_actions.go:27`,
  which then synthesises an entire `ai_service` block (hardcoding model and provider too — a separate
  smell, not touched).
- **`site-adoption-agent` x4 are DEAD CONFIG.** `analyze_site` 32000, `derive_content_direction` 6000,
  `classify_archetype` 4000, `generate_design_intent` 4000 — and **every logged call from all four runs
  at 16000**, the root `ai_service` value, because `ai_actions.go:357` reads the AGENT's config then
  `ai_service` and nothing looks at a step's top level. An operator asked for double on `analyze_site`
  and got half. Max observed output 7,708, so nothing truncates today.

**NOT changed here.** Moving those four under `ai_service` is live the instant it applies and raises
spend on those steps. Owner's call — this is the original 257 defect from the configuration end, and it
wants its own decision, not a side effect of a code commit.

### To close this bug

Unchanged bar: **fixed AND live**. Needs a fleet roll, then §6's verification route — prove it at the
binary with a negative control, then show a step whose declared budget **differs from 2000** sending the
larger number from `llm_call_log.max_tokens`. ⚠ Picking a 2000 step for that check reproduces §(3)'s
blindness in your own verification and it cannot fail. `page-content-writer.rewrite_negations` at 16000
is the right subject.

**Still open and still human calls:** candidate 2 (merging `ai_actions.go`'s inline block with the
helper — import cycle, different precedence level), the `llm_call_log` blindness of direct callers
(handoff §5 step 3), and now the four dead `site-adoption-agent` declarations above.

### ✅ COUNCIL APPROVED — round 2, corr `c8660cfb-690d-4dd2-8b1f-25828305133e` (2026-09-03)

`decided_by: "all reviewers approve"`, `gated_by_truncation: false`, no high-severity objection.
Dispatched 16:21, decided 16:37 — **16 minutes**. Two objections were acted on in commit `54ee2d261`
(`Council-Reviewed:`), not merely noted:

- **`reuse_agent` MEDIUM** — the new package-wide audit coexists with the older two-file guard with no
  stated consolidation plan. They differ by one assertion (the old one is a per-FILE "did my subject
  move" alarm; its empty-map half is redundant), and both files now say so, so the decision to keep it
  is stated rather than implicit.
- **`guardian` MEDIUM** — verify Path A rather than assume it, given a landmine that contradicts it.
  That landmine's heading is half false and its first bullet already says so; the seat read the heading.
  Verified: `anthropic.go:75` constructs with `resolvedMaxTokens(config)`, `5cafe18ef` is an ancestor of
  HEAD. **And the blast radius does not depend on Path A** — every live step here declares a budget, so
  the helper always supplies an explicit `max_tokens`; Path A governs only an unconfigured caller.
- **`guardian` LOW** — `fetchViaPerplexity` callers enumerated: exactly one, `feed_actions.go:419`,
  updated in the same commit. It should have been in the submission.

---

## ✅ ROUND 2 IS LIVE on `v1.0.1360` (rolled 2026-09-03 22:06:29Z) — proven, and with the one thing the proof CANNOT do stated

**The bug stays OPEN.** Two named residuals are untouched (candidate 2; the `llm_call_log` blindness of
direct callers), plus the four dead `site-adoption-agent` declarations found in round 2. The *round 2
defect* — hand-rolled budget literals — is fixed and live.

### How it was proven, and why not the way §6 prescribes

⚠ **This change is BEHAVIOUR-NEUTRAL BY CONSTRUCTION, so no post-roll reading can distinguish the new
code from the old.** That is not a gap in the verification; it is the same §2(c) blindness one level up,
now applying to my own check. Every live step reached by the five call sites declares its budget under
`ai_service`, so the helper supplies the same explicit number the hand-rolled block did. §6's prescribed
behavioural proof — *"a step whose declared budget exceeds the old literal actually sends the larger
number"* — cannot discriminate here either: `page-content-writer` sent 16000 **before** the roll too,
because its hand-rolled block read the config correctly. **Do not record the 16000 as evidence the fix
shipped.** It is evidence that nothing regressed, which is a different claim.

**What actually proves it, at the binary, by ancestry:**

- `PRC-003` (`bugs_open/453` shape 3), carried by **`681b0ee65`**, was proven present in the
  **v1.0.1360** binary by the 453 lane on 2026-09-04, through one `tr '\0' '\n' | grep -Fc` pipeline
  over `/proc/1/exe` with **four controls** — two new literals present, **the OLD literal
  `TEMPLATE RENDERED WITH MISSING DATA` ABSENT**, a must-be-present control present, a must-be-absent
  control absent. (Register `PRC-003`, `docs026_concept_register/register/prompt-composition.md`.)
- `git merge-base --is-ancestor 51357cf51 681b0ee65` → **passes**. My fix (17:21:53+01:00) is an
  ancestor of that commit (17:33:01+01:00), so any build containing `681b0ee65` contains `51357cf51`.
- Therefore **v1.0.1360 carries round 2**. The evidence is another lane's binary probe plus an ancestry
  check, not my own probe — stated plainly so nobody re-reads it as a first-hand measurement.
- ⚠ `54ee2d261` (the council-objection answers) is **not** an ancestor of `681b0ee65`. It is test-file
  comments only, so it cannot affect the binary; whether the build commit also includes it is unknown
  and immaterial.

**My own attempt at a direct probe FAILED and the failure is worth carrying:** a single
`grep -aoFf <474 candidate shas> /proc/1/exe` over the whole binary returned **empty after ~110s**, i.e.
it was killed, not answered. An empty result there reads exactly like *"none of these commits is in this
image"* — which would have been a false negative pointing the opposite way. The `dispatch_throughput`
lane had already recorded that a full-binary scan overruns a two-minute call (their NOTES, 2026-09-04).
**A binary probe must be a small number of KNOWN values, never an enumeration.**

### Behaviour after the roll — the "nothing regressed" half [MEASURED 2026-09-04]

Every step reached by the five changed call sites, since 22:06:58Z:

| agent | step | sent | calls | max out | truncated |
|---|---|---|---|---|---|
| offer-analyser | repair_hierarchy_register | 2000 | 4 | 125 | 0 |
| offer-analyser | repair_ordering_register | 2000 | 14 | 706 | 0 |
| page-content-writer | rewrite_negations (loop iters 0–4) | 16000 | 184 | 5,937 | 0 |
| tool-acceptance-agent | look | 4000 | 8 | — | 0 |

208 post-roll calls, every one at its configured value, zero truncations (counted from
`error_message ILIKE '%stop_reason=max_tokens%'`, never from `output_tokens >= max_tokens`).
`ch-llm-reviewer` and `design-critique-agent.critique` had no post-roll traffic.

### The one live check that WOULD discriminate — not run, because it writes to production config

Set a **top-level** `config.max_tokens` (outside `ai_service`) on a step reached by the new code and
read the next `llm_call_log.max_tokens`:

- **old code** ignores step-level entirely → sends the `ai_service` value;
- **new code** goes through `llmOptionsFromConfig`, whose documented precedence is *step config beats
  `ai_service`* → sends the step-level value.

`page-content-writer.rewrite_negations` with `config.max_tokens = 15999` against its `ai_service` 16000
is the cheapest form: one jsonb write, one token of difference, immediately reversible, and the two
hypotheses give different numbers. **It is a live config change to a production agent, so it is an
owner decision, not a session's.** Recorded here so the next session does not have to re-derive it —
and note it would also be the first live exercise of the step-level precedence arm, which
[MEASURED 2026-09-03] nothing on the fleet uses today.
