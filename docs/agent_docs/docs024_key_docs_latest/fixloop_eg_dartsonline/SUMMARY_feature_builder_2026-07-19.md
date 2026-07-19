# The feature builder — the story so far

*2026-07-19. Written to be read aloud. Supersedes the 2026-07-18 summary,
which was written before the machine had ever been approved by its council.*

## What this subproject is

The platform already had a repair loop. Something notices a fault, diagnoses
it against the real code and the live database, drafts a small careful fix
plan, has that plan argued over by a council of specialist reviewers, and —
only if the council approves — writes the code on a caged branch, proves it
still builds, and opens a pull request that a person must merge. Nothing in
that chain merges itself.

This subproject asked whether the same machinery could *build* rather than
just *mend*. A feature is a different shape of problem from a repair: it needs
several steps, files that don't exist yet, a workflow and the code it calls,
and a database change to switch the whole thing on.

Our answer was to keep every safeguard and change only the shape of the plan.
Instead of one small edit plan, the builder works from a *staged* plan — an
ordered list of steps, each one a small constrained plan with its own files,
its own goal, its own build check. One commit per step. One pull request at
the end. Anything that would touch the live database ships as a file in that
pull request, never executed by the machine, with an ordered checklist for a
human: deploy the code first, then apply the change.

## Where we've been

We designed it, had the design signed off decision by decision, and built it
in two halves. The first half turns an approved specification into a staged
plan and walks it past the council. The second half would take an approved
plan and build it, stage by stage, into a single pull request.

Then we ran it — five times over two days — and every single run taught us
something we could only have learned by running it.

The first run refused its own plan: the designer had invented a step to
"activate" a database change that activates itself. The second run refused
again, and this time **the machine was right and our rule was wrong** — it
produced an honest checklist, and our validator demanded a code deployment for
a feature that ships no code. We fixed the rule, not the machine.

The third run got the full council sitting for the first time, and all five
reviewers found something real. The fourth ran the entire loop to exhaustion
and escalated to a human — correctly, we thought, because the reviewers kept
objecting.

Then another team found the bug that changed that story. Because of a subtle
mismatch in how data is passed into a prompt, **the reviser had never seen a
single objection**. It had been revising blind, three rounds running, while
appearing merely stubborn. What disguised it was that the *facts* it looked up
were reaching it perfectly well; only the criticism was silently blank. We
fixed it, and we can now prove from the logs that the reviser reads the
council properly.

## What we've done

The fifth run, with a reviser that could finally hear its critics, came back
**approved — unanimously, all five seats, no objections**. The plan it
produced was genuinely good: a single careful database change, guarded so that
it refuses to run if anything about the target looks unexpected, touching only
the specific settings it means to touch and leaving everything else untouched
by construction. That last property is not an accident — the council had
rejected the dangerous version of exactly that change the day before, and the
lesson stuck.

There is an honest twist. While our machine was designing that fix, another
team fixed the same problem by hand — finishing sixty seconds before our run
started. So the plan our builder produced, though approved, must not be
applied: it would undo their work. We closed it out and recorded why. The
pilot proved the builder can produce an approvable plan; it did not, in the
end, produce a change we needed.

We also did something we had been neglecting: we put our own code through the
council that reviews everyone else's platform changes. It came back asking for
revisions, and it was right to. It found a genuine, serious flaw that our own
tests had not: in three places, our code quietly fell back to a default when a
value it expected was missing. In practice that meant a build step could have
read from the wrong branch, or committed a stage's work to the wrong branch
entirely — silently, with nothing failing. That is fixed now, and it was found
before the code had ever run in anger.

## Where we are

The designing half of the machine is live, exercised, and has earned an
approval from a five-seat council. The building half is written, reviewed, and
corrected — but has still never run. Not once. That is the honest headline.

Every consequential act remains behind a person: approving a specification,
spending the credits on each run, merging any pull request, applying any
database change.

## Where we're going

Three things, in order. First, pick a fresh target — a small, real capability
this platform actually wants, since the original pilot was overtaken. Second,
fire the building half for the first time and watch it take an approved plan
through one commit per stage into a single pull request. Third, review that
pull request as a human being would review anyone's, and decide whether it
merges.

If that lands, the platform will have crossed a line worth pausing on. Not
merely a system that mends itself when something breaks, but one that can be
asked for something it does not yet have and can build it — carefully,
accountably, with a council that has now repeatedly proved it catches real
problems rather than nodding along, and with a human hand on every gate that
matters.
