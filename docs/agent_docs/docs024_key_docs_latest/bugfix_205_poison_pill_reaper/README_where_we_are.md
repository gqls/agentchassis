# Where we are — bug 205 (plain prose, append-only)

## 2026-08-06/07 — what this bug turned out to be

The bug was filed as "a vet-data step has no output limit configured, so it runs at
a tiny built-in default, and two records are too big for it and retry forever."
That's true, but it's the smaller half.

The bigger half: there is a housekeeping job that runs every three minutes and
"rescues" any verification task that has been sitting in progress for more than
twenty minutes, by putting it back in the queue. That's the right thing to do when
a machine crashed mid-task. It's the wrong thing to do when the task itself is
doomed — it will fail exactly the same way every time. The housekeeping job can't
tell the difference, and it never counts how many times it has rescued the same
task. So a doomed task gets re-run, forever.

We measured it: in one day, 33 doomed tasks were re-run about fifty times each —
over 1,500 wasted runs, every one failing. Only one of those was the expensive kind
(it makes a paid AI call each time); the other 32 fail cheaply while trying to fetch
websites. Nobody noticed the cheap ones because the alarm that found this bug only
watches the paid calls.

The fix has two parts. First, teach the housekeeping job to count its rescues: after
five, it parks the task as failed with a note saying why, instead of re-queueing it.
That change is pure configuration — it takes effect the moment we apply it, no
software release needed, and it stops tonight's waste. Second, two small code
changes: one closes a back door where a parked task could be silently re-created if
a currently-switched-off sweep is ever switched back on, and one makes the platform
say out loud, in its logs, whenever a step runs without anyone having chosen its
output limit — so the next "nobody sized this" step is visible on its first run, not
after 64 failures.

What we're deliberately NOT doing: raising the vet step's output limit. That's the
owner's call (it costs money and belongs to the vet lane). Parking stops the
bleeding either way; the parked task is the reminder to make that call.
