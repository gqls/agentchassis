# PLAN 2026-08-16 — bugs_open/257: move the token budget INTO the provider client

**Lane opened 2026-08-16.** Bug: `bugs_open/257_HANDOFF_2026-08-05...` — *"the
token-budget contract has exactly ONE enforced entry point and using it is
OPTIONAL, so every direct-to-provider call silently runs at 2048."*

Ownership checked before starting: `scripts/who-owns.py 257` finds **one commit**
(the filing, `406595601`, 2026-08-12) and **no owning workstream**; a grep of all
live session transcripts (`~/.claude/projects/.../*.jsonl`, last 4h) shows no
session working 257. The bug file itself says *"Status: OPEN, unowned."*

## 1. The bug is still valid — re-verified at HEAD, and the line numbers have MOVED

The filing quotes `anthropic.go:109` and `:147`. Both have shifted; the code has
not. Re-read at HEAD 2026-08-16:

| claim | filed as | verified at HEAD | still true? |
|---|---|---|---|
| provider hardcodes 2048 | `anthropic.go:109` | **`anthropic.go:185`** `"max_tokens": 2048` | YES |
| only `options` can override | `anthropic.go:147` | **`anthropic.go:280`** `if maxTokens, ok := options["max_tokens"]` | YES |
| inline precedence copy | `ai_actions.go:358-364` | **`ai_actions.go:358-364`** (unmoved) | YES |
| `reasoning-agent` passes `nil` | `reasoning/agent.go:127` | **`reasoning/agent.go:127`** `GenerateText(a.ctx, prompt, nil)` | YES |
| `reasoning` has zero `max_tokens` | — | `grep -rn max_tokens internal/agents/reasoning/` → **0 hits** | YES |

**Gemini has the same hardcoded 2048** (`gemini.go:308`, `visibleBudget := 2048`)
— the filing did not mention it, because it read only the Anthropic client.

## 2. Census re-run at HEAD — it has GROWN by one since filing

`grep -rn --include=*.go -E '\.(GenerateText|GenerateWithImages)\(' . | grep -v ^./platform/aiservice/ | grep -v _test.go`

| call site | options | budget reaching the API |
|---|---|---|
| `ai_actions.go:415,624,764` | `options` built at `:358` | correct — the canonical path |
| `provocation_gate_action.go:842` | `opts` | fixed 2026-08-10 |
| `provocation_generator_action.go:548` | `opts` | fixed 2026-08-10 |
| `companies_house_llm_review_action.go:133` | `llmOptions` | sets `max_tokens` — OK |
| `execute_vision_prompt_action.go:197` | `options` | sets `max_tokens` — OK |
| `contentcreator/agent.go:305` | `options` | sets `max_tokens` (default 3000, `:720`) — OK |
| `tools-api/handlers/defend.go:101` | literal `{"max_tokens": 8192}` | hardcoded, config-blind |
| `tools-api/handlers/position.go:90` | literal `{"max_tokens": 8192}` | same |
| **`tools-api/handlers/gripper.go:161`** | `{"max_tokens": gripper.MaxTokensPerReply}` | **NOT IN THE FILED CENSUS — a new call site** |
| **`reasoning/agent.go:127`** | **`nil`** | **2048, always** |

`llm_options.go`'s own comment says *"a THIRD caller should be the extraction, not
another paste"*. A third caller has since arrived (`gripper.go`). It happens to
pass a Go constant rather than reading config, so it is not harmed — but the
prediction held, in under a week, which is the point.

## 3. The design question the filing did NOT ask: where does the client get config?

Every provider constructor already receives **the whole `ai_service` config map**
and already reads `model` out of it with a default fallback
(`anthropic.go:47-53`). It simply **discards `max_tokens`**. The struct keeps
`apiKey`, `model`, `httpClient` and nothing else.

So candidate 1 is not new plumbing — it is *reading one more key from a map the
client is already holding*, in exactly the shape `model` already uses.

## 4. DECISION: Path A = candidate 1, in all three provider clients

New precedence at the wire:

1. `options["max_tokens"]` — per-call, **wins** (unchanged)
2. **`ai_service.max_tokens` from the construction config — NEW**
3. provider hardcoded fallback — unchanged (2048 anthropic/gemini; omit
   `num_predict` entirely for ollama, which is today's behaviour)

Parsing accepts `int`, `int64`, `float64` and `json.Number`, and requires `> 0`
— because config arrives as **float64 from jsonb** (`agent_definitions`) but as
**int from YAML** (`configs/*.yaml` via viper), and the two paths must agree.
This is the same widening `intFromConfig` in `llm_options.go` already made.

### Why this is the candidate that closes the door

It removes the *choice*. A future action author who reaches for
`client.GenerateText(ctx, prompt, nil)` — the shape the filing says "reproduces
this bug by construction" — now inherits the configured budget instead of 2048.
Candidates 2 and 4 document or detect the hazard; only this one deletes it.

## 5. Blast radius — MEASURED, not argued (owner ruling 2026-07-28)

**Claim: this change alters ZERO requests in the fleet today.** It converts a
dead config key into a live one. Established by enumerating *every* construction
site, not by reasoning about the common case:

`grep -rn --include=*.go -E 'aiservice\.(NewClient|NewAnthropicClient|NewGeminiClient|NewOllamaClient)\(' . | grep -v _test.go` → **7 sites**

| # | construction site | config carries `max_tokens`? | options carry it? | effect of Path A |
|---|---|---|---|---|
| 1 | `ai_actions.go:1398` (canonical) | sometimes | **whenever the config does** | **none — see below** |
| 2 | `contentcreator/agent.go:123` | no (`configs/content-creator-agent.yaml`) | always (`:720`, default 3000) | none |
| 3 | `reasoning/agent.go:71` | **yes — `max_tokens: 2048`** | **no (`nil`)** | key becomes LIVE; value unchanged at 2048 |
| 4 | `tools-api/gripper.go:63` | no (inline literal map) | always | none |
| 5 | `tools-api/position.go:77` | no (inline literal map) | always | none |
| 6 | `tools-api/defend.go:83` | no (inline literal map) | always | none |
| 7 | `rag_actions.go:374` (ollama) | n/a — **embeddings only** | n/a | none |

**Site 1 is provably unchanged, and this is the load-bearing argument.**
`createAIClient(ctx, aiServiceConfig)` (`:287`) and the options build (`:361`)
read the **same merged map**, produced once by `resolveAIServiceConfig` (`:243`).
So:

- `aiServiceConfig["max_tokens"]` present → options carry it → **rule 1 fires →
  identical value**.
- absent → the client default is absent too → **rule 3 → identical**.
- only agent-level `agentConfig["max_tokens"]` set (outside the `ai_service`
  block, so never passed to the client) → options carry it → **rule 1 →
  identical**.

There is no case where the client default and the options value disagree on the
canonical path, because they come from one map.

**Live census of what is configured** (`agent_definitions`, active, non-snapshot):
193 active agents; **3** set `ai_service.max_tokens` (diagnose-agent 32000,
feed-triage 4000, site-adoption-agent 16000); **10** set agent-level
`max_tokens`. All 13 go through site 1, hence all 13 are unaffected.

### One honest divergence

`ai_actions.go:361` reads `aiServiceConfig["max_tokens"].(float64)` — float64
**only**. Path A's reader also accepts `int`. A config holding an `int` would
therefore start working where it silently did not. This cannot occur from jsonb
(always float64) and so cannot occur on the canonical path today; it is a
behaviour change in the correct direction, and it is stated rather than
discovered later.

## 6. §4's `[UNMEASURED]` question, ANSWERED: reasoning-agent is NOT a live burn

The filing recorded the hard-pinned 2048 on `reasoning-agent` and said plainly
that whether it truncates in production was not established, warning that
`llm_call_log` cannot answer it (that table is written by the very function being
bypassed). Measured today, by the route the filing prescribed — the pod, not the
table:

- **3 replicas running** (`reasoning-agent-75b7fb77d-*`, 5h old at the time of
  checking, all `1/1 Running`).
- **Their entire log is startup.** Zero request-handling lines, zero
  `TruncatedError`, zero AI calls.
- **Nothing publishes to its topic.** `grep -rn "system.agent.reasoning.requests"`
  finds the consumer itself and *one* other hit —
  `core-manager/admin/system_handlers.go:132`, a hardcoded "known topics" list in
  an admin endpoint that lists topic names. **There is no producer in the
  codebase.**
- `SELECT count(*) FROM llm_call_log WHERE agent_type ILIKE '%reason%'` → **0**,
  which on its own is uninformative (§2.3) but is consistent.

**So: a deployed service consuming a topic nothing writes to.** The hard pin is
real and the defect is real; the *burn* is zero. Recorded so nobody re-files it
as an outage, and so the fix is not oversold.

**Consequence for candidate 3:** Path A makes the key live, which is the fix.
**Its VALUE is deliberately left at 2048.** Raising a budget with no evidence of
truncation is precisely what §2.2 of the bug warns against — two live config
changes were made in one day against a key nothing read. There is now a real
lever; pulling it needs a real reason.

## 7. Deliberately OUT of scope: candidate 2 (extracting the duplicated rule)

Four council seats asked for `ai_actions.go:358-364` to be replaced by a call to
`llmOptionsFromConfig`. Not done here, on purpose:

- **Path A removes the HARM the duplication caused.** The cost of a caller using
  neither copy was a silent 2048; after Path A it is the configured value. What
  remains is a consistency cost, not a correctness cliff.
- **The two copies are not equivalent and cannot be naively merged.**
  `ai_actions.go`'s outer key is `agentConfig` (the *agent's* whole config);
  `llmOptionsFromConfig`'s is the *step's*. `llmOptionsFromConfig` also requires
  `> 0` (so `max_tokens: 0`, today a 400 from Anthropic, would become a warning)
  and folds in `budget_tokens`, whose precedence in `ai_actions.go:377` is
  `aiServiceConfig`-only. Each of those is a behaviour change on a hot path
  serving 127 live steps across 55 agents.
- Bundling a semantic refactor of the fleet's busiest code path into a change
  that already touches every LLM call is the trade `llm_options.go`'s own SCOPE
  note refused, for the same reason.

**Recorded as the follow-up**, with the three divergences above named so the next
session does not have to rediscover them.

## 8. Verification plan (§6 of the bug — the disconfirming result must be possible)

- **Wire-shape tests**, in the style of `prompt_caching_test.go`: assert against
  the request BODY via `capturingTransport`, because a client that dropped the
  config default would pass any response-only test.
- **Mutation-proven**: revert each arm and the matching test must fail. A test
  that passes on a binary still ignoring the config is not a check.
- **A negative control in the same run** — a client built with no `max_tokens`
  must still send 2048, or the test proves only that *something* was sent.
- **At the pod, post-roll**: grep a string the change ADDS, plus a control.

---

## Round 2, 2026-09-03 — the plan that was executed, and the corrections to the brief it was based on

This file plans **Path A** (budget resolution at the provider client), which shipped 2026-08-16. Round 2
is a different plan and it was written elsewhere:
`docs/agent_docs/docs024_key_docs_latest/bugfix_257_token_budget_at_the_client/HANDOFF_2026-09-03_continue_here.md`
§5. Recorded here so the standing five stay coherent — a reader of the PLAN should not have to discover
that a second phase happened.

**Executed:** §5 steps 1 and 2, in one commit (`51357cf51`), council APPROVED
(`c8660cfb-690d-4dd2-8b1f-25828305133e`), objections answered in `54ee2d261`. §5 step 3 (`llm_call_log`
blindness of direct callers) is untouched and still open.

**Three corrections to the originating brief, marked as corrections rather than edited away** — the
handoff's §4 census, which §5 step 1 was scoped from:

1. `companies_house_llm_review_action.go` was classified *"reads config, no literal — acceptable"*. It
   passed an **empty options map** and read no budget config at all.
2. `execute_vision_prompt_action.go`, same classification. It read the key **float64-only with no `> 0`
   guard**, so a configured `max_tokens: 0` would be sent verbatim — a hard 400.
3. `feed_actions.go`'s Perplexity path (`"max_tokens": 4096`, raw HTTP) **was not in the census at all,
   in any of its four runs**, because every run greps `\.GenerateText(` and that path never touches
   `platform/aiservice`.

**Consequence for the plan's shape:** step 1 as written was "delete two hand-rolled copies"; what
shipped routes **five** call sites onto the one resolver. That is not scope creep — step 2's own stated
rule (*"every GenerateText/GenerateWithImages call … must pass either options built by the canonical
path or llmOptionsFromConfig, and never a numeric literal"*) makes all five violations by construction,
and a guard cannot land while its own rule is broken in the same package.

**The decision worth carrying:** the three extra sites were found **by writing the enforcement, not by
reading the code**. A census keyed on an interface can only find callers who use it. The generalisation
is in `LANDMINES.md` (*"A budget census keyed on the CLIENT INTERFACE is blind to a provider called over
raw HTTP"*) and the incident in `WRONG_CALLS.md`, both 2026-09-03.

---

## ROUND 3 (2026-09-04) — the design, and the two decisions inside it

### The brief, corrected

The 09-04 handoff put decision 4 as *"four `site-adoption-agent` steps declare a budget in a place
nothing reads"*. **That was true and it was half the defect.** The recursive census found a second,
larger failure at the other end of the ladder — ten agents whose root-level bare key was read FIRST and
capped their own steps. Recorded as a correction rather than edited away: the brief was not wrong, it was
scoped by the same fixed-path assumption that had hidden the rest for three weeks.

### The design decision: read every place, in one order, most specific first

Three options were on the table.

| option | what it does | why not / why |
|---|---|---|
| (a) refuse the bare spelling | fail loud on a misplaced key | **Rejected.** It converts 17 misplaced declarations into 17 unconfigured steps at the 2048 floor — exactly `bugs_open/205`'s failure. A reader that goes quiet is worse than one that reads generously. |
| (b) read only the canonical place, migrate everything | one spelling, enforced | **Rejected as the whole answer.** It is the same as (a) for anything written after the migration, and the migration cannot be applied to `html-developer-chunked` before the code rolls. |
| (c) **read every place in a stated order, migrate what is safe, report the rest** | chosen | The reader can never be the reason a declaration is ignored; the migration removes today's misplacements; the report is what notices tomorrow's. |

**The ordering within a level was free, and that is why it is safe.** [MEASURED 2026-09-04] no live agent
declares both spellings at any one level — asserted as guard 3 of migration 769 — so choosing
canonical-before-bare re-decides nothing an operator has written. If that ever stops being true, the
detector's `ambiguous` kind fires and the run fails.

### Why the two halves are independent, deliberately

Code is inert until a chassis roll; configuration is live on apply. Making the migration depend on the
roll would have left both failures live for an unknown number of hours. Making the code depend on the
migration would have left the next misplacement silent. So each fixes both failures on its own, and after
both, every reader agrees. **The single exception is `html-developer-chunked`, and it is the only real
ordering constraint claimed** — `getMaxTokens` read the bare key and nothing else, so moving those three
before the code rolls silently substitutes a hardcoded 16000. That is migration 770, held.

### DECISION 2, ruled: option (c) of the three the handoff offered

The owner said "leave it" to merging the last two copies of the precedence rule. The handoff's three
options were (a) leave two copies, (b) extract into `datahelpers`, (c) decide the rules should differ
permanently and write that down as a contract.

**This is (c), and it is more than (a).** The two ladders now share the *walker* — how a number is read
out of a map, and which values count as configured (`int`/`int64`/`float64`/`json.Number`, positive only,
because `max_tokens: 0` is a hard 400 from Anthropic). They keep separate *level lists*, because a direct
caller genuinely has no agent-level config to consult and is handed an already-merged `ai_service` map it
cannot order against. `TestDirectCallerLadderStaysTwoLevels` pins the difference so a later reader cannot
merge them by accident, and `llm_options.go` states the contract in prose next to the code.

⚠ **Note the consequence for the induced check**: it works precisely because the direct-caller ladder
reads the step's bare key before the merged block. Unifying the ladders later would silently retire that
probe.

### What the detector is for, stated as a design constraint rather than a feature

`llm_budget_call_sites_test.go` binds "no hardcoded budget" at the package and says in its own header
that it cannot see config that is declared and read by nobody. That is not a gap in that test — it is
structural: a Go test is package-scoped and offline. The class needs a join between the compiled rule and
208 live definitions, which is what `cmd/config-key-audit` exists for.

Two design rules it follows, both learned the hard way in this round:
1. **It calls production's ladder** (`actions.ResolveStepBudget`), so there is no second copy of the rule
   to keep in parity. A detector that re-implements what it checks can only confirm itself.
2. **A finding must be a decision an operator can act on.** The first cut reported the documented
   overlay pattern as a fault and produced 18 findings, all healthy. A guard whose first run cries on
   the healthy majority is one nobody opens twice.
