# Where we are — news editorial features

The owner's running log. Plain prose, append-only, newest at the bottom.

---

## 2026-08-19 — first session: what the archaeology turned up

You asked for editorial feature articles built around the big current news
stories, filled out with graphs and charts of background information, put
together by pulling the concepts out of news articles coming in from several
different channels. And you asked me to collect the past discussions first,
before building anything. That is what this session did.

**The short version: we have designed this three times and never built it. But
we are much closer than the documents suggest, because two of the three hard
parts turn out to be already working.**

### The three previous designs

Back in April, two documents were written within days of each other. The first
(`027g`) laid out the pipeline as a chain of named agents: cluster the day's
news into stories, research the background behind each one, write the analysis,
draw the charts, publish the page. It also proposed a table called
`event_timeline` to give the articles memory — so that the tenth piece about gas
prices knows what the first nine said, and can write "this is the fifth supply
disruption in four weeks" instead of starting from scratch every time.

The second (`036b`) was less about plumbing and more about quality. It made the
point that "gas industry professionals" is too broad an audience to write for,
and split it into procurement, operations, trading and strategy — the same event
matters to all four, but for different reasons. It also drew a line I think is
worth keeping: structured "if this, then that" analysis with the reasoning shown
is fine; confident predictions about where a price is going are not, and
something in the pipeline has to enforce that rather than trusting the writer.

Then in July, during the news-pooling work, you raised the idea again yourself —
pick a topic a week, break it into its parts, keep updating it until it stops
mattering. That was written up as `features_open/001`, and it added the piece the
April designs were missing: **if we write one feature and publish it to 231
domains, that is worse than duplicating headlines, not better.** Long-form
near-identical prose is the shape search engines punish hardest. The fix was the
same trick that makes pooled feeds work, one level up — do the expensive research
once and share it, then let each site generate its own angle on top. Same facts,
same citations, genuinely different articles.

### What is already working

This is the part that surprised me, and it is why I checked the live system
rather than trusting the documents.

**Pulling the concepts out of articles already happens.** Every item that comes
through the feed gets scored by a triage step that also tags it with its topics
— and it has done so on 9,622 of the 10,855 articles we hold. It records where
each story originally came from, how we found it, and how credible the source
tier is. So the raw material for "what is this article about" is there, at
volume, today.

**We can already draw charts, properly.** There are three live chart components
— a bar chart, a time series, and a diagram for explaining how a process works.
They are strict in a way that matters here: a chart cannot state a number of its
own. Every plotted value has to resolve back to a registered fact, and every
point in a time series carries its own citation rather than borrowing the one
above it. That is exactly the discipline you would want on a feature article
full of background figures.

I nearly got this wrong. One of our own indexes still says the chart capability
was never started — it is simply out of date, because the charts are database
rows rather than code, and the usual way of searching for code cannot see them.
A previous session made that exact mistake in July, told you twice that there was
no chart renderer, and was wrong both times. I checked the live system instead.

### What is genuinely missing

Everything between one article and a story.

The feed table has three columns that were clearly designed for this and are
completely empty: one to mark that two articles are the same story, one to
record which entities an article mentions, and one to point at the page we
published from it. All three are zero across all 10,855 rows — and nothing in
the codebase writes to any of them. So the same event arriving from four
different channels is currently four unrelated rows, which is precisely the
"different channels" part of what you asked for.

Nor does anything publish. Every agent the April design named — the rewriter, the
publisher, the analyst, the researcher, the writer — is absent. The pipeline
genuinely does end at curation, as one of our own notes put it. And the
`event_timeline` table that was supposed to give the articles memory was never
created.

### One thing that gates the fleet-wide version

The per-site angle needs to know who each site is written for, and that audience
profile comes out of the pooling work — which you parked on 20 July, with the
seventeen pools sitting dormant and costing nothing. That parking is still in
force and I have not touched it.

But that only gates the *fleet* version. A first editorial feature on a single
site that already has a news feed and an evidence base needs no pool and no angle
layer at all — and both April designs independently argued for exactly that build
order: do not build this tier until the previous one has proven the writing is
good enough. That is the cheapest way to find out whether the output is any good
before we spend anything on scale.

### What I have run

I put a diagnosis run into the loop on the one question worth being sure about
before designing anything — where exactly the pipeline stops, and what those
three empty columns were meant to be for. Its verdict goes into the notes when it
lands. I have built nothing.

### What I need from you

Nothing blocking. But when you are ready to move to stage two, the one decision
that shapes everything else is whether we start with a single site to prove the
writing, or wait for the pooling to come off its pause and do it properly across
a pool. My recommendation is the single site, for the reason both April documents
gave.
