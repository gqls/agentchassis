# SUMMARY — work-item completion integrity

*Second in the series. The first (`SUMMARY_2026-07-18`) is what we believed when the fix
was written but not yet shipped; it stands unedited, including the parts this one
supersedes. Read them in order to see how the understanding moved.*

## What we're trying to do

Make sure that when the platform says a piece of work is finished, it actually happened.
The improvement loop is only as good as its record of what it has already fixed — if that
record can lie, the loop stops looking at real defects and its progress becomes fictional.

## Where we've come from

A bug filed from the robot-hands thread (`017`) noticed work items completing while
storing the error that proved nothing had run. Its diagnosis was wrong about the cause
and 27× short on the scale; the real confusion was two `status` fields one layer apart —
one recording that a reply *arrived*, the other what the reply *said*. That fix is now
live and the case is closed.

The second half of the story is the one that matters more. Closing `017` exposed the
layer above it: a completion verifier exists, and it had been registered exactly **once**
for 69 item types. So `017` stops a handler that reports failure from being marked
complete, but almost nothing stops one that reports *success* having done nothing.

## What we've done

`017` shipped in v1.0.1139 and was verified against the running pod — with the important
caveat that the obvious check passed on nothing, because the string we grepped for
predated the fix. Proof needed a symbol that could not exist unless the change shipped.
That lesson is now in the debugging guide, because the misleading version of that check
is the one the project's own instructions tell every thread to run.

On the verifier gap we found the filed diagnosis was again incomplete, and in the same
shape: it blamed the mechanism being opt-in — "stays at one unless an author remembers" —
which points the fix at discipline. In fact the verifier contract could not *express* a
verifier for most item types. It received only the defect's spec, and of 5,514 live work
items just nine carry a site id there. So for any site-wide defect a verifier was
unwritable, however willing the author — including the one a council-reviewed plan had
proposed adding. We widened the contract, which is the actual unblocking change, and
built a guard that now fails the build if any of the 69 item types is neither verified
nor classified with a reason.

And we wrote a verifier for the biggest population on the platform — 40% of all
completions — tested it, and **did not ship it**. Its handler only repairs a specific set
of components; everything else is deliberately left for a second detection pass to
escalate to a human. Our verifier judged the whole page, so it would have marked
correctly-handled work as unresolved and buried it in failures, across 1,849 items. The
tests all passed, because they tested the rule we had chosen rather than the one the
handler implements.

## Where we are now

`017` is closed and live. The verifier gap is worked but still open: verifiers are now
*writable*, and the gap is a categorised decision that breaks the build instead of an
invisible default — but the number of actual verifiers is unchanged at one, deliberately.
The contract widening and the guard are committed and inert until the next image ships.

Two process changes came out of this and are now fleet-wide rather than ours: a ledger of
wrong calls with a tally of the cheap check that would have caught each one, and a rule to
mark unverified claims inline so an inference cannot pass as a finding.

## Where we're going

Next is the provenance column the owner assigned to this thread — recording which
auditor's judgement created a work item, so that roughly fifteen thousand machine
judgements a month stop being unattributable. It is three council rounds in and genuinely
unstarted.

After that, real verifiers, written under a rule this thread paid to learn: read what the
*handler* is responsible for before deciding a defect can be verified. The guard's list of
gaps is the place to start, but its categories are marked as inferred, because assuming
otherwise is precisely the mistake that cost us a build.
