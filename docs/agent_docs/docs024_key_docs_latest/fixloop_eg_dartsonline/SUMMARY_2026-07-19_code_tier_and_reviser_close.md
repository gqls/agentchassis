# SUMMARY — 2026-07-19: the reviser bug closed, and the diagnoser learns to search code

*Milestone read-out, written to be read aloud. Current state only — the
chronology lives in `README_so_far.md`, the technical log in
`NOTES_running_fixloop(10).md`.*

---

## What we're trying to do

Build a system that finds its own bugs, diagnoses them with evidence, proposes a
fix, argues that fix out among a council of reviewers with different specialities,
and opens a pull request for a human to approve. Nothing merges itself. The
property that makes it trustworthy is that it refuses: it will not confirm a
diagnosis without evidence, will not confirm on code alone without seeing the
problem actually happen, and will not bless a fix that only half-solves a
problem.

## Where we've come from

The loop was designed and built over the past ten days. It has since diagnosed
its first real bug correctly, found a second bug by failing honestly, and grown
two new abilities directly out of moments where it admitted it could not answer
something. The council of reviewers has grown from two seats to thirteen. Most
of the recent work has been less about new capability and more about the seams:
things that were true when built and quietly stopped being true as the system
grew around them.

## What we've done today

**Closed a live bug that was reading as fixed.** When the council grew from six
seats to thirteen, the component that rewrites a rejected plan was still being
handed a hand-written list of reviewers, so it never saw most of the objections
it was supposed to address. That was known and had been fixed — but only on one
of the two roads out of the council. If the council merely asked for changes, the
fixed path ran. If it vetoed the plan outright, a different path ran, still
working from a list of two reviewers out of five.

Two patches, written the same day by people who understood the problem exactly.
The difference was where each put the shared lookup step: one put it before the
fork in the road, so everything downstream inherited it; the other put it on one
branch. Only one of them closed the bug — and the half-fix declared itself
complete in its own notes, which is precisely why nobody looked again. That
pattern is now written into the debugging guide, because it has nothing to do
with councils and will recur.

We also established that the third component carrying a council never had this
bug at all — it has no automatic rewriter, it hands objections straight back to a
person. Earlier handoffs carried an instruction to mirror the fix across to it.
That instruction was never needed, and saying so in writing saves the next person
an afternoon.

**Gave the diagnosing half the ability to search the codebase.** Until today the
loop could follow a trail — this function calls that one, look there next — but
could not ask a question about the code as a whole: does this same mistake exist
anywhere else, is there a second implementation, what else touches this? It could
only see code the trail already reached. That is a real limitation, because the
causes worth finding are usually *not* where the symptom is.

The reviewing half already had this ability, built last week after a reviewer
spent three rounds stuck on a question nothing could answer. So today's work was
mostly plumbing rather than invention — the same machinery, wired into the other
half of the loop.

Two details decided whether it would help or quietly harm:

- The loop stops itself when a round produces no new evidence. But a round that
  says "I can't settle this, go search the code for X" produces no evidence *by
  definition* — the answer comes next round. Without teaching that guard that
  asking a question is progress, using the new feature would have looked like
  going in circles, and the loop would have halted one round before the answer
  arrived.
- The loop deliberately refuses to confirm anything on code alone; it wants the
  mechanism in the code *and* evidence of it actually happening. Code-search
  results are code. Printed in the wrong place, they could have been cited as the
  "it really happened" half — letting a confident story with no evidence behind it
  pass the one check designed to stop exactly that. They now sit under their own
  heading that says, in words, that this is code and cannot show occurrence.

## Where we are now

The reviser fix is **live** — it was configuration, so it took effect
immediately. It is verified structurally but has not yet been through a real
veto, so the first one will be the true test.

The code-search ability is **written, tested and committed, but dormant**. It is
Go, and Go does nothing here until a new image is built and rolled out. The
accompanying instruction that tells the diagnoser this ability exists is written
and ready but deliberately **not** switched on, because turning it on early would
invite it to ask questions nothing can answer — and an unanswered question looks
exactly like an empty answer, which reads as "no, that doesn't exist anywhere".
That is the most dangerous wrong answer this feature could produce, so the order
is: image first, then the instruction.

One piece of housekeeping worth knowing: the shared working tree currently does
not compile, because another concurrent session changed a function without
updating its test. That is their work in flight, not something to fix around.
Today's changes were tested against a clean copy of the last committed state
instead, so "the tests pass" means what it says.

## Where we're going

1. **Roll an image** carrying the code-search work, then switch on the prompt and
   watch the first diagnosis that uses it. Until that happens the feature is real
   but unproven, and we should say so plainly.
2. **Prove the other half of the older reviser bug** — a related fix from
   yesterday is still unexercised, and there is a subtle timing trap in how you
   verify it (a run can *look* post-fix while having actually started before it).
3. **The approved fix for the first real bug** is sitting ready for the
   implementer, which would open a genuine pull request. That is a decision for
   the thread that owns that bug, not this one.

Everything remains manual. Nothing dispatches itself, each run costs credits, and
the owner says go per item.
