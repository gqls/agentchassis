# HANDOFF — provocation pipeline, 2026-09-02 (evening) — **LIVE AND SELF-DRIVING**

**Supersedes `HANDOFF_2026-09-02_continue_here.md`** (written this morning, when the change
was still waiting on a roll) **and everything before it.**

---

## In one line

vonc.com publishes a new provocation every day with no human in the loop; the code is
live, the schedules are running, six days are queued and four more are written — **nothing
is blocked and nothing needs doing to keep it going.**

## Status at a glance

| thing | state |
|---|---|
| chassis | **v1.0.1354**, both replicas, `326370d6c` proven live at the binary |
| human-approval stamp | **removed** from all three queries |
| readability rail | **FATAL**, and proven refusing in production |
| `685` | **APPLIED + recorded**, renamed off `_HOLD` |
| schedules | **3 live** — refill (12h), dating (24h), feed publish (6h) |
| queue | **2026-09-03 → 09-08 dated**, 4 more approved awaiting dates |
| councils | `c08d263a` **APPROVED**, `fb31e95e` **APPROVED** — both read, objections discharged |
| site today | still `22 Aug` — **correct**, see "the buffer" below |

## What is live

**The permission step is gone.** Three queries (feed / dating / exemplar selection) no
longer require `human_approved_at`. This is the owner's **third** position on the question
(none 07-31 → required 08-09 → none 09-02); the column and all history are retained, and
each site carries the instruction to restore it in **both** queries or neither.

**The readability rail rejects instead of recording**, as the deliberate replacement for
the reader that was removed — arithmetic rather than the judge, because the judge is
documented-stochastic on this corpus.

**The pipeline drives itself:** `provocation-shelf-refill` (12h, demand-driven — skips
entirely once 14 days of inventory exist) → `provocation-date-assign` (24h, dates up to 14,
tomorrow onwards) → `provocation-feed-refresh` (6h, publishes).

## How "live" was established — copy these, they are the discriminating checks

**1. The binary, not the tag.** The `build provenance` startup line had scrolled out of
`--tail=3000` (normal on this service; that means *not in range*, **never** *unstamped*).
The change removes a literal, so it was probed as a **capability**, on BOTH replicas, with
controls in the same breath:

| needle | expect | got |
|---|---|---|
| `human_approved_at IS NOT NULL` | absent | `0` ✅ |
| `ADVISORY: recorded, not fatal` | absent | `0` ✅ |
| `words/sentence, longest` | present | `1` ✅ |
| `publish_on IS NOT NULL` (POSITIVE CONTROL) | present | `1` ✅ |
| `zzz_string_that_cannot_exist` (NEGATIVE CONTROL) | absent | `0` ✅ |

**Never grep for the commit sha** — the binary stamps what it was BUILT from, so a correct
build made from a later commit reads absent. **Never use `strings`** (absent from the
debian-slim image; behind `2>/dev/null` its failure is indistinguishable from "not
stamped").

**2. The rail was INDUCED, because an all-pass run proves nothing.** The attended run
returned 4-of-4 approved with no rail note. That shows the new code ran; it does **not**
show the fatal path fires — a broken rule and a rule with nothing to object to look
identical. So a **real pre-rail body** was copied by `INSERT…SELECT` (never hand-composed —
see the 08-05 `WRONG_CALLS` entry) onto the isolation domain `calibration.vonc.com` and
judged by the live gate:

```
status = rejected · gate_version = 3
{"rule":"hard_to_read","fatal":true,"layer":"form",
 "detail":"grade 10.7, 16.4 words/sentence, longest 19 — …"}
```

Same text, same numbers the unit test sees.

## ⚠ The buffer — the site showing an old date is NOT a fault

`selectForDate` serves the latest entry whose publish date **has arrived**, and
`nextPublishDates` starts at **tomorrow at the earliest, never today**. So on the day
anything is written the site still shows the previous entry. That gap is the safety
property that replaced the human stamp: it is the window in which a bad row can be retired
before anyone sees it. **Do not "fix" it.**

## What to watch, and how to tell healthy from broken

```sql
-- 1. is anything driving it?  (three rows, all enabled)
SELECT name, interval_seconds, enabled, last_triggered_at
  FROM scheduled_tasks WHERE target_agent_type ILIKE '%provoc%';

-- 2. is there runway?  (want >= a few days dated ahead)
SELECT publish_on, left(title,50), gate_verdict->>'gate_version'
  FROM provocations WHERE domain='vonc.com' AND status='approved'
   AND publish_on > current_date ORDER BY publish_on;

-- 3. what is actually served?
-- curl -s https://vonc.com/data/provocations.json | jq '{today:.today.date, generated_at}'
```

**A refill that "did nothing" is usually correct** — the pre-query returns no rows once 14
days of inventory exist, and the scheduler logs *"Pre-query found no rows — task ran with
nothing to do"*. That is the design, not a failure.

## Known-inert, do not rely on it

**`concurrency_group` on these two tasks does NOTHING.** `cmd/scheduler/main.go` stamps
`last_triggered_at` **and** `last_completed_at` at FIRE time, and `countInFlightByGroup`
only counts a task whose `last_completed_at < last_triggered_at`. Measured on the live
rows: both timestamps identical, so neither task is ever "in flight".

I claimed the opposite in the `fb31e95e` submission and **that claim was false** — caught
by the `guardian` and `debug_historian` seats. Harmless here (the refill takes ~2 min,
runs every 12h, and inserts `ON CONFLICT DO NOTHING` on `(domain, slug)`) but **nobody
should build on it.**

## Owed — genuinely outstanding, not claimed

1. **The §10.6 LIVE calibration has NOT been re-run** (`PROVOCATION_LIVE_CALIBRATION=1`,
   real key, real tokens) and **two of its four bad-set fixtures were rewritten** to clear
   the rail so they reach the judge. **Do not cite that calibration as current.** This is
   the largest outstanding gap and it is a ~10-minute job for whoever has a key.
2. **Nobody has read the generated prose for voice.** The rail measures sentence and word
   length; it cannot see a riddle. The owner rejected the pool's *plainest* entry on 08-11
   with *"I don't even fully understand it"*. **A passing rail score is not "the reader
   will understand it."** The four new pieces are unreviewed by any human.
3. **First unattended day is 2026-09-03.** Nothing has yet published without a person
   watching. Worth one check that the site actually turns over.

## Traps this lane will hand you

- **`321` FAILS IF RE-RUN.** It RAISEs if the scheduler has a `scheduled_tasks` row, citing
  the 08-09 ruling the owner reversed. Expected. Do not repair the database to satisfy it
  and do not edit the applied file (checksum). Full entry in `LANDMINES.md`.
- **Both agent types are still named `-manual` while running on a schedule.** Renaming
  `agent_definitions.type` risks in-flight dispatch and breaks 321/371's own verify
  queries, so the truth lives in `description`. Read `scheduled_tasks`, never the name.
- **A `_HOLD` migration cannot be `--record-only`'d** while it carries the suffix — the
  sequence is hand-apply → rename → record, back to back.
- **Approved rows are never re-gated.** That is why the ordering guard existed, and it is
  also why a rule change never applies retroactively.

## Accepted debt, with a stated trigger

The `architecture` seat did not block but put on record that a yes/no policy flipped three
times in five weeks is being carried as **source-code surgery** (two paired predicates in
two files) rather than a config gate. **If it flips a fourth time, converting it to config
IS the work.** Not done now because building a switch on the day the owner rejected the
position it switches to is how mechanisms rot unexercised.

## Read next

`PLAN §15` (decisions and why, incl. §15.9 discharging the Go council's objections) ·
`NOTES` tail (measurements, the induced rail proof, and the misstep worth not repeating) ·
`README_where_we_are` tail (plain prose for the owner) · `RUNBOOK §16` (every command).

---

## > **CORRECTED 2026-09-02 (evening) — two of the three "Owed" items are RULED, not outstanding**

Caught by putting the list to the owner rather than carrying it forward. Full record with
its limits: `PLAN §15.10`.

**Owed 1 (the §10.6 LIVE calibration) is NOT work. Do not do it.** The owner ruled the
arithmetic rail sufficient: *"arithmetic rail is enough for now."* ⚠ The ruling retires the
**re-run**, not the **staleness** — the rest of that bullet stands word for word, and
**nobody may cite that calibration as current** (two of its four bad-set fixtures were
rewritten). If you are about to depend on a calibrated *judge* threshold rather than on the
rail's arithmetic, this ruling does not cover you.

**The "Accepted debt" section is CONFIRMED by the owner, not merely by the seat:** the
config gate is not built on this flip. The fourth-flip trigger stands unchanged.

**Owed 2 (nobody has read the prose) is the ONE live gap — and its date was wrong here.**
`[MEASURED 2026-09-02]` The rows dated **09-03 and 09-04 carry `human_approved_at`** and
were read before the change; the eight written on 09-02 do not. **So the first piece served
that no human has read is 2026-09-05.** "First unattended day is 2026-09-03" (Owed 3) is
true of the *machinery* and is **not** the reading deadline — reading it as one loses three
days in the wrong direction. All eight were presented to the owner on 09-02; retire recipe
is now `RUNBOOK §16g` (it did not exist when this handoff was written, which made the
buffer real but unusable).

⚠ Also do not chase the feed's `generated_at: 2026-08-22` **or** its
`Last-Modified: 01 Sep`. The two disagreeing is the no-change skip working as designed
(`provocation_feed_action.go:833`); neither is a liveness signal. See the `NOTES` tail.
