# HANDOFF — 2026-08-18, fresh chat starts here: 083's arc is DEMONSTRATED end to end; two defects in MY OWN 453 are owed; and the floor holds 10 rows nobody will ever be told about

**Supersedes `HANDOFF_2026-08-17c_continue_here.md`.** Measured 2026-08-18 ~13:10Z.
Another session worked this lane overnight (commits `03012d862`, `8d77196ad`, `6bfbe705b`) and
**corrected two things the 17c handoff asserted** — read §1 before trusting anything I wrote there.

## 1. What the overnight session established (read its commit bodies in full — they are the record)

- **Migration `458` applied 21:57Z + the scheduler Go half, BOTH now LIVE.** `kafka-scheduler`
  `v1.0.1308`, label `e7e5e4d53`, and `git merge-base --is-ancestor 03012d862 e7e5e4d53` → 0.
  Every CTE-only scheduled task now logs its own `pre_query` result.
- **⚠ MY 17c RESIDUAL REMEDY WAS INOPERATIVE, and they measured it.** I had recorded the guardian
  seat's suggestion verbatim — *"have the pre_query return a third column counting rows a door
  held"*. A third column would have been **write-only**: for `fire_message=false` the result was
  merged into inputData and discarded unlogged, so `promoted`/`pairs` had never been readable
  either. That is SCH-006's observability gap; 8 CTE-only tasks shared it. **Framework first, then
  the numbers** — the right order, and not what I proposed.
- **The promoter now says what it HELD.** Live tick, read from the running service:
  `{"held":"16","held_detail":"dead_fragment_link->page-build-handler (pair has never completed one …"}`
- **083's whole arc is DEMONSTRATED, and unlike criterion 2 it is discriminating.** The
  `page_component_status_drift → component-template-fixer` canary was hand-run 21:47:33Z, claimed
  64s later, component `approved→deployed` 21:49:05Z, item complete 21:49:16Z — **the artefact
  moved 11s BEFORE the row closed**. The pair became known-good and **the promoter took the rest
  unattended** on its next tick. Served page byte-identical before and after; demand control held
  (un-canaried siblings kept their `updated_at`). None of it was possible before 2026-08-15.
- **`bugs_open/300` filed, and it bites MY floor:** 1 of the 4 rows `453` escalated named a
  `page_components.id` with **zero** rows (page re-rendered 5 days after filing).
  `fixPageComponentStatus` hard-errors on `sql.ErrNoRows`, so canarying *that* row would have
  written a `failed` against the pair — and post-`444`/`454` failures are **scored**, so the run
  meant to QUALIFY a pair could instead push it **under my floor**. 016b already says do not key on
  `page_components.id`.
- **Council APPROVED round 1** on the 458 work (corr `8dc58e2a`), 4 advisories, architecture =
  `point_fix`.

## 2. ⚠ TWO DEFECTS IN MY OWN `453` — both named by that session as "its author's call". This is the top owed work.

### (a) The non-suppressing `WHERE` — visible in production right now

`453`'s final SELECT ends `WHERE (SELECT COUNT(*) FROM escalated) > 0`. **That does not suppress
anything.** The target list is aggregate-only with no `GROUP BY`, so Postgres returns exactly one
row regardless — `458` verified it read-only: the `WHERE` form returns `('0', NULL)`, the `HAVING`
form returns 0 rows. Live proof from the scheduler log this hour:

```
"task":"held-pair-canary-escalation","pre_query_result":"{\"escalated\":\"0\",\"pairs\":null}"
```

Two live consequences (`458`'s header states them): the scheduler cannot tell an acting tick from
an idle one (`dynamicData == nil`, main.go:200-216, is unreachable), and **the task claims its
concurrency slot on EVERY tick including idle ones** — `maintenance` is `max_concurrent=1` with
four other tasks in it, and `bugs_open/048`'s fix releases the slot on a no-op, which this defeats.
The overnight session measured the blast radius: **exactly 2 of 9 CTE-only tasks have it, and both
are this lane's.** They fixed the promoter and deliberately left `453` to me.

**Fix:** replace the `WHERE` with a `HAVING`, mirroring `458`. Note `458`'s subtlety — it suppresses
on `promoted = 0 AND held = 0`, not on `promoted` alone, *"because an idle tick that is holding rows
is the state this residual is about."* For `453` the analogue is: suppress only when there is
nothing escalated **and** nothing approaching the limit.

### (b) The escalation gap — `453` covers canary-held pairs ONLY, and the floor holds the rest for ever

`453` requires the pair to have **zero** successes. So a pair held by `444`/`454`'s **success floor**
— which has successes, just too few — is never escalated by anything. Measured today:

| pair | rows | oldest | successes | fails | fate |
|---|---|---|---|---|---|
| `literal_markdown → page-build-handler` | **10** | 08-17 | 3 | 24 | **FLOOR-held → `453` NEVER escalates** |
| `placeholder_contact → page-build-handler` | 3 | 08-16 | 0 | 4 | canary-held → escalates ~08-19 |
| `empty_internal_href → page-build-handler` | 1 | 08-18 | 0 | 1 | canary-held → escalates |
| `dead_fragment_link → page-build-handler` | 1 | 08-18 | 0 | 0 | canary-held → escalates |
| `missing_conversion_path → content-gap-planner` | 1 | 08-17 | 0 | 0 | canary-held → escalates |

**10 rows in about a day, growing, with no clock, no owner and no escalation path** — which is
exactly the disease `453` was built to cure, reproduced one category over. It is my design gap, not
a surprise: the consumer notice in `bugs_open/184` predicted findings would sit at `detected`, but
nothing tells anyone *when it has gone on too long*. Widen `453` to cover floor-held pairs too
(owner named per type; `literal_markdown`'s owner is the `bugs_open/201`/`184` lane).

## 3. ⚠ STILL OPEN from 17c — work-item history SHRINKS, cause UNDIAGNOSED

`required_fields_missing` completes read 64 (11:00Z) → 50 (18:30Z) on 08-17, not re-statused.
**A second instance today:** `literal_markdown → page-build-handler` was `1 complete / 28 failed`
yesterday and is `3 successes / 24 failed` now — successes rising is expected, **failures falling is
not.** Both of the promoter's tests read lifetime history, so a quiet pair can lose its evidence and
read as never having worked. `[UNDIAGNOSED]` — no `DELETE` in any scheduled task, no retention
window in the code, no CronJob naming the table, and the oldest surviving row fleet-wide is
2026-03-15. **Run `090` on it before designing any guard.** Full write-up: `bugs_open/083`, last section.

## 4. Owed work, in priority order
1. **`453` defect (a)** — the `HAVING` fix. Small, measured, and it is holding a shared concurrency
   slot every 24h. New numbered migration (`453` is ledger-recorded — never edit a recorded file);
   re-`ls` for the next free number, and expect collisions (`453`/`454` already collide with other
   sessions' files — **do NOT renumber those; the ledger keys on filename**).
2. **`453` defect (b)** — widen escalation to floor-held pairs. 10 rows are waiting on it.
3. **§3's `090` run** — the only thing that can quietly undo this lane's mechanism.
4. **`277` → `bugs_closed/`** ~**2026-08-22**: churn guard + the two cancelled conversions
   re-raising. **Both paths on the commit** — LANDMINE.
5. **`083` → `bugs_closed/`** — its arc is now demonstrated (§1); blocked only by §3.
6. **`router_engine` (RFC_030)** — phase 1 measurement DONE and in that lane's NOTES; phase 2 is a
   council design round on shape A vs B, **before building**. ⚠ its PLAN's guarantee 8 is STALE
   (RFC_022 is CLOSED, counter live since 08-13, owner-ruled N=10, daily CronJob) and that bears
   directly on the A-vs-B choice — fix the PLAN before submitting.
7. **Council gate's config blind spot** — `097` scopes on `platform/`/`internal/`/`pkg/`, so a
   config-only mechanism cannot be submitted; another lane filed a LANDMINE (`landmine(297)`).

## 5. Landmines this lane hit
- **`status` has TWO terminal success states** (`complete`, `verified`) — `GROUP BY status` first.
- **`failed` rows carry NO `completed_at`** — pair-health keyed on it returns a uniform 100%.
- **The row set itself is not stable** (§3).
- **All three are ONE class:** the population measured was not the population assumed. None was
  *unmeasured* — each was **measured against an incomplete definition**, which no marker and no
  council round detects (12 seats approved `444`).
- **An aggregate-only SELECT with a `WHERE` returns one row regardless** — use `HAVING` (§2a).
- **A verify block's control can be a tautology** — `453`'s draft had `EXISTS(X) AND NOT EXISTS(X)`.
- **A same-tag rebuild ships the cached image**; negative control must be a real-but-different sha.
- **Check whether YOUR OWN guard strands YOUR OWN lane** — and §2b is the case where mine did.

## 6. Session-start checklist
`git log --oneline -10` · re-read this file from disk · **verify the chassis/scheduler revision
before trusting any Go claim** · `scripts/who-owns.py 277` / `083` (by SLUG — 083 is ambiguous) ·
re-measure §2b's table · then §4 item 1.
