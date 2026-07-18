# The council gate — a summary to read aloud

*2026-07-18. Supersedes the 2026-07-17 summary, which was written before the
gate went live. Technical state: `RUNBOOK_council_gate.md`. Turn-by-turn
record: `NOTES_running_council_gate.md`.*

---

## What we set out to achieve

The fix loop has a council at its heart. When it proposes a fix, a panel of
reviewers examines the plan before anything is built: one checks the edits are
real and minimal, one checks the fix against the platform's own history of
recurring bugs, one guards the rest of the system from collateral damage and
can block outright. What comes out is decided by plain rules, not by another
opinion about the opinions.

The aim of this thread was simple to say and easy to get wrong: that council
should not be reserved for the fix loop's own fixes. Many working sessions
change platform code every day, and none of those changes were getting a
second pair of eyes. So we set out to open the council up as a service that
any session can put a change through before committing it — and to do it
without inventing a parallel system, without letting it become a bottleneck,
and without pretending it was compulsory when it isn't.

## What we've done

**We built it, and it is live.** Three parts. A submission script, where a
session describes its change and, importantly, why — the reasoning is what the
reviewers judge the edits against. The council service itself, which reuses
the existing machinery entirely: no new platform code was written for it, so
switching it on was a single database step. And a coverage report, which walks
recent history and says, commit by commit, which changes carried a verdict and
which never saw review.

That report gave us our baseline on the first run: in three days, twenty-eight
changes to platform code, none of them reviewed. That is the number this
exists to move.

**We proved it works on a real change, not a toy.** The first submission was a
genuine improvement to the fix loop's own digest. The council took about two
minutes, five of the seats then live examined it, two correctly sat the round
out as irrelevant, and the verdict came back "revise" with four fair
objections. Then the part that justified the whole endeavour: the reviewers
are allowed to ask questions of the live database before deciding, and their
questions found that my plan rested on an assumption that was simply false.
Had it been written as proposed, it would have produced a permanently empty
section in the digest — no error, no warning, just quietly nothing. That is
exactly the family of failure the platform has hit again and again, and the
council caught it on its first real outing, before a line of code existed.

**We kept it in step with a council that is growing fast.** In parallel,
another thread has been steadily adding reviewers drawn from the register of
verified platform concepts. Each time, the gate has to be re-synced or the two
councils drift apart — which is itself the kind of failure these reviewers
exist to catch. It grew from three seats to five, then seven, and this morning
to nine. Each time we re-read the live definition from the database rather
than trusting our own files, and mirrored what was actually there.

## Where we are now

The gate is live and commissioned, running the full nine-seat council. Two
seats always sit: edit quality, and the guardian who can block. The other
seven — the bug historian, the reuse agent, the guidelines agent, the tooling
and provenance reviewer, and guardians for the adoption, diagnosis and
improvement pipelines — only wake up when a change actually touches their
territory. That relevance filter is what makes this affordable at the pace of
everyday work: you pay for a specialist's opinion when you need it, not every
time.

And as of today there is a note in the repository's standing instructions —
the file every session reads at startup — telling threads how to submit a
change, what the verdicts mean, what to do with an approval, and the one
discipline that matters when the council changes: patch both councils
together, and check the live definition first.

## Where we're going

Three things, and the first one is not ours to force.

**Adoption.** The gate is advisory by design. It cannot intercept a commit,
and shouldn't yet. What happens next is that sessions start using it, and the
coverage report tells us honestly whether they do. That number — currently
zero out of twenty-eight — is the real measure of whether this was worth
building.

**More seats, cheaply.** Four candidates remain on the register's list. Now
that the filter exists, each new one costs almost nothing when it isn't
relevant, so the roster can keep growing without slowing anyone down.

**Then the real question, which is yours.** Whether to move from advisory to
enforcement: platform changes riding branches, and only council-approved work
merging. That would change how every session works, so it waits for evidence
from advisory mode and for your explicit decision. Nothing in what we've built
presumes the answer.

The short version: the council is no longer the fix loop's private machinery.
It's a service any thread can use, it's live, it has already caught a real bug
in a real plan before it was written, and the only thing left is for people to
start using it.
