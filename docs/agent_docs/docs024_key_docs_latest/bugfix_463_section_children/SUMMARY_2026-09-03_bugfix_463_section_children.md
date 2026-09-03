# SUMMARY — bugfix 463, a section index's new children (2026-09-03)

## What we're trying to do

Make it possible to add a page to a section of a website. Specifically: when the system plans a
site that already has an "Articles" or "Guides" or "News" section, the articles it plans for
that section should end up in it. Until today they could not — not through any number of
rebuilds.

## Where we've come from

gamedesign.uk shipped with an articles hub and no articles. It was rebuilt three times. Each
time the planner did its job correctly and the articles still did not appear. The third rebuild
was the clean case, because by then an earlier, unrelated defect in the planner had been fixed,
so nothing else could be blamed: nine pages proposed, four saved, five articles gone, and every
status green along the way.

The `gamedesign.uk` lane diagnosed it properly, filed it as bug 463, and passed the fix on. Two
other guards in the same routine had each been blamed for it at various points, and one of the
findings in the bug file is that they were interlocking: one guard deleted the articles, and a
second guard then held the empty hub back for having nothing in it. Each one's evidence read as
a reason for the other.

## What we've done

Fixed both halves and committed them.

The first half is the one that was filed: the cleanup step now compares the whole web address
rather than just its first segment, so it can tell an article inside a section from a page
clashing with that section. It keeps doing the job it was written for — the clash it was
designed to catch is still caught, and that behaviour was independently ratified in an earlier
bug, so it was important not to lose it.

The second half was not in the bug report, and the report explicitly says it is not part of the
problem. It is. When the plan is saved, the chosen web address is thrown away and rebuilt from
the page's type, which sends every article to `/blog/` unless it is told which section the page
belongs to — and nothing ever tells it, because the planner is not asked. Without this half the
first fix passes its own tests and changes nothing on the live page. We now work the section out
from the address the planner chose, carefully gated so it cannot move a page that already
exists.

Along the way we filed a second, separate fault (bug 467: a cap that discards every new page
once a site reaches twenty, which most of our sites have), corrected the original bug report
where it was wrong, recorded the trap for the next person who touches these functions, logged
one of our own mistaken measurements, and coordinated with five other threads working nearby —
one of which was building the reporting half of this same bug in the same file at the same time.

## Where we are now

The code is committed and under automated review. It is not live: changes of this kind do
nothing until the software is rebuilt and redeployed, which is a separate, fleet-wide action.

So the honest status is **fixed in the code, unproven in the world**, and the bug stays open on
that basis — this estate's bar for closing is fixed *and* live.

## Where we're going

One thing: prove it. When the next deployment goes out, re-plan gamedesign.uk and check three
points in order — that the plan keeps every page it proposed, that the articles are saved under
`/articles/` rather than `/blog/`, and only then look at the page itself. The middle check is
the one that matters most, because it is the one a half-fix would fail while still looking
successful.

Two things sit behind that. gamedesign.uk must be the test site, because it is small enough to
be under the twenty-page cap that bug 467 describes; on most of our sites that cap would eat the
articles immediately afterwards. And whether the hub then *displays* its new articles depends on
a separate fault another thread is already fixing.
