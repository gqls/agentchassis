# SUMMARY — the unpublish primitive (`bugs_closed/098`) — 2026-08-06 — CLOSED

*(Milestone read-out: the closing one. Previous: `SUMMARY_2026-08-04_unpublish_primitive.md`,
written the evening the population went to zero. The series is the record; this file is
current state only.)*

## What we're trying to do

Give the platform the ability to take a page OFF a live site as reliably as it puts one
on — and make sure that when the machinery declines to remove something, a human can see
that it declined, and why. When this bug was filed, archiving a page removed it from
every internal model while the public site kept serving its last rendered copy forever,
and nothing could detect that, let alone repair it.

## Where we've come from

The capability was built, vetoed on packaging grounds, and kept under the owner's RFC 011
ruling: the destructive verb exists but is reachable only through the guarded retraction
path. Along the way we fixed the opposite bug (a scheduler silently re-publishing an
archived page twice a day), paid five correctness debts the reviews raised, and had half
of one of those fixes refuted by its own first live run — which taught us that the
platform throws away an action's in-memory findings when it parks to await a reply, a
plumbing defect that turned out to have three documented faces across the fleet.

## What we've done (since the last summary)

Finished everything and closed the bug. The final acceptance held: all ten batch-retracted
pages stayed down through both scheduled refresh windows, with zero attempts to republish
any of them — that was the test that mattered, the one whose absence would have silently
un-fixed everything a week earlier. The last code repair (debt 5b) shipped, was approved
by the council, went live, and was proven at the database layer with a zero-side-effect
probe: a retraction aimed at a live page was refused, nothing was dispatched, and both
durable audit rows — the refusal with its reason, and the full "what I considered and
why" record — were sitting in the error log where monitoring already reads. The mechanism
has now performed five retractions in total (one single, one batch of ten, and two for a
neighbouring lane whose guard correctly refused the first attempt because the pages were
still linked), every one with both acceptance halves proven.

Today the owner ruled on the deeper plumbing question, RFC 012: **option B, as a
DB-backed helper**. The escape hatch every affected action has been hand-rolling becomes
one shared, named mechanism that writes findings to the database — the only sink proven
to survive the park — rather than to any in-memory record. The bigger rewiring of the
coordinator (option A) stays available later but needs a census of readers nobody has
run; doing nothing (option C) is what had already cost a production data loss. Building
the helper is not yet assigned to anyone; the ruling is recorded in the RFC.

## Where we are now

**`bugs_open/098` is `bugs_closed/098`.** Population zero, mechanism proven, and the
open questions all have recorded answers: archiving deliberately does NOT auto-retract
(the runbook's two-step procedure is the mechanism, because an automatic file-deleter
keyed on a hand-edited flag is exactly the unguarded authority this platform keeps
having to claw back); the acceptance test is the honest two-part one, not the stamp
count the file originally proposed; and the retraction audit survives durably on every
real run, clean or refused.

## Where we're going

Nothing further on this lane. Three things live on elsewhere: the RFC 012 helper needs
a builder (unassigned — the next findings-plus-await action should reach for it and
build it properly); RFC 011's deferred general question (is a destructive verb different
in kind from an inert field?) waits until the next destructive verb arrives; and the
class detector idea — curl-sweep what the platform believes is retired — remains the
cheapest way to catch the next instance of "the model and the artefact disagree", which
is the shape this whole bug was one case of.
