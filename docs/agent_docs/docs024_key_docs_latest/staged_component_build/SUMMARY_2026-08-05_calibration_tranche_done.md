# SUMMARY — the backlog is priced: four contracts proven, three live-tested, and the pace is now an informed choice

## What we're trying to do

Every tool and every shared page-piece on the estate should carry a written, tested
contract: what it promises a visitor, proven able to fail, checked in a real browser in
the live cluster. The machinery for this was finished and proven earlier in this lane;
the owner decided on 5 August that we take on the backlog of subjects that predate it —
and the agreed first step was a small calibration batch, timed, so the pace of the rest
could be chosen deliberately rather than discovered.

## Where we've come from

The last summary (3 August) closed with "what's left is deciding how much of the fleet to
point the machinery at." The owner made that call. The backlog then measured 36 tools and
111 section components with no contract at all; by the time this batch started, six new
tools had been born WITH contracts — the platform's tool-builder now writes one at birth —
so the backlog is genuinely old stock, not a growing pile.

## What we've done

Ran the full recipe on five subjects. Four contracts now exist, each read from the real
served page first, every check watched to fail under its own deliberate break, persisted
properly, and read back out to prove the stored copy is what runs: the fuel cost
estimator (13 checks), the loan-versus-savings calculator (14 checks, including its exact
arithmetic, its accessibility badge, and its deliberate tie rule), and the estate's two
most-placed page sections — the hero banner and the call-to-action block (about 200 pages
each). Three of the four were then dispatched through the live cluster and passed
everything, the two sections with a deliberately-wrong-page control that was correctly
refused.

The first subject picked was dropped at the reading stage because its live page is
genuinely broken — the gas unit converter serves a fully-styled form with every piece of
text empty, and a sibling on another site serves raw template code. The platform's own
checks had already filed both; the tickets are parked in human review, one marked
"won't fix". That finding, and a missing logo file that 404s on every gas wholesalers
page, went to the owner rather than into contracts for broken pages.

One new failure class was found and permanently recorded: the offline proving harnesses
run the newest code, so a fence using vocabulary another session added *hours earlier*
passes every offline proof and then fails in the live cluster, whose binary predates the
word — and the failing verdict auto-opens a "fix this tool" ticket for a healthy tool.
It cost one cluster round-trip to find, the spurious ticket was cancelled with the reason
written in, both affected fences were reworked and re-proven, and the trap is now a
landmines entry synced where every agent reads.

## Where we are now

The recipe is calibrated with real numbers: a static page section costs about 15 minutes
end to end; an interactive tool 30–45 minutes; a subject with a live defect costs a
finding instead of a contract, which is the reading step doing its job. Nothing is
blocked: one live test (fuel estimator) waits only on one missing image file, stated in
its own contract. The remaining backlog is 39 tools — of which only 15 have working pages
today — and 109 sections.

## Where we're going

The pace is the owner's call, posed with a recommendation in the running log: steady
background clearance, or a focused push on the ~15 ready tools plus the dozen most-placed
sections (three to four sessions, covers nearly everything a visitor actually meets), or
exhaustive clearance (ten to fifteen sessions, diminishing returns). The broken pages
found along the way need an owner decision too — they are detected, parked, and still
serving.
