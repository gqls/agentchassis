# SUMMARY — bugfix 187, 2026-08-04 (bug closed: fixed and live)

**What we're trying to do.** Stop the platform raising "build this page" work
requests that nobody — human or machine — can ever complete, and stop the ones
that become completable from sitting in the review queue for ever unnoticed.

**Where we've come from.** Ticket 177 fixed one emitter that raised impossible
requests. Its close-out found the same disease under five more sources, with
28 dead requests parked, and the council directed that population be its own
ticket. This lane picked that ticket up yesterday evening, measured all 28
rows, and read all five sources: two were raising impossible requests exactly
like 177's case; one is a legitimate reporter whose requests must NOT be
silenced (five of its rows point at genuinely broken pages); two were one-off
manual actions with nothing to fix. The ticket also claimed the review queue
had a drain for these — reading the code showed it never did, which is half
the reason nothing ever recovered.

**What we've done.** Built one shared piece of read-only code that answers
"could the page-builder actually do anything with this page?", and asked it at
both ends of a request's life: the two blind emitters now check before
raising (declining visibly when the answer is no), and the review-queue drain
now covers this request type — closing a request only when the page it asked
for demonstrably exists, section by section, with the evidence written onto
the row. The design was approved by the review council on the first round.
It went live this morning and proved itself immediately: twenty-six parked
requests closed with evidence, six buildable-but-unbuilt ones flagged for a
human, ten kept honestly parked because closing them would assert something
untrue. A clean-up script retired twelve provably-dead rows, guarded by a
census check that aborts if the world has moved — a check that caught a real
mistake of mine before it shipped.

**Where we are now.** The bug is closed: fix live on both running replicas,
proven at the binary in both directions. The queue holds only truthful rows.
One deployment hazard surfaced on the way: an image built *after* our fix,
which did not contain it — checking the image's contents before deploying is
now written into the runbook.

**Where we're going.** Two watch items, neither blocking: the next natural
image-landing on a section-less page should produce a visible "declined" log
and no request (expected within a day or two); and four calculator pages that
sit in a site plan without any declared sections remain an open design
question for the owner (ticket TL-009) — the guard rightly refuses to guess
that one.
