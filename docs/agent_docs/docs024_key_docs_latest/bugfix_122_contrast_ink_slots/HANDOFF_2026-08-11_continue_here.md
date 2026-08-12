# HANDOFF — bug 122 lane. Written 2026-08-11.

> **⚠ SUPERSEDED FOR STATE 2026-08-12 — START AT
> `HANDOFF_2026-08-12_continue_here.md` INSTEAD.** This file stays the reference for the
> improvement-sweep episode: the cost table, the guard arithmetic, and the two corrections it
> forced. Its banner below is RESOLVED, and its NEXT list has been overtaken — the contrast
> park's trigger changed, and detection is now driven by a rotation, not the sweep.

Supersedes `HANDOFF_2026-08-10_continue_here.md` for **state**. That file is still the
reference for the delivery evidence (§1–§4a: what shipped, the 10 closures, the two caps)
and is not repeated here. **§7 of it is superseded by this file.**

## ⚠ THE ONE THING THAT CANNOT WAIT

> **DONE 2026-08-11 18:00:39Z — `improvement-sweep` is DISABLED.** Ran 12:31:40Z → 18:00:39Z
> = **5h 29m**, ~22 fires. Evidence, cost table and the two corrections it forced are in
> `NOTES_contrast_ink_slots.md` (entry "the sweep ran 5h29m and was stopped"). Headline:
> `detected` 193 → 25, but that queue **moved rather than emptied** — 526 new rows filed
> (~105/h) against ~48/h completed, open work ~273 → **544** — and the guard counts
> `triaged`, so **sites over the guard went 5 → 1 (park) → 8 (now)**, worse than the state
> the park was done to fix. **Do not re-enable without reading that entry.**
>
> **RESOLVED 2026-08-12 12:39Z (v1.0.1290): the queue cleared overnight exactly as
> predicted, and the prediction could have come out otherwise.** `triaged` **446 → 0**,
> `complete` +542 (2,803), `failed` 66 → 15, and the guard census is **0 of 22 sites locked
> out** (from 8). Sweep still `enabled=false`; 226 contrast rows still `deferred`. This is
> the behavioural confirmation of MISSTEP 20(a): had the drain depended on the sweep,
> `triaged` would have sat at 446 all night.

**`improvement-sweep` is RUNNING and nobody has scheduled its stop.** The owner asked for
"a short while"; nothing expires on its own. Re-enabled 2026-08-11 12:31Z at 900s.

```sql
-- STOP IT:
UPDATE scheduled_tasks SET enabled = false, updated_at = now() WHERE name = 'improvement-sweep';
-- WATCH COST (baseline below):
SELECT date_trunc('hour', created_at) AS hr, count(*) AS calls,
       sum(input_tokens) AS in_tok, sum(output_tokens) AS out_tok
  FROM llm_call_log WHERE created_at > now() - interval '6 hours' GROUP BY 1 ORDER BY 1;
-- WATCH PROGRESS (the point of the exercise):
SELECT status, count(*) FROM site_work_items
 WHERE item_type='page_rerender' AND created_at > '2026-07-24' GROUP BY 1;
```

**Spend baseline, taken immediately BEFORE enabling** (calls / in / out per hour):
08:00 `8 / 40,076 / 13,228` · 09:00 `28 / 175,532 / 87,506` · 10:00 `134 / 514,812 / 361,053` ·
11:00 `57 / 397,350 / 80,514` · 12:00 `31 / 113,502 / 60,002`.
**Judge the sweep against 10:00's 134 calls — that hour had no sweep and is the fleet's
own busy-hour shape.** A rise to a few hundred calls/hour is the mechanism working; a
jump into the thousands is not, and the STOP is one UPDATE.

**WHAT ONE FIRE ACTUALLY COSTS — read before judging the spend.** An `improvement-sweep`
fire is *not* a cheap triage. Observed on the first fire (orchestration
`04b26f88-5ee2-44a8-8ff9-3ce371e8e3b2`, 12:31:40Z): the loop's first step is
`call_quality_discovery` — it runs **discovery agents against the site** before it ever
reaches `triage_findings`. So each fire is an LLM-heavy site pass, and the re-render drain
is a *downstream* effect of it, not the whole of it. That is why 180s was expensive and why
900s was chosen.

> **CORRECTED 2026-08-11 (evening): "the re-render drain is a *downstream* effect of it" is
> FALSE, and it was the claim the stop/continue decision rested on.** Completions ran **49/h
> and 48/h at 10:00 and 11:00 today**, before the 12:31Z re-enable, and **85 / 110 / 41**
> across 08-10 13:00–15:00 with the sweep disabled — against sweep-era 35 / 32 / 63 / 53 / 58.
> The drain is a **separate, always-on path at ~50/h**; the sweep adds **discovery and
> promotion, not execution**. So stopping it costs **nothing** in re-render throughput and
> only stops new arrivals. Caught by one `GROUP BY` over a window when the sweep was OFF —
> the check that should have preceded the claim. Also note the cost criterion below
> (calls/hour) never tripped while **input tokens ran 3.2x** the pre-sweep average: a call
> count is not a spend when each call is a whole site pass. **Expect promotion to lag the fire by minutes**, and do not read "fired,
but nothing triaged yet" as a failure — I did, three minutes in, and was simply early.

**Starting point for the drain:** `page_rerender` = 193 `detected`, 2,017 `complete`,
80 `unresolved`, 62 `failed` (12:31Z, at the moment of enabling).

## What happened today

**1. Owner decision executed — migration `389`** (applied + recorded 12:31Z, committed
`81000a1ed`). Two halves in one transaction:

- **Parked 226 `contrast_failure` items** `detected` → `deferred`.
- **Re-enabled `improvement-sweep`**, cutting its cadence `180s → 900s` (~20 site-runs an
  hour → ~4) because the owner is cost-wary.

**The park is not a liberty taken with "for the rerenders" — it is what makes it work.**
`improvement-sweep`'s `pre_query` skips any site with **≥50** `(triaged,detected)` items on
`pipeline='build'`. The overnight audit's 226 findings had pushed **five** sites over that
guard, including the two holding the most re-render work. Measured, before and after:

| site | backlog | contrast | rerenders | after park |
|---|---|---|---|---|
| robot-hands.com | 68 | 34 | 17 | **34 — in** |
| loancalculator.co.uk | 58 | 0 | 12 | 58 — still out (own backlog) |
| vonc.com | 55 | 38 | 13 | **17 — in** |
| leopardessconsulting.co.uk | 53 | 8 | 22 | **45 — in** |
| loanandmortgagecalculator.co.uk | 52 | 3 | **40** | **49 — in** |

Sites over the guard: **5 → 1**, confirmed by the migration's own post-check.

**Why park rather than promote:** `triage_detected_items` is site-scoped and **type-blind**
(`triage_detect_items_action.go:162-173` — `WHERE site_id = $1 AND status = 'detected'`), so
a sweep promotes contrast findings alongside re-renders whether or not anyone wants it to.
They route to `css-patch-agent`, where **`bugs_open/213`** is open. 226 possibly-false
`complete` rows are far harder to find later than 226 `detected` ones.

**`deferred` was verified as the correct park on both properties that matter**, not assumed:
triage only promotes `detected`; and `deferred` is **NOT** in `idx_swi_dedup`'s terminal
list, so a parked row **keeps its dedup slot** and next week's audit files no duplicate.
That is what makes the park safe to leave for weeks.

**2. Contributed into `bugs_open/213`** (OWNED by `bugfix_213_verifier_producer_join` —
`who-owns.py` says so; I did not compete). Its fix is **already live and pod-proven**; what
it lacked was behavioural proof. Measured: all 26 `_verification` rows predate the roll
(latest 2026-08-09), so `out_of_scope = 0` means the gate is **idle, not blind**. The
verifier demonstrably *can* refuse (4 `defect_persists`) and reports its own errors (2
`error`) — neither is the false-complete shape. **Enabling the sweep is what will finally
exercise it**, which is now written into their file along with my lane's stake.

**3. Verification method corrected — `strings` is retired.** CLAUDE.md changed today: the
`strings /app/agent-chassis | grep -c` recipe (which yesterday's handoff §1 enshrines)
"produced three confidently wrong readings in one day". Also `v1.0.1284` shipped **three
revisions under one tag** (`bugs_open/249`), so a tag is not a revision.

> **Today's engine proof, on `v1.0.1286`, both replicas:** the provenance log line had
> already rotated away, so I used the sanctioned binary probe with controls —
> `kubectl exec <pod> -- grep -aq "<needle>" /proc/1/exe`:
> `buildLegibleInkDefaults` **PRESENT**, `fillDarkSchemeSpecialisedSlots` **PRESENT**
> (positive control), `zzzInventedControlXyz` **absent** (negative). VIZ-014 is live.
> **Do not carry this forward past the next roll**, and do not use `strings`.

## Where the three decisions stand

| | decision | status |
|---|---|---|
| 1 | promotion gate | **RUN AND STOPPED** (12:31→18:00Z). Park held (226 still `deferred`); sweep re-locked 8 sites with its own promotions — see banner |
| 2 | `bugs_open/212` §8 — component-painted grounds (~24 failures) | **Still owner's.** Architecture, not a bug patch. Unchanged |
| 3 | `bugs_open/242` — the silent 25-page cap | **Open, unstarted.** Next code task in this lane |

## NEXT, in order

1. ~~**Watch the sweep, then stop it.**~~ **DONE 18:00:39Z** (see the banner). What replaces
   it: **nothing, deliberately.** 544 open re-render items drain at ~50/h with arrivals
   stopped ⇒ ~11h to clear, and the 8 guard-locked sites come back under 50 by themselves as
   they do. **Before re-enabling, decide the arrival-vs-completion question**, because the
   sweep files ~105/h into a path that clears ~48/h — a cadence slower than 900s does not fix
   that, it only slows the divergence. The honest options are: leave it off until the queue
   clears; raise the guard (it is the thing the park was spent on); or make the sweep's
   discovery step conditional on the site's own open-item count rather than firing every time.
2. **`bugs_open/242`** — the cheapest fix is parity with the drain one step downstream: put
   `pages_total` into the summary the adapter returns, so a capped sweep cannot read as a
   complete one. Every one of the 19 overnight sweeps measured **at most 25 pages** and
   none said whether that was all; **226 is a floor, not a census.**
3. ~~**Unpark the contrast items when 213 closes**~~ — **TRIGGER CORRECTED 2026-08-12: 213 is
   NOT the gate, and closing it changes nothing here.** `contrast_failure` has **no
   registered verifier** (filed at `write_render_audit_findings_action.go:258`; absent from
   every `RegisterVerifier*` call site, of which there are 12), so unparking mints 226 rows
   that complete **ungraded by construction** — not merely at risk from 213's predicate
   mismatch. **The real trigger: `contrast_failure` gets a verifier, or someone rules it
   needs none.** That is now the lane's cheapest high-value code task, and it is smaller
   than 242: one `RegisterVerifier` plus a predicate that re-measures the selector's
   contrast, which this lane already has the tooling for (VIZ-010, the Python contrast
   tool). 213's lane has been told they are no longer blocking us. The restore itself is
   still one UPDATE at the foot of migration `389`, predicated on
   `spec->>'parked_by' = 'migration_389'`; row-level backup
   `scratchpad/backups/backup_park_contrast_failure_20260811.tsv`.
4. **Free cross-check, if a lane re-renders robot-hands `/selection-guide.html`:** the
   audit filed `info-card-grid__card-link` + `__eyebrow` failures there, and migration
   `368` should close both. Grade at the next audit, never at the item status.

## Standing traps this lane has paid for

- **Grade per selector, never by fleet total.** It rose 109 → 112 while every targeted
  failure closed.
- **A filed count is not a found count.** "34 findings" was really 171 firm — 111 dropped by
  a cap, and the disconfirming keys were siblings of the one I quoted.
- **Read the selection before asserting it excludes your rows.** I wrote down two wrong
  causes for the 220 stalled items; each died to one further query, and the second was
  settled by opening the action.
- **A pathspec commit still takes a same-file passenger**, and the check is per-commit, not
  per-session — I ran it on one docs commit and not the next.
- **`pages.sections` is an array of plain strings**; an object-shaped census returns 0 rows
  silently. In `LANDMINES.md`.
