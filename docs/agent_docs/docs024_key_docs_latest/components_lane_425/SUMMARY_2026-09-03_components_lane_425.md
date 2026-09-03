# SUMMARY — components lane (`bugs_open/425`), 2026-09-03

Written to be read aloud. New file, not an edit of any previous one — this lane had none, so this
is the first entry in the series.

## What we're trying to do

Make the cards on our sites render properly. The owner's original words were that the cards need
better designs; underneath that was a specific, mechanical defect. A row of article cards on a
listing page was rendering with four empty boxes in every card — no summary line, no date, no
category, no image caption. The job is to make those slots either carry real content or not appear
at all, across every site that uses the shared card component, and to do it through the framework
rather than by patching pages.

## Where we've come from

The cause turned out to be a disagreement between two pieces of code that both write the same kind
of list item for one shared component. One called the summary text `meta_description`; the template
read `excerpt`. One stripped the site name off the end of each headline and the other left it on.
So the card template was asking for things by names nothing was writing, and rendered empty
elements that still carried their layout, which is why the gaps were visible rather than invisible.

Both halves were fixed. The template half shipped as a database migration on 2026-09-02, and it
worked — verified at the served markup on ten of fourteen pages, with the other four correctly
refused by an unrelated safety guard that stops a page shrinking too far in one go. The code half
went into the running system on the same day and then appeared to do nothing at all.

That apparent nothing consumed two days. We could see the fixed code inside the running binary and
prove it with controls. We could dispatch a page re-render, watch it complete successfully, and read
back the old broken shape. Three automated diagnosis runs came back unable to confirm anything.
Sixteen candidate explanations were eliminated one at a time by measuring every affected page on
the estate — the component, the data source, the re-render reason, the binary, timing, locks,
version pinning, five database columns. Every elimination was correct. None was the answer.

## What we've done

We found the cause, and it was not this bug. Another session, coming at the same wall from a
completely different symptom, filed it as bug 454: when a long function was split in two on
2026-09-02, one value stopped being handed across the new boundary. A single missing line of
assignment. The consequence was estate-wide and almost perfectly silent — **every light page
re-render spent a fortnight rendering each page's own stored data back at itself** and reporting
success, with healthy counts, no errors, and nothing blanked. The only observable was an
improvement that failed to arrive. Their fix is one line, and it went live at 12:18 today.

Rather than take that on trust, we proved it on our own class of defect, and the choice of test
page was the whole game. The two experiments filed before this one had both run on pages that
already contained the fix's output — where "the fix ran" and "the fix did nothing and the existing
good data survived" produce identical bytes. We picked a page that was still broken, re-rendered
it, and watched the summary text appear, the site name drop off the headlines, and the change reach
the live page on the internet within seconds. First run, no ambiguity.

Then the estate began repairing itself while we wrote it up. Two pages on the darts site fixed
themselves on ordinary background traffic that nobody aimed at this problem, which is stronger
evidence than our own test because no one chose those pages.

We also resolved a question the delivery session raised about broken cards on the boxing site's
articles page. It is a different fault, already filed from this lane as bug 457: a piece of code
has been appending a duplicate, unowned row to that page on every run, and the page now serves six
stacked copies of the same six-article list — thirty-six cards where there should be six. Each copy
is frozen with the template of the day it was created, which is why some still show empty boxes.
Our counts matched their independent measurement of the live page exactly on both axes.

## Where we are now

The card defect's code half is closed, proven, and live. The population of affected card rows has
gone from five of seventeen correct yesterday to nine of seventeen at 15:05 today, and it keeps
improving on its own every time a page re-renders for any reason. Two further check pages are
queued to confirm on deliberately unlike cases.

The same missing line also explains the other thing this lane was chasing — pages wearing the wrong
hero image. Six page-heading components were declaring no image field at all; we fixed that on
2026-09-02 and were puzzled when only one page visibly improved. It was the same cause: the
re-renders that should have carried the change were refreshing nothing. That is fixed by the same
one line, and the other session's evidence shows a heading component recovering its own image on a
single re-render.

What remains genuinely open is bug 457, the duplicate rows on the boxing articles page. It needs a
code fix and a rebuild; no re-render can touch it, because those rows have no component attached to
refresh from. We have deliberately not deleted them by hand — it is a live paid site, and a code
fix is the right instrument.

## Where we're going

Very little, deliberately. There is no repair wave to schedule: the fleet's ordinary traffic is
draining the remaining cases, and a wave would spend dozens of dispatches to arrive at the same
place. We read the two queued check pages, confirm the hero class on the one hero test still in the
queue, and hand bug 457 to whoever owns that code path.

The lasting output of the two hard days is not the fix — that was one line, and somebody else wrote
it. It is two written-down checks that would have saved them. The first: to say that a piece of
code produced a value, you have to read the state it replaced, because "it wrote" and "the row has
the value" do not combine into "it wrote the value". The second: an experiment can only tell you
something if its starting point lacks the thing you are testing for. Both are now in the estate's
shared landmine and wrong-calls records, where the next session meets them before it picks a test
page rather than afterwards.
