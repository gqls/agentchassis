# SUMMARY — gauntlet_dead_cta, 2026-07-28 (evening)

*The series is the record: previous entry SUMMARY_2026-07-27b. New file, never
an overwrite.*

## What we're trying to do

Make vonc.com's Gauntlet — a timed, judged argument against a real AI opponent
— something a first-time visitor can walk into, understand, play and come away
from with something in hand, without a single fabricated number or dead
control anywhere on the page.

## Where we've come from

By yesterday the machinery was honest and reliable: real rounds, honest
failures, a diagnosable engine, zero fabrication across 49 components. Then
the owner used the site as a visitor and found eight things no automated check
had seen — an invisible headline, content cut off phone screens, a primary
button that revealed nothing, a provocation that read as filler, a busy page
with no pecking order, and no way to keep a verdict. That list (bugs_open/131,
the gauntlet one — the number now names two cases) became the work.

## What we've done

All of it, in one day, most of it today. The headline reads (A). The cut
content is fixed on the page AND the fleet's blind checker can now see the
fault class everywhere (B — it caught its own author twice during the later
builds). The page opens sealed behind one door: pressing Enter the Gauntlet
starts a real round and the question is revealed by the engine's answer, on
the first screen instead of two and a half screens down (C — the owner chose
this over position-as-entry). The provocation sits in the page's only marked
card with the "take a position" line attached to it (E). The steps carry a
visible ranking — current bright, done receded, later dimmed — driven only by
real API responses, with nothing dimmed ever disabled (F). And a finished
round can be kept: a shareable 1200×630 card drawn from the visitor's own
provocation and the judge's actual verdict text, nothing else (G — the owner
picked the card over a permalink).

Alongside: the fleet-wide unbounded-HTTP-client defect this workstream
surfaced was filed, fixed, council-approved and closed the same day
(bugs_closed/130) — proven live on both the cluster and the island, the
island half by byte-identical binary hash after the owner ran the swap by
hand. The day also contained the fleet's Anthropic allowance running out and
being raised; the armed engine log caught it within seconds, its first-ever
catch.

## Where we are now

The Gauntlet is honest, reliable, and — for the first time — designed. A
visitor meets one door, one question, one thing to do at each moment, and
leaves with a real artefact. Nothing on the page claims what is not true by
construction; nothing changes state except on a real API response, including
the reveal itself and the card.

## Where we're going

Two residuals and one question. The rebuilt overflow checker is live but has
yet to be witnessed on a real acceptance run. The gauntlet page's social
preview image 404s (the og:image case that shares the number 131) — when that
is fixed fleet-wide, shares of the Gauntlet get a face. And the open question
is unchanged from yesterday, sharpened by a user's own words — "why argue with
an AI when Perplexity is free" (H): the site now deserves visitors; nothing on
it can manufacture them. That is a product decision, deliberately deferred by
the owner's design-first ruling, and it is the next real conversation.
