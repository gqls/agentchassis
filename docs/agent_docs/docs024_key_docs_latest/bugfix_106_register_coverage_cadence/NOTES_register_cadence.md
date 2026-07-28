# NOTES — bug 106, register cadence. Append-only, newest at the bottom.

## 2026-07-28 — picking it up

`who-owns.py 106` → no owning workstream. The `docs026_concept_register/` directory
has commits from today, but reading them they are *other threads adding entries*
(DMR-002, SCR-004, VIZ-010, my own WFA-002) — nobody working the detector. Last
commit touching the bug file itself: a triage sweep, 07-27.

The bug was already 90% done by the thread that filed it — both candidates
implemented within hours. What survived was one sentence of scope: the sensor has
no cadence. I liked this one because the remaining work is small, entirely
docs/scripts, and the *reasoning* in the file is unusually good — it caught its own
implementation being half-finished and said so.

## Why a commit-path trigger and not a cron

The bug asks for "a cadence instead of by coincidence", and a cron is the obvious
reading. I went with a commit-path check instead, deliberately:

- a cron reports drift up to a week after it appears, **to nobody in particular**;
- the commit hook reports it **the second it is created, to the person creating
  it**, who can close it in ten seconds by adding one register line.

Same standing practice as the warning I put inside `092_TRIGGER_experience_plan.sh`
earlier today rather than only in a doc: put the check where the error is MADE.

## Measured before inclusion — and the rate is oddly low, which needed checking

`pattern-check.py` sets its own bar: a new check states its fire rate against real
history. Mine: **4 fires in 1,500 commits, 0.27%** — quieter than every existing
check (README 0.7%, SUMMARY 2.0%, twin ~2%).

I nearly wrote that up as "high precision" and moved on. **A very low rate and a
dead check look identical from the rate alone**, so I inspected all four:

```
61a89a926  memory_index
50477bee5  bugs_sweep_2026_07
e6a1a7187  bugfix_066_spawn_image_tag
33d13b05a  gemini_content_provider
```

All four real. And the last two are **exactly the pair 106's own triage records the
sensor catching on 2026-07-27** — which is the strongest evidence available that
the trigger is correctly placed: the same two gaps, found at creation instead of
days later by a human who happened to run the tool.

Then the induced gap (three arms, all pass) because the bug demands it, plus a
negative control: 40 ms and silent on a commit touching no workstream directory.

## The other half of the register problem, which this fix does NOT touch

Twice today I hit the register being **wrong**, not **absent** — a different failure
mode with the same consequence:

1. `SCH-012` carried `verify-later: scheduled_tasks.enabled for
   'diagnose-pipeline-trigger' (should still be false unless deliberately turned
   on)`. It had been **true** for weeks. That stale expectation is part of why
   `bugs_closed/124` went unnoticed — a `verify-later` that states an *expected
   answer* rather than a *question* reads as reassurance and nobody re-runs it.
2. The `psql -t -A` command-tag trap had been found and fixed by two separate
   threads, each in a comment in its own script, neither anywhere findable — so a
   third thread (me) shipped it into a claim guard.

Neither is coverage. The sensor deliberately does not judge accuracy, and widening
it to do so is how a coverage check becomes an audit nobody runs. **Recorded here
rather than fixed**, because it is a real and separate defect: *the register can be
stale in content while being complete in coverage, and nothing detects that
either.* Candidate for its own bug if someone wants it — I am not filing one on a
sample of two.
