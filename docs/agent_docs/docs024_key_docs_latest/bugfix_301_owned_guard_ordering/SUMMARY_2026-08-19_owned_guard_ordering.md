# SUMMARY — 2026-08-19 — bug 301 closed: the owned-page guard now runs before the expensive work

*The lane's first and, unless the bug returns, last milestone read-out. Written to be said aloud.*

## What we're trying to do

Stop the page-building pipeline from paying for work it is about to throw away. Some pages on
our sites are "owned" — built and maintained by the tool pipeline, not by the generic page
builder. The generic builder is correctly forbidden from touching them. The problem was not the
rule; it was *when* the rule was applied.

## Where we've come from

On 2026-08-18 the lane that had just made these refusals visible (bug 295) noticed that every
refusal came at the very end of the workflow: the builder had already run the AI writer and the
link resolver — the two costly steps — and only then was told "not yours". On one site, 39 full
chains were run and discarded in two and a half hours. The page's ownership is on a row the
workflow loads at step 2 and did not consult until step 12.

## What we've done

We moved the same check to step 2, as an opt-in on that step, turned on for exactly the one
workflow that needs it (the tool pipeline shares the step and must never refuse owned pages, so it
stays off there and a guard in the migration aborts if anyone else turns it on). The late check
stays as a backstop. The change went through the reviewer council: round one asked for a
revision (we had skipped the house rule of snapshotting live configuration before changing it —
fair, conceded, remediated and written up in the shared mistakes log); round two approved, and
each of its advisory remarks was answered by going and checking rather than by arguing.

Then we watched real traffic rather than inducing anything: four owned pages were refused at the
load step with no AI writing run for any of them; two ordinary pages built all the way through —
writer, save, publish; refusals at the old late position since the roll: zero. Every pod that can
run this workflow — 22 of them — is on the same build, down to the image fingerprint.

## Where we are now

Closed. The bug file has moved to the closed folder with the evidence in a dated section; the
debugging guide has an index row and a new transferable pattern ("check WHERE a guard sits, not
only THAT it exists"); the memory archive has a line.

## Where we're going

One decision sits with the owner and has been deliberately not taken by this lane. The fix stops
the *waste*, not the *cause*: a dozen places in the code still send content-rewrite jobs to the
generic builder without first checking whether the page is owned, so owned pages keep queuing jobs
that can only be refused — 142 tonight. That upstream defect is now the untaken footnote of two
closed bug files and the subject of no open one. Recommendation: file it as its own small open bug,
cross-referenced to the Tier 2 repair discussion, because "routed to the wrong handler" is a small
routing fix and "how should these pages be repaired" is a design question — different bugs.
