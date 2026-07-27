# SUMMARY — oufe.com and oxenunity.com, 2026-07-26

The milestone read-out. Two days old as a workstream; both sites are live.

---

## What we are trying to do

Build a publication about how corporate finance actually works when a company is
under strain — restructuring, distressed debt, liability management, the tactics
that decide who gets paid and who does not. The focus is the United Kingdom,
where the machinery is unusually public: the courts publish, and in regulated
sectors the regulator publishes too.

The audience is the working professional in the middle of the market — the
boutique restructuring adviser, the corporate finance lawyer, the private credit
analyst, the person early in their career who needs to understand the mechanism
rather than read the headline. Not the large funds, who already buy the expensive
services and are not short of information.

What we offer is not more information. It is mechanism, explained clearly, and
tools that let a reader move the assumptions and watch who wins and who loses.

Alongside it, **oxenunity.com**: a single page carrying the name, pointing at the
work. There is no company, and the page makes no claim that there is.

## Where we have come from

The proposition was worked out in a long conversation with another AI, which then
could not export its own record of it — three failed attempts, and a fourth that
quietly dropped the strategy and printed its own source code. So the first job
was salvage: reconstructing what had been decided, and separating the owner's
decisions from the suggestions he had rejected or reordered.

That salvage turned up three things the conversation had been built on that were
no longer true, all of which changed what we built. A verification capability
described as unbuilt was in fact live. A charting library the architecture
assumed does not exist. And most consequentially, our automatic check for
invented numbers is close to useless on financial prose — it looks for numbers
near words like *clients* and *awards*, has no concept of a creditor, and skips
currency amounts entirely. On a site made of pound figures, that check reports
"clean" and means almost nothing.

We also disagreed with the owner's stated starting point, and said so. He had
ruled that an automated distress radar was the lowest-risk first move; it is the
highest, because we hold no market data, UK court listings have no feed, and a
distress signal is a factual claim about a named real company — the same shape as
the worst mistake this platform has made. We built the flagship-plus-tool path
instead, which he had separately called the primary magnet.

## What we have done

Both sites are live and verified.

The oufe build produced exactly the five pages we specified and no others, which
is the roadmap machinery working as intended rather than a happy accident. More
importantly, **not a single invented figure reached the site** — no currency
amount, no percentage, no statistic, across fifty thousand characters of
generated specification and every live page. Keeping numbers out of the briefs
entirely was the control that did it.

Then the site made a mistake worth the whole exercise. Its copy independently
wrote a promise of infallibility — *"a claim without a named, dated source does
not appear here"* — hours after the owner had struck a weaker version of exactly
that line. It was live on two pages before we caught it.

Nothing caught it but a person reading the page — and my first account of *why*
was wrong in an instructive way.

> **CORRECTED 2026-07-27.** I concluded the class was structurally invisible to
> every check we own. It is not. The banned-claim scanner is an ordinary text
> search over prose and would have caught all four phrases — **if anyone had ever
> written a pattern for this class on any site, which nobody had.** Only five of
> our fifteen live sites carry a single such pattern at all, and there is no way
> to write one once for the whole fleet. The gap was reach, not capability. The
> owner caught the error by refusing the build-something-new answer and telling
> me to look harder at what already exists.

So we made the lesson permanent in three places. The copy now says we cite
everything **so you can check us**, and that we can still be wrong — a source can
be wrong, and our reading of it can be wrong. The compliance seat on the review
council now looks for the class specifically, with the four phrases we shipped
quoted in its brief, and is asked to suggest the honest wording rather than only
name the fault. And both content writers carry a standing rule against promising
accuracy they cannot guarantee, so it shapes copy at the point of writing.

Alongside that we shipped the owner's framing as a general rule: **lead with
mechanism, treat real cases as clearly-marked illustration** — a possibly
inaccurate case study, never a definitive account. The insight underneath it is
that risk concentrates in asserting live facts about named companies while value
concentrates in explaining mechanism, and those two are separable. A mechanism
taught with openly hypothetical figures is completely accurate and cannot be
wrong about anybody.

The interactive tool now asks the reader to acknowledge that it can give a wrong
answer before it will work, and carries that caveat inside its own results, so it
travels when someone screenshots the output into a deck.

We also repaired what the first build got wrong: six broken links, including the
site's own navigation, and default commercial furniture — a "Get Started" button
and an "Our Services" heading — on a publication that sells nothing. The site now
has zero broken links.

Finally, and this is the piece with the longest reach, we built a **general
capability the whole fleet can use**: a content lane for pages that must not be
written from memory. It searches for sources rather than recalling them, requires
a verbatim quote for every claim, re-fetches each source and throws away any
quote that is not really there, writes only from what survived, audits its own
draft for anything it cannot trace, and then stops at human review. It has no
setting that lets it publish. That last part is deliberate: a content robot that
can publish will eventually publish something wrong while nobody is watching.

## Where we are now

Both domains serve. Home, about, cases and contact are live on oufe.com, with no
broken links and no unverified figure anywhere on them. The evidence register is
attached and deliberately empty, so the writers currently have no numbers they
are permitted to assert — which is the correct state for a site that has not yet
done its research.

The first grounded explainer is running now, on how a restructuring plan can bind
a class that voted against it. It is the first real exercise of the new lane, and
also the first real exercise of the citation-verification machinery, which had
been live for six days without ever completing a run.

Four platform defects found along the way are filed: a shared deployment script
whose documented branch cannot work, a page-render failure that reports success
while doing nothing, a dispatch queue where one long review run blocks everything
behind it, and a link checker that misses links inside content cards.

## Where we are going

The immediate work is the rest of the mechanism explainers — the creditor
waterfall, and the statutory framework — through the same grounded lane, and then
the Thames Water dossier, which needs its evidence gathered before a word of it
is written.

Two things need the owner. The standing disclaimer wording is drafted and needs
approval, because it has to sit *with* the content rather than in a footer: for
research someone relies on, the real exposure is not a regulator but a person who
acted on our page. And there is a live strategic question about audience — the
owner asked whether aiming at students would be safer. Our answer is that the
safety instinct is right but the narrowing is not needed to get it, because
professionals already expect "check this yourself", while narrowing to students
removes the ability to charge for anything. That is his call, and the site is at
the cheapest possible moment to make it.

Further out sits everything the owner planned and we deliberately deferred: the
deal packs, the subscriptions, the private data workspaces, the radar. None of it
is built and none of it should be until the site has proved it can keep one
flagship case genuinely current. That was his own instinct at the start — prove
regular value before asking anyone for money — and nothing since has argued
against it.
