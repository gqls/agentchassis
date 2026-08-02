# SUMMARY — 2026-08-03 — the repair learns what is not markup

*Read-out for the owner. Current state only — the chronology is in `README_where_we_are.md`
and the technical log in `NOTES_sibling_link_repair.md`.*

---

## What we're trying to do

Make sure that when the platform publishes a page, every link on it goes somewhere real —
and that the machinery which enforces that never damages the page while doing it. The first
half has been the work of this lane and its neighbours for a fortnight. The second half is
what this milestone is about, because we found the enforcement damaging pages.

## Where we've come from

The platform detects dead internal links accurately and has done for months, but nothing
acted on the findings, so 404s shipped. The answer was a *repair*: before a page leaves,
rewrite the links that can be corrected and unlink the ones that cannot, so the prose
survives and the dead link dies. That repair now runs at four points in the pipeline, and
this lane's last milestone (`bugs_closed/136`) added the fourth.

Closing that one turned up a new fault while re-running its census. A row in the count was
not a link at all. It was **JavaScript that builds a link when the visitor uses the tool** —
and our repair had been reading it as a broken link and deleting it.

## What we've done

**Reproduced it first, and the reproduction changed the fix.** Rather than patch the case in
the ticket, I ran the shipping repair over six pieces of realistic markup. Five came back
damaged, and only one was the case we knew about. A link built by joining strings is deleted
one way; a link built with the newer backtick syntax is deleted by a *different* branch of
the same function; markup quoted inside a stylesheet comment, inside a text box, and inside a
commented-out block are all rewritten too. One cause, five faces.

That killed the cheap fix. Teaching the repair to recognise the one JavaScript spelling from
the ticket would have fixed one of the five and rotted the moment somebody wrote a link a
sixth way.

**So the platform now has one shared answer to "which parts of this page are not really
markup?"** — scripts, stylesheets, text boxes, comments — and the two pieces of code that
*rewrite* pages both consult it. It reuses a page-scanner we already had rather than adding a
second one. Registered as LNK-029, submitted to the review council, and committed.

**Verified against production before submitting, not after.** I ran the old and the new
matching over all 509 of our pages on all 19 sites. One page is being damaged today (the
vet-comparison CMA tool), eleven places fleet-wide are exposed, and — the number that
mattered — **zero** genuine repairs are lost by the new guard. A guard that is too cautious
would quietly stop fixing real broken links, and that is the failure this change could
plausibly have introduced. It doesn't.

**And we confirmed the damage on the visitor's page**, which nobody had done: fetching the
live tool shows the link gone from the shipped JavaScript while our database still holds the
correct version. That matters beyond this bug — it means a database-only audit would have
reported that page healthy indefinitely.

**One honest failure, logged.** I made a deliberate technical choice, wrote a test to protect
it, and wrote a confident comment saying "change this and the test fails". Then I changed it,
and the test passed — a second safety net underneath had absorbed the change, so my test had
been proving nothing. The previous session in this lane had written that exact trap into its
handoff, which I read that morning. Reading it did not stop me repeating it, which is the
part worth recording.

## Where we are now

The fix is **committed and correct, and not yet live**. It missed the v1.0.1231 build by
about ninety seconds, and I have checked that on the running pods rather than assumed it —
the new code is absent from both replicas while the controls confirm the check itself works.
So the ticket stays **open**: the defect is still reproducible on the fleet until the next
chassis image ships. The council verdict is still running.

## Where we're going

1. **Next chassis build** carries it; then the pod check on both replicas, and the real proof
   — re-render that tool page and watch the deleted link come back on the live site.
2. **Then the piece this unblocked.** The previous milestone ended by saying the obvious next
   step — switching this repair on for tool pages — must *not* be taken, because tool pages
   are exactly where JavaScript builds links. That blocker is now gone.
3. **The detectors keep the same blind spot, on purpose.** They report a JavaScript-built
   link as a finding, which costs a person's attention rather than costing content. Fixing a
   thing that *judges* is a different decision from fixing a thing that *writes*, and it
   belongs to the detection lane. The shared helper is already there when they want it.
