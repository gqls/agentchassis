# SUMMARY — gauntlet_dead_cta, 2026-08-02

Written to be read aloud. A new file, as the rule requires. The previous read-out
is `SUMMARY_2026-07-31b_gauntlet_dead_cta.md`, which covers the share card and the
public record. This one covers the other thread of this lane's work — the
duplicate-content checker — which that summary does not mention. The series is
the record.

## What we're trying to do

This started as one bug on one page: vonc.com said the same things twice. The
lane's job became fixing the class rather than the instance — a checker that can
find byte-identical repeated content anywhere in the fleet, and a repair that can
remove it safely. "Safely" turned out to be most of the work, because a deletion
that is wrong looks exactly like one that is right, and nothing brings the row
back. The end state we wanted: the machinery built, reviewed, and proven — but
switched off, because the decision to point it at real data belongs to the lane
that owns the underlying bug, not to us.

## Where we've come from

By Thursday the checker existed and had been through review twice. The first
round said revise; the second said revise again, and in answering it I found the
review was right to push: my own measurement of "what would this delete today"
was wrong. I had measured with a different rule than the one the code actually
applies, and the honest answer was that it would have deleted a row it should
not — two different components on vonc sharing the same site-wide boilerplate
text. The identity rule was rebuilt so that a row is only a duplicate of another
if it sits in the same slot with byte-identical content, not merely similar
prose. Third round: approved, twelve reviewers in favour.

That approval left two loose ends, and the owner ruled on both. First, the lane
that owns the bug had no idea the checker existed — my note to them sat unread in
their directory. Second, one reviewer's objection survived the approval: the
repair deleted rows without asking the site's plan whether the repetition was
deliberate, so a page designed to repeat a component could have its design
deleted as a defect. The owner chose to deliver the news properly and to build
the guard now rather than later.

## What we've done

Three things, each through the review gate on its own merits.

The plan guard: before deleting anything, the repair now asks the site's own
plan how many copies of that component the page is supposed to have, and it
walks the same three sources of truth, in the same order, that the site builder
itself uses — so the guard and the builder can never disagree about what the
plan says. If the plan wants two copies, two copies stay. If the plan cannot be
read at all, the repair refuses to act rather than guessing. Approved first
round.

The lock guard: the review of the plan guard asked a question nobody had asked —
does the repair honour row locks? It did not, and forty-seven locked rows exist
in production today. The repair now treats a locked row the way every other
automated writer on the platform does: it will never delete one, and it reports
what it declined to touch. Approved first round.

The delivery: the owning lane's own cold-start document now carries a dated
section saying the checker is built, approved, and inert; that turning it on
means real deletions from the first site it runs against; and that the choice is
theirs. That is where a new session in their lane starts reading, so the news
can no longer be missed.

Worth saying out loud: two of the three review runs were killed mid-flight by
routine platform deploys — a deploy restarts the reviewer along with everything
else — and had to be re-fired. That is now written down as a known cost of
reviewing on this cluster, not an anomaly. And one reviewer objection in the
final round was simply wrong about the test file it criticised; we refuted it by
quoting the file rather than arguing, which is the same standard that caught my
own wrong measurement two days earlier. The standard cuts both ways and it is
the right one.

## Where we are now

The whole chain — detector, plan guard, lock guard — is built, tested, measured
against the live fleet, and approved end to end: three approvals across five
submissions in two days. Every claim in the record now has its evidence attached.

And it is all inert, deliberately. No agent in the fleet references the check;
run against today's data it would delete nothing — the one repetition it finds
is specified by that site's plan, which is exactly what the guard exists to
protect. Nothing here ships until the next routine deploy carries it, and when
that happens there is a recorded obligation to verify the code is actually in
the running pods rather than trusting the roll.

## Where we're going

The one open question is not ours to answer: the owning lane decides whether and
where to enable the checker, and everything they need is in their own directory.
If they take it on, the first enablement should be one site, watched. Two small
follow-ups are recorded for whenever they become worth doing — the three-source
plan walk could be shared with the two other places that do the same dance, and
the lock rule could move somewhere more central if a third consumer appears.
Neither blocks anything. This thread of the lane is, for the first time since
the vonc bug was filed, finished with nothing owed but a watch.
