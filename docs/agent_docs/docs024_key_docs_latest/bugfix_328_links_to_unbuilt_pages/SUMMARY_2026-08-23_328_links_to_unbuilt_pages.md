# SUMMARY 2026-08-23 — bug 328: a page that fails to build stays linked from the pages that did

## What we're trying to do

Stop the framework publishing links to its own pages that do not exist. When one page of a site
fails to build, every other page that links to it still goes out with the link in place, so the
reader clicks it and gets "page not found". A site with four good pages and no link to a missing
fifth reads as *small*; the same site with two dead links reads as *broken*, and that is a judgement
customers make about the whole product rather than about one page.

## Where we've come from

The bug was filed on 19 August from a live site, and a second lane added a contribution two days
later saying the platform already *notices* these links — it files a record naming the linking page
and quoting the link — and that those records simply sit in a queue nobody reads. On that account
the fix was plumbing: connect the queue to something.

That account was wrong in a way that mattered. The records had been picked up. A builder was
dispatched at each one and did run, and it failed — forty-eight times with "no sections ready to
build" and ten more with "content validation failed" — because the record's only instruction is
*build the missing page*, and the missing page is precisely the thing that cannot be built. The
platform detected the problem, dispatched correctly, tried the one repair it knows, and that repair
is the one that cannot work here. Nobody ever told the pages doing the linking.

Underneath sat a one-line gap with a wide reach. There is a shared piece of machinery that strips
dead links from a page just before it ships — it runs on the build path, the re-render paths, and at
the point content is saved. It asks "does this link point at a real page?" and answers by looking for
a row in the pages table. A page that was planned and never built *has* a row. So to every one of
those checks, a link to a page that has never existed on the web looked perfectly fine.

## What we've done

Built the fix, put it through the reviewer council, and revised it once on what the council found.

The rule is: just before a page's HTML leaves for publication, any link to a page of the same site
that would return "not found" and is not on its way comes out. Three properties make that safe
rather than merely clever. It does not touch what the page is made of — only the published copy
loses the link — so when the missing page is finally built the link comes back on its own at the
next render, with nothing to remember and nothing to go wrong later. It leaves the existing detector
seeing the link in the database, so we have not made the problem invisible by fixing it. And it is
off by default, switched on by a database change we have deliberately held back so it cannot go live
before the code that reads it.

Deciding *would this link break* was the hard part and the obvious answer was wrong. The natural
move — treat an empty "last published" column as "never published" — would have removed links to
nine pages that are serving perfectly well. The platform's existing narrower rule has the opposite
fault: it misses three genuinely dead pages, one of them the exact page this bug was filed about. So
a third signal was needed, and it is whether the page has any content stored against it at all:
twenty pages with none were all dead, nine with some were all alive, twenty-nine out of twenty-nine.

The scale is small and precise, which is what we wanted to see before shipping. Across every page on
every site there are 3,193 links between pages; this removes 36 of them, and every one of those 36
gives a reader "page not found" today.

## Where we are now

Committed, and with the council. The first review round sent it back. Two of its objections were
real and are fixed: we had claimed the change covered three publication routes when one of the three
is dead code that cannot run, and we had left one rendering agent switched off on the grounds that it
was "not the measured harm" — which is exactly the reasoning that leaves a second path broken until
someone hits it. We measured that agent instead of defending the gap, found it has produced no pages
at all, and switched it on. A third objection sent us to read a related bug file we had characterised
without opening; the distinction we drew turned out to hold, but our reasoning for it had been
unearned, and that is recorded.

The objection that actually blocked the round was mistaken, and the reviewer said how to check it —
it argued from a warning note's *title* and hedged that its full text might say otherwise. It does:
that note prescribes, almost word for word, the thing we built. One search settled it.

We also logged two mistakes of our own. A survey of 56 pages told us the exact opposite of the truth
because one domain in the fleet is parked at its registrar and answers "fine" for every address,
including one we invented to test it. And we wrote "fourteen days of live broken links" into our
notes when two thirds of that population is working now. Both are in the shared mistakes file,
because the tally is the point.

## Where we're going

Nothing takes effect until the next platform rebuild — normal here. After it: confirm the code is
actually running on both machines, switch it on, re-publish the 24 pages currently serving broken
links, and check the result on the real website rather than in the database. That last check has a
deliberate second half — as well as confirming a broken link is gone, confirm a *working* link is
still there, because a change that simply deleted all the links would otherwise look like a success.

One thing is deliberately left undone and written up rather than smuggled in. This is the third time
we have solved this shape of problem with a purpose-built piece rather than a general one, and our
own records flagged after the second that a third should prompt a wider rethink. We flagged it
ourselves in the submission; four reviewers objected on it anyway, which is fair — saying you owe a
debt is not paying it. So the count is now recorded where the next person will find it, and a ticket
is open asking what would make a *fourth* case a setting rather than another new predicate.
