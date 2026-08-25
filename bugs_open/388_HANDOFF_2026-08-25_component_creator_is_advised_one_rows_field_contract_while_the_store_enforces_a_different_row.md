# 388 — `component-creator` is advised ONE row's field contract while the store enforces a DIFFERENT row, on 27 of 117 section types

**Filed:** 2026-08-25 by the `bugfix_378_usage_count_derived` lane, found while answering a council
objection (`bug_historian`, medium) about whether repointing an `ORDER BY` could re-shape an enforced
schema. It could not — but measuring that question exposed this, which is a separate and larger
mismatch that predates `378` entirely.

**Severity:** latent, structural. Nothing errors. The cost is that the advice given to the component
writer and the contract actually enforced on it can name different rows, so a writer that obeys its
instructions can still be refused — or worse, silently overwrite a row it was never told about.

**Status: OPEN, not started.** Diagnosis below is first-hand; no code written.

## The finding in one paragraph

`load_existing_component_action.go` tells `component-creator` which field names to preserve. It picks
that row by `WHERE section_type = $1 … ORDER BY <most used>, updated_at DESC`. But the **store** —
`store_generated_component_action.go`, via `resolveStorageIdentity` — decides which row a
regeneration *overwrites and enforces* by **function name**, derived as
`NormaliseToKebab(section_type)`. Those are two different resolvers over the same table, and when the
winning row's `function` is not equal to its `section_type`, they name different rows.

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
- `[UNMEASURED]` **what the store's guard does on the resulting mismatch** — refuse loudly, or
  overwrite. `resolveContractViaStorageIdentity`'s comment implies a guard refuses a
  contract-breaking template, but I did not read that guard.
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
