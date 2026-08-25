# 388 — `component-creator` is advised ONE row's field contract while the store enforces a DIFFERENT row, on 27 of 117 section types

**Filed:** 2026-08-25 by the `bugfix_378_usage_count_derived` lane, found while answering a council
objection (`bug_historian`, medium) about whether repointing an `ORDER BY` could re-shape an enforced
schema. It could not — but measuring that question exposed this, which is a separate and larger
mismatch that predates `378` entirely.

**Severity:** latent, structural. Nothing errors. The cost is that the advice given to the component
writer and the contract actually enforced on it can name different rows, so a writer that obeys its
instructions can still be refused — or worse, silently overwrite a row it was never told about.

**Status: OPEN, IN PROGRESS** — picked up 2026-08-25 by the `bugfix_388_component_contract_identity`
lane (docs: `docs/agent_docs/docs024_key_docs_latest/bugfix_388_component_contract_identity/`).
Diagnosis below is first-hand; no code written **at filing time** — see the corrections, which change
the mechanism, the severity and the ranking of the fixes.

**FIX BUILT, COUNCIL-APPROVED AND COMMITTED 2026-08-25.** Council `5252bee6-0e49-4e41-81fc-6acb014a4802`
— round 1 REVISE (gating premise refuted: `decideStorageIdentity` contains no INSERT/UPDATE/DELETE/Exec
of any kind — the "scoped mint" mints a *name*), round 2 **APPROVED**. Commits: `30d223291` (the Go),
`df4802df4` (migration 612 + CLC-032 + landmine), `66243de7b` (the fourth finding code, from round 2's
advisory objection), `f8b529df6` (the optional-key cron literal). Registry declarations landed in
`eb7d92371` as a same-file passenger of the `bugs_open/358` lane.

**⚠ THE BUG IS NOT CLOSED, AND THE BAR IS "FIXED AND LIVE".** The Go is inert until the next chassis
roll, and the pin needs BOTH that roll AND migration **612**, which is committed and deliberately
**NOT APPLIED** (owner's call; no ordering constraint either way — the key is inert against a binary
whose spec does not declare it, and the new binary without the wire runs un-pinned). The defect is
reproducible until both land. Lane:
`docs/agent_docs/docs024_key_docs_latest/bugfix_388_component_contract_identity/`.

> **⚠ READ THE CORRECTIONS BEFORE THE BODY.** Three claims below are wrong or incomplete, and the
> file's own `[INFERRED]` marker pointed at the one that mattered. Nothing here has been edited away.

## The finding in one paragraph

`load_existing_component_action.go` tells `component-creator` which field names to preserve. It picks
that row by `WHERE section_type = $1 … ORDER BY <most used>, updated_at DESC`. But the **store** —
`store_generated_component_action.go`, via `resolveStorageIdentity` — decides which row a
regeneration *overwrites and enforces* by **function name**, derived as
`NormaliseToKebab(section_type)`. Those are two different resolvers over the same table, and when the
winning row's `function` is not equal to its `section_type`, they name different rows.

> **CORRECTED 2026-08-25 by the `388` lane — the store does NOT derive the function from
> `section_type` in all cases, and since 08-22 the two resolvers ARE bridged.** This file marked the
> claim `[INFERRED]` and the marker was well placed. `parseGeneratedTemplate`
> (`store_generated_component_action.go:798-806`) takes **the LLM's own emitted `function`** when it is
> non-empty, and falls back to `NormaliseToKebab(section_type)` only when the model supplies none. And
> the live `component-creator` prompt has carried an explicit pin since `e1951c24b` (the `337` fix, live
> 2026-08-22): *"Also set the top-level `function` in your output JSON to exactly:
> `{{.existing_component.function}}` ... a different name silently creates a parallel duplicate."*
>
> **So the defect is sharper than "two resolvers exist": the bridge between them is PROMPT TEXT.**
> Nothing validates that the writer complied, nothing records a divergence, and the pin renders only
> inside `{{if .existing_component.field_names}}` — so a row that resolves with an empty
> `input_schema.fields` keeps its identity and loses its pin. **A guard conditional on the very thing it
> protects.** 5 of 154 active section rows are schema-less today; 4 of those 5 happen to carry
> `function == section_type`, so it is presently benign. That is luck, not design.
>
> **And a name cannot be an identity here at all.** `lookupBaseComponent`
> (`component_storage_identity.go:167-176`) filters *only* `function = $1 AND forked_from IS NULL`,
> ordering `is_active DESC, updated_at DESC LIMIT 1` — **no `component_level` filter and no `is_active`
> filter**. `[MEASURED 2026-08-25]` of 330 non-forked rows, **25 `function` values carry more than one
> row**; `site-footer` and `site-header` carry **five each and span `section` AND `site` levels**. So
> the store picks among several rows by recency, and which one it picks can change with no code change.

## Evidence

### The design intent is explicit, and it is only implemented on the FALLBACK path

`load_existing_component_action.go`'s `resolveContractViaStorageIdentity` says in its own doc comment:

> *"The function name is derived exactly as `store_generated_component_action.go` derives it … so the
> prediction and the enforcement agree **by construction rather than by coincidence**."*

That is exactly right — and it is the **fallback**, reached only when the primary `section_type`
query finds **no** row. When the primary query *does* find a row, the careful agreement is bypassed
and the answer comes from a resolver the store does not share.

### The measurement

`[MEASURED 2026-08-25]`, active, non-forked, `component_level='section'`, over all **117**
section_types that have a candidate:

| ordering used by the advisory | rows whose `function` = the `section_type` the store enforces |
|---|---|
| the ordering in place until 2026-08-24 (`usage_count DESC, updated_at DESC`) | **88** of 117 |
| the ordering live today (derived site count `DESC, updated_at DESC`, `bugs_closed/378`) | **90** of 117 |

So **27 section_types still disagree**, and `378` improved this by 2 rather than causing it.

Eight of the 27, to show the shape — it is a *naming* divergence, not a data fault:

| section_type | advisory names this row | its `function` | store enforces function |
|---|---|---|---|
| `case-studies` | `case-studies-list` | `case-studies-list` | `case-studies` |
| `content-block-case-studies` | `case-studies-grid` | `case-studies-grid` | `content-block-case-studies` |
| `content-block-contact` | `contact-block` | `contact-block` | `content-block-contact` |
| `features` | `differentiators-section` | `differentiators` | `features` |
| `footer` | `site-footer` | `site-footer` | `footer` |
| `head` | `site-head` | `site-head` | `head` |
| `hero-carousel` | `hero-card-carousel` | `hero-card-carousel` | `hero-carousel` |
| `image-hover-cards` | `image-hover-card-grid` | `image-hover-card-grid` | `image-hover-cards` |

Reproduce:
```sql
WITH c AS (
  SELECT cc.section_type, cc.name, cc.function, cc.updated_at,
    (SELECT count(DISTINCT p.site_id) FROM page_components pc JOIN pages p ON p.id=pc.page_id
      WHERE pc.component_id=cc.id AND pc.build_status<>'removed') AS sites
  FROM content_components cc
  WHERE cc.is_active AND cc.forked_from IS NULL AND cc.component_level='section'
    AND cc.section_type IS NOT NULL)
SELECT section_type,
   (array_agg(function ORDER BY sites DESC, updated_at DESC))[1] AS advises_function,
   section_type AS store_enforces_function
FROM c GROUP BY 1
HAVING (array_agg(function ORDER BY sites DESC, updated_at DESC))[1] <> section_type;
```

## Why it matters

1. **The advice and the enforcement are independent resolvers over one table.** The file's own
   fallback exists *because* someone recognised they must agree; the primary path never got the same
   treatment.
2. **The failure is silent and asymmetric.** `load_existing_component` is explicitly advisory and
   "never blocks generation" — so a wrong advisory produces a component built to the wrong field
   contract, which the store's guard then refuses or overwrites. The refusal surfaces far from the
   cause.
3. **It is a naming-convention drift, so it will keep growing.** Every `content-block-*` /
   `site-*` / `*-grid` / `*-list` row whose `function` does not echo its `section_type` adds one.

## Fix candidates, ranked by what closes the door

1. **Make the primary path ask the store's resolver too** — i.e. resolve the contract row via
   `resolveStorageIdentity` *always*, not only when the `section_type` query misses, and use the
   `section_type` query only to decide whether there is anything to advise at all. Makes the
   disagreement unrepresentable; the fallback already proves the shape works.

   > **CORRECTED 2026-08-25 — THIS CANDIDATE IS WRONG AND WAS RANKED FIRST. The filing lane agrees
   > (in writing, 2026-08-25).** It achieves agreement by moving the *advice* onto the resolver that
   > cannot see the problem. `resolveStorageIdentity` keys on `NormaliseToKebab(section_type)`, which
   > for the 27 divergent section_types names a **different or nonexistent** row — so the advisory
   > would stop naming the row dependents are actually bound to, and would push generations toward
   > creating duplicates. It would make the bug worse. It also cannot touch the `function`
   > non-uniqueness above, because **any** name-keyed resolver inherits the name's ambiguity.
   >
   > **The reasoning gap is the transferable part:** this ranked "make the two resolvers agree"
   > without asking WHICH of them should win. When two mechanisms disagree, "make them agree" is half
   > a fix — write down which one is better informed, and why, before ranking anything.
   >
   > **The direction that does close the door is the inverse:** the STORE honours the identity the
   > advisory resolved, carried **by `content_components.id`**, with `bugs_open/311`'s diversion
   > decision re-applied on top at write time. That removes an LLM's output from an identity decision
   > and is immune to the name ambiguity. Plan and evidence:
   > `docs024_key_docs_latest/bugfix_388_component_contract_identity/PLAN_2026-08-25_component_contract_identity.md`;
   > council submission `5252bee6-0e49-4e41-81fc-6acb014a4802`. Recorded here as an amendment rather
   > than a candidate 5, because a ranked list whose top entry is wrong is worse than one visibly amended.
2. **Have the advisory report which row it means AND which row the store would take**, and say so
   when they differ. Cheap, honest, does not change behaviour — and it would make the 27 visible to
   whoever hits one.
3. **Reconcile the data** so `function` echoes `section_type` on the 27. Rejected on sight as the
   primary fix: it is a rename across a shared vocabulary with live bindings, and it would have to be
   re-done every time a new component is named to a different convention.
4. **Do nothing but document it.** The 27 are stable and nothing is on fire, but the set grows by
   naming convention and the failure it produces is silent.

## What is NOT established

- `[UNMEASURED]` **how often the disagreement is actually reached.** It matters only on a
  *regeneration* of one of those 27 section types. `component-creator` ran 17 times in the 14 hours
  after 2026-08-24 18:55Z, but I have not attributed those runs to section types.

  > **MEASURED 2026-08-25, and the answer bounds the severity: ZERO reached firings.**
  > `agent_error_log` carries no `component_validation_rejected` row attributable to this class. The
  > `removes/renames` curve runs **10** (08-15), **9** (08-17), **74** (08-18), **4** (08-19), then
  > **zero** on 08-20 and 08-21, and exactly **one** on 08-22 — `loans-application-tracker`, where
  > `function == section_type`, so no divergence was possible. The class stops where `bugs_open/311`'s
  > diversion (08-19) and `337`'s advisory (08-22) closed the routes that were producing it.
  > **The bug is latent, as this file says in its severity line.** Pin obedience is 11 of 11 observed,
  > which at n=11 with zero failures puts the 95% upper bound on the disobedience rate near **24%** —
  > it establishes nothing either way.
  >
  > ⚠ **One wrong attribution, recorded so it is not repeated.** The `388` lane initially read eight
  > `needs_new_component` items (cancelled 08-23, all refused 3/3 on `removes/renames`) as this bug's
  > first measured damage, on the strength of a census taken 08-25. The dates refute it: the row that
  > made them look like a divergence was created **2026-08-21 18:19Z**, four days AFTER the 08-17
  > refusals, and the advisory's fallback resolver shipped 08-22. Those eight are `bugs_open/337`'s
  > blind-writer casualties. **Before attributing a past event to a present configuration, `SELECT
  > created_at` on every row the explanation depends on.** Logged in `WRONG_CALLS.md`.
- `[UNMEASURED]` **what the store's guard does on the resulting mismatch** — refuse loudly, or
  overwrite. `resolveContractViaStorageIdentity`'s comment implies a guard refuses a
  contract-breaking template, but I did not read that guard.

  > **ANSWERED 2026-08-25 — it refuses LOUDLY, and the asymmetry is the other way round from this
  > file's summary line.** `store_generated_component_action.go:452-465` diffs the old and new schema
  > field sets, appends a blocking issue naming every stranded field, calls `recordValidationRejection`
  > and returns an error. It never silently overwrites. **But the guard is keyed on the row THE STORE
  > resolved**, so: if the store resolved a different row, the refusal is real, loud, and names a
  > contract the writer was never shown (the wrong-cause retry damage of `bugs_open/345` and `337`);
  > and if the store resolved NO row, `isRegeneration` is false, the guard is **vacuous**, and a
  > parallel duplicate is created with no error and no work item. `[MEASURED 2026-08-25]` the 27
  > partition **15 loud / 12 silent** on that test.
- `[INFERRED]` that `NormaliseToKebab(section_type)` is the store's derivation in all cases; taken
  from `resolveContractViaStorageIdentity`'s doc comment and the code path it calls, not from reading
  `store_generated_component_action.go`'s own derivation end to end.

## Diagnosis-loop note (owner ruling 2026-07-31)

`090` was **not** run. Stating the substitute plainly: the mechanism claim is two resolvers quoted
from the source, one of which documents the agreement requirement in its own comment; the population
claim is a single query over the live table, reproduced above, which would have returned zero had the
resolvers agreed. The residual uncertainty is recorded as `[UNMEASURED]`/`[INFERRED]` above rather
than hidden. A `090` run is worth its cost before acting on candidate 1, which changes what every
regeneration is advised.

## Related

- `bugs_closed/378` — where this was found; it moved the count from 29 to 27 and did not cause it.
- `bugs_open/357` — the other live case of a component's stored identity disagreeing with what it
  actually is.
- `bugs_open/337` — cited in `load_existing_component_action.go`'s own comments as the source-
  vocabulary failure that shaped the fallback path.

**PRIOR ART, found 2026-08-25 and not known at filing — this bug already has a register entry.**
`docs026_concept_register/register/component-lifecycle.md`, **CLC-006 — "F4 — regen-vs-create keyed on
the LLM-chosen function (silent fork)"**, status *partial*, records that *"a store-side advisory ... is
deliberately advisory-only and unbuilt, since multiple components per section_type can be legitimate"*,
and asks in its `verify-later` for exactly two things: *"duplicate non-forked function rows in
content_components; whether any store-side advisory exists"*. **Both are now answered and both answers
are the bad one** — two duplicate pairs exist from the `generated` route
(`tool-archetype-taster-quiz`, `tool-gripper-payload-calculator`, both second rows born 2026-05-06 with
`function == section_type`), and no store-side advisory exists. **388 is CLC-006's unbuilt half,
rediscovered from the data.**

⚠ CLC-006's stated reason for never building it is a **live constraint on any fix**: *multiple
components per section_type can be legitimate.* A fix must change WHICH row a regeneration lands on,
never HOW MANY rows a section_type may have — so no refusal, no rename, no reconciliation.

- `docs024_key_docs_latest/bugfix_388_component_contract_identity/` — the working lane (PLAN, NOTES,
  RUNBOOK, README). The `090` diagnosis run this file asked for was filed
  (`2f80ff5e-96db-4d9f-8dfa-f2b8ea9d52d0`) and returned **UNVERIFIABLE — stopped at the iteration cap**;
  it scoped the fallback function instead of its caller and settled nothing. Recorded in the lane NOTES
  rather than quietly dropped, because a discarded UNVERIFIABLE looks exactly like a run never filed.
