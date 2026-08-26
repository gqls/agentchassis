# The pages we retired and never took down

A read-out on `bugs_open/359`, written to be read aloud.

## What we're trying to do

Make the platform notice when it is still publishing something it decided to stop
publishing.

When we retire a page — because it was wrong, or superseded, or made a claim we could not
evidence — we mark it "archived" in the database. That mark does two useful things and
fails to do a third. It takes the page out of every menu, list and rebuild. It does not
take the file off the internet. Deleting the file is a separate action, and we built that
action back in early August.

Nothing has ever checked that it ran.

## Where we've come from

The gap was noticed on the 22nd of August by the thread closing a different piece of work,
which filed it, sampled three sites, found three retired pages still answering the public,
and moved on. Nobody picked it up. Meanwhile one of those pages — robot-hands.com's
gripper catalogue — had already been recorded as live-after-retirement on the 14th.

The reason nobody noticed is worth stating, because it is the shape of the whole class.
From inside the site, a retired page has genuinely gone: it is in no menu and nothing links
to it. It is only reachable from the outside — by a direct link, or by a search engine that
indexed it before we retired it. Every signal we had said the site was correct, and it was.

## What we've done

**Measured it properly first.** I re-ran the census across the whole fleet rather than
trusting four-day-old figures. **39 retired pages had been published at some point; 7 were
still serving.** More usefully: two of the three pages the original report named had since
gone, and five of my seven were new. **This is a flow, not a backlog.** Cleaning up the
seven by hand would have looked like a fix and changed nothing.

That census is now a command anyone can run, with the controls built into it rather than
written down beside it.

**Built the meter.** A check that rides the routine health sweep every site already gets —
roughly every four hours — and asks, of each retired page, whether the public still gets it.
It raises a flag for a person. It does not delete anything, deliberately: a page retired *by
mistake* that is serving perfectly well looks identical, from the outside, to a page retired
on purpose that should have gone. The first needs un-retiring, the second needs deleting,
and quietly getting that backwards would take a good page off a customer's website.

**And made it refuse to guess.** This is the part I'd want to explain if asked what was
interesting about the work. Every other check of this kind reports a problem — a broken
link, a dead site. When one of those goes blind it finds fewer problems than there are:
annoying, but honest. This check's "problem" is a page that *works*, so if the site happens
to be down while it runs, every retired page looks correctly gone and it reports nothing
wrong — which is exactly what it reports when everything is fine. So before it judges
anything it proves its own instrument twice: it asks for a page that cannot exist and
requires a refusal, and asks for a page it knows is live and requires an answer. If either
fails it declines, records that it declined, and is structurally prevented from marking any
existing flag as resolved. A blind run cannot tidy away real problems.

## Where we are now

Built, tested, and through the reviewer council — **approved first time**, with two advisory
objections and no vetoes, across seventeen reviewers.

The approval is not the interesting part. **Three of the objections found real defects**, and
all three are fixed:

- A reviewer pointed out I was deciding which sites to examine using a status field the
  platform treats as decorative. I had copied that from a sibling check. When I went and
  counted, one of the two values I was filtering on **does not exist anywhere in the fleet** —
  dead code — and the filter would silently have excluded any site at a status nobody has
  invented yet, from the only check that looks at this. Inverted: it now excludes one known
  class and lets everything else through.
- Another spotted that I had moved a safety guard into shared code and tested it only from
  the new side — while on the old side that same guard decides whether a live page's file
  gets deleted. Fair, and it was in my plan and missing from my commit. Now covered, three
  ways, each proven by breaking it.
- A third caught that my database change takes its backup *before* checking whether it has
  already run — so a second run would save a backup labelled "before" that actually contains
  "after". That is a documented trap, and I inherited it by copying two migrations that both
  have it. Fixed here, and written up so the next person copying them sees it.

I also got something wrong on my own account and it is in the record: the test file's table
of "break this line and that test will fail" was written from the design rather than from
running it, and when I did run it, **the two most important rows were false**. Both tests
were passing for the wrong reason. Fixed, and written up, because a table like that is proof
to whoever reads it next.

## Where we're going

The code is committed but switched off. Turning it on early would break two other working
checks — a check name the running software doesn't recognise fails the whole sweep — so the
switch is held until the next server update carries it.

After that update, in order: confirm the running software actually has the check (by asking
it, with a control, not by trusting a version number); turn it on; ask *what did I break*
before asking what worked; and then the real test — it must flag a page we know is live, and
**not** flag one we know is gone. Until it has done both, a quiet result from it means
nothing at all.
