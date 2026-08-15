# SUMMARY — bugfix 270, 2026-08-15: fix designed, coded and unit-tested; not yet committed, not yet shipped

## What we're trying to do

Stop a discovery check called `missing_structure` from ordering a full-site
rebuild on every single site, on every single pass, forever, for no reason.
It has been doing this since April. The rebuilds are not dangerous — they
just re-render what is already there — but each one jumps the queue ahead of
real work, and there have been about 31 of them for nothing so far.

## Where we've come from

The check was trying to answer a reasonable question — "did this site's
pages go live before their header and footer were ready?" — by reading three
columns on the `pages` table. Those three columns turned out to be empty on
every page, on every site, always, because the platform stopped writing
chrome there a long time ago and moved it to a different table
(`site_components`) without anyone updating this one check to follow. So the
check's question could never come back "no" — it was structurally
incapable of finding a healthy site. A second, unrelated filter in the same
query was also broken in the same "always true" direction, which made the
first problem worse rather than better.

This was found and written up as bug 270 two days ago, by a different
session that was building something else entirely and stumbled onto it while
answering a reviewer's question about a neighbouring check. That session did
the hard part — reproducing the false signal against the live database and
proving it — but correctly judged it wasn't part of what they were building,
left it for whoever picked it up next, and moved on.

## What we've done

Checked first that nobody else had picked it up (nobody had — the filing
session had explicitly marked it "unowned"), and that it was still
happening (it was — worse than when it was filed: 50 firings now versus 43
two days ago, most recent firing yesterday).

Then had a plan drawn up for the fix — deliberately using a different,
independent pass to weigh the two candidate approaches the original bug
report had already sketched, rather than just picking the first one. That
pass is worth explaining because it changed the shape of the fix:

The obvious-looking option was to delete the check entirely, since a newer,
better-designed check already exists that answers a very similar question in
a much more reliable way (it fetches the actual page a visitor would see,
rather than trusting a database column). But that newer check isn't switched
on yet — a different team deliberately left it off, as their own decision to
make on their own schedule — and deleting this one now would leave nothing
watching for this specific problem in the meantime. It would also require
juggling two changes in the right order on a shared system where "hold one
change back until the other is ready" isn't reliably possible.

So the fix instead points the SAME check at the RIGHT table. It now asks
`site_components` — the place chrome actually lives — whether a site's
header, footer and page-head have all genuinely rendered, and only orders a
rebuild when one of them hasn't. Because the check's name and its identifying
key didn't change, the fifteen-odd stale, wrongly-filed items already sitting
in the queue don't need a separate cleanup step: the platform has a
built-in mechanism for a check to say "I was wrong before, this is fine now,"
and the fixed check uses it. They will close themselves the next time each
affected site is checked again, automatically.

The code is written, it compiles, and five tests pass — including one that
specifically checks the fix can't quietly regress back to reading the wrong
column again, which is exactly the kind of one-line mistake that caused this
in the first place.

While researching this, a second, smaller instance of basically the same
mistake turned up in a completely different check (one that enforces
business decisions about page content) — it also reads the same empty
columns. It hasn't caused a visible problem yet, because none of the handful
of decisions on record happen to test anything in the header or footer. It's
being filed as its own separate bug rather than folded into this fix, since
it's a different check with a different failure shape.

## Where we are now

The fix is written and tested but not yet committed to the shared codebase,
not yet sent for the standing review this kind of change goes through, and
not yet built into the system that actually runs in production. None of the
50 historical wrongly-filed items have closed yet — that only happens once
the fixed code is live and each site gets its next routine check.

## Where we're going

Next: commit the fix narrowly (just the two files it touches), send it
through the standard review, file the second smaller bug as its own report,
then once an image is built and rolled, watch one full round of routine
checks confirm the fifteen-odd stale items close themselves and the check
goes quiet on healthy sites. Bug 270 stays open in the tracker until all of
that has actually happened and been checked, not just committed.
