# SUMMARY — pooled news feeds for the domain portfolio

**As at 2026-07-19.** Design only; nothing built.

## What we're trying to do

Give most of a thousand-plus domain portfolio a credible, regularly-updated news
section without paying for each domain separately. The target is a small number of
shared feeds — roughly a dozen — that between them serve the whole portfolio.

## Where we've come from

The platform already has a working news pipeline: it fetches RSS, runs LLM news
searches, scores articles for relevance, renders a JSON news block onto the page
and can publish an outbound RSS feed. It was built and proven on single sites —
gaswholesalers.com first, then relojistas.com, where it now serves a feed at a
legacy URL that still gets pulled around 136 times a day.

But it was built site by site, and every layer of it is site-scoped. Each site owns
its own list of sources, fetches those sources itself, stores its own copy of the
articles and pays for its own LLM call to score them. Two sites in the same
industry pointing at the same RSS URL do all of that work twice and pay twice.
There is no shared layer anywhere in it.

## What we've done

Surveyed the pipeline in code and in the live database, and analysed the 1,625-domain
list the owner supplied.

We established the real current state, which was smaller than assumed: the news
pipeline runs on **four sites**, not the ~1,176 an early miscount suggested. That
miscount came from reading rows in a versioned table as if they were sites; it was
caught by cross-checking against the site count and is corrected in the notes.

We sorted the domain list by subject and, more importantly, by whether real news
exists for it. About two thirds of the portfolio sits in a vertical with a genuine
news stream. The other third — short brandable names and very narrow product sites
— has no news supply at any price.

We also confirmed the two technical facts the proposed design depends on: the
database already runs the pgvector extension, and the platform already generates
its own embeddings in-house for the knowledge base. That means the per-site part of
the work can be done with a database query rather than an API call.

## Where we are now

We have a design and a straw-man pool list, and no code or schema changes proposed
for approval yet.

The design separates the expensive work from the differentiating work. Fetching,
parsing, de-duplicating and enriching an article are identical no matter which
domain asked, so they happen once per article in a shared pool. Deciding which
articles matter to a particular site is the only genuinely per-site step, and it
runs as a vector query at no marginal cost. The result is that spend scales with
the number of news stories rather than the number of domains.

The straw-man pool list is thirteen pools covering about a thousand domains. The
largest by far is a single UK money pool — mortgages, loans, savings, credit,
insurance, pensions and investing all draw on the same stream of rate and
regulation news, and that one pool serves around 231 domains.

The correction to the original brief is that a dozen pools covers roughly two
thirds of the portfolio rather than the vast majority, because the remaining third
should not have a news feed at all. Giving those domains a feed would produce
visible filler rather than cheap news.

## Where we're going

Three things need owner input before anything is built. Whether the many
near-duplicate domains — eleven variants of "insurance" and so on — are meant to be
separate live sites or redirects. Whether per-site sources should survive as a paid
tier above the shared pools. And whether to start with the small number of domains
that actually have traffic, given that only a fifth of the portfolio has any views
at all and a single domain accounts for over a quarter of them.

Two existing bugs should be cleared before a fleet rollout rather than after. One
causes news pages to render nothing without JavaScript, which would defeat the
purpose if the feed is there for search visibility. The other hardcodes English in
the news listing, which matters for the Spanish and other non-English domains in
the portfolio.
