# SUMMARY — gauntlet_dead_cta, 2026-07-31

Written to be read aloud. New file, as the rule requires; the previous read-out is
`SUMMARY_2026-07-29_gauntlet_dead_cta.md` and the series is the record.

## What we're trying to do

Make vonc.com a site a stranger would argue on, and prove the platform can build
and maintain an honest interactive product without a human checking every sentence.
The product half is settled for now: the owner ruled on 29 July that we run the
distribution experiment first — he posts the daily provocation and the share card
where people already argue — and let real behaviour choose between the two theses
we could pursue. The engineering half is everything that has to be true before
that experiment means anything.

## Where we've come from

The usability audit was fully built by the night of 28 July: a sealed door, a real
round against a real opponent, a verdict you can keep as a card. The opinion
ledger shipped on the 29th. The engine bug that was breaking judgements was closed
the same day. On the 30th we cleared the last engineering item that was gating
distribution — the social image, which had been serving a 404, now returns a real
1200×630 card.

Then two days of a different kind of work, and it is the more interesting part.
Almost nothing we did was building a feature. It was finding out that things we
believed were true were not, and deciding what to do about mechanisms that had
never actually run.

## What we've done

We ran, once, a scheduled task that had been switched off since May. It was the
only thing in the platform that moves a discovered defect into the queue a fixer
reads. Nothing else does it. One firing moved sixty-seven items, and a defect that
a check had correctly caught eighteen hours earlier — a control laid out off the
right edge of a phone screen — went to the repair agent, got fixed in
forty-six seconds, and the fix is verified on the live page. That is the first
time in this platform's life that particular loop has run end to end.

It also exposed four mechanisms getting their first ever execution, and half of
them failed. Two of the four repair items died at the first step. So the
convergence guard we built to stop a fixer looping forever still has never been
exercised, because nothing had ever reached it.

In the same run, the platform rewrote the copy on a live homepage and put a
fabricated human credential on it — a real sentence, "built for designers", turned
into "built by a designer" by one changed preposition, with invented supporting
detail generated to justify the new claim. We took that off the site. We also
found the platform correcting itself an hour later by accident rather than by any
control, which is not a control.

We took on and fixed a defect in our own service: every visitor to the Gauntlet
was being counted as the same person. The rate limiter was one bucket for the
whole site and the stored identity column had never distinguished anybody. That is
built, tested, and the image is waiting on the owner to swap it in.

And we ran the duplication check the owner asked for, which found that our own
about page was rendering every section twice, in public, for two days.

## Where we are now

The site is in good shape and better checked than it was. What is not finished is
knowingly not finished, and written down: four handoffs, one per issue, so each can
be picked up cold by a separate thread.

Two of those are about the same embarrassment. The daily provocation is not daily
— it never has been, and there is no mechanism by which it could be. It is a
hand-written file with a hardcoded "today", changed six times in a month by a
person. And the provocation the Gauntlet page carefully hides is printed in full
on the home page, so the sealed door only works for visitors who arrive by a route
almost nobody takes.

**And this is where the decisions came in, which is the point of this read-out.**

The owner asked a reasonable-sounding question: can we wire the duplication check
into the platform as a proper checker with a handler that fixes what it finds?
The honest answer turned out to be *not as one thing*, and the reason is worth
keeping.

The check finds two problems that look identical in the output and are opposites
underneath. The first is a page carrying the same section twice, word for word.
That has a deterministic fix — delete the later copy — and no judgement is
involved. The second is eight sibling pages that were each written on their own
and independently said much the same thing in different words. Fixing that means
rewriting prose, which means deciding what the page is *for*, which is judgement.

Treating them as one problem fails in whichever direction you pick. Build a
handler cautious enough for the second and it will not fix the first. Build one
confident enough for the first and you have pointed an automatic rewriter at copy
that might be perfectly good, on the strength of a similarity score — and a
similarity score says nothing about meaning. We already know what that costs,
because the same day we watched an automated rewrite invent a credential.

So the decision the owner made was to split it: the handler gets authority only
over the deterministic half, and the half needing judgement is recorded as work
the platform knows it cannot yet do. That second part is not a fudge — the
platform already has a way of saying "I found work I have no handler for", used
elsewhere, and it queues the problem for a human to decide on instead of guessing.
The structural fix for the judgement half belongs to another workstream that
already owns it, and telling them is part of this job rather than an afterthought.

There was a second, quieter decision in the same conversation. The duplication
check partly works by asking which approved facts a section repeats, and vonc has
only four approved facts, so that half of the check is nearly blind here. The
obvious response is to add more facts. But every fact we add is also a new thing
the site is licensed to state, and nothing yet checks the prose an AI writes
against those facts. Adding facts to improve a *check* would widen the surface for
exactly the failure we spent the previous day cleaning up. The owner's call was to
add only facts that carry their own verifying query, so that they cannot quietly
go stale, and to keep the pool tight rather than large.

Both decisions have the same shape, and it is the shape worth remembering: the
question was never "should we automate this". It was "what is this mechanism
allowed to be confident about". Every time we got that wrong this week — and we
got it wrong more than once — it was because something was given more authority
than its evidence supported.

## Where we're going

The owner posts. That is the experiment, and it is his leg.

Before that, the visitor-identity fix goes onto the island, because the identity
column is what the experiment would be measured on and it currently cannot tell
two people apart. The checker gets built for the deterministic half. The
provocation gets a mechanism, or the site stops promising a daily one. And the
home page stops giving away the thing the door is there to conceal.

One caution for whoever reads this next. Four separate times in two days a
measurement of ours was wrong because we compared against a rule we had invented
rather than the one the data declared — counting pages by the shape of their
names, testing for a technique by searching for its marketing name, calling a
pipeline idle from two observations four minutes apart, and building a URL instead
of reading the one on the record. Every one of those was caught, three of them
only because someone went back to check a number they had already reported. The
tally is the useful part, not any single instance: the failure is not carelessness,
it is that a plausible ruler feels exactly like the right one.
