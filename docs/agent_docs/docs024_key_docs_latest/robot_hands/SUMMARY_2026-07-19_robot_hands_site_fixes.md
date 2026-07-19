# SUMMARY — robot-hands.com site fixes (2026-07-19)

## What we're trying to do

robot-hands.com is the imagery workstream's testbed — a reference site for
engineers choosing industrial robot grippers, built and maintained entirely by
the agent platform. On 17 July the owner looked at it and found it visually and
functionally degraded: the dark technical look had turned into a generic blue
brochure, one page had yellow text on white, the tools mostly didn't work, a
"Load More" button did nothing, and two-thirds of the articles the site listed
were dead links. The job was to fix those six things and, wherever the cause
was a platform defect rather than bad data on one site, fix the platform.

## Where we've come from

The starting handoff had a theory: a regeneration run on 16 July had rewritten
the site's header without theme awareness. That turned out to be wrong in an
instructive way. The blue header had actually been there since 9 July — a week
earlier — and the 16 July run had only spread it across all 37 pages. Two other
premises in that handoff were also wrong: the snapshots we were told to restore
from don't exist, and the colour-fixing agent everyone suspected had never
actually changed anything, because it fails on every single run and reports
success anyway.

The real cause was more structural. Back on 10 July the site had been switched
to a dark layout, but that switch only changed one of four places colour lives.
Three copies of the palette stayed light, the site's header and footer were
still being rendered from deactivated components that bake colours directly
into the page, and — worst of it — a broken health check was telling the
platform, every single pass, that this site had no design. Each time it did,
the design agent re-invented the colour scheme from scratch. On 17 July alone
the site's stylesheet was rewritten four times, and one of those rewrites put a
white background on a site whose whole design is dark. That went live.

## What we've done

All six defects are fixed, and five of them are confirmed on the live site
rather than just in the database. The dark theme is back, built so it can't
break the same way again: the new header and footer carry no colours at all,
only references the stylesheet resolves, so any future regeneration stays
correct by construction. The learning centre now has one address instead of
three. The dead articles are retired and the listing shows only the three real
guides. The dead button is gone.

Underneath, we fixed the broken health check that caused the repainting — and
that fix is now live in production and proven: since the new image rolled last
night, that false alarm has not fired once anywhere in the fleet. We also
pinned this site's colours so the design agent stops inventing.

Three platform defects were written up for other threads: the colour-fixer that
can never run, the fact that a failed job can be recorded as complete, and the
missing guard that let a white background land on a dark site in the first
place.

The platform change went through the reviewer council, which was worth it. It
took three rounds and ended at seven of eight reviewers approving. Along the way
it caught a real bug in my fix — I was testing whether a value *existed* rather
than whether it had anything *in* it, which would have created the mirror image
of the problem I was fixing — and it correctly made me pull a broader change out
into its own piece of work rather than smuggling it in.

## Where we are now

The site looks right and its content links resolve, with two exceptions. Two
tool pages were planned but never built, so five buttons on the homepage lead
to "page not found". That is now the most visible thing wrong with the site.

Everything else outstanding is queued work rather than damage: the site carries
a large backlog of older items, none of it from this work, and the platform
works through it slowly by design — one item at a time per site.

## Where we're going

Next is building those two missing tool pages, which belongs with the
experience-loop workstream because it is already solving exactly that problem
across every site — so the next session should join that effort rather than
start a parallel one. If that has to wait, the quick mitigation is to point
those five buttons at a page that actually exists.

After that, the scheme guard: the reason a white background could land on a
dark site is that nothing anywhere checks a proposed colour against the design
the site has chosen. That is written up and ready to submit, with the
verification work already done.

One caution for whoever picks this up. The project's own guidance changed
yesterday: claims about how the platform works are now supposed to go through
the automated diagnosis loop *before* being asserted, because a session with
full context recently filed a confident structural claim that was refuted in
under ten minutes. Two of the write-ups from this work predate that change.
Their evidence is cited and checkable, but if either is about to become the
basis for someone else's work, it should go through the loop first.
