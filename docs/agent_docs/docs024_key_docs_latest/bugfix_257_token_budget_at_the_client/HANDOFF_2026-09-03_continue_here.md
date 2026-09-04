# HANDOFF 2026-09-03 — continue here (bugs_open/257, round 2: the class came back in a new shape)

> ## ✅ DELIVERED 2026-09-03 (later the same day) — §5 steps 1 AND 2 are WRITTEN AND COMMITTED (`51357cf51`)
>
> **Do not start this work; it is done, and it is LIVE on `v1.0.1360`.** Read this file for the
> ANALYSIS only. **The current entry point is
> `docs/agent_docs/docs024_key_docs_latest/bugfix_257_token_budget_at_the_client/HANDOFF_2026-09-04_continue_here.md`**,
> which carries the live proof, the three owner decisions, and the traps. Then
> `bugs_open/257…md` §2026-09-03b and §ROUND 2 IS LIVE for the detail.
>
> **⚠ §4's census below is WRONG in three places, and §5 step 1 is therefore NARROWER than what
> shipped.** `companies_house_llm_review_action.go` and `execute_vision_prompt_action.go` are listed
> as *"reads config, no literal — acceptable"*: the first passed an EMPTY options map, the second read
> the key with no `> 0` guard. And `feed_actions.go`'s Perplexity path hardcoded `4096` in a raw HTTP
> body — invisible to this census and to the three before it, because they all grep `\.GenerateText(`.
> All five are fixed. Details and the transferable lesson: `WRONG_CALLS.md` and `LANDMINES.md`
> (*"A budget census keyed on the CLIENT INTERFACE is blind to a provider called over raw HTTP"*),
> both 2026-09-03.
>
> **Still open, and still what a next session should pick up:** §5 step 3 (`llm_call_log` blindness),
> candidate 2 (merging the last two copies — a human call), the four dead `site-adoption-agent`
> `config.max_tokens` declarations found while doing this, and the post-roll live verification.
> Council: `c8660cfb-690d-4dd2-8b1f-25828305133e`.

**Written by:** the session that picked 257 up on 2026-09-03 after 18 days quiet.
**Status of that session's work: RESEARCH AND DOCUMENTATION ONLY. NO CODE WAS WRITTEN OR COMMITTED.**
**The fleet build of 2026-09-03 (`v1.0.1359`) therefore carries NOTHING from this session** — it is
someone else's roll and is named here only so a later reader does not mistake it for a delivery.

---

## 1. What this bug is, in plain terms, before any jargon

When the platform asks a language model to write something, it must tell the model how long the answer
may be. That number is called the **output-token budget**. It lives in configuration, in a block named
`ai_service`, under the key `max_tokens`, so that an operator can raise it for a step that needs longer
output without touching any code.

**The original defect (filed 2026-08-12):** only ONE function in the whole estate actually read that
configuration key and passed it to the model provider. Any code that talked to a provider directly
never read it, and silently used a hardcoded floor instead. The configuration looked perfect and did
nothing.

**That half is fixed and live.** On 2026-08-16 the fix ("Path A") moved the decision into the provider
clients themselves, so a caller that asks for no budget now inherits the configured one. It was council
APPROVED and verified at the running binaries. Nothing below retracts that.

**What this handoff adds:** the defect *class* has reappeared in a new shape, in code written AFTER the
fix, and this time it is measurable and it defeats the fix.

---

## 2. THE HEADLINE FINDING — a hand-rolled fallback beats the fix, and hides itself

Two actions written after Path A landed do not use any of the shared machinery. Each one re-implements
the budget rule by hand, and each ends with a **hardcoded number**:

| file:line | code | added |
|---|---|---|
| `platform/orchestration/actions/rewrite_negations_action.go:537-543` | `if mt, ok := aiServiceConfig["max_tokens"].(float64); ok && mt > 0 { … } else { options["max_tokens"] = 2000 }` | 2026-08-20, `a5d5d0728` |
| `platform/orchestration/actions/repair_ordering_register_action.go:436-440` | the same block, same literal | 2026-08-31, `f7156fb54` |

Three things follow, and the third is the one that makes this worth a round of its own.

**(a) The literal defeats Path A completely.** Path A works by having the client fall back to its own
construction config when the caller supplies no budget. These callers ALWAYS supply one — the config
value if they can read it, otherwise `2000`. An explicitly supplied option wins at the wire
(`platform/aiservice/anthropic.go:307`), so the client's new resolution never runs. A caller that passes
a literal is worse off than one that passes nothing: `nil` now inherits the configuration, `2000` never
can.

**(b) They are the fourth and fifth copies of one precedence rule.** The bug file's still-open candidate
2 describes "two hand-maintained copies". As of today there are five readers of `max_tokens`:
`ai_actions.go:357-360`, `llmOptionsFromConfig` (`llm_options.go:94`), `aiservice/max_tokens.go`, and
these two. `llm_options.go`'s own SCOPE note predicted this in writing — *"a THIRD caller should be the
extraction, not another paste"* — and two more pastes arrived anyway. Neither new copy reads the
**step's** own config, and neither forwards `budget_tokens`, so both are strictly less capable than the
helper sitting in the same Go package.

**(c) ⚠ THE LITERAL IS NUMERICALLY EQUAL TO THE CONFIGURED VALUE, SO THE FLEET'S OWN INSTRUMENT CANNOT
TELL THE TWO APART.** This is the important part.

`offer-analyser` declares `ai_service.max_tokens = 2000` on both steps that run
`repair_ordering_register`. The Go fallback is also `2000`. The value written to `llm_call_log.max_tokens`
is therefore `2000` **whether the configuration was read correctly or dropped on the floor**.

[MEASURED 2026-09-03] 29 calls to `repair_ordering_register`, every one at `max_tokens=2000`, zero
truncations, maximum observed output 600 tokens. That reading is **consistent with the code working and
equally consistent with the 257 defect being live at that call site**, and no query over `llm_call_log`
can separate them. The step is not currently burning — its answers are far below the cap — but the
instrument that would tell us if it started is blind by construction.

This is a new sub-class, and it is nastier than the original: the original defect at least showed a
*wrong* number (2048 where the config said 8000). This one shows the *right* number for the wrong
reason.

---

## 3. Why `rewrite_negations` is the worked example, and the near-miss it caused

`rewrite_negations` shows the whole mechanism, because its history happens to separate the two
hypotheses that `repair_ordering_register` fuses.

[MEASURED 2026-09-03] Every `rewrite_negations` call ever logged:

| budget sent | calls | truncated |
|---|---|---|
| 2000 | 99 | 4 |
| 16000 | 3,245 | 0 |

The 2000-token calls run 2026-08-21 to 2026-08-23; the 16000 ones from 2026-08-23 to today.

**I nearly recorded those four truncations as the Go literal biting production. They are not, and the
reason matters.** Migration `517` (2026-08-21) gave the step an `ai_service` block with
`max_tokens: 2000` — deliberately small, with a stated rationale. Migration `569` (2026-08-23) raised it
to 16000 after the 305 lane measured that a dense page with ten targets repaired **zero** of them. So
the four truncations are attributable to the **configured** 2000, which the platform then fixed the
correct way, through configuration.

**But the configured value and the Go literal were the same number for those three days**, so the
evidence could not have distinguished "the config arrived and was too small" from "the config was
dropped and the literal applied". The right answer was found by reading two migrations, not by any
query. The near-miss is logged in `WRONG_CALLS.md` under today's date.

The residual risk is real and unfixed: `rewrite_negations` still carries `else { options["max_tokens"]
= 2000 }` at line 542. Migration 569's repair lives **only in configuration**. Any future step, agent or
loop body that runs this action without declaring the key inherits 2000 again, silently, and 569's
measured fix does not transfer to it. Migration 517 documents exactly how easy that is: a loop substep
is named `process_sections_loop_iter_N_rewrite_negations`, which is not a top-level workflow step, so
`resolveAIServiceConfig` cannot find its block at all.

---

## 4. The census, re-run today — do NOT quote the bug file's §4

The bug file's §4 census (2026-08-12, nine sites) has now been wrong three times: it missed the Gemini
client, it missed a tenth site that arrived within days, and it predates these two. **Re-run it; the
population grows by addition and reads as current for ever.**

```bash
grep -rn '\.GenerateText(\|\.GenerateWithImages(' --include=*.go . \
  | grep -v '_test.go' | grep -v '^./platform/aiservice/'
```

[MEASURED 2026-09-03] Twelve production call sites. Grouped by what each actually does:

| group | sites | budget behaviour |
|---|---|---|
| canonical (`ExecuteAIStepAction`) | `ai_actions.go` :413 :463 :675 :853 | correct — the one enforced path |
| uses the shared helper | `provocation_gate_action.go:847`, `provocation_generator_action.go:554` | correct; step config + `budget_tokens` |
| **hand-rolled, hardcoded literal** | **`rewrite_negations_action.go:549`, `repair_ordering_register_action.go:446`** | **§2 — the new sub-class** |
| reads config, no literal | `companies_house_llm_review_action.go:133`, `execute_vision_prompt_action.go:212` | acceptable |
| outside the orchestration layer | `contentcreator/agent.go:305`, `reasoning/agent.go:127`, `tools-api` defend/gripper/position | see §6 |

Supporting live census [MEASURED 2026-09-03], 208 active agents:

| where the budget is declared | count |
|---|---|
| step `config.ai_service.max_tokens` | 149 |
| agent-level top-level `max_tokens` | 10 |
| root `ai_service.max_tokens` | 3 |
| step `config.max_tokens`, outside `ai_service` | 7 |
| any declared value `<= 0` or non-numeric | 0 |
| steps declaring `ai_service.max_tokens_ceiling` (bug 337's seam) | 1 |

Two notes on that table. The **7** top-level `config.max_tokens` entries (`site-adoption-agent` x4,
`html-developer-chunked` x3) are read by neither `ai_actions.go` nor the helper at the `ai_service`
level, and want checking separately — they may be a sixth spelling. And the **10** agent-level entries
all belong to `content-creator-*` agents whose step blocks declare 8000 while the agent top declares
1500-2000; since `ai_actions.go:357` gives the agent-level key priority over `ai_service`, the smaller
number wins. **None of those ten agents has ever written an `llm_call_log` row**, so this is a live
config trap with no traffic, not a burn. Worth a look, out of scope here.

---

## 5. THE PLAN — what the next session should build

Ordered by what makes the bad state unrepresentable, per the estate's own rule. **None of this is
written yet.** Steps 1 and 2 belong in one council submission; step 3 is separable.

### Step 1 — delete the two hand-rolled copies (the actual fix)

Replace both blocks with the helper that already exists in the same package:

```go
opts := llmOptionsFromConfig(config, aiServiceConfig, params.Logger, "rewrite_negations")
```

This is not cosmetic. It changes behaviour in three ways that must be stated in the submission, not
discovered by a reviewer:

1. **The step's own config becomes readable**, which it is not today. That is the helper's whole point.
2. **`budget_tokens` starts being forwarded** where a step declares it. [MEASURED 2026-09-03] **zero**
   active agents declare `budget_tokens` at root or step level, so this is inert on today's fleet.
3. **The hardcoded 2000 disappears.** An unconfigured call would then reach
   `aiservice.DefaultMaxOutputTokens` (2048) via Path A, not 2000. Both live steps declare the key, so
   [MEASURED] this changes zero requests today — but say so with the query, do not assert it.

⚠ **Preserve the small budget as CONFIGURATION, not as code.** Migration 517's rationale for 2000 is
sound and must not be lost: the answer is a few replacement sentences and a large ceiling only buys room
for an essay the gate will reject. `repair_ordering_register` keeps its declared 2000; the point is that
an operator can now change it and the code no longer pretends to decide.

### Step 2 — make the sixth paste impossible to merge quietly

The existing guard, `TestNoProvocationActionCallsAModelWithAnEmptyOptionsMap`
(`provocation_generator_action_test.go:460`), is well built — it refuses to pass when it matches nothing,
which is the failure mode register **OPP-003** exists to warn about. **But it is scoped to two named
files**, so it could not see either new caller. Generalise it to the package: every
`GenerateText`/`GenerateWithImages` call in `platform/orchestration/actions` must pass either `options`
built by the canonical path or `llmOptionsFromConfig(...)`, and never a numeric literal.

⚠ **Do not land that test before step 1.** On this shared tree, a test that fails on two existing
violations breaks HEAD for every other session. One commit, or the test second. And do **not**
allow-list the two current violations to make it green — an allow-list that silences the detector on
its own motivating case is a documented failure shape in this estate.

### Step 3 — `llm_call_log` blindness (the bug file's second open item, unchanged)

Still completely unaddressed, and §2(c) above raises its price. Direct callers are invisible to every
truncation instrument the estate has, because `llm_call_log` is written by the function being bypassed.
Since Path A they are invisible *and* running at whatever their construction config says.

Partial good news, [MEASURED 2026-09-03]: four of the six direct callers inside `platform/orchestration`
already call `LogLLMCall` — including both new ones, which is why §2 could be measured at all. The
genuinely blind population is now the two provocation actions plus everything outside the orchestration
layer (§6). The clients hold no database handle, so the write must stay at the caller's layer; that is
why this is a separate change with a different blast radius.

### Explicitly NOT in this plan

**Candidate 2 as originally framed** — unifying `ai_actions.go:357-360` with `llmOptionsFromConfig` —
stays open and stays a human call. The constraint found on 2026-08-16 still holds: `actions` imports
`aiservice`, so pushing the shared resolver down is an import cycle, and the two copies are not
behaviourally equivalent. Step 1 above reduces the number of copies from five to three without touching
that question.

---

## 6. Out of scope here, but do not lose them

- **`internal/agents/reasoning/agent.go:127`** passes `nil`. Path A made its configured 2048 live. It is
  **not a live burn** — no producer publishes to its topic. Do not re-file it as an outage; that has
  already been measured twice.
- **`tools-api` defend/gripper/position** hardcode 8192 inline and are config-blind. Same class as §2,
  different service, no configuration exists to drop.
- **`internal/agents/contentcreator/agent.go:305`** always passes a budget (default 3000).
- **The 7 top-level `config.max_tokens` entries** from §4 — possibly a sixth spelling nobody reads.

---

## 7. State of the record

| item | state |
|---|---|
| Path A (budget resolves at the client) | LIVE since v1.0.1305, council APPROVED `366efae9`, register **MDL-041** |
| Bug file | `bugs_open/257_HANDOFF_2026-08-12_the_token_budget_contract_has_one_entry_point_and_using_it_is_optional.md`, updated today with a §2026-09-03 re-validation section |
| Code written this session | **none** |
| Council submission for the plan above | **not made** — the plan is in §5, unsubmitted |
| Owner decisions still outstanding | candidate 2 (merge the copies); whether `llm_call_log` observability is in scope for 257 or a new lane |

**Verification route when step 1 ships** — the bug file's §6 still governs. Prove it at the binary with
a negative control, never at the tag; and the behavioural proof is that a step whose declared budget
exceeds the old literal actually sends the larger number, read from `llm_call_log.max_tokens`. ⚠ Pick a
step whose configured value **differs** from 2000 for that check, or §2(c) applies to your own
verification and it cannot fail.

## 8. Files

- **Bug:** `bugs_open/257_HANDOFF_2026-08-12_the_token_budget_contract_has_one_entry_point_and_using_it_is_optional.md`
- **Lane:** `docs/agent_docs/docs024_key_docs_latest/bugfix_257_token_budget_at_the_client/`
  (`NOTES_257_token_budget_at_the_client.md`, `README_where_we_are.md`,
  `RUNBOOK_257_token_budget_at_the_client.md`, `PLAN_2026-08-16_…`, two SUMMARY files, this handoff)
- **Code:** `platform/aiservice/max_tokens.go` · `platform/orchestration/actions/llm_options.go` ·
  `rewrite_negations_action.go:537-543` · `repair_ordering_register_action.go:436-440` ·
  `ai_actions.go:357-360`
- **Related bugs:** `bugs_open/205` (the unconfigured instance) · `bugs_open/305` (whose migrations 517
  and 569 are §3's evidence) · `bugs_open/337` (the escalation seam) · `bugs_closed/076`, `012`, `046`
  (the silent-truncation lineage)
- **Traps filed today:** `LANDMINES.md` (the equal-literal blindness) · `WRONG_CALLS.md` (the near-miss)
