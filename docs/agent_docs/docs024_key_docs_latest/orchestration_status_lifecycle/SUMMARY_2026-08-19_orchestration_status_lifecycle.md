# SUMMARY — 2026-08-19 — orchestration status lifecycle

## What we were trying to do

Make it impossible for a job to get stuck in a state that nothing ever looks at again.

## Where we came from

On 17 August a bug was filed saying that a job which stopped in one particular internal state would
never be cleaned up by anything. The oldest one found had been sitting there nineteen days, through
several fleet restarts, and each one was quietly holding on to two Kafka topics it would never
release — which fed a separate, open bug about the message system running out of memory.

The fix looked like a one-liner. A cleanup job runs every three minutes failing jobs that have
obviously died, and it worked from a list of states it knew about. One state was missing.

## What we did

We added the missing state — and then discovered we had fixed an instance rather than a defect, four
times over.

The bug file told us to re-run a particular measurement first, because that measurement was what made
the four-hour cutoff safe. **It had gone useless.** It now gave the same answer whether the system
was healthy or broken, because the evidence it counted had been cleared the week before. Following
the instruction to the letter would have produced a green light that could not have come out red. So
we read the code instead and found a permanent argument: the state is set on one line, by one caller,
and the very next thing that code does is move the job out of it. It is meant to last milliseconds.

Then the same gap appeared one state over — and another session was fixing it at the same moment. We
produced character-for-character identical SQL without knowing about each other. The one difference
mattered: they reused the reasoning from the first state and were wrong by a factor of about a
thousand, because the second state genuinely waits for a message, so its duration depends on how busy
the system is. We measured it instead — five and a half thousand cases, longest ever six seconds.

That is when it became clear the defect was the *list*. Any state nobody thought to add was invisible
for ever. So the list is gone: the cleaner now works from a rule — anything not finished, not
deliberately waiting, waiting for nothing, and not moved in four hours. We proved it by inventing a
state that has never existed, planting a job in it, and watching the real cleaner reap it while
leaving a healthy one alone. A second cleanup job had the same blind spot and now follows the same
rule.

Even then, "which states count as finished" was still written out by hand in two places. So there is
now one table saying what every state means, and the database itself refuses to write a state that is
not in it.

Along the way we deleted two things that were lying to us: a monitoring module that read a table
which has never existed, and a "pause for human approval" feature declared in five places,
implemented in none, whose two halves disagreed about its own name — so the four guards written to
protect a paused job could never have matched one.

## Where we are now

Every layer is live and council-approved. No job has been stranded for more than four hours since.
The vocabulary table holds seven states, the database enforces it, both cleanup jobs read it, and the
code changes shipped on the current build — checked against the running binary, not assumed.

The memory bug this fed into is closed, on the honest basis that its symptom is gone and a mitigation
is holding, **not** that its root cause was proven fixed.

## Where we are going

Three things are tracked elsewhere, deliberately. An RFC asks where the rule should live for choosing
between the two ways this codebase now enforces a vocabulary. A filed bug records that two pods can
still, in principle, take over the same job at once — decided by a clock rather than a lock, which
has been true for months and is not on fire. And one deployment file still carries the old memory
limit that caused the original incident, so deploying that service elsewhere would reproduce it.

**The lane itself is closed.**

## The thing worth taking away

Every expensive mistake in this work was the same shape: **trusting a check that could not have come
out any other way.** An expired census. A structural argument borrowed between states that were not
alike. An md5 that passed on a truncated file. A grep scoped to the wrong directory. A pod probe that
read one replica of twenty-seven. None of them was a missing check — each was a check that could only
ever say yes.
