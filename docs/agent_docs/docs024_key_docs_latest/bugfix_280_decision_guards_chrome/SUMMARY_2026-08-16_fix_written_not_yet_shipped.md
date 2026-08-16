# SUMMARY — bugfix 280, 2026-08-16: fix designed, coded, tested and committed; council review submitted; not yet shipped

## What we're trying to do

Stop a feature called "decision guards" from silently checking the wrong
thing. A decision guard is a promise recorded against a page — "this page
must always link to the audience-check tool" — that the platform re-checks
automatically every time it scans a site, flagging it if the promise stops
holding. The bug: the re-check was quietly leaving the header and footer out
of what it looked at, so a promise about anything in the header or footer
could never be checked correctly — either always wrongly flagged as broken,
or never flagged even when it genuinely broke.

## Where we've come from

This is the second half of a mistake bug 270 already found and fixed two
days ago in a different check. The platform used to store a page's header
and footer directly on the page's own database row, and at some point moved
to storing them somewhere else instead (a table called `site_components`)
without updating every place that still read the old location. 270 fixed
one such place. While fixing it, that session noticed a second, unrelated
check had the exact same problem, filed it separately since it fails in a
different way, and handed it off explicitly rather than trying to fix both
at once.

The owner asked this session to pick up "bug 180" — a different, older bug
that turned out to already be fixed and confirmed working two weeks ago.
Bug 270's own handoff note anticipated exactly this mix-up and said, in
effect, "you probably mean 280" — which, once asked, the owner confirmed.

## What we've done

Checked first that nobody else was already working on it. The usual
ownership check said "maybe" (because the session that found it is still
active on other things), but a closer look — no recent edits to the file,
no other live session actually discussing the specific line of code involved
— said it was genuinely free to pick up.

Read the broken check in full, plus the sibling fix from 270 as the pattern
to follow, and confirmed the shape of the underlying database table before
writing any new database query (a good habit here, since guessing the
table's structure wrong is a common way to introduce a new bug while fixing
another one).

Wrote the fix: point the same lookup at the right table for the header and
footer, exactly the same idea as 270, while being careful to keep intact a
separate, easy-to-miss behaviour the surrounding code depends on — what
happens when the check is asked about a page that doesn't exist at all
(it needs to notice "no such page," not quietly proceed as if the page
existed with an empty header). Caught that risk before it became a second,
new bug rather than after.

Added tests, including one specifically designed to fail if the fix is ever
accidentally reverted, and proved that test actually works by deliberately
breaking the fix, running the test, watching it fail, then putting the fix
back — rather than just trusting that the test looks right. Confirmed the
whole codebase still builds and every existing test elsewhere in the same
area still passes.

Committed the change on its own, then sent it through the platform's
standing advisory review process (a panel of automated reviewers) — that's
in progress now, expected to report back within about half an hour.

## Where we are now

The fix is written, tested, committed to the shared codebase, and under
review. Nothing has changed in production yet — the code is inert until it's
built into a new version of the service and that version is rolled out, and
that step is deliberately left for the owner to decide on and trigger, the
same way it was for bug 270's fix a day ago. No live guard has ever produced
a wrong result because of this bug (all five guards on record today happen
not to touch header or footer content), so there's nothing currently broken
to point at — this closes a trap before anyone fell into it, not after.

## Where we're going

Next: read the review's verdict once it lands and act on anything it raises.
Then it's the owner's call whether and when to build and roll this out.
Once it's live, there's nothing further to verify behaviourally (no guard on
record would change its answer), so "shipped" will be confirmed the same way
270's was — by asking the running service what code it's actually built
from, not by inferring it from the roll having happened.
