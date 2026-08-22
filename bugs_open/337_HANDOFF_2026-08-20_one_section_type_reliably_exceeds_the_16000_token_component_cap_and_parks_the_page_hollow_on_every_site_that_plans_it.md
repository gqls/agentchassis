# 337 — one section type reliably blows `generate_template`'s 16,000-token ceiling, so every site that plans it loses the page after burning three full generations

**Filed:** 2026-08-20 by the `bugfix_311_component_keys` lane, as the residual `311` named and
deliberately did not fix ("a cap decision, not a defect in the store"). **Status: OPEN, unowned.**
**Severity: low volume, 100% reproducible within it** — 3 items, 2 sites, 3 pages, and every
occurrence is the same section type.

**Not a duplicate of `bugs_open/205` or `bugs_open/257`, and the difference is the whole point.**
205 is a step whose budget nobody configured (runs at the hardcoded 2048); 257 is the class where a
configured budget never *arrives* at the provider. Here the budget is configured, it arrives, it is
respected — and it is simply **below what this task needs**. Different fix, different owner.

## The mechanism, in three lines

1. `component-creator`'s `generate_template` step is configured
   `{"model":"claude-sonnet-4-6","provider":"anthropic","max_tokens":16000}`
   (read from the live `agent_definitions` row 2026-08-20, not from a seed).
2. The LLM's template for one particular section type runs to **~47,000 characters**, hits the cap,
   and the transport correctly refuses a truncated body rather than storing a fragment
   (the `bugs_open/012` lesson, working as intended).
3. `needs_new_component` retries to `max_attempts=3`, each attempt paying a full 16,000-token
   generation, then parks `failed`. The page is left with no component and ships hollow.

## Evidence — every occurrence, and they are all ONE section type

`site_work_items.error` verbatim [MEASURED 2026-08-20 09:20Z]:

| site | page | error |
|---|---|---|
| loancalculator.co.uk | `tool-credit-roadmap` | `output_tokens=16000 reached the configured cap, 46637 chars recovered` |
| loanzy.uk | `tool-eligibility-checker` | `output_tokens=16000 reached the configured cap, 48553 chars recovered` |
| loanzy.uk | `tool-credit-health-check` | `output_tokens=16000 reached the configured cap, 47436 chars recovered` |

All three requested **`section_type = 'loans-credit-health-check'`**. All three reached
`attempt_count = 3`. Two different sites, three different page contexts, three consecutive
generations each — **nine cap-hits, zero successes, one section type.**

> **NARROWED the same day, 09:35Z, before anyone acts on it:** of the three pages, **two are the
> live loss** — `loancalculator.co.uk`/`tool-credit-roadmap` and `loanzy.uk`/`tool-credit-health-check`,
> both `status='active'`, both `build_status='needs_rebuild'`. The third,
> `loanzy.uk`/`tool-eligibility-checker`, is **`status='archived'`**, so its cap failure cost nine
> generations and no live page. Checked because a sibling repair in the same session spent a full
> build on a different archived page before the archived-page guard refused the stamp at the end.
> **The spend figure is unaffected; the "page ships hollow" claim applies to two pages, not three.** So this is not "sometimes
the LLM is verbose"; the ceiling is roughly a third of what this brief actually produces, every
time.

Adjacent cap failures on other agents, for scope only (do NOT fold them in — different steps,
different ceilings): `oufe.com` `improve_tool` (2026-08-11), and three `needs_diagnosis` items
(2026-08-14, 2026-08-19 ×2). Fleet total carrying a `stop_reason=max_tokens` error: **7 items as of 2026-08-20.**

**Blast radius is small today and grows with the vertical:** **4** pages fleet-wide **as of 2026-08-20** plan a
`credit-health`-shaped section. Both affected sites are loan/credit sites, which is the vertical the
portfolio buildout multiplies — so the next loans domain will plan it too and lose the same page.

## Fix candidates, ordered by what makes the bad state unrepresentable

1. **Make the writer's brief bound its own output, and prove the bound.** The ceiling is only
   wrong relative to a prompt that does not know it exists. A component brief that states a length
   budget (and a rich-tool decomposition rule: markup + a script the extractor separates, not one
   47k blob) closes the door for every future section type, not just this one. Requires reading
   `generate_template`'s prompt and measuring a real generation against it — the honest version of
   this candidate includes "and here is the length it actually came in at".
2. **Raise the step's cap to fit the measured need** (~24,000 would clear all three by ~50%).
   Cheapest, an owner threshold call, and it does not close the door — the next richer tool blows
   whatever number is chosen. Note the interaction: this multiplies the cost of a *failing* item's
   three attempts too.
3. **Handle truncation instead of refusing it** — a continuation call, or a retry at a higher
   ceiling on `stop_reason=max_tokens` specifically. Structural, cross-cutting (it would change the
   contract for every LLM step in the fleet), and therefore an `architecture_review` question
   rather than a bug patch. It is also the only candidate that helps the `improve_tool` and
   `needs_diagnosis` occurrences.
4. **Spend fewer attempts on a deterministic failure.** Three identical cap-hits is 48,000
   generated tokens thrown away per site. A cap refusal is not a transient — retrying it unchanged
   cannot succeed. Weakest as a fix (it saves money, repairs nothing) but it is the cheapest and it
   is honest about what a retry can and cannot do.

## How to verify a fix

Re-drive `loans-credit-health-check` on loanzy.uk (RUNBOOK recipe in
`docs/agent_docs/docs024_key_docs_latest/bugfix_311_component_keys/RUNBOOK_311_fix.md`) and assert:
the item completes with `attempt_count = 0`; a `content_components` row exists for the section with
`length(html_template) > 0` and containing `</section>`; and — the half that matters — the served
page gains real controls, `grep -c '<input'` on
`https://loanzy.uk/tools/credit-health-check/index.html` going above 0 from a pinned "before" of 0.
**Verify at the artefact, not at the item status:** a stored component is not a working calculator.

## Declared substitute for the `090` loop (owner ruling 2026-07-31)

No `090` run, because **no causal claim in this file is inferred.** The cap is read from the live
`agent_definitions` row; the failures are quoted verbatim from `site_work_items.error`, which names
the mechanism itself (`output_tokens=16000 reached the configured cap`); the census is a single
`ILIKE` over the same column and could have returned any number. What this file does NOT assert is
*why* this brief produces 47k chars, or that candidate 1 would fix it — those are open, and
candidate 3 in particular is architecture-scope and should not be taken as diagnosed.

## Related

- `bugs_open/311` — the parent. Its store-side collision is fixed and demand-proven; these two
  loanzy pages are the residual that fix cannot reach, because they die upstream of the store.
- `bugs_open/205` (a step with no configured cap, running at 2048; closed in substance) and
  `bugs_open/257` (the class: a configured budget that never arrives). Siblings, not duplicates.
- `bugs_open/012` — the truncation-and-config family. The refusal here is 012's lesson working:
  a cut completion is not stored as if it were whole.

---

## 2026-08-22 — taken; still valid; fix built (escalation seam + resize), council pending

Taken by a fresh session (lane docs:
`docs/agent_docs/docs024_key_docs_latest/bugfix_337_token_cap/` — PLAN, RUNBOOK, NOTES,
README). `who-owns` pointed at the 311 lane, but their handoff lists this as "filed from
this lane, unowned" and their live work is 311/345/351 — no competing fix in flight, and
no open work item targets this bug.

**Validity re-check [MEASURED 2026-08-22]:** cap still 16000 in the live row; the items
still parked `failed`/`attempt_count=3`; and a FOURTH occurrence this file predates —
loanzy.uk `needs_new_component:loans-credit-health-check_run1`, 08-18, same three-cut
signature. `llm_call_log` holds 9 generate_template truncations 08-15→08-19, recovered
46,441–48,817 chars each.

**Three facts this file did not have:**

1. **The cap was tight for the step's ORDINARY work, not just this section.**
   Successful calls, 14-day window: p95 13,633 of 16,000 (85%), max 15,374 (96%).
2. **The fleet headroom monitor flagged this step BEFORE and DURING this bug's life and
   nothing consumed the flag.** LCO-007 (`fleet-step-token-pressure`) doc_notes rows
   from 2026-08-18 onward: "T generate_template@16000 — n=229, p95 92.4%, peak 100.0%,
   truncated 9". Detection works; flag→action dispatch does not — that gap is a named
   residual here, not re-solved with a second monitor. (FIX-058's open question — does
   the near-miss threshold scale with the cap? "revisit when a 16000 seat first
   crosses" — has now had its trigger: this bug is that crossing.)
3. **The sibling whole-component writers were levelled and this step was missed:**
   `generate_tool_html`/`improve_tool` at 32000, `recreate_tool` at 64000
   (bugs_closed/012's table, migration 168 lineage) — generate_template sat at 16000.
   Per bugs_closed/067's sweep rule: generate_template is component-creator's only
   `execute_llm_prompt` step, so the levelling below is the whole sweep for this agent.

**Fix shipped (candidates 2+3 together, 3 in its narrow honest form):**

- **Framework seam (candidate 3, scoped):** `max_tokens_ceiling` — an opt-in
  `ai_service` key; on a typed `TruncatedError`, `execute_llm_prompt` retries ONCE at
  the ceiling, logging the cut call `success=false` with an `ESCALATED
  (bugs_open/337: …)` prefix; the escalated call's outcome flows through the existing
  machinery (tolerate/076 guard, transient ladder) verbatim. NOT a continuation call
  and NOT a fleet default — opt-in per owner ruling 2026-08-02 §2. Register
  **MDL-042**; code `platform/orchestration/actions/truncation_escalation.go` (+test,
  + wiring in `ai_actions.go`). Inert until the next chassis roll.
- **Resize from measurement (candidate 2):** migration
  `docs/agent_docs/sql_for_agents/549_generate_template_cap_resized_and_escalation_ceiling_armed.sql`
  — 16000→24000 routine (clears p95 by ~75% and this section's extrapolated ~19–22k
  need) + ceiling 32000 (clock-safe: step measures 92–127 tok/s, 600s non-streaming
  timeout; 40k+ risks trading truncation for a silent clock death). 415-pattern
  resolved-value asserts; 484-pattern gates; rollback sidecar. Live on apply.
- **Candidate 1 (bound the brief)** stays the named follow-up — FIX-059/migration-484
  precedent recorded in the lane PLAN; it changes what the step produces and needs a
  measured real generation to design. **Candidate 4** stays with the RSH-007/WII
  failure-ladder contract: a truncation AT the ceiling still burns the item's
  remaining attempts.

**Verification (per §How to verify, plus the RUNBOOK_311 corrections):** pre-flight
`pages.status` — `tool-eligibility-checker` is ARCHIVED, spend nothing on it; targets
are loanzy.uk `tool-credit-health-check` and loancalculator.co.uk `tool-credit-roadmap`.
Record WHICH half won: completion with zero `ESCALATED%` llm_call_log rows proves the
resize; with one, proves the escalation. Either passes; the attribution must be stated,
not assumed.

---

## ⚠ CORRECTED 2026-08-22 — THIS FILE'S CENTRAL MECHANISM CLAIM IS FALSE. The cap is not the binding constraint, and the truncation is a ~11% SIDE EFFECT of a regeneration loop whose real driver is pre-store validation.

Found while verifying the fix by demand. **The correction matters more than the fix**, so it
goes above the fix record. Nothing below invalidates the *change* that shipped (see
"disposition"), only the *diagnosis* this file asserted and that I repeated to a council.

### What the file said, and what is actually true

This file: *"nine cap-hits, zero successes, one section type"*, *"the ceiling is roughly a
third of what this brief actually produces, every time"*, *"100% reproducible within it"*,
and *"`needs_new_component` retries to `max_attempts=3`, each attempt paying a full
16,000-token generation"*.

**The exact census** — `llm_call_log` joined to `site_work_items` **on `work_item_id`**, and
filtered on the item's own `spec->>'section_type'` (the join is the load-bearing part; see
the misstep below) [MEASURED 2026-08-22]:

| | calls | outcome |
|---|---|---|
| `loans-credit-health-check` generations, all history | **82** | **73 SUCCEEDED at cap 16000** (outputs 8,641–15,374), **9 CUT** |

So the cap fits this brief **~89% of the time**, on the same two sites, in the same windows,
interleaved with the cuts hour by hour. "Every time" is false; "zero successes" is false.

**And the retry arithmetic was wrong in the opposite direction from the file's claim.** Per
item, `attempt_count` vs actual generations:

| site | item | attempt_count | LLM calls | of which CUT |
|---|---|---|---|---|
| loancalculator.co.uk | `needs_new_component:…` | 3 | **13** | 3 |
| loanzy.uk | `…_run1` | 3 | **55** | 3 |
| loanzy.uk | `needs_new_component:…` | 3 | **11** | 3 |

Each item generated 11–55 times, not three. The waste is **far larger** than this file
claimed — and it is not paid on truncations. It is an **in-workflow regeneration loop**, the
same shape `bugs_open/345`/migration 533 measured ("one item (8c8f5de5) produced 52
rejections in 3h34m while attempt_count capped at 3" — note 8c8f5de5 is the 55-call item
above, i.e. the two lanes measured the SAME item from different ends and neither saw it).

### What actually blocks these pages — proven today, at a taller cap

The 2026-08-22 re-drive generated cleanly at 24,000 (14,244 / 12,709 / 14,816 tokens, zero
truncations) and **loanzy STILL produced no component**, because `store_component` refused it:

> `generated template … rejected by pre-store validation: field "cta_primary_url" declares
> source "site_specs.ctas.primary_url" but no site carries a site_specs aspect named "ctas"
> … (bugs_open/309)`

That is the recurring blocker. The writer invents an unresolvable `site_specs` source in the
component's `input_schema`, the pre-store validator correctly refuses it, the workflow
regenerates, and round and round — occasionally drawing a long generation that hits the cap.

### Why the truncation looked like the cause (the transferable trap)

**A work item's `error` column holds the LAST failure, not the recurring one.** These items
looped on validation rejections and happened to die on a truncation, so `error` said
truncation — and the census that "confirmed" it (`WHERE error ILIKE '%reached the configured
cap%'`) could only ever return items whose final error was a truncation. It selected on the
symptom it was testing for. The 9-out-of-82 successes were invisible to it because
`site_work_items` records one error per item, while `llm_call_log` records one row per call.

**The lesson, generalised:** when an item's recorded error names a mechanism, count that
mechanism's occurrences AT THE CALL LEVEL before believing it is the mechanism. A ratio
(9 cut / 82 calls) answers "is this the binding constraint?"; a count of items (3 items, all
with truncation errors) cannot, however many sites it spans.

### Disposition of the change that shipped — re-scoped, not reverted

Migration 549 (16000→24000 + `max_tokens_ceiling` 32000) and the `MDL-042` escalation seam
stay, on evidence independent of this bug's false premise: the step's SUCCESSFUL calls run
p95 13,633 / max 15,374 against the old 16,000 (85%/96%), `fleet-step-token-pressure` has
flagged it since 08-18, and 9 genuine truncations across 82 calls is a real ~11% loss worth
removing. **But they must NOT be credited with healing these pages, and this file must not be
closed on them.** The three 08-22 generations all came in BELOW the old 16,000 cap, so they
do not even demonstrate the raise was necessary for those runs.

### What this bug now is

**Re-scoped to: a `needs_new_component` regeneration loop driven by pre-store validation
rejections (`bugs_open/309`'s unresolvable-source class), which burns 11–55 generations per
item behind an `attempt_count` of 3 and parks the page.** The cap work is done and is a
separate, smaller win. Whoever takes this next should start at the validator/loop, not the
ceiling — and should talk to the `bugs_open/345` lane, which has measured the same items.

**Still open, unchanged:** both pages remain without their component
(loancalculator's stored today and is awaiting a page re-render to attach it; loanzy's was
refused by the validator). `tool-eligibility-checker` remains archived — spend nothing on it.

### Repair state after the 2026-08-22 re-drive, and a correction to this file's own verification recipe

- **loancalculator.co.uk / `tool-credit-roadmap` — REPAIRED.** Component stored, page
  re-rendered, `page_components` 4 → 5 slots, deployed. The served page
  **`https://loancalculator.co.uk/tools/credit-roadmap.html`** (200, 46,594 B) carries
  `<section class="tool-credit-health-check-section">` with a working quiz — 13 `<button>`
  and 4,593 bytes of inline logic (`next()`, `showResult()`, listeners, step/points/meter).
- **loanzy.uk / `tool-credit-health-check` — STILL BLOCKED**, by the re-scoped cause above:
  the fresh 12,709-token generation was refused by `store_component`'s unresolvable-source
  rule. Page unchanged (24,323 B, 3 sections, 0 inputs).

⚠ **Two corrections to §How to verify a fix, both of which cost this lane a wrong reading:**

1. **The URL in that section is right for loanzy and WRONG for loancalculator.** URL shape is
   **per site**: loanzy serves `/tools/<name>/index.html`, loancalculator serves
   `/tools/<name>.html`. The name-derived guess returns that site's custom 404 — **1,201 bytes
   of real HTML with a stable md5 and zero `<input>`**, so it survives a two-reads stability
   check and a content grep while being the wrong document. Pin with the status code
   (`curl -s -o /dev/null -w '%{http_code} %{size_download}\n'`) and refuse anything but 200.
2. **`grep -c '<input'` is the wrong success predicate for this section type.** The component
   this bug is about is a **button-driven quiz**: it scores **0 `<input>` while fully working**.
   Assert the section's presence plus its behaviour (inline script bytes / declared handler
   names), not one tag borrowed from a calculator-shaped tool.

---

## 2026-08-22 (evening) — RE-SCOPED A SECOND TIME, and this time the class was counted before it was named. Fix built, committed, prompt half applied. STILL OPEN.

A fresh session took this after the re-scope above handed it on with *"start at the
validator/loop, not the ceiling"*. That instruction was right. The **class it named was
not**, and neither was my own first census — both are corrected here rather than quietly
replaced.

### Correction 1 — the re-scope above named the MINORITY class

The section above re-scoped this to *"`bugs_open/309`'s unresolvable-source class"*.
Counted at the call level rather than inferred [MEASURED 2026-08-22] — 101
`component_validation_rejected` rows, 2026-08-15→08-22, 4 sites:

| class | rows | what it is |
|---|---|---|
| **field contract** | **97** | "regeneration removes/renames N existing schema field(s)" |
| source vocabulary | 3 | "no site carries a `site_specs` aspect named …" (309's class) |
| other | 1 | |

The named class was 3 of 101. **The same lesson as the correction above, one level down:
the previous session stopped counting once it had a mechanism that fitted the case in
front of it.** Its own §"Why the truncation looked like the cause" says to count a
mechanism's occurrences before believing it is the mechanism — and the re-scope it
introduced did not do that for its own replacement claim.

### Correction 2 — my own census was wrong, and I stated it before I checked it

I reported "52 deadlocked components". That number was keyed on **function names** rather
than on **demand**, and a disconfirming check killed it: 21 of the 52 had regenerated
successfully since 2026-08-01. Re-measured on what is actually requested: of **14** section
types ever requested by a `needs_new_component` item, **0** are in the blind-and-stranding
state today. The reason is that **`bugs_open/311`'s site-scoped diversion (first diverted
row 2026-08-19 16:22:57Z) creates a `section_type`-carrying row as a side effect — and 97
of the 98 field-contract rejections predate it.** Logged in `WRONG_CALLS.md`.

### What the defect actually is

**The birth gate enforces two contracts the writer is never shown, from sources the writer
never sees**, so compliance is left to chance, the gate refuses, the item re-dispatches
unchanged, and the page parks.

- **Field names.** `load_existing_component_action.go` keys its advisory on `section_type`
  and also filters `is_active` and `component_level='section'`; the gate resolves the row
  it will overwrite by **function**, via `resolveStorageIdentity`, which deliberately has
  **neither** filter (`component_storage_identity.go:157-165` — the `is_active` filter was
  removed 2026-05-06 *because the paradigm regeneration target is a deactivated row*). A
  miss leaves the whole preservation block dormant behind
  `{{if .existing_component.field_names}}`, taking the function pin with it. Proven end to
  end: orchestration `4f321f85` had `field_names = ''` and was refused at 12:18:43Z for
  stranding 4 fields, then **succeeded at 12:53:07Z when the blind writer reproduced those
  4 generic names by chance**. The 18-name sibling (`button_1…button_18`) never succeeded
  in 70 attempts. So this is not "sometimes the LLM forgets" — **preservation is a
  lottery whose odds fall as the field names get more numerous and less guessable.**
- **Source vocabulary.** TIER D enumerates query names exactly (*"use these EXACTLY — do
  not invent new ones"*). TIER C says only `source: "site_specs.{path}"` with three prose
  examples and **no aspect list**, and the live `prompt_template` renders no part of
  `site_specs` **anywhere** despite it sitting in `input_fields`. The live blocker is a
  one-character invention: `site_specs.ctas.primary_url` / `.secondary_url` when the aspect
  is **`cta`**, which carries **exactly** `primary_url` and `secondary_url`.

### What shipped — commit `e1951c24b`, migration **565** APPLIED, `Council-Submitted: 9efde776-a210-42bc-aa99-899d0d301c67`

Reuse, not addition: every contract is computed by **the gate's own functions**, so offer
and enforcement cannot drift. This is a fourth application of a pattern this estate has
already approved, shipped and induced — `016b` §9/092 (*"the writer's allow-list and the
gate's accept-set cannot diverge"*).

1. **Arm A** — the advisory falls back to the store's own `resolveStorageIdentity` on a
   `section_type` miss, reporting only when `IsRegeneration`. **Not** a bare
   `lookupBaseComponent`, which would *manufacture* refusals on the two diversion paths.
2. **Arm B** — `known_aspects` (via the guard's `LoadKnownSpecAspects` and a newly extracted
   `KnownAspectsSorted`, now the single rendering used by both the refusal message and the
   advisory), `known_query_bases`, and `aspect_paths`: leaf `aspect.key` paths with **site
   coverage**, because the gate validates only the first segment so bare aspect names would
   trade a loud refusal for a silent blank. Migration 565 puts them in TIER C, dormancy-guarded.
3. **Arm C** — `section_type` healed on the **rejection** path, `is_active`-gated. The
   existing `COALESCE` runs only on a successful store, so the repair was locked behind the
   success its own absence prevents. Gated because healing makes a component **selectable**
   and migration 036 deactivates broken ones so pages stop choosing them.

**Inert until the next chassis roll.** Migration 565 is live but dormant (it renders nothing
until the Go half emits the key). No ordering constraint, and none is claimed.

### What this does NOT fix — read before crediting it with anything

- **Arm A repairs nothing today.** 0 of 14 requested section types are blind; 311's
  diversion closed that by side effect. It ships as prevention plus the `is_active` hole
  311 never touched.
- **Arm B's class is rare.** 72 section components created since 2026-05-07 and **none**
  carries a phantom aspect at rest; 3 rejections in 4 days. **What changed is the
  consequence, not the rate** — the birth gate went live 2026-08-18 and the first
  phantom-aspect rejection is 2026-08-21. Before the gate this output was *stored* and
  rendered silently blank (309's damage); since the gate it *parks the page*. The gate
  closed the silent door and opened a page-parking one.
- **`bugs_open/345` landed mid-build and narrows Arm B's claim.** Migration 555 was applied
  **2026-08-22 11:08:01Z** and `last_error` is demand-proven (5 orchestrations carry it, all
  since 11:08Z, against 0 all-history; their item `ceea0c07` was refused at 12:18:43Z and
  completed at attempt 1 at 12:53:07Z). The refusal message already lists every aspect
  **name**, so that half is now reachable reactively, one wasted generation later. Arm B's
  unique residual is the **leaf paths and coverage**, the **first** generation saved, and
  **items that have already spent their attempts** — like the 11 parked ones.
- **An avenue tested and closed, so nobody re-proposes it:** replacing the field-contract
  proxy with dependents' real `content_data` keys (which the guard's own comment at
  `:448-450` pre-authorises). **Refuted:** all 10 refused components have 1–2 live
  `page_components` dependents, and `loans-credit-health-check`'s dependent stores 19 keys
  of which **12 are `button_*`**. The rule was protecting real stored content.
- **Deep-path source validation is deliberately out** — guarantee-altering on a shared
  mechanism (RFC track), and the 309/362 lane's call. The prompt block instead tells the
  writer that an unlisted key renders blank, and that the answer is `static`-with-a-fallback
  or `llm` rather than an invented path — load-bearing, because the sibling
  `remortgagecalculator` case wanted `currency_symbol`, which **no site carries in any
  resolvable form**.

### Still open, and what the next session owes

**11 `needs_new_component` items are parked `failed` at attempt 3** across loanzy.uk (×8),
remortgagecalculator.uk, loancalculator.co.uk and gaswholesalers.com. The owner has ruled
**re-drive all 11** — but only **after the chassis roll**, because the Go half is inert
until then and a re-drive now measures nothing. `loanzy.uk/tool-credit-health-check` is
unchanged: **200, 24,323 B, 3 sections, md5 `bdc997300740`** [pinned 2026-08-22].

⚠ **A stale line in this file, corrected:** §Evidence records
`loanzy.uk`/`tool-eligibility-checker` as `status='archived'`. It reads **`active`/`planned`**
as of 2026-08-22 09:15Z. Re-check `pages.status` per item at re-drive time rather than
trusting either reading.

⚠ **Two verification corrections from the section above still stand and are easy to lose:**
loanzy serves `/tools/<name>/index.html` (loancalculator serves `/tools/<name>.html`, and a
wrong guess returns a 1,201-byte custom 404 with a stable md5 that survives a two-reads
check), and **`grep -c '<input'` is the wrong success predicate** — this component is a
button-driven quiz that scores 0 while working perfectly.

### ⚠ Council round 2 could not complete — FLEET-WIDE API LIMIT, not a verdict [2026-08-22 ~18:30Z]

Round 1 was **REVISE**, gated by `prior_art_librarian`, and the gating objection was right:
I described `KnownAspectsSorted` as *"the guard's own function"* in one edit while another
edit **created** it by extraction. Checked against the commit's parent rather than argued —
`resolveStorageIdentity`, `LoadKnownSpecAspects` and `KnownQueryBases` **did** pre-exist;
`KnownAspectsSorted` did not. Honest claim: **three reused, one extracted to be shared.**
Round 2 answers every seat with a reading rather than an argument (see
`COUNCIL_RESUBMISSION_2026-08-22_r2.json`) — including the guardian's most valuable
objection, that if Arm B's keys were a new `output_field` the migration's guard would never
fire and the fix would look shipped while delivering nothing (verified: `output_field` **is**
`existing_component`, which is already in `input_fields`, and the vocabulary is merged into
that same map on all five return paths).

**Round 2 then died at `review_editquality` with:**

> `API request failed with status 400: "You have reached your specified API usage limits.
> You will regain access on 2026-09-01 at 00:00 UTC."`

**This is an account-level cap, not a defect in the submission, and it is not confined to this
lane** — `llm_call_log` shows **10 usage-limit failures in the 18:00Z hour alone** across the
fleet, the first ones today. Until the owner raises the cap, **every LLM step in the estate
fails**, including the council gate, the diagnosis loop and every content build. The verdict
on record for this correlation remains round 1's REVISE; `Council-Submitted:` is therefore the
correct trailer on both commits and `098` will credit them automatically if a later round
approves.
