# Pending: instance-scope acceptance on the two locked evidence-timeseries rows

**Status as of 2026-08-25: WRITTEN, UNRUN.** Both `page_components` rows are still
`lock_type='permanent'` and unconverted, and no new dispatch items exist.

These live here, in the repo, **because a session scratchpad dies with its session**
and the 2026-08-25 handoff would otherwise point at a directory that no longer
exists. They are **operational one-offs against two named rows**, deliberately NOT
in `docs/agent_docs/sql_for_agents/` — putting them there would expose them to the
migration runner, which takes every pending file in the directory it is pointed at.

## Why

The 283 / RFC_034 lane converted the shared `evidence-timeseries` template on
2026-08-23 so its section id comes from `{{.InstanceID}}` rather than
`{{.ComponentID}}`. Our two locked instances refused delivery and filed
`lock_blocked_change` items. **Owner ruled ACCEPT, 2026-08-25.** The lane had
recommended honouring (the id is inert); the owner chose fleet-wide consistency.

Full evidence trail: `../HANDOFF_2026-08-24_continue_here.md` §8.
The dry-run recipe that gated this: `../RUNBOOK_news_editorial_features.md` §11.

## Run order — and it is an order, not a list

```bash
PSQL="kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db"

$PSQL -f - < A_unlock_and_dispatch.sql   # unlocks BOTH rows + re-dispatches the delivery
./verify.sh                              # re-run until it prints ALL PASS
$PSQL -f - < B_relock.sql                # restore permanent/news_editorial_features-lane
./verify.sh                              # confirm the lock came back
$PSQL -f - < C_close_lock_items.sql      # both items -> complete
```

⚠ **Between A and B the two rows are UNLOCKED** and exposed to every sweep on the
estate. This lane was hit once already by an improvement-loop misfire (2026-08-22).
**Do not start the sequence if you cannot finish it.**

⚠ **`verify.sh` checks the SERVED page, and that is the point.** A `complete` work
item and correct stored bytes are both consistent with the page still serving the
old version — measured by the 283 lane 2026-08-24: three repairs complete, stored
bytes correct, one page served the old version for hours. If steps 1 and 2 pass and
step 3 still shows the old id, that is a **publish lag, not a failed edit**: wait and
re-check. **Do NOT re-dispatch** — a second delivery on top of a pending publish
races two writes at an unlocked flagship row.

Every script is idempotent, guarded, and `RETURNING`s what it did, with a raw
read-back — so `UPDATE 0` can be told apart from "already applied".

## Expected result

| page | id before | id after |
|---|---|---|
| robot-demand-step-change | `evidence-timeseries-ifr` | `c-evidence-timeseries` |
| darts-calendar-density | `evidence-timeseries-pdc-calendar` | `c-evidence-timeseries` |

Served-byte baseline 2026-08-24: rh 94,351 / do 92,883. Predicted after: rh 94,349 /
do 92,872 — **`[PREDICTED]` from the dry run, not measured.** `verify.sh` prints the
delta so it can be checked rather than assumed.

## When it lands

Tell **`bugs_open/283`** the served ids — they are holding to move
`evidence-timeseries` 3 → 1 in RFC_032 §9a rather than re-deriving the census.

**Write the close-out as CONSISTENCY, not repair.** Our pages carry one ev-ts each,
so they were never in the population where a literal id can actually bite (the 283
lane's own narrowing: 48 unconverted → 8 multi-instance → 1 duplicated → 0 reaching
a visitor). A close-out implying a repair makes the next reader think a defect was
there.

## Then delete this directory

It is a pending-action folder, not a record. The record is 08-24 §8 and the NOTES
entry. Once C has run and the ids are verified served, `git rm -r` this directory in
the same commit that records the result.
