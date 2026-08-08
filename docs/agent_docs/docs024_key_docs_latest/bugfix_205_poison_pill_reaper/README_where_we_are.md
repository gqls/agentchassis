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

## 2026-08-07 morning — done, proven, and four decisions for you

The fix is now fully live. The housekeeping job has parked all 33 doomed tasks
(they stopped burning money at 01:40 last night and have stayed quiet since), and
this morning's software release carries the two code changes: the back door that
could have silently un-parked them is closed, and the platform now says out loud
when a step runs without a chosen output limit.

Decisions needed, in order of usefulness:

1. **The vet step's output limit.** The one expensive record can never verify
   until its step gets a real limit. 8000 matches what the rest of the fleet uses
   and is four times what its successful calls need. Say yes and we set it and
   un-park that one task to prove it completes.
2. **The other 32 parked tasks.** They fail fetching the practices' websites —
   likely dead or bot-blocking sites. Options: un-park them for one more bounded
   round (they'll park again after 5 tries if still broken), investigate the
   scrape errors first, or cancel them like the 574 before. No money burns while
   parked, so this can wait.
3. **Future-proofing note (from the review).** Any future task type on this queue
   inherits "park after 5" silently. Today none exist. If you'd rather each task
   type choose its own ceiling, that's a small follow-up; otherwise awareness is
   enough — it's written where the next builder will look.
4. **The bigger pattern.** This is the second or third housekeeping job that
   needed its own hand-written "count and park" logic. The review's architecture
   seat suggests a shared mechanism. That's an architecture-track item if you
   want it; nothing breaks without it.

## 2026-08-08 — all four decisions carried out; this workstream is done

You ruled on all four questions today, and all four are now done and proven:

1. **The vet step got its real output limit of 8000.** We then woke the one
   parked task that had been failing all along — it succeeded on its very first
   try. The document it produces needed about 3,100 tokens of output, half again
   more than the old built-in limit of 2048 could ever allow — arithmetic proof
   that no amount of retrying could ever have worked. That practice's record is
   now verified, after roughly 64 wasted attempts.
2. **The other 32 parked tasks are cancelled**, each with a dated note saying
   why — the same treatment as the 574 cancelled before them.
3. **and 4.** The housekeeping job's "count the rescues and park after five"
   logic is now a **shared mechanism** rather than a hand-written one-off: each
   kind of task can declare its own ceiling in a small table, and anything that
   doesn't declare one gets the safe default of five. The vet queue's
   housekeeping job now uses it, and it was tested by deliberately inducing both
   cases (an undeclared type parking at the default, a declared type honouring
   its own ceiling) with no residue left behind. We deliberately stopped there —
   no grand generalisation until a second queue actually wants it, because
   machinery with no user rots.

A later session re-checked the live system the same afternoon: the queue holds
only completed and cancelled work, nothing pending, nothing looping, and the
housekeeping job is running the new shared logic on schedule.

Nothing is left on this workstream. The one thing still worth watching for is
the new log warning: the platform will now say out loud when any of the seven
remaining unsized steps runs, and the first time that happens someone should go
and choose that step a limit.
