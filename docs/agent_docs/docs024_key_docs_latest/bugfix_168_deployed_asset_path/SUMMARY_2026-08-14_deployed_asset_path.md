# SUMMARY — 2026-08-14 — the safety gates are finished, and the useful discovery is that they had never been consulted at all

*Current state only. Chronology lives in `NOTES_deployed_asset_path.md` and `README_where_we_are.md`.
The previous read-out was 2026-08-10; this supersedes nothing — the series is the record.*

---

## What we're trying to do

The platform raises findings when a page says something the site cannot support — a statistic with no
source, a claim its own evidence register does not carry. Those findings are deliberately for a human
to judge, because deciding whether a sentence is *true* is not a machine's job.

But findings go stale. Somebody fixes the page and the finding sits in the queue for ever, because
nothing ever went back to look. So we built a daily sweep that re-examines parked findings and
withdraws the ones that genuinely no longer hold.

The whole difficulty is in one asymmetry. Leaving a finding open costs a human a glance. Closing one
wrongly means the platform has quietly decided a factual complaint about a live page was resolved when
it wasn't. So every piece of work on this lane has been about the same question: **what evidence is
good enough to let a machine close one of these?**

## Where we've come from

The answer has been tightened three times, and each tightening came from someone refusing to accept
the previous one.

We started with the obvious test: re-scan the page, and if it no longer trips the check, close the
finding. That is wrong, and a reviewer caught why. The standard a page is measured against is *itself
editable data* — add a line to the site's evidence register and a claim becomes "supported" with the
page untouched. So the register can retract a finding about a page nobody has fixed. The owner ruled
we must have positive proof the **page** moved, not just that the complaint evaporated. That was gate
one.

A reviewer then objected, three rounds running, that proving the *page* changed is not proving the
*disputed sentence* changed — a typo fix elsewhere on the page satisfies it. That became gate two: every
piece of text the finding actually quoted must be gone from the exact slot it was quoted from. Notably,
the reviewer's own suggested implementation was measured and rejected: it would have failed *open* on
roughly half of all claims, which is worse than what it replaced. The objection was right and the
proposed remedy was not, and we could only tell the difference by measuring.

A third reviewer then pointed out that all of this reads the database, while the finding is about what
a **live website** says — a correction sitting unpublished satisfies both gates. That became gate three:
the page must have been published since its copy last changed.

Seven council rounds, six of them "revise". Every one of those six caught something real.

Through all of it, one number stayed stubbornly uninformative. **No gate had ever refused anything.**
Zero. And we could not tell what that zero meant.

## What we've done

We made the zero explain itself, and that turned out to be the whole problem.

The realisation is simple once said. A refusal counter can only ever see one of a gate's two outcomes.
A gate that looked and approved reports nothing. A gate that was **never asked** also reports nothing.
And the three gates sit at the *end* of an eighteen-step process, so anything that stops earlier never
reaches them — which is not visible in a count either.

So each decision now records **which step decided it**. Eighteen fixed labels, with the gate steps
marked so they stand out. Nothing in the system reads the label to make a decision: an instrument that
can change what it measures is no longer evidence about it.

It went through the review council and was **approved first time**, with ten reviewers — worth noting
on a lane whose previous change took seven rounds. It is live, and it has run.

We wrote down what we expected to see *before* the first run, because a prediction made afterwards
proves nothing. It came out exactly as predicted: eighteen findings stopped at "the page still carries
these claims", **none** reached any gate, and nothing showed the "not yet instrumented" marker. Two of
the counts independently reproduce a reading we had done by hand the day before, one decision at a
time. Two different methods, same answer.

**So the answer to the two-day puzzle is that the gates were never consulted at all.** Not approving,
not refusing — never reached, because no page ever read clean enough to get that far. That was
unknowable from the counters and is now a query.

## Where we are now

**The mechanism is finished and correct, and it is doing almost nothing — which is the right amount.**
Eighteen findings are being refused because the claims really are still on the page. That is the check
working, not a fault.

**The gates remain unexercised, and we now say so from evidence rather than suspicion.** Nobody should
describe any of the three as having prevented anything. What they need is a page somebody has genuinely
cleaned, and nothing in the current population produces one.

**Of the eighteen refusals, only sixteen could ever reach a gate.** Two sit on pages that are not served
at all — one returns "not found", one was never published — so they will ask a human to fix copy on
invisible pages for ever. That is a consequence of a defect another lane owns, not a new one, and it is
recorded against their bug rather than duplicated into ours.

**Three times in two days, a claim of ours was refuted by a cheap measurement, and that is the part
worth repeating out loud.**

The first we caught ourselves, in a test, before anyone saw it: we had labelled the gate steps so they
could be found by a common prefix, and the label for "all three gates passed" deliberately reads as a
closure instead — so a search by prefix would have found only the cases where a gate *refused* and
missed every case where one *approved*. That is exactly the confusion the change exists to remove,
rebuilt inside the fix.

The second was caught by a reviewer, and it stings: we claimed the new label shows whether a gate has
**ever** been reached. It doesn't — each finding's record is overwritten daily, so it only ever
describes the latest run. The reviewer caught it by **quoting a warning this same lane had written two
days earlier about that exact column**. Same mistake, same column, three days apart.

The third arrived within hours of the first run. We had filed a note saying a blank value in the new
field could never happen; it happens on all nine already-closed findings, because a closed item is never
re-examined and its record is frozen at the day it closed. Anyone querying for "parts not yet
instrumented" the obvious way gets those nine and reads them as a gap.

Each was corrected where it was made, with what caught it named. And the third bought something back:
those blanks now *prove* that all nine historical closures happened before this measurement existed —
a question logged two days ago as unanswerable, now simply visible.

**One thing we nearly filed as a bug and shouldn't have.** The audit examines pages marked "archived",
which looks obviously wrong. Before writing it up we fetched them, with a deliberately fake address on
each site as a control. **One archived page is serving thirty-one kilobytes to the public.** So the
audit is right to look, and the obvious tidy-up — skip archived pages — would have stopped us auditing
a page genuinely making unsupported claims to real visitors. The distinction that matters is not
whether a page is archived but whether it is actually *being served*, and nothing currently checks
that.

## Where we're going

**One decision is yours, and it is the only thing standing between "built" and "proven in use".** The
gates cannot be observed working until a finding's page is genuinely cleaned. Either we wait for that
to happen naturally, or we deliberately create one — which means editing live site content, so we
haven't taken it unilaterally. Worth settling before more gates get built on the same pattern, because
the next one will have the same blind spot.

**Three pieces of tidying remain**, all small and all recorded. Only one of the five re-checkers names
its steps, so the others should be lifted onto the same shared machinery rather than each growing its
own. One reviewer's minor point from the approving round is still owed a test. And five older loose ends
from earlier in the lane are still listed.

**The transferable lesson, if this subproject is worth describing to anyone else, is not about claims or
gates.** It is that a zero is not a measurement. Over two days the same zero meant four different
things — never built, built but never asked, asked and approved, and the process stopped before it got
there — and no amount of counting could separate them. What separated them was recording *where the
process stopped*, which nobody thinks to build until they have been misled by the count. And the way we
found out was three cheap checks that each took under a minute and each overturned something we had
already written down with confidence.
