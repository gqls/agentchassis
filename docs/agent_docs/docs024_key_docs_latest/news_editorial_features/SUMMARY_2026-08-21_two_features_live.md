# Summary — news editorial features, 2026-08-21

> Third in the series. `SUMMARY_2026-08-19_collection_pass.md` was written mid-way
> through the first day's research; `SUMMARY_2026-08-19b_collection_complete.md`
> superseded it when the collection was finished and **nothing had been built**.
> This one is written because that is no longer true.

## What we are trying to do

Editorial feature articles: proper pieces about the bigger current news stories,
with the background filled out by real charts, assembled by extracting the
concepts out of news arriving on several different channels — and kept current
while the story matters.

## Where we have come from

The idea had been designed five times over five months and built none of them.
The last summary ended by saying the honest next step was not to design a
pipeline but to hand-build one worked example, because the expensive part is
choosing what is worth writing about, not generating the page.

## What we have done

**Two features are live**, on two different sites, both built through the
framework rather than hand-written.

The first reads this week's robot-demand coverage across four channels and sets
it against five years of installation figures, each point cited to the press
release it came from. The second does the same for professional darts: four
stories that read as a discipline row, set against the tournament calendar, which
turns out to have grown from thirty events a season to thirty-four. Both keep the
*dissenting* article in deliberately — the one whose headline contradicts the
others — because that is usually the most useful thing in the cluster and it is
what a feature can do that a headline list cannot.

Around them: heroes now carry a real image with a semi-transparent overlay, which
you ratified as the default. An **Insights** section sits in robot-hands' top
navigation. And the feed rows a feature draws on are now stamped with the page
they fed — a column designed for that purpose which, until this week, nothing in
the platform's history had ever written.

**The lifecycle policy you ratified is now the rule**, recorded as such: pages
stay up indefinitely at one URL, retirement is deliberate de-listing rather than
deletion, and refresh cadence is set per *fact* rather than per page.

The pattern is registered as `NEWS-020`, so another workstream can find it.

**A separate design lane exists** and has already shipped its first fix. We
measured the pages rather than designing from taste, found real contrast defects,
and fixed them: the two pages went from ten and eight failures to four and one.
What remains on the first is a pre-existing fault in a shared component that
belongs to another bug.

## Where we are now

The thing that has most changed is not the count of pages. It is that this work
is now being **checked by people who are not us**, and repeatedly caught.

Three other sessions found real defects in it this week. One noticed every feature
was quietly adding itself to the footer of every page on the site. One restored a
broken stylesheet and thereby proved that a measurement we had already acted on
was worthless — we had cleared a component of a fault using a page whose CSS
wasn't loading, and the false result looked like a pass. One found that the
sentence search engines print under the darts article was not a description at
all but a note-to-self about how to write the piece, which had been typed by hand
into a database seed and so had slipped past every automatic check the platform
has.

Each of those was invisible from where we were standing. All three are fixed.

The design work is in better shape than the plan that started it, because the
plan was wrong. It proposed choosing a single text colour that would be legible
on every site. That cannot work — some of these sites are dark and some are
light, and no one colour is readable on both. The framework already computes the
right colour per site, and had done since before we started; it was described in
our own index under a heading we read and did not open.

## Where we are going

Nothing here is blocked on a decision from you. Three things are open.

**Rollout.** Two sites are done. Six more have both a live feed and the evidence
register these pages need. The recipe is proven twice and takes about an hour a
site, and the expensive step remains choosing the story, which stays a human
judgement for now.

**One thing genuinely stuck.** The plan for components built out of other
components is still unwritten, because you asked for Fable to write it and Fable
has now hit its capacity limit four times. Nothing has been substituted. The
brief is ready and everything it needs to read is catalogued; it needs capacity,
not more thinking.

**One thing worth doing before it is needed.** Retiring a feature — the
de-listing half of the policy you ratified — needs two mechanisms that have never
been exercised together, and we know that only because a neighbouring lane built
half of one this week and told us what its limits were. It is cheap to prepare
now and expensive to discover during a retirement.
