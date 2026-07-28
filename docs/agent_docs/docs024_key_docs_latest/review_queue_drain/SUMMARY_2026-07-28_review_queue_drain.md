# SUMMARY 2026-07-28 — review queue drain

*The first summary for this workstream. Written the day the admin surface was
verified ready for the owner, at his request for a point-in-time read-out.*

## What we're trying to do

The platform constantly inspects its own sites and files what it finds — a
missing price list, a button that goes nowhere, content that failed a quality
check. Some of those findings a machine can fix; some genuinely need a person
to answer a question; and for months, both kinds have been accumulating in
queues that nobody read. We are trying to get to a world where every finding
either gets fixed by the framework, or reaches a person who can actually answer
it, or is refused at the point of writing because it names no reader. The
immediate, concrete goal: the owner can sit down at a screen, see the roughly
fifty questions only he can answer — some waiting since March — and answer
them, with each answer actually flowing back into the site it came from.

## Where we've come from

This started as "the human review queue has no working surface." That turned
out to be only the first layer. The surface was fixed in late July — the
dashboard had been silently showing at most fifty recent rows and reporting the
backlog as empty. Then we found the deeper problem: hundreds of parked items
describe pages that have since been rebuilt, and nothing distinguishes a
finding that is still true from one that stopped being true in April. So we
built an automatic re-checker that can close items whose complaint is provably
no longer true of the live page — it is written, reviewed and committed, and
deliberately ships switched off until we have read its first report.

Along the way the owner changed the shape of the job twice, both times for the
better. First: the machine-fixable failures should not be tidied into a queue
at all — the framework should fix them, so the queue should not fill. Second,
this week, shown the classification of everything unreadable: "I will take all
of them" — he makes every decision himself, working through the dashboard, and
asked for that surface to be checked before he starts.

There have also been two instructive wrong turns, both caught and logged. A
safety claim about the re-checker was written confidently and half of it did
not hold — the review council caught it. And yesterday's handoff claimed the
fifty owner-questions were invisible to the dashboard, parked under a status
the screen cannot show. Today's first measurement disproved that: they were on
the default screen all along. The lesson written into the wrong-calls log both
times is the same one — a confident sentence with no query behind it reads
exactly like a measured one.

## What we've done

As of today the admin surface is verified end to end, short of one step. The
dashboard loads, the login screen talks to the auth service, the API answers
behind it, and the running code was checked against what the pod actually
serves rather than a version label. The fifty owner-questions are findable in
two clicks — the Needs Review filter, then the type dropdown — and the form
for the biggest class is real working code: the owner types the missing
information, presses Save & Rebuild, the answer is written into the site's
own specification, and a rebuild of that page is queued automatically. The
one thing not exercised is his actual login, because the credentials are his.
The wrong claim in the handoff has been corrected in place, with the evidence,
so no future thread inherits it.

## Where we are now

The human queue holds 327 items; a second, machine-facing queue holds 157
(those mostly have live handlers and belong to a separate bug about a disabled
promoter — a different thread owns it). Of the 327, the fifty that need the
owner are ready for him now: a port-forward and a login is the entire journey.
Nothing blocks him except sitting down. His other decisions — turning the 186
advisory items into a periodic report, and refusing to write any finding whose
reader is "nobody" — are taken in principle but not yet executed, and today's
work deliberately did not pre-empt them. The automatic re-checker remains
switched off, waiting for someone to read its first dry-run report.

## Where we're going

First, the owner opens one item and answers it — that proves the login, the
last unverified step, and starts the fifty moving. Then, in order: build the
periodic report for the 186 advisory items, with a named reader and a cadence,
rather than a second queue nobody drains; and make the structural change that
closes the door — an item type must declare who reads it, and "nobody" must be
a refusable answer at the point of writing. That last one is the only change
here that prevents the next four-and-a-half-month-old unanswered question,
and it is a write-path change, not a dashboard one. Behind those sits the
larger job the owner has already named: giving the page builder a way to
rewrite content in response to what its own checker complained about, so
machine failures stop parking in a human's queue at all.
