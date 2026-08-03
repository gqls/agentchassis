# Summary — what counts as a live page (2026-08-03)

Second in this lane's series. The first (`SUMMARY_2026-08-02_page_role_upsert.md`) ended
with a bug fixed and one question put to the owner. This one is written because the answer
to that question turned into something larger: the platform now has a single, tested
definition of whether a page is live, and it did not have one before.

## What we're trying to do

Stop a family of quiet failures that all come from the same mistaken idea — that you can
tell whether a web page is live by reading one column. Two shapes of it: a page-creating
step that silently flattens an existing page, and a checker that never looks at a page
because it thinks the page isn't live yet.

## Where we've come from

A week ago a single instance was fixed: an upsert that overwrote a live page's content and
left its identity wrong, found looping on one site for three months. When the reviewing
council read that fix, a seat asked whether anyone had checked the siblings. There were
four, and the note recording them sat open and unowned for a day. That note is where this
session started.

## What we've done

Three things, and the second and third were not planned.

**One seam instead of five patches.** Every page-creating step whose role is fixed by what
the step *is* now goes through one piece of code that answers "that name is already taken"
in four explicit ways: create, refresh, take over a page that has never been published, or
— if the page is live and doing a different job — **stop, change nothing, and file the
decision for a person**. Four steps converted, a mechanical check added so a fifth can't be
written silently, and the owner's ruling that the widest of those powers must be switched
on explicitly by each caller rather than granted by a comment. All of it is live and
verified in the running binary.

**A definition we already had, and were not using.** Correcting my own work, I found the
estate had a shared, tested rule for "has this page shipped" — and that I had hand-written
a worse one. Mine was wrong on eleven live rows, including three tool pages created that
same day. Worse, the shared rule couldn't easily be *used*: it didn't cope with the table
alias that nearly every query needs, so consumer after consumer had written the judgement
out by hand, and a hand-written copy is where the wrong version creeps back in. That is now
fixed at the root — the rule takes an alias, and the old form is derived from it, so the
two cannot drift.

**The checkers that were never looking.** Applying that rule fleet-wide, twenty-eight live
pages turn out to be invisible to about ten of our automated checks, because those checks
ask the wrong column. Three are now converged, two deliberately left alone because the
narrow test is right for them, and the rest held back for their own measurement. Both the
converged and the untouched decisions are backed by running each query both ways against
production and looking at what actually changes.

## Where we are now

The first bug is closed and live. The owner's rulings are implemented and live. The
follow-up bug is filed, its first tranche approved by the council and awaiting the next
build. Three council rounds ran across this work: approved, then a revise, then approved
again — and the revise was correct, on a point worth keeping: **a change whose safety
depends on having swept every call site has to show the sweep, not assert it.**

The honest note on severity, which has not changed: **none of this was on fire.** The
original defect had a real surface — four live pages sitting on names those steps would
claim — but nothing had stepped on it. The checker blindness is a false negative, not
corruption: pages went unexamined, not damaged. This was prevention throughout, and the
work is worth more for what it stops than what it repaired.

## Where we're going

Two tranches remain on the follow-up. One is a group of eight queries where the right
answer genuinely isn't obvious yet and needs a measurement nobody has taken — how long a
page actually waits between being flagged for rebuild and getting one. The other is a
single function in the planner that can inject a layout into a live page; its exposure
today is one archived row, and fixing it honestly needs a change in a hot path, so it is
recorded with its measurement rather than rushed.

Beyond that, the thing to watch is whether the new checks start reporting: two orphaned
pages and eight untested tools will appear the first time the converged checkers run. They
are real and they have been invisible for weeks — but they will look like new problems
arriving rather than old ones surfacing, and somebody reading that queue should be told
which it is.

## The part I'd want said out loud

Five separate times this session I wrote something down that was wrong in the same way:
confidently, in the same voice as the things I had checked. A detector measured against a
tree that already had its fix. A predicate asserted without asking whether one existed. A
census filtered on a column I never enumerated. A count typed from memory. A claim about
what was running in production that another session had made false an hour earlier.

Every one was caught — by the council, by a mutation that refused to fail, or by rereading
my own numbers. None was caught by being careful. That is the argument for the review
machinery being expensive and slow: it is not there to catch carelessness, which is rare
here. It is there to catch the confident, plausible, unchecked sentence, which is not.
