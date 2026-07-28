# SUMMARY — 2026-07-28: the seat can see

*A new file, not an edit of `SUMMARY_2026-07-27c`. That one was called "the seat
that cannot see"; this one exists because that stopped being true. If the five
headings below had produced roughly the last summary again, there would be no file
here.*

---

## What we're trying to do

The owner asked for a process — possibly a council seat — that keeps the
architecture robust, stops it shifting underneath us, and keeps it sufficient for
plans we can anticipate, knowing those three goals pull against each other.

The conservative half already existed: a guardian seat holding the only hard veto.
Nothing argued the forward half. So the work has been to measure whether the
imbalance was real, build the counterweight, and then — the part that turned out to
matter most — make sure the seats can actually **look things up**, rather than
guessing confidently and being corrected a round later.

## Where we've come from

We measured the imbalance and it was real: the problem is **ossification, not
churn**. The platform core barely changes while pressure on it is high, and the
difference lands as workarounds stacked above it. That justified staffing a
forward-looking seat, which is now live.

Then the owner ruled, and the design settled: one conservative seat at full remit
keeping its veto, one forward seat, no duplicates. The balance comes from the two
arguing, not from trimming either. Nothing has been waiting on him since.

His last directive reframed the work. It was not enough for a seat to be *honest*
that it could not check something — it had to be *able to check*. We found the
seats were blind in two distinct ways: their questions were not routed to anything
that could answer them, and the thing meant to answer contained only declarations —
function names and signatures — never the code inside them. A reviewer asking "does
this codebase do X anywhere?" got silence, and read silence as "no".

## What we've done

The routing was fixed first, in configuration, and went live immediately.

The harder half — putting the actual source of every function into the index — went
through the council gate, was approved on its third round, and is now live and
proven. All 4,535 indexed symbols carry their source. The worked example written
into the tool's own documentation when it was built, which had never once matched
anything, now returns six results. Nothing was corrupted on the way, and we can
show that rather than assert it.

Along the way we corrected four defects in the plan the council had **approved**,
three of which were flagged by reviewers as minor objections that turned out to be
concretely right when someone actually ran them. The most serious: the plan's
central step claimed the indexer was already reading source files. It was not — it
reads a summary containing no source at all. The change works only because a piece
of live configuration had been altered months ago in a way our own stored setup
files still do not reflect. Had anyone trusted the repository over the live system,
this would have deployed looking finished and doing nothing.

## Where we are now

The seat can see code. That is a real capability change, not a configuration
tweak, and it is the first thing on this workstream that required a build and a
deploy rather than a database update.

Two things temper it, and both are worth saying plainly.

**First, we made one thing worse while making another better.** We fixed the case
where an empty answer was indistinguishable from a question nobody ran — the search
now says "I searched 4,535 symbols and found nothing" instead of showing blank
space. That was right. But that sentence is *more confident* than what it replaced,
and it is sitting on top of an index describing the codebase as it stood on 24
July. Within an hour of going live, a real diagnosis was told with that new
confidence that a function does not exist. It does. We improved how trustworthy the
system sounds without improving whether it is right, and the combination misleads
harder than either flaw did alone. That has been written up as a general trap,
because we do not think it is a one-off.

**Second, three separate checks in this one piece of work would have passed without
checking anything** — a deploy marker that would have read "not shipped" from a
perfectly good deploy, a comparison whose deciding case cannot occur in our data,
and a verification script still testing the code it was meant to be testing after
that code changed. None would have failed loudly. All three were caught, but by
running them rather than by reading them.

## Where we're going

The next job is settled and its priority has been **inverted** by what we learned:
make the index's freshness a function of how many commits behind it is, not how
long ago it ran. That used to look like a parallel improvement. It is now a
prerequisite, because we have just made the system state its findings more
confidently, and confidence on stale data is the failure we most need to avoid.

After that, prose. Our bug files, our log of wrong calls and our design register
are all still completely invisible to every reviewer — the machinery that reviews
our plans cannot read the record of how we have previously got things wrong. The
mechanism to fix that now exists; it needs one more schema change and a decision
about what to rank.

**And one thing needs the owner, not us.** The index can only ever mirror what has
been pushed, and roughly 955 commits of work have not been. No amount of
re-indexing changes this — we re-indexed and the gap did not move by a single
commit. Until that branch is pushed, every automated review on this platform is
reasoning about a fortnight-old copy of the code while being told it is current.
