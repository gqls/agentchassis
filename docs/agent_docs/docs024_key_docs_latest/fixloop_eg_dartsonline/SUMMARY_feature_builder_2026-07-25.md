# SUMMARY — 2026-07-25 — the feature builder built a feature, and we merged it

*New file per the never-overwrite rule. Previous read-out:
`SUMMARY_feature_builder_2026-07-19.md`. Written to be read aloud.*

## What we're trying to do

We already had a machine that could find a bug and fix it. The question this
workstream asks is whether the same machine can **build something that doesn't
exist yet** — not a one-line patch, but a whole small piece of software, written
across several files, in stages, under the same discipline: a plan reviewed by a
council of critics before a line is written, every write going through the one
narrow door that is allowed to touch the repository, and a human reading and
merging the result. The point was never to remove the person from the loop. It
was to move them to the end of it, where they read a pull request instead of
typing the code.

## Where we've come from

The designing half proved itself weeks ago — five runs, each one exposing a real
flaw in our own machinery before the next, ending in a plan the council approved
unanimously. The *building* half was the problem. It was written, seeded, live,
and had never once run. Three handoffs in a row said the same sentence: the
implementer has never executed. That was the whole remaining gap, and it stayed
open partly because nobody had a good enough job to give it.

Then a different workstream did. The vonc.com debate tool needed a backend built
from nothing — a genuinely useful, self-contained little service — and rather
than write it by hand, they gave it to our builder as its first real job.

It did not go smoothly, and it was not supposed to. The designer needed six
attempts before the council would approve its plan. The implementer needed
eight. Almost every failure taught us something and almost none of them were
what they looked like: two runs that appeared to be builder defects turned out to
be our own ten-minute housekeeping job quietly deleting the message channels the
running agents were replying on, because its "is anything still working?" check
looked for a label no agent has ever worn. Another was a chassis upgrade that
never reached the agents at all, because agents that run in their own pods pin
their image version in the database and nothing was updating it. Six durable
records came out of that fortnight; four are already closed.

## What we've done

On the morning of the 25th the builder ran the approved plan straight through.
Six stages — the scaffold, the container and deployment files, the shared error
handling and cross-origin rules, the rate limits and input caps, the database
table and the endpoint that serves today's debate, and finally the two endpoints
that actually talk to the AI. Each stage committed on its own, each gated before
the next was allowed to start, the whole thing checked by a test run derived from
the plan rather than declared by the machine, and then a single pull request:
eighteen files, eight hundred and eighty lines added, nothing deleted, nothing
touched outside its own service. Every commit message carries the plan's
reference number and the line *"Human review terminal — do not merge without
review."*

The owner read it and merged it at 09:19. It is on `main` now.

## Where we are now

The feature builder is proven end to end: a stated gap, a designed plan, a
council that argued with it until it was sound, a staged build, and a merge by a
human who could see exactly what they were agreeing to. That is the thing we set
out to demonstrate, and it is demonstrated.

Three honest qualifications, because one success is not a capability. First, the
run took fourteen fires across two days to produce one clean one — most of the
failures environmental rather than architectural, but that is a cost, not a
footnote. Second, the milestone was reached on someone else's target: the build
was driven by the gauntlet workstream, and we contributed the machine, not the
job. Third, one of the shakeout bugs is still open and is the nastiest kind —
when a chassis upgrade fails to reach a spawned agent, the symptom is
indistinguishable from the agent being broken, which is precisely how we lost two
runs to it.

There is also a piece of unfinished business that is now less urgent than it
looked. Our own stage-loop code has been sitting in front of the advisory review
council since the 18th, and its last verdict asked for another revision. The one
serious thing that review found was fixed a week ago. What remains are design
questions — did we reinvent a loop that already existed, should a read-only check
go through the adapter — asked about code that has since built and shipped a real
feature. Worth answering; no longer worth hurrying.

## Where we're going

The next thing that would actually teach us something is a **second** build, on a
target we choose ourselves, to find out whether the fortnight of fixes generalised
or merely got this one job over the line. Alongside that, two smaller decisions:
whether to spend one more council round closing out the review trail on our own
code, and getting the image-version bug into the hands of whoever owns deployment,
because it is the one remaining defect that can waste a run while looking like
something else entirely.

The merged service itself now belongs elsewhere — it has to be built into an
image, deployed to the island machine rather than the cluster, given its database
table, and smoke-tested through the public address. That is the island and
gauntlet threads' work, not ours. Our job was the machine that wrote it, and this
week it worked.
