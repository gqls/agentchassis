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

### Later the same day — the diagnosis came back, and it found something I had missed

The run confirmed all of it: the items really are handled one at a time, the
three columns really are written by nothing, and the topic tagging really is
running in production. It checked that last one properly — not just that the code
exists, but that the step ran successfully this morning.

But it opened a file I never did, and that file changes the picture in a way
worth telling you about.

**We already have something that spots when several headlines are about the same
story. We built it to throw that away.** There is a step in the news display code
that watches for one subject dominating the feed and quietly drops the extra
items so the page does not show four versions of the same thing. It is careful
work — it only kicks in once there are at least twelve articles to judge from,
because below that a genuine cluster looks the same as coincidence.

The comment its author left is almost a description of what you have asked for,
written by someone solving the opposite problem: it says the pool has to be big
enough that "appears a lot" means "is the subject matter", not "is one
well-covered story" — and gives four headlines about the same subject as the case
it must not mistake.

Four headlines about the same subject is exactly what a feature article is made
of. So the signal we need is already being detected; it is just being deleted at
the last moment, by a feature that is right to delete it. That is better news
than finding nothing — someone has already tuned the hard part — but it does mean
we cannot simply reuse that code where it sits, because the page it serves
depends on it doing the opposite.

One smaller thing worth knowing, because it would mislead anyone reading our own
documentation: the ingestion step is described internally as writing feed items
"with dedup", which sounds like it already merges the same story from different
sources. It does not. It only skips an article whose exact web address we have
already stored for that site. Reuters and the BBC covering one story are two
separate rows, by design.

### And then a second search turned up the most important document of all

I went back with different search words, and found two more discussions — one of
which is closer to what you asked me for today than anything else in the repo.

**On 28 July you raised this same idea against the Thames Water dossier**, and
you said at the time it was generic for "this type of site and similar to the
news editorial requirement". The note that came out of that conversation is
`docs/agent_docs/docs024_key_docs_latest/oufe/DESIGN_2026-07-28_premise_branching_and_deepthink.md`.
My first search missed it because it calls itself neither "editorial" nor
"feature article".

It contains the sharpest idea in the whole collection, and I think it is the one
that should shape what we build.

**The distinction is between a topic and a premise.** A topic — "the
stakeholders in the restructuring" — produces an encyclopaedia page. Nobody needs
it; it is a worse Wikipedia. A premise — "the outcome turns on whether the court
accepts this particular valuation" — produces a page with a reason to exist, and
it tells you what graph and what tool belong on it, because a premise is a claim
that can be tested and a tool is a way to test it.

The test for whether something is a premise is mechanical: **if this turned out
to be false, would the main article change?** If not, it is background, and it
belongs in a sentence rather than a page. That also answers "which one or two
points do we branch on" by ranking rather than by taste.

This matters directly for what you asked. "Extracting the concepts out of the
articles" could mean either thing. The tags we already produce on 9,622 articles
are the topic kind — useful for spotting that four articles are about the same
story, useless for deciding what the feature should argue. **Those are two
different extractions, and we have a weak version of the first and none of the
second.**

That note also spotted something I would have got wrong: the July idea spreads
one topic across many *sites*, and the Thames idea spreads one topic across many
*pages of one site*. They are the same feature seen from two directions — and
what you have asked for today wants both at once. The warning attached is that if
we build either half without noticing, the shared research gets invented twice in
incompatible ways.

**The best news is at the end of it.** That note said the real blocker was not
the drawing but the data — that our facts each carry a single value and no sense
of *when* that value applied, so a historical graph had nothing honest to plot,
and building the chart first would just tempt a writer to make the numbers up. It
recommended four steps, in order.

I checked the live system, and **the first two of those four are done.** The
series data shape was built and one site has a real one registered; the
time-series chart component is live and has served a real five-point series on a
public page. Both landed the day after that note was written.

So we are not at "design a pipeline". We are at **step three: hand-build one
worked example**, human-written, so we learn what good looks like before
automating anything. That was the recommendation then and everything I checked
today still supports it.

One last thing worth flagging, because it is a rule rather than a fact. There is
an existing defect where a tool that fails its own quality check automatically
raises a repair job carrying the failing criteria as the specification — and on
one occasion the only way to satisfy that was to delete a legally required
consent notice. It was caught by hand. That is fine while a person is watching
one tool at a time. A workflow generating pages and tools automatically, across
premises and sites, takes the person out of exactly that position. So the rule
the earlier note proposed, and I would keep: **anything this lane generates goes
to human review, and never into automatic repair.**

## 2026-08-19, later — the first editorial feature is live

The worked example is up: **https://robot-hands.com/insights/robot-demand-step-change.html**

I picked robot-hands rather than the gas site, for one decisive reason: our
charts refuse to draw a number that isn't backed by a registered, cited fact —
and the gas site has no fact register yet, so it would have produced a feature
with no charts at all.

The story chose itself from the site's own feed. The robot-demand story is on
at least four channels this week: one outlet framed record installations around
the labour shortage; a wire release about rising orders was carried, almost
word for word, by two different outlets — which is exactly the "same story,
different channels" signal this whole project is about; and one trade journal
ran the opposite headline, saying US orders were weak. I kept that dissenting
one in deliberately, because the background chart explains it: the Americas take
under a tenth of the world's new robots, so a soft American quarter and a strong
world are both true at once. That is what an editorial feature can do that a
headline list cannot.

The background is charted from the robot industry's own statistics body — five
years of worldwide installations, every point carrying its own citation to the
press release it came from, each one fetched and checked word-for-word before I
registered it. The page says what the series actually shows, which is more
interesting than the headlines: not a boom, but one sharp step up followed by
four years holding above half a million — a plateau at altitude. For this site's
readers the closing turn is that every one of those installations ends in a
gripper somebody had to specify, which hands off to the site's own selection
tools.

Everything went through the framework: the facts into the site's evidence
register, the page assembled from existing components, the honesty scan run
before deploy (clean), and the footer link requested through the framework's own
navigation queue rather than patched by hand.

Two things I'd note for next time. Writing the prose was quick; **choosing the
story and deciding what it actually shows was nearly all of the work** — which
was the bet, and it held, and it is what any future automation has to be
designed around. And the update question is now concrete: this page's series
moves once a year (each September, when the new report lands), while the
orders-and-earnings coverage it discusses moves quarterly — one page, two
natural clocks, which is why the policy proposal ties cadence to the fact, not
the page.

**Waiting on your say-so:** the lifecycle policy in
`DESIGN_2026-08-19_starting_point.md` §3 — stable URLs updated in place,
retirement by de-listing rather than deletion, cadence per fact. If you're
happy with it, it becomes the rule for every page this workstream ships.
