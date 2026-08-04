# README — where we are (bugfix 107, the same-homepage problem)

Append-only, newest at the bottom. Plain prose.

## 2026-08-04 — picked up, still real, getting the groundwork right first

You said (27 July, via the oufe build) that the new site looked like "the
standard looking site that it has produced before". The bug file for that is
107, and nobody had picked it up — the other open bugs all have active threads.

First thing we did was check it is still true, because a week is a long time on
this estate. It is: the newest site the framework built (lendzy.co.uk, 2
August) came out with exactly the same shape again — hero at the top, a strip
of cards in the middle, "call to action" at the bottom. The sites that look
different are the ones where a person forced them to differ.

The plan, in short: teach the part that plans a page WHAT KIND of site it is
building — a publication, a directory, a tool site, a brochure — so the shape
follows the kind, instead of every kind getting the brochure shape. There is
already a written-up grid of "kinds of site" in the docs from earlier work, so
this is wiring existing vocabulary in, not inventing new theory.

Before changing anything we are running the platform's own diagnosis loop on
it, as the house rules require for a claim this structural — it reads the code
and the live data independently and tells us if the cause is really where we
think it is. Two research passes over the code and docs are running now.

(Also closed a stale file on the way in: bug 121, the "house voice"
duplication, was actually fixed and live on 27 July — the paperwork just never
caught up. Verified it end to end today and moved it to closed.)
