# PLAN 2026-07-31 — `bugs_open/154`: a first-class column that no handler can see

**Bug:** `154_HANDOFF_2026-07-30_tool_improver_fails_at_load_tool_when_the_item_came_from_tool_auditor.md`
**Session:** "bugfix 19". **Ownership checked 2026-07-31 ~19:10** — no live `.jsonl`
transcript is working `tool-improver`/`load_tool`/`improve_tool`; the one active
session with a high symbol count (`631baa00`) had *deliberately declined* to
dispatch a `tool-improver` item on another lane's owned page, which is the
opposite of owning the fix. `who-owns.py` reported one filing commit and nothing
since.

## What the bug file asserted, and what it left open

154 recorded the symptom (two `improve_tool` items dead at `load_tool` with
`input_data.component_id resolved to nil`) and the tell (**the failures are the
rows whose `component_id` IS set**). It marked its own explanation `[INFERRED —
not yet read in the code]` and said: *confirm by reading `build-dispatch-loop`'s
`process_item` input_mapping and `tool-improver`'s `load_tool` step before
acting.* That instruction was right and is the whole of this plan's first phase.

## Root cause — read, not inferred

`site_work_items` carries **`component_id`, `entity_id` and `affected_url` as
real columns**. `LoadWorkItemsAction`
(`platform/orchestration/actions/load_work_item_actions.go`) builds the
`current_item` map from a `SELECT` that lists `wi.page_id` and **none of those
three**. So the only path a dispatcher can reference is
`current_item.spec.<key>` — a copy the creating agent must remember to
duplicate into the `spec` JSONB.

The live `build-dispatch-loop` maps, in `process_item.sub_workflow.call_handler`:

```json
"component_id?": "current_item.spec.component_id"
```

`ResolveInputMapping` treats the `?` suffix as optional and **silently skips**
an unresolved path (`input_contracts/input_mapping.go:122-129`). `tool-improver`'s
`load_tool` then runs `query_database` with `params: ["input_data.component_id"]`,
which is a hard error on nil.

**So the creator that used the schema properly is precisely the one whose items
cannot be dispatched.** Nothing is wrong with `tool-auditor`, nothing is wrong
with `tool-improver`; the framework drops the value between them.

## Evidence (all first-hand, 2026-07-31)

| claim | check | result |
|---|---|---|
| the split is exact | `site_work_items` census, `item_type='improve_tool'` | 4/4 `tool-auditor` rows column-set + spec-missing; 16/16 other creators the reverse |
| still live | same census | fresh failure **`a5d11c86`, robot-hands.com, 2026-07-31** — one day after filing |
| the fix is not inert | resolve the 4 column values | **4/4 exist in `content_components`, `is_active`, and join to a page** — exactly `load_tool`'s query |
| fleet size | census over all item types | `component_id` 21 col-set (**16 spec-missing**), 235 spec-only |
| `affected_url` | same | **0 rows anywhere** — the column has never been used |
| `entity_id` | same | 0 col-set, 12 spec-only |
| nothing reads the column path today | `agent_definitions` LIKE `%current_item.component_id%` | **0 agents** ⇒ exposing it is additive and inert |

## Design — and the candidate I rejected on evidence

Two ways to make one dispatcher path serve both populations:

**(A) Backfill the resolved value into the item's `spec` map.** Attractive
because it needs **no config change at all** (`current_item.spec.component_id`
would just start resolving). **Rejected.** `rerender-pages` reads
`input_data.spec.component_id`, and `create_rerender_items_action.go:219` gates

```go
scoped := (reason == "section_data_resolved" || reason == "image_landed") && componentIDStr != ""
```

on it. Writing into `spec` can therefore **flip a site-wide rerender into a
component-scoped one** — a behaviour change nobody asked for, in a path that has
nothing to do with this bug. (The two live `needs_rerender` rows in that shape
happen to carry an empty `reason`, so today it would not fire — but that is a
data-dependent argument that goes stale, and the design should not rest on it.)

**(B) Expose the columns top-level on the item map, column-first with a
`spec.<key>` fallback. CHOSEN.** `input_mapping` resolves exactly one source path
per destination and has no coalesce syntax, so the coalesce must happen in Go.
Every existing `spec.*` reader then reads exactly what it reads today.

### Deliberate non-changes

- **`page_id` stays column-only.** It is already exposed, so there is no bug to
  close, and **218 rows** have a NULL column with a `spec.page_id` — every one
  would newly gain `current_item.page_id`. Widening what reaches a handler
  changes it without editing it (the `149` landmine). Symmetry is not a reason.
- **`spec` is never mutated** — see (A).
- **The key stays ABSENT when neither source has a value**, rather than becoming
  `""`. An optional mapping path that *resolves* is forwarded; one that is
  *missing* is skipped. Handlers gate on presence, so materialising `""` would
  turn "not supplied" into "supplied empty".

## Phases

1. **Confirm the cause first-hand** — done, table above.
2. **Diagnosis loop** (`090`) on the structural claim — filed, `RUN_CORRELATION_ID=21758756-d7b3-444a-844e-b37e09b5c9ce`.
   CLAUDE.md makes this the default for a durable structural claim, and this is one.
3. **Go fix + regression test** — done; induced-fault proven non-vacuous.
4. **Council gate** — platform-code change, so it goes through.
5. **Config change, AFTER the image is live** — `build-dispatch-loop`'s
   `component_id?` mapping from `current_item.spec.component_id` →
   `current_item.component_id`. **Ordering is load-bearing:** the coalesce lives
   in the binary, so flipping the config first would break the 235 spec-only
   rows until the image ships. Image first, then config (CLAUDE.md).
6. **Verify at the artefact** — pod-grep for a symbol the change added, then
   dispatch a `tool-auditor`-raised item and watch it get past `load_tool`.

## The second finding in 154, and what I am doing with it

154 also noted that both failing items share the item_key
`audit_fix_gamesdesign.co.uk` — **one key per site, not per tool** — so a
site-scoped key on a per-tool fix cannot express *which* tool it is fixing, and
rows accumulate once one goes terminal. That is a **separate defect in
`tool-auditor`'s item creation**, not in the dispatch path, and it is marked
`[UNVERIFIED]` in the bug file as possibly intended. Fixing it would change what
gets deduped fleet-wide. **Left out of this change deliberately** and recorded
below, not silently dropped — it wants its own measurement and its own review.
