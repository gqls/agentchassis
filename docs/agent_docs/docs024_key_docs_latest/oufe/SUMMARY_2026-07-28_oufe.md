# SUMMARY — oufe.com, 28 July 2026

The third in the series. `SUMMARY_2026-07-26` and `SUMMARY_2026-07-27` stand as
written; this one records a real turn, because the site stopped being a
demonstration and became something with legal standing, verified behaviour, and a
substrate for the thing it was always meant to do.

---

## What we're trying to do

Build a specialist publication about how corporate finance actually works when a
company is under strain — restructuring plans, liability management, distressed
debt — written for the working mid-market professional. Restructuring advisers at
boutique firms, insolvency lawyers, private credit analysts. People who are
financially literate, short of time, sceptical of marketing language, and not
served by the trade press, which reports outcomes without mechanism.

The commercial idea is speed to synthesised clarity, and the differentiator is
interactive: not a PDF telling you who got paid, but a tool where you move the
assumptions yourself and watch who recovers and who is wiped out.

Underneath it sits a harder constraint that shapes everything. The site is
assembled with substantial machine assistance, and language models invent things
fluently. The most dangerous output is the plausible one: a real-looking figure in
a well-formed sentence next to a real citation. So the governing rule is that **a
figure reaches a page only by being registered with a source**, and the site says
plainly that it can be wrong. That is not modesty. On a site about named real
companies it is the product.

## Where we've come from

The site began as a long strategy conversation with another AI, which we unpicked
into a plan and then argued with. Several of its recommendations were wrong and
were recorded as wrong: the suggested "start with a distress radar" was the
highest-risk option available, not the lowest, and every figure it supplied about
Thames Water was an unverified recollection.

By the 27th we had two live sites, seven pages, a written Thames Water case
grounded in twenty-two quote-verified facts, and one working tool. We also had a
list of things the owner had asked for and not received: more tools, guides at
each point, a more readable layout, and infographics and graphs for the harder
concepts.

And we had a quieter problem. Several things were *believed* rather than checked.
The tool was described as working on the strength of its markup. A bug report
about accessibility carried a table of measurements that could not have been
right. A handoff asserted in bold that the platform had no chart renderer.

## What we've done

**Made the site legally coherent.** The disclaimer text approved on the 26th had
never actually been published — `/disclaimer`, `/terms` and `/privacy` all
returned "not found", and the footer's legal section was an empty box. So a site
publishing analysis of a named real company had no correction route a reader could
reach, and a live contact form with no privacy notice. Both pages are now live and
locked so no rebuild can rewrite them, linked from every page, alongside the
approved footer statement. Every claim in the privacy notice was checked against
the live site before being written: no cookies, no analytics, no third-party
scripts, and a contact form that opens the visitor's own email program rather than
sending anything to us.

The owner has no solicitor, so a decision was taken and recorded: publish the
approved text now, self-draft the privacy notice, and park the liability cap —
the one item genuinely needing legal review — until something is actually for
sale, which is also when it would be drafted properly.

**Proved the tool actually works.** The browser acceptance test we owed was run,
and it failed. The tool was fine; the test was impossible. It asked the consent
gate to be clicked twice on one page load, and the gate hides itself by design. We
changed only the test, not a line of the tool, and it now passes thirteen checks
across desktop and mobile — including one that had never once executed, so the
tool's most important behaviour had been untested the entire time it was described
as working.

**Found something dangerous while doing it.** That failure automatically raised a
job to *fix the tool*, handing a rewriting agent the impossible test as its
specification. The only ways to satisfy it were to weaken or delete the consent
gate, which is the legally load-bearing part of the page. It was cancelled by
hand. Nothing but attention prevented it, so it is filed as a platform bug.

**Made the case page readable.** The flagship Thames Water page was a single block
of text — the owner's "heavy textbook" complaint in its most literal form. It now
carries a mechanism diagram: seven steps from financial difficulty to a binding
court order, with a decision branch at the seventy-five per cent vote. Every step
paraphrases a fact already in the register, quote-verified with a source. Nothing
was written from memory.

**Corrected a bug report of our own that was wrong.** The accessibility table had
measured the wrong colour variable on every site but one, so on some sites it was
comparing a background against itself. Re-measured properly in a real browser, the
defect turned out to be bigger and elsewhere: the shared page header hardcodes
white text on each site's own accent colour, so legibility is luck, and five of
six sites fail. On one commercial site the main call-to-action is white text on a
white button and renders as a blank rectangle.

**Built the substrate for graphs, and then the graph.** The register could not
hold a time series at all: a fact carried one value and three dates, all of which
record when *we* did something, none of which is the date the value applies to. It
now holds dated observations, where each point carries its own source and can
never inherit one. The renderer followed, and each plotted point shows its own
citation beneath the chart.

**And discovered we already had chart renderers.** Two of them, one of which is
precisely the design we needed. The claim that we had none was our own, written
into a handoff in bold, and repeated to the owner twice.

## Where we are now

The site is eight pages, all live. The claims scanner reports zero findings across
sixteen components. Contrast passes on everything measurable. The tool is verified
in a real browser rather than asserted from its markup. The legal pages are
published and protected.

The first real chart is live on the Thames page, and getting there produced the
more interesting result. The task was to register a time series. The verified
figures do not form one: every debt figure sits at a single date, and the two
percentage figures measure different things — one is average yearly bills, the
other is an increase above inflation — so plotting them together would have
invented a trend out of unlike quantities. Forcing a series would have produced a
chart that looked authoritative and was analytically false.

So what went live is what the data actually supports: the capital structure as at
31 March 2024, drawn by class, scaled against the total drawn facilities — and
that total is itself a registered figure rather than a number typed into the
chart. Four figures, each quoted from a source that was fetched and checked
character by character rather than taken from a summary of it.

**The time-series renderer therefore remains built and unused**, which is the
correct state. Two primary sources tried today returned "forbidden" and "not
found", and a search engine's summary of a figure is not a verified quote. The
graph waits for data rather than being filled with something plausible.

Three things are owed to other people rather than to us: two sites with unreadable
buttons belong to other threads and were deliberately left alone rather than
fixed over the top of them, and the fleet-wide header defect needs a fix at the
generator rather than site by site.

The one unsatisfying thread is the review gate. The council caught a genuine and
subtle defect in our series work — a rule enforced in a validator that nothing
called, so it was not enforced at all — which we fixed. But the follow-up review
has failed three times with no verdict at all, in a way that looks like success to
anyone who does not go looking. We stopped after three runs and wrote up the
evidence rather than spending more.

## Where we're going

The immediate step is still the data, but the question is now sharper: not "get a
series" but "find a measurement that genuinely moves over time and that we can
reach a source for". Ofwat's determinations are the obvious candidate, and their
site refused us today. Until one lands, the time-series renderer stays unused and
the capital-structure chart carries the page.

Behind it sits the larger idea the owner set out — take a piece of writing, pull
out the premises it actually rests on, and branch the strongest one or two into
deeper pages with their own graphs and tools. Searching the backlog showed this is
already ours: it was raised in July as "packaged topic features", and what is being
described now is the same instrument turned ninety degrees, splitting one research
substrate across the pages of a site instead of across many sites. The advice
recorded from that analysis is to build one branch page by hand before designing
any workflow to generate them, because the expensive step is almost certainly
choosing *which* premise deserves a page, not writing the page.

The pattern worth carrying out of this week is smaller and more uncomfortable.
Almost every error corrected here was a claim we had written down ourselves and
then never re-derived — a warning in a handoff, a table in a bug report, a rule
stated in a validator. Each was stated confidently, which is exactly what stopped
anyone checking it. The things that caught them were cheap and dull: run the test,
open the browser, read the component library rather than the source code, query
your own correlation instead of the latest note.
