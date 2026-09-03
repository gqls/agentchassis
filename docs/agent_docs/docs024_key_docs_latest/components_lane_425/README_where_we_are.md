# README — where we are (components lane, `bugs_open/425`)

The owner's running log. Plain prose, append-only, newest at the bottom. Anyone may add; nobody
rewrites what is already here.

---

**2026-09-03, afternoon.** The thing that has been stuck for two days is solved, and it turned out
not to be our bug.

The complaint this lane started from was yours: the cards need better designs. Underneath that
was a real defect — a row of article cards on a listing page was rendering four empty boxes,
because two different bits of code write the same kind of list item and they disagreed about what
to call things. One wrote the summary text under the name `meta_description`; the template was
looking for `excerpt`. One stripped the site name off the end of each headline; the other left it
on. We fixed both halves. The template half went out on 2026-09-02 and worked. The code half went
into the running system and then appeared to do nothing at all, on one particular route.

That "appeared to do nothing" is what ate two days. We could see the fixed code inside the running
binary. We could watch a page re-render, complete successfully, and come back with the old broken
shape. Three separate automated diagnosis runs failed to explain it. The lane eliminated sixteen
possible causes, one at a time, by measuring every affected page on the estate. All sixteen
eliminations were correct and none of them was the answer.

The answer was in a different file, and another session had already found it. When someone split a
long function in two on 2026-09-02, one value stopped being handed across the new boundary — a
single missing line of assignment. The effect was that **every light page re-render across the
whole estate spent a fortnight rendering each page's own old data back at itself** and reporting
success. Nothing errored. Every count the system prints looked healthy. The only symptom was an
improvement that failed to appear, which is the hardest kind of thing to notice. That is filed as
bug 454 by the session that found it, fixed with one line, and live since 12:18 today.

So our card fix was never broken. It was being handed data that had not been refreshed.

We proved it properly this afternoon rather than taking it on trust. The trick was picking the
right page to test on, and the previous two attempts had got that wrong in an interesting way:
they both tested on a page that **already** had the fix's output in it, where "it worked" and "it
did nothing, and the old good data survived" look identical. We picked a page that was still
broken — a garden tools care page — re-rendered it, and watched the summary text appear and the
site name drop off the headlines, all the way through to the live page on the internet.

Then, while we were writing it up, the estate started fixing itself. Two darts pages repaired on
ordinary background traffic that nobody aimed at this problem — which is better evidence than our
own test, because no one chose those pages. The population of broken card rows has gone from five
of seventeen good yesterday to nine of seventeen good at the time of writing, and it keeps
improving on its own as pages get re-rendered for other reasons. **So there is nothing to schedule
here.** Two more check pages are queued to confirm on unlike cases, and that is the end of it.

Two things worth knowing beyond this lane.

First, the same missing line explains the other complaint we were chasing — the pages wearing the
wrong hero image. Six page-heading components were declaring no image field; we fixed that on
2026-09-02 and only one page visibly improved, which was baffling at the time. It was baffling for
the same reason: the re-renders that should have picked up the change were not refreshing anything.
That is now fixed too, by the same one line, and the other session's evidence shows a heading
component recovering its own image on a single re-render.

Second, a page on the boxing site is serving genuinely broken cards, and it is a different fault.
The delivery session swept the site before shipping and found empty card slots on the articles
listing page. We traced it: a piece of code has been quietly appending a duplicate, unowned row to
that page every time it runs, and the page is now serving **six** stacked copies of the same list
of six articles — thirty-six cards where there should be six. Each copy is frozen with whatever the
template looked like on the day it was created, which is why some of them still show the old empty
boxes. That is bug 457, already filed from this lane. It needs the code fix and a rebuild.

**Correction, same afternoon, and it was mine to make.** I first wrote that no re-render could fix
those rows at all, because they have no component attached. Another session challenged it and was
right: the re-render code falls back to matching on the slot name, and these rows do match, so a
re-render would refresh their content and clear the empty boxes. What it cannot fix is that six of
them exist — the page would then serve six up-to-date copies instead of six stale ones. Deletion is
still the answer, for the duplication rather than for the emptiness. Across the whole estate only
two rows anywhere are genuinely beyond a re-render, and neither is on the boxing site. We have not
touched the rows by hand — it is a paid site and that is the wrong instrument.

One correction to own: I told the delivery session that the card fix had "not reached" that
articles page. It never will, because the fault there was never the card fix. Written into the bug
file so the next person does not inherit my wrong version.
