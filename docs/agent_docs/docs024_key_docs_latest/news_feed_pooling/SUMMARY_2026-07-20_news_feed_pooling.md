# SUMMARY — pooled news feeds for the domain portfolio

**As at 2026-07-20, end of second day.** Design complete; first live
infrastructure exists.

## What we're trying to do

Give most of a thousand-plus domain portfolio a credible, regularly-updated news
section without paying for each domain separately: a small number of shared
article pools doing the expensive work once, with each domain still selecting
and presenting what suits its own audience — so two similarly-named sites never
read as copies of each other.

## Where we've come from

Yesterday established the architecture. The platform's existing news pipeline
works but is entirely per-site — every site fetches, stores and pays alone — and
it runs on just four sites, so there was nothing to migrate. The design splits
the work: fetching, de-duplicating and researching happen once per article in a
shared pool; only the choice of which articles matter to a given site is
per-site, and that runs on database machinery we already operate, at no marginal
cost. A dozen-odd pools cover about two thirds of the portfolio; the rest
shouldn't have news at all, because no real news exists for them. Each domain is
a separate site with its own target market, the near-duplicate names most of
all — which makes the per-site selection layer the product, not an optimisation.

## What we've done

Today turned decisions into live state. Every one of the eleven existing sites
now carries a structured audience profile — who it's for, how it differs from
our sibling sites, what that means for the copy — with the three genuinely
competing AI-services sites defined explicitly against each other. The two
rough profiles that already existed were migrated into the new shape with
nothing lost and nothing invented.

The seventeen pools now exist in the live system as synthetic sites — the same
device the platform already uses for its own internal work — each carrying a
default audience description that member sites will inherit and specialise.
Before creating them we verified, rather than assumed, that they are invisible
to every automated process that walks the fleet, and that the news machinery
cannot start spending money on them until we deliberately arm it.

The mystery of the high-traffic unknown domains is solved. An archive
investigation established what each one used to be, and the owner resolved the
rest: all nine are ours, one (nanangmrk — a thriving Indonesian tutorial site
with a half-million-subscriber YouTube channel behind it) is owner-run and will
be adopted into the framework as its own piece of work. The shopping-legacy
domains become honest retailer directories — categorised listings linking to
real suppliers, usefulness first, affiliate income later. The old fax-service
domain gets the same treatment for fax providers. Makeitaquote gets its
quote-image tool built, deliberately different from the Discord bot people are
guessing at, under our own branding. Komunikatif, a real Indonesian regional
news site until 2023, becomes the first foreign-language rebuild. And zdec — the
biggest number of the lot — turned out to be a hacked Chinese site full of
casino spam, so its traffic gets measured before anything is invested,
using the visitor-logging setup already built for relojistas.

The pilot cohort is chosen and grown: roughly thirty-seven domains, all with
real traffic, spanning fourteen of the seventeen pools and including the
portfolio's second-biggest traffic source.

## Where we are now

Design questions: none open. Live state: eleven profiled sites, seventeen inert
pools, a pilot list of thirty-seven, and four working documents plus a research
file recording how we got here — including the three wrong turns that were
caught and corrected along the way (a miscounted table, a misread revert, a
design of ours that ran against the platform's own idiom).

Money has not started being spent. The pools ingest nothing until sources are
chosen and switched on, and that switch is deliberately manual.

## Where we're going

Three steps, in order. First, pilot onboarding: the thirty-seven domains become
classified sites with audience profiles — profiles written from research, never
from the domain name alone. Second, one pool gets real news sources and starts
ingesting, alone, as the test case. Third, the duplication measurement runs on
that pool's real articles — how similar are member sites' feeds to each other —
and only when those numbers are acceptable does any member site render a feed.
The measurement gates the rollout because the risk it guards against — the
portfolio reading as one network of copies — is the one mistake that can't be
cleanly undone afterwards.

Two bugs still gate the pilot: news pages render nothing without JavaScript
(another thread has it), and the news listing hardcodes English, which blocks
the Romanian, Dutch, German, Spanish and now Indonesian domains in the cohort.
