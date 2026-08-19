# SUMMARY — 2026-08-19 — the evidence was never lost, and the instrument was the thing that was broken

*Third in the series. Previous: `SUMMARY_2026-08-18b_part_a_proven_and_the_wedge_has_a_shape.md`.
Written to be read aloud.*

## What we are trying to do

Bug 029 is the one where the site-building pipeline quietly stops. A build job starts a helper,
the helper answers, and then the job simply never does anything again — it does not crash, it does
not report an error, and nothing anywhere says the pipeline is down. Four hours later a caretaker
process notices the job has gone cold and kills it. We want to know what stops the job, and we want
to stop it happening.

## Where we have come from

We fixed one real thing on Monday and proved it in production on Tuesday: retries were being given
*less* time the more time a step had asked for, so work was being abandoned early. That is fixed,
approved, live, and still aboard today's build. But we were always clear that this was not the
freeze itself, and it never claimed to be.

By Tuesday evening we had the freeze characterised but not explained: twenty-odd jobs, all the same
shape, all on one day. Then the record of those jobs was deleted by ordinary housekeeping, and we
told ourselves — in the bug file, in the handoff, and to the diagnosis service — that the evidence
had gone and there was nothing to do but wait for it to happen again.

## What we have done

**We found out that was wrong, and it was the most valuable thing this week.** We keep two records:
one of each *job*, wiped after about a day, and one of each *step inside* a job, **kept for a week**.
Every one of Sunday's frozen jobs is still in the second record. It holds **twenty** of them, where
the record we lost had only ever shown us eighteen. We had declared the evidence destroyed on the
strength of looking in one drawer.

That mistake had already cost us something concrete. We had sent the automated diagnosis service
after this bug that morning, and it came back "cannot verify". It was right to: we had told it the
evidence was gone, so it went looking for a fresh example, found a candidate, checked it properly,
and rejected it. **We had it refute a live bug by pointing it at the emptied drawer.**

With the evidence back we could test the leading theory, and **we ruled it out.** The idea was that
a background repair process gives each job a sixty-second allowance and shares it between up to
twenty-five jobs at once, so a job runs out of time part-way through. Three separate measurements
say no: the allowance is never actually shared — every one of **31,548** repairs in the past week
handled exactly one job, never two; the jobs do not run out of time, dying about twelve to
thirty-five seconds into a sixty-second allowance; and the place they actually die has no time limit
on it at all.

**And we now know much more precisely where the freeze is.** The helper starts, answers, the job
records that answer and saves its progress — and then dies in the very next instruction, when it
tries to hand the real work to the helper it just started. That last save and the helper's reply
turned out to be the same moment, which we had been reading as two. The search area went from "some
time after the helper starts" to one step of one function on one thread.

**Then we sent the diagnosis service back in, and it stalled for a completely different reason —
one that is ours.** The corrected instructions worked: it went straight to the right record and
started walking the right code. It then asked us what columns that record has, because **the
evidence pack we hand it does not describe that table at all.** It could not write a query against
the only table that holds the answer.

That is not a new problem. The same thing happened to two earlier investigations on a different bug,
for the same reason, on a neighbouring table — and the fix that worked then was simply to add the
table to the list of ones we always describe. **We have made that change**, with a test that we
deliberately broke first to prove it would catch a regression, and it is with the review council now.
It is written in Go, so it does nothing until the next build goes out.

## Where we are now

The freeze has not happened since Sunday, and we can now say that over **seven days** rather than
twenty-six hours — which is the practical payoff of getting the evidence back. But we are being
careful with that: six of the eight days in the record are also zero, so a quiet week is what this
bug looks like anyway, and silence is not evidence that anything is fixed.

Today's fresh build is verified aboard: we checked the running binary directly and it matches the
expected build point on both copies, with two different wrong answers correctly coming back absent.
Monday's retry fix is still in it.

So: 029 stays open, honestly. We have removed a wrong theory, recovered the evidence, narrowed the
search to a single step, and fixed the instrument that stopped us investigating. We have not yet
explained the freeze.

## Where we are going

Once the council has ruled and the next build carries the evidence-pack fix, the diagnosis service
can be sent in a third time with everything it needs — the right table, its shape, and twenty real
examples, which are good until about Sunday the 24th. That is the shortest path to an explanation.

Alongside that there is a genuine defect we found by reading the code and have not yet fixed: the
check that asks "did the reply arrive before we finished filing the request?" identifies the request
by the *name of the step* rather than by the request itself. So when a step is retried, it matches
the previous reply, concludes it has nothing to do, and reports success without saving anything. It
is real, it is small, and it is worth fixing whether or not it turns out to be part of this bug.
