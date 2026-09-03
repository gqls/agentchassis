# 459 — the spend governor's level-change ALARM never fires: `FOR UPDATE` in the `old` CTE races the `UPDATE` of the same row in the same statement

**Filed** 2026-09-03 by the dispatch_throughput lane, from the first induced shed of the D4
governor (owner-authorised). **Status: OPEN — reproduced, root-caused, fix not written.**
**Severity: the governor sheds SILENTLY.** Shedding itself is proven working; what is dead is
the only mechanism that announces it.

## The symptom, measured

The governor was enabled 2026-09-03 10:14:32Z. At 11:14–11:33Z an induced window drove two
real level changes — **0 → 2 at ~11:17:07Z and 2 → 0 at ~11:33:20Z**, both visible in
`governor_state` and both accompanied by 114–115 correctly-classed rows in
`governor_withheld_now`. `doc_notes` gained **zero** rows for either:

```sql
SELECT created_at, left(body,150) FROM doc_notes WHERE subject_key='spend-governor'
ORDER BY created_at DESC LIMIT 10;
-- one row, 2026-09-02 14:03:13Z, the stage-B design record. Nothing from today.
```

## Root cause, with the induced-failure A/B

The task `spend-governor-state` runs one statement. Two of its CTEs touch `governor_state`:
`old AS (SELECT shed_level FROM governor_state WHERE id = 1 FOR UPDATE)` and
`upd AS (UPDATE governor_state s SET ... WHERE s.id = 1 ...)`. The note is
`noted AS (INSERT INTO doc_notes ... FROM old, new WHERE old.shed_level <> new.lvl)`.

Postgres runs a statement's sub-statements against one snapshot and does not support one
statement both row-locking and updating the same row this way. **With `FOR UPDATE` present,
`noted` selects zero rows even when the level genuinely changed.**

A/B on the LIVE stored text, both arms inside `BEGIN … ROLLBACK`, with the stored level forced
to 3 while the computed level is 0 `[MEASURED 2026-09-03 ~11:37Z]`:

| arm | statement | `level_changed` | note row |
|---|---|---|---|
| control | the `old` CTE alone (no `upd` in the statement) | — | `old` returns 1 row, `shed_level=3`, `new.lvl=0` — so old and new DO differ |
| **A** | the live text, verbatim | **0** | none |
| **B** | the live text with `FOR UPDATE` deleted — one token, nothing else | **1** | note lands |

Post-rollback control clean both times (`shed_level` 0, `spend-governor` note count unchanged
at 1). The control arm matters: it rules out "old and new did not actually differ", which is
the other way to get a zero.

## How it got here — and why the test that proved it did not catch it

- **672** installed the task AND proved the alarm: its verify drives the level 0→3→0 with a
  synthetic budget and asserts `after_notes - before_notes = 2`, then deletes its own two rows.
  It passed. The alarm worked when 672 applied.
- **673** then added `FOR UPDATE` to the `old` read, to close a real hole (a fire that BLOCKS
  on the advisory lock keeps its pre-block snapshot, so `old` could read a stale `shed_level`).
  Good intent. **Its verify asserts only that (a) the token is present in the stored text,
  (b) the text still `EXECUTE`s, (c) the level is 0 on a NULL budget.** It never drives a level
  change, so it cannot observe the note. `grep -c 'FOR UPDATE' 672…` = 0; `673…` = 8.
- So a hardening migration silently voided the assertion that had proved the alarm, and
  nothing re-ran it. **This is the transferable pattern** (016b §9 candidate): *when you change
  a statement, the verify that must be re-run is the one belonging to the BEHAVIOUR you might
  break, not the one belonging to the token you edited.* 673's verify tests its own change
  perfectly and is blind by construction to the only thing that change could break.

## What is NOT broken (measured the same hour, so nobody over-scopes this)

- Shedding works end to end: at L2, 3 dispatch loops handled 24 items, **all of them
  llm-free**, while 114 llm-bearing items sat withheld and eligible. The Go claim backstop
  fired once with `spend_governor_shed`.
- `governor_withheld_now` populates correctly and is the honest "paused vs stuck" answer.
- `governor_state.computed_at` (the heartbeat) advances normally.
- The state task's level arithmetic is correct in both directions.

## Fix candidates, ordered by what closes the door

1. **Restructure so the note reads the old value from the UPDATE's own output** — capture the
   pre-update level in the `upd` CTE's `FROM` list and `RETURNING` it, then have `noted` read
   `FROM upd` (a data-modifying CTE chain, which orders them). This removes the same-row
   `FOR UPDATE`+`UPDATE` pair entirely rather than trading one hazard for the other. Needs care:
   it must preserve 673's snapshot property, which is the whole reason `FOR UPDATE` is there.
2. **Split the note out of the statement** — a trigger on `governor_state` firing on
   `shed_level` change. Makes the alarm structurally independent of the task's statement shape,
   so no future hardening of that statement can silence it again. Costs a trigger on a hot row.
3. **Delete `FOR UPDATE`** — one token, restores the alarm today, and reopens exactly the hole
   673 was written to close. Only acceptable as a stop-gap with 673's hazard re-recorded.

Whichever is chosen, the verify must **drive a real level change and assert the note**, i.e.
re-run 672's assertion, not 673's. An appliable migration is council scope (CLAUDE.md).

## How to verify a fix

```sql
-- Must print 1 and land a row. Run inside BEGIN … ROLLBACK against the LIVE stored text.
BEGIN;
UPDATE governor_state SET shed_level = 3 WHERE id=1;      -- force old <> new
-- EXECUTE the stored pre_query verbatim, then:
SELECT count(*) FROM doc_notes WHERE subject_key='spend-governor'
  AND created_at > now() - interval '2 minutes';
ROLLBACK;
```
Full runnable A/B: `docs/agent_docs/docs024_key_docs_latest/dispatch_throughput/RUNBOOK_dispatch_throughput.md`
§"Spend governor (D4)".

## Substituted verification (owner ruling 2026-07-31)

Not put through the `090` diagnosis loop. Stated plainly, as the ruling requires: the cause is
not cross-cutting and is fully self-evidencing — a single statement, an A/B differing by one
token, run against the live stored text, with a control arm that rules out the competing
explanation and a post-rollback control proving nothing persisted. A diagnosis run would read
the same statement I read. **What it would add, and what I therefore did not get, is an
independent reading of whether candidate 1 preserves 673's snapshot property** — that question
is open and belongs with whoever writes the fix.

## Provenance

- Found by: the induced-shed window, dispatch_throughput lane, 2026-09-03 (NOTES 2026-09-03).
- Register: **AGOV-013**. Migrations `671` (task), `672` (alarm + its proof), `673` (the token).
- Related: `bugs_open/398_HANDOFF_2026-08-25_scheduled_tasks_row_is_not_single_flight.md`
  (same table, different defect — resolve 398 by SLUG, the number is ambiguous).
