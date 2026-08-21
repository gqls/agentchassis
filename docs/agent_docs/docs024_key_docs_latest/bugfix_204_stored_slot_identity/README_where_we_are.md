# Where we are — bug 204, the section names the planner keeps throwing away

Plain-prose log for the owner. Append only, newest at the bottom.

---

## 2026-08-21 — picking this up, and what I found

You asked me to look at bug 204. The first thing worth saying is that the headline
of that bug is already fixed, and has been since the 6th of August. So the question
was never "does this still bite" in the way the file's title suggests — it was "what
is left".

Here is the thing in plain terms.

When we build a site, each page is a list of **sections**. Most sections are named
after the kind of thing they are — `hero`, `faq`, `call-to-action`. But on sites we
took over from someone else, and then chopped their hand-written pages into pieces,
the sections are named by **position** instead: `prose-0`, `tool-1`, `prose-2`. The
name says "the first block of prose on this page", not "a component of type X". The
real identity of that block is a link stored in the database, pointing at whichever
component actually renders it.

Several parts of the platform look at a section name and try to work out which
component it means. They all did it the same way: check the catalogue of components
for one with that name. A positional name is not in the catalogue and never will be,
so the lookup fails. Back in August we fixed that in two places. The re-render path
and the page-build path both now look up the stored link first, and fall back to the
catalogue. Those work; I re-checked.

**What is left is a third place, and it behaves worse than the other two.** When the
site planner produces a new plan, a step called `validate_plan` checks every section
name the planner proposed and, if it cannot resolve one, **deletes it**. The other two
places merely postponed the section; this one removes it from the plan altogether. The
plan then gets written to the database, and the page's section list is overwritten with
the shortened version. The page keeps serving fine — the actual content rows are
untouched — but the record of what the page is *made of* is gone, so the next rebuild
has nothing to rebuild.

That is not theoretical. On the 20th of August, another session fired a replan at
loanandmortgagecalculator.co.uk to prove an unrelated fix, and **41 of that site's 45
live pages had their section lists emptied**. It was caught within the hour and put
back from a snapshot, so nothing broke in public. But the same run also queued 20 jobs
to "build" those now-empty pages, and had those been picked up they would have built
blank pages over live ones.

### The number that convinced me

Back in mid-August a different piece of work added a permanent record of every section
name this step throws away. It has been running since the 17th. In that time it has
recorded **140 discarded sections across 41 pages — and every single one of them is a
positional name.** Not one is a typo, a display name, or a stale component. The check
exists to catch mistakes of a kind that have not happened once; meanwhile the thing it
cannot see has accounted for one hundred per cent of what it threw away.

I like this number because it could easily have come out differently. If the check
were doing the job it claims, the list would be a mixture. It isn't a mixture.

### How much of the estate is exposed

Seven sites now carry section names that no component can resolve — 107 names in
total, up from 86 when the bug was filed. The bulk of it is
loanandmortgagecalculator.co.uk with 70 names across 41 pages. Six other sites carry
between two and eleven each. Any replan of any of them, today, would do what the 20th
of August did.

### What I have done so far

Checked that nobody else is working on it (two lanes have written *into* the bug file
recently, but both were reporting what they hit while doing something else, and both
said explicitly they were not taking it on). Confirmed the code is unchanged and the
setting that arms it is switched on for both planner agents. Filed the whole thing for
an independent second opinion through the diagnosis loop, rather than trusting my own
reading of it. And I have asked for a fix plan that covers the *whole* class rather
than the one site — because when I went looking, the same blind lookup turned out to
be used in **four** places, not one, and two of them write straight to live pages.

Next: the plan, then the review council, then the code.
