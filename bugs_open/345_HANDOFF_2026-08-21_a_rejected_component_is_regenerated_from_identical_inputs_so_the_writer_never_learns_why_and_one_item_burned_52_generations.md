# 345 — a component rejected by pre-store validation is regenerated from **identical inputs**, so the writer never learns why it was rejected: every repeat produces the *same* rejection, and one item burned **52 generations** under a 3-attempt budget

**Filed:** 2026-08-21 by the `bugfix_311_component_keys` lane, after it blocked the repair of
`bugs_open/311`'s **originating page**. **Status: OPEN — FIX BUILT AND COMMITTED 2026-08-21, INERT until the next chassis roll.**
`Council-Submitted: 67b07528-b40b-4eef-9abc-35ad70efae04`.

> **⚠ WHERE THE CODE ACTUALLY IS, because `git log` will not tell you.** The **Go half is at HEAD
> inside commit `0f80f5ea1`, whose message is about `bugs_open/344`** — another session edited
> `load_work_item_actions.go` for the failure-ladder work while my edit was in the shared working
> tree, and a pathspec commit takes the file from the tree. The documented same-file passenger
> trap, firing exactly as written. **Nothing is lost and forward-only holds**: verified by
> exporting `git archive HEAD` to a clean directory, where the package builds and all four 345
> tests pass. So `git log --oneline -S 'item["last_error"]'` finds it and `git log` on this bug's
> number does not. My own commit (tests + migration `533`) therefore describes a Go change it does
> not contain — stated here rather than left to mislead.
>
> **What shipped:**
> - **Go** (`load_work_item_actions.go`): `wi.error` joins the loader's SELECT, scanned as
>   `sql.NullString`, and surfaces as `current_item.last_error` **only when `attempt_count > 0`**
>   and non-blank, **capped at 2,000 chars** with an explicit `…[truncated]` marker. A first
>   attempt is byte-identical to before; NULL/empty/whitespace leave the key **absent** rather than
>   present-and-empty.
> - **Tests** (`load_work_items_last_error_test.go`): four properties, each **mutation-proven** —
>   removing the attempt gate, the blank check, or the cap fails its own test and only that one,
>   and the restored control is green. Two existing sqlmock tests were updated because a new SELECT
>   column breaks a positionally-declared column list.
> - **Config** (`docs/agent_docs/sql_for_agents/533_component_creator_prompt_reads_last_error.sql`
>   + `_ROLLBACK`): the prompt renders **named** placeholders, so the new key is invisible until it
>   is referenced. `{{if}}`-guarded, `DO`/`RAISE` guards on both sides (a verify block of `SELECT`s
>   cannot stop a `COMMIT`), anchored on a verbatim line. **NOT YET APPLIED.**
>
> **No ordering constraint, and this file does not claim one:** with the Go half absent the block
> renders nothing; with the config half absent the key is simply unread. Either may land first.
> **Candidates 2–4 of the list below are NOT done** — the retry budget is still spent on a
> deterministic refusal, the ~17-generations-per-attempt inner loop is untouched, and no source
> enumeration was added.
>
> **How to tell whether it worked, once rolled:** a second attempt's rejection reason should
> **differ** from the first. Under this bug it never has, in 99 rejections.
**Severity: live, wasteful, and it makes a whole class of page unrepairable** — 99 rejections
across 3 sites since 2026-08-15, and not one retry has ever produced a different outcome.

## The mechanism, and it is three lines of config

1. `store_generated_component` runs a **pre-store validation** that refuses a template whose
   declared field sources cannot resolve. Its message is **fully actionable**: it names the field,
   quotes the bad source, and enumerates every `site_specs` aspect that *does* exist (~60 of them).
2. The item goes back to `triaged` and `component-creator` regenerates.
3. `generate_template`'s step config carries **`input_fields: ["input_data", "site_record",
   "site_specs", "existing_component"]`** — read from the live `agent_definitions` row, 2026-08-21.
   **There is no field for the previous failure.** The writer is handed the same inputs it had the
   first time and asked the same question, so it produces the same answer.

**The rejection message and the regeneration are not connected.** The information needed to fix the
template exists, in full, at the moment of failure — and is discarded.

## Evidence — the retry has NEVER varied the outcome, and it could have

`agent_error_log`, `error_code='component_validation_rejected'` [MEASURED 2026-08-21 12:05Z]:

- **99 rejection rows**, 3 sites (`loanzy.uk` 78, `remortgagecalculator.uk` 11,
  `loancalculator.co.uk` 10), 2026-08-15 → 2026-08-21.
- Grouped per work item, **every single item with more than one rejection has exactly ONE distinct
  rejection reason** (`count(DISTINCT md5(error_message)) = 1`) — twelve such items, from 2 repeats
  up to 52. Across all 99 rows there are only **10 distinct reasons**.
- **This is the disconfirming test, and it failed to disconfirm.** A stochastic writer regenerating
  from scratch would be expected to invent a *different* bad field sometimes, or occasionally get it
  right. It never does. The mistake is a stable property of the prompt-plus-inputs, so the retry
  cannot succeed — a point the sibling entry in `016b` §9 makes for cap failures.

**And the spend is far worse than the item's budget suggests.** Item
`8c8f5de5-8078-440f-974d-70d0ba68346b` (`needs_new_component`, `loanzy.uk`) shows
`attempt_count = 3 / max_attempts = 3` and produced **52 rejection rows in 3 h 34 min**
(15:23:14 → 18:57:20 on 2026-08-18). So **something retries ~17× inside a single attempt**: the
work-item budget bounds the dispatches, not the generations. Two other items burned 10 and 8.

## The live instance that prompted this file

`remortgagecalculator.uk` / `index`, section **`mortgages-repayment`** — the calculator whose
absence is the owner's original complaint in `bugs_open/311` ("remortgagecalculator.uk left out the
actual tools"). Item `95fe67da`, two attempts (10:24:55Z, 11:42:34Z), **both rejected identically**:

> field `"currency_symbol"` declares source `"site_specs.locale.currency_symbol"` but no site
> carries a `site_specs` aspect named `"locale"` — the value would resolve nowhere on every site
> and the field would be silently omitted (`bugs_open/309`)

**The guard is right and must not be worked around.** `grep -rn "locale"` / `currency_symbol` over
`platform/`, `internal/` and `docs/agent_docs/sql_for_agents/` returns **nothing**: no such aspect
exists anywhere in the framework, so the writer invented the source. Seeding a `locale` aspect to
make one hallucinated field resolve would be fixing the checker to agree with the broken output.

**Worth recording for `311`'s sake:** the rejection names
`function="mortgages-repayment-remortgagecalculator-uk"` — **the site-scoped name**. So `311`'s
collision diversion FIRED correctly and derived the right identity; this defect sits *downstream*
of it. Identity resolution precedes pre-store validation.

## Fix candidates, ordered by what makes the bad state unrepresentable

1. **Feed the rejection back into the regeneration.** Add the previous validation failure to
   `generate_template`'s inputs (a `previous_rejection` / `last_error` field) and reference it in the
   prompt. This closes the door for **every** validation class the guard covers, not just invented
   sources, and it is the only candidate that makes a retry *capable* of succeeding. It is also
   the one that needs care: the failure text is model-authored in part, so it is untrusted input
   heading back into a prompt.
2. **Classify a validation rejection as non-retryable.** It cannot succeed unchanged, so stop
   paying for it: park on the first rejection with the actionable message attached. Saves the spend
   and repairs nothing — but it converts a silent 52-generation burn into one visible refusal.
3. **Bound the inner loop.** Whatever retries ~17× inside one attempt is the actual cost defect;
   `attempt_count` is not bounding generations. Find it and cap it regardless of candidates 1-2.
4. **Enumerate valid sources in the generate prompt.** Weakest, and the measurement argues against
   it: `site_specs` is **already** an input, so the writer could see which aspects exist and
   invented one anyway. Telling it again is not obviously different from telling it once.

## How to verify a fix

Re-drive `mortgages-repayment` on `remortgagecalculator.uk` (recipe:
`docs/agent_docs/docs024_key_docs_latest/bugfix_311_component_keys/RUNBOOK_311_fix.md`) and assert
**at the artefact, not the item status**: `https://remortgagecalculator.uk/index.html` goes above
**0 `<input>`** from a pinned before of **200 / 40,726 bytes / 0 `<input>` / md5
`89910f6e7875f1d310d962f83e443989`**. For the mechanism itself, the sharper assertion is that a
second attempt's rejection reason **differs from the first** — under this bug it never has, so
"two attempts, two different reasons" is the signal the feedback path is live.

## Declared substitute for the `090` loop (owner ruling 2026-07-31)

No `090` run, and the reason is that nothing here is inferred: the missing input field is **read
from the live `agent_definitions` row**; the identical-rejection finding is a `count(DISTINCT
md5(error_message))` over 99 rows that could have returned any number; the 52-generation burn is a
row count with timestamps; and the invented source is a `grep` returning zero over three trees.
What this file does **not** claim: which component retries inside an attempt (candidate 3 is a
lead, not a diagnosis), nor that candidate 1 would fix the hallucination rate — only that it makes
a retry able to differ.

## Related

- `bugs_open/311` — the originating page this blocks. Its own fix works here (the scoped name was
  derived); this defect is downstream.
- `bugs_open/337` — sibling, filed by the same lane the day before: a **cap** failure retried
  identically. Same shape (deterministic generation failure consuming a retry budget), different
  trigger. Both would be caught by candidate 2.
- `bugs_open/309` — the guard's own citation, and the reason the refusal is correct.
- `016b` §9, "N identical failures with IDENTICAL numbers is a deterministic refusal" — the
  diagnostic that turned this from "flaky generation" into a measurement.
