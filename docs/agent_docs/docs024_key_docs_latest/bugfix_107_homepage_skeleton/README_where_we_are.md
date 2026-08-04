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

## 2026-08-04, later — standing down: this fix already belongs to someone

The research came back with something the ownership checks had missed: you
parked this bug on 27 July, and the "vigilant designer" programme you approved
on 2 August already has this exact fix scheduled as its Phase 4. So we are not
building it — that would be two teams building one thing.

What we did instead: wrote everything we learned into the bug file itself (a
full map of where the sameness actually comes from — it is one AI planning
step, plus two fallback paths that copy the same shape back in), handed the
research to the vigilant-designer thread in their own directory, and recorded
how the ownership checks missed this so the next session's checks are better.
The research is not wasted — whoever builds Phase 4 starts from a map instead
of a blank page.

This directory stays as the record; no further work planned here.
