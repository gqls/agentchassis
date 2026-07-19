# Council gate — where things stand

*2026-07-19. Milestone read-out, written to be read aloud. Supersedes
`SUMMARY_council_gate_2026-07-18.md`, which was written before the gate was
adopted by other threads. Current state only — the chronology lives in
`README_where_we_are.md` (plain English) and `NOTES_running_council_gate.md`
(technical). Five parts, in order.*

---

## What we're trying to do

Give every working session a way to have its platform changes reviewed before
they land, using the reviewer council the fix loop already had, rather than
building a second thing that would drift from the first.

## Where we've come from

The council existed, but only the fix loop could use it: it reviewed the fix
loop's own proposed fixes and nothing else. Meanwhile a dozen sessions changed
platform code every day with no second opinion — and because they all share one
branch, one database and one image, a mistake in one thread is everyone's
problem. The council itself was ready for wider use, because it judges a plan by
its identifier and genuinely doesn't care who wrote it. The work was to open that
door without inventing new machinery, and without making review so slow or
expensive that people route around it.

## What we've done

We built three pieces and switched them on. A submission script, where a thread
describes its change and its reasoning. The council service itself — which
needed no new platform code at all, so turning it on was a single database step.
And a coverage report that walks recent history and says which changes were
reviewed and which weren't.

Then we spent most of the effort on the unglamorous half: keeping it honest.
The council grew from three seats to thirteen in about a day, and each new seat
had to exist in two places at once, so we replaced that hand-copying with a
mechanical mirror — the alternative was exactly the kind of quiet divergence the
council exists to catch. We fixed a fault that made the coverage report
undercount badly while appearing healthy, another that made it accuse an honest
commit of faking its review, and a third where the evidence a commit cited could
be deleted by a documented practice, retiring that practice. And we closed a bug
the owner flagged: the fix loop's reviser had been rewriting plans while
receiving blank reviews, and separately could only see six of the thirteen
reviewers — it now reads the whole council's report as one document, so future
seats reach it automatically.

## Where we are now

The gate is live and being used. Three threads have submitted without being
asked to, including one that went through a full revise-and-resubmit cycle as
intended. On its first real run it caught a genuine defect in a proposed change —
its reviewers queried the live database and disproved an assumption that would
have shipped a silently empty result.

Two honest qualifications. The reviser fix is applied and verified in
configuration but has not yet run, because the fix loop simply hasn't revised
anything since; we're watching for the first one rather than claiming it works.
And coverage remains low — most platform commits still carry no review, which is
what a day-old voluntary convention looks like.

## Where we're going

Three things. Adoption is the first and it isn't something we can force: the
note in the shared instructions file reaches every new session automatically,
and the coverage report tells us honestly whether people use it. More reviewer
seats are cheap now, because the relevance filter means an irrelevant specialist
costs nothing. And then the real question, which is the owner's: whether this
stays advisory, or becomes a gate that platform changes must pass — branches and
approved-only merges. That changes how every session works, so it waits for
evidence from the advisory period rather than being decided in advance.
