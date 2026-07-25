# relojistas.com — milestone summary, 25 July 2026

*Plain language, written to be read aloud. Previous read-out:
`SUMMARY_2026-07-24_relojistas_rebuild.md`. Technical entry point:
`HANDOFF_RESUME_relojistas_rebuild.md`. This one marks a real inflection: the
build list is finished — the last item by deciding not to build it — and the
mission metric has finally been measured.*

## What we're trying to do

Take a dead Spanish watch forum whose audience never fully left, and turn it into
a Spanish-language watch news portal that serves that audience automatically —
the old RSS subscribers at their original address, the searchers who still type
watch queries into it, and the crawlers that decide whether anyone else ever
finds it.

## Where we've come from

The feed was reactivated, the site gained real reference content behind a
fabrication fence, the news became visible to machines as well as browsers, and a
routing mistake of ours that had the homepage rewriting itself every six hours
was fixed and proven fixed by the site's own traffic. Search capture came back on
the homepage, and the engine behind the site was taught to answer searches rather
than merely record them. What remained was one feature and one number: per-board
feeds, and how many people actually came back.

## What we've done since the last read-out

**We finished the build list by deciding not to build the last thing on it.** The
plan was to map the old forum's boards — Panerai, Rolex, Seiko, divers — to
topic-filtered feeds, so someone who followed a board in 2015 would get that
subject's news today at the address they never unsubscribed from. Before building
it we went to find out which boards to serve, and the survey killed the premise.
Nearly nine in ten requests for board feeds come from crawlers that say so in
their own name, and not one of them behaves like a subscribed reader. The real
readers — including one that politely asks "anything new?" forty-two times and is
told no — are subscribed to the **site**, not to any board. We would have been
building for Google's indexer.

There was a second, independent reason. We could not have filled those feeds.
The most-requested boards are things a news service has no content for — legal
advice, group buys, member introductions, a pub-chat board — and across our
entire collection there is one Seiko story and no Louis Erard at all. A feed with
nothing in it looks, to a returning subscriber, exactly like the dead forum we are
undoing.

The decision is written down with the evidence, and with the specific condition
that would reverse it, so it is a judgement on record rather than a thing we
quietly dropped.

**We measured the thing the project exists for.** The old feed address failed
every request for years and until **17 July** this year; since the 18th it has
answered. Today it served 36 requests and failed none. Behind that there is at
least one genuine returning subscriber, plus Google's feed fetcher, which only
polls feeds a real person asked it to. What we still cannot do is count *people*,
because every visitor currently arrives wearing Cloudflare's address — which
turns a piece of pending housekeeping into the thing standing between us and the
headline number.

**Some of our own record turned out to be wrong, and has been corrected.** The
claim that we knew which boards people had subscribed to was never checked; it
was a count of requests described with a word that means people. It sat in the
plan for two weeks and nearly produced a feature. The correction is recorded where
the claim was made, and the cheap check that would have caught it — read the
column that says who is asking — is now written into the runbook and the
fleet-wide log of wrong calls.

## Where we are now

The site runs itself and the build list is empty. News arrives every six hours;
the feed serves its subscribers at the original address; reference content is
fenced against fabrication; the homepage survives its own refresh cycles; search
captures demand and can answer as soon as it is switched on. Everything claimed
here has been checked against the live site or the live logs rather than assumed.

One item remains, and it is the owner's: a single session on the server that
turns on search answers, starts counting real visitors, enables the collector,
and fixes the last handful of failing feed addresses. Every remaining unknown in
this project is downstream of it.

## Where we're going

After that session: confirm the residual feed failures have gone, and re-check
the one condition that would revive per-board feeds now that real visitor
addresses are visible. Then the measurement this was all for — how many real
people came back — becomes answerable for the first time.

Separately, the work on this box has opened a wider question the owner has now
directed us to take up: the server is still configured by a script a human runs
by hand, which is the same kind of untracked manual repair the platform refuses
to accept for web pages. That is now its own workstream, alongside merging it
with the tools-api box rather than running two estates.

## The one-sentence version

The dead forum is a self-running Spanish watch portal with nothing left on its
build list, its returning subscribers measurably served since 17 July, and one
owner session standing between it and the only number that ever mattered.
