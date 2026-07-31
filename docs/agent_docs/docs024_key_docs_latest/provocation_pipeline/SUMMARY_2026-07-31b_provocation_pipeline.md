# SUMMARY (2026-07-31b) — the daily provocation

Written to be read aloud. Current state only; the chronology is in
`README_where_we_are.md` and the technical log in `NOTES_provocation_pipeline.md`.

Second summary of the day. It exists because the route changed, not because time
passed: what the earlier one called a small configuration step turns out to be a
piece of software we have not written.

---

## What we're trying to do

vonc.com is built on one promise: every day, one provocation, and you argue it
against an AI on a twenty-minute clock. The provocation is the product — it is what
brings someone back tomorrow, and it is the thing you would share.

We want that promise to be true, and we want the provocations to be good: safe,
genuinely arguable, current, and interesting to people who like arguing. The
direction is that Grok generates them from what is actually being argued about on X,
behind a filter that catches slop and danger, with categories eventually spanning
politics to pets, and no human approving each one.

## Where we've come from

The provocation had never rotated, and the cause was worse than a broken job: there
was no mechanism at all. The feed was a static file with the provocation written
into it as a fixed paragraph, edited by hand six times since June. A plan from 25
June had designed the whole thing as a copy of the news pipeline; the half that
displays content shipped, and the half that chooses and regenerates it was never
started.

Earlier today we made rotation *possible*. The builder became schedule-driven —
each provocation carries the date it goes live, today's is whichever has most
recently arrived, and the archive is everything before it, which turns the archive
rule into a property of the data rather than a step someone has to remember.
Publishing and reverting became one tested command. That went live and nothing
broke.

## What we've done since

**Confirmed the site is coherent again.** A second team has been sealing today's
provocation so it is readable only once you enter the round. For a few hours this
afternoon the site was half-changed — the lobby said "sealed" while the page above
it printed the provocation in full. That is resolved: their display code is live,
today's headline and body appear nowhere on the home page, and the past-provocation
sample they added renders correctly. We had flagged that last point as unproven
because our earlier check matched text that old code could have produced too; it is
now confirmed against a label only the new code emits.

**Found that our own stated next step could not have worked.** The handoff written
this morning said to write a few provocations by hand and add a scheduled job to
rebuild and republish daily. There is no job to add. Scheduled jobs on this platform
dispatch to a named agent; no provocation agent exists, no deployed code knows how
to build this file, and the schedule lives in a script inside the documentation
folder that the cluster cannot run and never ships. The instruction would have
produced a database row pointing at nothing.

The error is worth stating plainly because it read as entirely sensible: we had
correctly established that the news feed on other sites rebuilds on a schedule and
commits itself, then let that stand as evidence that *our* feed had a path to the
same machinery. It only ever showed that a different feed does. Both halves of the
instruction were true alone — provocations are needed, a scheduled job is needed —
and nothing joined them. One query would have caught it.

**Established what daily rotation actually costs**, which is real but mostly
copying. The provocations move from a script into the database. A piece of code
picks today's, builds the file and commits it — and a near-identical one already
runs for the vet medicine directory, sharing a git-committing helper that already
serves two other exporters. Then a small agent definition and the scheduled row.
Writing the provocations comes last, because until the rest exists they cannot be
published.

## Where we are now

**The site still does not rotate, and "every day, one provocation" is still a false
claim.** That has not moved today and should not be buried under the rest.

What has changed is our understanding of the distance to it. This morning the
remaining work looked like content plus a configuration row. It is content plus a
new piece of platform software, which means it goes through code review before it
ships. Not large, and well precedented, but a different kind of task than the one
the earlier summary described.

We also now know that one attractive shortcut is closed. There was a much cheaper
design: publish a fortnight of provocations at once and let the site pick today's by
date — no scheduled job, no new code, and rotation that could not quietly stop
working, which any nightly job can. The seal rules it out, because publishing ahead
puts every future provocation in a file anyone can read, when the whole point is
that even today's is hidden until you step in. **The sealing decision is what makes
the nightly job mandatory.** That is a genuine cost of a good decision rather than a
fault in it, and it was invisible from inside either piece of work.

## Where we're going

Next is the mechanism, in the order the dependencies force: provocations into the
database, then the code that builds and commits the file, then the agent and the
schedule, then the provocations themselves. The one trap already written down is
that this new code must carry across the *sealing* rules as well as the rotation
ones — our current checks enforce both, and a rebuild that ports only rotation would
silently reopen the leak the other team closed today.

There is a faster alternative worth weighing: publish a bridge of hand-written
provocations manually for a week or so, which makes the claim true within a day at
the cost of somebody running a command each morning. It buys honesty on the front
page while the real mechanism is built, and it does not change any of the work
above.

After the mechanism comes the Grok generator and the filter in front of it. Because
no human approves each provocation, that filter is the only control, so it must fail
closed — keep yesterday rather than publish something unjudged — treat its own
errors as rejections, log what it rejects, and be calibrated against our existing
nine before it is wired to anything that publishes.

Then categories, which are not just labels: each needs its own safety threshold and
audience, and more than one live provocation at a time does not fit the engine's
current one-per-site assumption, so that is a conversation with the Gauntlet team
before it is a task.

The paired mode is prototyped and waiting on real identity, which the platform does
not have at all. It is worth doing sooner than it looks, because it is what produces
the named, returning contestants whose behaviour is the only thing that will ever
tell us which provocations are actually interesting.
