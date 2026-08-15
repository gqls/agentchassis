# SUMMARY — 2026-08-15 — the guardrails are live, the directory machinery is built

*The milestone read-out, written to be said aloud. Previous entry in the series:
`SUMMARY_2026-08-11_where_things_stand.md`. Current state only — the chronology
lives in NOTES and README_where_we_are.*

## What we're trying to do

Take the ~150 finance and insurance domains in the portfolio from parked names to
live, full-featured websites — twenty-plus pages each, with working calculators,
guides, images, newsfeeds, and provider directories built from real web research —
without hand-building any of them, and without shipping the kinds of quiet defects
the fleet has already been bitten by. The build pipeline does the work; our job has
been to make it trustworthy enough to point at a hundred and fifty domains.

## Where we've come from

A week ago the honest position was: only four of the 152 domains had ever been
built at all, every one of them by hand or by copying an existing site, and the
automated pipeline had never produced a single portfolio domain end to end. Its one
real test run, on a throwaway domain, worked but surfaced a list of gaps. The owner
set the gate plainly: nothing new goes through the pipeline until two guardrails
exist — a standing check that a deployed site is actually correct once served, and
a fix for the fact-discipline hole where a wrong fact, once written into a site's
own register, was trusted forever and defended by every checker.

## What we've done

**Both guardrails are built, reviewed, and live in production.** The
structural-validity gate is five automated checks against the live served site —
dead links, wrong canonical addresses, broken structured data, missing page
essentials, dead sitemap entries. The fact-discipline fix means a fact that cites a
code artefact can now be re-proved against that artefact on a schedule, and a fact
resting on a person's word now gets flagged for a fresh look after six months
rather than being trusted silently forever. Both went through the review council —
neither passed first time, and the review rounds caught real problems, including
one where the new checker could itself have been fooled by the very trap that
caused the original bug. Both are on the running system now, verified at the
binary on both replicas, not just at the tag.

**The review process also caught something nobody was looking for.** While
double-checking a claim at the owner's prompting, we found a four-month-old
check that fires on every site every time — its database columns have been empty
fleet-wide since the spring — and each firing orders a pointless full-site rebuild.
Roughly twenty-five of those rebuilds actually ran. Filed as bug 270 with the
evidence and fix candidates; not on our critical path, but a steady leak plugged
for whoever picks it up.

**The provider-directory machinery for the finance sites is built.** Three
directory kinds to start — mortgage lenders, savings providers, health insurers —
each a single fleet-wide list of named UK providers where every fact must carry a
verbatim quote from its source, and the source is re-fetched to prove the quote is
really there before anything is registered. Two compliance decisions are enforced
in the machinery itself, not just in instructions: the directories carry no prices,
rates or premiums ever (each kind has a closed list of permitted fact types, and a
price-shaped fact is refused at registration no matter how it was produced), and
the pages say "cited facts" rather than "verified facts", because a citation proves
where a claim came from, not that it is true. The research agent and its weekly
discovery sweeps are seeded and deliberately switched off; the page components are
written and validated against the live database without being applied. Everything
is staged behind one gate: the next image roll.

## Where we are now

The directory code is committed and in council review, which found one round of
real improvements (the price-enforcement mechanism exists because a reviewer
pointed out a ruling without enforcement is just a preference) and has since gone
two further rounds on the shape of the submission paperwork rather than the code —
including one requirement the submission format cannot actually satisfy (it caps
the list at eight edits; the change touches ten files). The gate is advisory by
design; the code is on the shared branch with its correlation recorded, and the
verdict trail — one round of substance, two of form — is honestly written down.
Whether to spend a fourth round or proceed on the advisory record is the owner's
call, and nothing else is blocked on it: the actual blocker for every remaining
step is the image roll, which nothing about the review changes.

## Where we're going

The next roll of the chassis image carries the directory code. After it is
pod-verified, the staged pieces switch on in order: apply the page components,
enable the three weekly research sweeps, run each once under supervision, and work
the human-review queues until each kind has a respectable set of cited providers.
Then the small wiring migrations — the publish trigger's kind-blindness fix, the
enrichment step that lets a freshly-classified site opt into a directory at plan
time, and the planner rule that actually places the pages. Then the piece this has
all been for: one real portfolio domain through the whole pipeline as a pilot, with
a measured cost baseline, brought to the owner for sign-off before the fleet
build-out begins in waves. Still parked deliberately: the owner's build-order call
across the ~140 remaining domains, the loanzy.uk conflict with the webdesign lane,
and the mortgagecalculator copy-voice review queued from last week.
