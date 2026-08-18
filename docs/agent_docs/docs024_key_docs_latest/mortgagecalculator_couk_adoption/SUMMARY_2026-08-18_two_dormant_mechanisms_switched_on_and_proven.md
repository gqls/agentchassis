# SUMMARY — mortgagecalculator.co.uk — 2026-08-18

*Milestone read-out. Previous in the series: `SUMMARY_2026-08-09_improvement_loop_engaged_and_legislation_watched.md`.
Written to be read aloud.*

---

## What we're trying to do

Take a mortgage calculator site that already existed, and bring it fully inside the framework —
so that everything about it is built, checked and repaired by the machinery rather than by hand.
Not just "get the pages up", but get to the point where the system notices its own faults: a
broken link, a stale figure, a page that promises something that isn't there.

The site is the product, but it is also the proof. If the framework can hold a real site that
real people use, with tax figures that change and copy that has to sound human, then it can hold
the others.

## Where we've come from

The last read-out, nine days ago, was about the improvement loop being engaged and the legislation
watch being switched on. Since then the work has been the slow, unglamorous half of adoption:
brand assets, the house voice, the tools rebuilt and checked against independent arithmetic.

Two things shaped this period more than anything planned. First, a lot of the machinery that was
supposed to be watching the site had been quietly switched off — some of it months ago, some of it
in a cost pause a week earlier that nobody had revisited. Second, when we did look properly, the
site was in worse shape than the paperwork said: not broken, but drifting, with links that led
nowhere and a to-do list that had gone stale without anyone noticing.

## What we've done

**We stopped trusting the list and started measuring.** The top item on the handoff — thirty pages
supposedly serving the wrong browser-tab titles — turned out to have fixed itself days earlier.
Had we worked from the list, we would have rebuilt thirty pages to change nothing. Every figure in
the current handoff now carries the date it was taken.

**We found eight links on the live site that led nowhere**, and fixed the ones we could by building
the missing pages through the framework rather than quietly re-pointing the copy. Two new guides
are live and reading in the right voice. A third page still won't build, because of a platform bug
that another team owns — we contributed our evidence to it rather than working around it.

**We turned the link checker back on across the whole estate**, having first measured what that
would cost: one site an hour, working through a backlog of twenty-two. It immediately started
finding real damage — the first site with real content had four dead links of its own. That week
of silence had been hiding faults fleet-wide, not just here.

**We removed a stray design file** that had been sitting publicly downloadable since adoption,
flagged three times and never dealt with — and caught, just in time, that deleting it from the web
server alone would have been silently undone by the next sync from this machine.

**And we proved a new safety mechanism actually works.** Another team had built something that
watches whether the tax figures our stamp-duty calculator depends on have moved. Our tool was its
first customer. We seeded it, and then — with your permission — deliberately broke a figure for
fourteen seconds to see whether the alarm went off. It did, with the right before-and-after
values, and went quiet again afterwards exactly as designed. We then checked the calculator itself
recomputes from those figures, and proved that too by perturbing one and watching exactly one
answer move.

## Where we are now

The site is healthy and nothing is half-finished. Two mechanisms that were dormant are now
running: the fleet-wide link checker, and the stamp-duty figure watch. Both were switched on
deliberately, both were measured first, and both have a one-line off switch that is written down.

Three things are open and none of them is ours to decide:

- **The trailing-slash problem.** On our style of hosting, any address ending in a slash fails —
  `/guides/` is a 404 while `/guides/index.html` works. We found the cause, and the fix is three
  lines in one file. But that file serves all thirty-six sites, and a bad deploy of it takes every
  one of them down, so it needs you and a review, not a lane.
- **Thirteen small review items** the new figure-watch created on purpose. We have answered their
  question with evidence. Whether to formally close them is a judgement about a queue that a
  separate bug says nobody reads, so we left them and said why.
- **The blocked page.** It cannot be built until another team fixes their bug. Worth knowing: every
  new page the framework writes here adds another link to that missing page, because the site's own
  brief names it. The problem is slowly getting bigger on its own, which is the strongest argument
  for prioritising it.

We also got two things wrong this week and both are written down. We reported a working alarm as
dead, because the result was nested one level deeper than we looked and a neighbouring counter
appeared to agree. And we filed a "new" discovery that had been in our own traps file for weeks,
because we searched for it in our words rather than by the name of the thing.

## Where we're going

The near work is other people's bugs clearing: when the template bug is fixed, the missing page
builds and four dead links close themselves. When the image-linking bug is fixed, ten hero images
we already paid for become visible.

Our own next step is smaller and worth doing properly: the link checker will work through the
estate over the next day and file real repair work. Someone should watch what it finds — not
because it is wrong, but because it is the first honest picture of the fleet's condition we have
had in a week, and it will be worth reading before it is acted on.

The larger question is the one the trailing-slash finding raises. We have been checking whether
links point at pages that exist. We have not been checking whether the addresses we publish
actually work when a human types them. Those turn out to be different questions, and only one of
them was being asked.
