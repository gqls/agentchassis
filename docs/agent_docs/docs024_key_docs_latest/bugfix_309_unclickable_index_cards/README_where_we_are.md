# Where we are — the unclickable article index (bug 309)

*(Owner's plain-prose log. Append-only, newest at the bottom.)*

## 2026-08-18 — picked up, cause found, and it's a family not a one-off

The Platform Log page on fundamentallyai.com lists six articles as cards — but not
one of them is clickable. The cards look finished (picture, date, title, excerpt)
and just have no link. Another session found and filed this yesterday; today's
session took it on and traced it end to end.

What happened: the card section on that page was built months ago by an AI
component generator, and the generator told the cards to fetch their links from a
data drawer called "blog" in the site's records. **That drawer has never existed —
not on this site, not on any site.** The rest of each card gets written by the
content writer, so everything else filled in fine; the link is the one part that
had to come from the missing drawer, and the system's rule for a missing value is
"quietly leave it out". So the page shipped looking complete and linking nowhere.

Two things follow. First, the fix for this page is easy and clean: we already have
a properly built article-list component that gets its links by asking the database
"which articles does this site actually have?" — we swap the page onto that and
rebuild it. That also fixes the card that currently advertises an archived article,
because the ask-the-database route only lists real, live pages.

Second, and more important: we checked whether other components make the same
mistake, and **ten other invented "drawers" are referenced by eleven components**,
plus seven made-up database queries that don't exist either. Most of those
components aren't currently on any page, but three are (one page on the
leopardess site, two on gaswholesalers). Nothing stops the generator inventing
more. So the real fix is a rule at the door: when a new component is stored, its
data sources must be checked against what actually exists — made-up sources get
sent back with a message saying what the real options are. That's the change going
to the council.

A machine diagnosis of the same page was fired by the previous session and is
still queued; we'll read it when it lands as a cross-check on our tracing.
