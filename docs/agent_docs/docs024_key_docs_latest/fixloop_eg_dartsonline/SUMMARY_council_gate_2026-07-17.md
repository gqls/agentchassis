# Where the council gate is — a summary to read aloud

*2026-07-17, the "fixloop council on every bugfix" thread. Companion to
`RUNBOOK_council_gate.md` (the technical state) and
`NOTES_running_council_gate.md` (the turn-by-turn record).*

---

## What we are doing

For a while now, the fix loop has had something quietly valuable at its
heart: a council. When the loop proposes a fix, three reviewers look at it
before anything happens — one checks the edits are real and minimal, one
checks the fix against the platform's own history of recurring bugs, and one
guards the rest of the platform from collateral damage, with the power to
block outright. The decision that comes out of them is made by plain rules,
not by another AI opinion.

The idea this thread is building: that council should not be reserved for
the fix loop's own fixes. Many different working sessions change platform
code every day, and none of those changes get a second pair of eyes. So we
are opening the council up as a service — any thread, before it commits a
change to platform code, can submit that change and get the same three-seat
review, with the verdict written durably on the record.

## Where we are now

The whole advisory version is built, and deliberately none of it is switched
on.

There are three pieces. First, a submission script: a thread describes its
change — the actual edits, and crucially the *why* — and the script checks
the submission is well-formed and in scope before a penny is spent, then
sends it to the council. Second, the council service itself, ready to apply
to the database in one step, with no new platform code needed at all —
everything it uses already exists and is already running. Third, a
visibility report: it walks the recent git history for platform code and
says, commit by commit, which changes carried a council verdict and which
never saw review.

We ran that report for the first time today. In the last three days there
were twenty-eight platform-code commits, and the number that had been
reviewed was zero. That's not a scandal — the gate didn't exist yet — but it
is the baseline number this whole effort exists to move.

Four decisions were collected from you today and are on the record: review
covers platform code only, never documents or site content; we launch in
advisory mode, where the council records verdicts but cannot block anyone;
one council run per task; and — the decision that sets the sequence — the
gate does not go live until the council has more seats.

That last decision connects this thread to the concept-register project.
Its register of sixteen hundred verified platform concepts is exactly where
new reviewer seats come from: the bug-historian seat, live since yesterday,
was built from it. And here the day moved faster than this page: while we
were building the gate, that thread — working concurrently, with your
go-ahead in its own conversation — built and applied the next two seats. A
"reuse agent," whose one job is to ask: does the platform already have
something that does this, and did the plan check? And a "guidelines agent,"
which checks a change against the platform's hard-won rules — and, unusually,
knows the difference between a change that breaks a rule and a change that
proves the rule itself is wrong. The council is now five seats.

So the seats you asked us to wait for have arrived. We synced the gate to
the five-seat roster the same afternoon, and we left a note on the record
about the one discipline this doubling creates: there are now two councils —
the fix loop's own, and this gate — and every future seat must land in both,
in the same migration, or they quietly drift apart. Fittingly, "two copies
of one thing drifting apart" is precisely the failure the reuse agent exists
to catch. We also flagged one small inherited gap in the live council for
the owning thread to fix: three of the advisory seats can ask for evidence
queries that are currently never run.

## What we will do next

Three things, in order.

First, one decision is back with you: does the five-seat council satisfy
your "more seats first" condition, so the gate can be applied? Or do you
want the relevance filter first — the mechanism the other thread has
designed, which picks only the seats relevant to each change instead of
running all of them every time? At the pace the gate will run — every task,
fleet-wide — that choice shapes both cost and speed, and it needs a small
piece of platform code, not just configuration.

Second, once you say the roster is enough, we apply the gate itself — one
database step, then a smoke test with a small real change, end to end.

Third, we let advisory mode do its work for a while. Threads submit, verdicts
accumulate, and the visibility report tells us honestly what fraction of
platform changes are getting reviewed. Only then, with numbers rather than
hopes, comes the real question: whether to flip from advisory to enforcement,
where platform changes ride branches and only council-approved work merges.
That is a change to how every session works, and it stays your call.

The short version: the review service is built, synced to a council that
grew to five seats today, and safe in its box. One question is waiting for
you — is this roster enough to launch? — and nothing goes live until you
say so.
