# SUMMARY 2026-08-17 — live, proven, closed; and the sibling defect it uncovered

Second read-out for this lane. Every heading answers differently from
`SUMMARY_2026-08-16`, whose "where we are now" said the fix was approved but **not
running** and the damaged rows deliberately **not** repaired. Both of those have
changed, so this is an inflection rather than a clock tick.

## What we're trying to do

Stop the platform recording its own correct observations as breakdowns. Some findings
name nobody to fix them **on purpose** — nothing can automatically repaint a brand,
restart a customer's virtual machine, or repoint a broken image reference — and those
were being swept into the work queue, handed to a dispatcher with nobody to give them
to, and stamped "blocked — cannot be routed to any agent".

## Where we've come from

Filed as one thing and it was another: the bug's own title accuses the dispatcher of
grabbing *parked* findings, and the dispatcher physically cannot see a parked one. The
damaged rows had never been parked — they were written in a state their authors
believed was inert, and it wasn't. Six separate checkers each stated that belief in a
comment at their own call site, and a rule kept as a comment in six places was wrong
in some of them.

Measured properly, the class was **sixty rows across four kinds of finding on fifteen
sites**, not eighteen of one kind. The review council sent the first submission back,
correctly: the marker I had used to attribute the damage could not identify which
checker produced a row, and a loose phrase in the write-up was hiding a sixth
producer. The corrected attribution was cleaner, and round two was approved.

## What we've done

The promoting step now asks exactly the question the dispatcher will ask a moment
later — does this name someone, and does that someone exist — with both drawing the
question from one shared piece of code so they cannot drift apart. It reports what it
holds back, so a filter that quietly does less can't pass as a quiet week.

Since yesterday, three things landed and none rests on an assumption:

- **The build is live and carries it.** Proven per service, not per fleet: the running
  images' own labels, plus a proof that the fix's commit is an ancestor of what was
  built, plus a direct interrogation of both running processes with a deliberate
  nonsense check alongside to show the question could have come back "no".
- **The guard was exercised, not merely deployed.** A site holding thirty-six of these
  flag-only findings and nothing else routable was chosen deliberately, and the
  promoting step fired at it: it held back all thirty-six and promoted none. Under the
  previous build those thirty-six would have been queued and then marked as failures.
- **All sixty rows are repaired and the last hole is shut.** Each row went back to the
  state its own producer files today, and is stamped with what happened to it so none
  looks spontaneously fixed. A database-level rule now makes the bad combination
  impossible to write at all — added in the correct order (after the build, never
  before) and tested by trying to break it three ways.

The bug is closed. Most of that final work was done by another session working the
same bug concurrently, which is how this estate runs; where our work overlapped, their
version is the one of record and ours is marked superseded.

## Where we are now

Closed and verified, with two things standing.

**A sibling defect is live and growing.** The same safety check has two arms: one for
a finding with nobody named, one for a finding addressed to an agent that does not
exist. We fixed the first. The second is happening now — the tool auditor is filing
genuinely useful findings about live tools and addressing them to something that has
never existed, fourteen of them across two sites, five yesterday and fourteen today,
each recorded as a routing failure and each silently blocking the auditor from
reporting that finding again. Filed as bug 291, producer identified. Neither our fix
nor the new database rule catches it, because the handler is *named* rather than
blank.

**And a decision is waiting.** Two reviewers wanted opposite things about the same
edit: one that we unify a third copy of the shared test, one that we not touch the
shared dispatch code at all. Nothing was decided unilaterally.

## Where we're going

Nothing on this lane needs a session sitting on it. Bug 291 needs picking up — it is
the live half of this family and the cheap first move is a decision about what that
never-existent handler was ever meant to be, since the sibling checks already have a
working idiom for "a human reads this". The owner's call on the third copy of the
predicate can land whenever; either direction is a one-file change.

The lesson we are carrying forward is about verification rather than diagnosis. The
diagnosis held up under a council round and a formal challenge; the **checking** phase
produced three confident false claims in ninety minutes, every one of them treating "I
found nothing" as "there is nothing" — a grep that had been killed mid-run, a table
that only keeps two days of history, and a switched-off scheduler read as proof the
code could not run. All three are logged where the next session will meet them.
