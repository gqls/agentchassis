# ✅ 319 (FIXED AND LIVE 2026-08-19) — the tool-suggester now FAILS on `max_tokens`: `bugs_open/275`'s fix widened its menu from 30 tools to 80 and nobody raised the answer budget

**Filed 2026-08-19** by the `bugfix_275_silent_row_caps` lane. **This is a regression caused by this
lane's own fix**, found by the very run dispatched to prove that fix worked. **Live now.**

## What happened

`bugs_open/275` removed `LIMIT 30` from `load_library_tools`, so `suggest_tools` is now handed the whole
library instead of the first thirty alphabetically. It works — see §The fix is proven. But the model's
*answer* scales with the menu it is given, and the step's own answer budget was left where it was.

**Measured 2026-08-19 09:44:26Z**, the first post-fix run:

```
step suggest_tools failed: failed to execute action execute_llm_prompt:
AI call failed with unhandled error: response truncated: stop_reason=max_tokens
(output_tokens=3000 reached the configured cap, 11090 chars recovered);
raise max_tokens or shorten the prompt
```

The step FAILED. The orchestration is `FAILED` at `suggest_tools`. Nothing downstream ran.

## The arithmetic, before and after

| | prompt chars | output tokens | cap | outcome |
|---|---|---|---|---|
| 2026-08-15 18:18 (pre-fix) | 25,207 | 2,679 | 3,000 | ok — **89% of cap** |
| 2026-08-15 18:20 (pre-fix) | 20,764 | 1,749 | 3,000 | ok |
| 2026-08-15 20:29 (pre-fix, last ever) | 24,327 | 2,438 | 3,000 | ok — 81% |
| **2026-08-19 09:44 (first post-fix)** | **33,818** | **3,000 (cut)** | 3,000 | **FAILED** |

**The headroom was already thin and nobody had looked.** Across all 59 successful calls in history:
output tokens ranged **1,178–2,921**, mean 1,932, and **5 of 59 came within 15% of the cap**. The
all-time maximum was **2,921 — 97.4% of the 3,000 budget** — *before* the menu tripled. The fix did not
create a fragile margin; it consumed one that was already nearly gone.

The budget is the step's own: `agent_definitions` → `suggest_tools.config.ai_service.max_tokens = 3000`
(`claude-sonnet-4-6`, anthropic). Since **MDL-041** (`bugs_open/257`) that value is resolved at the
provider client, so it is the live lever and it will be honoured.

## ✅ Two things this does NOT mean

1. **It is not a reason to re-cap the library.** The cap was the defect; the answer budget is the thing
   that is now mis-sized. Restoring `LIMIT 30` would trade a visible failure for the silent one that
   `bugs_open/275` exists to end.
2. **Nothing was corrupted.** The platform **failed closed**: the truncated 11,090 characters were
   discarded rather than persisted, the step errored, and `create_items_loop` never ran — **zero
   `add_tool` work items were created** on the target site. That is CLAUDE.md's
   `output_tokens == max_tokens means the completion was CUT` rule working as designed, and it is the
   only reason this was caught in one run rather than by someone noticing odd tool suggestions weeks later.

## ⚠ Scope: how many sites does this break?

**Unmeasured, and stated as such.** n=1 post-fix run (gamesdesign.co.uk: 7 tools deployed, 81 in the
library, so a large candidate set). The answer length depends on how many gaps the model finds, so a
site with fewer gaps may still fit. What IS established is that the margin is gone: the pre-fix
distribution already touched 97% of the cap with a menu a third the size. **Treat the agent as broken
until a run succeeds**, and note it has no automatic retry that would fix itself — the failure is
deterministic for a given prompt.

## Fix candidates, ordered by what closes the door

1. **Raise `suggest_tools.config.ai_service.max_tokens`.** The minimal restoration of function. Evidence
   for a value rather than a guess: the cut answer produced 11,090 chars in 3,000 tokens (~3.7 chars/token)
   and was *still incomplete*, so 3,000 is not marginally short. **6,000** gives roughly 2× the all-time
   maximum against a menu that grew ~2.7×. ⚠ It is an `agent_definitions` migration: snapshot first,
   pre-state gate on the current value, `DO`/`RAISE` verify — a bare `SELECT` verify block cannot stop the
   `COMMIT`. It is live on apply; no roll.
2. **Bound the ANSWER as well, in the prompt** — "return at most N suggestions". This is the output-side
   analogue of what 275 did on the input side (bound the dominant field rather than the row count), and it
   caps cost per call permanently instead of raising a ceiling that the next library growth will reach
   again. Best done *with* (1), not instead of it.
3. **Then add the check that would have caught it before the fix shipped:** for any step whose prompt
   payload is being widened, read `llm_call_log.output_tokens` against that step's `max_tokens` FIRST.
   One query, and it would have shown a 97.4% high-water mark before migration 445 was applied.

## How to verify a fix

- Dispatch one suggester run (recipe in the lane RUNBOOK) and require `llm_call_log.success = true` with
  `output_tokens < max_tokens` — **assert the strict inequality, not just success**, or you re-create the
  same blind spot one cap higher.
- The disconfirming arm already exists and is free: today's row is `success=false` with
  `output_tokens = max_tokens = 3000`.

## Filing basis (owner ruling 2026-07-31)

**No `090` run; substitution stated plainly.** No mechanism is asserted that is not already in the error
string the platform produced: the provider client reported `stop_reason=max_tokens` against a named
configured cap, and the before/after token counts are rows in `llm_call_log`. The causal claim (275
widened the prompt) is a one-line diff plus a dated migration, and the run that failed is the run
dispatched to test exactly that change. Grepped both bug directories: nothing covers this step or this cap.

## The transferable lesson

**Bounding an input payload can move the failure to the output budget.** `bugs_open/275` measured the
prompt carefully — per column, to find that `description` was 80% of it — and never once looked at what
the *answer* costs. The register entry for the fix (**LCO-009**) discusses payload size at length and does
not mention `max_tokens`. **When you widen what a model is shown, the next thing to measure is what it
says back.**

## Related

`bugs_open/275` (the fix that caused this; its §2026-08-19 entry records the same run) · **MDL-041** /
`bugs_open/257` (why the step's `ai_service.max_tokens` is now the live lever) · register **LCO-009** ·
CLAUDE.md, *"`output_tokens == max_tokens` means the completion was CUT"* · `bugs_open/012` (the class:
a truncated completion persisted as success — which is exactly what did NOT happen here).

## ✅ FIXED AND LIVE — 2026-08-19 10:25Z, migration 484

**The cause was one instruction, not "answers got longer".** The prompt asked the model to list every
library tool it rejected, with a reason, so `rejected_tools` carried roughly one entry per tool SHOWN.
Measured across the five most recent successful answers it was **37–66% of the response**, with counts
of 28, 30, 26, 30, 19 — i.e. the menu size. Tripling the menu tripled that section: ~4,400 chars
becomes ~11,500, which is exactly the 11,090 recovered before truncation.

**Migration `484_tool_suggester_answer_budget_and_bounded_lists.sql`** (applied by hand, single-file
psql, rehearsed `COMMIT`→`ROLLBACK` first with the live row verified unchanged; recorded in
`schema_migrations`; `_ROLLBACK` sidecar committed before the apply):

1. `suggest_tools.config.ai_service.max_tokens` **3000 → 6000** (owner-authorised).
2. **AT MOST 8** entries in `suggestions` — the highest any run has ever produced, so no observed
   behaviour is cut.
3. **AT MOST 5** entries in `rejected_tools` — the term that scales with the library, and the one
   nothing downstream reads (`save_tool_spec` does not persist it).

### Verified at the artefact, with the strict inequality the file asked for

| | before (2026-08-19 09:44) | after (2026-08-19 10:25) |
|---|---|---|
| `success` | **false** | **true** |
| `output_tokens` | **3000** (= cap, truncated) | **1761** |
| `max_tokens` | 3000 | 6000 |
| **strictly under cap** | **no** | **yes — 29% of budget** |
| `rejected_tools` listed | ~78 (cut mid-list) | **5** (bound held exactly) |
| `suggestions` | — (never returned) | 7 (under the 8 ceiling) |
| orchestration | FAILED at `suggest_tools` | **COMPLETED** |

**The answer is now smaller than the pre-fix average (1,761 vs 1,932) despite a 2.7× bigger menu** —
which is the evidence that the bound, not the raised ceiling, did the work. The extra 3,000 tokens of
headroom were not consumed; they are there so the next library growth does not re-open this bug.

⚠ **The run also exposed `bugs_open/321`:** of those 7 suggestions only **1** became a work item,
because every novel suggestion for a site collides on the same `item_key`. Fixing this bug made the
model's reply reachable; 321 is why most of it still goes nowhere.
