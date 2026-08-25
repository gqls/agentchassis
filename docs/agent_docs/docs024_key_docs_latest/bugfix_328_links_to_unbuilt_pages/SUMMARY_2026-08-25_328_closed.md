# SUMMARY 2026-08-25 — bug 328 is closed: no page on any of our sites now links to a page we never built

## What we're trying to do

Stop the framework publishing links to its own pages that do not exist. When one page of a site
fails to build, every other page that links to it still goes out with the link in place, so the
reader clicks it and gets "page not found". A site with four good pages and no link to a missing
fifth reads as *small*; the same site with two dead links reads as *broken*, and that is a
judgement customers make about the whole product rather than about one page.

## Where we've come from

Filed on 19 August from a live site. The first account of it was that the platform already
*notices* these links, files a record, and nobody reads the queue — so the fix was plumbing.

That turned out to be wrong in a way that mattered: the records *had* been picked up, and a
builder *had* run against them and failed, because the only remedy the system knew was "build the
missing page" and on more than nine in ten cases that page cannot be built. The other remedy —
stop the *linking* page advertising it — existed nowhere in the platform.

The fix went in on 23 August after four rounds of review, each of which found something real,
including that our own headline census was false. It went live on the afternoon of the 24th, and
by that evening it had been proven working on two sites.

## What we've done

Since that proof, one thing was left: the fix only takes effect when a page is rebuilt, so pages
that had not been rebuilt since it went live were still serving their old dead links. Yesterday we
judged the system's own rebuild rhythm would clear them within a day, and wrote that down as an
instruction not to intervene.

**This morning we checked, and that judgement was half right.** Of the nineteen pages involved,
eleven had cleaned themselves up overnight — and every single one of those eleven was a page that
happened to be rebuilt after the fix went live, while the eight still carrying dead links were
every page that had not been. Nineteen out of nineteen, no exceptions. Nothing but the fix
explains that.

But the rhythm had stopped. The platform rebuilt 1,671 pages in a day and a half, and none of them
were the eight we needed — it works page by page, not site by site, and one whole site had had
nothing queued at all. So the remaining eight would have waited on luck.

With the owner's go-ahead we pushed all eight through. They rebuilt in twenty-seven minutes.

## Where we are now

**Closed.** Every one of the nineteen public pages that was carrying a dead link now serves none.
Before this morning there were eleven such links across four live sites; there are now zero.

Three separate things say this is real rather than a lucky reading. On each page that changed, the
total number of internal links fell by *exactly* the number of dead ones it had been carrying —
one page lost one, another lost three — and the eleven pages that were already clean came back with
their counts unchanged to the digit, so the system removed the dead links and disturbed nothing
else. All seven pages being linked *to* are still unbuilt and still return "not found", so the
links went because we stopped publishing them, not because the missing pages quietly appeared. And
the system left its own record of each removal, timestamped seconds before each page went out,
naming the exact links.

Two things found along the way outlast the bug. Our own advice not to intervene was an example of a
recurring mistake worth naming: we answered a question about the *stragglers* with a statistic
about the *average*, and both can be true at once, so the number could never have shown us we were
wrong. And the query that decides whether this bug is fixed had never been written down anywhere —
it had been quoted as a *number* in five documents and as a *query* in none — so closing the bug
meant rebuilding it from the source code first. Both are now recorded, and the queries are in the
runbook.

## Where we're going

Nothing on this bug. Three follow-ups are noted and none is blocking: a small per-page database
query we could cache; a gap where the new setting is not covered by the fleet-wide config audit;
and `RFC_049`, which records that the estate has now answered "may I link to this page?" three
separate times in three separate places — three live mechanisms to keep in step, so the case for
consolidating grows rather than shrinks now that the third is in service.

One thing found this morning genuinely deserves a decision, and it is bigger than this bug. **There
are 297 jobs sitting in the platform's work queue that nothing will ever pick up.** Not one has
ever been attempted, and 205 of them name a real handler, so to anyone reading the queue they look
like work in progress. They are parked in a state the dispatcher does not select. Worse, each one
silently blocks any fresh request for the same page — so that page cannot be asked to rebuild at
all. One of them was in this batch: a mortgagecalculator guide page, stuck since 3 August, which
rebuilt two minutes after being woken. An existing bug covers one slice of this; the general case
is unowned.
