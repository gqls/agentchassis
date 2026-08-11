# 244 — The council gate sends 15 near-identical 106k-token prompts per round, uses no prompt caching, and orders the prompt so caching could not work if it were switched on

> # ✅ FIXED AND LIVE — 2026-08-10 evening, by another session, ~2 hours after this was filed
>
> **Do not build this fix. It exists.** Kept in `bugs_open/` per the owner ruling of 2026-08-06
> (a finished bug stays here), and left open only for the adoption gap in §8.
>
> - `3d6851d9b` **perf(aiservice)** — opt-in `cache_control` breakpoint on the shared Anthropic
>   client + the counters that make it falsifiable (migration 376 added
>   `cache_creation_input_tokens` / `cache_read_input_tokens` to `llm_call_log`).
> - `071adc44c` **perf(council-gate)** — hoisted the shared prefix in all 17 seats. Went through
>   the council gate itself; the edit-quality seat caught a real defect on the way (the marker
>   leaking into the prompt as content, correlation `b54f173e`).
>
> **Design is better than what §4 proposed.** Rather than a parameter, the split is a marker
> (`<!--CACHE_BREAKPOINT-->`, `platform/aiservice/anthropic.go:125`) placed **in the DB-held
> template**, so the shared/varying boundary lives where it is actually authored and visible to a
> reviewer. **Opt-in by construction** per the owner ruling of 2026-08-02 §2 (RFC_010): this client
> is the seam every agent calls, so the unsafe default is OFF — a caller that has never heard of
> caching runs the identical code path as before, byte for byte.
>
> ## Measured on live traffic, 2026-08-11 [MEASURED]
>
> | per council round | before | after |
> |---|---|---|
> | full-price input tokens | **806,024** | **127,783** |
> | cache reads | 0 | **973,554** |
> | cache writes | 0 | **93,333** |
> | rounds in sample | 221 | 17 |
>
> Effective cost (reads 0.1×, writes 1.25×): **806,024 → ~341,800 full-price-equivalent, a ~58%
> reduction per round.** Total prompt volume per round actually *rose* ~37% over the sample, so
> **per token of prompt the cost fell ~69%.** Cache hit rate on read-eligible seats: **157 of 170
> = 92.4%**; seat 1 writes (17 calls, 15 writes, 0 reads) exactly as designed.
>
> ## Two things this file got WRONG, both corrected by the measurement
>
> 1. **My "≈76% reduction" was optimistic.** Real figure is ~58% per round / ~69% per token. I
>    assumed only ~22,500 tokens per round would remain unshared; the true residue is ~127,783.
>    The estimate was directionally right and quantitatively loose — **it was arithmetic on a
>    measured input, never an observed bill, and it read as more precise than it was.**
> 2. **My `ttl: "1h"` recommendation was unnecessary, and the implementer was right to omit it.**
>    I reasoned that rounds run 459s mean / 1022s max, so a 5-minute TTL would expire mid-round.
>    `cacheTTL = ""` (no field ⇒ 5-minute default) and the data **refutes my concern**: seats
>    landing *past* 5 minutes have a **higher** hit rate (75/82 = 91%) than seats within it
>    (82/105 = 78%, and the misses there are the writing seat). Reads keep the entry alive, so
>    elapsed round time is not the variable I thought it was. The code comment invites exactly this
>    evidence-led check before adding a TTL — **that check has now been run, and the answer is
>    "leave it".**
>
> ## §8 — what is genuinely still open: ADOPTION
>
> The marker is opt-in, and **only `council-gate` has adopted it — 17 steps, zero other agent
> types** [MEASURED 2026-08-11]:
> ```sql
> SELECT type, count(*) FROM agent_definitions, jsonb_each(default_config->'workflow'->'steps') AS s(n, step)
> WHERE is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL
>   AND step::text LIKE '%CACHE_BREAKPOINT%' GROUP BY 1;
> ```
> That is the right first target (council was 87.8% of spend), but `page-content-writer` was 7.1%
> and others follow. **Before adopting a template, check it actually has a stable prefix** — the
> whole mechanism is a prefix match, so a per-call name, timestamp or id above the marker silently
> costs the write and never reads.

**Filed 2026-08-10 by the `bugfix_168_deployed_asset_path` lane**, after the fleet's
Anthropic API usage limit was exhausted at **14:51:47Z on the 10th of the month**, with the
stated reset **2026-09-01 00:00 UTC**.

> **CORRECTED 2026-08-11 — the outage was ~3h20m, not the 21 days this file originally claimed.**
> [MEASURED] last usage-limit failure **2026-08-10 17:02:12Z**, first success after it
> **18:12:11Z**: the owner raised the cap the same afternoon. I wrote "a 21-day fleet-wide LLM
> outage" by treating the vendor's *stated reset* as a forecast, and escalated on that basis.
> **The stated reset is a worst case; the binding variable is whether a human raises the limit.**
> Check the **success** side of `llm_call_log` before acting on the failure side.
>
> **None of the measurements below change, and the defect is not softened by the quick fix** —
> if anything it is sharpened: the budget was exhausted on the **10th of the month**, so absent
> the caching fix this recurs roughly monthly. Filing stays OPEN.

> **On the diagnosis-loop norm (CLAUDE.md, owner ruling 2026-07-31).** This file asserts a
> cross-cutting root cause and would normally have to go through `090` first. **The loop is
> LLM-backed and is itself down** — it is one of the things the cap has killed, so running it
> is not available. Substituted first-hand verification, stated per the escape hatch: I read
> the live request builder, extracted three real seat prompts from one production round and
> compared them byte-wise, and measured the token totals against `llm_call_log`. Every figure
> below carries its query or its `file:line`. **Route this through `090` once the cap lifts** —
> the mechanism is measured, but a second pair of eyes on the *fix* is still owed.

---

## 1. The symptom that started it

Council round 4 of this lane (`b67eb26a-14ef-45d7-b755-3e489fd57ef0`) reached
`COMPLETED / complete_invalid` with **no verdict**. `plan_valid` was `true` and the plan had
persisted — the submission was fine. The cause is only in `__step_error`:

```
step review_architecture failed: failed to execute action execute_llm_prompt:
AI endpoint unavailable: provider=anthropic model=claude-sonnet-5 ...
status 400: "You have reached your specified API usage limits.
You will regain access on 2026-09-01 at 00:00 UTC."
```

This exact shape is **already a documented landmine** (`LANDMINES.md`, added 2026-07-31 by the
`bugfix_149` lane) and its remedy — *"STOP and wait for the stated time"* — was correct then,
when the reset was 5 hours away. **This time the wait is 21 days**, which makes it an owner
decision rather than a remedy. That entry has been updated with this recurrence.

## 2. What is measured

**The outage.** [MEASURED 2026-08-10 16:54Z]

```sql
-- last success, and everything after it
SELECT success, model, count(*), min(created_at), max(created_at)
FROM llm_call_log WHERE created_at > '2026-08-10 14:00:00+00' GROUP BY 1,2 ORDER BY 4;
```
Last successful call fleet-wide: **14:51:45Z**. Every call since (n=5 at time of filing) fails
with the account-level cap. Blast radius over `orchestration_states`, using the needle the
landmine specifies (`ILIKE '%usage limit%'`, **not** `spending limit` — `WRONG_CALLS.md:14507`):
**5 runs killed**, first 14:51:49Z.

**Is the cap account-wide, or just `claude-sonnet-5`?** **Account-level — [MEASURED, by another
lane, not by me].** From my own evidence I could only mark this `[UNMEASURED]`: all 5 failures
are sonnet-5 because nothing had *attempted* another model since 14:43Z, and the vendor's
"your specified API usage limits" wording is not a measurement. The `webdesign.uk` chat lane
settled it the same afternoon by reproducing the identical refusal **from a standalone service
outside this cluster** (`6a4fbab21`, recorded in `LANDMINES.md`), which rules out a chassis
credential fault. **Read their entry before re-deriving this.**

⚠ **A pod-log grep cannot settle questions like that one** — I ran one across four services,
got 0 hits, and then found the chassis log window was ~2 minutes wide for a `--since=3h`
request. The silence was worthless. Measure at `llm_call_log` / `orchestration_states`.

**Where the month went.** [MEASURED, August 1–10]

```sql
SELECT agent_type, count(*) AS calls, sum(input_tokens) AS in_tok,
       round(100.0*sum(input_tokens)/sum(sum(input_tokens)) OVER (),1) AS pct
FROM llm_call_log WHERE created_at >= '2026-08-01' AND input_tokens IS NOT NULL
GROUP BY 1 ORDER BY 3 DESC;
```

| | |
|---|---|
| Fleet input tokens, Aug 1–10 | **188.1M** (output 11.6M) |
| `council-gate` share | **165.2M = 87.8%** |
| Council rounds | **209** in 10 days (~21/day) |
| Input tokens **per round** | **790,551** average |
| Mean input tokens **per call** | 2,600 (mid-July) → **62,000** (08-10), ~24× |

The `agent_type` relabelling trap (memory: `llm-call-log-agent-type-relabelled`) does not bite
here — the relabel was 2026-07-26 and this window is August-only, so attribution is clean.

**Why each call is so large — and why caching cannot currently help.** [MEASURED]

One production round, 15 seats, each prompt **272,000–300,000 chars / 106,000–118,000 input
tokens**. Three seat prompts pulled from `llm_call_log.prompt_rendered` and compared byte-wise:

| pair | common prefix | longest shared block |
|---|---|---|
| editquality vs compliance | **20 chars** | **268,980 chars = 98.6%** |
| editquality vs mission | 20 chars | — |

**98.6% of every seat's prompt is byte-identical.** The seat-specific head is only
**1,387–5,159 chars**. The shared block opens with a full DB schema dump
(`## Schema (the ONLY tables available to checks)`) followed by the submission.

**But the shared block sits AFTER the seat-specific header.** Prompt caching is a *prefix*
match, so with the volatile content first, a cache breakpoint could never hit — the prompts
diverge at character 21. This is the textbook silent-invalidator ordering.

**And caching is not enabled anyway.** `grep -rn "cache_control" --include=*.go platform/ pkg/
internal/` returns **nothing**; the only repo hits are a docs-only sample engine under
`docs/.../idea.uk/`. The live builder (`platform/aiservice/anthropic.go:103-116`) sends
**one `user` message with the whole prompt inline and no `system` field at all**:

```go
requestBody := map[string]interface{}{
    "model":      c.model,
    "max_tokens": 2048,
    "messages": []map[string]interface{}{{"role": "user", "content": content}},
}
```

## 3. Root cause

Three independent things compound, and **all three must be fixed for any of them to pay**:

1. **No prompt caching anywhere in the platform.** Every token is billed at full input rate.
2. **Prompt ordering makes caching unreachable** — volatile seat header first, 268k chars of
   stable shared content last. Fixing (1) without (2) buys nothing.
3. **The council fans out 11–15 seats per round**, each re-sending the same 268k chars. The
   relevance gate that CLAUDE.md describes as cost control gates *which seats fire*, not *how
   much context each one carries*.

## 4. The fix, and what it is worth

Restructure the council prompt so the shared block leads, then cache it:

- Move the shared block (schema + submission) into `system`, seat instruction into the `user`
  turn — `system` renders **before** `messages`, so this ordering is what the API wants.
- Add one `cache_control` breakpoint on the last shared block. Only 1 of the 4 allowed is needed.
- **Use `ttl: "1h"`, not the 5-minute default.** [MEASURED] a council round takes **459s mean,
  1022s max** over 193 August rounds — both exceed a 5-minute TTL, so the default would expire
  mid-round and every seat would pay a fresh write.

Arithmetic on the measured round (15 seats × ~106k tokens), 1h TTL at 2× write / 0.1× read:

| | input-token equivalents per round |
|---|---|
| today | 15 × 106,000 = **1,590,000** |
| cached | 212,000 write + 148,400 reads + ~22,500 unique heads ≈ **383,000** |

**≈76% reduction on the council**, which is 87.8% of fleet spend. Applied to August's 165.2M
council tokens that is ~40M instead — fleet input would have been ~63M rather than 188M.

⚠ **[INFERRED, not measured] the saving itself.** The percentages are arithmetic on measured
inputs, not an observed bill. The disconfirming result to watch for after shipping:
`cache_read_input_tokens` staying at 0, which is what a surviving silent invalidator looks like.
**`llm_call_log` has no cache columns**, so the fix must add them or the win is unverifiable —
that is part of the work, not a follow-up.

**Secondary, cheaper, and independent:** the schema dump at the head of the shared block is sent
to all 15 seats whether or not a seat writes SQL. Trimming it is worth doing but is not the
structural fix and should not be confused for one.

## 5. What this does NOT explain

- The cap's *level* is the owner's setting; I cannot see it. "Council is 88% of spend" says where
  it went, not whether the limit was set too low.
- **The lane's own revalidator sweep is unaffected and still running** — agent
  `diagnosis-review-queue-revalidator`, steps `sweep` + `complete`, no LLM step, and no LLM
  reference in `revalidate_review_queue_action.go`. It closed items on the 08-10 08:44Z run and
  will keep doing so through the outage. Pure-Go/DB work is not blocked; only LLM work is.

## 6. How to verify a fix

1. Pod-grep the running binary for a string the change adds **and a negative control** — a
   string it removes, expecting 0 (`bugs_open/153`).
2. After one council round: `cache_read_input_tokens > 0` on seats 2..N of the same round. A
   zero means the prefix is still diverging — diff two seat prompts byte-wise, as in §2.
3. Re-measure `sum(input_tokens)` per round and compare against the 790,551 baseline above.

## 7. Correlations

Round 4 orchestration `2f1b43f6-d92b-49eb-843b-204d0da235fa` (corr
`a7b35edc-4094-42b4-a2e7-a863de831e6b`), the run whose death exposed this.
Lane: `docs/agent_docs/docs024_key_docs_latest/bugfix_168_deployed_asset_path/`.
