# HANDOFF — 2026-08-18b (evening), fresh chat starts here: the undiagnosed risk is SOLVED and it was worse than filed; 465 + 466 applied; all four of my own defects closed

**Supersedes `HANDOFF_2026-08-18_continue_here.md`** (same day, earlier). Measured 2026-08-18
15:45–17:00Z. Read from disk, then `NOTES_required_fields_repair.md` from the bottom.

## 0. Build state

`agent-chassis` **v1.0.1309** verified shipped: label `f0117fb8b`, PRESENT in the binary,
superseded `a6d1c53c0` ABSENT, 28 commits behind HEAD (ordinary churn). A real tag bump, so no
repeat of 08-17's cached-layer no-op. `kafka-scheduler` **v1.0.1308** carries `458`'s Go half, so
every CTE-only task now logs its own `pre_query` result — that log is the instrument for everything
below.

## 1. THE UNDIAGNOSED RISK IS SOLVED — and it was already causing damage

`work-item-archiver`: **enabled**, daily, description *"Archives terminal work items older than 7
days to `site_work_items_archive`"*. The archive holds **20,184 rows against 8,702 live**. So both
of the promoter's success tests — reading `site_work_items` only — actually meant **"in the last 7
days"**, and `bugs_open/083`'s disease was being reintroduced by the mechanism built to cure it.

**Live victims, measured:**

| pair | live-only | TRUE (with archive) |
|---|---|---|
| `empty_internal_href → page-build-handler` | 0/1 = 0% — **held as "never completed"** | **9/5 = 64%** |
| `empty_section → page-build-handler` | 12/16 = 43% | **316/33 = 91%** |
| `literal_markdown → page-build-handler` | 3/24 = 11% | 3/36 = 8% (correctly held) |
| `placeholder_contact → page-build-handler` | 0/4 | 0/6 (genuinely never succeeded) |

**`465` (applied, ledger-recorded, `_ROLLBACK.sql`)** makes both tests read
`site_work_items UNION ALL site_work_items_archive`. **PROVEN end to end**: the stranded
`empty_internal_href` row was promoted on the first tick after apply (16:28:23) and claimed 24s
later. Cost 78 ms → 134.6 ms per 900 s tick.

> A shared VIEW over both tables is the tidier estate-wide answer and is **deliberately not taken** —
> a new shared object other pipelines may adopt is a shared-seam change (owner ruling 2026-07-28).
> That is the right RFC for whoever wants it; named so the omission reads as a decision.

## 2. `466` — three corrections to my own `453`, all applied

- **(a)** the final `WHERE` did not suppress (aggregate-only target list, no `GROUP BY`, returns one
  row regardless), so the task claimed its `max_concurrent=1` `maintenance` slot on every idle tick
  and defeated `bugs_open/048`'s no-op release. Now `HAVING`. Found by `458`'s author, left to me.
- **(b)** floor-held pairs were escalated by **nothing** — `literal_markdown`, 10 rows and growing,
  no clock, no owner, no path out. Now escalated too, with **opposite** remedy text: a canary on a
  floor-held pair adds a failure and pushes it *further* under, which `bugs_open/300` is the live
  case of, so the text says explicitly **do not canary — fix the handler**.
- **(c)** the one I had not spotted, and the worst: `453`'s held test read `status='complete'` over
  the live table alone, so after `454` and `465` it **disagreed with the promoter** — it saw **0**
  successes for `empty_internal_href` where the promoter saw **9**, and would have asked a human to
  canary a pair the promoter promotes unattended. Both are now one test.

Live after apply: **watching 15 rows — 5 canary-held, 10 floor-held.**

## 3. Owed work, in priority order

1. **Watch `466`'s first tick** (daily; last ran 12:57 under `453`, so the next is the first under
   `466`). Expect: `literal_markdown` (10 rows, floor-held, oldest 08-17) to escalate once past the
   3-day limit ~**08-20**; `placeholder_contact` (3 rows, canary-held, oldest 08-16) ~**08-19**. The
   `pre_query_result` line in `kafka-scheduler` logs is the instrument — it should now carry
   `watching` and `watching_detail` even on an idle tick, which is `466`(a) working.
2. **`277` → `bugs_closed/`** ~**2026-08-22**: churn guard + the two cancelled conversions
   re-raising. **Both paths on the commit** — LANDMINE.
3. **`083` → `bugs_closed/`** — its arc is demonstrated (the 08-17 canary, and now `465` proven);
   the risk that blocked it is solved. Check nothing else is outstanding, then move it.
4. **`router_engine` (RFC_030)** — phase 1 measurement DONE and in that lane's NOTES; phase 2 is a
   council design round on shape A vs B, **before building**. ⚠ its PLAN's guarantee 8 is STALE
   (RFC_022 is CLOSED; counter live since 08-13, owner-ruled N=10, daily CronJob) and that bears
   directly on the A-vs-B choice — fix the PLAN before submitting. **Now the largest real work here.**
5. **Council gate's config blind spot** — `097` scopes on `platform/`/`internal/`/`pkg/`, so a
   config-only mechanism cannot be submitted; another lane filed `landmine(297)`. Every migration in
   this lane since `444` has been unreviewable for that reason.
6. **Consider submitting `465`+`466` to the council** — they are config-only, so §3.5 applies and it
   needs `FORCE=1` or a Go anchor. Worth it: `465` changes what "lifetime" means for a shared gate.

## 4. Landmines this lane hit — the family is now FIVE and the shape never varied

Every one is **a population or a domain assumed rather than enumerated**, and none was caught by
review (twelve seats approved `444`):

1. **`failed` rows carry NO `completed_at`** — pair-health keyed on it returns a uniform 100%.
2. **`status` has TWO terminal success states** (`complete`, `verified`) — `GROUP BY status` first.
3. **The row set is not stable** — rows leave `site_work_items`.
4. **The row set is only a ~7-DAY WINDOW** — `work-item-archiver`; the archive is *bigger* than the
   live table. ⚠ **the three searches that will NOT find it**: no `DELETE` (it *moves*), a **NULL
   `pre_query`** (it is `fire_message=true`, so its SQL is in an agent, not the column you grep), and
   a description saying *"Archives"* not *"retention"*. ⚠ **and the control that cannot come out
   otherwise**: "the oldest surviving row is 2026-03-15" — that row is **non-terminal** and the
   archiver only takes terminal ones. Enumerate the ACTORS
   (`SELECT name, description, fire_message FROM scheduled_tasks`) before grepping for the verb.
5. **A control that cannot come out otherwise** — **THREE** tautological ones caught in this lane
   (`430`'s, `453`'s `EXISTS(X) AND NOT EXISTS(X)`, `466`'s `promotable AND NOT promotable`). The
   test is not *"is this control true?"* but ***"could it ever have come out non-zero?"***

Also live: **a same-tag rebuild ships the cached image** (negative control must be a real-but-
different sha, never a zeros run) · **an aggregate-only SELECT with a `WHERE` returns one row
regardless** — use `HAVING` · **check whether your own guard strands your own lane** · **a pathspec
commit still takes a same-file passenger** · **backticks in `git commit -m` EXECUTE** — they ate two
phrases from `466`'s message, one of them the substance; restated in NOTES since forward-only forbids
an amend. Use single quotes or `-F`.

## 5. Migration-number collisions — do NOT renumber
`453` and `454` each exist twice on disk and in the ledger (mine plus another session's). Nothing is
lost: `schema_migrations`' PK is `filename`, all are applied. **Renaming orphans the ledger row, the
file reads as pending, and the runner re-applies it** — and `454`'s positive control would now fail.

## 6. Session-start checklist
`git log --oneline -10` · re-read this file from disk · **verify the chassis/scheduler revision
before trusting any Go claim** · `scripts/who-owns.py 277` / `083` (by SLUG — 083 is ambiguous) ·
re-measure §2's watched set · then §3 item 1.
