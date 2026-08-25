# NOTES — `bugs_open/388`, component contract vs storage identity

Append-only, newest at the bottom. Missteps are not an appendix here; they are most of the value.

---

## 2026-08-25 — session 1, picking the lane up

### Ownership check first

`scripts/who-owns.py 388` said **OWNED or recently active**, pointing at
`bugfix_378_usage_count_derived` (14 commits/14d). That looked like a stop sign and was not one: 378
filed 388 and its own handoff says so explicitly —

> *"The 27-section_type contract mismatch is now `bugs_open/388`, filed 2026-08-25. It is NOT a 378
> residual — 378 *reduced* it from 29 to 27 — and it has its own evidence and `[UNMEASURED]` list. Do
> not re-derive it here."*

So the lane that shows as owner is the lane that handed it off. **`who-owns.py` cannot distinguish
"is working this" from "filed this and moved on"** — read the owning lane's handoff before treating
its verdict as a refusal. Confirmed live by messaging the `bugs_open/378` session, which replied
"I am NOT touching bugs_open/388. You own it."

### Is the bug still valid? Yes on the data, no on the mechanism as written

The population query reproduces: **27 of 120** section_types diverge as of 2026-08-25 (filed as 27 of
117 on 08-24 — the denominator moved by 3 in a day, which supports the file's "it will keep growing").

But the mechanism as filed is incomplete, and the file itself flagged the soft part `[INFERRED]`:

> *`[INFERRED]` that `NormaliseToKebab(section_type)` is the store's derivation in all cases; taken
> from `resolveContractViaStorageIdentity`'s doc comment ... not from reading
> `store_generated_component_action.go`'s own derivation end to end.*

Reading it end to end (`parseGeneratedTemplate`, `store_generated_component_action.go:797-807`): the
store uses **the LLM's own emitted `function`** when non-empty, and falls back to
`NormaliseToKebab(section_type)` only when the model supplies none. And since `e1951c24b` (the 337
fix, live 08-22) the live `component-creator` prompt carries an explicit pin:

> *"Also set the top-level "function" in your output JSON to exactly: `{{.existing_component.function}}`
> ... Do NOT choose a different function name — the component library matches regenerations by
> function, and a different name silently creates a parallel duplicate component instead of
> regenerating this one."*

**So the two resolvers ARE bridged. The bridge is prompt text.** That is the actual bug, and it is a
sharper one than "two resolvers exist": an identity decision is delegated to an LLM instruction, with
nothing validating it, nothing recording a divergence, and the instruction rendered only inside
`{{if .existing_component.field_names}}` — i.e. **the guard is conditional on the very thing it
protects**. A resolved row with an empty `input_schema.fields` gets an identity and no pin. 5 of 154
active section rows are schema-less today; 4 of those 5 happen to carry `function == section_type`, so
the hole is presently benign. **That is luck, not design** — the 378 lane's phrasing, and it is right.

### The measurements that were not in the bug file

`llm_call_log`, all-history to 2026-08-25: **672** `component-creator`/`generate_template` calls;
**11** ever rendered the pin; **11 of 11 obeyed it**; 0 disobeyed. Since the pin shipped (08-22), 37
calls, 7 carrying the regeneration block.

⚠ **0 of 11 is not evidence of reliability.** Naming what the sample could have detected: with n=11
and zero failures the 95% upper bound on the disobedience rate is ~24%. This cannot distinguish
"always obeyed" from "obeyed four times in five".

The 27 partition cleanly by what would happen if the pin were broken **[MEASURED 2026-08-25]**:

| outcome | count | why |
|---|---|---|
| **loud, wrong-cause refusal** | **15** | a row with `function == section_type` exists, and `lookupBaseComponent` has **no `is_active` filter** (deliberate, 2026-05-06), so the store finds it, `isRegeneration` is true, and the field-contract guard diffs a schema the writer was never shown |
| **silent duplicate creation** | **12** | no such row, so `isRegeneration` is false, the guard is vacuous, and a parallel row is born with no error and no work item |

The disconfirming result would have been a zero in either column. Neither is zero.

### What the guard actually does (the file's `[UNMEASURED]` #2, answered)

It **refuses loudly** — `store_generated_component_action.go:452-465` diffs old vs new schema field
sets, appends a blocking issue naming every stranded field, calls `recordValidationRejection`, and
returns an error. It does **not** silently overwrite. But it is keyed on the row *the store* resolved,
so when the store resolved the wrong row the refusal is real, loud, and names the wrong cause — and
when the store resolved no row the guard says nothing at all.

### The historical damage, and its signature

Two section_types carry a duplicate pair born from the `generated` route:

- `tool-archetype-taster-quiz`: `archetype-taster-quiz` (2026-04-13) → `tool-archetype-taster-quiz` (2026-05-06)
- `tool-gripper-payload-calculator`: `gripper-payload-calculator` (2026-05-01) → `tool-gripper-payload-calculator` (2026-05-06)

Both second rows carry `function == section_type`, which is the fingerprint of the section_type
fallback derivation. Both predate the pin by months.

⚠ `created_from` is what separates this from noise: the `hero` pool has 7 rows and looks worse, but
all 7 are `manual` and are deliberate vocabulary.

---

## 2026-08-25 — THE WRONG TURN, recorded where I made it

**I claimed eight cancelled work items were 388's first measured damage, on a census taken today, for
events that happened on 08-17.** The `bugs_open/345` lane handed me eight `needs_new_component` items
refused at 3/3 with `removes/renames`, noting that every one of their section types already had an
active component. I queried, saw the divergence, and wrote "this looks like a direct hit" before
checking a single date.

The dates refute it:

| row | `section_type` | fields | born |
|---|---|---|---|
| `mortgages-repayment` (the 8 fields the refusal actually names) | **NULL** | 8 | 2026-08-15 |
| `mortgages-repayment-remortgagecalculator-uk` (the "active component") | `mortgages-repayment` | 28 | **2026-08-21 18:19Z** |

The second row postdates the refusals by four days — it was minted by `bugs_open/311`'s diversion fix
(`17d883333`, 08-19) — and the advisory's fallback resolver shipped 08-22. On 08-17 there was no
divergence to have: one row, invisible to the advisory (NULL `section_type`), no fallback, blind
writer, store finds it by function, refuses. **That is `bugs_open/337`'s defect, since fixed.**

The peer's statement was true and still misled me: a present-tense fact offered as the explanation of
a past event. **The cheap check is one query — `SELECT created_at` on every row your explanation
depends on, and compare it to the event.** Anything that postdates the event is not available to it.

**The honest consequence, which is better than the claim I nearly made:** the `removes/renames` class
runs 10 (08-15), 9 (08-17), 74 (08-18), 4 (08-19), then **zero** on 08-20 and 08-21, and exactly one
on 08-22 (`loans-application-tracker`, where `function == section_type`, so no divergence). **388 has
ZERO measured firings.** It is latent and structural, exactly as filed. A latent bug stated as latent
is worth more than a latent bug dressed as an active one.

Logged fleet-wide in `WRONG_CALLS.md` under the same date.

**Second-order misstep, same session.** I concluded `WRONG_CALLS.md` "does not exist" — `tail`, `find`
and `git ls-files` all agreed. All three were relative-path calls and an earlier `cd` into this lane's
directory had persisted in the shell. **A tool whose working directory persists turns every relative
path into an unchecked claim about state.** Absolute paths, or `cd` to the repo root in the same
command.

---

## 2026-08-25 — what the peer lanes said (all four replied, none blocked)

- **`bugs_open/345`** — no file overlap; its hunks are `describeSchemaTemplateMismatch` (~:1406) and
  its call at :325, nowhere near my regions. ⚠ **Its repeat detector compares the refusal message
  BYTE-FOR-BYTE**, so a wording change to the "removes/renames" text straddling a roll costs one
  non-matching transition per in-flight item — say so in the commit if the text moves. Also: its
  `terminated_on_repeat` marker (WII-029, live) makes this class **countable going forward**, so no
  new counter is needed.
- **`bugs_open/378`** — no overlap; ⚠ **migration `610_content_components_drop_dead_usage_count_HOLD.sql`
  is committed and deliberately unapplied**, and DROPs `content_components.usage_count`. If my fix
  touches the birth INSERT, **do not re-add `usage_count` to its column list** — naming a column in an
  INSERT is enough to break every component creation once it is dropped. 610's guard aborts if the
  column has moved; if my work ever writes it, that alarm is real.
- **`bugs_open/283`** — measured negative, not a recollection: the per-instance conversions write
  `content_components` by direct id-keyed UPDATE (`fix_component_template_action.go:1178,1374`) and
  never call `store_generated_component` (one grep hit across four files, and it is a comment at
  :1082). One consequence to carry: `InstanceToken` derives from `function`, so if a regeneration
  lands on a differently-named row, that component's element ids change value at the next rerender and
  `page_components.rendered_html_digest` takes a new value — harmless, because all 6 writers compute
  the digest in the same statement as the bytes and all readers compare stored-to-stored, never
  against a fresh render. ⚠ Its scar tissue, worth obeying: **if this fix gains a "did the identity
  resolve correctly?" check, that check must not be answered by the same query that did the resolving**
  — `bugs_open/324` shipped 32 dangling rows because the completeness check re-grepped the renamer's
  own patterns.
- **`bugs_open/357`** — adjacency confirmed (binding seam vs birth seam); reply pending at time of
  writing.

---

## 2026-08-25 — the `090` diagnosis run: UNVERIFIABLE, and the way it failed is instructive

`RUN_CORRELATION_ID=2f80ff5e-96db-4d9f-8dfa-f2b8ea9d52d0`, filed 10:28Z, complete 10:39Z (11 minutes
— much faster than the ~30 the runbook budgets, so latency was not the constraint).

**Verdict: `NOT CONFIRMED (stopped: iteration-cap)`, status `UNVERIFIABLE`.** No fix proposed. It is
neither a confirmation nor a refutation and must not be quoted as either.

**What it actually did, because this is the useful part.** Its iteration-1 scope pulled
`load_existing_component_action.go:resolveContractViaStorageIdentity` — the **fallback** — and, reading
`functionName := datahelpers.NormaliseToKebab(sectionType)` as that function's first line, it produced:

> *"The premise that load_existing_component_action.go derives the field-contract advisory from 'a
> separate section_type-ordered query' is contradicted by the code in scope: resolveContractViaStorageIdentity
> does NOT run an independent section_type query."*

**That reading is wrong, and the reason is scope, not logic.** The separate section_type query lives in
`LoadExistingComponentAction` itself (`load_existing_component_action.go:163-182`), and
`resolveContractViaStorageIdentity` is called **only** from the `if err != nil` branch of that very
query — i.e. the function it read is by construction the path taken when the query it says does not
exist has already missed. The loop scoped the callee and never pulled the caller, then spent its
remaining iterations arguing with its own bundle about whether that callee had been fully included.

**So the run cost a round and settled nothing.** Recording it rather than quietly dropping it, because
a silently discarded UNVERIFIABLE looks identical to a run never filed.

Two things it *did* contribute, and they are worth the price:

1. Its `NeededEvidence` named exactly the right next step — *"whether [the store] also calls
   resolveStorageIdentity or independently resolves the row from parseGeneratedTemplate's
   functionName"* — which is the question, stated better than my symptom stated it.
2. It is a live instance of the failure mode this estate keeps logging in itself: **reading the
   function you were pointed at instead of the one that calls it.** `bugs_open/337`'s own refutation
   was the mirror image (a thread that skipped `rerenderLoadSections`). The loop is not immune to it.

**Practice note for the next symptom I write:** name the CALLER as well as the callee when the claim is
about which of two paths is taken. "the primary query in `LoadExistingComponentAction`, whose `err != nil`
branch calls `resolveContractViaStorageIdentity`" would have scoped both and cost nothing.

---

## 2026-08-25 — PRIOR ART: this bug already has a register entry, and it says the fix was flagged and not built

Found by grepping the concept register rather than the bug directories, which is why the filing lane
missed it. `docs026_concept_register/register/component-lifecycle.md`:

> **CLC-006 — F4 — regen-vs-create keyed on the LLM-chosen function (silent fork)**
> **status:** partial
> *"F4 (structural finding): regen-vs-create is keyed on the LLM-chosen function → nondeterministic.
> Miss case = silent FORK"; pin migration applied; "store-side advisory FLAGGED as follow-up, not built."*
> *"...a store-side advisory (warn when function misses but an active same-section_type row exists) is
> deliberately advisory-only and unbuilt, since multiple components per section_type can be legitimate."*
> **verify-later:** *duplicate non-forked function rows in content_components; whether any store-side
> advisory exists*

**388 is CLC-006's unbuilt store-side half, rediscovered from the data.** Three consequences:

1. **Its `verify-later` is now answered, and both answers are the bad one.** Duplicate non-forked rows
   from the `generated` route: **2** (`tool-archetype-taster-quiz`, `tool-gripper-payload-calculator`).
   A store-side advisory: **none exists** — grepped `store_generated_component_action.go` for any
   comparison against `collected_data.existing_component`; there is none.
2. **CLC-004 already documents the function pin as CLC-006's mitigation**, and its own 2026-08-22
   correction records that the pin "goes silent" whenever the advisory does. So the estate has written
   down, twice, that the bridge is conditional — and 388 is the measurement of how wide the gap is.
3. ⚠ **CLC-006's stated reason for not building it is a live constraint on any fix I propose:**
   *"multiple components per section_type can be legitimate."* So the fix must NOT make one-row-per-
   section_type an invariant, must not refuse a second row, and must not "reconcile" names. Honouring
   an identity resolved in Go before generation satisfies this — it changes WHICH row a regeneration
   lands on, never HOW MANY rows a section_type may have.

CLC-020's landmine says the same thing from the other side: *"the DIVERTED row's `function` is
site-suffixed while its `section_type` is not: resolve it by section_type (the selector) or exact
function, never by assuming function == section name."* The 27 are largely that landmine's population.
