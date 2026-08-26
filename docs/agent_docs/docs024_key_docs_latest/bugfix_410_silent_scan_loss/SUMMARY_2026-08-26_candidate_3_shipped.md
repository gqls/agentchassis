# Summary — 2026-08-26: the silent scan loss is closed, and the class behind it is pinned

## What we're trying to do

Stop the system quietly losing pieces of a page and then reporting that it rebuilt the page
successfully.

The specific problem: when the framework re-renders a page, it first reads that page's stored
pieces out of the database. If reading one of those pieces failed, the old code wrote a line to
the log, skipped it, and carried on. The caller had no way to tell a short answer from a genuinely
short page. Worse than it sounds, because the very next step **deletes and rewrites** the page's
rows from whatever it was handed — so a piece dropped during the read was not merely left out of
the render, it was **deleted from the database**. The page then shipped with a hole in it, under a
fresh "last built" stamp, with the job marked complete.

## Where we've come from

This was filed as a **pattern**, not a bug. Three different lanes hit the same shape in one week:
a listing that re-rendered without its images, a re-render instruction the code did not recognise,
and this one. Every time, the system chose the quiet option when it met something it did not
understand, and every time it reported success.

The insight in the original filing is the one that made it worth a lane of its own: **the safe
default and the silent-failure default are the same default.** Doing the quiet thing is almost
always right here, so nobody has any reason to look — which is precisely why every kind of drift
ends up landing there. Failing the other way would announce itself; failing this way announces
nothing by construction.

It was filed deliberately unowned. The lane that found this instance had it in their hands and
refused to fix it inside their own feature work, because it sits on the busiest pipeline we run.
That was the right call, and it left the work sitting there until now.

## What we've done

**Fixed the instance.** The reader now counts the rows the database actually handed it, compares
that to the rows it kept, and refuses outright if any went missing. The refusal happens before
anything is written, so a page keeps serving its last good version while the job fails and
retries.

Choosing *which* two numbers to compare was the whole design. The obvious comparison — "does this
page have as many pieces as the database says it has?" — is wrong, and would have killed the
guard within a week: the query deliberately filters some rows out, so that version fires on every
healthy page carrying a deleted component. A guard that cries wolf on correct input gets relaxed,
and a relaxed guard is a dead one.

**Answered the hard review question with a number instead of a promise.** The obvious objection
is that turning "quietly does the wrong thing" into "stops and complains" could light up the
busiest pipeline we have. So we measured it: it cannot fire on any data the database can currently
hold. Every value being read is either guaranteed to exist or already has a fallback. The day-one
effect is nothing — not "nothing as far as we sampled", but nothing by construction. It can only
ever speak when someone later changes the code in a way that breaks the read, which is exactly the
change that caused seven tests to fail this week and told nobody why.

**Pinned the class so it cannot grow.** There are 207 places in the codebase with this same shape.
Fixing them all blind would be a bigger risk than the bug. So instead there is now a tripwire that
records how many exist in each file and fails the build if any file grows a new one — and, notably,
fails just as loudly if a count goes *down* without the record being updated, so the tally can only
ever improve. A second, advisory copy covers the rest of the tree.

**Found out the shape is not the defect, which changed the design.** One of those 207 places is
our own best example of doing this correctly — it has the same shape *and* a proper guard, because
its guard sits after the loop. The obvious tripwire, one that simply banned the shape, would have
convicted the best code we have on this pattern. That is why it counts rather than forbids.

**Proved every guard can actually fail.** Each one was deliberately broken to confirm the test
goes red, then restored. When the main guard is removed, the test output reproduces the original
bug word for word — "returned 1 section with no error". A check nobody has proven can fail is the
exact defect this bug is about, reintroduced as its own fix.

## Where we are now

The fix is committed and will go out with the next chassis build. The tripwire and its advisory
twin are in. It is registered so other lanes can find it rather than build a second one. It has
gone to the review council and the verdict has not landed yet — which is normal, and the commit is
marked so it gets credited automatically when it does.

The two lanes working next to this have both confirmed no conflict, checked the change against
their own work, and each found something in it that mattered to them: one caught a stale figure in
their own review submission, and the other found that their in-flight refactor would have silently
changed the meaning of an unrelated check.

**Five measurement errors were made and logged along the way**, and they are all the same error
wearing different clothes: the measurement answered the question that had been encoded, not the
one that was asked. The most instructive was verifying another lane's figure by unknowingly
repeating their exact mistake and matching it to the digit — which felt like confirmation and was
actually two people using the same wrong method. The lesson written down is that re-running,
re-deriving and cross-checking all failed; the only thing that ever caught anything was opening one
member of the population and reading it.

## Where we're going

Three things remain, all named rather than quietly dropped:

- **The bug stays open.** One of its three instances is fixed. The other two belong to other
  lanes, and the expensive door-closing option — making "I did not understand this" a refusal
  everywhere rather than a fallback — still needs its own review.
- **A second kind of silent loss sits in the very same loop**, on a different axis: if a piece's
  stored content is corrupt, the row survives but its content is emptied, so the counts match and
  the new guard cannot see it. That needs a decision about whether an unreadable section may
  render as an empty one.
- **The 207 will go stale by addition**, so it carries its date and a command to re-check it
  before anyone quotes it.
