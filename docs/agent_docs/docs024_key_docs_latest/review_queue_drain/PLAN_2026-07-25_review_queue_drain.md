# PLAN 2026-07-25 — draining the `needs_human_review` queue (`bugs_open/033`)

**Workstream opened:** 2026-07-25, by the bugfix-033 thread.
**Bug:** `bugs_open/033_HANDOFF_2026-07-20_human_review_queue_has_no_working_surface.md`
**Prior owner rulings, both already given — this plan implements them, it does not re-ask:**

- 2026-07-20: *"split it — auto-drain what can be, queue the rest."*
- 2026-07-22 (re-affirmed via AskUserQuestion): *"this queue **is a queue, not a
  bin** — a human works these."*

The visibility half of 033 shipped on 2026-07-20 (read cap, pipeline filter,
VPN access; chassis + dashboard `v1.0.1141`). What is left is the drain.

---

## Where the queue actually is, measured live 2026-07-25

```
site_work_items WHERE status='needs_human_review'  : 370   (was 303 at grounding, 292 at filing)
  build pipeline                                   : 224
  content pipeline                                 : 145
  maintenance                                      :   1
ever actioned through the admin API (approved_by)  :   0
ever actioned through the admin API (result.resolved_by='admin')
                                                   :   0
```

The surface has been visible and reachable for five days. In those five days the
queue grew by 67 and drained by 0. That is not an argument that the owner ruling
was wrong — it is the reason the auto-drain half of it has to exist before a
human can sensibly work the rest.

### The queue by class

| item_type | n | oldest |
|---|---|---|
| `cta_names_unknown_destination` | 69 | 2026-07-14 |
| `unresolved_cta` | 68 | 2026-06-22 |
| `needs_section_data` | 45 | 2026-03-15 |
| `required_fields_missing` | 45 | 2026-07-14 |
| `needs_page` | 28 | 2026-07-12 |
| `content_rewrite` | 26 | 2026-04-08 |
| `voice_tells` | 25 | 2026-07-17 |
| `needs_content_page` | 19 | 2026-04-18 |
| `image_source_unsatisfiable` | 17 | 2026-07-17 |
| 11 smaller types | 28 | — |

---

## The finding this plan is built on

> **321 of the 370 parked items describe a page that has been redeployed since
> the item was filed.** (Measured: item's `spec->>'page_name'` joined to
> `pages.deployed_at > work_item.created_at`.) 033's grounding pass put this at
> "121 of 126 page-linked items"; scoped to the whole queue it is far larger.

A finding filed against a page that has since been rebuilt is not automatically
false — but nothing re-checks it, so nobody can tell which it is. **That
indistinguishability, not the backlog size, is what makes the queue unworkable.**
A human opening it today cannot know whether item #1 of 370 describes the live
site or a page state that stopped existing in April.

### Proof that the ghosts are real, not theoretical

`leopardessconsulting.co.uk` / `how-we-work`, two items parked 2026-07-10:

```
d0d5f910 unresolved_cta 'hero'           : missing cta_url, secondary_cta_url
bad6aa52 unresolved_cta 'call-to-action' : missing primary_cta_url, secondary_cta_url
```

The page was redeployed 2026-07-18. Its currently-deployed components:

```
hero            cta_url=/tools/password-entropy.html  secondary_cta_url=/services.html
call-to-action  primary_cta_url=/tools/password-entropy.html  secondary_cta_url=/services.html
```

Every field both items say is missing is populated on the live page. Both items
are ghosts, and a human working the queue would have spent their attention on
them.

### Why nothing can currently tell a ghost from a live finding

`insertWorkItem` (`platform/orchestration/actions/load_work_item_actions.go:1111`)
inserts discovery findings with:

```sql
ON CONFLICT (site_id, item_key) WHERE item_key IS NOT NULL AND status NOT IN (<terminal>)
DO NOTHING
```

`RunDiscoveryChecksAction` counts the suppressed insert into a local `skipped`
tally and nothing else (`discovery_checks.go:166`). So when a check runs again
and **re-confirms** a still-open finding, no column, no jsonb key and no log row
on that item changes. A re-confirmed finding and an abandoned one are
byte-identical in `site_work_items`.

[FILED FOR DIAGNOSIS 2026-07-25, corr `c19ed5b2-6d53-492a-af91-e78e175591d5`]
— the claim above is structural and load-bearing, so it goes through the loop
before it is asserted anywhere durable. Verdict to be recorded in NOTES.

---

## What we are building

**`revalidate_review_queue`** — a deterministic, no-LLM sweep that gives the
queue an exit and gives the survivors a trust signal.

Per parked item, look up a **revalidator** for its `item_type` and re-evaluate
the original finding against currently-deployed state. Three verdicts:

| verdict | meaning | what we write |
|---|---|---|
| `resolved` | the condition the item describes is provably no longer true of the live page | close: `status='complete'`, `resolution_path='auto:revalidated'`, evidence into `result.revalidation` |
| `still_holds` | the condition is provably still true | **no status change**; stamp `result.revalidation` so the human sees "re-confirmed <date>" |
| `unknown` | cannot be determined (no revalidator, ambiguous data shape) | stamp the reason; leave alone |

### Why auto-closing is safe

Every terminal status is excluded from the `idx_swi_dedup` predicate. So closing
an item **releases its dedup key**: if the finding is in fact still true, the
producing check raises it again, fresh. A wrong `resolved` verdict costs one
re-raise, not a lost finding. This is what makes the drain a reversible
operation rather than a bulk delete, and it is why the sweep is allowed to be
wrong.

> **QUALIFIED 2026-07-25 — the council gate was right to press on this, and the
> claim above is weaker than it was written.** (corr `ccba9c51`, `bug_historian`,
> medium: *"the entire safety case … rests on an unverified assumption"*.) It was
> reasoned from the index predicate, not measured. Verified afterwards:
>
> - **What holds.** All three producers insert with `ON CONFLICT DO NOTHING` on a
>   deterministic `item_key` (`resolve_internal_links_action.go:257`;
>   `plan_sections`' `createDeferredItems`; `RunDiscoveryChecks` →
>   `insertWorkItem`), so a terminal row genuinely does not block a re-raise.
> - **What does NOT hold.** "The check will run again." All three fire on a **page
>   build** or a discovery pass over that site — never on a timer. A page that is
>   never rebuilt again never re-raises, and a wrong close on such a page **is** a
>   silent loss.
> - **How untested this path is:** across the platform's entire history only **8**
>   items of these three types have ever reached a terminal status (7
>   `needs_section_data`, 1 `required_fields_missing`, and **zero** of the 70
>   `unresolved_cta`). The re-raise has essentially never been exercised here.
>
> The mitigation that holds unconditionally is the audit trail: every close
> records the fields it judged populated in `result.revalidation` and stamps
> `resolution_path='auto:revalidated'`, so a wrong close is individually
> identifiable and reversible by SQL whether or not anything re-raises.

The asymmetry is deliberate: `resolved` requires **positive** evidence that the
finding no longer holds. Anything ambiguous is `unknown` and stays queued.

### Revalidators shipping in v1

All three reduce to one primitive — *are the fields this item names now non-empty
on the currently-deployed component?* — keyed on **(page_name, slot_name)**,
never on `spec.component_id`.

> **Landmine, measured today.** `page_components.id` is not stable across
> re-renders. Keyed on `spec->>'component_id'`, 30 of 30 `needs_section_data`
> items and 11 of 45 `required_fields_missing` items resolve to a component that
> no longer exists — which would read as "target gone" when the section is in
> fact right there under a new row id. Key on the slot.

| revalidator | item_type | reads |
|---|---|---|
| CTA fields | `unresolved_cta` | `spec.missing[]` (field names) × `content_data` on `spec.section_name` |
| required fields | `required_fields_missing` | `spec.missing_fields[]` × `content_data` on `spec.slot_name` |
| section data | `needs_section_data` | `spec.missing[].field` × `content_data` on `spec.section_name` |

### What it drains on day one (measured, conservative)

| class | parked | provably resolved | still holds | not determinable |
|---|---|---|---|---|
| `unresolved_cta` | 68 | **39** | 29 | 0 |
| `required_fields_missing` | 45 | **10** | 4 | 31 (component row carries `content_data IS NULL`) |
| `needs_section_data` | 45 | **2** | 2 | 41 |
| **total** | **158** | **51** | **35** | **72** |

51 of 370 closed on the first non-dry sweep, with evidence, and the rate keeps
working as pages rebuild. The 72 "not determinable" are honest: those components
render from something other than `content_data`, and guessing would be exactly
the failure this whole bug is about.

---

## Corrections to 033 as filed

> **CORRECTED 2026-07-25 — the auto-drain the ruling named would drain nothing.**
> 033 says: *"The one genuine automated consumer, `reconcile_section_data_action.go`
> (re-opens `needs_section_data` when query-sourced data later resolves — **48
> items of the queue**), is registered as an action but wired to 0 live agents"*,
> and fix candidate A is "wire it". Measured live today: of the 45
> `needs_section_data` items parked, **0** have all-`query.*` missing sources —
> 30 carry `site_specs.*` / `site_assets.*` sources and 15 carry `missing: null`.
> `ReconcileSectionDataAction` skips both (it requires every missing field to be
> `query.`-prefixed). Wiring it today re-triggers **zero** pages. The action is
> not wrong; the population it was built for is not the population in the queue.
> The section-data revalidator above is the generalisation that does apply.

> **CORRECTED 2026-07-25 — the "stale 121" is now 321.** Not a contradiction, a
> re-measure at a wider scope (whole queue, not just `page_id`-carrying items).

---

## Explicitly NOT in scope

- **`cta_names_unknown_destination` (69).** The check belongs to the live
  `cta_link_integrity` workstream / `bugs_open/023`, which is mid-flight on it
  (18 of them are known false positives of the excluded-area branch). Two threads
  on one check is the thing to avoid. A revalidator for it is a later addition,
  after 023 lands.
- **D2 (residue aging).** 78 of the 370 carry an `error` and a machine
  `handler_agent` — they are failures parked by `FailWorkItemAction`'s
  `status_override` branch, which does not increment `attempt_count`, so they
  neither retry nor age out. That is a real defect and a real owner decision; it
  is recorded in 033 and stays there. The revalidator does not touch them.
- **D3 (identity/auth).** Unchanged: the admin handlers have no auth context, so
  "record who decided" is blocked on an auth decision. `resolution_path` now gets
  a real writer for the *machine* path only, which is honest — `auto:revalidated`
  claims no human.
- **`updated_at` maintenance generally** — `bugs_open/035` owns it. The sweep
  sets `updated_at` on rows it touches and does not go near the general path.

## Phasing

1. **P1** — action + revalidator registry + unit tests. *(this session)*
2. **P2** — seed as a manually-triggered agent, `dry_run=true` (mirrors
   `diagnosis-superseded-reviews`), after the image carrying the action is live.
3. **P3** — review the dry-run report, flip `dry_run=false`, measure the drain.
4. **P4** — council review; commit; 033 closes when the drain has run non-dry and
   the queue is measurably smaller.
