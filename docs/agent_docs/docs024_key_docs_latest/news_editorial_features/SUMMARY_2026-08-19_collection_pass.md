# Summary — news editorial features, 2026-08-19

## What we are trying to do

Build editorial feature articles: proper pieces about the bigger current news
stories, with the background research filled out by real graphs and charts, put
together by extracting the concepts out of news articles arriving from several
different channels and working out which of them are the same story.

Not a headline list. The feed we already run answers "what happened today". This
answers "what is going on, and what does it mean for you" — a different
instrument, with a different half-life and different value.

## Where we have come from

The idea has been designed three times and built none of them.

In April, two documents in quick succession. One laid out the pipeline as a chain
of agents — cluster the day's news into stories, research each one's background,
write the analysis, draw the charts, publish — and proposed a table to give the
articles memory across weeks, so the tenth piece on a subject knows what the
first nine said. The other worked on quality rather than plumbing: split a broad
trade audience into the four segments that actually read differently, put an
evaluation step between writing and publishing, and drew a hard line between
structured "if this, then that" analysis with the reasoning shown, which is
allowed, and confident directional prediction, which is not.

In July, during the news-pooling design, the owner raised it again — pick a topic
a week, break it into its parts, keep updating it until it stops mattering. That
round added the piece the April work lacked: one feature published across 231
domains is *worse* for duplicate content than duplicated headlines, because
long-form near-identical prose is the shape that gets punished hardest. The
answer is to do the expensive research once and share it, then generate each
site's own angle on top — same facts and citations, genuinely different articles.

Since then the platform itself has moved underneath all three designs, which is
why this collection pass was worth doing before any building.

## What we have done

Collected every past discussion into one workstream, and — more usefully — went
and measured what is actually live rather than trusting the documents.

Two of the three hard parts are already working. **Concept extraction runs
today**: the triage step tags every incoming article with its topics, and has
done so on 9,622 of the 10,855 articles we hold, alongside a credibility rating
and a record of where the story originally came from. **Chart rendering exists
and is proven**: three live components — a bar chart, a time series and a process
diagram — built so that a chart cannot state a number of its own. Every plotted
value resolves back to a registered fact, and every point in a series carries its
own citation rather than inheriting one. That is exactly the discipline a feature
article full of background figures needs.

One of our own indexes still records the chart capability as never started. It is
stale, because the charts are database rows rather than code and the usual code
search cannot see them — the same mistake a session made in July, when it told
the owner twice that no chart renderer existed while two were live. We checked
the live system instead.

We also filed a diagnosis run on the one question worth being certain about
before designing anything: where exactly the pipeline stops, and what three
conspicuously empty columns were meant to be for.

## Where we are now

The gap is not the concepts and not the charts. It is everything between one
article and a story.

Three columns in the feed table were clearly designed for this and are completely
empty across all 10,855 rows — one marking that two articles are the same story,
one recording the entities an article mentions, one pointing at the page we
published from it. Nothing in the codebase writes to any of them. So a story
arriving from four channels is four unrelated rows, and the "different channels"
half of the ask has no mechanism at all.

Nothing publishes, either. Every agent the April design named is absent from the
live roster, and the timeline table meant to give articles memory was never
created. Our own note put it plainly at the time: the pipeline ends at curation.

The fleet-wide version is gated on the news pooling, which the owner parked on
20 July with seventeen dormant pools costing nothing. That parking still holds
and we have not touched it.

## Where we are going

Stage two is a design, not a build, and it turns on one decision: prove the
writing on a single site first, or wait for pooling to come off its pause and do
it across a pool properly.

The recommendation is the single site. It needs no pool and no per-site angle
layer, it uses machinery that already exists, and it is the build order both
April documents independently argued for — do not build this tier until the
previous one has proven the writing is good enough. It is the cheapest way to
find out whether the output is any good before spending anything on scale.

After that decision, the substantive work is the grouping step: turning free-text
topic tags on individual articles into a story that several channels are covering
at once. That is the piece nobody has built, and everything else hangs off it.
