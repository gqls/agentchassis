# The feature builder — where we are, in plain words

*2026-07-17. A companion to the technical docs, written to be read aloud.*

## Where this came from

Over the last week this platform grew something unusual: a repair loop that
can notice a fault, diagnose it against the real code and the real database,
draft a careful little fix plan, have that plan argued over by a review
council, and — only if the council approves — write the code on a caged
branch, prove it still builds, and open a pull request that a human must
merge. Nothing in that chain merges itself. It has already caught and
diagnosed its first real bug.

Then came the obvious question: if it can fix one small thing, can it *build*
something? A real feature needs several steps — new files, a new workflow,
the wiring between them — and that's a different shape of problem from a
one-line repair.

## What we decided

The answer was to keep everything that made the fix loop trustworthy and
change only the shape of the plan. Instead of one small edit plan, the
builder works from a *staged* plan: an ordered list of steps, where each step
is itself a small, constrained plan — these files, this goal, this test. The
same council reviews it. The same cage applies when it's built: one step, one
commit, one build check, and if anything goes red the work stops and waits
for a person.

Two details matter most. First, anything that would touch the live database —
the seeds that switch a feature on — is never executed by the machine. It
ships as a *file* in the pull request, with an ordered checklist for a human:
deploy the code first, then apply the seed. The builder physically cannot
express that checklist in the wrong order. Second, a new review seat joined
the council today whose only job is asking "doesn't the platform already have
one of these?" — which is exactly the discipline a feature builder needs most.

## Where we are right now

All of it is built and committed, as of this afternoon. The plan format and
its validator. The designer, which turns an owner-approved specification into
a staged plan and walks it past the four-seat council. The implementer, which
walks an approved plan stage by stage — its own branch, one commit per stage,
a build check per stage, a test run at the end, one pull request. The scripts
that fire each piece.

None of it is switched on. The code needs a new image deployed, and the three
seeds are sitting in the repository as files, waiting for the owner to apply
them — deliberately, in that order, because that ordering *is* the
discipline, applied to the builder itself.

## Where we're going

The first feature the builder will build is a small repair to its own
sibling: the fix loop has a long-standing quirk where two settings were left
pointing at a stale branch. It's small, it has a known right answer we can
grade against, and fixing it makes the whole family safer. The owner writes
the spec, approves it by name, fires the designer, judges its plan against
our hand-written reference, and — if the council and the owner both say yes —
the machine builds it and opens the pull request.

If that goes well, the platform will have crossed a line worth pausing on:
not just a system that can mend itself when something breaks, but one that
can carefully, accountably, grow — with a human hand on every gate.
