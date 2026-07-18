# relojistas.com — project summary

*A plain-language status of the relojistas.com rebuild. Written to be read aloud.*
Last updated: 18 July 2026.

---

## What we set out to achieve

relojistas.com used to be a busy Spanish-language watch forum. The forum died years ago and
the domain sat parked. Before deciding its fate, we pointed it at a small server and simply
watched who still knocked on the door.

The final traffic reading was, on the face of it, discouraging. Over five and a half days the
domain took more than four hundred thousand requests, but around ninety-nine per cent of that
was automated crawler noise hammering the skeleton of the dead forum. Only eighty-nine of
those requests reached what you could call a real homepage visit.

But buried in that noise were two genuinely valuable things.

The first, and the reason this project exists: **the old forum's news feed is still being
pulled around a hundred and thirty-six times a day** — steadily, by real subscription
services including Google's feed fetcher, Facebook's indexer and Apple's crawler. Somewhere
out there, real people are still subscribed to this feed in their news readers. Every single
one of those requests was being answered with an error page. They had been knocking on a
bricked-up door for years.

The second: a small but real trickle of genuine human searches — people typing things like
"Omega Seamaster", "Certina 919" and "Casio GA-2100" into the site's search box, from Spain,
Chile and Mexico. Not many, but unmistakably real people looking for watches.

So the goal became: **rebuild relojistas.com as a Spanish-language watch news portal, and
reactivate that dormant feed at its original address** so the people still subscribed start
receiving real content again. For now the feed links out to the original news sources rather
than to us — that is both the honest thing to do and the legally safe one. The longer game is
that a reactivated audience is one we can eventually bring back to the site itself.

---

## Where we are now

**The site is live and rebuilt.** relojistas.com is now a Spanish watch news portal, built
automatically by our platform from a written brief. It has a homepage carrying live news, a
page on the site's own history that acknowledges the old forum honestly, an about page and a
contact page — all in Spanish, aimed at watch enthusiasts in Spain and Latin America.

**The news writes itself.** We found and verified five genuine Spanish watch magazines that
publish live feeds — among them Tiempo de Relojes, Debajo del Reloj and TR Magazine — and
paired them with an AI-powered news search. Every six hours the system now fetches new
articles, scores them for relevance and quality, and publishes the best of them. This runs
without anyone touching it. Current headlines cover Patek Philippe, Vacheron Constantin,
Grand Seiko and Zenith — exactly the right material.

**The dormant feed is answering again.** This is the headline result. The old feed address
now returns a proper, valid news feed containing thirty current Spanish watch stories, each
linking out to its original source. The hundred-plus daily requests that had been hitting a
dead end for years are now being served real content, at the exact address those subscribers
have always used. We did not ask anyone to re-subscribe; we simply started answering.

**Some platform improvements came out of it.** Along the way we found and fixed a significant
routing fault that was sending this kind of site's files to the wrong destination — a fault
that would have affected every similar site we build, not just this one. We also gave the
platform the ability to publish outgoing news feeds at all, which it could not do before, and
taught it to recognise the watch and horology sector.

**What is not finished, honestly stated.** Three links in the site's navigation — News,
Guides and Glossary — still lead nowhere, because those pages have not been built yet. The
homepage also carries a few links that the system invented to pages that were never created.
And the search box that was capturing those genuine watch searches is temporarily absent
from the new homepage. None of these are broken in a dangerous way, but they are visible and
they are next.

---

## What we aim to do next

**First, finish the navigation.** The News archive page is being built as this is written —
it will give the site a full, browsable news archive rather than just the headlines on the
homepage. Then we clear up the invented homepage links so nothing leads to a dead end.

**Then, give Guides and Glossary real substance.** We will write genuine Spanish content —
practical guides on choosing and maintaining a watch, and a glossary explaining watchmaking
terms. These are the pages that make the site worth returning to between news cycles, and
they are the kind of content that search engines reward.

**Then, bring back the search box — improved.** Rather than silently recording what people
type, it will actually answer them, matching their query against our news and reference
content. That restores the demand signal we were gathering while finally giving the visitor
something useful in return.

**After that, the refinements.** The old forum's feed had separate subscriptions for
individual discussion boards, and we can still see which ones people subscribed to. We can
map those to matching topic feeds, so someone who once followed the Rolex board receives
Rolex news rather than a general digest. And we will keep watching the logs to measure the
reactivation directly — counting how many of those daily subscriber requests turn into
satisfied ones.

---

## The one-sentence version

A dead forum domain turned out to still have a live audience nobody was serving; we rebuilt
it as a Spanish watch news portal that now feeds that audience automatically, and we are
finishing the remaining pages before turning the trickle of returning attention into
something we can measure and grow.
