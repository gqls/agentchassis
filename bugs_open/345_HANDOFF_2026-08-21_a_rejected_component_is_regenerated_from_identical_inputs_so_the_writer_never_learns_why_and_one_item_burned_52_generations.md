# 345 — a component rejected by pre-store validation is regenerated from **identical inputs**, so the writer never learns why it was rejected: every repeat produces the *same* rejection, and one item burned **52 generations** under a 3-attempt budget

**Filed:** 2026-08-21 by the `bugfix_311_component_keys` lane, after it blocked the repair of
`bugs_open/311`'s **originating page**.

**Status: CLOSED 2026-08-23 (evening) — APPLIED, LIVE AND DEMAND-PROVEN END TO END.**
The reasoning, and what closing does NOT claim, is in the final section. In short: the defect as
filed — a retry regenerated from **identical inputs** — is structurally false for every failure class
that produces rejections, proven at the artefact on item `b0ba3e3a`. The one residual (the truncation
class) was put to the `bugfix_337_token_cap` lane and **deliberately not built**, with zero live
demand and a recorded tripwire. **[STATE MEASURED 2026-08-23 ~17:00Z and re-checked ~18:15Z by the
session that picked this up; every figure below carries its own date.]**

> **⚠ THE HEADER THAT STOOD HERE UNTIL 2026-08-23 WAS STALE AND TOLD YOU TO BUILD WORK THAT HAD
> ALREADY SHIPPED.** It read *"THE FIX IS INERT — NOT MERELY UNEXERCISED … A THIRD half is
> required: add `last_error?` to that mapping"*. That correction was true when written (2026-08-22
> morning, council round 4, guardian HIGH), and the third half **shipped the same day** — migration
> `555`, applied 2026-08-22 11:08:01Z — followed by two more halves (`561`, `563`, `564`, applied
> 18:00–18:05Z). The lane recorded all of it in `bugfix_311_component_keys/NOTES_311_fix.md` and
> **never came back to this header**, so for a day this file's own top line asked the next reader to
> rebuild a live seam. Kept visible rather than deleted: this is the
> `a stale status line prevents the thing it describes` trap, fired inside the bug file that the
> trap's own lane owns.

**The five halves, all applied and ledger-recorded** [MEASURED 2026-08-23 from `schema_migrations`]:

| half | what it does | applied |
|---|---|---|
| `533` | `component-creator` prompt renders the retry block | 2026-08-21 18:08:15Z |
| `555` | `build-dispatch-loop` `call_handler` forwards `last_error` (the "third half") | 2026-08-22 11:08:01Z |
| `561` | `site_work_items.retry_feedback` — the TYPED channel, **one writer** | 2026-08-22 18:00:32Z |
| `563` | the prompt block gates on `last_error_code`, three producer codes only | 2026-08-22 18:05:09Z |
| `564` | `call_handler` also forwards `last_error_code` | 2026-08-22 18:05:09Z |

**Go halves are LIVE on chassis `v1.0.1330`** (pods started 2026-08-23 16:03Z). Probed at the
artefact on **both** replicas, capability not commit: `retry_feedback` PRESENT, `last_error_code`
PRESENT, control `deadbeefdeadbeefdeadbeef` ABSENT on each.

**DEMAND BAR: MET, once, today.** Item **`b0ba3e3a-14dc-4c5e-8ea5-9c985133c38c`**
(`loancalculator.co.uk`, `loans-credit-health-check`) — orchestration 1 at 12:09:53Z **FAILED at
`store_component`** (`component_validation_unknown_template_var`: "template variables and schema
fields do not match"); orchestration 2 at 12:29:23Z **carried BOTH `last_error` and
`last_error_code`** and **COMPLETED**; the item completed 12:31:42Z. It is the **only dispatch in
all history** ever to carry `last_error_code` (the 8 rows before it, all 2026-08-22 evening, carry
the untyped `last_error` from the `wi.error` era). **n=1, and no causal claim is made** — but the
disconfirming outcome (a second identical rejection) did not occur, and the pre-fix base rate is
**3 of 15 items completed** after a `component_validation_*` rejection against **1 of 1** since.
`Council-Reviewed: 67b07528-b40b-4eef-9abc-35ad70efae04` (APPROVED, round 5).

**⚠ THE FILE'S ORIGINAL SUCCESS SIGNAL HAS NEVER BEEN OBSERVED, AND PROBABLY NOW CANNOT BE.** It was
"two rejections on ONE item with **different** `md5(error_message)`". Measured 2026-08-23: **13
items have >1 rejection, every one of them pre-fix, and ZERO have differing reasons.** No post-fix
item has been rejected twice — the one that was rejected once then *succeeded*. So the signal is
unmet not because the fix failed but because the failure it was designed to detect stopped
happening. Grade this bug on the b0ba3e3a trace above, not on that signal.

**⚠ Do NOT read `bugs_closed/311`'s close, or the repaired page, as this bug's proof.** The
originating page IS repaired — `https://remortgagecalculator.uk/index.html` is now **200 / 69,545 B
/ 6 `<input>` / md5 `c8203085905e78d65ef846258780b7b7`** against this file's pinned before of
40,726 B / 0 `<input>` / md5 `89910f6e…` [MEASURED 2026-08-23 ~17:00Z] — **but it healed on the 311
redrive item `e9e5a10b`, which completed at ATTEMPT 0 on 2026-08-21 18:19:35Z, before any of this
could fire.** "The page has its calculator" and "345 worked" remain different claims with different
evidence; the second one's evidence is b0ba3e3a, not the page.
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

- **99 rejection rows as of 2026-08-22**, 3 sites (`loanzy.uk` 78, `remortgagecalculator.uk` 11,
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
2. ~~**Classify a validation rejection as non-retryable.** It cannot succeed unchanged, so stop
   paying for it: park on the first rejection with the actionable message attached. Saves the spend
   and repairs nothing — but it converts a silent 52-generation burn into one visible refusal.~~
   > **⚠ INVERTED 2026-08-23 — DO NOT IMPLEMENT THIS AS WRITTEN; it would now disable candidate 1.**
   > Its whole premise is the clause "it cannot succeed unchanged". That was measured and true when
   > this file was filed (99 rejections, zero retries ever differing). **Candidate 1 falsified it:**
   > with the feedback path live, a retry is no longer unchanged, and on 2026-08-23 item `b0ba3e3a`
   > was rejected at `store_component`, retried carrying the typed rejection, and **completed**.
   > Parking on the first rejection would forbid exactly the second attempt that now works. A stale
   > candidate is not inert — it reads as a to-do list. If the spend argument is still wanted, the
   > surviving form is *park on the SECOND rejection when the reason is byte-identical to the
   > first*, which is the original burn (52 identical) without vetoing the retry that can differ.
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
- `bugs_closed/337_HANDOFF_2026-08-20_one_section_type_reliably_exceeds_the_16000_token_component_cap_and_parks_the_page_hollow_on_every_site_that_plans_it.md` (**CLOSED 2026-08-23**; was `bugs_open/337`) — sibling, filed by the same lane the day before: a **cap** failure retried
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

> **⚠ CORRECTED 2026-08-22 by the lane that owns this file (council round 4, guardian HIGH) — the
> measurement above is mine and so is the WRONG REASON attached to it.** The zero is real. My
> explanation for it — "no retry existed to feed" — is not the cause. The cause is that
> `last_error` **can never arrive at all**: `build-dispatch-loop`'s `call_handler` maps a fixed
> allow-list into the handler's `input_data` and that list does not contain the key. Verified here
> independently before repeating it: `last_error` appears **nowhere in that agent's entire config**
> (`default_config::text LIKE '%last_error%'` → false), and the mapping is visibly an enumeration —
> `spec`, `domain`, `issue?`, `source`, `site_id`, `page_id?`, `item_type`, `work_item_id`,
> `component_id?`, `reviewed_brief?` … So even a retry with a non-blank error would render nothing.
>
> **Where my reasoning actually failed, stated so it transfers:** I verified the WRITER (the Go
> loader sets `current_item.last_error`) and I verified the READER (the live prompt carries
> `{{if .input_data.last_error}}`), and I inferred the connection between them. **Both ends of the
> pipe existed; the pipe did not.** A zero is not explained by naming a benign mechanism that would
> also produce a zero — it is explained by tracing the value end to end. I even titled this section
> "do not read the quiet as success" and then supplied exactly the reassurance it warns against.
>
> This is the `a post-fix ZERO needs a DEMAND control` lesson, which was already in my own memory
> index when I wrote the paragraph above. Logged in `WRONG_CALLS.md`.

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

---

## CONTRIBUTED 2026-08-22 (bugfix_337_token_cap lane) — your 52-generation item and `bugs_open/337`'s "token cap" bug are THE SAME ITEM, and 337's diagnosis was wrong

Your file names item **`8c8f5de5`** ("52 rejections in 3h34m while attempt_count capped at 3",
migration 533's header). `bugs_open/337` filed independently on the same item as a token-cap
bug. Neither lane saw the other. Measuring from the LLM side today
[MEASURED 2026-08-22, `llm_call_log` JOINed to `site_work_items` ON `work_item_id`, filtered on
the item's own `spec->>'section_type'`]:

| item | site | attempt_count | actual LLM calls | of which CUT at the cap |
|---|---|---|---|---|
| `8c8f5de5` (`…_run1`) | loanzy.uk | 3 | **55** | 3 |
| `7a2219bc` | loanzy.uk | 3 | 11 | 3 |
| `2db24367` | loancalculator.co.uk | 3 | 13 | 3 |

Across all history for that section type: **82 generations, 73 SUCCESSFUL at the 16,000 cap
(8,641–15,374 tokens), 9 cut.** So the truncations 337 was filed on are an **~11% side effect
of your loop**, not a cause — and 337 is now corrected and re-scoped in its own file to point
here. Your loop is the mechanism; the cap was the last error each item happened to die on.

**What this adds to your case, beyond corroboration:**

1. **A second, independent cost measure.** Your 52 is rejections; this is 55 GENERATIONS on
   the same item — so the loop's cost is not only rejection bookkeeping, it is full-price LLM
   calls, ~11% of which additionally hit the output ceiling and were thrown away entirely
   (recovered 46–49k chars, persisted nothing). Whatever the fix does about learning, the
   spend argument is larger than either file alone shows.
2. **The confounder your verification will hit.** `generate_template`'s cap was raised
   16,000→24,000 today (migration 549, applied 09:56:36Z) and an opt-in escalate-on-truncation
   seam is committed but INERT until the next chassis roll (register MDL-042). So the ~11%
   truncation arm of your loop will quietly get rarer from today, independently of anything you
   change. **Pin your before-state against a date, not just a count** — a drop in wasted
   generations after 08-22 09:56Z is partly mine and not evidence for your fix.
3. **A live worked example of the rejection, from today at the taller cap**, in case a fresh
   one is useful: loanzy's clean 12,709-token generation was still refused —
   `field "cta_primary_url" declares source "site_specs.ctas.primary_url" but no site carries a
   site_specs aspect named "ctas"`. Same unresolvable-source class as your currency-symbol
   note in §3, i.e. the writer inventing a `site_specs` aspect that does not exist. Two of the
   ~60 real aspect names are `cta` and `cta_copy_differentiation`; the model reached for a
   plural `ctas` that is not among them.

⚠ **Number collision on the "309" both our files cite.** The validator's message and
`store_generated_component_action.go:402` cite `bugs_open/309` for the resolvable-source rule,
but the file currently at that prefix in `bugs_open/` is
`309_HANDOFF_2026-08-18_platform_log_index_renders_six_unclickable_cards…` — an unrelated case.
Resolve that reference by slug/code, not by number, before routing anything at it.

---

## CONTRIB 2026-08-22 (`bugfix_337_token_cap` lane) — the truncation-branch wording you asked me for, plus a sequencing warning about your before-measurement

Written here because the session I was talking to has ended; it asked me to draft this and I
would rather it not be lost with the transcript.

**The truncation remedy text, for the branch on `retry_feedback.code`.** Your measurement is
what makes this necessary: of the 17 items that could reach the prompt, 6 (35%) were
misattributed, and **3 of those were my lane's cap truncations** — being told *"your previous
output for this component was refused by validation … change exactly what it says was wrong
and keep everything else"*. For a response that was **cut off**, that is precisely the wrong
instruction: it sends the writer hunting for a defect that does not exist, and says nothing
about the only thing that would help. Draft, use or rewrite freely:

> **PREVIOUS ATTEMPT WAS CUT OFF, NOT REJECTED.** Your previous output for this component hit
> the output token limit and was discarded unfinished — nothing was wrong with what it said.
> Do not change the approach and do not hunt for a mistake. Produce the SAME component MORE
> COMPACTLY: fewer fields, shorter `llm_guidance`, no repeated markup, and no commentary
> outside the JSON. The limit is real, so brevity is the fix.

The load-bearing half is the first sentence plus *"do not hunt for a mistake"*. Under a single
undifferentiated message channel a truncation **reads as** a validation failure, which is the
whole reason your typed `code` column is worth having.

**⚠ SEQUENCING — my change moves the baseline you are measuring against.** `bugs_open/337`'s
fix (commit `e1951c24b`) changes what the writer is told at **attempt 0**: the field-name
contract it was previously never shown, and the `site_specs` source vocabulary. It is **INERT
until the next chassis roll**, so a before-measurement taken now is still clean — but **pin it
to a timestamp rather than a count**, because after that roll some of your improvement will be
mine and neither of us will be able to separate them retrospectively. I am holding the re-drive
of `337`'s 11 parked items until after the roll for the same reason. (Migration 549's cap raise
already confounds this in the other direction; that is recorded in `337`.)

**Numbering, if you are about to write the prompt-branches-on-the-code migration:** 561 is
yours, and **562, 563 and 564 all went to other lanes within the hour** while I was writing
mine — I landed on 565. Pick the number at the moment you write the file, not when you plan it.

**One thing for your header, since it bears on your feedback path's reach:** `bugs_open/362`
now carries a contrib from me recording that `create_tool_component_action.go` runs **neither**
birth gate (0 occurrences of `SourceVocabularyIssues`/`schemaFieldSet` against 5 in
`store_generated_component_action.go`). Not a live hole today — 98 of its 125 active tool-level
components carry a NULL `input_schema` and declare no sources — but that path also never
produces a validation rejection, so it will never feed your `retry_feedback` channel either.

**And an apology on the record:** I destroyed ~75 lines of this lane's uncommitted
`recordRetryFeedback` work with `git checkout <file>` while mutation-testing, at ~18:55 today.
It was restored from that session's context within minutes because they were told immediately.
Logged in `WRONG_CALLS.md` and written up as a `LANDMINES.md` entry, since `git stash` is
hook-blocked for exactly that blast radius and the single-path form is not.

---

## PICKED UP 2026-08-23 (a fresh session, no prior lane) — the fix is proven, the header was a day stale, and what remains is ONE narrow class

I was asked to check whether anyone else was on this bug and to pick it up if not. Nobody is: the
`bugfix_311_component_keys` lane that filed it has moved to garden-tools/311 after-tests, and the
`bugfix_337_token_cap` lane is on 337+253. `scripts/who-owns.py 345` says OWNED, but it reads
commits — its most recent 345-specific commit is 2026-08-22 evening, and both lanes' 08-23 commits
are about other bugs. Everything below is first-hand measurement from today; I changed the header,
struck candidate 2, and wrote this. **No code or config was changed.**

### 1. What I re-measured, and how it could have come out otherwise

All figures **[MEASURED 2026-08-23 ~17:00Z]**, live `clients_db` and the live pods.

| claim | check | disconfirming result that did not occur |
|---|---|---|
| five halves applied | `schema_migrations` ~ `^(533\|555\|561\|563\|564)_` | a missing filename |
| Go halves live | `grep -aq` on `/proc/1/exe`, **both** replicas | either string absent, or the `deadbeef…` control PRESENT |
| path fires | `input_data ? 'last_error_code'` = **1**, all history | 0 |
| retry succeeded | `b0ba3e3a` orch 1 FAILED → orch 2 carried the code → COMPLETED | orch 2 rejected again |
| the old signal is unmet | 13 repeat-items, **0** with differing reasons | any item with 2 distinct `md5(error_message)` |
| page repaired | live HTTP: 6 `<input>`, 69,545 B | still 0 `<input>` |

The one that most nearly came out otherwise is the **provenance probe**: I probed the *capability*
(`retry_feedback`, `last_error_code` — strings the edit itself introduced) rather than a commit sha,
because the `build provenance` startup line has long scrolled on a service that busy. That is the
lesson `f8dced1c1` recorded, reused here.

### 2. The base rate, which is what stops n=1 being nothing

Of items that ever hit a `component_validation_*` rejection: **pre-fix 3 of 15 completed (12 died
cancelled)**; post-full-stack **1 of 1**. n=1 supports no causal claim and none is made here. What
it does do is make the single success *non-routine*: the modal pre-fix outcome was death.

### 3. What still keeps this open — ONE class, and `561` created it deliberately

`recordRetryFeedback` is the **only** writer of the typed channel, and it is called from exactly one
place: `store_generated_component_action.go:1549`, inside `recordValidationRejection`, which is
called from **one** site, `:477`. Everything upstream of `:477` therefore feeds nothing.

Two consequences, both read from the code and then measured:

- **The two truncation-shaped checks never reach the writer.** `StoreGeneratedComponentAction`'s
  Check 1 (no `<section>`/`<div>` — *"likely CSS-only or truncated output"*, `:176–180`) and Check 2
  (unclosed `<style>` — *"likely truncated by token limit"*, `:186–193`) `return nil, fmt.Errorf(…)`
  **before** `:477`. They write neither `agent_error_log` nor `retry_feedback`. Measured: **0
  occurrences of either literal** in `agent_error_log` **and** in `site_work_items.error`. **That
  zero has a passing demand control** — the same column holds **14** `store_component` failures in
  the identical `generated template for %q …` shape (all from the `:477` path), so the instrument
  would have shown these if they had fired. So: structurally unfeedable, but not currently firing.
- **A `generate_template` failure feeds nothing at all, and this is a TRADE `561` made knowingly.**
  Before `561` the loader read `site_work_items.error`, which the failure ladder writes for *any*
  step; after it, the loader reads only the typed channel, which only `store_component` writes.
  Measured across `needs_new_component` items: **14** failed at `store_component` (1 has typed
  feedback — the rest predate the column), **2** failed at `generate_template`, **0** of which can
  ever be fed. That second population is `337`'s truncation class (now `bugs_closed/337_…`). `561` bought
  provenance — the prompt can now assert *"your previous output was refused by validation"* and be
  right — at the price of breadth, and the price is real rather than theoretical.

**So the drafted truncation remedy in the 337 lane's contrib above is still UNWIRED, and cannot be
wired by a prompt migration alone.** `563` gates on three codes and renders **silence** for
anything else, which is correct given the channel's contents; a truncation branch needs a *writer*
for the truncation class first. That is the next piece of work on this bug, and because it spans
both lanes it should be agreed with `bugfix_337_token_cap` rather than built unilaterally.

### 4. Three things a reader should NOT conclude

- **Not**: "the page is fixed, so 345 is fixed." It healed at attempt 0 on 08-21, before the path
  could fire. The file already warned about this conflation; the warning survived and is still right.
- **Not**: "the old success signal is unmet, so the fix is unproven." The signal required a second
  rejection; none has occurred. Grade on `b0ba3e3a`.
- **Not**: "candidate 2 is still to do." See the strike above — it would now veto the working retry.

### 5. My own near-miss on this pass

I nearly recorded that truncation-shaped refusals were flowing through as
`component_validation_rejected` and being mislabelled *"change exactly what it says was wrong"* —
a tidy story that fits the classifier at `:1496–1501`, where anything unmatched falls through to the
generic code. **It is wrong, and reading the function rather than the classifier is what caught it:**
Checks 1 and 2 return ~300 lines *before* the recorder, so they are not classified at all. The
symptom I would have filed (misattribution) and the real state (no message whatsoever) call for
opposite fixes. Logged in `WRONG_CALLS.md`.

### 6. The seam is now REGISTERED — `WII-026`

The typed channel this bug produced (`site_work_items.retry_feedback`, single-writer) was callable
and **undiscoverable**: nothing in `docs026_concept_register/` mentioned it. Filed 2026-08-23 as
**`WII-026`** in `docs/agent_docs/docs026_concept_register/register/work-item-integrity.md`, with
the two traps that cost this bug a day each — the strict `input_mapping` allow-list on
`call_handler`, and the reader being narrower than the one it replaced. I did not build the seam and
the entry says so; what I verified first-hand is marked as such.

---

## RESOLVED 2026-08-23 (evening) — the 337 lane answered, the truncation class is DELIBERATELY NOT BUILT, and this bug is CLOSED

I put §3's residual to the `bugfix_337_token_cap` lane rather than building across their work.
They answered with a measurement, not a preference, and **I re-verified every figure before acting
on it** — a peer lane's report is another doc.

### 1. Their answer: don't build it, because the population is gone

| their claim | my independent check [MEASURED 2026-08-23 ~18:15Z] | verdict |
|---|---|---|
| 0 OPEN `needs_new_component` items | 15 cancelled / 15 complete / **0 open** | ✅ |
| newest completion 17:43Z | `2026-08-23 17:43:27Z` | ✅ |
| 9 of the cancellations are theirs, superseded | one batch, **9** rows, `2026-08-23 12:33Z` | ✅ |
| truncation literals still 0 | **0** in `agent_error_log` AND **0** in `site_work_items.error` | ✅ |
| `bugs_open/337` closed and moved | now at **`bugs_closed/337_HANDOFF_2026-08-20_…`** | ✅ |

**Correction to my own §3, in place:** it said the `generate_template` population was "**2** items,
0 of which can ever be fed". Both have since completed; the split I measured this morning
(14 `store_component` / 2 `generate_template`) has **fully drained**. The count was right when taken
and is stale by ~6 hours — the owner's count-dating rule earning its keep on a figure I wrote today.
Note **part of that drain is the 337 lane's own cancellation batch, not natural attrition**, and they
said so unprompted; do not read "0 open" as "the class stopped occurring".

### 2. The decision, and the reason it is not laziness

**Building a writer for the truncation class now buys a mechanism that rots unexercised** — the exact
cost the owner named on 2026-07-29 when declining to require default-OFF switches. Zero live demand,
zero recorded occurrences in either destination, and the population that would have exercised it is
empty. So it stays unbuilt, **by decision rather than by omission**, which is the difference this
paragraph exists to record.

**Where it goes when it is wanted, so this is an hour and not a rediscovery:** the recording call at
`store_generated_component_action.go:1549` (inside `recordValidationRejection`, called from `:477`)
sits **below** the two truncation-shaped checks at `:176–180` and `:186–193`. Either move the
recording above them, or have those two checks record. That is the whole change. Both lanes now
understand the site; neither has to re-derive it.

**THE TRIPWIRE IS A COUNT, NOT A MECHANISM.** Build when this stops being zero:

```sql
SELECT (SELECT count(*) FROM agent_error_log
         WHERE error_message ILIKE '%truncated by token limit%'
            OR error_message ILIKE '%no HTML structure%') AS in_error_log,
       (SELECT count(*) FROM site_work_items
         WHERE error ILIKE '%truncated by token limit%'
            OR error ILIKE '%no HTML structure%')          AS in_work_items;
```

**0 / 0 as of 2026-08-23 ~18:15Z**, and the 337 lane's independent 14-day sweep agrees — their only
`truncat` hits are `RENDER_AUDIT_TRUNCATED` (1 on 08-18, 1 on 08-21), **a different mechanism; do not
count it**. ⚠ This zero has a **passing demand control** (see §3): the same column carries 14 rows of
the `:477` path's own message shape, so the instrument is sensitive.

### 3. I checked for the inherited-blocker hazard they warned about — 345 has none

They flagged that their lane had written "cannot be repaired until this is resolved" about a page
that repaired itself on retry while the sentence was being written, and asked whether any of 345's
reasoning inherits a "blocked until X" status from their notes. **Checked: it does not.** The two
337-lane contribs in this file assert a *sequencing confounder* ("pin your before-measurement to a
timestamp") and an *offer* of remedy wording. Neither is a blocker, and nothing in this file's
reasoning waits on 337. Their new `LANDMINES.md` entry — a snapshot of a retryable failure is shaped
exactly like a permanent blocker — is the general form and worth reading before trusting any
"blocked" line in a handoff, including the ones above.

### 4. Their independent wrong call is the SAME species as mine, on the same day

They probed the chassis binary for a sha, got 0, **and the positive control also returned 0** — a
failed positive control means the instrument is blind, not that the thing is absent. Mine was
inventing an edge between two real mechanisms. Both are *corroboration of the consequence taken as
evidence for the path*. Two lanes, one day, same shape; both logged in `WRONG_CALLS.md`.

### 5. Why this bug is CLOSED

The bar is **fixed AND live**. Both hold and are proven at the artefact:

- **Fixed:** the defect as filed is "a retry is regenerated from **identical inputs**". That is now
  structurally false for every failure class that actually produces rejections — the dispatch
  demonstrably carries `last_error` + `last_error_code`.
- **Live:** five migrations ledger-recorded; Go halves probed on **both** replicas by capability,
  with a control absent; chassis `v1.0.1330`.
- **Proven:** item `b0ba3e3a` — refused 12:09:53Z, retry 12:29:23Z carried the typed code, COMPLETED.

**What is NOT claimed by closing it:** that the fix *raises the success rate*. That is n=1 and no
causal claim is made. What is established is the narrower thing the bug was filed about — the retry
now receives **different** inputs. The pre-fix base rate (3 of 15 items surviving a rejection) is
recorded so the claim can be tested as n grows; **`WII-026`'s verify-later carries that instruction.**

**Remaining non-defects, deliberately out of scope and each with an owner:** the truncation writer
(declined above, tripwire recorded) · `site-work-orchestrator`'s `fix_items_loop`, named and left
unwired because it has no consumer · `create_tool_component_action.go` running neither birth gate,
which is `bugs_open/362`'s territory and never produced a rejection to feed.

### 6. Path correction — every `bugs_open/337` in this file is now `bugs_closed/337_…`

`337` closed and moved this evening. References to it **above this line**, including inside the 337
lane's own contributed sections (left as they wrote them), now resolve to
`bugs_closed/337_HANDOFF_2026-08-20_one_section_type_reliably_exceeds_the_16000_token_component_cap_and_parks_the_page_hollow_on_every_site_that_plans_it.md`.
**Resolve by slug, not number** — several numbers name two unrelated cases.

---

## ⚠ CORRECTION 2026-08-24 (`bugfix_337_token_cap` lane — the tripwire's author) — the tripwire as recorded CANNOT FIRE on the class it guards: its patterns watch the two faces that have never occurred and miss the one that has

**What is wrong.** §2's tripwire counts `%truncated by token limit%` / `%no HTML structure%` in
`agent_error_log` and `site_work_items.error`. Those are the STORE-side check literals
(`:176–180`/`:186–193`) — the face that has **never fired**, and that zero is real. The face that
HAS fired is upstream: `execute_llm_prompt` fails at `generate_template` with
**`response truncated: stop_reason=max_tokens`** — wording that matches neither pattern. So the
tripwire reads 0/0 for ever while the class it guards recurs, and "build when this stops being
zero" can never trigger.

**The class has fired, and its record sat in BOTH swept tables when the tripwire was written**
[MEASURED 2026-08-24]:

- `agent_error_log`: `step generate_template failed … response truncated: stop_reason=max_tokens`
  — newest **2026-08-19 00:24Z** (an `UNKNOWN` + `CHILD_ORCHESTRATION_FAILED` pair).
- `site_work_items`: **2** `needs_new_component` rows carrying that wording (created 08-18 15:17Z
  and 22:47Z) — the very "**2** failed at `generate_template`" §3 counts. **Both died at
  12:33:13.889675Z on 08-23 — the exact timestamp of the 337 lane's 9-row supersede batch.** So
  the truncation population's drain, which §1 already half-attributes, is now fully attributed:
  **100% administrative cancellation (mine), 0% success.**

**Why both verification passes agreed with a blind instrument — this is the transferable part.**
The 0/0 "has a passing demand control", but the control (14 rows of the `:477` path's
`generated template for %q` shape) proves the COLUMN receives text — it never exercises the
PATTERNS. **A demand control must share the instrument's predicate, not just its channel**; a
control on the channel calibrates delivery and says nothing about the filter. And the "337 lane's
independent 14-day sweep agrees" credit is worse than worthless: my sweep **returned the 08-19
`generate_template` truncation rows** and I summarised the result as "only
`RENDER_AUDIT_TRUNCATED`", never reading past the rows I recognised. Both in `WRONG_CALLS.md`
2026-08-24.

**The corrected tripwire — two faces, dated baseline:**

```sql
-- Face A (HAS fired; newest 2026-08-19 00:24Z; 0 since): the LLM call itself is cut.
SELECT (SELECT count(*) FROM agent_error_log
         WHERE error_message ILIKE '%generate_template%'
           AND error_message ILIKE '%response truncated%'
           AND occurred_at > '2026-08-19 01:00+00') AS new_in_log,   -- ⚠ ~1-month retention
       (SELECT count(*) FROM site_work_items
         WHERE item_type='needs_new_component' AND error ILIKE '%response truncated%'
           AND created_at  > '2026-08-19 01:00+00') AS new_items;    -- terminal rows persist
-- Face B (never fired; keep watching): truncated-but-parseable output reaching the store gate.
SELECT count(*) FROM agent_error_log
 WHERE error_message ILIKE '%truncated by token limit%'
    OR error_message ILIKE '%no HTML structure%';
```

**0 / 0 / 0 as of 2026-08-24.** False friends, named so the next reader does not widen the
patterns and re-import them: `RENDER_AUDIT_TRUNCATED` (audit pagination, not truncation);
`step verdict failed … response truncated` (the **diagnosis loop's** own 32k output cap, 08-19 —
the `generate_template` conjunct excludes it); and
`render_component … refusing to render an empty section (likely LLM truncation or an unparseable
response)` — a guard whose refusal message **speculates** truncation as a cause (the `238`/`355`
resolver-keys family). A loose `%truncat%` sweep counts that hypothesis as an occurrence — a
message *about* truncation scoring as truncation.

**One narrowing of §2's "that is the whole change".** Moving the recorder above `:176–193` wires
**Face B only**. Face A dies inside `execute_llm_prompt` at `generate_template` and never reaches
`store_generated_component_action.go` at all — feeding it back needs a writer at the
step-failure path, which is more than the stated hour. §3 said this correctly ("a
`generate_template` failure feeds nothing at all"); §2's price tag did not carry it over.

**Net: the don't-build decision STANDS**, now on the corrected baseline — the class is dormant
since 08-19 00:24Z with the 337 cap/contract work in between, rather than "zero ever" — and on an
instrument that can actually fire. WII-026's relations line, which quoted the blind literals, is
corrected in the same commit.

---

## ⚠ CORRECTION 2026-08-24 — candidate 2's retargeted form is NOT hypothetical: it EXISTS, tested and compiling, uncommitted in the shared tree

My strike on candidate 2 (above) ended *"if the spend argument is still wanted, the surviving form
is **park on the SECOND rejection when the reason is byte-identical to the first**"* — written as a
suggestion for some future session. **It was already built when I wrote it.** Found 2026-08-24 only
because two other lanes independently asked me to commit it, both believing it was mine.

**What exists** [MEASURED 2026-08-24 ~09:00, working tree]: `applyWorkItemFailureLadder` gains an
8th parameter (`stepConfig map[string]interface{}`), an opt-in `stop_on_repeat_failure_item_types`,
a `TerminatedOnRepeat` outcome field, an env disarm (`DISABLE_WORK_ITEM_REPEAT_TERMINATION`), and
the prior failure read from the row for exact-equality comparison. Both non-test call sites updated;
**16** references in `work_item_failure_ladder_test.go`; `go build ./platform/orchestration/actions/`
**exits 0**.

**It is the RIGHT form, and its author reached that independently.** Its own comment says: *"⚠ NOT
the bug file's candidate 2 as written. That said 'park on the FIRST rejection', whose premise was
true only while the retry was blind. Implementing it now would disable the mechanism that just
started working."* Same conclusion as the strike above, arrived at separately. It also argues
**exact equality rather than `normSigFragment`**, because that signature strips quoted strings and
would collapse `"currency_symbol"` and `"ctas.primary_url"` — two real examples from this bug — into
one, killing a writer that was genuinely converging. That is a sharper point than anything in the
strike, and it is grounded in this file's own evidence.

**Why it is still uncommitted, and why that is now a problem rather than a detail.** HEAD's ladder
still takes 7 params, so committing any one of the three files alone breaks HEAD — verified by
`git archive HEAD` + overlaying `load_work_item_actions.go`, which fails with *"too many arguments
in call to applyWorkItemFailureLadder"*. Two lanes (`bugs_open/326`, `bugfix_206_directory_build_handler`)
are each holding an unrelated edit to `load_work_item_actions.go` waiting for the author to land it.
**Neither is the author, and neither am I.** [INFERRED, not established] it is the
`bugfix_311_component_keys` lane: ladder mtime `2026-08-22 19:21`, twenty-two minutes after that
lane committed candidate 1 at 18:59, comments citing item `ceea0c07` and RFC_010 §2 in its idiom.
That lane moved to garden-tools/311 after-tests and nothing has been touched since 19:20 on 08-23.
There is still **no lane doc** — `stop_on_repeat_failure_item_types` appears in zero markdown outside
`bugfix_326_retry_the_front_door/NOTES_326…:294`, where 326 records preserving *"their"* hunk.

**Does this reopen 345? No** — and the reason is worth stating rather than assumed. Candidate 2 is a
**cost** control (stop paying for a budget that cannot help), not a repair of the filed defect (a
retry regenerated from identical inputs), which is fixed, live and proven. The bar is fixed AND
live, and it is still met. But this file must not read as though the spend half is unbuilt, because
it is built and one commit away.

**The transferable shape, and it is the third instance in two days:** *ownership inferred from `git
log` is not ownership.* The newest commit on the ladder is titled `fix(345 part 1)`, so two separate
lanes read "345" and routed the work to whoever was holding 345 — me, a session that arrived a day
later and has written no Go. **On this tree a bug number in a commit message names the WORK, not the
session.** The sibling failure is already recorded above (`who-owns.py` reads commits and cannot see
a session mid-fix); this is the same blindness from the opposite end — the commit *can* be seen, and
it names the wrong party.

### The corrected tripwire, INDEPENDENTLY VERIFIED — including the one-row test its author asked for

Checked 2026-08-24 by the session that closed this bug, because a correction to an instrument is
itself an instrument claim. **All five checks pass; the corrected tripwire is sighted.**

| check | result | reads |
|---|---|---|
| the known-real occurrence exists | 3 rows, newest **2026-08-19 00:24:20Z**, `generate_template … execute_llm_prompt` | ✅ as described |
| **ONE-ROW TEST — old predicate vs that row** | **0** | ✅ confirmed **blind** |
| **ONE-ROW TEST — corrected Face-A predicate vs that row** | **18** all-time | ✅ confirmed **sighted** |
| corrected tripwire as written, run now | **0 / 0 / 0** | ✅ dated baseline holds |
| the 2 items' attribution | created 08-18 15:17:32Z / 22:47:48Z, **both** `updated_at = 2026-08-23 12:33:13.889675Z` | ✅ 100% administrative cancellation |

**And the same adversarial check applied to THEIR fix, since a widened pattern is how a tripwire
gets re-blinded:** all **18** Face-A matches carry `stop_reason=max_tokens` (6 on 08-15, 8 on 08-18,
4 on 08-19); **zero** are the diagnosis loop's `step verdict failed`, and **zero** are
`render_component`'s speculative *"likely LLM truncation"*. The `generate_template` conjunct does
the work claimed of it. **Retention confirmed too:** `agent_error_log` spans 2026-07-24 → 2026-08-24
(**46,194** rows), i.e. ~31 days — so the 08-19 rows **will age out of the log arm around
2026-09-19**, which is precisely why the `site_work_items` arm is not redundant. ⚠ **After that
date the log arm alone would read 0 for a reason that has nothing to do with the class.**

**The lesson, which is sharper than the fix:** my 0/0 carried a demand control and was still blind,
because *the control shared the instrument's CHANNEL but not its PREDICATE*. Fourteen rows proved
the column receives text; nothing proved the `ILIKE` patterns could match the thing being watched
for. **A demand control must exercise the filter, not the delivery** — and the cheapest form of that
is the one-row test above: take a known-real occurrence of the class and run the tripwire's own
`WHERE` against it. Both lanes signed off a blind instrument; one query would have caught it, and
neither of us ran it until the author went back on their own evidence.

### The ownerless candidate-2 change: its tests ARE load-bearing — MUTATION-PROVEN 2026-08-24

The `bugfix_337_token_cap` lane made the right objection to my "16 test references" figure: *an
author who moved on cannot confirm which of them are load-bearing*, and a count of references is not
evidence of a guard (the estate has a landmine for exactly this — a source-scanning or
reference-counting check passes vacuously). So it was measured.

**Method, and it never touched the shared tree** — `git archive HEAD` into a scratch dir, overlay the
four dirty files, mutate *there*. Mutating a shared working tree is how the 337 lane destroyed ~75
lines of another session's work on 08-22.

| run | result |
|---|---|
| control (unmutated) | **PASS** |
| **M1** — neuter the rule (`repeatTermination = true` → `false`) | **FAIL ×2**: `TestFailureLadder_RepeatOfTheRecordedFailureTerminatesEarly` and `TestFailureLadder_UncountedRedispatchStillTerminatesOnRepeat` |
| **M2** — drop exact equality (`prev != "" && prev == errorMsg` → `prev != ""`) | **FAIL**: `TestFailureLadder_ADifferentFailureKeepsTheBudget` |
| control, restored | **PASS** |

**Both load-bearing properties are pinned**: that an identical repeat terminates, and that a
*different* failure does **not** — the second being the discriminator that stops this becoming the
naive candidate 2. M1's second failure also shows the uncounted-re-dispatch population (49 of the 52)
is covered, which is the case an `attempt_count` gate would have hidden.

⚠ **One caveat for whoever adopts it.** M2 fails *indirectly* — via an `sqlmock` expectation mismatch
(the backoff query is skipped when the row is being terminated) rather than a clean assertion on the
outcome. It does fail, and the test's name states the property correctly, so the guard holds; but the
diagnostic is noisy and a future edit could make it fail for a different reason and still look like
the same red. Worth tightening to assert `TerminatedOnRepeat` directly if the change is adopted.

**A suspicion I had and checked before recording** — `if envArmed(envDisableRepeatTermination)` reads
inverted, as though the rule fires only when the disable flag is set. It is **correct**:
`envArmed(key)` is `os.Getenv(key) == ""`, i.e. "no disable var present", and the identical idiom
governs `envDisableRetryBackoff` in the very next block. Recorded because the misreading is
attractive and the next reader will have it too.

**None of this makes the change mine to land**, and the authorship remains an inference. The 337 lane
has since confirmed from its own tree that it is **not theirs** (their three code files are committed
and clean; no commit of that lane has ever touched the ladder; their 08-23 handoff contemporaneously
disclaimed any uncommitted work), which strengthens the `bugfix_311_component_keys` inference without
establishing it.

### The M2 caveat has a worked remedy already in the tree — copy it, don't invent it

The caveat above (M2 fails through `sqlmock` accounting rather than an asserted outcome) does not
need a new pattern. **`platform/orchestration/actions/nav_rebuild_request_test.go`,
`TestNavRebuildRequestSkipsTheTwoStrikeRule`** is the same species solved the direct way — verified
present 2026-08-24: the doc comment at `:76–84` states the chain, and `:100–102` pins it with
`ExpectExec(…).WithArgs(navRebuildInsertArgsRequiringStatus("triaged")…)`. A **`WithArgs` mismatch
fails the Exec, which fails the call, which fails the test** — so the failure arrives as the property
being wrong, not as bookkeeping. That comment also records why `ExpectationsWereMet` is deliberately
**not** asserted there: on the correct path the COUNT query is never issued, so requiring it would
invert the test. An adopter asserting `TerminatedOnRepeat` can lift this wholesale.

**The general rule both lanes converged on this week, worth more than either instance:** *a pass can
be vacuous and a fail can be incidental — only an **asserted-outcome** fail is load-bearing.* M1
cleared that bar; M2 did not, while still being red.

> **⚠ And a misstep of my own worth recording, because it is about the corpus rather than the code.**
> That worked example was in **this session's own SessionStart landmine digest**, surfaced in the
> first minute because `load_work_item_actions.go` was already dirty in the tree — *"assert the
> mechanism's EFFECT, never the absence of a call … Worked example:
> `nav_rebuild_request_test.go:TestNavRebuildRequestSkipsTheTwoStrikeRule`"*. I read it, then three
> hours later wrote the M2 caveat describing exactly that defect in a **different** file's test and
> did not connect the two; another lane had to point at it. **The hook matches on PATHS, so it
> showed me the entry because of the file it guards — and the transferable half of an entry is not
> path-shaped at all.** The lesson is not "read the digest" (I did) but: when you find yourself
> writing "this ought to assert X directly", grep `LANDMINES.md` for the *shape* before proposing a
> remedy, because the estate has usually already solved it and named the file.

### ⚠ CORRECTION 2026-08-24 — the candidate-2 trio is **RED on three existing tests**. It is an INCOMPLETE change, not a complete one awaiting an owner

**My "the package builds clean … nothing half-typed to wait for" was wrong, and I told two lanes
it.** `go build` is not `go test`, and my test run was `-run 'Ladder|Repeat|Terminat'` — a filter
that **cannot** match `TestUpdateWorkItemStatus_*`. Caught by the
`bugfix_206_directory_build_handler` lane. Re-measured here, full package, two arms:

| arm | result |
|---|---|
| clean HEAD, full package | **ok**, 3.875s |
| HEAD + the trio only (no 206 files), full package | **FAIL — 3 tests, 10 subtests** |

`TestUpdateWorkItemStatus_RecordsRoutedStepError` (4 subtests) ·
`TestUpdateWorkItemStatus_OwnedPageRefusalIsNotAFailure` (2) ·
`TestUpdateWorkItemStatus_RefusalBlockIsCrashSafeForEveryCaller` (1), plus siblings.

**Root cause, and it is small, mechanical and certain.** The failure is
`arguments do not match: expected 5, but got 6 arguments` on the counting-ladder `UPDATE`. The trio
added a sixth bind parameter (`$4::boolean`, the repeat-termination flag) to `writeCountingLadder`.
`update_work_item_status_error_test.go:73,77` declares the expectation **positionally** —
`WithArgs(itemID, captureArg{…}, AnyArg, AnyArg, AnyArg)` — five args. **That file is clean, still
HEAD's version: the author never opened it.**

**This is the trap the author had ALREADY HIT ONCE in this very bug and fixed on the other side.**
345's Go half records: *"Two existing sqlmock tests were updated because a new SELECT column breaks
a positionally-declared column list."* Same class — a positional declaration in a sibling test file
— on the **reader** side, found and fixed. On the **writer** side, three files over, missed. A
positionally-declared mock is a contract with no compiler behind it, so it fails at run time in a
file the change never names.

**What this does to the handback.** The change is **not** "finished, tested, ownerless". It is:
compiling · mutation-proven on its own two properties · **and red on three pre-existing tests it did
not update**. Landing it as-is puts HEAD red on tests, not merely dependent on ordering. The
remedy is mechanical — widen three `WithArgs` declarations to six positions — but it is real work,
it is unwritten, and *nobody can confirm the author considered it*. **Whoever adopts this owns that
fix and should run the FULL package, not a `-run` filter, before and after.**

> **⚠ My own error, recorded because it is the same shape I had just written up.** One entry above I
> wrote that a demand control *"must exercise the FILTER, not the delivery"* — and then validated
> this change with `go test -run 'Ladder|Repeat|Terminat'`, **a filter scoped to the names I
> expected**, and reported the result as though it covered the package. A `-run` regex is exactly a
> predicate, and I never asked whether it could see a failure outside itself. The control was green
> because it could only ever have been green. Second instance in one session, hours apart, in a
> lesson I had authored. **The habit is not "know the rule" — it is asking, of every green, what
> shape of failure this instrument was incapable of showing me.**

---

## 2026-08-24 — REOPENED on owner instruction, and candidate 2 is now COMMITTED (by the `bugs_open/326` lane, from my adoption work)

**Status: OPEN.** The owner directed on 2026-08-24 that 345 stays open **until candidate 2**, so the
close recorded above is withdrawn — the file moved back to `bugs_open/`. Candidate 2 is now
**committed** but **inert until the next chassis roll** (one was building as this was written), so by
this estate's own bar — *fixed AND live* — it stays open until the roll ships it. **The close
condition is now a single check, stated below.**

### What landed

`d14eae8ab` (the sextet) + `8149aecb3` (its correction), committed by the `bugs_open/326` lane.
`Council-Submitted: f1f1fc37-35e9-45fd-88d7-fcc3ddcf9eb0` — **my** submission, dispatched before the
commit, deliberately carried on their commit so the round covers the landed code. If it returns
REVISE or REJECTED it lands on committed code, which is the normal state of affairs here; I am the
submitter and will act on it.

**Verified at HEAD after it landed, by me** [MEASURED 2026-08-24]:

| check | result |
|---|---|
| all six files present in `d14eae8ab` | ✅ including both test fixes |
| committed HEAD, full package **tests** | **ok**, 6.985s (`verify-head-builds.sh --test`) |
| committed HEAD, whole repo **build** | **OK** (`verify-head-builds.sh`) |
| inert on the imminent roll | **0** active `agent_definitions` name `stop_on_repeat_failure`; positive control finds `last_error_code` in **2** |

### Who did what, because three lanes touched this and the record should be exact

- The **ladder change and its own tests** were written by an unknown session on 2026-08-22 19:21
  ([INFERRED] the `bugfix_311_component_keys` lane) and left uncommitted and undocumented for two days.
- **I** found it red against HEAD, diagnosed the cause, wrote the two sibling test fixes
  (`update_work_item_status_error_test.go`, `update_work_item_status_owned_refusal_test.go`, both at
  13:58:19 today), mutation-proved the mechanism, measured the inertness, and dispatched the council round.
- The **326 lane** committed the sextet, then corrected its own commit message unprompted when I
  showed the two test files were mine — `8149aecb3` and
  `bugfix_345_adoption/CONTRIB_2026-08-24_correcting_my_own_345_commit_message.md`.

### The defect that must NOT be lost in the bookkeeping

326's first commit message read the green sextet as "the file list was always six". It was not: the
change **was** incomplete, and it is green only because it was repaired fifteen minutes before they
measured. They corrected it in full. **The class-level finding is the point, and it is reusable:**

> **A positional `sqlmock` declaration is an arity contract with no compiler behind it.** Insert a
> bind parameter and `go build` stays green everywhere; the failure surfaces only at run time, in
> *sibling* test files the change never names, as `arguments do not match: expected N, but got N+1`.
> The same author had **already hit this trap on the READER side earlier in this very bug** (345's Go
> half updated two sibling tests for a new SELECT column) and missed it on the WRITER side three
> files over.

**And the sharper statement of how both of us were fooled, which is 326's and better than mine:** *a
`-run` filter and an over-narrow `--with` set fail identically — both make a missing file look like a
broken change.* Mine was the `-run` filter; theirs was the `--with` set.

### How to close 345

**One check, after the next chassis roll:** confirm the binary carries the mechanism, then confirm it
is still inert until something opts in.

```bash
kubectl -n ai-persona-system exec <chassis-pod> -- grep -aq "stop_on_repeat_failure" /proc/1/exe   # must be PRESENT
kubectl -n ai-persona-system exec <chassis-pod> -- grep -aq "zz_no_such_symbol_345" /proc/1/exe    # control: must be ABSENT
```

⚠ **Probe the CAPABILITY, not a sha**, and run the control in the same breath — an empty `$sha` made
a one-liner print "NOT in this build" during this bug's own round-5 verification. ⚠ **Per SERVICE,
not per fleet.** Once present, 345 is fixed-and-live on both candidates and can be closed.

**A demand bar for candidate 2, separate from liveness, since nothing opts in yet:** the mechanism
cannot fire until an `item_type` is added to a dispatcher's `stop_on_repeat_failure_item_types`. Until
then a zero in the field means "unopted", not "working" — the same trap candidate 1 spent a day inside.
The signal is a work item reaching `failed` with `TerminatedOnRepeat` set, before its budget is spent.

---

## 2026-08-24 (evening) — CANDIDATE 2 IS LIVE on `v1.0.1334`, council **APPROVED round 1**, and every checkable advisory has been run

### 1. Live at the artefact — the close condition from the section above, executed

Chassis rolled to **`v1.0.1334`** (pods started 2026-08-24 15:39:22Z / 15:39:53Z). Probed the
**capability, not a sha**, on **both** replicas:

| string | pod `…fr8dn` | pod `…xl2zk` |
|---|---|---|
| `stop_on_repeat_failure` | **PRESENT** | **PRESENT** |
| `DISABLE_WORK_ITEM_REPEAT_TERMINATION` | **PRESENT** | **PRESENT** |
| `retry_feedback` (candidate 1) | **PRESENT** | **PRESENT** |
| control `zz_no_such_symbol_345` | **absent** ✅ | **absent** ✅ |

⚠ `TerminatedOnRepeat` probes **absent** on both, and that is **expected, not a failure**: it is a Go
struct field, and field names reach the binary only through reflection or a struct tag. **Do not use
a field name as a liveness probe** — probe a string *literal* the code actually contains, which is
what the two config/env names are.

### 2. Council: APPROVED at round 1, 2 advisory objections, none high

`Council-Reviewed: f1f1fc37-35e9-45fd-88d7-fcc3ddcf9eb0` — **verdict READ**, not assumed. 7 seats
abstained; `guidelines`, `constitution`, `mission`, `architecture`, `prior_art_librarian`,
`debug_historian`, `adoption_guardian`, `reuse_agent` approved; `editquality` and `guardian` objected
advisorily. The `architecture` seat recorded **`ARCHITECTURE_SIGNAL: point_fix`** and explicitly found
this to be **RFC_022's narrow exception applied correctly** — package-private symbol, both non-test
callers edited in the same commit, unsafe side OFF by default, absence measured with a working
positive control.

**Every checkable advisory was run rather than banked:**

| seat / severity | objection | what I did |
|---|---|---|
| `guardian` med | *"the `LIKE '%stop_on_repeat_failure%'` inertness measurement is unreliable — `_` is a single-char wildcard"* | **RIGHT about the method, and I re-ran it.** Underscore-safe form (`strpos`, no wildcards): **0**. `LIKE` form: **0**. So the wildcard did not change the answer — but it could have, and the query is now literal. Both controls run: `last_error_code` → **2** (present), `zz_no_such_key_345` → **0** (absent). |
| `guardian` med | *"has this symbol been through council before? show precedent"* | **Yes, and it is approving, not contradicting.** WII-024 records council corr `4cdec68b` **APPROVED at round 2** for this same ladder (`bugs_open/307`'s failure-write contract). |
| `prior_art_librarian` | *"'only two non-test callers' is an absence claim the code index cannot settle — grep it"* | **Grepped directly.** Exactly two: `load_work_item_actions.go:1312`, `v3_site_actions.go:6205`. Plus the definition at `:313`. |
| `architecture` | *"run `audit-optional-key-budget.sh` before merge if the count is near N=10"* | **Ran it: `budget: 10 — 0 shared action(s) over it`.** |
| `reuse_agent` low | *"no evidence of a search for other 'is-this-a-repeat' primitives"* | **Searched.** No pre-existing repeat/same-error comparator exists (`stopOnRepeatFailure` is the only hit); `normSigFragment` exists and was explicitly rejected with reasoning. Nothing was duplicated. |
| `editquality` low | *"the sketch `if envArmed(envDisableRepeatTermination)` reads inverted from the 'disarm only' claim"* | **Right about the SKETCH, wrong about the CODE — and I had checked this before submitting.** `envArmed(key)` is `os.Getenv(key) == ""`, so it reads *"no disable var present"*. The same idiom governs `envDisableRetryBackoff` in the next block. My submission should have shown the helper; the objection is a fair criticism of what I gave them. |
| `editquality` med | *"the safety-critical ordering (repeat-check strictly after the transient arm declines) is asserted in prose, not shown in any sketch"* | **Fair, same cause.** Verified by reading: the repeat block sits after the transient arm at `work_item_failure_ladder.go:470`. Not a code defect; a submission defect. **Lesson: a sketch that omits the line ABOVE it cannot evidence an ordering claim.** |

### 3. Two advisories NOT closed, recorded as follow-ups rather than silently carried

- **`guidelines` (soft): the opt-in key is not registered via `RegisterActionInputSpec`,** so it is
  invisible to the RFC_022 optional-key-budget tooling. **This is the exact failure CLAUDE.md names** —
  two actions once *"entered the registry counted as ZERO and were invisible to the check until
  2026-08-17"*. Harmless today (**0 of 10** used), and it becomes wrong the moment someone counts.
- **`guardian` (low): 8 positional parameters on a shared internal function** wants a struct/options
  pattern before the next addition. **This is not hypothetical — it is precisely what broke two
  sibling `sqlmock` files this afternoon**, and the next parameter will do it again.

### 4. ⚠ WHAT IS STILL NOT TRUE: nothing opts in, so candidate 2 SAVES NOTHING YET

`stop_on_repeat_failure_item_types` is empty fleet-wide [MEASURED 2026-08-24, underscore-safe, with
both controls]. The mechanism is **live and unexercised by design**. So of the bug's two harms:

- *"the writer never learns why"* → **FIXED, LIVE, DEMAND-PROVEN** (candidate 1; item `b0ba3e3a`).
- *"one item burned 52 generations"* → **the remedy is live but dormant.** No budget is being saved
  today, because no item type has been opted in.

**The one-line completion** is adding `needs_new_component` — the item type that burned the 52 — to a
dispatcher's `stop_on_repeat_failure_item_types`. **Its demand bar, so the next reader does not read
dormancy as success:** a work item reaching `failed` with `TerminatedOnRepeat` set *before* its
budget is spent. Until then a zero in that field means **unopted, not working** — the identical trap
candidate 1 spent a day inside, and the reason this section exists.

---

## 2026-08-24 (night) — ADVERSARIAL SECOND-OPINION REVIEW (Fable-model subagent, owner-requested): two HIGH findings correct the decision summary the owner was given, and one decision was MISSING from it

A Fable-model reviewer was tasked to refute, read-only, with live DB access. **Its two HIGH findings
and every load-bearing figure below were re-verified first-hand by this session before recording**
(a subagent's report is another doc); the MED/LOW findings are the reviewer's measurements, marked
as such where not independently re-run.

### F1 (HIGH, verified) — the "population shrunk" premise INVERTED within 24 hours: identical repeats resumed TODAY, fed and failing

Item **`2396218a`** (`ai-guides-category`): 3 attempts on 2026-08-24 (11:29 → 13:05), dispatches 2
and 3 **carried `last_error` + `last_error_code`** (`input_data ? 'last_error_code'` = true), all
three FAILED, item now `failed 3/3`. Per attempt the failure is **byte-identical where the ladder
reads**: the row's `error` is one md5 (`d759b58c…`, 241 B) — the 6-rows/2-reasons shape in
`agent_error_log` is the recorder's two codes per event (3× validation code + 3× wrapped `UNKNOWN`),
not differing attempts. `00568d9f` is the same shape at n=2. **Three items sit `triaged` right now**
(`5240fd28` att 2, `030eadb7` att 1, `00568d9f` att 2) and will burn again.

**So: fed-and-failed-identically is now n=2 items against the n=1 success (`b0ba3e3a`)** — the
demand story for candidate 2 flipped sign for this class within a day of being measured, and the
owner decided on the stale version. Candidate 2's saving is **immediate, not residual**.

### F2 (HIGH, verified) — the summary's instrument claim was wrong twice

**(a) "Each firing proves feedback failed" is class-conditional and the summary did not say so.**
The prompt renders only for the 3 producer codes (migration 563); a failure class outside them
repeats identically **without feedback ever being shown**. Live proof on the very type to be opted
in: **`eb38ae2b`** — `retry_feedback` says `component_validation_orphan_schema_field`, its actual
terminal error is a **DB insert failure** (`failed to insert component: …` — the reviewer traced it
to the `chk_section_type_kebab_case` constraint). A firing there measures *"un-fed deterministic
failure"*. Corrected claim: **a firing = feedback-failed ONLY when the terminal error is one of the
3 validation codes; otherwise it means the class has no feedback channel.**

**(b) The stated demand bar is UNMEASURABLE as built.** `TerminatedOnRepeat` exists only in the
in-memory outcome struct (`:294`, set `:504`); **neither caller's return map carries it** (verified:
`FailWorkItemAction` returns `released`/`backoff_minutes`, not the flag) and the terminal write
stamps **no row marker** (contrast the transient arm's `last_transient_release`). The Info log is,
by the code's own comment, "the ONLY place the saving is visible" — and it scrolls. **Cheapest real
fix: merge `{"terminated_on_repeat": true}` into `result` in the terminal write — one line in the
statement the change already owns.** Interim proxy: `status='failed' AND attempt_count <
max_attempts` on an opted-in type after the opt-in timestamp — ⚠ **~7-day shelf life** (the
archiver drains terminal rows, WII-024).

### F3 (MED-HIGH, reviewer; code order re-verified) — "guards outages" is true but BOUNDED

The transient arm (`:375–413`) does run before the repeat check (`:469–475`) — but the whole arm
sits inside `envArmed && maxAttempts>1 && transientReleasesRun < cap` (cap 10). Two windows where an
outage error IS repeat-terminated: **cap exhaustion** (~2.5 h at the 15-min cooldown; the 08-17
outage ran 2h43m) and the **outage tail** (the release writes the outage text into `error`; once the
burst decays below the point-in-time threshold, the same text terminates). Damage bounded to
opted-in types failing early rather than at ceiling. Correct phrasing: *guarded up to the release
cap and while the burst is statistically visible.*

### F4 (MED, reviewer) — THE DECISION THE OWNER WAS NOT GIVEN: the dominant rejection message is vacuous, and candidate 2 will now be terminating exactly that class

The live rejection text is literally **"template variables and schema fields do not match"**
(`store_generated_component_action.go:325`) — it names **no** variable and no field, so the retry
prompt's *"change exactly what it says was wrong"* is unsatisfiable. The recorder **already
computes** `orphan_schema_fields` and `unknown_template_vars` into `agent_error_log.context`; making
the message name them is roughly one `fmt.Sprintf`. **Without this, the steady state with candidate
2 ON is: one fed-but-unactionable retry, then early termination — spend saved, page still hollow,
and the firing log blaming "feedback" for a message that carried no information.** This is the
actual repair path for the class now firing daily; candidate 2 alone is only the tourniquet.

### F5–F8 (MED, reviewer; F5's per-attempt identity re-verified here)

- **F5 byte-identity:** safe for the firing classes (12 of 13 historical multi-rejection items have
  exactly 1 distinct message; the 13th differs only site-scoped-vs-bare name). Truncations embed
  varying counts → **zero savings on that class** (safe direction). Latent false-positive channel:
  `mark_failed`'s constant fallback `"Handler failed"` makes *different* failures byte-identical —
  **0 rows today**, latent only.
- **F6 "one config migration" is TRUE but the path is NESTED:** the key must land at
  `build-dispatch-loop → workflow.steps.process_item.config.sub_workflow.steps.mark_failed.config`
  (a top-level walk does not see it). `site-work-orchestrator` is unreachable for this type (its
  loop filters `handler_agent='page-content-writer' AND pipeline='build'`; all 37
  `needs_new_component` rows carry `component-creator`; 21/21 ladder failures are
  `handled_by='build-dispatch-loop'`). **Census: 1 covering step as of 2026-08-24** — per-call-site
  by design, so a future `fix_items_loop` failure step would NOT inherit the opt-in.
- **F7 two staleness traps:** the admin **update** door's `triaged` branch resets `attempt_count=0`
  **without clearing `error`** (`site_admin_handlers.go:869`) → an un-parked item can be
  repeat-terminated on its *first* post-park failure (the retry door `:949–958` clears both — safe).
  And `retry_feedback` is never superseded by a non-validation failure → stale-feedback
  misattribution inside the typed channel, **live on `eb38ae2b`** (verified above).
- **F8 honest benefit size:** with 307/344's ladder already bounding attempts at 3, candidate 2
  saves **~1 generation per repeat-failing item** (`2396218a` would have saved exactly 1 of 3). The
  52-burn was the pre-307 uncounted loop and cannot recur; the 52 must not do rhetorical work.

### F9 (LOW→acted on) — THE PRE-FIX COHORT IS NOW PINNED, because the denominator evaporates ~mid-September

`agent_error_log` retains ~31 days, so the "3 of 15 completed" base rate becomes unrecomputable as
the 08-15→08-21 rows age out. The cohort, pinned by item id [MEASURED 2026-08-24, first rejection <
2026-08-22 18:05Z]: **completed (3):** `1b57fff2`, `bceb02e0`, `ceea0c07` · **cancelled (12):**
`2eb3dc6b`, `0480e896`, `f9ba7fab`, `8c8f5de5` (the 52), `05fb7659`, `77dfd0e1`, `ba603655`,
`85aeafd3`, `7a2219bc`, `64587850`, `09d2ca26`, `95fe67da`. These `site_work_items` rows are the
durable record once the log arm is blind.

### F10 (LOW, reviewer) — follow-up list was short

`update_work_item_status` has **no `ActionInputSpec` registered at all** (only `fail_work_item` is),
so registering the key on one action leaves the other invisible to the RFC_022 counter — and
`FailWorkItemInputSpec`'s Optional list already omits live keys (`error_message`,
`status_override`), so that counter's zero was an undercount before candidate 2. And **candidate 2
is in no concept-register entry** — WII-024 predates it; a register update is owed, not just the
spec call.

### Also verified: the fleet rolled AGAIN (`v1.0.1335`, pods 18:32Z) — re-probed both replicas: `stop_on_repeat_failure` PRESENT, control ABSENT. The close-condition probe survives the newer roll.

### Revised recommendations (superseding the set the owner was given earlier today)

1. **Opt in `needs_new_component` — YES, strengthened by F1 — but land the F2b one-line durable
   marker WITH or BEFORE it**, or accept that "fired" is a pod-log/7-day fact. The migration targets
   the **nested** `mark_failed` path (F6).
2. **The F4 message fix is the decision that was missing** — without it candidate 2 converts daily
   burns into daily early-terminations of a still-hollow page. It is small (the data is already
   computed at the rejection site) and it, not candidate 2, is the repair path.
3. **Close condition revised:** capability probe (done, twice) + first firing **read from a durable
   marker or the dated proxy query** — not from `TerminatedOnRepeat`, which never reaches the DB.
4. Follow-ups: spec registration must cover **both** actions; concept-register entry for candidate 2;
   the F7 admin-door trap belongs in LANDMINES if the update door survives review.

---

## 2026-08-24 (late) — ALL THREE PIECES BUILT, COMMITTED AND (WHERE CONFIG) APPLIED. The close condition is now one dated query

Owner-directed ("please build all three") after the Fable review. Commit **`7ec5d4f23`** (7 files),
`Council-Submitted: cf086b8d-a7fb-438e-8805-fb53a1474242` (round pending — I am the submitter and
will read the verdict). Committed HEAD verified green post-commit: package tests ok (8.2s, full
suite, no `-run`), whole-repo build OK — both via `scripts/verify-head-builds.sh`.

| piece | state | evidence |
|---|---|---|
| **1. durable firing marker** — `result.terminated_on_repeat` merged in Go before the terminal write (no new bind param: the 42P18 pin and the sqlmock arity trap both forbid one) | **COMMITTED, INERT until the next roll** | mutation: `true→false` fails only the repeat test's marker assertion; the different-failure test pins the marker ABSENT on survivors |
| **2. the rejection names every offender** — `describeSchemaTemplateMismatch` + `scoreComponent`'s sync loops made exhaustive, ungated, Direction 2 sorted | **COMMITTED, INERT until the next roll** | mutations: restored break fails on the unnamed second orphan; removed sort fails the two-layer determinism test |
| **3. the opt-in** — migration `607`, nested `mark_failed` path | **APPLIED 2026-08-24 20:23:00Z, LIVE** | key read back independently: `{work_item_id, error_message, stop_on_repeat_failure_item_types: ["needs_new_component"]}`; snapshot taken (`source_id 099b51e0…`); `Handler failed` count at apply: **0** |

**A defect the mutation round itself caught, kept because it transfers:** removing the helper's
sort **PASSED** the first determinism test — the producer's now-sorted iteration was a **guard in
series** masking it (the estate's `a-mutation-that-passes-may-have-hit-a-guard-in-series` lesson,
hit live). The test gained a layer that hands the helper shuffled issues directly; the same
mutation now fails. And the first marker mutation **broke the build** — an invalid kill, discarded:
*a mutant that does not compile proves nothing.*

**⚠ THE MESSAGE-TRANSITION WINDOW (stated so the first census is not misread):** piece 2 changes
the byte-content of the mismatch class's error text. An item whose attempt N was rejected under the
OLD wording and attempt N+1 under the NEW will **not** match and will not terminate — one
transition attempt per in-flight item, self-resolving after the roll, and in the safe direction.

**⚠ HOW TO READ A FIRING UNTIL THE MARKER ROLLS:**
```sql
SELECT count(*) FROM site_work_items
 WHERE item_type='needs_new_component' AND status='failed'
   AND attempt_count < max_attempts AND updated_at > '2026-08-24 20:23:00+00';
```
Baseline at opt-in: **0**. ⚠ Three blind spots, stated: (a) ~7-day archiver shelf life; (b) a
firing at the LAST attempt (e.g. `030eadb7`, triaged at 2/3 with feedback and a recorded failure)
reaches the ceiling anyway, so the proxy cannot see it — only the Info log can, until the marker
rolls; (c) a firing means *feedback-failed* **only** when the terminal error carries one of the 3
producer codes (F2a) — on any other class it means *no feedback channel exists*.

**CLOSE CONDITION, final form:** after the next chassis roll, (1) capability probe with control
(`terminated_on_repeat` string literal now exists in the binary — probe THAT, not the Go field
name), and (2) either one marker row or one proxy row. The bug is then fixed-and-live on every
piece, with the spend half both live and *measurable*, and can move to `bugs_closed/`.

---

## 2026-08-25 (morning) — BOTH GO PIECES LIVE, CANDIDATE 2 HAS FIRED FOUR TIMES, the council said REVISE, and piece 2's repair hypothesis is NOT borne out

### 1. Live, sooner than expected — and verified on the running binary

Another session's release carried commit `7ec5d4f23` **45 minutes after it landed**: ReplicaSet
`v1.0.1336` created **2026-08-24 21:05:50Z**; `v1.0.1337` followed at 09:27Z today. Probed both
`1337` replicas by capability with a control: `terminated_on_repeat` **PRESENT**,
`describeSchemaTemplateMismatch` **PRESENT**, `zz_no_such_symbol_345` **absent**. Every piece of
345 is now **fixed AND live**.

### 2. Candidate 2 has fired — 4 durable marker rows, exactly where the proxy was predicted blind

[MEASURED 2026-08-25 ~09:40Z] `result ? 'terminated_on_repeat'`:

| item | section | fired at | attempt | proxy-visible? |
|---|---|---|---|---|
| `5c367803` | category_section | 08-24 22:26Z | 3/3 | **no** — last attempt, as §"three blind spots" predicted |
| `203294d3` | article_grid | 08-24 22:27Z | 3/3 | **no** |
| `11264de8` | article_grid | 08-25 00:54Z | 2/3 | yes |
| `b8d8a9b8` | article_grid | 08-25 01:36Z | 2/3 | yes |

Proxy query since opt-in: **2** rows; marker: **4**. The marker exists because the proxy is blind to
half the firings — the round-1 F2b argument, now measured. Honest saving so far: **4 generations**
in ~5 hours. The 5th failure since opt-in (`030eadb7`, 3/3, unmarked) hit its ceiling on a
*different* message — correct non-firing.

### 3. ⚠ CORRECTION — "the message fix, not candidate 2, is the repair path" is REFUTED at n=2

I told the owner that naming the orphaned fields was the actual repair path (Fable F4, which I
endorsed and built). **Measured:** `11264de8`'s retry dispatch (00:53:53Z) carried
`last_error` = *"…do not match: schema field "read_article_label" has no template variable"* and the
matching code. The writer **acted** — its schema shrank from 13 fields to 7 — and **left
`read_article_label` orphaned again**, byte-identically. Same for `b8d8a9b8` (`read_more_label`).
All four firings are `*_label` orphans on article/category grids: the writer declares a CTA-label
field and renders the label as static text. Told exactly which field, it does not fix it.

**Tested and dead:** the "two prompt rules in conflict" hypothesis (the regeneration field-name
rule forcing the field to stay) — these are `needs_new_component` items with **no** existing
component (`existing_component_fields` NULL for all four), so that rule never applied.

So piece 2 improved the *information* and did not change the *outcome*. Logged in `WRONG_CALLS.md`.

### 4. THE FINDING THAT CAME OUT OF IT — orphan-only rejections are recorded as `warning` and enforced as `blocking`

`recordValidationRejection` scores an orphan-only mismatch as **severity `warning`** (an unrendered
schema field costs wasted content tokens, not a broken page). The store gate at `:324` treats the
same condition as **blocking**. [MEASURED 2026-08-25, all history]: **9 items FAILED with every
rejection orphan-only (21 rejections)**; **1** orphan-only item completed. Every one of those 9 is a
component the framework refused to store over a field it would merely have ignored.

**This is the next decision, and it is the owner's** (it changes what a shared gate GUARANTEES —
the 2026-07-29 §1 trigger): make orphan-only **non-blocking** (store + record the warning, or
auto-drop the orphaned schema fields at the gate). Candidate 2 now caps the *cost* of this class at
one wasted retry; only that change repairs the *pages*.

### 5. Council round 1 on `cf086b8d`: REVISE (gating: `editquality` HIGH) — every objection dispositioned, round 2 dispatched

| seat / severity | objection | disposition |
|---|---|---|
| `editquality` HIGH, `prior_art` HIGH | does the nested `sub_workflow.steps.mark_failed.config` write survive loop expansion (the `continue_on_error` landmine)? | **Artefact:** it fired 4× through that step. **Mechanism:** the expander stamps exactly ONE key — `injectedStep.Config["continue_on_error"]` at `loop_expansion_handler.go:157` — and deep-clones the rest; my key is not framework vocabulary (0 hits). Guard SQL now shown in the sketch; it fails closed at every branch. |
| `debug_historian` MED | no `snapshot_agent` | **Stale sketch** — the applied file takes it; log: `Snapshot captured … source_id=099b51e0…` |
| `debug_historian` MED | no pod-grep | Done, both `1337` replicas, control absent (§1) |
| `guardian`/`architecture` MED, `bug_historian` LOW | `QualityIssues` consumers — grep ≠ enumeration | **Enumerated:** 6 sites in the store action (log, persist, helper, classifier, context) + `check_component_template_corrupted.go` reading `::text` opaquely. **0** occurrences of `[0]`/`len==1`/`->0` anywhere. |
| `guardian` LOW | closed-key decoders of `result` | `DisallowUnknownFields` at 3 sites fleet-wide, none on a work item's result |
| `editquality` LOW | merge must be additive — test it | **`TestFailureLadder_RepeatMarkerIsAdditive`** added, green in the full suite |
| `bug_historian` LOW | regexes vs format strings coupled by test? | Yes — `TestMismatchMessageNamesEveryOffender` scores real inputs and requires each name; drift → bare fallback → red |
| `prior_art` MED | precedent | `f1f1fc37` APPROVED r1 (this mechanism); `4cdec68b` r2 + `df0748bf` r1 (the ladder, WII-024) |
| `reuse` LOW ×2 | existing formatter / existing item-type array? | 16 `Refusal|Rejection|Remedy` funcs enumerated — all recorders for other seams; `recordRetryFeedback` IS the seam. Array shape matches the estate's existing lists |

Resubmitted with `RESUBMIT_CORR` so the trail accumulates — corr in the next commit's trailer.

### 6. Close condition — MET on every piece

Fixed **and** live **and** measured: mechanism (f1f1fc37, live 1334+), feedback channel (WII-026,
demand-proven), marker + message (live 1336+, probed), opt-in (607 applied, **fired 4×**). **345
can close.** What it hands on is not a residual of *this* defect but the next one it exposed (§4),
which belongs in its own file.

---

## 2026-08-25 (afternoon) — the two "small follow-ups" INVESTIGATED; neither is a one-liner, and 524 is already resolved

The owner asked me to do the two follow-ups from decision 3 and to chase 524. Findings:

### 524's half-committed twin — ALREADY RESOLVED (nothing to do)

[MEASURED 2026-08-25] At HEAD only the un-held pair exists
(`524_claimed_item_timeout_honours_the_cooldown{,_ROLLBACK}.sql`); disk matches; **nothing staged**;
the class sweep (`_HOLD.sql` with an un-held twin at HEAD) is **empty**. The applied file's ledger
checksum `30af2a80…` equals the committed file's md5. Someone completed the deletion half since it
was flagged on 08-24. The LANDMINE entry for the trap stays (it is prospective); the live instance
is gone.

### Follow-up (a) "register the opt-in key via `RegisterActionInputSpec`" — the council seat's own remedy is mis-specified; DID NOT ship it, and here is why

The intent was: make `stop_on_repeat_failure_item_types` visible to the RFC_022 optional-key-budget
counter. Traced end to end:

1. **The counter reads `len(spec.Optional)`** (`cmd/config-key-audit/optionalbudget.go:153`), not
   actual config usage. So visibility requires the key in a spec's `Optional` list.
2. **The key lives on a `fail_work_item` step** (migration 607's `mark_failed`), so the action whose
   budget must count it is `fail_work_item` — whose spec **is passed to `ExtractActionInputs` at
   runtime** (`load_work_item_actions.go:1200`). Adding a key to `Optional` is therefore **not
   declaration-only** — it changes what the extractor walks on the failure hot path.
3. **The key is a directly-read literal** (`stopOnRepeatFailure` reads it straight from
   `StepConfig.Config`), exactly like `error_message`/`status_override` — which `FailWorkItemInputSpec`
   **deliberately excludes**, with a comment saying so. (The convention is not universal —
   `feed_actions.go:57` includes `error_message` — so practice is *inconsistent* fleet-wide.)
4. **`update_work_item_status` uses no `ExtractActionInputs` and has no spec** — but
   `censusUncountedActions` **already prints it as "unknowable"**, which its own comment says is the
   intended treatment ("registering a spec is work these actions may never need"). So a spec there
   adds nothing the census does not already show.
5. **Budget consequence today: nil.** `fail_work_item` really carries 3 optional keys, `update` ~7 —
   both far under N=10.

**So the honest remedy is not "register the key." It is a decision about the RFC_022 mechanism:** the
budget counts `spec.Optional`, but ~**35 actions** read optional config literals directly and specs
declare such keys inconsistently — so a whole class of optional authority is uncountable *by design*,
and 345's key is the third example on this one action (after `error_message`, `status_override`).
Jamming one literal into one runtime extraction spec, against that spec's stated convention, for zero
budget benefit, would be patching a shared seam blind — the exact thing this lane has spent a week
not doing. **Routed to the `config-key-audit`/RFC_022 owner as a mechanism question, not actioned
here.**

### Follow-up (b) "options struct for the 8-param ladder" — legitimate, non-urgent, and its own reviewed change

`applyWorkItemFailureLadder` has 8 positional params across **23 call sites** (21 test + 2 real). The
guardian seat rated the struct refactor **low, "before the next addition"** — i.e. when a 9th param
is actually needed, which nothing pending requires. It is a refactor of **the single failure-write
seam for the whole fleet**, so it wants its own council round, not bundling into a bug close. The
real pain it prevents (a silent sqlmock arity break in sibling files) is now already caught — the
sibling tests broke loudly this week and are fixed. **Recommend deferring** to the next param
addition; doing it now is a speculative fleet-seam change for a benefit already covered.

**Net:** 524 done; neither follow-up is a change I should ship unreviewed, and both are recorded for
their proper owners. This does not affect 345's close condition — every piece of the defect remains
fixed-and-live.
