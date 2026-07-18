# The feature builder — what we set out to do, and where it has got to

*2026-07-18. Written to be read aloud. Companion to the technical docs;
supersedes the 2026-07-17 summary, which was written before the machine had
ever run.*

## What we set out to do

The platform already had a repair loop: something notices a fault, diagnoses
it against the real code and the real database, drafts a small careful fix
plan, has that plan argued over by a review council, and — only if the council
approves — writes the code on a caged branch, proves it still builds, and
opens a pull request a human must merge. Nothing in that chain merges itself.

The question we set out to answer was whether the same machinery could *build*
rather than just *mend*. A real feature is a different shape of problem from a
one-line repair: it needs several steps, new files that don't exist yet, a
workflow and the code it calls, and a database change to switch it on.

The answer we committed to was: keep every safeguard, change only the shape of
the plan. Instead of one small edit plan, the builder works from a staged plan
— an ordered list of steps, each one itself a small constrained plan with its
own files, its own goal, its own build check. One commit per step. One pull
request at the end. And anything that would touch the live database ships as a
*file* in that pull request, never executed by the machine, with an ordered
checklist for a person: deploy the code first, then apply the change.

## What we've done

We built it, and then — the part that matters — we ran it.

The plan format and its validator went in first, then the designer that turns
an approved specification into a staged plan and walks it past the council,
then the implementer that would carry an approved plan stage by stage to a
pull request. All of it went into production yesterday and today: the code
shipped in the chassis image, the three agent definitions applied to the live
database, and a first real specification written and approved by name.

Then we fired it, three times, and each time it told us something true.

The first run refused its own plan. The designer had invented an extra step to
"activate" a database change that activates itself, and had described that step
in a way the validator rejected. We tightened the instructions.

The second run refused again — and this time **the machine was right and our
rule was wrong**. The designer produced exactly what we asked for: a single,
clean step, with an honest checklist. Our validator insisted every plan must
deploy a code image before applying a database change — but this feature ships
no code at all, so the only way to satisfy that rule would have been to write
a checklist that lied. We fixed the rule, not the machine.

The third run went furthest, and was the first time the full council ever sat.
The plan was accepted and stored. All five reviewers read it and each caught
something different and real. The one we built to remember old scars — the bug
historian — did precisely the job it was created for: it recognised in this new
plan the fingerprint of a failure this platform has suffered seven times before,
and demanded a guard against it. The reviewer whose job is to notice when a
*rule* is wrong rather than the plan approved it with a note, exactly as
designed. The council voted to revise, the system ran the reviewers' own
database queries to settle the facts they were unsure about, and the designer
rewrote its plan. The rewrite then tripped on one last unstated convention —
it invented a checklist step name we hadn't told it was a closed list.

We fixed that too. Three fires, three real defects found, every one of them
caught by the cheapest gate that could have caught it — which is exactly what
the design was for.

## Where we are now

Everything is live and working. The pilot feature — a small repair to the
repair loop's own sibling, chosen because it is small and because we know the
right answer and can therefore mark the machine's homework — has been through
design, storage, and a full council round with the reviewers' fact-checking
queries actually executed against the live database.

What has not happened yet: no plan has been *approved* by the council, so
nothing has been built, no branch has been cut, and no pull request exists.
Every consequential act still sits behind a human: approving the specification,
spending the credits on each run, merging the eventual pull request, applying
the eventual database change.

## Where we're going

The next run should carry the rewritten plan through the validator and into a
second council round, and may well reach approval. If it does, the staged plan
becomes real work: the implementer takes it stage by stage on its own branch,
proves each stage builds, runs the tests at the end, and opens a single pull
request — with the apply checklist written into the description, where the
person doing the merging will actually see it.

If that lands, the platform will have crossed a line worth pausing on. Not a
system that merely mends itself when something breaks, but one that can be
asked for something it doesn't yet have, and can carefully, accountably build
it — with a human hand on every gate that matters, and a council of specialists
that has already proved it catches real problems rather than nodding along.
