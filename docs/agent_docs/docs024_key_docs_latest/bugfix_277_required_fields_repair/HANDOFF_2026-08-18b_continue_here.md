# HANDOFF — 2026-08-18b (evening), fresh chat starts here: the undiagnosed risk is SOLVED and it was worse than filed; 465 + 466 applied; all four of my own defects closed

**Supersedes `HANDOFF_2026-08-18_continue_here.md`** (same day, earlier). Measured 2026-08-18
15:45–17:00Z. Read from disk, then `NOTES_required_fields_repair.md` from the bottom.

## 0. Build state

> **UPDATED 2026-08-18 18:20Z:** the fleet has rolled again since. `kafka-scheduler` now reports its
> own `build provenance` as **`0b185bad2a49c6e032352fa9e7d0b429f0a95104`** (pod 17m old at the time).
> `agent-chassis` provenance was **absent from `--tail=3000`** — that is the documented "startup line
> has scrolled" case for a busy service, **not** "unstamped"; use the binary probe if you need it.
> Nothing below depends on the Go half changing: `471`/`472` are config-only.

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
`site_work_items UNION ALL site_work_items_archive`. Cost 78 ms → 134.6 ms per 900 s tick.

**PROVEN for what it claims, and NO MORE — stated precisely because the two are easy to conflate.**
The stranded `empty_internal_href` row was promoted on the first tick after apply (16:28:23) and
claimed 24s later, so *the promoter no longer holds a pair with nine lifetime successes*: that is
`465`'s whole claim and it is demonstrated. **The dispatch then FAILED** (attempt 1 of 3). An earlier
draft of this handoff said "proven end to end", which overstated it.

**The failure is a GUARD WORKING, not a defect** — and it exposes something about my floor.
`vonc.com/tools/archetype-taster-quiz/index.html`:
`step save_sections failed: page … is rebuild_policy=owned (tool/widget-owned): a generic section save
would clobber it … Refusing to overwrite.` The handler correctly declined to overwrite a tool-owned
page. That is `bugs_open/295`'s family (*"SIX producer families die on owned pages"*).

> ### ⚠ NEW FINDING — the success floor counts PROTECTIVE REFUSALS as handler failure
>
> `444`/`454`/`465`'s floor counts `status='failed'` regardless of *why*. A handler that correctly
> refuses to clobber an owned page, and a handler that tried and produced nothing, score identically.
> [MEASURED 2026-08-18] on the floor-held pair `literal_markdown → page-build-handler`, of **24** live
> failures: **8 are owned-page refusals** (`rebuild_policy=owned`), **2 are transient delivery
> failures** (`failed_transient`, message validation), and the rest are genuine non-repairs
> (`completion blocked: post-fix verification found the defect still present` — `201`'s verifier
> doing its job).
>
> **The verdict on this pair SURVIVES the correction** — excluding refusals and transients it is
> still roughly 3/(3+14) ≈ 18%, under the 25% floor — so nothing is currently mis-held and this is
> not urgent. But the design is wrong in general: for a pair whose failures are mostly protective,
> the floor would hold a handler that is behaving correctly. The floor should distinguish
> *"refused on purpose"* and *"transient"* from *"tried and failed"*. Not fixed here; it wants its
> own measurement of how the error strings partition fleet-wide, and it touches `bugs_open/295`.
>
> **CORRECTED 2026-08-18 18:40Z — the figures in this block are LIVE-TABLE ONLY and are superseded.**
> The pair's true lifetime record is **3 successes / 36 failures, 16 protective / 16 genuine**, so the
> corrected rate is **16%**, not the ~18% estimated here. The conclusion is unchanged (still under the
> 25% floor, still correctly held). Full fleet-wide partition and the decision: §3 item 2.

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
   `466`). ~~Expect: `literal_markdown` (10 rows, floor-held, oldest 08-17) to escalate once past the
   3-day limit ~**08-20**; `placeholder_contact` (3 rows, canary-held, oldest 08-16) ~**08-19**.~~
   The `pre_query_result` line in `kafka-scheduler` logs is the instrument — it should now carry
   `watching` and `watching_detail` even on an idle tick, which is `466`(a) working.

   > **⚠ CORRECTED 2026-08-18 18:20Z — BOTH DATES ABOVE ARE ONE TICK EARLY, and the miss looks
   > exactly like `466` being broken.** I computed them as `(oldest + 3 days)::date`, which discards
   > the time of day. The task fires at **12:57 UTC** (`interval_seconds=86400`,
   > `last_triggered_at` 12:57:48) and its predicate is `min(created_at) < now() - interval '3 days'`,
   > so a row created at 19:17 is 6h20m short at the 12:57 tick and waits a **further full day**. A
   > "3-day limit" on a daily task is really **3–4 days**. Re-derived read-only from `466`'s own
   > `classified` CTE, [MEASURED 2026-08-18 18:19Z]:
   >
   > | pair | hold_kind | rows | oldest (UTC) | FIRST TICK THAT ESCALATES |
   > |---|---|---|---|---|
   > | `placeholder_contact → page-build-handler` | canary | 3 | 08-16 19:17 | **08-20 12:57** |
   > | `dead_fragment_link → page-build-handler` | canary | 1 | 08-18 01:38 | **08-21 12:57** |
   > | `literal_markdown → page-build-handler` | floor | 10 | 08-17 19:21 | **08-21 12:57** |
   > | `missing_conversion_path → content-gap-planner` | canary | 1 | 08-17 22:21 | **08-21 12:57** |
   >
   > **So the 08-19 12:57 tick — the first under `466` — escalates NOTHING, and that is CORRECT.**
   > It should log `escalated=0` with `watching=15`, which is `466`(a) working (a `HAVING` that still
   > speaks on an idle tick). Do not read that zero as a failed migration. Conditional on the held set
   > not changing: the clock keys on `min(created_at)` per PAIR, so if the oldest row leaves
   > `detected` the date moves OUT. Logged in `WRONG_CALLS.md` and `LANDMINES.md` (2026-08-18).
2. ~~**The floor counts protective refusals as failure**~~ — **MEASURED AND DECIDED 2026-08-18
   18:40Z; `471` applied. Do not redo the measurement.** Fleet-wide over the 948 `failed` rows the
   floor actually reads (live UNION archive): **protective refusal 434 (45.8%)** — 418 of them
   `rebuild_policy=owned` — **transient/infra 234 (24.7%)**, **housekeeping that was never an attempt
   110 (11.6%)**, **genuine non-repair 170 (17.9%)**, and that last is an UPPER bound (~9 rows
   misfiled into it, both corrections pushing the same way). Protective share by month: 04–06 **0%**,
   07 **1.5%**, **08 61.6%** — bursty, not drift; today 66 of 74.
   - **The floor's arithmetic is NOT changed, and that is the decision, not an omission.** Under the
     promoter's FULL predicate (`c = 0 OR under-floor` — *not* the floor formula alone) **0 pairs
     flip**. My first pass said `placeholder_contact` flips; that was my own error, caught before
     recording — it has zero successes, so the *canary* rule holds it whatever the floor says.
     `literal_markdown → page-build-handler` refines to **3/(3+16) = 16%**, still correctly held.
   - **Why not put the classifier in the gate:** `error ILIKE '%rebuild_policy=owned%'` in a live gate
     makes an error message's *wording* load-bearing across services — reword it, silently change what
     the promoter dispatches, no test fails. The sound fix is a **structured refusal signal** (a
     distinct terminal status, or a `result` key the handler sets), which is a new shared vocabulary
     on a shared seam ⇒ **architecture-scope**, to be taken with `bugs_open/295`. **Left open
     deliberately; named so the omission reads as a decision.**
   - **What WAS fixed (`471`, applied + ledger-recorded + `_ROLLBACK.sql`):** the floor-held
     escalation payload. `466`(b) told the reader *"FIX THE HANDLER"*; on the **08-21 12:57Z** tick
     that would have gone to `literal_markdown → page-build-handler`, **16 of whose 36 lifetime
     failures are the handler correctly refusing**. It now leads with *"FIRST PARTITION THE
     FAILURES"*, carries the numbers, and hands the reader the query. Text-only by construction
     (single `replace()` on the live `pre_query`); `EXPLAIN` proved it still parses without executing
     its `UPDATE`; positive control proved the text is reachable (10 rows).
   - **`472` (applied + ledger-recorded + `_ROLLBACK.sql`) corrects `471`'s onward pointer.** `471`
     shipped **`bugs_open/295`** into that payload — and 295 **closed 2026-08-17**, so the path does
     not exist. I copied the reference out of this handoff's own §1 without resolving it. The payload
     now names `bugs_closed/295` and its **live residual — fix candidate 3, route owned-page content
     findings to `section_edit` (18 completes)** — plus the `apply_section_edit` trap (right for
     REWRITING a component, a dead end for ADDING a section). ⚠ **§1 above still says
     "touches `bugs_open/295`" — it is `bugs_closed/`.** `WRONG_CALLS.md` has both.
3. **`277` → `bugs_closed/`** ~**2026-08-22**: churn guard + the two cancelled conversions
   re-raising. **Both paths on the commit** — LANDMINE.
4. **`083` → `bugs_closed/`** — its arc is demonstrated (the 08-17 canary, and now `465` proven);
   the risk that blocked it is solved. Check nothing else is outstanding, then move it.
5. **`router_engine` (RFC_030)** — phase 1 measurement DONE and in that lane's NOTES; phase 2 is a
   council design round on shape A vs B, **before building**. ⚠ its PLAN's guarantee 8 is STALE
   (RFC_022 is CLOSED; counter live since 08-13, owner-ruled N=10, daily CronJob) and that bears
   directly on the A-vs-B choice — fix the PLAN before submitting. **Now the largest real work here.**
6. **Council gate's config blind spot** — `097` scopes on `platform/`/`internal/`/`pkg/`, so a
   config-only mechanism cannot be submitted; another lane filed `landmine(297)`. Every migration in
   this lane since `444` has been unreviewable for that reason.
7. **Consider submitting `465`+`466` to the council** — they are config-only, so §3.5 applies and it
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
