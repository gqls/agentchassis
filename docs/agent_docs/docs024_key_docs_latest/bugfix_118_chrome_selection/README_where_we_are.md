# Where we are — chrome component selection (bug 118)

Plain-prose log, append-only, newest at the bottom.

---

**2026-07-31, evening.**

Picked up bug 118 off the open pile. The complaint, filed four days ago by the
relojistas lane, was this: somebody made an owner-approved fix to a site's
footer, applied it to the footer component that was marked *active*, rebuilt the
chrome, and nothing changed on any page. It turned out the code that chooses
which footer to use never looks at whether a component is active. It just takes
whichever name comes first alphabetically. Three of the five footers in the
library are switched off, and the alphabetically-first one — a switched-off one —
was winning on every site.

So the thing you'd naturally do to fix a footer (edit the live one) is the one
thing that cannot work, and it fails in a way that looks exactly like your fix
being wrong.

**What I found when I went to check it was still true.**

It was still true, and it was worse than filed in one way and much better in
another.

Worse: there are actually *three* separate bits of code asking "which footer
should this site use?", and all three answer differently, and all three are
wrong. One ignores active/inactive entirely. One checks active but doesn't
notice that a component can be a private *copy* belonging to one client — so it
would happily hand leopardessconsulting.co.uk's own bespoke header to every new
site we build. The third checks both of those but doesn't sort its results at
all, so it returns whichever row the database happens to hand back first, and
what it happens to hand back is a *page-body* component that shares the name.
Three answers to one question, each wrong in its own way. That is the actual
bug, and it's why patching one of them would have been pointless.

Better: the bug file said the fix "changes the rendered footer on every site" and
therefore needed your say-so. That isn't right, and it's the reason this sat
untouched for four days. The choosing code only runs for a site that has never
been assigned a footer at all. All fourteen real sites were assigned one long
ago, and that assignment is what they render from — so changing how we choose
affects exactly one site today: loancalculator.co.uk, which you created
yesterday and which has no chrome yet. Every other site is unaffected. No page
re-renders.

**What I've done.**

There is now one rule for what counts as a usable piece of site chrome — it must
be switched on, it must not be somebody's private copy, and it must actually be a
chrome component rather than a page section — written down once, used by the two
places that assign chrome to a site, with a test that fails if anyone writes a
fourth copy of the rule. Committed and sent to the review council; it goes live
with the next chassis build.

**Two things I deliberately have not done, because they're your call, not mine.**

First: eleven sites are still *pointed at* the switched-off footer, and seven at
a switched-off header. Fixing the chooser doesn't move them. Repointing them
would change how those eleven sites' footers look — different columns, different
headings — so that's a visible change to live sites and I'm not making it on my
own. Worth knowing: the platform has been *noticing* this since 17 July. It
raises a ticket saying "this site's footer points at a deactivated component" and
sends it to a repair job that re-renders... the same deactivated component. So
the ticket can never be satisfied, and two of them are sitting there marked
"unresolved after 2 attempts". That's a separate defect and I've written it up.

Second: the page-building path still uses that third, unsorted lookup, so it is
still capable of rendering a page-section component as a site header. Fixing that
changes the header and footer markup on every page we build from now on. Same
reason — visible, fleet-wide, your call. Measured and written down rather than
quietly shipped.

**The question I'd put to you**, when you want it: do we repoint the eleven sites
onto the active footer? It's a one-line database change plus a chrome re-render
per site, and it would let those stuck tickets close for the first time since
July. The risk is that the active footer looks different from what those sites
have been showing, and I'd want to do one site first and show you before and
after.
