# 362 — 69 fields across 17 active components declare a data source that resolves nowhere, and six of those components are live on 46 pages

**Filed 2026-08-22** by the `bugfix_309_unclickable_index_cards` lane, as the routing
target for `bugs_open/309`'s grandfather baseline. **Status: OPEN.** Nothing here is a
new discovery — this is the population `309` measured, made visible and given an owner
so that "grandfathered" means *routed*, not *excused*.

> **On the owner ruling of 2026-07-31** (a `bugs_open/` file asserting a cross-cutting or
> structural root cause is not filed until it has been through the `090` loop, or the
> filing session states plainly why it substituted equivalent first-hand verification).
> **Stated plainly: `090` was NOT run for this file, deliberately, and here is the
> substitution.**
>
> This file asserts no new causal theory. Its mechanism is `bugs_open/309`'s, which the
> loop already graded **CONFIRMED** on its first iteration set (run correlation
> `6e578bf5-778a-4e72-aab2-0531e45c07d8`, 2026-08-18) — independently re-deriving the
> chain and citing `plan_sections_action.go`'s `on_missing = "skip_field"` and the
> `site_specs` aspect emptiness itself. What this file adds is a **census**, not a cause:
> which rows carry that already-confirmed shape, and how much live page surface each sits
> on.
>
> The first-hand verification substituted for a second run, every part of it re-runnable
> from the query in the lane RUNBOOK: the population counted at the live DB; the claim
> "resolves nowhere" checked at the **resolver** rather than inferred from the guard
> (`plan_sections_action.go:623` returns `(nil,false)` for a dotless source;
> `resolveSpecAlias` step 2 is `if aspect != "identity" { return nil, false }`, so no
> phantom aspect can be rescued); and the census itself **executed as code** against the
> real library by the audit binary, which reproduces exactly these 69 across these 17
> components. A hand-counted census and a machine-counted one agreeing is a stronger check
> than either alone, and it is the one thing a `090` run could not have supplied.
>
> **What that substitution does NOT cover, said rather than left implied:** it establishes
> the population, not the repair. Each of the four repair shapes below is a judgement about
> a specific component, and none of them has been graded by anything.

**Every row of `component_source_baseline.json` names this file.** When one of these is
repaired, the daily check goes RED with a stale-entry message naming the exact line to
delete. That deletion is the burn-down: **the baseline's shrink history is this bug's
progress bar**, and there is no other bookkeeping to keep in step.

## What the defect is, in one paragraph

A component's `input_schema` declares, per field, where its value comes from. If that
`source` names something the platform cannot resolve — a `site_specs.<aspect>` no site
has ever carried, a `query.*` name the resolver never registered, or a prefix outside
the vocabulary — then `plan_sections`' resolver returns nothing, `on_missing` defaults
to `skip_field`, the key is omitted from `content_data`, and a `{{if}}`-gated template
silently drops the markup that would have used it. **The page renders complete and
data-less**, which is indistinguishable from success at every stage. That is
`bugs_open/309`: fundamentallyai.com's article index served six complete-looking cards
with zero links for four months.

`309` fixed the motivating page (migration 478) and shut the generation door
(**CLC-018**, the birth gate). This file is the population that was already through it.

## The 69, by component — repair-owed first

`[MEASURED 2026-08-22]` live `content_components WHERE is_active`; live instances are
`page_components → pages` with `status IN ('active','deployed')`.

| component | live instances | dead fields | class | sources |
|---|---|---|---|---|
| `info-card-grid` | **32** | 1 | prefix_outside_vocabulary | `config` (bare, no dot) |
| `Latest News Feed` | **6** | 1 | unregistered_query | `query.pages` |
| `featured_article` | **3** | 7 | unregistered_query | `query.featured_post` |
| `category-listing` | **2** | 3 | unregistered_query | `query.category`, `query.category_posts` |
| `testimonials` | **2** | 1 | phantom_aspect | `site_specs.social_proof.testimonials` |
| `social_proof` | **1** | 1 | phantom_aspect | `site_specs.social_proof.testimonials` |
| `footer-with-disclaimer_pre_037` | 0 | 18 | phantom_aspect | `site_specs.legal.*` |
| `Pricing Tiers` | 0 | 9 | phantom_aspect | `site_specs.pricing.tiers[N].*` |
| `header-with-categories_pre_037` | 0 | 8 | phantom_aspect | `site_specs.categories.cat_N_*` |
| `featured-inventory` | 0 | 7 | phantom_aspect | `site_specs.inventory.*` |
| `content-sidebar` | 0 | 4 | phantom_aspect | `site_specs.nav.link_*_url` |
| `filtered-result-grid` | 0 | 2 | unregistered_query | `query.comparison_results`, `query.comparison_filter_types` |
| `product-details_pre_037` | 0 | 2 | phantom_aspect | `site_specs.product.sku`, `.category` |
| `webdesign.co.uk Site Footer` | 0 | 2 | prefix_outside_vocabulary | `nav`, `site` |
| `header-docs` | 0 | 1 | phantom_aspect | `site_specs.social.github_url` |
| `product-card-with-cta` | 0 | 1 | unregistered_query | `query.affiliate_products` |
| `webdesign.co.uk Site Header` | 0 | 1 | prefix_outside_vocabulary | `nav` |

Totals **as of 2026-08-22**: **51** phantom_aspect · **14** unregistered_query · **4** prefix_outside_vocabulary = **69**, over **17** components, **6** live on **46** instances.

> **Every figure here is a census and carries its date, per the owner ruling of 2026-08-22 —
> a census does not go wrong, it goes STALE BY ADDITION and reads as current for ever.**
> Re-run it before quoting: the query is in the lane RUNBOOK, and `config-key-audit
> --component-source-vocabulary --emit-baseline` regenerates the whole list from the shipped
> rule. **The daily check is the live answer** — it reports the current population every
> morning into `doc_notes` (`source='component-source-vocabulary-check'`), so the freshest
> count is always one query away and this table never has to be trusted.

## Two things a reader will get wrong if they are not told

- **`info-card-grid` is the biggest live surface and the least obvious entry.** Its
  `carousel` field's source is the bare string `config` — **no dot**. That is not a
  `config.*` source: `plan_sections_action.go:623` splits on the first `.` and returns
  `(nil, false)` when there is none, so the field is dropped on all **32** live
  instances. A census that classifies by prefix alone does not see it.
- **A phantom aspect cannot be rescued by the alias fallback**, so "resolves nowhere" is
  literal. `resolveSpecAlias` step 1 needs `identityContainerAspects[aspect]` populated;
  step 2 is `if aspect != "identity" { return nil, false }`. Checked at the source
  before the claim was made, because the opposite error — a check flagging fields that
  actually work — is what gets a check switched off.

## Repair shapes, in the order that closes the door

1. **Repoint a listing at a registered `query.*` source** — the migration-478 pattern,
   and the right answer for `featured_article`, `category-listing`, `Latest News Feed`,
   `filtered-result-grid` and `product-card-with-cta`. Title and URL then come from the
   same page row and **cannot disagree**, which is what dissolved 309's "card 4
   advertises an archived page" symptom into the same root cause rather than a second
   repair. ⚠ It usually forces a template change too (numbered flat blocks → a `range`).
2. **Register the query** where the name is reasonable and the data exists —
   `queryresolve.queryHandlers` is one map and `IsKnownQueryName`/`KnownQueryBases`
   answer from it, so a new entry is picked up by the validator and the birth gate with
   no second edit.
3. **Seed the missing `site_specs` aspect** — weakest, and 309 §5 ranked it last for a
   reason: it makes one component work and leaves the trap armed for the next.
4. **Deactivate the component** where it is dormant AND superseded. This is a repair
   judgement, not bookkeeping: some of the eleven are awaiting regeneration, which the
   birth gate will now force to repoint sources anyway.

## How a repair is verified

Not by the component row. `[the 309 landmine]` **at the SERVED page**, and only for the
six with live instances: the field must appear in `content_data` after a rebuild AND the
markup it gates must be present. For the eleven dormant ones there is no artefact to
check, which is exactly why they are dormant-and-conditional rather than closed.

Then delete their lines from
`docs/agent_docs/docs024_key_docs_latest/bugfix_309_unclickable_index_cards/component_source_baseline.json`
and confirm the daily check returns to exit 0.

## What is already done, so nobody redoes it

- **The generation door is shut** — CLC-018's birth gate, live since v1.0.1314, and
  `[MEASURED 2026-08-22]` zero offending components created or updated since.
- **The at-rest door is watched, and PROVEN IN-CLUSTER 2026-08-22** — `config-key-audit
  --component-source-vocabulary`, daily CronJob `component-source-vocabulary-check` at 07:20
  UTC on image `v1.0.1326`: manual Job run, **pod exitCode 0**, `doc_notes` row written. It
  calls the birth gate's own function, so the two cannot drift.
- **A repair is not finished until the baseline shrinks.** Delete the entry's lines from
  `deployments/kustomize/services/component-source-vocabulary-check/base/component_source_baseline.json`
  (the docs path is a symlink to it) **and re-apply the overlay** — the baseline is mounted
  from a ConfigMap, so an edit that is not applied leaves the cluster on the old copy. Until
  you do, the check goes RED with a stale-entry message naming the exact lines.
- **A dormant component that gets DEPLOYED turns the job red**, so this backlog cannot
  quietly grow live surface while it waits.

## Ownership

Lane: `docs/agent_docs/docs024_key_docs_latest/bugfix_309_unclickable_index_cards/`
(PLAN_2026-08-22, NOTES, RUNBOOK carry the census query and its four measurement
gotchas). Related: `bugs_open/309` (the motivating case and the birth gate),
`bugs_open/238` (the sibling key-loss mechanism), CLC-018 / CLC-025 in the concept
register.
