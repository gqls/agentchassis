# SUMMARY — pooled news feeds for the domain portfolio

**As at 2026-07-19, end of first session.** Design settled on the major points.
Nothing built, no schema changed, no code written.

## What we're trying to do

Give most of a thousand-plus domain portfolio a credible, regularly-updated news
section without paying for each domain separately. The target is a small number of
shared feeds — roughly a dozen — that between them serve the whole estate, with
each individual domain still showing something that suits its own audience rather
than the same six headlines as everyone else.

## Where we've come from

The platform already has a working news pipeline. It fetches RSS, runs searches
against a live-search language model, scores articles for relevance, renders a news
block onto the page and can publish an outbound feed. It was built and proven one
site at a time — gaswholesalers first, then relojistas, where it now serves a feed
at a legacy address that outside subscribers are still pulling around a hundred and
thirty times a day.

Because it grew site by site, every layer of it is per-site. Each site owns its own
list of sources, fetches those sources itself, keeps its own copy of the articles
and pays for its own language-model call to score them. Two sites in the same
industry pointing at the same feed do all of that twice and pay twice. There is no
shared layer anywhere in it, and at a few sites that has never mattered.

## What we've done

We surveyed the pipeline in code and against the live system, analysed the
domain list, and settled the architecture on every major point except one.

We established what is actually running, which was smaller than assumed: the news
pipeline is live on four sites, not the eleven hundred an early miscount of mine
suggested. We sorted the domain list by subject and, more usefully, by whether real
news exists for it at all. We got decisions from the owner on the three questions
that were blocking, and then went looking for machinery the platform already has so
we could lean on it rather than invent.

That last step earned its keep twice. It found several mechanisms we can reuse
almost as-is, and it caught a design of mine that was wrong — described below,
because it changed the shape of the build.

We also created a new place in the repository for work that is designed but not
yet built, because there wasn't one. Bug reports are for things broken in
production now, and filling that directory with future plans would have wrecked the
one question it exists to answer.

## Where we are now

The design splits the expensive work from the differentiating work. Fetching,
parsing, de-duplicating and researching an article produce an identical result no
matter which of our domains asked for it, so that happens once, in a shared pool.
Deciding which articles matter to a particular site is the only genuinely per-site
step, and it turns out we can do that one for nothing, because the database already
has vector search installed and we already run our own embedding model in-house.
The effect is that cost stops scaling with the number of domains and starts scaling
with the number of news stories in the world, which does not go up when we buy
another domain.

A pool will be implemented as a synthetic site — a fake site record that subscribes
to sources and ingests articles exactly the way a real site does. That follows a
pattern the platform already uses and has already hardened in production, and it
means almost nothing has to change: no new database columns, no changes to the
fetching or de-duplication code, and no alteration to a database index that has
caused a fleet-wide outage here once before. This replaced an earlier design of
mine that would have added new columns and allowed articles to have no owning site.
That version was coherent and I had a story for it, but it worked against the way
the platform does things, and I had proposed it before looking at how the platform
already solves that problem.

There will be roughly thirteen pools, covering about a thousand domains. The
largest by a wide margin is a single money pool — mortgages, loans, savings,
credit, insurance, pensions and investing all draw on the same stream of rate and
regulation news, and that one pool serves around two hundred and thirty domains.

The correction to the original brief is that a dozen pools covers roughly two
thirds of the portfolio rather than the vast majority, because the remaining third
should not have a news feed at all. About a third of the list is either short
brandable names or very narrow product sites, and there is no news about plastic
ducks at any price. Giving those domains a feed produces visible filler rather than
cheap news, which is worse than having no news section, and at portfolio scale it
undermines the one thing those sites have going for them.

Each domain is to be a genuinely separate site aimed at its own target market. That
raises the stakes on the per-site part of the design considerably: it stops being a
cost optimisation and becomes the thing that makes two similarly-named domains
different sites rather than the same site twice. We now know where those audience
descriptions should live, and most of the groundwork is already in place — site
records carry a free-form set of descriptive aspects, one for audience already
exists in a rough form, there is a piece of code in the admin interface plainly
written in anticipation of it, and every one of our current sites already has a
written description of who it is for. The rule for how a pool's default audience
relates to an individual site's own is one we already use for shared components: a
shared default that any site can take its own copy of, with the standing rule that
anything carrying the site's own voice has to be its own copy rather than an edit
to the shared one.

The one thing genuinely missing is a way to record that one of our sites is
positioned differently from another of our sites. Everywhere the platform
currently talks about what makes a site distinctive, it means distinctive against
outside competitors. That gap is small but it has to be filled deliberately.

Alongside the pools we have written up the owner's idea for packaged topic
features: pick a subject each week, break it into its parts, and build a proper
piece out of it that gets updated until it stops mattering. That is a better
instrument than a headline list and it is the strongest answer we have to the
duplication problem — but only if the research is shared once and each site
generates its own angle on top of it. Written once and published everywhere, it
would be worse than headlines, because near-identical long-form text is the shape
that gets punished hardest.

The duplication risk is written up properly, including how to measure it rather
than just worry about it. It is the kind of problem that is invisible from inside
any single site — every site looks fine on its own, and the defect only exists in
the relationship between them. Encouragingly, the check we need is close to one
that already runs: the platform already spots two sites accidentally sharing a
colour palette by comparing every site against every other one, and comparing news
blocks is the same shape with a different comparand.

## Where we're going

The immediate next step is to settle how a domain's audience gets described, since
everything else depends on it and it is now the only unresolved design question.
Deriving it from the domain name is the obvious approach and is precisely wrong,
because the names are nearly identical in exactly the families we most need to keep
apart. Two things to check first: a very similar design was written before and then
deliberately reversed, and nobody currently knows why; and the question "who is your
target audience" used to be asked when a site was set up and has since been dropped
from the form, which is part of why we are short of the information now.

After that the build order is the audience descriptions, then one pool as a pilot,
then the duplication measurement on that pilot before anything goes wider. Pulling
feeds back later does not undo the damage they would do to the estate's standing,
so the measurement has to come before the rollout rather than after it.

We start with the domains that have readers. Only about a fifth of the portfolio
has any traffic at all, and a single domain accounts for more than a quarter of the
whole estate's views, so spreading the first effort thinly across dormant domains
would tell us very little.

Two existing faults should be cleared before any fleet-wide rollout. One causes
news pages to show nothing at all unless the visitor's browser runs scripts, which
would defeat the purpose entirely if the feed is there to be found by search
engines; another thread is already on that one. The other hardcodes English in the
news listing, which matters because the portfolio includes Spanish, Dutch and
German domains.

A paid tier above the shared pools is agreed in principle and deliberately not
designed yet. The owner's point that it needs to be more than news is the crux:
dedicated news sources alone are not something a buyer can tell apart from the
free version, so that decision waits until the free tier is actually working.
