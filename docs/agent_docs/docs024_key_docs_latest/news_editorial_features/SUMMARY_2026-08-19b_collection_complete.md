# Summary — news editorial features, 2026-08-19b

> **Supersedes `SUMMARY_2026-08-19_collection_pass.md`, written earlier the same
> session.** That one is kept, not corrected, because the difference between them
> is the finding: it was written after one grep pass that found three prior
> designs, and a second pass with different search terms found two more — one of
> which is the closest prior art in the repo and changes the recommendation.

## What we are trying to do

Editorial feature articles on the bigger current news stories, with the
background research filled out by real graphs and charts, assembled by extracting
the concepts out of news articles arriving on several different channels.

## Where we have come from

The idea has been designed **five** times and built none of them. Two in April
(the agent pipeline with a timeline table for continuity; the readership
segments, research-write-evaluate loop and the no-prediction boundary). One in
July, raised by the owner during news pooling: pick a topic a week, break it into
its parts, keep it updated until it stops mattering — plus the finding that
publishing one feature across 231 domains is *worse* for duplicate content than
duplicated headlines, so the research is shared once and each site generates its
own angle.

**And one on 28 July that the owner raised against the Thames Water dossier,
saying at the time it was generic for "this type of site and similar to the news
editorial requirement".** That is this ask, recorded three weeks ago. The first
search missed it because it calls itself neither editorial nor feature.

The earliest ancestor of all of them, from around March, already included a
near-duplicate headline detector. Cross-source grouping was in the design from the
start and was dropped on the way to what actually shipped.

## What we have done

Collected all five into one workstream, measured the live system rather than
trusting the documents, and put the central mechanism question through the
diagnosis loop, which returned **CONFIRMED** on all three clauses with citations.

Two of the three hard parts are already working. **Concept extraction runs
today** — topics tagged on 9,622 of 10,855 articles, with credibility and
provenance alongside. **Chart rendering is live and proven**, built so a chart
cannot state a number of its own: every value resolves to a registered fact and
every point in a series carries its own citation.

The diagnosis found a file the manual pass had missed. **We already detect when
several headlines are about one story — and we built it to throw that away.** A
display step suppresses the cluster so one story cannot dominate the feed, and its
author's own comment describes this workstream's problem exactly while solving the
opposite one: the pool must be big enough that "appears a lot" means "is the
subject matter", not "is one well-covered story". Four headlines on one subject is
what a feature is made of, and it is what that code deletes.

## Where we are now

The gap is everything between one article and a story. Three columns designed for
it are empty across all 10,855 rows and written by no code path; nothing
publishes; the timeline table was never created; every publishing agent the April
design named is absent.

But the July-28 note asked for something specific first, and **it is already
built.** It argued the blocker was the data, not the drawing — our facts each hold
one value and no sense of when that value applied, so a historical graph had
nothing honest to plot. It recommended four steps. Steps one and two, the series
data shape and the time-series chart component, both shipped the following day and
have served a real five-point series on a live page.

**So we are at step three of that recommendation: hand-build one worked example.**

The sharpest idea in the collection, and the one that should shape the build, is
the distinction between a topic and a premise. A topic produces an encyclopaedia
page nobody needs. A premise — a claim the article stands on — produces a page
with a reason to exist, and names what graph and what tool belong on it. The test
is mechanical: if this turned out to be false, would the main article change? That
matters directly here, because the tags we already extract are the topic kind:
good for spotting that four articles cover one story, useless for deciding what
the feature should argue. They are two different extractions and we have a weak
form of the first and none of the second.

## Where we are going

Hand-build one editorial feature, on one site, human-written, before designing any
pipeline — because the expensive step is almost certainly choosing what is worth
writing about, not generating the page, and guessing produces a lane that
automates the wrong thing.

Two constraints carry into whatever follows. The research substrate must be one
shared entity with the site-angle and the branch page as projections of it, or it
gets invented twice in incompatible ways. And anything this lane generates goes to
human review and never into automatic repair — there is an existing defect where a
tool failing its own quality check raises a repair job carrying the failing
criteria as the specification, and on one occasion satisfying it would have meant
deleting a legally required consent notice.

The fleet-wide version additionally needs the news pooling, which the owner parked
on 20 July with seventeen dormant pools. A single-site first example needs none of
it.
