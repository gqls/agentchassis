# SUMMARY — bugfix 280, 2026-08-17: fixed, reviewed, shipped, and confirmed live on both replicas. Closed.

## What we're trying to do

Stop a feature called "decision guards" from silently checking the wrong
thing. A decision guard is a promise recorded against a page — "this page
must always link to the audience-check tool" — that the platform re-checks
automatically every time it scans a site, flagging it if the promise stops
holding. The bug: the re-check was quietly leaving the header and footer out
of what it looked at, so a promise about anything in the header or footer
could never be checked correctly.

## Where we've come from

This was the second half of a mistake bug 270 already found and fixed
elsewhere: the platform used to store a page's header and footer directly on
the page's own database row, then moved that storage somewhere else without
updating every place that still read the old location. 270 fixed one such
place; this bug was the other one, found in passing while fixing 270 and
handed off deliberately rather than fixed on the spot.

Picked up after the owner asked for "bug 180" — a different, already-closed
bug — and confirmed, once asked, that 280 was the intended target.

## What we've done

Wrote the fix (point the same lookup at the right table for the header and
footer, same idea as 270), added tests including one proven by deliberately
breaking the fix and watching the test catch it, and sent it through the
platform's standing review.

The review came back once with a real, fair question: how confident are we
that the database column actually uses the words "header" and "footer",
given a similarly-named but different thing exists elsewhere in the
codebase? Checked directly against the live database and the two places in
the code that write that column — both confirmed the assumption was right —
and sent that evidence back on the same review. It passed clean the second
time, with three minor notes that didn't need any code changes.

The owner then built and rolled out a new version of the service. Rather
than take "it's deployed" at face value, checked it against the two actual
running copies of the service directly — asked each one, in a way that
proves the answer isn't just a lucky-looking coincidence, what exact version
of the code it's running, and confirmed both are running code that includes
this fix.

## Where we are now

Fixed, tested, reviewed, shipped, and confirmed live on both running copies
of the service. The bug file has been moved from the open list to the
closed list. There is nothing left to do — and importantly, nothing will
visibly change, because this was a silent, no-symptom-yet problem (none of
today's handful of recorded promises happen to be about header or footer
content), not an active one. The fix closes a trap before anyone fell into
it.

## Where we're going

Nothing further on this bug specifically. One thing worth carrying forward:
confirmed again that this particular service's "what are you running"
startup message disappears within minutes because it's a busy service —
worth remembering for the next person who needs to check what this service
is actually running, so they go straight to asking the running program
directly rather than trying the startup message first.
