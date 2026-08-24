# SUMMARY 2026-08-24b — approved, and deliberately smaller than it started

*Second summary today. The first (`SUMMARY_2026-08-24`) was written mid-session and said the fix
was committed and awaiting review. Both halves of that have since changed: the review is done,
and the review made the fix smaller. That is the inflection, so this is a new file rather than an
edit — the series is the record.*

## What we're trying to do

Unchanged: make the platform able to build the page types it never could — a directory of real
businesses, an index of a site's real guides — through the ordinary pipeline rather than by hand.

## Where we've come from

The August work built the missing machinery and both vetcomparison pages went live. Today we
checked that was still true (it is, and the pages survived a fleet-wide re-render, so the
pipeline reproduces them rather than preserving them), and then found that the August fix had
only ever been installed on **one of two doors**. A second routine — the one that reconciles a
site's plan against what has actually been built — had the answer "use the generic builder" typed
into it as text and never asked. Five pages on three other sites were parked because of it, one
of them for fifteen days while the builder it needed was running.

## What we've done

Built the missing piece: one function that answers "which builder builds this kind of page",
which that second routine now asks. Committed, reviewed, and inert until the fleet's next image
build.

**The review is the part worth reporting.** It took six rounds. Five came back asking for
changes, and rather than being friction, two of them changed the code and one changed the shape
of the fix:

- It caught a database column I had cited that does not exist.
- It caught that I had put a change I had *myself measured as dangerous* into the part of the
  submission that a machine executes — as a way of documenting that it should not be run.
- It caught that my submission had drifted out of step with my own code after I revised twice.
- And the important one: I had argued twice that a temporary inconsistency between the two
  routines was harmless, on the grounds that a database constraint stops a page ever having two
  conflicting jobs. A reviewer took that same constraint and showed it cuts the other way —
  whichever routine files first wins, and the other is silently discarded. So for any page the
  *other* routine reached first, my fix would simply never have run, and nothing anywhere would
  have said so.

**So the fix is now smaller than it was this morning, on purpose.** The one page type that made
the two routines disagree has been taken back out, which makes them identical again and the
problem impossible rather than merely unlikely. The case that started this — the directory page
parked for fifteen days — still gets fixed, because that page type was already handled correctly
on the other side. Two of the five parked pages are no longer covered and stay exactly as they
are today; they get fixed by a one-line change when a file another team has left in a broken
state becomes safe to touch.

I would rather report that honestly than describe a larger fix that had a silent hole in it.

## Where we are now

Approved, committed, inert until the next roll. The bug stays open, because the fault is still
happening on the fleet right now and stops when the code ships, not when it is written.

Two things remain in other people's hands, both recorded where those people will see them. One
team decided three weeks ago to deliberately leave that reconciling routine alone; half of my
change touches that decision, so I have put it to them in their own file and offered to take that
half straight back out if they disagree. And the remaining tidy-up is blocked on another team's
unfinished work, which I measured as breaking the build and three existing tests, and which I am
not going to adopt under my own name to make my change look complete.

## Where we're going

When the next fleet build lands: check the running services actually carry the change, watch the
reconciling routine file a page at the right builder, then re-trigger the parked pages using the
written recipe — and only then does the bug close.

Nothing else is owed on this. The wider version of this problem is larger than this lane and is
already someone else's: of the eighty-seven stuck pages across the fleet, sixty-nine are tool
pages being created faster than anyone drains them by a separate bug we filed in August.

One last thing worth the owner knowing, because it is about how we work rather than what we
built. This session logged ten of its own mistakes to the fleet's shared record. Three were the
same habit wearing different clothes — reading a result for what it confirmed instead of what it
ruled out — and one of them left three gigabytes of temporary files on the shared machine, which
another team had to write a rule about while I was still doing it. The review process caught what
my own checking did not, four separate times. That is the argument for the review process, and it
is also the argument for writing the mistakes down.

---

> **CORRECTED 2026-08-24, same day, by a measurement from the `loanzy_uk_example_site` lane.**
> Above I wrote that the two pages the narrowing leaves out *"stay exactly as they are today"*.
> That is true of the pages and **false of the sites**, and the difference matters to a reader
> deciding whether the narrowing was worth it.
>
> That lane ran an unaided greenfield build of `garden-tools.uk` overnight and measured the
> finished site: the parked page is the target of **three dead links from three live pages, one
> of them the home page**. So a visitor meets a 404 from the front door. Nothing suppresses a
> link when its target fails to build — that is a separate known defect, but it is the mechanism
> through which this decision's cost is actually paid.
>
> The narrowing still looks right to me, and to that lane: a dead link is *visible*, where a
> silently mis-routed page is not, and the reviewer's argument was precisely that the alternative
> fails without telling anyone. But "stays as it is" undersells it, and the honest version is:
> **leaving those pages unrouted costs a 404 from the home page on each greenfield site, until
> the follow-up lands.**
>
> The same lane also found that this problem is wider than the three page types this bug names:
> blog pages fail the same way, for a subtly different reason — their page type *is* mapped
> correctly, to a builder that simply cannot create a layout from nothing. That is the larger
> and better fix, it is not what shipped, and it is now written down as the next step rather
> than quietly folded into an approved change.
