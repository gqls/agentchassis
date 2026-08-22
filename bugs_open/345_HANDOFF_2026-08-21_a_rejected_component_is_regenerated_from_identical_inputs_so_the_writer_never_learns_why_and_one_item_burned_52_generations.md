# 345 — a component rejected by pre-store validation is regenerated from **identical inputs**, so the writer never learns why it was rejected: every repeat produces the *same* rejection, and one item burned **52 generations** under a 3-attempt budget

**Filed:** 2026-08-21 by the `bugfix_311_component_keys` lane, after it blocked the repair of
`bugs_open/311`'s **originating page**. **Status: OPEN — FIX LIVE ON BOTH HALVES since 2026-08-21 ~17:05Z** (v1.0.1322 carries the Go
half — provenance stamp `bac189921`, `0f80f5ea1` ancestor — and migration `533` is applied and
ledger-recorded; the CSS lane's contrib below independently probed both replicas with a control and
confirmed the live prompt carries the `{{if .input_data.last_error}}` block). **⚠ CORRECTED 2026-08-22 (council round 4, guardian HIGH): THE FIX IS INERT — NOT MERELY
UNEXERCISED.** `build-dispatch-loop`'s `call_handler` is a `call_agent` step with an explicit
**`input_mapping` allow-list of 14 keys, and `last_error` is not one of them** — so the loader's new
key never becomes part of the handler's `input_data` and `533`'s `{{if .input_data.last_error}}`
cannot fire. Fleet census of all 73 live `call_agent` mappings: **zero pass it.** A THIRD half is
required: add `last_error?` to that mapping (and check `site-work-orchestrator`'s 8-key
`sub:call_handler`) — a dispatcher migration on a shared seam, to be named as such.
**I had measured the zero and explained it away as "no retry yet"** — the "a post-fix ZERO needs a
DEMAND control" lesson, paid again. ~~DEMAND BAR OPEN: the
path has never fired~~ — `collected_data->'input_data' ? 'last_error'` is 0 across all history
(their query), and the first post-fix generation on the originating page succeeded on attempt 0, so
no retry existed to feed. The signal that closes this: two `component_validation_rejected` rows on
ONE item with **different** `md5(error_message)`. **⚠ Do NOT read `bugs_closed/311`'s close as this
bug's proof:** the originating page healed on an attempt-0 success — the feedback path never fired,
because there was no retry to feed. "The page has its calculator" and "345 worked" are different
claims, and only the first is established. Note the ROUND-2 REVISION (gate = non-blank
recorded failure, not `attempt_count>0`) is committed but rides the NEXT roll — the live binary
still carries the attempt-gated form.
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
(15:23:14 → 18:57:20 on 2026-08-18). ~~So **something retries ~17× inside a single attempt**: the
work-item budget bounds the dispatches, not the generations.~~
> **CORRECTED 2026-08-21 (council round 1, corr `67b07528` — the reviewers' own verification query
> refuted this).** The 52 rejections are **52 DISTINCT `orchestration_id`s, ONE rejection each**,
> ~4 minutes apart — not an inner loop. The burn was **dispatch-without-counting**: the old
> `isAIUnavailable` arm released the item to `triaged` forever without consuming an attempt. That
> arm was replaced by `bugs_open/307`/`344`'s failure ladder (live since v1.0.1322), which counts
> every real failure — but the ladder **still has a designed transient release**
> (`work_item_failure_ladder.go:279`, "attempt NOT consumed", writing `error` at `:570`).
> **Which is why the fix's gate changed in round 2:** `attempt_count > 0` would have hidden the
> failure text from exactly the uncounted re-dispatches; the shipped gate is **a NON-BLANK recorded
> failure** (a fresh item has `error` NULL, so first generations stay byte-identical). Candidate 3
> below ("bound the inner loop") is therefore RETARGETED: the loop was the dispatcher's, and
> 307/344 already bounded it.
Two other items burned 10 and 8.

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

---

## CONTRIB 2026-08-21 (evening, `remortgagecalculator.uk` CSS lane) — the fix is LIVE, it has NEVER FIRED, and the hallucination was a NEAR-MISS on a real namespace

The owner asked me why remortgagecalculator.uk still has no calculator. I traced it here rather
than filing anything: this file already owns the mechanism, `bugs_open/309` owns the guard that
emits the message, and `311`'s diversion demonstrably worked upstream. **Nothing new was filed.**
Three things below are state this file does not yet have.

### 1. Both halves of candidate 1 are now LIVE — this file still says "INERT until the next chassis roll"

| half | state | evidence |
|---|---|---|
| Go (`item["last_error"]`) | **LIVE** | chassis rolled to **v1.0.1322 at 16:54:34Z today**. Probed on BOTH replicas with a control in the same breath: the truncation literal present = 1 on each; invented control `zz_no_such_symbol_345` = 0 on each |
| config (migration `533`) | **APPLIED** | the live `component-creator` row's prompt contains `{{if .input_data.last_error}}\nPREVIOUS ATTEMPT REJECTED — THIS IS A RETRY.` |

Wiring reads consistent end to end (the loader's item map is what the handler receives as
`input_data`, and the prompt reads `.input_data.last_error`), **but that is a code reading, not a
demand proof** — see the next point.

### 2. It has never fired, and the reason is by design — so do not read the quiet as success

`SELECT count(*) FROM orchestration_states WHERE collected_data->'input_data' ? 'last_error'` →
**0, all history.** Also 0 source-vocabulary rejections in the last 24h. That is expected and not
reassuring: the fix is gated on `attempt_count > 0`, item `95fe67da` was **cancelled at attempt 2
at 12:13Z — four hours BEFORE the roll**, and nothing has retried since.

**The 311 lane has already re-driven it:** item **`e9e5a10b-928e-411a-8488-991dadec8afa`**, created
**18:08:44Z**, `created_by='bugfix_311_redrive'`, `status=triaged`, `attempt_count=0`, same section
`mortgages-repayment`. **That item is this fix's first real test.** Its attempt 0 is byte-identical
to pre-345 behaviour by construction, so **attempt 0 failing identically proves nothing** — this
file's own success signal (a second attempt whose reason DIFFERS) can only be read at attempt 1.

### 3. The invented source was a NEAR-MISS on a real namespace, and the value it wanted exists NOWHERE

This file says (§ around the grep) that `locale` returns nothing. **That is overstated, and the
correction strengthens the diagnosis rather than weakening the guard.** Measured against the live DB:

- **No `locale` ASPECT exists** — `SELECT count(*) FROM site_specs WHERE aspect='locale'` → **0**,
  across 20+ real aspects (`classification` 1227, `identity` 253, `evidence_base` 253, … `site_config` 28).
  So the guard is right and must still not be worked around.
- **But a `locale` NAMESPACE does exist, on this very site.** `remortgagecalculator.uk` carries
  `site_specs` aspect **`site_config`** = `{"locale": {"lang": "en-GB"}}` (27 sites carry a locale
  under `site_config`; migration `508_site_specs_locale_lang.sql` put it there, addressed by the
  `config.locale.*` dialect).
- **And `currency_symbol` resolves nowhere at all.** No site carries one in structured form;
  `currency` appears only inside prose-ish aspects (`strategy` 4, `vertical_landscape` 2,
  `content_direction` 1, `identity` 1, `tools` 1) — never as a resolvable field.

**So the model was not inventing arbitrarily.** It could see a real `locale` object with `lang` in
it and reached for `locale.currency_symbol`, a plausible sibling — then addressed it through the
`site_specs.<aspect>` dialect instead of `config.`. Two errors stacked: right namespace, wrong
dialect; and a leaf that does not exist under any dialect.

**What that does to the candidates:**
- **Candidate 4 ("enumerate valid sources") would NOT have saved this one.** Listing the aspects
  tells the model `locale` is not an aspect; it does not tell it where the currency symbol is,
  because *there isn't one anywhere*. This file's argument against candidate 4 survives, but the
  reason is stronger than "site_specs is already an input": the needed value is absent from the
  data model, not merely mis-addressed.
- **Candidate 1 is the right primary**, and the guard's own message already carries what a retry
  needs (it lists the aspects that exist).
- **A fifth thing this file does not list, and it is the real one for a UK mortgage calculator:**
  a component that needs a currency symbol has **no legitimate resolvable source for it**. Whatever
  the retry writes must either hardcode `£` or drop the field. If that is the intended answer it
  should be stated somewhere the writer can read; if it is not, the data-model gap is the fix.
  Recorded here rather than filed — it is one measurement, not a diagnosed defect, and it belongs
  to whoever owns the field-source vocabulary (`bugs_open/309`'s territory).

### 4. Symptom confirmed at the artefact, for the record

`https://remortgagecalculator.uk/` today: 200 / **41,136 bytes** / **0 `<form>`, 0 `<input>`,
0 `<select>`**; the only `<button>` is the mobile-menu toggle and the only `/tools/` asset is the
lender-directory JS. Three prose sections (`brief-explanation`, `info-card-grid`, `cta`). The copy
promises the tool it lacks — `<h1>See what your payment could be after your fix ends`,
"Your number takes seconds to work out." Note the byte count has MOVED from this file's pinned
`40,726 B / md5 89910f6e…` (an unrelated index rerender ran at 17:2xZ from the CSS lane), so
**re-pin before grading the repair** rather than diffing against the old md5.
