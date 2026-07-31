# 154 — `tool-improver` dies at its first step on items raised by `tool-auditor`: `input_data.component_id resolved to nil`

**Filed:** 2026-07-30.
**Status: FIXED IN CODE + REGISTERED, 2026-07-31 (session "bugfix 19"). OPEN only
until the image rolls AND the config half is applied.** Workstream docs:
`docs/agent_docs/docs024_key_docs_latest/bugfix_154_work_item_routing_columns/`.

> **THE `[INFERRED]` BLOCK BELOW WAS RIGHT ABOUT THE SPLIT AND WRONG ABOUT WHERE
> THE DEFECT LIVES.** It guessed the other creators "carry enough in
> `spec`/`item_key` for `load_tool` to resolve the tool another way". They do not
> — they carry `spec.component_id`, the same field, in the other of the two
> stores. **Neither agent is at fault. The framework drops the value between
> them:** `site_work_items` has `component_id`, `entity_id` and `affected_url` as
> real columns, and `LoadWorkItemsAction`'s SELECT listed only `page_id` among
> them — so `current_item` never carried them and `current_item.spec.<key>` was
> the only reachable path. The optional `"component_id?"` mapping then skips
> silently (`ResolveInputMapping`, `input_mapping.go:122-129`) and the nil
> surfaces two agents later at `query_database`.
> **So the creator that used the schema properly is the one whose items cannot be
> dispatched.** The instruction to read both configs before acting is what turned
> this into a framework fix instead of a patch on `tool-auditor` — it was the
> right call and it is why this file's own guess did no harm.
>
> **Still live at fix time:** `a5d11c86` (robot-hands.com) failed **2026-07-31**,
> one day after filing. 4/4 `tool-auditor` rows column-only and failed; 16/16
> rows from three other creators spec-only and fine.
>
> **Fix (committed):** `LoadWorkItemsAction` selects the three columns and
> exposes them top-level via `setRoutingField`, **column first, then
> `spec.<key>`**, so one dispatcher path serves both populations. Registered as
> **WDS-014** in the concept register; landmine appended to `LANDMINES.md`.
> Regression test `load_work_items_routing_fields_test.go`, induced-fault proven.
>
> **Diagnosis loop `21758756-d7b3-444a-844e-b37e09b5c9ce`: CONFIRMED**, first
> iteration, independently re-reading the same functions. It found a contrast
> pair I had not cited — the same mechanism, both directions, on one site:
> `a5d11c86` (column set, `spec={}`) **failed** at 13:21:26 today, one second
> after the `load_tool` "resolved to nil" line; `265f0c41` (column NULL,
> `spec.component_id` present) **complete**. All four symptom points `[explained]`.
> **Council `10be5ed9-3bd0-45ed-b6bb-4385a887967d`: APPROVED**, 6 advisory
> objections, none high-severity, 8 seats. (An earlier attempt was rejected at
> `persist_submission` for a malformed edit path of mine, before any seat ran —
> see `WRONG_CALLS.md`; resubmitted under the same correlation.) Four objections
> landed on the **un-applied** migration and were answered into it: the jsonb
> path is now asserted at apply time rather than assumed, `snapshot_agent()`
> replaces a hand-rolled pre-image, and a mechanically-derived pre-flight count
> runs immediately before the UPDATE.
>
> **One objection narrowed this bug's own claim, and it was right.** The Go half
> exposes THREE columns; the migration rewires ONE mapping — so the class is
> closed for `component_id` and **not** for `entity_id`/`affected_url`. Measured:
> both have **0** rows with the column set; `affected_url` is read by nothing and
> `entity_id` by exactly one agent (`asset-deployer`, via
> `input_data.spec.entity_id` — the `spec` passthrough, **not** a dispatcher
> mapping, so the coalesce cannot reach it). Nothing is broken through them
> today, but the first creator to write `entity_id` on the column hits this
> identical bug, and the fix then needs TWO edits (add an `entity_id?` mapping
> **and** repoint `asset-deployer`). Recorded in `LANDMINES.md` + WDS-014.
>
> **REMAINING, and the ONLY reason this is still open — ordering is
> load-bearing:** the coalesce is in the binary, the path that uses it is in the
> DB. After the image is live and **pod-grepped on every replica**, apply
> `"component_id?": "current_item.spec.component_id"` →
> `"current_item.component_id"` on `build-dispatch-loop`. Flipping it first
> strands the 235 spec-only rows. Exact statement in the workstream RUNBOOK.
>
> **NOT fixed here, deliberately:** the site-scoped `item_key` in the second
> finding below — a defect in item *creation* with fleet-wide dedup
> consequences, wanting its own measurement and its own review.
> **NOT changed, deliberately:** `page_id` keeps its column-only behaviour;
> **218** rows would newly gain `current_item.page_id` if the fallback were
> extended to it for symmetry.
**Found by:** the first `improve_tool` items ever to reach a handler — see
`bugs_open/083` BY SLUG (`…detected_findings_never_reach_a_handler`), whose
promoter was run by hand on owner instruction 2026-07-29. Until that run, no
`improve_tool` item had ever been dispatched, so this defect had never been
reachable and nothing had ever observed it.

## Symptom

Of the four `improve_tool` items dispatched on gamesdesign.co.uk, **two failed
immediately** at `tool-improver`'s first step:

```
step load_tool failed: failed to execute action query_database:
query param path 'input_data.component_id' resolved to nil
(code: CHILD_ORCHESTRATION_…)
```

## The tell: it is the items WITH a `component_id` that fail

```
 id         | status    | component_id IS NULL | created_by             | item_key
 e7ea0125…  | complete  | t                    | tool-acceptance-agent  | acceptance_fail:tool-loot-table-balancer:…
 e23548b6…  | complete  | t                    | design-discovery-agent | tool_health:tool-bayesian-ranking:…
 7c2d898a…  | failed    | f                    | tool-auditor           | audit_fix_gamesdesign.co.uk
 ee745694…  | failed    | f                    | tool-auditor           | audit_fix_gamesdesign.co.uk
```

Both successes have `component_id IS NULL` on the row. Both failures have it
**set**. So the error is not "the column is empty" — the column is populated
precisely where the workflow says it resolved to nil. The failing value is
`input_data.component_id` *inside the orchestration*, which means the row's
column is not what feeds it, and the two creators produce items that travel by
different routes.

`[INFERRED — not yet read in the code]` that `tool-acceptance-agent` and
`design-discovery-agent` items carry enough in `spec`/`item_key` for `load_tool`
to resolve the tool another way, while `tool-auditor`'s items rely on a
`component_id` that the dispatch does not map into `input_data`. **Confirm by
reading `build-dispatch-loop`'s `process_item` input_mapping and
`tool-improver`'s `load_tool` step before acting** — the shape above is evidence
of a split, not proof of which half is wrong.

## Second, smaller finding in the same rows

Both failing items share the item_key **`audit_fix_gamesdesign.co.uk`** — one key
per *site*, not per tool. `idx_swi_dedup` is unique on `(site_id, item_key)` only
`WHERE status NOT IN (<terminal>)`, so two rows with this key coexist as soon as
one goes terminal, and a site can accumulate an unbounded series of them. Whether
that is intended is `[UNVERIFIED]`; it is worth an owner's eye because a
site-scoped key on a per-tool fix cannot express *which* tool it is fixing.

## Why this matters more than 2 failed rows

`bugs_closed/010` built a convergence guard that escalates after two failed fix
cycles, and `bugs_open/126` warned that a repair loop inherits the authority of
whatever test it is given. Both have been **theoretical since they were written**
because nothing was ever dispatched (083). The moment dispatch happened, half the
population failed before the fixer did any work — so the guard still has not been
exercised, and the population it will eventually judge is smaller than it looks.

## How to verify a fix

Dispatch one `tool-auditor`-raised `improve_tool` item and assert the workflow
gets past `load_tool`:

```sql
SELECT current_step, status FROM orchestration_states
WHERE owner_agent_type = 'tool-improver' ORDER BY created_at DESC LIMIT 3;
-- expect current_step past load_tool, not a failure at it
```

Note the retention trap: `orchestration_states` history is short (an all-history
query for `owner_agent_type='improvement-loop'` returned exactly ONE row on
2026-07-29), so capture evidence at the time rather than expecting to query it
back later.

## What DID work, for contrast

`e7ea0125` — the Tier-4 mobile-fit catch from `bugs_open/131` § B — ran clean:
claimed 18:14:12Z, `tool-improver` 18:14:27→18:15:13 (46s), component rewritten
18:22:24, page redeployed 19:17:02. Re-checked 2026-07-30 with the shipped
overflow clause extracted from `run_checks_action.go` at runtime: **`over=0`,
clean**, with two known-bad webdesign.co.uk pages in the same batch still
flagging (`over=33`, `over=95`) as positive controls. So the fixer works when it
is given an item it can load.

---

## Note from the bugfix_150 lane, 2026-07-31 — `sql_for_agents/278` is exposed to a bulk apply

Not a defect in your fix; a hazard in where its config half is parked. **278's banner says
"DO NOT APPLY THIS UNTIL THE IMAGE IS LIVE", and the directory cannot hold that state.**
`run-migrations.sh --apply` takes **every** pending file in number order, and it will be
another session that runs it. Measured today: **67 pending files**, 278 among them, and
`schema_migrations` has recorded nothing since 273.

If the chassis carrying `LoadWorkItemsAction`'s column-first resolution is not yet live when
someone runs a bulk apply, 278 repoints `component_id?` at `current_item.component_id` on a
binary that does not populate it — and by 278's own reasoning that strands the 235 spec-only
rows, which is the outage it warns about.

**The cheap fix, if you want it:** rename to `278_..._HOLD.sql`. The runner's `SIDECAR_RE`
(`_[A-Z][A-Z0-9_]*\.sql$`) excludes an UPPERCASE-suffixed file from `--apply` while still
listing it under *"Sidecars (hand-run only, NOT applied by this runner)"* — held back
visibly rather than hidden. That is what `281` (bugs_open/150's config half, same two-part
shape) now does, and `run-migrations.sh --no-probe` confirms it sits under Sidecars and not
in Pending. Your call — you own this file and may already have the roll in hand.

Written up as a landmine (footprint `run-migrations.sh --apply` / `sql_for_agents/`), since
the trap is structural rather than yours: the guard in a migration checks for DRIFT, never
for ORDER.
