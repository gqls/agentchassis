# SUMMARY — 2026-08-21b — the work-item terminal-write contract (bugs_open/307 → closed)

*Written to be read aloud. The **second** summary of 2026-08-21, from the lane that BUILT the
contract; `SUMMARY_2026-08-21_…` is the first, from the session that VERIFIED and closed it. Same
day from the two ends of the work — the series is the record, so read both.*

> **Why this is a `b` file, twice over.** I wrote it with a shell redirect to a path I had never
> read and destroyed the first summary; the pre-commit pattern check caught it, and it was restored
> from git. The repair itself then mis-assigned — both filenames ended up holding the *other*
> lane's text, so for a short while this account was absent from `HEAD` entirely. Recovered from
> `83622e60b`. Recorded here rather than tidied away because the series exists precisely to show
> how a day's understanding moved, and a file that quietly lost half of it would be the failure the
> rule was written against. Incident: `WRONG_CALLS.md`, 2026-08-21.

## What we were trying to do

A three-hour GitHub outage on 17 August killed a hundred pieces of queued work outright. The
owner's ruling the next day was one sentence: **a transient blip should return the item to queued.**
The job was to make that true — and to do it as one mechanism rather than three patches, because
the bug's own diagnosis had already worked out that the three defects were one seam.

## Where we came from

The retry machinery existed and ran. It failed for two unrelated reasons that happened to meet.

It retried **immediately** — three attempts inside a few minutes, all into the same dead
dependency. Retrying with no wait is the same as not retrying, if the thing you depend on is down
for two hours.

And a second piece of code wrote "this failed" **without counting the attempt at all**. Five live
agents used it. For those, one failure was the end — on an ordinary day, with nothing wrong.
Counted properly, that path was costing **401 of 558 failures in a fortnight, 72%**, and nobody had
noticed because a dead item looks the same either way.

There was a third thing, contributed by another lane: when a handler deliberately decides
something — "a human needs to look at this" — the failure path could silently write over that
decision.

## What we did

One shared piece of code that every failure path now goes through. It counts the attempt, waits
before the next one (using timings from a policy table that already existed and had been waiting
for a second user since it was built), refuses to overwrite a deliberate decision, and — when it
can see the whole fleet failing the same way at once — puts the item back **without spending an
attempt**.

That last part is the only genuinely new machinery, and the design decision worth recording is what
it does *not* do. A deleted repository and a GitHub outage produce identical error text. So we
don't judge the message at all; we ask a different question — *is anything else failing this way
right now?* An outage shows up across different customers and different kinds of agent within
minutes. A deleted repo fails alone. Measured against a week of history: exactly three things
trigger it, all three were real outages, and nothing that was one item's own fault ever did.

It went through the review council twice, and the council **earned its place twice**.

## Where we are now

**It works, it is live on the whole fleet, and it is proven on real traffic** — not in a test.
Items have been returned to the queue without losing an attempt and have gone on to succeed; the
daily bleed reads zero; and a synthetic canary drove every remaining arm end to end.

**Three real defects were found after we thought it was done, and two of them were mine.**

- The council's first round caught me **reintroducing the exact bug I was fixing**, one line from
  where I fixed it — I reused a list of protected statuses on a neighbouring code path where it had
  to differ, and told the reviewers it matched when it didn't.
- The canary caught something worse: **the code that gives up honestly at three attempts crashed
  every time**. A parameter was passed to the database that the statement no longer mentioned.
  Fifteen tests and five deliberate sabotages missed it, because they all tested the SQL as *text*
  against a fake database, and the fault was in the SQL as a *program*.
- Fixing that exposed a fourth: once retries actually worked, the surrounding job started reporting
  success two seconds later and **stamping the retried item "complete"**, wiping the retry. Live,
  on a real customer page that never got its content while the record said it had.

All four are fixed, live and verified. `307` and `344` are closed; the fifth writer's cooldown is
applied and waiting for natural traffic to exercise it.

## Where we're going

Almost nothing, deliberately. One standing watch — the outage detector has still **never fired in
production**, because there has not been an outage since it shipped, and no test can substitute for
that. It lives in one file with a trigger that reopens the bug if it fires. Four design questions
are with the owner; the mechanism runs correctly whichever way they go, but they decide what the
next author copies.

**The thing worth carrying out of this lane is not the mechanism.** It is that every defect above
was found by something mechanical — a canary that drove the real path, a mutation that deleted the
code and checked something went red — and none was found by reading harder or being more careful.
Four separate times this lane produced a confident, wrong measurement: a filter that could not
match what it searched for, sabotages that never applied, an attribution by the wrong column, and a
time window that had not opened yet. Each was caught by the same habit: **before believing a
number, ask what result would have proved you wrong — and check that result was reachable.**
