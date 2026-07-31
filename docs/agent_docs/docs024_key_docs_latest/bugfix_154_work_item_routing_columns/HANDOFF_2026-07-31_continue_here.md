# HANDOFF — `bugs_open/154`, continue here

**Written 2026-07-31 ~22:05 by session "bugfix 19" (`9de5c96a`).**
Cold-start doc: read this first, then `NOTES_…` for the evidence trail.

## One-line state

**Both halves of the fix are LIVE and verified. `154` is ONE observation away
from closing** — a single dispatched `tool-auditor` item clearing `load_tool`.
The item is reset, queued and correctly ranked (2 of 36, inside the loop's window).

**It has not dispatched yet, and that is not evidence about the fix.** Fleet
dispatch runs in bursts with long gaps between them (measured: a ~90-minute quiet
spell earlier today, then 19 claims in half an hour). The item is queued behind
`find_dispatchable_site` picking this site — one arbitrary site per tick. See
"Why it has not fired" below, **including the two wrong readings I published
before getting it right** — the second of which wrongly accused `bugs_open/029`.

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

### Why it has not fired — dispatch is INTERMITTENT, and it is not this fix

> **CORRECTED TWICE. Read the third reading; the first two are kept because the
> way they failed is the useful part.**
>
> 1. ~~"Dispatch is alive but slow, the item is waiting its turn."~~
> 2. ~~"Dispatch is DEAD — nothing claimed fleet-wide for ~90 minutes, almost
>    certainly `bugs_open/029`."~~ **This was wrong, and wrong for an
>    embarrassing reason: I compared my shell's LOCAL clock (22:05 BST) against a
>    UTC timestamp (20:33Z) and turned a 31-minute gap into 90 — the exact
>    `+01:00` vs `Z` trap written down three paragraphs up in this very file.**
> 3. **Dispatch is intermittent/bursty. Reading 1 was substantially right.**

Every interval below is computed **by the database** (`now() - claimed_at`), never
by comparing a shell clock to a DB timestamp. That is the fix for how this went
wrong twice:

| claims, by half-hour (UTC) | count |
|---|---|
| 18:00–18:30 | 2 |
| **18:30–20:00** | **none — a ~90 min quiet spell** |
| 20:00–20:30 | 15 |
| 20:30–21:00 | 4 (last at 20:33:49Z) |
| 21:00– | none so far |

`mins_since_last_claim_anywhere` = **39** at 21:12Z.

So the fleet dispatches in **bursts separated by long gaps**, and a 39-minute gap
is *within the range already observed today* — there was a longer one earlier,
followed by 19 claims in half an hour. **A quiet lane is not a dead lane here.**

Everything else checks out and is unchanged: `build-pipeline-trigger` enabled,
firing every 120s, completing; **0** stuck `claimed` rows fleet-wide; target site
unlocked, `deployed`, 0 claims; item **rank 2 of 36**, inside `max_items: 5`.

**Conclusion: the item is simply queued and waiting for `find_dispatchable_site`
to pick this site** — one arbitrary site per tick (WDS-002). Nothing here is
evidence about the fix in either direction, and **there is no reason to route work
at `bugs_open/029` on the strength of this** (my previous version of this section
said there was — disregard it).

**How to check liveness without repeating my mistake:**

```sql
-- let the DB do the arithmetic; never subtract a shell clock from a DB timestamp
SELECT claimed_by,
       max(claimed_at) AS last_claim,
       round(extract(epoch from now()-max(claimed_at))/60) AS mins_ago
FROM site_work_items WHERE claimed_at > now() - interval '6 hours'
GROUP BY claimed_by ORDER BY last_claim DESC;
```

Two traps, one on each axis, and I hit both in sequence: **population** — a bare
`max(claimed_at)` reports another lane's diagnosis run and reads healthy (hence
`GROUP BY claimed_by`); **time anchor** — mixing local and UTC clocks. Fixing one
axis does not make the other sound.

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
