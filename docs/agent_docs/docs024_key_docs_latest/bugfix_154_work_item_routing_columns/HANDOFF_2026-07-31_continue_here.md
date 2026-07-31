# HANDOFF — `bugs_open/154`, continue here

**Written 2026-07-31 ~22:05 by session "bugfix 19" (`9de5c96a`).**
Cold-start doc: read this first, then `NOTES_…` for the evidence trail.

## One-line state

**Both halves of the fix are LIVE and verified. `154` is ONE observation away
from closing** — a single dispatched `tool-auditor` item clearing `load_tool`.
The item is reset, queued and correctly ranked (2 of 36, inside the loop's window).

**That observation is BLOCKED, and not by anything in this bug:** nothing has been
claimed anywhere on the fleet for ~90 minutes, while `build-pipeline-trigger`
fires every 120s and completes and no site is blocked. Almost certainly
`bugs_open/029`, which a session is already working. See "Why it has not fired"
below before drawing any conclusion from the item still sitting at `triaged` —
it is evidence about dispatch, not about this fix.

## What the bug was (30 seconds)

`site_work_items` carries `component_id`, `entity_id`, `affected_url` as **real
columns**. `LoadWorkItemsAction` built `current_item` from a SELECT that listed
only `page_id` among them, so the only path a dispatcher could reference was
`current_item.spec.<key>` — a copy each creating agent had to duplicate into the
`spec` JSONB. `tool-auditor` populated the column and not the blob ⇒ its items
were structurally undispatchable: the optional `"component_id?"` mapping resolved
nothing, `ResolveInputMapping` silently skips an unresolved optional path, and
`tool-improver`'s `load_tool` hard-errored on the nil param.

**The creator that used the schema properly was the only one whose items could
not be dispatched.**

## What is DONE (all verified, not asserted)

| thing | state | evidence |
|---|---|---|
| Go fix (`setRoutingField`, column-first + `spec` fallback) | **LIVE** | pod-grep **both** replicas on `v1.0.1219`: new symbol `1`, positive control `1` |
| Config (`build-dispatch-loop` `component_id?` → `current_item.component_id`) | **LIVE** | `sql_for_agents/278` applied: pre-flight count 1, `snapshot_agent` → `099b51e0-6dd0-4856-8f82-805a379e8b1d`, `UPDATE 1`, verified reads column path |
| Regression test | committed | `load_work_items_routing_fields_test.go`, induced-fault proven (fails on the bug AND on a regression of the 235-row majority) |
| Diagnosis loop | **CONFIRMED** | `21758756-d7b3-444a-844e-b37e09b5c9ce`, first iteration |
| Council gate | **APPROVED** | `10be5ed9-3bd0-45ed-b6bb-4385a887967d`, 8 seats, 6 advisory, none high |
| Concept register | WDS-014 | `docs026_concept_register/register/work-dispatch.md` + index |
| Landmine | appended + synced | `LANDMINES.md`, `landmines-sync.py --check` → in sync |
| Wrong calls | 2 logged | `WRONG_CALLS.md` |

Commits: `4667db235`, `9bde2aa14`, `c92a65dad`, `66e91f27e` (+ the docs commit
that carries this file).

## THE ONE REMAINING STEP

Item **`ee745694`** (`improve_tool`, `tool-bayesian-ranking`,
`gamesdesign.co.uk`) was reset to `triaged`/`attempt_count=0` at ~21:53 with the
owner's explicit go-ahead. Watch it clear `load_tool`:

```sql
SELECT status, claimed_by, left(error,150) AS error
FROM site_work_items WHERE id::text LIKE 'ee745694%';
```

- **PASS** = it reaches `complete` (or fails at ANY step *after* `load_tool`).
  The defect was `step load_tool failed: … input_data.component_id resolved to
  nil`. **Anything past `load_tool` proves the fix**, because that error was the
  first thing the workflow did.
- **FAIL** = the same `load_tool … resolved to nil` message. That would mean the
  column value is still not reaching the handler — re-read `278` and confirm the
  mapping actually took.

Capture the evidence **when it happens**: `orchestration_states` history is
short (rows for an 18:14 run were gone by ~18:40 — the retention trap `154`
itself warns about). `site_work_items` and `diagnosis_artifacts` outlive it.

Pre-state for change-detection on the component the fixer will rewrite:
`c345a76a` `tool-bayesian-ranking`, `md5=5de92eba982c30315b3886096b52dd87`,
9158 bytes, `updated_at=2026-07-29 18:48:23Z`.

### Why it has not fired — it is NOT this fix, and NOT ordinary slowness

> **CORRECTED 22:10, same session.** I first wrote that the dispatch lane was
> "alive, slow" and the item was waiting its turn. **That reading was already
> stale as I wrote it, and re-measuring inverted it.** The 28-minute gap I cited
> was not a lull — it was the *start* of a stall that is still going.

| check | result |
|---|---|
| item reachable? | **rank 2 of 36** dispatchable, inside the loop's `max_items: 5` — fine |
| site blocked? | `gamesdesign.co.uk` unlocked, `deployed`, **0** `claimed` items ⇒ not the `NOT EXISTS` whole-site blocker |
| stuck claims anywhere? | **0 rows fleet-wide** — no site is excluded by a stale claim |
| `build-pipeline-trigger` alive? | **enabled, fired 1 min ago, completed** — every 120s, healthy |
| anything claimed recently? | **NOTHING fleet-wide for ~90 min.** Last: `dartsonline.com` ×9 at 20:33Z, `finetuning.uk` ×1 at 20:04Z, then silence |

**The trigger fires on schedule and completes, no site is blocked, work is queued
and eligible — and nothing is being claimed anywhere.** That is a dispatch defect
and it is **not** in `154`'s scope: my change is in `LoadWorkItemsAction`, which
runs *after* a site is selected and cannot stop a selection happening. It cannot
produce a fleet-wide claim drought, and the drought **pre-dates the config apply**
(last claim 20:33Z, migration applied ~21:45Z).

**Do NOT diagnose it here — it is very likely `bugs_open/029`
(`hung_spawns_saturate_dispatch_group_and_halt_builds_fleetwide`), that shape
exactly, and a session is already on dispatch** (untracked in the tree as I write:
`bugfix_029_dispatch_gate/PLAN_2026-07-26_dispatch_gate.md`,
`sql_for_agents/213_dispatch_gate_matches_dispatcher.sql`,
`214_build_dispatch_watchdog.sql`). Hand them the measurement above; do not
compete with a second diagnosis.

**Consequence for `154`:** the last verification step is blocked on a **separate,
already-owned** bug, not on anything unresolved here. The item stays queued and
correctly ranked and should go on its own once dispatch recovers. Two ways to
finish:

1. **Wait** — re-run the status query above once claims are moving again.
2. **Judge it sufficient without the live dispatch.** Defensible on what is
   already collected (coalesce in both binaries, mapping reads the column path,
   induced-fault test over the exact column-only shape, all four items' ids
   satisfying every clause of `load_tool`'s query) — but if you take this route,
   **say in the close that no live dispatch was observed.** An unwitnessed step
   recorded as witnessed is precisely what these bug files exist to prevent.

**Use `GROUP BY claimed_by`** for dispatch liveness — a bare `max(claimed_at)`
reports another lane's diagnosis run and reads healthy (`WRONG_CALLS.md`, same
day) — and read the *absolute last claim time*, not the gap since one arbitrary
reading, which is the error corrected above.

## Then close it

`154`'s bar is *fixed AND live* — both halves are live, so once the dispatch is
observed, move it:

```bash
git mv bugs_open/154_HANDOFF_2026-07-30_tool_improver_fails_at_load_tool_when_the_item_came_from_tool_auditor.md bugs_closed/
git add bugs_closed/154_*.md
git commit bugs_open/154_HANDOFF_2026-07-30_tool_improver_fails_at_load_tool_when_the_item_came_from_tool_auditor.md \
           bugs_closed/154_HANDOFF_2026-07-30_tool_improver_fails_at_load_tool_when_the_item_came_from_tool_auditor.md \
           -m "close(154): ..."
```

**NAME BOTH PATHS ON THE COMMIT.** `git mv` stages a delete *and* an add; a
pathspec naming only the new path ships a **copy**, leaving the file in
`bugs_open/` as well — so the bug reads as still open, which is the one thing
closing it was for. Verify at HEAD, not on disk:

```bash
git ls-tree -r --name-only HEAD -- bugs_open/ bugs_closed/ | grep 154   # expect ONE line
```

(That is a `LANDMINES.md` entry from the `135` lane, 2026-07-31 — it bit two
closes that day.)

Also update `bugs_closed/README.md` if the index convention there expects a row.

## What is deliberately NOT fixed (do not "finish" these by accident)

1. **`entity_id` / `affected_url` are NOT closed** — narrowed after the council's
   `bug_historian` seat. The Go half exposes three columns; the migration rewired
   **one** mapping. Measured: both have **0** rows with the column set;
   `affected_url` is read by nothing; `entity_id` by exactly one agent
   (`asset-deployer`, via `input_data.spec.entity_id` — the **`spec` passthrough**,
   not a dispatcher mapping, so the coalesce cannot reach it). The first creator
   to write `entity_id` on the column hits this identical bug, and fixing it then
   needs **two** edits (add an `entity_id?` mapping AND repoint `asset-deployer`).
   Not pre-fixed because there is no failing population — that would be the same
   speculative widening refused for `page_id`.
2. **`page_id` stays column-only.** 218 rows have a NULL column and a
   `spec.page_id`; extending the fallback for symmetry would newly change what
   reaches those handlers.
3. **Never backfill the resolved value into `spec`.** It needs no config change
   and is therefore tempting. `create_rerender_items_action.go:219` gates
   `scoped := (reason=="section_data_resolved"||reason=="image_landed") && componentIDStr != ""`
   on it — a write there can flip a site-wide rerender into a component-scoped
   one. `TestSetRoutingField_NeverMutatesSpec` fails if someone tidies it in.
4. **`tool-auditor`'s site-scoped `item_key`** (`audit_fix_<domain>` — one key per
   *site* on a per-*tool* fix, so rows accumulate and cannot say which tool they
   mean). This is `154`'s own second finding, a defect in item **creation** with
   fleet-wide dedup consequences. Left open on the ticket; wants its own
   measurement and review.
5. **The remaining 3 stuck items** (`a5d11c86` robot-hands, `7c2d898a`
   gamesdesign) are still `failed` at 3/3. Once the fix is proven they can be
   reset the same way — their attempts were spent on this bug. **`5b4fd5cc` is
   `wont_fix` — a human decision. Do not resurrect it.**

## Open question the council raised (not mine to answer)

`architecture` seat, medium, explicitly not a block: the three new top-level keys
are a shared wire-shape change to `current_item`, and **no RFC describes that map
as a contract** — its key surface and resolution semantics are undocumented,
which is why a change like this has nowhere natural to be reviewed. Recorded as
WDS-014's open review question.
