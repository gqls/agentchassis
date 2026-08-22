# SUMMARY — 2026-08-22b: closed, live, and proven by a real release

Written to be read aloud. Supersedes nothing — `SUMMARY_2026-08-22_gate_inverted.md` is the
read-out from this morning, when the gate was built but had never met a release. This is
what changed.

## What we're trying to do

Make it impossible for one of our own services to be quietly left behind by a release —
running perfectly, on months-old code, with nothing anywhere saying so.

## Where we've come from

A service was found in August sitting eighty-six versions behind the fleet, serving a known
credential leak, for months. The team that fixed it built a gate to catch the next one, and
recorded honestly in their own notes that the gate could never have found the six *other*
frozen services they went on to discover — because of the way it asked its question. That
admission became this bug.

It asked "is this service's image one the release builds?" and skipped the service when the
answer was no — which is precisely what being left behind means. So the one case it existed
to catch was the one case it stepped over, and it printed "OK" while doing it. Eight
services fell in. Two of them after the ruling meant to close it, by people who had that
ruling in front of them.

## What we've done

**Found the next release already broken, and fixed it structurally.** Three services were on
the list of images a release ships and on none of the lists that say how to build them. The
release would have built twenty-two images, spent six minutes doing it, and stopped dead
uploading one that had never been built — before the deploy step, so nothing would have
reached the cluster. The build list is now *derived* from the ship list. There is one list.

**Inverted the gate.** An image of ours that no release builds is now a failure, and the only
way out is to say so explicitly, naming what moves that service instead. The code moved into
Go — not for tidiness, but because our review council does not read the makefile, so the old
gate had never been reviewed by anyone and could not be tested without editing a file forty
sessions share.

**Added the two things either side of it.** A warning at the moment the mistake is made,
measured against five weeks of history: it fires on seven commits and all seven are the known
incidents, silent on the other fourteen. And a cap on the excused list, because closing one
hiding place opens another.

**Built a second check that reads the cluster rather than the repository** — because what is
written down and what is running are two different lists and neither contains the other. It
found five real things on its first run, and it found a mistake in itself: it described two
services as *old* when they were newer. The council then found the same mistake a third
time, in the one place it does most damage.

## Where we are now

**Closed.** A whole-fleet release ran this afternoon under the new gate: twenty deployments
and eleven scheduled jobs, all on one version, `v1.0.1326`.

The cluster check went from five findings to one, and every one of the five resolved exactly
as it said it would. Two services that were declared everywhere and running nowhere were
**created by that release, two seconds apart**, which we had predicted in writing beforehand
and which could have come out otherwise. The one remaining finding is a service another team
created *since* — caught the day it appeared rather than in three months.

Being precise about what that proves: the tree was correct, so the gate **could only have
passed**. The release shows the new logic does not break a real release; what shows it can
actually catch something is the set of deliberate breakages we ran against a copy of the
makefile. Exercised, not proven — the same distinction the previous gate's own entry had to
be corrected for overstating.

Three things are deliberately unfinished and are debts to nobody, each written where someone
will trip over it: the cluster check runs only when a person runs it; a clean result means
every service is on the right *version*, not that it is running the right *code*; and there
are now seven places in the codebase that connect to Kubernetes with no shared helper.

## Where we're going

Nothing on this bug. It is closed and the lane's files say so.

The one thing worth a decision, and it is small: the cluster check has no driver. Making it a
daily job needs permissions, a place to write its findings, and its own review round. It is
worth doing — but this estate's own history is full of detectors that worked perfectly and
were never actually driven by anything, so it should be somebody's deliberate piece of work
rather than an afterthought attached to this one.
