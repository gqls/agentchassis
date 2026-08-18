# SUMMARY — 2026-08-18 · design-repair completion verification (`bugs_open/302`, and `201` closed)

Written to be read aloud.

## What we're trying to do

Stop the platform marking repair jobs "done" when nothing was repaired. There are two checks that run
before a job is stamped complete: one asks the repair agent "did you change anything?", the other
re-runs the original detector to ask "is the defect actually gone?". The point of this work was a
specific hole in the first of those checks.

## Where we've come from

The second check already had a firm rule, and it is the owner's: made on the 8th of August, it says
that if a check cannot run at all, the job does not pass — "I could not check" must never be read as
"I checked and it is fixed". The code enforcing that even refuses itself a loophole for unreadable
input, saying in as many words that allowing one would leave a second silent path to completion.

The first check was written five days after that ruling, and it was that second path. It only applies
to job types whose owners have explicitly signed them up — and signing up means asserting, with a
measurement attached, that for this type "nothing changed" cannot mean "repaired". But if the agent's
report came back unreadable, that assertion was silently waived and the job completed anyway.

This was raised as `bugs_open/302` by another lane, which filed it, scoped it and handed it on.

## What we've done

**Measured it first, and three of the filing's claims did not survive.** The checker registry holds
thirteen job types, not eleven. Seven of the eleven bad cases were a different bug's fault — one
fixed and shipped the afternoon before, which we confirmed at the fleet: the broken shape appears 939
times before that release and **zero** times in the 1,880 jobs completed after it. And the fix the
filing recommended — proper before-and-after checkers for the design repairs — is one this estate has
already declined in writing, for reasons recorded in the guard that exists to keep such gaps on the
record: it would need a browser on the completion path, and for three of the types "fixed" is an
aesthetic opinion with nothing to re-run. One of them is filed by four different producers, which is
a mistake we have made once already.

**Found the sharper version of the defect.** Of the eleven waived completions, **five were jobs the
check had already refused one attempt earlier**. It refused, the job retried, the retry came back
unreadable, and the waiver completed it. The arm was not merely failing to grade — it was reversing
the gate's own refusal.

**Fixed the arm, not the family.** A signed-up type must now *declare* what an unreadable report means
for it. Leaving that blank is a build failure, so nobody can inherit a waiver they never chose; and if
it is blank anyway the code refuses to block, because for a blank the dangerous direction is blocking
things nobody meant to block. The refusal gets its own status and its own operator message, because
the function that writes those messages has a history of reporting findings no check ever made.

**Proved it by breaking it.** Six deliberate mutations, each of which makes a named test fail. The
instructive one: delete the new message and the code does not error — it falls through to "the defect
is still present", handing an operator a finding no gate made.

**Put it through the council, which approved it first time and earned its keep.** Ten reviewers, two
advisory objections, none serious — and one of them caught a false sentence in our own submission: we
claimed a register entry had been amended in the same commit, and we had amended a different entry
that merely mentions it. Another objection asked us to size the blast radius with a query instead of
asserting it, and doing so found an effect nobody had enumerated: a live scheduler that decides which
job types are worth promoting reads this pair as 86.7% successful, which is an artefact of exactly the
false greens we just removed. It will become less flattering and more true — 75 refusals from now.

**Closed `bugs_open/201` on the way.** It had nothing left to fix, and we verified that at the data
rather than trusting its own summary: its checker has refused fifteen completions where the defect
survived and certified one genuine repair, which is the shape that matters, because a checker that
only ever passes tells you nothing.

## Where we are now

The fix is committed, council-approved, registered, and **not deployed** — releases are whole-fleet
and the owner runs them, so it rides the next roll. When it does roll it will do **nothing**, and we
would rather say so than dress it up: the job type it protects has had no traffic since the 17th and
both schedulers that feed it are switched off. So the honest status after the roll is "installed, not
yet seen working", carried by a test that drives the real completion path rather than by a live job.

The cost when it does fire is real: a refused job retries three times and then waits for a human. It
is not tidied up afterwards — the mechanism that would tidy it up is built and has never once run.

Three wrong calls were made in this session and all three were caught before they changed anything —
one by us, one by a peer lane, one by a reviewer. All three were the same error in different clothes:
a thing that exists, or was written down, is not a thing that operates, or was done.

## Where we're going

Nothing here is queued work for this lane; these are decisions somebody else owns.

- **A decision, not a fix:** two of the four jobs this bug was raised on are types where handing back
  an analysis may legitimately *be* the deliverable. Somebody has to say so before a rule can be
  written honestly.
- **A latent hole, deliberately left:** the fifteen-minute timeout sweep can complete one of these
  jobs without either check running, because its exclusion list is wired to the *other* check's list
  and structurally cannot see this one. It has never happened in 594 jobs across the full history, so
  we wrote it up rather than widening a shared scheduler rule on a hunch.
- **A watch item the architecture reviewer asked for:** the second job type to declare a refusal turns
  this from a point fix into a shared policy, and this round's approval will not cover it.
- **An operational gap that is probably not ours:** the release valve for refused jobs cannot work
  while the design audit's scheduler is off, which is a cost decision and belongs with the rotation
  work in `bugs_open/230`.
