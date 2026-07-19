# relojistas.com — project summary

*A plain-language account of the relojistas.com rebuild. Written to be read aloud.*
Last updated: 19 July 2026. Technical entry point: `HANDOFF_RESUME_relojistas_rebuild.md`.

---

## What we're doing

We are taking a dead domain that still has a living audience, and building it back into
something that serves that audience — a Spanish-language watch news portal at
relojistas.com that feeds itself, and that answers the people who never stopped knocking.

## Where we've come from

relojistas.com used to be a busy Spanish watch forum. The forum died years ago and the
domain sat parked. Rather than guess what it was worth, we pointed it at a small server and
watched who still arrived.

The headline reading was discouraging. Over five and a half days it took more than four
hundred thousand requests, and about ninety-nine per cent of that was automated crawler
noise hammering the skeleton of the dead forum. Only eighty-nine requests reached anything
resembling a real homepage visit.

But two things were hiding in that noise, and they changed the decision entirely.

The first is the reason this project exists. **The old forum's news feed was still being
pulled around a hundred and thirty-six times a day** — steadily, by real subscription
services including Google's feed fetcher, Facebook's indexer and Apple's crawler. Real
people were still subscribed to that feed in their news readers, and every single request
was being answered with an error page. They had been knocking on a bricked-up door for
years.

The second was small but unmistakable: genuine human searches. People typing "Omega
Seamaster", "Certina 919", "Casio GA-2100" into the old site's search box, from Spain, Chile
and Mexico. Not many. But real.

So the goal became: rebuild the site in Spanish as a watch news portal, and reactivate that
dormant feed **at its original address**, so the people still subscribed simply start
receiving real content again. For now the feed links out to the original sources rather than
to us — the honest choice, and the legally safe one. The longer game is that a reactivated
audience is one we can eventually bring back to the site.

## What we've done

**The site is rebuilt and live.** relojistas.com is now a Spanish watch news portal, built
by our own platform from a written brief. It has a homepage carrying live news, a full news
archive, a page on the site's own history that acknowledges the old forum honestly, an about
page and a contact page.

**The news writes itself.** We found and verified five genuine Spanish watch magazines that
publish live feeds — Tiempo de Relojes, Debajo del Reloj, TR Magazine, Máquinas del Tiempo
and Relojes y Estilo — and paired them with an AI news search. Every six hours the system
fetches new articles, scores them for relevance and quality, and publishes the best. Nobody
touches it. Current stories cover Patek Philippe, Vacheron Constantin, Grand Seiko and
Zenith — exactly the right material for the audience.

**The dormant feed is answering again.** This is the result that matters. The old feed
address now returns a proper, valid news feed of thirty current Spanish watch stories, each
linking to its original source. The hundred-plus daily requests that had been hitting a dead
end for years are now being served real content, at the exact address those subscribers have
always used. We never asked anyone to re-subscribe. We simply started answering.

**And the news archive went live this week** — a full browsable archive of twenty stories,
which had been the last visibly broken thing on the site.

**We also fixed things well beyond this one domain.** Along the way we found that files for
this whole class of site were being sent to the wrong destination — a fault that would have
affected every similar site we build. We gave the platform the ability to publish outgoing
news feeds at all, which it simply could not do before. And we taught it to recognise the
watch and horology sector, which it had never heard of.

**One episode is worth telling on its own,** because it says something about how we work
now. Before committing a proposed platform fix, we put it through the review council the
team has just brought online. The council didn't rubber-stamp it. One reviewer told us to go
looking for an existing tool before building a new one — and that search turned up a
component that already did the job, better than our version would have. Another reviewer
spotted that our design would have quietly declared a bug fixed while leaving it broken in
production. We withdrew the proposal and used what already existed. The review cost about
two minutes and saved us shipping a duplicate with a hidden flaw.

## Where we are now

The site is live, the news updates itself, the feed is reactivated, and the archive works.

Three things are still visibly unfinished, and we know exactly what each needs. Two sections
— Guides and Glossary — are in the navigation but have no content behind them yet, so those
links lead nowhere. The homepage carries a few links the system invented to pages that were
never created; the platform already has a checker for exactly that, it just hasn't been run
on this site. And the search box that was capturing those genuine watch searches is
temporarily absent from the rebuilt homepage, so we've paused collecting that signal.

None of these are dangerous. All are visible, and they're next.

## Where we're going

**First, finish what a visitor sees.** Write real Spanish content for Guides and Glossary —
practical pieces on choosing and maintaining a watch, and a glossary explaining watchmaking
terms. This is the content that makes the site worth returning to between news cycles, and
the kind search engines reward. Then run the existing checker to clear the invented links.

**Then bring the search box back, improved.** Instead of silently recording what people
type, it will answer them, matching the query against our news and reference content. That
restores the demand signal while finally giving the visitor something in return.

**Then make the feed personal.** The old forum had separate subscriptions for individual
discussion boards, and we can still see which boards people subscribed to. We can map those
to matching topic feeds, so someone who once followed the Rolex board receives Rolex news
rather than a general digest.

**And then we measure.** The whole premise is that a dormant audience can be reactivated. We
can count that directly in the server logs — how many of those daily subscriber requests
turn from errors into satisfied ones — and that number tells us whether this domain deserves
more investment or has simply been given a dignified second life.

---

## The one-sentence version

A dead forum domain turned out to still have a live audience nobody was serving; we rebuilt
it as a Spanish watch news portal that now feeds that audience automatically at the address
they never stopped using, and we're finishing the last visible gaps before turning that
returning attention into something we can measure and grow.
